package repo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// checkRepo writes a repo with <repoToml> and <toolsToml> in its values
// folder, sets the environment of a check without a clone, and gives its
// root.
func checkRepo(t *testing.T, repoToml, toolsToml string) string {
	t.Helper()
	root, bin := t.TempDir(), t.TempDir()
	writeFile(t, values.RepoFile(values.FolderOf(root)), repoToml)
	writeFile(t, values.ToolsFile(values.FolderOf(root)), toolsToml)
	// lookPath needs a mise; the check does not run it.
	if err := os.WriteFile(filepath.Join(bin, "mise"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{
		"MISE_PROJECT_ROOT": root, "PWD": root, "BOOTSTRAP_VALUES": "", "BOOTSTRAP_CLONE": "",
		"BOOTSTRAP_VERBOSE": "", "NO_COLOR": "1", "PATH": bin + ":" + os.Getenv("PATH"),
	} {
		t.Setenv(k, v)
	}
	return root
}

// statuses gives the name and the status of each stage, as "name=status".
func statuses(stages []Stage) string {
	var out []string
	for _, s := range stages {
		out = append(out, s.Name+"="+s.Status)
	}
	return strings.Join(out, " ")
}

// stageOf gives the stage <name> of <stages>.
func stageOf(t *testing.T, stages []Stage, name string) Stage {
	t.Helper()
	for _, s := range stages {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("no stage %s in %s", name, statuses(stages))
	return Stage{}
}

// runCheck runs Check and Run --check on the same repo, and gives the
// stages, the lines outside a stage, the code and the lines of Check. The
// lines of the two runs must be the same.
func runCheck(t *testing.T) ([]Stage, []Finding, fail.Code, string) {
	t.Helper()
	var out, errs, plainOut, plainErrs bytes.Buffer
	stages, loose, code := Check(&out, &errs)
	plainCode := Run([]string{"--check"}, &plainOut, &plainErrs)
	if code != plainCode || out.String() != plainOut.String() || errs.String() != plainErrs.String() {
		t.Errorf("Check: code %d, stderr %q; Run: code %d, stderr %q", code, errs.String(), plainCode, plainErrs.String())
	}
	return stages, loose, code, errs.String()
}

func TestCheckPass(t *testing.T) {
	checkRepo(t, repoHead+"environments = [\"dev\"]\n", "format = 3\n")
	stages, loose, code, errText := runCheck(t)
	if code != fail.OK || len(loose) != 0 {
		t.Fatalf("code %d, loose %+v, stderr %q", code, loose, errText)
	}
	want := "envs=pass values=pass old-forms=pass env-rules=pass core-tools=pass overrides=pass members=pass releases=pass projects=pass scripts=pass project-links=pass"
	if got := statuses(stages); got != want {
		t.Errorf("stages %s, want %s", got, want)
	}
	// Without a clone, the scripts stage passes with one warning.
	f := stageOf(t, stages, "scripts").Findings
	if len(f) != 1 || f[0].Severity != SevWarning || f[0].Code != fail.OK || !strings.Contains(f[0].Message, "no clone of bootstrap") {
		t.Errorf("scripts findings %+v", f)
	}
	if f := stageOf(t, stages, "envs").Findings; f == nil || len(f) != 0 {
		t.Errorf("envs findings %#v, want an empty list", f)
	}
}

func TestCheckFailAndSkipped(t *testing.T) {
	checkRepo(t, repoHead+"environments = [\"prod\"]\n", "format = 3\n")
	stages, _, code, errText := runCheck(t)
	if code != fail.Usage {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	want := "envs=fail values=pass old-forms=pass env-rules=skipped core-tools=pass overrides=pass members=pass releases=pass projects=pass scripts=skipped project-links=skipped"
	if got := statuses(stages); got != want {
		t.Errorf("stages %s, want %s", got, want)
	}
	envs := stageOf(t, stages, "envs")
	if len(envs.Findings) == 0 || envs.Findings[0].Code != fail.Usage || envs.Findings[0].Severity != SevError {
		t.Errorf("envs findings %+v", envs.Findings)
	}
	for _, name := range []string{"env-rules", "scripts"} {
		if s := stageOf(t, stages, name); s.Status != StageSkipped || len(s.Findings) != 0 {
			t.Errorf("stage %+v, want skipped", s)
		}
	}
}

func TestCheckUnknownKey(t *testing.T) {
	checkRepo(t, repoHead+"unknown_key = 1\n", "format = 3\n")
	stages, loose, code, errText := runCheck(t)
	if code != fail.Usage || len(loose) != 0 {
		t.Fatalf("code %d, loose %+v, stderr %q", code, loose, errText)
	}
	want := "envs=pass values=fail old-forms=skipped env-rules=skipped core-tools=pass overrides=pass members=pass releases=pass projects=skipped scripts=skipped project-links=skipped"
	if got := statuses(stages); got != want {
		t.Errorf("stages %s, want %s", got, want)
	}
	f := stageOf(t, stages, "values").Findings
	if len(f) != 1 || f[0].Code != fail.Usage || f[0].Severity != SevError ||
		f[0].Path != ".config/bootstrap/repo.toml" || !strings.Contains(f[0].Message, "unknown_key") {
		t.Errorf("values findings %+v", f)
	}
}

func TestCheckEarlyFailure(t *testing.T) {
	root := checkRepo(t, repoHead, "format = 3\n")
	if err := os.Remove(values.ToolsFile(values.FolderOf(root))); err != nil {
		t.Fatal(err)
	}
	stages, loose, code, _ := runCheck(t)
	if code != fail.Usage || stages != nil || len(loose) != 1 || loose[0].Severity != SevError ||
		!strings.Contains(loose[0].Message, "tools.toml not found") {
		t.Errorf("code %d, stages %+v, loose %+v", code, stages, loose)
	}
}

func TestCheckCoreTools(t *testing.T) {
	checkRepo(t, repoHead+"environments = [\"dev\"]\ncore_tools = [\"x\"]\n", "format = 3\n")
	stages, loose, code, errText := runCheck(t)
	if len(loose) != 0 || !strings.Contains(errText, "core_tools in") {
		t.Fatalf("code %d, loose %+v, stderr %q", code, loose, errText)
	}
	s := stageOf(t, stages, "core-tools")
	if s.Status != StagePass || len(s.Findings) != 1 || s.Findings[0].Severity != SevWarning ||
		s.Findings[0].Code != fail.OK || s.Findings[0].Path != ".config/bootstrap/repo.toml" ||
		!strings.Contains(s.Findings[0].Message, "core_tools in") {
		t.Errorf("core-tools stage %+v", s)
	}
}

func TestCheckOwnTestTask(t *testing.T) {
	// D4 and D7: a test/_default that the manifest does not list gives code 2
	// in the stage old-forms, with the fix of init.
	root := checkRepo(t, repoHead+"environments = [\"dev\"]\n", "format = 3\n")
	writeFile(t, root+"/.config/mise/tasks/test/_default", "#!/bin/bash\n")
	stages, _, code, errText := runCheck(t)
	if code != fail.Usage {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	s := stageOf(t, stages, "old-forms")
	if s.Status != StageFail || len(s.Findings) != 1 || s.Findings[0].Code != fail.Usage ||
		!strings.Contains(s.Findings[0].Message, ".config/mise/tasks/test/_default is a task of this repo") ||
		!strings.Contains(s.Findings[0].Message, "run bootstrap init") {
		t.Errorf("old-forms stage %+v", s)
	}
}

func TestCheckUnnamedPart(t *testing.T) {
	// D15: an executable part without an NN- name is a warning, code 0.
	root := checkRepo(t, repoHead+"environments = [\"dev\"]\n", "format = 3\n")
	p := root + "/.config/mise/tasks/code/lint/mylint"
	writeFile(t, p, "#!/bin/bash\n")
	if err := os.Chmod(p, 0o755); err != nil {
		t.Fatal(err)
	}
	stages, _, code, errText := runCheck(t)
	if code != fail.OK {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	s := stageOf(t, stages, "old-forms")
	if s.Status != StagePass || len(s.Findings) != 1 || s.Findings[0].Severity != SevWarning || s.Findings[0].Code != fail.OK ||
		!strings.Contains(s.Findings[0].Message, "code/lint/mylint has no NN- name") ||
		!strings.Contains(s.Findings[0].Message, "rename it to NN-<name>") {
		t.Errorf("old-forms stage %+v", s)
	}
}

func TestCheckOldWorktreeCmd(t *testing.T) {
	// D20: the old command of the worktree setup in .config/wt.toml and in a
	// member's worktree_install gives code 2 in the stage old-forms, with the
	// fix of init; after the rename of init, the check passes.
	repo := repoHead + "environments = [\"dev\"]\n\n[members.api]\ntype = \"folder\"\nworktree_install = \"mise run setup:worktree\"\n"
	root := checkRepo(t, repo, "format = 3\n")
	writeFile(t, root+"/"+oldform.WtFile, "[[pre-start]]\nsetup = \"mise run setup:worktree\"\n")
	writeFile(t, root+"/api/x", "")
	stages, _, code, errText := runCheck(t)
	if code != fail.Usage {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	s := stageOf(t, stages, "old-forms")
	if s.Status != StageFail || len(s.Findings) != 2 ||
		!strings.Contains(s.Findings[0].Message, "/.config/wt.toml runs mise run setup:worktree") ||
		!strings.Contains(s.Findings[1].Message, "members.api.worktree_install in ") ||
		!strings.Contains(s.Findings[1].Message, "run bootstrap init") {
		t.Errorf("old-forms stage %+v", s)
	}
	if !strings.Contains(errText, "Run bootstrap init") {
		t.Errorf("stderr %q", errText)
	}
	writeFile(t, values.RepoFile(values.FolderOf(root)), oldform.RenameCmd(repo))
	writeFile(t, root+"/"+oldform.WtFile, "[[pre-start]]\nsetup = \""+oldform.NewWorktreeCmd+"\"\n")
	if _, _, code, errText := runCheck(t); code != fail.OK {
		t.Fatalf("after the rename: code %d, stderr %q", code, errText)
	}
}
