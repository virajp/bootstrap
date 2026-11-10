package misegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// TestUsage: the usage errors of the functions, with their text.
func TestUsage(t *testing.T) {
	expect := func(err error, msg string) {
		t.Helper()
		var list []string
		for _, e := range source.Messages(err) {
			list = append(list, e.Msg)
		}
		if fail.CodeOf(err) != fail.Usage || len(list) != 1 || list[0] != msg {
			t.Errorf("want code 2 and %q, got %d %q", msg, fail.CodeOf(err), list)
		}
	}
	bad := &Generator{Side: "x"}
	side := "sync-engine/mise: the side must be repo or macos, not 'x'"
	expect(bad.Overrides(nil, nil, false), side)
	_, err := bad.Render("t", t.TempDir())
	expect(err, side)
	_, err = bad.OwnConfig(t.TempDir(), nil)
	expect(err, side)
	_, err = NewEnvs("x", nil, nil, "", nil)
	expect(err, side)
	repo := &Generator{Side: source.Repo, Values: filepath.Join(t.TempDir(), "v.toml")}
	_, err = repo.OwnConfig(t.TempDir(), nil)
	expect(err, "sync-engine/mise: the own config is only for the side macos")
	e, err := NewEnvs(source.Macos, []string{"dev"}, nil, "", nil)
	if err != nil || len(e.Selected) != 0 || e.ToolWords() != " " || e.MiseWords() != " " {
		t.Errorf("macos envs: want an empty result, got %+v %v", e, err)
	}
	for name, want := range map[string]bool{"dev": true, "ci": true, "prod": false, "": false, "d ev": false, "dev ": false} {
		if EnvName(name) != want {
			t.Errorf("EnvName(%q) = %v", name, !want)
		}
	}
}

// TestMarkOverrides compares the awk step of `_se_mise_generate` with awk,
// for lines that the generator does not write.
func TestMarkOverrides(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("no awk")
	}
	const prog = `
    FNR == NR { if (index($0, "#@override ") == 1) mark[substr($0, 12)] = 1; next }
    index($0, "#@override ") == 1 { next }
    /^[ \t]*\[/ { head = $0; gsub(/^[ \t]*\[+|\]+[ \t]*$|"/, "", head) }
    !/^[ \t]*\[/ && / = / {
      key = $0; sub(/ = .*/, "", key); gsub(/^[ \t]+|"/, "", key)
      if ((head == "" ? key : head "." key) in mark) $0 = $0 "  " note
    }
    { print }
  `
	for _, in := range []string{
		"a = 1\n[ \"x\".y ]]  \n  \"k\" = 2\n[[arr]]\nb = 3 = 4\nc=5\n#@override a\n#@override  x.y.k\n#@override arr.b",
		"\t[t]\n\tk = 1\n[t.\"a.b\"]\nq = \"x = y\"\n#@override t.k\n#@override t.a.b.q\n",
		"x = 1",
		"[a] # c\nk = 1\n#@override a] # c.k\n",
	} {
		f := filepath.Join(t.TempDir(), "in")
		if err := os.WriteFile(f, []byte(in), 0o644); err != nil {
			t.Fatal(err)
		}
		want, err := exec.Command("awk", "-v", "note=# N", prog, f, f).Output()
		if err != nil {
			t.Fatal(err)
		}
		if got := string(markOverrides([]byte(in), "# N")); got != string(want) {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
