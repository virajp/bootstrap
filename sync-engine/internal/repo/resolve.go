package repo

import (
	"os"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/selection"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// resolve is `_rs_resolve`: the tools of the values file, packs expanded and
// dependencies included, and the data of their `[files]` tables. It gives 2
// or 3, after a message, on any problem.
func (s *state) resolve() fail.Code {
	sel, err := s.readSel()
	if err != nil {
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
	s.fileMerge = map[string][2]string{}
	var ext []string
	// The fields of each tool, in the order of the walk, as the done
	// callback of bash collects them.
	for _, tool := range res.Tools {
		m := s.idx.Manifests[tool]
		ext = append(ext, m.Extensions...)
		for _, fe := range m.Files {
			key := tool + "/" + fe.Target
			for _, k := range fe.Keys.Keys() {
				switch k {
				case "mode":
					mode := strings.ReplaceAll(source.JSON(fe.Keys.Get(k)), `"`, "")
					s.fileModes[key] = mode
					if mode != "generated" {
						s.spare = append(s.spare, fe.Target)
					}
				case "format":
					s.fileFormats[key] = strings.ReplaceAll(source.JSON(fe.Keys.Get(k)), `"`, "")
				case "remove":
					m := s.fileMerge[key]
					m[0] = source.JSON(fe.Keys.Get(k))
					s.fileMerge[key] = m
				case "remove_items":
					m := s.fileMerge[key]
					m[1] = source.JSON(fe.Keys.Get(k))
					s.fileMerge[key] = m
				}
			}
			for _, item := range fe.Requires {
				s.fileConds = append(s.fileConds, cond{key, "requires", item})
			}
			for _, item := range fe.RequiresFound {
				s.fileConds = append(s.fileConds, cond{key, "requires_found", item})
			}
		}
	}
	s.extensions = sortedUnique(ext)
	return fail.OK
}

// sortedUnique is `sort -u` of the non-empty items, in byte order.
func sortedUnique(in []string) []string {
	set := map[string]bool{}
	var out []string
	for _, x := range in {
		if x != "" && !set[x] {
			set[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}

// addExtensions is `_rs_extensions`: it adds the `extensions` list of the
// values file. In the root repository, it also adds the extensions of each
// git submodule: the tools of its values file and its own `extensions` list.
// VS Code opens the root only, so the root recommends the extensions of the
// whole tree.
func (s *state) addExtensions() {
	extra := ""
	if out, ok := s.list("extensions"); ok {
		extra = out
	}
	if !s.inSubmodule() && isFile(s.root+"/.gitmodules") {
		for _, sub := range strings.Split(s.submodules(), "\n") {
			subValues := s.root + "/" + sub + "/.config/bootstrap.toml"
			if !isFile(subValues) {
				subValues = s.root + "/" + sub + "/.config/repo-config.toml"
			}
			if sub == "" || !isFile(subValues) {
				continue
			}
			extra += "\n" + strings.Join(s.toolExtensions(subValues), "\n")
			own, _ := s.gomplate(listTemplate("extensions"), nil, []ds{{"v", subValues, "toml"}}, nil, true)
			extra += "\n" + strings.TrimRight(own, "\n")
		}
	}
	s.extensions = sortedUnique(append(s.extensions, strings.Split(extra, "\n")...))
}

// toolExtensions is `_rs_tool_extensions`: the VS Code extensions of the
// tools of the values file <file>, packs expanded and dependencies
// included. Nothing is checked: a problem in the values file of a member is
// for the own sync of that member to report.
func (s *state) toolExtensions(file string) []string {
	var sel []values.Item
	if v, err := values.ReadRepo(file, s.path(file)); err == nil {
		sel = v.Sel
	}
	var pending []string
	for _, e := range partialSelect(s.src, sel) {
		pending = append(pending, e.Tool)
	}
	var out []string
	seen := map[string]bool{}
	for len(pending) > 0 {
		tool := pending[0]
		pending = pending[1:]
		if tool == "" || seen[tool] || !s.src.HasManifest(tool) {
			continue
		}
		seen[tool] = true
		m := s.idx.Manifests[tool]
		if m == nil {
			continue
		}
		out = append(out, m.Extensions...)
		for _, d := range m.Depends {
			pending = append(pending, fields(d)...)
		}
		for _, mt := range m.Mise {
			_, names := selection.MiseDepends(mt)
			for _, name := range names {
				if owner := s.idx.MiseOwner[name]; owner != "" {
					pending = append(pending, owner)
				}
			}
		}
	}
	return out
}

// partialSelect is `_se_selected` as `_rs_tool_extensions` reads it: the
// entries that bash prints before a failure. A failed check of the packs
// prints nothing; an unknown pack in the values file stops after the
// entries of the packs before it.
func partialSelect(src source.Source, sel []values.Item) []selection.Entry {
	packs, err := src.ReadPacks()
	if err != nil || selection.CheckPacks(src, packs) != nil {
		return nil
	}
	out, err := selection.Expand(packs, selection.BasePack, "-", "default")
	if err != nil {
		return nil
	}
	for _, it := range sel {
		if it.Kind != "packs" {
			continue
		}
		more, err := selection.Expand(packs, it.Name, it.Envs, "packs."+it.Name)
		if err != nil {
			return out
		}
		out = append(out, more...)
	}
	for _, it := range sel {
		if it.Kind == "tools" {
			out = append(out, selection.Entry{Tool: it.Name, Envs: it.Envs, Src: "tools." + it.Name})
		}
	}
	return out
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

// anyTemplate tells if the folder <dir> has a regular `*.tmpl` file, at any
// depth (`find <dir> -name '*.tmpl' -type f -print -quit`).
func anyTemplate(dir string) bool {
	found := false
	walk(dir, func(p string, info os.FileInfo) bool {
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".tmpl") {
			found = true
			return false
		}
		return true
	})
	return found
}

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
