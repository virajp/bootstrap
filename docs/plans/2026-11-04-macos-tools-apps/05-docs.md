# U5 — Docs

- **Read first:** the diffs of U1, U2 and U3 (`git log -p` of the branch),
  and the `DOCS FALSIFIED:` lines that the earlier units returned.
- **Lazy-load:** each owned doc that the docs-sync survey names.

## Ruling

The plan names no doc passage of its own. The docs unit runs `vwf:docs-sync`
over the branch delta and applies its findings plus every `DOCS FALSIFIED:`
line.

From the design rulings of 2026-11-04 (verbatim): "Contract docs: one place for
each rule. [...] CLAUDE.md keeps layout, change steps, task rules, test
rule, pointers."

## Edits

1. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line inside the owned
   files.
2. **Lists of Mac tools and packs.** When an owned doc lists the Mac tools
   or the Mac packs (for example a table that P10 to P13 added), complete
   it with the 52 tools and the packs `apple-dev`, `iwork`, `ms-office`,
   `disk`, `media` and `notes`. A count of Mac tools in an owned doc becomes 92.
3. **No `docs/sync/migrate.md` item.** This plan changes no file that a
   repo consumer gets.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Write identifiers in code spans.
- Write new prose in ASD-STE100 Simplified Technical English and in the
  style of the file.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the app-only Mac tools and the full Mac packs` — written by
the orchestrator after the unit gate and the wave gate, not by the unit.
When the unit changes no file, the orchestrator makes no commit.
