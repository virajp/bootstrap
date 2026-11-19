// Package tasks tests the parts of test:repo that ran one after the other,
// outside the case pool: the entry task setup:configs with a fake engine, the
// clone and the version check of the engine, the gate code:lint:60-bootstrap,
// gate_files.py and the orchestrator code:lint. The cases came from test:repo,
// lines 5572 to 6214.
package tasks

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

func TestMain(m *testing.M) { rc.Main(m) }

// helpersPath is the helpers.py library of the working tree's repo/.
func helpersPath() string {
	return filepath.Join(harness.Root, "repo", "tools", "mise", "tasks", "helpers.py")
}

// pyTask gives the command that runs the Python task file <path>: the file
// itself, with mode 755, so its shebang runs uv.
func pyTask(t testing.TB, path string) []string {
	t.Helper()
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return []string{path}
}

// copyFile copies <src> to <dst>, with the folders of <dst>.
func copyFile(t testing.TB, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, 0o755); err != nil {
		t.Fatal(err)
	}
}

// copyConsumer copies the base consumer of the repo cases to <dst>.
func copyConsumer(t testing.TB, dst string) {
	t.Helper()
	if err := harness.CopyTree(rc.BaseConsumer, dst); err != nil {
		t.Fatal(err)
	}
}

// gomplate renders the template <src> to <dst> as harness.Gomplate does.
// harness.Gomplate needs gomplate on PATH; the old task got it from its
// `#MISE tools.gomplate` line. Without gomplate on PATH, this helper runs it
// through `mise x gomplate@latest`.
func gomplate(t testing.TB, src, dst, ctx string) {
	t.Helper()
	if _, err := exec.LookPath("gomplate"); err == nil {
		harness.Gomplate(t, src, dst, ctx)
		return
	}
	args := []string{"x", "gomplate@latest", "--", "gomplate", "--left-delim", "<%", "--right-delim", "%>"}
	if ctx != "" {
		args = append(args, "-c", ".="+ctx)
	}
	args = append(args, "-f", src, "-o", dst)
	cmd := exec.Command(harness.RealMise, args...)
	cmd.Dir = harness.Root
	cmd.Env = harness.Env()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gomplate %s: %v\n%s", src, err, out)
	}
}
