# The `go` tool

Go, with the gofmt and go vet gates.

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
| `.config/mise/tasks/code/format/gofmt`   | Synced    | The task `code:format:gofmt`. |
| `.config/mise/tasks/code/lint/go`        | Synced    | The task `code:lint:go`.      |
| `.config/mise/conf.d/go/mise.<env>.toml` | Generated | The mise config of this tool. |

The VS Code extension is `golang.go`.

## mise tools

| mise tool | Version  | Environments | Note |
| --------- | -------- | ------------ | ---- |
| `go`      | `latest` | dev, ci      |      |

mise sets `GOBIN` to the folder of the active Go version, unless the
environment already sets `GOBIN`. Thus `go install` puts its binaries there.

The tool also sets `GOTOOLCHAIN = "local"`. The Go docs
([Go Toolchains](https://go.dev/doc/toolchain)) define a toolchain as "the
standard library as well as the compiler, assembler, and other tools". With
`local`, the `go` command always runs the Go of mise, and never downloads
another toolchain. The setting has no effect on module dependencies: they
download as before. When the `go` line of `go.mod` or `go.work` needs a
newer Go than the mise version, the `go` command stops with an error.

## Tasks

| Task                | Use                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------ |
| `code:format:gofmt` | Check the format of the `.go` files with `gofmt -l`. `--fix` formats them with `-w`. |
| `code:lint:go`      | Run `go vet ./...` in each Go module.                                                |

`code:lint` and `code:format` run the two gates. Without `go`, or without a
`go.mod` that git does not ignore, a gate prints nothing and exits 0.

- `code:format:gofmt` uses only the `.go` files. Without files, it formats
  each `.go` file that git does not ignore. `gofmt -l` exits 0 when a file is
  not formatted, so the task fails when the list is not empty.
- `code:lint:go` examines packages, not files. A given `.go` file, `go.mod`
  or `go.sum` selects the module of its nearest `go.mod`. Without files, the
  task examines each module with a `go.mod` that git does not ignore. `--fix`
  is accepted, but `go vet` only reports.

## Pre-commit hooks

None.

## Files the repository supplies

| File      | Used by                                                    |
| --------- | ---------------------------------------------------------- |
| `go.mod`  | `code:lint:go` and `code:format:gofmt`. It marks a module. |
| `go.work` | The `go` command, when the modules are in one workspace.   |

## Adapt to the repository

1. If `go.mod` or `go.work` needs a newer Go than the mise version, tell the
   user. `GOTOOLCHAIN = "local"` stops the `go` command in that module.
2. If the repository has its own Go linter, for example golangci-lint, tell
   the user. The tool ships only `go vet`.

## Migration

- Remove go and `GOTOOLCHAIN` from the repository-owned mise files.
