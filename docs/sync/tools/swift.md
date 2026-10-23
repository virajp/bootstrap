# The `swift` tool

Swift and SwiftPM from the host Xcode or Command Line Tools (macOS only), the
SwiftLint and swift-format gates, and the Xcode and simulator lookup of the
Xcode tasks.

| Property     | Value                                                |
| ------------ | ---------------------------------------------------- |
| Root only    | No                                                   |
| Depends      | None                                                 |
| Default envs | dev                                                  |
| Seeds        | `.config/swiftlint.yml`, `.config/swift-format.json` |

## When the tool is relevant

- The repository has `Package.swift` or an `*.xcodeproj`.
- The repository has the `ios/` or `macos/` folder of a Flutter app.

## What the tool ships

| Path in the repository                        | Kind      | Use                                                             |
| --------------------------------------------- | --------- | --------------------------------------------------------------- |
| `.config/mise/tasks/setup/deps/swift`         | Synced    | The task `setup:deps:swift`.                                    |
| `.config/mise/tasks/code/lint/swiftlint`      | Synced    | The task `code:lint:swiftlint`.                                 |
| `.config/mise/tasks/code/format/swift-format` | Synced    | The task `code:format:swift-format`.                            |
| `.config/mise/tasks/_scripts/xcode`           | Synced    | The library `xcode`, with `require_xcode` and `find_xcodeproj`. |
| `.config/swiftlint.yml`                       | Seed      | The SwiftLint rules. The layout rules are off for swift-format. |
| `.config/swift-format.json`                   | Seed      | The swift-format rules: 4 spaces, lines of 120 characters.      |
| `.config/mise/conf.d/swift/mise.<env>.toml`   | Generated | The mise config of this tool.                                   |

VS Code extensions: `swiftlang.swift-vscode`.

## mise tools

| mise tool   | Version  | Environments |
| ----------- | -------- | ------------ |
| `swiftlint` | `latest` | dev, ci      |

Swift, SwiftPM, swift-format and `sourcekit-lsp` come from the developer folder
that `xcode-select` selects: a full Xcode for a root `*.xcodeproj`, or the
Command Line Tools for a root with only `Package.swift`.

The tool also sets `os = "macos"` in `[mise_env.ci.settings]`. The `mise` tool
sets `os = "linux"` for ci, and `conf.d/swift/` loads after `conf.d/mise/`, so
this value wins.

## Tasks

| Task                       | Use                                                                                                                                         |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:swift`         | Resolve the SwiftPM and Xcode package dependencies from `Package.resolved`. `--upgrade` resolves them again to the newest allowed versions. |
| `code:lint:swiftlint`      | Lint the `.swift` files with `swiftlint lint --strict`. A warning is a failure.                                                             |
| `code:format:swift-format` | Check the format of the `.swift` files with `swift format lint --strict`. `--fix` formats them.                                             |

`code:lint` and `code:format` run the two gates. A gate skips without its tool
(SwiftLint from mise, swift-format from the host) and without a `.swift` file.
With a list of files, a gate uses only the `.swift` files. Without files, the
list comes from `git ls-files`, because SwiftLint and swift-format do not read
`.gitignore`. swift-format skips `*.generated.swift`. SwiftLint skips the
`excluded` paths of `.config/swiftlint.yml`.

The library `_scripts/xcode` is for the Xcode tasks of a stack, for example
`test:golden`. `require_xcode <task>` fails when no Xcode is selected. Then it
sets and exports four values. A value that the repository sets wins:

| Value                | Without a value of the repository                                                                                          |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `XCODE_VERSION`      | The selected Xcode. If the repository sets it, the selected Xcode must be that version, else the task fails.               |
| `SIMULATOR_PLATFORM` | `iOS`                                                                                                                      |
| `SIMULATOR_OS`       | The newest available iOS runtime that has an iPhone of the rule below.                                                     |
| `SIMULATOR_DEVICE`   | The available iPhone of that runtime with the highest model number and no other word in its name, for example `iPhone 17`. |

`find_xcodeproj <task>` prints the one `*.xcodeproj` at the root, and fails
for none or more than one.

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

| File                                          | Used by                                                                |
| --------------------------------------------- | ---------------------------------------------------------------------- |
| `.config/mise/conf.d/zz-repo/mise.<env>.toml` | Optional. `XCODE_VERSION` and the `SIMULATOR_*` values in its `[env]`. |

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
5. If the repository records golden images or needs one Xcode, set
   `XCODE_VERSION`, `SIMULATOR_PLATFORM`, `SIMULATOR_OS` and
   `SIMULATOR_DEVICE` in the `[env]` of
   `.config/mise/conf.d/zz-repo/mise.dev.toml`. A golden image is correct for
   one device only, and two machines can find two different simulators.
6. If the repository has its own SwiftLint or swift-format config, tell the
   user. The gates read only `.config/swiftlint.yml` and
   `.config/swift-format.json`.
7. To run the gates in CI, set `env = ["dev", "ci"]` in `[tools.swift]`.

   **CAUTION:** With `swift` in ci, mise downloads macOS binaries for each
   tool in ci. Thus each CI job of the repository must run on a macOS runner,
   for example `macos-latest` of GitHub Actions, or `saas-macos-medium-m1` of
   GitLab (beta, Premium and Ultimate only).

## Migration

- Delete the old stackgen copies: `conf.d/swiftlint/` and `conf.d/swiftui/`,
  and the task `code:lint:swift-format` (M19 of `migrate.md`).
- The old `.config/swiftlint.yml` and `.config/swift-format.json` of stackgen
  stay as they are: the sync does not write a seed that exists.
