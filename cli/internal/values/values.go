// Package values reads the values files (the folder `.config/bootstrap/` of
// a repo and `macos.toml` of the Mac) as far as the engine reads them, and
// merges the values of the selected tools: defaults, then contributes, then
// the values file. Contract: docs/design/cli.md ("The values of a tool").
package values

import (
	"os"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/schema"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// The values of a repo are the files of the folder Folder, below the root of
// the repo. BOOTSTRAP_VALUES names another folder.
const (
	Folder       = ".config/bootstrap"
	RepoName     = "repo.toml"     // The repo-level facts.
	ToolsName    = "tools.toml"    // The tool selection, the tool values and the overrides.
	ReleasesName = "releases.toml" // The release level of each project.
	ManifestName = "repo-generated.tsv"
	// Format is the value of the key format of each TOML file of the folder.
	Format = schema.Format
)

// FolderOf gives the values folder of the repo <root>.
func FolderOf(root string) string { return strings.TrimSuffix(root, "/") + "/" + Folder }

// RepoFile gives the file repo.toml of the values folder <dir>.
func RepoFile(dir string) string { return dir + "/" + RepoName }

// ToolsFile gives the file tools.toml of the values folder <dir>.
func ToolsFile(dir string) string { return dir + "/" + ToolsName }

// ReleasesFile gives the file releases.toml of the values folder <dir>.
func ReleasesFile(dir string) string { return dir + "/" + ReleasesName }

// ManifestFile gives the sync manifest of the values folder <dir>.
func ManifestFile(dir string) string { return dir + "/" + ManifestName }

// Item is one `[tools.<name>]` or `[packs.<name>]` entry of a values file.
type Item struct {
	Kind string // "tools" or "packs".
	Name string
	// Envs is the `env` list joined with commas, or "-" when the entry has no
	// `env`. The Mac side always has "-".
	Envs string
	// Env is the `env` list. It is nil when the entry has no `env`.
	Env []string
}

// Repo is the values of a repo: the files repo.toml and tools.toml of its
// values folder.
type Repo struct {
	Path string // The values folder.
	// Raw holds the keys of repo.toml, and the [tools.*] and [overrides.*]
	// tables of tools.toml without [overrides.tools.*], with sorted keys.
	Raw *source.Obj
	// Sel holds the `[tools.*]` entries and then the `[packs.*]` entries, each
	// group in name order.
	Sel []Item
	// User holds each `[tools.<name>]` table without `env`, with the keys of
	// repo.toml that a tool reads under their old names (see compat).
	User *source.Obj
	// Over holds the `[overrides.tools.<name>]` tables of tools.toml, or nil.
	Over *source.Obj
}

func usage(format string, args ...any) *fail.Error { return fail.New(fail.Usage, format, args...) }

// ReadFolder reads the files repo.toml and tools.toml of the values folder
// <dir>, shown as <shown>, and examines them with the schema (repo.toml) and
// the shape rules (tools.toml). Each file must exist and have format = 3.
// The error holds a message (code 2) for each problem.
func ReadFolder(dir, shown string) (repo, tools *source.Obj, err error) {
	var probs []*fail.Error
	var raws [2]*source.Obj
	for i, name := range []string{RepoName, ToolsName} {
		p := dir + "/" + name
		if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
			probs = append(probs, usage("%s has no %s - run bootstrap init", shown, name))
			continue
		}
		raw, err := source.ReadTOML(p)
		if err != nil {
			probs = append(probs, usage("cannot read %s/%s", shown, name))
			continue
		}
		raws[i] = raw
	}
	if len(probs) > 0 {
		return nil, nil, source.Problems(probs)
	}
	repo, tools = raws[0], raws[1]
	repoShown, toolsShown := shown+"/"+RepoName, shown+"/"+ToolsName
	if _, ok := repo.Lookup("format"); !ok {
		probs = append(probs, usage("%s without format = %d is not valid - run bootstrap init", repoShown, Format))
	}
	if _, ok := tools.Lookup("format"); !ok {
		probs = append(probs, usage("%s without format = %d is not valid - run bootstrap init", toolsShown, Format))
	}
	if len(probs) > 0 {
		return nil, nil, source.Problems(probs)
	}
	err = source.Problems(append(source.Messages(schema.CheckRepo(repo, repoShown)), source.Messages(schema.CheckToolsShape(tools, toolsShown))...))
	if err != nil {
		return nil, nil, err
	}
	// The projects of the tools name the projects of repo.toml.
	if err := schema.CheckProjects(nil, repo, tools, toolsShown); err != nil {
		return nil, nil, err
	}
	return repo, tools, nil
}

// ReadRepo reads the values folder <dir> of a repo (see ReadFolder) and
// checks the shape of the `[tools]` and `[packs]` tables of tools.toml, as
// `_rs_values_sel` does. <shown> is the folder as a message names it. The
// error is the first problem, with code 2. The `env` names are not checked
// here: they need the environments of the repo.
func ReadRepo(dir, shown string) (*Repo, error) {
	repo, tools, err := ReadFolder(dir, shown)
	if err != nil {
		return nil, err
	}
	shown += "/" + ToolsName
	v := &Repo{Path: dir, Raw: Combine(repo, tools), User: source.NewObj()}
	raw := tools
	for _, kind := range []string{"tools", "packs"} {
		if !raw.Has(kind) {
			continue
		}
		table, ok := raw.Get(kind).(*source.Obj)
		if !ok {
			name := "security"
			if kind == "tools" {
				name = "node"
			}
			return nil, usage("%s in %s is a list, the old format - write one table for each entry, for example [%s.%s] env = [\"dev\"]; the pack base is always added", kind, shown, kind, name)
		}
		for _, name := range table.Keys() {
			e, ok := table.Get(name).(*source.Obj)
			if !ok {
				return nil, usage("[%s.%s] in %s: must be a table", kind, name, shown)
			}
			if kind == "packs" {
				for _, ek := range e.Keys() {
					if ek != "env" {
						return nil, usage("[%s.%s] in %s: '%s' is not a key of a pack entry; the only key of a pack entry is env, a list", kind, name, shown, ek)
					}
				}
			}
			it := Item{Kind: kind, Name: name, Envs: "-"}
			if e.Has("env") {
				l, ok := e.Get("env").([]any)
				if !ok {
					return nil, usage("[%s.%s] in %s: env must be a list, for example env = [\"dev\"]", kind, name, shown)
				}
				it.Env = make([]string, len(l))
				for i, x := range l {
					it.Env[i] = stringOf(x)
				}
				it.Envs = strings.Join(it.Env, ",")
			}
			v.Sel = append(v.Sel, it)
			if kind == "tools" {
				u := e.Clone()
				omit(u, EngineKeys...)
				v.User.Set(name, u)
			}
		}
	}
	if o, ok := tools.Get("overrides").(*source.Obj); ok {
		v.Over, _ = o.Get("tools").(*source.Obj)
	}
	compat(v.User, repo)
	return v, nil
}

// Combine gives the keys of the repo.toml <repo>, then the [tools.*] and
// [overrides.*] tables of the tools.toml <tools>, without
// [overrides.tools.*].
func Combine(repo, tools *source.Obj) *source.Obj {
	out := repo.Clone()
	if tools.Has("tools") {
		out.Set("tools", tools.Get("tools"))
	}
	if o, ok := tools.Get("overrides").(*source.Obj); ok {
		rest := source.NewObj()
		for _, k := range o.Keys() {
			if k != "tools" {
				rest.Set(k, o.Get(k))
			}
		}
		if rest.Len() > 0 {
			out.Set("overrides", rest)
		}
	} else if tools.Has("overrides") {
		out.Set("overrides", tools.Get("overrides"))
	}
	out.Sort()
	return out
}

// SubmoduleMember tells if the repo with the values <raw> (the keys of
// repo.toml) leaves out its root_only tools: its [home] has type submodule.
// A sibling member is its own root, so it keeps them. A repo without
// [home].type gives <git>, the git test of a submodule (D57, D79).
func SubmoduleMember(raw *source.Obj, git func() bool) bool {
	if raw == nil {
		return git()
	}
	if h, ok := raw.Get("home").(*source.Obj); ok {
		if t, ok := h.Get("type").(string); ok {
			return t == "submodule"
		}
	}
	return git()
}

// compat gives the tools the keys of repo.toml <repo> that they read in
// their tables before the values folder, under their old names, so their
// templates and scripts read them as before: [commits].scopes as
// [tools.git].commit_scopes, [branches.<b>] as [tools.git.branch.<b>] (with
// merge_model and direct_commit from landing), and [worktree].files as
// [tools.mise].worktree_files.
func compat(user, repo *source.Obj) {
	table := func(tool string) *source.Obj {
		t, ok := user.Get(tool).(*source.Obj)
		if !ok {
			t = source.NewObj()
			user.Set(tool, t)
			user.Sort()
		}
		return t
	}
	defer func() {
		for _, k := range user.Keys() {
			if t, ok := user.Get(k).(*source.Obj); ok {
				t.Sort()
			}
		}
	}()
	if c, ok := repo.Get("commits").(*source.Obj); ok && c.Has("scopes") {
		table("git").Set("commit_scopes", c.Get("scopes"))
	}
	if bs, ok := repo.Get("branches").(*source.Obj); ok && bs.Len() > 0 {
		out := source.NewObj()
		for _, b := range bs.Keys() {
			t, _ := bs.Get(b).(*source.Obj)
			e := source.NewObj()
			if t != nil {
				for _, k := range []string{"merge_from", "merge_never"} {
					if t.Has(k) {
						e.Set(k, t.Get(k))
					}
				}
				switch l, _ := t.Get("landing").(string); l {
				case schema.LandingPR:
					e.Set("merge_model", "pr")
				case schema.LandingDirectMerge:
					e.Set("merge_model", "direct")
				case schema.LandingDirectCommit:
					e.Set("merge_model", "direct")
					e.Set("direct_commit", true)
				}
			}
			e.Sort()
			out.Set(b, e)
		}
		table("git").Set("branch", out)
	}
	if w, ok := repo.Get("worktree").(*source.Obj); ok && w.Has("files") {
		table("mise").Set("worktree_files", w.Get("files"))
	}
}

// omit removes the keys <keys> from <o>.
func omit(o *source.Obj, keys ...string) {
	drop := map[string]bool{}
	for _, k := range keys {
		if o.Has(k) {
			drop[k] = true
		}
	}
	if len(drop) == 0 {
		return
	}
	c := source.NewObj()
	for _, k := range o.Keys() {
		if !drop[k] {
			c.Set(k, o.Get(k))
		}
	}
	*o = *c
}

func stringOf(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return strings.TrimSpace(source.JSON(v))
}

// Macos is the values file of the Mac, `macos.toml`.
type Macos struct {
	// Exists is false when the file is not there: the selection is empty.
	Exists bool
	Path   string
	// Sel holds the `[packs.*]` entries and then the `[tools.*]` entries, each
	// group in name order. Every Envs is "-".
	Sel []Item
	// User holds the `[tools.<name>]` tables.
	User *source.Obj
}

const macosKeys = "bootstrap_commit, format, [tools.<name>], [overrides.mise.<path>] and [mise.*]"

// ReadMacos reads the Mac values file <path> if it exists, as `_ss_read_values`
// does. <shown> is the path as a message names it. The error holds each message
// of bash in order; its code is 3 when a name is not in bootstrap, else 2.
func ReadMacos(path, shown string) (*Macos, error) {
	v := &Macos{Path: path, User: source.NewObj()}
	st, err := os.Stat(path)
	if err != nil {
		return v, nil
	}
	if !st.Mode().IsRegular() {
		return v, usage("%s is not a file", shown)
	}
	raw, err := source.ReadTOML(path)
	if err != nil {
		return v, usage("cannot read %s", shown)
	}
	v.Exists = true
	var probs []*fail.Error
	bad := func(format string, args ...any) { probs = append(probs, usage(format, args...)) }
	named := func(kind, name string) {
		if !source.NameOK(name) {
			probs = append(probs, fail.New(fail.Missing, "%s is not a bootstrap %s (named in %s)", name, kind, shown))
			return
		}
		v.Sel = append(v.Sel, Item{Kind: kind + "s", Name: name, Envs: "-"})
	}
	typeBad := func(name string) {
		switch name {
		case "packs":
			bad("packs in %s must be [packs.<name>] tables - the top-level keys are %s", shown, macosKeys)
		case "bootstrap_commit":
			bad("bootstrap_commit in %s must be a string", shown)
		default:
			bad("%s in %s must be a table - the top-level keys are %s", name, shown, macosKeys)
		}
	}
	for _, k := range raw.Keys() {
		x := raw.Get(k)
		switch k {
		case "bootstrap_commit":
			if _, ok := x.(string); !ok {
				typeBad(k)
			}
		case "format":
			// The package oldform examines its value.
		case "tools":
			t, ok := x.(*source.Obj)
			if !ok {
				bad("tools = [...] is the old format - write a [tools.<name>] table for each tool")
				break
			}
			for _, n := range t.Keys() {
				e, ok := t.Get(n).(*source.Obj)
				if !ok {
					bad("[tools.%s] in %s must be a table", n, shown)
					continue
				}
				named("tool", n)
				if e.Has("env") {
					bad("[tools.%s] env is for repo tools only - the Mac side has no environments", n)
				}
			}
		case "packs":
			t, ok := x.(*source.Obj)
			if !ok {
				typeBad(k)
				break
			}
			for _, n := range t.Keys() {
				e, ok := t.Get(n).(*source.Obj)
				if !ok {
					bad("[packs.%s] in %s must be an empty table", n, shown)
					continue
				}
				named("pack", n)
				for _, ek := range e.Keys() {
					bad("[packs.%s] in %s has the key '%s' - a pack entry is an empty table", n, shown, ek)
				}
			}
		case "overrides":
			t, ok := x.(*source.Obj)
			if !ok {
				typeBad(k)
				break
			}
			for _, ok := range t.Keys() {
				if ok == "mise_env" {
					bad("[overrides.mise_env] in %s: the Mac side has no environments - use [overrides.mise.<path>]", shown)
				} else if ok != "mise" {
					bad("[overrides.%s] in %s is not valid - write [tools.<name>] for template values and [overrides.mise.<path>] for mise config", ok, shown)
				}
			}
		case "mise":
			if _, ok := x.(*source.Obj); !ok {
				typeBad(k)
			}
		default:
			bad("%s is not a key of %s - the top-level keys are %s", k, shown, macosKeys)
		}
	}
	if len(probs) > 0 {
		return v, source.Problems(probs)
	}
	if t, ok := raw.Get("tools").(*source.Obj); ok {
		v.User = t
	}
	return v, nil
}

// Names gives the names of <items> of <kind>, in the order of the items.
func Names(items []Item, kind string) []string {
	var out []string
	for _, it := range items {
		if it.Kind == kind {
			out = append(out, it.Name)
		}
	}
	return out
}

// OverridesPath gives the overrides file of the Mac values file <path>: the
// same folder and name, with `.overrides.toml` in place of `.toml`, for
// example `macos.overrides.toml`. A repo has its overrides in tools.toml.
func OverridesPath(path string) string {
	return strings.TrimSuffix(path, ".toml") + ".overrides.toml"
}

// ReadOverrides reads the overrides file <path>, when it exists, and gives
// its `[tools.<name>]` tables. It wins over all the other values. Its only
// key is `tools`. <shown> is the path as a message names it.
func ReadOverrides(path, shown string) (*source.Obj, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, nil
	}
	raw, err := source.ReadTOML(path)
	if err != nil {
		return nil, usage("cannot read %s", shown)
	}
	for _, k := range raw.Keys() {
		if k != "tools" {
			return nil, usage("%s is not a key of %s - its only key is [tools.<name>]", k, shown)
		}
	}
	t, ok := raw.Get("tools").(*source.Obj)
	if !ok {
		if raw.Has("tools") {
			return nil, usage("tools in %s must be [tools.<name>] tables", shown)
		}
		return nil, nil
	}
	for _, n := range t.Keys() {
		if _, ok := t.Get(n).(*source.Obj); !ok {
			return nil, usage("[tools.%s] in %s must be a table", n, shown)
		}
	}
	return t, nil
}
