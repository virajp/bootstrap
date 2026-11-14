# The `dprint` tool

dprint: the formatter of the sync and of `code:format`, the shared
`dprint.jsonc` and `setup:dprint`. A required member of the `base` pack.
`bootstrap remove tool <tool…>` refuses it.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

Always. The `base` pack adds this tool, and the sync writes an empty
`[tools.dprint]` table for it. The table gives values.

## What the tool ships

| Path in the repository                       | Kind      | Use                               |
| -------------------------------------------- | --------- | --------------------------------- |
| `dprint.jsonc`                               | Shared    | The dprint config. See below.     |
| `.config/mise/tasks/code/format/05-dprint`   | Synced    | The task `code:format:05-dprint`. |
| `.config/mise/tasks/setup/dprint`            | Synced    | The task `setup:dprint`.          |
| `.config/mise/conf.d/dprint/mise.<env>.toml` | Generated | The mise config of this tool.     |

`dprint.jsonc` is a shared file. The sync renders the spec
`files/dprint.jsonc.merge.toml` with the values of this tool, then merges it
into the file item by item:

- It removes an old `extends` key. The file has all its settings now.
- The top-level settings, and the keys of each plugin block, for example
  `markdown` or `toml`. A key that the file has keeps its value.
- Each item of `excludes`.
- Each item of `plugins`, found by its name with any version. Thus the
  version that `setup:dprint` writes stays.

The repository keeps its own keys, excludes and plugins directly in the file,
and the sync keeps them. To remove or skip an item of `bootstrap`, write an
override in the values file:

```toml
[tools.dprint.files."dprint.jsonc".markdown]
emphasisKind = { "$skip" = true }

[tools.dprint.files."dprint.jsonc".excludes]
"**/*.svg" = { "$mode" = "remove" }
```

The merge rules are in [`docs/design/tools.md`](../../design/tools.md).

### The formatter of the sync

The `[formatter]` table of this tool makes dprint the formatter of the sync.
The sync formats each rendered file with
`dprint fmt --no-gitignore --allow-no-files --stdin-files` before it compares
the file. It uses the first dprint config of the repository that it finds,
for example `dprint.jsonc`. See
[`docs/design/tools.md`](../../design/tools.md).

## mise tools

| mise tool | Version  | Environments | Note      |
| --------- | -------- | ------------ | --------- |
| `dprint`  | `latest` | dev, ci      | Formatter |

## Tasks

| Task                    | Use                                                                                                                                          |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `code:format:05-dprint` | Check the files with dprint. `--fix` formats them. It uses the given files, else the files of its config. Without dprint, it prints nothing. |
| `setup:dprint`          | Update the plugin versions in `dprint.jsonc` with `dprint config update`.                                                                    |

`code:format` of the `mise` tool runs `code:format:05-dprint` first, then the
other `code:format:<name>` tasks.

## Scripts

None.

## Contributions

| Tool         | Shared file                      | Items                                                                                  |
| ------------ | -------------------------------- | -------------------------------------------------------------------------------------- |
| `pre-commit` | `.config/pre-commit-config.yaml` | The hook block `format`, which calls `code:format --fix` with the files of the commit. |
| `vscode`     | `.vscode/extensions.json`        | `dprint.dprint`.                                                                       |

Other tools contribute blocks, plugins and excludes to `dprint.jsonc` in
`[contributes.dprint]`. Each tool file lists them.

## Values

The keys of `[tools.dprint]` in `.config/bootstrap.toml`:

| Key                                   | Use                                                                                          |
| ------------------------------------- | -------------------------------------------------------------------------------------------- |
| `env`                                 | The environments of the tool.                                                                |
| `settings`                            | The top-level settings. The default is `newLineKind = "lf"`.                                 |
| `blocks.<plugin>`                     | The keys of a plugin block. The defaults are `toml`, `json`, `markdown`, `yaml` and `shuck`. |
| `excludes`                            | More excludes. A list adds its items.                                                        |
| `plugins`                             | More plugins. A list adds its items.                                                         |
| `[tools.dprint.files."dprint.jsonc"]` | The overrides of the items of `bootstrap` (see "What the tool ships").                       |

No questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Move each formatter step of the repository to a task `code:format:<name>`.
2. Keep the own excludes of the repository in `dprint.jsonc`, for example the
   submodules.

## Migration

- Remove `dprint` from the mise files of the repository.
- Remove shfmt. dprint formats shell scripts now.
- Remove taplo. dprint formats TOML now. Also remove the `[toml]` and
  `evenBetterToml.*` keys from `.vscode/settings.json`.
- M34 of `migrate.md`: the sync now writes `dprint.jsonc`. It removes the old
  `extends` of the global config `~/.config/dprint/dprint.jsonc`, and it keeps
  the own excludes of the file.
