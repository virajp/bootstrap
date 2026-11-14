# The `grype` tool

grype: the dependency scanner, and `code:sec:20-grype`. A
required member of the `base` pack. `bootstrap remove tool <tool…>` refuses it.

| Property     | Value                |
| ------------ | -------------------- |
| Root only    | No                   |
| Depends      | `_security_base`     |
| Parent       | None                 |
| Default envs | dev, ci              |
| Seeds        | `.config/grype.yaml` |

## When the tool is relevant

Always. The `base` pack adds this tool, and the sync writes an empty
`[tools.grype]` table for it. The table gives values.

## What the tool ships

| Path in the repository                      | Kind      | Use                           |
| ------------------------------------------- | --------- | ----------------------------- |
| `.config/grype.yaml`                        | Seed      | The config of grype.          |
| `.config/mise/tasks/code/sec/20-grype`      | Synced    | The task `code:sec:20-grype`. |
| `.config/mise/conf.d/grype/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments | Note                  |
| --------- | -------- | ------------ | --------------------- |
| `grype`   | `latest` | dev, ci      | Vulnerability scanner |

## Tasks

| Task                | Use                                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------------------- |
| `code:sec:20-grype` | Scan the dependencies of the tree, without the git submodules. `--staged` does nothing. `--debug` adds `-vv`. |

`code:sec` of the `mise` tool runs this task.

## Scripts

None.

## Contributions

None. The `sec` hook of `_security_base` runs `code:sec`, which runs this
task.

## Values

Only `env`. No questions.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

- Remove `grype` from the mise files of the repository.
