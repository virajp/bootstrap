package core

import (
	"reflect"
	"sort"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

func keys(m map[string]bool) []string {
	var out []string
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func TestEffect(t *testing.T) {
	sd := macFixture(t, "format = 2\n[tools.a]\n[tools.b]\n[tools.e]\n")
	st, err := Open(sd)
	if err != nil {
		t.Fatal(problems(err))
	}
	// A removal takes the dependency that no tool needs; the values file
	// does not change.
	e, err := st.Effect([]Item{{Kind: RemoveTool, Names: []string{"b"}}})
	if err != nil {
		t.Fatal(problems(err))
	}
	if got := keys(e.Removed); !reflect.DeepEqual(got, []string{"b", "e"}) || len(e.Added) != 0 {
		t.Errorf("removed %v added %v", got, keys(e.Added))
	}
	// An add after the removal: the named tool is direct, its dependency
	// is not.
	e, err = st.Effect([]Item{{Kind: RemoveTool, Names: []string{"b"}}, {Kind: AddTool, Names: []string{"b"}}})
	if err != nil {
		t.Fatal(problems(err))
	}
	if !reflect.DeepEqual(keys(e.Added), []string{"b", "e"}) || !reflect.DeepEqual(keys(e.Direct), []string{"b"}) || len(keys(e.Removed)) != 2 {
		t.Errorf("added %v direct %v removed %v", keys(e.Added), keys(e.Direct), keys(e.Removed))
	}
	// The problem of an item that Plan refuses.
	if _, err := st.Effect([]Item{{Kind: RemoveTool, Names: []string{"e"}}}); !reflect.DeepEqual(problems(err), []string{"e is needed by b - remove b first"}) {
		t.Errorf("problems %q", problems(err))
	}
}

func TestInitState(t *testing.T) {
	tools := map[string]string{"a": "", "b": "depends = [ \"e\" ]\n", "e": "", "h": "selectable = false\n"}
	sd := fixture(t, source.Macos, tools, "[pack.base]\na = { required = true }\nb = {}\nh = {}\n", "")
	st, err := InitState(sd)
	if err != nil {
		t.Fatal(problems(err))
	}
	for _, n := range []string{"a", "b", "e"} {
		if !st.Selected(n) {
			t.Errorf("%s is not selected", n)
		}
	}
	if st.Named["h"] || !st.Named["b"] || Exists(sd.Values) {
		t.Errorf("named %v, values file there %v", st.Named, Exists(sd.Values))
	}
}
