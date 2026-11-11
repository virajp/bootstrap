# The `gcloud` tool

The Google Cloud CLI, for a repository that deploys to Google Cloud.

| Property     | Value    |
| ------------ | -------- |
| Root only    | No       |
| Depends      | `python` |
| Parent       | None     |
| Default envs | dev, ci  |
| Seeds        | None     |

## When the tool is relevant

- The repository deploys to Google Cloud.
- A script, a task or a CI workflow calls `gcloud`.
- The repository selects `pulumi` with the Google Cloud provider.

`bootstrap detect` runs the detect script of the tool (see "Scripts"). It
finds `.gcloudignore` and `cloudbuild.yaml` at the root. It does not find
`app.yaml`, because many tools use that name.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/gcloud/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `gcloud`  | `latest` | dev, ci      |

The tool sets `CLOUDSDK_PYTHON` to the Python 3.14 of mise. gcloud needs this
Python minor version. Without it, gcloud uses the system `python3` and fails.
The `python` tool pins the same version.

## Tasks

None.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                                 |
| ---------------- | ------ | ------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a `.gcloudignore` or a `cloudbuild.yaml`, else `none`. |

## Contributions

None.

## Values

Only `env`. No questions. See
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository sets its own `CLOUDSDK_PYTHON`, tell the user. This tool
   sets it.

## Migration

- Remove `gcloud` from the repository-owned mise files.
- Delete `CLOUDSDK_PYTHON` from `.config/mise.toml` (M5 of `migrate.md`).
