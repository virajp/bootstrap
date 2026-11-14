package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/scripts"
	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// Selected is the selection of a values file of the side.
type Selected struct {
	Idx   *source.Index
	Tools []string
	User  *source.Obj // The [tools.<name>] tables of the values file.
	Over  *source.Obj // The [tools.<name>] tables of its overrides file.
	Res   *selection.Resolved
	Items []values.Item // The [tools.*] and [packs.*] entries of the values file.
}

// SelectionOf reads the values file <file> of the side <sd> (the overrides
// file is next to <overridesOf>) and resolves its tools from the clone.
func SelectionOf(sd Side, file, overridesOf string) (*Selected, error) {
	side := source.Repo
	if sd.Mac {
		side = source.Macos
	}
	idx, err := source.Source{Root: sd.Clone, Side: side}.Open()
	if err != nil {
		return nil, err
	}
	s := &Selected{Idx: idx}
	shown := sd.Show(file)
	var res *selection.Resolved
	if sd.Mac {
		v, err := values.ReadMacos(file, shown)
		if err != nil {
			return nil, err
		}
		s.User, s.Items = v.User, v.Sel
		if res, err = selection.ResolveMacos(idx, v); err != nil {
			return nil, err
		}
	} else {
		v := &values.Repo{Path: file, Raw: source.NewObj(), User: source.NewObj()}
		if Exists(file) {
			if v, err = values.ReadRepo(file, shown); err != nil {
				return nil, err
			}
		}
		s.User, s.Items = v.User, v.Sel
		if res, err = selection.ResolveRepo(idx, v, InSubmodule(sd.Root)); err != nil {
			return nil, err
		}
	}
	s.Tools, s.Res = res.Tools, res
	ov := values.OverridesPath(overridesOf)
	if s.Over, err = values.ReadOverrides(ov, sd.Show(ov)); err != nil {
		return nil, err
	}
	return s, nil
}

// State is what each selection command reads first: the values file and its
// selection.
type State struct {
	Side
	Cur   []byte // The values file.
	Raw   *source.Obj
	Sel   *Selected
	Res   *selection.Resolved
	Named map[string]bool // The [tools.<name>] tables of the values file.
	Packs *source.Packs
}

// Open reads the values file of <sd> and its selection. The values file must
// exist and have the new form.
func Open(sd Side) (*State, error) {
	cur, err := os.ReadFile(sd.Values)
	if err != nil {
		return nil, problem(fail.Usage, sd.Show(sd.Values)+" is not there", "run "+CmdName(sd.Mac, "init")+" first")
	}
	raw, err := source.ReadTOML(sd.Values)
	if err != nil {
		return nil, problem(fail.Usage, "cannot read "+sd.Show(sd.Values), "")
	}
	if sd.Mac {
		err = oldform.Mac(raw, sd.Show(sd.Values))
	} else {
		err = oldform.Repo(sd.Root, raw, sd.Show(sd.Values))
	}
	if err != nil {
		return nil, err
	}
	return stateOf(sd, sd.Values, cur, raw)
}

// stateOf gives the state of the values text <cur> in the file <file>, with
// the overrides file of the values file of <sd>.
func stateOf(sd Side, file string, cur []byte, raw *source.Obj) (*State, error) {
	sel, err := SelectionOf(sd, file, sd.Values)
	if err != nil {
		return nil, err
	}
	s := &State{Side: sd, Cur: cur, Raw: raw, Sel: sel, Res: sel.Res, Named: map[string]bool{}}
	for _, it := range sel.Items {
		if it.Kind == "tools" {
			s.Named[it.Name] = true
		}
	}
	if s.Packs, err = sel.Idx.Source.ReadPacks(); err != nil {
		return nil, err
	}
	return s, nil
}

// Manifest gives the manifest of <tool>, or nil.
func (s *State) Manifest(tool string) *source.Manifest { return s.Sel.Idx.Manifests[tool] }

// Selected tells if <tool> is in the selection.
func (s *State) Selected(tool string) bool {
	for _, t := range s.Res.Tools {
		if t == tool {
			return true
		}
	}
	return false
}

// PackOf gives the pack that selects <tool>, or empty.
func (s *State) PackOf(tool string) string {
	for _, e := range s.Res.Sel {
		if e.Tool != tool || strings.HasPrefix(e.Src, "tools.") {
			continue
		}
		if e.Src == "default" {
			return selection.BasePack
		}
		return strings.TrimPrefix(e.Src, "packs.")
	}
	return ""
}

// Required tells if <tool> has required = true in a pack.
func (s *State) Required(tool string) bool {
	for k, v := range s.Packs.Required {
		if v && strings.HasSuffix(k, "/"+tool) {
			return true
		}
	}
	return false
}

// Closure gives <start> and each tool that they depend on, transitively.
func (s *State) Closure(start []string) map[string]bool {
	out := map[string]bool{}
	queue := append([]string(nil), start...)
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		if out[t] {
			continue
		}
		out[t] = true
		if m := s.Manifest(t); m != nil {
			queue = append(queue, m.Depends...)
		}
	}
	return out
}

// Runner gives the script runner of the side: the scripts run in the
// repository, or in $HOME on the Mac.
func (s *State) Runner() scripts.Runner {
	src := s.Sel.Idx.Source
	return scripts.Runner{ToolsDir: src.ToolsDir(), Shown: src.ToolsShown(), Root: s.Home()}
}

// Input gives the JSON input of a script of <tool>: the [tools.<from>]
// table of the values file as its values.
func (s *State) Input(tool, from string) []byte {
	vals := source.NewObj()
	if v, ok := s.Sel.User.Get(from).(*source.Obj); ok {
		vals = v
	}
	in := scripts.Input{Tool: tool, Values: vals}
	if s.Mac {
		in.Home = os.Getenv("HOME")
	} else {
		in.Repo = s.repoInput()
	}
	return in.JSON()
}

// repoInput gives the repo object of a script input: the root, the
// repo_name, and the [member.<name>] tables.
func (s *State) repoInput() *scripts.Repo {
	r := &scripts.Repo{Root: s.Root, Name: filepath.Base(s.Root), Members: []scripts.Member{}}
	if n, ok := s.Raw.Get("repo_name").(string); ok && n != "" {
		r.Name = n
	}
	if m, ok := s.Raw.Get("member").(*source.Obj); ok {
		for _, name := range m.Keys() {
			e, _ := m.Get(name).(*source.Obj)
			mem := scripts.Member{Name: name, Folder: name}
			if e != nil {
				if f, ok := e.Get("folder").(string); ok {
					mem.Folder = f
				}
				mem.Type, _ = e.Get("type").(string)
			}
			r.Members = append(r.Members, mem)
		}
	}
	return r
}

// Known gives code 3, with a problem for each name of <names> that is not a
// tool of the side.
func (s *State) Known(names []string) error {
	var errs []error
	for _, n := range names {
		if !source.NameOK(n) || s.Manifest(n) == nil {
			errs = append(errs, problem(fail.Missing, n+" is not a bootstrap tool", ""))
		}
	}
	return join(errs)
}

// Refuse gives code 2, with a problem for each name of <names> that a
// command cannot select by its name: a tool with selectable = false, and a
// tool whose hosts leave out this host. It examines only the names, not
// their dependencies.
func Refuse(idx *source.Index, names []string) error {
	var errs []error
	host := source.HostName()
	for _, n := range names {
		m := idx.Manifests[n]
		if m == nil {
			continue
		}
		if !m.Selectable {
			errs = append(errs, problem(fail.Usage, idx.NotSelectable(n), ""))
		}
		if !m.RunsOn(host) {
			errs = append(errs, problem(fail.Usage, m.HostMessage(host), ""))
		}
	}
	return join(errs)
}

// Proposable tells if detect can propose <m>: it is selectable and runs on
// this host.
func Proposable(m *source.Manifest) bool {
	return m != nil && m.Selectable && m.RunsOn(source.HostName())
}
