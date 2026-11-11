package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// defaultURL is the clone URL. BOOTSTRAP_URL changes it.
const defaultURL = "https://github.com/virajp/bootstrap.git"

// clone is the bootstrap clone of one command: BOOTSTRAP_CLONE, else a
// temporary copy of the cache clone.
type clone struct {
	a     *app
	dir   string // The clone that the command reads.
	cache string
	repo  string
	lock  string
	url   string
	held  bool
	copy  string // The temporary copy, removed by cleanup.
}

// acquireClone gives the clone of the command. Without BOOTSTRAP_CLONE, it
// takes the lock of the cache, fetches main into the cache clone and copies
// it. The lock stays held: the caller releases it with unlock.
func (a *app) acquireClone() (*clone, fail.Code) {
	c := &clone{a: a}
	if d := a.getenv("BOOTSTRAP_CLONE"); d != "" {
		c.dir = d
		return c, fail.OK
	}
	c.cache = a.getenv("BOOTSTRAP_CACHE")
	if c.cache == "" {
		xdg := a.getenv("XDG_CACHE_HOME")
		if xdg == "" {
			xdg = a.getenv("HOME") + "/.cache"
		}
		c.cache = xdg + "/bootstrap"
	}
	c.repo = c.cache + "/repo"
	c.lock = c.cache + "/lock"
	c.url = a.getenv("BOOTSTRAP_URL")
	if c.url == "" {
		c.url = defaultURL
	}
	if code := c.takeLock(); code != fail.OK {
		return c, code
	}
	if code := c.fetch(); code != fail.OK {
		c.cleanup()
		return c, code
	}
	tmp := a.getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	copyDir, err := os.MkdirTemp(tmp, "bootstrap-clone.")
	if err != nil {
		c.cleanup()
		a.errorf("cannot create a temp folder")
		return c, fail.Other
	}
	c.copy = copyDir
	if out, err := exec.Command("cp", "-R", "--", c.repo+"/.", copyDir).CombinedOutput(); err != nil {
		a.stderr.Write(out)
		c.cleanup()
		a.errorf("cannot copy the clone %s", c.repo)
		return c, fail.Other
	}
	c.dir = copyDir
	return c, fail.OK
}

// takeLock waits until this run holds the lock folder of the cache. A lock
// whose holder is dead, or a lock without a pid that is older than one
// minute, is removed.
func (c *clone) takeLock() fail.Code {
	if err := os.MkdirAll(c.cache, 0o777); err != nil {
		c.a.errorf("cannot create the cache %s", c.cache)
		return fail.Other
	}
	waited := false
	for os.Mkdir(c.lock, 0o777) != nil {
		if staleLock(c.lock) {
			continue
		}
		pid := readPid(c.lock)
		if pid != "" && !waited {
			c.a.infof("Waiting for another sync (pid %s) ...", pid)
			waited = true
		}
		time.Sleep(time.Second)
	}
	c.held = true
	_ = os.WriteFile(c.lock+"/pid", []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644)
	return fail.OK
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
			os.RemoveAll(take)
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
		os.RemoveAll(lock)
	}
	os.Remove(take)
	return stale
}

// unlock releases the lock when this run holds it.
func (c *clone) unlock() {
	if c.held {
		os.RemoveAll(c.lock)
		c.held = false
	}
}

// cleanup releases the lock and removes the copy.
func (c *clone) cleanup() {
	c.unlock()
	if c.copy != "" {
		os.RemoveAll(c.copy)
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

// gitEnv is the environment of a git call: without the variables of a git
// hook, and with no prompt for a network call.
func gitEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if k == "GIT_DIR" || k == "GIT_WORK_TREE" || k == "GIT_INDEX_FILE" {
			continue
		}
		env = append(env, kv)
	}
	return append(env, "GIT_TERMINAL_PROMPT=0", "GIT_SSH_COMMAND=ssh -o BatchMode=yes")
}

// git runs git with <args> and gives its stdout without the last newline.
func git(args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Env = gitEnv()
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\n"), err == nil
}

// gitQuiet runs git with <args>, with its stderr on the stderr of the run.
func (c *clone) gitQuiet(args ...string) bool {
	cmd := exec.Command("git", args...)
	cmd.Env = gitEnv()
	cmd.Stderr = c.a.stderr
	return cmd.Run() == nil
}

// fetch brings the cache clone to the newest commit of main: a clone when it
// is missing or empty, a fetch when its origin is the URL, a new clone for
// another origin. A folder that is not a clone is never deleted (code 2). A
// failed git call gives code 5.
func (c *clone) fetch() fail.Code {
	// A clone of the URL at the cache root is the old layout. Any other
	// repository there stays.
	if exists(c.cache + "/.git") {
		origin, _ := git("config", "--file", c.cache+"/.git/config", "--get", "remote.origin.url")
		if origin == "" || normalizeURL(origin) != normalizeURL(c.url) {
			c.a.fail(fmt.Sprintf("%s is a git repository with the origin '%s', not an old clone of %s", c.cache, origin, c.url), "move it, or set BOOTSTRAP_CACHE to another folder")
			return fail.Usage
		}
		entries, _ := os.ReadDir(c.cache)
		for _, e := range entries {
			switch e.Name() {
			case "repo", "lock", "lock.takeover", "macos-generated.tsv":
				continue
			}
			os.RemoveAll(filepath.Join(c.cache, e.Name()))
		}
		c.a.infof("Removed an old clone at %s", c.cache)
	}

	again := false
	st, err := os.Lstat(c.repo)
	switch {
	case err != nil:
		again = true
	case st.IsDir() && isEmptyDir(c.repo):
		again = true
	case st.IsDir() && exists(c.repo+"/.git"):
		origin, _ := git("-C", c.repo, "config", "--get", "remote.origin.url")
		if origin != "" && normalizeURL(origin) == normalizeURL(c.url) {
			// From the URL, not from origin: the compare drops the scheme,
			// and the transport must not come from the cached origin.
			if !c.gitQuiet("-C", c.repo, "fetch", "--quiet", "--depth", "1", "--", c.url, "main") {
				c.a.errorf("Cannot fetch %s: nothing synced", c.url)
				return fail.Network
			}
			if !c.gitQuiet("-C", c.repo, "reset", "--quiet", "--hard", "FETCH_HEAD") || !c.gitQuiet("-C", c.repo, "clean", "--quiet", "-ffdx") {
				c.a.errorf("Cannot update the clone in %s: nothing synced", c.repo)
				return fail.Network
			}
			return fail.OK
		}
		c.a.infof("The clone in %s has the origin '%s', not %s - cloning again", c.repo, origin, c.url)
		again = true
	default:
		c.a.fail(c.repo+" is not a git clone and not empty", "move it, or set BOOTSTRAP_CACHE to another folder")
		return fail.Usage
	}
	if !again {
		return fail.OK
	}
	c.a.infof("Cloning into %s", c.repo)
	os.RemoveAll(c.repo)
	_ = os.MkdirAll(filepath.Dir(c.repo), 0o777)
	if !c.gitQuiet("clone", "--quiet", "--depth", "1", "--single-branch", "--branch", "main", "--", c.url, c.repo) {
		os.RemoveAll(c.repo)
		c.a.errorf("Cannot clone %s: nothing synced", c.url)
		return fail.Network
	}
	return fail.OK
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func isEmptyDir(p string) bool {
	e, err := os.ReadDir(p)
	return err == nil && len(e) == 0
}

var semver = regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)$`)

// checkVersion compares the version of this binary with
// cli/VERSION of the clone <dir>, the minimum version. A `dev` build
// is new enough. An older binary stops with code 2.
func (a *app) checkVersion(dir string) fail.Code {
	b, _ := os.ReadFile(dir + "/" + source.CLIDir + "/VERSION")
	need, _, _ := strings.Cut(string(b), "\n")
	need = strings.TrimSpace(need)
	nm := semver.FindStringSubmatch(need)
	if nm == nil {
		a.errorf("%s/VERSION of the clone is not MAJOR.MINOR.PATCH ('%s'): nothing synced", source.CLIDir, need)
		return fail.Network
	}
	hm := semver.FindStringSubmatch(strings.TrimPrefix(a.version, "v"))
	if hm == nil {
		return fail.OK
	}
	for i := 1; i <= 3; i++ {
		h, _ := strconv.Atoi(hm[i])
		n, _ := strconv.Atoi(nm[i])
		if h > n {
			return fail.OK
		}
		if h < n {
			a.fail(fmt.Sprintf("bootstrap %s is older than %s, the version that the clone needs", a.version, need), "upgrade it with: mise use -g github:virajp/bootstrap")
			return fail.Usage
		}
	}
	return fail.OK
}
