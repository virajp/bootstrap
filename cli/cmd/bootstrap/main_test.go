package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		code   fail.Code
		stdout string // a prefix of stdout
		stderr string // a prefix of stderr
	}{
		{"version", []string{"version"}, fail.OK, "dev\n", ""},
		{"--version", []string{"--version"}, fail.OK, "dev\n", ""},
		{"mac --version", []string{"mac", "--version"}, fail.Usage, "", "bootstrap: unknown command mac --version\nusage:"},
		{"list argument", []string{"list", "x"}, fail.Usage, "", "bootstrap: unknown argument x\nusage:"},
		{"ls bad flag", []string{"ls", "--yes"}, fail.Usage, "", "bootstrap: unknown flag --yes\nusage:"},
		{"version argument", []string{"version", "x"}, fail.Usage, "", "bootstrap: unknown argument x\nusage:"},
		{"no command", nil, fail.Usage, "", "bootstrap: no command\nusage:"},
		{"unknown command", []string{"repo"}, fail.Usage, "", "bootstrap: unknown command repo\n"},
		{"help", []string{"--help"}, fail.OK, "usage:", ""},
		{"check and dry run", []string{"check", "--dry-run"}, fail.Usage, "", "bootstrap: unknown flag --dry-run\nusage:"},
		{"sync bad flag", []string{"sync", "--no-bootstrap"}, fail.Usage, "", "bootstrap: unknown flag --no-bootstrap\nusage:"},
		{"sync argument", []string{"sync", "x"}, fail.Usage, "", "bootstrap: unknown argument x\nusage:"},
		{"sync help", []string{"sync", "-h"}, fail.OK, "usage:", ""},
		{"init yes value", []string{"init", "--yes=1"}, fail.Usage, "", "bootstrap: --yes=1 takes no value\nusage:"},
		{"init answer without value", []string{"init", "--answer"}, fail.Usage, "", "bootstrap: --answer needs a value\nusage:"},
		{"jobs not a number", []string{"sync", "--jobs", "x"}, fail.Usage, "", "bootstrap: --jobs needs a number above 0, not 'x'\nusage:"},
		{"questions bad flag", []string{"questions", "--yes"}, fail.Usage, "", "bootstrap: unknown flag --yes\nusage:"},
		{"mac sync help", []string{"mac", "sync", "--dry-run", "-h"}, fail.OK, "usage:", ""},
		{"mac sync bad flag", []string{"mac", "sync", "--check"}, fail.Usage, "", "bootstrap: unknown flag --check\nusage:"},
		{"mac questions no bootstrap", []string{"mac", "questions", "--no-bootstrap"}, fail.Usage, "", "bootstrap: unknown flag --no-bootstrap\nusage:"},
		{"mac version", []string{"mac", "version"}, fail.Usage, "", "bootstrap: unknown command mac version\nusage:"},
		{"mac no command", []string{"mac"}, fail.Usage, "", "bootstrap: mac: no command\nusage:"},
		{"mac unknown command", []string{"mac", "macos"}, fail.Usage, "", "bootstrap: unknown command mac macos\nusage:"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", "1")
			var stdout, stderr bytes.Buffer
			if code := run(c.args, strings.NewReader(""), &stdout, &stderr); code != c.code {
				t.Errorf("code = %d, want %d", code, c.code)
			}
			if !strings.HasPrefix(stdout.String(), c.stdout) || (c.stdout == "" && stdout.Len() != 0) {
				t.Errorf("stdout = %q, want prefix %q", stdout.String(), c.stdout)
			}
			if !strings.HasPrefix(stderr.String(), c.stderr) || (c.stderr == "" && stderr.Len() != 0) {
				t.Errorf("stderr = %q, want prefix %q", stderr.String(), c.stderr)
			}
		})
	}
}
