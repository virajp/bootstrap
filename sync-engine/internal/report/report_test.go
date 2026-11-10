package report

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// The golden files in testdata/ hold the expected output of each case. The
// cases do not read the repository, so a change of the repository does not
// make a golden file old.

// step is one call of a case: an outcome, an info line, a log line or the
// start of a run. The summary is always the last call.
type step struct {
	call             string // "outcome", "info", "log" or "start"
	kind             Kind
	path, text, line string
	dry              bool
}

type goldenCase struct {
	name    string
	dryRun  bool
	verbose bool
	steps   []step
	rc      int
}

func outcome(kind Kind, path, text, line string) step {
	return step{call: "outcome", kind: kind, path: path, text: text, line: line}
}

var mixed = []step{
	outcome(Changed, ".config/mise/tasks/code/lint", "", "updated    .config/mise/tasks/code/lint"),
	outcome(New, ".gitignore", "", "updated    .gitignore"),
	outcome(New, "dir with space/a'b.toml", "", "updated    dir with space/a'b.toml"),
	outcome(Seeded, ".config/linter.yaml", "", "updated    .config/linter.yaml"),
	outcome(Removed, "old/file", "", "removed    old/file"),
	outcome(Removed, "old/file", "", "removed    old/file (again)"),
	outcome(Unchanged, "same", "", "unchanged  same"),
	outcome(Kept, "seed", "", "kept seed  seed"),
	outcome(Skipped, "skip", "", ""),
	outcome(Skipped, "skip2", "", "skip env   skip2"),
}

var goldenCases = []goldenCase{
	{name: "empty", rc: 0},
	{name: "empty-dry", dryRun: true, rc: 0},
	{name: "one-changed", steps: []step{outcome(Changed, "a", "", "updated    a")}, rc: 0},
	{name: "one-changed-dry", dryRun: true, steps: []step{outcome(Changed, "a", "", "would update a")}, rc: 0},
	{name: "mixed", steps: mixed, rc: 0},
	{name: "mixed-dry", dryRun: true, steps: mixed, rc: 0},
	{name: "mixed-verbose", verbose: true, steps: mixed, rc: 0},
	{name: "failed", steps: []step{
		outcome(Changed, "a", "", "updated    a"),
		outcome(Failed, "b", "cannot write the target", "cannot write b"),
		outcome(Failed, "b", "kept existing", "kept existing b"),
		outcome(Failed, "c", "template does not render", ""),
	}, rc: 2},
	{name: "failed-dry", dryRun: true, steps: []step{
		outcome(Failed, "b", "cannot write the target", "cannot write b"),
		outcome(New, "n", "", "would update n"),
	}, rc: 1},
	{name: "failed-verbose", verbose: true, steps: []step{
		outcome(Failed, "b", "cannot write the target", "cannot write b"),
		outcome(Failed, "b", "kept existing", "kept existing b"),
	}, rc: 1},
	{name: "code-no-failed", steps: []step{outcome(Unchanged, "u", "", "")}, rc: 5},
	{name: "info-and-log", steps: []step{
		{call: "info", text: "Using bootstrap at 1234567"},
		{call: "log", text: "kept a (link)"},
		outcome(New, "a", "", "updated    a"),
		{call: "info", text: "Set bootstrap_commit to 1234567"},
	}, rc: 0},
	{name: "restart", steps: []step{
		outcome(Changed, "a", "", ""),
		outcome(Failed, "f", "x", ""),
		{call: "start", dry: true},
		outcome(Failed, "f", "y", ""),
		outcome(Removed, "r", "", ""),
	}, rc: 0},
}

// run gives the output of the Go code for <c>.
func (c goldenCase) run(style bool) string {
	var b bytes.Buffer
	r := NewReport(&b, Options{Style: style, Verbose: c.verbose})
	r.Start(c.dryRun)
	for _, s := range c.steps {
		switch s.call {
		case "outcome":
			r.Outcome(s.kind, s.path, s.text, s.line)
		case "info":
			r.Info(s.text)
		case "log":
			r.Log(s.text)
		case "start":
			r.Start(s.dry)
		}
	}
	r.Summary(c.rc)
	return b.String()
}

func goldenPath(c goldenCase, style bool) string {
	suffix := "plain"
	if style {
		suffix = "style"
	}
	return filepath.Join("testdata", c.name+"."+suffix+".golden")
}

func TestGolden(t *testing.T) {
	for _, c := range goldenCases {
		for _, style := range []bool{false, true} {
			path := goldenPath(c, style)
			t.Run(filepath.Base(path), func(t *testing.T) {
				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if got := c.run(style); got != string(want) {
					t.Errorf("output differs from %s\ngot:\n%q\nwant:\n%q", path, got, want)
				}
			})
		}
	}
}

func TestStyleFromEnv(t *testing.T) {
	cases := []struct {
		name     string
		env      map[string]string
		terminal bool
		want     bool
	}{
		{"force", map[string]string{"FORCE_COLOR": "1", "NO_COLOR": "1"}, false, true},
		{"force 0 is not force", map[string]string{"FORCE_COLOR": "0", "TERM": "xterm"}, true, true},
		{"force 0 without terminal", map[string]string{"FORCE_COLOR": "0", "TERM": "xterm"}, false, false},
		{"no color", map[string]string{"NO_COLOR": "1", "TERM": "xterm"}, true, false},
		{"terminal", map[string]string{"TERM": "xterm-256color"}, true, true},
		{"no terminal", map[string]string{"TERM": "xterm-256color"}, false, false},
		{"dumb", map[string]string{"TERM": "dumb"}, true, false},
		{"no TERM", map[string]string{}, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			getenv := func(k string) string { return c.env[k] }
			if got := StyleFromEnv(getenv, c.terminal); got != c.want {
				t.Errorf("StyleFromEnv() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestVerboseFromEnv(t *testing.T) {
	for v, want := range map[string]bool{"1": true, "": false, "0": false, "true": false, "yes": false} {
		getenv := func(string) string { return v }
		if got := VerboseFromEnv(getenv); got != want {
			t.Errorf("VerboseFromEnv(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if IsTerminal(f) {
		t.Error("a regular file is a terminal")
	}
}

func TestPrefix(t *testing.T) {
	var b bytes.Buffer
	NewReport(&b, Options{Prefix: "setup-sync"}).Log("tools: a b")
	if got, want := b.String(), "setup-sync: tools: a b\n"; got != want {
		t.Errorf("Log() = %q, want %q", got, want)
	}
}

func TestCount(t *testing.T) {
	r := NewReport(&bytes.Buffer{}, Options{})
	r.Outcome(Failed, "a", "x", "")
	r.Outcome(Failed, "a", "y", "")
	r.Outcome(Removed, "a", "", "")
	r.Outcome(Unchanged, "b", "", "")
	r.Outcome(Unchanged, "b", "", "")
	for kind, want := range map[Kind]int{Failed: 1, Removed: 1, Unchanged: 2, Changed: 0} {
		if got := r.Count(kind); got != want {
			t.Errorf("Count(%d) = %d, want %d", kind, got, want)
		}
	}
}
