# U7 — Review: the TUI

## Scope

The review covers U1 to U6: the branch delta of `cli/`. The change lands
runnable code (Go), so the plan has one review unit.

## Ruling

Review against D1-D7 and D10-D14 of `index.json`, `screens.txt` of this folder
and
`docs/design/cli.md`. Look for: an import of `cli/internal/cli` from
`cli/internal/tui`; a string literal of `cli/internal/tui` that names a tool; a
write of a file from the TUI other than through `core.Apply`; a change list
that a cancel, a failed sync or a shrink of the terminal loses; a key of
`screens.txt` with no effect; a modal that a key cannot close; a new module
other than Lip Gloss.

## Verification

- `mise x -- mise run test` passes.

## Commit

None: a review unit commits nothing.
