# The `cloudflare` tool

The Cloudflare CLI (`npm:cf`).

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The repository has `wrangler.toml`, `wrangler.json` or `wrangler.jsonc`.
- The repository deploys to Cloudflare Workers or Pages.

## What the tool ships

| Path in the repository                           | Kind      | Use                           |
| ------------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/conf.d/cloudflare/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments | Note                                                   |
| --------- | -------- | ------------ | ------------------------------------------------------ |
| `npm:cf`  | `latest` | dev, ci      | Needs node and pnpm. The `workerd` build is permitted. |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If a script or a task calls `wrangler`, tell the user. `npm:cf` replaces
   it.

## Migration

- Remove `npm:cf` from the repository-owned mise files.
- Remove `npm:wrangler`. It is deprecated, and `npm:cf` replaces it.
