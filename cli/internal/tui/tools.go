package tui

import (
	"slices"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// The Tools tab (screens 1, 3, 4, 5 and 6 of the design): the categories
// in the left pane, the tools of the category in the right pane with the
// tree of their dependencies, and the details of the item under the cursor
// in the bottom pane. Each text of a tool comes from the data.

// The widths of the columns of a row (screens.txt, "The fixed columns").
const (
	catW     = 19 // The name of a category.
	countW   = 5  // The counter of a category.
	nameW    = 14 // The name of a tool.
	summaryW = 47 // The summary of a tool.
	kvKeyW   = 11 // The key of a detail of the bottom pane.
	kvValW   = 37 // The value of the first detail of a line.
)

// The text of a group of alternatives (ruling 28).
const groupText = " · select one or none"

// toolsTab is the index of the Tools tab.
const toolsTab = 0

// The mark of a tool row.
type mark int

const (
	markOff     mark = iota // [ ]: not selected, or only a dependency.
	markOn                  // [x]: selected.
	markAdd                 // [+]: the change list adds it.
	markRemove              // [-]: the change list removes it.
	markLocked              // [=]: required.
	markNotHere             // [·]: not available.
)

// catRow is a row of the left pane: a category and its counter.
type catRow struct {
	id, title, count string
	hint             bool // Detection proposes a change (the Detected tab fills it).
}

func (r catRow) render(_ int, cursor bool, tick int) line {
	cur := "  "
	if cursor {
		cur = "▸ "
	}
	hint := text(" ")
	if r.hint {
		hint = styled("◆", stHint)
	}
	return line{text(cur), text(cell(r.title, catW, tick)), text(" " + rightAlign(r.count, countW) + " "), hint}
}

// textRow is a row of plain text, for example the search field.
type textRow struct {
	s  string
	st lipgloss.Style
}

func (r textRow) render(w int, _ bool, _ int) line { return line{styled(fit(r.s, w), r.st)} }

// toolRow is a row of the right pane: a tool with its selector, a
// dependency of the tree above it (glyphs, no selector), or the label of a
// group of alternatives.
type toolRow struct {
	tool    string
	glyphs  string // The tree glyphs of a dependency row.
	group   string // The parent of a group row.
	mark    string
	markSt  lipgloss.Style
	nameSt  lipgloss.Style
	summary string
	sumSt   lipgloss.Style
}

func (r toolRow) render(w int, _ bool, tick int) line {
	if r.group != "" {
		return line{styled(" "+fit(r.group+groupText, w-1), stDim)}
	}
	sum := styled(fit(r.summary, summaryW), r.sumSt)
	if r.glyphs != "" {
		return line{text("     "), styled(cell(r.glyphs+r.tool, nameW, tick), r.nameSt), text(" "), sum, text(" ")}
	}
	return line{text(" "), styled(r.mark, r.markSt), text(" "), styled(cell(r.tool, nameW, tick), r.nameSt), text(" "), sum, text(" ")}
}

// The tools of the data.

func (m *Model) manifest(t string) *source.Manifest { return m.state.Manifest(t) }

// listed gives the tools that have a row with a selector: the selectable
// tools, in the order of their names.
func (m *Model) listed() []string {
	var out []string
	for _, n := range m.state.Sel.Idx.Names {
		if m.manifest(n).Selectable {
			out = append(out, n)
		}
	}
	return out
}

// categories gives the categories with at least one listed tool, in the
// order of the categories file.
func (m *Model) categories() []string {
	has := map[string]bool{}
	for _, t := range m.listed() {
		has[m.manifest(t).Category] = true
	}
	var out []string
	if c := m.state.Sel.Idx.Categories; c != nil {
		for _, id := range c.IDs {
			if has[id] {
				out = append(out, id)
			}
		}
	}
	return out
}

// categoryTitle gives the title of the category <id>.
func (m *Model) categoryTitle(id string) string {
	if c := m.state.Sel.Idx.Categories; c != nil {
		if t, ok := c.Titles[id]; ok {
			return t
		}
	}
	return id
}

// The selection after the change list.

func (m *Model) effect() *core.Effect {
	if m.eff == nil {
		return &core.Effect{}
	}
	return m.eff
}

// selected tells if <t> is selected after the change list.
func (m *Model) selected(t string) bool {
	e := m.effect()
	return e.Added[t] || (!e.Removed[t] && m.state.Selected(t))
}

// direct tells if the values file selects <t> by its table or by a pack,
// not only as a dependency.
func (m *Model) direct(t string) bool { return m.state.Named[t] || m.state.PackOf(t) != "" }

// requiredBy gives the pack where <t> is required, or "".
func (m *Model) requiredBy(t string) string {
	var keys []string
	for k, v := range m.state.Packs.Required {
		if v && strings.HasSuffix(k, "/"+t) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)
	return strings.TrimSuffix(keys[0], "/"+t)
}

// notHere gives why <t> is not available here, or "".
func (m *Model) notHere(t string) string {
	mf := m.manifest(t)
	if host := source.HostName(); !mf.RunsOn(host) {
		return mf.HostMessage(host)
	}
	if !m.opt.Side.Mac && mf.RootOnly && core.InSubmodule(m.opt.Side.Root) {
		return t + " is only for the root repository, and this repository is a submodule"
	}
	return ""
}

// markOf gives the mark of the tool <t> and the text of its status.
func (m *Model) markOf(t string) (mark, string) {
	e := m.effect()
	switch {
	case e.Removed[t] && m.state.Selected(t):
		return markRemove, "will remove"
	case e.Direct[t] && !m.direct(t):
		return markAdd, "will add"
	case !m.selected(t) && m.notHere(t) != "":
		return markNotHere, "not available"
	case m.selected(t) && m.requiredBy(t) != "":
		return markLocked, "required"
	case m.direct(t) || e.Direct[t]:
		return markOn, "selected"
	case e.Added[t]:
		return markOff, "will add as a dependency"
	case m.selected(t):
		return markOff, "selected as a dependency"
	}
	return markOff, "not selected"
}

// alternatives gives the listed tools with the parent of <t>, when the
// parent has two or more of them.
func (m *Model) alternatives(t string) []string {
	p := m.manifest(t).Parent
	if p == "" {
		return nil
	}
	var out []string
	for _, n := range m.listed() {
		if m.manifest(n).Parent == p {
			out = append(out, n)
		}
	}
	if len(out) < 2 {
		return nil
	}
	return out
}

// The rows of the tab.

// toolRows gives the rows of the tools <tools>: each tool with the tree of
// its dependencies, and the tools of a group of alternatives together under
// the label of the group.
func (m *Model) toolRows(tools []string) []row {
	byParent := map[string][]string{}
	for _, t := range tools {
		if p := m.manifest(t).Parent; p != "" {
			byParent[p] = append(byParent[p], t)
		}
	}
	var rows []row
	done := map[string]bool{}
	for _, t := range tools {
		if done[t] {
			continue
		}
		p := m.manifest(t).Parent
		if g := byParent[p]; p != "" && len(g) > 1 {
			rows = append(rows, toolRow{group: p})
			for _, n := range g {
				done[n] = true
				rows = append(rows, m.toolRow(n, true))
				rows = append(rows, m.tree(n, "", map[string]bool{n: true})...)
			}
			continue
		}
		done[t] = true
		rows = append(rows, m.toolRow(t, false))
		rows = append(rows, m.tree(t, "", map[string]bool{t: true})...)
	}
	return rows
}

// toolRow gives the row of the tool <t>; <radio> is a tool of a group of
// alternatives.
func (m *Model) toolRow(t string, radio bool) toolRow {
	r := toolRow{tool: t, summary: m.manifest(t).Summary, markSt: stPlain, nameSt: stPlain, sumSt: stPlain}
	mk, _ := m.markOf(t)
	switch mk {
	case markOn:
		r.mark, r.markSt = "[x]", stSel
		if radio {
			r.mark = "(•)"
		}
	case markAdd:
		r.mark, r.markSt = "[+]", stAdd
	case markRemove:
		r.mark, r.markSt = "[-]", stRemove
	case markLocked:
		r.mark, r.markSt = "[=]", stLock
	case markNotHere:
		r.mark, r.markSt, r.nameSt, r.sumSt = "[·]", stDim, stDim, stDim
	default:
		r.mark = "[ ]"
		if radio {
			r.mark = "( )"
		}
	}
	return r
}

// tree gives the dependency rows of <t>, each dependency under each tool
// that needs it (ruling 8). <path> holds the tools above, against a loop.
func (m *Model) tree(t, prefix string, path map[string]bool) []row {
	deps := m.manifest(t).Depends
	var rows []row
	for i, d := range deps {
		if path[d] || m.manifest(d) == nil {
			continue
		}
		glyph, next := "├─ ", "│  "
		if i == len(deps)-1 {
			glyph, next = "└─ ", "   "
		}
		st := stDep
		switch e := m.effect(); {
		case e.Removed[d]:
			st = stRemove
		case e.Added[d]:
			st = stAdd
		}
		rows = append(rows, toolRow{tool: d, glyphs: prefix + glyph, nameSt: st, summary: m.manifest(d).Summary, sumSt: stDim})
		path[d] = true
		rows = append(rows, m.tree(d, prefix+next, path)...)
		delete(path, d)
	}
	return rows
}

// countOf gives the counter of the tools <tools>: the selected ones and all.
func (m *Model) countOf(tools []string) string {
	n := 0
	for _, t := range tools {
		if mk, _ := m.markOf(t); mk == markOn || mk == markAdd || mk == markLocked {
			n++
		}
	}
	return itoa(n) + "/" + itoa(len(tools))
}

// inCategory gives the listed tools of the category <id>.
func (m *Model) inCategory(id string) []string {
	var out []string
	for _, t := range m.listed() {
		if m.manifest(t).Category == id {
			out = append(out, t)
		}
	}
	return out
}

// matches gives the listed tools whose name, summary, description or
// category has <q>, without regard to case (D10 f).
func (m *Model) matches(q string) []string {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}
	var out []string
	for _, t := range m.listed() {
		mf := m.manifest(t)
		for _, s := range []string{t, mf.Summary, mf.Description, m.categoryTitle(mf.Category)} {
			if strings.Contains(strings.ToLower(s), q) {
				out = append(out, t)
				break
			}
		}
	}
	return out
}

// refreshTools builds the panes of the Tools tab again from the state and
// the change list. The cursors stay where they are.
func (m *Model) refreshTools() {
	if m.state == nil {
		return
	}
	t := m.tabs[toolsTab]
	if m.search {
		t.left.title = "Search"
		t.left.rows = []row{textRow{" / " + m.query + "▏", stHead}, textRow{"", stPlain}, textRow{" " + plural(len(m.matches(m.query)), "result"), stDim}}
		t.left.cursor, t.left.top = 0, 0
		t.right.title = "Results"
		if m.query != "" {
			t.right.title = "Results for \"" + m.query + "\""
		}
		t.right.rows = m.toolRows(m.matches(m.query))
		t.onRight = true
	} else {
		t.left.title = "Categories"
		var rows []row
		for _, id := range m.categories() {
			rows = append(rows, catRow{id: id, title: m.categoryTitle(id), count: m.countOf(m.inCategory(id)), hint: m.hints[id]})
		}
		t.left.rows = rows
		t.left.move(0)
		t.right.title = ""
		t.right.rows = nil
		if c, ok := m.curCategory(); ok {
			t.right.title = c.title
			t.right.rows = m.toolRows(m.inCategory(c.id))
		}
	}
	t.right.move(0)
	t.bottom = m.toolsBottom()
}

// curCategory gives the category under the cursor of the left pane.
func (m *Model) curCategory() (catRow, bool) {
	l := m.tabs[toolsTab].left
	if l.cursor < len(l.rows) {
		c, ok := l.rows[l.cursor].(catRow)
		return c, ok
	}
	return catRow{}, false
}

// curTool gives the row under the cursor of the right pane.
func (m *Model) curTool() (toolRow, bool) {
	r := m.tabs[toolsTab].right
	if r.cursor < len(r.rows) {
		t, ok := r.rows[r.cursor].(toolRow)
		return t, ok
	}
	return toolRow{}, false
}

// The bottom pane.

// toolsBottom gives the lines of the bottom pane: the item under the cursor.
func (m *Model) toolsBottom() []line {
	t := m.tabs[toolsTab]
	if !t.onRight {
		c, ok := m.curCategory()
		if !ok {
			return nil
		}
		n := m.inCategory(c.id)
		return []line{
			{styled(" "+c.title, stHead)},
			{text(" " + strings.Replace(m.countOf(n), "/", " of ", 1) + " tools are selected.")},
		}
	}
	r, ok := m.curTool()
	switch {
	case !ok:
		return nil
	case r.group != "":
		return []line{
			{styled(" "+r.group+groupText, stHead)},
			{text(" The tools with the parent " + r.group + " are alternatives: select one of them, or none.")},
		}
	}
	return m.toolDetail(r.tool)
}

// toolDetail gives the name, the description, the dependencies and the
// status of the tool <t>, with the reason of a locked row, or the command
// of its change (ruling 54).
func (m *Model) toolDetail(t string) []line {
	mf := m.manifest(t)
	out := []line{{styled(" "+t, stHead)}}
	for _, l := range wrap(mf.Description, screenW-3) {
		out = append(out, line{text(" " + l)})
	}
	out = append(out, line{})
	_, status := m.markOf(t)
	pairs := [][2]string{{"Status", status}, {"Category", m.categoryTitle(mf.Category)}}
	needs := "nothing"
	if len(mf.Depends) > 0 {
		var l []string
		for _, d := range mf.Depends {
			if dm := m.manifest(d); dm != nil && !dm.Selectable {
				d += " (not selectable)"
			}
			l = append(l, d)
		}
		needs = strings.Join(l, ", ")
	}
	pairs = append(pairs, [2]string{"Needs", needs}, [2]string{"Packs", m.packsOf(t)})
	if mf.Parent != "" {
		pairs = append(pairs, [2]string{"Parent", mf.Parent})
	}
	if mf.Raw.Has("hosts") {
		pairs = append(pairs, [2]string{"Hosts", strings.Join(mf.Hosts, ", ")})
	}
	if len(mf.DefaultEnvs) > 0 {
		pairs = append(pairs, [2]string{"Envs", strings.Join(mf.DefaultEnvs, ", ")})
	}
	for i := 0; i < len(pairs); i += 2 {
		l := line{styled(" "+fit(pairs[i][0], kvKeyW), stDim), text(fit(pairs[i][1], kvValW))}
		if i+1 < len(pairs) {
			l = append(l, styled(" "+fit(pairs[i+1][0], kvKeyW), stDim), text(pairs[i+1][1]))
		}
		out = append(out, l)
	}
	mk, _ := m.markOf(t)
	switch {
	case mk == markLocked:
		out = append(out, line{styled(" ! "+t+" is required by the pack "+m.requiredBy(t)+". You cannot remove it.", stLock)})
	case mk == markNotHere:
		out = append(out, line{styled(" · "+m.notHere(t)+".", stDim)})
	}
	if same := m.sameAs(t); same != "" {
		out = append(out, line{styled(" Same as: "+same, stDim)})
	}
	return out
}

// packsOf gives the packs that have <t> as a member, with "(required)".
func (m *Model) packsOf(t string) string {
	p := m.state.Packs
	var out []string
	for _, n := range p.Names {
		if !p.HasMember(n, t) {
			continue
		}
		if p.Required[n+"/"+t] {
			n += " (required)"
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return "none"
	}
	return strings.Join(out, ", ")
}

// sameAs gives the command of the last change of the list that names <t>.
func (m *Model) sameAs(t string) string {
	if i := m.itemOf(t); i >= 0 {
		return m.cliOf(m.list.Items[i])
	}
	return ""
}

// cliOf gives the command of the CLI that makes the change <it>.
func (m *Model) cliOf(it core.Item) string {
	mac := m.opt.Side.Mac
	switch it.Kind {
	case core.AddTool:
		return core.CmdName(mac, "add tool") + " " + strings.Join(it.Names, " ")
	case core.AddPack:
		return core.CmdName(mac, "add pack") + " " + strings.Join(it.Names, " ")
	case core.RemoveTool:
		return core.CmdName(mac, "remove tool") + " " + strings.Join(it.Names, " ")
	case core.Replace:
		return core.CmdName(mac, "replace tool") + " " + it.Old + " " + it.New
	}
	return ""
}

// itemOf gives the index of the last change of the list that names <t>, or
// -1.
func (m *Model) itemOf(t string) int {
	for i := len(m.list.Items) - 1; i >= 0; i-- {
		it := m.list.Items[i]
		if (it.Kind != core.AddPack && slices.Contains(it.Names, t)) || it.Old == t || it.New == t {
			return i
		}
	}
	return -1
}

// The keys of the tab.

// toolsKey handles a key of the Tools tab that the main screen does not
// handle. It tells if it handled the key.
func (m *Model) toolsKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if m.state == nil {
		return nil, false
	}
	t := m.tabs[toolsTab]
	if m.search {
		switch k := msg.String(); {
		case k == "esc":
			m.search, m.query = false, ""
			t.onRight = false
		case k == "backspace":
			if r := []rune(m.query); len(r) > 0 {
				m.query = string(r[:len(r)-1])
			}
			t.right.cursor, t.right.top = 0, 0
		case k == "space":
			return m.space(), true
		case k == "up" || k == "down" || k == "shift+up" || k == "shift+down" || k == "ctrl+c":
			return nil, false
		case msg.Text != "":
			m.query += msg.Text
			t.right.cursor, t.right.top = 0, 0
		default:
			return nil, true
		}
		m.tick = 0
		t.bottomTop = 0
		m.refreshTools()
		return nil, true
	}
	switch k := msg.String(); k {
	case "/":
		m.search, m.query = true, ""
		t.right.cursor, t.right.top = 0, 0
		t.bottomTop = 0
		m.refreshTools()
		return nil, true
	case "space":
		if !t.onRight {
			return nil, true
		}
		return m.space(), true
	}
	return nil, false
}

// afterMove follows a move of the cursor or of the pane: a new category
// starts at the top of its tools, and the bottom pane shows the new item.
func (m *Model) afterMove(leftMoved bool) {
	t := m.tabs[toolsTab]
	if leftMoved && !m.search {
		t.right.cursor, t.right.top = 0, 0
	}
	t.bottomTop = 0
	m.refreshTools()
}

// space selects or unselects the tool under the cursor. A dependency row, a
// group row, a locked row and a row that is not available do nothing.
func (m *Model) space() tea.Cmd {
	r, ok := m.curTool()
	if !ok || r.group != "" || r.glyphs != "" {
		return nil
	}
	t := r.tool
	switch mk, _ := m.markOf(t); mk {
	case markLocked, markNotHere:
		return nil
	case markAdd, markRemove:
		return m.undo(t, mk)
	case markOn:
		return m.unselect(t)
	}
	return m.choose(t)
}

// choose selects <t>: in a group of alternatives with another selected
// tool, it asks to replace that tool (ruling 28).
func (m *Model) choose(t string) tea.Cmd {
	for _, o := range m.alternatives(t) {
		if o != t && m.selected(o) {
			m.modal = m.replaceModal(o, t)
			return nil
		}
	}
	return m.selectHook(core.Item{Kind: core.AddTool, Names: []string{t}})
}

// selectHook adds the select <it>. The questions form of a tool with
// questions opens here (U5); for now the change goes on the list with no
// form.
func (m *Model) selectHook(it core.Item) tea.Cmd {
	m.push(it)
	return nil
}

// push puts <items> at the end of the change list. When core refuses the
// list, the list stays as it was, and a dialog shows why.
func (m *Model) push(items ...core.Item) bool {
	return m.setItems(append(slices.Clone(m.list.Items), items...))
}

// setItems sets the items of the change list, after core accepts them.
func (m *Model) setItems(items []core.Item) bool {
	eff, err := m.state.Effect(items)
	if err != nil {
		m.modal = errModal(err)
		return false
	}
	m.list.Items, m.eff = items, eff
	m.refreshTools()
	return true
}

// undo takes the change of <t> off the list: a tool that it adds, or a
// removal. A tool that a pack adds, or a dependency that a removal takes,
// gets a change of its own.
func (m *Model) undo(t string, mk mark) tea.Cmd {
	i := m.itemOf(t)
	if i < 0 {
		if mk == markAdd {
			return m.unselect(t)
		}
		m.push(core.Item{Kind: core.AddTool, Names: []string{t}})
		return nil
	}
	items := slices.Clone(m.list.Items)
	it := items[i]
	if it.Kind == core.AddTool && len(it.Names) > 1 {
		it.Names = slices.DeleteFunc(slices.Clone(it.Names), func(n string) bool { return n == t })
		items[i] = it
	} else {
		items = slices.Delete(items, i, i+1)
	}
	m.setItems(items)
	return nil
}

// unselect removes <t>. When the removal takes other tools, the cascade
// dialog asks first; when a required tool needs <t>, a dialog says why it
// cannot (screen 4).
func (m *Model) unselect(t string) tea.Cmd {
	var deps, locked []string
	for _, n := range m.state.Sel.Idx.Names {
		if n == t || !m.selected(n) || !m.state.Closure([]string{n})[t] {
			continue
		}
		if m.requiredBy(n) != "" {
			locked = append(locked, n)
		}
		deps = append(deps, n)
	}
	if len(locked) > 0 {
		m.modal = m.lockedModal(t, locked)
		return nil
	}
	it := core.Item{Kind: core.RemoveTool, Names: append([]string{t}, deps...)}
	items := append(slices.Clone(m.list.Items), it)
	eff, err := m.state.Effect(items)
	if err != nil {
		m.modal = errModal(err)
		return nil
	}
	var extra []string
	for _, n := range m.state.Sel.Idx.Names {
		if eff.Removed[n] && !m.effect().Removed[n] && !slices.Contains(it.Names, n) && m.selected(n) {
			extra = append(extra, n)
		}
	}
	if len(deps) == 0 && len(extra) == 0 {
		m.setItems(items)
		return nil
	}
	m.modal = m.cascadeModal(t, deps, extra, items)
	return nil
}

// The dialogs of the tab.

// cascadeModal asks to remove <t> with the tools <deps> that need it, and
// the dependencies <extra> that no tool needs after the change.
func (m *Model) cascadeModal(t string, deps, extra []string, items []core.Item) *modal {
	gone := append([]string{t}, deps...)
	var body []line
	if len(deps) > 0 {
		body = append(body, line{text("These selected tools need " + t + ". They leave too:")}, line{})
		for _, d := range deps {
			need := t
			for _, x := range m.manifest(d).Depends {
				if slices.Contains(gone, x) {
					need = x
					break
				}
			}
			body = append(body, line{styled("- ", stRemove), text(fit(d, nameW)), styled("needs "+need, stDim)})
		}
	}
	if len(extra) > 0 {
		if len(body) > 0 {
			body = append(body, line{})
		}
		body = append(body, line{text("No other tool needs these dependencies. They leave too:")}, line{})
		for _, d := range extra {
			body = append(body, line{styled("- ", stRemove), text(fit(d, nameW)), styled("a dependency of "+t, stDim)})
		}
	}
	n := len(deps) + len(extra)
	body = append(body, line{}, line{text("Remove " + t + " and " + these(n) + "? "), styled("[y/N]", stHead)})
	return &modal{
		title: "Remove " + t + "?",
		warn:  true,
		width: 64,
		body:  body,
		keys:  "y remove the " + itoa(n+1) + " tools  any other key keeps " + t,
		onKey: func(m *Model, k tea.KeyPressMsg) tea.Cmd {
			m.modal = nil
			if k.String() == "y" {
				m.setItems(items)
			}
			return nil
		},
	}
}

// these gives "this tool" or "these <n> tools".
func these(n int) string {
	if n == 1 {
		return "this tool"
	}
	return "these " + itoa(n) + " tools"
}

// lockedModal says that <t> cannot leave: the required tools <locked> need
// it. Any key closes it (ruling 26).
func (m *Model) lockedModal(t string, locked []string) *modal {
	body := []line{{text(t + " is necessary for a required tool:")}, {}}
	for _, n := range locked {
		body = append(body, line{styled("= ", stLock), text(fit(n, nameW)), styled("required by the pack "+m.requiredBy(n), stDim)})
	}
	body = append(body, line{}, line{styled("Press any key to close.", stDim)})
	return &modal{title: "Cannot remove " + t, warn: true, width: 64, body: body, keys: "any key closes"}
}

// replaceModal asks to replace <old> with <new> of the same group. With no
// migrate script, it offers the long way: remove <old> and add <new>
// (ruling 27).
func (m *Model) replaceModal(old, new string) *modal {
	d := &modal{title: "Replace " + old + " with " + new + "?", warn: true, width: 72}
	r := m.state.Runner()
	if r.Has(new, source.MigratePrefix+old) {
		d.body = []line{
			{text("The migrate script of " + new + " for " + old + " runs at apply.")},
			{},
			{styled("Same as: "+m.cliOf(core.Item{Kind: core.Replace, Old: old, New: new}), stDim)},
			{},
			{text("Replace " + old + " with " + new + "? "), styled("[y/N]", stHead)},
		}
		d.keys = "y replace  any other key cancels"
		d.onKey = func(m *Model, k tea.KeyPressMsg) tea.Cmd {
			m.modal = nil
			if k.String() == "y" {
				m.push(core.Item{Kind: core.Replace, Old: old, New: new})
			}
			return nil
		}
		return d
	}
	d.body = []line{{text(new + " has no migrate script for " + old + ".")}, {}}
	oldLine := line{styled("- ", stRemove), text(fit(old, nameW))}
	if s := m.seeds(old, true); len(s) > 0 {
		oldLine = append(oldLine, styled("its seed files stay: "+strings.Join(s, ", "), stDim))
	}
	newLine := line{styled("+ ", stAdd), text(fit(new, nameW))}
	if s := m.seeds(new, false); len(s) > 0 {
		newLine = append(newLine, styled("its seed files are new: "+strings.Join(s, ", "), stDim))
	}
	d.body = append(d.body, oldLine, newLine)
	if len(m.seeds(old, true)) > 0 {
		d.body = append(d.body, line{}, line{styled("Move the content of the seed files yourself.", stDim)})
	}
	d.body = append(d.body, line{}, line{text("Remove " + old + " and add " + new + "? "), styled("[y/N]", stHead)})
	d.keys = "y remove " + old + " and add " + new + "  any other key cancels"
	d.onKey = func(m *Model, k tea.KeyPressMsg) tea.Cmd {
		m.modal = nil
		if k.String() == "y" {
			m.push(core.Item{Kind: core.RemoveTool, Names: []string{old}}, core.Item{Kind: core.AddTool, Names: []string{new}})
		}
		return nil
	}
	return d
}

// seeds gives the seed files of <t>; with <there>, only those that exist.
func (m *Model) seeds(t string, there bool) []string {
	root := m.state.Home()
	var out []string
	for _, f := range m.manifest(t).Files {
		if f.Mode == "seed" && (!there || core.Exists(root+"/"+f.Target)) {
			out = append(out, f.Target)
		}
	}
	return out
}

// errModal shows a problem of core with a message only; any key closes it
// (D13 b).
func errModal(err error) *modal {
	var body []line
	for _, l := range strings.Split(err.Error(), "\n") {
		for _, w := range wrap(l, 67) {
			body = append(body, line{text(w)})
		}
	}
	body = append(body, line{}, line{styled("Press any key to close.", stDim)})
	return &modal{title: "Not possible", warn: true, width: 72, body: body, keys: "any key closes"}
}

// afterKey follows a key of the main screen on the Tools tab: a move of the
// cursor in the left pane shows the tools of the new category from the top,
// and the bottom pane shows the item under the cursor.
func (m *Model) afterKey(msg tea.KeyPressMsg) {
	switch {
	case key.Matches(msg, m.keys.Move):
		m.afterMove(!m.tabs[toolsTab].onRight)
	case key.Matches(msg, m.keys.Pane):
		if m.search {
			m.tabs[toolsTab].onRight = true
		}
		m.afterMove(false)
	}
}
