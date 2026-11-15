# U1 — The helper, the global list and the orchestrators

- **Read first:** every owned file that an edit below names, top to bottom:
  `repo/tools/mise/tools.toml`, `repo/tools/mise/tasks/helpers`,
  `repo/tools/mise/tasks/code-format.tmpl`,
  `repo/tools/mise/tasks/code-lint.tmpl`,
  `repo/tools/mise/tasks/code-lint-bootstrap`.
- **Lazy-load:** `docs/design/tools.md` ("The `[files]` table", "The
  `[tasks]` table", "The header of a source file", "The `[defaults]`,
  `[questions]` and `[contributes]` tables"), `docs/design/cli.md` (the merge
  order of the values, about lines 1104-1129),
  `repo/tools/dprint/files/dprint.jsonc.merge.toml` (how a template reads
  `.config.<key>` and `.all.<tool>`), the design artifact (local copy:
  `/private/tmp/claude-502/-Users-virajpatel-Projects-github-com-virajp-bootstrap/727f621c-6790-45dd-b071-2fca2d5e3e90/scratchpad/gate-file-sets.html`;
  it can be gone, and the rulings below are complete without it).

## Ruling

- **D1:** with no files and no `--all`, the set is the uncommitted files: the
  staged files (`git diff --cached --name-only --diff-filter=d`), the changed
  files that are not staged (`git ls-files --modified`) and the new files
  (`git ls-files --others --exclude-standard`).
- **`--all`:** the set is `git ls-files --cached --others --exclude-standard`.
- **D16 (the repository filter):** paths are relative to the repository root
  (`MISE_PROJECT_ROOT`). The filter drops a gitlink (a submodule, mode
  `160000`), a path that ends with `/` (a nested repository that `--others`
  lists), a path that is not a regular file (a deleted file), and a given path
  that `git ls-files --cached --others --exclude-standard` does not list (an
  ignored path, a path in a submodule or outside the repository). It never
  recurses into a submodule.
- **D9:** the mise tool owns the key `code_files` with two lists, `includes`
  and `excludes`; an empty `includes` means every file. Its `[defaults]`:
  `includes = []`, `excludes = ["**/*-lock.yaml", "**/*.svg",
  "**/package-lock.json"]`. Other tools add patterns with `[contributes.mise]`;
  a consumer adds patterns in `[tools.mise]` of `.config/bootstrap/tools.toml`.
  The sync renders the lists to `.config/bootstrap/config-files.toml`, a
  template file of the mise tool. User: "agree but the location of the file
  must be `.config/bootstrap/config-files.toml`".
- **D3:** each gate names its own lists in a header comment of the gate file:
  `#FILES include="..." exclude="..."`, plus `shebang="..."` (D11). A gate with
  no `#FILES` header gets every file of the set.
- **D11:** with `shebang="sh bash dash ksh"`, a file is also included when its
  first line starts with `#!` and names one of these shells (as the
  interpreter, or after `env`). `shell_files_in_scope` of `helpers` goes.
- **D12 (pattern syntax):** a pattern has the gitignore glob form. A pattern
  without `/` matches the file name in each folder. A pattern with `/` matches
  from the repository root. `**` matches any number of folders. Values in a
  header are separated by spaces.
- **D7:** in a gate run the helper is the only place that removes a file.
  User: "I am more keen on standardizing the approach so it becomes easier to
  debug".
- **D13:** the mise tool ships `_scripts/gate_files.py` as a library; `helpers`
  gets the bash function `gate_files`, which each gate calls as
  `gate_files "$0"` with the gate's flags and files. It reads the `#FILES`
  header of the gate file, applies the mode, the repository filter, the global
  filter of `.config/bootstrap/config-files.toml` and the gate filter, in this
  order. It writes the list to a temporary file, one path for each line, and
  prints the path of that file. An empty list gives an empty file. The
  function runs the helper with `mise x -- python3`; without Python the gate
  stops with an error. With `--debug`, the helper prints each removed file to
  stderr: `drop <path>: <layer> <kind> "<pattern>"`, for example
  `drop vendor/x.js: global exclude "vendor/**"`. A path that a list of
  includes does not match prints `<layer> include` with no pattern.
- **D14:** the list file has one path on each line, so dprint can read it with
  `--stdin-files`, detekt as `@file`, and the other tools through `xargs`.
- **D17:** `code:format` and `code:lint` declare `--all`, `--fix`, `--debug`
  and the variadic `[files]`; they pass `--all`, `--fix`, `--debug` and the
  files to each gate. `--all` with files stops the orchestrator with a usage
  error, exit 2. The pre-commit hooks do not change.
- **D5:** the file logic is Python; the gates and the orchestrators stay bash.
- **D2:** the global list is a key in the values files of `.config/bootstrap/`,
  with `[contributes]` from the tools; the sync renders it into a file that the
  helper reads.
- **D4:** user: "I will suggest a help function that can be called by any of
  the gates to get the list". The orchestrator passes only the mode and
  `--fix` and `--debug`; it computes no list.
- **D6:** the gate tasks stay; the orchestrator runs them with `own_tasks`.
- **D8:** the values key is the one source; the sync also renders the same
  excludes into `dprint.jsonc` (U2's merge spec). A seed keeps its own text.
- **D15:** this plan runs after `2026-11-15-test-parts`; keep every change
  that the plans before it made in the owned files.
- **D22:** `gate_files.py` follows the formatter that `code:format` of this
  repository applies to `*.py` (dprint with its ruff plugin); U2 removes
  `*.py` from the shuck associations.
- **D24:** the gate deletes the list file: `trap 'rm -f "<list file>"' EXIT`
  right after the `gate_files` call. The helper does not delete it.
- **D25 (pattern syntax, replaces the D12 text):** the gitignore glob form
  without negation. A pattern without `/` (a trailing `/` aside) matches the
  name in each folder. A pattern with a leading or inner `/` matches from the
  root (a leading `/` is removed). A trailing `/` matches a folder and every
  path below it. `**` matches any number of folders. A pattern that starts
  with `!` is malformed: exit 2.
- **D27:** with a `shebang` key, the helper also drops a file that the
  include names match when its first line is a `#!` line that names a shell
  not in the list (for example zsh). `--debug`: `gate shebang`.
- **D28:** the task `code` declares `--all` and passes it to `code:format` and
  `code:lint`; `code:sec` does not change. Its description says so.
- **D29, D39:** a given path is relative to the repository root (as the
  pre-commit hook passes it). The helper removes `./`, resolves `..`, and
  makes an absolute path inside the repository relative. A path outside the
  repository is dropped (`repo unlisted`).
- **D37:** the orchestrator passes `--all` only to a gate whose file has a
  `#USAGE` line for `--all` (a grep of the gate file). A gate without it gets
  `--fix`, `--debug` and the files as today.
- **D42:** `gate_files.py` accepts `--fix` and ignores it.
- **D44:** the helper reads the top block of the gate file: every line from
  line 1 to the first line that is neither a comment nor blank. A `#FILES`
  line is a line of that block that starts with `#FILES `. A gate puts it
  right after its `#USAGE` lines. No `#FILES` line: every file. Malformed,
  exit 2: two `#FILES` lines, a key other than `include`, `exclude` or
  `shebang`, a key twice, a value without double quotes. An empty value
  (`include=""`) is an empty list.
- **D53:** a header with no `include` key, or with `include=""`, selects every
  file, as an empty global `includes` does.
- **D56 (replaces the 60-bootstrap header of Edit 7):**
  `#FILES include=".config/bootstrap/*.toml"
  exclude=".config/bootstrap/config-files.toml"`.
- **D45:** the drop lines go to stderr, only with `--debug`, and can come
  before the `print_header` of the gate and from a gate with an empty list.
  Without `--debug`, a gate with an empty list prints nothing.
- **D46:** a pattern without `/` (such as `node_modules`) matches, as in
  gitignore, a file or a folder with that name at any depth, and each path
  below such a folder.
- **D47:** the helper reads the interpreter of a `#!` line: the last path part
  of the first word, or after `env` (and its options such as `-S`) the first
  word that is not an option. Known shells: `sh bash dash ksh mksh zsh fish
  csh tcsh`. A file that the include names match is dropped only when its
  interpreter is a known shell that is not in the `shebang` list. Any other
  interpreter or no `#!` line keeps it. A file that only the `shebang` key
  selects needs an interpreter in the list.
- **D48:** the repository filter checks in this order: not listed by git
  (`repo unlisted`), a gitlink, a nested repository, a symlink (`repo
  symlink`: each symlink is dropped, whatever its target), not a regular file
  (`repo deleted`). Add `repo symlink` to the words of D33.
- **D30:** `gate_files.py` with `--all` and files stops with exit 2 and a usage
  message.
- **D34 (replaces `mise x -- python3` of D13):** the `gate_files` function
  runs `mise x python@3.14 -- python3`. The `TOOL@VERSION` argument names the
  version itself, so the helper gets Python 3.14 in each environment. The
  version is the pin of `repo/tools/python/tools.toml`; add a comment beside
  the call that names that pin. User: yes.
- **D33 (the `--debug` words):** `drop <path>: <layer> <kind> ["<pattern>"]`.
  `repo gitlink`, `repo nested-repo`, `repo deleted`, `repo unlisted` (no
  pattern); `global include` (no pattern), `global exclude "<pattern>"`;
  `gate include` (no pattern), `gate exclude "<pattern>"`, `gate shebang` (no
  pattern).

## Edits

1. **`repo/tools/mise/tasks/gate_files.py`** (new) — a Python 3.14 script that
   uses the standard library only (`tomllib`, `pathlib`, `subprocess`,
   `argparse`). It starts with the `Synced from bootstrap` header as Python
   comments (`docs/design/tools.md`, "The header of a source file"). Its
   arguments: the path of the gate file, `--all`, `--debug`, `--fix` (ignored,
   D42), and the files. It
   runs `git` in `MISE_PROJECT_ROOT` with `-c core.quotePath=off`, builds the
   base set (D1 or `--all` or the given files), applies the repository filter
   (D16), the global filter (the `includes` and `excludes` of
   `.config/bootstrap/config-files.toml`; a missing file means two empty
   lists), and the gate filter (the `#FILES` header of the gate file; the
   header is a comment line that starts with `#FILES `, in the comment block
   at the top of the file). It writes the sorted, unique list to a new
   temporary file (`tempfile.mkstemp`) and prints its path. Exit 0 on success,
   exit 2 with a message on stderr for a malformed header or an unreadable
   `config-files.toml`, and the exit code of `git` when `git` fails.
2. **`repo/tools/mise/tasks/helpers`** — add `gate_files`: it runs
   `mise x python@3.14 -- python3
   "${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/gate_files.py"`
   (D34)
   with its arguments and prints the path that the script prints. If the
   script fails, `gate_files` prints the error with `print_error` and returns
   its exit code. Remove `shell_files_in_scope` (D11). Bash 3.2 rules apply.
3. **`repo/tools/mise/tools.toml`** —
   - a `[defaults]` table with `code_files` (D9), with a comment that names
     the rendered file and `[contributes.mise]`;
   - `[files.".config/bootstrap/config-files.toml"]` with `mode = "template"`
     and `source = ".config/bootstrap/config-files.toml.tmpl"`;
   - `[tasks."_scripts/gate_files.py"]` with `library = true` and
     `source = "gate_files.py"`.
4. **`repo/tools/mise/files/.config/bootstrap/config-files.toml.tmpl`** (new) —
   the `Synced from bootstrap` header as TOML comments, then `includes` and
   `excludes` rendered from `.config.code_files` (the merged values of the
   mise tool), one TOML array each.
5. **`repo/tools/mise/tasks/code-format.tmpl`** and
   **`code-lint.tmpl`** — declare `#USAGE flag "--all"` (help: "format (or
   lint) every file of this repository, not only the uncommitted files"),
   keep `--fix`, `--debug` and `[files]` (help: "the files to format (or
   lint); the default is the uncommitted files"). Stop with exit 2 and
   `print_error` when `--all` comes with files. Pass `--fix`, `--debug` and
   the files to each `own_tasks` gate, and `--all` only to a gate that
   declares it (D37). Change the comments that
   say "the whole tree". Keep what the plans before this one left in these
   files (for example the removal of the member flags, D100 of
   2026-11-14-config-folder).
6. **`repo/tools/mise/tasks/code`** — declare `--all` and pass it to
   `code:format` and `code:lint` (D28); change its description.
7. **`repo/tools/mise/tasks/code-lint-bootstrap`** — declare `--all`; add the
   header `#FILES include=".config/bootstrap/repo.toml
   .config/bootstrap/tools.toml"`; get the list with `gate_files`; with an
   empty list, exit 0 with no output. Keep its check of the engine as the
   plans before this one left it.

## Verification

- `mise x -- mise run code:format` passes.
- `python3 -I -c "import ast,sys; ast.parse(open(sys.argv[1]).read())"
  repo/tools/mise/tasks/gate_files.py`
  exits 0.
- `grep -rn shell_files_in_scope repo/tools/mise/` finds nothing.
- In a scratch git repo with a submodule, a nested repository, a staged
  change, an unstaged change, a deleted file and a new file, the script prints
  a list file with exactly the paths of D1 and D16. Run it by hand; do not
  commit the scratch repo.

## Guardrails

- Touch nothing outside `repo/tools/mise/`. The other gates are U2's and
  U3's; `code-lint-linter` still calls `shell_files_in_scope` until U2 changes
  it, so the wave gate of wave 1 runs the synced copy at the root, not this
  file.
- Do not edit the synced copies at the root of this repository
  (`.config/mise/tasks/`, `.config/bootstrap/`).
- Write shell for bash 3.2 and bash 5.x; no namerefs, no `${var,,}`, no
  associative arrays.
- Python: standard library only (the plan lists no dependency).
- Delete with `rm`, never `git rm`.

## Commit

`feat: gate_files selects the files of each code:format and code:lint gate`
