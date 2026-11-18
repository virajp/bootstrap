package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// cloneCase is the state of one clone test: a remote (a clone of the
// bootstrap copy), a scratch consumer with values, and the folder of the
// test.
type cloneCase struct {
	t                     *testing.T
	dir, remote, consumer string
}

func newCloneCase(t *testing.T) *cloneCase {
	t.Helper()
	dir := harness.Dir(t)
	c := &cloneCase{t: t, dir: dir, remote: filepath.Join(dir, "remote"), consumer: filepath.Join(dir, "clone-consumer")}
	harness.Git(t, dir, "clone", "--quiet", "--", rc.BaseClone, c.remote)
	copyConsumer(t, c.consumer)
	rc.WriteValues(t, filepath.Join(c.consumer, ".config", "bootstrap"), `environments = [ "dev" ]`)
	harness.Git(t, c.consumer, "add", "-A")
	harness.Git(t, c.consumer, "commit", "--quiet", "--no-verify", "-m", "values")
	return c
}

// cache gives a new cache folder <name>/bootstrap (new_cache).
func (c *cloneCase) cache(name string) string {
	cache := filepath.Join(c.dir, name, "bootstrap")
	if err := os.MkdirAll(filepath.Dir(cache), 0o755); err != nil {
		c.t.Fatal(err)
	}
	return cache
}

// env is the environment of run_bin: no BOOTSTRAP_CLONE and no
// BOOTSTRAP_VALUES, the cache <cache> and the URL <url>.
func (c *cloneCase) env(cache, url string) []string {
	return harness.SetEnv(harness.Env(), "BOOTSTRAP_CLONE", "BOOTSTRAP_VALUES",
		"BOOTSTRAP_CACHE="+cache, "BOOTSTRAP_URL="+url, "MISE_PROJECT_ROOT="+c.consumer)
}

// run is run_bin: `bootstrap sync --dry-run` in the consumer, with the
// engine <bin> (empty for the engine of the test run).
func (c *cloneCase) run(cache, url, bin string) (int, string) {
	if bin == "" {
		bin = harness.Bin
	}
	cmd := &harness.Cmd{Dir: c.consumer, Env: c.env(cache, url), Args: []string{bin, "sync", "--dry-run"}}
	return cmd.Run()
}

// head gives the commit of HEAD of <repo>.
func (c *cloneCase) head(repo string) string {
	return harness.Git(c.t, repo, "rev-parse", "HEAD")
}

// checkClone is check_clone: the exit code, the commit of the copy (when
// <commit> is not empty), and no lock left.
func checkClone(t *testing.T, name string, want, got int, commit, log, cache string) {
	t.Helper()
	switch {
	case got != want:
		harness.Fail(t, fmt.Sprintf("%s: exit %d, expected %d", name, got, want), log)
	case commit != "" && !strings.Contains(log, "using bootstrap at "+commit[:7]):
		harness.Fail(t, fmt.Sprintf("%s: the sync did not use %s", name, commit[:7]), log)
	case !rc.NoFile(filepath.Join(cache, "lock")):
		harness.Fail(t, name+": the lock is still there", log)
	default:
		harness.Pass(t, fmt.Sprintf("%s: exit %d", name, got))
	}
}

// checkPaths is check_paths: each path of <present> is there, and each path
// of <absent> is not.
func checkPaths(t *testing.T, name, log string, present, absent []string) {
	t.Helper()
	bad := ""
	for _, p := range present {
		if _, err := os.Stat(p); err != nil {
			bad += " missing:" + p
		}
	}
	for _, p := range absent {
		if _, err := os.Stat(p); err == nil {
			bad += " present:" + p
		}
	}
	if bad != "" {
		harness.Fail(t, name+":"+bad, log)
	} else {
		harness.Pass(t, name+": the folders")
	}
}

var (
	releaseOnce    sync.Once
	releaseBin     string
	releaseVersion string
	releaseErr     error
)

// release gives a binary with the release version of cli/VERSION, for the
// version check, and that version.
func release(t *testing.T) (string, string) {
	t.Helper()
	releaseOnce.Do(func() {
		b, err := os.ReadFile(filepath.Join(rc.BaseClone, "cli", "VERSION"))
		if err != nil {
			releaseErr = err
			return
		}
		releaseVersion = strings.TrimSpace(string(b))
		releaseBin = filepath.Join(harness.Tmp, "engine-"+releaseVersion)
		cmd := exec.Command("go", "build", "-C", filepath.Join(harness.Root, "cli"), "-buildvcs=false",
			"-ldflags", "-X main.version="+releaseVersion, "-o", releaseBin, "./cmd/bootstrap")
		cmd.Env = harness.Env()
		if out, err := cmd.CombinedOutput(); err != nil {
			releaseErr = fmt.Errorf("go build of the release engine: %v\n%s", err, out)
		}
	})
	if releaseErr != nil {
		t.Fatal(releaseErr)
	}
	return releaseBin, releaseVersion
}

// versionRemote gives a remote with <version> in cli/VERSION.
func (c *cloneCase) versionRemote(version string) string {
	r := filepath.Join(c.dir, "remote-version")
	harness.Git(c.t, c.dir, "clone", "--quiet", "--", rc.BaseClone, r)
	harness.WriteFile(c.t, filepath.Join(r, "cli", "VERSION"), version+"\n")
	harness.Git(c.t, r, "commit", "--quiet", "--no-verify", "-am", "a VERSION")
	return r
}

// TestClone: the clone of bootstrap. Without BOOTSTRAP_CLONE, bootstrap
// takes the lock of its cache, fetches main of BOOTSTRAP_URL into the cache
// and syncs from a copy. Each run is `bootstrap sync --dry-run` in a scratch
// consumer. The log names the commit of the copy (BOOTSTRAP_VERBOSE=1).
func TestClone(t *testing.T) {
	// One cache through five steps, in order.
	t.Run("the cache steps", func(t *testing.T) {
		t.Parallel()
		c := newCloneCase(t)
		cache := c.cache("clone")
		url := "file://" + c.remote

		name := "bootstrap clones into the cache"
		got, log := c.run(cache, url, "")
		checkClone(t, name, 0, got, c.head(c.remote), log, cache)
		checkLog(t, name, log, true, "Cloning into")

		harness.Git(t, c.remote, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "second")
		name = "bootstrap fetches a new commit into the cache"
		got, log = c.run(cache, url, "")
		checkClone(t, name, 0, got, c.head(c.remote), log, cache)
		checkLog(t, name, log, true, " · ")

		liveLock(t, c, cache, url)

		remote2 := filepath.Join(c.dir, "remote2")
		harness.Git(t, c.dir, "clone", "--quiet", "--", rc.BaseClone, remote2)
		harness.Git(t, remote2, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "other")
		name = "bootstrap clones again for another URL"
		got, log = c.run(cache, "file://"+remote2, "")
		checkClone(t, name, 0, got, c.head(remote2), log, cache)
		checkLog(t, name, log, true, "cloning again")

		if err := os.Rename(remote2, remote2+".gone"); err != nil {
			t.Fatal(err)
		}
		name = "bootstrap, a fetch that fails"
		got, log = c.run(cache, "file://"+remote2, "")
		checkClone(t, name, 5, got, "", log, cache)
		checkPaths(t, name, log, []string{filepath.Join(cache, "repo", ".git")}, nil)
	})

	t.Run("a normalized origin", func(t *testing.T) {
		t.Parallel()
		c := newCloneCase(t)
		name := "bootstrap, a normalized origin is the same clone"
		cache := c.cache("normalized")
		got, log := c.run(cache, "file://"+c.remote, "")
		checkClone(t, name+", the first run", 0, got, "", log, cache)
		got, log = c.run(cache, c.remote, "")
		checkClone(t, name+", the second run", 0, got, "", log, cache)
		checkLog(t, name, log, true, " · ")
		checkLog(t, name, log, false, "Cloning into")
	})

	t.Run("a non-git folder", func(t *testing.T) {
		t.Parallel()
		c := newCloneCase(t)
		name := "bootstrap, a non-git folder stops with 2"
		cache := c.cache("not-a-clone")
		harness.WriteFile(t, filepath.Join(cache, "repo", "file"), "")
		got, log := c.run(cache, "file://"+c.remote, "")
		checkClone(t, name, 2, got, "", log, cache)
		checkPaths(t, name, log, []string{filepath.Join(cache, "repo", "file")}, nil)
	})

	t.Run("an old clone at the cache root", func(t *testing.T) {
		t.Parallel()
		c := newCloneCase(t)
		name := "bootstrap, an old clone at the cache root"
		cache := c.cache("old-layout")
		harness.Git(t, c.dir, "init", "--quiet", cache)
		harness.Git(t, cache, "remote", "add", "origin", "file://"+c.remote)
		got, log := c.run(cache, "file://"+c.remote, "")
		checkClone(t, name, 0, got, "", log, cache)
		checkPaths(t, name, log, []string{filepath.Join(cache, "repo", ".git")}, []string{filepath.Join(cache, ".git")})
	})

	t.Run("another origin at the cache root", func(t *testing.T) {
		t.Parallel()
		c := newCloneCase(t)
		name := "bootstrap, another origin at the cache root stops with 2"
		cache := c.cache("other-root-origin")
		harness.Git(t, c.dir, "init", "--quiet", cache)
		harness.Git(t, cache, "remote", "add", "origin", "file://"+filepath.Join(c.dir, "elsewhere"))
		harness.WriteFile(t, filepath.Join(cache, "file"), "")
		got, log := c.run(cache, "file://"+c.remote, "")
		checkClone(t, name, 2, got, "", log, cache)
		checkPaths(t, name, log, []string{filepath.Join(cache, ".git"), filepath.Join(cache, "file")},
			[]string{filepath.Join(cache, "repo")})
		checkLog(t, name, log, true, "is a git repository with the origin")
	})

	// The version check: cli/VERSION of the clone is the minimum version of
	// the binary. A dev build is new enough.
	t.Run("a binary older than VERSION", func(t *testing.T) {
		t.Parallel()
		bin, version := release(t)
		c := newCloneCase(t)
		r := c.versionRemote("99.0.0")
		name := "bootstrap, a binary older than VERSION stops with 2"
		cache := c.cache("version-old")
		got, log := c.run(cache, "file://"+r, bin)
		checkClone(t, name, 2, got, "", log, cache)
		checkLog(t, name, log, true, "bootstrap "+version+" is older than 99.0.0, the version that the clone needs")
		checkLog(t, name, log, true, "Upgrade it with: mise use -g github:virajp/bootstrap")
		checkLog(t, name, log, false, "using bootstrap at")
	})

	t.Run("a dev build", func(t *testing.T) {
		t.Parallel()
		c := newCloneCase(t)
		r := c.versionRemote("99.0.0")
		name := "bootstrap, a dev build is new enough"
		cache := c.cache("version-dev")
		got, log := c.run(cache, "file://"+r, "")
		checkClone(t, name, 0, got, "", log, cache)
	})

	t.Run("a binary at VERSION", func(t *testing.T) {
		t.Parallel()
		bin, _ := release(t)
		c := newCloneCase(t)
		name := "bootstrap, a binary at VERSION"
		cache := c.cache("version-same")
		got, log := c.run(cache, "file://"+c.remote, bin)
		checkClone(t, name, 0, got, "", log, cache)
	})

	t.Run("a bad VERSION", func(t *testing.T) {
		t.Parallel()
		bin, _ := release(t)
		c := newCloneCase(t)
		r := c.versionRemote("v1.2")
		name := "bootstrap, a VERSION that is not MAJOR.MINOR.PATCH stops with 5"
		cache := c.cache("version-bad")
		got, log := c.run(cache, "file://"+r, bin)
		checkClone(t, name, 5, got, "", log, cache)
		checkLog(t, name, log, true, "cli/VERSION of the clone is not MAJOR.MINOR.PATCH ('v1.2'): nothing synced")
	})
}

// liveLock: a live holder makes the run wait. When the holder dies, the run
// takes its stale lock.
func liveLock(t *testing.T, c *cloneCase, cache, url string) {
	t.Helper()
	holder := exec.Command("sleep", "30")
	if err := holder.Start(); err != nil {
		t.Fatal(err)
	}
	holderDone := make(chan struct{})
	go func() { _ = holder.Wait(); close(holderDone) }()
	defer func() {
		_ = holder.Process.Kill()
		<-holderDone
	}()
	harness.WriteFile(t, filepath.Join(cache, "lock", "pid"), fmt.Sprintf("%d\n", holder.Process.Pid))

	logPath := filepath.Join(c.dir, "lock-log")
	out, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	runner := exec.Command(harness.Bin, "sync", "--dry-run")
	runner.Dir = c.consumer
	runner.Env = c.env(cache, url)
	runner.Stdout, runner.Stderr = out, out
	if err := runner.Start(); err != nil {
		t.Fatal(err)
	}
	runnerDone := make(chan error, 1)
	go func() { runnerDone <- runner.Wait() }()

	time.Sleep(3 * time.Second)
	waitMsg := fmt.Sprintf("Waiting for another sync (pid %d) ...", holder.Process.Pid)
	var runErr error
	finished := false
	select {
	case runErr = <-runnerDone:
		finished = true
		harness.Fail(t, "bootstrap with a live lock: the run did not wait", harness.ReadFile(logPath))
	default:
		if !strings.Contains(harness.ReadFile(logPath), waitMsg) {
			harness.Fail(t, "bootstrap with a live lock: no wait message", harness.ReadFile(logPath))
		} else {
			harness.Pass(t, "bootstrap with a live lock: the run waits")
		}
	}
	_ = holder.Process.Kill()
	<-holderDone
	if !finished {
		select {
		case runErr = <-runnerDone:
		case <-time.After(10 * time.Second):
			_ = runner.Process.Kill()
			runErr = <-runnerDone
		}
	}
	_ = out.Close()
	got := 0
	if runErr != nil {
		got = -1
		if ee, ok := runErr.(*exec.ExitError); ok {
			got = ee.ExitCode()
		}
	}
	checkClone(t, "bootstrap takes the lock of a dead holder", 0, got, c.head(c.remote), harness.ReadFile(logPath), cache)
}
