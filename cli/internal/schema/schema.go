// Package schema is the key schema of the values files of a repo in
// `.config/bootstrap/`: one table of keys for `repo.toml` and one for
// `releases.toml`, and the rule of a `tools.toml` key, which the tool
// catalog of the clone gives. `check`, the sync and each command that reads
// the values use it, so an unknown key, a wrong type or a value outside its
// enum stops them with code 2.
//
// Contract: docs/design/bootstrap-config.md.
package schema

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/projects"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// Format is the value of the key format of each TOML file of the folder.
const Format = 3

// Type is the type of a value of a key.
type Type string

// The types of a value.
const (
	String     Type = "a string"
	Integer    Type = "an integer"
	StringList Type = "a list of strings"
	Table      Type = "a table"
)

// Key is one key of a file. A path segment "*" is a name: it matches any
// key, for example `branches.*.landing`.
type Key struct {
	Path     string
	Type     Type
	Enum     []string // The values of a String key, or of an Integer key as text.
	Required bool
	// Sync tells that a sync reads the key: a change needs a sync.
	Sync bool
	// Default is the value of a key that is not set, or nil.
	Default any
}

// MemberName is the rule of a member name: the name of a `[members.<name>]`
// table is a flag (`--<name>`) and a part of a shell variable of the tasks.
var MemberName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Landing values of `[branches.<b>].landing`.
const (
	LandingPR           = "pr"
	LandingDirectMerge  = "direct-merge"
	LandingDirectCommit = "direct-commit"
)

// Repo is the schema of `repo.toml`.
var Repo = []Key{
	{Path: "format", Type: Integer, Enum: []string{fmt.Sprint(Format)}, Required: true},
	{Path: "repo_name", Type: String, Required: true, Sync: true},
	{Path: "product_name", Type: String, Sync: true},
	{Path: "repo_url", Type: String, Sync: true},
	{Path: "core_branches", Type: StringList, Required: true, Sync: true},
	{Path: "default_branch", Type: String},
	{Path: "branches", Type: Table, Sync: true},
	{Path: "branches.*", Type: Table, Sync: true},
	{Path: "branches.*.landing", Type: String, Enum: []string{LandingPR, LandingDirectMerge, LandingDirectCommit}, Sync: true, Default: LandingDirectMerge},
	{Path: "branches.*.merge_from", Type: StringList, Sync: true},
	{Path: "branches.*.merge_never", Type: StringList, Sync: true},
	{Path: "commits", Type: Table, Sync: true},
	{Path: "commits.style", Type: String, Enum: []string{"conventional"}, Sync: true, Default: "conventional"},
	{Path: "commits.scopes", Type: StringList, Sync: true},
	{Path: "environments", Type: StringList, Sync: true, Default: []any{"dev"}},
	{Path: "bootstrap_commit", Type: String},
	{Path: "worktree", Type: Table, Sync: true},
	{Path: "worktree.files", Type: StringList, Sync: true},
	{Path: "topology", Type: String, Enum: []string{"repo", "monorepo", "multi-repo"}},
	{Path: "topology_reason", Type: String},
	{Path: "members", Type: Table, Sync: true},
	{Path: "members.*", Type: Table, Sync: true},
	{Path: "members.*.folder", Type: String, Sync: true},
	{Path: "members.*.type", Type: String, Enum: []string{"submodule", "folder", "sibling"}, Required: true, Sync: true},
	{Path: "members.*.url", Type: String, Sync: true},
	{Path: "members.*.label", Type: String, Sync: true},
	{Path: "members.*.worktree_install", Type: String, Sync: true},
	{Path: "home", Type: Table, Sync: true},
	{Path: "home.folder", Type: String, Sync: true},
	{Path: "home.type", Type: String, Enum: []string{"submodule", "sibling"}, Sync: true},
	{Path: "home.url", Type: String, Sync: true},
	{Path: "projects", Type: Table},
	{Path: "projects.*", Type: Table},
	{Path: "projects.*.folder", Type: String},
	{Path: "release_env", Type: String},
	{Path: "deploy", Type: Table},
	{Path: "deploy.*", Type: Table},
	{Path: "deploy.*.*", Type: String},
}

// Releases is the schema of `releases.toml`: one release level for each
// project.
var Releases = []Key{
	{Path: "format", Type: Integer, Enum: []string{fmt.Sprint(Format)}, Required: true},
	{Path: "*", Type: String, Enum: []string{"NONE", "PATCH", "MINOR", "MAJOR"}, Default: "NONE"},
}

// Moved are the keys of a `[tools.<name>]` table that live in `repo.toml`
// now, with their new key. A `[tools.*]` or `[overrides.tools.*]` table
// cannot have them.
var Moved = map[string]string{
	"git.commit_scopes":   "commits.scopes",
	"git.branch":          "branches.<branch>",
	"mise.worktree_files": "worktree.files",
}

// Find gives the key of <keys> at the path <segs>, or nil. A literal segment
// wins over a name segment "*".
func Find(keys []Key, segs []string) *Key {
	var best *Key
	bestScore := -1
	for i := range keys {
		k := &keys[i]
		parts := strings.Split(k.Path, ".")
		if len(parts) != len(segs) {
			continue
		}
		score, ok := 0, true
		for j, p := range parts {
			switch p {
			case segs[j]:
				score++
			case "*":
			default:
				ok = false
			}
			if !ok {
				break
			}
		}
		if ok && score > bestScore {
			best, bestScore = k, score
		}
	}
	return best
}

func usage(format string, args ...any) *fail.Error { return fail.New(fail.Usage, format, args...) }

// Of tells if the value <v> has the type <t>.
func Of(t Type, v any) bool {
	switch t {
	case String:
		_, ok := v.(string)
		return ok
	case Integer:
		_, ok := v.(int64)
		return ok
	case Table:
		_, ok := v.(*source.Obj)
		return ok
	case StringList:
		l, ok := v.([]any)
		if !ok {
			return false
		}
		for _, x := range l {
			if _, ok := x.(string); !ok {
				return false
			}
		}
		return true
	}
	return false
}

// InEnum tells if the value <v> of the key <k> is one of its enum values. A
// key without an enum takes each value.
func (k *Key) InEnum(v any) bool {
	if len(k.Enum) == 0 {
		return true
	}
	text := fmt.Sprint(v)
	for _, e := range k.Enum {
		if e == text {
			return true
		}
	}
	return false
}

// check walks <raw> against <keys>: an unknown key, a wrong type and a
// value outside its enum are each a problem. <file> names the file kind in a
// message, <shown> its path.
func check(keys []Key, raw *source.Obj, file, shown string) []*fail.Error {
	var probs []*fail.Error
	var walk func(o *source.Obj, segs []string)
	walk = func(o *source.Obj, segs []string) {
		for _, name := range o.Keys() {
			path := append(append([]string(nil), segs...), name)
			key := strings.Join(path, ".")
			k := Find(keys, path)
			if k == nil {
				probs = append(probs, usage("%s in %s is not a key of %s", key, shown, file))
				continue
			}
			v := o.Get(name)
			if !Of(k.Type, v) {
				probs = append(probs, usage("%s in %s must be %s", key, shown, k.Type))
				continue
			}
			if !k.InEnum(v) {
				probs = append(probs, usage("%s = %s in %s is not one of %s", key, strings.TrimSpace(source.JSON(v)), shown, strings.Join(k.Enum, ", ")))
				continue
			}
			if t, ok := v.(*source.Obj); ok {
				walk(t, path)
			}
		}
	}
	walk(raw, nil)
	for _, k := range keys {
		if k.Required && !strings.Contains(k.Path, ".") && !raw.Has(k.Path) {
			probs = append(probs, usage("%s has no %s - it is a required key of %s", shown, k.Path, file))
		}
	}
	return probs
}

// CheckRepo examines the `repo.toml` <raw>, shown as <shown>: its keys and
// the rules between them. The error holds a message (code 2) for each
// problem.
func CheckRepo(raw *source.Obj, shown string) error {
	probs := check(Repo, raw, "repo.toml", shown)
	if len(probs) == 0 {
		probs = append(branchRules(raw, shown), memberRules(raw, shown)...)
		probs = append(probs, projectRules(raw, shown)...)
	}
	return source.Problems(probs)
}

// Landing gives the landing of the branch <b> of the `repo.toml` <raw>, or
// its default.
func Landing(raw *source.Obj, b string) string {
	if bs, ok := raw.Get("branches").(*source.Obj); ok {
		if t, ok := bs.Get(b).(*source.Obj); ok {
			if l, ok := t.Get("landing").(string); ok {
				return l
			}
		}
	}
	return LandingDirectMerge
}

// branchRules are the rules between the keys of the branches: each
// `[branches.<b>]` names a core branch, with main as direct-commit a core
// branch develop is direct-commit too, and default_branch is a core branch
// that a branch from it can land on.
func branchRules(raw *source.Obj, shown string) []*fail.Error {
	var probs []*fail.Error
	core := map[string]bool{}
	list, _ := raw.Get("core_branches").([]any)
	for _, x := range list {
		core[x.(string)] = true
	}
	if bs, ok := raw.Get("branches").(*source.Obj); ok {
		for _, b := range bs.Keys() {
			if !core[b] {
				probs = append(probs, usage("[branches.%s] in %s: %s is not in core_branches - add it there, or remove the table", b, shown, b))
			}
		}
	}
	if core["main"] && core["develop"] && Landing(raw, "main") == LandingDirectCommit && Landing(raw, "develop") != LandingDirectCommit {
		probs = append(probs, usage("[branches.main] in %s has landing = %q, but [branches.develop] does not - set landing = %q for develop too, or change main", shown, LandingDirectCommit, LandingDirectCommit))
	}
	if d, ok := raw.Get("default_branch").(string); ok {
		if !core[d] {
			probs = append(probs, usage("default_branch = %q in %s is not in core_branches - add it there, or change default_branch", d, shown))
		} else if d == "main" && core["develop"] && !branchHasList(raw, "main") {
			// The defaults of a merge into main with develop let only develop in,
			// so a branch from main, for example the sync branch, cannot land.
			probs = append(probs, usage("default_branch = \"main\" in %s, but only develop merges into main - set merge_from or merge_never in [branches.main], or change default_branch", shown))
		}
	}
	return probs
}

// branchHasList tells whether [branches.<b>] has merge_from or merge_never.
func branchHasList(raw *source.Obj, b string) bool {
	if bs, ok := raw.Get("branches").(*source.Obj); ok {
		if t, ok := bs.Get(b).(*source.Obj); ok {
			return t.Has("merge_from") || t.Has("merge_never")
		}
	}
	return false
}

// memberRules are the rules between the keys of the members and of
// [home]: each member has a type, a sibling member has a url, and [home]
// has a folder and a type.
// The folder of a sibling is relative to the root of this repo, so it may
// start with ../.
func memberRules(raw *source.Obj, shown string) []*fail.Error {
	var probs []*fail.Error
	if ms, ok := raw.Get("members").(*source.Obj); ok {
		for _, n := range ms.Keys() {
			if !MemberName.MatchString(n) {
				probs = append(probs, usage("[members.%q] in %s has a name that is not valid - a member name has only letters, digits, - and _", n, shown))
			}
			m, _ := ms.Get(n).(*source.Obj)
			if m == nil {
				continue
			}
			if !m.Has("type") {
				probs = append(probs, usage("[members.%s] in %s has no type - set type to submodule, folder or sibling", n, shown))
				continue
			}
			if t, _ := m.Get("type").(string); t == "sibling" && !m.Has("url") {
				probs = append(probs, usage("[members.%s] in %s has type sibling, but no url - add the url of its repo", n, shown))
			}
		}
	}
	if h, ok := raw.Get("home").(*source.Obj); ok {
		for _, k := range []string{"folder", "type"} {
			if !h.Has(k) {
				probs = append(probs, usage("[home] in %s has no %s - run bootstrap sync from the home repo, which writes [home]", shown, k))
			}
		}
	}
	return probs
}

// projectRules are the rules of the [projects.<name>] tables: a name is
// lowercase kebab case, and each table has a folder of the repo, relative to
// its root (. is the root), that no other project has.
func projectRules(raw *source.Obj, shown string) []*fail.Error {
	var probs []*fail.Error
	ps, ok := raw.Get("projects").(*source.Obj)
	if !ok {
		return nil
	}
	seen := map[string]string{}
	for _, n := range ps.Keys() {
		if !projects.NameRule.MatchString(n) {
			probs = append(probs, usage("[projects.%q] in %s has a name that is not valid - a project name is lowercase kebab case, for example %s", n, shown, projects.Normalize(n)))
		}
		e, _ := ps.Get(n).(*source.Obj)
		if e == nil {
			continue
		}
		f, ok := e.Get("folder").(string)
		if !ok {
			probs = append(probs, usage("[projects.%s] in %s has no folder - set folder to its folder, relative to the root of the repo (. is the root)", n, shown))
			continue
		}
		c, ok := projects.CleanFolder(f)
		if !ok {
			probs = append(probs, usage("[projects.%s].folder = %q in %s is not a folder of the repo - give it relative to the root of the repo (. is the root)", n, f, shown))
			continue
		}
		if other, dup := seen[c]; dup {
			probs = append(probs, usage("[projects.%s] and [projects.%s] in %s have the same folder %s - a folder is one project", other, n, shown, c))
		}
		seen[c] = n
	}
	return probs
}

// CheckProjects examines the projects list of each [tools.<name>] table of
// the tools.toml values <tools>, shown as <shown>, against the
// [projects.<name>] tables of the repo.toml values <repo>: the list is a
// list of strings, each one a project of repo.toml. With the catalog <idx>,
// a root_only tool has no projects. The error holds a message (code 2) for
// each problem.
func CheckProjects(idx *source.Index, repo, tools *source.Obj, shown string) error {
	t, ok := tools.Get("tools").(*source.Obj)
	if !ok {
		return nil
	}
	known := map[string]bool{}
	for _, p := range projects.Listed(repo) {
		known[p.Name] = true
	}
	var probs []*fail.Error
	for _, n := range t.Keys() {
		e, ok := t.Get(n).(*source.Obj)
		if !ok || !e.Has("projects") {
			continue
		}
		if !Of(StringList, e.Get("projects")) {
			probs = append(probs, usage("[tools.%s].projects in %s must be %s - the names of [projects.<name>] tables of repo.toml", n, shown, StringList))
			continue
		}
		if idx != nil {
			if m := idx.Manifests[n]; m != nil && m.RootOnly {
				probs = append(probs, usage("[tools.%s].projects in %s: %s is root_only, so it has no projects - remove the key", n, shown, n))
				continue
			}
		}
		for _, x := range e.Get("projects").([]any) {
			if p := x.(string); !known[p] {
				probs = append(probs, usage("[tools.%s].projects in %s names %s, which is not a [projects.<name>] table of repo.toml - add [projects.%s], or remove the name", n, shown, p, p))
			}
		}
	}
	return source.Problems(probs)
}

// CheckReleases examines the `releases.toml` <raw>, shown as <shown>.
func CheckReleases(raw *source.Obj, shown string) error {
	return source.Problems(check(Releases, raw, "releases.toml", shown))
}

// toolsKeys are the top-level keys of `tools.toml`.
const toolsKeys = "format, [tools.<name>] and [overrides.*]"

// CheckToolsShape examines the `tools.toml` <raw>, shown as <shown>,
// without the catalog: its top-level keys, the key format, the
// `[overrides.tools.<name>]` tables, and the keys that live in `repo.toml`
// now. The `[tools.<name>]` tables and the other `[overrides.*]` keys have
// their own checks. The error holds a message (code 2) for each problem.
func CheckToolsShape(raw *source.Obj, shown string) error {
	var probs []*fail.Error
	for _, k := range raw.Keys() {
		v := raw.Get(k)
		switch k {
		case "format":
			if n, ok := v.(int64); !ok || n != Format {
				probs = append(probs, usage("format = %s in %s is not valid - the format of a values file is %d", strings.TrimSpace(source.JSON(v)), shown, Format))
			}
		case "tools":
			if t, ok := v.(*source.Obj); ok {
				probs = append(probs, moved(t, "tools", shown)...)
			}
		case "overrides":
			o, ok := v.(*source.Obj)
			if !ok {
				probs = append(probs, usage("overrides in %s must be a table", shown))
				break
			}
			if !o.Has("tools") {
				break
			}
			t, ok := o.Get("tools").(*source.Obj)
			if !ok {
				probs = append(probs, usage("overrides.tools in %s must be [overrides.tools.<name>] tables", shown))
				break
			}
			for _, n := range t.Keys() {
				if _, ok := t.Get(n).(*source.Obj); !ok {
					probs = append(probs, usage("[overrides.tools.%s] in %s must be a table", n, shown))
				}
			}
			probs = append(probs, moved(t, "overrides.tools", shown)...)
		default:
			probs = append(probs, usage("%s is not a key of %s - the top-level keys are %s", k, shown, toolsKeys))
		}
	}
	if !raw.Has("format") {
		probs = append(probs, usage("%s has no format - it is a required key of tools.toml", shown))
	}
	return source.Problems(probs)
}

// moved gives a problem for each key of the tables <t> (`[<prefix>.<name>]`)
// that lives in `repo.toml` now.
func moved(t *source.Obj, prefix, shown string) []*fail.Error {
	var probs []*fail.Error
	for _, n := range t.Keys() {
		e, ok := t.Get(n).(*source.Obj)
		if !ok {
			continue
		}
		for _, k := range e.Keys() {
			if to, ok := Moved[n+"."+k]; ok {
				probs = append(probs, usage("[%s.%s].%s in %s is a key of repo.toml now - write it as %s in repo.toml", prefix, n, k, shown, to))
			}
		}
	}
	return probs
}

// ToolKey tells if <key> is a key of the `[tools.<tool>]` table of
// `tools.toml` by the catalog <idx>: env, files, or a question or a default
// of the tool, or projects. A tool that the catalog does not have gives
// false.
func ToolKey(idx *source.Index, tool, key string) bool {
	m := idx.Manifests[tool]
	if m == nil {
		return false
	}
	if key == "env" || key == "files" || key == "projects" {
		return true
	}
	if m.Raw == nil {
		return false
	}
	for _, table := range []string{"questions", "defaults"} {
		if t, ok := m.Raw.Get(table).(*source.Obj); ok && t.Has(key) {
			return true
		}
	}
	return false
}

// CheckTools examines each key of the `[tools.<name>]` tables of the
// `tools.toml` <raw>, shown as <shown>, against the catalog <idx>. A tool
// that the catalog does not have is left to the selection, which stops with
// code 3. The keys of `[overrides.tools.*]` keep their own rules. The error
// holds a message (code 2) for each problem.
func CheckTools(idx *source.Index, raw *source.Obj, shown string) error {
	t, ok := raw.Get("tools").(*source.Obj)
	if !ok {
		return nil
	}
	var probs []*fail.Error
	for _, n := range t.Keys() {
		e, ok := t.Get(n).(*source.Obj)
		if !ok || idx.Manifests[n] == nil {
			continue
		}
		for _, k := range e.Keys() {
			if _, isMoved := Moved[n+"."+k]; isMoved {
				continue // CheckToolsShape names it.
			}
			if !ToolKey(idx, n, k) {
				probs = append(probs, usage("[tools.%s].%s in %s is not a key of the tool %s - its keys are env, files, projects, and its questions and defaults", n, k, shown, n))
			}
		}
	}
	return source.Problems(probs)
}
