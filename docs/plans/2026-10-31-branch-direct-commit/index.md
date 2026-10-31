# Plan — Allow direct commits on a core branch from bootstrap.toml (2026-10-31)

## Goal

After this plan lands, a consumer can set `direct_commit = true` in
`[branch.main]` or `[branch.develop]` of `.config/bootstrap.toml`, and
`setup:configs` keeps that branch out of the `no-commit-to-branch` hook. When
no protected branch is left, the sync removes the hook. The merge safety net
no longer fails on a protected source branch, so `code:merge:main` works from
a protected `develop`. The plan finishes backlog item B13. No reversal of a
standing decision.

## Facts the survey established

- The hook is the second `[[precommit]]` table of `tools/git/tools.toml`
  (`:21-29`, `pre-commit/pre-commit-hooks` `v6.0.0`). Its args (`:26`) are
  `["--branch", "main", "--branch", "develop"]`. Its name (`:29`) is "Refuse a
  direct commit to main or develop". Its description (`:27`) names no branch.
  The rendered copy in this repository is
  `.config/pre-commit-config.yaml:40-43`.
- `code:precommit` (`tools/pre-commit/.config/mise/tasks/code/precommit:11`)
  always runs `pre-commit run --all-files` and has no flags.
- `tools/git/.config/mise/tasks/code/merge/_default:7` says "merge into main
  or develop". That stays true and does not change.
- `_rs_context` in `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
  (`:1273-1299`, documented at `:133-134`) loads each tool's `tools.toml` as
  `tool_<name>` and the consumer file `.config/repo-sync/pre-commit.yaml` as
  `repo_precommit`. The gomplate line at `:1298` joins every `.precommit`
  entry in tool order, then the consumer entries, as the `precommit` key. It
  ends with `coll.Merge $ctx .v`, so every values key, `.branch` too, also
  reaches the template.
- The template `tools/pre-commit/.config/pre-commit-config.yaml.tmpl:28-47`
  loops over `.precommit` and its `.hooks` with no filter. No values key
  changes the args of a hook today.
- `repo-sync` does not read or validate `[branch.*]`. It validates the values
  file before it writes any file, at `:1671`:
  `{ _rs_mise_env && _rs_overrides "$dir" "$values"; } || { rc=$?; rm -rf --
  "$dir"; return "$rc"; }`.
  Values-file errors return 2 (header `:163-174`). `_rs_branch_exists`
  (`:699-702`) finds a branch locally or on `origin`, and `_rs_merge_model`
  (`:706-720`) uses it to select the merge tasks.
- The only reader of `[branch.*]` is `read_merge_config` in
  `tools/git/.config/mise/tasks/_scripts/merge:150-200`. It reads with
  `mise x -- yq -p toml`, ignores keys that it does not know, and reads any
  `merge_model` other than `pr` as `direct`.
- The merge is `git merge --no-ff --no-edit` (`merge:353`). git runs
  `pre-merge-commit` for it, and `default_install_hook_types` (template
  `:17-21`) does not install that type, so the hook never runs on a merge. The
  safety net `precommit_safety_net` (`merge:63-108`, called at `:311`) runs
  `mise x -- pre-commit run --config "${config}" --all-files` (`merge:78`) on
  the source branch, with no `SKIP`. `code:merge:main` must run from
  `develop`, so with a protected `develop` the safety net always fails.
- `test:repo-sync` (`.config/mise/tasks/test/repo-sync`) defines a case with
  `run_case <name> <code> <values> [<change>] [<message>] [<check>]`
  (`:76-95`, body `:97-130`). The case runs `repo_sync --dry-run`, compares
  the exit code, and greps the log for the message. A dry run prints a
  `diff -u` only when the destination exists (`repo-sync:1505-1507`); the
  pattern is `gitignore_file` (`:483`, expected lines `:492-499`).
  `check_log_lacks` (`:367`) is for "must not appear". The nearest cases are
  "[branch.main], the merge config of the repo" (`:235-238`, code 0) and the
  `[overrides.gitignore]` errors (`:500-513`, code 2).
- Docs that describe `[branch.*]` or the hook: `docs/design/bootstrap-config.md`
  (example `:33-39`, summary table `:61-63`, `:208`, key sections `:247`,
  `:265`, `:273-283`, next heading "Reserved names" `:285`);
  `docs/sync/tools/git.md` (`:44-47`, `:55-56`); `docs/sync/tools/pre-commit.md`
  (`:22`, `:60`); `docs/sync/index.md` (`:373`, key table `:407-409`, `:439`,
  step 4 `:518-545`); `docs/sync/vwf.md` (`:109`); `docs/sync/migrate.md`
  (`:85-86`); `README.md` (`:143-144`, `:246`); `CLAUDE.md` ("Branch model in
  the merge tasks", and the hook line in "Change a shared file");
  `docs/design/packs.md` (`:155`).
- Examples with `[branch.*]` tables: `polyrepo` (`:19`, `:24`),
  `polyrepo-main` (`:20`), `polyrepo-develop` (`:19`), `monorepo` (`:57`,
  `:62`), `repo-siblings/{acme-api :24, acme-app :26, acme-product :25}`,
  `repo-submodules` (root `:25`, `frontend :24/:29`, `backend :30/:35`,
  `iac :30`). `examples/last_sync.toml` has `commit_id` `97a1080`; on
  2026-10-31 the log `git log 97a1080..HEAD -- tools docs/sync CLAUDE.md` had
  14 contract commits.
- Commit convention (`.config/git-conventional-commits.yaml`): types `feat`,
  `fix`, `perf`, `refactor`, `revert`, `test`, `ops`, `docs`, `merge`, `wip`;
  `commitScopes: []`, so a commit has no scope.
- Gates: `code:format`, `code:lint`, `code:precommit` (hidden) and
  `test:repo-sync`. The repository has no CI.
- No new dependency: `repo-sync` already uses gomplate, the merge library
  already uses `yq`.
- The consumer `virajp/macos-setup` is at
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup`. Its
  `.config/bootstrap.toml:13-14` has only `[branch.main]` with
  `merge_model = "direct"`, and its hook has the same args as here.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file with Write, Edit, `sed` or
a script.

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

- The `macos-setup` change (`[branch.main].direct_commit = true`, then
  `mise run setup:configs` there): the user does it by hand after the landing
  (D1).
- `mise run setup:configs` in this repository: CLAUDE.md requires a
  `sync-<name>` branch and its own commit, which a step after the landing
  cannot make. The user runs it by hand. Until then, the synced copies here
  (`.config/pre-commit-config.yaml`, `.config/mise/tasks/_scripts/merge`)
  keep the old hook name and the safety net without `SKIP`. This is
  accepted: this repository has only `main`, so the safety net runs on a
  feature branch and passes.
- Other `[branch.*]` tables and other protected branches: the hook protects
  only `main` and `develop`.

## Parked

- B14 (`repo-sync-check` pre-commit hook): planned next as a separate folder
  that requires this one (D9). It reuses the `direct_commit` validation of U2.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-10-31-branch-direct-commit

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
