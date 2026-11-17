# U6 — Docs

- **Read first:** the branch delta, then each passage named below.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. The rulings are quoted in the unit files of
U1 to U3.

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings plus every
   `DOCS FALSIFIED:` line of the earlier units.
2. **`docs/design/tools.md`, `docs/design/repo.md`** — contributed merge items
   (D1, D7).
3. **`docs/sync/tools/claude-tool.md`** (D12; the pnpx text at about lines
   42-44) — setup:ai: install missing, the record,
   removal, `--update` (D2 to D5); no pnpx.
4. **`docs/sync/tools/git.md`** — the vwf lines are gone from `.gitignore`.
5. **`docs/sync/tools/virajp-plugins.md`** (new) — the tool (D6).
6. **`docs/sync/vwf.md`** — a vwf repo adds `virajp-plugins` (`bootstrap add
   tool virajp-plugins`); plugin instructions edit the claude tool's
   `plugins.*` values; removal is setup:ai's (D3), not a hand step.
7. **`docs/sync/migrate.md`** — a migrate item: a repo that relied on the pnpx
   install or on the vwf `.gitignore` lines adds `virajp-plugins`.
8. **`README.md`, `CLAUDE.md`** — the tool list and each mention of the pnpx
   install.
9. **`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`** — add to
   the amendments section: R2c's D1 (contributed merge items), D3 (the record)
   and D4 (update only on a flag). Do not rewrite the rest of the record.

## Verification

- `mise x -- mise run code:format` passes.

## Guardrails

- Touch nothing outside the owned paths. `docs/plans/` is not this unit's.
- Delete with `rm`, never `git rm`.

## Commit

`docs: contributed merge items, virajp-plugins and a declarative setup:ai`
