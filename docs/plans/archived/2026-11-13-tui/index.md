# Plan — The TUI: bootstrap tui and bootstrap mac tui (2026-11-13)

## Goal

Add the full-screen TUI `bootstrap tui` (repo) and `bootstrap mac tui` (Mac),
as the design "Bootstrap TUI Screens" (version 47) says. A copy of the design is
`screens.txt` in this folder. In a terminal, `init` and `mac init` open the
TUI. The TUI is for a person; the CLI is for an agent.

The TUI calls `cli/internal/core` of plan 2026-11-13-selection-model (its D23):
the change list, `Plan`, `Apply` and `Detect`. This plan requires that plan.

**After the landing, by hand** (not in `after_landing`, because a plan never
carries a release): the user tests the TUI, then tags `bootstrap@3.1.0` (D8).

## Facts the survey established

The names (D10 a): plan 1a is `2026-11-12-catalog-engine`, plan 1b is
`2026-11-12-catalog-data`, plan 2a is `2026-11-13-selection-model`, and plan 2c
is this plan. The line numbers are from `main` before plans 1a, 1b and 2a
landed (D7 j). Find
the code by the name of its function.

- Dispatch: `Usage` at `cli/internal/cli/cli.go:31`, `Run` at `:150`, the
  `allowed` map of the flags of each command at `:181-193`, the switch at
  `:242-263`.
- The terminal check: `a.interactive()` at `cli/internal/cli/git.go:121`.
- `init` at `cli/internal/cli/init.go:132`. Plan 2a adds `--no-tui` and moves
  the core of `finish` into `core.Plan` and `core.Apply`.
- The Mac package prefixes `brew:`, `brew-cask:` and `mas:` at
  `cli/internal/cli/packages.go:48-50`; plan 2a can move the scan into
  `cli/internal/core` (its D23).
- No package `cli/internal/tui` exists. Plan 1a adds Bubble Tea v2
  (`charm.land/bubbletea/v2`) and Bubbles v2 (`charm.land/bubbles/v2`) for the
  prompt of its D14.
- `TestEngineNamesNoTool` (`cli/internal/source/notools_test.go`) fails on a
  string literal of the engine that names a tool folder. Common words such as
  "network", "pages", "node", "python", "claude", "cli" and "age" are tool
  names, so the TUI takes each text of a tool from the data (D2).
- The gates: `test:go`, `test:repo`, `test:mac`, `test:macos-tools`,
  `code:format`, `code:lint` and `code:sec`. The commit types: feat, fix, perf,
  refactor, revert, test, ops, docs, merge, wip, with no scope. The only CI
  workflow is `release-engine.yml`.
- The docs that the `tui` command makes false: `docs/design/cli.md` ("The files
  of `cli/`", "The command line", "`init`", "The exit codes"), `README.md`,
  `CLAUDE.md`, `docs/sync/index.md` and `docs/setup/index.md`.

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
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence>
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

A block with no question carries the one line `QUESTION: none`.

## Out of scope

- Automated tests of the TUI screens: the user tests by hand (D3).
- `--exclude` in the TUI (D7 c).
- The change to the Mac tool `git` (ruling 44 of `screens.txt`): plan
  2026-11-12-catalog-data does it (D7 d).
- A bump of `cli/VERSION` without a release (D8).

## Parked

- B29: the automated tests of the TUI (golden `View` strings at 100x25, no
  `teatest`), P1 in the backlog.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-11-13-tui

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
