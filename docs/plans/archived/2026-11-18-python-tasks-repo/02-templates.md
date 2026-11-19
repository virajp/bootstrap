# U2 — The orchestrator templates in Python

- **Read first:** each owned file, top to bottom, and the Python
  libraries that it will import (`repo/tools/mise/tasks/helpers.py` and the
  others of the foundation plan).
- **Lazy-load:** `repo/tools/mise/files/` (how a template reads `.config` and
  `.all`), `e2e/repo/tasks/configs_test.go`, `e2e/repo/tasks/gate_test.go`,
  CLAUDE.md "Task scripts" and "Task output".

## Ruling

- **D1:** "The rulings D4, D5, D7, D8, D9, D10, D11, D12, D13 and D18 of
  docs/plans/2026-11-18-python-tasks-foundation apply to each task of this
  plan, as CLAUDE.md \"Task scripts\" states them after that plan."
- **D2:** "A converted task keeps its task name, its # USAGE flags and help
  texts, its #FILES header (as # FILES), its output [...], its exit codes and
  the order of its steps. Its # MISE lines stay, except: the #MISE tools.yq
  lines go (D4), and the tools line becomes one inline table with uv, python
  and the tools of its exception (foundation D29). [...] A behavior change is
  a QUESTION, never a silent edit."
- **D3:** "The key of a task gets .py [...] and source names the .py file. A
  template keeps mode = \"template\" and its source becomes <name>.py.tmpl.
  Each other key of the entry (requires_found, when, envs) stays."
- **D4:** each yq or jq call becomes `json`; the tool exception stays only
  for `github:virajp/bootstrap`, `pre-commit` and `dprint`, found with
  `shutil.which`. A gate reads its `[files]` with `helpers.usage_files()`.
- **D5:** "The engine renders a template with the delimiters <% and %> [...];
  {{ }} passes through unchanged [...]. So the Python of a template must not
  contain <% or %> outside a gomplate part. The --all search of foundation D18
  moves into the Python of code-format and code-lint. U2 checks a render with
  the engine: build it, sync a scratch consumer with --clone <worktree>, then
  compile the rendered files."
- **D7:** "U3 owns repo/tools/mise/tools.toml in wave 1 and writes the renamed
  keys of U1 (code/worktrees) and U2 (the five templates) too."

- **D14:** "Foundation D21 applies: mode 755 on each new file,
  sys.dont_write_bytecode = True in the preamble, python -B in a hand check,
  and a Go test of the engine from cli/. A unit checks the inventory of the
  tools.toml files with the engine: cd cli && BOOTSTRAP_CLONE=<worktree root>
  mise x -- go run ./cmd/bootstrap catalog exits 0."

- **D14 (order):** "U1 and U2 skip [the catalog check], because their keys
  are in the file of U3, and U3 runs it after U1 and U2 return."

- **D16:** "U2, U4, U5, U6 and U7 skip the catalog check and the render
  check. U3 runs both after all the units of wave 1 return (the team tells
  it) [...]."

## Edits

1. `code-format.tmpl`, `code-lint.tmpl`, `setup.tmpl`, `setup-configs.tmpl` and
   `setup-worktree.tmpl` become `<name>.py.tmpl` (D3, D5). Delete each old file
   with `rm`.
2. `code-format` and `code-lint` run each part from `helpers.own_tasks`, pass
   `--all` only to a part whose file has `# USAGE flag "--all"` or `#USAGE flag
   "--all"` (foundation D18), and keep the exit codes of the shell version (exit
   2 for `--all` with files).
3. `setup-configs` keeps the tool exception of `github:virajp/bootstrap` and
   finds the engine with `shutil.which`; its yq call and the one of `setup`
   become `json` (D4).
4. Do not sync a scratch consumer here (D16): U3 renders the templates and
   checks them after all the units of wave 1 return. Tell U3 the names of
   the five `.py.tmpl` files through the team.

## Verification

- The syntax check `python3.14 -B -c "import ast,sys; [ast.parse(open(p).read(),
  p) for p in sys.argv[1:]]" <files>`
  passes on each new `.py` file (render a template first). `py_compile`
  writes `__pycache__`, so it is not used.
- No ruff check here: a `.py.tmpl` source does not parse before a render,
  and U3 checks the rendered files (D16).
- The `# USAGE` lines of each new file have the same flags and help texts as
  the `#USAGE` lines of the old file (compare them with `grep`).
- No new file contains `shell=True`, `os.system`, `yq` or `jq` as a command.
- No catalog check here (D14): U3 runs it after this unit returns.
- Each new file has mode 755, and `find repo -name __pycache__` lists
  nothing.

## Guardrails

- Do not edit a file below `e2e/`: U8 updates the Go cases.
- Do not edit a shell library or its `_scripts/*` entry: U9 owns them.
- Delete with `rm`, never `git rm`.
- Write each comment in ASD-STE100. Indent with 2 spaces; lines to 120
  columns.
- Do not edit `repo/tools/mise/tools.toml`: tell U3 the new keys and sources
  (D7).

## Commit

`refactor: the orchestrator templates in Python`
