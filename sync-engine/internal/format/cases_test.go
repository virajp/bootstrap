package format

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The inputs of the golden cases. UPDATE_GOLDEN=1 builds the cases from the
// merge sources of the repository (macos/tools and the fixture merger) and from
// the live files of this file, then writes them to testdata/.

type source struct {
	name, format, text, remove, items string
	show                              string
}

func readRepo(t testing.TB, rel string) string {
	b, err := os.ReadFile(filepath.Join(root(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

var tmplAction = regexp.MustCompile(`<%.*?%>`)

func repoSources(t testing.TB) []source {
	gh := tmplAction.ReplaceAllString(readRepo(t, "macos/tools/gh/.config/gh/config.yml.tmpl"), "vim")
	var claude []string
	for _, l := range strings.Split(readRepo(t, "macos/tools/claude-cli/.claude/settings.json.tmpl"), "\n") {
		if !strings.HasPrefix(l, "<%-") {
			claude = append(claude, l)
		}
	}
	cl := strings.NewReplacer("<% .env.HOME %>", "/Users/test", "<% $projects %>", "/Users/test/Projects").Replace(strings.Join(claude, "\n"))
	m := "macos/fixtures/repo/tools/merger/.config/merger/"
	mjson := strings.ReplaceAll(readRepo(t, m+"m.json.tmpl"), "<% .config.items | data.ToJSON %>", `["a"]`)
	return []source{
		{name: "gem", format: "yaml", text: readRepo(t, "macos/tools/gem/.gemrc"), remove: `[]`, items: `[]`, show: "~/.gemrc"},
		{name: "gh", format: "yaml", text: gh, remove: `[]`, items: `[]`, show: "~/.config/gh/config.yml"},
		{name: "superfile", format: "toml", text: readRepo(t, "macos/tools/superfile/Library/Application Support/superfile/config.toml"),
			remove: `[["sidebar_sections"]]`, items: `[]`, show: "~/Library/Application Support/superfile/config.toml"},
		{name: "claude", format: "json", text: cl, remove: `[["deniedMcpServers"],["env","CONTEXT7_RUNNER"]]`, items: `[]`, show: "~/.claude/settings.json"},
		{name: "merger-json", format: "json", text: mjson, remove: `[["gone"],["nested","old"],["absent","path"]]`,
			items: `[{"path":["list"],"items":["drop",{"j":2,"k":1}]},{"path":["keep"],"items":["app"]}]`, show: "~/.config/merger/m.json"},
		{name: "merger-yaml", format: "yaml", text: readRepo(t, m+"m.yaml.tmpl"), remove: `[["gone"]]`,
			items: `[{"path":["list"],"items":["drop"]}]`, show: "~/.config/merger/m.yaml"},
		{name: "merger-toml", format: "toml", text: readRepo(t, m+"m.toml.tmpl"), remove: `[["gone"]]`,
			items: `[{"path":["notlist"],"items":["s"]}]`, show: "~/.config/merger/m.toml"},
	}
}

// Sources that are not usual: they test the edge cases.
func extraSources() []source {
	e := func(name, format, text, remove, items string) source {
		return source{name: name, format: format, text: text, remove: remove, items: items, show: "~/x." + format}
	}
	return []source{
		e("x-json-list", "json", "[1]\n", `[]`, `[]`),
		e("x-json-null", "json", "null\n", `[["a"]]`, `[{"path":["l"],"items":[1]}]`),
		e("x-json-empty-table", "json", "{}\n", `[["gone"]]`, `[{"path":["list"],"items":["drop"]}]`),
		e("x-json-two", "json", "{} {}\n", `[]`, `[]`),
		e("x-json-blank", "json", "", `[]`, `[]`),
		e("x-json-bad", "json", `{"a": `, `[]`, `[]`),
		e("x-json-string", "json", `"s"`, `[]`, `[]`),
		e("x-json-types", "json", `{"n":1.0,"e":1e2,"big":100000000000000000000,"s":"é\u0001\u007f<>&\/ \ud83d\ude00","t":true,"z":null,"o":{"b":1,"a":[]},"l":[1,1.0,2,{"a":1,"b":2},{"b":2,"a":1}],"d":{}}`,
			`[]`, `[{"path":["l"],"items":[2]}]`),
		e("x-json-edge", "json", `{"l":[1,"a",1.0,{"b":1,"a":2},[1,2],"new"],"o":{"p":{"q":1}},"s":"x","dup":[1,1,1],"u":{"é":1,"z":2},"arr":["a","b"]}`,
			`[["o","p"],["l"],["u","é"],["arr","0"]]`, `[{"path":["l"],"items":[1,[1,2]]},{"path":["arr"],"items":["a"]},{"path":["dup"],"items":[2]}]`),
		e("x-json-nested-items", "json", `{"a":{"b":["x","y","drop"]},"c":[["drop"],"k"]}`, `[["a","b","0"],["c"]]`,
			`[{"path":["a","b"],"items":["drop","x"]},{"path":["c"],"items":[["drop"]]}]`),
		e("x-yaml-list", "yaml", "- a\n- b\n", `[]`, `[]`),
		e("x-yaml-two", "yaml", "a: 1\n---\nb: 2\n", `[]`, `[]`),
		e("x-yaml-comment", "yaml", "# nothing\n", `[["gone"]]`, `[]`),
		e("x-yaml-bad", "yaml", "a: [1\n", `[]`, `[]`),
		e("x-yaml-blank", "yaml", "", `[]`, `[]`),
		e("x-yaml-scalar", "yaml", "just a string\n", `[]`, `[]`),
		e("x-yaml-rich", "yaml", "# head\nname: bootstrap # inline\nnum: 1.0\nbig: 1e2\nflag: yes\nlist:\n  - a\n  - b # item\n  - drop\n  - {k: 1, j: 2}\nnested:\n  deep:\n    x: 1\ntext: |\n  line one\n  line two\nnull_key: ~\n",
			`[["gone"]]`, `[{"path":["list"],"items":["drop"]}]`),
		e("x-toml-blank", "toml", "", `[]`, `[]`),
		e("x-toml-comment", "toml", "# nothing\n", `[["gone"]]`, `[]`),
		e("x-toml-bad", "toml", "a = [\n", `[]`, `[]`),
		e("x-toml-rich", "toml", "# head\nname = \"bootstrap\" # inline\nnum = 1.0\nbig = 1e2\nflag = true\nlist = [\"a\", \"b\", \"drop\"]\n\n[nested]\nx = 1\n\n[nested.deep]\ny = \"z\"\n",
			`[["gone"]]`, `[{"path":["list"],"items":["drop"]}]`),
	}
}

type liveSet struct {
	name string
	text *string
}

var jsonLives = []liveSet{
	{"absent", nil},
	{"empty", s("")},
	{"blank", s("  \n\t \n")},
	{"merger", s("{\n    \"keep\": \"app\",\n    \"scalar\": \"app\",\n    \"nested\": {\"set\": 0, \"old\": 1, \"other\": 2},\n    \"kind\": {\"was\": \"table\"},\n    \"list\": [\"x\", \"drop\", \"b\", {\"k\": 1, \"j\": 2}, \"drop\"],\n    \"maps\": [{\"a\": 1, \"b\": 2}],\n    \"gone\": true\n}\n")},
	{"claude", s(`{
  "model": "sonnet",
  "env": {"CONTEXT7_RUNNER": "x", "MY": "1", "MISE_ENV": "prod"},
  "deniedMcpServers": [{"serverName": "a"}],
  "permissions": {"allow": ["Read", "Bash(ls:*)", "Edit"], "deny": ["Bash(rm:*)"], "ask": []},
  "autoMode": {"environment": ["old"], "soft_deny": ["$defaults", "mine"]},
  "attribution": "legacy",
  "numbers": [1.0, 1e2, 0.10, 100000000000000000000],
  "text": "é\u0001 \ud83d\ude00 <&>",
  "skillOverrides": {"b": "on", "a": "off"}
}
`)},
	{"kinds", s(`{"permissions": ["x"], "env": "str", "autoMode": null, "attribution": {"commit": "x"}, "scalar": [1], "nested": 3, "kind": 5, "list": {"a": 1}, "maps": "m"}`)},
	{"edge", s(`{"l":[1.0,2,"a",[1,2],{"a":2,"b":1}],"o":{"p":[1],"r":2},"s":{"k":1},"dup":[1],"u":{"z":1,"é":5},"arr":["a","c"]}`)},
	{"list", s("[1]\n")},
	{"null", s("null\n")},
	{"string", s("\"s\"\n")},
	{"bad", s(`{"a": `)},
	{"two", s("{} {}\n")},
	{"bom", s("\xef\xbb\xbf{\"a\": 1}\n")},
	{"dupkeys", s(`{"a": 1, "b": 2, "a": 3}`)},
	{"crlf", s("{\r\n  \"keep\": 1\r\n}\r\n")},
}

var yamlLives = []liveSet{
	{"absent", nil},
	{"empty", s("")},
	{"blank", s("  \n\n")},
	{"comments", s("  \n# only a comment\n\n")},
	{"docmarks", s("---\n# c\n...\n")},
	{"gem", s("# my gemrc\n:verbose: true # loud\n:backtrace: false\ngem: --no-document\nupdate: --user-install\n:sources:\n  - https://rubygems.org/\n")},
	{"gh", s("# gh config\nversion: \"1\"\ngit_protocol: https # old\neditor: nano\naliases:\n  co: pr checkout\n  mine: pr list --author @me\nhttp_unix_socket:\nbrowser:\n")},
	{"merger", s("# top comment\nkeep: app # after keep\nscalar: app\nnested:\n  # own line\n  set: 0\nlist:\n  - x\n  - drop\n  - drop\nkind:\n  was: table\n")},
	{"rich", s("# head\nanchors: &a\n  k: 1\nuse: *a\nflow: {x: 1, y: [1, 2]}\nnum: 1.0\nbig: 1e2\nhex: 0x1F\nflag: yes\nlist:\n  - a\n  - b\ntext: |\n  one\n  two\n\n# tail\n")},
	{"kinds", s("scalar:\n  - 1\nnested: 3\nlist: x\nkind: 5\nversion:\n  a: 1\n")},
	{"list", s("- a\n- b\n")},
	{"null", s("~\n")},
	{"scalar", s("text\n")},
	{"bad", s("a: [1\n")},
	{"two", s("a: 1\n---\nb: 2\n")},
	{"tabs", s("a:\n\t- b\n")},
}

var tomlLives = []liveSet{
	{"absent", nil},
	{"empty", s("")},
	{"blank", s("\n\t \n")},
	{"comments", s("# only a comment\n\n# another\n")},
	{"superfile", s("# my config\ntheme = \"monokai\" # theme\nsidebar_sections = [\"home\", \"x\"]\nzoxide_support = false\nfile_size_use_si = false\n\n[plugins]\nenabled = [\"a\"]\n")},
	{"merger", s("# head comment\nkeep = \"app\" # after keep\nscalar = \"app\"\nnotlist = \"s\"\ngone = 1\n\n# own line\n[nested]\nset = 0\nlist = [\"x\"]\n")},
	{"rich", s("# head\ninline = {a = 1, b = \"x\"}\nnum = 1.0\nbig = 1e2\nlist = [\"a\", \"drop\", \"b\"]\n\n[[items]]\nname = \"one\"\n\n[[items]]\nname = \"two\"\n\n[nested.deep]\ny = 1 # inline\n")},
	{"kinds", s("scalar = [1]\nnested = 3\nlist = \"x\"\ncase_sensitive_sort = {a = 1}\ndefault_sort_type = [4]\n")},
	{"bad", s("a = [\n")},
	{"bad2", s("a = 1\na = 2\n")},
}

func buildMergeCases(t testing.TB) []mergeCase {
	livesOf := map[string][]liveSet{"json": jsonLives, "yaml": yamlLives, "toml": tomlLives}
	var out []mergeCase
	for _, src := range repoSources(t) {
		for _, l := range livesOf[src.format] {
			out = append(out, mergeCase{name: src.name + "--" + l.name, format: src.format, src: src.text, live: l.text, remove: src.remove, items: src.items, show: src.show})
		}
	}
	short := map[string][]string{
		"json": {"absent", "empty", "merger", "claude", "list", "edge"},
		"yaml": {"absent", "empty", "comments", "merger", "gh"},
		"toml": {"absent", "empty", "comments", "merger", "superfile"},
	}
	for _, src := range extraSources() {
		for _, name := range short[src.format] {
			for _, l := range livesOf[src.format] {
				if l.name == name {
					out = append(out, mergeCase{name: src.name + "--" + l.name, format: src.format, src: src.text, live: l.text, remove: src.remove, items: src.items, show: src.show})
				}
			}
		}
	}
	return out
}

var checkCases = []checkCase{
	{"json-ok", "json", "{\"a\": 1}\n"},
	{"json-multi", "json", "{} {}\n[1] 2\n"},
	{"json-nan", "json", "[nan, NaN, Infinity, -Infinity, 0123, .5, +1, 1.]"},
	{"json-bom", "json", "\xef\xbb\xbf{}"},
	{"json-empty", "json", ""},
	{"json-blank", "json", " \n\t\n"},
	{"json-unfinished", "json", "{\"a\": "},
	{"json-trailing-comma", "json", "[1,]"},
	{"json-no-separator", "json", "[1 2]"},
	{"json-single-quote", "json", "{'a': 1}"},
	{"json-control", "json", "\"a\tb\""},
	{"json-bad-escape", "json", "\"\\x\""},
	{"json-lone-high", "json", "\"\\ud800\""},
	{"json-lone-low", "json", "\"\\ude00\""},
	{"json-comment", "json", "{} # c"},
	{"json-literal", "json", "tru"},
	{"json-unmatched", "json", "[1]]"},
	{"json-key", "json", "{1: 2}"},
	{"json-vtab", "json", "\v{}"},
	{"yaml-ok", "yaml", "a: 1\nb:\n  - x\n"},
	{"yaml-docs", "yaml", "a: 1\n---\nb: 2\n"},
	{"yaml-empty", "yaml", ""},
	{"yaml-comment", "yaml", "# c\n"},
	{"yaml-bad", "yaml", "a: [1\n"},
	{"yaml-tab", "yaml", "a:\n\t- b\n"},
	{"yaml-dup", "yaml", "a: 1\na: 2\n"},
	{"yaml-colon", "yaml", "a: b: c\n"},
	{"yaml-anchor", "yaml", "a: &x 1\nb: *x\n"},
	{"yaml-undefined-alias", "yaml", "b: *x\n"},
	{"toml-ok", "toml", "a = 1\n[t]\nb = \"x\"\n"},
	{"toml-empty", "toml", ""},
	{"toml-comment", "toml", "# c\n"},
	{"toml-bad", "toml", "a = [\n"},
	{"toml-dup", "toml", "a = 1\na = 2\n"},
	{"toml-bad-key", "toml", "= 1\n"},
	{"toml-date", "toml", "d = 2020-01-02\n"},
	{"text-anything", "text", "{{ not ( parsed"},
}
