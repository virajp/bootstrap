package repo

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// TestCheckProjectFolders: a listed project whose folder is not there is a
// finding with code 2 (D10).
func TestCheckProjectFolders(t *testing.T) {
	root := checkRepo(t, repoHead+"environments = [\"dev\"]\n[projects.web]\nfolder = \"web\"\n[projects.api]\nfolder = \"api\"\n", "format = 3\n")
	if err := os.MkdirAll(filepath.Join(root, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	stages, _, code, errText := runCheck(t)
	st := stageOf(t, stages, "projects")
	if code != fail.Usage || st.Status != StageFail || len(st.Findings) != 1 ||
		!strings.Contains(st.Findings[0].Message, "[projects.web].folder = web in") || st.Findings[0].Code != fail.Usage {
		t.Errorf("code %d, stage %+v, stderr %q", code, st, errText)
	}
	if s := stageOf(t, stages, "project-links"); s.Status != StageSkipped {
		t.Errorf("project-links %+v", s)
	}
}

// TestCheckProjectNames: a projects list with a name that repo.toml does
// not have stops the values stage with code 2 (D7).
func TestCheckProjectNames(t *testing.T) {
	checkRepo(t, repoHead+"environments = [\"dev\"]\n", "format = 3\n[tools.a]\nprojects = [\"web\"]\n")
	stages, _, code, errText := runCheck(t)
	st := stageOf(t, stages, "values")
	if code != fail.Usage || st.Status != StageFail || !strings.Contains(errText, "names web, which is not a [projects.<name>] table of repo.toml") {
		t.Errorf("code %d, stage %+v, stderr %q", code, st, errText)
	}
}

// TestProjectLinks: a link passes when the detect script matches in its
// folder, or for a tool of the categories languages and package-managers
// when it matches at the root by a manifest that names the folder (D22,
// D31). Else it is a warning. A root_only tool and a missing folder are
// skipped (D21, D26).
func TestProjectLinks(t *testing.T) {
	clone, root := t.TempDir(), t.TempDir()
	files := map[string]string{
		filepath.Join(clone, "repo/categories.toml"): "[category.core]\ntitle = \"Core\"\n[category.package-managers]\ntitle = \"P\"\n[category.languages]\ntitle = \"L\"\n",
	}
	tool := func(name, cat, rootOnly, file string) {
		files[filepath.Join(clone, "repo/tools", name, "tools.toml")] = "summary = \"A\"\ncategory = \"" + cat + "\"\ndefault_envs = [ \"dev\" ]\ndepends = []\nroot_only = " + rootOnly + "\n"
		files[filepath.Join(clone, "repo/tools", name, "scripts/detect")] = "#!/bin/bash\ncat >/dev/null\nif [ -f " + file + " ]; then echo '{\"result\":\"match\",\"evidence\":[\"" + file + "\"]}'; else echo '{\"result\":\"none\",\"evidence\":[]}'; fi\n"
	}
	tool("lang", "languages", "false", "lang.txt")
	tool("pm", "package-managers", "false", "pnpm-workspace.yaml")
	tool("other", "core", "false", "pnpm-workspace.yaml")
	tool("ro", "core", "true", "nothing")
	for p, text := range map[string]string{
		"pnpm-workspace.yaml": "packages: [\"web\"]\n",
		"web/package.json":    "{}",
		"api/lang.txt":        "",
		"old/x.txt":           "",
	} {
		files[filepath.Join(root, p)] = text
	}
	files[values.RepoFile(values.FolderOf(root))] = repoHead + "[projects.web]\nfolder = \"web\"\n[projects.api]\nfolder = \"api\"\n[projects.old]\nfolder = \"old\"\n[projects.gone]\nfolder = \"gone\"\n"
	files[values.ToolsFile(values.FolderOf(root))] = "format = 3\n[tools.lang]\nprojects = [\"api\", \"old\", \"gone\"]\n[tools.pm]\nprojects = [\"web\"]\n[tools.other]\nprojects = [\"web\"]\n[tools.ro]\nprojects = [\"web\"]\n"
	for p, text := range files {
		writeFile(t, p, text)
		if strings.HasSuffix(p, "/detect") {
			_ = os.Chmod(p, 0o755)
		}
	}
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	t.Setenv("NO_COLOR", "1")
	var errs bytes.Buffer
	s := newState(io.Discard, &errs)
	s.root = root
	s.values = values.FolderOf(root)
	s.showTools = "tools.toml"
	s.src = source.Source{Root: clone, Side: source.Repo}
	idx, err := s.src.Open()
	if err != nil {
		t.Fatal(err)
	}
	s.idx = idx
	if s.vals, err = values.ReadRepo(s.values, "values"); err != nil {
		t.Fatal(err)
	}
	if code := s.projectLinks(); code != fail.OK {
		t.Fatalf("code %d, stderr %q", code, errs.String())
	}
	// lang matches in api; pm passes by its manifest; other is not of the
	// categories of a root link; ro is root_only; gone is not there.
	got := errs.String()
	for _, want := range []string{"lists old, but the detect script of lang gives no match in old", "[tools.other].projects in tools.toml lists web"} {
		if !strings.Contains(got, want) {
			t.Errorf("no %q in %q", want, got)
		}
	}
	if n := strings.Count(got, "[WARN]"); n != 2 {
		t.Errorf("%d warnings: %q", n, got)
	}
}
