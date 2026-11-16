package repo

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// TestFileSkipsAny: a [files] or [tasks] entry with two names in
// requires_found is written when any one of them is found, and skipped,
// with both names as "one.json or two.json", when neither is (D38, D39). A
// missing `requires` path still skips it with that path, and three names
// read "a, b or c".
func TestFileSkipsAny(t *testing.T) {
	conds := []cond{
		{"a/x", "requires_found", "one.json"},
		{"a/x", "requires_found", "two.json"},
		{"a/y", "requires", "here"},
		{"a/y", "requires_found", "one.json"},
		{"a/y", "requires_found", "two.json"},
		{"a/z", "requires_found", "one.json"},
		{"a/z", "requires_found", "two.json"},
		{"a/z", "requires_found", "three.json"},
		{"a/w", "requires_found", "one.json"},
	}
	for _, c := range []struct {
		name  string
		files []string
		want  map[string]string
	}{
		{"neither", []string{"here"}, map[string]string{"a/x": "one.json or two.json", "a/y": "one.json or two.json", "a/z": "one.json, two.json or three.json", "a/w": "one.json"}},
		{"first only", []string{"here", "sub/one.json"}, map[string]string{}},
		{"second only", []string{"here", "sub/two.json"}, map[string]string{"a/w": "one.json"}},
		{"no requires path", []string{"sub/two.json"}, map[string]string{"a/y": "here", "a/w": "one.json"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
				t.Fatalf("git init: %s", out)
			}
			for _, f := range c.files {
				writeFile(t, filepath.Join(root, f), "{}\n")
			}
			s := &state{root: root, fileConds: conds}
			s.fileSkips()
			if !reflect.DeepEqual(s.reqSkip, c.want) {
				t.Errorf("reqSkip %v, want %v", s.reqSkip, c.want)
			}
		})
	}
}
