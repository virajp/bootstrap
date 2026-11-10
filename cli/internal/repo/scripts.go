package repo

import (
	"path/filepath"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/scripts"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// runner gives the script runner of the run: the scripts run in the root of
// the repository.
func (s *state) runner() scripts.Runner {
	return scripts.Runner{ToolsDir: s.src.ToolsDir(), Shown: toolsShown, Root: s.root}
}

// memberList gives the members of the repo, once in each run. <name> starts
// each message of a member that breaks a rule.
func (s *state) memberList(name string) ([]member, fail.Code) {
	if !s.membersRead {
		s.memberRows, s.membersCode = s.members(name)
		s.membersRead = true
	}
	return s.memberRows, s.membersCode
}

// repoInput gives the `repo` object of the input of a script.
func (s *state) repoInput() *scripts.Repo {
	name := filepath.Base(s.root)
	if s.vals != nil {
		if n, ok := s.vals.Raw.Get("repo_name").(string); ok && n != "" {
			name = n
		}
	}
	r := &scripts.Repo{Root: s.root, Name: name, Members: []scripts.Member{}}
	rows, _ := s.memberList("template context")
	for _, m := range rows {
		r.Members = append(r.Members, scripts.Member{Name: m.name, Folder: m.folder, Type: m.kind})
	}
	return r
}

// jobs gives one job of the script <name> for each tool of <tools>.
func (s *state) jobs(name string, tools []string, vals func(string) *source.Obj) []scripts.Job {
	repo := s.repoInput()
	var out []scripts.Job
	for _, t := range tools {
		in := scripts.Input{Tool: t, Values: vals(t), Repo: repo}
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
		if o == nil {
			continue
		}
		rest, mise := source.NewObj(), source.NewObj()
		for _, k := range o.Keys() {
			if k == "mise" || k == "mise_env" {
				mise.Set(k, o.Get(k))
			} else {
				rest.Set(k, o.Get(k))
			}
		}
		if mise.Len() > 0 {
			s.derived[res.Tool] = mise
		}
		out[res.Tool] = rest
	}
	return out, source.Problems(probs)
}

// toolVals gives the merged values of <tool>.
func (s *state) toolVals(tool string) *source.Obj {
	if s.all != nil {
		if v, ok := s.all.Get(tool).(*source.Obj); ok {
			return v
		}
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

// when tells if the entry of <f> applies.
func (s *state) when(f source.File) bool { return source.When(s.all, f.When) }
