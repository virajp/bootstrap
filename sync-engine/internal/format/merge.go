package format

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseOps reads the two JSON texts that the bash merge gets: <remove>, a list
// of key paths, and <removeItems>, a list of `{path, items}` tables. The
// caller checked their shape (_se_files_check_shape). An empty text is an empty
// list.
func ParseOps(remove, removeItems string) (Ops, error) {
	var ops Ops
	if strings.TrimSpace(remove) != "" {
		vs, perr := parseStream([]byte(remove))
		if perr != nil {
			return ops, perr
		}
		if len(vs) != 1 || vs[0].k != kArray {
			return ops, errors.New("remove is not a list")
		}
		for _, p := range vs[0].a {
			keys, ok := stringList(p)
			if !ok {
				return ops, errors.New("remove holds a path that is not a list of strings")
			}
			ops.Remove = append(ops.Remove, keys)
		}
	}
	if strings.TrimSpace(removeItems) != "" {
		vs, perr := parseStream([]byte(removeItems))
		if perr != nil {
			return ops, perr
		}
		if len(vs) != 1 || vs[0].k != kArray {
			return ops, errors.New("remove_items is not a list")
		}
		for _, r := range vs[0].a {
			keys, ok := stringList(r.get("path"))
			items := r.get("items")
			if r.k != kObject || !ok || items.k != kArray {
				return ops, errors.New("remove_items holds a bad table")
			}
			ops.RemoveItems = append(ops.RemoveItems, RemoveRule{Path: keys, Items: items.a})
		}
	}
	return ops, nil
}

func stringList(v Value) ([]string, bool) {
	if v.k != kArray {
		return nil, false
	}
	out := make([]string, len(v.a))
	for i, e := range v.a {
		if e.k != kString {
			return nil, false
		}
		out[i] = e.s
	}
	return out, true
}

// nonBlank is the test `grep -q '[^[:space:]]'`.
func nonBlank(b []byte) bool {
	return strings.TrimFunc(string(b), isBashSpace) != ""
}

// onlyComments is true when each line is blank, a comment, or `---` or `...`
// (the test of a YAML file that holds no document).
var yamlEmptyLine = regexp.MustCompile(`^[ \t\n\v\f\r]*(#.*)?$|^(---|\.\.\.)[ \t\n\v\f\r]*$`)

func yamlHasNoContent(b []byte) bool {
	for _, line := range strings.Split(string(b), "\n") {
		if !yamlEmptyLine.MatchString(line) {
			return false
		}
	}
	return true
}

// typesOf gives what `jq -r type` prints for the values of <data>: one line
// for each value.
func typesOf(data string) (string, *parseError) {
	vs, err := parseStream([]byte(data))
	if err != nil {
		return "", err
	}
	names := make([]string, len(vs))
	for i, v := range vs {
		names[i] = v.typeName()
	}
	return strings.Join(names, "\n"), nil
}

// oneValue gives the one value of <data>, or false for none or many.
func oneValue(data []byte) (Value, bool) {
	vs, err := parseStream(data)
	if err != nil || len(vs) != 1 {
		return Value{}, false
	}
	return vs[0], true
}

func silent(code int, stderr string) *Failure { return &Failure{Code: code, Stderr: stderr} }

func failed(show, reason, stderr string) *Failure {
	return &Failure{Code: 2, Text: reason, Line: show + " " + reason, Stderr: stderr}
}

// Merge merges the merge source <src> over the live file <live> in format <f>,
// and gives the new content of the target. <live> is nil or holds only
// whitespace when the target is absent or empty. <show> names the target in
// each message. Format <f> is JSON, YAML or TOML. It returns a Failure with code
// 2 for a source or a live file that cannot merge; <live> is never changed.
//
// The steps and the messages are those of _se_merge_run in sync-engine/merge.
func Merge(src, live []byte, f Format, ops Ops, show string) ([]byte, *Failure) {
	dir, err := os.MkdirTemp("", "sync-engine-merge.")
	if err != nil {
		return nil, &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
	}
	defer os.RemoveAll(dir)
	return mergeRun(dir, src, live, f, ops, show)
}

// jqSlurpStderr gives what `jq -s 'if length == 1 then .[0] else error(...)'`
// prints to stderr for <data>: the parse error, or the error of a stream that
// does not hold exactly one value. jq shows the path of the file and the
// number of lines it read.
func jqSlurpStderr(path string, data []byte, perr *parseError) string {
	if perr != nil {
		return perr.Error() + "\n"
	}
	return fmt.Sprintf("jq: error (at %s:%d): more than one JSON value\n", path, strings.Count(string(data), "\n"))
}

func mergeRun(dir string, src, live []byte, f Format, ops Ops, show string) ([]byte, *Failure) {
	file := func(name string) string { return filepath.Join(dir, name) }
	srcPath := file("src")
	if err := os.WriteFile(srcPath, src, 0o600); err != nil {
		return nil, silent(1, err.Error())
	}
	format := string(f)

	// The source: one table; null is an empty source.
	var srcJSON string
	if f == JSON {
		vs, perr := parseStream(src)
		if perr != nil || len(vs) != 1 {
			return nil, &Failure{Code: 2, Stderr: jqSlurpStderr(srcPath, src, perr),
				Text: "the merge source is not one JSON value", Line: "the merge source of " + show + " is not one JSON value"}
		}
		srcJSON = string(vs[0].marshal())
	} else {
		if f == YAML {
			r, yerr := yq(nil, "ea", "-p", "yaml", "-o", "json", "[.] | length", srcPath)
			if yerr != nil {
				return nil, silent(2, r.stderr)
			}
			if strings.TrimRight(r.stdout, "\n") != "1" {
				return nil, &Failure{Code: 2, Stderr: r.stderr, Text: "the merge source has more than one YAML document",
					Line: "the merge source of " + show + " has more than one YAML document"}
			}
		}
		r, yerr := yq(nil, "eval", "-p", format, "-o", "json", ".", srcPath)
		if yerr != nil {
			return nil, silent(2, r.stderr)
		}
		srcJSON = r.stdout
	}
	kind, perr := typesOf(srcJSON)
	if perr != nil {
		return nil, silent(2, perr.Error()+"\n")
	}
	empty := false
	var srcV Value
	switch kind {
	case "null":
		srcV = objv(newObject())
		empty = true
	case "object":
		srcV, _ = oneValue([]byte(srcJSON))
	default:
		return nil, &Failure{Code: 2, Text: "the merge source is " + kindWord(kind) + ", not a table", Line: "the merge source of " + show + " is " + kindWord(kind) + ", not a table"}
	}

	// The live target: absent, or one table.
	var liveV Value
	var liveJSON string
	kind = "absent"
	if nonBlank(live) {
		livePath := file("live")
		if err := os.WriteFile(livePath, live, 0o600); err != nil {
			return nil, silent(1, err.Error())
		}
		switch f {
		case JSON:
			vs, perr := parseStream(live)
			if perr != nil || len(vs) != 1 {
				return nil, failed(show, "is not valid json - kept it as it is", jqSlurpStderr(livePath, live, perr))
			}
			liveJSON = string(vs[0].marshal())
		case YAML:
			r, yerr := yq(nil, "ea", "-p", "yaml", "-o", "json", "[.] | length", livePath)
			if yerr != nil {
				return nil, failed(show, "is not valid yaml - kept it as it is", r.stderr)
			}
			if strings.TrimRight(r.stdout, "\n") != "1" {
				return nil, failed(show, "has more than one YAML document - kept it as it is", r.stderr)
			}
			r, yerr = yq(nil, "eval", "-p", "yaml", "-o", "json", ".", livePath)
			if yerr != nil {
				return nil, silent(2, r.stderr)
			}
			liveJSON = r.stdout
		case TOML:
			r, yerr := yq(nil, "eval", "-p", "toml", "-o", "json", ".", livePath)
			if yerr != nil {
				return nil, failed(show, "is not valid toml - kept it as it is", r.stderr)
			}
			liveJSON = r.stdout
		}
		var perr *parseError
		kind, perr = typesOf(liveJSON)
		if perr != nil {
			return nil, silent(2, perr.Error()+"\n")
		}
		// A TOML file of comments only is an empty table; a YAML one is absent.
		switch {
		case kind == "null" && f == TOML:
			liveJSON = "{}\n"
			kind = "object"
		case kind == "null" && f == YAML && yamlHasNoContent(live):
			kind = "absent"
		}
		if kind != "object" && kind != "absent" {
			return nil, failed(show, "holds "+kindWord(kind)+", not a table - kept it as it is", "")
		}
		if kind == "object" {
			liveV, _ = oneValue([]byte(liveJSON))
		}
	}

	// An absent target gets the source without the `remove_items` items.
	if kind == "absent" {
		switch {
		case f == JSON:
			pl := mergePlan(objv(newObject()), srcV, Ops{RemoveItems: ops.RemoveItems})
			return pl.add.marshal(), nil
		case empty && f == YAML:
			return []byte("{}\n"), nil
		case empty:
			return []byte{}, nil
		}
		drop := sortPaths(itemPaths(srcV, ops.RemoveItems))
		for i, j := 0, len(drop)-1; i < j; i, j = i+1, j-1 {
			drop[i], drop[j] = drop[j], drop[i]
		}
		dropPath := file("src-drop.json")
		if err := os.WriteFile(dropPath, pathsValue(drop).marshal(), 0o600); err != nil {
			return nil, silent(1, err.Error())
		}
		r, yerr := yq([]string{"SE_MERGE_DROP=" + dropPath}, "eval", "-p", format, "-o", format,
			"load(strenv(SE_MERGE_DROP)) as $d | delpaths($d)", srcPath)
		if yerr != nil {
			return nil, silent(1, r.stderr)
		}
		return []byte(r.stdout), nil
	}

	pl := mergePlan(liveV, srcV, ops)
	if f == JSON {
		return appendMerge(delPaths(liveV, pl.del), pl.add).marshal(), nil
	}
	planPath := file("plan.json")
	if err := os.WriteFile(planPath, pl.value().marshal(), 0o600); err != nil {
		return nil, silent(1, err.Error())
	}
	r, yerr := yq([]string{"SE_MERGE_PLAN=" + planPath}, "eval", "-p", format, "-o", format,
		`load(strenv(SE_MERGE_PLAN)) as $m | delpaths($m.drop) *+ ($m.add | ((.. | (., key)) style = ""))`,
		file("live"))
	if yerr != nil {
		return nil, silent(1, r.stderr)
	}
	return []byte(r.stdout), nil
}

func pathsValue(ps []path) Value {
	items := make([]Value, len(ps))
	for i, p := range ps {
		items[i] = p.value()
	}
	return arr(items)
}
