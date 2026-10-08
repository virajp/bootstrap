# U1 — Coverage harness for the cli tests

- **Wave:** 1
- **Depends on:** —
- **Owns:** `cli/package.json`, `pnpm-lock.yaml`, `cli/vitest.config.ts`
- **Model:** opus
- **Kind:** code
- **Test first:** `mise x -- mise run p:cli:test -- --coverage` exits non-zero
  because the v8 coverage provider `@vitest/coverage-v8` is not installed. Done
  when the same command exits 0 and prints a coverage table for `src/**/*.ts`.
- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `.config/mise/tasks/p/cli/test`; Vitest coverage docs via
  Context7.

## Ruling

| #  | Decision                 | Ruling                                                  | Rejected                      | Unit |
| -- | ------------------------ | ------------------------------------------------------- | ----------------------------- | ---- |
| 11 | G1: no coverage provider | Add `@vitest/coverage-v8` so the coverage gate measures | Continue with `COVERAGE: n/a` | U1   |

## Edits

1. **`cli/package.json`** — add `@vitest/coverage-v8` to `devDependencies`, at
   the major that matches the installed `vitest`. Use the `catalog:` specifier
   if the other dev tools of `cli/package.json` use it; else an exact range in
   the style of the file.
2. **`pnpm-lock.yaml`** — the result of `pnpm install` at the repo root. Do not
   edit it by hand.
3. **`cli/vitest.config.ts`** — keep `provider: "v8"` and
   `include: ["src/**/*.ts"]`. Add a `reporter` list only if the default does
   not print a text table. Add no threshold.

## Verification

- `mise x -- mise run p:cli:test -- --coverage` exits 0 and prints coverage.
- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- Do not touch `.config/mise/tasks/p/cli/test`; the task passes `"$@"` already.
- Add no other package.
- The `ignoreScripts: true` setting of `pnpm-workspace.yaml` stays as it is.

## Commit

`ops: measure cli test coverage with the v8 provider`
