# The `claude` tool

The AI tools: `setup:ai`, the Claude Code plugin installs and updates.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

- The user uses Claude Code in the repository, for example the repository has
  `CLAUDE.md` or `.claude/`.

`setup:ai` skips in a submodule.

## What the tool ships

| Path in the repository        | Kind   | Use                  |
| ----------------------------- | ------ | -------------------- |
| `.config/mise/tasks/setup/ai` | Synced | The task `setup:ai`. |

## mise tools

None.

## Tasks

| Task       | Use                                                                                                                                                                                                                                                                                           |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:ai` | Install the missing plugins, update the marketplaces and the plugins of the repository, then remove the project-scope plugins that nothing references. It skips if `claude` is not installed, and in a git submodule. A failed step fails the task, and the other steps still run. See below. |

`setup:ai` does these steps in this sequence:

1. If no scope of the repository has `vwf@virajp-plugins`, it runs
   `pnpx @virajp.dev/claude-plugins@latest --all`, which installs it at user
   scope. Without `pnpx`, it prints a warning with the command.
2. It installs each plugin of `[claude.plugins.user]` and
   `[claude.plugins.project]` of `.config/bootstrap.toml` that is missing at
   that scope. If no scope has the marketplace of the plugin, it first adds
   the marketplace from the value, at the same scope.
3. It updates the marketplaces: `claude plugin marketplace update`.
4. It updates each plugin at user scope, and each plugin at project scope of
   this repository, one by one. `claude` has no command that updates all
   plugins.
5. It removes the project-scope plugins that nothing references.

A plugin at project scope of a different repository does not count for this
repository.

## Pre-commit hooks

None.

## Files the repository supplies

| File                     | Used by                                                                                                 |
| ------------------------ | ------------------------------------------------------------------------------------------------------- |
| `.config/bootstrap.toml` | Optional. `[claude.plugins.user]` and `[claude.plugins.project]`, the plugins that `setup:ai` installs. |

Each key is a plugin, `<name>@<marketplace>`. Its value is the source of the
marketplace for `claude plugin marketplace add`: a GitHub repository, a URL or
a path. Both tables are optional, and they can be empty:

```toml
[claude.plugins.user]
"taste-skill@taste-skill" = "Leonxlnx/taste-skill"

[claude.plugins.project]
```

The sync does not read these tables. `setup:ai` reads them when it runs.

## Adapt to the repository

1. If the user installs Claude Code plugins for this repository by hand, ask
   the user whether to list them in `[claude.plugins.user]` or
   `[claude.plugins.project]`.

## Migration

- The tool no longer ships graphify and mempalace (M24 of `migrate.md`).
- Delete an old `conf.d/ai/` (M7 of `migrate.md`).
