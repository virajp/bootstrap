# Plan — The Mac tasks in Python (2026-11-18)

## Goal

After this plan lands, each Mac task of `macos/tools/*/tasks/` is a Python uv
script with a `.py` name, and the shell Mac helpers are gone. The Mac gets
the Python tasks at its next Mac sync (`mise run updateall:configs`); the
stale rule removes the old shell copies. The task names, flags, output and
exit codes do not change. With this plan, B30 is done: every mise task of
bootstrap is Python, except the engine scripts that the gap record
`docs/memory/gaps/2026-11-18-engine-scripts-python-parked.md` parks.

This is plan 3 of 3 of B30. It requires
`docs/plans/2026-11-18-python-tasks-repo`. No reversal of its own.

## Facts the survey established

- **The Mac task files** (27 files, about 1000 lines):
  - mise: `helpers` (83 lines: `print_header`, `print_subheader`,
    `print_wait`, `print_newline`, `print_error`, `print_success`,
    `print_warn`, `line_sep`, `task_exists`, `run_task_if_exists`; colors and
    the emojis ✅ and 🚫), `updateall`, `updateall-configs` (the entry of the
    Mac sync), `upgrade-macos-download`, `upgrade-macos-upgrade`,
    `upgrade-power`. Keys: `macos/tools/mise/tools.toml` lines 71-88.
  - claude-cli: `func-claude` (ends with `exec claude`), `upgrade-ai`
    (`#MISE tools={jq = "latest"}`). cli: `cleanupDS`, `func-backup`,
    `func-code`, `func-listStartupItems`, `func-zip`, `keychain-delete`.
    homebrew: `brew-casks`, `upgrade-brew`. network: `ips-gateway`, `ips-ii`,
    `ips-ips`, `ips-ipv4`, `ips-ipv6`, `ips-remoteip`. node: `upgrade-npm`.
    pstree: `func-myps` (ends with `exec pstree`). tailscale:
    `tailscale-daemon`, `upgrade-tailscale`. telegram: `cleanupTelegram`.
  - Each Mac task sources `${HOME}/.config/mise/tasks/_scripts/helpers`.
- **The Mac dprint config** (`macos/tools/dprint/tools.toml:60`) sends every
  file below `**/.config/mise/tasks/**` to shuck, except `*.tmpl`, and it
  has no ruff plugin (D13).
- **The entry task** `updateall-configs` reads its flags from `"$@"`, not
  from `usage_*`, and calls `${BOOTSTRAP_BIN:-bootstrap}` (D11).
- **uv on the Mac.** The Mac python tool is required, the Mac uv tool is
  optional (D10). The foundation plan removed `UV_NO_CACHE` from
  `macos/tools/uv/tools.toml`.
- **The Go cases:** `e2e/mac/commands_test.go` runs the entry task
  `updateall:configs` with `bash <entryTask>` (`entryRun`, about line 368)
  and with `mise run updateall:configs` (`entryMise`);
  `e2e/mac/sync_test.go` and `e2e/mac/mac_test.go` expect the synced file
  lists; `e2e/macostools/macostools_test.go` syncs each Mac tool with
  `macos/examples/macos.toml`.
- **The docs:** `docs/setup/tools.md` (one section for each Mac tool; about
  line 721 names the shell helpers), `docs/design/mac.md` 67-68, CLAUDE.md
  (the Mac tools and "Tests").

- **`mise run` and a task flag.** `mise run <task> --all` gives the task
  `usage_all=true`, the same as `mise run <task> -- --all` (tested
  2026-11-18). The gate lines need no `--`.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file — the file a package
manager or a toolchain manager writes to pin what it resolved — with Write,
Edit, `sed` or a script: a lock file changes only as the output of the CLI
that owns it, the command the stack's package-manager conventions name, and
the unit keeps what that CLI wrote. A unit whose manifest change moves a lock
file lists that lock file in its `owns` beside the manifest.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    QUESTION Q<n>: <the question, readable alone>
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence> [| <option 3> | <option 4>]
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

The three QUESTION lines repeat per question, `n` counting from 1 in the
block, with two to four options; a block with no question carries the one
line `QUESTION: none`. A question is a point the plan does not decide and the
unit cannot settle inside its scope; the unit never proceeds on an
assumption.

A unit completes every edit that does not depend on an answer, never builds on
a part an answer can change, then returns all its open questions in one block.
Its edits stay uncommitted in the worktree.

## Out of scope

- The engine scripts stay bash (foundation D2).
- A change to the code of `cli/` (foundation D14).
- The Mac sync of the user's Mac: the user runs `mise run updateall:configs`
  by hand (D7).
- A change of behavior of a task while it moves.

## Parked

- The other consumer repositories run `mise run setup:configs` by hand, and
  the Mac runs `mise run updateall:configs` by hand, after this plan lands.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-18-python-tasks-mac

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
