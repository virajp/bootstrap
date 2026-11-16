package oldform

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/manifest"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// The old command of the worktree setup, and its new name (D20). The sync
// never rewrites the seed WtFile, and a member's worktree_install is a value
// of the repo, so init renames the command in both.
const (
	OldWorktreeCmd = "mise run setup:worktree"
	NewWorktreeCmd = "mise run git:worktree:setup"
	// RenameNote is the label of the change of init that renames it.
	RenameNote = "rename setup:worktree"
	// WtFile is the worktree config of a repo, relative to its root.
	WtFile = ".config/wt.toml"
)

// oldCmdRe is OldWorktreeCmd as a whole command: a character that can go on
// a task name does not follow it.
var oldCmdRe = regexp.MustCompile(regexp.QuoteMeta(OldWorktreeCmd) + `([^A-Za-z0-9_:-]|$)`)

// HasOldCmd tells if the text <s> runs OldWorktreeCmd.
func HasOldCmd(s string) bool { return oldCmdRe.MatchString(s) }

// RenameCmd gives <s> with NewWorktreeCmd in place of each OldWorktreeCmd;
// the rest of <s> stays as it is.
func RenameCmd(s string) string { return oldCmdRe.ReplaceAllString(s, NewWorktreeCmd+"${1}") }

// WtText gives the text of WtFile of the repo <root>, when it is a regular
// file that runs OldWorktreeCmd. The engine follows no symlink: a link, or a
// file below a linked folder, gives nothing.
func WtText(root string) ([]byte, bool) {
	root = strings.TrimSuffix(root, "/")
	if manifest.Linked(root, WtFile) || !isFile(root+"/"+WtFile) {
		return nil, false
	}
	b, err := os.ReadFile(root + "/" + WtFile)
	if err != nil || !HasOldCmd(string(b)) {
		return nil, false
	}
	return b, true
}

// OldMembers gives the name of each member of the table <table> of the keys
// <raw> ("members" of repo.toml, or "member" of the old values file) whose
// worktree_install runs OldWorktreeCmd, sorted, and the value of each.
func OldMembers(raw *source.Obj, table string) ([]string, map[string]string) {
	if raw == nil {
		return nil, nil
	}
	ms, ok := raw.Get(table).(*source.Obj)
	if !ok {
		return nil, nil
	}
	var names []string
	vals := map[string]string{}
	for _, n := range ms.Keys() {
		m, ok := ms.Get(n).(*source.Obj)
		if !ok {
			continue
		}
		if v, ok := m.Get("worktree_install").(string); ok && HasOldCmd(v) {
			names = append(names, n)
			vals[n] = v
		}
	}
	sort.Strings(names)
	return names, vals
}

// RenameProblems gives a problem with code 2, with the fix of init, for
// WtFile of the repo <root> and for each member of the values <raw> (the
// keys of repo.toml, shown as <shownRepo>) that runs OldWorktreeCmd (D20).
// While init plans (Migrating), it gives none: init renames them. <show>
// gives a path in a message.
func RenameProblems(root string, raw *source.Obj, shownRepo string, show func(string) string) error {
	if Migrating {
		return nil
	}
	root = strings.TrimSuffix(root, "/")
	var probs []*fail.Error
	if _, ok := WtText(root); ok {
		probs = append(probs, usage("%s runs %s, the old name of git:worktree:setup - %s", show(root+"/"+WtFile), OldWorktreeCmd, Hint))
	}
	names, _ := OldMembers(raw, "members")
	for _, n := range names {
		probs = append(probs, usage("members.%s.worktree_install in %s runs %s, the old name of git:worktree:setup - %s", n, shownRepo, OldWorktreeCmd, Hint))
	}
	if len(probs) == 0 {
		return nil
	}
	return source.Problems(probs)
}
