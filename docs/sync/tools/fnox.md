# The `fnox` tool

The secrets from `.config/fnox.toml`, checked by `setup:secrets:fnox`.

| Property     | Value               |
| ------------ | ------------------- |
| Root only    | No                  |
| Depends      | None                |
| Default envs | dev                 |
| Seeds        | `.config/fnox.toml` |

## When the tool is relevant

- The repository has `.config/fnox.toml` or `fnox.toml`.
- A script or a task calls `fnox exec`.
- The repository needs development secrets in the OS keychain.

Do not select both `fnox` and `doppler` without a reason. Ask the user which
one keeps the secrets of the repository.

## What the tool ships

| Path in the repository                   | Kind           | Use                                                       |
| ---------------------------------------- | -------------- | --------------------------------------------------------- |
| `.config/fnox.toml`                      | Seed, template | The fnox config, with a keychain provider and no secrets. |
| `.config/mise/tasks/setup/secrets/fnox`  | Synced         | The task `setup:secrets:fnox`.                            |
| `.config/mise/conf.d/fnox/mise.dev.toml` | Generated      | The mise config of this tool.                             |

The seed has `root = true` and `env = "exec"`. Thus fnox does not search the
parent directories, and the secrets go only to `fnox exec` subprocesses.

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `fnox`    | `latest` | dev          |

Only dev. The `ci` environment has its own secret management.

The tool sets `FNOX_CONFIG_DIR` to `.config/fnox/`. This folder has no
`config.toml`, so the global `~/.config/fnox/config.toml` does not load in the
repository. Do not put a `config.toml` in `.config/fnox/`.

## Tasks

| Task                 | Use                                                                                                                                                                                                                                                                             |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:secrets:fnox` | Make sure that each secret that fnox needs is configured: `fnox check` for the `default` profile, then `fnox -P <profile> check` for each other profile. It runs every check, and it fails at the end and names each profile that failed. It skips without `.config/fnox.toml`. |

The task needs fnox 1.39.0 or later.

The task checks each profile in its own call. `fnox -P a,b check` merges the
profiles into one overlay, where a later profile replaces a secret of the same
name, so it can hide a missing secret. A profile inherits the top-level
secrets, so its check examines them again.

## Pre-commit hooks

None.

## Files the repository supplies

None. The seed supplies `.config/fnox.toml`. Add each secret with
`mise x -- fnox set <NAME>`.

## Adapt to the repository

1. In a repository of a product with members, set `product_name` in the values
   file. Then the keychain provider uses `service = "<product_name>"` and
   `prefix = "<repo_name>/"`. Without it, `service` is `repo_name`.
2. Load the secrets with `mise x -- fnox exec -- <command>`. Do not use
   `--config .config/fnox.toml`: fnox finds the file by itself, and an
   explicit `--config` still loads the global config.

## Migration

1. If the repository has `fnox.toml` at its root, move it to
   `.config/fnox.toml`. Then the seed does not replace it.
2. Change each `fnox exec` call to `mise x -- fnox exec`. Remove each
   `--config .config/fnox.toml`.
3. If `mise x -- fnox --version` shows a version older than 1.39.0, run
   `mise upgrade fnox`. fnox 1.36.0 does not search `.config/`.
4. `.config/fnox.toml` is a seed, so the sync does not change an existing
   file. Change it by hand:
   1. Add `root = true` before `[providers]`. Then a submodule does not get
      the secrets of its root.
   2. In a repository of a product with members, set `product_name`. Then
      change the keychain provider: `service = "<product_name>"` and
      `prefix = "<repo_name>/"`. Move the existing secrets to the new names.
5. Remove `fnox` from the repository-owned mise files.
6. Delete an old `setup:secrets` task for fnox. `setup:secrets:fnox`
   replaces it (M3 of `migrate.md`).
