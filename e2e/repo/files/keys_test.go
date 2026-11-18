package files

import (
	"testing"

	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

const (
	msgOldSeed     = `repo/tools/fixture/tools.toml: seed is not supported; use [files."<target>"] with mode = "seed"`
	msgOldTable    = `repo/tools/fixture/tools.toml: [seed_requires] is not supported; use requires in [files."<target>"]`
	msgOldRequires = `fixture/tools.toml: requires of mise tool fixture-x is the old key - write requires_found = ["x.toml"]`
	msgOldFiles    = `repo/tools/fixture/tools.toml: files in [mise.tools.fixture-x] is not supported; use requires_found in [files."<target>"]`
)

// fixtureWith gives a change: the tool fixture with these lines in its
// tools.toml.
func fixtureWith(lines ...string) func(c *rc.Case) {
	return func(c *rc.Case) {
		fixtureTool(c)
		fixtureToml(c, lines...)
	}
}

// The old keys. The tool `fixture` is not selected: the sync examines each
// tool of the clone.
func TestOldKeys(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "the old seed key, a tool that is not selected", Code: 2, Values: noToolValues,
		Change: fixtureWith(`seed = [ "a.txt" ]`), Message: msgOldSeed})
	rc.Run(t, rc.Spec{Name: "the old [seed_requires], a tool that is not selected", Code: 2, Values: noToolValues,
		Change: fixtureWith(`[seed_requires]`, `"a.txt" = "package.json"`), Message: msgOldTable})
	rc.Run(t, rc.Spec{Name: "the old requires of a mise entry", Code: 2, Values: noToolValues,
		Change: fixtureWith(`[mise.tools.fixture-x]`, `requires = "x.toml"`, `version = "latest"`), Message: msgOldRequires})
	rc.Run(t, rc.Spec{Name: "the old files of a mise entry", Code: 2, Values: noToolValues,
		Change: fixtureWith(`[mise.tools.fixture-x]`, `files = [ "a.txt" ]`, `version = "latest"`), Message: msgOldFiles})
	rc.Run(t, rc.Spec{Name: "an old requires and an old files in one entry", Code: 2, Values: noToolValues,
		Change: fixtureWith(`[mise.tools.fixture-x]`, `files = [ "a.txt" ]`, `requires = "x.toml"`, `version = "latest"`),
		Check: func(c *rc.Case) {
			c.CheckThat("an old requires and an old files: the requires message", c.LogHas(msgOldRequires))
			c.CheckThat("an old requires and an old files: the files message", c.LogHas(msgOldFiles))
		}})
	rc.Run(t, rc.Spec{Name: "an old files in two entries", Code: 2, Values: noToolValues,
		Change: fixtureWith(`[mise.tools.fixture-a]`, `files = [ "a.txt" ]`, `version = "latest"`,
			`[mise.tools.fixture-b]`, `files = [ "a.txt" ]`, `version = "latest"`),
		Check: func(c *rc.Case) {
			c.CheckThat("an old files in two entries: one message", c.LogCount("files in [mise.tools.", 1))
			c.CheckThat("an old files in two entries: the first entry", c.LogHas("files in [mise.tools.fixture-a] is not supported"))
		}})
	rc.Run(t, rc.Spec{Name: "seed and [seed_requires]", Code: 2, Values: noToolValues,
		Change: fixtureWith(`seed = [ "a.txt" ]`, `[seed_requires]`, `"a.txt" = "package.json"`), Message: msgOldSeed,
		Check: func(c *rc.Case) {
			c.CheckThat("seed and [seed_requires]: the second message", c.LogHas(msgOldTable))
		}})
}

// The generated config of the tool fixture already exists in the consumer,
// so the dry run shows its diff.
func fixtureConf(c *rc.Case) {
	consumerFile(c, ".config/mise/conf.d/fixture/mise.dev.toml", "# x")
}

func miseFoundTwo(c *rc.Case) {
	fixtureTool(c)
	fixtureToml(c, `[mise.tools.fixture-x]`, `requires_found = [ "marker.toml", "other.toml" ]`, `version = "latest"`)
	fixtureConf(c)
	markerAt(c, ".")
}

func miseEnvFound(c *rc.Case) {
	fixtureTool(c)
	fixtureToml(c, `[mise_env.dev.tools.fixture-x]`, `requires_found = [ "marker.toml", "other.toml" ]`, `version = "latest"`)
	fixtureConf(c)
}

// The keys that are valid only in [mise.tools], and requires_found of a mise
// entry, also in [mise_env.<env>.tools] (D40).
func TestMiseKeys(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "[mise_env.dev.tools] requires and files", Code: 2, Values: noToolValues,
		Change: fixtureWith(`[mise_env.dev.tools.fixture-x]`, `files = [ "a.txt" ]`, `requires = "x.toml"`, `version = "latest"`),
		Check: func(c *rc.Case) {
			c.CheckThat("[mise_env.dev.tools]: requires", c.LogHas("fixture/tools.toml: requires of mise tool fixture-x in [mise_env.dev.tools] is valid only in [mise.tools]"))
			c.CheckThat("[mise_env.dev.tools]: files", c.LogHas("fixture/tools.toml: files of mise tool fixture-x in [mise_env.dev.tools] is valid only in [mise.tools]"))
		}})
	// requires_found of a [mise_env.<env>.tools] entry (D40): any-of, and the
	// written entry has no requires_found key.
	rc.Run(t, rc.Spec{Name: "[mise_env.dev.tools] requires_found, one of two names found", Values: fixtureValues,
		Change: func(c *rc.Case) {
			miseEnvFound(c)
			consumerFile(c, "other.toml")
		}, Message: "[tools.fixture-x]",
		Check: func(c *rc.Case) {
			checkLogLacks(c, "[mise_env.dev.tools] requires_found, one of two names found", "requires_found")
		}})
	rc.Run(t, rc.Spec{Name: "[mise_env.dev.tools] requires_found, no name found", Values: fixtureValues,
		Change: miseEnvFound,
		Check: func(c *rc.Case) {
			checkLogLacks(c, "[mise_env.dev.tools] requires_found, no name found", "fixture-x")
		}})
	rc.Run(t, rc.Spec{Name: "[mise_env.dev.tools] requires_found, an empty list", Code: 2, Values: noToolValues,
		Change:  fixtureWith(`[mise_env.dev.tools.fixture-x]`, `requires_found = []`, `version = "latest"`),
		Message: "fixture/tools.toml: requires_found of mise tool fixture-x in [mise_env.dev.tools] must be a non-empty list of file names"})
	rc.Run(t, rc.Spec{Name: "requires_found of a mise entry, a name with a /", Code: 2, Values: noToolValues,
		Change:  fixtureWith(`[mise.tools.fixture-x]`, `requires_found = [ "a/x.toml" ]`, `version = "latest"`),
		Message: "fixture/tools.toml: requires_found of mise tool fixture-x: 'a/x.toml' must be a file name, with no /"})
	rc.Run(t, rc.Spec{Name: "requires_found of a mise entry, an empty list", Code: 2, Values: noToolValues,
		Change:  fixtureWith(`[mise.tools.fixture-x]`, `requires_found = []`, `version = "latest"`),
		Message: "fixture/tools.toml: requires_found of mise tool fixture-x must be a non-empty list of file names"})
	// requires_found is any-of (D38): one of the two names is enough.
	rc.Run(t, rc.Spec{Name: "requires_found of a mise entry, one of two names found", Values: fixtureValues,
		Change: miseFoundTwo, Message: "[tools.fixture-x]",
		Check: func(c *rc.Case) {
			checkLogLacks(c, "requires_found of a mise entry, one of two names found", "requires_found")
		}})
	rc.Run(t, rc.Spec{Name: "requires_found of a mise entry, both names found", Values: fixtureValues,
		Change: func(c *rc.Case) {
			miseFoundTwo(c)
			consumerFile(c, "other.toml")
		}, Message: "[tools.fixture-x]",
		Check: func(c *rc.Case) {
			checkLogLacks(c, "requires_found of a mise entry, both names found", "requires_found")
		}})
}

// filesTable gives a change: the tool fixture with a.txt and a
// [files."a.txt"] table with these lines.
func filesTable(lines ...string) func(c *rc.Case) {
	return func(c *rc.Case) {
		fixtureTxt(c)
		fixtureToml(c, append([]string{`[files."a.txt"]`}, lines...)...)
	}
}

// The rules of a [files] table. The tool `fixture` is not selected.
func TestFilesTable(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "[files], an unknown key", Code: 2, Values: noToolValues,
		Change: filesTable(`owner = "me"`), Message: `fixture/tools.toml: [files."a.txt"] owner is not a key of a files entry`})
	rc.Run(t, rc.Spec{Name: "[files], a table without a source", Code: 2, Values: noToolValues,
		Change: func(c *rc.Case) {
			fixtureTxt(c)
			fixtureToml(c, `[files."nope.txt"]`, `mode = "seed"`)
		}, Message: `fixture/tools.toml: [files."nope.txt"] needs source, a path below files/ with no .. part`})
	rc.Run(t, rc.Spec{Name: "[files], a source that is not there", Code: 2, Values: noToolValues,
		Change:  fixtureWith(`[files."nope.txt"]`, `source = "nope.txt"`),
		Message: `fixture/tools.toml: [files."nope.txt"] source: files/nope.txt is not there`})
	rc.Run(t, rc.Spec{Name: "[files], a source with a .. part", Code: 2, Values: noToolValues,
		Change:  fixtureWith(`[files."up.txt"]`, `source = "../tools.toml"`),
		Message: `fixture/tools.toml: [files."up.txt"] needs source, a path below files/ with no .. part`})
	rc.Run(t, rc.Spec{Name: "a file below files/ without an entry", Code: 2, Values: noToolValues,
		Change: func(c *rc.Case) {
			fixtureTool(c)
			fixtureFile(c, "files/x.txt", "x")
		}, Message: `fixture/tools.toml: files/x.txt has no entry - write a [files."<path>"] table with source = "x.txt"`})
	rc.Run(t, rc.Spec{Name: "the old form: a file outside files/ and tasks/", Code: 2, Values: noToolValues, OldForm: true,
		Change:  fixtureTxt,
		Message: "fixture/tools.toml: a.txt is not below files/ or tasks/ - move it, and write an entry for it (the old form of a tool folder is gone)"})
	rc.Run(t, rc.Spec{Name: "[files], a target that is not a table", Code: 2, Values: noToolValues, OldForm: true,
		Change: func(c *rc.Case) {
			fixtureTxt(c)
			fixtureToml(c, `[files]`, `"a.txt" = 1`)
		}, Message: `fixture/tools.toml: [files."a.txt"] must be a table`})
	rc.Run(t, rc.Spec{Name: "[files], an unknown mode", Code: 2, Values: noToolValues,
		Change: filesTable(`mode = "owned"`), Message: `fixture/tools.toml: [files."a.txt"] mode must be "copy", "template", "seed" or "merge"`})
	rc.Run(t, rc.Spec{Name: "[files], an unknown format", Code: 2, Values: noToolValues,
		Change: filesTable(`format = "ini"`), Message: `fixture/tools.toml: [files."a.txt"] format must be "text", "json", "jsonc", "yaml" or "toml"`})
	rc.Run(t, rc.Spec{Name: "[files], requires with a .. part at the start", Code: 2, Values: noToolValues,
		Change:  filesTable(`requires = [ "../a" ]`),
		Message: `fixture/tools.toml: [files."a.txt"] requires: '../a' must be a path relative to the repository root, with no .. part`})
	rc.Run(t, rc.Spec{Name: "[files], requires with a .. part inside", Code: 2, Values: noToolValues,
		Change:  filesTable(`requires = [ "x/../a" ]`),
		Message: `fixture/tools.toml: [files."a.txt"] requires: 'x/../a' must be a path relative to the repository root, with no .. part`})
	rc.Run(t, rc.Spec{Name: "[files], requires with an absolute path", Code: 2, Values: noToolValues,
		Change:  filesTable(`requires = [ "/a" ]`),
		Message: `fixture/tools.toml: [files."a.txt"] requires: '/a' must be a path relative to the repository root, with no .. part`})
	rc.Run(t, rc.Spec{Name: "[files], requires an empty list", Code: 2, Values: noToolValues,
		Change: filesTable(`requires = []`), Message: `fixture/tools.toml: [files."a.txt"] requires must be a non-empty list of paths`})
	rc.Run(t, rc.Spec{Name: "[files], requires not a list", Code: 2, Values: noToolValues,
		Change: filesTable(`requires = 1`), Message: `fixture/tools.toml: [files."a.txt"] requires must be a non-empty list of paths`})
	rc.Run(t, rc.Spec{Name: "[files], requires_found with a /", Code: 2, Values: noToolValues,
		Change:  filesTable(`requires_found = [ "a/b" ]`),
		Message: `fixture/tools.toml: [files."a.txt"] requires_found: 'a/b' must be a file name, with no /`})
	rc.Run(t, rc.Spec{Name: "[files], requires_found an empty list", Code: 2, Values: noToolValues,
		Change:  filesTable(`requires_found = []`),
		Message: `fixture/tools.toml: [files."a.txt"] requires_found must be a non-empty list of file names`})
	rc.Run(t, rc.Spec{Name: "[files], two errors in one table", Code: 2, Values: noToolValues,
		Change: filesTable(`format = "ini"`, `mode = "owned"`), Message: `[files."a.txt"] format must be`,
		Check: func(c *rc.Case) {
			c.CheckThat("two errors in a [files] table: the second message", c.LogHas(`[files."a.txt"] mode must be`))
		}})
}
