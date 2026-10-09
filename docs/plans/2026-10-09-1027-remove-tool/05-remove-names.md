# U5 — Name validation of remove (pure)

- **Wave:** 3
- **Depends on:** U3
- **Owns:** `cli/src/remove/validate.ts` (new),
  `cli/test/remove/validate.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/remove/validate.test.ts` — a pure function over the
  named tools, the recorded `values.tools`, the repaired tools and the flags;
  tool names come from the catalog data or are data to check results:
  - A repeated name is used once, with no error.
  - A tool whose `removable` is false exits 2.
  - A tool that a tool still selected after the request `requires` exits 2
    naming the dependent tool (through `checkRequires`); removing both in one
    request passes.
  - An unremovable name and a dependent tool in one request give one exit-2
    error that names both findings (row 2).
  - A name not in `values.tools` goes to `not_added`.
  - No named tool in `values.tools` and no repair → the exit-0 shortcut (report
    `not_added` and `orphaned` only, no `-y` needed).
  - A repair with no named tool selected → the run continues.
  - Otherwise, without `-y` and without `--dry-run`, exit 2 "removing files
    needs -y"; `--dry-run` needs no `-y`.
  - It fails because `validate.ts` does not exist.
- **Read first:** `docs/blueprint/flows/cli/150-remove-tool/index.md` (steps 3
  and 4), `cli/src/report/errors.ts`, `cli/src/tool/catalog.ts`.
- **Lazy-load:** `cli/src/setup-config/**`.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                        | Rejected                  | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------- | -------------- |
| 2 | Two exit-2 findings in one step                         | Report every exit-2 finding of step 3 together (precedent: flow 160 "all usage errors are found together")                                                                                                                                                    | First finding only        | U5             |
| 5 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` already covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/change/`, `cli/src/command/` and `cli/src/remove/` | An allow-list of literals | U1, U4, U5, U6 |

## Edits

1. **`cli/src/remove/validate.ts`** — the step 3 and step 4 rules; it returns
   the removed tools (sorted by name), the `not_added` names (sorted by name),
   the shortcut flag, or a usage error (exit 2).
2. **`cli/test/remove/validate.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- No git, no disk, no prompt: a pure function.
- Do not touch `cli/src/remove/command.ts` or `usage.ts` (U4).
- Copy message texts from the flow byte for byte.

## Commit

`feat: validate the tools named for removal`
