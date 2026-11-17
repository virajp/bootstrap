package oldform

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/manifest"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// CopiesNote is the label of the change of init that removes the known old
// copies of the parts of the tool <tool> (D4, D5).
func CopiesNote(tool string) string { return "remove old copies of the " + tool + " test parts" }

// Part is one task of a tool with `replaces`: the path below TasksDir and
// the hashes of its known old copies.
type Part struct {
	Tool, Task string
	Replaces   []string
}

// Copy is one known old copy of a part: a file of the repository whose
// sha256 is in `replaces` of the part.
type Copy struct {
	Tool, Task string
	Path       string // Absolute.
}

// Drop is one file that a change removes, with its content and its mode, so
// a failed sync can write it back.
type Drop struct {
	Path string // Absolute.
	Old  []byte
	Mode os.FileMode
}

// Parts gives each task with `replaces` of the tools <tools> of the index
// <idx>, in the order of <tools>, then of the targets.
func Parts(idx *source.Index, tools []string) []Part {
	var out []Part
	if idx == nil {
		return nil
	}
	for _, tool := range tools {
		m := idx.Manifests[tool]
		if m == nil {
			continue
		}
		for _, fe := range m.Files {
			if !fe.Task || len(fe.Replaces) == 0 {
				continue
			}
			task := strings.TrimPrefix(fe.Target, TasksDir+"/")
			out = append(out, Part{Tool: tool, Task: task, Replaces: fe.Replaces})
		}
	}
	return out
}

// copyPlaces gives the paths below <root> (relative) where the known old
// copies of the part <task> are: the old path of its group and
// <group>/50-local (D5).
func copyPlaces(task string) []string {
	group := path.Dir(task)
	if group == "." {
		return nil
	}
	return []string{TasksDir + "/" + group, TasksDir + "/" + group + "/" + LocalPart}
}

// KnownCopies gives each known old copy of the parts <parts> in the repo
// <root>: a regular file at a place of copyPlaces whose sha256 is in
// `replaces` of its part. Another file there is the repo's own (D16). The
// engine follows no symlink: a path that is a link, or is below a linked
// folder, is not a copy.
func KnownCopies(root string, parts []Part) []Copy {
	root = strings.TrimSuffix(root, "/")
	var out []Copy
	seen := map[string]bool{}
	for _, p := range parts {
		for _, rel := range copyPlaces(p.Task) {
			abs := root + "/" + rel
			if seen[abs] || manifest.Linked(root, rel) {
				continue
			}
			info, err := os.Lstat(abs)
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			b, err := os.ReadFile(abs)
			if err != nil {
				continue
			}
			sum := sha256.Sum256(b)
			if !slices.Contains(p.Replaces, "sha256:"+hex.EncodeToString(sum[:])) {
				continue
			}
			seen[abs] = true
			out = append(out, Copy{Tool: p.Tool, Task: p.Task, Path: abs})
		}
	}
	return out
}

// CopyProblems gives a problem with code 2 for each known old copy of the
// parts <parts> in the repo <root>, with the fix of init (D5). While init
// plans (Migrating), it gives none: init removes them. <show> gives a path
// in a message.
func CopyProblems(root string, parts []Part, show func(string) string) error {
	if Migrating {
		return nil
	}
	var probs []*fail.Error
	for _, c := range KnownCopies(root, parts) {
		probs = append(probs, usage("%s is an old copy of the task %s of the tool %s - %s", show(c.Path), c.Task, c.Tool, Hint))
	}
	if len(probs) == 0 {
		return nil
	}
	return source.Problems(probs)
}

// DropsOf gives the drops of the copies <copies>, with the content and the
// mode of each file.
func DropsOf(copies []Copy) ([]Drop, error) {
	var out []Drop
	for _, c := range copies {
		info, err := os.Lstat(c.Path)
		if err != nil {
			return nil, fail.New(fail.Other, "cannot read %s", c.Path)
		}
		b, err := os.ReadFile(c.Path)
		if err != nil {
			return nil, fail.New(fail.Other, "cannot read %s", c.Path)
		}
		out = append(out, Drop{Path: c.Path, Old: b, Mode: info.Mode().Perm()})
	}
	return out, nil
}

// DropDetail gives the detail of a change with drops: each file.
func DropDetail(drops []Drop, show func(string) string) string {
	parts := make([]string, len(drops))
	for i, d := range drops {
		parts[i] = show(d.Path)
	}
	return strings.Join(parts, ", ")
}

// ApplyDrops removes each file of <drops>. A failure writes back the files
// of <drops> that it removed.
func ApplyDrops(drops []Drop) error {
	for i, d := range drops {
		if err := os.Remove(d.Path); err != nil && !os.IsNotExist(err) {
			_ = UndoDrops(drops[:i])
			return fail.New(fail.Other, "cannot remove %s", d.Path)
		}
	}
	return nil
}

// UndoDrops writes each file of <drops> back, with its mode. A path that is
// a folder now (the old path of a group, where the sync wrote the parts)
// goes first: before the drop it was a file, so the folder holds only what
// the sync put there.
func UndoDrops(drops []Drop) error {
	for i := len(drops) - 1; i >= 0; i-- {
		d := drops[i]
		if info, err := os.Lstat(d.Path); err == nil && info.IsDir() {
			if err := os.RemoveAll(d.Path); err != nil {
				return fail.New(fail.Other, "cannot write %s back", d.Path)
			}
		}
		if err := os.MkdirAll(path.Dir(d.Path), 0o777); err != nil {
			return fail.New(fail.Other, "cannot write %s back", d.Path)
		}
		if err := os.WriteFile(d.Path, d.Old, d.Mode); err != nil {
			return fail.New(fail.Other, "cannot write %s back", d.Path)
		}
		if err := os.Chmod(d.Path, d.Mode); err != nil {
			return fail.New(fail.Other, "cannot write %s back", d.Path)
		}
	}
	return nil
}
