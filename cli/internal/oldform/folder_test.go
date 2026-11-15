package oldform

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/schema"
	"github.com/virajp/bootstrap/cli/internal/source"
)

func TestLanding(t *testing.T) {
	cases := []struct {
		table string
		want  string
	}{
		{`merge_model = "pr"`, schema.LandingPR},
		{"merge_model = \"pr\"\ndirect_commit = true", schema.LandingPR},
		{"merge_model = \"direct\"\ndirect_commit = true", schema.LandingDirectCommit},
		{"direct_commit = true", schema.LandingDirectCommit},
		{`merge_model = "direct"`, schema.LandingDirectMerge},
		{"direct_commit = false", schema.LandingDirectMerge},
		{"", schema.LandingDirectMerge},
	}
	for _, c := range cases {
		p := filepath.Join(t.TempDir(), "t.toml")
		write(t, p, c.table+"\n")
		raw, err := source.ReadTOML(p)
		if err != nil {
			t.Fatal(err)
		}
		if got := Landing(raw); got != c.want {
			t.Errorf("%q: landing = %s, want %s", c.table, got, c.want)
		}
	}
	if Landing(nil) != schema.LandingDirectMerge {
		t.Error("no table is not direct-merge")
	}
}

// read parses the TOML text <text>.
func read(t *testing.T, text []byte) *source.Obj {
	t.Helper()
	p := filepath.Join(t.TempDir(), "x.toml")
	write(t, p, string(text))
	raw, err := source.ReadTOML(p)
	if err != nil {
		t.Fatalf("%v:\n%s", err, text)
	}
	return raw
}

func TestSplitRepo(t *testing.T) {
	in := `# The header.
format = 2
repo_name = "x"
environments = [ "dev" ]

[tools.git]
# The scopes.
commit_scopes = [ "api" ]

[tools.git.branch.main]
merge_model = "pr"
direct_commit = false

[tools.git.branch.develop]
merge_model = "pr"
merge_from = [ "main" ]

[tools.mise]
worktree_files = [ ".env" ]

[tools.node]
version = "22"

[member.web]
folder = "web"
type = "folder"

[overrides.mise.settings]
jobs = 2
`
	over := filepath.Join(t.TempDir(), "o.toml")
	write(t, over, "[tools.git]\ncommit_scopes = [ \"cli\" ]\n\n[tools.git.branch.develop]\nmerge_model = \"direct\"\n\n[tools.node]\nversion = \"24\"\n")
	s, err := SplitRepo([]byte(in), "v.toml", over, func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	repo, tools := read(t, s.Repo), read(t, s.Tools)
	if err := schema.CheckRepo(repo, "repo.toml"); err != nil {
		t.Errorf("repo.toml: %v\n%s", err, s.Repo)
	}
	if err := schema.CheckToolsShape(tools, "tools.toml"); err != nil {
		t.Errorf("tools.toml: %v\n%s", err, s.Tools)
	}
	text := string(s.Repo)
	for _, w := range []string{
		"# The header.", "format = 3", `repo_name = "x"`, `core_branches = ["main", "develop"]`,
		"[branches.main]", `landing = "pr"`, "[branches.develop]", `merge_from = ["main"]`,
		"[commits]", `scopes = ["cli"]`, `style = "conventional"`, "[worktree]", `files = [".env"]`,
		"[members.web]", `folder = "web"`,
	} {
		if !strings.Contains(text, w) {
			t.Errorf("repo.toml lacks %q:\n%s", w, text)
		}
	}
	// The override of a key of repo.toml wins over the file (D18).
	if l := schema.Landing(repo, "develop"); l != schema.LandingDirectMerge {
		t.Errorf("develop landing = %s", l)
	}
	for _, no := range []string{"merge_model", "direct_commit", "[tools", "[member.", "[overrides"} {
		if strings.Contains(text, no) {
			t.Errorf("repo.toml has %q:\n%s", no, text)
		}
	}
	text = string(s.Tools)
	for _, w := range []string{"format = 3", "[tools.git]", "[tools.mise]", `version = "22"`, "[overrides.tools.node]", `version = "24"`, "[overrides.mise.settings]"} {
		if !strings.Contains(text, w) {
			t.Errorf("tools.toml lacks %q:\n%s", w, text)
		}
	}
	for _, no := range []string{"repo_name", "commit_scopes", "branch", "worktree_files", "member", "[overrides.tools.git"} {
		if strings.Contains(text, no) {
			t.Errorf("tools.toml has %q:\n%s", no, text)
		}
	}
}

func TestSplitRepoComments(t *testing.T) {
	in := "format = 2\nrepo_name = \"x\"\n\n[tools.git]\n# The scopes.\ncommit_scopes = [ \"api\" ]\n"
	s, err := SplitRepo([]byte(in), "v.toml", "", func(b string) bool { return b == "main" })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(s.Repo), "# The scopes.\nscopes") {
		t.Errorf("the comment of a moved key does not move:\n%s", s.Repo)
	}
	if !strings.Contains(string(s.Repo), `core_branches = ["main"]`) {
		t.Errorf("core branches:\n%s", s.Repo)
	}
}

func TestSplitRepoStops(t *testing.T) {
	cases := []struct {
		name, in string
		exists   func(string) bool
		msg      string
	}{
		{"not a core branch", "format = 2\n[tools.git.branch.release]\n", func(string) bool { return true }, "[tools.git.branch.release] in v.toml names a branch"},
		{"a missing branch", "format = 2\n[tools.git.branch.develop]\n", func(b string) bool { return b == "main" }, "[tools.git.branch.develop] in v.toml names a branch"},
		{"no main and no develop", "format = 2\n", func(string) bool { return false }, "no branch main or develop"},
	}
	for _, c := range cases {
		_, err := SplitRepo([]byte(c.in), "v.toml", "", c.exists)
		if fail.CodeOf(err) != fail.Usage || !strings.Contains(err.Error(), c.msg) {
			t.Errorf("%s: %v", c.name, err)
		}
	}
	_, err := SplitRepo([]byte("format = 2\n[tools.git.branch.x]\n"), "v.toml", "", func(string) bool { return true })
	if err == nil || !strings.Contains(err.Error(), "remove the table or create the branch") {
		t.Errorf("fix: %v", err)
	}
}
