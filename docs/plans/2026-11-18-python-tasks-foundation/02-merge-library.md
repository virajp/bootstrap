# U2 — The Python merge library of the git tool

- **Read first:** `repo/tools/git/tasks/merge` and
  `repo/tools/git/tools.toml`, top to bottom; the
  `repo/tools/git/tasks/git-merge-*` tasks that source `merge`.
- **Lazy-load:** `docs/design/repo.md` (the landing modes and the merge
  rules), `repo/tools/mise/tasks/gate_files.py` (the form).

## Ruling

- **D4, D5, D7:** the form of `gate_files.py`, the import preamble, the
  standard library only. The module name is `merge`.
- **D10:** a value of `.config/bootstrap/` comes from
  `helpers.bootstrap_get(key)["value"]`.
- **D11:** each `yq` call of `merge` only parses JSON: use `json`. `git` runs
  with `subprocess` directly. `pre-commit` keeps the tool exception: find it
  with `shutil.which`.
- **D9:** each line that `merge` prints comes from the same `helpers` print
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

1. **`repo/tools/git/tasks/merge.py`** — `merge_to_destination_branch` with
   the same arguments, the same landing modes, the same git commands in the
   same order, the same output and the same exit codes as the shell version.
   The shell `set -euo pipefail` becomes a check of each `subprocess` result:
   a failed command stops the merge with the code that the shell version
   gives.
2. **`repo/tools/git/tools.toml`** — add `[tasks."_scripts/merge.py"]` with
   `library = true`, in the form of the `_scripts/merge` entry. Keep the shell
   entry.

## Verification

- `chmod 755` on the new file (D21), and no `__pycache__` folder below
  `repo/tools/` after the import (use `python -B`).
- An import of `merge` with `repo/tools/mise/tasks/` and
  `repo/tools/git/tasks/` on `sys.path` exits 0. U1 writes `helpers.py` in
  the same wave; if it is not there yet, ask U1 through the team, or test the
  import after U1 returns.
- `mise x ruff@latest -- ruff check repo/tools/git/tasks/merge.py` passes.

## Guardrails

- Do not edit `merge` or a `git-merge-*` task: plan 2 converts them.
- No `shell=True`. Pass each branch name as one list item.
- Write each comment in ASD-STE100.

## Commit

`refactor: add the Python merge library merge.py`
