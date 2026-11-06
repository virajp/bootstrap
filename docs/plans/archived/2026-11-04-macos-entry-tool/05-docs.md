# U5 — Docs

- **Read first:** every owned file that the list below names, before editing.
- **Lazy-load:** the diffs of U1, U2 and U3 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

From the design rulings of 2026-11-04, verbatim:

> - Tests: `test:repo-sync`, `test:setup-sync`, `test:macos-tools` (renamed
>   from macos-setup `test:tools`), and an orchestrator `test`
>   (`test/_default`) that runs all three in parallel. CLAUDE.md rule:
>   `sync-engine/` → `mise run test`; `repo/` → `test:repo-sync`; `macos/` →
>   `test:setup-sync` + `test:macos-tools`. No test in hooks or in
>   `code:merge` (a shared task; consumers have no such tests).

> - Contract docs: one place for each rule. `docs/design/` holds the full
>   contract: new `sync-engine.md`, `repo-sync.md`, `setup-sync.md` (from
>   macos-setup), and `packs.md`, `tools.md`, `mise.md` change to the merged
>   model. More files are fine, but CLAUDE.md and `docs/sync/index.md` load
>   them lazily ("read X when you change Y"). CLAUDE.md keeps layout, change
>   steps, task rules, test rule, pointers. Script headers keep usage, flags,
>   a pointer (the 300-line setup-sync header contract goes).
>   `macos-setup/CLAUDE.md` becomes short (macos.toml, how to edit, how to
>   sync).

Plan decision D11: "No. mise run setup:configs in bootstrap and in
consumers, and the first Mac run, stay hand steps (Out of scope)."

## Edits

1. **The explicit list.**
   - **`CLAUDE.md`**: the test rule for `macos/` names both
     `mise run test:setup-sync` and `mise run test:macos-tools`; the layout
     section names `macos/tools/`, `macos/packs.toml`,
     `macos/examples/macos.toml` (the sample values file, the model of
     `test:macos-tools`) and the task `test:macos-tools`. One sentence: a
     change to a Mac tool also changes `macos/examples/macos.toml` when the
     tool needs a value or a selection, because `test:macos-tools` fails for
     an unselected tool.
   - **`README.md`**: the Tests section names `test:macos-tools`.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units, inside the owned files. Do not commit.
3. **`docs/sync/migrate.md`**: no item. No repo consumer changes anything.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'test:macos-tools' CLAUDE.md README.md` finds a line in each.
- `grep -n 'macos/examples/macos.toml' CLAUDE.md` finds a line.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Write identifiers in code spans; write ASD-STE100.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the macos-setup tool and test:macos-tools` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
