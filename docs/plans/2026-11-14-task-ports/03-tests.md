# U3 — Tests

- **Read first:** `.config/mise/tasks/test/repo` (its case helpers, its
  scratch repos, `check_base_files` at about lines 348-372, and the `when`
  cases at about 2837-2847), and the work of U1 and U2.

## Ruling

D2, D3 and D4 are quoted in `01-test-dispatchers.md`; D4, D7 and D8 in
`02-init-move.md`. This unit tests the built engine and the synced tasks end to
end, in the style of the other cases of `test/repo`.

## Edits

1. **`.config/mise/tasks/test/repo`** — add cases:
   1. A fresh `init` of a scratch repo writes `test/_default`,
      `test/e2e/_default` and `test/golden/_default`; `mise run test`,
      `mise run test:e2e` and `mise run test:golden` each exit 0 and print the
      no-part line.
   2. A part `test/10-a` runs under `mise run test`; a file `test/b` with no
      `NN-` prefix does not run.
   3. A part `test/e2e/10-a` does not run under `mise run test` (D3).
   4. An unshaped scratch repo (a git repo with no `.config/bootstrap/`, D21)
      with its own `test/_default` (with a `depends` on a file `test/go`) and
      a stray `test/e2e` file: `check` exits 2 with `→ Run bootstrap init`
      (R1a's rule for a missing folder); `init` moves them to `test/50-local`
      and `test/e2e/50-local`; then `mise run test` runs the old dispatcher as
      a part, and `check` exits 0 with no warning for `test/go` (D18).
   5. A scratch repo with `test/50-local`, its own `test/_default` and a stray
      `test/e2e`: `init` exits 2, names both paths, and moves no file (D8,
      D20).
   6. A scratch repo with its own `test/golden/_default`: `init` moves it to
      `test/golden/50-local` (D12). A file `code/lint/mylint`: `check` exits
      0 and prints the rename warning (D15).
   7. A scratch repo whose `test/_default` carries the `Synced from bootstrap`
      header but no manifest entry: `init` overwrites it with the dispatcher,
      and no `test/50-local` appears (D26).
   8. Update `check_base_files` and any case that lists the synced task files
      of the mise tool, for the three new dispatchers.

## Verification

- `mise x -- mise run test:repo` passes.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo`.
- No case needs the network: use `--clone` with the worktree.
- Delete with `rm`, never `git rm`.

## Commit

`test: bash cases for the test ports and the init move`
