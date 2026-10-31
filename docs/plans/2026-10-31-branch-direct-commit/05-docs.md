# U5 — Docs

- **Read first:** every owned file that the list below names, at the named
  lines, before editing.
- **Lazy-load:** the diffs of U1, U2 and U3 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

D11: "All four parts: (1) the explicit list of passages in 05-docs.md, each
with its change; (2) then vwf:docs-sync over the full branch delta, with a fix
for each passage it finds that the list does not name; (3) a grep check: no
doc names the old hook text \"main or develop\", and each listed file that
describes the key has a hit for direct_commit; dprint check over the docs;
(4) the examples stay a separate unit (U3). The unit also adds one optional
M<n> item to docs/sync/migrate.md."

D2, D3, D4, D5, D6 and D7 are the behavior to describe; quote their rule, not
their wording, where a doc gives a rule.

## Edits

1. **The explicit list.** Line numbers are from the survey; find each passage
   by its text if the lines moved.
   - **`docs/design/bootstrap-config.md`**: add `direct_commit` to the
     example (`:33-39`) as a commented line, to the summary table
     (`:61-63`), and to the mention at `:208`. Add a key section for
     `direct_commit` after the `[branch.*]` key sections (`:273-283`, before
     "Reserved names" at `:285`). It gives the table of the five cases of D2
     (one core branch `main`; one core branch `develop`; both with `develop`
     only; both with both; both with `main` only, code 2), the code 2 errors
     of D4, D5 and D13, the effect of D7 and D14 on a hook in
     `.config/repo-sync/pre-commit.yaml`, and the D15 rule: the hook protects
     only the core branches that exist and do not allow direct commits.
   - **`docs/sync/tools/git.md`**: the merge config (`:44-47`) says that the
     merge tasks do not read `direct_commit`; the hook (`:55-56`) gets the new
     name of D6 and says that `direct_commit` removes a branch. git.md has no
     safety-net passage today: add one sentence to the merge-config passage
     (`:44-47`): before a merge, the merge tasks run the hooks over all files
     as a safety net, with `SKIP=no-commit-to-branch` (D3).
   - **`docs/sync/tools/pre-commit.md`** (`:22`, `:60`) and
     **`docs/sync/tools/repo-sync.md`**: the rendered hook args follow
     `direct_commit`, and `repo-sync` validates the key.
   - **`docs/sync/index.md`**: `:373`, the key table (`:407-409`), the
     example at `:439`, and step 4 (`:518-545`), which tells the adopting AI
     about the branch tables.
   - **`docs/sync/vwf.md`** (`:109`) and **`docs/design/packs.md`** (`:155`).
   - **`README.md`** (`:143-144`, `:246`).
   - **`CLAUDE.md`**: "Branch model in the merge tasks" (the key, the rule of
     D2, the safety-net skip of D3), and the hook line of "Change a shared
     file" (this repository sets no `direct_commit`, so the hook still refuses
     a commit on `main`).
   - **`docs/sync/migrate.md`**: one optional `M<n>` item, the next number,
     and its row in the "Find the items that apply" table: to allow direct
     commits on a core branch, set `direct_commit = true` in its table, then
     run `mise run setup:configs`. No old form exists, so the item says that
     a repository that does not want direct commits does nothing.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units that the list does not cover, inside the owned files.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rn --exclude-dir=plans 'direct commit to main or develop' README.md
  CLAUDE.md docs`
  finds nothing.
- `grep -ln 'direct_commit'` finds each of: `docs/design/bootstrap-config.md`,
  `docs/sync/tools/git.md`, `docs/sync/index.md`, `docs/sync/migrate.md`,
  `CLAUDE.md`.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Do not touch the examples (U3) or `examples/last_sync.toml` (U6).
- Do not edit a synced copy at the root; CLAUDE.md and README.md here are
  this repository's own files.
- Each new paragraph obeys the style of the file it is in.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe direct_commit and the safety-net skip` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
