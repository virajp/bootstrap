# U4 — Docs

- **Read first:** every owned file that the list below names, at the named
  passages, before editing.
- **Lazy-load:** the diffs of U1 and U2 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

D4: "The user, by hand, after the landing. It is out of scope, as in the
earlier plans."

D5: "Nothing. The first sync has no manifest and removes nothing.
docs/sync/migrate.md lists the old leftovers, and the owner deletes them by
hand as before."

D9: "It stays as it is. It is a rule of the branch model and works also
without a manifest. The stale rule replaces only _rs_stale and the removal in
_rs_skip_required."

## Edits

1. **The explicit list.** Find each passage by its text.
   - **`CLAUDE.md`**: "How the sync operates", the bullet "The sync never
     deletes a file, except ...": the exceptions are now the merge tasks of
     the other branch model, a synced file at a path where the sync needs a
     folder, and the stale rule: a file with the header that
     `.config/repo-generated.tsv` lists and this run does not write. Add a
     bullet for the manifest: its format, committed, one for each submodule,
     written only after a clean run, the first sync removes nothing. "mise
     config": the bullets about the removal of a stale `conf.d` file and of a
     `files` copy (after P03, `requires`/`requires_found` of a `[files]`
     table) now point to the stale rule.
   - **`README.md`**: the bullet "The sync never deletes a file, with four
     exceptions" says the same as `CLAUDE.md`.
   - **`docs/design/mise.md`** and **`docs/design/tools.md`**: each passage
     about stale conf.d files or the removal of a skipped file.
   - **`docs/sync/tools/repo-sync.md`**: the steps of the sync get the stale
     rule and the manifest; the files list of the consumer gets
     `.config/repo-generated.tsv` (commit it).
   - **`docs/sync/index.md`**: where the adoption commits the synced files,
     add `.config/repo-generated.tsv`.
   - **`docs/sync/migrate.md`**: add the next free `M<n>` item and its row in
     the "Find the items that apply" table: the first sync writes
     `.config/repo-generated.tsv`; commit it; from then on the sync removes
     stale synced files by itself; the first sync removes nothing, so delete
     old leftovers by hand as before (D5).
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units that the list does not cover, inside the owned files.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -l 'repo-generated.tsv'` finds `CLAUDE.md`, `README.md`,
  `docs/sync/tools/repo-sync.md`, `docs/sync/index.md` and
  `docs/sync/migrate.md`.
- `grep -n '_rs_stale' CLAUDE.md README.md docs` finds nothing.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Write identifiers in code spans.
- Each new paragraph obeys the style of the file it is in.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the repo manifest and the stale rule` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
