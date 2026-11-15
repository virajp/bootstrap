package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCurrentDirStalePWD is `mise x --cd=sub -- bootstrap`: mise changes the
// folder, but PWD still names the parent folder.
func TestCurrentDirStalePWD(t *testing.T) {
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(parent, "sub")
	if err := os.Mkdir(sub, 0o777); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	for _, c := range []struct{ name, pwd, want string }{
		{"stale", parent, sub},
		{"current", sub, sub},
		{"empty", "", sub},
	} {
		getenv := func(k string) string {
			if k == "PWD" {
				return c.pwd
			}
			return ""
		}
		if got := currentDir(getenv); got != c.want {
			t.Errorf("%s: currentDir = %q, want %q", c.name, got, c.want)
		}
	}
}
