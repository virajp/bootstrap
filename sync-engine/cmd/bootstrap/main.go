// Command bootstrap syncs the shared files of bootstrap into a repository
// (sync, check) or into the home folder of a Mac (mac sync, mac check).
//
// Contract: docs/design/sync-engine.md. sync and check are the repo front-end
// of internal/repo, and the mac group is the Mac front-end of internal/macos.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/macos"
	"github.com/virajp/bootstrap/sync-engine/internal/repo"
)

// version is the release version. The release build sets it with
// -ldflags "-X main.version=X.Y.Z". A local build prints "dev".
var version = "dev"

const usage = `usage: bootstrap <command> [flags]

commands:
  sync [--dry-run]                       sync the repository of the current folder
  check                                  examine the values file of the repository
  mac sync [--dry-run] [--no-bootstrap]  sync the Mac
  mac check [--no-bootstrap]             show what mac sync changes, and write nothing
  version                                print the version
`

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdout, os.Stderr)))
}

// run runs the command of <args> and gives its exit code.
func run(args []string, stdout, stderr io.Writer) fail.Code {
	// The front-ends print their own usage and errors, as repo_sync and
	// setup_sync of bash.
	if len(args) > 0 {
		switch cmd, rest := args[0], args[1:]; cmd {
		case "sync":
			return repo.Run(rest, stdout, stderr)
		case "check":
			return repo.Run(append([]string{"--check"}, rest...), stdout, stderr)
		case "mac":
			if len(rest) > 0 && rest[0] == "sync" {
				return macos.Run(rest[1:], stdout, stderr)
			}
			if len(rest) > 0 && rest[0] == "check" {
				return macos.Run(append([]string{"--dry-run"}, rest[1:]...), stdout, stderr)
			}
		}
	}
	err := dispatch(args, stdout)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(stdout, usage)
		return fail.OK
	}
	if err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
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
	case "version":
		if len(rest) != 0 {
			return fail.New(fail.Usage, "version: unknown argument %s", rest[0])
		}
		fmt.Fprintln(stdout, version)
		return nil
	case "mac":
		if len(rest) == 0 {
			return fail.New(fail.Usage, "mac: no command")
		}
		return fail.New(fail.Usage, "mac: unknown command %s", rest[0])
	case "-h", "-help", "--help", "help":
		return flag.ErrHelp
	default:
		return fail.New(fail.Usage, "unknown command %s", cmd)
	}
}
