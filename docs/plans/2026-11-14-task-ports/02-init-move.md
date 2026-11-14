# U2 — The init move and the dispatcher-path rule

- **Read first:** the `init` migration code that R1a leaves (the folder
  migration, `cli/internal/oldform/` and the `init` branch of
  `cli/internal/cli/`), the sync writer of `cli/internal/repo/` (`file.go`,
  how a task target is written, and the manifest read at `repo/repo.go`),
  and `check()` in `cli/internal/repo/repo.go` (with the stages that R1b's U2
  names).
- **Lazy-load:** `cli/internal/fail/fail.go`, `cli/internal/manifest/`,
  `docs/design/cli.md` ("The exit codes").

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D4 (the move):** `init` moves a stray `test/e2e` or `test/golden` file —
  one that the manifest does not list — to `<group>/50-local`
  (`test/e2e/50-local`, `test/golden/50-local`), then writes the dispatcher.
  `check` and `sync` report a file that the manifest does not list at a
  dispatcher path with code 2 and the fix line `→ Run bootstrap init`.
- **D7:** `init` also moves a repo's own `test` file or `test/_default` (not in
  the manifest) to `test/50-local` by the same rule. Its `depends` and its
  task names do not change.
- **D8:** if `<group>/50-local` already exists, `init` stops with code 2 and
  names both paths.
- **D12:** the same move applies to a repo's own `test/e2e/_default` and
  `test/golden/_default` (not in the manifest): six table paths in all.
- **D15:** `check` reports each executable file without an `NN-` name in a
  dispatcher folder (`code/format`, `code/lint`, `code/sec`, `setup/deps`,
  `setup/secrets`, `test`, `test/e2e`, `test/golden`) as a warning (code 0),
  with the fix "rename it to `NN-<name>`". `_default` and names that start
  with `_` are not parts and get no warning.
- **D18:** a file whose task name the `#MISE depends` of another file in the
  same folder names is not a part, and gets no D15 warning (for example this
  repo's `test/go`, which `test/50-local` names).
- **D20:** `init` examines every table entry first; if any target exists, it
  stops with code 2 before it moves a file (D8).
- **D25:** the move is one change in the change list (and in the TUI), with
  the label `move own test tasks into 50-local parts`; its detail lists each
  file and its target. It cannot be removed alone; to decline it, the user
  declines the `init` run. The same pattern as R1a's one folder-migration
  change.
- **D26:** a file at a table path that carries bootstrap's `Synced from
  bootstrap` header, but that the manifest does not list, is bootstrap's own:
  `init` and `sync` overwrite it with the new dispatcher and do not move it,
  and `check` gives no code 2 for it. Only a file without that header is the
  repo's own and moves.

## Edits

1. **One Go table of the dispatcher paths**, beside the other repo paths:
   `.config/mise/tasks/test` (a file), `.config/mise/tasks/test/_default`,
   `.config/mise/tasks/test/e2e` (a file),
   `.config/mise/tasks/test/e2e/_default`,
   `.config/mise/tasks/test/golden` (a file) and
   `.config/mise/tasks/test/golden/_default`, each with its `50-local` target
   (D12).
2. **The `init` migration** — before the sync, for each table entry that is a
   file the manifest does not list: examine every entry first, and stop with
   code 2 if any target exists (D8, D20); else move each to its target (D4,
   D7, D12). The move is part of the change list that `init` shows and
   applies, and of its restore on a failed sync.
3. **`check` and `sync`** — a file the manifest does not list at a table path
   gives code 2 and `→ Run bootstrap init`. In `check`, it is a finding of the
   stage that examines the values and the tree (R1b D8 names the stages).
   `check` also reports the D15 warnings, in the same stage.
4. **Tests** — a Go test for each case: a stray `test/e2e` file moved; a
   repo's own `test/_default` moved; a `50-local` that already exists (code 2);
   a manifest-listed `test/_default` left alone; `check` code 2; a repo's own
   `test/e2e/_default` moved (D12); a part `code/lint/mylint` gives a warning
   and code 0 (D15); a file `test/go` that `test/50-local` names in its
   `depends` gives no warning (D18); one target exists and another entry is
   due: code 2 and no file moved (D20); the move is one change with the D25
   label; a `test/_default` with the `Synced from bootstrap` header and no
   manifest entry is overwritten, not moved (D26).

## Verification

- `mise x -- mise run test:go` passes.

## Guardrails

- Touch nothing outside `cli/`.
- Never delete a moved file: a move only.
- Delete with `rm`, never `git rm`.

## Commit

`feat: init moves a repo's own test tasks into parts of the test ports`
