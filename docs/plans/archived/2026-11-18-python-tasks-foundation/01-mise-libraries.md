# U1 — The Python libraries of the mise tool

- **Read first:** `repo/tools/mise/tasks/helpers`, `checks`, `placeholder`,
  `worktrees-page` and `gate_files.py`, top to bottom, and
  `repo/tools/mise/tools.toml`.
- **Lazy-load:** `docs/design/tools.md` ("The `[tasks]` table", "The header
  of a source file"), `docs/design/cli.md` ("The JSON contract", for the
  answer of `bootstrap config get`).

## Ruling

- **D4:** each module has the form of `gate_files.py`: the shebang
  `#!/usr/bin/env -S uv run --script`, the Synced from bootstrap header on
  lines 2-3 (with its own source path), the `# /// script` block with
  `requires-python = ">=3.14"` and `dependencies = []`, then
  `# MISE description="..."`, `# MISE hide=true` and
  `# MISE tools={uv="latest", python="3.14"}`.
- **D5:** a task imports the modules after
  `sys.path.insert(0,
  f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")`.
  The module names are `helpers`, `checks`, `placeholder` and
  `worktrees_page`.
- **D7:** the standard library only.
- **D9:** "helpers.py keeps the names of the shell helpers and prints the same
  bytes: the same colors, symbols, widths and the same rule for color off as
  style_enabled."
- **D10:** `bootstrap_get(key, repo=None)` returns the parsed JSON answer as a
  dict. Engine code 1: `sys.exit(1)` after the message of the engine. No
  engine that answers, or engine code 2: the same `print_error` and
  `print_warn` lines as the shell version, then `sys.exit(2)`. With `repo`,
  it runs the engine in that folder with `MISE_PROJECT_ROOT=<repo>` and
  without `BOOTSTRAP_VALUES`, as the shell version does. `members(repo=None)`
  returns a list of `(name, folder, label, url)` tuples with the same filter,
  order and defaults as the shell version, parsed with `json`.
- **D11:** `tool(*args, **kwargs)` runs
  `subprocess.run(["mise", "x", "--", *args], **kwargs)` and returns the
  result. `has_tool(name)` runs `mise x -- which <name>` with no output and
  returns a bool. The engine is `BOOTSTRAP_BIN` when it is set, else
  `mise x -- bootstrap`; with no `BOOTSTRAP_BIN`, `has_tool("bootstrap")`
  decides "no engine".
- **D12:** `gate_files(gate, *args)` runs
  `[sys.executable, <_scripts>/gate_files.py, gate, *args]`, returns the path
  of the list file (the stdout line), and on a failure prints the same
  `print_error` line as the shell version and calls `sys.exit` with the code
  of the script. The caller deletes the list file.
- **D18, D22:** `own_tasks(group)` lists each executable regular file named
  `NN-<name>` or `NN-<name>.py` in the folder of the group, sorted, with the
  task name that mise gives it: `NN-<name>` for a `.py` file when no file
  without the extension has that name, else the file name. `has_own_task(name)`
  and the `--all` search read the file that mise runs for the name: `<path>`
  when it exists, else `<path>.py` (and `_default`, else `_default.py`).
- **S2:** this unit owns `repo/tools/mise/tools.toml` in this plan.

- **D21:** "Each new task or library file gets mode 755 (chmod 755) [...].
  The import preamble sets sys.dont_write_bytecode = True before the first
  import of a library [...]. The syntax check is python3.14 -B -c \"import
  ast,sys; [ast.parse(open(p).read(), p) for p in sys.argv[1:]]\" <files>;
  py_compile writes **pycache**. This repository has no pyproject.toml, so
  after D13 no gate lints its .py task files (accepted by the user); each unit
  lints its own files by hand with mise x ruff@latest -- ruff check. A Go test
  of the engine runs from cli/."
- **D22:** "print_success, print_warn, print_error and print_info take
  bold=True as a keyword where the shell takes a leading --bold. Each print
  function takes file=sys.stderr as a keyword (default sys.stdout) where a
  shell call site adds >&2 [...]. Each print function flushes its stream
  (flush=True) [...]. bootstrap_get: a missing or not executable
  BOOTSTRAP_BIN (FileNotFoundError, PermissionError), an engine code other
  than 0 and 1, and a code-0 answer that is not JSON all take the code 2 path
  [...]. own_tasks, has_own_task and the --all search name and read a file as
  mise does (tested 2026-11-18): mise removes .py from a task name only when no
  file without the extension has that name; when x and x.py both exist, mise
  lists x and x.py and runs x for the name x. [...] The libraries of_scripts/
  have both forms between plan 1 and plan 2; they are hidden and no task runs
  them by name, so that is accepted."
- **D23:** "worktrees_page.py gets write_worktrees_page(rows, attention, now,
  folder=\"/tmp/code-worktrees\") [...]. It keeps every check, the file name,
  the temp file and the rename, prints the path, and returns the path; on a
  failure it prints nothing and returns None, as the shell does. The G_* glyph
  constants and LEGEND
  move into worktrees_page.py as module constants with the same bytes;
  count_cell and state_parts move there too. U5 runs both sides with a scratch
  folder [the shell side as a copy of the function with its folder line
  changed to the scratch folder; TMPDIR does not move it], so parallel cases
  never share /tmp/code-worktrees, and it
  compares the HTML with the date line of both pages replaced by the same
  fixed text."
- **D27:** "placeholder.py accepts both forms [`#PLACEHOLDER`
  and `# PLACEHOLDER`], and lists task names without .py, as mise names them."

- **D26:** "helpers.py gets usage_files(), which returns
  shlex.split(os.environ.get(\"usage_files\", \"\")), so a Python gate keeps the
  same quoting for an odd file name."

## Edits

1. **`repo/tools/mise/tasks/helpers.py`** — the module of D9, D10, D11, D12,
   D18, D22 and D26. The print functions take the text as their arguments and
   write to
   stdout, or to stderr where the shell version uses `>&2`. A docstring on
   each function says what the shell comment of the same function says.
2. **`repo/tools/mise/tasks/checks.py`** — the git predicates of `checks`
   with the same names. Each returns a bool and prints nothing, except
   `get_main_worktree_path`, which returns the path.
3. **`repo/tools/mise/tasks/placeholder.py`** — the functions of
   `placeholder` with the same names and output; the search accepts
   `#PLACEHOLDER` and `# PLACEHOLDER` and lists task names without `.py`
   (D27).
4. **`repo/tools/mise/tasks/worktrees_page.py`** — `write_worktrees_page` of
   `worktrees-page` per D23: the signature
   `write_worktrees_page(rows, attention, now, folder)`, the checks of the
   folder, the file name, the temp file and the rename; the module constants
   `G_*` and `LEGEND`; `count_cell` and `state_parts` of `code-worktrees`;
   the same HTML bytes. Copy the
   embedded font data byte for byte; do not type it again.
5. **`repo/tools/mise/tools.toml`** — add `[tasks."_scripts/helpers.py"]`,
   `[tasks."_scripts/checks.py"]`, `[tasks."_scripts/placeholder.py"]` and
   `[tasks."_scripts/worktrees_page.py"]`, each with `library = true` and a
   `source` in the form of the entry of `_scripts/gate_files.py`. Keep the
   shell entries: plan 2 removes them.

## Verification

- `chmod 755` on each new file (D21).
- `PYTHONDONTWRITEBYTECODE=1 mise x -- uv run --quiet --no-project --python 3.14
  python -B -c`
  with an import of each module from `repo/tools/mise/tasks/` exits 0, and no
  `__pycache__` folder exists below `repo/tools/`.
- For each print function: the bytes of the Python call equal the bytes of
  the shell call (`bash -c 'source repo/tools/mise/tasks/helpers; print_header
  x'`),
  with color on and with color off (`NO_COLOR=1`, and a pipe). U5 writes the
  Go cases; this check is a hand check of the unit.
- `mise x ruff@latest -- ruff check repo/tools/mise/tasks/*.py` passes.
- `cd cli && mise x -- go test ./internal/source/...` passes (the library rule).

## Guardrails

- Do not edit the shell libraries, `gate_files.py`, the templates or any
  task: U4 and plan 2 own them.
- Write each comment in ASD-STE100.
- Indent with 2 spaces and keep lines to 120 columns (dprint ruff).
- No `shell=True` in a `subprocess` call.

## Commit

`refactor: add the Python libraries helpers.py, checks.py, placeholder.py and
worktrees_page.py`
