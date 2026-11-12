# U5 — Review: the engine change and the test tasks

## Scope

The review covers U3 and U4: the branch delta of `cli/`,
`macos/fixtures/` and `.config/mise/tasks/test/` since the branch base. The
change lands runnable code (the Go engine and the bash test tasks), so the plan
has one review unit (D6). U1 and U2 are data and templates: the wave review
judges them.

## Ruling

Review against D4, D5 and D8 of `index.json`, the rules of plan 1a, and
`docs/design/cli.md`. Look for: a message that does not have the form of D5, a
tool literal in the engine, a test case that asserts nothing, a fixture that
lost a case, and a template of D4 that writes an empty `email =` line.

## Verification

- `mise x -- mise run test` passes.

## Commit

None: a review unit commits nothing. Its findings go to the units that own the
files.
