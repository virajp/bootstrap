package tui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/selection"
)

// The Packs tab (screen 2 of the design): the packs with their marks in the
// left pane, the members of the pack under the cursor in the right pane,
// read-only, and the details of the item under the cursor in the bottom
// pane. A pack is a shortcut: select is one AddPack item, and deselect is
// one RemoveTool item of its members. Each text of a pack comes from the
// data.

// packsTab is the index of the Packs tab.
const packsTab = 1

// The widths of the columns of the Packs tab (screens.txt, "The fixed
// columns").
const (
	packW      = 16 // The name of a pack.
	memberSumW = 35 // The summary of a member.
	memberStW  = 11 // The status of a member.
)

// The marks of a pack row (ruling 47).
const (
	packOn      = "[x]" // Each member is selected.
	packSome    = "[~]" // Some members are selected.
	packOff     = "[ ]" // No member is selected.
	packAdd     = "[+]" // The change list adds the pack.
	packRemove  = "[-]" // The change list removes the pack.
	packLocked  = "[=]" // The pack base: the user cannot deselect it.
	packLockedS = "required"
)

// packRow is a row of the left pane: a pack, its mark and its counter.
type packRow struct {
	name, mark, count string
}

func (r packRow) render(_ int, cursor bool, tick int) line {
	cur := "  "
	if cursor {
		cur = "▸ "
	}
	return line{text(cur), styled(r.mark, packMarkStyle(r.mark)), text(" " + cell(r.name, packW, tick) + " " + rightAlign(r.count, countW) + " ")}
}

// packMarkStyle gives the style of the pack mark <mk>.
func packMarkStyle(mk string) lipgloss.Style {
	switch mk {
	case packOn, packSome:
		return stSel
	case packAdd:
		return stAdd
	case packRemove:
		return stRemove
	case packLocked:
		return stLock
	}
	return stPlain
}

// memberRow is a row of the right pane: a member of the pack, with no
// selector, its summary and a dim status.
type memberRow struct {
	tool, summary, status string
	nameSt                lipgloss.Style
}

func (r memberRow) render(_ int, _ bool, tick int) line {
	return line{
		text("     "), styled(cell(r.tool, nameW, tick), r.nameSt), text(" "),
		styled(fit(r.summary, memberSumW), stDim), text(" "),
		styled(fit(r.status, memberStW), stDim), text(" "),
	}
}

// The packs of the data.

// packNames gives the packs in the order of the packs file, then each pack
// with no member line.
func (m *Model) packNames() []string {
	p := m.state.Packs
	var out []string
	for _, o := range p.Order {
		if !slices.Contains(out, o.Pack) && p.HasPack(o.Pack) {
			out = append(out, o.Pack)
		}
	}
	for _, n := range p.Names {
		if !slices.Contains(out, n) {
			out = append(out, n)
		}
	}
	return out
}

// packMembers gives the tools of the pack <name> in file order, a nested
// pack "pack:<x>" expanded in turn.
func (m *Model) packMembers(name string) []string {
	var out []string
	var walk func(n string, seen map[string]bool)
	walk = func(n string, seen map[string]bool) {
		if seen[n] {
			return
		}
		seen[n] = true
		p := m.state.Packs
		for _, o := range p.Order {
			if o.Pack != n || !p.HasMember(n, o.Key) {
				continue
			}
			if rest, ok := strings.CutPrefix(o.Key, "pack:"); ok {
				walk(rest, seen)
				continue
			}
			if m.manifest(o.Key) != nil && !slices.Contains(out, o.Key) {
				out = append(out, o.Key)
			}
		}
	}
	walk(name, map[string]bool{})
	return out
}

// usable tells if the member <t> can be selected here: a pack leaves out a
// member with selectable = false, or that is not available here.
func (m *Model) usable(t string) bool {
	return m.manifest(t).Selectable && m.notHere(t) == ""
}

// packItem gives the index of the change of the pack <name> on the list, or
// -1.
func (m *Model) packItem(name string) int {
	for i := len(m.list.Items) - 1; i >= 0; i-- {
		it := m.list.Items[i]
		if (it.Kind == core.AddPack && slices.Contains(it.Names, name)) || (it.Kind == core.RemoveTool && it.Pack == name) {
			return i
		}
	}
	return -1
}

// packMark gives the mark of the pack <name>, and the counter of the
// members that the values file selects (screen 2).
func (m *Model) packMark(name string) (string, string) {
	all, n, now := 0, 0, 0
	for _, t := range m.packMembers(name) {
		if !m.usable(t) {
			continue
		}
		all++
		if m.state.Selected(t) {
			n++
		}
		if m.selected(t) {
			now++
		}
	}
	count := itoa(n) + "/" + itoa(all)
	if name == selection.BasePack {
		return packLocked, count
	}
	if i := m.packItem(name); i >= 0 {
		if m.list.Items[i].Kind == core.AddPack {
			return packAdd, count
		}
		return packRemove, count
	}
	switch {
	case all > 0 && now == all:
		return packOn, count
	case now > 0:
		return packSome, count
	}
	return packOff, count
}

// memberStatus gives the status of the member <t> of a pack.
func (m *Model) memberStatus(t string) string {
	e := m.effect()
	switch {
	case !m.selected(t) && !m.usable(t):
		return "not here"
	case e.Removed[t] && m.state.Selected(t):
		return "will remove"
	case e.Added[t]:
		return "will add"
	case m.selected(t) && m.state.Required(t):
		return packLockedS
	case m.selected(t):
		return "installed"
	}
	return "not added"
}

// refreshPacks builds the panes of the Packs tab again from the state and
// the change list. The cursors stay where they are.
func (m *Model) refreshPacks() {
	if m.state == nil {
		return
	}
	t := m.tabs[packsTab]
	var rows []row
	for _, n := range m.packNames() {
		mk, count := m.packMark(n)
		rows = append(rows, packRow{name: n, mark: mk, count: count})
	}
	t.left.rows = rows
	t.left.move(0)
	t.right.title, t.right.rows = "", nil
	if p, ok := m.curPack(); ok {
		t.right.title = p.name
		e := m.effect()
		for _, tool := range m.packMembers(p.name) {
			st := stPlain
			switch {
			case e.Removed[tool] && m.state.Selected(tool):
				st = stRemove
			case e.Added[tool]:
				st = stAdd
			case !m.selected(tool) && !m.usable(tool):
				st = stDim
			}
			t.right.rows = append(t.right.rows, memberRow{tool: tool, summary: m.manifest(tool).Summary, status: m.memberStatus(tool), nameSt: st})
		}
	}
	t.right.move(0)
	t.bottom = m.packsBottom()
}

// curPack gives the pack under the cursor of the left pane.
func (m *Model) curPack() (packRow, bool) {
	l := m.tabs[packsTab].left
	if l.cursor < len(l.rows) {
		p, ok := l.rows[l.cursor].(packRow)
		return p, ok
	}
	return packRow{}, false
}

// curMember gives the member under the cursor of the right pane.
func (m *Model) curMember() (memberRow, bool) {
	r := m.tabs[packsTab].right
	if r.cursor < len(r.rows) {
		t, ok := r.rows[r.cursor].(memberRow)
		return t, ok
	}
	return memberRow{}, false
}

// The bottom pane.

// packsBottom gives the lines of the bottom pane: on a pack, its name, its
// description, its status and its members (D7 a); on a member, the details
// of the tool.
func (m *Model) packsBottom() []line {
	if m.tabs[packsTab].onRight {
		r, ok := m.curMember()
		if !ok {
			return nil
		}
		out := m.toolDetail(r.tool)
		return append(out, line{styled(" The right pane is read-only. To remove "+r.tool+", use the Tools tab.", stDim)})
	}
	p, ok := m.curPack()
	if !ok {
		return nil
	}
	out := []line{{styled(" "+p.name, stHead)}}
	for _, l := range wrap(m.state.Packs.Descriptions[p.name], screenW-3) {
		out = append(out, line{text(" " + l)})
	}
	out = append(out, line{})
	members := m.packMembers(p.name)
	out = append(out, line{
		styled(" "+fit("Status", kvKeyW), stDim), text(fit(m.packStatus(p.name, p.mark), kvValW)),
		styled(" "+fit("Members", kvKeyW), stDim), text(itoa(len(members))),
	})
	switch p.mark {
	case packAdd:
		out = append(out, line{styled(" Same as: "+core.CmdName(m.opt.Side.Mac, "add pack")+" "+p.name, stDim)})
	case packRemove:
		out = append(out, line{styled(" Same as: one "+core.CmdName(m.opt.Side.Mac, "remove tool")+" for each member (no remove pack).", stDim)})
	}
	return out
}

// packStatus gives the status of the pack <name> with the mark <mk>.
func (m *Model) packStatus(name, mk string) string {
	e := m.effect()
	n := 0
	for _, t := range m.packMembers(name) {
		if (mk == packAdd && e.Added[t]) || (mk == packRemove && e.Removed[t] && m.state.Selected(t)) {
			n++
		}
	}
	switch mk {
	case packAdd:
		return "will add " + plural(n, "tool")
	case packRemove:
		return "will remove " + plural(n, "tool")
	case packLocked:
		return packLockedS
	case packOn:
		return "installed"
	case packSome:
		return "some members are selected"
	}
	return "not selected"
}

// The keys of the tab.

// packsKey handles a key of the Packs tab that the main screen does not
// handle. It tells if it handled the key. Space in the right pane does
// nothing (ruling 59).
func (m *Model) packsKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if m.state == nil || msg.String() != "space" {
		return nil, false
	}
	if m.tabs[packsTab].onRight {
		return nil, true
	}
	return m.packSpace(), true
}

// afterPacksKey follows a key of the main screen on the Packs tab: a move
// in the left pane shows the members of the new pack from the top, and the
// bottom pane shows the item under the cursor.
func (m *Model) afterPacksKey(msg tea.KeyPressMsg) {
	t := m.tabs[packsTab]
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
	m.refreshPacks()
}

// packSpace selects or deselects the pack under the cursor. [=] does
// nothing; [+] and [-] take the change of the pack off the list.
func (m *Model) packSpace() tea.Cmd {
	p, ok := m.curPack()
	if !ok {
		return nil
	}
	switch p.mark {
	case packLocked:
		return nil
	case packAdd, packRemove:
		m.setItems(slices.Delete(slices.Clone(m.list.Items), m.packItem(p.name), m.packItem(p.name)+1))
		return nil
	case packOn:
		return m.packDeselect(p.name)
	}
	return m.selectHook(core.Item{Kind: core.AddPack, Names: []string{p.name}})
}

// packDeselect removes the selected members of the pack <name>, except a
// required member and a member that a tool that stays needs. A selected
// tool outside the pack that needs a member leaves too, after the cascade
// dialog (ruling 48).
func (m *Model) packDeselect(name string) tea.Cmd {
	members := m.packMembers(name)
	// The tools that stay: each required tool and what it needs.
	var stay []string
	for _, n := range m.state.Sel.Idx.Names {
		if m.selected(n) && m.state.Required(n) {
			stay = append(stay, n)
		}
	}
	kept := m.state.Closure(stay)
	var gone []string
	for _, t := range members {
		if m.selected(t) && !kept[t] {
			gone = append(gone, t)
		}
	}
	// Each selected member stays: the deselect changes nothing, as space
	// on [=] does.
	if len(gone) == 0 {
		return nil
	}
	// The selected tools outside the pack that need a member that leaves.
	var outside []string
	for _, n := range m.state.Sel.Idx.Names {
		if slices.Contains(gone, n) || !m.selected(n) || kept[n] {
			continue
		}
		for c := range m.state.Closure([]string{n}) {
			if c != n && slices.Contains(gone, c) {
				outside = append(outside, n)
				break
			}
		}
	}
	it := core.Item{Kind: core.RemoveTool, Names: append(slices.Clone(gone), outside...), Pack: name}
	items := append(slices.Clone(m.list.Items), it)
	if len(outside) == 0 {
		m.setItems(items)
		return nil
	}
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
	m.modal = m.cascadeModal(name, gone, outside, extra, items)
	return nil
}
