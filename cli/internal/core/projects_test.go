package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/projects"
)

// A detect script that matches when the file of its first argument is in
// its folder, with that file as the evidence.
const fileScript = `#!/bin/bash
cat > /dev/null
if [ -f %s ]; then echo '{"result":"match","evidence":["%s"],"message":"found"}'
else echo '{"result":"none","evidence":[],"message":"no"}'; fi
`

// A detect script that searches two folders down, as find -maxdepth does:
// its evidence is the path below its folder.
const deepScript = `#!/bin/bash
cat > /dev/null
f=$(find . -maxdepth 2 -name %s | head -1)
if [ -n "$f" ]; then echo "{\"result\":\"match\",\"evidence\":[\"${f#./}\"],\"message\":\"found\"}"
else echo '{"result":"none","evidence":[],"message":"no"}'; fi
`

// projectTool is a tool of the project fixture: its category, root_only,
// and its detect script.
type projectTool struct {
	category string
	rootOnly bool
	script   string
}

// projectFixture is a repo side with the tools <tools>, each selected by
// <selected>, the repo.toml keys <repoExtra>, and the files <files> in the
// git repository of the values.
func projectFixture(t *testing.T, tools map[string]projectTool, selected []string, repoExtra string, files map[string]string) Side {
	t.Helper()
	clone, root := t.TempDir(), t.TempDir()
	write := func(p, text string, mode os.FileMode) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), mode); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(clone, "repo", "categories.toml"), "[category.core]\ntitle = \"Core\"\n[category.languages]\ntitle = \"Languages\"\n[category.package-managers]\ntitle = \"Package managers\"\n[category.containers]\ntitle = \"Containers\"\n", 0o644)
	write(filepath.Join(clone, "repo", "packs.toml"), "[pack.base]\n", 0o644)
	for name, tl := range tools {
		text := "summary = \"A tool\"\ncategory = \"" + tl.category + "\"\ndefault_envs = [ \"dev\" ]\ndepends = []\nroot_only = " + map[bool]string{true: "true", false: "false"}[tl.rootOnly] + "\n"
		write(filepath.Join(clone, "repo", "tools", name, "tools.toml"), text, 0o644)
		write(filepath.Join(clone, "repo", "tools", name, "scripts", "detect"), tl.script, 0o755)
	}
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	sd := Side{Root: root, Clone: clone, Pwd: root, Values: filepath.Join(root, ".config", "bootstrap")}
	write(filepath.Join(sd.Values, "repo.toml"), "format = 3\nrepo_name = \"My_Repo\"\ncore_branches = [ \"main\" ]\n"+repoExtra, 0o644)
	tt := "format = 3\n"
	for _, s := range selected {
		tt += "[tools." + s + "]\n"
	}
	write(sd.EditFile(), tt, 0o644)
	for p, text := range files {
		write(filepath.Join(root, p), text, 0o644)
	}
	return sd
}

func script(f, name string) string { return strings.ReplaceAll(f, "%s", name) }

// monorepoTools are the tools of the monorepo case.
func monorepoTools() map[string]projectTool {
	return map[string]projectTool{
		"golang": {"languages", false, script(fileScript, "go.mod")},
		"nodejs": {"languages", false, script(fileScript, "package.json")},
		"kt":     {"languages", false, script(deepScript, "build.gradle")},
		"pn":     {"package-managers", false, script(fileScript, "pnpm-workspace.yaml")},
		"gitx":   {"core", false, script(fileScript, "README.md")},
		"dock":   {"containers", true, script(fileScript, "Dockerfile")},
		"extra":  {"languages", false, script(fileScript, "extra.txt")},
	}
}

// monorepoFiles are the files of the monorepo case.
var monorepoFiles = map[string]string{
	".gitignore":                  "ignored/\n",
	"pnpm-workspace.yaml":         "packages:\n  - \"apps/*\"\n  - \".\"\n",
	"README.md":                   "",
	"apps/web/package.json":       "{}",
	"apps/web/README.md":          "",
	"apps/web/Dockerfile":         "",
	"apps/web/src/go.mod":         "",
	"apps/My_Api/go.mod":          "",
	"tools/x/web/package.json":    "{}",
	"node_modules/p/package.json": "{}",
	"ignored/go.mod":              "",
	"deep/a/b/c/d/go.mod":         "",
	"android/app/build.gradle":    "",
	"srv/go.mod":                  "",
	"srv/inner/go.mod":            "",
	"only/extra.txt":              "",
	"___/go.mod":                  "",
}

// TestDetectProjects: the project pass on a monorepo (D4, D6, D8, D14-D19,
// D24-D30).
func TestDetectProjects(t *testing.T) {
	sd := projectFixture(t, monorepoTools(), []string{"golang", "nodejs", "kt", "pn", "gitx", "dock"},
		"[projects.listed]\nfolder = \"srv\"\n", monorepoFiles)
	d, err := Detect(sd)
	if err != nil {
		t.Fatal(problems(err))
	}
	// The member of the manifest first; then the walk, a depth at a time:
	// android/app (kt matches in android only by app/build.gradle, D27),
	// apps/My_Api, ___ (no name, D16), only (a languages tool of no
	// selection), tools/x/web. srv is listed (D17), srv/inner and
	// apps/web/src are inside a project, node_modules and ignored are
	// skipped, and deep/a/b/c/d is 5 levels down (D18). The two web folders
	// take the path form (D8). The root is no project (D15).
	want := []projects.Project{
		{Name: "apps-web", Folder: "apps/web"},
		{Name: "only", Folder: "only"},
		{Name: "app", Folder: "android/app"},
		{Name: "my-api", Folder: "apps/My_Api"},
		{Name: "tools-x-web", Folder: "tools/x/web"},
	}
	if !reflect.DeepEqual(d.Projects, want) {
		t.Errorf("projects %+v", d.Projects)
	}
	if len(d.Warnings) != 1 || !strings.Contains(d.Warnings[0], "the folder ___ is not proposed") {
		t.Errorf("warnings %q", d.Warnings)
	}
	// golang links to its folders, the listed one too (D4, D19); pn only
	// matches at the root and links to the members of its manifest (D14);
	// gitx links by its match in a folder (D25); dock is root_only (D26).
	wantLinks := []Link{
		{Tool: "extra", Projects: []string{"only"}},
		{Tool: "gitx", Projects: []string{"apps-web"}},
		{Tool: "golang", Projects: []string{"listed", "my-api"}},
		{Tool: "kt", Projects: []string{"app"}},
		{Tool: "nodejs", Projects: []string{"apps-web", "tools-x-web"}},
		{Tool: "pn", Projects: []string{"apps-web"}},
	}
	if !reflect.DeepEqual(d.Links, wantLinks) {
		t.Errorf("links %+v", d.Links)
	}
	// extra matches only in a project folder: it is offered (D9).
	if !reflect.DeepEqual(d.Offers, []string{"extra"}) || d.Found["extra"].Evidence[0] != "only/extra.txt" {
		t.Errorf("offers %v, extra %+v", d.Offers, d.Found["extra"])
	}
}

// TestDetectRootProject: with no other project, the root is one project
// named repo_name; only a tool of the categories languages and
// package-managers links to it (D8, D16, D23, D25, D30). A second detect
// with the project and the links written proposes nothing.
func TestDetectRootProject(t *testing.T) {
	tools := monorepoTools()
	sd := projectFixture(t, tools, []string{"golang", "gitx", "pn"}, "", map[string]string{"go.mod": "", "README.md": "", "src/main.go": ""})
	d, err := Detect(sd)
	if err != nil {
		t.Fatal(problems(err))
	}
	if !reflect.DeepEqual(d.Projects, []projects.Project{{Name: "my-repo", Folder: "."}}) {
		t.Errorf("projects %+v", d.Projects)
	}
	if !reflect.DeepEqual(d.Links, []Link{{Tool: "golang", Projects: []string{"my-repo"}}}) {
		t.Errorf("links %+v", d.Links)
	}
	// The plan of detect --add writes both files together (D9, D19).
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Projects: d.Projects, Links: d.Links}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if _, _, _, err := Apply(p, Options{}); err != nil {
		t.Fatal(problems(err))
	}
	repoText, _ := os.ReadFile(filepath.Join(sd.Values, "repo.toml"))
	toolsText, _ := os.ReadFile(sd.EditFile())
	if !strings.Contains(string(repoText), "[projects.my-repo]\nfolder = \".\"") || !strings.Contains(string(toolsText), "[tools.golang]\nprojects = [\"my-repo\"]") {
		t.Errorf("repo.toml %q, tools.toml %q", repoText, toolsText)
	}
	d, err = Detect(sd)
	if err != nil {
		t.Fatal(problems(err))
	}
	if len(d.Projects) != 0 || len(d.Links) != 0 {
		t.Errorf("second detect: projects %+v links %+v", d.Projects, d.Links)
	}
}

// TestDetectNoRootProject: in a monorepo, a tool of the categories
// languages and package-managers with no manifest in its evidence gets no
// link (D23).
func TestDetectNoRootProject(t *testing.T) {
	sd := projectFixture(t, monorepoTools(), []string{"golang", "nodejs"}, "", map[string]string{"go.mod": "", "web/package.json": "{}"})
	d, err := Detect(sd)
	if err != nil {
		t.Fatal(problems(err))
	}
	if !reflect.DeepEqual(d.Projects, []projects.Project{{Name: "web", Folder: "web"}}) ||
		!reflect.DeepEqual(d.Links, []Link{{Tool: "nodejs", Projects: []string{"web"}}}) {
		t.Errorf("projects %+v links %+v", d.Projects, d.Links)
	}
}

// TestPlanProjectsRestore: a failed sync puts repo.toml back.
func TestPlanProjectsRestore(t *testing.T) {
	sd := projectFixture(t, monorepoTools(), []string{"golang"}, "", nil)
	before, _ := os.ReadFile(filepath.Join(sd.Values, "repo.toml"))
	(&fakeSync{code: 1}).install(t, sd)
	p, err := Plan(sd, List{Projects: []projects.Project{{Name: "x", Folder: "."}}, Links: []Link{{Tool: "golang", Projects: []string{"x"}}}}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if _, putBack, _, err := Apply(p, Options{}); !putBack || err != nil {
		t.Fatalf("put back %v err %v", putBack, err)
	}
	if after, _ := os.ReadFile(filepath.Join(sd.Values, "repo.toml")); string(after) != string(before) {
		t.Errorf("repo.toml %q", after)
	}
}
