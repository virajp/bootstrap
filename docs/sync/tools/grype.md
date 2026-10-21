# The `grype` tool

grype: the dependency scanner, and `code:sec:grype`.

| Property     | Value                |
| ------------ | -------------------- |
| Root only    | No                   |
| Depends      | None                 |
| Default envs | dev, ci              |
| Seeds        | `.config/grype.yaml` |

## When the tool is relevant

Always. The `base` pack always adds this tool, through the `security` pack.
Do not list it in the values file.

## What the tool ships

| Path in the repository                      | Kind      | Use                           |
| ------------------------------------------- | --------- | ----------------------------- |
| `.config/grype.yaml`                        | Seed      | The config of grype.          |
| `.config/mise/tasks/code/sec/grype`         | Synced    | The task `code:sec:grype`.    |
| `.config/mise/conf.d/grype/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments | Note                  |
| --------- | -------- | ------------ | --------------------- |
| `grype`   | `latest` | dev, ci      | Vulnerability scanner |

## Tasks

| Task             | Use                                                                                                           |
| ---------------- | ------------------------------------------------------------------------------------------------------------- |
| `code:sec:grype` | Scan the dependencies of the tree, without the git submodules. `--staged` does nothing. `--debug` adds `-vv`. |

`code:sec` runs this task.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

- Remove `grype` from the repository-owned mise files.
