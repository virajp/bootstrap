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
		{"version argument", []string{"version", "x"}, fail.Usage, "", "bootstrap: version: unknown argument x\n"},
		{"no command", nil, fail.Usage, "", "bootstrap: no command\nusage:"},
		{"unknown command", []string{"repo"}, fail.Usage, "", "bootstrap: unknown command repo\n"},
		{"help", []string{"--help"}, fail.OK, "usage:", ""},
		{"check and dry run", []string{"check", "--dry-run"}, fail.Usage, "", "repo-sync: usage: repo_sync [--dry-run | --check]\n"},
		{"sync check and dry run", []string{"sync", "--check", "--dry-run"}, fail.Usage, "", "repo-sync: usage: repo_sync [--dry-run | --check]\n"},
		{"sync bad flag", []string{"sync", "--no-bootstrap"}, fail.Usage, "", "repo-sync: usage: repo_sync [--dry-run | --check]\n"},
		{"sync argument", []string{"sync", "x"}, fail.Usage, "", "repo-sync: usage: repo_sync [--dry-run | --check]\n"},
		{"sync help", []string{"sync", "-h"}, fail.Usage, "", "repo-sync: usage: repo_sync [--dry-run | --check]\n"},
		{"mac sync help", []string{"mac", "sync", "--dry-run", "-h"}, fail.OK, "usage: setup_sync [--dry-run] [--no-bootstrap]\n", ""},
		{"mac check help", []string{"mac", "check", "-h"}, fail.OK, "usage: setup_sync [--dry-run] [--no-bootstrap]\n", ""},
		{"mac sync bad flag", []string{"mac", "sync", "--check"}, fail.Usage, "", "setup-sync: unknown flag '--check'\nusage: setup_sync"},
		{"mac no command", []string{"mac"}, fail.Usage, "", "bootstrap: mac: no command\nusage:"},
		{"mac unknown command", []string{"mac", "macos"}, fail.Usage, "", "bootstrap: mac: unknown command macos\nusage:"},
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
