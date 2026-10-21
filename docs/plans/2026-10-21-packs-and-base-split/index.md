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

**COMPLETE**

COMPLETE 2026-10-21 — c95e4ff, 39139fe, ce30bf1, 94c87f8, 1076790, 99bfe52,
b3667de, f633afe, 94b19ea, e9c8419

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

| Id | Wave | Unit file                    | Kind   | Owns                                                                                                                                                                          | Depends on | Status | Commit  |
| -- | ---- | ---------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------ | ------- |
| U1 | 1    | [01-split.md](01-split.md)   | edit   | `tools/**` (moves only for the files of U2)                                                                                                                                   | —          | green  | c95e4ff |
| U2 | 2    | [02-engine.md](02-engine.md) | edit   | `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`, `tools/repo-sync/.config/mise/tasks/setup/configs/_default`, `tools/repo-sync/.config/mise/tasks/setup/configs/sync` | U1         | green  | 39139fe |
| U3 | 3    | [03-tests.md](03-tests.md)   | edit   | `.config/mise/tasks/test/repo-sync`, `.config/repo-sync/pre-commit.yaml`                                                                                                      | U2         | green  | ce30bf1 |
| U4 | 4    | [04-review.md](04-review.md) | review | —                                                                                                                                                                             | U1, U2, U3 | green  |         |
| U5 | 5    | [05-docs.md](05-docs.md)     | edit   | `docs/sync/**`, `README.md`, `CLAUDE.md`, `docs/design/*.md`                                                                                                                  | U4         | green  | 94b19ea |
| U6 | 6    | [06-gates.md](06-gates.md)   | edit   | `.config/bootstrap.toml`, every synced or generated copy in `.config/` and at the root (`.gitignore`)                                                                         | U5         | green  | e9c8419 |

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

## Gaps surfaced during execution

- G1 (U4 round 1, plan gap, non-blocking): the first sync after this change
  fails in every existing consumer. The old synced
  `.config/mise/tasks/setup/configs/sync:20` sources
  `$REPO_SYNC_CLONE/tools/base/.../repo-sync`, which U1 removed, and a shim at
  `tools/base/` is not possible (the pack/tool clash check stops on it). U6
  edit 2 cannot run as written. Assumption taken: U6 does the first sync by
  sourcing `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync` from the
  clone directly; then `mise run setup:configs` works. No upgrade path for
  other consumers (D6: bootstrap is unreleased).
- G2 (U4 round 1, plan gap, non-blocking): 01-split.md edit 3 tells the
  `code:sec:<scanner>` leaves to print only `print_subheader` under the
  orchestrator's header, which breaks the task-output rules 2 and 5 of
  `CLAUDE.md`. Assumption taken: the plan's edit 3 stands.

- G3 (U4 round 2, plan gap, non-blocking): D15 replaces `conf.d/_base/` with
  `conf.d/<tool>/`, and the plan does not settle the load order. mise loads
  `_repo/` before `mise/`, `pre-commit/` and the other tool folders, and the
  later file wins. So an override in `_repo/mise.<env>.toml` of `os`,
  `PRE_COMMIT_HOME` or the `setup` alias now loses to the generated shared
  fragments; before, `_repo` came after `_base` and won. Assumption taken: no
  change in this run; a later plan rules the order (for example a folder name
  that sorts last).

- G4 (U4 round 4, review residual, contested, non-blocking):
  `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync:831` (U2) —
  `_rs_context` adds the `[[precommit]]` entries of every tool in `_RS_TOOLS`,
  but a values-file `env` can keep a tool out of dev (D1). Example:
  `[tools.github] env = ["ci"]` puts actionlint only in `mise.ci.toml`, but the
  dev pre-commit config still gets `mise x -- actionlint`. The loop ended at
  the round cap (4), and the convergence guard also tripped (1 finding in
  round 3, 1 in round 4, not the same finding). Diagnosis: each fix of the env
  rules exposed one more place that the new D1 rule reaches; the plan did not
  list every consumer of a tool's environments. Proposed fix: add a tool's
  pre-commit entries only when `_RS_TOOL_ENVS` has `<tool>:dev`.

- G5 (U6, plan gap, non-blocking): `code:graph` moved to the `claude` tool
  (D13), and this repository selects only the `base` pack (U6 edit 1). So the
  sync drops the `graphify-refresh` post-commit hook from this repository's
  `.config/pre-commit-config.yaml`, and `.config/mise/tasks/code/graph` stays as
  a stale synced copy with the header `tools/base/...` (the sync never
  deletes). U6 asked whether to add `[tools.claude]` or delete the stale copy.
  Ruling taken from the plan: U6 edit 1 says "no `[tools]` and no `[packs]`
  entry", so nothing is added and nothing is deleted. A later change decides
  whether this repository selects `claude`.

## Run log

| Wave | Unit                          | Model | Round | Outcome      | Detail                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | Commit  |
| ---- | ----------------------------- | ----- | ----- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| 0    | preflight                     | —     | 1     | pass         | doctor: mise 2026.10.11, graphify CLI present, graph in main checkout; no `.config/vwf.yaml` (defaults: review_round_cap 4); mempalace down, journal not written; wave gate `code:precommit` green                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | —       |
| 0    | format-check                  | —     | 1     | skipped      | why: no `covers:`, the plan reads no blueprint artifact                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | —       |
| 0    | stack conventions             | —     | 1     | skipped      | why: no `code` unit in the Units table                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | —       |
| 0    | wave order                    | —     | 1     | pass         | sequence: W1 U1 (edit), W2 U2 (edit), W3 U3 (edit), W4 U4 (review), acceptance/ux, reconcile, W5 U5 (edit, docs), W6 U6 (edit, gates)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | —       |
| 1    | U1 Split the base tool        | opus  | 1     | pass         | edit: base split into mise, repo-sync, git, pre-commit, dprint, linter, _security_base, gitleaks, grype, osv-scanner; packs.toml; claude not root_only, code:graph moved, setup:ai skips in a submodule; default_envs in every tools.toml. GAP: `setup/configs/sync:20` still names tools/base (U2 fixes it); the check `rg tools/base` skips hidden dirs. GAP: `code:precommit --all-files` skips untracked files; format and lint ran on the new files by hand                                                                                                                                                                                                                                            | —       |
| 1    | R1                            | opus  | 1     | findings(6)  | CONTRACT clean, RULINGS clean. U1: `code/sec/{gitleaks,grype,osv-scanner}:5` add `#MISE hide=true`, which edit 3 does not name → loop to U1. 5 Docs findings in U5 Owns → handed to U5 as DOCS FALSIFIED: `docs/sync/index.md:146` (keys table lacks default_envs), `README.md:44` (manifest example lacks default_envs), `docs/sync/claude-tool.md:22-25,41` (code/graph, setup:ai submodule skip), `docs/sync/migrate.md:68` (M4 deletes setup:ai in a submodule), `docs/design/mise.md:346` (tools/base path)                                                                                                                                                                                            | —       |
| 1    | U1 Split the base tool        | opus  | 2     | pass         | edit: removed `#MISE hide=true` from the three `code/sec/<tool>` leaves (R1 finding)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | c95e4ff |
| 1    | R1                            | opus  | 2     | pass         | CONTRACT clean, RULINGS clean; round-1 finding resolved                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | —       |
| 2    | U2 Teach repo-sync the tables | opus  | 1     | pass         | edit: `_rs_values_sel`, `_rs_pack`, `_rs_selected`; default_envs read; env inheritance rewritten, D5/D6/D8 errors; `_RS_DIRECT`, `_base` gone; setup aliases in mise; merge path in git; REPO_SYNC_NO_AUTOUPDATE; sync loads tools/repo-sync. DECIDED: env names dev/ci/prod/stage in envs; `env = []` and unknown keys stop with code 2; selection order base, packs, tools. GAP: a tool selected twice gets the union of its environments; D5 envs check runs after the `environments` filter; pack/tool clash checked for every pack; `_rs_envs` accepts dev and ci only                                                                                                                                 | 39139fe |
| 2    | R2                            | opus  | 1     | findings(5)  | CONTRACT clean, RULINGS clean. 5 Docs findings in U5 Owns → handed to U5 as DOCS FALSIFIED: `docs/sync/index.md:165` (envs and dependency-only rule), `docs/design/mise.md:146,154` (envs row, step 1), `docs/design/packs.md:52` (list format), `docs/design/packs.md:68` (pack envs), `docs/design/bootstrap-config.md:16,59` (tools as list); no loop                                                                                                                                                                                                                                                                                                                                                    | —       |
| 3    | U3 Test repo-sync             | opus  | 1     | pass         | edit: `test:repo-sync` (9 cases, all pass, bash 3.2 and 5) and the local hook `test-repo-sync` on `^tools/`. DECIDED: `#MISE tools.gomplate`; each case in a subshell; error cases use real tools or a changed copy. DOCS FALSIFIED: `CLAUDE.md:9` (no tests) → U5. GAP: the hook is active only after U6 renders the pre-commit config                                                                                                                                                                                                                                                                                                                                                                     | ce30bf1 |
| 3    | R3                            | opus  | 1     | pass         | CONTRACT clean, RULINGS clean                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | —       |
| 4    | U4                            | opus  | 1     | findings(1)  | security, range 31fcdb7..ac6b930; files: tools/** (U1 c95e4ff; U2 39139fe for repo-sync, configs/sync), .config test+hook (U3 ce30bf1); docs: commits excluded. [low] `_security_base/.../code/sec/_default:27` (U1): scan gate fails open when no executable `code:sec:<scanner>` task is found. VERDICT approve. Engine: no high-confidence vulnerabilities                                                                                                                                                                                                                                                                                                                                               | —       |
| 4    | U4                            | opus  | 1     | findings(4)  | review, range 31fcdb7..ac6b930. [low] `tools/dprint/tools.toml:10` (U1) format-first comment false; [low] `tools/mise/.../_scripts/helpers:246` (U1) setup:ai root-only comment; [low] `repo-sync:574` (U2) two env-name lists; [low] `test/repo-sync:76` (U3) exit code only, no message check. SPEC met; API COMPAT n/a. PLAN GAPS: first sync of every consumer fails (old synced `setup/configs/sync:20` sources tools/base); leaf `print_subheader` of edit 3 vs CLAUDE.md rules 2 and 5. VERDICT approve. 6 engine findings not kept by the reviewer (staged mise startups, duplicate pre-commit-hooks repo, hardcoded tool names, hook `files` filter, leaf header = plan gap, sync shim = plan gap) | —       |
| 4    | U1 Split the base tool        | opus  | 3     | pass         | edit (U4 loop-back): `code:sec` fails when a scanner task of the pack security is missing or not executable (security); dprint and helpers comments corrected                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | 94c87f8 |
| 4    | U2 Teach repo-sync the tables | opus  | 2     | pass         | edit (U4 loop-back): `_RS_ENV_NAMES` is the one list; `_rs_envs` reuses `_rs_env_name`; values file still accepts dev and ci only                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | 1076790 |
| 4    | U3 Test repo-sync             | opus  | 2     | pass         | edit (U4 loop-back): each error case also matches the engine message with grep -F                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | 99bfe52 |
| 4    | U4                            | opus  | 2     | pass         | security, range 31fcdb7..9fdb22f; round-1 fail-open fix (94c87f8) holds; engine: no high-confidence vulnerabilities. VERDICT approve                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | —       |
| 4    | U4                            | opus  | 2     | findings(2)  | review, range 31fcdb7..9fdb22f; round-1 fixes hold. [medium] `repo-sync:705` (U2) mise template ignores `_RS_TOOL_ENVS`: dev-only sections and folder sections not gated by the tool environments (D1); [low] `repo-sync:760,581` (U2) env names still hardcoded in the failure loop and `_rs_envs`. PLAN GAP G3: `conf.d/_repo/` now loads before the per-tool folders, so a repo override loses. SPEC met; API COMPAT n/a. VERDICT approve. Guard: 4 → 2, no resurfaced path:line                                                                                                                                                                                                                         | —       |
| 4    | U2 Teach repo-sync the tables | opus  | 3     | pass         | edit (U4 round-2 loop-back): `REPO_SYNC_IN` gates folder, dev-only and `[mise_env.*]` sections by `_RS_TOOL_ENVS`; failure loop over `$_RS_ENVS`; `_rs_envs` check simplified. GAP: the reviewer's node example is not a bug (node inherits dev through linter); checked with a scratch tool                                                                                                                                                                                                                                                                                                                                                                                                                | b3667de |
| 4    | U4                            | opus  | 3     | pass         | security, range 31fcdb7..71fcfa5; engine: no high-confidence vulnerabilities; flag pass-through to consumer `code:sec:*` tasks not a security finding. VERDICT approve                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | —       |
| 4    | U4                            | opus  | 3     | findings(1)  | review, range 31fcdb7..71fcfa5; round-2 fixes hold. [low] `repo-sync:375` (U2) an `env` item with whitespace passes the check and drops an environment silently (D5). SPEC met; API COMPAT n/a. VERDICT approve. Guard: 2 → 1, no resurfaced path:line                                                                                                                                                                                                                                                                                                                                                                                                                                                      | —       |
| 4    | U2 Teach repo-sync the tables | opus  | 4     | pass         | edit (U4 round-3 loop-back): `env` items split on commas only and checked by `_rs_env_name`; a bad or empty item stops with code 2. GAP: the case is not in test:repo-sync (U3 Owns)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | f633afe |
| 4    | U4                            | opus  | 4     | pass         | security, range 31fcdb7..00e083c; engine: no high-confidence vulnerabilities. VERDICT approve                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | —       |
| 4    | U4                            | opus  | 4     | findings(1)  | review, range 31fcdb7..00e083c; round-3 fix holds. [medium] `repo-sync:831` (U2) `_rs_context` adds the `[[precommit]]` entries of a tool that `env` keeps out of dev → **contested**: cap 4 reached and the guard tripped (1 → 1); gap G4. SPEC met; API COMPAT n/a. VERDICT approve. U4 green with residual                                                                                                                                                                                                                                                                                                                                                                                               | —       |
| 4    | R4                            | opus  | 1     | pass         | contract review of the U4 fix commits 94c87f8, 1076790, 99bfe52, b3667de, f633afe: CONTRACT clean, RULINGS clean; no late re-run                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | —       |
| 4    | wave gate                     | —     | 1     | pass         | `code:precommit` green; no UNRESOLVED                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | —       |
| —    | acceptance                    | —     | 1     | skipped      | why: no `covers:`, no acceptance criteria                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | —       |
| —    | ux                            | —     | 1     | skipped      | why: no `covers:`, no screens                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | —       |
| —    | renders                       | —     | 1     | skipped      | why: the ux stage did not run, no RENDER lines                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | —       |
| —    | reconcile                     | —     | 1     | skipped      | why: no `covers:` (no stamps, registry, environment or harness) and no `code` unit (nothing to persist); mempalace down                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | —       |
| 5    | U5 Docs                       | opus  | 1     | pass         | edit: tool docs moved to docs/sync/tools/, base.md removed, 10 new tool docs, packs/base.md and packs/security.md; index, vwf, migrate, CLAUDE.md, README.md, design docs (bootstrap-config, mise, packs, tools, fnox D19) updated; all 13 handed DOCS FALSIFIED lines applied. DECIDED: `environments` documented as dev and ci only (the engine rejects prod and stage); no migration item for the list format (D6); packs docs say not to write [packs.base] or [packs.security]. GAP: untracked new docs checked with dprint and lint by hand                                                                                                                                                           | —       |
| 5    | R5                            | opus  | 1     | findings(11) | CONTRACT clean, RULINGS clean. Docs: `migrate.md:14`, `README.md:164`, `index.md:220` leave out the G1 cause (old synced `setup/configs/sync` loads tools/base); `index.md:326` Migration items of the base tools no longer reached; `migrate.md:93` stale folder; `packs.md:7,148` stale `required` text; fold width broken at `README.md:154,229`, `CLAUDE.md:109`, `index.md:129-131,202-203`, `vwf.md:97-98,120` → loop to U5                                                                                                                                                                                                                                                                           | —       |
| 5    | U5 Docs                       | opus  | 2     | pass         | edit (R5 loop-back): G1 cause stated in index step 2, migrate row and README, all pointing to step 2 (no new procedure, D6); step 5 and M6 run the Migration of the base pack tools; migrate tools/ link; packs.md `required` text; refolds in README, CLAUDE.md, index, vwf, packs, kotlin                                                                                                                                                                                                                                                                                                                                                                                                                 | 94b19ea |
| 5    | R5                            | opus  | 2     | findings(3)  | round-1 findings resolved. **contested** (wave-review cap of 2 rounds): `docs/design/packs.md:30` says a pack names all its tools in `depends`, but `_security_base` is in `required`; `docs/sync/migrate.md:15` M15 first runs `setup:configs` with the old sync (G1) before step 2; `docs/sync/migrate.md:185` says the sync removes `conf.d/pre-commit/`, but the pre-commit tool now generates it. CONTRACT clean, RULINGS clean                                                                                                                                                                                                                                                                        | —       |
| 5    | wave gate                     | —     | 1     | pass         | `code:precommit` green; no UNRESOLVED                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | —       |
| 6    | U6 Gates                      | opus  | 1     | pass         | edit: `tools = ["base"]` removed; first sync with the new library from a scratch clone (G1): `using bootstrap at 855ab14`, 48/48, `conf.d/_base/` removed; `mise run init`; second sync: 40 unchanged, 4 kept seed, 1 skip env, 1 skip seed; test:repo-sync and code:precommit green. UNRESOLVED (add `[tools.claude]` for the code:graph hook, or delete the stale `code/graph`) → answered by the plan: U6 edit 1 rules no `[tools]` entry; recorded as gap G5. GAP: `skip seed` status accepted (seed does not apply)                                                                                                                                                                                    | e9c8419 |
| 6    | R6                            | opus  | 1     | pass         | CONTRACT clean, RULINGS clean; synced copies match their sources                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | —       |
| 6    | wave gate                     | —     | 1     | pass         | `code:precommit` green; no UNRESOLVED left (U6 question answered by the plan, G5)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | —       |
| —    | reconcile                     | —     | 1     | pass         | final wave gate `code:precommit` green over the finished tree                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | —       |
| —    | reconcile                     | —     | 1     | pass         | orchestrator gate 1: U6 sync printed `using bootstrap at 855ab14` (the run branch tip then); second sync only unchanged, kept seed, skip env, plus one `skip seed` (pnpm-workspace.yaml, no package.json); `conf.d/_base/` gone                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | —       |
| —    | reconcile                     | —     | 1     | pass         | orchestrator gate 2: `test:repo-sync` passes after U6 (all cases)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | —       |
| —    | reconcile                     | —     | 1     | pass         | orchestrator gate 3: `code:graph` in a scratch submodule runs and writes the submodule's own graphify-out/ (git-dir equals git-common-dir there). NOTE: line 25 `mise x -- which graphify` finds the global shim in a repo that declares no graphify, so the task says "refreshing" and the background update fails only in the log (moved code, unchanged; not fixed)                                                                                                                                                                                                                                                                                                                                      | —       |

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-21-packs-and-base-split

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
