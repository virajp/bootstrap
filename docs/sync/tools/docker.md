# The `docker` tool

The local services in `.tools/docker-compose.yaml`, run by the
`setup:external:*` tasks.

| Property     | Value                        |
| ------------ | ---------------------------- |
| Root only    | Yes                          |
| Depends      | None                         |
| Default envs | dev                          |
| Seeds        | `.tools/docker-compose.yaml` |

## When the tool is relevant

- The repository is a root repository, not a submodule.
- The development of the repository needs local services, for example a
  database, a queue or an emulator.
- The repository has a Docker Compose file for these services.

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

## Pre-commit hooks

None.

## Files the repository supplies

| File                         | Used by                                                                      |
| ---------------------------- | ---------------------------------------------------------------------------- |
| `.tools/docker-compose.yaml` | `setup:external:*`. The seed has no services. Ask the user for the services. |

## Adapt to the repository

1. If the repository has a Docker Compose file at a different path, ask the
   user whether to move its services to `.tools/docker-compose.yaml`.

**WARNING:** `setup:external:stop` removes the volumes of the services, and
`setup:external:update` runs it at most once a week. Tell the user before a
service keeps data that is not in a different location.

## Migration

- Delete an old `setup:deps:check` task. `setup:external:check` replaces it
  (M3 of `migrate.md`).
- Delete each `#PLACEHOLDER` stub in `setup/external/` (M2).
- In a submodule, delete `setup/external/*` (M4).
