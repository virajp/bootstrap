// Package oldform finds the old forms of a values file and of the files that
// went with it, and migrates them to the new form. `sync` and `check` stop
// with code 2 on an old form; only `bootstrap init` (`bootstrap mac init` on
// the Mac) migrates it.
//
// This package is the one place of the old names: the old tools of the entry
// and the old files of the own hooks and of the .gitignore lines.
//
// Contract: docs/design/bootstrap-config.md.
package oldform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tailscale/hujson"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// The old names.
const (
	// OldRepoTool is the old tool of the repo entry. The tool mise has the
	// entry now.
	OldRepoTool = "repo-sync"
	// OldMacTool is the old tool of the Mac entry. The Mac tool mise has the
	// entry now.
	OldMacTool = "macos-setup"
	// OldPrecommit is the old file of the own hooks of a repository.
	OldPrecommit = ".config/repo-sync/pre-commit.yaml"
	// OldGitignore is the older file of the own .gitignore lines.
	OldGitignore = ".config/repo-sync/gitignore"
	// precommit is the shared file that gets the own hooks.
	precommit  = ".config/pre-commit-config.yaml"
	extensions = ".vscode/extensions.json"
	gitignore  = ".gitignore"
)

// movedKeys are the old top-level keys of a repo values file.
var movedKeys = []string{"branch", "claude", "commit_scopes", "extensions", "worktree_files"}

// Migrating is true while `bootstrap init` plans the sync after a migration
// that it did not write yet: the old file of the own hooks is then not an
// error, because the migration removes it.
var Migrating bool

// Hint ends each message of an old form of the repo side.
const Hint = "run bootstrap init"

// MacHint ends each message of an old form of the Mac side.
const MacHint = "run bootstrap mac init"

func usage(format string, args ...any) *fail.Error { return fail.New(fail.Usage, format, args...) }

// Repo gives a message, with code 2, for each old form of the repository
// <root> with the values file <raw>, shown as <shown>. <raw> can be nil.
func Repo(root string, raw *source.Obj, shown string) error {
	var probs []*fail.Error
	if raw != nil {
		for _, k := range movedKeys {
			if raw.Has(k) {
				probs = append(probs, usage("%s in %s is the old form - %s", k, shown, Hint))
			}
		}
		if t, ok := raw.Get("tools").(*source.Obj); ok && t.Has(OldRepoTool) {
			probs = append(probs, usage("[tools.%s] in %s is the old form - %s", OldRepoTool, shown, Hint))
		}
		if o, ok := raw.Get("overrides").(*source.Obj); ok && o.Has("gitignore") {
			probs = append(probs, usage("[overrides.gitignore] in %s is the old form - %s", shown, Hint))
		}
		probs = append(probs, formatProblems(raw, shown, Hint)...)
	}
	if !Migrating && isFile(root+"/"+OldPrecommit) {
		probs = append(probs, usage("%s is the old form - %s", OldPrecommit, Hint))
	}
	if isFile(root + "/" + OldGitignore) {
		probs = append(probs, usage("%s is no longer used - move its lines into .gitignore, then delete it (M18 of docs/sync/migrate.md)", OldGitignore))
	}
	if len(probs) == 0 {
		return nil
	}
	return source.Problems(probs)
}

// Mac gives a message, with code 2, for each old form of the Mac values file
// <raw>, shown as <shown>. <raw> can be nil.
func Mac(raw *source.Obj, shown string) error {
	if raw == nil {
		return nil
	}
	var probs []*fail.Error
	if t, ok := raw.Get("tools").(*source.Obj); ok && t.Has(OldMacTool) {
		probs = append(probs, usage("[tools.%s] in %s is the old form - %s", OldMacTool, shown, MacHint))
	}
	probs = append(probs, formatProblems(raw, shown, MacHint)...)
	if len(probs) == 0 {
		return nil
	}
	return source.Problems(probs)
}

// Change is one file of a migration: its new content, or its removal.
type Change struct {
	Path    string // Absolute.
	Content []byte
	Remove  bool
	Note    string // What changes, for the change list.
}

// q is a string as a yq string literal.
func q(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// editTOML runs the yq expressions <exprs> on the TOML text <in>, so its
// comments stay.
func editTOML(in []byte, exprs []string) ([]byte, error) {
	dir, err := os.MkdirTemp("", "bootstrap-migrate.")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	p := filepath.Join(dir, "values.toml")
	if err := os.WriteFile(p, in, 0o600); err != nil {
		return nil, err
	}
	out, stderr, err := format.YQ("-p", "toml", "-o", "toml", strings.Join(exprs, " | "), p)
	if err != nil {
		return nil, usage("cannot migrate the values file: %s", strings.TrimSpace(stderr))
	}
	return []byte(out), nil
}

// strings of a TOML list, or nil.
func strList(v any) []string {
	l, _ := v.([]any)
	var out []string
	for _, x := range l {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// MigrateRepo gives the changes that move the repository <root> with the
// values file <values> to the new form. No change: an empty list. The tables
// of the packs need the clone of BOOTSTRAP_CLONE (see tablesOf).
func MigrateRepo(root, values string) ([]Change, error) {
	in, err := os.ReadFile(values)
	if err != nil {
		return nil, usage("cannot read %s", values)
	}
	raw, err := source.ReadTOML(values)
	if err != nil {
		return nil, usage("cannot read %s", values)
	}
	// The tables of the packs first: a later move adds keys to a table.
	exprs, notes, err := cloneTables(raw, values, source.Repo, root)
	if err != nil {
		return nil, err
	}
	var out []Change
	if raw.Has("commit_scopes") {
		exprs = append(exprs, ".tools.git.commit_scopes = .commit_scopes", "del(.commit_scopes)")
		notes = append(notes, "commit_scopes to [tools.git]")
	}
	if raw.Has("branch") {
		exprs = append(exprs, ".tools.git.branch = ((.tools.git.branch // {}) * .branch)", "del(.branch)")
		notes = append(notes, "[branch.*] to [tools.git.branch.*]")
	}
	if raw.Has("worktree_files") {
		exprs = append(exprs, ".tools.mise.worktree_files = .worktree_files", "del(.worktree_files)")
		notes = append(notes, "worktree_files to [tools.mise]")
	}
	if c, ok := raw.Get("claude").(*source.Obj); ok {
		if c.Has("plugins") {
			exprs = append(exprs, ".tools.claude.plugins = ((.tools.claude.plugins // {}) * .claude.plugins)", "del(.claude.plugins)")
		}
		exprs = append(exprs, "del(.claude)")
		notes = append(notes, "[claude.plugins.*] to [tools.claude.plugins.*]")
	}
	if t, ok := raw.Get("tools").(*source.Obj); ok && t.Has(OldRepoTool) {
		exprs = append(exprs, "del(.tools["+q(OldRepoTool)+"])")
		notes = append(notes, "[tools."+OldRepoTool+"] removed")
	}
	if o, ok := raw.Get("overrides").(*source.Obj); ok && o.Has("gitignore") {
		g, _ := o.Get("gitignore").(*source.Obj)
		if g != nil {
			for _, line := range strList(g.Get("remove")) {
				exprs = append(exprs, `.tools.git.files[".gitignore"][`+q(line)+`] = {"$mode": "remove"}`)
			}
			if add := strList(g.Get("add")); len(add) > 0 {
				c, err := addLines(root+"/"+gitignore, add)
				if err != nil {
					return nil, err
				}
				if c != nil {
					out = append(out, *c)
				}
			}
		}
		if o.Len() == 1 {
			exprs = append(exprs, "del(.overrides)")
		} else {
			exprs = append(exprs, "del(.overrides.gitignore)")
		}
		notes = append(notes, "[overrides.gitignore] to .gitignore and [tools.git.files.\".gitignore\"]")
	}
	if raw.Has("extensions") {
		c, err := addExtensions(root+"/"+extensions, strList(raw.Get("extensions")))
		if err != nil {
			return nil, err
		}
		if c != nil {
			out = append(out, *c)
		}
		exprs = append(exprs, "del(.extensions)")
		notes = append(notes, "extensions to "+extensions)
	}
	if isFile(root + "/" + OldPrecommit) {
		c, err := addHooks(root+"/"+precommit, root+"/"+OldPrecommit)
		if err != nil {
			return nil, err
		}
		out = append(out, c, Change{Path: root + "/" + OldPrecommit, Remove: true, Note: "its hooks move to " + precommit})
	}
	if len(exprs) > 0 {
		next, err := editTOML(in, exprs)
		if err != nil {
			return nil, err
		}
		out = append([]Change{{Path: values, Content: next, Note: "migrate the old form: " + strings.Join(notes, "; ")}}, out...)
	}
	return out, nil
}

// MigrateMac gives the change that moves the Mac values file <values> to the
// new form, or nil. The tables of the packs need the clone of BOOTSTRAP_CLONE
// (see tablesOf).
func MigrateMac(values string) (*Change, error) {
	raw, err := source.ReadTOML(values)
	if err != nil {
		return nil, usage("cannot read %s", values)
	}
	// The tables of the packs first: the move of the old tool adds keys to a
	// table.
	exprs, notes, err := cloneTables(raw, values, source.Macos, "")
	if err != nil {
		return nil, err
	}
	if t, ok := raw.Get("tools").(*source.Obj); ok && t.Has(OldMacTool) {
		exprs = append(exprs,
			".tools.mise = ((.tools.mise // {}) * .tools["+q(OldMacTool)+"])",
			"del(.tools["+q(OldMacTool)+"])")
		notes = append(notes, "[tools."+OldMacTool+"] to [tools.mise]")
	}
	if len(exprs) == 0 {
		return nil, nil
	}
	in, err := os.ReadFile(values)
	if err != nil {
		return nil, usage("cannot read %s", values)
	}
	next, err := editTOML(in, exprs)
	if err != nil {
		return nil, err
	}
	return &Change{Path: values, Content: next, Note: "migrate the old form: " + strings.Join(notes, "; ")}, nil
}

// addLines gives the text file <path> with each line of <lines> that it does
// not have at its end, or nil when it has each line.
func addLines(path string, lines []string) (*Change, error) {
	cur, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, usage("cannot read %s", path)
	}
	have := map[string]bool{}
	for _, l := range strings.Split(string(cur), "\n") {
		have[strings.TrimRight(l, "\r")] = true
	}
	text := string(cur)
	added := 0
	for _, l := range lines {
		if have[l] {
			continue
		}
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += l + "\n"
		have[l] = true
		added++
	}
	if added == 0 {
		return nil, nil
	}
	return &Change{Path: path, Content: []byte(text), Note: "add the lines of [overrides.gitignore] add"}, nil
}

// addExtensions gives the JSONC file <path> with each extension of <ids> in
// its `recommendations`, so its comments stay, or nil when it has each one.
func addExtensions(path string, ids []string) (*Change, error) {
	cur, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cur, err = []byte("{}\n"), nil
	}
	if err != nil {
		return nil, usage("cannot read %s", path)
	}
	v, err := hujson.Parse(cur)
	if err != nil {
		return nil, usage("cannot parse %s: %v", path, err)
	}
	std := v.Clone()
	std.Standardize()
	var doc map[string]any
	if err := json.Unmarshal(std.Pack(), &doc); err != nil {
		return nil, usage("%s is not a JSON object", path)
	}
	have := map[string]bool{}
	list, hasList := doc["recommendations"].([]any)
	for _, x := range list {
		if s, ok := x.(string); ok {
			have[s] = true
		}
	}
	var patch []map[string]any
	if !hasList {
		patch = append(patch, map[string]any{"op": "add", "path": "/recommendations", "value": []any{}})
	}
	for _, id := range ids {
		if !have[id] {
			have[id] = true
			patch = append(patch, map[string]any{"op": "add", "path": "/recommendations/-", "value": id})
		}
	}
	if len(patch) == 0 || (len(patch) == 1 && !hasList && len(ids) == 0) {
		return nil, nil
	}
	b, _ := json.Marshal(patch)
	if err := v.Patch(b); err != nil {
		return nil, usage("cannot change %s: %v", path, err)
	}
	v.Format()
	return &Change{Path: path, Content: v.Pack(), Note: "add the old top-level extensions"}, nil
}

// addHooks gives the YAML file <path> with the `repos` blocks of the old file
// <old> at the end of its `repos`, so its comments stay.
func addHooks(path, old string) (Change, error) {
	note := "add the hooks of " + OldPrecommit
	if !isFile(path) {
		b, err := os.ReadFile(old)
		if err != nil {
			return Change{}, usage("cannot read %s", old)
		}
		return Change{Path: path, Content: b, Note: note}, nil
	}
	out, stderr, err := format.YQ("-p", "yaml", "-o", "yaml", ".repos = ((.repos // []) + (load("+q(old)+").repos // []))", path)
	if err != nil {
		return Change{}, usage("cannot add the hooks of %s to %s: %s", OldPrecommit, path, strings.TrimSpace(stderr))
	}
	return Change{Path: path, Content: []byte(out), Note: note}, nil
}

// Apply writes each change. A removal also removes its folder when the folder
// is then empty.
func Apply(changes []Change) error {
	for _, c := range changes {
		if c.Remove {
			if err := os.Remove(c.Path); err != nil && !os.IsNotExist(err) {
				return fail.New(fail.Other, "cannot remove %s", c.Path)
			}
			_ = os.Remove(filepath.Dir(c.Path))
			continue
		}
		mode := os.FileMode(0o644)
		if st, err := os.Stat(c.Path); err == nil {
			mode = st.Mode().Perm()
		}
		if err := os.MkdirAll(filepath.Dir(c.Path), 0o777); err != nil {
			return fail.New(fail.Other, "cannot write %s", c.Path)
		}
		tmp := c.Path + ".bootstrap-new"
		if err := os.WriteFile(tmp, c.Content, mode); err != nil {
			os.Remove(tmp)
			return fail.New(fail.Other, "cannot write %s", c.Path)
		}
		if err := os.Rename(tmp, c.Path); err != nil {
			os.Remove(tmp)
			return fail.New(fail.Other, "cannot write %s", c.Path)
		}
	}
	return nil
}

// Paths gives the paths of <changes>, sorted.
func Paths(changes []Change) []string {
	var out []string
	for _, c := range changes {
		out = append(out, c.Path)
	}
	sort.Strings(out)
	return out
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.Mode().IsRegular()
}
