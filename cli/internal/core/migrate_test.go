package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// oldRepo is a repo fixture with the old values file <old>, the overrides
// file <over> and the manifest <manifest> (each skipped when empty), and a
// commit on main and on each branch of <branches>.
func oldRepo(t *testing.T, old, over, manifest string, branches ...string) Side {
	t.Helper()
	sd := fixture(t, source.Repo, map[string]string{"a": "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"},
		"[pack.base]\na = { required = true }\n", "")
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", sd.Root, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Env = GitEnv()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	git("commit", "-q", "--allow-empty", "-m", "init")
	for _, b := range branches {
		git("branch", b)
	}
	for p, text := range map[string]string{oldform.OldValues: old, oldform.OldOverrides: over, oldform.OldManifest: manifest} {
		if text == "" {
			continue
		}
		path := filepath.Join(sd.Root, p)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return sd
}

const oldValues = `# The header.
format = 2
repo_name = "x"
environments = [ "dev" ]

[tools.a]
`

func TestPlanMigrate(t *testing.T) {
	sd := oldRepo(t, oldValues, "[tools.a]\nkey = 1\n", "row\n", "develop")
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if !p.Migrated || p.Existed {
		t.Errorf("migrated = %v, existed = %v", p.Migrated, p.Existed)
	}
	root := sd.Root
	want := []string{
		root + "/" + oldform.OldValues, root + "/" + oldform.OldOverrides, root + "/" + oldform.OldManifest,
		values.ManifestFile(sd.Values), values.RepoFile(sd.Values), values.ToolsFile(sd.Values),
	}
	if got := oldform.Paths(p.Changes); !reflect.DeepEqual(got, sortStrings(want)) {
		t.Errorf("changes = %v, want %v", got, want)
	}
	for _, ch := range p.Changes {
		if strings.HasPrefix(ch.Path, root+"/.config/bootstrap.") || ch.Path == root+"/"+oldform.OldManifest {
			if !ch.Remove {
				t.Errorf("%s is not removed", ch.Path)
			}
		}
	}
	next := string(p.Next)
	for _, w := range []string{ToolsHeader + "\n", "format = 3", "[tools.a]", "[overrides.tools.a]", "key = 1"} {
		if !strings.Contains(next, w) {
			t.Errorf("tools.toml lacks %q:\n%s", w, next)
		}
	}
	if strings.Contains(next, "repo_name") {
		t.Errorf("tools.toml has repo_name:\n%s", next)
	}
	repo := string(p.RepoNew)
	for _, w := range []string{"# The header.", "format = 3", `repo_name = "x"`, `core_branches = ["main", "develop"]`, `style = "conventional"`} {
		if !strings.Contains(repo, w) {
			t.Errorf("repo.toml lacks %q:\n%s", w, repo)
		}
	}
	// The migration adds no member of base: the old file has its selection.
	if c, _ := InitChange(sd); c != "migrate to .config/bootstrap/" {
		t.Errorf("InitChange = %q", c)
	}
}

func TestApplyMigrate(t *testing.T) {
	sd := oldRepo(t, oldValues, "[tools.a]\nkey = 1\n", "row\n")
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if code, putBack, _, err := Apply(p, Options{}); code != fail.OK || putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	for _, f := range []string{oldform.OldValues, oldform.OldOverrides, oldform.OldManifest} {
		if Exists(filepath.Join(sd.Root, f)) {
			t.Errorf("%s is still there", f)
		}
	}
	if b, _ := os.ReadFile(values.ManifestFile(sd.Values)); string(b) != "row\n" {
		t.Errorf("manifest = %q", b)
	}
	entries, _ := os.ReadDir(filepath.Dir(sd.Values))
	for _, e := range entries {
		if strings.Contains(e.Name(), ".new.") {
			t.Errorf("the temp folder %s is still there", e.Name())
		}
	}
	// An already migrated repo is a no-op for the migration.
	if c, _ := InitChange(sd); c != "" {
		t.Errorf("InitChange after the migration = %q", c)
	}
	p, err = Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if p.Migrated || len(p.Changes) != 0 {
		t.Errorf("second init: migrated = %v, changes = %v", p.Migrated, oldform.Paths(p.Changes))
	}
}

func TestApplyMigrateRestore(t *testing.T) {
	sd := oldRepo(t, oldValues, "[tools.a]\nkey = 1\n", "row\n")
	f := &fakeSync{code: fail.Usage}
	f.install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	code, putBack, _, err := Apply(p, Options{})
	if code != fail.Usage || !putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	if Exists(sd.Values) {
		t.Error("the values folder is still there")
	}
	for f, want := range map[string]string{oldform.OldValues: oldValues, oldform.OldOverrides: "[tools.a]\nkey = 1\n", oldform.OldManifest: "row\n"} {
		if b, _ := os.ReadFile(filepath.Join(sd.Root, f)); string(b) != want {
			t.Errorf("%s = %q, want %q", f, b, want)
		}
	}
	if p.RestoredFile() != filepath.Join(sd.Root, oldform.OldValues) {
		t.Errorf("RestoredFile = %s", p.RestoredFile())
	}
}

func TestPlanMigrateStops(t *testing.T) {
	// D47: a branch table of a branch that does not exist.
	sd := oldRepo(t, oldValues+"\n[tools.git.branch.develop]\nmerge_model = \"pr\"\n", "", "")
	(&fakeSync{}).install(t, sd)
	_, err := Plan(sd, List{Init: true}, Options{})
	if fail.CodeOf(err) != fail.Usage || !strings.Contains(err.Error(), "[tools.git.branch.develop]") || !strings.Contains(err.Error(), "remove the table or create the branch") {
		t.Errorf("err = %v", err)
	}
	// D16: no main and no develop.
	sd = oldRepo(t, oldValues, "", "")
	if out, err := exec.Command("git", "-C", sd.Root, "branch", "-m", "main", "trunk").CombinedOutput(); err != nil {
		t.Fatalf("git: %s", out)
	}
	_, err = Plan(sd, List{Init: true}, Options{})
	if fail.CodeOf(err) != fail.Usage || !strings.Contains(err.Error(), "no branch main or develop") {
		t.Errorf("err = %v", err)
	}
}

func TestCheckLayout(t *testing.T) {
	sd := oldRepo(t, oldValues, "", "")
	// D28: the old file alone stops each command but init.
	if err := CheckLayout(sd, false); fail.CodeOf(err) != fail.Usage || !strings.Contains(problemFix(err), "run bootstrap migrate") {
		t.Errorf("old form: %v", err)
	}
	if err := CheckLayout(sd, true); err != nil {
		t.Errorf("old form, init: %v", err)
	}
	// D29: both stop each command, init too.
	if err := os.MkdirAll(sd.Values, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, initCmd := range []bool{false, true} {
		if err := CheckLayout(sd, initCmd); fail.CodeOf(err) != fail.Usage || problemFix(err) != oldform.FolderHint {
			t.Errorf("both, init %v: %v", initCmd, err)
		}
	}
	// D88: a folder without repo.toml stops each command but init.
	if err := os.Remove(filepath.Join(sd.Root, oldform.OldValues)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(values.ToolsFile(sd.Values), []byte("format = 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckLayout(sd, false); fail.CodeOf(err) != fail.Usage || !strings.Contains(err.Error(), "has no repo.toml") {
		t.Errorf("no repo.toml: %v", err)
	}
	if err := CheckLayout(sd, true); err != nil {
		t.Errorf("no repo.toml, init: %v", err)
	}
	// The Mac side has no folder.
	if err := CheckLayout(Side{Mac: true}, false); err != nil {
		t.Errorf("mac: %v", err)
	}
}

// problemFix gives the fix of the problem <err>.
func problemFix(err error) string {
	if p, ok := err.(*Problem); ok {
		return p.Fix
	}
	return ""
}

func sortStrings(l []string) []string {
	out := append([]string(nil), l...)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// TestMigrateStalePlanned runs the real sync of the repo side: the dry run of
// a migration reads the old manifest from the temp values folder, so it
// lists the stale removals that the sync then makes (D94).
func TestMigrateStalePlanned(t *testing.T) {
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise is not available")
	}
	sd := oldRepo(t, oldValues, "", "a\tstale.txt\n")
	stale := filepath.Join(sd.Root, "stale.txt")
	text := "# Synced from bootstrap: a. Do not edit this file.\n"
	if err := os.WriteFile(stale, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	// The sync takes a clone with a commit.
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "init"}} {
		cmd := exec.Command("git", append([]string{"-C", sd.Clone}, args...)...)
		cmd.Env = GitEnv()
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, b)
		}
	}
	t.Setenv("MISE_PROJECT_ROOT", sd.Root)
	t.Setenv("PWD", sd.Root)
	t.Setenv("BOOTSTRAP_CLONE", sd.Clone)
	var out strings.Builder
	p, err := Plan(sd, List{Init: true}, Options{Stdout: &out, Stderr: &out})
	if err != nil {
		t.Fatalf("%v %v\n%s", err, problems(err), out.String())
	}
	var removed []string
	for _, f := range p.Files {
		if f.Kind == report.Removed {
			removed = append(removed, f.Path)
		}
	}
	if !reflect.DeepEqual(removed, []string{stale}) {
		t.Errorf("planned removals = %v, want %v", removed, []string{stale})
	}
	if code, putBack, _, err := Apply(p, Options{}); code != fail.OK || putBack || err != nil {
		t.Fatalf("apply: %d %v %v", code, putBack, err)
	}
	if Exists(stale) {
		t.Errorf("the sync did not remove %s", stale)
	}
	// Without a migration, the plan reads the manifest of the values folder.
	if err := os.WriteFile(stale, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(values.ManifestFile(sd.Values), []byte("a\tstale.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if p, err = Plan(sd, List{Init: true}, Options{Stdout: &out, Stderr: &out}); err != nil {
		t.Fatalf("%v %v\n%s", err, problems(err), out.String())
	}
	removed = nil
	for _, f := range p.Files {
		if f.Kind == report.Removed {
			removed = append(removed, f.Path)
		}
	}
	if p.Migrated || !reflect.DeepEqual(removed, []string{stale}) {
		t.Errorf("second plan: migrated = %v, removals = %v", p.Migrated, removed)
	}
	// A manifest that is a link stops the dry run, as it stops the sync.
	real := filepath.Join(sd.Root, "real.tsv")
	if err := os.WriteFile(real, []byte("a\tstale.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(values.ManifestFile(sd.Values)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, values.ManifestFile(sd.Values)); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if _, err = Plan(sd, List{Init: true}, Options{Stdout: &out, Stderr: &out}); err == nil {
		t.Errorf("a linked manifest: the plan did not stop\n%s", out.String())
	} else if !strings.Contains(out.String(), "the manifest is a link") {
		t.Errorf("a linked manifest: %v %v\n%s", err, problems(err), out.String())
	}
}
