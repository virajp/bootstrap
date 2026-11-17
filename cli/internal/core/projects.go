package core

import (
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/projects"
	"github.com/virajp/bootstrap/cli/internal/scripts"
)

// The project pass of detect on the repo side. Contract:
// docs/design/cli.md ("detect").
//
// It finds the project folders, in order: the members of the workspace
// manifests at the root, then the folders below the root where the detect
// script of a tool of the category languages gives match, down to
// projects.MaxDepth folders. A folder inside a found project is not a new
// project, and a folder that repo.toml lists is never proposed again. With
// no project found, the root is one project named repo_name. It then runs
// each detect script in each project folder, and links a tool to each
// project where it matches.

// detectProjects runs the project pass: it sets Projects, the results of
// the scripts in the project folders, and, for a tool that is not selected
// and that matches only in a project folder, a match in Found with the
// evidence of those folders.
func (d *Detection) detectProjects() error {
	st := d.State
	root := st.Root
	d.rootFound = map[string]Detected{}
	for t, f := range d.Found {
		d.rootFound[t] = f
	}
	d.folder = map[string]map[string]Detected{}
	dirs, err := projects.Folders(root)
	if err != nil {
		d.Warnings = append(d.Warnings, "cannot list the folders of the repository with git: "+err.Error())
		dirs = map[string]bool{}
	}
	listed := projects.Listed(st.Raw)
	var found []string // The folders of the projects, for the nesting rule.
	listedAt := map[string]bool{}
	for _, p := range listed {
		listedAt[p.Folder] = true
		found = append(found, p.Folder)
	}
	inside := func(f string) bool {
		for _, p := range found {
			if projects.Inside(f, p) {
				return true
			}
		}
		return false
	}
	var fresh []string

	// 1. The members of the workspace manifests.
	d.members = map[string][]string{}
	seen := map[string]bool{}
	var members []string
	for _, m := range projects.Manifests {
		list, err := projects.Members(root, m, dirs)
		if err != nil {
			d.Warnings = append(d.Warnings, "cannot read the workspace manifest "+m+": "+err.Error())
			continue
		}
		d.members[m] = list
		for _, f := range list {
			if !seen[f] {
				seen[f] = true
				members = append(members, f)
			}
		}
	}
	byDepth(members)
	for _, f := range members {
		if !inside(f) {
			found = append(found, f)
			fresh = append(fresh, f)
		}
	}

	// 2. The walk below the root, one depth at a time.
	var lang []string
	for _, t := range d.Tools {
		if m := st.Manifest(t); m != nil && m.Category == projects.Languages {
			lang = append(lang, t)
		}
	}
	if len(lang) > 0 {
		all := make([]string, 0, len(dirs))
		for f := range dirs {
			all = append(all, f)
		}
		byDepth(all)
		for depth := 1; depth <= projects.MaxDepth; depth++ {
			var cands []string
			for _, f := range all {
				if projects.Depth(f) == depth && !inside(f) {
					cands = append(cands, f)
				}
			}
			if err := d.runIn(cands, lang); err != nil {
				return err
			}
			for _, f := range cands {
				for _, t := range lang {
					if r, ok := d.folder[f][t]; ok && matchIn(r) {
						found = append(found, f)
						fresh = append(fresh, f)
						break
					}
				}
			}
		}
	}

	// 3. With no project, the root is one project.
	used := map[string]bool{}
	for _, p := range listed {
		used[p.Name] = true
	}
	if len(found) == 0 {
		name, _ := st.Raw.Get("repo_name").(string)
		if name == "" {
			name = filepath.Base(root)
		}
		if n := projects.Normalize(name); n == "" {
			d.Warnings = append(d.Warnings, "the root is not proposed as a project: its name "+name+" has no letter or digit - add [projects.<name>] with folder = \".\" to repo.toml")
		} else if !used[n] {
			d.Projects = append(d.Projects, projects.Project{Name: n, Folder: projects.Root})
		}
	}

	// 4. The names: the basename, else the path with each / as -.
	count := map[string]int{}
	for _, f := range fresh {
		count[projects.Normalize(filepath.Base(f))]++
	}
	for _, f := range fresh {
		n := projects.Normalize(filepath.Base(f))
		if n == "" || count[n] > 1 || used[n] {
			n = projects.Normalize(strings.ReplaceAll(f, "/", "-"))
		}
		if n == "" || used[n] {
			d.Warnings = append(d.Warnings, "the folder "+f+" is not proposed as a project: it has no name of its own - add [projects.<name>] with folder = \""+f+"\" to repo.toml")
			continue
		}
		used[n] = true
		d.Projects = append(d.Projects, projects.Project{Name: n, Folder: f})
	}

	// The scripts of each tool in each project folder.
	var folders []string
	for _, p := range d.allProjects() {
		if p.Folder != projects.Root && dirs[p.Folder] {
			folders = append(folders, p.Folder)
		}
	}
	if err := d.runIn(folders, d.Tools); err != nil {
		return err
	}

	// A tool that is not selected and that matches only in a project
	// folder is offered too, with the evidence of those folders.
	for _, t := range d.Tools {
		if st.Selected(t) || d.Found[t].Result != "none" {
			continue
		}
		if ev := d.folderEvidence(t); len(ev) > 0 {
			msg := ""
			for _, p := range d.allProjects() {
				if r, ok := d.folder[p.Folder][t]; ok && matchIn(r) {
					msg = r.Message + " in " + p.Folder
					break
				}
			}
			d.Found[t] = Detected{Result: "match", Evidence: ev, Message: msg}
		}
	}
	return nil
}

// byDepth sorts the folders <fs> by depth, then by name.
func byDepth(fs []string) {
	sort.Slice(fs, func(i, j int) bool {
		a, b := projects.Depth(fs[i]), projects.Depth(fs[j])
		if a != b {
			return a < b
		}
		return fs[i] < fs[j]
	})
}

// matchIn tells if the result <r> of a script that ran in a folder counts
// there: a match with an evidence path in that folder itself.
func matchIn(r Detected) bool { return r.Result == "match" && projects.InFolder(r.Evidence) }

// runIn runs the detect script of each tool of <tools> in each folder of
// <folders> that has no result for it yet, and keeps the results.
func (d *Detection) runIn(folders, tools []string) error {
	r := d.Runner()
	var jobs []scripts.Job
	for _, f := range folders {
		for _, t := range tools {
			if _, done := d.folder[f][t]; done {
				continue
			}
			jobs = append(jobs, scripts.Job{Tool: t, Name: scripts.Detect, Input: d.Input(t, t), Dir: f})
		}
	}
	var errs []error
	for i, res := range r.Run(jobs) {
		f := jobs[i].Dir
		if !res.Ran {
			continue
		}
		got, err := Result(r, res, f)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if d.folder[f] == nil {
			d.folder[f] = map[string]Detected{}
		}
		d.folder[f][res.Tool] = got
	}
	return join(errs)
}

// allProjects gives the projects that repo.toml lists, then the proposed
// ones.
func (d *Detection) allProjects() []projects.Project {
	return append(projects.Listed(d.Raw), d.Projects...)
}

// folderEvidence gives the evidence of <tool> in each project folder where
// it matches, each path below its folder.
func (d *Detection) folderEvidence(tool string) []string {
	var out []string
	for _, p := range d.allProjects() {
		if p.Folder == projects.Root {
			continue
		}
		if r, ok := d.folder[p.Folder][tool]; ok && matchIn(r) {
			for _, e := range r.Evidence {
				out = append(out, p.Folder+"/"+strings.TrimPrefix(e, "./"))
			}
		}
	}
	return out
}

// links sets Links: for each selected or offered tool that is not
// root_only, the projects that it serves and that its projects list does
// not name yet.
//
//   - A match in a project folder links the tool to that project.
//   - The root project gets the link of a tool of the categories languages
//     and package-managers that matches at the root.
//   - A tool of those categories that matches at the root and in no other
//     project folder links to the projects of the workspace manifests in its
//     evidence; with none, to the root project, when there is one.
func (d *Detection) links() {
	st := d.State
	all := d.allProjects()
	for _, t := range d.Tools {
		if !st.Selected(t) && !slices.Contains(d.Offers, t) {
			continue
		}
		m := st.Manifest(t)
		if m == nil || m.RootOnly {
			continue
		}
		byRoot := projects.RootLinks(m.Category)
		atRoot := matchIn(d.rootFound[t])
		var names []string
		add := func(n string) {
			if !slices.Contains(names, n) {
				names = append(names, n)
			}
		}
		inProject := false
		for _, p := range all {
			if p.Folder == projects.Root {
				if byRoot && atRoot {
					add(p.Name)
				}
				continue
			}
			if r, ok := d.folder[p.Folder][t]; ok && matchIn(r) {
				add(p.Name)
				inProject = true
			}
		}
		if byRoot && atRoot && !inProject {
			for _, e := range d.rootFound[t].Evidence {
				e = strings.TrimPrefix(e, "./")
				if !projects.IsManifest(e) {
					continue
				}
				for _, p := range all {
					if slices.Contains(d.members[e], p.Folder) {
						add(p.Name)
					}
				}
			}
		}
		have := projects.Links(st.Raw, t)
		var adds []string
		for _, n := range names {
			if !slices.Contains(have, n) {
				adds = append(adds, n)
			}
		}
		if len(adds) > 0 {
			d.Links = append(d.Links, Link{Tool: t, Projects: adds})
		}
	}
}
