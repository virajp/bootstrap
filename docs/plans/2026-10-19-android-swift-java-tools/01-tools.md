# U1 — Add the java, android and swift tools

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/java/tool.toml` (new), `tools/android/tool.toml` (new),
  `tools/swift/tool.toml` (new), `tools/flutter/tool.toml`,
  `tools/kotlin/tool.toml`, `docs/sync/java.md` (new), `docs/sync/android.md`
  (new), `docs/sync/swift.md` (new), `docs/sync/flutter.md`,
  `docs/sync/kotlin.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** every owned file that exists, top to bottom, before editing;
  also `tools/node/tool.toml` (a tool with `version = "lts"`) and
  `tools/bun/tool.toml` (a small tool), as style models.
- **Lazy-load:** `CLAUDE.md` (sections "Layout rule" and "mise config"),
  `docs/sync/index.md:143-182` (the `tool.toml` keys and the mise environment
  rules), `tools/base/.config/mise/tasks/_scripts/repo-sync:171-209,443-470`
  (manifest read, duplicate check, environments).

## Ruling

From index.md, assumed decisions:

> D2 — A new `java` tool. `tools/java/tool.toml` defines only
> `[mise.tools.java]`, `version = "lts"`, `envs = ["dev", "ci"]`.

> D3 — A new `android` tool defines `[mise.tools.android-sdk]`,
> `version = "23"`, `envs = ["dev", "ci"]`, mise `depends = ["java"]` — moved
> unchanged from `tools/flutter/tool.toml`. No `[env]` table: the plugin
> exports `ANDROID_HOME`, `ANDROID_SDK_ROOT` and `PATH`.

> D4 — `tools/flutter/tool.toml` drops `[mise.tools.android-sdk]` and
> `[mise.tools.java]` and sets `depends = ["android"]`.

> D5 — `tools/kotlin/tool.toml` sets `depends = ["java"]`. Under the sync rule
> a tool added only as a dependency is dev only, so `docs/sync/kotlin.md`
> "Adapt to the repository" tells the agent to add `java` to `tools` when CI
> builds with Gradle.

> D6 — `tools/swift/tool.toml` has no mise tool, `depends = []`,
> `extensions = ["swiftlang.swift-vscode"]`, `root_only = false`, `seed = []`.
> `docs/sync/swift.md` says: macOS only; Swift and `sourcekit-lsp` come from
> the Xcode that `xcode-select` selects; no Linux toolchain. Its tasks come in
> plan 2.

> D12 — `docs/sync/<tool>.md` and `docs/sync/vwf.md` are shipped instructions
> that agents run, so the edit units own them.

User answers (2026-10-19): java version "lts"; "flutter depends on android";
kotlin depends on java "Yes, now"; swift "Manifest + extension".

## Edits

1. **`tools/java/tool.toml`** (new) — the same layout and comment style as
   `tools/flutter/tool.toml`. First line: a one-line comment that describes the
   tool, for example `# The Java development kit (JDK), for Gradle and the
   Android SDK.` Then `depends = []`, `extensions = []`, `root_only = false`,
   `seed = []`, then:

   ```toml
   [mise.tools.java]
   envs = [ "dev", "ci" ]
   version = "lts"
   ```

2. **`tools/android/tool.toml`** (new) — description comment, for example
   `# The Android SDK command-line tools.` Then `depends = []`,
   `extensions = []`, `root_only = false`, `seed = []`, then the
   `[mise.tools.android-sdk]` table cut from `tools/flutter/tool.toml` byte
   for byte (`depends = ["java"]`, `envs = ["dev", "ci"]`, `version = "23"`,
   and any comment it carries). No `[env]` table.

3. **`tools/swift/tool.toml`** (new) — description comment, for example
   `# Swift and SwiftPM from the host Xcode (macOS only).` Then
   `depends = []`, `extensions = [ "swiftlang.swift-vscode" ]`,
   `root_only = false`, `seed = []`. No `[mise.*]` table.

4. **`tools/flutter/tool.toml`** — remove the `[mise.tools.android-sdk]` and
   `[mise.tools.java]` tables. Set `depends = [ "android" ]`. Keep
   `[mise.tools.flutter]` and the other keys. Change the description comment
   so that it no longer says that the tool ships the Android toolchain (for
   example `# Flutter and Dart. Adds the android tool for the Android build.`).

5. **`tools/kotlin/tool.toml`** — set `depends = [ "java" ]`. Nothing else.

6. **``docs/sync/java.md``** (new) — the fixed structure of a tool file (see
   Facts in index.md and the model ``docs/sync/kotlin.md``): title
   ``# The \``java\` tool`, one-line description, Property | Value table (Root
   only No, Depends None, Seeds None), then each fixed heading:
   - When the tool is relevant: the repository builds with Gradle or Maven,
     or another selected tool adds it (`android`, `kotlin`).
   - What the tool ships: only the generated
     `.config/mise/conf.d/java/mise.<env>.toml`.
   - mise tools: `java`, `lts`, dev and ci, note "mise resolves `lts` to the
     newest LTS release".
   - Tasks, Pre-commit hooks, Files the repository supplies: None.
   - Adapt to the repository: if a tool adds `java` only as a dependency, it
     is in dev only; to have it in CI, add `java` to `tools`. If the
     repository pins a different JDK, tell the user.
   - Migration: remove `java` from the repository-owned mise files.

7. **`docs/sync/android.md`** (new) — same structure. Depends: `java`.
   - When the tool is relevant: the repository has an Android build
     (`settings.gradle.kts` with the Android Gradle plugin, or `android/` of a
     Flutter app). The `flutter` tool adds it.
   - mise tools: `android-sdk`, `23`, dev and ci, note "Needs java; sets
     `ANDROID_HOME` and `ANDROID_SDK_ROOT`".
   - Tasks: None yet (plan 2 adds them). Use the words "None." as the other
     files do; do not mention plans.
   - Adapt to the repository: if the repository sets `ANDROID_HOME` by hand,
     tell the user that the tool sets it.
   - Migration: remove `android-sdk` from the repository-owned mise files.

8. **`docs/sync/swift.md`** (new) — same structure. Depends None.
   - When the tool is relevant: the repository has `Package.swift`, an
     `*.xcodeproj`, or `ios/` or `macos/` of a Flutter app.
   - What the tool ships: nothing but the VS Code extension
     `swiftlang.swift-vscode`.
   - mise tools: None. Swift, SwiftPM and `sourcekit-lsp` come from the Xcode
     that `xcode-select` selects.
   - Adapt to the repository: macOS only; there is no Linux toolchain. If the
     user develops on another platform, tell the user. Xcode must be
     installed and selected (`sudo xcode-select -s /Applications/Xcode.app`).
   - Migration: None.

9. **`docs/sync/flutter.md`** — Property table: Depends `android`. The
   description line: drop "with the Android toolchain" or say that the tool
   adds the `android` tool. The mise tools table: only `flutter`. Remove the
   `android-sdk` and `java` rows and the "Needs java" note. Keep the Tasks
   table and the licenses note (the task still installs `platform-tools`
   through `android-sdk`, which now comes from the `android` tool; say so in
   one sentence). Migration: keep removing `flutter`; say that `android-sdk`
   and `java` are now in `docs/sync/android.md` and `docs/sync/java.md`.

10. **`docs/sync/kotlin.md`** — Property table: Depends `java`. Adapt to the
    repository: add the D5 item — `java` comes as a dependency in dev only; if
    CI builds with Gradle, add `java` to `tools`.

## Verification

- Wave-gate lines 1 and 2 (index.md, Wave gate) pass.
- The full-selection dry-run passes: the line-2 command with
  `tools = ["base", "flutter", "kotlin", "android", "java", "swift"]` exits 0.
- `grep -rn 'mise.tools.java\]' tools/*/tool.toml` prints exactly one line,
  in `tools/java/tool.toml`.
- `grep -rn 'mise.tools.android-sdk\]' tools/*/tool.toml` prints exactly one
  line, in `tools/android/tool.toml`.
- Each of the three new `tool.toml` files has the four keys `depends`,
  `root_only`, `extensions` and `seed`.
- Each new `docs/sync/*.md` has the eight fixed headings, in order.

## Guardrails

- Do not touch `docs/sync/index.md`, `docs/sync/migrate.md`, `README.md` or
  `CLAUDE.md` — U3 owns them. Report each passage they falsify as a
  `DOCS FALSIFIED:` line.
- Do not touch `docs/sync/vwf.md` — U2 owns it.
- Do not change any task under `tools/flutter/.config/` — plan 2 owns tasks.
- Do not edit the synced copies at the repository root (`.config/`,
  `.gitignore`).
- `tool.toml` files are formatted by dprint/taplo in `mise run code`: keep the
  array spacing of the existing files (`[ "a", "b" ]`) and sorted keys.
- Delete with `rm`, never `git rm`.
- Markdown: follow the style of the existing `docs/sync/*.md` files (STE
  prose, aligned tables).

## Commit

`feat: add the java, android and swift tools`
