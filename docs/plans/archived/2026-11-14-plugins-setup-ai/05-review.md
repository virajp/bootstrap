# U5 — Review: the engine, the tool, setup:ai and the tests

## Scope

U1 to U4, the whole branch delta since the branch base. The plan lands engine
Go code and a task that installs and uninstalls Claude plugins (D10). Focus:
contributed items never change a merge-mode file that has none (D1); setup:ai
never uninstalls a plugin that it did not record, never touches user scope,
and changes nothing on a second run (D3, D4, D18, D19); the files that R2c
touches name no vwf (D21).
