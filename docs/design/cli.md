# Sync engine design

This document holds the rules that both sides of the sync share: the command
line, the cache and the clone, the version check, the engine variables, the
config file, the tool calls, the scripts of a tool, the templates, the write of
one target, shared files, the header, the manifest and the stale rule, the
values, the `bootstrap_commit` key, the output, the exit codes and the release
of the engine. It is the one place of these rules.

See also: [packs.md](packs.md), [tools.md](tools.md), [mise.md](mise.md),
[repo.md](repo.md), [mac.md](mac.md).

## The engine knows features, not tools

The engine is the CLI `bootstrap`. Each tool describes itself in its
`tools.toml`, with the optional folders `files/`, `tasks/` and `scripts/`. See
[tools.md](tools.md). No Go file of the engine names a tool. The test
`TestEngineNamesNoTool` of `internal/source` examines this rule: no string
literal of the engine code names a tool folder, except the words of an exact
allow list (the old forms in `internal/oldform`, the old Mac header in
`internal/header`, the clone URL and some message words).

The engine knows only its frame:

- mise: it generates the mise config of each tool and writes it first, it runs
  `mise x`, and on the Mac it needs the standalone mise. It runs
  `mise bootstrap --yes` only for the first setup of a Mac.
- git: the repository, the members, the clone and `bootstrap_commit`.
- The values folder `.config/bootstrap/` (the Mac values file and its
  overrides file), its schema, the packs, the manifest, the header,
  the symlink rules, the stages and the exit codes.

Each tool installs only through mise: `[mise.tools]`, and the `brew:` and
`brew-cask:` entries of `[mise.bootstrap.packages]` on the Mac. bootstrap
touches only the files of the tools that a values file selects. Each other
file belongs to the user. Each sync handles only its own repository, never a
submodule: each submodule runs its own sync.

## The files of `cli/`

`cli/` is not a tool. The sync never copies it to a consumer. It is the
Go module `github.com/virajp/bootstrap/cli`.

| Path                 | Job                                                                                                                                                                                                                                                                                                                     |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `VERSION`            | The minimum version of the engine that a clone of `main` needs. See [The version check](#the-version-check).                                                                                                                                                                                                            |
| `go.mod`             | The module.                                                                                                                                                                                                                                                                                                             |
| `cmd/bootstrap`      | The command `bootstrap`. It calls `internal/cli`. `version` prints `main.version`, which the release build sets. A local build has `dev`.                                                                                                                                                                               |
| `internal/cli`       | The command line: it parses the arguments, calls `internal/core`, prints the results and the errors, and asks the questions and the consent, or starts `internal/tui`.                                                                                                                                                  |
| `internal/core`      | The preflight, the clone and its lock, the version check, the change list (`Plan`, `Apply`, `Effect`), the change and the state of `migrate` (`InitChange`, `InitState`), the scan of `detect` (`DetectState`) and its projects, the kind of a Mac tool (`KindOf`), and the clean-target rule. It never prints or asks. |
| `internal/contract`  | The JSON contract: the envelope, its version and the names of the task ports. See [The JSON contract](#the-json-contract).                                                                                                                                                                                              |
| `internal/tui`       | The TUI. See [The TUI](#the-tui). It calls `internal/core`, and never `internal/cli`.                                                                                                                                                                                                                                   |
| `internal/config`    | The user config file of the engine: its path, its keys and its checks. See [The config file](#the-config-file).                                                                                                                                                                                                         |
| `internal/oldform`   | The old forms of a values file, and their migration. It is the one place of the old names.                                                                                                                                                                                                                              |
| `internal/questions` | The questions of the selected tools, their proposals and their answers.                                                                                                                                                                                                                                                 |
| `internal/field`     | The text field of a question: the question, its note, and a field that starts with the proposal. It uses Bubble Tea v2 and the `textinput` of Bubbles v2.                                                                                                                                                               |
| `internal/scripts`   | The run of the scripts of a tool. See [The scripts of a tool](#the-scripts-of-a-tool).                                                                                                                                                                                                                                  |
| `internal/fail`      | The typed errors of the exit codes, and a collector where the highest code wins. See [The exit codes](#the-exit-codes).                                                                                                                                                                                                 |
| `internal/report`    | The output of a run. See [The output of a run](#the-output-of-a-run).                                                                                                                                                                                                                                                   |
| `internal/source`    | The read of the clone: the packs file, the categories file, the `tools.toml` of each tool, the entries and the files of a tool folder, and the checks of these files.                                                                                                                                                   |
| `internal/projects`  | The projects of a repository: the `[projects.<name>]` tables, the names, the workspace manifests and the folders that git does not ignore. `detect` and `check` use it. See [`detect`](#detect).                                                                                                                        |
| `internal/values`    | The values files and the overrides files of both sides, and the values merge. See [The values of a tool](#the-values-of-a-tool).                                                                                                                                                                                        |
| `internal/selection` | The checks of the packs, the required members of `base`, the depth-first walk of the `depends`, and the stages of the Mac.                                                                                                                                                                                              |
| `internal/tmpl`      | `text/template` with the delimiters `<%` and `%>`, and our own copy of the gomplate functions that the templates use. It does not import gomplate.                                                                                                                                                                      |
| `internal/format`    | The parse check, and the merge of a shared file: JSON, JSONC, YAML, TOML and text. See [Shared files](#shared-files).                                                                                                                                                                                                   |
| `internal/header`    | The header of a synced file and of a shared file, and its detection. See [The header](#the-header).                                                                                                                                                                                                                     |
| `internal/target`    | The write of one target, the modes, `bootstrap_commit`, and the tables of the required members of `base`. See [The write of one target](#the-write-of-one-target).                                                                                                                                                      |
| `internal/tomledit`  | The line editor of a values file, which keeps each other line. See [The layout of a values file](#the-layout-of-a-values-file).                                                                                                                                                                                         |
| `internal/manifest`  | The manifest and the stale rule, in the link modes `repo` and `macos`. See [The manifest](#the-manifest-and-the-stale-rule).                                                                                                                                                                                            |
| `internal/misegen`   | The mise config generator of both sides. See [mise.md](mise.md).                                                                                                                                                                                                                                                        |
| `internal/repo`      | The repo front-end, the sync of `bootstrap sync`. See [repo.md](repo.md).                                                                                                                                                                                                                                               |
| `internal/macos`     | The Mac front-end, the sync of `bootstrap mac sync`. See [mac.md](mac.md).                                                                                                                                                                                                                                              |

The task `test` of this repository runs `go test` in the module, and the
task `test:go` runs `go vet` and a `gofmt` check. Some unit tests compare the
output of a case with a
golden file in `testdata/`. `UPDATE_GOLDEN=1 go test ./internal/<package>`
writes the golden files of `selection`, `misegen`, `source`, `tmpl`, `manifest`
and `report` again.
The golden files of `internal/tmpl` hold the output of the gomplate CLI for the
template functions and the inline templates.

## The command line

```text
bootstrap <command> [<kind>] [<name>…] [flags]
bootstrap mac <command> [<kind>] [<name>…] [flags]
```

The group `mac` names the Mac side. Thus a folder with both `macos.toml` and
`.config/bootstrap/` has no doubt. Each command without `mac` is for the
repository of the current folder. The `mac` commands have the same grammar.
The current folder is `PWD` when it names the current folder, else the physical
path, as `$PWD` of bash. A parent that changes the folder and keeps its own
`PWD`, for example `mise x --cd=<dir>`, thus gives `<dir>`.

The kind word tells what the names are: `tool` or `pack`. `add` takes `tool` or
`pack`. `remove` and `replace` take `tool` only. The other commands take no
kind word.

| Command                    | Job                                                                                                                                                                                                                                                                                                                                       |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `migrate`                  | Migrates an old form, or creates the values folder with a table for each member of `base`, removes the known old copies of the test parts, moves the own test tasks into parts, renames the old command of the worktree setup, asks each question without an answer, then syncs (`mac migrate` does not sync). See [`migrate`](#migrate). |
| `tui`                      | Opens the TUI on the values. It needs a terminal, and stops when `migrate` has a change to make. See [The TUI](#the-tui).                                                                                                                                                                                                                 |
| `sync`                     | The sync of the side. When `migrate` has a change to make, `sync` does the `migrate` first. `setup:configs` and `updateall:configs` run it. See [`migrate` before `sync`](#migrate-before-sync).                                                                                                                                          |
| `check`                    | Examines the values folder against its schema and runs the check scripts. On the Mac, a dry run of the sync. `--json` reports each stage and its findings.                                                                                                                                                                                |
| `questions [tool…]`        | Lists the questions, their proposals and their answers. Without names, the questions of the selected tools. `--json` prints the envelope with the questions in `data`.                                                                                                                                                                    |
| `add tool <tool…>`         | Selects the tools and their dependencies, asks their questions, then syncs.                                                                                                                                                                                                                                                               |
| `add pack <pack…>`         | Selects each member of the packs that is not selected, and their dependencies, asks their questions, then syncs. See [`add` and `remove`](#add-and-remove).                                                                                                                                                                               |
| `remove tool <tool…>`      | Removes the tools, and each dependency that no other tool needs, then syncs. Each seed of a removed tool stays, with a line.                                                                                                                                                                                                              |
| `replace tool <old> <new>` | Runs the migrate script of `<new>` for `<old>`, swaps the tools, then syncs. See [`replace`](#replace).                                                                                                                                                                                                                                   |
| `detect`                   | Runs the detect scripts, or on the Mac the package scan, and lists each tool that matches. See [`detect`](#detect).                                                                                                                                                                                                                       |
| `list`, `ls`               | Prints the selected tools as a tree. See [`list`](#list). It writes nothing.                                                                                                                                                                                                                                                              |
| `config get <key>`         | Prints the value of a key of `.config/bootstrap/`, with its source. See [`config`](#config). There is no `mac config`.                                                                                                                                                                                                                    |
| `config set <key> <value>` | Writes a key of `.config/bootstrap/`, and changes nothing else. See [`config`](#config).                                                                                                                                                                                                                                                  |
| `config unset <key>`       | Removes a key of `.config/bootstrap/`. See [`config`](#config).                                                                                                                                                                                                                                                                           |
| `config show [<file>]`     | Prints `repo.toml`, `tools.toml` or `releases.toml`, or the three files as one table by file name. See [`config`](#config).                                                                                                                                                                                                               |
| `settings`                 | Prints the config file and each setting, with its value and its source. See [The config file](#the-config-file). There is no `mac settings`.                                                                                                                                                                                              |
| `contract`                 | Prints the versions of the JSON contract, the version of the engine, the commands that take `--json` and the names of the task ports. It needs no repository, no config file and no clone.                                                                                                                                                |
| `catalog`                  | Prints the tools of the clone by category. `--json` also gives the categories, the questions and defaults of each tool, and the packs. It needs no repository.                                                                                                                                                                            |
| `describe`                 | Prints the values of `repo.toml`, the selected tools with their values, the members, the task ports of the repository, each port present or missing. Only `--json` also gives its language servers, in `lsp`.                                                                                                                             |
| `lsp`                      | Prints the language servers of the selected tools of the repository, one line each: the server, its tool, its command and its extensions. See [`lsp`](#lsp).                                                                                                                                                                              |
| `version`, `--version`     | Prints the version. There is no `mac version`.                                                                                                                                                                                                                                                                                            |

| Flag                            | Commands                                                                                                                                           | Effect                                                                                                                                   |
| ------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `--dry-run`                     | `migrate`, `sync`, `add`, `remove`, `replace`, `detect`                                                                                            | Prints the change list, and writes nothing.                                                                                              |
| `--yes`, `-y`                   | `migrate`, `add`, `remove`, `replace`, `detect --add`                                                                                              | Writes the changes without a prompt.                                                                                                     |
| `--add`                         | `detect`                                                                                                                                           | Adds each matched tool that is not selected, after the consent.                                                                          |
| `--exclude <member>`            | `add pack`                                                                                                                                         | Leaves out a member of the pack, as the packs file writes it. The flag can repeat. It needs exactly one pack name.                       |
| `--backup <path>`               | `mac migrate`                                                                                                                                      | The backup clone: the key `backup` of `macos.toml`. See [mac.md](mac.md#the-backup-clone).                                               |
| `--answer <tool>.<key>=<value>` | `migrate`, `sync`, `add`, `replace`, `detect --add`                                                                                                | The answer of a question. The flag can repeat. `--answer <tool>.<key>=` gives an empty answer, only for an optional or a group question. |
| `--json`                        | `contract`, `catalog`, `describe`, `lsp`, `check`, `questions`, `mac questions`, `config`, `migrate`, `sync`, `add`, `remove`, `replace`, `detect` | Prints one JSON envelope on stdout, and each other line on stderr, for an AI agent. See [The JSON contract](#the-json-contract).         |
| `--format json\|yaml\|toml`     | `config get`, `config show`                                                                                                                        | The output format. The default is `json` when stdout is not a terminal, else `toml`. With `--json`, it is a usage error (code 2).        |
| `--no-bootstrap`                | each `mac` command except `questions`                                                                                                              | Skips the `mise bootstrap` of the first setup of a Mac. See [mac.md](mac.md#the-first-setup-and-exit-4).                                 |
| `--jobs <n>`                    | each command except `questions`, `list`, `config`, `contract`, `catalog`, `describe`, `lsp` and `version`                                          | The jobs at a time. The default is 2 × the CPU count.                                                                                    |
| `--verbose`                     | `migrate`, `tui`, `sync`, `check`, `add`, `remove`, `replace`, `detect`                                                                            | Sets `BOOTSTRAP_VERBOSE=1`. See [The output of a run](#the-output-of-a-run).                                                             |
| `--config <path>`               | each command except `version` and `contract`                                                                                                       | The config file. See [The config file](#the-config-file).                                                                                |
| `--cache-dir <path>`            | each command except `version` and `contract`                                                                                                       | The cache folder. It sets `BOOTSTRAP_CACHE`.                                                                                             |
| `--url <url>`                   | each command except `version` and `contract`                                                                                                       | The clone URL. It sets `BOOTSTRAP_URL`.                                                                                                  |
| `--clone <path>`                | each command except `version` and `contract`                                                                                                       | A clone of bootstrap, with no lock and no fetch. It sets `BOOTSTRAP_CLONE`.                                                              |

A flag of a repo command also applies to the same `mac` command, for example
`mac sync --verbose`. A flag with a value takes `--flag value` and
`--flag=value`. An empty value is a usage error. The one exception is the
empty answer `--answer <tool>.<key>=`. See [The questions](#the-questions).

A flag that a command does not take, an unknown command, and `add`, `remove`
or `replace` without a name are usage errors (code 2). `add`, `remove` and
`replace` without a kind word stop with code 2 and the hint of the new form,
for example `use: bootstrap add tool node`. `migrate` with names stops with code
2
and the hint `use: bootstrap add tool <names>`. A usage error prints the
problem, the usage line of that command only, and a pointer to the help. See
[The output of a run](#the-output-of-a-run). `bootstrap --help` and
`bootstrap <command> --help` print the full help on stdout, in the groups
`Repository`, `Mac` and `Flags`. With `--json` anywhere in the arguments, the
help goes to stderr, and a usage error also prints the envelope on stdout. See
[The JSON contract](#the-json-contract).

### The preflight

Before the preflight, each command except `version` and `contract` reads the
config file. An error of the file stops the command with code 2. See
[The config file](#the-config-file).

Each command except `version`, `settings`, `contract` and `catalog` then
examines these items. A failure stops it with code 2. `config` examines only
the git repository and the values folder.

- `mise` is on `PATH`.
- On the repo side, the folder is in a git repository: `MISE_PROJECT_ROOT`,
  else the git top level of the current folder. The Mac side needs no git
  repository.
- The values are the folder `.config/bootstrap/` of the repository, or the
  folder of `BOOTSTRAP_VALUES`. On the Mac, they are the fixed file
  `~/.config/bootstrap/macos.toml` (below `$XDG_CONFIG_HOME` when it is set),
  from any folder: see [mac.md](mac.md#the-place-of-macostoml).

`migrate` outside a git repository tells the user to create one with `git init`
(the repo side only).
On
the repo side, each command except `migrate`, `contract` and `catalog` needs the
values folder with its `repo.toml` and `tools.toml`: without them, it stops with
code 2 and the fix "run bootstrap migrate". `add`, `remove`, `replace` and
`detect`
need the values on both sides. The old values file `.config/bootstrap.toml`
stops each command except `migrate`, `version`, `settings`, `contract`,
`catalog`
and the `mac` commands with "run bootstrap migrate". The old file next to the
folder stops each command except `version`, `settings`, `contract`, `catalog`
and the `mac` commands, `migrate` included, with the fix "remove one: compare
.config/bootstrap.toml with .config/bootstrap/". A values file in an old form
stops each command except `migrate` with "run bootstrap migrate" ("run bootstrap
mac
migrate" on the Mac). An unknown key, a wrong type or a value outside its enum
in
the values folder stops each command that reads the values with code 2. See
[bootstrap-config.md](bootstrap-config.md#the-schema). `sync` and `check` also
stop for a question of a selected tool without an answer, and for a group of
questions without an answer that is not empty. `check` shows a warning for each
tool with `selectable = false` that the `[tools]` table of `tools.toml` names:
`<tool> is not selectable - remove [tools.<tool>] from <file>; a tool that needs
it selects it`. The warning does not change the exit code. `mac check` does the
same.

### The change list

Each writing command (`migrate`, `add`, `remove`, `replace`, `detect --add`)
builds one change list: the items `add tool`, `add pack` (with its
`--exclude` members), `remove tool`, `replace tool` and the answers, and on the
Mac the packages of `mac detect --add`. `internal/core` has these functions
for it, and the TUI calls the same functions:

- `Plan` applies the items in order to a temporary copy of the values (a
  temporary values folder on the repo side), examines the answers, finds the
  questions without an answer, and runs a quiet dry run of the front-end with
  the new values, with `BOOTSTRAP_VALUES` set to the temporary folder. It
  gives the new values, the migrations, and each file that the sync writes,
  changes or removes. It writes nothing. A command that changes the values
  also adds the missing tables of the required members of `base`, so
  `tools.toml` gets one write.
- `Apply` runs the script that comes before the writes, writes the values and
  the migrations, and runs the sync, with `BOOTSTRAP_VALUES` set to the values
  folder. Thus each command writes one time and syncs one time. The move of
  `migrate` (see [`migrate`](#migrate)) moves its files before the sync. When
  the sync
  fails, `Apply` moves those files back, puts the values back as they were
  before the command, or removes a file that did not exist. After a failed
  migration it puts back `.config/bootstrap.toml`,
  `.config/bootstrap.overrides.toml` and `.config/repo-generated.tsv`, and
  removes `.config/bootstrap/`. The other files stay as the sync left them. The
  one exception is code 4 of the Mac (a new terminal is necessary): the sync
  did its work, so the values file stays. `mac migrate` runs no sync: on a Mac
  with no manifest, `Apply` runs the first setup in its place, else nothing.
  `Apply` also gives the outcome of each target of the sync, also when the
  sync fails, for the `data` of `--json`. The front-ends give them from
  `RunWith`.
- `Effect` applies the items in order to the selection, as `Plan` does, and
  gives the tools that they add and remove. It writes nothing, and runs no
  script and no sync. Only the TUI calls it, after each change of its list.

- An empty change list prints `✔ Nothing to change` and exits 0. When the
  command has a reason, for example that a tool is already selected, the line
  adds it after `·`: `✔ Nothing to change · python is already selected`.
- `--dry-run` prints the list, names each dirty target, and writes nothing.
- Else the command applies the clean-target rule, asks for consent, writes the
  values and the migrations, and then runs the sync.

`sync` has no consent step. It builds the same plan with a quiet dry run, on
the Mac on each real run, and on the repo side when the work tree has a change
outside the values folder, and on each `sync --json`. Then it applies the
clean-target rule and writes.

### The consent

`--yes` writes the changes. Else the command asks at a prompt in a terminal:
`? Write <n> changes? [y/N]` (`1 change` for one). A "no" writes nothing and
exits 0 with `· Nothing changed`. Without a terminal, and without `--yes`, the
command stops with code 2: `changes need --yes`.

`--json` never prompts. `migrate`, `add`, `remove`, `replace` and `detect --add`
with `--json` also need `--yes` or `--dry-run`, else they stop with code 2:
`<command> --json needs --yes or --dry-run`. `sync --json` applies with no
`--yes`, as `sync` does, and a plain `detect --json` only reads. See
[The JSON contract](#the-json-contract).

### The clean-target rule

git is the history: the engine makes no backup and no commit. Thus a command
does not write when a target has changes that are not committed (modified,
staged or untracked). It names each such target and stops with code 2, and it
writes nothing.

- The rule applies only to a target inside a git work tree.
- `--dry-run` never stops. It names each dirty target in a warning line.
- `sync` does not apply the rule to the values folder, because it writes only
  `bootstrap_commit`, `[home]`, the manifest, the missing tables of the
  required members of `base` and the `mise` template
  `.config/bootstrap/config-files.toml` there. `migrate` applies the rule to the
  values.
- `replace` needs a clean work tree, because its migrate script can change any
  file.

### `migrate`

`migrate` is the first adoption, and also the update of an adopted repository or
Mac. It takes no tool or pack names: `add tool` and `add pack` select tools.
`init` is the old name of `migrate`, and stays as a hidden alias: `init` and
`mac init` run `migrate` and `mac migrate`. The help and the commands of
`contract` name only `migrate`, and the JSON envelope names `migrate`.

1. On the repo side, a new values folder gets `tools.toml` and `repo.toml`,
   each with a header comment and `format = 3`. `repo.toml` also gets
   `core_branches`, `environments = [ "dev" ]` and `repo_name`, the name of the
   folder, because the templates need it. `core_branches` is the `main` and
   `develop` that exist, else the current branch, born or unborn. With neither
   and a detached `HEAD`, `migrate` stops with code 2 and the fix "check out a
   branch, then run bootstrap migrate again". `tools.toml` gets a table for each
   member of `base`, as `add pack base` does. A folder without one of the two
   files gets the missing file. A new Mac values file starts with a header
   comment and `format = 2`.
2. An old `.config/bootstrap.toml` gets the migration of each old form, then
   the split into `.config/bootstrap/`. See
   [bootstrap-config.md](bootstrap-config.md#old-forms) for the old forms. On
   the Mac, an existing values file gets the migration of each old form, with
   its comments kept. Only `migrate` migrates: `sync` runs `migrate` first (see
   [`migrate` before `sync`](#migrate-before-sync)), and `check` stops with
   code 2 and "run bootstrap migrate" ("run bootstrap mac migrate" on the Mac)
   when
   they
   find an old form. A later `migrate` does not add a removed member of `base`
   again.
3. On the repo side, `migrate` removes each known old copy of a part of a
   selected tool: a regular file at the old path of the group of the part, for
   example `test/golden`, or at `<group>/50-local`, whose sha256 is in the
   `replaces` of the part. See
   [tools.md](tools.md#the-tasks-table). The stackgen packs wrote such copies,
   for example the old `test/golden` task of the compose pack. Each tool with
   copies is one change, `remove old copies of the <tool> test parts`, and its
   detail lists each file. It comes before the move of step 4, so the move
   does not move a removed file. A file with another hash stays, and so does a
   copy of a tool that is not selected: step 4 moves it as a task of the
   repository. `migrate` follows no link. On a failed sync, `migrate` writes
   each
   removed file back. `sync` runs `migrate` first, and `check` stops with code 2
   and "run bootstrap migrate" on a known copy at `<group>/50-local`. At the old
   path, the stop of
   step 4 for a task at the place of a test port comes first.
4. On the repo side, `migrate` moves each task of the repository at the place of
   a test port to the part `50-local` of its group, so the sync can write the
   dispatcher there. These are the paths below `.config/mise/tasks/`:

   | Path                                  | Moves to               |
   | ------------------------------------- | ---------------------- |
   | `test`, `test/_default`               | `test/50-local`        |
   | `test/e2e`, `test/e2e/_default`       | `test/e2e/50-local`    |
   | `test/golden`, `test/golden/_default` | `test/golden/50-local` |

   A path is a task of the repository when it is a regular file, the manifest
   does not list it, and it has no bootstrap header. A file with the header is
   bootstrap's own, and `migrate` does not move it (see the next paragraph).
   `migrate` follows
   no link: a path that is a link, or is below a linked folder, stays, and so
   does a path whose target is. All the moves are one change,
   `move own test tasks into 50-local parts`, and its detail lists each file
   and its target. A file moves through the temp name `<path>.bootstrap-move`
   next to it, so a file can become a part of the folder at its own path.
   `migrate` examines each path before it moves a file. When a target
   `<group>/50-local` exists, or a temp name exists, it stops with code 2,
   names both paths, and moves nothing. `sync` runs `migrate` first, and
   `check` stops with code 2 and "run bootstrap migrate" on such a task.

   Each dispatcher is a `.py` file, for example `test/_default.py`. mise runs
   the file without the extension first, so an old copy `<dir>/_default`
   hides the new dispatcher. On the repo side, `migrate` removes each old copy
   of a dispatcher: a regular file `<dir>/_default` with the bootstrap header,
   which the manifest does not list, in a folder of a dispatcher
   (`code/format`, `code/lint`, `code/sec`, `setup/deps`, `setup/secrets`,
   `test`, `test/e2e` and `test/golden`), when a selected tool ships
   `<dir>/_default.py`. All the removals are one change,
   `remove old copies of the dispatchers`, after the changes of step 3, and
   its detail lists each file. `migrate` follows no link. The sync removes a
   listed copy by its manifest. Only `migrate` acts on an unlisted copy:
   `core.MigrateNeeded` does not count it, so `sync` and the TUI do not run
   `migrate` for such a copy alone, and `check` does not report it.
5. On the repo side, `migrate` renames the old task `setup:worktree` to
   `git:worktree:setup` where the sync cannot. The seed `.config/wt.toml` and
   each `[members.<name>].worktree_install` of `repo.toml` can run
   `mise run setup:worktree`, and `migrate` writes `mise run git:worktree:setup`
   in its place. Only that command text changes, and `repo.toml` keeps its
   comments. Both files are one change, `rename setup:worktree`, and its detail
   lists each file and key. On a failed sync, `migrate` writes both files back.
   `sync` runs `migrate` first, and `check` stops with code 2 and "run
   bootstrap migrate" when it finds the old command in either place.
6. The command asks each question of a selected tool that has no answer. See
   [The questions](#the-questions).
7. It plans the sync, shows the change list, and after the consent writes the
   changes and syncs. `mac migrate` does not sync: on a Mac with no manifest it
   runs the first setup, and it ends with `Next: run bootstrap mac sync`. See
   [mac.md](mac.md#the-first-setup-and-exit-4).

`migrate` never opens the TUI, in a terminal or not. A person and an agent
run the same steps. In a terminal, the command asks each question without an
answer in a field, and asks for the consent. `--no-tui` is an old flag of
`migrate`: the command takes it, and it does nothing. The help does not name
it. `migrate` does not run the detect scripts.

### `migrate` before `sync`

A user does not know when a repository needs a `migrate`. Thus `sync` and
`mac sync` first examine whether `migrate` has a change to make
(`core.MigrateNeeded`):

- an old form of the values, or a values folder without `repo.toml` or
  `tools.toml`;
- a known old copy of a test part, an own test task to move, or the old
  worktree command;
- in a terminal only, a question of a selected tool without an answer. Without
  a terminal, the sync names such a question, with "run bootstrap migrate".

This examination runs no script and no dry run of the sync, so a `sync` with
nothing to migrate is as fast as before. A repository without values, and
without an old form, is not adopted: `sync` does not migrate it, because the
adoption is an explicit `migrate`.

When `migrate` has a change to make, `sync` prints the clone line and the info
line `bootstrap migrate has a change to make: it runs before the sync`. Then it
does the steps of `migrate`, with no consent, because `sync` writes with no
consent. In a terminal, it asks each question without an answer. Without a
terminal, a question without an answer of a tool that the migration selects
stops it with code 2, with the hint `--answer`. `sync --answer` is a usage error
(code 2) when `migrate` has no
change to make. `sync --dry-run` shows the change list and writes nothing.
`sync --json` keeps the plan and the outcomes in `data`, as `migrate --json`.
`mac migrate` runs no sync, so after it `mac sync` runs its own sync, unless
the migrate was a first setup that stopped with exit 4.
`check` does not migrate: it stops with code 2 and "run bootstrap migrate" on
an old form.

### The TUI

`tui` and `mac tui` open the full-screen TUI on the values. The TUI is for a
person; the command line is for an agent. `migrate` never opens it.

The TUI builds the change list of [The change list](#the-change-list) one key
at a time, with the same items. The review calls `Plan`, and Apply calls
`Apply`. The TUI gives each answer to `internal/core` directly, so the
questions of the command line never ask under the TUI. Apply applies the
clean-target rule: it names each dirty target, writes nothing, and keeps the
change list.

- `tui` and `mac tui` take the flags of `migrate` and `mac migrate` except
  `--yes`, `--dry-run`, `--answer` and `--json`. They take no names.
  `mac tui` opens the fixed `macos.toml`.
- Without a terminal, `tui` stops with code 2 and the hint
  `use: bootstrap add tool, bootstrap remove tool or bootstrap migrate` (the
  `mac` commands on the Mac).
- When `migrate` has a change to make (see
  [`migrate` before `sync`](#migrate-before-sync)), `tui` stops with code 2
  before the TUI opens: `bootstrap migrate has a change to make`, with the hint
  `run bootstrap migrate, then bootstrap tui again` (the `mac` commands on the
  Mac). A missing values folder and an old form stop it as before.
- A terminal smaller than 100 × 25 at the start stops `tui` and `mac tui` with
  code 2: `<command> needs a terminal of 100 × 25 or larger`.
- The screen fills the terminal. A terminal larger than 100 × 25 keeps the
  proportions of the panes: the left pane keeps 30 % of the width, and the top
  panes and the bottom pane share the extra lines 11 to 7. In a row of the
  left pane, the name gets the extra columns, and the counter stays at the
  right edge. The columns of a row of the right pane keep their proportions.
- The key line starts with `r apply`, bright green when the change list has a
  change. `r` opens the review, and the review starts its key line with
  `enter apply`, also bright green.
- A slow step shows a spinner. Before the TUI opens, the fetch of the clone
  shows `Updating the clone of bootstrap ...` on stderr, as in each command
  (see [The progress](#the-progress)). The plan of the review, and the plan
  again after the
  questions of Apply, show the dialog `Planning ...`. It takes no key but
  `ctrl+c`, which quits.
- `tui` without the values folder stops with code 2: `<folder> is not there`,
  with the hint `run bootstrap migrate first`. An old form of the values stops
  it
  with code 2, as it stops `sync`.
- A quit by the user after the TUI opened exits 0, also after a failed sync:
  the TUI shows the failure. Each error before the TUI opens keeps its code.

### The questions

A question is a `[questions.<key>]` table of a `tools.toml`. See
[tools.md](tools.md#the-defaults-questions-and-contributes-tables). `migrate`
asks
the questions of each selected tool, and `add` the questions of the tools that
it adds.

- An answer is a value of `[tools.<tool>]` or of `[overrides.tools.<tool>]` in
  `tools.toml` (on the Mac, the values file or its overrides file). A question
  with an answer gets no prompt. An empty string
  `""` is an answer, and it means no answer.
- `--answer <tool>.<key>=<value>` gives the answer, and that question gets no
  prompt. An `--answer` for a question that no selected tool has is a usage
  error. When `--answer` names one question two times, the last one wins.
- `--answer <tool>.<key>=` writes `""`, so a later `migrate` does not ask the
  question. It is only for an optional or a group question. For another
  question, it is a usage error (code 2):
  `--answer <tool>.<key>=: an empty answer is only for an optional or a group
  question`.
- In a terminal, the command asks only the questions that have no key in the
  files and no `--answer`. It also asks an optional question. If `--answer`
  gives 3 of 4 answers, it asks the fourth.
- The prompt shows `? <question> (<tool>.<key>)`, the `note` of the question on
  the next line, and a text field. The field starts with the proposal: the
  output of the `detect` command of the question, else its `example`. The user
  can edit or delete the text, and Enter accepts it. An empty text is accepted
  for an optional or a group question, and the command writes `""`. For
  another question, the field stays open and shows `an answer is necessary`.
  ctrl+c stops the command with code 2, and it writes nothing. The field uses
  Bubble Tea v2. The consent prompt `? Write <n> changes? [y/N]` stays a plain
  line.
- A group needs at least one answer that is not empty. The command also asks
  each question of a group whose answers in the files are all empty or
  missing. After the last question of a group, if all its answers are empty,
  the command shows `Give at least one answer: <key>, <key>` and asks again
  each question of the group, also the questions that an `--answer` gave, each
  field started with its proposal. It asks again until the group has an answer
  that is not empty.
- With `--yes`, or without a terminal, the command does not accept a proposal:
  a question without a key in the files and without an `--answer` stops the
  command with code 2. The hint is
  `give it with --answer <tool>.<key>=<value>`, and for an optional or a group
  question it adds `, or leave it empty with --answer <tool>.<key>=`. A group
  passes when one of its answers in the files or in the flags is not empty.
  Else the command stops with code 2: `the group <tool>.<group> has no answer`,
  with the hint
  `give at least one with --answer <tool>.<key>=<value>: <key>, <key>`.
- `questions` lists the questions in one group for each tool. Each question has
  its key, its text, and its answer: `✔ <answer>`, or `✗ no answer`. An answer
  `""` shows as `✔` alone. An optional question without an answer shows
  `· no answer (optional)`. A question of a group without an answer, in a group
  that has an answer, shows `· no answer (group <group> has an answer)`. Each
  question of a group without an answer that is not empty shows `✗ no answer`.
  A line `· proposal (<source>): <proposal>` follows only when the question has
  a proposal that is not the answer. The source is `detect` or `example`. A
  value longer than 40 characters shows its first 20 characters, `…` and its
  last 4 characters. The last line counts the questions:
  `· 6 questions · 5 answered · 1 without an answer` (a zero part is absent). A
  question with `· no answer` is not in the count of `without an answer`.
  Without a question, the output is `✔ No questions`. `--json`, also of
  `mac questions`, prints the envelope of
  [The JSON contract](#the-json-contract). Its `data` is a list with one item
  for each question: `tool`, `key`, `question`, `proposal` and `source` (`null`
  without a proposal), `answer` (`null` without an answer), `group` (a string,
  `""` without a group), `note` (a string) and `optional` (a boolean). The
  `detect` command runs in the repository, or in `$HOME` on the Mac.

```text
git
  email_github   Which email do your GitHub commits use?
                 ✔ me@users.noreply.github.com
  signing_key    Which SSH public key signs your commits?
                 ✔ ssh-ed25519 AAAAC3Nz…ICGo
                 · proposal (detect): ssh-ed25519 AAAAC3Nz…CBSM ssh-digitalocean

1password
  ssh_vault      Which 1Password vault holds your SSH keys?
                 ✗ no answer
                 · proposal (example): Personal

· 3 questions · 2 answered · 1 without an answer
```

### `add` and `remove`

`add tool` writes an empty `[tools.<name>]` table for each named tool and each
dependency that is not selected yet. It refuses some named tools: see
[Named tools that a command refuses](#named-tools-that-a-command-refuses). A
tool that is already selected gets no table. The reason
`<tool> is already selected` shows in the last line when the command has
nothing else to do.

`add pack <pack…>` writes a table for each member of the packs that is not
selected yet, and for each new dependency. A nested pack `pack:<name>`
expands the same way. The values keep no record of the pack. See
[packs.md](packs.md).

- `--exclude <member>` leaves out a member, with its key as the packs file
  writes it, also a nested pack (`pack:<name>`) or a tool that a nested pack
  adds. The flag needs exactly one pack name. A name that is not a member, a
  required member, and a member that another added member needs stop the
  command with code 2.
- It leaves out, with an info line, a member whose `hosts` do not name the
  host, a member with `selectable = false`, and in a submodule member a member
  with `root_only`.
- A pack whose members are all selected writes nothing:
  `<pack> is already selected`.
- There is no `remove pack`.

`remove tool` acts only on `[tools.<name>]` tables. It refuses each of these
with code 2:

- A required member of `base`.
- A tool that is not selected.
- A tool that another selected tool needs.

It also removes each dependency of a removed tool that has a
`[tools.<name>]` table and that no other selected tool needs. An empty
`[tools]` table goes too. An unknown tool name is code 3 for `add tool` and
`remove tool`, and an unknown pack name is code 3 for `add pack`.

#### Named tools that a command refuses

`add tool` and `replace tool` refuse a tool that the command names, with code 2,
when the tool has `selectable = false` or when its `hosts` do not name the
host. `replace` examines only its new tool: the old tool can be one that a
values file names directly. The command examines only the names, not their
dependencies. See [tools.md](tools.md#the-catalog-keys).

- `selectable = false`: `<tool> is not selectable. Select <tool>, <tool> or
  <tool>.`
  The message names the tools of the side whose `depends` name it, sorted.
  Without such a tool, it is only `<tool> is not selectable.`
- `hosts`: `<tool> needs the host <hosts> - this host is <host>`.

`detect` does not propose such a tool. `sync` accepts a values file that names
a tool with `selectable = false`, and `sync` and `check` do not examine
`hosts`. `questions` and `list` show such a tool when the values file names
it.

### `replace`

`replace tool <old> <new>` works only between two tools with the same `parent`,
for example `pnpm` and `aube` with `parent = "node"`. The new tool must have the
script `scripts/migrate-<old>`. `replace` refuses a new tool that is already
selected, and an old tool that another tool needs. It refuses a new tool with
`selectable = false`, or whose `hosts` do not name the host. See
[Named tools that a command refuses](#named-tools-that-a-command-refuses).

The migrate script runs after the consent and before the writes. Its input has
`tool` = the new tool and `values` = the old `[tools.<old>]` table. The new
table gets only the `env` of the old table.

### `list`

`list` prints a tree. It writes nothing. Each root is a selected tool that no
selected tool needs: a tool of a `[tools.<name>]` table, or a required member
of `base`, also without a table. The tree has no pack roots, because the
values file keeps no record of a pack.

Below a root are its dependencies, sorted by name, and a dependency of a
dependency goes one level lower. The subtree of a tool shows in full at its
first copy only, in the order of the print. Each later copy of a tool that has
dependencies shows `<name> ↑` and no children. A required tool has the marker
`● required`, also on a later copy, and all markers are in one column. The
last line counts the tools. It adds ` · ↑ = the subtree shows above` only when
the tree has an arrow.

```text
bash
└── mise          ● required
    └── cli
claude
fish
└── mise ↑        ● required
vscode

· 6 tools · ↑ = the subtree shows above
```

### `lsp`

`lsp` prints the language servers of the selected tools: each `[lsp.<server>]`
table of a selected tool (see
[tools.md](tools.md#the-lsp-table)). It writes nothing. A server is listed only
when its tool is in `dev`, and the `requires_found` of its source pin, in
`[mise.tools]` or `[mise_env.dev.tools]`, is met: any one of its names is found
at the root or down to 4 folders below it. A pin without `requires_found`, and
the source `host`, are always met. An `[lsp]` table has no gate of its own.

Each line has the server, its tool, its command with its args, and its
extensions. The last line counts the servers. With `--json`, `data` is
`{"servers", "checks"}`:

- `servers` maps each server to an entry of the `lspServers` of Claude, so a
  plugin copies it verbatim: `command` is `mise`, `args` is
  `["x", "--", <command>, <args>…]`, `extensionToLanguage` is the
  `extensions` table, and `startupTimeout` is `startup_timeout`, left out when
  it is not set.
- `checks` maps each server to `{"command", "source"}`: the command of its
  `[lsp]` table and its source. The task `code:lsp:check` reads it.

`describe --json` gives the same object in its `lsp` field. Outside a
repository, or without `.config/bootstrap/`, `lsp` stops with code 2 and "run
bootstrap migrate".

### `detect`

`detect` runs the detect script of each tool, in parallel, in the root of the
repository. Each script prints a result. See
[The scripts of a tool](#the-scripts-of-a-tool). On the repo side, it then finds
the projects, and runs each script in each project folder too. See
[The projects of `detect`](#the-projects-of-detect).

The output is the tree of the selected tools, as in [`list`](#list), with the
same roots and the same order. It shows only a tool that has a detect script,
and the path from its root to it. The result of each script is in one column,
and the evidence and the reason follow:

| Result     | Column                                                                                                  |
| ---------- | ------------------------------------------------------------------------------------------------------- |
| `match`    | `✔ found`, then `<evidence>: <reason>`. A list of more than one evidence shows the first and `+N more`. |
| `none`     | `· nothing found` (dim) for a required tool or a dependency. The user cannot remove it.                 |
| `none`     | `○ nothing found` (yellow) for a tool with a `[tools.<name>]` table. The user can remove it.            |
| `mismatch` | `✗ mismatch`. A `mismatch` of a selected tool is an error.                                              |

A tool that matches and is not selected shows after the tree, in the group
`Found, not selected`. Its rows use the columns of the tree: the status
(`offered` or `not offered`) is in the result column, and the evidence is in
the evidence column.

- `· <tool>  not offered  <evidence> is the evidence of <tool>`: the evidence is
  only the evidence of a selected tool that matches. The tool is not offered.
- `✚ <tool>  offered  <evidence>: <reason>`: `detect --add` adds the tool.

A tool with `selectable = false`, or whose `hosts` do not name the host, does
not show in `Found, not selected`, also on the Mac. Thus `detect --add` never
stops for such a tool.

```text
android                ✔ found           android/settings.gradle: the Gradle files use the Android Gradle plugin
go                     ○ nothing found   no go.mod
linter
└── node               · nothing found   no package.json

Found, not selected
  · kotlin             not offered       android/settings.gradle is the evidence of android
  ✚ github             offered           .github/: .github/ exists
```

- Without a flag, `detect` only lists, also in a terminal, and writes nothing.
  Each offered tool prints `✗ <tool> is found, but it is not selected` with the
  fix `Run bootstrap add tool <tool>, or bootstrap detect --add`, and the
  command exits 1. When each detected tool is selected, it exits 0, also with
  projects or links to write: then it prints the info line
  `run bootstrap detect --add to write the projects`.
- `--add` adds each offered tool, the new projects and the new links through
  the change list, after the consent `? Write <n> changes? [y/N]`. `--yes`
  answers the consent. `--yes` and `--answer` need `--add`, else code 2 with
  the hint `add --add`. `--dry-run` works with or without `--add`.
- `--no` is gone: it is an unknown flag (code 2).
- When nothing is offered, the last line is `✔ No tool to add`.

#### The projects of `detect`

On the repo side, `detect` also proposes the `[projects.<name>]` tables of
`repo.toml` and the `projects` list of each tool in `tools.toml`. See
[bootstrap-config.md](bootstrap-config.md#projectsname). The code is in
`internal/projects` and `internal/core`.

It finds the project folders in this order:

1. The members of the workspace manifests at the root: `packages` of
   `pnpm-workspace.yaml` and `aube-workspace.yaml`, `use` of `go.work`,
   `[workspace].members` of `Cargo.toml` and `[tool.uv.workspace].members` of
   `pyproject.toml`. It expands the globs, and applies the `!` patterns of
   pnpm and aube and the `exclude` lists of Cargo and uv. A member counts only
   when it is a folder with the package file of its workspace (`package.json`,
   `go.mod`, `Cargo.toml` or `pyproject.toml`), as pnpm does. A member `.` is
   left out.
2. The folders below the root where the detect script of a tool of the
   category `languages` gives `match`, down to 4 folders below the root. A
   match counts only when one of its evidence paths is in that folder itself.
3. With no project found, the root is one project, with the folder `.` and the
   name `repo_name`.

- The walk skips `.git`, `.config`, `node_modules`, `vendor` and each folder
  that git ignores (`git ls-files --cached --others --exclude-standard`). A
  `.config` folder holds the configs of the tools, never a project. These
  rules apply to the members of step 1 too.
- A folder inside a found project is not a new project. A folder that
  `repo.toml` lists is never proposed again.
- The name of a project is the basename of its folder in lowercase kebab case:
  `My_App` is `my-app`. When two new folders have one name, or a listed project
  has it, the new name is the path with each `/` as `-`: `a/api` is `a-api`. A
  folder with no such name is skipped with a warning.

Then `detect` runs the detect script of each tool in each project folder, the
listed and the new ones. It links a tool to a project, and proposes the name
for the `projects` list of the tool, when:

- the script gives `match` in the folder of the project, with an evidence path
  in that folder itself. A project with the folder `.` gets this link only from
  a tool of the categories `languages` and `package-managers`.
- the tool is of the categories `languages` or `package-managers`, matches at
  the root and in no other project folder, and a workspace manifest of its
  evidence names the folder of the project. For example, `pnpm` links to the
  `packages` of `pnpm-workspace.yaml`.

Only a selected or an offered tool gets links, and a `root_only` tool never
does. A tool that is not selected and that matches only in a project folder is
offered too. A name that the list already has is not proposed again.

The new projects and links show after the tree, in the group
`Projects, not written`:

```text
Projects, not written
  ✚ web   project  apps/web
  ✚ api   project  services/api
  ✚ go    link     api
  ✚ node  link     web
  ✚ pnpm  link     web
```

`detect --add` writes them with the offered tools, in one change list.

#### The package scan of `mac detect`

`mac detect` runs no detect script. It lists the installed packages of the
Mac, in `$HOME`:

| Kind of package | Command                                                     | Key                                          |
| --------------- | ----------------------------------------------------------- | -------------------------------------------- |
| Formula         | `brew leaves --installed-on-request`, `brew list --formula` | `brew:<name>` of `[mise.bootstrap.packages]` |
| Cask            | `brew list --cask`                                          | `brew-cask:<name>`                           |
| App Store app   | `mas list`                                                  | `mas:<id>`                                   |
| mise tool       | `mise ls --global --installed --json`                       | `<name>` of `[mise.tools]`                   |

- Without `brew` on `PATH`, the scan skips the formulae and the casks, with the
  info line `brew is not installed, so Homebrew packages are not scanned`.
  Without `mas`, it skips the apps the same way. A command that fails stops the
  run.
- brew prints a formula or a cask of a tap without the tap. Thus the scan
  compares a `brew:` or `brew-cask:` key by the part after its last `/`:
  `brew-cask:stablyai/orca/orca` matches the cask `orca`.
- Each installed formula of `brew list --formula` can match a Mac tool. Only a
  formula of `brew leaves --installed-on-request` can be an offer of a package
  of no tool. Thus a dependency is never an offer. `brew leaves` also does not
  list a formula of a tap that Homebrew does not trust.
- Each key of `[mise.bootstrap.packages]` and `[mise.tools]` in a `tools.toml`
  belongs to that tool. A Mac tool with one or more installed keys gives
  `match`, with those keys as the evidence and the reason `installed`. A Mac
  tool with keys and none installed gives `none`, with the reason
  `no package installed`. The tree and the offers of tools are as above.
- An installed key of no Mac tool, and not in the own `[mise]` tables of
  `macos.toml`, is an offered row of the group `Found, not selected`, with the
  reason `no Mac tool has it: add it to <table> of macos.toml`.
- An accepted key goes into `[mise.bootstrap.packages]` or `[mise.tools]` of
  `macos.toml`, with the version `latest`. See [mac.md](mac.md) for the own
  config of the values file.
- An accepted `mas:` key also adds the tool that installs `brew:mas` or the
  mise tool `mas`, when it is not selected: the own config has no `depends`.
- Without `--add`, each offered key prints
  `✗ <key> is installed, but no Mac tool has it` with the fix
  `Add it to <table> of macos.toml, or run bootstrap mac detect --add`, and
  `mac detect` exits 1. `--add` adds each offered tool and each offered key, in
  one write of the values file and one sync.

### The layout of a values file

The engine changes only the keys that it must change in a values file. The
package `cli/internal/tomledit` edits the lines of the file: it replaces the
line, or the lines, of a key, adds a key at its sorted place in its group, adds
a table at the end of the file with its keys in sorted order, and removes a key
or a table. Each other line stays as it is: the comments, the blank lines and
the layout of the values. A new value has the layout of dprint, for example
`[ "a", "b" ]`. Thus the format hook of the repository does not change the file
again. These writes use it: `config set`, `config unset`, the `[home]` of a
member sync and the empty `[tools.<name>]` tables of a sync.

For a form that the editor does not edit, an array of tables, a dotted key or
an inline table on the path of the change, the engine uses `yq`, which keeps
the comments but writes the file in its own layout. The plans of `add`,
`remove` and `migrate` also still use `yq`.

### `config`

`bootstrap config get|set|unset|show` reads and writes the values of the
repository in `.config/bootstrap/`. It is the one reader of a repo value for a
script, a task and a plugin: none of them reads a file of the folder. There is
no `mac config`. The settings of the engine are `bootstrap settings`: see
[The config file](#the-config-file).

A key is a path whose first segment names the file: `repo.<key>`,
`tools.<tool>.<key>` or `releases.<project>`, for example `repo.repo_name` or
`tools.node.registry`. A segment with a dot is double-quoted, as in yq:
`repo.branches."release.1".landing`. The keys and their schema are in
[bootstrap-config.md](bootstrap-config.md).

| Subcommand                 | Job                                                                                                                                                                                                                                                                            |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `config get <key>`         | Prints the object of the key: `{"key": <key>, "value": <value>, "source": "file" \| "default" \| "contributed" \| "derived" \| "unselected"}`, with `"contributed_by": [<tool>, …]` when other tools contributed to the value. The value is a scalar, a list or a whole table. |
| `config set <key> <value>` | Writes the value, typed by the schema or by the catalog: a string as it is, an integer, and a list or a table as JSON, for example `'["main"]'`.                                                                                                                               |
| `config unset <key>`       | Removes the key, or a whole table such as `repo.members.<name>` or `repo.branches.<b>`.                                                                                                                                                                                        |
| `config show [<file>]`     | Prints `repo`, `tools` or `releases`, or the three files as one table by file name.                                                                                                                                                                                            |

- `get` of a `repo.*` or `releases.*` key exits 0 for each valid key, set or
  not. A key that is not set prints its schema default with
  `"source": "default"`, or `null` when the schema has none. It never needs the
  clone.
- `get` of a `tools.*` key prints the value that the engine uses, as the merge
  of a sync makes it (see [The values of a tool](#the-values-of-a-tool)): the
  defaults, the contributions of the selected tools, the derive scripts of the
  selected tools, the file and its overrides, in that order. `get` runs the
  derive scripts from the clone, with the input of a sync. `source` names the
  last layer of that order that has the path: `file` when the values file or
  its overrides have it, else `derived` when the derive script of the tool gave
  it, else `contributed` when a contribution changed the value, else
  `default`. The source is one for the value as a whole. When a contribution
  changed the value, `contributed_by` lists the contributing tools, sorted. A
  value that a contribution changes and the values file does not have is
  `contributed`, also when the tool has a default for it.
- An engine key of a tool that the values merge does not merge, `env` and
  `projects`, prints the value of the file (`file`), else the default of the
  engine (`default`): for `env` the `default_envs` of the tool, for `projects`
  `[]`. Thus `set` then `get` of `tools.<tool>.env` gives the value that `set`
  wrote. `get` of a whole tool table, `tools.<tool>`, also has its `env`, and
  its `projects` when the file has them, with the source of the table.
- `get` of a `tools.*` key of a tool that is not selected prints `null` with
  `"source": "unselected"`.
- `get` of a `tools.*` key needs the catalog: `--clone`, else
  `BOOTSTRAP_CLONE`, else the cache clone of the origin. It reads either one as
  it is, with no age. `get` never fetches: with no catalog it stops with code
  2, an error, and no value.
- `--format json|yaml|toml` chooses the output of `get` and `show`. The default
  is `json` when stdout is not a terminal, else `toml`. `get` prints the
  object in each format: one line of JSON, a YAML map or a TOML document.
  TOML has no null, so a TOML object without a `value` key means `null`; JSON
  and YAML print `null`.
- `--json` on `get`, `set`, `unset` and `show` prints the envelope of
  [The JSON contract](#the-json-contract). Its `data` is the object of `get`,
  the files of `show`, `{"key": <key>, "value": <value>}` for `set`, with the
  value that the file holds after it, and `{"key": <key>}` for `unset`. The
  info lines of `set` and `unset` go to stderr only. `--json` with `--format`
  is a usage error (code 2). `--format` without `--json` stays as it is, so a
  hook or a tool script that reads `config get` does not change.
- `set` and `unset` of a `tools.*` key read the catalog of the clone: the key is
  `env`, a key below `files`, `projects`, or a question or a default of the
  tool. A tool
  that is not selected stops `set` with code 2 and the fix "run bootstrap add
  tool <tool>".
- `set` and `unset` write only the file of the key, with its comments kept,
  and run no sync. A change of a key that a sync reads prints the info line
  `run bootstrap sync to apply it`: each `tools.*` key, and each `repo.*` key
  except `topology`, `topology_reason`, `bootstrap_commit` and `format`. A
  change whose result breaks a rule of the
  schema writes nothing and stops with code 2. `unset` of a required key stops
  with code 2.
- `set repo.projects.<name> '{"folder": "<folder>"}'` writes a project. Each
  name of `tools.<tool>.projects` must be a `[projects.<name>]` of `repo.toml`,
  and a `root_only` tool has no `projects`, else code 2. Thus a `set` or an
  `unset` in either file that breaks this rule writes nothing.
  `unset repo.projects.<name>` stops with code 2 while a tool lists the name,
  and the error names those tools.
- `set releases.<project> <level>` keeps the higher level when the file already
  has one, and creates `releases.toml` with `format = 3` when it is not there.
  `show` of a missing `releases.toml` prints an empty table.
- With `BOOTSTRAP_VALUES` set, `config` reads and writes that folder and skips
  the stop of the old form. Else, without `.config/bootstrap/`, or without its
  `repo.toml` or `tools.toml`, each subcommand stops with code 2 and the fix
  "run bootstrap migrate".

| Code | `config`                                                                                                                                               |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 0    | A valid key, set or not (`get`), and a write that is done.                                                                                             |
| 1    | A file of the folder that cannot be read or parsed.                                                                                                    |
| 2    | A key that is not valid, a value that breaks the schema, no `.config/bootstrap/`, a `tools.*` `get` with no catalog, and a derive script that fails.   |
| 5    | `set` or `unset` of a `tools.*` key when the fetch fails and the cached clone cannot be used. See [The cache and the clone](#the-cache-and-the-clone). |

## The JSON contract

A plugin or an AI agent reads and drives the engine only through one JSON
envelope. `bootstrap contract`, `catalog`, `describe`, `lsp`, `check`,
`questions`, `mac questions`, `config get|set|unset|show`, `migrate`, `sync`,
`add`, `remove`,
`replace` and `detect` take `--json`. The code is in `internal/contract`.

**The envelope.** With `--json`, the command prints one JSON object on stdout,
on one line, then a newline:

```text
{"contract":1,"command":"config set","ok":true,"code":0,"data":{"key":"repo.repo_name","value":"orders-api"},"errors":[],"warnings":[]}
```

| Key        | Value                                                                                                                 |
| ---------- | --------------------------------------------------------------------------------------------------------------------- |
| `contract` | The version of the contract, `1`.                                                                                     |
| `command`  | The full command words, joined by one space: `check`, `config get`, `mac questions`. `""` when no command was parsed. |
| `ok`       | `true` only when `code` is `0`.                                                                                       |
| `code`     | The exit code of the command. The process exits with this code.                                                       |
| `data`     | The result of the command, below. `null` when the command fails before it has a result.                               |
| `errors`   | The errors, each `{"code", "message", "path"}`.                                                                       |
| `warnings` | The warnings, each `{"code", "message", "path"}`, with `code` `0`.                                                    |

`message` is the text that a person sees, with the fix after `-`. `path` is
the file when one is known, and is left out otherwise. A list is always `[]`,
never `null`.

- Each human line goes to stderr: the clone line, the change list, the trees,
  the output of the sync, the info lines and the errors. stdout holds the
  envelope only.
- The engine looks for a literal `--json` anywhere in the arguments before the
  parse. With one, a usage error also prints the envelope, with code 2, and the
  help goes to stderr. `--json` on a command that does not take it, for
  example `list` or `tui`, is such a usage error.
- On the Mac side, only `mac questions` takes `--json`. Each other `mac`
  command with `--json` stops with code 2. `contract`, `catalog`, `describe`
  and `lsp` have no `mac` form.

**The version.** `contract` is one integer. A new field, command or enum value
keeps it. A removed or renamed field, or a changed type or meaning, raises it.
`bootstrap contract` lists the versions that the engine supports.

**The data of each command.**

| Command                                                       | `data`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| ------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `contract`                                                    | `{"versions": [1], "cli_version", "commands", "ports"}`. `cli_version` is the version of the binary, `dev` for a local build. `commands` are the full words of each command that takes `--json`. `ports` are the names of the task ports. It needs no repository, no config file and no clone.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `catalog`                                                     | `{"fetched_at", "categories", "tools", "packs"}`. `fetched_at` is the time of the last successful fetch in ISO-8601 UTC, `""` under `--clone` or `BOOTSTRAP_CLONE`. Each tool has `name`, `category`, `description`, `questions` and `defaults`. It reads the repo side of the clone, and needs no repository.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `describe`                                                    | `{"repo", "tools", "members", "ports", "lsp"}`: the values of `repo.toml`, each selected tool with `name`, its `category` in the catalog and its `values` of `tools.toml` (with its `projects` list, without `env`), each `[members.<name>]` with `name`, `type`, `folder`, `url` and `present` (the folder exists), each port with `name` and `present`, and `lsp`, the `data` of `lsp`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `lsp`                                                         | `{"servers", "checks"}`: each listed server as an entry of the `lspServers` of Claude, and its command and source. See [`lsp`](#lsp).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `check`                                                       | `{"valid", "checks"}`. `checks` has one entry for each stage, in order: `{"name", "status", "findings"}`, with `status` `pass`, `fail` or `skipped`, and each finding `{"code", "severity", "message", "path"}`, with `severity` `error` or `warning`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `questions`, `mac questions`                                  | The list of the questions. See [The questions](#the-questions).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `config get`, `config show`                                   | The object of `get`, and the files of `show`. See [`config`](#config).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `config set`, `config unset`                                  | `{"key", "value"}` for `set`, and `{"key"}` for `unset`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `migrate`, `sync`, `add`, `remove`, `replace`, `detect --add` | `{"plan", "outcomes"}`. `plan` is `{"changes", "files", "skipped", "pending", "notes"}`: the writes before the sync (`path`, `content`, `remove`, `note`; the move of `migrate` is one change with the absolute `path` `<root>/.config/mise/tasks`, an empty `content`, and its files and targets only in `note`; the removal of the known old copies of `migrate` is one change for each tool with the same `path`, an empty `content`, and its files only in `note`; the rename of `migrate` is one change with the `path` `<root>/.config`, an empty `content`, and its files and keys only in `note`), the files of the sync (`path`, `kind`), the members that `add pack` leaves out (`tool`, `why`), the questions without an answer, and the notes. `outcomes` has one `{"kind", "path", "message"}` for each target of the sync, `[]` for a dry run. The warnings of the plan are the warnings of the envelope. |
| `detect`                                                      | `{"detected", "projects", "links"}`: one entry in `detected` for each tool that `detect` reports, with `tool`, `selected`, `result`, `evidence`, `message`, `offered`, `shared` and `of`. `projects` maps the name of each new project to `{"folder"}`, and `links` maps a tool to the project names to add to its `projects` list. Both are `{}` on the Mac side.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |

**The stages of `check`.** `envs`, `values`, `old-forms`, `env-rules`,
`core-tools`, `overrides`, `members`, `releases`, `projects`, `scripts` and
`project-links`, 11 in all. A stage whose input failed is `skipped`.
`projects` holds an error (code 2) for each `[projects.<name>]` whose folder
is not a folder of the repository. `project-links` runs the detect script of
each tool in the folder of each project of its `projects` list. A link passes
when the script gives `match` there, with an evidence path in that folder
itself, or, for a tool of the categories `languages` and `package-managers`,
when the tool still matches at the root and a workspace manifest of its
evidence still names the folder. A link that fails both tests is a warning. A
script that fails or prints no valid result is an error (code 2). A
`root_only` tool and a project whose folder is missing are skipped. `scripts`
also holds an error (code 2) for each known old copy of a part of a selected
tool at `<group>/50-local` (see [`migrate`](#migrate)). A known copy at the old
path
of the group is the error of `old-forms` for a task at the place of a test
port. `old-forms`
also holds an error for each task of
the repository at the place of a test port (see [`migrate`](#migrate)), an error
for `mise run setup:worktree` in `.config/wt.toml` or in a
`[members.<name>].worktree_install`, and a warning for each executable file
without an `NN-` name in the folder of a dispatcher: `code/format`,
`code/lint`, `code/sec`, `setup/deps`, `setup/secrets`, `test`, `test/e2e` and
`test/golden`. A file whose task name a `# MISE depends` (or `#MISE depends`)
of another file in the same folder names, and a name that starts with `_`,
get no warning. The task name of a file has no `.py`, as in mise: the depends
line `test:go` names the file `test/go.py`. `core-tools`
always passes: its line is a warning finding, and it never changes the code.
`valid` is `true` only when the code is `0`, and the code is the highest code of
the stages. A failure before the first stage gives
`{"valid": false, "checks": []}`, with the cause in `errors`.

**The write commands.** With `--json`, a command never prompts.

- `migrate`, `add`, `remove`, `replace` and `detect --add` need `--yes` or
  `--dry-run`, else they stop with code 2.
- `sync --json` applies with no `--yes`, as `sync` does. It always runs the
  quiet dry run of the plan first, reports its files in `data.plan.files`, and
  applies the clean-target rule.
- A plan with questions without an answer: with `--yes`, the command writes
  nothing and stops with code 2, and one error names the questions. With
  `--dry-run`, it exits 0, and one warning names them. `data.plan.pending`
  lists them, and a plugin gives the answers with `config set` or `--answer`.
- A plain `detect --json` only reads, and needs no `--yes` or `--dry-run`.

**The task ports.** A port is the name of a task that a plugin runs:
`code:format`, `code:lint`, `code:sec`, `code:precommit`,
`git:merge:<branch>`, `git:worktree:create`, `git:worktree:setup`,
`git:worktree:delete`, `code:lsp:check`, `setup`, `test`, `test:e2e`,
`test:golden` and `p:<id>:deploy`. `contract` lists the names only. `describe`
runs `mise tasks ls --json --no-global --hidden` in the repository, so a global
task is left out and a hidden task is in, and gives each port `present`. A
patterned port, such as `git:merge:<branch>`, also has `tasks`, the names of
the tasks that match it: `<branch>` and `<id>` match one segment without a
colon, and `present` is `true` when one task matches. When mise cannot list the
tasks, for example for a config that it does not trust, `describe` stops with
code 1 and the fix names `mise trust`. Outside a repository, or without
`.config/bootstrap/`, `describe` and `lsp` stop with code 2 and "run bootstrap
migrate".

## The cache and the clone

The cache is `${XDG_CACHE_HOME:-~/.cache}/bootstrap/`. `--cache-dir`,
`BOOTSTRAP_CACHE` or the key `cache_dir` of the config file changes the folder.
All the consumers of the machine and the Mac side share it.

| Entry                 | Use                                                  |
| --------------------- | ---------------------------------------------------- |
| `repo/`               | One shallow clone of this repository, branch `main`. |
| `lock/`               | The lock folder, with the `pid` of its holder.       |
| `lock.takeover/`      | The short lock of the removal of a dead lock.        |
| `fetched`             | The time of the last fetch of `repo/`.               |
| `macos-generated.tsv` | The manifest of the Mac side. See [mac.md](mac.md).  |

Each command except `version`, `settings`, `contract` and `config` takes the
clone in `internal/core`. `config show` and `config get` of a `repo.*` or
`releases.*` key never need the clone. `config get` of a `tools.*` key reads
`--clone`, `BOOTSTRAP_CLONE` or the cached clone as it is, and never fetches:
with no clone it stops with code 2.
After each successful fetch, the engine writes its time to the file `fetched`
in the cache root. The catalog reads (`config set` and `config unset` of a
`tools.*` key, `catalog` and `describe`) follow one fetch rule:

1. A last fetch of 30 minutes or less: the command uses the cached clone, with
   no fetch.
2. An older last fetch, or no recorded time: the command fetches.
3. A fetch that fails: the command uses the cached clone when its last fetch
   is less than 7 days old, with a warning. Else it stops with code 5, also
   for a cached clone with no recorded time.

Under `--clone` or `BOOTSTRAP_CLONE` the clone is used as it is, with no fetch
and no age rule. Each other command fetches each time.
With `--clone`, `BOOTSTRAP_CLONE` or the key `clone` of the config file, the
command uses that folder as it is: it takes no lock and makes no fetch. Else it
takes the lock, fetches, and copies the clone to a temporary folder. The sync
reads only this copy. Thus a change of the clone applies to a whole run, and the
next run can update the cache.

The URL is `https://github.com/virajp/bootstrap.git`. `--url`, `BOOTSTRAP_URL`
or the key `url` of the config file changes it. The fetch does these steps:

- It clones when `repo/` is missing or is an empty folder.
- It fetches the branch `main` when `repo/` is a clone whose `origin` matches
  the URL. It fetches from the URL, not from `origin`, because the compare
  drops the scheme, and the transport must not come from the cached `origin`.
  Then it runs `git reset --hard FETCH_HEAD` and `git clean -ffdx`. Do not keep
  work in the clone.
- It clones again when `repo/` is a clone with another `origin`, for example
  after a test with `BOOTSTRAP_URL`.
- It stops with code 2 when `repo/` is not a git clone and is not empty. It
  does not delete the folder.

The origin compare is normalized. Two URLs match when they are equal without the
scheme, the user name, the `user@host:path` form, a trailing `/` and a trailing
`.git`. Thus the SSH and the HTTPS URL of one repository match. A git call
that uses the network has no terminal prompt and uses SSH `BatchMode`. Thus a
missing key fails and does not wait for a password. A failed clone or fetch is
code 5. For a catalog read, a failed fetch is code 5 only when the cached clone
cannot be used: see the fetch rule above.

The sync uses the pushed `main` branch. Local changes have no effect on
consumers until you push them. To test a change before you push it, set
`BOOTSTRAP_URL` to a local clone with a commit of the change on `main` of that
clone, or set `BOOTSTRAP_CLONE` to a folder with the change. Use
`bootstrap sync --dry-run` to show a diff.

### The old clone at the cache root

A git repository at the cache root is the old layout. When its raw `origin`
normalizes to the URL, the run removes each entry of the cache root, except
`repo/`, `lock/`, `lock.takeover/`, `macos-generated.tsv` and `fetched`. Another
`origin`
or no `origin` stops the run with code 2, and the run deletes nothing in the
cache root.

### The lock

`lock/` makes a second run wait. The run creates the folder and writes its
`pid`. A waiter removes the lock of a dead holder, and a lock without a `pid`
that is older than one minute. The removal has its own lock, `lock.takeover/`:
without it, a second waiter can remove the lock that the first waiter took
after the removal. A signal removes the lock and the copy.

| Command                                                                                    | Release                      |
| ------------------------------------------------------------------------------------------ | ---------------------------- |
| Each repo command, `mac questions`, `mac check` and `mac sync --dry-run`                   | After the copy of the clone. |
| `mac sync`, `mac migrate`, `mac tui`, `mac add`, `mac remove`, `mac replace`, `mac detect` | At the end of the command.   |

## The version check

`cli/VERSION` of the clone is the minimum version of the engine that
the clone needs. After it takes the clone, each command compares its own
version with that value.

- A `VERSION` that is not `MAJOR.MINOR.PATCH` stops the command with code 5.
- An older binary stops the command with code 2, and tells the user to run
  `mise use -g github:virajp/bootstrap`.
- A build without a version, for example `dev`, is new enough.

mise installs the binary with `github:virajp/bootstrap`. The `mise` tool of each
side has it as a mise tool.
See [repo.md](repo.md) and [mac.md](mac.md).

## The engine variables

| Variable            | Use                                                                                                                                                                                                                                                                             |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BOOTSTRAP_CLONE`   | A folder with a clone of bootstrap. The command uses it, with no lock and no fetch. The gate `code:lint:60-bootstrap` sets it.                                                                                                                                                  |
| `BOOTSTRAP_URL`     | The clone URL. A test uses it for a change that is not pushed.                                                                                                                                                                                                                  |
| `BOOTSTRAP_CACHE`   | The cache folder. A test uses it.                                                                                                                                                                                                                                               |
| `BOOTSTRAP_VERBOSE` | `1` prints the log line of each target in place of the list. See [The output of a run](#the-output-of-a-run).                                                                                                                                                                   |
| `BOOTSTRAP_BIN`     | The path of a build of the engine. The entry tasks, the gate, the tool scripts and the tests run it in place of the `bootstrap` of mise. The engine exports its own path in it to each script and task that it runs.                                                            |
| `BOOTSTRAP_JOBS`    | The jobs at a time. `--jobs` sets it.                                                                                                                                                                                                                                           |
| `BOOTSTRAP_VALUES`  | The values folder. The default is `.config/bootstrap/` of the repository, and on the Mac the fixed folder of `macos.toml`. `migrate` and the sync set it to the folder of the planned values, so `config` reads the new keys; with it, `config` skips the stop of the old form. |
| `BOOTSTRAP_HOME`    | The home repository of a member. `setup:configs` sets it for the sync of a member, which writes `[home]`.                                                                                                                                                                       |
| `BOOTSTRAP_MEMBER`  | The name of the member in the `[members.*]` tables of its home. `setup:configs` sets it with `BOOTSTRAP_HOME`.                                                                                                                                                                  |
| `BOOTSTRAP_CONFIG`  | The config file. See [The config file](#the-config-file). The tests set it to a missing path.                                                                                                                                                                                   |

A flag sets its variable for the run: `--jobs`, `--cache-dir`, `--url`,
`--clone`, `--config` and `--verbose` (`BOOTSTRAP_VERBOSE=1`).
The config file sets `BOOTSTRAP_JOBS`, `BOOTSTRAP_CACHE`, `BOOTSTRAP_URL` and
`BOOTSTRAP_CLONE` when no flag and no variable sets them. See
[The config file](#the-config-file).

## The config file

The user config file holds the defaults of the settings of the engine. Thus a
user sets them one time, and does not export a variable in each shell. The path
of the file is the first of these that is set:

1. The flag `--config <path>`. It also sets `BOOTSTRAP_CONFIG`, so a child
   `bootstrap` reads the same file.
2. The variable `BOOTSTRAP_CONFIG`.
3. `$XDG_CONFIG_HOME/bootstrap/config.toml`.
4. `~/.config/bootstrap/config.toml`, when `XDG_CONFIG_HOME` is not set.

The engine reads only that path. When the file at that path is missing, the
engine does not try the next path: it uses the defaults. An empty
`BOOTSTRAP_CONFIG` or `XDG_CONFIG_HOME` is not set. The file has these keys:

| Key         | Type                | Variable          | Flag          |
| ----------- | ------------------- | ----------------- | ------------- |
| `jobs`      | An integer above 0. | `BOOTSTRAP_JOBS`  | `--jobs`      |
| `cache_dir` | A path.             | `BOOTSTRAP_CACHE` | `--cache-dir` |
| `url`       | A string.           | `BOOTSTRAP_URL`   | `--url`       |
| `clone`     | A path.             | `BOOTSTRAP_CLONE` | `--clone`     |

For example:

```toml
cache_dir = "~/.cache/bootstrap"
jobs = 4
```

Each value comes from the flag, then the variable, then the file, then the
default. An empty variable is not set, so the value of the file applies. A
variable with a value is set, also when the value is bad, for example
`BOOTSTRAP_JOBS=abc`: it blocks the value of the file. The engine reads the file
for each command except `version` and help, before the preflight. For each key
without a flag and without a variable, it sets the variable of the process from
the file. Thus the scripts of a tool and mise get the value of the file as a
variable, the same as an exported variable.

The rules of the file:

- A missing file gives the defaults, with no message.
- A path of `cache_dir` or `clone` is absolute or starts with `~/`. The engine
  expands a leading `~/` with `HOME`. The engine uses the value of a flag or of
  a variable as it is.
- The file is for the repo side and for the Mac side. A `[mac]` table is for the
  keys of the Mac side only. The table has no key now, so a key in it is an
  unknown key. An empty `[mac]` table is valid. The place of `macos.toml` is
  fixed, so
  the file has no key for it: see [mac.md](mac.md#the-place-of-macostoml).
- `verbose` is a flag only. A `verbose` key is an unknown key.
- The file holds only the settings of the engine, never the config of a tool.
  No sync, no seed and no tool writes the file. Only the user writes it.

These errors stop the command with code 2. The message names the file, and the
key when it has one, for example
`✗ <file>: jobs: needs a number above 0, not 0`:

- A file that is not valid TOML.
- An unknown key or table, for example `mac.verbose`.
- A wrong type, for example `jobs = "4"`.
- An empty string, a `jobs` below 1, and a path of `cache_dir` or `clone` that
  is not absolute and does not start with `~/`.
- A path that is a folder or that the engine cannot read.

`bootstrap settings` prints the file and each key, and changes nothing. It needs
no git repository, no mise and no clone. The first line is `file`, the path of
the file and its source (`flag --config`, `variable BOOTSTRAP_CONFIG` or
`default`), with `(missing)` after the path of a missing file. Then each key has
one line with its value and its source: `flag --<name>`, `variable <NAME>`,
`file` or `default`. `-` is no value. For example:

```text
file           /Users/me/.config/bootstrap/config.toml  default
jobs           4                                        flag --jobs
cache_dir      /Users/me/.cache/bootstrap               default
url            https://github.com/virajp/bootstrap.git  default
clone          -                                        default
```

Only the engine reads the file. The gate `code:lint:60-bootstrap` and the
cleanup of the old files of `updateall:configs` find the cache only from
`BOOTSTRAP_CACHE` or the default. With `cache_dir` only in the file, the gate
skips with no message, and the cleanup uses the default cache.

## The tool calls

The engine does not call gomplate, jq or yq as commands. It has its own
template code (`internal/tmpl`), and it parses and merges JSON, YAML and TOML
in its own process (`internal/format`). Thus the sync operates in a shell
without `MISE_ENV=dev`, before a first sync, on a new machine and in a scratch
folder.

The engine runs these commands:

| Side    | Command                                                                                                                                          |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Both    | `git` of `PATH`, for the clone, the commit, the work tree status and the branches. `diff -u` for the diff of a dry run.                          |
| Both    | `mise x -- env -0` once for each folder of the scripts of a tool, then `bash <script>`. See [The scripts of a tool](#the-scripts-of-a-tool).     |
| `repo`  | `mise x -- which <formatter>` and `mise x -- bash -c <formatter command>`. See [The formatter](#the-formatter).                                  |
| `repo`  | `describe`: `mise tasks ls --json --no-global --hidden` of the `mise` of `PATH`, in the repository. See [The JSON contract](#the-json-contract). |
| `macos` | The first setup: `mise bootstrap --yes` of the `mise` of `PATH`, in `$HOME`.                                                                     |
| `macos` | `mac detect`: `brew leaves`, `brew list`, `mas list` and `mise ls` of `PATH`, in `$HOME`. See [`detect`](#detect).                               |

## The scripts of a tool

A tool can have scripts in `scripts/` of its tool folder. The engine runs them
from the clone, and never copies them.

| Script               | When                                                         | Output                                                                                                                                                              |
| -------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `derive`             | Each sync and each plan, before the render.                  | One JSON object: the derived values of the tool. It can have a `mise` and a `mise_env` table, which merge into the mise config of the tool. See [mise.md](mise.md). |
| `check`              | `sync` and `check`, before a write.                          | Exit 0. Or each line of stdout is a message, and the exit code is 2.                                                                                                |
| `after`              | Mac only: at the end of a sync, and of a first setup.        | Nothing, or `{ "restart": "<reason>" }`, which asks for a new terminal (exit 4). See [mac.md](mac.md).                                                              |
| `detect`             | Repo side, `detect`, and `check` for the projects of a tool. | `{ "result": "match" \| "none" \| "mismatch", "evidence": [...], "message": "..." }`.                                                                               |
| `migrate-<old tool>` | `replace tool <old tool> <this tool>` only.                  | Exit 0. It moves the repository from the old tool to this tool.                                                                                                     |

- The engine runs each script as `bash
  <clone>/<side>/tools/<tool>/scripts/<name>`,
  in the root of the repository, or in `$HOME` on the Mac. The script gets the
  environment of `mise x` in its folder: the engine runs `mise x -- env -0`
  once for each folder of the scripts of one kind, and runs the bash of that
  `PATH`. Thus mise installs the missing tools, as for `mise x -- bash`, but
  resolves each folder once, not once for each script. When
  `mise x -- env -0` fails, each script of that folder runs as
  `mise x -- bash <script>`, so the error of mise goes into the messages.
  `detect` also runs
  a detect script in each project folder, and `check` in the folder of each
  project of the `projects` list of the tool. See
  [The projects of `detect`](#the-projects-of-detect).
- Its stdin is one JSON object: `tool`, `values`, and on the repo side `repo`
  (`root`, `name` and `members`, each member with `name`, `folder` and `type`:
  `submodule`, `folder` or `sibling`),
  or on the Mac side `home`.
- `values` of a `derive` script is the values of levels 1 and 2. `values` of a
  `check` and an `after` script is the merged values of the tool. `values` of a
  `detect` script is the `[tools.<tool>]` table of `tools.toml`.
- The scripts of one kind run in parallel, at most `--jobs` at a time. The
  results come back in tool order, so the writes and the report stay in a fixed
  order.
- A script that fails, prints JSON that is not one object, or exits with a code
  that its kind does not allow stops the run with code 2. Its stderr lines go
  into the messages.
- A script parses JSON with `jq` of `PATH`, else with `mise x jq@latest`.
- A script cannot read the output of a script of another tool.
- Evidence of a detect script is config files, never lock files. The script
  does the precise detection. The Mac side runs no detect script: `mac detect`
  scans the installed packages. See [`detect`](#detect).

Any other file in `scripts/` stops the sync. See [tools.md](tools.md).

## Templates

- The entry of a file makes it a template: `mode = "template"`, a seed with
  `render = true`, or `mode = "merge"` (the merge spec). See
  [tools.md](tools.md#the-files-table). Any other file is copied byte for byte.
  The `.tmpl` suffix of a source is a name convention only.
- The output file takes the mode of its source, or `perm` of its entry. Make a
  template that renders a task or a script executable (`755`).
- The delimiters are `<% %>`. Do not use `{{ }}` or `[[ ]]` for template logic.
  mise uses `{{ }}` and bash uses `[[ ]]`, and both must reach the output
  unchanged.
- `.config` is the values of the tool of the template. `.all` is the values of
  each selected tool, by tool name. A tool without values has an empty table.
  See [The values of a tool](#the-values-of-a-tool).
- A top-level key `all` or `config` of the repo values does not reach a
  template.

The other values of the context are different on each side. See
[repo.md](repo.md#the-template-context) and
[mac.md](mac.md#the-template-context).

## The write of one target

| Side    | Order                                                                              |
| ------- | ---------------------------------------------------------------------------------- |
| `repo`  | Render or copy, parse check, merge, formatter, compare, write.                     |
| `macos` | Render or copy, parse check, merge, compare, write. The Mac side has no formatter. |

- A destination is replaced only after a complete render or copy. The temp file
  is next to the destination, so the final `mv` is atomic. A dry run puts the
  temp file in `${TMPDIR:-/tmp}`.
- A failure leaves the existing file as it was, and the other files continue.
- A file takes the mode of its source, or `perm` of its entry, also when it
  exists already. An existing shared file keeps its own mode, and a new one
  takes the mode of its source. When only the mode differs, the sync runs
  `chmod`.
- `dir_perm` of an entry sets the mode of the folder of the target, once in
  each run, on a new folder and on an existing folder. A dry run reports the
  change only for an existing folder with another mode. A failed `chmod` fails
  that target and each later target in the folder.
- An entry whose `when` value is false is not shipped. See
  [tools.md](tools.md#the-files-table).
- A dry run lists the target as `would change`, `would add` or `would seed`,
  with a unified diff when the file exists. A change of the mode alone lists as
  `would change`, with no diff. See [The output of a run](#the-output-of-a-run).
- A file with the bootstrap header at a path where a target needs a folder is
  removed. A file without the header, or a seed or a shared file of a tool of
  the run on the repo side, fails that target.
- On each side, the generated mise files are written first: a later file can
  need the tools of the new mise config.

### Owned files, shared files and seeds

The modes of a file are in [tools.md](tools.md#the-files-table).

- An owned file (`copy` or `template`) is written on each run, when its content
  differs. It has a manifest row, and the stale rule removes it when its tool
  goes.
- A shared file (`merge`) is merged item by item on each run. A change of the
  user stays. See [Shared files](#shared-files).
- A seed is written only when the target is missing. After that, it belongs to
  the user. An existing seed gets no parse check, because the sync does not
  write it. A seed has no manifest row and is never in the stale rule. When its
  tool goes, the seed stays, and `remove` and `replace` list it.

### The parse check

Each JSON, YAML or TOML target must parse, after its render or copy and before
its merge and the formatter. The engine reads JSON as jq does, and YAML and
TOML with the yq code in its own process. The `format` key of its `[files]`
table selects the format, else the extension of the target.

- A `.json` target without a `format` key, with a line that starts with `//`
  after spaces, is JSONC. The sync does not examine it. This rule does not
  apply to a shared file.
- An empty JSON target does not parse.
- A target that does not parse keeps its existing file, the other files
  continue, and the run does not set `bootstrap_commit`.
- A dry run examines each target too.

### The formatter

A repo tool can have a `[formatter]` table: the command and the config file
names of a formatter. See [tools.md](tools.md#the-formatter-table). The `dprint`
tool has one. At most one selected tool can have a `[formatter]`; two such
tools stop the sync with code 2.

The sync copies each rendered template, each generated mise file and each
shared file to a stage folder, at its path in the repository, with a copy of
each config file of the formatter that the repository has. Then it formats the
stage in one call: `mise x -- bash -c "cd <stage> && <command>"`, with the
paths of the files on stdin, one for each line. A plain file needs no step,
because its source in bootstrap is already formatted.

Without the command of the formatter (`mise x -- which`), or if the formatter
fails for a file, the raw render of that file stays. Without this step, the
format hook changes the render, and the next sync writes it back.

## Shared files

A shared file is a target that bootstrap and the user both change, for example
`.config/pre-commit-config.yaml`, `dprint.jsonc`, `.gitignore` or
`.vscode/extensions.json`, and app settings on the Mac. Its entry has
`mode = "merge"`, and its source is a merge spec `files/<name>.merge.toml`.
The code is in `internal/format`.

On each run, the sync does these steps:

1. It renders the spec with the values. A spec is always a template. The
   result is a list of `[[item]]` tables. After them, it adds the items that the
   selected tools contribute to the target in
   `[[contributes_items.<owner>."<target>".item]]`, in the order of the tool
   names, each in the item format of the spec, without an item whose identity
   is already there. See
   [tools.md](tools.md#the-defaults-questions-and-contributes-tables).
2. It reads the live file. An absent file starts as an empty document.
3. For each item, it finds the item in the live file. A found item is kept or
   written again, as its kind says. An item that is not found is added, unless
   an override says `"$skip" = true`. An item with `"$mode" = "remove"` is
   removed if found. The override table applies to the contributed items too.
   An override of an item that any tool of the clone contributes to the target
   is known, also when that tool is not selected.
4. It removes each item that the manifest records as added by bootstrap, and
   that no rendered spec and no contributed item has now (a tool went, or
   bootstrap dropped the item).
   The identity of the item finds it, not its value.
5. It formats, compares and writes the file with the comments of the live
   file. It records the added items in the manifest.

### The spec

The top-level keys of a spec are `format`, `comment` and `[[item]]`. Any other
key stops that target.

- `format` is `text`, `json`, `jsonc`, `yaml` or `toml`. The default is the
  `format` of the entry, else the extension of the target.
- `comment` is the comment marker of a text file, for example `#`. The marker
  `<!--` writes each comment line as `<!-- <text> -->`, for a Markdown file.

| Key of an item       | Meaning                                                                                                                                                                         |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`               | `key`, `keys`, `item`, `block` or `section`. A text file has only `item` and `section`.                                                                                         |
| `path`               | The place in the file: a list of keys. An item of a text file has no path.                                                                                                      |
| `value`              | The bootstrap value.                                                                                                                                                            |
| `match`              | A regex for an item of a list of scalars or for a line. A non-empty list of keys for a block in a `table[]`; a key path through a list (`hooks.id`) compares the set of values. |
| `id`                 | The name of a block for an override. The live file does not have it. A block with `match` needs an `id`.                                                                        |
| `ignore`             | The keys of a found block that the engine never writes, for example `rev`.                                                                                                      |
| `type`               | Optional. Necessary for an empty list. One of `string`, `boolean`, `integer`, `float`, `<scalar>[]`, `table`, `table[]`.                                                        |
| `"$mode"`, `"$skip"` | `"$mode" = "remove"` removes the item. `"$skip" = true` does not add it.                                                                                                        |

| Kind      | What it is                                                                                                   | Found                                                                                |
| --------- | ------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `key`     | One leaf key. A list value is one value.                                                                     | The live value stays. A live value of another type is written again, with a warning. |
| `keys`    | A table of leaf keys: one `key` item for each leaf. The user can add keys to the table.                      | As `key`, for each leaf.                                                             |
| `item`    | One item of a list of scalars, or one line of a text file. Found by an equal `value`, or by a `match` regex. | The live item stays. A new item or line goes at the end, in spec order.              |
| `block`   | A value that bootstrap owns: a table, a list, or one table of a `table[]`, found by its `match` keys.        | Written again from the spec, except the `ignore` keys.                               |
| `section` | The part of a text file that bootstrap owns: a string of lines between two marker lines at the top.          | Written again from the spec. Each line outside the markers stays.                    |

A text file has at most one `section`, and its spec needs the `comment` key.
The marker lines are comments that start with `bootstrap:start` and
`bootstrap:end`. A new section goes at the top of the file, after the shared
header, with an empty line before the lines of the user. A change of the user
between the markers is lost at the next sync; the user adds own text after
`bootstrap:end`. A start line without an end line gets a warning, and the
section replaces that start line. A section has no override: `"$mode" =
"remove"` in the spec removes it, with the empty line after it.

A spec renders its items from the values, for example from the contributions of
the other tools: `[[contributes.pre-commit.repos]]` gives one block for each
hook repository, and `[contributes.vscode] extensions` gives one item for each
extension. The template function `data.ToTOMLInline` writes a value as inline
TOML.

### Overrides of a shared file

The user adds own items directly in the file: own lines, own excludes, own hook
blocks, own extensions. bootstrap never touches them. An override is only for a
bootstrap item, in the values of its tool:

```toml
[tools.dprint.files."dprint.jsonc".markdown]
emphasisKind = { "$skip" = true } # bootstrap does not add it
strongKind = "underscore" # own value

[tools.git.files.".gitignore"]
"*.log" = { "$mode" = "remove" }

[tools.pre-commit.files.".config/pre-commit-config.yaml".repos]
"conventional-commits" = { "$mode" = "remove" }
```

- The path of an override is the `path` of the item, then the value of an
  `item`, or the `id` of a `block`.
- An override has an own value, or `"$skip" = true`, or `"$mode" = "remove"`.
  Any other `$` key stops that target.
- An override with no bootstrap item at its path stops that target: remove it.

### The output of a shared file

- A JSON file is written in the form of jq: a 2-space indent, the key order of
  the live file, new keys at the end of their table, and one trailing newline.
- A JSONC file keeps its comments and its trailing commas.
- A YAML or TOML file is written with the yq code, so the comments of the live
  file stay. In TOML, a comment on a table-header line can be lost.
- A text file keeps each live line. New lines go at the end. The section goes
  at the top.
- A file in a format with comments gets the shared header. See
  [The header](#the-header).

A live file that does not parse, or whose top level is not a table, keeps its
content and fails that target. A shared file has no manifest row of its own,
so the stale rule leaves it. When its tool leaves the selection, the file
stays, and the sync removes the items that bootstrap added.

| Rule     | `repo`                                                                                                            | `macos`                                       |
| -------- | ----------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| Path     | The path in the repository. A target that is a symlink, or is below a symlinked folder, stays as it is and fails. | The real path. See [mac.md](mac.md#symlinks). |
| Format   | The formatter formats the result before the compare.                                                              | No formatter.                                 |
| Requires | `requires` and `requires_found`.                                                                                  | `requires` only.                              |

## The header

The header of an owned file is two comment lines:

```text
# Synced from bootstrap: <source>. Do not edit this file.
# The next sync writes it again, and each change in this file is lost.
```

The header of a shared file is three comment lines:

```text
# Shared with bootstrap: <spec>.
# A change in this file stays. A removed bootstrap item comes back at the next
# sync, unless <values file> has an override for it.
```

`<values file>` is `.config/bootstrap/tools.toml` on the repo side and
`macos.toml` on the Mac. A header whose third line names another values file,
for example the old `.config/bootstrap.toml`, is the same header: the sync
replaces it with the new line.

A file counts as a synced file (it "has the bootstrap header") when one of its
first 3 lines is a comment line that starts with a header text. The comment
line has an optional indent, then `#`, `//` or `<!--`. A text directly after a
`<!--` line that has nothing else also counts. A text in the middle of a
comment does not count.

| Header text                 | Use                                              |
| --------------------------- | ------------------------------------------------ |
| `Synced from bootstrap: `   | The header of each side now.                     |
| `Synced from repo-config: ` | The old name of `bootstrap`, in older copies.    |
| `Generated by macos-setup`  | The header of the old Mac sync, in older copies. |

Which file gets the header, and where, is in
[tools.md](tools.md#the-header-of-a-source-file).

## The manifest and the stale rule

A manifest lists the owned files of the last clean run and the items that
bootstrap added to each shared file. Its first line is
`# Written by bootstrap. Do not edit this file.`. Each other line is one row,
sorted, with the path relative to the target root:

- `<tool><TAB><path>` for an owned file.
- `<tool><TAB><path><TAB><item>` for an item that bootstrap added to a shared
  file. `<item>` is the identity of the item as JSON on one line.

A seed gets no row. A reader skips each `#` line. The sync writes the file
only when its content changes, with an atomic `mv` and the mode `644`. A write
counts as `changed` or `new`. On the repo side, a dry run lists the manifest as
`would change` or `would add` when the content differs. The dry run of the Mac
side writes no manifest and prints nothing for it. The item rows of a shared
file that a run did not merge stay, so a later run can still remove those
items.

| Side    | Manifest                               | Target root    | Details                         |
| ------- | -------------------------------------- | -------------- | ------------------------------- |
| `repo`  | `.config/bootstrap/repo-generated.tsv` | The repository | [repo.md](repo.md#the-manifest) |
| `macos` | `<cache>/macos-generated.tsv`          | `$HOME`        | [mac.md](mac.md#the-manifest)   |

The sync does not trust a manifest. A row with no TAB, an empty tool or path, an
absolute path, or an empty, `.`, `..` or `.git` segment (in any case) is skipped
with a log line.

**The stale rule** examines each owned-file row of the old manifest whose path
this run did not handle.

- A file that is not there, or that has no bootstrap header, stays with no log
  line.
- A listed path that is the same file as a handled path stays, which covers a
  case variant.
- Else the sync removes the file, then its empty parent folders, up to the
  target root and never past a symlinked folder. A dry run lists the file as
  `would remove`.
- A failed removal is a failure, and the other rows still run.
- The first sync has no manifest, so the stale rule removes nothing.

| Rule               | `repo` link mode                                                                                                    | `macos` link mode                 |
| ------------------ | ------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| A link             | A listed path that is a symlink, or is below a symlinked folder, stays. With verbose, it logs `kept <path> (link)`. | The sync removes the real file.   |
| Another repository | A file below a folder with its own `.git` stays. With verbose, it logs `kept <path> (another repository)`.          | Not examined.                     |
| Manifest file      | A manifest that is a symlink, is below a symlinked folder or is not a regular file is not read.                     | The sync reads the file as it is. |

## The values of a tool

`internal/values` merges the values of each selected tool in this order. A
later level wins:

1. `[defaults]` of the tool.
2. `[contributes.<tool>]` of each selected tool, in tool-name order.
3. The derived values: the output of the `derive` script of the tool. The
   script gets levels 1 and 2 on stdin. See
   [The scripts of a tool](#the-scripts-of-a-tool).
4. `[tools.<tool>]` of the values, without `env` and `projects`, then the
   overrides: `[overrides.tools.<tool>]` of `tools.toml` on the repo side,
   `[tools.<tool>]` of `macos.overrides.toml` on the Mac.

A contribution merges over `[defaults]`: tables merge key by key, a list appends
its new items, and any other value replaces the default. Two contributions that
set one key to different values stop the sync, unless both values are tables or
both are lists. A value in `[tools.<name>]` does not clear this conflict,
because the sync merges the contributions first. Remove one of the two.

The derived values merge over levels 1 and 2: tables merge key by key, and any
other value replaces. The `mise` and `mise_env` tables of the derived values do
not stay in the values: the mise generator gets them. See [mise.md](mise.md).

A contribution to a tool of the side that is not selected is ignored. A
contribution of a selected tool to a tool that is not in the clone stops the
sync with code 3.

A value of `[tools.<name>]` of the values depends on the tool:

- A base or dependency tool is a required member of the pack `base`, or a
  tool in the depends closure of a selected tool, a tool that a mise `depends`
  adds included. For such a tool, a list
  appends its new items, tables merge key by key, and a value whose type
  differs from the merged value stops the sync.
- For any other tool, a list appends its new items, tables merge key by key,
  and any other value replaces the merged one.

So a list of the values never removes an item of `[defaults]` or of a
contribution. To replace a list, give it in the overrides.

On the repo side, the overrides are the `[overrides.tools.<name>]` tables of
`tools.toml`. On the Mac, they are the `[tools.<name>]` tables of
`macos.overrides.toml` next to `macos.toml`. Their values win over all. An
answer in them counts as the answer of a question. A key of `repo.toml` has no
override. A leftover `.config/bootstrap.overrides.toml` next to the values
folder is not read.

The merge runs before the mise config and the file loop, so a failure stops the
sync before it writes a file of a tool, and `bootstrap_commit` stays. A dry run
merges too. Across tools, each message logs and the highest code returns.

When a contributor leaves the selection, its values are gone at the next render.
A seed gets its contributions only when it is first written. A shared file gets
them on each run.

The shape of these tables is in
[tools.md](tools.md#the-defaults-questions-and-contributes-tables).

## `bootstrap_commit`

After a sync without a failure, the front-end sets the top-level
`bootstrap_commit` of its values (`repo.toml` on the repo side, `macos.toml` on
the Mac) to the full commit of its clone. An agent
compares it with `main` to find what changed since the last sync.

- A sync that changed no file keeps a `bootstrap_commit` that has a value, so
  it adds no commit to the consumer. A changed, new, seeded or removed target
  (the manifest too), an added `[tools.<name>]` table or a changed `[home]`
  is a change. A dry run uses the same rule. An empty or missing key always
  gets the commit.
- A line that has the key before the first table is replaced. A line with the
  same value causes no write.
- Else the key goes after the header comments, when a blank line ends them, or
  else on the first line. A comment just above a key belongs to that key.
- A `format` key at that place, with a blank line and more keys after it, is
  a block of its own and stays first. The key goes after that blank line, in
  the block of the other keys, so the dprint `sortKeys` of a consumer keeps
  each line.
- Each other line stays. `sync` writes only this key and the tables of the
  required members of `base` to the values. See
  [The tables of the required members](#the-tables-of-the-required-members).
- Through a symlink, the sync writes the real file, so the link stays.
- A real run prints the info line `· bootstrap_commit changed to <short
  commit>`,
  and a dry run prints `· bootstrap_commit would change to <short commit>`. With
  verbose,
  the lines are `set bootstrap_commit <commit> in <path>` and
  `would set bootstrap_commit <commit> in <path>`.
- A missing `macos.toml` is created. See [mac.md](mac.md#macostoml). On the repo
  side, a missing values folder stops the sync before this step.

### The tables of the required members

`sync`, on both sides, writes an empty `[tools.<name>]` table into the values
(`tools.toml` on the repo side) for each required member of `base` that has no
table, so that the user
can add values to it. This also applies when the user deleted the table by
hand. It skips a member whose `hosts` do not name the host, a member with
`selectable = false`, and in a submodule member a member with `root_only`. A
dependency without a table gets no table: the sync installs it through
`depends`.

- The write keeps each other line of the file (see
  [The layout of a values file](#the-layout-of-a-values-file)), and it comes
  before the files of the tools.
- Each table prints the info line `[tools.<name>] added to <values file>`.
- `sync --dry-run` and `check` write nothing. They print
  `[tools.<name>] would be added to <values file>` for each table. The line
  does not change the exit code.

## The output of a run

The code of the output is in `internal/report`, so both front-ends print the
same lines. `internal/cli` prints its own lines, for example the change list,
the prompts and the status lines, with the same glyphs and the same style rule.
With `--json`, stdout holds the envelope only, and each line of this section
goes to stderr. See [The JSON contract](#the-json-contract).

**The glyphs.** One set of glyphs is for all commands. Each glyph is one column
wide, so the columns stay aligned.

| Glyph     | Meaning                                 | Color                    | Style off            |
| --------- | --------------------------------------- | ------------------------ | -------------------- |
| `✔`       | A success, a match, a valid item        | Green                    | `[OK]`               |
| `✗`       | An error                                | Red                      | `[ERROR]`            |
| `!`       | A warning                               | Yellow                   | `[WARN]`             |
| `·`       | An info line                            | Dim                      | `[INFO]`             |
| `↻`       | The clone and its commit                | Dim                      | `[INFO]`             |
| `▸`       | A step that runs a command              | Bold                     | `[INFO]`             |
| `→`       | The fix of an error                     | Dim                      | `  -> <fix>`         |
| `○`       | A tool selected directly, nothing found | Yellow                   | `[WARN]`             |
| `●`       | A required tool in `list`               | None                     | `*`                  |
| `↑`       | A later copy of a subtree in a tree     | None                     | `^`                  |
| `✎ ✚ ❋ ✖` | A changed, new, seeded, removed file    | Yellow, green, blue, red | The word of the kind |

An error line is `✗ <problem>`. When the error has a fix, a second line
`  → <fix>` follows, with an upper case first letter. The fix is a command to
run, for example `  → Run bootstrap add tool github`, or a short instruction. An
error that has no fix is one line.

**The first line.** A command that uses the clone prints one line after the
clone update, except `tui` and `mac tui`: `↻ bootstrap <short commit> · <cache
path>`, with `$HOME` shown as
`~`. With `--clone` or `BOOTSTRAP_CLONE`, the path is the clone that the command
uses. `sync`, `mac sync` and `mac check` print `↻ bootstrap <short commit> ·
<n> tools · env <environments>` in place of that line, and the Mac side has no
env part. A command prints one clone line only: after the change list of
`migrate`, `add`, `remove`, `replace` and `detect --add`, the sync does not
print
its own. The other lines of the clone are info lines: `· Cloning into <path>`,
`· Waiting for another sync (pid <pid>) ...` and `· Removed an old clone at
<path>`.

**The outcome of a target.** The outcome of a target (a kind, a path, and an
optional text and log line) replaces each per-file log line. It increments the
count of the kind and prints the correct line.

| Kind        | Lines of the log                                      | List                | Dry-run word   | Count in the last line |
| ----------- | ----------------------------------------------------- | ------------------- | -------------- | ---------------------- |
| `changed`   | `updated`, `chmod`, `wrote` of a manifest that exists | `✎ changed`, yellow | `would change` | In the headline        |
| `new`       | `updated`, `wrote` of a target that does not exist    | `✚ new`, green      | `would add`    | `new`                  |
| `seeded`    | `updated` of a seed                                   | `❋ seeded`, blue    | `would seed`   | `seeded`               |
| `removed`   | `removed`                                             | `✖ removed`, red    | `would remove` | `removed`              |
| `failed`    | Each failure message of a target                      | `✗ failed`, red     | `failed`       | `failed`               |
| `unchanged` | `unchanged`                                           | None                | None           | `unchanged`            |
| `kept`      | `kept seed`                                           | None                | None           | `kept seeds`           |
| `skipped`   | `skip seed`, `skip merge`, `skip file`, `skip env`    | None                | None           | `skipped`              |

The list lines start at the start of the line, with no indent and no blank line
before the list: `✎ changed  <path>`, `✚ new      <path>`.

- **A failed target:** It counts once, and it has one list line, from its first
  failure message: `✗ failed   <path>: <text>`. `<path>` is the target, and the
  text does not repeat it. A target that fails with no message gets the text
  `cannot write the target`. A later line for the same target, for example
  `kept existing`, prints only with verbose.
- **No count:** A failure of the mise config of a tool or of the manifest file
  has no count, and its message stays. The info line `FAILED: <list>` stays
  before the last line. It names the source file of each failure, not the
  target, so it does not repeat the list lines of the failed targets.
- `bootstrap_commit` prints the info line `· bootstrap_commit would change to
  <short commit>` (`changed to` for a real run) before the last line.

**The last line.** The summary of a run prints it once, as the last output of
the run: after the stale rule, the manifest write and `bootstrap_commit`. A run
prints it after the file loop, also on a failure. A run that stops before the
file loop has no last line. There is no table. `n` is the sum of the counts
`changed`, `new`, `seeded` and `removed`.

| Run              | Headline                                                                  |
| ---------------- | ------------------------------------------------------------------------- |
| Real run, code 0 | `✔ Synced: <n> files changed`, or `✔ Synced: no file changed`             |
| Dry run, code 0  | `✔ Dry run: <n> files would change`, or `✔ Dry run: no file would change` |
| Another code     | A red `✗ The sync failed with code <rc>`                                  |

The headline uses `1 file` for one file. After the headline, ` · <count>
<label>` follows for each kind that is not zero, in the order `new`, `seeded`,
`removed`, `unchanged`, `kept seeds`, `skipped`, `failed`; the kind `changed` is
in the headline. A run with another code lists `failed` first, and then each
other kind that is not zero, with `changed` after `failed`. For example:

```text
✎ changed  ~/.config/mise/config.toml
✚ new      ~/.config/zed/settings.json
✔ Synced: 2 files changed · 1 new · 161 unchanged · 4 kept seeds
```

```text
✗ failed   .vscode/settings.json: the merge spec is not valid JSON
✗ The sync failed with code 1 · 1 failed · 1 changed · 60 unchanged
```

**The change list.** The change list of `migrate`, `add`, `remove`, `replace`
and
`detect --add` has the header `Changes`. Each line is `  <glyph> <path>` with
the glyph of what the sync does to the file (`✎` changed, `✚` new, `❋` seeded,
`✖` removed), and a note column that is aligned, for example `remove claude` for
the values file. A script that runs before the writes has the glyph `▸`. A dry
run ends with `✔ Dry run: <n> changes, nothing written`. Else the prompt is
`? Write <n> changes? [y/N]`.

```text
Changes
  ✎ .config/bootstrap/tools.toml          remove claude
  ✖ .config/mise/tasks/setup/ai.py
  ✎ .config/bootstrap/repo-generated.tsv

✔ Dry run: 3 changes, nothing written
```

When the sync of the command fails, the values go back as they were before
the command (see [The change list](#the-change-list)), and the info line
`· Put back <values file> as it was before the command` follows the error of
the sync. After a failed migration, the line names `.config/bootstrap.toml`.
On the Mac, code 4 keeps the values file and prints no such line.

**Errors.** Each error prints `✗ <problem>` and, when it has a fix, the line
`  → <fix>`. The problem is a short statement, and the name of a tool has no
quotes: `✗ nosuchtool is not a bootstrap tool`, `✗ git is required by the base
pack, so remove cannot remove it`. The front-ends and the packages below them
write a fix after `-` in the message; the output puts it on its own line.
Verbose does not change an error.

A usage error prints `✗ <problem>`, then `  usage: bootstrap <command> <its
flags>` for that command only, then `  → Run bootstrap --help for all commands`.
An unknown command, or a command that is missing, has no usage line:
`✗ frob is not a command`.

```text
✗ sync has no flag --bogus
  usage: bootstrap sync [--dry-run] [--jobs <n>] [--verbose] [--json]
  → Run bootstrap --help for all commands
```

**The check.** `check` prints `✔ <values folder> is valid` when the local rules
and the schema of the values folder pass. It lists each missing table of a
required member of
`base` and writes nothing. See
[The tables of the required members](#the-tables-of-the-required-members). Then
it prints `✔ check script of <tool>` for each check
script that ran. A script that fails prints `✗ check script of <tool>: <text>`
and, when the script gives a fix after `-`, the fix line. The exit codes do not
change. The local rules are the rules that the engine checks without a tool
script: the environments, the packs, the members, the overrides and the
project folders. After the check scripts, it runs the detect scripts of the
projects of each tool, and prints a warning for each link that no longer
matches. When `BOOTSTRAP_CLONE` is not set, the check scripts do not run, and a
warning says so.

**The step of the first setup.** The first setup of a Mac prints
`▸ Running mise bootstrap --yes` before the command and `✔ mise bootstrap done`
after a success. The output of
mise stays as it is. A failure is the error `✗ mise bootstrap failed with exit
code <rc>`.

**The style rule.** One rule turns the glyphs and the colors on and off
together: `FORCE_COLOR` (not `0`) turns them on, `NO_COLOR` turns them off, and
otherwise they need a terminal on stderr and a `TERM` that is not `dumb`. This
is the rule of `style_enabled` of `helpers.py`. The engine does not call
`helpers.py`, because the local copy can be stale and the Mac side does not have
it. The style table:

| Line    | Style on                          | Style off                      |
| ------- | --------------------------------- | ------------------------------ |
| List    | `<glyph> <word> <path>`, color    | `<word> <path>`                |
| Change  | `  <glyph> <path>  <note>`, color | `  <word> <path>  <note>`      |
| Success | `✔ <text>`                        | `[OK] <text>`                  |
| Error   | `✗ <text>`, red                   | `[ERROR] <text>`               |
| Fix     | `  → <fix>`, dim                  | `  -> <fix>`                   |
| Warning | `! <text>`                        | `[WARN] <text>`                |
| Info    | `· <text>`, dim                   | `[INFO] <text>`                |
| Clone   | `↻ <text>`, dim                   | `[INFO] <text>`                |
| Step    | `▸ <text>`, bold                  | `[INFO] <text>`                |
| Tree    | `●`, `↑`, `✔ ○ ·`                 | `*`, `^`, `[OK] [WARN] [INFO]` |
| Prompt  | `? <text>`                        | `? <text>`                     |

The usage line of a usage error and the output of `settings` and `config` have
no glyph, except the info lines of `config set` and `config unset`. The
verbose log lines have the prefix `bootstrap: ` with style on and with style
off.

**Verbose.** `--verbose` of the engine sets `BOOTSTRAP_VERBOSE=1`, on `migrate`,
`tui`, `sync`, `check`, `add`, `remove`, `replace` and `detect`. `--verbose` of
`setup:configs` and `updateall:configs` also sets the variable. Each entry task
consumes the flag and does not pass it to the engine. A user can also set the
variable, and only the value `1` turns it on. The config file has no `verbose`
key. Each target then prints its log line in place of its list line, for
example `unchanged  <path>`, `updated    <path>`, `kept seed  <path>`,
`skip env   <path>` or `would update <path>`. The last line stays. These lines
print only with verbose:

- `using bootstrap at <commit>`, `tools:` and, on the repo side,
  `environments:`. By default, they are the first line. See
  [repo.md](repo.md#the-entry-tasks) and [mac.md](mac.md#the-output).
- `kept <path> (link)` and `kept <path> (another repository)`.
- `done: <ok>/<total> files in sync`, `<ok>/<total> files in sync; existing
  copies of the rest were left untouched` and, on the Mac side, `dry run:
  <ok>/<total> files checked, nothing written`.

The tests set the variable, so a case compares the log lines. The cases of
the default output unset it.

### The progress

Each command shows a spinner `<frame> <text>` on the last line of stderr while
a slow step runs. The lines of the step show above the spinner, and the
spinner line goes at the end of the step. The spinner shows only when stderr
is a terminal, and never with `--json`.

| Step                                         | Text                                  | Commands                                                                   |
| -------------------------------------------- | ------------------------------------- | -------------------------------------------------------------------------- |
| The fetch of the clone (not with `--clone`)  | `Updating the clone of bootstrap ...` | Each command that uses the clone, except `config get`, which never fetches |
| The plan: the dry run of the sync            | `Planning ...`                        | `migrate`, `sync`, `add`, `remove`, `replace`, `detect --add`              |
| The write and the sync                       | `Syncing ...`                         | `migrate`, `sync`, `add`, `remove`, `replace`, `detect --add`              |
| The detect scripts, or the package scan      | `Detecting ...`                       | `detect`                                                                   |
| The check scripts, or the dry run of the Mac | `Checking ...`                        | `check`                                                                    |

A step that asks the user for input never spins: the questions and the consent
come between the spinners. The first setup of a Mac does not spin, because
its `mise bootstrap` can ask for a password; with `--no-bootstrap`, it spins.

## The exit codes

A failure of one file does not stop the other files. A run returns the highest
code of all its failures. This section is the one place of the codes 1, 2, 3
and 5. Code 4 is only for the Mac side: see
[mac.md](mac.md#the-first-setup-and-exit-4).

| Code | Meaning                                                                                                |
| ---- | ------------------------------------------------------------------------------------------------------ |
| 0    | Done, nothing to change, a "no" at the consent prompt, or a quit of the TUI.                           |
| 1    | Any other failure, for example a local file system error. `detect` without `--add` with a tool to add. |
| 2    | A usage error, or a config that breaks a rule. See the lists below.                                    |
| 3    | A name that the clone does not have. See the lists below.                                              |
| 5    | A failure of the network side or of the clone. See the lists below.                                    |

**Code 2, of the command line:**

- A bad flag or command, `add`, `remove` or `replace` without a kind word or
  without a name (the old grammar, with the hint of the new form), `migrate`
  with
  names, `remove pack` and `replace pack`, `--exclude` outside `add pack` or
  with more than one pack, and `--check` with `--dry-run` of the front-end.
- A config file that breaks a rule: see [The config file](#the-config-file).
- A missing `mise`, a folder outside a git repository, and a missing values
  folder, `repo.toml` or `tools.toml` for each command except `migrate`,
  `version`, `settings`, `contract`, `catalog` and `mac`. Each `config`
  subcommand also stops with code 2 and "Run bootstrap migrate" when
  `.config/bootstrap/` is absent, unless `BOOTSTRAP_VALUES` is set.
- The old values file `.config/bootstrap.toml` (except for `migrate`, `version`,
  `settings`, `contract`, `catalog` and `mac`), and the old file next to the
  values folder (each command but `version`, `settings`, `contract`, `catalog`
  and `mac`). An old form of the Mac values file (except for `mac migrate`): a
  file without `format = 2`, and a `[packs.*]` table. A `format` other than `3`
  in the values folder, or `2` on the Mac, is not valid.
- A fresh `migrate` with no `main`, no `develop` and a detached `HEAD`, and a
  migration of a repository with no `main` and no `develop`, or with a
  `[tools.git.branch.<b>]` table of another branch or of a branch that does not
  exist.
- A `config` key that is not valid: not a key path, a first segment other than
  `repo`, `tools` or `releases`, a key that the schema does not have, or on
  `set` and `unset` a `tools.*` key that is not `env`, `files.*`, a question or
  a default of the tool. A `config set` of a value of the wrong type or outside
  its enum, of a key of a tool that is not selected (fix "run bootstrap add tool
  <tool>"), a `config unset` of a required key, and a `config set` or `unset`
  whose result breaks a rule of the schema.
  A question of a selected tool without an answer, and a group of questions
  without an answer that is not empty, with the message "run bootstrap migrate"
  ("run bootstrap mac migrate" on the Mac).
- An `--answer` for a question that no selected tool has, an empty `--answer`
  for a question that is not optional and not in a group, a question without
  an answer with `--yes` or without a terminal, a group without an answer that
  is not empty with `--yes` or without a terminal, and ctrl+c in the field of
  a question.
- An `add tool` or `replace tool` of a named tool with `selectable = false`,
  or whose `hosts` do not name the host. See
  [Named tools that a command refuses](#named-tools-that-a-command-refuses).
- Changes without `--yes` and without a terminal: `changes need --yes`.
- `--json` on `migrate`, `add`, `remove`, `replace` or `detect --add` without
  `--yes` and without `--dry-run`, `--json` on a command that does not take it
  (also each `mac` command except `mac questions`), `--json` with `--format`,
  and with `--json --yes` a question without an answer. See
  [The JSON contract](#the-json-contract).
- `tui` and `mac tui` without a terminal, with a change to migrate, or in a
  terminal smaller than 100 × 25. See [The TUI](#the-tui).
- A dirty target: see [The clean-target rule](#the-clean-target-rule).
- A binary that is older than `VERSION` of the clone.
- A `remove tool` of a required tool, of a tool that is not selected, or of a
  tool that another tool needs.
- An `add pack --exclude` of a name that is not a member of the pack, of a
  required member, or of a member that another added member needs.
- A `replace` between tools without the same `parent`, without a migrate
  script, of a new tool that is already selected, or of an old tool that
  another tool needs.
- A `detect --yes` or `detect --answer` without `--add`, and a `mismatch` of a
  selected tool.

**Code 2, on both sides:**

- A `repo/` of the cache that is not a clone and not empty, and a repository at
  the cache root with another `origin` or no `origin`.
- A `tools.toml` that cannot be read or that breaks a rule, also an old key.
- A `[pack.*]` table that breaks a rule: an old key, a member that is not a
  table or is not on one line, the key `default` in a member (the old form),
  another key in a member, a `description` that is not a string, a pack name
  that is also a tool folder, and a pack loop.
- A catalog key of a tool that breaks a rule, and a categories file that breaks
  a rule. See [tools.md](tools.md#the-catalog-keys).
- A `required` key outside the pack `base`, or on a nested pack.
- A tool folder that is named `zz-repo` or `zz-macos` for its side, or whose
  name sorts after that name.
- A file of a tool folder outside `files/`, `tasks/` and `scripts/` (the old
  form), a file below `files/` or `tasks/` without an entry, and a file of
  `scripts/` that is not a script of a tool.
- A tool that ships a mise file that the sync also generates.
- A bad `[files]` or `[tasks]` table: another key, a target or a source that
  fails, or a bad value. A `parent` that is not another tool. A `replaces`
  that is not a non-empty list of `sha256:<64 hex>` in lower case, or of a
  task outside the folder of a group.
- A bad `[defaults]`, `[questions]` or `[contributes]` table, and two
  contributions or a user value that conflict.
- A bad `[[contributes_items.<owner>."<target>".item]]` table (repo side): a
  bad shape, an owner that is not a tool of the clone, or a target that the
  owner does not declare as a merge-mode file.
- A bad overrides file: a key other than `[tools.<name>]` tables.
- A script that fails, prints bad output, or a check script that finds a
  problem.
- A template that cannot render, and a target that does not parse.
- A merge spec that breaks a rule, an override of a shared file that breaks a
  rule, and a live shared file that cannot merge.
- A `[mise.*]` config that the generator cannot write: another section, a failed
  render, a rendered copy that is not valid TOML or that changes the key paths.
- A bad `[overrides.mise*]`: a path that no selected tool has, another key of
  `[overrides]`, an environment outside `environments`, and the `envs`, the
  `depends`, the `requires`, the `requires_found` or the `files` of a mise tool.

**Code 2, only on the repo side:**

- A values file in the old list format (`tools = [...]`, `packs = [...]`).
- A bad `environments`: an unknown name, or a list without `dev`.
- A bad `env`: a name that is not in `environments`, a list without `dev`, an
  empty list, or an item with a space.
- A name other than `dev` and `ci` in the `default_envs` or the `envs` of a
  `tools.toml`, and a `default_envs` without `dev`, also of a tool that is not
  selected.
- An environment that a mise tool does not allow in its `envs`, also of a
  dependency or through `default_envs`.
- `requires` or `files` in a `[mise_env.<env>.tools]` entry, and a
  `requires_found` there that is not a non-empty list of file names.
- Two tools that define the same mise tool, and two selected tools with a
  `[formatter]`.
- A bad `[lsp]` of a `tools.toml`, also of a tool that is not selected: an
  `[lsp]` that is not a table, an `[lsp.<server>]` that is not a table, a key
  of a server other than `command`, `args`, `extensions`, `startup_timeout`
  and `source`, no `command` or a `command` with whitespace, `args` that is
  not a list of strings, no or empty `extensions` or a key of it that is not
  `".<ext>" = "<language id>"`, a `startup_timeout` that is not an integer
  above 0, and no `source`.
- One server name in two tools, and a `source` that is neither `"host"` nor a
  mise tool of the same tool.
- A key of the values folder that the schema does not have, a value of the
  wrong type or outside its enum, a missing required key, and a key of
  `repo.toml` in `[tools.*]` or `[overrides.tools.*]`. See
  [bootstrap-config.md](bootstrap-config.md#the-schema).
- A `root_only` tool that a `[tools.*]` entry names in a submodule member.
- A bad `[projects.<name>]` table: a name that is not lowercase kebab case, no
  `folder`, a folder that is absolute or has a `..` segment, two projects with
  one folder, and, for `check`, a folder that is not there. A
  `[tools.<tool>].projects` that is not a list of strings, that names a project
  that `repo.toml` does not have, or of a `root_only` tool. A detect script that
  fails in a project folder, for `detect` and `check`.
- A bad `[members.*]` table: a name with other characters, two members with one
  name, no `type`, a `sibling` without `url`, a folder that is not in the
  repository, a `folder` of `.`, and a `type` that is not correct. A `[home]`
  without `folder` or `type`.
- A shared file that is a symlink or is below a symlinked folder.
- A `.config/bootstrap/repo-generated.tsv` that is a symlink, is below a
  symlinked folder or is not a regular file.
- A task of the repository at the place of a test port, for `sync` and
  `check`, with the fix "run bootstrap migrate". On `migrate`, a target
  `<group>/50-local` of its move that exists, and a temp name
  `<path>.bootstrap-move` that exists. See [`migrate`](#migrate).
- A known old copy of a part of a selected tool at `<group>/50-local`, for
  `sync` and `check`, with the message
  `<path> is an old copy of the task <part> of the tool <tool> - run bootstrap
  migrate`.
  At the old path of the group, the message is that of a task at the place of
  a test port. See [`migrate`](#migrate).

**Code 2, only on the Mac side:**

- A `mise` that is not at `~/.local/bin/mise`.
- A `HOME` that is not set.
- A Mac with a manifest and no `macos.toml` at its fixed place. See
  [mac.md](mac.md#the-place-of-macostoml).
- A `BOOTSTRAP_CLONE` without `cli/` or `macos/tools/`.
- A `macos.toml` that breaks a rule, also a `[overrides.mise_env.*]` table.
- A nested run of `updateall:configs`.
- A key of a Mac `tools.toml` that is for repo tools only, a `requires_found` of
  a `[files]` table, and a `[mise_env.*]`, an `envs` or a `requires_found` of a
  mise tool.
- Two selected tools that ship one target, and a tool that ships a file below
  `.config/mise/conf.d/`.

**Code 3:**

- A tool that is not in the clone, also a member of a pack that is not a tool
  folder or a pack, and a tool in a `depends`.
- A packs file that is missing or has no `[pack.base]`:
  `base is not a bootstrap pack (named in <side>/packs.toml)`.
- A tool name of `add tool`, `remove tool`, `replace tool` or `questions`, or a
  pack name of `add pack`, that is not in the clone.
- An old values file that names a pack that the frozen copy of the migration
  does not have: `migrate` stops with
  `<pack> is not a bootstrap pack (named in <file>)`. See
  [bootstrap-config.md](bootstrap-config.md#old-forms).
- A contribution of a selected tool to a tool that is not in the clone.

**Code 5:**

- A clone or a fetch that fails. For a catalog read, only when the cached
  clone cannot be used: see [The cache and the clone](#the-cache-and-the-clone).
- A `cli/VERSION` of the clone that is not `MAJOR.MINOR.PATCH`.
- A `BOOTSTRAP_CLONE` that is not usable, on the repo side.

`bootstrap check` returns the highest code of its stages. See
[The JSON contract](#the-json-contract) for the stages. With `--json`, the
exit code of each command is the `code` of its envelope.

## The release of the engine

The engine is a binary of a GitHub release of `virajp/bootstrap`. mise
installs it with its `github:` backend:
`mise use -g github:virajp/bootstrap`. The release tags
are `bootstrap@X.Y.Z`, so that other tags of the repository can have other uses.
mise removes a tag prefix that is the name of the repository, so it shows the
tag `bootstrap@2.0.2` as the version `2.0.2`, and a command needs no option. The
workflow
`.github/workflows/release-engine.yml` builds the binary and makes
the release.

### `VERSION`

`cli/VERSION` is the minimum version of the engine that a clone of
`main` needs. See [The version check](#the-version-check).

- It has one line, `MAJOR.MINOR.PATCH`, with no `v` and no comment.
- The tag `bootstrap@X.Y.Z` is the source of the version. The binary gets
  `X.Y.Z`
  from
  the build, and the release workflow writes `X.Y.Z` to `VERSION`.
- Nobody edits `VERSION` by hand. A manual change can name a version that has no
  release.

### The rule

The release exists before `VERSION` moves. The workflow commits `VERSION` only
as its last step, after a successful build and release. If a step fails before
it, `VERSION` does not change, so no consumer needs a version without a release.

A tool change that needs a new feature of the engine lands after the pin commit.
Before that commit, the clone does not ask for the new engine.

### Release a version

**CAUTION:** Do not tag a commit that is not on `main`. The workflow stops, and
the ruleset `all-tags` does not let you move or delete the tag.

1. Land the engine change on `main` with `mise run git:merge:main <branch>`.
2. Pull `main`.
3. Select the next version `X.Y.Z`.
4. Make a signed tag on the commit of `main`:
   `git tag -s bootstrap@X.Y.Z -m "bootstrap@X.Y.Z"`.
5. Push the tag: `git push origin bootstrap@X.Y.Z`.
6. Monitor the run of `release-engine` in GitHub Actions.
7. When the run is done, pull `main`. Its last commit is
   `ops: pin bootstrap X.Y.Z`.

The ruleset `all-tags` requires a signed tag. The tag must be
`bootstrap@MAJOR.MINOR.PATCH`, with no leading zero and no suffix.

### What the workflow does

The workflow starts on a push of a tag `bootstrap@*.*.*`. Its one job runs on
`macos-15`,
an arm64 runner, with the permission `contents: write` only.

1. It examines the tag. The tag must be `bootstrap@MAJOR.MINOR.PATCH`, and its
   commit
   must be on `main`. Else the run stops.
2. It installs the Go version of `cli/go.mod`.
3. It builds `./cmd/bootstrap` in `cli/` for `darwin/arm64`, with
   `CGO_ENABLED=0` and `-ldflags "-X main.version=X.Y.Z"`.
4. It runs `bootstrap version`. The output must be `X.Y.Z`.
5. It writes the SHA-256 checksum of the binary.
6. It creates the GitHub release `bootstrap@X.Y.Z` with two assets:
   `bootstrap-darwin-arm64` and `bootstrap-darwin-arm64.sha256`.
7. It commits `cli/VERSION` = `X.Y.Z` on `main`.

The build is for macOS arm64 only. No Linux build and no Intel build exist. The
pre-commit hook runs the gate, and a gate without its tool exits 0, so a CI job
without `bootstrap` passes.

mise reads release assets, not workflow artifacts. It selects the asset of the
OS and the CPU, and it removes the OS and CPU suffix from a single-file binary.
Thus `bootstrap-darwin-arm64` installs as `bootstrap`. mise also examines the
download against the digest of the asset or the checksum file.

### The pin commit

The ruleset `all-branches` requires signed commits. A `git push` from the runner
is not signed, so the ruleset refuses it. Thus the last step uses the GraphQL
mutation `createCommitOnBranch`, and GitHub signs that commit.

- The step reads the head of `main`, and gives it to the mutation as
  `expectedHeadOid`.
- If `main` moves before the mutation, the mutation fails. The step reads the
  head again and tries again, up to five times.
- If `VERSION` on `main` is already `X.Y.Z`, the step does nothing.
- The commit is from `github-actions[bot]`. A commit of the workflow token
  starts no other workflow.

If the last step fails, the release exists and `VERSION` does not move. Run the
failed job again in GitHub Actions. The release step finds the release and does
not change it, and the last step tries the pin commit again.
