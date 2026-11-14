// Package selection expands the packs, checks them, and walks the `depends` of
// the selected tools depth first, as `select` of the bash engine does. It also
// resolves a whole selection for a side: the repo side and the Mac side.
// Contract: docs/design/packs.md.
package selection

import (
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// Entry is one selected tool: `<tool> <env,...|-> <source>`. Src is `default`
// for a required member of the pack base, and `tools.<name>` for a tool that
// the values file names. A tool can be in the list more than once.
type Entry struct {
	Tool, Envs, Src string
}

// Line gives the entry as the bash engine prints it.
func (e Entry) Line() string { return e.Tool + " " + e.Envs + " " + e.Src }

func usage(format string, args ...any) *fail.Error { return fail.New(fail.Usage, format, args...) }
func missing(format string, args ...any) *fail.Error {
	return fail.New(fail.Missing, format, args...)
}

// Select is `_se_selected`: the required members of the pack <base>, then
// each `tools` item of <sel>. The values file keeps no record of a pack, so a
// `packs` item adds nothing. It examines every pack of the packs file first.
// The error is one message, with code 2 or 3.
func Select(src source.Source, base string, sel []values.Item) ([]Entry, error) {
	packs, err := src.ReadPacks()
	if err != nil {
		return nil, err
	}
	if err := CheckPacks(src, packs); err != nil {
		return nil, err
	}
	var out []Entry
	for _, t := range Required(packs, base) {
		out = append(out, Entry{t, "-", "default"})
	}
	for _, it := range sel {
		if it.Kind == "tools" {
			out = append(out, Entry{it.Name, it.Envs, "tools." + it.Name})
		}
	}
	return out, nil
}

// Required gives the members of the pack <base> with `required = true`, in
// file order.
func Required(p *source.Packs, base string) []string {
	var out []string
	for _, o := range p.Order {
		if o.Pack == base && p.Required[base+"/"+o.Key] {
			out = append(out, o.Key)
		}
	}
	return out
}

// Expand is `_se_pack`: the tools of the pack <name> as entries: each member
// in file order, a `pack:<x>` member expanded in turn. Code 3 for an unknown
// pack.
func Expand(p *source.Packs, name, envs, srcName string) ([]Entry, error) {
	if !p.HasPack(name) {
		return nil, missing("%s is not a bootstrap pack", name)
	}
	var out []Entry
	for _, o := range p.Order {
		if o.Pack != name || !p.HasMember(name, o.Key) {
			continue
		}
		if rest, ok := strings.CutPrefix(o.Key, "pack:"); ok {
			more, err := Expand(p, rest, envs, srcName)
			if err != nil {
				return nil, err
			}
			out = append(out, more...)
			continue
		}
		out = append(out, Entry{o.Key, envs, srcName})
	}
	return out, nil
}

// PackLoop is `_se_pack_loop`: stop on a loop through the pack <name> over each
// `pack:<x>` member. <chain> holds the packs that include this one. Code 2.
func PackLoop(p *source.Packs, name string, chain []string, shown string) error {
	for i, c := range chain {
		if c == name {
			loop := name
			for _, w := range chain[i+1:] {
				loop += " -> " + w
			}
			return usage("%s: a pack loop: %s -> %s", shown, loop, name)
		}
	}
	chain = append(chain[:len(chain):len(chain)], name)
	for _, m := range p.Members {
		if m.Pack != name {
			continue
		}
		if rest, ok := strings.CutPrefix(m.Key, "pack:"); ok {
			if err := PackLoop(p, rest, chain, shown); err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckPacks is `_se_pack_check`: it examines every `[pack.*]` table, expanded
// or not. The error is the first problem, with code 2 or 3.
func CheckPacks(src source.Source, p *source.Packs) error {
	shown := p.Shown
	if len(p.BadDescriptions) > 0 {
		return usage("%s: [pack.%s] description must be a string", shown, p.BadDescriptions[0])
	}
	for _, b := range p.Bad {
		switch b.Key + ":" + b.Kind {
		case "depends:slice", "required:slice", "tools:slice", "default:slice":
			return usage("%s: [pack.%s] %s is the old format - write each member as <tool> = {}", shown, b.Pack, b.Key)
		}
		return usage("%s: [pack.%s] %s must be a table, {} or { required = true }", shown, b.Pack, b.Key)
	}
	for _, m := range p.Members {
		if rest, ok := strings.CutPrefix(m.Key, "p:"); ok {
			return usage("%s: [pack.%s] \"%s\" is the old form - write \"pack:%s\" = {}", shown, m.Pack, m.Key, rest)
		}
	}
	for _, m := range p.Members {
		if v, ok := strings.CutPrefix(m.State, "bad:key "); ok {
			if v == "default" {
				return usage("%s: [pack.%s] %s has the entry key 'default', the old form - each member is a default: write %s = {}", shown, m.Pack, m.Key, m.Key)
			}
			return usage("%s: [pack.%s] %s has the entry key '%s' - a member is {}, or { required = true } in pack %s", shown, m.Pack, m.Key, v, BasePack)
		}
		if v, ok := strings.CutPrefix(m.State, "bad:required "); ok {
			return usage("%s: [pack.%s] %s: required must be true or false, not %s", shown, m.Pack, m.Key, v)
		}
		if p.Required[m.Pack+"/"+m.Key] && (m.Pack != BasePack || strings.HasPrefix(m.Key, "pack:")) {
			return usage("%s: [pack.%s] %s: required = true is only for a tool of pack %s", shown, m.Pack, m.Key, BasePack)
		}
	}
	for _, m := range p.Members {
		if !p.HasOrder(m.Pack, m.Key) {
			return usage("%s: [pack.%s] %s must be on one line", shown, m.Pack, m.Key)
		}
	}
	for _, name := range p.Names {
		if src.IsDir(name) {
			return usage("%s: pack '%s' is also a tool folder - rename one", shown, name)
		}
	}
	for _, m := range p.Members {
		if rest, ok := strings.CutPrefix(m.Key, "pack:"); ok {
			if !p.HasPack(rest) {
				return missing("%s is not a bootstrap pack (named in pack %s)", rest, m.Pack)
			}
		} else if !src.IsDir(m.Key) {
			return missing("%s is not a bootstrap tool (named in pack %s)", m.Key, m.Pack)
		}
	}
	for _, name := range p.Names {
		if err := PackLoop(p, name, nil, shown); err != nil {
			return err
		}
	}
	return nil
}

// DepRow is a tool and its dependencies, as a line of `_SE_WALK_DEPS`.
type DepRow struct {
	Tool string
	Deps []string
}

// Walked is the result of Walk.
type Walked struct {
	Visited []string // The finished tools, in visit order: a dependency before its tool.
	Deps    []DepRow // The tools that have dependencies, in visit order.
}

// Walk is `_se_walk`: it walks <start> depth first, so each dependency
// finishes before the tool that named it. A loop of `depends` is not an error:
// the walk finishes the tool that it met first. <deps> gives the dependency
// names of a tool, or an error that stops the walk. <done> runs when the walk
// finishes a tool.
func Walk(start []string, deps func(tool string) ([]string, error), done func(tool string) error) (*Walked, error) {
	w := &Walked{}
	visited := map[string]bool{}
	visiting := map[string]bool{}
	pending := append([]string(nil), start...)
	for len(pending) > 0 {
		tool := pending[0]
		pending = pending[1:]
		if tool == "" || visited[tool] {
			continue
		}
		all, err := deps(tool)
		if err != nil {
			return w, err
		}
		var open []string
		for _, d := range all {
			if !visited[d] {
				open = append(open, d)
			}
		}
		if len(open) > 0 && !visiting[tool] {
			visiting[tool] = true
			pending = append(append(open, tool), pending...)
			continue
		}
		if len(all) > 0 {
			w.Deps = append(w.Deps, DepRow{tool, all})
		}
		if err := done(tool); err != nil {
			return w, err
		}
		visited[tool] = true
		w.Visited = append(w.Visited, tool)
	}
	return w, nil
}

func entryTools(sel []Entry) []string {
	out := make([]string, len(sel))
	for i, e := range sel {
		out[i] = e.Tool
	}
	return out
}

// keepTools gives the required base tools (source `default`) and each
// dependency.
func keepTools(sel []Entry, deps []DepRow) []string {
	var out []string
	for _, e := range sel {
		if e.Src == "default" {
			out = append(out, e.Tool)
		}
	}
	for _, r := range deps {
		out = append(out, r.Deps...)
	}
	return out
}

// MissingTables gives the required members of the pack <base> that have no
// `[tools.<name>]` table in <named>, in file order: the tables that a sync
// writes into the values file. It skips a member whose hosts leave out this
// host, a member with `selectable = false`, and, when <sub> is true (a git
// submodule), a `root_only` member.
func MissingTables(idx *source.Index, base string, named map[string]bool, sub bool) ([]string, error) {
	packs, err := idx.Source.ReadPacks()
	if err != nil {
		return nil, err
	}
	host := source.HostName()
	var out []string
	for _, t := range Required(packs, base) {
		m := idx.Manifests[t]
		if named[t] || m == nil || !m.Selectable || !m.RunsOn(host) || (sub && m.RootOnly) {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}
