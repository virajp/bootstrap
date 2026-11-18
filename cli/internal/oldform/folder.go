package oldform

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/schema"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// The files of the values of a repo before the folder `.config/bootstrap/`.
const (
	OldValues    = ".config/bootstrap.toml"
	OldOverrides = ".config/bootstrap.overrides.toml"
	OldManifest  = ".config/repo-generated.tsv"
)

// FolderHint is the fix of a repo with both the old values file and the
// values folder.
const FolderHint = "remove one: compare .config/bootstrap.toml with .config/bootstrap/"

// toolsKeys are the top-level keys of an old values file that stay in
// tools.toml. Each other top-level key moves to repo.toml as it is, and the
// schema of repo.toml then names a key that it does not have.
var toolsKeys = map[string]bool{"tools": true, "overrides": true}

// Split is an old values file split into the files of the values folder.
type Split struct {
	Repo, Tools []byte // repo.toml and tools.toml, each with format = 3.
	Notes       []string
}

// SplitRepo splits the values file <values> of the form format = 2 (the
// text <in>), and its overrides file <overrides> when it exists, into
// repo.toml and tools.toml with format = 3, so the comments of the keys that
// stay in place, and of most moved keys, stay (D4, D5, D11, D16, D17, D18,
// D32, D47). <exists> tells if a branch exists in git, locally or on origin,
// and <submodule> tells if a folder is the path of a git submodule in
// .gitmodules.
func SplitRepo(in []byte, values, overrides string, exists, submodule func(string) bool) (*Split, error) {
	dir, err := os.MkdirTemp("", "bootstrap-split.")
	if err != nil {
		return nil, fail.New(fail.Other, "cannot create a temp folder")
	}
	defer func() { _ = os.RemoveAll(dir) }()
	p := filepath.Join(dir, "values.toml")
	if err := os.WriteFile(p, in, 0o600); err != nil {
		return nil, fail.New(fail.Other, "cannot write a temp file")
	}
	raw, err := source.ReadTOML(p)
	if err != nil {
		return nil, usage("cannot read %s", values)
	}
	var over *source.Obj
	if isFile(overrides) {
		o, err := source.ReadTOML(overrides)
		if err != nil {
			return nil, usage("cannot read %s", overrides)
		}
		over, _ = o.Get("tools").(*source.Obj)
	}
	s := &Split{}

	// The keys of the overrides file that live in repo.toml now go into the
	// values first, so they win over the values of the file (D18).
	var fold []string
	if over != nil {
		if g, ok := over.Get("git").(*source.Obj); ok {
			if g.Has("commit_scopes") {
				fold = append(fold, ".tools.git.commit_scopes = "+source.JSON(g.Get("commit_scopes")))
			}
			if g.Has("branch") {
				fold = append(fold, ".tools.git.branch = ((.tools.git.branch // {}) * "+source.JSON(g.Get("branch"))+")")
			}
		}
		if m, ok := over.Get("mise").(*source.Obj); ok && m.Has("worktree_files") {
			fold = append(fold, ".tools.mise.worktree_files = "+source.JSON(m.Get("worktree_files")))
		}
		if len(fold) > 0 {
			if in, err = editTOML(in, fold); err != nil {
				return nil, err
			}
			if err := os.WriteFile(p, in, 0o600); err != nil {
				return nil, fail.New(fail.Other, "cannot write a temp file")
			}
			if raw, err = source.ReadTOML(p); err != nil {
				return nil, usage("cannot read %s", values)
			}
			s.Notes = append(s.Notes, "the keys of "+OldOverrides+" that live in repo.toml go there")
		}
	}

	// The branches: each table names main or develop, a branch that exists
	// (D47); the core branches are the main and develop that exist (D16).
	var branches *source.Obj
	if t, ok := raw.Get("tools").(*source.Obj); ok {
		if g, ok := t.Get("git").(*source.Obj); ok {
			branches, _ = g.Get("branch").(*source.Obj)
		}
	}
	var probs []*fail.Error
	if branches != nil {
		for _, b := range branches.Keys() {
			if (b != "main" && b != "develop") || !exists(b) {
				probs = append(probs, usage("[tools.git.branch.%s] in %s names a branch that is not main or develop, or that does not exist in git - remove the table or create the branch", b, values))
			}
		}
	}
	if len(probs) > 0 {
		return nil, source.Problems(probs)
	}
	var core []any
	for _, b := range []string{"main", "develop"} {
		if exists(b) {
			core = append(core, b)
		}
	}
	if len(core) == 0 {
		return nil, usage("the repository has no branch main or develop, so %s cannot migrate - create main or develop, then run bootstrap init again", values)
	}

	// repo.toml: the old file without its tool tables, with the moved keys.
	exprs := []string{".core_branches = " + source.JSON(core)}
	if branches != nil {
		exprs = append(exprs, ".branches = .tools.git.branch")
		for _, b := range branches.Keys() {
			t, _ := branches.Get(b).(*source.Obj)
			exprs = append(exprs,
				".branches["+q(b)+"].landing = "+q(Landing(t)),
				"del(.branches["+q(b)+"].merge_model)",
				"del(.branches["+q(b)+"].direct_commit)")
		}
		s.Notes = append(s.Notes, "[tools.git.branch.*] to [branches.*] with landing")
	}
	if hasKey(raw, "git", "commit_scopes") {
		exprs = append(exprs, ".commits.scopes = .tools.git.commit_scopes")
		s.Notes = append(s.Notes, "[tools.git].commit_scopes to [commits].scopes")
	}
	exprs = append(exprs, `.commits.style = "conventional"`)
	if hasKey(raw, "mise", "worktree_files") {
		exprs = append(exprs, ".worktree.files = .tools.mise.worktree_files")
		s.Notes = append(s.Notes, "[tools.mise].worktree_files to [worktree].files")
	}
	exprs = append(exprs, membersExprs(raw, submodule)...)
	exprs = append(exprs, "del(.tools)", "del(.overrides)", "."+FormatKey+" = "+itoa(schema.Format))
	if s.Repo, err = editTOML(in, exprs); err != nil {
		return nil, err
	}
	s.Repo = formatFirst(s.Repo)

	// tools.toml: the tool tables and the overrides of the old file, the
	// overrides file as [overrides.tools.*] (D5).
	exprs = nil
	for _, k := range raw.Keys() {
		if !toolsKeys[k] && k != FormatKey {
			exprs = append(exprs, "del(.["+q(k)+"])")
		}
	}
	for _, mv := range sortedMoved() {
		tool, key, _ := strings.Cut(mv, ".")
		if hasKey(raw, tool, key) {
			exprs = append(exprs, "del(.tools["+q(tool)+"]["+q(key)+"])")
		}
	}
	exprs = append(exprs, "."+FormatKey+" = "+itoa(schema.Format))
	if over != nil {
		rest := source.NewObj()
		for _, n := range over.Keys() {
			e, ok := over.Get(n).(*source.Obj)
			if !ok {
				continue
			}
			kept := source.NewObj()
			for _, k := range e.Keys() {
				if _, isMoved := schema.Moved[n+"."+k]; !isMoved {
					kept.Set(k, e.Get(k))
				}
			}
			if kept.Len() > 0 {
				rest.Set(n, kept)
			}
		}
		if rest.Len() > 0 {
			exprs = append(exprs, ".overrides.tools = "+source.JSON(rest))
			s.Notes = append(s.Notes, OldOverrides+" to [overrides.tools.*]")
		}
	}
	if s.Tools, err = editTOML(in, exprs); err != nil {
		return nil, err
	}
	return s, nil
}

// membersExprs moves the [member.<name>] tables of the old file <raw> to
// [members.<name>] of repo.toml, with their keys as they are. A table
// without a type gets the type the old engine gave it (D99): submodule when
// its folder (default the name) is a git submodule in .gitmodules
// (<submodule>), else folder. The migration reads no vwf file: a sibling
// member, with its url, is added later with bootstrap config set.
func membersExprs(raw *source.Obj, submodule func(string) bool) []string {
	ms, ok := raw.Get("member").(*source.Obj)
	if !ok {
		return nil
	}
	out := []string{".members = .member", "del(.member)"}
	for _, n := range ms.Keys() {
		m, ok := ms.Get(n).(*source.Obj)
		if !ok || m.Has("type") {
			continue
		}
		folder := n
		if f, ok := m.Get("folder").(string); ok {
			folder = f
		}
		kind := "folder"
		if submodule(strings.TrimSuffix(folder, "/")) {
			kind = "submodule"
		}
		out = append(out, ".members["+q(n)+"].type = "+q(kind))
	}
	return out
}

// Landing gives the landing of an old [tools.git.branch.<b>] table <t>
// (D11, D17): merge_model = "pr" is pr; direct_commit = true with
// merge_model = "direct" or no merge_model is direct-commit; each other
// table is direct-merge.
func Landing(t *source.Obj) string {
	if t == nil {
		return schema.LandingDirectMerge
	}
	model, _ := t.Get("merge_model").(string)
	dc, _ := t.Get("direct_commit").(bool)
	switch {
	case model == "pr":
		return schema.LandingPR
	case (model == "direct" || model == "") && dc:
		return schema.LandingDirectCommit
	}
	return schema.LandingDirectMerge
}

// hasKey tells if the old file <raw> has [tools.<tool>].<key>.
func hasKey(raw *source.Obj, tool, key string) bool {
	t, ok := raw.Get("tools").(*source.Obj)
	if !ok {
		return false
	}
	e, ok := t.Get(tool).(*source.Obj)
	return ok && e.Has(key)
}

// sortedMoved gives the keys of schema.Moved, sorted.
func sortedMoved() []string {
	out := make([]string, 0, len(schema.Moved))
	for k := range schema.Moved {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func itoa(n int) string {
	return strings.TrimSpace(source.JSON(int64(n)))
}
