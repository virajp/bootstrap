package projects

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

func write(t *testing.T, p, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// gitRepo gives a git repository with the files <files> (path: text).
func gitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	for p, text := range files {
		write(t, filepath.Join(root, p), text)
	}
	return root
}

// TestNormalize: a name is lowercase kebab case (D16).
func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{"My_App": "my-app", "web": "web", "a.b--c": "a-b-c", "___": "", "API2": "api2"} {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
		if want != "" && !NameRule.MatchString(want) {
			t.Errorf("%q breaks the name rule", want)
		}
	}
	for _, bad := range []string{"My", "a_b", "-a", "a-", ""} {
		if NameRule.MatchString(bad) {
			t.Errorf("%q passes the name rule", bad)
		}
	}
}

// TestCleanFolder: a folder is relative to the root; . is the root (D7).
func TestCleanFolder(t *testing.T) {
	for in, want := range map[string]string{".": ".", "./apps/web/": "apps/web", "a//b": "a/b"} {
		if got, ok := CleanFolder(in); !ok || got != want {
			t.Errorf("CleanFolder(%q) = %q %v", in, got, ok)
		}
	}
	for _, bad := range []string{"", "/abs", "../x", "a/../../b"} {
		if _, ok := CleanFolder(bad); ok {
			t.Errorf("CleanFolder(%q) passes", bad)
		}
	}
}

// TestInFolder: a match counts in a folder only with an evidence path in
// that folder itself (D27).
func TestInFolder(t *testing.T) {
	if !InFolder([]string{"app/build.gradle", "go.mod"}) || !InFolder([]string{"./go.mod"}) || !InFolder([]string{".github/"}) {
		t.Error("a path in the folder does not count")
	}
	if InFolder([]string{"app/build.gradle"}) || InFolder(nil) {
		t.Error("a path below the folder counts")
	}
}

// TestInside: the root holds no other project (D15).
func TestInside(t *testing.T) {
	if !Inside("apps/web/src", "apps/web") || !Inside("apps/web", "apps/web") || Inside("apps/webx", "apps/web") || Inside("apps", Root) {
		t.Error("Inside")
	}
}

// TestFolders: the folders of the files that git tracks or does not ignore,
// without .git, node_modules, vendor and the gitignored paths (D8).
func TestFolders(t *testing.T) {
	root := gitRepo(t, map[string]string{
		".gitignore":                  "ignored/\n",
		"a/b/c.txt":                   "",
		"ignored/x/go.mod":            "",
		"node_modules/p/package.json": "",
		"x/vendor/y/go.mod":           "",
		"top.txt":                     "",
	})
	got, err := Folders(root)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"a": true, "a/b": true, "x": true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("folders %v", got)
	}
}

// TestMembers: each workspace manifest, its globs, its excludes, the root
// dropped, and only the folders that exist (D8, D24, D28, D29).
func TestMembers(t *testing.T) {
	root := gitRepo(t, map[string]string{
		"pnpm-workspace.yaml":          "packages:\n  - \"apps/*\"\n  - \"!apps/docs\"\n  - \".\"\n  - \"libs/**\"\n  - \"gone\"\n",
		"aube-workspace.yaml":          "packages: [\"apps/web\"]\n",
		"go.work":                      "go 1.22\n\nuse ./svc/a // the a service\nuse (\n\t./svc/b\n\t.\n)\n",
		"Cargo.toml":                   "[workspace]\nmembers = [\"crates/*\"]\nexclude = [\"crates/old\"]\n",
		"pyproject.toml":               "[tool.uv.workspace]\nmembers = [\"py/*\"]\nexclude = [\"py/skip\"]\n",
		"apps/web/package.json":        "{}",
		"apps/docs/package.json":       "{}",
		"apps/empty/README.md":         "",
		"libs/a/b/package.json":        "{}",
		"svc/a/go.mod":                 "",
		"svc/b/go.mod":                 "",
		"crates/x/Cargo.toml":          "",
		"crates/old/Cargo.toml":        "",
		"py/one/pyproject.toml":        "",
		"py/skip/pyproject.toml":       "",
		"apps/web/node_modules/x/a.js": "",
	})
	dirs, err := Folders(root)
	if err != nil {
		t.Fatal(err)
	}
	for m, want := range map[string][]string{
		"pnpm-workspace.yaml": {"apps/web", "libs/a/b"},
		"aube-workspace.yaml": {"apps/web"},
		"go.work":             {"svc/a", "svc/b"},
		"Cargo.toml":          {"crates/x"},
		"pyproject.toml":      {"py/one"},
	} {
		got, err := Members(root, m, dirs)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%s: members %v (%v), want %v", m, got, err, want)
		}
	}
	if got, err := Members(root, "Cargo.toml", nil); err != nil || got != nil {
		t.Errorf("no folders: %v %v", got, err)
	}
	if !IsManifest("./go.work") || IsManifest("apps/web/package.json") {
		t.Error("IsManifest")
	}
}

// TestMembersNotThere: a manifest that is not there, or without a
// workspace, gives no member.
func TestMembersNotThere(t *testing.T) {
	root := gitRepo(t, map[string]string{"pyproject.toml": "[project]\nname = \"x\"\n", "a/b.txt": ""})
	dirs, _ := Folders(root)
	for _, m := range Manifests {
		if got, err := Members(root, m, dirs); err != nil || got != nil {
			t.Errorf("%s: %v %v", m, got, err)
		}
	}
}

// TestGlob: ** matches zero or more segments.
func TestGlob(t *testing.T) {
	dirs := map[string]bool{"a": true, "a/b": true, "a/b/c": true, "x": true}
	if got := Glob("a/**", dirs); strings.Join(got, " ") != "a a/b a/b/c" {
		t.Errorf("a/** = %v", got)
	}
	if got := Glob("*/b", dirs); strings.Join(got, " ") != "a/b" {
		t.Errorf("*/b = %v", got)
	}
	if got := Glob(".", dirs); got != nil {
		t.Errorf(". = %v", got)
	}
}

// TestListedAndLinks reads the projects of repo.toml and the projects list
// of a tool.
func TestListedAndLinks(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "v.toml")
	write(t, p, "[projects.web]\nfolder = \"./apps/web\"\n[projects.api]\nfolder = \"apps/api\"\n[tools.go]\nprojects = [\"api\"]\n")
	raw, err := source.ReadTOML(p)
	if err != nil {
		t.Fatal(err)
	}
	want := []Project{{"api", "apps/api"}, {"web", "apps/web"}}
	if got := Listed(raw); !reflect.DeepEqual(got, want) {
		t.Errorf("listed %v", got)
	}
	if got := Links(raw, "go"); !reflect.DeepEqual(got, []string{"api"}) || Links(raw, "node") != nil {
		t.Errorf("links %v", got)
	}
}
