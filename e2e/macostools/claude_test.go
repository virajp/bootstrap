package macostools

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

// The Claude files of the Mac tools (plan 2026-11-27-tool-skills, D8, D9,
// D13 c, D15, D17 d): the skill secrets and the npm hook below $HOME, the
// skills mac-setup and git-identity in the backup clone, and the hook itself.
func TestClaudeFiles(t *testing.T) {
	t.Run("backup", testClaudeBackup)
	t.Run("no backup", testClaudeNoBackup)
	t.Run("hook", testClaudeHook)
}

const hookCommand = "$HOME/.claude/hooks/npm-normalize.sh"

// withBackup writes the copy of the example with the key backup set to a new
// git repository, and gives the folder of that clone.
func (c *macCase) withBackup() string {
	clone := filepath.Join(c.dir, "backup")
	c.git("init", "--quiet", clone)
	harness.WriteFile(c.t, filepath.Join(c.config, "macos.toml"), "backup = \""+clone+"\"\n"+harness.ReadFile(example))
	c.commitConfig()
	return clone
}

func isRegular(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode().IsRegular()
}

func isRealDir(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.IsDir()
}

func statOK(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// hasHookItem: settings.json has a PreToolUse item with the matcher Bash and
// the command of the npm hook.
func hasHookItem(path string) bool {
	var s struct {
		Hooks struct {
			PreToolUse []struct {
				Matcher string `json:"matcher"`
				Hooks   []struct {
					Command string `json:"command"`
					Type    string `json:"type"`
				} `json:"hooks"`
			} `json:"PreToolUse"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(harness.ReadFile(path)), &s); err != nil {
		return false
	}
	for _, it := range s.Hooks.PreToolUse {
		for _, h := range it.Hooks {
			if it.Matcher == "Bash" && h.Command == hookCommand && h.Type == "command" {
				return true
			}
		}
	}
	return false
}

// checkHomeFiles: the skill secrets, the hook and its settings item below
// $HOME.
func (c *macCase) checkHomeFiles() {
	c.checkFile("~/.claude/skills/secrets/SKILL.md exists", statOK(filepath.Join(c.home, ".claude/skills/secrets/SKILL.md")))
	c.checkFile("~/.claude/skills/secrets/references/local.md exists",
		statOK(filepath.Join(c.home, ".claude/skills/secrets/references/local.md")))
	c.checkFile("~/.claude/hooks/npm-normalize.sh exists", statOK(filepath.Join(c.home, ".claude/hooks/npm-normalize.sh")))
	c.checkFile("~/.claude/settings.json has the PreToolUse item of the hook",
		hasHookItem(filepath.Join(c.home, ".claude/settings.json")))
}

// testClaudeBackup: with backup, the two backup skills are real files in the
// clone, and a foreign item of ~/.claude/skills/ stays a real item in $HOME.
func testClaudeBackup(t *testing.T) {
	t.Parallel()
	c := newCase(t, "claude backup")
	clone := c.withBackup()
	foreign := filepath.Join(c.home, ".claude/skills/foreign")
	harness.WriteFile(t, filepath.Join(foreign, "SKILL.md"), "foreign\n")
	c.runReal()
	c.checkExit(0)
	c.checkHomeFiles()
	for _, skill := range []string{"mac-setup", "git-identity"} {
		c.checkFile(".claude/skills/"+skill+"/SKILL.md is a real file in the clone",
			isRegular(filepath.Join(clone, ".claude/skills", skill, "SKILL.md")))
		c.checkFile(".claude/skills/"+skill+"/ is not below $HOME", !statOK(filepath.Join(c.home, ".claude/skills", skill)))
	}
	c.checkFile("~/.claude/skills is a real folder", isRealDir(filepath.Join(c.home, ".claude/skills")))
	c.checkFile("~/.claude/hooks is a real folder", isRealDir(filepath.Join(c.home, ".claude/hooks")))
	c.checkFile("the foreign item of ~/.claude/skills/ stays a real folder", isRealDir(foreign) &&
		isRegular(filepath.Join(foreign, "SKILL.md")))
	c.checkFile("the foreign item is not moved into the clone", !statOK(filepath.Join(clone, "dotfiles/.claude/skills/foreign")))
	c.runReal()
	c.checkExit(0)
}

// testClaudeNoBackup: without backup, the two backup skills are skipped.
func testClaudeNoBackup(t *testing.T) {
	t.Parallel()
	c := newCase(t, "claude no backup")
	c.runReal()
	c.checkExit(0)
	c.checkHomeFiles()
	for _, skill := range []string{"mac-setup", "git-identity"} {
		c.check(strings.Contains(c.logText(), "skip file  @backup/.claude/skills/"+skill+"/SKILL.md (no backup in macos.toml)"),
			"the skill "+skill+" is skipped", "no skip line for the skill "+skill)
		c.checkFile("no skill "+skill+" below $HOME", !statOK(filepath.Join(c.home, ".claude/skills", skill)))
	}
}

// runHook runs the synced hook with <input> on stdin and a PATH of the
// folder of jq, the system folders and, with <pnpm>, a fake pnpm.
func runHook(t *testing.T, hook, dir, input string, pnpm bool) string {
	jq, err := exec.LookPath("jq")
	if err != nil {
		t.Fatal("the hook test needs jq on PATH")
	}
	path := filepath.Dir(jq) + ":/usr/bin:/bin"
	if pnpm {
		bin := filepath.Join(dir, "pnpm-bin")
		harness.WriteFile(t, filepath.Join(bin, "pnpm"), "#!/bin/sh\nexit 0\n")
		if err := os.Chmod(filepath.Join(bin, "pnpm"), 0o755); err != nil {
			t.Fatal(err)
		}
		path = bin + ":" + path
	}
	cmd := exec.Command(hook)
	cmd.Env = []string{"PATH=" + path, "HOME=" + dir}
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("the hook: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

// sameJSON: <got> and <want> are the same JSON value.
func sameJSON(got, want string) bool {
	var g, w any
	if json.Unmarshal([]byte(got), &g) != nil || json.Unmarshal([]byte(want), &w) != nil {
		return false
	}
	gb, _ := json.Marshal(g)
	wb, _ := json.Marshal(w)
	return string(gb) == string(wb)
}

// testClaudeHook: the synced npm hook changes npm only when pnpm is on PATH,
// gives no permission decision, keeps the other keys of tool_input, and
// changes no command with a newline, a quote or a backslash.
func testClaudeHook(t *testing.T) {
	t.Parallel()
	c := newCase(t, "claude hook")
	c.runReal()
	c.checkExit(0)
	hook := filepath.Join(c.home, ".claude/hooks/npm-normalize.sh")
	fi, err := os.Stat(hook)
	c.checkFile("the hook has mode 755", err == nil && fi.Mode().Perm() == 0o755)
	if err != nil {
		return
	}
	input := `{"tool_name":"Bash","tool_input":{"command":"npm ci && npx tsc && npm install x","timeout":5000,"description":"d"}}`
	want := `{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":` +
		`{"command":"pnpm install --frozen-lockfile && pnpm exec tsc && pnpm install x","timeout":5000,"description":"d"}}}`
	for _, tc := range []struct {
		name, input, want string
		pnpm              bool
	}{
		{"pnpm on PATH: npm, npx and npm ci change, the other keys stay, no decision", input, want, true},
		{"no pnpm on PATH: {}", input, `{}`, false},
		{"no npm in the command: {}", `{"tool_input":{"command":"ls -la"}}`, `{}`, true},
		{"a command with a newline: {}", `{"tool_input":{"command":"npm ci\nnpx tsc"}}`, `{}`, true},
		{"npm ci before ; with no space", `{"tool_input":{"command":"npm ci;npm test"}}`,
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":{"command":"pnpm install --frozen-lockfile;pnpm test"}}}`, true},
		{"npm ci with two spaces", `{"tool_input":{"command":"npm  ci"}}`,
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":{"command":"pnpm install --frozen-lockfile"}}}`, true},
		{"a quoted ; npm: {}", `{"tool_input":{"command":"git commit -m \"docs; npm install gone\""}}`, `{}`, true},
		{"a backslash: {}", `{"tool_input":{"command":"echo a\\; npm i"}}`, `{}`, true},
		{"npx becomes pnpm exec", `{"tool_input":{"command":"npx tsc"}}`,
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":{"command":"pnpm exec tsc"}}}`, true},
		{"npx keeps its flags", `{"tool_input":{"command":"npx -y tsc"}}`,
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":{"command":"pnpm exec -y tsc"}}}`, true},
	} {
		got := runHook(t, hook, c.dir, tc.input, tc.pnpm)
		c.check(sameJSON(got, tc.want), "hook: "+tc.name, "hook: "+tc.name+": got "+got)
	}
}
