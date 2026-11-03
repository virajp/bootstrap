# U4 — Review: U1 to U3

## Scope

This review covers U1, U2 and U3, the whole branch delta since the branch
base. U1 adds `sync-engine/clone` and changes the three tasks of the
`repo-sync` tool, U2 moves the library to `sync-engine/repo-sync` and adds the
forwarding file, and U3 changes the test task. Each consumer runs this code,
so the plan lands runnable code (D1).

D1: "Yes. U1 and U2 change bash code that each consumer runs, and U3 changes
the test task. The plan lands runnable code, so one review unit in wave 3
covers U1, U2 and U3."
