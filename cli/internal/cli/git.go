package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

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
	_, _ = fmt.Fprintf(a.stdout, "? Write %s? [y/N] ", count(n, "change"))
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
