package misegen

import (
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/tmpl"
)

// extend gives the tools.toml <t> of <tool> with the derived `mise` and
// `mise_env` tables merged in, and with the `depends` of each mise tool of
// [mise.tools] for the environment <env> (empty on the macos side).
func (g *Generator) extend(tool, env string, t any) any {
	m, ok := t.(map[string]any)
	if !ok || tool == "" {
		return t
	}
	if d := g.Derived[tool]; d != nil {
		for _, key := range []string{"mise", "mise_env"} {
			if !d.Has(key) {
				continue
			}
			v, err := tmpl.DecodeJSON([]byte(source.JSON(d.Get(key))))
			if err != nil {
				continue
			}
			m[key] = deepMerge(m[key], v)
		}
	}
	mise, ok := m["mise"].(map[string]any)
	if !ok {
		return m
	}
	tools, ok := mise["tools"].(map[string]any)
	if !ok {
		return m
	}
	for name, deps := range g.depends(tool, env) {
		switch v := tools[name].(type) {
		case map[string]any:
			v["depends"] = deps
		case nil:
		default:
			tools[name] = map[string]any{"version": v, "depends": deps}
		}
	}
	return m
}

// deepMerge merges <b> over <a>: tables merge key by key, any other value of
// <b> replaces.
func deepMerge(a, b any) any {
	am, ok1 := a.(map[string]any)
	bm, ok2 := b.(map[string]any)
	if !ok1 || !ok2 {
		return b
	}
	out := map[string]any{}
	for k, v := range am {
		out[k] = v
	}
	for k, v := range bm {
		out[k] = deepMerge(am[k], v)
	}
	return out
}

// present tells if the mise tool <name> of the selected tool is in the
// config of <env>: on the repo side, in the environment and with its
// requires_found met; on the macos side, always.
func (g *Generator) present(name, env string) bool {
	if g.Side == source.Macos {
		return true
	}
	return g.Envs != nil && g.Envs.mise.has(name+":"+env)
}

// depends gives the `depends` of each mise tool of <tool> in <env>: the mise
// tools of each tool of its `depends` and `parent`, then the mise tools above
// it in the same tool, each one only when it is in the config.
func (g *Generator) depends(tool, env string) map[string][]any {
	if g.Index == nil {
		return nil
	}
	m := g.Index.Manifests[tool]
	if m == nil {
		return nil
	}
	selected := map[string]bool{}
	for _, t := range g.Tools {
		selected[t] = true
	}
	var from []string
	seenTool := map[string]bool{}
	for _, d := range append(append([]string(nil), m.Depends...), m.Parent) {
		if d != "" && d != tool && !seenTool[d] && selected[d] {
			seenTool[d] = true
			from = append(from, d)
		}
	}
	var outside []string
	for _, d := range from {
		if dm := g.Index.Manifests[d]; dm != nil {
			for _, name := range dm.MiseOrder {
				if g.present(name, env) {
					outside = append(outside, name)
				}
			}
		}
	}
	out := map[string][]any{}
	var above []string
	for _, name := range m.MiseOrder {
		var deps []any
		seen := map[string]bool{name: true}
		for _, d := range append(append([]string(nil), outside...), above...) {
			if !seen[d] {
				seen[d] = true
				deps = append(deps, d)
			}
		}
		if len(deps) > 0 {
			out[name] = deps
		}
		if g.present(name, env) {
			above = append(above, name)
		}
	}
	return out
}
