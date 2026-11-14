package cli

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// listRoot is a root of a tree: a title with its tools below it. A root
// without a title draws each of its tools as a root.
type listRoot struct {
	name  string
	tools []string
}

// list is `bootstrap [mac] list` (and `ls`): the selected tools as a tree. A
// root is each selected tool that no selected tool needs, and a tool shows its
// dependencies below it. It writes nothing.
func (a *app) list(mac bool) fail.Code {
	s, code := a.open(mac)
	if s != nil && s.c != nil {
		defer s.c.cleanup()
	}
	if code != fail.OK {
		return code
	}
	roots := listRoots(s)
	deps := depsOf(s)
	style := a.styled
	opt := treeOpts{style: style, arrowCell: true, cell: func(tool string) (cell, string) {
		if !s.required(tool) {
			return cell{}, ""
		}
		g := report.Glyph("required", style)
		return cell{vis: g + " required", colored: g + " required"}, ""
	}}
	out, arrows := renderTree(roots, deps, opt)
	fmt.Fprint(a.stdout, out)
	text := count(len(s.res.Tools), "tool")
	if arrows {
		text += " · " + report.Arrow(style) + " = the subtree shows above"
	}
	fmt.Fprintln(a.stdout)
	report.NewPrinter(a.stdout, style).Info(text)
	return fail.OK
}

// depsOf gives the dependencies of each selected tool.
func depsOf(s *side) map[string][]string {
	deps := map[string][]string{}
	for _, d := range s.res.Deps {
		deps[d.Tool] = d.Deps
	}
	return deps
}

// listRoots gives the roots of the tree of the selection of <s>: each tool of
// a [tools.<name>] table, and each required member of the pack base, that no
// selected tool needs.
func listRoots(s *side) []listRoot {
	needed := map[string]bool{}
	for _, d := range s.res.Deps {
		for _, t := range d.Deps {
			if t != d.Tool {
				needed[t] = true
			}
		}
	}
	var tools []string
	for _, e := range s.res.Sel {
		if !needed[e.Tool] {
			tools = append(tools, e.Tool)
		}
	}
	return []listRoot{{tools: tools}}
}

// cell is the result column of a row of the tree: the visible text, its
// painted form, and a detail after it.
type cell struct{ vis, colored string }

// treeOpts sets what a row of the tree shows.
type treeOpts struct {
	style bool
	// include tells if a tool shows. A tool shows too when a tool below it
	// does. Nil shows each tool.
	include func(tool string) bool
	// cell gives the result column of a tool, and the detail after it. It
	// is called for the first copy of a tool only.
	cell func(tool string) (cell, string)
	// arrowCell also gives the result column to a later copy of a tool.
	arrowCell bool
	// groupTitle and group are rows after the tree, with the columns of the
	// tree: a lead (for example a glyph and a tool), a result and a detail.
	groupTitle string
	group      []groupRow
}

// groupRow is one row after the tree. lead is the visible text of the first
// column and painted its painted form.
type groupRow struct {
	lead, painted string
	cell          cell
	det           string
}

// renderTree draws the roots, each with its sorted tools, and below a tool its
// sorted dependencies, recursively. A dependency that is already on the path
// to it (a loop of `depends`) is not drawn again. The subtree of a tool shows
// at its first copy only; a later copy has the arrow. The result columns
// share one column. It also tells if it drew an arrow.
func renderTree(roots []listRoot, deps map[string][]string, opt treeOpts) (string, bool) {
	type line struct {
		text    string
		painted string // The text with colors; empty for the plain text.
		root    bool
		blank   bool
		cell    cell
		det     string
	}
	var lines []line
	drawn := map[string]bool{}
	arrows := false
	memo := map[string]bool{}
	var keep func(tool string, path map[string]bool) bool
	keep = func(tool string, path map[string]bool) bool {
		if opt.include == nil {
			return true
		}
		if v, ok := memo[tool]; ok {
			return v
		}
		if path[tool] {
			return false
		}
		path[tool] = true
		v := opt.include(tool)
		for _, d := range deps[tool] {
			if keep(d, path) {
				v = true
			}
		}
		delete(path, tool)
		memo[tool] = v
		return v
	}
	var walk func(tool, prefix string, last, top bool, path map[string]bool)
	walk = func(tool, prefix string, last, top bool, path map[string]bool) {
		branch, next := "├── ", "│   "
		if last {
			branch, next = "└── ", "    "
		}
		if top {
			branch, next = "", ""
		}
		var kids []string
		for _, d := range deps[tool] {
			if !path[d] && keep(d, map[string]bool{tool: true}) {
				kids = append(kids, d)
			}
		}
		sort.Strings(kids)
		l := line{text: prefix + branch + tool}
		if drawn[tool] && len(kids) > 0 {
			arrows = true
			l.text += " " + report.Arrow(opt.style)
			if opt.arrowCell && opt.cell != nil {
				l.cell, l.det = opt.cell(tool)
			}
			lines = append(lines, l)
			return
		}
		drawn[tool] = true
		if opt.cell != nil {
			l.cell, l.det = opt.cell(tool)
		}
		lines = append(lines, l)
		path[tool] = true
		for i, d := range kids {
			walk(d, prefix+next, i == len(kids)-1, false, path)
		}
		delete(path, tool)
	}
	for _, r := range roots {
		tools := append([]string(nil), r.tools...)
		sort.Strings(tools)
		tools = slices.Compact(tools)
		var shown []string
		for _, t := range tools {
			if keep(t, map[string]bool{}) {
				shown = append(shown, t)
			}
		}
		if len(shown) == 0 {
			continue
		}
		if r.name != "" {
			lines = append(lines, line{text: r.name, root: true})
		}
		for i, t := range shown {
			walk(t, "", i == len(shown)-1, r.name == "", map[string]bool{})
		}
	}
	if len(opt.group) > 0 {
		if len(lines) > 0 {
			lines = append(lines, line{blank: true})
		}
		lines = append(lines, line{text: opt.groupTitle, root: true, blank: true})
		for _, g := range opt.group {
			lines = append(lines, line{text: g.lead, painted: g.painted, cell: g.cell, det: g.det})
		}
	}
	width, cellW := 0, 0
	for _, l := range lines {
		if !l.root && !l.blank {
			width = max(width, utf8.RuneCountInString(l.text))
			cellW = max(cellW, utf8.RuneCountInString(l.cell.vis))
		}
	}
	var b strings.Builder
	for _, l := range lines {
		if l.blank {
			// A blank line, or the title of the group: not bold.
			b.WriteString(l.text)
			b.WriteByte('\n')
			continue
		}
		if l.root {
			if opt.style {
				b.WriteString(report.Bold(l.text))
			} else {
				b.WriteString(l.text)
			}
			b.WriteByte('\n')
			continue
		}
		if l.painted != "" {
			b.WriteString(l.painted)
		} else {
			b.WriteString(l.text)
		}
		if l.cell.vis != "" {
			b.WriteString(strings.Repeat(" ", width-utf8.RuneCountInString(l.text)+2) + l.cell.colored)
			if l.det != "" {
				b.WriteString(strings.Repeat(" ", cellW-utf8.RuneCountInString(l.cell.vis)+3) + l.det)
			}
		}
		b.WriteByte('\n')
	}
	return b.String(), arrows
}
