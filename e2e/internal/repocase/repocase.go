// Package repocase runs a case of the repo side: `bootstrap sync --dry-run`,
// `bootstrap check` or a call of the case, with the working tree's repo/ and
// cli/, against a scratch values folder and a scratch consumer.
//
// Each case has its own folder: a clone of the bootstrap copy, a copy of the
// consumer, a cache and a log. The cases run in parallel.
package repocase

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

var (
	// BaseClone is the bootstrap copy: the working tree's repo/ and, of
	// cli/, VERSION and go.mod (the engine reads only VERSION of a clone),
	// committed in a temporary repository on main.
	BaseClone string
	// fullClone is BaseClone with all of cli/, for a case with FullCLI. It is
	// made at the first use.
	fullClone     string
	fullCloneOnce sync.Once
	fullCloneErr  error
	// BaseConsumer is a repository with a main branch and one empty commit.
	BaseConsumer string
	// BaseCommit is the commit of BaseClone.
	BaseCommit string
)

// Main is the TestMain of a package of repo cases.
func Main(m *testing.M, inputs ...string) {
	harness.Main(m, setup, append([]string{"repo"}, inputs...)...)
}

// makeClone writes a bootstrap copy at <dir>: repo/ and either all of cli/
// or its VERSION and go.mod, committed on main.
func makeClone(dir string, fullCLI bool) error {
	if err := harness.CopyTree(filepath.Join(harness.Root, "repo"), filepath.Join(dir, "repo")); err != nil {
		return err
	}
	if fullCLI {
		if err := harness.CopyTree(filepath.Join(harness.Root, "cli"), filepath.Join(dir, "cli")); err != nil {
			return err
		}
	} else {
		for _, f := range []string{"VERSION", "go.mod"} {
			b, err := os.ReadFile(filepath.Join(harness.Root, "cli", f))
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Join(dir, "cli"), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, "cli", f), b, 0o644); err != nil {
				return err
			}
		}
	}
	// One pack in place of the loose objects: each case clones this copy,
	// and the engine copies the clone again, so fewer files cost less.
	steps := [][]string{{"init", "--quiet", "-b", "main"}, {"add", "-A"},
		{"commit", "--quiet", "--no-verify", "-m", "test copy"}, {"repack", "-a", "-d", "-q"}, {"prune-packed"}}
	for _, s := range steps {
		if out, err := harness.GitErr(dir, s...); err != nil {
			return fmt.Errorf("git %v: %v\n%s", s, err, out)
		}
	}
	return nil
}

func setup() error {
	BaseClone = filepath.Join(harness.Tmp, "bootstrap")
	if err := makeClone(BaseClone, false); err != nil {
		return err
	}
	var err error
	if BaseCommit, err = harness.GitErr(BaseClone, "rev-parse", "HEAD"); err != nil {
		return err
	}
	BaseConsumer = filepath.Join(harness.Tmp, "consumer")
	if err := os.MkdirAll(filepath.Join(BaseConsumer, ".config"), 0o755); err != nil {
		return err
	}
	for _, s := range [][]string{{"init", "--quiet", "-b", "main"}, {"commit", "--quiet", "--no-verify", "--allow-empty", "-m", "init"}} {
		if out, err := harness.GitErr(BaseConsumer, s...); err != nil {
			return fmt.Errorf("git %v: %v\n%s", s, err, out)
		}
	}
	return nil
}

// Spec is one case. It replaces `run_case` of the old bash test task: the
// fields that start with an old name do what CASE_<NAME> did.
type Spec struct {
	Name    string
	Code    int    // the expected exit code
	Values  string // one TOML text in the folder form (see WriteValues)
	Change  func(c *Case)
	Message string // a fixed string that the flattened output must have
	Check   func(c *Case)

	// Args replaces the arguments of `bootstrap sync` (default --dry-run).
	// With Cmd "check", the case runs `bootstrap check` with Args.
	Args []string
	Cmd  string
	// BootstrapURL, when set, becomes BOOTSTRAP_URL of the case.
	BootstrapURL string
	// NoClone sets no BOOTSTRAP_CLONE, so bootstrap fetches its clone into the
	// cache of the case (CASE_CLONE=none).
	NoClone bool
	// Call runs instead of bootstrap and gives the exit code (CASE_CALL). Its
	// commands write to the log of the case: use c.Engine and c.Sh.
	Call func(c *Case) int
	// Dirty does not commit the consumer before the run.
	Dirty bool
	// OldValues writes Values to the old file .config/bootstrap.toml, with
	// format = 2 and the common keys. NoFormat also leaves out format = 2.
	OldValues bool
	NoFormat  bool
	// OldForm keeps the old form of the scratch tools (no NewForm).
	OldForm bool
	// KeysAsWritten names the tools that keep their catalog keys as the case
	// writes them.
	KeysAsWritten []string
	// RealMise runs the case with the real mise, not the shim.
	RealMise bool
	// FullCLI clones a bootstrap copy with all of cli/, for a case that
	// builds or reads the engine code of the clone.
	FullCLI bool
}

// Case is the state of one running case. Change, Call and Check read and
// change its files.
type Case struct {
	T *testing.T
	// Dir is the folder of the case (tmp of the old task).
	Dir string
	// Clone is the clone of the bootstrap copy, Consumer the consumer.
	Clone, Consumer string
	// Values is the values folder; RepoValues and ToolsValues its two files;
	// OldFile the old values file .config/bootstrap.toml.
	Values, RepoValues, ToolsValues, OldFile string
	// Log is the file with the output of the run.
	Log string
	// Env is the environment of the run of the engine.
	Env  []string
	spec Spec
	out  *os.File
}

// Run runs one case as a parallel subtest of <t>.
func Run(t *testing.T, s Spec) {
	t.Helper()
	t.Run(s.Name, func(t *testing.T) {
		t.Parallel()
		runCase(t, s)
	})
}

// Check runs one case with `bootstrap check` (run_check), with no argument
// when Args is nil.
func Check(t *testing.T, s Spec) {
	t.Helper()
	s.Cmd = "check"
	if s.Args == nil {
		s.Args = []string{}
	}
	Run(t, s)
}

func runCase(t *testing.T, s Spec) {
	endSetup := harness.Span(t, "setup")
	dir := harness.Dir(t)
	c := &Case{T: t, Dir: dir, spec: s,
		Clone: filepath.Join(dir, "bootstrap"), Consumer: filepath.Join(dir, "consumer"),
		Log: filepath.Join(dir, "log")}
	c.Values = filepath.Join(c.Consumer, ".config", "bootstrap")
	c.OldFile = filepath.Join(c.Consumer, ".config", "bootstrap.toml")
	c.RepoValues = filepath.Join(c.Values, "repo.toml")
	c.ToolsValues = filepath.Join(c.Values, "tools.toml")
	base := BaseClone
	if s.FullCLI {
		fullCloneOnce.Do(func() {
			fullClone = filepath.Join(harness.Tmp, "bootstrap-full")
			fullCloneErr = makeClone(fullClone, true)
		})
		if fullCloneErr != nil {
			t.Fatal(fullCloneErr)
		}
		base = fullClone
	}
	harness.Git(t, dir, "clone", "--quiet", "--", base, c.Clone)
	if err := harness.CopyTree(BaseConsumer, c.Consumer); err != nil {
		t.Fatal(err)
	}
	if s.OldValues || s.NoFormat {
		var b strings.Builder
		if !s.NoFormat {
			b.WriteString("format = 2\n")
		}
		b.WriteString("repo_name = \"test\"\nrepo_url = \"https://example.com/test\"\n")
		b.WriteString(s.Values + "\n")
		harness.WriteFile(t, c.OldFile, b.String())
	} else {
		WriteValues(t, c.Values, s.Values)
	}
	if s.Change != nil {
		s.Change(c)
	}
	CatalogKeysAll(t, c.Clone, s.KeysAsWritten)
	if !s.OldForm {
		NewFormAll(t, c.Clone)
	}
	// The clean-target rule: a target with changes that are not committed
	// stops a sync, so the case commits its consumer first.
	if !s.Dirty {
		c.CommitConsumer()
	}
	c.Env = harness.Env()
	c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_URL", "BOOTSTRAP_CLONE",
		"BOOTSTRAP_CACHE="+filepath.Join(dir, "cache"),
		"BOOTSTRAP_VALUES="+c.Values, "MISE_PROJECT_ROOT="+c.Consumer)
	if s.BootstrapURL != "" {
		c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_URL="+s.BootstrapURL)
	}
	if !s.NoClone {
		c.Env = harness.SetEnv(c.Env, "BOOTSTRAP_CLONE="+c.Clone)
	}
	if s.RealMise {
		c.Env = harness.WithRealMise(c.Env)
	}
	endSetup()

	endRun := harness.Span(t, "run")
	f, err := os.Create(c.Log)
	if err != nil {
		t.Fatal(err)
	}
	c.out = f
	var got int
	if s.Call != nil {
		got = s.Call(c)
	} else {
		cmd := s.Cmd
		if cmd == "" {
			cmd = "sync"
		}
		args := s.Args
		if args == nil {
			args = []string{"--dry-run"}
		}
		got = c.Engine(append([]string{cmd}, args...)...)
	}
	_ = f.Close()
	c.out = nil
	endRun()

	endCheck := harness.Span(t, "check")
	defer endCheck()
	switch {
	case got != s.Code:
		harness.Fail(t, fmt.Sprintf("%s: exit %d, expected %d", s.Name, got, s.Code), c.LogText())
	case s.Message != "" && !c.LogHas(s.Message):
		harness.Fail(t, fmt.Sprintf("%s: exit %d, but the output lacks: %s", s.Name, got, s.Message), c.LogText())
	default:
		harness.Pass(t, fmt.Sprintf("%s: exit %d", s.Name, got))
		if s.Check != nil {
			s.Check(c)
		}
	}
}

// Engine runs the engine with <args> in the environment of the case, and
// gives its exit code. Its output goes to the log of the case while the case
// runs, else it is lost.
func (c *Case) Engine(args ...string) int {
	return c.Exec(append([]string{harness.Bin}, args...)...)
}

// EngineIn is Engine with the standard input <stdin>.
func (c *Case) EngineIn(stdin string, args ...string) int {
	cmd := &harness.Cmd{Env: c.Env, Stdin: stdin, Out: c.out, Args: append([]string{harness.Bin}, args...)}
	code, _ := cmd.Run()
	return code
}

// Exec runs a command in the environment of the case, in the root of the
// repository, with its output to the log of the case.
func (c *Case) Exec(args ...string) int {
	cmd := &harness.Cmd{Env: c.Env, Out: c.out, Args: args}
	code, _ := cmd.Run()
	return code
}

// Sh runs a bash script in the environment of the case, with $bin, $clone,
// $consumer, $values, $tmp and $log set, and its output to the log of the
// case. Use it only where the old case was a pipeline of other commands.
func (c *Case) Sh(script string) int {
	env := harness.SetEnv(c.Env, "bin="+harness.Bin, "clone="+c.Clone, "consumer="+c.Consumer,
		"values="+c.Values, "tmp="+c.Dir, "log="+c.Log)
	cmd := &harness.Cmd{Env: env, Out: c.out, Args: []string{"bash", "-c", "set -euo pipefail\n" + script}}
	code, _ := cmd.Run()
	return code
}

// Synced syncs and commits the consumer, so a selection command starts from
// a clean work tree. It gives the exit code of the sync; its output is lost.
func (c *Case) Synced() int {
	cmd := &harness.Cmd{Env: c.Env, Args: []string{harness.Bin, "sync"}}
	if code, _ := cmd.Run(); code != 0 {
		return code
	}
	c.CommitConsumer()
	return 0
}

// CommitConsumer commits each file of the consumer.
func (c *Case) CommitConsumer() {
	_, _ = harness.GitErr(c.Consumer, "add", "-A")
	_, _ = harness.GitErr(c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "case")
}

// LogText gives the output of the run.
func (c *Case) LogText() string { return harness.ReadFile(c.Log) }

// LogHas tells whether the flattened output has the fixed string <s>.
func (c *Case) LogHas(s string) bool {
	return strings.Contains(harness.FlattenLog(c.LogText()), s)
}

// LogLacks is the opposite of LogHas.
func (c *Case) LogLacks(s string) bool { return !c.LogHas(s) }

// CheckThat records the check <name>, with the output of the case as the
// detail of a failure (check_that).
func (c *Case) CheckThat(name string, ok bool) bool {
	c.T.Helper()
	return harness.Check(c.T, name, ok, c.LogText())
}

// Pass and Fail record a check of a custom check function.
func (c *Case) Pass(name string) { c.T.Helper(); harness.Pass(c.T, name) }

// Fail records a failed check with the output of the case.
func (c *Case) Fail(name string) { c.T.Helper(); harness.Fail(c.T, name, c.LogText()) }

// Path gives the path of <rel> in the consumer.
func (c *Case) Path(rel string) string { return filepath.Join(c.Consumer, rel) }

// ClonePath gives the path of <rel> in the clone.
func (c *Case) ClonePath(rel string) string { return filepath.Join(c.Clone, rel) }

// WriteValues writes the values folder <dir> of a case. The common keys
// (format, repo_name, repo_url, and core_branches when <body> has none) and
// the top-level keys and tables of <body> go to repo.toml; the [tools.*] and
// [overrides.*] tables, and a top-level tools or overrides key, go to
// tools.toml, after its format. A line inside a multi-line string stays with
// its key.
func WriteValues(t testing.TB, dir, body string) {
	t.Helper()
	var repo, tools strings.Builder
	repo.WriteString("format = 3\nrepo_name = \"test\"\nrepo_url = \"https://example.com/test\"\n")
	hasCore := false
	for _, l := range strings.Split(body, "\n") {
		if strings.HasPrefix(l, "core_branches") && strings.HasPrefix(strings.TrimLeft(l[len("core_branches"):], " \t"), "=") {
			hasCore = true
		}
	}
	if !hasCore {
		repo.WriteString("core_branches = [ \"main\" ]\n")
	}
	tools.WriteString("format = 3\n")
	out, ml := &repo, false
	for _, l := range strings.Split(body, "\n") {
		if !ml && out == &repo && isKey(l, "tools", "overrides") {
			tools.WriteString(l + "\n")
			continue
		}
		if !ml && isTableHeader(l) {
			if isToolsTable(l) {
				out = &tools
			} else {
				out = &repo
			}
		}
		out.WriteString(l + "\n")
		if strings.Count(l, `"""`)%2 == 1 {
			ml = !ml
		}
	}
	harness.WriteFile(t, filepath.Join(dir, "repo.toml"), repo.String())
	harness.WriteFile(t, filepath.Join(dir, "tools.toml"), tools.String())
}

// isKey: the line is a top-level key of one of <names> (/^name[[:space:]]*=/).
func isKey(l string, names ...string) bool {
	for _, n := range names {
		if strings.HasPrefix(l, n) && strings.HasPrefix(strings.TrimLeft(l[len(n):], " \t"), "=") {
			return true
		}
	}
	return false
}

// isTableHeader: /^\[\[?[A-Za-z_"]/.
func isTableHeader(l string) bool {
	r := strings.TrimPrefix(l, "[")
	if r == l {
		return false
	}
	r = strings.TrimPrefix(r, "[")
	if r == "" {
		return false
	}
	ch := r[0]
	return ch == '_' || ch == '"' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

// isToolsTable: /^\[(tools|overrides)([].]|$)/.
func isToolsTable(l string) bool {
	for _, n := range []string{"[tools", "[overrides"} {
		if rest, ok := strings.CutPrefix(l, n); ok {
			return rest == "" || rest[0] == ']' || rest[0] == '.'
		}
	}
	return false
}

// CatalogKeysAll gives each scratch tool of the clone the summary and the
// category that it lacks, before its other lines (each tool needs both). The
// tools of <keep> keep their keys as the case writes them.
func CatalogKeysAll(t testing.TB, clone string, keep []string) {
	t.Helper()
	dirs, _ := filepath.Glob(filepath.Join(clone, "repo", "tools", "*", "tools.toml"))
	for _, toml := range dirs {
		tool := filepath.Base(filepath.Dir(toml))
		if contains(keep, tool) {
			continue
		}
		catalogKeys(t, toml, tool)
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func catalogKeys(t testing.TB, toml, tool string) {
	t.Helper()
	text := harness.ReadFile(toml)
	s, c := false, false
	for _, l := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimLeft(l, " \t"), "[") {
			break
		}
		s = s || isKey(l, "summary")
		c = c || isKey(l, "category")
	}
	missing := ""
	if !s {
		missing += "summary = \"The scratch tool " + tool + ".\"\n"
	}
	if !c {
		missing += "category = \"core\"\n"
	}
	if missing != "" {
		harness.WriteFile(t, toml, missing+text)
	}
}

// NewFormAll moves the files of each scratch tool of the clone to the new
// form (NewForm).
func NewFormAll(t testing.TB, clone string) {
	t.Helper()
	dirs, _ := filepath.Glob(filepath.Join(clone, "repo", "tools", "*"))
	for _, d := range dirs {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			NewForm(t, d)
		}
	}
}

// NewForm: a case writes its scratch tools in the old form of a tool folder:
// each file at its target path, and a [files."<target>"] table only for a file
// with a mode or a condition. NewForm moves each such file below files/ (or
// tasks/ for a task), and gives it its entry: it adds `source` to the table
// of the file, or writes a new table. A `.tmpl` file becomes
// mode = "template", or render = true for a seed.
func NewForm(t testing.TB, dir string) {
	t.Helper()
	toml := filepath.Join(dir, "tools.toml")
	if _, err := os.Stat(toml); err != nil {
		return
	}
	var rels []string
	_ = filepath.WalkDir(dir, func(p string, e os.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		if rel == "tools.toml" || strings.HasPrefix(rel, "files/") || strings.HasPrefix(rel, "tasks/") || strings.HasPrefix(rel, "scripts/") {
			return nil
		}
		rels = append(rels, rel)
		return nil
	})
	sort.Strings(rels)
	for _, rel := range rels {
		target := strings.TrimSuffix(rel, ".tmpl")
		tmpl := rel != target
		var key, table, src string
		if k, ok := strings.CutPrefix(target, ".config/mise/tasks/"); ok {
			key, table = k, "tasks"
			src = strings.ReplaceAll(key, "/", "-")
			if l, ok := strings.CutPrefix(key, "_scripts/"); ok {
				src = l
			}
		} else {
			key, table, src = target, "files", target
		}
		if tmpl {
			src += ".tmpl"
		}
		dest := filepath.Join(dir, table, src)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(filepath.Join(dir, rel), dest); err != nil {
			t.Fatal(err)
		}
		lib := table == "tasks" && strings.HasPrefix(key, "_scripts/")
		h := `[files."` + target + `"]`
		lines := strings.Split(strings.TrimSuffix(harness.ReadFile(toml), "\n"), "\n")
		mode, inside := "", false
		for _, l := range lines {
			if l == h {
				inside = true
				continue
			}
			if strings.HasPrefix(l, "[") {
				inside = false
			}
			if f := strings.Fields(l); inside && len(f) >= 3 && f[0] == "mode" {
				mode = strings.ReplaceAll(f[2], `"`, "")
			}
		}
		extra := func(b *strings.Builder) {
			b.WriteString(`source = "` + src + `"` + "\n")
			if tmpl && mode == "seed" {
				b.WriteString("render = true\n")
			} else if tmpl && (mode == "" || mode == "generated") {
				b.WriteString("mode = \"template\"\n")
			}
			if lib {
				b.WriteString("library = true\n")
			}
		}
		var b strings.Builder
		found := false
		inside = false
		for _, l := range lines {
			if l == h {
				b.WriteString("[" + table + `."` + key + `"]` + "\n")
				extra(&b)
				found, inside = true, true
				continue
			}
			if strings.HasPrefix(l, "[") {
				inside = false
			}
			if f := strings.Fields(l); inside && len(f) >= 3 && f[0] == "mode" && f[2] == `"generated"` {
				continue
			}
			b.WriteString(l + "\n")
		}
		if !found {
			b.WriteString("\n[" + table + `."` + key + `"]` + "\n")
			extra(&b)
		}
		harness.WriteFile(t, toml, b.String())
	}
}

// TomlQ tells whether the yq expression <expr> is true in <file>
// (tools.toml of the case when <file> is empty), as `tomlq` did.
func (c *Case) TomlQ(expr, file string) bool {
	if file == "" {
		file = c.ToolsValues
	}
	cmd := &harness.Cmd{Env: c.Env, Args: []string{"yq", "-p", "toml", "-o", "json", expr, file}}
	code, out := cmd.Run()
	return code == 0 && strings.TrimSpace(out) == "true"
}

// FileHas tells whether the file <path> has the fixed string <s>.
func FileHas(path, s string) bool { return strings.Contains(harness.ReadFile(path), s) }

// NoFile tells whether nothing is at <path>.
func NoFile(path string) bool {
	_, err := os.Lstat(path)
	return err != nil
}

// ModeIs tells whether the permission bits of <path> are <mode>, for
// example "644".
func ModeIs(path, mode string) bool {
	fi, err := os.Stat(path)
	return err == nil && fmt.Sprintf("%o", fi.Mode().Perm()) == mode
}

// LogCount tells whether the output has <n> lines with the fixed string <s>.
func (c *Case) LogCount(s string, n int) bool {
	k := 0
	for _, l := range strings.Split(c.LogText(), "\n") {
		if strings.Contains(l, s) {
			k++
		}
	}
	return k == n
}

// CheckFile records the check <name>: the file <path> holds <content>,
// without the newlines at the end (check_file).
func (c *Case) CheckFile(name, path, content string) {
	c.T.Helper()
	b, err := os.ReadFile(path)
	if err == nil && strings.TrimRight(string(b), "\n") == content {
		harness.Pass(c.T, name)
		return
	}
	harness.Fail(c.T, name+": "+path+" is not the expected file", string(b))
}
