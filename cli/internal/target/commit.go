package target

import (
	"os"
	"regexp"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// tableLine is a TOML table header line, as the awk of bash finds it.
var tableLine = regexp.MustCompile(`^[ \t]*\[`)

var (
	blankLine   = regexp.MustCompile(`^[ \t]*$`)
	commentLine = regexp.MustCompile(`^[ \t]*#`)
)

// awkLines splits <b> into the records of awk: a last line without a newline
// is a record, and a newline at the end gives no empty record.
func awkLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// RecordCommit is _se_record_commit: it sets the top-level <key> of the TOML
// file <file> to <commit>. A line that has the key before the first table is
// replaced, and a line with the same value causes no write. Else the key and
// its <comment> go after the header comments, when a blank line ends them,
// or else on the first line. Each other line stays as it is. Through a
// symlink, the real file is written, so the link stays. A <file> that is not
// there is created with <header> and the key; without <header>, that is an
// error. An empty <comment> adds no comment line. With a dry run, it only
// prints the change. It gives 1, after a log line, when it cannot write.
func (e *Env) RecordCommit(file, key, commit, comment, header string) fail.Code {
	r := e.Report
	show := e.Show(file)
	line := key + ` = "` + commit + `"`
	cannot := func() fail.Code {
		r.Log("cannot write " + show)
		return fail.Other
	}
	real := file
	for hops := 1; isLink(real); hops++ {
		if hops > 40 {
			return cannot()
		}
		link, err := os.Readlink(real)
		if err != nil || link == "" {
			return cannot()
		}
		if strings.HasPrefix(link, "/") {
			real = link
		} else {
			real = Dirname(real) + "/" + link
		}
	}
	isFile := isRegular(real)
	if !isFile && header == "" {
		return cannot()
	}
	var old []byte
	var readErr error
	if isFile {
		// When awk cannot read the file, bash goes on, and the write fails.
		old, readErr = os.ReadFile(real)
		for _, l := range awkLines(old) {
			if tableLine.MatchString(l) {
				break
			}
			if l == line {
				return fail.OK
			}
		}
	}
	short := commit
	if len(short) > 7 {
		short = short[:7]
	}
	if e.DryRun {
		e.Plan(real)
		if r.Verbose() {
			r.Log("would set " + key + " " + commit + " in " + show)
		} else {
			r.Info("Would set " + key + " to " + short)
		}
		return fail.OK
	}
	tmp, err := e.Temps.CreateNext(real)
	if err != nil {
		return cannot()
	}
	mode := os.FileMode(0o644)
	var out strings.Builder
	if isFile && readErr != nil {
		os.Remove(tmp)
		return cannot()
	}
	if isFile {
		if m, err := Mode(real); err == nil {
			mode = m
		}
		insert(&out, awkLines(old), key, line, comment)
	} else {
		out.WriteString(header + "\n\n")
		if comment != "" {
			out.WriteString(comment + "\n")
		}
		out.WriteString(line + "\n")
	}
	if err := os.WriteFile(tmp, []byte(out.String()), 0o600); err != nil {
		os.Remove(tmp)
		return cannot()
	}
	if err := os.Chmod(tmp, mode); err != nil {
		os.Remove(tmp)
		return cannot()
	}
	if err := Move(tmp, real); err != nil {
		os.Remove(tmp)
		return cannot()
	}
	if r.Verbose() {
		r.Log("set " + key + " " + commit + " in " + show)
	} else {
		r.Info("Set " + key + " to " + short)
	}
	return fail.OK
}

// insert is the awk program of _se_record_commit: the first pass finds the
// key before the first table, and the line to insert before (after the first
// blank line that ends the header comments, else the first line); the second
// pass writes the lines.
func insert(out *strings.Builder, lines []string, key, line, comment string) {
	keyLine := regexp.MustCompile("^" + regexp.QuoteMeta(key) + "[ \t]*=")
	tables, has, header, found := false, false, false, false
	at := 0
	for i, l := range lines {
		n := i + 1
		if tableLine.MatchString(l) {
			tables = true
		}
		if !tables && keyLine.MatchString(l) {
			has = true
		}
		if n == 1 {
			header = commentLine.MatchString(l)
			at = 1
		}
		if header && !found && blankLine.MatchString(l) {
			at = n + 1
			found = true
		}
		if header && !found && !commentLine.MatchString(l) {
			header = false
		}
	}
	done := false
	for i, l := range lines {
		n := i + 1
		if has && !done && keyLine.MatchString(l) {
			out.WriteString(line + "\n")
			done = true
			continue
		}
		if !has && !done && n == at {
			if comment != "" {
				out.WriteString(comment + "\n")
			}
			out.WriteString(line + "\n")
			done = true
		}
		out.WriteString(l + "\n")
	}
	if !done {
		if comment != "" {
			out.WriteString(comment + "\n")
		}
		out.WriteString(line + "\n")
	}
}
