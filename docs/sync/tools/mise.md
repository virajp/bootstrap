# The `mise` tool

The shared mise config, the task library, the `setup`, `code` and `test`
orchestrators, and the entry of the sync: `setup:configs` and the gate
`code:lint:60-bootstrap`. The `base` pack adds it with `required = true`, so
`bootstrap remove tool <tool…>` refuses it.

| Property     | Value                                         |
| ------------ | --------------------------------------------- |
| Root only    | No                                            |
| Depends      | None                                          |
| Parent       | None                                          |
| Default envs | dev, ci                                       |
| Seeds        | `.config/mise/conf.d/zz-repo/mise.<env>.toml` |

## When the tool is relevant

Always. The `base` pack adds this tool. Do not select it in the values file.
A `[tools.mise]` table gives values.

## What the tool ships

| Path in the repository                        | Kind      | Use                                                                                                                                                                             |
| --------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.config/mise.toml`                           | Template  | The shared `[settings]`, and the `[env]` values `MEMBERS` and `REPO_NAME` from `.config/bootstrap/repo.toml`.                                                                   |
| `.config/mise/conf.d/mise/mise.<env>.toml`    | Generated | The mise config of this tool, one file for each environment.                                                                                                                    |
| `.config/mise/conf.d/zz-repo/mise.<env>.toml` | Seed      | The mise config and the `[env]` values of the repository, one file for `dev` and one for `ci`. The folder name sorts last, so this config wins over each tool config.           |
| `.config/mise/tasks/_scripts/helpers`         | Library   | The print functions, `members`, `own_tasks` (the `NN-<name>` tasks of a group), `has_own_task`, `shell_files_in_scope` and `bootstrap_get`, the reader of `.config/bootstrap/`. |
| `.config/mise/tasks/_scripts/checks`          | Library   | The git predicates of the merge tasks.                                                                                                                                          |
| `.config/mise/tasks/_scripts/placeholder`     | Library   | The prompt of a task that no stack fills yet.                                                                                                                                   |
| `.config/mise/tasks/_scripts/worktrees-page`  | Library   | The HTML page of `code:worktrees`.                                                                                                                                              |
| `.config/mise/tasks/setup/_default`           | Template  | The task `setup`.                                                                                                                                                               |
| `.config/mise/tasks/setup/configs/_default`   | Template  | The task `setup:configs`.                                                                                                                                                       |
| `.config/mise/tasks/setup/mise`               | Synced    | The task `setup:mise`.                                                                                                                                                          |
| `.config/mise/tasks/setup/worktree`           | Template  | The task `setup:worktree`.                                                                                                                                                      |
| `.config/mise/tasks/code/_default`            | Synced    | The task `code`.                                                                                                                                                                |
| `.config/mise/tasks/code/format/_default`     | Template  | The task `code:format`.                                                                                                                                                         |
| `.config/mise/tasks/code/lint/_default`       | Template  | The task `code:lint`.                                                                                                                                                           |
| `.config/mise/tasks/code/lint/60-bootstrap`   | Synced    | The gate `code:lint:60-bootstrap`.                                                                                                                                              |
| `.config/mise/tasks/code/sec/_default`        | Synced    | The task `code:sec`.                                                                                                                                                            |
| `.config/mise/tasks/code/merge/_default`      | Synced    | The task `code:merge`. Only with `when = "!git.branch_pair"`: a repository with one of `main` and `develop`. See [git.md](git.md).                                              |
| `.config/mise/tasks/code/worktrees`           | Synced    | The task `code:worktrees`.                                                                                                                                                      |
| `.config/mise/tasks/test/_default`            | Synced    | The task `test`.                                                                                                                                                                |
| `.config/mise/tasks/test/e2e/_default`        | Synced    | The task `test:e2e`.                                                                                                                                                            |
| `.config/mise/tasks/test/golden/_default`     | Synced    | The task `test:golden`.                                                                                                                                                         |

The `mise` tool ships each orchestrator `_default`. The tool of a task group
ships its leaf tasks. A leaf task that an orchestrator runs with `own_tasks`
must have a number prefix, for example `code:lint:70-ruff`. The prefix gives
the run order. `own_tasks` takes only an executable file named `NN-<name>`,
and skips each other file, `_default` and each folder. `bootstrap check` warns
about a file without the prefix in the folder of an orchestrator.

## mise tools

| mise tool                 | Version  | Environments | Note                                                                                                                                                                                 |
| ------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `jq`                      | `latest` | dev, ci      | JSON queries                                                                                                                                                                         |
| `yq`                      | `latest` | dev, ci      | YAML queries                                                                                                                                                                         |
| `github:virajp/bootstrap` | `latest` | dev          | The CLI `bootstrap`, from `[mise_env.dev.tools]`. `setup:configs` and the gate run it. `minimum_release_age_excludes` of `.config/mise.toml` names it, so a new release has no wait. |

The tool also sets:

- The shell aliases `precommit`, `setup` and `worktrees`. With submodule or
  sibling members that have a `[members.<name>]` table, the derive script adds
  `setup-all` and `setup-<member>` (see "Scripts").
- The hidden task `init`, which makes the task scripts executable.
- `os = "macos"` in `dev` and `os = "linux"` in `ci`, from
  `[mise_env.dev.settings]` and `[mise_env.ci.settings]`.

## Tasks

| Task                     | Use                                                                                                                                                                                                                                  |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `setup`                  | Install the tools, the dependencies and the hooks, and run `setup:vscode` when the repository has it. `--upgrade` also upgrades them. With members, `--all` and `--<member>` also set up those members.                              |
| `setup:configs`          | Sync the shared files: it runs `bootstrap sync`. `--dry-run` shows the changes. `--verbose` lists each file. With members, `--all` and `--<member>` also sync those members.                                                         |
| `setup:mise`             | Reshim, doctor and install. `--upgrade` also upgrades.                                                                                                                                                                               |
| `setup:worktree`         | Prepare a new git worktree. It copies the `[worktree].files` of `repo.toml` from the main worktree, runs the `worktree_install` of each member, and it does less than `setup`.                                                       |
| `code`                   | Run `code:format`, `code:lint` and `code:sec`. `--debug` goes to each of them.                                                                                                                                                       |
| `code:format`            | Run each `code:format:<name>` task in number order, on the given files, else on the whole tree. `--fix` applies the fixes.                                                                                                           |
| `code:lint`              | Run each `code:lint:<name>` task in number order, on the given files, else on the whole tree. `--fix` applies the fixes.                                                                                                             |
| `code:lint:60-bootstrap` | Validate the files of `.config/bootstrap/` with `bootstrap check`. Hidden. See below.                                                                                                                                                |
| `code:sec`               | Run each `code:sec:<scanner>` task. `--staged` scans only the staged changes for secrets. See [_security_base.md](_security_base.md).                                                                                                |
| `code:merge <branch>`    | Land a branch on the only long-lived branch. See [git.md](git.md).                                                                                                                                                                   |
| `code:worktrees`         | List the worktrees and the branches of the product: the home repository (`[home]`), its submodules and its submodule and sibling members of `[members.*]`. A terminal gets a table and an HTML page. Another reader gets JSON Lines. |
| `test`                   | Run each `test:NN-<name>` task in number order. Each argument goes to each task unchanged, so a flag of a task needs no `--`. It does not run `test:e2e` or `test:golden`. With no task, it prints one line and exits 0.             |
| `test:e2e`               | Run each `test:e2e:NN-<name>` task, as `test` does.                                                                                                                                                                                  |
| `test:golden`            | Run each `test:golden:NN-<name>` task, as `test` does, for example `mise run test:golden --update`.                                                                                                                                  |

`setup:configs` gets the same member flags as `setup`. Only a submodule or a
sibling member with a `[members.<name>]` table gets a flag. A folder member of
a monorepo is part of the root tree, so it gets no flag. `code:format` and
`code:lint` have no member flags: each repository and each submodule runs its
own, and a task runs in a member only through `mise run --cd`. The same holds
for `test`, `test:e2e` and `test:golden`, for example
`mise run --cd <member> test`.

`setup`, `setup:configs` and `code:worktrees` read the members with
`bootstrap config get repo.members` through `members` of `helpers`, and
`setup` and `setup:configs` have `#MISE tools.yq="latest"`. They need
`bootstrap` 4.0.0 or newer: without an engine that answers, they stop with
code 2. A member that is not on the machine is skipped with a warning.

`setup:configs` runs `bootstrap sync` in the root, then in each selected
member, with `BOOTSTRAP_HOME` and `BOOTSTRAP_MEMBER`, so the sync of the member
writes its `[home]`. It skips a member without `.config/bootstrap/` with the
warning `member <name>: run bootstrap init there`. `BOOTSTRAP_BIN`, a local
build, replaces the `bootstrap` of mise.
The exit code of `bootstrap` is the exit code of the task. The clone, the
cache `~/.cache/bootstrap/` and the output of a sync are in
[`docs/design/cli.md`](../../design/cli.md).

`code:lint:60-bootstrap` runs `bootstrap check` against the cache clone
`~/.cache/bootstrap/repo`. It never fetches. Without that clone, without
`bootstrap`, or with another `origin` in the clone, it prints nothing and
exits 0. Thus no output does not show that the values are correct.

## Scripts

The sync runs these scripts from the clone, and never copies them.

| Script           | Kind   | Use                                                                                                                                                                                                                                    |
| ---------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scripts/derive` | derive | In a repository with submodule or sibling members that have a `[members.<name>]` table, it gives the dev aliases `setup-all` and `setup-<member>` as `mise.shell_alias`. It reads the tables with `bootstrap config get repo.members`. |

## Contributions

| Tool     | Shared file               | Items                                                                    |
| -------- | ------------------------- | ------------------------------------------------------------------------ |
| `vscode` | `.vscode/extensions.json` | `hverlin.mise-vscode`, `redhat.vscode-yaml`, `tamasfe.even-better-toml`. |

## Values

The keys of `[tools.mise]` in `.config/bootstrap/tools.toml`:

| Key   | Use                           |
| ----- | ----------------------------- |
| `env` | The environments of the tool. |

The tool reads these keys of `.config/bootstrap/repo.toml`:

| Key                | Use                                                                                                               |
| ------------------ | ----------------------------------------------------------------------------------------------------------------- |
| `[worktree].files` | The gitignored files that `setup:worktree` copies into a new worktree, for example `[".config/mise.local.toml"]`. |
| `[members.<name>]` | The members of `setup`, `setup:configs`, `setup:worktree` and `code:worktrees`.                                   |

No questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Write the values of the repository ("Step 6: Write the other values and
   sync again" of [`index.md`](../index.md)).
2. Move each shell function of the repository to
   `.config/mise/tasks/_scripts/local`.
3. Move each mise setting and `[env]` value of the repository to the seeds
   `.config/mise/conf.d/zz-repo/mise.<env>.toml`.

## Migration

- Remove `jq` and `yq` from the mise files of the repository.
- The general items M3, M5, M10, M14 and M17 of `migrate.md` are about the
  files of this tool.
- M34 of `migrate.md` installs `bootstrap`, and moves `worktree_files` into
  `[tools.mise]`. This tool now ships `setup:configs` and the gate
  `code:lint:60-bootstrap`, which replace the tasks of an old tool.
- M41 of `migrate.md`: `bootstrap init` moves `worktree_files` to
  `[worktree].files` of `repo.toml`. `code:format` and `code:lint` lose their
  member flags, and `code:worktrees` no longer reads `.config/vwf.yaml` or
  `.config/vwf-membership.yaml`.
