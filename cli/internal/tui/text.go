package tui

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// The text of the screen: a line is a list of segments, each with its style.
// The plain text of a line is the text of its segments, so a modal dialog can
// dim the screen behind it.

// seg is a text with its style. A raw segment holds text that is styled
// already, for example the view of the help component.
type seg struct {
	text  string
	style lipgloss.Style
	raw   bool
}

// line is one line of the screen.
type line []seg

// text gives a segment with no style.
func text(s string) seg { return seg{text: s, style: lipgloss.NewStyle()} }

// styled gives a segment with the style <st>.
func styled(s string, st lipgloss.Style) seg { return seg{text: s, style: st} }

// width gives the number of columns of <l>.
func (l line) width() int {
	n := 0
	for _, s := range l {
		n += lipgloss.Width(s.text)
	}
	return n
}

// pad gives <l> with spaces at its end up to <w> columns.
func (l line) pad(w int) line {
	if n := l.width(); n < w {
		return append(l, text(strings.Repeat(" ", w-n)))
	}
	return l
}

// plain gives the text of <l> with no style.
func (l line) plain() string {
	var b strings.Builder
	for _, s := range l {
		if s.raw {
			b.WriteString(stripStyle(s.text))
			continue
		}
		b.WriteString(s.text)
	}
	return b.String()
}

// render gives <l> with its styles.
func (l line) render() string {
	var b strings.Builder
	for _, s := range l {
		if s.raw {
			b.WriteString(s.text)
			continue
		}
		b.WriteString(s.style.Render(s.text))
	}
	return b.String()
}

// cursorRow gives <l> with the background of the cursor row on each
// segment.
func (l line) cursorRow() line {
	out := make(line, len(l))
	for i, s := range l {
		s.style = s.style.Background(colCursor)
		if _, none := s.style.GetForeground().(lipgloss.NoColor); none {
			s.style = s.style.Foreground(colText)
		}
		out[i] = s
	}
	return out
}

// stripStyle gives <s> without its ANSI escape sequences.
func stripStyle(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			in = true
		case in:
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				in = false
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// fit gives <s> at exactly <w> columns: cut with "…" when it is longer, else
// with spaces at its end.
func fit(s string, w int) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) > w {
		return string(r[:w-1]) + "…"
	}
	return s + strings.Repeat(" ", w-len(r))
}

// Scroll of the long text of the cursor row: the text holds for holdTicks
// ticks, then moves one character for each tick, with loopGap spaces between
// its end and its start.
const (
	holdTicks = 10
	loopGap   = 3
)

// cell gives the text <s> in a column of <w> columns. <tick> is the count of
// ticks since the cursor came to the row; a tick below 0 is a row without
// the cursor. On the cursor row, a text longer than the column scrolls.
func cell(s string, w, tick int) string {
	r := []rune(s)
	if tick < 0 || len(r) <= w {
		return fit(s, w)
	}
	loop := append(r, []rune(strings.Repeat(" ", loopGap))...)
	p := tick % (holdTicks + len(loop))
	if p < holdTicks {
		return fit(s, w)
	}
	start := p - holdTicks + 1
	out := make([]rune, w)
	for k := range out {
		out[k] = loop[(start+k)%len(loop)]
	}
	return string(out)
}

// rightAlign gives <s> with spaces before it up to <w> columns.
func rightAlign(s string, w int) string {
	if n := len([]rune(s)); n < w {
		return strings.Repeat(" ", w-n) + s
	}
	return s
}

// plural gives "<n> <noun>" with the plural of <noun> when <n> is not 1.
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}

// wrap gives the words of <s> in lines of <w> columns or fewer. A word
// longer than <w> is a line of its own.
func wrap(s string, w int) []string {
	var out []string
	cur := ""
	for _, word := range strings.Fields(s) {
		switch {
		case cur == "":
			cur = word
		case len([]rune(cur))+1+len([]rune(word)) > w:
			out = append(out, cur)
			cur = word
		default:
			cur += " " + word
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// itoa gives <n> as a decimal text.
func itoa(n int) string { return strconv.Itoa(n) }

// andList gives the names <l> joined with ", " and a last " and ".
func andList(l []string) string {
	if len(l) < 2 {
		return strings.Join(l, "")
	}
	return strings.Join(l[:len(l)-1], ", ") + " and " + l[len(l)-1]
}
