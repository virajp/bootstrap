GAP 2026-11-14 ★3 | plan:2026-11-14-plugins-setup-ai, skill:init,
plugin:stackgen | vwf repos must add the virajp-plugins tool | parked: R3 in
claude-plugins

**Source** The R2c interview of `docs/plans/2026-11-14-plugins-setup-ai/`.

From R2c on, bootstrap's `setup:ai` no longer installs vwf through
`pnpx @virajp.dev/claude-plugins`, and the git tool no longer writes vwf's
`.gitignore` lines. A new `virajp-plugins` tool carries both: it contributes
the vwf `.gitignore` lines to the git tool's `.gitignore`, and the
`vwf@virajp-plugins` plugin entry to the claude tool's `plugins.user`.
`setup:ai` also removes, at project scope, the plugins it installed when they
leave the list.

What claude-plugins must change in R3:

- `/vwf:init` adds the tool: `bootstrap add tool virajp-plugins`.
- The stackgen handoff
  (`stackgen-stack-template/references/bootstrap-handoff.md`)
  edits the claude tool's `plugins.*` values and stops telling the agent to
  uninstall a removed plugin by hand.
- claude-plugins' `CLAUDE.md`, `readme.md` and `installer/CLAUDE.md` stop
  describing the pnpx install inside `setup:ai`.
