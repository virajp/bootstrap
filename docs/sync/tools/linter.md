# The `linter` tool

The house linter and ShellCheck, `code:lint:05-linter` and `setup:linter`.
A required member of the `base` pack. `bootstrap remove tool <tool…>` refuses
it.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | `node`  |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

Always. The `base` pack adds this tool, and the sync writes an empty
`[tools.linter]` table for it. The table gives values.

## What the tool ships

| Path in the repository                       | Kind      | Use                             |
| -------------------------------------------- | --------- | ------------------------------- |
| `.config/mise/tasks/code/lint/05-linter`     | Synced    | The task `code:lint:05-linter`. |
| `.config/mise/tasks/setup/linter`            | Synced    | The task `setup:linter`.        |
| `.config/mise/conf.d/linter/mise.<env>.toml` | Generated | The mise config of this tool.   |

The `mise` tool ships the orchestrator `code:lint`.

## mise tools

| mise tool                | Version  | Environments | Note                                                                        |
| ------------------------ | -------- | ------------ | --------------------------------------------------------------------------- |
| `npm:@virajp.dev/linter` | `latest` | dev, ci      | Linter. Needs node. `minimum_release_age = "0h"` and `allow_low_downloads`. |
| `shellcheck`             | `latest` | dev, ci      | Shell linter                                                                |

## Tasks

| Task                  | Use                                                                                                                                                                                                        |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `code:lint:05-linter` | Run the linter on the given files, else on the whole tree, then ShellCheck on the shell files. `--fix` applies the fixes. Without the linter, it prints nothing. ShellCheck skips zsh files and templates. |
| `setup:linter`        | Write the linter config files, if they are missing.                                                                                                                                                        |

`code:lint` of the `mise` tool runs `code:lint:05-linter` first, then the
other `code:lint:<name>` tasks.

## Scripts

None.

## Contributions

| Tool         | Shared file                      | Items                                                                              |
| ------------ | -------------------------------- | ---------------------------------------------------------------------------------- |
| `pre-commit` | `.config/pre-commit-config.yaml` | The hook block `lint`, which calls `code:lint --fix` with the files of the commit. |
| `vscode`     | `.vscode/extensions.json`        | `dbaeumer.vscode-eslint`, `foxundermoon.shell-format`.                             |

## Values

Only `env`. No questions.

## Files the repository supplies

None.

## Adapt to the repository

1. Move each linter step of the repository to a task `code:lint:<name>`.

## Migration

- Remove `npm:@virajp.dev/linter`, the old `npm:@askviraj/linter` and
  `shellcheck` from the mise files of the repository.
- Replace the old package name `@askviraj/linter` with `@virajp.dev/linter`
  in `eslint.config.mjs` and `.config/linter.yaml` (M21 of `migrate.md`).
