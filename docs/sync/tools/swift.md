# The `swift` tool

Swift and SwiftPM from the host Xcode or Command Line Tools (macOS only), the
SwiftLint and swift-format gates, and the Xcode and simulator lookup of the
Xcode tasks.

| Property     | Value                                                |
| ------------ | ---------------------------------------------------- |
| Root only    | No                                                   |
| Depends      | None                                                 |
| Parent       | None                                                 |
| Default envs | dev                                                  |
| Hosts        | macos                                                |
| Seeds        | `.config/swiftlint.yml`, `.config/swift-format.json` |

## When the tool is relevant

The tool has `hosts = ["macos"]`. `bootstrap add tool <tool…>` and the new tool
of `bootstrap replace tool <old> <new>` refuse it on a host that is not macOS,
`bootstrap add pack <pack…>` leaves it out there, and `bootstrap detect` does
not propose it there. `sync` and `check` do not examine `hosts`, so a Linux CI
job still syncs a repository that has the tool.

- The repository has `Package.swift` or an `*.xcodeproj`.
- The repository has the `ios/` or `macos/` folder of a Flutter app.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
the root has `Package.swift` or an `*.xcodeproj` (see "Scripts"). It does not
look in the folders of a Flutter app.

## What the tool ships

| Path in the repository                           | Kind      | Use                                                             |
| ------------------------------------------------ | --------- | --------------------------------------------------------------- |
| `.config/mise/tasks/setup/deps/60-swift`         | Synced    | The task `setup:deps:60-swift`.                                 |
| `.config/mise/tasks/code/lint/80-swiftlint`      | Synced    | The task `code:lint:80-swiftlint`.                              |
| `.config/mise/tasks/code/format/50-swift-format` | Synced    | The task `code:format:50-swift-format`.                         |
| `.config/mise/tasks/test/golden/40-swift`        | Synced    | The part `test:golden:40-swift`.                                |
| `.config/mise/tasks/_scripts/xcode`              | Library   | The library `xcode`, with `require_xcode` and `find_xcodeproj`. |
| `.config/swiftlint.yml`                          | Seed      | The SwiftLint rules. The layout rules are off for swift-format. |
| `.config/swift-format.json`                      | Seed      | The swift-format rules: 4 spaces, lines of 120 characters.      |
| `.config/mise/conf.d/swift/mise.<env>.toml`      | Generated | The mise config of this tool.                                   |

## mise tools

| mise tool   | Version  | Environments |
| ----------- | -------- | ------------ |
| `swiftlint` | `latest` | dev, ci      |

Swift, SwiftPM, swift-format and `sourcekit-lsp` come from the developer folder
that `xcode-select` selects: a full Xcode for a root `*.xcodeproj`, or the
Command Line Tools for a root with only `Package.swift`.

The tool declares the language server `sourcekit-lsp` (`.swift`, a startup
timeout of 1000 ms) with the source `host`: no mise pin, and `code:lsp:check`
finds it with `xcrun --find sourcekit-lsp`. The language server is in
[`docs/design/tools.md`](../../design/tools.md) ("The `[lsp]` table").
`bootstrap lsp` lists it, and `code:lsp:check` of the `mise` tool tests that it
resolves.

The tool also sets `os = "macos"` in `[mise_env.ci.settings]`. The `mise` tool
sets `os = "linux"` for ci, and `conf.d/swift/` loads after `conf.d/mise/`, so
this value wins.

## Tasks

| Task                          | Use                                                                                                                                         |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:60-swift`         | Resolve the SwiftPM and Xcode package dependencies from `Package.resolved`. `--upgrade` resolves them again to the newest allowed versions. |
| `code:lint:80-swiftlint`      | Lint the `.swift` files with `swiftlint lint --strict`. A warning is a failure.                                                             |
| `code:format:50-swift-format` | Check the format of the `.swift` files with `swift format lint --strict`. `--fix` formats them.                                             |
| `test:golden:40-swift`        | A part of `test:golden`. Compare the views of the Xcode app against their golden images. `--record` records them again, then compares.      |

`code:lint` and `code:format` run the two gates. A gate skips without its tool
(SwiftLint from mise, swift-format from the host) and without a `.swift` file.
Each gate gets its files from `gate_files`: the given files, else with
`--all` every file, else the uncommitted files. Its `#FILES` header takes the
`.swift` files and skips `*.generated.swift`, `.build/`, `.swiftpm/`,
`Derived/` and `DerivedData/`. SwiftLint and swift-format get only that list,
so the `excluded` paths of the seed `.config/swiftlint.yml` do not act in a
gate run: put a path exclude in `code_files.excludes` of `[tools.mise]`. See
[`mise.md`](mise.md#the-file-sets-of-the-gates).

`test:golden:40-swift` has only the flag `--record`, because `test:golden`
passes the same arguments to each part. The other options are env values,
with a default:

| Value                   | Use                                                                                |
| ----------------------- | ---------------------------------------------------------------------------------- |
| `XCODE_SCHEME`          | The scheme whose test action holds the target. The name of the project by default. |
| `XCODE_SNAPSHOT_TARGET` | The snapshot test target to run, `SnapshotTests` by default.                       |
| `SIMULATOR_*`           | The simulator to render on. See `_scripts/xcode` below.                            |

- Without an `*.xcodeproj` at the root, the part prints nothing and exits 0.
- With one, it applies. More than one `*.xcodeproj` fails it through
  `find_xcodeproj`, and no Xcode fails it through `require_xcode`.
- A project whose `xcodebuild -list` has no target `XCODE_SNAPSHOT_TARGET`
  prints nothing and exits 0.
- It runs `xcodebuild test -only-testing:<target>` with the packages of
  `Package.resolved`. On a platform other than macOS, it needs
  `SIMULATOR_DEVICE` and `SIMULATOR_OS`, else it fails. For iOS,
  `require_xcode` finds them. `--record` records
  each golden first, then compares against the new ones. The failed
  comparisons leave their images in `.build/snapshot-artifacts`.

The library `_scripts/xcode` is for the Xcode tasks, for example
`test:golden:40-swift`. `require_xcode <task>` fails when no Xcode is selected.
Then it
sets and exports four values. A value that the repository sets wins:

| Value                | Without a value of the repository                                                                                          |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `XCODE_VERSION`      | The selected Xcode. If the repository sets it, the selected Xcode must be that version, else the task fails.               |
| `SIMULATOR_PLATFORM` | `iOS`                                                                                                                      |
| `SIMULATOR_OS`       | The newest available iOS runtime that has an iPhone of the rule below.                                                     |
| `SIMULATOR_DEVICE`   | The available iPhone of that runtime with the highest model number and no other word in its name, for example `iPhone 17`. |

`find_xcodeproj <task>` prints the one `*.xcodeproj` at the root, and fails
for none or more than one.

`setup:deps:60-swift` acts on the repository root only. It applies when the root
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

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                                |
| ---------------- | ------ | ------------------------------------------------------------------ |
| `scripts/detect` | detect | `match` with a root `Package.swift` or `*.xcodeproj`, else `none`. |

## Contributions

| Tool     | Shared file               | Items                    |
| -------- | ------------------------- | ------------------------ |
| `vscode` | `.vscode/extensions.json` | `swiftlang.swift-vscode` |

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File                                          | Used by                                                                                                         |
| --------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `.config/mise/conf.d/zz-repo/mise.<env>.toml` | Optional. `XCODE_VERSION`, `XCODE_SCHEME`, `XCODE_SNAPSHOT_TARGET` and the `SIMULATOR_*` values in its `[env]`. |

## Adapt to the repository

1. The tool is for macOS only. There is no Linux toolchain. On a different
   host, `setup:deps:60-swift` skips. If the user develops on a different
   platform, tell the user.
2. For a root `*.xcodeproj`, Xcode must be installed and selected:
   `sudo xcode-select -s /Applications/Xcode.app`. A root with only
   `Package.swift` needs only the Command Line Tools: `xcode-select --install`.
3. Commit `Package.resolved` (for an `*.xcodeproj`, the file in
   `<project>/project.xcworkspace/xcshareddata/swiftpm/`). A normal run of
   `setup:deps:60-swift` installs only the versions that it names. The task does
   not read a root `*.xcworkspace` or its `Package.resolved`.
4. In a Flutter app, `setup:deps:20-flutter` handles the `ios/` and `macos/`
   folders. `setup:deps:60-swift` acts only on the repository root.
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
- `bootstrap migrate` removes the old `test/golden` task of the swiftui pack of
  stackgen, at its old path or at `test/golden/50-local`. It knows each
  released version by its sha256 (`replaces`). A copy that the repository
  changed stays: compare it with the part, and delete it when the part does
  its job, else both run. The old `--target`, `--scheme` and simulator flags
  are now the env values above.
- The old `.config/swiftlint.yml` and `.config/swift-format.json` of stackgen
  stay as they are: the sync does not write a seed that exists.
