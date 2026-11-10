package source

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The golden file testdata/toml-yq.golden holds, for each TOML file of repo/
// and macos/, the hash of the JSON that `yq -p toml -o json` gives, with the
// keys sorted. TestTOMLMatchesYq shows that ReadTOML decodes each file to the
// same JSON. To write the golden file again from yq, run:
//
//	UPDATE_GOLDEN=1 go test ./internal/source

func treeRoot(t *testing.T) string {
	t.Helper()
	d, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func tomlFiles(t *testing.T) []string {
	t.Helper()
	root := treeRoot(t)
	var out []string
	for _, side := range []string{"repo", "macos"} {
		_ = filepath.WalkDir(filepath.Join(root, side), func(p string, d fs.DirEntry, err error) error {
			// A merge spec is a template: the engine reads it after the render.
			if err == nil && d.Type().IsRegular() && strings.HasSuffix(p, ".toml") && !strings.HasSuffix(p, ".merge.toml") {
				rel, _ := filepath.Rel(root, p)
				out = append(out, filepath.ToSlash(rel))
			}
			return nil
		})
	}
	sort.Strings(out)
	return out
}

// canonical decodes JSON text and writes it again with sorted keys.
func canonical(t *testing.T, text string) string {
	t.Helper()
	var v any
	if strings.TrimSpace(text) == "null" {
		text = "{}" // yq gives null for an empty file.
	}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("bad JSON: %v\n%s", err, text)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func yqAvailable() bool {
	out, err := exec.Command("sh", "-c", "cd / && MISE_OFFLINE=1 mise ls --installed yq 2>/dev/null </dev/null").Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

func yqJSON(t *testing.T, path string) string {
	t.Helper()
	cmd := exec.Command("sh", "-c", `cd / && MISE_OFFLINE=1 mise x yq -- yq -p toml -o json -I0 "$1"`, "sh", path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("yq %s: %v", path, err)
	}
	return string(out)
}

func TestTOMLMatchesYq(t *testing.T) {
	golden := filepath.Join("testdata", "toml-yq.golden")
	files := tomlFiles(t)
	if len(files) < 100 {
		t.Fatalf("found %d TOML files, want more than 100", len(files))
	}
	root := treeRoot(t)
	var lines []string
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if !yqAvailable() {
			t.Fatal("UPDATE_GOLDEN=1 needs yq")
		}
		for _, f := range files {
			lines = append(lines, hash(canonical(t, yqJSON(t, filepath.Join(root, f))))+" "+f)
		}
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("no golden file: run UPDATE_GOLDEN=1 go test ./internal/source")
	}
	want := map[string]string{}
	for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		h, f, _ := strings.Cut(l, " ")
		want[f] = h
	}
	live := os.Getenv("BOOTSTRAP_LIVE") == "1" && yqAvailable()
	stale := 0
	for _, f := range files {
		obj, err := ReadTOML(filepath.Join(root, f))
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		got := canonical(t, JSON(obj))
		wantHash, ok := want[f]
		if !ok {
			stale++
			continue
		}
		if hash(got) != wantHash {
			// The golden can be old when the file changed: ask yq.
			if !yqAvailable() {
				stale++
				continue
			}
			if y := canonical(t, yqJSON(t, filepath.Join(root, f))); y != got {
				t.Errorf("%s: ReadTOML differs from yq\ngo: %s\nyq: %s", f, got, y)
			}
			continue
		}
		if live {
			if y := canonical(t, yqJSON(t, filepath.Join(root, f))); y != got {
				t.Errorf("%s: ReadTOML differs from yq (live)", f)
			}
		}
	}
	if stale > 0 {
		t.Skip(fmt.Sprintf("%d TOML files are not in the golden file and yq cannot run: run UPDATE_GOLDEN=1", stale))
	}
}
