# Plan — R1c: the test ports as group dispatchers (2026-11-14)

## Goal

After this plan lands, every shaped repo has the ports `test`, `test:e2e` and
`test:golden` as group dispatchers of the mise tool. A tool or the repo adds
`NN-<name>` parts, and a dispatcher with no part exits 0. A repo's own test
task keeps working as a `50-local` part, which `bootstrap init` moves into
place. The other plain ports of R1b D3 already exist (`code:format`,
`code:lint`, `code:sec`, `code:precommit`, `setup`) or are R2's
(`git:merge:<branch>`, `git:worktree:*`, `code:lsp:check`).

This is phase R1c of
`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`. It requires R1a
and R1b. It amends the record's rule that every port exists in every shaped
repo: a patterned port (`p:<id>:deploy`) has no default (D5).

## Facts the survey established

- Task targets are `.config/mise/tasks/<key>`, from the `[tasks."<key>"]`
  tables of a tool's `tools.toml` (`cli/internal/source/inventory.go`). A
  folder with `_default` and parts is a mise task group.
- `code:format`, `code:lint` and `code:sec` are dispatchers of the mise tool
  (`repo/tools/mise/tools.toml`, about lines 84-122) that loop `own_tasks`
  (`repo/tools/mise/tasks/helpers`, about 256-264). Every shipped part has an
  `NN-` name (for example `code/format/05-dprint`, `setup/deps/40-node`).
- `code:precommit` (pre-commit) and `setup` (mise) exist. The base pack
  (`repo/packs.toml`, about lines 7-17) requires mise, git, pre-commit,
  dprint, linter, gitleaks, grype and osv-scanner, so every shaped repo has
  these ports.
- No tool ships `test`, `test:e2e`, `test:golden`, `p:<id>:deploy`,
  `code:lsp:check`, `git:worktree:create` or `delete`. `git:merge` exists only
  as `code:merge*`, and `git:worktree:setup` as `setup:worktree`; R2 renames
  them.
- stackgen packs in claude-plugins copy task files into a repo:
  `test/golden` (two packs, for example `app-framework/compose`), `test/e2e`
  (`framework/android`) and `p/_project/deploy` (three packs, for example
  `cloud-service/containers`), from their `config/.config/mise/tasks/`.
- This repo has its own `.config/mise/tasks/test/_default` with
  `depends=["test:repo", "test:mac", "test:macos-tools", "test:go"]` and the
  parts `go`, `mac`, `macos-tools` and `repo`. It is not in the manifest.
- `repo/tools/mise/tasks/placeholder` is an unused library for no-op tasks.
- On mise 2026.11.7, a task that declares `--all` and one variadic argument
  gets an undeclared flag (`mise run t --update a`, with or without `--`) in
  the variadic argument and in `$*`; `--all` is parsed as the flag (tested in
  a scratch repo, 2026-11-14).
- R1a U4 edits files in `repo/tools/` (the merge, worktree and setup tasks);
  R1b edits `cli/` and adds the port table (D3) and `describe`'s `present`
  (D19, D27, D39). The TUI plan owns no task file.
- Docs of the task library: `docs/design/mise.md` (about 583-649),
  `docs/sync/tools/mise.md` (about 68-82), `docs/design/packs.md` (about
  190-203), `CLAUDE.md` (about 145-195), `docs/design/cli.md`,
  `docs/sync/migrate.md`.
- Gates: `code:format`, `code:lint`, `test:go`, `test`, `code:sec`. Commit
  types are the ten of `.config/git-conventional-commits.yaml`, with no
  scopes. No backlog item covers R1c.

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

- `git:merge:<branch>`, `git:worktree:*` and `code:lsp:check`: phase R2 (D1).
- `p:<id>:deploy` and per-project ids in bootstrap: phase R5 (D5).
- The unused placeholder library: no change.
- This repo's own `.config/` files: the user runs `bootstrap init` here after
  the 4.0.0 tag, which moves its `test/_default` (D7).
- A version bump or a tag (R1b D16).

## Parked

- The stackgen packs in claude-plugins must ship `test/e2e/NN-<pack>` and
  `test/golden/NN-<pack>` parts, not `test/e2e` and `test/golden` files (R3 or
  R5): `docs/memory/gaps/2026-11-14-pack-test-parts-parked.md`.
- Who ships the test parts of a stack after R5 (D23): bootstrap owns the
  dispatchers; the R5 interview decides the parts:
  `docs/memory/gaps/2026-11-14-r5-test-parts-owner-parked.md`.
- The claude-plugins copy of the design record needs D1 and D5 of this plan.
- R2 is planned next.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-14-task-ports

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
