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
// with the members of FAKE_MEMBERS.
const fakeEngine = `#!/usr/bin/env bash
if [[ ${1:-} == config && ${2:-} == get && ${3:-} == repo.members ]]; then
  printf '{"key":"repo.members","value":%s,"source":"file"}\n' "${FAKE_MEMBERS:-null}"
  exit 0
fi
[[ ${1:-} == sync && $(pwd -P) == "$(cd -- "$MISE_PROJECT_ROOT" && pwd -P)" ]] || exit 1
[[ -z ${BOOTSTRAP_MEMBER:-} || ${BOOTSTRAP_HOME:-} == "$SYNC_ROOT" ]] || exit 1
sub=${MISE_PROJECT_ROOT#"$SYNC_ROOT"}
sub=${sub#/}
printf '%s\n' "sync${sub:+ $sub}${2:+ $2}${BOOTSTRAP_MEMBER:+ member=$BOOTSTRAP_MEMBER}" >> "$SYNC_OUT"
[[ -z ${SYNC_FAIL:-} || ${SYNC_FAIL_IN:-$sub} != "$sub" ]] || exit "$SYNC_FAIL"
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
// with its own values folder. <path> runs the fake engine as `bootstrap` of
// PATH, without BOOTSTRAP_BIN. <change> changes the root before the run.
// <kv> are more variables, for example usage_all=true.
func (e configsEnv) run(t *testing.T, path bool, change func(root string), kv ...string) configsRun {
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
	out := filepath.Join(dir, "synced")
	env := harness.Env()
	if path {
		env = harness.SetEnv(env, "BOOTSTRAP_BIN")
	} else {
		env = harness.SetEnv(env, "BOOTSTRAP_BIN="+filepath.Join(e.binDir, "bootstrap"))
	}
	env = harness.SetEnv(env, "FAKE_MEMBERS="+defaultMembers, "MISE_PROJECT_ROOT="+root,
		"PATH="+e.binDir+":"+harness.GetEnv(env, "PATH"),
		"SYNC_OUT="+out, "SYNC_ROOT="+root)
	env = harness.SetEnv(env, kv...)
	cmd := &harness.Cmd{Env: env, Args: pyTask(t, e.task)}
	code, log := cmd.Run()
	return configsRun{code: code, log: log, synced: strings.TrimRight(harness.ReadFile(out), "\n")}
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
	if err := os.Chmod(filepath.Join(e.binDir, "bootstrap"), 0o755); err != nil {
		t.Fatal(err)
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
}
