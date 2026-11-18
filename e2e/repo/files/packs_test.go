package files

import (
	"regexp"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

const oldFormat = "is the old format - write each member as <tool> = {}"

// packsLines gives a change: these lines after an empty line at the end of
// the packs of the scratch clone.
func packsLines(lines ...string) func(c *rc.Case) {
	return func(c *rc.Case) {
		harness.AppendFile(c.T, c.ClonePath("repo/packs.toml"), harness.Lines(append([]string{""}, lines...)...))
	}
}

var (
	packsExtra   = packsLines("[pack.extra]", "java = {}", "pulumi = {}")
	packsFlutter = packsLines("[pack.mobile]", "flutter = {}", "android = {}", "go = {}")
)

func packsOuter(c *rc.Case) {
	packsExtra(c)
	packsLines("[pack.outer]", "go = {}", `"pack:extra" = {}`)(c)
}

// addPack syncs and commits the consumer, then runs
// `bootstrap add pack --yes <args>`.
func addPack(args ...string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		return c.Engine(append([]string{"add", "pack", "--yes"}, args...)...)
	}
}

func hasTool(c *rc.Case, tool string) bool { return c.TomlQ(`.tools | has("`+tool+`")`, "") }

func checkAddPackKept(c *rc.Case) {
	c.CheckThat("add pack: the values file stays", notGrep("pulumi", c.ToolsValues))
}

// The packs model: each member of a pack is a default, and a pack is a
// shortcut of add (D2, D4). The sync examines each pack of the file.
func TestPackRules(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "[pack.*], the old depends", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd]", `depends = [ "p:security" ]`), Message: "repo/packs.toml: [pack.odd] depends " + oldFormat})
	rc.Run(t, rc.Spec{Name: "[pack.*], the old required", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd]", `required = [ "_security_base" ]`), Message: "repo/packs.toml: [pack.odd] required " + oldFormat})
	rc.Run(t, rc.Spec{Name: "[pack.*], the old tools", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd]", `tools = [ "node" ]`), Message: "repo/packs.toml: [pack.odd] tools " + oldFormat})
	rc.Run(t, rc.Spec{Name: "[pack.*], the old default list", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd]", `default = [ "node" ]`), Message: "repo/packs.toml: [pack.odd] default " + oldFormat})
	rc.Run(t, rc.Spec{Name: "[pack.*], a key with the old prefix p:", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd]", `"p:x" = {}`), Message: `repo/packs.toml: [pack.odd] "p:x" is the old form - write "pack:x" = {}`})
	rc.Run(t, rc.Spec{Name: "[pack.*], an entry key other than required", Code: 2, Values: noToolValues,
		Change:  packsLines("[pack.odd]", `node = { version = "1" }`),
		Message: "repo/packs.toml: [pack.odd] node has the entry key 'version' - a member is {}, or { required = true } in pack base"})
	rc.Run(t, rc.Spec{Name: "[pack.*], the key default", Code: 2, Values: noToolValues,
		Change:  packsLines("[pack.odd]", `node = { default = true }`),
		Message: "repo/packs.toml: [pack.odd] node has the entry key 'default', the old form - each member is a default: write node = {}"})
	rc.Run(t, rc.Spec{Name: "[pack.*], required = true in a pack other than base", Code: 2, Values: noToolValues,
		Change:  packsLines("[pack.odd]", `node = { required = true }`),
		Message: "repo/packs.toml: [pack.odd] node: required = true is only for a tool of pack base"})
	rc.Run(t, rc.Spec{Name: "[pack.*], a member that is a string", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd]", `node = "x"`), Message: "repo/packs.toml: [pack.odd] node must be a table, {} or { required = true }"})
	rc.Run(t, rc.Spec{Name: "[pack.*], a loop through nested packs", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.x]", `"pack:y" = {}`, "", "[pack.y]", `"pack:x" = {}`), Message: "repo/packs.toml: a pack loop: x -> y -> x"})
	rc.Run(t, rc.Spec{Name: "[pack.*], an unknown tool in a pack", Code: 3, Values: noToolValues,
		Change: packsLines("[pack.odd]", "nope = {}"), Message: "nope is not a bootstrap tool (named in pack odd)"})
	rc.Run(t, rc.Spec{Name: "[pack.*], an unknown pack in a pack", Code: 3, Values: noToolValues,
		Change: packsLines("[pack.odd]", `"pack:nope" = {}`), Message: "nope is not a bootstrap pack (named in pack odd)"})
	rc.Run(t, rc.Spec{Name: "[pack.*], a member on two lines", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd]", "node = {", "}"), Message: "repo/packs.toml: [pack.odd] node must be on one line"})
	rc.Run(t, rc.Spec{Name: "[pack.*], a member as its own table", Code: 2, Values: noToolValues,
		Change: packsLines("[pack.odd.node]"), Message: "repo/packs.toml: [pack.odd] node must be on one line"})
}

// add pack: a table for each member that is not selected and for its new
// dependencies; --exclude leaves out a member (D2, D16 e, D17 f). The members
// of the pack extra are outside the pack base and its depends. No tool is
// selected by a pack only.
func TestAddPack(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "a pack is not a selection", Values: devValues, Change: packsExtra,
		Check: func(c *rc.Case) {
			c.CheckThat("a pack: its members are not selected by a sync", !toolsHas(c, "pulumi"))
			c.CheckThat("a pack: no table of its member", notGrep("pulumi", c.ToolsValues))
		}})
	rc.Run(t, rc.Spec{Name: "add pack extra", Values: devValues, Change: packsExtra, Call: addPack("extra"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add pack extra: the table of java", hasTool(c, "java"))
			c.CheckThat("add pack extra: the table of pulumi", hasTool(c, "pulumi"))
			c.CheckThat("add pack extra: no table of a pack", notGrep(`^\[packs\.`, c.ToolsValues))
		}})
	rc.Run(t, rc.Spec{Name: "add pack extra --exclude java", Values: devValues, Change: packsExtra,
		Call: addPack("extra", "--exclude", "java"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add pack --exclude: no table of java", notGrep("java", c.ToolsValues))
			c.CheckThat("add pack --exclude: the table of pulumi", hasTool(c, "pulumi"))
		}})
	rc.Run(t, rc.Spec{Name: "add pack again adds only the members that are not selected", Values: devValues + `
[tools.java]
env = [ "dev" ]`, Change: packsExtra, Call: addPack("extra"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add pack again: java keeps its env", c.TomlQ(`.tools.java.env[0] == "dev"`, ""))
			c.CheckThat("add pack again: the table of pulumi", hasTool(c, "pulumi"))
		}})
	rc.Run(t, rc.Spec{Name: "add pack: a nested pack", Values: devValues, Change: packsOuter, Call: addPack("outer"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add pack outer: the table of go", hasTool(c, "go"))
			c.CheckThat("add pack outer: the nested pack: java", hasTool(c, "java"))
			c.CheckThat("add pack outer: the nested pack: pulumi", hasTool(c, "pulumi"))
		}})
	rc.Run(t, rc.Spec{Name: "add pack --exclude of a nested pack", Values: devValues, Change: packsOuter,
		Call: addPack("outer", "--exclude", "pack:extra"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add pack --exclude pack:extra: the table of go", hasTool(c, "go"))
			c.CheckThat("add pack --exclude pack:extra: no table of java", notGrep("java", c.ToolsValues))
		}})
	rc.Run(t, rc.Spec{Name: "add pack --exclude of a member of a nested pack", Values: devValues, Change: packsOuter,
		Call: addPack("outer", "--exclude", "java"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add pack --exclude of a member of a nested pack: no table of java", notGrep("java", c.ToolsValues))
			c.CheckThat("add pack --exclude of a member of a nested pack: the table of pulumi", hasTool(c, "pulumi"))
		}})
	rc.Run(t, rc.Spec{Name: "add pack --exclude of a name that is not a member", Code: 2, Values: devValues, Change: packsExtra,
		Call: addPack("extra", "--exclude", "nope"), Message: "nope is not a member of the pack extra", Check: checkAddPackKept})
	rc.Run(t, rc.Spec{Name: "add pack --exclude of a required member", Code: 2, Values: devValues,
		Call: addPack("base", "--exclude", "mise"), Message: "mise is required in the pack base, so add pack cannot exclude it"})
	rc.Run(t, rc.Spec{Name: "add pack of two packs with --exclude", Code: 2, Values: devValues, Change: packsOuter,
		Call:    addPack("extra", "outer", "--exclude", "java"),
		Message: "--exclude needs exactly one pack name - run bootstrap add pack once for each pack", Check: checkAddPackKept})
	rc.Run(t, rc.Spec{Name: "add pack of two packs without --exclude", Values: devValues, Change: packsOuter,
		Call: addPack("extra", "outer"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			c.CheckThat("add pack of two packs: the table of go", hasTool(c, "go"))
			c.CheckThat("add pack of two packs: the table of pulumi once", countLine(c.ToolsValues, "[tools.pulumi]") == 1)
		}})
	rc.Run(t, rc.Spec{Name: "add pack --exclude of a member that another member needs", Code: 2, Values: devValues,
		Change: packsFlutter, Call: addPack("mobile", "--exclude", "android"),
		Message: "android is needed by flutter - exclude flutter too, or keep android"})
	rc.Run(t, rc.Spec{Name: "add pack of an unknown pack", Code: 3, Values: devValues, Call: addPack("nope"),
		Message: "nope is not a bootstrap pack"})
	rc.Run(t, rc.Spec{Name: "add pack without a pack", Code: 2, Values: devValues,
		Call: func(c *rc.Case) int {
			if code := c.Synced(); code != 0 {
				return code
			}
			return c.Engine("add", "pack")
		}, Message: "add pack needs at least one pack"})
}

// baseLacksSecurity: the scratch pack base without the scanners, so a
// scanner comes only from the case.
func baseLacksSecurity(c *rc.Case) {
	p := c.ClonePath("repo/packs.toml")
	re := regexp.MustCompile(`(?m)^(gitleaks|grype|osv-scanner) = .*$\n?`)
	harness.WriteFile(c.T, p, re.ReplaceAllString(harness.ReadFile(p), ""))
}

// The tool order, as at 3b199d2, the commit before the merged packs model.
// The tools before the security tools stay. At that commit _security_base
// came after the scanners; it now comes before them, as a dependency of each
// one.
const orderTools = "mise git python pre-commit dprint node linter _security_base gitleaks grype osv-scanner"

func TestScannersAndOrder(t *testing.T) {
	// One scanner brings _security_base through its depends, without the
	// others.
	rc.Run(t, rc.Spec{Name: "[tools.gitleaks], a pack base without the scanners", Values: `environments = [ "dev" ]
[tools.gitleaks]`, Change: baseLacksSecurity,
		Check: func(c *rc.Case) {
			c.CheckThat("[tools.gitleaks], no scanner in base: a file of _security_base",
				c.LogHas("would update "+c.Path(".config/mise/tasks/code/sec/_default")))
			c.CheckThat("[tools.gitleaks], no scanner in base: no other scanner", !toolsHas(c, "grype"))
		}})
	rc.Run(t, rc.Spec{Name: "no scanner, a pack base without the scanners", Values: noToolValues, Change: baseLacksSecurity,
		Check: func(c *rc.Case) {
			c.CheckThat("no scanner in base: no _security_base", !toolsHas(c, "_security_base"))
		}})
	rc.Run(t, rc.Spec{Name: "the tool order does not change", Values: noToolValues,
		Check: func(c *rc.Case) {
			c.CheckThat("the order of the pack base: the tools", hasLine(c.Log, "bootstrap: tools: "+orderTools))
		}})
}
