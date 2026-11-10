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

	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// Question is one `[questions.<key>]` table of a tool.
type Question struct {
	Tool, Key string
	Text      string // The `question` string.
	Detect    string // The `detect` shell command, or empty.
	Example   string // The `example`, or empty.
}

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

// Missing gives each question of <tools> that has no answer.
func Missing(idx *source.Index, tools []string, user, over *source.Obj) []Question {
	var out []Question
	for _, q := range Of(idx, tools) {
		if _, ok := Answer(user, over, q); !ok {
			out = append(out, q)
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
