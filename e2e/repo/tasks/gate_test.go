package tasks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

func miseTask(name string) string {
	return filepath.Join(harness.Root, "repo", "tools", "mise", "tasks", name)
}

// TestRunGate: the gate code:lint:60-bootstrap. It runs `bootstrap check`
// with the cache clone, and it never fetches. It runs with --all, and gets
// its files from gate_files of the working tree's helpers and gate_files.py.
func TestRunGate(t *testing.T) {
	// runGate runs the gate in a new consumer. <origin>, when not empty, is
	// the origin of a cache clone.
	runGate := func(t *testing.T, origin string) (int, string) {
		dir := harness.Dir(t)
		root := filepath.Join(dir, "gate-consumer")
		cache := filepath.Join(dir, "gate-cache")
		copyConsumer(t, root)
		scripts := filepath.Join(root, ".config", "mise", "tasks", "_scripts")
		copyFile(t, helpersPath(), filepath.Join(scripts, "helpers"))
		copyFile(t, miseTask("gate_files.py"), filepath.Join(scripts, "gate_files.py"))
		rc.WriteValues(t, filepath.Join(root, ".config", "bootstrap"), `environments = [ "dev" ]`)
		if origin != "" {
			repo := filepath.Join(cache, "repo")
			harness.Git(t, dir, "clone", "--quiet", "--", rc.BaseClone, repo)
			harness.Git(t, repo, "remote", "set-url", "origin", origin)
		}
		env := harness.SetEnv(harness.Env(), "usage_files", "BOOTSTRAP_VALUES", "BOOTSTRAP_CLONE",
			"usage_all=true", "MISE_PROJECT_ROOT="+root, "BOOTSTRAP_CACHE="+cache,
			"BOOTSTRAP_URL=file://"+filepath.Join(dir, "no-such-remote"), "BOOTSTRAP_BIN="+harness.Bin)
		cmd := &harness.Cmd{Env: env, Args: []string{"bash", miseTask("code-lint-bootstrap")}}
		return cmd.Run()
	}

	t.Run("without a cache", func(t *testing.T) {
		t.Parallel()
		got, log := runGate(t, "")
		if got == 0 && log == "" {
			harness.Pass(t, "the gate without a cache: exit 0, no output")
		} else {
			harness.Fail(t, fmt.Sprintf("the gate without a cache: exit %d, or an output", got), log)
		}
	})
	t.Run("with a cache", func(t *testing.T) {
		t.Parallel()
		got, log := runGate(t, "https://github.com/virajp/bootstrap.git")
		if got == 0 && strings.Contains(log, " is valid") {
			harness.Pass(t, "the gate with a cache: exit 0, valid")
		} else {
			harness.Fail(t, fmt.Sprintf("the gate with a cache: exit %d, or no valid line", got), log)
		}
	})
	t.Run("with another origin", func(t *testing.T) {
		t.Parallel()
		got, log := runGate(t, "https://github.com/virajp/other.git")
		if got == 0 && log == "" {
			harness.Pass(t, "the gate with another origin: exit 0, no output")
		} else {
			harness.Fail(t, fmt.Sprintf("the gate with another origin: exit %d, or an output", got), log)
		}
	})
}

var (
	gfOnce         sync.Once
	gfPython, gfUV string
	gfErr          error
)

// gfTools gives python 3.14 and uv of mise, as the old task found them.
func gfTools(t *testing.T) (string, string) {
	t.Helper()
	gfOnce.Do(func() {
		which := func(tool, bin string) string {
			cmd := exec.Command(harness.RealMise, "x", tool, "--", "which", bin)
			cmd.Dir = harness.Root
			cmd.Env = harness.Env()
			out, err := cmd.Output()
			if err != nil && gfErr == nil {
				gfErr = fmt.Errorf("mise x %s -- which %s: %v", tool, bin, err)
			}
			return strings.TrimSpace(string(out))
		}
		gfPython = which("python@3.14", "python3")
		gfUV = which("uv@latest", "uv")
	})
	if gfErr != nil {
		t.Fatal(gfErr)
	}
	return gfPython, gfUV
}

// gf is one scratch repository of a gate_files case, with its gate file
// outside it.
type gf struct {
	t                  *testing.T
	dir, gate, errFile string
	env                []string
	code               int
	list               string
	listed             bool
}

// newGF is gf_new: a new scratch repository with no commit, and its gate
// file without a #FILES header.
func newGF(t *testing.T) *gf {
	t.Helper()
	base := harness.Dir(t)
	g := &gf{t: t, dir: filepath.Join(base, "repo"), gate: filepath.Join(base, "gate"), errFile: filepath.Join(base, "err")}
	// git without the global and the system config, and without the default
	// excludes file $XDG_CONFIG_HOME/git/ignore, so an ignore file of the
	// user never hides a file of a case.
	g.env = harness.SetEnv(harness.Env(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"XDG_CONFIG_HOME="+filepath.Join(base, "gf-no-config"))
	if err := os.MkdirAll(g.dir, 0o755); err != nil {
		t.Fatal(err)
	}
	g.git("", "init", "--quiet", "-b", "main")
	g.header()
	return g
}

// git runs git in <sub> of the repository, as gf_git did.
func (g *gf) git(sub string, args ...string) {
	g.t.Helper()
	all := append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false",
		"-c", "core.hooksPath=/dev/null", "-C", filepath.Join(g.dir, sub)}, args...)
	cmd := exec.Command("git", all...)
	cmd.Env = g.env
	if out, err := cmd.CombinedOutput(); err != nil {
		g.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// header is gf_header: the top block of the gate file, with <lines> after
// its #USAGE line.
func (g *gf) header(lines ...string) {
	all := append([]string{"#!/usr/bin/env bash", "", `#USAGE flag "--all"`}, lines...)
	all = append(all, "", "set -e")
	harness.WriteFile(g.t, g.gate, harness.Lines(all...))
}

// file is gf_file: a file of the repository, with <lines> (default "x").
func (g *gf) file(path string, lines ...string) {
	if len(lines) == 0 {
		lines = []string{"x"}
	}
	harness.WriteFile(g.t, filepath.Join(g.dir, path), harness.Lines(lines...))
}

func (g *gf) commit() {
	g.git("", "add", "-A")
	g.git("", "commit", "--quiet", "--no-verify", "-m", "files")
}

// run is gf_run: the script, with --debug.
func (g *gf) run(args ...string) {
	python, uv := gfTools(g.t)
	errF, err := os.Create(g.errFile)
	if err != nil {
		g.t.Fatal(err)
	}
	defer func() { _ = errF.Close() }()
	var out bytes.Buffer
	cmd := exec.Command(uv, append([]string{"run", "--quiet", "--python", python, "--script", miseTask("gate_files.py"),
		g.gate, "--debug"}, args...)...)
	cmd.Dir = g.dir
	cmd.Env = harness.SetEnv(g.env, "MISE_PROJECT_ROOT="+g.dir)
	cmd.Stdout, cmd.Stderr = &out, errF
	g.code, g.list, g.listed = 0, "", false
	if err := cmd.Run(); err != nil {
		g.code = 127
		if ee, ok := err.(*exec.ExitError); ok {
			g.code = ee.ExitCode()
		}
	}
	p := strings.TrimRight(out.String(), "\n")
	if p != "" {
		if b, err := os.ReadFile(p); err == nil {
			g.listed = true
			g.list = strings.TrimRight(string(b), "\n")
			_ = os.Remove(p)
		}
	}
}

// check is gf_check: the exit code, the list (one path for each line), and
// each fixed string of <stderr> on stderr. A case with exit 0 needs the list
// file.
func (g *gf) check(name string, want int, list []string, stderr ...string) {
	g.t.Helper()
	wantList := strings.Join(list, "\n")
	errText := harness.ReadFile(g.errFile)
	bad := ""
	switch {
	case g.code != want:
		bad = fmt.Sprintf("exit %d, expected %d", g.code, want)
	case want == 0 && !g.listed:
		bad = "no list file"
	case g.list != wantList:
		bad = fmt.Sprintf("the list is '%s', expected '%s'", strings.ReplaceAll(g.list, "\n", " "), strings.Join(list, " "))
	default:
		for _, s := range stderr {
			if !strings.Contains(errText, s) {
				bad = "stderr lacks: " + s
				break
			}
		}
	}
	if bad == "" {
		harness.Pass(g.t, "gate_files, "+name)
		return
	}
	harness.Fail(g.t, "gate_files, "+name+": "+bad, errText)
}

// changes is gf_changes: a staged change, an unstaged change, a deleted
// file, a staged deletion and a new file.
func changes(t *testing.T) *gf {
	g := newGF(t)
	for _, f := range []string{"staged.txt", "changed.txt", "deleted.txt", "removed.txt", "same.txt"} {
		g.file(f)
	}
	g.commit()
	g.file("staged.txt", "x", "y")
	g.git("", "add", "staged.txt")
	g.file("changed.txt", "x", "y")
	_ = os.Remove(filepath.Join(g.dir, "deleted.txt"))
	g.git("", "rm", "--quiet", "removed.txt")
	g.file("new.txt")
	return g
}

// repoFilter is gf_repo: a submodule (a gitlink), a nested repository, an
// ignored file and a symlink.
func repoFilter(t *testing.T) *gf {
	g := newGF(t)
	g.file(".gitignore", "ignored.txt")
	g.file("a.txt")
	g.file("ignored.txt")
	for _, d := range []string{"sub", "nested"} {
		if err := os.MkdirAll(filepath.Join(g.dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	g.git("sub", "init", "--quiet", "-b", "main")
	g.file("sub/x.txt")
	g.git("sub", "add", "-A")
	g.git("sub", "commit", "--quiet", "--no-verify", "-m", "sub")
	g.git("", "add", "-A")
	g.git("", "commit", "--quiet", "--no-verify", "-m", "files")
	g.git("nested", "init", "--quiet", "-b", "main")
	g.file("nested/n.txt")
	g.file("sub/x.txt", "x", "y")
	g.file("sub/new.txt")
	g.file("new.txt")
	if err := os.Symlink("a.txt", filepath.Join(g.dir, "link.txt")); err != nil {
		t.Fatal(err)
	}
	return g
}

// TestGateFiles: gate_files.py, the file list of a code:format or code:lint
// gate. Each case runs the working tree's script in its own scratch
// repository, with the gate file outside it, and compares the exit code, the
// list and the --debug lines on stderr.
func TestGateFiles(t *testing.T) {
	sub := func(name string, f func(t *testing.T)) {
		t.Run(name, func(t *testing.T) { t.Parallel(); f(t) })
	}

	// The uncommitted files. --all: each tracked and new file.
	sub("changes, the default", func(t *testing.T) {
		g := changes(t)
		g.run()
		g.check("the default: the staged, the changed and the new files", 0,
			[]string{"changed.txt", "new.txt", "staged.txt"}, "drop deleted.txt: repo deleted")
	})
	sub("changes, --all", func(t *testing.T) {
		g := changes(t)
		g.run("--all")
		g.check("--all: each tracked and new file, no deleted file", 0,
			[]string{"changed.txt", "new.txt", "same.txt", "staged.txt"}, "drop deleted.txt: repo deleted")
	})

	// The repository filter, in each mode.
	sub("repo, the default", func(t *testing.T) {
		g := repoFilter(t)
		g.run()
		g.check("the default: no file of a submodule or a nested repository", 0, []string{"new.txt"},
			"drop nested/: repo nested-repo", "drop link.txt: repo symlink")
	})
	sub("repo, --all", func(t *testing.T) {
		g := repoFilter(t)
		g.run("--all")
		g.check("--all: no file of a submodule or a nested repository", 0, []string{".gitignore", "a.txt", "new.txt"},
			"drop sub: repo gitlink", "drop nested/: repo nested-repo", "drop link.txt: repo symlink")
	})
	sub("repo, given files", func(t *testing.T) {
		g := repoFilter(t)
		g.run("a.txt", "ignored.txt", "sub/x.txt", "sub", "nested/n.txt", "missing.txt", "link.txt")
		g.check("given files: an ignored, a submodule, a nested and a missing path go", 0, []string{"a.txt"},
			"drop ignored.txt: repo unlisted", "drop sub/x.txt: repo unlisted", "drop sub: repo gitlink",
			"drop nested/n.txt: repo unlisted", "drop missing.txt: repo unlisted", "drop link.txt: repo symlink")
	})
	sub("repo, --all with files", func(t *testing.T) {
		g := repoFilter(t)
		g.run("--all", "a.txt")
		g.check("--all with files: exit 2", 2, nil)
	})

	// The global filter, then the gate filter and a malformed header, in
	// one repository, in order.
	sub("global and gate filters", func(t *testing.T) {
		g := newGF(t)
		g.file("a.txt")
		g.file("skip.txt")
		g.file("b.md")
		g.file(".config/bootstrap/config-files.toml", `includes = [ "*.txt" ]`, `excludes = [ "skip.txt" ]`)
		g.commit()
		g.run("--all")
		g.check("the global includes and excludes", 0, []string{"a.txt"},
			"drop b.md: global include", `drop skip.txt: global exclude "skip.txt"`,
			"drop .config/bootstrap/config-files.toml: global include")
		// A missing file means two empty lists, and a gate without a #FILES
		// header gets each file.
		_ = os.Remove(filepath.Join(g.dir, ".config", "bootstrap", "config-files.toml"))
		g.commit()
		g.run("--all")
		g.check("no config-files.toml and no #FILES header: each file", 0, []string{"a.txt", "b.md", "skip.txt"})

		g.header(`#FILES include="*.txt" exclude="skip.txt"`)
		g.run("--all")
		g.check("the #FILES include and exclude", 0, []string{"a.txt"},
			"drop b.md: gate include", `drop skip.txt: gate exclude "skip.txt"`)

		// An empty list: an empty list file, exit 0.
		g.header(`#FILES include="*.go"`)
		g.run("--all")
		g.check("an empty list", 0, nil)

		// A malformed header stops with exit 2 and a message.
		g.header(`#FILES include="*.txt"`, `#FILES exclude="b.txt"`)
		g.run("--all")
		g.check("two #FILES lines", 2, nil, "two #FILES lines")
		g.header(`#FILES only="*.txt"`)
		g.run("--all")
		g.check("an unknown #FILES key", 2, nil, "unknown #FILES key only")
		g.header(`#FILES include=*.txt`)
		g.run("--all")
		g.check("a #FILES value without quotes", 2, nil, "the #FILES value of include has no double quotes")
	})

	// The shebang key: a file that only the shebang selects needs a shell of
	// the list; a file of the include names goes only for a known shell that
	// is not in the list.
	sub("the shebang key", func(t *testing.T) {
		g := newGF(t)
		g.header(`#FILES include="*.sh" shebang="sh bash"`)
		g.file("run", "#!/usr/bin/env bash")
		g.file("run-s", "#!/usr/bin/env -S bash -e")
		g.file("zrun", "#!/bin/zsh")
		g.file("py.sh", "#!/usr/bin/env python3")
		g.file("zsh.sh", "#!/bin/zsh")
		g.file("plain.sh", "echo x")
		g.file("notes.txt")
		g.commit()
		g.run("--all")
		g.check("the shebang key", 0, []string{"plain.sh", "py.sh", "run", "run-s"},
			"drop zrun: gate include", "drop zsh.sh: gate shebang", "drop notes.txt: gate include")
	})

	// The patterns: a pattern without / matches in each folder, a pattern
	// with / from the root, and ** any number of folders.
	sub("the patterns", func(t *testing.T) {
		g := newGF(t)
		g.header(`#FILES exclude="node_modules *.log gen/*.txt docs/**/*.md"`)
		for _, f := range []string{"node_modules/a.js", "web/node_modules/pkg/b.js", "node_modules.txt", "a.log",
			"x/y/b.log", "gen/a.txt", "x/gen/a.txt", "docs/a.md", "docs/x/y/b.md", "top.md"} {
			g.file(f)
		}
		g.commit()
		g.run("--all")
		g.check("the patterns", 0, []string{"node_modules.txt", "top.md", "x/gen/a.txt"},
			`drop web/node_modules/pkg/b.js: gate exclude "node_modules"`, `drop x/y/b.log: gate exclude "*.log"`,
			`drop gen/a.txt: gate exclude "gen/*.txt"`, `drop docs/x/y/b.md: gate exclude "docs/**/*.md"`,
			`drop docs/a.md: gate exclude "docs/**/*.md"`)
	})
}

// TestLintOrchestrator: the orchestrator code:lint stops with exit 2 for
// --all with files.
func TestLintOrchestrator(t *testing.T) {
	t.Parallel()
	dir := harness.Dir(t)
	root := filepath.Join(dir, "lint-orchestrator")
	copyFile(t, helpersPath(), filepath.Join(root, ".config", "mise", "tasks", "_scripts", "helpers"))
	task := filepath.Join(dir, "code-lint")
	gomplate(t, miseTask("code-lint.tmpl"), task, "")
	env := harness.SetEnv(harness.Env(), "MISE_PROJECT_ROOT="+root, "usage_all=true", "usage_files=a.txt")
	cmd := &harness.Cmd{Env: env, Args: []string{"bash", task}}
	got, log := cmd.Run()
	if got == 2 && strings.Contains(log, "Use --all or files, not both.") {
		harness.Pass(t, "code:lint --all with files: exit 2")
	} else {
		harness.Fail(t, fmt.Sprintf("code:lint --all with files: exit %d, or no usage message", got), log)
	}
}
