package target

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

func TestAddTables(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "v.toml")
	text := "# Head.\nformat = 2\n\n[tools.a]\n# own\nx = 1\n"
	if err := os.WriteFile(file, []byte(text), 0o640); err != nil {
		t.Fatal(err)
	}
	run := func(dry, listOnly bool, names ...string) (*Env, string) {
		var b bytes.Buffer
		e := &Env{Report: report.NewReport(&b, report.Options{Prefix: "bootstrap"}), Pwd: dir, DryRun: dry}
		if code := e.AddTables(file, names, "", listOnly); code != fail.OK {
			t.Fatalf("code %d: %s", code, b.String())
		}
		return e, b.String()
	}
	// A dry run plans the file and writes nothing; check lists only.
	e, out := run(true, false, "m", "g")
	if b, _ := os.ReadFile(file); string(b) != text || len(e.Planned) != 1 || !strings.Contains(out, "[tools.m] would be added to v.toml") {
		t.Errorf("dry run: %q, planned %v, out %q", b, e.Planned, out)
	}
	e, out = run(false, true, "m")
	if b, _ := os.ReadFile(file); string(b) != text || len(e.Planned) != 0 || !strings.Contains(out, "[tools.m] would be added to v.toml") {
		t.Errorf("check: %q, planned %v, out %q", b, e.Planned, out)
	}
	_, out = run(false, false, "m", "g")
	b, _ := os.ReadFile(file)
	for _, want := range []string{"# Head.", "# own", "[tools.m]", "[tools.g]", "x = 1"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("%q is missing:\n%s", want, b)
		}
	}
	if !strings.Contains(out, "[tools.g] added to v.toml") {
		t.Errorf("out = %q", out)
	}
	if st, _ := os.Stat(file); st.Mode().Perm() != 0o640 {
		t.Errorf("mode = %v", st.Mode().Perm())
	}
}

func TestAddTablesNewFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "macos.toml")
	var b bytes.Buffer
	e := &Env{Report: report.NewReport(&b, report.Options{Prefix: "bootstrap"}), Pwd: dir}
	if code := e.AddTables(file, []string{"m"}, "# New.\n\nformat = 2\n", false); code != fail.OK {
		t.Fatalf("code %d: %s", code, b.String())
	}
	got, _ := os.ReadFile(file)
	// The blank line after the header stays, so bootstrap_commit goes after
	// the header and the header stays on line 1.
	if !strings.HasPrefix(string(got), "# New.\n\nformat = 2\n") || !strings.Contains(string(got), "[tools.m]") {
		t.Errorf("new file = %q", got)
	}
	if code := e.RecordCommit(file, "bootstrap_commit", "abc", "", ""); code != fail.OK {
		t.Fatalf("record: code %d: %s", code, b.String())
	}
	got, _ = os.ReadFile(file)
	if !strings.HasPrefix(string(got), "# New.\n\nbootstrap_commit = \"abc\"\n") {
		t.Errorf("after record = %q", got)
	}
	// Without a head, a missing file is an error.
	if code := e.AddTables(filepath.Join(dir, "nope.toml"), []string{"m"}, "", false); code != fail.Other {
		t.Errorf("code = %d", code)
	}
}
