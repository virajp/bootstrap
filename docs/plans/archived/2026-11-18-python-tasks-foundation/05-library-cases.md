# U5 — The Go cases of the Python libraries

- **Read first:** `e2e/internal/harness/harness.go`,
  `e2e/repo/tasks/main_test.go` and `e2e/repo/tasks/gate_test.go` (the form
  of a case that runs a library), then the modules that U1, U2 and U3 wrote.
- **Lazy-load:** `e2e/repo/commands/tasks_test.go` (the stub engines and the
  stub merge library), `docs/design/cli.md` ("The JSON contract").

## Ruling

- **D9:** the output bytes of each `helpers.py` print function equal the
  bytes of the shell function, with color on and off.
- **D10:** `bootstrap_get` returns the dict; engine code 1 gives exit 1;
  engine code 2 or no engine gives the error lines and exit 2. `members`
  returns the tuples of the submodule and sibling members only.
- **D11, D12, D18:** `tool`, `has_tool`, `gate_files`, `own_tasks` and
  `has_own_task`.
- **D19:** `require_xcode` returns a dict and sets `os.environ`;
  `find_xcodeproj` returns the path.
- **D30:** "[U5] fixes GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL, GIT_COMMITTER_NAME,
  GIT_COMMITTER_EMAIL, GIT_AUTHOR_DATE and GIT_COMMITTER_DATE, so the two git
  logs (git log --format=\"%H %P %s\" --all) and the output are the same,
  hashes included."
- **D15:** the cases run with `mise run test`.

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

## Edits

1. **`e2e/repo/pylib/`** — a new Go package with a package comment. Each
   case renders or copies the modules of `repo/tools/*/tasks/*.py` into a
   scratch `.config/mise/tasks/_scripts/`, then runs a small Python probe
   with `uv run --quiet --no-project --python 3.14` (or `python3.14` when
   the harness gives it) that imports the module. Cases:
   - each print function of `helpers.py` against the shell `helpers`, byte
     for byte, with color on (`TERM`, a TTY flag if the shell function reads
     one) and with `NO_COLOR=1`;
   - `bootstrap_get` with a stub engine in `BOOTSTRAP_BIN` that answers JSON
     (exit 0), exits 1, and exits 2; and with no engine;
   - `members` with a stub answer that holds a submodule, a sibling and a
     folder member;
   - `own_tasks` and `has_own_task` with `NN-x`, `NN-x.py`, `x`, `x.py`,
     `_default.py`, a folder with both `NN-x` and `NN-x.py` (the names of
     D22, the same as `mise tasks ls`), and a file without the execute bit;
   - a print function with `file=sys.stderr` writes to stderr only (D22);
   - `gate_files` against the real `gate_files.py` in a scratch git repo;
   - `usage_files` (D26) with a name that has a space, given quoted;
   - `checks.py` predicates against the shell predicates in a scratch repo
     with a worktree;
   - `worktrees_page.write_worktrees_page` against the shell function (D23):
     the same HTML bytes for the same input, with the date line fixed, both
     sides in a scratch folder of the case;
   - `placeholder.py` (D27) with a `.py` task that has `# PLACEHOLDER`;
   - `merge.merge_to_destination_branch` in a scratch repo with a local bare
     origin, in the landing mode `direct-merge`, with the fixed git identity
     and dates of D30: the same git log, hashes included, and the same
     output as the shell function;
   - `xcode.find_xcodeproj` in a folder with no Xcode project (the error
     lines and exit 1) and with one (the returned path); `xcode.require_xcode`
     when `xcodebuild` is on the PATH (a dict, and the same names in
     `os.environ`), else `t.Skip` with one line.

## Verification

- `mise x -- go vet ./e2e/repo/pylib/` passes.
- `mise x -- go test ./e2e/repo/pylib/` passes, and `-count=1` passes again.

## Guardrails

- Do not edit a module: a failure that a module causes is a QUESTION to the
  orchestrator, which routes it to U1, U2 or U3.
- Do not edit a file below `e2e/internal/`. Write a private helper in the
  package when the harness lacks something.
- Each case uses `t.Parallel()`, `harness.Dir(t)` and `harness.Env()`;
  never `os.Setenv` or `os.Chdir`.
- Write each comment in ASD-STE100.

## Commit

`test: compare the Python libraries with the shell libraries`
