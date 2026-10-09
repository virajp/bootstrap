---
type: vwf-plan
title: Remove a tool (bootstrap remove)
requires:
  - docs/plans/2026-10-09-0931-setup-repository
backlog: []
backlog_pieces: []
covers:
  - docs/blueprint/flows/cli/150-remove-tool/index.md
---

# Plan — Remove a tool (bootstrap remove) (2026-10-09)

## Status

**APPROVED**

APPROVED 2026-10-09 by the user

## Consent

| Action                                            | Granted   |
| ------------------------------------------------- | --------- |
| Merge to the integration branch and push on green | yes       |
| After landing: archive the plan folder            | run       |
| After landing: `mise run code:graph`              | run       |
| LSP typescript                                    | installed |
| End an `all` run after landing                    | no        |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt, and `run` is the only mode. `/vwf:execute` asks nothing at run
time. The archive step moves the folder only when its gap list has no open entry
(user, 2026-10-09); `plan-management archive` without a prompt leaves such a
folder live and reports the warning.

## Release levels

| Project | Level | Reason                                                                                          |
| ------- | ----- | ----------------------------------------------------------------------------------------------- |
| cli     | NONE  | Pre-1.0: the version stays `0.0.1` until the single 1.0 release, cut by hand (user, 2026-10-09) |

## Goal

After this plan lands, `bootstrap remove <tool>...` implements the flow
[Remove a tool](../../blueprint/flows/cli/150-remove-tool/index.md), and that
doc reads `implementation: complete`. On the way, the init-only parts of plan 2
become shared modules that `init`, `remove` and later `add` use (row 1). No
standing decision is reversed.

## Slice

Flow [Remove a tool](../../blueprint/flows/cli/150-remove-tool/index.md) —
`implementation: none`.

Plan 3 of the slice-2 chain — requires
`docs/plans/2026-10-09-0931-setup-repository`; required by the plan for
[Add a tool](../../blueprint/flows/cli/140-add-tool/index.md) (flow 140), which
is planned next. Flow 140 step 5 removes a replaced tool "as in Remove a tool",
so this flow is planned first.

## Facts the survey established

- **Survey base:** `develop` at `7ffcf60`. Plan 1 is `RUNNING` in a parallel
  session; plan 2 is `APPROVED`. Everything plans 1 and 2 deliver is treated as
  present: the catalog fields and `unremovableTools`, `checkRequires`; the
  renderer with `createOnly`; `requireSetUp`, `versionGuard`, Validity, repair
  on read, the remove-mode refresh of `files` and the rewrite rule (plan 1 U2,
  U4, U5); the error model `cli/src/report/errors.ts`, git access and preflight
  `cli/src/git/**`, the writer `cli/src/write/apply.ts`, the report
  `cli/src/report/{human,json,progress}.ts`, the change computation
  `cli/src/init/changes.ts`, the pipeline `cli/src/init/run.ts`, the E2E harness
  `p:cli:e2e` and its helpers `cli/test/e2e/support/**` (plan 2 U1–U8).
- **Init-shaped parts of plan 2** (generalised here, row 1):
  - `cli/src/init/changes.ts` takes init's resolution and the init rewrite mode.
    Remove needs the step 5 table: a shared file of a still-selected tool is
    created again, a non-shared absent file is left alone, paths not recorded in
    `files` are not reported, the files of a repaired tool are written.
  - `cli/src/report/{human,json}.ts` fix the init key set. Remove reports
    `removed`, `not_added`, `deleted`, `created`, `changed`, `unchanged`,
    `kept`, `orphaned`, `warnings` in that order; `removed` and `not_added` are
    sorted by name. Glyphs per `design-system.md` `#terminal-ux`: `-` removed
    (emphasis), `·` not added (muted).
  - `cli/src/init/command.ts` holds the common flags; `cli/src/init/run.ts`
    holds the pipeline skeleton (preflight, write unless `--dry-run`, print, the
    `dry_run` error document).
  - The closing-line rule of U7 already takes "a tool added or a repair" and "a
    `mise` config file touched"; remove gives only the repair and the file
    inputs.
- **Absent:** the `remove` command and its usage errors, the name validation of
  step 3, the exit-0 shortcut, the `-y` rule of step 4, the remove pipeline, the
  remove tests.
- **Backlog:** unreadable — no backlog project `bootstrap` exists under
  `virajp`. No backlog id applies.
- **Blind spots:** none. Monorepo; `site/` is out of scope.
- **Stack gate:** `/vwf:doctor cli` found no blocking finding on 2026-10-08; no
  file under `cli/` changed since. The TypeScript LSP is installed.
- **Harness:** `p:cli:test`, `p:cli:check`, `p:cli:build`; `p:cli:e2e` lands
  with plan 2 U1. No local stack. Run `pnpm install --frozen-lockfile` first.
- **Stack conventions:** `project/typescript-effect-cli` (effect 4.x:
  `effect/cli`, `NodeServices.layer`; Vitest + `@effect/vitest`; `@/` alias;
  errors are values), `repo/pnpm-workspace`, `deploy/npm-package`. No `node:fs`
  and no `process.exit` in handlers.
- **Commit types** allowed by `.config/git-conventional-commits.yaml`: `ops`,
  `docs`, `merge`, `feat`, `fix`, `refactor`. No scopes.
- **Docs:** `cli/` has no README and no CHANGELOG. The docs unit runs
  `vwf:docs-sync`.

## Assumed decisions — confirm or override at review

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                        | Rejected                                        | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- | -------------- |
| 1 | Shared code (user, 2026-10-09)                          | Plan 3 moves the init-only parts into shared modules: a change computation with a mode, a report with a key set, common flags and a pipeline skeleton. Init calls them, and its tests stay green                                                              | Revise plan 2 now; separate modules per command | U1, U2, U3     |
| 2 | Two exit-2 findings in one step                         | Report every exit-2 finding of step 3 together (precedent: flow 160 "all usage errors are found together")                                                                                                                                                    | First finding only                              | U5             |
| 3 | Order in step 2                                         | Validity (exit 3), then the version guard (exit 1), then the repair. An older `version` exits 1 with no repair warning                                                                                                                                        | Repair first                                    | U6             |
| 4 | Remove-mode cases                                       | The change computation gets a remove mode for the step 5 table: a shared file is created again, a non-shared absent file is left alone, unrecorded paths are not reported, the files of a repaired tool are written                                           | A separate remove computation                   | U1             |
| 5 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` already covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/change/`, `cli/src/command/` and `cli/src/remove/` | An allow-list of literals                       | U1, U4, U5, U6 |

## New dependencies

none.

## Units

| Id | Wave | Unit file                                    | Kind   | Owns                                                                                                                                       | Depends on             | Status  | Commit |
| -- | ---- | -------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------- | ------- | ------ |
| U1 | 1    | [01-change-modes.md](01-change-modes.md)     | code   | `cli/src/change/**`, `cli/src/init/changes.ts`, `cli/src/init/run.ts` (import only), `cli/test/change/**`, `cli/test/init/changes.test.ts` | —                      | pending |        |
| U2 | 1    | [02-report-keys.md](02-report-keys.md)       | code   | `cli/src/report/human.ts`, `cli/src/report/json.ts`, `cli/test/report/output.test.ts`                                                      | —                      | pending |        |
| U3 | 2    | [03-command-shared.md](03-command-shared.md) | code   | `cli/src/command/**`, `cli/src/init/command.ts`, `cli/src/init/run.ts`, `cli/test/command/**`                                              | U1, U2                 | pending |        |
| U4 | 3    | [04-remove-command.md](04-remove-command.md) | code   | `cli/src/remove/command.ts`, `cli/src/remove/usage.ts`, `cli/src/cli.ts`, `cli/test/remove/usage.test.ts`                                  | U3                     | pending |        |
| U5 | 3    | [05-remove-names.md](05-remove-names.md)     | code   | `cli/src/remove/validate.ts`, `cli/test/remove/validate.test.ts`                                                                           | U3                     | pending |        |
| U6 | 4    | [06-remove-run.md](06-remove-run.md)         | code   | `cli/src/remove/run.ts`, `cli/src/remove/command.ts`, `cli/test/remove/run.test.ts`, `cli/test/e2e/remove.e2e.test.ts`                     | U4, U5                 | pending |        |
| U7 | 5    | [07-review.md](07-review.md)                 | review | —                                                                                                                                          | U1, U2, U3, U4, U5, U6 | pending |        |
| U8 | 6    | [08-docs.md](08-docs.md)                     | edit   | the repo's docs (README, CLAUDE.md, `docs/**` outside `docs/blueprint/` and `docs/plans/`)                                                 | all                    | pending |        |
| U9 | 7    | [09-gates-and-bump.md](09-gates-and-bump.md) | edit   | — (no version bump; no generated file)                                                                                                     | U8                     | pending |        |

## Shared-file rule

| File                        | Why it collides                                      | Owner                      |
| --------------------------- | ---------------------------------------------------- | -------------------------- |
| `cli/src/init/run.ts`       | U1 changes its import; U3 moves its skeleton out     | U1 in wave 1, U3 in wave 2 |
| `cli/src/remove/command.ts` | U4 registers the command; U6 connects the pipeline   | U4 in wave 3, U6 in wave 4 |
| `cli/src/report/errors.ts`  | the error model of plan 2; reused, never edited here | no unit                    |
| `pnpm-lock.yaml`            | generated                                            | no unit (no dependency)    |
| every human-facing doc      | n units editing one doc                              | U8 only                    |

## Waves

- **Wave 1:** U1 and U2 — disjoint; the shared change computation and the shared
  report. U1 also edits the one import line of `cli/src/init/run.ts`.
- **Wave 2:** U3 — common flags and the pipeline skeleton, on top of U1 and U2.
- **Wave 3:** U4 and U5 — disjoint files under `cli/src/remove/`.
- **Wave 4:** U6 — the remove pipeline and its acceptance tests.
- **Wave 5:** U7 — the one review row, over U1–U6.
- **Wave 6:** U8 — docs.
- **Wave 7:** U9 — the final gate.

Waves are ordering only. `execute` runs the units serially.

## Wave gate

```text
pnpm install --frozen-lockfile
mise x -- mise run p:cli:check
mise x -- mise run p:cli:test
mise x -- mise run p:cli:build
mise x -- mise run p:cli:e2e
```

Plus the wave review, plus every report read for `UNRESOLVED:`. Plan 2 lands
first (`requires:`), so `p:cli:e2e` exists before wave 1.

## After landing

| Step                    | Mode | Notes                                                                                |
| ----------------------- | ---- | ------------------------------------------------------------------------------------ |
| archive the plan folder | run  | `plan-management archive`; moves the folder only when its gap list has no open entry |
| `mise run code:graph`   | run  | refreshes `graphify-out/` for the merged code                                        |

## Gates the orchestrator keeps

- After `mise run p:cli:build`, in a new scratch directory with `git init`, a
  remote `origin` `git@github.com:o/r.git` and `mise` on `PATH`: run
  `node <repo>/cli/dist/bin.mjs init --merge-develop direct --merge-main pr -y`
  and commit. Then `node <repo>/cli/dist/bin.mjs remove grype -y` exits 0,
  deletes the files of `grype`, and `.config/bootstrap.yaml` no longer lists
  `grype`. After a commit, a second `remove grype -y` writes nothing, reports
  `grype` under `not_added` and exits 0.
- `node <repo>/cli/dist/bin.mjs remove mise -y` in the same repository exits 2
  and writes nothing.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this file, its owned
paths plus "touch nothing outside this list", the facts section, the shared-file
rule, and the return block below. A unit never bumps a version, never runs a
generator, never edits a doc, never adds a dependency this file does not list,
never commits. A unit deletes with plain `rm`, never `git rm` — it stages
nothing.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

A `GAP:` is a hole in the plan the unit could proceed past on a stated
assumption; it is recorded and the run continues. An `UNRESOLVED:` is a ruling
the unit could not proceed without; it blocks the unit and its dependents.

## Out of scope

- The `add`, `tui` and `show` commands — their own flows and plans. The add plan
  reuses the shared modules of U1–U3.
- The `site` project — slice 6.

## Parked

none.

## Run log

| Wave | Unit | Model | Round | Outcome | Detail | Commit |
| ---- | ---- | ----- | ----- | ------- | ------ | ------ |

## Acceptance criteria (from blueprint)

From [Remove a tool](../../blueprint/flows/cli/150-remove-tool/index.md) —
copied verbatim. U6 covers every criterion with an E2E test on the built binary,
except the write-failure and interrupt criteria, which U6 covers in-process with
the FileSystem seam of plan 2.

A criterion that says only "when remove runs" runs with `-y`; a criterion that
names its flags (no `-y`, `--dry-run`) runs with those only.

- Given a repository with github added, when `bootstrap remove github -y` runs,
  then github's files are deleted, `values.tools` no longer contains github,
  `files` no longer lists them, nothing is committed and the exit code is 0.
- Given a removable base tool such as `dprint`, with `taplo` not selected, when
  `bootstrap remove dprint -y` runs, then its files are deleted,
  `.config/mise/conf.d/_base/mise.dev.toml` is reported `changed` without its
  entries, the next command is `MISE_ENV=dev mise run setup:all` and the exit
  code is 0.
- Given `mempalace` is selected, when `bootstrap remove mempalace -y` runs, then
  `mempalace.yaml` (`create_only`) stays on disk, is reported `kept` and is no
  longer in `files`, the tool's other files are deleted and the exit code is 0.
- Given a recorded shared file missing from disk whose tool stays selected, when
  remove runs and its absence is committed, then the file is created again,
  reported `created` and the exit code is 0.
- Given a recorded shared file of a still-selected tool that is deleted in the
  working tree with the deletion not committed, when remove runs, then nothing
  is written, the path is listed and the exit code is 1.
- Given a recorded path of a removed tool that is deleted in the working tree
  with the deletion not committed, when remove runs, then nothing is written,
  the path is listed and the exit code is 1.
- Given a committed hand edit to a file of a still-selected tool, when remove
  runs with `-y`, then the file is overwritten with its render, reported
  `changed` and the exit code is 0.
- Given a tool whose `removable` is false (e.g. `mise`), when remove runs, then
  nothing is written and the exit code is 2.
- Given an unknown tool name, when remove runs, then the exit code is 2.
- Given a tool not in `values.tools`, when remove runs with its name, then
  nothing is written, it is reported `not_added` and the exit code is 0.
- Given `bootstrap remove` with no names, when it runs, then nothing is written,
  a short usage is shown and the exit code is 2.
- Given `taplo` is selected, when `bootstrap remove dprint -y` runs, then
  nothing is written and the exit code is 2 naming `taplo`.
- Given `taplo` and `dprint` are selected, when
  `bootstrap remove taplo dprint -y` runs, then the requires check passes on the
  selection after the request, both are removed and the exit code is 0.
- Given `bootstrap remove` with no names in a directory that is not a git
  repository or not set up, when it runs, then the exit code is 2, not 3 or 1.
- Given a recorded path the running cli no longer renders that is present on
  disk, when remove runs, then the path is not deleted, stays in `files` and is
  reported `orphaned`.
- Given a recorded path the running cli no longer renders that is absent on
  disk, when remove runs, then the path is dropped from `files`, is not reported
  and the exit code is 0.
- Given a tool name given twice, when `bootstrap remove github github -y` runs,
  then the tool is removed once and the exit code is 0.
- Given a recorded path of the removed tool already absent from the repository,
  with the absence committed, when remove runs, then the path is dropped from
  `files` without a report entry and the exit code is 0.
- Given a file of the tool that is modified or staged, when remove runs, then
  nothing is written, the path is listed and the exit code is 1.
- Given a target that is untracked or ignored in git, when remove runs, then
  nothing is written, the path is listed and the exit code is 1.
- Given a target that is a directory or a symlink, when remove runs, then
  nothing is written and the exit code is 3 naming that path.
- Given one target that is a directory and another that is modified, when remove
  runs, then nothing is written, the exit code is 3 and the message lists both
  paths.
- Given `.config/bootstrap.yaml` has an uncommitted change, when remove runs,
  then nothing is written and the exit code is 1 listing it.
- Given a recorded `values.tools` missing an unremovable tool and no other tool
  of its `max: one` category is recorded, when remove runs, even with none of
  the named tools selected, then the repair is written: the tool is in
  `values.tools`, its files are rendered and reported `created` or `changed`,
  `.config/bootstrap.yaml` is listed `changed`, the warning "added back
  unremovable tool `<name>`" is in `warnings` and the exit code is 0.
- Given a setup file recorded with an older `format`, when remove runs, then the
  older format is read and not refused, and the next write records the running
  cli's format.
- Given a recorded `values.tools` with a tool name the running catalog does not
  have, when remove runs, then nothing is written and the exit code is 3 naming
  the fix.
- Given a run in any terminal that reaches step 4 (a named tool is selected, or
  a repair happened) without `-y`, when remove runs, then the exit code is 2 and
  nothing is written.
- Given an interactive terminal, when remove runs without `-y`, then the exit
  code is 2 with "removing files needs -y", nothing is written and no prompt is
  shown.
- Given a `--dry-run` without `-y`, in any terminal, when remove runs, then the
  plan is shown, nothing is written and the exit code is 0.
- Given `--json`, when remove exits on any code, then exactly one JSON result is
  printed, with `removed`, `not_added`, `deleted`, `created`, `changed`,
  `unchanged`, `kept`, `orphaned` and `warnings` on success; a successful JSON
  result has a top-level `next_command` key only per
  [errors](../../blueprint/flows/cli/../blueprint/conventions.md#errors), and an
  error document still carries `error.next_command`.
- Given `--dry-run --json`, when remove runs, then nothing is written and the
  JSON result the real run would return is printed plus `"dry_run": true`; on a
  non-zero exit it is the same error document plus `"dry_run": true`.
- Given no `.config/bootstrap.yaml`, when remove runs, then the exit code is 1
  and the message names `bootstrap init`.
- Given a recorded version older than the running cli, when remove runs, then
  the exit code is 1 and the next command is `bootstrap init`.
- Given an invalid setup file or a newer recorded version, when remove runs,
  then the exit code is 3 and nothing is written.
- Given a directory outside a git repository, or no `mise` on the `PATH`, when
  remove runs, then nothing is written and the exit code is 3.
- Given `mise` found at a path other than `~/.local/bin/mise`, when remove runs,
  then one warning names the path, it is listed in `warnings` and the run
  continues.
- Given a write failure injected mid-run, when remove runs, then every changed
  or deleted target is restored from `HEAD`, every created file is deleted and
  the exit code is 3.
- Given a run, when the actor presses Ctrl-C during the write, then the
  repository is as it was before and the exit code is 130.
- Given a write failure or an interrupt during the writes whose restore also
  fails, when remove runs, then the exit code is 3 and the output lists every
  path not restored (`--json`: `unrestored`).
- Given a path the cli renders for a tool that stays selected but that is not in
  `files` and is absent on disk, when remove runs, then that path is not created
  and not reported.
- Given a recorded `values.tools` that misses an unremovable tool while another
  tool of its `max: one` category is recorded, when remove runs, then nothing is
  written and the exit code is 3 naming the fix
  ([Validity](../../blueprint/flows/cli/../blueprint/entities/setup-config/index.md#validity)).
- Abuse case: n/a — runs locally with the caller's own permissions on the
  caller's own repository; every input is validated per
  [baseline](../../blueprint/flows/cli/../blueprint/conventions.md#baseline)
  boundary-validation.

## Gaps surfaced during execution

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-09-1027-remove-tool

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
