package format

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The golden files in testdata/ hold the result of the merge and of the parse
// check of the bash engine (its files merge and write), from before the
// removal of that engine. Each golden file holds the inputs of its case and the
// output of bash, so the golden files cannot get old.

const prefix = "x"

// A merge case.

type mergeCase struct {
	name   string
	format string
	src    string
	live   *string // nil: the file is absent.
	remove string
	items  string
	show   string
}

// stored is a case as testdata/<kind>/<name>.golden holds it: the inputs, and
// the result of bash. One JSON file holds all bytes of a case, so no hook or
// git setting changes them.
type stored struct {
	Format string  `json:"format"`
	Show   string  `json:"show"`
	Src    string  `json:"src"`
	Live   *string `json:"live"`
	Remove string  `json:"remove"`
	Items  string  `json:"remove_items"`
	RC     int     `json:"rc"`
	Stderr string  `json:"stderr"`
	Out    *string `json:"out"`
}

func loadJSON(t testing.TB, path string, v any) {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

// goMerge runs Merge for <c> and gives what bash would give: the stderr, the
// exit code and the output.
func goMerge(t testing.TB, c mergeCase) (stderr string, rc int, out []byte) {
	t.Helper()
	ops, err := ParseOps(c.remove, c.items)
	if err != nil {
		t.Fatal(err)
	}
	var live []byte
	if c.live != nil {
		live = []byte(*c.live)
	}
	o, f := Merge([]byte(c.src), live, Format(c.format), ops, c.show)
	if f == nil {
		return "", 0, o
	}
	if f.Text != "" {
		return f.Stderr + prefix + ": " + f.Line + "\n", f.Code, nil
	}
	return f.Stderr, f.Code, nil
}

func readCase(t testing.TB, name string) (mergeCase, stored) {
	var st stored
	loadJSON(t, filepath.Join("testdata", "merge", name+".golden"), &st)
	return mergeCase{name: name, format: st.Format, show: st.Show, src: st.Src, live: st.Live, remove: st.Remove, items: st.Items}, st
}

func caseNames(t testing.TB) []string {
	entries, err := os.ReadDir(filepath.Join("testdata", "merge"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if n, ok := strings.CutSuffix(e.Name(), ".golden"); ok {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

// normalize gives <s> with the temp paths of a tool message made equal.
var (
	tmpPath = regexp.MustCompile(`'/[^'\n]*'`)
	jqAt    = regexp.MustCompile(`\(at /[^ )]*:(\d+)\)`)
)

func normalize(s string) string {
	return jqAt.ReplaceAllString(tmpPath.ReplaceAllString(s, "'<path>'"), "(at <path>:$1)")
}

func TestMergeGolden(t *testing.T) {
	names := caseNames(t)
	if len(names) == 0 {
		t.Fatal("no merge cases in testdata/merge")
	}
	equal := 0
	for _, name := range names {
		c, g := readCase(t, name)
		t.Run(name, func(t *testing.T) {
			stderr, rc, out := goMerge(t, c)
			if rc != g.RC {
				t.Errorf("exit code %d, bash gave %d\ngo stderr: %q\nbash stderr: %q", rc, g.RC, stderr, g.Stderr)
				return
			}
			if normalize(stderr) != normalize(g.Stderr) {
				t.Errorf("stderr differs\ngo:   %q\nbash: %q", stderr, g.Stderr)
				return
			}
			if (g.Out == nil) != (out == nil) || (g.Out != nil && string(out) != *g.Out) {
				t.Errorf("output differs\ngo:\n%s\nbash:\n%v", out, g.Out)
				return
			}
			equal++
		})
	}
	t.Logf("%d of %d merge cases byte-equal", equal, len(names))
}

// The parse check.

type checkCase struct {
	name, format, content string
}

type storedCheck struct {
	Format  string `json:"format"`
	Content string `json:"content"`
	RC      int    `json:"rc"`
	Stderr  string `json:"stderr"`
}

func TestParseCheckGolden(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("testdata", "check"))
	if err != nil {
		t.Fatal(err)
	}
	n, equal := 0, 0
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".golden")
		if !ok {
			continue
		}
		n++
		var st storedCheck
		loadJSON(t, filepath.Join("testdata", "check", e.Name()), &st)
		c := checkCase{name: name, format: st.Format, content: st.Content}
		wantErr := st.Stderr
		t.Run(c.name, func(t *testing.T) {
			stderr, rc := goCheck(c)
			if rc != st.RC {
				t.Errorf("exit code %d, bash gave %d", rc, st.RC)
				return
			}
			want := wantErr
			if c.format == "json" {
				// The text of a jq parse error is not copied: compare the first line.
				stderr, want = firstLine(stderr), firstLine(want)
			}
			if normalize(stderr) != normalize(want) {
				t.Errorf("stderr differs\ngo:   %q\nbash: %q", stderr, want)
				return
			}
			equal++
		})
	}
	t.Logf("%d of %d parse check cases equal", equal, n)
}

func firstLine(s string) string {
	l, _, _ := strings.Cut(s, "\n")
	return l
}

func goCheck(c checkCase) (string, int) {
	f := ParseCheck([]byte(c.content), Format(c.format), "show")
	if f == nil {
		return "", 0
	}
	return prefix + ": " + f.Line + "\n" + f.IndentedDetail(), f.Code
}

// The format of a target.

func TestDetect(t *testing.T) {
	jsonc := []byte("  // header\n{}\n")
	cases := []struct {
		target, explicit string
		content          []byte
		want             Format
	}{
		{"a.json", "", []byte("{}"), JSON},
		{"a.json.tmpl", "", []byte("{}"), JSON},
		{"a.json", "", jsonc, Text},
		{"a.json", "", []byte("{\n// a line comment\n}"), Text},
		{"a.json", "", []byte("{\n\t// a header\n}"), Text},
		{"a.json", "yaml", jsonc, YAML},
		{"a.jsonc", "", nil, Text},
		{"a.yaml", "", nil, YAML},
		{"a.yml.tmpl", "", nil, YAML},
		{"a.toml", "", nil, TOML},
		{".gemrc", "", nil, Text},
		{".gemrc", "yaml", nil, YAML},
		{"a.YAML", "", nil, Text},
		{"a.tmpl", "", nil, Text},
		{"x/a.json/b", "", nil, Text},
	}
	for _, c := range cases {
		if got := Detect(c.target, c.content, c.explicit); got != c.want {
			t.Errorf("Detect(%q, %q) = %q, want %q", c.target, c.explicit, got, c.want)
		}
	}
}
