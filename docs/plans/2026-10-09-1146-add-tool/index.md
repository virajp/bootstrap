---
type: vwf-plan
title: Add a tool (bootstrap add)
requires:
  - docs/plans/2026-10-09-1027-remove-tool
backlog: []
backlog_pieces: []
covers:
  - docs/blueprint/flows/cli/140-add-tool/index.md
---

# Plan — Add a tool (bootstrap add) (2026-10-09)

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

After this plan lands, `bootstrap add <tool>...` implements the flow
[Add a tool](../../blueprint/flows/cli/140-add-tool/index.md), and that doc
reads `implementation: complete`. Add reuses the shared modules of plan 3 and
shares one replacement rule with init. No standing decision is reversed.

## Slice

Flow [Add a tool](../../blueprint/flows/cli/140-add-tool/index.md) —
`implementation: none`.

Last plan of the slice-2 chain — requires
`docs/plans/2026-10-09-1027-remove-tool`; required by no plan yet.

## Facts the survey established

- **Survey base:** `develop` at `7ffcf60`. Plan 1 is `RUNNING` in a parallel
  session; plans 2 and 3 are `APPROVED`. Everything they deliver is treated as
  present: the catalog fields, `checkRequires`, `unremovableTools`, the
  renderer, `requireSetUp`, `versionGuard`, Validity, repair, the full-mode
  refresh of `files` and the add rewrite rule (plan 1); the `Catalog` service
  and its fixture layer, the error model, git access and preflight, the writer
  with rollback, the report, the E2E harness and the init replacement rule in
  `cli/src/init/resolve.ts` (plan 2); the shared change computation
  `cli/src/change/compute.ts` with its init and remove modes, the report with a
  key set per command, the common flags and the pipeline skeleton
  `cli/src/command/**`, and the deletion of a removed tool's files (plan 3).
- **1.0 catalog:** no `max: one` category holds two tools, so no real
  replacement exists (row 5).
- **Absent:** the `add` command and its usage errors, the merge of repeated
  names, `already_added`, the selection after the request (held minus replaced
  plus named), the dependent check on a replaced tool, the `add` mode of the
  change computation, the report glyphs for `added` and `already_added`, the
  exit-0 shortcut when every name is already added and nothing was repaired, the
  add pipeline and its tests.
- **Report keys of add** (flow 140 step 7): `added`, `already_added`, `replaced`
  (`{from, to}`), `created`, `changed`, `deleted`, `unchanged`, `kept`,
  `orphaned`, `warnings`, and `next_command` per `conventions.md` `#errors`.
  Every key except `next_command` is always present. Glyphs per
  `design-system.md` `#terminal-ux`: `+` added (success), `·` already added
  (muted).
- **Backlog:** unreadable — no backlog project `bootstrap` exists under
  `virajp`. No backlog id applies.
- **Blind spots:** none. Monorepo; `site/` is out of scope.
- **Stack gate:** `/vwf:doctor cli` found no blocking finding on 2026-10-08; no
  file under `cli/` changed on `develop` since. The TypeScript LSP is installed.
- **Harness:** `p:cli:test`, `p:cli:check`, `p:cli:build`, and `p:cli:e2e` from
  plan 2. No local stack. Run `pnpm install --frozen-lockfile` first.
- **Stack conventions:** `project/typescript-effect-cli` (effect 4.x:
  `effect/cli`, `NodeServices.layer`; Vitest + `@effect/vitest`; `@/` alias;
  errors are values), `repo/pnpm-workspace`, `deploy/npm-package`. No `node:fs`
  and no `process.exit` in handlers.
- **Commit types** allowed by `.config/git-conventional-commits.yaml`: `ops`,
  `docs`, `merge`, `feat`, `fix`, `refactor`. No scopes.
- **Docs:** `cli/` has no README and no CHANGELOG. The docs unit runs
  `vwf:docs-sync`.

## Assumed decisions — confirm or override at review

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                                                                       | Rejected              | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------- | -------------- |
| 1 | Shared code (user, 2026-10-09)                          | Add uses the shared modules of plan 3. The change computation gets an `add` mode: the init classification, plus the deletion of a replaced tool's files as in remove, plus the add rewrite rule                                                                                                              | Own modules for add   | U3             |
| 2 | One replacement rule                                    | The replacement rule of init (plan 2 U4) moves into a shared module that init and add use: `--replace` is necessary, `-y` is never consent, `replaceable: false` exits 2, `--replace` with nothing to replace is ignored                                                                                     | A second copy for add | U2             |
| 3 | Two exit-2 findings in one step                         | Report every exit-2 finding of a step together (precedent: flow 160 "all usage errors are found together"; the same as plan 3)                                                                                                                                                                               | First finding only    | U2             |
| 4 | Order in step 2                                         | Validity (exit 3), then the version guard (exit 1), then the repair (flow 140 step 2). An older `version` exits 1 with no repair warning                                                                                                                                                                     | Repair first          | U4             |
| 5 | Criteria that need two tools of one `max: one` category | They run in-process with the fixture catalog layer of plan 2 (its row 12): the replacement criteria, the dependent-of-replaced criterion, two named tools of one category, two recorded tools of one category, and another tool held in place of a missing unremovable one; the 1.0 catalog has no such pair | Leave them uncovered  | U1, U2, U3, U4 |
| 6 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/add/`                                                                                                     | An allow-list         | U1, U2, U3, U4 |

## New dependencies

none.

## Units

| Id | Wave | Unit file                                    | Kind   | Owns                                                                                                                                              | Depends on     | Status  | Commit |
| -- | ---- | -------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- | ------- | ------ |
| U1 | 1    | [01-add-command.md](01-add-command.md)       | code   | `cli/src/add/command.ts`, `cli/src/add/usage.ts`, `cli/src/cli.ts`, `cli/test/add/usage.test.ts`                                                  | —              | pending |        |
| U2 | 1    | [02-add-selection.md](02-add-selection.md)   | code   | `cli/src/add/select.ts`, `cli/src/change/replace.ts`, `cli/src/init/resolve.ts`, `cli/test/add/select.test.ts`, `cli/test/change/replace.test.ts` | —              | pending |        |
| U3 | 1    | [03-add-mode.md](03-add-mode.md)             | code   | `cli/src/change/compute.ts`, `cli/src/report/human.ts`, `cli/test/change/add-mode.test.ts`, `cli/test/report/output.test.ts`                      | —              | pending |        |
| U4 | 2    | [04-add-run.md](04-add-run.md)               | code   | `cli/src/add/run.ts`, `cli/src/add/command.ts`, `cli/test/add/run.test.ts`, `cli/test/e2e/add.e2e.test.ts`                                        | U1, U2, U3     | pending |        |
| U5 | 3    | [05-review.md](05-review.md)                 | review | —                                                                                                                                                 | U1, U2, U3, U4 | pending |        |
| U6 | 4    | [06-docs.md](06-docs.md)                     | edit   | the repo's docs (README, CLAUDE.md, `docs/**` outside `docs/blueprint/` and `docs/plans/`)                                                        | all            | pending |        |
| U7 | 5    | [07-gates-and-bump.md](07-gates-and-bump.md) | edit   | — (no version bump; no generated file)                                                                                                            | U6             | pending |        |

## Shared-file rule

| File                       | Why it collides                                      | Owner                      |
| -------------------------- | ---------------------------------------------------- | -------------------------- |
| `cli/src/add/command.ts`   | U1 registers the command; U4 connects the pipeline   | U1 in wave 1, U4 in wave 2 |
| `cli/src/report/errors.ts` | the error model of plan 2; reused, never edited here | no unit                    |
| `pnpm-lock.yaml`           | generated                                            | no unit (no dependency)    |
| every human-facing doc     | n units editing one doc                              | U6 only                    |

## Waves

- **Wave 1:** U1, U2 and U3 — disjoint files: the add command, the selection
  with the shared replacement rule, and the add mode with its report glyphs.
- **Wave 2:** U4 — the add pipeline and its acceptance tests.
- **Wave 3:** U5 — the one review row, over U1–U4.
- **Wave 4:** U6 — docs.
- **Wave 5:** U7 — the final gate.

Waves are ordering only. `execute` runs the units serially.

## Wave gate

```text
pnpm install --frozen-lockfile
mise x -- mise run p:cli:check
mise x -- mise run p:cli:test
mise x -- mise run p:cli:build
mise x -- mise run p:cli:e2e
```

Plus the wave review, plus every report read for `UNRESOLVED:`. Plans 2 and 3
land first (`requires:`), so `p:cli:e2e` exists before wave 1.

## After landing

| Step                    | Mode | Notes                                                                                |
| ----------------------- | ---- | ------------------------------------------------------------------------------------ |
| archive the plan folder | run  | `plan-management archive`; moves the folder only when its gap list has no open entry |
| `mise run code:graph`   | run  | refreshes `graphify-out/` for the merged code                                        |

## Gates the orchestrator keeps

- After `mise run p:cli:build`, in a new scratch directory with `git init`, a
  remote `origin` `git@github.com:o/r.git` and `mise` on `PATH`: run
  `node <repo>/cli/dist/bin.mjs init --merge-develop direct --merge-main pr -y`,
  then `remove grype -y`, and commit. Then
  `node <repo>/cli/dist/bin.mjs add grype` exits 0, creates the files of `grype`
  and ends with "commit the changes, then run
  `MISE_ENV=dev mise run setup:all`". After a commit, a second `add grype`
  writes nothing, reports `grype` under `already_added`, prints no next command
  and exits 0.

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

- The `tui` and `show` commands — their own flows (160, 170) and plans.
- The `site` project — slice 6.

## Parked

none.

## Run log

| Wave | Unit | Model | Round | Outcome | Detail | Commit |
| ---- | ---- | ----- | ----- | ------- | ------ | ------ |

## Acceptance criteria (from blueprint)

From [Add a tool](../../blueprint/flows/cli/140-add-tool/index.md) — copied
verbatim. U4 covers every criterion with an E2E test on the built binary, except
the write-failure, failed-restore and interrupt criteria, which U4 covers
in-process with the FileSystem seam of plan 2, and the criteria of row 5, which
U4 covers in-process with the fixture catalog.

- Given a repository at the running version, when `bootstrap add <tool> -y` runs
  for a tool in no occupied `max: one` category, then its files exist,
  `values.tools` contains it, `files` includes them and the exit code is 0; then
  a re-run of `bootstrap init -y` reports every file `unchanged`.
- Given two tool names, when add runs, then both tools are added.
- Given a tool name given twice, when add runs, then the tool is added once and
  the exit code is 0.
- Given a tool already in `values.tools` and no repair to make, when add runs
  with its name, then nothing is written, the full step 7 report is printed with
  the tool in `already_added`, empty path lists except `orphaned`, `warnings`
  and no next command (no `next_command` key under `--json`), and the exit code
  is 0.
- Given `bootstrap add` with no tool name, when add runs, then nothing is
  written, short usage is printed and the exit code is 2.
- Given an unknown tool name, when add runs, then the exit code is 2 and nothing
  is written.
- Given a named tool whose `requires` tool is neither selected nor named (for
  example `taplo` without `dprint`), when add runs, then the exit code is 2, the
  message names the required tool and nothing is written.
- Given `bootstrap add taplo dprint` with neither held, when add runs, then the
  requires check passes on the selection after the request and both are added.
- Given a held tool that another selected tool requires and an add that would
  replace it, when add runs, then the exit code is 2, the message names the
  dependent tool and nothing is written.
- Given `bootstrap add` with no name in a directory that is not a git repository
  or not set up, when add runs, then the exit code is 2, not 3 or 1.
- Given two named tools of one `max: one` category, when add runs, then the exit
  code is 2 and nothing is written.
- Given a held replaceable tool, when add runs with `--replace` in any terminal,
  then the new tool's files are written, the old tool's files are removed except
  `create_only` ones, `values.tools` swaps the two and `replaced` lists
  `{from, to}`, and the exit code is 0.
- Given a run in any terminal that would replace a tool, when add runs without
  `--replace`, then the exit code is 2 and nothing is written.
- Given a run that would replace a tool, when add runs with `-y` but without
  `--replace`, then the exit code is 2 and nothing is written.
- Given an interactive terminal, when add needs a replacement and runs without
  `--replace`, then the exit code is 2, nothing is written and no prompt is
  shown.
- Given `--replace` and no replacement to make, when add runs, then the flag is
  ignored and the run proceeds as without it.
- Given a held tool whose `replaceable` is false, when add names a tool of its
  category, then the exit code is 2 and nothing is written, even with
  `--replace`.
- Given a shared file the render would change (for example
  `.config/mise/conf.d/_base/mise.dev.toml`) that is modified, staged, untracked
  or ignored, when add runs, then nothing is written, the exit code is 1 and the
  path is listed.
- Given a base tool with mise entries is added, when add runs, then the full-set
  render changes `.config/mise/conf.d/_base/mise.dev.toml`, and it is listed in
  `changed`.
- Given a setup file that is modified, staged, untracked or ignored, when add
  runs, then nothing is written, the exit code is 1 and `.config/bootstrap.yaml`
  is listed; on a clean run it is listed `changed` (or `unchanged`).
- Given a recorded `values.tools` that misses an unremovable tool and no other
  tool of its `max: one` category is recorded, when add runs, even with every
  named tool already added, then the tool is added again, the repair is written
  back, `.config/bootstrap.yaml` is listed `changed`, `warnings` reports "added
  back unremovable tool `<name>`" and the exit code is 0.
- Given a setup file recorded with an older `format`, when add runs, then the
  older format is read and not refused, and the next write records the running
  cli's format.
- Given a recorded `values.tools` with two tools of one `max: one` category, a
  tool name the running catalog does not have, or a tool whose `requires` tool
  is not selected, when add runs, then nothing is written and the exit code is 3
  naming the fix.
- Given a path in `files` that the running cli does not render for a selected
  tool, when add runs, then it is left in place, stays in `files` and is listed
  `orphaned`.
- Given a recorded version older than the running cli, when add runs, then the
  exit code is 1, the next command is `bootstrap init` and nothing is written.
- Given a recorded version newer than the running cli, or a setup file failing
  Setup config validity, when add runs, then the exit code is 3 and nothing is
  written.
- Given no `.config/bootstrap.yaml`, when add runs, then the exit code is 1 and
  the message names `bootstrap init`.
- Given a directory outside a git repository, or no `mise` on the `PATH`, when
  add runs, then nothing is written and the exit code is 3.
- Given `mise` found elsewhere than `~/.local/bin/mise`, when add runs, then a
  warning names the path and the run continues.
- Given a target that is a directory or a symlink, when add runs, then nothing
  is written and the exit code is 3 naming that path.
- Given an existing `create_only` file of a new tool, when add runs, then it is
  not changed and is listed in `kept`.
- Given a path whose render equals the working copy, when add runs, then it is
  listed `unchanged` and not rewritten.
- Given a write failure injected mid-run, when add runs, then the tracked tree
  matches `HEAD` for every target, created files are gone and the exit code is
  3.
- Given a write failure whose restore also fails, when add runs, then the exit
  code is 3, the output lists every path not restored and, under `--json`,
  `error.unrestored` lists the same paths.
- Given an interrupt (Ctrl-C) during the write, when add runs, then the targets
  are restored, created files are gone and the exit code is 130.
- Given `--dry-run`, when add runs, then nothing is written, no `-y` is needed
  and the exit code is the one the real run would return, including 2 for a
  replacement without `--replace`.
- Given `--json`, when add runs on any exit, then stdout is exactly one document
  with `exit`; on success it has `added`, `already_added`, `replaced`,
  `created`, `changed`, `deleted`, `unchanged`, `kept`, `orphaned`, `warnings`
  and `next_command` equal to `MISE_ENV=dev mise run setup:all` (absent when
  every named tool is already added and nothing was repaired), with sorted
  paths.
- Given `--dry-run --json`, when add runs, then the document is the one the real
  run would return plus `"dry_run": true`, and on a non-zero exit it is the same
  error document plus `"dry_run": true`; no file changes.
- Given a recorded `values.tools` that misses an unremovable tool while another
  tool of its `max: one` category is recorded, when add runs, then nothing is
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

/vwf:execute docs/plans/2026-10-09-1146-add-tool

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
