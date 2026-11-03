# Plan — Design docs and lazy-loaded CLAUDE.md (2026-11-04)

## Goal

After this plan lands, each rule of the merged sync has one place, in
`docs/design/`: the new `sync-engine.md`, `repo-sync.md` and
`setup-sync.md`, and the updated `packs.md`, `tools.md` and `mise.md`.
`CLAUDE.md` keeps the layout, the change steps, the task rules, the test
rule and a list of pointers ("Read X before you change Y").
`docs/sync/index.md` points the same way. The script headers of
`sync-engine/*` keep usage, flags and a pointer; the long contract comments
go. No behavior changes. The design docs are the owned files of U1 in this
plan; the docs unit owns `CLAUDE.md`, `README.md` and `docs/sync/`.

Plan 15 of 16 — requires `2026-11-04-macos-tools-apps`; required by
`2026-11-04-docs-setup`. No reversal of a standing decision.

## Facts the survey established

The design rulings come from the planning session of 2026-11-04 (artifact
https://claude.ai/artifact/7LKisXPLbSGtrFfVD36ez7). Each unit quotes the
rulings it implements; no unit reads another file for them.

`MS` is `/Users/virajpatel/Projects/github.com/virajp/macos-setup` (read
only).

- **Survey state.** The survey ran before P01 to P14 landed. Line numbers
  and section texts below are from that state; the run finds each passage by
  its heading, because P01 to P14 changed the text below each heading.
- **`CLAUDE.md`** (554 lines at the survey) has these sections: Purpose,
  Layout rule (`:13`, also the `tools.toml` key rules), How the sync
  operates (`:128`, cache, lock, clone, values file, exit codes,
  `--check`, gate, dprint, gomplate, fnox profiles, the stale removal,
  `bootstrap_commit`, `[overrides.gitignore]`, `[claude.plugins.*]`,
  pre-commit render and `autoupdate`, `test:repo-sync`), Templates
  (`:242`), Task scripts (`:275`), Task output (`:361`), Branch model in
  the merge tasks (`:391`), mise config (`:431`), Change a shared file
  (`:534`). P01 to P14 added text for the new layout, packs, `[files]`,
  manifest, values schema, the Mac front-end and the test rule.
- **`docs/design/`** at the survey: `index.md` (overview and references),
  `bootstrap-config.md` (the keys of `.config/bootstrap.toml`),
  `fnox.md`, `mise.md` (416 lines: files, environments, the `[mise.*]`
  sections, mise tools, tasks, shell aliases), `packs.md` (183 lines),
  `tools.md` (93 lines). `sync-engine.md`, `repo-sync.md` and
  `setup-sync.md` do not exist.
- **`MS/docs/setup-sync.md`** (559 lines) is the contract of the old Mac
  engine: Variables, The entry task, Run order, mise, The cache clone, The
  `tools/` layout, `tools.toml` (file attributes, questions,
  contributions), `packs.toml`, The config folder, Selection, Values and
  the template context, Generated files, seeds and merge files (merge
  files, the header, `generated.txt` and the stale rule, blockers),
  Symlinks, Stages, Failures, Where setup-sync writes, Return codes, Tests.
  Parts of it are out of date under the rulings: the iCloud config folder,
  `<tool>.toml` values files, `tools = [...]`, `[overrides.<tool>]`,
  `generated.txt`, the pinned gomplate, jq and yq, the iCloud wait,
  `macos_setup_commit`.
- **`sync-engine/`** after P14: `clone`, `common`, `write`, `select`,
  `merge`, `manifest`, `values`, `mise`, `repo-sync`, `setup-sync`. The old
  B header contract (about 300 lines) moved into `sync-engine/setup-sync`
  with P09, unless P09 already cut it.
- **`docs/sync/index.md`** (711 lines at the survey) is the adoption prompt
  for an AI agent: Rules, Requirements, the repository shapes, Steps 1 to 9.
  Step 3 lists the keys of the values file.
- **Tests:** `mise run test` runs `test:repo-sync`, `test:setup-sync` and
  `test:macos-tools`. A comment change does not change a test result.
- **Commit convention:** types `feat`, `fix`, `perf`, `refactor`, `revert`,
  `test`, `ops`, `docs`; no scopes. No CI.

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

- The Mac adoption and migration docs (`docs/setup/`): P16
  (`2026-11-04-docs-setup`).
- `macos-setup/CLAUDE.md`, which becomes short: a change in the
  `macos-setup` repo (see P16 Parked).
- Any change to a rule: this plan moves text and does not change behavior.
- `mise run setup:configs` in this repository and in consumers: a hand step.

## Parked

See P16 Parked (`docs/plans/2026-11-04-docs-setup/index.md`).

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-04-docs-design-split

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
