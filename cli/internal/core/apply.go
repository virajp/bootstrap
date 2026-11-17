package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/scripts"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// StepKind is the kind of a step of Apply.
type StepKind int

// The steps of Apply before the sync.
const (
	// Ran: the script Path ran; Stderr is its output on stderr.
	Ran StepKind = iota + 1
	// Wrote: Apply wrote the file Path.
	Wrote
	// Removed: Apply removed the file Path.
	Removed
	// Moved: Apply moved the file Path to To.
	Moved
)

// Step is one step of Apply before the sync.
type Step struct {
	Kind   StepKind
	Path   string // The shown path of a script; the absolute path of a file.
	To     string // Moved: the absolute path that the file moved to.
	Stderr []byte
}

// Apply applies the plan <p>: it runs the script that comes before the
// writes, writes the values file and the changes of a migration, and runs the
// sync. When the sync fails, it puts the values file back as it was before
// Apply (or removes a values file that did not exist), and gives the code of
// the sync and true; the other files stay as the sync left them. A sync that
// stops with fail.Terminal (the Mac init stage changed) did its work, so the
// values file stays and Apply gives that code and false. A problem
// before the sync gives its code and an error. It also gives the outcomes of
// the sync, one for each call of report.Outcome, also when the sync fails.
func Apply(p *PlanResult, o Options) (fail.Code, bool, []report.Result, error) {
	step := func(s Step) {
		if o.Step != nil {
			o.Step(s)
		}
	}
	if sc := p.Before; sc != nil {
		res := sc.Runner.Run([]scripts.Job{sc.Job})[0]
		shown := sc.Runner.ShownPath(res.Tool, res.Name)
		if res.Code != 0 {
			var errs []error
			for _, l := range strings.Split(strings.TrimSpace(string(res.Stderr)), "\n") {
				if l != "" {
					errs = append(errs, problem(fail.Usage, shown+": "+l, ""))
				}
			}
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("%s failed with exit code %d", shown, res.Code), ""))
			return fail.Usage, false, nil, join(errs)
		}
		step(Step{Kind: Ran, Path: shown, Stderr: res.Stderr})
	}
	rest := p.Changes
	if p.Migrated {
		var err error
		if rest, err = p.writeFolder(); err != nil {
			return fail.CodeOf(err), false, nil, err
		}
	}
	if err := oldform.Apply(rest); err != nil {
		return fail.CodeOf(err), false, nil, err
	}
	for _, ch := range p.Changes {
		for _, m := range ch.Moves {
			step(Step{Kind: Moved, Path: m.From, To: m.To})
		}
		for _, e := range ch.Edits {
			step(Step{Kind: Wrote, Path: e.Path})
		}
		for _, d := range ch.Drops {
			step(Step{Kind: Removed, Path: d.Path})
		}
		if len(ch.Moves) > 0 || len(ch.Edits) > 0 || len(ch.Drops) > 0 {
			continue
		}
		if ch.Remove {
			step(Step{Kind: Removed, Path: ch.Path})
		} else {
			step(Step{Kind: Wrote, Path: ch.Path})
		}
	}
	out, errw := o.writers()
	// The scripts of the sync read the values with bootstrap config (D19).
	restoreEnv := setEnv("BOOTSTRAP_VALUES", p.Values, !p.Mac)
	_, outcomes, code := frontEnd(p.Mac)(SyncArgs(p.Mac, false, o.NoBootstrap), out, errw)
	restoreEnv()
	if code != fail.OK && code != fail.Terminal {
		if err := p.restore(); err != nil {
			return code, false, outcomes, err
		}
		return code, true, outcomes, nil
	}
	return code, false, outcomes, nil
}

// writeFolder writes the files of the migration below the values folder
// into a temp folder next to it and renames it into place (D70). It gives
// the other changes: the removals of the old files come after the rename.
func (p *PlanResult) writeFolder() ([]oldform.Change, error) {
	dir := strings.TrimSuffix(p.Values, "/")
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, 0o777); err != nil {
		return nil, problem(fail.Other, "cannot create "+p.Show(parent), "")
	}
	tmp, err := os.MkdirTemp(parent, "."+filepath.Base(dir)+".new.")
	if err != nil {
		return nil, problem(fail.Other, "cannot create a temp folder in "+p.Show(parent), "")
	}
	var rest []oldform.Change
	for _, ch := range p.Changes {
		rel, ok := strings.CutPrefix(ch.Path, dir+"/")
		if !ok || ch.Remove {
			rest = append(rest, ch)
			continue
		}
		if err := os.WriteFile(filepath.Join(tmp, rel), ch.Content, 0o644); err != nil {
			_ = os.RemoveAll(tmp)
			return nil, problem(fail.Other, "cannot write "+p.Show(ch.Path), "")
		}
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		_ = os.RemoveAll(tmp)
		return nil, problem(fail.Other, "cannot write "+p.Show(dir), "")
	}
	if err := os.Rename(tmp, dir); err != nil {
		_ = os.RemoveAll(tmp)
		return nil, problem(fail.Other, "cannot write "+p.Show(dir), "")
	}
	return rest, nil
}

// setEnv sets the variable <key> to <value> when <on>, and gives the
// function that puts its old value back.
func setEnv(key, value string, on bool) func() {
	if !on {
		return func() {}
	}
	old, had := os.LookupEnv(key)
	_ = os.Setenv(key, value)
	return func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

// RestoredFile gives the file that a failed Apply put back: the old values
// file of a migration, else the edited file.
func (p *PlanResult) RestoredFile() string {
	if p.Migrated && len(p.Old) > 0 {
		return p.Old[0].Path
	}
	return p.EditFile()
}

// restore moves the moved files back, writes the files of a change with
// Edits back, puts the edited values file back as it
// was before Apply, and removes a repo.toml that init created, and then the
// values folder when it is empty.
func (p *PlanResult) restore() error {
	// The removed old copies and the moved own test tasks go back first
	// (D5, D7), then the files of the rename of setup:worktree (D20).
	for _, ch := range p.Changes {
		if err := oldform.UndoDrops(ch.Drops); err != nil {
			return problem(fail.Other, err.Error(), "")
		}
		if err := oldform.UndoMoves(ch.Moves); err != nil {
			return problem(fail.Other, err.Error(), "")
		}
		for _, e := range ch.Edits {
			if err := oldform.WriteFile(e.Path, e.Old); err != nil {
				return problem(fail.Other, "cannot write "+p.Show(e.Path)+" back", "")
			}
		}
	}
	if p.Migrated {
		// The old files come back, and the folder goes (D83).
		if err := os.RemoveAll(p.Values); err != nil {
			return problem(fail.Other, "cannot remove "+p.Show(p.Values), "")
		}
		for _, o := range p.Old {
			if err := os.WriteFile(o.Path, o.Content, 0o644); err != nil {
				return problem(fail.Other, "cannot write "+p.Show(o.Path)+" back", "")
			}
		}
		return nil
	}
	file := p.EditFile()
	if p.RepoNew != nil {
		repoFile := values.RepoFile(p.Values)
		if err := os.Remove(repoFile); err != nil && !os.IsNotExist(err) {
			return problem(fail.Other, "cannot remove "+p.Show(repoFile), "")
		}
	}
	if !p.Existed {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return problem(fail.Other, "cannot remove "+p.Show(file), "")
		}
		if !p.Mac {
			_ = os.Remove(p.Values)
		}
		return nil
	}
	mode := os.FileMode(0o644)
	if st, err := os.Stat(file); err == nil {
		mode = st.Mode().Perm()
	}
	if err := os.WriteFile(file, p.Cur, mode); err != nil {
		return problem(fail.Other, "cannot write "+p.Show(file)+" back", "")
	}
	return nil
}
