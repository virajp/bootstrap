package cli

import (
	"strings"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// writeData is the data of a write command with --json (init, sync, add,
// remove, replace, detect --add): the plan, and after an apply one outcome
// for each call of report.Outcome. A dry run has no outcome.
type writeData struct {
	Plan     planData      `json:"plan"`
	Outcomes []outcomeData `json:"outcomes"`
}

// planData is data.plan (D23): the fields of core.PlanResult that a caller
// reads. The warnings of the plan go to the warnings of the envelope.
type planData struct {
	Changes []changeData  `json:"changes"`
	Files   []fileData    `json:"files"`
	Skipped []skipData    `json:"skipped"`
	Pending []pendingData `json:"pending"`
	Notes   []string      `json:"notes"`
}

// changeData is an oldform.Change: a write before the sync. Content is the
// text of the file.
type changeData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Remove  bool   `json:"remove"`
	Note    string `json:"note"`
}

// fileData is a target.Planned: a file that the sync writes, changes or
// removes.
type fileData struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

// skipData is a core.Skip: a member of a pack that add pack leaves out.
type skipData struct {
	Tool string `json:"tool"`
	Why  string `json:"why"`
}

// pendingData is a questions.Question with no answer.
type pendingData struct {
	Tool     string `json:"tool"`
	Key      string `json:"key"`
	Text     string `json:"text"`
	Detect   string `json:"detect"`
	Example  string `json:"example"`
	Group    string `json:"group"`
	Note     string `json:"note"`
	Optional bool   `json:"optional"`
}

// outcomeData is one call of report.Outcome. Message is the reason of a
// failed target.
type outcomeData struct {
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// planDataOf gives data.plan of <p>, with each list [] when it is empty.
func planDataOf(p *core.PlanResult) planData {
	d := planData{
		Changes: []changeData{}, Files: []fileData{}, Skipped: []skipData{},
		Pending: []pendingData{}, Notes: []string{},
	}
	for _, c := range p.Changes {
		d.Changes = append(d.Changes, changeData{Path: c.Path, Content: string(c.Content), Remove: c.Remove, Note: c.Note})
	}
	for _, f := range p.Files {
		d.Files = append(d.Files, fileData{Path: f.Path, Kind: f.Kind.String()})
	}
	for _, s := range p.Skipped {
		d.Skipped = append(d.Skipped, skipData{Tool: s.Tool, Why: s.Why})
	}
	for _, q := range p.Pending {
		d.Pending = append(d.Pending, pendingData{
			Tool: q.Tool, Key: q.Key, Text: q.Text, Detect: q.Detect,
			Example: q.Example, Group: q.Group, Note: q.Note, Optional: q.Optional,
		})
	}
	d.Notes = append(d.Notes, p.Notes...)
	return d
}

// outcomesOf gives the outcomes of a run, [] when it has none.
func outcomesOf(rs []report.Result) []outcomeData {
	out := []outcomeData{}
	for _, r := range rs {
		out = append(out, outcomeData{Kind: r.Kind.String(), Path: r.Path, Message: r.Text})
	}
	return out
}

// pendingJSON ends a write command with --json whose plan has questions with
// no answer, and never asks them. A dry run exits 0 with one warning (D45);
// with --yes, it writes nothing and stops with code 2 and one error (D22).
// data.plan.pending lists the questions.
func (a *app) pendingJSON(p *core.PlanResult, o opts) fail.Code {
	names := make([]string, len(p.Pending))
	for i, q := range p.Pending {
		names[i] = q.Name()
	}
	list := strings.Join(names, ", ")
	if o.dryRun {
		a.warnf("the questions %s have no answer - a run with --yes stops", list)
		return fail.OK
	}
	a.fail("the questions "+list+" have no answer; nothing written",
		"set each with bootstrap config set tools.<tool>.<key> <value>, or give it with --answer <tool>.<key>=<value>")
	return fail.Usage
}
