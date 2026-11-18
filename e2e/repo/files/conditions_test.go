package files

import (
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

const (
	nodeValues = `environments = [ "dev" ]
[tools.pnpm]`
	aubeValues = `environments = [ "dev" ]
[tools.aube]`
	nodeOwnSeed = "# my own settings"
)

func nodePackage(c *rc.Case)    { consumerFile(c, "package.json") }
func nodeSeedExists(c *rc.Case) { consumerFile(c, "pnpm-workspace.yaml", nodeOwnSeed) }

func checkSeedKept(c *rc.Case) {
	c.CheckThat("a seed that exists: kept seed", c.LogHas("kept seed  "+c.Path("pnpm-workspace.yaml")))
	c.CheckThat("a seed that exists: no skip seed", c.LogLacks("skip seed"))
	checkExact(c, "a seed that exists: not changed", c.Path("pnpm-workspace.yaml"), nodeOwnSeed+"\n")
}

// A seed: the pnpm-workspace.yaml of the pnpm tool, with requires =
// package.json.
func TestSeeds(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "a seed, its requires absent", Values: nodeValues,
		Check: func(c *rc.Case) {
			p := c.Path("pnpm-workspace.yaml")
			c.CheckThat("a seed, its requires absent: skip seed", c.LogHas("skip seed  "+p+" (no package.json)"))
			c.CheckThat("a seed, its requires absent: not written", c.LogLacks("would update "+p))
		}})
	rc.Run(t, rc.Spec{Name: "a seed that exists, its requires absent", Values: nodeValues,
		Change: nodeSeedExists, Check: checkSeedKept})
	rc.Run(t, rc.Spec{Name: "a seed that exists, its requires present", Values: nodeValues,
		Change: func(c *rc.Case) {
			nodePackage(c)
			nodeSeedExists(c)
		}, Check: checkSeedKept})
	rc.Run(t, rc.Spec{Name: "a seed, its requires present", Values: nodeValues, Args: realRun,
		Change: nodePackage,
		Check: func(c *rc.Case) {
			c.CheckThat("a seed, its requires present: written", notEmpty(c.Path("pnpm-workspace.yaml")))
		}})
	rc.Run(t, rc.Spec{Name: "the aube seed, its requires present", Values: aubeValues, Args: realRun,
		Change: nodePackage,
		Check: func(c *rc.Case) {
			c.CheckThat("the aube seed: written", notEmpty(c.Path("aube-workspace.yaml")))
			c.CheckThat("the aube seed: no pnpm seed", !exists(c.Path("pnpm-workspace.yaml")))
		}})
	rc.Run(t, rc.Spec{Name: "a seed, its requires_found absent", Values: fixtureValues,
		Change: func(c *rc.Case) {
			fixtureTxt(c)
			fixtureToml(c, `[files."a.txt"]`, `mode = "seed"`, `requires_found = [ "marker.toml" ]`)
		},
		Check: func(c *rc.Case) {
			c.CheckThat("a seed, its requires_found absent: skip seed", c.LogHas("skip seed  "+c.Path("a.txt")+" (no marker.toml)"))
		}})
}

// excludeIs: the top-level exclude of the pre-commit config is <value>.
func excludeIs(c *rc.Case, value string) bool {
	cmd := &harness.Cmd{Env: c.Env, Args: []string{"yq", ".exclude", c.Path(".config/pre-commit-config.yaml")}}
	code, out := cmd.Run()
	return code == 0 && strings.TrimSuffix(out, "\n") == value
}

// The top-level exclude of the pre-commit config: the default, then a
// contribution of a selected tool, then the values file.
func TestPrecommitExclude(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "pre-commit exclude, no values", Values: devValues, Args: realRun,
		Check: func(c *rc.Case) {
			c.CheckThat("exclude, only the default", excludeIs(c, `(^|/)\.config/mise/locks/`))
		}})
	rc.Run(t, rc.Spec{Name: "pre-commit exclude, default, contribution and values", Args: realRun,
		Values: `environments = [ "dev" ]
[tools.pre-commit]
exclude = [ '(^|/)vendor/', '\.snap$' ]`,
		Change: func(c *rc.Case) {
			harness.AppendFile(c.T, c.ClonePath("repo/tools/dprint/tools.toml"),
				harness.Lines("[contributes.pre-commit]", "exclude = [ '(^|/)contributed/' ]"))
		},
		Check: func(c *rc.Case) {
			c.CheckThat("exclude, all sources merged in order",
				excludeIs(c, `(^|/)\.config/mise/locks/|(^|/)contributed/|(^|/)vendor/|\.snap$`))
		}})
	rc.Run(t, rc.Spec{Name: "pre-commit exclude, a string and not a list", Code: 2, Values: `environments = [ "dev" ]
[tools.pre-commit]
exclude = "vendor"`, Message: "sets exclude to a string, but it is a list"})
}

// genFound: a generated target with requires_found.
func genFound(c *rc.Case) {
	fixtureTxt(c)
	fixtureToml(c, `[files."a.txt"]`, `requires_found = [ "marker.toml" ]`)
}

func genFound5(c *rc.Case) {
	genFound(c)
	markerAt(c, "a/b/c/d/e")
}

func genFound5Synced(c *rc.Case) {
	genFound5(c)
	copySynced(c)
	manifestFile(c, "fixture"+tab+"a.txt")
}

// A generated target with requires_found: found 3 folders down, not found 5
// folders down.
func TestRequiresFoundDepth(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "a generated target, requires_found 3 folders down", Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			genFound(c)
			markerAt(c, "a/b/c")
		},
		Check: func(c *rc.Case) {
			c.CheckThat("requires_found 3 folders down: written", hasLine(c.Path("a.txt"), "a"))
			c.CheckThat("requires_found 3 folders down: no skip", c.LogLacks("(no marker.toml)"))
		}})
	rc.Run(t, rc.Spec{Name: "a generated target, requires_found 5 folders down", Values: fixtureValues,
		Change: genFound5,
		Check: func(c *rc.Case) {
			c.CheckThat("requires_found 5 folders down: skip file", c.LogHas("skip file  "+c.Path("a.txt")+" (no marker.toml)"))
			c.CheckThat("requires_found 5 folders down: not written", c.LogLacks("would update "+c.Path("a.txt")))
		}})
	rc.Run(t, rc.Spec{Name: "a generated target, requires_found 5 folders down, a synced copy", Values: fixtureValues,
		Change: genFound5Synced,
		Check: func(c *rc.Case) {
			c.CheckThat("requires_found 5 folders down, a synced copy: would remove", c.LogHas("would remove "+c.Path("a.txt")))
			c.CheckThat("requires_found 5 folders down, a synced copy: the dry run keeps it", isFile(c.Path("a.txt")))
		}})
	rc.Run(t, rc.Spec{Name: "a generated target, requires_found 5 folders down, a synced copy, a real run", Values: fixtureValues,
		Args: realRun, Change: genFound5Synced,
		Check: func(c *rc.Case) {
			c.CheckThat("requires_found 5 folders down, a synced copy: removed", c.LogHas("removed    "+c.Path("a.txt")))
			c.CheckThat("requires_found 5 folders down, a synced copy: gone", rc.NoFile(c.Path("a.txt")))
		}})
	rc.Run(t, rc.Spec{Name: "a generated target, requires_found 5 folders down, a copy of the repo", Values: fixtureValues,
		Args: realRun,
		Change: func(c *rc.Case) {
			genFound5(c)
			copyMine(c)
			manifestFile(c, "fixture"+tab+"a.txt")
		},
		Check: func(c *rc.Case) {
			c.CheckThat("requires_found 5 folders down, a copy of the repo: skip file", c.LogHas("skip file  "+c.Path("a.txt")+" (no marker.toml)"))
			checkExact(c, "requires_found 5 folders down, a copy of the repo: stays", c.Path("a.txt"), copyOwn+"\n")
		}})
}

// bothLists: a target with requires and requires_found.
func bothLists(c *rc.Case) {
	fixtureTxt(c)
	fixtureToml(c, `[files."a.txt"]`, `requires = [ "need.txt" ]`, `requires_found = [ "marker.toml" ]`)
}

func bothNeed(c *rc.Case) {
	bothLists(c)
	consumerFile(c, "need.txt")
}

func bothMarker(c *rc.Case) {
	bothLists(c)
	markerAt(c, ".")
}

// A target with requires and requires_found: it needs every entry of both,
// and the message names the first entry that is not there.
func TestRequiresBoth(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "requires present, requires_found absent", Values: fixtureValues, Change: bothNeed,
		Check: func(c *rc.Case) {
			c.CheckThat("requires present, requires_found absent", c.LogHas("skip file  "+c.Path("a.txt")+" (no marker.toml)"))
			c.CheckThat("requires present, requires_found absent: not written", c.LogLacks("would update "+c.Path("a.txt")))
		}})
	rc.Run(t, rc.Spec{Name: "requires absent, requires_found present", Values: fixtureValues, Change: bothMarker,
		Check: func(c *rc.Case) {
			c.CheckThat("requires absent, requires_found present", c.LogHas("skip file  "+c.Path("a.txt")+" (no need.txt)"))
			c.CheckThat("requires absent, requires_found present: not written", c.LogLacks("would update "+c.Path("a.txt")))
		}})
	rc.Run(t, rc.Spec{Name: "requires and requires_found absent", Values: fixtureValues, Change: bothLists,
		Check: func(c *rc.Case) {
			c.CheckThat("both absent: the requires entry", c.LogHas("skip file  "+c.Path("a.txt")+" (no need.txt)"))
			c.CheckThat("both absent: not the requires_found entry", c.LogLacks("(no marker.toml)"))
		}})
	rc.Run(t, rc.Spec{Name: "requires and requires_found present", Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			bothNeed(c)
			markerAt(c, ".")
		},
		Check: func(c *rc.Case) { c.CheckThat("both present: written", hasLine(c.Path("a.txt"), "a")) }})
	rc.Run(t, rc.Spec{Name: "requires, the file only in a folder", Values: fixtureValues,
		Change: func(c *rc.Case) {
			bothMarker(c)
			consumerFile(c, "a/need.txt")
		},
		Check: func(c *rc.Case) {
			c.CheckThat("requires is an exact path", c.LogHas("skip file  "+c.Path("a.txt")+" (no need.txt)"))
		}})
}
