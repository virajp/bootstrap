# The `rust` tool

Rust and the zola site generator.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `Cargo.toml`.
- The repository has a zola site, for example a `config.toml` with
  `base_url` and a `content/` folder.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
the root has a `Cargo.toml` (see "Scripts"). It does not find a zola site.

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

None. dprint formats the `.rs` files with rustfmt (see "Contributions"), so
`code:format:05-dprint` of the `dprint` tool formats them.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                       |
| ---------------- | ------ | ----------------------------------------- |
| `scripts/detect` | detect | `match` with a `Cargo.toml`, else `none`. |

## Contributions

| Tool     | Shared file    | Items                                                                                                    |
| -------- | -------------- | -------------------------------------------------------------------------------------------------------- |
| `dprint` | `dprint.jsonc` | The plugin `@dprint/exec`, and an `exec` command for the `rs` files: `mise x -- rustfmt --edition 2024`. |

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository has `rust-toolchain.toml`, tell the user. It can select
   a different toolchain than mise.
2. If the repository needs a Rust linter in `code:lint`, for example clippy,
   write a task `code:lint:<name>`.

## Migration

- Remove `core:rust` and zola from the repository-owned mise files.
