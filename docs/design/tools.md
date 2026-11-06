# Tools design

A tool is the building block of `bootstrap`. A tool folder holds the files of
the
tool at the path that they get in the target, and a `tools.toml`. This document
holds the `tools.toml` schema of both sides, the `[files]` keys, the folder
rules and
the rule for the header of a source file.

See also: [packs.md](packs.md), [mise.md](mise.md) for `[mise.*]`,
[sync-engine.md](sync-engine.md) for merge mode and the values merge,
[repo-sync.md](repo-sync.md) and [setup-sync.md](setup-sync.md). The exit code
of each rule that stops the sync is in
[sync-engine.md](sync-engine.md#the-exit-codes).

## The two sides

| Side    | Tools root     | A file goes to             | Reads the tools of       |
| ------- | -------------- | -------------------------- | ------------------------ |
| `repo`  | `repo/tools/`  | The path in the repository | `sync-engine/repo-sync`  |
| `macos` | `macos/tools/` | The path below `$HOME`     | `sync-engine/setup-sync` |

Each file in a tool folder is at the path that it gets in the target, without
the
`.tmpl` suffix of a template. For example,
`repo/tools/git/.config/mise/tasks/code/merge/develop` becomes
`.config/mise/tasks/code/merge/develop` in a consumer. Some files refer to files
that
are not in the tool folder, for example `dprint.jsonc`. The consumer supplies
them.

## The `tools.toml` file

Each tool folder has a `tools.toml`. A Mac tool folder without it stops the
sync.

| Key                    | `repo`    | `macos`  | Meaning                                                                                                                    |
| ---------------------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------------------- |
| `default_envs`         | Necessary | Refused  | A list: the environments of the tool when the values file gives it no `env`, for example `["dev"]`. It must include `dev`. |
| `depends`              | Necessary | Optional | The tools that the sync adds with this one, transitively. On the repo side, each gets the environments of this tool.       |
| `root_only`            | Necessary | Refused  | `true`: the tool is not for a git submodule. See [repo-sync.md](repo-sync.md#root_only).                                   |
| `extensions`           | Necessary | Refused  | The VS Code extensions of the tool.                                                                                        |
| `description`          | Not read  | Optional | A string: one line that tells the adopting AI what the tool is.                                                            |
| `stage`                | Not read  | Optional | `"init"` puts the tool in the init stage. See [setup-sync.md](setup-sync.md#stages-and-exit-4).                            |
| `[files."<target>"]`   | Optional  | Optional | The mode and the condition of one file. See [The `[files]` table](#the-files-table).                                       |
| `[defaults]`           | Optional  | Optional | The values of the tool. See [below](#the-defaults-questions-and-contributes-tables).                                       |
| `[questions.<key>]`    | Optional  | Optional | The question for one value.                                                                                                |
| `[contributes.<tool>]` | Optional  | Optional | Values for another selected tool.                                                                                          |
| `[[precommit]]`        | Optional  | Refused  | One `repos:` entry of `.config/pre-commit-config.yaml`, with its `[[precommit.hooks]]`.                                    |
| `[mise.*]`             | Optional  | Optional | The mise config of the tool. See [mise.md](mise.md#the-mise-sections).                                                     |
| `[mise_env.<env>.*]`   | Optional  | Refused  | More mise config for one environment. See [mise.md](mise.md#the-mise-sections).                                            |

On the repo side, the four necessary keys are necessary because the sync reads
them
without defaults. The sync examines the manifest of each tool of the clone,
selected
or not.

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
- A tool folder with both `<path>` and `<path>.tmpl` stops the sync.
- A tool does not ship mise files, and a tool that ships a file that the sync
  also
  generates stops the sync. See [mise.md](mise.md#the-mise-sections)
  for the
  exceptions. A Mac tool that ships a file below `.config/mise/conf.d/` also
  stops the
  sync.

A tool name is the name of its folder. A pack name must not be a tool folder.

### The old keys

These keys are gone. Each stops the sync with a message that names
the new
key: the top-level `seed` key, the old table of seed conditions
`[seed_requires]`, and
the `requires` and the `files` of a mise entry. On the Mac side, `stage =
"shell"` also
stops the sync and names `stage = "init"`.

## The `[files]` table

An optional `[files."<target>"]` table of a `tools.toml` gives the mode and the
condition of one file of the tool. A file without a table is `generated`.

The target is the path of the file without `.tmpl`. On the repo side, the
target, or the
target with `.tmpl`, must be the same string as a path that `find -type f` gives
in the
tool folder. Thus a case variant fails on each file system, and a path through a
symlinked folder fails. A target that ends in `.tmpl` also fails. On the Mac
side, the target is relative to `$HOME`, with
no
`~/`, and a selected tool must have a file that produces it. Thus the path of a
template with its `.tmpl` also fails. A target that fails
stops
the sync.

| Key              | Side | Meaning                                                                                                                                         |
| ---------------- | ---- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `mode`           | Both | `generated` (the default): written on each run. `seed`: written only when it is missing. `merge`: merged into the live target.                  |
| `requires`       | Both | A list of exact paths. Each path must exist. It is relative to the root of the repository, or to `$HOME` on the Mac side. It has no `..` part.  |
| `requires_found` | Repo | A list of file names. See [repo-sync.md](repo-sync.md#requires_found). On the Mac side, it stops the sync.                                      |
| `format`         | Both | `text`, `json`, `yaml` or `toml`. It selects the parse check. The default comes from the extension.                                             |
| `remove`         | Both | Only with `merge`. A list of key paths, each a non-empty list of strings. It deletes each path from the live target.                            |
| `remove_items`   | Both | Only with `merge`. A list of tables with exactly `path` and `items`. It deletes each listed item from that list, in the live target and source. |

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
- Merge mode is in [sync-engine.md](sync-engine.md#merge-mode). The
  `claude-cli`, `gem`,
  `gh` and `superfile` tools of `macos/tools/` use merge mode now. No tool of
  `repo/tools/` uses it.
- `test:macos-tools` stops when a `[files]` table gives `mode = "seed"` to a
  JSON, YAML
  or TOML target of a Mac tool.

## The header of a source file

Each file that is neither a seed nor a merge source has this header in its
source, in
the comment syntax of the file. In a template, `<path>` keeps the `.tmpl`
suffix:

```text
# Synced from bootstrap: repo/tools/<tool>/<path>. Do not edit this file.
# The next sync replaces it, and a manual change is lost.
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

A seed and a merge source have no header:

- A YAML or TOML merge source can have the comment `# Merged by bootstrap:
  <path>`. A
  JSON merge source has no comment. The sync strips nothing from a merge source.
- The forwarding file keeps its old header with `tools/repo-sync/...`, because
  the sync
  does not ship it.

## The `[defaults]`, `[questions]` and `[contributes]` tables

A `tools.toml` can have three optional tables for the values of a tool. A
template reads
the values as `.config` (the values of its tool) and `.all` (the values of each
selected
tool, by tool name). The `pre-commit` tool of `repo/tools/` uses `[defaults]`,
and no other tool of `repo/tools/` uses these tables now. Mac tools such as
`fish`, `git`, `ssh`, `gh` and `macos-defaults` use `[defaults]` or
`[questions]`. The
Mac tools `1password` and `orbstack` use `[contributes]`. The merge of these
values is
in [sync-engine.md](sync-engine.md#the-values-merge).

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
each
tool of the clone. `repo_sync --check` does not examine them, because they need
the
clone. These stop the sync:

- `[defaults]` is not a table.
- `[questions]` or `[contributes]` is not a table, or one of its entries is not
  a table.
- A `[questions.<key>]` has another key, has no `question`, or has a `question`
  or a
  `detect` that is not a string.

The sync does not ask a question and does not run `detect`. A missing answer is
not an
error of the engine: a template that needs an answer uses `required`, which
fails that
file with its message. Exit code 4 is only for the Mac front-end.

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

- `repo/tools/flutter/tools.toml` gives the mise tool `flutter` the static mise
  `depends` `android-sdk`. Thus each Flutter consumer gets `android` and
  `java`, also a web-only or a desktop-only app.
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
