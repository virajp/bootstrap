# The `node` tool

Node. A repository with a `package.json` also selects one package manager: the
[`pnpm`](pnpm.md) tool or the [`aube`](aube.md) tool.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `package.json`.
- The core language of the repository is JavaScript or TypeScript.

The `linter` tool adds `node` in dev and ci, through its mise `depends`. A
repository that uses Node only as a development tool does not need to list
`node`, and it needs no package manager. A TypeScript repository lists
`[tools.node]`, which then goes to each environment of its `default_envs` that
the repository selects.

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/tasks/setup/deps/node`       | Synced    | The task `setup:deps:node`.   |
| `.config/mise/conf.d/node/mise.<env>.toml` | Generated | The mise config of this tool. |

VS Code extensions: `christian-kohler.npm-intellisense`,
`meganrogge.template-string-converter`.

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

| Task              | Use                                                                                                         |
| ----------------- | ----------------------------------------------------------------------------------------------------------- |
| `setup:deps:node` | Install from the lockfile, then report the outdated and vulnerable dependencies. `--upgrade` upgrades them. |

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

## Pre-commit hooks

None.

## Files the repository supplies

| File           | Used by                                       |
| -------------- | --------------------------------------------- |
| `package.json` | `setup:deps:node`. The task skips without it. |

## Adapt to the repository

1. If the repository has a `package.json`, ask the user which package manager
   the repository uses, and select `[tools.pnpm]` or `[tools.aube]`. A
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
