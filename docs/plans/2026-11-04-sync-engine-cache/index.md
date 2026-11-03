# Plan — Shared cache and clone for the sync engine (2026-11-04)

## Goal

After this plan lands, bootstrap has the folder `sync-engine/`. It holds the
clone library `sync-engine/clone` and the repo front-end
`sync-engine/repo-sync`, which moves there from the `repo-sync` tool. Every
`setup:configs` run uses one shared cache, `~/.cache/sync-engine/`, with the
clone in `repo/` and the lock in `lock/`. It clones
`https://github.com/virajp/bootstrap.git` on the branch `main`, and it removes
the old cache folders `~/.cache/bootstrap/` and `~/.cache/bootstrap.lock`. A
forwarding file at the old library path keeps the
old consumer stubs in operation for one more sync. Consumers change nothing.

Plan 1 of 16 — requires no plan; required by
`docs/plans/2026-11-04-sync-engine-core`. The chain moves the tools of
`macos-setup` into bootstrap with one shared sync engine. The design rulings
of 2026-11-04 are the input; each unit quotes the rulings that it applies.

## Facts the survey established

- **The design rulings** come from the planning session of 2026-11-04
  (artifact https://claude.ai/artifact/7LKisXPLbSGtrFfVD36ez7). Each unit
  quotes the rulings it implements; no unit reads another file for them.

Paths: `BS` is this repository. `MS` is
`/Users/virajpatel/Projects/github.com/virajp/macos-setup` (read only; a unit
reads it by its absolute path; no unit edits it).

- **The engine today.** A =
  `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
  (2024 lines, `_rs_*` functions, entry `repo_sync`). It does not lock. It
  clones by itself in `repo_sync` (A:1870-1888) when `REPO_SYNC_CLONE` is not
  set: `git clone --depth 1 --single-branch` of
  `${REPO_SYNC_URL:-git@github.com:virajp/bootstrap.git}`. With
  `REPO_SYNC_CLONE`, it copies that clone into a temp folder (A:1877-1884,
  code 5 when it cannot).
- **The entry task.** A_TMPL =
  `tools/repo-sync/.config/mise/tasks/setup/configs/_default.tmpl` holds the
  clone, cache and lock code (A_TMPL:31-112): URL
  `${REPO_SYNC_URL:-git@github.com:virajp/bootstrap.git}` (31), cache
  `${XDG_CACHE_HOME:-${HOME}/.cache}/bootstrap` (32), lock `${cache}.lock`
  (33), `remove_stale_lock` with a `.takeover` folder (54-71), the wait loop
  with the message `Waiting for another sync of bootstrap (pid N) ...`
  (76-85), an exact origin compare (91-92), `fetch --depth 1 origin HEAD`
  (94), a clone again for another origin (100-108), a copy in
  `${TMPDIR:-/tmp}/bootstrap.XXXXXX`, then the lock release (110-113). Then
  `REPO_SYNC_CLONE="${copy}" mise run setup:configs:sync` (117) and the same
  for each member submodule (118-138).
- **The Mac entry task.** B_DEF =
  `MS/tools/macos-setup/.config/mise/tasks/updateall/configs/_default`
  (lines 91-216) is a near copy: cache
  `${XDG_CACHE_HOME:-${HOME}/.cache}/macos-setup`,
  clone `${cache}/repo`, lock `${cache}/lock`, HTTPS URL, explicit branch
  `main` (152, 184), `normalize_url` (the origin compare that matches the SSH
  and HTTPS URLs of one repository, B_DEF:110-125), the raw origin from
  `git config --get remote.origin.url`, and a non-git, non-empty clone folder
  that stops with exit 2 (B_DEF:197-200). It unsets `GIT_DIR GIT_WORK_TREE
  GIT_INDEX_FILE` first. Its `generated.txt` already has
  `<tool><TAB><path>` lines.
- **The sync task.** A_SYNC =
  `tools/repo-sync/.config/mise/tasks/setup/configs/sync`
  has `#MISE tools.pre-commit="latest"` and `#MISE tools.gomplate="latest"`,
  and sources the local library, or
  `${REPO_SYNC_CLONE}/tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
  when that variable is set (A_SYNC:19-24). Each consumer has this old task
  and the old stub, so the old library path must keep its function in the
  clone during the transition.
- **The gate.** A_LINT =
  `tools/repo-sync/.config/mise/tasks/code/lint/repo-sync`
  (44 lines) exits 0 without gomplate (19), filters the file list (25-35),
  prints its header (37), unsets `REPO_SYNC_CLONE` (41) and sources the local
  library (43).
- **Template context.** `_rs_context` (A:1428-1487) builds the JSON context
  of each template from `REPO_SYNC_*` environment variables in one gomplate
  call. The context keys are in the header comment of A.
- **Header and lint.** Every file of `tools/` that is not a seed starts with
  the bootstrap header. `shell_files_in_scope` of `helpers` lints a file only
  when its first line is a shell shebang or its name ends in `.sh`/`.bash`.
- **Tests.** `.config/mise/tasks/test/repo-sync` (978 lines) copies only
  `tools/` of the working tree into a scratch clone (:46-50). `case_body`
  (:100-138) exports `REPO_SYNC_CLONE`, `REPO_SYNC_VALUES`,
  `MISE_PROJECT_ROOT`, `REPO_SYNC_NO_AUTOUPDATE=1`, maps `CASE_REPO_SYNC_URL`
  to `REPO_SYNC_URL`, unsets `REPO_SYNC_CLONE` for `--check`, and sources
  `$clone/tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`. The
  `setup:configs` cases (:812-972) render `_default.tmpl` with gomplate and a
  `ctx.json`, replace `mise` with a fake that checks the copy and the lock,
  and test the clone, the fetch, the wait for a live holder, the takeover of a
  dead holder, a clone again for another URL, a failed fetch (5) and the
  member flags.
- **Docs that describe the cache and the variables:** `CLAUDE.md` ("How the
  sync operates", lines 130-205; "Layout rule"; "Task scripts"),
  `README.md:133`, `docs/sync/tools/repo-sync.md` (table of files :19-26,
  cache :42-55, gate :67-80, variables :121-130), `docs/sync/index.md:669`
  (code 5), `docs/design/bootstrap-config.md:364`. `docs/sync/migrate.md`
  ends with M26 today.
- **Repo facts.** Commit types: feat, fix, perf, refactor, revert, test, ops,
  docs, merge, wip; no scopes. `.config/releases.yaml` holds
  `bootstrap: MAJOR` (pending). bash 3.2: neither engine uses a bash 4
  feature; keep it so.

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

- `mise run setup:configs` in this repository and in each consumer (D4): the
  user runs them by hand after the landing.
- The shared libraries `common`, `write` and `select`: plan 2
  (`2026-11-04-sync-engine-core`).
- `~/.cache/macos-setup/` and the move of its `generated.txt` to
  `macos-generated.tsv`: the Mac entry task of plan 10 does them (D12).
- The Mac entry task and its lock over the whole run: plan 10
  (`2026-11-04-macos-entry-tool`). This plan only gives the lock functions.
- The examples and `examples/last_sync.toml`: the values-file contract does
  not change.

## Parked

See P16 Parked (`docs/plans/2026-11-04-docs-setup`).

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-04-sync-engine-cache

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
