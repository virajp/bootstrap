# The `gitleaks` tool

gitleaks: the secret scanner, and `code:sec:gitleaks`.

| Property     | Value                   |
| ------------ | ----------------------- |
| Root only    | No                      |
| Depends      | None                    |
| Default envs | dev, ci                 |
| Seeds        | `.config/gitleaks.toml` |

## When the tool is relevant

Always. The `base` pack always adds this tool, through the `security` pack.
Do not list it in the values file.

## What the tool ships

| Path in the repository                         | Kind           | Use                           |
| ---------------------------------------------- | -------------- | ----------------------------- |
| `.config/gitleaks.toml`                        | Seed, template | The gitleaks rules.           |
| `.config/mise/tasks/code/sec/gitleaks`         | Synced         | The task `code:sec:gitleaks`. |
| `.config/mise/conf.d/gitleaks/mise.<env>.toml` | Generated      | The mise config of this tool. |

## mise tools

| mise tool  | Version  | Environments | Note           |
| ---------- | -------- | ------------ | -------------- |
| `gitleaks` | `latest` | dev, ci      | Secret scanner |

## Tasks

| Task                | Use                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------ |
| `code:sec:gitleaks` | Scan the files that git knows for secrets. `--staged` scans only the staged changes. |

`code:sec` runs this task.

## Pre-commit hooks

None. The `sec` hook of `_security_base` runs this task.

## Files the repository supplies

None.

## Adapt to the repository

1. Examine an existing `.config/gitleaks.toml`. A seed keeps an existing
   file. Without `[extend] useDefault = true`, gitleaks uses only the rules of
   the file. Tell the user.

## Migration

- Remove `gitleaks` from the repository-owned mise files.
- Remove an old `gitleaks` hook (M12 of `migrate.md`).
