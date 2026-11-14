package misegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// Each case runs this package in the folder of the case, and prints the exit
// code and the log lines of each step, the written files, and the bytes and
// the mode of each written file. The golden file testdata/golden/<case>.golden
// holds the expected text. To write the golden files again from the Go
// output, run: UPDATE_GOLDEN=1 go test ./internal/misegen

func copyDir(t *testing.T, from, to string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-Rp", from, to).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v\n%s", err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// mcase is one case of the Mac side.
type mcase struct {
	name    string
	mode    string            // full, direct, own, render
	values  string            // the values file (full: config/macos.toml; else values.toml)
	noValue bool              // no values file
	tools   string            // direct and render: the tools
	files   map[string]string // more files, relative to the case folder
}

// build writes the inputs of <c> into a new folder.
func (c mcase) build(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(dir, "case")
	if err := os.MkdirAll(filepath.Join(dir, "clone", "macos", "tools"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The categories file names the category of macosTool.
	writeFile(t, filepath.Join(dir, "clone", "macos", "categories.toml"), "[category.core]\ntitle = \"Core\"\n")
	switch {
	case c.mode == "full":
		if err := os.MkdirAll(filepath.Join(dir, "config"), 0o755); err != nil {
			t.Fatal(err)
		}
		if !c.noValue {
			writeFile(t, filepath.Join(dir, "config", "macos.toml"), c.values)
		}
	case !c.noValue:
		writeFile(t, filepath.Join(dir, "values.toml"), c.values)
	}
	for p, content := range c.files {
		writeFile(t, filepath.Join(dir, p), content)
	}
	return dir
}

var (
	workRe  = regexp.MustCompile(`bootstrap-mise\.[A-Za-z0-9]+`)
	spaceRe = regexp.MustCompile(`(?m)[ \t]+$`)
)

// scrub writes the paths of one run as words that do not change. A line that
// ends with a space or a tab gets a "$" after it, and the text ends with
// "--- end": the hooks of the repository change no byte of a golden file.
func scrub(s, dir string) string {
	s = strings.ReplaceAll(s, dir, "@CASE@")
	s = workRe.ReplaceAllString(s, "bootstrap-mise.@")
	if home := os.Getenv("HOME"); home != "" {
		s = strings.ReplaceAll(s, home, "@HOME@")
	}
	s = spaceRe.ReplaceAllStringFunc(s, func(m string) string { return m + "$" })
	return s + "\n--- end\n"
}

// golang runs the case with this package in a copy of <in>.
func (c mcase) golang(t *testing.T, in string) string {
	t.Helper()
	dir := in + "-go"
	copyDir(t, in, dir)
	if err := os.MkdirAll(filepath.Join(dir, "tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := &runner{dir: dir, env: map[string]string{"USER": "tester", "TMPDIR": filepath.Join(dir, "tmp")}}
	r.macos(c)
	return scrub(r.b.String(), dir)
}

type runner struct {
	dir   string
	env   map[string]string
	b     strings.Builder
	files []string
}

func (r *runner) getenv(k string) string {
	if v, has := r.env[k]; has {
		return v
	}
	return os.Getenv(k)
}

// step prints the exit code and the messages of one step.
func (r *runner) step(name string, err error) bool {
	fmt.Fprintf(&r.b, "%s: rc=%d\n", name, fail.CodeOf(err))
	for _, e := range source.Messages(err) {
		if e.Msg != "" {
			r.b.WriteString("| mise: " + e.Msg + "\n")
		}
	}
	return err == nil
}

func (r *runner) written(list []string, prefix string) {
	for _, w := range list {
		r.b.WriteString("  " + w + "\n")
		r.files = append(r.files, prefix+w)
	}
}

func (r *runner) finish() {
	r.b.WriteString("--- files\n")
	for _, f := range r.files {
		p := filepath.Join(r.dir, f)
		st, err := os.Stat(p)
		mode := "?"
		if err == nil {
			mode = fmt.Sprintf("%o", st.Mode().Perm())
		}
		b, _ := os.ReadFile(p)
		fmt.Fprintf(&r.b, "== %s %s\n%s", f, mode, b)
	}
}

func (r *runner) macos(c mcase) {
	defer r.finish()
	clone := filepath.Join(r.dir, "clone")
	src := source.Source{Root: clone, Side: source.Macos}
	confd := filepath.Join(r.dir, "out", "conf.d")
	g := &Generator{Side: source.Macos, ToolsDir: src.ToolsDir(), Shown: src.ToolsShown(), Getenv: r.getenv}
	defer func() { _ = g.Close() }()
	var tools []string
	switch c.mode {
	case "full":
		cfg := filepath.Join(r.dir, "config")
		v, err := values.ReadMacos(filepath.Join(cfg, "macos.toml"), "macos.toml")
		if err != nil {
			fmt.Fprintf(&r.b, "pre: rc=%d\n", fail.CodeOf(err))
			return
		}
		idx, err := src.Open()
		if err != nil {
			fmt.Fprintf(&r.b, "pre: rc=%d\n", fail.CodeOf(err))
			return
		}
		res, err := selection.ResolveMacos(idx, v)
		if err != nil {
			fmt.Fprintf(&r.b, "pre: rc=%d\n", fail.CodeOf(err))
			return
		}
		m, err := values.Merge(values.Input{Index: idx, Selected: res.Tools, Keep: res.Keep, User: v.User,
			ShownRoot: src.ToolsShown(), ShownValues: "macos.toml"})
		if err != nil {
			fmt.Fprintf(&r.b, "pre: rc=%d\n", fail.CodeOf(err))
			return
		}
		r.b.WriteString("pre: rc=0\n")
		if v.Exists {
			g.Values = filepath.Join(cfg, "macos.toml")
			g.ValuesShown = "macos.toml"
		}
		g.All = m.All()
		tools = res.Tools
	default:
		if p := filepath.Join(r.dir, "values.toml"); fileExists(p) {
			g.Values, g.ValuesShown = p, "values.toml"
		}
		tools = strings.Fields(c.tools)
		if c.mode == "direct" {
			// The values come from the tools.toml files with no check: a bad
			// file is for the generator to find.
			idx, _ := src.Open()
			m, err := values.Merge(values.Input{Index: idx, Selected: tools, User: source.NewObj(),
				ShownRoot: src.ToolsShown(), ShownValues: "macos.toml"})
			if err != nil {
				r.b.WriteString("pre: rc=1\n")
				return
			}
			g.All = m.All()
		}
		r.b.WriteString("pre: rc=0\n")
	}
	if (c.mode == "full" || c.mode == "direct") && g.Values != "" {
		if !r.step("overrides", g.Overrides(tools, nil, false)) {
			return
		}
	}
	if c.mode != "own" {
		for _, tool := range tools {
			w, err := g.Render(tool, confd)
			r.step("render "+tool, err)
			r.written(w, "out/conf.d/")
		}
	}
	if c.mode == "own" || (c.mode == "full" && g.Values != "") {
		own := tools
		if c.mode == "own" {
			own = nil
		}
		w, err := g.OwnConfig(confd, own)
		r.step("own", err)
		r.written(w, "out/conf.d/")
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// check runs <c> and compares the output with the golden file.
func check(t *testing.T, c mcase) {
	t.Parallel()
	got := c.golang(t, c.build(t))
	golden := filepath.Join("testdata", "golden", c.name+".golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		writeFile(t, golden, got)
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run UPDATE_GOLDEN=1 go test ./internal/misegen)", err)
	}
	if got != string(want) {
		t.Errorf("%s: the output differs from the golden file\n--- got\n%s\n--- want\n%s", c.name, got, want)
	}
}

// run runs each case of <cases> as a parallel subtest.
func run(t *testing.T, cases []mcase) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { check(t, c) })
	}
}
