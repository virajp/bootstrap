package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

// read decodes the TOML text <text>.
func read(t *testing.T, text string) *source.Obj {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.toml")
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := source.ReadTOML(p)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// messages gives the messages of <err>, one per line.
func messages(err error) string {
	var out []string
	for _, m := range source.Messages(err) {
		out = append(out, m.Error())
	}
	return strings.Join(out, "\n")
}

const head = "format = 3\nrepo_name = \"r\"\ncore_branches = [ \"main\", \"develop\" ]\n"

func TestCheckRepo(t *testing.T) {
	cases := []struct{ name, text, want string }{
		{"valid", head + `product_name = "p"
repo_url = "https://example.com/r"
environments = [ "dev", "ci" ]
bootstrap_commit = ""
topology = "multi-repo"
topology_reason = "one repo"
release_env = "prod"

[branches.main]
landing = "pr"
merge_from = [ "develop" ]

[branches.develop]
landing = "direct-merge"
merge_never = [ "main" ]

[commits]
style = "conventional"
scopes = [ "api" ]

[worktree]
files = [ ".config/mise.local.toml" ]

[members.web]
folder = "../web"
type = "sibling"
url = "https://example.com/web"
label = "Web"
worktree_install = "pnpm install"

[home]
folder = ".."
type = "submodule"
url = "https://example.com/home"

[deploy.prod]
web = "https://web.example.com"
`, ""},
		{"unknown key", head + "colour = \"red\"\n", "colour in repo.toml is not a key of repo.toml"},
		{"unknown nested key", head + "[branches.main]\nmerge_model = \"pr\"\n", "branches.main.merge_model in repo.toml is not a key of repo.toml"},
		{"bad enum", head + "[branches.main]\nlanding = \"merge\"\n", `branches.main.landing = "merge" in repo.toml is not one of pr, direct-merge, direct-commit`},
		{"bad member type", head + "[members.x]\ntype = \"repo\"\n", `members.x.type = "repo" in repo.toml is not one of submodule, folder, sibling`},
		{"bad topology", head + "topology = \"polyrepo\"\n", `topology = "polyrepo" in repo.toml is not one of repo, monorepo, multi-repo`},
		{"sibling without url", head + "[members.web]\nfolder = \"../web\"\ntype = \"sibling\"\n", "[members.web] in repo.toml has type sibling, but no url - add the url of its repo"},
		{"submodule without url", head + "[members.web]\ntype = \"submodule\"\n", ""},
		{"bad member name", head + "[members.\"a[$(cmd)]\"]\ntype = \"folder\"\n", `[members."a[$(cmd)]"] in repo.toml has a name that is not valid - a member name has only letters, digits, - and _`},
		{"member without type", head + "[members.web]\nfolder = \"web\"\n", "[members.web] in repo.toml has no type - set type to submodule, folder or sibling"},
		{"home without url", head + "[home]\nfolder = \"../home\"\ntype = \"sibling\"\n", ""},
		{"home without folder and type", head + "[home]\nurl = \"u\"\n", "[home] in repo.toml has no folder - run bootstrap sync from the home repo, which writes [home]\n[home] in repo.toml has no type - run bootstrap sync from the home repo, which writes [home]"},
		{"home of a folder member", head + "[home]\nfolder = \"..\"\ntype = \"folder\"\n", `home.type = "folder" in repo.toml is not one of submodule, sibling`},
		{"bad format", "format = 2\nrepo_name = \"r\"\ncore_branches = [ \"main\" ]\n", "format = 2 in repo.toml is not one of 3"},
		{"wrong type", head + "repo_url = 1\n", "repo_url in repo.toml must be a string"},
		{"wrong list type", "format = 3\nrepo_name = \"r\"\ncore_branches = \"main\"\n", "core_branches in repo.toml must be a list of strings"},
		{"missing repo_name", "format = 3\ncore_branches = [ \"main\" ]\n", "repo.toml has no repo_name - it is a required key of repo.toml"},
		{"missing all", "", "repo.toml has no format - it is a required key of repo.toml\nrepo.toml has no repo_name - it is a required key of repo.toml\nrepo.toml has no core_branches - it is a required key of repo.toml"},
		{"branch not core", "format = 3\nrepo_name = \"r\"\ncore_branches = [ \"main\" ]\n[branches.develop]\nlanding = \"pr\"\n", "[branches.develop] in repo.toml: develop is not in core_branches - add it there, or remove the table"},
		{"main direct-commit alone", head + "[branches.main]\nlanding = \"direct-commit\"\n", `[branches.main] in repo.toml has landing = "direct-commit", but [branches.develop] does not - set landing = "direct-commit" for develop too, or change main`},
		{"both direct-commit", head + "[branches.main]\nlanding = \"direct-commit\"\n[branches.develop]\nlanding = \"direct-commit\"\n", ""},
		{"main direct-commit, no develop", "format = 3\nrepo_name = \"r\"\ncore_branches = [ \"main\" ]\n[branches.main]\nlanding = \"direct-commit\"\n", ""},
	}
	for _, c := range cases {
		if got := messages(CheckRepo(read(t, c.text), "repo.toml")); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

func TestCheckReleases(t *testing.T) {
	cases := []struct{ text, want string }{
		{"format = 3\napi = \"MINOR\"\nweb = \"NONE\"\n", ""},
		{"format = 3\napi = \"BIG\"\n", `api = "BIG" in releases.toml is not one of NONE, PATCH, MINOR, MAJOR`},
		{"api = \"PATCH\"\n", "releases.toml has no format - it is a required key of releases.toml"},
		{"format = 3\n[api]\nlevel = \"PATCH\"\n", "api in releases.toml must be a string"},
	}
	for _, c := range cases {
		if got := messages(CheckReleases(read(t, c.text), "releases.toml")); got != c.want {
			t.Errorf("%q:\n got %q\nwant %q", c.text, got, c.want)
		}
	}
}

func TestFind(t *testing.T) {
	for path, want := range map[string]string{
		"branches.main.landing": "branches.*.landing",
		"members.web":           "members.*",
		"deploy.prod.web":       "deploy.*.*",
		"format":                "format",
		"repo_name.x":           "",
		"nope":                  "",
	} {
		got := ""
		if k := Find(Repo, strings.Split(path, ".")); k != nil {
			got = k.Path
		}
		if got != want {
			t.Errorf("Find(%s) = %q, want %q", path, got, want)
		}
	}
	// A literal segment wins over a name.
	if k := Find(Releases, []string{"format"}); k == nil || k.Type != Integer {
		t.Errorf("Find(releases format) = %+v", k)
	}
	if k := Find(Releases, []string{"api"}); k == nil || k.Default != "NONE" {
		t.Errorf("Find(releases api) = %+v", k)
	}
	if k := Find(Repo, []string{"branches", "x", "landing"}); k == nil || k.Default != LandingDirectMerge || !k.Sync {
		t.Errorf("landing = %+v", k)
	}
	for _, path := range []string{"deploy.prod.web", "release_env", "topology", "topology_reason", "bootstrap_commit", "format"} {
		if k := Find(Repo, strings.Split(path, ".")); k == nil || k.Sync {
			t.Errorf("%s: a sync does not read it (D90): %+v", path, k)
		}
	}
}

func TestCheckToolsShape(t *testing.T) {
	cases := []struct{ name, text, want string }{
		{"valid", "format = 3\n[tools.a]\nk = 1\n[overrides.tools.a]\nk = 2\n[overrides.mise.x]\nv = 1\n", ""},
		{"no format", "[tools.a]\n", "tools.toml has no format - it is a required key of tools.toml"},
		{"bad format", "format = 2\n", "format = 2 in tools.toml is not valid - the format of a values file is 3"},
		{"unknown key", "format = 3\nrepo_name = \"r\"\n", "repo_name is not a key of tools.toml - the top-level keys are format, [tools.<name>] and [overrides.*]"},
		{"override not table", "format = 3\n[overrides.tools]\na = 1\n", "[overrides.tools.a] in tools.toml must be a table"},
		{"moved tools key", "format = 3\n[tools.git.branch.main]\nmerge_model = \"pr\"\n", "[tools.git].branch in tools.toml is a key of repo.toml now - write it as branches.<branch> in repo.toml"},
		{"moved override key", "format = 3\n[overrides.tools.mise]\nworktree_files = []\n", "[overrides.tools.mise].worktree_files in tools.toml is a key of repo.toml now - write it as worktree.files in repo.toml"},
	}
	for _, c := range cases {
		if got := messages(CheckToolsShape(read(t, c.text), "tools.toml")); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

// TestCheckTools: a key of [tools.<name>] is env, files, or a question or a
// default of the tool in the catalog.
func TestCheckTools(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"repo/categories.toml":    "[category.core]\ntitle = \"Core\"\n",
		"repo/tools/a/tools.toml": "summary = \"A\"\ncategory = \"core\"\ndefault_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n[defaults]\nlevel = 1\n[questions.name]\nquestion = \"Name?\"\n",
	}
	for p, text := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	idx, err := source.Source{Root: root, Side: source.Repo}.Open()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, text, want string }{
		{"valid", "format = 3\n[tools.a]\nenv = [ \"dev\" ]\nlevel = 2\nname = \"x\"\n[tools.a.files.\".gitignore\"]\n\"*.log\" = { \"$mode\" = \"remove\" }\n", ""},
		{"unknown key", "format = 3\n[tools.a]\ncolour = 1\n", "[tools.a].colour in tools.toml is not a key of the tool a - its keys are env, files, and its questions and defaults"},
		{"unknown tool", "format = 3\n[tools.zzz]\ncolour = 1\n", ""},
		{"override keeps its rules", "format = 3\n[overrides.tools.a]\ncolour = 1\n", ""},
	}
	for _, c := range cases {
		if got := messages(CheckTools(idx, read(t, c.text), "tools.toml")); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}
