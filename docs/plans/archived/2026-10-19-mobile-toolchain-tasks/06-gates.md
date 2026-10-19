# U6 — Gates: the wave gate, the skip runs and the scratch Kotlin check

- **Wave:** 4
- **Depends on:** U5
- **Owns:** none. A fix that a gate needs goes back to the unit that owns the
  file, through the orchestrator.
- **Model:** opus
- **Kind:** edit
- **Read first:** index.md sections "Wave gate" and "Gates the orchestrator
  keeps"; `tools/kotlin/.config/mise/tasks/setup/deps/kotlin`.

## Ruling

From index.md, assumed decisions:

> E3 — … A normal run … resolves every resolvable configuration of every
> project … with `--write-locks`, then fails when any `gradle.lockfile` or
> `settings-gradle.lockfile` is new or changed against git — a new module
> without a committed lockfile fails, so the check does not depend on
> `LockMode.STRICT` — and restores every lockfile it changed and removes every
> lockfile it created before it exits.

This plan names no generator. No version file exists to bump. The gates unit
uses `mise x gradle@latest -- gradle init` in a temporary folder only.

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
      kotlin`
      (non-interactive flags as `gradle init --help` gives them), then add a
      second subproject with one external dependency;
   2. add `dependencyLocking { lockAllConfigurations() }` to every project,
      and `distributionSha256Sum` (the real checksum of the wrapper
      distribution) to `gradle/wrapper/gradle-wrapper.properties`;
   3. `./gradlew dependencies --write-locks` for every project, `git init`,
      commit everything;
   4. copy the helpers as in step 2, run `setup:deps:kotlin` with
      `MISE_PROJECT_ROOT` set to the folder;
   5. add a third subproject with one external dependency and no lockfile,
      commit it, run the task again.

## Verification

- Lines 1 and 2 exit 0.
- Each skip run exits 0 and prints nothing.
- Scratch check: step 4 exits 0 and `git status --short` is empty after it;
  step 5 exits non-zero, its output names the new subproject's lockfile, and
  `git status --short` is empty after it.

## Guardrails

- Change no file in the worktree. If a check fails, return `UNRESOLVED:` with
  the failing command, its exit code, the last lines of its output, and the
  unit whose file causes it.
- Remove the temporary folders when done.
- Do not run `mise run setup:configs` without `--dry-run`, and do not run the
  after-landing steps.

## Commit

None when nothing changes. If the orchestrator routes a fix through this
unit: `fix: pass the wave gate of the mobile tool tasks`.
