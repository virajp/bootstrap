package source

import (
	"reflect"
	"strings"
	"testing"
)

// TestReplaces: a [tasks] entry may have replaces, a list of sha256:<hex>;
// another form of an item is a problem (D4).
func TestReplaces(t *testing.T) {
	good := "sha256:" + strings.Repeat("ab", 32)
	root := t.TempDir()
	write(t, root, map[string]string{
		"macos/categories.toml": "[category.core]\ntitle = \"Core\"\n",
		"macos/tools/a/tools.toml": `summary = "A"
category = "core"
[tasks."grp/40-a"]
source = "part"
replaces = ["` + good + `"]
`,
		"macos/tools/a/tasks/part": "p",
		"macos/tools/b/tools.toml": `summary = "B"
category = "core"
[tasks."grp/40-b"]
source = "one"
replaces = ["md5:00", "sha256:` + strings.Repeat("AB", 32) + `"]
[tasks."grp/40-c"]
source = "two"
replaces = []
[tasks."top"]
source = "three"
replaces = ["` + good + `"]
[files."x"]
source = "x"
replaces = ["` + good + `"]
`,
		"macos/tools/b/tasks/one":   "1",
		"macos/tools/b/tasks/two":   "2",
		"macos/tools/b/tasks/three": "3",
		"macos/tools/b/files/x":     "x",
	})
	x, err := Source{Root: root, Side: Macos}.Open()
	fe := x.Manifests["a"].Files
	if len(fe) != 1 || !reflect.DeepEqual(fe[0].Replaces, []string{good}) {
		t.Errorf("a: files %+v", fe)
	}
	var got []string
	for _, e := range Messages(err) {
		got = append(got, e.Msg)
	}
	want := []string{
		`macos/tools/b/tools.toml: [files."x"] replaces is not a key of a files entry - the keys are source, mode, render, format, requires, requires_found, when, perm, dir_perm`,
		`macos/tools/b/tools.toml: [tasks."grp/40-b"] replaces: 'md5:00' must be "sha256:<64 hex>", in lower case`,
		`macos/tools/b/tools.toml: [tasks."grp/40-b"] replaces: 'sha256:` + strings.Repeat("AB", 32) + `' must be "sha256:<64 hex>", in lower case`,
		`macos/tools/b/tools.toml: [tasks."grp/40-c"] replaces must be a non-empty list of "sha256:<64 hex>" strings`,
		`macos/tools/b/tools.toml: [tasks."top"] replaces is for a part in the folder of its group`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("messages:\n%v\nwant\n%v", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
