# Tools design

A tool is the building block of `bootstrap`. A tool folder holds a
`tools.toml`, the config files of the tool in `files/`, its task files in
`tasks/` and its scripts in `scripts/`. The `tools.toml` gives the target path
of each file. This document holds the `tools.toml` schema of both sides, the
`[files]`, `[tasks]` and `[formatter]` keys, the scripts, the folder rules and
the rule for the header of a source file.

See also: [packs.md](packs.md), [mise.md](mise.md) for `[mise.*]`,
[sync-engine.md](sync-engine.md) for shared files, the scripts and the values,
[repo.md](repo.md) and [mac.md](mac.md). The exit code of each rule that stops
the sync is in [sync-engine.md](sync-engine.md#the-exit-codes).

## The two sides

| Side    | Tools root     | A file goes to             | Reads the tools of   |
| ------- | -------------- | -------------------------- | -------------------- |
| `repo`  | `repo/tools/`  | The path in the repository | `bootstrap sync`     |
| `macos` | `macos/tools/` | The path below `$HOME`     | `bootstrap mac sync` |

A tool folder has four parts:

```text
repo/tools/<tool>/
├── tools.toml   necessary
├── files/       optional: the config files and merge specs, one entry each in [files]
├── tasks/       optional: the task files, one entry each in [tasks]
└── scripts/     optional: derive, check, after, detect, migrate-<tool>
```

Each file below `files/` and `tasks/` has an entry. The entry gives its target.
For example, the entry `[tasks."code/merge/develop"]` with
`source = "code-merge-develop"` copies `repo/tools/git/tasks/code-merge-develop`
to `.config/mise/tasks/code/merge/develop` in a consumer. A script is not
copied: the engine runs it from the clone. See
[sync-engine.md](sync-engine.md#the-scripts-of-a-tool).

## The `tools.toml` file

Each tool folder has a `tools.toml`. A Mac tool folder without it stops the
sync.

| Key                    | `repo`    | `macos`  | Meaning                                                                                                                          |
| ---------------------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `default_envs`         | Necessary | Refused  | A list: the environments of the tool when the values file gives it no `env`, for example `["dev"]`. It must include `dev`.       |
| `depends`              | Optional  | Optional | The tools that the sync adds with this one, transitively. On the repo side, each gets the environments of this tool.             |
| `root_only`            | Optional  | Refused  | `true`: the tool is not for a git submodule. See [repo.md](repo.md#root_only).                                                   |
| `description`          | Optional  | Optional | A string: one line that tells the adopting AI what the tool is.                                                                  |
| `parent`               | Optional  | Optional | The name of another tool, for example `parent = "node"` in `pnpm` and `aube`. Tools with the same parent can replace each other. |
| `stage`                | Not read  | Optional | `"init"` puts the tool in the init stage. See [mac.md](mac.md#stages-and-exit-4).                                                |
| `[files."<target>"]`   | Optional  | Optional | The source, the mode and the condition of one config file. See [The `[files]` table](#the-files-table).                          |
| `[tasks."<path>"]`     | Optional  | Optional | The source, the mode and the condition of one task file. See [The `[tasks]` table](#the-tasks-table).                            |
| `[defaults]`           | Optional  | Optional | The values of the tool. See [below](#the-defaults-questions-and-contributes-tables).                                             |
| `[questions.<key>]`    | Optional  | Optional | The question for one value.                                                                                                      |
| `[contributes.<tool>]` | Optional  | Optional | Values for another selected tool, for example the hooks of `pre-commit` or the extensions of `vscode`.                           |
| `[formatter]`          | Optional  | Refused  | The formatter of the sync. See [The `[formatter]` table](#the-formatter-table).                                                  |
| `[mise.*]`             | Optional  | Optional | The mise config of the tool. See [mise.md](mise.md#the-mise-sections).                                                           |
| `[mise_env.<env>.*]`   | Optional  | Refused  | More mise config for one environment. See [mise.md](mise.md#the-mise-sections).                                                  |

The sync examines the `tools.toml` of each tool of the clone, selected or not.

On the Mac side, "refused" means that the key stops the sync, with the message
`is for
repo tools only`. Any other top-level key also stops the sync, for example a
typo such
as `depend`. `depends` must be a list of tool names. The only value of `stage`
is
`"init"`.

`default_envs` and the `envs` of a mise tool can name only `dev` and `ci`, the
scope of
`bootstrap`. Any other name stops the sync. A name of `default_envs`
that
the values file selects, and that a mise tool of the tool does not allow in its
`envs`,
stops the sync. See
[mise.md](mise.md#the-environments-of-a-mise-tool). A
`default_envs` without `dev` stops the sync, also for a tool that
the values
file does not select.

### The folder rules

- A repo tool folder must not be `zz-repo`, and its name must not sort after
  `zz-repo`
  in byte order. A Mac tool folder has the same rule for `zz-macos`. If not, the
  sync
  stops, because the folder of the owner in `conf.d/` must load
  last. See
  [mise.md](mise.md#zz-repo-and-zz-macos).
- A Mac tool folder name has no whitespace and does not start with a dot.
- A folder whose name starts with `_` is a common tool, for example
  `_security_base`. It
  is not a member of a pack. See [packs.md](packs.md#common-tools).
- Two tools must not ship the same path. The repo side does not examine this
  rule, so
  examine it yourself when you add a file. On the Mac side, two selected tools
  that ship
  one target stop the sync.
- Each regular file of the tool folder, other than `tools.toml`, is below
  `files/`, `tasks/` or `scripts/`. A file outside these folders is the old
  form of a tool folder, and it stops the sync.
- A file of `scripts/` must be `derive`, `check`, `after`, `detect` or
  `migrate-<tool>`. Any other file there stops the sync.
- Each file below `files/` and `tasks/` has an entry: the entry names it in
  `source`, or a folder entry names a folder above it. A file without an entry
  stops the sync.
- A tool does not ship mise files, and a tool that ships a file that the sync
  also
  generates stops the sync. See [mise.md](mise.md#the-mise-sections)
  for the
  exceptions. A Mac tool that ships a file below `.config/mise/conf.d/` also
  stops the
  sync.

A tool name is the name of its folder. A pack name must not be a tool folder.

### The old keys

These keys are gone. Each stops the sync with a message that names the new
key: the top-level `seed` key, the old table of seed conditions
`[seed_requires]`, the `requires` and the `files` of a mise entry, the
`extensions` key (now `[contributes.vscode] extensions`) and `[[precommit]]`
(now `[[contributes.pre-commit.repos]]` with an `id`). On the Mac side,
`stage = "shell"` also stops the sync and names `stage = "init"`.

## The `[files]` table

A `[files."<target>"]` table of a `tools.toml` is the entry of one config file
of the tool. The key is the target: the path in the repository, or the path
below `$HOME` on the Mac side, with no `~/`. A target has no `..` part and does
not start with `/`.

| Key              | Side | Meaning                                                                                                                                                                         |
| ---------------- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `source`         | Both | Necessary. The path below `files/`, with no `..` part. A folder copies each regular file below it, to the same path below the target. For `merge`, a spec `<name>.merge.toml`.  |
| `mode`           | Both | `copy` (the default): copied on each run. `template`: rendered on each run. `seed`: written only when it is missing. `merge`: a shared file, merged item by item from its spec. |
| `render`         | Both | Only with `seed`. `true` renders the seed at its first write.                                                                                                                   |
| `requires`       | Both | A list of exact paths. Each path must exist. It is relative to the root of the repository, or to `$HOME` on the Mac side. It has no `..` part.                                  |
| `requires_found` | Repo | A list of file names. See [repo.md](repo.md#requires_found). On the Mac side, it stops the sync.                                                                                |
| `format`         | Both | `text`, `json`, `jsonc`, `yaml` or `toml`: the type of the file, not the extension. It selects the parse check and the merge. The default comes from the extension.             |
| `when`           | Both | A value path `<tool>.<path>` in the values of all tools, with an optional `!` at the start. The entry applies only when the value is `true` (with `!`: not `true`).             |
| `perm`           | Both | The mode of the file, an octal string, for example `"600"`. The default is the mode of the source.                                                                              |
| `dir_perm`       | Both | The mode of the folder of the target, an octal string, for example `"700"` for each file below `~/.ssh`.                                                                        |

The `.tmpl` suffix of a source is a name convention for a template. The `mode`
key, not the suffix, makes a file a template. A `copy` or `template` file is an
owned file: the sync writes it on each run, and it gets a row in the manifest.

An entry whose `when` value is not `true` is not shipped, and the stale rule
removes a synced copy. For example, the merge tasks of `git` have
`when = "git.branch_pair"`, a value of the derive script of `git`. See
[repo.md](repo.md#the-branch-model-and-the-merge-tasks).

Any other key stops the sync. `requires` and `requires_found` must
be
non-empty lists.

When a `requires` or `requires_found` entry is not there, the sync does not
write the
target. It counts a seed that exists as kept, so that seed is in `Kept seeds`.
It counts each other target as skipped. A generated target gets no row in
the manifest, so
after
a run, the stale rule removes a listed copy with the bootstrap header. A copy
without
the header, a seed and a merge target stay.

- A file that the consumer owns after the first write is a seed. A config file
  that a
  project tunes, for example the rules of a linter (`.config/swiftlint.yml`,
  `.config/detekt.yml`), is a seed.
- A seed that applies only when the consumer has a specific file also gets
  `requires`,
  for example `requires = ["package.json"]`.
- A shared file is in [sync-engine.md](sync-engine.md#shared-files). The
  `pre-commit`, `dprint`, `git` and `vscode` tools of `repo/tools/` and the
  `claude-cli`, `gem`, `gh` and `superfile` tools of `macos/tools/` have one.
- `test:macos-tools` stops when a `[files]` table gives `mode = "seed"` to a
  JSON, YAML
  or TOML target of a Mac tool.

## The `[tasks]` table

A `[tasks."<path>"]` table is the entry of one task file of the tool. The key is
the path below `.config/mise/tasks/` in the target, for example
`code/format/40-ruff`. `#MISE` and `#USAGE` stay in the task file.

| Key              | Side | Meaning                                                                                               |
| ---------------- | ---- | ----------------------------------------------------------------------------------------------------- |
| `source`         | Both | Necessary. The path below `tasks/`, with no `..` part. The tools use a flat name: `code-format-ruff`. |
| `mode`           | Both | `copy` (the default) or `template`.                                                                   |
| `requires`       | Both | As in `[files]`.                                                                                      |
| `requires_found` | Repo | As in `[files]`.                                                                                      |
| `when`           | Both | As in `[files]`.                                                                                      |
| `library`        | Both | `true` for a library below `_scripts/`, and only there. A library is not a task.                      |

The rules of the task files:

- The `mise` tool ships each orchestrator `_default`, for example
  `code/format/_default`. Each other tool ships its own task files below it.
- An orchestrator runs the leaf tasks of its group with `own_tasks`, which
  returns the names sorted. Thus a leaf task of the groups `code:format`,
  `code:lint`, `code:sec`, `setup:deps` and `setup:secrets` has a number
  prefix for its run order: 10, 20, 30 and so on, for example
  `code/lint/70-ruff`. A task of the repository without a prefix runs after
  them.
- A task that a user or another task calls by its name has no prefix, for
  example `code:merge:develop` and `setup:dprint`.

## The `[formatter]` table

A repo tool can have one `[formatter]` table: the formatter that the sync runs
on each rendered file before the compare. See
[sync-engine.md](sync-engine.md#the-formatter).

| Key       | Meaning                                                                                                              |
| --------- | -------------------------------------------------------------------------------------------------------------------- |
| `command` | A shell command. It runs in the stage folder, and gets the paths of the files on stdin, one for each line.           |
| `configs` | A list of the names of the config files of the formatter. The stage gets a copy of each one that the repository has. |

The table has exactly these two keys. At most one selected tool can have a
`[formatter]`. The `dprint` tool has it:

```toml
[formatter]
command = "dprint fmt --no-gitignore --allow-no-files --stdin-files"
configs = [ "dprint.json", "dprint.jsonc", ".config/dprint.jsonc" ]
```

## The header of a source file

Each file that is neither a seed nor a merge source has this header in its
source, in the comment syntax of the file. `<path>` is the path of the source in
the tool folder, for example `files/.gitignore` or `tasks/code-format.tmpl`:

```text
# Synced from bootstrap: repo/tools/<tool>/<path>. Do not edit this file.
# The next sync writes it again, and each change in this file is lost.
```

A file of a Mac tool has `macos/tools/<tool>/<path>` in the header.

| Kind of file   | Place of the header                            |
| -------------- | ---------------------------------------------- |
| A script       | After the shebang.                             |
| YAML           | After `---` when the file has it.              |
| TOML           | After a `# :schema` line when the file has it. |
| JSONC          | Inside `{`.                                    |
| Markdown       | As an HTML comment.                            |
| Any other file | On the first line.                             |

The sync does not add this header to a shipped file. The generator of the mise
config
writes the same two lines in each generated mise file. See
[mise.md](mise.md#the-generated-files). The detection of the header is in
[sync-engine.md](sync-engine.md#the-header).

A seed and a merge spec have no header. The sync writes the shared header into
a shared file of a format with comments. See
[sync-engine.md](sync-engine.md#the-header).

## The `[defaults]`, `[questions]` and `[contributes]` tables

A `tools.toml` can have three optional tables for the values of a tool. A
template reads
the values as `.config` (the values of its tool) and `.all` (the values of each
selected
tool, by tool name). The `pre-commit`, `dprint` and `vscode` tools of
`repo/tools/` use `[defaults]`, and many repo tools use `[contributes]` for the
hooks of `pre-commit`, the blocks of `dprint` and the extensions of `vscode`.
Mac tools such as
`fish`, `git`, `ssh`, `gh` and `macos-defaults` use `[defaults]` or
`[questions]`. The
Mac tools `1password` and `orbstack` use `[contributes]`. The merge of these
values is
in [sync-engine.md](sync-engine.md#the-values-of-a-tool).

```toml
[defaults]
registry = "https://registry.example.com"

[questions.team]
detect = "git config user.email"
example = "platform"
question = "Which team owns this repository?"

[contributes.vscode]
settings = [ "editor.formatOnSave" ]
```

| Table                  | Keys                                                                              | Meaning                           |
| ---------------------- | --------------------------------------------------------------------------------- | --------------------------------- |
| `[defaults]`           | Any key                                                                           | The values of the tool.           |
| `[questions.<key>]`    | `question` (necessary, a string), `detect` (a string, a shell command), `example` | The question for one value.       |
| `[contributes.<tool>]` | Any key                                                                           | Values for another selected tool. |

The repo side examines these tables for each selected tool, and the Mac side for
each tool of the clone. These stop the sync:

- `[defaults]` is not a table.
- `[questions]` or `[contributes]` is not a table, or one of its entries is not
  a table.
- A `[questions.<key>]` has another key, has no `question`, or has a `question`
  or a
  `detect` that is not a string.

`bootstrap init` and `bootstrap add` ask the questions, and `detect` of a
question gives the proposal. See
[sync-engine.md](sync-engine.md#the-questions). `sync` and `check` stop with
code 2 for a question of a selected tool without an answer: run
`bootstrap init`. A template that needs a value without a question uses
`required`, which fails that file with its message.

## Tools and packs

A tool, for example `claude`, is a set of config for mise and for other CLIs.

A pack groups tools, for example the pack `security`. A tool never selects a
group of tools: a pack does. The config, the tasks and the files that all the
packs of one kind use are in one tool that each of these packs names as a
member, for example `flutter` for the Flutter packs, or in a common tool
`_<name>` that each member tool names in `depends`, for example
`_security_base`. See [packs.md](packs.md).

## The Flutter example

### Platforms

The Flutter documentation (Context7 `/flutter/website`) gives these platforms:

| Group   | Platform | Host OS for the build | Native toolchain                         |
| ------- | -------- | --------------------- | ---------------------------------------- |
| Mobile  | Android  | macOS, Windows, Linux | Android SDK, JDK, Gradle, Kotlin or Java |
| Mobile  | iOS      | macOS only            | Xcode, Swift or Objective-C, CocoaPods   |
| Desktop | macOS    | macOS only            | Xcode, Swift or Objective-C              |
| Desktop | Windows  | Windows only          | Visual Studio, C++                       |
| Desktop | Linux    | Linux only            | Clang, CMake, Ninja, GTK, C              |
| Web     | Web      | All                   | A browser (Chrome for `flutter run`)     |

### The Android dependencies

The user gave these dependencies for a Flutter app for Android:

- `Dart`: the primary language of Flutter. It has its own package manager.
- `Android-SDK`: necessary to build an Android app.
- `Kotlin`: necessary for `Android-SDK`. `Gradle` is the package manager and
  is part of `Kotlin`.
- Kotlin DSL (`kotlin-dsl`) is the only supported build script language.
  Android and Flutter support it since 2025.

### The core of Flutter

`repo/tools/flutter/tools.toml` has the mise tool `flutter`, which is the
Flutter SDK. `flutter` is its name in the mise registry. `flutter-sdk` is not a
name in the registry.

### The current state

- `repo/tools/flutter/tools.toml` has `depends = ["android"]`, so the mise tool
  `flutter` depends on `android-sdk`. Thus each Flutter consumer gets `android`
  and `java`, also a web-only or a desktop-only app.
- Gradle does not come from mise. The Android tasks use the `gradlew` wrapper
  of the project. The `kotlin` tool ships `ktlint`, `detekt` and `kotlin-lsp`.

## Tasks of a shared tool for one platform

A task that is for Flutter on one platform goes in the `flutter` tool, not in
the platform tool. For example, a task that runs `flutter build apk` is in
`repo/tools/flutter/`, not in `repo/tools/android/`.

- Each Flutter consumer gets the task, also without an Android target.
- If the project has no `android/` folder, the task prints nothing and exits 0.

## Secrets

Each task that runs the application uses fnox. See [fnox.md](fnox.md).
