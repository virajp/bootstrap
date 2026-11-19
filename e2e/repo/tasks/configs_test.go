package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

// The fake engine of setup:configs. It records each call: the folder below
// the root, the flag, and the member name of BOOTSTRAP_MEMBER. It fails when
// it does not run in the folder of MISE_PROJECT_ROOT, or when a member sync
// has no BOOTSTRAP_HOME of the root. It answers `config get repo.members`
// with the members of FAKE_MEMBERS, `config get repo.core_branches` with
// main, the landing of main and its table with FAKE_LANDING, and each other
// key with null.
// With SYNC_WRITE, a sync writes the file synced.txt.
const fakeEngine = `#!/usr/bin/env bash
if [[ ${1:-} == config && ${2:-} == get ]]; then
  case ${3:-} in
    repo.members) value=${FAKE_MEMBERS:-null} ;;
    repo.core_branches) value='["main"]' ;;
    'repo.branches."main".landing') value=${FAKE_LANDING:-null} ;;
    'repo.branches."main"') value=null; [[ -z ${FAKE_LANDING:-} ]] || value="{\"landing\":$FAKE_LANDING}" ;;
    *) value=null ;;
  esac
  printf '{"key":"k","value":%s,"source":"file"}\n' "$value"
  exit 0
fi
[[ ${1:-} == sync && $(pwd -P) == "$(cd -- "$MISE_PROJECT_ROOT" && pwd -P)" ]] || exit 1
[[ -z ${BOOTSTRAP_MEMBER:-} || ${BOOTSTRAP_HOME:-} == "$SYNC_ROOT" ]] || exit 1
sub=${MISE_PROJECT_ROOT#"$SYNC_ROOT"}
sub=${sub#/}
printf '%s\n' "sync${sub:+ $sub}${2:+ $2}${BOOTSTRAP_MEMBER:+ member=$BOOTSTRAP_MEMBER}" >> "$SYNC_OUT"
[[ -z ${SYNC_WRITE:-} ]] || printf 'synced\n' > synced.txt
[[ -z ${SYNC_FAIL:-} || ${SYNC_FAIL_IN:-$sub} != "$sub" ]] || exit "$SYNC_FAIL"
`

// fakeMise runs `mise run --cd <folder> git:merge:<branch> <source>` as the
// task file of the folder, with the PATH of LAND_PATH, which has no gh, glab
// and pre-commit: so the landing makes no request and runs no hook. It passes
// each other call to the mise shim of the harness.
const fakeMise = `#!/usr/bin/env bash
if [[ ${1:-} == run && ${2:-} == --cd && ${4:-} == git:merge:* ]]; then
  cd "$3" || exit 1
  export MISE_PROJECT_ROOT="$3" usage_source="$5" PATH="$LAND_PATH"
  unset MISE_ORIGINAL_CWD
  exec "$3/.config/mise/tasks/git/merge/${4#git:merge:}.py"
fi
exec "$HARNESS_MISE" "$@"
`

const defaultMembers = `{"sub":{"folder":"sub","type":"submodule"}}`

// configsRun is one run of setup:configs with the fake engine.
type configsRun struct {
	code   int
	log    string
	synced string
}

// configsEnv holds the rendered task and the fake engine of the test.
type configsEnv struct {
	task, binDir string
}

// run runs the task in a new scratch root with one member `sub`, a folder
// with its own values folder. The root and sub are git repositories on main,
// each with one commit; the root ignores sub/. <path> runs the fake engine as
// `bootstrap` of PATH, without BOOTSTRAP_BIN. <change> changes the root
// before the commits. <kv> are more variables, for example usage_all=true.
func (e configsEnv) run(t *testing.T, path bool, change func(root string), kv ...string) configsRun {
	return e.runAfter(t, path, change, nil, kv...)
}

// runAfter is run with <after>, a change of the root after the commits.
func (e configsEnv) runAfter(t *testing.T, path bool, change, after func(root string), kv ...string) configsRun {
	t.Helper()
	dir := harness.Dir(t)
	root := filepath.Join(dir, "members")
	copyFile(t, helpersPath(), filepath.Join(root, ".config", "mise", "tasks", "_scripts", "helpers.py"))
	if err := os.MkdirAll(filepath.Join(root, "sub", ".config", "bootstrap"), 0o755); err != nil {
		t.Fatal(err)
	}
	if change != nil {
		change(root)
	}
	harness.Git(t, root, "init", "--quiet", "-b", "main")
	harness.AppendFile(t, filepath.Join(root, ".git", "info", "exclude"), "/sub/\n")
	harness.Git(t, root, "add", "--all")
	harness.Git(t, root, "commit", "--quiet", "-m", "root")
	harness.Git(t, filepath.Join(root, "sub"), "init", "--quiet", "-b", "main")
	harness.Git(t, filepath.Join(root, "sub"), "commit", "--quiet", "--allow-empty", "-m", "sub")
	if after != nil {
		after(root)
	}
	out := filepath.Join(dir, "synced")
	env := harness.Env()
	if path {
		env = harness.SetEnv(env, "BOOTSTRAP_BIN")
	} else {
		env = harness.SetEnv(env, "BOOTSTRAP_BIN="+filepath.Join(e.binDir, "bootstrap"))
	}
	// git of the task runs without the config of the user, so a commit of the
	// task needs no key and no signing.
	env = harness.SetEnv(env, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	env = harness.SetEnv(env, "HARNESS_MISE="+filepath.Join(harness.ShimDir, "mise"), "LAND_PATH="+landPath(env))
	env = harness.SetEnv(env, "FAKE_MEMBERS="+defaultMembers, "MISE_PROJECT_ROOT="+root,
		"PATH="+e.binDir+":"+harness.GetEnv(env, "PATH"),
		"SYNC_OUT="+out, "SYNC_ROOT="+root)
	env = harness.SetEnv(env, kv...)
	cmd := &harness.Cmd{Env: env, Args: pyTask(t, e.task)}
	code, log := cmd.Run()
	return configsRun{code: code, log: log, synced: strings.TrimRight(harness.ReadFile(out), "\n")}
}

// landPath is the PATH of <env> without each folder that has gh, glab or
// pre-commit.
func landPath(env []string) string {
	var keep []string
	for _, dir := range filepath.SplitList(harness.GetEnv(env, "PATH")) {
		drop := false
		for _, bin := range []string{"gh", "glab", "pre-commit"} {
			if _, err := os.Stat(filepath.Join(dir, bin)); err == nil {
				drop = true
			}
		}
		if !drop {
			keep = append(keep, dir)
		}
	}
	return strings.Join(keep, string(os.PathListSeparator))
}

// check is check_configs: the exit code, and the calls of the fake engine.
func (r configsRun) check(t *testing.T, name string, want int, synced string) {
	t.Helper()
	switch {
	case r.code != want:
		harness.Fail(t, fmt.Sprintf("%s: exit %d, expected %d", name, r.code, want), r.log)
	case r.synced != synced:
		harness.Fail(t, fmt.Sprintf("%s: synced '%s', expected '%s'", name, r.synced, synced), r.log)
	default:
		harness.Pass(t, fmt.Sprintf("%s: exit %d", name, r.code))
	}
}

// checkLog is check_configs_log: the output has (<has>) or lacks the fixed
// string <s>.
func checkLog(t *testing.T, name, log string, has bool, s string) {
	t.Helper()
	flag := "-l"
	if has {
		flag = "-h"
	}
	if strings.Contains(log, s) == has {
		harness.Pass(t, name+": the output")
	} else {
		harness.Fail(t, fmt.Sprintf("%s: the check %s of the output fails for: %s", name, flag, s), log)
	}
}

// TestSetupConfigs: setup:configs runs `bootstrap sync` for the root, then
// for each selected member. The task is rendered with one submodule `sub`
// that has a [members.sub] table.
func TestSetupConfigs(t *testing.T) {
	dir := harness.Dir(t)
	ctx := filepath.Join(dir, "ctx.json")
	harness.WriteFile(t, ctx, fmt.Sprintf(`{"members":["sub"],"member_tables":["sub"],"member_types":{"sub":"submodule"},"member_folders":{"sub":"sub"},"cli_dir":"%s/cli"}`+"\n", harness.Root))
	e := configsEnv{task: filepath.Join(dir, "setup-configs.py"), binDir: filepath.Join(dir, "bin")}
	gomplate(t, filepath.Join(harness.Root, "repo", "tools", "mise", "tasks", "setup-configs.py.tmpl"), e.task, ctx)
	harness.WriteFile(t, filepath.Join(e.binDir, "bootstrap"), fakeEngine)
	harness.WriteFile(t, filepath.Join(e.binDir, "mise"), fakeMise)
	for _, bin := range []string{"bootstrap", "mise"} {
		if err := os.Chmod(filepath.Join(e.binDir, bin), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	both := "sync\nsync sub member=sub"

	cases := []struct {
		name   string
		path   bool
		kv     []string
		code   int
		synced string
	}{
		{"setup:configs without a member flag", false, nil, 0, "sync"},
		{"setup:configs runs bootstrap of PATH", true, nil, 0, "sync"},
		{"setup:configs --all", false, []string{"usage_all=true"}, 0, both},
		{"setup:configs --all, the root fails", false, []string{"SYNC_FAIL=2", "usage_all=true"}, 2, "sync"},
		{"setup:configs --all, the member fails", false, []string{"SYNC_FAIL=2", "SYNC_FAIL_IN=sub", "usage_all=true"}, 2, both},
		{"setup:configs --dry-run --sub", false, []string{"usage_dry_run=true", "usage_sub=true"}, 0, "sync --dry-run\nsync sub --dry-run member=sub"},
		// A folder member is part of the root tree: members leaves it out.
		{"setup:configs --all, a folder member", false, []string{"usage_all=true",
			`FAKE_MEMBERS={"sub":{"folder":"sub","type":"submodule"},"web":{"folder":"web","type":"folder"}}`}, 0, both},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e.run(t, tc.path, nil, tc.kv...).check(t, tc.name, tc.code, tc.synced)
		})
	}

	t.Run("setup:configs --sub", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs --sub"
		r := e.run(t, false, nil, "usage_sub=true")
		r.check(t, name, 0, both)
		if !strings.Contains(r.log, "Syncing sub ...") {
			harness.Fail(t, name+": no header for sub", r.log)
		}
	})

	// A member of the old form, with no .config/bootstrap/, is skipped with a
	// warning, and the task goes on.
	t.Run("setup:configs --all, a member of the old form", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs --all, a member of the old form"
		r := e.run(t, false, func(root string) {
			if err := os.Rename(filepath.Join(root, "sub", ".config", "bootstrap"), filepath.Join(root, "sub", ".config", "bootstrap.off")); err != nil {
				t.Fatal(err)
			}
		}, "usage_all=true")
		r.check(t, name, 0, "sync")
		checkLog(t, name, r.log, true, "member sub: run bootstrap migrate there")
	})

	// A member that is not on this machine is skipped with a warning.
	t.Run("setup:configs --all, a sibling that is not there", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs --all, a sibling that is not there"
		r := e.run(t, false, nil, "usage_all=true",
			`FAKE_MEMBERS={"sub":{"folder":"sub","type":"submodule"},"gone":{"folder":"../gone","type":"sibling","url":"u"}}`)
		r.check(t, name, 0, both)
		checkLog(t, name, r.log, true, "member gone: ../gone is not on this machine")
	})

	// The branch and land flow. A file task git/merge/main.py makes the flow
	// use ops/bootstrap-sync; with no SYNC_WRITE, the fake sync changes
	// nothing, so the flow never runs the merge task.
	mergeTask := func(root string) {
		harness.WriteFile(t, filepath.Join(root, ".config", "mise", "tasks", "git", "merge", "main.py"), "# stub\n")
	}
	gitOut := func(dir string, args ...string) string { out, _ := harness.GitErr(dir, args...); return out }

	t.Run("setup:configs, uncommitted files stop it", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs, uncommitted files"
		r := e.runAfter(t, false, nil, func(root string) {
			harness.WriteFile(t, filepath.Join(root, "new.txt"), "x\n")
		})
		r.check(t, name, 1, "")
		checkLog(t, name, r.log, true, "this repository has uncommitted or untracked files.")
	})
	t.Run("setup:configs --all, a dirty member stops it before the root", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs --all, a dirty member"
		r := e.runAfter(t, false, nil, func(root string) {
			harness.WriteFile(t, filepath.Join(root, "sub", "new.txt"), "x\n")
		}, "usage_all=true")
		r.check(t, name, 1, "")
		checkLog(t, name, r.log, true, "member sub has uncommitted or untracked files.")
	})
	t.Run("setup:configs, a sync with no change", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs, no change"
		var root string
		r := e.runAfter(t, false, mergeTask, func(r string) { root = r })
		r.check(t, name, 0, "sync")
		checkLog(t, name, r.log, true, "Syncing on ops/bootstrap-sync, from main")
		checkLog(t, name, r.log, true, "The sync changed nothing, so there is nothing to land")
		harness.Check(t, name+": no ops/bootstrap-sync stays", gitOut(root, "branch", "--list", "ops/bootstrap-sync") == "", r.log)
		harness.Check(t, name+": back on main", gitOut(root, "branch", "--show-current") == "main", r.log)
	})
	t.Run("setup:configs, ops/bootstrap-sync exists", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs, the branch exists"
		r := e.runAfter(t, false, mergeTask, func(root string) {
			harness.Git(t, root, "branch", "ops/bootstrap-sync")
		})
		r.check(t, name, 1, "")
		checkLog(t, name, r.log, true, "the branch ops/bootstrap-sync exists, here or on origin.")
	})
	t.Run("setup:configs, no merge task: a direct commit", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs, a direct commit"
		var root string
		r := e.runAfter(t, false, nil, func(r string) {
			root = r
			harness.Git(t, r, "switch", "--quiet", "-c", "feature")
		}, "SYNC_WRITE=1")
		r.check(t, name, 0, "sync")
		checkLog(t, name, r.log, true, "No task git:merge:main, so the sync is committed directly on main")
		harness.Check(t, name+": the commit is on main", gitOut(root, "log", "-1", "--format=%s", "main") == "ops: sync bootstrap", r.log)
		harness.Check(t, name+": back on feature", gitOut(root, "branch", "--show-current") == "feature", r.log)
		harness.Check(t, name+": a clean tree", gitOut(root, "status", "--porcelain") == "", r.log)
	})
	t.Run("setup:configs --dry-run makes no branch", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs --dry-run"
		r := e.run(t, false, mergeTask, "usage_dry_run=true")
		r.check(t, name, 0, "sync --dry-run")
		checkLog(t, name, r.log, false, "ops/bootstrap-sync")
	})

	// The landing with the real merge task of the git tool: the root gets the
	// task git:merge:main, merge.py and checks.py, and the fake sync writes
	// synced.txt.
	realMerge := func(root string) {
		scripts := filepath.Join(root, ".config", "mise", "tasks", "_scripts")
		gitTasks := filepath.Join(harness.Root, "repo", "tools", "git", "tasks")
		copyFile(t, filepath.Join(gitTasks, "git-merge-main.py"), filepath.Join(root, ".config", "mise", "tasks", "git", "merge", "main.py"))
		copyFile(t, filepath.Join(gitTasks, "merge.py"), filepath.Join(scripts, "merge.py"))
		copyFile(t, miseTask("checks.py"), filepath.Join(scripts, "checks.py"))
	}
	t.Run("setup:configs lands the sync with git:merge:main", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs, a direct landing"
		var root string
		r := e.runAfter(t, false, realMerge, func(r string) { root = r }, "SYNC_WRITE=1")
		r.check(t, name, 0, "sync")
		checkLog(t, name, r.log, true, "Landing ops/bootstrap-sync on main (direct-merge) ...")
		harness.Check(t, name+": main has the merge", gitOut(root, "log", "-1", "--format=%s", "main") == "Merge branch 'ops/bootstrap-sync'", r.log)
		harness.Check(t, name+": main has the sync", gitOut(root, "show", "main:synced.txt") == "synced", r.log)
		harness.Check(t, name+": no ops/bootstrap-sync stays", gitOut(root, "branch", "--list", "ops/bootstrap-sync") == "", r.log)
		harness.Check(t, name+": back on main", gitOut(root, "branch", "--show-current") == "main", r.log)
	})
	t.Run("setup:configs opens a pull request under landing pr", func(t *testing.T) {
		t.Parallel()
		name := "setup:configs, landing pr"
		var root, origin string
		r := e.runAfter(t, false, realMerge, func(r string) {
			root = r
			origin = filepath.Join(filepath.Dir(r), "origin.git")
			harness.Git(t, r, "init", "--quiet", "--bare", "-b", "main", origin)
			harness.Git(t, r, "remote", "add", "origin", origin)
			harness.Git(t, r, "push", "--quiet", "-u", "origin", "main")
		}, "SYNC_WRITE=1", `FAKE_LANDING="pr"`)
		r.check(t, name, 0, "sync")
		checkLog(t, name, r.log, true, "Landing ops/bootstrap-sync on main (pr) ...")
		harness.Check(t, name+": origin has the branch", gitOut(origin, "rev-parse", "--verify", "--quiet", "refs/heads/ops/bootstrap-sync") != "", r.log)
		harness.Check(t, name+": main does not change", gitOut(root, "rev-parse", "main") == gitOut(origin, "rev-parse", "main"), r.log)
		harness.Check(t, name+": no local ops/bootstrap-sync", gitOut(root, "branch", "--list", "ops/bootstrap-sync") == "", r.log)
		harness.Check(t, name+": back on main", gitOut(root, "branch", "--show-current") == "main", r.log)
	})
}
