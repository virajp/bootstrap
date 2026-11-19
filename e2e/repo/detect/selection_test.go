package detect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// oldValues and oldFiles are the old form of a values file and of its files.
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

func oldFiles(c *rc.Case) {
	harness.WriteFile(c.T, c.Path(".config/repo-sync/pre-commit.yaml"), harness.Lines(
		"repos:", "  - repo: https://example.com/own-hooks", "    rev: v1", "    hooks:", "      - id: own"))
}

func fileHas(path, s string) bool { return rc.FileHas(path, s) }

func appendClone(c *rc.Case, rel string, lines ...string) {
	harness.AppendFile(c.T, c.ClonePath(rel), harness.Lines(lines...))
}

func goQuestion(c *rc.Case) {
	appendClone(c, "repo/tools/go/tools.toml", "", "[questions.flavor]", `question = "Which flavor?"`, `example = "plain"`)
}

var addGo = engineCall("add", "tool", "--yes", "go")

func checkAddKept(c *rc.Case) {
	c.CheckThat("add: the values file stays", notGrep(c.ToolsValues, "go"))
}

func TestAdd(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "add go", Values: devValues + "\n[tools.vscode]", Call: addGo, Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add: the values file selects go", c.TomlQ(`.tools | has("go")`, ""))
			c.CheckThat("add: the go task", exists(c.Path(".config/mise/tasks/code/lint/45-gomod.py")))
			c.CheckThat("add: the go extension", fileHas(c.Path(".vscode/extensions.json"), `"golang.go"`))
			c.CheckThat("add: one clone line", c.LogCount("[INFO] bootstrap ", 1))
		}})
	rc.Run(t, rc.Spec{Name: "add writes the dependencies", Values: devValues, Call: engineCall("add", "tool", "--yes", "flutter"),
		Message: "select java", Check: func(c *rc.Case) {
			c.CheckThat("add: flutter", c.TomlQ(`.tools | has("flutter")`, ""))
			c.CheckThat("add: its dependency android", c.TomlQ(`.tools | has("android")`, ""))
			c.CheckThat("add: the dependency of android", c.TomlQ(`.tools | has("java")`, ""))
		}})
	rc.Run(t, rc.Spec{Name: "add without --yes and without a terminal", Code: 2, Values: devValues,
		Call: engineCall("add", "tool", "go"), Message: "changes need --yes", Check: checkAddKept})
	rc.Run(t, rc.Spec{Name: "add --dry-run", Values: devValues, Call: engineCall("add", "tool", "--dry-run", "go"),
		Message: "changes, nothing written", Check: checkAddKept})
	rc.Run(t, rc.Spec{Name: "add tool without a tool", Code: 2, Values: devValues, Call: engineCall("add", "tool"),
		Message: "add tool needs at least one tool"})
	rc.Run(t, rc.Spec{Name: "add of an unknown tool", Code: 3, Values: devValues, Call: engineCall("add", "tool", "--yes", "nope"),
		Message: "nope is not a bootstrap tool"})
	rc.Run(t, rc.Spec{Name: "add of a selected tool", Values: devValues, Call: engineCall("add", "tool", "--yes", "dprint"),
		Message: "Nothing to change"})
	rc.Run(t, rc.Spec{Name: "add stops on a values file with changes that are not committed", Code: 2, Values: devValues,
		Call: func(c *rc.Case) int {
			if code := c.Synced(); code != 0 {
				return code
			}
			harness.AppendFile(c.T, c.ToolsValues, "# a change that is not committed\n")
			return c.Engine("add", "tool", "--yes", "go")
		}, Message: ".config/bootstrap/tools.toml has changes that are not committed"})
	rc.Run(t, rc.Spec{Name: "add --yes: a question without an answer", Code: 2, Values: devValues, Change: goQuestion,
		Call: addGo, Message: "the question go.flavor has no answer - give it with --answer go.flavor=<value>"})
	rc.Run(t, rc.Spec{Name: "add --answer", Values: devValues, Change: goQuestion,
		Call: engineCall("add", "tool", "--yes", "--answer", "go.flavor=sweet", "go"), Message: "answer go.flavor",
		Check: func(c *rc.Case) { c.CheckThat("add: the answer", c.TomlQ(`.tools.go.flavor == "sweet"`, "")) }})
	rc.Run(t, rc.Spec{Name: "add without a values file", Code: 2, Values: devValues, Call: func(c *rc.Case) int {
		if err := os.RemoveAll(c.Values); err != nil {
			c.T.Fatal(err)
		}
		c.CommitConsumer()
		return c.Engine("add", "tool", "--yes", "go")
	}, Message: "is not there - run bootstrap migrate first"})
	rc.Run(t, rc.Spec{Name: "add on the old form: run bootstrap migrate", Code: 2, NoFormat: true, Values: oldValues,
		Change: oldFiles, Call: func(c *rc.Case) int { return c.Engine("add", "tool", "--yes", "go") }, Message: msgOldForm})
}

func removeTool(tool string) func(c *rc.Case) int { return engineCall("remove", "tool", "--yes", tool) }

func TestRemove(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "remove go", Values: devValues + "\n[tools.vscode]\n[tools.go]", Call: removeTool("go"),
		Message: "Synced: ", Check: func(c *rc.Case) {
			c.CheckThat("remove: the values file without go", notGrep(c.ToolsValues, "go]"))
			c.CheckThat("remove: the go task is gone", rc.NoFile(c.Path(".config/mise/tasks/code/lint/45-gomod.py")))
			c.CheckThat("remove: the go extension is gone", !fileHas(c.Path(".vscode/extensions.json"), `"golang.go"`))
		}})
	rc.Run(t, rc.Spec{Name: "remove keeps a seed and lists it", Values: devValues + "\n[tools.pnpm]", Change: pnpmFiles,
		Call: removeTool("pnpm"), Message: "pnpm-workspace.yaml of a removed tool", Check: func(c *rc.Case) {
			c.CheckThat("remove: the seed stays", exists(c.Path("pnpm-workspace.yaml")))
			c.CheckThat("remove: the mise config of pnpm is gone", rc.NoFile(c.Path(".config/mise/conf.d/pnpm/mise.dev.toml")))
		}})
	rc.Run(t, rc.Spec{Name: "remove: the dependencies that no tool needs go",
		Values: devValues + "\n[tools.flutter]\n[tools.android]\n[tools.java]\n[tools.go]", Call: removeTool("flutter"),
		Message: "remove java (a dependency that no tool needs)", Check: func(c *rc.Case) {
			c.CheckThat("remove: android goes", notGrep(c.ToolsValues, "android"))
			c.CheckThat("remove: java goes", notGrep(c.ToolsValues, "java"))
			c.CheckThat("remove: go stays", c.TomlQ(`.tools | has("go")`, ""))
		}})
	rc.Run(t, rc.Spec{Name: "remove: a tool that another selected tool needs", Code: 2,
		Values: devValues + "\n[tools.flutter]\n[tools.android]", Call: removeTool("android"),
		Message: "android is needed by flutter - remove flutter first"})
	rc.Run(t, rc.Spec{Name: "remove: a required tool", Code: 2, Values: devValues, Call: removeTool("mise"),
		Message: "mise is required by the base pack, so remove cannot remove it"})
	rc.Run(t, rc.Spec{Name: "remove: a required scanner of the pack base", Code: 2, Values: devValues, Call: removeTool("gitleaks"),
		Message: "gitleaks is required by the base pack, so remove cannot remove it"})
	rc.Run(t, rc.Spec{Name: "remove: a member of the pack base that is not required", Values: devValues + "\n[tools.vscode]",
		Call: removeTool("vscode"), Message: "Synced: ", Check: func(c *rc.Case) {
			c.CheckThat("remove: the table of vscode is gone", notGrep(c.ToolsValues, "vscode"))
		}})
	rc.Run(t, rc.Spec{Name: "remove: a tool that is not selected", Code: 2, Values: devValues, Call: removeTool("go"),
		Message: "go is not selected"})
	rc.Run(t, rc.Spec{Name: "remove: an unknown tool", Code: 3, Values: devValues, Call: removeTool("nope"),
		Message: "nope is not a bootstrap tool"})
}

// The exec commands of dprint.jsonc: each command is a block, found by its
// exts.
const (
	gofmtCmd   = `"command": "mise x -- gofmt"`
	rustfmtCmd = `"command": "mise x -- rustfmt --edition 2024"`
	ownCmd     = `"command": "own-fmt"`
)

func checkExecBoth(c *rc.Case) {
	d := c.Path("dprint.jsonc")
	c.CheckThat("exec: the gofmt command", fileHas(d, gofmtCmd))
	c.CheckThat("exec: the rustfmt command", fileHas(d, rustfmtCmd))
	c.CheckThat("exec: no cacheKeyFiles", !fileHas(d, "cacheKeyFiles"))
}

func TestExec(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "exec: add go and rust", Values: devValues, Call: engineCall("add", "tool", "--yes", "go", "rust"),
		Message: "Synced: ", Check: checkExecBoth})
	rc.Run(t, rc.Spec{Name: "exec: add rust to a repository with go", Values: devValues + "\n[tools.go]",
		Call: engineCall("add", "tool", "--yes", "rust"), Message: "Synced: ", Check: checkExecBoth})
	rc.Run(t, rc.Spec{Name: "exec: remove rust", Values: devValues + "\n[tools.go]\n[tools.rust]", Call: removeTool("rust"),
		Message: "Synced: ", Check: func(c *rc.Case) {
			c.CheckThat("exec: the gofmt command stays", fileHas(c.Path("dprint.jsonc"), gofmtCmd))
			c.CheckThat("exec: the rustfmt command is gone", !fileHas(c.Path("dprint.jsonc"), rustfmtCmd))
		}})
	rc.Run(t, rc.Spec{Name: "exec: a command of the user stays", Values: devValues + "\n[tools.go]", Call: func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		// sed 's#"commands": \[#...#': the first match of each line.
		p := c.Path("dprint.jsonc")
		lines := strings.Split(harness.ReadFile(p), "\n")
		for i, l := range lines {
			lines[i] = strings.Replace(l, `"commands": [`, `"commands": [ { "command": "own-fmt", "exts": ["own"] },`, 1)
		}
		harness.WriteFile(c.T, p, strings.Join(lines, "\n"))
		c.CommitConsumer()
		return c.Engine("add", "tool", "--yes", "rust")
	}, Message: "Synced: ", Check: func(c *rc.Case) {
		c.CheckThat("exec: the command of the user stays", fileHas(c.Path("dprint.jsonc"), ownCmd))
		checkExecBoth(c)
	}})
	// A block without keys still gets its keys item: tailwindcss of node stays.
	rc.Run(t, rc.Spec{Name: "dprint: an empty block stays", Values: devValues + "\n[tools.node]", Call: engineCall("sync"),
		Message: "Synced: ", Check: func(c *rc.Case) {
			c.CheckThat("dprint: the empty block tailwindcss stays", fileHas(c.Path("dprint.jsonc"), `"tailwindcss": {}`))
		}})
	// The old associations of shuck, in an adopted dprint.jsonc, go at the
	// next sync: shuck formats only the shell files with an extension.
	rc.Run(t, rc.Spec{Name: "dprint: the old associations of shuck go", Values: devValues, Call: func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		// sed 's#"shuck": {#...#': the first match of each line.
		p := c.Path("dprint.jsonc")
		lines := strings.Split(harness.ReadFile(p), "\n")
		for i, l := range lines {
			lines[i] = strings.Replace(l, `"shuck": {`,
				`"shuck": { "associations": [ "**/.config/mise/tasks/**", "!**/*.tmpl" ],`, 1)
		}
		harness.WriteFile(c.T, p, strings.Join(lines, "\n"))
		if !fileHas(p, `"associations"`) {
			return 9
		}
		c.CommitConsumer()
		return c.Engine("sync")
	}, Message: "Synced: ", Check: func(c *rc.Case) {
		c.CheckThat("dprint: the old associations of shuck are gone", !fileHas(c.Path("dprint.jsonc"), `"associations"`))
		c.CheckThat("dprint: the shuck block stays", fileHas(c.Path("dprint.jsonc"), `"shuck": {`))
	}})
}

// fakeAube writes a fake aube in <tmp>/fakebin. It records its calls in
// <tmp>/aube-calls and writes the lock file of its format.
func fakeAube(c *rc.Case) {
	p := filepath.Join(c.Dir, "fakebin", "aube")
	harness.WriteFile(c.T, p, `#!/usr/bin/env bash
printf '%s\n' "$*" >> "${0%/*}/../aube-calls"
case $1 in
  import) printf '%s\n' 'lockfileVersion: aube' > aube-lock.yaml ;;
  install) printf '%s\n' "lockfileVersion: '9.0'" > "${AUBE_DEFAULT_LOCKFILE:-aube-lock.yaml}" ;;
esac
`)
	if err := os.Chmod(p, 0o755); err != nil {
		c.T.Fatal(err)
	}
}

func pnpmLockFiles(c *rc.Case) {
	pnpmFiles(c)
	harness.WriteFile(c.T, c.Path("pnpm-lock.yaml"), "lockfileVersion: '9.0'\n")
	fakeAube(c)
}

func aubeLockFiles(c *rc.Case) {
	aubeFiles(c)
	harness.WriteFile(c.T, c.Path("aube-lock.yaml"), "lockfileVersion: aube\n")
	fakeAube(c)
}

// replaceTools syncs, then runs `bootstrap replace tool --yes` with the fake
// aube first on PATH.
func replaceTools(tools ...string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		c.Env = harness.SetEnv(c.Env, "PATH="+filepath.Join(c.Dir, "fakebin")+string(os.PathListSeparator)+harness.GetEnv(c.Env, "PATH"))
		return c.Engine(append([]string{"replace", "tool", "--yes"}, tools...)...)
	}
}

func hasLine(path, line string) bool {
	for _, l := range strings.Split(harness.ReadFile(path), "\n") {
		if l == line {
			return true
		}
	}
	return false
}

func TestReplace(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "replace pnpm aube", Values: devValues + "\n[tools.pnpm]\nenv = [ \"dev\" ]", Change: pnpmLockFiles,
		Call: replaceTools("pnpm", "aube"), Message: "Ran repo/tools/aube/scripts/migrate-pnpm", Check: func(c *rc.Case) {
			c.CheckThat("replace: the values file selects aube", c.TomlQ(`.tools | has("aube") and (has("pnpm") | not)`, ""))
			c.CheckThat("replace: aube gets the env of pnpm", c.TomlQ(`.tools.aube.env[0] == "dev"`, ""))
			c.CheckThat("replace: aube-workspace.yaml", exists(c.Path("aube-workspace.yaml")))
			c.CheckThat("replace: no pnpm-workspace.yaml", rc.NoFile(c.Path("pnpm-workspace.yaml")))
			c.CheckThat("replace: aube-lock.yaml", exists(c.Path("aube-lock.yaml")))
			c.CheckThat("replace: no pnpm-lock.yaml", rc.NoFile(c.Path("pnpm-lock.yaml")))
			c.CheckThat("replace: aube import ran", hasLine(filepath.Join(c.Dir, "aube-calls"), "import --lockfile-only"))
			c.CheckThat("replace: the mise config of aube", exists(c.Path(".config/mise/conf.d/aube/mise.dev.toml")))
			c.CheckThat("replace: no mise config of pnpm", rc.NoFile(c.Path(".config/mise/conf.d/pnpm/mise.dev.toml")))
		}})
	rc.Run(t, rc.Spec{Name: "replace aube pnpm", Values: devValues + "\n[tools.aube]", Change: aubeLockFiles,
		Call: replaceTools("aube", "pnpm"), Message: "review its diff", Check: func(c *rc.Case) {
			c.CheckThat("replace: pnpm-workspace.yaml", exists(c.Path("pnpm-workspace.yaml")))
			c.CheckThat("replace: pnpm-lock.yaml", exists(c.Path("pnpm-lock.yaml")))
			c.CheckThat("replace: no aube-lock.yaml", rc.NoFile(c.Path("aube-lock.yaml")))
		}})
	rc.Run(t, rc.Spec{Name: "replace --dry-run", Values: devValues + "\n[tools.pnpm]", Change: pnpmFiles,
		Call: engineCall("replace", "tool", "--dry-run", "pnpm", "aube"), Message: "run     repo/tools/aube/scripts/migrate-pnpm",
		Check: func(c *rc.Case) {
			c.CheckThat("replace --dry-run: pnpm-workspace.yaml stays", exists(c.Path("pnpm-workspace.yaml")))
			c.CheckThat("replace --dry-run: the values file stays", c.TomlQ(`.tools | has("pnpm")`, ""))
		}})
	rc.Run(t, rc.Spec{Name: "replace: tools without the same parent", Code: 2, Values: devValues + "\n[tools.pnpm]",
		Call:    engineCall("replace", "tool", "--yes", "pnpm", "go"),
		Message: "works only between two tools with the same parent (pnpm has the parent 'node', go has '')"})
	rc.Run(t, rc.Spec{Name: "replace: no migrate script", Code: 2, Values: devValues + "\n[tools.pnpm]",
		Change: func(c *rc.Case) {
			harness.WriteFile(c.T, c.ClonePath("repo/tools/yarn/tools.toml"),
				harness.Lines(`default_envs = [ "dev" ]`, `depends = [ "node" ]`, `parent = "node"`, `root_only = false`))
		}, Call: engineCall("replace", "tool", "--yes", "pnpm", "yarn"),
		Message: "yarn has no script repo/tools/yarn/scripts/migrate-pnpm, so replace cannot replace pnpm with it"})
	rc.Run(t, rc.Spec{Name: "replace: the new tool is selected", Code: 2, Values: devValues + "\n[tools.pnpm]\n[tools.aube]",
		Call: engineCall("replace", "tool", "--yes", "pnpm", "aube"), Message: "aube is already selected"})
	rc.Run(t, rc.Spec{Name: "replace: a work tree with changes that are not committed", Code: 2, Values: devValues + "\n[tools.pnpm]",
		Change: pnpmFiles, Call: func(c *rc.Case) int {
			if code := c.Synced(); code != 0 {
				return code
			}
			harness.WriteFile(c.T, c.Path("untracked.txt"), "x\n")
			return c.Engine("replace", "tool", "--yes", "pnpm", "aube")
		}, Message: "the work tree has changes that are not committed", Check: func(c *rc.Case) {
			c.CheckThat("replace: pnpm-workspace.yaml stays", exists(c.Path("pnpm-workspace.yaml")))
		}})
	rc.Run(t, rc.Spec{Name: "replace with one tool", Code: 2, Values: devValues, Call: engineCall("replace", "tool", "--yes", "pnpm"),
		Message: "replace tool needs two tools"})
}

// The catalog keys: summary, category, selectable and description of a tool.
const (
	vtestValues = "environments = [ \"dev\" ]\n[tools.vtest]"
	vtestAt     = "repo/tools/vtest/tools.toml: "
)

var summary47 = strings.Repeat("x", 47)

// catTool writes the tool vtest with no file, an empty depends list, and
// <lines> after its top-level keys.
func catTool(lines ...string) func(c *rc.Case) {
	return func(c *rc.Case) {
		harness.WriteFile(c.T, c.ClonePath("repo/tools/vtest/tools.toml"),
			harness.Lines(append([]string{`default_envs = [ "dev" ]`, "depends = [  ]", "root_only = false"}, lines...)...))
	}
}

func TestCatalog(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "catalog: a summary of 48 characters", Code: 2, Values: vtestValues,
		Change: catTool(`summary = "` + summary47 + `y"`), Message: vtestAt + "summary must be a string of 47 characters or fewer"})
	rc.Run(t, rc.Spec{Name: "catalog: a summary that is not a string", Code: 2, Values: vtestValues,
		Change: catTool("summary = 1"), Message: vtestAt + "summary must be a string of 47 characters or fewer"})
	rc.Run(t, rc.Spec{Name: "catalog: a category that is not a string", Code: 2, Values: vtestValues,
		Change: catTool("category = 1"), Message: vtestAt + "category must be a string"})
	rc.Run(t, rc.Spec{Name: "catalog: selectable that is not a boolean", Code: 2, Values: vtestValues,
		Change: catTool(`selectable = "no"`), Message: vtestAt + "selectable must be true or false"})
	rc.Run(t, rc.Spec{Name: "catalog: a description that is not a string", Code: 2, Values: vtestValues,
		Change: catTool("description = 1"), Message: vtestAt + "description must be a string"})
	rc.Run(t, rc.Spec{Name: "catalog: a summary of 47 characters and a description", Values: vtestValues,
		Change:  catTool(`summary = "`+summary47+`"`, `description = "The scratch tool of the catalog cases."`, "selectable = true"),
		Message: "Dry run: "})
}
