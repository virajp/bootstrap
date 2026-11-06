# The `crane` tool

crane, which operates on container images and registries. It is a member of the
`container-build` pack.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- A script, a task or a CI job of the repository calls `crane`.

## What the tool ships

| Path in the repository                      | Kind      | Use                           |
| ------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/crane/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `crane`   | `latest` | dev, ci      |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

- Remove `crane` from the repository-owned mise files.
