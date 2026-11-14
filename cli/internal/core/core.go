// Package core is the shared core of the selection commands: the preflight
// of a side, the values file and its selection, the change list with Plan
// and Apply, the scan of detect, and the check of the clean-target rule. The
// command line (internal/cli) and the TUI call it.
//
// A function of core gives values and errors only: it never prints, never
// prompts and never reads a terminal. The one output that it passes on is
// the output of the sync front-end, to the writers of Options.
//
// Contract: docs/design/cli.md.
package core

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/macos"
	"github.com/virajp/bootstrap/cli/internal/repo"
	"github.com/virajp/bootstrap/cli/internal/target"
)

// Problem is a failure with its exit code, its message and the line of its
// fix. A Problem with no message is a failure that the output of the sync
// front-end already shows.
type Problem struct {
	Code fail.Code
	Msg  string
	Fix  string
}

func (p *Problem) Error() string {
	if p.Fix == "" {
		return p.Msg
	}
	return p.Msg + " - " + p.Fix
}

// Unwrap gives the problem as a *fail.Error, so fail.CodeOf finds its code.
func (p *Problem) Unwrap() error { return &fail.Error{Code: p.Code, Msg: p.Msg} }

func problem(code fail.Code, msg, fix string) *Problem {
	return &Problem{Code: code, Msg: msg, Fix: fix}
}

// Side is the folder and the values file of a command, and its clone.
type Side struct {
	Mac    bool
	Root   string // The repository; on the Mac, the repository of macos.toml.
	Values string // The values file.
	Clone  string // The folder of the clone of bootstrap.
	Pwd    string // The current folder: a message shows a path relative to it.
}

// Show gives <p> relative to the current folder when it is below it.
func (sd Side) Show(p string) string { return target.Show(p, sd.Pwd) }

// Home gives the folder where the scripts and the detect scan run: $HOME on
// the Mac, else the repository.
func (sd Side) Home() string {
	if sd.Mac {
		return os.Getenv("HOME")
	}
	return sd.Root
}

// CmdName gives the command as the user types it.
func CmdName(mac bool, verb string) string { return "bootstrap " + CmdWord(mac, verb) }

// CmdWord gives the command word, with mac on the Mac side.
func CmdWord(mac bool, cmd string) string {
	if mac {
		return "mac " + cmd
	}
	return cmd
}

// Prepare is the preflight of a command of the side: it finds its folder and
// its values file. <initCmd> is init, which runs in the current folder <pwd>
// on the Mac, and which tells the user to create a git repository. It sets
// MISE_PROJECT_ROOT or MACOS_SETUP_REPO_PATH, which the front-ends read.
func Prepare(mac, initCmd bool, pwd string, getenv func(string) string) (Side, error) {
	sd := Side{Mac: mac, Pwd: pwd}
	if _, err := exec.LookPath("mise"); err != nil {
		return sd, problem(fail.Usage, "mise is not available", "install it from https://mise.run")
	}
	if mac {
		dir := getenv("MACOS_SETUP_REPO_PATH")
		if initCmd {
			dir = pwd
		} else if dir == "" {
			return sd, problem(fail.Usage, "MACOS_SETUP_REPO_PATH is not set", "set it to the repository of macos.toml, or run bootstrap mac init in that repository")
		}
		if _, ok := Toplevel(dir); !ok {
			if initCmd {
				return sd, problem(fail.Usage, dir+" is not a git repository", "create one with git init, then run bootstrap mac init again")
			}
			return sd, problem(fail.Usage, "MACOS_SETUP_REPO_PATH "+dir+" is not a git repository", "")
		}
		_ = os.Setenv("MACOS_SETUP_REPO_PATH", dir)
		sd.Root, sd.Values = dir, strings.TrimSuffix(dir, "/")+"/macos.toml"
		return sd, nil
	}
	root := getenv("MISE_PROJECT_ROOT")
	dir := root
	if dir == "" {
		dir = pwd
	}
	top, ok := Toplevel(dir)
	if !ok {
		if initCmd {
			return sd, problem(fail.Usage, dir+" is not a git repository", "create one with git init, then run bootstrap init again")
		}
		return sd, problem(fail.Usage, dir+" is not a git repository", "")
	}
	if root == "" {
		root = top
		_ = os.Setenv("MISE_PROJECT_ROOT", root)
	}
	sd.Root = root
	sd.Values = getenv("BOOTSTRAP_VALUES")
	if sd.Values == "" {
		sd.Values = root + "/.config/bootstrap.toml"
	}
	return sd, nil
}

// Options are the options of Plan and Apply.
type Options struct {
	// DryRun: the command writes nothing. replace then gives a warning for
	// a work tree that is not clean, not a problem.
	DryRun bool
	// NoBootstrap skips each mise bootstrap step of the Mac sync.
	NoBootstrap bool
	// Stdout and Stderr take the output of the sync front-end: the sync of
	// Apply, and the dry run of a plan that fails. Nil discards it.
	Stdout, Stderr io.Writer
	// Step tells each step of Apply before the sync, when it is not nil.
	Step func(Step)
}

func (o Options) writers() (io.Writer, io.Writer) {
	out, errw := o.Stdout, o.Stderr
	if out == nil {
		out = io.Discard
	}
	if errw == nil {
		errw = io.Discard
	}
	return out, errw
}

// FrontEnd is the sync of a side, with its change list.
type FrontEnd func(args []string, stdout, stderr io.Writer) ([]target.Planned, fail.Code)

// FrontEndOf gives the sync of the side.
func FrontEndOf(mac bool) FrontEnd {
	if mac {
		return macos.RunWith
	}
	return repo.RunWith
}

// frontEnd gives the sync that Plan and Apply run. Tests replace it.
var frontEnd = FrontEndOf

// SyncArgs are the arguments of the front-end.
func SyncArgs(mac, dryRun, noBootstrap bool) []string {
	var args []string
	if dryRun {
		args = append(args, "--dry-run")
	}
	if mac && noBootstrap {
		args = append(args, "--no-bootstrap")
	}
	return args
}

// PlanPaths gives the path of each file of a change list.
func PlanPaths(list []target.Planned) []string {
	paths := make([]string, len(list))
	for i, p := range list {
		paths[i] = p.Path
	}
	return paths
}
