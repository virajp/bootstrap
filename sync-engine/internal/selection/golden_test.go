package selection

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/report"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// The golden file testdata/<case>.golden holds the expected output of each
// case: the exit code, the log and the result of the read steps (the tool
// list, the stages, the merged values). To write the golden files again from
// the Go output, run: UPDATE_GOLDEN=1 go test ./internal/selection

// A case is one run of the read steps against a clone and a values file.
type pcase struct {
	name   string
	side   source.Side
	clone  string // The clone copy.
	values string // The values file (repo) or the config folder (macos).
	shown  string // The values file as a message names it.
	sub    bool   // Repo: the repo is a git submodule.
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

// check compares the result of <c> with the golden file.
func check(t *testing.T, c pcase) {
	t.Helper()
	golden := filepath.Join("testdata", c.name+".golden")
	got := c.golang()
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run UPDATE_GOLDEN=1 go test ./internal/selection)", err)
	}
	if got != string(want) {
		t.Errorf("%s: the output differs from the golden file\n--- got\n%s\n--- want\n%s", c.name, got, want)
	}
}
