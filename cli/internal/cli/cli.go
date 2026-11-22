// Package cli is the command line of `bootstrap`: the commands init, tui,
// sync, check, questions, add, remove, replace, detect and list of the repo
// side and of the Mac side (`bootstrap mac <command>`), config, settings,
// contract, catalog, describe, lsp and version. It is the shell of the command
// line: it parses the arguments, takes the clone and its lock, checks the
// version, calls internal/core (the preflight, Plan, Apply, Detect and the
// clean-target rule), prints the results and the errors, and asks the
// questions and the consent, or starts the TUI of internal/tui. The
// front-ends internal/repo and internal/macos do the sync itself.
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
	"text/tabwriter"

	"github.com/virajp/bootstrap/cli/internal/contract"
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
  migrate              migrate an old form, or create the values file with
                       the members of the pack base, ask the questions
                       without an answer, then sync
  tui                  open the TUI on the values file of the repository;
                       when migrate has a change to make, run migrate first
  sync                 sync the repository of the current folder; when
                       migrate has a change to make, migrate first
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
  config get <key>     print the value of a key of .config/bootstrap/, with
                       its source (file, default, contributed, derived or
                       unselected); exit 2 for a tools.* key with no catalog
  config set <key> <value>
                       write a key of .config/bootstrap/
  config unset <key>   remove a key of .config/bootstrap/
  config show [repo|tools|releases]
                       print the values of .config/bootstrap/
  settings             print each setting of the engine, with its value and
                       its source
  contract             print the versions of the JSON contract, the commands
                       that take --json, and the names of the task ports
  catalog              print the tools of the clone by category, with the
                       categories and the packs; it needs no repository
  describe             print the values of the repository, its selected
                       tools, its members, its task ports and its language
                       servers
  lsp                  print the language servers of the selected tools of
                       the repository
  version, --version   print the version

Mac
  mac migrate          migrate for the Mac: ~/.config/bootstrap/macos.toml
  mac tui              open the TUI on macos.toml
  mac sync             sync the Mac from ~/.config/bootstrap/macos.toml
  mac check            show what mac sync changes, and write nothing
  mac questions [tool...]
                       the questions of the Mac tools
  mac add tool|pack, mac remove tool, mac replace tool, mac detect
                       the selection commands for macos.toml
  mac list, mac ls     print the selected Mac tools as a tree

Flags
  --dry-run            migrate, sync, add, remove, replace, detect: show the
                       changes, and write nothing
  --yes, -y            migrate, add, remove, replace, detect --add: write the
                       changes without a prompt
  --add                detect: add each matched tool that is not selected,
                       after the consent
  --exclude <member>   add pack: leave out a member of the pack, as
                       packs.toml writes it (repeatable; one pack only)
  --answer <tool>.<key>=<value>
                       migrate, sync, add, replace, detect --add: the answer
                       of a question (repeatable); --answer <tool>.<key>= writes
                       an empty answer for an optional or a group question
  --json               contract, catalog, describe, lsp, check, questions, mac
                       questions, config get|set|unset|show, migrate, sync,
                       add, remove, replace, detect: print one JSON envelope
                       on stdout, and each other line on stderr; migrate,
                       add, remove, replace and detect --add also need --yes
                       or --dry-run
  --format json|yaml|toml
                       config get, config show: the output format (default:
                       json when stdout is not a terminal, else toml)
  --no-bootstrap       mac migrate, mac sync: skip the mise bootstrap of the
                       first setup of a Mac
  --jobs <n>           the jobs at a time (default: 2 x the CPU count)
  --verbose            migrate, sync, check, add, remove, replace, detect: print
                       the log line of each target
  --config <path>      each command except version and contract: the config
                       file (default: ~/.config/bootstrap/config.toml)
  --cache-dir <path>   each command except version and contract: the cache
                       folder
  --url <url>          each command except version and contract: the clone URL
  --clone <path>       each command except version and contract: a clone of
                       bootstrap, with no fetch
`

// argsOf gives the arguments of a command in its usage line.
var argsOf = map[string]string{
	"questions": " [tool...]", "add": " tool|pack <name...>",
	"remove": " tool <tool...>", "replace": " tool <old> <new>",
	"config": " get|set|unset|show [<key>] [<value>]",
}

// flagUsage gives each flag of a usage line, by the name of the flag.
var flagUsage = map[string]string{
	"answer": "[--answer <tool>.<key>=<value>]", "jobs": "[--jobs <n>]",
	"exclude": "[--exclude <member>]",
	"format":  "[--format json|yaml|toml]",
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

// allowed gives the own flags of each command; a name that it does not have
// is not a command. The Mac side takes json only on questions.
var allowed = map[string]string{
	"migrate":   "dry-run yes answer jobs verbose json",
	"tui":       "jobs verbose",
	"sync":      "dry-run answer jobs verbose json",
	"check":     "jobs verbose json",
	"questions": "json",
	"add":       "dry-run yes exclude answer jobs verbose json",
	"remove":    "dry-run yes jobs verbose json",
	"replace":   "dry-run yes answer jobs verbose json",
	"detect":    "dry-run add yes answer jobs verbose json",
	"list":      "",
	"settings":  "jobs",
	"config":    "format json",
	"contract":  "json",
	"catalog":   "json",
	"describe":  "json",
	"lsp":       "json",
	"version":   "",
}

// commandOrder is the order of the commands in the commands of contract.
var commandOrder = []string{
	"contract", "catalog", "describe", "lsp", "migrate", "tui", "sync", "check",
	"questions", "add", "remove", "replace", "detect", "list", "config",
	"settings", "version",
}

// writeCommands are the commands that plan and write changes: with --json,
// each human line of them goes to stderr.
var writeCommands = map[string]bool{
	"migrate": true, "sync": true, "add": true, "remove": true, "replace": true, "detect": true,
}

// writes tells if the command <cmd> with <o> writes after a consent: migrate,
// add, remove, replace and detect --add. sync writes with no consent.
func writes(cmd string, o opts) bool {
	return writeCommands[cmd] && cmd != "sync" && (cmd != "detect" || o.add)
}

// configSubs are the subcommands of config.
var configSubs = []string{"get", "set", "unset", "show"}

// repoOnly are the commands that the Mac side does not have.
var repoOnly = map[string]bool{
	"version": true, "settings": true, "config": true, "contract": true,
	"catalog": true, "describe": true, "lsp": true,
}

// takesJSON tells if the command <cmd> of the side takes --json: on the Mac
// side, only questions.
func takesJSON(mac bool, cmd string) bool {
	if mac && cmd != "questions" {
		return false
	}
	return hasWord(allowed[cmd], "json")
}

// jsonCommands gives the full words of each command that takes --json, for
// example "config get" and "mac questions".
func jsonCommands() []string {
	var out []string
	for _, cmd := range commandOrder {
		if !takesJSON(false, cmd) {
			continue
		}
		if cmd == "config" {
			for _, sub := range configSubs {
				out = append(out, "config "+sub)
			}
			continue
		}
		out = append(out, cmd)
	}
	for _, cmd := range commandOrder {
		if takesJSON(true, cmd) {
			out = append(out, cmdWord(true, cmd))
		}
	}
	return out
}

func hasWord(list, w string) bool {
	for _, f := range strings.Fields(list) {
		if f == w {
			return true
		}
	}
	return false
}

// withoutWord gives <list> without the word <w>.
func withoutWord(list, w string) string {
	var out []string
	for _, f := range strings.Fields(list) {
		if f != w {
			out = append(out, f)
		}
	}
	return strings.Join(out, " ")
}

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
	// progress: stderr is a terminal and the command has no --json, so a
	// slow step shows a spinner.
	progress bool
	// json: the arguments hold --json, so the command prints the envelope on
	// stdout and each human line on stderr. command is the full command words
	// of the envelope, data its result, and errors and warnings its items.
	json             bool
	command          string
	data             any
	errors, warnings []contract.Item
}

// opts are the flags and the arguments of a command.
type opts struct {
	dryRun, yes, json, noBootstrap, help bool
	verbose, add                         bool
	// auto: the command writes its changes with no consent, as sync does
	// when it migrates first.
	auto    bool
	answers []string
	kind    string            // add, remove, replace: the kind word, tool or pack.
	names   []string          // The names after the kind word.
	exclude []string          // add pack: the members of --exclude.
	values  map[string]string // The value flags of the settings, by flag name.
	format  string            // config get, config show: --format.
}

// Run runs the command of <args> and gives its exit code. With a literal
// --json anywhere in <args>, it prints the envelope of the command on stdout,
// also for a usage error, and its exit code is the code of the envelope.
func Run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer) fail.Code {
	a := &app{version: version, stdin: stdin, stdout: stdout, stderr: stderr, getenv: os.Getenv}
	for _, arg := range args {
		a.json = a.json || arg == "--json"
	}
	code := a.run(args)
	if !a.json {
		return code
	}
	for i := range a.errors {
		// An error line does not know the code of the command that stops.
		if a.errors[i].Code == fail.OK {
			a.errors[i].Code = code
		}
	}
	return contract.Write(stdout, contract.Envelope{
		Command: a.command, Code: code, Data: a.data, Errors: a.errors, Warnings: a.warnings,
	})
}

// run runs the command of <args> for Run.
func (a *app) run(args []string) fail.Code {
	version, stdout := a.version, a.helpOut()
	isTerm := false
	if f, ok := a.stderr.(*os.File); ok {
		isTerm = report.IsTerminal(f)
	}
	a.styled = report.StyleFromEnv(a.getenv, isTerm)
	a.progress = isTerm && !a.json
	a.pwd = currentDir(a.getenv)
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
	// init is the old name of migrate: a hidden alias, not in the help and
	// not in the commands of contract.
	if cmd == "init" {
		cmd = "migrate"
	}
	flags, ok := allowed[cmd]
	if !ok || (mac && repoOnly[cmd]) {
		switch cmd {
		case "-h", "-help", "--help", "help":
			if !mac {
				_, _ = fmt.Fprint(stdout, Usage)
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
	a.command = cmdWord(mac, cmd)
	if mac && !takesJSON(mac, cmd) {
		flags = withoutWord(flags, "json")
	}
	own := flags
	if mac && cmd != "questions" {
		flags += " no-bootstrap"
	}
	if cmd != "version" && cmd != "contract" {
		flags += " config cache-dir url clone"
	}
	if cmd == "migrate" {
		// --no-tui is an old flag of migrate, which never opens the TUI
		// now: it stays out of the help, and does nothing.
		flags += " no-tui"
	}
	takesNames := cmd == "migrate" || cmd == "questions" || cmd == "add" || cmd == "remove" || cmd == "replace" || cmd == "config"
	o, err := parse(cmdWord(mac, cmd), rest, flags, takesNames)
	if err != nil {
		a.usageErr(err.Error(), usageLine(mac, cmd, own))
		return fail.Usage
	}
	if o.help {
		_, _ = fmt.Fprint(stdout, Usage)
		return fail.OK
	}
	if code := a.kindWord(mac, cmd, own, &o); code != fail.OK {
		return code
	}
	if cmd == "migrate" && len(o.names) > 0 {
		a.fail(cmdWord(mac, "migrate")+" takes no tool names", "use: "+cmdName(mac, "add tool")+" "+strings.Join(o.names, " "))
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
	if o.json && writes(cmd, o) && !o.yes && !o.dryRun {
		// --json never prompts (D4); sync applies with no consent (D21).
		a.fail(cmdWord(mac, cmd)+" --json needs --yes or --dry-run", "add --yes to write the changes, or --dry-run to show them")
		return fail.Usage
	}
	if o.json && writeCommands[cmd] {
		// stdout holds the envelope only: the change list, the trees and the
		// output of the sync go to stderr.
		a.stdout = a.stderr
	}
	if cmd == "version" {
		_, _ = fmt.Fprintln(stdout, version)
		return fail.OK
	}
	if cmd == "contract" {
		// It takes no repository and no config file.
		return a.contractCmd(o)
	}
	// The config file, before any other work.
	if code := a.applySettings(o); code != fail.OK {
		return code
	}
	switch cmd {
	case "settings":
		return a.printSettings(o)
	case "config":
		return a.configCmd(o)
	case "catalog":
		return a.catalogCmd(o)
	case "describe":
		return a.describe(o)
	case "lsp":
		return a.lsp(o)
	case "sync":
		return a.sync(mac, o)
	case "check":
		return a.check(mac, o)
	case "migrate":
		return a.init(mac, o)
	case "tui":
		return a.tui(mac, o)
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
			name == "url" || name == "clone" || name == "format"
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
		case "add":
			o.add = true
		case "exclude":
			if value == "" {
				return o, fmt.Errorf("--exclude needs a value")
			}
			o.exclude = append(o.exclude, value)
		case "json":
			o.json = true
		case "format":
			o.format = value
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
		case "config", "cache-dir", "url", "clone":
			if value == "" {
				return o, fmt.Errorf("--%s needs a value", name)
			}
			o.values[name] = value
		}
	}
	return o, nil
}

// helpOut gives the writer of the help: stdout, or stderr under --json,
// where stdout holds the envelope only.
func (a *app) helpOut() io.Writer {
	if a.json {
		return a.stderr
	}
	return a.stdout
}

// contractCmd is `bootstrap contract`: the versions of the JSON contract,
// the version of the engine, the commands that take --json and the names of
// the task ports.
func (a *app) contractCmd(o opts) fail.Code {
	commands := jsonCommands()
	if o.json {
		a.data = struct {
			Versions   []int    `json:"versions"`
			CLIVersion string   `json:"cli_version"`
			Commands   []string `json:"commands"`
			Ports      []string `json:"ports"`
		}{contract.Supported, a.version, commands, contract.Ports}
		return fail.OK
	}
	versions := make([]string, len(contract.Supported))
	for i, v := range contract.Supported {
		versions[i] = strconv.Itoa(v)
	}
	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "versions\t%s\n", strings.Join(versions, " "))
	_, _ = fmt.Fprintf(w, "version\t%s\n", a.version)
	_, _ = fmt.Fprintf(w, "commands\t%s\n", strings.Join(commands, ", "))
	_, _ = fmt.Fprintf(w, "ports\t%s\n", strings.Join(contract.Ports, ", "))
	_ = w.Flush()
	return fail.OK
}

// record keeps an error (<warn> false) or a warning of the envelope. An
// error with code 0 takes the code of the command.
func (a *app) record(warn bool, code fail.Code, msg string) {
	if !a.json {
		return
	}
	if warn {
		a.warnings = append(a.warnings, contract.Item{Message: msg})
		return
	}
	a.errors = append(a.errors, contract.Item{Code: code, Message: msg})
}

// printer gives the Printer of the command: it writes to stderr.
func (a *app) printer() report.Printer { return report.NewPrinter(a.stderr, a.styled) }

// errorf prints an error line, with no fix.
func (a *app) errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.record(false, fail.OK, msg)
	a.printer().ErrorFix(msg, "")
}

// fail prints an error line with the line of its fix.
func (a *app) fail(problem, fix string) {
	a.record(false, fail.OK, withFix(problem, fix))
	a.printer().ErrorFix(problem, fix)
}

// withFix gives the message of an error and its fix as one line.
func withFix(problem, fix string) string {
	if fix == "" {
		return problem
	}
	return problem + " - " + fix
}

func (a *app) infof(format string, args ...any) { a.printer().Info(fmt.Sprintf(format, args...)) }
func (a *app) warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.record(true, fail.OK, msg)
	a.printer().Warn(msg)
}

// usageErr prints a usage error: the problem, the usage line of the command
// when it has one, and the pointer to the help.
func (a *app) usageErr(problem, usage string) {
	a.record(false, fail.Usage, problem)
	p := a.printer()
	p.ErrorFix(problem, "")
	if usage != "" {
		_, _ = fmt.Fprintf(a.stderr, "  usage: %s\n", usage)
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
			a.record(false, p.Code, withFix(p.Msg, p.Fix))
			a.printer().ErrorFix(p.Msg, p.Fix)
		}
		return
	}
	for _, m := range source.Messages(err) {
		a.record(false, fail.CodeOf(m), m.Error())
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
	// The fetch of the clone can take some seconds.
	stop := func() {}
	if a.getenv("BOOTSTRAP_CLONE") == "" {
		stop = a.spin("Updating the clone of bootstrap ...")
	}
	c, err := core.OpenClone(core.CloneOptions{
		Getenv:  a.getenv,
		Version: a.version,
		Hold:    hold,
		Note:    func(s string) { a.infof("%s", s) },
		Stderr:  a.stderr,
	})
	stop()
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

// sync is `bootstrap [mac] sync`. When migrate has a change to make, sync
// does the migrate first, with no consent, and its sync; else it syncs.
func (a *app) sync(mac bool, o opts) fail.Code {
	// On the repo side, an old form of the values does not stop the
	// preflight: the migrate of sync migrates it.
	w, code := a.prepare(mac, !mac)
	if code != fail.OK {
		return code
	}
	if !mac {
		// The scripts of the sync read the values with bootstrap config (D19).
		_ = os.Setenv("BOOTSTRAP_VALUES", w.Values)
	}
	c, code := a.openClone(mac && !o.dryRun, false)
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	w.Clone = c.Dir
	// A question without an answer is a change of migrate only in a
	// terminal, where sync can ask it; else the sync names it.
	need, err := core.MigrateNeeded(w, a.interactive())
	if err != nil {
		return a.logErr(err)
	}
	if need {
		a.printer().Clone(a.cloneText(c))
		report.CloneShown = true
		a.infof("%s has a change to make: it runs before the sync", cmdName(mac, "migrate"))
		o.auto = true
		code := a.finish(w, o, core.List{Init: true})
		if !mac || o.dryRun || code != fail.OK {
			return code
		}
		// mac migrate runs no sync: the sync of this command follows.
		o.answers = nil
	}
	if len(o.answers) > 0 {
		a.fail("sync --answer is only for a question of a migrate", "use: "+cmdName(mac, "config set")+" tools.<tool>.<key> <value>")
		return fail.Usage
	}
	run := core.FrontEndOf(mac)
	args := core.SyncArgs(mac, o.dryRun, o.noBootstrap)
	skip := core.ValuesSkip(w.Values)
	if o.dryRun {
		stop := a.spin("Planning ...")
		planned, _, code := run(args, a.stdout, a.stderr)
		stop()
		a.syncData(o, planned, nil)
		a.warnDirty(core.DirtyTargets(core.PlanPaths(planned), skip))
		return code
	}
	// The plan, then the write. A repository without a change that is not
	// committed needs no plan, except for the plan of --json.
	var planned []target.Planned
	if o.json || mac || core.DirtyBesides(w.Root, skip) {
		var code fail.Code
		stop := a.spin("Planning ...")
		planned, _, code = run(append(args, "--dry-run"), io.Discard, io.Discard)
		stop()
		a.syncData(o, planned, nil)
		if code == fail.OK {
			if code := a.stopDirty(core.DirtyTargets(core.PlanPaths(planned), skip)); code != fail.OK {
				return code
			}
		}
	}
	stop := a.syncSpin(false)
	_, outcomes, code := run(args, a.stdout, a.stderr)
	stop()
	a.syncData(o, planned, outcomes)
	return code
}

// syncSpin shows the spinner of a sync that writes. The first setup of a
// Mac runs mise bootstrap, which can ask for a password, so it does not
// spin (<prompts>).
func (a *app) syncSpin(prompts bool) (stop func()) {
	if prompts {
		return func() {}
	}
	return a.spin("Syncing ...")
}

// syncData keeps the data of sync --json: the files of the plan in
// data.plan, and the outcomes of the run.
func (a *app) syncData(o opts, planned []target.Planned, outcomes []report.Result) {
	if o.json {
		a.data = writeData{Plan: planDataOf(&core.PlanResult{Files: planned}), Outcomes: outcomesOf(outcomes)}
	}
}

// checkData is the data of check --json: valid is true only when the code
// is 0, and checks has one entry for each stage of the check, in order.
type checkData struct {
	Valid  bool         `json:"valid"`
	Checks []repo.Stage `json:"checks"`
}

// check is `bootstrap [mac] check`.
func (a *app) check(mac bool, o opts) fail.Code {
	if o.json {
		// A failure before the first stage keeps this data.
		a.data = checkData{Checks: []repo.Stage{}}
	}
	w, code := a.prepare(mac, false)
	if code != fail.OK {
		return code
	}
	c, code := a.openClone(false, !mac)
	defer c.Cleanup()
	if code != fail.OK {
		return code
	}
	a.warnNotSelectable(mac, c.Dir, w.EditFile())
	stop := a.spin("Checking ...")
	defer stop()
	if mac {
		o.dryRun = true
		return macos.Run(core.SyncArgs(true, o.dryRun, o.noBootstrap), a.stdout, a.stderr)
	}
	if !o.json {
		return repo.Run([]string{"--check"}, a.stdout, a.stderr)
	}
	// stdout holds the envelope only.
	stages, loose, code := repo.Check(a.stderr, a.stderr)
	for _, f := range loose {
		item := contract.Item{Code: f.Code, Message: f.Message, Path: f.Path}
		if f.Severity == repo.SevWarning {
			a.warnings = append(a.warnings, item)
		} else {
			a.errors = append(a.errors, item)
		}
	}
	if stages != nil {
		a.data = checkData{Valid: code == fail.OK, Checks: stages}
	}
	return code
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

// currentDir is $PWD of bash: the PWD variable when it names the current
// folder, else the physical path of the current folder. A parent can change
// the folder and keep its own PWD, for example `mise x --cd=<dir>`.
func currentDir(getenv func(string) string) string {
	p := getenv("PWD")
	if strings.HasPrefix(p, "/") {
		a, errA := os.Stat(p)
		b, errB := os.Stat(".")
		if errA == nil && errB == nil && os.SameFile(a, b) {
			return p
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return p
}
