package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/scripts"
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
)

// Step is one step of Apply before the sync.
type Step struct {
	Kind   StepKind
	Path   string // The shown path of a script; the absolute path of a file.
	Stderr []byte
}

// Apply applies the plan <p>: it runs the script that comes before the
// writes, writes the values file and the changes of a migration, and runs the
// sync. When the sync fails, it puts the values file back as it was before
// Apply (or removes a values file that did not exist), and gives the code of
// the sync and true; the other files stay as the sync left them. A sync that
// stops with fail.Terminal (the Mac init stage changed) did its work, so the
// values file stays and Apply gives that code and false. A problem
// before the sync gives its code and an error.
func Apply(p *PlanResult, o Options) (fail.Code, bool, error) {
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
			return fail.Usage, false, join(errs)
		}
		step(Step{Kind: Ran, Path: shown, Stderr: res.Stderr})
	}
	if err := oldform.Apply(p.Changes); err != nil {
		return fail.CodeOf(err), false, err
	}
	for _, ch := range p.Changes {
		if ch.Remove {
			step(Step{Kind: Removed, Path: ch.Path})
		} else {
			step(Step{Kind: Wrote, Path: ch.Path})
		}
	}
	out, errw := o.writers()
	_, code := frontEnd(p.Mac)(SyncArgs(p.Mac, false, o.NoBootstrap), out, errw)
	if code != fail.OK && code != fail.Terminal {
		if err := p.restore(); err != nil {
			return code, false, err
		}
		return code, true, nil
	}
	return code, false, nil
}

// restore puts the values file back as it was before Apply.
func (p *PlanResult) restore() error {
	if !p.Existed {
		if err := os.Remove(p.Values); err != nil && !os.IsNotExist(err) {
			return problem(fail.Other, "cannot remove "+p.Show(p.Values), "")
		}
		return nil
	}
	mode := os.FileMode(0o644)
	if st, err := os.Stat(p.Values); err == nil {
		mode = st.Mode().Perm()
	}
	if err := os.WriteFile(p.Values, p.Cur, mode); err != nil {
		return problem(fail.Other, "cannot write "+p.Show(p.Values)+" back", "")
	}
	return nil
}
