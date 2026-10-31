# The `repo-sync` tool

The sync function `repo_sync` and the `setup:configs` tasks that copy the
tools from `bootstrap`.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not list it in the values
file.

## What the tool ships

| Path in the repository                        | Kind      | Use                                                                 |
| --------------------------------------------- | --------- | ------------------------------------------------------------------- |
| `.config/mise/tasks/_scripts/repo-sync`       | Synced    | The library with the sync function `repo_sync`.                     |
| `.config/mise/tasks/setup/configs/_default`   | Synced    | The task `setup:configs`. It fetches `bootstrap` into the cache.    |
| `.config/mise/tasks/setup/configs/sync`       | Synced    | The task `setup:configs:sync`. It loads `repo-sync` from the clone. |
| `.config/mise/tasks/code/lint/repo-sync`      | Synced    | The gate `code:lint:repo-sync`. It runs `repo_sync --check`.        |
| `.config/mise/conf.d/repo-sync/mise.dev.toml` | Generated | The mise config of this tool.                                       |

## mise tools

| mise tool  | Version  | Environments | Note                  |
| ---------- | -------- | ------------ | --------------------- |
| `gomplate` | `latest` | dev          | Renders the templates |

## Tasks

| Task                  | Use                                                                     |
| --------------------- | ----------------------------------------------------------------------- |
| `setup:configs`       | Sync the shared files from `bootstrap`. `--dry-run` shows the diff.     |
| `setup:configs:sync`  | Run `repo_sync` from the clone of `setup:configs`. Do not run it alone. |
| `code:lint:repo-sync` | Validate `.config/bootstrap.toml` with `repo_sync --check`. Hidden.     |

`setup:configs` keeps a shallow clone of `bootstrap` in
`${XDG_CACHE_HOME:-~/.cache}/bootstrap`, for each consumer of the machine:

- Each run fetches the default branch into the clone, and gives a copy of it
  to `setup:configs:sync`.
- The lock folder `bootstrap.lock` next to the clone makes a second run wait
  until the first has made its copy. A waiter removes the lock of a dead
  holder.
- A clone with another `origin`, for example after a test with
  `REPO_SYNC_URL`, is cloned again.
- If the fetch or the clone fails, `setup:configs` stops with code 5. It
  never uses an old clone.

`code:lint:repo-sync` is a gate of `code:lint`. It accepts `--debug` and
`--fix`, and ignores them, because the check has no fixes. It prints nothing
and exits 0 in these conditions:

- gomplate is not available (`mise x -- which gomplate`). The gate does not
  install it, so a commit never needs the network.
- A list of files does not include `.config/bootstrap.toml`.
- No list of files is given, and `.config/bootstrap.toml` does not exist.

Else it unsets `REPO_SYNC_CLONE`, loads the local `_scripts/repo-sync`, and
runs `repo_sync --check`.

`repo_sync --check` examines only the rules that need the values file and the
local git repository. It runs after the checks of the values file and of
gomplate, and it stops before the clone. It makes no clone, writes no file and
does not set `bootstrap_commit`. It runs each local check, prints each
message, and returns the highest code. `--check` with `--dry-run` is a usage
error, code 2. These rules need the clone, so only `setup:configs` examines
them: a tool or a pack that is not in `bootstrap`, a pack loop or a pack named
like a tool, an environment that a mise tool does not allow, an
`[overrides.mise*]` path that no selected tool has,
`[overrides.graphifyignore]` without the `claude` tool, a `remove` line that
the shared file does not have, and a `root_only` tool in a submodule.

Besides the copy of the files, `repo_sync` does these steps:

- It writes the merge tasks of the branch model that it finds, and removes the
  others. If neither `main` nor `develop` exists, it stops with code 2.
- It removes each `conf.d/*/mise*.toml` with the bootstrap header that it did
  not write.
- It applies `[overrides]` of the values file to the generated mise config
  and to `.gitignore`. Each overridden line gets an `# override:` comment.
  A bad `[overrides]` stops the sync with code 2. See "Overrides" in step 3
  of `index.md`.
- It validates `direct_commit` of the `[branch.*]` tables before it writes a
  file. These values stop the sync with code 2: a value that is not a
  boolean, the key in a table other than `[branch.main]` and
  `[branch.develop]`, the key for a core branch that does not exist (also
  `false`), and `direct_commit = true` for `main` but not for `develop` when
  both branches exist.
- It changes the `no-commit-to-branch` hook in `.precommit`, the hooks of the
  tools and of `.config/repo-sync/pre-commit.yaml`. It removes each
  `"--branch", "<branch>"` pair of a core branch that does not exist or that
  has `direct_commit = true`. A hook with no pair left goes, and an entry with
  no hook left goes too. Thus a repository with one core branch gets only
  `--branch <its branch>`.
- After a sync without a failure, it sets the top-level `bootstrap_commit` of
  the values file to the full commit of its clone. It replaces only that
  line, or adds it after the header comments. A dry run prints
  `would set bootstrap_commit <commit>`. See "A repository that already has
  `.config/bootstrap.toml`" of `index.md`.

These environment variables change the sync. `--check` does not read
`REPO_SYNC_URL` and `REPO_SYNC_CLONE`, because it makes no clone:

| Variable                  | Use                                                                    |
| ------------------------- | ---------------------------------------------------------------------- |
| `REPO_SYNC_URL`           | The repository to clone. The default is the `bootstrap` repository.    |
| `XDG_CACHE_HOME`          | The folder of the clone of `setup:configs`. The default is `~/.cache`. |
| `REPO_SYNC_CLONE`         | An existing clone. `setup:configs` sets it.                            |
| `REPO_SYNC_VALUES`        | The values file. The default is `.config/bootstrap.toml`.              |
| `REPO_SYNC_NO_AUTOUPDATE` | `1` skips `pre-commit autoupdate`, so no network is necessary.         |

## Pre-commit hooks

None. The `lint` hook of the `linter` tool runs `code:lint`, and `code:lint`
runs `code:lint:repo-sync` with the staged files.

## Files the repository supplies

| File                                | Used by                                                                                          |
| ----------------------------------- | ------------------------------------------------------------------------------------------------ |
| `.config/bootstrap.toml`            | `repo_sync`. The values file of step 3 of `index.md`. `repo_sync` writes its `bootstrap_commit`. |
| `.config/repo-sync/pre-commit.yaml` | `.config/pre-commit-config.yaml`. Optional: its own hooks.                                       |

## Adapt to the repository

1. Copy the sync files by hand before the first sync (step 2 of `index.md`).

## Migration

- Remove `gomplate` from the repository-owned mise files.
- The general items M1, M15 and M23 of `migrate.md` are about the files of
  this tool.
