---
type: vwf-change-plan
title: Close the open gaps of the mobile tool plans
requires:
  - docs/plans/2026-10-19-android-swift-java-tools
  - docs/plans/2026-10-19-mobile-toolchain-tasks
source: []
backlog: []
backlog_pieces: []
gaps_of:
  - docs/plans/2026-10-19-android-swift-java-tools
  - docs/plans/2026-10-19-mobile-toolchain-tasks
---

# Plan — Close the open gaps of the mobile tool plans (2026-10-20)

## Status

**RUNNING**

RUNNING since 2026-10-19T19:56Z in .worktrees/2026-10-20-mobile-gap-fixes

## Consent

| Action                                            | Granted |
| ------------------------------------------------- | ------- |
| Merge to the integration branch and push on green | yes     |
| After landing: archive the plan folder            | run     |
| After landing: archive the two `gaps_of:` folders | run     |
| After landing: `mise run code:graph`              | run     |
| End an `all` run after landing                    | no      |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt, and `run` is the only mode. The integration branch of this
repository is `main`; the run lands with `--no-ff`.

## Release levels

| Project   | Level | Reason                                                                                                                                                                        |
| --------- | ----- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bootstrap | MINOR | New behaviour: Groovy-DSL Android builds, SwiftPM roots with only the Command Line Tools, a Swift root on a Linux host skips, and `remove tools` also deletes Template files. |

## Goal

After this plan lands, each of the 16 open gaps of the two mobile tool plans has
a recorded ruling, and the kotlin, android and swift tasks and
`docs/sync/vwf.md` follow those rulings. The open gaps are G3, G4 and G5 of
`2026-10-19-android-swift-java-tools` (called A-G3 to A-G5 here) and G2 to G14
of `2026-10-19-mobile-toolchain-tasks` (called G2 to G14 here). At hand-off,
both source rows read `0 — moved to` this folder; after a green landing, the
run archives both source folders.

No backlog item asked for this plan. The rulings change four rulings of the
earlier plans. These are plan facts, not standing decisions, so no decisions
doc is necessary:

- mobile E3: the normal run used `--write-locks` and "does not depend on
  `LockMode.STRICT`". Now the task sets STRICT itself (R4).
- mobile E4: an unconditional Xcode check, and a failure on a non-Darwin host
  (R14, R15).
- mobile E6: the lint path came from the folder path (R8).
- android-swift-java-tools D8: `remove` deleted only Synced files (R1).

## Facts the survey established

- No backlog item covers this work (project `bootstrap` under `virajp`, read
  2026-10-20). There is no `docs/memory/decisions/`, no archived plan, no
  `.config/vwf.yaml`, no CI and no test suite.
- `docs/sync/vwf.md` (Remove tools section, :95-107): deletes `Synced` files
  (:102), keeps `Seed` and `Seed, template` (:104), leaves `Generated` to the
  sync (:106), and keeps any other kind, for example `Template` (:107). It
  stops on `remove tools [base]` (:63). It says nothing about the tools that
  only the removed tool's `depends` added; only the Add section computes
  transitive additions (:83-86).
- `repo-sync` adds no header to a rendered `.tmpl`. The header comes from the
  template source. 8 of the 10 `.tmpl` files carry it in line 1 or 2;
  `tools/base/.config/gitleaks.toml.tmpl` and
  `tools/fnox/.config/fnox.toml.tmpl`
  do not. `repo-sync:635` treats `Synced from (bootstrap|repo-config)` in the
  first 3 lines as the marker. The `Template` kind is in `docs/sync/base.md`,
  `github.md:25` and `vscode.md:25`.
- `tools/kotlin/.config/mise/tasks/setup/deps/kotlin` (called KOTLIN):
  - skip without an executable `./gradlew` (:17); checksum check (:31-37);
  - init script (:77-85) registers `bootstrapResolveAll` in `allprojects`;
    it iterates `confs.matching { it.canBeResolved }.each { it.resolve() }`
    (:83), the live view, inside `doLast`;
  - normal run: lockfile snapshot (:125-134), resolve with `--write-locks`
    (:137), compare with git (:141), split into RESOLVED and UNCOMMITTED
    (:146-155), restore (:157-159), hints and exit 1 (:168-183); a build
    with no lockfile warns and exits 0 (:184-187);
  - `--upgrade`: `--write-locks --refresh-dependencies` (:89), then
    `dependencyUpdates` (:108-118); no lockfile warns (:96-97);
  - `print_yellow` at :34, :177, :181; `print_files` at :161-166 (fixed in
    `2333ff1`).
- Status printers in `tools/base/.config/mise/tasks/_scripts/helpers`:
  `print_header` (:163), `print_subheader` (:170), `print_success` (:187),
  `print_warn` (:196), `print_error` (:206), `print_info` (:216), `print_wait`
  (:232). `print_yellow` (:236) is not a status printer.
- `docs/sync/kotlin.md`: only projects with `dependencyLocking` are covered
  (:4, :20-21); normal run (:49-55); lockfiles only with `dependencyLocking`
  (:56-58); `--upgrade` (:59-63); an unlocked project is not checked
  (:83-86). It says nothing on dynamic versions or on configurations that
  fail resolution.
- Gradle (Context7, `/websites/gradle_current_userguide`, dependency locking):
  with lock state, Gradle adds the locked versions as `strictly` constraints,
  so dynamic versions stay pinned. `LockMode.STRICT` also fails when a
  configuration marked as locked has no lock state. `--write-locks`
  overwrites the lock state from a fresh resolve.
- `tools/android/.config/mise/tasks/setup/deps/android` (called SETUP):
  - AGP detect: `grep -rqsE` over `.` with `--include='*.gradle.kts'` (:24-25),
    excludes only `build`, `.gradle`, `.git`; `build-logic/` and catalog grep
    (:27-28);
  - level detect: `find . -name build.gradle.kts` (:47-48), prunes `build`,
    `.gradle`, `.git`, `build-logic`; it walks `.worktrees/`,
    `.claude/worktrees/` and `node_modules`;
  - `;` to `/` in `EMULATOR_IMAGE` (:96-99); image id (:111); platform id
    `platforms/android-<lvl>` (:122);
  - installed check by `$ANDROID_HOME/<id>/package.xml` (:122-136); the
    `_ps16k` swap happens only after a network list (:136-160, :153);
  - missing packages go to `android sdk install` (:172), and the task prints
    "installed" without a check.
- `tools/android/.config/mise/tasks/code/lint/android` (called LINT):
  - `module_of` (:38-55) walks up to the nearest folder with
    `build.gradle.kts` or `build.gradle` (:48) and prints `":${dir//\//:}"`
    (:49); `buildSrc/*`, `build-logic/*`, `gradle/*` and
    `settings.gradle.kts` map to `all` (:35-44);
  - file filter (:67) has `*.gradle.kts` but no `*.gradle`;
  - init script (:90-104) registers `bootstrapLint` in
    `gradle.projectsEvaluated`; it resolves paths with `root.project(it)`
    (:99); "every module" is `root.allprojects` under the root dir (:100);
    it depends on `lint` or `lintFix` only where the task exists (:102);
    invocation (:107-108).
- `docs/sync/android.md`: Kotlin DSL only (:38-39); Flutter `android/` not
  handled (:40-41); `/` ids, `;` accepted (:48-51); offline promise "When all
  four are installed, the task does not list or install, so it needs no
  network." (:61-63); lint-all files (:71-73); every module (:74-77);
  version-catalog example (:88-97).
- `tools/swift/.config/mise/tasks/setup/deps/swift` (called SWIFT):
  - root only (:16); `Package.swift` (:18-19); root `*.xcodeproj` (:20-23);
    neither → exit 0, no output (:25); more than one `.xcodeproj` fails
    (:33-36);
  - non-Darwin → "Swift is macOS only", exit 1 (:38-42);
  - `xcodebuild -version` runs for every Swift root (:46-49), before the
    `Package.swift` block (:51); `swift package resolve` (:58); `-project`
    only (:76, :84).
- `docs/sync/swift.md`: macOS only (:48, :61-62); Xcode necessary (:48-49,
  :63-64); a root `*.xcworkspace` is not read (:67-68).
- Docs that name the mobile tools: `README.md:16-18,46-54`,
  `CLAUDE.md:19-25,161`,
  `docs/sync/index.md:31,33-36,170`, `docs/sync/migrate.md:324-344`. None of
  them names the kotlin, swift or android task behaviour that this plan
  changes.
- Gates: `mise run code` runs dprint, the linter and ShellCheck (`code:lint`),
  and gitleaks, grype and osv-scanner (`code:sec`). The sync dry-run of
  wave-gate line 2 is the line of the earlier plan.
- Commit convention (`.config/git-conventional-commits.yaml`): types `feat`,
  `fix`, `perf`, `refactor`, `revert`, `test`, `ops`, `docs`, `merge`,
  `wip`; `commitScopes: []`; the log uses tool names as scopes.

## Assumed decisions — confirm or override at review

| #   | Decision                    | Ruling                                                                                                                                                                                                                                                                                                                                                                              | Rejected                                                        | Unit |
| --- | --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- | ---- |
| R1  | A-G3: Template outputs      | `remove tools` also deletes the `Template` outputs that the tool's `docs/sync/<tool>.md` lists, but only a file that carries the `Synced from bootstrap` header in its first three lines. A `Template` output without the header stays.                                                                                                                                             | Keep them and document that the agent deletes them by hand      | U1   |
| R2  | A-G4: `remove tools [base]` | The stop on `remove tools [base]` is ratified, because `docs/sync/index.md:138,207` make `base` necessary. No change.                                                                                                                                                                                                                                                               | —                                                               | U1   |
| R3  | A-G5: dependency-only tools | Before the change, the agent finds the tools that the `depends` of the remaining tools no longer need and that the values file does not list. It deletes their `Synced` files and their `Template` outputs with the header, the same way. The values file does not change for them, because they were never listed.                                                                 | Document that their files stay                                  | U1   |
| R4  | G11: lock enforcement       | The init script sets `LockMode.STRICT` on each project. The normal run resolves without `--write-locks`, so Gradle pins the locked versions and fails when a locked configuration has no lock state (a new module). The git compare, the RESOLVED and UNCOMMITTED split and the restore leave the normal run. `--upgrade` does not change (`--write-locks --refresh-dependencies`). | Keep `--write-locks` and document that dynamic versions fail    | U2   |
| R5  | G14: configurations         | The init script resolves only the configurations that are `canBeResolved` and not `canBeConsumed`. A failure in what remains still fails the task.                                                                                                                                                                                                                                  | try/catch and report each failure; out of scope for AGP and KMP | U2   |
| R6  | G6: unlocked project        | A project with no `dependencyLocking` is not checked. Ratified; `docs/sync/kotlin.md:83-86` already says so. The warning for a build with no lockfile stays.                                                                                                                                                                                                                        | Fail the task                                                   | U2   |
| R7  | G10 kotlin residuals        | The init script copies the `matching {}` view to a list before it resolves. The hints use status printers (`print_info` or `print_warn`), not `print_yellow`.                                                                                                                                                                                                                       | —                                                               | U2   |
| R8  | G3, G5: Gradle paths        | `code:lint:android` gives the init script the module folders, not Gradle paths. The init script finds the project whose `projectDir` is that folder. A folder with no project lints every module, which also covers `buildSrc/` (G5).                                                                                                                                               | Keep the folder-to-path mapping and document the limit          | U3   |
| R9  | G7: unknown platform id     | After the install, the task examines `$ANDROID_HOME/<id>/package.xml` for each package it asked for. If one is missing, it fails and names the id. No network.                                                                                                                                                                                                                      | Check the list before the install; accept                       | U3   |
| R10 | G12: Groovy DSL             | Supported: `build.gradle` and `*.gradle` in the AGP detect, the level detect and the lint file filter, and `settings.gradle` in the lint-all list. The `compileSdk` and `targetSdk` detects also accept the Groovy form without `=` (`compileSdk 35`, `compileSdkVersion 35`).                                                                                                      | Kotlin DSL only                                                 | U3   |
| R11 | G8: lint scope              | No AGP check in `code:lint:android`. The departures from E6 are ratified: `:bootstrapLint` from a temporary init script depends on `lint` or `lintFix` only where the task exists, and "every module" means the projects below the Gradle root.                                                                                                                                     | Skip without an AGP plugin                                      | U3   |
| R12 | G10 android residuals       | The AGP and level detects read the file list from `git ls-files --cached --others --exclude-standard`, so ignored folders (`.worktrees/`, `.claude/worktrees/`, `node_modules`) are not read. The installed check also accepts an installed `_ps16k` variant of the image, uses it, and prints which id it used, with no network list.                                              | —                                                               | U3   |
| R13 | G2, G4: ratified departures | `docs/sync/android.md` keeps the version-catalog example (no Flutter example, by E2). The package ids use the `/` form of the `android` CLI, and a `;` in `EMULATOR_IMAGE` is converted. No change.                                                                                                                                                                                 | —                                                               | U3   |
| R14 | G9: Xcode                   | `xcodebuild -version` runs only for a root with an `.xcodeproj`. A SwiftPM-only root resolves with `swift package resolve` and the Command Line Tools. A root `*.xcworkspace` stays unsupported, as `docs/sync/swift.md:67-68` says.                                                                                                                                                | Also read workspaces; keep full Xcode for every root            | U4   |
| R15 | G13: non-Darwin host        | On a non-Darwin host with a Swift root, the task prints its header, then `print_warn` "Swift dependencies are macOS only, skipped", and exits 0.                                                                                                                                                                                                                                    | Keep the failure                                                | U4   |
| R16 | Review row                  | The plan has one review row, U5, because U2, U3 and U4 change shipped shell scripts, which are runnable code.                                                                                                                                                                                                                                                                       | No row (wave review only)                                       | U5   |
| R17 | G10: review loop            | The loop of the earlier review row did not settle. That loop is in the vwf executor in `claude-plugins`, so it is parked here. G10 closes when R7 and R12 land, with G7, G9, G11, G12 and G13 ruled.                                                                                                                                                                                | Change the review loop in this plan                             | —    |

## New dependencies

None in what ships. The gates unit runs `mise x gradle@latest -- gradle init`
in a temporary folder, as the earlier plan did; nothing of it is committed.

## Units

| Id | Wave | Unit file                      | Kind   | Owns                                                                                                                                                       | Depends on | Status  | Commit    |
| -- | ---- | ------------------------------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------- | --------- |
| U1 | 1    | [01-vwf.md](01-vwf.md)         | edit   | `docs/sync/vwf.md`                                                                                                                                         | —          | green   | `7ad3632` |
| U2 | 1    | [02-kotlin.md](02-kotlin.md)   | edit   | `tools/kotlin/.config/mise/tasks/setup/deps/kotlin`, `docs/sync/kotlin.md`                                                                                 | —          | green   | `141d056` |
| U3 | 1    | [03-android.md](03-android.md) | edit   | `tools/android/.config/mise/tasks/setup/deps/android`, `tools/android/.config/mise/tasks/code/lint/android`, `docs/sync/android.md`                        | —          | green   | `b64d96b` |
| U4 | 1    | [04-swift.md](04-swift.md)     | edit   | `tools/swift/.config/mise/tasks/setup/deps/swift`, `docs/sync/swift.md`                                                                                    | —          | green   | `7289d1f` |
| U5 | 2    | [05-review.md](05-review.md)   | review | —                                                                                                                                                          | U2, U3, U4 | pending |           |
| U6 | 3    | [06-docs.md](06-docs.md)       | edit   | `README.md`, `CLAUDE.md`, `docs/sync/index.md`, `docs/sync/migrate.md`, and any other doc that `vwf:docs-sync` finds falsified, except the files U1–U4 own | U1, U5     | pending |           |
| U7 | 4    | [07-gates.md](07-gates.md)     | edit   | none (a fix it needs goes back to the unit that owns the file)                                                                                             | U6         | pending |           |

## Shared-file rule

| File                                                         | Why it collides                  | Owner          |
| ------------------------------------------------------------ | -------------------------------- | -------------- |
| `docs/sync/vwf.md`                                           | shipped instructions for agents  | U1             |
| `docs/sync/kotlin.md`, `android.md`, `swift.md`              | shipped instructions of one tool | U2, U3, U4     |
| `README.md`, `CLAUDE.md`, `docs/sync/index.md`, `migrate.md` | describe tools and tasks         | U6 (docs unit) |
| `tools/*/tool.toml`                                          | not changed by this plan         | nobody         |
| The two `gaps_of:` folders                                   | plan history                     | nobody         |
| Synced copies at the root (`.config/`, `.gitignore`)         | written by `setup:configs` only  | nobody         |
| Version files                                                | none in this repository          | nobody         |

## Waves

- Wave 1: U1, U2, U3 and U4. Each owns the files of one tool or of
  `docs/sync/vwf.md`; the Owns are disjoint.
- Wave 2: U5, the review row, over the commits of wave 1. It covers U2, U3
  and U4; U1 is a doc and the wave review judges it.
- Wave 3: U6, the docs unit.
- Wave 4: U7, the gates unit.

## Wave gate

One command per line; line 1 is the first line of the block, line 2 the second.

```text
mise run code
bash -c 't=$(mktemp -d) && sed "s/^tools *=.*/tools = [\"base\", \"flutter\", \"kotlin\", \"android\", \"java\", \"swift\"]/; s/^environments *=.*/environments = [\"dev\", \"ci\"]/" .config/bootstrap.toml > "$t/values.toml" && REPO_SYNC_CLONE="$PWD" REPO_SYNC_VALUES="$t/values.toml" mise run setup:configs:sync --dry-run'
```

Plus the wave review, plus every report read for `UNRESOLVED:`. Both lines run
from the root of the run's worktree and are green on `main` now. Line 1 runs
ShellCheck over the changed task scripts.

## After landing

| Step                               | Mode | Notes                                                                                                                                                                      |
| ---------------------------------- | ---- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| archive the plan folder            | run  | `plan-management archive`, only when the folder has no open gap                                                                                                            |
| archive the two `gaps_of:` folders | run  | `plan-management archive` for `2026-10-19-android-swift-java-tools` and `2026-10-19-mobile-toolchain-tasks`, each only when it has no open gap that this plan did not take |
| `mise run code:graph`              | run  | refreshes the graphify graph of this repository                                                                                                                            |

## Gates the orchestrator keeps

- **Skip runs** (in U7): in a temporary git repository with no mobile project,
  each of `setup:deps:kotlin`, `setup:deps:swift`, `setup:deps:android` and
  `code:lint:android`, run from the worktree copy with `MISE_PROJECT_ROOT` set
  to the temporary repository, exits 0 and prints nothing.
- **Scratch Kotlin lock check** (in U7), in a temporary Gradle project with two
  locked subprojects:
  - locked and committed: `setup:deps:kotlin` exits 0, and
    `git status --short` is empty;
  - a dynamic version (`com.google.guava:guava:31.+`) whose committed lockfile
    names an older version in the range (`31.0-jre`): the task exits 0, and
    `git status --short` is empty (the lock stays pinned);
  - a third subproject with a dependency, `dependencyLocking`, and no
    lockfile, committed: the task exits non-zero, its output names the new
    subproject, and `git status --short` is empty.
- **Scratch SwiftPM check** (in U7): a temporary git repository with
  `swift package init --type library`, committed: `setup:deps:swift` exits 0.
  If `/Library/Developer/CommandLineTools` exists, the task also exits 0 with
  `DEVELOPER_DIR` set to it.
- No Android scratch run: it installs SDK packages over the network. The review
  row and the wave review are the check for U3.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this file, its owned
paths plus "touch nothing outside this list", the facts section, the shared-file
rule, and the return block below. A unit never bumps a version, never runs a
generator, never edits a doc outside its Owns, never adds a dependency this file
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file in this repository.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

## Out of scope

- A root `*.xcworkspace` in `setup:deps:swift` — declined at R14; it needs a
  scheme choice.
- An AGP check in `code:lint:android` — declined at R11.
- A failure for a project without `dependencyLocking` — declined at R6.
- An Android scratch run — declined at the gate question; it needs the network
  and several GB of SDK packages.
- Edits to the two `gaps_of:` folders — their gap lists are plan history.

## Parked

- G10 of `2026-10-19-mobile-toolchain-tasks`: the review loop of its U4 did
  not settle (round 1 had 10 findings, round 2 had 10, and the convergence
  guard ended the loop). The loop is in the vwf executor in `claude-plugins`
  (`skills/execute/references/review-unit.md`). Examine there why fresh
  `/code-review` rounds keep the count level, for a later plan in that repo.

## Gaps surfaced during execution

- G1 (U2, non-blocking): the init script sets `LockMode.STRICT` before the build
  script runs, so a build script that sets its own `lockMode` (for example
  `LENIENT`) overrides it. The plan does not say which wins. Assumption taken:
  an explicit lock mode in the build is the build owner's choice.
- G2 (U3, non-blocking): R12 makes `git ls-files` the file source of the android
  detects, and the plan does not cover a Gradle root that is not a git work
  tree. Assumption taken: `git ls-files` lists nothing there, so
  `setup:deps:android` does not apply and prints nothing.
- G3 (R1, non-blocking): the comment at `tools/swift/tool.toml:1` ("Swift and
  SwiftPM from the host Xcode (macOS only).") is false after R14 and R15. No
  unit owns the file (Shared-file rule: `tools/*/tool.toml` not changed), and U6
  must not touch `tools/**`. It stays as it is.
- G4 (U5, non-blocking, ruling question on R8 and U3 Edit 4): `module_of` maps a
  file at the Gradle root to `.`, the root project. In a multi-module build the
  root has no `lint` task, so a change to the root `build.gradle(.kts)`,
  `lint.xml` or `lint-baseline.xml` lints nothing; the code before this plan
  linted every module. The rulings do not say whether root-level build and lint
  files lint every module. Kept as Edit 4 says.
- G5 (U5, non-blocking, ruling question on R5): the filter `canBeResolved &&
  !canBeConsumed` skips configurations made with
  `configurations.create()`/`creating`, which Gradle 8 makes resolvable and
  consumable by default (detekt, ktlint, custom tool configurations). STRICT
  never checks their locks and `--upgrade` never writes them. Kept as R5 says.
- G6 (U5, non-blocking, ruling question on R12): the installed `_ps16k` shortcut
  also replaces an explicit `EMULATOR_IMAGE` at any API level, so the 4 KB image
  the user asked for is never installed. R12 does not say whether the shortcut
  applies to an explicit image. Kept as R12 says.
- G7 (U5, non-blocking, ruling question on R4 and U2 Edit 2): the normal run
  reports every `./gradlew` failure as a lockfile mismatch and advises
  `--upgrade` — a network outage, a 401, a compile error or a JDK mismatch
  included. After a transient failure that advice rewrites every lockfile with
  `--refresh-dependencies`. The plan does not separate a lock failure from other
  Gradle failures. Kept as Edit 2 says.

## Run log

| Wave | Unit            | Model  | Round | Outcome     | Detail                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Commit    |
| ---- | --------------- | ------ | ----- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| 0    | preflight       | sonnet | 1     | blocked     | wave gate lines 1 and 2 green; format check skipped (no `covers:`); doctor BLOCKING: graphify CLI shim has no version (`mise x -- graphify` → "No version is set for shim: graphify"), no `graphify-out/` graph in the worktree or the main checkout; why: hard halt on a doctor blocking finding                                                                                                                                                                         | —         |
| 0    | preflight       | sonnet | 2     | pass        | resume after `30b9e8f` (graphify in mise) merged from main at `9c283db`; doctor BLOCKING none, LSP none (no `.config/vwf.yaml`); non-blocking: no `.graphifyignore`, no `mempalace.yaml`, no `post-merge` hook type; wave gate lines 1 and 2 green; stack conventions skipped (no `code` unit); order: W1 U1,U2,U3,U4 → W2 U5 → W3 U6 → W4 U7                                                                                                                             | —         |
| 1    | U1 remove tools | opus   | 1     | pass        | edit; DECIDED: step 1 cites the Add closure rule; step 4 names dependency-only tools; Template path is the output path. GAP: assumed every tool has a `docs/sync/<tool>.md` ships table                                                                                                                                                                                                                                                                                   | —         |
| 1    | U4 swift        | opus   | 1     | pass        | edit; DECIDED: SwiftPM-only check runs `swift --version` (the `/usr/bin/swift` stub is always on PATH); both `Package.swift` and `.xcodeproj` → Xcode check only; multi-`.xcodeproj` check stays before the host check. Linux simulated with stub `uname`: header + warn, exit 0                                                                                                                                                                                          | —         |
| 1    | U2 kotlin       | opus   | 1     | pass        | edit; DECIDED: one init script, STRICT line added only without `--upgrade`; `lockMode.set(LockMode.STRICT)` in `allprojects` (Context7: Property since Gradle 6.4); no-lockfile check after the resolve. GAP: G1 (build-script `lockMode` overrides STRICT)                                                                                                                                                                                                               | —         |
| 1    | U3 android      | opus   | 1     | pass        | edit; DECIDED: one `git ls-files -z` list, `build/` and `.gradle/` still skipped; level pattern already matched Groovy forms; root-level file lints every module; folders passed one per line; Context7 confirmed `projectsEvaluated`, `allprojects`, `projectDir`. GAP: G2 (non-git Gradle root); R9 and the init script not run end to end (no SDK install, Gradle runs are U7's)                                                                                       | —         |
| 1    | R1              | opus   | 1     | findings(7) | RULINGS: U4 departed from R15 (multi-`.xcodeproj` fails on non-Darwin); U4 swift.md:26 Xcode-only text; U2 kotlin.md:42 + task:124 "every configuration"; U2 kotlin.md:52 "committed lockfile"; U3 android.md:13 Groovy relevance; U3 android.md:39,40,84 fold; U3 module_of prints `all`, not `.` (Edit 4). tool.toml:1 → G3. CONTRACT clean                                                                                                                             | —         |
| 1    | U4 swift        | opus   | 2     | pass        | edit; R1 loop-back: host check before the multi-`.xcodeproj` check (non-Darwin: warn, exit 0); swift.md:3,:26,:48 and multi-project text fixed. DECIDED: multi-`.xcodeproj` failure stays on macOS only                                                                                                                                                                                                                                                                   | —         |
| 1    | U2 kotlin       | opus   | 2     | pass        | edit; R1 loop-back: "resolvable-only configurations" in kotlin.md table and task success line; "lockfiles on the disk", "without a lockfile". DECIDED: same fix in `#MISE description`, cache reason and doc intro; "commit the lockfiles" hints kept                                                                                                                                                                                                                     | —         |
| 1    | U3 android      | opus   | 2     | pass        | edit; R1 loop-back: `module_of` prints `.` for the root (Edit 4); android.md relevance names Groovy DSL; lint paragraph and refold to 80 columns. DECIDED: a file under a folder with no build file maps to `.`. Note: a root module folder named `all` collides with the lint-all marker (lints too much) — a code finding for U5, not a plan gap                                                                                                                        | —         |
| 1    | R1              | opus   | 2     | pass        | wave review round 2: FINDINGS 0, CONTRACT clean, RULINGS clean                                                                                                                                                                                                                                                                                                                                                                                                            | —         |
| 1    | gate            | —      | 1     | pass        | wave gate lines 1 and 2 exit 0 (57/57 in sync); no `UNRESOLVED:`; commits U1 `7ad3632`, U2 `141d056`, U3 `b64d96b`, U4 `7289d1f`                                                                                                                                                                                                                                                                                                                                          | —         |
| 2    | U5              | opus   | 1     | pass        | security; range `30b9e8f..0f9c230`; files: kotlin task+doc (U2), android tasks+doc (U3), swift task+doc (U4), vwf.md (U1, uncovered); engine `/security-review` no findings (engine/U5-1-security.log); reviewer FINDINGS none, VERDICT approve; tag U5/security/1                                                                                                                                                                                                        | —         |
| 2    | U5              | opus   | 1     | findings(8) | review; range `30b9e8f..0f9c230`; engine `/code-review` 10 findings (engine/U5-1-review.log); reviewer 8 findings, SPEC COMPLIANCE met, API COMPAT n/a, VERDICT approve; ruling questions → G4 (R8 root lint), G5 (R5 tool configs), G6 (R12 explicit image); loop back U2: kotlin:119 non-git false warning; U3: android:37 build-logic arm, :203 ANDROID_HOME, lint:81 `all` collision, android:31 pathspecs; 0 findings on uncovered units; tags U5/review/1, U5/gap/1 | —         |
| 2    | U2 kotlin       | opus   | 3     | pass        | edit; U5 r1 loop-back: `lockfiles()` uses `find` on the disk (skips hidden, `build`, `node_modules`), so a non-git root or ignored lockfiles no longer warn falsely. Note: `--upgrade` "did not change" check still uses git (not named by the finding)                                                                                                                                                                                                                   | `7a276fb` |
| 2    | U3 android      | opus   | 3     | pass        | edit; U5 r1 loop-back: `git ls-files` pathspecs; every `build-logic/` file in the AGP detect (level detect `grep -I`); fail after the header when `ANDROID_HOME` is unset; lint-all marker `/` (no folder collision). DECIDED: unset `ANDROID_HOME` is an error, not "does not apply"                                                                                                                                                                                     | `b7ecfd4` |
| 2    | U5              | opus   | 2     | pass        | security; range `30b9e8f..2919bfe`; engine `/security-review` no findings (engine/U5-2-security.log); reviewer FINDINGS none, VERDICT approve; tag U5/security/2                                                                                                                                                                                                                                                                                                          | —         |
| 2    | U5              | opus   | 2     | findings(3) | review; range `30b9e8f..2919bfe`; engine `/code-review` 10 findings (engine/U5-2-review.log), 6 already recorded as G1, G2, G4–G6 or ruled by R15; reviewer 3 findings, SPEC COMPLIANCE met, VERDICT approve; guard: 8 → 3, none resurfaced; ruling question → G7 (every Gradle failure reported as a lock mismatch); loop back U2: kotlin:67 STRICT on buildscript locks, kotlin:84-93 + kotlin.md:65 `--upgrade` summary uses git; tags U5/review/2, U5/gap/2           | —         |
| 2    | U2 kotlin       | opus   | 4     | pass        | edit; U5 r2 loop-back: STRICT also on `p.buildscript.dependencyLocking` (Context7: `ScriptHandler.getDependencyLocking()`); `--upgrade` summary compares `cksum` snapshots of `lockfiles()`, no git; kotlin.md agrees                                                                                                                                                                                                                                                     | `7637767` |

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-20-mobile-gap-fixes

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
