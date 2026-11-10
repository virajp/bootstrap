# The `pulumi` tool

Pulumi. A repository that deploys to Google Cloud also selects `gcloud`.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Parent       | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

- The repository has `Pulumi.yaml` or `Pulumi.yml`.

`bootstrap detect` runs the detect script of the tool (see "Scripts").

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/conf.d/pulumi/mise.dev.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `pulumi`  | `latest` | dev          |

## Tasks

None.

## Scripts

| Script           | Kind   | Use                                                                  |
| ---------------- | ------ | -------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` with `Pulumi.yaml` or `Pulumi.yml` at the root, else `none`. |

The sync runs the scripts from the clone of `bootstrap`, and never copies them.
See [`docs/design/tools.md`](../../design/tools.md).

## Contributions

None.

## Values

Only `env`. No questions. See
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. If the Pulumi program uses the Google Cloud provider, also select
   `gcloud`.
2. If CI runs Pulumi, tell the user. The mise tool `pulumi` allows only dev.

## Migration

- Remove `pulumi` from the repository-owned mise files.
