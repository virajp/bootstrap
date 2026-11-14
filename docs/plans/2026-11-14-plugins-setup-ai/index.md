# Plan — R2c: contributed merge items, the virajp-plugins tool and a declarative setup:ai (2026-11-14)

## Goal

After this plan lands, `setup:ai` is declarative and names no plugin: it
installs what the claude tool's values list, removes the project plugins it
installed when they leave the list, and updates only with `--update`. A new
`virajp-plugins` tool carries vwf's `.gitignore` lines and its plugin entry,
through a general engine mechanism for contributed merge items. It is the last
of the three R2 plans of
`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`, and it finishes
backlog item B28.

## Facts the survey established

- `repo/tools/claude/tasks/setup-ai` today: exits in a submodule; warns and
  exits 0 with no `claude`; `claude_step` collects failures; `plugins_here`
  reads `claude plugin list --json`; a hardcoded block runs
  `pnpx @virajp.dev/claude-plugins@latest --all` when `vwf@virajp-plugins` is
  absent (about lines 64-79); declared plugins `"<plugin>@<marketplace>" =
  "<source>"` per scope, each installed at its scope with its marketplace
  added when missing (about 81-101); always updates every plugin (about
  103-110); `claude plugin autoremove --scope project` (about 112); never
  uninstalls a plugin that left the list. R1a U4 changes only its read to
  `bootstrap config get` and adds `plugins.user`/`plugins.project` defaults
  (R1a D86).
- The git tool owns `.gitignore` as a merge-mode file
  (`repo/tools/git/tools.toml`, about 61-64); vwf's lines are in
  `gitignore.merge.toml` (about 182-186 and 611-619). An override table can
  only change or remove existing items (`cli/internal/format/shared.go`, about
  306-347); `[contributes.<tool>]` merges values only
  (`cli/internal/values/merge.go`, about 93-125, 289-318). No tool can add
  lines to another tool's merge file today.
- `virajp-plugins` exists only as a name inside `setup-ai`.
- claude-plugins: stackgen's handoff sends plugin instructions to bootstrap's
  `docs/sync/vwf.md`, which says removal is a hand step. `/vwf:init` names no
  plugin.
- Backlog B28 (In progress, Planned in `docs/plans/2026-11-14-lsp`): install a
  Claude marketplace and plugins, identify LSP servers. R2b lands the LSP piece.
- Gates: `code:format`, `code:lint`, `test:go`, `test`, `code:sec`. Commit
  types: the ten of `.config/git-conventional-commits.yaml`, no scopes.

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

- The Mac side: plugin installs are per repo (D8).
- claude-plugins (`/vwf:init`, the stackgen handoff, their docs): parked below.
- This repo's own `.config/` files, except `.config/mise/tasks/test/`, which
  holds bootstrap's own tests. After the plan lands, re-sync this repo and add
  `virajp-plugins`: its `.config/mise/tasks/setup/ai` and root `.gitignore`
  keep the pnpx block and the vwf lines until then (D30).
- A version bump or a tag (R1b D16).

## Parked

- claude-plugins R3: `/vwf:init` and the stackgen handoff add the
  `virajp-plugins` tool (`bootstrap add tool virajp-plugins`) and stop naming
  the pnpx install and the hand removal of plugins:
  `docs/memory/gaps/2026-11-14-plugins-add-virajp-plugins-parked.md`.
- The vwf mentions outside the files of R2c (D21):
  `repo/tools/mise/tasks/placeholder:64`,
  `macos/tools/claude-cli/files/.claude/settings.json.merge.toml:41` and
  `cli/internal/oldform/oldform_test.go:63`.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-14-plugins-setup-ai

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
