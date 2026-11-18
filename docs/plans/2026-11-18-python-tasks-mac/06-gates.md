# U6 — Gates

- **Read first:** `index.json` of this folder, `gates` and `release_levels`.

## Ruling

The plan names no generator. No unit bumps `cli/VERSION`, and the release
level is NONE (foundation D14).

## Edits

1. Run `mise x -- mise run code:lint --all`.
2. Run each line of `gates.plan`.
3. Run K1 of `gates.keeps`.

## Verification

- Every command above passes.

## Guardrails

- Edit no file. A failure is a question, never a fix in this unit.

## Commit

No commit when no file changes.
