package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/contract"
	"github.com/virajp/bootstrap/cli/internal/fail"
)

// envelope is the envelope of a --json command, as a test reads it.
type envelope struct {
	Contract int             `json:"contract"`
	Command  string          `json:"command"`
	OK       bool            `json:"ok"`
	Code     fail.Code       `json:"code"`
	Data     json.RawMessage `json:"data"`
	Errors   []contract.Item `json:"errors"`
	Warnings []contract.Item `json:"warnings"`
}

// readEnvelope parses <out>: one line of JSON, then a newline.
func readEnvelope(t *testing.T, out string) envelope {
	t.Helper()
	if strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
		t.Fatalf("stdout = %q, want one line", out)
	}
	var e envelope
	if err := json.Unmarshal([]byte(out), &e); err != nil {
		t.Fatalf("stdout = %q: %v", out, err)
	}
	if e.Contract != contract.Version || e.OK != (e.Code == fail.OK) {
		t.Errorf("envelope = %+v", e)
	}
	return e
}

func runArgs(args ...string) (fail.Code, string, string) {
	var stdout, stderr bytes.Buffer
	code := Run("dev", args, strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestContractJSON(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	code, out, errText := runArgs("contract", "--json")
	if code != fail.OK || errText != "" {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	e := readEnvelope(t, out)
	if e.Command != "contract" || e.Code != fail.OK {
		t.Errorf("envelope = %+v", e)
	}
	want := `{"versions":[1],"cli_version":"dev","commands":["contract","check","questions","config get","config set","config unset","config show","mac questions"],` +
		`"ports":["code:format","code:lint","code:sec","code:precommit","git:merge:<branch>","git:worktree:create","git:worktree:setup","git:worktree:delete","code:lsp:check","setup","test","test:e2e","test:golden","p:<id>:deploy"]}`
	if string(e.Data) != want {
		t.Errorf("data = %s, want %s", e.Data, want)
	}
}

func TestContractText(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	code, out, _ := runArgs("contract")
	if code != fail.OK || !strings.HasPrefix(out, "versions  1\nversion   dev\ncommands  contract, check, questions,") {
		t.Errorf("code %d, stdout %q", code, out)
	}
}

func TestJSONUsageErrors(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	cases := []struct {
		args    []string
		command string
		message string
	}{
		{[]string{"--json"}, "", "--json is not a command"},
		{[]string{"repo", "--json"}, "", "repo is not a command"},
		{[]string{"list", "--json"}, "list", "list has no flag --json"},
		{[]string{"tui", "--json"}, "tui", "tui has no flag --json"},
		{[]string{"version", "--json"}, "version", "version has no flag --json"},
		{[]string{"contract", "--json", "--yes"}, "contract", "contract has no flag --yes"},
		{[]string{"contract", "x", "--json"}, "contract", "contract takes no argument x"},
		{[]string{"mac", "add", "tool", "x", "--json"}, "mac add", "mac add has no flag --json"},
		{[]string{"mac", "check", "--json"}, "mac check", "mac check has no flag --json"},
		{[]string{"mac", "contract", "--json"}, "", "mac contract is not a command"},
		{[]string{"mac", "describe", "--json"}, "", "mac describe is not a command"},
		{[]string{"mac", "config", "get", "x", "--json"}, "", "mac config is not a command"},
		{[]string{"config", "--json"}, "config", "config needs get, set, unset or show"},
		{[]string{"config", "get", "--json"}, "config get", "config get takes the wrong number of arguments"},
		{[]string{"config", "get", "repo.repo_name", "--json", "--format", "toml"}, "config get", "config get takes --json or --format, not both"},
		{[]string{"config", "show", "--format=json", "--json"}, "config show", "config show takes --json or --format, not both"},
	}
	for _, c := range cases {
		code, out, errText := runArgs(c.args...)
		if code != fail.Usage {
			t.Errorf("%v: code %d, want 2", c.args, code)
			continue
		}
		e := readEnvelope(t, out)
		if e.Code != fail.Usage || e.Command != c.command || string(e.Data) != "null" || len(e.Errors) != 1 ||
			e.Errors[0].Code != fail.Usage || !strings.HasPrefix(e.Errors[0].Message, c.message) {
			t.Errorf("%v: envelope %+v, want command %q and the error %q", c.args, e, c.command, c.message)
		}
		if !strings.Contains(errText, c.message) {
			t.Errorf("%v: stderr %q, want the human line", c.args, errText)
		}
	}
}

func TestJSONHelpOnStderr(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	code, out, errText := runArgs("contract", "--json", "-h")
	if code != fail.OK || !strings.HasPrefix(errText, "usage:") {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	if e := readEnvelope(t, out); e.Command != "contract" || string(e.Data) != "null" {
		t.Errorf("envelope = %+v", e)
	}
}

func TestConfigJSON(t *testing.T) {
	_, run := valuesRepo(t)
	code, out, errText := run("config", "get", "repo.repo_name", "--json")
	if code != fail.OK {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	e := readEnvelope(t, out)
	if e.Command != "config get" || string(e.Data) != `{"key":"repo.repo_name","value":"x","source":"file"}` {
		t.Errorf("envelope = %+v, data %s", e, e.Data)
	}
	code, out, _ = run("config", "show", "tools", "--json")
	if e := readEnvelope(t, out); code != fail.OK || e.Command != "config show" || string(e.Data) != `{"format":3,"tools":{"a":{"flag":false}}}` {
		t.Errorf("code %d, envelope %+v, data %s", code, e, e.Data)
	}
	// An error of the command: data null, the error with its code.
	code, out, _ = run("config", "get", "repo.nope", "--json")
	e = readEnvelope(t, out)
	if code != fail.Usage || string(e.Data) != "null" || len(e.Errors) != 1 || e.Errors[0].Code != fail.Usage ||
		!strings.Contains(e.Errors[0].Message, "is not a key of the values") {
		t.Errorf("code %d, envelope %+v", code, e)
	}
	// Without --json, the output stays.
	code, out, _ = run("config", "get", "repo.repo_name")
	if code != fail.OK || out != `{"key":"repo.repo_name","value":"x","source":"file"}`+"\n" {
		t.Errorf("code %d, stdout %q", code, out)
	}
}

func TestConfigSetUnsetJSON(t *testing.T) {
	_, run := valuesRepo(t)
	cases := []struct {
		args []string
		data string
	}{
		{[]string{"config", "set", "repo.product_name", "P"}, `{"key":"repo.product_name","value":"P"}`},
		{[]string{"config", "set", "repo.core_branches", `[ "main", "develop" ]`}, `{"key":"repo.core_branches","value":["main","develop"]}`},
		{[]string{"config", "set", "tools.a.flag", "true"}, `{"key":"tools.a.flag","value":true}`},
		{[]string{"config", "set", "releases.web", "MAJOR"}, `{"key":"releases.web","value":"MAJOR"}`},
		// The file keeps the higher level, so data holds it.
		{[]string{"config", "set", "releases.web", "MINOR"}, `{"key":"releases.web","value":"MAJOR"}`},
		{[]string{"config", "unset", "repo.product_name"}, `{"key":"repo.product_name"}`},
		// A key that is not set: the same data.
		{[]string{"config", "unset", "repo.product_name"}, `{"key":"repo.product_name"}`},
	}
	for _, c := range cases {
		code, out, errText := run(append(c.args, "--json")...)
		e := readEnvelope(t, out)
		if code != fail.OK || e.Command != c.args[0]+" "+c.args[1] || string(e.Data) != c.data || len(e.Warnings) != 0 {
			t.Errorf("%v: code %d, envelope %+v, data %s, stderr %q", c.args, code, e, e.Data, errText)
		}
	}
	// The info lines stay on stderr, out of the warnings (D48).
	code, out, errText := run("config", "set", "releases.web", "PATCH", "--json")
	if e := readEnvelope(t, out); code != fail.OK || len(e.Warnings) != 0 || !strings.Contains(errText, "releases.web stays MAJOR") {
		t.Errorf("code %d, envelope %+v, stderr %q", code, e, errText)
	}
	// A failure: data null.
	code, out, _ = run("config", "set", "repo.core_branches", "main", "--json")
	if e := readEnvelope(t, out); code != fail.Usage || string(e.Data) != "null" || len(e.Errors) != 1 {
		t.Errorf("code %d, envelope %+v", code, e)
	}
	code, out, _ = run("config", "unset", "repo.repo_name", "--json")
	if e := readEnvelope(t, out); code != fail.Usage || string(e.Data) != "null" || len(e.Errors) != 1 {
		t.Errorf("code %d, envelope %+v", code, e)
	}
	// Without --json, stdout stays empty.
	if code, out, _ := run("config", "set", "repo.product_name", "Q"); code != fail.OK || out != "" {
		t.Errorf("code %d, stdout %q", code, out)
	}
}

func TestQuestionsRunJSON(t *testing.T) {
	_, run := valuesRepo(t)
	code, out, errText := run("questions", "--json")
	if code != fail.OK {
		t.Fatalf("code %d, stderr %q", code, errText)
	}
	e := readEnvelope(t, out)
	var rows []map[string]any
	if err := json.Unmarshal(e.Data, &rows); err != nil || e.Command != "questions" || len(rows) != 1 || rows[0]["key"] != "name" {
		t.Errorf("envelope %+v, data %s (%v)", e, e.Data, err)
	}
	code, out, _ = run("questions", "nope", "--json")
	if e := readEnvelope(t, out); code != fail.Missing || e.Code != fail.Missing || len(e.Errors) != 1 || e.Errors[0].Code != fail.Missing {
		t.Errorf("code %d, envelope %+v", code, e)
	}
}

func TestCheckJSON(t *testing.T) {
	dir, run := valuesRepo(t)
	code, out, errText := run("check", "--json")
	e := readEnvelope(t, out)
	var data struct {
		Valid  bool `json:"valid"`
		Checks []struct {
			Name     string `json:"name"`
			Status   string `json:"status"`
			Findings []struct {
				Code     fail.Code `json:"code"`
				Severity string    `json:"severity"`
				Message  string    `json:"message"`
				Path     string    `json:"path"`
			} `json:"findings"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(e.Data, &data); err != nil {
		t.Fatalf("data %s: %v", e.Data, err)
	}
	if code != e.Code || e.Command != "check" || data.Valid != (code == fail.OK) || len(data.Checks) != 9 || data.Checks[0].Name != "envs" {
		t.Errorf("code %d, envelope %+v, data %s, stderr %q", code, e, e.Data, errText)
	}
	// An unknown key: the values stage fails with one finding of the file.
	writeFile(t, filepath.Join(dir, "repo.toml"), "format = 3\nrepo_name = \"x\"\ncore_branches = [ \"main\" ]\nunknown_key = 1\n", 0o644)
	code, out, errText = run("check", "--json")
	e = readEnvelope(t, out)
	data.Checks = nil
	if err := json.Unmarshal(e.Data, &data); err != nil {
		t.Fatalf("data %s: %v", e.Data, err)
	}
	if code != fail.Usage || e.Code != fail.Usage || data.Valid || len(data.Checks) != 9 {
		t.Fatalf("code %d, envelope %+v, data %s, stderr %q", code, e, e.Data, errText)
	}
	v := data.Checks[1]
	if v.Name != "values" || v.Status != "fail" || len(v.Findings) != 1 || v.Findings[0].Path != ".config/bootstrap/repo.toml" ||
		!strings.Contains(v.Findings[0].Message, "unknown_key") || v.Findings[0].Code != fail.Usage {
		t.Errorf("values stage %+v", v)
	}
	for _, i := range []int{2, 3, 8} {
		if data.Checks[i].Status != "skipped" {
			t.Errorf("stage %+v, want skipped", data.Checks[i])
		}
	}
	// Without --json, stdout stays empty and the lines stay on stderr.
	code, out, plain := run("check")
	if code != fail.Usage || out != "" || !strings.Contains(plain, "unknown_key") {
		t.Errorf("code %d, stdout %q, stderr %q", code, out, plain)
	}
}

func TestCheckJSONEarlyFailure(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	root := t.TempDir()
	t.Setenv("MISE_PROJECT_ROOT", root)
	t.Setenv("PWD", root)
	t.Chdir(root)
	code, out, _ := runArgs("check", "--json")
	e := readEnvelope(t, out)
	if code == fail.OK || e.Command != "check" || string(e.Data) != `{"valid":false,"checks":[]}` || len(e.Errors) == 0 {
		t.Errorf("code %d, envelope %+v, data %s", code, e, e.Data)
	}
}
