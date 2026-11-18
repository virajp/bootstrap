package commands

import (
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// The JSON contract (D1, D12, D26): each --json command prints one compact
// envelope line on stdout, and its exit code is the envelope code. The human
// lines go to stderr, so a case that reads the envelope alone drops them.
func TestJSONEnvelopes(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "contract --json outside a repository", Values: devValues,
		Call: func(c *rc.Case) int {
			dir := filepath.Join(c.Dir, "outside")
			mkdir(c, dir)
			env := harness.SetEnv(c.Env, "BOOTSTRAP_VALUES", "MISE_PROJECT_ROOT", "BOOTSTRAP_CLONE")
			return run{dir: dir, env: env, quiet: true}.do(c, harness.Bin, "contract", "--json")
		},
		Message: `{"contract":1,"command":"contract","ok":true,"code":0,"data":{"versions":[1],`,
		Check: func(c *rc.Case) {
			c.CheckThat("contract --json: one line", jsonLines(c))
			c.CheckThat("contract --json: the 14 ports", c.LogHas(`"ports":["code:format","code:lint","code:sec","code:precommit","git:merge:<branch>","git:worktree:create","git:worktree:setup","git:worktree:delete","code:lsp:check","setup","test","test:e2e","test:golden","p:<id>:deploy"]`))
		}})
	rc.Run(t, rc.Spec{Name: "config get --json", Values: devValues,
		Call: func(c *rc.Case) int { return engineOut(c, nil, "config", "get", "repo.repo_name", "--json") },
		Check: func(c *rc.Case) {
			c.CheckThat("config get --json: the envelope", logIs(c, `{"contract":1,"command":"config get","ok":true,"code":0,"data":{"key":"repo.repo_name","value":"test","source":"file"},"errors":[],"warnings":[]}`))
		}})
	rc.Run(t, rc.Spec{Name: "config get --json --format toml", Code: 2, Values: devValues,
		Call: func(c *rc.Case) int {
			return engineOut(c, nil, "config", "get", "repo.repo_name", "--json", "--format", "toml")
		},
		Message: `{"contract":1,"command":"config get","ok":false,"code":2,"data":null,"errors":[{"code":2,"message":"config get takes --json or --format, not both"}]`})
	rc.Run(t, rc.Spec{Name: "check --json: an unknown key in repo.toml", Code: 2, Values: devValues + "\nbogus = 1",
		Call:    func(c *rc.Case) int { return engineOut(c, nil, "check", "--json") },
		Message: `{"name":"values","status":"fail","findings":[{"code":2,"severity":"error","message":"bogus in `,
		Check: func(c *rc.Case) {
			c.CheckThat("check --json: the path of the finding", c.LogHas(`/repo.toml is not a key of repo.toml","path":".config/bootstrap/repo.toml"}]}`))
			c.CheckThat("check --json: a skipped stage", c.LogHas(`{"name":"old-forms","status":"skipped","findings":[]}`))
			c.CheckThat("check --json: one line", jsonLines(c))
		}})
	noGitChange := func(c *rc.Case) {
		out, err := harness.GitErr(c.Consumer, "status", "--porcelain")
		c.CheckThat("add --json: no change in git", err == nil && out == "")
	}
	rc.Run(t, rc.Spec{Name: "add tool --json without --yes or --dry-run", Code: 2, Values: devValues,
		Call:    engine("add", "tool", "--json", "go"),
		Message: `{"contract":1,"command":"add","ok":false,"code":2,"data":null,"errors":[{"code":2,"message":"add --json needs --yes or --dry-run`,
		Check:   noGitChange})
	rc.Run(t, rc.Spec{Name: "add tool --json --dry-run", Values: devValues,
		Call:    func(c *rc.Case) int { return engineOut(c, nil, "add", "tool", "--json", "--dry-run", "go") },
		Message: `{"contract":1,"command":"add","ok":true,"code":0,`,
		Check: func(c *rc.Case) {
			c.CheckThat("add --json --dry-run: a plan", c.LogHas(`"data":{"plan":{"changes":[`))
			c.CheckThat("add --json --dry-run: no outcomes", c.LogHas(`"outcomes":[]}`))
			noGitChange(c)
		}})
	rc.Run(t, rc.Spec{Name: "catalog --json --clone", Values: devValues,
		Call: func(c *rc.Case) int {
			return engineOut(c, harness.SetEnv(c.Env, "BOOTSTRAP_CLONE"), "catalog", "--json", "--clone", c.Clone)
		},
		Message: `{"contract":1,"command":"catalog","ok":true,"code":0,"data":{"fetched_at":"",`})
	rc.Run(t, rc.Spec{Name: "describe --json", Values: devValues,
		Call:    func(c *rc.Case) int { return engineOut(c, nil, "describe", "--json") },
		Message: `{"contract":1,"command":"describe","ok":true,"code":0,"data":{"repo":{`,
		Check: func(c *rc.Case) {
			c.CheckThat("describe --json: tools", c.LogHas(`"tools":[`))
			c.CheckThat("describe --json: members", c.LogHas(`"members":[`))
			c.CheckThat("describe --json: ports", c.LogHas(`"ports":[{"name":"code:format","present":`))
		}})
	rc.Run(t, rc.Spec{Name: "an unknown flag with --json", Code: 2, Values: devValues,
		Call:    func(c *rc.Case) int { return engineOut(c, nil, "sync", "--bogus", "--json") },
		Message: `{"contract":1,"command":"sync","ok":false,"code":2,"data":null,"errors":[{"code":2,"message":"sync has no flag --bogus"}],"warnings":[]}`})
}

const (
	lspValues  = devValues + "\n[tools.go]\n[tools.node]"
	nodeValues = devValues + "\n[tools.node]"
)

// lspJSON is the Call of `bootstrap lsp --json`, with its stdout alone.
func lspJSON(c *rc.Case) int { return engineOut(c, nil, "lsp", "--json") }

// typescript-lsp follows its pin, which needs a package.json or a
// tsconfig.json (D35, D37, D38): node alone has no server and no pin.
func lspPackageJSON(c *rc.Case) { writeConsumer(c, "package.json", "{}") }
func lspTsconfig(c *rc.Case)    { writeConsumer(c, "tsconfig.json", "{}") }

func checkLspTypescript(c *rc.Case) {
	c.CheckThat("lsp --json: the checks of typescript-lsp", lspJq(c,
		`.data.checks."typescript-lsp" == {"command":"typescript-language-server","source":"npm:typescript-language-server"}`))
}

// The language servers (D4, D6, D7, D12, D28, D30): bootstrap lsp --json lists
// the [lsp] servers of the selected tools as entries of the lspServers of
// Claude, with the command and the source of each in checks, and describe
// --json has the same servers in its lsp field.
func TestLspJSON(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "lsp --json: the servers of go and node", Values: lspValues, Change: lspPackageJSON,
		Call: func(c *rc.Case) int {
			if code := lspJSON(c); code != 0 {
				return code
			}
			return run{out: filepath.Join(c.Dir, "describe.json"), trunc: true, quiet: true}.do(c, harness.Bin, "describe", "--json")
		},
		Message: `{"contract":1,"command":"lsp","ok":true,"code":0,"data":{"servers":{`,
		Check: func(c *rc.Case) {
			c.CheckThat("lsp --json: one line", jsonLines(c))
			c.CheckThat("lsp --json: the gopls entry", lspJq(c,
				`.data.servers.gopls == {"command":"mise","args":["x","--","gopls"],"extensionToLanguage":{".go":"go"}}`))
			c.CheckThat("lsp --json: the typescript-lsp entry", lspJq(c,
				`.data.servers."typescript-lsp" | .command == "mise" and .args == ["x","--","typescript-language-server","--stdio"] and .startupTimeout == 60000 and .extensionToLanguage.".ts" == "typescript"`))
			c.CheckThat("lsp --json: the checks of gopls", lspJq(c,
				`.data.checks.gopls == {"command":"gopls","source":"go:golang.org/x/tools/gopls"}`))
			checkLspTypescript(c)
			servers := jqOut(c, ".data.servers", c.Log)
			c.CheckThat("describe --json: lsp.servers is the servers of lsp --json",
				servers == jqOut(c, ".data.lsp.servers", filepath.Join(c.Dir, "describe.json")))
		}})
	rc.Run(t, rc.Spec{Name: "lsp --json: node alone has no server", Values: nodeValues, Call: lspJSON,
		Message: `"command":"lsp","ok":true`,
		Check: func(c *rc.Case) {
			c.CheckThat("lsp --json: no typescript-lsp", lspJq(c,
				`(.data.servers | has("typescript-lsp") | not) and (.data.checks | has("typescript-lsp") | not)`))
		}})
	rc.Run(t, rc.Spec{Name: "lsp --json: a tsconfig.json alone lists typescript-lsp", Values: nodeValues,
		Change: lspTsconfig, Call: lspJSON, Message: `"data":{"servers":{"typescript-lsp":`, Check: checkLspTypescript})
}

// confFile: a generated config of the consumer (conf_file), so the dry run
// shows its diff.
func confFile(c *rc.Case, tool, env string) {
	writeConsumer(c, ".config/mise/conf.d/"+tool+"/mise."+env+".toml", "# x")
}

// The generated dev config of node already exists in the consumer, so the
// dry run shows its diff. With a package.json or a tsconfig.json, the dev
// config of node has both pins, without the requires_found key of their
// [mise_env.dev.tools] entries (D40).
func TestNodePins(t *testing.T) {
	nodeConf := func(c *rc.Case) { confFile(c, "node", "dev") }
	checkNodePins := func(name string) func(c *rc.Case) {
		return func(c *rc.Case) {
			c.CheckThat(name+": the typescript pin", c.LogHas(`"npm:typescript"`))
			checkLogLacks(c, name, "requires_found")
		}
	}
	rc.Run(t, rc.Spec{Name: "node alone: no typescript pins in dev", Values: nodeValues, Change: nodeConf,
		Message: "[tools.node]",
		Check:   func(c *rc.Case) { checkLogLacks(c, "node alone: no typescript pins", "npm:typescript") }})
	rc.Run(t, rc.Spec{Name: "node and a package.json: the typescript pins in dev", Values: nodeValues,
		Change:  func(c *rc.Case) { nodeConf(c); lspPackageJSON(c) },
		Message: `"npm:typescript-language-server"`, Check: checkNodePins("node and a package.json")})
	rc.Run(t, rc.Spec{Name: "node and a tsconfig.json: the typescript pins in dev", Values: nodeValues,
		Change:  func(c *rc.Case) { nodeConf(c); lspTsconfig(c) },
		Message: `"npm:typescript-language-server"`, Check: checkNodePins("node and a tsconfig.json")})
}

// check refuses a wrong [lsp] block (D7, D30). Each change adds a table to
// the go tool of the clone of the case.
func TestLspRefusals(t *testing.T) {
	lspAdd := func(lines ...string) func(c *rc.Case) {
		return func(c *rc.Case) { appendClone(c, "repo/tools/go/tools.toml", lines...) }
	}
	rc.Check(t, rc.Spec{Name: "[lsp.x] with no extensions", Code: 2, Values: lspValues,
		Change:  lspAdd("[lsp.x]", `command = "x"`, `source = "host"`),
		Message: `repo/tools/go/tools.toml: [lsp.x] needs extensions, a table of ".<ext>" = "<language id>" that is not empty`})
	rc.Check(t, rc.Spec{Name: "[lsp.x] with empty extensions", Code: 2, Values: lspValues,
		Change:  lspAdd("[lsp.x]", `command = "x"`, "extensions = {}", `source = "host"`),
		Message: `repo/tools/go/tools.toml: [lsp.x] needs extensions, a table of ".<ext>" = "<language id>" that is not empty`})
	rc.Check(t, rc.Spec{Name: "[lsp.x] with no command", Code: 2, Values: lspValues,
		Change:  lspAdd("[lsp.x]", `extensions = { ".x" = "x" }`, `source = "host"`),
		Message: "repo/tools/go/tools.toml: [lsp.x] needs command, the binary of the server (a string)"})
	rc.Check(t, rc.Spec{Name: "one server name in two tools", Code: 2, Values: lspValues,
		Change:  lspAdd("[lsp.typescript-lsp]", `command = "x"`, `extensions = { ".x" = "x" }`, `source = "host"`),
		Message: "language server 'typescript-lsp' is defined in repo/tools/go and in repo/tools/node - keep one"})
	rc.Check(t, rc.Spec{Name: "[lsp.x] with a source that the tool does not pin", Code: 2, Values: lspValues,
		Change:  lspAdd("[lsp.x]", `command = "x"`, `extensions = { ".x" = "x" }`, `source = "nope"`),
		Message: "repo/tools/go/tools.toml: source 'nope' of [lsp.x] is not a mise tool of this tool - name its pin"})
	rc.Check(t, rc.Spec{Name: "[lsp.x] with a command with spaces", Code: 2, Values: lspValues,
		Change:  lspAdd("[lsp.x]", `command = "pyrefly lsp"`, `extensions = { ".x" = "x" }`, `source = "host"`),
		Message: "repo/tools/go/tools.toml: [lsp.x] command 'pyrefly lsp' has whitespace - command is one binary; put the rest in args"})
}

// A server follows its source pin and the envs of its tool (D12, D28):
// pyrefly needs the pyproject.toml of its pin, and a tool without dev has no
// server. lsp does not run the rules of check, so it reads a values file in
// which go has env ci only.
func TestLspPins(t *testing.T) {
	pyprojectRoot := func(c *rc.Case) {
		confFile(c, "python", "dev")
		harness.WriteFile(c.T, c.Path("pyproject.toml"), "")
	}
	rc.Run(t, rc.Spec{Name: "lsp --json: python without a pyproject.toml", Values: pythonValues, Call: lspJSON,
		Message: `"command":"lsp","ok":true`,
		Check: func(c *rc.Case) {
			c.CheckThat("lsp --json: no pyrefly", lspJq(c, `.data.servers | has("pyrefly") | not`))
		}})
	rc.Run(t, rc.Spec{Name: "lsp --json: python with a pyproject.toml", Values: pythonValues, Change: pyprojectRoot,
		Call: lspJSON, Message: `"command":"lsp","ok":true`,
		Check: func(c *rc.Case) {
			c.CheckThat("lsp --json: the pyrefly entry", lspJq(c,
				`.data.servers.pyrefly == {"command":"mise","args":["x","--","pyrefly","lsp"],"extensionToLanguage":{".py":"python",".pyi":"python"}}`))
			c.CheckThat("lsp --json: the checks of pyrefly", lspJq(c,
				`.data.checks.pyrefly == {"command":"pyrefly","source":"pipx:pyrefly"}`))
		}})
	rc.Run(t, rc.Spec{Name: "lsp --json: a tool without dev has no server",
		Values: "environments = [ \"dev\", \"ci\" ]\n[tools.go]\nenv = [ \"ci\" ]", Call: lspJSON,
		Message: `"command":"lsp","ok":true`,
		Check: func(c *rc.Case) {
			c.CheckThat("lsp --json: no gopls", lspJq(c,
				`(.data.servers | has("gopls") | not) and (.data.checks | has("gopls") | not)`))
		}})
}
