package tui

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// The review and Apply (screen 8 of the design): a full screen with one line
// for each change of the list (a pack is one line), the files of the dry
// run of core.Plan, the dirty targets of the clean-target rule, and Apply.
// Apply asks the questions with no answer, releases the terminal, and calls
// core.Apply, so the output of the sync goes to the terminal.

// nextRepo is the next command after a sync of the repo side that passes
// (D12 b). The Mac side has none.
const nextRepo = "MISE_ENV=dev mise run setup"

// The review screen.
const (
	reviewH  = screenH - 5 // The body: the screen without the header, two rules and the key line.
	filesMax = 6           // The files that show before f.
	opW      = 7           // The column of the kind of a file.
	chgW     = 18          // The column of the name of a change.
)

// review is the state of the review screen.
type review struct {
	plan  *core.PlanResult
	dirty []string
	all   bool // f: each file shows.
	top   int  // The first body line that shows.
}

// applied is the end of an Apply: the lines of its result, which show in
// place of the screen until a key.
type applied struct {
	lines []line
}

// planOf plans the change list with core.Plan. The output of a dry run that
// fails is the message of its problem.
func (m *Model) planOf() (*core.PlanResult, error) {
	var out bytes.Buffer
	p, err := core.Plan(m.opt.Side, m.list, core.Options{NoBootstrap: m.opt.NoBootstrap, Stdout: &out, Stderr: &out})
	if err != nil && err.Error() == "" {
		msg := strings.TrimSpace(stripStyle(out.String()))
		if msg == "" {
			msg = fmt.Sprintf("the dry run of the sync failed with exit code %d", fail.CodeOf(err))
		}
		err = &core.Problem{Code: fail.CodeOf(err), Msg: msg}
	}
	return p, err
}

// openReview plans the change list and shows the review. A problem of
// core shows in a dialog, and the change list stays (D13 b).
func (m *Model) openReview() tea.Cmd {
	if m.changes() == 0 {
		return nil
	}
	p, err := m.planOf()
	if err != nil {
		m.modal = errModal(err)
		return nil
	}
	m.review = &review{plan: p, dirty: p.Dirty()}
	return nil
}

// reviewKey handles a key of the review.
func (m *Model) reviewKey(k tea.KeyPressMsg) tea.Cmd {
	r := m.review
	switch k.String() {
	case "esc", "q":
		m.review = nil
	case "ctrl+c":
		return m.quit()
	case "enter":
		return m.apply()
	case "f":
		r.all = !r.all
		r.top = 0
	case "up":
		r.top = max(r.top-1, 0)
	case "down":
		r.top = min(r.top+1, max(len(m.reviewBody())-reviewH, 0))
	}
	return nil
}

// apply asks the questions with no answer, one form for each tool in the
// order of the tools, plans again with the answers, and applies the plan
// after the consent of writeModal.
// Cancel in a form stops Apply and keeps the change list (D6). A dirty
// target refuses Apply (D10 e).
func (m *Model) apply() tea.Cmd {
	r := m.review
	if len(r.dirty) > 0 {
		return nil
	}
	if len(r.plan.Pending) == 0 {
		m.modal = m.writeModal(r.plan)
		return nil
	}
	return m.askEach(byTool(r.plan.Pending),
		func(tool string, n int) string { return "Apply · " + tool + " needs " + plural(n, "value") },
		func(m *Model, a []core.Answer) { m.setAnswers(a) },
		func(m *Model) tea.Cmd {
			p, err := m.planOf()
			if err != nil {
				m.modal = errModal(err)
				return nil
			}
			m.review = &review{plan: p, dirty: p.Dirty(), all: r.all}
			return m.apply()
		},
		func(m *Model) tea.Cmd {
			m.review = nil
			return nil
		})
}

// writeModal asks "Write <n> changes? [y/N]" before Apply writes, as the
// consent of the CLI (D23): y applies the plan <p>, and any other key goes
// back to the review with the change list kept.
func (m *Model) writeModal(p *core.PlanResult) *modal {
	return &modal{
		title: "Apply",
		width: 50,
		body:  []line{{text("Write " + plural(m.changes(), "change") + "? [y/N]")}},
		keys:  "y write  any other key goes back",
		onKey: func(m *Model, k tea.KeyPressMsg) tea.Cmd {
			m.modal = nil
			if k.String() == "y" {
				return m.runApply(p)
			}
			return nil
		},
	}
}

// applyCmd is core.Apply as a command of tea.Exec: the TUI releases the
// terminal while it runs, so the script before the writes, the sync and its
// output write to the terminal (D12 c).
type applyCmd struct {
	m         *Model
	p         *core.PlanResult
	n         int
	out, errw io.Writer
	code      fail.Code
	putBack   bool
	err       error
}

func (c *applyCmd) SetStdin(io.Reader)    {}
func (c *applyCmd) SetStdout(w io.Writer) { c.out = w }
func (c *applyCmd) SetStderr(w io.Writer) { c.errw = w }

// Run applies the plan, and prints each step before the sync.
func (c *applyCmd) Run() error {
	sd := c.m.opt.Side
	fmt.Fprintf(c.out, "\nApplying %s ...\n", plural(c.n, "change"))
	opt := core.Options{NoBootstrap: c.m.opt.NoBootstrap, Stdout: c.out, Stderr: c.errw, Step: func(s core.Step) {
		switch s.Kind {
		case core.Ran:
			_, _ = c.errw.Write(s.Stderr)
			fmt.Fprintf(c.out, "  ✓ ran %s\n", s.Path)
		case core.Wrote:
			fmt.Fprintf(c.out, "  ✓ wrote %s\n", sd.Show(s.Path))
		case core.Removed:
			fmt.Fprintf(c.out, "  ✓ removed %s\n", sd.Show(s.Path))
		}
	}}
	c.code, c.putBack, c.err = core.Apply(c.p, opt)
	return nil
}

// appliedMsg is the end of core.Apply.
type appliedMsg struct {
	c   *applyCmd
	err error
}

// runApply applies the plan <p> with the terminal released.
func (m *Model) runApply(p *core.PlanResult) tea.Cmd {
	c := &applyCmd{m: m, p: p, n: m.changes()}
	// An empty result leaves the alternate screen before the terminal is
	// released, so the output of the sync stays in the main buffer.
	m.applied = &applied{}
	return tea.Exec(c, func(err error) tea.Msg { return appliedMsg{c: c, err: err} })
}

// afterApply shows the result of Apply. After a sync that passes, the
// change list is empty and the TUI reads the values file again; after a
// failed sync, core.Apply put the values file back, and the change list
// stays.
func (m *Model) afterApply(msg appliedMsg) {
	m.review = nil
	sd := m.opt.Side
	sync := core.CmdName(sd.Mac, "sync")
	var out []line
	ok := false
	switch {
	case msg.err != nil:
		out = append(out, line{styled("  ✗ ", stRemove), text(msg.err.Error())})
	case msg.c.err != nil:
		for _, l := range strings.Split(msg.c.err.Error(), "\n") {
			out = append(out, line{styled("  ✗ ", stRemove), text(l)})
		}
	case msg.c.code != fail.OK && msg.c.code != fail.Terminal:
		out = append(out, line{styled("  ✗ ", stRemove), text(fmt.Sprintf("%s: exit code %d", sync, msg.c.code))})
		if msg.c.putBack {
			out = append(out, line{styled("  ↺ ", stLock), text("put " + sd.Show(sd.Values) + " back as it was before Apply")})
		}
	default:
		ok = true
		out = append(out, line{styled("  ✓ ", stSel), text(sync + ": done")})
	}
	out = append(out, line{})
	if !ok {
		n := m.changes()
		kept := "Your " + plural(n, "change") + " are kept."
		if n == 1 {
			kept = "Your 1 change is kept."
		}
		out = append(out, line{styled("The sync failed. "+kept+" Press any key to go back.", stLock)})
		m.applied = &applied{lines: out}
		return
	}
	if !sd.Mac {
		out = append(out, line{styled("Next: "+nextRepo, stDim)}, line{})
	}
	out = append(out, line{styled("Done. Press any key to go back.", stSel)})
	m.applied = &applied{lines: out}

	st, err := core.Open(sd)
	if err != nil {
		m.modal = errModal(err)
		return
	}
	m.state, m.eff = st, nil
	m.list = core.List{}
	m.fixed, m.edits = nil, nil
	m.opt.Init = false
	// The detection was of the old values file.
	m.det = nil
	if m.cur == detectedTab {
		m.cur = toolsTab
	}
	m.refreshAll()
}

// appliedView gives the result of Apply in the main buffer of the
// terminal, under the output of the sync.
func (m *Model) appliedView() string {
	out := make([]string, len(m.applied.lines))
	for i, l := range m.applied.lines {
		out[i] = l.render()
	}
	return strings.Join(out, "\n") + "\n"
}

// The lines of the review.

// reviewScreen gives the 25 lines of the review.
func (m *Model) reviewScreen() []line {
	out := m.header()
	out = append(out, line{styled(strings.Repeat("─", screenW), stDim)})
	body := m.reviewBody()
	r := m.review
	r.top = min(r.top, max(len(body)-reviewH, 0))
	for i := range reviewH {
		var l line
		if n := r.top + i; n < len(body) {
			l = body[n]
		}
		l = clip(l, screenW-1).pad(screenW - 1)
		out = append(out, append(l, scrollMark(i, reviewH, r.top > 0, r.top+reviewH < len(body))))
	}
	out = append(out, line{styled(strings.Repeat("─", screenW), stDim)})
	keys := "enter apply  f all files  ↑↓ scroll  esc back"
	if len(r.dirty) > 0 {
		keys = "f all files  ↑↓ scroll  esc back"
	}
	return append(out, line{styled(" "+keys, stDim)})
}

// reviewBody gives the lines of the body of the review.
func (m *Model) reviewBody() []line {
	r := m.review
	sd := m.opt.Side
	out := []line{{styled(" Review · "+plural(m.changes(), "change"), stHead)}, {}}
	if len(m.fixed) > 0 {
		out = append(out, line{styled(" Values file", stHead)})
		for _, f := range m.fixed {
			out = append(out, line{styled("   + ", stAdd), text(f)})
		}
	}
	packs, tools, pkgs := m.changeLines()
	for _, s := range []struct {
		title string
		lines []line
	}{{"Packs", packs}, {"Tools", tools}, {"Packages", pkgs}} {
		if len(s.lines) > 0 {
			out = append(out, line{styled(" "+s.title, stHead)})
			out = append(out, s.lines...)
		}
	}
	if len(m.edits) > 0 {
		out = append(out, line{styled(" Answers", stHead)})
		for _, t := range m.edits {
			out = append(out, changeLine("~", stLock, t, "answers: "+strings.Join(m.answerKeys(t), ", ")))
		}
	}
	out = append(out, line{})
	if len(r.dirty) > 0 {
		out = append(out, line{styled(" Not committed", stHead)})
		for _, d := range r.dirty {
			out = append(out, line{styled("   ! ", stLock), text(sd.Show(d)), styled(" has changes that are not committed", stDim)})
		}
		out = append(out, line{styled("   Apply writes nothing. Commit or discard the changes of these targets, then review again.", stLock)}, line{})
	}
	out = append(out, line{styled(" Files", stHead), styled("   same as "+core.CmdName(sd.Mac, "sync")+" --dry-run", stDim)})
	out = append(out, m.fileLines()...)
	out = append(out, line{})
	if !sd.Mac {
		out = append(out, line{styled(" Next: "+nextRepo, stDim)}, line{})
	}
	return append(out, line{text("   "), button("Apply", len(r.dirty) == 0), text("    "), button("Back", false)})
}

// changeLine gives one line of a change: the sign, the name and a note.
func changeLine(sign string, st lipgloss.Style, name, note string) line {
	return line{styled("   "+sign+" ", st), text(fit(name, chgW)), styled(note, stDim)}
}

// changeLines gives the lines of the items of the change list: a pack is
// one line, a tool has its new dependencies under it, and an own package
// is one line.
func (m *Model) changeLines() (packs, tools, pkgs []line) {
	items := m.list.Items
	prev := &core.Effect{Added: map[string]bool{}, Removed: map[string]bool{}}
	for i, it := range items {
		eff, err := m.state.Effect(items[:i+1])
		if err != nil {
			eff = prev
		}
		var added, removed []string
		for _, n := range m.state.Sel.Idx.Names {
			if eff.Added[n] && !prev.Added[n] {
				added = append(added, n)
			}
			if eff.Removed[n] && !prev.Removed[n] {
				removed = append(removed, n)
			}
		}
		prev = eff
		switch {
		case it.Kind == core.AddPack:
			packs = append(packs, changeLine("+", stAdd, strings.Join(it.Names, ", "), "adds "+listOr(added)))
		case it.Kind == core.RemoveTool && it.Pack != "":
			packs = append(packs, changeLine("-", stRemove, it.Pack, "removes "+listOr(removed)))
		case it.Kind == core.AddTool:
			tools = append(tools, m.toolLines("+", stAdd, it.Names, added)...)
		case it.Kind == core.RemoveTool:
			tools = append(tools, m.toolLines("-", stRemove, it.Names, removed)...)
		case it.Kind == core.Replace:
			tools = append(tools, changeLine("-", stRemove, it.Old, "replaced by "+it.New))
			tools = append(tools, changeLine("+", stAdd, it.New, "replaces "+it.Old+"; its migrate script runs first"))
		case it.Kind == core.AddPackage:
			pkgs = append(pkgs, changeLine("+", stAdd, it.Package.Key, it.Package.Table()))
		}
	}
	return packs, tools, pkgs
}

// toolLines gives a line for each tool of <names>, with its answers, and
// the dependencies of <all> that are not in <names> under it.
func (m *Model) toolLines(sign string, st lipgloss.Style, names, all []string) []line {
	var out []line
	for _, n := range names {
		note := ""
		if keys := m.answerKeys(n); len(keys) > 0 {
			note = "answers: " + strings.Join(keys, ", ")
		}
		out = append(out, changeLine(sign, st, n, note))
	}
	for _, d := range all {
		if slices.Contains(names, d) {
			continue
		}
		out = append(out, line{styled("     └─ ", stDim), styled(fit(d, chgW-2), st), styled("dependency", stDim)})
	}
	return out
}

// answerKeys gives the keys of the answers of <t> on the change list.
func (m *Model) answerKeys(t string) []string {
	var out []string
	for _, a := range m.list.Answers {
		if answerTool(a) == t {
			out = append(out, strings.TrimPrefix(a.Name, t+"."))
		}
	}
	return out
}

// listOr gives the names <l> joined, or "nothing".
func listOr(l []string) string {
	if len(l) == 0 {
		return "nothing"
	}
	return strings.Join(l, ", ")
}

// fileLines gives the lines of the files of the plan: the script before the
// writes, the changes of the values file, then the files of the dry run.
// Without f, only the first filesMax show, with the count of the others.
func (m *Model) fileLines() []line {
	r := m.review
	p := r.plan
	sd := m.opt.Side
	if len(p.Pending) > 0 {
		var tools []string
		for _, g := range byTool(p.Pending) {
			tools = append(tools, g[0].Tool)
		}
		return []line{{styled("   The questions of "+strings.Join(tools, ", ")+" have no answer yet. Apply asks them, then plans the files.", stDim)}}
	}
	var out []line
	if b := p.Before; b != nil {
		out = append(out, line{styled("   "+fit("run", opW), stLock), text(b.Shown()), styled("   "+b.Note, stDim)})
	}
	for _, c := range p.Changes {
		op, st, note := "change", stDim, ""
		switch {
		case c.Remove:
			op, st = "remove", stRemove
		case !core.Exists(c.Path):
			op, st, note = "write", stAdd, "new file"
		}
		out = append(out, fileLine(op, st, sd.Show(c.Path), note))
	}
	for _, f := range p.Files {
		op, st, note := "write", stAdd, ""
		switch f.Kind {
		case report.New:
			note = "new file"
		case report.Seeded:
			note = "seed"
		case report.Removed:
			op, st = "remove", stRemove
		}
		out = append(out, fileLine(op, st, sd.Show(f.Path), note))
	}
	if len(out) == 0 {
		return []line{{styled("   No file changes.", stDim)}}
	}
	if !r.all && len(out) > filesMax {
		more := len(out) - filesMax
		out = append(out[:filesMax-1:filesMax-1], append(out[filesMax-1], styled("   + "+itoa(more)+" more", stDim)))
	}
	return out
}

// fileLine gives the line of a file: its kind, its path and a note.
func fileLine(op string, st lipgloss.Style, path, note string) line {
	l := line{styled("   "+fit(op, opW), st), text(path)}
	if note != "" {
		l = append(l, styled("   "+note, stDim))
	}
	return l
}
