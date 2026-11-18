# U1 — The Mac helpers and the Mac mise tasks in Python

- **Read first:** each owned file, top to bottom, and
  `repo/tools/mise/tasks/helpers.py` (the form of a Python helpers module).
- **Lazy-load:** `docs/design/mac.md` (the Mac sync and its entry task),
  `e2e/mac/commands_test.go` (the entry cases), CLAUDE.md "Task scripts".

## Ruling

- **D1:** the rules of the foundation and repo plans apply; the import
  preamble is
  `sys.path.insert(0, os.path.expanduser("~/.config/mise/tasks/_scripts"))`.
- **D2:** "macos/tools/mise/tasks/helpers.py keeps the names of the shell Mac
  helpers [...] and prints the same bytes. It stays a module of its own."
- **D10:** each task has `# MISE tools={uv="latest", python="3.14"}`.
- **D10:** "uv becomes a required member of the Mac base pack: uv = { required =
  true } in macos/packs.toml [...]. [...] the Mac mise tool gets depends =
  [\"cli\", \"python\", \"uv\"]: then both are in the init stage, and its mise
  bootstrap step installs them before the run asks for a new terminal."
- **D11:** "It keeps what it does now: it stays standalone and imports no
  helpers [...]; it reads its flags from sys.argv, never from usage_*[...];
  it keeps its usage and -h text, legacy_cleanup, the nested-run guard
  (_UPDATEALL_CONFIGS_RUNNING), the unset of the GIT_* variables, and
  BOOTSTRAP_CACHE, XDG_CACHE_HOME, FORCE_COLOR and NO_COLOR; it calls
  BOOTSTRAP_BIN, else bootstrap from PATH, as now. A direct run (outside mise
  run) needs uv on PATH for the shebang."
- **D16:** the syntax check writes no file: `ast.parse`, not `py_compile`.
- **D12:** "The Mac helpers.py has no tool or has_tool. Each Mac task keeps
  its current call form: mise x -- for a tool of mise (dprint in updateall),
  shutil.which for a command that it finds with command -v now."
- Repo plan D3: a task key gets `.py`, and `source` names the `.py` file.

- **D14:** "Foundation D21 applies: mode 755 on each new file,
  sys.dont_write_bytecode = True in the import preamble before the first
  import of helpers, python -B in a hand check, and a Go test of the engine
  from cli/."

## Edits

1. **`macos/tools/mise/tasks/helpers.py`** — the module of D2, in the form of
   a library (`# MISE hide=true`, a description).
2. `updateall`, `updateall-configs`, `upgrade-macos-download`,
   `upgrade-macos-upgrade` and `upgrade-power` become `.py` files. Delete
   each old file with `rm`. `updateall-configs` follows D11: it stays
   standalone and imports no helpers.
3. **`macos/tools/mise/tools.toml`** — add
   `[tasks."_scripts/helpers.py"]` with `library = true`, and rename the five
   task keys. Keep `[tasks."_scripts/helpers"]`: U2 removes it in wave 2.
4. **`macos/packs.toml`** — D10: `uv = { required = true }` in the base
   pack, and its `description` names uv among the tools that a Mac cannot
   remove.
5. **`macos/tools/mise/tools.toml`** — D10: `depends = [ "cli", "python",
   "uv" ]`, and the comment above it says why: the entry tasks are Python,
   so uv and Python must be installed in the init stage.

## Verification

- Each new file has mode 755, and `find macos -name __pycache__` lists
  nothing.
- The syntax check `python3.14 -B -c "import ast,sys; [ast.parse(open(p).read(),
  p) for p in sys.argv[1:]]" <files>` and `mise x ruff@latest -- ruff check`
  pass on
  each new file (D16).
- The bytes of each print function of `helpers.py` equal the bytes of the
  shell Mac helpers (a hand check: `bash -c 'source
  macos/tools/mise/tasks/helpers; print_header x'`).
- `cd cli && mise x -- go test ./internal/source/... ./internal/macos/...`
  passes.

## Guardrails

- Do not edit a file below `e2e/`: U3 updates the Go cases.
- Do not delete the shell Mac helpers: the tasks of the other Mac tools
  still source it until U2 converts them.
- Delete with `rm`, never `git rm`. No `shell=True`.
- Write each comment in ASD-STE100.

## Commit

`refactor: the Mac helpers and the Mac mise tasks in Python`
