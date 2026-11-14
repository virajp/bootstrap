package repo

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
	"github.com/virajp/bootstrap/cli/internal/values"
)

func TestToolValuesGroup(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "repo", "tools", "t", "tools.toml")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	const toml = "default_envs = [\"dev\"]\ndepends = []\nroot_only = false\nsummary = \"A tool\"\ncategory = \"core\"\n[questions.a]\nquestion = \"A?\"\ngroup = \"g\"\n[questions.b]\nquestion = \"B?\"\ngroup = \"g\"\n[questions.d]\nquestion = \"D?\"\n"
	if err := os.WriteFile(p, []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "repo", "categories.toml"), []byte("[category.core]\ntitle = \"Core\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := source.Source{Root: root, Side: source.Repo}
	idx, err := src.Open()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		user map[string]any
		want string
	}{
		{map[string]any{"a": "", "d": ""}, "[ERROR] the group t.g has no answer in .config/bootstrap.toml\n  -> Run bootstrap init\n"},
		{map[string]any{"a": "x"}, "[ERROR] the question t.d has no answer in .config/bootstrap.toml\n  -> Run bootstrap init\n"},
		{map[string]any{"b": "x", "d": ""}, ""},
	} {
		tt := source.NewObj()
		for k, v := range c.user {
			tt.Set(k, v)
		}
		user := source.NewObj()
		user.Set("t", tt)
		var buf bytes.Buffer
		r := report.NewReport(&buf, report.Options{Prefix: Prefix})
		s := &state{stderr: &buf, r: r, e: &target.Env{Report: r, Pwd: root}, root: root, dir: root, src: src, idx: idx,
			values: filepath.Join(root, ".config", "bootstrap.toml"), show: ".config/bootstrap.toml",
			vals: &values.Repo{Raw: source.NewObj(), User: user}, membersRead: true, res: &selection.Resolved{Tools: []string{"t"}}}
		code := s.toolValues()
		if buf.String() != c.want || (code == fail.OK) != (c.want == "") {
			t.Errorf("%v: code %d, output %q, want %q", c.user, code, buf.String(), c.want)
		}
	}
}
