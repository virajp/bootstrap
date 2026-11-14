# Plan — R2d: projects in bootstrap (2026-11-15)

## Goal

After this plan lands, each repo lists its projects in `repo.toml`
(`[projects.<name>]` with `folder`), each tool in `tools.toml` has a
`projects` list of those names, `bootstrap detect` proposes both, and
`bootstrap check` verifies them. `describe` gives each tool its category. vwf
format 28 (claude-plugins R4b) then reads a project's languages and package
manager through the bootstrap CLI. This plan makes the ST2 line of
`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md` true.

## Facts the survey established

- No tool has a `projects` list today, and no plan adds one. The CLI has no
  `projects` field.
- A tool's `tools.toml` in the clone has `category`; `languages` (10 tools)
  and `package-managers` (pnpm, aube) are two of the categories.
- `bootstrap detect` runs each detect script once, in the repo root
  (`cli/internal/scripts/scripts.go`, `cmd.Dir = r.Root`, about line 159;
  `cli/internal/core/detect.go`). The node script tests `package.json` in its
  working folder.
- `cli/internal/repo/rules.go` (about line 405) already lists files with
  `git ls-files --cached --others --exclude-standard`.
- R1a, R1b, R1c and R2a to R2c are APPROVED and not yet run. This plan builds
  on the code they add: the `internal/schema` table, `config`, `describe`,
  `detect --json`, `check`.
- vwf keeps each project's role, platforms and other facts in its own
  `registry.yaml`, keyed by name (D3).
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

- The Mac side: projects are per repo.
- The TUI and `init`: they write no projects (D9).
- claude-plugins: R4b reads the projects after the bootstrap 4.0.0 release.
- This repo's own `.config/` files, except `.config/mise/tasks/test/`, which
  holds bootstrap's own tests.
- A version bump or a tag (R1b D16).

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-15-projects

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
