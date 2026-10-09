# U3 — Common flags and the pipeline skeleton

- **Wave:** 2
- **Depends on:** U1, U2
- **Owns:** `cli/src/command/**` (new), `cli/src/init/command.ts`,
  `cli/src/init/run.ts`, `cli/test/command/**` (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/command/pipeline.test.ts` (in-process) — a test
  command built from the shared parts: the common flags (`-y`/`--yes`,
  `--dry-run`, `--json`, `-q`/`--quiet`, `-v`/`--verbose`, `--no-color`,
  `--help`) parse as in init; the skeleton runs the preflight, skips the write
  under `--dry-run`, prints the result with the command's key set, and on an
  error prints the error document with `"dry_run": true` under
  `--dry-run --json`. It fails because `cli/src/command/` does not exist. Every
  init test and init E2E test stays green.
- **Read first:** `cli/src/init/command.ts`, `cli/src/init/run.ts`,
  `cli/src/report/*.ts`, `cli/src/write/apply.ts`, `cli/src/git/preflight.ts`.
- **Lazy-load:** the `effect/cli` docs via Context7.

## Ruling

| # | Decision                       | Ruling                                                                                                                                                                                           | Rejected                                        | Unit       |
| - | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------- | ---------- |
| 1 | Shared code (user, 2026-10-09) | Plan 3 moves the init-only parts into shared modules: a change computation with a mode, a report with a key set, common flags and a pipeline skeleton. Init calls them, and its tests stay green | Revise plan 2 now; separate modules per command | U1, U2, U3 |

## Edits

1. **`cli/src/command/flags.ts`** — the common flags, moved out of
   `cli/src/init/command.ts`.
2. **`cli/src/command/pipeline.ts`** — the skeleton moved out of
   `cli/src/init/run.ts`: preflight, the command's own steps, write unless
   `--dry-run`, print, the `dry_run` error document, progress.
3. **`cli/src/init/command.ts`**, **`cli/src/init/run.ts`** — use the shared
   parts; init behaviour unchanged.
4. **`cli/test/command/pipeline.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Init behaviour and output do not change; do not edit init tests.
- Do not touch `cli/src/change/**` (U1) or `cli/src/report/**` (U2).

## Commit

`refactor: share the command flags and the pipeline skeleton`
