package target

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// AddTables writes an empty `[tools.<name>]` table into the values file
// <file> for each name of <names>, so its comments stay (the yq edit of the
// migration). A <file> that is not there is created from <head>; without
// <head>, that is an error. Through a symlink, the real file is written. It
// logs one line for each table. With a dry run, or with <listOnly> (check), it
// writes nothing and only logs the tables; a dry run also plans the file. It
// gives 1, after a log line, when it cannot write.
func (e *Env) AddTables(file string, names []string, head string, listOnly bool) fail.Code {
	if len(names) == 0 {
		return fail.OK
	}
	r := e.Report
	show := e.Show(file)
	line := func(name, verb string) {
		text := "[tools." + name + "] " + verb + " " + show
		if r.Verbose() {
			r.Log(text)
		} else {
			r.Info(text)
		}
	}
	if listOnly || e.DryRun {
		if e.DryRun && !listOnly {
			real := file
			if p, err := filepath.EvalSymlinks(file); err == nil {
				real = p
			}
			e.Plan(real, report.Changed)
		}
		for _, n := range names {
			line(n, "would be added to")
		}
		return fail.OK
	}
	var exprs []string
	for _, n := range names {
		b, _ := json.Marshal(n)
		exprs = append(exprs, ".tools["+string(b)+"] = (.tools["+string(b)+"] // {})")
	}
	if !e.editTOML(file, head, exprs) {
		r.Error("cannot write " + show)
		return fail.Other
	}
	for _, n := range names {
		line(n, "added to")
	}
	return fail.OK
}

// editTOML applies the yq expressions <exprs> to the TOML file <file>, so
// its comments stay. A <file> that is not there is created from <head>;
// without <head>, that is an error. Through a symlink, the real file is
// written. It tells if the write succeeded.
func (e *Env) editTOML(file, head string, exprs []string) bool {
	real := file
	if p, err := filepath.EvalSymlinks(file); err == nil {
		real = p
	}
	in, err := os.ReadFile(real)
	mode := os.FileMode(0o644)
	switch {
	case os.IsNotExist(err) && head != "":
		in = []byte(head)
	case err != nil:
		return false
	default:
		if m, err := Mode(real); err == nil {
			mode = m
		}
	}
	tmp, err := e.Temps.CreateNext(real)
	if err != nil {
		return false
	}
	if err := os.WriteFile(tmp, in, 0o600); err != nil {
		_ = os.Remove(tmp)
		return false
	}
	out, _, err := format.YQ("-p", "toml", "-o", "toml", strings.Join(exprs, " | "), tmp)
	if err != nil {
		_ = os.Remove(tmp)
		return false
	}
	// yq drops the comments of a file that has no key: keep its first
	// comment lines.
	var lead strings.Builder
	for _, l := range strings.SplitAfter(string(in), "\n") {
		if !strings.HasPrefix(l, "#") {
			break
		}
		lead.WriteString(l)
	}
	// yq also drops the blank line that ends them, and bootstrap_commit
	// goes after the header comments only when a blank line ends them.
	if h := lead.String(); h != "" {
		blank := strings.HasPrefix(strings.TrimLeft(string(in)[len(h):], " \t"), "\n")
		switch {
		case !strings.HasPrefix(out, h):
			out = h + "\n" + out
		case blank && !strings.HasPrefix(out[len(h):], "\n"):
			out = h + "\n" + out[len(h):]
		}
	}
	if err := os.WriteFile(tmp, []byte(out), 0o600); err != nil {
		_ = os.Remove(tmp)
		return false
	}
	if err := os.Chmod(tmp, mode); err != nil {
		_ = os.Remove(tmp)
		return false
	}
	if err := Move(tmp, real); err != nil {
		_ = os.Remove(tmp)
		return false
	}
	return true
}

// SetTable sets the top-level table <key> of the TOML file <file> to the
// JSON object <value>, so the comments of the file stay. It logs one line.
// With a dry run, it writes nothing and plans the file. It gives 1, after a
// log line, when it cannot write.
func (e *Env) SetTable(file, key, value string) fail.Code {
	r := e.Report
	show := e.Show(file)
	line := func(verb string) {
		text := "[" + key + "] " + verb + " in " + show
		if r.Verbose() {
			r.Log(text)
		} else {
			r.Info(text)
		}
	}
	if e.DryRun {
		real := file
		if p, err := filepath.EvalSymlinks(file); err == nil {
			real = p
		}
		e.Plan(real, report.Changed)
		line("would change")
		return fail.OK
	}
	if !e.editTOML(file, "", []string{"." + key + " = " + value}) {
		r.Error("cannot write " + show)
		return fail.Other
	}
	line("changed")
	return fail.OK
}
