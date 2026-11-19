# Plan — The shipped tasks of repo/tools in Python (2026-11-18)

## Goal

After this plan lands, each task that `repo/tools/` ships is a Python uv
script with a `.py` name, and the shell libraries are gone. A consumer gets
the Python tasks at its next sync; the stale rule removes the old shell
copies. The task names, flags, output and exit codes do not change. The own
tasks of this repository (`test:go`, `test/50-local`, `vorssaint:diff`) also
move to Python, because they source the shell `helpers` at the root, which
the self-sync after this plan removes (D13).

This is plan 2 of 3 of B30. It requires
`docs/plans/2026-11-18-python-tasks-foundation`, which adds the Python
libraries, the rules of a Python task and the gate fixes. Plan 3,
`docs/plans/2026-11-18-python-tasks-mac`, requires this plan. No reversal of
its own: the reversals are in the foundation plan and its record
`docs/memory/decisions/2026-11-18-python-tasks.md`.

## Facts the survey established

- **The task files** (68 in `repo/tools/*/tasks/`, about 6000 lines):
  - mise: `code` (zsh), `code-format.tmpl`, `code-lint.tmpl`,
    `code-lint-bootstrap`, `code-lsp-check`, `code-sec`, `code-worktrees`
    (873 lines, `git:worktree:list`), `setup-configs.tmpl`, `setup-mise`,
    `setup-worktree.tmpl`, `setup.tmpl`, `test`, `test-e2e`, `test-golden`;
    libraries `helpers`, `checks`, `placeholder`, `worktrees-page`.
  - git: `code-git-config`, `git-merge-develop`, `git-merge-main`,
    `git-worktree-create`, `git-worktree-delete`; library `merge`.
  - android: `code-lint-android`, `setup-deps-android` (245 lines),
    `test-e2e-android`, `test-golden-android`. kotlin: `code-format-ktlint`,
    `code-lint-detekt`, `setup-deps-kotlin`. flutter: `code-format-dart`,
    `code-lint-dart`, `setup-deps-flutter`.
  - swift: `code-format-swift-format`, `code-lint-swiftlint`,
    `setup-deps-swift`, `test-golden-swift`; library `xcode`. go:
    `code-lint-go`, `code-lint-gomod`, `setup-deps-go`. node:
    `setup-deps-node`. python: `code-lint-pyrefly`, `code-lint-ruff`,
    `code-lint-uv`, `setup-deps-python`. docker:
    `setup-external-{check,pull,start,stop,update}`.
  - claude: `setup-ai`. vscode: `setup-vscode`. dprint:
    `code-format-dprint`, `setup-dprint`. linter: `code-lint-linter`,
    `code-lint-shellcheck`, `setup-linter`. pre-commit: `code-precommit`,
    `setup-precommit`. gitleaks, grype, osv-scanner: one `code-sec-*` each.
    doppler, fnox: one `setup-secrets-*` each.
- **The `#MISE tools.` lines:** `setup-dprint` (dprint), `git-merge-main` and
  `git-merge-develop` (pre-commit, yq), `git-worktree-create` and
  `git-worktree-delete` (yq), `setup.tmpl` (yq), `setup-configs.tmpl`
  (`github:virajp/bootstrap`, yq), `code-worktrees` (yq), `setup-precommit`
  and `code-precommit` (pre-commit). Each yq call parses the JSON of
  `bootstrap config get`.
- **The templates** render with the delimiters `<%` and `%>` (README.md
  229-233); `{{ }}` passes through. They have `mode = "template"` in
  `repo/tools/mise/tools.toml`
  (lines 104-134) and render with gomplate. The engine renders a source by
  its `mode`, not by its name, and `cli/internal/format/format.go:39` reads
  the format of a target from its extension without `.tmpl`.
- **The Go cases that run or name a task:** `e2e/repo/commands/tasks_test.go`
  (`code:lsp:check`, `setup:ai`, `setup --upgrade`, `setup:configs --all`,
  the landing and entry cases of `git:merge:<branch>` with a stub `merge`
  library, the worktree tasks, `code_files`),
  `e2e/repo/commands/helpers_test.go`,
  `e2e/repo/tasks/{configs,gate,clone,main}_test.go`,
  `e2e/repo/{config,detect,files,migrate,sync}/*_test.go` (task paths in the
  expected file lists), `e2e/macostools/macostools_test.go`, and the
  `e2e/repo/pylib/` package of the foundation plan.
  `e2e/internal/repocase/repocase.go:525` maps a task target to its source
  name generically; do not edit `e2e/internal/`.
- **The docs** of the shell rules and of each tool: CLAUDE.md "Task scripts";
  `docs/design/{mise,tools,repo,cli}.md`; `docs/sync/index.md` 126-155, 701
  and "Files that bootstrap no longer ships" (316-356);
  `docs/sync/migrate.md` (last item M45); `docs/sync/tools/*.md` for each
  tool of this plan; `README.md` 229-233.

- **`mise run` and a task flag.** `mise run <task> --all` gives the task
  `usage_all=true`, the same as `mise run <task> -- --all` (tested
  2026-11-18). The gate lines need no `--`.

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
Its edits stay uncommitted in the worktree.

## Out of scope

- The engine scripts of `repo/tools/*/scripts/` stay bash (foundation D2).
- A change to the code of `cli/` (foundation D14), except the one fix of
  `UnnamedParts` (D15, U14).
- The Mac tasks: plan 3.
- A change of behavior of a task while it moves (D2).

## Parked

- B30: the Mac tasks and the Mac helpers in Python —
  `docs/plans/2026-11-18-python-tasks-mac`.
- The other consumer repositories run `mise run setup:configs` by hand after
  plan 3.

- The `code:check` type-check group for each language (backlog item B31, P0):
  a change plan of its own that requires
  `docs/plans/2026-11-18-python-tasks-repo`.
  pyrefly stays `code:lint:50-pyrefly` in B30.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-18-python-tasks-repo

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
