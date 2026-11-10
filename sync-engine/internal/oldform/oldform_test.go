package oldform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/source"
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
	if len(msgs) != 5 {
		t.Fatalf("messages = %v", msgs)
	}
	for _, m := range msgs {
		if !strings.HasSuffix(m.Error(), "is the old form - run bootstrap init") {
			t.Errorf("message = %q", m.Error())
		}
	}
	Migrating = true
	defer func() { Migrating = false }()
	if n := len(source.Messages(Repo(root, raw, "v.toml"))); n != 4 {
		t.Errorf("messages while migrating = %d, want 4", n)
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

func TestMigrateMac(t *testing.T) {
	values := t.TempDir() + "/macos.toml"
	write(t, values, "# Mac.\n[tools.macos-setup]\nconfig_path = \"~/m\"\n\n[tools.mise]\nx = 1\n")
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
	if c, err := MigrateMac(values); c != nil || err != nil {
		t.Errorf("a second migration = %v, %v", c, err)
	}
}
