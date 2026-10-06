# U3 — Apply engine: plan, backups, atomic write, rollback

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `cli/src/apply/**`, `cli/test/apply/**`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/apply/apply.test.ts` — against a stub file system:
  (a) the plan marks each target `create`, `backup` (existing, different) or
  `unchanged` (byte-identical, no backup); (b) a target that is a directory or a
  symlink, or whose parent directory is a symlink or a regular file, fails the
  whole plan before any write with every such path named; (c) backups are
  `<name>.bak`, then the lowest free `<name>.N.bak`, never overwriting a backup;
  (d) a write failure on the k-th file rolls back: every created file and
  created directory removed, every backup restored, tree byte-identical, error
  names the failing path; (e) an interrupt during the write phase rolls back the
  same way; (f) a rollback that itself fails returns the list of
  `{path, backup}` not restored; (g) the setup-config file is written last.
  Fails because the module does not exist.
- **Read first:** flow 110 steps 6–7 and Guarantees,
  `docs/blueprint/conventions.md` `#backups` and `#baseline` (atomic
  multi-write).
- **Lazy-load:** Effect finalizer / interruption and file-system service docs
  via Context7.

## Ruling

| # | Decision               | Ruling                                                                                                                                                                                                                                                                                                                     | Rejected                                                                    | Unit   |
| - | ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- | ------ |
| 4 | Interrupt and rollback | rollback is an Effect finalizer on the write phase; `NodeRuntime.runMain` turns Ctrl-C into a graceful interrupt; the command maps an interrupt to exit 130                                                                                                                                                                | a hand-written signal handler                                               | U3, U6 |
| 5 | Fault injection in e2e | "Real faults + test pause": a write failure is a real one (a target directory made read-only); an interrupt uses one hidden env variable `BOOTSTRAP_TEST_PAUSE_AFTER_WRITE=<n>` that pauses after the n-th write so the test sends SIGINT; it does nothing else, is undocumented, and is read only at the process boundary | real faults only (timer SIGINT, flaky); interrupt tested at unit level only | U3, U6 |
| 6 | Rollback scope         | rollback removes every file and every directory the run created and restores every backup, so the tree is byte-identical to before                                                                                                                                                                                         | remove files only                                                           | U3     |

## Edits

1. **`cli/src/apply/plan.ts`** — input: the rendered map (path → content) and
   the repository root; output: the sorted per-path plan, or the typed "not a
   regular file" failure listing every offending path (checked before any
   write).
2. **`cli/src/apply/backup.ts`** — backup-name choice per `#backups`.
3. **`cli/src/apply/write.ts`** — executes a plan all-or-nothing: backups moved
   aside, files and directories written, a final file (the setup config) written
   last; a finalizer rolls back on failure or interruption; a failed rollback
   returns the unrestored list. Accepts an optional "pause after n writes" hook
   as a plain parameter (U6 wires it from the env variable at the process
   boundary; this module never reads the environment).
4. **`cli/test/apply/**`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- No `node:fs`; use the platform file-system service so tests use a stub.
- This module never reads env variables or prints anything.
- Do not touch `cli/src/repo/**`, `cli/src/report/**`, `cli/src/commands/**`.

## Commit

`feat: plan and apply file writes with backups and rollback`
