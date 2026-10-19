# U6 — Docs: reconcile the docs with the gap fixes

- **Wave:** 3
- **Depends on:** U1, U5
- **Owns:** `README.md`, `CLAUDE.md`, `docs/sync/index.md`,
  `docs/sync/migrate.md`, and any other doc that `vwf:docs-sync` finds
  falsified, except the files that U1, U2, U3 and U4 own
- **Model:** opus
- **Kind:** edit
- **Read first:** every owned file you edit, top to bottom; the `DOCS
  FALSIFIED:` lines of U1–U4, which the orchestrator passes in.
- **Lazy-load:** `docs/sync/vwf.md`, `docs/sync/kotlin.md`,
  `docs/sync/android.md`, `docs/sync/swift.md` (read only, to quote them
  correctly).

## Ruling

From index.md, assumed decisions:

> R1 — `remove tools` also deletes the `Template` outputs that the tool's
> `docs/sync/<tool>.md` lists, but only a file that carries the
> `Synced from bootstrap` header in its first three lines.

> R4 — The init script sets `LockMode.STRICT` on each project. The normal run
> resolves without `--write-locks` …

> R14 — `xcodebuild -version` runs only for a root with an `.xcodeproj`. …

> R15 — On a non-Darwin host with a Swift root, the task … exits 0.

## Edits

1. Run `vwf:docs-sync` over the run's branch delta (the skill file is
   `/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`).
   Apply its findings that fall inside this unit's Owns, plus every
   `DOCS FALSIFIED:` line from U1–U4.
2. **`README.md`**, **`CLAUDE.md`**, **`docs/sync/index.md`**: change only a
   passage that describes `remove tools`, the lock check of the kotlin task,
   the Xcode or macOS requirement of the swift task, or the Kotlin-DSL-only
   detect of the android tasks. Add nothing that only repeats the per-tool
   files.
3. **`docs/sync/migrate.md`**: no change, unless docs-sync finds a passage
   that the rulings falsify. The rulings ask no action of a consumer.
4. Do not write a decisions doc: the plan changes earlier plan rulings, not
   standing decisions (index.md, Goal).

## Verification

- Wave-gate lines 1 and 2 pass.
- No passage in the owned files says that `remove tools` keeps every
  `Template` output, or that the swift task fails on a Linux host.

## Guardrails

- Do not touch `tools/**` or the files of U1–U4. If one of them is wrong,
  return an `UNRESOLVED:` line that names the file and the passage.
- Do not edit the two `gaps_of:` folders or the synced copies at the
  repository root.
- Delete with `rm`, never `git rm`.

## Commit

`docs: reconcile the docs with the mobile gap fixes`
