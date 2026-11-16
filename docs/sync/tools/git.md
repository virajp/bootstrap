# The `git` tool

The shared `.gitignore` and `.gitattributes`, the commit convention, the
git-config check, the merge tasks, the worktree tasks and worktrunk. The `base`
pack adds it with `required = true`, so `bootstrap remove tool <tool…>` refuses
it.

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

| Path in the repository                   | Kind      | Use                                                                                                   |
| ---------------------------------------- | --------- | ----------------------------------------------------------------------------------------------------- |
| `.gitignore`                             | Shared    | The shared ignore lines. See below.                                                                   |
| `.gitattributes`                         | Synced    | LF line endings, the generated lockfiles, and the binary files of each stack.                         |
| `.config/git-conventional-commits.yaml`  | Template  | The commit types, and the scopes from `[commits].scopes`, else the names of the `[members.*]` tables. |
| `.config/wt.toml`                        | Seed      | The worktrunk config: the `pre-start` steps and the commit convention (see "worktrunk").              |
| `.config/mise/tasks/_scripts/merge`      | Library   | `merge_to_destination_branch`, which the merge tasks use.                                             |
| `.config/mise/tasks/code/git-config`     | Synced    | The task `code:git-config`.                                                                           |
| `.config/mise/tasks/git/merge/develop`   | Synced    | The task `git:merge:develop`. Only with `when = "git.has_develop"`: `core_branches` has `develop`.    |
| `.config/mise/tasks/git/merge/main`      | Synced    | The task `git:merge:main`. Only with `when = "git.has_main"`: `core_branches` has `main`.             |
| `.config/mise/tasks/git/worktree/create` | Synced    | The task `git:worktree:create`.                                                                       |
| `.config/mise/tasks/git/worktree/delete` | Synced    | The task `git:worktree:delete`.                                                                       |
| `.config/mise/conf.d/git/mise.dev.toml`  | Generated | The mise config of this tool.                                                                         |

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
3. `mise run git:worktree:setup`.

A `.config/wt.toml` from before bootstrap 4.0.0 runs `mise run setup:worktree`,
and no task has that name now. The sync does not write a seed again, so
`bootstrap init` renames that command to `mise run git:worktree:setup`, and
the rest of the file stays. Until then, `sync` and `check` stop with code 2.

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

| Task                                                        | Use                                                                                                                                        |
| ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `code:git-config`                                           | Find forbidden keys in the local git config. `--fix` removes them.                                                                         |
| `git:merge:develop [<branch>]`                              | Land a branch on `develop`. The default is the current branch. Only when `core_branches` has `develop`.                                    |
| `git:merge:main [<branch>]`                                 | Land a branch on `main`. The default is `develop` when it is a core branch, else the current branch. Only when `core_branches` has `main`. |
| `git:worktree:create <branch> [--from <base>] [--no-setup]` | Add a worktree for the branch, then run `mise trust` and `git:worktree:setup` in it. See below.                                            |
| `git:worktree:delete <branch>`                              | Remove the worktree of the branch, and the branch when a core branch has each of its commits. See below.                                   |

The name of a merge task is its destination. The derive script gives
`has_main` and `has_develop`, and the `when` of each merge task selects the
task of each core branch. The sync removes the merge task of a branch that is
not a core branch. The `mise` tool ships `git:worktree:setup` and
`git:worktree:list`.

The merge tasks read `[branches.<destination>]` of the
`.config/bootstrap/repo.toml` of the repository, with `bootstrap config get`.
They need `bootstrap` 4.0.0 or newer: without an engine that answers, a merge
stops with code 2. Its `landing` selects `direct-merge` (the default),
`direct-commit` (the same local merge) or `pr`. Its `merge_from` or
`merge_never`
limits the source branches. Without both lists, only `develop` merges into
`main` when `develop` exists, and nothing merges from `main`. A table with
both lists stops the merge. The `pr` model needs the CLI of the forge, from
the `github` or the `gitlab` tool. In a repository without an `origin`
remote, a direct landing merges locally without the pull and the push, and
`pr` stops before the hooks run. Before a merge, the merge tasks run the hooks
over all
files, with `SKIP=no-commit-to-branch`.

A merge task stops when the repository of `MISE_ORIGINAL_CWD` is not the
repository of the task. In a submodule without that merge task, mise finds the
task of the root repository and runs it in the root folder. Without the stop,
that task lands a branch of the root. Run `mise run setup:configs` in the
submodule to get its own merge tasks.

`git:worktree:create` adds the worktree in `.worktrees/<branch>` of the main
checkout, or in `worktrees/<branch>` when only `worktrees/` exists. It checks
out a local branch. A branch only on `origin` becomes a local branch that
tracks it: the task first prints a warning, because it trusts and runs the mise
config of that branch as is. Any other branch is new, from `--from`, else from
the current branch. `--no-setup` skips `git:worktree:setup`; `mise trust`
always runs. The task stops with code 2 for a core branch, and for `--from`
with a branch that exists. It stops with code 1 when the path exists or git
refuses the add. When the trust or the setup fails, the worktree stays, and
the task prints the commands that set it up again or delete it.

`git:worktree:delete` finds the worktree of the branch at any path, and runs
`git worktree remove`: git refuses a worktree with changes. It deletes the
branch with `git branch -D` only when a core branch, local or on `origin`, has
each of its commits. Else the branch stays, and the task prints the `-D`
command and exits 1. It stops with code 2 for a core branch and inside the
worktree that it removes, and with code 1 when no worktree has the branch.

## Scripts

The sync runs these scripts from the clone, and never copies them.

| Script           | Kind   | Use                                                                                                                                                                                                                                                                                                                                                                                         |
| ---------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scripts/derive` | derive | It gives `branch_pair` (`core_branches` has `main` and `develop`), `has_main` and `has_develop` (`core_branches` has `main`, or `develop`) and `protected_branches` (each core branch whose `landing` is not `direct-commit`). It reads them with `bootstrap config get`. An engine before 4.0.0, with the old values file, gets them from the branches that exist, locally or on `origin`. |
| `scripts/check`  | check  | A branch of `core_branches` that does not exist, locally or on `origin`, and is not the unborn branch that HEAD names, stops the sync with code 2. So does an engine that cannot answer `bootstrap config get`. The schema of the values examines the `[branches.*]` tables.                                                                                                                |

The `no-commit-to-branch` hook gets one `--branch` pair for each branch of
`protected_branches`. Without a protected branch, the sync removes the hook.

## Contributions

| Tool         | Shared file                      | Items                                                                                                                                                                     |
| ------------ | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pre-commit` | `.config/pre-commit-config.yaml` | The hook blocks `git-config` (calls `code:git-config --fix`), `no-commit-to-branch` of `pre-commit/pre-commit-hooks`, and `conventional-commits` in the commit-msg stage. |

The tools that add lines to `.gitignore` say so in their own tool file.

## Values

The keys of `[tools.git]` in `.config/bootstrap/tools.toml`:

| Key                              | Use                                                            |
| -------------------------------- | -------------------------------------------------------------- |
| `env`                            | The environments of the tool.                                  |
| `[tools.git.files.".gitignore"]` | The overrides of the shared lines (see "What the tool ships"). |

The tool reads these keys of `.config/bootstrap/repo.toml`:

| Key                               | Use                                                                                                              |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `core_branches`                   | The core branches, in order. Each one must exist.                                                                |
| `[branches.<name>]` `landing`     | `pr`, `direct-merge` (the default) or `direct-commit`, which allows a direct commit. See "Step 2" of `index.md`. |
| `[branches.<name>]` `merge_from`  | The only branches that merge into the branch.                                                                    |
| `[branches.<name>]` `merge_never` | The branches that never merge into the branch.                                                                   |
| `[commits]` `scopes`              | The scopes of the commit convention. The default is the names of the `[members.*]` tables, sorted.               |

No questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Select the branch model ("Step 2: Select the branch model" of
   [`index.md`](../index.md)).
2. Set `[commits].scopes` with `bootstrap config set repo.commits.scopes`
   ("Step 6: Write the other values and sync again").
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
- M41 of `migrate.md`: `bootstrap init` moves `commit_scopes` to
  `[commits].scopes`
  and `[tools.git.branch.*]` to `[branches.*]` of `repo.toml`, with `landing`
  in place of `merge_model` and `direct_commit`. The merge tasks and the
  scripts need `bootstrap` 4.0.0 or newer.
