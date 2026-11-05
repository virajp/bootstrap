# Plan — Changed-only output and a count table for the sync (2026-11-05)

## Goal

`setup:configs` and `updateall:configs` list only what changed and end with a
table of counts, with glyphs and colors. `setup:configs` also runs the current
sync task, and it shows this below the line `Updating the clone`. Each task
reports the old cache only when it removes that cache.

The user saw and accepted a sample of the output before this plan. This plan
has no reversal of a standing decision.

## Facts the survey established

Line numbers are from commit `5961aaf`.

- **Run path of `setup:configs`:** The rendered `_default`
  (`repo/tools/repo-sync/.config/mise/tasks/setup/configs/_default.tmpl`) embeds
  `sync-engine/clone`, prints `Fetching bootstrap ...` (:35), fetches the clone
  and makes the copy (`_se_copy`, :41). Then it runs `mise run
  setup:configs:sync ${DRY_RUN}` (:46), which is the local plain file
  `setup/configs/sync`. That file sources `sync-engine/repo-sync` from
  `SYNC_ENGINE_CLONE`.
- **Members:** `sync_member` (`_default.tmpl`:52-60) runs the sync task in each
  selected submodule. Its guard at :55 needs an executable local `sync`.
- **Stale in a run today:** the local `_default`, `sync` and `helpers`. They are
  replaced in the file loop with `mv -f`, so the running scripts keep the old
  inode. All of `sync-engine/*` is fresh.
- **Clone copy paths:**
  `${_SE_COPY}/repo/tools/repo-sync/.config/mise/tasks/setup/configs/sync` and
  `${_SE_COPY}/repo/tools/mise/.config/mise/tasks/_scripts/helpers`. Both tools
  are in the `base` pack, and neither file has a condition.
- **`_default.tmpl`:20** sources `helpers` before the copy. The task keeps the
  old print functions for that run, which is accepted.
- **Old-cache bug:** `sync-engine/clone`:115 continues when only
  `bootstrap.lock` exists, and :124 is
  `[[ -e ${old} ]] || _se_say info "Removed the old cache ${old}"`.
- **Per-file lines:**
  - `sync-engine/write`:106-134 has `unchanged`, `would chmod`, `chmod`,
    `would update` with a diff, and `updated`. A new seed reports `updated`
    today. :89-91 has `removed` for a file at a folder path. :55 has the parse
    failure.
  - `sync-engine/repo-sync` has `kept seed` (1092, 1422), `skip seed` (1094),
    `skip merge` (1100), `skip file` (1104), `skip env` (1415), `kept existing`
    (1437, 1529, 1594), symlink lines (1436, 1508), `removed` (1040),
    `would remove` (1033) and `cannot remove` (1037). The end lines are at
    1892-1896. The first lines are at 1739, 1754 and 1755.
  - `sync-engine/manifest`:182-242 has the lines of the stale rule and of the
    manifest write.
  - `sync-engine/common`:93 and :140 have the `bootstrap_commit` lines, and :14
    has `_se_log`, which writes to stderr with no color.
- **Mac entry:**
  `macos/tools/macos-setup/.config/mise/tasks/updateall/configs/_default.tmpl`
  runs `bash "${sync}" "$@"` from the copy (:125), so its `sync` is never stale.
  It has its own `log()` (:15), its flags are read at :33-50, and it has no
  `helpers`. `configs/sync`:12 prints `setup-sync: running ...`.
- **`sync-engine/setup-sync`:** One loop handles all tools (`_ss_render`,
  :672-688). `_ss_file` has its own lines at :555, :562-565 and :631. The end
  lines are at :745, :831, :835 and :885. It returns 4 at :863 with no end line.
  `mise bootstrap` runs before the end line (:879).
- **`_se_say`** (`sync-engine/clone`:25-35) uses `print_info` and `print_error`
  when they exist. Otherwise it prints with the prefix `sync-engine: `.
- **Colors of `helpers`:** `repo/tools/mise/.config/mise/tasks/_scripts/helpers`
  has the colors at :30-36, the glyphs at :39-44 and `style_enabled` at :66-75.
- **Tests:** `test/repo-sync` has `case_body` (:111) with its exports at :128,
  plus `case_rerun` (:1075-1076) and `CASE_CALL` paths. `test/setup-sync` has
  `sync_run` (:164-169) and entry runs (:1265, :1276, :1389). `test/macos-tools`
  calls the engine at :111. `Updating the clone` is matched at repo-sync :3157,
  :3216 and setup-sync :1318, :1345. No test matches `Removed the old cache` or
  `Fetching bootstrap`.
- **Docs that describe today's output:**
  - `docs/sync/tools/repo-sync.md`:22-23, 48-49, 52-74, 84-85, 171, 196, 208
  - `README.md`:392-398
  - `CLAUDE.md`:214-222, 368, 439-447
  - `docs/sync/migrate.md`:330
  - `docs/sync/index.md`:744
- **Other plans in the queue:** No Mac plan owns a file of U1, U2 or U3.
  `2026-11-04-docs-design-split` owns `sync-engine/**` and runs later.
- **Commit types:** `feat`, `fix`, `perf`, `refactor`, `revert`, `test`, `ops`,
  `docs`, `merge`, `wip`. The repository has no scopes.

## The accepted sample

The file names and the counts are examples. `setup/configs/sync` does not show
as changed, because of D5.

```text
🚀 Fetching bootstrap ...
ℹ️ Updating the clone in ~/.cache/sync-engine/repo
ℹ️ Updated the local sync task to 5961aaf
🚀 Syncing the shared files from bootstrap ...
ℹ️ bootstrap 5961aaf · 14 tools · environments: dev ci

  ✎ changed   .gitignore
  ✚ new       .config/mise/tasks/code/lint/repo-sync
  ❋ seeded    .config/swiftlint.yml
  ✖ removed   .config/mise/conf.d/old-tool/mise.dev.toml

  ┌───────────────┬───────┐
  │ ✎ Changed     │     1 │
  │ ✚ New         │     1 │
  │ ❋ Seeded      │     1 │
  │ ✖ Removed     │     1 │
  │ = Unchanged   │   143 │
  │ ◦ Kept seeds  │     6 │
  │ ⊘ Skipped     │    11 │
  │ ✗ Failed      │     0 │
  ├───────────────┼───────┤
  │ Total         │   164 │
  └───────────────┴───────┘
✅ Synced: 4 files changed
```

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file.

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

- The output of other tasks, for example `setup:ai`: the request names only the
  sync.
- A `docs/sync/migrate.md` item: no consumer contract changes.
- The sync of the other consumer repositories: it stays a hand step.
- A self-update of `_default`: it needs the full render, and the file loop
  already writes it for the next run.

## Parked

- A total table across the members for `setup:configs --all`.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-11-05-setup-configs-output

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
