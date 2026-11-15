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

func TestCatalogRule(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	offline := problem(fail.Network, "Cannot fetch u: nothing synced", "")
	cases := []struct {
		name    string
		age     time.Duration // The age of the last fetch; -1: no recorded time.
		usable  bool          // The cache clone exists with the origin of the URL.
		fetchOK bool
		fetched bool // The rule fetches.
		code    fail.Code
		warn    bool
		at      time.Time // The fetch time that the rule gives.
	}{
		{"10 minutes old: no fetch", 10 * time.Minute, true, false, false, fail.OK, false, now.Add(-10 * time.Minute)},
		{"30 minutes old: no fetch", 30 * time.Minute, true, false, false, fail.OK, false, now.Add(-30 * time.Minute)},
		{"2 hours old: fetch", 2 * time.Hour, true, true, true, fail.OK, false, now},
		{"no cache clone: fetch", 10 * time.Minute, false, true, true, fail.OK, false, now},
		{"fetch fails at 2 days: cache and warning", 2 * 24 * time.Hour, true, false, true, fail.OK, true, now.Add(-2 * 24 * time.Hour)},
		{"fetch fails at 8 days: code 5", 8 * 24 * time.Hour, true, false, true, fail.Network, false, time.Time{}},
		{"fetch fails with no cache clone: code 5", 2 * time.Hour, false, false, true, fail.Network, false, time.Time{}},
		{"no fetch time: fetch", -1, true, true, true, fail.OK, false, now},
		{"no fetch time and the fetch fails: code 5", -1, true, false, true, fail.Network, false, time.Time{}},
	}
	for _, c := range cases {
		cache := t.TempDir()
		if c.age >= 0 {
			stamp(cache, now.Add(-c.age))
		}
		fetched := false
		fetch := func() error {
			fetched = true
			if c.fetchOK {
				return nil
			}
			return offline
		}
		at, warn, err := catalogRule(cache, now, c.usable, fetch)
		if fetched != c.fetched {
			t.Errorf("%s: fetched %v", c.name, fetched)
		}
		if fail.CodeOf(err) != c.code {
			t.Errorf("%s: error %v", c.name, err)
		}
		if (warn != "") != c.warn {
			t.Errorf("%s: warning %q", c.name, warn)
		}
		if !at.Equal(c.at) {
			t.Errorf("%s: fetch time %v, want %v", c.name, at, c.at)
		}
		// A successful fetch records its time.
		if got, ok := fetchedAt(cache); c.fetchOK && (!ok || !got.Equal(now)) {
			t.Errorf("%s: recorded %v %v", c.name, got, ok)
		}
	}
}

func TestCloneCatalog(t *testing.T) {
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
	// A fetch of 10 minutes ago: no fetch, so the origin that does not
	// exist does not matter.
	stamp(cache, time.Now().Add(-10*time.Minute))
	c, err := acquireClone(CloneOptions{Getenv: getenv, Catalog: true})
	c.Cleanup()
	if err != nil || c.Fetched.IsZero() || c.Warning != "" {
		t.Fatalf("fresh: %v %+v", err, c)
	}
	// An old fetch fetches again; offline, the cache clone of 2 days ago
	// is used with a warning.
	stamp(cache, time.Now().Add(-2*24*time.Hour))
	c, err = acquireClone(CloneOptions{Getenv: getenv, Catalog: true})
	c.Cleanup()
	if err != nil || c.Warning == "" {
		t.Errorf("offline at 2 days: %v %+v", err, c)
	}
	// Without Catalog, each command fetches.
	stamp(cache, time.Now())
	c, err = acquireClone(CloneOptions{Getenv: getenv})
	c.Cleanup()
	if fail.CodeOf(err) != fail.Network {
		t.Errorf("no catalog: %v", err)
	}
	// BOOTSTRAP_CLONE: no fetch and no fetch time.
	env["BOOTSTRAP_CLONE"] = repo
	c, err = acquireClone(CloneOptions{Getenv: getenv, Catalog: true})
	c.Cleanup()
	if err != nil || c.Dir != repo || !c.Fetched.IsZero() {
		t.Errorf("BOOTSTRAP_CLONE: %v %+v", err, c)
	}
}
