// Command bootstrap syncs the shared files of bootstrap into a repository
// (init, sync, check, questions) or into the home folder of a Mac (the same
// commands below `mac`).
//
// Contract: docs/design/sync-engine.md. internal/cli is the command line,
// internal/repo the repo front-end and internal/macos the Mac front-end.
package main

import (
	"io"
	"os"

	"github.com/virajp/bootstrap/sync-engine/internal/cli"
	"github.com/virajp/bootstrap/sync-engine/internal/fail"
)

// version is the release version. The release build sets it with
// -ldflags "-X main.version=X.Y.Z". A local build prints "dev".
var version = "dev"

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)))
}

// run runs the command of <args> and gives its exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) fail.Code {
	return cli.Run(version, args, stdin, stdout, stderr)
}
