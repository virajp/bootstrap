# U1 — E2E harness: test:e2e and test:e2e:staging

- **Wave:** 1
- **Depends on:** —
- **Owns:** `.config/mise/tasks/test/**`, `cli/vitest.e2e.config.ts`,
  `cli/test/e2e/helpers.ts`, `cli/test/e2e/smoke.test.ts`, `.config/vwf.yaml`
  (the `harness:` block only)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/e2e/smoke.test.ts` — through the helper, run the
  built binary with `--version` inside a fresh temp git repository and assert
  stdout equals the `cli/package.json` version and the exit code is 0. Fails
  because no e2e config, helper or task exists.
- **Read first:** `cli/package.json`, `cli/vitest.config.ts`,
  `cli/tsdown.config.ts`, `.config/mise/tasks/p/cli/*`, `.config/vwf.yaml`.
- **Lazy-load:** the vwf harness contract `assets/harness.md` (task names
  `test:e2e`, `test:e2e:staging`).

## Ruling

| # | Decision    | Ruling                                                                                                                                                                           | Rejected                               | Unit |
| - | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- | ---- |
| 2 | e2e_staging | "Pack-and-install e2e": `test:e2e:staging` packs the cli with `pnpm pack`, installs the tarball into a clean temp directory and runs the e2e suite against that installed binary | park it; publish a registry prerelease | U1   |

## Edits

1. **`cli/vitest.e2e.config.ts`** — a Vitest config for `cli/test/e2e/**` only,
   serial, with a timeout that fits a full `init` run.
2. **`cli/test/e2e/helpers.ts`** — creates a fresh temp git repository (with an
   optional `origin`), runs the binary named by one env variable
   (`BOOTSTRAP_E2E_BIN`, default `cli/dist/bin.mjs` under `node`) with given
   args, env and stdin, and returns stdout, stderr, exit code and a snapshot of
   the repository tree (paths and bytes) for byte-identical assertions.
3. **`cli/test/e2e/smoke.test.ts`** — the test-first case.
4. **`.config/mise/tasks/test/…`** — task `test:e2e`: build the cli, then run
   the e2e config. Task `test:e2e:staging`: build, `pnpm pack` the cli into a
   temp directory, install the tarball into a clean temp project, then run the
   same suite with `BOOTSTRAP_E2E_BIN` pointing at the installed `bootstrap`
   bin. Executable, with a shebang. Pick the file layout that gives exactly
   these two task names and record it in `DECIDED:`.
5. **`.config/vwf.yaml`** — in the `harness:` block only, set `e2e_local: true`
   and `e2e_staging: true`; change nothing else in the file.

## Verification

- `mise x -- mise run test:e2e` exits 0.
- `mise x -- mise run test:e2e:staging` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- Never write outside the OS temp directory in a test.
- `.config/vwf.yaml`: edit the two keys surgically (no `yq -i` whole-file
  rewrite — it reformats the file).

## Commit

`feat: add the cli e2e harness`
