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

**COMPLETE**

COMPLETE 2026-10-21 — 05ccf8c, d8e2768, 6a2b4cc, f536a32, 19257a9, c90c987,
df16e1c, 58f1fd1, 7f2e23b

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
| D7 | G2, hook variables    | Added after the block of wave 1, by the user: "do steps 1 to 3". Before its first `git` call, `test:repo-sync` unsets each `GIT_*` variable that a git hook exports, so its scratch repositories never write to the real repository.            | Run the hook with `pass_filenames` only; no hook    | U3     |

## New dependencies

None.

## Units

| Id | Wave | Unit file                    | Kind   | Owns                                                                                                                        | Depends on | Status | Commit  |
| -- | ---- | ---------------------------- | ------ | --------------------------------------------------------------------------------------------------------------------------- | ---------- | ------ | ------- |
| U1 | 1    | [01-engine.md](01-engine.md) | edit   | `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`                                                                     | —          | green  | d8e2768 |
| U2 | 1    | [02-rename.md](02-rename.md) | edit   | `tools/mise/.config/mise/conf.d/_repo/` (moves to `zz-repo/`), `tools/mise/tools.toml`, `tools/mise/.config/mise.toml.tmpl` | —          | green  | 6a2b4cc |
| U3 | 1    | [03-tests.md](03-tests.md)   | edit   | `.config/mise/tasks/test/repo-sync`                                                                                         | —          | green  | 05ccf8c |
| U4 | 2    | [04-review.md](04-review.md) | review | —                                                                                                                           | U1, U2, U3 | green  |         |
| U5 | 3    | [05-docs.md](05-docs.md)     | edit   | `docs/sync/**`, `README.md`, `CLAUDE.md`, `docs/design/*.md`                                                                | U4         | green  | 58f1fd1 |
| U6 | 4    | [06-gates.md](06-gates.md)   | edit   | `.config/bootstrap.toml`, `.config/mise/conf.d/**`, every synced copy in `.config/` and at the root                         | U5         | green  | 7f2e23b |

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

- G1 (R1, plan gap, non-blocking — RESOLVED on branch `gap-fixes`: the three
  paths now name `zz-repo`):
  `tools/android/.config/mise/tasks/setup/deps/android` lines 113, 137 and 192
  still print `.config/mise/conf.d/_repo/mise.<env>.toml`. The Facts list of
  `_repo` mentions left this file out, and no unit owns it, so the run does not
  change it. After the landing, the task tells a consumer to edit a folder that
  no longer exists. Fix: rename the three paths to `zz-repo` in a later change.

- G2 (wave 1 commit, blocking — RESOLVED by D7: commits c90c987 and 19257a9; the
  hook passed inside every later `tools/` commit, and `core.bare` stayed
  `false`): `test:repo-sync` runs `git` in scratch
  repositories without clearing the `GIT_*` variables that a git hook sets. As
  the `test-repo-sync` pre-commit hook, it wrote to the real repository: it
  set `core.bare = true` in `.git/config` and made the commits `test copy` and
  `init` on the run branch. Fix needed in U3's file: unset `GIT_DIR`,
  `GIT_INDEX_FILE`, `GIT_WORK_TREE`, `GIT_OBJECT_DIRECTORY` and the other
  `GIT_*` variables of a hook before any `git` call of the task.
  Ruling added as D7 (U3 edit 4). Recovery done by hand: `core.bare` set back
  to `false`, the run branch reset (soft) to `b0696c8`, and the index reset;
  the unit edits stay on disk.

## Run log

| Wave | Unit                               | Model | Round | Outcome     | Detail                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Commit  |
| ---- | ---------------------------------- | ----- | ----- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| 0    | preflight                          | —     | 1     | pass        | doctor: mise 2026.10.17, graphify CLI present, graph in main checkout; no `.config/vwf.yaml` (review_round_cap 4); mempalace down, journal not written; wave gate `code:precommit` and `test:repo-sync` green                                                                                                                                                                                                                                                                                                                                                                           | —       |
| 0    | format-check                       | —     | 1     | skipped     | why: no `covers:`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | —       |
| 0    | stack conventions                  | —     | 1     | skipped     | why: no `code` unit                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | —       |
| 0    | wave order                         | —     | 1     | pass        | W1 U1, U2, U3 (edit, one message); W2 U4 (review); acceptance/ux; reconcile; W3 U5 (docs); W4 U6 (gates)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | —       |
| 1    | U2 Rename the seed folder          | opus  | 1     | pass        | edit: `tools/mise/.config/mise/conf.d/_repo/` → `zz-repo/` (mode 644 kept); seed paths in `tools/mise/tools.toml:6`; comment in `mise.toml.tmpl:21`. DOCS FALSIFIED: every `_repo` mention in the Facts doc list → U5                                                                                                                                                                                                                                                                                                                                                                   | —       |
| 1    | U3 Test the dev and zz-repo rules  | opus  | 1     | pass        | edit: pulumi case uses `env = ["dev", "ci"]`; 3 new code-2 cases (`env = ["ci"]`, `default_envs = ["ci"]`, `tools/zzz-late`) with message checks; 14 checks pass. DECIDED: the node dependency case also moves to `["dev", "ci"]`; message checks leave out the absolute path. GAP: the plan named only line 142; line 148 also needed the change                                                                                                                                                                                                                                       | —       |
| 1    | U1 repo-sync dev and zz-repo rules | opus  | 1     | pass        | edit: code 2 for an `env` without dev, a `default_envs` without dev (and `[]`), and a tools/ folder that sorts after `zz-repo` (LC_ALL=C, every clone folder, in `_rs_mise_index`); header comments. DOCS FALSIFIED: CLAUDE.md mise-config env rules → U5. GAP: a folder named exactly `zz-repo` is not stopped (wording taken literally)                                                                                                                                                                                                                                               | —       |
| 1    | R1                                 | opus  | 1     | findings(4) | CONTRACT clean, RULINGS clean. `repo-sync:339` (U1) `> "zz-repo"` lets a folder named `zz-repo` through → loop to U1. Docs → U5: `docs/sync/index.md:184-199,349` (dev rule, exit-code causes), `docs/design/mise.md:152` (env rules). `tools/android/.config/mise/tasks/setup/deps/android:113,137,192` still prints `conf.d/_repo/` in a file no unit owns → gap G1                                                                                                                                                                                                                   | —       |
| 1    | U1 repo-sync dev and zz-repo rules | opus  | 2     | pass        | edit (R1 loop-back): `>= "zz-repo"`, so a folder named `zz-repo` also stops with code 2; message adds "or is named zz-repo" (the part U3 matches is unchanged). DOCS FALSIFIED: the equal name is not in D3 → U5 docs say "sorts after or is named zz-repo"                                                                                                                                                                                                                                                                                                                             | —       |
| 1    | R1                                 | opus  | 2     | pass        | CONTRACT clean, RULINGS clean; `>=` fix holds                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | —       |
| 1    | wave commit                        | —     | 1     | blocked     | why: the `test-repo-sync` pre-commit hook ran `test:repo-sync` inside `git commit`; the task ran `git init/add/commit` with the hook's inherited GIT_DIR/GIT_INDEX_FILE, so it set `core.bare = true` in the main `.git/config` and made the commits `test copy` and `init` on the run branch (bc2db48, 2dbb8ed, 13e844b, cf40a8a). The unit commits df5a624 and 9a8d923 sit on them and their trees are wrong. The working files on disk hold the unit edits. origin/main is unchanged (b0696c8)                                                                                       | —       |
| 1    | resume                             | —     | 1     | pass        | resume after the G2 block: worktree exists, row RUNNING, ruling D7 in the plan (U3 edit 4); U1-U3 edits on disk, uncommitted; wave 1 runs again. Commit order: U3 first, so the `test-repo-sync` hook of the `tools/` commits runs the fixed task                                                                                                                                                                                                                                                                                                                                       | —       |
| 1    | U1 repo-sync dev and zz-repo rules | opus  | 3     | pass        | resume: edits on disk verified (dev rule, default_envs rule, `>= "zz-repo"` folder rule); no change                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | d8e2768 |
| 1    | U2 Rename the seed folder          | opus  | 2     | pass        | resume: edits on disk verified (`zz-repo/` seeds equal HEAD `_repo/`, mode 644; seed paths; comment); no change. GAP: the unit file's `rg` check skips `.config/`; with `--hidden --no-ignore` only the G1 lines remain                                                                                                                                                                                                                                                                                                                                                                 | 6a2b4cc |
| 1    | U3 Test the dev and zz-repo rules  | opus  | 2     | pass        | resume: edits 1-3 on disk verified; edit 4 (D7): `unset GIT_DIR GIT_INDEX_FILE GIT_WORK_TREE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_COMMON_DIR GIT_PREFIX GIT_EXEC_PATH` after `set -euo pipefail`, with a G2 comment; 14 checks pass, bash 3.2 and 5; hook-style run against a scratch repo: its HEAD, status, core.bare unchanged; real repo unchanged                                                                                                                                                                                                             | 05ccf8c |
| 1    | R1                                 | opus  | 3     | pass        | resume: contract review of U3 edit 4 (D7): CONTRACT clean, RULINGS clean                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | —       |
| 1    | wave gate                          | —     | 1     | pass        | `code:precommit` and `test:repo-sync` green; commit order U3, U1, U2 (so the hook runs the fixed task); hook `test-repo-sync` passed inside both `tools/` commits; core.bare stayed false, no stray commits                                                                                                                                                                                                                                                                                                                                                                             | —       |
| 2    | U4                                 | opus  | 1     | findings(1) | security, range b0696c8..2b5d8c9; files: repo-sync (U1 d8e2768), tools/mise (U2 6a2b4cc), test/repo-sync (U3 05ccf8c). [low] `test/repo-sync:16` (U3): the unset list leaves out GIT_CONFIG, GIT_CONFIG_PARAMETERS, GIT_CONFIG_COUNT/KEY_n/VALUE_n and others that `git rev-parse --local-env-vars` lists; D7 closes G2 (no write to the real .git); fix `unset $(git rev-parse --local-env-vars)` → must fix. VERDICT approve. Engine: no high-confidence vulnerabilities                                                                                                              | —       |
| 2    | U4                                 | opus  | 1     | findings(4) | review, range b0696c8..2b5d8c9. [medium] `repo-sync:494` (U1) default_envs dev check only for selected tools → move into `_rs_mise_index` for every tool; [low] `repo-sync:339` (U1) the empty-string fallback of `late` fails open; [low] `test/repo-sync:175` (U3) no case for a folder named `zz-repo` or `default_envs = []`; [low] `test/repo-sync:16` (U3) incomplete unset list (= the security finding). SPEC met; API COMPAT n/a. VERDICT approve                                                                                                                              | —       |
| 2    | U1 repo-sync dev and zz-repo rules | opus  | 4     | pass        | edit (U4 loop-back): default_envs dev check moved into `_rs_mise_index` for every tool of the clone (old per-selected check removed); zz-repo guard `{ print; exit }` without head, returns 2 when the listing fails                                                                                                                                                                                                                                                                                                                                                                    | f536a32 |
| 2    | U3 Test the dev and zz-repo rules  | opus  | 3     | pass        | edit (U4 loop-back): unset the fixed list + the GIT_CONFIG variables + `git rev-parse --local-env-vars` + GIT_CONFIG_KEY_n/VALUE_n (security); cases for a folder named `zz-repo` and `default_envs = []`; 16 checks; hook-style run with injected core.bare=true config left the scratch repo unchanged; real repo unchanged                                                                                                                                                                                                                                                           | 19257a9 |
| 2    | U4                                 | opus  | 2     | pass        | security, range b0696c8..a8dff9b; round-1 unset fix (19257a9) holds; engine finding 2 (engine does not clear GIT_*) not a security finding of this change. Engine: no high-confidence vulnerabilities. VERDICT approve                                                                                                                                                                                                                                                                                                                                                                  | —       |
| 2    | U4                                 | opus  | 2     | pass        | review, range b0696c8..a8dff9b; round-1 fixes hold (f536a32, 19257a9); no findings. SPEC met; API COMPAT n/a. VERDICT approve. U4 green                                                                                                                                                                                                                                                                                                                                                                                                                                                 | —       |
| 2    | R2                                 | opus  | 1     | findings(2) | contract review of f536a32, 19257a9. CONTRACT clean. RULINGS: U3 departed from D7 — `git rev-parse --local-env-vars` runs before the unset (read-only) → loop to U3. `repo-sync:331` (U1) `_rs_mise_index` header does not name the new default_envs return → loop to U1                                                                                                                                                                                                                                                                                                                | —       |
| 2    | U1 repo-sync dev and zz-repo rules | opus  | 5     | pass        | edit (R2 loop-back): `_rs_mise_index` header names the default_envs dev return; comment only                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | df16e1c |
| 2    | U3 Test the dev and zz-repo rules  | opus  | 4     | pass        | edit (R2 loop-back, D7): the fixed list (and GIT_CONFIG_KEY_n/VALUE_n via `"${!prefix@}"`) is unset before any git call; only then `git rev-parse --local-env-vars` runs and its output is unset; 16 checks; hook-style run unchanged; real repo unchanged                                                                                                                                                                                                                                                                                                                              | c90c987 |
| 2    | R2                                 | opus  | 2     | pass        | CONTRACT clean, RULINGS clean; both round-1 items resolved (c90c987, df16e1c) → U4 late re-run over a8dff9b..df16e1c                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | —       |
| 2    | U4                                 | opus  | 1     | pass        | re-run 1 (U4-late1), security, range a8dff9b..df16e1c (c90c987 U3, df16e1c U1); isolation complete; engine: no high-confidence vulnerabilities. VERDICT approve                                                                                                                                                                                                                                                                                                                                                                                                                         | —       |
| 2    | U4                                 | opus  | 1     | pass        | re-run 1 (U4-late1), review, range a8dff9b..df16e1c; no findings; SPEC met; API COMPAT n/a. VERDICT approve. No fix commits, so no R2-late contract review                                                                                                                                                                                                                                                                                                                                                                                                                              | —       |
| 2    | wave gate                          | —     | 1     | pass        | `code:precommit` and `test:repo-sync` (16 checks) green after the late re-run                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | —       |
| —    | acceptance                         | —     | 1     | skipped     | why: no `covers:`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | —       |
| —    | ux                                 | —     | 1     | skipped     | why: no `covers:`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | —       |
| —    | renders                            | —     | 1     | skipped     | why: the ux stage did not run                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | —       |
| —    | reconcile                          | —     | 1     | skipped     | why: no `covers:` and no `code` unit                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | —       |
| 3    | U5 Docs                            | opus  | 1     | pass        | edit: `_repo` → `zz-repo` (load-last reason) in CLAUDE.md, README, docs/sync (index, migrate, tools/mise, tools/android), docs/design (mise, packs); dev rule for env and default_envs (every clone tool) and the zz-repo folder rule (named or after) in CLAUDE.md, index (sequence, exit codes), design docs; D1: the AI asks the user to run `mise run setup:configs`; migrate M17 (rename `_repo`, env needs dev). DECIDED: dry run still by the AI; M15.4/M16 later syncs unchanged; `docs/design/tools.md` also updated. GAP: index step 2 read as "the AI does not run the sync" | —       |
| 3    | R3                                 | opus  | 1     | findings(5) | CONTRACT clean. RULINGS: U5 departed from D1 — the AI still runs the first sync in migrate M16 step 4 and M15.4 step 5. Others: `migrate.md:14` order M16 before M17; `index.md:249` "do not run" contradicts the dry run of step 6.2; `bootstrap-config.md:100` lacks the default_envs and folder rules (edit 4); `index.md:56` lacks "in byte order" → loop to U5                                                                                                                                                                                                                     | —       |
| 3    | U5 Docs                            | opus  | 2     | pass        | edit (R3 loop-back): migrate — M17 before M16 in the row; M15.4 "first sync, do not run it yourself" with the hand copy, tables, M17, M16 first; M15.4 step 6 and M16 step 4 ask the user (D1); index step 2 names the real sync, byte order in the folder rule; bootstrap-config gets the default_envs and folder rules. DECIDED: M15.7 and vwf.md syncs are later syncs, AI-run                                                                                                                                                                                                       | 58f1fd1 |
| 3    | R3                                 | opus  | 2     | findings(1) | round-1 findings resolved; CONTRACT clean, RULINGS clean. **contested** (wave-review cap of 2 rounds): `docs/sync/migrate.md:361` and `:17` say M17 runs "before the first sync", but a repository with `conf.d/_repo/` synced before; it means "before the next sync"                                                                                                                                                                                                                                                                                                                  | —       |
| 3    | wave gate                          | —     | 1     | pass        | `code:precommit` and `test:repo-sync` green; no UNRESOLVED                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | —       |
| 4    | U6 Gates                           | opus  | 1     | pass        | edit: `[tools.claude]` added; `conf.d/_repo/` → `zz-repo/`, `"pipx:graphifyy"` and the empty `[tools]` removed; sync from a scratch clone: `using bootstrap at eadc4d4`, 51/51; `mise run init`; second sync: 40 unchanged, 4 kept seed, 1 skip env, 1 skip seed (pnpm-workspace.yaml, no package.json); `code/graph` header now tools/claude, `graphify-refresh` hook back, new `conf.d/claude/`, `setup/ai`; test:repo-sync green; core.bare false                                                                                                                                    | 7f2e23b |
| 4    | R4                                 | opus  | 1     | pass        | CONTRACT clean, RULINGS clean; synced copies match their sources                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | —       |
| 4    | wave gate                          | —     | 1     | pass        | `code:precommit` and `test:repo-sync` green; no UNRESOLVED                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | —       |
| —    | reconcile                          | —     | 1     | pass        | final wave gate green over the finished tree (`code:precommit`, `test:repo-sync` 16 checks)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | —       |
| —    | reconcile                          | —     | 1     | pass        | orchestrator gates: `conf.d/zz-repo/` exists, `conf.d/_repo/` gone; `graphify-refresh` in `.config/pre-commit-config.yaml`; `code/graph` header `tools/claude/...`; U6 second sync only unchanged, kept seed, skip env, plus one accepted skip seed                                                                                                                                                                                                                                                                                                                                     | —       |

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-21-packs-gap-fixes

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
