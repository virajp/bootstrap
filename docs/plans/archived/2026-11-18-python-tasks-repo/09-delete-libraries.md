# U9 — Delete the shell libraries

- **Read first:** `repo/tools/mise/tools.toml`, `repo/tools/git/tools.toml`
  and `repo/tools/swift/tools.toml` (the `_scripts/*` entries).
- **Lazy-load:** `docs/design/cli.md` ("The manifest and the stale rule").

## Ruling

- **D8:** "U9 deletes helpers, checks, placeholder, worktrees-page (mise),
  merge (git) and xcode (swift) with rm, and removes their
  [tasks.\"_scripts/<name>\"] entries. In a consumer, the stale rule removes
  the synced copies at the next sync."
- **D14:** "U9 searches repo/tools/*/tasks/ and the three own tasks of U13 only
  (not
  the synced copies at the root of .config/mise/tasks/, which the self-sync
  replaces; not macos/; not the tools.toml files)."
- **D7:** "U9 owns the three tools.toml files in wave 2 and removes only the
  library entries."

## Edits

1. `grep -rlE
   '_scripts/(helpers|checks|placeholder|worktrees-page|merge|xcode)([^._]|$)'
   repo/tools/*/tasks/ .config/mise/tasks/test/go.py
   .config/mise/tasks/test/50-local.py .config/mise/tasks/vorssaint/diff.py`
   must find no task file that sources a shell library (D14: not the synced
   copies at the root, which the self-sync replaces; not `macos/`, whose Mac
   helpers are plan 3; not the `tools.toml` files). If it finds one, stop and
   return a QUESTION.
2. `rm` the six shell libraries.
3. Remove the `[tasks."_scripts/helpers"]`, `checks`, `placeholder`,
   `worktrees-page`, `merge` and `xcode` entries, with their comments. Keep
   the `.py` entries and `_scripts/gate_files.py`.
4. Update the comments of the three `tools.toml` files that name a shell
   library (for example the header comment of `repo/tools/swift/tools.toml`).

## Verification

- `cd cli && BOOTSTRAP_CLONE=<worktree root> mise x -- go run ./cmd/bootstrap
  catalog`
  exits 0.
- The grep of edit 1 finds nothing.

## Guardrails

- Do not edit a task key: U3, U4 and U6 renamed them in wave 1.
- Do not edit `macos/`: the Mac helpers are plan 3.
- Delete with `rm`, never `git rm`.

## Commit

`refactor: delete the shell libraries of the tasks`
