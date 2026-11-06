# The `apko` tool

apko, which builds OCI images from apk packages. It is a member of the
`container-build` pack.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has apko image configs.
- A script or a task of the repository calls `apko`.

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/conf.d/apko/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `apko`    | `latest` | dev, ci      |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

| File               | Used by |
| ------------------ | ------- |
| The image configs. | `apko`  |

## Adapt to the repository

None.

## Migration

- Remove `apko` from the repository-owned mise files.
