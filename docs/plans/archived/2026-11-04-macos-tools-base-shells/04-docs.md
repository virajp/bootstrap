# U4 — Docs

- **Read first:** every owned file that `vwf:docs-sync` names, before editing.
- **Lazy-load:** the diffs of U1 and U2 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

From the design rulings of 2026-11-04, verbatim:

> - Contract docs: one place for each rule. `docs/design/` holds the full
>   contract: new `sync-engine.md`, `repo-sync.md`, `setup-sync.md` (from
>   macos-setup), and `packs.md`, `tools.md`, `mise.md` change to the merged
>   model. More files are fine, but CLAUDE.md and `docs/sync/index.md` load
>   them lazily ("read X when you change Y"). CLAUDE.md keeps layout, change
>   steps, task rules, test rule, pointers. Script headers keep usage, flags,
>   a pointer (the 300-line setup-sync header contract goes).
>   `macos-setup/CLAUDE.md` becomes short (macos.toml, how to edit, how to
>   sync).

Plan decision D10: "No. mise run setup:configs in bootstrap and in
consumers, and the first Mac run, stay hand steps (Out of scope)."

## Edits

1. **The explicit list:** none. This plan adds tool folders, pack members and
   sample answers; the per-tool Mac guide `docs/setup/tools.md` comes in P16.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units, inside the owned files. A doc that counts the Mac tools, lists the
   members of pack base, or describes the fish `login_shell` value gets the
   new text. Do not commit.
3. **`docs/sync/migrate.md`**: no item. No repo consumer changes anything.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `git diff --stat` of this unit names only owned files.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Write identifiers in code spans; write ASD-STE100.
- Delete with `rm`, never `git rm`.

## Commit

`docs: update the docs for the shell and identity Mac tools` — written by the
orchestrator after the unit gate and the wave gate, not by the unit. When the
unit changes no file, the orchestrator makes no commit for it.
