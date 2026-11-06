# U4 — Review: U1 to U3

## Scope

This review covers U1, U2 and U3, the whole branch delta since the branch
base. U1 changes `sync-engine/select` and `sync-engine/repo-sync`, U2
rewrites `tools/packs.toml` and three scanner manifests, and U3 changes the
test task. Each consumer runs this code, so the plan lands runnable code
(D1).

D1: "Yes. U1 changes the selection code that each consumer runs, U2 changes
the data that it reads, and U3 changes the test task. The plan lands runnable
code, so one review unit in wave 3 covers U1, U2 and U3."
