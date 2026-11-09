package selection

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/report"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// The golden files in testdata/ hold the output of the bash engine for each
// case: the exit code, the log and the result of the read steps (the tool
// list, the stages, the merged values), from before the removal of that
// engine. A case has a fingerprint of its inputs. When the fingerprint differs
// from the golden, the golden is old: the test skips the case, and
// test:repo-sync, test:setup-sync and test:macos-tools test that behavior.

// engineDir is sync-engine/.
func engineDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// treeRoot is the root of the working tree.
func treeRoot(t *testing.T) string {
	t.Helper()
	return filepath.Dir(engineDir(t))
}

func copyDir(t *testing.T, from, to string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-R", from, to).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v\n%s", err, out)
	}
}

// fingerprint hashes the files under <dirs>.
func fingerprint(t *testing.T, dirs ...string) string {
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
	for _, d := range dirs {
		var paths []string
		_ = filepath.WalkDir(d, func(p string, e fs.DirEntry, err error) error {
			if err == nil && !e.IsDir() {
				paths = append(paths, p)
			}
			return nil
		})
		sort.Strings(paths)
		for _, p := range paths {
			rel, _ := filepath.Rel(d, p)
			add(p, rel)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// A case is one run of the read steps against a clone and a values file.
type pcase struct {
	name   string
	side   source.Side
	clone  string // The clone copy.
	values string // The values file (repo) or the config folder (macos).
	shown  string // The values file as a message names it.
	sub    bool   // Repo: the repo is a git submodule.
	dirs   []string
}

func (c pcase) fingerprint(t *testing.T) string {
	return fingerprint(t, append([]string{c.clone, c.values}, c.dirs...)...)
}

// compose gives the text of a result.
func compose(rc int, stderr, stdout string) string {
	return fmt.Sprintf("rc=%d\n--- stderr\n%s--- stdout\n%s", rc, stderr, stdout)
}

func (c pcase) golang() string {
	var eb bytes.Buffer
	prefix := "repo-sync"
	if c.side == source.Macos {
		prefix = "setup-sync"
	}
	log := report.NewReport(&eb, report.Options{Prefix: prefix})
	rc, out := 0, ""
	if c.side == source.Macos {
		rc, out = runMacos(c, log)
	} else {
		rc, out = runRepo(c, log)
	}
	return compose(rc, eb.String(), out)
}

func runRepo(c pcase, log source.Logger) (int, string) {
	src := source.Source{Root: c.clone, Side: source.Repo}
	idx, err := src.Open()
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	v, err := values.ReadRepo(c.values, c.shown)
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	res, err := ResolveRepo(idx, v, c.sub)
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	m, err := values.Merge(values.Input{Index: idx, Selected: res.Tools, Keep: res.Keep, User: v.User,
		ShownRoot: src.ToolsShown(), ShownValues: c.shown})
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	var b strings.Builder
	b.WriteString("tools:" + prefixed(res.Tools) + "\nsel:\n")
	for _, e := range res.Sel {
		b.WriteString(e.Line() + "\n")
	}
	b.WriteString("deps:\n")
	for _, r := range res.Deps {
		b.WriteString(r.Tool + " " + strings.Join(r.Deps, " ") + "\n")
	}
	b.WriteString("values:\n" + source.JSON(m.All()) + "\n")
	return 0, b.String()
}

func prefixed(words []string) string {
	var b strings.Builder
	for _, w := range words {
		b.WriteString(" " + w)
	}
	return b.String()
}

func runMacos(c pcase, log source.Logger) (int, string) {
	src := source.Source{Root: c.clone, Side: source.Macos}
	// The order of setup_sync: the values file, the manifests, the selection.
	v, err := values.ReadMacos(filepath.Join(c.values, "macos.toml"), c.shown)
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	idx, err := src.Open()
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	res, err := ResolveMacos(idx, v)
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	m, err := values.Merge(values.Input{Index: idx, Selected: res.Tools, Keep: res.Keep, User: v.User,
		ShownRoot: src.ToolsShown(), ShownValues: "macos.toml"})
	if err != nil {
		return int(source.LogAll(log, err)), ""
	}
	var b strings.Builder
	b.WriteString("tools:" + prefixed(res.Tools) + "\ninit:" + prefixed(res.InitOrder) + "\nkeep:\n")
	b.WriteString(strings.Join(res.Keep, "\n") + "\nvalues:\n" + source.JSON(m.All()) + "\n")
	return 0, b.String()
}

// check compares the Go result of <c> with the golden file.
func check(t *testing.T, c pcase) {
	t.Helper()
	golden := filepath.Join("testdata", c.name+".golden")
	fp := c.fingerprint(t)
	got := c.golang()
	data, err := os.ReadFile(golden)
	if err != nil {
		t.Skipf("%s: no golden file, and the bash engine that wrote the golden files is gone", golden)
	}
	head, want, _ := strings.Cut(string(data), "\n")
	if head != "fingerprint: "+fp {
		t.Skipf("%s: the golden file is old (the inputs changed), and the bash engine that wrote it is gone", golden)
	}
	if got != want {
		t.Errorf("%s: Go differs from bash\n--- go\n%s\n--- bash\n%s", c.name, got, want)
	}
}

// ---- The real tree

func realClone(t *testing.T, side source.Side) string {
	t.Helper()
	d := t.TempDir()
	copyDir(t, filepath.Join(treeRoot(t), string(side)), filepath.Join(d, string(side)))
	// Examples and fixtures are not part of the clone that the sync reads.
	_ = os.RemoveAll(filepath.Join(d, string(side), "examples"))
	_ = os.RemoveAll(filepath.Join(d, string(side), "fixtures"))
	return d
}

func exampleName(root, path string) string {
	rel, _ := filepath.Rel(filepath.Join(root, "repo", "examples"), path)
	return "example-" + strings.ReplaceAll(strings.TrimSuffix(rel, "/bootstrap.toml"), "/", "-")
}

func TestRepoExamples(t *testing.T) {
	root := treeRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "repo", "examples", "*", "bootstrap.toml"))
	if err != nil {
		t.Fatal(err)
	}
	more, _ := filepath.Glob(filepath.Join(root, "repo", "examples", "*", "*", "bootstrap.toml"))
	files = append(files, more...)
	sort.Strings(files)
	if len(files) != 11 {
		t.Fatalf("found %d example values files, want 11", len(files))
	}
	for _, f := range files {
		name := exampleName(root, f)
		t.Run(name, func(t *testing.T) {
			clone := realClone(t, source.Repo)
			vdir := t.TempDir()
			copyDir(t, f, filepath.Join(vdir, "bootstrap.toml"))
			check(t, pcase{name: name, side: source.Repo, clone: clone, values: filepath.Join(vdir, "bootstrap.toml"), shown: "bootstrap.toml"})
		})
	}
}

func TestMacosExample(t *testing.T) {
	clone := realClone(t, source.Macos)
	cfg := t.TempDir()
	copyDir(t, filepath.Join(treeRoot(t), "macos", "examples", "macos.toml"), filepath.Join(cfg, "macos.toml"))
	check(t, pcase{name: "example-macos", side: source.Macos, clone: clone, values: cfg, shown: "macos.toml"})
}
