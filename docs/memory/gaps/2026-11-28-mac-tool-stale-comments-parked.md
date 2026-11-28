GAP 2026-11-28 ★2 | repo:bootstrap, tree:macos/tools,
plan:2026-11-28-mac-tools-doc-audit | stale comments in the shipped files of
the Mac tools | parked: a small change plan after the docs audit of B18 lands

**Source** The planning of B18 (the audit of `docs/setup/tools.md` against
`macos/tools/`), 2026-11-28. The user ruled that the audit is a docs plan, and
that the stale comments in the tool files are a separate item, because they
change shipped files.

The comments that the survey of 2026-11-28 found:

1. `macos/tools/homebrew/files/.config/fish/conf.d/05-homebrew.fish:16` says
   that mise activates at "51-". The mise drop-in is `06-mise.fish`.
2. `macos/tools/fish/tools.toml:3`, `macos/tools/zsh/tools.toml:3` and
   `macos/tools/starship/tools.toml:3` say that `cli` adds drop-ins. The
   `zoxide` and `dasel` tools add them now.
3. `macos/tools/claude/files/.claude/CLAUDE.md.merge.toml:44-46` names the
   conf.d files `cli.toml` and `devtools.toml`. There is no `devtools` tool.
4. `macos/tools/claude-cli/tasks/func-claude.py:52` says "the claude tool's".
5. `macos/tools/mise/tasks/updateall.py:145-146` says that `files` examines
   `sudo_local` again, but the command is `--only macos-defaults,defaults`.

The audit units can find more. The audit files of the plan,
`docs/plans/2026-11-28-mac-tools-doc-audit/audit/*.md`, list them in their
"Stale comments" sections, and the change plan of this item takes them too.
