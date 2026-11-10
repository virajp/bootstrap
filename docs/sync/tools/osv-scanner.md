# The `osv-scanner` tool

osv-scanner: the dependency scanner, and `code:sec:30-osv-scanner`. A member of
the
`security` pack, with `default = true`. `bootstrap remove` refuses it, because
only a pack selects it.

| Property     | Value                      |
| ------------ | -------------------------- |
| Root only    | No                         |
| Depends      | `_security_base`           |
| Parent       | None                       |
| Default envs | dev, ci                    |
| Seeds        | `.config/osv-scanner.toml` |

## When the tool is relevant

Always. The `base` pack adds this tool, through the `security` pack. Do not
select it in the values file. A `[tools.osv-scanner]` table gives values.

## What the tool ships

| Path in the repository                            | Kind      | Use                                 |
| ------------------------------------------------- | --------- | ----------------------------------- |
| `.config/osv-scanner.toml`                        | Seed      | The config of osv-scanner.          |
| `.config/mise/tasks/code/sec/30-osv-scanner`      | Synced    | The task `code:sec:30-osv-scanner`. |
| `.config/mise/conf.d/osv-scanner/mise.<env>.toml` | Generated | The mise config of this tool.       |

## mise tools

| mise tool     | Version  | Environments | Note                  |
| ------------- | -------- | ------------ | --------------------- |
| `osv-scanner` | `latest` | dev, ci      | Vulnerability scanner |

## Tasks

| Task                      | Use                                                                                                                               |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `code:sec:30-osv-scanner` | Scan the dependencies of the tree, without the git submodules. `--staged` does nothing. `--debug` raises the verbosity to `info`. |

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

- Remove `osv-scanner` from the mise files of the repository.
