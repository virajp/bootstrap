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

| Path in the repository                      | Kind      | Use                                                                                                   |
| ------------------------------------------- | --------- | ----------------------------------------------------------------------------------------------------- |
| `.gitignore`                                | Shared    | The shared ignore lines. See below.                                                                   |
| `.gitattributes`                            | Synced    | LF line endings, the generated lockfiles, and the binary files of each stack.                         |
| `.config/git-conventional-commits.yaml`     | Template  | The commit types, and the scopes from `[commits].scopes`, else the names of the `[members.*]` tables. |
| `.config/wt.toml`                           | Seed      | The worktrunk config: the `pre-start` steps and the commit convention (see "worktrunk").              |
| `.config/mise/tasks/_scripts/merge.py`      | Library   | `merge_to_destination_branch`, which the merge tasks use.                                             |
| `.config/mise/tasks/_scripts/release.py`    | Library   | The steps of the release tasks.                                                                       |
| `.config/mise/tasks/code/git-config.py`     | Synced    | The task `code:git-config`.                                                                           |
| `.config/mise/tasks/git/pull.py`            | Synced    | The task `git:pull`.                                                                                  |
| `.config/mise/tasks/git/merge/develop.py`   | Synced    | The task `git:merge:develop`. Only with `when = "git.has_develop"`: `core_branches` has `develop`.    |
| `.config/mise/tasks/git/merge/main.py`      | Synced    | The task `git:merge:main`. Only with `when = "git.has_main"`: `core_branches` has `main`.             |
| `.config/mise/tasks/git/tag.py`             | Synced    | The task `git:tag`.                                                                                   |
| `.config/mise/tasks/release/<p>/<k>.py`     | Template  | The task `release:<p>:<k>`, one for each `[release.<p>.<k>]` table of `repo.toml`.                    |
| `.config/mise/tasks/git/worktree/create.py` | Synced    | The task `git:worktree:create`.                                                                       |
| `.config/mise/tasks/git/worktree/delete.py` | Synced    | The task `git:worktree:delete`.                                                                       |
| `.config/mise/conf.d/git/mise.dev.toml`     | Generated | The mise config of this tool.                                                                         |

`.gitignore` is a shared file. The spec `files/gitignore.merge.toml` merges
into it line by line: the sync adds each missing line of `bootstrap` at the
end. The repository keeps its own lines directly in `.gitignore`, and the sync
keeps them. A later line wins, so an own line with `!` cancels a shared
pattern. Another selected tool can add lines in its
`[[contributes_items.git.".gitignore".item]]` tables, for example the vwf lines
of `virajp-plugins`: the sync adds them after the lines of the spec. The spec
no longer has the vwf lines (`docs/memory/handoff/`,
`docs/memory/doctor/`, `docs/memory/runs/`, `docs/scratchpad/` and
`!docs/plans/**/*.log`): `virajp-plugins` has them. To remove a line of
`bootstrap`, a contributed line too, write an override in the values file:

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
`bootstrap migrate` renames that command to `mise run git:worktree:setup`, and
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

| Task                                                                     | Use                                                                                                                                            |
| ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `code:git-config`                                                        | Find forbidden keys in the local git config. `--fix` removes them.                                                                             |
| `git:pull [--all]`                                                       | Bring the core branches and the default branch up to date from `origin`, here and in each member. `--all`: each branch of `origin`. See below. |
| `git:merge:develop [<branch>]`                                           | Land a branch on `develop`. The default is the current branch. Only when `core_branches` has `develop`.                                        |
| `git:merge:main [<branch>]`                                              | Land a branch on `main`. The default is `develop` when it is a core branch, else the current branch. Only when `core_branches` has `main`.     |
| `git:tag <name> [--message <text>] [--push]`                             | Tag `HEAD` with a signed tag. `--push` pushes only the tag. See below.                                                                         |
| `release:<project>:<kind> [--push] [--dry-run]`                          | Release the project as its `[release.<project>.<kind>]` table says. See below.                                                                 |
| `git:worktree:create <branch> [--from <base> \| --current] [--no-setup]` | Add a worktree for the branch, then run `mise trust` and `git:worktree:setup` in it. See below.                                                |
| `git:worktree:delete <branch>`                                           | Remove the worktree of the branch, and the branch when a core branch has each of its commits. See below.                                       |

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
`pr` stops before the hooks run. When `core_branches` has `main` and
`develop`, a direct landing of `develop` on `main` is a fast-forward, so `main`
is always an ancestor of `develop`. If `main` or
`origin/main` has a commit that `develop` does not have, the task stops before
it changes a branch: merge `main` into `develop`, push `develop`, then land
again. Each other direct landing is a `--no-ff` merge. Before a merge, the
merge tasks run the hooks
over all
files, with `SKIP=no-commit-to-branch`.

A merge task stops when `MISE_ORIGINAL_CWD`, the folder of the shell, is in
the tree of the repository of the task but in another git repository, for
example a submodule. A run with `mise run --cd <repo>` or `mise -C <repo>` from
a folder outside that tree passes: mise keeps the folder of the shell in
`MISE_ORIGINAL_CWD`, and no variable gives `<repo>`. In a submodule without that
merge task, mise finds the
task of the root repository and runs it in the root folder. Without the stop,
that task lands a branch of the root. Run `mise run setup:configs` in the
submodule to get its own merge tasks.

`git:pull` fetches `origin` in the repository and in each submodule and
sibling member on the machine. For each core branch and the default branch, it
creates a missing local branch from `origin`, and fast-forwards a local branch
that `origin` is ahead of. It keeps a branch that has local commits, and warns
about a branch that has diverged. `--all` does the same for each branch of
`origin`.

`git:worktree:create` adds the worktree in `.worktrees/<branch>` of the main
checkout, or in `worktrees/<branch>` when only `worktrees/` exists. It checks
out a local branch. A branch only on `origin` becomes a local branch that
tracks it: the task first prints a warning, because it trusts and runs the mise
config of that branch as is. Any other branch is new, from `--from`, from the
current branch with `--current`, else from the default branch
(`default_branch` of `repo.toml`, local or on `origin`). A new branch gets
`--no-track`, so it does not track its base: `git:merge:<b>` reads the
upstream of the branch to find unpushed commits, and a branch that tracks
`origin/<b>` shows each new commit as unpushed. `--no-setup` skips
`git:worktree:setup`; `mise trust`
always runs. The task stops with code 2 for a core branch, for `--from` and
`--current` together, and for either flag with a branch that exists. It stops
with code 1 when the path exists or git
refuses the add. When the trust or the setup fails, the worktree stays, and
the task prints the commands that set it up again or delete it.

`git:worktree:delete` finds the worktree of the branch at any path, and runs
`git worktree remove`: git refuses a worktree with changes. It deletes the
branch with `git branch -D` only when a core branch, local or on `origin`, has
each of its commits. Else the branch stays, and the task prints the `-D`
command and exits 1. It stops with code 2 for a core branch and inside the
worktree that it removes, and with code 1 when no worktree has the branch.

`git:tag` stops, and changes nothing, when the tree is not clean, when the
tag exists here or on `origin`, and with `--push` when no branch of `origin`
has `HEAD`. A `release:<project>:<kind>` task of a `task` kind runs its
`preflight` tasks, then its `task` with its `args`. A task of a `tag` kind reads
the level of
`releases.<project>` of `.config/bootstrap/releases.toml`, bumps the version,
makes one commit `ops: release <tag>` with the version file and the cleared
level, lands it as `landing` of its branch says, and calls `git:tag`. Its
work branch `release/<tag>` gets `--no-track`, as a new worktree branch does. A
kind
with `from` promotes the last version of the other kind. A release commit
without its tag gets the tag on the next run. The full rules are in
[`docs/design/repo.md`](../../design/repo.md#the-release-tasks).

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

The tools that add lines to `.gitignore`, in
`[[contributes_items.git.".gitignore".item]]`, say so in their own tool file,
for example `virajp-plugins`.

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
| `[release.<project>.<kind>]`      | One release task for each table. The keys are in `bootstrap-config.md`.                                          |

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
- M34 of `migrate.md`: `bootstrap migrate` moves the top-level `commit_scopes`
  and `[branch.*]` into `[tools.git]`. It moves the `add` lines of
  `[overrides.gitignore]` into `.gitignore`, and its `remove` lines into
  `[tools.git.files.".gitignore"]`.
- M41 of `migrate.md`: `bootstrap migrate` moves `commit_scopes` to
  `[commits].scopes`
  and `[tools.git.branch.*]` to `[branches.*]` of `repo.toml`, with `landing`
  in place of `merge_model` and `direct_commit`. The merge tasks and the
  scripts need `bootstrap` 4.0.0 or newer.
- M44 of `migrate.md`: the spec no longer has the vwf lines. Add the
  `virajp-plugins` tool to keep them.
- M52 of `migrate.md`: `release_env` and `[deploy.<env>]` are not keys of
  `repo.toml`.
- M53 of `migrate.md`: an own release task moves to a
  `[release.<project>.<kind>]` table.
- M54 of `migrate.md`: the type `local` of a release table is now `task`.
