package macos

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/header"
	"github.com/virajp/bootstrap/sync-engine/internal/manifest"
)

// exists is the bash test -e: the path exists after symlinks.
func exists(path string) bool { _, err := os.Stat(path); return err == nil }

// isDir is the bash test -d: a folder after symlinks.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func hasHeader(path string) bool { return header.HasFile(path) }

// pwd is $PWD of bash: the PWD variable when it names the current folder,
// else the physical path of the current folder.
func pwd(getenv func(string) string) string {
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

// lookPath is `command -v <name>` of bash: the first executable file of that
// name in the folders of PATH. An empty folder is the current folder.
func lookPath(name, path string) string {
	for _, dir := range strings.Split(path, ":") {
		if dir == "" {
			dir = "."
		}
		p := dir + "/" + name
		info, err := os.Stat(p)
		if err != nil || info.IsDir() || syscall.Access(p, 1) != nil {
			continue
		}
		return p
	}
	return ""
}

// checkMise is `_ss_check_mise`: mise must be the standalone install at
// ~/.local/bin/mise, never a copy of a package manager. Each link is
// followed, so a link into another folder also fails. It gives 2, after a
// message.
func (s *state) checkMise() fail.Code {
	real := ""
	if found := lookPath("mise", s.getenv("PATH")); found != "" {
		real = s.abs(found)
		if r, ok := manifest.Real(real); ok {
			real = r
		}
	}
	if !strings.HasSuffix(real, "/.local/bin/mise") {
		shown := real
		if shown == "" {
			shown = "no mise on PATH"
		}
		s.log("mise must be installed with `curl https://mise.run | sh` at ~/.local/bin/mise, not from a package manager (found: %s)", shown)
		return fail.Usage
	}
	return fail.OK
}

// output runs <name> with <args> and gives its stdout without the trailing
// newlines, as a bash command substitution. Its stderr is dropped.
func output(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\n"), err
}

// gitHead is `git -C <dir> rev-parse HEAD`, or "" when it fails.
func gitHead(dir string) string {
	out, err := output("", "git", "-C", dir, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return out
}

// loginShell is `_ss_login_shell`: the login shell as dscl reports it, or ""
// when dscl cannot tell.
func (s *state) loginShell() string {
	out, _ := output("", "dscl", ".", "-read", s.getenv("HOME"), "UserShell")
	return out
}

// bootstrap is `_ss_bootstrap`: `mise bootstrap --yes` in $HOME. It gives 1,
// after a message, when it fails.
func (s *state) bootstrap() fail.Code {
	s.log("running mise bootstrap --yes")
	rc := 127
	if mise := lookPath("mise", s.getenv("PATH")); mise != "" {
		rc = 0
		cmd := exec.Command(mise, "bootstrap", "--yes")
		cmd.Dir = s.getenv("HOME")
		cmd.Stdout = s.stdout
		cmd.Stderr = s.stderr
		if err := cmd.Run(); err != nil {
			rc = 1
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				rc = ee.ExitCode()
				if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
					rc = 128 + int(ws.Signal())
				}
			}
		}
	}
	if rc != 0 {
		s.log("mise bootstrap failed with exit code %d", rc)
		return fail.Other
	}
	return fail.OK
}
