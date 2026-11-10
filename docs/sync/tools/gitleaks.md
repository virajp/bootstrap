# The `gitleaks` tool

gitleaks: the secret scanner, and `code:sec:10-gitleaks`. A member of the
`security` pack, with `default = true`. `bootstrap remove` refuses it, because
only a pack selects it.

| Property     | Value                   |
| ------------ | ----------------------- |
| Root only    | No                      |
| Depends      | `_security_base`        |
| Parent       | None                    |
| Default envs | dev, ci                 |
| Seeds        | `.config/gitleaks.toml` |

## When the tool is relevant

Always. The `base` pack adds this tool, through the `security` pack. Do not
select it in the values file. A `[tools.gitleaks]` table gives values.

## What the tool ships

| Path in the repository                         | Kind           | Use                              |
| ---------------------------------------------- | -------------- | -------------------------------- |
| `.config/gitleaks.toml`                        | Seed, template | The gitleaks rules.              |
| `.config/mise/tasks/code/sec/10-gitleaks`      | Synced         | The task `code:sec:10-gitleaks`. |
| `.config/mise/conf.d/gitleaks/mise.<env>.toml` | Generated      | The mise config of this tool.    |

## mise tools

| mise tool  | Version  | Environments | Note           |
| ---------- | -------- | ------------ | -------------- |
| `gitleaks` | `latest` | dev, ci      | Secret scanner |

## Tasks

| Task                   | Use                                                                                                              |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `code:sec:10-gitleaks` | Scan the files that git knows for secrets. `--staged` scans only the staged changes. `--debug` adds `--verbose`. |

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

1. Examine an existing `.config/gitleaks.toml`. A seed keeps an existing
   file. Without `[extend] useDefault = true`, gitleaks uses only the rules of
   the file. Tell the user.

## Migration

- Remove `gitleaks` from the mise files of the repository.
- Remove an old `gitleaks` hook (M12 of `migrate.md`).
