package tomledit

import (
	"testing"
)

// member is a repo.toml of a member in the layout of dprint: groups, blank
// lines, comments and arrays with spaces.
const member = `# The values of this repository.

format = 3

# The bootstrap commit of the last sync. The sync writes it.
bootstrap_commit = "abc"
core_branches = [ "main", "develop" ]
environments = [ "dev" ]
repo_name = "backend"

[branches.main]
landing = "direct-merge"
merge_from = [
  "develop",
]

[commits]
scopes = [ "api" ]
`

func decode(t *testing.T, text string) any {
	t.Helper()
	v, ok := Decode(text)
	if !ok {
		t.Fatalf("decode %s", text)
	}
	return v
}

func TestSet(t *testing.T) {
	cases := []struct {
		name, src  string
		path       []string
		value, out string
	}{
		{"a new table goes at the end, its keys sorted", member, []string{"home"},
			`{"url":"https://example.com/root","type":"submodule","folder":".."}`,
			member + "\n[home]\nfolder = \"..\"\ntype = \"submodule\"\nurl = \"https://example.com/root\"\n"},
		{"a table in place of the old one", "a = 1\n\n[home]\nfolder = \"x\"\n\n[z]\nb = 2\n", []string{"home"},
			`{"folder":".."}`, "a = 1\n\n[home]\nfolder = \"..\"\n\n[z]\nb = 2\n"},
		{"a table and its sub-table go, the new one comes", "[home]\nx = 1\n\n[home.sub]\ny = 2\n\n[z]\nb = 2\n", []string{"home"},
			`{"x":3}`, "[home]\nx = 3\n\n[z]\nb = 2\n"},
		{"a top-level key changes in place", member, []string{"repo_name"}, `"api"`,
			replace(member, `repo_name = "backend"`, `repo_name = "api"`)},
		{"an array on more lines becomes one line", member, []string{"branches", "main", "merge_from"}, `["develop","release"]`,
			replace(member, "merge_from = [\n  \"develop\",\n]", `merge_from = [ "develop", "release" ]`)},
		{"a new top-level key goes to its sorted place", member, []string{"default_branch"}, `"develop"`,
			replace(member, "core_branches = [ \"main\", \"develop\" ]\n", "core_branches = [ \"main\", \"develop\" ]\ndefault_branch = \"develop\"\n")},
		{"a new key of a table goes to its sorted place", member, []string{"commits", "style"}, `"conventional"`,
			replace(member, "scopes = [ \"api\" ]\n", "scopes = [ \"api\" ]\nstyle = \"conventional\"\n")},
		{"a key of a new table", member, []string{"worktree", "files"}, `[]`,
			member + "\n[worktree]\nfiles = []\n"},
		{"a key of an empty table goes after its header", "[tools.go]\n\n[tools.rust]\n", []string{"tools", "go", "v"}, `1`,
			"[tools.go]\nv = 1\n\n[tools.rust]\n"},
		{"a quoted key and an escaped string", "", []string{"branches", "release.1", "landing"}, `"a\"b"`,
			"[branches.\"release.1\"]\nlanding = \"a\\\"b\"\n"},
		{"an inline table in an array", "", []string{"x"}, `[{"b":1,"a":"y"}]`,
			"x = [ { a = \"y\", b = 1 } ]\n"},
	}
	for _, c := range cases {
		got, ok := Set([]byte(c.src), c.path, decode(t, c.value))
		if !ok {
			t.Errorf("%s: not ok", c.name)
			continue
		}
		if string(got) != c.out {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.out)
		}
	}
}

func TestSetNotOk(t *testing.T) {
	cases := []struct {
		name, src string
		path      []string
	}{
		{"a dotted key", "home.folder = \"..\"\n", []string{"repo_name"}},
		{"an array of tables", "[[x]]\na = 1\n", []string{"repo_name"}},
		{"an inline table on the path", "home = { folder = \"..\" }\n", []string{"home", "folder"}},
		{"an inline table at the path", "home = { folder = \"..\" }\n", []string{"home"}},
		{"a string that does not end", "a = \"x\n", []string{"b"}},
	}
	for _, c := range cases {
		value := any("v")
		if c.name == "an inline table at the path" {
			value = map[string]any{"folder": "x"}
		}
		if _, ok := Set([]byte(c.src), c.path, value); ok {
			t.Errorf("%s: ok, want not ok", c.name)
		}
	}
}

func TestDelete(t *testing.T) {
	cases := []struct {
		name, src string
		path      []string
		prune     bool
		out       string
	}{
		{"a key, the rest stays", member, []string{"environments"}, false,
			replace(member, "environments = [ \"dev\" ]\n", "")},
		{"a key on more lines", member, []string{"branches", "main", "merge_from"}, false,
			replace(member, "merge_from = [\n  \"develop\",\n]\n", "")},
		{"the last key, and its empty table with prune", member, []string{"commits", "scopes"}, true,
			replace(member, "\n[commits]\nscopes = [ \"api\" ]\n", "")},
		{"the last key, and the table stays without prune", member, []string{"commits", "scopes"}, false,
			replace(member, "scopes = [ \"api\" ]\n", "")},
		{"a table and its sub-tables", "a = 1\n\n[p]\nx = 1\n\n[p.q]\ny = 2\n\n[z]\nb = 2\n", []string{"p"}, false,
			"a = 1\n\n[z]\nb = 2\n"},
		{"a key alone between blank lines takes one", "a = 1\n\nb = 2\n\nc = 3\n", []string{"b"}, false,
			"a = 1\n\nc = 3\n"},
		{"a missing key changes nothing", member, []string{"nothing"}, true, member},
	}
	for _, c := range cases {
		got, ok := Delete([]byte(c.src), c.path, c.prune)
		if !ok {
			t.Errorf("%s: not ok", c.name)
			continue
		}
		if string(got) != c.out {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.out)
		}
	}
}

func TestEnsureTable(t *testing.T) {
	tools := "format = 3\n\n[tools.go]\n\n[tools.mise]\nx = 1\n\n[overrides.tools.mise]\ny = 2\n"
	cases := []struct {
		name, src string
		path      []string
		out       string
	}{
		{"after the last tools table", tools, []string{"tools", "rust"},
			"format = 3\n\n[tools.go]\n\n[tools.mise]\nx = 1\n\n[tools.rust]\n\n[overrides.tools.mise]\ny = 2\n"},
		{"a table that is there", tools, []string{"tools", "go"}, tools},
		{"a table with only a sub-table is there", "[tools.go.env]\na = 1\n", []string{"tools", "go"}, "[tools.go.env]\na = 1\n"},
		{"no tools table: at the end", "format = 3\n", []string{"tools", "go"}, "format = 3\n\n[tools.go]\n"},
		{"the last table of the file", "[tools.go]\n", []string{"tools", "rust"}, "[tools.go]\n\n[tools.rust]\n"},
	}
	for _, c := range cases {
		got, ok := EnsureTable([]byte(c.src), c.path)
		if !ok {
			t.Errorf("%s: not ok", c.name)
			continue
		}
		if string(got) != c.out {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.out)
		}
	}
}

func replace(s, old, new string) string {
	for i := 0; i+len(old) <= len(s); i++ {
		if s[i:i+len(old)] == old {
			return s[:i] + new + s[i+len(old):]
		}
	}
	panic("no " + old)
}
