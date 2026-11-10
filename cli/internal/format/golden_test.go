package format

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The golden files in testdata/check/ hold the cases of the parse check. Each
// golden file holds the inputs of its case and the expected result.

const prefix = "x"

// normalize gives <s> with the temp paths of a tool message made equal.
var (
	tmpPath = regexp.MustCompile(`'/[^'\n]*'`)
	jqAt    = regexp.MustCompile(`\(at /[^ )]*:(\d+)\)`)
)

func normalize(s string) string {
	return jqAt.ReplaceAllString(tmpPath.ReplaceAllString(s, "'<path>'"), "(at <path>:$1)")
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
				t.Errorf("exit code %d, want %d", rc, st.RC)
				return
			}
			want := wantErr
			if c.format == "json" {
				// The text of a jq parse error is not copied: compare the first line.
				stderr, want = firstLine(stderr), firstLine(want)
			}
			if normalize(stderr) != normalize(want) {
				t.Errorf("stderr differs\ngot:  %q\nwant: %q", stderr, want)
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

func loadJSON(t testing.TB, path string, v any) {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}
