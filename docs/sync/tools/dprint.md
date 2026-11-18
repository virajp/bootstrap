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
- Each command of the `exec` plugin block, found by its `exts`. The sync writes
  the command of each selected tool again, adds a new one, and removes the
  command of a tool that went. A command that the repository adds for other
  `exts` stays.
- Each item of `excludes`: first the patterns of `code_files.excludes` of the
  `mise` tool, the one global exclude list of the gates, then the `excludes`
  of this tool. Each pattern has the gitignore form and gets the dprint form:
  a leading `/` is removed; a pattern `p` without `/` gives `**/p` and
  `**/p/**`; a pattern whose only `/` is trailing (`build/`) gives
  `**/build/**`; any other trailing `/` gets `**` after it.
- Each item of `plugins`, found by its name with any version. Thus the
  version that `setup:dprint` writes stays.

The repository keeps its own keys and plugins directly in the file, and the
sync keeps them. An own exclude also stays, but `code:format` never sees it:
put the excludes of the repository in `code_files.excludes` of `[tools.mise]`,
so `gate_files` drops those files first and `--debug` names them. To remove or
skip an item of `bootstrap`, write an override in the values file:

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
the file. The `shuck` block has no `associations`: shuck formats only the shell
files that it knows by extension (`.sh`, `.bash`, `.zsh` and `.ksh`), not the
task files below `.config/mise/tasks/`, which have no extension. The block
`ruff` of the `python` tool formats each `.py` and `.pyi` file, also a Python
task file below `.config/mise/tasks/`: dprint is the one formatter of the
Python files, and no `ruff format` gate runs. The sync
removes old `associations` from the block. It uses the first dprint config of
the repository that it
finds, for example `dprint.jsonc`. See
[`docs/design/tools.md`](../../design/tools.md).

## mise tools

| mise tool | Version  | Environments | Note      |
| --------- | -------- | ------------ | --------- |
| `dprint`  | `latest` | dev, ci      | Formatter |

## Tasks

| Task                    | Use                                                                                                                                                                                                                                                                      |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `code:format:05-dprint` | Check the files of `gate_files` with dprint (`--stdin-files`): the given files, else with `--all` every file, else the uncommitted files. It has no `#FILES` header, and dprint picks the files of its plugins. `--fix` formats them. Without dprint, it prints nothing. |
| `setup:dprint`          | Update the plugin versions in `dprint.jsonc` with `dprint config update`.                                                                                                                                                                                                |

`code:format` of the `mise` tool runs `code:format:05-dprint` first, then the
other `code:format:NN-<name>` tasks.

## Scripts

None.

## Contributions

| Tool         | Shared file                      | Items                                                                                  |
| ------------ | -------------------------------- | -------------------------------------------------------------------------------------- |
| `pre-commit` | `.config/pre-commit-config.yaml` | The hook block `format`, which calls `code:format --fix` with the files of the commit. |
| `vscode`     | `.vscode/extensions.json`        | `dprint.dprint`.                                                                       |

Other tools contribute blocks and plugins to `dprint.jsonc` in
`[contributes.dprint]`, and each tool file lists them. A tool can also add
excludes in `[contributes.mise]` (`code_files`).

## Values

The keys of `[tools.dprint]` in `.config/bootstrap/tools.toml`:

| Key                                   | Use                                                                                                                                                                       |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `env`                                 | The environments of the tool.                                                                                                                                             |
| `settings`                            | The top-level settings. The default is `newLineKind = "lf"`.                                                                                                              |
| `blocks.<plugin>`                     | The keys of a plugin block. The defaults are `toml`, `json`, `markdown`, `yaml` and `shuck`.                                                                              |
| `excludes`                            | More excludes of `dprint.jsonc` only, after `code_files.excludes` of `[tools.mise]`. The default is `[]`. A list adds its items. Use `code_files.excludes` for the gates. |
| `plugins`                             | More plugins. A list adds its items.                                                                                                                                      |
| `[tools.dprint.files."dprint.jsonc"]` | The overrides of the items of `bootstrap` (see "What the tool ships").                                                                                                    |

No questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Move each formatter step of the repository to a task `code:format:NN-<name>`.
2. Put the own excludes of the repository in `code_files.excludes` of
   `[tools.mise]`, for example a vendored folder. A submodule needs none:
   `gate_files` never examines a file of a submodule.

## Migration

- Remove `dprint` from the mise files of the repository.
- Remove shfmt. dprint formats shell scripts now.
- Remove taplo. dprint formats TOML now. Also remove the `[toml]` and
  `evenBetterToml.*` keys from `.vscode/settings.json`.
- M34 of `migrate.md`: the sync now writes `dprint.jsonc`. It removes the old
  `extends` of the global config `~/.config/dprint/dprint.jsonc`, and it keeps
  the own excludes of the file.
- M45 of `migrate.md`: the base excludes (`**/*-lock.yaml`, `**/*.svg`,
  `**/package-lock.json`) moved to `code_files.excludes` of the `mise` tool.
  Move `[tools.dprint] excludes` and the own excludes of `dprint.jsonc` there.
  The sync removes the `associations` of the `shuck` block, so shuck no longer
  formats the task files.
