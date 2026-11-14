// Package values reads the values files (`.config/bootstrap.toml` of a repo and
// `macos.toml` of the Mac) as far as the engine reads them, and merges the
// values of the selected tools: defaults, then contributes, then the values
// file. Contract: docs/design/cli.md ("The values of a tool").
package values

import (
	"os"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// Item is one `[tools.<name>]` or `[packs.<name>]` entry of a values file.
type Item struct {
	Kind string // "tools" or "packs".
	Name string
	// Envs is the `env` list joined with commas, or "-" when the entry has no
	// `env`. The Mac side always has "-".
	Envs string
	// Env is the `env` list. It is nil when the entry has no `env`.
	Env []string
}

// Repo is the values file of a repo, `.config/bootstrap.toml`.
type Repo struct {
	Path string
	Raw  *source.Obj // The whole file, with sorted keys.
	// Sel holds the `[tools.*]` entries and then the `[packs.*]` entries, each
	// group in name order.
	Sel []Item
	// User holds each `[tools.<name>]` table without `env`.
	User *source.Obj
}

func usage(format string, args ...any) *fail.Error { return fail.New(fail.Usage, format, args...) }

// ReadRepo reads a repo values file and checks the shape of its `[tools]` and
// `[packs]` tables, as `_rs_values_sel` does. <shown> is the path as a message
// names it. The error is the first problem, with code 2. The `env` names are
// not checked here: they need the environments of the repo.
func ReadRepo(path, shown string) (*Repo, error) {
	raw, err := source.ReadTOML(path)
	if err != nil {
		return nil, usage("cannot read %s", shown)
	}
	v := &Repo{Path: path, Raw: raw, User: source.NewObj()}
	for _, kind := range []string{"tools", "packs"} {
		if !raw.Has(kind) {
			continue
		}
		table, ok := raw.Get(kind).(*source.Obj)
		if !ok {
			name := "security"
			if kind == "tools" {
				name = "node"
			}
			return nil, usage("%s in %s is a list, the old format - write one table for each entry, for example [%s.%s] env = [\"dev\"]; the pack base is always added", kind, shown, kind, name)
		}
		for _, name := range table.Keys() {
			e, ok := table.Get(name).(*source.Obj)
			if !ok {
				return nil, usage("[%s.%s] in %s: must be a table", kind, name, shown)
			}
			if kind == "packs" {
				for _, ek := range e.Keys() {
					if ek != "env" {
						return nil, usage("[%s.%s] in %s: '%s' is not a key of a pack entry; the only key of a pack entry is env, a list", kind, name, shown, ek)
					}
				}
			}
			it := Item{Kind: kind, Name: name, Envs: "-"}
			if e.Has("env") {
				l, ok := e.Get("env").([]any)
				if !ok {
					return nil, usage("[%s.%s] in %s: env must be a list, for example env = [\"dev\"]", kind, name, shown)
				}
				it.Env = make([]string, len(l))
				for i, x := range l {
					it.Env[i] = stringOf(x)
				}
				it.Envs = strings.Join(it.Env, ",")
			}
			v.Sel = append(v.Sel, it)
			if kind == "tools" {
				u := e.Clone()
				omit(u, "env")
				v.User.Set(name, u)
			}
		}
	}
	return v, nil
}

func omit(o *source.Obj, key string) {
	if !o.Has("env") {
		return
	}
	c := source.NewObj()
	for _, k := range o.Keys() {
		if k != key {
			c.Set(k, o.Get(k))
		}
	}
	*o = *c
}

func stringOf(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return strings.TrimSpace(source.JSON(v))
}

// Macos is the values file of the Mac, `macos.toml`.
type Macos struct {
	// Exists is false when the file is not there: the selection is empty.
	Exists bool
	Path   string
	// Sel holds the `[packs.*]` entries and then the `[tools.*]` entries, each
	// group in name order. Every Envs is "-".
	Sel []Item
	// User holds the `[tools.<name>]` tables.
	User *source.Obj
}

const macosKeys = "bootstrap_commit, format, [tools.<name>], [overrides.mise.<path>] and [mise.*]"

// ReadMacos reads the Mac values file <path> if it exists, as `_ss_read_values`
// does. <shown> is the path as a message names it. The error holds each message
// of bash in order; its code is 3 when a name is not in bootstrap, else 2.
func ReadMacos(path, shown string) (*Macos, error) {
	v := &Macos{Path: path, User: source.NewObj()}
	st, err := os.Stat(path)
	if err != nil {
		return v, nil
	}
	if !st.Mode().IsRegular() {
		return v, usage("%s is not a file", shown)
	}
	raw, err := source.ReadTOML(path)
	if err != nil {
		return v, usage("cannot read %s", shown)
	}
	v.Exists = true
	var probs []*fail.Error
	bad := func(format string, args ...any) { probs = append(probs, usage(format, args...)) }
	named := func(kind, name string) {
		if !source.NameOK(name) {
			probs = append(probs, fail.New(fail.Missing, "%s is not a bootstrap %s (named in %s)", name, kind, shown))
			return
		}
		v.Sel = append(v.Sel, Item{Kind: kind + "s", Name: name, Envs: "-"})
	}
	typeBad := func(name string) {
		switch name {
		case "packs":
			bad("packs in %s must be [packs.<name>] tables - the top-level keys are %s", shown, macosKeys)
		case "bootstrap_commit":
			bad("bootstrap_commit in %s must be a string", shown)
		default:
			bad("%s in %s must be a table - the top-level keys are %s", name, shown, macosKeys)
		}
	}
	for _, k := range raw.Keys() {
		x := raw.Get(k)
		switch k {
		case "bootstrap_commit":
			if _, ok := x.(string); !ok {
				typeBad(k)
			}
		case "format":
			// The package oldform examines its value.
		case "tools":
			t, ok := x.(*source.Obj)
			if !ok {
				bad("tools = [...] is the old format - write a [tools.<name>] table for each tool")
				break
			}
			for _, n := range t.Keys() {
				e, ok := t.Get(n).(*source.Obj)
				if !ok {
					bad("[tools.%s] in %s must be a table", n, shown)
					continue
				}
				named("tool", n)
				if e.Has("env") {
					bad("[tools.%s] env is for repo tools only - the Mac side has no environments", n)
				}
			}
		case "packs":
			t, ok := x.(*source.Obj)
			if !ok {
				typeBad(k)
				break
			}
			for _, n := range t.Keys() {
				e, ok := t.Get(n).(*source.Obj)
				if !ok {
					bad("[packs.%s] in %s must be an empty table", n, shown)
					continue
				}
				named("pack", n)
				for _, ek := range e.Keys() {
					bad("[packs.%s] in %s has the key '%s' - a pack entry is an empty table", n, shown, ek)
				}
			}
		case "overrides":
			t, ok := x.(*source.Obj)
			if !ok {
				typeBad(k)
				break
			}
			for _, ok := range t.Keys() {
				if ok == "mise_env" {
					bad("[overrides.mise_env] in %s: the Mac side has no environments - use [overrides.mise.<path>]", shown)
				} else if ok != "mise" {
					bad("[overrides.%s] in %s is not valid - write [tools.<name>] for template values and [overrides.mise.<path>] for mise config", ok, shown)
				}
			}
		case "mise":
			if _, ok := x.(*source.Obj); !ok {
				typeBad(k)
			}
		default:
			bad("%s is not a key of %s - the top-level keys are %s", k, shown, macosKeys)
		}
	}
	if len(probs) > 0 {
		return v, source.Problems(probs)
	}
	if t, ok := raw.Get("tools").(*source.Obj); ok {
		v.User = t
	}
	return v, nil
}

// Names gives the names of <items> of <kind>, in the order of the items.
func Names(items []Item, kind string) []string {
	var out []string
	for _, it := range items {
		if it.Kind == kind {
			out = append(out, it.Name)
		}
	}
	return out
}

// OverridesPath gives the overrides file of the values file <path>: the
// same folder and name, with `.overrides.toml` in place of `.toml`, for
// example `.config/bootstrap.overrides.toml`.
func OverridesPath(path string) string {
	return strings.TrimSuffix(path, ".toml") + ".overrides.toml"
}

// ReadOverrides reads the overrides file <path>, when it exists, and gives
// its `[tools.<name>]` tables. It wins over all the other values. Its only
// key is `tools`. <shown> is the path as a message names it.
func ReadOverrides(path, shown string) (*source.Obj, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, nil
	}
	raw, err := source.ReadTOML(path)
	if err != nil {
		return nil, usage("cannot read %s", shown)
	}
	for _, k := range raw.Keys() {
		if k != "tools" {
			return nil, usage("%s is not a key of %s - its only key is [tools.<name>]", k, shown)
		}
	}
	t, ok := raw.Get("tools").(*source.Obj)
	if !ok {
		if raw.Has("tools") {
			return nil, usage("tools in %s must be [tools.<name>] tables", shown)
		}
		return nil, nil
	}
	for _, n := range t.Keys() {
		if _, ok := t.Get(n).(*source.Obj); !ok {
			return nil, usage("[tools.%s] in %s must be a table", n, shown)
		}
	}
	return t, nil
}
