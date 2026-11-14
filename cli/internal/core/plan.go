package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/questions"
	"github.com/virajp/bootstrap/cli/internal/scripts"
	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// The first line of a new values file.
const (
	RepoHeader = "# The values of bootstrap for this repository. See docs/sync/index.md of bootstrap."
	MacHeader  = "# The values of the Mac sync of bootstrap. See docs/setup/index.md."
)

// ItemKind is the kind of an item of the change list.
type ItemKind int

// The kinds of an item.
const (
	// AddTool selects the tools of Names and their new dependencies.
	AddTool ItemKind = iota + 1
	// AddPack selects the members of the packs of Names, without the members
	// of Exclude, and their new dependencies.
	AddPack
	// RemoveTool removes the tools of Names, and each dependency that no
	// other tool needs.
	RemoveTool
	// Replace runs the migrate script of New for Old and swaps the tools.
	Replace
	// AddPackage writes Package to the own [mise] tables of macos.toml (Mac
	// only).
	AddPackage
)

// Item is one item of the change list.
type Item struct {
	Kind     ItemKind
	Names    []string // AddTool, RemoveTool: the tools; AddPack: the packs.
	Exclude  []string // AddPack: the member keys that the pack does not add.
	Old, New string   // Replace.
	Package  Package  // AddPackage.
}

// Answer is the answer of a question <tool>.<key>; "" is an empty answer.
// Flag is the text of the --answer flag that gives it: Plan reads the name
// and the value from it, and examines it. An answer with no Flag (the answer
// of a field) is not examined.
type Answer struct {
	Name, Value string
	Flag        string
}

// FlagAnswer gives the answer of the flag --answer <text>.
func FlagAnswer(text string) Answer {
	name, value, _ := strings.Cut(text, "=")
	return Answer{Name: name, Value: value, Flag: text}
}

// List is the change list of a command.
type List struct {
	// Init creates the values file, or migrates its old form, and asks the
	// questions of each selected tool. A new values file gets the key format
	// and a table for each member of the pack base (add pack base).
	Init    bool
	Items   []Item
	Answers []Answer
}

// Skip is a member of a pack that add pack leaves out, and why: "hosts"
// (its hosts leave out this host), "selectable" (selectable = false) or
// "root_only" (the repository is a submodule).
type Skip struct{ Tool, Why string }

// Script is one script of a tool that Apply runs before the writes.
type Script struct {
	Runner scripts.Runner
	Job    scripts.Job
	Note   string
}

// Shown gives the path of the script for a message.
func (s *Script) Shown() string { return s.Runner.ShownPath(s.Job.Tool, s.Job.Name) }

// PlanResult is the result of Plan.
type PlanResult struct {
	Side
	Cur     []byte // The values file before the command.
	Existed bool
	Next    []byte // The values file after the command, with the answers.
	Notes   []string
	// Changes are the writes before the sync: the values file first, then
	// the changes of a migration.
	Changes []oldform.Change
	// Sel is the selection of the new values file, before the answers.
	Sel *Selected
	// AskOnly holds the tools whose questions the command asks; nil for
	// each tool.
	AskOnly map[string]bool
	// Pending are the questions with no answer. Plan runs no dry run while
	// a question is pending: the caller asks them, and plans again with the
	// answers.
	Pending []questions.Question
	// Files are the files that the sync writes, changes or removes, sorted.
	Files    []target.Planned
	Before   *Script  // A script that runs after the consent, before the writes.
	Seeds    []string // The seeds of the removed tools that stay.
	Reasons  []string // Why the command has nothing to add.
	Warnings []string // The lines that the caller shows first.
	Skipped  []Skip   // The members of a pack that add pack leaves out.
}

// Empty tells if the plan changes nothing.
func (p *PlanResult) Empty() bool {
	return len(p.Changes) == 0 && len(p.Files) == 0 && p.Before == nil
}

// Paths gives each path that the plan writes.
func (p *PlanResult) Paths() []string {
	return append(oldform.Paths(p.Changes), PlanPaths(p.Files)...)
}

// Dirty is the check of the clean-target rule for the plan: each path that
// it writes and that has changes that are not committed.
func (p *PlanResult) Dirty() []string { return DirtyTargets(p.Paths(), nil) }

func join(errs []error) error { return errors.Join(errs...) }

// q is a string as a yq string literal.
func q(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// EditTOML runs the yq expressions <exprs> on the TOML text <in>, so its
// comments stay.
func EditTOML(in []byte, exprs []string) ([]byte, error) {
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

func tmpDir() string {
	if d := os.Getenv("TMPDIR"); d != "" {
		return d
	}
	return "/tmp"
}

// Plan plans the change list <l> on the side <sd>: it applies the items in
// order to a temporary copy of the values file, examines the answers, finds
// the questions with no answer, and runs the dry run of the sync with the new
// values file. It writes nothing. It gives each problem as an error; the
// result also comes with an error, for its Warnings.
func Plan(sd Side, l List, o Options) (*PlanResult, error) {
	p := &PlanResult{Side: sd}
	var st *State
	if l.Init {
		if err := p.initValues(); err != nil {
			return p, err
		}
	} else {
		var err error
		if st, err = Open(sd); err != nil {
			return p, err
		}
		p.Cur, p.Existed, p.Next = st.Cur, true, st.Cur
		p.AskOnly = map[string]bool{}
	}

	// A temporary copy of the values file, with its overrides file, for the
	// selection and the plan.
	tmp, err := os.MkdirTemp(tmpDir(), "bootstrap-init.")
	if err != nil {
		return p, problem(fail.Other, "cannot create a temp folder", "")
	}
	defer os.RemoveAll(tmp)
	tmpValues := tmp + "/" + filepath.Base(sd.Values)
	write := func() error {
		if err := os.WriteFile(tmpValues, p.Next, 0o644); err != nil {
			return problem(fail.Other, "cannot write a temp file", "")
		}
		return nil
	}
	if err := write(); err != nil {
		return p, err
	}
	copyOverrides(sd.Values, tmpValues)

	items := l.Items
	newFile := l.Init && !p.Existed
	if len(items) > 0 || newFile {
		if st == nil {
			raw, err := source.ReadTOML(tmpValues)
			if err != nil {
				return p, problem(fail.Usage, "cannot read "+sd.Show(sd.Values), "")
			}
			if st, err = stateOf(sd, tmpValues, p.Next, raw); err != nil {
				return p, err
			}
		}
		// A new values file gets the members of base, as add pack base does
		// (D6, D16 c).
		if newFile && st.Packs.HasPack(selection.BasePack) {
			items = append([]Item{{Kind: AddPack, Names: []string{selection.BasePack}}}, items...)
		}
		pl := &planner{st: st, p: p, o: o, added: map[string]bool{}, removed: map[string]bool{}}
		for _, it := range items {
			if err := pl.item(it); err != nil {
				return p, err
			}
		}
		if len(pl.exprs) > 0 {
			// A [tools] table without a tool goes.
			next, err := EditTOML(p.Next, append(pl.exprs, "del(.tools | select(length == 0))"))
			if err != nil {
				return p, err
			}
			p.Next = next
			if err := write(); err != nil {
				return p, err
			}
		}
	}
	if l.Init {
		p.AskOnly = nil
	}

	// The questions and the answers.
	sel, err := SelectionOf(sd, tmpValues, sd.Values)
	if err != nil {
		return p, err
	}
	p.Sel = sel
	given, err := CheckAnswers(sel, l.Answers)
	if err != nil {
		return p, err
	}
	p.Pending = Pending(sel, given, p.AskOnly)
	if len(given) > 0 {
		var exprs []string
		for _, qq := range questions.Of(sel.Idx, sel.Tools) {
			if v, ok := given[qq.Name()]; ok {
				exprs = append(exprs, ".tools["+q(qq.Tool)+"]["+q(qq.Key)+"] = "+q(v))
				p.Notes = append(p.Notes, "answer "+qq.Name())
			}
		}
		if p.Next, err = EditTOML(p.Next, exprs); err != nil {
			return p, err
		}
		if err := write(); err != nil {
			return p, err
		}
	}
	if len(p.Pending) > 0 {
		return p, nil
	}
	// A command that writes the values file also adds the missing tables of
	// the required members of base, so the sync of Apply finds none and the
	// values file is written one time (D21 d). A command that does not write
	// it leaves them to the sync (D14 d).
	if !p.Existed || !bytes.Equal(p.Next, p.Cur) {
		if err := p.requiredTables(sel, write); err != nil {
			return p, err
		}
	}
	if !p.Existed || !bytes.Equal(p.Next, p.Cur) {
		p.Changes = append([]oldform.Change{{Path: sd.Values, Content: p.Next, Note: strings.Join(p.Notes, "; ")}}, p.Changes...)
	}

	// The plan of the sync, with the new values file.
	run := frontEnd(sd.Mac)
	envKey, envVal := "BOOTSTRAP_VALUES", tmpValues
	if sd.Mac {
		envKey, envVal = "MACOS_SETUP_REPO_PATH", tmp
	}
	old, had := os.LookupEnv(envKey)
	_ = os.Setenv(envKey, envVal)
	oldform.Migrating = l.Init
	args := SyncArgs(sd.Mac, true, o.NoBootstrap)
	planned, pc := run(args, io.Discard, io.Discard)
	if pc != fail.OK {
		// Again, with the output, so the messages show.
		out, errw := o.writers()
		_, pc = run(args, out, errw)
	}
	oldform.Migrating = false
	if had {
		_ = os.Setenv(envKey, old)
	} else {
		_ = os.Unsetenv(envKey)
	}
	if pc != fail.OK {
		return p, &Problem{Code: pc}
	}
	// The temporary copy is not a change; the sync can show it by its real
	// path.
	realTmp := tmp
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		realTmp = r
	}
	seen := map[string]bool{}
	for _, f := range planned {
		if strings.HasPrefix(f.Path, tmp+"/") || strings.HasPrefix(f.Path, realTmp+"/") || seen[f.Path] {
			continue
		}
		seen[f.Path] = true
		p.Files = append(p.Files, f)
	}
	sort.Slice(p.Files, func(i, j int) bool { return p.Files[i].Path < p.Files[j].Path })
	return p, nil
}

// requiredTables adds to p.Next an empty [tools.<name>] table for each
// required member of base that has no table in the selection <sel>, as the
// sync does, and writes the temporary copy again with <write>.
func (p *PlanResult) requiredTables(sel *Selected, write func() error) error {
	named := map[string]bool{}
	for _, it := range sel.Items {
		if it.Kind == "tools" {
			named[it.Name] = true
		}
	}
	names, err := selection.MissingTables(sel.Idx, selection.BasePack, named, !p.Mac && InSubmodule(p.Root))
	if err != nil || len(names) == 0 {
		return err
	}
	var exprs []string
	for _, n := range names {
		// An answer can have given the table already: keep its keys.
		exprs = append(exprs, ".tools["+q(n)+"] = (.tools["+q(n)+"] // {})")
		p.Notes = append(p.Notes, "add [tools."+n+"] (required by the base pack)")
	}
	if p.Next, err = EditTOML(p.Next, exprs); err != nil {
		return err
	}
	return write()
}

// copyOverrides copies the overrides file of <file> next to <to>.
func copyOverrides(file, to string) {
	if b, err := os.ReadFile(values.OverridesPath(file)); err == nil {
		_ = os.WriteFile(values.OverridesPath(to), b, 0o644)
	}
}

// initValues gives the values file of init: a new file with the key
// format, or the file migrated from its old form.
func (p *PlanResult) initValues() error {
	sd := p.Side
	cur, err := os.ReadFile(sd.Values)
	p.Cur, p.Existed = cur, err == nil
	p.Next = cur
	if !p.Existed {
		format := fmt.Sprintf("%s = %d\n", oldform.FormatKey, oldform.Format)
		if sd.Mac {
			p.Next = []byte(MacHeader + "\n\n" + format)
		} else {
			// The templates need repo_name: the name of the folder.
			p.Next = []byte(RepoHeader + "\n\n" + format + "environments = [ \"dev\" ]\nrepo_name = " + q(filepath.Base(sd.Root)) + "\n")
		}
		p.Notes = append(p.Notes, "create the values file")
		return nil
	}
	var migs []oldform.Change
	if sd.Mac {
		ch, err := oldform.MigrateMac(sd.Values)
		if err != nil {
			return err
		}
		if ch != nil {
			migs = append(migs, *ch)
		}
	} else {
		if migs, err = oldform.MigrateRepo(sd.Root, sd.Values); err != nil {
			return err
		}
	}
	for _, ch := range migs {
		if ch.Path == sd.Values {
			p.Next = ch.Content
			p.Notes = append(p.Notes, ch.Note)
		} else {
			p.Changes = append(p.Changes, ch)
		}
	}
	return nil
}

// InitChange gives the change that init makes before each other change of
// the list: the new values file with the members of the pack base, or the
// migration of an old form. A values file of the new form gives "". It writes
// nothing. The TUI shows the change on its change list.
func InitChange(sd Side) (string, error) {
	if !Exists(sd.Values) {
		return "create the values file with " + selection.BasePack, nil
	}
	p := &PlanResult{Side: sd}
	if err := p.initValues(); err != nil {
		return "", err
	}
	if len(p.Changes) > 0 || !bytes.Equal(p.Next, p.Cur) {
		return fmt.Sprintf("migrate the values file to %s %d", oldform.FormatKey, oldform.Format), nil
	}
	return "", nil
}

// planner applies the items of a change list to the state of the values
// file. added and removed hold the changes of the items before, so an item
// sees them.
type planner struct {
	st             *State
	p              *PlanResult
	o              Options
	exprs          []string
	added, removed map[string]bool
	// direct holds the tools that an item names and that it selects, when
	// it is not nil (Effect).
	direct map[string]bool
}

func (pl *planner) selected(t string) bool {
	return pl.added[t] || (!pl.removed[t] && pl.st.Selected(t))
}

func (pl *planner) named(t string) bool {
	return pl.added[t] || (!pl.removed[t] && pl.st.Named[t])
}

func (pl *planner) note(expr, note string) {
	pl.exprs = append(pl.exprs, expr)
	pl.p.Notes = append(pl.p.Notes, note)
}

// item applies one item of the change list.
func (pl *planner) item(it Item) error {
	st := pl.st
	switch it.Kind {
	case AddTool:
		if err := st.Known(it.Names); err != nil {
			return err
		}
		if err := Refuse(st.Sel.Idx, it.Names); err != nil {
			return err
		}
		pl.add(it.Names, true)
	case AddPack:
		return pl.addPack(it.Names, it.Exclude)
	case RemoveTool:
		if err := st.Known(it.Names); err != nil {
			return err
		}
		return pl.remove(it.Names, nil)
	case Replace:
		return pl.replace(it.Old, it.New)
	case AddPackage:
		if !st.Mac {
			return problem(fail.Usage, "a package of macos.toml is only for the Mac side", "")
		}
		pl.note(it.Package.expr(), "add "+it.Package.Key+" to "+it.Package.Table())
	default:
		return problem(fail.Other, fmt.Sprintf("an unknown item kind %d", it.Kind), "")
	}
	return nil
}

// add adds <names> and each tool that they depend on, that the selection
// does not have yet. A dependency with selectable = false gets no table: the
// sync installs it through depends. With <reason>, a name that is selected
// gives the reason "<name> is already selected".
func (pl *planner) add(names []string, reason bool) {
	st := pl.st
	named := map[string]bool{}
	for _, n := range names {
		named[n] = true
		if pl.direct != nil {
			pl.direct[n] = true
		}
		if reason && pl.selected(n) {
			pl.p.Reasons = append(pl.p.Reasons, n+" is already selected")
		}
	}
	// Each name and its dependencies, a dependency before its tool.
	w, _ := selection.Walk(names, func(t string) ([]string, error) {
		return append([]string(nil), st.Manifest(t).Depends...), nil
	}, func(string) error { return nil })
	for _, t := range w.Visited {
		if pl.selected(t) {
			continue
		}
		if m := st.Manifest(t); !named[t] && m != nil && !m.Selectable {
			continue
		}
		pl.added[t] = true
		if pl.p.AskOnly != nil {
			pl.p.AskOnly[t] = true
		}
		pl.note(".tools["+q(t)+"] = {}", "select "+t)
	}
}

// expansion is the expansion of a pack: its tools, its excluded tools, and
// the pack of each member key.
type expansion struct {
	tools, excluded []string
	keys            map[string]string
}

// walk expands the pack <name>: each member in file order, a "pack:<x>"
// member expanded in turn. A member of <ex>, and each tool below it, is
// excluded.
func (e *expansion) walk(packs *source.Packs, name string, ex map[string]bool, dropped bool, seen map[string]bool) {
	if seen[name] {
		return
	}
	seen[name] = true
	defer delete(seen, name)
	for _, o := range packs.Order {
		if o.Pack != name || !packs.HasMember(name, o.Key) {
			continue
		}
		if _, ok := e.keys[o.Key]; !ok {
			e.keys[o.Key] = name
		}
		d := dropped || ex[o.Key]
		if rest, ok := strings.CutPrefix(o.Key, "pack:"); ok {
			e.walk(packs, rest, ex, d, seen)
			continue
		}
		if d {
			e.excluded = append(e.excluded, o.Key)
		} else {
			e.tools = append(e.tools, o.Key)
		}
	}
}

// addPack adds the members of the packs <names> that are not selected yet,
// without the members of <exclude>, and their new dependencies. It leaves
// out a member whose hosts leave out this host, a member with selectable =
// false, and in a submodule a member with root_only.
func (pl *planner) addPack(names, exclude []string) error {
	st := pl.st
	if len(names) > 1 && len(exclude) > 0 {
		return problem(fail.Usage, "--exclude needs exactly one pack name", "run "+CmdName(st.Mac, "add pack")+" once for each pack")
	}
	var errs []error
	for _, n := range names {
		if !st.Packs.HasPack(n) {
			errs = append(errs, problem(fail.Missing, n+" is not a bootstrap pack", ""))
		}
	}
	if err := join(errs); err != nil {
		return err
	}
	ex := map[string]bool{}
	for _, m := range exclude {
		ex[m] = true
	}
	host := source.HostName()
	sub := !st.Mac && InSubmodule(st.Root)
	for _, name := range names {
		e := &expansion{keys: map[string]string{}}
		e.walk(st.Packs, name, ex, false, map[string]bool{})
		for _, m := range exclude {
			pack, ok := e.keys[m]
			switch {
			case !ok:
				errs = append(errs, problem(fail.Usage, m+" is not a member of the pack "+name, ""))
			case st.Packs.Required[pack+"/"+m]:
				errs = append(errs, problem(fail.Usage, m+" is required in the pack "+pack+", so "+CmdWord(st.Mac, "add pack")+" cannot exclude it", ""))
			}
		}
		if err := join(errs); err != nil {
			return err
		}
		out := map[string]bool{}
		for _, t := range e.excluded {
			out[t] = true
		}
		var news []string
		seen := map[string]bool{}
		for _, t := range e.tools {
			if out[t] || seen[t] {
				continue
			}
			seen[t] = true
			m := st.Manifest(t)
			switch {
			case m == nil:
			case !m.RunsOn(host):
				pl.p.Skipped = append(pl.p.Skipped, Skip{t, "hosts"})
				continue
			case !m.Selectable:
				pl.p.Skipped = append(pl.p.Skipped, Skip{t, "selectable"})
				continue
			case sub && m.RootOnly:
				pl.p.Skipped = append(pl.p.Skipped, Skip{t, "root_only"})
				continue
			}
			if !pl.selected(t) {
				news = append(news, t)
			}
		}
		if len(news) == 0 {
			pl.p.Reasons = append(pl.p.Reasons, name+" is already selected")
			continue
		}
		// An excluded tool that a tool of the pack needs.
		for _, x := range sortedKeys(out) {
			if pl.selected(x) {
				continue
			}
			var by []string
			for _, t := range news {
				if t != x && st.Closure([]string{t})[x] {
					by = append(by, t)
				}
			}
			if len(by) > 0 {
				sort.Strings(by)
				list := strings.Join(by, ", ")
				errs = append(errs, problem(fail.Usage, x+" is needed by "+list, "exclude "+list+" too, or keep "+x))
			}
		}
		if err := join(errs); err != nil {
			return err
		}
		pl.add(news, false)
	}
	return nil
}

// sortedKeys gives the keys of <m> with a true value, sorted.
func sortedKeys(m map[string]bool) []string {
	var out []string
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// remove removes <names>, and each dependency of them that is a
// [tools.<name>] table and that no other selected tool needs. <extra> are
// tools that stay selected too (replace: the new tool). It refuses a
// required tool, a tool that only a pack selects, a tool that is not
// selected, and a tool that another selected tool needs (code 2).
func (pl *planner) remove(names, extra []string) error {
	st := pl.st
	verb := CmdWord(st.Mac, "remove")
	var errs []error
	removed := map[string]bool{}
	for _, n := range names {
		removed[n] = true
	}
	for _, n := range names {
		switch {
		case st.Required(n):
			pack := st.PackOf(n)
			if pack == "" {
				pack = selection.BasePack
			}
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("%s is required by the %s pack, so %s cannot remove it", n, pack, verb), ""))
		case !pl.named(n) && st.PackOf(n) != "":
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("%s comes from the %s pack, so %s cannot remove it", n, st.PackOf(n), verb), ""))
		case !pl.named(n) && !pl.selected(n):
			errs = append(errs, problem(fail.Usage, n+" is not selected", ""))
		}
	}
	if err := join(errs); err != nil {
		return err
	}
	// The pack tools stay; each other tool stays when it is a [tools.*]
	// table that is not removed.
	var packTools []string
	for _, e := range st.Res.Sel {
		if !strings.HasPrefix(e.Src, "tools.") {
			packTools = append(packTools, e.Tool)
		}
	}
	tables := map[string]bool{}
	for t := range st.Named {
		tables[t] = pl.named(t)
	}
	for t := range pl.added {
		tables[t] = true
	}
	var deps []string
	for _, n := range names {
		deps = append(deps, st.Manifest(n).Depends...)
	}
	cands := map[string]bool{}
	for t := range st.Closure(deps) {
		if tables[t] && !removed[t] {
			cands[t] = true
		}
	}
	kept := map[string]bool{}
	var need map[string]bool
	for {
		start := append(append([]string(nil), packTools...), extra...)
		for t, v := range tables {
			if v && !removed[t] && (!cands[t] || kept[t]) {
				start = append(start, t)
			}
		}
		need = st.Closure(start)
		changed := false
		for t := range cands {
			if need[t] && !kept[t] {
				kept[t] = true
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	for _, n := range names {
		if !need[n] {
			continue
		}
		var by []string
		for t := range need {
			if removed[t] {
				continue
			}
			for _, d := range st.Manifest(t).Depends {
				if d == n {
					by = append(by, t)
				}
			}
		}
		sort.Strings(by)
		errs = append(errs, problem(fail.Usage, n+" is needed by "+strings.Join(by, ", "), "remove "+strings.Join(by, ", ")+" first"))
	}
	if err := join(errs); err != nil {
		return err
	}
	for _, n := range names {
		if pl.named(n) {
			pl.note("del(.tools["+q(n)+"])", "remove "+n)
		}
	}
	gone := append([]string(nil), names...)
	for _, t := range sortedKeys(cands) {
		if !kept[t] {
			pl.note("del(.tools["+q(t)+"])", "remove "+t+" (a dependency that no tool needs)")
			gone = append(gone, t)
		}
	}
	for _, t := range gone {
		pl.removed[t] = true
		delete(pl.added, t)
		delete(pl.direct, t)
	}
	// The seeds of the removed tools stay; the command lists them.
	root := st.Home()
	for _, t := range gone {
		if need[t] {
			continue
		}
		for _, fe := range st.Manifest(t).Files {
			if p := root + "/" + fe.Target; fe.Mode == "seed" && Exists(p) {
				pl.p.Seeds = append(pl.p.Seeds, p)
			}
		}
	}
	return nil
}

// replace runs the migrate script of <newT> for <oldT> after the consent,
// and swaps the tools. The new tool gets the env of the old tool.
func (pl *planner) replace(oldT, newT string) error {
	st := pl.st
	verb := CmdWord(st.Mac, "replace")
	if err := st.Known([]string{oldT, newT}); err != nil {
		return err
	}
	if err := Refuse(st.Sel.Idx, []string{newT}); err != nil {
		return err
	}
	mo, mn := st.Manifest(oldT), st.Manifest(newT)
	if mo.Parent == "" || mo.Parent != mn.Parent {
		return problem(fail.Usage, fmt.Sprintf("%s works only between two tools with the same parent (%s has the parent '%s', %s has '%s')", verb, oldT, mo.Parent, newT, mn.Parent), "")
	}
	if pl.selected(newT) {
		return problem(fail.Usage, newT+" is already selected", "")
	}
	r := st.Runner()
	name := source.MigratePrefix + oldT
	if !r.Has(newT, name) {
		return problem(fail.Usage, fmt.Sprintf("%s has no script %s, so %s cannot replace %s with it", newT, r.ShownPath(newT, name), verb, oldT), "")
	}
	// The migrate script can change any file: the work tree must be clean.
	if !st.Mac && DirtyBesides(st.Root, nil) {
		if !pl.o.DryRun {
			return problem(fail.Usage, "the work tree has changes that are not committed", "commit or discard them, then run the command again")
		}
		pl.p.Warnings = append(pl.p.Warnings, "the work tree has changes that are not committed; a run without --dry-run stops")
	}
	if err := pl.remove([]string{oldT}, []string{newT}); err != nil {
		return err
	}
	pl.add([]string{newT}, true)
	if t, ok := st.Raw.Get("tools").(*source.Obj); ok {
		if e, ok := t.Get(oldT).(*source.Obj); ok && e.Has("env") {
			pl.exprs = append(pl.exprs, ".tools["+q(newT)+"].env = "+source.JSON(e.Get("env")))
		}
	}
	pl.p.Before = &Script{Runner: r, Job: scripts.Job{Tool: newT, Name: name, Input: st.Input(newT, oldT)}, Note: "move the files of " + oldT + " to " + newT}
	return nil
}
