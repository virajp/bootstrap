# U6 — Gates

- **Read first:** `index.json` of this folder, `gates` and `release_levels`.

## Ruling

The plan names no generator. No unit bumps `cli/VERSION`: the release
workflow writes it when the user tags `bootstrap@4.0.0` after R1a, R1b, R1c
and R2 all land (R1b D16).

## Edits

1. Run `mise x -- mise run code:lint` and `mise x -- mise run test:go`.
2. Run `mise x -- mise run test` and `mise x -- mise run code:sec`.
3. Build the engine once, to prove it builds, with its output in a temp path:
   `go build -o "$(mktemp -d)/bootstrap" ./cmd/bootstrap` in `cli/`.

## Verification

- Every command above passes.

## Guardrails

- Edit no file. A failure is a question, never a fix in this unit.

## Commit

No commit when no file changes.
