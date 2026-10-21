# U7 — Gates: the wave gate, the skip runs and the scratch checks

- **Wave:** 4
- **Depends on:** U6
- **Owns:** none. A fix that a gate needs goes back to the unit that owns the
  file, through the orchestrator.
- **Model:** opus
- **Kind:** edit
- **Read first:** index.md sections "Wave gate" and "Gates the orchestrator
  keeps"; `tools/kotlin/.config/mise/tasks/setup/deps/kotlin`;
  `tools/swift/.config/mise/tasks/setup/deps/swift`.

## Ruling

From index.md, assumed decisions:

> R4 — The init script sets `LockMode.STRICT` on each project. The normal run
> resolves without `--write-locks`, so Gradle pins the locked versions and
> fails when a locked configuration has no lock state (a new module). The git
> compare, the RESOLVED and UNCOMMITTED split and the restore leave the normal
> run. `--upgrade` does not change (`--write-locks --refresh-dependencies`).

> R14 — `xcodebuild -version` runs only for a root with an `.xcodeproj`. A
> SwiftPM-only root resolves with `swift package resolve` and the Command Line
> Tools.

This plan names no generator. No version file exists to bump. The gates unit
uses `mise x gradle@latest -- gradle init` and `swift package init` in
temporary folders only.

## Edits

1. Run wave-gate lines 1 and 2 (index.md, Wave gate), from the root of the
   worktree.
2. Skip runs: make an empty temporary git repository (`git init`). For each
   of `tools/kotlin/.config/mise/tasks/setup/deps/kotlin`,
   `tools/swift/.config/mise/tasks/setup/deps/swift`,
   `tools/android/.config/mise/tasks/setup/deps/android` and
   `tools/android/.config/mise/tasks/code/lint/android`, run the script from
   the worktree with `MISE_PROJECT_ROOT` set to the temporary repository and
   a copy of `tools/base/.config/mise/tasks/_scripts/helpers` at
   `<tmp>/.config/mise/tasks/_scripts/helpers`.
3. Scratch Kotlin lock check, in a new temporary folder:
   1. `mise x gradle@latest -- gradle init --type kotlin-application --dsl
      kotlin` (non-interactive flags as `gradle init --help` gives them), then
      add a second subproject with one external dependency;
   2. add `dependencyLocking { lockAllConfigurations() }` to every project,
      and `distributionSha256Sum` (the real checksum of the wrapper
      distribution) to `gradle/wrapper/gradle-wrapper.properties`;
   3. `./gradlew dependencies --write-locks` for every project, `git init`,
      commit everything; copy the helpers as in step 2; run
      `setup:deps:kotlin` with `MISE_PROJECT_ROOT` set to the folder;
   4. add `com.google.guava:guava:31.+` to one subproject, run
      `./gradlew dependencies --write-locks`, then change the guava lines of
      that lockfile to `31.0-jre` (a scratch file, not a file of this
      repository), commit, and run the task again;
   5. add a third subproject with one external dependency and
      `dependencyLocking`, and no lockfile, commit it, and run the task
      again.
4. Scratch SwiftPM check, in a new temporary folder: `swift package init
   --type library`, `git init`, commit, copy the helpers, run
   `setup:deps:swift` with `MISE_PROJECT_ROOT` set to the folder. If
   `/Library/Developer/CommandLineTools` exists, run it again with
   `DEVELOPER_DIR=/Library/Developer/CommandLineTools`.

## Verification

- Lines 1 and 2 exit 0.
- Each skip run exits 0 and prints nothing.
- Scratch Kotlin: step 3 exits 0 and `git status --short` is empty; step 4
  exits 0, `git status --short` is empty, and the lockfile still names
  `31.0-jre`; step 5 exits non-zero, its output names the third subproject,
  and `git status --short` is empty.
- Scratch SwiftPM: each run exits 0.

## Guardrails

- Change no file in the worktree. If a check fails, return `UNRESOLVED:` with
  the failing command, its exit code, the last lines of its output, and the
  unit whose file causes it.
- Remove the temporary folders when done.
- Do not run `mise run setup:configs` without `--dry-run`, and do not run the
  after-landing steps.

## Commit

None when nothing changes. If the orchestrator routes a fix through this
unit: `fix: pass the wave gate of the mobile gap fixes`.
