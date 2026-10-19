# The `swift` tool

Swift and SwiftPM from the host Xcode (macOS only).

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The repository has `Package.swift` or an `*.xcodeproj`.
- The repository has the `ios/` or `macos/` folder of a Flutter app.

## What the tool ships

No files. The tool adds only a VS Code extension.

VS Code extensions: `swiftlang.swift-vscode`.

## mise tools

None. Swift, SwiftPM and `sourcekit-lsp` come from the Xcode that
`xcode-select` selects.

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. The tool is for macOS only. There is no Linux toolchain. If the user
   develops on a different platform, tell the user.
2. Xcode must be installed and selected:
   `sudo xcode-select -s /Applications/Xcode.app`.

## Migration

None.
