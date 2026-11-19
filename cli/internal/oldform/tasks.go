package oldform

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/header"
	"github.com/virajp/bootstrap/cli/internal/manifest"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// TasksDir is the folder of the task files of a repository.
const TasksDir = ".config/mise/tasks"

// LocalPart is the name of the part that a repo's own test task becomes.
const LocalPart = "50-local"

// MoveNote is the label of the change of init that moves the own test tasks
// of a repo into parts (D25).
const MoveNote = "move own test tasks into 50-local parts"

// TaskPorts are the paths of the test ports where a repo can have its own
// task, each with the part that init moves it to (D4, D7, D12). The paths
// are relative to the root of the repository.
var TaskPorts = []Move{
	{From: TasksDir + "/test", To: TasksDir + "/test/" + LocalPart},
	{From: TasksDir + "/test/_default", To: TasksDir + "/test/" + LocalPart},
	{From: TasksDir + "/test/e2e", To: TasksDir + "/test/e2e/" + LocalPart},
	{From: TasksDir + "/test/e2e/_default", To: TasksDir + "/test/e2e/" + LocalPart},
	{From: TasksDir + "/test/golden", To: TasksDir + "/test/golden/" + LocalPart},
	{From: TasksDir + "/test/golden/_default", To: TasksDir + "/test/golden/" + LocalPart},
}

// DispatcherDirs are the folders of the dispatchers that run their NN-
// parts, relative to TasksDir (D15).
var DispatcherDirs = []string{"code/format", "code/lint", "code/sec", "setup/deps", "setup/secrets", "test", "test/e2e", "test/golden"}

// Move is one file that init moves: From to To.
type Move struct{ From, To string }

// Moving holds the absolute paths of the files that init moves, while
// `bootstrap migrate` plans the sync before the move: the dry run then does not
// stop on such a file at the place of a folder.
var Moving map[string]bool

// Listed gives the path of each owned-file row of the manifest text <text>.
func Listed(text []byte) map[string]bool {
	out := map[string]bool{}
	for _, line := range strings.Split(string(text), "\n") {
		if line == "" || line[0] == '#' {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) == 2 && f[0] != "" && f[1] != "" {
			out[f[1]] = true
		}
	}
	return out
}

// ManifestText gives the text of the manifest <file> of the repo <root>, and
// whether <file> exists. It reads <file> only past the link and kind rule of
// manifest.Read: a link, a file below a linked folder or a file that is not
// regular (a FIFO would block the read) gives no text. The sync refuses such
// a manifest later, with its message.
func ManifestText(root, file string) ([]byte, bool) {
	root = strings.TrimSuffix(root, "/")
	info, err := os.Lstat(file)
	if err != nil {
		return nil, false
	}
	rest, below := strings.CutPrefix(file, root+"/")
	if below && manifest.Linked(root, rest) || !below && info.Mode()&os.ModeSymlink != 0 {
		return nil, true
	}
	if info, err := os.Stat(file); err != nil || !info.Mode().IsRegular() {
		return nil, true
	}
	b, _ := os.ReadFile(file)
	return b, true
}

// ListedFile gives Listed of the manifest <file> of the repo <root>
// (ManifestText); a missing or refused file lists nothing.
func ListedFile(root, file string) map[string]bool {
	b, _ := ManifestText(root, file)
	return Listed(b)
}

// OwnTasks gives each path of TaskPorts below <root> (relative) that is a
// regular file, that <listed> does not name and that has no bootstrap header:
// a task of the repo, at the place of a dispatcher. A file with the header is
// bootstrap's own: the sync overwrites it (D26). The engine follows no
// symlink: a path that is a link, or is below a linked folder, and a path
// whose target is, stay (manifest.Linked); the sync then fails the target
// below the link by its own rule.
func OwnTasks(root string, listed map[string]bool) []Move {
	root = strings.TrimSuffix(root, "/")
	var out []Move
	for _, m := range TaskPorts {
		if listed[m.From] {
			continue
		}
		if manifest.Linked(root, m.From) || manifest.Linked(root, m.To) {
			continue
		}
		p := root + "/" + m.From
		if !isFile(p) || header.HasFile(p) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// HeaderNote is the label of the change of init that removes the old header
// copies of the dispatchers (HeaderCopies).
const HeaderNote = "remove old copies of the dispatchers"

// HeaderCopies gives each old copy of a dispatcher in the repo <root>: the
// file <dir>/_default of a folder of DispatcherDirs, when a tool of <tools>
// of the index <idx> ships <dir>/_default.py. mise runs the file without an
// extension, so the old copy hides the new dispatcher. The file is a regular
// file with the bootstrap header, and <listed> does not name it: before the
// dispatcher became a .py file, the sync wrote over such a file (D26). The
// sync removes a listed file by its manifest. A file without the header is
// the repo's own task: OwnTasks moves it. The engine follows no symlink: a
// path that is a link, or is below a linked folder, is not a copy.
func HeaderCopies(root string, listed map[string]bool, idx *source.Index, tools []string) []Copy {
	root = strings.TrimSuffix(root, "/")
	if idx == nil {
		return nil
	}
	ships := map[string]string{}
	for _, tool := range tools {
		m := idx.Manifests[tool]
		if m == nil {
			continue
		}
		for _, fe := range m.Files {
			if fe.Task && ships[fe.Target] == "" {
				ships[fe.Target] = tool
			}
		}
	}
	var out []Copy
	for _, dir := range DispatcherDirs {
		rel := TasksDir + "/" + dir + "/_default"
		tool := ships[rel+".py"]
		if tool == "" || listed[rel] || manifest.Linked(root, rel) {
			continue
		}
		p := root + "/" + rel
		if info, err := os.Lstat(p); err != nil || !info.Mode().IsRegular() || !header.HasFile(p) {
			continue
		}
		out = append(out, Copy{Tool: tool, Task: dir + "/_default.py", Path: p})
	}
	return out
}

// PlanMoves gives the moves of init for the repo <root>, with absolute paths:
// OwnTasks with <listed>. When the target of any of them exists, it moves
// nothing and gives a problem with code 2 for each, which names both paths
// (D8, D20). The same holds for the temp name of a move (TempName) that
// exists: the move would replace it. <show> gives a path in a message.
func PlanMoves(root string, listed map[string]bool, show func(string) string) ([]Move, error) {
	root = strings.TrimSuffix(root, "/")
	var out []Move
	var probs []*fail.Error
	for _, m := range OwnTasks(root, listed) {
		from, to := root+"/"+m.From, root+"/"+m.To
		if _, err := os.Lstat(to); err == nil {
			probs = append(probs, usage("%s is a task of this repo at the place of a dispatcher of bootstrap, and %s exists - join the two tasks in %s, then run bootstrap migrate again", show(from), show(to), show(to)))
			continue
		}
		if _, err := os.Lstat(TempName(from)); err == nil {
			probs = append(probs, usage("%s exists, and bootstrap migrate moves %s through it - move it away, then run bootstrap migrate again", show(TempName(from)), show(from)))
			continue
		}
		out = append(out, Move{From: from, To: to})
	}
	if len(probs) > 0 {
		return nil, source.Problems(probs)
	}
	return out, nil
}

// MoveDetail gives the detail of the move change: each file and its target.
func MoveDetail(moves []Move, show func(string) string) string {
	var parts []string
	for _, m := range moves {
		parts = append(parts, show(m.From)+" → "+show(m.To))
	}
	return strings.Join(parts, ", ")
}

// TaskProblems gives a problem with code 2 for each own task of the repo
// <root> at the place of a dispatcher (OwnTasks with the manifest <manifest>),
// with the fix of init (D4, D7). While init plans (Migrating), it gives none:
// init moves them. <show> gives a path in a message.
func TaskProblems(root, manifest string, show func(string) string) error {
	if Migrating {
		return nil
	}
	root = strings.TrimSuffix(root, "/")
	var probs []*fail.Error
	for _, m := range OwnTasks(root, ListedFile(root, manifest)) {
		probs = append(probs, usage("%s is a task of this repo at the place of a dispatcher of bootstrap - %s", show(root+"/"+m.From), Hint))
	}
	if len(probs) == 0 {
		return nil
	}
	return source.Problems(probs)
}

// partRe is the name of a part that a dispatcher runs, as own_tasks of the
// helpers finds it.
var partRe = regexp.MustCompile(`^[0-9][0-9]-.`)

// dependsRe is a `#MISE depends` line of a task file.
var dependsRe = regexp.MustCompile(`^#\s*\[?MISE\]?\s+depends\s*=(.*)$`)

// quotedRe is a quoted item of a depends line.
var quotedRe = regexp.MustCompile(`"([^"]*)"|'([^']*)'`)

// UnnamedParts gives each executable file in a dispatcher folder of the repo
// <root> (DispatcherDirs) without an NN- name, relative to <root>, sorted: a
// dispatcher does not run it (D15). A name that starts with `_` is not a
// part, and neither is a file whose task name the `#MISE depends` of another
// file in the same folder names (D18). The task name of a `.py` file is its
// file name without `.py`, as mise gives it. The output keeps the file name.
func UnnamedParts(root string) []string {
	root = strings.TrimSuffix(root, "/")
	var out []string
	for _, d := range DispatcherDirs {
		dir := root + "/" + TasksDir + "/" + d
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		var cands []string
		named := map[string]bool{}
		for _, e := range entries {
			p := filepath.Join(dir, e.Name())
			info, err := os.Stat(p)
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			for _, n := range dependsOf(p) {
				named[n] = true
			}
			if strings.HasPrefix(e.Name(), "_") || partRe.MatchString(e.Name()) || info.Mode().Perm()&0o111 == 0 {
				continue
			}
			cands = append(cands, e.Name())
		}
		task := strings.ReplaceAll(d, "/", ":")
		for _, n := range cands {
			if !named[task+":"+strings.TrimSuffix(n, ".py")] {
				out = append(out, TasksDir+"/"+d+"/"+n)
			}
		}
	}
	sort.Strings(out)
	return out
}

// dependsOf gives the task names of the `#MISE depends` lines of the file
// <p>: the first word of each quoted item.
func dependsOf(p string) []string {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		m := dependsRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		for _, q := range quotedRe.FindAllStringSubmatch(m[1], -1) {
			item := q[1] + q[2]
			if f := strings.Fields(item); len(f) > 0 {
				out = append(out, f[0])
			}
		}
	}
	return out
}

// ApplyMoves moves each file of <moves>: From to To, through a temp name next
// to From, so a file can become the part of a folder at its own path. A
// failure moves the files of <moves> back that it moved.
func ApplyMoves(moves []Move) error {
	for i, m := range moves {
		if err := move(m.From, m.To); err != nil {
			_ = UndoMoves(moves[:i])
			return fail.New(fail.Other, "cannot move %s to %s", m.From, m.To)
		}
	}
	return nil
}

// UndoMoves moves each file of <moves> back, in the reverse order: To to
// From. A From that is a folder now (a file that became the folder of its
// part) goes first: before the move it was a file, so the folder holds only
// what the move and the sync put there.
func UndoMoves(moves []Move) error {
	for i := len(moves) - 1; i >= 0; i-- {
		m := moves[i]
		if _, err := os.Lstat(m.To); err != nil {
			continue
		}
		tmp := TempName(m.From)
		if err := renameFree(m.To, tmp); err != nil {
			return fail.New(fail.Other, "cannot move %s back to %s", m.To, m.From)
		}
		if info, err := os.Lstat(m.From); err == nil && info.IsDir() {
			if err := os.RemoveAll(m.From); err != nil {
				return fail.New(fail.Other, "cannot move %s back to %s", m.To, m.From)
			}
		}
		if err := os.Rename(tmp, m.From); err != nil {
			return fail.New(fail.Other, "cannot move %s back to %s", m.To, m.From)
		}
	}
	return nil
}

// TempName is the temp name of a move of the file <from>, next to it.
func TempName(from string) string { return from + ".bootstrap-move" }

// renameFree renames <from> to <to> only when <to> does not exist, so a
// rename never replaces a file.
func renameFree(from, to string) error {
	if _, err := os.Lstat(to); err == nil {
		return os.ErrExist
	}
	return os.Rename(from, to)
}

// move moves the file <from> to <to>, through TempName.
func move(from, to string) error {
	tmp := TempName(from)
	if err := renameFree(from, tmp); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o777); err != nil {
		_ = os.Rename(tmp, from)
		return err
	}
	if err := os.Rename(tmp, to); err != nil {
		_ = os.Rename(tmp, from)
		return err
	}
	return nil
}
