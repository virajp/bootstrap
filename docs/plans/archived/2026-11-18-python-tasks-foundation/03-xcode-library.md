# U3 — The Python xcode library of the swift tool

- **Read first:** `repo/tools/swift/tasks/xcode` and
  `repo/tools/swift/tools.toml`, top to bottom; the swift tasks that source
  `xcode`.
- **Lazy-load:** `repo/tools/mise/tasks/gate_files.py` (the form),
  `docs/sync/tools/swift.md` (what the library does).

## Ruling

- **D4, D5, D7:** the form of `gate_files.py`, the import preamble, the
  standard library only. The module name is `xcode`.
- **D19:** "xcode.py has the functions of the shell xcode (require_xcode,
  find_xcodeproj and the simulator functions) with the same names, messages and
  exit codes [...]. require_xcode(task) returns a dict of the names that the
  shell exports (XCODE_VERSION and the SIMULATOR_*names) to their values, and
  also sets each name in os.environ [...]. find_xcodeproj(task) returns the
  path of the one*.xcodeproj as a str, where the shell prints it [...]. The
  plan 2 swift tasks use these return values."
- **D11:** `xcodebuild`, `xcrun` and `swift` are system commands: run them
  with `subprocess` directly, with no `mise x --`.
- **D9:** each line that `xcode` prints comes from the same `helpers` print
  function with the same text.

- **D21:** "Each new task or library file gets mode 755 (chmod 755) [...].
  The import preamble sets sys.dont_write_bytecode = True before the first
  import of a library [...]. The syntax check is python3.14 -B -c \"import
  ast,sys; [ast.parse(open(p).read(), p) for p in sys.argv[1:]]\" <files>;
  py_compile writes **pycache**. This repository has no pyproject.toml, so
  after D13 no gate lints its .py task files (accepted by the user); each unit
  lints its own files by hand with mise x ruff@latest -- ruff check. A Go test
  of the engine runs from cli/."

## Edits

1. **`repo/tools/swift/tasks/xcode.py`** — the functions of D19, with the
   same lookups, the same messages and the same exit codes as the shell
   version.
2. **`repo/tools/swift/tools.toml`** — add `[tasks."_scripts/xcode.py"]` with
   `library = true`, in the form of the `_scripts/xcode` entry. Keep the shell
   entry.

## Verification

- `chmod 755` on the new file (D21), and no `__pycache__` folder below
  `repo/tools/` after the import (use `python -B`).
- An import of `xcode` with `repo/tools/mise/tasks/` and
  `repo/tools/swift/tasks/` on `sys.path` exits 0.
- `mise x ruff@latest -- ruff check repo/tools/swift/tasks/xcode.py` passes.
- On this Mac, `require_xcode()` and `find_xcodeproj()` give the same result
  as the shell functions in a scratch folder with no Xcode project.

## Guardrails

- Do not edit `xcode` or a swift task: plan 2 converts them.
- No `shell=True`.
- Write each comment in ASD-STE100.

## Commit

`refactor: add the Python xcode library xcode.py`
