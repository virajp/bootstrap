# U5 — Docs

- **Read first:** every passage that the list below names, before editing.
- **Lazy-load:** the diffs of U1 to U3 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned, the vwf
  documentation-standards skill.

## Ruling

From the design rulings of 2026-11-04:

> - One merged packs model for both sides (user's proposal, 2026-11-04):
>   - `[packs.<name>]` adds each `default = true` member (option A). A new
>     `default = true` member reaches every consumer of the pack; a
>     `default = false` member arrives only with `[tools.<name>]`. Still open,
>     as in `docs/design/packs.md`: no way to drop one default member.
>   - `required` goes: a common tool (`_security_base`) becomes a `depends` of
>     each member tool (gitleaks, grype, osv-scanner), so a single-tool
>     selection is correct. Repo consumers get a `migrate.md` item for it.

From this plan:

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires plan 3 only."

D4: "No. mise run setup:configs in bootstrap and in each consumer stays a hand
step (Out of scope), as in earlier plans."

D10: "The next free number after the highest item at the time of the run
(plan 1 adds one before this plan)."

## Edits

1. **The explicit list.** Find each passage by its text.
   - **`docs/design/packs.md`**: the whole model. A pack is a group of
     members; the member forms `<tool> = { default = true }` and
     `<tool> = {}`; a nested pack `"pack:<name>"`; `[packs.<name>]` adds the
     default members; a tool in two packs comes once; no pack-level keys;
     the old keys and `p:` stop with code 2; an unknown member is code 3.
     Replace "Common tools and required tools" (:135-150): a common tool is
     a `depends` of each member tool. Keep the open point: no way to drop
     one default member. Change the examples (:23-29, :118-130) to the new
     shape.
   - **`docs/sync/packs/base.md`** and **`docs/sync/packs/security.md`**: the
     member lists with their `default` flags; `security.md` says that each
     scanner brings `_security_base`.
   - **`docs/sync/tools/gitleaks.md`**, **`grype.md`**, **`osv-scanner.md`**:
     the Depends row is `_security_base`.
   - **`CLAUDE.md`**, "Layout rule": the pack bullet (`depends`, `p:<name>`,
     `required`) becomes the member shape; the bullet on a common tool says
     that each member tool depends on it.
   - **`docs/sync/index.md`** and **`docs/sync/vwf.md`**: each passage on
     `depends`, `required` or `p:` of a pack.
   - **`docs/sync/migrate.md`**: the new item (D10) and its row in "Find the
     items that apply": "A values file that selects one scanner, or a
     repository that read `required`". Text: `required` is gone; a single
     scanner now brings `_security_base`; values files do not change; the
     consumer does nothing.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`).
   Apply each finding and each `DOCS FALSIFIED:` line of the earlier units
   that the list does not cover, inside the owned files.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rnE 'required = |"p:|p:security' CLAUDE.md docs/design docs/sync/packs
  docs/sync/index.md docs/sync/vwf.md`
  finds nothing.
- `grep -n 'default = true' docs/design/packs.md docs/sync/packs/base.md`
  finds the new shape in both files.
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

`docs: describe the merged packs model` — written by the orchestrator after
the unit gate and the wave gate, not by the unit.
