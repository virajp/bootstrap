package cli

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/selection"
)

// list is `bootstrap [mac] list` (and `ls`): each tool that the values file
// selects, sorted, with the reasons of its selection. It writes nothing.
func (a *app) list(mac bool) fail.Code {
	s, code := a.open(mac)
	if s != nil && s.c != nil {
		defer s.c.cleanup()
	}
	if code != fail.OK {
		return code
	}
	why := map[string][]string{}
	add := func(tool, reason string) {
		for _, r := range why[tool] {
			if r == reason {
				return
			}
		}
		why[tool] = append(why[tool], reason)
	}
	for _, e := range s.res.Sel {
		switch {
		case strings.HasPrefix(e.Src, "tools."):
			add(e.Tool, "tools")
		case e.Src == "default":
			add(e.Tool, "pack "+selection.BasePack)
		default:
			add(e.Tool, "pack "+strings.TrimPrefix(e.Src, "packs."))
		}
	}
	needs := map[string][]string{}
	for _, d := range s.res.Deps {
		for _, dep := range d.Deps {
			needs[dep] = append(needs[dep], d.Tool)
		}
	}
	for dep, by := range needs {
		sort.Strings(by)
		add(dep, "dependency of "+strings.Join(by, ", "))
	}
	tools := append([]string(nil), s.res.Tools...)
	sort.Strings(tools)
	tw := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	for _, t := range tools {
		reasons := why[t]
		if s.required(t) {
			reasons = append(reasons, "required")
		}
		fmt.Fprintf(tw, "%s\t%s\n", t, strings.Join(reasons, "; "))
	}
	_ = tw.Flush()
	return fail.OK
}
