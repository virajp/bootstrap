# The `apko` tool

apko, which builds OCI images from apk packages.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has apko image configs.
- A script or a task of the repository calls `apko`.

`bootstrap detect` runs the detect script of the tool (see "Scripts"). It
finds an apko config by its top-level keys `contents:` and `entrypoint:`,
because apko has no fixed file name. It does not find a config without
`entrypoint:`.

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

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                                                                                                                  |
| ---------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a `*.yaml` or `*.yml` file, at most two folders below the root, that has the top-level keys `contents:` and `entrypoint:`, else `none`. |

## Contributions

None.

## Values

Only `env`. No questions. See
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File               | Used by |
| ------------------ | ------- |
| The image configs. | `apko`  |

## Adapt to the repository

None.

## Migration

- Remove `apko` from the repository-owned mise files.
