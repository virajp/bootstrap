package core

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// DefaultURL is the clone URL. BOOTSTRAP_URL changes it.
const DefaultURL = "https://github.com/virajp/bootstrap.git"

// CloneOptions are the options of OpenClone.
type CloneOptions struct {
	Getenv  func(string) string
	Version string // The version of this binary.
	// Hold keeps the lock of the cache held to Cleanup.
	Hold bool
	// Note tells each info line of the clone ("Waiting for another sync",
	// "Cloning into"), when it is not nil.
	Note func(string)
	// Stderr takes the output of git and cp on stderr. Nil discards it.
	Stderr io.Writer
}

// Clone is the bootstrap clone of one command: BOOTSTRAP_CLONE, else a
// temporary copy of the cache clone.
type Clone struct {
	Dir    string // The clone that the command reads.
	Cache  string
	Repo   string // The cache clone; empty with BOOTSTRAP_CLONE.
	lock   string
	url    string
	held   bool
	copy   string // The temporary copy, removed by Cleanup.
	note   func(string)
	stderr io.Writer
}

// OpenClone takes the clone of the command and checks its version, and sets
// BOOTSTRAP_CLONE to its folder. Without BOOTSTRAP_CLONE, it takes the lock
// of the cache, fetches main into the cache clone and copies it. A signal
// removes the lock and the copy. It always gives a clone: the caller defers
// Cleanup.
func OpenClone(o CloneOptions) (*Clone, error) {
	c, err := acquireClone(o)
	if err != nil {
		return c, err
	}
	if !o.Hold {
		c.Unlock()
	}
	// A signal removes the lock and the copy.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		got := <-sig
		c.Cleanup()
		if got == syscall.SIGINT {
			os.Exit(130)
		}
		os.Exit(143)
	}()
	if err := CheckVersion(o.Version, c.Dir); err != nil {
		return c, err
	}
	_ = os.Setenv("BOOTSTRAP_CLONE", c.Dir)
	return c, nil
}

// acquireClone gives the clone of the command. Without BOOTSTRAP_CLONE, it
// takes the lock of the cache, fetches main into the cache clone and copies
// it. The lock stays held: the caller releases it with Unlock.
func acquireClone(o CloneOptions) (*Clone, error) {
	c := &Clone{note: o.Note, stderr: o.Stderr}
	if c.note == nil {
		c.note = func(string) {}
	}
	if c.stderr == nil {
		c.stderr = io.Discard
	}
	if d := o.Getenv("BOOTSTRAP_CLONE"); d != "" {
		c.Dir = d
		return c, nil
	}
	c.Cache = o.Getenv("BOOTSTRAP_CACHE")
	if c.Cache == "" {
		xdg := o.Getenv("XDG_CACHE_HOME")
		if xdg == "" {
			xdg = o.Getenv("HOME") + "/.cache"
		}
		c.Cache = xdg + "/bootstrap"
	}
	c.Repo = c.Cache + "/repo"
	c.lock = c.Cache + "/lock"
	c.url = o.Getenv("BOOTSTRAP_URL")
	if c.url == "" {
		c.url = DefaultURL
	}
	if err := c.takeLock(); err != nil {
		return c, err
	}
	if err := c.fetch(); err != nil {
		c.Cleanup()
		return c, err
	}
	tmp := o.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	copyDir, err := os.MkdirTemp(tmp, "bootstrap-clone.")
	if err != nil {
		c.Cleanup()
		return c, problem(fail.Other, "cannot create a temp folder", "")
	}
	c.copy = copyDir
	if out, err := exec.Command("cp", "-R", "--", c.Repo+"/.", copyDir).CombinedOutput(); err != nil {
		_, _ = c.stderr.Write(out)
		c.Cleanup()
		return c, problem(fail.Other, fmt.Sprintf("cannot copy the clone %s", c.Repo), "")
	}
	c.Dir = copyDir
	return c, nil
}

// takeLock waits until this run holds the lock folder of the cache. A lock
// whose holder is dead, or a lock without a pid that is older than one
// minute, is removed.
func (c *Clone) takeLock() error {
	if err := os.MkdirAll(c.Cache, 0o777); err != nil {
		return problem(fail.Other, fmt.Sprintf("cannot create the cache %s", c.Cache), "")
	}
	waited := false
	for os.Mkdir(c.lock, 0o777) != nil {
		if staleLock(c.lock) {
			continue
		}
		pid := readPid(c.lock)
		if pid != "" && !waited {
			c.note(fmt.Sprintf("Waiting for another sync (pid %s) ...", pid))
			waited = true
		}
		time.Sleep(time.Second)
	}
	c.held = true
	_ = os.WriteFile(c.lock+"/pid", []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644)
	return nil
}

func readPid(lock string) string {
	b, err := os.ReadFile(lock + "/pid")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// olderThanMinute tells if <path> exists and its mtime is more than one
// minute old.
func olderThanMinute(path string) bool {
	st, err := os.Stat(path)
	return err == nil && time.Since(st.ModTime()) > time.Minute
}

// staleLock removes the lock of a dead holder and gives true. The removal
// has its own lock: without it, a second waiter can remove the lock that the
// first waiter took after the removal.
func staleLock(lock string) bool {
	take := lock + ".takeover"
	if os.Mkdir(take, 0o777) != nil {
		if olderThanMinute(take) {
			_ = os.RemoveAll(take)
		}
		return false
	}
	stale := false
	if pid := readPid(lock); pid != "" {
		if n, err := strconv.Atoi(pid); err == nil && n > 0 {
			if err := syscall.Kill(n, 0); err != nil && err != syscall.EPERM {
				stale = true
			}
		}
	} else if olderThanMinute(lock) {
		stale = true
	}
	if stale {
		_ = os.RemoveAll(lock)
	}
	_ = os.Remove(take)
	return stale
}

// Unlock releases the lock when this run holds it.
func (c *Clone) Unlock() {
	if c.held {
		_ = os.RemoveAll(c.lock)
		c.held = false
	}
}

// Cleanup releases the lock and removes the copy.
func (c *Clone) Cleanup() {
	c.Unlock()
	if c.copy != "" {
		_ = os.RemoveAll(c.copy)
		c.copy = ""
	}
}

// normalizeURL drops the scheme, the user name, the `user@host:path` form, a
// trailing `/` and a trailing `.git`, so the SSH and HTTPS URL of one
// repository match.
func normalizeURL(u string) string {
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
		host, _, _ := strings.Cut(u, "/")
		if strings.Contains(host, "@") {
			u = u[strings.Index(u, "@")+1:]
		}
	} else if before, _, ok := strings.Cut(u, ":"); ok && !strings.Contains(before, "/") {
		if i := strings.Index(u, "@"); i >= 0 {
			u = u[i+1:]
		}
		u = strings.Replace(u, ":", "/", 1)
	}
	u = strings.TrimSuffix(u, "/")
	u = strings.TrimSuffix(u, ".git")
	return strings.TrimSuffix(u, "/")
}

// gitQuiet runs git with <args>, with its stderr on the stderr writer of the
// clone.
func (c *Clone) gitQuiet(args ...string) bool {
	cmd := exec.Command("git", args...)
	cmd.Env = GitEnv()
	cmd.Stderr = c.stderr
	return cmd.Run() == nil
}

// fetch brings the cache clone to the newest commit of main: a clone when it
// is missing or empty, a fetch when its origin is the URL, a new clone for
// another origin. A folder that is not a clone is never deleted (code 2). A
// failed git call gives code 5.
func (c *Clone) fetch() error {
	// A clone of the URL at the cache root is the old layout. Any other
	// repository there stays.
	if Exists(c.Cache + "/.git") {
		origin, _ := Git("config", "--file", c.Cache+"/.git/config", "--get", "remote.origin.url")
		if origin == "" || normalizeURL(origin) != normalizeURL(c.url) {
			return problem(fail.Usage, fmt.Sprintf("%s is a git repository with the origin '%s', not an old clone of %s", c.Cache, origin, c.url), "move it, or set BOOTSTRAP_CACHE to another folder")
		}
		entries, _ := os.ReadDir(c.Cache)
		for _, e := range entries {
			switch e.Name() {
			case "repo", "lock", "lock.takeover", "macos-generated.tsv":
				continue
			}
			_ = os.RemoveAll(filepath.Join(c.Cache, e.Name()))
		}
		c.note(fmt.Sprintf("Removed an old clone at %s", c.Cache))
	}

	again := false
	st, err := os.Lstat(c.Repo)
	switch {
	case err != nil:
		again = true
	case st.IsDir() && isEmptyDir(c.Repo):
		again = true
	case st.IsDir() && Exists(c.Repo+"/.git"):
		origin, _ := Git("-C", c.Repo, "config", "--get", "remote.origin.url")
		if origin != "" && normalizeURL(origin) == normalizeURL(c.url) {
			// From the URL, not from origin: the compare drops the scheme,
			// and the transport must not come from the cached origin.
			if !c.gitQuiet("-C", c.Repo, "fetch", "--quiet", "--depth", "1", "--", c.url, "main") {
				return problem(fail.Network, fmt.Sprintf("Cannot fetch %s: nothing synced", c.url), "")
			}
			if !c.gitQuiet("-C", c.Repo, "reset", "--quiet", "--hard", "FETCH_HEAD") || !c.gitQuiet("-C", c.Repo, "clean", "--quiet", "-ffdx") {
				return problem(fail.Network, fmt.Sprintf("Cannot update the clone in %s: nothing synced", c.Repo), "")
			}
			return nil
		}
		c.note(fmt.Sprintf("The clone in %s has the origin '%s', not %s - cloning again", c.Repo, origin, c.url))
		again = true
	default:
		return problem(fail.Usage, c.Repo+" is not a git clone and not empty", "move it, or set BOOTSTRAP_CACHE to another folder")
	}
	if !again {
		return nil
	}
	c.note(fmt.Sprintf("Cloning into %s", c.Repo))
	_ = os.RemoveAll(c.Repo)
	_ = os.MkdirAll(filepath.Dir(c.Repo), 0o777)
	if !c.gitQuiet("clone", "--quiet", "--depth", "1", "--single-branch", "--branch", "main", "--", c.url, c.Repo) {
		_ = os.RemoveAll(c.Repo)
		return problem(fail.Network, fmt.Sprintf("Cannot clone %s: nothing synced", c.url), "")
	}
	return nil
}

func isEmptyDir(p string) bool {
	e, err := os.ReadDir(p)
	return err == nil && len(e) == 0
}

var semver = regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)$`)

// CheckVersion compares the version <version> of this binary with
// cli/VERSION of the clone <dir>, the minimum version. A `dev` build
// is new enough. An older binary is code 2.
func CheckVersion(version, dir string) error {
	b, _ := os.ReadFile(dir + "/" + source.CLIDir + "/VERSION")
	need, _, _ := strings.Cut(string(b), "\n")
	need = strings.TrimSpace(need)
	nm := semver.FindStringSubmatch(need)
	if nm == nil {
		return problem(fail.Network, fmt.Sprintf("%s/VERSION of the clone is not MAJOR.MINOR.PATCH ('%s'): nothing synced", source.CLIDir, need), "")
	}
	hm := semver.FindStringSubmatch(strings.TrimPrefix(version, "v"))
	if hm == nil {
		return nil
	}
	for i := 1; i <= 3; i++ {
		h, _ := strconv.Atoi(hm[i])
		n, _ := strconv.Atoi(nm[i])
		if h > n {
			return nil
		}
		if h < n {
			return problem(fail.Usage, fmt.Sprintf("bootstrap %s is older than %s, the version that the clone needs", version, need), "upgrade it with: mise use -g github:virajp/bootstrap")
		}
	}
	return nil
}
