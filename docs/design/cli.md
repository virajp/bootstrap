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
  `mise x` and `mise bootstrap --yes`, and on the Mac it needs the standalone
  mise.
- git: the repository, the members, the clone and `bootstrap_commit`.
- The values file and its overrides file, the packs, the manifest, the header,
  the symlink rules, the stages and the exit codes.

Each tool installs only through mise: `[mise.tools]`, and the `brew:` and
`brew-cask:` entries of `[mise.bootstrap.packages]` on the Mac. bootstrap
touches only the files of the tools that a values file selects. Each other
file belongs to the user. Each sync handles only its own repository, never a
submodule: each submodule runs its own sync.

## The files of `cli/`

`cli/` is not a tool. The sync never copies it to a consumer. It is the
Go module `github.com/virajp/bootstrap/cli`.

| Path                 | Job                                                                                                                                                                                 |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `VERSION`            | The minimum version of the engine that a clone of `main` needs. See [The version check](#the-version-check).                                                                        |
| `go.mod`             | The module.                                                                                                                                                                         |
| `cmd/bootstrap`      | The command `bootstrap`. It calls `internal/cli`. `version` prints `main.version`, which the release build sets. A local build has `dev`.                                           |
| `internal/cli`       | The command line: it parses the arguments, calls `internal/core`, prints the results and the errors, and asks the questions and the consent.                                        |
| `internal/core`      | The preflight, the clone and its lock, the version check, the change list with `Plan` and `Apply`, the scan of `detect`, and the clean-target rule. It never prints and never asks. |
| `internal/config`    | The user config file of the engine: its path, its keys and its checks. See [The config file](#the-config-file).                                                                     |
| `internal/oldform`   | The old forms of a values file, and their migration. It is the one place of the old names.                                                                                          |
| `internal/questions` | The questions of the selected tools, their proposals and their answers.                                                                                                             |
| `internal/field`     | The text field of a question: the question, its note, and a field that starts with the proposal. It uses Bubble Tea v2 and the `textinput` of Bubbles v2.                           |
| `internal/scripts`   | The run of the scripts of a tool. See [The scripts of a tool](#the-scripts-of-a-tool).                                                                                              |
| `internal/fail`      | The typed errors of the exit codes, and a collector where the highest code wins. See [The exit codes](#the-exit-codes).                                                             |
| `internal/report`    | The output of a run. See [The output of a run](#the-output-of-a-run).                                                                                                               |
| `internal/source`    | The read of the clone: the packs file, the categories file, the `tools.toml` of each tool, the entries and the files of a tool folder, and the checks of these files.               |
| `internal/values`    | The values files and the overrides files of both sides, and the values merge. See [The values of a tool](#the-values-of-a-tool).                                                    |
| `internal/selection` | The checks of the packs, the required members of `base`, the depth-first walk of the `depends`, and the stages of the Mac.                                                          |
| `internal/tmpl`      | `text/template` with the delimiters `<%` and `%>`, and our own copy of the gomplate functions that the templates use. It does not import gomplate.                                  |
| `internal/format`    | The parse check, and the merge of a shared file: JSON, JSONC, YAML, TOML and text. See [Shared files](#shared-files).                                                               |
| `internal/header`    | The header of a synced file and of a shared file, and its detection. See [The header](#the-header).                                                                                 |
| `internal/target`    | The write of one target, the modes, `bootstrap_commit`, and the tables of the required members of `base`. See [The write of one target](#the-write-of-one-target).                  |
| `internal/manifest`  | The manifest and the stale rule, in the link modes `repo` and `macos`. See [The manifest](#the-manifest-and-the-stale-rule).                                                        |
| `internal/misegen`   | The mise config generator of both sides. See [mise.md](mise.md).                                                                                                                    |
| `internal/repo`      | The repo front-end, the sync of `bootstrap sync`. See [repo.md](repo.md).                                                                                                           |
| `internal/macos`     | The Mac front-end, the sync of `bootstrap mac sync`. See [mac.md](mac.md).                                                                                                          |

The task `test:go` of this repository runs `go vet`, a `gofmt` check and
`go test` in the module. Some unit tests compare the output of a case with a
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
`.config/bootstrap.toml` has no doubt. Each command without `mac` is for the
repository of the current folder. The `mac` commands have the same grammar.

The kind word tells what the names are: `tool` or `pack`. `add` takes `tool` or
`pack`. `remove` and `replace` take `tool` only. The other commands take no
kind word.

| Command                    | Job                                                                                                                                                             |
| -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `init`                     | Migrates an old form, or creates the values file with a table for each member of `base`, asks each question without an answer, then syncs. See [`init`](#init). |
| `sync`                     | The sync of the side. `setup:configs` and `updateall:configs` run it.                                                                                           |
| `check`                    | Examines the values file and runs the check scripts. On the Mac, a dry run of the sync.                                                                         |
| `questions [tool…]`        | Lists the questions, their proposals and their answers. Without names, the questions of the selected tools. `--json` prints one JSON array.                     |
| `add tool <tool…>`         | Selects the tools and their dependencies, asks their questions, then syncs.                                                                                     |
| `add pack <pack…>`         | Selects each member of the packs that is not selected, and their dependencies, asks their questions, then syncs. See [`add` and `remove`](#add-and-remove).     |
| `remove tool <tool…>`      | Removes the tools, and each dependency that no other tool needs, then syncs. Each seed of a removed tool stays, with a line.                                    |
| `replace tool <old> <new>` | Runs the migrate script of `<new>` for `<old>`, swaps the tools, then syncs. See [`replace`](#replace).                                                         |
| `detect`                   | Runs the detect scripts, or on the Mac the package scan, and lists each tool that matches. See [`detect`](#detect).                                             |
| `list`, `ls`               | Prints the selected tools as a tree. See [`list`](#list). It writes nothing.                                                                                    |
| `config`                   | Prints the config file and each setting, with its value and its source. See [The config file](#the-config-file). There is no `mac config`.                      |
| `version`, `--version`     | Prints the version. There is no `mac version`.                                                                                                                  |

| Flag                            | Commands                                                      | Effect                                                                                                                                   |
| ------------------------------- | ------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `--dry-run`                     | `init`, `sync`, `add`, `remove`, `replace`, `detect`          | Prints the change list, and writes nothing.                                                                                              |
| `--yes`, `-y`                   | `init`, `add`, `remove`, `replace`, `detect --add`            | Writes the changes without a prompt.                                                                                                     |
| `--no-tui`                      | `init`                                                        | Selects no tool on a screen. Until the TUI is there, `init` takes this path in each case.                                                |
| `--add`                         | `detect`                                                      | Adds each matched tool that is not selected, after the consent.                                                                          |
| `--exclude <member>`            | `add pack`                                                    | Leaves out a member of the pack, as the packs file writes it. The flag can repeat. It needs exactly one pack name.                       |
| `--answer <tool>.<key>=<value>` | `init`, `add`, `replace`, `detect --add`                      | The answer of a question. The flag can repeat. `--answer <tool>.<key>=` gives an empty answer, only for an optional or a group question. |
| `--json`                        | `questions`                                                   | JSON output, for an AI agent.                                                                                                            |
| `--no-bootstrap`                | each `mac` command except `questions`                         | Skips each `mise bootstrap` step.                                                                                                        |
| `--jobs <n>`                    | each command except `questions`, `list` and `version`         | The jobs at a time. The default is 2 × the CPU count.                                                                                    |
| `--verbose`                     | `init`, `sync`, `check`, `add`, `remove`, `replace`, `detect` | Sets `BOOTSTRAP_VERBOSE=1`. See [The output of a run](#the-output-of-a-run).                                                             |
| `--config <path>`               | each command except `version`                                 | The config file. See [The config file](#the-config-file).                                                                                |
| `--cache-dir <path>`            | each command except `version`                                 | The cache folder. It sets `BOOTSTRAP_CACHE`.                                                                                             |
| `--url <url>`                   | each command except `version`                                 | The clone URL. It sets `BOOTSTRAP_URL`.                                                                                                  |
| `--clone <path>`                | each command except `version`                                 | A clone of bootstrap, with no lock and no fetch. It sets `BOOTSTRAP_CLONE`.                                                              |
| `--repo-path <path>`            | each `mac` command except `mac init`, and `config`            | The repository of `macos.toml`. It sets `MACOS_SETUP_REPO_PATH`.                                                                         |

A flag of a repo command also applies to the same `mac` command, for example
`mac sync --verbose`. A flag with a value takes `--flag value` and
`--flag=value`. An empty value is a usage error. The one exception is the
empty answer `--answer <tool>.<key>=`. See [The questions](#the-questions).

A flag that a command does not take, an unknown command, and `add`, `remove`
or `replace` without a name are usage errors (code 2). `add`, `remove` and
`replace` without a kind word stop with code 2 and the hint of the new form,
for example `use: bootstrap add tool node`. `init` with names stops with code 2
and the hint `use: bootstrap add tool <names>`. A usage error prints the
problem, the usage line of that command only, and a pointer to the help. See
[The output of a run](#the-output-of-a-run). `bootstrap --help` and
`bootstrap <command> --help` print the full help on stdout, in the groups
`Repository`, `Mac` and `Flags`.

### The preflight

Before the preflight, each command except `version` reads the config file. An
error of the file stops the command with code 2. See
[The config file](#the-config-file).

Each command except `version` and `config` then examines these items. A failure
stops it with code 2.

- `mise` is on `PATH`.
- The folder is in a git repository. The repo side uses `MISE_PROJECT_ROOT`,
  else the git top level of the current folder. `mac init` uses the current
  folder. Each other `mac` command uses `--repo-path` or
  `MACOS_SETUP_REPO_PATH`: see [mac.md](mac.md).
- The values file is `.config/bootstrap.toml` of the repository, or the file of
  `BOOTSTRAP_VALUES`. On the Mac, it is `macos.toml` of the repository of
  `macos.toml`.

`init` outside a git repository tells the user to create one with `git init`.
On the repo side, each command except `init` needs a values file. `add`,
`remove`, `replace` and `detect` need a values file on both sides. A values file
in the old form stops each command except `init` with "run bootstrap init"
("run bootstrap mac init" on the Mac). `sync` and `check` also stop for a
question of a selected tool without an
answer, and for a group of questions without an answer that is not empty.
`check` shows a warning for each tool with `selectable = false` that the
`[tools]` table of the values file names:
`<tool> is not selectable - remove [tools.<tool>] from <file>; a tool that needs
it selects it`.
The warning does not change the exit code. `mac check` does the same.

### The change list

Each writing command (`init`, `add`, `remove`, `replace`, `detect --add`)
builds one change list: the items `add tool`, `add pack` (with its
`--exclude` members), `remove tool`, `replace tool` and the answers, and on the
Mac the packages of `mac detect --add`. `internal/core` has two functions for
it, and the TUI calls the same functions:

- `Plan` applies the items in order to a temporary copy of the values file,
  examines the answers, finds the questions without an answer, and runs a
  quiet dry run of the front-end with the new values file. It gives the new
  values file, the migrations, and each file that the sync writes, changes or
  removes. It writes nothing. A command that changes the values file also adds
  the missing tables of the required members of `base`, so the values file
  gets one write.
- `Apply` runs the script that comes before the writes, writes the values file
  and the migrations, and runs the sync. Thus each command writes one time
  and syncs one time. When the sync fails, `Apply` puts the values file back
  as it was before the command, or removes it when it did not exist. The other
  files stay as the sync left them. The one exception is code 4 of the Mac
  (the stage `init` changed): the sync did its work, so the values file stays.

- An empty change list prints `✔ Nothing to change` and exits 0. When the
  command has a reason, for example that a tool is already selected, the line
  adds it after `·`: `✔ Nothing to change · python is already selected`.
- `--dry-run` prints the list, names each dirty target, and writes nothing.
- Else the command applies the clean-target rule, asks for consent, writes the
  values file and the migrations, and then runs the sync.

`sync` has no consent step. It builds the same plan with a quiet dry run, on
the Mac on each real run, and on the repo side only when the work tree has a
change outside the values file. Then it applies the clean-target rule and
writes.

### The consent

`--yes` writes the changes. Else the command asks at a prompt in a terminal:
`? Write <n> changes? [y/N]` (`1 change` for one). A "no" writes nothing and
exits 0 with `· Nothing changed`. Without a terminal, and without `--yes`, the
command stops with code 2: `changes need --yes`.

### The clean-target rule

git is the history: the engine makes no backup and no commit. Thus a command
does not write when a target has changes that are not committed (modified,
staged or untracked). It names each such target and stops with code 2, and it
writes nothing.

- The rule applies only to a target inside a git work tree.
- `--dry-run` never stops. It names each dirty target in a warning line.
- `sync` does not apply the rule to the values file, because it writes only
  `bootstrap_commit` and the missing tables of the required members of `base`
  there. `init` applies the rule to the values file.
- `replace` needs a clean work tree, because its migrate script can change any
  file.

### `init`

`init` is the first adoption, and also the update of an adopted repository or
Mac. It takes no tool or pack names: `add tool` and `add pack` select tools.

1. A new values file starts with a header comment and `format = 2`. On the
   repo side, it also gets `environments = [ "dev" ]` and `repo_name`, the name
   of the folder, because the templates need it. It gets a table for each
   member of `base`, as `add pack base` does.
2. An existing values file gets the migration of each old form, with its
   comments kept. See [bootstrap-config.md](bootstrap-config.md) for the old
   forms. Only `init` migrates: `sync` and `check` stop with code 2 and
   "run bootstrap init" ("run bootstrap mac init" on the Mac) when they find
   an old form. A later `init` on a file with `format = 2` does not add a
   removed member of `base` again.
3. The command asks each question of a selected tool that has no answer. See
   [The questions](#the-questions).
4. It plans the sync, shows the change list, and after the consent writes the
   changes and syncs.

With a terminal, `init` will open the TUI. Until the TUI is there, `init`
takes the `--no-tui` path in each case: it selects no tool on a screen.
`init` does not run the detect scripts.

### The questions

A question is a `[questions.<key>]` table of a `tools.toml`. See
[tools.md](tools.md#the-defaults-questions-and-contributes-tables). `init` asks
the questions of each selected tool, and `add` the questions of the tools that
it adds.

- An answer is a value of `[tools.<tool>]` in the values file or in its
  overrides file. A question with an answer gets no prompt. An empty string
  `""` is an answer, and it means no answer.
- `--answer <tool>.<key>=<value>` gives the answer, and that question gets no
  prompt. An `--answer` for a question that no selected tool has is a usage
  error. When `--answer` names one question two times, the last one wins.
- `--answer <tool>.<key>=` writes `""`, so a later `init` does not ask the
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
  Without a question, the output is `✔ No questions`. `--json` gives each
  question the keys `group` (a string, `""` without a group), `note` (a string)
  and `optional` (a boolean), with the keys of before. The `detect` command runs
  in the repository, or in `$HOME` on the Mac.

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
expands the same way. The values file keeps no record of the pack. See
[packs.md](packs.md).

- `--exclude <member>` leaves out a member, with its key as the packs file
  writes it, also a nested pack (`pack:<name>`) or a tool that a nested pack
  adds. The flag needs exactly one pack name. A name that is not a member, a
  required member, and a member that another added member needs stop the
  command with code 2.
- It leaves out, with an info line, a member whose `hosts` do not name the
  host, a member with `selectable = false`, and in a git submodule a member
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

### `detect`

`detect` runs the detect script of each tool, in parallel. Each script prints a
result. See [The scripts of a tool](#the-scripts-of-a-tool).

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
  command exits 1. When each detected tool is selected, it exits 0.
- `--add` adds each offered tool through the change list, after the consent
  `? Write <n> changes? [y/N]`. `--yes` answers the consent. `--yes` and
  `--answer` need `--add`, else code 2 with the hint `add --add`. `--dry-run`
  works with or without `--add`.
- `--no` is gone: it is an unknown flag (code 2).
- When nothing is offered, the last line is `✔ No tool to add`.

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

## The cache and the clone

The cache is `${XDG_CACHE_HOME:-~/.cache}/bootstrap/`. `--cache-dir`,
`BOOTSTRAP_CACHE` or the key `cache_dir` of the config file changes the folder.
All the consumers of the machine and the Mac side share it.

| Entry                 | Use                                                  |
| --------------------- | ---------------------------------------------------- |
| `repo/`               | One shallow clone of this repository, branch `main`. |
| `lock/`               | The lock folder, with the `pid` of its holder.       |
| `lock.takeover/`      | The short lock of the removal of a dead lock.        |
| `macos-generated.tsv` | The manifest of the Mac side. See [mac.md](mac.md).  |

Each command except `version` and `config` takes the clone in `internal/core`.
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
code 5.

The sync uses the pushed `main` branch. Local changes have no effect on
consumers until you push them. To test a change before you push it, set
`BOOTSTRAP_URL` to a local clone with a commit of the change on `main` of that
clone, or set `BOOTSTRAP_CLONE` to a folder with the change. Use
`bootstrap sync --dry-run` to show a diff.

### The old clone at the cache root

A git repository at the cache root is the old layout. When its raw `origin`
normalizes to the URL, the run removes each entry of the cache root, except
`repo/`, `lock/`, `lock.takeover/` and `macos-generated.tsv`. Another `origin`
or no `origin` stops the run with code 2, and the run deletes nothing in the
cache root.

### The lock

`lock/` makes a second run wait. The run creates the folder and writes its
`pid`. A waiter removes the lock of a dead holder, and a lock without a `pid`
that is older than one minute. The removal has its own lock, `lock.takeover/`:
without it, a second waiter can remove the lock that the first waiter took
after the removal. A signal removes the lock and the copy.

| Command                                                                      | Release                      |
| ---------------------------------------------------------------------------- | ---------------------------- |
| Each repo command, `mac questions`, `mac check` and `mac sync --dry-run`     | After the copy of the clone. |
| `mac sync`, `mac init`, `mac add`, `mac remove`, `mac replace`, `mac detect` | At the end of the command.   |

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

| Variable                | Use                                                                                                                            |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `BOOTSTRAP_CLONE`       | A folder with a clone of bootstrap. The command uses it, with no lock and no fetch. The gate `code:lint:60-bootstrap` sets it. |
| `BOOTSTRAP_URL`         | The clone URL. A test uses it for a change that is not pushed.                                                                 |
| `BOOTSTRAP_CACHE`       | The cache folder. A test uses it.                                                                                              |
| `BOOTSTRAP_VERBOSE`     | `1` prints the log line of each target in place of the list. See [The output of a run](#the-output-of-a-run).                  |
| `BOOTSTRAP_BIN`         | The path of a local build of the engine. The entry tasks, the gate and the tests run it in place of the `bootstrap` of mise.   |
| `BOOTSTRAP_JOBS`        | The jobs at a time. `--jobs` sets it.                                                                                          |
| `BOOTSTRAP_VALUES`      | The values file of the repo side. The default is `.config/bootstrap.toml`.                                                     |
| `MACOS_SETUP_REPO_PATH` | The repository of `macos.toml`. See [mac.md](mac.md).                                                                          |
| `BOOTSTRAP_CONFIG`      | The config file. See [The config file](#the-config-file). The test tasks set it to a missing path.                             |

A flag sets its variable for the run: `--jobs`, `--cache-dir`, `--url`,
`--clone`, `--repo-path`, `--config` and `--verbose` (`BOOTSTRAP_VERBOSE=1`).
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
  unknown key. An empty `[mac]` table is valid. The file has no key for
  `MACOS_SETUP_REPO_PATH`: see [mac.md](mac.md).
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

`bootstrap config` prints the file and each key, and changes nothing. It needs
no git repository, no mise and no clone. The first line is `file`, the path of
the file and its source (`flag --config`, `variable BOOTSTRAP_CONFIG` or
`default`), with `(missing)` after the path of a missing file. Then each key has
one line with its value and its source: `flag --<name>`, `variable <NAME>`,
`file` or `default`. `-` is no value. The line `mac.repo_path` shows the value
of `--repo-path` or `MACOS_SETUP_REPO_PATH`. For example:

```text
file           /Users/me/.config/bootstrap/config.toml  default
jobs           4                                        flag --jobs
cache_dir      /Users/me/.cache/bootstrap               default
url            https://github.com/virajp/bootstrap.git  default
clone          -                                        default
mac.repo_path  /Users/me/Projects/macos-setup           variable MACOS_SETUP_REPO_PATH
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

| Side    | Command                                                                                                                 |
| ------- | ----------------------------------------------------------------------------------------------------------------------- |
| Both    | `git` of `PATH`, for the clone, the commit, the work tree status and the branches. `diff -u` for the diff of a dry run. |
| Both    | `mise x -- bash <script>` for each script of a tool. See [The scripts of a tool](#the-scripts-of-a-tool).               |
| `repo`  | `mise x -- which <formatter>` and `mise x -- bash -c <formatter command>`. See [The formatter](#the-formatter).         |
| `macos` | `mise bootstrap --yes` of the `mise` of `PATH`, in `$HOME`.                                                             |
| `macos` | `mac detect`: `brew leaves`, `brew list`, `mas list` and `mise ls` of `PATH`, in `$HOME`. See [`detect`](#detect).      |

## The scripts of a tool

A tool can have scripts in `scripts/` of its tool folder. The engine runs them
from the clone, and never copies them.

| Script               | When                                            | Output                                                                                                                                                              |
| -------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `derive`             | Each sync and each plan, before the render.     | One JSON object: the derived values of the tool. It can have a `mise` and a `mise_env` table, which merge into the mise config of the tool. See [mise.md](mise.md). |
| `check`              | `sync` and `check`, before a write.             | Exit 0. Or each line of stdout is a message, and the exit code is 2.                                                                                                |
| `after`              | Mac only: after `mise bootstrap` of each stage. | Nothing, or `{ "restart": "<reason>" }`, which asks for a new terminal (exit 4). See [mac.md](mac.md).                                                              |
| `detect`             | Repo side, `detect` only.                       | `{ "result": "match" \| "none" \| "mismatch", "evidence": [...], "message": "..." }`.                                                                               |
| `migrate-<old tool>` | `replace tool <old tool> <this tool>` only.     | Exit 0. It moves the repository from the old tool to this tool.                                                                                                     |

- The engine runs each script as `mise x -- bash
  <clone>/<side>/tools/<tool>/scripts/<name>`,
  in the root of the repository, or in `$HOME` on the Mac.
- Its stdin is one JSON object: `tool`, `values`, and on the repo side `repo`
  (`root`, `name` and `members`, each member with `name`, `folder` and `type`),
  or on the Mac side `home`.
- `values` of a `derive` script is the values of levels 1 and 2. `values` of a
  `check` and an `after` script is the merged values of the tool. `values` of a
  `detect` script is the `[tools.<tool>]` table of the values file.
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
- A top-level key `all` or `config` of the repo values file does not reach a
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
   result is a list of `[[item]]` tables.
2. It reads the live file. An absent file starts as an empty document.
3. For each item, it finds the item in the live file. A found item is kept or
   written again, as its kind says. An item that is not found is added, unless
   an override says `"$skip" = true`. An item with `"$mode" = "remove"` is
   removed if found.
4. It removes each item that the manifest records as added by bootstrap, and
   that no rendered spec has now (a tool went, or bootstrap dropped the item).
   The identity of the item finds it, not its value.
5. It formats, compares and writes the file with the comments of the live
   file. It records the added items in the manifest.

### The spec

The top-level keys of a spec are `format`, `comment` and `[[item]]`. Any other
key stops that target.

- `format` is `text`, `json`, `jsonc`, `yaml` or `toml`. The default is the
  `format` of the entry, else the extension of the target.
- `comment` is the comment marker of a text file, for example `#`.

| Key of an item       | Meaning                                                                                                                                                                         |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`               | `key`, `keys`, `item` or `block`. A text file has only `item`.                                                                                                                  |
| `path`               | The place in the file: a list of keys. An item of a text file has no path.                                                                                                      |
| `value`              | The bootstrap value.                                                                                                                                                            |
| `match`              | A regex for an item of a list of scalars or for a line. A non-empty list of keys for a block in a `table[]`; a key path through a list (`hooks.id`) compares the set of values. |
| `id`                 | The name of a block for an override. The live file does not have it. A block with `match` needs an `id`.                                                                        |
| `ignore`             | The keys of a found block that the engine never writes, for example `rev`.                                                                                                      |
| `type`               | Optional. Necessary for an empty list. One of `string`, `boolean`, `integer`, `float`, `<scalar>[]`, `table`, `table[]`.                                                        |
| `"$mode"`, `"$skip"` | `"$mode" = "remove"` removes the item. `"$skip" = true` does not add it.                                                                                                        |

| Kind    | What it is                                                                                                   | Found                                                                                |
| ------- | ------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `key`   | One leaf key. A list value is one value.                                                                     | The live value stays. A live value of another type is written again, with a warning. |
| `keys`  | A table of leaf keys: one `key` item for each leaf. The user can add keys to the table.                      | As `key`, for each leaf.                                                             |
| `item`  | One item of a list of scalars, or one line of a text file. Found by an equal `value`, or by a `match` regex. | The live item stays. A new item or line goes at the end, in spec order.              |
| `block` | A value that bootstrap owns: a table, a list, or one table of a `table[]`, found by its `match` keys.        | Written again from the spec, except the `ignore` keys.                               |

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
- A text file keeps each live line. New lines go at the end.
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

| Side    | Manifest                      | Target root    | Details                         |
| ------- | ----------------------------- | -------------- | ------------------------------- |
| `repo`  | `.config/repo-generated.tsv`  | The repository | [repo.md](repo.md#the-manifest) |
| `macos` | `<cache>/macos-generated.tsv` | `$HOME`        | [mac.md](mac.md#the-manifest)   |

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
4. `[tools.<tool>]` of the values file, without `env`, then `[tools.<tool>]` of
   its overrides file.

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

A value of `[tools.<name>]` of the values file depends on the tool:

- A base or dependency tool is a required member of the pack `base`, or a
  tool in the depends closure of a selected tool, a tool that a mise `depends`
  adds included. For such a tool, a list
  appends its new items, tables merge key by key, and a value whose type
  differs from the merged value stops the sync.
- For any other tool, a value replaces the merged one, and tables merge key by
  key.

The overrides file is `<values file name>.overrides.toml` next to the values
file: `.config/bootstrap.overrides.toml` or `macos.overrides.toml`. It holds
only
`[tools.<name>]` tables. Its values win over all. It is for a value that a
user keeps out of the committed values file. An answer in it counts as the
answer of a question.

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
`bootstrap_commit` of its values file to the full commit of its clone. An agent
compares it with `main` to find what changed since the last sync.

- A line that has the key before the first table is replaced. A line with the
  same value causes no write.
- Else the key goes after the header comments, when a blank line ends them, or
  else on the first line. A comment just above a key belongs to that key.
- Each other line stays. `sync` writes only this key and the tables of the
  required members of `base` to the values file. See
  [The tables of the required members](#the-tables-of-the-required-members).
- Through a symlink, the sync writes the real file, so the link stays.
- A real run prints the info line `· bootstrap_commit changed to <short
  commit>`,
  and a dry run prints `· bootstrap_commit would change to <short commit>`. With
  verbose,
  the lines are `set bootstrap_commit <commit> in <path>` and
  `would set bootstrap_commit <commit> in <path>`.
- A missing `macos.toml` is created. See [mac.md](mac.md#macostoml). On the repo
  side, a missing values file stops the sync before this step.

### The tables of the required members

`sync`, on both sides, writes an empty `[tools.<name>]` table into the values
file for each required member of `base` that has no table, so that the user
can add values to it. This also applies when the user deleted the table by
hand. It skips a member whose `hosts` do not name the host, a member with
`selectable = false`, and in a git submodule a member with `root_only`. A
dependency without a table gets no table: the sync installs it through
`depends`.

- The write keeps the comments of the file, and it comes before the files of
  the tools.
- Each table prints the info line `[tools.<name>] added to <values file>`.
- `sync --dry-run` and `check` write nothing. They print
  `[tools.<name>] would be added to <values file>` for each table. The line
  does not change the exit code.

## The output of a run

The code of the output is in `internal/report`, so both front-ends print the
same lines. `internal/cli` prints its own lines, for example the change list,
the prompts and the status lines, with the same glyphs and the same style rule.

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
clone update: `↻ bootstrap <short commit> · <cache path>`, with `$HOME` shown as
`~`. With `--clone` or `BOOTSTRAP_CLONE`, the path is the clone that the command
uses. `sync`, `mac sync` and `mac check` print `↻ bootstrap <short commit> ·
<n> tools · env <environments>` in place of that line, and the Mac side has no
env part. A command prints one clone line only: after the change list of
`init`, `add`, `remove`, `replace` and `detect --add`, the sync does not print
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

**The change list.** The change list of `init`, `add`, `remove`, `replace` and
`detect --add` has the header `Changes`. Each line is `  <glyph> <path>` with
the glyph of what the sync does to the file (`✎` changed, `✚` new, `❋` seeded,
`✖` removed), and a note column that is aligned, for example `remove claude` for
the values file. A script that runs before the writes has the glyph `▸`. A dry
run ends with `✔ Dry run: <n> changes, nothing written`. Else the prompt is
`? Write <n> changes? [y/N]`.

```text
Changes
  ✎ .config/bootstrap.toml          remove claude
  ✖ .config/mise/tasks/setup/ai
  ✎ .config/repo-generated.tsv

✔ Dry run: 3 changes, nothing written
```

When the sync of the command fails, the values file goes back as it was
before the command (see [The change list](#the-change-list)), and the info line
`· Put back <values file> as it was before the command` follows the error of
the sync. On the Mac, code 4 keeps the values file and prints no such line.

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
  usage: bootstrap sync [--dry-run] [--jobs <n>] [--verbose]
  → Run bootstrap --help for all commands
```

**The check.** `check` prints `✔ <values file> is valid` when the local rules of
the values file pass. It lists each missing table of a required member of
`base` and writes nothing. See
[The tables of the required members](#the-tables-of-the-required-members). Then
it prints `✔ check script of <tool>` for each check
script that ran. A script that fails prints `✗ check script of <tool>: <text>`
and, when the script gives a fix after `-`, the fix line. The exit codes do not
change. The local rules are the rules that the engine checks without a tool
script: the environments, the packs, the members and the overrides. When
`BOOTSTRAP_CLONE` is not set, the check scripts do not run, and a warning says
so.

**The step of `mac sync`.** `mac sync` prints `▸ Running mise bootstrap --yes`
before the command and `✔ mise bootstrap done` after a success. The output of
mise stays as it is. A failure is the error `✗ mise bootstrap failed with exit
code <rc>`.

**The style rule.** One rule turns the glyphs and the colors on and off
together: `FORCE_COLOR` (not `0`) turns them on, `NO_COLOR` turns them off, and
otherwise they need a terminal on stderr and a `TERM` that is not `dumb`. This
is the rule of `style_enabled` of `helpers`. The engine does not call
`helpers`, because the local copy can be stale and the Mac side does not have
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

The usage line of a usage error and the output of `config` have no glyph. The
verbose log lines have the prefix `bootstrap: ` with style on and with style
off.

**Verbose.** `--verbose` of the engine sets `BOOTSTRAP_VERBOSE=1`, on `init`,
`sync`, `check`, `add`, `remove`, `replace` and `detect`. `--verbose` of
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

The test tasks set the variable, so a case compares the log lines. The cases of
the default output unset it.

## The exit codes

A failure of one file does not stop the other files. A run returns the highest
code of all its failures. This section is the one place of the codes 1, 2, 3
and 5. Code 4 is only for the Mac side: see [mac.md](mac.md#stages-and-exit-4).

| Code | Meaning                                                                                                |
| ---- | ------------------------------------------------------------------------------------------------------ |
| 0    | Done, nothing to change, or a "no" at the consent prompt.                                              |
| 1    | Any other failure, for example a local file system error. `detect` without `--add` with a tool to add. |
| 2    | A usage error, or a config that breaks a rule. See the lists below.                                    |
| 3    | A name that the clone does not have. See the lists below.                                              |
| 5    | A failure of the network side or of the clone. See the lists below.                                    |

**Code 2, of the command line:**

- A bad flag or command, `add`, `remove` or `replace` without a kind word or
  without a name (the old grammar, with the hint of the new form), `init` with
  names, `remove pack` and `replace pack`, `--exclude` outside `add pack` or
  with more than one pack, and `--check` with `--dry-run` of the front-end.
- A config file that breaks a rule: see [The config file](#the-config-file).
- A missing `mise`, a folder outside a git repository, and a missing values
  file for each command except `init`.
- An old form of the values file (except for `init`): also a file without
  `format = 2`, and a `[packs.*]` table. A `format` other than `2` is not valid.
  A question of a selected tool without an answer, and a group of questions
  without an answer that is not empty, with the message "run bootstrap init"
  ("run bootstrap mac init" on the Mac).
- An `--answer` for a question that no selected tool has, an empty `--answer`
  for a question that is not optional and not in a group, a question without
  an answer with `--yes` or without a terminal, a group without an answer that
  is not empty with `--yes` or without a terminal, and ctrl+c in the field of
  a question.
- An `add tool` or `replace tool` of a named tool with `selectable = false`,
  or whose `hosts` do not name the host. See
  [Named tools that a command refuses](#named-tools-that-a-command-refuses).
- Changes without `--yes` and without a terminal: `changes need --yes`.
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
  fails, or a bad value. A `parent` that is not another tool.
- A bad `[defaults]`, `[questions]` or `[contributes]` table, and two
  contributions or a user value that conflict.
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
- `requires_found`, `requires` or `files` in a `[mise_env.<env>.tools]` entry.
- Two tools that define the same mise tool, and two selected tools with a
  `[formatter]`.
- A `root_only` tool that a `[tools.*]` entry names in a submodule.
- A bad `[member.*]` table: a name with other characters, two members with one
  name, a folder that is not in the repository, a `folder` of `.`, and a `type`
  that is not correct.
- A shared file that is a symlink or is below a symlinked folder.
- A `.config/repo-generated.tsv` that is a symlink, is below a symlinked folder
  or is not a regular file.

**Code 2, only on the Mac side:**

- A `mise` that is not at `~/.local/bin/mise`.
- A `MACOS_SETUP_REPO_PATH` (or `--repo-path`) that is not set, is not a folder
  or is not a git repository.
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
  does not have: `init` stops with
  `<pack> is not a bootstrap pack (named in <file>)`. See
  [bootstrap-config.md](bootstrap-config.md#old-forms).
- A contribution of a selected tool to a tool that is not in the clone.

**Code 5:**

- A clone or a fetch that fails.
- A `cli/VERSION` of the clone that is not `MAJOR.MINOR.PATCH`.
- A `BOOTSTRAP_CLONE` that is not usable, on the repo side.

`bootstrap check` returns the highest code of its checks.

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

1. Land the engine change on `main` with `mise run code:merge <branch>`.
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
