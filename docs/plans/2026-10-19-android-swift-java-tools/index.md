---
type: vwf-change-plan
title: Add the java, android and swift tools and docs/sync/vwf.md
requires: []
source: [ B11, B03 ]
backlog: [ B11 ]
backlog_pieces: [ B03 ]
gaps_of: []
---

# Plan — Add the java, android and swift tools and docs/sync/vwf.md (2026-10-19)

## Status

**RUNNING**

RUNNING since 2026-10-19 in .worktrees/2026-10-19-android-swift-java-tools

## Consent

| Action                                            | Granted |
| ------------------------------------------------- | ------- |
| Merge to the integration branch and push on green | yes     |
| After landing: archive the plan folder            | run     |
| After landing: `mise run code:graph`              | run     |
| End an `all` run after landing                    | no      |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt, and `run` is the only mode. The integration branch of this
repository is `main`; the run lands with `--no-ff`.

## Release levels

| Project   | Level | Reason                                                                                                                                         |
| --------- | ----- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| bootstrap | MINOR | Three new tools and a new entry point for vwf and stackgen. No consumer breaks: a flutter consumer gets the same mise tools through `depends`. |

## Goal

After this plan lands, `bootstrap` `main` has `docs/sync/vwf.md` (the one entry
point for vwf and stackgen), an `android` tool and a `swift` tool, and the mise
tool `java` has exactly one owner tool. The `virajp/claude-plugins` plan
`docs/plans/2026-10-18-packs-hand-off-to-bootstrap` then passes its first
wave-gate line, which checks that the three files `docs/sync/vwf.md`,
`tools/android/tool.toml` and `tools/swift/tool.toml` are on `main`.

This plan is plan 1 of 2 for backlog item B03 (move mobile toolchain setup into
bootstrap). It finishes B11 and lands one piece of B03. Plan 2,
`docs/plans/2026-10-19-mobile-toolchain-tasks`, requires this plan and finishes
B03 (see Parked). No standing decision is reversed.

## Facts the survey established

- `tools/` has 17 tools: `base`, `bun`, `claude`, `cloudflare`, `docker`,
  `doppler`, `flutter`, `fnox`, `gcloud`, `github`, `gitlab`, `kotlin`, `node`,
  `pulumi`, `python`, `rust`, `vscode`. No `android`, `swift`, `java` or
  `xcode` tool exists.
- `tools/flutter/tool.toml`:
  - keys `depends = []`, `extensions = ["dart-code.dart-code",
    "dart-code.flutter"]`,
    `root_only = false`, `seed = []`;
  - mise tools, all `envs = ["dev", "ci"]` (lines 8-19): `android-sdk`
    version `"23"` with mise `depends = ["java"]`; `flutter` version `"3"`;
    `java` version `"21"`;
  - no `[env]`, no `[mise_env]`, no `[[precommit]]`.
- `tools/flutter` ships the tasks `setup/deps/flutter`, `code/format/dart` and
  `code/lint/dart`. `setup/deps/flutter` reads `${ANDROID_HOME:-}` (:62) and
  runs `mise x -- android sdk install platform-tools` when `android/gradlew`
  exists (:64-70). This plan does not change these tasks.
- `tools/kotlin/tool.toml`:
  - keys `depends = []`, `extensions = []`, `root_only = false`, `seed = []`;
  - one mise tool `http:kotlin-lsp`, version `latest`, dev only, macOS
    platforms only (lines 8-15).
- `java` is defined only in `tools/flutter/tool.toml` (lines 17-19).
  `android-sdk` is defined only in `tools/flutter/tool.toml` (lines 8-11).
- The mise registry maps `android-sdk` to `vfox:jdx/vfox-android-sdk`
  (versions 21.0, 22.0, 23.0). The plugin exports `ANDROID_HOME`,
  `ANDROID_SDK_ROOT` and the `cmdline-tools/<v>/bin` and `platform-tools`
  `PATH` entries. A tool that ships `android-sdk` needs no `[env]` table for
  them.
- `mise latest java@lts` resolves to `25.0.4+101.0.LTS`.
  `tools/node/tool.toml:19`
  already uses `version = "lts"`.
- repo-sync (`tools/base/.config/mise/tasks/_scripts/repo-sync`):
  - `_rs_mise_index` (:187-209) indexes every tool in the clone and fails with
    exit 2 when two tools define the same mise tool.
  - `_rs_resolve` (:224-) adds `depends` depth-first, exits 3 for a tool that
    is not in bootstrap, and errors on a mise `depends` that no tool defines.
  - `_rs_mise_env` (:443-): a tool listed in `tools` uses its `envs`; a tool
    added only as a dependency is dev only. Each mise `depends` of a mise tool
    is in the environments of that mise tool too.
- Each `docs/sync/<tool>.md` has a one-line description, a Property | Value
  table (Root only, Depends, Seeds), then these headings in this order:
  `## When the tool is relevant`, `## What the tool ships`, `## mise tools`,
  `## Tasks`, `## Pre-commit hooks`, `## Files the repository supplies`,
  `## Adapt to the repository`, `## Migration`. `docs/sync/flutter.md` and
  `docs/sync/kotlin.md` are the models.
- `docs/sync/index.md`: the files table is at :24-30, the tool list at :32-36
  (one sentence, alphabetical), the `tool.toml` key table at :143-150, the
  mise and environment text at :157-182.
- `README.md:44-60` shows a `tool.toml` example where `flutter` defines
  `android-sdk` with `depends = ["java"]`, and a `[mise_env.ci.env]`
  `ANDROID_HOME`.
- `CLAUDE.md` "Layout rule" lists the files of `docs/sync/`: `index.md`,
  `migrate.md`, `<tool>.md`. It does not list `vwf.md`.
- `docs/sync/vwf.md` does not exist. vwf is not mentioned in `README.md` or
  `docs/`.
- The contract that `virajp/claude-plugins` expects of `docs/sync/vwf.md`
  (packs plan, decision 8, `index.md:152`): a plugin shallow-clones bootstrap
  (`git clone --depth 1 git@github.com:virajp/bootstrap.git` into a
  `mktemp -d` folder), reads `docs/sync/vwf.md` in the clone, and gives it
  instructions — `bootstrap: add tools [<tool>, …]` or
  `bootstrap: remove tools [<tool>, …]`, plus `reason: <why>`. The file then
  follows `docs/sync/<tool>.md` of each tool. A plugin never edits
  `.config/bootstrap.toml` itself. The plugins name this path and no other
  bootstrap path. Plugins may run `mise run setup:configs`. The materializer's
  reason string is `stackgen pack <type>/<slug> pinned`.
- The `claude-plugins` checker pins the tool list to the 17 tools plus
  `android` and `swift`. `java` is not in it; the packs never name `java`
  directly, so the new tool does not break the checker.
- Gates: no CI workflow, no test suite. `mise run code` runs `code:format`
  (check), `code:lint` and `code:sec`; it exited 0 on 2026-10-19. The sync
  dry-run of wave-gate line 2 exited 0 on 2026-10-19 with
  `tools = ["base", "flutter", "kotlin"]`; it reports "would update" for the
  flutter and kotlin files, because this repository selects only `base`, and
  writes nothing.
- Commit convention (`.config/git-conventional-commits.yaml`): types `feat`,
  `fix`, `perf`, `refactor`, `revert`, `test`, `ops`, `docs`, `merge`, `wip`;
  `commitScopes: []`. The log uses tool names as scopes (`fix(flutter):`).

## Assumed decisions — confirm or override at review

| #   | Decision                        | Ruling                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | Rejected                                                                                                                                                        | Unit |
| --- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---- |
| D1  | Scope                           | Two plans. This plan finishes B11 and lands one piece of B03. Plan 2 requires this plan and finishes B03.                                                                                                                                                                                                                                                                                                                                                                                                           | One plan for B11 and B03 (keeps claude-plugins blocked longer); only B11 now                                                                                    | —    |
| D2  | Owner of mise `java`            | A new `java` tool. `tools/java/tool.toml` defines only `[mise.tools.java]`, `version = "lts"`, `envs = ["dev", "ci"]`.                                                                                                                                                                                                                                                                                                                                                                                              | java in the `android` tool (a Kotlin repo without Android must select android); java in the `kotlin` tool (Flutter pulls kotlin-lsp); `temurin-25`; `21`        | U1   |
| D3  | The `android` tool              | A new `android` tool defines `[mise.tools.android-sdk]`, `version = "23"`, `envs = ["dev", "ci"]`, mise `depends = ["java"]` — moved unchanged from `tools/flutter/tool.toml`. No `[env]` table: the plugin exports `ANDROID_HOME`, `ANDROID_SDK_ROOT` and `PATH`.                                                                                                                                                                                                                                                  | Keep `android-sdk` in `flutter`                                                                                                                                 | U1   |
| D4  | Flutter and Android             | `tools/flutter/tool.toml` drops `[mise.tools.android-sdk]` and `[mise.tools.java]`, keeps `depends = []`, and gives `[mise.tools.flutter]` the mise `depends = ["android-sdk"]`. The sync then adds the `android` tool, and each mise `depends` is in the environments of `flutter` (dev and ci), so `android-sdk` and `java` are in dev and ci. Ruled 2026-10-19 (G2): "`android`, `java`, `android-sdk`: all must support `dev` as well as `ci` since they are required to build the app".                        | Tool-level `depends = ["android"]` (a dependency-only tool is dev only, so CI loses `android-sdk` and `java`); a sync-rule change; a doc item to list `android` | U1   |
| D5  | Kotlin and Java                 | `tools/kotlin/tool.toml` sets `depends = ["java"]`. Under the sync rule a tool added only as a dependency is dev only, so `docs/sync/kotlin.md` "Adapt to the repository" tells the agent to add `java` to `tools` when CI builds with Gradle.                                                                                                                                                                                                                                                                      | Add the dependency in plan 2                                                                                                                                    | U1   |
| D6  | The `swift` tool                | `tools/swift/tool.toml` has no mise tool, `depends = []`, `extensions = ["swiftlang.swift-vscode"]`, `root_only = false`, `seed = []`. `docs/sync/swift.md` says: macOS only; Swift and `sourcekit-lsp` come from the Xcode that `xcode-select` selects; no Linux toolchain. Its tasks come in plan 2.                                                                                                                                                                                                              | Manifest only (empty lists); also the `setup:deps:swift` task now                                                                                               | U1   |
| D7  | Instructions of `vwf.md`        | `docs/sync/vwf.md` accepts `bootstrap: add tools [<tool>, …]` and `bootstrap: remove tools [<tool>, …]`, each with `reason: <why>`. For each tool it follows `docs/sync/<tool>.md`. A plugin never edits `.config/bootstrap.toml` itself; the agent that reads `vwf.md` does.                                                                                                                                                                                                                                       | —                                                                                                                                                               | U2   |
| D8  | `remove tools` and files        | The agent removes the tool from `tools`, deletes each file that `docs/sync/<tool>.md` lists as Synced (they carry the bootstrap header), keeps each Seed (the repository owns it), then runs `mise run setup:configs`, which removes the stale generated mise files.                                                                                                                                                                                                                                                | Keep all files (orphans); ask the user per file                                                                                                                 | U2   |
| D9  | Consent and commit              | The instruction is the consent: the plugin asked the user before it sent it. The agent applies the change, runs `mise run setup:configs`, shows the changed files and the reason, and does not commit. The caller commits.                                                                                                                                                                                                                                                                                          | Ask the user first; apply and commit                                                                                                                            | U2   |
| D10 | Stop conditions                 | The agent stops and changes nothing when: the repository has no `.config/bootstrap.toml` (it says the repository is not adopted, and that `docs/sync/index.md` is the adoption); a tool name is not a folder in `tools/` (it names the tool); `remove` names a tool that another selected tool needs through `depends` or a mise `depends` (it names that tool). `add` of a tool that is already in `tools` changes nothing for that tool. `remove` of a tool that is not in `tools` changes nothing for that tool. | —                                                                                                                                                               | U2   |
| D11 | `bootstrap: adopt`              | Not in this plan. `vwf.md` documents only `add tools`, `remove tools` and `reason:`.                                                                                                                                                                                                                                                                                                                                                                                                                                | Document `bootstrap: adopt` now                                                                                                                                 | U2   |
| D12 | Owner of the per-tool sync docs | `docs/sync/<tool>.md` and `docs/sync/vwf.md` are shipped instructions that agents run, so the edit units own them. The docs unit owns `README.md`, `CLAUDE.md`, `docs/sync/index.md` and `docs/sync/migrate.md`.                                                                                                                                                                                                                                                                                                    | Every doc in the docs unit                                                                                                                                      | all  |
| D13 | Review row                      | None. The plan lands TOML data and Markdown, no script; the wave review is the only check.                                                                                                                                                                                                                                                                                                                                                                                                                          | One review row                                                                                                                                                  | —    |
| D14 | Model                           | Every unit runs on `opus`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | —                                                                                                                                                               | all  |

## New dependencies

None. `java` and `android-sdk` come from the mise registry and are in use today.

## Units

| Id | Wave | Unit file                  | Kind | Owns                                                                                                                                                                                                                                              | Depends on | Status  | Commit  |
| -- | ---- | -------------------------- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------- | ------- |
| U1 | 1    | [01-tools.md](01-tools.md) | edit | `tools/java/tool.toml`, `tools/android/tool.toml`, `tools/swift/tool.toml`, `tools/flutter/tool.toml`, `tools/kotlin/tool.toml`, `docs/sync/java.md`, `docs/sync/android.md`, `docs/sync/swift.md`, `docs/sync/flutter.md`, `docs/sync/kotlin.md` | —          | green   | da0e607 |
| U2 | 1    | [02-vwf.md](02-vwf.md)     | edit | `docs/sync/vwf.md`                                                                                                                                                                                                                                | —          | green   | 632e820 |
| U3 | 2    | [03-docs.md](03-docs.md)   | edit | `README.md`, `CLAUDE.md`, `docs/sync/index.md`, `docs/sync/migrate.md`, and any other doc that `vwf:docs-sync` finds falsified, except the files U1 and U2 own                                                                                    | U1, U2     | green   | 48d791c |
| U4 | 3    | [04-gates.md](04-gates.md) | edit | none (a fix it needs goes back to the unit that owns the file)                                                                                                                                                                                    | U3         | pending |         |

## Shared-file rule

| File                                                                     | Why it collides                         | Owner          |
| ------------------------------------------------------------------------ | --------------------------------------- | -------------- |
| `docs/sync/index.md`                                                     | lists every tool and every sync file    | U3 (docs unit) |
| `README.md`, `CLAUDE.md`, `docs/sync/migrate.md`                         | describe tools and the sync folder      | U3 (docs unit) |
| `docs/sync/<tool>.md` of `java`, `android`, `swift`, `flutter`, `kotlin` | shipped instructions of the tools of U1 | U1             |
| `docs/sync/vwf.md`                                                       | the entry point                         | U2             |
| Version files                                                            | none in this repository                 | nobody         |
| Synced copies at the root (`.config/`, `.gitignore`)                     | written by `setup:configs` only         | nobody         |

## Waves

- Wave 1: U1 and U2. Their Owns are disjoint. `vwf.md` names tools only through
  the generic rule `docs/sync/<tool>.md`, so U2 needs nothing that U1 writes.
- Wave 2: U3, the docs unit, after both.
- Wave 3: U4, the gates unit.

## Wave gate

One command per line; line 1 is the first line of the block, line 2 the second.

```text
mise run code
bash -c 't=$(mktemp -d) && sed "s/^tools *=.*/tools = [\"base\", \"flutter\", \"kotlin\"]/; s/^environments *=.*/environments = [\"dev\", \"ci\"]/" .config/bootstrap.toml > "$t/values.toml" && REPO_SYNC_CLONE="$PWD" REPO_SYNC_VALUES="$t/values.toml" mise run setup:configs:sync --dry-run'
```

Plus the wave review, plus every report read for `UNRESOLVED:`. Both lines run
from the root of the run's worktree. Line 2 syncs from the worktree itself
(`REPO_SYNC_CLONE`) into a dry-run, so it writes nothing; it fails on a mise
tool that two tools define (exit 2) or on an unknown tool (exit 3). Both lines
were green on 2026-10-19.

## After landing

| Step                    | Mode | Notes                                                           |
| ----------------------- | ---- | --------------------------------------------------------------- |
| archive the plan folder | run  | `plan-management archive`, only when the folder has no open gap |
| `mise run code:graph`   | run  | refreshes the graphify graph of this repository                 |

## Gates the orchestrator keeps

- **Full-selection dry-run** (in U4): the line-2 command with
  `tools = ["base", "flutter", "kotlin", "android", "java", "swift"]`. Pass:
  exit 0; the output lists `.config/mise/conf.d/java/mise.dev.toml`,
  `.config/mise/conf.d/java/mise.ci.toml`,
  `.config/mise/conf.d/android/mise.dev.toml` and
  `.config/mise/conf.d/android/mise.ci.toml`; the rendered
  `conf.d/flutter/mise.*.toml` hold no `android-sdk` and no `java`.
- **The claude-plugins gate line** (after the landing pushes `main`, by hand
  or by the next claude-plugins run). This command exits 0:

  ```text
  bash -c 'for p in docs/sync/vwf.md tools/android/tool.toml tools/swift/tool.toml; do gh api "repos/virajp/bootstrap/contents/$p" --silent || { echo "bootstrap main lacks $p"; exit 1; }; done'
  ```

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this file, its owned
paths plus "touch nothing outside this list", the facts section, the shared-file
rule, and the return block below. A unit never bumps a version, never runs a
generator, never edits a doc outside its Owns, never adds a dependency this file
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

## Out of scope

- `bootstrap: adopt` in `vwf.md` — the user declined it for this plan (D11);
  the `claude-plugins` vwf hand-off plan needs it later (see Parked).
- B04 gate tools (ktlint, detekt, swiftlint, swift-format) — a separate backlog
  item.
- `test:*` tasks (`test:golden`, `test:e2e`) — they stay in `claude-plugins`,
  because they need an LLM decision.
- Any change to the tasks of `flutter` — plan 2.

## Parked

- B03: the rest of the mobile toolchain move, in plan 2,
  `docs/plans/2026-10-19-mobile-toolchain-tasks`:
  - the kotlin and Gradle deps tasks, with the wrapper checksum check and the
    lock check for a new module (K);
  - the Swift and SwiftPM deps task;
  - the Android setup and lint tasks, where the lint trigger misses `lint.xml`
    and `lint-baseline.xml` and the lint must run only on the changed modules
    (G8);
  - the `EMULATOR_IMAGE` `_ps16k` check (G7);
  - the SDK-level values read from `build-logic/` and
    `gradle/libs.versions.toml`;
  - the exact cmdline-tools pin (G1);
  - a home for `XCODE_VERSION` and `SIMULATOR_*`.
- `bootstrap: adopt` in `docs/sync/vwf.md`: the `claude-plugins` plan
  `2026-10-18-vwf-hands-off-to-bootstrap` checks for it in its wave gate
  (`index.md:181-182`). It needs its own backlog item and plan.

## Gaps surfaced during execution

- G1 (resolved 2026-10-19 by ruling (a), `9a3f5e8` on `main`, merged into the
  run branch as `cd03113`): wave-gate line 2 cannot run from the run's worktree.
  `tools/base/.config/mise/tasks/_scripts/repo-sync:850` refuses a
  `REPO_SYNC_CLONE` whose `.git` is not a directory, and `.git` is a file in a
  linked worktree. The line was green on 2026-10-19 because it ran in the main
  checkout. Ruling needed: one of (a) fix `repo-sync` to accept a `.git` file
  (`-e` in place of `-d` at :850 and :858) in its own change, landed on `main`
  first; (b) change wave-gate line 2 in this folder so that it gives
  `REPO_SYNC_CLONE` a path with a real `.git` directory that holds the
  uncommitted wave edits; (c) another gate that the folder states.
- G2 (resolved 2026-10-19: mise `depends = ["android-sdk"]` on
  `[mise.tools.flutter]`, D4 amended): with `flutter` in `tools`, `android` is a
  dependency only, so under the `_rs_mise_env` rule `android-sdk` and `java`
  render to dev only. A flutter consumer with `ci` loses both in CI, which it
  has today. This contradicts the Release-levels reason "a flutter consumer gets
  the same mise tools through `depends`". D4 gives no CI item for flutter (D5
  gives one for kotlin). Ruling needed, one of: (a) a `docs/sync/flutter.md`
  "Adapt to the repository" item, and a `docs/sync/migrate.md` item for U3, that
  tell the agent to add `android` to `tools` when CI builds Android; (b) change
  `_rs_mise_env` so that a dependency tool takes the environments of the tool
  that depends on it (a `tools/base` change, outside this plan's Owns); (c)
  accept dev only and correct the Release-levels reason.
- G3 (non-blocking, U2): D8 does not name the kind `Template`.
  `docs/sync/vwf.md` follows D8 literally, so `remove` deletes only `Synced`
  files, and the `Template` outputs of a removed tool (for example in `base`,
  `github`, `vscode`) stay in the consumer. A ruling is needed only if they must
  be deleted too.
- G4 (non-blocking, U2): `docs/sync/vwf.md` stops on `remove tools [base]`. D10
  does not list it; assumed because `docs/sync/index.md:138,207` make `base`
  necessary.
- G5 (non-blocking, U2): `remove` does not take out of `tools` the tools that
  only the removed tool's `depends` added. They drop from the next selection;
  their synced task files stay, because the sync never deletes files.

## Run log

| Wave | Unit                   | Model | Round | Outcome     | Detail                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Commit |
| ---- | ---------------------- | ----- | ----- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| 0    | preflight              | —     | 1     | pass        | node: doctor — no blocking finding (repo has no `.config/vwf.yaml`: "not onboarded", information only); edit units only, so no LSP or conventions read                                                                                                                                                                                                                                                                                                                                                                                                 | —      |
| 0    | preflight              | —     | 1     | pass        | node: wave-gate line 1 `mise run code` — exit 0                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | —      |
| 0    | preflight              | —     | 1     | blocked     | node: wave-gate line 2 — exit 5, `repo-sync: cannot use the clone <worktree> - nothing synced`; why: `repo-sync:850` tests `-d $REPO_SYNC_CLONE/.git`, and `.git` is a file in a linked worktree. Inherited from `main`, not caused by a unit (G1)                                                                                                                                                                                                                                                                                                     | —      |
| 0    | preflight              | —     | 2     | pass        | node: resume — `main` merged in (`9a3f5e8`, repo-sync accepts a `.git` file); wave-gate line 1 exit 0; line 2 exit 0, 49/49 in sync                                                                                                                                                                                                                                                                                                                                                                                                                    | —      |
| 1    | U2 docs/sync/vwf.md    | opus  | 1     | pass        | node: edit — DECIDED: remove also deletes `Template` files (they carry the header), keeps files without it; `remove base` stops; all instructions checked before any change; failed sync points to the exit-code table. GAP: remove leaves tools added only by the removed tool's `depends` to drop from the next selection. DOCS FALSIFIED: `docs/sync/index.md` files table, `CLAUDE.md` Layout rule (both U3)                                                                                                                                       | —      |
| 1    | U1 tools and tool docs | opus  | 1     | blocked     | node: edit — UNRESOLVED: flutter CI regression. With `flutter` in `tools`, `android` is a dependency only, so `android-sdk` and `java` render to dev only and a flutter consumer with `ci` loses them in CI; contradicts the Release-levels reason. Verification otherwise green (gate lines 1-2, full-selection dry-run 53/53). why: ruling needed; edits set aside as `/tmp/vwf-run/U1-round1.patch`, not committed                                                                                                                                  | —      |
| 2    | U3 docs                | —     | —     | skipped     | why: depends on U1                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | —      |
| 3    | U4 gates               | —     | —     | skipped     | why: depends on U3, which depends on U1                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | —      |
| 1    | R1 wave review         | opus  | 1     | findings(4) | U2: add step misses tools added through a mise `depends` (:83); kind `Seed, template` not handled (:101); failed-sync pointer only in add, and unnamed (:90); `remove base` stop not ruled (:63). RULINGS: U2 departed from D8 (deletes `Template` files too, adds keep-and-report). CONTRACT clean                                                                                                                                                                                                                                                    | —      |
| 1    | U2 docs/sync/vwf.md    | opus  | 2     | pass        | node: edit — remove now deletes only `Synced` (D8), keeps `Seed` and `Seed, template`; add follows mise `depends` too; failed-sync pointer moved to "After the change" (DECIDED). GAP: `Template` files of a removed tool stay (D8 literal); `remove base` stop not in D10, kept because `docs/sync/index.md:138,207` make base necessary; dependency-only tools drop at next selection. Its UNRESOLVED was the orchestrator's unformatted Run log, fixed by `dprint fmt`                                                                              | —      |
| 1    | R1 wave review         | opus  | 2     | pass        | FINDINGS 0; CONTRACT clean; RULINGS clean                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | —      |
| 0    | preflight              | —     | 3     | pass        | node: resume 2 — G2 ruled by the user (mise `depends = ["android-sdk"]` on `[mise.tools.flutter]`), D4 and `01-tools.md` amended; U1, U3, U4 reset to pending; wave-gate lines 1 and 2 exit 0                                                                                                                                                                                                                                                                                                                                                          | —      |
| 1    | U1 tools and tool docs | opus  | 2     | pass        | node: edit — flutter: tool `depends = []`, `[mise.tools.flutter]` mise `depends = ["android-sdk"]` (D4 as ruled); gate line 2 lists android and java `mise.dev.toml` and `mise.ci.toml`. DECIDED: began from the round-1 patch; flutter mise table keeps a Note column; java.md covers both the depends-key (dev only) and mise-depends (flutter envs) paths. GAP: dry-run lists names only, so rendered flutter content is assumed free of android-sdk and java. DOCS FALSIFIED: README.md:44-60, docs/sync/index.md:32-36, docs/sync/migrate.md (U3) | —      |
| 1    | R1 wave review         | opus  | 3     | findings(4) | U1 loop round 1. android.md:8 Depends `java` but key is `[]` (index.md:147 convention); kotlin.md:46 "java in dev only" lacks the mise-depends condition; flutter.md:30 note style; flutter.md:46 paragraph not reflowed. RENDER: android-sdk and java only in their own `mise.dev.toml` and `mise.ci.toml`; flutter files clean. CONTRACT, RULINGS clean                                                                                                                                                                                              | —      |
| 1    | U1 tools and tool docs | opus  | 3     | pass        | node: edit — 4 R1 findings fixed. DECIDED: android.md Depends "None. Its mise tool adds `java`" (index.md:147 defines the row as the `depends` key; unit edit 7 said `java`). Its render GAP is closed by the R1 RENDER line                                                                                                                                                                                                                                                                                                                           | —      |
| 1    | R1 wave review         | opus  | 4     | pass        | U1 loop round 2: FINDINGS 0; CONTRACT clean; RULINGS clean                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | —      |
| —    | acceptance             | —     | —     | skipped     | why: no `covers:`, no acceptance criteria                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | —      |
| —    | ux                     | —     | —     | skipped     | why: no `covers:`, no Screens contract                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | —      |
| —    | renders                | —     | —     | skipped     | why: ux stage did not run                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | —      |
| —    | reconcile              | —     | —     | skipped     | why: no `covers:` (no registry, environment, harness or stamp edit); no `code` unit, so nothing to persist                                                                                                                                                                                                                                                                                                                                                                                                                                             | —      |
| 2    | U3 docs                | opus  | 1     | pass        | node: edit — docs-sync over `9a3f5e8..HEAD` matched the DOCS FALSIFIED lines. README example now the real `android` tool and `base` `os = "linux"`; CLAUDE.md and index.md list vwf.md; index.md tool list adds android, java, swift; migrate.md M16. DECIDED: followed amended D4 over the 03-docs.md quote; index.md step 1 sentence on mise `depends` envs. GAP: M16 placement and the "Find the items" rows (`repo-config.toml` row left at M15)                                                                                                   | —      |
| 2    | R2 wave review         | opus  | 1     | findings(3) | migrate.md:15 `repo-config.toml` row must read "M15, then M16" (U3 GAP call wrong); migrate.md:339 M16 step 3 has nothing for a kotlin-only repo; index.md:36 paragraph not refolded. CONTRACT, RULINGS clean                                                                                                                                                                                                                                                                                                                                          | —      |
| 2    | U3 docs                | opus  | 2     | pass        | node: edit — 3 R2 findings fixed: `repo-config.toml` row "M15, then M16"; M16 step 3 gets a kotlin-only check; index.md tool list refolded. GAP closed                                                                                                                                                                                                                                                                                                                                                                                                 | —      |
| 2    | R2 wave review         | opus  | 2     | pass        | FINDINGS 0; CONTRACT clean; RULINGS clean                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | —      |

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-19-android-swift-java-tools

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
