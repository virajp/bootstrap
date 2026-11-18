package core

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
