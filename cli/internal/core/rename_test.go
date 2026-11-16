package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

const (
	oldWt = "# My worktree config.\n[[pre-start]]\ntrust = \"mise trust\"\n\n[[pre-start]]\nsetup = \"mise run setup:worktree\"\n"
	newWt = "# My worktree config.\n[[pre-start]]\ntrust = \"mise trust\"\n\n[[pre-start]]\nsetup = \"mise run git:worktree:setup\"\n"
	// A repo.toml with a comment, a member with the old command and one
	// without it.
	memberRepo = "format = 3\nrepo_name = \"x\"\ncore_branches = [ \"main\" ]\n\n# The API.\n[members.api]\ntype = \"submodule\"\nworktree_install = \"mise run setup:worktree && pnpm install\"\n\n[members.web]\ntype = \"submodule\"\nworktree_install = \"pnpm install\"\n"
)

// renameRepo is a repo fixture for init with a values folder, and the files
// <files> (path below the root to text).
func renameRepo(t *testing.T, files map[string]string) Side {
	t.Helper()
	sd := fixture(t, source.Repo, map[string]string{"a": "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"},
		"[pack.base]\na = { required = true }\n", "format = 3\n\n[tools.a]\n")
	for rel, text := range files {
		p := filepath.Join(sd.Root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return sd
}

// renameChanges gives the changes of <p> that edit files.
func renameChanges(p *PlanResult) []oldform.Change {
	var out []oldform.Change
	for _, c := range p.Changes {
		if len(c.Edits) > 0 {
			out = append(out, c)
		}
	}
	return out
}

func readText(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPlanInitRenamesWtToml(t *testing.T) {
	// D20: init rewrites the old command in .config/wt.toml, and only it.
	sd := renameRepo(t, map[string]string{oldform.WtFile: oldWt})
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	rc := renameChanges(p)
	if len(rc) != 1 || len(rc[0].Edits) != 1 || rc[0].Note != oldform.RenameNote+": .config/wt.toml" {
		t.Fatalf("rename changes %+v", rc)
	}
	if c, err := RenameChange(sd); c != oldform.RenameNote || err != nil {
		t.Errorf("RenameChange = %q, %v", c, err)
	}
	if code, putBack, _, err := Apply(p, Options{}); code != fail.OK || putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	if got := readText(t, filepath.Join(sd.Root, oldform.WtFile)); got != newWt {
		t.Errorf("wt.toml = %q", got)
	}
	if c, err := RenameChange(sd); c != "" || err != nil {
		t.Errorf("RenameChange after init = %q, %v", c, err)
	}
}

func TestPlanInitRenamesWorktreeInstall(t *testing.T) {
	// D20: init rewrites the old command in a member's worktree_install of
	// repo.toml; the comments, the rest of the value and the other member
	// stay.
	sd := renameRepo(t, map[string]string{".config/bootstrap/repo.toml": memberRepo})
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	rc := renameChanges(p)
	if len(rc) != 1 || rc[0].Note != oldform.RenameNote+": .config/bootstrap/repo.toml members.api.worktree_install" {
		t.Fatalf("rename changes %+v", rc)
	}
	if code, putBack, _, err := Apply(p, Options{}); code != fail.OK || putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	got := readText(t, values.RepoFile(sd.Values))
	if !strings.Contains(got, `worktree_install = "mise run git:worktree:setup && pnpm install"`) ||
		!strings.Contains(got, `worktree_install = "pnpm install"`) || !strings.Contains(got, "# The API.") ||
		strings.Contains(got, "setup:worktree") {
		t.Errorf("repo.toml = %q", got)
	}
}

func TestPlanInitNoRename(t *testing.T) {
	// A wt.toml and a repo.toml with no old command do not change; a task
	// name that only starts with the old one is not the old command.
	wt := "[[pre-start]]\nsetup = \"mise run git:worktree:setup\"\nother = \"mise run setup:worktree-x\"\n"
	sd := renameRepo(t, map[string]string{oldform.WtFile: wt})
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if rc := renameChanges(p); len(rc) != 0 {
		t.Fatalf("rename changes %+v", rc)
	}
	if c, err := RenameChange(sd); c != "" || err != nil {
		t.Errorf("RenameChange = %q, %v", c, err)
	}
	if code, _, _, err := Apply(p, Options{}); code != fail.OK || err != nil {
		t.Fatalf("apply: %d %v", code, err)
	}
	if got := readText(t, filepath.Join(sd.Root, oldform.WtFile)); got != wt {
		t.Errorf("wt.toml = %q", got)
	}
}

func TestPlanInitRenameRestore(t *testing.T) {
	// A failed sync puts wt.toml and repo.toml back as they were.
	sd := renameRepo(t, map[string]string{oldform.WtFile: oldWt, ".config/bootstrap/repo.toml": memberRepo})
	(&fakeSync{code: fail.Usage}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if rc := renameChanges(p); len(rc) != 1 || len(rc[0].Edits) != 2 {
		t.Fatalf("rename changes %+v", rc)
	}
	if code, putBack, _, err := Apply(p, Options{}); code != fail.Usage || !putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	if got := readText(t, filepath.Join(sd.Root, oldform.WtFile)); got != oldWt {
		t.Errorf("wt.toml = %q", got)
	}
	if got := readText(t, values.RepoFile(sd.Values)); got != memberRepo {
		t.Errorf("repo.toml = %q", got)
	}
}

func TestApplyMigrateRenamesWorktreeInstall(t *testing.T) {
	// D20 with the migration of the old values file: the member moves to
	// repo.toml with the new command, and a failed sync puts wt.toml back.
	old := oldValues + "\n[member.api]\nworktree_install = \"mise run setup:worktree\"\n"
	sd := oldRepo(t, old, "", "")
	if c, err := RenameChange(sd); c != oldform.RenameNote || err != nil {
		t.Errorf("RenameChange = %q, %v", c, err)
	}
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if code, putBack, _, err := Apply(p, Options{}); code != fail.OK || putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	if got := readText(t, values.RepoFile(sd.Values)); !strings.Contains(got, `worktree_install = "mise run git:worktree:setup"`) {
		t.Errorf("repo.toml = %q", got)
	}

	sd = oldRepo(t, old, "", "")
	if err := os.WriteFile(filepath.Join(sd.Root, oldform.WtFile), []byte(oldWt), 0o644); err != nil {
		t.Fatal(err)
	}
	(&fakeSync{code: fail.Usage}).install(t, sd)
	if p, err = Plan(sd, List{Init: true}, Options{}); err != nil {
		t.Fatal(problems(err))
	}
	if code, putBack, _, err := Apply(p, Options{}); code != fail.Usage || !putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	if got := readText(t, filepath.Join(sd.Root, oldform.WtFile)); got != oldWt {
		t.Errorf("wt.toml = %q", got)
	}
	if Exists(sd.Values) || !Exists(OldValuesOf(sd.Root)) {
		t.Error("the migration was not put back")
	}
}
