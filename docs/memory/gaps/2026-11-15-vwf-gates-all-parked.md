GAP 2026-11-15 ★3 | plan:2026-11-15-gate-file-sets, repo:claude-plugins,
plugin:vwf | the vwf gate lines of a consumer must use --all | parked: a
change in claude-plugins

**Source** The rulings review of `docs/plans/2026-11-15-gate-file-sets/` (D55).
The user approved this park.

What the plan decides: after it lands, `code:format` and `code:lint` examine
only the uncommitted files (staged, changed and new) when they get no files and
no `--all`. With `--all`, they examine every file of the repository.

What is open: `/vwf:execute` runs the wave gate and the plan gate of a plan
(for example `mise x -- mise run code:lint`) after the units commit. On a clean
worktree the default set is empty, so such a gate examines nothing. The vwf
gate ports and the default gate lines of `/vwf:plan` and `/vwf:change-plan`
(`gates-consent.md`, the plan template) must use `--all` for `code:lint` and
`code:format`. The plans that are already queued keep their gate lines; their
gates still pass, but examine fewer files.
