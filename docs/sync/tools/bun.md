# The `bun` tool

Bun, with a `bunfig.toml` that the repository owns.

| Property     | Value                |
| ------------ | -------------------- |
| Root only    | No                   |
| Depends      | None                 |
| Default envs | dev, ci, prod, stage |
| Seeds        | `bunfig.toml`        |

## When the tool is relevant

- The repository has `bun.lock`, `bun.lockb` or `bunfig.toml`.
- A script or a task calls `bun` or `bunx`.

## What the tool ships

| Path in the repository                    | Kind      | Use                                                         |
| ----------------------------------------- | --------- | ----------------------------------------------------------- |
| `bunfig.toml`                             | Seed      | The Bun config. The sync writes it only when it is missing. |
| `.config/mise/conf.d/bun/mise.<env>.toml` | Generated | The mise config of this tool.                               |

## mise tools

| mise tool | Version  | Environments         |
| --------- | -------- | -------------------- |
| `bun`     | `latest` | dev, ci, prod, stage |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository has a `bunfig.toml`, the seed keeps it. Tell the user.
2. If the repository uses both Bun and pnpm, ask the user which package
   manager owns the dependencies. The `node` tool runs `pnpm install` when a
   `package.json` exists.

## Migration

- Remove `bun` from the repository-owned mise files.
