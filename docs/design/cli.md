# Sync engine design

This document holds the rules that both sides of the sync share: the command
line, the cache and the clone, the version check, the engine variables, the
tool calls, the scripts of a tool, the templates, the write of one target,
shared files, the header, the manifest and the stale rule, the values, the
`bootstrap_commit` key, the output, the exit codes and the release of the
engine. It is the one place of these rules.

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

| Path                 | Job                                                                                                                                                |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `VERSION`            | The minimum version of the engine that a clone of `main` needs. See [The version check](#the-version-check).                                       |
| `go.mod`             | The module.                                                                                                                                        |
| `cmd/bootstrap`      | The command `bootstrap`. It calls `internal/cli`. `version` prints `main.version`, which the release build sets. A local build has `dev`.          |
| `internal/cli`       | The command line: the preflight, the clone and its lock, the version check, the change list, the clean-target rule, the consent and each command.  |
| `internal/oldform`   | The old forms of a values file, and their migration. It is the one place of the old names.                                                         |
| `internal/questions` | The questions of the selected tools, their proposals and their answers.                                                                            |
| `internal/scripts`   | The run of the scripts of a tool. See [The scripts of a tool](#the-scripts-of-a-tool).                                                             |
| `internal/fail`      | The typed errors of the exit codes, and a collector where the highest code wins. See [The exit codes](#the-exit-codes).                            |
| `internal/report`    | The output of a run. See [The output of a run](#the-output-of-a-run).                                                                              |
| `internal/source`    | The read of the clone: the packs file, the `tools.toml` of each tool, the entries and the files of a tool folder, and the checks of these files.   |
| `internal/values`    | The values files and the overrides files of both sides, and the values merge. See [The values of a tool](#the-values-of-a-tool).                   |
| `internal/selection` | The pack expansion, the checks of the packs, the depth-first walk of the `depends`, and the stages of the Mac.                                     |
| `internal/tmpl`      | `text/template` with the delimiters `<%` and `%>`, and our own copy of the gomplate functions that the templates use. It does not import gomplate. |
| `internal/format`    | The parse check, and the merge of a shared file: JSON, JSONC, YAML, TOML and text. See [Shared files](#shared-files).                              |
| `internal/header`    | The header of a synced file and of a shared file, and its detection. See [The header](#the-header).                                                |
| `internal/target`    | The write of one target, the modes, and `bootstrap_commit`. See [The write of one target](#the-write-of-one-target).                               |
| `internal/manifest`  | The manifest and the stale rule, in the link modes `repo` and `macos`. See [The manifest](#the-manifest-and-the-stale-rule).                       |
| `internal/misegen`   | The mise config generator of both sides. See [mise.md](mise.md).                                                                                   |
| `internal/repo`      | The repo front-end, the sync of `bootstrap sync`. See [repo.md](repo.md).                                                                          |
| `internal/macos`     | The Mac front-end, the sync of `bootstrap mac sync`. See [mac.md](mac.md).                                                                         |

The task `test:go` of this repository runs `go vet`, a `gofmt` check and
`go test` in the module. Some unit tests compare the output of a case with a
golden file in `testdata/`. `UPDATE_GOLDEN=1 go test ./internal/<package>`
writes the golden files of `selection`, `misegen`, `source` and `tmpl` again.
The golden files of `internal/tmpl` hold the output of the gomplate CLI for the
template functions and the inline templates.

## The command line

```text
bootstrap <command> [flags]
bootstrap mac <command> [flags]
```

The group `mac` names the Mac side. Thus a folder with both `macos.toml` and
`.config/bootstrap.toml` has no doubt. Each command without `mac` is for the
repository of the current folder.

| Command                | Job                                                                                                                                                  |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `init [tool…]`         | Selects `base` and the named tools, migrates an old form, asks each question without an answer, then syncs. See [`init`](#init).                     |
| `sync`                 | The sync of the side. `setup:configs` and `updateall:configs` run it.                                                                                |
| `check`                | Examines the values file and runs the check scripts. On the Mac, a dry run of the sync.                                                              |
| `questions [tool…]`    | Lists the questions, their proposals and their answers. Without names, the questions of the selected tools. `--json` prints one JSON array.          |
| `add <tool…>`          | Selects the tools and their dependencies, asks their questions, then syncs.                                                                          |
| `remove <tool…>`       | Removes the tools, and each dependency that the user named and that no other tool needs, then syncs. Each seed of a removed tool stays, with a line. |
| `replace <old> <new>`  | Runs the migrate script of `<new>` for `<old>`, swaps the tools, then syncs. See [`replace`](#replace).                                              |
| `detect`               | Runs the detect scripts, and offers to add each tool that matches. See [`detect`](#detect).                                                          |
| `list`, `ls`           | Prints the selected tools as a tree. See [`list`](#list). It writes nothing.                                                                         |
| `version`, `--version` | Prints the version. There is no `mac version`.                                                                                                       |

| Flag                            | Commands                                              | Effect                                                                 |
| ------------------------------- | ----------------------------------------------------- | ---------------------------------------------------------------------- |
| `--dry-run`                     | `init`, `sync`, `add`, `remove`, `replace`, `detect`  | Prints the change list, and writes nothing.                            |
| `--yes`, `-y`                   | `init`, `add`, `remove`, `replace`, `detect`          | Writes the changes without a prompt. `detect`: adds each matched tool. |
| `--no`                          | `detect`                                              | Adds nothing. A matched tool that is not selected exits 1.             |
| `--answer <tool>.<key>=<value>` | `init`, `add`, `replace`, `detect`                    | The answer of a question. The flag can repeat.                         |
| `--json`                        | `questions`                                           | JSON output, for an AI agent.                                          |
| `--no-bootstrap`                | each `mac` command except `questions`                 | Skips each `mise bootstrap` step.                                      |
| `--jobs <n>`                    | each command except `questions`, `list` and `version` | The jobs at a time. The default is 2 × the CPU count.                  |

A flag that a command does not take, an unknown command, and `add` or `remove`
without a tool name are usage errors (code 2).

### The preflight

Each command first examines these items. A failure stops it with code 2.

- `mise` is on `PATH`.
- The folder is in a git repository. The repo side uses `MISE_PROJECT_ROOT`,
  else the git top level of the current folder. `mac init` uses the current
  folder. Each other `mac` command uses `MACOS_SETUP_REPO_PATH`: see
  [mac.md](mac.md).
- The values file is `.config/bootstrap.toml` of the repository, or the file of
  `BOOTSTRAP_VALUES`. On the Mac, it is `macos.toml` of the repository of
  `macos.toml`.

`init` outside a git repository tells the user to create one with `git init`.
On the repo side, each command except `init` needs a values file. `add`,
`remove`, `replace` and `detect` need a values file on both sides. A values file
in the old form stops each command except `init` with "run bootstrap init"
("run bootstrap mac init" on the Mac). `sync` and `check` also stop for a
question of a selected tool without an
answer.

### The change list

Each writing command (`init`, `add`, `remove`, `replace`, `detect`) builds the
full change list in memory first: the writes of the values file, the
migrations, and each file that the sync writes, changes or removes. The sync
part comes from a quiet dry run of the front-end with the new values file.

- An empty change list prints `Nothing to change` and exits 0.
- `--dry-run` prints the list, names each dirty target, and writes nothing.
- Else the command applies the clean-target rule, asks for consent, writes the
  values file and the migrations, and then runs the sync.

`sync` has no consent step. It builds the same plan with a quiet dry run, on
the Mac on each real run, and on the repo side only when the work tree has a
change outside the values file. Then it applies the clean-target rule and
writes.

### The consent

`--yes` writes the changes. Else the command asks at a prompt in a terminal.
A "no" writes nothing and exits 0 with `Nothing changed`. Without a terminal,
and without `--yes`, the command stops with code 2: `changes need --yes`.

### The clean-target rule

git is the history: the engine makes no backup and no commit. Thus a command
does not write when a target has changes that are not committed (modified,
staged or untracked). It names each such target and stops with code 2, and it
writes nothing.

- The rule applies only to a target inside a git work tree.
- `--dry-run` never stops. It names each dirty target in an info line.
- `sync` does not apply the rule to the values file, because it writes only
  `bootstrap_commit` there. `init` applies the rule to the values file.
- `replace` needs a clean work tree, because its migrate script can change any
  file.

### `init`

`init` is the first adoption, and also the update of an adopted repository or
Mac.

1. A new values file starts with a header comment. On the repo side, it also
   gets `environments = [ "dev" ]` and `repo_name`, the name of the folder,
   because the templates need it.
2. An existing values file gets the migration of each old form, with its
   comments kept. See [bootstrap-config.md](bootstrap-config.md) for the old
   forms. Only `init` migrates: `sync` and `check` stop with code 2 and
   "run bootstrap init" ("run bootstrap mac init" on the Mac) when they find
   an old form.
3. Each named tool that is not selected gets an empty `[tools.<name>]` table.
4. The command asks each question of a selected tool that has no answer. See
   [The questions](#the-questions).
5. It plans the sync, shows the change list, and after the consent writes the
   changes and syncs.

`init` does not run the detect scripts.

### The questions

A question is a `[questions.<key>]` table of a `tools.toml`. See
[tools.md](tools.md#the-defaults-questions-and-contributes-tables). `init` asks
the questions of each selected tool, and `add` the questions of the tools that
it adds.

- An answer is a value of `[tools.<tool>]` in the values file or in its
  overrides file. A question with an answer gets no prompt.
- `--answer <tool>.<key>=<value>` gives the answer, and that question gets no
  prompt. An `--answer` for a question that no selected tool has is a usage
  error.
- The prompt shows the proposal: the output of the `detect` command of the
  question, else its `example`. An empty input takes the proposal. Without a
  proposal, an empty input is not accepted.
- With `--yes`, or without a terminal, the command does not accept a proposal:
  a question without an `--answer` stops the command with code 2.
- `questions` lists each question with its proposal and its source (`detect` or
  `example`), and its answer. The `detect` command runs in the repository, or
  in `$HOME` on the Mac.

### `add` and `remove`

`add` writes an empty `[tools.<name>]` table for each named tool and each
dependency that is not selected yet. A tool that is already selected, also
through a pack, gets an info line and no table.

`remove` acts only on `[tools.<name>]` tables. It refuses each of these with
code 2:

- A tool that a pack marks `required = true`.
- A tool that only a pack selects. The message names the pack.
- A tool that is not selected.
- A tool that another selected tool needs.

It also removes each dependency of a removed tool that has a
`[tools.<name>]` table and that no other selected tool needs. An empty
`[tools]` table goes too. An unknown tool name is code 3 for `add` and
`remove`.

### `replace`

`replace <old> <new>` works only between two tools with the same `parent`, for
example `pnpm` and `aube` with `parent = "node"`. The new tool must have the
script `scripts/migrate-<old>`. `replace` refuses a new tool that is already
selected, and an old tool that another tool needs.

The migrate script runs after the consent and before the writes. Its input has
`tool` = the new tool and `values` = the old `[tools.<old>]` table. The new
table gets only the `env` of the old table.

### `list`

`list` prints a tree. It writes nothing. Each root is `<pack> (pack)` for a pack
that selects a tool, `base` first and then the other packs by name. The last
root is `tools (selected directly)` for the tools of `[tools.<name>]`, and it
is absent when there are none.

Below a root are its tools, sorted by name. Below a tool are its dependencies,
sorted by name, and a dependency of a dependency goes one level lower. A tool
with more than one parent shows below each parent. A required tool has the
marker `[required]`, and all markers are in one column.

```text
base (pack)
├── git         [required]
├── pre-commit  [required]
│   └── python
└── vscode
tools (selected directly)
└── claude
```

### `detect`

`detect` runs the detect script of each tool, in parallel. Each script prints a
result. See [The scripts of a tool](#the-scripts-of-a-tool).

- `match` of a tool that is not selected: the command offers to add it. A
  match whose evidence is only the evidence of a selected tool that matches is
  not offered.
- `mismatch` of a selected tool is an error.
- `none` of a selected tool is an info line. The tool stays selected.
- `--yes` adds each offered tool, and `--no` adds nothing. With `--no`, an
  offered tool exits 1. Without both flags and without a terminal, the command
  stops with code 2. `--yes` with `--no` is a usage error.

## The cache and the clone

The cache is `${XDG_CACHE_HOME:-~/.cache}/bootstrap/`. `BOOTSTRAP_CACHE`
changes the folder. All the consumers of the machine and the Mac side share it.

| Entry                 | Use                                                  |
| --------------------- | ---------------------------------------------------- |
| `repo/`               | One shallow clone of this repository, branch `main`. |
| `lock/`               | The lock folder, with the `pid` of its holder.       |
| `lock.takeover/`      | The short lock of the removal of a dead lock.        |
| `macos-generated.tsv` | The manifest of the Mac side. See [mac.md](mac.md).  |

Each command except `version` takes the clone in `internal/cli`. With
`BOOTSTRAP_CLONE` set, the command uses that folder as it is: it takes no lock
and makes no fetch. Else it takes the lock, fetches, and copies the clone to a
temporary folder. The sync reads only this copy. Thus a change of the clone
applies to a whole run, and the next run can update the cache.

The URL is `https://github.com/virajp/bootstrap.git`. `BOOTSTRAP_URL` changes
it. The fetch does these steps:

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

## The scripts of a tool

A tool can have scripts in `scripts/` of its tool folder. The engine runs them
from the clone, and never copies them.

| Script               | When                                            | Output                                                                                                                                                              |
| -------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `derive`             | Each sync and each plan, before the render.     | One JSON object: the derived values of the tool. It can have a `mise` and a `mise_env` table, which merge into the mise config of the tool. See [mise.md](mise.md). |
| `check`              | `sync` and `check`, before a write.             | Exit 0. Or each line of stdout is a message, and the exit code is 2.                                                                                                |
| `after`              | Mac only: after `mise bootstrap` of each stage. | Nothing, or `{ "restart": "<reason>" }`, which asks for a new terminal (exit 4). See [mac.md](mac.md).                                                              |
| `detect`             | `detect` only.                                  | `{ "result": "match" \| "none" \| "mismatch", "evidence": [...], "message": "..." }`.                                                                               |
| `migrate-<old tool>` | `replace <old tool> <this tool>` only.          | Exit 0. It moves the repository from the old tool to this tool.                                                                                                     |

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
  does the precise detection.

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

- A base or dependency tool is a tool that the pack `base` adds, also through a
  nested `default = true` pack, or a tool in the depends closure of a selected
  tool, a tool that a mise `depends` adds included. For such a tool, a list
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
- Each other line stays. This is the only write of `sync` to the values file.
- Through a symlink, the sync writes the real file, so the link stays.
- A real run prints the info line `Set bootstrap_commit to <short commit>`, and
  a dry run prints `Would set bootstrap_commit to <short commit>`. With verbose,
  the lines are `set bootstrap_commit <commit> in <path>` and
  `would set bootstrap_commit <commit> in <path>`.
- A missing `macos.toml` is created. See [mac.md](mac.md#macostoml). On the repo
  side, a missing values file stops the sync before this step.

## The output of a run

The code of the output is in `internal/report`, so both front-ends print the
same lines. `internal/cli` prints its own lines, for example the change list,
the prompts and the info lines, with the same style rule.

**The outcome of a target.** The outcome of a target (a kind, a path, and an
optional text and log line) replaces each per-file log line. It increments the
count of the kind and prints the correct line.

| Kind        | Lines of the log                                      | List                | Dry-run word   | Table row      |
| ----------- | ----------------------------------------------------- | ------------------- | -------------- | -------------- |
| `changed`   | `updated`, `chmod`, `wrote` of a manifest that exists | `✎ changed`, yellow | `would change` | `Changed`      |
| `new`       | `updated`, `wrote` of a target that does not exist    | `✚ new`, green      | `would add`    | `New`          |
| `seeded`    | `updated` of a seed                                   | `❋ seeded`, blue    | `would seed`   | `Seeded`       |
| `removed`   | `removed`                                             | `✖ removed`, red    | `would remove` | `Removed`      |
| `failed`    | Each failure message of a target                      | `✗ failed`, red     | `failed`       | `Failed`       |
| `unchanged` | `unchanged`                                           | None                | None           | `= Unchanged`  |
| `kept`      | `kept seed`                                           | None                | None           | `◦ Kept seeds` |
| `skipped`   | `skip seed`, `skip merge`, `skip file`, `skip env`    | None                | None           | `⊘ Skipped`    |

- **A failed target:** It counts once, and it has one list line, from its first
  failure message: `failed <path>: <text>`. `<path>` is the target, and the text
  does not repeat it. A target that fails with no message gets the text
  `cannot write the target`. A later line for the same target, for example
  `kept existing`, prints only with verbose.
- **No count:** A failure of the mise config of a tool or of the manifest file
  has no count, and its message stays. The line `FAILED: <list>` stays, before
  the table. The other log messages stay too.

**The table and the last line.** The summary of a run prints them once, as the
last output of the run: after the stale rule, the manifest write and
`bootstrap_commit`. A run prints it after the file loop, also on a failure. A
run that stops before the file loop has no table.

- The table always has the rows `Changed`, `New`, `Seeded`, `Removed`,
  `Unchanged`, `Kept seeds`, `Skipped`, `Failed` and `Total`, in that order.
  `Total` is the sum of the rows.
- A row with a zero count is dim. Another row takes the color of its kind in the
  list. `Unchanged`, `Kept seeds`, `Skipped` and `Total` have no color.
- `n` is the sum of `Changed`, `New`, `Seeded` and `Removed`.

| Run              | Last line                                                                   |
| ---------------- | --------------------------------------------------------------------------- |
| Real run, code 0 | `✅ Synced: <n> files changed`, or `✅ Synced: no file changed`             |
| Dry run, code 0  | `✅ Dry run: <n> files would change`, or `✅ Dry run: no file would change` |
| Another code     | A red `❌ The sync failed with code <rc>: <failed count> files failed`      |

The last line uses `1 file` for one file. With no failed file, the line with
another code stops after `<rc>`.

**The style rule.** One rule turns the glyphs and the colors on and off
together: `FORCE_COLOR` (not `0`) turns them on, `NO_COLOR` turns them off, and
otherwise they need a terminal on stderr and a `TERM` that is not `dumb`. This
is the rule of `style_enabled` of `helpers`. The engine does not call
`helpers`, because the local copy can be stale and the Mac side does not have
it.

| Line      | Style on                       | Style off                         |
| --------- | ------------------------------ | --------------------------------- |
| List      | `<glyph> <word> <path>`, color | `<word> <path>`                   |
| Info      | `ℹ️ <text>`, no color          | `[INFO] <text>`                   |
| Last line | `✅ <text>` or `❌ <text>`     | `[OK] <text>` or `[ERROR] <text>` |
| Table     | Box characters                 | ASCII borders (`+`, `-`, `\|`)    |

These lines have no prefix. A log line of a front-end, on the two sides, has
the prefix `bootstrap: ` with style off. A line of `internal/cli` shows the
info glyph or the error glyph, or the same prefix with style off.

**Verbose.** `--verbose` of `setup:configs` and `updateall:configs` sets
`BOOTSTRAP_VERBOSE=1`. A user can also set the variable, and only the value `1`
turns it on. Each entry task consumes the flag and does not pass it to the
engine. Each target then prints its log line in place of its list line, for
example `unchanged  <path>`, `updated    <path>`, `kept seed  <path>`,
`skip env   <path>` or `would update <path>`. The table and the last line stay.
These lines print only with verbose:

- `using bootstrap at <commit>`, `tools:` and, on the repo side,
  `environments:`. By default, they are one info line. See
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

| Code | Meaning                                                                                     |
| ---- | ------------------------------------------------------------------------------------------- |
| 0    | Done, nothing to change, or a "no" at the consent prompt.                                   |
| 1    | Any other failure, for example a local file system error. `detect --no` with a tool to add. |
| 2    | A usage error, or a config that breaks a rule. See the lists below.                         |
| 3    | A name that the clone does not have. See the lists below.                                   |
| 5    | A failure of the network side or of the clone. See the lists below.                         |

**Code 2, of the command line:**

- A bad flag or command, `add` or `remove` without a tool, `--yes` with `--no`,
  and `--check` with `--dry-run` of the front-end.
- A missing `mise`, a folder outside a git repository, and a missing values
  file for each command except `init`.
- An old form of the values file (except for `init`), and a question of a
  selected tool without an answer, with the message "run bootstrap init"
  ("run bootstrap mac init" on the Mac).
- An `--answer` for a question that no selected tool has, and a question without
  an answer with `--yes` or without a terminal.
- Changes without `--yes` and without a terminal: `changes need --yes`.
- A dirty target: see [The clean-target rule](#the-clean-target-rule).
- A binary that is older than `VERSION` of the clone.
- A `remove` of a required tool, of a tool that only a pack selects, of a tool
  that is not selected, or of a tool that another tool needs.
- A `replace` between tools without the same `parent`, without a migrate
  script, of a new tool that is already selected, or of an old tool that
  another tool needs.
- A `detect` without `--yes` or `--no` and without a terminal, and a
  `mismatch` of a selected tool.

**Code 2, on both sides:**

- A `repo/` of the cache that is not a clone and not empty, and a repository at
  the cache root with another `origin` or no `origin`.
- A `tools.toml` that cannot be read or that breaks a rule, also an old key.
- A `[pack.*]` table that breaks a rule: an old key, a member that is not a
  table or is not on one line, another key in a member, a `default` that is not
  a boolean, a pack name that is also a tool folder, and a pack loop.
- A `required` member of a pack that is not a `default = true` member of
  `base`.
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

- A values file in the old list format (`tools = [...]`, `packs = [...]`), and
  another key than `env` in a `[packs.<name>]` entry.
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
- A `MACOS_SETUP_REPO_PATH` that is not set, is not a folder or is not a git
  repository.
- A `BOOTSTRAP_CLONE` without `cli/` or `macos/tools/`.
- A `macos.toml` that breaks a rule, also a `[overrides.mise_env.*]` table.
- A nested run of `updateall:configs`.
- A key of a Mac `tools.toml` that is for repo tools only, a `requires_found` of
  a `[files]` table, and a `[mise_env.*]`, an `envs` or a `requires_found` of a
  mise tool.
- Two selected tools that ship one target, and a tool that ships a file below
  `.config/mise/conf.d/`.

**Code 3:**

- A tool or a pack that is not in the clone, also a member of a pack that is not
  a tool folder or a pack, and a tool in a `depends`.
- A tool name of `add`, `remove`, `replace` or `questions` that is not in the
  clone.
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
