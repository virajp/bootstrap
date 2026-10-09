# U1 — E2E harness for the cli binary

- **Wave:** 1
- **Depends on:** —
- **Owns:** `.config/mise/tasks/p/cli/e2e`, `cli/vitest.config.ts`,
  `cli/vitest.e2e.config.ts` (new), `cli/test/e2e/support/**` (new),
  `cli/test/e2e/smoke.e2e.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/e2e/smoke.e2e.test.ts` — in a temporary git
  repository made by the support helpers, the built `cli/dist/bin.mjs` run with
  no argument exits 2 and prints the top-level help; run with `--version` it
  exits 0 and prints the version of `cli/package.json`. It fails because no E2E
  config, task or helper exists.
- **Read first:** every owned file that exists, `.config/mise/tasks/p/cli/test`,
  `.config/mise/tasks/p/cli/build`, `cli/test/cli.test.ts`.
- **Lazy-load:** Vitest `globalSetup` and project config docs via Context7.

## Ruling

| #  | Decision         | Ruling                                                                                                                          | Rejected              | Unit   |
| -- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------- | --------------------- | ------ |
| 11 | Acceptance tests | The built binary runs in temporary git repositories with a `mise` stub on `PATH`; the two fault criteria run in-process (row 2) | Only in-process tests | U1, U8 |

## Edits

1. **`.config/mise/tasks/p/cli/e2e`** — a task file in the style of `p/cli/test`
   (shebang, `#MISE` description, executable). It runs
   `pnpm exec vitest run --config vitest.e2e.config.ts "$@"` in `cli/`. It
   builds first through a `depends` on `p:cli:build`, so the binary is fresh.
2. **`cli/vitest.e2e.config.ts`** — includes `test/e2e/**/*.e2e.test.ts` only;
   no coverage; a per-test timeout long enough for a spawned process.
3. **`cli/vitest.config.ts`** — exclude `test/e2e/**` from the unit run. Keep
   the rest.
4. **`cli/test/e2e/support/`** — helpers:
   - make a temporary directory with `git init`, a git user, and an optional
     remote `origin` URL; remove it after the test;
   - make a `mise` stub (an executable shell script that exits 0) in a temporary
     `bin` directory, and a `PATH` that has the stub, `git` and `node`, and no
     real `mise` — or no `mise` at all, for the "mise not installed" case;
   - run `node <repo>/cli/dist/bin.mjs <args>` with a given working directory,
     environment and stdin closed, and return the exit code, stdout and stderr;
   - send a signal to a running process (for tests that need it);
   - read the git status of the temporary repository.
5. **`cli/test/e2e/smoke.e2e.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:e2e` exits 0.
- `mise x -- mise run p:cli:test` exits 0 and runs no file under
  `cli/test/e2e/`.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- Do not touch `cli/src/**`.
- The helpers never touch the repo's own working tree; every write goes to an OS
  temporary directory.
- The task file is executable and has no extension.

## Commit

`ops: run the cli binary end to end in temporary git repositories`
