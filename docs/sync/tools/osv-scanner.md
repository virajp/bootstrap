# The `osv-scanner` tool

osv-scanner: the dependency scanner, and `code:sec:osv-scanner`.

| Property     | Value                      |
| ------------ | -------------------------- |
| Root only    | No                         |
| Depends      | None                       |
| Default envs | dev, ci                    |
| Seeds        | `.config/osv-scanner.toml` |

## When the tool is relevant

Always. The `base` pack always adds this tool, through the `security` pack.
Do not list it in the values file.

## What the tool ships

| Path in the repository                            | Kind      | Use                              |
| ------------------------------------------------- | --------- | -------------------------------- |
| `.config/osv-scanner.toml`                        | Seed      | The config of osv-scanner.       |
| `.config/mise/tasks/code/sec/osv-scanner`         | Synced    | The task `code:sec:osv-scanner`. |
| `.config/mise/conf.d/osv-scanner/mise.<env>.toml` | Generated | The mise config of this tool.    |

## mise tools

| mise tool     | Version  | Environments | Note                  |
| ------------- | -------- | ------------ | --------------------- |
| `osv-scanner` | `latest` | dev, ci      | Vulnerability scanner |

## Tasks

| Task                   | Use                                                                                                                               |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `code:sec:osv-scanner` | Scan the dependencies of the tree, without the git submodules. `--staged` does nothing. `--debug` raises the verbosity to `info`. |

`code:sec` runs this task.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

- Remove `osv-scanner` from the repository-owned mise files.
