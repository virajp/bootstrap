# U6 — The init command and its acceptance e2e tests

- **Wave:** 4
- **Depends on:** U3, U4, U5
- **Owns:** `cli/src/commands/init/command.ts`, `cli/src/cli.ts`,
  `cli/src/bin.ts`, `cli/test/e2e/init/**`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/e2e/init/*.test.ts` — one e2e test per criterion in
  index.md "Acceptance criteria (from blueprint)" (29), each run through
  `cli/test/e2e/helpers.ts` against the built binary. The write-failure case
  uses a read-only target directory; the interrupt case sets
  `BOOTSTRAP_TEST_PAUSE_AFTER_WRITE` and sends SIGINT while the run pauses.
  Fails because `init` does not exist.
- **Read first:** flow 110 (all of it), `cli/src/cli.ts`, `cli/src/bin.ts`,
  `cli/src/repo/**`, `cli/src/apply/**`, `cli/src/report/**`,
  `cli/src/commands/init/{options,prompts}.ts`, `cli/src/tool/render.ts`,
  `cli/src/setup-config/**`, `cli/test/e2e/helpers.ts`.
- **Lazy-load:** `docs/blueprint/conventions.md` `#errors`, `#backups`.

## Ruling

| # | Decision               | Ruling                                                                                                                                                                                                                                                                                                                     | Rejected                                                                    | Unit   |
| - | ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- | ------ |
| 4 | Interrupt and rollback | rollback is an Effect finalizer on the write phase; `NodeRuntime.runMain` turns Ctrl-C into a graceful interrupt; the command maps an interrupt to exit 130                                                                                                                                                                | a hand-written signal handler                                               | U3, U6 |
| 5 | Fault injection in e2e | "Real faults + test pause": a write failure is a real one (a target directory made read-only); an interrupt uses one hidden env variable `BOOTSTRAP_TEST_PAUSE_AFTER_WRITE=<n>` that pauses after the n-th write so the test sends SIGINT; it does nothing else, is undocumented, and is read only at the process boundary | real faults only (timer SIGINT, flaky); interrupt tested at unit level only | U3, U6 |

## Edits

1. **`cli/src/commands/init/command.ts`** — steps 1–8 of flow 110 in order:
   repository root (exit 3); config present (exit 1, points to
   `bootstrap update`); values and tools (U5); plan and confirm; render core
   plus the chosen tools; apply (U3) with the setup config written last
   (`created` includes it, `files` excludes it); report (U4) with the next
   command `MISE_ENV=dev mise run setup:all`; `--dry-run` stops after the plan
   and reports with `dry_run: true`. Every exit code per `#errors`; every
   `--json` exit prints exactly one document.
2. **`cli/src/cli.ts`** — add `init` as a subcommand of the root command.
3. **`cli/src/bin.ts`** — read `BOOTSTRAP_TEST_PAUSE_AFTER_WRITE` here only and
   pass it to the apply engine; map interrupt to exit 130 and every typed error
   to its exit code.
4. **`cli/test/e2e/init/**`** — the 29 acceptance tests; name each test after
   its criterion so the acceptance verifier can map them.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run test:e2e` exits 0 and runs all 29 init tests.
- `mise x -- mise run test:e2e:staging` exits 0.

## Guardrails

- Business decisions (what to render, what to back up) stay out of the technical
  helpers (`#baseline` business-technical separation).
- Do not edit `cli/src/apply/**`, `cli/src/report/**`, `cli/src/repo/**` — if
  one needs a change, return `UNRESOLVED:` naming it.

## Commit

`feat: add bootstrap init`
