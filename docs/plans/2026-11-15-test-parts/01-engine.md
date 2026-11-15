# U1 — The engine: the replaces key, init removal, check

- **Read first:** the `tools.toml` reader of `cli/` (the `[tasks.*]` table and
  its keys), and the code that R1c added: the `init` move of a repo's own
  test file to `<group>/50-local`, and its `check` and `sync` reports.
- **Lazy-load:** D4, D7, D12, D15 and D25 of
  `docs/plans/2026-11-14-task-ports/index.json` (or its archived path).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D4:** a `[tasks."<path>"]` entry may have `replaces`, a list of strings
  `sha256:<64 hex>`. Another form of an element stops the sync, as an unknown
  key does today.
- **D4, D5:** for each part with `replaces`, the files it replaces are found at
  the old path of its group (`test/e2e` or `test/golden`, the group being the
  part's folder) and at `<group>/50-local`:
  - A file whose sha256 is in `replaces` is a known copy. `init` removes it,
    as one change labelled `remove old copies of the <tool> test parts`,
    whose detail lists each file. `check` and `sync` report it with code 2 and
    `→ Run bootstrap init`. The removal runs before R1c's move, so a known copy
    at the old path is removed, not moved.
  - Another file at `<group>/50-local` is the repo's own, with no warning
    (D16).
- **D22:** the change cannot be declined alone, as R1c D25 rules for its
  move: to decline it, the user declines the `init` run.
- A file is only removed when its tool is selected.
- `describe` and `catalog` show no `replaces`.

## Verification

- `mise x -- mise run test:go` passes, with unit tests for the `replaces`
  parse, a known copy at each of the two paths, an edited copy, and a tool
  that is not selected.

## Guardrails

- Touch nothing outside `cli/`.

## Commit

`feat: a tool part can replace known old copies, and init removes them`
