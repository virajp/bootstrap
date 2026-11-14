package tui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/virajp/bootstrap/cli/internal/core"
)

// The Detected tab (screens 2a and 2b of the design): each time the user
// opens it, the dialog "Detecting ..." runs core.DetectState. Then the left
// pane lists the categories with a proposal (D5), the right pane the
// proposals of the category with [?], and the bottom pane the evidence of
// the proposal under the cursor. A proposal is not a change: space accepts
// one, and a accepts all of them, the same as detect --add. Each text of a
// tool comes from the data.

// detectedTab is the index of the Detected tab.
const detectedTab = 2

// The category of the own packages of the Mac (D11 c): the last one. Its id
// is no id of the category data.
const (
	ownID    = "\x00own"
	ownTitle = "Own packages"
)

// The frames of the spinner of the dialog "Detecting ...".
const spinFrames = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"

// proposal is one proposal of the detection: a tool to add, a selected tool
// to remove (its result is mismatch), or on the Mac an own package to add.
type proposal struct {
	tool   string
	pkg    core.Package
	isPkg  bool
	remove bool
	found  core.Detected
	cat    string
}

// detRow is a row of the right pane: a proposal with its mark.
type detRow struct {
	toolRow
	prop proposal
}

// The dialog.

// startDetect opens the dialog "Detecting ..." and runs the detection of
// the state in a command. The tab stays the tab of before until the result
// arrives.
func (m *Model) startDetect() tea.Cmd {
	m.detSeq++
	seq, st := m.detSeq, m.state
	m.detecting = true
	m.modal = &modal{
		title:   "Detecting ...",
		width:   56,
		body:    m.detectingBody(),
		buttons: line{styled(" Cancel ", stButton)},
		keys:    "esc cancel and go back to [" + m.tabs[m.cur].name + "]",
		onKey: func(m *Model, k tea.KeyPressMsg) tea.Cmd {
			switch k.String() {
			case "esc", "enter":
				// The result of the detection is dropped (D12 d).
				m.detSeq++
				m.detecting, m.modal = false, nil
			}
			return nil
		},
	}
	return func() tea.Msg {
		d, err := core.DetectState(st)
		return detectMsg{seq: seq, d: d, err: err}
	}
}

// detectingBody gives the body of the dialog "Detecting ...", with the
// frame of the spinner of the tick.
func (m *Model) detectingBody() []line {
	f := []rune(spinFrames)
	what := "Running the detect scripts"
	if m.opt.Side.Mac {
		what = "Scanning the installed packages"
	}
	return []line{{styled(string(f[m.tick%len(f)])+" "+what, stPlain)}}
}

// detected takes the result of a detection: the Detected tab shows it, or a
// dialog shows the problem and the tab of before stays (D13 b). A result
// after esc is dropped.
func (m *Model) detected(msg detectMsg) {
	if msg.seq != m.detSeq || !m.detecting {
		return
	}
	m.detecting, m.modal = false, nil
	if msg.err != nil {
		m.modal = errModal(msg.err)
		return
	}
	m.det = msg.d
	m.cur = detectedTab
	m.tick = 0
	t := m.tabs[detectedTab]
	t.right.cursor, t.right.top = 0, 0
	t.bottomTop = 0
	m.refreshDetected()
	m.refreshTools()
}

// The proposals.

// proposals gives the proposals of the last detection: in the order of the
// tools, each offered tool and each selected tool whose result is
// mismatch, then the own packages.
func (m *Model) proposals() []proposal {
	d := m.det
	if d == nil {
		return nil
	}
	var out []proposal
	for _, t := range d.Tools {
		mf := m.manifest(t)
		if mf == nil {
			continue
		}
		f := d.Found[t]
		switch {
		case f.Result == "mismatch" && d.Selected(t):
			out = append(out, proposal{tool: t, remove: true, found: f, cat: mf.Category})
		case slices.Contains(d.Offers, t):
			out = append(out, proposal{tool: t, found: f, cat: mf.Category})
		}
	}
	for _, p := range d.Unknown {
		out = append(out, proposal{pkg: p, isPkg: true, found: core.Detected{Result: "match", Evidence: []string{p.Key}, Message: "installed"}, cat: ownID})
	}
	return out
}

// detCategories gives the categories of <props>: in the order of the
// categories file, then the others, then the own packages.
func (m *Model) detCategories(props []proposal) []string {
	has := map[string]bool{}
	for _, p := range props {
		has[p.cat] = true
	}
	var out []string
	if c := m.state.Sel.Idx.Categories; c != nil {
		for _, id := range c.IDs {
			if has[id] {
				out = append(out, id)
			}
		}
	}
	for _, p := range props {
		if p.cat != ownID && !slices.Contains(out, p.cat) {
			out = append(out, p.cat)
		}
	}
	if has[ownID] {
		out = append(out, ownID)
	}
	return out
}

// detTitle gives the title of the category <id> of the Detected tab.
func (m *Model) detTitle(id string) string {
	if id == ownID {
		return ownTitle
	}
	return m.categoryTitle(id)
}

// pkgItem gives the index of the change of the own package <p> on the
// list, or -1.
func (m *Model) pkgItem(p core.Package) int {
	for i, it := range m.list.Items {
		if it.Kind == core.AddPackage && it.Package == p {
			return i
		}
	}
	return -1
}

// propMark gives the mark of the proposal <p>, its style, and if it is on
// the change list: [?] when it is not (green to add, red to remove), else
// [+] or [-] from the change list (D12 d).
func (m *Model) propMark(p proposal) (string, lipgloss.Style, bool) {
	if p.isPkg {
		if m.pkgItem(p.pkg) >= 0 {
			return "[+]", stAdd, true
		}
		return "[?]", stAdd, false
	}
	mk, _ := m.markOf(p.tool)
	if p.remove {
		if mk == markRemove {
			return "[-]", stRemove, true
		}
		return "[?]", stRemove, false
	}
	if mk == markAdd {
		return "[+]", stAdd, true
	}
	return "[?]", stAdd, false
}

// propRow gives the row of the proposal <p>.
func (m *Model) propRow(p proposal) detRow {
	mk, st, _ := m.propMark(p)
	r := toolRow{tool: p.pkg.Key, summary: "no Mac tool has it", markSt: stPlain, nameSt: stPlain, sumSt: stPlain}
	if !p.isPkg {
		r = m.toolRow(p.tool, false)
	}
	r.mark, r.markSt = mk, st
	return detRow{toolRow: r, prop: p}
}

// refreshDetected builds the panes of the Detected tab again from the last
// detection and the change list, and the ◆ hints of the Tools tab: a
// category with a proposal that is not on the change list. The cursors stay
// where they are.
func (m *Model) refreshDetected() {
	if m.state == nil {
		return
	}
	t := m.tabs[detectedTab]
	props := m.proposals()
	byCat := map[string][]proposal{}
	open := map[string]bool{}
	for _, p := range props {
		byCat[p.cat] = append(byCat[p.cat], p)
		if _, _, on := m.propMark(p); !on {
			open[p.cat] = true
		}
	}
	m.hints = map[string]bool{}
	var rows []row
	for _, id := range m.detCategories(props) {
		count := m.countOf(m.inCategory(id))
		if id == ownID {
			count = itoa(len(byCat[id])-m.openCount(byCat[id])) + "/" + itoa(len(byCat[id]))
		} else if open[id] {
			m.hints[id] = true
		}
		rows = append(rows, catRow{id: id, title: m.detTitle(id), count: count, hint: open[id]})
	}
	t.left.rows = rows
	t.left.move(0)
	t.right.title, t.right.rows = "", nil
	if len(props) == 0 {
		t.right.rows = []row{textRow{" No tool detected", stDim}}
		for _, l := range m.noteLines(rightW - 2) {
			t.right.rows = append(t.right.rows, textRow{" " + l, stDim})
		}
	} else if c, ok := m.curDetCat(); ok {
		t.right.title = c.title + " · detected"
		for _, p := range byCat[c.id] {
			t.right.rows = append(t.right.rows, m.propRow(p))
			if !p.isPkg {
				t.right.rows = append(t.right.rows, m.tree(p.tool, "", map[string]bool{p.tool: true})...)
			}
		}
	}
	t.right.move(0)
	t.bottom = m.detectedBottom()
}

// openCount gives the count of the proposals of <props> that are not on the
// change list.
func (m *Model) openCount(props []proposal) int {
	n := 0
	for _, p := range props {
		if _, _, on := m.propMark(p); !on {
			n++
		}
	}
	return n
}

// curDetCat gives the category under the cursor of the left pane.
func (m *Model) curDetCat() (catRow, bool) {
	l := m.tabs[detectedTab].left
	if l.cursor < len(l.rows) {
		c, ok := l.rows[l.cursor].(catRow)
		return c, ok
	}
	return catRow{}, false
}

// curDetRow gives the row under the cursor of the right pane.
func (m *Model) curDetRow() row {
	r := m.tabs[detectedTab].right
	if r.cursor < len(r.rows) {
		return r.rows[r.cursor]
	}
	return nil
}

// The bottom pane.

// detectedBottom gives the lines of the bottom pane: on a category, its
// count of proposals and the notes of the package scan (D19); on a proposal, its evidence (ruling 16); on a
// dependency, the details of the tool.
func (m *Model) detectedBottom() []line {
	if !m.tabs[detectedTab].onRight {
		c, ok := m.curDetCat()
		if !ok {
			return nil
		}
		var props []proposal
		for _, p := range m.proposals() {
			if p.cat == c.id {
				props = append(props, p)
			}
		}
		on := len(props) - m.openCount(props)
		out := []line{
			{styled(" "+c.title, stHead)},
			{text(" " + plural(len(props), "proposal") + " of the detection, " + itoa(on) + " on the change list.")},
		}
		for _, l := range m.noteLines(screenW - 2) {
			out = append(out, line{styled(" "+l, stDim)})
		}
		return out
	}
	switch r := m.curDetRow().(type) {
	case detRow:
		return m.proposalDetail(r.prop)
	case toolRow:
		if r.glyphs != "" {
			return m.toolDetail(r.tool)
		}
	}
	return nil
}

// noteLines gives the notes of the package scan of the last detection
// (D19), for example a package manager that is not installed, wrapped to
// the width <w>.
func (m *Model) noteLines(w int) []string {
	if m.det == nil {
		return nil
	}
	var out []string
	for _, n := range m.det.Notes {
		out = append(out, wrap(n, w)...)
	}
	return out
}

// proposalDetail gives the name, the description, the proposal and the
// evidence of the proposal <p>, and its command when it is on the list.
func (m *Model) proposalDetail(p proposal) []line {
	_, _, on := m.propMark(p)
	action, sign, st := "add", "+", stAdd
	if p.remove {
		action, sign, st = "remove", "-", stRemove
	}
	prop := action + " (space: accept)"
	if on {
		prop = "on the change list (space: undo)"
	}
	if p.isPkg {
		out := []line{{styled(" "+p.pkg.Key, stHead)}, {text(" An installed package that no Mac tool has.")}, {}}
		out = append(out, pairLines([][2]string{{"Proposal", prop}, {"Category", ownTitle}, {"Table", p.pkg.Table()}, {"Evidence", p.pkg.Key}})...)
		if on {
			return append(out, line{styled(" Same as: "+core.CmdName(true, "detect")+" --add", stDim)})
		}
		return append(out, line{styled(" "+sign+" Space adds it to "+p.pkg.Table()+" of macos.toml.", st)})
	}
	t := p.tool
	mk, _ := m.markOf(t)
	locked := p.remove && mk == markLocked
	if locked {
		prop = action + " (not possible)"
	}
	out := []line{{styled(" "+t, stHead)}}
	for _, l := range wrap(m.manifest(t).Description, screenW-3) {
		out = append(out, line{text(" " + l)})
	}
	out = append(out, line{})
	ev := strings.Join(p.found.Evidence, ", ")
	if ev == "" {
		ev = "none"
	}
	out = append(out, pairLines([][2]string{{"Proposal", prop}, {"Category", m.categoryTitle(m.manifest(t).Category)}, {"Detected", p.found.Result}, {"Evidence", ev}})...)
	switch {
	case locked:
		return append(out, line{styled(" ! "+t+" is required by the pack "+m.requiredBy(t)+". You cannot remove it.", stLock)})
	case on:
		if same := m.sameAs(t); same != "" {
			out = append(out, line{styled(" Same as: "+same, stDim)})
		}
		return out
	}
	msg := sign + " "
	if p.found.Message != "" {
		msg += p.found.Message + ". "
	}
	return append(out, line{styled(" "+msg+"Press space to accept.", st)})
}

// The keys of the tab.

// detectedKey handles a key of the Detected tab that the main screen does
// not handle. It tells if it handled the key.
func (m *Model) detectedKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if m.state == nil {
		return nil, false
	}
	switch msg.String() {
	case "space":
		if !m.tabs[detectedTab].onRight {
			return nil, true
		}
		return m.detSpace(), true
	case "a":
		return m.acceptAll(), true
	}
	return nil, false
}

// afterDetectedKey follows a key of the main screen on the Detected tab: a
// move in the left pane shows the proposals of the new category from the
// top, and the bottom pane shows the item under the cursor.
func (m *Model) afterDetectedKey(msg tea.KeyPressMsg) {
	t := m.tabs[detectedTab]
	switch {
	case key.Matches(msg, m.keys.Move):
		if !t.onRight {
			t.right.cursor, t.right.top = 0, 0
		}
	case key.Matches(msg, m.keys.Pane):
	default:
		return
	}
	t.bottomTop = 0
	m.refreshDetected()
}

// detSpace accepts the proposal under the cursor, or takes it off the list
// when it is on it. A dependency row does nothing, and so does the removal
// of a required tool.
func (m *Model) detSpace() tea.Cmd {
	r, ok := m.curDetRow().(detRow)
	if !ok {
		return nil
	}
	p := r.prop
	if p.isPkg {
		if i := m.pkgItem(p.pkg); i >= 0 {
			m.setItems(slices.Delete(slices.Clone(m.list.Items), i, i+1))
			return nil
		}
		m.push(m.detItems(nil, []core.Package{p.pkg})...)
		return nil
	}
	mk, _ := m.markOf(p.tool)
	switch {
	case mk == markAdd || mk == markRemove:
		return m.undo(p.tool, mk)
	case p.remove && mk == markLocked:
		return nil
	case p.remove:
		return m.unselect(p.tool)
	}
	return m.selectHook(core.Item{Kind: core.AddTool, Names: []string{p.tool}})
}

// acceptAll puts each proposal to add that is not on the list on it: the
// tools and the own packages, the same as detect --add. A proposal to
// remove stays a proposal.
func (m *Model) acceptAll() tea.Cmd {
	var tools []string
	var pkgs []core.Package
	for _, p := range m.proposals() {
		if _, _, on := m.propMark(p); on || p.remove {
			continue
		}
		if p.isPkg {
			pkgs = append(pkgs, p.pkg)
		} else {
			tools = append(tools, p.tool)
		}
	}
	items := m.detItems(tools, pkgs)
	if len(items) == 0 {
		return nil
	}
	if items[0].Kind != core.AddTool {
		m.push(items...)
		return nil
	}
	if len(items) > 1 && !m.push(items[1:]...) {
		return nil
	}
	return m.selectHook(items[0])
}

// detItems gives the items that add the tools <tools> and the own packages
// <pkgs>, as detect --add does. The tool that core adds for an App Store
// app stays out when it is selected already.
func (m *Model) detItems(tools []string, pkgs []core.Package) []core.Item {
	items := m.det.Items(tools, pkgs)
	if len(items) == 0 || items[0].Kind != core.AddTool {
		return items
	}
	names := slices.DeleteFunc(slices.Clone(items[0].Names), func(n string) bool {
		return !slices.Contains(tools, n) && m.selected(n)
	})
	if len(names) == 0 {
		return items[1:]
	}
	items[0].Names = names
	return items
}
