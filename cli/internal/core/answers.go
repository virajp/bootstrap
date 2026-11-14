package core

import (
	"fmt"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/questions"
)

// CheckAnswers examines the answers <answers> of the selection <sel>, and
// gives each one by the name of its question. The answer of a flag must name
// a question of the selected tools, and an empty answer is only for an
// optional or a group question: else a problem with code 2.
func CheckAnswers(sel *Selected, answers []Answer) (map[string]string, error) {
	byName := map[string]questions.Question{}
	for _, qq := range questions.Of(sel.Idx, sel.Tools) {
		byName[qq.Name()] = qq
	}
	given := map[string]string{}
	var errs []error
	for _, ans := range answers {
		if ans.Flag == "" {
			given[ans.Name] = ans.Value
			continue
		}
		name, value, ok := strings.Cut(ans.Flag, "=")
		qq, known := byName[name]
		if !ok || !known {
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("--answer %s: no question %s of the selected tools", ans.Flag, name), "the form is <tool>.<key>=<value>"))
			continue
		}
		if value == "" && qq.Necessary() {
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("--answer %s: an empty answer is only for an optional or a group question", ans.Flag), fmt.Sprintf("give it with --answer %s=<value>", name)))
			continue
		}
		given[name] = value
	}
	if err := join(errs); err != nil {
		return nil, err
	}
	return given, nil
}

// Pending gives the questions of <sel> with no answer, in the order of the
// questions: each question with no answer in <given> and in the files, each
// question of a group with no answer that is not empty, and the last
// question of such a group. With <only>, it gives only the questions of
// those tools.
func Pending(sel *Selected, given map[string]string, only map[string]bool) []questions.Question {
	all := questions.Of(sel.Idx, sel.Tools)
	filled := func(q questions.Question) bool {
		if v, ok := given[q.Name()]; ok {
			return v != ""
		}
		v, ok := questions.Answer(sel.User, sel.Over, q)
		return ok && !questions.Empty(v)
	}
	open := map[string]bool{}
	last := map[string]int{}
	for i, qq := range all {
		if qq.Group == "" {
			continue
		}
		empty := true
		for _, m := range questions.Members(all, qq) {
			if filled(m) {
				empty = false
			}
		}
		open[qq.GroupName()] = empty
		last[qq.GroupName()] = i
	}
	var out []questions.Question
	for i, qq := range all {
		if only != nil && !only[qq.Tool] {
			continue
		}
		inOpen := qq.Group != "" && open[qq.GroupName()]
		_, flagged := given[qq.Name()]
		_, inFiles := questions.Answer(sel.User, sel.Over, qq)
		if (!flagged && (!inFiles || inOpen)) || (inOpen && last[qq.GroupName()] == i) {
			out = append(out, qq)
		}
	}
	return out
}
