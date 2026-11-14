package format

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tailscale/hujson"
)

// This file holds the read and the write of a shared file in each format:
// JSON in the form of jq, JSONC with hujson, YAML and TOML with yq (so the
// comments of the live file stay), and a text file by line.

func failed(show, reason, stderr string) *Failure {
	return &Failure{Code: 2, Text: reason, Line: show + " " + reason, Stderr: stderr}
}

// specFailure is a spec or an override that the merge cannot use.
func specFailure(msg string) *Failure {
	return &Failure{Code: 2, Text: msg, Line: msg}
}

// Marker gives the comment marker of a format, or "" when it has no comments.
func Marker(f Format, sp *Spec) string {
	switch f {
	case YAML, TOML:
		return "#"
	case JSONC:
		return "//"
	case Text:
		if sp != nil {
			return sp.Comment
		}
	}
	return ""
}

// MergeShared merges the rendered spec of a shared file into its live file.
// The live file stays as it is on a failure.
func MergeShared(in SharedInput) (SharedResult, *Failure) {
	var res SharedResult
	sp, err := ParseSpec(in.Spec, in.SpecShow)
	if err != nil {
		return res, specFailure(err.Error())
	}
	f := FormatOf(sp, in.Explicit, in.Target)
	marker := Marker(f, sp)
	items := sp.expand()
	ov := Value{}
	if len(in.Overrides) > 0 {
		if v, ok := oneValue(in.Overrides); ok {
			ov = v
		}
	}
	used, err := overrideItems(items, ov)
	if err != nil {
		return res, specFailure(in.Show + ": " + err.Error())
	}
	if u := unused(ov, nil, used); u != "" {
		return res, specFailure(in.Show + ": the override " + u + " has no item of bootstrap - remove it")
	}

	live := stripHeader(in.Live, marker, in.Header)
	absent := strings.TrimFunc(string(live), isBashSpace) == ""
	var cur Value
	var hv *hujson.Value
	if !absent {
		var fl *Failure
		cur, absent, hv, fl = readLive(live, f, in.Show)
		if fl != nil {
			return res, fl
		}
	}
	if absent {
		cur = objv(newObject())
		if f == Text {
			cur = arr(nil)
		}
	}
	d := &doc{v: cur}
	res.Records = mergeItems(d, items, parseRecords(in.Records), func(w string) {
		res.Warnings = append(res.Warnings, in.Show+": "+w)
	})

	var out []byte
	var fl *Failure
	switch {
	case f == Text:
		out = textOut(d.v)
	case !absent && len(d.ops) == 0:
		out = live
	case f == JSON:
		out = d.v.marshal()
	case f == JSONC && absent:
		out = d.v.marshal()
	case f == JSONC:
		out, fl = patchJSONC(hv, d.ops, in.Show)
	case absent:
		out, fl = fromJSON(d.v, f)
	default:
		out, fl = applyYQ(live, f, d.ops)
	}
	if fl != nil {
		return res, fl
	}
	res.Out = addHeader(out, marker, in.Header)
	return res, nil
}

// readLive parses the live file. It gives absent for a YAML file without a
// document.
func readLive(live []byte, f Format, show string) (Value, bool, *hujson.Value, *Failure) {
	switch f {
	case Text:
		return textLines(live), false, nil, nil
	case JSON:
		vs, perr := parseStream(live)
		if perr != nil || len(vs) != 1 {
			msg := ""
			if perr != nil {
				msg = perr.Error() + "\n"
			}
			return Value{}, false, nil, failed(show, "is not valid json - kept it as it is", msg)
		}
		if vs[0].k != kObject {
			return Value{}, false, nil, failed(show, "holds "+kindWord(vs[0].typeName())+", not a table - kept it as it is", "")
		}
		return vs[0], false, nil, nil
	case JSONC:
		hv, err := hujson.Parse(live)
		if err != nil {
			return Value{}, false, nil, failed(show, "is not valid jsonc - kept it as it is", err.Error()+"\n")
		}
		std := hv.Clone()
		std.Standardize()
		v, ok := oneValue(std.Pack())
		if !ok || v.k != kObject {
			return Value{}, false, nil, failed(show, "holds no table - kept it as it is", "")
		}
		return v, false, &hv, nil
	}
	dir, err := os.MkdirTemp("", "bootstrap-merge.")
	if err != nil {
		return Value{}, false, nil, &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
	}
	defer func() { _ = os.RemoveAll(dir) }()
	p := filepath.Join(dir, "live")
	if err := os.WriteFile(p, live, 0o600); err != nil {
		return Value{}, false, nil, &Failure{Code: 1, Text: "cannot write a temp file", Line: "cannot write a temp file"}
	}
	if f == YAML {
		r, yerr := yq(nil, "ea", "-p", "yaml", "-o", "json", "[.] | length", p)
		if yerr != nil {
			return Value{}, false, nil, failed(show, "is not valid yaml - kept it as it is", r.stderr)
		}
		if strings.TrimRight(r.stdout, "\n") != "1" {
			return Value{}, false, nil, failed(show, "has more than one YAML document - kept it as it is", r.stderr)
		}
	}
	r, yerr := yq(nil, "eval", "-p", string(f), "-o", "json", ".", p)
	if yerr != nil {
		return Value{}, false, nil, failed(show, "is not valid "+string(f)+" - kept it as it is", r.stderr)
	}
	v, ok := oneValue([]byte(r.stdout))
	switch {
	case !ok:
		return Value{}, false, nil, failed(show, "is not valid "+string(f)+" - kept it as it is", "")
	case v.k == kNull && f == YAML:
		return Value{}, true, nil, nil
	case v.k == kNull:
		return objv(newObject()), false, nil, nil
	case v.k != kObject:
		return Value{}, false, nil, failed(show, "holds "+kindWord(v.typeName())+", not a table - kept it as it is", "")
	}
	return v, false, nil, nil
}

// textLines gives the lines of a text file as a list of strings.
func textLines(b []byte) Value {
	s := strings.TrimSuffix(string(b), "\n")
	var out []Value
	for _, l := range strings.Split(s, "\n") {
		out = append(out, str(l))
	}
	return arr(out)
}

func textOut(v Value) []byte {
	if len(v.a) == 0 {
		return nil
	}
	var b strings.Builder
	for _, l := range v.a {
		b.WriteString(l.s + "\n")
	}
	return []byte(b.String())
}

// patchJSONC applies the edits to the live JSONC value as a JSON patch, so
// its comments stay.
func patchJSONC(hv *hujson.Value, ops []op, show string) ([]byte, *Failure) {
	var cur Value
	std := hv.Clone()
	std.Standardize()
	cur, _ = oneValue(std.Pack())
	var patch []Value
	for _, o := range ops {
		p := newObject()
		p.set("path", str(pointer(o.path)))
		switch {
		case o.del:
			p.set("op", str("remove"))
			cur = delIn(cur, o.path)
		default:
			name := "add"
			if i, isIdx := o.path[len(o.path)-1].(int); isIdx {
				if parent, _ := at(cur, o.path[:len(o.path)-1]); i < len(parent.a) {
					name = "replace"
				}
			}
			p.set("op", str(name))
			p.set("value", o.value)
			cur = setIn(cur, o.path, o.value)
		}
		patch = append(patch, objv(p))
	}
	if err := hv.Patch(arr(patch).marshal()); err != nil {
		return nil, failed(show, "cannot be merged - kept it as it is", err.Error()+"\n")
	}
	return hv.Pack(), nil
}

// fromJSON writes <v> in the format <f> with yq, for an absent live file.
func fromJSON(v Value, f Format) ([]byte, *Failure) {
	if f == TOML && len(v.o.keys) == 0 {
		return nil, nil
	}
	dir, err := os.MkdirTemp("", "bootstrap-merge.")
	if err != nil {
		return nil, &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
	}
	defer func() { _ = os.RemoveAll(dir) }()
	p := filepath.Join(dir, "new.json")
	if err := os.WriteFile(p, v.marshal(), 0o600); err != nil {
		return nil, &Failure{Code: 1, Text: "cannot write a temp file", Line: "cannot write a temp file"}
	}
	r, yerr := yq(nil, "eval", "-p", "json", "-o", string(f), `(.. | (., key)) style = ""`, p)
	if yerr != nil {
		return nil, &Failure{Code: 1, Stderr: r.stderr}
	}
	return []byte(r.stdout), nil
}

// applyYQ applies the edits to the live YAML or TOML text with yq, so its
// comments stay.
func applyYQ(live []byte, f Format, ops []op) ([]byte, *Failure) {
	dir, err := os.MkdirTemp("", "bootstrap-merge.")
	if err != nil {
		return nil, &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
	}
	defer func() { _ = os.RemoveAll(dir) }()
	livePath := filepath.Join(dir, "live")
	opsPath := filepath.Join(dir, "ops.json")
	var list []Value
	expr := []string{`(load(strenv(BOOTSTRAP_MERGE_OPS)) | ((.. | (., key)) style = "")) as $o`}
	for i, o := range ops {
		e := newObject()
		var path []Value
		for _, s := range o.path {
			switch k := s.(type) {
			case string:
				path = append(path, str(k))
			case int:
				path = append(path, intVal(k))
			}
		}
		e.set("path", arr(path))
		e.set("value", o.value)
		list = append(list, objv(e))
		n := strconv.Itoa(i)
		if o.del {
			expr = append(expr, "delpaths([$o["+n+"].path])")
		} else {
			expr = append(expr, "setpath($o["+n+"].path; $o["+n+"].value)")
		}
	}
	if os.WriteFile(livePath, live, 0o600) != nil || os.WriteFile(opsPath, arr(list).marshal(), 0o600) != nil {
		return nil, &Failure{Code: 1, Text: "cannot write a temp file", Line: "cannot write a temp file"}
	}
	r, yerr := yq([]string{"BOOTSTRAP_MERGE_OPS=" + opsPath}, "eval", "-p", string(f), "-o", string(f), strings.Join(expr, " | "), livePath)
	if yerr != nil {
		return nil, &Failure{Code: 1, Stderr: r.stderr}
	}
	return []byte(r.stdout), nil
}

// keepFirst tells if a first line stays above the shared header: a YAML
// document start, a TOML schema line or a shebang.
func keepFirst(line string) bool {
	t := strings.TrimSpace(line)
	return t == "---" || strings.HasPrefix(t, "#:schema") || strings.HasPrefix(t, "# :schema") || strings.HasPrefix(t, "#!")
}

// ownedPrefix starts the first line of the header of an owned file.
const ownedPrefix = "Synced from bootstrap: "

// HeaderPrefix starts the first line of the header of a shared file.
const HeaderPrefix = "Shared with bootstrap: "

// stripHeader removes the shared header from the top of <b>: a line that
// starts with HeaderPrefix, and each next line that is the next line of
// <header>.
func stripHeader(b []byte, marker string, header []string) []byte {
	if marker == "" || len(b) == 0 {
		return b
	}
	lines := strings.SplitAfter(string(b), "\n")
	start := 0
	if len(lines) > 0 && keepFirst(lines[0]) {
		start = 1
	}
	// The header of an owned file, from the time when bootstrap wrote the
	// whole file: its two lines go.
	if start < len(lines) && strings.HasPrefix(lines[start], marker+" "+ownedPrefix) {
		end := start + 1
		if end < len(lines) && strings.HasPrefix(lines[end], marker+" The next sync ") {
			end++
		}
		return []byte(strings.Join(lines[:start], "") + strings.Join(lines[end:], ""))
	}
	if start >= len(lines) || !strings.HasPrefix(lines[start], marker+" "+HeaderPrefix) {
		return b
	}
	end := start + 1
	for i := 1; i < len(header) && end < len(lines) && strings.TrimRight(lines[end], "\r\n") == marker+" "+header[i]; i++ {
		end++
	}
	return []byte(strings.Join(lines[:start], "") + strings.Join(lines[end:], ""))
}

// addHeader puts the shared header at the top of <b>, after a line that
// keepFirst keeps.
func addHeader(b []byte, marker string, header []string) []byte {
	if marker == "" || len(header) == 0 {
		return b
	}
	var h strings.Builder
	for _, l := range header {
		h.WriteString(marker + " " + l + "\n")
	}
	s := string(b)
	first, rest, cut := strings.Cut(s, "\n")
	if cut && keepFirst(first) {
		return []byte(first + "\n" + h.String() + rest)
	}
	return []byte(h.String() + s)
}
