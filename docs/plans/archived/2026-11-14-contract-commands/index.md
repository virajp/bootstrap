# Plan — R1b: the contract commands and the JSON envelope (2026-11-14)

## Goal

After this plan lands, a plugin reads and drives the bootstrap CLI only
through one versioned JSON envelope on stdout (D1): `bootstrap contract`
(versions, commands, ports), `describe` (repo facts, tools, members, ports),
`catalog` (the tool catalog of the clone), `check --json` (each stage and its
findings), and `--json` on `questions`, `config` and the six write commands
(`add`, `remove`, `replace`, `init`, `detect`, `sync`). All human lines go to
stderr, and the exit code equals the envelope `code`.

This is phase R1b of
`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`. It requires R1a
(`docs/plans/2026-11-14-config-folder`). It has two reversals:

- **D16** reverses R1a D69. R1a, R1b, R1c and R2 land before one
  `bootstrap@4.0.0` tag, not a tag after each phase. It restores R1a D36.
- **D6** reverses R1a D23 and D52 for `config set` and `unset`: one fetch rule
  (D5) for every catalog read.

## Facts the survey established

The survey ran in the session that planned R1a and is recorded in
`docs/memory/handoff/bootstrap-plugins-split.md`. The code is the selection
model's (`.worktrees/all-2026-11-13-1753/cli/`); R1a and the TUI plan change it
before this plan runs, so line numbers are approximate.

- `docs/plans/2026-11-13-selection-model` is RUNNING; `2026-11-13-tui` and
  `2026-11-14-config-folder` (R1a) are APPROVED behind it. All three own
  `cli/`. This plan requires R1a, so it runs on the code all three leave.
- The only JSON today is `questions --json` (`cli/internal/cli/init.go`, about
  line 380, `printQuestions`, shared by `mac questions`): a bare array on
  stdout, with no envelope and no version. The bash cases assert it at about
  `.config/mise/tasks/test/repo:3198-3199`, `:3836-3844` and
  `.config/mise/tasks/test/mac:1992`, `:2428`; `cli/cmd/bootstrap/main_test.go`
  holds only usage strings. The flag table in `cli.go` is shared by the Mac
  side. No file in `claude-plugins/plugins` or `claude-plugins/.claude` calls
  it.
- The binary reports `main.version`, set by ldflags; a local build reports
  `"dev"` (`cli/cmd/bootstrap/main.go`). `cli/VERSION` is the minimum engine a
  clone needs, not the binary's version.
- `sync` takes no `--yes` and applies with no consent; `init` always calls
  `ask()` for pending questions.
- R1a adds `bootstrap config get|set|unset|show --format json|yaml|toml`;
  `config get` prints `{"key","value","source"}` in every format (R1a D41).
- `check` (`cli/internal/cli/cli.go` → `repo.Run(--check)`,
  `cli/internal/repo/repo.go` `check()`) runs the stages envs, values, old
  forms, env rules, overrides, members and scripts. It skips a stage whose
  input failed, prints findings at once and keeps only the highest code.
  `fail.Collector` exists but `check` does not use it.
- Write commands go through `core.Plan` → `core.Apply`
  (`cli/internal/core/plan.go`, `PlanResult`: Changes, Files, Skipped, Pending,
  Warnings, Notes). The per-target outcomes of a sync are only counted inside
  `report.Report` (`cli/internal/report/report.go`, `Outcome`), never
  returned.
- stdout carries human output today (change list, trees, prompts).
- Catalog data, in `cli/internal/source/`: `Source.Open()` returns the
  `Index` (`check.go`) of `Manifest`s (`manifest.go`);
  `Source.ReadCategories()` (`catalog.go`) and `Source.ReadPacks()`
  (`packs.go`) read the categories and the packs.
- No interface version exists; `oldform.Format` and `cli/VERSION` with
  `core.CheckVersion` are the patterns.
- R1a D52 writes the fetch time to a file in the cache root after each
  successful fetch. `--clone` and `BOOTSTRAP_CLONE` skip the TTL.
- Docs that describe the changed output: `docs/design/cli.md`,
  `docs/sync/vwf.md`, `docs/sync/index.md`, `docs/setup/index.md`,
  `docs/setup/migrate-dotfiles.md`, `README.md`, `CLAUDE.md`. Stale points in
  `cli.md`: `list` still describes packs, `detect --no` is gone, and
  `init [tool…]` now refuses names.
- No backlog item of the `bootstrap` project covers R1b. B28 overlaps R2.
- Gates: `code:format`, `code:lint`, `test:go`, `test` (go, repo, mac,
  macos-tools), `code:sec`. This repo has no `code:precommit`. Commit types
  are the ten of `.config/git-conventional-commits.yaml`, with no scopes.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file — the file a package
manager or a toolchain manager writes to pin what it resolved — with Write,
Edit, `sed` or a script: a lock file changes only as the output of the CLI
that owns it, the command the stack's package-manager conventions name, and
the unit keeps what that CLI wrote. A unit whose manifest change moves a lock
file lists that lock file in its `owns` beside the manifest.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    QUESTION Q<n>: <the question, readable alone>
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence> [| <option 3> | <option 4>]
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

The three QUESTION lines repeat per question, `n` counting from 1 in the
block, with two to four options; a block with no question carries the one
line `QUESTION: none`. A question is a point the plan does not decide and the
unit cannot settle inside its scope; the unit never proceeds on an
assumption.

A unit completes every edit that does not depend on an answer, never builds on
a part an answer can change, then returns all its open questions in one block.
Its edits stay uncommitted in the worktree. The orchestrator asks the user,
records each answer in `decisions`, and resumes the unit with the question ids
and answers; the resumed unit applies them and returns the whole block again.

## Out of scope

- The task ports as no-op defaults: plan R1c. This plan only names them (D3).
- The `lsp` field of `describe`, `bootstrap lsp` and `code:lsp:check`: phase
  R2 (D10).
- The `git:*` task names: phase R2.
- The Mac side (`macos.toml`, `mac` commands): no contract command reads it.
- A version bump or a tag: the user tags `bootstrap@4.0.0` after R1a, R1b, R1c
  and R2 (D16).
- This repo's own `.config/` files: the user migrates them by hand after the
  tag.

## Parked

- R1a's folder still says "tag 4.0.0 at once after R1a" (its D69, U8, U9 and
  Parked). Amend it to D16 before it runs:
  `docs/memory/gaps/2026-11-14-one-4-0-0-tag-parked.md`.
- The claude-plugins copy of the design record needs R1a's amendments and
  D16; this plan edits only the bootstrap copy.
- The R2 interview decides whether R2 keeps the old task names as aliases,
  since no release comes between R2 and 4.0.0.
- R1c (the task ports) and R2 are planned next; R1c requires R1a.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-14-contract-commands

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
