# U3 — Test the dev rule and the zz-repo rule

- **Wave:** 1
- **Depends on:** —
- **Owns:** `.config/mise/tasks/test/repo-sync`
- **Model:** opus
- **Kind:** edit
- **Read first:** the owned file, top to bottom (`run_case` at line 46).

## Ruling

Quoted from `index.md`:

- D3: "A tool folder whose name sorts after `zz-repo` stops the sync with
  code 2."
- D4: "Each `env` of the values file and each `default_envs` without `dev`
  stops the sync with code 2."
- D7: "Before its first `git` call, `test:repo-sync` unsets each `GIT_*`
  variable that a git hook exports, so its scratch repositories never write
  to the real repository."

## Edits

1. Change the case "env ci on a tool whose mise tool allows dev only"
   (line 142) to `env = ["dev", "ci"]`, so it still tests the `envs` of the
   mise tool and not the new `dev` rule.
2. Add three cases, each with code 2:
   - `env = ["ci"]` on a `[tools.*]` entry (no `dev`);
   - a `tools.toml` in the copy with `default_envs = ["ci"]`;
   - a tool folder in the copy named `zzz-late`, which sorts after `zz-repo`.
3. Keep the message checks of `run_case` for each new case: match the words
   that U1's messages use (`dev`, `zz-repo`).
4. Before the first `git` call of the task, unset the `GIT_*` variables that
   a git hook exports: `GIT_DIR`, `GIT_INDEX_FILE`, `GIT_WORK_TREE`,
   `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`,
   `GIT_COMMON_DIR`, `GIT_PREFIX` and `GIT_EXEC_PATH`. Explain why in a short
   comment: the `test-repo-sync` pre-commit hook runs the task inside
   `git commit`, and with these variables set, the `git init`, `git add` and
   `git commit` of the scratch repositories write to the real repository
   (gap G2).

## Verification

- `mise x -- mise run test:repo-sync` passes after wave 1, with U1 and U2.
- `bash -n` and `mise x -- shellcheck -x -s bash` pass.
- `mise x -- mise run code:precommit` passes.
- Run the task the way a hook does, from the worktree root, with
  `GIT_DIR="$(git rev-parse --git-dir)"` and
  `GIT_INDEX_FILE="$(git rev-parse --git-dir)/index"` exported. The task
  passes, and these do not change: `git rev-parse HEAD`,
  `git status --porcelain`, and `git config --get core.bare` (it prints
  `false`).

## Guardrails

- Touch only the owned file. The task needs no network.

## Commit

`test: cover the dev rule and the zz-repo rule of repo-sync`
