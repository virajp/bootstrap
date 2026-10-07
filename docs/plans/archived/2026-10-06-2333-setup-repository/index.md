---
type: vwf-plan
title: Set up a repository (bootstrap init)
requires:
  - docs/plans/2026-10-06-2332-tool-setup-config
backlog: []
backlog_pieces: []
covers:
  - docs/blueprint/flows/cli/110-setup-repository/index.md
---

# Plan — Set up a repository (bootstrap init) (2026-10-06)

## Status

**ARCHIVED**

ARCHIVED 2026-10-08 — not run; was APPROVED 2026-10-06 by the user (superseded by the mise rescope of the blueprint)

## Consent

| Action                                            | Granted |
| ------------------------------------------------- | ------- |
| Merge to the integration branch and push on green | yes     |
| After landing: archive the plan folder            | run     |
| After landing: `mise run code:graph`              | run     |
| After landing: `/vwf:handoff` for slice 2         | ask     |
| Release cli publicly                              | none    |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt; an `ask` step stops the run once before it, reports what it
would do, and waits. The release row records the user's answer "No release yet":
no version bump, no publish.

## Goal

`bootstrap init` works as
[Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
specifies: steps 1–8, the mode table, `--dry-run`, `--json`, the guarantees
(all-or-nothing writes, full rollback on a write failure or an interrupt), and
every acceptance criterion below, proven by an e2e test against the built binary
and against an installed tarball. After landing, the flow reads
`implementation: complete`.

## Slice

Flow
[Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
(`cli/110-setup-repository`).

Plan 2 of 2 — requires `docs/plans/2026-10-06-2332-tool-setup-config` (the tool
catalog, the renderer, the templates and the setup config); required by nothing
yet.

## Facts the survey established

- Plan 1 provides: the `cli` package and workspace, `effect` 4.x and
  `@effect/platform-node` 4.x, `effect/cli`, the tool catalog
  (`cli/src/tool/catalog.ts`), the renderer (`cli/src/tool/render.ts`), the
  templates, and setup-config read / validate / encode
  (`cli/src/setup-config/**`). This plan adds only the orchestration.
- `effect/cli` holds `Command`, `Flag`, and `Prompt` with `Text` (validation),
  `MultiSelect` and `Confirm`. `NodeRuntime.runMain` interrupts the program
  gracefully on Ctrl-C, so finalizers run.
- Contract sources: flow 110 (steps, the mode table under step 3, Modes,
  Guarantees, Acceptance); `docs/blueprint/conventions.md` `#errors` (exit codes
  0/1/2/3/130, the `--json` document with top-level `exit` and
  `error = {what, why, next_command}`, `unrestored` on a failed rollback,
  interactive = stdin and stdout are terminals and no `--json`, the repository
  root from any subdirectory), `#backups` (`.bak`, `.N.bak`, byte-identical left
  alone, directory or symlink refused, every output path list sorted by path),
  `#baseline` (atomic multi-write, boundary validation of every `--json`
  document before it prints); `docs/blueprint/design-system.md` Terminal UX
  (stdout results, stderr progress, colour rules, `--quiet` errors only,
  `--json` ignores `--quiet`).
- Harness (`.config/vwf.yaml` `harness:`): `e2e_local: false`,
  `e2e_staging: false`; no `test:e2e` task exists. `local_stack` is not needed
  (no backing services).
- No code in the repo parses a git remote, names backups or rolls back writes.
- Backlog: unreadable — no GitHub project; no ids.

## Assumed decisions — confirm or override at review

| # | Decision               | Ruling                                                                                                                                                                                                                                                                                                                     | Rejected                                                                    | Unit   |
| - | ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- | ------ |
| 1 | Git access             | "Run the git binary": `git rev-parse --show-toplevel` and `git remote get-url origin` through the Effect process service                                                                                                                                                                                                   | parse `.git` directly                                                       | U2     |
| 2 | e2e_staging            | "Pack-and-install e2e": `test:e2e:staging` packs the cli with `pnpm pack`, installs the tarball into a clean temp directory and runs the e2e suite against that installed binary                                                                                                                                           | park it; publish a registry prerelease                                      | U1     |
| 3 | Prompts                | the `effect/cli` `Prompt` module (`Text`, `MultiSelect`, `Confirm`)                                                                                                                                                                                                                                                        | `@clack/prompts` (superseded by the Effect CLI choice)                      | U5     |
| 4 | Interrupt and rollback | rollback is an Effect finalizer on the write phase; `NodeRuntime.runMain` turns Ctrl-C into a graceful interrupt; the command maps an interrupt to exit 130                                                                                                                                                                | a hand-written signal handler                                               | U3, U6 |
| 5 | Fault injection in e2e | "Real faults + test pause": a write failure is a real one (a target directory made read-only); an interrupt uses one hidden env variable `BOOTSTRAP_TEST_PAUSE_AFTER_WRITE=<n>` that pauses after the n-th write so the test sends SIGINT; it does nothing else, is undocumented, and is read only at the process boundary | real faults only (timer SIGINT, flaky); interrupt tested at unit level only | U3, U6 |
| 6 | Rollback scope         | rollback removes every file and every directory the run created and restores every backup, so the tree is byte-identical to before                                                                                                                                                                                         | remove files only                                                           | U3     |
| 7 | `--json` validation    | every `--json` document is an Effect Schema value, encoded and validated before it prints                                                                                                                                                                                                                                  | build JSON by hand                                                          | U4     |
| 8 | Review placement       | "One per plan": one review row after the last code unit                                                                                                                                                                                                                                                                    | an extra earlier row                                                        | U7     |

## New dependencies

none — plan 1 adds every package this plan uses.

## Units

| Id | Wave | Unit file                                    | Kind   | Owns                                                                                                                                                              | Depends on             | Status  | Commit |
| -- | ---- | -------------------------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- | ------- | ------ |
| U1 | 1    | [01-e2e-harness.md](01-e2e-harness.md)       | code   | `.config/mise/tasks/test/**`, `cli/vitest.e2e.config.ts`, `cli/test/e2e/helpers.ts`, `cli/test/e2e/smoke.test.ts`, `.config/vwf.yaml` (the `harness:` block only) | —                      | pending |        |
| U2 | 2    | [02-repo-context.md](02-repo-context.md)     | code   | `cli/src/repo/**`, `cli/test/repo/**`                                                                                                                             | U1                     | pending |        |
| U3 | 2    | [03-apply-engine.md](03-apply-engine.md)     | code   | `cli/src/apply/**`, `cli/test/apply/**`                                                                                                                           | U1                     | pending |        |
| U4 | 2    | [04-output.md](04-output.md)                 | code   | `cli/src/report/**`, `cli/test/report/**`                                                                                                                         | U1                     | pending |        |
| U5 | 3    | [05-init-options.md](05-init-options.md)     | code   | `cli/src/commands/init/options.ts`, `cli/src/commands/init/prompts.ts`, `cli/test/commands/init/options.test.ts`, `cli/test/commands/init/prompts.test.ts`        | U2                     | pending |        |
| U6 | 4    | [06-init-command.md](06-init-command.md)     | code   | `cli/src/commands/init/command.ts`, `cli/src/cli.ts`, `cli/src/bin.ts`, `cli/test/e2e/init/**`                                                                    | U3, U4, U5             | pending |        |
| U7 | 5    | [07-review.md](07-review.md)                 | review | —                                                                                                                                                                 | U1, U2, U3, U4, U5, U6 | pending |        |
| U8 | 6    | [08-docs.md](08-docs.md)                     | edit   | the repo's docs (README, CLAUDE.md, `docs/**` outside `docs/blueprint/` and `docs/plans/`, `cli/README.md`)                                                       | all                    | pending |        |
| U9 | 7    | [09-gates-and-bump.md](09-gates-and-bump.md) | edit   | — (no release: no version file, no generator)                                                                                                                     | U8                     | pending |        |

## Shared-file rule

| File                                 | Why it collides                                  | Owner   |
| ------------------------------------ | ------------------------------------------------ | ------- |
| `cli/src/cli.ts`, `cli/src/bin.ts`   | the root command and the process boundary        | U6 only |
| `.config/vwf.yaml`                   | shared config; only the `harness:` block changes | U1 only |
| `cli/package.json`, `pnpm-lock.yaml` | no dependency changes in this plan               | nobody  |
| README, CLAUDE.md, `docs/**`         | n units editing one doc                          | U8 only |

## Waves

- Wave 1: U1 — the e2e harness the acceptance tests run on.
- Wave 2: U2, U3, U4 — disjoint trees (`repo/`, `apply/`, `report/`).
- Wave 3: U5 — options and prompts need the repository context.
- Wave 4: U6 — wires everything into `bootstrap init` and owns the acceptance
  e2e tests.
- Wave 5: U7 review; wave 6: U8 docs; wave 7: U9 gates.

Waves are ordering only; `execute` runs the units serially.

## Wave gate

`MISE_ENV=dev mise run code:all`, `mise x -- mise run p:cli:check` and
`mise x -- mise run p:cli:test`, plus the wave review, plus every report read
for `UNRESOLVED:`. These tasks exist once plan 1 has landed, which `requires:`
guarantees.

## After landing

| Step                                                       | Mode | Notes                                                                  |
| ---------------------------------------------------------- | ---- | ---------------------------------------------------------------------- |
| archive the plan folder (`plan-management archive`)        | run  | moves this folder to `docs/plans/archived/` and updates the plan index |
| `mise run code:graph`                                      | run  | refreshes the code graph                                               |
| `/vwf:handoff` pointing at `/vwf:plan` for slice 2 (check) | ask  | stops once and offers the handoff                                      |

## Gates the orchestrator keeps

- In a scratch git repository with an `origin`,
  `node cli/dist/bin.mjs init --repo acme/widgets --merge-develop direct --merge-main pr`
  writes the core files and `.config/bootstrap.yaml` and exits 0.
- In that repository, `mise trust --all` then `MISE_ENV=dev mise run setup:all`
  (the next command init prints) completes with exit 0.

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

## Out of scope

- The commands `check`, `update`, `add` and `remove` — slices 2–5 (they reuse U3
  and U4).
- Publishing to npm — the release pipeline is parked.

## Parked

- Release pipeline: npm trusted publishing, the tag shape, the npm version on
  the runner (`docs/memory/gaps/npm-release-open-decisions.md`), and the stale
  `.config/mise/tasks/p/i/*` tasks — the release plan.
- The `site` project — slice 6.
- Flows `cli/120-check-drift`, `cli/130-update-repository`, `cli/140-add-tool`,
  `cli/150-remove-tool` — slices 2–5.
- Dogfooding: re-render this repo's own tooling with `bootstrap` once `check`
  and `update` exist.

## Run log

| Wave | Unit | Model | Round | Outcome | Detail | Commit |
| ---- | ---- | ----- | ----- | ------- | ------ | ------ |

## Acceptance criteria (from blueprint)

- [ ] Given an empty git repository, when `bootstrap init` runs
      non-interactively with all flags, then every core file and the chosen
      non-core tools' files exist, `.config/bootstrap.yaml` records the version,
      values (including `values.tools`) and files, and the exit code is 0. —
      from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given a non-interactive run with every required value flagged, no `-y` and
      no `--tool`, when init runs, then only core files are written and the exit
      code is 0. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `bootstrap init --yes` with no `--tool` in a repository with a
      readable `origin`, when init runs, then only core files are written, the
      picker is not shown, `values.tools` is an empty list and the exit code
      is 0. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `.config/bootstrap.yaml` exists, when init runs, then no file
      changes, the exit code is 1 and the message names `bootstrap update`. —
      from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given the directory is not inside a git repository, when init runs, then
      nothing is written and the exit code is 3. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given init runs from a subdirectory of a repository, when it succeeds,
      then `.config/bootstrap.yaml` and every tool file are at the repository
      root. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given a `.gitignore` exists and `.config/bootstrap.yaml` does not, when
      init runs, then the original is preserved byte-identical as
      `.gitignore.bak`, the new `.gitignore` is written, and the backup is
      listed in the output. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given a successful run, when init finishes, then `created` lists
      `.config/bootstrap.yaml` and every path list is sorted by path. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `--tool core` or `--tool mise`, when init runs, then nothing is
      written and the exit code is 2; given `--tool github --tool github`,
      github is applied once. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given a tool target path that exists as a directory or a symlink, when
      init runs, then nothing is written and the exit code is 3 naming that
      path. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `.gitignore.bak` already exists, when init backs up `.gitignore`,
      then the new backup is `.gitignore.1.bak` and the older backup is
      untouched. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given a pre-existing file byte-identical to what init would write, when
      init runs, then it is untouched, listed `unchanged`, and no `.bak` is
      created. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given an interactive run, when the actor declines at the confirm, then
      nothing is written and the exit code is 1. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given a write failure injected mid-run, when init runs, then the
      repository tree is identical to before and the exit code is 3 naming the
      failing path. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given an interrupt signal mid-write, when init is running, then the
      repository tree is identical to before and the exit code is 130. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `--dry-run`, when init runs, then no file changes, the would-be
      created, backed-up and unchanged files are reported, and the exit code
      is 0. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given non-interactive mode with a required value missing, when init runs,
      then the exit code is 2 and the message names the missing flag. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given non-interactive mode, no `-y`, no `--repo` flag and an `origin` URL
      not readable as a repo path, when init runs, then there is no default and
      the exit code is 2. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `-y`, no `--repo` and no repo path readable from `origin`, when init
      runs in an interactive terminal or not, then it does not prompt, the exit
      code is 2 and the message names `--repo`. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given an interactive run without `-y` or `--scope`, when the actor answers
      `api, web-ui` (or nothing), then scopes are `api` and `web-ui` (or zero);
      an invalid answer repeats the prompt with the rule. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `--dry-run` and a required value with neither flag nor `-y`, when
      init runs, then it does not prompt, the exit code is 2 and the message
      names the flag. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given a successful run with `--json`, when init finishes, then the
      document has `exit` 0, `created`, `backed_up` (each `{path, backup}`
      pairing an original with its backup), `unchanged`, and `next_command`
      equal to `MISE_ENV=dev mise run setup:all`, the same text the human output
      prints. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `--dry-run --json`, when init runs, then the document has the same
      keys and values a real run would return, `backed_up` naming the backup it
      would use, plus `"dry_run": true`, and no file changes. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given an unknown `--tool` name, when init runs, then nothing is written
      and the exit code is 2. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given an invalid flag value (for example `--merge-main squash`), when init
      runs, then the exit code is 2 and the message names the flag and the
      allowed form. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given `--json`, when init runs, then stdout parses as exactly one JSON
      document and contains nothing else. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given an interactive terminal, `--json`, and no `-y` or flag for a
      required value, when init runs, then it does not prompt, the exit code is
      2, and stdout is one JSON document carrying the error. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Given the remote `origin` is `git@<host>:o/r.git` or
      `https://<host>/o/r.git`, when init reads defaults, then the repo default
      is `o/r`; for `https://<host>/a/b/c.git` it is `a/b/c`. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
- [ ] Abuse case: n/a — runs locally with the caller's own permissions on the
      caller's own repository; no remote surface; every input is validated per
      [baseline](../../../conventions.md#baseline) boundary-validation. — from
      [Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)

## Gaps surfaced during execution

- …

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the fresh session's worktree — cut from the integration branch — can see it.

Run in a fresh session:

/vwf:execute docs/plans/2026-10-06-2333-setup-repository

or let the queue pick it, by priority:

/vwf:execute next
