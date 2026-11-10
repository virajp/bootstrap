package values

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

func obj(m map[string]any) *source.Obj { return source.FromMap(m).(*source.Obj) }

// The value order: defaults, contributions, the derive output, the values
// file, then the overrides file.
func TestMergeOrder(t *testing.T) {
	idx := &source.Index{Names: []string{"a", "b"}, Manifests: map[string]*source.Manifest{
		"a": {Tool: "a", Raw: obj(map[string]any{"defaults": map[string]any{"d": 1, "c": 1, "r": 1, "u": 1, "o": 1}})},
		"b": {Tool: "b", Raw: obj(map[string]any{"contributes": map[string]any{"a": map[string]any{"c": 2, "r": 2}}})},
	}}
	var gotBase string
	in := Input{
		Index:     idx,
		Selected:  []string{"a", "b"},
		User:      obj(map[string]any{"a": map[string]any{"u": 4, "o": 4}}),
		Overrides: obj(map[string]any{"a": map[string]any{"o": 5}}),
		Derive: func(tools []string, base map[string]*source.Obj) (map[string]*source.Obj, error) {
			gotBase = source.JSON(base["a"])
			return map[string]*source.Obj{"a": obj(map[string]any{"r": 3, "u": 3})}, nil
		},
	}
	r, err := Merge(in)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"c":2,"d":1,"o":1,"r":2,"u":1}`; gotBase != want {
		t.Errorf("derive input = %s, want %s", gotBase, want)
	}
	if got, want := source.JSON(r.Values["a"]), `{"c":2,"d":1,"o":5,"r":3,"u":4}`; got != want {
		t.Errorf("values = %s, want %s", got, want)
	}
}

func TestReadOverrides(t *testing.T) {
	dir := t.TempDir()
	values := filepath.Join(dir, "bootstrap.toml")
	if got := OverridesPath(values); got != filepath.Join(dir, "bootstrap.overrides.toml") {
		t.Errorf("OverridesPath = %s", got)
	}
	if o, err := ReadOverrides(OverridesPath(values), "x"); o != nil || err != nil {
		t.Errorf("missing file = %v, %v", o, err)
	}
	write := func(text string) {
		if err := os.WriteFile(OverridesPath(values), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("[tools.git]\nk = 1\n")
	if o, err := ReadOverrides(OverridesPath(values), "x"); err != nil || source.JSON(o) != `{"git":{"k":1}}` {
		t.Errorf("tools = %v, %v", o, err)
	}
	write("repo_name = \"x\"\n")
	if _, err := ReadOverrides(OverridesPath(values), "x"); err == nil || !strings.Contains(err.Error(), "its only key is [tools.<name>]") {
		t.Errorf("other key = %v", err)
	}
}
