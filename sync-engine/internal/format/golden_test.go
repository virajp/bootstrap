package format

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

// The golden files in testdata/ hold the result of the bash functions
// _se_merge, _se_parse_check and _se_format of sync-engine/merge and
// sync-engine/write. The inputs of a case are in the folder of the case, and
// the output of bash is next to them.
//
//   - UPDATE_GOLDEN=1 go test ./internal/format   writes the folders again from
//     the sources of the repository and from bash (it needs jq and yq in mise).
//   - LIVE_BASH=1 go test ./internal/format   also runs bash for each case and
//     compares it with Go. It is slow.

const prefix = "x"

// root gives the root of the repository.
func root(t testing.TB) string {
	t.Helper()
	r, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func syncDir(t testing.TB) string { return filepath.Join(root(t), "sync-engine") }

// runBash runs the bash script <script> with the arguments <args> and gives
// its stderr and its exit code.
func runBash(t testing.TB, script string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", append([]string{"-c", script, "bash"}, args...)...)
	cmd.Env = append(os.Environ(), "SYNC_ENGINE_VERBOSE=1", "NO_COLOR=1", "SE_DIR="+syncDir(t))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	rc := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		rc = ee.ExitCode()
	}
	return stderr.String(), rc
}

const mergeScript = `
source "$SE_DIR/common"; source "$SE_DIR/merge"
_SE_PREFIX=x
_se_merge "$1" "$2" "$3" "$(cat "$4")" "$(cat "$5")" "$6" "$7"
`

const checkScript = `
source "$SE_DIR/common"; source "$SE_DIR/write"
_SE_PREFIX=x
_se_parse_check "$1" "$2" "$3"
`

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

func s(x string) *string { return &x }

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

func (c mergeCase) path() string { return filepath.Join("testdata", "merge", c.name+".golden") }

func saveJSON(t testing.TB, path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
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

// bashMerge runs _se_merge for <c> and gives the stderr, the exit code and
// the output file (nil when bash wrote none).
func bashMerge(t testing.TB, c mergeCase) (stderr string, rc int, out []byte) {
	t.Helper()
	tmp := t.TempDir()
	write := func(name, data string) string {
		p := filepath.Join(tmp, name)
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	src := write("src", c.src)
	live := filepath.Join(tmp, "live-absent")
	if c.live != nil {
		live = write("live", *c.live)
	}
	remove := write("remove", c.remove)
	items := write("items", c.items)
	outPath := filepath.Join(tmp, "out")
	stderr, rc = runBash(t, mergeScript, src, live, c.format, remove, items, outPath, c.show)
	if b, err := os.ReadFile(outPath); err == nil {
		out = b
	}
	return
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
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		updateMergeGolden(t)
	}
	names := caseNames(t)
	if len(names) == 0 {
		t.Fatal("no merge cases in testdata/merge")
	}
	live := os.Getenv("LIVE_BASH") == "1"
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
			if live {
				bs, brc, bout := bashMerge(t, c)
				if brc != g.RC || normalize(bs) != normalize(g.Stderr) || (g.Out != nil && string(bout) != *g.Out) {
					t.Errorf("the golden files are old: run UPDATE_GOLDEN=1")
					return
				}
			}
			equal++
		})
	}
	t.Logf("%d of %d merge cases byte-equal", equal, len(names))
}

func updateMergeGolden(t *testing.T) {
	cases := buildMergeCases(t)
	if err := os.RemoveAll(filepath.Join("testdata", "merge")); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for _, c := range cases {
		wg.Add(1)
		sem <- struct{}{}
		go func(c mergeCase) {
			defer wg.Done()
			defer func() { <-sem }()
			stderr, rc, out := bashMerge(t, c)
			st := stored{Format: c.format, Show: c.show, Src: c.src, Live: c.live, Remove: c.remove, Items: c.items, RC: rc, Stderr: stderr}
			if rc == 0 {
				st.Out = s(string(out))
			}
			saveJSON(t, c.path(), st)
		}(c)
	}
	wg.Wait()
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

func (c checkCase) path() string { return filepath.Join("testdata", "check", c.name+".golden") }

func TestParseCheckGolden(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		updateCheckGolden(t)
	}
	entries, err := os.ReadDir(filepath.Join("testdata", "check"))
	if err != nil {
		t.Fatal(err)
	}
	live := os.Getenv("LIVE_BASH") == "1"
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
			if live {
				bs, brc := bashCheck(t, c)
				if c.format == "json" {
					bs = firstLine(bs)
				}
				if brc != st.RC || normalize(bs) != normalize(want) {
					t.Errorf("the golden files are old: run UPDATE_GOLDEN=1")
					return
				}
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

func bashCheck(t testing.TB, c checkCase) (string, int) {
	p := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(p, []byte(c.content), 0o644); err != nil {
		t.Fatal(err)
	}
	return runBash(t, checkScript, p, c.format, "show")
}

func updateCheckGolden(t *testing.T) {
	if err := os.RemoveAll(filepath.Join("testdata", "check")); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for _, c := range checkCases {
		wg.Add(1)
		sem <- struct{}{}
		go func(c checkCase) {
			defer wg.Done()
			defer func() { <-sem }()
			stderr, rc := bashCheck(t, c)
			saveJSON(t, c.path(), storedCheck{Format: c.format, Content: c.content, RC: rc, Stderr: stderr})
		}(c)
	}
	wg.Wait()
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

// TestDetectMatchesBash compares Detect with _se_format.
func TestDetectMatchesBash(t *testing.T) {
	if os.Getenv("LIVE_BASH") != "1" {
		t.Skip("set LIVE_BASH=1 to compare with bash")
	}
	script := `source "$SE_DIR/common"; source "$SE_DIR/write"; _se_format "$1" "$2" "$3"`
	for _, tc := range []struct{ target, content, explicit string }{
		{"a.json", "{}", ""}, {"a.json.tmpl", "  // x\n{}", ""}, {"a.json", "{\n// c\n}", ""},
		{"a.yml", "", ""}, {"a.toml.tmpl", "", ""}, {"a.txt", "", ""}, {".gemrc", "", "yaml"},
		{"a.json", "\t//x", ""}, {"a.json", "x // y", ""},
	} {
		p := filepath.Join(t.TempDir(), "f")
		_ = os.WriteFile(p, []byte(tc.content), 0o644)
		cmd := exec.Command("bash", "-c", script, "bash", tc.target, p, tc.explicit)
		cmd.Env = append(os.Environ(), "SE_DIR="+syncDir(t))
		out, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		want := strings.TrimSpace(string(out))
		if got := Detect(tc.target, []byte(tc.content), tc.explicit); string(got) != want {
			t.Errorf("Detect(%q) = %q, bash gave %q", tc.target, got, want)
		}
	}
}
