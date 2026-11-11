# The `melange` tool

melange, which builds apk packages from a declarative pipeline. It is a member
of the `container-build` pack.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has melange package configs.
- A script or a task of the repository calls `melange`.

`bootstrap detect` runs the detect script of the tool (see "Scripts"). It
finds a melange config by its top-level keys `package:` and `pipeline:`,
because melange has no fixed file name.

## What the tool ships

| Path in the repository                        | Kind      | Use                           |
| --------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/melange/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments |
| --------- | -------- | ------------ |
| `melange` | `latest` | dev, ci      |

## Tasks

None.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                                                                                                               |
| ---------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a `*.yaml` or `*.yml` file, at most two folders below the root, that has the top-level keys `package:` and `pipeline:`, else `none`. |

## Contributions

None.

## Values

Only `env`. No questions. See
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File                 | Used by   |
| -------------------- | --------- |
| The package configs. | `melange` |

## Adapt to the repository

1. If the repository sets the runner with `MELANGE_RUNNER`, keep the value in
   the `zz-repo` mise files. The runners are `bubblewrap`, `docker` and `qemu`.
   The correct runner depends on the host and on the CI job, so the tool does
   not set it. For example, macOS has no bubblewrap, so a repository can use
   `docker` in dev and `bubblewrap` in a CI job that has no Docker daemon.

## Migration

- Remove `melange` from the repository-owned mise files. Keep `MELANGE_RUNNER`.
