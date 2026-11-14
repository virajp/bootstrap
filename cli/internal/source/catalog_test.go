package source

import (
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// repoTool gives a repo tools.toml with the keys that each one needs, and
// <extra>.
func repoTool(extra string) string {
	return "default_envs = [\"dev\"]\ndepends = []\nroot_only = false\n" + extra
}

func messages(err error) []string {
	var out []string
	for _, e := range Messages(err) {
		out = append(out, e.Msg)
	}
	return out
}

func TestCatalogKeysRepo(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"repo/categories.toml": "[category.b]\ntitle = \"B\"\n[category.a]\ntitle = \"A\"\n",
		"repo/tools/t1/tools.toml": repoTool("summary = \"" + strings.Repeat("é", SummaryMax) + "\"\ncategory = \"a\"\n" +
			"selectable = false\nhosts = [\"macos\", \"linux\"]\ndescription = \"long text\"\n"),
		"repo/tools/t2/tools.toml": "default_envs = [\"dev\"]\ndepends = [\"t1\"]\nroot_only = false\nsummary = \"t2\"\ncategory = \"b\"\n",
	})
	x, err := Source{Root: root, Side: Repo}.Open()
	if err != nil {
		t.Fatalf("open: %v", messages(err))
	}
	m := x.Manifests["t1"]
	if m.Summary == "" || m.Category != "a" || m.Selectable || m.Description != "long text" ||
		!reflect.DeepEqual(m.Hosts, []string{"macos", "linux"}) {
		t.Errorf("manifest = %+v", m)
	}
	if !x.Manifests["t2"].Selectable {
		t.Error("t2: selectable must be true without the key")
	}
	if !reflect.DeepEqual(x.Categories.IDs, []string{"b", "a"}) || x.Categories.Titles["a"] != "A" {
		t.Errorf("categories = %+v", x.Categories)
	}
	if !m.RunsOn("linux") || m.RunsOn("windows") || !x.Manifests["t2"].RunsOn("plan9") {
		t.Error("RunsOn")
	}
	if got := x.NotSelectable("t1"); got != "t1 is not selectable. Select t2." {
		t.Errorf("NotSelectable = %q", got)
	}
	if got := x.NotSelectable("t2"); got != "t2 is not selectable." {
		t.Errorf("NotSelectable = %q", got)
	}
	if got := m.HostMessage("windows"); got != "t1 needs the host macos or linux - this host is windows" {
		t.Errorf("HostMessage = %q", got)
	}
}

func TestCatalogKeysRepoBad(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"repo/tools/a/tools.toml": repoTool("summary = \"" + strings.Repeat("x", SummaryMax+1) + "\"\ncategory = 1\nselectable = \"no\"\ndescription = 2\nhosts = []\n"),
		"repo/tools/b/tools.toml": repoTool("summary = \"b\"\nhosts = [\"darwin\"]\ncategory = \"x\"\n"),
		"repo/tools/c/tools.toml": repoTool("hosts = \"macos\"\nsummary = 1\ncategory = \"x\"\n"),
	})
	_, err := Source{Root: root, Side: Repo}.Open()
	want := []string{
		"repo/tools/a/tools.toml: summary must be a string of 47 characters or fewer",
		"repo/tools/a/tools.toml: category must be a string",
		"repo/tools/a/tools.toml: selectable must be true or false",
		"repo/tools/a/tools.toml: description must be a string",
		"repo/tools/a/tools.toml: hosts must name at least one host",
		"repo/tools/b/tools.toml: hosts must be a list of macos, linux and windows",
		"repo/tools/b/tools.toml: category x needs the file repo/categories.toml",
		"repo/tools/c/tools.toml: summary must be a string of 47 characters or fewer",
		"repo/tools/c/tools.toml: hosts must be a list of macos, linux and windows",
		"repo/tools/c/tools.toml: category x needs the file repo/categories.toml",
	}
	if got := messages(err); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestCategoriesFile(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"macos/categories.toml":     "top = 1\n[category.a]\ntitle = \"A\"\n[category.b]\n[category.c]\ntitle = 3\n[category.d]\ntitle = \"D\"\nicon = \"x\"\n",
		"macos/tools/m1/tools.toml": "summary = \"s\"\ncategory = \"a\"\n",
		"macos/tools/m2/tools.toml": "summary = \"s\"\ncategory = \"zz\"\n",
	})
	x, err := Source{Root: root, Side: Macos}.Open()
	want := []string{
		"macos/categories.toml: top is not a key of the categories file - each entry is [category.<id>] with a title",
		"macos/categories.toml: [category.b] needs a title string",
		"macos/categories.toml: [category.c] title must be a string",
		"macos/categories.toml: [category.d] icon is not a key of a category - the only key is title",
		"macos/tools/m2/tools.toml: category zz is not a category of macos/categories.toml",
	}
	if got := messages(err); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if !reflect.DeepEqual(x.Categories.IDs, []string{"a", "d"}) {
		t.Errorf("ids = %v", x.Categories.IDs)
	}

	write(t, root, map[string]string{"macos/categories.toml": "[category\n"})
	_, err = Source{Root: root, Side: Macos}.Open()
	if got := messages(err); len(got) == 0 || got[0] != "cannot read macos/categories.toml" {
		t.Errorf("a file that does not parse: %v", got)
	}
}

func TestCatalogKeysMacos(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"macos/categories.toml": "[category.a]\ntitle = \"A\"\n",
		"macos/tools/m1/tools.toml": "summary = \"s\"\ncategory = \"a\"\nselectable = false\nhosts = [\"macos\"]\n" +
			"[questions.a]\nquestion = \"A?\"\ngroup = \"g\"\nnote = \"n\"\noptional = false\n" +
			"[questions.b]\nquestion = \"B?\"\ngroup = \"g\"\noptional = true\n" +
			"[questions.c]\nquestion = \"C?\"\ngroup = 1\nnote = 2\noptional = \"yes\"\n",
	})
	_, err := Source{Root: root, Side: Macos}.Open()
	want := []string{
		"macos/tools/m1/tools.toml: hosts is for repo tools only",
		"macos/tools/m1/tools.toml: [questions.b] has a group, so it cannot be optional",
		"macos/tools/m1/tools.toml: [questions.c] group must be a string",
		"macos/tools/m1/tools.toml: [questions.c] note must be a string",
		"macos/tools/m1/tools.toml: [questions.c] optional must be true or false",
	}
	if got := messages(err); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// A tool with no summary or no category, or with an empty one, stops the
// sync, on both sides.
func TestCatalogKeysNeeded(t *testing.T) {
	for _, side := range []Side{Repo, Macos} {
		head := ""
		if side == Repo {
			head = repoTool("")
		}
		at := string(side) + "/tools/"
		root := t.TempDir()
		write(t, root, map[string]string{
			at + "a/tools.toml":               head,
			at + "b/tools.toml":               head + "summary = \"\"\ncategory = \"\"\n",
			at + "c/tools.toml":               head + "summary = \"s\"\n",
			at + "d/tools.toml":               head + "category = \"k\"\n",
			at + "e/tools.toml":               head + "summary = \"s\"\ncategory = \"k\"\n",
			string(side) + "/categories.toml": "[category.k]\ntitle = \"K\"\n",
		})
		_, err := Source{Root: root, Side: side}.Open()
		want := []string{
			at + "a/tools.toml: summary is missing",
			at + "a/tools.toml: category is missing",
			at + "b/tools.toml: summary is missing",
			at + "b/tools.toml: category is missing",
			at + "c/tools.toml: category is missing",
			at + "d/tools.toml: summary is missing",
		}
		if got := messages(err); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got\n%s\nwant\n%s", side, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
		if fail.CodeOf(err) != fail.Usage {
			t.Errorf("%s: code = %v, want %v", side, fail.CodeOf(err), fail.Usage)
		}
	}
}

func TestHostName(t *testing.T) {
	old := GOOS
	defer func() { GOOS = old }()
	for in, want := range map[string]string{"darwin": "macos", "linux": "linux", "windows": "windows", "plan9": "plan9"} {
		GOOS = in
		if got := HostName(); got != want {
			t.Errorf("HostName(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestPackDescription(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"repo/packs.toml": "[pack.base]\ndescription = \"The base pack.\"\na = { default = true }\n[pack.odd]\ndescription = 1\n",
	})
	p, err := Source{Root: root, Side: Repo}.ReadPacks()
	if err != nil {
		t.Fatal(err)
	}
	if p.Descriptions["base"] != "The base pack." {
		t.Errorf("descriptions = %v", p.Descriptions)
	}
	if !reflect.DeepEqual(p.BadDescriptions, []string{"odd"}) {
		t.Errorf("bad descriptions = %v", p.BadDescriptions)
	}
	if len(p.Members) != 1 || p.Members[0].Key != "a" || len(p.Bad) != 0 {
		t.Errorf("members = %v, bad = %v", p.Members, p.Bad)
	}
	if !reflect.DeepEqual(p.Order, []Order{{"base", "a"}}) {
		t.Errorf("order = %v", p.Order)
	}
}
