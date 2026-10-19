# U5 — Docs: reconcile the docs with the new mobile tasks

- **Wave:** 3
- **Depends on:** U4
- **Owns:** `README.md`, `CLAUDE.md`, `docs/sync/index.md`,
  `docs/sync/migrate.md`, and any other doc that `vwf:docs-sync` finds
  falsified, except the files that U1, U2 and U3 own
- **Model:** opus
- **Kind:** edit
- **Read first:** every owned file you edit, top to bottom; the `DOCS
  FALSIFIED:` lines of U1, U2 and U3, which the orchestrator passes in.
- **Lazy-load:** `docs/sync/kotlin.md`, `docs/sync/swift.md`,
  `docs/sync/android.md` (read only, to quote them correctly).

## Ruling

From index.md, assumed decisions:

> E10 — Each task unit owns its `docs/sync/<tool>.md` (the "Tasks" table and
> "What the tool ships"), as D12 of plan 1.

> E1 — One task `setup:deps:<tool>` per tool, with a `--upgrade` flag, as
> `setup:deps:flutter`.

## Edits

1. Run `vwf:docs-sync` over the run's branch delta (the skill file is
   `/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`).
   Apply its findings that fall inside this unit's Owns, plus every
   `DOCS FALSIFIED:` line from U1, U2 and U3.
2. **`docs/sync/index.md`**, **`README.md`**, **`CLAUDE.md`**: change only a
   passage that lists the tasks of a tool or says that the kotlin, swift or
   android tool ships no task. Add nothing that only repeats the per-tool
   files.
3. **`docs/sync/migrate.md`**: if it has a list of old stackgen pack files to
   delete, add the pack tasks that the new tasks replace
   (`setup/deps/{install,audit,cleanup,outdated,upgrade}/{kotlin,swift,swiftui}`,
   `setup/deps/install/android`, `code/lint/android`); otherwise do not
   change it.

## Verification

- Wave-gate lines 1 and 2 pass.
- No passage in the owned files says that the kotlin, swift or android tool
  ships no task.

## Guardrails

- Do not touch `tools/**` or the `docs/sync/<tool>.md` files of U1–U3. If one
  of them is wrong, return an `UNRESOLVED:` line that names the file and the
  passage.
- Do not edit the synced copies at the repository root.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the tasks of the kotlin, swift and android tools`
