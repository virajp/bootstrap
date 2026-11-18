// Package pylib compares the Python libraries of the tasks with the shell
// libraries: helpers.py, checks.py, placeholder.py and worktrees_page.py of
// the mise tool, merge.py of the git tool and xcode.py of the swift tool.
// Each case copies the libraries of the working tree's repo/ into a scratch
// .config/mise/tasks/_scripts/, then runs a small Python probe that imports a
// module, and, where a shell library has the same function, a shell probe
// that sources it. The case compares the output bytes, the exit code and the
// return values of the two sides.
package pylib

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
)

func TestMain(m *testing.M) { harness.Main(m, nil, "repo") }

// libs are the libraries that a case copies, with the folder of their tool.
var libs = map[string]string{
	"helpers":           "mise",
	"helpers.py":        "mise",
	"checks":            "mise",
	"checks.py":         "mise",
	"placeholder":       "mise",
	"placeholder.py":    "mise",
	"worktrees-page":    "mise",
	"worktrees_page.py": "mise",
	"gate_files.py":     "mise",
	"merge":             "git",
	"merge.py":          "git",
	"xcode":             "swift",
	"xcode.py":          "swift",
}

// libPath gives the path of the library <name> in the working tree's repo/.
func libPath(name string) string {
	return filepath.Join(harness.Root, "repo", "tools", libs[name], "tasks", name)
}

// scriptsDir gives the _scripts folder of the project root <root>.
func scriptsDir(root string) string {
	return filepath.Join(root, ".config", "mise", "tasks", "_scripts")
}

// installLibs copies each library into the _scripts folder of <root>, with
// mode 755.
func installLibs(t testing.TB, root string) {
	t.Helper()
	for name := range libs {
		b, err := os.ReadFile(libPath(name))
		if err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(scriptsDir(root), name)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

var (
	pyOnce sync.Once
	pyArgv []string
	pyErr  error
)

// python gives the command that runs python 3.14: python3.14 of PATH when
// the harness gives it, else uv run --quiet --no-project --python 3.14.
func python(t testing.TB) []string {
	t.Helper()
	pyOnce.Do(func() {
		if p, err := exec.LookPath("python3.14"); err == nil {
			pyArgv = []string{p}
			return
		}
		uv, err := exec.LookPath("uv")
		if err != nil {
			pyErr = fmt.Errorf("no python3.14 and no uv on the PATH: %v", err)
			return
		}
		pyArgv = []string{uv, "run", "--quiet", "--no-project", "--python", "3.14", "python"}
	})
	if pyErr != nil {
		t.Fatal(pyErr)
	}
	return pyArgv
}

// bash5 gives the bash of PATH, which `#!/usr/bin/env bash` runs (D32).
func bash5(t testing.TB) string {
	t.Helper()
	p, err := exec.LookPath("bash")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// bash32 is the bash of macOS, which the page comparison follows (D31).
const bash32 = "/bin/bash"

// result is the exit code and the output of one probe.
type result struct {
	code        int
	out, errOut string
}

func (r result) String() string {
	return fmt.Sprintf("exit %d\n--- stdout\n%s--- stderr\n%s", r.code, r.out, r.errOut)
}

// probe is one run of a probe file.
type probe struct {
	dir   string   // the working folder
	env   []string // the environment
	merge bool     // stdout and stderr to one pipe, in the order of the writes
}

// run runs <args> and gives its result.
func (p probe) run(t testing.TB, args ...string) result {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = p.dir
	cmd.Env = p.env
	cmd.Stdin = strings.NewReader("")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	if p.merge {
		cmd.Stderr = &out
	} else {
		cmd.Stderr = &errOut
	}
	err := cmd.Run()
	r := result{out: out.String(), errOut: errOut.String()}
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("%s: %v", strings.Join(args, " "), err)
		}
		r.code = ee.ExitCode()
	}
	return r
}

// pyPreamble is the import preamble of a task (D5, D21).
const pyPreamble = `import json
import os
import sys

sys.dont_write_bytecode = True
sys.path.insert(0, f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")
import helpers  # noqa: E402


def mark(name):
  for stream in (sys.stdout, sys.stderr):
    stream.write(f"@@ {name}\n")
    stream.flush()


`

// shPreamble sources helpers, as a task does.
const shPreamble = `source "${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/helpers"
mark() { printf '@@ %s\n' "$1"; printf '@@ %s\n' "$1" >&2; }
`

// py writes <body> after the preamble to a probe file below <scratch> and
// runs it with python -B.
func (p probe) py(t testing.TB, scratch, body string) result {
	t.Helper()
	f, err := os.CreateTemp(scratch, "probe-*.py")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(pyPreamble + body); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return p.run(t, append(append([]string(nil), python(t)...), "-B", f.Name())...)
}

// sh writes <body> after the preamble to a probe file below <scratch> and
// runs it with the bash <bash>.
func (p probe) sh(t testing.TB, bash, scratch, body string) result {
	t.Helper()
	f, err := os.CreateTemp(scratch, "probe-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(shPreamble + body); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return p.run(t, bash, f.Name())
}

// env gives the environment of a probe: the case environment with the
// project root <root>, no color variables, and git without the config of the
// user, so a setting of the user never changes a case.
func env(base, root string, kv ...string) []string {
	e := harness.SetEnv(harness.Env(), "MISE_PROJECT_ROOT="+root, "FORCE_COLOR", "NO_COLOR", "TERM=dumb",
		"BOOTSTRAP_BIN", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"XDG_CONFIG_HOME="+filepath.Join(base, "no-config"), "MISE_ORIGINAL_CWD", "MISE_TASK_NAME", "SKIP")
	return harness.SetEnv(e, kv...)
}

// newCase gives the folder of a case and its project root, with the
// libraries in the _scripts folder of the root.
func newCase(t testing.TB) (string, string) {
	t.Helper()
	dir := harness.Dir(t)
	root := filepath.Join(dir, "root")
	installLibs(t, root)
	return dir, root
}

// same records a check that <py> and <sh> are equal.
func same(t testing.TB, name string, py, sh result) bool {
	t.Helper()
	return harness.Check(t, name, py == sh, "python:\n"+py.String()+"shell:\n"+sh.String())
}

// pyQuote gives <s> as a Python string literal.
func pyQuote(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`, "\n", `\n`).Replace(s) + "'"
}

// shQuote gives <s> as a shell word.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// writeExec writes <content> to <path> with mode 755.
func writeExec(t testing.TB, path, content string) {
	t.Helper()
	harness.WriteFile(t, path, content)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// git runs git in <dir> with <env>, and fails the test on an error.
func git(t testing.TB, e []string, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = e
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimRight(string(out), "\n")
}
