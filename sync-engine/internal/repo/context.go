package repo

import (
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/tmpl"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// toolValues is `_rs_tool_values`: it merges the values of the selected
// tools. The user values are each `[tools.<name>]` table of the values file
// without `env`. It gives the code of the merge, after its messages.
func (s *state) toolValues() fail.Code {
	m, err := values.Merge(values.Input{
		Index:       s.idx,
		Selected:    s.res.Tools,
		Keep:        s.res.Keep,
		User:        s.vals.User,
		ShownRoot:   toolsShown,
		ShownValues: s.show,
	})
	if err != nil {
		return s.logAll(err)
	}
	s.all = m.All()
	s.allJSON = source.JSON(s.all)
	return fail.OK
}

// contexts holds the shared template context: the JSON text of
// `_rs_context`.
type contexts struct{ shared []byte }

// of gives the context of the templates of <tool>: the shared context plus
// `config`, the values of that tool (the file <tool>.json of
// `_rs_tool_contexts`).
func (c *contexts) of(tool string) (any, error) {
	v, err := tmpl.DecodeJSON(c.shared)
	if err != nil {
		return nil, err
	}
	m, ok := v.(map[string]any)
	if !ok {
		m = map[string]any{}
	}
	var config any
	if all, ok := m["all"].(map[string]any); ok {
		config = all[tool]
	}
	m["config"] = config
	return m, nil
}

// contexts is `_rs_context`: the values of the values file, with `members`,
// `member_folders`, `member_types`, `submodules`, `tools`, `extensions`,
// `sync_engine`, `precommit` and `all` merged in first, so a key of the same
// name in the values file cannot replace them. The keys `all` and `config` of
// the values file do not go in. It gives nil, after the messages, when the
// context cannot be built.
func (s *state) contexts() *contexts {
	rows, code := s.members("template context")
	if code != fail.OK {
		return nil
	}
	var names, folders, kinds []string
	for _, r := range rows {
		names = append(names, r.name)
		folders = append(folders, r.folder)
		kinds = append(kinds, r.kind)
	}
	// The git submodules alone: `members` also has the folder members.
	submodules := ""
	if isFile(s.root + "/.gitmodules") {
		submodules = s.submodules()
	}
	v, err := s.valuesDS()
	if err != nil {
		gomplateError(s.stderr, "failed to read datasource v ("+s.values+"): "+err.Error())
		return nil
	}
	all, err := tmpl.DecodeJSON([]byte(s.allJSON))
	if err != nil {
		return nil
	}
	ctx := map[string]any{"v": v, "all_values": all}
	// One context for each tool manifest, `tool_<name>`, for the
	// `[[precommit]]` entries; `repo_precommit` for the own entries of the
	// repo.
	var sources []ds
	for _, tool := range s.res.Tools {
		sources = append(sources, ds{"tool_" + tool, s.src.ToolDir(tool) + "/tools.toml", "toml"})
	}
	own := s.root + "/.config/repo-sync/pre-commit.yaml"
	hasOwn := ""
	if isFile(own) {
		sources = append(sources, ds{"repo_precommit", own, "yaml"})
		hasOwn = "1"
	}
	env := map[string]string{
		"REPO_SYNC_SYNC_ENGINE":    s.dir + "/sync-engine",
		"REPO_SYNC_HOOK_DROP":      " " + strings.Join(append(s.hookDrop, ""), " "),
		"REPO_SYNC_OWN_PRECOMMIT":  hasOwn,
		"REPO_SYNC_MEMBERS":        strings.Join(names, "\n"),
		"REPO_SYNC_MEMBER_FOLDERS": strings.Join(folders, "\n"),
		"REPO_SYNC_MEMBER_TYPES":   strings.Join(kinds, "\n"),
		"REPO_SYNC_SUBMODULES":     submodules,
		"REPO_SYNC_TOOLS":          strings.Join(s.res.Tools, "\n"),
		"REPO_SYNC_EXTENSIONS":     strings.Join(s.extensions, "\n"),
	}
	out, ok := s.gomplate(tmplContext, ctx, sources, env, false)
	if !ok {
		return nil
	}
	return &contexts{shared: []byte(out)}
}
