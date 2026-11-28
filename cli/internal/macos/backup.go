package macos

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/manifest"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// This file holds the backup clone: the key `backup` of macos.toml. The sync
// links the synced files of $HOME into <backup>/dotfiles/, before the render,
// so the render writes through the links into the clone. Contract:
// docs/design/mac.md ("The backup clone").

// backupDir is the folder of the clone that holds the files, with the paths
// of $HOME.
const backupDir = "dotfiles"

// link is one link of the backup: a path relative to $HOME, a folder or a
// file.
type link struct {
	rel    string
	folder bool
}

// neverLinked are the folders of $HOME that the rule never links: their
// children are linked one by one.
var neverLinked = map[string]bool{".config": true}

// linkPlan gives the links of the targets <targets> (relative to $HOME) with
// the runtime files <runtime> (relative to $HOME), sorted by path. A folder is
// linked when no runtime file is in it; else each of its children that holds a
// target, with the same rule. $HOME itself and the folders of neverLinked are
// never linked.
func linkPlan(targets, runtime []string) []link {
	hasRuntime := func(dir string) bool {
		for _, r := range runtime {
			if r == dir || strings.HasPrefix(r, dir+"/") {
				return true
			}
		}
		return false
	}
	seen := map[string]bool{}
	var out []link
	for _, t := range targets {
		parts := strings.Split(t, "/")
		for i := 1; i <= len(parts); i++ {
			p := strings.Join(parts[:i], "/")
			if i == len(parts) {
				if !seen[p] {
					seen[p] = true
					out = append(out, link{rel: p})
				}
				break
			}
			if neverLinked[p] || hasRuntime(p) {
				continue
			}
			if !seen[p] {
				seen[p] = true
				out = append(out, link{rel: p, folder: true})
			}
			break
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out
}

// linkTargets gives the targets of <plan> that get a link: each target
// below $HOME. A target of root = "backup" is a real file in the clone, so
// it gets no link.
func linkTargets(plan []planItem) []string {
	var out []string
	for _, it := range plan {
		if !it.backup {
			out = append(out, it.target)
		}
	}
	return out
}

// step is one change of the adoption of a link.
type step struct {
	kind     int    // stepMove, stepRemove, stepMkdir or stepLink.
	from, to string // Absolute paths. stepRemove and stepMkdir use from only.
}

const (
	stepMove   = iota // Move <from> into the clone at <to>.
	stepRemove        // Remove <from>: a link or a file that the clone has.
	stepMkdir         // Make the folder <from> in the clone.
	stepLink          // Make the link <from> to <to>.
)

// adoption holds the steps and the conflicts of the links of a run.
type adoption struct {
	steps     []step
	conflicts []string
}

// sameFile tells if the files <a> and <b> have the same bytes.
func sameFile(a, b string) bool {
	x, err1 := os.ReadFile(a)
	y, err2 := os.ReadFile(b)
	return err1 == nil && err2 == nil && bytes.Equal(x, y)
}

// realOf gives the real path of <p>, or "" when it cannot resolve.
func realOf(p string) string {
	r, ok := manifest.Real(p)
	if !ok {
		return ""
	}
	return r
}

// merge examines the item <h> of $HOME against its place <c> in the clone,
// and adds the steps that put its content into the clone without loss. A
// difference that can lose data is a conflict. <show> names a path in a
// message.
func (a *adoption) merge(h, c string, show func(string) string) {
	hi, err := os.Lstat(h)
	if err != nil {
		return
	}
	ci, cerr := os.Lstat(c)
	switch {
	case hi.Mode()&os.ModeSymlink != 0:
		if cerr == nil && realOf(h) != "" && realOf(h) == realOf(c) {
			a.steps = append(a.steps, step{kind: stepRemove, from: h})
			return
		}
		dest, _ := os.Readlink(h)
		a.conflicts = append(a.conflicts, show(h)+" is a link to "+dest+", not to "+show(c))
	case cerr != nil:
		a.steps = append(a.steps, step{kind: stepMove, from: h, to: c})
	case hi.Mode().IsRegular() && ci.Mode().IsRegular():
		if sameFile(h, c) {
			a.steps = append(a.steps, step{kind: stepRemove, from: h})
			return
		}
		a.conflicts = append(a.conflicts, show(h)+" and "+show(c)+" differ")
	case hi.IsDir() && ci.IsDir():
		entries, err := os.ReadDir(h)
		if err != nil {
			a.conflicts = append(a.conflicts, "cannot read "+show(h))
			return
		}
		for _, e := range entries {
			a.merge(h+"/"+e.Name(), c+"/"+e.Name(), show)
		}
		a.steps = append(a.steps, step{kind: stepRemove, from: h})
	default:
		a.conflicts = append(a.conflicts, show(h)+" and "+show(c)+" are not the same kind of item")
	}
}

// plan examines the link <l> of $HOME <home> into the clone folder <root>.
func (a *adoption) plan(l link, home, root string, show func(string) string) {
	h, c := home+"/"+l.rel, root+"/"+l.rel
	// An item above the link must be a real folder: a link there already
	// holds the path, and the rule does not link that folder.
	parts := strings.Split(l.rel, "/")
	for i := 1; i < len(parts); i++ {
		up := home + "/" + strings.Join(parts[:i], "/")
		if fi, err := os.Lstat(up); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			a.conflicts = append(a.conflicts, show(up)+" is a link, but a tool writes runtime files in it - replace the link with a folder")
			return
		}
	}
	if fi, err := os.Lstat(h); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		if dest, _ := os.Readlink(h); dest == c {
			return
		}
	}
	a.merge(h, c, show)
	if len(a.conflicts) > 0 {
		return
	}
	if l.folder {
		if _, err := os.Stat(c); err != nil {
			if mv := a.movesTo(c); !mv {
				a.steps = append(a.steps, step{kind: stepMkdir, from: c})
			}
		}
	}
	a.steps = append(a.steps, step{kind: stepLink, from: h, to: c})
}

// movesTo tells if a step moves an item to <c>.
func (a *adoption) movesTo(c string) bool {
	for _, st := range a.steps {
		if st.kind == stepMove && st.to == c {
			return true
		}
	}
	return false
}

// backup links the targets of the plan into the backup clone, when macos.toml
// has `backup`. It runs before the render. A conflict stops the run with
// code 2, before any change. A dry run shows the changes and writes nothing.
func (s *state) backup(dryRun bool) fail.Code {
	if s.clone == "" {
		return fail.OK
	}
	clone := s.clone
	if fi, err := os.Stat(clone); err != nil || !fi.IsDir() {
		s.log("the backup clone %s is not a folder - set backup in %s to your clone, or remove the key", clone, valuesName)
		return fail.Usage
	}
	s.cloneReal = realOf(clone)
	root := clone + "/" + backupDir

	targets := linkTargets(s.plan)
	var runtime []string
	for _, g := range s.gen {
		targets = append(targets, g.target)
	}
	if rel, ok := strings.CutPrefix(s.config, s.home+"/"); ok {
		targets = append(targets, rel+"/"+valuesName)
	}
	for _, tool := range s.res.Tools {
		if m := s.idx.Manifests[tool]; m != nil {
			runtime = append(runtime, m.Runtime...)
		}
	}

	show := s.showPath
	a := &adoption{}
	for _, l := range linkPlan(targets, runtime) {
		a.plan(l, s.home, root, show)
	}
	if len(a.conflicts) > 0 {
		for _, c := range a.conflicts {
			s.log("backup: %s - fix it by hand, then run the sync again", c)
		}
		s.r.Summary(int(fail.Usage))
		return fail.Usage
	}
	for _, st := range a.steps {
		if code := s.doStep(st, dryRun, show); code != fail.OK {
			s.r.Summary(int(code))
			return code
		}
	}
	return fail.OK
}

// doStep does one step of the adoption, or shows it in a dry run.
func (s *state) doStep(st step, dryRun bool, show func(string) string) fail.Code {
	var err error
	switch st.kind {
	case stepMove:
		if dryRun {
			s.e.Plan(st.from, report.Changed)
			s.r.Outcome(report.Changed, show(st.from), "", "would move "+show(st.from)+" into the backup "+show(st.to))
			return fail.OK
		}
		if err = os.MkdirAll(filepath.Dir(st.to), 0o755); err == nil {
			err = os.Rename(st.from, st.to)
		}
	case stepRemove:
		if dryRun {
			return fail.OK
		}
		err = os.Remove(st.from)
	case stepMkdir:
		if dryRun {
			return fail.OK
		}
		err = os.MkdirAll(st.from, 0o755)
	case stepLink:
		if dryRun {
			s.e.Plan(st.from, report.New)
			s.r.Outcome(report.New, show(st.from), "", "would link "+show(st.from)+" to "+show(st.to))
			return fail.OK
		}
		if err = os.MkdirAll(filepath.Dir(st.from), 0o755); err == nil {
			err = os.Symlink(st.to, st.from)
		}
		if err == nil {
			s.r.Outcome(report.New, show(st.from), "", "linked     "+show(st.from)+" to "+show(st.to))
		}
	}
	if err != nil {
		s.log("backup: cannot change %s: %v", show(st.from), err)
		return fail.Other
	}
	if st.kind == stepMove {
		s.r.Outcome(report.Changed, show(st.from), "", "moved      "+show(st.from)+" into the backup")
	}
	return fail.OK
}
