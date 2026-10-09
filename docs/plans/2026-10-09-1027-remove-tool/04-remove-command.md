# U4 — The remove command and its usage errors

- **Wave:** 3
- **Depends on:** U3
- **Owns:** `cli/src/remove/command.ts` (new), `cli/src/remove/usage.ts` (new),
  `cli/src/cli.ts`, `cli/test/remove/usage.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/remove/usage.test.ts` (in-process, in a directory
  that is not a git repository) — `remove` with no name, with an unknown tool
  name, with an unknown flag or a bad flag value exits 2 with the short usage
  and writes nothing, before any git or disk access; `remove --help` exits 0
  with the remove help; `remove` takes the common flags of U3 and no `--tool`,
  `--repo` or scope flag. It fails because `remove` is not registered.
- **Read first:** `docs/blueprint/flows/cli/150-remove-tool/index.md` (usage
  errors, Modes), `docs/blueprint/design-system.md` `#terminal-ux` (help),
  `cli/src/cli.ts`, `cli/src/command/flags.ts`, `cli/src/init/usage.ts`.
- **Lazy-load:** the `effect/cli` docs via Context7 (variadic positional
  arguments).

## Ruling

| # | Decision                                                | Ruling                                                                                                                                                                                                                                                        | Rejected                  | Unit           |
| - | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------- | -------------- |
| 5 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool or category name in logic; the literal scan of `cli/test/catalog-driven.test.ts` already covers every `.ts` file under `cli/src` except the two catalogs and `cli/src/git/`, so it covers `cli/src/change/`, `cli/src/command/` and `cli/src/remove/` | An allow-list of literals | U1, U4, U5, U6 |

## Edits

1. **`cli/src/remove/command.ts`** — the `remove <tool>...` subcommand with the
   common flags; the handler runs `usage.ts`, then a placeholder that U6
   replaces.
2. **`cli/src/remove/usage.ts`** — the usage checks, pure, over the parsed input
   and the catalog; all errors found are reported together.
3. **`cli/src/cli.ts`** — register `remove`.
4. **`cli/test/remove/usage.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Do not touch `cli/src/init/**` or `cli/src/command/**`.
- Copy message texts from the flow byte for byte.

## Commit

`feat: add the remove command surface with usage errors`
