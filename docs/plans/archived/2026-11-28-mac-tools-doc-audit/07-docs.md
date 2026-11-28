# U7 — Docs

- **Read first:** the branch delta of the run (`git diff <base>...HEAD`), the
  `DOCS FALSIFIED:` lines of U1 to U6, and the facts section of `index.md`.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`,
  `docs/design/mac.md`, `docs/setup/index.md`, `CLAUDE.md`.

## Ruling

> D1: Audit docs/setup/tools.md against the current macos/tools/. The Notes of
> the deleted macos-setup docs/tool-reference.md (lines 71-201 at commit
> a4ec825^) are a checklist, not a source: each claim that is still true must
> be in the doc, and a false claim is dropped.

> D3: Each fact goes in the ### section of the tool that owns it. A rule for
> many tools goes in 'How a tool works', as a paragraph before the item legend
> ('Each section below gives these items for one tool:').

> D9: No unit changes a file of macos/tools/. An audit unit lists each stale
> comment that it finds in the '## Stale comments' section of its audit file.
> U6 does not apply them. The parked item
> docs/memory/gaps/2026-11-28-mac-tool-stale-comments-parked.md takes them,
> and its later change plan reads the audit files.

## Edits

1. Run `vwf:docs-sync` over the branch delta of the run. Apply its findings
   and each `DOCS FALSIFIED:` line of U1 to U6 to the doc that it names.
   `docs/setup/tools.md` is U6's: change it only for a finding of docs-sync
   that U6 did not cover.

## Verification

- `mise x -- mise run code:format --all` passes.
- `mise x -- mise run code:lint --all` passes.
- No doc outside `docs/setup/tools.md` describes a Mac tool fact that U6
  changed in a different way (grep the tool names of the changed sections in
  `docs/design/mac.md`, `docs/setup/index.md`, `docs/setup/migrate-dotfiles.md`
  and `CLAUDE.md`).

## Guardrails

- Do not edit a file of `macos/tools/`, `repo/` or `cli/`.
- Do not edit the audit files.
- Delete with `rm`, never `git rm`.

## Commit

`docs: reconcile the docs with the audit of the Mac tools`
