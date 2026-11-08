package manifest

// The parity harness of P4. Each case builds the same file tree in two
// scratch folders, runs the bash functions of sync-engine/write,
// sync-engine/manifest and sync-engine/common in one and the Go functions in
// the other, then compares the exit codes, stdout, stderr and the file tree.
// The golden files in testdata/ hold the bash result. TestParityGo compares
// Go with them (offline). TestParityBash compares bash with them, while the
// bash files exist (P4_NO_BASH=1 skips it). A case that runs jq or yq (through
// mise) runs live only with P4_LIVE_TOOLS=1, because it is slow. To write the
// golden files again from bash, run:
// UPDATE_GOLDEN=1 go test ./internal/manifest
// P4_SHOW_NOISE=1 prints the error lines of commands that the compare drops.

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/format"
	"github.com/virajp/bootstrap/sync-engine/internal/header"
	"github.com/virajp/bootstrap/sync-engine/internal/report"
	"github.com/virajp/bootstrap/sync-engine/internal/target"
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

// op is one step of a case: a bash snippet and the same step in Go. In bash,
// $R is the root, $DRY the dry-run flag and $TMPDIR the folder of the temp
// files of a dry run. Each step prints `rc=<code>` to stdout.
type op struct {
	bash  string
	run   func(h *run) int
	tools bool // the step runs jq or yq
}

type pcase struct {
	name  string
	dry   bool
	setup func(root string)
	ops   []op
}

func q(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

func qr(rel string) string {
	if rel == "" {
		return `"$R"`
	}
	return `"$R"/` + q(rel)
}

func lines(list []string) string {
	if len(list) == 0 {
		return "''"
	}
	return q(strings.Join(list, "\n"))
}

// --- the ops ---

// write is the write of one target as a front-end does it: mkdir (a real
// run), the temp file, the content, then with <check> the format and the parse
// check (a failure removes the temp file and keeps the existing file), then
// _se_finish.
func write(mode, content, dest, fileMode string, check bool, explicit string, merge bool) op {
	var b strings.Builder
	fmt.Fprintf(&b, "dest=%s; show=$(_se_path \"$dest\")\n", qr(dest))
	b.WriteString(`if ((DRY)); then d=$TMPDIR; else d=$(dirname -- "$dest"); mkdir -p -- "$d"; fi` + "\n")
	b.WriteString(`_se_tmp "$d" > /dev/null; tmp=$_SE_TMP` + "\n")
	fmt.Fprintf(&b, "printf %%s %s > \"$tmp\"\n", q(content))
	b.WriteString("rc=0\n")
	if check {
		file := `"$tmp"`
		if merge {
			file = "/dev/null"
		}
		fmt.Fprintf(&b, "tf=$(_se_format \"$dest\" %s %s)\n", file, q(explicit))
		b.WriteString(`_se_parse_check "$tmp" "$tf" "$show" || rc=$?` + "\n")
	}
	b.WriteString(`if ((rc != 0)); then rm -f -- "$tmp"; [[ -e $dest ]] && _se_outcome failed "$show" 'kept existing' "kept existing $show"; echo "rc=$rc"` + "\n")
	fmt.Fprintf(&b, "else _se_finish %s \"$tmp\" \"$dest\" \"$show\" \"$DRY\" %s; echo \"rc=$?\"; fi", mode, q(fileMode))
	m, _ := strconv.ParseUint(mode, 8, 32)
	return op{bash: b.String(), tools: check, run: func(h *run) int {
		e := h.env
		d := h.tmp
		path := h.abs(dest)
		show := e.Show(path)
		if !e.DryRun {
			d = target.Dirname(path)
			os.MkdirAll(d, 0o777)
		}
		tmp, err := e.Temps.Create(d)
		if err != nil {
			return 99
		}
		os.WriteFile(tmp, []byte(content), 0o600)
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
	bash := fmt.Sprintf(`if _se_manifest_real %s; then dest=$_SE_MANIFEST_REAL; show="~/"%s
  if ((DRY)); then d=$TMPDIR; else d=$(dirname -- "$dest"); mkdir -p -- "$d"; fi
  _se_tmp "$d" > /dev/null; printf %%s %s > "$_SE_TMP"
  _se_finish %s "$_SE_TMP" "$dest" "$show" "$DRY" generated; echo "rc=$?"
else echo rc=loop; fi`, qr(rel), q(rel), q(content), mode)
	m, _ := strconv.ParseUint(mode, 8, 32)
	return op{bash: bash, run: func(h *run) int {
		e := h.env
		dest, ok := Real(h.abs(rel))
		if !ok {
			fmt.Fprint(h.out, "rc=loop\n")
			return -1
		}
		d := h.tmp
		if !e.DryRun {
			d = target.Dirname(dest)
			os.MkdirAll(d, 0o777)
		}
		tmp, _ := e.Temps.Create(d)
		os.WriteFile(tmp, []byte(content), 0o600)
		return int(e.Finish(os.FileMode(m), tmp, dest, "~/"+rel, "generated"))
	}}
}

func seedKept(dest string) op {
	return op{bash: fmt.Sprintf(`_se_seed_kept %s; echo "rc=$?"`, qr(dest)), run: func(h *run) int {
		if target.SeedKept(h.abs(dest)) {
			return 0
		}
		return 1
	}}
}

func unblock(dir, root string, spare []string, tgt string) op {
	return op{
		bash: fmt.Sprintf(`_se_unblock %s %s "$DRY" %s %s; echo "rc=$?"`, qr(dir), qr(root), lines(spare), q(tgt)),
		run: func(h *run) int {
			return int(h.env.Unblock(h.abs(dir), h.abs(root), spare, tgt))
		},
	}
}

func hasHeader(file string) op {
	return op{bash: fmt.Sprintf(`_se_has_header %s; echo "rc=$?"`, qr(file)), run: func(h *run) int {
		if header.HasFile(h.abs(file)) {
			return 0
		}
		return 1
	}}
}

func badPath(p string) op {
	return op{bash: fmt.Sprintf(`_se_manifest_bad_path %s; echo "rc=$?"`, q(p)), run: func(h *run) int {
		if BadPath(p) {
			return 0
		}
		return 1
	}}
}

func realPath(rel string) op {
	return op{
		bash: fmt.Sprintf(`if _se_manifest_real %s; then echo "real=$_SE_MANIFEST_REAL"; echo rc=0; else echo rc=1; fi`, qr(rel)),
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
		bash: fmt.Sprintf(`rows=$(_se_manifest_read %s %s %s); rc=$?; printf 'rows=%%s\n' "$rows"; echo "rc=$rc"`, mode, qr(file), qr(root)),
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
			continue // bash skips an empty line
		}
		tool, path, _ := strings.Cut(l, "\t")
		rows = append(rows, Row{tool, path})
	}
	return rows
}

func stale(mode LinkMode, rows, handled []string, root string) op {
	return op{
		bash: fmt.Sprintf(`_se_manifest_stale %s %s %s %s "$DRY"; echo "rc=$?"`, mode, lines(rows), lines(handled), qr(root)),
		run: func(h *run) int {
			return int(Stale(h.env, mode, toRows(rows), handled, h.abs(root)))
		},
	}
}

func writeManifest(file string, rows []string) op {
	return op{
		bash: fmt.Sprintf(`_se_manifest_write %s %s "$DRY"; echo "rc=$?"`, qr(file), lines(rows)),
		run: func(h *run) int {
			return int(Write(h.env, h.abs(file), toRows(rows)))
		},
	}
}

func commit(file, key, sha, comment, hdr string) op {
	return op{
		bash: fmt.Sprintf(`_se_record_commit %s %s %s "$DRY" %s %s; echo "rc=$?"`, qr(file), q(key), q(sha), q(comment), q(hdr)),
		run: func(h *run) int {
			return int(h.env.RecordCommit(h.abs(file), key, sha, comment, hdr))
		},
	}
}

func prune(root, path string) op {
	return op{bash: fmt.Sprintf(`_se_manifest_prune %s %s; echo "rc=$?"`, qr(root), q(path)), run: func(h *run) int {
		Prune(h.abs(root), path)
		return 0
	}}
}

// merge is a merge target as a front-end writes it: the parse check of the
// source as <f>, then the merge with the live file, then _se_finish with the
// mode of the live file (or 644).
func merge(content, dest, f string) op {
	bash := fmt.Sprintf(`dest=%s; show=$(_se_path "$dest")
if ((DRY)); then d=$TMPDIR; else d=$(dirname -- "$dest"); mkdir -p -- "$d"; fi
_se_tmp "$d" > /dev/null; tmp=$_SE_TMP; printf %%s %s > "$tmp"; rc=0
tf=$(_se_format "$dest" /dev/null '')
_se_parse_check "$tmp" "$tf" "$show" || rc=$?
((rc != 0)) || _se_merge "$tmp" "$dest" "$tf" '[]' '[]' "$tmp" "$show" || rc=$?
if ((rc != 0)); then rm -f -- "$tmp"; [[ -e $dest ]] && _se_outcome failed "$show" 'kept existing' "kept existing $show"; echo "rc=$rc"
else if [[ -e $dest ]]; then m=$(_se_mode "$dest"); else m=644; fi
_se_finish "$m" "$tmp" "$dest" "$show" "$DRY" merge; echo "rc=$?"; fi`, qr(dest), q(content))
	return op{bash: bash, tools: true, run: func(h *run) int {
		e := h.env
		path := h.abs(dest)
		show := e.Show(path)
		d := h.tmp
		if !e.DryRun {
			d = target.Dirname(path)
			os.MkdirAll(d, 0o777)
		}
		tmp, _ := e.Temps.Create(d)
		os.WriteFile(tmp, []byte(content), 0o600)
		tf, rc := e.Check(tmp, path, "", true, show)
		if rc == 0 {
			rc = e.Merge(tmp, path, tf, format.Ops{}, show)
		}
		if rc != 0 {
			e.Abort(tmp, path, show)
			return int(rc)
		}
		m := os.FileMode(0o644)
		if _, err := os.Stat(path); err == nil {
			m, _ = target.Mode(path)
		}
		return int(e.Finish(m, tmp, path, show, "merge"))
	}}
}

// --- the runs ---

const (
	tmpName  = "tmp"
	rootName = "root"
)

func (c pcase) needsTools() bool {
	for _, o := range c.ops {
		if o.tools {
			return true
		}
	}
	return false
}

func (c pcase) script(se, base string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "source %s; source %s; source %s; source %s\n",
		q(se+"/common"), q(se+"/write"), q(se+"/merge"), q(se+"/manifest"))
	fmt.Fprintf(&b, "R=%s; DRY=%d\ncd \"$R\" || exit 99\n", q(base+"/"+rootName), b2i(c.dry))
	fmt.Fprintf(&b, "_se_outcome_start \"$DRY\"\n")
	for _, o := range c.ops {
		b.WriteString(o.bash + "\n")
	}
	b.WriteString("_se_summary 0\n")
	return b.String()
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

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

func (c pcase) runBash(t *testing.T, shell, se string, verbose bool) result {
	base := c.prepare(t)
	cmd := exec.Command(shell, "-c", c.script(se, base))
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		switch k {
		case "SYNC_ENGINE_VERBOSE", "FORCE_COLOR", "NO_COLOR", "TMPDIR", "TERM":
			continue
		}
		env = append(env, kv)
	}
	env = append(env, "NO_COLOR=1", "TERM=dumb", "TMPDIR="+base+"/"+tmpName)
	if verbose {
		env = append(env, "SYNC_ENGINE_VERBOSE=1")
	}
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", shell, err, stderr.String())
	}
	return finishResult(t, base, stdout.String(), dropNoise(stderr.String()))
}

// noise is an error line of a command or of bash itself, for example
// `mktemp: mkstemp failed on ...`. Its text depends on the system and the bash
// version, and the Go engine does not print it, so the compare drops it.
var noise = regexp.MustCompile(`^(mktemp|rm|mv|chmod|mkdir|awk|cmp|readlink|dirname): |^/\S+: line [0-9]+: `)

func dropNoise(s string) string {
	var out []string
	for _, l := range strings.SplitAfter(s, "\n") {
		if noise.MatchString(l) && os.Getenv("P4_SHOW_NOISE") == "1" {
			fmt.Fprintf(os.Stderr, "noise: %s", l)
		}
		if !noise.MatchString(l) {
			out = append(out, l)
		}
	}
	return strings.Join(out, "")
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

var tempName = regexp.MustCompile(`\.sync-engine\.[A-Za-z0-9]+`)

// finishResult normalizes the output and takes the tree of <base>.
func finishResult(t *testing.T, base, stdout, stderr string) result {
	norm := func(s string) string {
		s = strings.ReplaceAll(s, base, "@")
		return tempName.ReplaceAllString(s, ".sync-engine.*")
	}
	var tree strings.Builder
	var paths []string
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
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
				os.Chmod(p, 0o755)
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

func TestParityGo(t *testing.T) {
	n := 0
	for _, c := range cases() {
		for _, verbose := range []bool{false, true} {
			path := goldenPath(c, verbose)
			t.Run(filepath.Base(path), func(t *testing.T) {
				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				got := c.runGo(t, verbose).String()
				// The position in a jq parse error is best effort in
				// internal/format (see p3-format.md).
				switch {
				case got == string(want):
					n++
				case jqPos.ReplaceAllString(got, "$1") == jqPos.ReplaceAllString(string(want), "$1"):
					n++
					t.Logf("only the position of a jq parse error differs from %s", path)
				default:
					t.Errorf("Go differs from %s\ngot:\n%s\nwant:\n%s", path, got, want)
				}
			})
		}
	}
	t.Logf("%d runs equal to the bash golden files", n)
}

func TestParityBash(t *testing.T) {
	se, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(se, "manifest")); err != nil {
		t.Skip("no sync-engine/manifest")
	}
	if os.Getenv("P4_NO_BASH") == "1" {
		t.Skip("P4_NO_BASH=1")
	}
	var shells []string
	for _, name := range []string{"/bin/bash", "bash"} {
		if p, err := exec.LookPath(name); err == nil {
			shells = append(shells, p)
		}
	}
	if len(shells) == 0 {
		t.Skip("no bash")
	}
	_, misErr := exec.LookPath("mise")
	update := os.Getenv("UPDATE_GOLDEN") == "1"
	for _, c := range cases() {
		for _, verbose := range []bool{false, true} {
			path := goldenPath(c, verbose)
			t.Run(filepath.Base(path), func(t *testing.T) {
				t.Parallel()
				if c.needsTools() && !update && os.Getenv("P4_LIVE_TOOLS") != "1" {
					t.Skip("the case runs jq or yq: P4_LIVE_TOOLS=1 runs it")
				}
				if c.needsTools() && misErr != nil {
					t.Skip("no mise for jq and yq")
				}
				for i, shell := range shells {
					got := c.runBash(t, shell, se, verbose).String()
					if update && i == 0 {
						if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
							t.Fatal(err)
						}
					}
					want, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if got != string(want) {
						t.Errorf("%s differs from %s\ngot:\n%s\nwant:\n%s", shell, path, got, want)
					}
				}
			})
		}
	}
}
