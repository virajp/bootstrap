package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

// partsRoot is a scratch repository root with the working tree's helpers.py and
// the executable task files <files> (paths below .config/mise/tasks/, each
// with the lines <body>).
func partsRoot(t *testing.T, body string, files ...string) string {
	t.Helper()
	root := filepath.Join(harness.Dir(t), "parts-consumer")
	tasks := filepath.Join(root, ".config", "mise", "tasks")
	copyFile(t, helpersPath(), filepath.Join(tasks, "_scripts", "helpers.py"))
	for _, f := range files {
		harness.WriteFile(t, filepath.Join(tasks, f), body)
		if err := os.Chmod(filepath.Join(tasks, f), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// helpersCall imports the helpers.py of <root> and runs <script> with
// python -B.
func helpersCall(t *testing.T, root, script string) (int, string) {
	t.Helper()
	python, _ := gfTools(t)
	env := harness.SetEnv(harness.Env(), "MISE_PROJECT_ROOT="+root)
	src := `import os
import sys

sys.dont_write_bytecode = True
sys.path.insert(0, f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")
import helpers
` + script
	cmd := &harness.Cmd{Env: env, Args: []string{python, "-B", "-c", src}}
	return cmd.Run()
}

// TestPythonParts: helpers.py and the orchestrators find a Python part as
// mise names it: NN-x.py is the task NN-x when no NN-x exists.
func TestPythonParts(t *testing.T) {
	t.Run("own_tasks", func(t *testing.T) {
		t.Parallel()
		root := partsRoot(t, "x\n", "code/lint/70-probe.py", "code/lint/60-both", "code/lint/60-both.py",
			"code/lint/80-shell")
		got, log := helpersCall(t, root, "for task in helpers.own_tasks('code:lint'):\n  print(task)\n")
		want := harness.Lines("code:lint:60-both", "code:lint:60-both.py", "code:lint:70-probe", "code:lint:80-shell")
		if got == 0 && log == want {
			harness.Pass(t, "own_tasks: 70-probe.py is code:lint:70-probe, 60-both.py keeps .py")
		} else {
			harness.Fail(t, fmt.Sprintf("own_tasks: exit %d, or a wrong list", got), log)
		}
	})
	t.Run("has_own_task", func(t *testing.T) {
		t.Parallel()
		root := partsRoot(t, "x\n", "setup/ai.py", "setup/external/_default.py")
		script := `for task in ("setup:ai", "setup:external", "setup:none"):
  print(task, "yes" if helpers.has_own_task(task) else "no")
`
		got, log := helpersCall(t, root, script)
		want := harness.Lines("setup:ai yes", "setup:external yes", "setup:none no")
		if got == 0 && log == want {
			harness.Pass(t, "has_own_task: finds x.py and x/_default.py")
		} else {
			harness.Fail(t, fmt.Sprintf("has_own_task: exit %d, or a wrong answer", got), log)
		}
	})
	t.Run("code:lint --all", func(t *testing.T) {
		t.Parallel()
		root := partsRoot(t, "", "code/lint/70-probe.py", "code/lint/80-old")
		tasks := filepath.Join(root, ".config", "mise", "tasks")
		harness.WriteFile(t, filepath.Join(tasks, "code", "lint", "70-probe.py"),
			harness.Lines("#!/usr/bin/env -S uv run --script", `# USAGE flag "--all"`))
		harness.WriteFile(t, filepath.Join(tasks, "code", "lint", "80-old"), harness.Lines("#!/usr/bin/env bash"))
		dir := filepath.Dir(root)
		task := filepath.Join(dir, "code-lint.py")
		gomplate(t, miseTask("code-lint.py.tmpl"), task, "")
		// A fake mise first on PATH: it writes the arguments of each call.
		bin := filepath.Join(dir, "fake-bin")
		calls := filepath.Join(dir, "calls")
		harness.WriteFile(t, filepath.Join(bin, "mise"),
			harness.Lines("#!/usr/bin/env bash", `printf '%s\n' "$*" >> "`+calls+`"`))
		if err := os.Chmod(filepath.Join(bin, "mise"), 0o755); err != nil {
			t.Fatal(err)
		}
		env := harness.Env()
		env = harness.SetEnv(env, "usage_files", "usage_fix", "usage_debug", "MISE_PROJECT_ROOT="+root,
			"usage_all=true", "PATH="+bin+string(os.PathListSeparator)+harness.GetEnv(env, "PATH"))
		cmd := &harness.Cmd{Env: env, Args: pyTask(t, task)}
		got, log := cmd.Run()
		want := harness.Lines("run code:lint:70-probe --all", "run code:lint:80-old")
		ran := harness.ReadFile(calls)
		if got == 0 && ran == want {
			harness.Pass(t, "code:lint --all: --all goes to a .py part with # USAGE flag \"--all\" only")
		} else {
			harness.Fail(t, fmt.Sprintf("code:lint --all: exit %d, or wrong calls", got), log+"calls:\n"+ran)
		}
	})
}
