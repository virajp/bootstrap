# U3 — Docs

- **Read first:** `CLAUDE.md`, `docs/sync/index.md`, `README.md`, and the
  three new files of `docs/setup/`.
- **Lazy-load:** the diffs of U1 and U2 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that they returned.

## Ruling

From the design rulings of 2026-11-04 (verbatim):

"Contract docs: one place for each rule. [...] More files are fine, but
CLAUDE.md and `docs/sync/index.md` load them lazily ("read X when you change
Y"). CLAUDE.md keeps layout, change steps, task rules, test rule, pointers.
[...]"

"Placement (accepted): [...] `docs/` stays at the root: `docs/sync/` (no
move; plugins name `docs/sync/vwf.md`), `docs/setup/` (Mac guide: sync.md,
tool-reference.md), `docs/design/` (both sides). [...]"

The placement bullet above is older: D7 names the three files.

From this plan:

D7: "docs/setup/index.md, docs/setup/migrate-dotfiles.md and
docs/setup/tools.md. The later "Mac docs" ruling wins over the older
placement bullet, which names sync.md and tool-reference.md."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-docs-design-split."

## Edits

1. **`CLAUDE.md`** — the Layout rule names `docs/setup/` and its three
   files. The list `## Where the rules are` (P15) gets the lines: the Mac
   adoption → `docs/setup/index.md`; the migration from `dotfiles/` →
   `docs/setup/migrate-dotfiles.md`; a Mac tool → `docs/setup/tools.md`.
   Change a shared file says: a change to a tool in `macos/tools/` changes
   its section in `docs/setup/tools.md` too.
2. **`docs/sync/index.md`** — one pointer line: a Mac uses
   `docs/setup/index.md`, not this prompt.
3. **`README.md`** — one line under the adoption section that points a Mac
   owner to `docs/setup/index.md`.
4. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line inside the owned
   files.
5. **No `docs/sync/migrate.md` item.** No repo consumer file changes.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.
- `grep -c 'docs/setup/' CLAUDE.md` is at least 3;
  `grep -n 'docs/setup/index.md' docs/sync/index.md README.md` finds a line
  in each file.

## Guardrails

- Touch only the owned files. Do not edit `docs/setup/`; report a finding
  there as a `GAP:`.
- Write new prose in ASD-STE100 Simplified Technical English. Write
  identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: point CLAUDE.md, README.md and docs/sync to the Mac docs` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
