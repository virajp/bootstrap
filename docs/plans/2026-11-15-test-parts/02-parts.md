# U2 — The test parts of the android and swift tools

- **Read first:** `repo/tools/android/tools.toml`,
  `repo/tools/android/tasks/code-lint-android` (the header, the "does not
  apply" rule), `repo/tools/swift/tools.toml`, `repo/tools/swift/tasks/xcode`,
  and the three pack tasks named in the facts, read from the claude-plugins
  repo at `/Users/virajpatel/Projects/github.com/virajp/claude-plugins`
  (`git show 0562cb44:<path>`).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1:** three parts, each a task file below `tasks/` with its
  `[tasks."<path>"]` entry:
  - android `test/e2e/40-android`, source `test-e2e-android`;
  - android `test/golden/40-android`, source `test-golden-android`;
  - swift `test/golden/40-swift`, source `test-golden-swift`.
- **D2:** `test/golden/40-android` reads the Gradle build files: the
  Roborazzi plugin id (`io.github.takahirom.roborazzi`) runs
  `recordRoborazzi<Variant>` or `verifyRoborazzi<Variant>`; the Paparazzi
  plugin id (`app.cash.paparazzi`) runs `recordPaparazzi<Variant>` or
  `verifyPaparazzi<Variant>`; neither, nothing. **D11:** it scans
  `settings.gradle(.kts)`, every `build.gradle(.kts)` and
  `gradle/libs.versions.toml` for the ids; with both ids it runs both
  tasks; with an id and no `./gradlew` it prints nothing and exits 0.
- **D3:** each golden part declares one flag, `--record`; each e2e part
  declares none. Every other option is an environment value, read with a
  default: `ANDROID_VARIANT` (`debug`), `ANDROID_DEVICE` (`e2e`),
  `ANDROID_CONNECTED` (`true` runs `connected<Variant>AndroidTest`, and then
  `ANDROID_DEVICE` is ignored, D12),
  `XCODE_SCHEME` (the project name), `XCODE_SNAPSHOT_TARGET`
  (`SnapshotTests`), and the existing `SIMULATOR_PLATFORM`,
  `SIMULATOR_DEVICE`, `SIMULATOR_OS`. The header comment of each part names
  its values and says a repo sets them in the `[env]` of
  `.config/mise/conf.d/zz-repo/mise.<env>.toml`.
- **D6:** a part that does not apply prints nothing and exits 0:
  - `40-android` (e2e): no `./gradlew` at the root, or no
    `src/androidTest` folder in a module;
  - `40-android` (golden): neither plugin id in the Gradle build files;
  - `40-swift`: no `.xcodeproj` at the root. **D23:** once an `.xcodeproj`
    exists the part applies: more than one at the root exits 1 with the
    `find_xcodeproj` message; no Xcode exits 1 through `require_xcode`; it
    finds the target with `xcodebuild -list`, and no target named
    `XCODE_SNAPSHOT_TARGET` prints nothing and exits 0.
    A failure after the part applies (a failed Gradle or `xcodebuild` run) exits
    with that command's code. **D12:** the e2e part with no Android SDK after it
    applies exits 1 with the pack task's message.
- **D4:** each entry's `replaces` holds the `sha256:` of the matching pack file
  at claude-plugins commit `0562cb44` (`git show 0562cb44:<path> | shasum -a
  256`). **D13:** the swift part also lists the hash of the same file at the
  tags `stackgen-v1.31.0`, `stackgen-v1.33.0`, `stackgen-v2.0.0` and
  `stackgen-v3.0.0`. Record every value in the DECIDED line.
- Each new file has the header of the other task files: the synced-from line
  and the "do not edit" lines, `#MISE description` and `#MISE hide=true`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Touch nothing outside `repo/tools/android/` and `repo/tools/swift/`.
- Do not change the other tasks of the two tools.

## Commit

`feat: the android and swift tools ship their e2e and golden test parts`
