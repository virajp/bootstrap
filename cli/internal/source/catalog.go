package source

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/BurntSushi/toml"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// The catalog keys of a tool: summary, category, selectable, description
// and, on the repo side, hosts. Contract: docs/design/tools.md.

// SummaryMax is the most characters that a summary can have.
const SummaryMax = 47

// HostNames are the host names of `hosts`: the OS names of mise.
var HostNames = []string{"macos", "linux", "windows"}

// GOOS is the Go name of the host. Tests replace it.
var GOOS = runtime.GOOS

// HostName gives the name of the host as `hosts` names it: darwin is macos,
// and another Go name stays as it is.
func HostName() string {
	if GOOS == "darwin" {
		return "macos"
	}
	return GOOS
}

// RunsOn tells if the tool can be selected on the host <host>: it has no
// hosts, or its hosts name <host>.
func (m *Manifest) RunsOn(host string) bool {
	if !m.Raw.Has("hosts") {
		return true
	}
	for _, h := range m.Hosts {
		if h == host {
			return true
		}
	}
	return false
}

// catalogRules examines the catalog keys of <tool>, on both sides.
func (x *Index) catalogRules(tool string) []*fail.Error {
	m := x.Manifests[tool]
	at := x.Source.ManifestShown(tool)
	var out []*fail.Error
	bad := func(text string) { out = append(out, fail.New(fail.Usage, "%s: %s", at, text)) }
	// A tool needs a summary and a category: a missing key and an empty
	// string are the same problem.
	if v, has := m.Raw.Lookup("summary"); !has || v == "" {
		bad("summary is missing")
	} else if s, ok := v.(string); !ok || utf8.RuneCountInString(s) > SummaryMax {
		bad("summary must be a string of 47 characters or fewer")
	}
	if v, has := m.Raw.Lookup("category"); !has || v == "" {
		bad("category is missing")
	} else if _, ok := v.(string); !ok {
		bad("category must be a string")
	}
	if v, has := m.Raw.Lookup("selectable"); has {
		if _, ok := v.(bool); !ok {
			bad("selectable must be true or false")
		}
	}
	// macosRules examines the description of a Mac tool.
	if v, has := m.Raw.Lookup("description"); has && x.Source.Side == Repo {
		if _, ok := v.(string); !ok {
			bad("description must be a string")
		}
	}
	if v, has := m.Raw.Lookup("hosts"); has && x.Source.Side == Repo {
		l, ok := v.([]any)
		switch {
		case !ok:
			bad("hosts must be a list of macos, linux and windows")
		case len(l) == 0:
			bad("hosts must name at least one host")
		default:
			for _, e := range l {
				if h, ok := e.(string); !ok || !hostKnown(h) {
					bad("hosts must be a list of macos, linux and windows")
					break
				}
			}
		}
	}
	return out
}

func hostKnown(h string) bool {
	for _, n := range HostNames {
		if n == h {
			return true
		}
	}
	return false
}

// Categories is the categories file of a side: `[category.<id>]` tables, each
// with a title.
type Categories struct {
	Shown  string
	Exists bool
	IDs    []string          // The ids, in the order of the file.
	Titles map[string]string // The title of each id.
}

// Has tells if the file has the category <id>.
func (c *Categories) Has(id string) bool { _, ok := c.Titles[id]; return ok }

// CategoriesShown gives the categories file as a message names it.
func (s Source) CategoriesShown() string { return string(s.Side) + "/categories.toml" }

// CategoriesFile gives the absolute path of the categories file.
func (s Source) CategoriesFile() string {
	return filepath.Join(s.Root, string(s.Side), "categories.toml")
}

// ReadCategories reads the categories file of the side. A missing file gives
// no category and no problem.
func (s Source) ReadCategories() (*Categories, []*fail.Error) {
	c := &Categories{Shown: s.CategoriesShown(), Titles: map[string]string{}}
	file := s.CategoriesFile()
	st, err := os.Stat(file)
	if err != nil || !st.Mode().IsRegular() {
		return c, nil
	}
	c.Exists = true
	var raw map[string]any
	md, err := toml.DecodeFile(file, &raw)
	if err != nil {
		return c, []*fail.Error{readErr("cannot read %s", c.Shown)}
	}
	var out []*fail.Error
	bad := func(text string) { out = append(out, fail.New(fail.Usage, "%s: %s", c.Shown, text)) }
	root := FromMap(raw).(*Obj)
	for _, k := range root.Keys() {
		if k != "category" {
			bad(k + " is not a key of the categories file - each entry is [category.<id>] with a title")
		}
	}
	cats, ok := root.Get("category").(*Obj)
	if root.Has("category") && !ok {
		bad("[category] must be a table of [category.<id>] tables")
		return c, out
	}
	if cats == nil {
		return c, out
	}
	// The ids in the order of the file.
	seen := map[string]bool{}
	var order []string
	for _, k := range md.Keys() {
		if len(k) >= 2 && k[0] == "category" && !seen[k[1]] {
			seen[k[1]] = true
			order = append(order, k[1])
		}
	}
	for _, id := range order {
		e, ok := cats.Get(id).(*Obj)
		if !ok {
			bad("[category." + id + "] must be a table")
			continue
		}
		for _, k := range e.Keys() {
			if k != "title" {
				bad("[category." + id + "] " + k + " is not a key of a category - the only key is title")
			}
		}
		t, has := e.Lookup("title")
		title, isStr := t.(string)
		switch {
		case !has:
			bad("[category." + id + "] needs a title string")
		case !isStr:
			bad("[category." + id + "] title must be a string")
		default:
			c.IDs = append(c.IDs, id)
			c.Titles[id] = title
		}
	}
	return c, out
}

// categoryRule examines the category of <tool> against the categories file.
func (x *Index) categoryRule(tool string) []*fail.Error {
	m := x.Manifests[tool]
	if m.Category == "" {
		return nil // catalogRules names a missing category.
	}
	if _, ok := m.Raw.Get("category").(string); !ok {
		return nil // catalogRules names the type.
	}
	at := x.Source.ManifestShown(tool)
	c := x.Categories
	if c == nil || !c.Exists {
		return []*fail.Error{fail.New(fail.Usage, "%s: category %s needs the file %s", at, m.Category, x.Source.CategoriesShown())}
	}
	if !c.Has(m.Category) {
		return []*fail.Error{fail.New(fail.Usage, "%s: category %s is not a category of %s", at, m.Category, c.Shown)}
	}
	return nil
}

// Dependents gives the tools of the index whose depends name <tool>, sorted.
func (x *Index) Dependents(tool string) []string {
	var out []string
	for _, n := range x.Names {
		for _, d := range x.Manifests[n].Depends {
			if d == tool {
				out = append(out, n)
				break
			}
		}
	}
	return out
}

// NotSelectable gives the message of a tool with `selectable = false`: it
// names the tools that need it.
func (x *Index) NotSelectable(tool string) string {
	by := x.Dependents(tool)
	if len(by) == 0 {
		return tool + " is not selectable."
	}
	return tool + " is not selectable. Select " + orList(by) + "."
}

// HostMessage gives the message of a tool whose hosts leave out <host>.
func (m *Manifest) HostMessage(host string) string {
	return m.Tool + " needs the host " + orList(m.Hosts) + " - this host is " + host
}

// orList gives "a", "a or b", "a, b or c".
func orList(l []string) string {
	if len(l) <= 1 {
		return strings.Join(l, "")
	}
	return strings.Join(l[:len(l)-1], ", ") + " or " + l[len(l)-1]
}
