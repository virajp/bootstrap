# The `go` tool

Go.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `go.mod` or `go.work`.

## What the tool ships

| Path in the repository                   | Kind      | Use                           |
| ---------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/go/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments | Note |
| --------- | -------- | ------------ | ---- |
| `go`      | `latest` | dev, ci      |      |

mise sets `GOBIN` to the folder of the active Go version, unless the
environment already sets `GOBIN`. Thus `go install` puts its binaries there.

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If `go.mod` or `go.work` has a `toolchain` line, or a `go` line with a
   patch version, tell the user. The Go command can download that toolchain
   when it is newer than the mise version.
2. If the repository needs a Go formatter or linter in `code:format` or
   `code:lint`, write a task `code:format:go` or `code:lint:go`.

## Migration

- Remove go from the repository-owned mise files.
