# U3 — Tests

- **Read first:** `.config/mise/tasks/test/repo` (the merge cases, about lines
  432-436, 2088, 2141-2150, 2806-2808 and 2834-2848, and any case that names
  `setup/worktree` or `code/worktrees`), `cli/internal/manifest/cases_test.go`
  (about line 167) and the golden files
  `cli/internal/manifest/testdata/read-repo*.golden`, and the example files
  under `repo/examples/` that name the old tasks.

## Ruling

D2, D3 and D13 to D16 are quoted in `01-git-tool.md`; D4 and D5 in
`02-mise-tool.md`. This unit tests the synced tasks end to end and updates the
test data that names the old tasks.

- **D16:** every case that runs `git:worktree:create` passes `--no-setup`.
- **D33:** every case that runs a `git:*` task exports
  `BOOTSTRAP_BIN=<the engine that test/repo builds>`, so the task's
  `config get` reaches this build, not a released engine.
- **D23:** the merge cases test only the entry tasks: put a stub
  `.config/mise/tasks/_scripts/merge` in the scratch repo that prints the
  destination and the source it gets. No case runs the real library (K2 does).
- **D20:** U7's `init` rename of `setup:worktree` in `.config/wt.toml` and a
  member's `worktree_install`.
- **D17:** in `repo/examples/`, edit the member's `worktree_install` value in
  the file that R1a U5 leaves: `repo.toml` under `[members.<name>]`.
- **D18:** a recorded exception to the unit contract: you may run the test's
  own `UPDATE_GOLDEN=1` for `cli/internal/manifest`; then review every changed
  golden file.

## Edits

1. **`.config/mise/tasks/test/repo`** — update every case that names
   `code/merge`, `code:merge`, `setup/worktree` or `code/worktrees` for the new
   names, and add:
   1. A scratch repo with `main` and `develop` as core branches syncs
      `git/merge/main`, `git/merge/develop` and
      `git/worktree/{create,setup,delete,list}`;
      no `code/merge*`, `setup/worktree` or `code/worktrees` file remains, and
      a repo synced before has them removed through the manifest (D1).
   2. A main-only scratch repo syncs only `git/merge/main`; with the D23 stub,
      `mise run git:merge:main feat` passes `main feat`. In the two-branch
      repo, `git:merge:main` with no argument passes `main develop`, and
      `git:merge:develop` passes `develop <current branch>`.
   3. `git:worktree:create x --no-setup` makes `.worktrees/x` on a new branch
      `x`; after a merge of `x` into `develop`, `git:worktree:delete x`
      removes the worktree and the branch (D22).
   4. `git:worktree:delete y` for a branch merged into no core branch exits 1,
      prints the `git branch -D y` command, and keeps the branch.
   5. A worktree that `git worktree add` made at another path is found and
      removed by `delete` (D28). `create` onto an existing path exits 1; `create
      develop` exits 2;
      `delete develop` exits 2; `delete` of a branch with no worktree exits 1
      (D13, D15, D25).
   6. A scratch repo whose `.config/wt.toml` runs `mise run setup:worktree`:
      `check` exits 2; `init` rewrites it to `mise run git:worktree:setup`
      (D20).
2. **`cli/internal/manifest/cases_test.go` and its golden files** — the new
   task paths where a case names a real task; regenerate the golden files
   with `UPDATE_GOLDEN=1` (D18) and review each one.
3. **`repo/examples/`** — only the lines that name the old tasks.

## Verification

- `mise x -- mise run test:repo` and `mise x -- mise run test:go` pass.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo`,
  `cli/internal/manifest/` and `repo/examples/`.
- No case needs the network: use `--clone` with the worktree.
- Delete with `rm`, never `git rm`.

## Commit

`test: the git:* tasks`
