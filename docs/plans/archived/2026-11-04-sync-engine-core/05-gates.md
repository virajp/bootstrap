# U5 — Gates

- **Read first:** nothing; this unit edits no file.

## Ruling

The plan names no generator. The unit runs every gate line of the plan over
the finished tree and reports the result.

## Edits

1. **Generators:** none.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run code:precommit` passes.
- `mise x -- mise run test:repo-sync` passes.
- `git status --short` is empty.

## Guardrails

- Edit no file. Do not bump a version.
- A red gate is a failure of this unit: report the gate and its output.

## Commit

None: the unit changes no file, so the orchestrator makes no commit for it.
