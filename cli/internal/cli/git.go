package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// toplevel gives the root of the git work tree of the folder <dir>.
func toplevel(dir string) (string, bool) {
	top, ok := git("-C", dir, "rev-parse", "--show-toplevel")
	return top, ok && top != ""
}

// realDir gives the real path of the nearest folder of <p> that exists, and
// the rest of <p> below it.
func realDir(p string) (string, string) {
	dir, rest := filepath.Dir(p), filepath.Base(p)
	for !exists(dir) && dir != "/" && dir != "." {
		rest = filepath.Join(filepath.Base(dir), rest)
		dir = filepath.Dir(dir)
	}
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	return dir, rest
}

// dirtyTargets gives each path of <paths> that is inside a git work tree and
// has changes that are not committed there: modified, staged or untracked.
// A path of <skip> is not examined.
func dirtyTargets(paths []string, skip map[string]bool) []string {
	tops := map[string]string{}            // A real folder to its work tree.
	status := map[string]map[string]bool{} // A work tree to its dirty paths.
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		if skip[p] || seen[p] {
			continue
		}
		seen[p] = true
		dir, rest := realDir(p)
		top, ok := tops[dir]
		if !ok {
			top, _ = toplevel(dir)
			if r, err := filepath.EvalSymlinks(top); err == nil {
				top = r
			}
			tops[dir] = top
		}
		if top == "" {
			continue
		}
		set, ok := status[top]
		if !ok {
			set = statusSet(top)
			status[top] = set
		}
		rel, err := filepath.Rel(top, filepath.Join(dir, rest))
		if err == nil && set[rel] {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// statusSet gives the paths with changes that are not committed in the work
// tree <top>, relative to it.
func statusSet(top string) map[string]bool {
	set := map[string]bool{}
	out, ok := git("-C", top, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if !ok {
		return set
	}
	entries := strings.Split(out, "\x00")
	for i := 0; i < len(entries); i++ {
		e := entries[i]
		if len(e) < 4 {
			continue
		}
		set[e[3:]] = true
		// A rename or a copy: the next entry is the old path.
		if e[0] == 'R' || e[0] == 'C' {
			i++
			if i < len(entries) {
				set[entries[i]] = true
			}
		}
	}
	return set
}

// stopDirty prints each dirty target and gives code 2, or gives 0 when
// <dirty> is empty.
func (a *app) stopDirty(dirty []string) fail.Code {
	if len(dirty) == 0 {
		return fail.OK
	}
	for _, p := range dirty {
		a.errorf("%s has changes that are not committed", a.show(p))
	}
	a.fail("nothing written", "commit or discard the changes of these targets, then run the command again")
	return fail.Usage
}

// warnDirty prints each dirty target of a dry run.
func (a *app) warnDirty(dirty []string) {
	for _, p := range dirty {
		a.warnf("%s has changes that are not committed; a run without --dry-run stops", a.show(p))
	}
}

// interactive tells if stdin and stdout are terminals.
func (a *app) interactive() bool {
	if a.terminal {
		return true
	}
	in, ok1 := a.stdin.(*os.File)
	out, ok2 := a.stdout.(*os.File)
	return ok1 && ok2 && report.IsTerminal(in) && report.IsTerminal(out)
}

// readLine reads one line of stdin, without its newline.
func (a *app) readLine() (string, bool) {
	if a.in == nil {
		a.in = bufio.NewReader(a.stdin)
	}
	line, err := a.in.ReadString('\n')
	if err != nil && line == "" {
		return "", false
	}
	return strings.TrimRight(line, "\r\n"), true
}

// consent gives OK when the changes can be written: --yes, else a "y" at the
// prompt of a terminal. A "no" gives done = true: nothing is written, and the
// command exits 0. Without a terminal, it gives code 2.
func (a *app) consent(yes bool, n int) (done bool, code fail.Code) {
	if yes {
		return false, fail.OK
	}
	if !a.interactive() {
		a.fail("changes need --yes", "run the command in a terminal, or add --yes")
		return true, fail.Usage
	}
	fmt.Fprintf(a.stdout, "? Write %s? [y/N] ", count(n, "change"))
	line, _ := a.readLine()
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return false, fail.OK
	}
	a.infof("Nothing changed")
	return true, fail.OK
}

// count gives "<n> <noun>" with the plural of <noun> when <n> is not 1.
func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
