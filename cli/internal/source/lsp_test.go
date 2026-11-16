package source

import (
	"reflect"
	"testing"
)

// lspTool is a repo tool with the keys that the checks need, then <rest>.
func lspTool(rest string) string {
	return "summary = \"T\"\ncategory = \"core\"\ndefault_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n" + rest
}

func TestLSPParse(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"repo/categories.toml": "[category.core]\ntitle = \"Core\"\n",
		"repo/tools/a/tools.toml": lspTool(`
[mise.tools.pa]
version = "latest"

[mise_env.dev.tools.da]
version = "latest"

[lsp.pa-lsp]
command = "pa"
args = [ "lsp", "--stdio" ]
extensions = { ".py" = "python", ".pyi" = "python" }
source = "pa"

[lsp.da-lsp]
command = "da"
extensions = { ".d" = "d" }
startup_timeout = 1000
source = "da"

[lsp.host-lsp]
command = "h"
extensions = { ".h" = "h" }
source = "host"
`),
	})
	x, err := Source{Root: root, Side: Repo}.Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	want := []LSPServer{
		{Name: "da-lsp", Command: "da", Args: []string{}, Extensions: map[string]string{".d": "d"}, StartupTimeout: 1000, Source: "da"},
		{Name: "host-lsp", Command: "h", Args: []string{}, Extensions: map[string]string{".h": "h"}, Source: "host"},
		{Name: "pa-lsp", Command: "pa", Args: []string{"lsp", "--stdio"}, Extensions: map[string]string{".py": "python", ".pyi": "python"}, Source: "pa"},
	}
	if got := x.Manifests["a"].LSP; !reflect.DeepEqual(got, want) {
		t.Errorf("LSP = %+v, want %+v", got, want)
	}
	if x.LSPOwner["pa-lsp"] != "a" {
		t.Errorf("LSPOwner = %v", x.LSPOwner)
	}
}

// TestLSPRules: each wrong [lsp] shape stops the check with code 2 (D7, D30).
func TestLSPRules(t *testing.T) {
	cases := []struct {
		name, a, b string
		want       []string
	}{
		{"no command", "[lsp.s]\nextensions = { \".x\" = \"x\" }\nsource = \"host\"\n", "",
			[]string{"repo/tools/a/tools.toml: [lsp.s] needs command, the binary of the server (a string)"}},
		{"command with whitespace", "[lsp.s]\ncommand = \"x lsp\"\nextensions = { \".x\" = \"x\" }\nsource = \"host\"\n", "",
			[]string{"repo/tools/a/tools.toml: [lsp.s] command 'x lsp' has whitespace - command is one binary; put the rest in args"}},
		{"no extensions", "[lsp.s]\ncommand = \"x\"\nsource = \"host\"\n", "",
			[]string{`repo/tools/a/tools.toml: [lsp.s] needs extensions, a table of ".<ext>" = "<language id>" that is not empty`}},
		{"empty extensions", "[lsp.s]\ncommand = \"x\"\nextensions = {}\nsource = \"host\"\n", "",
			[]string{`repo/tools/a/tools.toml: [lsp.s] needs extensions, a table of ".<ext>" = "<language id>" that is not empty`}},
		{"bad extension", "[lsp.s]\ncommand = \"x\"\nextensions = { \"x\" = \"x\" }\nsource = \"host\"\n", "",
			[]string{`repo/tools/a/tools.toml: [lsp.s] extensions: 'x' must be ".<ext>" = "<language id>"`}},
		{"bad args", "[lsp.s]\ncommand = \"x\"\nargs = \"--stdio\"\nextensions = { \".x\" = \"x\" }\nsource = \"host\"\n", "",
			[]string{"repo/tools/a/tools.toml: [lsp.s] args must be a list of strings"}},
		{"bad timeout", "[lsp.s]\ncommand = \"x\"\nextensions = { \".x\" = \"x\" }\nstartup_timeout = \"1s\"\nsource = \"host\"\n", "",
			[]string{"repo/tools/a/tools.toml: [lsp.s] startup_timeout must be an integer of milliseconds above 0"}},
		{"unknown key", "[lsp.s]\ncommand = \"x\"\nextensions = { \".x\" = \"x\" }\nsource = \"host\"\nenv = 1\n", "",
			[]string{"repo/tools/a/tools.toml: [lsp.s] env is not a key of a server - the keys are command, args, extensions, startup_timeout and source"}},
		{"no source", "[lsp.s]\ncommand = \"x\"\nextensions = { \".x\" = \"x\" }\n", "",
			[]string{`repo/tools/a/tools.toml: [lsp.s] needs source, the mise tool of this tool that provides the command, or "host"`}},
		{"not a table", "lsp = 1\n", "",
			[]string{"repo/tools/a/tools.toml: [lsp] must be a table of [lsp.<server>] tables"}},
		{"source of another tool", "[lsp.s]\ncommand = \"x\"\nextensions = { \".x\" = \"x\" }\nsource = \"pb\"\n", "[mise.tools.pb]\nversion = \"1\"\n",
			[]string{`repo/tools/a/tools.toml: source 'pb' of [lsp.s] is not a mise tool of this tool - name its pin of [mise.tools] or [mise_env.<env>.tools], or write "host"`}},
		{"one name in two tools", "[lsp.s]\ncommand = \"x\"\nextensions = { \".x\" = \"x\" }\nsource = \"host\"\n", "[lsp.s]\ncommand = \"y\"\nextensions = { \".y\" = \"y\" }\nsource = \"host\"\n",
			[]string{"language server 's' is defined in repo/tools/a and in repo/tools/b - keep one"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, map[string]string{
				"repo/categories.toml":    "[category.core]\ntitle = \"Core\"\n",
				"repo/tools/a/tools.toml": lspTool(c.a),
				"repo/tools/b/tools.toml": lspTool(c.b),
			})
			_, err := Source{Root: root, Side: Repo}.Open()
			var got []string
			for _, e := range Messages(err) {
				got = append(got, e.Msg)
				if e.Code != 2 {
					t.Errorf("code %d for %s", e.Code, e.Msg)
				}
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("messages:\n%q\nwant\n%q", got, c.want)
			}
		})
	}
}
