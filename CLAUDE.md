# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Purpose

This repository is the source of shared files for other repositories. Its only
application code is the sync engine in `cli/`: the Go module of the
engine, the CLI `bootstrap`. Its tests are Go tests in `cli/` and in `e2e/`,
and the task `test` runs them (see "Tests").
Other repositories copy files from it with their own `setup:configs` task. Read
`docs/design/index.md` for the contract of the sync.

## Layout rule

The shared files of a repository are in `repo/tools/<tool>/`, one folder for
each tool: a `tools.toml`, the config files and merge specs in `files/`, the
task files in `tasks/` and the scripts in `scripts/`. Each file of `files/` and
`tasks/` has an entry in `tools.toml`. The engine runs a script from the clone
and never copies it. The packs are in
`repo/packs.toml`, outside each tool folder, so the
sync never copies it. The examples are in `repo/examples/`. The Mac tools are in
`macos/tools/<tool>/`, one folder for
each tool, and their packs are in `macos/packs.toml`. These
files are for this repository only:

- At the root: `README.md`, `CLAUDE.md`, `LICENSE`, `dprint.jsonc` (with the
  excludes of this repository), `eslint.config.mjs` and `go.mod` (the Go
  module of the end-to-end tests).
- `e2e/`: the end-to-end tests, Go tests of the root module (see "Tests").
- `cli/`: the engine of the sync. It is not a tool, so the sync never
  copies it to a consumer. It is the Go module
  `github.com/virajp/bootstrap/cli` (`go.mod`, `cmd/bootstrap/` and
  `internal/`), the CLI `bootstrap` with the commands `migrate`, `tui`, `check`,
  `sync`, `questions`, `detect`, `add`, `remove`, `replace`, `list`, `config`
  (`get`, `set`, `unset` and `show` of a value of `.config/bootstrap/`),
  `settings`, `contract`, `catalog`, `describe`, `lsp` and the group `mac`. A
  plugin
  reads and drives it through the JSON envelope of `--json`: see
  `docs/design/cli.md` ("The JSON contract").
  No Go
  file of the engine names a tool: `TestEngineNamesNoTool` examines this rule.
- `docs/design/`: the contract of the sync, one place for each rule (see "Where
  the rules are"). `docs/plans/`: the vwf plan index.
- `.config/mise/tasks/test/`: the task `test:go` (`go`) and the runner of
  this repository that `test` runs (`_default`, which `bootstrap migrate`
  moves to the part `50-local`). They are not in `repo/tools/` or
  `macos/tools/`, so no consumer gets them.
- `.github/workflows/release-engine.yml` and `cli/VERSION`: the release
  of the Go engine and the minimum version that a clone needs. A signed tag
  `bootstrap@X.Y.Z` on `main` starts the
  release, and the workflow writes `VERSION`. Do not edit `VERSION` by hand. See
  `docs/design/cli.md` ("The release of the engine").
- `.config/mise/tasks/vorssaint/diff`: the task `vorssaint:diff`. It compares
  a Vorssaint backup file with the defaults of `macos/tools/vorssaint/`, and
  exits 1 when a key differs or is missing.
- `macos/fixtures/`: the fixtures of `e2e/mac`.
- `macos/examples/macos.toml`: the sample values file of the Mac side, and the
  model of `e2e/macostools`. A change to a Mac tool also changes this file
  when the tool needs a value or a selection, because `e2e/macostools` fails
  for an unselected tool.

`repo/examples/<example>/` holds sample values files (`repo.toml` and
`tools.toml`, the files of `.config/bootstrap/`), one folder for each
repository shape: `polyrepo`, `polyrepo-main`,
`polyrepo-develop`, `monorepo`, `repo-submodules` and `repo-siblings`.
`repo-submodules` also has the two files in each submodule folder, and
`repo-siblings` has them in each sibling folder only. The sync never
copies or generates them. `docs/sync/index.md` and `docs/sync/vwf.md`
use them as the model of a values file, so a change to the contract of the
values file changes the examples in the same commit.

Each example has an empty `bootstrap_commit`, so a consumer that copies an
example gets no wrong value: the sync replaces the empty line with the
commit of its sync. `e2e/repo/migrate` runs `bootstrap sync --dry-run` with each
example, in a consumer with its members. That case finds an unknown tool, a
wrong member, a wrong shape and an unknown key, because the schema of the
values refuses one.

`docs/sync/` tells an AI agent how to adopt `bootstrap` in a consumer
repository, how to bring it up to date, and how to add and remove tools after
the adoption. After the adoption, `setup:configs` (`bootstrap sync`) keeps the
consumer in sync.

- `docs/sync/index.md`: the adoption prompt. It runs
  `bootstrap questions --json` (the questions are in `.data` of the
  envelope), then `bootstrap migrate --answer …` and
  `bootstrap add tool <tool…> --answer …`, and it brings an adopted
  repository up to date with `bootstrap migrate`.
- `docs/sync/migrate.md`: the items for a repository with older copies.
- `docs/sync/kotlin-dsl.md`: the static prompt that migrates a Gradle build
  from the Groovy DSL to the Kotlin DSL. `vwf.md` and the "Migration"
  sections of `kotlin.md`, `android.md` and `flutter.md` send an agent there.
- `docs/sync/vwf.md`: the entry point for the vwf and stackgen plugins. It
  accepts the `init`, `add`, `remove`, `add plugins` and `remove plugins`
  instructions. A plugin names this path
  and no other bootstrap path.
- `docs/sync/tools/<tool>.md`: one file for each folder in `repo/tools/`, with
  the same sections in each file. The file of the `claude` tool is
  `claude-tool.md`, because a file system that ignores case reads `claude.md`
  as `CLAUDE.md`.
- `docs/sync/packs/<pack>.md`: one file for each pack of `repo/packs.toml`,
  with the tools that it adds and the questions of the adopting AI.

`docs/setup/` tells an AI agent how to adopt `bootstrap` on a Mac, and how to
bring the Mac up to date with `bootstrap mac migrate`. After the adoption, the
task
`updateall:configs` (`bootstrap mac sync`) keeps the Mac in sync. The sync never
copies this folder.

- `docs/setup/index.md`: the Mac adoption prompt. It does the first adoption,
  and it brings a Mac up to date with `bootstrap_commit`.
- `docs/setup/migrate-dotfiles.md`: the method that moves a Mac from the symlink
  model of `macos-setup/dotfiles/` to `macos.toml` and the Mac sync.
- `docs/setup/tools.md`: the reference of the Mac tools, with one section for
  each folder in `macos/tools/`.

This repository is also a consumer of its own `base` pack and of the `claude`
tool. Its values file is still the old `.config/bootstrap.toml` until the user
migrates it with `bootstrap migrate` after the `bootstrap@4.0.0` tag; the engine
of this branch reads `.config/bootstrap/`. `setup:configs` writes
the synced copies at the root (`.config/`, `.vscode/`, `.gitignore`,
`.gitattributes`). Do not edit a synced copy. Change the
file in `repo/tools/<tool>/`, push it, then run `mise run setup:configs` here
too.

## Where the rules are

Each rule of the sync has one place in `docs/design/`. Read the doc before you
change the path:

- Read `docs/design/cli.md` before you change `cli/`.
- Read `docs/design/repo.md` before you change `cli/internal/repo`
  or the entry tasks of `repo/tools/mise/`.
- Read `docs/design/mac.md` before you change
  `cli/internal/macos` or the entry tasks of `macos/tools/mise/`.
- Read `docs/design/packs.md` before you change `repo/packs.toml` or
  `macos/packs.toml`.
- Read `docs/design/tools.md` and `docs/design/mise.md` before you change a
  `tools.toml`. `docs/design/mise.md` also holds
  `cli/internal/misegen`.
- Read `docs/design/bootstrap-config.md` before you change the keys of
  `.config/bootstrap/` or `cli/internal/schema`.
- Read `docs/design/fnox.md` before you change fnox.
- Read `docs/sync/index.md` before you change `docs/sync/`.
- Read `docs/setup/index.md` before you change the Mac adoption.
- Read `docs/setup/migrate-dotfiles.md` before you change the migration from
  `dotfiles/`.
- Read `docs/setup/tools.md` before you change a tool in `macos/tools/`.

Each shipped file that is neither a seed nor a merge source starts with the
`Synced from bootstrap` header: see `docs/design/tools.md` ("The header of a
source file").

## Task scripts

Task scripts are in `repo/tools/<tool>/tasks/`, each with a `[tasks."<path>"]`
entry in its `tools.toml`. The key is the path below `.config/mise/tasks/` in a
consumer. mise uses that path as the task name, for example `git/merge/develop`
is `git:merge:develop`. A `_default` file is the task for the directory name.
The `mise` tool ships each `_default`. A leaf task of a group that an
orchestrator runs with `own_tasks` (`code:format`, `code:lint`, `code:sec`,
`setup:deps`, `setup:secrets`, `test`, `test:e2e`, `test:golden`) has a
number prefix for its run order, for example `code:lint:70-ruff`. `own_tasks`
skips a file without the prefix. A task file is a shell file without an
extension, or a Python file with the `.py` extension. mise removes `.py` from
the task name when no file without the extension has that name, so the part
`70-ruff.py` is the task `code:lint:70-ruff`, and `_default.py` is the task
for the directory name. The `mise` tool ships `test`, `test:e2e` and
`test:golden` as dispatchers: each passes its arguments unchanged to each
part, and with no part it prints one line and exits 0. See
`docs/design/tools.md` ("The `[tasks]` table").

- `_scripts/` contains libraries, not tasks. Each shell library has a Python
  module beside it, with the same names and the same output: `helpers.py`,
  `checks.py`, `placeholder.py`, `worktrees_page.py`, `merge.py` and
  `xcode.py`. A shell task sources a shell library through
  `${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/<name>`. A Python task
  imports a module (see "Python tasks" below).
  - `helpers`: the print functions (`print_header`, `print_success`,
    `print_wait`, and others), `bootstrap_get`, `members`, and
    `gate_files`. Use these functions for all task output.
    `bootstrap_get <key>` is the one reader of a value of
    `.config/bootstrap/`: it prints the JSON answer of `bootstrap config get`
    (read `.value`), with `BOOTSTRAP_BIN`, else `mise x -- bootstrap`. A read
    failure stops the task with code 1. No engine that answers (none, or
    one older than 4.0.0), and an engine that refuses the read with code 2,
    for example a `tools.*` key with no catalog (`get` never fetches), stop
    it with code 2. `members` lists the submodule and
    sibling members of `[members.*]` through it. Call both as
    `VAR=$(bootstrap_get <key>) || exit $?`. `gate_files` runs
    `gate_files.py` with
    `mise x uv@latest python@3.14 -- uv run --quiet --script`.
  - `helpers.py`: the Python form of `helpers`, with the same names. Each
    print function writes the same bytes as the shell function. The
    differences: `bold=True` replaces the leading `--bold`, and
    `file=sys.stderr` replaces `>&2`. Each print function flushes its stream.
    `bootstrap_get(key, repo=None)` returns the JSON answer as a dict (read
    `["value"]`). A read failure stops the task with `sys.exit(1)`, and no
    engine that answers, a refusal with code 2, another code or an answer
    that is not JSON stop it with `sys.exit(2)`, after the same lines as the
    shell. `members(repo=None)` returns a list of
    `(name, folder, label, url)` tuples and stops the task in the same way.
    `gate_files(gate, *args)` runs `gate_files.py` with `sys.executable` and
    returns the path of the list file. `usage_files()` returns the variadic
    `[files]` argument as a list (`shlex.split`). `tool(*args, **kwargs)` and
    `has_tool(name)` are the `mise x --` rule (see below).
  - `gate_files.py`: the file list of one `code:format` or `code:lint` gate.
    See the gate rules below.
  - `checks` and `checks.py`: git predicates. They return an exit status
    (in Python, a bool) only and print nothing. The one exception is
    `get_main_worktree_path`, which prints the path of the main worktree (in
    Python, returns it).
  - `merge` and `merge.py`: `merge_to_destination_branch`, which the
    `git:merge:<branch>` tasks use.
  - `placeholder` and `placeholder.py`: for a task that no stack fills yet. A
    placeholder task contains the `#PLACEHOLDER` marker (`# PLACEHOLDER` in a
    `.py` file, because dprint adds the space) and always exits 0. A stack
    overlay replaces the file. Do not edit a placeholder in place. No task in
    `repo/tools/` uses it now.
  - `xcode` and `xcode.py`: `require_xcode`, `find_xcodeproj` and the
    simulator functions for the Apple tasks. In Python, `require_xcode(task)`
    returns a dict of the names that the shell exports, and also sets them in
    `os.environ`. `find_xcodeproj(task)` returns the path where the shell
    prints it.
  - `worktrees-page` and `worktrees_page.py`: `write_worktrees_page`, which
    writes the HTML page of `git:worktree:list`. It holds a small font with
    the glyphs of that task. The shell function takes no argument and reads
    the arrays of its caller. The Python function is
    `write_worktrees_page(rows, attention, now, folder="/tmp/code-worktrees")`:
    `rows` is a list of dicts, one for each row; `attention` is a list of
    dicts in rank order; `now` is the time text of the page. It returns the
    path of the page, or `None` on a failure.

  The `mise` tool ships `helpers`, `helpers.py`, `gate_files.py`, `checks`,
  `checks.py`, `placeholder`, `placeholder.py`, `worktrees-page` and
  `worktrees_page.py`, the `git` tool ships `merge` and `merge.py`, and the
  `swift` tool ships `xcode` and `xcode.py`. Plan 2 of B30
  (`docs/plans/2026-11-18-python-tasks-repo`) deletes the shell libraries.
- Prefix each tool call with `mise x --`, for example `mise x -- dprint check`.
  To find whether a tool is available, use `mise x -- which <tool>`, not
  `command -v`. System commands (`git`, `awk`, `sed`) and `mise` itself do not
  get the prefix. The tools that mise does not install also do not get it, for
  example `xcodebuild`, `swift`, `pod`, `./gradlew` and the VS Code `code`
  command. In a Python task, `helpers.tool(*args, **kwargs)` runs
  `mise x -- <args>` with `subprocess.run`, and `helpers.has_tool(name)`
  runs `mise x -- which <name>` and returns a bool. A system command runs
  with `subprocess` directly.
- Exception: a tool that a task must have in every repository and environment
  gets a `#MISE tools.<tool>="latest"` directive in that task, and the task
  and the libraries that it sources call it without `mise x --` and find it
  with `command -v`. `mise x` finds only the tools of a config file, and a
  repository before its first sync, or a shell without `MISE_ENV=dev`, has no
  config that names them. The tools are `github:virajp/bootstrap` in
  `setup:configs`, `yq` in `setup:configs` and `setup`, `pre-commit` in
  `setup:precommit` and `code:precommit`, `pre-commit` and `yq` in the
  `git:merge:<branch>` tasks, `yq` in `git:worktree:list`,
  `git:worktree:create` and `git:worktree:delete`, and `dprint` in
  `setup:dprint`. A Python task finds the tool with `shutil.which`, and
  names it in its one `# MISE tools={...}` line, for example
  `# MISE tools={uv="latest", python="3.14", pre-commit="latest"}`. TOML
  does not accept an inline table and a dotted key for the same table, so a
  Python task has no second `tools` line. The engine is the exception to this
  exception: `helpers`
  (`bootstrap_get`), `code:lsp:check` and the scripts of the tools call it as
  `BOOTSTRAP_BIN`,
  else `mise x -- bootstrap`, and find it with `mise x -- which bootstrap`.
  The tool calls of the engine in `cli/` are in
  `docs/design/cli.md` ("The tool calls").
- To find whether the repository has an optional task, use `has_own_task` and
  `own_tasks` from `helpers` (or `helpers.py`). Do not use `mise tasks info`
  or a wildcard `mise run`. As mise does, both find a `.py` part:
  `has_own_task` reads `<file>` when it exists, else `<file>.py`, and
  `own_tasks` lists the part `NN-<name>.py` as `NN-<name>` when no
  `NN-<name>` exists. In a submodule, mise also finds the tasks of the root
  repository, and it runs them in the root directory.
- The rules of a shell task:
  - Write shell for bash 3.2 and bash 5.x. macOS has bash 3.2 as
    `/bin/bash`. Do not use namerefs, `${var,,}`, or associative arrays.
  - Under `set -e`, write a command substitution that can fail as
    `VAR=$(cmd) || VAR=''`. The existing scripts explain this in comments.
    The one exception is `bootstrap_get` and `members`: pass their code on
    with `VAR=$(bootstrap_get <key>) || exit $?`.
- Tasks use `#MISE` and `#USAGE` header comments for metadata and flags (in a
  Python file, `# MISE` and `# USAGE` with a space, which dprint writes; mise
  reads both forms). A `#USAGE flag "--fix"` is available in the script as
  the environment variable `usage_fix`. A Python task reads it with
  `os.environ.get("usage_fix")`. Each search for a header line accepts both
  forms.
- Python tasks: the rules of a Python task or library file.
  - The file has the `.py` extension and mode 755. It has the form of
    `gate_files.py`: line 1 is the shebang
    `#!/usr/bin/env -S uv run --script`, lines 2-3 are the
    `Synced from bootstrap` header, then the
    `# /// script` block with `requires-python = ">=3.14"` and
    `dependencies = []`, then the `# MISE` lines with
    `# MISE tools={uv="latest", python="3.14"}`, then the `# USAGE` lines. A
    library also has `# MISE hide=true` and a description, so `mise tasks`
    does not list it.
  - Use the standard library only (`dependencies = []`). A third-party
    package needs a plan that names it and a `stackgen-reputation` check
    that passes.
  - A task imports the libraries with this preamble, before the first import
    of a library:

    ```python
    sys.dont_write_bytecode = True
    sys.path.insert(0, f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")
    import helpers  # noqa: E402
    ```

    `sys.dont_write_bytecode` stops a `__pycache__` folder below
    `.config/mise/tasks/`. A hand check imports with `python -B`.
  - `bootstrap_get` and `members` stop the task with `sys.exit` (code 1 or
    2), so a task does not catch their exit.
  - dprint (its ruff plugin) is the one formatter of a `.py` file. The ruff
    and pyrefly gates run only in a repository with a `pyproject.toml`, and
    skip `.config/mise/tasks/`.
- Only the `init` task is in mise config (`[mise.tasks.init]` of
  `repo/tools/mise/tools.toml`, always dev). It makes the task scripts
  executable. The task files load in every environment.
- A gate of a language is a `code:lint:<gate>` or `code:format:<gate>` task
  in the tool of that language, for example `code:lint:70-ruff` in `python`.
  `code:lint:60-bootstrap` of `mise` follows the same rules for the files of
  the values folder `.config/bootstrap/`:
  - Without its tool (`mise x -- which <tool>`) or without a file of its
    language, the gate prints nothing and exits 0. Some gates examine a
    project file instead, for example `pubspec.yaml` for dart and `gradlew`
    for android. The task files load in every environment, also where the
    tool is not installed.
  - It gets its files only from `gate_files` of `helpers`, called as
    `list=$(gate_files "$0" <its flags and files>) || exit $?`, then
    `trap 'rm -f "${list}"' EXIT` (in Python,
    `helpers.gate_files(<gate file>, <its flags and files>)`, which returns
    the path of the list; the gate deletes the list). The list is a temporary
    file with one path for each line, and an empty file means that the gate
    does not apply. The base set is the given files, else with `--all` every
    file of the
    repository, else the uncommitted files (staged, changed and new). Then
    three filters remove files, in this order: the repository filter (a
    submodule, a nested repository, a symlink, a deleted file, a path that
    git does not list or ignores, and a name with a line break or a leading
    `-`), the global lists `code_files` of the `mise` tool
    (`.config/bootstrap/config-files.toml`), and the `#FILES` header of the
    gate. In a gate run, `gate_files` is the only place that removes a file:
    a gate has no own filter and no `--force-exclude`, and each tool gets
    only that list (dprint with `--stdin-files`, detekt through its argument
    file, each other tool through `tr '\n' '\0' < "${list}" | xargs -0`).
    `--debug` prints each removed file on stderr.
  - The `#FILES` header of a gate, right after its `#USAGE` lines, names the
    files of its language: `#FILES include="..." exclude="..."` and, for
    ShellCheck, `shebang="..."`. A `.py` gate writes `# FILES` with a space.
    See `docs/design/tools.md` ("The `#FILES` header of a gate").
  - Each gate declares `--all`, `--fix`, `--debug` and the variadic
    `[files]`. A gate of a module (android, go, gomod, uv) maps the files of
    its list to their modules.
  - Without `--fix`, a format task is the check of its tool. Do not add a
    lint task that runs the same check, for example `code:lint:ktlint` next
    to `code:format:30-ktlint`.
  - Its config file, when it has one, is a seed.

`bootstrap` does not ship these items on purpose. Do not add them again
without a new reason:

- `code:count`: it only counts the lines of code.
- `.config/claude-status.json`: claude-status finds the name of the
  repository from `origin`. The file is plain JSON, so it cannot have the
  bootstrap header.
- `.config/miserc.toml` with `env_conf_d` in a repository: the sync writes only
  folder fragments `conf.d/<tool>/mise.<env>.toml`, and that setting does not
  apply to them. The `mise` tool of `macos/tools/` ships the user-level
  `.config/mise/miserc.toml` with that setting.
- A `code:check` task group: a check of a language is a `code:lint:<name>`
  task, for example `code:lint:90-uv` for `uv lock --check`.

## Task output: headers and subheaders

An orchestrator is a task that runs other tasks with `mise run`. A leaf is a
task that runs only commands.

1. A task first checks whether it applies. If it does not apply, it prints
   nothing and exits 0. For example, `setup:deps:50-python` without a
   `pyproject.toml`.
2. A leaf prints `print_header` first, and only once. Nothing comes before
   it.
3. Each later step of a leaf that needs separation is a `print_subheader`.
4. Status lines (`print_wait`, `print_success`, `print_info`, `print_warn`,
   `print_error`) go under a header or a subheader. They never replace one.
5. An orchestrator prints nothing for a `mise run`. The subtask prints its own
   header. Exception: before `mise run --cd <member>`, it prints a header that
   names the member, because a member is a separate repository.
6. A group of commands in an orchestrator is its own subtask, for example
   `setup:linter`. A single direct command in an orchestrator gets a
   `print_header`, because it is at the level of the sibling subtasks.
7. The top-level orchestrator ends with `line_sep1` and
   `print_success "<Done message>"`, not with a header.
8. Header and subheader text is an "-ing" verb phrase with " ..." at the end,
   for example "Installing pre-commit hooks ...".
9. A task that runs only inside a git hook (`code:git-config`) prints status
   lines only, no header.
10. Exception to rules 2 and 5: `code:sec` prints a `print_header` before it
    runs its `code:sec:<scanner>` tasks. Each scanner task prints only a
    `print_subheader` under that header. With `--staged`, a scanner task
    without a staged mode prints nothing.
11. Exception to rule 2: when its output is not a terminal, `git:worktree:list`
    prints only JSON Lines, because an agent reads each line.
12. Exception to rule 2: with `--debug`, the drop lines of `gate_files` go to
    stderr and can come before the header of a gate, also of a gate with an
    empty list.

## Tests

The tests are Go tests, in two Go modules: `cli/` (the unit tests of the
engine) and the root module `github.com/virajp/bootstrap` (`go.mod` at the
root), with the end-to-end tests in `e2e/`. The end-to-end tests run the
engine, built from the working tree's `cli/`, with the working tree's
`repo/` and `macos/`, against scratch consumers and scratch homes. The root
module holds them because `go test` keeps a result only for the files inside
the module of the test (see "The test cache" below).

- `e2e/internal/harness`: the common parts. `Main` (the `TestMain` of each
  package) opens each file of the inputs of the package, builds the engine
  once (or uses `BOOTSTRAP_BIN`, an absolute path), and writes the mise shim.
  The environment of a case has `BOOTSTRAP_VERBOSE=1`, so a case compares the
  per-file log lines of the sync (the cases of the default output unset it;
  see `docs/design/cli.md`, "The output of a run"), `BOOTSTRAP_CONFIG` at a
  missing path, so the config file of the user never changes a case (see
  `docs/design/cli.md`, "The config file"), and no `GIT_*` variable of a git
  hook, so a scratch repository never writes to this repository.
- The mise shim: `mise x -- <command>` runs the command directly, with the
  PATH of the test run, and each other `mise` call goes to the real mise. A
  `mise x` costs 0.1 s alone and more than 1 s under load, and the engine
  makes one for each folder of its scripts. A case with `RealMise` uses the
  real mise. The shim is at a path that ends in `.local/bin/mise`, because
  `bootstrap mac sync` accepts only such a path.
- `e2e/internal/repocase`: the runner of a repo case (`rc.Run`, `rc.Check`):
  a clone of the bootstrap copy (`repo/` and `cli/VERSION`; all of `cli/` with
  `FullCLI`), a copy of a consumer, the values folder of the case, then
  `bootstrap sync --dry-run`, `bootstrap check` or the `Call` of the case.
- `e2e/repo/sync`, `config`, `files`, `migrate`, `selection`, `detect` and
  `commands`: the repo cases. They run `bootstrap sync --dry-run`,
  `bootstrap check` and, for a case that compares a written file,
  `bootstrap sync`; the migrate cases run `bootstrap migrate` on a new values
  folder, on an old values file and on its migration to `.config/bootstrap/`;
  the selection cases run `bootstrap add tool <tool…>`,
  `bootstrap add pack <pack…>`, `bootstrap remove tool <tool…>`,
  `bootstrap replace tool <old> <new>` and `bootstrap detect --add`; the
  config verb cases run `bootstrap config get|set|unset|show`; the settings
  cases point `BOOTSTRAP_CONFIG` or `--config` at a file of the case; the JSON
  cases run `--json` on `contract`, `catalog`, `describe`, `lsp`,
  `questions`, `config`, `check` and `add`, and compare the envelope on
  stdout; the project cases run `bootstrap detect --json`, `detect` and
  `detect --add --yes` in a scratch consumer with project folders. The
  examples cases run `bootstrap sync --dry-run` with each example of
  `repo/examples/`. The tool-test-part cases read an old stackgen copy of the
  compose `test/golden` task from the claude-plugins clone of
  `CLAUDE_PLUGINS_CLONE`; without it, each of the four cases is skipped with
  one line.
- `e2e/repo/commands` and `e2e/repo/tasks`: the rendered tasks. The entry
  cases run `setup:configs` with a fake engine and with the real engine, and
  the clone, lock and version cases; the landing cases run `git:merge:main`
  with the real merge library, and the other merge cases run it with a stub
  library; the worktree cases run `git:worktree:create` with `--no-setup` and
  `git:worktree:delete`; the LSP cases run `code:lsp:check`, the setup:ai
  cases `setup:ai`, and the setup case `setup`; the gate_files cases run the
  working tree's `gate_files.py` in a scratch repository with `--debug`; the
  `run_gate` cases run `code:lint:60-bootstrap`; the orchestrator case runs
  `code:lint` with `--all` and files, and expects exit 2; the Python part
  cases run `own_tasks`, `has_own_task` and the `--all` search with `.py`
  parts.
- `e2e/repo/pylib`: the Python libraries against the shell libraries. Each
  case copies the libraries of the working tree's `repo/` into a scratch
  `_scripts/`, runs a Python probe that imports a module and a shell probe
  that sources the shell library, and compares the output bytes, the exit
  code and the return values.
- `e2e/mac`: `bootstrap mac sync` against the fixtures in `macos/fixtures/`,
  with the exit code, the output and the files of each case. The entry cases
  run the task `updateall:configs` of `macos/tools/mise/` against a local bare
  origin. Each case has its own temporary `HOME`.
- `e2e/macostools`: `bootstrap mac sync` with each tool of `macos/tools/` and a
  copy of `macos/examples/macos.toml`, in a temporary `HOME`. No `[files]`
  table gives `mode = "seed"` to a JSON, YAML or TOML target; the example
  selects each tool and each rendered file parses; an example without the
  answer `config_path` of `mise` stops the sync with code 2; a mise tool on
  both sides of `repo/tools/` and `macos/tools/` has one version and one
  `depends`, and a mise setting on both sides has one value (an intended
  difference goes in `driftAllowed`, with a reason); the git tool needs one
  email only.

Each check of a case logs one line `CHECK <name>`, and each step of a case one
line `TIMING <step> <seconds>s`. The cases run in parallel (`t.Parallel`), and
`go test` runs the packages in parallel. A case writes only below its own
folder; `HARNESS_KEEP=1` keeps the folder of a failed case, and
`HARNESS_KEEP=all` keeps each folder. The tests need no network when
gomplate, jq, yq, uv and Python 3.14 are installed.

The test cache: `go test` runs a package again only when the package, a file
that it opened, or an environment variable that it read changed. `Main` opens
each file of `cli/` and of the inputs of the package (`repo/`, `macos/`), so a
change to one of them runs the packages that read it, and only them. The
cache key also has the variables that a test read, PATH included: mise gives
the tool folders of a task in a different order on each run, so `test`
sorts them first. A change to an outside tool (gomplate, yq, mise) changes
PATH only when its version folder changes; use `--all` after such a
change. The slots: a case runs only while it holds one of the slots that
all test packages share (one for each CPU, `HARNESS_SLOTS` to change), so
the packages share the CPUs and do not overload them.

The tasks:

- `test` (the dispatcher of the `mise` tool) runs the part `test/50-local` of
  this repository. It runs `test:go` first, then gotestsum in the two modules
  at the same time. `mise run test --all` runs each test again
  (`go test -count=1`). Each run writes the JSON log of `go test -json` to
  `${XDG_STATE_HOME:-$HOME/.local/state}/bootstrap/test-runs/<UTC time>.json`
  and prints the slowest tests (`gotestsum tool slowest`).
- `test:go` runs `go vet` and a `gofmt` check in the two modules. Some unit
  tests compare a case with a golden file in `testdata/`; see
  `docs/design/cli.md` ("The files of `cli/`"). `TestEngineNamesNoTool`
  fails when a string literal of the engine code names a tool, outside its
  allow list. The task runs the Go of the `go` tool, which this repository
  selects.
- To run one package or one case: `mise x -- go test ./e2e/repo/detect/` or
  `mise x -- go test ./e2e/repo/detect/ -run 'TestDetect/detect_--yes'`.

No hook and no `git:merge:<branch>` task runs a test, because a consumer does
not have the tests. Run `mise run test` by hand before you commit.

## Change a shared file

1. Change the file in `repo/tools/<tool>/`.
2. If the change adds, removes or changes a file, a mise tool, a task, a hook,
   a seed or a value of the tool, change `docs/sync/tools/<tool>.md` too
   (`claude-tool.md` for `claude`). A
   new tool needs a new file in `docs/sync/tools/`, and a line in the list of
   `docs/sync/index.md`. A removed file needs a row in "Files that bootstrap no
   longer ships" of `docs/sync/index.md`, and a `migrate.md` item. A change to
   `repo/packs.toml` changes
   `docs/sync/packs/<pack>.md` too, and a new pack needs a new file there. A
   change to a tool in `macos/tools/` changes its section in
   `docs/setup/tools.md` too.
3. Commit on a local work branch. Land it on `main` with
   `mise run git:merge:main <branch>`, which merges and pushes.
4. In this repository and in each consumer repository, run
   `mise run setup:configs`, then commit the result. This repository allows no
   direct commit on `main` (no `direct_commit`, which `bootstrap migrate` makes
   `landing = "direct-merge"`), so the `no-commit-to-branch` hook refuses a
   commit on `main`. Thus in this
   repository, sync on a local `sync-<name>` branch, commit
   `ops: sync bootstrap with <change>`, and land it with
   `mise run git:merge:main sync-<name>`. The sync also updates
   `bootstrap_commit`, which stays one merge behind `main`.
