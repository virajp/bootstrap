// Command sync-engine syncs the shared files of bootstrap into a repository
// (repo) or into the home folder of a Mac (macos).
//
// Contract: docs/design/sync-engine.md. Phase P6: repo only parses its flags.
// macos is the Mac front-end of internal/macos.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/macos"
)

// version is the release version. The release build sets it with
// -ldflags "-X main.version=X.Y.Z". A local build prints "dev".
var version = "dev"

const usage = `usage: sync-engine <command> [flags]

commands:
  repo [--dry-run | --check]          sync the repository of the current folder
  macos [--dry-run] [--no-bootstrap]  sync the Mac
  version                             print the version
`

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdout, os.Stderr)))
}

// run runs the command of <args> and gives its exit code.
func run(args []string, stdout, stderr io.Writer) fail.Code {
	// macos prints its own usage and errors, as setup_sync of bash.
	if len(args) > 0 && args[0] == "macos" {
		return macos.Run(args[1:], stdout, stderr)
	}
	err := dispatch(args, stdout)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(stdout, usage)
		return fail.OK
	}
	if err != nil {
		fmt.Fprintf(stderr, "sync-engine: %v\n", err)
		if fail.CodeOf(err) == fail.Usage {
			fmt.Fprint(stderr, usage)
		}
	}
	return fail.CodeOf(err)
}

func dispatch(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fail.New(fail.Usage, "no command")
	}
	switch cmd, rest := args[0], args[1:]; cmd {
	case "repo":
		return repo(rest)
	case "version":
		if len(rest) != 0 {
			return fail.New(fail.Usage, "version: unknown argument %s", rest[0])
		}
		fmt.Fprintln(stdout, version)
		return nil
	case "-h", "-help", "--help", "help":
		return flag.ErrHelp
	default:
		return fail.New(fail.Usage, "unknown command %s", cmd)
	}
}

// parse parses <args> with <fs>. A bad flag or an argument is a usage error.
func parse(fs *flag.FlagSet, args []string) error {
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return fail.New(fail.Usage, "%s: %v", fs.Name(), err)
	}
	if fs.NArg() != 0 {
		return fail.New(fail.Usage, "%s: unknown argument %s", fs.Name(), fs.Arg(0))
	}
	return nil
}

func repo(args []string) error {
	fs := flag.NewFlagSet("repo", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", false, "show the changes, write nothing")
	check := fs.Bool("check", false, "examine the values file only")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *dryRun && *check {
		return fail.New(fail.Usage, "repo: --check and --dry-run cannot be used together")
	}
	return fail.New(fail.Other, "repo: not implemented")
}
