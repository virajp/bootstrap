# U4 — Review: the dispatchers, the init move and the tests

## Scope

U1 to U3, the whole branch delta since the branch base. The task scripts run in
every shaped repo, and the engine change moves files in a user's repo (D9).
Focus: the `NN-` filter of `own_tasks` does not drop a part that a shipped tool
names (D4); a dispatcher with no part exits 0 (D2); `init` never loses a file,
and its restore puts a moved file back when the sync fails (D4, D7, D8); a
manifest-listed path is never moved.
