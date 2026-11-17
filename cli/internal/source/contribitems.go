package source

import (
	"sort"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// This file holds the contributed items of a merge-mode file:
// `[[contributes_items.<owner>."<target>".item]]` of a tools.toml adds items,
// in the item format of the merge spec of the owner, to the merge-mode file
// <target> of the tool <owner>. The values merge of [contributes.<tool>] does
// not read this key.

// contribItemsKey is the top-level key of the contributed items.
const contribItemsKey = "contributes_items"

// itemsOf gives the `item` list of the contributed items of <m> for <owner>
// and <target>, or nil.
func itemsOf(m *Manifest, owner, target string) []any {
	if m == nil || m.Raw == nil {
		return nil
	}
	byOwner, ok := m.Raw.Get(contribItemsKey).(*Obj)
	if !ok {
		return nil
	}
	byTarget, ok := byOwner.Get(owner).(*Obj)
	if !ok {
		return nil
	}
	t, ok := byTarget.Get(target).(*Obj)
	if !ok {
		return nil
	}
	l, _ := t.Get("item").([]any)
	return l
}

// ContributedItems gives the items that the tools <tools> contribute to the
// merge-mode file <target> of <owner>, in the order of the tool names, each
// tool's items in the order of its file. A tool that is not in the index, and
// <owner> itself, add nothing.
func (x *Index) ContributedItems(owner, target string, tools []string) []any {
	names := append([]string(nil), tools...)
	sort.Strings(names)
	var out []any
	seen := map[string]bool{}
	for _, t := range names {
		if t == owner || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, itemsOf(x.Manifests[t], owner, target)...)
	}
	return out
}

// KnownItems gives the items that any tool of the index contributes to the
// merge-mode file <target> of <owner>, selected or not (D17).
func (x *Index) KnownItems(owner, target string) []any {
	return x.ContributedItems(owner, target, x.Names)
}

// mergeTarget tells if <owner> declares <target> as a merge-mode file.
func (x *Index) mergeTarget(owner, target string) bool {
	m := x.Manifests[owner]
	if m == nil {
		return false
	}
	for _, f := range m.Files {
		if !f.Task && f.Target == target && f.Mode == "merge" {
			return true
		}
	}
	return false
}

// contribItemsRules examines the contributed items of <tool> (D7): their
// shape, an owner that the catalog has, and a target that the owner declares
// as a merge-mode file. Each problem is code 2.
func (x *Index) contribItemsRules(tool string) []*fail.Error {
	m := x.Manifests[tool]
	v, has := m.Raw.Lookup(contribItemsKey)
	if !has {
		return nil
	}
	at := x.Source.ManifestShown(tool)
	var out []*fail.Error
	bad := func(format string, args ...any) {
		out = append(out, fail.New(fail.Usage, "%s: "+format, append([]any{at}, args...)...))
	}
	byOwner, ok := v.(*Obj)
	if !ok {
		bad("[contributes_items] must be a table of [contributes_items.<tool>] tables")
		return out
	}
	for _, owner := range byOwner.Keys() {
		byTarget, ok := byOwner.Get(owner).(*Obj)
		if !ok {
			bad("[contributes_items.%s] must be a table of [contributes_items.%s.\"<target>\"] tables", owner, owner)
			continue
		}
		if !x.Has(owner) {
			bad("[contributes_items.%s] names a tool that bootstrap does not have", owner)
			continue
		}
		for _, target := range byTarget.Keys() {
			t, ok := byTarget.Get(target).(*Obj)
			if !ok || t.Len() != 1 || !t.Has("item") {
				bad("[contributes_items.%s.\"%s\"] must hold only [[contributes_items.%s.\"%s\".item]] tables", owner, target, owner, target)
				continue
			}
			l, ok := t.Get("item").([]any)
			okItems := ok
			for _, e := range l {
				if _, isObj := e.(*Obj); !isObj {
					okItems = false
				}
			}
			if !okItems {
				bad("[contributes_items.%s.\"%s\"] item must be a list of [[contributes_items.%s.\"%s\".item]] tables", owner, target, owner, target)
				continue
			}
			if !x.mergeTarget(owner, target) {
				bad("[contributes_items.%s.\"%s\"]: %s declares no merge-mode file %s - name a [files.\"<target>\"] of %s with mode = \"merge\"", owner, target, owner, target, owner)
			}
		}
	}
	return out
}
