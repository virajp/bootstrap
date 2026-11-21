package source

import (
	"reflect"
	"strings"
	"testing"
)

// TestEachEntry: a [tasks] entry of a repo tool may have each, a value path
// with {name} parts that the path of the entry has too.
func TestEachEntry(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"repo/categories.toml": "[category.core]\ntitle = \"Core\"\n",
		"repo/tools/a/tools.toml": lspTool(`
[tasks."rel/{project}/{kind}.py"]
source = "rel.py.tmpl"
mode = "template"
each = "release.{project}.{kind}"
`),
		"repo/tools/a/tasks/rel.py.tmpl": "x",
		"repo/tools/b/tools.toml": lspTool(`
[tasks."one/{p}.py"]
source = "one"
each = "release.{p}"
[tasks."two/{p}.py"]
source = "two"
mode = "template"
each = "release.{q}"
[tasks."three/{p}.py"]
source = "three"
mode = "template"
[tasks."four/{p}.py"]
source = "four"
mode = "template"
each = "{p}.x"
[tasks."five/{value}.py"]
source = "five"
mode = "template"
each = "release.{value}"
`),
		"repo/tools/b/tasks/one":   "1",
		"repo/tools/b/tasks/two":   "2",
		"repo/tools/b/tasks/three": "3",
		"repo/tools/b/tasks/four":  "4",
		"repo/tools/b/tasks/five":  "5",
	})
	x, err := Source{Root: root, Side: Repo}.Open()
	if f := x.Files("a"); len(f) != 1 || f[0].Each != "release.{project}.{kind}" {
		t.Errorf("a: files %+v", f)
	}
	var got []string
	for _, e := range Messages(err) {
		got = append(got, e.Msg)
	}
	const at = "repo/tools/b/tools.toml: "
	bad := `each must be a value path with {name} parts, for example "release.{project}.{kind}"`
	want := []string{
		at + `[tasks."five/{value}.py"] ` + bad,
		at + `[tasks."four/{p}.py"] ` + bad,
		at + `[tasks."one/{p}.py"] each needs mode = "template"`,
		at + `[tasks."three/{p}.py"] a {name} part of the path needs each`,
		at + `[tasks."two/{p}.py"] the path must have each {name} part of each, and no other`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("messages:\n%v\nwant\n%v", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// TestEachMacos: each is for repo tools only.
func TestEachMacos(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"macos/categories.toml": "[category.core]\ntitle = \"Core\"\n",
		"macos/tools/a/tools.toml": `summary = "A"
category = "core"
[tasks."x/{p}"]
source = "x"
mode = "template"
each = "y.{p}"
`,
		"macos/tools/a/tasks/x": "x",
	})
	_, err := Source{Root: root, Side: Macos}.Open()
	if m := Messages(err); len(m) != 1 || !strings.HasSuffix(m[0].Msg, "each is for repo tools only") {
		t.Errorf("messages %+v", m)
	}
}

func TestExpand(t *testing.T) {
	values := map[string]any{
		"release": map[string]any{
			"web": map[string]any{
				"prod":  map[string]any{"type": "tag"},
				"local": map[string]any{"type": "task"},
			},
			"api": map[string]any{"prod": map[string]any{"type": "tag"}},
			"odd": "not a table",
		},
	}
	f := File{Path: "tasks/rel.py.tmpl", Target: ".config/mise/tasks/rel/{project}/{kind}.py", Tmpl: true, Each: "release.{project}.{kind}"}
	got, err := Expand(values, f)
	if err != nil {
		t.Fatal(err)
	}
	var targets []string
	for _, g := range got {
		targets = append(targets, g.Target)
	}
	want := []string{
		".config/mise/tasks/rel/api/prod.py",
		".config/mise/tasks/rel/web/local.py",
		".config/mise/tasks/rel/web/prod.py",
	}
	if !reflect.DeepEqual(targets, want) {
		t.Errorf("targets %v, want %v", targets, want)
	}
	if v := got[1].Vars; v["project"] != "web" || v["kind"] != "local" || !reflect.DeepEqual(v[EachValue], map[string]any{"type": "task"}) {
		t.Errorf("vars %+v", v)
	}

	if got, err := Expand(map[string]any{}, f); err != nil || len(got) != 0 {
		t.Errorf("no table: %v %v", got, err)
	}
	bad := map[string]any{"release": map[string]any{"../x": map[string]any{"prod": map[string]any{}}}}
	if _, err := Expand(bad, f); err == nil || !strings.Contains(err.Error(), "'../x'") {
		t.Errorf("bad key: %v", err)
	}
}
