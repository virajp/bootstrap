# U2 — Report with a key set per command

- **Wave:** 1
- **Depends on:** —
- **Owns:** `cli/src/report/human.ts`, `cli/src/report/json.ts`,
  `cli/test/report/output.test.ts`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/report/output.test.ts` — new cases for a remove
  result: human groups and JSON keys in the order `removed`, `not_added`,
  `deleted`, `created`, `changed`, `unchanged`, `kept`, `orphaned`, `warnings`;
  `removed` with `-` (emphasis) and `not_added` with `·` (muted); `removed` and
  `not_added` sorted by name, path lists sorted by path, `warnings` in the order
  found; under `--json` every key present (empty lists included), and no
  `next_command` unless the next command is printed; the closing line from a
  repair or a touched `mise` config file only. They fail because the key set is
  fixed to init's. The init cases stay green.
- **Read first:** `cli/src/report/human.ts`, `cli/src/report/json.ts`,
  `cli/test/report/output.test.ts`, `docs/blueprint/design-system.md`
  `#terminal-ux`, `docs/blueprint/flows/cli/150-remove-tool/index.md` (step 9).
- **Lazy-load:** `docs/blueprint/conventions.md` `#errors`.

## Ruling

| # | Decision                       | Ruling                                                                                                                                                                                           | Rejected                                        | Unit       |
| - | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------- | ---------- |
| 1 | Shared code (user, 2026-10-09) | Plan 3 moves the init-only parts into shared modules: a change computation with a mode, a report with a key set, common flags and a pipeline skeleton. Init calls them, and its tests stay green | Revise plan 2 now; separate modules per command | U1, U2, U3 |

## Edits

1. **`cli/src/report/human.ts`**, **`cli/src/report/json.ts`** — the key set and
   its order become a parameter of the result; the glyph and role table gains
   `removed` and `not_added`; the init key set stays the default for init.
2. **`cli/test/report/output.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Do not touch `cli/src/report/errors.ts` or `progress.ts`.
- Init output does not change.

## Commit

`refactor: make the result report take a key set per command`
