package target

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/format"
	"github.com/virajp/bootstrap/sync-engine/internal/header"
	"github.com/virajp/bootstrap/sync-engine/internal/report"
)

// SeedKept is _se_seed_kept: true when the seed <dest> exists (after
// symlinks). It then stays as it is, and the caller reports `kept seed`.
func SeedKept(dest string) bool {
	_, err := os.Stat(dest)
	return err == nil
}

// exists is the bash test -e: the path exists after symlinks.
func exists(path string) bool { _, err := os.Stat(path); return err == nil }

// isLink is the bash test -L.
func isLink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

// isRegular is the bash test -f: a regular file after symlinks.
func isRegular(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// Unblock is _se_unblock. It walks from the folder <dir> up to <root> (not
// <root> itself). The first regular file (not a symlink) on that walk is at
// the place of a folder: a synced file (with the header) is removed, or with
// a dry run listed. <spare> holds the paths, relative to <root>, that the sync
// must not remove: the seeds and the merge targets. <target> names the file
// that needs the folder, as a failure reports it. It gives 1, after a message,
// when that file has no header, is in <spare>, or cannot be removed.
func (e *Env) Unblock(dir, root string, spare []string, target string) fail.Code {
	r := e.Report
	for strings.HasPrefix(dir, root+"/") {
		if isRegular(dir) && !isLink(dir) {
			show := e.Show(dir)
			if !header.HasFile(dir) || contains(spare, dir[len(root)+1:]) {
				text := show + " is a file of this repo, but the sync needs a folder there - move it"
				r.Outcome(report.Failed, target, text, text)
				return fail.Other
			}
			if e.DryRun {
				r.Outcome(report.Removed, show, "", "would remove "+show+" (a folder replaces it)")
				return fail.OK
			}
			if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
				r.Outcome(report.Failed, target, "cannot write the target", "")
				return fail.Other
			}
			r.Outcome(report.Removed, show, "", "removed    "+show+" (a folder replaces it)")
			return fail.OK
		}
		dir = Dirname(dir)
	}
	return fail.OK
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Failure reports a failed parse check or merge of internal/format in the
// order of bash: the tool output, the outcome `failed`, then the detail.
func (e *Env) Failure(f *format.Failure, show string) {
	if f.Stderr != "" {
		fmt.Fprint(e.Stderr, f.Stderr)
	}
	if f.Text != "" {
		e.Report.Outcome(report.Failed, show, f.Text, f.Line)
	}
	fmt.Fprint(e.Stderr, f.IndentedDetail())
}

// Check finds the format of the target <target> (its path, with or without
// `.tmpl`) and runs the parse check of the temp file <tmp>. <explicit> is the
// `format` key of its `[files]` table, or "". For a merge target (<merge>),
// the JSONC rule does not apply. A failure gives 2, after the messages.
func (e *Env) Check(tmp, target, explicit string, merge bool, show string) (format.Format, fail.Code) {
	content, _ := os.ReadFile(tmp)
	var f format.Format
	if merge {
		f = format.Detect(target, nil, explicit)
	} else {
		f = format.Detect(target, content, explicit)
	}
	if fl := format.ParseCheckFile(tmp, f, show); fl != nil {
		e.Failure(fl, show)
		return f, fail.Code(fl.Code)
	}
	return f, fail.OK
}

// Merge is the call of _se_merge in a front-end: it merges the temp file
// <tmp> (the merge source) over the live target <live> into <tmp>. A <live>
// that is not a regular file is absent. A failure gives its code, after the
// messages.
func (e *Env) Merge(tmp, live string, f format.Format, ops format.Ops, show string) fail.Code {
	src, err := os.ReadFile(tmp)
	if err != nil {
		return fail.Other
	}
	var cur []byte
	if isRegular(live) {
		if cur, err = os.ReadFile(live); err != nil {
			cur = nil
		}
	}
	out, fl := format.Merge(src, cur, f, ops, show)
	if fl != nil {
		e.Failure(fl, show)
		return fail.Code(fl.Code)
	}
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return fail.Other
	}
	return fail.OK
}

// Abort ends a failed target: it removes the temp file <tmp>, and when <dest>
// exists, it reports `kept existing`.
func (e *Env) Abort(tmp, dest, show string) {
	if tmp != "" {
		os.Remove(tmp)
	}
	if exists(dest) {
		e.Report.Outcome(report.Failed, show, "kept existing", "kept existing "+show)
	}
}

// sameContent is `cmp -s <a> <b>`.
func sameContent(a, b string) bool {
	x, err := os.ReadFile(a)
	if err != nil {
		return false
	}
	y, err := os.ReadFile(b)
	if err != nil {
		return false
	}
	return bytes.Equal(x, y)
}

// Finish is _se_finish: it compares the temp file <tmp> with <dest>, then
// writes it with <mode>, or with a dry run prints the change and a unified
// diff. <show> names <dest> in each message. <fileMode> is the mode of the
// `[files]` table: a written `seed` reports as seeded. It gives 1, after a
// message, when the chmod or the rename fails. The temp file is gone after
// the call.
func (e *Env) Finish(mode os.FileMode, tmp, dest, show, fileMode string) fail.Code {
	r := e.Report
	kind := report.Changed
	if fileMode == "seed" {
		kind = report.Seeded
	} else if !exists(dest) && !isLink(dest) {
		kind = report.New
	}
	octal := fmt.Sprintf("%o", mode)
	if isRegular(dest) && sameContent(tmp, dest) {
		os.Remove(tmp)
		cur, err := Mode(dest)
		switch {
		case err == nil && cur == mode:
			r.Outcome(report.Unchanged, show, "", "unchanged  "+show)
		case e.DryRun:
			r.Outcome(report.Changed, show, "", "would chmod "+octal+" "+show)
		default:
			if err := os.Chmod(dest, mode); err != nil {
				r.Outcome(report.Failed, show, "cannot chmod", "cannot chmod "+show)
				return fail.Other
			}
			r.Outcome(report.Changed, show, "", "chmod "+octal+"  "+show)
		}
		return fail.OK
	}

	if e.DryRun {
		r.Outcome(kind, show, "", "would update "+show)
		if isRegular(dest) {
			cmd := exec.Command("diff", "-u", "--label", show+" (current)", "--label", show+" (bootstrap)", "--", dest, tmp)
			cmd.Stdout = e.Stderr
			cmd.Stderr = e.Stderr
			_ = cmd.Run()
		}
		os.Remove(tmp)
		return fail.OK
	}

	_ = os.Chmod(tmp, mode)
	if err := Move(tmp, dest); err != nil {
		os.Remove(tmp)
		r.Outcome(report.Failed, show, "cannot write", "cannot write "+show)
		return fail.Other
	}
	r.Outcome(kind, show, "", "updated    "+show)
	return fail.OK
}
