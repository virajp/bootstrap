// Package field is the text field of a question: the question, a note on a
// line under it, and a field that starts with a proposal. The user edits or
// deletes the text, and Enter accepts it. A necessary field stays open on an
// empty text. It knows no command: init and the TUI give it the texts.
//
// Contract: docs/design/cli.md.
package field

import (
	"errors"
	"io"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// ErrInterrupted is the error of a field that the user stopped (ctrl+c).
var ErrInterrupted = errors.New("interrupted")

// Necessary is the line of a necessary field with an empty text.
const Necessary = "an answer is necessary"

// Spec is the text of a field.
type Spec struct {
	Question  string // The question line.
	Note      string // A line under the question, or empty.
	Proposal  string // The start text of the field.
	Necessary bool   // An empty text keeps the field open.
}

// Model is the Bubble Tea model of one field.
type Model struct {
	spec        Spec
	input       textinput.Model
	need        bool // The last Enter had an empty text on a necessary field.
	done        bool
	interrupted bool
}

// New gives the model of <s>, with the field started with its proposal.
func New(s Spec) Model {
	in := textinput.New()
	in.Prompt = "  "
	in.SetValue(s.Proposal)
	in.CursorEnd()
	in.Focus()
	return Model{spec: s, input: in}
}

// Init is a part of tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update is a part of tea.Model: Enter accepts the text, ctrl+c stops the
// field, and each other key goes to the text field.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.InterruptMsg:
		m.interrupted = true
		return m, tea.Quit
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			m.interrupted = true
			return m, tea.Quit
		case "enter":
			if m.spec.Necessary && m.Value() == "" {
				m.need = true
				return m, nil
			}
			m.done = true
			m.input.Blur()
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.Value() != "" {
		m.need = false
	}
	return m, cmd
}

// View is a part of tea.Model.
func (m Model) View() tea.View {
	var b strings.Builder
	b.WriteString("? " + m.spec.Question + "\n")
	if m.spec.Note != "" {
		b.WriteString("  " + m.spec.Note + "\n")
	}
	if m.done {
		b.WriteString("  " + m.Value() + "\n")
		return tea.NewView(b.String())
	}
	b.WriteString(m.input.View() + "\n")
	if m.need {
		b.WriteString("  " + Necessary + "\n")
	}
	return tea.NewView(b.String())
}

// Value gives the text of the field, without the spaces at its ends.
func (m Model) Value() string { return strings.TrimSpace(m.input.Value()) }

// Done tells if Enter accepted the text.
func (m Model) Done() bool { return m.done }

// Interrupted tells if the user stopped the field.
func (m Model) Interrupted() bool { return m.interrupted }

// Need tells if the field shows the line of a necessary field.
func (m Model) Need() bool { return m.need }

// Ask runs the field of <s> on <in> and <out>, and gives the accepted text,
// or ErrInterrupted.
func Ask(in io.Reader, out io.Writer, s Spec) (string, error) {
	p := tea.NewProgram(New(s), tea.WithInput(in), tea.WithOutput(out))
	end, err := p.Run()
	if errors.Is(err, tea.ErrInterrupted) {
		return "", ErrInterrupted
	}
	if err != nil {
		return "", err
	}
	m, ok := end.(Model)
	if !ok || !m.done {
		return "", ErrInterrupted
	}
	return m.Value(), nil
}
