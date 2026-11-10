package macos

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/target"
)

func TestLookPath(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	for _, d := range []string{a, b, filepath.Join(a, "mise")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A folder and a file that is not executable do not count.
	if err := os.WriteFile(filepath.Join(b, "mise"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := lookPath("mise", a+":"+b); got != "" {
		t.Errorf("lookPath = %q, want none", got)
	}
	if err := os.Chmod(filepath.Join(b, "mise"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := lookPath("mise", a+":"+b); got != filepath.Join(b, "mise") {
		t.Errorf("lookPath = %q, want %q", got, filepath.Join(b, "mise"))
	}
}

func TestCheckMise(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(dir, "home", ".local", "bin")
	other := filepath.Join(dir, "brew", "bin")
	linked := filepath.Join(dir, "link", ".local", "bin")
	for _, d := range []string{good, other, linked} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{good, other} {
		if err := os.WriteFile(filepath.Join(d, "mise"), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(other, "mise"), filepath.Join(linked, "mise")); err != nil {
		t.Fatal(err)
	}
	const msg = "bootstrap: mise must be installed with `curl https://mise.run | sh` at ~/.local/bin/mise, not from a package manager (found: "
	cases := []struct{ path, want string }{
		{good, ""},
		{other, msg + filepath.Join(other, "mise") + ")\n"},
		{linked, msg + filepath.Join(other, "mise") + ")\n"},
		{filepath.Join(dir, "none"), msg + "no mise on PATH)\n"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		r := report.NewReport(&buf, report.Options{Prefix: Prefix})
		s := &state{r: r, e: &target.Env{Report: r}, pwd: dir,
			getenv: func(k string) string {
				if k == "PATH" {
					return c.path
				}
				return ""
			}}
		code := s.checkMise()
		if buf.String() != c.want || (code == 0) != (c.want == "") {
			t.Errorf("PATH=%s: code %d, output %q, want %q", c.path, code, buf.String(), c.want)
		}
	}
}

func TestUsage(t *testing.T) {
	var out, errs bytes.Buffer
	if code := Run([]string{"--no-bootstrap", "--help", "--bogus"}, &out, &errs); code != 0 || out.String() != Usage || errs.Len() != 0 {
		t.Errorf("help: code %d, stdout %q, stderr %q", code, out.String(), errs.String())
	}
	out.Reset()
	if code := Run([]string{"--dry-run", "--bogus", "-h"}, &out, &errs); code != 2 || out.Len() != 0 ||
		errs.String() != "bootstrap: unknown flag '--bogus'\n"+Usage {
		t.Errorf("bad flag: code %d, stdout %q, stderr %q", code, out.String(), errs.String())
	}
	if !strings.HasPrefix(Usage, "usage: bootstrap mac sync [--dry-run] [--no-bootstrap]\n") {
		t.Errorf("usage: %q", Usage)
	}
}
