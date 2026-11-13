# U6 — Review: the engine and the test tasks

## Scope

The review covers U1, U3, U4 and U5: the branch delta of `cli/`,
`.config/mise/tasks/test/` and `macos/fixtures/`. The change lands runnable
code (Go and bash), so the plan has one review unit (D11). The data of U2 is
judged by the wave review.

## Ruling

Review against D1-D9, D12 and D14-D22 of `index.json` and `docs/design/cli.md`.
Look for: a command that writes or syncs more than one time; an error that
prints instead of being returned by `Plan` or `Apply`; a restore that misses a
new values file; the frozen copy read as the current packs; a tool literal in
the engine; a test case that asserts nothing.

## Verification

- `mise x -- mise run test` passes.

## Commit

None: a review unit commits nothing.
