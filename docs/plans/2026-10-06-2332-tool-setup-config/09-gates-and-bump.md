# U9 — Gates (no version bump)

- **Wave:** 7
- **Depends on:** U8
- **Owns:** — (no release: no version file, no generator)
- **Model:** opus
- **Kind:** edit

## Ruling

The consent block records "Release cli publicly: none" — the user's answer "No
release yet". No version bump, no publish. This unit runs the final gate only.

## Edits

none.

## Verification

- `pnpm install --frozen-lockfile` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:build` exits 0.
- `MISE_ENV=dev mise run code:all` exits 0.

## Guardrails

- Do not change `cli/package.json`'s version.
- Do not write `implementation:` stamps — that is the executor's Reconcile step.

## Commit

none — this unit changes no file.
