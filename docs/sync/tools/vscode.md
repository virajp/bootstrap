# The `vscode` tool

The VS Code settings, and the shared `extensions.json` with the extensions of
the selected tools. A member of the `base` pack that is not required:
`bootstrap init` writes its table in a new values file, and
`bootstrap remove tool vscode` removes it.

| Property     | Value |
| ------------ | ----- |
| Root only    | Yes   |
| Depends      | None  |
| Parent       | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

Always in a root repository: the `base` pack adds it. VS Code opens the root
repository. In a git submodule, `bootstrap init` writes no table for the tool,
and a `[tools.vscode]` table in a submodule stops the sync with code 2.

## What the tool ships

| Path in the repository                     | Kind      | Use                                                               |
| ------------------------------------------ | --------- | ----------------------------------------------------------------- |
| `.vscode/settings.json`                    | Synced    | The shared editor settings. The sync replaces the file.           |
| `.vscode/extensions.json`                  | Shared    | The extensions of this tool and of the selected tools. See below. |
| `.config/mise/tasks/setup/vscode`          | Synced    | The task `setup:vscode`.                                          |
| `.config/mise/conf.d/vscode/mise.dev.toml` | Generated | The mise config of this tool.                                     |

`.vscode/extensions.json` is a shared file. The sync renders the spec
`files/.vscode/extensions.json.merge.toml`, then adds each missing extension
to `recommendations`. The extensions are the `[defaults]` of this tool and the
`[contributes.vscode]` extensions of each selected tool. The repository keeps
its own extensions directly in the file, and the sync keeps them. To remove or
skip an extension of `bootstrap`, write an override in the values file:

```toml
[tools.vscode.files.".vscode/extensions.json".recommendations]
"usernamehw.errorlens" = { "$mode" = "remove" }
```

The merge rules are in [`docs/design/tools.md`](../../design/tools.md). The
root file lists only the extensions of the root. It does not list the
extensions of its git submodules.

## mise tools

None.

The tool also sets the shell alias `code`. It opens the root of the repository
in the VS Code profile `$REPO_NAME`, also from a subdirectory. The first
`code` or `setup:vscode` creates the profile. `code` associates the folder
with the profile only when no window of the folder is open. A new profile does
not use the user settings and keybindings of the Default profile. To use them,
turn on "Use Default profile" for them in the profile editor.

## Tasks

| Task           | Use                                                                                                                                                                                                                                                                        |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:vscode` | Make the VS Code profile of the repository match `extensions.json`. It installs the missing extensions and removes the others. If the profile does not exist, it opens a new empty window with the profile, which creates it. Then `code` opens the root with the profile. |

`setup` runs `setup:vscode`. The task uses a profile for each repository, not a
global install. Thus an extension of one repository does not stay on in each
other window.

## Scripts

None.

## Contributions

None. Other tools contribute extensions to `.vscode/extensions.json` in
`[contributes.vscode]`. Each tool file lists them.

## Values

The keys of `[tools.vscode]` in `.config/bootstrap.toml`:

| Key                                              | Use                                                                                                                                                               |
| ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `env`                                            | The environments of the tool.                                                                                                                                     |
| `extensions`                                     | More extensions. A list adds its items to the defaults `gruntfuggly.todo-tree`, `oderwat.indent-rainbow`, `pkief.material-icon-theme` and `usernamehw.errorlens`. |
| `[tools.vscode.files.".vscode/extensions.json"]` | The overrides of the extensions of `bootstrap` (see "What the tool ships").                                                                                       |

No questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. Move the keys of `.vscode/settings.json` that the repository owns to the
   user settings or to the workspace, and tell the user. The sync replaces the
   file. Examples are `[dart]`, `files.associations` and `markdownlint.*`.
2. If the repository needs an extension that no tool has, for example astro,
   tailwind or effect, keep it in `recommendations` of
   `.vscode/extensions.json`.

## Migration

- Remove a `[tools.vscode]` table without values from the values file (M20 of
  `migrate.md`).
- Delete `.config/vscode.d/`. Remove the `// >>> vscode.d` blocks from
  `.vscode/settings.json` and `.config/pre-commit-config.yaml`.
- In a submodule, delete `.vscode/` and `.config/mise/tasks/setup/vscode`
  (M4 of `migrate.md`).
- M34 of `migrate.md`: `bootstrap init` moves the old top-level `extensions`
  list into `.vscode/extensions.json`.
