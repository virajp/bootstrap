// Package report prints the output of a run: the list lines, the verbose log
// lines, the info lines, the table and the last line. It writes the same bytes
// as the output functions of the bash file common.
//
// Contract: docs/design/cli.md ("The output of a run").
package report

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Kind is the outcome of one target.
type Kind int

// The outcome kinds, in the order of the table rows.
const (
	Changed Kind = iota
	New
	Seeded
	Removed
	Unchanged
	Kept
	Skipped
	Failed
	nKinds
)

const (
	yellow = "\033[0;33m"
	green  = "\033[0;32m"
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

// The look of each kind: the color, the glyph and the word of its list line,
// the word of a dry run, and the label of its table row. A kind without a word
// has no list line.
var kinds = [nKinds]struct {
	color, glyph, word, dryWord, label string
}{
	Changed:   {yellow, "✎", "changed", "would change", "Changed"},
	New:       {green, "✚", "new", "would add", "New"},
	Seeded:    {blue, "❋", "seeded", "would seed", "Seeded"},
	Removed:   {red, "✖", "removed", "would remove", "Removed"},
	Unchanged: {"", "=", "", "", "Unchanged"},
	Kept:      {"", "◦", "", "", "Kept seeds"},
	Skipped:   {"", "⊘", "", "", "Skipped"},
	Failed:    {red, "✗", "failed", "failed", "Failed"},
}

// Options sets the look of a Report.
type Options struct {
	// Prefix is the name at the start of each log line. Empty gives
	// "bootstrap".
	Prefix string
	// Style turns the glyphs and the colors on. See StyleFromEnv.
	Style bool
	// Verbose prints the log line of each target in place of its list line.
	// See VerboseFromEnv.
	Verbose bool
}

// Report counts the outcomes of a run and prints its lines to one writer,
// the stderr of the engine.
type Report struct {
	w      io.Writer
	opt    Options
	dryRun bool
	counts [nKinds]int
	seen   map[string]bool
	listed bool
}

// NewReport gives a Report that writes to <w>. It starts as a real run.
func NewReport(w io.Writer, opt Options) *Report {
	if opt.Prefix == "" {
		opt.Prefix = "bootstrap"
	}
	r := &Report{w: w, opt: opt}
	r.Start(false)
	return r
}

// StyleFromEnv applies the style rule: FORCE_COLOR (not "0") turns the style
// on, NO_COLOR turns it off, else it needs a terminal on stderr and a TERM
// that is not "dumb". <getenv> is os.Getenv in the engine.
func StyleFromEnv(getenv func(string) string, stderrIsTerminal bool) bool {
	if v := getenv("FORCE_COLOR"); v != "" && v != "0" {
		return true
	}
	if getenv("NO_COLOR") != "" {
		return false
	}
	term := getenv("TERM")
	if term == "" {
		term = "dumb"
	}
	return stderrIsTerminal && term != "dumb"
}

// IsTerminal reports whether <f> is a character device, as the bash test
// `-t 2` does for a terminal. Another character device, for example
// /dev/null, also gives true, but no output of the engine there can be seen.
func IsTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// VerboseFromEnv is true when BOOTSTRAP_VERBOSE is "1".
func VerboseFromEnv(getenv func(string) string) bool {
	return getenv("BOOTSTRAP_VERBOSE") == "1"
}

// Start starts the counts of a run. <dryRun> gives the words of a dry run.
func (r *Report) Start(dryRun bool) {
	r.dryRun = dryRun
	r.counts = [nKinds]int{}
	r.seen = map[string]bool{}
	r.listed = false
}

// Count gives the count of <kind>.
func (r *Report) Count(kind Kind) int { return r.counts[kind] }

// Verbose is true when the report prints the log line of each target.
func (r *Report) Verbose() bool { return r.opt.Verbose }

// Log prints a log line with the prefix.
func (r *Report) Log(msg string) {
	fmt.Fprintf(r.w, "%s: %s\n", r.opt.Prefix, msg)
}

// Info prints an info line.
func (r *Report) Info(text string) {
	if r.opt.Style {
		fmt.Fprintf(r.w, "ℹ️ %s\n", text)
	} else {
		fmt.Fprintf(r.w, "[INFO] %s\n", text)
	}
}

// Outcome reports one target: it counts <kind> and prints its line. <text> is
// the reason of a failed target, and it does not repeat <path>. <line> is the
// log line that prints with verbose; without it, verbose prints nothing. A
// target counts once as failed and once as removed: a later report of it
// prints only its <line>.
func (r *Report) Outcome(kind Kind, path, text, line string) {
	seen := false
	if kind == Failed || kind == Removed {
		key := fmt.Sprintf("%d %s", kind, path)
		seen = r.seen[key]
		r.seen[key] = true
	}
	if !seen {
		r.counts[kind]++
	}
	if r.opt.Verbose {
		if line != "" {
			r.Log(line)
		}
	} else if !seen {
		r.listLine(kind, path, text)
	}
}

// listLine prints the list line of <kind>, when the list has that kind.
func (r *Report) listLine(kind Kind, path, text string) {
	k := kinds[kind]
	if k.word == "" {
		return
	}
	word, width := k.word, 9
	if kind != Failed && r.dryRun {
		word, width = k.dryWord, 12
	}
	if !r.opt.Style {
		if kind == Failed {
			fmt.Fprintf(r.w, "%s %s: %s\n", word, path, text)
		} else {
			fmt.Fprintf(r.w, "%s %s\n", word, path)
		}
		return
	}
	if !r.listed {
		r.listed = true
		fmt.Fprint(r.w, "\n")
	}
	if kind == Failed {
		fmt.Fprintf(r.w, "  %s%s %s %s: %s%s\n", k.color, k.glyph, word, path, text, reset)
	} else {
		fmt.Fprintf(r.w, "  %s%s %-*s %s%s\n", k.color, k.glyph, width, word, path, reset)
	}
}

// tableRow prints one row of the table. A row with a zero count is dim.
func (r *Report) tableRow(kind Kind) {
	k, count := kinds[kind], r.counts[kind]
	if !r.opt.Style {
		fmt.Fprintf(r.w, "| %-13s | %5d |\n", k.label, count)
		return
	}
	color, off := k.color, ""
	if count == 0 {
		color = dim
	}
	if color != "" {
		off = reset
	}
	fmt.Fprintf(r.w, "  │ %s%s %-11s%s │ %s%5d%s │\n", color, k.glyph, k.label, off, color, count, off)
}

// Summary prints the table of the counts and the last line of a run that
// ended with <rc>.
func (r *Report) Summary(rc int) {
	top, middle, bottom := "+---------------+-------+", "+---------------+-------+", "+---------------+-------+"
	if r.opt.Style {
		top = "  ┌───────────────┬───────┐"
		middle = "  ├───────────────┼───────┤"
		bottom = "  └───────────────┴───────┘"
		fmt.Fprint(r.w, "\n")
	}
	n := r.counts[Changed] + r.counts[New] + r.counts[Seeded] + r.counts[Removed]
	total := 0
	for _, c := range r.counts {
		total += c
	}
	fmt.Fprintln(r.w, top)
	for kind := Changed; kind < nKinds; kind++ {
		r.tableRow(kind)
	}
	fmt.Fprintln(r.w, middle)
	if r.opt.Style {
		fmt.Fprintf(r.w, "  │ %-13s │ %5d │\n", "Total", total)
	} else {
		fmt.Fprintf(r.w, "| %-13s | %5d |\n", "Total", total)
	}
	fmt.Fprintln(r.w, bottom)

	if rc != 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "The sync failed with code %d", rc)
		switch failed := r.counts[Failed]; {
		case failed == 1:
			b.WriteString(": 1 file failed")
		case failed > 1:
			fmt.Fprintf(&b, ": %d files failed", failed)
		}
		if r.opt.Style {
			fmt.Fprintf(r.w, "%s❌ %s%s\n", red, b.String(), reset)
		} else {
			fmt.Fprintf(r.w, "[ERROR] %s\n", b.String())
		}
		return
	}
	files := "files"
	if n == 1 {
		files = "file"
	}
	var text string
	switch {
	case r.dryRun && n == 0:
		text = "Dry run: no file would change"
	case r.dryRun:
		text = fmt.Sprintf("Dry run: %d %s would change", n, files)
	case n == 0:
		text = "Synced: no file changed"
	default:
		text = fmt.Sprintf("Synced: %d %s changed", n, files)
	}
	if r.opt.Style {
		fmt.Fprintf(r.w, "✅ %s\n", text)
	} else {
		fmt.Fprintf(r.w, "[OK] %s\n", text)
	}
}
