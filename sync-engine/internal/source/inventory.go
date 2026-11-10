package source

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
)

// The folders of a tool folder that hold the sources of its entries.
const (
	FilesDir = "files"
	TasksDir = "tasks"
	// TasksTarget is the folder of the task files in a target.
	TasksTarget = ".config/mise/tasks"
)

// File is one file that a tool ships: one file of an entry, or of a folder
// entry.
type File struct {
	Path   string // The source, relative to the tool folder: files/<source> or tasks/<source>.
	Target string // The path in the target.
	Tmpl   bool   // The source is a template: mode template, a seed with render, or merge.
}

// Files gives the files that <tool> ships, from the entries of its
// tools.toml, in byte order of the target. A folder entry gives each regular
// file below its folder.
func (x *Index) Files(tool string) []File {
	m := x.Manifests[tool]
	if m == nil {
		return nil
	}
	out := make([]File, 0, len(m.Files))
	for _, fe := range m.Files {
		out = append(out, File{Path: fe.Source, Target: fe.Target, Tmpl: fe.Tmpl})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Target < out[j].Target })
	return out
}

// AnyTemplate tells if a tool of the index has a template file.
func (x *Index) AnyTemplate() bool {
	for _, m := range x.Manifests {
		for _, fe := range m.Files {
			if fe.Tmpl {
				return true
			}
		}
	}
	return false
}

// regularFiles gives the regular files below <dir>, relative to it, sorted.
// The walk does not follow a link.
func regularFiles(dir string) []string {
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Type().IsRegular() {
			rel, _ := filepath.Rel(dir, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

// entryKind is the kind of an entry table: [files] or [tasks].
type entryKind struct {
	table  string // "files" or "tasks"
	dir    string // FilesDir or TasksDir
	modes  []string
	keys   []string
	target func(key string) string
}

var filesKind = entryKind{
	table: "files", dir: FilesDir,
	modes:  []string{"copy", "template", "seed", "merge"},
	keys:   []string{"source", "mode", "render", "format", "requires", "requires_found", "remove", "remove_items"},
	target: func(k string) string { return k },
}

var tasksKind = entryKind{
	table: "tasks", dir: TasksDir,
	modes:  []string{"copy", "template"},
	keys:   []string{"source", "mode", "requires", "requires_found", "library"},
	target: func(k string) string { return TasksTarget + "/" + k },
}

func oneOf(s string, list []string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func quoted(list []string) string {
	q := make([]string, len(list))
	for i, s := range list {
		q[i] = `"` + s + `"`
	}
	if len(q) == 1 {
		return q[0]
	}
	return strings.Join(q[:len(q)-1], ", ") + " or " + q[len(q)-1]
}

// entries reads the [files] and [tasks] tables of <raw> into typed entries,
// and gives a message for each problem. A folder entry gives one entry for
// each regular file below its folder. <shown> is the tools.toml as a message
// names it, and <dir> the absolute tool folder.
func (s Source) entries(shown, dir string, raw *Obj) ([]FileEntry, []*fail.Error) {
	var out []FileEntry
	var probs []*fail.Error
	add := func(format string, args ...any) {
		probs = append(probs, fail.New(fail.Usage, "%s: %s", shown, fmt.Sprintf(format, args...)))
	}
	// covered holds each source file that an entry ships.
	covered := map[string]bool{}
	for _, kind := range []entryKind{filesKind, tasksKind} {
		v, has := raw.Lookup(kind.table)
		if !has {
			continue
		}
		tbl, ok := v.(*Obj)
		if !ok {
			add("[%s] must be a table of [%s.\"<path>\"] tables", kind.table, kind.table)
			continue
		}
		for _, key := range tbl.Keys() {
			at := fmt.Sprintf("[%s.\"%s\"]", kind.table, key)
			e, ok := tbl.Get(key).(*Obj)
			if !ok {
				add("%s must be a table", at)
				continue
			}
			bad := false
			fail1 := func(format string, args ...any) {
				add("%s %s", at, fmt.Sprintf(format, args...))
				bad = true
			}
			if badRelPath(key) || strings.HasSuffix(key, "/") {
				fail1("must be a path relative to the target root, with no .. part")
			}
			for _, k := range e.Keys() {
				if !oneOf(k, kind.keys) {
					fail1("%s is not a key of a %s entry - the keys are %s", k, kind.table, strings.Join(kind.keys, ", "))
				}
			}
			if s.Side == Macos && e.Has("requires_found") {
				fail1("requires_found is for repo tools only")
			}
			fe := FileEntry{Target: kind.target(key), Keys: e, Task: kind.table == "tasks"}
			src, isStr := e.Get("source").(string)
			if !isStr || badRelPath(src) || strings.HasSuffix(src, "/") {
				fail1("needs source, a path below %s/ with no .. part", kind.dir)
			}
			mode := "copy"
			if v, has := e.Lookup("mode"); has {
				m, isStr := v.(string)
				if !isStr || !oneOf(m, kind.modes) {
					fail1("mode must be %s", quoted(kind.modes))
				} else {
					mode = m
				}
			}
			render := false
			if v, has := e.Lookup("render"); has {
				b, isBool := v.(bool)
				switch {
				case !isBool:
					fail1("render must be true or false")
				case b && mode != "seed":
					fail1("render works only with mode = \"seed\" - a template has mode = \"template\"")
				default:
					render = b
				}
			}
			if v, has := e.Lookup("library"); has {
				b, isBool := v.(bool)
				switch {
				case !isBool:
					fail1("library must be true or false")
				case b != strings.HasPrefix(key, "_scripts/"):
					fail1("library = true is for a library below _scripts/, and each file below _scripts/ is a library")
				}
			} else if strings.HasPrefix(key, "_scripts/") {
				fail1("a file below _scripts/ is a library - write library = true")
			}
			if v, has := e.Lookup("format"); has {
				f, isStr := v.(string)
				if !isStr || !oneOf(f, []string{"text", "json", "yaml", "toml"}) {
					fail1("format must be \"text\", \"json\", \"yaml\" or \"toml\"")
				} else {
					fe.Format = f
				}
			}
			for _, k := range []string{"requires", "requires_found"} {
				v, has := e.Lookup(k)
				if !has {
					continue
				}
				items, good := checkedList(v)
				if !good {
					what := "file names"
					if k == "requires" {
						what = "paths"
					}
					fail1("%s must be a non-empty list of %s", k, what)
					continue
				}
				for _, it := range items {
					if k == "requires" && badRelPath(it) {
						where := "the repository root"
						if s.Side == Macos {
							where = "$HOME"
						}
						fail1("requires: '%s' must be a path relative to %s, with no .. part", it, where)
					} else if k == "requires_found" && (it == "" || strings.Contains(it, "/")) {
						fail1("requires_found: '%s' must be a file name, with no /", it)
					}
				}
				if k == "requires" {
					fe.Requires = items
				} else {
					fe.RequiresFound = items
				}
			}
			for _, k := range []string{"remove", "remove_items"} {
				v, has := e.Lookup(k)
				if !has {
					continue
				}
				if mode != "merge" {
					fail1("%s works only with mode = \"merge\"", k)
				}
				if er := CheckFileShape(shown, key, k, v); er != nil {
					probs = append(probs, er)
					bad = true
				}
			}
			if mode == "merge" && formatOf(key, fe.Format) == "text" {
				fail1("mode = \"merge\" needs a json, yaml or toml target")
			}
			fe.Remove = e.Get("remove")
			fe.RemoveItems = e.Get("remove_items")
			switch mode {
			case "seed", "merge":
				fe.Mode = mode
			default:
				fe.Mode = "generated"
			}
			fe.Tmpl = mode == "template" || mode == "merge" || render
			rel := kind.dir + "/" + src
			if bad {
				// The source counts as covered, so its file gets no second
				// message.
				if isStr {
					covered[rel] = true
					for _, f := range regularFiles(filepath.Join(dir, rel)) {
						covered[rel+"/"+f] = true
					}
				}
				continue
			}
			abs := filepath.Join(dir, rel)
			st, err := os.Lstat(abs)
			switch {
			case err != nil:
				fail1("source: %s/%s is not there", kind.dir, src)
			case st.Mode().IsRegular():
				fe.Source = rel
				covered[rel] = true
				out = append(out, fe)
			case st.IsDir():
				list := regularFiles(abs)
				if len(list) == 0 {
					fail1("source: the folder %s/%s has no file", kind.dir, src)
				}
				for _, f := range list {
					one := fe
					one.Source = rel + "/" + f
					one.Target = fe.Target + "/" + f
					covered[one.Source] = true
					out = append(out, one)
				}
			default:
				fail1("source: %s/%s is not a file or a folder", kind.dir, src)
			}
		}
	}
	// D1: each file of files/ and tasks/ has an entry. The old form: a file
	// outside these folders.
	for _, f := range regularFiles(dir) {
		top, _, _ := strings.Cut(f, "/")
		switch {
		case f == "tools.toml":
		case top == FilesDir || top == TasksDir:
			if !covered[f] {
				table := "files"
				if top == TasksDir {
					table = "tasks"
				}
				add("%s has no entry - write a [%s.\"<path>\"] table with source = \"%s\"", f, table, strings.TrimPrefix(f, top+"/"))
			}
		default:
			add("%s is not below files/ or tasks/ - move it, and write an entry for it (the old form of a tool folder is gone)", f)
		}
	}
	return out, probs
}
