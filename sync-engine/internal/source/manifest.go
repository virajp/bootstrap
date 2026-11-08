package source

import (
	"fmt"
	"strings"
)

// FileEntry is one `[files."<target>"]` table of a tools.toml. A key that the
// table does not have is nil. Values are as the TOML gives them.
type FileEntry struct {
	Target        string
	Mode          string // "generated", "seed" or "merge"; empty when unset or not valid.
	Format        string // "text", "json", "yaml" or "toml"; empty when unset or not valid.
	Requires      []string
	RequiresFound []string
	Remove        any // The value of `remove`, or nil.
	RemoveItems   any // The value of `remove_items`, or nil.
	Keys          *Obj
}

// MiseTool is one `[mise.tools.<name>]` entry (repo side).
type MiseTool struct {
	Name          string
	Envs          []string // nil when the entry has no envs.
	Depends       []string
	RequiresFound []string
}

// Manifest is the tools.toml of one tool.
type Manifest struct {
	Tool        string
	Raw         *Obj // The whole file, with sorted keys.
	Description string
	Stage       string   // Mac side: "init", or empty.
	Depends     []string // The tool names of `depends`.
	DefaultEnvs []string // Repo side.
	Extensions  []string // Repo side.
	RootOnly    bool     // Repo side: `root_only = true`.
	Files       []FileEntry
	Mise        []MiseTool // Repo side.

	// lines is the flat form that the rules read (repo side).
	lines []mline
}

// mline is one line of the flat manifest form of `_RS_MANIFEST_TEMPLATE`.
type mline struct {
	field, a, b, c string
	rest           []string // mise: the mise depends.
}

func asList(v any) ([]any, bool) {
	l, ok := v.([]any)
	return l, ok
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return strings.TrimSpace(strings.ReplaceAll(fmt.Sprint(v), "\n", " "))
}

func strList(l []any) []string {
	out := make([]string, len(l))
	for i, e := range l {
		out[i] = str(e)
	}
	return out
}

// listOf reads a key that must be a list or absent (the template ranges over
// it). It gives false for another type.
func listOf(o *Obj, key string) ([]string, bool) {
	v, ok := o.Lookup(key)
	if !ok || v == nil {
		return nil, true
	}
	l, ok := asList(v)
	if !ok {
		return nil, false
	}
	return strList(l), true
}

// LoadManifest reads the tools.toml of <tool> and types it for the side of the
// source. It returns the message `cannot read <side>/tools/<tool>/tools.toml`
// (code 2) when the file does not parse or a key has a type that the sync
// cannot read. The rules of the shape are in the checks (CheckRepoIndex and
// CheckMacosIndex), not here.
func (s Source) LoadManifest(tool string) (*Manifest, error) {
	bad := readErr("cannot read %s", s.ManifestShown(tool))
	raw, err := readTOML(s.ToolDir(tool) + "/tools.toml")
	if err != nil {
		return nil, bad
	}
	m := &Manifest{Tool: tool, Raw: raw}
	if v, ok := raw.Get("description").(string); ok {
		m.Description = v
	}
	if v, ok := raw.Get("stage").(string); ok {
		m.Stage = v
	}
	if s.Side == Repo {
		if err := m.repoLines(); err != nil {
			return nil, bad
		}
	} else if l, ok := raw.Get("depends").([]any); ok {
		for _, e := range l {
			if d, ok := e.(string); ok {
				m.Depends = append(m.Depends, d)
			}
		}
	}
	m.Files = filesOf(raw)
	return m, nil
}

// filesOf types the valid `[files]` tables. A bad table is a message of the
// checks, and it has no entry here.
func filesOf(raw *Obj) []FileEntry {
	files, ok := raw.Get("files").(*Obj)
	if !ok {
		return nil
	}
	var out []FileEntry
	for _, target := range files.Keys() {
		e, ok := files.Get(target).(*Obj)
		if !ok {
			continue
		}
		fe := FileEntry{Target: target, Keys: e}
		if v, ok := e.Get("mode").(string); ok && (v == "generated" || v == "seed" || v == "merge") {
			fe.Mode = v
		}
		if v, ok := e.Get("format").(string); ok && (v == "text" || v == "json" || v == "yaml" || v == "toml") {
			fe.Format = v
		}
		fe.Requires = pathList(e.Get("requires"))
		fe.RequiresFound = pathList(e.Get("requires_found"))
		fe.Remove = e.Get("remove")
		fe.RemoveItems = e.Get("remove_items")
		out = append(out, fe)
	}
	return out
}

// pathList gives the items of a non-empty list of strings, else nil.
func pathList(v any) []string {
	l, ok := asList(v)
	if !ok || len(l) == 0 {
		return nil
	}
	out := make([]string, len(l))
	for i, e := range l {
		s, ok := e.(string)
		if !ok {
			return nil
		}
		out[i] = s
	}
	return out
}

// repoLines is the port of `_RS_MANIFEST_TEMPLATE`. It gives an error when the
// template would fail.
func (m *Manifest) repoLines() error {
	t := m.Raw
	// The template of bash fails for a key that is not there.
	for _, k := range []string{"default_envs", "depends", "extensions", "root_only"} {
		if !t.Has(k) {
			return fmt.Errorf("%s", k)
		}
	}
	var ok bool
	if m.DefaultEnvs, ok = listOf(t, "default_envs"); !ok {
		return fmt.Errorf("default_envs")
	}
	if m.Depends, ok = listOf(t, "depends"); !ok {
		return fmt.Errorf("depends")
	}
	if m.Extensions, ok = listOf(t, "extensions"); !ok {
		return fmt.Errorf("extensions")
	}
	if b, isBool := t.Get("root_only").(bool); isBool {
		m.RootOnly = b
	}
	add := func(field, a, b, c string) { m.lines = append(m.lines, mline{field: field, a: a, b: b, c: c}) }
	for _, e := range m.DefaultEnvs {
		add("default_env", e, "", "")
	}
	for _, e := range m.Depends {
		add("depends", e, "", "")
	}
	for _, e := range m.Extensions {
		add("extension", e, "", "")
	}
	if t.Has("seed") {
		add("old", "seed", "", "")
	}
	if t.Has("seed_requires") {
		add("old", "seed_requires", "", "")
	}
	if t.Has("files") {
		if files, isMap := t.Get("files").(*Obj); isMap {
			for _, f := range files.Keys() {
				e, isMap := files.Get(f).(*Obj)
				if !isMap {
					add("bad", fmt.Sprintf("[files.\"%s\"] must be a table", f), "", "")
					continue
				}
				add("files", f, "-", "-")
				for _, k := range e.Keys() {
					v := e.Get(k)
					add("files", f, k, JSON(v))
					if k == "requires" || k == "requires_found" {
						items, good := checkedList(v)
						if good {
							for _, it := range items {
								add("files_item", f, k, it)
							}
						} else {
							what := "file names"
							if k == "requires" {
								what = "paths"
							}
							add("bad", fmt.Sprintf("[files.\"%s\"] %s must be a non-empty list of %s", f, k, what), "", "")
						}
					}
				}
			}
		} else {
			add("bad", "[files] must be a table of [files.\"<target>\"] tables", "", "")
		}
	}
	if mise, has := t.Get("mise").(*Obj); has {
		if tools, has := mise.Get("tools").(*Obj); has {
			for _, k := range tools.Keys() {
				v := tools.Get(k)
				mt := MiseTool{Name: k}
				ve, isMap := v.(*Obj)
				if isMap {
					if v, has := ve.Lookup("envs"); has {
						l, isList := asList(v)
						if !isList {
							return fmt.Errorf("envs")
						}
						mt.Envs = strList(l)
						if mt.Envs == nil {
							mt.Envs = []string{}
						}
					}
					if v, has := ve.Lookup("depends"); has {
						if s, isStr := v.(string); isStr {
							mt.Depends = []string{s}
						} else if l, isList := asList(v); isList {
							mt.Depends = strList(l)
						} else {
							return fmt.Errorf("depends")
						}
					}
				}
				envs := "-"
				if mt.Envs != nil {
					envs = strings.Join(mt.Envs, ",")
				}
				m.lines = append(m.lines, mline{field: "mise", a: k, b: envs, rest: mt.Depends})
				if isMap {
					if v, has := ve.Lookup("requires_found"); has {
						items, good := checkedList(v)
						if good {
							for _, it := range items {
								add("mise_found", k, it, "")
							}
							mt.RequiresFound = items
						} else {
							add("bad", fmt.Sprintf("requires_found of mise tool %s must be a non-empty list of file names", k), "", "")
						}
					}
					if v, has := ve.Lookup("requires"); has {
						add("old", "requires", k, str(v))
					}
					if ve.Has("files") {
						add("old", "files", k, "")
					}
				}
				m.Mise = append(m.Mise, mt)
			}
		}
	} else if t.Has("mise") {
		return fmt.Errorf("mise")
	}
	if me, has := t.Get("mise_env").(*Obj); has {
		for _, en := range me.Keys() {
			b, isMap := me.Get(en).(*Obj)
			if !isMap {
				continue
			}
			tools, isMap := b.Get("tools").(*Obj)
			if !isMap {
				continue
			}
			for _, k := range tools.Keys() {
				v, isMap := tools.Get(k).(*Obj)
				if !isMap {
					continue
				}
				for _, key := range []string{"requires_found", "requires", "files"} {
					if v.Has(key) {
						add("mise_env_key", key, k, en)
					}
				}
			}
		}
	}
	return nil
}

// checkedList tells if <v> is a non-empty list of strings, and gives the items.
func checkedList(v any) ([]string, bool) {
	l, ok := asList(v)
	if !ok || len(l) == 0 {
		return nil, false
	}
	out := make([]string, len(l))
	for i, e := range l {
		s, ok := e.(string)
		if !ok {
			return nil, false
		}
		out[i] = s
	}
	return out, true
}
