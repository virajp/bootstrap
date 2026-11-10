package misegen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/virajp/bootstrap/sync-engine/internal/header"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/selection"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// The parity test runs each case with this package in a copy of the case
// folder, and prints the exit code and the log lines of each step, the
// written files, and the bytes and the mode of each written file. The golden
// files in testdata/golden/ hold the same text of the bash engine (its file
// mise), from before the removal of that engine, and a fingerprint of the
// inputs of the case. A golden file with another fingerprint is old: the test
// skips the case, and test:repo-sync, test:setup-sync and test:macos-tools
// test that behavior.

func engineDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func treeRoot(t *testing.T) string { return filepath.Dir(engineDir(t)) }

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

// mcase is one case.
type mcase struct {
	name    string
	side    source.Side
	mode    string            // repo: full; macos: full, direct, own, render
	values  string            // the values file (repo: .config/bootstrap.toml; macos: values or config/macos.toml)
	noValue bool              // macos full: no macos.toml
	tools   string            // macos direct and render: the tools
	files   map[string]string // more files, relative to the case folder
	local   bool              // repo: the --check form of the overrides
	aliases string            // repo: the dev aliases of the tool mise
	setup   func(t *testing.T, dir string)
	real    bool // copy the real tools into the clone
}

// build writes the inputs of <c> into a new folder.
func (c mcase) build(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(dir, "case")
	side := string(c.side)
	if c.real {
		copyDir(t, filepath.Join(treeRoot(t), side), filepath.Join(dir, "clone", side))
		_ = os.RemoveAll(filepath.Join(dir, "clone", side, "examples"))
		_ = os.RemoveAll(filepath.Join(dir, "clone", side, "fixtures"))
	} else {
		if err := os.MkdirAll(filepath.Join(dir, "clone", side, "tools"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, filepath.Join(dir, "side"), side)
	writeFile(t, filepath.Join(dir, "mode"), c.mode)
	switch {
	case c.side == source.Repo:
		writeFile(t, filepath.Join(dir, "root", ".config", "bootstrap.toml"), c.values)
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
	if c.tools != "" {
		writeFile(t, filepath.Join(dir, "tools"), c.tools)
	}
	if c.local {
		writeFile(t, filepath.Join(dir, "local"), "")
	}
	if c.aliases != "" {
		writeFile(t, filepath.Join(dir, "aliases"), c.aliases)
	}
	for p, content := range c.files {
		writeFile(t, filepath.Join(dir, p), content)
	}
	if c.setup != nil {
		c.setup(t, dir)
	}
	return dir
}

// fingerprint hashes the files of the case folder.
func fingerprint(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	add := func(path, name string) {
		b, err := os.ReadFile(path)
		if err != nil {
			return
		}
		fmt.Fprintf(h, "%s %d\n", name, len(b))
		h.Write(b)
	}
	var paths []string
	_ = filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err == nil && !e.IsDir() {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)
	for _, p := range paths {
		rel, _ := filepath.Rel(dir, p)
		add(p, rel)
	}
	return hex.EncodeToString(h.Sum(nil))
}

var (
	workRe  = regexp.MustCompile(`sync-engine-mise\.[A-Za-z0-9]+`)
	spaceRe = regexp.MustCompile(`(?m)[ \t]+$`)
)

// scrub writes the paths of one run as words that do not change. A line that
// ends with a space or a tab gets a "$" after it, and the text ends with
// "--- end": the hooks of the repository change no byte of a golden file.
func scrub(s, dir string) string {
	s = strings.ReplaceAll(s, dir, "@CASE@")
	s = workRe.ReplaceAllString(s, "sync-engine-mise.@")
	if home := os.Getenv("HOME"); home != "" {
		s = strings.ReplaceAll(s, home, "@HOME@")
	}
	s = spaceRe.ReplaceAllStringFunc(s, func(m string) string { return m + "$" })
	return s + "\n--- end\n"
}

func caseEnv(dir string) map[string]string {
	return map[string]string{"USER": "tester", "TMPDIR": filepath.Join(dir, "tmp")}
}

// golang runs the case with this package in a copy of <in>.
func (c mcase) golang(t *testing.T, in string) string {
	t.Helper()
	dir := in + "-go"
	copyDir(t, in, dir)
	if err := os.MkdirAll(filepath.Join(dir, "tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := &runner{dir: dir, env: caseEnv(dir)}
	if c.side == source.Repo {
		r.repo(c)
	} else {
		r.macos(c)
	}
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

// environments reads `environments` of the values file as `_rs_envs` does,
// for a values file that has no problem.
func environments(raw *source.Obj) []string {
	list := map[string]bool{}
	if l, isList := raw.Get("environments").([]any); isList {
		for _, x := range l {
			if s, isStr := x.(string); isStr {
				list[s] = true
			}
		}
	}
	if len(list) == 0 {
		list["dev"] = true
	}
	var out []string
	for _, n := range EnvNames {
		if list[n] {
			out = append(out, n)
		}
	}
	return out
}

// hasFile is `_rs_has_file` for a root without a .gitignore: a file named
// <name> at the root or down to 4 folders below it.
func hasFile(root string) func(string) bool {
	return func(name string) bool {
		found := false
		_ = filepath.WalkDir(root, func(p string, e fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if e.IsDir() && e.Name() == ".git" {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(root, p)
			if !e.IsDir() && e.Name() == name && len(strings.Split(rel, "/")) <= 5 {
				found = true
			}
			return nil
		})
		return found
	}
}

func (r *runner) repo(c mcase) {
	defer r.finish()
	clone := filepath.Join(r.dir, "clone")
	root := filepath.Join(r.dir, "root")
	vpath := filepath.Join(root, ".config", "bootstrap.toml")
	src := source.Source{Root: clone, Side: source.Repo}
	pre := func(err error) bool {
		if err != nil {
			fmt.Fprintf(&r.b, "pre: rc=%d\n", fail.CodeOf(err))
			return false
		}
		return true
	}
	idx, err := src.Open()
	if !pre(err) {
		return
	}
	v, err := values.ReadRepo(vpath, ".config/bootstrap.toml")
	if !pre(err) {
		return
	}
	res, err := selection.ResolveRepo(idx, v, false)
	if !pre(err) {
		return
	}
	r.b.WriteString("pre: rc=0\n")
	envs := environments(v.Raw)
	e, err := NewEnvs(source.Repo, envs, res, src.ToolsShown(), hasFile(root))
	if !r.step("envs", err) {
		return
	}
	fmt.Fprintf(&r.b, "selected:%s\ntool_envs:%s\nmise_envs:%s\n", words(e.Selected).String(), e.ToolWords(), e.MiseWords())
	g := &Generator{Side: source.Repo, ToolsDir: src.ToolsDir(), Shown: src.ToolsShown(), Values: vpath,
		ValuesShown: ".config/bootstrap.toml", ProjectRoot: root, Envs: e, Getenv: r.getenv}
	defer g.Close()
	if !r.step("overrides", g.Overrides(res.Tools, envs, c.local)) || c.local {
		return
	}
	m, err := values.Merge(values.Input{Index: idx, Selected: res.Tools, Keep: res.Keep, User: v.User,
		ShownRoot: src.ToolsShown(), ShownValues: ".config/bootstrap.toml"})
	if err != nil {
		r.b.WriteString("values: failed\n")
		return
	}
	g.All = m.All()
	// The member aliases come from the derive script of mise now.
	if c.aliases != "" {
		al := source.NewObj()
		for _, line := range strings.Split(c.aliases, "\n") {
			if k, v, ok := strings.Cut(line, "="); ok {
				al.Set(k, v)
			}
		}
		mise := source.NewObj()
		mise.Set("shell_alias", al)
		d := source.NewObj()
		d.Set("mise", mise)
		g.Derived = map[string]*source.Obj{"mise": d}
	}
	for _, tool := range res.Tools {
		w, err := g.Render(tool, filepath.Join(src.ToolDir(tool), ".config", "mise", "conf.d"))
		r.step("render "+tool, err)
		r.written(w, "clone/repo/tools/"+tool+"/.config/mise/conf.d/")
	}
}

func (r *runner) macos(c mcase) {
	defer r.finish()
	clone := filepath.Join(r.dir, "clone")
	src := source.Source{Root: clone, Side: source.Macos}
	confd := filepath.Join(r.dir, "out", "conf.d")
	g := &Generator{Side: source.Macos, ToolsDir: src.ToolsDir(), Shown: src.ToolsShown(), Getenv: r.getenv}
	defer g.Close()
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
			// `_se_values` reads the tools.toml files with no check: a bad
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

// check runs <c> and compares Go with the golden file.
func check(t *testing.T, c mcase, stats *counter) {
	t.Parallel()
	in := c.build(t)
	fp := fingerprint(t, in)
	golden := filepath.Join("testdata", "golden", c.name+".golden")
	got := c.golang(t, in)
	data, err := os.ReadFile(golden)
	if err != nil {
		t.Skipf("%s: no golden file, and the bash engine that wrote the golden files is gone", golden)
	}
	head, want, _ := strings.Cut(string(data), "\n")
	if head != "fingerprint: "+fp {
		t.Skipf("%s: the golden file is old (the inputs changed), and the bash engine that wrote it is gone", golden)
	}
	if !strings.Contains(want, "pre: rc=0\n") {
		t.Fatalf("%s: the steps before the generator failed in bash:\n%s", c.name, want)
	}
	// B3 changed line 2 of the header; the golden files have the old line.
	want = strings.ReplaceAll(want, "The next sync replaces it, and a manual change is lost.", header.Line2)
	if got != want {
		t.Errorf("%s: Go differs from bash\n--- go\n%s\n--- bash\n%s", c.name, got, want)
		return
	}
	stats.add(want)
}

// counter counts the compared results of a test, for its log.
type counter struct {
	mu                  sync.Mutex
	cases, files, steps int
}

func (s *counter) add(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cases++
	for _, l := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(l, "== "):
			s.files++
		case strings.Contains(l, ": rc="):
			s.steps++
		}
	}
}

// run runs each case of <cases> as a parallel subtest, and logs the counts.
func run(t *testing.T, cases []mcase) {
	stats := &counter{}
	t.Cleanup(func() {
		t.Logf("%d cases equal to bash: %d steps (exit code and messages), %d written files (path, mode, bytes)", stats.cases, stats.steps, stats.files)
	})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { check(t, c, stats) })
	}
}
