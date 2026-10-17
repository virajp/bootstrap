# The `claude` tool

The AI tools: graphify, mempalace and `setup:ai`. Submodules inherit the tools
from the root.

| Property  | Value |
| --------- | ----- |
| Root only | Yes   |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The repository is a root repository, not a submodule. The sync refuses this
  tool in a submodule.
- The user uses Claude Code in the repository, for example the repository has
  `CLAUDE.md` or `.claude/`.
- The repository uses graphify (`graphify-out/`) or mempalace.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/tasks/setup/ai`                | Synced    | The task `setup:ai`.          |
| `.config/mise/conf.d/claude/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool        | Version  | Environments | Note         |
| ---------------- | -------- | ------------ | ------------ |
| `pipx:graphifyy` | `latest` | dev          | Needs python |
| `pipx:mempalace` | `latest` | dev          | Needs python |

A submodule in its root checkout gets these tools from the root. Thus the
`code:graph` task of `base` also works in a submodule.

## Tasks

| Task       | Use                                                                                     |
| ---------- | --------------------------------------------------------------------------------------- |
| `setup:ai` | Update the Claude Code marketplaces and plugins. It skips if `claude` is not installed. |

## Pre-commit hooks

None. The `graphify-refresh` hook comes from `base`.

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository has git hooks from `graphify hook install`, tell the
   user. The `graphify-refresh` hook of `base` replaces them.

## Migration

- Remove `pipx:graphifyy` and `pipx:mempalace` from the repository-owned mise
  files.
- Delete an old `conf.d/ai/` (M7 of `migrate.md`).
- In a submodule, delete `setup:ai` and `conf.d/claude/` (M4).
