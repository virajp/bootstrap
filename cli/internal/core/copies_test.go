package core

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// TestKnownCopiesBeforeMove: init removes a known old copy at the old path
// of its group as one change, before the move of the own test tasks, which
// then does not move it (D4, D5); the own task of another group still moves.
func TestKnownCopiesBeforeMove(t *testing.T) {
	root := t.TempDir()
	old := "#!/bin/bash\necho old\n"
	for rel, text := range map[string]string{"test/golden": old, "test/e2e": "#!/bin/bash\necho own\n"} {
		p := filepath.Join(root, oldform.TasksDir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	h := sha256.Sum256([]byte(old))
	idx := &source.Index{Manifests: map[string]*source.Manifest{"t1": {Tool: "t1", Files: []source.FileEntry{
		{Target: oldform.TasksDir + "/test/golden/40-t1", Task: true, Replaces: []string{"sha256:" + hex.EncodeToString(h[:])}},
	}}}}
	p := &PlanResult{Side: Side{Root: root, Pwd: root}}
	p.Changes = []oldform.Change{{Path: root + "/first", Note: "first"}}
	if err := p.taskMoves(); err != nil {
		t.Fatal(err)
	}
	p.Changes = append(p.Changes, oldform.Change{Path: root + "/last", Note: "last"})
	if err := p.knownCopies(&Selected{Idx: idx, Tools: []string{"t1"}}); err != nil {
		t.Fatal(err)
	}
	if len(p.Changes) != 4 {
		t.Fatalf("changes %+v", p.Changes)
	}
	drop, move := p.Changes[1], p.Changes[2]
	golden := filepath.Join(root, oldform.TasksDir, "test/golden")
	if len(drop.Drops) != 1 || drop.Drops[0].Path != golden || drop.Note != "remove old copies of the t1 test parts: "+filepath.Join(oldform.TasksDir, "test/golden") {
		t.Errorf("drop change %+v", drop)
	}
	if len(move.Moves) != 1 || !strings.HasSuffix(move.Moves[0].From, "/test/e2e") || strings.Contains(move.Note, "golden") {
		t.Errorf("move change %+v", move)
	}
	if p.Changes[0].Note != "first" || p.Changes[3].Note != "last" {
		t.Errorf("order %+v", p.Changes)
	}
	if !p.moving()[golden] {
		t.Errorf("moving %v", p.moving())
	}
	// A tool that is not selected keeps its copy.
	q := &PlanResult{Side: Side{Root: root, Pwd: root}}
	if err := q.knownCopies(&Selected{Idx: idx, Tools: nil}); err != nil || len(q.Changes) != 0 {
		t.Errorf("unselected: %v %+v", err, q.Changes)
	}
}
