# Plan — One packs model (2026-11-04)

## Goal

After this plan lands, `tools/packs.toml` has the merged shape of both sides:
each key of a `[pack.<name>]` table is a member, `<tool> = { default = true }`
or `<tool> = {}`, and a nested pack is `"pack:<name>" = { default = true }`.
`[packs.<name>]` of a values file adds each default member; a tool can be in
several packs, and the engine adds it once. The old keys `depends`,
`required`, `tools`, `default = [...]` and the prefix `p:` stop the sync with
code 2. `required` goes: each scanner (`gitleaks`, `grype`, `osv-scanner`)
depends on `_security_base`, so one scanner alone brings it. Values files do
not change; a `docs/sync/migrate.md` item explains the change.

Plan 4 of 16 — requires `docs/plans/2026-11-04-repo-files-table`; required by
`docs/plans/2026-11-04-repo-manifest`.

## Facts the survey established

- **The design rulings** come from the planning session of 2026-11-04
  (artifact https://claude.ai/artifact/7LKisXPLbSGtrFfVD36ez7). Each unit
  quotes the rulings it implements; no unit reads another file for them.

`A` is `sync-engine/repo-sync`; `select` is `sync-engine/select` (from plan
2: `_se_pack`, `_se_selected`, `_se_walk`).

- **packs.toml today** (8 lines):
  `[pack.base] depends = [ "mise", "repo-sync", "git", "pre-commit", "dprint",
  "linter", "vscode", "p:security" ]`;
  `[pack.security] depends = [ "gitleaks", "grype", "osv-scanner" ]`,
  `required = [ "_security_base" ]`. The header comment names
  `docs/design/packs.md`.
- **The reader today** (`_rs_selected`, A:403-431, now `_se_selected`): one
  gomplate call prints `pack <name>` and `item <name> <entry>` lines from
  `depends` and `required`; a pack named like a tool folder is code 2.
  `_rs_pack`
  (A:372-396, now `_se_pack`) expands `p:<name>` items, finds a loop
  (`a pack loop: a -> b -> a`, code 2) and an unknown pack
  (`pack '<name>' is not in bootstrap`, code 3). An unknown tool is found
  later, in the walk (`tool '<x>' is not in bootstrap`, code 3).
- **Key order.** The global dprint config sets `sortKeys: true` for TOML, so
  `code:format` sorts the keys of `tools/packs.toml`. The root
  `dprint.jsonc` has an empty `excludes` list today. Today's pack base order
  (mise, repo-sync, git, pre-commit, dprint, linter, vscode, security) is the
  tool order and the order of the pre-commit hooks.
- **The three scanners** have `depends = []` today; `_security_base` arrives
  only through `required`.
- **The Mac packs today** (`MS/tools/packs.toml`, read only):
  `[pack.<name>]` with `tools` and an optional `default` list, one pack per
  tool. The merged model replaces both formats; the Mac side adopts it in
  plan 9.
- **Tests:** `test:repo-sync` writes old-shape packs in `pack_loop` and
  `pack_named_like_tool` (:187-194) and has the cases "a pack loop",
  "a pack named like a tool folder", "[packs.nope], an unknown pack"
  (:241-266). `base_tools` (:159) lists `_security_base` among the tools of
  the pack base.
- **Docs:** `docs/design/packs.md` (183 lines; `depends`, `required`, `p:`,
  "Common tools and required tools" :135-150), `docs/sync/packs/base.md`,
  `docs/sync/packs/security.md`, `CLAUDE.md` ("Layout rule": the pack
  bullet with `depends`, `p:<name>` and `required`), the header comment of A.

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

- A way to drop one default member of a selected pack: still open, as in
  `docs/design/packs.md` (the design rulings of 2026-11-04).
- The Mac `packs.toml`: plan 10 and the import plans write
  `macos/packs.toml` in this shape.
- The move of `tools/packs.toml` to `repo/packs.toml`: plan 6.
- `mise run setup:configs` in this repository and in each consumer (D4).

## Parked

See P16 Parked (`docs/plans/2026-11-04-docs-setup`).

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-04-packs-merged

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
