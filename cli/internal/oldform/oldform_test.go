package oldform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRepoFindsEachOldForm(t *testing.T) {
	root := t.TempDir()
	values := root + "/.config/bootstrap.toml"
	write(t, values, "commit_scopes = [\"a\"]\nextensions = [\"x.y\"]\n[tools.repo-sync]\n[overrides.gitignore]\nadd = [\"a/\"]\n")
	write(t, root+"/"+OldPrecommit, "repos: []\n")
	raw, err := source.ReadTOML(values)
	if err != nil {
		t.Fatal(err)
	}
	msgs := source.Messages(Repo(root, raw, "v.toml"))
	if len(msgs) != 6 {
		t.Fatalf("messages = %v", msgs)
	}
	for _, m := range msgs {
		if !strings.HasSuffix(m.Error(), "is the old form - run bootstrap init") {
			t.Errorf("message = %q", m.Error())
		}
	}
	Migrating = true
	defer func() { Migrating = false }()
	if n := len(source.Messages(Repo(root, raw, "v.toml"))); n != 5 {
		t.Errorf("messages while migrating = %d, want 5", n)
	}
}

func TestMigrateRepo(t *testing.T) {
	root := t.TempDir()
	values := root + "/.config/bootstrap.toml"
	write(t, values, `# The header.
repo_name = "x"
commit_scopes = [ "api" ]
worktree_files = [ ".env" ]
extensions = [ "foo.bar" ]

[tools.node]

[branch.main]
# Inside.
direct_commit = false

[claude.plugins.project]
vwf = true

[overrides.gitignore]
add = [ "docs/prompts/" ]
remove = [ "*.log" ]

[tools.repo-sync]
`)
	write(t, root+"/.gitignore", "node_modules/\n")
	write(t, root+"/"+OldPrecommit, "repos:\n  - repo: https://example.com/own\n    hooks:\n      - id: own\n")
	write(t, root+"/"+precommit, "# Kept.\nrepos:\n  - repo: https://example.com/base\n    hooks:\n      - id: base\n")
	t.Setenv("BOOTSTRAP_CLONE", repoClone(t, nil))
	changes, err := MigrateRepo(root, values)
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(changes); err != nil {
		t.Fatal(err)
	}
	raw, err := source.ReadTOML(values)
	if err != nil {
		t.Fatal(err)
	}
	if err := Repo(root, raw, "v.toml"); err != nil {
		t.Errorf("an old form stays: %v", err)
	}
	tools := raw.Get("tools").(*source.Obj)
	git := tools.Get("git").(*source.Obj)
	if source.JSON(git.Get("commit_scopes")) != `["api"]` {
		t.Errorf("commit_scopes = %s", source.JSON(git.Get("commit_scopes")))
	}
	if !strings.Contains(source.JSON(git), `"*.log":{"$mode":"remove"}`) {
		t.Errorf("[tools.git] = %s", source.JSON(git))
	}
	if !strings.Contains(source.JSON(tools.Get("claude")), `"vwf":true`) {
		t.Errorf("[tools.claude] = %s", source.JSON(tools.Get("claude")))
	}
	// The frozen pack base: each member with a table, except vscode (not
	// required in the new base) and _security_base (selectable = false).
	if got := strings.Join(tools.Keys(), " "); got != "claude dprint git gitleaks grype linter mise node osv-scanner pre-commit" {
		t.Errorf("the tables = %s", got)
	}
	if !IsFormat(raw) {
		t.Errorf("format = %v", raw.Get(FormatKey))
	}
	b, _ := os.ReadFile(values)
	for _, want := range []string{"# The header.", "# Inside."} {
		if !strings.Contains(string(b), want) {
			t.Errorf("the comment %q is gone:\n%s", want, b)
		}
	}
	if b, _ := os.ReadFile(root + "/.gitignore"); string(b) != "node_modules/\ndocs/prompts/\n" {
		t.Errorf(".gitignore = %q", b)
	}
	if b, _ := os.ReadFile(root + "/" + extensions); !strings.Contains(string(b), `"foo.bar"`) {
		t.Errorf("extensions.json = %q", b)
	}
	b, _ = os.ReadFile(root + "/" + precommit)
	if !strings.Contains(string(b), "# Kept.") || !strings.Contains(string(b), "https://example.com/own") || !strings.Contains(string(b), "https://example.com/base") {
		t.Errorf("pre-commit config = %s", b)
	}
	if isFile(root + "/" + OldPrecommit) {
		t.Error("the old hook file stays")
	}
	again, err := MigrateRepo(root, values)
	if err != nil || len(again) != 0 {
		t.Errorf("a second migration = %v, %v", again, err)
	}
}

func TestMacFindsTheOldForm(t *testing.T) {
	values := t.TempDir() + "/macos.toml"
	write(t, values, "format = 2\n[tools.macos-setup]\n")
	raw, err := source.ReadTOML(values)
	if err != nil {
		t.Fatal(err)
	}
	msgs := source.Messages(Mac(raw, "macos.toml"))
	if len(msgs) != 1 || msgs[0].Error() != "[tools.macos-setup] in macos.toml is the old form - run bootstrap mac init" {
		t.Errorf("messages = %v", msgs)
	}
}

func TestMigrateMac(t *testing.T) {
	values := t.TempDir() + "/macos.toml"
	write(t, values, "# Mac.\n[tools.macos-setup]\nconfig_path = \"~/m\"\n\n[tools.mise]\nx = 1\n\n[packs.ai]\n")
	t.Setenv("BOOTSTRAP_CLONE", macClone(t))
	c, err := MigrateMac(values)
	if err != nil || c == nil {
		t.Fatalf("change = %v, %v", c, err)
	}
	if err := Apply([]Change{*c}); err != nil {
		t.Fatal(err)
	}
	raw, _ := source.ReadTOML(values)
	if err := Mac(raw, "macos.toml"); err != nil {
		t.Errorf("the old form stays: %v", err)
	}
	mise := raw.Get("tools").(*source.Obj).Get("mise").(*source.Obj)
	if mise.Get("config_path") != "~/m" || mise.Get("x") == nil {
		t.Errorf("[tools.mise] = %s", source.JSON(mise))
	}
	// The frozen base and the frozen pack ai: uv is not required in the new
	// base, and devtools is not in the clone.
	if got := strings.Join(raw.Get("tools").(*source.Obj).Keys(), " "); got != "claude-desktop homebrew mise" || raw.Has("packs") || !IsFormat(raw) {
		t.Errorf("the tables = %s, packs %v, format %v", got, raw.Has("packs"), raw.Get(FormatKey))
	}
	if c, err := MigrateMac(values); c != nil || err != nil {
		t.Errorf("a second migration = %v, %v", c, err)
	}
}

// clone writes a clone with the tools <tools> of <side> (name to the extra
// text of its tools.toml) and the packs file <packs>, and gives its root.
func clone(t *testing.T, side source.Side, tools map[string]string, packs string) string {
	t.Helper()
	root := t.TempDir()
	dir := root + "/" + string(side)
	write(t, dir+"/categories.toml", "[category.core]\ntitle = \"Core\"\n")
	write(t, dir+"/packs.toml", packs)
	for name, extra := range tools {
		head := "summary = \"A tool\"\ncategory = \"core\"\n"
		if side == source.Repo {
			head += "default_envs = [ \"dev\" ]\n"
			if !strings.Contains(extra, "root_only") {
				head += "root_only = false\n"
			}
			if !strings.Contains(extra, "depends") {
				head += "depends = []\n"
			}
		}
		write(t, dir+"/tools/"+name+"/tools.toml", head+extra)
	}
	return root
}

// repoClone is a repo clone with the tools of the frozen pack base, node and
// claude. <over> replaces the text of a tool.
func repoClone(t *testing.T, over map[string]string) string {
	tools := map[string]string{
		"mise": "", "git": "", "pre-commit": "", "dprint": "", "linter": "", "vscode": "",
		"gitleaks": "depends = [ \"_security_base\" ]\n", "grype": "", "osv-scanner": "",
		"_security_base": "selectable = false\n", "node": "", "claude": "",
	}
	for k, v := range over {
		tools[k] = v
	}
	return clone(t, source.Repo, tools, "[pack.base]\nmise = { required = true }\ngit = { required = true }\npre-commit = { required = true }\n"+
		"dprint = { required = true }\nlinter = { required = true }\ngitleaks = { required = true }\ngrype = { required = true }\n"+
		"osv-scanner = { required = true }\nvscode = {}\n")
}

// macClone is a Mac clone with a few tools of the frozen packs.
func macClone(t *testing.T) string {
	return clone(t, source.Macos, map[string]string{"mise": "", "homebrew": "", "uv": "", "claude-desktop": "", "claude-cli": ""},
		"[pack.base]\nmise = { required = true }\nhomebrew = { required = true }\nuv = {}\n\n[pack.claude]\nclaude-desktop = {}\nclaude-cli = {}\n")
}

func TestFormatKey(t *testing.T) {
	cases := []struct{ text, want string }{
		{"format = 2\n[tools.a]\n", ""},
		{"[tools.a]\n", "v.toml without format = 2 is the old form - run bootstrap init"},
		{"format = 3\n", "format = 3 in v.toml is not valid - the format of a values file is 2"},
		{"format = \"2\"\n", "format = \"2\" in v.toml is not valid - the format of a values file is 2"},
		{"format = 2\n[packs.security]\n", "[packs.security] in v.toml is the old form - run bootstrap init"},
	}
	for _, c := range cases {
		values := t.TempDir() + "/v.toml"
		write(t, values, c.text)
		raw, err := source.ReadTOML(values)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, m := range source.Messages(Repo(t.TempDir(), raw, "v.toml")) {
			got = append(got, m.Error())
		}
		if strings.Join(got, "\n") != c.want {
			t.Errorf("%q: messages = %q, want %q", c.text, got, c.want)
		}
		// The Mac side has the same rules, with its hint.
		var mac []string
		for _, m := range source.Messages(Mac(raw, "v.toml")) {
			mac = append(mac, m.Error())
		}
		if want := strings.ReplaceAll(c.want, "run bootstrap init", MacHint); strings.Join(mac, "\n") != want {
			t.Errorf("%q: Mac messages = %q, want %q", c.text, mac, want)
		}
	}
}

// migrated runs TablesIn on the values text <text> and gives the values file
// after the edit.
func migrated(t *testing.T, text string, side source.Side, root string, sub bool) (*source.Obj, []string) {
	t.Helper()
	values := t.TempDir() + "/v.toml"
	write(t, values, text)
	raw, err := source.ReadTOML(values)
	if err != nil {
		t.Fatal(err)
	}
	exprs, notes, err := TablesIn(raw, "v.toml", side, root, sub)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(values)
	next, err := editTOML(b, exprs)
	if err != nil {
		t.Fatal(err)
	}
	write(t, values, string(next))
	out, err := source.ReadTOML(values)
	if err != nil {
		t.Fatal(err)
	}
	return out, notes
}

func TestMigratePacks(t *testing.T) {
	root := repoClone(t, map[string]string{"linter": "root_only = true\n"})
	// A file of the new form: the pack tables only, no base. Two packs give
	// gitleaks the union of their env; a table keeps its own env; a table of
	// a tool that the clone does not have goes.
	raw, notes := migrated(t, "format = 2\n\n[packs.base]\nenv = [ \"ci\" ]\n\n[packs.security]\nenv = [ \"dev\" ]\n\n"+
		"[tools.grype]\nenv = [ \"dev\" ]\n\n[tools.devtools]\n", source.Repo, root, true)
	tools := raw.Get("tools").(*source.Obj)
	want := map[string]string{
		"mise": `["ci"]`, "git": `["ci"]`, "pre-commit": `["ci"]`, "dprint": `["ci"]`,
		"gitleaks": `["ci","dev"]`, "grype": `["dev"]`, "osv-scanner": `["ci","dev"]`,
	}
	if got := strings.Join(tools.Keys(), " "); got != "dprint git gitleaks grype mise osv-scanner pre-commit" {
		t.Errorf("the tables = %s (linter is root_only in a submodule, vscode is not required)", got)
	}
	for k, w := range want {
		if e := source.JSON(tools.Get(k).(*source.Obj).Get("env")); e != w {
			t.Errorf("[tools.%s] env = %s, want %s", k, e, w)
		}
	}
	if raw.Has("packs") {
		t.Error("the [packs] tables stay")
	}
	if !strings.Contains(strings.Join(notes, "; "), "[tools.devtools] removed") {
		t.Errorf("notes = %v", notes)
	}
	// A pack without env gives no env, also when another pack has one.
	raw, _ = migrated(t, "format = 2\n[packs.base]\n[packs.security]\nenv = [ \"dev\" ]\n", source.Repo, root, false)
	if e := raw.Get("tools").(*source.Obj).Get("gitleaks").(*source.Obj); e.Has("env") {
		t.Errorf("[tools.gitleaks] = %s", source.JSON(e))
	}
	// An unknown pack stops the migration.
	values := t.TempDir() + "/v.toml"
	write(t, values, "format = 2\n[packs.nope]\n")
	r, _ := source.ReadTOML(values)
	if _, _, err := TablesIn(r, "v.toml", source.Repo, root, false); err == nil || !strings.Contains(err.Error(), "nope is not a bootstrap pack") {
		t.Errorf("err = %v", err)
	}
	// A file of the new form without packs needs nothing.
	if exprs, notes, err := TablesIn(raw, "v.toml", source.Repo, root, false); len(exprs) != 0 || len(notes) != 0 || err != nil {
		t.Errorf("again: %v %v %v", exprs, notes, err)
	}
}

func TestFrozenTools(t *testing.T) {
	packs, err := frozenPacks(source.Repo)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := frozenTools(packs, basePack)
	if !ok || strings.Join(got, " ") != "dprint git linter mise gitleaks grype osv-scanner pre-commit vscode" {
		t.Errorf("base = %v %v", got, ok)
	}
	mac, err := frozenPacks(source.Macos)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := frozenTools(mac, "ai"); strings.Join(got, " ") != "claude-desktop" {
		t.Errorf("ai = %v", got)
	}
	if _, ok := frozenTools(mac, "nope"); ok {
		t.Error("an unknown pack")
	}
}
