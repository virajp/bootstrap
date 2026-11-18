package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// A tool with no summary or no category, or with category = "" (D5, D10 d).
// KeysAsWritten leaves vtest as the case writes it.
func TestCatalogKeys(t *testing.T) {
	keep := []string{"vtest"}
	rc.Run(t, rc.Spec{Name: "catalog: a tool with no summary", Code: 2, Values: vtestValues, KeysAsWritten: keep,
		Change:  func(c *rc.Case) { catTool(c, "vtest", "", `category = "core"`) },
		Message: vtestAt + "summary is missing"})
	rc.Run(t, rc.Spec{Name: "catalog: a tool with no category", Code: 2, Values: vtestValues, KeysAsWritten: keep,
		Change:  func(c *rc.Case) { catTool(c, "vtest", "", `summary = "The scratch tool vtest."`) },
		Message: vtestAt + "category is missing"})
	rc.Run(t, rc.Spec{Name: `catalog: a tool with category = ""`, Code: 2, Values: vtestValues, KeysAsWritten: keep,
		Change:  func(c *rc.Case) { catTool(c, "vtest", "", `summary = "The scratch tool vtest."`, `category = ""`) },
		Message: vtestAt + "category is missing"})
}

// categoriesFile adds tables after the categories of the copy of
// repo/categories.toml (categories_file).
func categoriesFile(c *rc.Case, lines ...string) { appendClone(c, "repo/categories.toml", lines...) }

// The categories file repo/categories.toml (D2). A case changes only the copy
// of the file in its own clone (D12 b).
func TestCategories(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "categories: a category and no repo/categories.toml", Code: 2, Values: vtestValues,
		Change: func(c *rc.Case) {
			_ = os.Remove(c.ClonePath("repo/categories.toml"))
			catTool(c, "vtest", "", `category = "lang"`)
		},
		Message: vtestAt + "category lang needs the file repo/categories.toml"})
	rc.Run(t, rc.Spec{Name: "categories: an unknown category", Code: 2, Values: vtestValues,
		Change: func(c *rc.Case) {
			categoriesFile(c, "[category.lang]", `title = "Languages"`)
			catTool(c, "vtest", "", `category = "nope"`)
		},
		Message: vtestAt + "category nope is not a category of repo/categories.toml"})
	rc.Run(t, rc.Spec{Name: "categories: a key other than title", Code: 2, Values: noToolValues,
		Change:  func(c *rc.Case) { categoriesFile(c, "[category.lang]", `color = "red"`, `title = "Languages"`) },
		Message: "repo/categories.toml: [category.lang] color is not a key of a category - the only key is title"})
	rc.Run(t, rc.Spec{Name: "categories: a category with no title", Code: 2, Values: noToolValues,
		Change:  func(c *rc.Case) { categoriesFile(c, "[category.lang]") },
		Message: "repo/categories.toml: [category.lang] needs a title string"})
	rc.Run(t, rc.Spec{Name: "categories: a title that is not a string", Code: 2, Values: noToolValues,
		Change:  func(c *rc.Case) { categoriesFile(c, "[category.lang]", "title = 1") },
		Message: "repo/categories.toml: [category.lang] title must be a string"})
	rc.Run(t, rc.Spec{Name: "categories: a top-level key", Code: 2, Values: noToolValues,
		Change:  func(c *rc.Case) { prependFile(c, c.ClonePath("repo/categories.toml"), `title = "x"`, "") },
		Message: "repo/categories.toml: title is not a key of the categories file - each entry is [category.<id>] with a title"})
	rc.Run(t, rc.Spec{Name: "categories: a valid file and category", Values: vtestValues,
		Change: func(c *rc.Case) {
			categoriesFile(c, "[category.lang]", `title = "Languages"`, "", "[category.tool]", `title = "Tools"`)
			catTool(c, "vtest", "", `category = "tool"`)
		},
		Message: "Dry run: "})

	// The description of a pack (D3).
	rc.Run(t, rc.Spec{Name: "[pack.*], a description", Values: noToolValues,
		Change: func(c *rc.Case) {
			appendClone(c, "repo/packs.toml", "[pack.odd]", `description = "The odd tools."`, "node = {}")
		},
		Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "[pack.*], a description that is not a string", Code: 2, Values: noToolValues,
		Change:  func(c *rc.Case) { appendClone(c, "repo/packs.toml", "[pack.odd]", "description = 1", "node = {}") },
		Message: "repo/packs.toml: [pack.odd] description must be a string"})
}

// The scratch tools and the calls of the hosts and selectable cases.
func packVtest(c *rc.Case) {
	appendClone(c, "repo/packs.toml", "[pack.extra]", "vtest = {}", "go = {}")
}

var (
	addVtest         = synced(engine("add", "tool", "--yes", "vtest"))
	replacePnpmVtest = synced(engine("replace", "tool", "--yes", "pnpm", "vtest"))
	addPackExtra     = synced(engine("add", "pack", "--yes", "extra"))
)

func checkVtestKept(c *rc.Case) {
	c.CheckThat("the values file does not select vtest", !strings.Contains(harness.ReadFile(c.ToolsValues), "vtest"))
}

func checkPackSkipsVtest(c *rc.Case) {
	checkVtestKept(c)
	c.CheckThat("add pack: the other member is selected", c.TomlQ(`.tools | has("go")`, ""))
}

// hosts (D4). The tool names a host other than the host of the test, so the
// cases do not depend on the machine. The engine names darwin macos.
func TestHosts(t *testing.T) {
	this, other := hosts()
	msgHost := "vtest needs the host " + other + " - this host is " + this
	hostOther := func(c *rc.Case) { catTool(c, "vtest", "", `hosts = [ "`+other+`" ]`) }
	rc.Run(t, rc.Spec{Name: "hosts: add of a tool for another host", Code: 2, Values: devValues, Change: hostOther,
		Call: addVtest, Message: msgHost, Check: checkVtestKept})
	// add pack leaves out, with an info line, a member for another host and a
	// member with selectable = false (D17 g).
	rc.Run(t, rc.Spec{Name: "hosts: add pack leaves out a member for another host", Values: devValues,
		Change: func(c *rc.Case) { hostOther(c); packVtest(c) }, Call: addPackExtra,
		Message: "vtest is left out: its hosts leave out this host", Check: checkPackSkipsVtest})
	rc.Run(t, rc.Spec{Name: "hosts: sync of a values file that names a tool for another host", Values: vtestValues,
		Change: hostOther, Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "hosts: a name that is not a mise OS name", Code: 2, Values: vtestValues,
		Change:  func(c *rc.Case) { catTool(c, "vtest", "", `hosts = [ "darwin" ]`) },
		Message: vtestAt + "hosts must be a list of macos, linux and windows"})
	rc.Run(t, rc.Spec{Name: "hosts: an empty list", Code: 2, Values: vtestValues,
		Change:  func(c *rc.Case) { catTool(c, "vtest", "", "hosts = []") },
		Message: vtestAt + "hosts must name at least one host"})
	rc.Run(t, rc.Spec{Name: "hosts: replace with a new tool for another host", Code: 2, Values: devValues + "\n[tools.pnpm]",
		Change:  func(c *rc.Case) { catTool(c, "vtest", `"node"`, `parent = "node"`, `hosts = [ "`+other+`" ]`) },
		Call:    replacePnpmVtest,
		Message: msgHost})
}

// selectable = false (D5).
func TestSelectable(t *testing.T) {
	notSelectable := func(c *rc.Case) { catTool(c, "vtest", "", "selectable = false") }
	notSelectableNeeded := func(c *rc.Case) {
		notSelectable(c)
		catTool(c, "vneed", `"vtest"`)
		catTool(c, "vuse", `"vtest"`)
	}
	rc.Run(t, rc.Spec{Name: "selectable: add of a tool with selectable = false", Code: 2, Values: devValues,
		Change: notSelectableNeeded, Call: addVtest, Message: "vtest is not selectable. Select vneed or vuse.",
		Check: checkVtestKept})
	rc.Run(t, rc.Spec{Name: "selectable: add pack leaves out a member with selectable = false", Values: devValues,
		Change: func(c *rc.Case) { notSelectable(c); packVtest(c) }, Call: addPackExtra,
		Message: "vtest is left out: it is not selectable; a tool that needs it selects it", Check: checkPackSkipsVtest})
	rc.Run(t, rc.Spec{Name: "selectable: a tool that no tool needs", Code: 2, Values: devValues,
		Change: notSelectable, Call: addVtest, Message: "vtest is not selectable.",
		Check: func(c *rc.Case) { c.CheckThat("selectable = false: no tool to select", c.LogLacks("Select ")) }})
	rc.Run(t, rc.Spec{Name: "selectable: sync of a values file that names it", Values: vtestValues,
		Change: notSelectableNeeded, Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "selectable: check of a values file that names it", Values: vtestValues,
		Change: notSelectableNeeded, Call: synced(engine("check")),
		Message: "vtest is not selectable - remove [tools.vtest] from ",
		Check: func(c *rc.Case) {
			c.CheckThat("check: the warning names the values file",
				c.LogHas("from "+c.ToolsValues+"; a tool that needs it selects it"))
		}})
	rc.Run(t, rc.Spec{Name: "selectable: replace with a new tool with selectable = false", Code: 2,
		Values: devValues + "\n[tools.pnpm]",
		Change: func(c *rc.Case) { catTool(c, "vtest", `"node"`, `parent = "node"`, "selectable = false") },
		Call:   replacePnpmVtest, Message: "vtest is not selectable."})
}

// detect does not offer a tool with selectable = false or for another host.
func TestDetectNotOffered(t *testing.T) {
	_, other := hosts()
	pnpmFiles := func(c *rc.Case) {
		writeConsumer(c, "package.json", "{}")
		writeConsumer(c, "pnpm-workspace.yaml", "packages: []")
	}
	checkNoPnpm := func(c *rc.Case) {
		c.CheckThat("detect: pnpm is not offered", !logRe(c, `\[NEW\] pnpm|pnpm is found`))
		c.CheckThat("detect: the values file stays", !strings.Contains(harness.ReadFile(c.ToolsValues), "pnpm"))
	}
	detectDry := synced(engine("detect", "--add", "--yes", "--dry-run"))
	rc.Run(t, rc.Spec{Name: "detect: a tool with selectable = false", Values: devValues,
		Change: func(c *rc.Case) { pnpmFiles(c); firstLine(c, "pnpm", "selectable = false") },
		Call:   detectDry, Check: checkNoPnpm})
	rc.Run(t, rc.Spec{Name: "detect: a tool for another host", Values: devValues,
		Change: func(c *rc.Case) { pnpmFiles(c); firstLine(c, "pnpm", `hosts = [ "`+other+`" ]`) },
		Call:   detectDry, Check: checkNoPnpm})
}

// The scratch tools of the question cases (D6, D14, D15).
func groupTool(c *rc.Case) {
	catTool(c, "vtest", "", "[questions.a]", `group = "email"`, `note = "One answer of the group is enough."`,
		`question = "Which a?"`, "", "[questions.b]", `group = "email"`, `question = "Which b?"`)
}

func optionalTool(c *rc.Case) {
	catTool(c, "vtest", "", "[questions.q]", `example = "x"`, "optional = true", `question = "Which q?"`)
}

func necessaryTool(c *rc.Case) { catTool(c, "vtest", "", "[questions.q]", `question = "Which q?"`) }

func questionsTool(c *rc.Case) {
	groupTool(c)
	harness.AppendFile(c.T, c.ClonePath("repo/tools/vtest/tools.toml"), harness.Lines("", "[questions.q]",
		`example = "x"`, "optional = true", `question = "Which q?"`, "", "[questions.r]", "optional = true",
		`question = "Which r?"`))
}

const (
	msgGroup     = "the group vtest.email has no answer in "
	msgGroupFlag = "the group vtest.email has no answer - give at least one with --answer <tool>.<key>=<value>: a, b"
)

func checkGroupFile(c *rc.Case) {
	c.CheckThat("the group message names the values file", c.LogHas("in "+c.ToolsValues+" - run bootstrap migrate"))
}

func checkNoQuestionLine(c *rc.Case) {
	c.CheckThat("init: only the group error", c.LogLacks("the question vtest."))
}

func checkEmptyAnswer(c *rc.Case) {
	c.CheckThat("the empty answer", c.TomlQ(`.tools.vtest.q == ""`, ""))
}

// migrateTwice runs migrate with <args>, commits the consumer, then runs
// migrate --yes.
func migrateTwice(args ...string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		if code := c.Engine(append([]string{"migrate"}, args...)...); code != 0 {
			return code
		}
		c.CommitConsumer()
		return c.Engine("migrate", "--yes")
	}
}

var initYes = engine("migrate", "--yes")

func TestQuestions(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "questions: a group with no answer", Code: 2, Values: vtestValues, Change: groupTool,
		Message: msgGroup, Check: checkGroupFile})
	rc.Run(t, rc.Spec{Name: "questions: a group with empty answers", Code: 2, Values: vtestValues + "\na = \"\"\nb = \"\"",
		Change: groupTool, Message: msgGroup, Check: checkGroupFile})
	rc.Run(t, rc.Spec{Name: "questions: a group with one answer", Values: vtestValues + "\na = \"x\"", Change: groupTool,
		Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "questions: an optional question with no answer", Values: vtestValues, Change: optionalTool,
		Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "questions: an empty answer of a question", Values: vtestValues + "\nq = \"\"",
		Change: necessaryTool, Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "questions: a question with a group and optional = true", Code: 2, Values: vtestValues,
		Change: func(c *rc.Case) {
			catTool(c, "vtest", "", "[questions.a]", `group = "email"`, "optional = true", `question = "Which a?"`)
		},
		Message: vtestAt + "[questions.a] has a group, so it cannot be optional"})
	rc.Run(t, rc.Spec{Name: "questions: a question with a group and optional = false", Values: vtestValues + "\na = \"x\"",
		Change: func(c *rc.Case) {
			catTool(c, "vtest", "", "[questions.a]", `group = "email"`, "optional = false", `question = "Which a?"`)
		},
		Message: "Dry run: "})
}

func TestInitAnswers(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "init --yes: a group with no answer", Code: 2, Values: vtestValues, Change: groupTool,
		Call: initYes, Message: msgGroupFlag, Check: checkNoQuestionLine})
	rc.Run(t, rc.Spec{Name: "init --yes: empty answers of each question of a group", Code: 2, Values: vtestValues,
		Change: groupTool, Call: engine("migrate", "--yes", "--answer", "vtest.a=", "--answer", "vtest.b="),
		Message: msgGroupFlag, Check: checkNoQuestionLine})
	rc.Run(t, rc.Spec{Name: "init --yes again on a values file that answers a group", Values: vtestValues,
		Change: groupTool, Call: migrateTwice("--yes", "--answer", "vtest.a=x", "--answer", "vtest.b="),
		Message: "Nothing to change",
		Check: func(c *rc.Case) {
			c.CheckThat("init: the answer of the group", c.TomlQ(`.tools.vtest.a == "x"`, ""))
			c.CheckThat("init: the empty answer of the group", c.TomlQ(`.tools.vtest.b == ""`, ""))
		}})
	rc.Run(t, rc.Spec{Name: "init --yes: an optional question with no flag", Code: 2, Values: vtestValues,
		Change: optionalTool, Call: initYes,
		Message: "the question vtest.q has no answer - give it with --answer vtest.q=<value>, or leave it empty with --answer vtest.q="})
	rc.Run(t, rc.Spec{Name: "init --answer <q>= of an optional question, then init --yes", Values: vtestValues,
		Change: optionalTool, Call: migrateTwice("--yes", "--answer", "vtest.q="), Message: "Nothing to change",
		Check: checkEmptyAnswer})
	rc.Run(t, rc.Spec{Name: "init --answer <q>= of a necessary question", Code: 2, Values: vtestValues,
		Change: necessaryTool, Call: engine("migrate", "--yes", "--answer", "vtest.q="),
		Message: "--answer vtest.q=: an empty answer is only for an optional or a group question - give it with --answer vtest.q=<value>"})
	rc.Run(t, rc.Spec{Name: "add --answer <q>= of an optional question", Values: devValues, Change: optionalTool,
		Call:    synced(engine("add", "tool", "--yes", "--answer", "vtest.q=", "vtest")),
		Message: "Synced: ", Check: checkEmptyAnswer})
	rc.Run(t, rc.Spec{Name: "init --skip-answer is not a flag", Code: 2, Values: vtestValues, Change: optionalTool,
		Call: engine("migrate", "--yes", "--skip-answer", "vtest.q"), Message: "migrate has no flag --skip-answer"})
}

func TestQuestionsCommand(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "questions --json: group, note and optional", Values: vtestValues, Change: questionsTool,
		Call: engine("questions", "--json"), Message: `"key":"r"`,
		Check: func(c *rc.Case) {
			c.CheckThat("questions --json: the envelope", c.LogHas(`{"contract":1,"command":"questions","ok":true,"code":0,"data":[{`))
			c.CheckThat("questions --json: the group", c.LogHas(`"group":"email"`))
			c.CheckThat("questions --json: the note", c.LogHas(`"note":"One answer of the group is enough."`))
			c.CheckThat("questions --json: optional", c.LogHas(`"optional":true`))
			c.CheckThat("questions --json: not optional", c.LogHas(`"optional":false`))
		}})
	rc.Run(t, rc.Spec{Name: "questions: optional and group questions, and an empty answer",
		Values: vtestValues + "\na = \"x\"\nr = \"\"", Change: questionsTool, Call: engine("questions"),
		Message: "Which r?",
		Check: func(c *rc.Case) {
			c.CheckThat("questions: an empty answer is the ok glyph alone", logRe(c, `^ +\[OK\]$`))
			c.CheckThat("questions: an optional question with no answer", c.LogHas("[INFO] no answer (optional)"))
			c.CheckThat("questions: a group question with no answer", c.LogHas("[INFO] no answer (group email has an answer)"))
			c.CheckThat("questions: no question without an answer", c.LogLacks("without an answer"))
		}})
}
