# The `pnpm` tool

pnpm, the package manager of a repository that selects it. A repository with a
`package.json` selects this tool or the [`aube`](aube.md) tool, not both.

| Property     | Value                                                      |
| ------------ | ---------------------------------------------------------- |
| Root only    | No                                                         |
| Depends      | `node`, through its mise `depends`                         |
| Parent       | `node`. The tools with this parent can replace each other. |
| Default envs | dev, ci                                                    |
| Seeds        | `pnpm-workspace.yaml`, only with a `package.json`          |

## When the tool is relevant

- The repository has `pnpm-lock.yaml` or `pnpm-workspace.yaml`, and the user
  does not want to move to aube.
- The user selects pnpm as the package manager of the repository.

## What the tool ships

| Path in the repository                     | Kind      | Use                                                                            |
| ------------------------------------------ | --------- | ------------------------------------------------------------------------------ |
| `pnpm-workspace.yaml`                      | Seed      | The pnpm settings. The sync writes it only with a `package.json` (`requires`). |
| `.config/mise/conf.d/pnpm/mise.<env>.toml` | Generated | The mise config of this tool.                                                  |

The seed has strict settings: `strictDepBuilds`, a minimum release age of 24
hours, `blockExoticSubdeps`, `trustPolicy: no-downgrade` and
`verifyDepsBeforeRun: error`.

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `pnpm`    | `latest` | dev, ci      |

The tool also sets the shell alias `npx = "pnpm dlx"`.

## Tasks

None. `setup:deps:40-node` of the `node` tool runs pnpm when this tool is
selected.

## Pre-commit hooks

None.

## Files the repository supplies

| File             | Used by                                              |
| ---------------- | ---------------------------------------------------- |
| `package.json`   | `setup:deps:40-node`. The task skips without it.     |
| `pnpm-lock.yaml` | `setup:deps:40-node`. The install uses its versions. |

## Settings outside the repository

The repository needs no user-level pnpm config. `pnpm-workspace.yaml` holds the
important settings, and the pnpm defaults supply the others. Two items are not
in the seed:

- A login for a private registry. pnpm reads it from `~/.npmrc` or from the
  `auth.ini` of its global config folder. Do not commit a token.
- An exclusion from the release age. If the repository needs a new release of
  its own package, add the scope to `minimumReleaseAgeExclude`.

## Adapt to the repository

1. If the repository has a `pnpm-workspace.yaml`, the seed keeps it. Compare
   it with the seed, and tell the user about the strict settings that it does
   not have.
2. Move the pnpm settings of `.npmrc` to `pnpm-workspace.yaml`, except the
   auth and registry settings. pnpm 11 reads only those from `.npmrc`.
3. If the repository uses npm or yarn, tell the user. The task uses pnpm.
4. Add each package with a build script to `allowBuilds`.

## Migration

- Remove pnpm from the repository-owned mise files.
- If the old files pin a different version, tell the user. For example, a
  repository pins pnpm `12` (M6 of `migrate.md`).
- Delete the old `setup/pnpm/*` tasks (M2).
