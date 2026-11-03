# U4 — Review: U1 to U3

## Scope

This review covers U1, U2 and U3, the whole branch delta since the branch
base. U1 changes `sync-engine/repo-sync` and `sync-engine/write` and adds
`sync-engine/merge`, U2 moves every `tools.toml` to the `[files]` tables, and
U3 changes the test task. Each consumer runs this code, so the plan lands
runnable code (D1).

D1: "Yes. U1 changes the engine that each consumer runs and adds
sync-engine/merge, U2 changes the manifests that the engine reads, and U3
changes the test task. The plan lands runnable code, so one review unit in
wave 3 covers U1, U2 and U3."
