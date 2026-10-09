# U5 — Atomic writer with rollback

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/src/write/**` (new), `cli/test/write/**` (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/write/apply.test.ts` (in-process, a temporary
  directory, a FileSystem layer wrapped by a test seam):
  - A change set (create with missing parent directories, change, delete) is
    applied; every path under `.config/mise/tasks/` gets mode `0755`; the setup
    config `.config/bootstrap.yaml` is written last of all writes.
  - The seam fails the n-th write: every changed target is back to its content
    and mode before the run, every created file is deleted, and the error has
    exit 3 naming the failing path.
  - The seam fails the n-th write and also fails one restore: the error has exit
    3 and lists every path not restored.
  - The seam interrupts the fiber at the n-th write: the tree is as before, and
    the result is the interrupted error, exit 130, "interrupted — repository
    restored".
  - An interrupt before the first write gives exit 130 "interrupted — nothing
    written".
  - It fails because `cli/src/write/` does not exist.
- **Read first:** `docs/blueprint/conventions.md` `#safety`, `#errors` (the
  interrupt and restore lines), `docs/blueprint/entities/setup-config/index.md`
  (invariant 1), `cli/src/report/errors.ts`.
- **Lazy-load:** the Effect `FileSystem`, `Effect.onExit` and interruption docs
  via Context7.

## Ruling

| # | Decision                                         | Ruling                                                                                                                                                                                   | Rejected                                   | Unit   |
| - | ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ | ------ |
| 1 | Restore after a write failure or an interrupt    | In-memory snapshot of the content and mode of each target before the write, written back; created files deleted. It equals HEAD because every written target is clean (user, 2026-10-09) | `git restore --source=HEAD`                | U5     |
| 2 | Fault tests (write failure, interrupt mid-write) | In-process FileSystem seam that fails or stops at the n-th write; no test code ships in the binary (user, 2026-10-09)                                                                    | Environment variable in the binary; timing | U5, U8 |
| 6 | Rollback trigger                                 | One `onExit` handler covers a write failure and an interrupt                                                                                                                             | Two separate paths                         | U5     |
| 9 | Plan 1 row 9 (setup-config invariant 1)          | Lands here: a dirty target, `.config/bootstrap.yaml` included, refuses the run only when its render differs from the working copy; the setup config is written last                      | —                                          | U5, U6 |

## Edits

1. **`cli/src/write/apply.ts`** — take a change set (path, content, mode,
   action) with the setup config as its last entry; snapshot each existing
   target; write; on any failure or interrupt, restore the snapshots and delete
   the created files (row 6), and return the matching error from
   `cli/src/report/errors.ts`.
2. **`cli/test/write/**`** — the test-first cases and the FileSystem seam (test
   code only).

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- The writer reads the FileSystem service from the Effect context only, so the
  test seam can wrap it; no `node:fs`.
- The writer decides nothing about which files change; it applies a set.
- No git call during the write or the rollback (row 1).

## Commit

`feat: write the change set atomically with rollback on failure or interrupt`
