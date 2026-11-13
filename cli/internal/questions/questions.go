// Package questions reads the `[questions.<key>]` tables of the selected
// tools, finds their answers in a values file and its overrides file, and
// gives the proposal of a question: the output of its `detect` command, else
// its `example`.
//
// Contract: docs/design/tools.md ("The defaults, questions and contributes
// tables").
package questions

import (
	"os/exec"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/source"
)

// Question is one `[questions.<key>]` table of a tool.
type Question struct {
	Tool, Key string
	Text      string // The `question` string.
	Detect    string // The `detect` shell command, or empty.
	Example   string // The `example`, or empty.
	Group     string // The `group`, or empty: a group needs one answer that is not empty.
	Note      string // The `note`, or empty.
	Optional  bool   // `optional = true`: the question can have no answer.
}

// Necessary tells if <q> needs an answer that is not empty: it is not
// optional and not in a group.
func (q Question) Necessary() bool { return !q.Optional && q.Group == "" }

// GroupName is `<tool>.<group>`.
func (q Question) GroupName() string { return q.Tool + "." + q.Group }

// Name is `<tool>.<key>`.
func (q Question) Name() string { return q.Tool + "." + q.Key }

// Of gives the questions of <tools>, in tool order, each tool in key order.
func Of(idx *source.Index, tools []string) []Question {
	var out []Question
	for _, tool := range tools {
		m := idx.Manifests[tool]
		if m == nil || m.Raw == nil {
			continue
		}
		qs, ok := m.Raw.Get("questions").(*source.Obj)
		if !ok {
			continue
		}
		keys := append([]string(nil), qs.Keys()...)
		sort.Strings(keys)
		for _, k := range keys {
			e, ok := qs.Get(k).(*source.Obj)
			if !ok {
				continue
			}
			q := Question{Tool: tool, Key: k}
			q.Text, _ = e.Get("question").(string)
			q.Detect, _ = e.Get("detect").(string)
			q.Example, _ = e.Get("example").(string)
			q.Group, _ = e.Get("group").(string)
			q.Note, _ = e.Get("note").(string)
			q.Optional, _ = e.Get("optional").(bool)
			out = append(out, q)
		}
	}
	return out
}

// Answer gives the answer of <q>: its key in `[tools.<tool>]` of the
// overrides file <over>, else of the values file <user>. Both hold the
// `[tools.<name>]` tables by tool name, and each can be nil.
func Answer(user, over *source.Obj, q Question) (any, bool) {
	for _, o := range []*source.Obj{over, user} {
		if o == nil {
			continue
		}
		if t, ok := o.Get(q.Tool).(*source.Obj); ok && t.Has(q.Key) {
			return t.Get(q.Key), true
		}
	}
	return nil, false
}

// Empty tells if the answer <v> is the empty string: no answer.
func Empty(v any) bool {
	s, ok := v.(string)
	return ok && s == ""
}

// Gap is a question without an answer, or a group without an answer that is
// not empty.
type Gap struct {
	Question Question // The question; for a group, its first question.
	Group    bool     // A group: Question.GroupName() names it.
	Keys     []string // A group: the keys of its questions, in key order.
}

// Name gives `<tool>.<key>`, or `<tool>.<group>` for a group.
func (g Gap) Name() string {
	if g.Group {
		return g.Question.GroupName()
	}
	return g.Question.Name()
}

// Members gives the questions of the group of <q> in <all>, in their order.
func Members(all []Question, q Question) []Question {
	var out []Question
	for _, o := range all {
		if q.Group != "" && o.Tool == q.Tool && o.Group == q.Group {
			out = append(out, o)
		}
	}
	return out
}

// GroupAnswered tells if a question of the group of <q> has an answer that
// is not empty in the files.
func GroupAnswered(all []Question, q Question, user, over *source.Obj) bool {
	for _, m := range Members(all, q) {
		if v, ok := Answer(user, over, m); ok && !Empty(v) {
			return true
		}
	}
	return false
}

// Missing gives each question of <tools> that has no answer, and each group
// whose answers are all empty or missing. An optional question can have no
// answer. A question of a group needs no key; its group needs one answer
// that is not empty.
func Missing(idx *source.Index, tools []string, user, over *source.Obj) []Gap {
	all := Of(idx, tools)
	var out []Gap
	seen := map[string]bool{}
	for _, q := range all {
		switch {
		case q.Optional:
		case q.Group != "":
			if seen[q.GroupName()] {
				continue
			}
			seen[q.GroupName()] = true
			if !GroupAnswered(all, q, user, over) {
				g := Gap{Question: q, Group: true}
				for _, m := range Members(all, q) {
					g.Keys = append(g.Keys, m.Key)
				}
				out = append(out, g)
			}
		default:
			if _, ok := Answer(user, over, q); !ok {
				out = append(out, Gap{Question: q})
			}
		}
	}
	return out
}

// Proposal gives the proposal of <q> and its source: the first line of the
// output of `detect`, run with bash in <dir>, else the example. A detect that
// fails or gives nothing falls back to the example. With neither, both are
// empty.
func Proposal(q Question, dir string) (value, from string) {
	if q.Detect != "" {
		cmd := exec.Command("bash", "-c", q.Detect)
		cmd.Dir = dir
		if out, err := cmd.Output(); err == nil {
			line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
			if line = strings.TrimSpace(line); line != "" {
				return line, "detect"
			}
		}
	}
	if q.Example != "" {
		return q.Example, "example"
	}
	return "", ""
}
