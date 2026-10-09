---
type: vwf-plan
title: Set up a repository (bootstrap init)
requires:
  - docs/plans/2026-10-08-1124-tool-catalog
backlog: []
backlog_pieces: []
covers:
  - docs/blueprint/flows/cli/110-setup-repository/index.md
---

# Plan — Set up a repository (bootstrap init) (2026-10-09)

## Status

**APPROVED**

APPROVED 2026-10-09 by the user

## Consent

| Action                                                                | Granted   |
| --------------------------------------------------------------------- | --------- |
| Merge to the integration branch and push on green                     | yes       |
| After landing: archive the plan folder                                | run       |
| After landing: `mise run code:graph`                                  | run       |
| After landing: archive `docs/plans/2026-10-06-2332-tool-setup-config` | run       |
| LSP typescript                                                        | installed |
| End an `all` run after landing                                        | no        |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt, and `run` is the only mode. `/vwf:execute` asks nothing at run
time. Both archive steps move a folder only when its gap list has no open entry
(user, 2026-10-09); `plan-management archive` without a prompt leaves such a
folder live and reports the warning.

## Release levels

| Project | Level | Reason                                                                                                     |
| ------- | ----- | ---------------------------------------------------------------------------------------------------------- |
| cli     | NONE  | Pre-1.0: the version stays `0.0.1` until the single 1.0 release, cut by hand (user, 2026-10-09, as plan 1) |

## Goal

After this plan lands, `bootstrap init` implements the flow
[Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md),
and that doc reads `implementation: complete`. A first run sets up a git
repository from flags, `-y` and the remote `origin`, and a re-run renders with
the running bootstrap and writes only what changed, atomically. No standing
decision is reversed.

## Slice

Flow
[Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md) —
`implementation: none`.

Plan 2 of 2 — requires `docs/plans/2026-10-08-1124-tool-catalog` (the entity
cycle tool, tool category and setup config); required by no plan yet.

## Facts the survey established

- **Survey base:** `develop` at `c29139d`. Plan 1 is approved and not executed.
  Everything plan 1 delivers is treated as present: the catalog data,
  `defaultSelection`, `unremovableTools`, `checkRequires`, the templates, the
  renderer with the selection and the `origin` host as inputs, and the pure
  setup-config rules (schema, Validity, repair, write format, `requireSetUp`,
  `versionGuard`, refresh of `files` with the full and remove modes, the rewrite
  rule). Plan 1 has no disk writer (its row 9).
- **Backlog:** unreadable — no backlog project `bootstrap` exists under
  `virajp`. No backlog id applies.
- **Blind spots:** none. Monorepo; `site/` is out of scope.
- **Stack gate:** `/vwf:doctor cli` found no blocking finding on 2026-10-08; no
  file under `cli/` changed since. The TypeScript LSP is installed.
- **Present today:** `setupConfigPath` (`cli/src/setup-config/schema.ts:5`);
  `read`, `parse`, `write` (text only) (`cli/src/setup-config/io.ts:28,56,93`);
  the Validity tagged errors (`cli/src/setup-config/validity.ts:26-160`);
  `renderFiles` (`cli/src/tool/render.ts:98`); the embedded templates
  (`cli/src/tool/templates.ts:36`, `cli/tsdown.config.ts`); the process boundary
  (`cli/src/bin.ts:1-18`; `Runtime.errorExitCode` is honoured by
  `NodeRuntime.runMain`); bare `bootstrap` exits 2 with help through `NoCommand`
  (`cli/src/cli.ts:14-17,34`).
- **Partial:** the root command (`cli/src/cli.ts:21`) has no subcommand and no
  flag. Only `NoCommand` maps an exit code. `cli/test/cli.test.ts:18-31` runs
  `run(args)` in-process with a captured Console — the one command-level test
  pattern.
- **Absent:** every part of `init` — the command and its flags, the usage
  errors, the preflight, the `origin` parser, value and tool resolution, the
  replacement and drop rules, the target checks, `create_only` / `kept`, the
  `orphaned` report, the disk writer with rollback and interrupt, the result
  output (human and `--json`), the next-command rule and `--dry-run`. No code in
  `cli/src` spawns a process.
- **Earlier gaps:** G2 — effect/cli exits 1 on an unknown flag, where the
  contract wants 2. G3 — the framework gives `-v` to `--version`, where
  `design-system.md` `#terminal-ux` gives `-v` to `--verbose` on each command
  and `--version` exists only as `bootstrap --version`. This plan closes both
  (rows 3, 4).
- **Flags of `init`:** the flow's flags (`--repo`, `--add-scope`,
  `--remove-scope`, `--reset-scope`, `--merge-develop`, `--merge-main`,
  `--tool`, `--replace`, `-y`/`--yes`, `--dry-run`) plus the common flags of
  `design-system.md` `#terminal-ux`: `--json`, `-q`/`--quiet`, `-v`/`--verbose`,
  `--no-color`, `--help`.
- **`mise` config files** (next-command rule, `conventions.md` `#errors`):
  `.config/mise.toml`, `.config/miserc.toml` and every file under
  `.config/mise/conf.d/`.
- **Harness:** `mise run p:cli:test` (Vitest), `p:cli:check` (`tsc --noEmit`),
  `p:cli:build` (`tsdown` → `cli/dist/bin.mjs`). `harness.e2e_local` is `false`;
  flow 110 has acceptance criteria, so U1 injects it. A local stack is not
  necessary (no backing service). Run `pnpm install --frozen-lockfile` first.
- **Stack conventions:** `project/typescript-effect-cli` (effect 4.x:
  `effect/cli`, `NodeServices.layer` for FileSystem, Path and child processes;
  Vitest + `@effect/vitest`; `@/` alias; errors are values),
  `repo/pnpm-workspace`, `deploy/npm-package`. No `node:fs` and no
  `process.exit` in handlers.
- **Commit types** allowed by `.config/git-conventional-commits.yaml`: `ops`,
  `docs`, `merge`, `feat`, `fix`, `refactor`. No scopes.
- **Docs:** `cli/` has no README and no CHANGELOG. The docs unit runs
  `vwf:docs-sync`.

## Assumed decisions — confirm or override at review

| #  | Decision                                                | Ruling                                                                                                                                                                                                          | Rejected                                   | Unit   |
| -- | ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ | ------ |
| 1  | Restore after a write failure or an interrupt           | In-memory snapshot of the content and mode of each target before the write, written back; created files deleted. It equals HEAD because every written target is clean (user, 2026-10-09)                        | `git restore --source=HEAD`                | U5     |
| 2  | Fault tests (write failure, interrupt mid-write)        | In-process FileSystem seam that fails or stops at the n-th write; no test code ships in the binary (user, 2026-10-09)                                                                                           | Environment variable in the binary; timing | U5, U8 |
| 3  | G2: effect/cli exits 1 on an unknown flag               | Map the framework's parse errors to a usage error, exit 2 with the short usage, at the command boundary                                                                                                         | An own argument pre-parser                 | U2     |
| 4  | G3: effect/cli gives `-v` to `--version`                | `-v` is `--verbose` on each command; `--version` only at the top level, with no short form                                                                                                                      | Keep the framework default                 | U2     |
| 5  | Git access                                              | `git` runs as a child process through the Effect platform                                                                                                                                                       | A git library (a new dependency)           | U3     |
| 6  | Rollback trigger                                        | One `onExit` handler covers a write failure and an interrupt                                                                                                                                                    | Two separate paths                         | U5     |
| 7  | Plan constraint: catalog-driven code (user, 2026-10-08) | The literal scan of plan 1 (`cli/test/catalog-driven.test.ts`) extends to every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`. `git` and `mise` there are process names, not tool logic | An allow-list of literals                  | U2, U3 |
| 8  | Color and spinner                                       | No new dependency: ANSI codes by color role and a small spinner                                                                                                                                                 | `picocolors`, `ora`                        | U7     |
| 9  | Plan 1 row 9 (setup-config invariant 1)                 | Lands here: a dirty target, `.config/bootstrap.yaml` included, refuses the run only when its render differs from the working copy; the setup config is written last                                             | —                                          | U5, U6 |
| 10 | Command surface before the pipeline                     | U2 registers `init`, its flags and usage checks; its handler ends in a placeholder that U8 replaces with the pipeline before the review                                                                         | One unit for both                          | U2, U8 |
| 11 | Acceptance tests                                        | The built binary runs in temporary git repositories with a `mise` stub on `PATH`; the two fault criteria run in-process (row 2)                                                                                 | Only in-process tests                      | U1, U8 |

## New dependencies

none — Effect gives the file system, the child processes and interruption;
`yaml` and `liquidjs` are installed; color and spinner are own code (row 8).

## Units

| Id  | Wave | Unit file                                      | Kind   | Owns                                                                                                                                                                                                                                           | Depends on                     | Status  | Commit |
| --- | ---- | ---------------------------------------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ | ------- | ------ |
| U1  | 1    | [01-e2e-harness.md](01-e2e-harness.md)         | code   | `.config/mise/tasks/p/cli/e2e`, `cli/vitest.config.ts`, `cli/vitest.e2e.config.ts`, `cli/test/e2e/support/**`, `cli/test/e2e/smoke.e2e.test.ts`                                                                                                | —                              | pending |        |
| U2  | 2    | [02-command-surface.md](02-command-surface.md) | code   | `cli/src/cli.ts`, `cli/src/bin.ts`, `cli/src/init/command.ts`, `cli/src/init/usage.ts`, `cli/src/report/errors.ts`, `cli/test/cli.test.ts`, `cli/test/init/usage.test.ts`, `cli/test/report/errors.test.ts`, `cli/test/catalog-driven.test.ts` | U1                             | pending |        |
| U3  | 3    | [03-git-preflight.md](03-git-preflight.md)     | code   | `cli/src/git/**`, `cli/test/git/**`                                                                                                                                                                                                            | U2                             | pending |        |
| U4  | 3    | [04-resolution.md](04-resolution.md)           | code   | `cli/src/init/resolve.ts`, `cli/test/init/resolve.test.ts`                                                                                                                                                                                     | U2                             | pending |        |
| U5  | 3    | [05-writer.md](05-writer.md)                   | code   | `cli/src/write/**`, `cli/test/write/**`                                                                                                                                                                                                        | U2                             | pending |        |
| U6  | 4    | [06-changes.md](06-changes.md)                 | code   | `cli/src/init/changes.ts`, `cli/test/init/changes.test.ts`                                                                                                                                                                                     | U3, U4                         | pending |        |
| U7  | 4    | [07-report.md](07-report.md)                   | code   | `cli/src/report/human.ts`, `cli/src/report/json.ts`, `cli/src/report/progress.ts`, `cli/test/report/output.test.ts`                                                                                                                            | U2                             | pending |        |
| U8  | 5    | [08-pipeline.md](08-pipeline.md)               | code   | `cli/src/init/run.ts`, `cli/src/init/command.ts`, `cli/test/init/run.test.ts`, `cli/test/e2e/init.e2e.test.ts`                                                                                                                                 | U5, U6, U7                     | pending |        |
| U9  | 6    | [09-review.md](09-review.md)                   | review | —                                                                                                                                                                                                                                              | U1, U2, U3, U4, U5, U6, U7, U8 | pending |        |
| U10 | 7    | [10-docs.md](10-docs.md)                       | edit   | the repo's docs (README, CLAUDE.md, `docs/**` outside `docs/blueprint/` and `docs/plans/`)                                                                                                                                                     | all                            | pending |        |
| U11 | 8    | [11-gates-and-bump.md](11-gates-and-bump.md)   | edit   | — (no version bump; no generated file)                                                                                                                                                                                                         | U10                            | pending |        |

## Shared-file rule

| File                       | Why it collides                                                | Owner                      |
| -------------------------- | -------------------------------------------------------------- | -------------------------- |
| `cli/src/init/command.ts`  | U2 registers the command; U8 connects the pipeline             | U2 in wave 2, U8 in wave 5 |
| `cli/src/report/errors.ts` | the error model every unit raises; U3–U8 import it, never edit | U2 only                    |
| `cli/vitest.config.ts`     | unit tests exclude `cli/test/e2e/**`                           | U1 only                    |
| `pnpm-lock.yaml`           | generated                                                      | no unit (no dependency)    |
| every human-facing doc     | n units editing one doc                                        | U10 only                   |

## Waves

- **Wave 1:** U1 — the E2E harness, before every unit whose Verification runs
  `p:cli:e2e`.
- **Wave 2:** U2 — the command surface and the error model every later unit
  raises.
- **Wave 3:** U3, U4 and U5 — disjoint folders; each stands on U2 only.
- **Wave 4:** U6 (stands on the git status of U3 and the resolution of U4) and
  U7 (stands on U2); disjoint files.
- **Wave 5:** U8 — the pipeline joins U5, U6 and U7 and adds the acceptance
  tests.
- **Wave 6:** U9 — the one review row, over U1–U8.
- **Wave 7:** U10 — docs.
- **Wave 8:** U11 — the final gate.

Waves are ordering only. `execute` runs the units serially.

## Wave gate

```text
pnpm install --frozen-lockfile
mise x -- mise run p:cli:check
mise x -- mise run p:cli:test
mise x -- mise run p:cli:build
```

Plus the wave review, plus every report read for `UNRESOLVED:`.
`mise x -- mise run p:cli:e2e` exists only after U1, so it is in the
Verification of U1–U8 and in U11, not here.

## After landing

| Step                                                   | Mode | Notes                                                                                      |
| ------------------------------------------------------ | ---- | ------------------------------------------------------------------------------------------ |
| archive the plan folder                                | run  | `plan-management archive`; moves the folder only when its gap list has no open entry       |
| `mise run code:graph`                                  | run  | refreshes `graphify-out/` for the merged code                                              |
| archive `docs/plans/2026-10-06-2332-tool-setup-config` | run  | `plan-management archive`; moves it only when its gap list has no open entry, else reports |

## Gates the orchestrator keeps

- After `mise run p:cli:build`, in a new scratch directory with `git init`, a
  remote `origin` `git@github.com:o/r.git` and `mise` on `PATH`:
  `node <repo>/cli/dist/bin.mjs init --merge-develop direct --merge-main pr -y`
  exits 0, writes the files of the default tools (with `github`) and
  `.config/bootstrap.yaml`, and ends with "commit the changes, then run
  `MISE_ENV=dev mise run setup:all`". After `git add -A && git commit`, a second
  run writes nothing, has no closing line and exits 0.
- `node cli/dist/bin.mjs --version` prints `0.0.1`;
  `node cli/dist/bin.mjs init
  --bogus` exits 2.

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

- The `add`, `remove`, `tui` and `show` commands — their own flows (140, 150,
  160, 170) and plans. This plan builds the shared parts they reuse (error
  model, git access, writer, report) only as far as `init` needs them.
- The `site` project — slice 5.
- Installing software or running `setup:all` — init never does it (flow 110 step
  9).

## Parked

none — this plan closes G2 and G3.

## Run log

| Wave | Unit | Model | Round | Outcome | Detail | Commit |
| ---- | ---- | ----- | ----- | ------- | ------ | ------ |

## Acceptance criteria (from blueprint)

From
[Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md) —
copied verbatim. U8 covers every criterion with an E2E test on the built binary,
except the write-failure and interrupt criteria, which U5 and U8 cover
in-process (row 2).

- Given an empty git repository, when `bootstrap init` runs with all flags, then
  the files of the selected tools exist, `.config/bootstrap.yaml` records the
  version, values (`values.tools` = every selected tool) and files, and the exit
  code is 0.
- Given a first run with every required value flagged, no `-y` and no `--tool`,
  when init runs, then only the unremovable tools' files are written and the
  exit code is 0.
- Given `bootstrap init --yes` with no `--tool` on a first run, when init runs,
  then the default tools are selected and the exit code is 0.
- Given the remote `origin` is `git@<host>:o/r.git` or `https://<host>/o/r.git`,
  when init reads defaults, then the repo default is `o/r`; for
  `https://<host>/a/b/c.git` it is `a/b/c`; for `ssh://git@github.com/o/r.git`
  the repo default is not readable; and whenever the `origin` host is
  `github.com`, `github` is selected by default.
- Given a set-up repository and a re-run with nothing to create, change, delete
  or replace, when init runs (in any terminal, with or without `-y`), then
  nothing is written, an existing `create_only` file is listed `kept`, an
  unrendered recorded path is listed `orphaned`, every other file is listed
  `unchanged`, there is no closing line and no `next_command`, and the exit code
  is 0 (a `--dry-run` follows the same rule).
- Given a re-run with `-y` and `--merge-main direct` over a recorded `pr`, when
  init runs, then the flag value is rendered and recorded and the other recorded
  values are kept as supplied.
- Given a re-run whose config records a newer version or format, when init runs,
  then nothing is written and the exit code is 3 "upgrade bootstrap".
- Given a repository set up by an older bootstrap `version` or `format`, when a
  newer bootstrap re-runs init, then the older format is read and not refused,
  the changed files are rewritten, `version` and `format` move to the running
  bootstrap's, `files` is refreshed and orphaned paths are reported `orphaned`.
- Given a first run, when init finishes, then no path is reported `orphaned`.
- Given a config that fails the schema, or a `files` entry that is empty,
  absolute, uses `..` or has a wildcard, when init runs, then nothing is
  written, the exit code is 3 and the message names the failing field and says
  "fix the file, then run again".
- Given an unknown flag, an invalid flag value (for example
  `--merge-main squash`), an unknown `--tool` name or two `--tool` names of one
  `max: one` category, when init runs (even in a directory that is not a git
  repository or with a config that fails the schema), then nothing is written
  and the exit code is 2, not 3; for an invalid flag value the message names the
  flag and the allowed form.
- Given a `values.tools` name the running catalog does not have, or two tools of
  one `max: one` category, or a tool whose `requires` tool is not selected, when
  init runs, then nothing is written and the exit code is 3 naming the fix (for
  example "remove `fnox` from values.tools").
- Given a `values.tools` that misses an unremovable tool and no other tool of
  its `max: one` category is recorded, when init runs, then the tool is added
  again, the warning "added back unremovable tool `<name>`" is reported
  (`warnings`) and the exit code is 0.
- Given `.config/bootstrap.yaml` has uncommitted changes and its content differs
  from the render, when init re-runs, then nothing is written and the exit code
  is 1 listing it; once committed, it is listed `changed` when rewritten and
  `unchanged` when not.
- Given `--tool taplo` without `--tool dprint` (even with `dprint` recorded),
  when init runs, then nothing is written and the exit code is 2 naming
  `dprint`; given `--tool taplo --tool dprint`, the requires check passes on the
  selection after the request and both are selected.
- Given `--tool github --tool github`, when init runs, then github is applied
  once.
- Given a recorded list without `api`, when init runs with `--add-scope api`
  (also repeated as `--add-scope api --add-scope api`), then `api` is recorded
  once, the files are rendered again and every changed file is a target; given
  `api` already in the list, the list is unchanged.
- Given a recorded list with `api`, when init runs with `--remove-scope api`,
  then `api` is no longer recorded and the changed files are rewritten; given
  `--remove-scope web` with `web` not in the list, then the list is unchanged
  and `warnings` has "scope `web` is not recorded".
- Given a recorded list with scopes, when init runs with `--reset-scope`, then
  `values.commit_scopes` is empty and the changed files are rewritten; it also
  works together with every non-scope flag.
- Given `--reset-scope` with `--add-scope` or `--remove-scope`, when init runs,
  then nothing is written and the exit code is 2.
- Given `--add-scope api --remove-scope api`, when init runs, then nothing is
  written and the exit code is 2.
- Given `--add-scope api --remove-scope web` with `web` recorded, when init
  runs, then both are applied: `api` is added and `web` is removed.
- Given a first run with `--add-scope api`, when init runs, then the list is
  `api` alone.
- Given `--add-scope Api` (not lowercase kebab-case), when init runs, then
  nothing is written and the exit code is 2 naming the flag and the allowed
  form.
- Given the directory is not inside a git repository, when init runs, then
  nothing is written and the exit code is 3.
- Given `mise` is not on `PATH`, when init runs, then nothing is written, the
  exit code is 3 and the message says "mise not installed" with the install
  command.
- Given `mise` is found at a path other than `~/.local/bin/mise`, when init
  runs, then it continues and prints one warning naming that path (`--json`: in
  `warnings`).
- Given init runs from a subdirectory of a repository, when it succeeds, then
  `.config/bootstrap.yaml` and every tool file are at the repository root.
- Given a target path that is modified, staged, untracked or ignored in git,
  when init runs, then nothing is written, the exit code is 1 and the output
  lists the path with "commit or stash these files, then run again".
- Given a tracked target deleted in the working tree with the deletion not
  committed, when init runs, then nothing is written and the exit code is 1
  listing the path; once the deletion is committed, init creates the file again.
- Given a dirty file (including `.config/bootstrap.yaml`) whose content and mode
  already equal the render, when init runs, then it is listed `unchanged` and
  the run is not refused.
- Given a tool target path that exists as a directory or a symlink, or whose
  parent directory is a symlink or a regular file, when init runs, then nothing
  is written and the exit code is 3 naming that path.
- Given both a not-a-file target and an uncommitted target, when init runs, then
  nothing is written, the exit code is 3 and the output lists every path of both
  kinds.
- Given an existing `create_only` file, when init runs, then it is untouched and
  listed `kept`.
- Given a path in `files` that the running bootstrap no longer renders for a
  selected tool, when init runs, then it is left in place and listed `orphaned`;
  a recorded path that is no longer rendered and is absent on disk leaves
  `files` silently.
- Given a re-run that needs a replacement (a different tool in a `max: one`
  category than recorded) without `--replace`, when init runs in any terminal
  (also with `-y`, `--dry-run` or `--json`), then it does not prompt, nothing is
  written and the exit code is 2; with `--replace` the replacement is applied
  (deleting the replaced tool's files) and listed in `replaced`.
- Given a re-run that replaces a tool with `--replace` and without `-y`, in any
  terminal, when init runs, then the replacement is applied, the replaced tool's
  non-`create_only` files are deleted, it is listed in `replaced` and the exit
  code is 0.
- Given `--replace` and no replacement to make, when init runs, then the flag is
  ignored and the run proceeds as without it.
- Given `-y` or `--yes`, when init runs, then both behave identically.
- Given a replacement of a tool whose `replaceable` is false, when init runs,
  then nothing is written and the exit code is 2.
- Given a re-run with `-y` whose selection drops a recorded removable tool, when
  init runs, then that tool's non-`create_only` files are deleted and its
  `create_only` files are kept and reported.
- Given a re-run whose selection drops a recorded removable tool (whether or not
  it deletes files), when init runs (including `--json`) without `-y`, then
  nothing is written and the exit code is 2 "removing files needs -y"; with `-y`
  the files are deleted, and `--dry-run` (also `--dry-run --json`) reports the
  deletions without needing `-y`.
- Given a successful run, when init finishes, then a first run's `created` lists
  `.config/bootstrap.yaml`, every task file has mode `0755` and every path list
  is sorted by path.
- Given a write failure injected mid-run, when init runs, then every changed
  target is restored from `HEAD`, every created file is deleted and the exit
  code is 3 naming the failing path.
- Given an interrupt signal mid-write, when init is running, then the working
  tree is as before and the exit code is 130.
- Given `--dry-run`, when init runs, then no file changes, the would-be created,
  changed, deleted, kept, orphaned and replaced paths are reported, and the exit
  code is 0.
- Given a required value with no flag, no recorded value and no default accepted
  by `-y` (also with `--dry-run` or `--json`), when init runs, then it does not
  prompt, the exit code is 2 and the message names the missing flag; with
  `--json` stdout is one JSON document carrying the error.
- Given an interactive terminal and a missing required value without `-y`, when
  init runs, then it does not prompt, the exit code is 2 naming the flag, and
  the error's next command offers `bootstrap tui` or the missing flag, and `-y`
  only when `-y` would supply the value.
- Given a first run, a repo path readable from `origin`, `--merge-develop` and
  `--merge-main` given and no `-y`, when init runs with no `--repo`, then the
  path from `origin` is used and the exit code is 0.
- Given a re-run with a recorded `values.repo` that differs from the path read
  from `origin`, when init runs with no `--repo`, then the recorded value is
  kept; with `--repo` the flag value is rendered and recorded.
- Given a first run, no `--repo` and no repo path readable from `origin`, when
  init runs with or without `-y`, then there is no default, it does not prompt,
  the exit code is 2 and the message names `--repo`.
- Given a successful run or a `--dry-run` that adds a tool or creates, changes
  or deletes a `mise` config file, when init finishes, then the human output
  ends with exactly "commit the changes, then run
  `MISE_ENV=dev mise run setup:all`"; with `--json` the success document has
  `exit` 0, `created`, `changed`, `deleted`, `unchanged`, `kept`, `orphaned`,
  `replaced` (each `{from, to}`), `warnings`, and `next_command` equal to
  `MISE_ENV=dev mise run setup:all`.
- Given a re-run that only changes a value (for example `--merge-main direct`)
  and no `mise` config file, when init finishes, then the human output ends with
  exactly "commit the changes", there is no `next_command` and the exit code
  is 0.
- Given `--dry-run --json`, when init runs, then the document has the same keys
  and values a real run would return plus `"dry_run": true`, and no file
  changes; on a non-zero exit it is the same error document plus
  `"dry_run": true`.
- Given `--json`, when init runs, then stdout parses as exactly one JSON
  document and contains nothing else.
- Given a recorded `values.tools` that misses an unremovable tool while another
  tool of its `max: one` category is recorded, when init runs, then nothing is
  written and the exit code is 3 naming the fix
  ([Validity](../../blueprint/flows/cli/../blueprint/entities/setup-config/index.md#validity)).
- Abuse case: n/a — runs locally with the caller's own permissions on the
  caller's own repository; no remote surface; every input is validated per
  [baseline](../../blueprint/flows/cli/../blueprint/conventions.md#baseline)
  boundary-validation.

## Gaps surfaced during execution

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-09-0931-setup-repository

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
