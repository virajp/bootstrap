package source

import "testing"

func TestWhen(t *testing.T) {
	all := FromMap(map[string]any{"git": map[string]any{"pair": true, "n": 1, "sub": map[string]any{"off": false}}}).(*Obj)
	for path, want := range map[string]bool{
		"":             true,
		"git.pair":     true,
		"!git.pair":    false,
		"git.n":        false,
		"git.sub.off":  false,
		"!git.sub.off": true,
		"git.nope":     false,
		"!nope.x.y":    true,
	} {
		if got := When(all, path); got != want {
			t.Errorf("When(%q) = %v, want %v", path, got, want)
		}
	}
	if !When(nil, "!a") || When(nil, "a") {
		t.Error("When with no values")
	}
}

func TestParsePerm(t *testing.T) {
	for p, want := range map[string]int{"700": 0o700, "0644": 0o644, "4755": 0o4755} {
		if m, ok := ParsePerm(p); !ok || int(m) != want {
			t.Errorf("ParsePerm(%q) = %o, %v", p, m, ok)
		}
	}
	for _, p := range []string{"", "7", "800", "07777x", "abc"} {
		if _, ok := ParsePerm(p); ok {
			t.Errorf("ParsePerm(%q) ok", p)
		}
	}
}
