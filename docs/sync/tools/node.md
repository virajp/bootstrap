# The `node` tool

Node. A repository with a `package.json` also selects one package manager: the
[`pnpm`](pnpm.md) tool or the [`aube`](aube.md) tool.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `package.json`.
- The core language of the repository is JavaScript or TypeScript.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
the root has a `package.json` (see "Scripts").

The `linter`, `cloudflare` and `firebase` tools have `node` in their
`depends`. A repository that uses Node only as a development tool does not
need to list `node`, and it needs no package manager. A TypeScript repository
lists `[tools.node]`, which then goes to each environment of its
`default_envs` that the repository selects. The environments of a dependency
are in [`docs/design/mise.md`](../../design/mise.md).

## What the tool ships

| Path in the repository                     | Kind      | Use                            |
| ------------------------------------------ | --------- | ------------------------------ |
| `.config/mise/tasks/setup/deps/40-node`    | Synced    | The task `setup:deps:40-node`. |
| `.config/mise/conf.d/node/mise.<env>.toml` | Generated | The mise config of this tool.  |

## mise tools

| mise tool | Version | Environments |
| --------- | ------- | ------------ |
| `node`    | `lts`   | dev, ci      |

The tool also sets:

- `npm.package_manager = "aube"` and `node.compile = false`. This is the
  installer of the `npm:` mise tools. It is not the package manager of the
  repository.
- `node_modules/.bin` in the `PATH`.

## Tasks

| Task                 | Use                                                                                                         |
| -------------------- | ----------------------------------------------------------------------------------------------------------- |
| `setup:deps:40-node` | Install from the lockfile, then report the outdated and vulnerable dependencies. `--upgrade` upgrades them. |

The task finds the package manager from `.config/mise/conf.d/`:

| Selection                   | Result                                              |
| --------------------------- | --------------------------------------------------- |
| No `package.json`           | The task prints nothing and exits 0.                |
| `pnpm`                      | The task runs pnpm.                                 |
| `aube`                      | The task runs aube.                                 |
| `bun`, and no other         | The task prints nothing and exits 0.                |
| `pnpm` and `aube`           | The task stops with a message, and the setup stops. |
| None, with a `package.json` | The task stops with a message, and the setup stops. |

A vulnerability at or above `audit.level` of the workspace file stops the setup.
The task never runs a self-update, because mise installs the package manager.

`--upgrade` runs `update --recursive --no-save`. It moves the lockfile to the
newest versions inside the ranges of each `package.json`, and it does not change
a range. A package in `update.ignoreDeps` of the workspace file stays at its
locked version.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                     |
| ---------------- | ------ | ------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a `package.json` at the root, else `none`. |

## Contributions

| Tool     | Shared file               | Items                                                                                                                                              |
| -------- | ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `vscode` | `.vscode/extensions.json` | `christian-kohler.npm-intellisense`, `meganrogge.template-string-converter`.                                                                       |
| `dprint` | `dprint.jsonc`            | The blocks `typescript`, `markup` and `tailwindcss`, and the plugins `@dprint/typescript`, `dprint-plugin-markup` and `dprint-plugin-tailwindcss`. |

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File           | Used by                                          |
| -------------- | ------------------------------------------------ |
| `package.json` | `setup:deps:40-node`. The task skips without it. |

## Adapt to the repository

1. If the repository has a `package.json`, ask the user which package manager
   the repository uses. Select `pnpm` or `aube` in
   "Step 1: Select the tools and the packs" of [`index.md`](../index.md). A
   `pnpm-lock.yaml` is not a decision: aube can also use that file.
2. If the repository uses `typescript` or `tsx`, keep them as
   `devDependencies` in `package.json`. `bootstrap` ships no tool for them,
   because each package can need its own version, and other packages need
   `typescript` as a peer.

## Migration

- Remove node from the repository-owned mise files.
- If a repository-owned mise file has `npm:typescript` or `npm:tsx`, move it to
  `devDependencies` in `package.json`. Then make sure that the scripts that
  call it still find it.
- If the old files pin a different version, tell the user. For example, a
  repository pins node `26` (M6 of `migrate.md`).
- Do the "Migration" items of the package manager that the repository selects.
  M30 of `migrate.md` has the change from the old `node` tool.
