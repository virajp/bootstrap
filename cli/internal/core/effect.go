package core

import (
	"os"
	"path/filepath"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// Effect is what the items of a change list do to the selection of a state,
// before the questions and the sync.
type Effect struct {
	Added   map[string]bool // The tools that the items select, with their new dependencies.
	Removed map[string]bool // The tools that the items remove, with the dependencies that no tool needs.
	Direct  map[string]bool // The tools of Added that an item names, or that a pack of an item adds.
}

// Effect applies the items <items> in order to the selection of <s>, as
// Plan does, and gives the tools that they add and remove. It writes
// nothing, runs no script and no sync, and gives the problem of the first
// item that Plan refuses. The TUI calls it after each change of its list.
func (s *State) Effect(items []Item) (*Effect, error) {
	p := &PlanResult{Side: s.Side}
	pl := &planner{st: s, p: p, o: Options{DryRun: true}, added: map[string]bool{}, removed: map[string]bool{}, direct: map[string]bool{}}
	for _, it := range items {
		if err := pl.item(it); err != nil {
			return nil, err
		}
	}
	return &Effect{Added: pl.added, Removed: pl.removed, Direct: pl.direct}, nil
}

// InitState gives the state of the values file after the change of init
// (InitChange): a new values file with the members of the pack base, or the
// values file migrated from its old form. It writes only to a temporary
// folder. The TUI shows the tools of init with it.
func InitState(sd Side) (*State, error) {
	p := &PlanResult{Side: sd}
	if err := p.initValues(); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp(tmpDir(), "bootstrap-init.")
	if err != nil {
		return nil, problem(fail.Other, "cannot create a temp folder", "")
	}
	defer os.RemoveAll(tmp)
	tmpValues := tmp + "/" + filepath.Base(sd.Values)
	read := func() (*State, error) {
		if err := os.WriteFile(tmpValues, p.Next, 0o644); err != nil {
			return nil, problem(fail.Other, "cannot write a temp file", "")
		}
		raw, err := source.ReadTOML(tmpValues)
		if err != nil {
			return nil, problem(fail.Usage, "cannot read "+sd.Show(sd.Values), "")
		}
		return stateOf(sd, tmpValues, p.Next, raw)
	}
	copyOverrides(sd.Values, tmpValues)
	st, err := read()
	if err != nil || p.Existed || !st.Packs.HasPack(selection.BasePack) {
		return st, err
	}
	// A new values file gets the members of base, as Plan does.
	pl := &planner{st: st, p: p, added: map[string]bool{}, removed: map[string]bool{}}
	if err := pl.addPack([]string{selection.BasePack}, nil); err != nil {
		return nil, err
	}
	if len(pl.exprs) == 0 {
		return st, nil
	}
	if p.Next, err = EditTOML(p.Next, pl.exprs); err != nil {
		return nil, err
	}
	return read()
}
