# U4 — Docs

- **Read first:** `docs/design/tools.md` (the key table, the questions),
  `docs/design/packs.md`, `docs/design/cli.md` ("The command line"), and each
  other owned file at the passages that the facts section of `index.md` names.
- **Lazy-load:** `CLAUDE.md` ("Where the rules are").

## Ruling

The decisions D1–D8, D14 and D15 of `index.json`, as the code of U1 implements
them.

## Edits

1. Run `vwf:docs-sync` over the branch delta of the run. Apply its findings,
   each `DOCS FALSIFIED:` line of U1 and U2, and the docs of the facts section
   of `index.md`.
2. **`docs/design/tools.md`** — add `summary`, `category`, `selectable`,
   `hosts` and the repo `description` to the key table, with their types and
   rules (D1, D4, D5). Add a section on the categories file (D2). Add `group`,
   `note` and `optional` to the questions, with the rules of D6 (rules 1–6).
   Say that `summary` and `category` become necessary in a later change.
3. **`docs/design/packs.md`** — the reserved key `description` (D3).
4. **`docs/design/cli.md`** — the refusals of `init`, `add` and `replace`
   (D4, D5), the new keys of `questions --json` (D6), the warning of `check`
   (D5), the list of `list` (D5), the empty
   `--answer <q>=`, the prompt with the pre-filled field (D6 rules 4 and 5,
   D14), and the new Go dependencies Bubble Tea v2 and Bubbles v2 where the
   doc names the dependencies of the engine.
5. **`docs/design/mac.md`** — the keys that a Mac tool accepts now, and that
   `hosts` is for repo tools only.
6. **`docs/sync/index.md`, `docs/setup/tools.md`, `docs/setup/index.md`,
   `README.md`, `CLAUDE.md`** — only where a passage names the keys of a tool,
   a pack or a question, or the rule that each question needs an answer.

Write the docs in ASD-STE100 Simplified Technical English, in the style of the
existing text.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'selectable' docs/design/tools.md docs/design/cli.md` finds each
  file.

## Guardrails

- Do not touch `cli/`, the test tasks, `repo/` or `macos/`.
- Do not edit a synced copy (see `CLAUDE.md`).

## Commit

`docs: describe the catalog keys of a tool, a pack and a question`
