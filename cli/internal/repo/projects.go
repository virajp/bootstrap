package repo

import (
	"encoding/json"
	"os"
	"slices"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/projects"
	"github.com/virajp/bootstrap/cli/internal/scripts"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// The checks of the projects of `bootstrap check`. Contract:
// docs/design/cli.md ("check").

// projectFolders examines the folder of each [projects.<name>] table of
// repo.toml: a folder that does not exist is a problem with code 2.
func (s *state) projectFolders() fail.Code {
	if s.vals == nil {
		return fail.OK
	}
	code := fail.OK
	for _, p := range projects.Listed(s.vals.Raw) {
		if !isDir(s.root + "/" + p.Folder) {
			s.log("[projects." + p.Name + "].folder = " + p.Folder + " in " + s.showRepo + " is not a folder of the repo - create it, or remove [projects." + p.Name + "] and its name from the projects of each tool")
			code = fail.Usage
		}
	}
	return code
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// linkResult is the part of the result of a detect script that the check of
// a link reads.
type linkResult struct {
	Result   string   `json:"result"`
	Evidence []string `json:"evidence"`
}

func (r linkResult) in() bool { return r.Result == "match" && projects.InFolder(r.Evidence) }

// projectLinks examines each name of the projects list of each tool, with
// the clone: its detect script still gives match in the folder of the
// project, or, for a tool of the categories languages and package-managers,
// it still matches at the root and a workspace manifest of its evidence
// still names the folder. A link that passes neither test is a warning. A
// root_only tool, and a project whose folder is not there, are skipped.
func (s *state) projectLinks() fail.Code {
	if s.idx == nil || s.vals == nil {
		return fail.OK
	}
	folderOf := map[string]string{}
	for _, p := range projects.Listed(s.vals.Raw) {
		folderOf[p.Name] = p.Folder
	}
	tools, ok := s.vals.Raw.Get("tools").(*source.Obj)
	if !ok {
		return fail.OK
	}
	r := s.runner()
	type link struct{ tool, name, folder string }
	var links []link
	var jobs []scripts.Job
	queued := map[string]bool{}
	repo := s.repoInput()
	queue := func(tool, dir string) {
		if queued[tool+"\x00"+dir] {
			return
		}
		queued[tool+"\x00"+dir] = true
		vals, _ := s.vals.User.Get(tool).(*source.Obj)
		in := scripts.Input{Tool: tool, Values: vals, Repo: repo}
		jobs = append(jobs, scripts.Job{Tool: tool, Name: scripts.Detect, Input: in.JSON(), Dir: dir})
	}
	for _, t := range tools.Keys() {
		m := s.idx.Manifests[t]
		if m == nil || m.RootOnly || !r.Has(t, scripts.Detect) {
			continue
		}
		for _, n := range projects.Links(s.vals.Raw, t) {
			f, ok := folderOf[n]
			if !ok || !isDir(s.root+"/"+f) {
				continue
			}
			links = append(links, link{t, n, f})
			queue(t, f)
			if projects.RootLinks(m.Category) {
				queue(t, projects.Root)
			}
		}
	}
	if len(links) == 0 {
		return fail.OK
	}
	got := map[string]linkResult{}
	code := fail.OK
	for i, res := range r.Run(jobs) {
		dir := jobs[i].Dir
		var lr linkResult
		if res.Code != 0 || json.Unmarshal(res.Stdout, &lr) != nil {
			s.log(r.ShownPath(res.Tool, res.Name) + " in " + dir + " gave no result - run bootstrap detect to see its error")
			code = fail.Usage
			continue
		}
		got[res.Tool+"\x00"+dir] = lr
	}
	var members map[string][]string
	membersOf := func(manifest string) []string {
		if members == nil {
			members = map[string][]string{}
			dirs, err := projects.Folders(s.root)
			if err != nil {
				dirs = map[string]bool{}
			}
			for _, m := range projects.Manifests {
				members[m], _ = projects.Members(s.root, m, dirs)
			}
		}
		return members[manifest]
	}
	for _, l := range links {
		if got[l.tool+"\x00"+l.folder].in() {
			continue
		}
		root := got[l.tool+"\x00"+projects.Root]
		ok := false
		if projects.RootLinks(s.idx.Manifests[l.tool].Category) && root.in() {
			for _, e := range root.Evidence {
				e = strings.TrimPrefix(e, "./")
				if projects.IsManifest(e) && slices.Contains(membersOf(e), l.folder) {
					ok = true
				}
			}
		}
		if !ok {
			s.r.Warn("[tools." + l.tool + "].projects in " + s.showTools + " lists " + l.name + ", but the detect script of " + l.tool + " gives no match in " + l.folder + " - remove " + l.name + " from the list, or run bootstrap detect")
		}
	}
	return code
}
