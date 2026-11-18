# U2 — The tasks of the other Mac tools in Python

- **Read first:** each task file of the owned tool folders, top to bottom,
  and `macos/tools/mise/tasks/helpers.py` (from U1).
- **Lazy-load:** `docs/setup/tools.md` (the section of each tool).

## Ruling

- **D1:** the rules of the foundation and repo plans apply, with the Mac
  import preamble.
- **D3:** "A task that ends with exec in shell ends with os.execvp with the
  same arguments, so the process of the tool replaces the Python process.
  uv run --script can stay as the parent process of Python; that is
  accepted."
- **D4:** a jq or yq call that parses JSON becomes `json`, and its
  `# MISE tools` line for jq or yq goes.
- **D5:** "U2 deletes macos/tools/mise/tasks/helpers with rm after it
  converts the last task that sources it, and removes its
  [tasks.\"_scripts/helpers\"] entry."
- **D12:** "Each Mac task keeps its current call form: mise x -- for a tool
  of mise, shutil.which for a command that it finds with command -v now
  (claude, code, fnox, brew, mas, security)."
- **D13:** "dprint formats every .py file [...], so U2 adds the plugin
  npm:@dprint/ruff@0.11.3 [...] to the plugins of the Mac dprint config, a
  [defaults.blocks.ruff] with indentWidth = 2 and lineLength = 120 [...], and
  \"!**/*.py\" to the shuck associations."
- **D15:** "A task that starts a tool for the user (func-claude and func-myps
  with os.execvp, func-code with VS Code) gives the tool an environment
  without VIRTUAL_ENV and without the bin folder of sys.prefix on PATH [...].
  The uv and Python folders that # MISE tools adds stay."
- **D16:** the syntax check writes no file: `ast.parse`, not `py_compile`.
- **D10:** each task has `# MISE tools={uv="latest", python="3.14"}`.

- **D14:** "Foundation D21 applies: mode 755 on each new file,
  sys.dont_write_bytecode = True in the import preamble before the first
  import of helpers, python -B in a hand check, and a Go test of the engine
  from cli/."

## Edits

1. Each task file of
   `macos/tools/{claude-cli,cli,homebrew,network,node,pstree,tailscale,telegram}/tasks/`
   becomes a `.py` file. Delete each old file with `rm`.
2. Each `tools.toml` of the eight tools: rename the task keys (repo plan D3).
3. **`macos/tools/dprint/tools.toml`** — D13: add `npm:@dprint/ruff@0.11.3`
   to the plugins; add `[defaults.blocks.ruff]` with `indentWidth = 2` and
   `lineLength = 120`; the shuck associations become
   `[ "**/.config/mise/tasks/**", "!**/*.tmpl", "!**/*.py" ]`; the comment
   says that a `.py` task goes to the ruff plugin; the header comment of the
   file (lines 8-11, "the base formats only") also names Python.
4. `grep -rl '_scripts/helpers"' macos/` finds no task; then
   `rm macos/tools/mise/tasks/helpers` and remove its entry from
   `macos/tools/mise/tools.toml`.

## Verification

- Each new file has mode 755, and `find macos -name __pycache__` lists
  nothing.
- The syntax check `python3.14 -B -c "import ast,sys; [ast.parse(open(p).read(),
  p) for p in sys.argv[1:]]" <files>` and `mise x ruff@latest -- ruff check`
  pass on
  each new file (D16).
- The `# USAGE` lines of each new file have the same flags and help texts as
  the old `#USAGE` lines.
- No new file contains `shell=True`, `os.system`, `jq` or `yq` as a command.
- `cd cli && mise x -- go test ./internal/source/...` passes.

## Guardrails

- Do not edit `macos/tools/mise/tasks/helpers.py` or a task of the Mac mise
  tool: U1 owns them.
- Do not edit a file below `e2e/`.
- Delete with `rm`, never `git rm`.
- Write each comment in ASD-STE100.

## Commit

`refactor: the tasks of the Mac tools in Python, and delete the shell Mac
helpers`
