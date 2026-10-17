# The `rust` tool

Rust and the zola site generator.

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The repository has `Cargo.toml`.
- The repository has a zola site, for example a `config.toml` with
  `base_url` and a `content/` folder.

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/conf.d/rust/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool   | Version  | Environments | Note                                                 |
| ----------- | -------- | ------------ | ---------------------------------------------------- |
| `core:rust` | `latest` | dev, ci      | The `default` profile, which has clippy and rustfmt. |
| `zola`      | `latest` | dev, ci      |                                                      |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository has `rust-toolchain.toml`, tell the user. It can select
   a different toolchain than mise.
2. If the repository needs a Rust formatter or linter in `code:format` or
   `code:lint`, write a task `code:format:rust` or `code:lint:rust`.

## Migration

- Remove `core:rust` and zola from the repository-owned mise files.
