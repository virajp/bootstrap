# The `docker` tool

The local services in `.tools/docker-compose.yaml`, run by the
`setup:external:*` tasks.

| Property     | Value                        |
| ------------ | ---------------------------- |
| Root only    | Yes                          |
| Depends      | None                         |
| Parent       | None                         |
| Default envs | dev                          |
| Seeds        | `.tools/docker-compose.yaml` |

## When the tool is relevant

- The repository is a root repository, not a submodule.
- The development of the repository needs local services, for example a
  database, a queue or an emulator.
- The repository has a Docker Compose file for these services.

`bootstrap detect` runs the detect script of the tool (see "Scripts"). It
finds a Dockerfile, not a Compose file. A Dockerfile alone does not show that
the repository needs local services. Ask the user before you add the tool.

## What the tool ships

| Path in the repository               | Kind   | Use                                             |
| ------------------------------------ | ------ | ----------------------------------------------- |
| `.tools/docker-compose.yaml`         | Seed   | The local services. It starts with no services. |
| `.config/mise/tasks/setup/external/` | Synced | The tasks in "Tasks".                           |

## mise tools

None. The tool uses the Docker of the machine.

## Tasks

| Task                    | Use                                                                         |
| ----------------------- | --------------------------------------------------------------------------- |
| `setup:external:check`  | Make sure that the Docker server is reachable.                              |
| `setup:external:pull`   | Build the local images and pull the other images.                           |
| `setup:external:start`  | Start the local services.                                                   |
| `setup:external:stop`   | Stop the local services and remove their volumes.                           |
| `setup:external:update` | Stop, pull and start, at most once a week. `-f`/`--force` ignores the week. |

`setup:external:pull`, `start`, `stop` and `update` apply only when
`MISE_ENV` is `dev`. In another environment, they print nothing and exit 0.

**WARNING:** `setup:external:stop` removes the volumes of the services, and
`setup:external:update` runs it at most once a week. Tell the user before a
service keeps data that is not in a different location.

## Scripts

| Script           | Kind   | Use                                                                                                            |
| ---------------- | ------ | -------------------------------------------------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a `Dockerfile`, `Dockerfile.*` or `*.Dockerfile` at most two folders below the root, else `none`. |

The sync runs the scripts from the clone of `bootstrap`, and never copies
them. See [`docs/design/tools.md`](../../design/tools.md).

## Contributions

| Tool     | Shared file    | Items                                                                 |
| -------- | -------------- | --------------------------------------------------------------------- |
| `dprint` | `dprint.jsonc` | The plugin `npm:@dprint/dockerfile@0.7.0` and the block `dockerfile`. |

## Values

Only `env`. No questions. See
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File                         | Used by                                                                      |
| ---------------------------- | ---------------------------------------------------------------------------- |
| `.tools/docker-compose.yaml` | `setup:external:*`. The seed has no services. Ask the user for the services. |

## Adapt to the repository

1. If the repository has a Docker Compose file at a different path, ask the
   user whether to move its services to `.tools/docker-compose.yaml`.

## Migration

- Delete an old `setup:deps:check` task. `setup:external:check` replaces it
  (M3 of `migrate.md`).
- Delete each `#PLACEHOLDER` stub in `setup/external/` (M2).
- In a submodule, delete `setup/external/*` (M4).
