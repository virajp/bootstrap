package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/field"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/questions"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/tui"
)

// init is `bootstrap [mac] init [--no-tui]`. In a terminal, it opens the
// TUI. With --no-tui, --yes, --dry-run or --answer (they come from an agent
// or a script), or with no terminal, it migrates an old form, or creates the
// values file with a table for each member of the pack base, asks the
// questions without an answer, and syncs.
func (a *app) init(mac bool, o opts) fail.Code {
	if a.interactive() && !o.noTUI && !o.yes && !o.dryRun && len(o.answers) == 0 {
		sd, code := a.prepare(mac, true)
		if code != fail.OK {
			return code
		}
		return a.runTUI(sd, o, true)
	}
	sd, c, code := a.side(mac, true, mac)
	if c != nil {
		defer c.Cleanup()
	}
	if code != fail.OK {
		return code
	}
	return a.finish(sd, o, core.List{Init: true})
}

// tui is `bootstrap [mac] tui`: the TUI on the values file of the side. It
// needs a terminal. mac tui finds macos.toml with --repo-path, then
// MACOS_SETUP_REPO_PATH, then in the current folder.
func (a *app) tui(mac bool, o opts) fail.Code {
	if !a.interactive() {
		a.fail(cmdName(mac, "tui")+" needs a terminal", "use: "+cmdName(mac, "init")+" --no-tui")
		return fail.Usage
	}
	getenv := a.getenv
	if mac {
		getenv = func(k string) string {
			if v := a.getenv(k); v != "" || k != "MACOS_SETUP_REPO_PATH" {
				return v
			}
			return a.pwd
		}
	}
	sd, err := core.Prepare(mac, false, a.pwd, getenv)
	if err != nil {
		return a.logErr(err)
	}
	return a.runTUI(sd, o, false)
}

// runTUI takes the clone of the side <sd> and runs the TUI on it. With
// <initCmd>, a missing values file or its old form is the first change.
func (a *app) runTUI(sd core.Side, o opts, initCmd bool) fail.Code {
	c, code := a.openClone(sd.Mac, false)
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	sd.Clone = c.Dir
	code, err := tui.Run(tui.Options{
		Side: sd, Init: initCmd, NoBootstrap: o.noBootstrap,
		Home: a.getenv("HOME"), Stdin: a.stdin, Stdout: a.stdout,
	})
	if err != nil {
		a.printErr(err)
	}
	return code
}

// side does the preflight of a command of the side and takes its clone. The
// caller defers c.Cleanup when c is not nil. With <hold>, the lock of the
// clone stays held to the end.
func (a *app) side(mac, initCmd, hold bool) (core.Side, *core.Clone, fail.Code) {
	sd, code := a.prepare(mac, initCmd)
	if code != fail.OK {
		return sd, nil, code
	}
	c, code := a.openClone(hold, true)
	sd.Clone = c.Dir
	return sd, c, code
}

// finish plans the change list <l> with core.Plan, asks the questions
// without an answer, shows the change list, and after the consent applies it
// with core.Apply: it writes the changes and syncs.
func (a *app) finish(sd core.Side, o opts, l core.List) fail.Code {
	for _, ans := range o.answers {
		l.Answers = append(l.Answers, core.FlagAnswer(ans))
	}
	opt := core.Options{DryRun: o.dryRun, NoBootstrap: o.noBootstrap, Stdout: a.stdout, Stderr: a.stderr}
	p, err := core.Plan(sd, l, opt)
	for _, w := range p.Warnings {
		a.warnf("%s", w)
	}
	if err != nil {
		return a.logErr(err)
	}
	if len(p.Pending) > 0 {
		answered, code := a.ask(sd.Mac, p.Sel, o, p.AskOnly)
		if code != fail.OK {
			return code
		}
		l.Answers = nil
		for _, ans := range answered {
			l.Answers = append(l.Answers, core.Answer{Name: ans[0] + "." + ans[1], Value: ans[2]})
		}
		if p, err = core.Plan(sd, l, opt); err != nil {
			return a.logErr(err)
		}
	}
	a.reasons = append(a.reasons, p.Reasons...)
	for _, s := range p.Skipped {
		a.infof("%s", skipText(s))
	}

	if p.Empty() {
		a.nothingToChange()
		return fail.OK
	}
	var entries []entry
	if p.Before != nil {
		entries = append(entries, entry{step: true, path: p.Before.Shown(), note: p.Before.Note})
	}
	for _, c := range p.Changes {
		kind := report.Changed
		switch {
		case c.Remove:
			kind = report.Removed
		case !core.Exists(c.Path):
			kind = report.New
		}
		entries = append(entries, entry{kind: kind, path: a.show(c.Path), note: c.Note})
	}
	for _, f := range p.Files {
		entries = append(entries, entry{kind: f.Kind, path: a.show(f.Path)})
	}
	for _, r := range a.reasons {
		a.infof("%s", r)
	}
	a.printChanges(entries)
	if o.dryRun {
		a.warnDirty(p.Dirty())
		a.keptSeeds(p.Seeds)
		a.printer().Success("Dry run: " + count(len(entries), "change") + ", nothing written")
		return fail.OK
	}
	if code := a.stopDirty(p.Dirty()); code != fail.OK {
		return code
	}
	if done, code := a.consent(o.yes, len(entries)); done {
		return code
	}
	opt.Step = a.step
	code, putBack, err := core.Apply(p, opt)
	if err != nil {
		a.logErr(err)
		return code
	}
	if putBack {
		a.infof("Put back %s as it was before the command", a.show(p.Values))
	}
	a.keptSeeds(p.Seeds)
	return code
}

// skipText gives the line of a member of a pack that the command leaves out.
func skipText(s core.Skip) string {
	switch s.Why {
	case "hosts":
		return s.Tool + " is left out: its hosts leave out this host"
	case "selectable":
		return s.Tool + " is left out: it is not selectable; a tool that needs it selects it"
	case "root_only":
		return s.Tool + " is left out: it is root_only, and this repository is a submodule"
	}
	return s.Tool + " is left out"
}

// step prints a step of core.Apply.
func (a *app) step(s core.Step) {
	switch s.Kind {
	case core.Ran:
		_, _ = a.stderr.Write(s.Stderr)
		a.infof("Ran %s", s.Path)
	case core.Wrote:
		a.infof("Wrote %s", a.show(s.Path))
	case core.Removed:
		a.infof("Removed %s", a.show(s.Path))
	}
}

// entry is one line of the change list.
type entry struct {
	kind report.Kind
	step bool // A script that runs before the writes.
	path string
	note string
}

// printChanges prints the change list: a blank line, the header, one line for
// each entry with its glyph, the notes in one column, and a blank line.
func (a *app) printChanges(entries []entry) {
	width := 0
	for _, e := range entries {
		width = max(width, utf8.RuneCountInString(e.path))
	}
	_, _ = fmt.Fprintln(a.stdout)
	_, _ = fmt.Fprintln(a.stdout, "Changes")
	for _, e := range entries {
		mark := report.Mark(e.kind, a.styled)
		if e.step {
			mark = "▸"
			if !a.styled {
				mark = "run    "
			}
		}
		line := "  " + mark + " " + e.path
		if e.note != "" {
			line += strings.Repeat(" ", width-utf8.RuneCountInString(e.path)+2) + e.note
		}
		_, _ = fmt.Fprintln(a.stdout, line)
	}
	_, _ = fmt.Fprintln(a.stdout)
}

// nothingToChange ends a command that has nothing to do. The reasons of the
// command follow the first words.
func (a *app) nothingToChange() {
	text := "Nothing to change"
	for _, r := range a.reasons {
		text += " · " + r
	}
	a.printer().Success(text)
}

// keptSeeds lists each seed of a removed tool: the seed stays.
func (a *app) keptSeeds(seeds []string) {
	for _, p := range seeds {
		if !core.Exists(p) {
			continue
		}
		a.infof("kept seed %s of a removed tool - it is yours now; remove it when you do not need it", a.show(p))
	}
}

// ask gives the answer of each question of the selected tools that has no
// answer, and of each --answer, as (tool, key, value). An empty --answer is
// only for an optional or a group question. With --yes, a question without a
// key and an --answer is an error, and a group without an answer that is not
// empty, too. Without a terminal, too. In a terminal, it asks each such
// question in a field that starts with its proposal, and each question of a
// group without an answer that is not empty, until the group has one. With
// <only>, it asks only the questions of those tools.
func (a *app) ask(mac bool, sel *core.Selected, o opts, only map[string]bool) ([][3]string, fail.Code) {
	all := questions.Of(sel.Idx, sel.Tools)
	var flags []core.Answer
	for _, ans := range o.answers {
		flags = append(flags, core.FlagAnswer(ans))
	}
	given, err := core.CheckAnswers(sel, flags)
	if err != nil {
		return nil, a.logErr(err)
	}
	rc := fail.OK
	dir := a.home(mac)
	terminal := !o.yes && a.interactive()
	// The answers that the command writes: the flags, then the fields.
	answers := map[string]string{}
	for name, v := range given {
		answers[name] = v
	}
	filled := func(q questions.Question) bool {
		if v, ok := answers[q.Name()]; ok {
			return v != ""
		}
		v, ok := questions.Answer(sel.User, sel.Over, q)
		return ok && !questions.Empty(v)
	}
	groupEmpty := func(q questions.Question) bool {
		for _, m := range questions.Members(all, q) {
			if filled(m) {
				return false
			}
		}
		return true
	}
	// The groups without an answer that is not empty in the files and the
	// flags, and the last question of each group.
	open := map[string]bool{}
	last := map[string]int{}
	for i, qq := range all {
		if qq.Group != "" {
			open[qq.GroupName()] = groupEmpty(qq)
			last[qq.GroupName()] = i
		}
	}
	for i, qq := range all {
		if only != nil && !only[qq.Tool] {
			continue
		}
		name := qq.Name()
		inOpen := qq.Group != "" && open[qq.GroupName()]
		_, flagged := given[name]
		_, inFiles := questions.Answer(sel.User, sel.Over, qq)
		if !flagged && (!inFiles || inOpen) {
			switch {
			case terminal:
				v, code := a.askField(qq, dir)
				if code != fail.OK {
					return nil, code
				}
				answers[name] = v
			case !inOpen:
				hint := fmt.Sprintf("give it with --answer %s=<value>", name)
				if !qq.Necessary() {
					hint += fmt.Sprintf(", or leave it empty with --answer %s=", name)
				}
				a.fail(fmt.Sprintf("the question %s has no answer", name), hint)
				rc = fail.Usage
			}
		}
		if !inOpen || last[qq.GroupName()] != i {
			continue
		}
		members := questions.Members(all, qq)
		var keys []string
		for _, m := range members {
			keys = append(keys, m.Key)
		}
		if !terminal {
			a.fail(fmt.Sprintf("the group %s has no answer", qq.GroupName()), "give at least one with --answer <tool>.<key>=<value>: "+strings.Join(keys, ", "))
			rc = fail.Usage
			continue
		}
		for groupEmpty(qq) {
			_, _ = fmt.Fprintf(a.stdout, "Give at least one answer: %s\n", strings.Join(keys, ", "))
			for _, m := range members {
				v, code := a.askField(m, dir)
				if code != fail.OK {
					return nil, code
				}
				answers[m.Name()] = v
			}
		}
	}
	if rc != fail.OK {
		return nil, rc
	}
	var out [][3]string
	for _, qq := range all {
		if v, ok := answers[qq.Name()]; ok {
			out = append(out, [3]string{qq.Tool, qq.Key, v})
		}
	}
	return out, fail.OK
}

// askField asks <qq> in a field that starts with its proposal. An interrupt
// gives code 2.
func (a *app) askField(qq questions.Question, dir string) (string, fail.Code) {
	proposal, _ := questions.Proposal(qq, dir)
	spec := field.Spec{Question: qq.Text + " (" + qq.Name() + ")", Note: qq.Note, Proposal: proposal, Necessary: qq.Necessary()}
	ask := a.prompt
	if ask == nil {
		ask = func(s field.Spec) (string, error) { return field.Ask(a.stdin, a.stdout, s) }
	}
	v, err := ask(spec)
	if err != nil {
		a.errorf("the question %s has no answer", qq.Name())
		return "", fail.Usage
	}
	return v, fail.OK
}

// home gives the folder where the proposals of the questions come from:
// $HOME on the Mac, else the repository.
func (a *app) home(mac bool) string {
	if mac {
		return a.getenv("HOME")
	}
	return a.getenv("MISE_PROJECT_ROOT")
}

// questions is `bootstrap [mac] questions [tool...]`.
func (a *app) questions(mac bool, o opts) fail.Code {
	sd, c, code := a.side(mac, false, false)
	if c != nil {
		defer c.Cleanup()
	}
	if code != fail.OK {
		return code
	}
	if core.Exists(sd.Values) {
		raw, err := source.ReadTOML(sd.Values)
		if err != nil {
			a.errorf("cannot read %s", a.show(sd.Values))
			return fail.Usage
		}
		if mac {
			err = oldform.Mac(raw, a.show(sd.Values))
		} else {
			err = oldform.Repo(sd.Root, raw, a.show(sd.Values))
		}
		if err != nil {
			return a.logErr(err)
		}
	}
	sel, err := core.SelectionOf(sd, sd.Values, sd.Values)
	if err != nil {
		return a.logErr(err)
	}
	return a.printQuestions(mac, sel, o)
}

// printQuestions prints the questions of the tools of <sel>, or of the tools
// that <o> names, as JSON or as text.
func (a *app) printQuestions(mac bool, sel *core.Selected, o opts) fail.Code {
	tools := sel.Tools
	if len(o.names) > 0 {
		for _, n := range o.names {
			if sel.Idx.Manifests[n] == nil {
				a.errorf("%s is not a bootstrap tool", n)
				return fail.Missing
			}
		}
		tools = o.names
	}
	type row struct {
		Tool     string  `json:"tool"`
		Key      string  `json:"key"`
		Question string  `json:"question"`
		Proposal *string `json:"proposal"`
		Source   *string `json:"source"`
		Answer   any     `json:"answer"`
		Group    string  `json:"group"`
		Note     string  `json:"note"`
		Optional bool    `json:"optional"`
		// state is the text state of a question without an answer that is
		// not an error: "optional" or "group".
		state string
		// gone: a question of a group without an answer that is not empty.
		gone bool
	}
	rows := []row{}
	dir := a.home(mac)
	all := questions.Of(sel.Idx, tools)
	for _, qq := range all {
		r := row{Tool: qq.Tool, Key: qq.Key, Question: qq.Text, Group: qq.Group, Note: qq.Note, Optional: qq.Optional}
		if qq.Group != "" {
			r.gone = !questions.GroupAnswered(all, qq, sel.User, sel.Over)
			if !r.gone {
				r.state = "group"
			}
		} else if qq.Optional {
			r.state = "optional"
		}
		if v, from := questions.Proposal(qq, dir); from != "" {
			r.Proposal, r.Source = &v, &from
		}
		if v, ok := questions.Answer(sel.User, sel.Over, qq); ok {
			r.Answer = v
		}
		rows = append(rows, r)
	}
	if o.json {
		b, _ := json.MarshalIndent(rows, "", "  ")
		_, _ = fmt.Fprintln(a.stdout, string(b))
		return fail.OK
	}
	out := report.NewPrinter(a.stdout, a.styled)
	if len(rows) == 0 {
		out.Success("No questions")
		return fail.OK
	}
	width := 0
	for _, r := range rows {
		width = max(width, utf8.RuneCountInString(r.Key))
	}
	pad := strings.Repeat(" ", width+5)
	tool, answered, others := "", 0, 0
	for _, r := range rows {
		if r.Tool != tool {
			if tool != "" {
				_, _ = fmt.Fprintln(a.stdout)
			}
			tool = r.Tool
			_, _ = fmt.Fprintln(a.stdout, tool)
		}
		_, _ = fmt.Fprintf(a.stdout, "  %-*s   %s\n", width, r.Key, r.Question)
		answer := ""
		switch {
		case r.Answer != nil && !r.gone:
			answered++
			answer = fmt.Sprint(r.Answer)
			line := pad + report.Glyph("ok", a.styled)
			if answer != "" {
				line += " " + shortValue(answer)
			}
			_, _ = fmt.Fprintln(a.stdout, line)
		case r.state == "optional":
			others++
			_, _ = fmt.Fprintf(a.stdout, "%s%s no answer (optional)\n", pad, report.Glyph("info", a.styled))
		case r.state == "group":
			others++
			_, _ = fmt.Fprintf(a.stdout, "%s%s no answer (group %s has an answer)\n", pad, report.Glyph("info", a.styled), r.Group)
		default:
			_, _ = fmt.Fprintf(a.stdout, "%s%s no answer\n", pad, report.Glyph("error", a.styled))
		}
		if r.Proposal != nil && *r.Proposal != answer {
			_, _ = fmt.Fprintf(a.stdout, "%s%s proposal (%s): %s\n", pad, report.Glyph("info", a.styled), *r.Source, shortValue(*r.Proposal))
		}
	}
	_, _ = fmt.Fprintln(a.stdout)
	text := count(len(rows), "question")
	if answered > 0 {
		text += fmt.Sprintf(" · %d answered", answered)
	}
	if without := len(rows) - answered - others; without > 0 {
		text += fmt.Sprintf(" · %d without an answer", without)
	}
	out.Info(text)
	return fail.OK
}

// shortValue gives <v>, or for a value longer than 40 characters its first
// 20 characters, "…" and its last 4 characters.
func shortValue(v string) string {
	r := []rune(v)
	if len(r) <= 40 {
		return v
	}
	return string(r[:20]) + "…" + string(r[len(r)-4:])
}
