package oldform

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

const oldCopy = "#!/bin/bash\necho old pack task\n"

func sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return "sha256:" + hex.EncodeToString(h[:])
}

// copyIndex is an index with the tool "t1", whose part grp/40-t1 replaces
// oldCopy, and the tool "t2" with no replaces.
func copyIndex() *source.Index {
	return &source.Index{Manifests: map[string]*source.Manifest{
		"t1": {Tool: "t1", Files: []source.FileEntry{
			{Target: TasksDir + "/grp/40-t1", Task: true, Replaces: []string{sum("other"), sum(oldCopy)}},
			{Target: TasksDir + "/grp/10-t1", Task: true},
		}},
		"t2": {Tool: "t2", Files: []source.FileEntry{{Target: TasksDir + "/grp/40-t2", Task: true}}},
	}}
}

func TestPartsOfSelectedTools(t *testing.T) {
	got := Parts(copyIndex(), []string{"t2", "t1"})
	want := []Part{{Tool: "t1", Task: "grp/40-t1", Replaces: []string{sum("other"), sum(oldCopy)}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parts %v, want %v", got, want)
	}
	if got := Parts(copyIndex(), []string{"t2"}); got != nil {
		t.Fatalf("parts of t2: %v", got)
	}
}

func TestKnownCopyOldPath(t *testing.T) {
	// D5: a known copy at the old path of the group.
	root := t.TempDir()
	p := taskFile(t, root, "grp", oldCopy, 0o755)
	got := KnownCopies(root, Parts(copyIndex(), []string{"t1"}))
	want := []Copy{{Tool: "t1", Task: "grp/40-t1", Path: p}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("copies %v, want %v", got, want)
	}
}

func TestKnownCopyLocalPart(t *testing.T) {
	// D5: a known copy at <group>/50-local.
	root := t.TempDir()
	p := taskFile(t, root, "grp/50-local", oldCopy, 0o755)
	got := KnownCopies(root, Parts(copyIndex(), []string{"t1"}))
	if len(got) != 1 || got[0].Path != p {
		t.Fatalf("copies %v", got)
	}
	err := CopyProblems(root, Parts(copyIndex(), []string{"t1"}), show)
	if fail.CodeOf(err) != fail.Usage || !strings.Contains(err.Error(), p+" is an old copy of the task grp/40-t1 of the tool t1 - "+Hint) {
		t.Fatalf("problems: %v", err)
	}
	Migrating = true
	defer func() { Migrating = false }()
	if err := CopyProblems(root, Parts(copyIndex(), []string{"t1"}), show); err != nil {
		t.Fatalf("problems while init plans: %v", err)
	}
}

func TestEditedCopyStays(t *testing.T) {
	// D16: an edited copy is the repo's own, with no problem.
	root := t.TempDir()
	taskFile(t, root, "grp/50-local", oldCopy+"# edited\n", 0o755)
	parts := Parts(copyIndex(), []string{"t1"})
	if got := KnownCopies(root, parts); got != nil {
		t.Fatalf("copies %v", got)
	}
	if err := CopyProblems(root, parts, show); err != nil {
		t.Fatalf("problems: %v", err)
	}
}

func TestCopyOfUnselectedTool(t *testing.T) {
	// A file is only removed when its tool is selected.
	root := t.TempDir()
	taskFile(t, root, "grp", oldCopy, 0o755)
	parts := Parts(copyIndex(), []string{"t2"})
	if got := KnownCopies(root, parts); got != nil {
		t.Fatalf("copies %v", got)
	}
	if err := CopyProblems(root, parts, show); err != nil {
		t.Fatalf("problems: %v", err)
	}
}

func TestCopyLinkStays(t *testing.T) {
	// The engine follows no symlink.
	root := t.TempDir()
	target := taskFile(t, root, "real", oldCopy, 0o755)
	if err := os.MkdirAll(root+"/"+TasksDir+"/grp", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, root+"/"+TasksDir+"/grp/50-local"); err != nil {
		t.Fatal(err)
	}
	if got := KnownCopies(root, Parts(copyIndex(), []string{"t1"})); got != nil {
		t.Fatalf("copies %v", got)
	}
}

func TestDropsApplyAndUndo(t *testing.T) {
	// init removes a known copy, and a failed sync writes it back with its
	// mode, also where a folder replaced it.
	root := t.TempDir()
	p := taskFile(t, root, "grp", oldCopy, 0o750)
	drops, err := DropsOf(KnownCopies(root, Parts(copyIndex(), []string{"t1"})))
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply([]Change{{Path: root + "/" + TasksDir, Drops: drops}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(p); !os.IsNotExist(err) {
		t.Fatalf("copy not removed: %v", err)
	}
	if got := Paths([]Change{{Drops: drops}}); !reflect.DeepEqual(got, []string{p}) {
		t.Fatalf("paths %v", got)
	}
	taskFile(t, root, "grp/40-t1", "part", 0o755)
	if err := UndoDrops(drops); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(p)
	b, _ := os.ReadFile(p)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o750 || string(b) != oldCopy {
		t.Fatalf("written back: %v %v %q", info, err, b)
	}
}
