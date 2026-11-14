package report

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Printer prints the status lines of a command: one glyph for each meaning.
// With the style off, the glyph becomes a word prefix. Contract:
// docs/design/cli.md ("The output of a run").
type Printer struct {
	W     io.Writer
	Style bool
}

// NewPrinter gives a Printer that writes to <w>.
func NewPrinter(w io.Writer, style bool) Printer { return Printer{W: w, Style: style} }

// status prints one line. <color> paints the glyph; with <whole>, it paints
// the whole line.
func (p Printer) status(glyph, color, plain, text string, whole bool) {
	if !p.Style {
		_, _ = fmt.Fprintf(p.W, "%s %s\n", plain, text)
		return
	}
	switch {
	case color == "":
		_, _ = fmt.Fprintf(p.W, "%s %s\n", glyph, text)
	case whole:
		_, _ = fmt.Fprintf(p.W, "%s%s %s%s\n", color, glyph, text, reset)
	default:
		_, _ = fmt.Fprintf(p.W, "%s%s%s %s\n", color, glyph, reset, text)
	}
}

// Success prints `✔ <text>`.
func (p Printer) Success(text string) { p.status("✔", green, "[OK]", text, false) }

// Warn prints `! <text>`.
func (p Printer) Warn(text string) { p.status("!", yellow, "[WARN]", text, false) }

// Info prints `· <text>`, dim.
func (p Printer) Info(text string) { p.status("·", dim, "[INFO]", text, true) }

// Clone prints `↻ <text>`, dim.
func (p Printer) Clone(text string) { p.status("↻", dim, "[INFO]", text, true) }

// Step prints `▸ <text>`, bold.
func (p Printer) Step(text string) { p.status("▸", bold, "[INFO]", text, true) }

// ErrorFix prints `✗ <problem>` and, when <fix> is not empty, the fix line.
func (p Printer) ErrorFix(problem, fix string) {
	p.status("✗", red, "[ERROR]", problem, true)
	p.Fix(fix)
}

// Fix prints the fix line `  → <fix>`, dim. An empty <fix> prints nothing.
func (p Printer) Fix(fix string) {
	switch {
	case fix == "":
	case p.Style:
		_, _ = fmt.Fprintf(p.W, "%s  → %s%s\n", dim, Capital(fix), reset)
	default:
		_, _ = fmt.Fprintf(p.W, "  -> %s\n", Capital(fix))
	}
}

// Error prints an error message. A message of the form `<problem> - <fix>`
// prints the fix on its own line.
func (p Printer) Error(msg string) {
	problem, fix, _ := strings.Cut(msg, " - ")
	p.ErrorFix(problem, fix)
}

// Capital gives <s> with an upper case first letter.
func Capital(s string) string {
	for i, r := range s {
		return string(unicode.ToUpper(r)) + s[i+len(string(r)):]
	}
	return s
}

// Mark gives the mark of <kind> in a change list: its colored glyph with the
// style on, else its word padded to 7 characters.
func Mark(kind Kind, style bool) string {
	k := kinds[kind]
	if !style {
		return fmt.Sprintf("%-7s", k.word)
	}
	return k.color + k.glyph + reset
}

// GlyphParts gives the glyph of a status line for the name "ok", "error",
// "warn", "info", "none" or "required": its visible text and its painted
// form. With the style off, the glyph is a word in brackets and both are the
// same. It is for a line that a command builds by itself.
func GlyphParts(name string, style bool) (visible, painted string) {
	glyph, color, plain := "·", dim, "[INFO]"
	switch name {
	case "ok":
		glyph, color, plain = "✔", green, "[OK]"
	case "error":
		glyph, color, plain = "✗", red, "[ERROR]"
	case "warn":
		glyph, color, plain = "!", yellow, "[WARN]"
	case "none":
		glyph, color, plain = "○", yellow, "[WARN]"
	case "required":
		glyph, color, plain = "●", "", "*"
	}
	if !style {
		return plain, plain
	}
	if color == "" {
		return glyph, glyph
	}
	return glyph, color + glyph + reset
}

// Glyph gives the painted glyph of GlyphParts.
func Glyph(name string, style bool) string {
	_, painted := GlyphParts(name, style)
	return painted
}

// Arrow gives the mark of a later copy of a subtree in a tree.
func Arrow(style bool) string {
	if style {
		return "↑"
	}
	return "^"
}

// Bold gives <s> in bold.
func Bold(s string) string { return bold + s + reset }
