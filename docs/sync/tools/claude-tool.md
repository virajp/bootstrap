# The `claude` tool

The AI tools: graphify, mempalace, `setup:ai` and `code:graph`.

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

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/tasks/setup/ai`                | Synced    | The task `setup:ai`.          |
| `.config/mise/tasks/code/graph`              | Synced    | The task `code:graph`.        |
| `.config/mise/conf.d/claude/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool        | Version  | Environments | Note                |
| ---------------- | -------- | ------------ | ------------------- |
| `pipx:graphifyy` | `latest` | dev          | Needs python and uv |
| `pipx:mempalace` | `latest` | dev          | Needs python and uv |

A submodule in its root checkout also gets these tools from the root.

## Tasks

| Task         | Use                                                                                                                                                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:ai`   | Update the Claude Code marketplaces and plugins, then remove the project-scope plugins that nothing references. It skips if `claude` is not installed, and in a git submodule. A failed step fails the task, and the other steps still run. |
| `code:graph` | Refresh the graphify graph. Only the post-commit hook runs it. See below.                                                                                                                                                                   |

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

## Pre-commit hooks

- The local hook `graphify-refresh` in the post-commit stage, which calls
  `code:graph`.

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository has git hooks from `graphify hook install`, tell the
   user. The `graphify-refresh` hook of this tool replaces them.

## Migration

- Remove `pipx:graphifyy` and `pipx:mempalace` from the repository-owned mise
  files.
- Delete an old `conf.d/ai/` (M7 of `migrate.md`).
