# U7 — Gates

- **Read first:** `index.json` of this folder, `gates` and `release_levels`.

## Ruling

The plan names no generator. No unit bumps `cli/VERSION`.

## Edits

1. Run `mise x -- mise run code:lint`.
2. Run `mise x -- mise run test` and `mise x -- mise run code:sec`.
3. Run K1 of `gates.keeps`. D43: K1 runs the probe gate directly, so no other
   gate runs.

## Verification

- Every command above passes.

## Guardrails

- Edit no file. A failure is a question, never a fix in this unit.

## Commit

No commit when no file changes.
