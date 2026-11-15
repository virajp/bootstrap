package tui

import "strings"

// The smallest layout of the main screen, in columns and lines (screens.txt,
// "The fixed columns"): the header (2 lines), a rule, the titles of the two
// top panes, 11 rows, a rule, the bottom pane (7 lines), a rule and the key
// line. A larger terminal gives a larger screen with the same proportions
// (layoutFor).
const (
	screenW  = 100
	screenH  = 25
	leftW    = 30           // The left pane, with its scroll mark.
	paneH    = 11           // The rows of the two top panes.
	bottomH  = 7            // The lines of the bottom pane.
	modalMax = 20           // The most lines of a modal dialog.
	modalTop = 3            // The first line of the area of a dialog.
	bodyMax  = modalMax - 6 // The body lines of a dialog with buttons.
	upMark   = "▲"          // More lines above.
	downMark = "▼"          // More lines below.
)

// layout is the size of the screen and of its panes, in columns and lines.
type layout struct {
	w, h          int // The screen.
	left, right   int // The two top panes, each with its scroll mark.
	paneH, bottom int // The rows of the two top panes, the lines of the bottom pane.
}

// layoutFor gives the layout of a terminal of <w> × <h>: the full terminal,
// and never less than screenW × screenH. The left pane keeps its part of
// the width, and the top panes and the bottom pane share the extra lines
// in the proportion paneH to bottomH.
func layoutFor(w, h int) layout {
	w, h = max(w, screenW), max(h, screenH)
	left := w * leftW / screenW
	extra := h - screenH
	panes := paneH + extra*paneH/(paneH+bottomH)
	return layout{
		w: w, h: h,
		left: left, right: w - left - 1,
		paneH: panes, bottom: bottomH + extra - (panes - paneH),
	}
}

// row is one row of a top pane. render gives the row at <width> columns.
// <cursor> tells if the cursor is on the row; <tick> is the step of the
// scroll of its long text, or -1 when the row does not scroll.
type row interface {
	render(width int, cursor bool, tick int) line
}

// pane is a top pane: its title, its rows, the cursor, the first row that
// shows and the count of its rows that show.
type pane struct {
	title  string
	rows   []row
	cursor int
	top    int
	height int
}

// move moves the cursor by <d> rows, and keeps it in view.
func (p *pane) move(d int) {
	if len(p.rows) == 0 {
		p.cursor, p.top = 0, 0
		return
	}
	p.cursor = min(max(p.cursor+d, 0), len(p.rows)-1)
	if p.cursor < p.top {
		p.top = p.cursor
	}
	if p.cursor >= p.top+p.height {
		p.top = p.cursor - p.height + 1
	}
}

// lines gives the height lines of the pane at <width> columns, the scroll
// mark in the last column. The cursor row of the active pane has the
// background of the cursor, and its long text scrolls with <tick>.
func (p pane) lines(width int, active bool, tick int) []line {
	out := make([]line, p.height)
	for i := range out {
		var l line
		n := p.top + i
		if n < len(p.rows) {
			cur := n == p.cursor
			t := -1
			if cur && active {
				t = tick
			}
			l = p.rows[n].render(width-1, cur, t).pad(width - 1)
			if cur && active {
				l = l.cursorRow()
			}
		} else {
			l = line{text(strings.Repeat(" ", width-1))}
		}
		out[i] = append(l, scrollMark(i, p.height, p.top > 0, p.top+p.height < len(p.rows)))
	}
	return out
}

// scrollMark gives the mark of line <i> of an area of <n> lines: ▲ on the
// first line with more lines above, ▼ on the last line with more lines
// below, else a space.
func scrollMark(i, n int, up, down bool) seg {
	switch {
	case up && i == 0:
		return styled(upMark, stMark)
	case down && i == n-1:
		return styled(downMark, stMark)
	}
	return text(" ")
}

// tab is one tab of the main screen: its two top panes, the pane with the
// cursor, and the lines of the bottom pane with their first line that
// shows. The tabs Tools, Packs and Detected fill them.
type tab struct {
	name        string
	left, right pane
	onRight     bool
	bottom      []line
	bottomTop   int
}

// active gives the pane with the cursor.
func (t *tab) active() *pane {
	if t.onRight {
		return &t.right
	}
	return &t.left
}

// scrollBottom scrolls the bottom pane of <h> lines by <d> lines.
func (t *tab) scrollBottom(d, h int) {
	t.bottomTop = min(max(t.bottomTop+d, 0), max(len(t.bottom)-h, 0))
}

// bottomLines gives the <h> lines of the bottom pane at <w> columns, the
// scroll mark in the last column.
func (t *tab) bottomLines(w, h int) []line {
	t.scrollBottom(0, h)
	out := make([]line, h)
	for i := range out {
		var l line
		if n := t.bottomTop + i; n < len(t.bottom) {
			l = t.bottom[n]
		}
		l = clip(l, w-1).pad(w - 1)
		out[i] = append(l, scrollMark(i, h, t.bottomTop > 0, t.bottomTop+h < len(t.bottom)))
	}
	return out
}

// clip cuts the line <l> at <w> columns, with "…" at its end. A raw segment
// stays whole.
func clip(l line, w int) line {
	if l.width() <= w {
		return l
	}
	var out line
	left := w
	for _, s := range l {
		n := len([]rune(s.text))
		if s.raw || n < left {
			out = append(out, s)
			left -= n
			continue
		}
		s.text = fit(s.text, left)
		return append(out, s)
	}
	return out
}
