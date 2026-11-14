# Plan — R1a: the .config/bootstrap/ folder, the key schema and the config verb (2026-11-14)

## Goal

After this plan lands, the engine reads and writes the values of a repo only in
`.config/bootstrap/`: `repo.toml` (repo-level facts), `tools.toml` (the tool
selection, tool values and `[overrides.tools.*]`), `releases.toml` (the schema
only, in this plan) and `repo-generated.tsv` (the sync manifest), the three TOML
files
with `format = 3` (D74). A Go schema validates every key, and `check` refuses an
unknown
one. `bootstrap config get|set|unset|show --format json|yaml|toml` reads and
writes the values; the old `config` command becomes `settings`. Members are
`[members.<name>]` with `submodule`, `folder` or `sibling`, a `url`, and a
`[home]` back-link in each member, and no bootstrap file reads a vwf file.
`init` migrates the old `.config/bootstrap.toml`; every other command stops
with code 2.

This is phase R1a of
`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`.
It reverses that record's D2 ("contract v1 ships inside 3.0.0"): R1 ships as
`bootstrap@4.0.0`, after 3.0.0 (the selection model) and 3.1.0 (the TUI). The
docs unit records the reversal in the record. R1b (the contract commands) and
R1c (the task ports) are separate plans that require this one.

## Facts the survey established

- `docs/plans/2026-11-13-selection-model` is RUNNING (waves 1-3 green) and
  makes the values file `format = 2`, one `[tools.<name>]` table per tool, no
  `[packs.*]`, with `core.Plan`/`core.Apply` in `cli/internal/core`. Its
  landing is tagged 3.0.0 at once. `docs/plans/2026-11-13-tui` (APPROVED,
  priority 40, 3.1.0) requires it and owns `cli/`. This plan requires the TUI
  plan, so it runs on the code both leave.
- The engine is about 15k lines of Go in `cli/`. Commands are dispatched by
  hand in `cli/internal/cli/cli.go` (`Usage`, a flag-allow table, `parse`, a
  `switch cmd`). `--json` exists only on `questions`. Exit codes are in
  `cli/internal/fail` (0, 1, 2, 3, 4, 5).
- The path `.config/bootstrap.toml` is a literal at
  `cli/internal/core/core.go:123-125`
  in the selection model's code (D73; `BOOTSTRAP_VALUES`, with the temp values
  copy at `cli/internal/core/plan.go:314` and `:383`),
  `cli/internal/repo/repo.go`
  (about 194), `repo/file.go` (about 374) and `repo/rules.go` (about 360). The
  overrides path is derived in `values.OverridesPath`; the manifest path is
  `repo/repo.go:45`; `bootstrap_commit` is written by
  `target.Env.RecordCommit`.
- TOML is read with BurntSushi (comments lost) and edited with in-process yq
  (comments kept). `oldform` reports old forms with code 2, and `init`
  migrates them.
- No machine schema exists today; an unknown top-level key passes.
- `bootstrap config` today prints the engine settings from
  `~/.config/bootstrap/config.toml` (`cli/internal/cli/config.go`).
  `bootstrap check` exists.
- Tool scripts that read the values file: `repo/tools/git/tasks/merge`,
  `git/scripts/derive`, `git/scripts/check`, `claude/tasks/setup-ai`,
  `mise/tasks/code-lint-bootstrap`; `mise/tasks/code-worktrees` reads
  `.config/vwf.yaml` and `.config/vwf-membership.yaml`. About ten more files
  name the path in a comment or message.
- 11 example values files sit under `repo/examples/`. `test/repo` has about 60
  cases that name the values file.
- This repo's own `.config/` holds synced copies of the tool tasks and its own
  old-form `.config/bootstrap.toml`. They keep working with the old engine
  until the user migrates this repo after the 4.0.0 tag.
- Gates: `code:format`, `code:lint`, `test:go`, `test` (go, repo, mac,
  macos-tools), `code:sec`. Commit types are the ten of
  `.config/git-conventional-commits.yaml`, with no scopes.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file with Write, Edit, `sed` or a
script.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    QUESTION Q<n>: <the question, readable alone>
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence> [| <option 3> | <option 4>]
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

The three QUESTION lines repeat per question, `n` counting from 1 in the
block, with two to four options; a block with no question carries the one
line `QUESTION: none`. A unit completes every edit that does not depend on an
answer, then returns all its open questions in one block.

## Out of scope

- The Mac side (`macos.toml`, `mac` commands).
- The contract commands `contract`, `describe`, `catalog`, `check --json` and
  `--json` on `add`/`remove`/`init`: plan R1b.
- The task ports as no-op defaults: plan R1c.
- The `git:*` task names and the LSP blocks: phase R2.
- Moving `.config/releases.yaml` into `releases.toml`: phase R3 (D10).
- This repo's own `.config/` files: the user migrates them by hand after the
  tag.

## Parked

- R1b and R1c are planned next, each requiring this folder.
- The claude-plugins copy of the design record needs the same amendments as
  U8 makes here.
- **Hand steps for the user:** tag `bootstrap@4.0.0` at once after this plan
  lands (D69); then run
  `bootstrap init` in this repo and in each consumer repo.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-14-config-folder

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
