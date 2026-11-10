# The `_security_base` tool

The common tool of the `security` pack: the `code:sec` orchestrator and the
`sec` hook.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

Always. The `base` pack always adds this tool: each scanner of the `security`
pack has it in its `depends`. Do not select it in the values file. A
`[tools._security_base]` table gives values.

## What the tool ships

| Path in the repository                 | Kind   | Use                                            |
| -------------------------------------- | ------ | ---------------------------------------------- |
| `.config/mise/tasks/code/sec/_default` | Synced | The task `code:sec`. The `mise` tool ships it. |

## mise tools

None.

## Tasks

| Task       | Use                                                                                                                                                                                          |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `code:sec` | Run each `code:sec:<scanner>` task, and pass on `--staged` and `--debug` when they are given. `--staged` scans only the staged changes for secrets, skips grype, and prints no success line. |

`code:sec` stops with an error if one of `code:sec:10-gitleaks`,
`code:sec:20-grype` and `code:sec:30-osv-scanner` is missing or not executable.
Then
run `mise run init`. A repository can add its own `code:sec:<name>` task, and
`code:sec` runs it too.

## Pre-commit hooks

- The local hook `sec`, which calls `code:sec --staged`.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

None.
