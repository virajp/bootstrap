// Package commands tests the catalog keys, the questions, the JSON envelopes,
// the language servers, the rendered tasks code:lsp:check, setup:ai, setup,
// setup:configs, git:merge:<branch> and git:worktree:*, the renames of
// setup:worktree and the code_files of the mise tool. The cases came from
// test:repo, lines 4560 to 5568.
package commands

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

func TestMain(m *testing.M) { rc.Main(m) }

const (
	devValues    = `environments = [ "dev" ]`
	noToolValues = devValues
	vtestValues  = devValues + "\n[tools.vtest]"
	vtestAt      = "repo/tools/vtest/tools.toml: "
	pythonValues = devValues + "\n[tools.python]"
	bothCore     = devValues + "\ncore_branches = [ \"main\", \"develop\" ]"
	tasksPath    = ".config/mise/tasks"
)

// thisHost and otherHost: the host of the test, as the engine names it, and
// a host other than it.
func hosts() (this, other string) {
	this = runtime.GOOS
	if this == "darwin" {
		this = "macos"
	}
	other = "macos"
	if this == "macos" {
		other = "linux"
	}
	return this, other
}

var (
	toolDirOnce sync.Once
	jqDir       string
)

// toolDirs finds the folder of jq (from `mise x jq@latest -- which jq`), as
// the old task did.
func toolDirs(t testing.TB) string {
	t.Helper()
	toolDirOnce.Do(func() {
		find := func(args ...string) string {
			cmd := &harness.Cmd{Args: args}
			code, out := cmd.Run()
			if code != 0 {
				return ""
			}
			lines := strings.Split(strings.TrimSpace(out), "\n")
			return filepath.Dir(strings.TrimSpace(lines[len(lines)-1]))
		}
		jqDir = find("mise", "x", "jq@latest", "--", "which", "jq")
	})
	if jqDir == "" {
		t.Fatalf("no jq: jq in %q", jqDir)
	}
	return jqDir
}

// run is one command of a Call or of a check.
type run struct {
	dir   string   // the working folder; empty is the root of the repository
	env   []string // the environment; nil is the environment of the case
	out   string   // the file of stdout; empty is the log of the case
	trunc bool     // write the file again, not at its end
	quiet bool     // stderr goes nowhere (2> /dev/null)
}

// do runs <args> and gives its exit code. Its output goes to the end of the
// log of the case (or of <out>). A Call that uses do does not use c.Engine.
func (r run) do(c *rc.Case, args ...string) int {
	c.T.Helper()
	out := r.out
	if out == "" {
		out = c.Log
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if r.trunc {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	f, err := os.OpenFile(out, flag, 0o644)
	if err != nil {
		c.T.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = r.dir
	if cmd.Dir == "" {
		cmd.Dir = harness.Root
	}
	cmd.Env = r.env
	if cmd.Env == nil {
		cmd.Env = c.Env
	}
	cmd.Stdout = f
	if !r.quiet {
		cmd.Stderr = f
	}
	err = cmd.Run()
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	_, _ = f.WriteString(err.Error() + "\n")
	return 127
}

// engineOut runs the engine with its stdout alone to the log (2> /dev/null).
func engineOut(c *rc.Case, env []string, args ...string) int {
	return run{env: env, quiet: true}.do(c, append([]string{harness.Bin}, args...)...)
}

// prependPath gives <env> with <dirs> before its PATH.
func prependPath(env []string, dirs ...string) []string {
	return harness.SetEnv(env, "PATH="+strings.Join(append(dirs, harness.GetEnv(env, "PATH")), string(os.PathListSeparator)))
}

// catTool writes the tool <tool> of the clone with no file, the items
// <depends> in its depends list, and <lines> after its top-level keys
// (cat_tool).
func catTool(c *rc.Case, tool, depends string, lines ...string) {
	head := []string{`default_envs = [ "dev" ]`, "depends = [ " + depends + " ]", "root_only = false"}
	harness.WriteFile(c.T, c.ClonePath("repo/tools/"+tool+"/tools.toml"), harness.Lines(append(head, lines...)...))
}

// prependFile writes <lines> before the content of the file <path>.
func prependFile(c *rc.Case, path string, lines ...string) {
	harness.WriteFile(c.T, path, harness.Lines(lines...)+harness.ReadFile(path))
}

// firstLine: a top-level line before the other lines of a tools.toml of the
// clone (first_line).
func firstLine(c *rc.Case, tool, line string) {
	prependFile(c, c.ClonePath("repo/tools/"+tool+"/tools.toml"), line)
}

// appendClone appends an empty line and <lines> to the file <rel> of the
// clone.
func appendClone(c *rc.Case, rel string, lines ...string) {
	harness.AppendFile(c.T, c.ClonePath(rel), harness.Lines(append([]string{""}, lines...)...))
}

// writeConsumer writes <lines> to the file <rel> of the consumer.
func writeConsumer(c *rc.Case, rel string, lines ...string) {
	harness.WriteFile(c.T, c.Path(rel), harness.Lines(lines...))
}

// mkdir makes the folder <path>.
func mkdir(c *rc.Case, path string) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		c.T.Fatal(err)
	}
}

// git runs git in <dir> and fails the test on an error (_git).
func git(c *rc.Case, dir string, args ...string) string {
	c.T.Helper()
	return harness.Git(c.T, dir, args...)
}

// plainGit runs git in <dir> and gives its output, or "" on an error.
func plainGit(dir string, args ...string) string {
	out, err := harness.GitErr(dir, args...)
	if err != nil {
		return ""
	}
	return out
}

// fileHasLine tells whether a line of the file <path> is <line> (grep -qxF).
func fileHasLine(path, line string) bool {
	for _, l := range strings.Split(harness.ReadFile(path), "\n") {
		if l == line {
			return true
		}
	}
	return false
}

// logRe tells whether the log of the case matches the regular expression <re>
// on a line.
func logRe(c *rc.Case, re string) bool {
	return regexp.MustCompile("(?m)" + re).MatchString(c.LogText())
}

// checkLogLacks <name> <text>: the output of the case must not have the fixed
// string <text> (check_log_lacks).
func checkLogLacks(c *rc.Case, name, text string) {
	c.T.Helper()
	if c.LogHas(text) {
		harness.Fail(c.T, name+": the output has: "+text, c.LogText())
		return
	}
	c.Pass(name + ": no " + text)
}

// synced runs a Call after a sync and a commit of the consumer, and gives the
// code of the failed sync or of the Call.
func synced(then func(c *rc.Case) int) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		return then(c)
	}
}

// engine gives a Call that runs the engine with <args>.
func engine(args ...string) func(c *rc.Case) int {
	return func(c *rc.Case) int { return c.Engine(args...) }
}

// jsonLines: the log has one line.
func jsonLines(c *rc.Case) bool { return strings.Count(c.LogText(), "\n") == 1 }

// logIs: the log, without the newlines at its end, is <s>.
func logIs(c *rc.Case, s string) bool { return strings.TrimRight(c.LogText(), "\n") == s }

// jqTrue tells whether the jq filter <filter> on the file <file> prints true.
func jqTrue(c *rc.Case, filter, file string) bool {
	jq := toolDirs(c.T)
	cmd := &harness.Cmd{Env: c.Env, Args: []string{filepath.Join(jq, "jq"), "-c", filter, file}}
	code, out := cmd.Run()
	return code == 0 && strings.TrimSpace(out) == "true"
}

// jqOut gives the output of the jq filter <filter> on the file <file>.
func jqOut(c *rc.Case, filter, file string) string {
	jq := toolDirs(c.T)
	cmd := &harness.Cmd{Env: c.Env, Args: []string{filepath.Join(jq, "jq"), "-cS", filter, file}}
	_, out := cmd.Run()
	return strings.TrimRight(out, "\n")
}

// lspJq: the jq filter is true for the envelope of the case (lsp_jq).
func lspJq(c *rc.Case, filter string) bool { return jqTrue(c, filter, c.Log) }

// writeScript writes an executable script.
func writeScript(c *rc.Case, path, content string) {
	harness.WriteFile(c.T, path, content)
	if err := os.Chmod(path, 0o755); err != nil {
		c.T.Fatal(err)
	}
}

// isExec tells whether <path> is a file that can run (test -x).
func isExec(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().Perm()&0o111 != 0
}

// isFile tells whether <path> is a regular file (test -f).
func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}

// isDir tells whether <path> is a folder (test -d).
func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
