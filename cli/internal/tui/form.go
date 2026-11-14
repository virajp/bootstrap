package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/questions"
)

// The questions form (screen 7 of the design): a modal dialog with one field
// for each question of a tool, the note under a question, "optional" next
// to an optional field, the line of a group, and the buttons Accept and
// Cancel. A field starts with the current answer, else with the output of
// the detect command of its question; an example shows dim in an empty
// field. Accept puts the answers on the change list; nothing is written
// before Apply (ruling 36).

// The widths of the form.
const (
	formW  = 86 // The dialog.
	fieldW = 62 // The text of a field.
)

// The line of a field that has no answer, and the reason of a refused
// Accept (ruling 43).
const needText = "An answer is necessary"

// formField is one field of the form: its question and its text.
type formField struct {
	q      questions.Question
	in     textinput.Model
	tag    string // The tag next to the field: optional, detected or empty.
	detect string // The output of the detect command, or empty.
	need   bool   // Accept refused the field: it shows needText.
}

// form is the questions form of one tool. focus is a field, or the button
// Accept (len(fields)) or Cancel (len(fields)+1).
type form struct {
	title  string
	tool   string
	fields []formField
	focus  int
	// done takes the answers of an Accept; cancel follows a Cancel.
	done   func(m *Model, answers []core.Answer) tea.Cmd
	cancel func(m *Model) tea.Cmd
}

// newForm gives the form of the questions <qs> of <tool>. <current> gives
// the start text of a question that has one.
func (m *Model) newForm(title, tool string, qs []questions.Question, current func(questions.Question) (string, bool)) *form {
	f := &form{title: title, tool: tool}
	dir := m.state.Home()
	for _, q := range qs {
		in := textinput.New()
		in.Prompt = ""
		in.SetWidth(fieldW)
		fd := formField{q: q}
		v, ok := current(q)
		if p, from := questions.Proposal(q, dir); from == "detect" {
			fd.detect = p
			if !ok {
				v = p
			}
		} else if from == "example" {
			in.Placeholder = "example: " + p
		}
		in.SetValue(v)
		in.CursorEnd()
		fd.in = in
		switch {
		case q.Optional:
			fd.tag = "optional"
		case fd.detect != "" && v == fd.detect:
			fd.tag = "detected"
		case v == "":
			fd.tag = "empty"
		}
		f.fields = append(f.fields, fd)
	}
	f.setFocus(0)
	return f
}

// setFocus moves the focus to <i>: a field, Accept or Cancel.
func (f *form) setFocus(i int) {
	n := len(f.fields) + 2
	f.focus = (i%n + n) % n
	for k := range f.fields {
		if k == f.focus {
			f.fields[k].in.Focus()
		} else {
			f.fields[k].in.Blur()
		}
	}
}

// value gives the text of field <i>, without the spaces at its ends.
func (f *form) value(i int) string { return strings.TrimSpace(f.fields[i].in.Value()) }

// body gives the lines of the dialog, and the first line of each field.
func (f *form) body() ([]line, []int) {
	var out []line
	starts := make([]int, len(f.fields))
	groups := map[string]bool{}
	for i, fd := range f.fields {
		q := fd.q
		if q.Group != "" && !groups[q.Group] {
			groups[q.Group] = true
			n := 0
			for _, o := range f.fields {
				if o.q.Group == q.Group {
					n++
				}
			}
			out = append(out, line{styled(fmt.Sprintf("At least one of these %d answers is necessary.", n), stLock)}, line{})
		}
		starts[i] = len(out)
		for _, l := range wrap(q.Text, formW-5) {
			out = append(out, line{text(l)})
		}
		out = append(out, f.fieldLine(i))
		if fd.need {
			out = append(out, line{styled(needText, stRemove)})
		}
		for _, l := range wrap(q.Note, formW-5) {
			out = append(out, line{styled(l, stLock)})
		}
		out = append(out, line{})
	}
	return out, starts
}

// fieldLine gives the line of the text of field <i> with its tag. The
// field with the focus has the background of the cursor and the caret ▏.
func (f *form) fieldLine(i int) line {
	fd := f.fields[i]
	tag := styled(fd.tag, stDim)
	v := fd.in.Value()
	if i != f.focus {
		if v == "" {
			return line{styled(fit(fd.in.Placeholder, fieldW), stDim), text(" "), tag}
		}
		return line{text(fit(v, fieldW)), text(" "), tag}
	}
	r := []rune(v)
	pos := min(fd.in.Position(), len(r))
	shown := string(r[:pos]) + "▏" + string(r[pos:])
	// A long text shows the part before the caret.
	if s := []rune(shown); len(s) > fieldW {
		start := max(0, min(pos+1-fieldW, len(s)-fieldW))
		shown = string(s[start : start+fieldW])
	}
	return line{styled(fit(shown, fieldW), stButton), text(" "), tag}
}

// buttons gives the line of the buttons Accept and Cancel.
func (f *form) buttons() line {
	n := len(f.fields)
	return line{text(" "), button("Accept", f.focus == n || f.focus < n), text("   "), button("Cancel", f.focus == n+1)}
}

// button gives the button <label>; <on> shows it with the background of
// the cursor.
func button(label string, on bool) seg {
	if on {
		return styled(" "+label+" ", stButton)
	}
	return text(" " + label + " ")
}

// check tells if Accept can take the answers: each necessary field has a
// text, and each group has one text in the form or one answer outside it.
// It marks each field that needs an answer, and moves the focus to the
// first one.
func (f *form) check(m *Model) bool {
	first := -1
	for i := range f.fields {
		f.fields[i].need = false
	}
	for i, fd := range f.fields {
		q := fd.q
		need := false
		switch {
		case q.Group != "":
			need = !f.groupFilled(m, q)
		case !q.Optional:
			need = f.value(i) == ""
		}
		if need {
			f.fields[i].need = true
			if first < 0 {
				first = i
			}
		}
	}
	if first >= 0 {
		f.setFocus(first)
		return false
	}
	return true
}

// groupFilled tells if the group of <q> has a text in the form, or an answer
// that is not empty outside the form.
func (f *form) groupFilled(m *Model, q questions.Question) bool {
	var in []string
	for i, fd := range f.fields {
		if fd.q.Group == q.Group && fd.q.Tool == q.Tool {
			in = append(in, fd.q.Name())
			if f.value(i) != "" {
				return true
			}
		}
	}
	for _, o := range questions.Members(questions.Of(m.state.Sel.Idx, []string{q.Tool}), q) {
		if slices.Contains(in, o.Name()) {
			continue
		}
		if v, ok := m.answerOf(o); ok && v != "" {
			return true
		}
	}
	return false
}

// answers gives the answers of the form.
func (f *form) answers() []core.Answer {
	out := make([]core.Answer, len(f.fields))
	for i, fd := range f.fields {
		out[i] = core.Answer{Name: fd.q.Name(), Value: f.value(i)}
	}
	return out
}

// showForm shows the form <f> as the modal dialog, with the field of the
// focus in view.
func (m *Model) showForm(f *form) {
	body, starts := f.body()
	top := 0
	if m.modal != nil && m.form == f {
		top = m.modal.top
	}
	d := &modal{
		title:   f.title,
		width:   formW,
		body:    body,
		buttons: f.buttons(),
		keys:    "tab next field  shift ↑↓ scroll  enter accept  esc cancel",
		onKey:   func(m *Model, k tea.KeyPressMsg) tea.Cmd { return m.formKey(f, k) },
	}
	d.top = top
	if f.focus < len(f.fields) {
		first := starts[f.focus]
		last := len(body)
		if f.focus+1 < len(starts) {
			last = starts[f.focus+1]
		}
		if last > d.top+d.rows() {
			d.top = last - d.rows()
		}
		if first < d.top {
			d.top = first
		}
	}
	d.scroll(0)
	m.form, m.modal = f, d
}

// formKey handles a key of the form <f>.
func (m *Model) formKey(f *form, k tea.KeyPressMsg) tea.Cmd {
	n := len(f.fields)
	switch k.String() {
	case "esc", "ctrl+c":
		m.form, m.modal = nil, nil
		return f.cancel(m)
	case "tab", "down":
		f.setFocus(f.focus + 1)
	case "shift+tab", "up":
		f.setFocus(f.focus - 1)
	case "shift+up":
		m.modal.scroll(-1)
		return nil
	case "shift+down":
		m.modal.scroll(1)
		return nil
	case "enter":
		if f.focus == n+1 {
			m.form, m.modal = nil, nil
			return f.cancel(m)
		}
		if !f.check(m) {
			break
		}
		m.form, m.modal = nil, nil
		return f.done(m, f.answers())
	case "left", "right":
		if f.focus >= n {
			f.setFocus(n + (f.focus-n+1)%2)
			break
		}
		return m.formInput(f, k)
	default:
		if f.focus < n {
			return m.formInput(f, k)
		}
	}
	m.showForm(f)
	return nil
}

// formInput gives the message <msg> to the field with the focus.
func (m *Model) formInput(f *form, msg tea.Msg) tea.Cmd {
	if f.focus >= len(f.fields) {
		return nil
	}
	fd := &f.fields[f.focus]
	var cmd tea.Cmd
	fd.in, cmd = fd.in.Update(msg)
	if f.value(f.focus) != "" {
		fd.need = false
	}
	m.showForm(f)
	return cmd
}

// The answers of the change list.

// answerOf gives the answer of <q>: on the change list, else in the files.
func (m *Model) answerOf(q questions.Question) (string, bool) {
	for i := len(m.list.Answers) - 1; i >= 0; i-- {
		if a := m.list.Answers[i]; a.Name == q.Name() {
			return a.Value, true
		}
	}
	v, ok := questions.Answer(m.state.Sel.User, m.state.Sel.Over, q)
	if !ok {
		return "", false
	}
	if s, isText := v.(string); isText {
		return s, true
	}
	return fmt.Sprint(v), true
}

// given gives the answers of the change list by the name of their question.
func (m *Model) given() map[string]string {
	out := map[string]string{}
	for _, a := range m.list.Answers {
		out[a.Name] = a.Value
	}
	return out
}

// setAnswers puts <answers> on the change list, in place of the answers of
// the same questions.
func (m *Model) setAnswers(answers []core.Answer) {
	for _, a := range answers {
		m.list.Answers = slices.DeleteFunc(m.list.Answers, func(o core.Answer) bool { return o.Name == a.Name })
		m.list.Answers = append(m.list.Answers, a)
	}
}

// answerTool gives the tool of the answer <a>.
func answerTool(a core.Answer) string {
	t, _, _ := strings.Cut(a.Name, ".")
	return t
}

// pruneAnswers takes the answers and the edits of each tool that is not
// selected after the change list off the list.
func (m *Model) pruneAnswers() {
	m.list.Answers = slices.DeleteFunc(m.list.Answers, func(a core.Answer) bool { return !m.selected(answerTool(a)) })
	m.edits = slices.DeleteFunc(m.edits, func(t string) bool { return !m.selected(t) })
}

// pendingOf gives the questions of the tools <tools> with no answer in the
// files and on the change list, one list for each tool, in the order of the
// tools.
func (m *Model) pendingOf(tools []string) [][]questions.Question {
	sel := *m.state.Sel
	sel.Tools = tools
	return byTool(core.Pending(&sel, m.given(), nil))
}

// byTool gives the questions <qs> in one list for each tool, in their order.
func byTool(qs []questions.Question) [][]questions.Question {
	var out [][]questions.Question
	for _, q := range qs {
		if n := len(out); n > 0 && out[n-1][0].Tool == q.Tool {
			out[n-1] = append(out[n-1], q)
			continue
		}
		out = append(out, []questions.Question{q})
	}
	return out
}

// askEach opens one form for each list of <groups>, one after the other.
// <title> gives the title of the form of a tool. Each Accept goes to
// <accept>; after the last one, <done> runs. A Cancel runs <cancel>, and
// the forms after it do not open.
func (m *Model) askEach(groups [][]questions.Question, title func(tool string, n int) string,
	accept func(m *Model, answers []core.Answer), done func(m *Model) tea.Cmd, cancel func(m *Model) tea.Cmd) tea.Cmd {
	if len(groups) == 0 {
		return done(m)
	}
	qs := groups[0]
	tool := qs[0].Tool
	f := m.newForm(title(tool, len(qs)), tool, qs, func(questions.Question) (string, bool) { return "", false })
	f.done = func(m *Model, answers []core.Answer) tea.Cmd {
		accept(m, answers)
		return m.askEach(groups[1:], title, accept, done, cancel)
	}
	f.cancel = cancel
	m.showForm(f)
	return nil
}

// selectHook adds the select <items>. When the tools that they add have
// questions with no answer, a form for each such tool asks them first.
// Cancel in a form cancels the select: no change goes on the list (D6).
func (m *Model) selectHook(items ...core.Item) tea.Cmd {
	eff, err := m.state.Effect(append(slices.Clone(m.list.Items), items...))
	if err != nil {
		m.modal = errModal(err)
		return nil
	}
	var tools []string
	for _, n := range m.state.Sel.Idx.Names {
		if eff.Added[n] && !m.selected(n) {
			tools = append(tools, n)
		}
	}
	groups := m.pendingOf(tools)
	if len(groups) == 0 {
		m.push(items...)
		return nil
	}
	var answers []core.Answer
	return m.askEach(groups,
		func(tool string, n int) string { return tool + " needs " + plural(n, "value") },
		func(_ *Model, a []core.Answer) { answers = append(answers, a...) },
		func(m *Model) tea.Cmd {
			if m.push(items...) {
				m.setAnswers(answers)
				m.refreshAll()
			}
			return nil
		},
		func(*Model) tea.Cmd { return nil })
}

// editForm opens the form of the tool under the cursor of the Tools tab
// with its current answers: a selected tool with questions (e). Cancel
// keeps the old answers (D6).
func (m *Model) editForm() tea.Cmd {
	r, ok := m.curTool()
	if !ok || r.group != "" || !m.selected(r.tool) {
		return nil
	}
	t := r.tool
	qs := questions.Of(m.state.Sel.Idx, []string{t})
	if len(qs) == 0 {
		return nil
	}
	f := m.newForm("The values of "+t, t, qs, m.answerOf)
	f.done = func(m *Model, answers []core.Answer) tea.Cmd {
		var changed []core.Answer
		for _, a := range answers {
			if v, ok := m.answerOf(questionOf(qs, a.Name)); !ok || v != a.Value {
				changed = append(changed, a)
			}
		}
		if len(changed) == 0 {
			return nil
		}
		m.setAnswers(changed)
		if !m.effect().Added[t] && !slices.Contains(m.edits, t) {
			m.edits = append(m.edits, t)
		}
		m.refreshAll()
		return nil
	}
	f.cancel = func(*Model) tea.Cmd { return nil }
	m.showForm(f)
	return nil
}

// questionOf gives the question of <qs> with the name <name>.
func questionOf(qs []questions.Question, name string) questions.Question {
	for _, q := range qs {
		if q.Name() == name {
			return q
		}
	}
	return questions.Question{}
}

// refreshAll builds the panes of each tab again.
func (m *Model) refreshAll() {
	m.refreshDetected()
	m.refreshTools()
	m.refreshPacks()
}
