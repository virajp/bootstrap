# Plan — Shared mise generator (2026-11-04)

## Goal

After this plan lands, the mise config generator, its env rules and the
`[overrides.mise*]` rules are one shared library, `sync-engine/mise`, with a
side parameter. The `repo` side writes `conf.d/<tool>/mise.<env>.toml` as
today, byte for byte. The `macos` side writes `conf.d/<tool>/mise.toml` and
refuses environments. Both sides accept `[mise.bootstrap.*]` and
`[mise.dotfiles]`, and a `[mise.*]` string can hold `<% %>`, rendered with the
values of the tool. `[overrides.gitignore]` stays in `sync-engine/repo-sync`.

Plan 8 of 16 — requires `docs/plans/2026-11-04-values-schema`; required by
`docs/plans/2026-11-04-setup-sync-front-end`. No reversal of a standing
decision.

## Facts the survey established

- The design rulings come from the planning session of 2026-11-04
  (artifact https://claude.ai/artifact/7LKisXPLbSGtrFfVD36ez7). Each unit
  quotes the rulings it implements; no unit reads another file for them.

The line numbers are of today's
`tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
(2024 lines). P01-P07 move that code to `sync-engine/repo-sync` and change it;
find each passage by its name.

- **What moves.** `_RS_DEFINES` (1065, the `rs_leaves` and `rs_has`
  templates), `_RS_OVERRIDES_TEMPLATE` (1086-~1160), `_rs_overrides`
  (1168-1256; its `gitignore` problems stay in repo-sync), `_rs_mise_env`
  (983-1059, the environments of each tool and the `envs` check),
  `_RS_MISE_TEMPLATE` (1270-1334), `_rs_mise_render` (1342-1404). The
  template reads `REPO_SYNC_ENV`, `REPO_SYNC_ENVS`, `REPO_SYNC_IN` and
  `REPO_SYNC_DEV_ALIASES` from the environment.
- **The section lists in `_RS_MISE_TEMPLATE`:** entry sections `tools`,
  `tool_alias`, `plugins`; every-environment sections `settings`, `env`,
  `tool_config`, `task_config`, `task_templates`, `doctor`, `deps`, `oci`;
  dev-only sections `vars`, `secrets`, `redactions`, `shell_alias`, `hooks`,
  `watch_files`, `daemons`, `daemon_groups`, `daemons_settings`,
  `daemon_providers`, `tasks`. Any other section → `test.Fail`
  "[mise.<s>] is not supported in a tools.toml" (code 2 for that tool; its
  existing conf.d files stay). The keys `envs`, `requires`, `files` of a
  `[mise.tools]` entry are dropped from the output (P03 changes `requires` to
  `requires_found` and removes `files`).
- **The generated header** (after P06):
  `# Synced from bootstrap: repo/tools/<tool>/tools.toml. Do not edit this
  file.`
  - `# The next sync replaces it, and a manual change is lost.`
- **The `mise` tool aliases.** `_rs_mise_render` adds `setup-all` and
  `setup-<submodule>` to the dev file of the `mise` tool when the repository
  has git submodules (1350-1362).
- **The values** (P07): `sync-engine/values` gives the merged values of each
  selected tool; the file templates get `.config` and `.all`.
- **gomplate** runs through the wrapper of `sync-engine/common`
  (`mise x gomplate@latest --`, from `/`), with the delimiters `<% %>`.
- **Tests.** `.config/mise/tasks/test/repo-sync` has the override cases
  (`java_conf`, `[overrides.mise.tools.java] version`, `[overrides.mise_env.ci]
  wins`, unknown paths, `envs` not overridable), the `requires` cases of
  python, and the kotlin and swift `ci` cases (near 311-500 today). A dry run
  shows a diff only for a consumer file that already exists.
- **Docs.** `CLAUDE.md` "mise config" (the section lists, `[mise_env]`, the
  overrides), `docs/design/mise.md`, `docs/design/tools.md`,
  `docs/sync/index.md`, `README.md`.
- Bash 3.2: neither engine uses a bash-4 feature. Keep it so.
- Commit types: `feat`, `fix`, `perf`, `refactor`, `revert`, `test`, `ops`,
  `docs`; no scope.

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
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

A `GAP:` is a hole in the plan the unit could proceed past on a stated
assumption; it is recorded and the run continues. An `UNRESOLVED:` is a ruling
the unit could not proceed without; it blocks the unit and its dependents.

## Out of scope

- The Mac front-end that calls the library with side `macos`, the
  `zz-macos` file and the Mac values file: P09.
- A repo tool that uses `[mise.bootstrap.*]`, `[mise.dotfiles]` or `<% %>`:
  none needs one now.
- `mise run setup:configs` in this repository and in each consumer: the user
  runs it by hand (D4).

## Parked

See P16 Parked.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-04-mise-generation-shared

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
