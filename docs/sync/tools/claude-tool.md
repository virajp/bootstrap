# The `claude` tool

The AI tools: `setup:ai`, the Claude Code plugin installs and updates. No
pack has this tool: select it with `[tools.claude]`.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Parent       | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

- The user uses Claude Code in the repository, for example the repository has
  `CLAUDE.md` or `.claude/`.

`bootstrap detect` runs the detect script of the tool (see "Scripts"). It
finds `CLAUDE.md` and `.claude/` at the root.

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
2. It installs each plugin of `[tools.claude.plugins.user]` and
   `[tools.claude.plugins.project]` of `.config/bootstrap.toml` that is missing
   at
   that scope. If no scope has the marketplace of the plugin, it first adds
   the marketplace from the value, at the same scope.
3. It updates the marketplaces: `claude plugin marketplace update`.
4. It updates each plugin at user scope, and each plugin at project scope of
   this repository, one by one. `claude` has no command that updates all
   plugins.
5. It removes the project-scope plugins that nothing references.

A plugin at project scope of a different repository does not count for this
repository.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                      |
| ---------------- | ------ | -------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a `CLAUDE.md` or a `.claude/`, else `none`. |

## Contributions

None.

## Values

The keys of `[tools.claude]` in `.config/bootstrap.toml`:

| Key                              | Use                                                    |
| -------------------------------- | ------------------------------------------------------ |
| `env`                            | The environments of the tool.                          |
| `[tools.claude.plugins.user]`    | The plugins that `setup:ai` installs at user scope.    |
| `[tools.claude.plugins.project]` | The plugins that `setup:ai` installs at project scope. |

Each key is a plugin, `<name>@<marketplace>`. Its value is the source of the
marketplace for `claude plugin marketplace add`: a GitHub repository, a URL or
a path. Both tables are optional, and they can be empty:

```toml
[tools.claude.plugins.user]
"taste-skill@taste-skill" = "Leonxlnx/taste-skill"

[tools.claude.plugins.project]
```

The sync does not use these tables. `setup:ai` reads them when it runs. No
questions. The rules of the keys are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. If the user installs Claude Code plugins for this repository by hand, ask
   the user whether to list them in `[tools.claude.plugins.user]` or
   `[tools.claude.plugins.project]`.

## Migration

- The tool no longer ships graphify and mempalace (M24 of `migrate.md`).
- Delete an old `conf.d/ai/` (M7 of `migrate.md`).
- M34 of `migrate.md`: `bootstrap init` moves the old top-level
  `[claude.plugins.*]` tables into `[tools.claude.plugins.*]`.
