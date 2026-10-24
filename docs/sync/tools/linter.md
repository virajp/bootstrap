# The `linter` tool

The house linter and ShellCheck, `code:lint` and `setup:linter`.

| Property     | Value                                               |
| ------------ | --------------------------------------------------- |
| Root only    | No                                                  |
| Depends      | None. Its mise tool adds `node` (see "mise tools"). |
| Default envs | dev, ci                                             |
| Seeds        | None                                                |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not list it in the values
file.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/tasks/code/lint/_default`      | Template  | The task `code:lint`.         |
| `.config/mise/tasks/setup/linter`            | Synced    | The task `setup:linter`.      |
| `.config/mise/conf.d/linter/mise.<env>.toml` | Generated | The mise config of this tool. |

VS Code extensions: `dbaeumer.vscode-eslint`, `foxundermoon.shell-format`.

## mise tools

| mise tool                | Version  | Environments | Note                |
| ------------------------ | -------- | ------------ | ------------------- |
| `npm:@virajp.dev/linter` | `latest` | dev, ci      | Linter. Needs node. |
| `shellcheck`             | `latest` | dev, ci      | Shell linter        |

## Tasks

| Task           | Use                                                                                                                                                                                                                                                                                               |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `code:lint`    | Run the linters, then each `code:lint:<name>`, on the `[files]` arguments, else on the whole tree. `--fix` applies fixes, and `--debug` prints debug logs. With git submodules, `--all` and `--<submodule>` also run `code:lint` in those submodules. A folder member of a monorepo gets no flag. |
| `setup:linter` | Write the linter config files, if they are missing.                                                                                                                                                                                                                                               |

## Pre-commit hooks

- The local hook `lint`, which calls `code:lint --fix` with the files of the
  commit.

## Files the repository supplies

None.

## Adapt to the repository

1. Move each repository linter step to a task `code:lint:<name>`.

## Migration

- Remove `npm:@virajp.dev/linter`, the old `npm:@askviraj/linter` and
  `shellcheck` from the repository-owned mise files.
- Replace the old package name `@askviraj/linter` with `@virajp.dev/linter`
  in `eslint.config.mjs` and `.config/linter.yaml` (M21 of `migrate.md`).
