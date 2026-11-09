package selection

import (
	"sort"

	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// ResolveMacos is `_ss_select`: the selection against `macos/`, packs expanded
// and depends included. <idx> is the checked index of the clone. The error is
// the first problem, with code 2 or 3.
func ResolveMacos(idx *source.Index, v *values.Macos) (*Resolved, error) {
	src := idx.Source
	if _, err := src.ReadPacks(); err != nil {
		return nil, err
	}
	sel, err := Select(src, BasePack, v.Sel)
	if err != nil {
		return nil, err
	}
	shown := src.ToolsShown()
	deps := func(from string) ([]string, error) {
		m := idx.Manifests[from]
		if m == nil {
			return nil, missing("tool '%s' is not in bootstrap", from)
		}
		for _, dep := range m.Depends {
			if !source.NameOK(dep) || idx.Manifests[dep] == nil {
				return nil, missing("%s/%s/tools.toml depends on the tool '%s', which bootstrap does not have", shown, from, dep)
			}
		}
		return append([]string(nil), m.Depends...), nil
	}
	w, err := Walk(entryTools(sel), deps, func(string) error { return nil })
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var tools []string
	for _, t := range w.Visited {
		if !seen[t] {
			seen[t] = true
			tools = append(tools, t)
		}
	}
	sort.Strings(tools)
	r := &Resolved{Index: idx, Sel: sel, Tools: tools, Deps: w.Deps, Keep: keepTools(sel, w.Deps), Init: map[string]bool{}}
	var queue []string
	for _, t := range tools {
		if idx.StageInit[t] {
			r.Init[t] = true
			r.InitOrder = append(r.InitOrder, t)
			queue = append(queue, t)
		}
	}
	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		for _, dep := range idx.Manifests[from].Depends {
			if !r.Init[dep] {
				r.Init[dep] = true
				r.InitOrder = append(r.InitOrder, dep)
				queue = append(queue, dep)
			}
		}
	}
	return r, nil
}
