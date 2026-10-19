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

| Path in the repository                | Kind   | Use                          |
| ------------------------------------- | ------ | ---------------------------- |
| `.config/mise/tasks/setup/deps/swift` | Synced | The task `setup:deps:swift`. |

VS Code extensions: `swiftlang.swift-vscode`.

## mise tools

None. Swift, SwiftPM and `sourcekit-lsp` come from the Xcode that
`xcode-select` selects.

## Tasks

| Task               | Use                                                                                                                                         |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:swift` | Resolve the SwiftPM and Xcode package dependencies from `Package.resolved`. `--upgrade` resolves them again to the newest allowed versions. |

`setup:deps:swift` acts on the repository root only. It applies when the root
has `Package.swift` or one `*.xcodeproj`, and runs for both when both exist.
More than one root `*.xcodeproj` fails.

- `Package.swift`: a normal run is
  `swift package resolve --force-resolved-versions`. `--upgrade` runs
  `swift package update`.
- `*.xcodeproj`: a normal run is `xcodebuild -resolvePackageDependencies` with
  `-onlyUsePackageVersionsFromResolvedFile`, and the packages go to
  `.build/SourcePackages`. `--upgrade` removes the project's
  `Package.resolved` and resolves again. If the resolve fails, the task puts
  the old file back.

The task fails on a host that is not macOS. It uses the Xcode that
`xcode-select` selects, and fails when no Xcode is selected.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. The tool is for macOS only. There is no Linux toolchain. If the user
   develops on a different platform, tell the user.
2. Xcode must be installed and selected:
   `sudo xcode-select -s /Applications/Xcode.app`.
3. Commit `Package.resolved` (for an `*.xcodeproj`, the file in
   `<project>/project.xcworkspace/xcshareddata/swiftpm/`). A normal run of
   `setup:deps:swift` installs only the versions that it names.
4. In a Flutter app, `setup:deps:flutter` handles the `ios/` and `macos/`
   folders. `setup:deps:swift` acts only on the repository root.
5. bootstrap does not read `XCODE_VERSION` or `SIMULATOR_*`. They stay with
   the `test:*` tasks of the stack that ships them.

## Migration

None.
