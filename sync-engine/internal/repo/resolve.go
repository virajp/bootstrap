package repo

import (
	"os"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/oldform"
	"github.com/virajp/bootstrap/sync-engine/internal/selection"
)

// resolve is `_rs_resolve`: the tools of the values file, packs expanded and
// dependencies included, and the data of their `[files]` tables. It gives 2
// or 3, after a message, on any problem.
func (s *state) resolve() fail.Code {
	sel, err := s.readSel()
	if err != nil {
		return s.logAll(err)
	}
	if err := oldform.Repo(s.root, s.vals.Raw, s.show); err != nil {
		return s.logAll(err)
	}
	if code := s.envRules(sel); code != fail.OK {
		return code
	}
	res, err := selection.ResolveRepo(s.idx, s.vals, s.inSubmodule())
	if err != nil {
		return s.logAll(err)
	}
	s.res = res
	s.fileModes = map[string]string{}
	s.fileFormats = map[string]string{}
	// The fields of each tool, in the order of the walk, as the done
	// callback of bash collects them.
	for _, tool := range res.Tools {
		m := s.idx.Manifests[tool]
		for _, fe := range m.Files {
			key := tool + "/" + fe.Target
			if fe.Mode != "generated" {
				s.fileModes[key] = fe.Mode
				s.spare = append(s.spare, fe.Target)
			}
			if fe.Format != "" {
				s.fileFormats[key] = fe.Format
			}
			for _, item := range fe.Requires {
				s.fileConds = append(s.fileConds, cond{key, "requires", item})
			}
			for _, item := range fe.RequiresFound {
				s.fileConds = append(s.fileConds, cond{key, "requires_found", item})
			}
		}
	}
	return s.formatterOf()
}

// formatterOf sets the formatter of the run: the [formatter] of the
// selected tool that has one. Two such tools stop the run.
func (s *state) formatterOf() fail.Code {
	var with []string
	s.formatter = nil
	for _, tool := range s.res.Tools {
		if f := s.idx.Manifests[tool].Formatter; f != nil {
			with = append(with, tool)
			s.formatter = f
		}
	}
	if len(with) > 1 {
		s.log("the selected tools " + strings.Join(with, ", ") + " each have a [formatter] - at most one selected tool can have one")
		return fail.Usage
	}
	return fail.OK
}

// fileSkips is `_rs_file_skips`: it sets reqSkip to the targets with a
// `requires` path that does not exist, or a `requires_found` name that
// hasFile does not find. The value is the first such item, in the order of
// the lists.
func (s *state) fileSkips() {
	s.reqSkip = map[string]string{}
	for _, c := range s.fileConds {
		if c.key == "" {
			continue
		}
		if _, done := s.reqSkip[c.key]; done {
			continue
		}
		if c.kind == "requires" {
			if exists(s.root + "/" + c.item) {
				continue
			}
		} else if s.hasFile(c.item) {
			continue
		}
		s.reqSkip[c.key] = c.item
	}
}

// inReqSkip tells if <key> has an entry in reqSkip.
func (s *state) inReqSkip(key string) bool { _, ok := s.reqSkip[key]; return ok }

// walk visits each entry below <dir>, without a symlink to a folder, until
// <visit> gives false.
func walk(dir string, visit func(p string, info os.FileInfo) bool) bool {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return true
	}
	for _, e := range ents {
		p := dir + "/" + e.Name()
		info, err := e.Info()
		if err != nil {
			continue
		}
		if !visit(p, info) {
			return false
		}
		if e.IsDir() && !walk(p, visit) {
			return false
		}
	}
	return true
}
