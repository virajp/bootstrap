package macos

import (
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/scripts"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// runner gives the script runner of the run: the scripts run in $HOME.
func (s *state) runner() scripts.Runner {
	return scripts.Runner{ToolsDir: s.src.ToolsDir(), Shown: s.src.ToolsShown(), Root: s.home}
}

// jobs gives one job of the script <name> for each tool of <tools>, with
// the values of <vals>.
func (s *state) jobs(name string, tools []string, vals func(string) *source.Obj) []scripts.Job {
	var out []scripts.Job
	for _, t := range tools {
		in := scripts.Input{Tool: t, Values: vals(t), Home: s.home}
		out = append(out, scripts.Job{Tool: t, Name: name, Input: in.JSON()})
	}
	return out
}

// derive runs the derive script of each tool of <tools> with the values of
// levels 1 and 2 in <base>. It keeps the derived `mise` and `mise_env`
// tables for the generator, and gives the other values.
func (s *state) derive(tools []string, base map[string]*source.Obj) (map[string]*source.Obj, error) {
	r := s.runner()
	results := r.Run(s.jobs(scripts.Derive, tools, func(t string) *source.Obj { return base[t] }))
	out := map[string]*source.Obj{}
	var probs []*fail.Error
	for _, res := range results {
		o, p := r.Derived(res)
		if p != nil {
			probs = append(probs, p...)
			continue
		}
		if res.Ran && len(res.Stderr) > 0 {
			s.stderr.Write(res.Stderr)
		}
		if o != nil {
			out[res.Tool] = splitMise(o, res.Tool, s.derived)
		}
	}
	return out, source.Problems(probs)
}

// splitMise moves the `mise` and `mise_env` keys of the derived values <o>
// of <tool> into <into>, and gives the rest.
func splitMise(o *source.Obj, tool string, into map[string]*source.Obj) *source.Obj {
	rest, mise := source.NewObj(), source.NewObj()
	for _, k := range o.Keys() {
		if k == "mise" || k == "mise_env" {
			mise.Set(k, o.Get(k))
		} else {
			rest.Set(k, o.Get(k))
		}
	}
	if mise.Len() > 0 {
		into[tool] = mise
	}
	return rest
}

// toolVals gives the merged values of <tool>.
func (s *state) toolVals(tool string) *source.Obj {
	if v, ok := s.all.Get(tool).(*source.Obj); ok {
		return v
	}
	return source.NewObj()
}

// checks runs the check script of each selected tool. It gives 2, after the
// messages, when a script finds a problem.
func (s *state) checks() fail.Code {
	r := s.runner()
	results := r.Run(s.jobs(scripts.Check, s.res.Tools, s.toolVals))
	var probs []*fail.Error
	for _, res := range results {
		if p := r.Checked(res); p != nil {
			probs = append(probs, p...)
		} else if res.Ran && len(res.Stderr) > 0 {
			s.stderr.Write(res.Stderr)
		}
	}
	return s.logAll(source.Problems(probs))
}

// after runs the after script of each selected tool of <stage>. It gives the
// restart reasons, as `<tool>: <reason>`, and 2 when a script fails.
func (s *state) after(stage string) ([]string, fail.Code) {
	var tools []string
	for _, t := range s.res.Tools {
		if s.inStage(stage, t) {
			tools = append(tools, t)
		}
	}
	r := s.runner()
	results := r.Run(s.jobs(scripts.After, tools, s.toolVals))
	var probs []*fail.Error
	var restart []string
	for _, res := range results {
		reason, p := r.Restart(res)
		if p != nil {
			probs = append(probs, p...)
			continue
		}
		if res.Ran && len(res.Stderr) > 0 {
			s.stderr.Write(res.Stderr)
		}
		if reason != "" {
			restart = append(restart, res.Tool+": "+reason)
		}
	}
	return restart, s.logAll(source.Problems(probs))
}
