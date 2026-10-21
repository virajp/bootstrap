---
type: vwf-change-plan
title: Close the gaps of the packs and base split
requires: []
source: []
backlog: []
backlog_pieces: []
gaps_of:
  - docs/plans/2026-10-21-packs-and-base-split
---

# Plan — Close the gaps of the packs and base split (2026-10-21)

## Status

**APPROVED**

APPROVED 2026-10-21 by the user

## Consent

| Action                                                   | Granted |
| -------------------------------------------------------- | ------- |
| Merge to the integration branch and push on green        | yes     |
| After landing: archive the plan folder                   | run     |
| After landing: archive `2026-10-21-packs-and-base-split` | run     |
| After landing: `mise run code:graph`                     | run     |
| End an `all` run after landing                           | no      |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt, and `run` is the only mode. The integration branch of this
repository is `main`; the run lands with `--no-ff`. The archive of this folder
runs only when it has no open gap (standing rule of the user).

## Release levels

| Project     | Level | Reason                                                                                                          |
| ----------- | ----- | --------------------------------------------------------------------------------------------------------------- |
| `bootstrap` | MAJOR | Consumers must rename `conf.d/_repo/` to `conf.d/zz-repo/`, and a values file with a `ci`-only `env` now fails. |

## Goal

After this plan lands, the overrides of a repository win over the tool
fragments again, each tool is in `dev`, the adoption docs ask the user to run
the first sync, and this repository is in sync with its graphify hook again.
The plan takes the five gaps of `2026-10-21-packs-and-base-split`.

Reversals, each confirmed by the user: `conf.d/_repo/` becomes
`conf.d/zz-repo/`, and a `ci`-only tool is no longer valid.

## Facts the survey established

- mise loads the folders of `conf.d/` "in folder-name order" and evaluates the
  files "from the lowest precedence to the highest", so a later folder wins
  (Context7 `/jdx/mise` v2026.10.7, `docs/configuration.md` and the 2026.9.14
  release notes). `_repo` sorts before every lowercase tool folder, so it
  loses (G3).
- `_repo` appears in `tools/mise/tools.toml:6` (the two seeds),
  `tools/mise/.config/mise.toml.tmpl:21` (a comment), and the seed folder
  `tools/mise/.config/mise/conf.d/_repo/`. `repo-sync` has no `_repo` literal.
  Docs with `_repo`: `CLAUDE.md`, `README.md`, `docs/sync/index.md`,
  `docs/sync/migrate.md`, `docs/sync/tools/mise.md`,
  `docs/sync/tools/android.md`, `docs/sync/tools/github.md`,
  `docs/design/mise.md`, `docs/design/packs.md`.
- No `tools.toml` has a `default_envs` without `dev`. `repo-sync` stops with
  code 2 when `environments` lacks `dev` (`repo-sync:593`), but no rule stops
  `env = ["ci"]` in the values file (G4).
- `test:repo-sync` (`.config/mise/tasks/test/repo-sync`) passes now. Its case
  "env ci on a tool whose mise tool allows dev only" (line 142) uses
  `env = ["ci"]`, and `run_case` compares the exit code and the message
  (line 46).
- `code:sec` prints the header
  (`tools/_security_base/.config/mise/tasks/code/sec/_default:20-22`),
  and each scanner prints a subheader; `CLAUDE.md` rule 10 allows this (G2).
- `setup` does not run `setup:configs`. `docs/sync/index.md:221-232` and
  `docs/sync/migrate.md:14` tell the AI to copy the sync files by hand (G1).
- This repository: `.config/bootstrap.toml` has no `[tools]`;
  `.config/pre-commit-config.yaml` has no `graphify-refresh` hook;
  `.config/mise/tasks/code/graph` is a stale copy with the header
  `tools/base/...`; `.config/mise/conf.d/_repo/mise.dev.toml` has
  `[tools] "pipx:graphifyy" = "latest"` (G5).
- The commit convention allows feat, fix, perf, refactor, revert, test, ops,
  docs, merge, wip, with no scopes.

## Assumed decisions — confirm or override at review

| #  | Decision              | Ruling                                                                                                                                                                                                                                          | Rejected                                            | Unit   |
| -- | --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------- | ------ |
| D1 | G1, the first sync    | "ask user to run `mise run setup:configs`". The AI copies the sync files by hand, then asks the user to run `mise run setup:configs`.                                                                                                           | The AI runs the sync; `setup` runs the sync         | U5     |
| D2 | G2, `code:sec` output | Accepted: "Assuming that `code:sec` prints the header and the subtasks print the subheader. If the assumption is correct then ok" — it is correct.                                                                                              | —                                                   | —      |
| D3 | G3, load order        | "zz-repo is fine". Rename `conf.d/_repo/` to `conf.d/zz-repo/`. A tool folder whose name sorts after `zz-repo` stops the sync with code 2.                                                                                                      | `~repo`; `.config/mise/config.<env>.toml`           | U1, U2 |
| D4 | G4, `dev` always      | "yes, every tool must be in dev". Each `env` of the values file and each `default_envs` without `dev` stops the sync with code 2. The pre-commit code does not change.                                                                          | Skip the hooks of a tool without `dev`              | U1, U3 |
| D5 | G5, this repository   | "Let AI fix the changes due to migration and get the repo in sync". This repository selects `[tools.claude]`, moves its `conf.d/_repo/` files to `conf.d/zz-repo/`, and removes `"pipx:graphifyy"` from its seed, because `claude` supplies it. | Delete the stale `code/graph` and keep `claude` out | U6     |
| D6 | Review row            | One review row after U3, because the plan changes shell scripts.                                                                                                                                                                                | No review row                                       | U4     |

## New dependencies

None.

## Units

| Id | Wave | Unit file                    | Kind   | Owns                                                                                                                        | Depends on | Status  | Commit |
| -- | ---- | ---------------------------- | ------ | --------------------------------------------------------------------------------------------------------------------------- | ---------- | ------- | ------ |
| U1 | 1    | [01-engine.md](01-engine.md) | edit   | `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`                                                                     | —          | pending |        |
| U2 | 1    | [02-rename.md](02-rename.md) | edit   | `tools/mise/.config/mise/conf.d/_repo/` (moves to `zz-repo/`), `tools/mise/tools.toml`, `tools/mise/.config/mise.toml.tmpl` | —          | pending |        |
| U3 | 1    | [03-tests.md](03-tests.md)   | edit   | `.config/mise/tasks/test/repo-sync`                                                                                         | —          | pending |        |
| U4 | 2    | [04-review.md](04-review.md) | review | —                                                                                                                           | U1, U2, U3 | pending |        |
| U5 | 3    | [05-docs.md](05-docs.md)     | edit   | `docs/sync/**`, `README.md`, `CLAUDE.md`, `docs/design/*.md`                                                                | U4         | pending |        |
| U6 | 4    | [06-gates.md](06-gates.md)   | edit   | `.config/bootstrap.toml`, `.config/mise/conf.d/**`, every synced copy in `.config/` and at the root                         | U5         | pending |        |

## Shared-file rule

| File                                                 | Why it collides       | Owner   |
| ---------------------------------------------------- | --------------------- | ------- |
| `.config/**` synced copies, `.config/mise/conf.d/**` | generated by the sync | U6 only |
| `.config/bootstrap.toml`                             | the values file       | U6 only |
| `README.md`, `CLAUDE.md`, `docs/**`                  | human-facing docs     | U5 only |

## Waves

- Wave 1: U1, U2 and U3. They own different files. U3 is in the same wave as
  U1, because the new `dev` rule of U1 changes the message of an existing
  test case, and the gate after wave 1 runs the tests.
- Wave 2: U4, the review of U1 to U3.
- Wave 3: U5, the docs.
- Wave 4: U6, the sync of this repository and the final gate.

## Wave gate

```text
mise x -- mise run code:precommit
mise x -- mise run test:repo-sync
```

plus the wave review, plus every report read for `UNRESOLVED:`. Both lines are
green before wave 1.

## After landing

| Step                                      | Mode | Notes                                                   |
| ----------------------------------------- | ---- | ------------------------------------------------------- |
| archive the plan folder                   | run  | `plan-management archive`, only when it has no open gap |
| archive `2026-10-21-packs-and-base-split` | run  | its gaps moved to this plan                             |
| `mise run code:graph`                     | run  | refreshes the graphify graph of the main checkout       |

## Gates the orchestrator keeps

After U6: `.config/mise/conf.d/zz-repo/` exists and `.config/mise/conf.d/_repo/`
does not; `.config/pre-commit-config.yaml` has the `graphify-refresh` hook;
`.config/mise/tasks/code/graph` has the header `tools/claude/...`; a second
`mise run setup:configs` shows only `unchanged`, `kept seed` or `skip env`.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this file, its owned
paths plus "touch nothing outside this list", the facts section, the shared-file
rule, and the return block below. A unit never bumps a version, never runs a
generator, never edits a doc, never adds a dependency this file does not list,
never commits. A unit deletes with plain `rm`, never `git rm` — it stages
nothing. A unit never writes a lock file with Write, Edit, `sed` or a script.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

## Out of scope

- The pre-commit fix that the earlier run proposed for G4: D4 makes it
  unnecessary.
- An upgrade path for consumers with `conf.d/_repo/`: bootstrap is unreleased;
  `docs/sync/migrate.md` tells them to rename it.

## Parked

- The Flutter packs, and the removal of the mise `depends` on `android-sdk`
  from `tools/flutter/tools.toml`.
- The member keys and the merge keys of `docs/design/bootstrap-config.md`.
- The fnox redesign of `docs/design/fnox.md`, with the secret helper.
- B06 (`env_conf_d`) and B07 (`.gitattributes`).
- A key of the values file that removes a tool of a pack.

## Gaps surfaced during execution

## Run log

| Wave | Unit | Model | Round | Outcome | Detail | Commit |
| ---- | ---- | ----- | ----- | ------- | ------ | ------ |

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-21-packs-gap-fixes

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
