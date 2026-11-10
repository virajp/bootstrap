// Package cli is the command line of `bootstrap`: the commands init, sync,
// check, questions, add, remove, replace, detect and list of the repo side and of the Mac side (`bootstrap mac
// <command>`), and version. It does the steps that each command shares: the
// preflight, the clone and its lock, the version check, the change list, the
// clean-target rule and the consent. The front-ends internal/repo and
// internal/macos do the sync itself.
//
// Contract: docs/design/cli.md.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/macos"
	"github.com/virajp/bootstrap/cli/internal/repo"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
)

// Usage is the help text of the command.
const Usage = `usage: bootstrap <command> [flags]

commands:
  init [tool...]       select base and the named tools, migrate an old form,
                       ask the questions without an answer, then sync
  sync                 sync the repository of the current folder
  check                examine the values file of the repository
  questions [tool...]  list the questions of the tools, with their proposals
                       and their answers
  add <tool...>        select the tools and their dependencies, ask their
                       questions, then sync
  remove <tool...>     remove the tools and the dependencies that no other
                       tool needs, then sync
  replace <old> <new>  run the migrate script of <new> for <old>, swap the
                       tools, then sync
  detect               run the detect scripts, and offer to add each tool
                       that matches
  list, ls             list the selected tools, with the reason of each one
  mac init [tool...]   init for the Mac, in the current folder (the
                       repository of macos.toml)
  mac sync             sync the Mac from MACOS_SETUP_REPO_PATH
  mac check            show what mac sync changes, and write nothing
  mac questions [tool...]
                       the questions of the Mac tools
  mac add, mac remove, mac replace, mac detect
                       the selection commands for macos.toml
  mac list, mac ls     list the selected Mac tools
  version, --version   print the version

flags:
  --dry-run            init, sync, add, remove, replace, detect: show the
                       changes, and write nothing
  --yes, -y            init, add, remove, replace, detect: write the changes
                       without a prompt; detect: add each matched tool
  --no                 detect: add nothing; a matched tool that is not
                       selected exits 1
  --answer <tool>.<key>=<value>
                       init, add, replace, detect: the answer of a question
                       (repeatable)
  --json               questions, mac questions: JSON output
  --no-bootstrap       mac init, mac sync, mac check: skip mise bootstrap
  --jobs <n>           the jobs at a time (default: 2 x the CPU count)
`

// app is the state of one command.
type app struct {
	version        string
	stdin          io.Reader
	stdout, stderr io.Writer
	getenv         func(string) string
	styled         bool
	in             *bufio.Reader
	pwd            string
}

// opts are the flags and the arguments of a command.
type opts struct {
	dryRun, yes, no, json, noBootstrap, help bool
	confirmed                                bool // The user said yes to the changes at a prompt.
	answers                                  []string
	names                                    []string
}

// Run runs the command of <args> and gives its exit code.
func Run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer) fail.Code {
	a := &app{version: version, stdin: stdin, stdout: stdout, stderr: stderr, getenv: os.Getenv}
	isTerm := false
	if f, ok := stderr.(*os.File); ok {
		isTerm = report.IsTerminal(f)
	}
	a.styled = report.StyleFromEnv(a.getenv, isTerm)
	a.pwd = a.getenv("PWD")
	if wd, err := os.Getwd(); a.pwd == "" && err == nil {
		a.pwd = wd
	}
	if len(args) == 0 {
		a.errorf("no command")
		fmt.Fprint(stderr, Usage)
		return fail.Usage
	}
	mac := false
	cmd, rest := args[0], args[1:]
	if cmd == "--version" {
		cmd = "version"
	}
	if cmd == "mac" {
		if len(rest) == 0 {
			a.errorf("mac: no command")
			fmt.Fprint(stderr, Usage)
			return fail.Usage
		}
		mac = true
		cmd, rest = rest[0], rest[1:]
	}
	if cmd == "ls" {
		cmd = "list"
	}
	allowed := map[string]string{
		"init":      "dry-run yes answer jobs",
		"sync":      "dry-run jobs",
		"check":     "jobs",
		"questions": "json",
		"add":       "dry-run yes answer jobs",
		"remove":    "dry-run yes jobs",
		"replace":   "dry-run yes answer jobs",
		"detect":    "dry-run yes no answer jobs",
		"list":      "",
		"version":   "",
	}
	flags, ok := allowed[cmd]
	if !ok || (mac && cmd == "version") {
		switch cmd {
		case "-h", "-help", "--help", "help":
			if !mac {
				fmt.Fprint(stdout, Usage)
				return fail.OK
			}
		}
		name := cmd
		if mac {
			name = "mac " + cmd
		}
		a.errorf("unknown command %s", name)
		fmt.Fprint(stderr, Usage)
		return fail.Usage
	}
	if mac && cmd != "questions" {
		flags += " no-bootstrap"
	}
	takesNames := cmd == "init" || cmd == "questions" || cmd == "add" || cmd == "remove" || cmd == "replace"
	o, err := parse(rest, flags, takesNames)
	if err != nil {
		a.errorf("%s", err)
		fmt.Fprint(stderr, Usage)
		return fail.Usage
	}
	if o.help {
		fmt.Fprint(stdout, Usage)
		return fail.OK
	}
	if (cmd == "add" || cmd == "remove") && len(o.names) == 0 {
		a.errorf("%s needs at least one tool", cmdName(mac, cmd))
		fmt.Fprint(stderr, Usage)
		return fail.Usage
	}
	switch cmd {
	case "version":
		fmt.Fprintln(stdout, version)
		return fail.OK
	case "sync":
		return a.sync(mac, o)
	case "check":
		return a.check(mac, o)
	case "init":
		return a.init(mac, o)
	case "add":
		return a.add(mac, o)
	case "remove":
		return a.remove(mac, o)
	case "replace":
		return a.replace(mac, o)
	case "detect":
		return a.detect(mac, o)
	case "list":
		return a.list(mac)
	default:
		return a.questions(mac, o)
	}
}

// parse reads the flags of <flags> (names without dashes) and, with
// <names>, the arguments.
func parse(args []string, flags string, names bool) (opts, error) {
	var o opts
	has := map[string]bool{}
	for _, f := range strings.Fields(flags) {
		has[f] = true
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, value, withValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if !strings.HasPrefix(arg, "-") {
			if !names {
				return o, fmt.Errorf("unknown argument %s", arg)
			}
			o.names = append(o.names, arg)
			continue
		}
		if arg == "-h" || arg == "--help" {
			o.help = true
			continue
		}
		if arg == "-y" {
			name = "yes"
		}
		if !has[name] {
			return o, fmt.Errorf("unknown flag %s", arg)
		}
		takes := name == "answer" || name == "jobs"
		if takes && !withValue {
			if i+1 >= len(args) {
				return o, fmt.Errorf("%s needs a value", arg)
			}
			i++
			value = args[i]
		} else if !takes && withValue {
			return o, fmt.Errorf("%s takes no value", arg)
		}
		switch name {
		case "dry-run":
			o.dryRun = true
		case "yes":
			o.yes = true
		case "no":
			o.no = true
		case "json":
			o.json = true
		case "no-bootstrap":
			o.noBootstrap = true
		case "answer":
			o.answers = append(o.answers, value)
		case "jobs":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return o, fmt.Errorf("--jobs needs a number above 0, not '%s'", value)
			}
			// The script runner reads it.
			_ = os.Setenv("BOOTSTRAP_JOBS", value)
		}
	}
	return o, nil
}

func (a *app) line(glyph, text string) {
	if a.styled {
		fmt.Fprintf(a.stderr, "%s %s\n", glyph, text)
	} else {
		fmt.Fprintf(a.stderr, "bootstrap: %s\n", text)
	}
}

func (a *app) infof(format string, args ...any)  { a.line("ℹ️", fmt.Sprintf(format, args...)) }
func (a *app) errorf(format string, args ...any) { a.line("❌", fmt.Sprintf(format, args...)) }

// logErr prints each message of <err> and gives its code.
func (a *app) logErr(err error) fail.Code {
	if err == nil {
		return fail.OK
	}
	for _, m := range source.Messages(err) {
		a.errorf("%s", m.Error())
	}
	return fail.CodeOf(err)
}

// show gives <p> relative to the current folder when it is below it.
func (a *app) show(p string) string { return target.Show(p, a.pwd) }

// where is the folder and the values file of a command.
type where struct {
	root   string // The repository; on the Mac, the repository of macos.toml.
	values string
}

// prepare does the preflight of a command of the side, and finds its
// folder and its values file.
func (a *app) prepare(mac, initCmd bool) (where, fail.Code) {
	var w where
	if _, err := exec.LookPath("mise"); err != nil {
		a.errorf("mise is not available - install it from https://mise.run")
		return w, fail.Usage
	}
	if mac {
		dir := a.getenv("MACOS_SETUP_REPO_PATH")
		if initCmd {
			dir = a.pwd
		} else if dir == "" {
			a.errorf("MACOS_SETUP_REPO_PATH is not set - set it to the repository of macos.toml, or run bootstrap mac init in that repository")
			return w, fail.Usage
		}
		if _, ok := toplevel(dir); !ok {
			if initCmd {
				a.errorf("%s is not a git repository - create one with git init, then run bootstrap mac init again", dir)
			} else {
				a.errorf("MACOS_SETUP_REPO_PATH %s is not a git repository", dir)
			}
			return w, fail.Usage
		}
		_ = os.Setenv("MACOS_SETUP_REPO_PATH", dir)
		w.root, w.values = dir, strings.TrimSuffix(dir, "/")+"/macos.toml"
		return w, fail.OK
	}
	root := a.getenv("MISE_PROJECT_ROOT")
	dir := root
	if dir == "" {
		dir = a.pwd
	}
	top, ok := toplevel(dir)
	if !ok {
		if initCmd {
			a.errorf("%s is not a git repository - create one with git init, then run bootstrap init again", dir)
		} else {
			a.errorf("%s is not a git repository", dir)
		}
		return w, fail.Usage
	}
	if root == "" {
		root = top
		_ = os.Setenv("MISE_PROJECT_ROOT", root)
	}
	w.root = root
	w.values = a.getenv("BOOTSTRAP_VALUES")
	if w.values == "" {
		w.values = root + "/.config/bootstrap.toml"
	}
	return w, fail.OK
}

// openClone takes the clone of the command and checks its version. The
// caller defers cleanup. With <hold>, the lock stays held to the end.
func (a *app) openClone(hold bool) (*clone, fail.Code) {
	c, code := a.acquireClone()
	if code != fail.OK {
		return c, code
	}
	if !hold {
		c.unlock()
	}
	// A signal removes the lock and the copy.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		got := <-sig
		c.cleanup()
		if got == syscall.SIGINT {
			os.Exit(130)
		}
		os.Exit(143)
	}()
	if code := a.checkVersion(c.dir); code != fail.OK {
		return c, code
	}
	_ = os.Setenv("BOOTSTRAP_CLONE", c.dir)
	return c, fail.OK
}

// frontEnd is the sync of a side, with its change list.
type frontEnd func(args []string, stdout, stderr io.Writer) ([]string, fail.Code)

func frontEndOf(mac bool) frontEnd {
	if mac {
		return macos.RunWith
	}
	return repo.RunWith
}

// syncArgs are the arguments of the front-end.
func syncArgs(mac bool, o opts) []string {
	var args []string
	if o.dryRun {
		args = append(args, "--dry-run")
	}
	if mac && o.noBootstrap {
		args = append(args, "--no-bootstrap")
	}
	return args
}

// valuesSkip gives the values file and its real path: sync writes only
// bootstrap_commit there, so the clean-target rule skips it.
func valuesSkip(values string) map[string]bool {
	skip := map[string]bool{values: true}
	if r, err := realPath(values); err == nil {
		skip[r] = true
	}
	return skip
}

// sync is `bootstrap [mac] sync`.
func (a *app) sync(mac bool, o opts) fail.Code {
	w, code := a.prepare(mac, false)
	if code != fail.OK {
		return code
	}
	c, code := a.openClone(mac && !o.dryRun)
	defer c.cleanup()
	if code != fail.OK {
		return code
	}
	run := frontEndOf(mac)
	args := syncArgs(mac, o)
	skip := valuesSkip(w.values)
	if o.dryRun {
		planned, code := run(args, a.stdout, a.stderr)
		a.warnDirty(dirtyTargets(planned, skip))
		return code
	}
	// The plan, then the write. A repository without a change that is not
	// committed needs no plan.
	if mac || a.dirtyBesides(w.root, skip) {
		planned, code := run(append(args, "--dry-run"), io.Discard, io.Discard)
		if code == fail.OK {
			if code := a.stopDirty(dirtyTargets(planned, skip)); code != fail.OK {
				return code
			}
		}
	}
	_, code = run(args, a.stdout, a.stderr)
	return code
}

// dirtyBesides tells if the work tree of <root> has a change that is not
// committed in a path that is not in <skip>.
func (a *app) dirtyBesides(root string, skip map[string]bool) bool {
	top, ok := toplevel(root)
	if !ok {
		return false
	}
	if r, err := realPath(top); err == nil {
		top = r
	}
	for p := range statusSet(top) {
		if !skip[top+"/"+p] {
			return true
		}
	}
	return false
}

// check is `bootstrap [mac] check`.
func (a *app) check(mac bool, o opts) fail.Code {
	w, code := a.prepare(mac, false)
	if code != fail.OK {
		return code
	}
	_ = w
	c, code := a.openClone(false)
	defer c.cleanup()
	if code != fail.OK {
		return code
	}
	if mac {
		o.dryRun = true
		return macos.Run(syncArgs(true, o), a.stdout, a.stderr)
	}
	return repo.Run([]string{"--check"}, a.stdout, a.stderr)
}
