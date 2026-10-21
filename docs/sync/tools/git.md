# The `git` tool

The shared `.gitignore`, the commit convention, the git-config check and the
merge tasks.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not list it in the values
file.

## What the tool ships

| Path in the repository                  | Kind     | Use                                                                                        |
| --------------------------------------- | -------- | ------------------------------------------------------------------------------------------ |
| `.gitignore`                            | Synced   | The shared ignore lines, then the lines of `.config/repo-sync/gitignore`.                  |
| `.config/git-conventional-commits.yaml` | Template | The commit types, and the scopes from `commit_scopes`, else the members of the repository. |
| `.config/mise/tasks/_scripts/merge`     | Synced   | The library `merge`, with `merge_to_destination_branch`.                                   |
| `.config/mise/tasks/code/git-config`    | Synced   | The task `code:git-config`.                                                                |
| `.config/mise/tasks/code/merge/`        | Synced   | The merge tasks of the branch model (see "Tasks").                                         |

## mise tools

None.

## Tasks

| Task                          | Use                                                                |
| ----------------------------- | ------------------------------------------------------------------ |
| `code:git-config`             | Find forbidden keys in the local git config. `--fix` removes them. |
| `code:merge <branch>`         | Land a branch on the only long-lived branch.                       |
| `code:merge:develop <branch>` | Land a branch on `develop`.                                        |
| `code:merge:main`             | Land `develop` on `main`.                                          |

The sync writes the merge tasks of the branch model that it finds (step 4 of
`index.md`), and removes the merge tasks of the other model.
`MERGE_MODEL_DEVELOP` and `MERGE_MODEL_MAIN` select `direct` or `pr`. The
`pr` model needs the CLI of the forge, from the `github` or the `gitlab` tool.

## Pre-commit hooks

- The local hook `git-config`, which calls `code:git-config --fix`.
- `no-commit-to-branch` of `pre-commit/pre-commit-hooks`, for `main` and
  `develop`.
- `qoomon/git-conventional-commits` in the commit-msg stage.

## Files the repository supplies

None.

## Adapt to the repository

1. Select the branch model (step 4 of `index.md`).
2. Set `commit_scopes` in the values file.
3. Move each repository ignore line to `.config/repo-sync/gitignore`.

## Migration

- The general items M10 and M11 of `migrate.md` are about the files of this
  tool.
