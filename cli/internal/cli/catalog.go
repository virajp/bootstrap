package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/virajp/bootstrap/cli/internal/contract"
	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/projects"
	"github.com/virajp/bootstrap/cli/internal/questions"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// catalogData is the data of catalog --json (D14): the time of the last
// successful fetch, the categories, each tool of the clone and the packs of
// the repo side. Each field is the Go field in snake case.
type catalogData struct {
	FetchedAt  string         `json:"fetched_at"`
	Categories categoriesData `json:"categories"`
	Tools      []catalogTool  `json:"tools"`
	Packs      packsData      `json:"packs"`
}

// categoriesData is a source.Categories.
type categoriesData struct {
	Shown  string            `json:"shown"`
	Exists bool              `json:"exists"`
	IDs    []string          `json:"ids"`
	Titles map[string]string `json:"titles"`
}

// catalogTool is one tool of the catalog: its name, category and
// description, its questions and its [defaults] table.
type catalogTool struct {
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Questions   []questionData  `json:"questions"`
	Defaults    json.RawMessage `json:"defaults"`
}

// questionData is a questions.Question.
type questionData struct {
	Tool     string `json:"tool"`
	Key      string `json:"key"`
	Text     string `json:"text"`
	Detect   string `json:"detect"`
	Example  string `json:"example"`
	Group    string `json:"group"`
	Note     string `json:"note"`
	Optional bool   `json:"optional"`
}

// packsData is a source.Packs.
type packsData struct {
	Shown           string            `json:"shown"`
	Names           []string          `json:"names"`
	Members         []memberEntry     `json:"members"`
	Bad             []badEntry        `json:"bad"`
	Order           []orderEntry      `json:"order"`
	Required        map[string]bool   `json:"required"`
	Descriptions    map[string]string `json:"descriptions"`
	BadDescriptions []string          `json:"bad_descriptions"`
}

// memberEntry is a source.Member.
type memberEntry struct {
	Pack  string `json:"pack"`
	Key   string `json:"key"`
	State string `json:"state"`
}

// badEntry is a source.BadValue.
type badEntry struct {
	Pack string `json:"pack"`
	Key  string `json:"key"`
	Kind string `json:"kind"`
}

// orderEntry is a source.Order.
type orderEntry struct {
	Pack string `json:"pack"`
	Key  string `json:"key"`
}

// orEmpty gives <l>, or an empty list for nil: a list of the envelope is
// [], never null (D37).
func orEmpty[T any](l []T) []T {
	if l == nil {
		return []T{}
	}
	return l
}

// catalogCmd is `bootstrap catalog`: the tool catalog of the repo side of
// the clone (D14, D29). It needs no repository.
func (a *app) catalogCmd(o opts) fail.Code {
	c, idx, code := a.catalogIndex()
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	src := source.Source{Root: c.Dir, Side: source.Repo}
	packs, err := src.ReadPacks()
	if err != nil {
		return a.logErr(err)
	}
	cats := idx.Categories
	if !o.json {
		return a.printCatalog(idx, cats)
	}
	d := catalogData{
		Categories: categoriesData{Shown: cats.Shown, Exists: cats.Exists, IDs: orEmpty(cats.IDs), Titles: cats.Titles},
		Tools:      []catalogTool{},
		Packs: packsData{
			Shown: packs.Shown, Names: orEmpty(packs.Names), Members: []memberEntry{}, Bad: []badEntry{},
			Order: []orderEntry{}, Required: packs.Required, Descriptions: packs.Descriptions,
			BadDescriptions: orEmpty(packs.BadDescriptions),
		},
	}
	if !c.Fetched.IsZero() {
		d.FetchedAt = c.Fetched.UTC().Format(time.RFC3339)
	}
	for _, m := range packs.Members {
		d.Packs.Members = append(d.Packs.Members, memberEntry{m.Pack, m.Key, m.State})
	}
	for _, b := range packs.Bad {
		d.Packs.Bad = append(d.Packs.Bad, badEntry{b.Pack, b.Key, b.Kind})
	}
	for _, r := range packs.Order {
		d.Packs.Order = append(d.Packs.Order, orderEntry{r.Pack, r.Key})
	}
	for _, name := range idx.Names {
		m := idx.Manifests[name]
		t := catalogTool{Name: name, Category: m.Category, Description: m.Description, Questions: []questionData{}, Defaults: json.RawMessage("{}")}
		for _, q := range questions.Of(idx, []string{name}) {
			t.Questions = append(t.Questions, questionData{q.Tool, q.Key, q.Text, q.Detect, q.Example, q.Group, q.Note, q.Optional})
		}
		if def, ok := m.Raw.Get("defaults").(*source.Obj); ok {
			t.Defaults = json.RawMessage(source.JSON(def))
		}
		d.Tools = append(d.Tools, t)
	}
	a.data = d
	return fail.OK
}

// printCatalog prints the tools of <idx> by category, each with its
// summary, in the order of the categories file.
func (a *app) printCatalog(idx *source.Index, cats *source.Categories) fail.Code {
	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	for i, id := range cats.IDs {
		var names []string
		for _, n := range idx.Names {
			if idx.Manifests[n].Category == id {
				names = append(names, n)
			}
		}
		if len(names) == 0 {
			continue
		}
		if i > 0 {
			_, _ = fmt.Fprintln(w)
		}
		_, _ = fmt.Fprintln(w, a.title(cats.Titles[id]))
		for _, n := range names {
			_, _ = fmt.Fprintf(w, "  %s\t%s\n", n, idx.Manifests[n].Summary)
		}
	}
	_ = w.Flush()
	_, _ = fmt.Fprintln(a.stdout)
	report.NewPrinter(a.stdout, a.styled).Info(count(len(idx.Names), "tool"))
	return fail.OK
}

// title gives the title of a section: bold with the style.
func (a *app) title(s string) string {
	if a.styled {
		return report.Bold(s)
	}
	return s
}

// describeData is the data of describe --json (D10): the values of
// repo.toml, the selected tools with their values of tools.toml, the
// members, the task ports and the language servers (the data of lsp).
type describeData struct {
	Repo    json.RawMessage `json:"repo"`
	Tools   []describeTool  `json:"tools"`
	Members []describeMem   `json:"members"`
	Ports   []portData      `json:"ports"`
	LSP     lspData         `json:"lsp"`
}

// describeTool is a selected tool, its category in the catalog, and its
// [tools.<name>] table without env, {} when the table is not there. Its
// projects list is in the table when the file has it.
type describeTool struct {
	Name     string          `json:"name"`
	Category string          `json:"category"`
	Values   json.RawMessage `json:"values"`
}

// describeMem is a [members.<name>] table. Folder is the name when the
// table has no folder; present tells if the folder exists.
type describeMem struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Folder  string `json:"folder"`
	URL     string `json:"url"`
	Present bool   `json:"present"`
}

// portData is a task port (D3, D19, D39). A patterned port, for example
// git:merge:<branch>, also lists each task that matches it.
type portData struct {
	Name    string    `json:"name"`
	Present bool      `json:"present"`
	Tasks   *[]string `json:"tasks,omitempty"`
}

// describe is `bootstrap describe`: the facts of the repository (D10, D28,
// D40).
func (a *app) describe(o opts) fail.Code {
	sd, code := a.prepare(false, false)
	if code != fail.OK {
		return code
	}
	if !sd.Present() {
		a.fail(sd.Show(sd.Values)+" is not there", "run bootstrap init")
		return fail.Usage
	}
	c, _, code := a.catalogIndex()
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	sd.Clone = c.Dir
	repoRaw, err := source.ReadTOML(values.RepoFile(sd.Values))
	if err != nil {
		a.fail("cannot read "+sd.Show(values.RepoFile(sd.Values))+": "+err.Error(), "")
		return fail.Other
	}
	sel, err := core.SelectionOf(sd, sd.Values, sd.Values)
	if err != nil {
		return a.logErr(err)
	}
	tasks, err := core.TaskNames(sd.Root)
	if err != nil {
		return a.logErr(err)
	}
	servers, err := lspOf(sd.Root, sel)
	if err != nil {
		return a.logErr(err)
	}
	d := describeData{
		Repo: json.RawMessage(source.JSON(repoRaw)), Tools: []describeTool{},
		Members: []describeMem{}, Ports: []portData{}, LSP: servers,
	}
	for _, t := range sel.Tools {
		tbl, ok := sel.User.Get(t).(*source.Obj)
		if !ok {
			tbl = source.NewObj()
		}
		if l := projects.Links(sel.Raw, t); l != nil {
			tbl = tbl.Clone()
			list := make([]any, len(l))
			for i, n := range l {
				list[i] = n
			}
			tbl.Set("projects", list)
			tbl.Sort()
		}
		cat := ""
		if m := sel.Idx.Manifests[t]; m != nil {
			cat = m.Category
		}
		d.Tools = append(d.Tools, describeTool{Name: t, Category: cat, Values: json.RawMessage(source.JSON(tbl))})
	}
	if ms, ok := repoRaw.Get("members").(*source.Obj); ok {
		for _, name := range ms.Keys() {
			m := describeMem{Name: name, Folder: name}
			if e, ok := ms.Get(name).(*source.Obj); ok {
				if f, ok := e.Get("folder").(string); ok && f != "" {
					m.Folder = f
				}
				m.Type, _ = e.Get("type").(string)
				m.URL, _ = e.Get("url").(string)
			}
			st, err := os.Stat(filepath.Join(sd.Root, m.Folder))
			m.Present = err == nil && st.IsDir()
			d.Members = append(d.Members, m)
		}
	}
	for _, port := range contract.Ports {
		p := portData{Name: port}
		if !strings.Contains(port, "<") {
			for _, t := range tasks {
				p.Present = p.Present || t == port
			}
		} else {
			matched := []string{}
			for _, t := range tasks {
				if core.PortMatch(port, t) {
					matched = append(matched, t)
				}
			}
			p.Present, p.Tasks = len(matched) > 0, &matched
		}
		d.Ports = append(d.Ports, p)
	}
	if o.json {
		a.data = d
		return fail.OK
	}
	return a.printDescribe(repoRaw, d)
}

// printDescribe prints the four parts of describe as human lines: repo,
// tools, members and ports, each port present or missing.
func (a *app) printDescribe(repoRaw *source.Obj, d describeData) fail.Code {
	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	state := func(ok bool) string {
		if ok {
			return report.Glyph("ok", a.styled) + " present"
		}
		return report.Glyph("none", a.styled) + " missing"
	}
	_, _ = fmt.Fprintln(w, a.title("repo"))
	for _, k := range repoRaw.Keys() {
		if k == "members" {
			continue
		}
		_, _ = fmt.Fprintf(w, "  %s\t%s\n", k, source.JSON(repoRaw.Get(k)))
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, a.title("tools"))
	for _, t := range d.Tools {
		v := string(t.Values)
		if v == "{}" {
			v = ""
		}
		_, _ = fmt.Fprintf(w, "  %s\t%s\n", t.Name, v)
	}
	if len(d.Members) > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, a.title("members"))
		for _, m := range d.Members {
			_, _ = fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", m.Name, m.Type, m.Folder, state(m.Present))
		}
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, a.title("ports"))
	for _, p := range d.Ports {
		line := state(p.Present)
		if p.Tasks != nil && len(*p.Tasks) > 0 {
			line += " (" + strings.Join(*p.Tasks, ", ") + ")"
		}
		_, _ = fmt.Fprintf(w, "  %s\t%s\n", p.Name, line)
	}
	_ = w.Flush()
	return fail.OK
}
