# U2 — The SwiftPM and Xcode deps task of the swift tool

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/swift/.config/mise/tasks/setup/deps/swift` (new, mode
  `755`), `docs/sync/swift.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** `tools/flutter/.config/mise/tasks/setup/deps/flutter` (the
  model task), `docs/sync/swift.md`, `CLAUDE.md` sections "Task scripts" and
  "Task output: headers and subheaders".
- **Lazy-load:** `tools/base/.config/mise/tasks/_scripts/helpers`,
  `docs/sync/flutter.md` (a Tasks table to copy the style of).

## Ruling

From index.md, assumed decisions:

> E1 — One task `setup:deps:<tool>` per tool, with a `--upgrade` flag, as
> `setup:deps:flutter`. A normal run installs only what the lockfiles name. No
> audit step: `code:sec` runs osv-scanner and grype over the repository.

> E2 — `setup:deps:*` acts on the repository root only (`./gradlew`,
> `Package.swift`, one `*.xcodeproj`). In a Flutter repository,
> `setup:deps:flutter` keeps the `android/` and `ios/` deps, so nothing runs
> twice.

> E4 — Applies when the root has `Package.swift` or exactly one
> `*.xcodeproj`; else it prints nothing and exits 0. More than one root
> `*.xcodeproj` fails and names them. On a host that is not macOS (`uname -s`
> is not `Darwin`) with a Swift project, it fails with a message that Swift is
> macOS only. A normal run: `Package.swift` →
> `swift package resolve --force-resolved-versions`; `*.xcodeproj` →
> `xcodebuild -resolvePackageDependencies -project <p>
> -clonedSourcePackagesDirPath .build/SourcePackages
> -onlyUsePackageVersionsFromResolvedFile`. `--upgrade`:
> `swift package update`; for the project, remove its `Package.resolved` and
> resolve without `-onlyUsePackageVersionsFromResolvedFile` (restore the file
> if the resolve fails). It reads no `XCODE_VERSION`; it uses the Xcode that
> `xcode-select` selects.

> E8 — `XCODE_VERSION` and `SIMULATOR_*` are not in bootstrap. They stay with
> the `test:*` tasks that stay in `claude-plugins`.

> E10 — Each task unit owns its `docs/sync/<tool>.md` (the "Tasks" table and
> "What the tool ships"), as D12 of plan 1. Tool calls go through `mise x --`;
> host commands (`git`, `awk`, `sed`, `find`, `uname`, `swift`, `xcodebuild`)
> do not.

## Edits

1. **`tools/swift/.config/mise/tasks/setup/deps/swift`** (new, `755`):
   - shebang, the bootstrap header for
     `tools/swift/.config/mise/tasks/setup/deps/swift`, `#MISE` and `#USAGE`
     headers in the flutter task's style, with `--upgrade`;
   - `set -euo pipefail`, source the helpers, `cd "${MISE_PROJECT_ROOT}"`;
   - detect `Package.swift` and the root `*.xcodeproj` folders with a bash 3.2
     safe glob (`shopt -s nullglob` or a `for` loop with an existence test);
     none → exit 0, no output;
   - more than one root `*.xcodeproj` → `print_error` naming them, exit 1;
   - a host that is not `Darwin` → `print_error` ("Swift is macOS only; this
     host is <name>"), exit 1;
   - `xcodebuild -version` failing → `print_error` with the fix
     `sudo xcode-select -s /Applications/Xcode.app`, exit 1;
   - `print_header` once, `print_subheader` per project kind;
   - the E4 commands, normal and `--upgrade`. The project's
     `Package.resolved` is at
     `<p>/project.xcworkspace/xcshareddata/swiftpm/Package.resolved`; for the
     upgrade copy it to a `mktemp` file first, and copy it back in a `trap`
     when the resolve fails;
   - both kinds can exist at the root; then run both.
2. **`docs/sync/swift.md`**: "What the tool ships" gets the task row;
   "Tasks" describes `setup:deps:swift` (root only, the two kinds, locked
   normal run, `--upgrade`, macOS only, the selected Xcode); "Adapt to the
   repository" says that `Package.resolved` must be committed, that the iOS
   part of a Flutter app is handled by `setup:deps:flutter`, and that
   `XCODE_VERSION` and `SIMULATOR_*` are not read by bootstrap.

## Verification

- Wave-gate lines 1 and 2 pass.
- `bash -n` on the task passes; the unit sets `chmod 755`.
- With `MISE_PROJECT_ROOT` set to an empty temporary git repository, the task
  prints nothing and exits 0.
- `grep -n 'XCODE_VERSION' tools/swift/.config/mise/tasks/setup/deps/swift`
  prints nothing.

## Guardrails

- Do not touch `tools/swift/tool.toml`, any other tool, or the docs that U5
  owns. Report each falsified passage as `DOCS FALSIFIED:`.
- Do not use `mise x --` for `swift` or `xcodebuild`: they come from the host
  Xcode.
- Delete with `rm`, never `git rm`.

## Commit

`feat(swift): add setup:deps:swift for SwiftPM and Xcode projects`
