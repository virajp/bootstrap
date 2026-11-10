package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/oldform"
	"github.com/virajp/bootstrap/sync-engine/internal/scripts"
	"github.com/virajp/bootstrap/sync-engine/internal/selection"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// The selection commands: add, remove, replace and detect. Each one changes
// the [tools.<name>] tables of the values file, then syncs with the steps of
// init (finish).

// side is the state that each selection command reads first.
type side struct {
	mac   bool
	w     where
	c     *clone
	cur   []byte
	sel   *selected
	res   *selection.Resolved
	named map[string]bool // The [tools.<name>] tables of the values file.
	packs *source.Packs
	raw   *source.Obj // The values file.
}

// cmdName gives the command as the user types it.
func cmdName(mac bool, verb string) string {
	if mac {
		return "bootstrap mac " + verb
	}
	return "bootstrap " + verb
}

// open does the preflight, takes the clone, and reads the values file and
// its selection. The values file must exist and have the new form. The
// caller defers s.c.cleanup when s.c is set.
func (a *app) open(mac bool) (*side, fail.Code) {
	w, code := a.prepare(mac, false)
	if code != fail.OK {
		return nil, code
	}
	c, code := a.openClone(mac)
	s := &side{mac: mac, w: w, c: c}
	if code != fail.OK {
		return s, code
	}
	cur, err := os.ReadFile(w.values)
	if err != nil {
		a.errorf("%s is not there - run %s first", a.show(w.values), cmdName(mac, "init"))
		return s, fail.Usage
	}
	s.cur = cur
	raw, err := source.ReadTOML(w.values)
	if err != nil {
		a.errorf("cannot read %s", a.show(w.values))
		return s, fail.Usage
	}
	s.raw = raw
	if mac {
		err = oldform.Mac(raw, a.show(w.values))
	} else {
		err = oldform.Repo(w.root, raw, a.show(w.values))
	}
	if err != nil {
		return s, a.logErr(err)
	}
	sel, code := a.selectionOf(mac, c.dir, w.root, w.values, w.values)
	if code != fail.OK {
		return s, code
	}
	s.sel, s.res = sel, sel.res
	s.named = map[string]bool{}
	for _, it := range sel.items {
		if it.Kind == "tools" {
			s.named[it.Name] = true
		}
	}
	if s.packs, err = sel.idx.Source.ReadPacks(); err != nil {
		return s, a.logErr(err)
	}
	return s, fail.OK
}

// change gives a change of the values file of <s>, which starts from its
// current text.
func (s *side) change() *change {
	return &change{w: s.w, c: s.c, cur: s.cur, existed: true, next: s.cur}
}

// manifest gives the manifest of <tool>, or nil.
func (s *side) manifest(tool string) *source.Manifest { return s.sel.idx.Manifests[tool] }

// selected tells if <tool> is in the selection.
func (s *side) selected(tool string) bool {
	for _, t := range s.res.Tools {
		if t == tool {
			return true
		}
	}
	return false
}

// packOf gives the pack that selects <tool>, or empty.
func (s *side) packOf(tool string) string {
	for _, e := range s.res.Sel {
		if e.Tool != tool || strings.HasPrefix(e.Src, "tools.") {
			continue
		}
		if e.Src == "default" {
			return selection.BasePack
		}
		return strings.TrimPrefix(e.Src, "packs.")
	}
	return ""
}

// required tells if <tool> has required = true in a pack.
func (s *side) required(tool string) bool {
	for k, v := range s.packs.Required {
		if v && strings.HasSuffix(k, "/"+tool) {
			return true
		}
	}
	return false
}

// closure gives <start> and each tool that they depend on, transitively.
func (s *side) closure(start []string) map[string]bool {
	out := map[string]bool{}
	queue := append([]string(nil), start...)
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		if out[t] {
			continue
		}
		out[t] = true
		if m := s.manifest(t); m != nil {
			queue = append(queue, m.Depends...)
		}
	}
	return out
}

// known gives code 3, after a message, for each name of <names> that is not
// a tool of the side.
func (a *app) known(s *side, names []string) fail.Code {
	code := fail.OK
	for _, n := range names {
		if !source.NameOK(n) || s.manifest(n) == nil {
			a.errorf("tool '%s' is not in bootstrap", n)
			code = fail.Missing
		}
	}
	return code
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

// planAdd adds <names> and each tool that they depend on, that the
// selection does not have yet, to <ch>. It gives the added tools.
func (a *app) planAdd(s *side, ch *change, names []string) map[string]bool {
	added := map[string]bool{}
	for _, n := range names {
		if s.selected(n) {
			a.infof("tool %s is already selected", n)
		}
	}
	// Each name and its dependencies, a dependency before its tool.
	w, _ := selection.Walk(names, func(t string) ([]string, error) {
		return append([]string(nil), s.manifest(t).Depends...), nil
	}, func(string) error { return nil })
	for _, t := range w.Visited {
		if s.selected(t) || added[t] {
			continue
		}
		added[t] = true
		ch.exprs = append(ch.exprs, ".tools["+q(t)+"] = {}")
		ch.notes = append(ch.notes, "select "+t)
	}
	return added
}

// planRemove removes <names> from <ch>, and each dependency of them that is
// a [tools.<name>] table and that no other selected tool needs. <extra> are
// tools that stay selected too (replace: the new tool). It refuses a
// required tool, a tool that only a pack selects, a tool that is not
// selected, and a tool that another selected tool needs (code 2).
func (a *app) planRemove(s *side, ch *change, names, extra []string) fail.Code {
	code := fail.OK
	removed := map[string]bool{}
	for _, n := range names {
		removed[n] = true
	}
	for _, n := range names {
		switch {
		case s.required(n):
			a.errorf("tool %s is required - %s cannot remove it", n, cmdName(s.mac, "remove"))
			code = fail.Usage
		case !s.named[n] && s.packOf(n) != "":
			a.errorf("tool %s comes from the pack %s - %s removes only a tool of [tools.*]", n, s.packOf(n), cmdName(s.mac, "remove"))
			code = fail.Usage
		case !s.named[n] && !s.selected(n):
			a.errorf("tool %s is not selected", n)
			code = fail.Usage
		}
	}
	if code != fail.OK {
		return code
	}
	// The pack tools stay; each other tool stays when it is a [tools.*]
	// table that is not removed.
	var packTools []string
	for _, e := range s.res.Sel {
		if !strings.HasPrefix(e.Src, "tools.") {
			packTools = append(packTools, e.Tool)
		}
	}
	var deps []string
	for _, n := range names {
		deps = append(deps, s.manifest(n).Depends...)
	}
	cands := map[string]bool{}
	for t := range s.closure(deps) {
		if s.named[t] && !removed[t] {
			cands[t] = true
		}
	}
	kept := map[string]bool{}
	var need map[string]bool
	for {
		start := append(append([]string(nil), packTools...), extra...)
		for t := range s.named {
			if !removed[t] && (!cands[t] || kept[t]) {
				start = append(start, t)
			}
		}
		need = s.closure(start)
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
			for _, d := range s.manifest(t).Depends {
				if d == n {
					by = append(by, t)
				}
			}
		}
		sort.Strings(by)
		a.errorf("tool %s is needed by %s - remove that tool first", n, strings.Join(by, ", "))
		code = fail.Usage
	}
	if code != fail.OK {
		return code
	}
	for _, n := range names {
		if s.named[n] {
			ch.exprs = append(ch.exprs, "del(.tools["+q(n)+"])")
			ch.notes = append(ch.notes, "remove "+n)
		}
	}
	gone := append([]string(nil), names...)
	for _, t := range sortedKeys(cands) {
		if !kept[t] {
			ch.exprs = append(ch.exprs, "del(.tools["+q(t)+"])")
			ch.notes = append(ch.notes, "remove "+t+" (a dependency that no tool needs)")
			gone = append(gone, t)
		}
	}
	// The seeds of the removed tools stay; the command lists them.
	root := s.w.root
	if s.mac {
		root = os.Getenv("HOME")
	}
	for _, t := range gone {
		if need[t] {
			continue
		}
		for _, fe := range s.manifest(t).Files {
			if p := root + "/" + fe.Target; fe.Mode == "seed" && exists(p) {
				ch.seeds = append(ch.seeds, p)
			}
		}
	}
	return fail.OK
}

// edited applies the expressions of <ch> to its values file.
func (a *app) edited(ch *change) fail.Code {
	if len(ch.exprs) == 0 {
		return fail.OK
	}
	// A [tools] table without a tool goes.
	next, err := editTOML(ch.next, append(ch.exprs, "del(.tools | select(length == 0))"))
	if err != nil {
		return a.logErr(err)
	}
	ch.next, ch.exprs = next, nil
	return fail.OK
}

// add is `bootstrap [mac] add <tool...>`.
func (a *app) add(mac bool, o opts) fail.Code {
	s, code := a.open(mac)
	if s != nil && s.c != nil {
		defer s.c.cleanup()
	}
	if code != fail.OK {
		return code
	}
	return a.addNames(s, o, o.names)
}

// addNames adds <names> to the selection of <s>, asks their questions, and
// syncs.
func (a *app) addNames(s *side, o opts, names []string) fail.Code {
	if code := a.known(s, names); code != fail.OK {
		return code
	}
	ch := s.change()
	ch.askOnly = a.planAdd(s, ch, names)
	if code := a.edited(ch); code != fail.OK {
		return code
	}
	return a.finish(s.mac, o, ch)
}

// remove is `bootstrap [mac] remove <tool...>`.
func (a *app) remove(mac bool, o opts) fail.Code {
	s, code := a.open(mac)
	if s != nil && s.c != nil {
		defer s.c.cleanup()
	}
	if code != fail.OK {
		return code
	}
	if code := a.known(s, o.names); code != fail.OK {
		return code
	}
	ch := s.change()
	ch.askOnly = map[string]bool{}
	if code := a.planRemove(s, ch, o.names, nil); code != fail.OK {
		return code
	}
	if code := a.edited(ch); code != fail.OK {
		return code
	}
	return a.finish(mac, o, ch)
}

// replace is `bootstrap [mac] replace <old> <new>`.
func (a *app) replace(mac bool, o opts) fail.Code {
	if len(o.names) != 2 {
		a.errorf("%s needs two tools: <old> <new>", cmdName(mac, "replace"))
		return fail.Usage
	}
	s, code := a.open(mac)
	if s != nil && s.c != nil {
		defer s.c.cleanup()
	}
	if code != fail.OK {
		return code
	}
	oldT, newT := o.names[0], o.names[1]
	if code := a.known(s, o.names); code != fail.OK {
		return code
	}
	mo, mn := s.manifest(oldT), s.manifest(newT)
	if mo.Parent == "" || mo.Parent != mn.Parent {
		a.errorf("%s works only between two tools with the same parent - %s has the parent '%s', %s has '%s'", cmdName(mac, "replace"), oldT, mo.Parent, newT, mn.Parent)
		return fail.Usage
	}
	if s.selected(newT) {
		a.errorf("tool %s is already selected", newT)
		return fail.Usage
	}
	r := s.runner(a)
	name := source.MigratePrefix + oldT
	if !r.Has(newT, name) {
		a.errorf("tool %s has no script %s - %s cannot replace %s with it", newT, r.ShownPath(newT, name), cmdName(mac, "replace"), oldT)
		return fail.Usage
	}
	// The migrate script can change any file: the work tree must be clean.
	if !mac {
		if o.dryRun {
			if a.dirtyBesides(s.w.root, nil) {
				a.infof("the work tree has changes that are not committed - a run without --dry-run stops")
			}
		} else if a.dirtyBesides(s.w.root, nil) {
			a.errorf("the work tree has changes that are not committed - commit or discard them, then run the command again - nothing written")
			return fail.Usage
		}
	}
	ch := s.change()
	if code := a.planRemove(s, ch, []string{oldT}, []string{newT}); code != fail.OK {
		return code
	}
	ch.askOnly = a.planAdd(s, ch, []string{newT})
	// The new tool gets the env of the old tool.
	if t, ok := s.raw.Get("tools").(*source.Obj); ok {
		if e, ok := t.Get(oldT).(*source.Obj); ok && e.Has("env") {
			ch.exprs = append(ch.exprs, ".tools["+q(newT)+"].env = "+source.JSON(e.Get("env")))
		}
	}
	if code := a.edited(ch); code != fail.OK {
		return code
	}
	in := s.input(newT, oldT)
	ch.before = &script{runner: r, job: scripts.Job{Tool: newT, Name: name, Input: in}, note: "move the files of " + oldT + " to " + newT}
	return a.finish(mac, o, ch)
}

// runner gives the script runner of the side: the scripts run in the
// repository, or in $HOME on the Mac.
func (s *side) runner(a *app) scripts.Runner {
	src := s.sel.idx.Source
	root := s.w.root
	if s.mac {
		root = a.getenv("HOME")
	}
	return scripts.Runner{ToolsDir: src.ToolsDir(), Shown: src.ToolsShown(), Root: root}
}

// input gives the JSON input of a script of <tool>: the [tools.<from>]
// table of the values file as its values.
func (s *side) input(tool, from string) []byte {
	vals := source.NewObj()
	if v, ok := s.sel.user.Get(from).(*source.Obj); ok {
		vals = v
	}
	in := scripts.Input{Tool: tool, Values: vals}
	if s.mac {
		in.Home = os.Getenv("HOME")
	} else {
		in.Repo = s.repoInput()
	}
	return in.JSON()
}

// repoInput gives the repo object of a script input: the root, the
// repo_name, and the [member.<name>] tables.
func (s *side) repoInput() *scripts.Repo {
	r := &scripts.Repo{Root: s.w.root, Name: filepath.Base(s.w.root), Members: []scripts.Member{}}
	if n, ok := s.raw.Get("repo_name").(string); ok && n != "" {
		r.Name = n
	}
	if m, ok := s.raw.Get("member").(*source.Obj); ok {
		for _, name := range m.Keys() {
			e, _ := m.Get(name).(*source.Obj)
			mem := scripts.Member{Name: name, Folder: name}
			if e != nil {
				if f, ok := e.Get("folder").(string); ok {
					mem.Folder = f
				}
				mem.Type, _ = e.Get("type").(string)
			}
			r.Members = append(r.Members, mem)
		}
	}
	return r
}

// runScript runs the script of a command. A failure gives code 2.
func (a *app) runScript(sc *script) fail.Code {
	res := sc.runner.Run([]scripts.Job{sc.job})[0]
	if res.Code != 0 {
		for _, l := range strings.Split(strings.TrimSpace(string(res.Stderr)), "\n") {
			if l != "" {
				a.errorf("%s: %s", sc.runner.ShownPath(res.Tool, res.Name), l)
			}
		}
		a.errorf("%s failed with exit code %d", sc.runner.ShownPath(res.Tool, res.Name), res.Code)
		return fail.Usage
	}
	a.stderr.Write(res.Stderr)
	a.infof("Ran %s", sc.runner.ShownPath(res.Tool, res.Name))
	return fail.OK
}

// detected is the output of a detect script.
type detected struct {
	Result   string   `json:"result"`
	Evidence []string `json:"evidence"`
	Message  string   `json:"message"`
}

// detect is `bootstrap [mac] detect`.
func (a *app) detect(mac bool, o opts) fail.Code {
	if o.yes && o.no {
		a.errorf("--yes and --no cannot be used together")
		return fail.Usage
	}
	s, code := a.open(mac)
	if s != nil && s.c != nil {
		defer s.c.cleanup()
	}
	if code != fail.OK {
		return code
	}
	r := s.runner(a)
	var tools []string
	for t := range s.sel.idx.Manifests {
		if r.Has(t, scripts.Detect) {
			tools = append(tools, t)
		}
	}
	sort.Strings(tools)
	if len(tools) == 0 {
		if mac {
			a.infof("No Mac tool has a detect script")
		} else {
			a.infof("No tool has a detect script")
		}
		return fail.OK
	}
	var jobs []scripts.Job
	for _, t := range tools {
		jobs = append(jobs, scripts.Job{Tool: t, Name: scripts.Detect, Input: s.input(t, t)})
	}
	results := r.Run(jobs)
	found := map[string]detected{}
	code = fail.OK
	for _, res := range results {
		shown := r.ShownPath(res.Tool, res.Name)
		var d detected
		if res.Code != 0 {
			for _, l := range strings.Split(strings.TrimSpace(string(res.Stderr)), "\n") {
				if l != "" {
					a.errorf("%s: %s", shown, l)
				}
			}
			a.errorf("%s failed with exit code %d", shown, res.Code)
			code = fail.Usage
			continue
		}
		if err := json.Unmarshal(res.Stdout, &d); err != nil {
			a.errorf("%s did not print a JSON object: %s", shown, err)
			code = fail.Usage
			continue
		}
		switch d.Result {
		case "match", "none", "mismatch":
		default:
			a.errorf("%s printed the result '%s' - the results are match, none and mismatch", shown, d.Result)
			code = fail.Usage
			continue
		}
		found[res.Tool] = d
	}
	if code != fail.OK {
		return code
	}

	// The evidence of the selected tools that match.
	selEvidence := map[string]string{}
	for _, t := range tools {
		if d := found[t]; d.Result == "match" && s.selected(t) {
			for _, e := range d.Evidence {
				selEvidence[e] = t
			}
		}
	}
	var cands []string
	fmt.Fprintln(a.stdout, "Detected:")
	for _, t := range tools {
		d := found[t]
		sel := s.selected(t)
		if d.Result == "none" && !sel {
			continue
		}
		state := "not selected"
		if sel {
			state = "selected"
		}
		line := fmt.Sprintf("  %-8s %-14s %s", d.Result, t, state)
		if len(d.Evidence) > 0 {
			line += " (" + strings.Join(d.Evidence, ", ") + ")"
		}
		if d.Message != "" {
			line += " - " + d.Message
		}
		fmt.Fprintln(a.stdout, line)
		switch {
		case d.Result == "mismatch" && sel:
			a.errorf("tool %s is selected, but its detect script gives mismatch: %s", t, d.Message)
			code = fail.Usage
		case d.Result == "none" && sel:
			a.infof("tool %s is selected, but its detect script finds nothing - it stays selected", t)
		case d.Result == "match" && !sel:
			// Two tools with the same evidence: the selected one is enough.
			same := ""
			for _, e := range d.Evidence {
				if u, ok := selEvidence[e]; ok {
					same = u
				}
			}
			if same == "" {
				cands = append(cands, t)
			}
		}
	}
	if code != fail.OK {
		return code
	}
	if len(cands) == 0 {
		a.infof("No tool to add")
		return fail.OK
	}
	var add []string
	switch {
	case o.no:
		for _, t := range cands {
			a.errorf("tool %s matches, but it is not selected - add it with %s %s", t, cmdName(mac, "add"), t)
		}
		return fail.Other
	case o.yes:
		add = cands
	case !a.interactive():
		a.errorf("the detected tools need --yes or --no")
		return fail.Usage
	default:
		for _, t := range cands {
			fmt.Fprintf(a.stdout, "Add %s? [y/N] ", t)
			line, _ := a.readLine()
			switch strings.ToLower(strings.TrimSpace(line)) {
			case "y", "yes":
				add = append(add, t)
			}
		}
		if len(add) == 0 {
			a.infof("Nothing changed")
			return fail.OK
		}
		o.confirmed = true
	}
	return a.addNames(s, o, add)
}
