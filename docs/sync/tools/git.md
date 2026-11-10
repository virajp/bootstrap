# The `git` tool

The shared `.gitignore` and `.gitattributes`, the commit convention, the
git-config check, the merge tasks and worktrunk. The `base` pack adds it with
`required = true`, so `bootstrap remove` refuses it.

| Property     | Value             |
| ------------ | ----------------- |
| Root only    | No                |
| Depends      | None              |
| Parent       | None              |
| Default envs | dev               |
| Seeds        | `.config/wt.toml` |

## When the tool is relevant

Always. The `base` pack adds this tool. Do not select it in the values file.
A `[tools.git]` table gives values.

## What the tool ships

| Path in the repository                  | Kind      | Use                                                                                                           |
| --------------------------------------- | --------- | ------------------------------------------------------------------------------------------------------------- |
| `.gitignore`                            | Shared    | The shared ignore lines. See below.                                                                           |
| `.gitattributes`                        | Synced    | LF line endings, the generated lockfiles, and the binary files of each stack.                                 |
| `.config/git-conventional-commits.yaml` | Template  | The commit types, and the scopes from `commit_scopes`, else the members of the repository.                    |
| `.config/wt.toml`                       | Seed      | The worktrunk config: the `pre-start` steps and the commit convention (see "worktrunk").                      |
| `.config/mise/tasks/_scripts/merge`     | Library   | `merge_to_destination_branch`, which the merge tasks use.                                                     |
| `.config/mise/tasks/code/git-config`    | Synced    | The task `code:git-config`.                                                                                   |
| `.config/mise/tasks/code/merge/develop` | Synced    | The task `code:merge:develop`. Only with `when = "git.branch_pair"`: the repository has `main` and `develop`. |
| `.config/mise/tasks/code/merge/main`    | Synced    | The task `code:merge:main`. Only with `when = "git.branch_pair"`.                                             |
| `.config/mise/conf.d/git/mise.dev.toml` | Generated | The mise config of this tool.                                                                                 |

`.gitignore` is a shared file. The spec `files/gitignore.merge.toml` merges
into it line by line: the sync adds each missing line of `bootstrap` at the
end. The repository keeps its own lines directly in `.gitignore`, and the sync
keeps them. A later line wins, so an own line with `!` cancels a shared
pattern. To remove a line of `bootstrap`, write an override in the values
file:

```toml
[tools.git.files.".gitignore"]
"*.log" = { "$mode" = "remove" }
```

`"$skip" = true` keeps the sync from adding a line. The merge rules are in
[`docs/design/tools.md`](../../design/tools.md).

### worktrunk

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

## mise tools

| mise tool   | Version  | Environments | Note                           |
| ----------- | -------- | ------------ | ------------------------------ |
| `worktrunk` | `latest` | dev          | The git worktree manager `wt`. |

The tool is in `dev` only. Its files and hooks do not depend on an
environment, so a CI job still gets the merge tasks and the hooks.

## Tasks

| Task                          | Use                                                                |
| ----------------------------- | ------------------------------------------------------------------ |
| `code:git-config`             | Find forbidden keys in the local git config. `--fix` removes them. |
| `code:merge:develop <branch>` | Land a branch on `develop`, with `main` and `develop`.             |
| `code:merge:main`             | Land `develop` on `main`, with `main` and `develop`.               |

The `mise` tool ships `code:merge <branch>` for a repository with one of
`main` and `develop`. The derive script gives `branch_pair`, and the `when`
of each merge task selects the tasks of the branch model. The sync removes
the merge tasks of the other model.

The merge tasks read `[tools.git.branch.<destination>]` of the
`.config/bootstrap.toml` of the repository, with `yq`. Its `merge_model`
selects `direct` (the default) or `pr`. Its `merge_from` or `merge_never`
limits the source branches. Without both lists, only `develop` merges into
`main` when `develop` exists, and nothing merges from `main`. A table with
both lists stops the merge. The `pr` model needs the CLI of the forge, from
the `github` or the `gitlab` tool. In a repository without an `origin`
remote, `direct` merges locally without the pull and the push, and `pr` stops
before the hooks run. Before a merge, the merge tasks run the hooks over all
files, with `SKIP=no-commit-to-branch`.

A merge task stops when the repository of `MISE_ORIGINAL_CWD` is not the
repository of the task. In a submodule without that merge task, mise finds the
task of the root repository and runs it in the root folder. Without the stop,
that task lands a branch of the root. Run `mise run setup:configs` in the
submodule to get its own merge tasks.

## Scripts

The sync runs these scripts from the clone, and never copies them.

| Script           | Kind   | Use                                                                                                                                                                                       |
| ---------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scripts/derive` | derive | It gives `branch_pair` (the repository has `main` and `develop`) and `protected_branches` (`main` and `develop`, without each branch with `direct_commit = true`).                        |
| `scripts/check`  | check  | It examines the branches and `direct_commit`. Neither `main` nor `develop`, a bad `direct_commit`, or `direct_commit` on `main` only when both branches exist stops the sync with code 2. |

The `no-commit-to-branch` hook gets one `--branch` pair for each branch of
`protected_branches`. Without a protected branch, the sync removes the hook.

## Contributions

| Tool         | Shared file                      | Items                                                                                                                                                                     |
| ------------ | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pre-commit` | `.config/pre-commit-config.yaml` | The hook blocks `git-config` (calls `code:git-config --fix`), `no-commit-to-branch` of `pre-commit/pre-commit-hooks`, and `conventional-commits` in the commit-msg stage. |

The tools that add lines to `.gitignore` say so in their own tool file.

## Values

The keys of `[tools.git]` in `.config/bootstrap.toml`:

| Key                                         | Use                                                                               |
| ------------------------------------------- | --------------------------------------------------------------------------------- |
| `env`                                       | The environments of the tool.                                                     |
| `commit_scopes`                             | The scopes of the commit convention. The default is the member list.              |
| `[tools.git.branch.<name>]` `direct_commit` | `true` allows a direct commit to `main` or `develop`. See "Step 2" of `index.md`. |
| `[tools.git.branch.<name>]` `merge_model`   | `direct` or `pr`.                                                                 |
| `[tools.git.branch.<name>]` `merge_from`    | The only branches that merge into the branch.                                     |
| `[tools.git.branch.<name>]` `merge_never`   | The branches that never merge into the branch.                                    |
| `[tools.git.files.".gitignore"]`            | The overrides of the shared lines (see "What the tool ships").                    |

No questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Select the branch model ("Step 2: Select the branch model" of
   [`index.md`](../index.md)).
2. Set `commit_scopes` in `[tools.git]` ("Step 6: Write the other values and
   sync again").
3. Keep the ignore lines of the repository in `.gitignore`. For each shared
   line that the repository must not ignore, write a `"$mode" = "remove"`
   override.
4. If the repository has its own lines in `.gitattributes`, tell the user. The
   sync replaces the file, and it has no overrides.

## Migration

- The general items M10, M11 and M18 of `migrate.md` are about the files of
  this tool.
- M34 of `migrate.md`: `bootstrap init` moves the top-level `commit_scopes`
  and `[branch.*]` into `[tools.git]`. It moves the `add` lines of
  `[overrides.gitignore]` into `.gitignore`, and its `remove` lines into
  `[tools.git.files.".gitignore"]`.
