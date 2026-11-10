package format

import (
	"strings"
	"testing"
)

var testHeader = []string{"Shared with bootstrap: x.", "A change stays."}

type sharedCase struct {
	name      string
	target    string
	spec      string
	live      *string
	overrides string
	records   []string
	want      string
	wantRows  []string
	warn      int
	fail      string
}

func sp(s string) *string { return &s }

func runShared(t *testing.T, c sharedCase) {
	t.Helper()
	var live []byte
	if c.live != nil {
		live = []byte(*c.live)
	}
	var ov []byte
	if c.overrides != "" {
		ov = []byte(c.overrides)
	}
	res, fl := MergeShared(SharedInput{
		Spec: []byte(c.spec), Live: live, Target: c.target, Overrides: ov,
		Records: c.records, Header: testHeader, Show: c.target, SpecShow: "spec",
	})
	if c.fail != "" {
		if fl == nil {
			t.Fatalf("no failure, want %q; out:\n%s", c.fail, res.Out)
		}
		if !strings.Contains(fl.Line+fl.Text, c.fail) {
			t.Fatalf("failure %q, want %q", fl.Line, c.fail)
		}
		return
	}
	if fl != nil {
		t.Fatalf("failure: %s %s %s", fl.Line, fl.Text, fl.Stderr)
	}
	if string(res.Out) != c.want {
		t.Errorf("output differs\ngot:\n%s\nwant:\n%s", res.Out, c.want)
	}
	if c.wantRows != nil && strings.Join(res.Records, "\n") != strings.Join(c.wantRows, "\n") {
		t.Errorf("rows differ\ngot:  %q\nwant: %q", res.Records, c.wantRows)
	}
	if len(res.Warnings) != c.warn {
		t.Errorf("warnings %q, want %d", res.Warnings, c.warn)
	}
	// A second run on the output changes nothing.
	again, fl := MergeShared(SharedInput{
		Spec: []byte(c.spec), Live: res.Out, Target: c.target, Overrides: ov,
		Records: res.Records, Header: testHeader, Show: c.target, SpecShow: "spec",
	})
	if fl != nil || string(again.Out) != string(res.Out) {
		t.Errorf("second run differs\n%s", again.Out)
	}
}

func TestSharedKinds(t *testing.T) {
	cases := []sharedCase{
		{
			name: "key-added-and-kept", target: "a.json",
			spec: `
[[item]]
kind = "key"
path = [ "a", "b" ]
value = 1
[[item]]
kind = "key"
path = [ "c" ]
value = "x"
`,
			live:     sp(`{"c": "mine", "d": true}`),
			want:     "{\n  \"c\": \"mine\",\n  \"d\": true,\n  \"a\": {\n    \"b\": 1\n  }\n}\n",
			wantRows: []string{`{"kind":"key","path":["a","b"]}`},
		},
		{
			name: "key-of-another-type", target: "a.json",
			spec: `
[[item]]
kind = "key"
path = [ "a" ]
value = [ "x" ]
`,
			live: sp(`{"a": "x"}`),
			want: "{\n  \"a\": [\n    \"x\"\n  ]\n}\n",
			warn: 1,
		},
		{
			name: "keys-of-a-table", target: "a.json",
			spec: `
[[item]]
kind = "keys"
path = [ "t" ]
value = { a = 1, b = { c = 2 }, e = {} }
`,
			live: sp(`{"t": {"a": 5, "z": 0}}`),
			want: "{\n  \"t\": {\n    \"a\": 5,\n    \"z\": 0,\n    \"b\": {\n      \"c\": 2\n    },\n    \"e\": {}\n  }\n}\n",
		},
		{
			name: "item-of-a-list", target: "a.json",
			spec: `
[[item]]
kind = "item"
path = [ "l" ]
value = "npm:a@2"
match = '^npm:a@'
[[item]]
kind = "item"
path = [ "l" ]
value = "b"
`,
			live:     sp(`{"l": ["mine", "npm:a@1"]}`),
			want:     "{\n  \"l\": [\n    \"mine\",\n    \"npm:a@1\",\n    \"b\"\n  ]\n}\n",
			wantRows: []string{`{"kind":"item","path":["l"],"value":"b"}`},
		},
		{
			name: "block-in-a-table-list", target: "a.yaml",
			spec: `
[[item]]
kind = "block"
path = [ "repos" ]
id = "one"
match = [ "repo", "hooks.id" ]
ignore = [ "rev" ]
value = { repo = "r", rev = "v2", hooks = [ { id = "h", name = "new" } ] }
`,
			live: sp("# top comment\nrepos:\n  # own\n  - repo: mine\n  - repo: r\n    rev: v1 # pinned\n    hooks:\n      - id: h\n        name: old\n"),
			want: "# " + testHeader[0] + "\n# " + testHeader[1] + "\n# top comment\nrepos:\n  # own\n  - repo: mine\n  - repo: r\n    rev: v1\n    hooks:\n      - id: h\n        name: new\n",
		},
		{
			name: "block-of-a-table", target: "a.toml",
			spec: `
[[item]]
kind = "block"
path = [ "owned" ]
value = [ "a", "b" ]
`,
			live: sp("# keep\nowned = [ \"b\", \"x\" ]\nmine = 1 # mine\n"),
			want: "# " + testHeader[0] + "\n# " + testHeader[1] + "\n# keep\nowned = [\"a\", \"b\"]\nmine = 1  # mine\n",
		},
		{
			name: "text-lines", target: ".gitignore",
			spec: `
format = "text"
comment = "#"
[[item]]
value = "node_modules/"
[[item]]
match = '^\.env(\..*)?$'
value = ".env"
[[item]]
value = "*.log"
`,
			live: sp("# mine\n.env.local\nbuild/\n"),
			want: "# " + testHeader[0] + "\n# " + testHeader[1] + "\n# mine\n.env.local\nbuild/\nnode_modules/\n*.log\n",
		},
		{
			name: "absent-yaml", target: "a.yaml",
			spec: `
[[item]]
kind = "keys"
value = { a = 1, b = [ "x" ] }
`,
			want: "# " + testHeader[0] + "\n# " + testHeader[1] + "\na: 1\nb:\n  - x\n",
		},
		{
			name: "absent-toml", target: "a.toml",
			spec: `
[[item]]
kind = "keys"
path = [ "t" ]
value = { a = 1 }
`,
			want: "# " + testHeader[0] + "\n# " + testHeader[1] + "\n[t]\na = 1\n",
		},
		{
			name: "jsonc-comments", target: "d.jsonc",
			spec: `
[[item]]
kind = "keys"
path = [ "m" ]
value = { b = 2 }
[[item]]
kind = "item"
path = [ "p" ]
value = "y"
[[item]]
kind = "key"
path = [ "extends" ]
"$mode" = "remove"
`,
			live: sp("{\n  \"extends\": \"~/x\",\n  // keep me\n  \"m\": { \"a\": 1 }, // and me\n  \"p\": [\"x\",],\n}\n"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.name == "jsonc-comments" {
				res, fl := MergeShared(SharedInput{Spec: []byte(c.spec), Live: []byte(*c.live), Target: c.target, Header: testHeader, Show: c.target, SpecShow: "spec"})
				if fl != nil {
					t.Fatal(fl.Line, fl.Stderr)
				}
				out := string(res.Out)
				for _, want := range []string{"// keep me", "// and me", `"b"`, `"y"`, "// " + testHeader[0]} {
					if !strings.Contains(out, want) {
						t.Errorf("no %q in\n%s", want, out)
					}
				}
				if strings.Contains(out, "extends") {
					t.Errorf("extends stays in\n%s", out)
				}
				return
			}
			runShared(t, c)
		})
	}
}

func TestSharedDollarKeys(t *testing.T) {
	spec := `
[[item]]
kind = "keys"
path = [ "m" ]
value = { a = 1, b = 2, c = 3 }
[[item]]
kind = "item"
path = [ "l" ]
value = "x"
[[item]]
kind = "block"
path = [ "r" ]
id = "one"
match = [ "n" ]
value = { n = "one", v = 1 }
`
	runShared(t, sharedCase{
		name: "overrides", target: "a.json", spec: spec,
		live:      sp(`{"m": {"c": 3}, "l": ["x"], "r": [{"n": "one", "v": 0}]}`),
		overrides: `{"m": {"a": 9, "b": {"$skip": true}, "c": {"$mode": "remove"}}, "l": {"x": {"$mode": "remove"}}, "r": {"one": {"$mode": "remove"}}}`,
		want:      "{\n  \"m\": {\n    \"a\": 9\n  },\n  \"l\": [],\n  \"r\": []\n}\n",
	})
	runShared(t, sharedCase{
		name: "unused", target: "a.json", spec: spec, live: sp(`{}`),
		overrides: `{"m": {"zz": 1}}`,
		fail:      "the override m.zz has no item",
	})
	runShared(t, sharedCase{
		name: "spec-remove", target: "a.json",
		spec: "[[item]]\nkind = \"key\"\npath = [ \"old\" ]\n\"$mode\" = \"remove\"\n",
		live: sp(`{"old": 1, "x": 2}`),
		want: "{\n  \"x\": 2\n}\n",
	})
}

func TestSharedCleanup(t *testing.T) {
	// The last run added "gone" and the block "b"; the spec no longer has
	// them, so the merge removes them. "mine" was never added by bootstrap.
	runShared(t, sharedCase{
		name: "cleanup", target: "a.json",
		spec: "[[item]]\nkind = \"item\"\npath = [ \"l\" ]\nvalue = \"keep\"\n",
		live: sp(`{"l": ["keep", "gone", "mine"], "r": [{"n": "b"}, {"n": "c"}], "k": 1}`),
		records: []string{
			`{"kind":"item","path":["l"],"value":"keep"}`,
			`{"kind":"item","path":["l"],"value":"gone"}`,
			`{"kind":"block","path":["r"],"id":"b","find":{"n":"b"}}`,
			`{"kind":"key","path":["k"]}`,
		},
		want:     "{\n  \"l\": [\n    \"keep\",\n    \"mine\"\n  ],\n  \"r\": [\n    {\n      \"n\": \"c\"\n    }\n  ]\n}\n",
		wantRows: []string{`{"kind":"item","path":["l"],"value":"keep"}`},
	})
	// A block whose match values changed is found by its recorded find.
	runShared(t, sharedCase{
		name: "block-moved", target: "a.json",
		spec: "[[item]]\nkind = \"block\"\npath = [ \"r\" ]\nid = \"b\"\nmatch = [ \"n\" ]\nvalue = { n = \"new\" }\n",
		live: sp(`{"r": [{"n": "old"}]}`),
		records: []string{
			`{"kind":"block","path":["r"],"id":"b","find":{"n":"old"}}`,
		},
		want:     "{\n  \"r\": [\n    {\n      \"n\": \"new\"\n    }\n  ]\n}\n",
		wantRows: []string{`{"kind":"block","path":["r"],"id":"b","find":{"n":"new"}}`},
	})
}

func TestSharedRoundTrip(t *testing.T) {
	// A change in each format keeps the comments of the live file.
	for _, c := range []struct{ target, live, comment string }{
		{"a.yaml", "# head\na: 1 # inline\n# before b\nb: 2\n", "# inline"},
		{"a.toml", "# head\na = 1 # inline\n\n[t]\n# in t\nb = 2\n", "# in t"},
		{"a.jsonc", "{\n  // head\n  \"a\": 1, // inline\n}\n", "// inline"},
		{".x", "# head\nline\n", "# head"},
	} {
		t.Run(c.target, func(t *testing.T) {
			spec := "[[item]]\nkind = \"key\"\npath = [ \"new\" ]\nvalue = \"v\"\n"
			if c.target == ".x" {
				spec = "format = \"text\"\ncomment = \"#\"\n[[item]]\nvalue = \"new\"\n"
			}
			res, fl := MergeShared(SharedInput{Spec: []byte(spec), Live: []byte(c.live), Target: c.target, Header: testHeader, Show: c.target, SpecShow: "spec"})
			if fl != nil {
				t.Fatal(fl.Line, fl.Stderr)
			}
			if !strings.Contains(string(res.Out), c.comment) || !strings.Contains(string(res.Out), "new") {
				t.Errorf("comment %q or new item lost:\n%s", c.comment, res.Out)
			}
		})
	}
}

func TestSharedOldHeader(t *testing.T) {
	// The header of the owned file of an older sync goes.
	runShared(t, sharedCase{
		name: "old-header", target: ".gitignore",
		spec: "format = \"text\"\ncomment = \"#\"\n[[item]]\nvalue = \"a\"\n",
		live: sp("# Synced from bootstrap: x. Do not edit this file.\n# The next sync replaces it, and a manual change is lost.\n\n# mine\na\n"),
		want: "# " + testHeader[0] + "\n# " + testHeader[1] + "\n\n# mine\na\n",
	})
}

func TestSpecErrors(t *testing.T) {
	for spec, want := range map[string]string{
		"foo = 1":                           "foo is not a key",
		"[[item]]\nkind = \"x\"\nvalue = 1": "kind must be",
		"[[item]]\nkind = \"block\"\npath = [\"a\"]\nmatch = [\"n\"]\nvalue = {n=1}": "needs an id",
		"[[item]]\nkind = \"key\"\npath = [\"a\"]":                                   "needs a value",
		"[[item]]\nkind = \"key\"\npath = [\"a\"]\nvalue = []\ntype = \"string\"":    "not of the type",
	} {
		if _, err := ParseSpec([]byte(spec), "spec"); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: error %v, want %q", spec, err, want)
		}
	}
}
