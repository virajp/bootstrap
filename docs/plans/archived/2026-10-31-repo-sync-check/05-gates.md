# U5 — Gates

- **Read first:** nothing; this unit edits no file.

## Ruling

The plan names no generator. The unit runs the full wave gate over the
finished tree and reports the result.

## Edits

1. **Generators:** none.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `git status --short` is empty.

## Guardrails

- Edit no file. Do not bump a version.

## Commit

None: the unit changes no file, so the orchestrator makes no commit for it.
