package source

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
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
		"macos/categories.toml": "[category.core]\ntitle = \"Core\"\n",
		"macos/tools/a/tools.toml": `summary = "A"
category = "core"
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
		"macos/tools/b/tools.toml": `summary = "B"
category = "core"
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
		{"files/d/one", ".config/d/one", false, "", 0, 0, "", nil, false},
		{"files/d/sub/two", ".config/d/sub/two", false, "", 0, 0, "", nil, false},
		{"tasks/lib", ".config/mise/tasks/_scripts/lib", false, "", 0, 0, "", nil, false},
		{"tasks/run-it", ".config/mise/tasks/run/it", false, "", 0, 0, "", nil, false},
		{"files/y.json.tmpl", ".config/y.json", true, "", 0, 0, "", nil, false},
		{"files/s.md", "s.md", true, "", 0, 0, "", nil, false},
		{"files/x.txt", "x.txt", false, "", 0, 0, "", nil, false},
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

func TestRootKey(t *testing.T) {
	root := t.TempDir()
	tool := `summary = "A"
category = "core"
[files."home.txt"]
source = "x"
root = "home"
[files.".claude/skills/s/SKILL.md"]
source = "x"
root = "backup"
[files."bad.txt"]
source = "x"
root = "clone"
[files."dotfiles/x"]
source = "x"
root = "backup"
[files."DotFiles/x"]
source = "x"
root = "backup"
[files."a/.Git/x"]
source = "x"
root = "backup"
[files."m.json"]
source = "m.merge.toml"
mode = "merge"
root = "backup"
[files."@backup/x"]
source = "x"
[files."r.txt"]
source = "x"
root = "backup"
requires = [ "/abs" ]
`
	write(t, root, map[string]string{
		"macos/categories.toml":            "[category.core]\ntitle = \"Core\"\n",
		"macos/tools/a/tools.toml":         tool,
		"macos/tools/a/files/x":            "x",
		"macos/tools/a/files/m.merge.toml": "",
		"repo/categories.toml":             "[category.core]\ntitle = \"Core\"\n",
		"repo/tools/a/tools.toml":          repoTool("summary = \"A\"\ncategory = \"core\"\n[files.\"home.txt\"]\nsource = \"x\"\nroot = \"home\"\n"),
		"repo/tools/a/files/x":             "x",
	})
	x, err := Source{Root: root, Side: Macos}.Open()
	var got []string
	for _, f := range x.Files("a") {
		got = append(got, f.Target+" "+map[bool]string{true: "backup", false: "home"}[f.Backup])
	}
	if want := []string{".claude/skills/s/SKILL.md backup", "home.txt home"}; !reflect.DeepEqual(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
	var msgs []string
	for _, e := range Messages(err) {
		msgs = append(msgs, e.Msg)
	}
	const at = "macos/tools/a/tools.toml: "
	wantMsgs := []string{
		at + `[files."@backup/x"] must be a path relative to the target root, with no .. part`,
		at + `[files."DotFiles/x"] a target of root = "backup" cannot be below dotfiles/ or a .git folder of the backup clone`,
		at + `[files."a/.Git/x"] a target of root = "backup" cannot be below dotfiles/ or a .git folder of the backup clone`,
		at + `[files."bad.txt"] root must be "home" or "backup"`,
		at + `[files."dotfiles/x"] a target of root = "backup" cannot be below dotfiles/ or a .git folder of the backup clone`,
		at + `[files."m.json"] root = "backup" works only with mode = "copy", "template" or "seed"`,
		at + `[files."r.txt"] requires: '/abs' must be a path relative to the backup clone, with no .. part`,
	}
	if !reflect.DeepEqual(msgs, wantMsgs) {
		t.Errorf("messages:\n%v\nwant\n%v", strings.Join(msgs, "\n"), strings.Join(wantMsgs, "\n"))
	}

	// On the repo side, root is an unknown key, with the list of the repo
	// keys.
	_, err = Source{Root: root, Side: Repo}.Open()
	msgs = nil
	for _, e := range Messages(err) {
		msgs = append(msgs, e.Msg)
	}
	want := `repo/tools/a/tools.toml: [files."home.txt"] root is not a key of a files entry - the keys are source, mode, render, format, requires, requires_found, when, perm, dir_perm`
	if !slices.Contains(msgs, want) {
		t.Errorf("repo messages = %v, want %s", msgs, want)
	}
}
