# The `git` tool

The shared `.gitignore` and `.gitattributes`, the commit convention, the
git-config check and the merge tasks.

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
| `.gitignore`                            | Synced   | The shared ignore lines, with `[overrides.gitignore]` of the values file applied.          |
| `.gitattributes`                        | Synced   | LF line endings, the generated lockfiles, and the binary files of each stack.              |
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
The merge tasks read `[branch.<destination>]` of the
repository's own `.config/bootstrap.toml`, with `mise x -- yq`. Its
`merge_model` selects `direct` (the default) or `pr`. Its `merge_from` or
`merge_never` limits the source branches. Without both lists, only `develop`
merges into `main` when `develop` exists, and nothing merges from `main`. A
table with both lists stops the merge. The `pr` model needs the CLI of the
forge, from the `github` or the `gitlab` tool. The merge tasks do not read
`direct_commit`: `repo-sync` reads it. Before a merge, the merge tasks run the
hooks over all files as a safety net, with `SKIP=no-commit-to-branch`. The
safety net makes no commit, so that hook does not apply to it.

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
