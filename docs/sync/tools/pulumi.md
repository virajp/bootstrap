# The `pulumi` tool

Pulumi. A repository that deploys to Google Cloud also selects `gcloud`.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

- The repository has `Pulumi.yaml`.

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

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If the Pulumi program uses the Google Cloud provider, also select
   `gcloud`.
2. If CI runs Pulumi, tell the user. This tool installs Pulumi in dev only.

## Migration

- Remove `pulumi` from the repository-owned mise files.
