package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// The mise shim of code:lsp:check: `which` gives a path only in
// MISE_ENV=dev, and fails with LSP_WHICH_FAIL=1; it runs the command of
// `x --` itself.
const lspShim = `#!/usr/bin/env bash
case $1 in
  which)
    [[ ${MISE_ENV:-} == dev && -z ${LSP_WHICH_FAIL:-} ]] || exit 1
    printf '%s\n' "/servers/$2"
    ;;
  x)
    shift
    [[ ${1:-} != -- ]] || shift
    exec "$@"
    ;;
  *) exit 1 ;;
esac
`

// The stub engine of code:lsp:check: an envelope with the one server gopls,
// or an exit with LSP_STUB_CODE.
const lspStub = `#!/usr/bin/env bash
[[ $* == "lsp --json" ]] || exit 1
if [[ -n ${LSP_STUB_CODE:-} ]]; then
  printf '%s\n' '[ERROR] the stub fails' >&2
  exit "$LSP_STUB_CODE"
fi
printf '%s\n' '{"contract":1,"command":"lsp","ok":true,"code":0,"data":{"servers":{"gopls":{"command":"mise","args":["x","--","gopls"],"extensionToLanguage":{".go":"go"}}},"checks":{"gopls":{"command":"gopls","source":"go:golang.org/x/tools/gopls"}}},"errors":[],"warnings":[]}'
`

// lspCheckTask runs the synced task file code:lsp:check, not with mise run.
// <bin> is the engine of the task (empty is the stub).
func lspCheckTask(bin, whichFail, stubCode string) func(c *rc.Case) int {
	return synced(func(c *rc.Case) int {
		shim := filepath.Join(c.Dir, "shim")
		writeScript(c, filepath.Join(shim, "mise"), lspShim)
		stub := filepath.Join(c.Dir, "stub")
		writeScript(c, stub, lspStub)
		if bin == "" {
			bin = stub
		}
		env := harness.SetEnv(c.Env, "BOOTSTRAP_BIN="+bin, "LSP_WHICH_FAIL="+whichFail, "LSP_STUB_CODE="+stubCode)
		env = prependPath(env, shim)
		return run{dir: c.Consumer, env: env}.do(c, c.Path(tasksPath+"/code/lsp/check.py"))
	})
}

// code:lsp:check (D16, D20, D29).
func TestLspCheckTask(t *testing.T) {
	checkHeader := func(c *rc.Case) {
		c.CheckThat("code:lsp:check: the header", c.LogHas("Checking the LSP servers ..."))
	}
	rc.Run(t, rc.Spec{Name: "code:lsp:check: each server resolves", Values: devValues, Call: lspCheckTask("", "", ""),
		Message: "gopls: /servers/gopls", Check: checkHeader})
	rc.Run(t, rc.Spec{Name: "code:lsp:check: a server that does not resolve", Code: 1, Values: devValues,
		Call: lspCheckTask("", "1", ""), Message: "gopls: gopls not found - run MISE_ENV=dev mise install",
		Check: checkHeader})
	rc.Run(t, rc.Spec{Name: "code:lsp:check: lsp --json exits 2", Code: 2, Values: devValues,
		Call: lspCheckTask("", "", "2"), Message: "the stub fails"})
	rc.Run(t, rc.Spec{Name: "code:lsp:check: lsp --json exits 3", Code: 1, Values: devValues,
		Call: lspCheckTask("", "", "3"), Message: "the stub fails"})
	// A repo with no server: the task prints nothing and exits 0, with the
	// real engine. The repo has node through base, but no package.json or
	// tsconfig.json.
	rc.Run(t, rc.Spec{Name: "code:lsp:check: a repo with no server", Values: devValues,
		Call: lspCheckTask(harness.Bin, "", ""),
		Check: func(c *rc.Case) {
			fi, err := os.Stat(c.Log)
			c.CheckThat("code:lsp:check: no output without a server", err != nil || fi.Size() == 0)
		}})
}

// The mise shim of setup:ai: it runs the command of `x --` itself.
const aiMise = `#!/usr/bin/env bash
case $1 in
  x)
    shift
    [[ ${1:-} != -- ]] || shift
    exec "$@"
    ;;
  *) exit 1 ;;
esac
`

// The stub claude of setup:ai: it logs each call to $AI_STATE/calls and
// answers from the state files installed and marketplaces.
const aiClaude = `#!/usr/bin/env bash
state=$AI_STATE
printf '%s\n' "$*" >> "$state/calls"
case "$*" in
  "plugin list --json")
    jq -Rnc --arg root "$AI_ROOT" '[inputs | select(length > 0) | split("|")
      | {id: .[0], scope: .[1]} + (if .[1] == "project" then {projectPath: $root} else {} end)]' \
      < "$state/installed"
    ;;
  "plugin marketplace list --json")
    jq -Rnc '[inputs | select(length > 0) | {name: .}]' < "$state/marketplaces"
    ;;
  "plugin marketplace add --scope "*)
    case $5 in
      virajp/claude-plugins) printf '%s\n' virajp-plugins >> "$state/marketplaces" ;;
      *) printf '%s\n' "${5##*/}" >> "$state/marketplaces" ;;
    esac
    ;;
  "plugin install --scope "*)
    if [[ ${AI_FAIL:-} == install ]]; then
      printf '%s\n' 'the stub install fails' >&2
      exit 1
    fi
    printf '%s|%s\n' "$5" "$4" >> "$state/installed"
    ;;
  "plugin uninstall --scope project --prune -y "*)
    if [[ ${AI_FAIL:-} == uninstall ]]; then
      printf '%s\n' 'the stub uninstall fails' >&2
      exit 1
    fi
    grep -vFx "$7|project" "$state/installed" > "$state/installed.new" || true
    mv -- "$state/installed.new" "$state/installed"
    ;;
esac
`

// The engine of setup:ai for AI_SET_FAIL: the real engine, except that
// `config set` exits 5.
const aiEngine = `#!/usr/bin/env bash
if [[ $1 == config && $2 == set && -n ${AI_SET_FAIL:-} ]]; then
  printf '%s\n' '[ERROR] the stub set fails' >&2
  exit 5
fi
exec "$AI_REAL_BIN" "$@"
`

// A project plugin of the list, and the vwf plugin that virajp-plugins
// contributes at user scope.
const aiValues = devValues + `
[tools.virajp-plugins]
[tools.claude.plugins.project]
"lsp@tools" = "acme/tools"`

// aiRecordValues is the same list with a record.
func aiRecordValues(record string) string {
	return aiValues + "\n[tools.claude.plugins]\ninstalled = " + record
}

// aiSetup syncs the consumer, and writes the shims, the stub claude, the
// stub engine and the empty state files. It gives the code of the sync.
func aiSetup(c *rc.Case) int {
	if code := c.Synced(); code != 0 {
		return code
	}
	ai := filepath.Join(c.Dir, "ai")
	harness.WriteFile(c.T, filepath.Join(ai, "installed"), "")
	harness.WriteFile(c.T, filepath.Join(ai, "marketplaces"), "")
	writeScript(c, filepath.Join(ai, "bin", "mise"), aiMise)
	writeScript(c, filepath.Join(ai, "bin", "claude"), aiClaude)
	writeScript(c, filepath.Join(ai, "engine"), aiEngine)
	return 0
}

// aiState appends <lines> to the state file <name>.
func aiState(c *rc.Case, name string, lines ...string) {
	harness.AppendFile(c.T, filepath.Join(c.Dir, "ai", name), harness.Lines(lines...))
}

// aiRun runs the task setup:ai for run <n>, with --upgrade when <upgrade>,
// and the variables <extra>. Its output goes to the log of the case, and its
// claude calls to calls<n>.
func aiRun(c *rc.Case, n int, upgrade bool, extra ...string) int {
	jq := toolDirs(c.T)
	ai := filepath.Join(c.Dir, "ai")
	calls := filepath.Join(ai, "calls")
	harness.WriteFile(c.T, calls, "")
	harness.AppendFile(c.T, c.Log, fmt.Sprintf("--- run %d\n", n))
	env := harness.SetEnv(c.Env, "BOOTSTRAP_BIN="+harness.Bin, "AI_REAL_BIN="+harness.Bin, "AI_STATE="+ai,
		"AI_ROOT="+c.Consumer, fmt.Sprintf("usage_upgrade=%t", upgrade))
	env = prependPath(env, filepath.Join(ai, "bin"), jq)
	env = harness.SetEnv(env, extra...)
	got := run{dir: c.Consumer, env: env}.do(c, c.Path(tasksPath+"/setup/ai.py"))
	harness.WriteFile(c.T, fmt.Sprintf("%s%d", calls, n), harness.ReadFile(calls))
	return got
}

func aiCalls(c *rc.Case, n int) string {
	return filepath.Join(c.Dir, "ai", fmt.Sprintf("calls%d", n))
}

// aiCalled: run <n> made the claude call <call>.
func aiCalled(c *rc.Case, n int, call string) bool { return fileHasLine(aiCalls(c, n), call) }

var aiChangeRe = regexp.MustCompile(`(?m)^plugin (install|uninstall|update|marketplace (add|update)) `)

// aiNoChange: run <n> made no install, update, uninstall or add call.
func aiNoChange(c *rc.Case, n int) bool {
	return !aiChangeRe.MatchString(harness.ReadFile(aiCalls(c, n)))
}

// aiRecord: the record tools.claude.plugins.installed of the values file,
// its entries joined with commas; no key is an empty record.
func aiRecord(c *rc.Case, record string) bool {
	return c.TomlQ(`((.tools.claude.plugins.installed // []) | join(",")) == "`+record+`"`, "")
}

// setup:ai (D2 to D5, D13, D18, D19, D25 to D27, D29, D34): the case runs the
// synced task file, with the real engine in BOOTSTRAP_BIN and the
// clone of the case in BOOTSTRAP_CLONE (D27). No case uses the network or a
// real claude.
func TestSetupAI(t *testing.T) {
	// 1. A first run installs; 2. a second run changes nothing (D4); 3.
	// --upgrade makes the update calls (D4).
	rc.Run(t, rc.Spec{Name: "setup:ai: install, a second run, then --upgrade", Values: aiValues,
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			if code := aiRun(c, 1, false); code != 0 {
				return code
			}
			if code := aiRun(c, 2, false); code != 0 {
				return code
			}
			return aiRun(c, 3, true)
		},
		Message: "Installing vwf@virajp-plugins (user)",
		Check: func(c *rc.Case) {
			c.CheckThat("setup:ai: the first run adds the marketplace of vwf",
				aiCalled(c, 1, "plugin marketplace add --scope user virajp/claude-plugins"))
			c.CheckThat("setup:ai: the first run installs vwf at user scope",
				aiCalled(c, 1, "plugin install --scope user vwf@virajp-plugins"))
			c.CheckThat("setup:ai: the first run installs the listed project plugin",
				aiCalled(c, 1, "plugin install --scope project lsp@tools"))
			c.CheckThat("setup:ai: the project plugin is recorded", aiRecord(c, "lsp@tools"))
			c.CheckThat("setup:ai: the second run changes nothing (D4)", aiNoChange(c, 2))
			c.CheckThat("setup:ai: --upgrade updates the marketplaces (D4)", aiCalled(c, 3, "plugin marketplace update"))
			c.CheckThat("setup:ai: --upgrade updates vwf at user scope (D4)",
				aiCalled(c, 3, "plugin update --scope user vwf@virajp-plugins"))
			c.CheckThat("setup:ai: --upgrade updates the project plugin (D4)",
				aiCalled(c, 3, "plugin update --scope project lsp@tools"))
			c.CheckThat("setup:ai: --upgrade installs nothing", !aiCalled(c, 3, "plugin install --scope project lsp@tools"))
		}})

	// A plugin that leaves plugins.project is uninstalled, with --prune -y,
	// and leaves the record (D3, D29). A project plugin installed by hand, not
	// in the record, stays.
	rc.Run(t, rc.Spec{Name: "setup:ai: a project plugin leaves the list", Values: aiValues,
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			aiState(c, "installed", "hand@tools|project")
			if code := aiRun(c, 1, false); code != 0 {
				return code
			}
			var kept strings.Builder
			for _, l := range strings.SplitAfter(harness.ReadFile(c.ToolsValues), "\n") {
				if l != "" && !strings.Contains(l, `"lsp@tools" = `) {
					kept.WriteString(l)
				}
			}
			harness.WriteFile(c.T, c.ToolsValues, kept.String())
			return aiRun(c, 2, false)
		},
		Message: "Uninstalling lsp@tools (project)",
		Check: func(c *rc.Case) {
			c.CheckThat("setup:ai: a plugin that left the list is uninstalled (D3, D29)",
				aiCalled(c, 2, "plugin uninstall --scope project --prune -y lsp@tools"))
			c.CheckThat("setup:ai: the uninstalled plugin leaves the record (D3)", aiRecord(c, ""))
			c.CheckThat("setup:ai: a plugin installed by hand is never uninstalled (D3)",
				!aiCalled(c, 2, "plugin uninstall --scope project --prune -y hand@tools"))
			c.CheckThat("setup:ai: user scope is never uninstalled (D3)",
				!strings.Contains(harness.ReadFile(aiCalls(c, 2)), "uninstall --scope user"))
		}})

	// A failed install records nothing; a failed uninstall keeps the record;
	// both exit 1 (D13).
	rc.Run(t, rc.Spec{Name: "setup:ai: a failed install", Code: 1, Values: aiValues,
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			return aiRun(c, 1, false, "AI_FAIL=install")
		},
		Message: "the stub install fails",
		Check: func(c *rc.Case) {
			c.CheckThat("setup:ai: a failed install is not recorded (D13)", aiRecord(c, ""))
		}})
	rc.Run(t, rc.Spec{Name: "setup:ai: a failed uninstall", Code: 1, Values: aiRecordValues(`[ "lsp@tools", "old@tools" ]`),
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			aiState(c, "installed", "vwf@virajp-plugins|user", "old@tools|project", "lsp@tools|project")
			return aiRun(c, 1, false, "AI_FAIL=uninstall")
		},
		Message: "the stub uninstall fails",
		Check: func(c *rc.Case) {
			c.CheckThat("setup:ai: a failed uninstall stays recorded (D13)", aiRecord(c, "lsp@tools,old@tools"))
		}})

	// A recorded plugin that is not installed leaves the record with no
	// uninstall call (D18); a listed project plugin that is installed is
	// recorded with no install call (D19).
	rc.Run(t, rc.Spec{Name: "setup:ai: the record follows the installed plugins", Values: aiRecordValues(`[ "gone@tools" ]`),
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			aiState(c, "installed", "vwf@virajp-plugins|user", "lsp@tools|project")
			aiState(c, "marketplaces", "virajp-plugins", "tools")
			return aiRun(c, 1, false)
		},
		Check: func(c *rc.Case) {
			c.CheckThat("setup:ai: a recorded plugin that is not installed, no uninstall call (D18)",
				!aiCalled(c, 1, "plugin uninstall --scope project --prune -y gone@tools"))
			c.CheckThat("setup:ai: an installed listed plugin, no install call (D19)", aiNoChange(c, 1))
			c.CheckThat("setup:ai: the record drops gone@tools and adds lsp@tools (D18, D19)", aiRecord(c, "lsp@tools"))
		}})

	// A recorded, listed plugin that is not installed is installed again and
	// stays recorded (D26).
	rc.Run(t, rc.Spec{Name: "setup:ai: a recorded plugin installed again", Values: aiRecordValues(`[ "lsp@tools" ]`),
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			aiState(c, "installed", "vwf@virajp-plugins|user")
			aiState(c, "marketplaces", "virajp-plugins", "tools")
			return aiRun(c, 1, false)
		},
		Check: func(c *rc.Case) {
			c.CheckThat("setup:ai: a recorded plugin that is not installed is installed again (D26)",
				aiCalled(c, 1, "plugin install --scope project lsp@tools"))
			c.CheckThat("setup:ai: the plugin stays recorded (D26)", aiRecord(c, "lsp@tools"))
		}})

	// A record write that fails is a failed step (D25).
	rc.Run(t, rc.Spec{Name: "setup:ai: a config set that fails", Code: 1, Values: aiValues,
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			return aiRun(c, 1, false, "BOOTSTRAP_BIN="+filepath.Join(c.Dir, "ai", "engine"), "AI_SET_FAIL=1")
		},
		Message: "the stub set fails"})

	// No catalog: the first config get exits 2, so the task stops with code 2
	// before any change: it installs nothing, uninstalls nothing and keeps the
	// record (D34).
	rc.Run(t, rc.Spec{Name: "setup:ai: config get reads no catalog", Code: 2,
		Values: aiRecordValues(`[ "old@tools", "gone@tools" ]`),
		Call: func(c *rc.Case) int {
			if code := aiSetup(c); code != 0 {
				return code
			}
			aiState(c, "installed", "vwf@virajp-plugins|user", "old@tools|project")
			aiState(c, "marketplaces", "virajp-plugins", "tools")
			return aiRun(c, 1, false, "BOOTSTRAP_CLONE", "BOOTSTRAP_CACHE="+filepath.Join(c.Dir, "no-cache"),
				"BOOTSTRAP_URL=file://"+filepath.Join(c.Dir, "no-such-remote"))
		},
		Message: "config get tools.claude.plugins.user needs the catalog",
		Check: func(c *rc.Case) {
			calls := harness.ReadFile(aiCalls(c, 1))
			c.CheckThat("setup:ai, no catalog: no install call (D34)", !strings.Contains(calls, "install"))
			c.CheckThat("setup:ai, no catalog: no uninstall call (D34)", !strings.Contains(calls, "uninstall"))
			c.CheckThat("setup:ai, no catalog: the record is unchanged (D34)", aiRecord(c, "old@tools,gone@tools"))
		}})
}

// setup passes --upgrade to setup:ai (D39): the case runs the synced task
// file of setup, a plain run, then a run with --upgrade. A PATH shim for
// mise logs each `mise run` to runs<n> and does nothing else.
func TestSetupUpgrade(t *testing.T) {
	const setupMise = "#!/usr/bin/env bash\n[[ $1 != run ]] || printf '%s\\n' \"$*\" >> \"$SETUP_RUNS\"\n"
	runs := func(c *rc.Case, n int) string { return filepath.Join(c.Dir, "setup", fmt.Sprintf("runs%d", n)) }
	setupRun := func(c *rc.Case, n int, upgrade bool) int {
		env := harness.SetEnv(c.Env, "SETUP_RUNS="+runs(c, n), fmt.Sprintf("usage_upgrade=%t", upgrade))
		env = prependPath(env, filepath.Join(c.Dir, "setup", "bin"))
		return run{dir: c.Consumer, env: env}.do(c, c.Path(tasksPath+"/setup/_default.py"))
	}
	rc.Run(t, rc.Spec{Name: "setup: --upgrade passes --upgrade to setup:ai", Values: aiValues,
		Call: synced(func(c *rc.Case) int {
			writeScript(c, filepath.Join(c.Dir, "setup", "bin", "mise"), setupMise)
			if code := setupRun(c, 1, false); code != 0 {
				return code
			}
			return setupRun(c, 2, true)
		}),
		Message: "Setup complete!",
		Check: func(c *rc.Case) {
			c.CheckThat("setup: the synced setup:ai is there", isExec(c.Path(tasksPath+"/setup/ai.py")))
			c.CheckThat("setup: a plain run does not run setup:ai (D4)",
				!strings.Contains(harness.ReadFile(runs(c, 1)), "run setup:ai"))
			c.CheckThat("setup: --upgrade runs setup:ai --upgrade (D39)", fileHasLine(runs(c, 2), "run setup:ai --upgrade"))
		}})
}

// setup:configs --all with the real engine (D54, D62, D63): the sibling api is
// a repository next to the consumer with its own .config/bootstrap/, so its
// sync writes [home]. The sibling web is not there: the sync of the home skips
// it with an info line, and the task with a warning. The task and helpers.py
// come from a sync of the consumer.
func TestConfigsAll(t *testing.T) {
	const siblingValues = devValues + `
[members.api]
folder = "../api"
type = "sibling"
url = "https://example.com/api"
[members.web]
folder = "../web"
type = "sibling"
url = "https://example.com/web"`
	configsAll := synced(func(c *rc.Case) int {
		env := harness.SetEnv(c.Env, "BOOTSTRAP_VALUES", "BOOTSTRAP_BIN="+harness.Bin, "usage_all=true")
		return run{dir: c.Consumer, env: env}.do(c, c.Path(tasksPath+"/setup/configs/_default.py"))
	})
	siblingRepo := func(c *rc.Case) string {
		api := filepath.Join(c.Dir, "api")
		mkdir(c, api)
		git(c, api, "init", "--quiet", "-b", "main")
		return api
	}
	commitSibling := func(c *rc.Case, api string) {
		git(c, api, "add", "-A")
		git(c, api, "commit", "--quiet", "--no-verify", "-m", "values")
	}
	rc.Run(t, rc.Spec{Name: "setup:configs --all writes [home] in a sibling that is there", Values: siblingValues,
		Change: func(c *rc.Case) {
			api := siblingRepo(c)
			vals := filepath.Join(api, ".config", "bootstrap")
			rc.WriteValues(c.T, vals, devValues)
			repo := filepath.Join(vals, "repo.toml")
			text := regexp.MustCompile(`(?m)^repo_name = .*$`).ReplaceAllString(harness.ReadFile(repo), `repo_name = "api"`)
			harness.WriteFile(c.T, repo, text)
			commitSibling(c, api)
		},
		Call: configsAll,
		Check: func(c *rc.Case) {
			home := filepath.Join(c.Dir, "api", ".config", "bootstrap", "repo.toml")
			c.CheckThat("--all: the sibling gets [home] type sibling", c.TomlQ(`.home.type == "sibling"`, home))
			c.CheckThat("--all: [home] has the url of the home", c.TomlQ(`.home.url == "https://example.com/test"`, home))
			c.CheckThat("--all: [home] has the folder of the home", c.TomlQ(`.home | has("folder")`, home))
			c.CheckThat("--all: the header of the sibling", c.LogHas("Syncing api ..."))
			c.CheckThat("--all: the missing sibling, the info line of the sync", c.LogHas("member web skipped: ../web is not there"))
			c.CheckThat("--all: the missing sibling, the warning of the task", c.LogHas("member web: ../web is not on this machine"))
		}})
	rc.Run(t, rc.Spec{Name: "setup:configs --all skips a sibling of the old form", Values: siblingValues,
		Change: func(c *rc.Case) {
			api := siblingRepo(c)
			harness.WriteFile(c.T, filepath.Join(api, ".config", "bootstrap.toml"), harness.Lines("format = 2", `repo_name = "api"`))
			commitSibling(c, api)
		},
		Call:    configsAll,
		Message: "member api: run bootstrap migrate there",
		Check: func(c *rc.Case) {
			c.CheckThat("--all, an old-form sibling: no values folder written", rc.NoFile(filepath.Join(c.Dir, "api", ".config", "bootstrap")))
		}})
}

// The merge task git:merge:main honours the landing of [branches.main]: pr
// opens no direct merge (with no origin, it stops before it touches a
// branch), and direct-merge merges feature into main locally. The case
// removes the hook config, so the hook safety net does not run.
func TestMergeLanding(t *testing.T) {
	mergeFeature := synced(func(c *rc.Case) int {
		_ = os.Remove(c.Path(".config/pre-commit-config.yaml"))
		c.CommitConsumer()
		git(c, c.Consumer, "checkout", "--quiet", "-b", "feature")
		git(c, c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "feature")
		// mise run of the test sets MISE_ORIGINAL_CWD to this repository, and
		// the merge task compares it with its own repository.
		env := harness.SetEnv(c.Env, "MISE_ORIGINAL_CWD", "MISE_TASK_NAME", "BOOTSTRAP_BIN="+harness.Bin, "usage_source=feature")
		return run{dir: c.Consumer, env: env}.do(c, c.Path(tasksPath+"/git/merge/main.py"))
	})
	mergesOnMain := func(c *rc.Case) string { return plainGit(c.Consumer, "rev-list", "--merges", "--count", "main") }
	rc.Run(t, rc.Spec{Name: "git:merge:main, landing pr opens no direct merge", Code: 1,
		Values: devValues + "\n[branches.main]\nlanding = \"pr\"", Call: mergeFeature,
		Message: `landing "pr" of [branches.main] needs an origin remote`,
		Check: func(c *rc.Case) {
			c.CheckThat("landing pr: no merge on main", mergesOnMain(c) == "0")
			c.CheckThat("landing pr: the branch stays", plainGit(c.Consumer, "rev-parse", "--abbrev-ref", "HEAD") == "feature")
		}})
	rc.Run(t, rc.Spec{Name: "git:merge:main, landing direct-merge merges",
		Values: devValues + "\n[branches.main]\nlanding = \"direct-merge\"", Call: mergeFeature,
		Message: "Landing feature on main (direct-merge) ...",
		Check: func(c *rc.Case) {
			c.CheckThat("landing direct-merge: one merge on main", mergesOnMain(c) == "1")
			c.CheckThat("landing direct-merge: no push without origin", c.LogHas("No origin remote"))
		}})
}

// git:merge:main lands develop on main with a fast-forward. When main, local
// or on origin, has a commit that develop does not have, it stops before it
// changes a branch.
func TestMergeFastForward(t *testing.T) {
	// mergeDevelop: develop is main with one commit more, then <setup>, then
	// the task runs from develop. The case removes the hook config, so the hook
	// safety net does not run. The file main-before of the case holds main
	// before the run.
	mergeDevelop := func(setup func(c *rc.Case)) func(c *rc.Case) int {
		return synced(func(c *rc.Case) int {
			_ = os.Remove(c.Path(".config/pre-commit-config.yaml"))
			c.CommitConsumer()
			git(c, c.Consumer, "checkout", "--quiet", "develop")
			git(c, c.Consumer, "merge", "--quiet", "--ff-only", "main")
			git(c, c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "work")
			if setup != nil {
				setup(c)
			}
			harness.WriteFile(c.T, filepath.Join(c.Dir, "main-before"), plainGit(c.Consumer, "rev-parse", "main"))
			env := harness.SetEnv(c.Env, "MISE_ORIGINAL_CWD", "MISE_TASK_NAME", "BOOTSTRAP_BIN="+harness.Bin, "usage_source=develop")
			return run{dir: c.Consumer, env: env}.do(c, c.Path(tasksPath+"/git/merge/main.py"))
		})
	}
	// withOrigin: a bare origin with main and develop, both pushed.
	withOrigin := func(c *rc.Case) {
		origin := filepath.Join(c.Dir, "origin.git")
		git(c, c.Dir, "init", "--quiet", "--bare", "-b", "main", origin)
		git(c, c.Consumer, "remote", "add", "origin", origin)
		git(c, c.Consumer, "push", "--quiet", "--no-verify", "-u", "origin", "main", "develop")
	}
	rev := func(c *rc.Case, dir, ref string) string { return plainGit(dir, "rev-parse", ref) }
	mainUnchanged := func(c *rc.Case) bool {
		return harness.ReadFile(filepath.Join(c.Dir, "main-before")) == rev(c, c.Consumer, "main")
	}

	rc.Run(t, rc.Spec{Name: "git:merge:main, develop fast-forwards main", Values: bothCore,
		Change: precommitBothBranches, Call: mergeDevelop(nil),
		Message: "Fast-forwarding main to develop ... [OK]",
		Check: func(c *rc.Case) {
			c.CheckThat("fast-forward: the check passes", c.LogHas("Checking if main can fast-forward to develop ... [OK]"))
			c.CheckThat("fast-forward: main is develop", rev(c, c.Consumer, "main") == rev(c, c.Consumer, "develop"))
			c.CheckThat("fast-forward: no merge commit on main", plainGit(c.Consumer, "rev-list", "--merges", "--count", "main") == "0")
			c.CheckThat("fast-forward: the run goes back to develop", plainGit(c.Consumer, "rev-parse", "--abbrev-ref", "HEAD") == "develop")
		}})
	rc.Run(t, rc.Spec{Name: "git:merge:main, develop fast-forwards main of origin", Values: bothCore,
		Change: precommitBothBranches, Call: mergeDevelop(withOrigin),
		Message: "Pushing main branch to origin ... [OK]",
		Check: func(c *rc.Case) {
			origin := filepath.Join(c.Dir, "origin.git")
			c.CheckThat("fast-forward, origin: main of origin is develop", rev(c, origin, "main") == rev(c, c.Consumer, "develop"))
			c.CheckThat("fast-forward, origin: no merge commit on main of origin",
				plainGit(origin, "rev-list", "--merges", "--count", "main") == "0")
		}})
	rc.Run(t, rc.Spec{Name: "git:merge:main, main has a commit that develop does not have", Values: bothCore, Code: 1,
		Change: precommitBothBranches,
		Call: mergeDevelop(func(c *rc.Case) {
			git(c, c.Consumer, "checkout", "--quiet", "main")
			git(c, c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "hotfix")
			git(c, c.Consumer, "checkout", "--quiet", "develop")
		}),
		Message: "main has 1 commit(s) that develop does not have, so main cannot fast-forward to develop.",
		Check: func(c *rc.Case) {
			c.CheckThat("no fast-forward: the commit is listed", c.LogHas("hotfix"))
			c.CheckThat("no fast-forward: the fix", c.LogHas("git switch develop && git merge main"))
			c.CheckThat("no fast-forward: main does not change", mainUnchanged(c))
			c.CheckThat("no fast-forward: the branch stays", plainGit(c.Consumer, "rev-parse", "--abbrev-ref", "HEAD") == "develop")
			checkLogLacks(c, "no fast-forward: no hook pass", "hook safety net")
		}})
	rc.Run(t, rc.Spec{Name: "git:merge:main, main of origin has a commit that develop does not have", Values: bothCore, Code: 1,
		Change: precommitBothBranches,
		Call: mergeDevelop(func(c *rc.Case) {
			withOrigin(c)
			// origin gets a commit on main that the local main does not have.
			git(c, c.Consumer, "checkout", "--quiet", "main")
			git(c, c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "remote fix")
			git(c, c.Consumer, "push", "--quiet", "--no-verify", "origin", "main")
			git(c, c.Consumer, "checkout", "--quiet", "develop")
			git(c, c.Consumer, "branch", "--force", "main", "main~1")
		}),
		Message: "origin/main has 1 commit(s) that develop does not have, so main cannot fast-forward to develop.",
		Check: func(c *rc.Case) {
			c.CheckThat("no fast-forward, origin: the fix", c.LogHas("git switch develop && git merge origin/main"))
			c.CheckThat("no fast-forward, origin: main does not change", mainUnchanged(c))
			c.CheckThat("no fast-forward, origin: the branch stays", plainGit(c.Consumer, "rev-parse", "--abbrev-ref", "HEAD") == "develop")
		}})
}

var (
	oldGitTasks = []string{"code/merge/_default", "code/merge/main", "code/merge/develop", "setup/worktree", "code/worktrees"}
	newGitTasks = []string{"git/merge/main.py", "git/merge/develop.py", "git/worktree/create.py", "git/worktree/setup.py",
		"git/worktree/delete.py", "git/worktree/list.py"}
)

func precommitBothBranches(c *rc.Case) { git(c, c.Consumer, "branch", "develop") }

// The git:* tasks (D1, D2, D3). A real run of a repo with main and develop
// writes git:merge:<branch> for each core branch and the four
// git:worktree:* tasks. The old names of a run before are rows of its
// manifest, and the run removes them (D1).
func TestGitTasksSynced(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "the git tasks: a real run writes the new names and removes the old", Values: bothCore,
		Args: []string{},
		Change: func(c *rc.Case) {
			precommitBothBranches(c)
			rows := []string{"# Written by bootstrap. Do not edit this file."}
			for _, p := range oldGitTasks {
				rel := tasksPath + "/" + p
				writeConsumer(c, rel, "# Synced from bootstrap: tools/oldtool/"+rel+". Do not edit this file.",
					"# The next sync writes it again, and each change in this file is lost.")
				rows = append(rows, "git\t"+rel)
			}
			writeConsumer(c, ".config/bootstrap/repo-generated.tsv", rows...)
		},
		Check: func(c *rc.Case) {
			manifest := harness.ReadFile(c.Path(".config/bootstrap/repo-generated.tsv"))
			for _, p := range newGitTasks {
				c.CheckThat("the git tasks: "+p+" is written", isFile(c.Path(tasksPath+"/"+p)))
			}
			for _, p := range oldGitTasks {
				c.CheckThat("the git tasks: the old "+p+" is removed", c.LogHas("removed    "+c.Path(tasksPath+"/"+p)))
				c.CheckThat("the git tasks: no "+p+" stays", rc.NoFile(c.Path(tasksPath+"/"+p)))
				c.CheckThat("the git tasks: the manifest has no "+p, !strings.Contains(manifest, tasksPath+"/"+p))
			}
			c.CheckThat("the git tasks: the folder code/merge is gone", rc.NoFile(c.Path(tasksPath+"/code/merge")))
		}})
}

// gitTask runs a git:* task with mise run in the consumer of the case, as a
// user runs it there, with the engine of this test (D33). The log gets its
// output, and the function gives its exit code (git_task_code).
func gitTask(c *rc.Case, args ...string) int {
	env := harness.SetEnv(harness.Env(), "MISE_ENV", "MISE_ORIGINAL_CWD", "MISE_TASK_NAME", "MISE_TASK_DIR",
		"MISE_TASK_FILE", "MISE_CONFIG_ROOT", "FORCE_COLOR", "BOOTSTRAP_BIN="+harness.Bin, "MISE_PROJECT_ROOT="+c.Consumer)
	return run{dir: c.Consumer, env: env, trunc: true}.do(c, append([]string{"mise", "run"}, args...)...)
}

// The merge entry tasks with a stub library (D23): the stub prints the
// destination and the source that the task gives it, so no case runs the real
// merge.
func TestMergeEntryTasks(t *testing.T) {
	mergeStubbed := synced(func(c *rc.Case) int {
		harness.WriteFile(c.T, c.Path(tasksPath+"/_scripts/merge.py"), `def merge_to_destination_branch(destination_branch, source_branch=None):
  print(f"stub-merge: [{destination_branch}] [{source_branch or ''}]", flush=True)
`)
		c.CommitConsumer()
		return 0
	})
	rc.Run(t, rc.Spec{Name: "git:merge, main only: the source it names", Values: devValues, Call: mergeStubbed,
		Check: func(c *rc.Case) {
			c.CheckThat("git:merge, main only: no git:merge:develop", rc.NoFile(c.Path(tasksPath+"/git/merge/develop.py")))
			got := gitTask(c, "git:merge:main", "feat")
			c.CheckThat("git:merge, main only: git:merge:main feat exits 0", got == 0)
			c.CheckThat("git:merge, main only: git:merge:main feat gives main feat", c.LogHas("stub-merge: [main] [feat]"))
		}})
	// git:merge:develop gives an empty source, which the library reads as the
	// current branch.
	rc.Run(t, rc.Spec{Name: "git:merge, main and develop: the default sources", Values: bothCore,
		Change: precommitBothBranches, Call: mergeStubbed,
		Check: func(c *rc.Case) {
			got := gitTask(c, "git:merge:main")
			c.CheckThat("git:merge, both branches: git:merge:main exits 0", got == 0)
			c.CheckThat("git:merge, both branches: git:merge:main gives main develop", c.LogHas("stub-merge: [main] [develop]"))
			git(c, c.Consumer, "checkout", "--quiet", "-b", "feature")
			got = gitTask(c, "git:merge:develop")
			c.CheckThat("git:merge, both branches: git:merge:develop exits 0", got == 0)
			c.CheckThat("git:merge, both branches: git:merge:develop gives develop and the current branch",
				c.LogHas("stub-merge: [develop] []"))
		}})
}

// git:worktree:create and git:worktree:delete in a repo with main and develop
// (D3, D13, D15, D16, D22, D25, D28). Each create passes --no-setup.
func TestWorktreeTasks(t *testing.T) {
	syncedCall := synced(func(*rc.Case) int { return 0 })
	hasBranch := func(c *rc.Case, b string) bool {
		_, err := harness.GitErr(c.Consumer, "show-ref", "--verify", "--quiet", "refs/heads/"+b)
		return err == nil
	}
	rc.Run(t, rc.Spec{Name: "git:worktree: create, merge into develop, delete", Values: bothCore,
		Change: precommitBothBranches, Call: syncedCall,
		Check: func(c *rc.Case) {
			wt := c.Path(".worktrees/x")
			got := gitTask(c, "git:worktree:create", "x", "--no-setup")
			c.CheckThat("worktree, merged: create exits 0", got == 0)
			c.CheckThat("worktree, merged: create makes .worktrees/x", isDir(wt))
			c.CheckThat("worktree, merged: .worktrees/x is on the new branch x", plainGit(wt, "rev-parse", "--abbrev-ref", "HEAD") == "x")
			git(c, wt, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "x")
			git(c, c.Consumer, "checkout", "--quiet", "develop")
			git(c, c.Consumer, "merge", "--quiet", "--no-ff", "--no-verify", "-m", "merge x", "x")
			git(c, c.Consumer, "checkout", "--quiet", "main")
			got = gitTask(c, "git:worktree:delete", "x")
			c.CheckThat("worktree, merged: delete exits 0", got == 0)
			c.CheckThat("worktree, merged: delete removes the worktree", rc.NoFile(wt))
			c.CheckThat("worktree, merged: delete removes the branch", !hasBranch(c, "x"))
		}})
	rc.Run(t, rc.Spec{Name: "git:worktree: delete of a branch that no core branch has", Values: bothCore,
		Change: precommitBothBranches, Call: syncedCall,
		Check: func(c *rc.Case) {
			gitTask(c, "git:worktree:create", "y", "--no-setup")
			git(c, c.Path(".worktrees/y"), "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "y")
			got := gitTask(c, "git:worktree:delete", "y")
			c.CheckThat("worktree, not merged: delete exits 1", got == 1)
			c.CheckThat("worktree, not merged: delete prints the -D command", c.LogHas("git branch -D y"))
			c.CheckThat("worktree, not merged: the branch stays", hasBranch(c, "y"))
		}})
	rc.Run(t, rc.Spec{Name: "git:worktree: another path, and the refusals", Values: bothCore,
		Change: precommitBothBranches, Call: syncedCall,
		Check: func(c *rc.Case) {
			elsewhere := filepath.Join(c.Dir, "elsewhere")
			git(c, c.Consumer, "worktree", "add", "--quiet", "-b", "z", elsewhere)
			got := gitTask(c, "git:worktree:delete", "z")
			c.CheckThat("worktree, another path: delete exits 0", got == 0)
			c.CheckThat("worktree, another path: delete removes it", rc.NoFile(elsewhere))
			mkdir(c, c.Path(".worktrees/w"))
			got = gitTask(c, "git:worktree:create", "w", "--no-setup")
			c.CheckThat("worktree: create onto an existing path exits 1", got == 1)
			got = gitTask(c, "git:worktree:create", "develop", "--no-setup")
			c.CheckThat("worktree: create develop exits 2", got == 2)
			got = gitTask(c, "git:worktree:delete", "develop")
			c.CheckThat("worktree: delete develop exits 2", got == 2)
			git(c, c.Consumer, "branch", "nowt")
			got = gitTask(c, "git:worktree:delete", "nowt")
			c.CheckThat("worktree: delete of a branch with no worktree exits 1", got == 1)
		}})
}

// D20: a .config/wt.toml or a member's worktree_install that runs the old
// setup:worktree stops check with code 2; init renames it to
// git:worktree:setup.
func TestWorktreeSetupRename(t *testing.T) {
	wtOldCmd := func(c *rc.Case) {
		writeConsumer(c, ".config/wt.toml", "[[pre-start]]", `setup = "mise run setup:worktree"`)
	}
	rc.Check(t, rc.Spec{Name: "--check, .config/wt.toml runs setup:worktree", Code: 2, Values: devValues, Change: wtOldCmd,
		Message: ".config/wt.toml runs mise run setup:worktree, the old name of git:worktree:setup - run bootstrap migrate"})
	rc.Run(t, rc.Spec{Name: "init renames setup:worktree in .config/wt.toml", Values: devValues, Change: wtOldCmd,
		Call: initYes, Message: "rename setup:worktree",
		Check: func(c *rc.Case) {
			wt := harness.ReadFile(c.Path(".config/wt.toml"))
			c.CheckThat("init: wt.toml runs git:worktree:setup", strings.Contains(wt, `setup = "mise run git:worktree:setup"`))
			c.CheckThat("init: wt.toml has no setup:worktree", !strings.Contains(wt, "setup:worktree"))
		}})
	const memberOldCmd = devValues + `
[members.web]
folder = "projects/web"
type = "folder"
worktree_install = "mise run setup:worktree"`
	memberFolder := func(c *rc.Case) { mkdir(c, c.Path("projects/web")) }
	// The message shows the absolute path of repo.toml, so the case names its
	// two ends.
	rc.Check(t, rc.Spec{Name: "--check, a member's worktree_install runs setup:worktree", Code: 2, Values: memberOldCmd,
		Change:  memberFolder,
		Message: "/.config/bootstrap/repo.toml runs mise run setup:worktree, the old name of git:worktree:setup - run bootstrap migrate",
		Check: func(c *rc.Case) {
			c.CheckThat("check: the message names members.web.worktree_install", c.LogHas("members.web.worktree_install in "))
		}})
	rc.Run(t, rc.Spec{Name: "init renames setup:worktree in a member's worktree_install", Values: memberOldCmd,
		Change: memberFolder, Call: initYes, Message: "rename setup:worktree",
		Check: func(c *rc.Case) {
			repo := harness.ReadFile(c.RepoValues)
			c.CheckThat("init: worktree_install runs git:worktree:setup",
				strings.Contains(repo, `worktree_install = "mise run git:worktree:setup"`))
			c.CheckThat("init: repo.toml has no setup:worktree", !strings.Contains(repo, "setup:worktree"))
		}})
}

// The global file lists of the gates (D8, D9): the sync renders code_files of
// the mise tool, its defaults and the value of the repository, to
// .config/bootstrap/config-files.toml, and the same excludes to dprint.jsonc.
func TestCodeFiles(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "code_files: config-files.toml and dprint.jsonc",
		Values: devValues + "\n[tools.mise]\ncode_files.excludes = [ \"generated\" ]",
		Call:   engine("sync"), Message: "Synced: ",
		Check: func(c *rc.Case) {
			file := c.Path(".config/bootstrap/config-files.toml")
			dprint := c.Path("dprint.jsonc")
			c.CheckThat("code_files: the mise defaults",
				c.TomlQ(`.excludes | contains(["**/*-lock.yaml", "**/*.svg", "**/package-lock.json"])`, file))
			c.CheckThat("code_files: the value of the repository", c.TomlQ(`.excludes | contains(["generated"])`, file))
			c.CheckThat("code_files: no includes", c.TomlQ(`(.includes | length) == 0`, file))
			c.CheckThat("code_files: dprint.jsonc, a default", rc.FileHas(dprint, `"**/*-lock.yaml"`))
			c.CheckThat("code_files: dprint.jsonc, the value of the repository", rc.FileHas(dprint, `"**/generated"`))
			c.CheckThat("code_files: dprint.jsonc, below the folder of the value", rc.FileHas(dprint, `"**/generated/**"`))
		}})
}
