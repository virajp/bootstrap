# U9 — Gates

- **Read first:** `index.json` of this folder, `gates` and `release_levels`.

## Ruling

The plan names no generator. No unit bumps `cli/VERSION`: the release
workflow writes it when the user tags `bootstrap@4.0.0` at once after this
plan lands (D69).

### Rulings review, round 2

These rulings override any earlier text of this file that they contradict.

- **D40** — bootstrap@4.0.0 is tagged after R1a, R1b, R1c and R2 (D36).
- **D59** — U9 builds with -o to a temp path.

### Rulings review, round 3

These rulings override any earlier text of this file that they contradict.

- **D69** — bootstrap@4.0.0 is tagged at once after R1a lands, as the selection
  model does for 3.0.0. R1b, R1c and R2 become minor releases (4.1, 4.2, 4.3);
  R2 keeps the old task names as aliases for one release, as the design says, so
  it does not break. R3 waits for the release that carries the contract. D36 and
  D40 are superseded on the tag.

## Edits

1. Run `mise x -- mise run code:lint` and `mise x -- mise run test:go`.
2. Run `mise x -- mise run test` and `mise x -- mise run code:sec`.
3. Build the engine once, to prove it builds, with its output in a temp
   path: `go build -o "$(mktemp -d)/bootstrap" ./cmd/bootstrap` in `cli/` (D59).

## Verification

- Every command above passes.

## Guardrails

- Edit no file. A failure is a question, never a fix in this unit.

## Commit

No commit when no file changes.
