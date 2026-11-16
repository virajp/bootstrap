package misegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// TestMiseEnvRequiresFound: a [mise_env.dev.tools] entry with requires_found
// is in mise.dev.toml only when any one of its names is found, and the file
// never has the key (D40). An entry without requires_found is always there.
func TestMiseEnvRequiresFound(t *testing.T) {
	clone := t.TempDir()
	for path, body := range map[string]string{
		"repo/categories.toml": "[category.core]\ntitle = \"Core\"\n",
		"repo/tools/a/tools.toml": `summary = "T"
category = "core"
default_envs = [ "dev" ]
depends = []
root_only = false

[mise_env.dev.tools.ts]
version = "1"
requires_found = [ "package.json", "tsconfig.json" ]

[mise_env.dev.tools.other]
version = "2"
`,
	} {
		p := filepath.Join(clone, path)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	src := source.Source{Root: clone, Side: source.Repo}
	idx, err := src.Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	for _, c := range []struct {
		name  string
		found []string
		want  string
	}{
		{"neither", nil, "[tools]\n[tools.other]\nversion = \"2\"\n"},
		{"first only", []string{"package.json"}, "[tools]\n[tools.other]\nversion = \"2\"\n[tools.ts]\nversion = \"1\"\n"},
		{"second only", []string{"tsconfig.json"}, "[tools]\n[tools.other]\nversion = \"2\"\n[tools.ts]\nversion = \"1\"\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			has := func(name string) bool {
				for _, f := range c.found {
					if f == name {
						return true
					}
				}
				return false
			}
			r := &selection.Resolved{Index: idx, Sel: []selection.Entry{{Tool: "a", Envs: "dev", Src: "tools.a"}}, Tools: []string{"a"}}
			envs, err := NewEnvs(source.Repo, []string{"dev"}, r, "repo/tools", has)
			if err != nil {
				t.Fatalf("envs: %v", err)
			}
			g := &Generator{Side: source.Repo, ToolsDir: src.ToolsDir(), Shown: "repo/tools", Envs: envs, Index: idx, Tools: r.Tools}
			defer func() { _ = g.Close() }()
			confd := t.TempDir()
			if _, err := g.Render("a", confd); err != nil {
				t.Fatalf("render: %v", err)
			}
			got, err := os.ReadFile(filepath.Join(confd, "a", "mise.dev.toml"))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			body := string(got)
			if i := strings.Index(body, "\n\n"); i >= 0 {
				body = body[i+2:]
			}
			if body != c.want {
				t.Errorf("mise.dev.toml:\n%s\nwant\n%s", got, c.want)
			}
		})
	}
}
