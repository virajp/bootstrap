# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Purpose

This repository is the source of shared files for other repositories. Its only
application code is the sync engine in `cli/`: the Go module of the
engine, the CLI `bootstrap`. Its tests are `test:repo`, `test:mac`,
`test:macos-tools` and `test:go`, and the task `test` runs each test task (see
"Tests").
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
  excludes of this repository) and `eslint.config.mjs`.
- `cli/`: the engine of the sync. It is not a tool, so the sync never
  copies it to a consumer. It is the Go module
  `github.com/virajp/bootstrap/cli` (`go.mod`, `cmd/bootstrap/` and
  `internal/`), the CLI `bootstrap` with the commands `init`, `check`, `sync`,
  `questions`, `detect`, `add`, `remove`, `replace`, `list`, `config` and the
  group `mac`.
  No Go
  file of the engine names a tool: `TestEngineNamesNoTool` examines this rule.
- `docs/design/`: the contract of the sync, one place for each rule (see "Where
  the rules are"). `docs/plans/`: the vwf plan index.
- `.config/mise/tasks/test/`: the tasks `test:repo` (`repo`), `test:mac`
  (`mac`), `test:macos-tools` (`macos-tools`),
  `test:go` (`go`) and `test` (`_default`). They are not in `repo/tools/` or
  `macos/tools/`, so no consumer gets them.
- `.github/workflows/release-engine.yml` and `cli/VERSION`: the release
  of the Go engine and the minimum version that a clone needs. A signed tag
  `bootstrap@X.Y.Z` on `main` starts the
  release, and the workflow writes `VERSION`. Do not edit `VERSION` by hand. See
  `docs/design/cli.md` ("The release of the engine").
- `.config/mise/tasks/vorssaint/diff`: the task `vorssaint:diff`. It compares
  a Vorssaint backup file with the defaults of `macos/tools/vorssaint/`, and
  exits 1 when a key differs or is missing.
- `macos/fixtures/`: the fixtures of `test:mac`.
- `macos/examples/macos.toml`: the sample values file of the Mac side, and the
  model of `test:macos-tools`. A change to a Mac tool also changes this file
  when the tool needs a value or a selection, because `test:macos-tools` fails
  for an unselected tool.

`repo/examples/<example>/` holds sample values files (`bootstrap.toml`), one
folder for each repository shape: `polyrepo`, `polyrepo-main`,
`polyrepo-develop`, `monorepo`, `repo-submodules` and `repo-siblings`.
`repo-submodules` also has one file in each submodule folder, and
`repo-siblings` has one file in each sibling folder only. The sync never
copies or generates them. `docs/sync/index.md` and `docs/sync/vwf.md`
use them as the model of a values file, so a change to the contract of the
values file changes the examples in the same commit.

Each example has an empty `bootstrap_commit`, so a consumer that copies an
example gets no wrong value: the sync replaces the empty line with the
commit of its sync. `test:repo` runs `bootstrap sync --dry-run` with each
example, in a consumer with its members. That case finds an unknown tool, a
wrong member and a wrong shape. It does not find an unknown key, because
`bootstrap sync` accepts one.

`docs/sync/` tells an AI agent how to adopt `bootstrap` in a consumer
repository, how to bring it up to date, and how to add and remove tools after
the adoption. After the adoption, `setup:configs` (`bootstrap sync`) keeps the
consumer in sync.

- `docs/sync/index.md`: the adoption prompt. It runs
  `bootstrap questions --json`, then `bootstrap init <tool…> --answer …`, and
  it brings an adopted repository up to date with `bootstrap init`.
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
bring the Mac up to date with `bootstrap mac init`. After the adoption, the task
`updateall:configs` (`bootstrap mac sync`) keeps the Mac in sync. The sync never
copies this folder.

- `docs/setup/index.md`: the Mac adoption prompt. It does the first adoption,
  and it brings a Mac up to date with `bootstrap_commit`.
- `docs/setup/migrate-dotfiles.md`: the method that moves a Mac from the symlink
  model of `macos-setup/dotfiles/` to `macos.toml` and the Mac sync.
- `docs/setup/tools.md`: the reference of the Mac tools, with one section for
  each folder in `macos/tools/`.

This repository is also a consumer of its own `base` pack and of the `claude`
tool. Its values file is `.config/bootstrap.toml`, and `setup:configs` writes
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
  `.config/bootstrap.toml`.
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
consumer. mise uses that path as the task name, for example `code/merge/develop`
is `code:merge:develop`. A `_default` file is the task for the directory name.
The `mise` tool ships each `_default`. A leaf task of a group that an
orchestrator runs with `own_tasks` (`code:format`, `code:lint`, `code:sec`,
`setup:deps`, `setup:secrets`) has a number prefix for its run order, for
example `code:lint:70-ruff`. See `docs/design/tools.md` ("The `[tasks]`
table").

- `_scripts/` contains libraries, not tasks. Each task sources them through
  `${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/<name>`.
  - `helpers`: the print functions (`print_header`, `print_success`,
    `print_wait`, and others), `members`, and `shell_files_in_scope`. Use these
    functions for all task output.
  - `checks`: git predicates. They return an exit status only and print
    nothing. The one exception is `get_main_worktree_path`, which prints the
    path of the main worktree.
  - `merge`: `merge_to_destination_branch`, which the `code:merge` tasks
    use.
  - `placeholder`: for a task that no stack fills yet. A placeholder task
    contains the `#PLACEHOLDER` marker and always exits 0. A stack overlay
    replaces the file. Do not edit a placeholder in place. No task in
    `repo/tools/` uses it now.
  - `xcode`: `require_xcode`, `find_xcodeproj` and the simulator functions for
    the Apple tasks.
  - `worktrees-page`: `write_worktrees_page`, which writes the HTML page of
    `code:worktrees`. It holds a small font with the glyphs of that task.

  The `mise` tool ships `helpers`, `checks`, `placeholder` and
  `worktrees-page`, the `git` tool ships `merge`, and the `swift` tool ships
  `xcode`.
- Prefix each tool call with `mise x --`, for example `mise x -- dprint check`.
  To find whether a tool is available, use `mise x -- which <tool>`, not
  `command -v`. System commands (`git`, `awk`, `sed`) and `mise` itself do not
  get the prefix. The tools that mise does not install also do not get it, for
  example `xcodebuild`, `swift`, `pod`, `./gradlew` and the VS Code `code`
  command.
- Exception: a tool that a task must have in every repository and environment
  gets a `#MISE tools.<tool>="latest"` directive in that task, and the task
  and the libraries that it sources call it without `mise x --` and find it
  with `command -v`. `mise x` finds only the tools of a config file, and a
  repository before its first sync, or a shell without `MISE_ENV=dev`, has no
  config that names them. The tools are `github:virajp/bootstrap` in
  `setup:configs`, `pre-commit` in `setup:precommit` and `code:precommit`,
  `pre-commit` and `yq` in the `code:merge` tasks, `yq` in `code:worktrees`,
  and `dprint` in `setup:dprint`. The tool calls of the engine in `cli/`
  are in
  `docs/design/cli.md` ("The tool calls").
- To find whether the repository has an optional task, use `has_own_task` and
  `own_tasks` from `helpers`. Do not use `mise tasks info` or a wildcard
  `mise run`. In a submodule, mise also finds the tasks of the root
  repository, and it runs them in the root directory.
- Write shell for bash 3.2 and bash 5.x. macOS has bash 3.2 as `/bin/bash`. Do
  not use namerefs, `${var,,}`, or associative arrays.
- Tasks use `#MISE` and `#USAGE` header comments for metadata and flags. A
  `#USAGE flag "--fix"` is available in the script as `usage_fix`.
- Under `set -e`, write a command substitution that can fail as
  `VAR=$(cmd) || VAR=''`. The existing scripts explain this in comments.
- Only the `init` task is in mise config (`[mise.tasks.init]` of
  `repo/tools/mise/tools.toml`, always dev). It makes the task scripts
  executable. The task files load in every environment.
- A gate of a language is a `code:lint:<gate>` or `code:format:<gate>` task
  in the tool of that language, for example `code:lint:70-ruff` in `python`.
  `code:lint:60-bootstrap` of `mise` follows the same rules for the values
  file `.config/bootstrap.toml`:
  - Without its tool (`mise x -- which <tool>`) or without a file of its
    language, the gate prints nothing and exits 0. Some gates examine a
    project file instead, for example `pubspec.yaml` for dart and `gradlew`
    for android. The task files load in every environment, also where the
    tool is not installed.
  - With a list of files, it uses only the files of its language. Without
    files, it lists them with `git ls-files --cached --others
    --exclude-standard`, because most linters do not read `.gitignore`. A
    tool that scans its project itself (ruff, pyrefly, detekt) uses that list
    only to find whether a file exists.
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
11. Exception to rule 2: when its output is not a terminal, `code:worktrees`
    prints only JSON Lines, because an agent reads each line.

## Tests

Each test task sets `BOOTSTRAP_VERBOSE=1`, so a case compares the per-file log
lines of the sync. The cases of the default output unset it. See
`docs/design/cli.md` ("The output of a run"). Each test task also sets
`BOOTSTRAP_CONFIG` to a missing path, so the config file of the user never
changes a case. See `docs/design/cli.md` ("The config file").

- `test:repo` is the task of this repository in
  `.config/mise/tasks/test/repo`. It runs `bootstrap sync --dry-run`,
  `bootstrap check` and, for a case that compares a written file,
  `bootstrap sync`, built from the working tree's `cli/`, with the
  working tree's `repo/` and `cli/` against scratch values files and
  scratch consumers. The init cases run `bootstrap init` on an old values
  file, and the clone cases run the binary against a local origin. The entry
  cases run the rendered task `setup:configs` with a fake engine. The config
  cases point `BOOTSTRAP_CONFIG` or `--config` at a file of the case, or at a
  missing file. Most of them run `bootstrap config`. The cases of a `verbose`
  key and of the key `clone` run `bootstrap sync --dry-run`. A case of the
  default output also runs `bootstrap sync --verbose`. With
  `BOOTSTRAP_BIN` set to an absolute path, it tests that binary and does not
  build one. It
  compares the exit code and the message of each case. It needs no network when
  gomplate, jq and yq are installed. Before its
  first `git` call, it unsets the `GIT_*` variables of a git hook, so its
  scratch repositories never write to this repository. It runs the cases in
  parallel, one job for each CPU, and each case in its own folder.
- `test:mac` is the task of this repository in
  `.config/mise/tasks/test/mac`. It runs `bootstrap mac sync`, built
  from the working tree's `cli/`, against the fixtures in
  `macos/fixtures/`, and it
  compares the exit code, the output and the files of each case. The entry cases
  run the task `updateall:configs` of `macos/tools/mise/`
  against a local bare origin, for the clone, the lock and the temporary copy.
  Each case has its own temporary `HOME`, so it never changes the real `$HOME`.
  It needs no network when gomplate, jq and yq are installed. With
  `BOOTSTRAP_BIN` set to an absolute path, this task and `test:macos-tools`
  test that binary and do not build one.
- `test:macos-tools` is the task of this repository in
  `.config/mise/tasks/test/macos-tools`. It runs `bootstrap mac sync`, built
  from the working tree's `cli/`, with each tool of `macos/tools/` and
  a copy of
  `macos/examples/macos.toml`, in a temporary `HOME`. It examines five items:
  no `[files]` table gives `mode = "seed"` to a JSON, YAML or TOML target, the
  example selects each tool and each rendered file parses, an example without
  the answer `config_path` of `mise` stops the sync with code 2, a
  mise tool on both sides of `repo/tools/` and `macos/tools/` has one version
  and one `depends`, a mise setting on both sides has one value, and the
  git tool needs one email only. To change a mise version, a `depends` or a
  mise setting, change it on both sides. An intended
  difference goes in `drift_allowed` of the task, with a reason. In the git
  cases, an empty `email_github` or an empty `email_gitlab` gets a first run,
  then a second run with exit 0. After it, `identity` has the other email, the
  config of the empty email has no email line, and `allowed_signers` has one
  signer line. Two empty emails stop the sync with code 2 and the message of
  the group `git.email`. It needs no
  network when gomplate, jq and yq are installed.
- `test:go` is the task of this repository in `.config/mise/tasks/test/go`.
  It runs `go vet`, a `gofmt` check and `go test` in the Go module of
  `cli/`. Some unit tests compare a case with a golden file in
  `testdata/`; see `docs/design/cli.md` ("The files of
  `cli/`"). `TestEngineNamesNoTool` fails when a string literal of the
  engine code names a tool, outside its allow list. The task runs the Go of the
  `go` tool, which this repository selects.
- `test` (`test/_default`) runs each test task in parallel. No hook and no
  `code:merge` task runs a test, because the tests are slow and a consumer does
  not have them. Run the tests by hand before you commit:
  - A change to `cli/`: `mise run test`.
  - A change to the Go code only: `mise run test:go`.
  - A change to `repo/`: `mise run test:repo`.
  - A change to `macos/`: `mise run test:mac` and
    `mise run test:macos-tools`.

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
   `mise run code:merge <branch>`, which merges and pushes.
4. In this repository and in each consumer repository, run
   `mise run setup:configs`, then commit the result. This repository sets no
   `direct_commit`, so the `no-commit-to-branch` hook refuses a commit on
   `main`. Thus in this
   repository, sync on a local `sync-<name>` branch, commit
   `ops: sync bootstrap with <change>`, and land it with
   `mise run code:merge sync-<name>`. The sync also updates
   `bootstrap_commit`, which stays one merge behind `main`.
