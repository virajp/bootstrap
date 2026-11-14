# The `go` tool

Go, with golangci-lint, the `go.mod` check, `setup:deps:25-go` and
`govulncheck`. dprint formats the Go files with `gofmt`.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `go.mod` or `go.work`.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
the root has a `go.mod` (see "Scripts").

## What the tool ships

| Path in the repository                   | Kind      | Use                                                |
| ---------------------------------------- | --------- | -------------------------------------------------- |
| `.config/mise/tasks/code/lint/40-go`     | Synced    | The task `code:lint:40-go`. Only with a `go.mod`.  |
| `.config/mise/tasks/code/lint/45-gomod`  | Synced    | The task `code:lint:45-gomod`.                     |
| `.config/mise/tasks/setup/deps/25-go`    | Synced    | The task `setup:deps:25-go`. Only with a `go.mod`. |
| `.config/mise/conf.d/go/mise.<env>.toml` | Generated | The mise config of this tool.                      |

## mise tools

| mise tool                              | Version  | Environments | Note                                                                                 |
| -------------------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------ |
| `go`                                   | `latest` | dev, ci      |                                                                                      |
| `golangci-lint`                        | `latest` | dev, ci      | Only when the sync finds a `go.mod` (`requires_found`).                              |
| `go:golang.org/x/vuln/cmd/govulncheck` | `latest` | dev, ci      | Only when the sync finds a `go.mod`. The `go` backend builds it with the Go of mise. |

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

| Task                 | Use                                                                                                                      |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `code:lint:40-go`    | Run `golangci-lint run ./...` in each Go module. `--fix` applies the fixes that it can make.                             |
| `code:lint:45-gomod` | Run `go mod tidy -diff` in each Go module. It fails when `go.mod` or `go.sum` is not tidy. `--fix` runs `go mod tidy`.   |
| `setup:deps:25-go`   | Download and verify the modules, report the outdated ones, and audit them with `govulncheck`. `--upgrade` upgrades them. |

`code:lint` runs the two gates, and `setup` runs `setup:deps:25-go`. Without
the tool of a task, or without a `go.mod` that git does not ignore, the task
prints nothing and exits 0.

- `code:lint:40-go` and `code:lint:45-gomod` examine modules, not files. A
  given `.go` file, `go.mod` or `go.sum` selects the module of its nearest
  `go.mod`. Without files, they examine each module with a `go.mod` that git
  does not ignore.
- Without a config file, `golangci-lint` runs its standard linters: `errcheck`,
  `govet`, `ineffassign`, `staticcheck` and `unused`. A `.golangci.yml` in the
  module changes the set. The tool ships no config file.
- `setup:deps:25-go` runs `go mod download` and `go mod verify`: `go.sum`
  decides the versions. `--upgrade` runs `go get -u ./...` and `go mod tidy`.
  `-u` never changes a major version, because the major version is part of the
  module path. The report lists each direct dependency with a newer version.
  `govulncheck ./...` stops the setup when the code calls a vulnerable
  function. `osv-scanner` of `code:sec` also reads `go.mod`, but only by
  version.

dprint formats the `.go` files: the tool adds an `exec` command for `gofmt` to
the shared `dprint.jsonc` (see "Contributions"). `code:format:05-dprint` runs
dprint, so the tool has no format task of its own.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                   |
| ---------------- | ------ | ------------------------------------- |
| `scripts/detect` | detect | `match` with a `go.mod`, else `none`. |

## Contributions

| Tool     | Shared file               | Items                                                                                   |
| -------- | ------------------------- | --------------------------------------------------------------------------------------- |
| `dprint` | `dprint.jsonc`            | The plugin `@dprint/exec`, and an `exec` command for the `go` files: `mise x -- gofmt`. |
| `vscode` | `.vscode/extensions.json` | `golang.go`                                                                             |

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File      | Used by                                                                                         |
| --------- | ----------------------------------------------------------------------------------------------- |
| `go.mod`  | Each task of the tool, and the install of `golangci-lint` and `govulncheck`. It marks a module. |
| `go.work` | The `go` command, when the modules are in one workspace.                                        |

## Adapt to the repository

1. If `go.mod` or `go.work` needs a newer Go than the mise version, tell the
   user. `GOTOOLCHAIN = "local"` stops the `go` command in that module.
2. If the repository has its own golangci-lint config, keep it:
   `code:lint:40-go` uses it. If the repository has another Go linter, tell
   the user.
3. If `code:lint:40-go` finds problems in the code, show them to the user.

## Migration

- Remove go, golangci-lint, govulncheck and `GOTOOLCHAIN` from the
  repository-owned mise files.
- See M39 of `migrate.md` for a repository with `code:format:20-gofmt`.
