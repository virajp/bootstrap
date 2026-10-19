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

**BLOCKED**

BLOCKED at wave 0 — preflight: doctor blocking findings (graphify CLI has no
version; no graph reachable); worktree .worktrees/2026-10-20-mobile-gap-fixes

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

| Id | Wave | Unit file                      | Kind   | Owns                                                                                                                                                       | Depends on | Status  | Commit |
| -- | ---- | ------------------------------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------- | ------ |
| U1 | 1    | [01-vwf.md](01-vwf.md)         | edit   | `docs/sync/vwf.md`                                                                                                                                         | —          | pending |        |
| U2 | 1    | [02-kotlin.md](02-kotlin.md)   | edit   | `tools/kotlin/.config/mise/tasks/setup/deps/kotlin`, `docs/sync/kotlin.md`                                                                                 | —          | pending |        |
| U3 | 1    | [03-android.md](03-android.md) | edit   | `tools/android/.config/mise/tasks/setup/deps/android`, `tools/android/.config/mise/tasks/code/lint/android`, `docs/sync/android.md`                        | —          | pending |        |
| U4 | 1    | [04-swift.md](04-swift.md)     | edit   | `tools/swift/.config/mise/tasks/setup/deps/swift`, `docs/sync/swift.md`                                                                                    | —          | pending |        |
| U5 | 2    | [05-review.md](05-review.md)   | review | —                                                                                                                                                          | U2, U3, U4 | pending |        |
| U6 | 3    | [06-docs.md](06-docs.md)       | edit   | `README.md`, `CLAUDE.md`, `docs/sync/index.md`, `docs/sync/migrate.md`, and any other doc that `vwf:docs-sync` finds falsified, except the files U1–U4 own | U1, U5     | pending |        |
| U7 | 4    | [07-gates.md](07-gates.md)     | edit   | none (a fix it needs goes back to the unit that owns the file)                                                                                             | U6         | pending |        |

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

## Run log

| Wave | Unit      | Model  | Round | Outcome | Detail                                                                                                                                                                                                                                                                                            | Commit |
| ---- | --------- | ------ | ----- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 0    | preflight | sonnet | 1     | blocked | wave gate lines 1 and 2 green; format check skipped (no `covers:`); doctor BLOCKING: graphify CLI shim has no version (`mise x -- graphify` → "No version is set for shim: graphify"), no `graphify-out/` graph in the worktree or the main checkout; why: hard halt on a doctor blocking finding | —      |

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-20-mobile-gap-fixes

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
