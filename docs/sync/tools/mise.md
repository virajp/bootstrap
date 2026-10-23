# The `mise` tool

The shared mise config, the task library, and the `setup` and `code`
orchestrators.

| Property     | Value                 |
| ------------ | --------------------- |
| Root only    | No                    |
| Depends      | None                  |
| Default envs | dev, ci               |
| Seeds        | `zz-repo/mise.*.toml` |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not list it in the values
file.

## What the tool ships

| Path in the repository                          | Kind      | Use                                                                                                                                                               |
| ----------------------------------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.config/mise.toml`                             | Template  | The shared `[settings]`, and the `[env]` values from the values file.                                                                                             |
| `.config/mise/conf.d/mise/mise.<env>.toml`      | Generated | The mise config of this tool, one file for each environment.                                                                                                      |
| `.config/mise/conf.d/zz-repo/mise.<env>.toml`   | Seed      | The repository's own mise config and `[env]` values, one file for each environment. The folder name sorts last, so this config wins over the config of each tool. |
| `.config/mise/tasks/_scripts/`                  | Synced    | The libraries `helpers`, `checks` and `placeholder`.                                                                                                              |
| `.config/mise/tasks/setup/_default`             | Template  | The task `setup`.                                                                                                                                                 |
| `.config/mise/tasks/setup/worktree`             | Template  | The task `setup:worktree`.                                                                                                                                        |
| `.config/mise/tasks/setup/mise`                 | Synced    | The task `setup:mise`.                                                                                                                                            |
| `.config/mise/tasks/code/_default`, `worktrees` | Synced    | The tasks `code` and `code:worktrees`.                                                                                                                            |

VS Code extensions: `hverlin.mise-vscode`, `redhat.vscode-yaml`,
`tamasfe.even-better-toml`.

## mise tools

| mise tool | Version  | Environments | Note         |
| --------- | -------- | ------------ | ------------ |
| `jq`      | `latest` | dev, ci      | JSON queries |
| `yq`      | `latest` | dev, ci      | YAML queries |

The tool also sets:

- The shell alias `setup`. With git submodules, the sync also adds
  `setup-all` and `setup-<submodule>`.
- The hidden task `init`, which makes the task scripts executable.
- `os = "macos"` in `dev` and `os = "linux"` in `ci`.

## Tasks

| Task             | Use                                                                                                                                                                                                                                                          |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `setup`          | Install tools, dependencies and hooks, and run `setup:vscode` when the repository has it. `--upgrade` also upgrades them. With git submodules, `--all` and `--<submodule>` also run `setup` in those submodules. A folder member of a monorepo gets no flag. |
| `setup:mise`     | Reshim, doctor and install. `--upgrade` also upgrades.                                                                                                                                                                                                       |
| `setup:worktree` | Initialize a new git worktree. It is lighter than `setup`.                                                                                                                                                                                                   |
| `code`           | Run `code:format`, `code:lint` and `code:sec`. `--debug` passes on to each of them.                                                                                                                                                                          |
| `code:worktrees` | List the worktrees of the root repository and of each submodule.                                                                                                                                                                                             |

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. Write the values file (step 3 of `index.md`).
2. Move each repository shell function to `.config/mise/tasks/_scripts/local`.

## Migration

- Remove `jq` and `yq` from the repository-owned mise files.
- The general items M3, M5, M10, M14 and M17 of `migrate.md` are about the
  files of this tool.
