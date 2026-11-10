# The `pnpm` tool

pnpm, the package manager of a repository that selects it. A repository with a
`package.json` selects this tool or the [`aube`](aube.md) tool, not both.

| Property     | Value                                                                    |
| ------------ | ------------------------------------------------------------------------ |
| Root only    | No                                                                       |
| Depends      | `node`                                                                   |
| Parent       | `node`. `bootstrap replace` changes a tool for another with that parent. |
| Default envs | dev, ci                                                                  |
| Seeds        | `pnpm-workspace.yaml`, only with a `package.json`                        |

## When the tool is relevant

- The repository has `pnpm-lock.yaml` or `pnpm-workspace.yaml`, and the user
  does not want to move to aube.
- The user selects pnpm as the package manager of the repository.

`bootstrap detect` runs the detect script of the tool (see "Scripts").

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

## Scripts

The sync and the CLI run the scripts of the tool from the clone, and never
copy them.

| Script                 | Kind    | Use                                                                                                                                                                                                 |
| ---------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scripts/detect`       | detect  | `match` with a `pnpm-workspace.yaml`. `mismatch` with an `aube-workspace.yaml`, because then aube is the package manager. Else `none`.                                                              |
| `scripts/migrate-aube` | migrate | For `bootstrap replace aube pnpm`. It renames `aube-workspace.yaml` to `pnpm-workspace.yaml`, and writes a new `pnpm-lock.yaml` in place of `aube-lock.yaml`. Examine the diff of the new lockfile. |

`bootstrap replace` works only between two tools with the same parent, and
only in a work tree without changes that are not committed.

## Contributions

None.

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

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
5. To move an adopted repository from aube to pnpm, run
   `bootstrap replace aube pnpm`.

## Migration

- Remove pnpm from the repository-owned mise files.
- If the old files pin a different version, tell the user. For example, a
  repository pins pnpm `12` (M6 of `migrate.md`).
- Delete the old `setup/pnpm/*` tasks (M2).
