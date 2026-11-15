# U6 — Gates

- **Read first:** `index.json` of this folder, `gates` and `release_levels`.

## Ruling

The plan names no generator. No unit bumps `cli/VERSION`: the release
workflow writes it when the user tags `bootstrap@4.0.0` (R1b D16).

## Edits

1. Run `mise x -- mise run code:lint` and `mise x -- mise run test:go`.
2. Run `mise x -- mise run test` and `mise x -- mise run code:sec`.
3. Run K1 of `gates.keeps`.

## Verification

- Every command above passes.

## Guardrails

- Edit no file. A failure is a question, never a fix in this unit.

## Commit

No commit when no file changes.
