# The `pre-commit` tool

pre-commit: the rendered pre-commit config, its setup task and
`code:precommit`.

| Property     | Value                                                 |
| ------------ | ----------------------------------------------------- |
| Root only    | No                                                    |
| Depends      | None. Its mise tool adds `python` (see "mise tools"). |
| Default envs | dev                                                   |
| Seeds        | None                                                  |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not select it in the values
file. A `[tools.pre-commit]` table gives values.

## What the tool ships

| Path in the repository                         | Kind      | Use                                                                                                                                  |
| ---------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `.config/pre-commit-config.yaml`               | Template  | The hooks of the selected tools, then `.config/repo-sync/pre-commit.yaml`. The args of `no-commit-to-branch` follow `direct_commit`. |
| `.config/mise/tasks/setup/precommit`           | Synced    | The task `setup:precommit`.                                                                                                          |
| `.config/mise/tasks/code/precommit`            | Synced    | The task `code:precommit`.                                                                                                           |
| `.config/mise/conf.d/pre-commit/mise.dev.toml` | Generated | The mise config of this tool.                                                                                                        |

## mise tools

| mise tool    | Version  | Environments | Note                 |
| ------------ | -------- | ------------ | -------------------- |
| `pre-commit` | `latest` | dev          | Hooks. Needs python. |

The tool also sets:

- `PRE_COMMIT_HOME = "$HOME/.cache/pre-commit"`.
- The shell alias `precommit`.

## Tasks

| Task              | Use                                               |
| ----------------- | ------------------------------------------------- |
| `setup:precommit` | Install the hooks. `--upgrade` also updates them. |
| `code:precommit`  | Run the hooks on all files.                       |

## Pre-commit hooks

`pre-commit/pre-commit-hooks`: large files, case conflicts, Windows names,
shebangs, merge conflicts, symlinks, JSON, TOML and YAML syntax, private keys,
end of file, trailing whitespace and line endings.

The sync runs `pre-commit autoupdate` on the rendered config, so each `rev`
becomes the newest tag, not the `rev` of the `tools.toml`. If pre-commit is not
available or the autoupdate fails, the sync stops with code 5.
`REPO_SYNC_NO_AUTOUPDATE=1` skips the autoupdate.

Before the render, `repo-sync` changes the `no-commit-to-branch` hook of each
entry. It removes each `"--branch", "<branch>"` pair of a core branch (`main`,
`develop`) that does not exist, or whose `[branch.<branch>]` table has
`direct_commit = true`. A hook with no pair left goes, and an entry with no
hook left goes too. Thus a repository with one core branch gets only
`--branch <its branch>`. `repo-sync` validates `direct_commit` first: see
`repo-sync.md`.

## Excludes

The rendered config has one top-level `exclude`. No hook reads a path that
matches it. The sync merges the items of the `exclude` list from three sources,
in this order, and the template joins them with `|`:

1. `[defaults]` of this tool: the mise lock files.
2. `[contributes.pre-commit]` of each selected tool.
3. `[tools.pre-commit]` of the values file.

A later source adds to the list. It does not replace the earlier items, and a
repeated item stays once. Each item is a Python regular expression.

```toml
[tools.pre-commit]
exclude = [ "(^|/)vendor/", '\.snap$' ]
```

Do not use an empty item. An empty pattern matches each path, so no hook reads
a file.

## Files the repository supplies

| File                                | Used by                                                                                                                             |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `.config/repo-sync/pre-commit.yaml` | `.config/pre-commit-config.yaml`. Optional: its own hooks. A `no-commit-to-branch` hook here changes as the hook of the `git` tool. |

## Adapt to the repository

1. Move each repository hook that no tool ships to
   `.config/repo-sync/pre-commit.yaml`.
2. Move each repository `exclude` pattern to the `exclude` list of
   `[tools.pre-commit]`.

## Migration

- Remove `pre-commit` and `PRE_COMMIT_HOME` from the repository-owned mise
  files.
- The general item M12 of `migrate.md` is about the files of this tool.
