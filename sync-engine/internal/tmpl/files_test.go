package tmpl

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// The archive testdata/renders.tar.gz holds, for each render of a template by
// the bash engine, the inputs and the output of the gomplate CLI. For a
// template file (folder c.*): the source path, the JSON context, the flag
// --missing-key and the output. For an inline template (folder i.*): the
// template, the files of the -c flags, the environment variables of the engine,
// the output and the error. The script testdata/capture.sh writes it from the
// bash engine. To write it again, run: UPDATE_GOLDEN=1 go test ./internal/tmpl
// The scratch folder of the capture is @TMP@ in the files.

const rendersFile = "testdata/renders.tar.gz"

// render is one folder of the archive: the file name to its content.
type render struct {
	name  string
	files map[string]string
}

func (r render) get(n string) string { return r.files[n] }

var loadRenders = sync.OnceValues(func() ([]render, error) {
	f, err := os.Open(rendersFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(zr)
	byDir := map[string]*render{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		name := strings.TrimPrefix(h.Name, "./")
		parts := strings.SplitN(name, "/", 3) // group/case/file
		if len(parts) < 3 {
			continue
		}
		key := parts[0] + "/" + parts[1]
		r := byDir[key]
		if r == nil {
			r = &render{name: key, files: map[string]string{}}
			byDir[key] = r
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		r.files[parts[2]] = string(b)
	}
	var out []render
	for _, r := range byDir {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, nil
})

// repoRoot is the root of the bootstrap repository.
func repoRoot(t testing.TB) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestUpdateFilesGolden(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") != "1" {
		t.Skip("set UPDATE_GOLDEN=1 to write the golden files")
	}
	root := repoRoot(t)
	cmd := exec.Command("bash", filepath.Join(root, "sync-engine", "internal", "tmpl", "testdata", "capture.sh"),
		filepath.Join(root, "sync-engine", "internal", "tmpl", rendersFile))
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

// fileCases returns the renders of template files.
func fileCases(t testing.TB) []render {
	t.Helper()
	all, err := loadRenders()
	if err != nil {
		t.Fatalf("%v (run UPDATE_GOLDEN=1 go test ./internal/tmpl)", err)
	}
	var out []render
	for _, r := range all {
		if strings.HasPrefix(path.Base(r.name), "c.") {
			out = append(out, r)
		}
	}
	return out
}

// scratch returns a folder with the links that the contexts name: the
// sync_engine value is @TMP@/bootstrap/sync-engine or @TMP@/mtree/sync-engine.
func scratch(t testing.TB) string {
	t.Helper()
	tmp := t.TempDir()
	root := repoRoot(t)
	for _, n := range []string{"bootstrap", "mtree"} {
		if err := os.Symlink(root, filepath.Join(tmp, n)); err != nil {
			t.Fatal(err)
		}
	}
	return tmp
}

func renderFileCase(t testing.TB, root, tmp string, c render) ([]byte, error) {
	t.Helper()
	srcPath := c.get("src")
	src, err := os.ReadFile(filepath.Join(root, srcPath))
	if err != nil {
		t.Fatalf("%s: %v", c.name, err)
	}
	ctx, err := DecodeJSON([]byte(strings.ReplaceAll(c.get("ctx.json"), "@TMP@", tmp)))
	if err != nil {
		t.Fatalf("%s: %v", c.name, err)
	}
	return Render(srcPath, src, ctx, &Options{MissingKey: c.get("missingkey")})
}

// TestTemplateFiles renders each captured template file and compares the
// bytes with the output of gomplate.
func TestTemplateFiles(t *testing.T) {
	cases := fileCases(t)
	if len(cases) == 0 {
		t.Fatal("no captured renders: run UPDATE_GOLDEN=1 go test ./internal/tmpl")
	}
	root := repoRoot(t)
	tmp := scratch(t)
	equal := map[string]int{}
	for _, c := range cases {
		got, err := renderFileCase(t, root, tmp, c)
		src := c.get("src")
		if c.get("rc") != "0" {
			if err == nil {
				t.Errorf("%s (%s): gomplate failed, Go did not", c.name, src)
			}
			continue
		}
		want := []byte(strings.ReplaceAll(c.get("out"), "@TMP@", tmp))
		if err != nil {
			t.Errorf("%s (%s): %v", c.name, src, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s (%s): output differs\n--- want\n%s\n--- got\n%s", c.name, src, want, got)
			continue
		}
		equal[src]++
	}
	// The template files of the repository: each must be in the captures.
	var missing []string
	n := 0
	for _, base := range []string{"repo", "macos"} {
		_ = filepath.WalkDir(filepath.Join(root, base), func(p string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(p, ".tmpl") {
				n++
				rel, _ := filepath.Rel(root, p)
				if equal[rel] == 0 {
					missing = append(missing, rel)
				}
			}
			return nil
		})
	}
	t.Logf("%d of %d templates (.tmpl files) byte-equal, in %d contexts; %d of the captured files are tools.toml files",
		n-len(missing), n, len(cases), countSuffix(equal, "tools.toml"))
	for _, m := range missing {
		t.Errorf("no captured render of %s", m)
	}
}

func countSuffix(m map[string]int, suffix string) int {
	n := 0
	for k := range m {
		if strings.HasSuffix(k, suffix) {
			n++
		}
	}
	return n
}
