package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// writeRepo is a repository for the write commands with --json: the tools a
// and c with a detect script that matches, the tool q with a question, and
// the children r1 and r2 of p, where r2 can replace r1. a, p and r1 are
// selected. It gives the values file, its text, and the runner of a command.
func writeRepo(t *testing.T) (string, string, func(args ...string) (fail.Code, string, string)) {
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
	writeFile(t, filepath.Join(clone, "repo", "tools", "q", "tools.toml"), keys+"\n[questions.name]\nquestion = \"N?\"\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "p", "tools.toml"), keys, 0o644)
	for _, tool := range []string{"r1", "r2"} {
		writeFile(t, filepath.Join(clone, "repo", "tools", tool, "tools.toml"), keys+"parent = \"p\"\n", 0o644)
	}
	writeFile(t, filepath.Join(clone, "repo", "tools", "r2", "scripts", "migrate-r1"), "#!/bin/bash\nexit 0\n", 0o755)
	dir := filepath.Join(root, ".config", "bootstrap")
	writeFile(t, filepath.Join(dir, "repo.toml"), "format = 3\nrepo_name = \"x\"\ncore_branches = [ \"main\" ]\n", 0o644)
	values := filepath.Join(dir, "tools.toml")
	const text = "format = 3\n\n[tools.a]\n\n[tools.p]\n\n[tools.r1]\n"
	writeFile(t, values, text, 0o644)
	for _, dir := range []string{clone, root} {
		for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--no-gpg-sign", "-m", "c"}} {
			if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
				t.Fatalf("git %v: %s", args, out)
			}
		}
	}
	t.Setenv("MISE_PROJECT_ROOT", root)
	t.Setenv("BOOTSTRAP_VALUES", "")
	t.Chdir(root)
	t.Setenv("PWD", root)
	return values, text, func(args ...string) (fail.Code, string, string) {
		var stdout, stderr bytes.Buffer
		code := Run("dev", append(args, "--clone", clone), strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
}

// writeJSON is the data of a write command, as a test reads it.
type writeJSON struct {
	Plan *struct {
		Changes []struct {
			Path    string `json:"path"`
			Content string `json:"content"`
			Remove  bool   `json:"remove"`
			Note    string `json:"note"`
		} `json:"changes"`
		Files []struct {
			Path string `json:"path"`
			Kind string `json:"kind"`
		} `json:"files"`
		Skipped []json.RawMessage `json:"skipped"`
		Pending []struct {
			Tool string `json:"tool"`
			Key  string `json:"key"`
			Text string `json:"text"`
		} `json:"pending"`
		Notes []string `json:"notes"`
	} `json:"plan"`
	Outcomes []struct {
		Kind    string `json:"kind"`
		Path    string `json:"path"`
		Message string `json:"message"`
	} `json:"outcomes"`
}

// readWrite parses the data of <e>, and fails when a list is null.
func readWrite(t *testing.T, e envelope) writeJSON {
	t.Helper()
	var d writeJSON
	if err := json.Unmarshal(e.Data, &d); err != nil {
		t.Fatalf("data %s: %v", e.Data, err)
	}
	for _, list := range []string{`"changes":null`, `"files":null`, `"skipped":null`, `"pending":null`, `"notes":null`, `"outcomes":null`} {
		if strings.Contains(string(e.Data), list) {
			t.Errorf("data %s has %s", e.Data, list)
		}
	}
	if d.Plan == nil || d.Outcomes == nil {
		t.Fatalf("data %s, want plan and outcomes", e.Data)
	}
	return d
}

// TestWriteJSONNeedsConsent: with --json, init, add, remove, replace and
// detect --add need --yes or --dry-run: code 2, and nothing written (D4).
func TestWriteJSONNeedsConsent(t *testing.T) {
	values, text, run := writeRepo(t)
	for _, args := range [][]string{
		{"migrate"}, {"add", "tool", "c"}, {"remove", "tool", "r1"}, {"replace", "tool", "r1", "r2"}, {"detect", "--add"},
	} {
		code, out, errText := run(append(args, "--json")...)
		e := readEnvelope(t, out)
		if code != fail.Usage || e.Command != args[0] || string(e.Data) != "null" || len(e.Errors) != 1 ||
			!strings.Contains(e.Errors[0].Message, args[0]+" --json needs --yes or --dry-run") {
			t.Errorf("%v: code %d, envelope %+v, stderr %q", args, code, e, errText)
		}
		if b, _ := os.ReadFile(values); string(b) != text {
			t.Errorf("%v wrote the values file: %q", args, b)
		}
	}
}

// TestWriteJSONDryRun: --json --dry-run gives the plan and no outcome, and
// writes nothing; stdout holds the envelope only.
func TestWriteJSONDryRun(t *testing.T) {
	values, text, run := writeRepo(t)
	for _, c := range []struct {
		args    []string
		changes int
	}{
		{[]string{"migrate"}, 0},
		{[]string{"sync"}, 0},
		{[]string{"add", "tool", "c"}, 1},
		{[]string{"remove", "tool", "r1"}, 1},
		{[]string{"replace", "tool", "r1", "r2"}, 1},
		// The tool c, and the root project of a repo with no project.
		{[]string{"detect", "--add"}, 2},
	} {
		code, out, errText := run(append(c.args, "--json", "--dry-run")...)
		e := readEnvelope(t, out)
		if code != fail.OK || e.Code != fail.OK {
			t.Errorf("%v: code %d, envelope %+v, stderr %q", c.args, code, e, errText)
			continue
		}
		d := readWrite(t, e)
		if len(d.Outcomes) != 0 || len(d.Plan.Changes) != c.changes {
			t.Errorf("%v: data %s", c.args, e.Data)
		}
		if b, _ := os.ReadFile(values); string(b) != text {
			t.Errorf("%v wrote the values file: %q", c.args, b)
		}
	}
	// The change of add is the values file, with its text.
	_, out, _ := run("add", "tool", "c", "--json", "--dry-run")
	d := readWrite(t, readEnvelope(t, out))
	if ch := d.Plan.Changes[0]; ch.Path != values || !strings.Contains(ch.Content, "[tools.c]") || ch.Remove {
		t.Errorf("change %+v", ch)
	}
}

// TestWriteJSONApply: an apply gives the plan and one outcome for each target
// of the sync, and writes the change.
func TestWriteJSONApply(t *testing.T) {
	values, _, run := writeRepo(t)
	code, out, errText := run("add", "tool", "c", "--json", "--yes")
	e := readEnvelope(t, out)
	if code != fail.OK || e.Command != "add" {
		t.Fatalf("code %d, envelope %+v, stderr %q", code, e, errText)
	}
	d := readWrite(t, e)
	if len(d.Plan.Changes) != 1 || len(d.Outcomes) == 0 {
		t.Errorf("data %s", e.Data)
	}
	for _, o := range d.Outcomes {
		if o.Kind == "" || o.Path == "" {
			t.Errorf("outcome %+v", o)
		}
	}
	if b, _ := os.ReadFile(values); !strings.Contains(string(b), "[tools.c]") {
		t.Errorf("values file %q", b)
	}
	// The human lines stay on stderr, and stdout without --json keeps them.
	if !strings.Contains(errText, "Changes") {
		t.Errorf("stderr %q", errText)
	}
	if code, out, _ := run("remove", "tool", "r1", "--dry-run"); code != fail.OK || !strings.Contains(out, "Changes") {
		t.Errorf("without --json: code %d, stdout %q", code, out)
	}
}

// TestWriteJSONApplyEach: each of the six commands applies with --json (sync
// with no --yes, D21) and gives one outcome for each target of its sync.
func TestWriteJSONApplyEach(t *testing.T) {
	for _, args := range [][]string{
		{"migrate", "--yes"}, {"sync"}, {"add", "tool", "c", "--yes"}, {"remove", "tool", "r1", "--yes"},
		{"replace", "tool", "r1", "r2", "--yes"}, {"detect", "--add", "--yes"},
	} {
		t.Run(args[0], func(t *testing.T) {
			_, _, run := writeRepo(t)
			code, out, errText := run(append(args, "--json")...)
			e := readEnvelope(t, out)
			if code != fail.OK || e.Command != args[0] {
				t.Fatalf("code %d, envelope %+v, stderr %q", code, e, errText)
			}
			if d := readWrite(t, e); len(d.Outcomes) == 0 {
				t.Errorf("data %s", e.Data)
			}
		})
	}
}

// TestSyncJSONPlanFiles: a real sync --json in a clean repository still runs
// the plan, and data.plan.files lists the files that the run wrote (D50).
func TestSyncJSONPlanFiles(t *testing.T) {
	_, _, run := writeRepo(t)
	code, out, errText := run("sync", "--json")
	e := readEnvelope(t, out)
	if code != fail.OK || e.Command != "sync" {
		t.Fatalf("code %d, envelope %+v, stderr %q", code, e, errText)
	}
	d := readWrite(t, e)
	if len(d.Plan.Files) == 0 || len(d.Outcomes) == 0 {
		t.Fatalf("data %s", e.Data)
	}
	for _, f := range d.Plan.Files {
		if f.Path == "" || f.Kind == "" {
			t.Errorf("file %+v", f)
		}
	}
}

// TestWriteJSONPending: a question with no answer stops --json --yes with
// code 2 and writes nothing (D22); --json --dry-run exits 0 with a warning
// (D45). Both list the question in data.plan.pending.
func TestWriteJSONPending(t *testing.T) {
	values, text, run := writeRepo(t)
	code, out, errText := run("add", "tool", "q", "--json", "--yes")
	e := readEnvelope(t, out)
	d := readWrite(t, e)
	if code != fail.Usage || len(e.Errors) != 1 || !strings.Contains(e.Errors[0].Message, "q.name") ||
		len(d.Plan.Pending) != 1 || d.Plan.Pending[0].Key != "name" || d.Plan.Pending[0].Text != "N?" {
		t.Errorf("--yes: code %d, envelope %+v, data %s, stderr %q", code, e, e.Data, errText)
	}
	if b, _ := os.ReadFile(values); string(b) != text {
		t.Errorf("--yes wrote the values file: %q", b)
	}
	code, out, errText = run("add", "tool", "q", "--json", "--dry-run")
	e = readEnvelope(t, out)
	d = readWrite(t, e)
	if code != fail.OK || len(e.Errors) != 0 || len(e.Warnings) != 1 || !strings.Contains(e.Warnings[0].Message, "q.name") ||
		len(d.Plan.Pending) != 1 {
		t.Errorf("--dry-run: code %d, envelope %+v, data %s, stderr %q", code, e, e.Data, errText)
	}
	// With the answer, the apply goes on.
	code, out, errText = run("add", "tool", "q", "--json", "--yes", "--answer", "q.name=v")
	if e := readEnvelope(t, out); code != fail.OK || len(readWrite(t, e).Plan.Pending) != 0 {
		t.Errorf("--answer: code %d, envelope %+v, stderr %q", code, e, errText)
	}
}

// TestDetectJSON: a plain detect --json is a read: data.detected has one
// entry for each tool that detect reports (D30).
func TestDetectJSON(t *testing.T) {
	_, _, run := writeRepo(t)
	code, out, errText := run("detect", "--json")
	e := readEnvelope(t, out)
	var d struct {
		Detected []detectedData               `json:"detected"`
		Projects map[string]map[string]string `json:"projects"`
		Links    map[string][]string          `json:"links"`
	}
	if err := json.Unmarshal(e.Data, &d); err != nil {
		t.Fatalf("data %s: %v", e.Data, err)
	}
	// A repo with no project has the root as one project, named repo_name;
	// a tool of the category core links to no project (D8, D9, D25).
	if len(d.Projects) != 1 || d.Projects["x"]["folder"] != "." || d.Links == nil || len(d.Links) != 0 {
		t.Errorf("projects %v, links %v", d.Projects, d.Links)
	}
	if code != fail.Other || e.Command != "detect" || len(d.Detected) != 2 || len(e.Errors) != 1 {
		t.Fatalf("code %d, envelope %+v, data %s, stderr %q", code, e, e.Data, errText)
	}
	a, c := d.Detected[0], d.Detected[1]
	if a.Tool != "a" || !a.Selected || a.Result != "match" || c.Tool != "c" || c.Selected || !c.Offered ||
		len(c.Evidence) != 1 || c.Evidence[0] != "g" {
		t.Errorf("detected %+v", d.Detected)
	}
}
