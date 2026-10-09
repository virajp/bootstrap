# U1 — The add command and its usage errors

- **Wave:** 1
- **Depends on:** —
- **Owns:** `cli/src/add/command.ts` (new), `cli/src/add/usage.ts` (new),
  `cli/src/cli.ts`, `cli/test/add/usage.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/add/usage.test.ts` (in-process, in a directory that
  is not a git repository) — `add` with no name, with an unknown tool name, with
  an unknown flag or a bad flag value exits 2 with the short usage and writes
  nothing, before any git or disk access; two distinct names of one `max: one`
  category exit 2 after repeated names are merged (with the fixture catalog
  layer of plan 2); a name given twice is not an error; `add --help` exits 0
  with the add help; `add` takes the common flags and `--replace`, and no
  `--tool`, `--repo` or scope flag. It fails because `add` is not registered.
- **Read first:** `docs/blueprint/flows/cli/140-add-tool/index.md` (usage
  errors, Modes), `docs/blueprint/design-system.md` `#terminal-ux` (help),
  `cli/src/cli.ts`, `cli/src/command/flags.ts`, `cli/src/remove/command.ts` and
  `usage.ts` (as the model).
- **Lazy-load:** `cli/src/catalog/service.ts`,
  `cli/test/support/catalog-fixture.ts`.

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                                                                       | Rejected             | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------- | -------------- |
| 5 | Criteria that need two tools of one `max: one` category | They run in-process with the fixture catalog layer of plan 2 (its row 12): the replacement criteria, the dependent-of-replaced criterion, two named tools of one category, two recorded tools of one category, and another tool held in place of a missing unremovable one; the 1.0 catalog has no such pair | Leave them uncovered | U1, U2, U3, U4 |
| 6 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/add/`                                                                                                     | An allow-list        | U1, U2, U3, U4 |

## Edits

1. **`cli/src/add/command.ts`** — the `add <tool>...` subcommand with the common
   flags and `--replace`; the handler runs `usage.ts`, then a placeholder that
   U4 replaces.
2. **`cli/src/add/usage.ts`** — the usage checks, pure, over the parsed input
   and the catalog from the `Catalog` service; all errors found are reported
   together.
3. **`cli/src/cli.ts`** — register `add`.
4. **`cli/test/add/usage.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Do not touch `cli/src/init/**`, `cli/src/remove/**` or `cli/src/command/**`.
- Copy message texts from the flow byte for byte.

## Commit

`feat: add the add command surface with usage errors`
