package manifest

// The test harness. Each case builds a file tree in a scratch folder, runs
// the Go functions on it, then compares the exit codes, stdout, stderr and the
// file tree with the golden files in testdata/. The golden files hold the
// expected result of each case. The cases do not read the repository, so a
// change of the repository does not make a golden file old.

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/header"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/target"
)

// run is the state of one Go run of a case.
type run struct {
	root, tmp string
	env       *target.Env
	out       *bytes.Buffer // stdout of the ops
}

func (h *run) abs(rel string) string {
	if rel == "" {
		return h.root
	}
	return h.root + "/" + rel
}

// op is one step of a case. Each step prints `rc=<code>` to stdout. The
// golden files hold the expected output of each step.
type op struct {
	run func(h *run) int
}

type pcase struct {
	name  string
	dry   bool
	setup func(root string)
	ops   []op
}

// --- the ops ---

// write is the write of one target as a front-end does it: mkdir (a real
// run), the temp file, the content, then with <check> the format and the parse
// check (a failure removes the temp file and keeps the existing file), then
// the finish of the target.
func write(mode, content, dest, fileMode string, check bool, explicit string, merge bool) op {
	m, _ := strconv.ParseUint(mode, 8, 32)
	return op{run: func(h *run) int {
		e := h.env
		d := h.tmp
		path := h.abs(dest)
		show := e.Show(path)
		if !e.DryRun {
			d = target.Dirname(path)
			_ = os.MkdirAll(d, 0o777)
		}
		tmp, err := e.Temps.Create(d)
		if err != nil {
			return 99
		}
		_ = os.WriteFile(tmp, []byte(content), 0o600)
		if check {
			if _, rc := e.Check(tmp, path, explicit, merge, show); rc != 0 {
				e.Abort(tmp, path, show)
				return int(rc)
			}
		}
		return int(e.Finish(os.FileMode(m), tmp, path, show, fileMode))
	}}
}

func finish(mode, content, dest, fileMode string) op {
	return write(mode, content, dest, fileMode, false, "", false)
}

// writeReal is the write of the Mac side to the real path of <rel>.
func writeReal(mode, content, rel string) op {
	m, _ := strconv.ParseUint(mode, 8, 32)
	return op{run: func(h *run) int {
		e := h.env
		dest, ok := Real(h.abs(rel))
		if !ok {
			fmt.Fprint(h.out, "rc=loop\n")
			return -1
		}
		d := h.tmp
		if !e.DryRun {
			d = target.Dirname(dest)
			_ = os.MkdirAll(d, 0o777)
		}
		tmp, _ := e.Temps.Create(d)
		_ = os.WriteFile(tmp, []byte(content), 0o600)
		return int(e.Finish(os.FileMode(m), tmp, dest, "~/"+rel, "generated"))
	}}
}

func seedKept(dest string) op {
	return op{run: func(h *run) int {
		if target.SeedKept(h.abs(dest)) {
			return 0
		}
		return 1
	}}
}

func unblock(dir, root string, spare []string, tgt string) op {
	return op{
		run: func(h *run) int {
			return int(h.env.Unblock(h.abs(dir), h.abs(root), spare, tgt))
		},
	}
}

func hasHeader(file string) op {
	return op{run: func(h *run) int {
		if header.HasFile(h.abs(file)) {
			return 0
		}
		return 1
	}}
}

func badPath(p string) op {
	return op{run: func(h *run) int {
		if BadPath(p) {
			return 0
		}
		return 1
	}}
}

func realPath(rel string) op {
	return op{
		run: func(h *run) int {
			p, ok := Real(h.abs(rel))
			if !ok {
				return 1
			}
			fmt.Fprintf(h.out, "real=%s\n", p)
			return 0
		},
	}
}

func read(mode LinkMode, file, root string) op {
	return op{
		run: func(h *run) int {
			rows, rc := Read(h.env, mode, h.abs(file), h.abs(root))
			var s []string
			for _, r := range rows {
				s = append(s, r.String())
			}
			fmt.Fprintf(h.out, "rows=%s\n", strings.Join(s, "\n"))
			return int(rc)
		},
	}
}

func toRows(list []string) []Row {
	var rows []Row
	for _, l := range list {
		if l == "" {
			continue // an empty line is not a row
		}
		tool, path, _ := strings.Cut(l, "\t")
		rows = append(rows, Row{Tool: tool, Path: path})
	}
	return rows
}

func stale(mode LinkMode, rows, handled []string, root string) op {
	return op{
		run: func(h *run) int {
			return int(Stale(h.env, mode, toRows(rows), handled, h.abs(root)))
		},
	}
}

func writeManifest(file string, rows []string) op {
	return op{
		run: func(h *run) int {
			return int(Write(h.env, h.abs(file), toRows(rows)))
		},
	}
}

func commit(file, key, sha, comment, hdr string) op {
	return op{
		run: func(h *run) int {
			return int(h.env.RecordCommit(h.abs(file), key, sha, comment, hdr))
		},
	}
}

func prune(root, path string) op {
	return op{run: func(h *run) int {
		Prune(h.abs(root), path)
		return 0
	}}
}

// --- the runs ---

const (
	tmpName  = "tmp"
	rootName = "root"
)

// prepare makes the scratch folder of one run: <base>/root and <base>/tmp.
func (c pcase) prepare(t *testing.T) string {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{rootName, tmpName} {
		if err := os.Mkdir(filepath.Join(base, d), 0o777); err != nil {
			t.Fatal(err)
		}
	}
	if c.setup != nil {
		c.setup(base + "/" + rootName)
	}
	return base
}

type result struct{ stdout, stderr, tree string }

// String gives the golden text. A CR shows as <CR>, so no hook of the
// repository changes a golden file.
func (r result) String() string {
	cr := func(s string) string { return strings.ReplaceAll(s, "\r", "<CR>") }
	r.stdout, r.stderr = cr(r.stdout), cr(r.stderr)
	return "== stdout ==\n" + r.stdout + "== stderr ==\n" + r.stderr + "== tree ==\n" + r.tree
}

func (c pcase) runGo(t *testing.T, verbose bool) result {
	base := c.prepare(t)
	var stdout, stderr bytes.Buffer
	rep := report.NewReport(&stderr, report.Options{Verbose: verbose})
	rep.Start(c.dry)
	h := &run{
		root: base + "/" + rootName,
		tmp:  base + "/" + tmpName,
		out:  &stdout,
		env:  &target.Env{Report: rep, Stderr: &stderr, Pwd: base + "/" + rootName, DryRun: c.dry},
	}
	for _, o := range c.ops {
		if rc := o.run(h); rc >= 0 {
			fmt.Fprintf(&stdout, "rc=%d\n", rc)
		}
	}
	rep.Summary(0)
	h.env.Temps.Cleanup()
	return finishResult(t, base, stdout.String(), stderr.String())
}

var jqPos = regexp.MustCompile(`(jq: parse error: .*) at line [0-9]+, column [0-9]+`)

var tempName = regexp.MustCompile(`\.bootstrap\.[A-Za-z0-9]+`)

// finishResult normalizes the output and takes the tree of <base>.
func finishResult(t *testing.T, base, stdout, stderr string) result {
	norm := func(s string) string {
		s = strings.ReplaceAll(s, base, "@")
		return tempName.ReplaceAllString(s, ".bootstrap.*")
	}
	var tree strings.Builder
	var paths []string
	_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if p != base {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)
	for _, p := range paths {
		info, err := os.Lstat(p)
		if err != nil {
			continue
		}
		rel := norm(strings.TrimPrefix(p, base+"/"))
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			l, _ := os.Readlink(p)
			fmt.Fprintf(&tree, "l %s -> %s\n", rel, norm(l))
		case info.IsDir():
			fmt.Fprintf(&tree, "d %o %s\n", info.Mode().Perm(), rel)
			// A read-only folder of a case: open it again, so the test can
			// remove it.
			if info.Mode().Perm()&0o200 == 0 {
				_ = os.Chmod(p, 0o755)
			}
		default:
			b, _ := os.ReadFile(p)
			fmt.Fprintf(&tree, "f %o %s %q\n", info.Mode().Perm(), rel, b)
		}
	}
	return result{norm(stdout), norm(stderr), tree.String()}
}

func goldenPath(c pcase, verbose bool) string {
	v := "plain"
	if verbose {
		v = "verbose"
	}
	return filepath.Join("testdata", c.name+"."+v+".golden")
}

func TestCases(t *testing.T) {
	n := 0
	for _, c := range cases() {
		for _, verbose := range []bool{false, true} {
			path := goldenPath(c, verbose)
			t.Run(filepath.Base(path), func(t *testing.T) {
				got := c.runGo(t, verbose).String()
				if os.Getenv("UPDATE_GOLDEN") == "1" {
					if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
						t.Fatal(err)
					}
					return
				}
				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				// The position in a jq parse error is best effort in
				// internal/format (see p3-format.md).
				switch {
				case got == string(want):
					n++
				case jqPos.ReplaceAllString(got, "$1") == jqPos.ReplaceAllString(string(want), "$1"):
					n++
					t.Logf("only the position of a jq parse error differs from %s", path)
				default:
					t.Errorf("the output differs from %s\ngot:\n%s\nwant:\n%s", path, got, want)
				}
			})
		}
	}
	t.Logf("%d runs equal to the golden files", n)
}
