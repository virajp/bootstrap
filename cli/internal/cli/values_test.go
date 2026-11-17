package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// valuesRepo makes the fixture of the config tests: a clone with the tools
// a (selected, with defaults and a question) and b (not selected), and a
// repository with a values folder. It gives the folder and a function that
// runs bootstrap with the clone.
func valuesRepo(t *testing.T) (string, func(args ...string) (fail.Code, string, string)) {
	t.Helper()
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	clone, root := t.TempDir(), t.TempDir()
	const keys = catalogKeys + "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"
	writeFile(t, filepath.Join(clone, "cli", "VERSION"), "0.0.1\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "categories.toml"), catalogFile, 0o644)
	writeFile(t, filepath.Join(clone, "repo", "packs.toml"), "[pack.base]\na = { required = true }\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "a", "tools.toml"), keys+"\n[defaults]\nflag = true\ncount = 1\nlist = [ \"x\" ]\n\n[questions.name]\nquestion = \"N?\"\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "b", "tools.toml"), keys+"\n[defaults]\nflag = true\n", 0o644)
	dir := filepath.Join(root, ".config", "bootstrap")
	writeFile(t, filepath.Join(dir, "repo.toml"), "# Head.\nformat = 3\nrepo_name = \"x\"\ncore_branches = [ \"main\" ]\n\n[branches.main]\n# The landing.\nlanding = \"pr\"\n", 0o644)
	writeFile(t, filepath.Join(dir, "tools.toml"), "format = 3\n\n[tools.a]\nflag = false\n", 0o644)
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	t.Setenv("MISE_PROJECT_ROOT", root)
	t.Setenv("BOOTSTRAP_VALUES", "")
	t.Setenv("PWD", root)
	return dir, func(args ...string) (fail.Code, string, string) {
		var stdout, stderr bytes.Buffer
		code := Run("dev", append(args, "--clone", clone), strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
}

func TestConfigGet(t *testing.T) {
	_, run := valuesRepo(t)
	cases := []struct {
		args []string
		out  string
	}{
		{[]string{"repo.repo_name"}, `{"key":"repo.repo_name","value":"x","source":"file"}`},
		{[]string{"repo.branches.main.landing"}, `{"key":"repo.branches.main.landing","value":"pr","source":"file"}`},
		{[]string{"repo.branches.develop.landing"}, `{"key":"repo.branches.develop.landing","value":"direct-merge","source":"default"}`},
		{[]string{`repo.branches."release.1".landing`}, `{"key":"repo.branches.\"release.1\".landing","value":"direct-merge","source":"default"}`},
		{[]string{"repo.branches"}, `{"key":"repo.branches","value":{"main":{"landing":"pr"}},"source":"file"}`},
		{[]string{"repo.core_branches"}, `{"key":"repo.core_branches","value":["main"],"source":"file"}`},
		{[]string{"repo.environments"}, `{"key":"repo.environments","value":["dev"],"source":"default"}`},
		{[]string{"repo.commits.style"}, `{"key":"repo.commits.style","value":"conventional","source":"default"}`},
		{[]string{"repo.product_name"}, `{"key":"repo.product_name","value":null,"source":"default"}`},
		{[]string{"releases.web"}, `{"key":"releases.web","value":"NONE","source":"default"}`},
		{[]string{"tools.a.flag"}, `{"key":"tools.a.flag","value":false,"source":"file","catalog":true}`},
		{[]string{"tools.a.count"}, `{"key":"tools.a.count","value":1,"source":"default","catalog":true}`},
		{[]string{"tools.nope.key"}, `{"key":"tools.nope.key","value":null,"source":"default","catalog":true}`},
		{[]string{"repo.repo_name", "--format", "toml"}, "key = \"repo.repo_name\"\nvalue = \"x\"\nsource = \"file\""},
		{[]string{"repo.repo_name", "--format", "yaml"}, "key: repo.repo_name\nvalue: x\nsource: file"},
		{[]string{"repo.product_name", "--format", "toml"}, "key = \"repo.product_name\"\nsource = \"default\""},
	}
	for _, c := range cases {
		code, out, errText := run(append([]string{"config", "get"}, c.args...)...)
		if code != fail.OK || strings.TrimSpace(out) != c.out {
			t.Errorf("%v: code %d, out %q, want %q (%s)", c.args, code, out, c.out, errText)
		}
	}
	for _, key := range []string{"repo.nope", "repo", "other.x", "repo..x", `repo."a`} {
		if code, _, errText := run("config", "get", key); code != fail.Usage || !strings.Contains(errText, "is not a key of the values") {
			t.Errorf("%s: code %d, %s", key, code, errText)
		}
	}
	// A file that does not parse is a read failure.
	dir := os.Getenv("MISE_PROJECT_ROOT") + "/.config/bootstrap"
	writeFile(t, dir+"/releases.toml", "web = \n", 0o644)
	if code, _, _ := run("config", "get", "releases.web"); code != fail.Other {
		t.Errorf("a bad file: code %d", code)
	}
}

// contribRepo makes the fixture of the contributed values: the tool a
// (base) with a list default, and the tool c (selected) that contributes to
// a list item and a key that a does not have. The values file sets the list
// of a when <list> is not empty.
func contribRepo(t *testing.T, list string) (string, string) {
	t.Helper()
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	clone, root := t.TempDir(), t.TempDir()
	const keys = catalogKeys + "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"
	writeFile(t, filepath.Join(clone, "cli", "VERSION"), "0.0.1\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "categories.toml"), catalogFile, 0o644)
	writeFile(t, filepath.Join(clone, "repo", "packs.toml"), "[pack.base]\na = { required = true }\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "a", "tools.toml"), keys+"\n[defaults]\nlist = [ \"x\" ]\nflag = true\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "c", "tools.toml"), keys+"\n[contributes.a]\nlist = [ \"y\" ]\n\n[contributes.a.extra]\nk = \"v\"\n", 0o644)
	dir := filepath.Join(root, ".config", "bootstrap")
	writeFile(t, filepath.Join(dir, "repo.toml"), "format = 3\nrepo_name = \"x\"\ncore_branches = [ \"main\" ]\n", 0o644)
	tools := "format = 3\n\n[tools.c]\n"
	if list != "" {
		tools += "\n[tools.a]\nlist = " + list + "\n"
	}
	writeFile(t, filepath.Join(dir, "tools.toml"), tools, 0o644)
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	t.Setenv("MISE_PROJECT_ROOT", root)
	t.Setenv("BOOTSTRAP_VALUES", "")
	t.Setenv("PWD", root)
	return clone, root
}

// getOut runs bootstrap config get <args> and gives the code and stdout.
func getOut(args ...string) (fail.Code, string, string) {
	var stdout, stderr bytes.Buffer
	code := Run("dev", append([]string{"config", "get"}, args...), strings.NewReader(""), &stdout, &stderr)
	return code, strings.TrimSpace(stdout.String()), stderr.String()
}

// TestConfigGetContributed: a tools.* key gives the merged value, with the
// contributions of the selected tools (D15, D20), from --clone or
// BOOTSTRAP_CLONE (D27).
func TestConfigGetContributed(t *testing.T) {
	clone, _ := contribRepo(t, "")
	cases := []struct{ key, out string }{
		{"tools.a.extra", `{"key":"tools.a.extra","value":{"k":"v"},"source":"contributed","contributed_by":["c"],"catalog":true}`},
		{"tools.a.extra.k", `{"key":"tools.a.extra.k","value":"v","source":"contributed","contributed_by":["c"],"catalog":true}`},
		{"tools.a.list", `{"key":"tools.a.list","value":["x","y"],"source":"contributed","contributed_by":["c"],"catalog":true}`},
		{"tools.a.flag", `{"key":"tools.a.flag","value":true,"source":"default","catalog":true}`},
	}
	for _, c := range cases {
		if code, out, errText := getOut(c.key, "--clone", clone); code != fail.OK || out != c.out {
			t.Errorf("--clone %s: code %d, out %s, want %s (%s)", c.key, code, out, c.out, errText)
		}
		t.Setenv("BOOTSTRAP_CLONE", clone)
		if code, out, errText := getOut(c.key); code != fail.OK || out != c.out {
			t.Errorf("BOOTSTRAP_CLONE %s: code %d, out %s, want %s (%s)", c.key, code, out, c.out, errText)
		}
		t.Setenv("BOOTSTRAP_CLONE", "")
	}
	// A value of the file with contributed entries: source file, and the
	// contributing tools.
	clone, _ = contribRepo(t, `[ "z" ]`)
	want := `{"key":"tools.a.list","value":["x","y","z"],"source":"file","contributed_by":["c"],"catalog":true}`
	if code, out, errText := getOut("tools.a.list", "--clone", clone); code != fail.OK || out != want {
		t.Errorf("file and contributed: code %d, out %s, want %s (%s)", code, out, want, errText)
	}
}

// TestConfigGetNoCatalog: with no clone and no cache clone, get gives the
// file value alone, with catalog false, and fetches nothing (D23, D28). A
// cache clone is read with no fetch.
func TestConfigGetNoCatalog(t *testing.T) {
	clone, _ := contribRepo(t, `[ "z" ]`)
	cache := t.TempDir()
	t.Setenv("BOOTSTRAP_CACHE", cache)
	// A fetch from this URL fails.
	url := "file://" + t.TempDir() + "/no-such-origin"
	t.Setenv("BOOTSTRAP_URL", url)
	want := `{"key":"tools.a.list","value":["z"],"source":"file","catalog":false}`
	if code, out, errText := getOut("tools.a.list"); code != fail.OK || out != want {
		t.Errorf("no catalog: code %d, out %s, want %s (%s)", code, out, want, errText)
	}
	if code, out, _ := getOut("tools.a.extra"); code != fail.OK || out != `{"key":"tools.a.extra","value":null,"source":"default","catalog":false}` {
		t.Errorf("no catalog, contributed key: code %d, out %s", code, out)
	}
	if entries, _ := os.ReadDir(cache); len(entries) != 0 {
		t.Errorf("no catalog: the cache has %d entries, want none (no fetch)", len(entries))
	}
	// A cache clone with the origin of the URL: read as it is, with no fetch.
	repo := filepath.Join(cache, "repo")
	if out, err := exec.Command("cp", "-R", clone+"/.", repo).CombinedOutput(); err != nil {
		t.Fatalf("cp: %s", out)
	}
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"remote", "add", "origin", url}} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	want = `{"key":"tools.a.list","value":["x","y","z"],"source":"file","contributed_by":["c"],"catalog":true}`
	if code, out, errText := getOut("tools.a.list"); code != fail.OK || out != want {
		t.Errorf("cache clone: code %d, out %s, want %s (%s)", code, out, want, errText)
	}
	if _, err := os.Stat(filepath.Join(cache, "fetched")); err == nil {
		t.Errorf("cache clone: get fetched")
	}
}

func TestConfigShow(t *testing.T) {
	_, run := valuesRepo(t)
	code, out, _ := run("config", "show", "--format", "json")
	if code != fail.OK || !strings.Contains(out, `"repo": {`) || !strings.Contains(out, `"tools": {`) || !strings.Contains(out, `"releases": {}`) {
		t.Errorf("show: %d\n%s", code, out)
	}
	code, out, _ = run("config", "show", "repo", "--format", "toml")
	if code != fail.OK || !strings.Contains(out, "repo_name = \"x\"") || strings.Contains(out, "[repo]") {
		t.Errorf("show repo: %d\n%s", code, out)
	}
	if code, out, _ := run("config", "show", "releases", "--format", "yaml"); code != fail.OK || strings.TrimSpace(out) != "{}" {
		t.Errorf("show releases: %d %q", code, out)
	}
	if code, _, _ := run("config", "show", "nope"); code != fail.Usage {
		t.Errorf("show nope: %d", code)
	}
}

func TestConfigSet(t *testing.T) {
	dir, run := valuesRepo(t)
	read := func(name string) string {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		return string(b)
	}
	// A repo key: comments stay, and the sync line shows (D9, D26).
	code, _, errText := run("config", "set", "repo.product_name", "P")
	if code != fail.OK || !strings.Contains(errText, "run bootstrap sync to apply it") {
		t.Errorf("set product_name: %d %s", code, errText)
	}
	if r := read("repo.toml"); !strings.Contains(r, `product_name = "P"`) || !strings.Contains(r, "# The landing.") || !strings.HasPrefix(r, "# Head.") {
		t.Errorf("repo.toml:\n%s", r)
	}
	// A key without a sync: no sync line (D90).
	if code, _, errText := run("config", "set", "repo.topology", "multi-repo"); code != fail.OK || strings.Contains(errText, "sync") {
		t.Errorf("set topology: %d %s", code, errText)
	}
	if code, _, _ := run("config", "set", "repo.core_branches", `["main","develop"]`); code != fail.OK {
		t.Errorf("set core_branches: %d", code)
	}
	if code, _, errText := run("config", "set", "repo.core_branches", "main"); code != fail.Usage || !strings.Contains(errText, "needs a JSON value") {
		t.Errorf("set a list as text: %d %s", code, errText)
	}
	// A result that breaks a rule writes nothing (D27).
	before := read("repo.toml")
	if code, _, errText := run("config", "set", "repo.branches.feature.landing", "pr"); code != fail.Usage || !strings.Contains(errText, "not in core_branches") {
		t.Errorf("set a branch that is not core: %d %s", code, errText)
	}
	if code, _, _ := run("config", "set", "repo.branches.main.landing", "fast"); code != fail.Usage {
		t.Errorf("set a value outside the enum: %d", code)
	}
	if read("repo.toml") != before {
		t.Error("a refused set wrote repo.toml")
	}
	// tools keys: typed by the catalog (D23, D86).
	for _, c := range []struct{ key, value, want string }{
		{"tools.a.count", "3", "count = 3"},
		{"tools.a.flag", "true", "flag = true"},
		{"tools.a.name", "n", `name = "n"`},
		{"tools.a.list", `["y"]`, `list = ["y"]`},
		{"tools.a.env", `["dev"]`, `env = ["dev"]`},
		{"tools.a.files", `{".gitignore":{"x":{"$mode":"remove"}}}`, `[tools.a.files.".gitignore".x]`},
	} {
		code, _, errText := run("config", "set", c.key, c.value)
		if code != fail.OK || !strings.Contains(errText, "run bootstrap sync to apply it") || !strings.Contains(read("tools.toml"), c.want) {
			t.Errorf("set %s: %d %s\n%s", c.key, code, errText, read("tools.toml"))
		}
	}
	for _, c := range []struct{ key, value, msg string }{
		{"tools.a.count", "x", "needs an integer"},
		{"tools.a.bogus", "1", "bogus is not a key of the tool a"},
		{"tools.zz.k", "1", "zz is not a bootstrap tool"},
		{"tools.b.flag", "true", "Run bootstrap add tool b"},
		{"tools.a", "{}", "names the table of a tool"},
	} {
		if code, _, errText := run("config", "set", c.key, c.value); code != fail.Usage || !strings.Contains(errText, c.msg) {
			t.Errorf("set %s: %d %s", c.key, code, errText)
		}
	}
}

func TestConfigReleases(t *testing.T) {
	dir, run := valuesRepo(t)
	file := filepath.Join(dir, "releases.toml")
	steps := []struct{ level, info, want string }{
		{"PATCH", "releases.web = PATCH", "web = \"PATCH\""},
		{"MAJOR", "releases.web = MAJOR", "web = \"MAJOR\""},
		{"MINOR", "releases.web stays MAJOR: it is higher than MINOR", "web = \"MAJOR\""},
	}
	for _, s := range steps {
		code, _, errText := run("config", "set", "releases.web", s.level)
		b, _ := os.ReadFile(file)
		if code != fail.OK || !strings.Contains(errText, s.info) || strings.Contains(errText, "sync") || !strings.Contains(string(b), s.want) || !strings.HasPrefix(string(b), "format = 3") {
			t.Errorf("set %s: %d %s\n%s", s.level, code, errText, b)
		}
	}
	if code, _, _ := run("config", "set", "releases.web", "HUGE"); code != fail.Usage {
		t.Errorf("a bad level: %d", code)
	}
	if code, _, _ := run("config", "unset", "releases.web"); code != fail.OK {
		t.Errorf("unset: %d", code)
	}
	if b, _ := os.ReadFile(file); strings.Contains(string(b), "web") {
		t.Errorf("unset left:\n%s", b)
	}
}

func TestConfigUnset(t *testing.T) {
	dir, run := valuesRepo(t)
	if code, _, errText := run("config", "unset", "repo.repo_name"); code != fail.Usage || !strings.Contains(errText, "required key") {
		t.Errorf("unset a required key: %d %s", code, errText)
	}
	if code, _, _ := run("config", "unset", "repo.branches.main"); code != fail.OK {
		t.Errorf("unset a table: %d", code)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "repo.toml"))
	if strings.Contains(string(b), "[branches") {
		t.Errorf("repo.toml:\n%s", b)
	}
	if code, _, errText := run("config", "unset", "repo.product_name"); code != fail.OK || !strings.Contains(errText, "is not set") {
		t.Errorf("unset an absent key: %d %s", code, errText)
	}
	if code, _, _ := run("config", "unset", "tools.a.flag"); code != fail.OK {
		t.Errorf("unset a tool key: %d", code)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "tools.toml")); strings.Contains(string(b), "flag") || !strings.Contains(string(b), "[tools.a]") {
		t.Errorf("tools.toml:\n%s", b)
	}
}

func TestConfigLayout(t *testing.T) {
	dir, run := valuesRepo(t)
	root := filepath.Dir(filepath.Dir(dir))
	// D24: no values folder.
	if err := os.Rename(dir, root+"/kept"); err != nil {
		t.Fatal(err)
	}
	if code, _, errText := run("config", "get", "repo.repo_name"); code != fail.Usage || !strings.Contains(errText, "Run bootstrap init") {
		t.Errorf("no folder: %d %s", code, errText)
	}
	// D28: the old file alone.
	writeFile(t, root+"/.config/bootstrap.toml", "format = 2\n", 0o644)
	if code, _, errText := run("config", "get", "repo.repo_name"); code != fail.Usage || !strings.Contains(errText, "old form") {
		t.Errorf("old form: %d %s", code, errText)
	}
	// D19: BOOTSTRAP_VALUES skips the stop of the old form.
	t.Setenv("BOOTSTRAP_VALUES", root+"/kept")
	if code, out, errText := run("config", "get", "repo.repo_name"); code != fail.OK || !strings.Contains(out, `"x"`) {
		t.Errorf("BOOTSTRAP_VALUES: %d %s %s", code, out, errText)
	}
	t.Setenv("BOOTSTRAP_VALUES", "")
	// D29: both.
	if err := os.Rename(root+"/kept", dir); err != nil {
		t.Fatal(err)
	}
	if code, _, errText := run("config", "get", "repo.repo_name"); code != fail.Usage || !strings.Contains(errText, "Remove one") {
		t.Errorf("both: %d %s", code, errText)
	}
	// D88: a folder without tools.toml.
	_ = os.Remove(root + "/.config/bootstrap.toml")
	_ = os.Remove(dir + "/tools.toml")
	if code, _, errText := run("config", "get", "repo.repo_name"); code != fail.Usage || !strings.Contains(errText, "has no tools.toml") {
		t.Errorf("no tools.toml: %d %s", code, errText)
	}
	// settings needs no values.
	if code, _, _ := run("settings"); code != fail.OK {
		t.Errorf("settings: %d", code)
	}
}
