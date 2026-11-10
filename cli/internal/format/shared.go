package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// This file holds the merge of a shared file: each item of a spec is found
// in the live document, then kept, written, added or removed. The edits go to
// a list of operations, which the writer of the format applies to the live
// text, so its comments stay. Contract: docs/design/cli.md ("Shared
// files").

// SharedInput is the input of MergeShared.
type SharedInput struct {
	Spec      []byte // The rendered spec.
	Live      []byte // The live file; nil when it is absent.
	Target    string // The target path, for the format of its extension.
	Explicit  string // The `format` key of the [files] entry, or "".
	Overrides []byte // JSON: the override table of the target, or nil.
	Records   []string
	Header    []string // The lines of the shared header, without a comment marker.
	Show      string   // The target in a message.
	SpecShow  string   // The spec in a message.
}

// SharedResult is the output of MergeShared.
type SharedResult struct {
	Out      []byte   // The new content of the target.
	Records  []string // The rows of the items that bootstrap added, as JSON.
	Warnings []string
}

// op is one edit of the live document: set <value> at <path>, or delete
// <path>. A path step is a string (a key) or an int (a list index).
type op struct {
	del   bool
	path  []any
	value Value
}

// doc is the live document and the edits made to it.
type doc struct {
	v   Value
	ops []op
}

// at gives the value at <p>.
func at(v Value, p []any) (Value, bool) {
	for _, s := range p {
		switch k := s.(type) {
		case string:
			if v.k != kObject || !v.has(k) {
				return Value{}, false
			}
			v = v.o.m[k]
		case int:
			if v.k != kArray || k < 0 || k >= len(v.a) {
				return Value{}, false
			}
			v = v.a[k]
		}
	}
	return v, true
}

// setIn gives a copy of <v> with <x> at <p>. Each step of <p> exists, except
// the last, which can be a new key or the index after the end of a list.
func setIn(v Value, p []any, x Value) Value {
	if len(p) == 0 {
		return x
	}
	switch k := p[0].(type) {
	case string:
		o := newObject()
		if v.k == kObject {
			o = v.o.clone()
		}
		o.set(k, setIn(o.m[k], p[1:], x))
		return objv(o)
	case int:
		a := append([]Value(nil), v.a...)
		if k == len(a) {
			a = append(a, setIn(Value{}, p[1:], x))
		} else {
			a[k] = setIn(a[k], p[1:], x)
		}
		return arr(a)
	}
	return v
}

// delIn gives a copy of <v> without <p>.
func delIn(v Value, p []any) Value {
	if len(p) == 0 {
		return v
	}
	switch k := p[0].(type) {
	case string:
		if v.k != kObject || !v.has(k) {
			return v
		}
		o := v.o.clone()
		if len(p) == 1 {
			o.del(k)
		} else {
			o.set(k, delIn(o.m[k], p[1:]))
		}
		return objv(o)
	case int:
		if v.k != kArray || k >= len(v.a) {
			return v
		}
		a := append([]Value(nil), v.a...)
		if len(p) == 1 {
			a = append(a[:k], a[k+1:]...)
		} else {
			a[k] = delIn(a[k], p[1:])
		}
		return arr(a)
	}
	return v
}

// set writes <x> at <p>. A missing step on the way, or a step that is not a
// table where a key follows, becomes a table: the edit sets that step with
// the rest of <x> nested below it.
func (d *doc) set(p []any, x Value) {
	cur := d.v
	for i := 0; i < len(p)-1; i++ {
		next, ok := at(cur, p[i:i+1])
		_, nextIsKey := p[i+1].(string)
		if !ok || (nextIsKey && next.k != kObject) {
			d.put(p[:i+1], nest(p[i+1:], x))
			return
		}
		cur = next
	}
	d.put(p, x)
}

// nest gives <x> below the key path <p>.
func nest(p []any, x Value) Value {
	for i := len(p) - 1; i >= 0; i-- {
		o := newObject()
		o.set(p[i].(string), x)
		x = objv(o)
	}
	return x
}

func (d *doc) put(p []any, x Value) {
	d.v = setIn(d.v, p, x)
	d.ops = append(d.ops, op{path: append([]any(nil), p...), value: x})
}

func (d *doc) del(p []any) {
	d.v = delIn(d.v, p)
	d.ops = append(d.ops, op{del: true, path: append([]any(nil), p...)})
}

func keys2path(p []string) []any {
	out := make([]any, len(p))
	for i, s := range p {
		out[i] = s
	}
	return out
}

func with(p []any, i int) []any { return append(append([]any(nil), p...), i) }

// sameType tells if two values have the same type: a number is one type.
func sameType(a, b Value) bool { return a.k == b.k }

// scalarText gives the text of a scalar, as an override key.
func scalarText(v Value) string {
	if v.k == kString {
		return v.s
	}
	return string(v.marshal()[:len(v.marshal())-1])
}

// identity gives the identity of an item: what finds it in a later run. The
// row of the manifest is the identity, and for a block in a table[] also
// `find`, the values of its match keys.
func (it *specItem) identity() *object {
	o := newObject()
	o.set("kind", str(it.kind))
	path := make([]Value, len(it.path))
	for i, s := range it.path {
		path[i] = str(s)
	}
	o.set("path", arr(path))
	switch {
	case it.kind == kindItem && it.match != "":
		o.set("match", str(it.match))
	case it.kind == kindItem:
		o.set("value", it.value)
	case it.kind == kindBlock && it.id != "":
		o.set("id", str(it.id))
	}
	return o
}

func identityKey(o *object) string { return compact(objv(o)) }

// compact gives <v> as JSON on one line.
func compact(v Value) string {
	var b bytes.Buffer
	if json.Compact(&b, v.marshal()) != nil {
		return ""
	}
	return b.String()
}

// finder gives the values of the match keys of a block value <v>. A key path
// through a list (`hooks.id`) gives the sorted set of the values.
func finder(v Value, keys []string) *object {
	o := newObject()
	for _, k := range keys {
		head, rest, through := strings.Cut(k, ".")
		if !through {
			if v.k == kObject && v.has(k) {
				o.set(k, v.o.m[k])
			}
			continue
		}
		l := v.get(head)
		if l.k != kArray {
			continue
		}
		var set []Value
		for _, e := range l.a {
			if x := e.get(rest); x.k != kNull && !anyEqual(set, x) {
				set = append(set, x)
			}
		}
		sort.SliceStable(set, func(i, j int) bool { return compare(set[i], set[j]) < 0 })
		o.set(k, arr(set))
	}
	return o
}

func finderEqual(a, b *object) bool { return equal(objv(a), objv(b)) }

// findItems gives the indexes of the items of <list> that <it> finds: an
// equal value, or a string that matches the regex.
func findItems(list Value, value Value, re *regexp.Regexp) []int {
	var out []int
	for i, e := range list.a {
		if re != nil {
			if e.k == kString && re.MatchString(e.s) {
				out = append(out, i)
			}
		} else if equal(e, value) {
			out = append(out, i)
		}
	}
	return out
}

// findBlocks gives the indexes of the tables of <list> whose match keys have
// the values of <find>.
func findBlocks(list Value, keys []string, find *object) []int {
	var out []int
	if len(find.keys) != len(keys) {
		return nil
	}
	for i, e := range list.a {
		if e.k == kObject && finderEqual(finder(e, keys), find) {
			out = append(out, i)
		}
	}
	return out
}

// delAll deletes the items <idx> of the list at <p>, from the last one.
func (d *doc) delAll(p []any, idx []int) {
	for i := len(idx) - 1; i >= 0; i-- {
		d.del(with(p, idx[i]))
	}
}

// withIgnored gives the block value <v> with the live value of each ignored
// key of <live>.
func withIgnored(v, live Value, ignore []string) Value {
	if v.k != kObject || live.k != kObject || len(ignore) == 0 {
		return v
	}
	o := v.o.clone()
	for _, k := range ignore {
		if live.has(k) {
			o.set(k, live.o.m[k])
		}
	}
	return objv(o)
}

// override applies the override table <ov> to the items, and gives the paths
// of <ov> that an item used.
func overrideItems(items []*specItem, ov Value) (map[string]bool, error) {
	used := map[string]bool{}
	for _, it := range items {
		p := append([]string(nil), it.path...)
		switch {
		case it.kind == kindItem && it.hasVal:
			p = append(p, scalarText(it.value))
		case it.kind == kindBlock && it.id != "":
			p = append(p, it.id)
		}
		node, ok := at(ov, keys2path(p))
		if !ok || len(p) == 0 {
			continue
		}
		used[strings.Join(p, "\x00")] = true
		dollar := false
		if node.k == kObject {
			for _, k := range node.o.keys {
				if strings.HasPrefix(k, "$") {
					dollar = true
				}
			}
		}
		if !dollar {
			it.value, it.hasVal = node, true
			continue
		}
		for _, k := range node.o.keys {
			x := node.o.m[k]
			switch {
			case k == "$skip" && x.k == kBool:
				it.skip = x.b
			case k == "$mode" && x.k == kString && x.s == "remove":
				it.remove = true
			default:
				return nil, fmt.Errorf("the override %s has %s - an override has an own value, or \"$skip\" = true, or \"$mode\" = \"remove\"", strings.Join(p, "."), k)
			}
		}
	}
	return used, nil
}

// unused gives the path of the first leaf of <ov> that no item used.
func unused(ov Value, path []string, used map[string]bool) string {
	if used[strings.Join(path, "\x00")] && len(path) > 0 {
		return ""
	}
	if ov.k == kObject && len(ov.o.keys) > 0 {
		dollar := false
		for _, k := range ov.o.keys {
			if strings.HasPrefix(k, "$") {
				dollar = true
			}
		}
		if !dollar {
			for _, k := range ov.o.keys {
				if u := unused(ov.o.m[k], append(append([]string(nil), path...), k), used); u != "" {
					return u
				}
			}
			return ""
		}
	}
	if len(path) == 0 {
		return ""
	}
	return strings.Join(path, ".")
}

// mergeItems merges the items into the document <d>. <records> holds the
// rows of the last run, by identity. It gives the rows of this run.
func mergeItems(d *doc, items []*specItem, records map[string]*object, warn func(string)) []string {
	var rows []string
	current := map[string]bool{}
	record := func(it *specItem, find *object) {
		o := it.identity()
		if find != nil {
			o.set("find", objv(find))
		}
		rows = append(rows, compact(objv(o)))
	}
	for _, it := range items {
		key := identityKey(it.identity())
		current[key] = true
		rec := records[key]
		p := keys2path(it.path)
		switch {
		case it.kind == kindKey || (it.kind == kindBlock && it.keys == nil):
			cur, found := at(d.v, p)
			if it.remove {
				if found {
					d.del(p)
				}
				continue
			}
			if found {
				switch {
				case it.kind == kindBlock:
					if nv := withIgnored(it.value, cur, it.ignore); !equal(cur, nv) {
						d.set(p, nv)
					}
				case !sameType(cur, it.value):
					warn(fmt.Sprintf("%s is %s, not %s - written again", strings.Join(it.path, "."), kindWord(cur.typeName()), kindWord(it.value.typeName())))
					d.set(p, it.value)
				}
				if rec != nil {
					record(it, nil)
				}
				continue
			}
			if it.skip {
				continue
			}
			d.set(p, it.value)
			record(it, nil)
		case it.kind == kindItem:
			list, found := at(d.v, p)
			if found && list.k != kArray {
				if it.remove || it.skip {
					continue
				}
				warn(fmt.Sprintf("%s is %s, not a list - written again", strings.Join(it.path, "."), kindWord(list.typeName())))
				d.set(p, arr([]Value{it.value}))
				record(it, nil)
				continue
			}
			idx := findItems(list, it.value, it.re)
			if it.remove {
				d.delAll(p, idx)
				continue
			}
			if len(idx) > 0 {
				if rec != nil {
					record(it, nil)
				}
				continue
			}
			if it.skip {
				continue
			}
			if !found {
				d.set(p, arr([]Value{it.value}))
			} else {
				d.set(with(p, len(list.a)), it.value)
			}
			record(it, nil)
		default: // A block in a table[].
			list, found := at(d.v, p)
			if found && list.k != kArray {
				if it.remove || it.skip {
					continue
				}
				warn(fmt.Sprintf("%s is %s, not a list - written again", strings.Join(it.path, "."), kindWord(list.typeName())))
				d.set(p, arr([]Value{it.value}))
				record(it, finder(it.value, it.keys))
				continue
			}
			var idx []int
			if it.hasVal {
				idx = findBlocks(list, it.keys, finder(it.value, it.keys))
			}
			if len(idx) == 0 && rec != nil {
				if f, ok := rec.m["find"]; ok && f.k == kObject {
					idx = findBlocks(list, it.keys, f.o)
				}
			}
			if it.remove {
				d.delAll(p, idx)
				continue
			}
			if len(idx) > 0 {
				live := list.a[idx[0]]
				if nv := withIgnored(it.value, live, it.ignore); !equal(live, nv) {
					d.set(with(p, idx[0]), nv)
				}
				if rec != nil {
					record(it, finder(it.value, it.keys))
				}
				continue
			}
			if it.skip {
				continue
			}
			if !found {
				d.set(p, arr([]Value{it.value}))
			} else {
				d.set(with(p, len(list.a)), it.value)
			}
			record(it, finder(it.value, it.keys))
		}
	}
	// The items that bootstrap added before, and that no spec has now.
	var keys []string
	for k := range records {
		if !current[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		cleanup(d, records[k])
	}
	return rows
}

// cleanup removes the item of the row <rec> from the document.
func cleanup(d *doc, rec *object) {
	var path []string
	if p := rec.m["path"]; p.k == kArray {
		path, _ = strings1(p)
	}
	p := keys2path(path)
	switch rec.m["kind"].s {
	case kindKey, kindBlock:
		if f, ok := rec.m["find"]; ok && f.k == kObject {
			list, _ := at(d.v, p)
			if list.k == kArray {
				d.delAll(p, findBlocks(list, f.o.keys, f.o))
			}
			return
		}
		if _, found := at(d.v, p); found && len(p) > 0 {
			d.del(p)
		}
	case kindItem:
		list, _ := at(d.v, p)
		if list.k != kArray {
			return
		}
		var re *regexp.Regexp
		if m, ok := rec.m["match"]; ok && m.k == kString {
			re, _ = regexp.Compile(m.s)
			if re == nil {
				return
			}
		}
		d.delAll(p, findItems(list, rec.m["value"], re))
	}
}

// parseRecords reads the rows of the manifest of a target.
func parseRecords(rows []string) map[string]*object {
	out := map[string]*object{}
	for _, r := range rows {
		v, ok := oneValue([]byte(r))
		if !ok || v.k != kObject || v.get("kind").k != kString {
			continue
		}
		id := v.o.clone()
		id.del("find")
		out[identityKey(id)] = v.o
	}
	return out
}

// pointer gives <p> as a JSON pointer.
func pointer(p []any) string {
	var b strings.Builder
	for _, s := range p {
		b.WriteByte('/')
		switch k := s.(type) {
		case string:
			b.WriteString(strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1"))
		case int:
			b.WriteString(strconv.Itoa(k))
		}
	}
	return b.String()
}

func anyEqual(list []Value, x Value) bool {
	for _, e := range list {
		if equal(e, x) {
			return true
		}
	}
	return false
}

// oneValue gives the one value of <data>, or false for none or many.
func oneValue(data []byte) (Value, bool) {
	vs, err := parseStream(data)
	if err != nil || len(vs) != 1 {
		return Value{}, false
	}
	return vs[0], true
}
