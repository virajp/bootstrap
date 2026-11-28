# Plan — Tool skills and the bootstrap section of CLAUDE.md (2026-11-28)

## Goal

After this plan lands, each consumer that selects `claude` gets a bootstrap
section at the top of its `CLAUDE.md` (Tasks, Dependencies, Gates), filled
from its selected tools, and the skills of its selected tools in
`.claude/skills/`: `typescript`, `pnpm`, `kotlin`, `gradle` and `swift`.
On the Mac, `~/.claude/skills/secrets/` and the npm hook ship, and the backup
clone (`macos-setup`) gets `.claude/skills/mac-setup/` and
`.claude/skills/git-identity/`.

Plan 3 of the skill-delivery work of 2026-11-27; it requires
`2026-11-27-tool-selection` and `2026-11-27-mac-backup-target`. The user
designed it over 2026-11-27 and 2026-11-28, one question at a time.

**Reversals.** (1) The 2026-11-14 split
(`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`) gave stack
knowledge to stackgen; tool knowledge, keyed by a bootstrap tool, now belongs
to bootstrap. (2) The handoff prompt
`docs/prompts/2026-11-27-mise-skill-and-flutter-test.md` (item 1) and D14 of
the claude-plugins plan `2026-11-26-stackgen-bootstrap-tasks` said that
bootstrap ships a repo mise skill; the user ruled that the task text goes into
the bootstrap section of `CLAUDE.md`, and no mise skill ships. U6 records both
(D12).

## Facts the survey established

- The engine already merges a Markdown file with one `section` (marker
  comments `bootstrap:start` / `bootstrap:end`, `comment = "<!--"`),
  renders a merge spec as a template with `.config` and `.all`
  (`cli/internal/source/inventory.go:364`,
  `cli/internal/macos/file.go:182-200`),
  and puts a new section at the top of the file
  (`cli/internal/format/section.go:115-118`). `TestSharedSection`
  (`cli/internal/format/shared_test.go:363`) covers a `CLAUDE.md`. The model
  spec is `macos/tools/claude/files/.claude/CLAUDE.md.merge.toml`.
- `when = "<tool>.<path>"` with a missing value is false
  (`cli/internal/source/when.go:8-24`), so `when = "claude.skills"` ships
  nothing when `claude` is not selected.
- `header.Has` reads the first 3 lines (`cli/internal/header/header.go:39-68`),
  so a header on lines 2-3 under `---` is found. The owned header is in the
  source only.
- No repo tool ships `CLAUDE.md` or a file below `.claude/` today. The repo
  `claude` tool has `[defaults.plugins]` and the task `setup/ai.py`, and no
  `[files]`.
- Gate tasks: dprint 05-dprint; linter 05-linter, 06-shellcheck; kotlin
  30-ktlint, 30-detekt; swift 50-swift-format, 80-swiftlint; flutter 10-dart,
  20-dart; python 50-pyrefly, 70-ruff, 90-uv; go 40-go, 45-gomod; android
  10-android; mise 60-bootstrap. Deps tasks: android 10, flutter 20, go 25,
  kotlin 30, node 40, python 50, swift 60.
- The Mac `claude` tool ships `.claude/CLAUDE.md` and `.claude/settings.json`
  as merge files; the settings spec has one hooks item, the claude-status
  `PostToolUse` block (:58-63). `1password-cli`, `age`, `glab`, `mas`,
  `setapp` and `macos-defaults` have no `[files]`. `macos/examples/macos.toml`
  selects all fourteen Mac tools of U3.
- The stackgen sources are in
  `$HOME/Projects/github.com/virajp/claude-plugins/plugins/stackgen/tools/<slug>/`;
  the claude-plugins plan `2026-11-26-stackgen-bootstrap-tasks` was running on
  2026-11-28 and edits them, so U1 and U3 read `origin/develop`.
- The audit of 2026-11-27 (three subagents) cut each stackgen entry by D5; its
  surviving rules are in `01-repo-tools.md` ("The skill texts").
- The consumers `linter` and `virajp.dev` have stackgen copies in
  `.claude/skills/` (pnpm, tsconfig, typescript, eslint, ux-gate and others),
  `.claude/stackgen/lock.yaml`, and `linter` has
  `.claude/hooks/npm-normalize.sh`.
- The next migrate item is M62 (M60 and M61 come from plan 1).
- Gates, the commit convention and the project `bootstrap` (`cli/`) are as in
  plan 1.

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

- The stackgen change in claude-plugins (folder 4): parked, see below.
- The `test` part of `flutter` (item 2 of the handoff prompt): a later plan.
- A BSD `sed` line in `~/.claude/CLAUDE.md`: not asked.
- The Mac sync of this plan on the Mac (`mise run updateall:configs`): a hand
  step after the user releases the engine of plan
  `2026-11-27-mac-backup-target`.

## Parked

- The stackgen change in claude-plugins (stop landing tool knowledge, the sync
  conversion, `ux-gate` moves to a pack, the materializer reads
  `references/conventions.md` from the consumer, the mise-skill line becomes a
  `CLAUDE.md` pointer):
  `docs/memory/gaps/2026-11-28-stackgen-tool-knowledge-moves-parked.md`.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-27-tool-skills

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
