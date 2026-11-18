package cli

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/misegen"
	"github.com/virajp/bootstrap/cli/internal/repo"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// lspData is the data of lsp --json, and the lsp field of describe --json:
// the language servers of the selected tools. Each entry of servers is an
// entry of the lspServers of Claude, launched with `mise x --`; checks gives
// the command and the source of each server, for code:lsp:check.
type lspData struct {
	Servers map[string]lspEntry `json:"servers"`
	Checks  map[string]lspCheck `json:"checks"`
	tools   map[string]string   // The tool of each server, for the human lines.
	lines   map[string]string   // The command and its args, for the human lines.
}

// lspEntry is one entry of the lspServers of Claude.
type lspEntry struct {
	Command             string            `json:"command"`
	Args                []string          `json:"args"`
	ExtensionToLanguage map[string]string `json:"extensionToLanguage"`
	StartupTimeout      int64             `json:"startupTimeout,omitempty"`
}

// lspCheck is the command of the [lsp] table of a server, and its source.
type lspCheck struct {
	Command string `json:"command"`
	Source  string `json:"source"`
}

// lspOf gives the language servers of the selected tools of <sel> in the
// repository <root>. A server is there only when its tool is in dev and the
// requires_found of its source pin is met; a "host" source has no
// requires_found.
func lspOf(root string, sel *core.Selected) (lspData, error) {
	d := lspData{
		Servers: map[string]lspEntry{}, Checks: map[string]lspCheck{},
		tools: map[string]string{}, lines: map[string]string{},
	}
	has := func(name string) bool { return repo.HasFile(root, name) }
	envs, err := misegen.NewEnvs(source.Repo, []string{"dev"}, sel.Res, sel.Idx.Source.ToolsShown(), has)
	if err != nil {
		return d, err
	}
	for _, tool := range sel.Tools {
		m := sel.Idx.Manifests[tool]
		if m == nil || !envs.In(tool, "dev") {
			continue
		}
		for _, srv := range m.LSP {
			if !foundFor(m, srv.Source, has) {
				continue
			}
			d.Servers[srv.Name] = lspEntry{
				Command:             "mise",
				Args:                append([]string{"x", "--", srv.Command}, srv.Args...),
				ExtensionToLanguage: srv.Extensions,
				StartupTimeout:      srv.StartupTimeout,
			}
			d.Checks[srv.Name] = lspCheck{Command: srv.Command, Source: srv.Source}
			d.tools[srv.Name] = tool
			d.lines[srv.Name] = strings.Join(append([]string{srv.Command}, srv.Args...), " ")
		}
	}
	return d, nil
}

// foundFor tells if the requires_found of the mise tool <pin> of <m>, in
// [mise.tools] or in [mise_env.dev.tools], is met. A pin without
// requires_found, and "host", always are.
func foundFor(m *source.Manifest, pin string, has func(string) bool) bool {
	for _, mt := range m.Mise {
		if mt.Name == pin && !anyFound(mt.RequiresFound, has) {
			return false
		}
	}
	for _, mt := range m.MiseEnv {
		if mt.Env == "dev" && mt.Name == pin && !anyFound(mt.RequiresFound, has) {
			return false
		}
	}
	return true
}

// anyFound tells if the repository has a file of <names>, or <names> is
// empty: requires_found is met by any one of its names.
func anyFound(names []string, has func(string) bool) bool {
	for _, f := range names {
		if has(f) {
			return true
		}
	}
	return len(names) == 0
}

// lsp is `bootstrap lsp`: the language servers of the selected tools of the
// repository, one line each, or in the envelope with --json.
func (a *app) lsp(o opts) fail.Code {
	sd, code := a.prepare(false, false)
	if code != fail.OK {
		return code
	}
	if !sd.Present() {
		a.fail(sd.Show(sd.Values)+" is not there", "run bootstrap migrate")
		return fail.Usage
	}
	c, _, code := a.catalogIndex()
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	sd.Clone = c.Dir
	sel, err := core.SelectionOf(sd, sd.Values, sd.Values)
	if err != nil {
		return a.logErr(err)
	}
	d, err := lspOf(sd.Root, sel)
	if err != nil {
		return a.logErr(err)
	}
	if o.json {
		a.data = d
		return fail.OK
	}
	names := make([]string, 0, len(d.Servers))
	for n := range d.Servers {
		names = append(names, n)
	}
	sort.Strings(names)
	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	for _, n := range names {
		exts := make([]string, 0, len(d.Servers[n].ExtensionToLanguage))
		for x := range d.Servers[n].ExtensionToLanguage {
			exts = append(exts, x)
		}
		sort.Strings(exts)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", n, d.tools[n], d.lines[n], strings.Join(exts, " "))
	}
	_ = w.Flush()
	if len(names) > 0 {
		_, _ = fmt.Fprintln(a.stdout)
	}
	report.NewPrinter(a.stdout, a.styled).Info(count(len(names), "language server"))
	return fail.OK
}
