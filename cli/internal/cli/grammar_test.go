package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
)

// TestGrammar: add takes the kind word tool or pack, remove and replace take
// tool (D1). A name with no kind word is the old form, with the hint of the
// new form (D8). --exclude is for add pack only (D2). init takes no names
// (D17 i). detect has no --no, and --yes and --answer need --add (D7, D17 d).
func TestGrammar(t *testing.T) {
	cases := []struct {
		name string
		args []string
		msgs []string
	}{
		{"add with no kind", []string{"add", "node"}, []string{"add needs the kind word tool or pack before the names", "-> Use: bootstrap add tool node"}},
		{"mac add with no kind", []string{"mac", "add", "node", "uv"}, []string{"-> Use: bootstrap mac add tool node uv"}},
		{"remove with no kind", []string{"remove", "node"}, []string{"remove needs the kind word tool before the names", "-> Use: bootstrap remove tool node"}},
		{"mac remove with no kind", []string{"mac", "remove", "node"}, []string{"-> Use: bootstrap mac remove tool node"}},
		{"replace with no kind", []string{"replace", "a", "b"}, []string{"-> Use: bootstrap replace tool a b"}},
		{"remove pack", []string{"remove", "pack", "dev"}, []string{"remove takes the kind word tool only - there is no remove pack"}},
		{"replace pack", []string{"mac", "replace", "pack", "a", "b"}, []string{"mac replace takes the kind word tool only"}},
		{"add alone", []string{"add"}, []string{"add needs the kind word tool or pack", "usage: bootstrap add tool|pack <name...>"}},
		{"add tool alone", []string{"add", "tool"}, []string{"add tool needs at least one tool"}},
		{"add pack alone", []string{"mac", "add", "pack"}, []string{"mac add pack needs at least one pack"}},
		{"remove tool alone", []string{"remove", "tool"}, []string{"remove tool needs at least one tool"}},
		{"replace tool with one name", []string{"replace", "tool", "a"}, []string{"replace tool needs two tools", "usage: bootstrap replace tool <old> <new>"}},
		{"exclude of add tool", []string{"add", "tool", "a", "--exclude", "b"}, []string{"--exclude is only for add pack"}},
		{"exclude of remove", []string{"remove", "tool", "a", "--exclude", "b"}, []string{"remove has no flag --exclude"}},
		{"exclude with no value", []string{"add", "pack", "a", "--exclude="}, []string{"--exclude needs a value"}},
		{"init with names", []string{"init", "node", "uv"}, []string{"init takes no tool names", "-> Use: bootstrap add tool node uv"}},
		{"mac init with names", []string{"mac", "init", "a"}, []string{"-> Use: bootstrap mac add tool a"}},
		{"detect --no", []string{"detect", "--no"}, []string{"detect has no flag --no"}},
		{"detect --yes", []string{"detect", "--yes"}, []string{"detect --yes needs --add", "-> Add --add"}},
		{"mac detect --answer", []string{"mac", "detect", "--answer", "a.b=c"}, []string{"mac detect --answer needs --add"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setupEnv(t, "/nonexistent/bootstrap/config.toml")
			var stdout, stderr bytes.Buffer
			if code := Run("dev", c.args, strings.NewReader(""), &stdout, &stderr); code != fail.Usage {
				t.Errorf("code %d, want %d; stderr %q", code, fail.Usage, stderr.String())
			}
			for _, m := range c.msgs {
				if !strings.Contains(stderr.String(), m) {
					t.Errorf("stderr %q lacks %q", stderr.String(), m)
				}
			}
		})
	}
}

// TestParseKind: the kind word, the names after it, --exclude and the flags
// of init and detect.
func TestParseKind(t *testing.T) {
	var stderr bytes.Buffer
	a := &app{stderr: &stderr}
	o, err := parse("add", []string{"pack", "dev", "--exclude", "uv", "--exclude=pack:security"}, "dry-run yes exclude answer jobs verbose", true)
	if err != nil {
		t.Fatal(err)
	}
	if code := a.kindWord(false, "add", "", &o); code != fail.OK {
		t.Fatalf("code %d: %s", code, stderr.String())
	}
	want := core.Item{Kind: core.AddPack, Names: []string{"dev"}, Exclude: []string{"uv", "pack:security"}}
	if got := itemOf("add", o); !reflect.DeepEqual(got, want) {
		t.Errorf("item %+v, want %+v", got, want)
	}
	for _, c := range []struct {
		cmd  string
		args []string
		want core.Item
	}{
		{"add", []string{"tool", "a", "b"}, core.Item{Kind: core.AddTool, Names: []string{"a", "b"}}},
		{"remove", []string{"tool", "a"}, core.Item{Kind: core.RemoveTool, Names: []string{"a"}}},
		{"replace", []string{"tool", "a", "b"}, core.Item{Kind: core.Replace, Old: "a", New: "b"}},
	} {
		o, err := parse(c.cmd, c.args, "", true)
		if err != nil {
			t.Fatal(err)
		}
		if code := a.kindWord(false, c.cmd, "", &o); code != fail.OK {
			t.Fatalf("%s: code %d", c.cmd, code)
		}
		if got := itemOf(c.cmd, o); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: item %+v, want %+v", c.cmd, got, c.want)
		}
	}
	o, err = parse("init", []string{"--no-tui"}, "dry-run yes no-tui", false)
	if err != nil || !o.noTUI {
		t.Errorf("init --no-tui: %v %+v", err, o)
	}
	o, err = parse("detect", []string{"--add", "-y"}, "dry-run add yes answer", false)
	if err != nil || !o.add || !o.yes {
		t.Errorf("detect --add -y: %v %+v", err, o)
	}
}

// detectRepo makes the fixture of the detect tests: a clone with the tools a
// (a required member of base) and c, whose detect scripts match, and a
// repository whose values file selects a. It gives the values file, its
// first text and a function that runs bootstrap with the clone.
func detectRepo(t *testing.T) (string, string, func(args ...string) (fail.Code, string, string)) {
	t.Helper()
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise is not available")
	}
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	clone, root := t.TempDir(), t.TempDir()
	const keys = catalogKeys + "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"
	writeFile(t, filepath.Join(clone, "cli", "VERSION"), "0.0.1\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "categories.toml"), catalogFile, 0o644)
	writeFile(t, filepath.Join(clone, "repo", "packs.toml"), "[pack.base]\na = { required = true }\n", 0o644)
	for tool, out := range map[string]string{
		"a": `{"result":"match","evidence":["f"],"message":"m"}`,
		"c": `{"result":"match","evidence":["g"],"message":"n"}`,
	} {
		writeFile(t, filepath.Join(clone, "repo", "tools", tool, "tools.toml"), keys, 0o644)
		writeFile(t, filepath.Join(clone, "repo", "tools", tool, "scripts", "detect"), "#!/bin/bash\necho '"+out+"'\n", 0o755)
	}
	values := filepath.Join(root, ".config", "bootstrap.toml")
	const text = "format = 2\nrepo_name = \"x\"\n\n[tools.a]\n"
	writeFile(t, values, text, 0o644)
	// The sync copies a clone with a commit, and writes only to a clean tree.
	for _, dir := range []string{clone, root} {
		for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--no-gpg-sign", "-m", "c"}} {
			if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
				t.Fatalf("git %v: %s", args, out)
			}
		}
	}
	t.Setenv("MISE_PROJECT_ROOT", root)
	t.Setenv("BOOTSTRAP_VALUES", "")
	t.Setenv("PWD", root)
	return values, text, func(args ...string) (fail.Code, string, string) {
		var stdout, stderr bytes.Buffer
		code := Run("dev", append(args, "--clone", clone), strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
}

// writeFile writes <text> to <p> with the mode <mode>, and makes its folder.
func writeFile(t *testing.T, p, text string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), mode); err != nil {
		t.Fatal(err)
	}
}

// TestDetectList: detect with no flag lists the tools, and exits 1 when a
// detected tool is not selected; it writes nothing (D7).
func TestDetectList(t *testing.T) {
	values, text, run := detectRepo(t)
	code, out, errs := run("detect")
	if code != fail.Other || !strings.Contains(out, "c") || !strings.Contains(errs, "c is found, but it is not selected") ||
		!strings.Contains(errs, "Run bootstrap add tool c, or bootstrap detect --add") {
		t.Errorf("code %d stdout %q stderr %q", code, out, errs)
	}
	// --dry-run without --add lists too.
	if code, _, errs := run("detect", "--dry-run"); code != fail.Other {
		t.Errorf("--dry-run: code %d stderr %q", code, errs)
	}
	if b, _ := os.ReadFile(values); string(b) != text {
		t.Errorf("detect wrote the values file: %q", b)
	}
	// Each detected tool selected: exit 0.
	writeFile(t, values, text+"[tools.c]\n", 0o644)
	if code, out, errs := run("detect"); code != fail.OK || !strings.Contains(out, "No tool to add") {
		t.Errorf("selected: code %d stdout %q stderr %q", code, out, errs)
	}
}

// TestDetectAdd: detect --add asks the consent, and --yes answers it; it
// writes a table for each detected tool that is not selected, and syncs one
// time (D7, D9). --add --dry-run writes nothing.
func TestDetectAdd(t *testing.T) {
	values, text, run := detectRepo(t)
	if code, _, errs := run("detect", "--add", "--dry-run"); code != fail.OK {
		t.Errorf("--add --dry-run: code %d stderr %q", code, errs)
	}
	if b, _ := os.ReadFile(values); string(b) != text {
		t.Errorf("--add --dry-run wrote the values file: %q", b)
	}
	// No terminal and no --yes: no consent, nothing written.
	if code, _, errs := run("detect", "--add"); code != fail.Usage || !strings.Contains(errs, "changes need --yes") {
		t.Errorf("--add with no consent: code %d stderr %q", code, errs)
	}
	if b, _ := os.ReadFile(values); string(b) != text {
		t.Errorf("--add with no consent wrote the values file: %q", b)
	}
	code, out, errs := run("detect", "--add", "--yes")
	if code != fail.OK {
		t.Fatalf("--add --yes: code %d stdout %q stderr %q", code, out, errs)
	}
	b, _ := os.ReadFile(values)
	if !strings.Contains(string(b), "[tools.c]") || !strings.Contains(string(b), "[tools.a]") {
		t.Errorf("--add --yes: values file %q", b)
	}
	// Then each detected tool is selected.
	if code, out, errs := run("detect"); code != fail.OK || !strings.Contains(out, "No tool to add") {
		t.Errorf("after --add: code %d stdout %q stderr %q", code, out, errs)
	}
}

// TestMacDetectAdd: mac detect --add --yes writes a table for each detected
// Mac tool, and each installed package of no Mac tool as an AddPackage item
// to the own [mise] tables of macos.toml, in one write and one sync (D7,
// D16 f). The fake brew, mas and mise give the installed packages; the fake
// mise runs the real one for each command but ls.
func TestMacDetectAdd(t *testing.T) {
	mise, err := exec.LookPath("mise")
	if err != nil {
		t.Skip("mise is not available")
	}
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	// The sync needs mise at ~/.local/bin/mise: the fake one is there.
	clone, cfg, home := t.TempDir(), t.TempDir(), t.TempDir()
	bin := filepath.Join(home, ".local", "bin")
	writeFile(t, filepath.Join(clone, "cli", "VERSION"), "0.0.1\n", 0o644)
	writeFile(t, filepath.Join(clone, "macos", "categories.toml"), catalogFile, 0o644)
	writeFile(t, filepath.Join(clone, "macos", "packs.toml"), "[pack.base]\n", 0o644)
	writeFile(t, filepath.Join(clone, "macos", "tools", "pkgtool", "tools.toml"),
		catalogKeys+"\n[mise.bootstrap.packages]\n\"brew:pkgfmla\" = \"latest\"\n", 0o644)
	writeFile(t, filepath.Join(clone, "macos", "tools", "appcmd", "tools.toml"),
		catalogKeys+"\n[mise.bootstrap.packages]\n\"brew:mas\" = \"latest\"\n", 0o644)
	writeFile(t, filepath.Join(bin, "brew"), "#!/bin/sh\ncase \"$1 $2\" in\n"+
		"leaves*|\"list --formula\") printf 'pkgfmla\\nloosefmla\\n' ;;\n\"list --cask\") echo loosecask ;;\nesac\n", 0o755)
	writeFile(t, filepath.Join(bin, "mas"), "#!/bin/sh\necho '123  Some App  (1.0)'\n", 0o755)
	writeFile(t, filepath.Join(bin, "mise"), "#!/bin/sh\nif [ \"$1\" = ls ]; then echo '{\"loosetool\": []}'; exit 0; fi\n"+
		"exec '"+mise+"' \"$@\"\n", 0o755)
	values := filepath.Join(cfg, "macos.toml")
	const text = "format = 2\n"
	writeFile(t, values, text, 0o644)
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--no-gpg-sign", "-m", "c"}} {
		if out, err := exec.Command("git", append([]string{"-C", cfg}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("MACOS_SETUP_REPO_PATH", cfg)
	t.Setenv("PWD", cfg)
	run := func(args ...string) (fail.Code, string, string) {
		var stdout, stderr bytes.Buffer
		code := Run("dev", append(append([]string{"mac"}, args...), "--no-bootstrap", "--clone", clone), strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	code, out, errs := run("detect", "--add", "--yes")
	if code != fail.OK {
		t.Fatalf("mac detect --add --yes: code %d stdout %q stderr %q", code, out, errs)
	}
	b, _ := os.ReadFile(values)
	for _, line := range []string{
		"[tools.pkgtool]", "[tools.appcmd]", "[mise.tools.loosetool]",
		`"brew:loosefmla" = "latest"`, `"brew-cask:loosecask" = "latest"`, `"mas:123" = "latest"`,
	} {
		if !strings.Contains(string(b), line+"\n") {
			t.Errorf("macos.toml has no line %s: %q", line, b)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "mise", "conf.d", "pkgtool", "mise.toml")); err != nil {
		t.Errorf("the sync did not write the file of pkgtool: %v", err)
	}
	if code, out, errs := run("detect"); code != fail.OK || !strings.Contains(out, "No tool to add") {
		t.Errorf("after --add: code %d stdout %q stderr %q", code, out, errs)
	}
}

// TestSkipText: the info line of a member of a pack that add pack leaves out
// (D16 d, D17 g).
func TestSkipText(t *testing.T) {
	for _, c := range []struct {
		s    core.Skip
		want string
	}{
		{core.Skip{Tool: "x", Why: "hosts"}, "x is left out: its hosts leave out this host"},
		{core.Skip{Tool: "x", Why: "selectable"}, "x is left out: it is not selectable; a tool that needs it selects it"},
		{core.Skip{Tool: "x", Why: "root_only"}, "x is left out: it is root_only, and this repository is a submodule"},
	} {
		if got := skipText(c.s); got != c.want {
			t.Errorf("skipText(%v) = %q, want %q", c.s, got, c.want)
		}
	}
}
