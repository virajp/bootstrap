// Package harness holds the common parts of the end-to-end tests of
// bootstrap: the root of the repository, the engine, the environment of a
// case, the mise shim, git, the checks and the timing lines.
//
// Each test package calls Main from its TestMain. The first case opens each
// file of the input folders, so `go test` runs the package again only when
// one of these files changes (the test cache). The module root is the root of the
// repository, so repo/, macos/ and cli/ are inside it.
package harness

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

var (
	// Root is the root of the repository.
	Root string
	// Tmp is the temporary folder of the test binary. Main removes it at the
	// end, unless HARNESS_KEEP is set.
	Tmp string
	// Bin is the engine: BOOTSTRAP_BIN, else one build of cli/.
	Bin string
	// RealMise is the path of the real mise.
	RealMise string
	// ShimDir is the folder of the mise shim. Env puts it first on PATH.
	ShimDir string

	baseEnv []string
)

// The variables that the environment of a case always reads, so that the
// test cache sees a change of one of them (os.Environ is not tracked).
var trackedVars = []string{
	"PATH", "HOME", "TMPDIR", "MISE_ENV", "BOOTSTRAP_BIN", "CLAUDE_PLUGINS_CLONE",
	"MISE_DATA_DIR", "MISE_CACHE_DIR", "XDG_DATA_HOME", "XDG_CACHE_HOME", "HARNESS_KEEP",
}

// The variables that a git hook exports. With them set, the git commands of
// the scratch repositories write to the real repository.
var gitHookVars = []string{
	"GIT_DIR", "GIT_INDEX_FILE", "GIT_WORK_TREE", "GIT_OBJECT_DIRECTORY",
	"GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_COMMON_DIR", "GIT_PREFIX", "GIT_EXEC_PATH",
	"GIT_CONFIG", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_COUNT", "GIT_NAMESPACE",
	"GIT_CEILING_DIRECTORIES", "GIT_DISCOVERY_ACROSS_FILESYSTEM", "GIT_QUARANTINE_PATH",
}

// Main prepares the package and runs its tests. <inputs> are the folders
// below the root that the tests of the package read, for example "repo".
// cli/ is always an input, because the engine is built from it. The first
// case opens them (see TrackInputs). setup, when
// not nil, runs after the engine build and before the tests.
func Main(m *testing.M, setup func() error, inputs ...string) {
	os.Exit(run(m, setup, inputs))
}

func run(m *testing.M, setup func() error, inputs []string) int {
	start := time.Now()
	if err := prepare(inputs); err != nil {
		fmt.Fprintln(os.Stderr, "harness:", err)
		return 1
	}
	if os.Getenv("HARNESS_KEEP") == "" {
		defer func() { _ = os.RemoveAll(Tmp) }()
	} else {
		fmt.Fprintln(os.Stderr, "harness: the temporary folder stays:", Tmp)
	}
	if setup != nil {
		if err := setup(); err != nil {
			fmt.Fprintln(os.Stderr, "harness: setup:", err)
			return 1
		}
	}
	fmt.Printf("TIMING package-setup %.2fs\n", time.Since(start).Seconds())
	return m.Run()
}

func prepare(inputs []string) error {
	for _, k := range trackedVars {
		_ = os.Getenv(k)
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	Root, err = findRoot(wd)
	if err != nil {
		return err
	}
	inputDirs = append([]string{"cli"}, inputs...)
	Tmp, err = os.MkdirTemp("", "bootstrap-e2e-")
	if err != nil {
		return err
	}
	RealMise, err = exec.LookPath("mise")
	if err != nil {
		return fmt.Errorf("mise is not on PATH: %w", err)
	}
	if err := writeShim(); err != nil {
		return err
	}
	baseEnv = cleanEnv(os.Environ())
	if b := os.Getenv("BOOTSTRAP_BIN"); b != "" {
		Bin = b
		return nil
	}
	Bin = filepath.Join(Tmp, "engine", "bootstrap")
	cmd := exec.Command("go", "build", "-C", filepath.Join(Root, "cli"), "-buildvcs=false", "-o", Bin, "./cmd/bootstrap")
	cmd.Env = baseEnv
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build of the engine: %v\n%s", err, out)
	}
	return nil
}

// findRoot finds the folder above <dir> that has cli/go.mod.
func findRoot(dir string) (string, error) {
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "cli", "go.mod")); err == nil {
			return d, nil
		}
		if filepath.Dir(d) == d {
			return "", fmt.Errorf("no cli/go.mod above %s", dir)
		}
	}
}

var (
	inputDirs  []string
	inputsOnce sync.Once
	inputsErr  error
)

// TrackInputs opens each folder and file below the input folders of the
// package. The test log of `go test` records each open, so a change of a
// file, or a new or removed file, runs the package again. The test log
// starts in m.Run, so Main cannot do this: Dir calls it in the first case.
func TrackInputs(t testing.TB) {
	t.Helper()
	inputsOnce.Do(func() { inputsErr = trackInputs(inputDirs) })
	if inputsErr != nil {
		t.Fatal(inputsErr)
	}
}

func trackInputs(dirs []string) error {
	for _, d := range dirs {
		err := filepath.WalkDir(filepath.Join(Root, d), func(p string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if e.Type()&fs.ModeSymlink != 0 {
				return nil
			}
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			return f.Close()
		})
		if err != nil {
			return fmt.Errorf("the inputs of %s: %w", d, err)
		}
	}
	return nil
}

// writeShim writes the mise shim. `mise x -- <command>` and
// `mise exec -- <command>` run the command directly, with the PATH of the
// test run, which has the tools of this repository. Each other call goes to
// the real mise. A `mise x` costs about 0.1 s alone and more than 1 s when
// many cases run at the same time; the shim keeps the arguments of the
// engine and of the scripts, and removes that cost.
func writeShim() error {
	// The path ends in /.local/bin/mise, because the Mac sync accepts only a
	// mise at such a path (checkMise of cli/internal/macos).
	ShimDir = filepath.Join(Tmp, "shim", ".local", "bin")
	if err := os.MkdirAll(ShimDir, 0o755); err != nil {
		return err
	}
	script := `#!/bin/bash
# The mise shim of the end-to-end tests (e2e/internal/harness).
if [ "${1:-}" = x ] || [ "${1:-}" = exec ]; then
  if [ "${2:-}" = -- ]; then
    shift 2
    exec "$@"
  fi
fi
exec ` + shellQuote(RealMise) + ` "$@"
`
	return os.WriteFile(filepath.Join(ShimDir, "mise"), []byte(script), 0o755)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// cleanEnv removes the git hook variables, the bootstrap variables and the
// usage_* variables of the caller, and puts the mise shim first on PATH.
func cleanEnv(env []string) []string {
	drop := map[string]bool{"BOOTSTRAP_URL": true, "BOOTSTRAP_CLONE": true, "BOOTSTRAP_VALUES": true,
		"BOOTSTRAP_CACHE": true, "BOOTSTRAP_HOME": true, "BOOTSTRAP_MEMBER": true}
	for _, k := range gitHookVars {
		drop[k] = true
	}
	var out []string
	for _, kv := range env {
		k, _, _ := strings.Cut(kv, "=")
		// usage_<flag> is a flag of the mise task that runs the tests (for
		// example usage_all of `mise run test --all`): a rendered task of a
		// case reads the same names.
		if drop[k] || strings.HasPrefix(k, "GIT_CONFIG_KEY_") || strings.HasPrefix(k, "GIT_CONFIG_VALUE_") ||
			strings.HasPrefix(k, "usage_") {
			continue
		}
		out = append(out, kv)
	}
	real, err := filepath.EvalSymlinks(Tmp)
	if err != nil {
		real = Tmp
	}
	trusted := Tmp + ":" + real
	if t := os.Getenv("MISE_TRUSTED_CONFIG_PATHS"); t != "" {
		trusted += ":" + t
	}
	return SetEnv(out,
		"BOOTSTRAP_VERBOSE=1",
		"BOOTSTRAP_CONFIG="+filepath.Join(Tmp, "no-config", "bootstrap", "config.toml"),
		"MISE_TRUSTED_CONFIG_PATHS="+trusted,
		"PATH="+ShimDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

// Env is a copy of the base environment of a case: the environment of the
// test run without the git hook and bootstrap variables, with
// BOOTSTRAP_VERBOSE=1, BOOTSTRAP_CONFIG at a missing file, the temporary
// folder in MISE_TRUSTED_CONFIG_PATHS, and the mise shim first on PATH.
func Env() []string {
	return append([]string(nil), baseEnv...)
}

// SetEnv sets each KEY=VALUE of <kv> in <env>. A KEY without "=" removes
// the variable.
func SetEnv(env []string, kv ...string) []string {
	for _, s := range kv {
		k, _, set := strings.Cut(s, "=")
		out := env[:0:0]
		for _, e := range env {
			if ek, _, _ := strings.Cut(e, "="); ek != k {
				out = append(out, e)
			}
		}
		if set {
			out = append(out, s)
		}
		env = out
	}
	return env
}

// GetEnv gives the value of <key> in <env>.
func GetEnv(env []string, key string) string {
	for i := len(env) - 1; i >= 0; i-- {
		if k, v, _ := strings.Cut(env[i], "="); k == key {
			return v
		}
	}
	return ""
}

// WithRealMise removes the mise shim from PATH of <env>, for a case that
// tests what the real `mise x` does.
func WithRealMise(env []string) []string {
	p := GetEnv(env, "PATH")
	p = strings.TrimPrefix(p, ShimDir+string(os.PathListSeparator))
	return SetEnv(env, "PATH="+p)
}

// Cmd is one command of a case.
type Cmd struct {
	Dir   string   // the working folder; empty is the root of the repository
	Env   []string // the environment; nil is Env()
	Stdin string   // the standard input; empty is no input
	Out   *os.File // stdout and stderr go here; nil collects them
	Args  []string // the command and its arguments
	buf   bytes.Buffer
}

// Run runs <c> and gives its exit code and, when Out is nil, its combined
// output. A command that does not start gives code 127 and the error.
func (c *Cmd) Run() (int, string) {
	cmd := exec.Command(c.Args[0], c.Args[1:]...)
	cmd.Dir = c.Dir
	if cmd.Dir == "" {
		cmd.Dir = Root
	}
	cmd.Env = c.Env
	if cmd.Env == nil {
		cmd.Env = Env()
	}
	cmd.Stdin = strings.NewReader(c.Stdin)
	if c.Out != nil {
		cmd.Stdout, cmd.Stderr = c.Out, c.Out
	} else {
		cmd.Stdout, cmd.Stderr = &c.buf, &c.buf
	}
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 127
			c.buf.WriteString(err.Error() + "\n")
		}
	}
	return code, c.buf.String()
}

// Git runs git in <dir> without the hooks, signing or identity of the user,
// and fails the test on an error. It gives the output, without the last
// newline.
func Git(t testing.TB, dir string, args ...string) string {
	t.Helper()
	out, err := GitErr(dir, args...)
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

// GitErr is Git that gives the error.
func GitErr(dir string, args ...string) (string, error) {
	all := append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "-C", dir}, args...)
	cmd := exec.Command("git", all...)
	cmd.Env = baseEnv
	out, err := cmd.CombinedOutput()
	return strings.TrimRight(string(out), "\n"), err
}

// CopyTree copies the folder <src> to <dst>, with a copy-on-write clone of
// each file where the file system has one (cp -c on APFS).
func CopyTree(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := exec.Command("cp", "-cR", src, dst).CombinedOutput()
	if err != nil {
		out, err = exec.Command("cp", "-R", src, dst).CombinedOutput()
	}
	if err != nil {
		return fmt.Errorf("cp %s %s: %v\n%s", src, dst, err, out)
	}
	return nil
}

// Dir gives a new folder below Tmp for a case, and first takes a slot for
// the case (see Slot). The test removes the folder at its end, unless it
// fails and HARNESS_KEEP is set, or HARNESS_KEEP=all.
func Dir(t testing.TB) string {
	t.Helper()
	TrackInputs(t)
	Slot(t)
	d, err := os.MkdirTemp(Tmp, "case-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		keep := os.Getenv("HARNESS_KEEP")
		if keep == "all" || (keep != "" && t.Failed()) {
			t.Logf("the folder of the case stays: %s", d)
			return
		}
		_ = os.RemoveAll(d)
	})
	return d
}

// Span starts a timing span of a step of a test. The returned function ends
// it and logs one line "TIMING <label> <seconds>s". `mise run test` keeps the
// log of the run, so a later analysis finds the slow steps.
func Span(t testing.TB, label string) func() {
	start := time.Now()
	return func() {
		t.Logf("TIMING %s %.3fs", label, time.Since(start).Seconds())
	}
}

// Pass records a passed check. The line "CHECK <name>" is in the log of the
// run, so a check that disappears shows in a comparison of two runs.
func Pass(t testing.TB, name string) {
	t.Helper()
	t.Logf("CHECK %s", name)
}

// Fail records a failed check with <detail>, for example the output of the
// case.
func Fail(t testing.TB, name, detail string) {
	t.Helper()
	if detail != "" {
		detail = "\n" + Indent(detail)
	}
	t.Errorf("FAIL %s%s", name, detail)
}

// Check records a check: Pass when <ok>, else Fail with <detail>.
func Check(t testing.TB, name string, ok bool, detail string) bool {
	t.Helper()
	if ok {
		Pass(t, name)
	} else {
		Fail(t, name, detail)
	}
	return ok
}

// Indent puts four spaces before each line of <s>.
func Indent(s string) string {
	var b strings.Builder
	sc := bufio.NewScanner(strings.NewReader(s))
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		b.WriteString("    " + sc.Text() + "\n")
	}
	return b.String()
}

// ReadFile gives the content of <path>, or "" when it cannot be read.
func ReadFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

// WriteFile writes <content> to <path>, with its folders, and fails the test
// on an error.
func WriteFile(t testing.TB, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// AppendFile appends <content> to <path>.
func AppendFile(t testing.TB, path, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
}

// Lines gives the lines of <s> (lines that "\n" ends), as printf '%s\n' of
// each argument writes them.
func Lines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// FlattenLog joins each fix line ("  -> <fix>") of <out> to its error line,
// as "<problem> - <fix>", with a lower case first letter of the fix, so a case
// names a message in one string.
func FlattenLog(out string) string {
	var b strings.Builder
	prev, have := "", false
	flush := func() {
		if have {
			b.WriteString(prev + "\n")
		}
		have = false
	}
	for _, line := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
		if strings.HasPrefix(line, "  -> ") {
			f := line[5:]
			if have {
				prev = prev + " - " + strings.ToLower(f[:min(1, len(f))]) + f[min(1, len(f)):]
			} else {
				b.WriteString(line + "\n")
			}
			continue
		}
		flush()
		prev, have = line, true
	}
	flush()
	return b.String()
}

// Gomplate renders the template <src> to <dst> with the delimiters of a
// bootstrap task (<% %>) and the context file <ctx> (empty for none).
func Gomplate(t testing.TB, src, dst, ctx string) {
	t.Helper()
	args := []string{"--left-delim", "<%", "--right-delim", "%>"}
	if ctx != "" {
		args = append(args, "-c", ".="+ctx)
	}
	args = append(args, "-f", src, "-o", dst)
	cmd := exec.Command("gomplate", args...)
	cmd.Env = baseEnv
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gomplate %s: %v\n%s", src, err, out)
	}
}

// The slots: one limit for the cases of all test packages, which `go test`
// runs in separate processes at the same time. A case runs only while it
// holds a slot: an exclusive flock on one file of the slot folder. With more
// cases than CPUs, the processes of the cases compete for the CPUs and the
// kernel, and each case takes much more time.
var (
	slotMu   sync.Mutex
	slotHeld = map[testing.TB]bool{}
)

// slotCount is the number of slots: HARNESS_SLOTS, else the number of CPUs.
func slotCount() int {
	if n, err := strconv.Atoi(os.Getenv("HARNESS_SLOTS")); err == nil && n > 0 {
		return n
	}
	return runtime.NumCPU()
}

// Slot waits for a free slot and holds it until the end of the test. A test
// that holds a slot gets no second one. The time of the wait is in the
// TIMING line "slot-wait".
func Slot(t testing.TB) {
	t.Helper()
	slotMu.Lock()
	if slotHeld[t] {
		slotMu.Unlock()
		return
	}
	slotHeld[t] = true
	slotMu.Unlock()
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("bootstrap-e2e-slots-%d", os.Getuid()))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	n := slotCount()
	start := time.Now()
	for {
		for i := 0; i < n; i++ {
			f, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("slot-%d", i)), os.O_CREATE|os.O_RDWR, 0o600)
			if err != nil {
				t.Fatal(err)
			}
			if syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) == nil {
				t.Logf("TIMING slot-wait %.3fs", time.Since(start).Seconds())
				t.Cleanup(func() {
					_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
					_ = f.Close()
					slotMu.Lock()
					delete(slotHeld, t)
					slotMu.Unlock()
				})
				return
			}
			_ = f.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}
}
