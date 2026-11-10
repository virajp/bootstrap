package target

import (
	"fmt"
	"os"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/report"
)

// DirPerm keeps the folder <dir> at the mode <mode>: the `dir_perm` of a
// `[files]` entry. <show> names the folder in a line. <done> holds the
// folders of the run that are already at their mode, so each one is examined
// once. The check reads `<dir>/.`, so a linked folder gives the mode of its
// target. A real run creates the folder when it is missing. It gives true
// when it changed the mode. After a failure, the folder stays not done, so
// each later file below it also fails.
func (e *Env) DirPerm(dir, show string, mode os.FileMode, done map[string]bool) (bool, fail.Code) {
	if done[dir] {
		return false, fail.OK
	}
	octal := fmt.Sprintf("%o", mode)
	st, err := os.Stat(dir)
	if err == nil && st.IsDir() {
		if m, err := Mode(dir + "/."); err == nil && m == mode {
			done[dir] = true
			return false, fail.OK
		}
	} else if e.DryRun {
		done[dir] = true
		return false, fail.OK
	}
	if e.DryRun {
		done[dir] = true
		e.Report.Outcome(report.Changed, show, "", "would chmod "+octal+" "+show)
		return false, fail.OK
	}
	if os.MkdirAll(dir, 0o777) != nil || os.Chmod(dir, mode) != nil {
		e.Report.Outcome(report.Failed, show, "cannot chmod", "cannot chmod "+show)
		return false, fail.Other
	}
	done[dir] = true
	e.Report.Outcome(report.Changed, show, "", "chmod "+octal+"  "+show)
	return true, fail.OK
}
