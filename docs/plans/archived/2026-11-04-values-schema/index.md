# Plan — [defaults], [questions], [contributes] (2026-11-04)

## Goal

After this plan lands, a repo tool can declare `[defaults]`, `[questions.<key>]`
and `[contributes.<tool>]` in its `tools.toml`, and a consumer gives values in
the keys of `[tools.<name>]` next to `env`. The engine merges the values in one
shared library, `sync-engine/values`, a port of the values merge of
setup-sync, so the Mac front-end (P09) uses the same code. Each template gets
`.config` (the values of its tool) and `.all` (the values of every selected
tool). The top-level keys of `bootstrap.toml` do not move.

Plan 7 of 16 — requires `docs/plans/2026-11-04-repo-layout`; required by
`docs/plans/2026-11-04-mise-generation-shared`. No reversal of a standing
decision.

## Facts the survey established

- The design rulings come from the planning session of 2026-11-04
  (artifact https://claude.ai/artifact/7LKisXPLbSGtrFfVD36ez7). Each unit
  quotes the rulings it implements; no unit reads another file for them.

- **The source of the port.** B =
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/macos-setup/.config/mise/tasks/_scripts/setup-sync`
  (read only). `_SS_INPUTS_TPL` (near 744) builds one JSON input with
  gomplate; `_SS_VALUES_JQ` (749-~829) merges: `replace_merge`, `append_new`,
  `contribution_merge`, `contributors_merge` (two different scalars → error
  with the path), `keep_merge` (a base or dependency tool: lists append, a
  type change → error), `user_merge`; it prints `ctx<TAB><tool><TAB><json>` or
  `bad<TAB><message>` lines. `_ss_values` (831-857) runs it and returns 2 on a
  `bad` line. `_SS_KEEP` (587, 656-681) holds the base and dependency tools.
  The `[questions]` checks are in `_SS_MANIFEST_TPL` (455): the keys of a
  question are `question` (required string), `detect` (string) and `example`.
  The contract comment is at B 128-160. The contribution checks are at B
  700-716 (a contribution of a selected tool to a tool that the repo does not
  have → exit 3).
- **The repo side today** (line numbers of today's
  `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`; P01-P06 move the
  code into `sync-engine/repo-sync`; find each by its function name):
  `_rs_values_sel` (333-370) stops with code 2 on a key other than `env` in a
  `[tools.*]` or `[packs.*]` entry ("'<key>' is not a key of an entry; the
  only key is env, a list"). `_rs_context` (1428-1487) builds one shared JSON
  context, `coll.Merge $ctx .v`, so the engine keys (`members`, `submodules`,
  `tools`, `extensions`, `precommit`, `member_folders`, `member_types`) win
  over a values key of the same name. The repo engine does not reject an
  unknown top-level key of a `tools.toml`.
- **After P02.** `sync-engine/common` has the wrappers
  `mise x jq@latest --` and `mise x gomplate@latest --`, run from `/`.
- **After P06.** Tool folders are `repo/tools/<tool>/`; messages name
  `repo/tools/<tool>/tools.toml`.
- **Tests.** `.config/mise/tasks/test/repo-sync` has one `--check` case for
  "a key other than `env`" in a `[tools.*]` entry (lines 696-698 today); it
  expects code 2. No sync case tests that rule. A case can change the scratch
  clone with a change
  function (for example `cp -- "$tmp/python.toml" "$clone/.../tools.toml"`).
  A dry run shows a diff only for a consumer file that already exists.
- **Docs.** `docs/design/tools.md` (the `tools.toml` keys),
  `docs/design/bootstrap-config.md` (the key table of the values file and
  `[tools.<name>]`), `docs/sync/index.md` step 3 (the values keys),
  `CLAUDE.md` ("Templates", the `tools.toml` keys bullet, "How the sync
  operates" for `[tools.<folder>]`), `README.md`.
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

- A repo tool that uses the new tables: no tool needs one now.
- The move of the top-level keys of `bootstrap.toml` (`commit_scopes`,
  `extensions`, `[claude.plugins.*]`, `[branch.*]`) into `[tools.<name>]`:
  the rulings keep them.
- Values in generated mise config (`<% %>` in `[mise.*]`): P08.
- The Mac values file `macos.toml`: P09.
- `mise run setup:configs` in this repository and in each consumer: the user
  runs it by hand (D4).

## Parked

See P16 Parked.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-04-values-schema

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
