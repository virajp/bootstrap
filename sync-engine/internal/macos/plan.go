package macos

import (
	"sort"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// shipped is one file of a selected tool: a line of `files` in `_ss_plan`.
type shipped struct{ tool, path, target string }

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
		list, _ := s.src.Files(tool)
		for _, f := range list {
			files = append(files, shipped{tool, f.Path, f.Target})
		}
	}
	rc := fail.OK

	// A target that more than one file gives.
	count := map[string]int{}
	for _, f := range files {
		count[f.target]++
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
		for _, f := range files {
			if f.target == t {
				b.WriteString(" " + shown + "/" + f.tool + "/" + f.path)
			}
		}
		s.log("%s/%s comes from more than one source:%s", tilde, t, b.String())
		rc = fail.Usage
	}
	for _, f := range files {
		if strings.HasPrefix(f.target, confTarget+"/") {
			s.log("%s/%s/%s: the sync generates .config/mise/conf.d/ from [mise.*] of tools.toml - remove the file", shown, f.tool, f.path)
			rc = fail.Usage
		}
	}
	// Each [files] table of a selected tool names a target of its tool.
	selected := map[string]bool{}
	for _, t := range s.res.Tools {
		selected[t] = true
	}
	for _, t := range s.idx.Names {
		if !selected[t] {
			continue
		}
		for _, fe := range s.idx.Manifests[t].Files {
			found := false
			for _, f := range files {
				if f.tool == t && f.target == fe.Target {
					found = true
					break
				}
			}
			if !found {
				s.log("%s/%s/tools.toml: [files.\"%s\"] names a target that no file of the tool produces", shown, t, fe.Target)
				rc = fail.Usage
			}
		}
	}
	if rc != fail.OK {
		return rc
	}

	// The mode and the format of each file: its [files] table, else the
	// defaults. The generated mise files come first.
	var all []planItem
	for _, g := range s.gen {
		all = append(all, planItem{g.tool, g.file, g.target, "generated", "toml"})
	}
	for _, f := range files {
		mode, format := "generated", "-"
		if fe := s.entry(f.tool, f.target); fe != nil {
			if fe.Mode != "" {
				mode = fe.Mode
			}
			if fe.Format != "" {
				format = fe.Format
			}
		}
		all = append(all, planItem{f.tool, f.path, f.target, mode, format})
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
				if strings.HasPrefix(it.target, ".config/mise/") != (pass == "mise") {
					continue
				}
				s.plan = append(s.plan, it)
				s.handled = append(s.handled, it.target)
			}
		}
	}
	return fail.OK
}

// entry gives the [files] table of <target> of <tool>, or nil.
func (s *state) entry(tool, target string) *source.FileEntry {
	m := s.idx.Manifests[tool]
	if m == nil {
		return nil
	}
	for i := range m.Files {
		if m.Files[i].Target == target {
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
