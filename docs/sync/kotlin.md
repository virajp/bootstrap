# The `kotlin` tool

The Kotlin language server.

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The repository has Kotlin code: `*.kt` or `*.kts` files, for example the
  Android part of a Flutter app.
- The user edits Kotlin with an editor or an agent that uses a language
  server.

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/conf.d/kotlin/mise.dev.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool         | Version  | Environments | Note                                                        |
| ----------------- | -------- | ------------ | ----------------------------------------------------------- |
| `http:kotlin-lsp` | `latest` | dev          | From the JetBrains downloads, for macOS arm64 and x64 only. |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. The tool has binaries for macOS only. If the user develops on a different
   platform, tell the user.

## Migration

- Remove `http:kotlin-lsp` from the repository-owned mise files.
