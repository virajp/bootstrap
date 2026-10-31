# Plan — Validate bootstrap.toml at commit time with repo_sync --check (2026-10-31)

## Goal

After this plan lands, each consumer gets the gate `code:lint:repo-sync`. It
runs `repo_sync --check` on `.config/bootstrap.toml` with no bootstrap clone
and no network. The existing `lint` pre-commit hook runs it, so a commit with
a values file that breaks a local rule of the sync is blocked, and every error
shows at once. The plan finishes backlog item B14. No reversal of a standing
decision.

## Facts the survey established

`S` is `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`.

- **Main flow of `repo_sync`** (`S:1663`):
  1. Arguments (`S:1665-1676`): a `while`/`case` that accepts only
     `--dry-run`; anything else prints the usage (`S:1672`, also `S:13`,
     `S:1659`) and returns 2.
  2. The values file and its old name (`S:1678-1689`), gomplate (`S:1690`).
  3. The clone or the copy of `REPO_SYNC_CLONE` (`S:1695-1722`).
  4. Validation: `_rs_envs && _rs_mise_index && _rs_resolve && _rs_extensions`
     (`S:1729`), `_rs_merge_model` (`S:1736`), `_rs_file_skips` (`S:1741`),
     `core_tools` (`S:1744`), then
     `_rs_mise_env && _rs_overrides && _rs_direct_commit` (`S:1753`).
  5. The writes: `_rs_mise_render` (`S:1758`), context (`S:1770`), the file
     loop (`S:1798`), format (`S:1819`), stale removal (`S:1835`),
     `bootstrap_commit` (`S:1845`).
- **Local rules** (values file and local git only):
  - values file missing, code 2 (`S:1686`); old name `.config/repo-config.toml`,
    warning (`S:1682`);
  - `_rs_values_sel` (`S:291-323`): the old list format, an entry that is not
    a table, a key other than `env`, a non-list `env`, code 2;
  - `_rs_envs` (`S:836`): no `environments`, warning (`S:843`); a name other
    than `dev`/`ci`, or no `dev`, code 2;
  - the env rules inside `_rs_resolve` (`S:517-542`): an empty `env`, a name
    that is not `dev`/`ci`, a name outside `environments`, no `dev`, code 2.
    `_rs_resolve` also looks up the tools in the clone (code 3, `S:571`), so
    the local part must be split out (D5);
  - `_rs_merge_model` (`S:713`): neither `main` nor `develop`, code 2;
  - `core_tools`, warning (`S:1744-1747`);
  - `_rs_overrides` (`S:1069`, template `S:990-1054`), the structure only: the
    old `.config/repo-sync/gitignore` (`S:1073`), and the problems `table`,
    `key`, `gikey`, `gilist`, `giempty`, `envname`, `env`, `lock`
    (`S:1086-1125`), and a line in both `add` and `remove` (`S:1141`);
  - `_rs_direct_commit` (`S:735-767`): all four rules, with
    `_rs_branch_exists` (`S:706`);
  - `_rs_members` (`S:220-272`): a member `.`, a bad `type`, a folder type on
    a submodule, a submodule type not in `.gitmodules`, a missing folder. It
    is called only from `_rs_context` (`S:1324`) today.
- **Clone rules** (out of scope, D3): unknown pack (code 3, `S:340`), pack
  loop or pack named like a tool (`S:330-378`), unknown tool (code 3,
  `S:571`), a mise `depends` that no tool defines (`S:590`), `root_only` in a
  submodule (`S:640`), `_rs_mise_env` (`S:886-960`), `_rs_mise_index`
  (`S:432`), the `[overrides]` paths (`S:1035-1047`), graphifyignore without
  `claude` (`S:1106`), a `remove` line that is not shipped (`S:1144`).
- **Not validated anywhere** (parked, D7): unknown top-level keys,
  `[claude.plugins.*]`, `commit_scopes`, `extensions` (`S:691`, errors
  ignored), `repo_name`, the `[branch.*]` merge keys, and
  `.config/repo-sync/pre-commit.yaml` (read only in `_rs_context`, `S:1339`).
- **The consumer task** `tools/repo-sync/.config/mise/tasks/setup/configs/sync`
  has `#MISE tools.gomplate="latest"` (`:7`). It sources
  `${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/repo-sync`, or the
  clone's copy when `REPO_SYNC_CLONE` is set (`:17-23`), then calls
  `repo_sync [--dry-run]` (`:29-33`).
- **The gate pattern.** `code:lint`
  (`tools/linter/.config/mise/tasks/code/lint/_default.tmpl:57-62`) runs each
  `code:lint:<name>` task with the staged files. The `lint` hook
  (`tools/linter/tools.toml:16-20`) has `pass_filenames = true` and
  `require_serial = true`. A model gate is
  `tools/python/.config/mise/tasks/code/lint/ruff` (header, `#MISE hide=true`,
  `#USAGE` flags `--debug`, `--fix` and `arg "[files]" var=#true`, the
  applies-check, then `print_header`). The `repo-sync` tool has no
  `code/` tasks today and no `[[precommit]]` table
  (`docs/sync/tools/repo-sync.md:91`). It is in the `base` pack
  (`tools/packs.toml:4`).
- **`test:repo-sync`** (`.config/mise/tasks/test/repo-sync`):
  `run_case <name> <code> <values> [<change>] [<message>] [<check>]`
  (`:76-95`). `case_body` (`:97-132`) exports `REPO_SYNC_CLONE`,
  `REPO_SYNC_VALUES`, `MISE_PROJECT_ROOT` and `REPO_SYNC_NO_AUTOUPDATE=1`,
  sources the clone's `S`, and runs `repo_sync --dry-run` (`:110-117`), which
  is hard coded. Similar cases: `direct_commit` (`:583-614`), the old
  gitignore (`:542`), overrides (`:320-339`). The collect loop is at `:617`;
  the `setup:configs` cases start at `:621`.
- **Docs:** `docs/sync/tools/repo-sync.md` (Tasks `:33-56`, steps `:56-80`,
  env vars `:80-89`, Pre-commit hooks `:91-93`, Files `:95-100`, Migration
  `:106`); `docs/design/bootstrap-config.md` (key table `:43-66`,
  `direct_commit` `:288-330`); `docs/sync/index.md` (dry run `:218`, `:358`,
  `:594-597`; values keys step 3, `:361-`); `docs/sync/migrate.md` (highest
  item M22, `:487`); `README.md` ("How a repository syncs" `:130-163`, Tests
  `:294-299`); `CLAUDE.md` ("How the sync operates", exit codes, the test
  task; the gate rules in "Task scripts").
- **Commit convention:** types `feat`, `fix`, `perf`, `refactor`, `revert`,
  `test`, `ops`, `docs`, `merge`, `wip`; no scopes.
- **Gates:** `code:format`, `code:lint`, `code:precommit` (hidden, always
  `--all-files`, no flags) and `test:repo-sync`. No CI.
- **Lessons of the B13 run** (its gaps G1 to G7): give exact rules where a unit
  must choose, and make each keep check runnable as written. In a temporary
  folder, `mise x` drops the mise install folders from `PATH`; K1 of B13
  needed a symlink of the `pre-commit` binary in a scratch `bin` folder.

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

- The seven rules that need the clone (D3). `setup:configs` still examines
  them.
- The examples and `examples/last_sync.toml`: the values-file contract does
  not change, so no example changes.
- `mise run setup:configs` in this repository (on a `sync-<name>` branch) and
  in each consumer: the user runs them by hand.

## Parked

- New rules that no code examines today (D7): unknown top-level keys, the
  shape of `[claude.plugins.user]` and `[claude.plugins.project]`, the types
  of `commit_scopes` and `extensions`, the `[branch.*]` keys `merge_model`,
  `merge_from` and `merge_never`, and the shape of
  `.config/repo-sync/pre-commit.yaml`. Each new rule can stop the sync of a
  consumer that works today, so each one needs a `migrate.md` item.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-10-31-repo-sync-check

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
