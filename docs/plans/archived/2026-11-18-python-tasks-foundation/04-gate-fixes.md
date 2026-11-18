# U4 — The gate fixes, dprint as the one Python formatter and the uv cache

- **Read first:** `repo/tools/mise/tasks/gate_files.py`,
  `repo/tools/mise/tasks/helpers` (`has_own_task`, `own_tasks`),
  `repo/tools/mise/tasks/code-format.tmpl`,
  `repo/tools/mise/tasks/code-lint.tmpl`, `repo/tools/python/tools.toml`,
  `repo/tools/python/tasks/code-lint-ruff`, `macos/tools/uv/tools.toml`,
  `e2e/repo/tasks/gate_test.go`.
- **Lazy-load:** `e2e/internal/harness/harness.go` and
  `e2e/repo/tasks/main_test.go` (how a case runs a rendered task);
  `docs/design/tools.md` ("The `#FILES` header of a gate").

## Ruling

- **D12:** "gate_files.py keeps its command line; it accepts the header line
  `# FILES` as well as `#FILES`, because ruff format adds the space."
- **D8:** every search for a USAGE line accepts `#USAGE` and `# USAGE`.
- **D18:** "The shell own_tasks and has_own_task, and the --all search of
  code-format.tmpl and code-lint.tmpl, follow the naming of mise in D22: a
  part NN-x.py is the task NN-x when no NN-x exists." The search accepts
  `#USAGE` and `# USAGE` (D8).
- **D13:** "dprint is the one formatter of every .py file [...]. Remove the
  gate code:format:40-ruff ([tasks.\"code/format/40-ruff\"] and
  repo/tools/python/tasks/code-format-ruff). ruff (lint) and pyrefly (types)
  keep requires_found = [\"pyproject.toml\"]: they run only in a repository
  that builds something with Python, not for the synced task files."
- **D6:** remove `UV_NO_CACHE = true` from `[mise.env]` of
  `repo/tools/python/tools.toml` and of `macos/tools/uv/tools.toml`. Keep
  `UV_NO_MANAGED_PYTHON` and `UV_PYTHON_DOWNLOADS`.
- **D15:** each changed behavior gets a Go case.

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

- **D28:** "the # FILES header of both gates [code:lint:70-ruff and
  code:lint:50-pyrefly] gets exclude=\".config/mise/tasks/**\"."

## Edits

1. **`repo/tools/mise/tasks/gate_files.py`** — the header parser accepts a
   line that starts with `# FILES ` as well as `#FILES `; two such lines in
   any mix are still the error "two #FILES lines". Update the comments that
   name the header.
2. **`repo/tools/mise/tasks/helpers`** — `has_own_task` and `own_tasks` per
   D18. Keep bash 3.2: use `${name%.py}`, no new bash features.
3. **`repo/tools/mise/tasks/code-format.tmpl`** and
   **`code-lint.tmpl`** — the `--all` search per D18: the part file is
   `<path>` or `<path>.py`, and the pattern is
   `^# \?USAGE flag "--all"` (grep basic regex).
4. **`repo/tools/python/tools.toml`** — D13: remove
   `[tasks."code/format/40-ruff"]`; keep every `requires_found`. D6: remove
   `UV_NO_CACHE`. Update the comments, the tool `description` and `summary`,
   so they say that dprint formats the Python files, and that ruff and
   pyrefly lint and check a repository with a `pyproject.toml`.
5. **`repo/tools/python/tasks/code-format-ruff`** — D13: `rm` the file.
6. **`repo/tools/python/tasks/code-lint-ruff`** and **`code-lint-pyrefly`** —
   D28: the `#FILES` line gets `exclude=".config/mise/tasks/**"`.
7. **`macos/tools/uv/tools.toml`** — D6.
8. **`e2e/repo/tasks/gate_test.go`** — a case of a gate file with a
   `# FILES` header (the same list as `#FILES`), and a case with one
   `#FILES` and one `# FILES` line (the error of two lines).
9. **`e2e/repo/tasks/parts_test.go`** — new file in the package of
   `gate_test.go`: cases of the rendered helpers and orchestrators with a
   part `70-probe.py`: `own_tasks` prints `code:lint:70-probe`;
   `has_own_task` finds `x.py` and `x/_default.py`; the rendered
   `code:lint --all` passes `--all` to a `.py` part with `# USAGE flag "--all"`.

## Verification

- `mise x -- go test ./e2e/repo/tasks/ ./e2e/repo/config/` passes. If a Go
  case names `code:format:40-ruff`, it is a QUESTION: no unit owns that file.
- `mise x -- go test ./e2e/macostools/` passes (the drift check of the mise
  settings and versions on both sides).
- `grep -n UV_NO_CACHE repo/tools/python/tools.toml macos/tools/uv/tools.toml`
  finds nothing.

## Guardrails

- Do not edit `repo/tools/mise/tools.toml`: U1 owns it (S2).
- Do not edit a file below `e2e/internal/`.
- A case uses `t.Parallel()`, `harness.Dir(t)` and `harness.Env()`; never
  `os.Setenv` or `os.Chdir`.
- Write each comment in ASD-STE100.

## Commit

`feat: find Python parts, accept # FILES, format Python with dprint only and
turn the uv cache on`
