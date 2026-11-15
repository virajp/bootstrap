package repo

import (
	"path/filepath"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// homeKeys are the keys of [home], in the order of the table.
var homeKeys = []string{"folder", "type", "url"}

// homeLink gives the [home] table of this repo when the sync runs in a
// member: BOOTSTRAP_HOME names the home repo and BOOTSTRAP_MEMBER the name
// of this member in its [members.*] tables (D63, D89). [home] has folder,
// the path from this repo back to the home, type, from the table of the
// member in the home, and url, the repo_url of the home when it has one
// (D56, D57). A member without a type there is a submodule when this repo
// is a git submodule, else a folder member. A folder member gets no [home]
// (D55). Outside a member, or when the home has no repo.toml, it gives nil.
func (s *state) homeLink() *source.Obj {
	home, name := s.getenv("BOOTSTRAP_HOME"), s.getenv("BOOTSTRAP_MEMBER")
	if home == "" || name == "" {
		return nil
	}
	if !filepath.IsAbs(home) {
		home = filepath.Join(s.pwd, home)
	}
	file := values.RepoFile(values.FolderOf(home))
	if !isFile(file) {
		s.r.Warn("BOOTSTRAP_HOME " + home + " has no " + values.Folder + "/" + values.RepoName + " - [home] was not written")
		return nil
	}
	raw, err := source.ReadTOML(file)
	if err != nil {
		s.r.Warn("cannot read " + s.path(file) + " - [home] was not written")
		return nil
	}
	kind := ""
	if ms, ok := raw.Get("members").(*source.Obj); ok {
		if m, ok := ms.Get(name).(*source.Obj); ok {
			kind, _ = m.Get("type").(string)
		}
	}
	if kind == "" {
		kind = "folder"
		if s.inSubmodule() {
			kind = "submodule"
		}
	}
	if kind == "folder" {
		return nil
	}
	root := s.root
	if root == "" {
		root = s.pwd
	}
	folder, err := filepath.Rel(realPath(root), realPath(home))
	if err != nil {
		return nil
	}
	h := source.NewObj()
	h.Set("folder", filepath.ToSlash(folder))
	h.Set("type", kind)
	if u, ok := raw.Get("repo_url").(string); ok && u != "" {
		h.Set("url", u)
	}
	return h
}

// realPath gives <p> with its symlinks resolved, or <p> when it cannot.
func realPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

// sameHome tells if the [home] table <cur> has the keys and values of <want>.
func sameHome(cur, want *source.Obj) bool {
	if cur == nil || cur.Len() != want.Len() {
		return false
	}
	for _, k := range homeKeys {
		if cur.Get(k) != want.Get(k) {
			return false
		}
	}
	return true
}

// writeHome writes [home] into repo.toml when the sync runs in a member and
// [home] is not there or differs (D30, D63). A dry run plans the file.
func (s *state) writeHome() fail.Code {
	if s.home == nil {
		return fail.OK
	}
	cur, _ := s.vals.Raw.Get("home").(*source.Obj)
	if sameHome(cur, s.home) {
		return fail.OK
	}
	return s.e.SetTable(values.RepoFile(s.values), "home", strings.TrimSpace(source.JSON(s.home)))
}

// asSubmodule tells if this repo leaves out its root_only tools: [home] has
// type submodule, the [home] that this run writes first (D57, D79).
func (s *state) asSubmodule() bool {
	if s.home != nil {
		return s.home.Get("type") == "submodule"
	}
	var raw *source.Obj
	if s.vals != nil {
		raw = s.vals.Raw
	}
	return values.SubmoduleMember(raw, s.inSubmodule)
}
