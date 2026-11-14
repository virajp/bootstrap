package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// GitEnv is the environment of a git call: without the variables of a git
// hook, and with no prompt for a network call.
func GitEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if k == "GIT_DIR" || k == "GIT_WORK_TREE" || k == "GIT_INDEX_FILE" {
			continue
		}
		env = append(env, kv)
	}
	return append(env, "GIT_TERMINAL_PROMPT=0", "GIT_SSH_COMMAND=ssh -o BatchMode=yes")
}

// Git runs git with <args> and gives its stdout without the last newline.
func Git(args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Env = GitEnv()
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\n"), err == nil
}

// Toplevel gives the root of the git work tree of the folder <dir>.
func Toplevel(dir string) (string, bool) {
	top, ok := Git("-C", dir, "rev-parse", "--show-toplevel")
	return top, ok && top != ""
}

// InSubmodule tells if the repository <root> is a git submodule.
func InSubmodule(root string) bool {
	sub, _ := Git("-C", root, "rev-parse", "--show-superproject-working-tree")
	return sub != ""
}

// Exists tells if the path <p> exists.
func Exists(p string) bool { _, err := os.Stat(p); return err == nil }

// realDir gives the real path of the nearest folder of <p> that exists, and
// the rest of <p> below it.
func realDir(p string) (string, string) {
	dir, rest := filepath.Dir(p), filepath.Base(p)
	for !Exists(dir) && dir != "/" && dir != "." {
		rest = filepath.Join(filepath.Base(dir), rest)
		dir = filepath.Dir(dir)
	}
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	return dir, rest
}

// DirtyTargets is the check of the clean-target rule: it gives each path of
// <paths> that is inside a git work tree and has changes that are not
// committed there: modified, staged or untracked. A path of <skip> is not
// examined.
func DirtyTargets(paths []string, skip map[string]bool) []string {
	tops := map[string]string{}            // A real folder to its work tree.
	status := map[string]map[string]bool{} // A work tree to its dirty paths.
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		if skip[p] || seen[p] {
			continue
		}
		seen[p] = true
		dir, rest := realDir(p)
		top, ok := tops[dir]
		if !ok {
			top, _ = Toplevel(dir)
			if r, err := filepath.EvalSymlinks(top); err == nil {
				top = r
			}
			tops[dir] = top
		}
		if top == "" {
			continue
		}
		set, ok := status[top]
		if !ok {
			set = statusSet(top)
			status[top] = set
		}
		rel, err := filepath.Rel(top, filepath.Join(dir, rest))
		if err == nil && set[rel] {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// DirtyBesides tells if the work tree of <root> has a change that is not
// committed in a path that is not in <skip>.
func DirtyBesides(root string, skip map[string]bool) bool {
	top, ok := Toplevel(root)
	if !ok {
		return false
	}
	if r, err := filepath.EvalSymlinks(top); err == nil {
		top = r
	}
	for p := range statusSet(top) {
		if !skip[top+"/"+p] {
			return true
		}
	}
	return false
}

// ValuesSkip gives the values file and its real path: sync writes only
// bootstrap_commit there, so the clean-target rule skips it.
func ValuesSkip(values string) map[string]bool {
	skip := map[string]bool{values: true}
	if r, err := filepath.EvalSymlinks(values); err == nil {
		skip[r] = true
	}
	return skip
}

// statusSet gives the paths with changes that are not committed in the work
// tree <top>, relative to it.
func statusSet(top string) map[string]bool {
	set := map[string]bool{}
	out, ok := Git("-C", top, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if !ok {
		return set
	}
	entries := strings.Split(out, "\x00")
	for i := 0; i < len(entries); i++ {
		e := entries[i]
		if len(e) < 4 {
			continue
		}
		set[e[3:]] = true
		// A rename or a copy: the next entry is the old path.
		if e[0] == 'R' || e[0] == 'C' {
			i++
			if i < len(entries) {
				set[entries[i]] = true
			}
		}
	}
	return set
}
