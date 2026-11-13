package field

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func press(m Model, keys ...tea.KeyPressMsg) Model {
	for _, k := range keys {
		next, _ := m.Update(k)
		m = next.(Model)
	}
	return m
}

func text(s string) []tea.KeyPressMsg {
	var out []tea.KeyPressMsg
	for _, r := range s {
		out = append(out, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return out
}

var (
	enter     = tea.KeyPressMsg{Code: tea.KeyEnter}
	backspace = tea.KeyPressMsg{Code: tea.KeyBackspace}
	ctrlC     = tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
)

func TestStartsWithProposal(t *testing.T) {
	m := New(Spec{Question: "Which name?", Proposal: "abc"})
	if m.Value() != "abc" {
		t.Fatalf("start text %q, want abc", m.Value())
	}
	m = press(m, enter)
	if !m.Done() || m.Value() != "abc" {
		t.Fatalf("done %v value %q", m.Done(), m.Value())
	}
}

func TestDeleteAndNewText(t *testing.T) {
	m := New(Spec{Question: "Which name?", Proposal: "abc"})
	m = press(m, backspace, backspace, backspace)
	m = press(m, text("xy")...)
	m = press(m, enter)
	if !m.Done() || m.Value() != "xy" {
		t.Fatalf("done %v value %q, want xy", m.Done(), m.Value())
	}
}

func TestNecessaryStaysOpen(t *testing.T) {
	m := New(Spec{Question: "Which name?", Proposal: "a", Necessary: true})
	m = press(m, backspace, enter)
	if m.Done() || !m.Need() {
		t.Fatalf("done %v need %v, want an open field", m.Done(), m.Need())
	}
	if !strings.Contains(m.View().Content, Necessary) {
		t.Fatalf("view %q has no %q", m.View().Content, Necessary)
	}
	m = press(m, text("b")...)
	m = press(m, enter)
	if !m.Done() || m.Value() != "b" {
		t.Fatalf("done %v value %q, want b", m.Done(), m.Value())
	}
}

func TestEmptyAccepted(t *testing.T) {
	m := New(Spec{Question: "Which name?", Proposal: "a"})
	m = press(m, backspace, enter)
	if !m.Done() || m.Value() != "" {
		t.Fatalf("done %v value %q, want an empty answer", m.Done(), m.Value())
	}
}

func TestInterrupt(t *testing.T) {
	m := press(New(Spec{Question: "Which name?"}), ctrlC)
	if !m.Interrupted() || m.Done() {
		t.Fatalf("interrupted %v done %v", m.Interrupted(), m.Done())
	}
}

func TestViewHasNote(t *testing.T) {
	v := New(Spec{Question: "Which name?", Note: "a note"}).View().Content
	if !strings.Contains(v, "? Which name?\n  a note\n") {
		t.Fatalf("view %q", v)
	}
}
