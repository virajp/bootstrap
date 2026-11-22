package mac

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

var (
	// realTmp is the real path of harness.Tmp. The Mac sync logs the real
	// path of a target behind a symlink.
	realTmp string
	// fixtures is macos/fixtures of the repository.
	fixtures string
	// tree is cli/ and the fixture repo as macos/, in a git repository.
	tree       string
	treeCommit string
	// The real mise folders of the user, so that a real `mise x` finds its
	// tools with no download.
	miseData, miseCache string
	// The entry task and its origin.
	entryTool, entryTask, entrySrc, entryURL, entryHead string
)

// gitCmd runs git in <dir> with <env>, without the config of the user, the
// hooks, signing or identity (the _git function of the old task).
func gitCmd(env []string, dir string, args ...string) (string, error) {
	all := append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "-c", "init.defaultBranch=main"}, args...)
	cmd := exec.Command("git", all...)
	cmd.Dir = dir
	cmd.Env = harness.SetEnv(env, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	return strings.TrimRight(string(out), "\n"), err
}

// setup checks the fixture files, and makes the tree and the origin of the
// entry cases once.
func setup() error {
	var err error
	if realTmp, err = filepath.EvalSymlinks(harness.Tmp); err != nil {
		return err
	}
	fixtures = filepath.Join(harness.Root, "macos", "fixtures")
	env := harness.Env()

	// No fixture file may be one that git ignores: it is not in a fresh
	// checkout.
	var files []string
	err = filepath.WalkDir(fixtures, func(p string, e fs.DirEntry, err error) error {
		if err == nil && e.Type().IsRegular() {
			rel, _ := filepath.Rel(harness.Root, p)
			files = append(files, rel)
		}
		return err
	})
	if err != nil {
		return err
	}
	ci := exec.Command("git", "check-ignore", "--no-index", "--stdin")
	ci.Dir = harness.Root
	ci.Env = env
	ci.Stdin = strings.NewReader(strings.Join(files, "\n") + "\n")
	if out, _ := ci.Output(); strings.TrimSpace(string(out)) != "" {
		return fmt.Errorf("fixture files that .gitignore matches:\n%s", harness.Indent(string(out)))
	}

	miseData = os.Getenv("MISE_DATA_DIR")
	if miseData == "" {
		d := os.Getenv("XDG_DATA_HOME")
		if d == "" {
			d = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
		miseData = filepath.Join(d, "mise")
	}
	miseCache = os.Getenv("MISE_CACHE_DIR")
	if miseCache == "" {
		d := os.Getenv("XDG_CACHE_HOME")
		if d == "" {
			d = filepath.Join(os.Getenv("HOME"), ".cache")
		}
		miseCache = filepath.Join(d, "mise")
	}

	// The tree: cli/ of the working tree and the fixture repo as macos/,
	// without tools/bad/ (one case adds it), committed.
	tree = filepath.Join(realTmp, "tree")
	if err := harness.CopyTree(filepath.Join(harness.Root, "cli"), filepath.Join(tree, "cli")); err != nil {
		return err
	}
	if err := harness.CopyTree(filepath.Join(fixtures, "repo"), filepath.Join(tree, "macos")); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(tree, "macos", "tools", "bad")); err != nil {
		return err
	}
	for _, s := range [][]string{{"init", "--quiet", "-b", "main"}, {"add", "-A"}, {"commit", "--quiet", "--no-verify", "-m", "fixture"}} {
		if out, err := gitCmd(env, tree, s...); err != nil {
			return fmt.Errorf("git %v: %v\n%s", s, err, out)
		}
	}
	if treeCommit, err = gitCmd(env, tree, "rev-parse", "HEAD"); err != nil {
		return err
	}

	// The entry task of the Mac tool mise, and the origin of the clone of
	// bootstrap that the engine fetches: the fixture tree.
	entryTool = filepath.Join(harness.Root, "macos", "tools", "mise")
	entryTask = filepath.Join(realTmp, "updateall-configs.py")
	if err := copyFile(filepath.Join(entryTool, "tasks", "updateall-configs.py"), entryTask); err != nil {
		return err
	}
	if err := os.Chmod(entryTask, 0o755); err != nil {
		return err
	}
	entrySrc = filepath.Join(realTmp, "entry-src")
	if err := harness.CopyTree(tree, entrySrc); err != nil {
		return err
	}
	entryURL = "file://" + realTmp + "/entry-origin.git"
	if out, err := gitCmd(env, realTmp, "clone", "--quiet", "--bare", "--", entrySrc, filepath.Join(realTmp, "entry-origin.git")); err != nil {
		return fmt.Errorf("git clone: %v\n%s", err, out)
	}
	entryHead, err = gitCmd(env, entrySrc, "rev-parse", "HEAD")
	return err
}

// copyFile copies <src> to <dst> with the mode of <src>.
func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, fi.Mode().Perm())
}

// mcase is one case: the paths of the variables of case_body, its
// environment and the state of its last run.
type mcase struct {
	t                                 *testing.T
	name                              string
	dir, home, user, cfg, values      string
	cache, manifest, xdgCache, xdgCfg string
	tmpdir                            string
	clone, work                       string
	repo, lock, oldCache              string // the entry cases
	env                               []string
	log                               string
	rc, runs                          int
}

// runCase is run_case: <body> runs in parallel, in its own folder and
// environment.
func runCase(t *testing.T, name string, body func(c *mcase)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		c := newCase(t, name)
		defer harness.Span(t, "case")()
		body(c)
	})
}

// newCase is case_body: the isolated environment of a case.
func newCase(t *testing.T, name string) *mcase {
	d, err := filepath.EvalSymlinks(harness.Dir(t))
	if err != nil {
		t.Fatal(err)
	}
	c := &mcase{t: t, name: name, dir: d, clone: tree}
	c.home = filepath.Join(d, "home")
	c.user = filepath.Join(d, "user")
	c.xdgCache = filepath.Join(d, "xdg", "cache")
	c.xdgCfg = filepath.Join(d, "xdg", "config")
	c.tmpdir = filepath.Join(d, "tmp")
	c.cache = filepath.Join(c.xdgCache, "bootstrap")
	// The fixed folder of macos.toml: $XDG_CONFIG_HOME/bootstrap.
	c.cfg = filepath.Join(c.xdgCfg, "bootstrap")
	c.values = filepath.Join(c.cfg, "macos.toml")
	c.manifest = filepath.Join(c.cache, "macos-generated.tsv")
	for _, p := range []string{c.home, c.xdgCache, c.xdgCfg, filepath.Join(d, "xdg", "state"),
		filepath.Join(d, "xdg", "data"), c.tmpdir, c.user, c.cfg} {
		c.mkdir(p)
	}
	gitconfig := filepath.Join(d, "gitconfig")
	harness.WriteFile(t, gitconfig, "")
	c.env = harness.SetEnv(harness.Env(),
		"HOME="+c.home,
		"XDG_CACHE_HOME="+c.xdgCache, "XDG_CONFIG_HOME="+c.xdgCfg,
		"XDG_STATE_HOME="+filepath.Join(d, "xdg", "state"), "XDG_DATA_HOME="+filepath.Join(d, "xdg", "data"),
		"BOOTSTRAP_CACHE="+c.cache,
		"GIT_CONFIG_GLOBAL="+gitconfig, "GIT_CONFIG_NOSYSTEM=1",
		"TMPDIR="+c.tmpdir,
		"MISE_DATA_DIR="+miseData, "MISE_CACHE_DIR="+miseCache,
		"BOOTSTRAP_BIN="+harness.Bin,
		"BOOTSTRAP_URL", "BOOTSTRAP_CLONE", "BOOTSTRAP_VALUES", "MACOS_SETUP_REPO_PATH", "MISE_ENV",
	)
	c.git("init", "--quiet", c.cfg)
	return c
}

// setEnv sets or removes variables of the environment of the case.
func (c *mcase) setEnv(kv ...string) { c.env = harness.SetEnv(c.env, kv...) }

// prependPath puts <dir> first on PATH of the case.
func (c *mcase) prependPath(dir string) {
	c.setEnv("PATH=" + dir + string(os.PathListSeparator) + harness.GetEnv(c.env, "PATH"))
}

// git runs git with the environment of the case and fails the test on an
// error.
func (c *mcase) git(args ...string) string {
	c.t.Helper()
	out, err := gitCmd(c.env, c.dir, args...)
	if err != nil {
		c.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

// ---- the runs ------------------------------------------------------------

// execIn runs <args> in <dir> with <env>, with its output in a new log
// file of the case, and keeps its exit code.
func (c *mcase) execIn(dir string, env []string, args ...string) {
	c.runs++
	c.log = filepath.Join(c.dir, fmt.Sprintf("run-%d.txt", c.runs))
	f, err := os.Create(c.log)
	if err != nil {
		c.t.Fatal(err)
	}
	cmd := &harness.Cmd{Dir: dir, Env: env, Out: f, Args: args}
	c.rc, _ = cmd.Run()
	_ = f.Close()
}

// sync is sync_run: the engine as its entry task runs it, from the case
// folder. <extra> are variables for this run only.
func (c *mcase) syncWith(extra []string, args ...string) {
	env := harness.SetEnv(c.env, append([]string{"BOOTSTRAP_CLONE=" + c.clone}, extra...)...)
	c.execIn(c.dir, env, append([]string{harness.Bin, "mac", "sync"}, args...)...)
}
func (c *mcase) runReal() { c.syncWith(nil, "--no-bootstrap") }
func (c *mcase) runDry()  { c.syncWith(nil, "--dry-run") }

// mac is mac_run: `bootstrap mac <args>` in the config folder, without a
// terminal.
func (c *mcase) mac(args ...string) {
	env := harness.SetEnv(c.env, "BOOTSTRAP_CLONE="+c.clone)
	c.execIn(c.cfg, env, append([]string{harness.Bin, "mac"}, args...)...)
}

// ---- the checks ----------------------------------------------------------

func (c *mcase) pass(what string) { c.t.Helper(); harness.Pass(c.t, c.name+": "+what) }
func (c *mcase) fail(what string) {
	c.t.Helper()
	harness.Fail(c.t, c.name+": "+what, harness.ReadFile(c.log))
}

// check records <what>: pass when <ok>, else fail.
func (c *mcase) check(what string, ok bool) {
	c.t.Helper()
	if ok {
		c.pass(what)
	} else {
		c.fail(what)
	}
}

// exit is check_exit.
func (c *mcase) exit(want int) {
	c.t.Helper()
	if c.rc == want {
		c.pass(fmt.Sprintf("exit %d", c.rc))
	} else {
		c.fail(fmt.Sprintf("exit %d, expected %d", c.rc, want))
	}
}

func (c *mcase) flat() string { return harness.FlattenLog(harness.ReadFile(c.log)) }

// has is check_output.
func (c *mcase) has(s string) {
	c.t.Helper()
	c.check("the output has: "+s, strings.Contains(c.flat(), s))
}

// lacks is check_no_output.
func (c *mcase) lacks(s string) {
	c.t.Helper()
	c.check("the output lacks: "+s, !strings.Contains(c.flat(), s))
}

// file is check_file: the file exists with exactly <content>, without the
// newlines at the end.
func (c *mcase) file(path, content string) {
	c.t.Helper()
	c.check(strings.TrimPrefix(path, c.dir+"/")+" has the expected content", fileIs(path, content))
}

// same is check_same: the file holds exactly <content>, byte for byte.
func (c *mcase) same(path, content string) {
	c.t.Helper()
	b, err := os.ReadFile(path)
	c.check(strings.TrimPrefix(path, c.dir+"/")+" is kept byte-equal", err == nil && isFile(path) && string(b) == content)
}

// logHasRe tells whether a line of the log matches <re>.
func (c *mcase) logHasRe(re string) bool { return countRe(c.log, re) > 0 }

// ---- the file predicates -------------------------------------------------

func exists(p string) bool { _, err := os.Stat(p); return err == nil }
func isFile(p string) bool { fi, err := os.Stat(p); return err == nil && fi.Mode().IsRegular() }
func isDir(p string) bool  { fi, err := os.Stat(p); return err == nil && fi.IsDir() }
func isLink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&fs.ModeSymlink != 0
}

// catT is $(cat <p>): the content without the newlines at the end.
func catT(p string) string { return strings.TrimRight(harness.ReadFile(p), "\n") }

func fileIs(p, content string) bool { return isFile(p) && catT(p) == content }

// lines gives the lines of the file <p>, as grep reads them.
func lines(p string) []string {
	s := harness.ReadFile(p)
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// hasLine is grep -qxF.
func hasLine(p, line string) bool { return countLine(p, line) > 0 }

func countLine(p, line string) int {
	n := 0
	for _, l := range lines(p) {
		if l == line {
			n++
		}
	}
	return n
}

// hasStr is grep -qF.
func hasStr(p, s string) bool { return strings.Contains(harness.ReadFile(p), s) }

// countRe is grep -cE.
func countRe(p, re string) int {
	r := regexp.MustCompile(re)
	n := 0
	for _, l := range lines(p) {
		if r.MatchString(l) {
			n++
		}
	}
	return n
}

func hasRe(p, re string) bool { return countRe(p, re) > 0 }

// firstLine is head -n 1.
func firstLine(p string) string {
	if l := lines(p); len(l) > 0 {
		return l[0]
	}
	return ""
}

// cmpFiles is cmp -s.
func cmpFiles(a, b string) bool {
	x, err1 := os.ReadFile(a)
	y, err2 := os.ReadFile(b)
	return err1 == nil && err2 == nil && string(x) == string(y)
}

// modeOf is stat -f %Lp.
func modeOf(p string) string {
	fi, err := os.Lstat(p)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%o", fi.Mode().Perm())
}

// listOf gives the list of the key <key> of a JSON file, without spaces
// and newlines (tr -d ' \n' | sed 's/.*"<key>":\(\[[^]]*\]\).*/\1/').
func listOf(p, key string) string {
	s := strings.NewReplacer(" ", "", "\n", "").Replace(harness.ReadFile(p))
	m := regexp.MustCompile(`.*"` + key + `":(\[[^]]*\]).*`).FindStringSubmatch(s)
	if m == nil {
		return s
	}
	return m[1]
}

// snapshot gives the name, mode, size and time of each path below
// <roots>, and the hash of each regular file, sorted.
func snapshot(roots ...string) string {
	var rows []string
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(p string, e fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			fi, err := os.Lstat(p)
			if err != nil {
				return nil
			}
			row := fmt.Sprintf("%s %s %d %d", p, fi.Mode(), fi.Size(), fi.ModTime().Unix())
			if fi.Mode().IsRegular() {
				b, _ := os.ReadFile(p)
				row += fmt.Sprintf(" %x", sha1.Sum(b))
			}
			rows = append(rows, row)
			return nil
		})
	}
	sort.Strings(rows)
	return strings.Join(rows, "\n")
}

// ---- the changes of a case -----------------------------------------------

func (c *mcase) mkdir(p string) {
	c.t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		c.t.Fatal(err)
	}
}

func (c *mcase) write(p, content string) { c.t.Helper(); harness.WriteFile(c.t, p, content) }

// touch is `: > <p>`.
func (c *mcase) touch(p string) { c.t.Helper(); c.write(p, "") }

func (c *mcase) chmod(p string, mode os.FileMode) {
	c.t.Helper()
	if err := os.Chmod(p, mode); err != nil {
		c.t.Fatal(err)
	}
}

func (c *mcase) symlink(src, link string) {
	c.t.Helper()
	if err := os.Symlink(src, link); err != nil {
		c.t.Fatal(err)
	}
}

func (c *mcase) remove(p string) {
	c.t.Helper()
	if err := os.RemoveAll(p); err != nil {
		c.t.Fatal(err)
	}
}

// useConfig is use_config: a copy of a fixture config folder; "-" is an
// empty one.
func (c *mcase) useConfig(name string) {
	c.t.Helper()
	c.remove(c.cfg)
	if name == "-" {
		c.mkdir(c.cfg)
	} else if err := harness.CopyTree(filepath.Join(fixtures, "config", name), c.cfg); err != nil {
		c.t.Fatal(err)
	}
	c.git("init", "--quiet", c.cfg)
	c.commitConfig()
}

// commitConfig commits each file of the config folder.
func (c *mcase) commitConfig() {
	c.t.Helper()
	c.git("-C", c.cfg, "add", "-A")
	c.git("-C", c.cfg, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "config")
}

// valuesFile is values_file: a macos.toml of the new form, with the tables
// of the required members core and mise, then <ls>.
func (c *mcase) valuesFile(ls ...string) {
	c.write(c.values, harness.Lines(append([]string{"format = 2", "", "[tools.core]", "", "[tools.mise]", ""}, ls...)...))
}

// caseTree is case_tree: a copy of the tree for this case only, after
// <change> changed its macos/ folder, then the new form of each tool.
func (c *mcase) caseTree(change func(c *mcase)) {
	c.t.Helper()
	c.clone = filepath.Join(c.dir, "tree")
	if err := harness.CopyTree(tree, c.clone); err != nil {
		c.t.Fatal(err)
	}
	c.work = filepath.Join(c.clone, "macos")
	change(c)
	dirs, _ := filepath.Glob(filepath.Join(c.work, "tools", "*"))
	sort.Strings(dirs)
	for _, d := range dirs {
		if isDir(d) {
			rc.NewForm(c.t, d)
		}
	}
}

// seedManifest is seed_manifest: a manifest from an earlier run.
func (c *mcase) seedManifest(rows ...string) { c.write(c.manifest, harness.Lines(rows...)) }

// oldHeaderFile writes a file with the header of the old Mac sync, which the
// sync no longer counts as a header.
func (c *mcase) oldHeaderFile(p, line string) {
	c.write(p, harness.Lines("# Generated by macos-setup: test", line))
}

func (c *mcase) newHeaderFile(p, line string) {
	c.write(p, harness.Lines("# Synced from bootstrap: test. Do not edit this file.", line))
}

// toolsToml is the tools.toml of <tool> in the tree of the case.
func (c *mcase) toolsToml(tool string) string {
	return filepath.Join(c.work, "tools", tool, "tools.toml")
}

// appendPacks appends a blank line and <ls> to macos/packs.toml.
func (c *mcase) appendPacks(ls ...string) {
	harness.AppendFile(c.t, filepath.Join(c.work, "packs.toml"), "\n"+harness.Lines(ls...))
}

// prependKey is prepend_key: a top-level line before the tables of a
// tools.toml.
func (c *mcase) prependKey(tool, line string) {
	p := c.toolsToml(tool)
	c.write(p, line+"\n"+catT(p)+"\n")
}

// appendLines is append_lines: lines at the end of a tools.toml. When the
// tools.toml has the first line already, a table header, the other lines go
// below that header instead.
func (c *mcase) appendLines(tool string, ls ...string) {
	p := c.toolsToml(tool)
	if hasLine(p, ls[0]) {
		var b strings.Builder
		for _, l := range lines(p) {
			b.WriteString(l + "\n")
			if l == ls[0] {
				b.WriteString(strings.Join(ls[1:], "\n") + "\n")
			}
		}
		c.write(p, b.String())
		return
	}
	harness.AppendFile(c.t, p, "\n"+harness.Lines(ls...))
}

// dropKey is drop_key: a tools.toml without its top-level line of <key>.
func (c *mcase) dropKey(tool, key string) {
	p := c.toolsToml(tool)
	var b strings.Builder
	tables := false
	for _, l := range lines(p) {
		if strings.HasPrefix(strings.TrimLeft(l, " \t"), "[") {
			tables = true
		}
		if f := strings.Fields(l); !tables && len(f) >= 2 && f[0] == key && f[1] == "=" {
			continue
		}
		b.WriteString(l + "\n")
	}
	c.write(p, b.String())
}

// setKey is set_key: <line> in place of the top-level line of <key>.
func (c *mcase) setKey(tool, key, line string) {
	c.dropKey(tool, key)
	c.prependKey(tool, line)
}

// synced is the header of a synced file of the Mac tool file <rel>.
func synced(rel string) string {
	return "# Synced from bootstrap: macos/tools/" + rel + ". Do not edit this file.\n" +
		"# The next sync writes it again, and each change in this file is lost."
}
