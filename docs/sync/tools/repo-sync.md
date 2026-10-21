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

## Pre-commit hooks

None.

## Files the repository supplies

| File                     | Used by                                               |
| ------------------------ | ----------------------------------------------------- |
| `.config/bootstrap.toml` | `repo_sync`. The values file of step 3 of `index.md`. |

## Adapt to the repository

1. Copy the sync files by hand before the first sync (step 2 of `index.md`).

## Migration

- Remove `gomplate` from the repository-owned mise files.
- The general items M1 and M15 of `migrate.md` are about the files of this
  tool.
