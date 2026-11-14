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
2. **`docs/design/tools.md`** — the `[lsp.<server>]` keys (D4) and the D7
   rules.
3. **`docs/design/cli.md`** — the `lsp` command (D4, D6), the `lsp` field of
   `describe`, and `lsp` in `contract`'s commands.
4. **`docs/design/mise.md`, `docs/sync/tools/mise.md`** — the
   `code:lsp:check` task (D5).
5. **`docs/sync/tools/{go,node,python,kotlin,swift,flutter,rust,java}.md`** —
   each tool's server and its dev pin (D1 to D3, D8).
6. **`docs/sync/vwf.md`** — a plugin reads the LSP servers with
   `bootstrap lsp --json` (or `describe`'s `lsp` field) and checks them with
   `code:lsp:check`.
7. **`README.md`, `CLAUDE.md`** — the command list and the task list.
8. **`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`** — add to
   the amendments section: R2b's D1 (every language tool), D4 (Claude-ready
   output of `bootstrap lsp`) and D5 (what `code:lsp:check` tests). Do not
   rewrite the rest of the record.

## Verification

- `mise x -- mise run code:format` passes.

## Guardrails

- Touch nothing outside the owned paths. `docs/plans/` is not this unit's.
- Delete with `rm`, never `git rm`.

## Commit

`docs: the LSP blocks, bootstrap lsp and code:lsp:check`
