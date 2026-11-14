package questions

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

// index gives the Mac index of one tool t with the questions of <toml>.
func index(t *testing.T, toml string) *source.Index {
	t.Helper()
	root := t.TempDir()
	p := filepath.Join(root, "macos", "tools", "t", "tools.toml")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	x, err := source.Source{Root: root, Side: source.Macos}.Open()
	if err != nil {
		t.Fatal(err)
	}
	return x
}

// answers gives the [tools.t] table of a values file.
func answers(kv map[string]any) *source.Obj {
	t := source.NewObj()
	for k, v := range kv {
		t.Set(k, v)
	}
	o := source.NewObj()
	o.Set("t", t)
	return o
}

func names(gs []Gap) []string {
	var out []string
	for _, g := range gs {
		out = append(out, g.Name())
	}
	return out
}

const groupTOML = `
[questions.a]
question = "A?"
group = "g"
note = "a note"
[questions.b]
question = "B?"
group = "g"
[questions.c]
question = "C?"
optional = true
[questions.d]
question = "D?"
`

func TestOfReadsTheNewKeys(t *testing.T) {
	qs := Of(index(t, groupTOML), []string{"t"})
	if qs[0].Group != "g" || qs[0].Note != "a note" || !qs[2].Optional || qs[3].Optional {
		t.Errorf("questions = %+v", qs)
	}
	if qs[0].Necessary() || qs[2].Necessary() || !qs[3].Necessary() {
		t.Error("Necessary")
	}
}

func TestMissing(t *testing.T) {
	x := index(t, groupTOML)
	cases := []struct {
		name string
		user map[string]any
		want []string
	}{
		{"nothing", map[string]any{}, []string{"t.g", "t.d"}},
		{"empty group, empty d", map[string]any{"a": "", "b": "", "d": ""}, []string{"t.g"}},
		{"one answer of the group", map[string]any{"b": "x", "d": "y"}, nil},
		{"optional empty", map[string]any{"a": "x", "c": "", "d": "y"}, nil},
	}
	for _, c := range cases {
		got := Missing(x, []string{"t"}, answers(c.user), nil)
		if !reflect.DeepEqual(names(got), c.want) {
			t.Errorf("%s: %v, want %v", c.name, names(got), c.want)
		}
	}
	g := Missing(x, []string{"t"}, answers(map[string]any{"d": "y"}), nil)[0]
	if !g.Group || !reflect.DeepEqual(g.Keys, []string{"a", "b"}) {
		t.Errorf("gap = %+v", g)
	}
}

func TestEmpty(t *testing.T) {
	if !Empty("") || Empty("x") || Empty(false) || Empty(nil) {
		t.Error("Empty")
	}
}
