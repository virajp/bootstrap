package target

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// TestUnblockMoving: a file of the repo at the place of a folder stops the
// dry run, unless init moves it before its sync (Moving).
func TestUnblockMoving(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "tasks", "test")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("own\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	e := &Env{Report: report.NewReport(&b, report.Options{Prefix: "bootstrap"}), Pwd: root, DryRun: true}
	if code := e.Unblock(file, root, nil, "x"); code != fail.Other {
		t.Errorf("without Moving: code %d", code)
	}
	e.Moving = map[string]bool{file: true}
	if code := e.Unblock(file, root, nil, "x"); code != fail.OK || len(e.Planned) != 0 {
		t.Errorf("with Moving: code %d, planned %v", code, e.Planned)
	}
}
