# U1 — The git tool: git:merge per core branch and the worktree tasks

- **Read first:** `repo/tools/git/tools.toml` (the `[tasks.*]` entries, about
  lines 68-84, and the `when` keys), `repo/tools/git/tasks/merge` (the
  library: `merge_to_destination_branch`, `open_pull_request`, the landing read
  that R1a U4 rewrote), `repo/tools/git/tasks/code-merge-develop`,
  `repo/tools/git/tasks/code-merge-main`, `repo/tools/git/scripts/derive` (the
  JSON it prints, which the `when` keys read),
  `repo/tools/git/files/.config/wt.toml`.
- **Lazy-load:** `repo/tools/mise/tasks/code-merge` (the one-branch entry task
  that U2 removes; its destination pick), `repo/tools/mise/tasks/helpers`,
  `cli/internal/source/inventory.go` (how `when` is read),
  `/Users/virajpatel/Projects/github.com/virajp/claude-plugins/plugins/vwf/skills/git-workflow/scripts/lib/land.mjs`
  (lines about 99-133 and 372-390: the worktree add and remove rules to match).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1:** no aliases. The old task files go away; the sync removes them from a
  repo through the manifest.
- **D2:** the git tool ships `git/merge/main` and `git/merge/develop`
  (tasks `git:merge:main`, `git:merge:develop`). Each has a `when` that is true
  only when that branch is in `repo.core_branches`. `<branch>` in the task name
  is always the destination. Each takes an optional source argument: for
  `git:merge:main` the default is `develop` when `develop` is a core branch,
  else the current branch; for `git:merge:develop` the default is the current
  branch. Both call `merge_to_destination_branch <dest> <src>`; the library and
  its landing (`pr | direct-merge | direct-commit`) do not change.
- **D3:** `git:worktree:create <branch> [--from <base>]` and
  `git:worktree:delete <branch>`, in raw git, with the rules of `land.mjs`:
  - create: the folder is `.worktrees/<branch>`; a `worktrees/` folder is used
    only when `.worktrees/` is absent and `worktrees/` exists. A branch that
    does not exist gets `-b`; `--from` sets its base, default the current
    branch. A branch that exists is checked out with no `-b` (and `--from` is
    then a usage error, code 2). After the add, run `mise run
    git:worktree:setup` inside the new worktree.
  - delete: `git worktree remove <path>` (git refuses a dirty tree; pass that
    error on with code 1), then the D22 branch rule below.
- **D21 (create):** run `mise trust` in the new worktree before
  `git:worktree:setup`.
- **D22 (delete):** for each branch of `repo.core_branches`, test
  `git merge-base --is-ancestor <branch> <core>`. If one passes, delete the
  branch with `git branch -D` (safe after that proof). Else exit 1 and print
  the `git branch -D <branch>` command; do not run it.
- **D24:** `git:merge:main` has no "current branch must be develop" rule; the
  library's `merge_from` default decides the allowed source.
- **D25 (create):** a branch in `repo.core_branches`: exit 2. A branch that
  another worktree has checked out: git refuses; exit 1 with git's message.
- **D26:** `create` and `delete` read `repo.core_branches` as D19 says.
- **D27 (delete):** in the D22 test, use each core branch's local ref when it
  exists, else `origin/<core>` when it exists, else skip that core branch.
- **D28 (delete):** find the worktree by its branch in
  `git worktree list --porcelain`, at any path, and remove that path.
- **D29 (create):** `--no-setup` skips only `git:worktree:setup`;
  `mise trust` always runs.
- **D5:** `wt.toml`'s `pre-start` step runs `mise run git:worktree:setup`.
  `code:git-config` keeps its name.
- **D10:** this unit also owns `repo/tools/github/tools.toml` and
  `repo/tools/gitlab/tools.toml`, for the comments that name `code:merge`.
- **D11:** keep `branch_pair` in `derive`, and add `has_main` and
  `has_develop`. No `cli/` change.
- **D12:** no new rule for a source equal to the destination: the library
  already refuses it, and the entry task passes its code on.
- **D19:** `git:merge:main` reads `repo.core_branches` through
  `${BOOTSTRAP_BIN:-mise x -- bootstrap} config get repo.core_branches --format
  json`
  (`.value`), with the D85 exit-2 rule of R1a U4 when no engine is installed.
- **D13 (create):** `.worktrees/` is relative to the main checkout. An existing
  path exits 1. A branch that exists only on `origin` is checked out as a local
  branch that tracks `origin/<branch>`; `--from` is then a usage error (code 2).
- **D14 (create):** if `git:worktree:setup` fails, keep the worktree, exit 1,
  and print the command that runs the setup again and the `delete` command.
- **D16 (create):** `--no-setup` skips the setup step.
- **D15 (delete):** no worktree for the branch: exit 1. Run from inside the
  worktree it removes: exit 2, "run it from another worktree". A branch in
  `repo.core_branches`: exit 2; it never deletes main or develop.

## Edits

1. **`repo/tools/git/scripts/derive`** — print one flag per core branch the
   `when` keys can read (for example `has_main`, `has_develop`), from
   `repo.core_branches` as R1a U4 reads it. Keep `branch_pair` (D11).
2. **`repo/tools/git/tasks/git-merge-main`, `git-merge-develop`** (new) — the
   D2 entry tasks, with `#USAGE` for the optional source. Remove
   `code-merge-main` and `code-merge-develop` with `rm`.
3. **`repo/tools/git/tasks/git-worktree-create`, `git-worktree-delete`** (new) —
   D3, each sourcing the helpers, with `#USAGE` for the branch and `--from`.
4. **`repo/tools/git/tools.toml`** — the `[tasks."git/merge/main"]`,
   `[tasks."git/merge/develop"]` (with their `when`),
   `[tasks."git/worktree/create"]` and `[tasks."git/worktree/delete"]`
   entries; remove the `code/merge/*` entries. Update comments that name the
   old tasks.
5. **`repo/tools/git/files/.config/wt.toml`** — `pre-start` runs
   `mise run git:worktree:setup` (D5).
6. **`repo/tools/git/tasks/merge`** — only the comments and messages that name
   `code:merge*` (for example the safety-net comment); no logic change.
7. **`repo/tools/github/tools.toml`, `repo/tools/gitlab/tools.toml`** — the
   comments that name `code:merge` (D10).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rn 'code:merge\|code/merge\|setup:worktree' repo/tools/git` finds
  nothing.

## Guardrails

- Touch nothing outside `repo/tools/git/`, `repo/tools/github/tools.toml` and
  `repo/tools/gitlab/tools.toml`. `repo/tools/mise/` is U2's.
- Do not change the landing logic of the library.
- Delete with `rm`, never `git rm`.

## Commit

`feat: git:merge per core branch and git:worktree create and delete`
