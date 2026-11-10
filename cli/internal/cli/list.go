package cli

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/selection"
)

// listRoot is a root of the tree of `list`: a pack or the direct selection.
type listRoot struct {
	name  string
	tools []string
}

// list is `bootstrap [mac] list` (and `ls`): the selected tools as a tree. A
// root is each pack that selects a tool, then the tools of the values file,
// and a tool shows its dependencies below it. It writes nothing.
func (a *app) list(mac bool) fail.Code {
	s, code := a.open(mac)
	if s != nil && s.c != nil {
		defer s.c.cleanup()
	}
	if code != fail.OK {
		return code
	}
	direct := "tools (selected directly)"
	base := selection.BasePack + " (pack)"
	var roots []listRoot
	at := map[string]int{}
	root := func(name string) *listRoot {
		i, ok := at[name]
		if !ok {
			i = len(roots)
			at[name] = i
			roots = append(roots, listRoot{name: name})
		}
		return &roots[i]
	}
	for _, e := range s.res.Sel {
		var r *listRoot
		switch {
		case strings.HasPrefix(e.Src, "tools."):
			r = root(direct)
		case e.Src == "default":
			r = root(base)
		case strings.HasPrefix(e.Src, "packs."):
			r = root(strings.TrimPrefix(e.Src, "packs.") + " (pack)")
		default:
			continue
		}
		r.tools = append(r.tools, e.Tool)
	}
	sort.SliceStable(roots, func(i, j int) bool {
		rank := func(n string) int {
			switch n {
			case base:
				return 0
			case direct:
				return 2
			}
			return 1
		}
		if ri, rj := rank(roots[i].name), rank(roots[j].name); ri != rj {
			return ri < rj
		}
		return roots[i].name < roots[j].name
	})
	deps := map[string][]string{}
	for _, d := range s.res.Deps {
		deps[d.Tool] = d.Deps
	}
	fmt.Fprint(a.stdout, renderTree(roots, deps, s.required))
	return fail.OK
}

// renderTree draws the roots, each with its sorted tools, and below a tool its
// sorted dependencies, recursively. A dependency that is already on the path
// to it (a loop of `depends`) is not drawn again. The `[required]` markers
// share one column.
func renderTree(roots []listRoot, deps map[string][]string, required func(string) bool) string {
	type line struct {
		text string
		req  bool
	}
	var lines []line
	isRoot := map[int]bool{}
	var walk func(tool, prefix string, last bool, path map[string]bool)
	walk = func(tool, prefix string, last bool, path map[string]bool) {
		branch, next := "├── ", "│   "
		if last {
			branch, next = "└── ", "    "
		}
		lines = append(lines, line{prefix + branch + tool, required(tool)})
		path[tool] = true
		var kids []string
		for _, d := range deps[tool] {
			if !path[d] {
				kids = append(kids, d)
			}
		}
		sort.Strings(kids)
		for i, d := range kids {
			walk(d, prefix+next, i == len(kids)-1, path)
		}
		delete(path, tool)
	}
	for _, r := range roots {
		isRoot[len(lines)] = true
		lines = append(lines, line{text: r.name})
		tools := append([]string(nil), r.tools...)
		sort.Strings(tools)
		tools = slices.Compact(tools)
		for i, t := range tools {
			walk(t, "", i == len(tools)-1, map[string]bool{})
		}
	}
	width := 0
	for i, l := range lines {
		if !isRoot[i] {
			width = max(width, utf8.RuneCountInString(l.text))
		}
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.text)
		if l.req {
			b.WriteString(strings.Repeat(" ", width-utf8.RuneCountInString(l.text)+2) + "[required]")
		}
		b.WriteByte('\n')
	}
	return b.String()
}
