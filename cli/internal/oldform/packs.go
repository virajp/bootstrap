package oldform

import (
	"embed"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// FormatKey is the top-level key of a values file of the new form, and
// Format its value. A values file without it is an old form.
const (
	FormatKey = "format"
	Format    = 2
)

// The frozen copy of the packs files of both sides, as they were before the
// values file had the key format: a member with `default = true` comes with
// its pack, and the pack base comes with each values file. Only the migration
// of a values file reads it. dprint.jsonc excludes the folder, so the copies
// stay byte for byte.
//
//go:embed packs-v2/repo.toml packs-v2/macos.toml
var frozen embed.FS

// basePack is the pack that each values file of the old form had.
const basePack = "base"

// frozenPacks reads the frozen packs file of <side>: each pack and its
// members, with the value of `default` of each member.
func frozenPacks(side source.Side) (map[string]map[string]bool, error) {
	b, err := frozen.ReadFile("packs-v2/" + string(side) + ".toml")
	if err != nil {
		return nil, err
	}
	var doc struct {
		Pack map[string]map[string]map[string]any `toml:"pack"`
	}
	if _, err := toml.Decode(string(b), &doc); err != nil {
		return nil, err
	}
	out := map[string]map[string]bool{}
	for name, members := range doc.Pack {
		out[name] = map[string]bool{}
		for key, m := range members {
			d, _ := m["default"].(bool)
			out[name][key] = d
		}
	}
	return out, nil
}

// frozenTools gives the tools that the pack <name> of the frozen packs
// <packs> added: its `default = true` members, and a nested `pack:<x>` with
// `default = true` expanded in turn. It is false for a pack that the frozen
// copy does not have.
func frozenTools(packs map[string]map[string]bool, name string) ([]string, bool) {
	if _, ok := packs[name]; !ok {
		return nil, false
	}
	var out []string
	seen := map[string]bool{}
	var walk func(p string)
	walk = func(p string) {
		if seen["pack:"+p] {
			return
		}
		seen["pack:"+p] = true
		keys := make([]string, 0, len(packs[p]))
		for k := range packs[p] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !packs[p][k] {
				continue
			}
			if rest, ok := strings.CutPrefix(k, "pack:"); ok {
				walk(rest)
			} else if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	walk(name)
	return out, true
}

// IsFormat tells if the values file <raw> has the key format with the value
// of the new form.
func IsFormat(raw *source.Obj) bool {
	v, ok := raw.Lookup(FormatKey)
	n, isInt := v.(int64)
	return ok && isInt && n == Format
}

// formatProblems gives the messages of the key format and of the
// `[packs.<name>]` tables of the values file <raw>, shown as <shown>.
func formatProblems(raw *source.Obj, shown, hint string) []*fail.Error {
	var probs []*fail.Error
	if p, ok := raw.Get("packs").(*source.Obj); ok {
		for _, name := range p.Keys() {
			probs = append(probs, usage("[packs.%s] in %s is the old form - %s", name, shown, hint))
		}
	} else if raw.Has("packs") {
		probs = append(probs, usage("packs in %s is the old form - %s", shown, hint))
	}
	v, ok := raw.Lookup(FormatKey)
	switch {
	case !ok:
		probs = append(probs, usage("%s without %s = %d is the old form - %s", shown, FormatKey, Format, hint))
	case !IsFormat(raw):
		probs = append(probs, usage("%s = %s in %s is not valid - the %s of a values file is %d", FormatKey, strings.TrimSpace(source.JSON(v)), shown, FormatKey, Format))
	}
	return probs
}

// source of the tools of one old pack: its name and its env.
type packSource struct {
	name   string
	env    []string
	hasEnv bool
}

// tablesOf is the part of a migration that moves the pack selection to
// `[tools.<name>]` tables. <raw> is the values file, shown as <shown>; <idx>
// is the checked index of the clone, <newPacks> its packs file, and <sub>
// tells if the repository is a git submodule. It gives the yq expressions and
// the notes. A file without the key format gets the tables of the frozen pack
// base too; each `[packs.<name>]` table gets the tables of the frozen pack.
// A tool gets no table when it has one, when the clone does not have it, when
// it has `selectable = false`, when it is `root_only` in a submodule, or when
// it is a member of the new pack base without `required = true`. Each
// dependency of such a tool gets a table too, with the same rules. A
// `[tools.<name>]` table of a tool that the clone does not have goes, and the
// note lists it.
func tablesOf(raw *source.Obj, shown string, side source.Side, idx *source.Index, newPacks *source.Packs, sub bool) ([]string, []string, error) {
	packs, err := frozenPacks(side)
	if err != nil {
		return nil, nil, fail.New(fail.Other, "cannot read the frozen packs: %v", err)
	}
	tools, _ := raw.Get("tools").(*source.Obj)
	has := func(t string) bool { return tools != nil && tools.Has(t) }
	_, hasFormat := raw.Lookup(FormatKey)
	var sources []packSource
	if !hasFormat {
		sources = append(sources, packSource{name: basePack})
	}
	old, _ := raw.Get("packs").(*source.Obj)
	if old != nil {
		for _, name := range old.Keys() {
			ps := packSource{name: name}
			if e, ok := old.Get(name).(*source.Obj); ok && e.Has("env") && side == source.Repo {
				ps.hasEnv = true
				for _, x := range strList(e.Get("env")) {
					ps.env = append(ps.env, x)
				}
			}
			sources = append(sources, ps)
		}
	}
	// The members of the new pack base without required = true.
	optional := map[string]bool{}
	for _, m := range newPacks.Members {
		if m.Pack == basePack && !newPacks.Required[m.Pack+"/"+m.Key] {
			optional[m.Key] = true
		}
	}
	skip := func(t string) bool {
		m := idx.Manifests[t]
		return m == nil || !m.Selectable || (sub && m.RootOnly) || optional[t] || has(t)
	}
	var order []string
	noEnv := map[string]bool{}
	envs := map[string][]string{}
	for _, ps := range sources {
		members, ok := frozenTools(packs, ps.name)
		if !ok {
			return nil, nil, fail.New(fail.Missing, "%s is not a bootstrap pack (named in %s)", ps.name, shown)
		}
		// The members that get a table, then each dependency of them.
		var start []string
		for _, t := range members {
			if m := idx.Manifests[t]; m != nil && !optional[t] && !(sub && m.RootOnly) {
				start = append(start, t)
			}
		}
		seen := map[string]bool{}
		for len(start) > 0 {
			t := start[0]
			start = start[1:]
			if seen[t] {
				continue
			}
			seen[t] = true
			m := idx.Manifests[t]
			if m == nil {
				continue
			}
			start = append(start, m.Depends...)
			if skip(t) {
				continue
			}
			if _, ok := envs[t]; !ok && !noEnv[t] {
				order = append(order, t)
				envs[t] = nil
			}
			if !ps.hasEnv {
				noEnv[t] = true
				continue
			}
			for _, e := range ps.env {
				if !contains(envs[t], e) {
					envs[t] = append(envs[t], e)
				}
			}
		}
	}
	var exprs, notes []string
	for _, t := range order {
		value := "{}"
		if !noEnv[t] {
			value = `{"env": ` + source.JSON(toAny(envs[t])) + `}`
		}
		exprs = append(exprs, ".tools["+q(t)+"] = ((.tools["+q(t)+"] // {}) * "+value+")")
	}
	if len(order) > 0 {
		notes = append(notes, "the tables of the packs: "+strings.Join(order, ", "))
	}
	if old != nil {
		exprs = append(exprs, "del(.packs)")
		names := old.Keys()
		for i, n := range names {
			names[i] = "[packs." + n + "]"
		}
		notes = append(notes, strings.Join(names, ", ")+" removed")
	}
	if tools != nil {
		var gone []string
		for _, t := range tools.Keys() {
			if idx.Manifests[t] == nil && t != OldRepoTool && t != OldMacTool {
				exprs = append(exprs, "del(.tools["+q(t)+"])")
				gone = append(gone, "[tools."+t+"]")
			}
		}
		if len(gone) > 0 {
			notes = append(notes, strings.Join(gone, ", ")+" removed: bootstrap no longer has the tool")
		}
	}
	if !hasFormat {
		exprs = append(exprs, "."+FormatKey+" = 2")
		notes = append(notes, FormatKey+" = 2")
	}
	return exprs, notes, nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func toAny(list []string) []any {
	out := make([]any, len(list))
	for i, s := range list {
		out[i] = s
	}
	return out
}

// needsTables tells if the values file <raw> needs the part of the migration
// that reads the clone.
func needsTables(raw *source.Obj) bool {
	_, hasFormat := raw.Lookup(FormatKey)
	return !hasFormat || raw.Has("packs")
}

// cloneTables gives the result of tablesOf with the clone of the command, in
// BOOTSTRAP_CLONE. Without a clone, a values file of the new form loses no
// table, and another file cannot migrate.
func cloneTables(raw *source.Obj, shown string, side source.Side, root string) ([]string, []string, error) {
	clone := os.Getenv("BOOTSTRAP_CLONE")
	if clone == "" {
		if needsTables(raw) {
			return nil, nil, usage("cannot migrate %s without a clone of bootstrap - BOOTSTRAP_CLONE is not set", shown)
		}
		return nil, nil, nil
	}
	return TablesIn(raw, shown, side, clone, side == source.Repo && inSubmodule(root))
}

// TablesIn is tablesOf with the clone <clone>.
func TablesIn(raw *source.Obj, shown string, side source.Side, clone string, sub bool) ([]string, []string, error) {
	src := source.Source{Root: clone, Side: side}
	idx, err := src.Open()
	if err != nil {
		return nil, nil, err
	}
	newPacks, err := src.ReadPacks()
	if err != nil {
		return nil, nil, err
	}
	return tablesOf(raw, shown, side, idx, newPacks, sub)
}

// inSubmodule tells if the repository <root> is a git submodule.
func inSubmodule(root string) bool {
	out, err := exec.Command("git", "-C", root, "rev-parse", "--show-superproject-working-tree").Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}
