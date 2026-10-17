# The `gitlab` tool

GitLab: the `glab` CLI, and the issue and merge request templates.

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The `origin` remote of the repository is on GitLab.
- The repository has `.gitlab/` or `.gitlab-ci.yml`.

Select `github` or `gitlab`, not both. The `pr` merge model needs the CLI of
the forge.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.gitlab/issue_templates/Bug.md`             | Synced    | The bug report template.      |
| `.gitlab/issue_templates/Feature.md`         | Synced    | The feature request template. |
| `.gitlab/merge_request_templates/Default.md` | Synced    | The merge request template.   |
| `.config/mise/conf.d/gitlab/mise.<env>.toml` | Generated | The mise config of this tool. |

VS Code extension: `gitlab.gitlab-workflow`.

## mise tools

| mise tool | Version  | Environments | Note                                                             |
| --------- | -------- | ------------ | ---------------------------------------------------------------- |
| `glab`    | `latest` | dev, ci      | Issues, merge requests, `glab ci lint`, and the `pr` merge model |

## Tasks

None. The `code:merge` tasks of `base` use `glab` under the `pr` merge model.

## Pre-commit hooks

- `glab-ci-lint` on `.gitlab-ci.yml`. It asks the GitLab API, so it needs the
  network and `glab auth`.

## Files the repository supplies

None.

## Adapt to the repository

1. Set `repo_url` in the values file to the GitLab web URL of the repository.
2. Tell the user that the `glab-ci-lint` hook needs `glab auth login`.

## Migration

`gitlab` replaces these files, if they exist:
`.gitlab/issue_templates/Bug.md`, `Feature.md`, and
`.gitlab/merge_request_templates/Default.md`.

1. Before the first sync, compare them with the shared files. Tell the user
   about each repository-specific change.
2. Remove `glab` from the repository-owned mise files.
