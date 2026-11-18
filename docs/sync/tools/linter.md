# The `linter` tool

The house linter and ShellCheck, `code:lint:05-linter`,
`code:lint:06-shellcheck` and `setup:linter`.
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

| Path in the repository                       | Kind      | Use                                 |
| -------------------------------------------- | --------- | ----------------------------------- |
| `.config/mise/tasks/code/lint/05-linter`     | Synced    | The task `code:lint:05-linter`.     |
| `.config/mise/tasks/code/lint/06-shellcheck` | Synced    | The task `code:lint:06-shellcheck`. |
| `.config/mise/tasks/setup/linter`            | Synced    | The task `setup:linter`.            |
| `.config/mise/conf.d/linter/mise.<env>.toml` | Generated | The mise config of this tool.       |

The `mise` tool ships the orchestrator `code:lint`.

## mise tools

| mise tool                | Version  | Environments | Note                                                                        |
| ------------------------ | -------- | ------------ | --------------------------------------------------------------------------- |
| `npm:@virajp.dev/linter` | `latest` | dev, ci      | Linter. Needs node. `minimum_release_age = "0h"` and `allow_low_downloads`. |
| `shellcheck`             | `latest` | dev, ci      | Shell linter                                                                |

## Tasks

| Task                      | Use                                                                                                                                                                                                                                                                                                     |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `code:lint:05-linter`     | Run the linter on the files of `gate_files`: the given files, else with `--all` every file, else the uncommitted files. Its `#FILES` header names the file types of `@virajp.dev/linter` 2.1.0 and excludes its built-in ignore list. `--fix` applies the fixes. Without the linter, it prints nothing. |
| `code:lint:06-shellcheck` | Run ShellCheck on the shell files of `gate_files`: `*.sh`, `*.bash` and each file whose `#!` names `sh`, `bash`, `dash` or `ksh`, but not a template (`*.tmpl`) or a file whose `#!` names another shell, for example zsh. Only warnings and errors fail. Without ShellCheck, it prints nothing.        |
| `setup:linter`            | Write the linter config files, if they are missing.                                                                                                                                                                                                                                                     |

`code:lint` of the `mise` tool runs `code:lint:05-linter` first, then
`code:lint:06-shellcheck`, then the other `code:lint:NN-<name>` tasks. The
file sets of the gates are in
[`mise.md`](mise.md#the-file-sets-of-the-gates).

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

1. Move each linter step of the repository to a task `code:lint:NN-<name>`.

## Migration

- Remove `npm:@virajp.dev/linter`, the old `npm:@askviraj/linter` and
  `shellcheck` from the mise files of the repository.
- Replace the old package name `@askviraj/linter` with `@virajp.dev/linter`
  in `eslint.config.mjs` and `.config/linter.yaml` (M21 of `migrate.md`).
- ShellCheck is now its own gate, `code:lint:06-shellcheck`. Move the
  `ignores` of `.config/linter.yaml` into `code_files.excludes` of
  `[tools.mise]` (M45 of `migrate.md`).
