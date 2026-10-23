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
| `.config/mise/tasks/setup/configs/_default`   | Synced    | The task `setup:configs`. It clones `bootstrap`.                    |
| `.config/mise/tasks/setup/configs/sync`       | Synced    | The task `setup:configs:sync`. It loads `repo-sync` from the clone. |
| `.config/mise/conf.d/repo-sync/mise.dev.toml` | Generated | The mise config of this tool.                                       |

## mise tools

| mise tool  | Version  | Environments | Note                  |
| ---------- | -------- | ------------ | --------------------- |
| `gomplate` | `latest` | dev          | Renders the templates |

## Tasks

| Task                 | Use                                                                     |
| -------------------- | ----------------------------------------------------------------------- |
| `setup:configs`      | Sync the shared files from `bootstrap`. `--dry-run` shows the diff.     |
| `setup:configs:sync` | Run `repo_sync` from the clone of `setup:configs`. Do not run it alone. |

`setup:configs` stops with code 5 if the clone fails. Besides the copy of the
files, `repo_sync` does these steps:

- It writes the merge tasks of the branch model that it finds, and removes the
  others. If neither `main` nor `develop` exists, it stops with code 2.
- It removes each `conf.d/*/mise*.toml` with the bootstrap header that it did
  not write.
- It applies `[overrides]` of the values file to the generated mise config
  and to `.gitignore`. Each overridden line gets an `# override:` comment.
  A bad `[overrides]` stops the sync with code 2. See "Overrides" in step 3
  of `index.md`.
- After a sync without a failure, it sets the top-level `bootstrap_commit` of
  the values file to the full commit of its clone. It replaces only that
  line, or adds it after the header comments. A dry run prints
  `would set bootstrap_commit <commit>`. See "A repository that already has
  `.config/bootstrap.toml`" of `index.md`.

These environment variables change the sync:

| Variable                  | Use                                                                 |
| ------------------------- | ------------------------------------------------------------------- |
| `REPO_SYNC_URL`           | The repository to clone. The default is the `bootstrap` repository. |
| `REPO_SYNC_CLONE`         | An existing clone. `setup:configs` sets it.                         |
| `REPO_SYNC_VALUES`        | The values file. The default is `.config/bootstrap.toml`.           |
| `REPO_SYNC_NO_AUTOUPDATE` | `1` skips `pre-commit autoupdate`, so no network is necessary.      |

## Pre-commit hooks

None.

## Files the repository supplies

| File                                | Used by                                                                                          |
| ----------------------------------- | ------------------------------------------------------------------------------------------------ |
| `.config/bootstrap.toml`            | `repo_sync`. The values file of step 3 of `index.md`. `repo_sync` writes its `bootstrap_commit`. |
| `.config/repo-sync/pre-commit.yaml` | `.config/pre-commit-config.yaml`. Optional: its own hooks.                                       |

## Adapt to the repository

1. Copy the sync files by hand before the first sync (step 2 of `index.md`).

## Migration

- Remove `gomplate` from the repository-owned mise files.
- The general items M1 and M15 of `migrate.md` are about the files of this
  tool.
