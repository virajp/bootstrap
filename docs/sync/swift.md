# The `swift` tool

Swift and SwiftPM from the host Xcode or Command Line Tools (macOS only).

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

None. Swift, SwiftPM and `sourcekit-lsp` come from the developer folder that
`xcode-select` selects: a full Xcode for a root `*.xcodeproj`, or the Command
Line Tools for a root with only `Package.swift`.

## Tasks

| Task               | Use                                                                                                                                         |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:swift` | Resolve the SwiftPM and Xcode package dependencies from `Package.resolved`. `--upgrade` resolves them again to the newest allowed versions. |

`setup:deps:swift` acts on the repository root only. It applies when the root
has `Package.swift` or one `*.xcodeproj`, and runs for both when both exist.
On macOS, more than one root `*.xcodeproj` fails.

- `Package.swift`: a normal run is
  `swift package resolve --force-resolved-versions`. `--upgrade` runs
  `swift package update`.
- `*.xcodeproj`: a normal run is `xcodebuild -resolvePackageDependencies` with
  `-onlyUsePackageVersionsFromResolvedFile`, and the packages go to
  `.build/SourcePackages`. `--upgrade` removes the project's
  `Package.resolved` and resolves again. If the resolve fails, the task puts
  the old file back.

On a host that is not macOS, the task prints a warning, skips and exits 0 for
each Swift root, also for a root with more than one `*.xcodeproj`. Thus `setup`
on a Linux CI host continues. It uses the developer folder that
`xcode-select` selects. A root `*.xcodeproj` needs a full Xcode, and the task
fails when no Xcode is selected. A root with only `Package.swift` needs only
the Command Line Tools, and the task fails when `swift` is not available.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. The tool is for macOS only. There is no Linux toolchain. On a different
   host, `setup:deps:swift` skips. If the user develops on a different
   platform, tell the user.
2. For a root `*.xcodeproj`, Xcode must be installed and selected:
   `sudo xcode-select -s /Applications/Xcode.app`. A root with only
   `Package.swift` needs only the Command Line Tools: `xcode-select --install`.
3. Commit `Package.resolved` (for an `*.xcodeproj`, the file in
   `<project>/project.xcworkspace/xcshareddata/swiftpm/`). A normal run of
   `setup:deps:swift` installs only the versions that it names. The task does
   not read a root `*.xcworkspace` or its `Package.resolved`.
4. In a Flutter app, `setup:deps:flutter` handles the `ios/` and `macos/`
   folders. `setup:deps:swift` acts only on the repository root.
5. bootstrap does not read `XCODE_VERSION` or `SIMULATOR_*`. They stay with
   the `test:*` tasks of the stack that ships them.

## Migration

None.
