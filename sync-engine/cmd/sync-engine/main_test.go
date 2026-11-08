package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
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
		{"version argument", []string{"version", "x"}, fail.Usage, "", "sync-engine: version: unknown argument x\n"},
		{"no command", nil, fail.Usage, "", "sync-engine: no command\nusage:"},
		{"unknown command", []string{"sync"}, fail.Usage, "", "sync-engine: unknown command sync\n"},
		{"help", []string{"--help"}, fail.OK, "usage:", ""},
		{"repo", []string{"repo"}, fail.Other, "", "sync-engine: repo: not implemented\n"},
		{"repo dry run", []string{"repo", "--dry-run"}, fail.Other, "", "sync-engine: repo: not implemented\n"},
		{"repo check", []string{"repo", "--check"}, fail.Other, "", "sync-engine: repo: not implemented\n"},
		{"repo check and dry run", []string{"repo", "--check", "--dry-run"}, fail.Usage, "", "sync-engine: repo: --check and --dry-run"},
		{"repo bad flag", []string{"repo", "--no-bootstrap"}, fail.Usage, "", "sync-engine: repo: flag provided but not defined"},
		{"repo argument", []string{"repo", "x"}, fail.Usage, "", "sync-engine: repo: unknown argument x\n"},
		{"repo help", []string{"repo", "-h"}, fail.OK, "usage:", ""},
		{"macos", []string{"macos", "--dry-run", "--no-bootstrap"}, fail.Other, "", "sync-engine: macos: not implemented\n"},
		{"macos bad flag", []string{"macos", "--check"}, fail.Usage, "", "sync-engine: macos: flag provided but not defined"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run(c.args, &stdout, &stderr); code != c.code {
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
