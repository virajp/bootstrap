# U3 — Review: U1 and U2

## Scope

This review covers U1 and U2, the whole branch delta since the branch base.
U1 adds `sync-engine/setup-sync`, which writes in `$HOME` and runs
`mise bootstrap`. U2 adds the task scripts `test:setup-sync` and `test`. All
of them run, which is the reason for this unit (D12: "Yes. U1 adds
sync-engine/setup-sync and U2 adds two task scripts, and all of them run. One
review unit in wave 3 covers U1 and U2.").
