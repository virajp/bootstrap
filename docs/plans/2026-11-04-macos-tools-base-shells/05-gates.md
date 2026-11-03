# U5 — Gates

- **Read first:** nothing; this unit edits no file.

## Ruling

The plan names no generator. The unit runs every gate line of the plan over
the finished tree and reports the result. `mise x -- mise run test` runs
`test:macos-tools`, which now renders the init stage (exit 4, then 0) and
each answer of the example.

## Edits

1. **Generators:** none.
2. Run `mise x -- mise run code:precommit` and `mise x -- mise run test`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run code:precommit` passes.
- `mise x -- mise run test` passes.
- `git status --short` is empty.

## Guardrails

- Edit no file. Do not bump a version.

## Commit

None: the unit changes no file, so the orchestrator makes no commit for it.
