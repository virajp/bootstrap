# The `git` tool

The shared `.gitignore` and `.gitattributes`, the commit convention, the
git-config check, the merge tasks and worktrunk.

| Property     | Value             |
| ------------ | ----------------- |
| Root only    | No                |
| Depends      | None              |
| Default envs | dev               |
| Seeds        | `.config/wt.toml` |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not select it in the values
file. A `[tools.git]` table gives values.

## What the tool ships

| Path in the repository                  | Kind     | Use                                                                                             |
| --------------------------------------- | -------- | ----------------------------------------------------------------------------------------------- |
| `.gitignore`                            | Synced   | The shared ignore lines, with `[overrides.gitignore]` of the values file applied.               |
| `.gitattributes`                        | Synced   | LF line endings, the generated lockfiles, and the binary files of each stack.                   |
| `.config/git-conventional-commits.yaml` | Template | The commit types, and the scopes from `commit_scopes`, else the members of the repository.      |
| `.config/mise/tasks/_scripts/merge`     | Synced   | The library `merge`, with `merge_to_destination_branch`.                                        |
| `.config/mise/tasks/code/git-config`    | Synced   | The task `code:git-config`.                                                                     |
| `.config/mise/tasks/code/merge/`        | Synced   | The merge tasks of the branch model (see "Tasks"). The `mise` tool ships `code/merge/_default`. |
| `.config/wt.toml`                       | Seed     | The worktrunk config: the `pre-start` hooks and the commit convention (see "worktrunk").        |

## mise tools

| mise tool   | Version  | Environments | Note                           |
| ----------- | -------- | ------------ | ------------------------------ |
| `worktrunk` | `latest` | dev          | The git worktree manager `wt`. |

The tool is in `dev` only. Its files and hooks do not depend on an
environment, so a CI job still gets the merge tasks and the hooks.

## worktrunk

The seed `.config/wt.toml` is the project config of worktrunk. A new worktree
of `wt switch --create` runs its `pre-start` steps in sequence:

1. `mise trust`.
2. `wt step copy-ignored --require-include`. It copies the gitignored files
   of the main worktree that `.worktreeinclude` lists, for example build
   caches. Without `.worktreeinclude`, it copies nothing.
3. `mise run setup:worktree`.

Its `template-append` tells the LLM commit messages of `wt` to use the
commit types of `.config/git-conventional-commits.yaml`. wt asks the user to
approve each project command the first time that it runs.

Land a branch with the merge tasks, not with `wt merge`. By default,
`wt merge` squashes, rebases and fast-forwards, and the project config cannot
change that. The user config `~/.config/worktrunk/config.toml` and the shell
integration (`wt config shell install`) belong to the user, and the sync does
not write them.

## Tasks

| Task                          | Use                                                                |
| ----------------------------- | ------------------------------------------------------------------ |
| `code:git-config`             | Find forbidden keys in the local git config. `--fix` removes them. |
| `code:merge <branch>`         | Land a branch on the only long-lived branch.                       |
| `code:merge:develop <branch>` | Land a branch on `develop`.                                        |
| `code:merge:main`             | Land `develop` on `main`.                                          |

The sync writes the merge tasks of the branch model that it finds (step 4 of
`index.md`), and removes the merge tasks of the other model.
The merge tasks read `[branch.<destination>]` of the
repository's own `.config/bootstrap.toml`, with `yq` (a tools directive of each
merge task). Its
`merge_model` selects `direct` (the default) or `pr`. Its `merge_from` or
`merge_never` limits the source branches. Without both lists, only `develop`
merges into `main` when `develop` exists, and nothing merges from `main`. A
table with both lists stops the merge. The `pr` model needs the CLI of the
forge, from the `github` or the `gitlab` tool. In a repository without an
`origin` remote, `direct` merges locally without the pull and the push, and
`pr` stops before the hooks run. The merge tasks do not read
`direct_commit`: `repo-sync` reads it. Before a merge, the merge tasks run the
hooks over all files as a safety net, with `SKIP=no-commit-to-branch`. The
safety net makes no commit, so that hook does not apply to it.

A merge task stops when the repository of `MISE_ORIGINAL_CWD` is not the
repository of the task. In a submodule without that merge task, mise finds the
task of the root repository and runs it in the root folder. Without the stop,
that task lands a branch of the root. Run `mise run setup:configs` in the
submodule to get its own merge tasks.

## Pre-commit hooks

- The local hook `git-config`, which calls `code:git-config --fix`.
- `no-commit-to-branch` of `pre-commit/pre-commit-hooks`, with the name
  "Refuse a direct commit to a protected branch". Its args are
  `--branch main --branch develop`. The sync removes the `--branch` pair of a
  core branch that does not exist, and of a branch with `direct_commit = true`
  in its `[branch.<branch>]` table. When no branch is left, the sync removes
  the hook. See step 4 of `index.md`.
- `qoomon/git-conventional-commits` in the commit-msg stage.

## Files the repository supplies

None.

## Adapt to the repository

1. Select the branch model (step 4 of `index.md`).
2. Set `commit_scopes` in the values file.
3. Put each repository ignore line in `add` of `[overrides.gitignore]` in the
   values file. Put each shared line that the repository must not ignore in
   `remove`.
4. If the repository has its own lines in `.gitattributes`, tell the user. The
   sync replaces the file, and it has no overrides.

## Migration

- The general items M10, M11 and M18 of `migrate.md` are about the files of
  this tool.
