package cli

import (
	"bytes"
	"os"
	"runtime"
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

// setupEnv sets the variables of the settings with t.Setenv, so that Go
// resets them after the test, and points BOOTSTRAP_CONFIG at <config>.
func setupEnv(t *testing.T, config string) {
	t.Helper()
	for _, v := range []string{"BOOTSTRAP_JOBS", "BOOTSTRAP_CACHE", "BOOTSTRAP_URL", "BOOTSTRAP_CLONE", "MACOS_SETUP_REPO_PATH", "BOOTSTRAP_VERBOSE"} {
		t.Setenv(v, "")
	}
	t.Setenv("BOOTSTRAP_CONFIG", config)
	t.Setenv("HOME", "/h")
	t.Setenv("XDG_CACHE_HOME", "/xc")
	t.Setenv("NO_COLOR", "1")
}

// writeConfig writes a config file with <body> and gives its path.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := t.TempDir() + "/config.toml"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// configLines runs `bootstrap <args>` and gives the fields of each line of
// stdout, by the first field.
func configLines(t *testing.T, args ...string) map[string][]string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := Run("dev", args, strings.NewReader(""), &stdout, &stderr); code != fail.OK {
		t.Fatalf("%v: code %d, stderr %q", args, code, stderr.String())
	}
	lines := map[string][]string{}
	for _, l := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		f := strings.Fields(l)
		lines[f[0]] = f[1:]
	}
	return lines
}

func TestConfigSources(t *testing.T) {
	file := writeConfig(t, "jobs = 3\ncache_dir = \"~/x\"\n")
	want := func(lines map[string][]string, key, value, source string) {
		t.Helper()
		got := strings.Join(lines[key], " ")
		if got != value+" "+source {
			t.Errorf("%s: %q, want %q", key, got, value+" "+source)
		}
	}
	t.Run("default", func(t *testing.T) {
		setupEnv(t, "/nonexistent/bootstrap/config.toml")
		lines := configLines(t, "config")
		want(lines, "file", "/nonexistent/bootstrap/config.toml (missing)", "variable BOOTSTRAP_CONFIG")
		want(lines, "jobs", strconv.Itoa(2*runtime.NumCPU()), "default")
		want(lines, "cache_dir", "/xc/bootstrap", "default")
		want(lines, "url", defaultURL, "default")
		want(lines, "clone", "-", "default")
		want(lines, "mac.repo_path", "-", "default")
	})
	t.Run("file", func(t *testing.T) {
		setupEnv(t, file)
		lines := configLines(t, "config")
		want(lines, "file", file, "variable BOOTSTRAP_CONFIG")
		want(lines, "jobs", "3", "file")
		want(lines, "cache_dir", "/h/x", "file")
		if os.Getenv("BOOTSTRAP_JOBS") != "3" || os.Getenv("BOOTSTRAP_CACHE") != "/h/x" {
			t.Error("the values of the file are not variables of the process")
		}
	})
	t.Run("variable", func(t *testing.T) {
		setupEnv(t, file)
		t.Setenv("BOOTSTRAP_JOBS", "abc")
		t.Setenv("BOOTSTRAP_CACHE", "/c")
		t.Setenv("MACOS_SETUP_REPO_PATH", "/r")
		lines := configLines(t, "config")
		want(lines, "jobs", "abc", "variable BOOTSTRAP_JOBS")
		want(lines, "cache_dir", "/c", "variable BOOTSTRAP_CACHE")
		want(lines, "mac.repo_path", "/r", "variable MACOS_SETUP_REPO_PATH")
	})
	t.Run("flag", func(t *testing.T) {
		setupEnv(t, "/nonexistent/bootstrap/config.toml")
		t.Setenv("BOOTSTRAP_JOBS", "4")
		t.Setenv("BOOTSTRAP_CACHE", "/c")
		lines := configLines(t, "config", "--config", file, "--jobs", "5", "--cache-dir=/f",
			"--url", "/u", "--clone", "/k", "--repo-path", "/r")
		want(lines, "file", file, "flag --config")
		want(lines, "jobs", "5", "flag --jobs")
		want(lines, "cache_dir", "/f", "flag --cache-dir")
		want(lines, "url", "/u", "flag --url")
		want(lines, "clone", "/k", "flag --clone")
		want(lines, "mac.repo_path", "/r", "flag --repo-path")
		if os.Getenv("BOOTSTRAP_CONFIG") != file || os.Getenv("BOOTSTRAP_JOBS") != "5" || os.Getenv("MACOS_SETUP_REPO_PATH") != "/r" {
			t.Error("a flag does not set its variable")
		}
	})
}

func TestConfigErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
		body string
		msg  string
	}{
		{"unknown key", []string{"config"}, "verbose = true\n", ": verbose: unknown key"},
		{"unknown key before the preflight", []string{"sync"}, "[mac]\nverbose = 1\n", ": mac.verbose: unknown key"},
		{"mac config", []string{"mac", "config"}, "", "unknown command mac config"},
		{"empty url", []string{"config", "--url="}, "", "--url needs a value"},
		{"verbose of config", []string{"config", "--verbose"}, "", "unknown flag --verbose"},
		{"repo-path of mac init", []string{"mac", "init", "--repo-path", "/r"}, "", "unknown flag --repo-path"},
		{"config of version", []string{"version", "--config", "/c"}, "", "unknown flag --config"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := writeConfig(t, c.body)
			setupEnv(t, path)
			var stdout, stderr bytes.Buffer
			if code := Run("dev", c.args, strings.NewReader(""), &stdout, &stderr); code != fail.Usage {
				t.Errorf("code %d, want %d", code, fail.Usage)
			}
			if !strings.Contains(stderr.String(), c.msg) {
				t.Errorf("stderr %q lacks %q", stderr.String(), c.msg)
			}
			if strings.HasPrefix(c.msg, ": ") && !strings.Contains(stderr.String(), "bootstrap: "+path+c.msg) {
				t.Errorf("stderr %q does not name the file", stderr.String())
			}
		})
	}
}

func TestVerboseFlag(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	// A folder that is not a git repository: check stops at the preflight.
	t.Setenv("PWD", t.TempDir())
	t.Setenv("MISE_PROJECT_ROOT", "")
	var stdout, stderr bytes.Buffer
	if code := Run("dev", []string{"check", "--verbose"}, strings.NewReader(""), &stdout, &stderr); code != fail.Usage {
		t.Errorf("code %d, stderr %q", code, stderr.String())
	}
	if os.Getenv("BOOTSTRAP_VERBOSE") != "1" {
		t.Error("--verbose does not set BOOTSTRAP_VERBOSE=1")
	}
}
