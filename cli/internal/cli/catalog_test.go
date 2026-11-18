package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// catalogJSON is the data of catalog --json, as a test reads it.
type catalogJSON struct {
	FetchedAt  *string `json:"fetched_at"`
	Categories struct {
		IDs    []string          `json:"ids"`
		Titles map[string]string `json:"titles"`
	} `json:"categories"`
	Tools []struct {
		Name      string `json:"name"`
		Category  string `json:"category"`
		Questions []struct {
			Key  string `json:"key"`
			Text string `json:"text"`
		} `json:"questions"`
		Defaults map[string]any `json:"defaults"`
	} `json:"tools"`
	Packs struct {
		Names   []string `json:"names"`
		Members []struct {
			Pack string `json:"pack"`
			Key  string `json:"key"`
		} `json:"members"`
	} `json:"packs"`
}

func TestCatalogJSON(t *testing.T) {
	_, run := valuesRepo(t)
	check := func(where string) {
		code, out, errText := run("catalog", "--json")
		e := readEnvelope(t, out)
		if code != fail.OK || e.Command != "catalog" {
			t.Fatalf("%s: code %d, envelope %+v, stderr %q", where, code, e, errText)
		}
		var d catalogJSON
		if err := json.Unmarshal(e.Data, &d); err != nil {
			t.Fatalf("%s: data %s: %v", where, e.Data, err)
		}
		// --clone: no fetch, so no fetch time.
		if d.FetchedAt == nil || *d.FetchedAt != "" {
			t.Errorf("%s: fetched_at %v", where, d.FetchedAt)
		}
		if strings.Join(d.Categories.IDs, ",") != "core" || d.Categories.Titles["core"] != "Core" {
			t.Errorf("%s: categories %+v", where, d.Categories)
		}
		if len(d.Tools) != 2 || d.Tools[0].Name != "a" || d.Tools[1].Name != "b" || d.Tools[0].Category != "core" {
			t.Fatalf("%s: tools %s", where, e.Data)
		}
		if len(d.Tools[0].Questions) != 1 || d.Tools[0].Questions[0].Key != "name" || d.Tools[0].Questions[0].Text != "N?" {
			t.Errorf("%s: questions of a %+v", where, d.Tools[0].Questions)
		}
		if d.Tools[0].Defaults["flag"] != true || d.Tools[1].Questions == nil {
			t.Errorf("%s: tools %s", where, e.Data)
		}
		if strings.Join(d.Packs.Names, ",") != "base" || len(d.Packs.Members) != 1 || d.Packs.Members[0].Key != "a" {
			t.Errorf("%s: packs %+v", where, d.Packs)
		}
		if strings.Contains(string(e.Data), "null") {
			t.Errorf("%s: data %s has null", where, e.Data)
		}
	}
	check("in a repository")
	// Outside a repository: the catalog needs only the clone (D28).
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("PWD", dir)
	t.Setenv("MISE_PROJECT_ROOT", "")
	check("outside a repository")
}

func TestCatalogText(t *testing.T) {
	_, run := valuesRepo(t)
	code, out, _ := run("catalog")
	if code != fail.OK || !strings.HasPrefix(out, "Core\n  a  A tool\n  b  A tool\n") || !strings.Contains(out, "2 tools") {
		t.Errorf("code %d, stdout %q", code, out)
	}
}

// describeRepo is the fixture of valuesRepo with the folder member web, the
// sibling member api whose folder is not there, and the task git:merge:main.
func describeRepo(t *testing.T) func(args ...string) (fail.Code, string, string) {
	t.Helper()
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise is not available")
	}
	dir, run := valuesRepo(t)
	root := filepath.Dir(filepath.Dir(dir))
	writeFile(t, filepath.Join(dir, "repo.toml"), "format = 3\nrepo_name = \"x\"\ncore_branches = [ \"main\" ]\n\n"+
		"[members.web]\ntype = \"folder\"\nfolder = \"web\"\n\n[members.api]\ntype = \"sibling\"\nfolder = \"../api\"\nurl = \"u\"\n", 0o644)
	writeFile(t, filepath.Join(root, "web", "f"), "", 0o644)
	writeFile(t, filepath.Join(root, ".config", "mise", "tasks", "git", "merge", "main"), "#!/bin/bash\n", 0o755)
	t.Setenv("MISE_TRUSTED_CONFIG_PATHS", root)
	return run
}

func TestDescribeJSON(t *testing.T) {
	run := describeRepo(t)
	code, out, errText := run("describe", "--json")
	e := readEnvelope(t, out)
	if code != fail.OK || e.Command != "describe" {
		t.Fatalf("code %d, envelope %+v, stderr %q", code, e, errText)
	}
	var d struct {
		Repo  map[string]any `json:"repo"`
		Tools []struct {
			Name     string         `json:"name"`
			Category string         `json:"category"`
			Values   map[string]any `json:"values"`
		} `json:"tools"`
		Members []describeMem `json:"members"`
		Ports   []struct {
			Name    string    `json:"name"`
			Present bool      `json:"present"`
			Tasks   *[]string `json:"tasks"`
		} `json:"ports"`
	}
	if err := json.Unmarshal(e.Data, &d); err != nil {
		t.Fatalf("data %s: %v", e.Data, err)
	}
	if d.Repo["repo_name"] != "x" {
		t.Errorf("repo %v", d.Repo)
	}
	// Each tool has its category from the catalog (D11).
	if len(d.Tools) != 1 || d.Tools[0].Name != "a" || d.Tools[0].Category != "core" || d.Tools[0].Values["flag"] != false {
		t.Errorf("tools %s", e.Data)
	}
	want := []describeMem{
		{Name: "api", Type: "sibling", Folder: "../api", URL: "u", Present: false},
		{Name: "web", Type: "folder", Folder: "web", Present: true},
	}
	if len(d.Members) != 2 || d.Members[0] != want[0] || d.Members[1] != want[1] {
		t.Errorf("members %+v", d.Members)
	}
	ports := map[string]bool{}
	for _, p := range d.Ports {
		ports[p.Name] = p.Present
		switch p.Name {
		case "git:merge:<branch>":
			if p.Tasks == nil || strings.Join(*p.Tasks, ",") != "git:merge:main" {
				t.Errorf("port %+v", p)
			}
		case "p:<id>:deploy":
			if p.Tasks == nil || len(*p.Tasks) != 0 {
				t.Errorf("port %+v", p)
			}
		default:
			if p.Tasks != nil {
				t.Errorf("a plain port with tasks: %+v", p)
			}
		}
	}
	if len(d.Ports) != 14 || !ports["git:merge:<branch>"] || ports["code:lint"] || ports["p:<id>:deploy"] {
		t.Errorf("ports %s", e.Data)
	}
}

func TestDescribeText(t *testing.T) {
	run := describeRepo(t)
	code, out, errText := run("describe")
	for _, w := range []string{"repo\n", "  repo_name ", "tools\n  a ", "members\n", "  web  folder   web     [OK] present", "[OK] present", "ports\n", "  git:merge:<branch>   [OK] present (git:merge:main)", "  code:lint            [WARN] missing"} {
		if !strings.Contains(out, w) {
			t.Errorf("stdout lacks %q: code %d\n%s%s", w, code, out, errText)
		}
	}
}

// TestDescribeUntrusted: a mise that cannot list the tasks stops describe
// with code 1 and an error that names mise trust (D27).
func TestDescribeUntrusted(t *testing.T) {
	run := describeRepo(t)
	t.Setenv("MISE_TRUSTED_CONFIG_PATHS", "")
	code, out, _ := run("describe", "--json")
	e := readEnvelope(t, out)
	if code != fail.Other || string(e.Data) != "null" || len(e.Errors) != 1 || !strings.Contains(e.Errors[0].Message, "mise trust") {
		t.Errorf("code %d, envelope %+v", code, e)
	}
}

// TestDescribeNoValues: a repository without .config/bootstrap/ stops with
// code 2 and the fix bootstrap migrate (D28).
func TestDescribeNoValues(t *testing.T) {
	dir, run := valuesRepo(t)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	code, out, errText := run("describe", "--json")
	e := readEnvelope(t, out)
	if code != fail.Usage || string(e.Data) != "null" || !strings.Contains(errText, "Run bootstrap migrate") {
		t.Errorf("code %d, envelope %+v, stderr %q", code, e, errText)
	}
}

func TestPortsMac(t *testing.T) {
	setupEnv(t, "/nonexistent/bootstrap/config.toml")
	for _, cmd := range []string{"catalog", "describe"} {
		code, out, _ := runArgs("mac", cmd, "--json")
		e := readEnvelope(t, out)
		if code != fail.Usage || e.Command != "" {
			t.Errorf("mac %s: code %d, envelope %+v", cmd, code, e)
		}
	}
}

// TestDescribeProjects: the projects come through data.repo and the values
// of each tool (D11).
func TestDescribeProjects(t *testing.T) {
	run := describeRepo(t)
	if code, _, errText := run("config", "set", "repo.projects.web", `{"folder":"web"}`); code != fail.OK {
		t.Fatalf("set: %d %s", code, errText)
	}
	if code, _, errText := run("config", "set", "tools.a.projects", `["web"]`); code != fail.OK {
		t.Fatalf("set: %d %s", code, errText)
	}
	code, out, errText := run("describe", "--json")
	e := readEnvelope(t, out)
	var d struct {
		Repo struct {
			Projects map[string]map[string]string `json:"projects"`
		} `json:"repo"`
		Tools []struct {
			Name   string         `json:"name"`
			Values map[string]any `json:"values"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(e.Data, &d); err != nil || code != fail.OK {
		t.Fatalf("code %d, data %s: %v, stderr %q", code, e.Data, err, errText)
	}
	if d.Repo.Projects["web"]["folder"] != "web" || len(d.Tools) != 1 || fmt.Sprint(d.Tools[0].Values["projects"]) != "[web]" {
		t.Errorf("data %s", e.Data)
	}
}
