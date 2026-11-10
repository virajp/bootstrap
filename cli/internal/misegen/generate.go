package misegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/header"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/tmpl"
)

// Generator holds the inputs of one run of the generator. Close removes its
// temp folder.
type Generator struct {
	Side source.Side
	// ToolsDir is the absolute folder of the tools, and Shown is that folder
	// in a message and a header, for example repo/tools.
	ToolsDir string
	Shown    string
	// Values is the values file. Empty: no values file (macos without
	// macos.toml), so no overrides.
	Values string
	// ValuesShown names Values in a message (`_se_path`: relative to the
	// current folder when it is below it).
	ValuesShown string
	// ProjectRoot is MISE_PROJECT_ROOT. On the repo side, the override note
	// names Values relative to it.
	ProjectRoot string
	// All holds the merged values: tool name to its values. Nil: no values.
	All *source.Obj
	// Envs is the result of NewEnvs (repo). Nil on the macos side.
	Envs *Envs
	// Getenv reads HOME, USER, TMPDIR and each env.Getenv of a template. Nil
	// reads the real environment.
	Getenv func(string) string
	// Ships tells if <tool> has an entry for the target <path> (repo). Nil:
	// no tool ships a mise file.
	Ships func(tool, path string) bool
	// Index and Tools are the clone and the selected tools: the generator
	// computes the `depends` of each mise tool from them. Nil: no depends.
	Index *source.Index
	Tools []string
	// Derived holds the `mise` and `mise_env` tables that the derive script
	// of a tool gives. They merge into the mise config of the tool before
	// [overrides.mise].
	Derived map[string]*source.Obj

	work string
}

func (g *Generator) getenv(name string) string {
	if g.Getenv != nil {
		return g.Getenv(name)
	}
	return os.Getenv(name)
}

// Close removes the temp folder of the generator.
func (g *Generator) Close() error {
	if g.work == "" {
		return nil
	}
	err := os.RemoveAll(g.work)
	g.work = ""
	return err
}

// workDir is `_se_mise_work`: the temp folder of the renders.
func (g *Generator) workDir() error {
	if g.work != "" {
		if st, err := os.Stat(g.work); err == nil && st.IsDir() {
			return nil
		}
	}
	base := g.getenv("TMPDIR")
	if base == "" {
		base = "/tmp"
	}
	base = strings.TrimSuffix(base, "/")
	if base == "" {
		base = "/"
	}
	d, err := os.MkdirTemp(base, "bootstrap-mise.")
	if err != nil {
		return fail.New(fail.Other, "cannot create a temp directory")
	}
	g.work = d
	return nil
}

// spaceSep is the value of a tools list in a bash variable: the words with
// one space between them.
func spaceSep(tools []string) string { return strings.Join(tools, " ") }

// envsText is the environments as the caller gives them in bash (_RS_ENVS):
// " dev ci ", or "" when there are none.
func envsText(envs []string) string {
	if len(envs) == 0 {
		return ""
	}
	return words(envs).String()
}

func readTOMLFile(path string) (any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return tmpl.DecodeTOML(b)
}

// envWith gives the Getenv of a template: <vars> first, then g.getenv.
func (g *Generator) envWith(vars map[string]string) func(string) string {
	return func(name string) string {
		if v, has := vars[name]; has {
			return v
		}
		return g.getenv(name)
	}
}

// Overrides is `_se_mise_overrides`. It examines the `mise` and `mise_env`
// keys of `[overrides]` of the values file against the tools.toml of each
// tool of <tools>. Each value must replace a value that a selected tool has,
// at the same path below [mise] or [mise_env.<env>]. <envs> holds the
// environments of the values file (repo); empty when the caller already
// reported their problem. <local> is for `repo_sync --check`: it needs no
// tools, and it skips the check of each path. The error holds a message (code
// 2) for each problem.
func (g *Generator) Overrides(tools, envs []string, local bool) error {
	if err := checkSide(g.Side); err != nil {
		return err
	}
	if local {
		tools = nil
	}
	envw := envsText(envs)
	ctx := map[string]any{}
	bad := func() error { return usage("cannot read [overrides] of %s", g.ValuesShown) }
	o, err := readTOMLFile(g.Values)
	if err != nil {
		return bad()
	}
	ctx["o"] = o
	for _, tool := range tools {
		t, err := readTOMLFile(g.ToolsDir + "/" + tool + "/tools.toml")
		if err != nil {
			return bad()
		}
		ctx["tool_"+tool] = t
	}
	out, err := tmpl.Render("<arg>", []byte(overridesTemplate), ctx, &tmpl.Options{Env: g.envWith(map[string]string{
		"BOOTSTRAP_MISE_SIDE":  string(g.Side),
		"BOOTSTRAP_MISE_ENVS":  envw,
		"BOOTSTRAP_MISE_TOOLS": spaceSep(tools),
	})})
	if err != nil {
		return bad()
	}
	lockHint := " - use env of [tools.<tool>]"
	if g.Side != source.Repo {
		lockHint = ""
	}
	show := g.ValuesShown
	var list []*fail.Error
	for _, line := range bashLines(string(out)) {
		problem, path, arg := readTabs(line)
		var msg string
		switch problem {
		case "table":
			msg = fmt.Sprintf("%s in %s must be a table", path, show)
		case "noenv":
			msg = fmt.Sprintf("%s in %s: the macos side has no environments - use [overrides.mise.<path>]", path, show)
		case "envname", "env":
			if strings.ReplaceAll(envw, " ", "") == "" {
				continue
			}
			if problem == "envname" {
				msg = fmt.Sprintf("%s in %s: '%s' is not one of %s", path, show, arg, envNamesText)
			} else {
				msg = fmt.Sprintf("%s in %s: '%s' is not in environments:%s", path, show, arg, strings.TrimSuffix(envw, " "))
			}
		case "lock":
			msg = fmt.Sprintf("%s in %s: the envs, the depends, the requires, the requires_found and the files of a mise tool cannot be overridden%s", path, show, lockHint)
		case "none":
			if local {
				continue
			}
			msg = fmt.Sprintf("%s in %s: no selected tool has %s - the path after overrides. is the path in %s/<tool>/tools.toml", path, show, arg, g.Shown)
		default:
			continue
		}
		list = append(list, usage("%s", msg))
	}
	return source.Problems(list)
}

// bashLines splits <s> as `while read` does: each line, also a last line
// without a newline.
func bashLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// readTabs is `IFS=$'\t' read -r a b c`: the tabs at the start and the end go,
// a run of tabs splits two fields, and c keeps the rest of the line.
func readTabs(line string) (a, b, c string) {
	line = strings.Trim(line, "\t")
	field := func() string {
		i := strings.IndexByte(line, '\t')
		if i < 0 {
			f := line
			line = ""
			return f
		}
		f := line[:i]
		line = strings.TrimLeft(line[i:], "\t")
		return f
	}
	a = field()
	b = field()
	return a, b, line
}

// Render is `_se_mise_render`: it generates the mise config of <tool> into
// <confd>/<tool>/. Repo: one mise.<env>.toml for each selected environment
// with content. Macos: one mise.toml. It gives the written files as
// `<tool>/<name>` (_SE_MISE_WRITTEN), also after a failure. A failure writes
// nothing more for this tool, and the existing files of the tool stay.
func (g *Generator) Render(tool, confd string) ([]string, error) {
	if err := checkSide(g.Side); err != nil {
		return nil, err
	}
	if err := g.workDir(); err != nil {
		return nil, err
	}
	file := g.Shown + "/" + tool + "/tools.toml"
	src, err := g.read(g.ToolsDir+"/"+tool+"/tools.toml", file, tool, func(all map[string]any) map[string]any {
		cfg := all[tool]
		if cfg == nil || cfg == false {
			cfg = map[string]any{}
		}
		return map[string]any{"config": cfg, "all": all}
	})
	if err != nil {
		return nil, err
	}
	out := filepath.Join(confd, tool)
	var written []string
	if g.Side == source.Macos {
		note := "# override: " + MacosValues
		body, full, err := g.generate(tool, src, g.Values, "", "1", note, file)
		if err != nil {
			return nil, err
		}
		if full {
			w, err := put(file, out, "mise.toml", body)
			if err != nil {
				return written, err
			}
			written = append(written, w)
		}
		return written, nil
	}
	note := "# override: " + strings.TrimPrefix(g.Values, g.ProjectRoot+"/")
	var selected []string
	if g.Envs != nil {
		selected = g.Envs.Selected
	}
	for _, env := range selected {
		name := "mise." + env + ".toml"
		in := "0"
		if g.Envs.In(tool, env) {
			in = "1"
		}
		body, full, err := g.generate(tool, src, g.Values, env, in, note, file)
		if err != nil {
			return written, err
		}
		if !full {
			continue
		}
		shipped := g.Ships != nil && g.Ships(tool, ".config/mise/conf.d/"+tool+"/"+name)
		if _, err := os.Stat(filepath.Join(out, name)); err == nil || shipped {
			return written, usage("%s/%s ships .config/mise/conf.d/%s/%s and also generates it from [mise] - remove one", g.Shown, tool, tool, name)
		}
		w, err := put(file, out, name, body)
		if err != nil {
			return written, err
		}
		written = append(written, w)
	}
	return written, nil
}

// OwnConfig is `_se_mise_own_config`, only on the macos side: it generates
// the [mise.*] tables of the values file into <confd>/zz-macos/mise.toml.
// The `.all` of a render holds the values of the tools of <tools>. A values
// file without a [mise] table gets no file. It gives the written files, as
// Render does.
func (g *Generator) OwnConfig(confd string, tools []string) ([]string, error) {
	if err := checkSide(g.Side); err != nil {
		return nil, err
	}
	if g.Side != source.Macos {
		return nil, usage("misegen: the own config is only for the side macos")
	}
	v, err := readTOMLFile(g.Values)
	if err != nil {
		return nil, usage("cannot read %s", g.ValuesShown)
	}
	if m, isMap := v.(map[string]any); !isMap || !hasKey(m, "mise") {
		return nil, nil
	}
	if err := g.workDir(); err != nil {
		return nil, err
	}
	// The bash list is `printf '%s ' $tools`, then split(" ") in jq.
	list := ""
	for _, t := range tools {
		list += t + " "
	}
	if len(tools) == 0 {
		list = " "
	}
	keep := map[string]bool{}
	for _, t := range strings.Split(list, " ") {
		keep[t] = true
	}
	src, err := g.read(g.Values, MacosValues, "zz-macos", func(all map[string]any) map[string]any {
		some := map[string]any{}
		for k, v := range all {
			if keep[k] {
				some[k] = v
			}
		}
		return map[string]any{"all": some}
	})
	if err != nil {
		return nil, err
	}
	body, full, err := g.generate("", src, "", "", "1", "", MacosValues)
	if err != nil || !full {
		return nil, err
	}
	w, err := put(MacosValues+" [mise]", filepath.Join(confd, "zz-macos"), "mise.toml", body)
	if err != nil {
		return nil, err
	}
	return []string{w}, nil
}

func hasKey(m map[string]any, k string) bool {
	_, has := m[k]
	return has
}

// read is `_se_mise_read`. It gives the file that the generator reads for
// <file>: <file> itself, or a temp copy, rendered with the template engine,
// when it holds `<%`. <shown> names <file> in a message, and <name> names the
// temp files. <build> builds the context from the merged values; .env is added
// to it. The error (code 2) is a render that fails, a copy that is not valid
// TOML, or a copy with other key paths than <file>.
func (g *Generator) read(file, shown, name string, build func(all map[string]any) map[string]any) (string, error) {
	raw, err := os.ReadFile(file)
	if err != nil || !bytes.Contains(raw, []byte("<%")) {
		// grep fails on a file that it cannot read: the caller reads it.
		return file, nil
	}
	all := map[string]any{}
	if g.All != nil {
		v, err := tmpl.DecodeJSON([]byte(source.JSON(g.All)))
		m, isMap := v.(map[string]any)
		if err != nil || !isMap {
			return "", usage("%s: cannot build the context of the render", shown)
		}
		all = m
	}
	ctx := build(all)
	ctx["env"] = map[string]any{"HOME": g.getenv("HOME"), "USER": g.getenv("USER")}
	copyPath := filepath.Join(g.work, name+".toml")
	// `--missing-key default`: `required` must see a key that is missing.
	out, err := tmpl.Render(file, raw, ctx, &tmpl.Options{MissingKey: "default", Env: g.getenv})
	if err != nil {
		return "", usage("%s: cannot render: %s", shown, renderMessage(file, err))
	}
	// gomplate writes no file for an empty output.
	if len(out) > 0 {
		if err := os.WriteFile(copyPath, out, 0o644); err != nil {
			return "", fail.New(fail.Other, "cannot write %s", copyPath)
		}
	}
	if bytes.Contains(out, []byte("<no value>")) {
		return "", usage("%s: cannot render: an expression has no value - a key is missing", shown)
	}
	copyJSON, stderr, err := format.YQ("-p", "toml", "-o", "json", ".", copyPath)
	if err != nil {
		return "", usage("%s: not valid TOML after render: %s", shown, firstLine(stderr))
	}
	rawJSON, stderr, err := format.YQ("-p", "toml", "-o", "json", ".", file)
	if err != nil {
		return "", usage("%s: not valid TOML before render: %s", shown, firstLine(stderr))
	}
	a, errA := jsonPaths(rawJSON)
	b, errB := jsonPaths(copyJSON)
	if errA != nil || errB != nil || !equalPaths(a, b) {
		return "", usage("%s: a rendered value changes the keys of the file", shown)
	}
	return copyPath, nil
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// renderMessage gives the message of a failed render of a file, as the bash
// engine cuts it from the JSON log line of gomplate.
func renderMessage(file string, err error) string {
	var full string
	var ee template.ExecError
	if errors.As(err, &ee) {
		full = "renderTemplate: failed to render template " + file + ": " + err.Error()
	} else if strings.HasPrefix(err.Error(), "template: ") {
		full = "renderTemplate: parse template " + file + ": " + err.Error()
	} else {
		full = "renderTemplate: failed to render template " + file + ": " + err.Error()
	}
	line := jsonEscape(full)
	if loc := callRe.FindStringIndex(line); loc != nil {
		return line[loc[1]:]
	}
	return line
}

var (
	callRe = regexp.MustCompile(`.*error calling [A-Za-z.]*: `)
	failRe = regexp.MustCompile(`.*error calling Fail: template generation failed: `)
)

// jsonEscape escapes a string as the JSON log handler of gomplate (log/slog)
// does, without the quotes: the bash engine reads the message from that line.
func jsonEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			b.WriteString(`�`)
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20:
			fmt.Fprintf(&b, `\u%04x`, r)
		case r == ' ' || r == ' ':
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}

// jsonPaths is jq `[paths]` of a JSON text, as a set.
func jsonPaths(text string) (map[string]bool, error) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	set := map[string]bool{}
	var walk func(v any, p []any)
	walk = func(v any, p []any) {
		if len(p) > 0 {
			k, _ := json.Marshal(p)
			set[string(k)] = true
		}
		switch x := v.(type) {
		case map[string]any:
			for k, e := range x {
				walk(e, append(append([]any(nil), p...), k))
			}
		case []any:
			for i, e := range x {
				walk(e, append(append([]any(nil), p...), i))
			}
		}
	}
	walk(v, nil)
	return set, nil
}

func equalPaths(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// generate is `_se_mise_generate`: one output of the file <src>, with
// `[overrides]` of <values> applied (empty: the own config, no overrides).
// Each line that an override changed gets the comment <note>. <full> is true
// when the output has content. The error (code 2) names <shown>.
func (g *Generator) generate(tool, src, values, env, in, note, shown string) (body []byte, full bool, err error) {
	ctx := map[string]any{}
	own := "1"
	var terr error
	if t, err := readTOMLFile(src); err == nil {
		ctx["t"] = g.extend(tool, env, t)
	} else {
		terr = err
	}
	if values != "" {
		own = ""
		if o, err := readTOMLFile(values); err == nil {
			ctx["o"] = o
		} else {
			terr = err
		}
	}
	miseEnvs := ""
	if g.Envs != nil {
		miseEnvs = g.Envs.MiseWords()
	}
	var out []byte
	if terr == nil {
		out, terr = tmpl.Render("<arg>", []byte(miseTemplate), ctx, &tmpl.Options{Env: g.envWith(map[string]string{
			"BOOTSTRAP_MISE_SIDE": string(g.Side),
			"BOOTSTRAP_MISE_OWN":  own,
			"BOOTSTRAP_MISE_ENV":  env,
			"BOOTSTRAP_MISE_ENVS": miseEnvs,
			"BOOTSTRAP_MISE_IN":   in,
		})})
	}
	if terr != nil {
		msg := ""
		line := jsonEscape(terr.Error())
		if loc := failRe.FindStringIndex(line); loc != nil {
			msg = line[loc[1]:]
		}
		return nil, false, usage("%s: cannot generate the mise config: %s", shown, msg)
	}
	if len(out) == 0 {
		return nil, false, nil
	}
	body = markOverrides(out, note)
	full = strings.ContainsFunc(string(body), func(r rune) bool { return r != ' ' && r != '\n' })
	return body, full, nil
}

var (
	headRe   = regexp.MustCompile(`^[ \t]*\[`)
	headCut  = regexp.MustCompilePOSIX(`^[ \t]*\[+|\]+[ \t]*$|"`)
	keyStart = regexp.MustCompilePOSIX(`^[ \t]+|"`)
)

// markOverrides is the awk step of `_se_mise_generate`: the `#@override
// <path>` lines name the overridden keys. Each of them gets <note> in the
// TOML, and those lines go. <path> is the table header and the key, joined
// with dots and without quotes.
func markOverrides(out []byte, note string) []byte {
	lines := bashLines(string(out))
	const prefix = "#@override "
	mark := map[string]bool{}
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			mark[l[len(prefix):]] = true
		}
	}
	var b strings.Builder
	head := ""
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			continue
		}
		if headRe.MatchString(l) {
			head = headCut.ReplaceAllString(l, "")
		} else if i := strings.Index(l, " = "); i >= 0 {
			key := keyStart.ReplaceAllString(l[:i], "")
			full := key
			if head != "" {
				full = head + "." + key
			}
			if mark[full] {
				l += "  " + note
			}
		}
		b.WriteString(l + "\n")
	}
	return []byte(b.String())
}

// put is `_se_mise_put`: it writes <body> to <folder>/<name>, below the
// header that names <source>, with mode 644. It gives `<folder name>/<name>`.
func put(sourceName, folder, name string, body []byte) (string, error) {
	path := filepath.Join(folder, name)
	data := "# " + header.Line1(sourceName) + "\n# " + header.Line2 + "\n\n" + string(body)
	if err := os.MkdirAll(folder, 0o777); err != nil {
		return "", &fail.Error{Code: fail.Other, Err: err}
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return "", &fail.Error{Code: fail.Other, Err: err}
	}
	if err := os.Chmod(path, 0o644); err != nil {
		return "", &fail.Error{Code: fail.Other, Err: err}
	}
	return filepath.Base(folder) + "/" + name, nil
}
