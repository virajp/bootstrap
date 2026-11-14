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
	cannot := func() fail.Code {
		r.Error("cannot write " + show)
		return fail.Other
	}
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
		return cannot()
	default:
		if m, err := Mode(real); err == nil {
			mode = m
		}
	}
	var exprs []string
	for _, n := range names {
		b, _ := json.Marshal(n)
		exprs = append(exprs, ".tools["+string(b)+"] = (.tools["+string(b)+"] // {})")
	}
	tmp, err := e.Temps.CreateNext(real)
	if err != nil {
		return cannot()
	}
	if err := os.WriteFile(tmp, in, 0o600); err != nil {
		os.Remove(tmp)
		return cannot()
	}
	out, _, err := format.YQ("-p", "toml", "-o", "toml", strings.Join(exprs, " | "), tmp)
	if err != nil {
		os.Remove(tmp)
		return cannot()
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
	if h := lead.String(); h != "" && !strings.HasPrefix(out, h) {
		out = h + "\n" + out
	}
	if err := os.WriteFile(tmp, []byte(out), 0o600); err != nil {
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
	for _, n := range names {
		line(n, "added to")
	}
	return fail.OK
}
