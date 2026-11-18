# Plan — The Python task libraries and the rules of a Python task (2026-11-18)

## Goal

After this plan lands, each shell library of the tasks has a Python module
beside it: `helpers.py`, `checks.py`, `placeholder.py` and `worktrees_page.py`
in the mise tool, `merge.py` in the git tool and `xcode.py` in the swift tool.
Each module has the same names and gives the same output as its shell library.
The gates find a Python part, dprint is the one formatter of the Python
files (the gate `code:format:40-ruff` is gone), and the uv cache is on. The docs
give the rules of a Python task.
No shipped task changes yet, so a consumer sees no change in its task output.

This is plan 1 of 3 of B30 ("Convert every shell mise task to Python run by
uv"). It requires nothing. Plan 2, `docs/plans/2026-11-18-python-tasks-repo`,
converts the shipped tasks of `repo/tools/` and deletes the shell libraries.
It also converts the own tasks of this repository. Plan 3,
`docs/plans/2026-11-18-python-tasks-mac`, converts the Mac tasks.

Reversals (D3, confirmed by the user): the bash 3.2 rule of CLAUDE.md "Task
scripts" ends for a Python task; D4 and D5 of
`docs/plans/archived/2026-11-15-gate-file-sets` get a Python form; the
gate `code:format:40-ruff` ends; the `UV_NO_CACHE` setting ends.

## Facts the survey established

- **Scale.** `repo/tools/*/tasks/`: 68 files, about 6000 lines. The largest
  are `code-worktrees` (873 lines), `merge` (433), `helpers` (363),
  `worktrees-page` (254) and `setup-deps-android` (245). The mise tool has 19
  task files (2624 lines). `macos/tools/*/tasks/`: about 33 files and 1000
  lines, with an own `helpers` (83 lines). The own tasks of this repository:
  `test/go`, `test/50-local` and `vorssaint/diff`.
- **The one Python task now.** `repo/tools/mise/tasks/gate_files.py` is the
  model of D4: shebang, header on lines 2-3, `# /// script`, `# MISE` lines.
  ruff format rewrote `#MISE` to `# MISE`; mise still reads it. mise also
  reads `# USAGE` and `# [USAGE]`, removes `.py` from the task name, and takes
  `_default.py` as the task of its folder. Each flag comes as `usage_<flag>`.
  Each new task file needs `mise trust` in a scratch repository.
- **The shell helpers** (`repo/tools/mise/tasks/helpers`): `style_enabled`
  (line 66), `print_line` (90), `_line_sep`, `line_sep1`, `line_sep2`,
  `print_header` (163), `print_subheader`, `print_success`, `print_warn`,
  `print_error`, `print_info`, `print_wait`, `print_header_wait`,
  `print_yellow`, `print_newline`, `has_own_task` (247), `own_tasks` (258),
  `bootstrap_get` (288), `members` (332) and `gate_files` (353). `members`
  parses with yq. `own_tasks` matches `[0-9][0-9]-?*` and prints the file
  name, so a `70-ruff.py` part prints `code:lint:70-ruff.py` now (D18).
- **The library entries.** `repo/tools/mise/tools.toml` lines 81-98 hold
  `[tasks."_scripts/<name>"]` with `library = true` for `checks`,
  `gate_files.py`, `helpers`, `placeholder` and `worktrees-page`;
  `repo/tools/git/tools.toml:68` holds `_scripts/merge`;
  `repo/tools/swift/tools.toml:48` holds `_scripts/xcode`.
  `cli/internal/source/inventory.go:250` refuses a file below `_scripts/`
  without `library = true`.
- **The init task** (`[mise.tasks.init]`, `repo/tools/mise/tools.toml:40`)
  runs `chmod 755` on each file below `.config/mise/tasks/`, so a `.py` file
  gets the execute bit.
- **The --all search.** `code-format.tmpl:38` and `code-lint.tmpl:41` run
  `grep -q '^#USAGE flag "--all"'` on the part file `${task//://}`.
- **The ruff tool** (`repo/tools/python/tools.toml`): `[mise.tools.ruff]`,
  `[mise.tools."pipx:pyrefly"]`, `code/format/40-ruff`, `code/lint/50-pyrefly`
  and `code/lint/70-ruff` all have `requires_found = [ "pyproject.toml" ]`.
  `[mise.env]` has `UV_NO_CACHE = true`; `macos/tools/uv/tools.toml:16` too.
  Each repository has the python tool, because the pre-commit tool depends on
  it. The dprint ruff plugin formats `.py` with `indentWidth = 2` and
  `lineLength = 120`.
- **dprint.** Commit `1b01eb6` removed the shuck association of the task files
  without an extension. No dprint change is necessary.
- **The tests.** The Go tests are on main (merge `56a6640`): module
  `github.com/virajp/bootstrap`, folder `e2e/`. The common code is
  `e2e/internal/harness` and `e2e/internal/repocase`; do not edit it. The
  gate_files cases are in `e2e/repo/tasks/gate_test.go`. A test of a rendered
  task gets `harness.Env()` and runs with `t.Parallel()`; it never uses
  `os.Setenv` or `os.Chdir`. One package runs with
  `mise x -- go test ./e2e/<package>/`.
- **The engine.** `cli/internal/oldform/tasks.go:179` `dependsRe` already
  accepts `# MISE depends`. The engine runs no task file, so no engine change
  is necessary (D14).
- **The docs that state the shell rules** (U7): CLAUDE.md "Task scripts" and
  "Task output"; `docs/design/mise.md` 580-629, 669-676, 738-759, 792-793;
  `docs/design/tools.md` 223-318, 380-394; `docs/design/cli.md` 1284-1291;
  `docs/design/repo.md` 27-37, 187-191, 248-250, 369-372;
  `docs/design/bootstrap-config.md` 356-370; `docs/sync/index.md` 126-155,
  640 (`UV_NO_CACHE`), 701; `docs/sync/tools/{mise,git,swift,python}.md`
  (`python.md:59` names `UV_NO_CACHE`); `README.md` 229-233. The line numbers
  are from the survey; read the passage, not the number.

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

- The engine scripts of `repo/tools/*/scripts/` stay bash (D2): the gap
  record `docs/memory/gaps/2026-11-18-engine-scripts-python-parked.md`.
- A change to the code of `cli/` (D14). A unit that needs one stops and asks.
- ruff and pyrefly in a repository without a `pyproject.toml`: the synced
  task files get no lint and no type check (D13).
- The synced copies at the root of this repository: the self-sync after the
  landing (A1) writes them. No unit edits `.config/mise/tasks/`.

## Parked

- B30: the shipped tasks of `repo/tools/` in Python, the deletion of the shell
  libraries and the `migrate.md` item —
  `docs/plans/2026-11-18-python-tasks-repo`.
- B30: the own tasks of this repository in Python —
  `docs/plans/2026-11-18-python-tasks-repo`.
- B30: the Mac tasks and the Mac helpers in Python —
  `docs/plans/2026-11-18-python-tasks-mac`.
- The other consumer repositories run `mise run setup:configs` by hand after
  plan 3.

- The `code:check` type-check group for each language (backlog item B31, P0):
  a change plan of its own that requires
  `docs/plans/2026-11-18-python-tasks-repo`.
  pyrefly stays `code:lint:50-pyrefly` in B30.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-18-python-tasks-foundation

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
