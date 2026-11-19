# U13 — The own tasks of this repository in Python

- **Read first:** `.config/mise/tasks/test/go`,
  `.config/mise/tasks/test/50-local`
  and `.config/mise/tasks/vorssaint/diff`, top to bottom, and
  `.config/mise/tasks/_scripts/helpers.py` (the synced copy that the self-sync
  of the foundation plan wrote).
- **Lazy-load:** CLAUDE.md "Tests" (what `test:go` and `test/50-local` run),
  `macos/tools/vorssaint/` (the defaults that `vorssaint:diff` compares).

## Ruling

- **D13:** "The own tasks of this repository (test/go, test/50-local,
  vorssaint/diff) source the shell helpers at the root. The self-sync after
  this plan removes that copy, so they move to Python in this plan, with the
  rules of D1 and D2. U13 also deletes the stale synced copy
  .config/mise/tasks/_scripts/repo-sync: no file sources it and no manifest
  row lists it, so the stale rule never removes it."
- **D13 (tools):** "An own task keeps its own tools (test/50-local:
  gotestsum and gomplate) in the one inline table of foundation D29 and calls
  them directly, as now."
- These files are own files of this repository, not synced copies: they have
  no Synced from bootstrap header, and they keep none.
- `vorssaint/diff` embeds Python in bash now: its Python part becomes the body
  of `diff.py`, with the same flags, output and exit codes (exit 1 when a key
  differs or is missing).

- **D14:** "Foundation D21 applies: mode 755 on each new file,
  sys.dont_write_bytecode = True in the preamble, python -B in a hand check,
  and a Go test of the engine from cli/. A unit checks the inventory of the
  tools.toml files with the engine: cd cli && BOOTSTRAP_CLONE=<worktree root>
  mise x -- go run ./cmd/bootstrap catalog exits 0."

## Edits

1. `test/go` becomes `test/go.py`, `test/50-local` becomes `test/50-local.py`,
   `vorssaint/diff` becomes `vorssaint/diff.py`, each with the uv form of the
   foundation plan without the Synced header, and the import preamble. Delete
   each old file with `rm`.
2. `rm .config/mise/tasks/_scripts/repo-sync`.

## Verification

- `mise trust` on each new file, then `mise tasks ls` lists `test:go`,
  `test:50-local` and `vorssaint:diff` with their descriptions.
- `mise x -- mise run test:go` passes.
- `mise x ruff@latest -- ruff check` passes on the three files.

## Guardrails

- Do not edit a synced copy below `.config/mise/tasks/` (a file with the
  Synced from bootstrap header). The one exception is the `rm` of
  `_scripts/repo-sync` (D14).
- Delete with `rm`, never `git rm`.
- Write each comment in ASD-STE100.

## Commit

`refactor: the own test and vorssaint tasks in Python, and delete the stale
repo-sync`
