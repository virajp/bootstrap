package target

import "testing"

// The parity of the writes with bash is in the tests of internal/manifest.
func TestDirnameShow(t *testing.T) {
	for in, want := range map[string]string{
		"/a/b": "/a", "/a/b/": "/a", "a": ".", "/a": "/", "/": "/", "//": "/",
		"a//b": "a", "/a/../b": "/a/..", "a/b//": "a",
	} {
		if got := Dirname(in); got != want {
			t.Errorf("Dirname(%q) = %q, want %q", in, got, want)
		}
	}
	for _, c := range [][3]string{
		{"/r/a/b", "/r", "a/b"}, {"/r", "/r", "/r"}, {"/rx/a", "/r", "/rx/a"}, {"/a", "/", "/a"},
	} {
		if got := Show(c[0], c[1]); got != c[2] {
			t.Errorf("Show(%q, %q) = %q, want %q", c[0], c[1], got, c[2])
		}
	}
}
