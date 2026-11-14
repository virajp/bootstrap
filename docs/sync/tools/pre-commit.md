# The `pre-commit` tool

pre-commit: the shared pre-commit config, its setup task and
`code:precommit`. The `base` pack adds it with `required = true`, so
`bootstrap remove tool <tool…>` refuses it.

| Property     | Value    |
| ------------ | -------- |
| Root only    | No       |
| Depends      | `python` |
| Parent       | None     |
| Default envs | dev      |
| Seeds        | None     |

## When the tool is relevant

Always. The `base` pack adds this tool. Do not select it in the values file.
A `[tools.pre-commit]` table gives values.

## What the tool ships

| Path in the repository                         | Kind      | Use                                                          |
| ---------------------------------------------- | --------- | ------------------------------------------------------------ |
| `.config/pre-commit-config.yaml`               | Shared    | The hooks of this tool and of the selected tools. See below. |
| `.config/mise/tasks/setup/precommit`           | Synced    | The task `setup:precommit`.                                  |
| `.config/mise/tasks/code/precommit`            | Synced    | The task `code:precommit`.                                   |
| `.config/mise/conf.d/pre-commit/mise.dev.toml` | Generated | The mise config of this tool.                                |

`.config/pre-commit-config.yaml` is a shared file. The sync renders the spec
`files/.config/pre-commit-config.yaml.merge.toml` with the values of this
tool, then merges it into the file item by item:

- The keys `minimum_pre_commit_version`, `default_install_hook_types` and
  `default_stages`.
- The top-level `exclude`: the `exclude` items of the values, joined with `|`.
- One block in `repos` for each hook repository, found by its `repo` and its
  hook ids. The sync writes a found block again, but it keeps its `rev`.

The repository keeps its own hook blocks directly in the file, and the sync
keeps them. To remove or skip a block of `bootstrap`, write an override in the
values file, with the `id` of the block:

```toml
[tools.pre-commit.files.".config/pre-commit-config.yaml".repos]
"conventional-commits" = { "$mode" = "remove" }
```

The merge rules are in [`docs/design/tools.md`](../../design/tools.md).

pre-commit does not find a config below `.config/`. `setup:precommit`
installs the hooks with the path, so `git commit` needs no flag.
`pre-commit run` needs `--config .config/pre-commit-config.yaml`.

## mise tools

| mise tool    | Version  | Environments | Note                 |
| ------------ | -------- | ------------ | -------------------- |
| `pre-commit` | `latest` | dev          | Hooks. Needs python. |

The tool also sets:

- `PRE_COMMIT_HOME = "$HOME/.cache/pre-commit"`.
- The shell alias `precommit`.

## Tasks

| Task              | Use                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------- |
| `setup:precommit` | Install the hooks. `--upgrade` also runs `pre-commit autoupdate`, which updates each `rev`. |
| `code:precommit`  | Run the hooks on all files.                                                                 |

The sync keeps the `rev` of a found block. Thus the `rev` that
`setup:precommit --upgrade` writes stays at the next sync.

## Scripts

None.

## Contributions

None. Other tools contribute hook blocks to `.config/pre-commit-config.yaml`
in `[[contributes.pre-commit.repos]]`. Each tool file lists its blocks.

This tool has the block `pre-commit-hooks` of
`pre-commit/pre-commit-hooks` in its `[defaults]`: large files, case
conflicts, Windows names, shebangs, merge conflicts, symlinks, JSON, TOML and
YAML syntax, private keys, end of file, trailing whitespace and line endings.

## Values

The keys of `[tools.pre-commit]` in `.config/bootstrap.toml`:

| Key                                                         | Use                                                                                                                      |
| ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `env`                                                       | The environments of the tool.                                                                                            |
| `exclude`                                                   | More items of the top-level `exclude`, each a Python regular expression. A list adds its items to the default `exclude`. |
| `repos`                                                     | More hook blocks, each with an `id`. A list adds its items. A block directly in the file does the same.                  |
| `[tools.pre-commit.files.".config/pre-commit-config.yaml"]` | The overrides of the items of `bootstrap` (see "What the tool ships").                                                   |

```toml
[tools.pre-commit]
exclude = [ "(^|/)vendor/", '\.snap$' ]
```

Do not use an empty `exclude` item. An empty pattern matches each path, so no
hook reads a file.

No questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Keep each hook of the repository that no tool ships as a block in
   `.config/pre-commit-config.yaml`.
2. Move each `exclude` pattern of the repository to the `exclude` list of
   `[tools.pre-commit]`.

## Migration

- Remove `pre-commit` and `PRE_COMMIT_HOME` from the mise files of the
  repository.
- The general item M12 of `migrate.md` is about the files of this tool.
- M34 of `migrate.md`: `bootstrap init` moves the hooks of the old file of
  the repository hooks into `.config/pre-commit-config.yaml`, and deletes the
  old file.
