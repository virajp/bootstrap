package source

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for p, c := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEntries(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"macos/tools/a/tools.toml": `
[files."x.txt"]
source = "x.txt"
[files.".config/y.json"]
source = "y.json.tmpl"
mode = "template"
[files.".config/d"]
source = "d"
[files."s.md"]
source = "s.md"
mode = "seed"
render = true
[tasks."_scripts/lib"]
source = "lib"
library = true
[tasks."run/it"]
source = "run-it"
`,
		"macos/tools/a/files/x.txt":       "x",
		"macos/tools/a/files/y.json.tmpl": "y",
		"macos/tools/a/files/d/one":       "1",
		"macos/tools/a/files/d/sub/two":   "2",
		"macos/tools/a/files/s.md":        "s",
		"macos/tools/a/tasks/lib":         "l",
		"macos/tools/a/tasks/run-it":      "r",
		"macos/tools/b/tools.toml": `
[files."x"]
source = "x"
mode = "copy"
render = true
[tasks."lib"]
source = "lib"
library = true
`,
		"macos/tools/b/files/x":     "x",
		"macos/tools/b/files/extra": "e",
		"macos/tools/b/tasks/lib":   "l",
		"macos/tools/b/old.txt":     "o",
	})
	x, err := Source{Root: root, Side: Macos}.Open()
	want := []File{
		{"files/d/one", ".config/d/one", false, "", 0, 0},
		{"files/d/sub/two", ".config/d/sub/two", false, "", 0, 0},
		{"tasks/lib", ".config/mise/tasks/_scripts/lib", false, "", 0, 0},
		{"tasks/run-it", ".config/mise/tasks/run/it", false, "", 0, 0},
		{"files/y.json.tmpl", ".config/y.json", true, "", 0, 0},
		{"files/s.md", "s.md", true, "", 0, 0},
		{"files/x.txt", "x.txt", false, "", 0, 0},
	}
	if g := x.Files("a"); !reflect.DeepEqual(g, want) {
		t.Errorf("files = %v, want %v", g, want)
	}
	var got []string
	for _, e := range Messages(err) {
		got = append(got, e.Msg)
	}
	wantMsgs := []string{
		`macos/tools/b/tools.toml: [files."x"] render works only with mode = "seed" - a template has mode = "template"`,
		`macos/tools/b/tools.toml: [tasks."lib"] library = true is for a library below _scripts/, and each file below _scripts/ is a library`,
		`macos/tools/b/tools.toml: files/extra has no entry - write a [files."<path>"] table with source = "extra"`,
		`macos/tools/b/tools.toml: old.txt is not below files/ or tasks/ - move it, and write an entry for it (the old form of a tool folder is gone)`,
	}
	if !reflect.DeepEqual(got, wantMsgs) {
		t.Errorf("messages:\n%v\nwant\n%v", got, wantMsgs)
	}
}

func TestObjJSON(t *testing.T) {
	o, err := ReadTOML(writeFile(t, "b = 1\na = \"q\\\"\\\\\\n\\t\\u007f\\u00e9<>&\"\nc = [ 1, 2.5, true ]\n[t]\nz = 1\ny = 2\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":"q\"\\\n\t\u007fé<>&","b":1,"c":[1,2.5,true],"t":{"y":2,"z":1}}`
	if got := JSON(o); got != want {
		t.Errorf("JSON = %s, want %s", got, want)
	}
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.toml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEqual(t *testing.T) {
	a, _ := ReadTOML(writeFile(t, "x = 1\ny = [ 1, 2 ]\n"))
	b, _ := ReadTOML(writeFile(t, "y = [ 1, 2 ]\nx = 1.0\n"))
	if !Equal(a, b) {
		t.Error("objects with the same values and numbers 1 and 1.0 must be equal")
	}
	c, _ := ReadTOML(writeFile(t, "y = [ 2, 1 ]\nx = 1\n"))
	if Equal(a, c) {
		t.Error("lists with another order must differ")
	}
}

func TestNameOK(t *testing.T) {
	for name, want := range map[string]bool{"a": true, "a-b": true, "": false, "a/b": false, ".a": false, "a b": false, "a\tb": false} {
		if got := NameOK(name); got != want {
			t.Errorf("NameOK(%q) = %v, want %v", name, got, want)
		}
	}
}
