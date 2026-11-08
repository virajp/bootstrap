package format

import "sort"

// This file is the Go form of _SE_MERGE_JQ in sync-engine/merge. Each function
// carries the name of its jq def.

// step is one step of a path: a key of a table or an index of a list.
type step struct {
	idx   int
	key   string
	isIdx bool
}

type path []step

func keyPath(keys []string) path {
	p := make(path, len(keys))
	for i, k := range keys {
		p[i] = step{key: k}
	}
	return p
}

func (p path) with(s step) path {
	c := make(path, len(p)+1)
	copy(c, p)
	c[len(p)] = s
	return c
}

func (p path) value() Value {
	items := make([]Value, len(p))
	for i, s := range p {
		if s.isIdx {
			items[i] = intVal(s.idx)
		} else {
			items[i] = str(s.key)
		}
	}
	return arr(items)
}

// cmpPath is the jq order of two paths: a number sorts before a string.
func cmpPath(a, b path) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		x, y := a[i], b[i]
		switch {
		case x.isIdx && !y.isIdx:
			return -1
		case !x.isIdx && y.isIdx:
			return 1
		case x.isIdx:
			if c := cmpInt(x.idx, y.idx); c != 0 {
				return c
			}
		default:
			if x.key != y.key {
				if x.key < y.key {
					return -1
				}
				return 1
			}
		}
	}
	return cmpInt(len(a), len(b))
}

func sortPaths(ps []path) []path {
	c := append([]path(nil), ps...)
	sort.SliceStable(c, func(i, j int) bool { return cmpPath(c[i], c[j]) < 0 })
	return c
}

func uniquePaths(ps []path) []path {
	var out []path
	for _, p := range sortPaths(ps) {
		if len(out) == 0 || cmpPath(out[len(out)-1], p) != 0 {
			out = append(out, p)
		}
	}
	return out
}

// outer is `outer`: the paths that are not under another one, the last first.
func outer(ps []path) []path {
	all := uniquePaths(ps)
	var out []path
	for _, p := range all {
		under := false
		for _, q := range all {
			if len(q) < len(p) && cmpPath(q, p[:len(q)]) == 0 {
				under = true
				break
			}
		}
		if !under {
			out = append(out, p)
		}
	}
	// `sort | reverse`
	out = sortPaths(out)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// hasPath is `has_path`: each step is a table that has the key.
func hasPath(v Value, p []string) bool {
	for _, k := range p {
		if v.k != kObject || !v.has(k) {
			return false
		}
		v = v.o.m[k]
	}
	return true
}

func getPath(v Value, p []string) Value {
	for _, k := range p {
		v = v.get(k)
	}
	return v
}

// delPaths is the jq `delpaths`: the paths are unique and sorted, and the
// last is deleted first. A path that is not there is skipped.
func delPaths(v Value, ps []path) Value {
	sorted := uniquePaths(ps)
	for i := len(sorted) - 1; i >= 0; i-- {
		v = delPath(v, sorted[i])
	}
	return v
}

func delPath(v Value, p path) Value {
	if len(p) == 0 {
		return v
	}
	s := p[0]
	switch {
	case v.k == kObject && !s.isIdx:
		child, ok := v.o.m[s.key]
		if !ok {
			return v
		}
		o := v.o.clone()
		if len(p) == 1 {
			o.del(s.key)
		} else {
			o.m[s.key] = delPath(child, p[1:])
		}
		return objv(o)
	case v.k == kArray && s.isIdx:
		if s.idx < 0 || s.idx >= len(v.a) {
			return v
		}
		a := append([]Value(nil), v.a...)
		if len(p) == 1 {
			a = append(a[:s.idx], a[s.idx+1:]...)
		} else {
			a[s.idx] = delPath(a[s.idx], p[1:])
		}
		return arr(a)
	}
	return v
}

// RemoveRule is one `remove_items` table: the items to delete from the list
// at Path.
type RemoveRule struct {
	Path  []string
	Items []Value
}

// itemPaths is `item_paths`: the paths in <v> of each list item equal to an
// item of a rule.
func itemPaths(v Value, rules []RemoveRule) []path {
	var out []path
	for _, r := range rules {
		if !hasPath(v, r.Path) {
			continue
		}
		list := getPath(v, r.Path)
		if list.k != kArray {
			continue
		}
		for i, e := range list.a {
			for _, it := range r.Items {
				if equal(it, e) {
					out = append(out, keyPath(r.Path).with(step{idx: i, isIdx: true}))
					break
				}
			}
		}
	}
	return out
}

// newItems is `new_items`: the items of <s> that are not equal to an item of
// <l> or to an earlier item.
func newItems(l, s []Value) []Value {
	out := []Value{}
	for _, x := range s {
		if anyEqual(l, x) || anyEqual(out, x) {
			continue
		}
		out = append(out, x)
	}
	return out
}

func anyEqual(list []Value, x Value) bool {
	for _, e := range list {
		if equal(e, x) {
			return true
		}
	}
	return false
}

// additions is `additions`.
func additions(l, s Value) Value {
	switch s.k {
	case kObject:
		o := newObject()
		for _, k := range s.o.keys {
			o.set(k, additions(l.get(k), s.o.m[k]))
		}
		return objv(o)
	case kArray:
		var ll []Value
		if l.k == kArray {
			ll = l.a
		}
		return arr(newItems(ll, s.a))
	}
	return s
}

// conflicts is `conflicts`: the paths under <p> of <l> (a live value) whose
// kind differs from <s>.
func conflicts(l, s Value, p path) []path {
	if l.k == kObject && s.k == kObject {
		var out []path
		for _, k := range s.o.keys {
			if lv, ok := l.o.m[k]; ok {
				out = append(out, conflicts(lv, s.o.m[k], p.with(step{key: k}))...)
			}
		}
		return out
	}
	if l.kindOf() != s.kindOf() {
		return []path{p}
	}
	return nil
}

// appendMerge is `append_merge`.
func appendMerge(a, b Value) Value {
	switch {
	case a.k == kObject && b.k == kObject:
		o := a.o.clone()
		for _, k := range b.o.keys {
			o.set(k, appendMerge(a.get(k), b.o.m[k]))
		}
		return objv(o)
	case a.k == kArray && b.k == kArray:
		out := append(append([]Value(nil), a.a...), b.a...)
		return arr(out)
	}
	return b
}

// Ops are the removals of a merge: the `remove` key paths and the
// `remove_items` rules.
type Ops struct {
	Remove      [][]string
	RemoveItems []RemoveRule
}

type plan struct {
	del, replace, drop []path
	add                Value
}

// mergePlan is `merge_plan`.
func mergePlan(live, src Value, ops Ops) plan {
	var removes []path
	for _, r := range ops.Remove {
		if hasPath(live, r) {
			removes = append(removes, keyPath(r))
		}
	}
	items := itemPaths(live, ops.RemoveItems)
	src = delPaths(src, itemPaths(src, ops.RemoveItems))
	rm := append(append([]path(nil), removes...), items...)
	kept := delPaths(live, rm)
	cf := conflicts(kept, src, nil)
	return plan{
		del:     outer(rm),
		replace: outer(cf),
		drop:    outer(append(append([]path(nil), rm...), cf...)),
		add:     additions(kept, src),
	}
}

// value gives the plan as the table {del, replace, drop, add}.
func (pl plan) value() Value {
	list := func(ps []path) Value {
		items := make([]Value, len(ps))
		for i, p := range ps {
			items[i] = p.value()
		}
		return arr(items)
	}
	o := newObject()
	o.set("del", list(pl.del))
	o.set("replace", list(pl.replace))
	o.set("drop", list(pl.drop))
	o.set("add", pl.add)
	return objv(o)
}
