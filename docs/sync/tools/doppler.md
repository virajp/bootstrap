# The `doppler` tool

The Doppler CLI and `setup:secrets:doppler`. A repository selects it only when
it keeps its secrets in Doppler.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

- The repository has `doppler.yaml`.
- A script or a task calls `doppler run` or `doppler secrets`.

Do not select both `doppler` and `fnox` without a reason. Ask the user which
one keeps the secrets of the repository.

## What the tool ships

| Path in the repository                      | Kind      | Use                               |
| ------------------------------------------- | --------- | --------------------------------- |
| `.config/mise/tasks/setup/secrets/doppler`  | Synced    | The task `setup:secrets:doppler`. |
| `.config/mise/conf.d/doppler/mise.dev.toml` | Generated | The mise config of this tool.     |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `doppler` | `latest` | dev          |

Only dev. The `ci` environment has its own secret management.

## Tasks

| Task                    | Use                                                                                                                                                             |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:secrets:doppler` | Select the Doppler project and config from `doppler.yaml`. It skips without `doppler.yaml`, without a login, or with a warning when `doppler` is not installed. |

The task does not run `doppler login`, because that command opens a browser.

## Pre-commit hooks

None.

## Files the repository supplies

| File           | Used by                                                           |
| -------------- | ----------------------------------------------------------------- |
| `doppler.yaml` | `setup:secrets:doppler`. It names the Doppler project and config. |

## Adapt to the repository

1. If `doppler.yaml` is missing, ask the user for the project and the config.

## Migration

- Remove `doppler` from the repository-owned mise files.
- Delete an old `setup:secrets` task for Doppler. `setup:secrets:doppler`
  replaces it (M3 of `migrate.md`).
