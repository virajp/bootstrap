# The `pre-commit` tool

pre-commit: the rendered pre-commit config, its setup task and
`code:precommit`.

| Property     | Value                                                 |
| ------------ | ----------------------------------------------------- |
| Root only    | No                                                    |
| Depends      | None. Its mise tool adds `python` (see "mise tools"). |
| Default envs | dev                                                   |
| Seeds        | None                                                  |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not list it in the values
file.

## What the tool ships

| Path in the repository                         | Kind      | Use                                                                        |
| ---------------------------------------------- | --------- | -------------------------------------------------------------------------- |
| `.config/pre-commit-config.yaml`               | Template  | The hooks of the selected tools, then `.config/repo-sync/pre-commit.yaml`. |
| `.config/mise/tasks/setup/precommit`           | Synced    | The task `setup:precommit`.                                                |
| `.config/mise/tasks/code/precommit`            | Synced    | The task `code:precommit`.                                                 |
| `.config/mise/conf.d/pre-commit/mise.dev.toml` | Generated | The mise config of this tool.                                              |

## mise tools

| mise tool    | Version  | Environments | Note                 |
| ------------ | -------- | ------------ | -------------------- |
| `pre-commit` | `latest` | dev          | Hooks. Needs python. |

The tool also sets:

- `PRE_COMMIT_HOME = "$HOME/.cache/pre-commit"`.
- The shell alias `precommit`.

## Tasks

| Task              | Use                                               |
| ----------------- | ------------------------------------------------- |
| `setup:precommit` | Install the hooks. `--upgrade` also updates them. |
| `code:precommit`  | Run the hooks on all files.                       |

## Pre-commit hooks

`pre-commit/pre-commit-hooks`: large files, case conflicts, Windows names,
shebangs, merge conflicts, symlinks, JSON, TOML and YAML syntax, private keys,
end of file, trailing whitespace and line endings.

The sync runs `pre-commit autoupdate` on the rendered config, so each `rev`
becomes the newest tag, not the `rev` of the `tools.toml`. If pre-commit is not
available or the autoupdate fails, the sync stops with code 5.
`REPO_SYNC_NO_AUTOUPDATE=1` skips the autoupdate.

## Files the repository supplies

| File                                | Used by                                                    |
| ----------------------------------- | ---------------------------------------------------------- |
| `.config/repo-sync/pre-commit.yaml` | `.config/pre-commit-config.yaml`. Optional: its own hooks. |

## Adapt to the repository

1. Move each repository hook that no tool ships to
   `.config/repo-sync/pre-commit.yaml`.

## Migration

- Remove `pre-commit` and `PRE_COMMIT_HOME` from the repository-owned mise
  files.
- The general item M12 of `migrate.md` is about the files of this tool.
