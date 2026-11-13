package repo

import (
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/questions"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/tmpl"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// toolValues is `_rs_tool_values`: it merges the values of the selected
// tools. The user values are each `[tools.<name>]` table of the values file
// without `env`. It gives the code of the merge, after its messages.
func (s *state) toolValues() fail.Code {
	overrides := values.OverridesPath(s.values)
	over, err := values.ReadOverrides(overrides, s.path(overrides))
	if err != nil {
		return s.logAll(err)
	}
	m, err := values.Merge(values.Input{
		Index:       s.idx,
		Selected:    s.res.Tools,
		Keep:        s.res.Keep,
		User:        s.vals.User,
		ShownRoot:   toolsShown,
		ShownValues: s.show,
		Overrides:   over,
		Derive:      s.derive,
	})
	if err != nil {
		return s.logAll(err)
	}
	if missing := questions.Missing(s.idx, s.res.Tools, s.vals.User, over); len(missing) > 0 {
		for _, g := range missing {
			what := "the question "
			if g.Group {
				what = "the group "
			}
			s.log(what + g.Name() + " has no answer in " + s.show + " - run bootstrap init")
		}
		return fail.Usage
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
// `cli_dir`, `precommit` and `all` merged in first, so a key of the same
// name in the values file cannot replace them. The keys `all` and `config` of
// the values file do not go in. It gives nil, after the messages, when the
// context cannot be built.
func (s *state) contexts() *contexts {
	rows, code := s.memberList("template context")
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
	var sources []ds
	env := map[string]string{
		"BOOTSTRAP_CLI_DIR":        s.dir + "/" + source.CLIDir,
		"BOOTSTRAP_MEMBERS":        strings.Join(names, "\n"),
		"BOOTSTRAP_MEMBER_FOLDERS": strings.Join(folders, "\n"),
		"BOOTSTRAP_MEMBER_TYPES":   strings.Join(kinds, "\n"),
		"BOOTSTRAP_SUBMODULES":     submodules,
		"BOOTSTRAP_TOOLS":          strings.Join(s.res.Tools, "\n"),
	}
	out, ok := s.gomplate(tmplContext, ctx, sources, env, false)
	if !ok {
		return nil
	}
	return &contexts{shared: []byte(out)}
}
