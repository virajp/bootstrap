# U8 — The init pipeline and the acceptance tests

- **Wave:** 5
- **Depends on:** U5, U6, U7
- **Owns:** `cli/src/init/run.ts` (new), `cli/src/init/command.ts`,
  `cli/test/init/run.test.ts` (new), `cli/test/e2e/init.e2e.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/e2e/init.e2e.test.ts` — one test per criterion of *Acceptance
    criteria (from blueprint)* in `index.md`, on the built binary in temporary
    git repositories (U1 helpers), except the write-failure and the interrupt
    criteria and the criteria that need two tools of one `max: one` category
    (row 12). Each test names its criterion.
  - `cli/test/init/run.test.ts` (in-process, with the FileSystem seam of U5):
    "Given a write failure injected mid-run…" and "Given an interrupt signal
    mid-write…" end to end through the pipeline — the tree is as before and the
    exit codes are 3 (naming the failing path) and 130.
  - `cli/test/init/run.test.ts` also covers, in-process with the fixture catalog
    layer, every criterion that needs two tools of one `max: one` category (row
    12). Each test names its criterion.
  - They fail because the handler is the placeholder of U2.
- **Read first:** `docs/blueprint/flows/cli/110-setup-repository/index.md` (the
  whole flow), `cli/src/init/command.ts`, `usage.ts`, `resolve.ts`,
  `changes.ts`, `cli/src/git/preflight.ts`, `cli/src/write/apply.ts`,
  `cli/src/report/*.ts`.
- **Lazy-load:** `cli/src/setup-config/**`, `cli/src/tool/render.ts`.

## Ruling

| #  | Decision                                                     | Ruling                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Rejected                                             | Unit           |
| -- | ------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | -------------- |
| 2  | Fault tests (write failure, interrupt mid-write)             | In-process FileSystem seam that fails or stops at the n-th write; no test code ships in the binary (user, 2026-10-09)                                                                                                                                                                                                                                                                                                                                                                                                                                      | Environment variable in the binary; timing           | U5, U8         |
| 10 | Command surface before the pipeline                          | U2 registers `init`, its flags and usage checks; its handler ends in a placeholder that U8 replaces with the pipeline before the review                                                                                                                                                                                                                                                                                                                                                                                                                    | One unit for both                                    | U2, U8         |
| 11 | Acceptance tests                                             | The built binary runs in temporary git repositories with a `mise` stub on `PATH`; the two fault criteria run in-process (row 2)                                                                                                                                                                                                                                                                                                                                                                                                                            | Only in-process tests                                | U1, U8         |
| 12 | Replacement criteria with the 1.0 catalog (user, 2026-10-09) | Every reader of catalog data (the catalog functions, the renderer, the setup-config rules, the command code) gets the catalog and the template sources from an Effect `Catalog` service; its default layer is the static data. A test fixture catalog holds a `max: one` category with two replaceable tools. A criterion that needs two tools of one `max: one` category (a replacement, two such tools named or recorded, or another tool held in place of a missing unremovable one) runs in-process with the fixture; the 1.0 catalog has no such pair | Catalog as a parameter; leave the criteria uncovered | U2, U4, U6, U8 |

## Edits

1. **`cli/src/init/run.ts`** — the pipeline in flow order: usage check (U2),
   preflight (U3), read the setup config — absent is a first run; present is
   validated and repaired (plan 1); init takes no version guard, so an older
   `version` or `format` is an upgrade re-run — then resolution (U4), change set
   (U6), write (U5) unless `--dry-run`, print (U7). `--dry-run` runs the usage
   check and steps 1–6 and writes nothing.
2. **`cli/src/init/command.ts`** — replace the placeholder of U2 with `run.ts`.
3. **Tests** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:build` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0 and runs a test for every acceptance
  criterion except the two fault criteria and the row 12 criteria.
- `grep -rn "not implemented" cli/src` finds nothing.

## Guardrails

- Do not edit `usage.ts`, `resolve.ts`, `changes.ts`, `cli/src/git/**`,
  `cli/src/write/**` or `cli/src/report/**`. A defect there is a `GAP:` naming
  the unit.
- Init never installs software and never runs `setup:all`.

## Commit

`feat: run bootstrap init end to end`
