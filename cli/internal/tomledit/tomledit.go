// Package tomledit changes keys and tables of a TOML file and keeps each
// other line as it is: the comments, the blank lines, the order and the
// layout of the values. A new value has the layout of dprint, for example
// `[ "a", "b" ]`, and a new key goes to its sorted place in its group, as the
// sortKeys of dprint puts it.
//
// The editor knows the forms that bootstrap writes: `[table]` headers and
// `key = value` lines, also a value on more than one line. For each other
// form, an array of tables (`[[x]]`), a dotted key (`a.b = 1`) or an inline
// table on the path of the change, it gives ok false, and the caller uses
// another method.
package tomledit

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type kind int

const (
	blank kind = iota
	comment
	header
	keyval
)

// stmt is one statement of the file: lines first to last.
type stmt struct {
	kind        kind
	first, last int
	path        []string // the table of a header, or the table of a key
	key         string   // the key of a keyval, unquoted
}

type doc struct {
	lines []string // without the newline at the end of the file
	stmts []stmt
}

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// parse splits <src> into statements. ok is false for a form that the
// editor does not edit.
func parse(src []byte) (*doc, bool) {
	text := strings.TrimSuffix(string(src), "\n")
	d := &doc{}
	if text != "" {
		d.lines = strings.Split(text, "\n")
	}
	var table []string
	for i := 0; i < len(d.lines); i++ {
		t := strings.TrimSpace(d.lines[i])
		switch {
		case t == "":
			d.stmts = append(d.stmts, stmt{kind: blank, first: i, last: i})
		case strings.HasPrefix(t, "#"):
			d.stmts = append(d.stmts, stmt{kind: comment, first: i, last: i})
		case strings.HasPrefix(t, "[["):
			return nil, false
		case strings.HasPrefix(t, "["):
			p, ok := parseHeader(t)
			if !ok {
				return nil, false
			}
			table = p
			d.stmts = append(d.stmts, stmt{kind: header, first: i, last: i, path: p})
		default:
			key, rest, ok := parseKey(t)
			if !ok {
				return nil, false
			}
			last, ok := valueEnd(d.lines, i, rest)
			if !ok {
				return nil, false
			}
			d.stmts = append(d.stmts, stmt{kind: keyval, first: i, last: last, path: table, key: key})
			i = last
		}
	}
	return d, true
}

// parseKey reads `key =` at the start of <t>. It gives the key and the text
// after `=`. A dotted key is not ok.
func parseKey(t string) (string, string, bool) {
	key, rest, ok := readKey(t)
	if !ok {
		return "", "", false
	}
	rest = strings.TrimLeft(rest, " \t")
	if !strings.HasPrefix(rest, "=") {
		return "", "", false
	}
	return key, rest[1:], true
}

// readKey reads one bare or quoted key at the start of <t>.
func readKey(t string) (string, string, bool) {
	switch {
	case strings.HasPrefix(t, `"`):
		end := 1
		for end < len(t) && t[end] != '"' {
			if t[end] == '\\' {
				end++
			}
			end++
		}
		if end >= len(t) {
			return "", "", false
		}
		k, err := strconv.Unquote(t[:end+1])
		if err != nil {
			return "", "", false
		}
		return k, t[end+1:], true
	case strings.HasPrefix(t, "'"):
		end := strings.Index(t[1:], "'")
		if end < 0 {
			return "", "", false
		}
		return t[1 : end+1], t[end+2:], true
	}
	n := 0
	for n < len(t) && (t[n] == '_' || t[n] == '-' || t[n] >= '0' && t[n] <= '9' || t[n] >= 'a' && t[n] <= 'z' || t[n] >= 'A' && t[n] <= 'Z') {
		n++
	}
	if n == 0 {
		return "", "", false
	}
	return t[:n], t[n:], true
}

// parseHeader reads `[a.b."c"]`, with an optional comment after it.
func parseHeader(t string) ([]string, bool) {
	rest := strings.TrimLeft(t[1:], " \t")
	var path []string
	for {
		k, r, ok := readKey(rest)
		if !ok {
			return nil, false
		}
		path = append(path, k)
		r = strings.TrimLeft(r, " \t")
		switch {
		case strings.HasPrefix(r, "."):
			rest = strings.TrimLeft(r[1:], " \t")
		case strings.HasPrefix(r, "]"):
			tail := strings.TrimSpace(r[1:])
			return path, tail == "" || strings.HasPrefix(tail, "#")
		default:
			return nil, false
		}
	}
}

// valueEnd finds the last line of a value that starts with <rest> on line
// <i>: an array or a string can go on over more lines.
func valueEnd(lines []string, i int, rest string) (int, bool) {
	depth := 0
	inStr := ""
	text := rest
	for {
		for j := 0; j < len(text); j++ {
			c := text[j]
			if inStr != "" {
				switch {
				case inStr == `"` && c == '\\', inStr == `"""` && c == '\\':
					j++
				case strings.HasPrefix(text[j:], inStr):
					j += len(inStr) - 1
					inStr = ""
				}
				continue
			}
			switch {
			case c == '#':
				j = len(text)
			case strings.HasPrefix(text[j:], `"""`), strings.HasPrefix(text[j:], `'''`):
				inStr = text[j : j+3]
				j += 2
			case c == '"' || c == '\'':
				inStr = string(c)
			case c == '[' || c == '{':
				depth++
			case c == ']' || c == '}':
				depth--
			}
		}
		if inStr == `"` || inStr == "'" {
			return 0, false
		}
		if depth <= 0 && inStr == "" {
			return i, true
		}
		i++
		if i >= len(lines) {
			return 0, false
		}
		text = lines[i]
	}
}

func hasPrefix(p, pre []string) bool {
	if len(p) < len(pre) {
		return false
	}
	for i := range pre {
		if p[i] != pre[i] {
			return false
		}
	}
	return true
}

func equal(a, b []string) bool { return len(a) == len(b) && hasPrefix(a, b) }

// inlineOnPath tells if a key on <path>, or a key above it, has its value
// in the file: an inline table or a value that the editor cannot go into.
func (d *doc) inlineOnPath(path []string) bool {
	for _, s := range d.stmts {
		if s.kind != keyval {
			continue
		}
		full := append(append([]string{}, s.path...), s.key)
		if len(full) < len(path) && hasPrefix(path, full) {
			return true
		}
	}
	return false
}

// section gives the index of the header of <table>, or -1. The top level,
// an empty <table>, has no header and gives -1 too.
func (d *doc) section(table []string) int {
	for i, s := range d.stmts {
		if s.kind == header && equal(s.path, table) {
			return i
		}
	}
	return -1
}

// regionEnd gives the index of the first statement after the region of the
// header at <h>: the next header.
func (d *doc) regionEnd(h int) int {
	for i := h + 1; i < len(d.stmts); i++ {
		if d.stmts[i].kind == header {
			return i
		}
	}
	return len(d.stmts)
}

func (d *doc) out(lines []string) []byte {
	if len(lines) == 0 {
		return nil
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

// splice gives the lines with lines[from:to] replaced by <with>.
func splice(lines []string, from, to int, with []string) []string {
	out := append([]string{}, lines[:from]...)
	out = append(out, with...)
	return append(out, lines[to:]...)
}

// Set sets the key at <path> to <value>, a value of encoding/json (with
// json.Number for a number). A map value becomes a table. ok is false for a
// form that the editor does not edit; out is then nil.
func Set(src []byte, path []string, value any) ([]byte, bool) {
	d, ok := parse(src)
	if !ok || len(path) == 0 || d.inlineOnPath(path) {
		return nil, false
	}
	if m, isMap := value.(map[string]any); isMap {
		for _, s := range d.stmts {
			if s.kind == keyval && equal(append(append([]string{}, s.path...), s.key), path) {
				return nil, false // an inline table at the path: not this editor
			}
		}
		block, ok := renderTable(path, m)
		if !ok {
			return nil, false
		}
		return d.setTable(path, block), true
	}
	text, ok := render(value)
	if !ok {
		return nil, false
	}
	table, key := path[:len(path)-1], path[len(path)-1]
	line := quoteKey(key) + " = " + text
	for _, s := range d.stmts {
		if s.kind == header && equal(s.path, path) {
			return nil, false // a table becomes a value: not this editor
		}
		if s.kind == keyval && equal(s.path, table) && s.key == key {
			indent := d.lines[s.first][:len(d.lines[s.first])-len(strings.TrimLeft(d.lines[s.first], " \t"))]
			return d.out(splice(d.lines, s.first, s.last+1, []string{indent + line})), true
		}
	}
	return d.out(d.insertKey(table, key, line)), true
}

// insertKey puts the new line of <key> into <table>: at its sorted place in
// the last group of keys of the table, or after the header when the table
// has no key. A table that is not there goes at the end of the file.
func (d *doc) insertKey(table []string, key, line string) []string {
	start, end := 0, d.regionEnd(-1)
	if len(table) > 0 {
		h := d.section(table)
		if h < 0 {
			return d.appendBlock([]string{"[" + headerPath(table) + "]", line})
		}
		start, end = h+1, d.regionEnd(h)
	}
	last := -1
	for i := start; i < end; i++ {
		if d.stmts[i].kind == keyval {
			last = i
		}
	}
	if last < 0 {
		if len(table) > 0 {
			return splice(d.lines, d.stmts[start-1].last+1, d.stmts[start-1].last+1, []string{line})
		}
		// A top level with no key: the key goes before the first table, after
		// the comments at the top.
		at := 0
		for i := 0; i < end && d.stmts[i].kind == comment; i++ {
			at = d.stmts[i].last + 1
		}
		with := []string{line}
		if end < len(d.stmts) {
			with = append(with, "")
		}
		if at > 0 {
			with = append([]string{""}, with...)
		}
		return splice(d.lines, at, at, with)
	}
	first := last
	for first > start && d.stmts[first-1].kind == keyval {
		first--
	}
	for i := first; i <= last; i++ {
		if d.stmts[i].key > key {
			return splice(d.lines, d.stmts[i].first, d.stmts[i].first, []string{line})
		}
	}
	at := d.stmts[last].last + 1
	return splice(d.lines, at, at, []string{line})
}

// appendBlock puts <block> at the end of the file, after one blank line.
func (d *doc) appendBlock(block []string) []string {
	lines := append([]string{}, d.lines...)
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	return append(lines, block...)
}

// setTable puts <block>, the lines of the table <path> and its sub-tables,
// at the place of the old table, or at the end of the file.
func (d *doc) setTable(path []string, block []string) []byte {
	var keep []string
	at := -1
	for i := 0; i < len(d.stmts); i++ {
		s := d.stmts[i]
		if s.kind == header && hasPrefix(s.path, path) {
			end := d.regionEnd(i)
			if at < 0 {
				at = len(keep)
				// The blank lines at the end of the old table stay after the new.
				trail := 0
				for j := end - 1; j > i && d.stmts[j].kind == blank; j-- {
					trail++
				}
				keep = append(keep, block...)
				for ; trail > 0; trail-- {
					keep = append(keep, "")
				}
			}
			i = end - 1
			continue
		}
		keep = append(keep, d.lines[s.first:s.last+1]...)
	}
	if at < 0 {
		return d.out(d.appendBlock(block))
	}
	return d.out(trimEnd(keep))
}

func trimEnd(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// Delete removes the key or the table at <path>. With <prune>, each table
// above it that is then empty goes too, down to the second level. ok is false
// for a form that the editor does not edit; out is then nil. A path that is
// not there gives <src>.
func Delete(src []byte, path []string, prune bool) ([]byte, bool) {
	d, ok := parse(src)
	if !ok || len(path) == 0 || d.inlineOnPath(path) {
		return nil, false
	}
	lines, found := d.remove(path)
	if !found {
		return src, true
	}
	for n := len(path) - 1; prune && n > 0; n-- {
		next, _ := parse(d.out(lines))
		h := next.section(path[:n])
		if h < 0 || !next.empty(h) {
			break
		}
		lines, _ = next.remove(path[:n])
	}
	return d.out(lines), true
}

// empty tells if the table of the header at <h> has no key and no sub-table.
func (d *doc) empty(h int) bool {
	for i := h + 1; i < d.regionEnd(h); i++ {
		if d.stmts[i].kind == keyval {
			return false
		}
	}
	for _, s := range d.stmts {
		if s.kind == header && len(s.path) > len(d.stmts[h].path) && hasPrefix(s.path, d.stmts[h].path) {
			return false
		}
	}
	return true
}

// remove gives the lines without the table <path> and its sub-tables, or
// without the key <path>.
func (d *doc) remove(path []string) ([]string, bool) {
	var keep []string
	found := false
	table, key := path[:len(path)-1], path[len(path)-1]
	for i := 0; i < len(d.stmts); i++ {
		s := d.stmts[i]
		switch {
		case s.kind == header && hasPrefix(s.path, path):
			found = true
			i = d.regionEnd(i) - 1
			continue
		case s.kind == keyval && equal(s.path, table) && s.key == key:
			found = true
			// A key alone between two blank lines takes one of them with it.
			if len(keep) > 0 && strings.TrimSpace(keep[len(keep)-1]) == "" && i+1 < len(d.stmts) && d.stmts[i+1].kind == blank {
				i++
			}
			continue
		}
		keep = append(keep, d.lines[s.first:s.last+1]...)
	}
	return trimEnd(keep), found
}

// EnsureTable adds an empty table <path> when the file has no table on it:
// after the last table that starts with the first segment of <path>, or at
// the end of the file. ok is false for a form that the editor does not edit.
func EnsureTable(src []byte, path []string) ([]byte, bool) {
	d, ok := parse(src)
	if !ok || len(path) == 0 || d.inlineOnPath(path) {
		return nil, false
	}
	after := -1
	for i, s := range d.stmts {
		if s.kind == keyval && equal(append(append([]string{}, s.path...), s.key), path) {
			return src, true // a value at the path: not a table, no change
		}
		if s.kind != header {
			continue
		}
		if hasPrefix(s.path, path) {
			return src, true
		}
		if s.path[0] == path[0] {
			after = i
		}
	}
	line := "[" + headerPath(path) + "]"
	if after < 0 {
		return d.out(d.appendBlock([]string{line})), true
	}
	end := d.regionEnd(after)
	if end == len(d.stmts) {
		return d.out(d.appendBlock([]string{line})), true
	}
	at := d.stmts[end].first
	return d.out(splice(d.lines, at, at, []string{line, ""})), true
}

// renderTable gives the lines of the table <path> with the values of <m>:
// the keys in sorted order, then each sub-table after a blank line.
func renderTable(path []string, m map[string]any) ([]string, bool) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := []string{"[" + headerPath(path) + "]"}
	var subs []string
	for _, k := range keys {
		if _, ok := m[k].(map[string]any); ok {
			subs = append(subs, k)
			continue
		}
		text, ok := render(m[k])
		if !ok {
			return nil, false
		}
		out = append(out, quoteKey(k)+" = "+text)
	}
	for _, k := range subs {
		block, ok := renderTable(append(append([]string{}, path...), k), m[k].(map[string]any))
		if !ok {
			return nil, false
		}
		out = append(out, "")
		out = append(out, block...)
	}
	return out, true
}

// render gives the TOML text of a value, in the layout of dprint.
func render(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return quote(x), true
	case bool:
		return strconv.FormatBool(x), true
	case json.Number:
		return x.String(), true
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10), true
		}
		return strconv.FormatFloat(x, 'g', -1, 64), true
	case int:
		return strconv.Itoa(x), true
	case int64:
		return strconv.FormatInt(x, 10), true
	case []any:
		if len(x) == 0 {
			return "[]", true
		}
		parts := make([]string, len(x))
		for i, e := range x {
			t, ok := render(e)
			if !ok {
				return "", false
			}
			parts[i] = t
		}
		return "[ " + strings.Join(parts, ", ") + " ]", true
	case map[string]any:
		if len(x) == 0 {
			return "{}", true
		}
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			t, ok := render(x[k])
			if !ok {
				return "", false
			}
			parts[i] = quoteKey(k) + " = " + t
		}
		return "{ " + strings.Join(parts, ", ") + " }", true
	}
	return "", false
}

// quote gives a TOML basic string.
func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&b, `\u%04X`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func quoteKey(k string) string {
	if bareKey.MatchString(k) {
		return k
	}
	return quote(k)
}

func headerPath(path []string) string {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = quoteKey(p)
	}
	return strings.Join(parts, ".")
}

// Decode reads the JSON text <text> as a value for Set, with json.Number
// for a number.
func Decode(text string) (any, bool) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, false
	}
	return v, true
}
