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

func TestFilesRepoAndMacos(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"repo/tools/a/tools.toml":      "",
		"repo/tools/a/x.txt":           "x",
		"repo/tools/a/y.json.tmpl":     "y",
		"repo/tools/a/sub/tools.toml":  "",
		"repo/tools/a/sub/z":           "z",
		"macos/tools/a/tools.toml":     "",
		"macos/tools/a/x.txt":          "x",
		"macos/tools/a/sub/tools.toml": "",
	})
	if err := os.Symlink("x.txt", filepath.Join(root, "repo/tools/a/link")); err != nil {
		t.Fatal(err)
	}
	got := func(side Side) []File {
		f, err := Source{Root: root, Side: side}.Files("a")
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	wantRepo := []File{{"sub/z", "sub/z", false}, {"x.txt", "x.txt", false}, {"y.json.tmpl", "y.json", true}}
	if g := got(Repo); !reflect.DeepEqual(g, wantRepo) {
		t.Errorf("repo files = %v, want %v", g, wantRepo)
	}
	wantMacos := []File{{"sub/tools.toml", "sub/tools.toml", false}, {"x.txt", "x.txt", false}}
	if g := got(Macos); !reflect.DeepEqual(g, wantMacos) {
		t.Errorf("macos files = %v, want %v", g, wantMacos)
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
