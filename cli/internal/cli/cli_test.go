package cli

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

func TestNormalizeURL(t *testing.T) {
	same := []string{
		"https://github.com/virajp/bootstrap.git",
		"https://github.com/virajp/bootstrap",
		"https://user@github.com/virajp/bootstrap/",
		"git@github.com:virajp/bootstrap.git",
		"ssh://git@github.com/virajp/bootstrap.git",
	}
	for _, u := range same {
		if got := normalizeURL(u); got != "github.com/virajp/bootstrap" {
			t.Errorf("normalizeURL(%q) = %q", u, got)
		}
	}
	if got := normalizeURL("/tmp/remote"); got != "/tmp/remote" {
		t.Errorf("a path = %q", got)
	}
}

func TestCheckVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir+"/cli", 0o777); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		version, need string
		code          fail.Code
		msg           string
	}{
		{"dev", "1.2.3", fail.OK, ""},
		{"1.2.3", "1.2.3", fail.OK, ""},
		{"1.10.0", "1.9.9", fail.OK, ""},
		{"2.0.0", "1.9.9", fail.OK, ""},
		{"1.2.2", "1.2.3", fail.Usage, "bootstrap 1.2.2 is older than 1.2.3"},
		{"0.9.0", "1.0.0", fail.Usage, "upgrade it with: mise use -g github:virajp/bootstrap"},
		{"1.0.0", "v1.2", fail.Network, "is not MAJOR.MINOR.PATCH ('v1.2')"},
	}
	for _, c := range cases {
		if err := os.WriteFile(dir+"/cli/VERSION", []byte(c.need+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		var stderr bytes.Buffer
		a := &app{version: c.version, stderr: &stderr}
		if code := a.checkVersion(dir); code != c.code {
			t.Errorf("%s with %s: code %d, want %d", c.version, c.need, code, c.code)
		}
		if !strings.Contains(stderr.String(), c.msg) {
			t.Errorf("%s with %s: stderr %q lacks %q", c.version, c.need, stderr.String(), c.msg)
		}
	}
}

func TestStaleLock(t *testing.T) {
	lock := t.TempDir() + "/lock"
	if err := os.Mkdir(lock, 0o777); err != nil {
		t.Fatal(err)
	}
	// A pid that is not alive: a process id far above the usual range.
	if err := os.WriteFile(lock+"/pid", []byte("999999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !staleLock(lock) || exists(lock) {
		t.Error("the lock of a dead holder stays")
	}
	if err := os.Mkdir(lock, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lock+"/pid", []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if staleLock(lock) || !exists(lock) {
		t.Error("the lock of a live holder is removed")
	}
}

func TestRenderTree(t *testing.T) {
	deps := map[string][]string{
		"gitleaks":   {"_sec"},
		"grype":      {"_sec"},
		"linter":     {"node"},
		"pre-commit": {"python"},
		"python":     {"uv"},
		"uv":         {"python"}, // a loop of depends
	}
	required := func(t string) bool { return t == "git" || t == "pre-commit" }
	roots := []listRoot{
		{"base (pack)", []string{"pre-commit", "git", "gitleaks", "grype", "linter"}},
		{"extra (pack)", []string{"zed"}},
		{"tools (selected directly)", []string{"git", "claude"}},
	}
	want := `base (pack)
├── git         [required]
├── gitleaks
│   └── _sec
├── grype
│   └── _sec
├── linter
│   └── node
└── pre-commit  [required]
    └── python
        └── uv
extra (pack)
└── zed
tools (selected directly)
├── claude
└── git         [required]
`
	if got := renderTree(roots, deps, required); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	if got := renderTree(roots[:2], deps, required); strings.Contains(got, "tools (selected") {
		t.Errorf("a tools root without tools: %s", got)
	}
}
