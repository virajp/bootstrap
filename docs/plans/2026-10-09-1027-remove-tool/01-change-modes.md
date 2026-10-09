# U1 — Shared change computation with an init and a remove mode

- **Wave:** 1
- **Depends on:** —
- **Owns:** `cli/src/change/**` (new), `cli/src/init/changes.ts` (moved out and
  deleted), `cli/test/change/**` (new), `cli/test/init/changes.test.ts` (moved
  out and deleted); plus the one import line in `cli/src/init/run.ts`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/change/remove-mode.test.ts` (in a temporary git
  repository) — one case per row of the step 5 table of
  [Remove a tool](../../blueprint/flows/cli/150-remove-tool/index.md): each
  recorded or rendered path gets the target decision, the report key and the
  `files` decision of its row; a shared file is one of the "Selection-dependent
  files" of the Tool entity, read from the catalog data, not a literal list in
  logic; `.config/bootstrap.yaml` is a target listed `changed` and never in
  `files`. It fails because no remove mode exists. The moved init cases
  (`cli/test/change/init-mode.test.ts`) stay green.
- **Read first:** `cli/src/init/changes.ts`, `cli/test/init/changes.test.ts`,
  `docs/blueprint/flows/cli/150-remove-tool/index.md` (step 5),
  `docs/blueprint/entities/setup-config/index.md` (invariants 2, 4, 5, 6, 7),
  `docs/blueprint/entities/tool/index.md` (Selection-dependent files).
- **Lazy-load:** `cli/src/setup-config/**` (refresh modes and the rewrite rule
  of plan 1), `cli/src/git/git.ts`.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                        | Rejected                                        | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- | -------------- |
| 1 | Shared code (user, 2026-10-09)                          | Plan 3 moves the init-only parts into shared modules: a change computation with a mode, a report with a key set, common flags and a pipeline skeleton. Init calls them, and its tests stay green                                                              | Revise plan 2 now; separate modules per command | U1, U2, U3     |
| 4 | Remove-mode cases                                       | The change computation gets a remove mode for the step 5 table: a shared file is created again, a non-shared absent file is left alone, unrecorded paths are not reported, the files of a repaired tool are written                                           | A separate remove computation                   | U1             |
| 5 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` already covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/change/`, `cli/src/command/` and `cli/src/remove/` | An allow-list of literals                       | U1, U4, U5, U6 |

## Edits

1. **`cli/src/change/compute.ts`** — the change computation of plan 2 U6, moved
   with its init behaviour unchanged, plus a mode parameter: `init` (plan 2
   behaviour; refresh in full mode, init rewrite rule) and `remove` (the step 5
   table; refresh in remove mode, remove rewrite rule). Inputs are command
   neutral: the recorded config, the old and new selection, the values, the
   removed and replaced tools, the repaired tools.
2. **`cli/src/init/changes.ts`** — delete with `rm`; move its tests to
   `cli/test/change/init-mode.test.ts` unchanged.
3. **`cli/src/init/run.ts`** — change only the import to the moved module.
4. **`cli/test/change/remove-mode.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0 (every init E2E test stays green).

## Guardrails

- Init behaviour does not change; the moved init tests are not edited.
- Do not touch `cli/src/report/**` (U2) or anything else in `cli/src/init/`.
- Delete with `rm`, never `git rm`.

## Commit

`refactor: share the change computation with a remove mode`
