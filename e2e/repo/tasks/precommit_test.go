package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

var (
	pcOnce sync.Once
	pcDir  string
	pcErr  error
)

// pcTool gives the folder of pre-commit: the one on PATH, else the one of
// `mise x pre-commit@latest`.
func pcTool(t *testing.T) string {
	t.Helper()
	pcOnce.Do(func() {
		path := harness.GetEnv(harness.Env(), "PATH")
		for _, dir := range filepath.SplitList(path) {
			if info, err := os.Stat(filepath.Join(dir, "pre-commit")); err == nil && !info.IsDir() {
				pcDir = dir
				return
			}
		}
		cmd := exec.Command(harness.RealMise, "x", "pre-commit@latest", "--", "which", "pre-commit")
		cmd.Dir = harness.Root
		cmd.Env = harness.Env()
		out, err := cmd.Output()
		if err != nil {
			pcErr = fmt.Errorf("mise x pre-commit@latest -- which pre-commit: %v", err)
			return
		}
		pcDir = filepath.Dir(strings.TrimSpace(string(out)))
	})
	if pcErr != nil {
		t.Fatal(pcErr)
	}
	return pcDir
}

// pcConfig is a pre-commit config with one local hook. The hook prints each
// file that it gets, one on each line, and needs no network.
const pcConfig = `repos:
  - repo: local
    hooks:
      - id: show
        name: show the files
        language: system
        entry: printf 'got %s\n'
        types: []
        verbose: true
`

// pcRepo is a scratch repository with the config and same.txt committed.
func pcRepo(t *testing.T) (string, []string) {
	t.Helper()
	base := harness.Dir(t)
	root := filepath.Join(base, "repo")
	copyFile(t, helpersPath(), filepath.Join(root, ".config", "mise", "tasks", "_scripts", "helpers.py"))
	harness.WriteFile(t, filepath.Join(root, ".config", "pre-commit-config.yaml"), pcConfig)
	harness.WriteFile(t, filepath.Join(root, "same.txt"), "x\n")
	harness.Git(t, root, "init", "--quiet", "-b", "main")
	harness.Git(t, root, "add", "-A")
	harness.Git(t, root, "commit", "--quiet", "--no-verify", "-m", "files")
	// git without the global and the system config, and without the default
	// excludes file, so an ignore file of the user never hides a file of a
	// case.
	env := harness.SetEnv(harness.Env(), "MISE_PROJECT_ROOT="+root,
		"PATH="+pcTool(t)+string(os.PathListSeparator)+harness.GetEnv(harness.Env(), "PATH"),
		"PRE_COMMIT_HOME="+filepath.Join(base, "pre-commit-home"),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "XDG_CONFIG_HOME="+filepath.Join(base, "no-config"))
	return root, env
}

// pcRun runs the working tree's code:precommit in <root>.
func pcRun(t *testing.T, root string, env []string) (int, string) {
	t.Helper()
	task := filepath.Join(harness.Root, "repo", "tools", "pre-commit", "tasks", "code-precommit.py")
	cmd := &harness.Cmd{Dir: root, Env: env, Args: pyTask(t, task)}
	return cmd.Run()
}

// pcCheck compares the exit code, the lines of <want> and the hook lines of
// the output. <files> are the files that the hook got, sorted: pre-commit
// can change the order of the files.
func pcCheck(t *testing.T, name string, got int, log string, want []string, files []string) {
	t.Helper()
	var hook []string
	for _, line := range strings.Split(log, "\n") {
		if f, ok := strings.CutPrefix(strings.TrimSpace(line), "got "); ok {
			hook = append(hook, f)
		}
	}
	slices.Sort(hook)
	bad := ""
	switch {
	case got != 0:
		bad = fmt.Sprintf("exit %d, expected 0", got)
	case strings.Join(hook, " ") != strings.Join(files, " "):
		bad = fmt.Sprintf("the hook got '%s', expected '%s'", strings.Join(hook, " "), strings.Join(files, " "))
	default:
		for _, s := range want {
			if !strings.Contains(log, s) {
				bad = "the output lacks: " + s
				break
			}
		}
	}
	if bad == "" {
		harness.Pass(t, "code:precommit, "+name)
		return
	}
	harness.Fail(t, "code:precommit, "+name+": "+bad, log)
}

// TestPrecommit: the task code:precommit. Without --all, it runs the hooks
// on the uncommitted files only; with --all, on each tracked file. With no
// uncommitted file, it says so and names --all.
func TestPrecommit(t *testing.T) {
	hint := "Run `mise run code:precommit --all` to run the hooks on every file"

	t.Run("no uncommitted file", func(t *testing.T) {
		t.Parallel()
		root, env := pcRepo(t)
		got, log := pcRun(t, root, env)
		pcCheck(t, "no uncommitted file: a message and the --all hint", got, log,
			[]string{"No uncommitted files, so no hook ran", hint}, nil)
	})

	t.Run("uncommitted files", func(t *testing.T) {
		t.Parallel()
		root, env := pcRepo(t)
		for _, f := range []string{"staged.txt", "changed.txt", "deleted.txt", "removed.txt"} {
			harness.WriteFile(t, filepath.Join(root, f), "x\n")
		}
		harness.Git(t, root, "add", "-A")
		harness.Git(t, root, "commit", "--quiet", "--no-verify", "-m", "more")
		harness.WriteFile(t, filepath.Join(root, "staged.txt"), "x\ny\n")
		harness.Git(t, root, "add", "staged.txt")
		harness.WriteFile(t, filepath.Join(root, "changed.txt"), "x\ny\n")
		_ = os.Remove(filepath.Join(root, "deleted.txt"))
		harness.Git(t, root, "rm", "--quiet", "removed.txt")
		harness.WriteFile(t, filepath.Join(root, "new.txt"), "x\n")
		harness.WriteFile(t, filepath.Join(root, ".gitignore"), "ignored.txt\n")
		harness.WriteFile(t, filepath.Join(root, "ignored.txt"), "x\n")
		if err := os.Symlink("missing.txt", filepath.Join(root, "link.txt")); err != nil {
			t.Fatal(err)
		}
		got, log := pcRun(t, root, env)
		pcCheck(t, "the staged, changed and new files, and a symlink; no deleted or ignored file", got, log,
			[]string{"Running the hooks on 5 uncommitted file(s)", hint},
			[]string{".gitignore", "changed.txt", "link.txt", "new.txt", "staged.txt"})
	})

	t.Run("--all", func(t *testing.T) {
		t.Parallel()
		root, env := pcRepo(t)
		harness.WriteFile(t, filepath.Join(root, "new.txt"), "x\n")
		harness.Git(t, root, "add", "new.txt")
		got, log := pcRun(t, root, harness.SetEnv(env, "usage_all=true"))
		pcCheck(t, "--all: each tracked file", got, log,
			[]string{"Running the hooks on every file"},
			[]string{".config/mise/tasks/_scripts/helpers.py", ".config/pre-commit-config.yaml", "new.txt", "same.txt"})
	})
}
