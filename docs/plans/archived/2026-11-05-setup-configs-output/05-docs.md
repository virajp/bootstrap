# U5 — Docs

- **Read first:** the skill file `vwf:docs-sync`, then the passages below.

## Ruling

Run `vwf:docs-sync` over the branch delta. Apply its findings, each
`DOCS FALSIFIED:` line of the earlier units, and the list below. D2 and D5 give
the behaviour to describe.

## Edits

1. **`docs/sync/tools/repo-sync.md`** — Lines 22-23, 48-49, 52-74, 84-85, 171,
   196 and 208: describe the list, the table, `--verbose`, `SYNC_ENGINE_VERBOSE`
   and the copy of the local sync task. State that a repository with an old
   `_default` must run `mise run setup:configs` once without `--verbose`,
   because
   mise refuses a flag that the old task does not define.
2. **`README.md`** — Lines 392-398, where they describe the output.
3. **`CLAUDE.md`** — Lines 214-222, 368 and 439-447: add the output rules
   to "How the sync operates". State that the test tasks set
   `SYNC_ENGINE_VERBOSE=1`.
4. **`docs/sync/migrate.md`:330 and `docs/sync/index.md`:744** — These passages
   tell an agent to read `kept seed` or `unchanged` lines. Tell the agent to use
   `mise run setup:configs --verbose`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rn 'SYNC_ENGINE_VERBOSE' CLAUDE.md docs/sync/tools/repo-sync.md` shows
  a hit in each file.

## Guardrails

- Change only the passages that the change falsifies.
- Write in STE100.

## Commit

`docs: describe the changed-only output of the sync`
