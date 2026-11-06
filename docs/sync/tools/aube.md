# The `aube` tool

aube, the package manager of a repository that selects it. A repository with a
`package.json` selects this tool or the [`pnpm`](pnpm.md) tool, not both.

| Property     | Value                                             |
| ------------ | ------------------------------------------------- |
| Root only    | No                                                |
| Depends      | `node`, through its mise `depends`                |
| Default envs | dev, ci                                           |
| Seeds        | `aube-workspace.yaml`, only with a `package.json` |

## When the tool is relevant

- The repository has `aube-lock.yaml` or `aube-workspace.yaml`.
- The user selects aube as the package manager of the repository.

## What the tool ships

| Path in the repository                     | Kind      | Use                                                                            |
| ------------------------------------------ | --------- | ------------------------------------------------------------------------------ |
| `aube-workspace.yaml`                      | Seed      | The aube settings. The sync writes it only with a `package.json` (`requires`). |
| `.config/mise/conf.d/aube/mise.<env>.toml` | Generated | The mise config of this tool.                                                  |

The seed has the same strict settings as the seed of the `pnpm` tool:
`strictDepBuilds`, a minimum release age of 24 hours, `blockExoticSubdeps`,
`trustPolicy: no-downgrade` and `verifyDepsBeforeRun: error`.

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `aube`    | `latest` | dev, ci      |

The tool also sets the shell alias `npx = "aubx"`.

## Tasks

None. `setup:deps:node` of the `node` tool runs aube when this tool is selected.

## Pre-commit hooks

None.

## Files the repository supplies

| File                                 | Used by                                           |
| ------------------------------------ | ------------------------------------------------- |
| `package.json`                       | `setup:deps:node`. The task skips without it.     |
| `aube-lock.yaml` or `pnpm-lock.yaml` | `setup:deps:node`. The install uses its versions. |

aube writes a new project to `aube-lock.yaml`. It keeps an existing
`pnpm-lock.yaml` and updates it in place, at lockfile version 9 only. An older
version stops with `ERR_AUBE_UNSUPPORTED_PNPM_LOCKFILE_VERSION`. To convert it,
run `pnpm install` one time with a current pnpm.

## Settings outside the repository

The repository needs no user-level aube config. `aube-workspace.yaml` holds the
important settings, and the aube defaults supply the others. Two items are not
in the seed:

- A login for a private registry. Keep it outside the repository. Do not commit
  a token.
- An exclusion from the release age. If the repository needs a new release of
  its own package, add the scope to the workspace file.

## Adapt to the repository

1. If the repository has a `pnpm-workspace.yaml`, rename it to
   `aube-workspace.yaml` before the first sync. If both files exist,
   `aube-workspace.yaml` takes precedence, so the seed can replace the settings
   of the repository without a message.
2. Keep an existing `pnpm-lock.yaml`. Do not create `aube-lock.yaml` next to it.
3. Do not use pnpm and aube on the same `node_modules`. The package manager that
   installs last controls the top-level links, so run `aube install` after the
   move.
4. Tell the user that `aube run` and `aube exec` differ from pnpm: they install
   automatically when the state is stale, unless `verifyDepsBeforeRun` says
   otherwise.
5. Add each package with a build script to `allowBuilds`, or run
   `aube approve-builds`.

## Migration

- Remove aube and pnpm from the repository-owned mise files.
- Replace `pnpm` with `aube` in the scripts and the CI jobs of the repository.
