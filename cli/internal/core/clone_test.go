package core

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

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
		err := CheckVersion(c.version, dir)
		if code := fail.CodeOf(err); code != c.code {
			t.Errorf("%s with %s: code %d, want %d", c.version, c.need, code, c.code)
		}
		got := ""
		if err != nil {
			got = err.Error()
		}
		if !strings.Contains(got, c.msg) {
			t.Errorf("%s with %s: error %q lacks %q", c.version, c.need, got, c.msg)
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
	if !staleLock(lock) || Exists(lock) {
		t.Error("the lock of a dead holder stays")
	}
	if err := os.Mkdir(lock, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lock+"/pid", []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if staleLock(lock) || !Exists(lock) {
		t.Error("the lock of a live holder is removed")
	}
}

func TestCloneAge(t *testing.T) {
	cache := t.TempDir()
	url := "file://" + cache + "/no-origin"
	repo := cache + "/repo"
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	if out, err := exec.Command("git", "-C", repo, "remote", "add", "origin", url).CombinedOutput(); err != nil {
		t.Fatalf("git remote: %s", out)
	}
	env := map[string]string{"BOOTSTRAP_CACHE": cache, "BOOTSTRAP_URL": url, "TMPDIR": t.TempDir()}
	getenv := func(k string) string { return env[k] }
	stamp := func(at time.Time) {
		if err := os.WriteFile(cache+"/"+FetchedName, []byte(at.UTC().Format(time.RFC3339)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A fetch younger than the age: no fetch, so the origin that does not
	// exist does not matter.
	stamp(time.Now().Add(-time.Hour))
	c, err := acquireClone(CloneOptions{Getenv: getenv, MaxAge: CatalogAge})
	if err != nil {
		t.Fatalf("fresh: %v", err)
	}
	c.Cleanup()
	// A stale fetch fetches again; offline, that is code 5.
	stamp(time.Now().Add(-8 * 24 * time.Hour))
	c, err = acquireClone(CloneOptions{Getenv: getenv, MaxAge: CatalogAge})
	c.Cleanup()
	if fail.CodeOf(err) != fail.Network {
		t.Errorf("stale: %v", err)
	}
	// Without MaxAge, each command fetches.
	stamp(time.Now())
	c, err = acquireClone(CloneOptions{Getenv: getenv})
	c.Cleanup()
	if fail.CodeOf(err) != fail.Network {
		t.Errorf("no age: %v", err)
	}
	if fresh(cache, 0) || !fresh(cache, time.Hour) {
		t.Error("fresh")
	}
}
