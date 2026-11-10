package selection

import (
	"strings"

	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// BasePack is the pack that every repo and every Mac gets.
const BasePack = "base"

// Resolved is the selection of one side, packs expanded and dependencies
// included.
type Resolved struct {
	Index *source.Index
	Sel   []Entry  // `_RS_SEL`: the entries of the packs and the values file.
	Tools []string // The tools to sync. On the repo side in walk order (a dependency before its tool). On the Mac side sorted.
	Deps  []DepRow // The tools that have dependencies.
	Keep  []string // The base and dependency tools, for the values merge.
	// Init holds the Mac tools of the init stage: the selected tools with
	// `stage = "init"`, and all their dependencies. It is empty on the repo
	// side.
	Init map[string]bool
	// InitOrder is Init in the order of bash: the init-stage tools sorted, then
	// each added dependency in the order of the search.
	InitOrder []string
}

// Stage gives the stage of a Mac tool: "init" or "main".
func (r *Resolved) Stage(tool string) string {
	if r.Init[tool] {
		return "init"
	}
	return "main"
}

// ResolveRepo is `_rs_resolve` without the check of the `env` names, which
// needs the environments of the repo (the repo front-end does it). <idx> is
// the checked index of the clone. <inSubmodule> is true when the repo is a git
// submodule: a `root_only` tool that a pack adds is skipped, and a
// `[tools.*]` entry of it stops the sync. The error is the first problem, with
// code 2 or 3.
func ResolveRepo(idx *source.Index, v *values.Repo, inSubmodule bool) (*Resolved, error) {
	src := idx.Source
	sel, err := Select(src, BasePack, v.Sel)
	if err != nil {
		return nil, err
	}
	if inSubmodule {
		kept := sel[:0:0]
		for _, e := range sel {
			if m := idx.Manifests[e.Tool]; !strings.HasPrefix(e.Src, "tools.") && m != nil && m.RootOnly {
				continue
			}
			kept = append(kept, e)
		}
		sel = kept
	}
	shown := src.ToolsShown()
	deps := func(tool string) ([]string, error) {
		m := idx.Manifests[tool]
		if m == nil {
			if !src.HasManifest(tool) {
				return nil, missing("tool '%s' is not in bootstrap", tool)
			}
			return nil, usage("cannot read %s/%s/tools.toml", shown, tool)
		}
		return append([]string(nil), m.Depends...), nil
	}
	done := func(tool string) error {
		m := idx.Manifests[tool]
		for _, env := range m.DefaultEnvs {
			if !envNameOK(env) {
				return usage("default_envs of %s/%s/tools.toml: '%s' is not one of %s", shown, tool, env, strings.Join(source.EnvNames, ", "))
			}
		}
		if m.RootOnly && inSubmodule {
			return usage("tool '%s' is for the root repository only - remove it from the tools of this submodule", tool)
		}
		return nil
	}
	w, err := Walk(entryTools(sel), deps, done)
	if err != nil {
		return nil, err
	}
	return &Resolved{Index: idx, Sel: sel, Tools: w.Visited, Deps: w.Deps, Keep: keepTools(sel, w.Deps)}, nil
}

func envNameOK(name string) bool {
	for _, n := range source.EnvNames {
		if n == name {
			return name != "" && !strings.ContainsAny(name, " \t\n")
		}
	}
	return false
}
