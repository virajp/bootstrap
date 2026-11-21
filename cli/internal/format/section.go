package format

import "strings"

// This file holds the section of a text file: the one part of the file that
// bootstrap owns, between two marker lines at the top. Each line outside the
// markers belongs to the user. Contract: docs/design/cli.md ("Shared
// files").

const (
	sectionStart = "bootstrap:start"
	sectionEnd   = "bootstrap:end"
	startText    = sectionStart + " - the next sync writes this part again, and a change in it is lost"
	endText      = sectionEnd + " - add own text below this line"
	htmlComment  = "<!--"
)

// asComment gives <text> as a comment line with <marker>. The marker <!--
// closes the line with -->.
func asComment(marker, text string) string {
	if marker == htmlComment {
		return htmlComment + " " + text + " -->"
	}
	return marker + " " + text
}

// uncomment gives the text of the comment line <line> with <marker>, or false
// when it is not such a line.
func uncomment(marker, line string) (string, bool) {
	s, ok := strings.CutPrefix(strings.TrimSpace(line), marker+" ")
	if ok && marker == htmlComment {
		s, ok = strings.CutSuffix(s, " -->")
	}
	return s, ok
}

// isMarker tells if the comment text <t> starts with the marker word <word>.
func isMarker(t, word string) bool {
	return t == word || strings.HasPrefix(t, word+" ")
}

// findSection gives the indexes of the start line and of the end line of the
// section in <lines>, or -1 for each line that is not there.
func findSection(lines []Value, marker string) (int, int) {
	start := -1
	for i, l := range lines {
		t, ok := uncomment(marker, l.s)
		switch {
		case !ok:
		case start < 0 && isMarker(t, sectionStart):
			start = i
		case start >= 0 && isMarker(t, sectionEnd):
			return start, i
		}
	}
	return start, -1
}

// sectionLines gives the lines of the section of <value>, with its markers.
func sectionLines(value, marker string) []Value {
	out := []Value{str(asComment(marker, startText))}
	for _, l := range strings.Split(strings.TrimSuffix(value, "\n"), "\n") {
		out = append(out, str(l))
	}
	return append(out, str(asComment(marker, endText)))
}

// setLines replaces the lines of the text document. The writer of a text
// file reads only the value, so the op only marks the change.
func (d *doc) setLines(lines []Value) {
	d.v = arr(lines)
	d.ops = append(d.ops, op{})
}

// cutSection removes the lines <start> to <end> and one empty line after
// them, the line that mergeSection adds before the text of the user.
func (d *doc) cutSection(start, end int) {
	lines := d.v.a
	if end+1 < len(lines) && lines[end+1].s == "" {
		end++
	}
	d.setLines(append(append([]Value(nil), lines[:start]...), lines[end+1:]...))
}

// mergeSection writes the section <it> into the text document. A found
// section is written again from the spec. A new section goes at the top,
// with an empty line before the text of the user. It tells if the file has
// the section after the merge.
func (d *doc) mergeSection(it *specItem, warn func(string)) bool {
	lines := d.v.a
	start, end := findSection(lines, d.comment)
	if start >= 0 && end < 0 {
		warn(sectionStart + " has no " + sectionEnd + " - wrote the section again at that line")
		end = start
	}
	if it.remove {
		if start >= 0 {
			d.cutSection(start, end)
		}
		return false
	}
	want := sectionLines(it.value.s, d.comment)
	if start >= 0 {
		if !equal(arr(lines[start:end+1]), arr(want)) {
			d.setLines(append(append(append([]Value(nil), lines[:start]...), want...), lines[end+1:]...))
		}
		return true
	}
	if it.skip {
		return false
	}
	if len(lines) > 0 {
		want = append(want, str(""))
	}
	d.setLines(append(want, lines...))
	return true
}

// cleanupSection removes the section that bootstrap added before and that
// no spec has now.
func (d *doc) cleanupSection() {
	start, end := findSection(d.v.a, d.comment)
	if start >= 0 && end >= 0 {
		d.cutSection(start, end)
	}
}
