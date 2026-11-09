# The `fnox` tool

The secrets from `.config/fnox.toml`, checked by `setup:secrets:20-fnox`.

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

| Path in the repository                     | Kind           | Use                                                                            |
| ------------------------------------------ | -------------- | ------------------------------------------------------------------------------ |
| `.config/fnox.toml`                        | Seed, template | The fnox config, with a keychain provider and the empty profile `<repo_name>`. |
| `.config/mise/tasks/setup/secrets/20-fnox` | Synced         | The task `setup:secrets:20-fnox`.                                              |
| `.config/mise/conf.d/fnox/mise.dev.toml`   | Generated      | The mise config of this tool.                                                  |

The seed has `root = true` and `env = "exec"`. Thus fnox does not search the
parent directories, and the secrets go only to `fnox exec` subprocesses.

## Profiles

A repository does not use the fnox `default` profile. The `default` profile is
only for the user-level fnox settings.

- The secrets live only in `[profiles.<name>.secrets]` tables of
  `.config/fnox.toml`. The top-level `[secrets]` table is the `default`
  profile, so the file does not have it, also not an empty one.
- A repository with one profile uses its `repo_name` as the profile name. The
  seed has the empty table `[profiles.<repo_name>.secrets]`.
- A repository with more sets of secrets, for example one for each Pulumi
  stack, adds one `[profiles.<name>.secrets]` table for each set. The names
  are free, but not `default`.
- Each `fnox` call names the profile with `-P <name>`. Without `-P`, fnox uses
  `default`, and the command gets no secret. The tool does not set
  `FNOX_PROFILE`.

The sync and `repo_sync --check` stop with code 2 when `.config/fnox.toml`
has a top-level `[secrets]` table, a `[profiles.default]` table, or no named
profile. The message names M25 of `migrate.md`. The check applies when
`.config/fnox.toml` exists. The lint gate `code:lint:60-repo-sync` runs
`--check` only when `.config/bootstrap.toml` is staged, so a commit that
changes only `.config/fnox.toml` does not run it. The next sync does.

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `fnox`    | `latest` | dev          |

Only dev. The `ci` environment has its own secret management.

The tool sets `FNOX_CONFIG_DIR` to `.config/fnox/`. This folder has no
`config.toml`, so the global `~/.config/fnox/config.toml` does not load in the
repository. Do not put a `config.toml` in `.config/fnox/`.

## Tasks

| Task                    | Use                                                                                                                                                                                                                                                                                               |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:secrets:20-fnox` | Make sure that each secret of each profile is configured: `fnox -P <profile> check` for each named profile. It does not check `default`. It runs every check, and it fails at the end and names each profile that failed. It fails without a named profile. It skips without `.config/fnox.toml`. |

The task needs fnox 1.39.0 or later.

The task checks each profile in its own call. `fnox -P a,b check` merges the
profiles into one overlay, where a later profile replaces a secret of the same
name, so it can hide a missing secret.

## Pre-commit hooks

None.

## Files the repository supplies

None. The seed supplies `.config/fnox.toml`. Add each secret with
`mise x -- fnox -P <profile> set <NAME>`.

## Adapt to the repository

1. In a repository of a product with members, set `product_name` in the values
   file. Then the keychain provider uses `service = "<product_name>"` and
   `prefix = "<repo_name>/"`. Without it, `service` is `repo_name`.
2. If the repository has more sets of secrets, add one
   `[profiles.<name>.secrets]` table for each set. Keep the table of
   `<repo_name>` only if it is one of the sets.
3. Load the secrets with `mise x -- fnox -P <profile> exec -- <command>`. Do not
   use
   `--config .config/fnox.toml`: fnox finds the file by itself, and an
   explicit `--config` still loads the global config.

## Migration

1. If the repository has `fnox.toml` at its root, move it to
   `.config/fnox.toml`. Then the seed does not replace it.
2. Change each `fnox exec` call to `mise x -- fnox -P <profile> exec`.
   Remove each `--config .config/fnox.toml`.
3. If `mise x -- fnox --version` shows a version older than 1.39.0, run
   `mise upgrade fnox`. fnox 1.36.0 does not search `.config/`.
4. `.config/fnox.toml` is a seed, so the sync does not change an existing
   file. Change it by hand:
   1. Add `root = true` before `[providers]`. Then a submodule does not get
      the secrets of its root.
   2. In a repository of a product with members, set `product_name`. Then
      change the keychain provider: `service = "<product_name>"` and
      `prefix = "<repo_name>/"`. Move the existing secrets to the new names.
   3. Move each secret of the top-level `[secrets]` table to a named profile,
      and delete the table (M25 of `migrate.md`).
5. Remove `fnox` from the repository-owned mise files.
6. Delete an old `setup:secrets` task for fnox. `setup:secrets:20-fnox`
   replaces it (M3 of `migrate.md`).
