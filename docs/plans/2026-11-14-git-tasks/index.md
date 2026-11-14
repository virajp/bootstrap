# Plan — R2a: the git:* tasks (2026-11-14)

## Goal

After this plan lands, the git actions of a shaped repo have their `git:*`
names: `git:merge:<branch>` for each core branch (the destination in the
name), `git:worktree:create`, `git:worktree:setup`, `git:worktree:delete` and
`git:worktree:list`. `code:merge`, `code:merge:<b>`, `setup:worktree` and
`code:worktrees` are gone, with no aliases (D1).

This is the first of the three R2 plans of
`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md` (D9): R2a git
tasks, R2b LSP, R2c plugins. It requires R1c. D1 reverses the design text that
kept the old names as aliases for one release.

## Facts the survey established

- One library does every merge: `repo/tools/git/tasks/merge`
  (`merge_to_destination_branch <dest> [<src>]`; `open_pull_request` with `gh`
  or `glab`; `--no-ff`; a safety net that runs pre-commit; a hop to the main
  worktree). R1a U4 changes its landing read to `repo.branches.<b>.landing`
  (`pr | direct-merge | direct-commit`) and renames no task.
- Three entry tasks call it: `code:merge:develop <src>` and `code:merge:main`
  (git tool, `when = "git.branch_pair"`), and `code:merge <src>` (mise tool,
  `when = "!git.branch_pair"`, picks main or develop). `branch_pair` comes from
  `repo/tools/git/scripts/derive`; R1a U4 derives it from `repo.core_branches`.
- No task creates or deletes a worktree. `setup:worktree`
  (`repo/tools/mise/tasks/setup-worktree.tmpl`) prepares an existing one;
  `code:worktrees` (881 lines) lists worktrees, alias `worktrees`
  (`repo/tools/mise/tools.toml`, about line 29). `wt.toml` (git tool) runs
  `mise run setup:worktree` in `pre-start`.
- claude-plugins' git-workflow calls no bootstrap task in code: `land.mjs` runs
  raw `git worktree add/remove`, `git merge --no-ff`, `gh pr create` and
  `glab`. Its docs (`SKILL.md`, `references/landing.md`) name
  `code:merge:<b>`; R3 moves them.
- Tests and data that name the old tasks: `.config/mise/tasks/test/repo` (about
  432-436, 2088, 2141-2150, 2806-2808, 2834-2848),
  `cli/internal/manifest/cases_test.go:167`, 12 golden files
  `cli/internal/manifest/testdata/read-repo*.golden`, and
  `repo/examples/repo-submodules/bootstrap.toml`.
- Docs that name them: `README.md`, `CLAUDE.md`,
  `docs/design/{repo,mise,tools,packs,cli,bootstrap-config}.md`,
  `docs/sync/index.md`, `docs/sync/tools/{git,github,gitlab,mise}.md`.
- R1b D3 names the ports `git:merge:<branch>` and
  `git:worktree:create|setup|delete`; R1b's `describe` reports them per D19.
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

- R2b (LSP blocks, `bootstrap lsp`, `code:lsp:check`) and R2c (the
  `virajp-plugins` tool, a declarative `setup:ai`); backlog item B28 (D9).
- claude-plugins' git-workflow and `land.mjs`: R3 moves them onto the `git:*`
  tasks.
- The landing logic of the merge library: R1a U4 owns it.
- This repo's own `.config/` files and the synced copies in this checkout,
  except `.config/mise/tasks/test/`, which holds bootstrap's own tests and
  which U3 owns (D32).
- `docs/memory/handoff/bootstrap-plugins-split.md`, a session record (D31).
- A version bump or a tag (R1b D16).

## Parked

- R2b and R2c are planned next in this session.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-14-git-tasks

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
