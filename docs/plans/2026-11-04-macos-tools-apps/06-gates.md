# U6 — Gates

- **Read first:** nothing; this unit edits no file.

## Ruling

The plan names no generator. The unit runs every gate line of the plan over
the finished tree and reports the result.

## Edits

1. **Generators:** none.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run code:precommit` passes.
- `mise x -- mise run test` passes: `test:repo-sync`, `test:setup-sync` and
  `test:macos-tools`, with all 92 Mac tools in the render of
  `test:macos-tools`.
- `command ls macos/tools | wc -l` prints 92.
- `git status --short` is empty.

## Guardrails

- Edit no file. Do not bump a version.

## Commit

None: the unit changes no file, so the orchestrator makes no commit for it.
