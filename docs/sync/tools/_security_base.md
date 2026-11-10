# The `_security_base` tool

The common tool of the `security` pack: the `sec` hook. Each scanner of the
`security` pack has it in its `depends`, so the `base` pack adds it.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

Always. A scanner of the `security` pack adds this tool. Do not select it in
the values file. A `[tools._security_base]` table gives values.

## What the tool ships

None. The `mise` tool ships the orchestrator `code:sec`
(`.config/mise/tasks/code/sec/_default`).

## mise tools

None.

## Tasks

None. The task `code:sec` of the `mise` tool runs each `code:sec:<scanner>`
task, and passes on `--staged` and `--debug` when they are given. `--staged`
scans only the staged changes for secrets, skips grype, and prints no success
line.

`code:sec` stops with an error if one of `code:sec:10-gitleaks`,
`code:sec:20-grype` and `code:sec:30-osv-scanner` is missing or not
executable. Then run `mise run init`. A repository can add its own
`code:sec:<name>` task, and `code:sec` runs it too.

## Scripts

None.

## Contributions

| Tool         | Shared file                      | Items                                                                 |
| ------------ | -------------------------------- | --------------------------------------------------------------------- |
| `pre-commit` | `.config/pre-commit-config.yaml` | The hook block `sec`, which calls `code:sec --staged` at each commit. |

## Values

Only `env`. No questions.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

None.
