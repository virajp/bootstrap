# U9 — Gates

- **Read first:** `index.json` of this folder, `gates` and `release_levels`.

## Ruling

The plan names no generator. No unit bumps `cli/VERSION`: the release
workflow writes it when the user tags `bootstrap@4.0.0` once R1a, R1b, R1c
and R2 all land (D92).

### Rulings review, round 2

These rulings override any earlier text of this file that they contradict.

- **D40** — bootstrap@4.0.0 is tagged after R1a, R1b, R1c and R2 (D36).
- **D59** — U9 builds with -o to a temp path.

### Rulings review, round 3

These rulings override any earlier text of this file that they contradict.

- **D92** (supersedes D69) — bootstrap@4.0.0 is tagged once, after R1a, R1b,
  R1c and R2 all land (R1b D16); until then the user tests with a local binary.
  No tag after R1a; R1b, R1c and R2 are not 4.x minors. D36 and D40 hold again.
  Whether R2 keeps alias task names is for the R2 interview.

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
