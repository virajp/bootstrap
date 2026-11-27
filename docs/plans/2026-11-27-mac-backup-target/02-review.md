# U2 — Review: the root key of the Mac sync

## Scope

U1, the engine change and its tests: the branch delta since the branch base.
D7: the change lands runnable engine code in `cli/internal/`, so this review
runs `/code-review` and `/security-review` and the two reviewers. Pay
attention to the path handling of a `backup` target and of a `@backup/`
manifest row: a target must never resolve outside the clone or `$HOME`.
