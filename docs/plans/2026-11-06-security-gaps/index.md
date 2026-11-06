# Plan — Rulings and fixes for the open security gaps of the bootstrap 2.0 run (2026-11-06)

## Goal

Each open security gap of the bootstrap 2.0 run has one ruling: fix, or accept
with a reason. The fixes land in `macos/tools/` and `sync-engine/setup-sync`.
The plan `docs/plans/2026-11-06-dotfiles-security` of the `macos-setup`
repository lands the same values in the live `dotfiles/`, so that the user can
test them before the migration.

The user brought this request. No backlog item asked for it. This plan has no
reversal of a standing decision (D12).

## Facts the survey established

Line numbers are from commit `0ccae6e`.

- **Trust path:** `macos/tools/mise/.config/zsh/01b-mise.zsh.tmpl`:6,
  `bash/01b-mise.bash.tmpl`:10 and `fish/conf.d/06-mise.fish.tmpl`:9 hold
  `$HOME/.config/mise:<% $pd %>`. Line 3 of each file expands the `~` of
  `projects_dir` into `$pd`.
  `macos/tools/claude-cli/.claude/settings.json.tmpl`:55
  holds `<% .env.HOME %>/.config/mise:<% $projects %>`. Lines 1-2 expand the `~`
  into `$projects`.
- **mise:** The docs state that `[settings]` of an untrusted config are ignored,
  and that `MISE_TRUSTED_CONFIG_PATHS` trusts the configs under its paths with
  no prompt.
- **uv:** The docs state that `uv run` creates the project environment if it
  does not exist. `python.uv_venv_auto` has the values `source` and
  `create|source`, and no value that only creates.
- **`repo/tools/python/tools.toml`:21** keeps the setting, and
  `docs/sync/tools/python.md`:57 stays true.
- **Starship:** `macos/tools/starship/.config/starship.toml`:147 is the `when`
  command of `custom.submodule_status`. `PROJECTS_DIR` comes from
  `macos/tools/cli/tools.toml`:29. A scratch test with starship 1.27.0 shows
  that the `git_status` module runs the filter driver of a repository, the same
  as `git status`.
- **brew:** `macos/tools/homebrew/tools.toml`:25 and
  `macos/tools/macos-setup/.config/mise/tasks/updateall/_default`:69. A scratch
  test of `mise hook-env -s` shows `alias foo='echo $HOME/x'` in zsh and bash,
  and `alias foo 'echo $HOME/x'` in fish.
- **Context7:** `settings.json.tmpl`:52 holds `CONTEXT7_RUNNER`, and `:76-77`
  hold the allow lines that stay. `macos/tools/claude-cli/tools.toml`:16 holds
  `remove`. `macos/tools/fnox/.config/fnox/config.toml`:29-32 holds the profile,
  and `macos/tools/fnox/tools.toml`:2 names it in a comment. No test and no
  fixture names `context7` or `uv_venv`.
- **`~/.ssh`:** `_ss_file` of `sync-engine/setup-sync` starts at :518, and
  `mkdir -p` of the file writer is at :572. `_se_mode` is in
  `sync-engine/common`:21. `_se_finish` of `sync-engine/write` has the lines
  `would chmod <mode> <show>` and `chmod <mode>  <show>`, which D6 copies. The
  live folder of this Mac already has mode 700.
- **Tailscale:** `macos/tools/tailscale/.config/mise/tasks/upgrade/tailscale`:44
  holds the `readlink -l` line. `readlink -l` gives `illegal option` on this
  Mac.
- **Gates:** `test:setup-sync` covers `sync-engine/setup-sync` with
  `macos/fixtures/`. `test:macos-tools` renders each Mac tool. `code:precommit`
  already runs all files and has no flag.
- **Gap sources:** The open security entries are in the `gaps` lists of the nine
  `docs/plans/2026-11-04-*` folders that `04-docs.md` names. "The bootstrap 2.0
  run" means all plan folders dated 2026-11-04. Each folder validates today.
  `2026-11-04-macos-entry-tool` has G5 and G6 closed already.
- **Commit convention:** `.config/git-conventional-commits.yaml` allows `fix`,
  `docs` and the others. No scope is necessary.
- **Dependencies:** This plan adds none.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its owned
paths plus "touch nothing outside this list", the facts section, the plan's
shared files, and the return block below. A unit never bumps a version, never
runs a generator, never edits a doc that it does not own, never adds a
dependency, never commits. A unit deletes with plain `rm`, never `git rm`. A
unit never writes a lock file by hand.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

A `GAP:` is a hole in the plan that the unit can pass on a stated assumption. It
is recorded, and the run continues. An `UNRESOLVED:` is a ruling that the unit
cannot continue without. It blocks the unit and its dependents.

## Out of scope

- The non-security gaps of the same folders. The request names only the security
  gaps.
- `repo/tools/python/tools.toml`. A repository that adopts the tool keeps the
  setting (D2).
- The `macos-setup` repository. Its own plan owns it (D11).
- `mcp__plugin_context7_context7__*` allow lines. The official plugin uses them
  (D5).

## Parked

- The fixes reach the rendered Mac files through this repository only at the
  migration in `docs/setup/migrate-dotfiles.md`.
- The `CONTEXT7_API_KEY` item stays in the Keychain of the user. The user can
  delete it by hand.
- `stackgen` in `claude-plugins` still reads `CONTEXT7_RUNNER`, with the
  fallback
  `pnpm dlx`. Nothing breaks.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-06-security-gaps

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
