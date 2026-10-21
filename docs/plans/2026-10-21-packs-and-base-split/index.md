---
type: vwf-change-plan
title: Build the packs and split the base tool
requires: []
source: []
backlog: []
backlog_pieces: []
gaps_of: []
---

# Plan — Build the packs and split the base tool (2026-10-21)

## Status

**RUNNING**

RUNNING since 2026-10-21 10:19 in
/Users/virajpatel/Projects/github.com/virajp/bootstrap/.worktrees/2026-10-21-packs-and-base-split

## Consent

| Action                                            | Granted |
| ------------------------------------------------- | ------- |
| Merge to the integration branch and push on green | yes     |
| After landing: archive the plan folder            | run     |
| After landing: `mise run code:graph`              | run     |
| End an `all` run after landing                    | no      |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt, and `run` is the only mode. The integration branch of this
repository is `main`; the run lands with `--no-ff`. The archive step runs only
when the folder has no open gap (standing rule of the user).

## Release levels

| Project     | Level | Reason                                                                                                                                 |
| ----------- | ----- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `bootstrap` | MAJOR | The values file (`[tools.*]`, `[packs.*]`) and `tools.toml` (`default_envs`) change format; each consumer must change its values file. |

## Goal

After this plan lands, `repo-sync` resolves `[tools.<folder>]` and
`[packs.<name>]` tables of `.config/bootstrap.toml`, and it always adds the
`base` pack of `tools/packs.toml`. The `base` tool no longer exists: its files
are in ten tools, each with its own `.config/mise/conf.d/<tool>/` folder. A
test task checks the new error cases at each commit that changes `tools/`.

The design is in `docs/design/packs.md`, `docs/design/mise.md`,
`docs/design/bootstrap-config.md` and `docs/design/tools.md`. Where this file
and a design document differ, this file wins: the interview of 2026-10-21 made
later rulings, and unit U5 brings the design documents up to date.

Reversals of standing decisions, each confirmed by the user:

- `tools = [...]` changes from a list to a table of tool folders. This reverses
  `docs/design/bootstrap-config.md` and every current values file.
- `tools.toml` has five necessary keys, not four: `default_envs` is new.
- `docs/sync/<tool>.md` moves to `docs/sync/tools/<tool>.md`; packs get
  `docs/sync/packs/<pack>.md`.
- The rule "a tool that only a dependency adds gets `dev` only" goes away.

## Facts the survey established

- No test suite and no CI exist (`CLAUDE.md:9`, no `.github/`). The hooks are
  in `.config/pre-commit-config.yaml`; `code:merge` runs every hook over every
  file before it merges.
- The engine is `tools/base/.config/mise/tasks/_scripts/repo-sync` (954 lines).
  Functions to change: `_rs_resolve` (223-309, reads `tools` at 225, empty
  check 229-232, `_RS_DIRECT` at 235), `_rs_manifest` (173-185),
  `_rs_mise_env` (439-479, `_RS_DIRECT` rule 444-446), `_rs_mise_render`
  (529-577, `_base` at 531, `setup-all` aliases tied to `base` at 538-549),
  `_rs_merge_model` (367-381, hardcodes `base/.config/mise/tasks/code/merge` at
  368 — after the split both merge models would sync, silently),
  `_rs_tool_extensions` (311-341, reads a submodule's `tools` list),
  `_rs_context` (598-624, `-c tool_<name>=` aliases; a hyphen in a tool name,
  for example `tool_repo-sync`, is not tested with gomplate). Literal `base`
  also at lines 2, 15, 37-38, 48, 230, 365.
- `setup/configs/sync:20` hardcodes `tools/base/.../repo-sync`;
  `setup/configs/sync:7` holds `#MISE tools.gomplate="latest"`, which gomplate
  needs. `setup/configs/_default` has no literal `base`.
- The file loop syncs each file of each tool except `tools.toml`; a file at
  `tools/packs.toml` is outside every tool folder, so it is never synced.
- gomplate reads any TOML file with `-c <alias>=<path>`, so `-c
  p=<clone>/tools/packs.toml` works like the `tools.toml` contexts.
- Test a local commit with `REPO_SYNC_URL=<local clone>` and
  `mise run setup:configs --dry-run` (`CLAUDE.md:84-85`). A shallow clone
  ignores uncommitted work. The sync runs `pre-commit autoupdate`, which needs
  the network (code 5 on failure).
- `code:sec` (`tools/base/.config/mise/tasks/code/sec`): `--staged` (lines
  15-20) runs only `gitleaks git --staged` and exits; a full run builds
  `GRYPE_EXCLUDES` and `OSV_EXCLUDES` from the submodules (24-33), runs grype
  (35-37), osv-scanner (39-44), and gitleaks over a tar copy in `SCAN_DIR` with
  an EXIT trap (46-67). `own_tasks` (`_scripts/helpers:256`) lists executable
  files of a task folder except `_*`, so the orchestrator must be
  `code/sec/_default`.
- `code:graph` (`tools/base/.config/mise/tasks/code/graph`) warns and exits 0
  without graphify; it skips a linked worktree with `git rev-parse --git-dir`
  (30-35). In a submodule that check is not confirmed.
- Every tool outside `base` sources only `_scripts/helpers`. The `base` pack is
  always present, so no tool needs `depends = ["mise"]`.
- The files of `tools/base/` and their target tool are listed in U1. Each
  moved file has a `Synced from bootstrap: tools/base/<path>` header at line 1
  or 2.
- This repository consumes `base`: `.config/bootstrap.toml` has `tools =
  ["base"]`, `environments = ["dev"]`, and the generated
  `.config/mise/conf.d/_base/mise.dev.toml`.
- `.config/releases.yaml` has `bootstrap: MINOR`.
- The commit convention (`.config/git-conventional-commits.yaml`) allows feat,
  fix, perf, refactor, revert, test, ops, docs, merge, wip, with no scopes.
- Docs that name `base`, `_base`, `tools/base`, `root_only` of `claude` or
  `code:graph` (owner U5): `README.md` 37-41, 49, 61, 99, 120, 135, 170, 186;
  `CLAUDE.md` 31, 37, 40, 62, 69, 74, 88, 152, 178, 202, 221, 231-234;
  `docs/sync/base.md` (whole file); `docs/sync/index.md` 33, 118-120, 140,
  150, 162, 198-202, 211, 237, 294; `docs/sync/migrate.md` 52, 64, 69, 96,
  108-124, 175; `docs/sync/vwf.md` 52-63; `docs/sync/claude-tool.md` 3-4, 8,
  14-15, 34-35, 45, 54, 61; `docs/sync/node.md:16`; `docs/sync/python.md:16`;
  `docs/sync/flutter.md:51`; `docs/sync/github.md:40`;
  `docs/sync/gitlab.md:38`.

## Assumed decisions — confirm or override at review

| #   | Decision                          | Ruling                                                                                                                                                                                                                                                                                                                                                                                    | Rejected                                             | Unit   |
| --- | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | ------ |
| D1  | Meaning of `envs`                 | "`envs = ["dev", "ci"]` means that this tool can be installed in either or both these environments." A dependency inherits the environments of its parent; a tool that the consumer selects gets the environments that the consumer chooses.                                                                                                                                              | Each dependency always gets its own `envs`           | U2     |
| D2  | Values file shape                 | `[tools.<folder>]` and `[packs.<name>]` tables. "keys are tool folders and only env". `[packs.*]`: "Same applies to packs as well".                                                                                                                                                                                                                                                       | A list; a `version` or `depends` key                 | U2, U3 |
| D3  | `default_envs`                    | Each `tools.toml` has a fifth necessary key `default_envs`, a list: "list, and yes it must be in each tools.toml. Change the name to `default_envs`". An entry without `env` gets `default_envs`.                                                                                                                                                                                         | A string `default_env`; no default                   | U1, U2 |
| D4  | Environments of the `base` pack   | "base uses default_envs". Each tool of `base` gets its `default_envs`.                                                                                                                                                                                                                                                                                                                    | Each environment that `envs` allows                  | U2     |
| D5  | Limits and errors of environments | `environments` limits every result. An `env` name that is not in `environments`, or that a mise tool of the tool does not allow in its `envs`, stops the sync with code 2. A dependency whose `envs` lacks an inherited environment stops the sync with code 2.                                                                                                                           | Skip the tool without a message                      | U2, U3 |
| D6  | The old list format               | `tools = [...]` (a list) or `packs = [...]` stops the sync with code 2; the message names the table format. No upgrade path (bootstrap is unreleased).                                                                                                                                                                                                                                    | Read both formats                                    | U2, U3 |
| D7  | Packs                             | `tools/packs.toml` has `[pack.<name>]` with `depends` (tool folders, or `p:<name>` for a pack) and optional `required` (tools always installed, not in `depends`). The `base` pack name is fixed in `repo-sync` and is always added. Empty or absent `[tools]` is valid.                                                                                                                  | `base` listed in the values file                     | U1, U2 |
| D8  | Pack errors                       | A pack loop stops the sync with code 2 and shows the loop (`a -> b -> a`). A pack name that is also a tool folder stops it with code 2. An unknown pack stops it with code 3. The `[packs.*]` keys of the values file have no `p:` prefix.                                                                                                                                                | —                                                    | U2, U3 |
| D9  | The split of `base`               | The ten tools and the file of each, as U1 lists them. The orchestrators stay in their tools: "That's not possible since they are part of `base` pack and that is not removable".                                                                                                                                                                                                          | Orchestrators in `mise`                              | U1     |
| D10 | `security`                        | `security` is a pack: `depends = ["gitleaks", "grype", "osv-scanner"]`, `required = ["_security_base"]`. `_security_base` holds `code:sec` and the `sec` hook; each scanner ships `code:sec:<tool>`.                                                                                                                                                                                      | A `security` tool                                    | U1     |
| D11 | `code:sec` split                  | `code/sec/_default` runs `own_tasks code:sec` and passes `--staged` and `--debug` to each. `code:sec:gitleaks` keeps both modes. `code:sec:grype` and `code:sec:osv-scanner` print nothing and exit 0 with `--staged`.                                                                                                                                                                    | One script that calls the scanners                   | U1     |
| D12 | Extensions of `base`              | `dprint.dprint` → `dprint`; `hverlin.mise-vscode`, `redhat.vscode-yaml`, `tamasfe.even-better-toml` → `mise`; `dbaeumer.vscode-eslint`, `foxundermoon.shell-format` → `linter`; `gruntfuggly.todo-tree`, `oderwat.indent-rainbow`, `pkief.material-icon-theme`, `usernamehw.errorlens` → `vscode`.                                                                                        | The four editor preferences in `mise`                | U1     |
| D13 | `claude`                          | "keep `claude` as is ... move `code:graph` inside `claude` ... add logic in `setup:ai` where it silently skips when it's being run inside git-submodule ... `claude` will NOT be `root_only`".                                                                                                                                                                                            | A separate `graphify` tool                           | U1     |
| D14 | `jq` and `yq`                     | In the `mise` tool with `envs = ["dev", "ci"]`: "they will be required in `dev` as well as `ci`".                                                                                                                                                                                                                                                                                         | In `claude`; remove `yq`                             | U1     |
| D15 | Fixed paths in the engine         | `_rs_merge_model` uses the `git` tool path; `setup/configs/sync` loads `tools/repo-sync/.../repo-sync`; the `setup-all` and `setup-<submodule>` aliases go to the `mise` tool. `_base` goes away.                                                                                                                                                                                         | —                                                    | U2     |
| D16 | Autoupdate in tests               | `repo-sync` skips `pre-commit autoupdate` when `REPO_SYNC_NO_AUTOUPDATE=1`, so the tests need no network.                                                                                                                                                                                                                                                                                 | Tests that need the network                          | U2, U3 |
| D17 | Tests                             | "Tests must be in `bootstrap` and must run locally before merge to main". `test:repo-sync` is owned by this repository; a local hook in `.config/repo-sync/pre-commit.yaml` runs it: "run the hook on every commit that changes tools/".                                                                                                                                                  | No tests; a check by hand                            | U3     |
| D18 | Docs layout                       | "one file for each tool, as the rule says. Also we need to have a file for pack." Tool docs in `docs/sync/tools/`, pack docs in `docs/sync/packs/`. `index.md`, `migrate.md` and `vwf.md` stay in `docs/sync/`. `claude-tool.md` keeps its name.                                                                                                                                          | One `base.md` for the ten tools                      | U5     |
| D19 | Secret helper                     | Record in `docs/design/fnox.md`: a run task calls a helper of `_scripts/helpers`, for example `run_with_secrets <command>`. With fnox (`mise x -- which fnox`) it runs `mise x -- fnox exec --profile "${MISE_ENV%%,*}" -- <command>`; with doppler, `doppler run -- <command>`; else the command directly. Both `fnox` and `doppler` selected stops the sync with code 2. Not built now. | A `.tmpl` run task with `<% if has .tools "fnox" %>` | U5     |
| D20 | Review row                        | One review row after U3, because the plan lands shell scripts (`repo-sync`, task scripts).                                                                                                                                                                                                                                                                                                | No review row                                        | U4     |

## New dependencies

None. The tests are plain bash.

## Units

| Id | Wave | Unit file                    | Kind   | Owns                                                                                                                                                                          | Depends on | Status  | Commit  |
| -- | ---- | ---------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------- | ------- |
| U1 | 1    | [01-split.md](01-split.md)   | edit   | `tools/**` (moves only for the files of U2)                                                                                                                                   | —          | green   | c95e4ff |
| U2 | 2    | [02-engine.md](02-engine.md) | edit   | `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`, `tools/repo-sync/.config/mise/tasks/setup/configs/_default`, `tools/repo-sync/.config/mise/tasks/setup/configs/sync` | U1         | green   |         |
| U3 | 3    | [03-tests.md](03-tests.md)   | edit   | `.config/mise/tasks/test/repo-sync`, `.config/repo-sync/pre-commit.yaml`                                                                                                      | U2         | pending |         |
| U4 | 4    | [04-review.md](04-review.md) | review | —                                                                                                                                                                             | U1, U2, U3 | pending |         |
| U5 | 5    | [05-docs.md](05-docs.md)     | edit   | `docs/sync/**`, `README.md`, `CLAUDE.md`, `docs/design/*.md`                                                                                                                  | U4         | pending |         |
| U6 | 6    | [06-gates.md](06-gates.md)   | edit   | `.config/bootstrap.toml`, every synced or generated copy in `.config/` and at the root (`.gitignore`)                                                                         | U5         | pending |         |

## Shared-file rule

| File                                                               | Why it collides                    | Owner                                |
| ------------------------------------------------------------------ | ---------------------------------- | ------------------------------------ |
| `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`            | U1 moves it, U2 edits it           | U1 moves (wave 1), U2 edits (wave 2) |
| `tools/repo-sync/.config/mise/tasks/setup/configs/*`               | U1 moves them, U2 edits them       | U1 moves (wave 1), U2 edits (wave 2) |
| `.config/**` synced copies, `.config/mise/conf.d/**`, `.gitignore` | generated by the sync              | U6 only                              |
| `.config/bootstrap.toml`                                           | the values file of this repository | U6 only                              |
| `README.md`, `CLAUDE.md`, `docs/**`                                | human-facing docs                  | U5 only                              |

## Waves

- Wave 1: U1 alone. It moves every file, so no other unit can run with it.
- Wave 2: U2 alone. It edits the engine files that U1 moved.
- Wave 3: U3 alone. Its tests need the engine of U2.
- Wave 4: U4, the review of U1 to U3.
- Wave 5: U5, the docs.
- Wave 6: U6, the sync of this repository and the final gate.

## Wave gate

```text
mise x -- mise run code:precommit
```

plus the wave review, plus every report read for `UNRESOLVED:`. Every line here
must be green before wave 1. `test:repo-sync` is not a gate line, because it
exists only after U3: it is in the Verification of U3 and U6.

## After landing

| Step                    | Mode | Notes                                                           |
| ----------------------- | ---- | --------------------------------------------------------------- |
| archive the plan folder | run  | `plan-management archive`, only when the folder has no open gap |
| `mise run code:graph`   | run  | refreshes the graphify graph of the main checkout               |

## Gates the orchestrator keeps

- U6 syncs this repository from the commit of the run. Pass: `repo-sync`
  prints `using bootstrap at <rev>` with the short hash of the run branch; a
  second `mise run setup:configs` shows only `unchanged`, `kept seed` or
  `skip env`; `.config/mise/conf.d/_base/` is gone.
- `mise x -- mise run test:repo-sync` passes after U6.
- `code:graph` runs once in a scratch repository with a git submodule that has
  graphify. Pass: the task runs in the submodule. If it skips there, record a
  gap and do not fix it in this run.

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

- The Flutter packs, and the removal of the mise `depends` on `android-sdk`
  from `flutter`: the tools of each Flutter pack are not decided.
- The member keys and the merge keys (`folder`, `branch.<branch>`): their
  design has open gaps.
- The fnox redesign: the run tasks are not decided.
- B06 (`env_conf_d`) and B07 (`.gitattributes`): related to `base`, but new
  behaviour.

## Parked

- The Flutter packs (`flutter-ios`, `flutter-android`, `flutter-web`,
  `flutter-macos`, `flutter-win`, `flutter-linux`) and the removal of the mise
  `depends` on `android-sdk` from `tools/flutter/tools.toml`.
- The member keys and the merge keys of `docs/design/bootstrap-config.md`:
  the self member with `folder = "."`, a submodule without a `[member.*]`
  table, and the `code:merge:<branch>` tasks that the sync writes.
- The fnox redesign of `docs/design/fnox.md`, with the secret helper
  `run_with_secrets` and its rule for `fnox` and `doppler` (D19).
- A key of the values file that removes a tool of a pack (deferred by the user).
- `product_name`: no file reads it after the fnox redesign.

## Run log

| Wave | Unit                          | Model | Round | Outcome     | Detail                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | Commit  |
| ---- | ----------------------------- | ----- | ----- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| 0    | preflight                     | —     | 1     | pass        | doctor: mise 2026.10.11, graphify CLI present, graph in main checkout; no `.config/vwf.yaml` (defaults: review_round_cap 4); mempalace down, journal not written; wave gate `code:precommit` green                                                                                                                                                                                                                                                                                                                                                                          | —       |
| 0    | format-check                  | —     | 1     | skipped     | why: no `covers:`, the plan reads no blueprint artifact                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | —       |
| 0    | stack conventions             | —     | 1     | skipped     | why: no `code` unit in the Units table                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | —       |
| 0    | wave order                    | —     | 1     | pass        | sequence: W1 U1 (edit), W2 U2 (edit), W3 U3 (edit), W4 U4 (review), acceptance/ux, reconcile, W5 U5 (edit, docs), W6 U6 (edit, gates)                                                                                                                                                                                                                                                                                                                                                                                                                                       | —       |
| 1    | U1 Split the base tool        | opus  | 1     | pass        | edit: base split into mise, repo-sync, git, pre-commit, dprint, linter, _security_base, gitleaks, grype, osv-scanner; packs.toml; claude not root_only, code:graph moved, setup:ai skips in a submodule; default_envs in every tools.toml. GAP: `setup/configs/sync:20` still names tools/base (U2 fixes it); the check `rg tools/base` skips hidden dirs. GAP: `code:precommit --all-files` skips untracked files; format and lint ran on the new files by hand                                                                                                            | —       |
| 1    | R1                            | opus  | 1     | findings(6) | CONTRACT clean, RULINGS clean. U1: `code/sec/{gitleaks,grype,osv-scanner}:5` add `#MISE hide=true`, which edit 3 does not name → loop to U1. 5 Docs findings in U5 Owns → handed to U5 as DOCS FALSIFIED: `docs/sync/index.md:146` (keys table lacks default_envs), `README.md:44` (manifest example lacks default_envs), `docs/sync/claude-tool.md:22-25,41` (code/graph, setup:ai submodule skip), `docs/sync/migrate.md:68` (M4 deletes setup:ai in a submodule), `docs/design/mise.md:346` (tools/base path)                                                            | —       |
| 1    | U1 Split the base tool        | opus  | 2     | pass        | edit: removed `#MISE hide=true` from the three `code/sec/<tool>` leaves (R1 finding)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | c95e4ff |
| 1    | R1                            | opus  | 2     | pass        | CONTRACT clean, RULINGS clean; round-1 finding resolved                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | —       |
| 2    | U2 Teach repo-sync the tables | opus  | 1     | pass        | edit: `_rs_values_sel`, `_rs_pack`, `_rs_selected`; default_envs read; env inheritance rewritten, D5/D6/D8 errors; `_RS_DIRECT`, `_base` gone; setup aliases in mise; merge path in git; REPO_SYNC_NO_AUTOUPDATE; sync loads tools/repo-sync. DECIDED: env names dev/ci/prod/stage in envs; `env = []` and unknown keys stop with code 2; selection order base, packs, tools. GAP: a tool selected twice gets the union of its environments; D5 envs check runs after the `environments` filter; pack/tool clash checked for every pack; `_rs_envs` accepts dev and ci only | —       |
| 2    | R2                            | opus  | 1     | findings(5) | CONTRACT clean, RULINGS clean. 5 Docs findings in U5 Owns → handed to U5 as DOCS FALSIFIED: `docs/sync/index.md:165` (envs and dependency-only rule), `docs/design/mise.md:146,154` (envs row, step 1), `docs/design/packs.md:52` (list format), `docs/design/packs.md:68` (pack envs), `docs/design/bootstrap-config.md:16,59` (tools as list); no loop                                                                                                                                                                                                                    | —       |

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-21-packs-and-base-split

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
