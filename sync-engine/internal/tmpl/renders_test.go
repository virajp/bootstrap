package tmpl

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// The archive testdata/renders.tar.gz holds inline templates (folder i.*) with
// the output of the gomplate CLI for each one: the template, the files of the
// -c flags, the environment variables, the output and the error. The scratch
// folder of the capture is @TMP@ in the files.

const rendersFile = "testdata/renders.tar.gz"

// render is one folder of the archive: the file name to its content.
type render struct {
	name  string
	files map[string]string
}

func (r render) get(n string) string { return r.files[n] }

var loadRenders = sync.OnceValues(readRenders)

func readRenders() ([]render, error) {
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
}

// repoRoot is the root of the bootstrap repository.
func repoRoot(t testing.TB) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
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
