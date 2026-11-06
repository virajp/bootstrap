# The `cosign` tool

cosign, which signs and verifies container images. It is a member of the
`container-build` pack.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- A script, a task or a CI job of the repository calls `cosign`.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/cosign/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `cosign`  | `latest` | dev, ci      |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

- Remove `cosign` from the repository-owned mise files.
