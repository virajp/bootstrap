package migrate

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// bootstrap migrate. The old form of a values file and of its files: init
// migrates them, with consent, and then sync runs.
const oldValues = `environments = [ "dev" ]
commit_scopes = [ "api" ]
worktree_files = [ ".env" ]
extensions = [ "foo.bar" ]

[tools.repo-sync]

[branch.main]
# The protected main branch.
direct_commit = false

[overrides.gitignore]
add = [ "docs/prompts/" ]
remove = [ "*.log" ]`

const msgOldForm = ".config/bootstrap.toml is the old form of the values of the repository - run bootstrap migrate"

// msgMigrateFirst is the line of a sync that runs migrate first.
const msgMigrateFirst = "bootstrap migrate has a change to make: it runs before the sync"

func oldFiles(c *rc.Case) {
	writeLines(c, c.Path(".config/repo-sync/pre-commit.yaml"), "repos:", "  - repo: https://example.com/own-hooks",
		"    rev: v1", "    hooks:", "      - id: own")
}

// step is one command of the engine. The output of a quiet step is lost.
type step struct {
	args  []string
	quiet bool
}

// engineSeq runs the engine commands of <steps> in order, with a commit of
// the consumer between two steps. It stops at the first step that fails.
func engineSeq(steps ...step) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		for i, s := range steps {
			if i > 0 {
				c.CommitConsumer()
			}
			var code int
			if s.quiet {
				cmd := &harness.Cmd{Env: c.Env, Args: append([]string{harness.Bin}, s.args...)}
				code, _ = cmd.Run()
			} else {
				code = c.Engine(s.args...)
			}
			if code != 0 || i == len(steps)-1 {
				return code
			}
		}
		return 0
	}
}

func run(args ...string) step   { return step{args: args} }
func quiet(args ...string) step { return step{args: args, quiet: true} }

var (
	initYes      = engineSeq(run("migrate", "--yes"))
	initNoTui    = engineSeq(run("migrate", "--no-tui", "--yes"))
	syncOldForm  = engineSeq(run("sync"))
	checkOldForm = engineSeq(run("check"))
)

// initNew removes the values folder, commits, and runs migrate --yes.
func initNew(c *rc.Case) int {
	if err := os.RemoveAll(c.Values); err != nil {
		c.T.Fatal(err)
	}
	c.CommitConsumer()
	return c.Engine("migrate", "--yes")
}

func checkValuesUnchanged(c *rc.Case) {
	c.CheckThat("the old form stays", reIn(c.OldFile, "^commit_scopes"))
	c.CheckThat("no values folder", !exists(c.RepoValues))
	c.CheckThat("the old hook file stays", exists(c.Path(".config/repo-sync/pre-commit.yaml")))
}

func checkInitMigrated(c *rc.Case) {
	rv := c.RepoValues
	c.CheckThat("init: the old values file is gone", rc.NoFile(c.OldFile))
	c.CheckThat("init: repo.toml has format = 3", c.TomlQ(".format == 3", rv))
	c.CheckThat("init: tools.toml has format = 3", c.TomlQ(".format == 3", ""))
	c.CheckThat("init: commit_scopes is [commits] scopes", c.TomlQ(`.commits.scopes[0] == "api"`, rv))
	c.CheckThat("init: [branch.main] is [branches.main]", c.TomlQ(`.branches.main.landing == "direct-merge"`, rv))
	c.CheckThat("init: core_branches", c.TomlQ(`(.core_branches | join(",")) == "main"`, rv))
	c.CheckThat("init: worktree_files is [worktree] files", c.TomlQ(`.worktree.files[0] == ".env"`, rv))
	c.CheckThat("init: no [branch.main] in tools.toml", notGrep(c.ToolsValues, "branch"))
	c.CheckThat("init: no [tools.repo-sync]", notGrep(c.ToolsValues, "repo-sync"))
	c.CheckThat("init: no [overrides.gitignore]", notGrep(c.ToolsValues, "overrides"))
	c.CheckThat("init: the remove line of .gitignore", c.TomlQ(`.tools.git.files[".gitignore"]["*.log"]["$mode"] == "remove"`, ""))
	c.CheckThat("init: the add line in .gitignore", lineIn(c.Path(".gitignore"), "docs/prompts/"))
	c.CheckThat("init: the extension", rc.FileHas(c.Path(".vscode/extensions.json"), `"foo.bar"`))
	c.CheckThat("init: the own hook", rc.FileHas(c.Path(".config/pre-commit-config.yaml"), "https://example.com/own-hooks"))
	c.CheckThat("init: the old hook file is gone", !exists(c.Path(".config/repo-sync/pre-commit.yaml")))
	c.CheckThat("init: bootstrap_commit", reIn(rv, `^bootstrap_commit = "[0-9a-f]{40}"`))
	c.CheckThat("init: the manifest in the folder", isFile(filepath.Join(c.Values, "repo-generated.tsv")))
	c.CheckThat("init: the sync after init", c.LogHas("Synced: "))
	c.CheckThat("init: the tables of the members of base", c.TomlQ(`.tools | (has("dprint") and has("gitleaks"))`, ""))
	c.CheckThat("init: no table of a member of base that is not required", c.TomlQ(`.tools | has("vscode") | not`, ""))
	c.CheckThat("init: no table for _security_base", c.TomlQ(`.tools | has("_security_base") | not`, ""))
}

func TestMigrateOldForm(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "init migrates the old form, then sync runs", NoFormat: true, Values: oldValues,
		Change: oldFiles, Call: engineSeq(run("migrate", "--yes"), run("sync")), Check: checkInitMigrated})
	rc.Run(t, rc.Spec{Name: "init without --yes and without a terminal", Code: 2, NoFormat: true, Values: oldValues,
		Change: oldFiles, Call: engineSeq(run("migrate")), Message: "changes need --yes", Check: checkValuesUnchanged})
	rc.Run(t, rc.Spec{Name: "init --dry-run", NoFormat: true, Values: oldValues, Change: oldFiles,
		Call: engineSeq(run("migrate", "--dry-run")), Message: "changes, nothing written", Check: func(c *rc.Case) {
			checkValuesUnchanged(c)
			c.CheckThat("init --dry-run: the values file in the list", reIn(c.Log, `consumer/\.config/bootstrap\.toml +moved to \.config/bootstrap/`))
			c.CheckThat("init --dry-run: the old hook file in the list", reIn(c.Log, `consumer/\.config/repo-sync/pre-commit\.yaml +its hooks move`))
		}})
	rc.Run(t, rc.Spec{Name: "sync on the old form: migrate first", NoFormat: true, Values: oldValues,
		Change: oldFiles, Call: syncOldForm, Message: msgMigrateFirst, Check: checkInitMigrated})
	rc.Run(t, rc.Spec{Name: "check on the old form: run bootstrap migrate", Code: 2, NoFormat: true, Values: oldValues,
		Change: oldFiles, Call: checkOldForm, Message: msgOldForm, Check: checkValuesUnchanged})
	rc.Run(t, rc.Spec{Name: "sync on the old form with format = 2: migrate first", OldValues: true,
		Values: devValues, Call: syncOldForm, Message: msgMigrateFirst})
	rc.Run(t, rc.Spec{Name: "check on the old form with format = 2: run bootstrap migrate", Code: 2, OldValues: true,
		Values: devValues, Call: checkOldForm, Message: msgOldForm})
	rc.Run(t, rc.Spec{Name: "init stops on a values file with changes that are not committed", Code: 2, NoFormat: true,
		Values: oldValues, Change: oldFiles, Call: func(c *rc.Case) int {
			appendLines(c, c.OldFile, "# a change that is not committed")
			return c.Engine("migrate", "--yes")
		}, Message: ".config/bootstrap.toml has changes that are not committed", Check: checkValuesUnchanged})
}

// init of the old form with format = 2 (D4, D5, D11, D15, D18): the keys
// that move to repo.toml, the branch tables as [branches.*] with landing, the
// members as [members.*], the overrides file as [overrides.tools.*], and the
// manifest in the folder. An override of a key that moves to repo.toml wins.
// A second init changes nothing.
const movedValues = `environments = [ "dev" ]
[tools.git]
commit_scopes = [ "api", "web" ]
[tools.git.branch.main]
merge_model = "pr"
[tools.git.branch.develop]
direct_commit = true
merge_never = [ "main" ]
[tools.mise]
worktree_files = [ ".env" ]
[member.web]
folder = "projects/web"
type = "folder"`

func movedFiles(c *rc.Case) {
	precommitBothBranches(c)
	memberFolder(c)
	writeLines(c, c.Path(".config/bootstrap.overrides.toml"), "[tools.git]", `commit_scopes = [ "from-overrides" ]`, "",
		`[tools.git.files.".gitignore"]`, `"*.log" = { "$mode" = "remove" }`)
	writeLines(c, c.Path(".config/repo-generated.tsv"), manifestFirst)
}

func checkInitMoved(c *rc.Case) {
	rv := c.RepoValues
	c.CheckThat("init: the old values file is gone", rc.NoFile(c.OldFile))
	c.CheckThat("init: the overrides file is gone", rc.NoFile(c.Path(".config/bootstrap.overrides.toml")))
	c.CheckThat("init: the old manifest is gone", rc.NoFile(c.Path(".config/repo-generated.tsv")))
	c.CheckThat("init: the manifest in the folder", isFile(filepath.Join(c.Values, "repo-generated.tsv")))
	c.CheckThat("init: core_branches, main first", c.TomlQ(`(.core_branches | join(",")) == "main,develop"`, rv))
	c.CheckThat("init: merge_model pr is landing pr", c.TomlQ(`.branches.main.landing == "pr"`, rv))
	c.CheckThat("init: direct_commit is landing direct-commit", c.TomlQ(`.branches.develop.landing == "direct-commit"`, rv))
	c.CheckThat("init: merge_never moves unchanged", c.TomlQ(`.branches.develop.merge_never[0] == "main"`, rv))
	c.CheckThat("init: the override of commit_scopes wins", c.TomlQ(`(.commits.scopes | join(",")) == "from-overrides"`, rv))
	c.CheckThat("init: worktree_files is [worktree] files", c.TomlQ(`.worktree.files[0] == ".env"`, rv))
	c.CheckThat("init: [member.web] is [members.web]",
		c.TomlQ(`.members.web.folder == "projects/web" and .members.web.type == "folder"`, rv))
	c.CheckThat("init: tools.toml has no key of repo.toml", notGrep(c.ToolsValues, "commit_scopes|worktree_files|branch|member"))
	c.CheckThat("init: the overrides file is [overrides.tools.git]",
		c.TomlQ(`.overrides.tools.git.files[".gitignore"]["*.log"]["$mode"] == "remove"`, ""))
	c.CheckThat("init: the overrides of tools.toml have no commit_scopes", c.TomlQ(`.overrides.tools.git | has("commit_scopes") | not`, ""))
}

func checkNoFolder(c *rc.Case) {
	c.CheckThat("a refused migration: no values folder", !exists(c.Values))
	c.CheckThat("a refused migration: the old values file stays", isFile(c.OldFile))
}

func TestMigrateFormat2(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "init migrates the keys of format = 2, then init changes nothing", OldValues: true,
		Values: movedValues, Change: movedFiles, Call: engineSeq(run("migrate", "--yes"), run("migrate")),
		Message: "Nothing to change", Check: checkInitMoved})
	rc.Run(t, rc.Spec{Name: "init: a branch table of a branch that does not exist", Code: 2, OldValues: true,
		Values: `environments = [ "dev" ]
[tools.git.branch.develop]
direct_commit = true`, Call: initYes,
		Message: ".config/bootstrap.toml names a branch that is not main or develop, or that does not exist in git - remove the table or create the branch",
		Check:   checkNoFolder})
	rc.Run(t, rc.Spec{Name: "init: a branch table of a branch that is not main or develop", Code: 2, OldValues: true,
		Values: `environments = [ "dev" ]
[tools.git.branch.feature]
direct_commit = false`, Change: func(c *rc.Case) { git(c, "branch", "feature") }, Call: initYes,
		Message: "remove the table or create the branch", Check: checkNoFolder})
	rc.Run(t, rc.Spec{Name: "init: neither main nor develop", Code: 2, OldValues: true, Values: devValues,
		Change: consumerNoCoreBranch, Call: initYes,
		Message: ".config/bootstrap.toml cannot migrate - create main or develop, then run bootstrap migrate again", Check: checkNoFolder})
	rc.Run(t, rc.Spec{Name: "init: an old member without a type gets type submodule", OldValues: true, Values: `environments = [ "dev" ]
[member.lib]
folder = "sub"`, Change: consumerWithSubmodule, Call: initYes, Message: "Synced: ", Check: func(c *rc.Case) {
		c.CheckThat("init: an old member without a type in .gitmodules is a submodule",
			c.TomlQ(`.members.lib.type == "submodule" and .members.lib.folder == "sub"`, c.RepoValues))
	}})
}

// TestMigrateNew: init in a repository without values: format = 3 and a
// table for each member of base in tools.toml, and the branch of HEAD in
// core_branches.
func TestMigrateNew(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "init without a values file", Values: devValues, Call: initNew, Message: "Synced: ", Check: func(c *rc.Case) {
		rv := c.RepoValues
		c.CheckThat("init: the new repo.toml has format = 3", c.TomlQ(".format == 3", rv))
		c.CheckThat("init: the new tools.toml has format = 3", c.TomlQ(".format == 3", ""))
		c.CheckThat("init: the new repo.toml has core_branches", c.TomlQ(`(.core_branches | join(",")) == "main"`, rv))
		c.CheckThat("init: the new values file has the members of base",
			c.TomlQ(`.tools | (has("mise") and has("dprint") and has("vscode") and has("osv-scanner"))`, ""))
		c.CheckThat("init: no table for _security_base", c.TomlQ(`.tools | has("_security_base") | not`, ""))
		c.CheckThat("init: the sync wrote the mise config", isFile(c.Path(".config/mise.toml")))
	}})
	rc.Run(t, rc.Spec{Name: "init without a values file on a branch that is not main or develop", Values: devValues,
		Change: consumerNoCoreBranch, Call: initNew, Message: "Synced: ", Check: func(c *rc.Case) {
			c.CheckThat("init on trunk: core_branches", c.TomlQ(`(.core_branches | join(",")) == "trunk"`, c.RepoValues))
		}})
	// D93: with neither main nor develop and a detached HEAD, init stops.
	rc.Run(t, rc.Spec{Name: "init without a values file on a detached HEAD", Code: 2, Values: devValues,
		Change: func(c *rc.Case) {
			git(c, "checkout", "--quiet", "--detach")
			git(c, "branch", "-m", "main", "trunk")
		}, Call: initNew,
		Message: "has no branch main or develop, and HEAD is not on a branch - check out a branch, then run bootstrap migrate again"})
	rc.Run(t, rc.Spec{Name: "init creates a missing tools.toml", Values: devValues, Call: func(c *rc.Case) int {
		_ = os.Remove(c.ToolsValues)
		c.CommitConsumer()
		return c.Engine("migrate", "--yes")
	}, Message: "Synced: ", Check: func(c *rc.Case) {
		c.CheckThat("init: tools.toml is there again", c.TomlQ(".format == 3", ""))
	}})
	rc.Run(t, rc.Spec{Name: "init outside a git repository", Code: 2, Values: devValues, Call: func(c *rc.Case) int {
		plain := filepath.Join(c.Dir, "plain")
		mkdir(c, plain)
		c.Env = harness.SetEnv(c.Env, "MISE_PROJECT_ROOT="+plain)
		return c.Engine("migrate", "--yes")
	}, Message: "is not a git repository - create one with git init, then run bootstrap migrate again"})
	rc.Run(t, rc.Spec{Name: "init with nothing to change", Values: devValues, Call: func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		return c.Engine("migrate")
	}, Message: "Nothing to change"})
	rc.Run(t, rc.Spec{Name: "init --answer of an unknown question", Code: 2, Values: devValues,
		Call:    engineSeq(run("migrate", "--yes", "--answer", "nope.key=1")),
		Message: "--answer nope.key=1: no question nope.key of the selected tools"})
	rc.Run(t, rc.Spec{Name: "init with a tool name", Code: 2, Values: devValues, Call: engineSeq(run("migrate", "--yes", "node")),
		Message: "migrate takes no tool names - use: bootstrap add tool node"})
}

// ownTask writes an executable task of the repo of the case below
// .config/mise/tasks/, with <lines> after the shebang.
func ownTask(c *rc.Case, rel string, lines ...string) {
	file := c.Path(tasksDir + "/" + rel)
	writeLines(c, file, append([]string{"#!/usr/bin/env bash"}, lines...)...)
	if err := os.Chmod(file, 0o755); err != nil {
		c.T.Fatal(err)
	}
}

// The mise variables of the task that runs the tests. A mise run in the
// consumer runs without them.
var miseTaskVars = []string{"MISE_ENV", "MISE_ORIGINAL_CWD", "MISE_TASK_NAME", "MISE_TASK_DIR", "MISE_TASK_FILE",
	"MISE_CONFIG_ROOT", "FORCE_COLOR"}

// consumerMise runs `mise run <args>` in the consumer of the case, as a user
// runs it there. The log gets its output.
func consumerMise(c *rc.Case, args ...string) int {
	env := harness.SetEnv(harness.Env(), miseTaskVars...)
	env = harness.SetEnv(env, "MISE_PROJECT_ROOT="+c.Consumer)
	return runToLog(c, env, c.Consumer, append([]string{"mise", "run"}, args...)...)
}

// isDispatcher tells whether the task <port> is the dispatcher of bootstrap.
func isDispatcher(c *rc.Case, port string) bool {
	return rc.FileHas(c.Path(tasksDir+"/"+strings.ReplaceAll(port, ":", "/")+"/_default.py"), `helpers.own_tasks("`+port+`")`)
}

func taskHas(c *rc.Case, rel, text string) bool { return rc.FileHas(c.Path(tasksDir+"/"+rel), text) }

// TestTestPorts: the test ports (D2, D3, D4, D7, D8, D12, D15, D18, D20, D21,
// D26): the dispatchers test, test:e2e and test:golden, their NN- parts, and
// the init move of a repo's own test task to <group>/50-local.
func TestTestPorts(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "the test ports: init writes the three dispatchers", Values: devValues, Call: initNew,
		Message: "Synced: ", Check: func(c *rc.Case) {
			for _, port := range []string{"test", "test:e2e", "test:golden"} {
				c.CheckThat("the test ports: init writes the dispatcher of "+port, isDispatcher(c, port))
				got := consumerMise(c, port)
				c.CheckThat("the test ports: mise run "+port+" with no part exits 0", got == 0)
				c.CheckThat("the test ports: mise run "+port+" with no part prints the line",
					c.LogHas("No "+port+" part in this repository"))
			}
		}})
	rc.Run(t, rc.Spec{Name: "the test ports: the NN- parts of a dispatcher", Values: devValues, Change: func(c *rc.Case) {
		ownTask(c, "test/10-a", "echo part-test-a")
		ownTask(c, "test/b", "echo part-test-b")
		ownTask(c, "test/e2e/10-a", "echo part-e2e-a")
	}, Call: initNew, Message: "Synced: ", Check: func(c *rc.Case) {
		got := consumerMise(c, "test")
		c.CheckThat("the test ports: mise run test exits 0", got == 0)
		c.CheckThat("the test ports: mise run test runs test:10-a", c.LogHas("part-test-a"))
		c.CheckThat("the test ports: mise run test skips test:b, no NN- name", c.LogLacks("part-test-b"))
		c.CheckThat("the test ports: mise run test does not run test:e2e:10-a (D3)", c.LogLacks("part-e2e-a"))
		got = consumerMise(c, "test:e2e")
		c.CheckThat("the test ports: mise run test:e2e exits 0", got == 0)
		c.CheckThat("the test ports: mise run test:e2e runs test:e2e:10-a", c.LogHas("part-e2e-a"))
	}})
	// D21: an unshaped repo, with no values folder, has its own test/_default
	// with a depends on test:go, and a stray test/e2e file.
	rc.Run(t, rc.Spec{Name: "the init move: an unshaped repo with its own test tasks", Code: 2, Values: devValues,
		Change: func(c *rc.Case) {
			if err := os.RemoveAll(c.Values); err != nil {
				c.T.Fatal(err)
			}
			ownTask(c, "test/_default", `#MISE depends=["test:go"]`, "echo own-test-default")
			ownTask(c, "test/go", "echo own-test-go")
			ownTask(c, "test/e2e", "echo own-test-e2e")
		}, Call: checkOldForm, Message: ".config/bootstrap not found - run bootstrap migrate", Check: func(c *rc.Case) {
			got := caseEngine(c, nil, "init", "--yes")
			c.CheckThat("the init move: init exits 0", got == 0)
			c.CheckThat("the init move: init names the move", c.LogHas("move own test tasks into 50-local parts"))
			c.CheckThat("the init move: test/_default is test/50-local", taskHas(c, "test/50-local", "own-test-default"))
			c.CheckThat("the init move: test/e2e is test/e2e/50-local", taskHas(c, "test/e2e/50-local", "own-test-e2e"))
			c.CheckThat("the init move: test/go stays", taskHas(c, "test/go", "own-test-go"))
			c.CheckThat("the init move: the dispatcher of test", isDispatcher(c, "test"))
			c.CheckThat("the init move: the dispatcher of test:e2e", isDispatcher(c, "test:e2e"))
			got = consumerMise(c, "test")
			c.CheckThat("the init move: mise run test exits 0", got == 0)
			c.CheckThat("the init move: mise run test runs the old test as a part", c.LogHas("own-test-default"))
			c.CheckThat("the init move: its depends still runs test:go", c.LogHas("own-test-go"))
			got = caseEngine(c, nil, "check")
			c.CheckThat("the init move: check exits 0", got == 0)
			c.CheckThat("the init move: no warning for test:go (D18)", c.LogLacks("has no NN- name"))
		}})
	// D8, D20: test/50-local exists, so init moves no file, also not
	// test/e2e.
	rc.Run(t, rc.Spec{Name: "the init move: a 50-local that exists stops init", Code: 2, Values: devValues,
		Change: func(c *rc.Case) {
			ownTask(c, "test/_default", "echo own-test-default")
			ownTask(c, "test/50-local", "echo own-test-local")
			ownTask(c, "test/e2e", "echo own-test-e2e")
		}, Call: initYes, Message: "is a task of this repo at the place of a dispatcher of bootstrap, and ",
		Check: func(c *rc.Case) {
			c.CheckThat("a taken 50-local: names test/_default", c.LogHas(tasksDir+"/test/_default is a task of this repo"))
			c.CheckThat("a taken 50-local: names test/50-local", c.LogHas(tasksDir+"/test/50-local exists"))
			c.CheckThat("a taken 50-local: test/_default stays", taskHas(c, "test/_default", "own-test-default"))
			c.CheckThat("a taken 50-local: test/50-local stays", taskHas(c, "test/50-local", "own-test-local"))
			c.CheckThat("a taken 50-local: test/e2e stays a file", isFile(c.Path(tasksDir+"/test/e2e")))
		}})
	// D12 and D15: the repo's own test/golden/_default moves; a file of a
	// dispatcher folder with no NN- name is a warning of check.
	rc.Run(t, rc.Spec{Name: "the init move: test/golden/_default, and a part with no NN- name", Values: devValues,
		Change: func(c *rc.Case) {
			ownTask(c, "test/golden/_default", "echo own-golden")
			ownTask(c, "code/lint/mylint", "echo own-lint")
		}, Call: engineSeq(quiet("migrate", "--yes"), run("check")),
		Message: tasksDir + "/code/lint/mylint has no NN- name, so its dispatcher does not run it - rename it to NN-<name>",
		Check: func(c *rc.Case) {
			c.CheckThat("the init move: test/golden/_default is test/golden/50-local", taskHas(c, "test/golden/50-local", "own-golden"))
			c.CheckThat("the init move: the dispatcher of test:golden", isDispatcher(c, "test:golden"))
		}})
	// D26: a test/_default with the header of bootstrap and no manifest row
	// is bootstrap's own old copy. D22 of
	// docs/plans/2026-11-18-python-tasks-repo: init writes the dispatcher
	// test/_default.py and removes the old copy.
	rc.Run(t, rc.Spec{Name: "the init move: a test/_default with the header is removed", Values: devValues,
		Change: func(c *rc.Case) { headerFile(c, tasksDir+"/test/_default") }, Call: initYes, Check: func(c *rc.Case) {
			c.CheckThat("a synced test/_default: the dispatcher", isDispatcher(c, "test"))
			c.CheckThat("a synced test/_default: no test/50-local", rc.NoFile(c.Path(tasksDir+"/test/50-local")))
			c.CheckThat("a synced test/_default: no old test/_default", rc.NoFile(c.Path(tasksDir+"/test/_default")))
		}})
}

// The test parts of the tools (R2e D4, D5, D16, D17, D19, D25, D26): init
// removes a known old copy of a part, the compose test/golden task of the
// stackgen pack, at claude-plugins commit 0562cb44: its bytes give the
// replaces hash of the android part test/golden/40-android. Without
// CLAUDE_PLUGINS_CLONE, each case of a copy is skipped with one line (D21).
const msgOldCopy = "is an old copy of the task test/golden/40-android.py of the tool android - run bootstrap migrate"

// oldGolden reads the old copy from CLAUDE_PLUGINS_CLONE. It gives nil when
// the variable is empty.
func oldGolden(t *testing.T) []byte {
	clone := os.Getenv("CLAUDE_PLUGINS_CLONE")
	if clone == "" {
		return nil
	}
	cmd := exec.Command("git", "-C", clone, "show",
		"0562cb44:plugins/stackgen/stacks/app-framework/compose/config/.config/mise/tasks/test/golden")
	cmd.Env = harness.Env()
	out, err := cmd.Output()
	if err != nil {
		harness.Fail(t, "the tool test parts: "+clone+" has no compose test/golden at 0562cb44", err.Error())
		return nil
	}
	return out
}

// copyCase runs the case with the old copy, else skips it with one line.
func copyCase(t *testing.T, old []byte, s rc.Spec) {
	t.Helper()
	if old == nil {
		t.Run(s.Name, func(t *testing.T) {
			t.Skip(s.Name + ": skipped - CLAUDE_PLUGINS_CLONE names no claude-plugins clone")
		})
		return
	}
	rc.Run(t, s)
}

// oldCopy writes the old copy, executable, below .config/mise/tasks/ of the
// consumer of the case.
func oldCopy(c *rc.Case, old []byte, rel string) {
	file := c.Path(tasksDir + "/" + rel)
	mkdir(c, filepath.Dir(file))
	if err := os.WriteFile(file, old, 0o755); err != nil {
		c.T.Fatal(err)
	}
}

// addLocalCopy: init, add tool android, then the old copy at
// test/golden/50-local. With <edited>, a line goes at the end of the copy,
// so it is the repo's own.
func addLocalCopy(old []byte, edited bool) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		for _, args := range [][]string{{"migrate", "--yes"}, {"add", "tool", "android", "--yes"}} {
			cmd := &harness.Cmd{Env: c.Env, Args: append([]string{harness.Bin}, args...)}
			if code, _ := cmd.Run(); code != 0 {
				return code
			}
			c.CommitConsumer()
		}
		oldCopy(c, old, "test/golden/50-local")
		if edited {
			appendLines(c, c.Path(tasksDir+"/test/golden/50-local"), "# an edit of the repo")
		}
		c.CommitConsumer()
		return c.Engine("check")
	}
}

func TestToolTestParts(t *testing.T) {
	old := oldGolden(t)
	androidValues := devValues + `
[tools.android]`
	oldAtGroup := func(c *rc.Case) { oldCopy(c, old, "test/golden") }
	// D25, D26: the values select android before the first init; the old
	// copy is at the old path test/golden. check gives code 2 and the fix of
	// init; init removes it, and R1c does not move it.
	copyCase(t, old, rc.Spec{Name: "the tool test parts: an old copy at test/golden", Code: 2, Cmd: "check", Args: noArgs,
		Values: androidValues, Change: oldAtGroup, Message: " - run bootstrap migrate", Check: func(c *rc.Case) {
			got := caseEngine(c, nil, "init", "--yes")
			c.CheckThat("an old copy at test/golden: init exits 0", got == 0)
			c.CheckThat("an old copy at test/golden: init names the removal", c.LogHas("remove old copies of the android test parts"))
			c.CheckThat("an old copy at test/golden: the dispatcher of test:golden", isDispatcher(c, "test:golden"))
			c.CheckThat("an old copy at test/golden: not moved to 50-local", rc.NoFile(c.Path(tasksDir+"/test/golden/50-local")))
			c.CheckThat("an old copy at test/golden: the part 40-android", isFile(c.Path(tasksDir+"/test/golden/40-android.py")))
			got = caseEngine(c, nil, "check")
			c.CheckThat("an old copy at test/golden: check after init exits 0", got == 0)
		}})
	// D25: init, add tool android, then the old copy at
	// test/golden/50-local.
	copyCase(t, old, rc.Spec{Name: "the tool test parts: an old copy at test/golden/50-local", Code: 2, Values: devValues,
		Call: addLocalCopy(old, false), Message: msgOldCopy, Check: func(c *rc.Case) {
			got := caseEngine(c, nil, "init", "--yes")
			c.CheckThat("an old copy at test/golden/50-local: init exits 0", got == 0)
			c.CheckThat("an old copy at test/golden/50-local: init names the removal", c.LogHas("remove old copies of the android test parts"))
			c.CheckThat("an old copy at test/golden/50-local: removed", rc.NoFile(c.Path(tasksDir+"/test/golden/50-local")))
			got = caseEngine(c, nil, "check")
			c.CheckThat("an old copy at test/golden/50-local: check after init exits 0", got == 0)
		}})
	// D16: an edited copy is the repo's own: no warning, and init keeps it.
	copyCase(t, old, rc.Spec{Name: "the tool test parts: an edited copy at test/golden/50-local stays", Values: devValues,
		Call: addLocalCopy(old, true), Check: func(c *rc.Case) {
			c.CheckThat("an edited copy: check names no old copy", c.LogLacks("is an old copy"))
			got := caseEngine(c, nil, "init", "--yes")
			c.CheckThat("an edited copy: init exits 0", got == 0)
			c.CheckThat("an edited copy: init removes nothing", c.LogLacks("remove old copies"))
			c.CheckThat("an edited copy: it stays", lineIn(c.Path(tasksDir+"/test/golden/50-local"), "# an edit of the repo"))
		}})
	// D17: without the android tool, R1c moves the old copy to 50-local.
	copyCase(t, old, rc.Spec{Name: "the tool test parts: no android tool, the old copy moves to 50-local", Values: devValues,
		Change: oldAtGroup, Call: initYes, Message: "move own test tasks into 50-local parts", Check: func(c *rc.Case) {
			c.CheckThat("no android tool: init removes nothing", c.LogLacks("remove old copies"))
			b, err := os.ReadFile(c.Path(tasksDir + "/test/golden/50-local"))
			c.CheckThat("no android tool: the copy is test/golden/50-local", err == nil && bytes.Equal(b, old))
		}})
	// D14, D15, D19: with the android and swift tools and no Gradle or Xcode
	// project, the parts print nothing, so stdout is the success line of the
	// dispatcher; stderr is not examined. mise trust and the run keep their
	// state in the case folder, and install no tool.
	rc.Run(t, rc.Spec{Name: "the tool test parts: mise run test:golden with no Gradle or Xcode project", Values: devValues + `
[tools.android]

[tools.swift]`, Call: initYes, Message: "Synced: ", Check: checkGoldenPartsQuiet})
}

func checkGoldenPartsQuiet(c *rc.Case) {
	env := harness.SetEnv(harness.Env(), "FORCE_COLOR")
	// The bytes of print_success "Golden testing done" of the shell helpers
	// of bootstrap 9476e49, with no color: stdout is not a terminal.
	wantOut := []byte("[OK] Golden testing done\n")
	env = harness.SetEnv(env, miseTaskVars...)
	env = harness.SetEnv(env, "MISE_STATE_DIR="+filepath.Join(c.Dir, "mise-state"),
		"MISE_TASK_RUN_AUTO_INSTALL=false", "MISE_PROJECT_ROOT="+c.Consumer)
	trust := exec.Command("mise", "trust", "--all")
	trust.Dir, trust.Env = c.Consumer, env
	_ = trust.Run()
	logf, err := os.Create(c.Log)
	if err != nil {
		c.T.Fatal(err)
	}
	var out bytes.Buffer
	run := exec.Command("mise", "run", "test:golden")
	run.Dir, run.Env, run.Stdout, run.Stderr = c.Consumer, env, &out, logf
	runErr := run.Run()
	_ = logf.Close()
	c.CheckThat("the golden parts: the parts are there",
		isFile(c.Path(tasksDir+"/test/golden/40-android.py")) && isFile(c.Path(tasksDir+"/test/golden/40-swift.py")))
	c.CheckThat("the golden parts: mise run test:golden exits 0", runErr == nil)
	c.CheckThat("the golden parts: stdout is the success line of the dispatcher", bytes.Equal(wantOut, out.Bytes()))
}

// TestMigratePacks: the migration of [packs.*] tables with the frozen packs
// of the engine (D5, D16 b, D18 a). A tool of the pack gets a table with the
// env of the pack; a tool with a table keeps its env; a table of a tool that
// bootstrap no longer has goes; a member of the new pack base without
// required = true gets no table.
func TestMigratePacks(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "init --no-tui migrates the [packs.*] tables", NoFormat: true, Values: `environments = [ "dev", "ci" ]

[packs.container-build]
env = [ "dev" ]

[tools.cosign]
env = [ "dev", "ci" ]

[tools.devtools]`, Call: initNoTui, Message: "Synced: ", Check: func(c *rc.Case) {
		tv := c.ToolsValues
		c.CheckThat("migration: format = 3", c.TomlQ(".format == 3", ""))
		c.CheckThat("migration: no [packs.*] table", notGrep(tv, `^\[packs\.`))
		c.CheckThat("migration: a tool of the pack gets the env of the pack", c.TomlQ(`(.tools.apko.env | join(",")) == "dev"`, ""))
		c.CheckThat("migration: a tool with a table keeps its env", c.TomlQ(`(.tools.cosign.env | join(",")) == "dev,ci"`, ""))
		c.CheckThat("migration: the removed tool goes", notGrep(tv, "devtools"))
		c.CheckThat("migration: the removed tool is listed", c.LogHas("[tools.devtools] removed: bootstrap no longer has the tool"))
		c.CheckThat("migration: a tool of the frozen pack base", c.TomlQ(`.tools | (has("dprint") and has("pre-commit"))`, ""))
		c.CheckThat("migration: no table of a member of base that is not required", notGrep(tv, "vscode"))
		c.CheckThat("migration: no table for _security_base", notGrep(tv, "_security_base"))
	}})
	rc.Run(t, rc.Spec{Name: "init of an old values file with an unknown pack", Code: 3, NoFormat: true, Values: `environments = [ "dev" ]
[packs.nope]`, Call: initNoTui, Message: "nope is not a bootstrap pack (named in "})
	rc.Run(t, rc.Spec{Name: "init migrates a [packs.*] table of a values file with format = 2", OldValues: true,
		Values: `environments = [ "dev" ]
[packs.container-build]`, Call: initNoTui, Message: "Synced: ", Check: func(c *rc.Case) {
			c.CheckThat("a [packs.*] table with format = 2: the table of apko", c.TomlQ(`.tools | has("apko")`, ""))
			c.CheckThat("a [packs.*] table with format = 2: no [packs.*] table", notGrep(c.ToolsValues, `^\[packs\.`))
			c.CheckThat("a [packs.*] table with format = 2: no table of vscode", notGrep(c.ToolsValues, "vscode"))
		}})
	rc.Run(t, rc.Spec{Name: "init again does not add a removed member of base", Values: devValues, Call: func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		return c.Engine("migrate", "--no-tui", "--yes")
	}, Message: "Nothing to change", Check: func(c *rc.Case) {
		c.CheckThat("init again: no table of the removed vscode", notGrep(c.ToolsValues, "vscode"))
	}})
	rc.Run(t, rc.Spec{Name: "init in a submodule: no table of a root_only member", Values: devValues, Call: func(c *rc.Case) int {
		consumerSubmodule(c)
		return initNew(c)
	}, Message: "Synced: ", Check: func(c *rc.Case) {
		c.CheckThat("init in a submodule: no table of vscode", notGrep(c.ToolsValues, "vscode"))
		c.CheckThat("init in a submodule: the table of dprint", c.TomlQ(`.tools | has("dprint")`, ""))
	}})
}

// TestRequiredTables: the sync writes an empty table for each required
// member of base that has no table, and logs it; a dry run and check list it
// and write nothing (D18 b, c).
func TestRequiredTables(t *testing.T) {
	values := devValues + `
[tools.mise]`
	checkListed := func(c *rc.Case) {
		c.CheckThat("a dry run: the values file has no new table", notGrep(c.ToolsValues, "dprint"))
		c.CheckThat("a dry run: a member with a table is not listed", c.LogLacks("[tools.mise] would be added"))
	}
	rc.Run(t, rc.Spec{Name: "sync writes the table of a required member of base", Values: values, Call: syncOldForm,
		Message: "bootstrap: [tools.dprint] added to ", Check: func(c *rc.Case) {
			c.CheckThat("sync: the tables of the required members",
				c.TomlQ(`.tools | (has("mise") and has("dprint") and has("linter") and has("osv-scanner"))`, ""))
			c.CheckThat("sync: no table of a member that is not required", notGrep(c.ToolsValues, "vscode"))
			c.CheckThat("sync: no table of a dependency", notGrep(c.ToolsValues, "_security_base"))
		}})
	rc.Run(t, rc.Spec{Name: "sync --dry-run lists the table of a required member and writes nothing", Values: values,
		Message: "[tools.dprint] would be added to ", Check: checkListed})
	rc.Check(t, rc.Spec{Name: "check lists the table of a required member and writes nothing", Values: values,
		Message: "[tools.dprint] would be added to ", Check: checkListed})
}

// TestAddFail: a sync that fails after add tool puts the values file back
// (D9, D26). A folder that the sync cannot write makes the sync fail.
func TestAddFail(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "a sync that fails after add tool puts the values file back", Code: 1, Values: devValues,
		Call: func(c *rc.Case) int {
			if code := c.Synced(); code != 0 {
				return code
			}
			conf := c.Path(".config/mise/conf.d")
			if err := os.Chmod(conf, 0o555); err != nil {
				c.T.Fatal(err)
			}
			code := c.Engine("add", "tool", "--yes", "go")
			_ = os.Chmod(conf, 0o755)
			return code
		}, Message: "as it was before the command", Check: func(c *rc.Case) {
			c.CheckThat("a failed sync: the values file does not select go", notGrep(c.ToolsValues, `go\]`))
			c.CheckThat("a failed sync: the info line names the values file", c.LogHas("Put back "+c.ToolsValues+" as it was"))
		}})
}
