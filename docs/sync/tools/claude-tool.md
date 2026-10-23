# The `claude` tool

The AI tools: graphify and its `.graphifyignore`, mempalace, `setup:ai` (the
Claude Code plugin installs and updates) and `code:graph`.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

- The user uses Claude Code in the repository, for example the repository has
  `CLAUDE.md` or `.claude/`.
- The repository uses graphify (`graphify-out/`) or mempalace.

A submodule can select the tool too. Then it gets `code:graph` and the
`graphify-refresh` hook. `setup:ai` skips in a submodule.

## What the tool ships

| Path in the repository                       | Kind      | Use                                                                                                  |
| -------------------------------------------- | --------- | ---------------------------------------------------------------------------------------------------- |
| `.graphifyignore`                            | Synced    | The files that graphify does not read, with `[overrides.graphifyignore]` of the values file applied. |
| `.config/mise/tasks/setup/ai`                | Synced    | The task `setup:ai`.                                                                                 |
| `.config/mise/tasks/code/graph`              | Synced    | The task `code:graph`.                                                                               |
| `.config/mise/conf.d/claude/mise.<env>.toml` | Generated | The mise config of this tool.                                                                        |

## mise tools

| mise tool        | Version  | Environments | Note                |
| ---------------- | -------- | ------------ | ------------------- |
| `pipx:graphifyy` | `latest` | dev          | Needs python and uv |
| `pipx:mempalace` | `latest` | dev          | Needs python and uv |

A submodule in its root checkout also gets these tools from the root.

## Tasks

| Task         | Use                                                                                                                                                                                                                                                                                           |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:ai`   | Install the missing plugins, update the marketplaces and the plugins of the repository, then remove the project-scope plugins that nothing references. It skips if `claude` is not installed, and in a git submodule. A failed step fails the task, and the other steps still run. See below. |
| `code:graph` | Refresh the graphify graph. Only the post-commit hook runs it. See below.                                                                                                                                                                                                                     |

`code:graph` runs `graphify update` in the background, with its log in
`~/.cache/graphify-rebuild.log`, so the commit does not wait. It exits 0 in
these cases:

- graphify is not installed. The task prints a warning.
- `GRAPHIFY_SKIP_HOOK=1` is set.
- The checkout is a linked worktree.
- A rebase, merge or cherry-pick is in progress.
- The commit changed only `graphify-out/`.

The argument `--force` rebuilds the graph, also when the commit changed only
`graphify-out/`.

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

graphify reads `.gitignore` and `.graphifyignore`. The shared
`.graphifyignore` holds only the committed vwf folders that are not code:
`docs/memory/`, `docs/plans/archived/`, `docs/prompts/` and `archived/`.

## Pre-commit hooks

- The local hook `graphify-refresh` in the post-commit stage, which calls
  `code:graph`.

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

1. If the repository has git hooks from `graphify hook install`, tell the
   user. The `graphify-refresh` hook of this tool replaces them.
2. If the user installs Claude Code plugins for this repository by hand, ask
   the user whether to list them in `[claude.plugins.user]` or
   `[claude.plugins.project]`.
3. If the repository has its own `.graphifyignore`, put each of its lines that
   `.gitignore` and the shared file do not have in `add` of
   `[overrides.graphifyignore]`. Put each shared line that graphify must read
   in `remove`. The sync replaces the file.

## Migration

- Remove `pipx:graphifyy` and `pipx:mempalace` from the repository-owned mise
  files.
- Delete an old `conf.d/ai/` (M7 of `migrate.md`).
