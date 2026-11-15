package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/schema"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// The files of the values folder, by the first segment of a key.
const (
	fileRepo     = "repo"
	fileTools    = "tools"
	fileReleases = "releases"
)

var fileNames = []string{fileRepo, fileTools, fileReleases}

// The release levels, lowest first (D10).
var levels = []string{"NONE", "PATCH", "MINOR", "MAJOR"}

// needsSync tells if a change of a key of <file> (schema key <k>) needs a
// sync (D26, D90): each tools.toml key, and each repo.toml key with Sync in
// the schema.
func needsSync(file string, k *schema.Key) bool {
	switch file {
	case fileTools:
		return true
	case fileRepo:
		return k != nil && k.Sync
	}
	return false
}

// configUsage is the usage line of config.
const configUsage = "bootstrap config get <key> | set <key> <value> | unset <key> | show [repo|tools|releases] [--format json|yaml|toml] [--json]"

// configCmd is `bootstrap config get|set|unset|show`: it reads and writes
// the values of the repo in .config/bootstrap/ (D3, D21-D27).
func (a *app) configCmd(o opts) fail.Code {
	if len(o.names) == 0 {
		a.usageErr("config needs get, set, unset or show; the settings of the engine are bootstrap settings", configUsage)
		return fail.Usage
	}
	sub, args := o.names[0], o.names[1:]
	want := map[string][2]int{"get": {1, 1}, "set": {2, 2}, "unset": {1, 1}, "show": {0, 1}}
	n, ok := want[sub]
	if !ok {
		a.usageErr("config has no subcommand "+sub+" - use get, set, unset or show; the settings of the engine are bootstrap settings", configUsage)
		return fail.Usage
	}
	a.command = "config " + sub
	if len(args) < n[0] || len(args) > n[1] {
		a.usageErr("config "+sub+" takes the wrong number of arguments", configUsage)
		return fail.Usage
	}
	if o.format != "" && (sub == "set" || sub == "unset") {
		a.usageErr("config "+sub+" has no flag --format", configUsage)
		return fail.Usage
	}
	if o.json && o.format != "" {
		a.usageErr("config "+sub+" takes --json or --format, not both", configUsage)
		return fail.Usage
	}
	switch o.format {
	case "":
		if o.json {
			// Run prints the envelope; printDoc keeps the data.
			o.format = "json"
			break
		}
		o.format = "json"
		if f, ok := a.stdout.(*os.File); ok && report.IsTerminal(f) {
			o.format = "toml"
		}
	case "json", "yaml", "toml":
	default:
		a.usageErr("--format needs json, yaml or toml, not "+o.format, configUsage)
		return fail.Usage
	}
	sd, code := a.valuesSide()
	if code != fail.OK {
		return code
	}
	switch sub {
	case "get":
		return a.configGet(sd, args[0], o.format)
	case "show":
		file := ""
		if len(args) == 1 {
			file = args[0]
		}
		return a.configShow(sd, file, o.format)
	case "set":
		return a.configSet(sd, args[0], args[1])
	}
	return a.configUnset(sd, args[0])
}

// valuesSide finds the values folder of config: BOOTSTRAP_VALUES, which
// skips the stop of the old form (D19), else .config/bootstrap/ of the
// repository. A repository without the folder stops with code 2 (D24).
func (a *app) valuesSide() (core.Side, fail.Code) {
	sd := core.Side{Pwd: a.pwd}
	if v := strings.TrimSuffix(a.getenv("BOOTSTRAP_VALUES"), "/"); v != "" {
		sd.Root, sd.Values = a.getenv("MISE_PROJECT_ROOT"), v
	} else {
		root := a.getenv("MISE_PROJECT_ROOT")
		if root == "" {
			top, ok := core.Toplevel(a.pwd)
			if !ok {
				a.fail(a.pwd+" is not a git repository", "")
				return sd, fail.Usage
			}
			root = top
		}
		sd.Root, sd.Values = root, values.FolderOf(root)
		if err := core.CheckLayout(sd, false); err != nil {
			return sd, a.logErr(err)
		}
	}
	if !core.Exists(sd.Values) {
		a.fail(sd.Show(sd.Values)+" is not there", "run bootstrap init")
		return sd, fail.Usage
	}
	for _, f := range []string{values.RepoFile(sd.Values), values.ToolsFile(sd.Values)} {
		if !core.Exists(f) {
			a.fail(sd.Show(sd.Values)+" has no "+filepath.Base(f), "run bootstrap init")
			return sd, fail.Usage
		}
	}
	return sd, fail.OK
}

// splitKey splits the key <key> into its segments. A segment with a dot is
// double-quoted, as in yq: repo.branches."release.1".landing (D49).
func splitKey(key string) ([]string, bool) {
	var segs []string
	for len(key) > 0 {
		var seg string
		if strings.HasPrefix(key, `"`) {
			end := strings.Index(key[1:], `"`)
			if end < 0 {
				return nil, false
			}
			seg, key = key[1:1+end], key[2+end:]
			if key != "" && !strings.HasPrefix(key, ".") {
				return nil, false
			}
		} else {
			i := strings.Index(key, ".")
			if i < 0 {
				i = len(key)
			}
			seg, key = key[:i], key[i:]
		}
		if seg == "" {
			return nil, false
		}
		segs = append(segs, seg)
		if key != "" {
			key = key[1:]
			if key == "" {
				return nil, false
			}
		}
	}
	return segs, len(segs) > 0
}

// valuesKey is a key of config: its file, its path in the file, and its
// schema key (nil for a tools.toml key).
type valuesKey struct {
	text  string
	file  string
	path  []string
	key   *schema.Key
	shown string // The path of the file for a message.
}

// keyOf parses and checks the key path <text>, with no catalog (D23). An
// invalid key is code 2.
func (a *app) keyOf(sd core.Side, text string) (*valuesKey, fail.Code) {
	segs, ok := splitKey(text)
	bad := func(why string) (*valuesKey, fail.Code) {
		a.fail(text+" is not a key of the values: "+why, "a key starts with repo., tools. or releases., for example repo.repo_name")
		return nil, fail.Usage
	}
	if !ok {
		return bad("it is not a key path")
	}
	k := &valuesKey{text: text, file: segs[0], path: segs[1:]}
	if len(k.path) == 0 {
		return bad("it names a file and no key")
	}
	switch k.file {
	case fileRepo:
		k.key = schema.Find(schema.Repo, k.path)
		k.shown = sd.Show(values.RepoFile(sd.Values))
	case fileReleases:
		k.key = schema.Find(schema.Releases, k.path)
		k.shown = sd.Show(values.ReleasesFile(sd.Values))
	case fileTools:
		k.shown = sd.Show(values.ToolsFile(sd.Values))
		return k, fail.OK
	default:
		return bad("the first segment is not repo, tools or releases")
	}
	if k.key == nil {
		return bad("the schema of " + k.file + ".toml has no such key")
	}
	return k, fail.OK
}

// filePath gives the file of the key <k>.
func (k *valuesKey) filePath(sd core.Side) string {
	switch k.file {
	case fileRepo:
		return values.RepoFile(sd.Values)
	case fileTools:
		return values.ToolsFile(sd.Values)
	}
	return values.ReleasesFile(sd.Values)
}

// docPath gives the path of the key in its file: below [tools] for a key of
// tools.toml.
func (k *valuesKey) docPath() []string {
	if k.file == fileTools {
		return append([]string{"tools"}, k.path...)
	}
	return k.path
}

// yqPath gives the path of the key in its file as a yq path.
func (k *valuesKey) yqPath() string {
	b := "."
	for _, s := range k.docPath() {
		b += "[" + strings.TrimSpace(source.JSON(s)) + "]"
	}
	return b
}

// readValues reads the file <path>; a missing file is an empty table. A
// file that does not parse is code 1.
func (a *app) readValues(sd core.Side, path string) (*source.Obj, fail.Code) {
	if !core.Exists(path) {
		return source.NewObj(), fail.OK
	}
	raw, err := source.ReadTOML(path)
	if err != nil {
		a.fail("cannot read "+sd.Show(path)+": "+err.Error(), "")
		return nil, fail.Other
	}
	return raw, fail.OK
}

// lookup gives the value at <path> of <raw>.
func lookup(raw *source.Obj, path []string) (any, bool) {
	var cur any = raw
	for _, s := range path {
		o, ok := cur.(*source.Obj)
		if !ok {
			return nil, false
		}
		if cur, ok = o.Lookup(s); !ok {
			return nil, false
		}
	}
	return cur, true
}

// configGet prints the envelope of the key <text>: its value and its
// source, file or default (D21, D22, D41, D50, D51). Only an invalid key
// (code 2) and a read failure (code 1) are errors.
func (a *app) configGet(sd core.Side, text, form string) fail.Code {
	k, code := a.keyOf(sd, text)
	if code != fail.OK {
		return code
	}
	raw, code := a.readValues(sd, k.filePath(sd))
	if code != fail.OK {
		return code
	}
	env := source.NewObj()
	env.Set("key", text)
	if v, ok := lookup(raw, k.docPath()); ok {
		env.Set("value", v)
		env.Set("source", "file")
	} else {
		var def any
		if k.key != nil {
			def = k.key.Default
		}
		env.Set("value", def)
		env.Set("source", "default")
	}
	return a.printDoc(env, form, true)
}

// configShow prints the file <file> (repo, tools or releases), or the three
// files as one table by file name. A missing releases.toml is an empty
// table (D24).
func (a *app) configShow(sd core.Side, file, form string) fail.Code {
	files := fileNames
	if file != "" {
		ok := false
		for _, f := range fileNames {
			ok = ok || f == file
		}
		if !ok {
			a.usageErr("config show takes repo, tools or releases, not "+file, configUsage)
			return fail.Usage
		}
		files = []string{file}
	}
	doc := source.NewObj()
	for _, f := range files {
		k := &valuesKey{file: f}
		raw, code := a.readValues(sd, k.filePath(sd))
		if code != fail.OK {
			return code
		}
		if file != "" {
			return a.printDoc(raw, form, false)
		}
		doc.Set(f, raw)
	}
	return a.printDoc(doc, form, false)
}

// printDoc prints <doc> on stdout in the format <form>. <compact> prints
// JSON on one line. Under --json, <doc> is the data of the envelope.
func (a *app) printDoc(doc *source.Obj, form string, compact bool) fail.Code {
	text := source.JSON(doc)
	if a.json {
		// Run prints the envelope with <doc> in data.
		a.data = json.RawMessage(text)
		return fail.OK
	}
	if form == "json" && compact {
		_, _ = fmt.Fprintln(a.stdout, text)
		return fail.OK
	}
	dir, err := os.MkdirTemp("", "bootstrap-config.")
	if err != nil {
		a.errorf("cannot create a temp folder")
		return fail.Other
	}
	defer func() { _ = os.RemoveAll(dir) }()
	p := filepath.Join(dir, "doc.json")
	if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
		a.errorf("cannot write a temp file")
		return fail.Other
	}
	out, stderr, err := format.YQ("-p", "json", "-o", form, ".", p)
	if err != nil {
		a.errorf("cannot print the values as %s: %s", form, strings.TrimSpace(stderr))
		return fail.Other
	}
	_, _ = fmt.Fprint(a.stdout, out)
	return fail.OK
}

// catalog opens the clone of a set or an unset of a tools.toml key: the
// cached clone, fetched again when its last fetch is 7 days old (D23, D52).
// The caller defers Cleanup.
func (a *app) catalog() (*core.Clone, *source.Index, fail.Code) {
	c, err := core.OpenClone(core.CloneOptions{
		Getenv: a.getenv, Version: a.version, MaxAge: core.CatalogAge,
		Note: func(s string) { a.infof("%s", s) }, Stderr: a.stderr,
	})
	if err != nil {
		return c, nil, a.logErr(err)
	}
	idx, err := source.Source{Root: c.Dir, Side: source.Repo}.Open()
	if err != nil {
		return c, nil, a.logErr(err)
	}
	return c, idx, fail.OK
}

// toolKey checks the tools.toml key <k> against the catalog <idx> (D86):
// env and files.* are always keys of a tool; any other key is a question or
// a default of the tool.
func (a *app) toolKey(idx *source.Index, k *valuesKey) fail.Code {
	if len(k.path) < 2 {
		a.fail(k.text+" names the table of a tool, not a key", "select or remove a tool with bootstrap add tool or bootstrap remove tool")
		return fail.Usage
	}
	tool := k.path[0]
	if idx.Manifests[tool] == nil {
		a.fail(k.text+" is not a key of the values: "+tool+" is not a bootstrap tool", "")
		return fail.Usage
	}
	if !schema.ToolKey(idx, tool, k.path[1]) {
		a.fail(k.text+" is not a key of the values: "+k.path[1]+" is not a key of the tool "+tool, "its keys are env, files, and its questions and defaults")
		return fail.Usage
	}
	return fail.OK
}

// valueText gives the value <v> of a set as a JSON literal, typed by the
// schema (repo.toml, releases.toml) or by the catalog (tools.toml).
func valueText(idx *source.Index, k *valuesKey, v string) (string, error) {
	asJSON := func() (string, error) {
		if !json.Valid([]byte(v)) {
			return "", fmt.Errorf("%s needs a JSON value, for example '[\"main\",\"develop\"]', not %s", k.text, v)
		}
		return v, nil
	}
	str := func() (string, error) { return strings.TrimSpace(source.JSON(v)), nil }
	if k.key != nil {
		switch k.key.Type {
		case schema.String:
			return str()
		case schema.Integer:
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return "", fmt.Errorf("%s needs an integer, not %s", k.text, v)
			}
			return strconv.FormatInt(n, 10), nil
		}
		return asJSON()
	}
	tool, key := k.path[0], k.path[1]
	if len(k.path) > 2 {
		// A key below a table of a tool: JSON when it is JSON.
		if json.Valid([]byte(v)) {
			return v, nil
		}
		return str()
	}
	if key == "env" || key == "files" {
		return asJSON()
	}
	var def any
	if m := idx.Manifests[tool]; m != nil && m.Raw != nil {
		if d, ok := m.Raw.Get("defaults").(*source.Obj); ok {
			def = d.Get(key)
		}
	}
	switch def.(type) {
	case nil, string:
		return str()
	case bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return "", fmt.Errorf("%s needs true or false, not %s", k.text, v)
		}
		return strconv.FormatBool(b), nil
	case int64:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return "", fmt.Errorf("%s needs an integer, not %s", k.text, v)
		}
		return strconv.FormatInt(n, 10), nil
	case float64:
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return "", fmt.Errorf("%s needs a number, not %s", k.text, v)
		}
		return v, nil
	}
	return asJSON()
}

// configSet writes <v> at the key <text> (D9, D10, D24-D27).
func (a *app) configSet(sd core.Side, text, v string) fail.Code {
	k, code := a.keyOf(sd, text)
	if code != fail.OK {
		return code
	}
	var idx *source.Index
	if k.file == fileTools {
		c, ix, code := a.catalog()
		defer c.Cleanup()
		if code != fail.OK {
			return code
		}
		if code := a.toolKey(ix, k); code != fail.OK {
			return code
		}
		sd.Clone, idx = c.Dir, ix
		sel, err := core.SelectionOf(sd, sd.Values, sd.Values)
		if err != nil {
			return a.logErr(err)
		}
		selected := false
		for _, t := range sel.Tools {
			selected = selected || t == k.path[0]
		}
		if !selected {
			a.fail(k.path[0]+" is not selected, so "+text+" cannot be set", "run bootstrap add tool "+k.path[0])
			return fail.Usage
		}
	}
	lit, err := valueText(idx, k, v)
	if err != nil {
		a.fail(err.Error(), "")
		return fail.Usage
	}
	path := k.filePath(sd)
	cur, err := os.ReadFile(path)
	if err != nil {
		if k.file != fileReleases || core.Exists(path) {
			a.fail("cannot read "+k.shown, "")
			return fail.Other
		}
		// A new releases.toml (D24).
		cur = []byte("format = " + strconv.Itoa(values.Format) + "\n")
	}
	if k.file == fileReleases {
		raw, code := a.readValues(sd, path)
		if code != fail.OK {
			return code
		}
		if old, ok := lookup(raw, k.path); ok {
			o, n := rank(fmt.Sprint(old)), rank(v)
			if o >= 0 && n >= 0 && o > n {
				a.infof("%s stays %s: it is higher than %s", text, old, v)
				// The file keeps the old level, so data holds it (D47).
				a.setData(text, json.RawMessage(strings.TrimSpace(source.JSON(old))))
				return fail.OK
			}
		}
	}
	next, err := core.EditTOML(cur, []string{k.yqPath() + " = " + lit})
	if err != nil {
		return a.logErr(err)
	}
	if code := a.writeValues(sd, k, idx, next); code != fail.OK {
		return code
	}
	if k.file == fileReleases {
		a.infof("%s = %s", text, v)
	}
	if needsSync(k.file, k.key) {
		a.infof("run bootstrap sync to apply it")
	}
	a.setData(text, json.RawMessage(lit))
	return fail.OK
}

// setData keeps the data of config set and config unset for the envelope
// (D47): the key, and for a set the value the file holds after it. <value>
// is nil for an unset.
func (a *app) setData(key string, value json.RawMessage) {
	if !a.json {
		return
	}
	if value == nil {
		a.data = struct {
			Key string `json:"key"`
		}{key}
		return
	}
	a.data = struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}{key, value}
}

// rank gives the place of the release level <l>, or -1.
func rank(l string) int {
	for i, x := range levels {
		if x == l {
			return i
		}
	}
	return -1
}

// configUnset removes the key <text> (D27). A required key stays; a whole
// table can go.
func (a *app) configUnset(sd core.Side, text string) fail.Code {
	k, code := a.keyOf(sd, text)
	if code != fail.OK {
		return code
	}
	var idx *source.Index
	if k.file == fileTools {
		c, ix, code := a.catalog()
		defer c.Cleanup()
		if code != fail.OK {
			return code
		}
		if code := a.toolKey(ix, k); code != fail.OK {
			return code
		}
		idx = ix
	}
	if k.key != nil && k.key.Required {
		a.fail(text+" is a required key of "+k.file+".toml, so config unset cannot remove it", "set another value with bootstrap config set "+text+" <value>")
		return fail.Usage
	}
	path := k.filePath(sd)
	raw, code := a.readValues(sd, path)
	if code != fail.OK {
		return code
	}
	if _, ok := lookup(raw, k.docPath()); !ok {
		a.infof("%s is not set", text)
		a.setData(text, nil)
		return fail.OK
	}
	cur, err := os.ReadFile(path)
	if err != nil {
		a.fail("cannot read "+k.shown, "")
		return fail.Other
	}
	exprs := []string{"del(" + k.yqPath() + ")"}
	if k.file != fileTools {
		// A table that the unset leaves empty goes too. In tools.toml, an
		// empty [tools.<name>] table selects the tool, so it stays.
		for n := len(k.path) - 1; n > 0; n-- {
			p := (&valuesKey{file: k.file, path: k.path[:n]}).yqPath()
			exprs = append(exprs, "del("+p+" | select(length == 0))")
		}
	}
	next, err := core.EditTOML(cur, exprs)
	if err != nil {
		return a.logErr(err)
	}
	if code := a.writeValues(sd, k, idx, next); code != fail.OK {
		return code
	}
	if needsSync(k.file, k.key) {
		a.infof("run bootstrap sync to apply it")
	}
	a.setData(text, nil)
	return fail.OK
}

// writeValues examines the new text <next> of the file of <k> with the
// schema, and the catalog <idx> for tools.toml, and writes it. A text that
// breaks a rule is code 2, and nothing is written (D27).
func (a *app) writeValues(sd core.Side, k *valuesKey, idx *source.Index, next []byte) fail.Code {
	dir, err := os.MkdirTemp("", "bootstrap-config.")
	if err != nil {
		a.errorf("cannot create a temp folder")
		return fail.Other
	}
	defer func() { _ = os.RemoveAll(dir) }()
	tmp := filepath.Join(dir, k.file+".toml")
	if err := os.WriteFile(tmp, next, 0o600); err != nil {
		a.errorf("cannot write a temp file")
		return fail.Other
	}
	raw, err := source.ReadTOML(tmp)
	if err != nil {
		a.fail("cannot read the new "+k.shown+": "+err.Error(), "")
		return fail.Usage
	}
	switch k.file {
	case fileRepo:
		err = schema.CheckRepo(raw, k.shown)
	case fileReleases:
		err = schema.CheckReleases(raw, k.shown)
	default:
		err = schema.CheckToolsShape(raw, k.shown)
		if err == nil && idx != nil {
			err = schema.CheckTools(idx, raw, k.shown)
		}
	}
	if err != nil {
		return a.logErr(err)
	}
	if err := oldform.Apply([]oldform.Change{{Path: k.filePath(sd), Content: next}}); err != nil {
		return a.logErr(err)
	}
	return fail.OK
}
