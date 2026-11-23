// Package selection tests the old grammar of the selection commands, the
// clean-target rule of a sync, questions, list and ls, the user config file,
// the config verb (get, set, unset and show) and settings. The cases came
// from test:repo, lines 3756 to 4112.
package selection

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

func TestMain(m *testing.M) { rc.Main(m) }

const devValues = `environments = [ "dev" ]`

const dirtyLine = "# a change that is not committed"

// engineSeq runs the engine with each argument list in order. It stops at the
// first exit code that is not 0 and gives that code (cmd || exit).
func engineSeq(c *rc.Case, calls ...[]string) int {
	for _, a := range calls {
		if code := c.Engine(a...); code != 0 {
			return code
		}
	}
	return 0
}

// unsetSettings removes the variables of the settings of the engine from the
// environment of the case.
func unsetSettings(c *rc.Case) {
	c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_CACHE", "BOOTSTRAP_CLONE", "BOOTSTRAP_JOBS")
}

// configHas tells whether a line of the output, with each run of spaces as
// one space, is the words of <words> joined by one space (config_has).
func configHas(c *rc.Case, words ...string) bool {
	want := strings.Join(words, " ")
	sp := regexp.MustCompile(` +`)
	for _, l := range strings.Split(c.LogText(), "\n") {
		if sp.ReplaceAllString(l, " ") == want {
			return true
		}
	}
	return false
}

// logIs tells whether the output, without the newlines at the end, is <s>.
func logIs(c *rc.Case, s string) bool {
	return strings.TrimRight(c.LogText(), "\n") == s
}

// logJSON decodes the output of the case as one JSON value.
func logJSON(c *rc.Case) (map[string]any, bool) {
	var m map[string]any
	if err := json.Unmarshal([]byte(c.LogText()), &m); err != nil {
		return nil, false
	}
	return m, true
}

// fileHasLine tells whether a line of the file <path> is <line>.
func fileHasLine(path, line string) bool {
	for _, l := range strings.Split(harness.ReadFile(path), "\n") {
		if l == line {
			return true
		}
	}
	return false
}

func oldGrammar(args string) func(c *rc.Case) int {
	return func(c *rc.Case) int { return c.Engine(strings.Fields(args)...) }
}

func TestOldGrammar(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "the old grammar: add node", Code: 2, Values: devValues, Call: oldGrammar("add node"),
		Message: "add needs the kind word tool or pack before the names - use: bootstrap add tool node"})
	rc.Run(t, rc.Spec{Name: "the old grammar: remove node", Code: 2, Values: devValues, Call: oldGrammar("remove node"),
		Message: "remove needs the kind word tool before the names - use: bootstrap remove tool node"})
	rc.Run(t, rc.Spec{Name: "the old grammar: replace pnpm aube", Code: 2, Values: devValues, Call: oldGrammar("replace pnpm aube"),
		Message: "replace needs the kind word tool before the names - use: bootstrap replace tool pnpm aube"})
	rc.Run(t, rc.Spec{Name: "remove pack is not a command", Code: 2, Values: devValues, Call: oldGrammar("remove pack base"),
		Message: "remove takes the kind word tool only - there is no remove pack"})
}

// syncDirty syncs, commits the consumer, changes a target and syncs with
// <args>.
func syncDirty(args ...string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		harness.AppendFile(c.T, c.Path(".config/mise.toml"), dirtyLine+"\n")
		return c.Engine(append([]string{"sync"}, args...)...)
	}
}

func checkDirtyKept(c *rc.Case) {
	c.CheckThat("the dirty target keeps its change", fileHasLine(c.Path(".config/mise.toml"), dirtyLine))
}

func listBase(c *rc.Case) int { return c.Engine("list") }

func TestSyncAndList(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "sync stops on a target with changes that are not committed", Code: 2, Values: devValues,
		Call: syncDirty(), Message: ".config/mise.toml has changes that are not committed", Check: checkDirtyKept})
	rc.Run(t, rc.Spec{Name: "sync --dry-run names a target with changes that are not committed", Values: devValues,
		Call:    syncDirty("--dry-run"),
		Message: ".config/mise.toml has changes that are not committed; a run without --dry-run stops", Check: checkDirtyKept})
	rc.Run(t, rc.Spec{Name: "sync with a values file with changes that are not committed", Values: devValues,
		Call: func(c *rc.Case) int {
			harness.AppendFile(c.T, c.ToolsValues, dirtyLine+"\n")
			return c.Engine("sync")
		}, Message: "Synced: "})
	rc.Run(t, rc.Spec{Name: "sync --jobs 1", Values: devValues,
		Call: func(c *rc.Case) int { return c.Engine("sync", "--dry-run", "--jobs", "1") }, Message: "Dry run: "})
	rc.Run(t, rc.Spec{Name: "questions --json without questions", Values: devValues,
		Call:    func(c *rc.Case) int { return c.Engine("questions", "--json") },
		Message: `{"contract":1,"command":"questions","ok":true,"code":0,"data":[],"errors":[],"warnings":[]}`})
	rc.Run(t, rc.Spec{Name: "list: a dependency below its tool", Values: devValues, Call: listBase, Message: "└── python"})
	rc.Run(t, rc.Spec{Name: "list: a required member of base with no table is a root", Values: devValues, Call: listBase,
		Message: "dprint              * required",
		Check:   func(c *rc.Case) { c.CheckThat("list: no root of a pack", c.LogLacks("(pack)")) }})
	rc.Run(t, rc.Spec{Name: "ls: a required tool", Values: devValues,
		Call: func(c *rc.Case) int { return c.Engine("ls") }, Message: "* required"})
}

// configBad writes a config file with <lines> and points BOOTSTRAP_CONFIG at
// it.
func configBad(c *rc.Case, lines ...string) {
	harness.WriteFile(c.T, filepath.Join(c.Dir, "config.toml"), harness.Lines(lines...))
	c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_CONFIG="+filepath.Join(c.Dir, "config.toml"))
}

// engineSplit runs the engine with stdout to the log of the case and stderr
// to the file <stderr>.
func engineSplit(c *rc.Case, stderr string, args ...string) int {
	logf, err := os.OpenFile(c.Log, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		c.T.Fatal(err)
	}
	defer func() { _ = logf.Close() }()
	errf, err := os.Create(stderr)
	if err != nil {
		c.T.Fatal(err)
	}
	defer func() { _ = errf.Close() }()
	cmd := exec.Command(harness.Bin, args...)
	cmd.Dir = harness.Root
	cmd.Env = c.Env
	cmd.Stdout, cmd.Stderr = logf, errf
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		return 127
	}
	return 0
}

func TestConfigFile(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "config: the source of each value", Values: devValues,
		Call: func(c *rc.Case) int {
			harness.WriteFile(c.T, filepath.Join(c.Dir, "config.toml"), harness.Lines(
				"jobs = 3", `cache_dir = "~/cache-from-file"`, `url = "https://example.com/file.git"`))
			unsetSettings(c)
			c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_CONFIG="+filepath.Join(c.Dir, "config.toml"),
				"BOOTSTRAP_URL=https://example.com/variable.git")
			return c.Engine("settings", "--jobs", "5")
		},
		Check: func(c *rc.Case) {
			home := harness.GetEnv(c.Env, "HOME")
			c.CheckThat("config: the file and its source", configHas(c, "file", filepath.Join(c.Dir, "config.toml"), "variable", "BOOTSTRAP_CONFIG"))
			c.CheckThat("config: a flag over the file", configHas(c, "jobs", "5", "flag", "--jobs"))
			c.CheckThat("config: a value of the file, ~/ expanded", configHas(c, "cache_dir", home+"/cache-from-file", "file"))
			c.CheckThat("config: a variable over the file", configHas(c, "url", "https://example.com/variable.git", "variable", "BOOTSTRAP_URL"))
			c.CheckThat("config: a default with no value", configHas(c, "clone", "-", "default"))
		}})
	rc.Run(t, rc.Spec{Name: "config --config over BOOTSTRAP_CONFIG", Values: devValues,
		Call: func(c *rc.Case) int {
			harness.WriteFile(c.T, filepath.Join(c.Dir, "other.toml"), "jobs = 7\n")
			c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_JOBS")
			return c.Engine("settings", "--config", filepath.Join(c.Dir, "other.toml"))
		},
		Check: func(c *rc.Case) {
			c.CheckThat("config --config: the file and its source", configHas(c, "file", filepath.Join(c.Dir, "other.toml"), "flag", "--config"))
			c.CheckThat("config --config: a value of that file", configHas(c, "jobs", "7", "file"))
		}})
	rc.Run(t, rc.Spec{Name: "config: a missing file gives the defaults", Values: devValues,
		Call: func(c *rc.Case) int {
			unsetSettings(c)
			c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_CONFIG="+filepath.Join(c.Dir, "missing", "config.toml"))
			return engineSplit(c, filepath.Join(c.Dir, "stderr"), "settings")
		},
		Check: func(c *rc.Case) {
			cache := harness.GetEnv(c.Env, "XDG_CACHE_HOME")
			if cache == "" {
				cache = harness.GetEnv(c.Env, "HOME") + "/.cache"
			}
			fi, err := os.Stat(filepath.Join(c.Dir, "stderr"))
			c.CheckThat("config, a missing file: (missing)",
				configHas(c, "file", filepath.Join(c.Dir, "missing", "config.toml")+" (missing)", "variable", "BOOTSTRAP_CONFIG"))
			c.CheckThat("config, a missing file: the default jobs",
				regexp.MustCompile(`(?m)^jobs +[0-9]+ +default$`).MatchString(c.LogText()))
			c.CheckThat("config, a missing file: the default cache_dir", configHas(c, "cache_dir", cache+"/bootstrap", "default"))
			c.CheckThat("config, a missing file: the default url", configHas(c, "url", "https://github.com/virajp/bootstrap.git", "default"))
			c.CheckThat("config, a missing file: the default clone", configHas(c, "clone", "-", "default"))
			c.CheckThat("config, a missing file: no message", err == nil && fi.Size() == 0)
		}})
	rc.Run(t, rc.Spec{Name: "a config file with a verbose key", Code: 2, Values: devValues,
		Call: func(c *rc.Case) int {
			configBad(c, "verbose = true")
			return c.Engine("sync", "--dry-run")
		}, Message: "verbose: unknown key",
		Check: func(c *rc.Case) {
			c.CheckThat("a verbose key: names the file and the key",
				c.LogHas("[ERROR] "+filepath.Join(c.Dir, "config.toml")+": verbose: unknown key"))
			c.CheckThat("a verbose key: the sync does not start", c.LogLacks("Dry run: "))
		}})
	rc.Run(t, rc.Spec{Name: "a config file with jobs as a string", Code: 2, Values: devValues,
		Call: func(c *rc.Case) int {
			configBad(c, `jobs = "4"`)
			return c.Engine("settings")
		}, Message: "jobs: needs a number above 0",
		Check: func(c *rc.Case) {
			c.CheckThat("jobs as a string: names the file and the key",
				c.LogHas("[ERROR] "+filepath.Join(c.Dir, "config.toml")+": jobs: needs a number above 0, not a string"))
		}})
	rc.Run(t, rc.Spec{Name: "a config file that is not TOML", Code: 2, Values: devValues,
		Call: func(c *rc.Case) int {
			configBad(c, "jobs = = 4")
			return c.Engine("settings")
		}, Message: "not valid TOML",
		Check: func(c *rc.Case) {
			c.CheckThat("not TOML: names the file", c.LogHas("[ERROR] "+filepath.Join(c.Dir, "config.toml")+": not valid TOML"))
		}})
	// clone in the file: the sync uses that clone. The URL names no remote,
	// so a fetch into the cache fails.
	rc.Run(t, rc.Spec{Name: "clone in the config file, without BOOTSTRAP_CLONE", Values: devValues,
		Call: func(c *rc.Case) int {
			harness.WriteFile(c.T, filepath.Join(c.Dir, "config.toml"), `clone = "`+c.Clone+`"`+"\n")
			c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_CLONE", "BOOTSTRAP_CONFIG="+filepath.Join(c.Dir, "config.toml"),
				"BOOTSTRAP_URL=file://"+filepath.Join(c.Dir, "no-such-remote"))
			return c.Engine("sync", "--dry-run")
		}, Message: "Dry run: "})
}

// configCmd runs `bootstrap config <args>` with BOOTSTRAP_BIN at the engine
// of the test (config_cmd).
func configCmd(args string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_BIN="+harness.Bin)
		return c.Engine(append([]string{"config"}, strings.Fields(args)...)...)
	}
}

// configNoCatalog is configCmd with no clone and no cache.
func configNoCatalog(args string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_CLONE", "BOOTSTRAP_CACHE="+filepath.Join(c.Dir, "no-cache"),
			"BOOTSTRAP_URL=file://"+filepath.Join(c.Dir, "no-such-remote"))
		return configCmd(args)(c)
	}
}

func valueJSON(c *rc.Case) {
	c.CheckThat("config get --format json: the envelope", logIs(c, `{"key":"repo.repo_name","value":"test","source":"file"}`))
}

func valueTOML(c *rc.Case) {
	c.CheckThat("config get --format toml: the envelope", logIs(c, "key = \"repo.repo_name\"\nvalue = \"test\"\nsource = \"file\""))
}

func valueYAML(c *rc.Case) {
	c.CheckThat("config get --format yaml: the envelope", logIs(c, "key: repo.repo_name\nvalue: test\nsource: file"))
}

func TestConfigGet(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "config get --format json", Values: devValues, Call: configCmd("get repo.repo_name --format json"), Check: valueJSON})
	rc.Run(t, rc.Spec{Name: "config get, json by default", Values: devValues, Call: configCmd("get repo.repo_name"), Check: valueJSON})
	rc.Run(t, rc.Spec{Name: "config get --format toml", Values: devValues, Call: configCmd("get repo.repo_name --format toml"), Check: valueTOML})
	rc.Run(t, rc.Spec{Name: "config get --format yaml", Values: devValues, Call: configCmd("get repo.repo_name --format yaml"), Check: valueYAML})
	rc.Run(t, rc.Spec{Name: "config get of a key that is not set", Values: devValues, Call: configCmd("get repo.product_name --format json"),
		Message: `{"key":"repo.product_name","value":null,"source":"default"}`})
	rc.Run(t, rc.Spec{Name: "config get of a key with a default", Values: devValues, Call: configCmd("get repo.commits.style --format yaml"),
		Message: "source: default"})
	rc.Run(t, rc.Spec{Name: "config get, a quoted segment and a default", Values: devValues, Call: configCmd(`get repo.branches."main".landing`),
		Message: `"value":"direct-merge","source":"default"}`})
	rc.Run(t, rc.Spec{Name: "config get of a table", Values: devValues + "\n[branches.main]\nlanding = \"pr\"", Call: configCmd("get repo.branches.main"),
		Message: `{"key":"repo.branches.main","value":{"landing":"pr"},"source":"file"}`})
	// A whole tool table carries env: the file value, else default_envs.
	rc.Run(t, rc.Spec{Name: "config get of a tool table, with env", Values: devValues + "\n[tools.node]", Call: configCmd("get tools.node"),
		Message: `{"key":"tools.node","value":{"audit":"osv-scanner","env":["dev","ci"]},"source":"file"}`})
	rc.Run(t, rc.Spec{Name: "config get of env, default_envs", Values: devValues + "\n[tools.node]", Call: configCmd("get tools.node.env"),
		Message: `{"key":"tools.node.env","value":["dev","ci"],"source":"default"}`})
	// A tool that the catalog does not have, or that is not selected: value
	// null, source unselected.
	rc.Run(t, rc.Spec{Name: "config get of a tool that bootstrap does not have", Values: devValues, Call: configCmd("get tools.nope.x"),
		Message: `{"key":"tools.nope.x","value":null,"source":"unselected"}`})
	rc.Run(t, rc.Spec{Name: "config get of a tool that is not selected", Values: devValues, Call: configCmd("get tools.go.env"),
		Message: `{"key":"tools.go.env","value":null,"source":"unselected"}`})
	// A value whose winner is a derive script, here the git script of the
	// base pack, which reads repo.core_branches.
	rc.Run(t, rc.Spec{Name: "config get of a derived value", Values: devValues, Call: configCmd("get tools.git.branch_pair"),
		Message: `{"key":"tools.git.branch_pair","value":false,"source":"derived"}`})
	rc.Run(t, rc.Spec{Name: "config get of a derived list", Values: "environments = [ \"dev\" ]\ncore_branches = [ \"main\", \"develop\" ]",
		Call:    configCmd("get tools.git.protected_branches"),
		Message: `{"key":"tools.git.protected_branches","value":["main","develop"],"source":"derived"}`})
	// No catalog: a tools.* get exits 2 with no value; a repo.* get needs
	// none.
	rc.Run(t, rc.Spec{Name: "config get of a tools key, no catalog", Code: 2, Values: devValues + "\n[tools.node]",
		Call: configNoCatalog("get tools.node.env"), Message: "config get tools.node.env needs the catalog, and no clone is cached",
		Check: func(c *rc.Case) {
			c.CheckThat("config get, no catalog: no envelope", !strings.Contains(c.LogText(), `"value"`))
		}})
	rc.Run(t, rc.Spec{Name: "config get of a repo key, no catalog", Values: devValues, Call: configNoCatalog("get repo.repo_name"),
		Message: `{"key":"repo.repo_name","value":"test","source":"file"}`})
	// A value that virajp-plugins contributes to the claude tool: with no
	// file entry, its source is contributed; with a file entry too, it is
	// file, and contributed_by names the contributor.
	rc.Run(t, rc.Spec{Name: "config get of a contributed tools value", Values: devValues + "\n[tools.virajp-plugins]",
		Call:    configCmd("get tools.claude.plugins.user --format json"),
		Message: `{"key":"tools.claude.plugins.user","value":{"vwf@virajp-plugins":"virajp/claude-plugins"},"source":"contributed","contributed_by":["virajp-plugins"]}`})
	rc.Run(t, rc.Spec{Name: "config get of a tools value with file and contributed entries",
		Values: devValues + "\n[tools.virajp-plugins]\n[tools.claude.plugins.user]\n\"own@acme\" = \"acme/own\"",
		Call:   configCmd("get tools.claude.plugins.user --format json"), Message: `"source":"file","contributed_by":["virajp-plugins"]}`,
		Check: func(c *rc.Case) {
			m, ok := logJSON(c)
			want := map[string]any{"own@acme": "acme/own", "vwf@virajp-plugins": "virajp/claude-plugins"}
			c.CheckThat("config get: the file entry and the contributed entry", ok && reflect.DeepEqual(m["value"], want))
		}})
	rc.Run(t, rc.Spec{Name: "config get of a release level, no releases.toml", Values: devValues, Call: configCmd("get releases.api"),
		Message: `{"key":"releases.api","value":"NONE","source":"default"}`})
	rc.Run(t, rc.Spec{Name: "config get of an unknown key", Code: 2, Values: devValues, Call: configCmd("get repo.nope"),
		Message: "repo.nope is not a key of the values"})
	rc.Run(t, rc.Spec{Name: "config get without a values folder", Code: 2, Values: devValues,
		Change: func(c *rc.Case) { _ = os.RemoveAll(c.Values) }, Call: configCmd("get repo.repo_name"), Message: "run bootstrap migrate"})
	rc.Run(t, rc.Spec{Name: "config alone", Code: 2, Values: devValues, Call: configCmd(""),
		Message: "config needs get, set, unset or show; the settings of the engine are bootstrap settings"})
}

// keepRepo copies repo.toml before a refused write.
func keepRepo(c *rc.Case) {
	harness.WriteFile(c.T, filepath.Join(c.Dir, "repo.before"), harness.ReadFile(c.RepoValues))
}

func refused(args ...string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		keepRepo(c)
		return c.Engine(args...)
	}
}

func checkUnchanged(c *rc.Case) {
	before, err1 := os.ReadFile(filepath.Join(c.Dir, "repo.before"))
	after, err2 := os.ReadFile(c.RepoValues)
	c.CheckThat("a refused set: repo.toml stays", err1 == nil && err2 == nil && string(before) == string(after))
}

func TestConfigSet(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "config set, then get", Values: devValues,
		Call: func(c *rc.Case) int {
			return engineSeq(c, []string{"config", "set", "repo.product_name", "acme"},
				[]string{"config", "get", "repo.product_name", "--format", "json"})
		}, Message: `{"key":"repo.product_name","value":"acme","source":"file"}`,
		Check: func(c *rc.Case) {
			c.CheckThat("config set: the sync line", c.LogHas("run bootstrap sync to apply it"))
			c.CheckThat("config set: the value in repo.toml", c.TomlQ(`.product_name == "acme"`, c.RepoValues))
		}})
	rc.Run(t, rc.Spec{Name: "config set of an unknown key", Code: 2, Values: devValues,
		Call: refused("config", "set", "repo.nope", "x"), Message: "repo.nope is not a key of the values", Check: checkUnchanged})
	rc.Run(t, rc.Spec{Name: "config set that breaks a rule between keys", Code: 2, Values: devValues,
		Call:    refused("config", "set", "repo.branches.feature.landing", "pr"),
		Message: "feature is not in core_branches - add it there, or remove the table", Check: checkUnchanged})
	rc.Run(t, rc.Spec{Name: "config set of a value outside its enum", Code: 2, Values: devValues,
		Call:    refused("config", "set", "repo.branches.main.landing", "merge"),
		Message: "is not one of pr, direct-merge, direct-commit", Check: checkUnchanged})
	rc.Run(t, rc.Spec{Name: "config unset of a required key", Code: 2, Values: devValues,
		Call:    refused("config", "unset", "repo.repo_name"),
		Message: "repo.repo_name is a required key of repo.toml, so config unset cannot remove it", Check: checkUnchanged})
	rc.Run(t, rc.Spec{Name: "config unset, then get", Values: "environments = [ \"dev\" ]\nproduct_name = \"acme\"",
		Call: func(c *rc.Case) int {
			return engineSeq(c, []string{"config", "unset", "repo.product_name"}, []string{"config", "get", "repo.product_name"})
		}, Message: `{"key":"repo.product_name","value":null,"source":"default"}`,
		Check: func(c *rc.Case) {
			c.CheckThat("config unset: the key is gone", !strings.Contains(harness.ReadFile(c.RepoValues), "product_name"))
		}})
	rc.Run(t, rc.Spec{Name: "config set of a tools key, then get", Values: devValues + "\n[tools.node]",
		Call: func(c *rc.Case) int {
			if code := c.Engine("config", "set", "tools.node.env", `["dev"]`); code != 0 {
				return code
			}
			c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_BIN="+harness.Bin)
			return c.Engine("config", "get", "tools.node.env")
		}, Message: `{"key":"tools.node.env","value":["dev"],"source":"file"}`})
	rc.Run(t, rc.Spec{Name: "config set of a tool that is not selected", Code: 2, Values: devValues,
		Call:    func(c *rc.Case) int { return c.Engine("config", "set", "tools.go.env", `["dev"]`) },
		Message: "go is not selected, so tools.go.env cannot be set - run bootstrap add tool go"})
	rc.Run(t, rc.Spec{Name: "config set of a key that the tool does not have", Code: 2, Values: devValues + "\n[tools.node]",
		Call: configCmd("set tools.node.version 1"), Message: "version is not a key of the tool node"})
}

func TestConfigReleases(t *testing.T) {
	// set keeps the higher level, unset clears it, and the first set creates
	// releases.toml with format = 3.
	rc.Run(t, rc.Spec{Name: "config set releases.x PATCH, MINOR, PATCH keeps MINOR", Values: devValues,
		Call: func(c *rc.Case) int {
			return engineSeq(c, []string{"config", "set", "releases.api", "PATCH"}, []string{"config", "set", "releases.api", "MINOR"},
				[]string{"config", "set", "releases.api", "PATCH"}, []string{"config", "get", "releases.api"})
		}, Message: `{"key":"releases.api","value":"MINOR","source":"file"}`,
		Check: func(c *rc.Case) {
			f := filepath.Join(c.Values, "releases.toml")
			c.CheckThat("releases: the file has format = 3", c.TomlQ(".format == 3", f))
			c.CheckThat("releases: MINOR stays", c.TomlQ(`.api == "MINOR"`, f))
			c.CheckThat("releases: the info line", c.LogHas("releases.api stays MINOR: it is higher than PATCH"))
		}})
	rc.Run(t, rc.Spec{Name: "config unset of a release level", Values: devValues,
		Call: func(c *rc.Case) int {
			return engineSeq(c, []string{"config", "set", "releases.api", "MAJOR"}, []string{"config", "unset", "releases.api"},
				[]string{"config", "get", "releases.api"})
		}, Message: `{"key":"releases.api","value":"NONE","source":"default"}`})
	rc.Run(t, rc.Spec{Name: "config set of a release level outside its enum", Code: 2, Values: devValues,
		Call: configCmd("set releases.api BIG"), Message: "is not one of NONE, PATCH, MINOR, MAJOR"})
}

// emptyValue tells whether <v> is null, an empty table or an empty list.
func emptyValue(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case map[string]any:
		return len(x) == 0
	case []any:
		return len(x) == 0
	}
	return false
}

func TestConfigShow(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "config show repo --format json", Values: devValues, Call: configCmd("show repo --format json"),
		Check: func(c *rc.Case) {
			m, ok := logJSON(c)
			c.CheckThat("config show repo --format json", ok && m["repo_name"] == "test" && m["format"] == float64(3))
		}})
	rc.Run(t, rc.Spec{Name: "config show tools --format toml", Values: devValues + "\n[tools.node]", Call: configCmd("show tools --format toml"),
		Message: "[tools.node]"})
	rc.Run(t, rc.Spec{Name: "config show releases, no releases.toml", Values: devValues, Call: configCmd("show releases --format yaml"),
		Check: func(c *rc.Case) { c.CheckThat("config show releases: an empty table", logIs(c, "{}")) }})
	rc.Run(t, rc.Spec{Name: "config show, all the files", Values: devValues + "\n[tools.node]", Call: configCmd("show --format json"),
		Check: func(c *rc.Case) {
			m, ok := logJSON(c)
			repo, _ := m["repo"].(map[string]any)
			tools, _ := m["tools"].(map[string]any)
			inner, _ := tools["tools"].(map[string]any)
			_, node := inner["node"]
			c.CheckThat("config show: the three files", ok && repo["repo_name"] == "test" && node && emptyValue(m["releases"]))
		}})
	rc.Run(t, rc.Spec{Name: "config show of an unknown file", Code: 2, Values: devValues, Call: configCmd("show nope"),
		Message: "config show takes repo, tools or releases, not nope"})
	rc.Run(t, rc.Spec{Name: "settings prints the settings of the engine", Values: devValues,
		Call: func(c *rc.Case) int {
			unsetSettings(c)
			return c.Engine("settings")
		}, Message: "cache_dir "})
}
