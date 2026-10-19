# U4 — setup:deps:swift: Xcode only for a project, skip off macOS

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/swift/.config/mise/tasks/setup/deps/swift`,
  `docs/sync/swift.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** both owned files, top to bottom; `CLAUDE.md` sections "Task
  scripts" and "Task output: headers and subheaders".
- **Lazy-load:** `tools/base/.config/mise/tasks/_scripts/helpers` (status
  printers).

## Ruling

From index.md, assumed decisions:

> R14 — `xcodebuild -version` runs only for a root with an `.xcodeproj`. A
> SwiftPM-only root resolves with `swift package resolve` and the Command Line
> Tools. A root `*.xcworkspace` stays unsupported, as
> `docs/sync/swift.md:67-68` says.

> R15 — On a non-Darwin host with a Swift root, the task prints its header,
> then `print_warn` "Swift dependencies are macOS only, skipped", and exits 0.

## Edits

1. **`tools/swift/.config/mise/tasks/setup/deps/swift`, host check (:38-42)**
   (R15): keep the applicability exit at :25 first. On a non-Darwin host,
   `print_header` (the task's normal header), then `print_warn` "Swift
   dependencies are macOS only, skipped", then exit 0.
2. **Same file, Xcode check (:46-49)** (R14): move `xcodebuild -version` into
   the `.xcodeproj` path, so it runs only when PROJECTS is not empty. For a
   SwiftPM-only root, examine that `swift` is on `PATH` (a host command, no
   `mise x --`); if it is not, `print_error` names the Command Line Tools
   (`xcode-select --install`) and the task exits 1.
3. **Same file**: no workspace support; no other behaviour change.
4. **`docs/sync/swift.md`**:
   - :48, :61-62: on a non-macOS host, the task prints a warning, skips and
     exits 0, so `setup` on a Linux CI host continues.
   - :48-49, :63-64: full Xcode is necessary only for a root `*.xcodeproj`; a
     SwiftPM-only root needs only the Command Line Tools.
   - :67-68: keep the workspace sentence.

## Verification

- Wave-gate lines 1 and 2 pass (ShellCheck runs in line 1).
- `bash -n` on the task passes.
- With `MISE_PROJECT_ROOT` set to an empty temporary git repository, the task
  prints nothing and exits 0.

## Guardrails

- Do not touch `tools/swift/tool.toml`, any other tool, or the docs that U6
  owns. Report each falsified passage as `DOCS FALSIFIED:`.
- Keep the script safe for bash 3.2 and 5.x.
- Delete with `rm`, never `git rm`.

## Commit

`fix(swift): need Xcode only for a project and skip off macOS`
