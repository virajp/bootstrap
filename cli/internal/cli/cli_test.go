package cli

import (
	"bytes"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/selection"
)

func TestRenderTree(t *testing.T) {
	deps := map[string][]string{
		"gitleaks":   {"_sec"},
		"grype":      {"_sec"},
		"linter":     {"node"},
		"pre-commit": {"python"},
		"python":     {"uv"},
		"uv":         {"python"}, // a loop of depends
	}
	required := func(tool string) (cell, string) {
		if tool == "git" || tool == "pre-commit" {
			return cell{vis: "* required", colored: "* required"}, ""
		}
		return cell{}, ""
	}
	roots := []listRoot{
		{"base (pack)", []string{"pre-commit", "git", "gitleaks", "grype", "linter"}},
		{"extra (pack)", []string{"zed"}},
		{"tools (selected directly)", []string{"git", "claude", "gitleaks", "pre-commit"}},
	}
	want := `base (pack)
├── git           * required
├── gitleaks
│   └── _sec
├── grype
│   └── _sec
├── linter
│   └── node
└── pre-commit    * required
    └── python
        └── uv
extra (pack)
└── zed
tools (selected directly)
├── claude
├── git           * required
├── gitleaks ^
└── pre-commit ^  * required
`
	opt := treeOpts{cell: required, arrowCell: true}
	got, arrows := renderTree(roots, deps, opt)
	if got != want || !arrows {
		t.Errorf("got (arrows %v):\n%s\nwant:\n%s", arrows, got, want)
	}
	if got, arrows := renderTree(roots[:2], deps, opt); strings.Contains(got, "tools (selected") || arrows {
		t.Errorf("a tools root without tools: %s", got)
	}
	// A tree without a repeated subtree has no arrow.
	if _, arrows := renderTree(roots[1:2], deps, opt); arrows {
		t.Error("an arrow without a repeated subtree")
	}
	// include keeps a tool and the path to it.
	only := treeOpts{include: func(tool string) bool { return tool == "node" }}
	if got, _ := renderTree(roots[:1], deps, only); got != "base (pack)\n└── linter\n    └── node\n" {
		t.Errorf("include: %s", got)
	}
}

// TestListRoots: the roots of list are the tools, with no pack root. A
// tool that a selected tool needs is not a root; a required member of base
// without a table is one.
func TestListRoots(t *testing.T) {
	s := &core.State{Res: &selection.Resolved{
		Sel: []selection.Entry{
			{Tool: "mise", Envs: "-", Src: "default"},
			{Tool: "git", Envs: "-", Src: "default"},
			{Tool: "python", Envs: "-", Src: "tools.python"},
			{Tool: "uv", Envs: "-", Src: "tools.uv"},
		},
		Deps: []selection.DepRow{{Tool: "python", Deps: []string{"uv"}}},
	}}
	roots := listRoots(s)
	if len(roots) != 1 || roots[0].name != "" || strings.Join(roots[0].tools, " ") != "mise git python" {
		t.Fatalf("roots = %v", roots)
	}
	required := func(tool string) (cell, string) {
		if tool == "git" || tool == "mise" {
			return cell{vis: "* required", colored: "* required"}, ""
		}
		return cell{}, ""
	}
	got, _ := renderTree(roots, depsOf(s), treeOpts{cell: required})
	want := `git     * required
mise    * required
python
└── uv
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
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
		want(lines, "url", core.DefaultURL, "default")
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
		{"mac config", []string{"mac", "config"}, "", "mac config is not a command"},
		{"empty url", []string{"config", "--url="}, "", "--url needs a value"},
		{"verbose of config", []string{"config", "--verbose"}, "", "config has no flag --verbose"},
		{"repo-path of mac init", []string{"mac", "init", "--repo-path", "/r"}, "", "mac init has no flag --repo-path"},
		{"config of version", []string{"version", "--config", "/c"}, "", "version has no flag --config"},
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
			if strings.HasPrefix(c.msg, ": ") && !strings.Contains(stderr.String(), "[ERROR] "+path+c.msg) {
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

func TestShortValue(t *testing.T) {
	long := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICGo"
	for in, want := range map[string]string{
		"short":                 "short",
		strings.Repeat("a", 40): strings.Repeat("a", 40),
		long + " comment":       "ssh-ed25519 AAAAC3Nz…ment",
		strings.Repeat("é", 20) + strings.Repeat("x", 30) + "abcd": strings.Repeat("é", 20) + "…abcd",
	} {
		if got := shortValue(in); got != want {
			t.Errorf("shortValue(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUsageLine(t *testing.T) {
	for _, c := range []struct {
		mac        bool
		cmd, flags string
		want       string
	}{
		{false, "sync", "dry-run jobs verbose", "bootstrap sync [--dry-run] [--jobs <n>] [--verbose]"},
		{true, "add", "dry-run yes answer jobs verbose no-bootstrap", "bootstrap mac add <tool...> [--dry-run] [--yes] [--answer <tool>.<key>=<value>] [--jobs <n>] [--verbose] [--no-bootstrap]"},
		{false, "list", "", "bootstrap list"},
	} {
		if got := usageLine(c.mac, c.cmd, c.flags); got != c.want {
			t.Errorf("usageLine = %q, want %q", got, c.want)
		}
	}
}

func TestDetectText(t *testing.T) {
	for _, c := range []struct {
		d    core.Detected
		want string
	}{
		{core.Detected{Evidence: []string{"a", "b", "c"}, Message: "m"}, "a, +2 more: m"},
		{core.Detected{Evidence: []string{"a"}, Message: "m"}, "a: m"},
		{core.Detected{Evidence: []string{"a"}}, "a"},
		{core.Detected{Message: "m"}, "m"},
	} {
		if got := detectText(c.d); got != c.want {
			t.Errorf("detectText(%+v) = %q, want %q", c.d, got, c.want)
		}
	}
}
