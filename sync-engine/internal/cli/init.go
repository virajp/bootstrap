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

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/format"
	"github.com/virajp/bootstrap/sync-engine/internal/oldform"
	"github.com/virajp/bootstrap/sync-engine/internal/questions"
	"github.com/virajp/bootstrap/sync-engine/internal/selection"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
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
		s.user = v.User
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
		s.user = v.User
		sub, _ := git("-C", root, "rev-parse", "--show-superproject-working-tree")
		res, err = selection.ResolveRepo(idx, v, sub != "")
		if err != nil {
			return nil, a.logErr(err)
		}
	}
	s.tools = res.Tools
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
	c, code := a.openClone(mac)
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
	answered, code := a.ask(mac, sel, o)
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
	envKey, envVal := "REPO_SYNC_VALUES", tmpValues
	if mac {
		envKey, envVal = "MACOS_SETUP_REPO_PATH", tmp
	}
	old, had := os.LookupEnv(envKey)
	_ = os.Setenv(envKey, envVal)
	oldform.Migrating = true
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
	var syncs []string
	seen := map[string]bool{}
	for _, p := range planned {
		if strings.HasPrefix(p, tmp+"/") || seen[p] {
			continue
		}
		seen[p] = true
		syncs = append(syncs, p)
	}
	sort.Strings(syncs)

	if len(changes) == 0 && len(syncs) == 0 {
		a.infof("Nothing to change")
		return fail.OK
	}
	fmt.Fprintln(a.stdout, "Changes:")
	for _, ch := range changes {
		verb := "write "
		if ch.Remove {
			verb = "remove"
		}
		fmt.Fprintf(a.stdout, "  %s %s - %s\n", verb, a.show(ch.Path), ch.Note)
	}
	for _, p := range syncs {
		fmt.Fprintf(a.stdout, "  sync   %s\n", a.show(p))
	}
	paths := append(oldform.Paths(changes), syncs...)
	if o.dryRun {
		a.warnDirty(dirtyTargets(paths, nil))
		a.infof("Dry run: nothing written")
		return fail.OK
	}
	if code := a.stopDirty(dirtyTargets(paths, nil)); code != fail.OK {
		return code
	}
	if done, code := a.consent(o.yes); done {
		return code
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
	return code
}

// ask gives the answer of each question of the selected tools that has no
// answer, and of each --answer, as (tool, key, value). With --yes, a question
// without an --answer is an error. Without a terminal, too.
func (a *app) ask(mac bool, sel *selected, o opts) ([][3]string, fail.Code) {
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
			a.errorf("--answer %s: no question %s of the selected tools - the form is <tool>.<key>=<value>", ans, name)
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
		if o.yes || !a.interactive() {
			a.errorf("the question %s has no answer - give it with --answer %s=<value>", qq.Name(), qq.Name())
			rc = fail.Usage
			continue
		}
		proposal, _ := questions.Proposal(qq, dir)
		for {
			if proposal != "" {
				fmt.Fprintf(a.stdout, "%s (%s)\n  [%s]: ", qq.Text, qq.Name(), proposal)
			} else {
				fmt.Fprintf(a.stdout, "%s (%s)\n  : ", qq.Text, qq.Name())
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
	c, code := a.openClone(false)
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
				a.errorf("unknown tool %s", n)
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
	if len(rows) == 0 {
		fmt.Fprintln(a.stdout, "No questions.")
		return fail.OK
	}
	for _, r := range rows {
		fmt.Fprintf(a.stdout, "%s.%s: %s\n", r.Tool, r.Key, r.Question)
		if r.Proposal != nil {
			fmt.Fprintf(a.stdout, "  proposal: %s (%s)\n", *r.Proposal, *r.Source)
		}
		if r.Answer != nil {
			fmt.Fprintf(a.stdout, "  answer:   %v\n", r.Answer)
		} else {
			fmt.Fprintln(a.stdout, "  answer:   none")
		}
	}
	return fail.OK
}
