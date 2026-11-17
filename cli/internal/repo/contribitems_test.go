package repo

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// contribKeys are the keys that each tool of the contributed-items clone has.
const contribKeys = "summary = \"A tool\"\ncategory = \"core\"\ndefault_envs = [\"dev\"]\ndepends = []\nroot_only = false\n"

// contribClone writes a clone with the tool own, which ships the merge-mode
// file list.txt, and the tools <extra> (name to tools.toml body, after the
// common keys). It gives the clone.
func contribClone(t *testing.T, extra map[string]string) string {
	t.Helper()
	clone := t.TempDir()
	writeFile(t, filepath.Join(clone, "cli", "VERSION"), "0.0.1\n")
	writeFile(t, filepath.Join(clone, "repo", "categories.toml"), "[category.core]\ntitle = \"Core\"\n")
	writeFile(t, filepath.Join(clone, "repo", "packs.toml"), "[pack.base]\n")
	writeFile(t, filepath.Join(clone, "repo", "tools", "own", "tools.toml"),
		contribKeys+"\n[files.\"list.txt\"]\nmode = \"merge\"\nsource = \"list.merge.toml\"\n")
	writeFile(t, filepath.Join(clone, "repo", "tools", "own", "files", "list.merge.toml"),
		"format = \"text\"\ncomment = \"#\"\n\n[[item]]\nvalue = \"a\"\n\n[[item]]\nvalue = \"b\"\n")
	for name, body := range extra {
		writeFile(t, filepath.Join(clone, "repo", "tools", name, "tools.toml"), contribKeys+body)
	}
	// The sync reads a git clone.
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"}, {"add", "-A"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "c"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", clone}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	return clone
}

// contribSync runs a sync of a new repo that selects <tools> with <toolsToml>
// after them, against <clone>. It gives the code, stderr and the root.
func contribSync(t *testing.T, clone string, tools []string, toolsToml string) (fail.Code, string, string) {
	t.Helper()
	body := "format = 3\n"
	for _, n := range tools {
		body += "\n[tools." + n + "]\n"
		if n == "own" {
			body += toolsToml
		}
	}
	root := checkRepo(t, repoHead+"environments = [\"dev\"]\n", body)
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	t.Setenv("BOOTSTRAP_CLONE", clone)
	var out, errs bytes.Buffer
	code := Run(nil, &out, &errs)
	return code, errs.String(), root
}

// lines gives the lines of the file <p> that are not a comment or empty.
func lines(t *testing.T, p string) []string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		if l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}

// TestContributedItems: a selected tool adds its items to the merge-mode file
// of a selected owner, after the owner's items, in the order of the tool
// names, each line once; an override of the owner removes a contributed
// line; an override of a line that only an unselected tool contributes is not
// unused (D1, D7, D17).
func TestContributedItems(t *testing.T) {
	extra := map[string]string{
		"zeta":  "\n[[contributes_items.own.\"list.txt\".item]]\nvalue = \"z\"\n\n[[contributes_items.own.\"list.txt\".item]]\nvalue = \"a\"\n",
		"alpha": "\n[[contributes_items.own.\"list.txt\".item]]\nvalue = \"y\"\n\n[[contributes_items.own.\"list.txt\".item]]\nvalue = \"z\"\n",
		"gamma": "\n[[contributes_items.own.\"list.txt\".item]]\nvalue = \"g\"\n",
	}
	cases := []struct {
		name  string
		tools []string
		over  string // Lines of [tools.own].
		want  string
		fail  string
	}{
		{"no contributor", []string{"own"}, "", "a b", ""},
		{"one contributor", []string{"own", "zeta"}, "", "a b z", ""},
		{"two contributors in name order, each line once", []string{"own", "zeta", "alpha"}, "", "a b y z", ""},
		{"an override removes a contributed line", []string{"own", "alpha"}, "files.\"list.txt\".y = { \"$mode\" = \"remove\" }\n", "a b z", ""},
		{"an override of a line of an unselected tool is known", []string{"own"}, "files.\"list.txt\".g = { \"$mode\" = \"remove\" }\n", "a b", ""},
		{"an override of an unknown line is unused", []string{"own"}, "files.\"list.txt\".q = { \"$mode\" = \"remove\" }\n", "", "the override q has no item of bootstrap"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			clone := contribClone(t, extra)
			code, errText, root := contribSync(t, clone, c.tools, c.over)
			if c.fail != "" {
				if code != fail.Usage || !strings.Contains(errText, c.fail) {
					t.Fatalf("code %d, want 2 with %q:\n%s", code, c.fail, errText)
				}
				return
			}
			if code != fail.OK {
				t.Fatalf("code %d:\n%s", code, errText)
			}
			if got := strings.Join(lines(t, filepath.Join(root, "list.txt")), " "); got != c.want {
				t.Errorf("list.txt %q, want %q", got, c.want)
			}
		})
	}
}

// TestContributedItemsNoOwner: a contributing tool whose owner is not
// selected adds nothing, and no error (D1).
func TestContributedItemsNoOwner(t *testing.T) {
	clone := contribClone(t, map[string]string{
		"zeta": "\n[[contributes_items.own.\"list.txt\".item]]\nvalue = \"z\"\n",
	})
	code, errText, root := contribSync(t, clone, []string{"zeta"}, "")
	if code != fail.OK {
		t.Fatalf("code %d:\n%s", code, errText)
	}
	if _, err := os.Stat(filepath.Join(root, "list.txt")); err == nil {
		t.Errorf("list.txt is there, want no file")
	}
}
