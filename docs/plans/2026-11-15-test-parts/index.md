# Plan — R2e: test parts of the android and swift tools (2026-11-15)

## Goal

After this plan lands, the `android` tool ships `test/e2e/40-android` and
`test/golden/40-android`, and the `swift` tool ships `test/golden/40-swift`,
as parts under the test dispatchers of R1c. `bootstrap init` removes the old
copies of these tasks that the stackgen packs wrote. claude-plugins R5a then
deletes the files from its packs (R5a D8). It closes the two parked
test-parts gaps of `docs/memory/gaps/`.

## Facts the survey established

- A tool ships a task with `[tasks."<path below .config/mise/tasks/>"]` and
  `source = "<file below tasks/>"` in its `tools.toml`
  (`repo/tools/android/tools.toml`: `code/lint/10-android`,
  `setup/deps/10-android`). A task that does not apply prints nothing and
  exits 0 (`repo/tools/android/tasks/code-lint-android`).
- The `swift` tool ships `_scripts/xcode` (`library = true`), the Xcode and
  simulator lookup, read from `SIMULATOR_PLATFORM`, `SIMULATOR_DEVICE` and
  `SIMULATOR_OS`.
- R1c (approved, not run): the mise tool ships `test/_default`,
  `test/e2e/_default` and `test/golden/_default`; each runs the executable
  `NN-<name>` parts and passes the raw arguments to each part (D2, D19, D24).
  `init` moves a repo's own `test/e2e` or `test/golden` file to
  `<group>/50-local` (D4, D12); `check` reports such a file with code 2.
- The three stackgen pack tasks, in claude-plugins at commit `0562cb44`:
  - `plugins/stackgen/stacks/framework/android/config/.config/mise/tasks/test/e2e`
    (flags `--device`, `--variant`, `--connected`; Gradle managed devices);
  - `plugins/stackgen/stacks/app-framework/compose/config/.config/mise/tasks/test/golden`
    (`--record`, `--variant`; `record`/`verifyRoborazzi<Variant>`);
  - `plugins/stackgen/stacks/app-framework/swiftui/config/.config/mise/tasks/test/golden`
    (`--record`, `--target`, `--scheme`, `--platform`, `--device`, `--os`;
    sources `_scripts/xcode`).
- `[questions]` values reach templates at sync time only; a copied task file
  cannot read them (`docs/design/tools.md`, "The defaults, questions and
  contributes tables").
- Gates: `code:format`, `code:lint`, `test:go`, `test`, `code:sec`. Commit
  types: the ten of `.config/git-conventional-commits.yaml`, no scopes.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc it does not own, never adds a
dependency the plan does not list, never commits. A unit deletes with plain
`rm`, never `git rm` — it stages nothing. A unit never writes a lock file with
Write, Edit, `sed` or a script.

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

- claude-plugins: R5a deletes the pack files and drops their lock rows.
- A Flutter app's own `android/` folder: the Flutter tool owns its tests.
- `p:<id>:deploy` and `p/_project/icons`: they stay in the stackgen packs.
- This repo's own `.config/` files, except `.config/mise/tasks/test/`.
- A version bump or a tag (R1b D16).

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-15-test-parts

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
