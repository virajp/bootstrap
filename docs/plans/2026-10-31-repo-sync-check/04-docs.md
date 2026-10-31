# U4 — Docs

- **Read first:** every owned file that the list below names, at the named
  lines, before editing.
- **Lazy-load:** the diffs of U1 and U2 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

D10: "The method of the B13 plan: (1) the explicit list of passages in
04-docs.md, each with its change; (2) then vwf:docs-sync over the full branch
delta, with a fix for each passage it finds that the list does not name; (3)
grep checks: each listed file that describes the check has a hit for --check
or code:lint:repo-sync; dprint check over the docs. It adds M23 to
docs/sync/migrate.md: after the sync, run mise run code:lint once and fix each
error that code:lint:repo-sync reports."

D3: "Finish B14. The seven rules that need the clone stay with setup:configs
[...]."

D4: "A gate task code:lint:repo-sync in the repo-sync tool. [...] so no new
hook is added. [...]"

D6: "--check runs each local check, prints each message, and returns the
highest code. The sync does not change [...]."

## Edits

1. **The explicit list.** Line numbers are from the survey; find each passage
   by its text if the lines moved.
   - **`docs/sync/tools/repo-sync.md`**: the Tasks section (`:33-56`) gets
     `code:lint:repo-sync`, with the gate rules of D4; the steps (`:56-80`)
     say that `--check` stops after the local rules; the env vars
     (`:80-89`) say that `--check` reads no `REPO_SYNC_CLONE`; Pre-commit
     hooks (`:91-93`) says that the existing `lint` hook runs the gate
     through `code:lint`; the table "What the tool ships" (`:17-26`) lists
     the new task.
   - **`docs/design/mise.md`**: the task table row of `repo-sync` (`:348`)
     adds `code:lint:repo-sync`.
   - **`docs/design/packs.md`**: the row of `repo-sync` (`:154`), which lists
     what the tool ships, adds `code:lint:repo-sync`.
   - **`docs/design/bootstrap-config.md`**: near the key table (`:43-66`),
     one paragraph: `code:lint:repo-sync` examines the local rules at each
     commit, and the rules that need the clone (D3, a short list) only at
     `setup:configs`.
   - **`docs/sync/index.md`**: next to the dry-run passages (`:218`, `:358`,
     `:594-597`), say that the adopting AI can run
     `mise run code:lint:repo-sync` to validate the values file without a
     sync.
   - **`docs/sync/migrate.md`**: add `M23` after `M22` (`:487`) and its row in
     the "Find the items that apply" table: after the sync, run
     `mise run code:lint` once, and fix each error that `code:lint:repo-sync`
     reports. A repository with a valid values file does nothing.
   - **`README.md`**: "How a repository syncs" (`:130-163`) gets one sentence
     about the commit-time check; Tests (`:294-299`) mentions the `--check`
     cases of `test:repo-sync`.
   - **`CLAUDE.md`**: "How the sync operates" gets a bullet for `--check`
     (local rules only, every error at once, no clone, no write; the clone
     rules stay with the sync); the exit-codes bullet notes that `--check`
     returns 2 at most; the `test:repo-sync` bullet mentions the `--check`
     mode.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units that the list does not cover, inside the owned files. Do not commit;
   the orchestrator commits the unit.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -lE 'code:lint:repo-sync|--check'` finds each of:
  `docs/sync/tools/repo-sync.md`, `docs/design/bootstrap-config.md`,
  `docs/sync/index.md`, `docs/sync/migrate.md`, `docs/design/mise.md`,
  `docs/design/packs.md`,
  `README.md`, `CLAUDE.md`.
- `grep -n '^## M23' docs/sync/migrate.md` finds the new item.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Write identifiers in code spans: the linter reads `_a ... _b` in prose as
  emphasis and deletes a space.
- Each new paragraph obeys the style of the file it is in.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe repo_sync --check and the code:lint:repo-sync gate` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
