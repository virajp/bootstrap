# Plan — The catalog data: categories, texts, selectable, hosts and the Mac git emails (2026-11-12)

## Goal

After this plan lands, each tool of both sides has a `summary` and a
`category`, and the engine stops the sync for a tool without them. Each repo
tool has a `description`. Each side has a
categories file. `_security_base` is not selectable, `swift` runs only on
macOS, and the Mac git tool needs one email only.

This is plan 1b of three chained plans for the TUI `bootstrap tui` and
`bootstrap mac tui` (artifact "Bootstrap TUI Screens",
https://claude.ai/artifact/QGqhhDrKesC3N2Kz7x1nr5). It requires plan 1a,
`docs/plans/2026-11-12-catalog-engine`, which adds the keys and their rules to
the engine. It lands after the user tags `bootstrap@2.1.0` (D7). Plan 2 (the
TUI) requires this plan. The plan reverses no standing decision.

## Facts the survey established

- The tools: 36 folders in `repo/tools/` and 109 folders in `macos/tools/` on
  2026-11-12 (after `c0bb6f8`, one `fonts` tool, and `5bb99b2`, `gem` replaced
  by `ruby`). `texts.md` covers each one.
- The packs: `repo/packs.toml` has `base`, `security` and `container-build`.
  `macos/packs.toml` has `base`, `ai`, `essentials`, `security`, `dev`,
  `terminal`, `apple-dev`, `iwork`, `ms-office`, `disk`, `media`, `chat`,
  `notes` and `cli-extras`. No pack has a `description` now. The local branch
  `packs-shortcut` (`420e3e6`) reshapes both files for plan 2; this plan does
  not change them (D13).
- Each Mac tool has a `description` now. No repo tool has one: its text is
  the header comment of its `tools.toml`.
- The Mac git questions are at `macos/tools/git/tools.toml:12-29`. The four
  templates that use `required` on an email are in
  `macos/tools/git/files/.config/git/`: `identity.tmpl:9`,
  `allowed_signers.tmpl:3-4`, `github.config.tmpl:6` and
  `gitlab.config.tmpl:6`. They use the delimiters `<%` and `%>`.
  `macos/examples/macos.toml:72-75` answers both emails.
- No values file of this repository or of `repo/examples/` names
  `_security_base` or `swift` directly.
- The engine of plan 1a validates `summary`, `category`, `selectable`, `hosts`,
  `description`, the categories file and the question keys in
  `cli/internal/source/check.go` (`repoRules` and `macosRules`) and in
  `cli/internal/values/merge.go:224`. Search for `summary` and `category`
  there. The keys are optional after plan 1a.
- The test tasks write their own `tools.toml` files: about 85 places in
  `.config/mise/tasks/test/repo` and about 29 in `.config/mise/tasks/test/mac`.
  `macos/fixtures/` holds 18 `tools.toml` files. The Go tests write
  `tools.toml` text in `cli/internal/source/source_test.go`,
  `cli/internal/cli/packages_test.go`,
  `cli/internal/selection/form_test.go`, `cli/internal/selection/cases_test.go`,
  `cli/internal/misegen/golden_test.go` and
  `cli/internal/misegen/cases_test.go`.
  `cli/internal/source/testdata/toml-yq.golden`
  hashes each `.toml` below `repo/` and `macos/`, the fixtures included
  (`cli/internal/source/toml_test.go:17-49`, `:36`).
- `test:macos-tools` (`.config/mise/tasks/test/macos-tools`) syncs each tool
  of `macos/tools/` with `macos/examples/macos.toml`.
- `--clone <path>` makes a command use that folder as the clone, as it is
  (`docs/design/cli.md:422`, `cli/internal/cli/cli.go:83`).
- The last item of `docs/sync/migrate.md` is M36.
- The commit types are in `.config/git-conventional-commits.yaml`: `feat`,
  `fix`, `perf`, `refactor`, `revert`, `test`, `ops`, `docs`, `merge` and
  `wip`. The repository has no scopes.

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

- The TUI, and every engine change other than D5: plan 2, or plan 1a.
- The packs files: no description and no new comment (D13). Plan 2 writes
  them with the packs of the branch `packs-shortcut`.
- A change of the Mac tool descriptions of today, and of the header comment of
  a repo tool.
- A `conflicts` key: `parent` is the only conflict data.

## Parked

- Plan 2, the TUI: `bootstrap tui` and `bootstrap mac tui` per the artifact.
  It requires this plan.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-11-12-catalog-data

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
