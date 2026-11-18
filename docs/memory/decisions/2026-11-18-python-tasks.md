DECISION 2026-11-18 ★4 | repo:bootstrap, repo/tools/mise, repo/tools/git,
repo/tools/swift, repo/tools/python, macos/tools/uv, the mise task files |
each mise task becomes a Python file run by uv (B30, three plans); plan 1 adds
a Python module beside each shell library, dprint is the one formatter of the
Python files, the gate code:format:40-ruff ends and the uv cache is on | alt
rejected: one plan, the engine scripts in Python too, argparse, third-party
packages, ruff format beside dprint, keep the shell rules beside the Python
rules

# The mise tasks in Python

The user ruled this in the plan interview of
`docs/plans/2026-11-18-python-tasks-foundation` (its decisions D1 to D30),
and during its run (D31 to D33). It is plan 1 of 3 of the backlog item B30,
"Convert every shell mise task to Python run by uv".

## Reversals (D3)

The user confirmed these reversals ("yes"):

- The bash 3.2 rule of CLAUDE.md "Task scripts" ends for a Python task. The
  shell rules stay, marked as the rules of a shell task, until plan 2 deletes
  the shell libraries.
- D4 and D5 of `docs/plans/2026-11-15-gate-file-sets` (record
  `2026-11-15-gate-file-sets.md`) get a Python form. D4 gave each gate the
  call `gate_files "$0"` of the shell `helpers`; a Python gate now calls
  `helpers.gate_files(gate, *args)` of `helpers.py`. D5 kept the gate files
  and the orchestrators in bash, with Python only for the file logic; the
  gates and orchestrators now become Python too (plan 2).
- The `UV_NO_CACHE` setting ends (D6).
- The gate `code:format:40-ruff` ends (D13).

Rejected: keep the shell rules beside the Python rules.

## The plans (D1, D2, D14 to D17, D20)

- **Three chained plans (D1).** User: "Yes, three plans". Plan 1 (this one):
  the Python libraries beside the shell libraries, the gate fixes, the rules
  and this record. Plan 2, `docs/plans/2026-11-18-python-tasks-repo`: the
  shipped tasks of `repo/tools/`, the own tasks of this repository, the
  deletion of the shell libraries and the `migrate.md` item M48. Plan 3,
  `docs/plans/2026-11-18-python-tasks-mac`: the Mac tasks and the Mac
  helpers. Rejected: one plan (too large to review); four plans with a
  separate test plan.
- **Only the mise tasks (D2).** User: "Let's focus on `mise tasks` only for
  now". The engine scripts of `repo/tools/*/scripts/` stay bash: the gap
  record `docs/memory/gaps/2026-11-18-engine-scripts-python-parked.md`.
  Rejected: convert the engine scripts too.
- **No release (D14).** User: "Yes, accept both". No engine release and no
  tag for B30: the tasks reach a consumer from the clone at
  `bootstrap_commit`. `cli/internal/oldform` needs no change. The false
  warning of `bootstrap migrate` for an own unnamed `.py` file that another
  file names in `# MISE depends` is accepted. Rejected: `bootstrap@4.2.0`;
  one release for each plan.
- **The gates of the plan (D15).** User: "Yes, this smoke check" and "Yes,
  three plans". The plan lands only when `mise run test`,
  `code:format --all`, `code:precommit` and `code:sec` pass, plus the smoke
  check K1. The wave gate is `code:lint --all`. Rejected: land without the
  test suite.
- **After the landing (D16).** User: "yes, sync after each plan lands". The
  run does the self-sync on a local branch `sync-python-tasks-foundation`, and
  archives the folder only with no open gap. Rejected: the user syncs by hand.
- **The run settings (D17).** User: "Yes, accept". Model opus for each unit,
  one unit for each library group, one review unit (D20) before the docs
  unit, execution teams. Rejected: solo.

## The form of a Python task (D4, D5, D7, D8, D11, D21, D29)

- **The file (D4).** User: "yes". Each Python task and library file has the
  `.py` extension. mise removes `.py` from the task name; `_default.py` is the
  task of its folder. The form is the form of `gate_files.py`: the shebang
  `#!/usr/bin/env -S uv run --script`, the `Synced from bootstrap` header on
  lines 2-3, the `# /// script` block with `requires-python = ">=3.14"` and
  `dependencies = []`, the `# MISE` lines with
  `# MISE tools={uv="latest", python="3.14"}` (a library also
  `# MISE hide=true` and a description), then the `# USAGE` lines. Rejected:
  no extension (dprint and the Python gates select by extension).
- **The import preamble (D5, D21).** A task adds
  `${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts` to `sys.path`, then
  `import helpers  # noqa: E402`. The Mac tasks use
  `~/.config/mise/tasks/_scripts`. The preamble sets
  `sys.dont_write_bytecode = True` before the first import of a library, so no
  `__pycache__` folder appears below `.config/mise/tasks/` (the engine refuses
  a file there with no entry). Each new file gets mode 755, because the hook
  `check-shebang-scripts-are-executable` refuses a file with a shebang and no
  execute bit. Rejected: `PYTHONPATH` in the mise conf.d; a path dependency
  through `[tool.uv.sources]`; delete `__pycache__` by hand.
- **The standard library only (D7).** User: "yes and yes for Q5 as well".
  `dependencies = []`. A third-party package needs a plan that names it and a
  `stackgen-reputation` check that passes. Rejected: a fixed set (rich, a YAML
  parser); no rule.
- **The flags (D8).** User: "Yes, keep # USAGE". mise parses the flags and
  gives each one as the environment variable `usage_<flag>`. ruff format (and
  dprint) writes `# USAGE` with a space, so each search for a USAGE line
  accepts both forms. Rejected: argparse (no `--help` and no completions from
  mise; each flag in two places).
- **The mise x rule (D11).** User: "Yes, both". `helpers.tool(*args,
  **kwargs)` runs `mise x -- <args>`; `helpers.has_tool(name)` runs
  `mise x -- which <name>`. System commands run with `subprocess` directly.
  Each yq call of the tasks only parses JSON, so it becomes `json` in plan 2.
  The tool exception stays for `github:virajp/bootstrap`, `pre-commit` and
  `dprint`; those tasks find the tool with `shutil.which`. The engine call
  keeps `BOOTSTRAP_BIN`, else `mise x -- bootstrap`. Rejected: build the
  prefix by hand in each task; keep yq.
- **One tools line (D29).** TOML does not accept an inline table and a dotted
  key for the same table, so a Python task has one line, for example
  `# MISE tools={uv="latest", python="3.14", pre-commit="latest"}`. Rejected:
  two lines.

## The libraries (D9, D10, D12, D18, D19, D22, D23, D26, D27)

- **helpers.py (D9).** User: "Yes, same names, bytes". The names of the shell
  helpers, the same bytes, colors, symbols, widths, and the same rule for
  color off as `style_enabled`. Rejected: a new output API.
- **bootstrap_get and members (D10).** User: "Yes, value and sys.exit".
  `bootstrap_get(key, repo=None)` returns the parsed JSON answer; the caller
  reads `["value"]`. A read failure calls `sys.exit(1)`; no engine that
  answers, or a refusal with code 2, gives the same lines as the shell and
  `sys.exit(2)`. `members(repo=None)` returns a list of
  `(name, folder, label, url)` tuples. Rejected: raise an exception (a
  traceback exits 1, which breaks the code 2 contract).
- **The API details (D22).** `bold=True` replaces the leading `--bold`;
  `file=sys.stderr` replaces `>&2`; each print function flushes. A missing
  `BOOTSTRAP_BIN`, an engine code other than 0 and 1, and a code-0 answer
  that is not JSON take the code 2 path. Rejected: a leading `"--bold"`
  string; no flush; a traceback on a bad engine.
- **gate_files (D12).** `helpers.gate_files(gate, *args)` runs
  `gate_files.py` with `sys.executable` and returns the path of the list file;
  a failure stops the task with the code of the script. `gate_files.py`
  accepts `# FILES` as well as `#FILES`. Rejected: import `gate_files.py` into
  `helpers.py`.
- **A .py part (D18, D22).** `own_tasks`, `has_own_task` and the `--all`
  search of `code:format` and `code:lint` follow mise: a part `NN-x.py` is the
  task `NN-x` when no `NN-x` exists, else `NN-x.py`; `has_own_task` and the
  `--all` search read `x` when it exists, else `x.py`. The shell and the
  Python libraries do the same. Rejected: change the orchestrators only in
  plan 2.
- **xcode.py (D19).** The functions of the shell `xcode`, with the same
  names, messages and exit codes. `require_xcode(task)` returns a dict of
  `XCODE_VERSION` and the `SIMULATOR_*` names, and sets them in `os.environ`.
  `find_xcodeproj(task)` returns the path. Rejected: return values only.
- **write_worktrees_page (D23).** `write_worktrees_page(rows, attention, now,
  folder="/tmp/code-worktrees")`: a list of row dicts, a list of attention
  dicts in rank order, and the time text. It keeps each check, the file name,
  the temp file and the rename, prints the path and returns it; on a failure
  it prints nothing and returns `None`. The glyphs, `LEGEND`, `count_cell`
  and `state_parts` are module members. Rejected: global state in the module.
- **usage_files (D26).** `helpers.usage_files()` returns
  `shlex.split(os.environ.get("usage_files", ""))`, the same quoting as the
  shell `eval`. Rejected: `str.split`.
- **The placeholder marker (D27).** dprint writes `# PLACEHOLDER` in a `.py`
  file, so `placeholder.py` accepts both forms, and lists task names without
  `.py`. Rejected: one form only.
- **Run-time answers (D31 to D33).** `worktrees_page.py` keeps the output of
  `/bin/bash` 3.2 (correct escaping and `~` paths), not of bash 5.2 (D31).
  `print_line` of `helpers.py` treats the `\u` and `\U` escapes as bash 5.x,
  the bash that `#!/usr/bin/env bash` runs on this Mac (D32). Outside a git
  repository `checks.py` prints the git fatal line fewer times than the shell
  on stderr; the difference is accepted, because `checks` prints nothing by
  contract (D33).

## Format, lint and the uv cache (D6, D13, D24, D25, D28)

- **The uv cache (D6).** User: "how about enabling cache with `uv` ?", then
  "yes". `UV_NO_CACHE = true` leaves `[mise.env]` of the `python` tool and of
  the Mac `uv` tool. uv keeps one environment for each script path. Rejected:
  keep the cache off.
- **dprint only (D13).** User: "remove ruff formatter, only use dprint", and
  "pyrefly and ruff will only go if the repo is building something with
  python, not just for mise task files". dprint (its ruff plugin,
  `indentWidth` 2, `lineLength` 120) formats each `.py` file. The gate
  `code:format:40-ruff` is gone. ruff (lint) and pyrefly (types) keep
  `requires_found = ["pyproject.toml"]`, so this repository has no gate that
  lints its own `.py` task files. Rejected: ruff format beside dprint; ruff
  lint in every repository.
- **The task files and the Python gates (D28).** The `# FILES` header of
  `code:lint:70-ruff` and `code:lint:50-pyrefly` gets
  `exclude=".config/mise/tasks/**"`, because pyrefly cannot resolve
  `import helpers`. Rejected: suppress the finding; accept the failure.
- **A consumer override of UV_NO_CACHE (D24).** An override cannot add a key,
  so an `[overrides.mise.env] UV_NO_CACHE` now stops the sync, also on the
  Mac. `migrate.md` item M46, and a step in the update part of
  `docs/setup/index.md`. Rejected: no migrate item.
- **A consumer with ruff format settings (D25).** `migrate.md` item M47: move
  the format settings of `[tool.ruff]` and `[tool.ruff.format]` to the block
  `ruff` of `dprint.jsonc`, under the names of the dprint ruff plugin
  (`lineLength`, `indentWidth`, `indentStyle`, `lineEnding`, `quoteStyle`,
  `skipMagicTrailingComma`, `docstringCodeFormat`, `docstringCodeLineWidth`,
  `preview`, `targetVersion`). Rejected: no migrate item.

## The tests (D21, D30)

- `e2e/repo/pylib` compares each Python module with its shell library: the
  output bytes, the exit code and the return values.
- Two merge runs (D30) fix the git author, committer and dates, so the two
  logs match with their hashes. Rejected: compare the subjects only.
