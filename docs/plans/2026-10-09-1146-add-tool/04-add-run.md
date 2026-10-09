# U4 — The add pipeline and the acceptance tests

- **Wave:** 2
- **Depends on:** U1, U2, U3
- **Owns:** `cli/src/add/run.ts` (new), `cli/src/add/command.ts`,
  `cli/test/add/run.test.ts` (new), `cli/test/e2e/add.e2e.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/e2e/add.e2e.test.ts` — one test per criterion of *Acceptance
    criteria (from blueprint)* in `index.md`, on the built binary in temporary
    git repositories set up with `init` (plan 2 U1 helpers), except the
    write-failure, failed-restore and interrupt criteria and the criteria of row
    5. Each test names its criterion.
  - `cli/test/add/run.test.ts` (in-process) — the write-failure, failed-restore
    and interrupt criteria with the FileSystem seam of plan 2, and the criteria
    of row 5 with the fixture catalog layer of plan 2. Each test names its
    criterion.
  - They fail because the handler is the placeholder of U1.
- **Read first:** `docs/blueprint/flows/cli/140-add-tool/index.md` (the whole
  flow), `cli/src/add/{command,usage,select}.ts`, `cli/src/command/pipeline.ts`,
  `cli/src/change/compute.ts`, `cli/src/remove/run.ts` (as the model).
- **Lazy-load:** `cli/src/setup-config/**`, `cli/src/write/apply.ts`,
  `cli/test/e2e/support/**`, `cli/test/support/catalog-fixture.ts`.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                                                                       | Rejected             | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------- | -------------- |
| 4 | Order in step 2                                         | Validity (exit 3), then the version guard (exit 1), then the repair (flow 140 step 2). An older `version` exits 1 with no repair warning                                                                                                                                                                     | Repair first         | U4             |
| 5 | Criteria that need two tools of one `max: one` category | They run in-process with the fixture catalog layer of plan 2 (its row 12): the replacement criteria, the dependent-of-replaced criterion, two named tools of one category, two recorded tools of one category, and another tool held in place of a missing unremovable one; the 1.0 catalog has no such pair | Leave them uncovered | U4             |
| 6 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/add/`                                                                                                     | An allow-list        | U1, U2, U3, U4 |

## Edits

1. **`cli/src/add/run.ts`** — the pipeline in flow order on the shared skeleton:
   usage (U1), preflight, `requireSetUp`, Validity, `versionGuard`, repair (row
   4), selection (U2), the exit-0 shortcut (render only to find `orphaned`, no
   target checks), render of the new selection, change set in `add` mode (U3),
   target checks, write (unless `--dry-run`), report with the add key set.
2. **`cli/src/add/command.ts`** — replace the placeholder of U1 with `run.ts`.
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
  `cli/src/init/**`, `cli/src/remove/**`, `usage.ts` or `select.ts`. A defect
  there is a `GAP:` naming the unit.
- Add never commits and never prompts.

## Commit

`feat: run bootstrap add end to end`
