package values

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// The tests of the merge are in internal/selection. These tests cover the
// small rules of this package alone.

func TestReadRepoSelection(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bootstrap.toml")
	text := "[tools.b]\nenv = [ \"dev\", \"ci\" ]\nk = 1\n\n[tools.a]\n\n[packs.z]\nenv = [ \"dev\" ]\n"
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := ReadRepo(p, "bootstrap.toml")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, it := range v.Sel {
		got = append(got, it.Kind+" "+it.Name+" "+it.Envs)
	}
	want := []string{"tools a -", "tools b dev,ci", "packs z dev"}
	if len(got) != len(want) {
		t.Fatalf("Sel = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Sel[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if j := source.JSON(v.User); j != `{"a":{},"b":{"k":1}}` {
		t.Errorf("User = %s", j)
	}
}

func TestMergeEmpty(t *testing.T) {
	r, err := Merge(Input{})
	if err != nil || len(r.Tools) != 0 || r.Lines() != "" {
		t.Errorf("Merge(empty) = %v, %v", r, err)
	}
}
