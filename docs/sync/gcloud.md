# The `gcloud` tool

The Google Cloud CLI, for a repository that deploys to Google Cloud.

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The repository deploys to Google Cloud.
- A script, a task or a CI workflow calls `gcloud`.
- The repository selects `pulumi` with the Google Cloud provider.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/gcloud/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments | Note         |
| --------- | -------- | ------------ | ------------ |
| `gcloud`  | `latest` | dev, ci      | Needs python |

The tool sets `CLOUDSDK_PYTHON` to the Python 3.14 of mise. gcloud needs this
Python minor version. Without it, gcloud uses the system `python3` and fails.
The `python` tool pins the same version.

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository sets its own `CLOUDSDK_PYTHON`, tell the user. This tool
   sets it.

## Migration

- Remove `gcloud` from the repository-owned mise files.
- Delete `CLOUDSDK_PYTHON` from `.config/mise.toml` (M5 of `migrate.md`).
