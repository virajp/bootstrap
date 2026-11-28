package macos

import (
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// shipped is one file of a selected tool: a line of `files` in `_ss_plan`.
type shipped struct {
	tool, path, target string
	backup, tmpl       bool
}

// row gives the manifest row path of the file: its handled target.
func (f shipped) row() string { return rowPath(f.target, f.backup) }

// makePlan is `_ss_plan`. Each file of a selected tool, except its tools.toml,
// gives a target relative to $HOME, without `.tmpl`. It sets the plan (in the
// order of the run: each stage, its mise files first) and the handled targets
// for the stale rule. The handled targets also hold the mise file of each tool
// whose mise config cannot generate, so the stale rule keeps its copy. It
// gives 2, after a message for each problem.
func (s *state) makePlan() fail.Code {
	shown := s.src.ToolsShown()
	var files []shipped
	for _, tool := range s.res.Tools {
		for _, f := range s.idx.Files(tool) {
			// An entry whose `when` is false is not shipped.
			if !source.When(s.all, f.When) {
				continue
			}
			files = append(files, shipped{tool, f.Path, f.Target, f.Backup, f.Tmpl})
		}
	}
	rc := fail.OK

	// A target that more than one file gives. A target in the backup clone
	// and the same path below $HOME are two targets.
	count := map[string]int{}
	for _, f := range files {
		count[f.row()]++
	}
	var dups []string
	for t, n := range count {
		if n > 1 {
			dups = append(dups, t)
		}
	}
	sort.Strings(dups)
	for _, t := range dups {
		var b strings.Builder
		var where string
		for _, f := range files {
			if f.row() == t {
				b.WriteString(" " + shown + "/" + f.tool + "/" + f.path)
				where = s.showTarget(f.target, f.backup)
			}
		}
		s.log("%s comes from more than one source:%s", where, b.String())
		rc = fail.Usage
	}
	for _, f := range files {
		if !f.backup && strings.HasPrefix(f.target, confTarget+"/") {
			s.log("%s/%s/%s: the sync generates .config/mise/conf.d/ from [mise.*] of tools.toml - remove the file", shown, f.tool, f.path)
			rc = fail.Usage
		}
	}
	if rc != fail.OK {
		return rc
	}

	// The mode and the format of each file: its [files] table, else the
	// defaults. The generated mise files come first.
	var all []planItem
	for _, g := range s.gen {
		all = append(all, planItem{g.tool, g.file, g.target, false, "generated", "toml", false})
	}
	for _, f := range files {
		mode, format := "generated", "-"
		if fe := s.entry(f.tool, f.row()); fe != nil {
			if fe.Mode != "" {
				mode = fe.Mode
			}
			if fe.Format != "" {
				format = fe.Format
			}
		}
		all = append(all, planItem{f.tool, f.path, f.target, f.backup, mode, format, f.tmpl})
	}

	s.plan = nil
	s.handled = nil
	for _, b := range s.miseBad {
		s.handled = append(s.handled, confTarget+"/"+b.tool+"/mise.toml")
	}
	for _, stage := range []string{"init", "main"} {
		for _, pass := range []string{"mise", "rest"} {
			for _, it := range all {
				if s.res.Init[it.tool] != (stage == "init") {
					continue
				}
				if (!it.backup && strings.HasPrefix(it.target, ".config/mise/")) != (pass == "mise") {
					continue
				}
				s.plan = append(s.plan, it)
				s.handled = append(s.handled, it.row())
			}
		}
	}
	return fail.OK
}

// row gives the manifest row path of the item: its handled target.
func (it planItem) row() string { return rowPath(it.target, it.backup) }

// showTarget gives the target <target> as an output line shows it: ~/<target>,
// or for a target in the backup clone its path in the clone. Without backup
// in macos.toml, a target in the clone shows as @backup/<target>.
func (s *state) showTarget(target string, backup bool) string {
	if !backup {
		return tilde + "/" + target
	}
	if s.clone == "" {
		return rowPath(target, true)
	}
	return s.showPath(s.clone + "/" + target)
}

// entry gives the [files] table of <tool> whose manifest row path is <row>,
// or nil.
func (s *state) entry(tool, row string) *source.FileEntry {
	m := s.idx.Manifests[tool]
	if m == nil {
		return nil
	}
	for i := range m.Files {
		if rowPath(m.Files[i].Target, m.Files[i].Backup) == row {
			return &m.Files[i]
		}
	}
	return nil
}

// unhandle removes the first <target> from the handled targets, so the stale
// rule can remove a listed copy.
func (s *state) unhandle(target string) {
	for i, h := range s.handled {
		if h == target {
			s.handled = append(s.handled[:i:i], s.handled[i+1:]...)
			return
		}
	}
}
