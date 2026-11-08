package fail

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"
)

func TestCodeOf(t *testing.T) {
	plain := errors.New("plain")
	cases := []struct {
		name string
		err  error
		want Code
	}{
		{"nil", nil, OK},
		{"plain error", plain, Other},
		{"usage", New(Usage, "bad flag %s", "--x"), Usage},
		{"missing", New(Missing, "no tool"), Missing},
		{"terminal", New(Terminal, "open a new terminal"), Terminal},
		{"network", Wrap(Network, plain, "fetch"), Network},
		{"wrapped by fmt", fmt.Errorf("outer: %w", New(Usage, "inner")), Usage},
		{"joined", errors.Join(plain, New(Network, "a"), New(Usage, "b")), Network},
		{"joined plain", errors.Join(plain, plain), Other},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CodeOf(c.err); got != c.want {
				t.Errorf("CodeOf() = %d, want %d", got, c.want)
			}
		})
	}
}

func TestErrorText(t *testing.T) {
	cases := []struct {
		err  *Error
		want string
	}{
		{New(Usage, "bad flag: %s", "--x"), "bad flag: --x"},
		{Wrap(Other, fs.ErrNotExist, "cannot read a.toml"), "cannot read a.toml: file does not exist"},
		{Wrap(Other, fs.ErrNotExist, ""), "file does not exist"},
	}
	for _, c := range cases {
		if got := c.err.Error(); got != c.want {
			t.Errorf("Error() = %q, want %q", got, c.want)
		}
	}
}

func TestUnwrap(t *testing.T) {
	err := Wrap(Network, fs.ErrPermission, "clone")
	if !errors.Is(err, fs.ErrPermission) {
		t.Error("errors.Is does not find the wrapped error")
	}
}

func TestCollector(t *testing.T) {
	var c Collector
	if c.Code() != OK || c.Err() != nil {
		t.Fatalf("empty collector: code %d, err %v", c.Code(), c.Err())
	}
	c.Add(nil)
	if c.Code() != OK || c.Err() != nil {
		t.Fatalf("after nil: code %d, err %v", c.Code(), c.Err())
	}
	steps := []struct {
		err  error
		want Code
	}{
		{New(Usage, "a"), Usage},
		{errors.New("b"), Usage},
		{New(Network, "c"), Network},
		{New(Missing, "d"), Network},
	}
	for i, s := range steps {
		c.Add(s.err)
		if got := c.Code(); got != s.want {
			t.Errorf("step %d: Code() = %d, want %d", i, got, s.want)
		}
	}
	if got := CodeOf(c.Err()); got != Network {
		t.Errorf("CodeOf(Err()) = %d, want %d", got, Network)
	}
	if !errors.Is(c.Err(), steps[0].err) {
		t.Error("Err() does not hold the first failure")
	}
}
