# U2 — Selection of add and the shared replacement rule

- **Wave:** 1
- **Depends on:** —
- **Owns:** `cli/src/add/select.ts` (new), `cli/src/change/replace.ts` (new),
  `cli/src/init/resolve.ts`, `cli/test/add/select.test.ts` (new),
  `cli/test/change/replace.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/change/replace.test.ts` (pure, with the fixture catalog layer of
    plan 2): a different tool in a held `max: one` category is a replacement;
    without `--replace` exit 2; `-y` is not consent; a held tool whose
    `replaceable` is false exits 2, also with `--replace`; `--replace` with
    nothing to replace is ignored; the result lists `{from, to}`. The init
    resolution tests of plan 2 stay green, now on the shared rule.
  - `cli/test/add/select.test.ts` (pure): repeated names are used once; a name
    already in `values.tools` goes to `already_added`; the selection after the
    request is the held tools minus the replaced ones plus the named ones; a
    named tool whose `requires` tool is not in it exits 2 naming the required
    tool; a replacement of a tool that another selected tool requires exits 2
    naming the dependent tool (fixture); two exit-2 findings are reported
    together (row 3); every name already added and no repair → the exit-0
    shortcut; `-y` changes nothing.
  - They fail because `select.ts` and `replace.ts` do not exist.
- **Read first:** `docs/blueprint/flows/cli/140-add-tool/index.md` (steps 3 and
  4), `docs/blueprint/entities/tool-category/index.md` (invariant 3),
  `cli/src/init/resolve.ts`, `cli/src/report/errors.ts`.
- **Lazy-load:** `cli/src/catalog/service.ts`,
  `cli/test/support/catalog-fixture.ts`, `cli/src/tool/catalog.ts`.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                                                                       | Rejected              | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------- | -------------- |
| 2 | One replacement rule                                    | The replacement rule of init (plan 2 U4) moves into a shared module that init and add use: `--replace` is necessary, `-y` is never consent, `replaceable: false` exits 2, `--replace` with nothing to replace is ignored                                                                                     | A second copy for add | U2             |
| 3 | Two exit-2 findings in one step                         | Report every exit-2 finding of a step together (precedent: flow 160 "all usage errors are found together"; the same as plan 3)                                                                                                                                                                               | First finding only    | U2             |
| 5 | Criteria that need two tools of one `max: one` category | They run in-process with the fixture catalog layer of plan 2 (its row 12): the replacement criteria, the dependent-of-replaced criterion, two named tools of one category, two recorded tools of one category, and another tool held in place of a missing unremovable one; the 1.0 catalog has no such pair | Leave them uncovered  | U1, U2, U3, U4 |
| 6 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/add/`                                                                                                     | An allow-list         | U1, U2, U3, U4 |

## Edits

1. **`cli/src/change/replace.ts`** — the replacement rule, moved out of
   `cli/src/init/resolve.ts` with its behaviour unchanged.
2. **`cli/src/init/resolve.ts`** — call the shared rule; no behaviour change.
3. **`cli/src/add/select.ts`** — the step 3 and step 4 rules of add; it returns
   the added tools, the `already_added` names, the replaced pairs, the new
   selection and the shortcut flag, or a usage error (exit 2).
4. **Tests** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0 (every init E2E test stays green).

## Guardrails

- No git, no disk, no prompt: pure functions.
- Init behaviour does not change; do not edit the init tests.
- Do not touch `cli/src/change/compute.ts` (U3).

## Commit

`feat: select the tools to add with one shared replacement rule`
