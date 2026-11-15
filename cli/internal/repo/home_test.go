package repo

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// writeFile writes <text> to <p>, with its folders.
func writeFile(t *testing.T, p, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// testState gives a state of the repo <root> with its values folder, its
// output in <buf>, and <env> as the environment.
func testState(t *testing.T, root string, env map[string]string, buf *bytes.Buffer) *state {
	t.Helper()
	r := report.NewReport(buf, report.Options{Prefix: Prefix})
	dir := values.FolderOf(root)
	s := &state{stderr: buf, r: r, e: &target.Env{Report: r, Pwd: root}, root: root, pwd: root,
		values: dir, show: ".config/bootstrap", showRepo: ".config/bootstrap/repo.toml", showTools: ".config/bootstrap/tools.toml",
		getenv: func(k string) string { return env[k] }}
	v, err := values.ReadRepo(dir, s.show)
	if err != nil {
		t.Fatal(err)
	}
	s.vals = v
	return s
}

const repoHead = "format = 3\nrepo_name = \"r\"\ncore_branches = [\"main\"]\n"

func TestMembersSibling(t *testing.T) {
	for _, present := range []bool{true, false} {
		top := t.TempDir()
		root := filepath.Join(top, "home")
		writeFile(t, values.RepoFile(values.FolderOf(root)), repoHead+"[members.api]\nfolder = \"api\"\ntype = \"folder\"\n[members.web]\nfolder = \"../web\"\ntype = \"sibling\"\nurl = \"https://example.com/web\"\n")
		writeFile(t, values.ToolsFile(values.FolderOf(root)), "format = 3\n")
		if err := os.MkdirAll(filepath.Join(root, "api"), 0o755); err != nil {
			t.Fatal(err)
		}
		if present {
			if err := os.MkdirAll(filepath.Join(top, "web"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		var buf bytes.Buffer
		s := testState(t, root, nil, &buf)
		got, code := s.members("test")
		want := []member{{"api", "api", "folder"}}
		if present {
			want = append(want, member{"web", "../web", "sibling"})
		}
		if code != fail.OK || !reflect.DeepEqual(got, want) {
			t.Errorf("present %v: code %d, members %v, want %v: %s", present, code, got, want, buf.String())
		}
		info := strings.Contains(buf.String(), "member web skipped: ../web is not there - clone https://example.com/web into it")
		if info == present {
			t.Errorf("present %v: output %q", present, buf.String())
		}
	}
}

func TestHomeLink(t *testing.T) {
	cases := []struct {
		name, home, member, want string
	}{
		{"sibling with url", "repo_url = \"https://example.com/home\"\n[members.web]\nfolder = \"../web\"\ntype = \"sibling\"\nurl = \"u\"\n", "web", `{"folder":"../home","type":"sibling","url":"https://example.com/home"}`},
		{"submodule without url", "[members.web]\ntype = \"submodule\"\n", "web", `{"folder":"../home","type":"submodule"}`},
		{"folder member", "[members.web]\ntype = \"folder\"\n", "web", ""},
		{"no table, not a git submodule", "", "web", ""},
	}
	for _, c := range cases {
		top := t.TempDir()
		home, root := filepath.Join(top, "home"), filepath.Join(top, "web")
		writeFile(t, values.RepoFile(values.FolderOf(home)), repoHead+c.home)
		writeFile(t, values.RepoFile(values.FolderOf(root)), "# Head.\n\n"+repoHead+"# The members.\n[members.x]\nfolder = \"x\"\ntype = \"folder\"\n")
		writeFile(t, values.ToolsFile(values.FolderOf(root)), "format = 3\n")
		var buf bytes.Buffer
		s := testState(t, root, map[string]string{"BOOTSTRAP_HOME": home, "BOOTSTRAP_MEMBER": c.member}, &buf)
		s.home = s.homeLink()
		got := ""
		if s.home != nil {
			got = strings.TrimSpace(source.JSON(s.home))
		}
		if got != c.want {
			t.Errorf("%s: [home] = %s, want %s: %s", c.name, got, c.want, buf.String())
			continue
		}
		if s.home == nil {
			continue
		}
		// The write, then no write when [home] is the same.
		if code := s.writeHome(); code != fail.OK {
			t.Fatalf("%s: write: %d %s", c.name, code, buf.String())
		}
		b, _ := os.ReadFile(values.RepoFile(s.values))
		for _, w := range []string{"# Head.", "# The members.", "[home]", `folder = "../home"`} {
			if !strings.Contains(string(b), w) {
				t.Errorf("%s: repo.toml has no %q:\n%s", c.name, w, b)
			}
		}
		if !strings.Contains(buf.String(), "[home] changed in") {
			t.Errorf("%s: output %q", c.name, buf.String())
		}
		buf.Reset()
		s2 := testState(t, root, map[string]string{"BOOTSTRAP_HOME": home, "BOOTSTRAP_MEMBER": c.member}, &buf)
		s2.home = s2.homeLink()
		s2.e.DryRun = true
		if code := s2.writeHome(); code != fail.OK || buf.Len() != 0 || len(s2.e.Planned) != 0 {
			t.Errorf("%s: second run: %d %q %v", c.name, code, buf.String(), s2.e.Planned)
		}
	}
	// Outside a member: no [home].
	var buf bytes.Buffer
	s := &state{getenv: func(string) string { return "" }, r: report.NewReport(&buf, report.Options{Prefix: Prefix})}
	if s.homeLink() != nil {
		t.Error("[home] outside a member")
	}
}

func TestAsSubmodule(t *testing.T) {
	obj := func(kind string) *source.Obj {
		o := source.NewObj()
		o.Set("folder", "..")
		o.Set("type", kind)
		return o
	}
	root := t.TempDir() // Not a git repo: the git test is false.
	for _, c := range []struct {
		name      string
		home, raw *source.Obj
		want      bool
	}{
		{"this run writes submodule", obj("submodule"), nil, true},
		{"this run writes sibling", obj("sibling"), nil, false},
		{"repo.toml says submodule", nil, obj("submodule"), true},
		{"repo.toml says sibling", nil, obj("sibling"), false},
		{"no [home]", nil, nil, false},
	} {
		raw := source.NewObj()
		if c.raw != nil {
			raw.Set("home", c.raw)
		}
		s := &state{root: root, home: c.home, vals: &values.Repo{Raw: raw}}
		if got := s.asSubmodule(); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}
