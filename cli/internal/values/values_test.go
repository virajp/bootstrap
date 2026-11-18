package values

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

// The tests of the merge are in internal/selection. These tests cover the
// small rules of this package alone.

// folder writes a values folder with the files <repo> and <tools>; "-"
// leaves a file out.
func folder(t *testing.T, repo, tools string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "bootstrap")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{RepoName: repo, ToolsName: tools} {
		if text == "-" {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const repoHead = "format = 3\nrepo_name = \"r\"\ncore_branches = [ \"main\", \"develop\" ]\n"

func TestReadRepoSelection(t *testing.T) {
	text := "format = 3\n[tools.b]\nenv = [ \"dev\", \"ci\" ]\nk = 1\n\n[tools.a]\n\n[overrides.tools.b]\nk = 2\n\n[overrides.mise.x]\nv = 1\n"
	v, err := ReadRepo(folder(t, repoHead, text), "bootstrap")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, it := range v.Sel {
		got = append(got, it.Kind+" "+it.Name+" "+it.Envs)
	}
	want := []string{"tools a -", "tools b dev,ci"}
	if len(got) != len(want) {
		t.Fatalf("Sel = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Sel[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if j := source.JSON(v.User); j != `{"a":{},"b":{"k":1}}` {
		t.Errorf("User = %s", j)
	}
	if j := source.JSON(v.Over); j != `{"b":{"k":2}}` {
		t.Errorf("Over = %s", j)
	}
	if got := strings.Join(v.Raw.Keys(), " "); got != "core_branches format overrides repo_name tools" {
		t.Errorf("Raw keys = %s", got)
	}
	if j := source.JSON(v.Raw.Get("overrides")); j != `{"mise":{"x":{"v":1}}}` {
		t.Errorf("Raw overrides = %s", j)
	}
}

// TestReadRepoCompat: the keys of repo.toml that a tool read in its table
// before the values folder reach its values under their old names.
func TestReadRepoCompat(t *testing.T) {
	repo := repoHead + "[commits]\nscopes = [ \"api\" ]\n[worktree]\nfiles = [ \"a.toml\" ]\n" +
		"[branches.main]\nlanding = \"pr\"\nmerge_from = [ \"develop\" ]\n[branches.develop]\nlanding = \"direct-commit\"\n"
	v, err := ReadRepo(folder(t, repo, "format = 3\n[tools.git]\n"), "bootstrap")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"git":{"branch":{"develop":{"direct_commit":true,"merge_model":"direct"},"main":{"merge_from":["develop"],"merge_model":"pr"}},"commit_scopes":["api"]},"mise":{"worktree_files":["a.toml"]}}`
	if j := source.JSON(v.User); j != want {
		t.Errorf("User = %s\nwant   %s", j, want)
	}
}

// TestReadFolder: the rules of the files of the values folder.
func TestReadFolder(t *testing.T) {
	cases := []struct{ name, repo, tools, want string }{
		{"valid", repoHead, "format = 3\n", ""},
		{"no folder files", "-", "-", "bootstrap has no repo.toml - run bootstrap migrate\nbootstrap has no tools.toml - run bootstrap migrate"},
		{"no tools.toml", repoHead, "-", "bootstrap has no tools.toml - run bootstrap migrate"},
		{"no format", "repo_name = \"r\"\ncore_branches = [ \"main\" ]\n", "format = 3\n", "bootstrap/repo.toml without format = 3 is not valid - run bootstrap migrate"},
		{"old format", repoHead, "format = 2\n", "format = 2 in bootstrap/tools.toml is not valid - the format of a values file is 3"},
		{"unknown repo key", repoHead + "colour = 1\n", "format = 3\n", "colour in bootstrap/repo.toml is not a key of repo.toml"},
		{"unknown tools key", repoHead, "format = 3\n[packs.x]\n", "packs is not a key of bootstrap/tools.toml - the top-level keys are format, [tools.<name>] and [overrides.*]"},
		{"moved key", repoHead, "format = 3\n[overrides.tools.git]\ncommit_scopes = []\n", "[overrides.tools.git].commit_scopes in bootstrap/tools.toml is a key of repo.toml now - write it as commits.scopes in repo.toml"},
		{"missing repo_name", "format = 3\ncore_branches = [ \"main\" ]\n", "format = 3\n", "bootstrap/repo.toml has no repo_name - it is a required key of repo.toml"},
	}
	for _, c := range cases {
		_, _, err := ReadFolder(folder(t, c.repo, c.tools), "bootstrap")
		var got []string
		for _, m := range source.Messages(err) {
			got = append(got, m.Error())
		}
		if strings.Join(got, "\n") != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestMergeEmpty(t *testing.T) {
	r, err := Merge(Input{})
	if err != nil || len(r.Tools) != 0 || r.Lines() != "" {
		t.Errorf("Merge(empty) = %v, %v", r, err)
	}
}
