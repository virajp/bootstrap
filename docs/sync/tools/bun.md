# The `bun` tool

Bun, with a `bunfig.toml` that the repository owns.

| Property     | Value         |
| ------------ | ------------- |
| Root only    | No            |
| Depends      | None          |
| Parent       | None          |
| Default envs | dev, ci       |
| Seeds        | `bunfig.toml` |

## When the tool is relevant

- The repository has `bun.lock`, `bun.lockb` or `bunfig.toml`.
- A script or a task calls `bun` or `bunx`.

`bootstrap detect` runs the detect script of the tool (see "Scripts").

## What the tool ships

| Path in the repository                    | Kind      | Use                                                         |
| ----------------------------------------- | --------- | ----------------------------------------------------------- |
| `bunfig.toml`                             | Seed      | The Bun config. The sync writes it only when it is missing. |
| `.config/mise/conf.d/bun/mise.<env>.toml` | Generated | The mise config of this tool.                               |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `bun`     | `latest` | dev, ci      |

## Tasks

None.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                        |
| ---------------- | ------ | ------------------------------------------ |
| `scripts/detect` | detect | `match` with a `bunfig.toml`, else `none`. |

## Contributions

| Tool     | Shared file    | Items                                                                                                                                              |
| -------- | -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `dprint` | `dprint.jsonc` | The blocks `typescript`, `markup` and `tailwindcss`, and the plugins `@dprint/typescript`, `dprint-plugin-markup` and `dprint-plugin-tailwindcss`. |

The `node` tool adds the same blocks and plugins.

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository has a `bunfig.toml`, the seed keeps it. Tell the user.
2. If the repository uses Bun and also pnpm or aube, ask the user which package
   manager owns the dependencies. If the repository also selects `pnpm` or
   `aube`, `setup:deps:40-node` runs it when a `package.json` exists. With only
   `bun`, that task prints nothing.

## Migration

- Remove `bun` from the repository-owned mise files.
