// Package target writes one target of the sync: the temp file, the seed gate,
// the unblock of a folder path, the parse check and the merge of the temp
// file, and the compare, the mode, the dry-run diff and the rename. It also
// writes `bootstrap_commit` in a values file. It replaces sync-engine/write
// (with internal/format) and _se_mode, _se_path, _se_cleanup and
// _se_record_commit of sync-engine/common, with the same output, files and
// return codes.
//
// A front-end calls the steps in the order of the bash front-ends. For each
// target: SeedKept, Unblock, its own mkdir, Temps.Create, the render or the
// copy into the temp file, Check, Merge (merge mode only), and on a failure
// Abort, else Finish. The front-end keeps its own steps between them (the
// skip of an environment, `requires`, the symlink checks of a merge target,
// the `.ssh` folder, `pre-commit autoupdate`, dprint and the overrides).
//
// Contract: docs/design/sync-engine.md ("The write of one target",
// "`bootstrap_commit`").
package target

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/report"
)

// Env is the context of the writes of one run.
type Env struct {
	// Report counts the outcomes and prints the lines.
	Report *report.Report
	// Stderr is the stream of Report. The dry-run diff and the detail of a
	// failed parse check go there.
	Stderr io.Writer
	// Pwd is the current folder, as $PWD of bash. Show makes a path relative
	// to it.
	Pwd string
	// DryRun prints each change and writes no target.
	DryRun bool
	// Temps holds the temp files of the run. The front-end defers
	// Temps.Cleanup.
	Temps Temps
}

// Show is _se_path: <path> relative to Env.Pwd when it is below it, else
// <path> as it is.
func (e *Env) Show(path string) string { return Show(path, e.Pwd) }

// Show is _se_path with the current folder <pwd>.
func Show(path, pwd string) string {
	if strings.HasPrefix(path, pwd+"/") {
		return path[len(pwd)+1:]
	}
	return path
}

// Mode is _se_mode: the permission bits of <path> (the link itself for a
// symlink, as `stat` without -L).
func Mode(path string) (os.FileMode, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
}

// DryRunDir gives the folder of the temp files of a dry run: $TMPDIR, else
// /tmp. <getenv> is os.Getenv in the engine.
func DryRunDir(getenv func(string) string) string {
	if d := getenv("TMPDIR"); d != "" {
		return d
	}
	return "/tmp"
}

// Temps is _SE_TMPS: the temp files and folders that are still on disk. The
// zero value is ready to use.
type Temps struct{ paths []string }

// Add records <path>, to remove in Cleanup.
func (t *Temps) Add(path string) { t.paths = append(t.paths, path) }

// Create is _se_tmp: it creates an empty temp file `.sync-engine.*` with mode
// 0600 in <dir>, records it and gives its path.
func (t *Temps) Create(dir string) (string, error) {
	return t.create(strings.TrimSuffix(dir, "/"), ".sync-engine.*")
}

// CreateNext creates an empty temp file `<path>.*` next to <path>, as
// `mktemp "<path>.XXXXXX"`, records it and gives its path.
func (t *Temps) CreateNext(path string) (string, error) {
	return t.create(Dirname(path), filepath.Base(path)+".*")
}

func (t *Temps) create(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	t.Add(name)
	return name, nil
}

// Cleanup is _se_cleanup: it removes each recorded path that is still there.
// A front-end defers it, and calls it on SIGINT (exit 130) and SIGTERM
// (exit 143), as the traps of common do.
func (t *Temps) Cleanup() {
	for _, p := range t.paths {
		os.RemoveAll(p)
	}
	t.paths = nil
}

// Move is `mv -f <from> <to>`. As mv, it moves <from> into <to> when <to> is
// a folder (also through a symlink).
func Move(from, to string) error {
	if info, err := os.Stat(to); err == nil && info.IsDir() {
		to = filepath.Join(to, filepath.Base(from))
	}
	return os.Rename(from, to)
}

// Dirname is the `dirname` command. It does not clean the path, so a `..`
// after a symlink keeps its meaning.
func Dirname(p string) string {
	for len(p) > 1 && strings.HasSuffix(p, "/") {
		p = p[:len(p)-1]
	}
	i := strings.LastIndex(p, "/")
	switch {
	case i < 0:
		return "."
	case i == 0:
		return "/"
	}
	p = p[:i]
	for len(p) > 1 && strings.HasSuffix(p, "/") {
		p = p[:len(p)-1]
	}
	return p
}
