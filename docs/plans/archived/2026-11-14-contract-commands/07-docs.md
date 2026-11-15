# U7 — Docs

- **Read first:** the branch delta, then each passage named below.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. The engine rulings are quoted in the unit
files of U1 to U4.

- **D16**, a reversal of R1a D69, in the user's words: "Keep the release to
  4.0.0 since we will be doing all these changes and only then releasing", and
  for the scope: "all, we will be building a local binary and testing it". So
  R1a, R1b, R1c and R2 land before one `bootstrap@4.0.0` tag; until then the
  user tests with a local binary. It restores R1a D36.
- **D6**, a reversal of R1a D23 and D52: one fetch rule (D5) for `config set`,
  `config unset`, `catalog` and `describe`.
- **D2:** `questions --json` prints the envelope; every doc that shows the bare
  array changes.

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings plus every
   `DOCS FALSIFIED:` line of the earlier units.
2. **`docs/design/cli.md`** — the command and flag tables (`contract`,
   `describe`, `catalog`, `--json` on `check`, `config`, `questions` and the
   six write commands); a new section "The JSON contract" (the envelope D1,
   the version rule D7, the items D11, the exit code D12, and the `data` of
   each command: D4, D8, D9, D10, D13, D14, D19); "Consent" (no prompt under
   `--json`, D4); "Questions" (D2); the clone cache text (D5, D6); "The exit
   codes" (code 2 for `--json` with no `--yes` or `--dry-run`, code 5 for a
   stale clone). Fix the stale points the survey found: `list` still describes
   packs, `detect --no` is gone, and `init [tool…]` now refuses names.
3. **`docs/sync/vwf.md`** — the plugin contract names `bootstrap contract` as
   the version check, `describe` and `catalog` as the reads, and `--json` with
   `--yes` or `--dry-run` for the writes. It shows the envelope once.
4. **`docs/sync/index.md`, `docs/setup/index.md`,
   `docs/setup/migrate-dotfiles.md`** — each passage that shows the bare
   `questions --json` array (D2).
5. **`README.md`, `CLAUDE.md`** — the command list and each mention of
   `questions --json`.
6. **`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`** — add a
   section "Amendments 2026-11-14 (R1b planning)" that records D16 (one
   `bootstrap@4.0.0` tag after R1a, R1b, R1c and R2; local-binary tests until
   then; it supersedes R1a D69), D1 (the envelope), D7 (the version rule) and
   D10 (no LSP field until R2). Fix the phase table and every 4.1, 4.2 or 4.3
   reference to match D16. Do not rewrite the rest of the record.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rn '4\.1\.0\|4\.2\.0\|4\.3\.0' README.md CLAUDE.md docs/design
  docs/sync docs/setup docs/memory/decisions` finds nothing that names a
  bootstrap release of the split.

## Guardrails

- Touch nothing outside the owned paths. `docs/plans/` is not this unit's.
- Delete with `rm`, never `git rm`.

## Commit

`docs: the JSON contract commands and one 4.0.0 tag`
