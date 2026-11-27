# Plan — Drop the bun tool, and android depends on kotlin (2026-11-27)

## Goal

After this plan lands, bootstrap has no `bun` tool, and each selection of
`android` (and so of `flutter`) also selects `kotlin`.

This is plan 1 of the skill-delivery work of 2026-11-27. The user ruled both
changes while the delivery of tool skills was designed: no project uses
`bun`, and Android is Kotlin-first. Plan 3, `2026-11-27-tool-skills`, requires
this plan, so that no skill or contribution names `bun`, and the `kotlin`
skill reaches Android repositories. No recorded decision is reversed.

## Facts the survey established

- `bun` appears in: `repo/tools/bun/` (tools.toml, files/bunfig.toml,
  scripts/detect), `repo/tools/node/tasks/setup-deps-node.py:33-35` (a skip
  when only `bun` is selected), its synced copy
  `.config/mise/tasks/setup/deps/40-node.py:33-34`,
  `cli/internal/source/testdata/toml-yq.golden:186-187`,
  `e2e/repo/detect/detect_test.go:123`, `docs/design/mise.md:538`,
  `docs/sync/index.md:58`, `docs/sync/migrate.md:684` (M30),
  `docs/sync/tools/node.md:79` and `docs/sync/tools/bun.md`. No hit in
  `repo/packs.toml`, `repo/categories.toml`, `repo/examples/`, `macos/` or the
  Go sources of `cli/`.
- Today `android` depends on `java`, `kotlin` depends on `java`, and
  `flutter` depends on `android`.
- Tests that state the closure: `e2e/repo/detect/selection_test.go:55-59`,
  `:109-114`, `:116-118`; `e2e/repo/files/packs_test.go:23`, `:145-146`;
  `e2e/repo/detect/detect_test.go:95-98`;
  `e2e/repo/migrate/migrate_test.go:443`,
  `:461-536`.
- Docs that state the closure: `docs/sync/tools/android.md:8`,
  `docs/sync/tools/flutter.md:8`, `:35-37`, `docs/sync/tools/kotlin.md:140-146`
  (the CI text), `README.md:128-136` (a sample android tools.toml),
  `docs/design/mise.md:524`, `:806-807`, `docs/design/tools.md:584-585`,
  `docs/design/cli.md:704-716` (a detect sample with "kotlin not offered").
- `toml-yq.golden` holds every TOML file of `repo/`. `UPDATE_GOLDEN=1 go test
  ./internal/source` in `cli/` writes it again.
- The last migrate item is M59 (`docs/sync/migrate.md:1803`). Items use
  `## M<n>: <title>`.
- Gates: `code:format` (dprint), `code:lint` (linter, shellcheck, go, gomod,
  bootstrap), `code:precommit`, `test` (test:go, then gotestsum in `cli/` and
  the root module).
- Commit convention: `.config/git-conventional-commits.yaml`, types feat, fix,
  perf, refactor, revert, test, ops, docs, merge, wip; no scopes.
- The project `bootstrap` is the folder `cli/`.

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

- The tool skills, the `CLAUDE.md` section and the Mac changes: plan 3,
  `2026-11-27-tool-skills`.
- The `backup:` target of the Mac sync: plan 2, `2026-11-27-mac-backup-target`.
- A split of `android` into an SDK tool and an app tool: the user chose that
  `flutter` gets `kotlin`.

## Parked

none

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-27-tool-selection

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
