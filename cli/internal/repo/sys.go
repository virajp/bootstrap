package repo

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/virajp/bootstrap/cli/internal/source"
)

// exists is the bash test -e: the path exists after symlinks.
func exists(path string) bool { _, err := os.Stat(path); return err == nil }

// isFile is the bash test -f: a regular file after symlinks.
func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// isLink is the bash test -L.
func isLink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

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

// command runs <name> with <args> in the folder <dir> ("" is the current
// folder). <stderr> gets the error output of the command; nil drops it. It
// gives the stdout without the trailing newlines, as a bash command
// substitution, and false when the command fails.
func command(dir string, stderr io.Writer, name string, args ...string) (string, bool) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stderr = stderr
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\n"), err == nil
}

// run runs <name> with <args> in the folder <dir> and drops its output. It
// gives true when the command succeeds.
func run(dir string, name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run() == nil
}

// git runs git in the folder of the repository <root>, as `git -C <root>`,
// and drops its error output.
func git(root string, args ...string) (string, bool) {
	return command("", nil, "git", append([]string{"-C", root}, args...)...)
}

// bashLines splits <s> as `while read` does: each line, also a last line
// without a newline.
func bashLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// fields is the word split of bash with the default IFS.
func fields(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' })
}

// readTabs is `IFS=$'\t' read -r a b c`: the tabs at the start and the end go,
// a run of tabs splits two fields, and c keeps the rest of the line.
func readTabs(line string) (a, b, c string) {
	line = strings.Trim(line, "\t")
	field := func() string {
		i := strings.IndexByte(line, '\t')
		if i < 0 {
			f := line
			line = ""
			return f
		}
		f := line[:i]
		line = strings.TrimLeft(line[i:], "\t")
		return f
	}
	a = field()
	b = field()
	return a, b, line
}

// copyFile is `cp -- <src> <dst>`: the content of <src> into <dst>. An
// existing <dst> keeps its mode.
func copyFile(src, dst string) bool {
	b, err := os.ReadFile(src)
	if err != nil {
		return false
	}
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o644)
	if err != nil {
		return false
	}
	_, err = f.Write(b)
	return f.Close() == nil && err == nil
}

// gomplateError prints <msg> as the JSON log line of a failed gomplate.
func gomplateError(w io.Writer, msg string) {
	_, _ = fmt.Fprintf(w, `{"time":"%s","level":"ERROR","msg":"","err":%s}`+"\n",
		time.Now().Format("2006-01-02T15:04:05.999999Z07:00"), source.JSON(msg))
}
