# The `vscode` tool

The VS Code settings, and `extensions.json` from the extensions of the
selected tools.

| Property     | Value |
| ------------ | ----- |
| Root only    | Yes   |
| Depends      | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

Always in a root repository: the `base` pack adds it. Do not write
`[tools.vscode]` in the values file. VS Code opens the root repository, and
its `extensions.json` also has the extensions of each git submodule. In a
submodule, the sync skips the tool. A `[tools.vscode]` entry in a submodule
stops the sync with code 2.

## What the tool ships

| Path in the repository                     | Kind      | Use                                                                                          |
| ------------------------------------------ | --------- | -------------------------------------------------------------------------------------------- |
| `.vscode/settings.json`                    | Synced    | The shared editor settings. The sync replaces the file.                                      |
| `.vscode/extensions.json`                  | Template  | The extensions of the selected tools, of the values key `extensions`, and of each submodule. |
| `.config/mise/tasks/setup/vscode`          | Synced    | The task `setup:vscode`.                                                                     |
| `.config/mise/conf.d/vscode/mise.dev.toml` | Generated | The mise config of this tool.                                                                |

VS Code extensions: `gruntfuggly.todo-tree`, `oderwat.indent-rainbow`,
`pkief.material-icon-theme`, `usernamehw.errorlens`.

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

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. Move the repository-owned keys of `.vscode/settings.json` to the user
   settings or to the workspace, and tell the user. The sync replaces the
   file. Examples are `[dart]`, `files.associations` and `markdownlint.*`.
2. If the repository needs an extension that no tool has, for example astro,
   tailwind or effect, tell the user. Put it in the values key `extensions` of
   the repository that needs it. A submodule lists its own.

## Migration

- Remove `[tools.vscode]` from the values file (M20 of `migrate.md`).
- Delete `.config/vscode.d/`. Remove the `// >>> vscode.d` blocks from
  `.vscode/settings.json` and `.config/pre-commit-config.yaml`. This tool
  writes `.vscode/extensions.json` from the tools.
- In a submodule, delete `.vscode/` and `.config/mise/tasks/setup/vscode`
  (M4 of `migrate.md`).
