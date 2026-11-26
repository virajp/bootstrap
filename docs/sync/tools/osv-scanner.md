# The `osv-scanner` tool

osv-scanner: the dependency and container image scanner, and
`code:sec:30-osv-scanner`. It is the default audit of the dependencies of each
language that it reads. A required member of the `base` pack. `bootstrap remove
tool <tool…>` refuses it.

| Property     | Value                      |
| ------------ | -------------------------- |
| Root only    | No                         |
| Depends      | `_security_base`           |
| Parent       | None                       |
| Default envs | dev, ci                    |
| Seeds        | `.config/osv-scanner.toml` |

## When the tool is relevant

Always. The `base` pack adds this tool, and the sync writes an empty
`[tools.osv-scanner]` table for it. The table gives values.

## What the tool ships

| Path in the repository                            | Kind      | Use                                 |
| ------------------------------------------------- | --------- | ----------------------------------- |
| `.config/osv-scanner.toml`                        | Seed      | The config of osv-scanner.          |
| `.config/mise/tasks/code/sec/30-osv-scanner.py`   | Synced    | The task `code:sec:30-osv-scanner`. |
| `.config/mise/conf.d/osv-scanner/mise.<env>.toml` | Generated | The mise config of this tool.       |

## mise tools

| mise tool     | Version  | Environments | Note                  |
| ------------- | -------- | ------------ | --------------------- |
| `osv-scanner` | `latest` | dev, ci      | Vulnerability scanner |

## Tasks

| Task                      | Use                                                                                                                                                          |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `code:sec:30-osv-scanner` | Scan the dependencies of the tree, without the git submodules, then each container image. `--staged` does nothing. `--debug` raises the verbosity to `info`. |

`code:sec` of the `mise` tool runs this task. It does these steps:

1. It scans the lockfiles of each folder that has a file that git tracks,
   with `osv-scanner scan source --no-ignore` and no `--recursive`. Thus git
   decides which files to scan, and the scan does not go into a submodule.
   osv-scanner also applies each `.gitignore` above the repository (its
   issue 209), so a recursive scan of a worktree below an ignored folder, for
   example `.worktrees/`, finds nothing. osv-scanner has no parser for
   `aube-lock.yaml`, so the task gives each `aube-lock.yaml` that git tracks
   with `--lockfile pnpm-lock.yaml:<path>`: the file has the format of
   `pnpm-lock.yaml`. If a later aube format does not parse, the scan stops
   with an error.
2. It scans each image of `.tools/docker-compose.yaml` (the services with an
   `image`; a service with only a `build` has no image) and each entry of the
   value `images`. The task saves an image name with `docker save` (with
   `DOCKER_DEFAULT_PLATFORM` when it is set), and pulls it first when docker
   has no copy. Then osv-scanner reads the archive with
   `osv-scanner scan image --archive`. An entry that ends in `.tar` is an image
   archive, and docker is not necessary for it.

A vulnerability in a dependency fails the task. CI builds the images of the
value `images`, so a vulnerability in one of them also fails the task, and so
does an entry that the task cannot scan: docker is not installed, the pull
fails, or the archive is missing. The images of `.tools/docker-compose.yaml`
are for local use only. A clean local image is preferred, but a vulnerability
in it, or a local image that the task cannot scan, is only a warning.
`.config/osv-scanner.toml` holds the exceptions (`[[IgnoredVulns]]`).

## Scripts

None.

## Contributions

None. The `sec` hook of `_security_base` runs `code:sec`, which runs this
task.

## Values

| Key      | Default | Use                                                                                                                                                                                 |
| -------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `images` | `[]`    | The container images that the repository builds: an image name, for example `"app:dev"`, or the path of an image archive from the root, for example `"dist/app.tar"`. No questions. |

The `node` and `go` tools have the value `audit`: osv-scanner is their
default audit.

## Files the repository supplies

None.

## Adapt to the repository

None.

## Migration

- Remove `osv-scanner` from the mise files of the repository.
