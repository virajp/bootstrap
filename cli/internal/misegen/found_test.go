package misegen

import (
	"testing"

	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// TestRequiresFoundAny: a mise tool with two names in requires_found is in
// the config when any one of them is found, and not when neither is (D38).
func TestRequiresFoundAny(t *testing.T) {
	r := &selection.Resolved{
		Index: &source.Index{Manifests: map[string]*source.Manifest{"a": {
			Tool: "a",
			Mise: []source.MiseTool{{Name: "pa", RequiresFound: []string{"one.json", "two.json"}}, {Name: "pb"}},
		}}},
		Sel:   []selection.Entry{{Tool: "a", Envs: "dev", Src: "tools.a"}},
		Tools: []string{"a"},
	}
	for _, c := range []struct {
		name  string
		found []string
		want  string
	}{
		{"neither", nil, " pb:dev "},
		{"first only", []string{"one.json"}, " pa:dev pb:dev "},
		{"second only", []string{"two.json"}, " pa:dev pb:dev "},
		{"both", []string{"one.json", "two.json"}, " pa:dev pb:dev "},
	} {
		has := func(name string) bool {
			for _, f := range c.found {
				if f == name {
					return true
				}
			}
			return false
		}
		e, err := NewEnvs(source.Repo, []string{"dev"}, r, "repo/tools", has)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := e.MiseWords(); got != c.want {
			t.Errorf("%s: mise words %q, want %q", c.name, got, c.want)
		}
	}
}
