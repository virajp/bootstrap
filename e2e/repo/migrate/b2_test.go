package migrate

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// yqLines runs yq with <expr> on the TOML file <file> and gives the lines of
// its TSV output.
func yqLines(c *rc.Case, expr, file string) []string {
	cmd := &harness.Cmd{Env: c.Env, Args: []string{"yq", "-p", "toml", "-o", "tsv", expr, file}}
	code, out := cmd.Run()
	if code != 0 {
		c.T.Fatalf("yq %s %s: exit %d\n%s", expr, file, code, out)
	}
	out = strings.TrimRight(out, "\n")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// exampleConsumer gives the case the values of the example <example>, in a
// consumer with its core branches and its members: a folder for each folder
// member, and a git submodule for each submodule member. A sibling member is
// not there, so the sync skips it with an info line.
func exampleConsumer(example string) func(c *rc.Case) {
	return func(c *rc.Case) {
		if err := os.RemoveAll(c.Values); err != nil {
			c.T.Fatal(err)
		}
		for _, f := range []string{"repo.toml", "tools.toml"} {
			harness.WriteFile(c.T, filepath.Join(c.Values, f),
				harness.ReadFile(filepath.Join(harness.Root, "repo/examples", example, f)))
		}
		for _, branch := range yqLines(c, ".core_branches // [] | .[]", c.RepoValues) {
			if branch == "" {
				continue
			}
			if _, err := harness.GitErr(c.Consumer, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch); err != nil {
				git(c, "branch", branch)
			}
		}
		for _, row := range yqLines(c, `.members // {} | to_entries | .[] | [.key, .value.type // "", .value.folder // ""]`, c.RepoValues) {
			f := strings.Split(row, "\t")
			for len(f) < 3 {
				f = append(f, "")
			}
			name, typ, folder := f[0], f[1], f[2]
			if folder == "" {
				folder = name
			}
			switch typ {
			case "submodule":
				git(c, "-c", "protocol.file.allow=always", "submodule", "--quiet", "add", "--", rc.BaseClone, folder)
			case "folder":
				mkdir(c, c.Path(folder))
			}
		}
	}
}

// TestExamples: each folder of repo/examples/ with a repo.toml is the values
// folder of a repository, with its tools.toml next to it. Each one must pass
// a dry run. The dry run refuses an unknown key, an unknown tool, a wrong
// member and a wrong shape.
func TestExamples(t *testing.T) {
	base := filepath.Join(harness.Root, "repo/examples")
	var examples []string
	_ = filepath.WalkDir(base, func(p string, e fs.DirEntry, err error) error {
		if err == nil && !e.IsDir() && e.Name() == "repo.toml" {
			rel, _ := filepath.Rel(base, filepath.Dir(p))
			examples = append(examples, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(examples)
	for _, example := range examples {
		rc.Run(t, rc.Spec{Name: "the example " + example, Change: exampleConsumer(example)})
	}
}

// TestOldKeys: an old key of the values file. A top-level key or a table of
// the old form is not a key of repo.toml, and a key of a [tools.*] table that
// lives in repo.toml now names its new place.
func TestOldKeys(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "an old key: commit_scopes at the top", Code: 2, Values: devValues + `
commit_scopes = [ "api" ]`, Message: "[ERROR] commit_scopes in "})
	rc.Run(t, rc.Spec{Name: "an old key: worktree_files at the top", Code: 2, Values: devValues + `
worktree_files = [ ".env" ]`, Message: "[ERROR] worktree_files in "})
	rc.Run(t, rc.Spec{Name: "an old key: [claude.plugins.user]", Code: 2, Values: devValues + `
[claude.plugins.user]`, Message: "[ERROR] claude in "})
	rc.Check(t, rc.Spec{Name: "--check, an old key: [branch.main]", Code: 2, Values: devValues + `
[branch.main]
merge_model = "pr"`, Message: "[ERROR] branch in "})
	rc.Run(t, rc.Spec{Name: "an old key: commit_scopes of [tools.git]", Code: 2, Values: devValues + `
[tools.git]
commit_scopes = [ "api" ]`, Message: ".config/bootstrap/tools.toml is a key of repo.toml now - write it as commits.scopes in repo.toml"})
	rc.Run(t, rc.Spec{Name: "an old key: worktree_files of [tools.mise]", Code: 2, Values: devValues + `
[tools.mise]
worktree_files = [ ".env" ]`, Message: ".config/bootstrap/tools.toml is a key of repo.toml now - write it as worktree.files in repo.toml"})
	rc.Check(t, rc.Spec{Name: "--check, an old key: [overrides.tools.git] commit_scopes", Code: 2, Values: devValues + `
[overrides.tools.git]
commit_scopes = [ "api" ]`, Message: ".config/bootstrap/tools.toml is a key of repo.toml now - write it as commits.scopes in repo.toml"})
}

// TestCommits: the scopes of [commits] in repo.toml reach the template of
// the commit types. A real run, because a dry run prints no content for a
// new file.
func TestCommits(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "[commits] scopes of repo.toml", Args: noArgs, Values: devValues + `
[commits]
scopes = [ "from-values" ]`, Check: func(c *rc.Case) {
		c.CheckThat("[commits] scopes: in the template", rc.FileHas(c.Path(".config/git-conventional-commits.yaml"), "from-values"))
	}})
	rc.Run(t, rc.Spec{Name: "the default scopes are the [members.*] keys", Args: noArgs, Values: devValues + `
[members.web]
folder = "projects/web"
type = "folder"
[members.api]
folder = "../api"
type = "sibling"
url = "https://example.com/api"`, Change: memberFolder, Check: func(c *rc.Case) {
		file := c.Path(".config/git-conventional-commits.yaml")
		c.CheckThat("the default scopes: a folder member", reIn(file, `^ +- ["']web["']$`))
		c.CheckThat("the default scopes: a sibling that is not there", reIn(file, `^ +- ["']api["']$`))
	}})
	rc.Run(t, rc.Spec{Name: "[commits] style, not one of its values", Code: 2, Values: devValues + `
[commits]
style = "free"`, Message: `commits.style = "free" in `})
}

// consumerUnborn: in a repository with no commit, the unborn branch of HEAD
// is a core branch that exists (D48).
func consumerUnborn(c *rc.Case) {
	if err := os.RemoveAll(c.Path(".git")); err != nil {
		c.T.Fatal(err)
	}
	git(c, "init", "--quiet", "-b", "main")
}

// TestBranches: the derive script of git gives has_main and has_develop, and
// the merge task of each core branch follows its flag.
func TestBranches(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "when: main only gives git:merge:main", Values: devValues, Check: func(c *rc.Case) {
		c.CheckThat("when, main only: git:merge:main", c.LogHas("/.config/mise/tasks/git/merge/main.py"))
		c.CheckThat("when, main only: no git:merge:develop", c.LogLacks("/.config/mise/tasks/git/merge/develop.py"))
	}})
	rc.Run(t, rc.Spec{Name: "when: main and develop give git:merge:main and develop", Values: bothCore,
		Change: precommitBothBranches, Check: func(c *rc.Case) {
			c.CheckThat("when, both branches: git:merge:main", c.LogHas("/.config/mise/tasks/git/merge/main.py"))
			c.CheckThat("when, both branches: git:merge:develop", c.LogHas("/.config/mise/tasks/git/merge/develop.py"))
		}})
	rc.Run(t, rc.Spec{Name: "the git check: a core branch that does not exist", Code: 2, Values: devValues,
		Change: consumerNoCoreBranch, Message: "core_branches in "})
	rc.Run(t, rc.Spec{Name: "the git check: the unborn branch of HEAD", Dirty: true, Values: devValues, Change: consumerUnborn})
	rc.Check(t, rc.Spec{Name: "--check, core_branches with the unborn branch of HEAD", Dirty: true, Values: devValues,
		Change: consumerUnborn, Message: " is valid"})
}

// TestAliases: the derive script of mise gives the aliases of the submodule
// and sibling members that have a [members.<name>] table.
func TestAliases(t *testing.T) {
	checkNoAliases := func(c *rc.Case) {
		c.CheckThat("no member table: no setup-all", !rc.FileHas(c.Path(".config/mise/conf.d/mise/mise.dev.toml"), "setup-all"))
	}
	rc.Run(t, rc.Spec{Name: "the aliases of the submodule members", Args: noArgs, Values: devValues + `
[members.sub]
folder = "sub"
type = "submodule"`, Change: consumerWithSubmodule, Check: func(c *rc.Case) {
		file := c.Path(".config/mise/conf.d/mise/mise.dev.toml")
		c.CheckThat("the aliases: setup-all", reIn(file, `^setup-all = "mise run setup --all"`))
		c.CheckThat("the aliases: setup-sub", reIn(file, `^setup-sub = "mise run setup --sub"`))
		c.CheckThat("the aliases: the alias of the tool stays", reIn(file, `^setup = "mise run setup"`))
	}})
	rc.Run(t, rc.Spec{Name: "a submodule without a table, no member aliases", Args: noArgs, Values: devValues,
		Change: consumerWithSubmodule, Check: checkNoAliases})
	rc.Run(t, rc.Spec{Name: "no submodule, no member aliases", Args: noArgs, Values: devValues, Check: checkNoAliases})
}

// TestMiseDepends: the mise depends come from depends and parent of the
// tool, and the mise tools above in the same tool.
func TestMiseDepends(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "the mise depends are computed", Args: noArgs, Values: devValues + `
[tools.pnpm]`, Check: func(c *rc.Case) {
		c.CheckThat("depends: pnpm on node", reIn(c.Path(".config/mise/conf.d/pnpm/mise.dev.toml"), `^depends = \[ ?"node" ?\]`))
		c.CheckThat("depends: uv on python", reIn(c.Path(".config/mise/conf.d/python/mise.dev.toml"), `^depends = \[ ?"python" ?\]`))
		c.CheckThat("depends: node has none", !rc.FileHas(c.Path(".config/mise/conf.d/node/mise.dev.toml"), "depends"))
	}})
	rc.Run(t, rc.Spec{Name: "a hand-written mise depends", Code: 2, Values: devValues + `
[tools.node]`, Change: func(c *rc.Case) {
		appendLines(c, c.ClonePath("repo/tools/node/tools.toml"), "", "[mise.tools.extra]", `version = "1"`, `depends = [ "node" ]`)
	}, Message: "depends of mise tool extra in [mise.tools] - the sync computes it"})
}

// scriptTool: the tool fixture with a scripts/ folder.
func scriptTool(c *rc.Case) {
	fixtureTool(c)
	mkdir(c, c.ClonePath("repo/tools/fixture/scripts"))
}

func scriptWrite(c *rc.Case, name string, lines ...string) {
	writeLines(c, c.ClonePath("repo/tools/fixture/scripts/"+name), lines...)
}

func checkFails(c *rc.Case) {
	scriptTool(c)
	scriptWrite(c, "check", "cat > /dev/null", `echo "fixture: the first problem"`, `echo "fixture: the second problem"`, "exit 2")
}

func checkCrashes(c *rc.Case) {
	scriptTool(c)
	scriptWrite(c, "check", "cat > /dev/null", `echo "a crash" >&2`, "exit 7")
}

func deriveBad(c *rc.Case) {
	scriptTool(c)
	scriptWrite(c, "derive", "cat > /dev/null", `echo "[1]"`)
}

// deriveValue: the input of the script, and the folder where it runs.
func deriveValue(c *rc.Case) {
	scriptTool(c)
	fixtureToml(c, "[defaults]", `word = "default"`)
	scriptWrite(c, "derive", `in=$(cat)`, `printf "%s\n" "$in" > "$PWD/.derive-input"`,
		`printf "{\"word\": \"derived\", \"other\": \"derived\"}\n"`)
	fixtureFile(c, ".config/fixture.txt.tmpl", "word=<% .config.word %> other=<% .config.other %>")
}

func unknownScript(c *rc.Case) {
	scriptTool(c)
	scriptWrite(c, "build", "x")
}

func checkDerived(c *rc.Case) {
	in := c.Path(".derive-input")
	checkExact(c, "derive: the values file wins over the derived value", c.Path(".config/fixture.txt"), "word=user other=derived\n")
	c.CheckThat("derive: the input has the tool", rc.FileHas(in, `"tool":"fixture"`))
	c.CheckThat("derive: the input has levels 1 and 2", rc.FileHas(in, `"values":{"word":"default"}`))
	c.CheckThat("derive: the input has the repo name", rc.FileHas(in, `"name":"test"`))
}

// permFiles: perm and dir_perm of a [files] entry.
func permFiles(c *rc.Case) {
	fixtureTool(c)
	fixtureFile(c, "files/secret.txt", "secret")
	fixtureToml(c, `[files."private/secret.txt"]`, `source = "secret.txt"`, `perm = "600"`, `dir_perm = "700"`)
}

// misesFirst: each "updated" line of a mise file comes before the other
// "updated" lines.
func misesFirst(c *rc.Case) bool {
	conf := regexp.MustCompile(`\.config/mise/conf\.d/`)
	other := false
	for _, l := range strings.Split(c.LogText(), "\n") {
		if !strings.Contains(l, "updated ") {
			continue
		}
		if conf.MatchString(l) {
			if other {
				return false
			}
			continue
		}
		other = true
	}
	return true
}

func TestScripts(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "a check script that finds problems", Code: 2, Values: fixtureValues, Change: checkFails,
		Message: "fixture: the second problem"})
	rc.Run(t, rc.Spec{Name: "a check script that crashes", Code: 2, Values: fixtureValues, Change: checkCrashes,
		Message: "repo/tools/fixture/scripts/check failed with exit code 7"})
	rc.Run(t, rc.Spec{Name: "a derive script that prints no object", Code: 2, Values: fixtureValues, Change: deriveBad,
		Message: "repo/tools/fixture/scripts/derive did not print a JSON object"})
	rc.Run(t, rc.Spec{Name: "the derive values: level 3 below the values file", Args: noArgs, Values: fixtureValues + `
word = "user"`, Change: deriveValue, Check: checkDerived})
	rc.Run(t, rc.Spec{Name: "a file of scripts/ that is not a script", Code: 2, Values: fixtureValues, Change: unknownScript,
		Message: "scripts/build is not a script of a tool"})
	rc.Check(t, rc.Spec{Name: "--check runs the check scripts", Code: 2, Values: fixtureValues, Change: checkFails,
		Message: "fixture: the first problem"})
	rc.Run(t, rc.Spec{Name: "perm and dir_perm", Args: noArgs, Values: fixtureValues, Change: permFiles, Check: func(c *rc.Case) {
		c.CheckThat("perm: the file has mode 600", rc.ModeIs(c.Path("private/secret.txt"), "600"))
		c.CheckThat("dir_perm: the folder has mode 700", rc.ModeIs(c.Path("private"), "700"))
		c.CheckThat("the mise files first", misesFirst(c))
	}})
}
