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

// lspRepo is a clone with three tools, each with a language server: a pins
// pa with requires_found = ["pyproject.toml"] and da in [mise_env.dev.tools],
// b has a host server, and c is selected in ci only. It gives the root of
// the repository and the run function.
func lspRepo(t *testing.T) (string, func(args ...string) (fail.Code, string, string)) {
	t.Helper()
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	clone, root := t.TempDir(), t.TempDir()
	const keys = catalogKeys + "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"
	writeFile(t, filepath.Join(clone, "cli", "VERSION"), "0.0.1\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "categories.toml"), catalogFile, 0o644)
	writeFile(t, filepath.Join(clone, "repo", "packs.toml"), "[pack.base]\na = { required = true }\n", 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "a", "tools.toml"), keys+`
[mise.tools.pa]
requires_found = [ "pyproject.toml" ]
version = "latest"

[mise_env.dev.tools.da]
version = "latest"

[lsp.pa-lsp]
command = "pa"
args = [ "lsp" ]
extensions = { ".py" = "python", ".pyi" = "python" }
source = "pa"

[lsp.da-lsp]
command = "da"
args = [ "--stdio" ]
extensions = { ".d" = "d" }
startup_timeout = 1000
source = "da"
`, 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "b", "tools.toml"), keys+`
[lsp.host-lsp]
command = "hl"
extensions = { ".h" = "h" }
source = "host"
`, 0o644)
	writeFile(t, filepath.Join(clone, "repo", "tools", "c", "tools.toml"), keys+`
[lsp.ci-lsp]
command = "cl"
extensions = { ".c" = "c" }
source = "host"
`, 0o644)
	dir := filepath.Join(root, ".config", "bootstrap")
	writeFile(t, filepath.Join(dir, "repo.toml"), "format = 3\nrepo_name = \"x\"\ncore_branches = [ \"main\" ]\nenvironments = [ \"dev\", \"ci\" ]\n", 0o644)
	writeFile(t, filepath.Join(dir, "tools.toml"), "format = 3\n\n[tools.a]\n\n[tools.b]\n\n[tools.c]\nenv = [ \"ci\" ]\n", 0o644)
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	t.Setenv("MISE_PROJECT_ROOT", root)
	t.Setenv("BOOTSTRAP_VALUES", "")
	t.Setenv("PWD", root)
	return root, func(args ...string) (fail.Code, string, string) {
		var stdout, stderr bytes.Buffer
		code := Run("dev", append(args, "--clone", clone), strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
}

// lspJSON reads the lsp data of an envelope.
func lspJSON(t *testing.T, raw json.RawMessage) (map[string]json.RawMessage, map[string]lspCheck) {
	t.Helper()
	var d struct {
		Servers map[string]json.RawMessage `json:"servers"`
		Checks  map[string]lspCheck        `json:"checks"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatalf("data %s: %v", raw, err)
	}
	return d.Servers, d.Checks
}

// TestLSPJSON: the servers of the tools in dev, as Claude entries launched
// with mise x --, with startupTimeout only when it is set (D4); a server of a
// pin with requires_found is there only with the file (D28); checks gives the
// command and the source (D24).
func TestLSPJSON(t *testing.T) {
	root, run := lspRepo(t)
	code, out, errText := run("lsp", "--json")
	e := readEnvelope(t, out)
	if code != fail.OK || e.Command != "lsp" {
		t.Fatalf("code %d, envelope %+v, stderr %q", code, e, errText)
	}
	servers, checks := lspJSON(t, e.Data)
	want := map[string]string{
		"da-lsp":   `{"command":"mise","args":["x","--","da","--stdio"],"extensionToLanguage":{".d":"d"},"startupTimeout":1000}`,
		"host-lsp": `{"command":"mise","args":["x","--","hl"],"extensionToLanguage":{".h":"h"}}`,
	}
	if len(servers) != len(want) {
		t.Errorf("servers %s", e.Data)
	}
	for n, w := range want {
		if string(servers[n]) != w {
			t.Errorf("server %s = %s, want %s", n, servers[n], w)
		}
	}
	if len(checks) != 2 || checks["da-lsp"] != (lspCheck{"da", "da"}) || checks["host-lsp"] != (lspCheck{"hl", "host"}) {
		t.Errorf("checks %+v", checks)
	}
	writeFile(t, filepath.Join(root, "sub", "pyproject.toml"), "", 0o644)
	code, out, _ = run("lsp", "--json")
	servers, checks = lspJSON(t, readEnvelope(t, out).Data)
	if code != fail.OK || string(servers["pa-lsp"]) != `{"command":"mise","args":["x","--","pa","lsp"],"extensionToLanguage":{".py":"python",".pyi":"python"}}` ||
		checks["pa-lsp"] != (lspCheck{"pa", "pa"}) || len(servers) != 3 {
		t.Errorf("with pyproject.toml: code %d, servers %v, checks %v", code, servers, checks)
	}
}

func TestLSPText(t *testing.T) {
	_, run := lspRepo(t)
	code, out, errText := run("lsp")
	want := "da-lsp    a  da --stdio  .d\nhost-lsp  b  hl          .h\n\n"
	if code != fail.OK || !strings.HasPrefix(out, want) || !strings.Contains(out, "2 language servers") {
		t.Errorf("code %d, stdout %q, stderr %q", code, out, errText)
	}
}

// TestLSPNoRepo: lsp exits 2 with no .config/bootstrap/, and outside a
// repository (D6).
func TestLSPNoRepo(t *testing.T) {
	root, run := lspRepo(t)
	if err := os.RemoveAll(filepath.Join(root, ".config", "bootstrap")); err != nil {
		t.Fatal(err)
	}
	code, out, errText := run("lsp", "--json")
	e := readEnvelope(t, out)
	if code != fail.Usage || string(e.Data) != "null" || !strings.Contains(errText, "Run bootstrap init") {
		t.Errorf("no values: code %d, envelope %+v, stderr %q", code, e, errText)
	}
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("PWD", dir)
	t.Setenv("MISE_PROJECT_ROOT", "")
	code, out, errText = run("lsp", "--json")
	e = readEnvelope(t, out)
	if code != fail.Usage || string(e.Data) != "null" {
		t.Errorf("outside a repository: code %d, envelope %+v, stderr %q", code, e, errText)
	}
}

// TestDescribeLSP: describe --json has the lsp field, the data of lsp.
func TestDescribeLSP(t *testing.T) {
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise is not available")
	}
	root, run := lspRepo(t)
	t.Setenv("MISE_TRUSTED_CONFIG_PATHS", root)
	code, out, errText := run("describe", "--json")
	e := readEnvelope(t, out)
	if code != fail.OK {
		t.Fatalf("code %d, envelope %+v, stderr %q", code, e, errText)
	}
	var d struct {
		LSP json.RawMessage `json:"lsp"`
	}
	if err := json.Unmarshal(e.Data, &d); err != nil {
		t.Fatal(err)
	}
	_, lspOut, _ := run("lsp", "--json")
	if string(d.LSP) != string(readEnvelope(t, lspOut).Data) {
		t.Errorf("describe lsp %s, lsp data %s", d.LSP, readEnvelope(t, lspOut).Data)
	}
}

func TestLSPMac(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	code, out, _ := runArgs("mac", "lsp", "--json")
	if e := readEnvelope(t, out); code != fail.Usage || e.Command != "" {
		t.Errorf("mac lsp: code %d, envelope %+v", code, e)
	}
}
