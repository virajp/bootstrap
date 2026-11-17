# The `virajp-plugins` tool

The vwf plugin of Claude Code: the `.gitignore` lines of the vwf working notes
and scratch, and the plugin entry that `setup:ai` of the `claude` tool
installs. No pack has this tool: select it with
`bootstrap add tool virajp-plugins`.

| Property     | Value           |
| ------------ | --------------- |
| Root only    | Yes             |
| Depends      | `claude`, `git` |
| Parent       | None            |
| Default envs | dev             |
| Seeds        | None            |

## When the tool is relevant

- The repository uses the vwf plugin, for example it has `docs/plans/`,
  `docs/memory/` or `.config/vwf.yaml`.

Only the root repository can select it. A submodule member cannot.

## What the tool ships

None. The tool has no file of its own: it adds to the files and values of
other tools (see "Contributions").

## mise tools

None.

## Tasks

None. `setup:ai` of the `claude` tool installs the plugin.

## Scripts

None.

## Contributions

| Tool     | Shared file or values         | Items                                                                                                                                                  |
| -------- | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `git`    | `.gitignore`                  | `docs/memory/handoff/`, `docs/memory/doctor/`, `docs/memory/runs/`, `docs/scratchpad/`, and `!docs/plans/**/*.log`, which keeps the plan logs tracked. |
| `claude` | `[tools.claude.plugins.user]` | `"vwf@virajp-plugins" = "virajp/claude-plugins"`.                                                                                                      |

The lines of `.gitignore` are `[[contributes_items.git.".gitignore".item]]`
tables: the sync adds them after the lines of the `git` tool, and the override
table `[tools.git.files.".gitignore"]` applies to them. The plugin entry is a
`[contributes.claude]` value: `bootstrap config get tools.claude.plugins.user`
returns it with `contributed_by` `["virajp-plugins"]`, and `setup:ai` installs
it at user scope.

## Values

Only `env`. No questions.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

- M44 of `migrate.md`: the `git` tool no longer ships the vwf lines, and
  `setup:ai` no longer runs the `pnpx` install of vwf. Add this tool.
