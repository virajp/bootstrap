GAP 2026-11-16 ★3 | plan:2026-11-14-task-ports, plan:2026-11-14-git-tasks |
remove the fallback for an engine older than 4.0.0 from three scripts |
resolved 2026-11-18: the three fallback blocks are removed before the tag,
because the user runs no sync before bootstrap@4.0.0

**Source** The user approved this fix after the `/vwf:execute all` run stopped
at the preflight of `docs/plans/2026-11-14-git-tasks/`. Gap G2 of
`docs/plans/2026-11-14-task-ports/` records the cause.

The cause: the config-folder merge (`37d6f13`) made three scripts call
`bootstrap config get`, which only bootstrap 4.0.0 has. The released engine is
3.3.1, and the gate `code:lint:60-bootstrap` runs the scripts of the shared
clone of `main`. Thus the gate failed in this repository and in each consumer
with engine 3.3.1.

The fix: each script has a fallback. When `BOOTSTRAP_VALUES` names a file, or
when it is empty and only the old `.config/bootstrap.toml` exists, the script
uses its logic from before 4.0.0 (`8abecac`):

- `repo/tools/git/scripts/derive`
- `repo/tools/git/scripts/check`
- `repo/tools/mise/scripts/derive`

What is open: after the release of bootstrap@4.0.0, remove the three fallback
blocks. Each block starts with the comment "The fallback for an engine older
than 4.0.0".
