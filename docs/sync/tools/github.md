# The `github` tool

GitHub: the `gh` CLI, actionlint, and the issue and pull request templates.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The `origin` remote of the repository is on GitHub.
- The repository has `.github/`.

`bootstrap detect` runs the detect script of the tool (see "Scripts").

Select `github` or `gitlab`, not both. The `pr` merge model needs the CLI of
the forge.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.github/ISSUE_TEMPLATE/bug_report.yml`      | Synced    | The bug report form.          |
| `.github/ISSUE_TEMPLATE/feature_request.yml` | Synced    | The feature request form.     |
| `.github/ISSUE_TEMPLATE/config.yml`          | Template  | The issue template config.    |
| `.github/pull_request_template.md`           | Synced    | The pull request template.    |
| `.config/mise/conf.d/github/mise.<env>.toml` | Generated | The mise config of this tool. |

`config.yml` has the contact links only when the values file has `repo_url`.

## mise tools

| mise tool    | Version  | Environments | Note                                            |
| ------------ | -------- | ------------ | ----------------------------------------------- |
| `gh`         | `latest` | dev, ci      | Issues, pull requests, and the `pr` merge model |
| `actionlint` | `latest` | dev, ci      | The workflow linter                             |

## Tasks

None. The `code:merge` tasks of the `git` tool use `gh` under the `pr` merge
model.

## Scripts

| Script           | Kind   | Use                                            |
| ---------------- | ------ | ---------------------------------------------- |
| `scripts/detect` | detect | `match` with a `.github/` folder, else `none`. |

The sync runs the scripts from the clone of `bootstrap`, and never copies
them. See [`docs/design/tools.md`](../../design/tools.md).

## Contributions

| Tool         | Shared file                      | Items                                                                                                        |
| ------------ | -------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `pre-commit` | `.config/pre-commit-config.yaml` | The local hook block `actionlint`. It runs `mise x -- actionlint` on `.github/workflows/*.yml` and `*.yaml`. |
| `vscode`     | `.vscode/extensions.json`        | `github.vscode-github-actions`.                                                                              |

## Values

Only `env`. No questions. The template reads the top-level key `repo_url`.
See [`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Set `repo_url` in the values file to the GitHub web URL of the repository
   ("Step 6: Write the other values and sync again" of
   [`index.md`](../index.md)).

## Migration

`github` replaces these files, if they exist:
`.github/ISSUE_TEMPLATE/bug_report.yml`, `feature_request.yml`, `config.yml`,
and `.github/pull_request_template.md`.

1. Before the first sync, compare them with the shared files. Tell the user
   about each repository-specific change, for example an extra contact link.
2. Remove `gh` and `actionlint` from the repository-owned mise files.
