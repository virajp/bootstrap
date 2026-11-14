// Package cli is the command line of `bootstrap`: the commands init, sync,
// check, questions, add, remove, replace, detect and list of the repo side and of the Mac side (`bootstrap mac
// <command>`), and version. It is the shell of the command line: it parses
// the arguments, takes the clone and its lock, checks the version, calls
// internal/core (the preflight, Plan, Apply, Detect and the clean-target
// rule), prints the results and the errors, and asks the questions and the
// consent. The front-ends internal/repo and internal/macos do the sync itself.
//
// Contract: docs/design/cli.md.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/field"
	"github.com/virajp/bootstrap/cli/internal/macos"
	"github.com/virajp/bootstrap/cli/internal/repo"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
)

// Usage is the help text of the command.
const Usage = `usage: bootstrap <command> [flags]

Repository
  init                 migrate an old form, or create the values file with
                       the members of the pack base, ask the questions
                       without an answer, then sync
  sync                 sync the repository of the current folder
  check                examine the values file of the repository
  questions [tool...]  list the questions of the tools, with their proposals
                       and their answers
  add tool <tool...>   select the tools and their dependencies, ask their
                       questions, then sync
  add pack <pack...>   select each member of the packs that is not selected,
                       and their dependencies, ask their questions, then sync
  remove tool <tool...>
                       remove the tools and the dependencies that no other
                       tool needs, then sync
  replace tool <old> <new>
                       run the migrate script of <new> for <old>, swap the
                       tools, then sync
  detect               run the detect scripts and list the tools that match;
                       exit 1 when a tool that matches is not selected
  list, ls             print the selected tools as a tree
  config               print each setting of the engine, with its value and
                       its source
  version, --version   print the version

Mac
  mac init             init for the Mac, in the current folder (the
                       repository of macos.toml)
  mac sync             sync the Mac from MACOS_SETUP_REPO_PATH
  mac check            show what mac sync changes, and write nothing
  mac questions [tool...]
                       the questions of the Mac tools
  mac add tool|pack, mac remove tool, mac replace tool, mac detect
                       the selection commands for macos.toml
  mac list, mac ls     print the selected Mac tools as a tree

Flags
  --dry-run            init, sync, add, remove, replace, detect: show the
                       changes, and write nothing
  --yes, -y            init, add, remove, replace, detect --add: write the
                       changes without a prompt
  --no-tui             init: select no tool on a screen; init takes this
                       path in each case for now
  --add                detect: add each matched tool that is not selected,
                       after the consent
  --exclude <member>   add pack: leave out a member of the pack, as
                       packs.toml writes it (repeatable; one pack only)
  --answer <tool>.<key>=<value>
                       init, add, replace, detect --add: the answer of a
                       question (repeatable); --answer <tool>.<key>= writes
                       an empty answer for an optional or a group question
  --json               questions, mac questions: JSON output
  --no-bootstrap       mac init, mac sync, mac check: skip mise bootstrap
  --jobs <n>           the jobs at a time (default: 2 x the CPU count)
  --verbose            init, sync, check, add, remove, replace, detect: print
                       the log line of each target
  --config <path>      each command except version: the config file (default:
                       ~/.config/bootstrap/config.toml)
  --cache-dir <path>   each command except version: the cache folder
  --url <url>          each command except version: the clone URL
  --clone <path>       each command except version: a clone of bootstrap, with
                       no fetch
  --repo-path <path>   each mac command except mac init, and config: the
                       repository of macos.toml
`

// argsOf gives the arguments of a command in its usage line.
var argsOf = map[string]string{
	"questions": " [tool...]", "add": " tool|pack <name...>",
	"remove": " tool <tool...>", "replace": " tool <old> <new>",
}

// flagUsage gives each flag of a usage line, by the name of the flag.
var flagUsage = map[string]string{
	"answer": "[--answer <tool>.<key>=<value>]", "jobs": "[--jobs <n>]",
	"repo-path": "[--repo-path <path>]", "exclude": "[--exclude <member>]",
}

// usageLine gives the usage line of a command: its arguments and its own
// flags, not the flags of every command.
func usageLine(mac bool, cmd, flags string) string {
	b := "bootstrap " + cmdWord(mac, cmd) + argsOf[cmd]
	for _, f := range strings.Fields(flags) {
		switch f {
		case "config", "cache-dir", "url", "clone":
			continue
		}
		if u, ok := flagUsage[f]; ok {
			b += " " + u
		} else {
			b += " [--" + f + "]"
		}
	}
	return b
}

func cmdWord(mac bool, cmd string) string { return core.CmdWord(mac, cmd) }

// app is the state of one command.
type app struct {
	version        string
	stdin          io.Reader
	stdout, stderr io.Writer
	getenv         func(string) string
	styled         bool
	in             *bufio.Reader
	pwd            string
	settings       settings // The settings of the command and their sources.
	reasons        []string // Why a command has nothing to add: the end line shows them.
	// prompt asks a question in a field; nil is the field of the terminal.
	// Tests replace it.
	prompt func(field.Spec) (string, error)
	// terminal makes interactive true. Tests set it with prompt.
	terminal bool
}

// opts are the flags and the arguments of a command.
type opts struct {
	dryRun, yes, json, noBootstrap, help bool
	verbose, noTUI, add                  bool
	answers                              []string
	kind                                 string            // add, remove, replace: the kind word, tool or pack.
	names                                []string          // The names after the kind word.
	exclude                              []string          // add pack: the members of --exclude.
	values                               map[string]string // The value flags of the settings, by flag name.
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
		a.usageErr("no command given", "")
		return fail.Usage
	}
	mac := false
	cmd, rest := args[0], args[1:]
	if cmd == "--version" {
		cmd = "version"
	}
	if cmd == "mac" {
		if len(rest) == 0 {
			a.usageErr("mac needs a command", "")
			return fail.Usage
		}
		mac = true
		cmd, rest = rest[0], rest[1:]
	}
	if cmd == "ls" {
		cmd = "list"
	}
	allowed := map[string]string{
		"init":      "dry-run yes no-tui answer jobs verbose",
		"sync":      "dry-run jobs verbose",
		"check":     "jobs verbose",
		"questions": "json",
		"add":       "dry-run yes exclude answer jobs verbose",
		"remove":    "dry-run yes jobs verbose",
		"replace":   "dry-run yes answer jobs verbose",
		"detect":    "dry-run add yes answer jobs verbose",
		"list":      "",
		"config":    "jobs repo-path",
		"version":   "",
	}
	flags, ok := allowed[cmd]
	if !ok || (mac && (cmd == "version" || cmd == "config")) {
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
		a.usageErr(name+" is not a command", "")
		return fail.Usage
	}
	own := flags
	if mac && cmd != "questions" {
		flags += " no-bootstrap"
	}
	if cmd != "version" {
		flags += " config cache-dir url clone"
	}
	if mac && cmd != "init" {
		flags += " repo-path"
	}
	takesNames := cmd == "init" || cmd == "questions" || cmd == "add" || cmd == "remove" || cmd == "replace"
	o, err := parse(cmdWord(mac, cmd), rest, flags, takesNames)
	if err != nil {
		a.usageErr(err.Error(), usageLine(mac, cmd, own))
		return fail.Usage
	}
	if o.help {
		fmt.Fprint(stdout, Usage)
		return fail.OK
	}
	if code := a.kindWord(mac, cmd, own, &o); code != fail.OK {
		return code
	}
	if cmd == "init" && len(o.names) > 0 {
		a.fail(cmdWord(mac, "init")+" takes no tool names", "use: "+cmdName(mac, "add tool")+" "+strings.Join(o.names, " "))
		return fail.Usage
	}
	if cmd == "detect" && !o.add && (o.yes || len(o.answers) > 0) {
		flag := "--yes"
		if !o.yes {
			flag = "--answer"
		}
		a.fail(cmdWord(mac, "detect")+" "+flag+" needs --add", "add --add")
		return fail.Usage
	}
	if cmd == "version" {
		fmt.Fprintln(stdout, version)
		return fail.OK
	}
	// The config file, before any other work.
	if code := a.applySettings(o); code != fail.OK {
		return code
	}
	switch cmd {
	case "config":
		return a.config(o)
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

// kindWord takes the kind word of add, remove and replace from the names of
// <o>: add takes tool or pack, remove and replace take tool. A name with no
// kind word is the old form (code 2, with the hint of the new form).
// --exclude is only for add pack.
func (a *app) kindWord(mac bool, cmd, own string, o *opts) fail.Code {
	if cmd != "add" && cmd != "remove" && cmd != "replace" {
		return fail.OK
	}
	kinds := "tool"
	if cmd == "add" {
		kinds = "tool or pack"
	}
	usage := usageLine(mac, cmd, own)
	if len(o.names) == 0 {
		a.usageErr(cmdWord(mac, cmd)+" needs the kind word "+kinds, usage)
		return fail.Usage
	}
	switch kind := o.names[0]; {
	case kind == "tool" || (kind == "pack" && cmd == "add"):
		o.kind, o.names = kind, o.names[1:]
	case kind == "pack":
		a.usageErr(cmdWord(mac, cmd)+" takes the kind word tool only - there is no "+cmd+" pack", usage)
		return fail.Usage
	default:
		a.fail(cmdWord(mac, cmd)+" needs the kind word "+kinds+" before the names", "use: "+cmdName(mac, cmd+" tool")+" "+strings.Join(o.names, " "))
		return fail.Usage
	}
	if len(o.exclude) > 0 && o.kind != "pack" {
		a.usageErr("--exclude is only for "+cmdWord(mac, "add pack"), usage)
		return fail.Usage
	}
	if len(o.names) == 0 && cmd != "replace" {
		a.usageErr(cmdWord(mac, cmd+" "+o.kind)+" needs at least one "+o.kind, usage)
		return fail.Usage
	}
	return fail.OK
}

// parse reads the flags of <flags> (names without dashes) and, with
// <names>, the arguments.
func parse(cmd string, args []string, flags string, names bool) (opts, error) {
	o := opts{values: map[string]string{}}
	has := map[string]bool{}
	for _, f := range strings.Fields(flags) {
		has[f] = true
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, value, withValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if !strings.HasPrefix(arg, "-") {
			if !names {
				return o, fmt.Errorf("%s takes no argument %s", cmd, arg)
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
			return o, fmt.Errorf("%s has no flag %s", cmd, arg)
		}
		takes := name == "answer" || name == "exclude" || name == "jobs" || name == "config" || name == "cache-dir" ||
			name == "url" || name == "clone" || name == "repo-path"
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
		case "no-tui":
			o.noTUI = true
		case "add":
			o.add = true
		case "exclude":
			if value == "" {
				return o, fmt.Errorf("--exclude needs a value")
			}
			o.exclude = append(o.exclude, value)
		case "json":
			o.json = true
		case "no-bootstrap":
			o.noBootstrap = true
		case "verbose":
			o.verbose = true
		case "answer":
			o.answers = append(o.answers, value)
		case "jobs":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return o, fmt.Errorf("--jobs needs a number above 0, not %s", value)
			}
			// applySettings sets BOOTSTRAP_JOBS, which the script runner reads.
			o.values[name] = value
		case "config", "cache-dir", "url", "clone", "repo-path":
			if value == "" {
				return o, fmt.Errorf("--%s needs a value", name)
			}
			o.values[name] = value
		}
	}
	return o, nil
}

// printer gives the Printer of the command: it writes to stderr.
func (a *app) printer() report.Printer { return report.NewPrinter(a.stderr, a.styled) }

// errorf prints an error line, with no fix.
func (a *app) errorf(format string, args ...any) {
	a.printer().ErrorFix(fmt.Sprintf(format, args...), "")
}

// fail prints an error line with the line of its fix.
func (a *app) fail(problem, fix string) { a.printer().ErrorFix(problem, fix) }

func (a *app) infof(format string, args ...any) { a.printer().Info(fmt.Sprintf(format, args...)) }
func (a *app) warnf(format string, args ...any) { a.printer().Warn(fmt.Sprintf(format, args...)) }

// usageErr prints a usage error: the problem, the usage line of the command
// when it has one, and the pointer to the help.
func (a *app) usageErr(problem, usage string) {
	p := a.printer()
	p.ErrorFix(problem, "")
	if usage != "" {
		fmt.Fprintf(a.stderr, "  usage: %s\n", usage)
	}
	p.Fix("Run bootstrap --help for all commands")
}

// logErr prints each message of <err> and gives its code. A problem of core
// prints its message and its fix; a problem with no message prints nothing,
// because the output of the sync shows it.
func (a *app) logErr(err error) fail.Code {
	if err == nil {
		return fail.OK
	}
	a.printErr(err)
	return fail.CodeOf(err)
}

func (a *app) printErr(err error) {
	if j, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range j.Unwrap() {
			a.printErr(e)
		}
		return
	}
	var p *core.Problem
	if errors.As(err, &p) {
		if p.Msg != "" {
			a.fail(p.Msg, p.Fix)
		}
		return
	}
	for _, m := range source.Messages(err) {
		a.printer().Error(m.Error())
	}
}

// show gives <p> relative to the current folder when it is below it.
func (a *app) show(p string) string { return target.Show(p, a.pwd) }

// prepare does the preflight of a command of the side, and finds its
// folder and its values file.
func (a *app) prepare(mac, initCmd bool) (core.Side, fail.Code) {
	sd, err := core.Prepare(mac, initCmd, a.pwd, a.getenv)
	if err != nil {
		return sd, a.logErr(err)
	}
	return sd, fail.OK
}

// cloneText gives the clone line of a command: the short commit of the
// clone and its folder, with the home folder as `~`. The commands sync and
// mac check print their own line.
func (a *app) cloneText(c *core.Clone) string {
	commit, ok := core.Git("-C", c.Dir, "rev-parse", "--short", "HEAD")
	if !ok || commit == "" {
		commit = "an unknown commit"
	}
	where := c.Repo
	if where == "" {
		where = c.Dir
	}
	if home := a.getenv("HOME"); home != "" && (where == home || strings.HasPrefix(where, home+"/")) {
		where = "~" + strings.TrimPrefix(where, home)
	}
	return "bootstrap " + commit + " · " + where
}

// openClone takes the clone of the command with core.OpenClone, and prints
// its notes and its errors. The caller defers Cleanup. With <hold>, the lock
// stays held to the end. With <line>, it prints the clone line.
func (a *app) openClone(hold, line bool) (*core.Clone, fail.Code) {
	c, err := core.OpenClone(core.CloneOptions{
		Getenv:  a.getenv,
		Version: a.version,
		Hold:    hold,
		Note:    func(s string) { a.infof("%s", s) },
		Stderr:  a.stderr,
	})
	if err != nil {
		return c, a.logErr(err)
	}
	if line {
		a.printer().Clone(a.cloneText(c))
		// The sync of the command does not print a second clone line.
		report.CloneShown = true
	}
	return c, fail.OK
}

// sync is `bootstrap [mac] sync`.
func (a *app) sync(mac bool, o opts) fail.Code {
	w, code := a.prepare(mac, false)
	if code != fail.OK {
		return code
	}
	c, code := a.openClone(mac && !o.dryRun, false)
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	run := core.FrontEndOf(mac)
	args := core.SyncArgs(mac, o.dryRun, o.noBootstrap)
	skip := core.ValuesSkip(w.Values)
	if o.dryRun {
		planned, code := run(args, a.stdout, a.stderr)
		a.warnDirty(core.DirtyTargets(core.PlanPaths(planned), skip))
		return code
	}
	// The plan, then the write. A repository without a change that is not
	// committed needs no plan.
	if mac || core.DirtyBesides(w.Root, skip) {
		planned, code := run(append(args, "--dry-run"), io.Discard, io.Discard)
		if code == fail.OK {
			if code := a.stopDirty(core.DirtyTargets(core.PlanPaths(planned), skip)); code != fail.OK {
				return code
			}
		}
	}
	_, code = run(args, a.stdout, a.stderr)
	return code
}

// check is `bootstrap [mac] check`.
func (a *app) check(mac bool, o opts) fail.Code {
	w, code := a.prepare(mac, false)
	if code != fail.OK {
		return code
	}
	c, code := a.openClone(false, !mac)
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	a.warnNotSelectable(mac, c.Dir, w.Values)
	if mac {
		o.dryRun = true
		return macos.Run(core.SyncArgs(true, o.dryRun, o.noBootstrap), a.stdout, a.stderr)
	}
	return repo.Run([]string{"--check"}, a.stdout, a.stderr)
}

// warnNotSelectable prints a warning for each tool with selectable = false
// that the values file <file> names in its [tools] table. A problem of the
// clone or of the file shows in the check itself.
func (a *app) warnNotSelectable(mac bool, dir, file string) {
	side := source.Repo
	if mac {
		side = source.Macos
	}
	raw, err := source.ReadTOML(file)
	if err != nil {
		return
	}
	tools, ok := raw.Get("tools").(*source.Obj)
	if !ok {
		return
	}
	idx, _ := source.Source{Root: dir, Side: side}.Open()
	for _, t := range tools.Keys() {
		if m := idx.Manifests[t]; m != nil && !m.Selectable {
			a.warnf("%s is not selectable - remove [tools.%s] from %s; a tool that needs it selects it", t, t, a.show(file))
		}
	}
}
