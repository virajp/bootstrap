# U6 — The remove pipeline and the acceptance tests

- **Wave:** 4
- **Depends on:** U4, U5
- **Owns:** `cli/src/remove/run.ts` (new), `cli/src/remove/command.ts`,
  `cli/test/remove/run.test.ts` (new), `cli/test/e2e/remove.e2e.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/e2e/remove.e2e.test.ts` — one test per criterion of *Acceptance
    criteria (from blueprint)* in `index.md`, on the built binary in temporary
    git repositories set up with `init` (plan 2 U1 helpers), except the
    write-failure, failed-restore and interrupt criteria. Each test names its
    criterion.
  - `cli/test/remove/run.test.ts` (in-process, with the FileSystem seam of plan
    2 U5) — the write-failure, failed-restore and interrupt criteria through the
    remove pipeline.
  - `cli/test/remove/run.test.ts` also covers, in-process with the fixture
    catalog layer, every criterion that needs two tools of one `max: one`
    category (row 6).
  - They fail because the handler is the placeholder of U4.
- **Read first:** `docs/blueprint/flows/cli/150-remove-tool/index.md` (the whole
  flow), `cli/src/remove/{command,usage,validate}.ts`,
  `cli/src/command/pipeline.ts`, `cli/src/change/compute.ts`,
  `cli/src/init/run.ts` (as the model).
- **Lazy-load:** `cli/src/setup-config/**`, `cli/src/write/apply.ts`,
  `cli/test/e2e/support/**`.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                        | Rejected                  | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------- | -------------- |
| 3 | Order in step 2                                         | Validity (exit 3), then the version guard (exit 1), then the repair. An older `version` exits 1 with no repair warning                                                                                                                                        | Repair first              | U6             |
| 5 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` already covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/change/`, `cli/src/command/` and `cli/src/remove/` | An allow-list of literals | U1, U4, U5, U6 |
| 6 | Criteria that need two tools of one `max: one` category | They run in-process with the fixture catalog layer of plan 2 (its row 12); the 1.0 catalog has no such pair                                                                                                                                                   | Leave them uncovered      | U6             |

## Edits

1. **`cli/src/remove/run.ts`** — the pipeline in flow order on the shared
   skeleton: usage (U4), preflight, `requireSetUp`, Validity, `versionGuard`,
   repair (row 3), name validation and the `-y` rule (U5), the exit-0 shortcut
   (render only to find `orphaned`), render of the new selection, change set in
   `remove` mode (U1), target checks, write (unless `--dry-run`), report with
   the remove key set (U2).
2. **`cli/src/remove/command.ts`** — replace the placeholder of U4 with
   `run.ts`.
3. **Tests** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:build` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0 and runs a test for every acceptance
  criterion except the in-process criteria.
- `grep -rn "not implemented" cli/src` finds nothing.

## Guardrails

- Do not edit `cli/src/change/**`, `cli/src/command/**`, `cli/src/report/**`,
  `cli/src/init/**`, `usage.ts` or `validate.ts`. A defect there is a `GAP:`
  naming the unit.
- Remove never commits.

## Commit

`feat: run bootstrap remove end to end`
