# U7 — init renames setup:worktree in wt.toml and worktree_install

- **Read first:** the `init` migration code that R1a leaves (the folder
  migration and its change list, `cli/internal/oldform/` and the `init`
  branch of `cli/internal/cli/`), the reader of `[members.<name>]` in
  `.config/bootstrap/repo.toml` (R1a U3), and `check()` in
  `cli/internal/repo/repo.go` with the stages that R1b's U2 names.
- **Lazy-load:** `repo/tools/git/tools.toml` (`.config/wt.toml` is a seed
  file, about lines 54-56), `repo/tools/git/files/.config/wt.toml`,
  `cli/internal/fail/fail.go`, `docs/design/cli.md` ("The exit codes").

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1:** no aliases for the old task names.
- **D20:** `init` rewrites `mise run setup:worktree` to
  `mise run git:worktree:setup` in `.config/wt.toml` (a seed file, which the
  sync never rewrites) and in each `[members.<name>].worktree_install` value of
  `.config/bootstrap/repo.toml`. It is one change in the change list (and in
  the TUI), with the label `rename setup:worktree`, whose detail lists each file
  and key. `check` and `sync` report an old name in either place with code 2
  and the fix line `→ Run bootstrap init`. Only that exact command text
  changes; the rest of the file or value stays as it is.

## Edits

1. **The `init` migration** — find the old command text in `.config/wt.toml`
   and in each member's `worktree_install`; add the D20 change; apply it with
   the rest of the change list, inside `init`'s restore on a failed sync. The
   `repo.toml` edit goes through the same writer that `config set` uses
   (comments kept).
2. **`check` and `sync`** — the D20 code 2, as a finding of the stage that
   examines the values and the tree.
3. **Tests** — a Go test for each case: `wt.toml` with the old command is
   rewritten; a member's `worktree_install` with the old command is rewritten;
   a file with no old command is not changed; `check` exits 2 before `init`
   and 0 after.

## Verification

- `mise x -- mise run test:go` passes.

## Guardrails

- Touch nothing outside `cli/`. `cli/internal/manifest/` test data is U3's
  (wave 2); do not edit it.
- Delete with `rm`, never `git rm`.

## Commit

`feat: init renames setup:worktree in wt.toml and member worktree_install`
