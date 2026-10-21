# The `node` tool

Node and pnpm.

| Property     | Value                                             |
| ------------ | ------------------------------------------------- |
| Root only    | No                                                |
| Depends      | None                                              |
| Default envs | dev, ci                                           |
| Seeds        | `pnpm-workspace.yaml`, only with a `package.json` |

## When the tool is relevant

- The repository has `package.json`.
- The core language of the repository is JavaScript or TypeScript.

The `linter` tool adds `node` in dev and ci, through its mise `depends`. A
repository that uses Node only as a development tool does not need to list
`node`. A TypeScript repository lists `[tools.node]`, which then goes to each
environment of its `default_envs` that the repository selects.

## What the tool ships

| Path in the repository                     | Kind      | Use                                                               |
| ------------------------------------------ | --------- | ----------------------------------------------------------------- |
| `pnpm-workspace.yaml`                      | Seed      | The pnpm settings. The sync writes it only with a `package.json`. |
| `.config/mise/tasks/setup/deps/node`       | Synced    | The task `setup:deps:node`.                                       |
| `.config/mise/conf.d/node/mise.<env>.toml` | Generated | The mise config of this tool.                                     |

The seed has strict settings: `strictDepBuilds`, a minimum release age of 24
hours, `blockExoticSubdeps`, `trustPolicy: no-downgrade` and
`verifyDepsBeforeRun: error`.

VS Code extensions: `christian-kohler.npm-intellisense`,
`meganrogge.template-string-converter`.

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `node`    | `lts`    | dev, ci      |
| `pnpm`    | `latest` | dev, ci      |

The tool also sets:

- `npm.package_manager = "pnpm"` and `node.compile = false`.
- `node_modules/.bin` in the `PATH`.
- The shell alias `npx = "pnpm dlx"`.

## Tasks

| Task              | Use                                                                                                         |
| ----------------- | ----------------------------------------------------------------------------------------------------------- |
| `setup:deps:node` | Install from the lockfile, then report the outdated and vulnerable dependencies. `--upgrade` upgrades them. |

A vulnerability at or above `audit.level` of `pnpm-workspace.yaml` stops the
setup. The task never runs `pnpm self-update`, because mise installs pnpm.

## Pre-commit hooks

None.

## Files the repository supplies

| File             | Used by                                           |
| ---------------- | ------------------------------------------------- |
| `package.json`   | `setup:deps:node`. The task skips without it.     |
| `pnpm-lock.yaml` | `setup:deps:node`. The install uses its versions. |

## Adapt to the repository

1. If the repository has a `pnpm-workspace.yaml`, the seed keeps it. Compare
   it with the seed, and tell the user about the strict settings that it does
   not have.
2. Move the pnpm settings of `.npmrc` to `pnpm-workspace.yaml`, except the
   auth and registry settings. pnpm 11 reads only those from `.npmrc`.
3. If the repository uses npm or yarn, tell the user. The task uses pnpm.
4. Add each package with a build script to `allowBuilds`.

## Migration

- Remove node and pnpm from the repository-owned mise files.
- If the old files pin a different version, tell the user. For example,
  95octane pinned node `26` and pnpm `12` (M6 of `migrate.md`).
- Delete the old `setup/pnpm/*` tasks (M2).
