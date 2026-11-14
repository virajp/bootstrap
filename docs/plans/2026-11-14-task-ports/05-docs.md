# U5 — Docs

- **Read first:** the branch delta, then each passage named below.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. The task and engine rulings are quoted in
`01-test-dispatchers.md` and `02-init-move.md`.

- **D5**, an amendment of the design record: R1c ships no `p:<id>:deploy`. A
  patterned port has no default; it exists for an id only when a task for that
  id exists. The rule "every port exists in every shaped repo" holds for the
  plain ports only.
- **D1:** R1c covers `test`, `test:e2e` and `test:golden`; R2 makes
  `git:merge:<branch>`, `git:worktree:*` and `code:lsp:check` with real work.

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings plus every
   `DOCS FALSIFIED:` line of the earlier units.
2. **`docs/design/mise.md`** — the orchestrator table (about lines 583-591):
   the three test dispatchers; the task-script rule (about line 638):
   parts are `NN-<name>` files, and `own_tasks` skips any other file and every
   folder; a task script that does not apply prints nothing, and a dispatcher
   with no part prints one line (D16).
3. **`docs/sync/tools/mise.md`** (about lines 68-82) — the three dispatchers.
4. **`docs/design/packs.md`** (the base pack table, about lines 190-203) — the
   three dispatchers.
5. **`CLAUDE.md`** ("Task scripts", about lines 145-195) — the three
   dispatchers and the `NN-` rule.
6. **`docs/design/cli.md`** — the `init` move (D4, D7, D8, D12), the code 2
   of `check` and `sync` for a file at a dispatcher path, and the D15 warning.
7. **`docs/sync/migrate.md`** — a migrate item: run `bootstrap init`; a repo's
   own `test`, `test/_default`, `test/e2e/_default` or `test/golden/_default`,
   and a stray `test/e2e` or `test/golden` file, become `50-local` parts
   (D7, D12). A `test` task in `mise.toml` `[tasks.test]` or in another task
   folder: move it by hand into `.config/mise/tasks/test/50-local` (D14).
   A part without an `NN-` name in a dispatcher folder no longer runs: rename
   it to `NN-<name>`; `check` warns about it (D15).
8. **`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`** — add to
   the amendments section (or add one, "Amendments 2026-11-14 (R1c planning)",
   if none exists): D1 and D5 of this plan. Do not rewrite the rest of the
   record.

## Verification

- `mise x -- mise run code:format` passes.

## Guardrails

- Touch nothing outside the owned paths. `docs/plans/` is not this unit's.
- Delete with `rm`, never `git rm`.

## Commit

`docs: the test ports and the init move`
