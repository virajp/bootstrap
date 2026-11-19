# U7 — The setup, format and security tasks of the other tools in Python

- **Read first:** each owned file, top to bottom, and the Python
  libraries that it will import (`repo/tools/mise/tasks/helpers.py` and the
  others of the foundation plan).
- **Lazy-load:** `e2e/repo/commands/tasks_test.go` (`TestSetupAI`),
  `docs/sync/tools/*.md` of the ten tools, CLAUDE.md "Task scripts" and "Task
  output".

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

- **D14:** "Foundation D21 applies: mode 755 on each new file,
  sys.dont_write_bytecode = True in the preamble, python -B in a hand check,
  and a Go test of the engine from cli/. A unit checks the inventory of the
  tools.toml files with the engine: cd cli && BOOTSTRAP_CLONE=<worktree root>
  mise x -- go run ./cmd/bootstrap catalog exits 0."

- **D16:** "U2, U4, U5, U6 and U7 skip the catalog check and the render
  check. U3 runs both after all the units of wave 1 return (the team tells
  it) [...]."

## Edits

1. Each task file of
   `repo/tools/{claude,vscode,dprint,linter,pre-commit,gitleaks,grype,osv-scanner,doppler,fnox}/tasks/`
   becomes a `.py` file. Delete each old file with `rm`.
2. `setup-dprint` keeps `dprint`, and `setup-precommit` and `code-precommit`
   keep `pre-commit`, in the one tools line, for example
   `# MISE tools={uv="latest", python="3.14", pre-commit="latest"}`. Each finds
   its tool with `shutil.which` (D4).
3. The `code-sec-*` scanners print only a `print_subheader` ("Task output" rule
   10); with `--staged`, a scanner without a staged mode prints nothing.
4. `code-lint-shellcheck` stays the ShellCheck gate of the shell files of a
   consumer; only its own file becomes Python.
5. Each `tools.toml` of the ten tools: rename the task keys per D3.

## Verification

- The syntax check `python3.14 -B -c "import ast,sys; [ast.parse(open(p).read(),
  p) for p in sys.argv[1:]]" <files>`
  passes on each new `.py` file (render a template first). `py_compile`
  writes `__pycache__`, so it is not used.
- `mise x ruff@latest -- ruff check` passes on each new `.py` file (foundation
  D21: no ruff is installed in this repository).
- The `# USAGE` lines of each new file have the same flags and help texts as
  the `#USAGE` lines of the old file (compare them with `grep`).
- No new file contains `shell=True`, `os.system`, `yq` or `jq` as a command.
- No catalog check here (D16): U3 runs it after all the units of wave 1
  return.
- Each new file has mode 755, and `find repo -name __pycache__` lists
  nothing.

## Guardrails

- Do not edit a file below `e2e/`: U8 updates the Go cases.
- Do not edit a shell library or its `_scripts/*` entry: U9 owns them.
- Delete with `rm`, never `git rm`.
- Write each comment in ASD-STE100. Indent with 2 spaces; lines to 120
  columns.

## Commit

`refactor: the setup, format and security tasks in Python`
