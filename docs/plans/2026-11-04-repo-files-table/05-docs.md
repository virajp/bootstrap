# U5 — Docs

- **Read first:** every passage that the list below names, before editing.
- **Lazy-load:** the diffs of U1 to U3 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned, the vwf
  documentation-standards skill.

## Ruling

From the design rulings of 2026-11-04:

> - Seeds on both sides use the setup-sync form: `[files."<target>"]` (target
>   path, no `.tmpl`), `mode = "seed"`, `requires = [...]` (exact paths
>   relative to the target root, each must exist). `seed = [...]` and
>   `[seed_requires]` go (exit 2 naming the new key). No `[files]` table =
>   `mode = "generated"`. The repo mise-entry `requires` (a file name searched
>   4 folders deep) is a second meaning of the word: solve it in the "mise
>   config" row.

> - `requires` has one meaning: exact paths, in `[files]` tables. New key
>   `requires_found` = a list of file names searched down to 4 folders below
>   the root, valid in a mise entry and in a `[files]` table, repo side only
>   (Mac → exit 2). The `files` list of a mise entry goes: each file states
>   its condition in its own `[files]` table (python, node).

From this plan:

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires plan 2 only."

D4: "No. mise run setup:configs in bootstrap and in each consumer stays a hand
step (Out of scope), as in earlier plans."

D14: "No. A merge source carries no two-line bootstrap header, because the
app owns the target. A YAML or TOML merge source may carry the comment #
Merged by bootstrap: <path>; a JSON merge source carries no comment. The
engine strips nothing. The test fixtures follow this. The docs unit rewrites
the header rule of CLAUDE.md to: each file that is neither a seed nor a merge
source."

D6: "No. tools.toml is internal to bootstrap; a consumer changes nothing, and
its synced files stay the same."

## Edits

1. **The explicit list.** Find each passage by its text.
   - **`CLAUDE.md`**:
     - "Layout rule": `tools.toml` has four keys (`default_envs`,
       `depends`, `root_only`, `extensions`); a seed is a
       `[files."<target>"]` table with `mode = "seed"`; the bullet on
       `[seed_requires]` becomes `requires` of that table; add the
       `[files]` keys (`mode`, `requires`, `requires_found`, `format`,
       `remove`, `remove_items`) and merge mode in one short bullet each.
     - The header rule ("Each file that is not a seed starts with this
       header"): it becomes "each file that is neither a seed nor a merge
       source", and one sentence on the optional `# Merged by bootstrap:
       <path>` comment of a YAML or TOML merge source (D14).
     - "mise config": the optional `requires` of an entry becomes
       `requires_found`, a list; the `files` bullet becomes "each file
       states its condition in its own `[files]` table".
     - "How the sync operates": the deletion bullet (a synced file of a
       `files` list) now names a `[files]` target whose condition fails.
   - **`docs/design/tools.md`** and **`docs/design/mise.md`**: the keys of
     `tools.toml`, the `[files]` table, merge mode, `requires_found`.
   - **`docs/sync/tools/*.md`**: each tool with a seed or a condition (bun,
     docker, fnox, git, gitleaks, grype, kotlin, mise, node, osv-scanner,
     python, swift) names the new keys where it names the old ones.
   - **`docs/sync/index.md`**: each passage on `seed` or `seed_requires`.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`).
   Apply each finding and each `DOCS FALSIFIED:` line of the earlier units
   that the list does not cover, inside the owned files. Add no
   `docs/sync/migrate.md` item (D6).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rn 'seed_requires\|seed = \[' CLAUDE.md README.md docs/sync
  docs/design`
  finds only `docs/sync/migrate.md` items of the past.
- `grep -ln 'requires_found' CLAUDE.md docs/design/tools.md
  docs/sync/tools/python.md`
  lists all three files.
- `mise x -- dprint check CLAUDE.md README.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Do not edit `docs/plans/**`. Do not change a past item of
  `docs/sync/migrate.md`.
- Write identifiers in code spans.
- Each new paragraph obeys the style of the file it is in (ASD-STE100, short
  sentences).
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the [files] tables, requires_found and merge mode` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
