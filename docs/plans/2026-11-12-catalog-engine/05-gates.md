# U5 — Gates

- **Read first:** `index.json` `gates`.

## Ruling

The plan has no generator to run. The release level of `bootstrap` is MINOR,
which `/vwf:execute` records at landing; this unit bumps no version.

## Edits

1. Run the wave gate over the finished tree, and report each result. No file
   changes are expected.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Change no file. Bump no version.

## Commit

`ops: none` — this unit makes no commit when no file changes.
