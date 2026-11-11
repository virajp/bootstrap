package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/questions"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/scripts"
	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// The first line of a new values file.
const (
	repoHeader = "# The values of bootstrap for this repository. See docs/sync/index.md of bootstrap."
	macHeader  = "# The values of the Mac sync of bootstrap. See docs/setup/index.md."
)

func realPath(p string) (string, error) { return filepath.EvalSymlinks(p) }

// q is a string as a yq string literal.
func q(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// editTOML runs the yq expressions <exprs> on the TOML text <in>, so its
// comments stay.
func editTOML(in []byte, exprs []string) ([]byte, error) {
	if len(exprs) == 0 {
		return in, nil
	}
	dir, err := os.MkdirTemp("", "bootstrap-init.")
	if err != nil {
		return nil, fail.New(fail.Other, "cannot create a temp folder")
	}
	defer os.RemoveAll(dir)
	p := dir + "/values.toml"
	if err := os.WriteFile(p, in, 0o600); err != nil {
		return nil, fail.New(fail.Other, "cannot write a temp file")
	}
	out, stderr, err := format.YQ("-p", "toml", "-o", "toml", strings.Join(exprs, " | "), p)
	if err != nil {
		return nil, fail.New(fail.Usage, "cannot change the values file: %s", strings.TrimSpace(stderr))
	}
	// yq drops the comments of a file that has no key: keep its first
	// comment lines.
	var head strings.Builder
	for _, line := range strings.SplitAfter(string(in), "\n") {
		if !strings.HasPrefix(line, "#") {
			break
		}
		head.WriteString(line)
	}
	if h := head.String(); h != "" && !strings.HasPrefix(out, h) {
		out = h + "\n" + out
	}
	return []byte(out), nil
}

// selected is the selection of a values file of the side.
type selected struct {
	idx   *source.Index
	tools []string
	user  *source.Obj // The [tools.<name>] tables of the values file.
	over  *source.Obj // The [tools.<name>] tables of its overrides file.
	res   *selection.Resolved
	items []values.Item // The [tools.*] and [packs.*] entries of the values file.
}

// selectionOf reads the values file <file> (the overrides file is next to
// <overridesOf>) and resolves its tools from the clone <dir>.
func (a *app) selectionOf(mac bool, dir, root, file, overridesOf string) (*selected, fail.Code) {
	side := source.Repo
	if mac {
		side = source.Macos
	}
	idx, err := source.Source{Root: dir, Side: side}.Open()
	if err != nil {
		return nil, a.logErr(err)
	}
	s := &selected{idx: idx}
	shown := a.show(file)
	var res *selection.Resolved
	if mac {
		v, err := values.ReadMacos(file, shown)
		if err != nil {
			return nil, a.logErr(err)
		}
		s.user, s.items = v.User, v.Sel
		res, err = selection.ResolveMacos(idx, v)
		if err != nil {
			return nil, a.logErr(err)
		}
	} else {
		v := &values.Repo{Path: file, Raw: source.NewObj(), User: source.NewObj()}
		if exists(file) {
			if v, err = values.ReadRepo(file, shown); err != nil {
				return nil, a.logErr(err)
			}
		}
		s.user, s.items = v.User, v.Sel
		sub, _ := git("-C", root, "rev-parse", "--show-superproject-working-tree")
		res, err = selection.ResolveRepo(idx, v, sub != "")
		if err != nil {
			return nil, a.logErr(err)
		}
	}
	s.tools, s.res = res.Tools, res
	ov := values.OverridesPath(overridesOf)
	s.over, err = values.ReadOverrides(ov, a.show(ov))
	if err != nil {
		return nil, a.logErr(err)
	}
	return s, fail.OK
}

// init is `bootstrap [mac] init [tool...]`.
func (a *app) init(mac bool, o opts) fail.Code {
	w, code := a.prepare(mac, true)
	if code != fail.OK {
		return code
	}
	c, code := a.openClone(mac, true)
	defer c.cleanup()
	if code != fail.OK {
		return code
	}

	// The values file: new, or migrated from its old form.
	cur, err := os.ReadFile(w.values)
	existed := err == nil
	next := cur
	var notes []string
	var changes []oldform.Change
	if !existed {
		if mac {
			next = []byte(macHeader + "\n")
		} else {
			// The templates need repo_name: the name of the folder.
			next = []byte(repoHeader + "\n\nenvironments = [ \"dev\" ]\nrepo_name = " + q(filepath.Base(w.root)) + "\n")
		}
		notes = append(notes, "create the values file")
	} else {
		var migs []oldform.Change
		if mac {
			ch, err := oldform.MigrateMac(w.values)
			if err != nil {
				return a.logErr(err)
			}
			if ch != nil {
				migs = append(migs, *ch)
			}
		} else {
			migs, err = oldform.MigrateRepo(w.root, w.values)
			if err != nil {
				return a.logErr(err)
			}
		}
		for _, ch := range migs {
			if ch.Path == w.values {
				next = ch.Content
				notes = append(notes, ch.Note)
			} else {
				changes = append(changes, ch)
			}
		}
	}

	// The named tools.
	raw := source.NewObj()
	if len(next) > 0 {
		if r, err := readTOMLText(next); err == nil {
			raw = r
		}
	}
	tools, _ := raw.Get("tools").(*source.Obj)
	var exprs []string
	for _, name := range o.names {
		if tools == nil || !tools.Has(name) {
			exprs = append(exprs, ".tools["+q(name)+"] = {}")
			notes = append(notes, "select "+name)
		}
	}
	if next, err = editTOML(next, exprs); err != nil {
		return a.logErr(err)
	}

	return a.finish(mac, o, &change{w: w, c: c, cur: cur, existed: existed, next: next, notes: notes, changes: changes, migrating: true})
}

// change is the change of a writing command: the values file before and
// after, the other changes, and a script that runs after the consent.
type change struct {
	w         where
	c         *clone
	cur       []byte // The values file before the command.
	existed   bool
	next      []byte // The values file after the command, without the answers.
	notes     []string
	exprs     []string // yq expressions for next, before the answers.
	changes   []oldform.Change
	migrating bool            // init: the plan accepts the old form of the targets.
	askOnly   map[string]bool // The tools whose questions the command asks; nil for each tool.
	before    *script         // A script that runs after the consent, before the writes.
	seeds     []string        // The seeds of the removed tools that stay.
}

// script is one script of a tool that a command runs.
type script struct {
	runner scripts.Runner
	job    scripts.Job
	note   string
}

// finish asks the questions, plans the sync with the new values file, shows
// the change list, and after the consent writes the changes and syncs.
func (a *app) finish(mac bool, o opts, ch *change) fail.Code {
	w, c, cur, existed, next, notes, changes := ch.w, ch.c, ch.cur, ch.existed, ch.next, ch.notes, ch.changes
	var err error
	// A temporary copy of the values file, with its overrides file, for the
	// selection and the plan.
	tmp, err := os.MkdirTemp(tmpDir(a.getenv), "bootstrap-init.")
	if err != nil {
		a.errorf("cannot create a temp folder")
		return fail.Other
	}
	defer os.RemoveAll(tmp)
	tmpValues := tmp + "/" + filepath.Base(w.values)
	if err := os.WriteFile(tmpValues, next, 0o644); err != nil {
		a.errorf("cannot write a temp file")
		return fail.Other
	}
	if b, err := os.ReadFile(values.OverridesPath(w.values)); err == nil {
		_ = os.WriteFile(values.OverridesPath(tmpValues), b, 0o644)
	}

	// The questions.
	sel, code := a.selectionOf(mac, c.dir, w.root, tmpValues, w.values)
	if code != fail.OK {
		return code
	}
	answered, code := a.ask(mac, sel, o, ch.askOnly)
	if code != fail.OK {
		return code
	}
	if len(answered) > 0 {
		var exprs []string
		for _, ans := range answered {
			exprs = append(exprs, ".tools["+q(ans[0])+"]["+q(ans[1])+"] = "+q(ans[2]))
			notes = append(notes, "answer "+ans[0]+"."+ans[1])
		}
		if next, err = editTOML(next, exprs); err != nil {
			return a.logErr(err)
		}
		if err := os.WriteFile(tmpValues, next, 0o644); err != nil {
			a.errorf("cannot write a temp file")
			return fail.Other
		}
	}
	if !existed || !bytes.Equal(next, cur) {
		changes = append([]oldform.Change{{Path: w.values, Content: next, Note: strings.Join(notes, "; ")}}, changes...)
	}

	// The plan of the sync, with the new values file.
	run := frontEndOf(mac)
	envKey, envVal := "BOOTSTRAP_VALUES", tmpValues
	if mac {
		envKey, envVal = "MACOS_SETUP_REPO_PATH", tmp
	}
	old, had := os.LookupEnv(envKey)
	_ = os.Setenv(envKey, envVal)
	oldform.Migrating = ch.migrating
	planned, pc := run(syncArgs(mac, opts{dryRun: true, noBootstrap: o.noBootstrap}), io.Discard, io.Discard)
	if pc != fail.OK {
		// Again, with the output, so the messages show.
		_, pc = run(syncArgs(mac, opts{dryRun: true, noBootstrap: o.noBootstrap}), a.stdout, a.stderr)
	}
	oldform.Migrating = false
	if had {
		_ = os.Setenv(envKey, old)
	} else {
		_ = os.Unsetenv(envKey)
	}
	if pc != fail.OK {
		return pc
	}
	var syncs []target.Planned
	seen := map[string]bool{}
	for _, p := range planned {
		if strings.HasPrefix(p.Path, tmp+"/") || seen[p.Path] {
			continue
		}
		seen[p.Path] = true
		syncs = append(syncs, p)
	}
	sort.Slice(syncs, func(i, j int) bool { return syncs[i].Path < syncs[j].Path })

	if len(changes) == 0 && len(syncs) == 0 && ch.before == nil {
		a.nothingToChange()
		return fail.OK
	}
	var entries []entry
	if ch.before != nil {
		entries = append(entries, entry{step: true, path: ch.before.runner.ShownPath(ch.before.job.Tool, ch.before.job.Name), note: ch.before.note})
	}
	for _, c := range changes {
		kind := report.Changed
		switch {
		case c.Remove:
			kind = report.Removed
		case !exists(c.Path):
			kind = report.New
		}
		entries = append(entries, entry{kind: kind, path: a.show(c.Path), note: c.Note})
	}
	for _, p := range syncs {
		entries = append(entries, entry{kind: p.Kind, path: a.show(p.Path)})
	}
	for _, r := range a.reasons {
		a.infof("%s", r)
	}
	a.printChanges(entries)
	paths := append(oldform.Paths(changes), planPaths(syncs)...)
	if o.dryRun {
		a.warnDirty(dirtyTargets(paths, nil))
		a.keptSeeds(ch.seeds)
		a.printer().Success("Dry run: " + count(len(entries), "change") + ", nothing written")
		return fail.OK
	}
	if code := a.stopDirty(dirtyTargets(paths, nil)); code != fail.OK {
		return code
	}
	if done, code := a.consent(o.yes || o.confirmed, len(entries)); done {
		return code
	}
	if ch.before != nil {
		if code := a.runScript(ch.before); code != fail.OK {
			return code
		}
	}
	if err := oldform.Apply(changes); err != nil {
		return a.logErr(err)
	}
	for _, ch := range changes {
		if ch.Remove {
			a.infof("Removed %s", a.show(ch.Path))
		} else {
			a.infof("Wrote %s", a.show(ch.Path))
		}
	}
	_, code = run(syncArgs(mac, opts{noBootstrap: o.noBootstrap}), a.stdout, a.stderr)
	a.keptSeeds(ch.seeds)
	return code
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
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Changes")
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
		fmt.Fprintln(a.stdout, line)
	}
	fmt.Fprintln(a.stdout)
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
		if !exists(p) {
			continue
		}
		a.infof("kept seed %s of a removed tool - it is yours now; remove it when you do not need it", a.show(p))
	}
}

// ask gives the answer of each question of the selected tools that has no
// answer, and of each --answer, as (tool, key, value). With --yes, a question
// without an --answer is an error. Without a terminal, too. With <only>, it
// asks only the questions of those tools.
func (a *app) ask(mac bool, sel *selected, o opts, only map[string]bool) ([][3]string, fail.Code) {
	all := questions.Of(sel.idx, sel.tools)
	byName := map[string]questions.Question{}
	for _, qq := range all {
		byName[qq.Name()] = qq
	}
	given := map[string]string{}
	rc := fail.OK
	for _, ans := range o.answers {
		name, value, ok := strings.Cut(ans, "=")
		if _, known := byName[name]; !ok || !known {
			a.fail(fmt.Sprintf("--answer %s: no question %s of the selected tools", ans, name), "the form is <tool>.<key>=<value>")
			rc = fail.Usage
			continue
		}
		given[name] = value
	}
	if rc != fail.OK {
		return nil, rc
	}
	dir := sel.root(a, mac)
	var out [][3]string
	for _, qq := range all {
		if v, ok := given[qq.Name()]; ok {
			out = append(out, [3]string{qq.Tool, qq.Key, v})
			continue
		}
		if _, ok := questions.Answer(sel.user, sel.over, qq); ok {
			continue
		}
		if only != nil && !only[qq.Tool] {
			continue
		}
		if o.yes || !a.interactive() {
			a.fail(fmt.Sprintf("the question %s has no answer", qq.Name()), fmt.Sprintf("give it with --answer %s=<value>", qq.Name()))
			rc = fail.Usage
			continue
		}
		proposal, _ := questions.Proposal(qq, dir)
		for {
			if proposal != "" {
				fmt.Fprintf(a.stdout, "? %s (%s)\n  [%s]: ", qq.Text, qq.Name(), proposal)
			} else {
				fmt.Fprintf(a.stdout, "? %s (%s)\n  : ", qq.Text, qq.Name())
			}
			line, ok := a.readLine()
			if !ok {
				a.errorf("the question %s has no answer", qq.Name())
				return nil, fail.Usage
			}
			line = strings.TrimSpace(line)
			if line == "" {
				line = proposal
			}
			if line != "" {
				out = append(out, [3]string{qq.Tool, qq.Key, line})
				break
			}
		}
	}
	return out, rc
}

// root gives the folder where a detect command runs: $HOME on the Mac, else
// the repository.
func (s *selected) root(a *app, mac bool) string {
	if mac {
		return a.getenv("HOME")
	}
	return a.getenv("MISE_PROJECT_ROOT")
}

// readTOMLText reads TOML text.
func readTOMLText(b []byte) (*source.Obj, error) {
	f, err := os.CreateTemp("", "bootstrap-values.*.toml")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return nil, err
	}
	f.Close()
	return source.ReadTOML(f.Name())
}

func tmpDir(getenv func(string) string) string {
	if d := getenv("TMPDIR"); d != "" {
		return d
	}
	return "/tmp"
}

// questions is `bootstrap [mac] questions [tool...]`.
func (a *app) questions(mac bool, o opts) fail.Code {
	w, code := a.prepare(mac, false)
	if code != fail.OK {
		return code
	}
	c, code := a.openClone(false, true)
	defer c.cleanup()
	if code != fail.OK {
		return code
	}
	if exists(w.values) {
		raw, err := source.ReadTOML(w.values)
		if err != nil {
			a.errorf("cannot read %s", a.show(w.values))
			return fail.Usage
		}
		if mac {
			err = oldform.Mac(raw, a.show(w.values))
		} else {
			err = oldform.Repo(w.root, raw, a.show(w.values))
		}
		if err != nil {
			return a.logErr(err)
		}
	}
	sel, code := a.selectionOf(mac, c.dir, w.root, w.values, w.values)
	if code != fail.OK {
		return code
	}
	tools := sel.tools
	if len(o.names) > 0 {
		for _, n := range o.names {
			if sel.idx.Manifests[n] == nil {
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
	}
	rows := []row{}
	dir := sel.root(a, mac)
	for _, qq := range questions.Of(sel.idx, tools) {
		r := row{Tool: qq.Tool, Key: qq.Key, Question: qq.Text}
		if v, from := questions.Proposal(qq, dir); from != "" {
			r.Proposal, r.Source = &v, &from
		}
		if v, ok := questions.Answer(sel.user, sel.over, qq); ok {
			r.Answer = v
		}
		rows = append(rows, r)
	}
	if o.json {
		b, _ := json.MarshalIndent(rows, "", "  ")
		fmt.Fprintln(a.stdout, string(b))
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
	tool, answered := "", 0
	for _, r := range rows {
		if r.Tool != tool {
			if tool != "" {
				fmt.Fprintln(a.stdout)
			}
			tool = r.Tool
			fmt.Fprintln(a.stdout, tool)
		}
		fmt.Fprintf(a.stdout, "  %-*s   %s\n", width, r.Key, r.Question)
		answer := ""
		if r.Answer != nil {
			answered++
			answer = fmt.Sprint(r.Answer)
			fmt.Fprintf(a.stdout, "%s%s %s\n", pad, report.Glyph("ok", a.styled), shortValue(answer))
		} else {
			fmt.Fprintf(a.stdout, "%s%s no answer\n", pad, report.Glyph("error", a.styled))
		}
		if r.Proposal != nil && *r.Proposal != answer {
			fmt.Fprintf(a.stdout, "%s%s proposal (%s): %s\n", pad, report.Glyph("info", a.styled), *r.Source, shortValue(*r.Proposal))
		}
	}
	fmt.Fprintln(a.stdout)
	text := count(len(rows), "question")
	if answered > 0 {
		text += fmt.Sprintf(" · %d answered", answered)
	}
	if answered < len(rows) {
		text += fmt.Sprintf(" · %d without an answer", len(rows)-answered)
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
