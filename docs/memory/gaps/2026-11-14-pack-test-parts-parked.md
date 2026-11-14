GAP 2026-11-14 ★3 | plugin:stackgen, plan:2026-11-14-task-ports | stackgen packs
copy test/e2e and test/golden as files, where bootstrap now ships dispatchers |
parked: the packs must ship NN- parts

**Source** The R1c interview of `docs/plans/2026-11-14-task-ports/` (D4).

From R1c on, bootstrap's mise tool ships `test/_default`, `test/e2e/_default`
and `test/golden/_default` as group dispatchers, which run `NN-<name>` parts
only. In claude-plugins, these stackgen packs still copy a file to a
dispatcher path from `config/.config/mise/tasks/`:

- `test/golden`: `app-framework/compose` and one more pack;
- `test/e2e`: `framework/android`.

`bootstrap init` moves such a file to `<group>/50-local`, and `check` reports
it with code 2 until then. So a repo shaped with those packs works, but each
new materialize writes the file again.

Where it belongs: a claude-plugins plan (R3, or R5 when the packs go) changes
each pack to ship `test/e2e/NN-<pack>` or `test/golden/NN-<pack>` parts.
`p/_project/deploy` (three packs) is not affected: R1c ships no
`p:<id>:deploy` (D5).
