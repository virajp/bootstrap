# Plan — A Mac [files] entry can write into the backup clone (2026-11-27)

## Goal

After this plan lands, a Mac `[files]` entry with `root = "backup"` writes
its target as a real file below the backup clone (the `backup` key of
`macos.toml`), not below `$HOME`. The sync skips such an entry when
`macos.toml` has no `backup`.

This is plan 2 of the skill-delivery work of 2026-11-27. Plan 3,
`2026-11-27-tool-skills`, uses it to ship the skills `mac-setup` and
`git-identity` into `macos-setup/.claude/skills/`, where Claude Code finds
project skills. No recorded decision is reversed.

## Facts the survey established

- A Mac `[files]` key is parsed in `Source.entries`
  (`cli/internal/source/inventory.go:182`), checked by `badRelPath`
  (`cli/internal/source/check.go:179`) and by `manifest.BadPath`
  (`cli/internal/manifest/manifest.go:63`, used at :126).
- `makePlan` (`cli/internal/macos/plan.go:23`) checks `when` and duplicate
  targets. Each resolution builds `s.home + "/" + it.target`:
  `cli/internal/macos/file.go:39, :44, :55, :78, :100`; the stale rule roots
  at `s.home` (`cli/internal/macos/macos.go:631`); `s.home` is set at
  `macos.go:269`.
- The `backup` key is read in `cli/internal/values/values.go:368-373`. The
  link rule is `state.backup` (`cli/internal/macos/backup.go:197`), called
  at `macos.go:363`; it expands `~` (:202), makes the clone absolute (:205),
  takes the targets of the plan (:213) and passes them to `linkPlan` (:40).
  `backupDir` is `dotfiles` (:22).
- The Mac manifest is `~/.cache/bootstrap/macos-generated.tsv`; a row is
  `<tool><TAB><path>` with the path relative to `$HOME`
  (`docs/design/cli.md` "The manifest and the stale rule").
- `when` with a missing value is false (`cli/internal/source/when.go:8-24`).
- Tests: `cli/internal/macos/backup_test.go` (`TestLinkPlan` :11),
  `cli/internal/source/source_test.go`,
  `cli/internal/source/notools_test.go:107`
  (`TestEngineNamesNoTool`), `e2e/mac/sync_test.go:1344-1410` (the backup
  cases), `e2e/mac/mac_test.go:119-122`, `:201`,
  `e2e/macostools/macostools_test.go:205`.
- Docs: `docs/design/tools.md` "The `[files]` table" (:206),
  `docs/design/mac.md` "The backup clone" (:360), "The manifest" (:432),
  "Where the Mac side writes" (:446), `docs/design/cli.md` "The manifest and
  the stale rule" (:1599).
- Gates, the commit convention and the project `bootstrap` (`cli/`) are as in
  plan 1, `docs/plans/2026-11-27-tool-selection/index.md`.

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
and answers; the resumed unit applies them and returns the whole block again
(the execute skill's `references/questions.md`).

## Out of scope

- The first real `root = "backup"` entries (`mac-setup`, `git-identity`):
  plan 3, `2026-11-27-tool-skills`.
- A commit in the clone: bootstrap never runs git in the clone (D6).

## Parked

none

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-27-mac-backup-target

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
