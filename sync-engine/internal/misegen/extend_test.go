package misegen

import (
	"fmt"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// A mise tool depends on the mise tools of the depends and the parent of its
// tool, then on the mise tools above it in the same tool.
func TestDepends(t *testing.T) {
	idx := &source.Index{Manifests: map[string]*source.Manifest{
		"node":   {MiseOrder: []string{"node"}},
		"python": {MiseOrder: []string{"python", "uv", "ruff"}},
		"pnpm":   {Depends: []string{"node"}, Parent: "node", MiseOrder: []string{"pnpm"}},
		"gcloud": {Depends: []string{"python", "gone"}, MiseOrder: []string{"gcloud", "extra"}},
	}}
	g := &Generator{Side: source.Macos, Index: idx, Tools: []string{"node", "python", "pnpm", "gcloud"}}
	for tool, want := range map[string]string{
		"node":   "map[]",
		"pnpm":   "map[pnpm:[node]]",
		"python": "map[ruff:[python uv] uv:[python]]",
		"gcloud": "map[extra:[python uv ruff gcloud] gcloud:[python uv ruff]]",
	} {
		if got := fmt.Sprint(g.depends(tool, "")); got != want {
			t.Errorf("depends(%s) = %s, want %s", tool, got, want)
		}
	}
	// On the repo side, only the mise tools in the config of the environment.
	g = &Generator{Side: source.Repo, Index: idx, Tools: []string{"python", "gcloud"},
		Envs: &Envs{mise: words{"python:dev", "uv:dev", "gcloud:dev", "extra:dev"}}}
	if got, want := fmt.Sprint(g.depends("gcloud", "dev")), "map[extra:[python uv gcloud] gcloud:[python uv]]"; got != want {
		t.Errorf("repo depends = %s, want %s", got, want)
	}
}

func TestExtend(t *testing.T) {
	d := source.FromMap(map[string]any{"mise": map[string]any{"shell_alias": map[string]any{"a": "x"}}}).(*source.Obj)
	idx := &source.Index{Manifests: map[string]*source.Manifest{"t": {MiseOrder: []string{"one", "two"}}}}
	g := &Generator{Side: source.Macos, Index: idx, Tools: []string{"t"}, Derived: map[string]*source.Obj{"t": d}}
	in := map[string]any{"mise": map[string]any{
		"shell_alias": map[string]any{"b": "y"},
		"tools":       map[string]any{"one": "1", "two": map[string]any{"version": "2"}},
	}}
	got := fmt.Sprint(g.extend("t", "", in))
	want := "map[mise:map[shell_alias:map[a:x b:y] tools:map[one:1 two:map[depends:[one] version:2]]]]"
	if got != want {
		t.Errorf("extend = %s, want %s", got, want)
	}
}
