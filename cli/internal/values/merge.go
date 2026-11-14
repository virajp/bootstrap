package values

import (
	"fmt"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// Input is the input of Merge.
type Input struct {
	Index *source.Index // The tools.toml files of the clone.
	// Selected holds the selected tool names, and Keep the base and dependency
	// tool names. Both can hold a name twice, and they need no order.
	Selected, Keep []string
	User           *source.Obj // The user values: tool name to table.
	ShownRoot      string      // The folder of the tools in a message, for example repo/tools.
	ShownValues    string      // The values file in a message.
	// Overrides holds the `[tools.<name>]` tables of the overrides file. They
	// win over all.
	Overrides *source.Obj
	// Derive runs the derive scripts. It gets the selected tools and their
	// values of levels 1 and 2, and gives the values of level 3 of each tool
	// (nil for a tool without a derive script).
	Derive func(tools []string, base map[string]*source.Obj) (map[string]*source.Obj, error)
}

// Result holds the merged values of the selected tools.
type Result struct {
	Tools  []string               // The selected tools, sorted, without a repeat.
	Values map[string]*source.Obj // The merged values of each tool.
}

// Lines gives the `ctx<TAB><tool><TAB><json>` lines of `_se_values`, one for
// each selected tool, each with a trailing newline.
func (r *Result) Lines() string {
	var b strings.Builder
	for _, t := range r.Tools {
		b.WriteString("ctx\t" + t + "\t" + source.JSON(r.Values[t]) + "\n")
	}
	return b.String()
}

// All gives the values of all the tools as one object: the tool name to its
// values, in tool-name order.
func (r *Result) All() *source.Obj {
	o := source.NewObj()
	for _, t := range r.Tools {
		o.Set(t, r.Values[t])
	}
	return o
}

// pathErr is the error of the three merge functions that can fail.
type pathErr struct {
	path []string
	a, b string // The jq types, for the type change.
}

func appendNew(a, b []any) []any {
	out := append([]any(nil), a...)
	for _, x := range b {
		dup := false
		for _, e := range out {
			if source.Equal(e, x) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, x)
		}
	}
	return out
}

// replaceMerge: a later layer. Tables merge deeply, any other value replaces.
func replaceMerge(a, b any) any {
	ao, ok1 := a.(*source.Obj)
	bo, ok2 := b.(*source.Obj)
	if ok1 && ok2 {
		out := ao.Clone()
		for _, k := range bo.Keys() {
			out.Set(k, replaceMerge(ao.Get(k), bo.Get(k)))
		}
		return out
	}
	return b
}

// contributionMerge: the contributions over [defaults]. Tables merge, lists
// append, else <b> wins.
func contributionMerge(a, b any) any {
	ao, ok1 := a.(*source.Obj)
	bo, ok2 := b.(*source.Obj)
	if ok1 && ok2 {
		out := ao.Clone()
		for _, k := range bo.Keys() {
			out.Set(k, contributionMerge(ao.Get(k), bo.Get(k)))
		}
		return out
	}
	al, ok1 := a.([]any)
	bl, ok2 := b.([]any)
	if ok1 && ok2 {
		return appendNew(al, bl)
	}
	return b
}

// contributorsMerge: two contributions. Two different values that are not both
// tables or both lists are an error with their path.
func contributorsMerge(a, b any, p []string) (any, *pathErr) {
	if a == nil {
		return b, nil
	}
	ao, ok1 := a.(*source.Obj)
	bo, ok2 := b.(*source.Obj)
	if ok1 && ok2 {
		out := ao.Clone()
		for _, k := range bo.Keys() {
			v, err := contributorsMerge(ao.Get(k), bo.Get(k), append(append([]string(nil), p...), k))
			if err != nil {
				return nil, err
			}
			out.Set(k, v)
		}
		return out, nil
	}
	al, ok1 := a.([]any)
	bl, ok2 := b.([]any)
	if ok1 && ok2 {
		return appendNew(al, bl), nil
	}
	if source.Equal(a, b) {
		return a, nil
	}
	return nil, &pathErr{path: p}
}

// keepMerge: the user values of a base or dependency tool. A value whose type
// differs from the merged one is an error with its path and both types.
func keepMerge(a, b any, p []string) (any, *pathErr) {
	if a == nil {
		return b, nil
	}
	ao, ok1 := a.(*source.Obj)
	bo, ok2 := b.(*source.Obj)
	if ok1 && ok2 {
		out := ao.Clone()
		for _, k := range bo.Keys() {
			v, err := keepMerge(ao.Get(k), bo.Get(k), append(append([]string(nil), p...), k))
			if err != nil {
				return nil, err
			}
			out.Set(k, v)
		}
		return out, nil
	}
	al, ok1 := a.([]any)
	bl, ok2 := b.([]any)
	if ok1 && ok2 {
		return appendNew(al, bl), nil
	}
	if source.Kind(a) != source.Kind(b) {
		return nil, &pathErr{path: p, a: source.Kind(a), b: source.Kind(b)}
	}
	return b, nil
}

// typeWord is a jq type as the TOML word for it.
func typeWord(t string) string {
	switch t {
	case "object":
		return "table"
	case "array":
		return "list"
	}
	return t
}

func getpath(v any, path []string) any {
	for _, k := range path {
		o, ok := v.(*source.Obj)
		if !ok {
			return nil
		}
		v = o.Get(k)
	}
	return v
}

func tableOf(v any) *source.Obj {
	if o, ok := v.(*source.Obj); ok {
		return o
	}
	return source.NewObj()
}

// shapeMessages gives the messages of the tables of the manifest <t> of <tool>.
func shapeMessages(shown, tool string, t *source.Obj) []string {
	at := fmt.Sprintf("%s/%s/tools.toml: ", shown, tool)
	var out []string
	if t.Has("defaults") {
		if _, ok := t.Get("defaults").(*source.Obj); !ok {
			out = append(out, at+"[defaults] must be a table")
		}
	}
	if t.Has("questions") {
		qs, ok := t.Get("questions").(*source.Obj)
		if !ok {
			out = append(out, at+"[questions] must be a table of [questions.<key>] tables")
		} else {
			for _, q := range qs.Keys() {
				e, ok := qs.Get(q).(*source.Obj)
				if !ok {
					out = append(out, fmt.Sprintf("%s[questions.%s] must be a table", at, q))
					continue
				}
				for _, k := range e.Keys() {
					if !source.QuestionKey(k) {
						out = append(out, fmt.Sprintf("%s[questions.%s] %s is not a key of a question - the keys are %s", at, q, k, source.QuestionKeys))
					}
				}
				for _, text := range source.QuestionTypes(q, e) {
					out = append(out, at+text)
				}
				if !e.Has("question") {
					out = append(out, fmt.Sprintf("%s[questions.%s] needs a question string", at, q))
				} else if _, ok := e.Get("question").(string); !ok {
					out = append(out, fmt.Sprintf("%s[questions.%s] question must be a string", at, q))
				}
				if e.Has("detect") {
					if _, ok := e.Get("detect").(string); !ok {
						out = append(out, fmt.Sprintf("%s[questions.%s] detect must be a string (a shell command)", at, q))
					}
				}
			}
		}
	}
	if t.Has("contributes") {
		cs, ok := t.Get("contributes").(*source.Obj)
		if !ok {
			out = append(out, at+"[contributes] must be a table of [contributes.<tool>] tables")
		} else {
			for _, o := range cs.Keys() {
				if _, ok := cs.Get(o).(*source.Obj); !ok {
					out = append(out, fmt.Sprintf("%s[contributes.%s] must be a table", at, o))
				}
			}
		}
	}
	return out
}

func sortedUnique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// Merge merges the values of the selected tools, as `_se_values` does. The
// error holds each message of bash in its order, and the code is 3 when a
// contribution names a tool that bootstrap does not have, else 2. With no
// selected tool, the result is empty.
func Merge(in Input) (*Result, error) {
	names := sortedUnique(in.Selected)
	res := &Result{Tools: names, Values: map[string]*source.Obj{}}
	if len(names) == 0 {
		return res, nil
	}
	keep := map[string]bool{}
	for _, k := range in.Keep {
		keep[k] = true
	}
	root := map[string]bool{}
	for _, n := range in.Index.Names {
		root[n] = true
	}
	// The picked manifest of each tool: defaults, questions and contributes.
	raw := map[string]*source.Obj{}
	defaults := map[string]*source.Obj{}
	contribs := map[string]*source.Obj{}
	for _, n := range names {
		m, ok := in.Index.Manifests[n]
		if !ok {
			return nil, fail.New(fail.Usage, "cannot read the tools.toml files of the selected tools")
		}
		raw[n] = m.Raw
		defaults[n] = tableOf(m.Raw.Get("defaults"))
		c := source.NewObj()
		if co, ok := m.Raw.Get("contributes").(*source.Obj); ok {
			for _, k := range co.Keys() {
				if _, ok := co.Get(k).(*source.Obj); ok {
					c.Set(k, co.Get(k))
				}
			}
		}
		contribs[n] = c
	}

	var bad2, bad3 []string
	for _, n := range names {
		bad2 = append(bad2, shapeMessages(in.ShownRoot, n, raw[n])...)
	}
	for _, t := range names {
		for _, o := range contribs[t].Keys() {
			if !root[o] {
				bad3 = append(bad3, fmt.Sprintf("%s/%s/tools.toml: [contributes.%s] names a tool that bootstrap does not have", in.ShownRoot, t, o))
			}
		}
	}

	type contribution struct {
		n   string
		v   any
		bad string
	}
	var contributions []contribution
	for _, n := range names {
		var from []string
		for _, t := range names {
			if contribs[t].Has(n) {
				from = append(from, t)
			}
		}
		acc := any(source.NewObj())
		var failed *pathErr
		failTool := ""
		for _, t := range from {
			v, err := contributorsMerge(acc, contribs[t].Get(n), nil)
			if err != nil {
				failed, failTool = err, t
				break
			}
			acc = v
		}
		if failed == nil {
			contributions = append(contributions, contribution{n: n, v: acc})
			continue
		}
		first := "null"
		for _, t := range from {
			if t < failTool && getpath(contribs[t].Get(n), failed.path) != nil {
				first = t
				break
			}
		}
		contributions = append(contributions, contribution{n: n, bad: fmt.Sprintf("%s/%s/tools.toml and %s/%s/tools.toml set [contributes.%s] %s to different values - remove one",
			in.ShownRoot, first, in.ShownRoot, failTool, n, strings.Join(failed.path, "."))})
	}

	var probs []*fail.Error
	for _, m := range bad2 {
		probs = append(probs, fail.New(fail.Usage, "%s", m))
	}
	for _, m := range bad3 {
		probs = append(probs, fail.New(fail.Missing, "%s", m))
	}
	for _, c := range contributions {
		if c.bad != "" {
			probs = append(probs, fail.New(fail.Usage, "%s", c.bad))
		}
	}
	if len(probs) > 0 {
		return nil, source.Problems(probs)
	}

	// Levels 1 and 2: the defaults, then the contributions.
	base := map[string]*source.Obj{}
	for _, x := range contributions {
		base[x.n] = asObj(contributionMerge(defaults[x.n], x.v))
	}
	// Level 3: the output of the derive script of each tool wins over 1 and 2.
	if in.Derive != nil {
		derived, err := in.Derive(names, base)
		if err != nil {
			return nil, err
		}
		for _, n := range names {
			if d := derived[n]; d != nil {
				base[n] = asObj(replaceMerge(base[n], d))
			}
		}
	}

	type merged struct {
		n   string
		v   *source.Obj
		bad string
	}
	var mergedList []merged
	for _, x := range contributions {
		b := base[x.n]
		user := any(source.NewObj())
		if in.User != nil && in.User.Has(x.n) && in.User.Get(x.n) != nil {
			user = in.User.Get(x.n)
		}
		var v *source.Obj
		if keep[x.n] {
			kv, err := keepMerge(b, user, nil)
			if err != nil {
				mergedList = append(mergedList, merged{n: x.n, bad: fmt.Sprintf("[tools.%s] of %s sets %s to a %s, but it is a %s in the base or dependency tool %s - write a %s there (a list adds its items)",
					x.n, in.ShownValues, strings.Join(err.path, "."), typeWord(err.b), typeWord(err.a), x.n, typeWord(err.a))})
				continue
			}
			v = asObj(kv)
		} else {
			v = asObj(replaceMerge(b, user))
		}
		// The overrides file wins over all.
		if in.Overrides != nil && in.Overrides.Has(x.n) {
			if o, ok := in.Overrides.Get(x.n).(*source.Obj); ok {
				v = asObj(replaceMerge(v, o))
			}
		}
		mergedList = append(mergedList, merged{n: x.n, v: v})
	}
	for _, m := range mergedList {
		if m.bad != "" {
			probs = append(probs, fail.New(fail.Usage, "%s", m.bad))
		}
	}
	if len(probs) > 0 {
		return nil, source.Problems(probs)
	}
	for _, m := range mergedList {
		res.Values[m.n] = m.v
	}
	return res, nil
}

func asObj(v any) *source.Obj {
	if o, ok := v.(*source.Obj); ok {
		return o
	}
	return source.NewObj()
}
