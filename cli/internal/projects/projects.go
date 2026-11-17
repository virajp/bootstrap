// Package projects holds the rules of the projects of a repo: the
// `[projects.<name>]` tables of `repo.toml`, the names, the workspace
// manifests that list project folders, and the folders that git tracks or
// does not ignore. detect and check use them.
//
// Contract: docs/design/bootstrap-config.md.
package projects

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// Project is a `[projects.<name>]` table: its name and its folder, relative
// to the root of the repo. The folder "." is the root.
type Project struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

// Root is the folder of the root project.
const Root = "."

// MaxDepth is the depth of the folder walk of detect: 4 folders below the
// root, as the test of requires_found.
const MaxDepth = 4

// The categories of the tools whose detect script finds a project folder
// (Languages), and whose match at the root links to the projects of a
// workspace manifest (Languages and PackageManagers).
const (
	Languages       = "languages"
	PackageManagers = "package-managers"
)

// RootLinks tells if a tool of the category <cat> links to a project by its
// match at the root: the root project, or the members of a workspace
// manifest in its evidence.
func RootLinks(cat string) bool { return cat == Languages || cat == PackageManagers }

// NameRule is the rule of a project name: lowercase kebab case.
var NameRule = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var nonName = regexp.MustCompile(`[^a-z0-9]+`)

// Normalize gives <s> in lowercase kebab case: My_App is my-app. The result
// is empty when <s> has no letter or digit.
func Normalize(s string) string {
	return strings.Trim(nonName.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// CleanFolder gives the folder <f> of a project in its one form, and false
// when it is not a folder of the repo: empty, absolute, or with a ..
// segment.
func CleanFolder(f string) (string, bool) {
	if f == "" || strings.HasPrefix(f, "/") {
		return "", false
	}
	c := path.Clean(f)
	for _, seg := range strings.Split(c, "/") {
		if seg == ".." {
			return "", false
		}
	}
	return c, true
}

// Listed gives the `[projects.<name>]` tables of the repo.toml values <raw>
// in name order. A table without a folder string is left out: the schema
// names it.
func Listed(raw *source.Obj) []Project {
	var out []Project
	if raw == nil {
		return nil
	}
	ps, ok := raw.Get("projects").(*source.Obj)
	if !ok {
		return nil
	}
	for _, n := range ps.Keys() {
		e, ok := ps.Get(n).(*source.Obj)
		if !ok {
			continue
		}
		f, ok := e.Get("folder").(string)
		if !ok {
			continue
		}
		if c, ok := CleanFolder(f); ok {
			f = c
		}
		out = append(out, Project{Name: n, Folder: f})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Links gives the projects list of the `[tools.<tool>]` table of the values
// <raw> (the [tools] table at its top), or nil.
func Links(raw *source.Obj, tool string) []string {
	if raw == nil {
		return nil
	}
	ts, ok := raw.Get("tools").(*source.Obj)
	if !ok {
		return nil
	}
	e, ok := ts.Get(tool).(*source.Obj)
	if !ok {
		return nil
	}
	l, _ := e.Get("projects").([]any)
	var out []string
	for _, x := range l {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// Inside tells if the folder <f> is <parent> or below it. Each folder is
// below the root, but the root holds no other project: it is a project only
// when no other one is found.
func Inside(f, parent string) bool {
	if parent == Root {
		return false
	}
	return f == parent || strings.HasPrefix(f, parent+"/")
}

// Depth gives the number of segments of the folder <f>; the root is 0.
func Depth(f string) int {
	if f == Root || f == "" {
		return 0
	}
	return strings.Count(f, "/") + 1
}

// InFolder tells if a match counts in the folder where its script ran: one
// of its evidence paths sits in that folder itself, so the path has no /.
// A script that searches below its folder matches in each parent folder too.
func InFolder(evidence []string) bool {
	for _, e := range evidence {
		e = strings.TrimSuffix(strings.TrimPrefix(e, "./"), "/")
		if e != "" && !strings.Contains(e, "/") {
			return true
		}
	}
	return false
}

// skipped are the folders that the walk never enters.
var skipped = map[string]bool{".git": true, "node_modules": true, "vendor": true}

// Folders gives each folder of the repo <root> that holds a file that git
// tracks or does not ignore, at any depth, without the root and without a
// folder below .git, node_modules or vendor. A gitignored folder is not
// there. A git submodule is not entered.
func Folders(root string) (map[string]bool, error) {
	cmd := exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	dirs := map[string]bool{}
	for _, f := range strings.Split(string(out), "\x00") {
		if f == "" {
			continue
		}
		segs := strings.Split(f, "/")
		for i := 1; i < len(segs); i++ {
			if skipped[segs[i-1]] {
				break
			}
			dirs[strings.Join(segs[:i], "/")] = true
		}
	}
	return dirs, nil
}

// Manifests are the workspace manifests at the root of a repo that list
// project folders.
var Manifests = []string{"pnpm-workspace.yaml", "aube-workspace.yaml", "go.work", "Cargo.toml", "pyproject.toml"}

// marker is the file that a member folder of each workspace manifest
// holds: a member is a package of its workspace.
var marker = map[string]string{
	"pnpm-workspace.yaml": "package.json",
	"aube-workspace.yaml": "package.json",
	"go.work":             "go.mod",
	"Cargo.toml":          "Cargo.toml",
	"pyproject.toml":      "pyproject.toml",
}

// IsManifest tells if the evidence path <e> is a workspace manifest at the
// root.
func IsManifest(e string) bool {
	e = strings.TrimPrefix(e, "./")
	for _, m := range Manifests {
		if e == m {
			return true
		}
	}
	return false
}

// Members gives the member folders of the workspace manifest <name> at the
// root <root>, among the folders <dirs> (see Folders): its globs expanded,
// without its excluded folders and without the root. A member folder holds
// the package file of its workspace, for example package.json for
// pnpm-workspace.yaml. A manifest that is not there, or that lists no
// member, gives nil.
func Members(root, name string, dirs map[string]bool) ([]string, error) {
	p := filepath.Join(root, name)
	if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
		return nil, nil
	}
	var include, exclude []string
	switch name {
	case "pnpm-workspace.yaml", "aube-workspace.yaml":
		out, _, err := format.YQ("-p", "yaml", "-o", "json", "-I", "0", "[.packages // [] | .[] | select(tag == \"!!str\")]", p)
		if err != nil {
			return nil, err
		}
		var list []string
		if err := json.Unmarshal([]byte(out), &list); err != nil {
			return nil, err
		}
		for _, g := range list {
			if rest, ok := strings.CutPrefix(g, "!"); ok {
				exclude = append(exclude, rest)
			} else {
				include = append(include, g)
			}
		}
	case "go.work":
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		include = goWorkUse(string(b))
	case "Cargo.toml", "pyproject.toml":
		raw, err := source.ReadTOML(p)
		if err != nil {
			return nil, err
		}
		keys := []string{"workspace"}
		if name == "pyproject.toml" {
			keys = []string{"tool", "uv", "workspace"}
		}
		var cur any = raw
		for _, k := range keys {
			o, ok := cur.(*source.Obj)
			if !ok {
				cur = nil
				break
			}
			cur = o.Get(k)
		}
		ws, ok := cur.(*source.Obj)
		if !ok {
			return nil, nil
		}
		include, exclude = stringList(ws.Get("members")), stringList(ws.Get("exclude"))
	}
	excluded := map[string]bool{}
	for _, g := range exclude {
		for _, f := range Glob(g, dirs) {
			excluded[f] = true
		}
	}
	seen := map[string]bool{}
	var out []string
	for _, g := range include {
		for _, f := range Glob(g, dirs) {
			if st, err := os.Stat(filepath.Join(root, f, marker[name])); err != nil || !st.Mode().IsRegular() {
				continue
			}
			if !excluded[f] && !seen[f] {
				seen[f] = true
				out = append(out, f)
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// stringList gives the strings of the list <v>.
func stringList(v any) []string {
	l, _ := v.([]any)
	var out []string
	for _, x := range l {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// goWorkUse gives the folders of the use directives of the go.work text
// <text>: `use ./a` and `use ( ./a ./b )`.
func goWorkUse(text string) []string {
	var out []string
	block := false
	for _, line := range strings.Split(text, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch {
		case block && f[0] == ")":
			block = false
		case block:
			out = append(out, strings.Trim(f[0], "\"`"))
		case f[0] == "use" && len(f) >= 2 && f[1] == "(":
			block = true
		case f[0] == "use" && len(f) >= 2 && strings.HasPrefix(f[1], "("):
			block = true
			if rest := strings.TrimSuffix(strings.TrimPrefix(f[1], "("), ")"); rest != "" {
				out = append(out, strings.Trim(rest, "\"`"))
			}
		case f[0] == "use" && len(f) >= 2:
			out = append(out, strings.Trim(f[1], "\"`"))
		}
	}
	return out
}

// Glob gives the folders of <dirs> that the pattern <g> names, sorted. A
// segment ** matches zero or more segments; the other segments match as
// path.Match does. The pattern . (the root) names no folder.
func Glob(g string, dirs map[string]bool) []string {
	g = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(g), "./"), "/")
	if g == "" || g == "." || strings.HasPrefix(g, "/") {
		return nil
	}
	g = path.Clean(g)
	pat := strings.Split(g, "/")
	var out []string
	for d := range dirs {
		if globMatch(pat, strings.Split(d, "/")) {
			out = append(out, d)
		}
	}
	sort.Strings(out)
	return out
}

func globMatch(pat, segs []string) bool {
	if len(pat) == 0 {
		return len(segs) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(segs); i++ {
			if globMatch(pat[1:], segs[i:]) {
				return true
			}
		}
		return false
	}
	if len(segs) == 0 {
		return false
	}
	if ok, err := path.Match(pat[0], segs[0]); err != nil || !ok {
		return false
	}
	return globMatch(pat[1:], segs[1:])
}
