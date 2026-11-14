package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// modal is a modal dialog (ruling 22): it takes over the center of the
// screen, over the dimmed panes, up to modalMax lines. A long body scrolls
// in the dialog, with ▲ ▼ at its right edge. A yes/no question is the last
// line of the body and ends with [y/N]. The key line shows the keys of the
// dialog.
type modal struct {
	title   string
	warn    bool // The title has the color of a warning.
	width   int
	body    []line
	top     int  // The first body line that shows.
	buttons line // The buttons of a form, or nil.
	keys    string
	onKey   func(m *Model, k tea.KeyPressMsg) tea.Cmd
}

// key handles a key of the dialog.
func (d *modal) key(m *Model, k tea.KeyPressMsg) tea.Cmd {
	if d.onKey == nil {
		m.modal = nil
		return nil
	}
	return d.onKey(m, k)
}

// rows gives the count of the body lines that show.
func (d *modal) rows() int {
	frame := 4 // The title, a blank line, a blank line and the bottom.
	if d.buttons != nil {
		frame++
	}
	return min(len(d.body), modalMax-frame)
}

// scroll scrolls the body by <n> lines.
func (d *modal) scroll(n int) {
	d.top = min(max(d.top+n, 0), len(d.body)-d.rows())
}

// box gives the lines of the dialog.
func (d *modal) box() []line {
	w := d.width
	edge := func(s string) seg { return styled(s, stHead) }
	titleSt := stHead
	if d.warn {
		titleSt = stLock.Bold(true)
	}
	out := []line{{
		edge("┌─ "), styled(d.title, titleSt),
		edge(" " + strings.Repeat("─", max(0, w-5-len([]rune(d.title)))) + "┐"),
	}}
	blank := line{edge("│"), text(strings.Repeat(" ", w-2)), edge("│")}
	out = append(out, blank)
	n := d.rows()
	for i := range n {
		body := clip(d.body[d.top+i], w-5).pad(w - 5)
		l := append(line{edge("│ ")}, body...)
		l = append(l, scrollMark(i, n, d.top > 0, d.top+n < len(d.body)), edge(" │"))
		out = append(out, l)
	}
	out = append(out, blank)
	if d.buttons != nil {
		l := append(line{edge("│ ")}, clip(d.buttons, w-4).pad(w-4)...)
		out = append(out, append(l, edge(" │")))
	}
	return append(out, line{edge("└" + strings.Repeat("─", w-2) + "┘")})
}

// over gives the lines of the screen <bg> with the dialog over its center:
// the screen is dimmed, and its key line shows the keys of the dialog.
func (d *modal) over(bg []line) []line {
	box := d.box()
	top := modalTop + (modalMax-len(box))/2
	left := (screenW - d.width) / 2
	out := make([]line, len(bg))
	for i, l := range bg {
		plain := []rune(fit(l.plain(), screenW))
		switch {
		case i == len(bg)-1:
			out[i] = line{styled(" "+d.keys, stDim)}
		case i >= top && i < top+len(box):
			o := line{styled(string(plain[:left]), stBehind)}
			o = append(o, box[i-top]...)
			out[i] = append(o, styled(string(plain[left+d.width:]), stBehind))
		default:
			out[i] = line{styled(string(plain), stBehind)}
		}
	}
	return out
}
