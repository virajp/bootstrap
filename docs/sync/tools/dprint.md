# The `dprint` tool

dprint: the formatter, `code:format` and `setup:dprint`.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not list it in the values
file.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/tasks/code/format/_default`    | Template  | The task `code:format`.       |
| `.config/mise/tasks/setup/dprint`            | Synced    | The task `setup:dprint`.      |
| `.config/mise/conf.d/dprint/mise.<env>.toml` | Generated | The mise config of this tool. |

VS Code extension: `dprint.dprint`.

## mise tools

| mise tool | Version  | Environments | Note      |
| --------- | -------- | ------------ | --------- |
| `dprint`  | `latest` | dev, ci      | Formatter |

## Tasks

| Task           | Use                                                                                                                                                                                                                                         |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `code:format`  | Run dprint, then each `code:format:<name>`, on the `[files]` arguments, else on the whole tree. `--fix` applies the fixes, and `--debug` prints debug logs. With members, `--all` and `--<member>` also run `code:format` in those members. |
| `setup:dprint` | Upgrade the dprint plugins.                                                                                                                                                                                                                 |

## Pre-commit hooks

- The local hook `format`, which calls `code:format --fix` with the files of
  the commit.

## Files the repository supplies

| File                            | Used by                                                                                                                                               |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `dprint.jsonc`                  | `code:format`. It must have `"extends": "~/.config/dprint/dprint.jsonc"`, which has the plugins. It only adds `excludes`, for example the submodules. |
| `~/.config/dprint/dprint.jsonc` | `setup:dprint`                                                                                                                                        |

## Adapt to the repository

1. Move each repository formatter step to a task `code:format:<name>`.

## Migration

- Remove `dprint` from the repository-owned mise files.
- Remove shfmt. dprint formats shell scripts now.
- Remove taplo. dprint formats TOML now. Also remove the `[toml]` and
  `evenBetterToml.*` keys from `.vscode/settings.json`.
