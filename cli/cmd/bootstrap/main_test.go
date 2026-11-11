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
		{"mac --version", []string{"mac", "--version"}, fail.Usage, "", "[ERROR] mac --version is not a command\n  -> Run bootstrap --help"},
		{"list argument", []string{"list", "x"}, fail.Usage, "", "[ERROR] list takes no argument x\n  usage: bootstrap list\n  -> Run"},
		{"ls bad flag", []string{"ls", "--yes"}, fail.Usage, "", "[ERROR] list has no flag --yes\n  usage: bootstrap list\n  -> Run"},
		{"version argument", []string{"version", "x"}, fail.Usage, "", "[ERROR] version takes no argument x\n  usage: bootstrap version\n  -> Run"},
		{"no command", nil, fail.Usage, "", "[ERROR] no command given\n  -> Run bootstrap --help"},
		{"unknown command", []string{"repo"}, fail.Usage, "", "[ERROR] repo is not a command\n  -> Run bootstrap --help"},
		{"help", []string{"--help"}, fail.OK, "usage:", ""},
		{"check and dry run", []string{"check", "--dry-run"}, fail.Usage, "", "[ERROR] check has no flag --dry-run\n  usage: bootstrap check [--jobs <n>] [--verbose]\n  -> Run"},
		{"sync bad flag", []string{"sync", "--no-bootstrap"}, fail.Usage, "", "[ERROR] sync has no flag --no-bootstrap\n  usage: bootstrap sync [--dry-run] [--jobs <n>] [--verbose]\n  -> Run"},
		{"sync argument", []string{"sync", "x"}, fail.Usage, "", "[ERROR] sync takes no argument x\n  usage: bootstrap sync [--dry-run] [--jobs <n>] [--verbose]\n  -> Run"},
		{"sync help", []string{"sync", "-h"}, fail.OK, "usage:", ""},
		{"init yes value", []string{"init", "--yes=1"}, fail.Usage, "", "[ERROR] --yes=1 takes no value\n  usage: bootstrap init"},
		{"init answer without value", []string{"init", "--answer"}, fail.Usage, "", "[ERROR] --answer needs a value\n  usage: bootstrap init"},
		{"jobs not a number", []string{"sync", "--jobs", "x"}, fail.Usage, "", "[ERROR] --jobs needs a number above 0, not x\n  usage: bootstrap sync"},
		{"questions bad flag", []string{"questions", "--yes"}, fail.Usage, "", "[ERROR] questions has no flag --yes\n  usage: bootstrap questions [tool...] [--json]\n  -> Run"},
		{"mac sync help", []string{"mac", "sync", "--dry-run", "-h"}, fail.OK, "usage:", ""},
		{"mac sync bad flag", []string{"mac", "sync", "--check"}, fail.Usage, "", "[ERROR] mac sync has no flag --check\n  usage: bootstrap mac sync [--dry-run] [--jobs <n>] [--verbose]\n  -> Run"},
		{"mac questions no bootstrap", []string{"mac", "questions", "--no-bootstrap"}, fail.Usage, "", "[ERROR] mac questions has no flag --no-bootstrap\n  usage: bootstrap mac questions [tool...] [--json]\n  -> Run"},
		{"mac version", []string{"mac", "version"}, fail.Usage, "", "[ERROR] mac version is not a command\n  -> Run bootstrap --help"},
		{"mac no command", []string{"mac"}, fail.Usage, "", "[ERROR] mac needs a command\n  -> Run bootstrap --help"},
		{"mac unknown command", []string{"mac", "macos"}, fail.Usage, "", "[ERROR] mac macos is not a command\n  -> Run bootstrap --help"},
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
