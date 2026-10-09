# U3 — Add mode of the change computation and the add report keys

- **Wave:** 1
- **Depends on:** —
- **Owns:** `cli/src/change/compute.ts`, `cli/src/report/human.ts`,
  `cli/test/change/add-mode.test.ts` (new), `cli/test/report/output.test.ts`
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/change/add-mode.test.ts` (in a temporary git repository): in `add`
    mode, the full file set of the new selection is rendered; only paths that
    differ from the working copy are targets, shared files included; the files
    of a replaced tool are `deleted`, except its `create_only` files, which are
    `kept` (fixture catalog layer); an existing `create_only` file of a new tool
    is `kept`; `files` is refreshed in full mode; the setup config follows the
    add rewrite rule (rewritten only on a tool change or a repair). The
    init-mode and remove-mode tests stay green.
  - `cli/test/report/output.test.ts`: an add result prints `added` with `+`
    (success) and `already_added` with `·` (muted), in the key order of flow 140
    step 7; under `--json` every key except `next_command` is present.
  - They fail because no `add` mode and no add glyphs exist.
- **Read first:** `cli/src/change/compute.ts`, `cli/src/report/human.ts`,
  `docs/blueprint/flows/cli/140-add-tool/index.md` (steps 5–7),
  `docs/blueprint/design-system.md` `#terminal-ux`.
- **Lazy-load:** `cli/src/setup-config/**`, `cli/src/report/json.ts`.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                                                                       | Rejected             | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------- | -------------- |
| 1 | Shared code (user, 2026-10-09)                          | Add uses the shared modules of plan 3. The change computation gets an `add` mode: the init classification, plus the deletion of a replaced tool's files as in remove, plus the add rewrite rule                                                                                                              | Own modules for add  | U3             |
| 5 | Criteria that need two tools of one `max: one` category | They run in-process with the fixture catalog layer of plan 2 (its row 12): the replacement criteria, the dependent-of-replaced criterion, two named tools of one category, two recorded tools of one category, and another tool held in place of a missing unremovable one; the 1.0 catalog has no such pair | Leave them uncovered | U1, U2, U3, U4 |
| 6 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/add/`                                                                                                     | An allow-list        | U1, U2, U3, U4 |

## Edits

1. **`cli/src/change/compute.ts`** — the `add` mode.
2. **`cli/src/report/human.ts`** — the glyph and role entries for `added` and
   `already_added`.
3. **Tests** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- The init and remove modes do not change; do not edit their tests.
- Do not touch `cli/src/change/replace.ts` (U2).

## Commit

`feat: compute and report an add change set`
