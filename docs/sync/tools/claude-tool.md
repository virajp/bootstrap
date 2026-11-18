# The `claude` tool

The AI tools: `setup:ai`, which installs the listed Claude Code plugins,
uninstalls the project plugins that it installed and that left the list, and
updates the plugins with `--upgrade`. No pack has this tool: select it with
`[tools.claude]`.

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

| Task       | Use                                                                                                                                                                                                                                                                                                                                                                                       |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:ai` | Install the listed plugins that are missing, then uninstall the project-scope plugins that it installed and that left the list. With `--upgrade`, also update the marketplaces and the listed plugins. `setup --upgrade` runs it with `--upgrade`. It skips if `claude` is not installed, and in a git submodule. A failed step fails the task, and the other steps still run. See below. |

`setup:ai` names no plugin. It reads the installed plugins once, at the start,
then does these steps in this sequence:

1. It reads `[tools.claude.plugins.user]`, `[tools.claude.plugins.project]` and
   the record `tools.claude.plugins.installed` with `bootstrap config get`:
   the values of `.config/bootstrap/tools.toml`, with the entries that other
   selected tools contribute, for example `vwf@virajp-plugins` of the
   `virajp-plugins` tool. A failed read stops the task with the code of
   `bootstrap config get`: code 2 without an engine that answers, for example
   one older than 4.0.0, and code 2 with no catalog (no `--clone`, no
   `BOOTSTRAP_CLONE` and no cache clone; `get` never fetches).
2. It drops from the record each recorded plugin that is no longer installed at
   project scope.
3. It installs each listed plugin that is missing at its scope. If no scope has
   the marketplace of the plugin, it first adds the marketplace from the value,
   at the same scope. Each listed project plugin that is installed, before or
   now, goes in the record. A plugin goes in the record only after its install
   succeeds.
4. It uninstalls each recorded plugin that is no longer in
   `[tools.claude.plugins.project]`:
   `claude plugin uninstall --scope project --prune -y`, which also removes the
   dependencies that only it needed. A failed uninstall keeps the plugin in the
   record, so the next run tries again. A plugin that is not in the record, for
   example one installed by hand, and a plugin at user scope, are never
   uninstalled.
5. With `--upgrade` only, it updates the marketplaces
   (`claude plugin marketplace update`), then each listed plugin that is
   installed, at its scope, one by one. `claude` has no command that updates
   all plugins.
6. When the record changed, it writes it with
   `bootstrap config set tools.claude.plugins.installed '<JSON list>'`. A failed
   write is a failed step. The next run corrects the record from the installed
   plugins.

A plain run with nothing to change changes nothing.

A plugin at project scope of a different repository does not count for this
repository.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                      |
| ---------------- | ------ | -------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a `CLAUDE.md` or a `.claude/`, else `none`. |

## Contributions

None. Other tools contribute plugin entries in `[contributes.claude]`, for
example `virajp-plugins`. Each tool file lists them.

## Values

The keys of `[tools.claude]` in `.config/bootstrap/tools.toml`:

| Key                              | Use                                                                           |
| -------------------------------- | ----------------------------------------------------------------------------- |
| `env`                            | The environments of the tool.                                                 |
| `[tools.claude.plugins.user]`    | The plugins that `setup:ai` installs at user scope.                           |
| `[tools.claude.plugins.project]` | The plugins that `setup:ai` installs at project scope.                        |
| `plugins.installed`              | The record of `setup:ai`: the project plugins that it installed. Not by hand. |

Each key of the two plugin tables is a plugin, `<name>@<marketplace>`. Its value
is the source of the
marketplace for `claude plugin marketplace add`: a GitHub repository, a URL or
a path. Both tables are `[defaults]` of the tool, empty by default, so
`bootstrap config set` accepts them. Both tables are optional, and they can be
empty:

```toml
[tools.claude.plugins.user]
"taste-skill@taste-skill" = "Leonxlnx/taste-skill"

[tools.claude.plugins.project]
```

`plugins.installed` is a list of `<name>@<marketplace>`, empty by default.
`setup:ai` writes it: each project plugin that it installed, or that it found
installed while listed. It uninstalls a recorded plugin that leaves
`[tools.claude.plugins.project]`. Do not edit it by hand.

The sync does not use these keys. `setup:ai` reads them with
`bootstrap config get tools.claude.plugins.<key>` when it runs, so a table has
the entries that other selected tools contribute too. No questions. The rules of
the keys are in
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
- M34 of `migrate.md`: `bootstrap migrate` moves the old top-level
  `[claude.plugins.*]` tables into `[tools.claude.plugins.*]`.
- M41 of `migrate.md`: the tables are in `.config/bootstrap/tools.toml`, and
  `setup:ai` needs `bootstrap` 4.0.0 or newer.
- M44 of `migrate.md`: `setup:ai` no longer runs
  `pnpx @virajp.dev/claude-plugins@latest --all`. Add the `virajp-plugins`
  tool for the vwf plugin. `setup:ai` no longer updates the plugins on each
  run: run `mise run setup:ai --upgrade`, or `mise run setup --upgrade`.
