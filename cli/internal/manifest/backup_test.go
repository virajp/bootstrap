package manifest

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/target"
)

// TestBadRow: the @backup segment of a manifest row. A Mac row can start
// with it; a repo row cannot; no row has it as another segment.
func TestBadRow(t *testing.T) {
	for _, c := range []struct {
		mode LinkMode
		path string
		bad  bool
	}{
		{Macos, "@backup/x", false},
		{Macos, "@backup/.claude/skills/s/SKILL.md", false},
		{Macos, "@backup", true},
		{Macos, "x/@backup/y", true},
		{Macos, "@backup/.git/x", true},
		{Macos, ".config/x", false},
		{Repo, "@backup/x", true},
		{Repo, "x/@backup/y", false},
		{Repo, ".config/x", false},
	} {
		if got := BadRow(c.mode, c.path); got != c.bad {
			t.Errorf("BadRow(%s, %q) = %v, want %v", c.mode, c.path, got, c.bad)
		}
	}
}

// TestReadBackupRows: Read keeps a Mac @backup/ row, and skips the bad rows
// with a log line.
func TestReadBackupRows(t *testing.T) {
	file := filepath.Join(t.TempDir(), "m.tsv")
	content := FirstLine + "\nt\t@backup/a.md\nt\tx/@backup/y\nt\tb.txt\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		mode LinkMode
		want []Row
	}{
		{Macos, []Row{{"t", "@backup/a.md", ""}, {"t", "b.txt", ""}}},
		{Repo, []Row{{"t", "x/@backup/y", ""}, {"t", "b.txt", ""}}},
	} {
		var out bytes.Buffer
		e := &target.Env{Report: report.NewReport(&out, report.Options{Prefix: "bootstrap"})}
		rows, code := Read(e, c.mode, file, filepath.Dir(file))
		if code != 0 || !reflect.DeepEqual(rows, c.want) {
			t.Errorf("%s: rows = %v (code %d), want %v", c.mode, rows, code, c.want)
		}
		if !bytes.Contains(out.Bytes(), []byte("is not <tool><TAB><path> - skipped")) {
			t.Errorf("%s: no log line for the bad row: %q", c.mode, out.String())
		}
	}
}
