package tui

import "testing"

func TestFit(t *testing.T) {
	for _, c := range []struct {
		in   string
		w    int
		want string
	}{
		{"abc", 5, "abc  "},
		{"abcde", 5, "abcde"},
		{"abcdef", 5, "abcd…"},
		{"éééééé", 3, "éé…"},
		{"abc", 0, ""},
	} {
		if got := fit(c.in, c.w); got != c.want {
			t.Errorf("fit(%q, %d) = %q, want %q", c.in, c.w, got, c.want)
		}
	}
}

func TestCell(t *testing.T) {
	long := "abcdefgh"
	// A row without the cursor, a short text and the hold: the cut text.
	for _, tick := range []int{-1, 0, holdTicks - 1} {
		if got := cell(long, 5, tick); got != "abcd…" {
			t.Errorf("cell at tick %d = %q", tick, got)
		}
	}
	if got := cell("abc", 5, 30); got != "abc  " {
		t.Errorf("a short text scrolls: %q", got)
	}
	// After the hold, the text moves one character for each tick, and comes
	// back after its end and the gap.
	if got := cell(long, 5, holdTicks); got != "bcdef" {
		t.Errorf("first step = %q", got)
	}
	if got := cell(long, 5, holdTicks+len(long)); got != "  abc" {
		t.Errorf("gap = %q", got)
	}
	period := holdTicks + len(long) + loopGap
	if got := cell(long, 5, period); got != "abcd…" {
		t.Errorf("a new hold = %q", got)
	}
}

func TestPaneMove(t *testing.T) {
	p := pane{}
	for range 20 {
		p.rows = append(p.rows, nil)
	}
	p.move(15)
	if p.cursor != 15 || p.top != 15-paneH+1 {
		t.Errorf("cursor %d top %d", p.cursor, p.top)
	}
	p.move(-100)
	if p.cursor != 0 || p.top != 0 {
		t.Errorf("cursor %d top %d", p.cursor, p.top)
	}
	p.move(100)
	if p.cursor != 19 {
		t.Errorf("cursor %d", p.cursor)
	}
}

func TestWrap(t *testing.T) {
	got := wrap("one two three  four", 9)
	want := []string{"one two", "three", "four"}
	if len(got) != len(want) {
		t.Fatalf("wrap = %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("wrap = %q, want %q", got, want)
		}
	}
	if got := wrap("", 5); len(got) != 0 {
		t.Errorf("wrap of nothing = %q", got)
	}
}
