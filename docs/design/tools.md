# Tools design

`Tools` are fundamental building block of `bootstrap`. They define `mise`
configuration to install the required applications (also called as tools in
mise), tasks, environment, dependencies, etc. They also contain configuration of
the tool they are installing

## The `tools.toml` file

Each tool folder has a `tools.toml` with four necessary keys. The sync reads
them without defaults.

| Key            | Meaning                                                                                                                    |
| -------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `default_envs` | A list: the environments of the tool when the values file gives it no `env`, for example `["dev"]`. It must include `dev`. |
| `depends`      | The tools that the sync adds with this one. Each gets the environments of this tool.                                       |
| `root_only`    | `true`: the sync refuses the tool in a git submodule.                                                                      |
| `extensions`   | The VS Code extensions of the tool.                                                                                        |

`default_envs` and the `envs` of a mise tool can name only `dev` and `ci`, the
scope of `bootstrap`. Any other name stops the sync with code 2. A name of
`default_envs` that the values file selects, and that a
mise tool of the tool does not allow in its `envs`, stops the sync with
code 2. See [mise.md](mise.md#the-environments-of-a-mise-tool).

A `default_envs` without `dev` stops the sync with code 2, also for a tool
that the values file does not select. A tool folder that is `zz-repo`, or
whose name sorts after `zz-repo` in byte order, also stops the sync with
code 2: the folder `conf.d/zz-repo/` of the repository must load last.

## The `[files]` table

An optional `[files."<target>"]` table of a `tools.toml` gives the mode and the
condition of one file of the tool. A file without a table is `generated`.

The target is the path of the file without `.tmpl`. The target, or the target
with `.tmpl`, must be the same string as a path that `find -type f` gives in the
tool folder. Thus a case variant fails on each file system, and a path through a
symlinked folder fails. A target that fails stops the sync with code 2.

| Key              | Meaning                                                                                                                                 |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `mode`           | `generated` (the default): written on each run. `seed`: written only when it is missing. `merge`: merged into the live target.          |
| `requires`       | A list of exact paths, relative to the root of the repository. Each path must exist.                                                    |
| `requires_found` | A list of file names. The sync must find each name at the root or down to 4 folders below it, in a file that git does not ignore.       |
| `format`         | `text`, `json`, `yaml` or `toml`. It selects the parse check. The default comes from the extension.                                     |
| `remove`         | Only with `merge`. A list of key paths to delete from the live target.                                                                  |
| `remove_items`   | Only with `merge`. A list of tables with `path` and `items`. It deletes each listed item from that list, in the live target and source. |

When a `requires` or `requires_found` entry is not there, the sync does not
write the target. A generated target gets no row in
`.config/repo-generated.tsv`, so after a sync without a failure, the stale rule
removes a listed copy with the bootstrap header. A copy without the header, a
seed and a merge target stay.

The top-level `seed` key, the old table of seed conditions, and the `requires`
and the `files` of a mise entry are gone. Each stops the sync with code 2, and
the message names the new key.

### Merge mode

Merge mode is for a target that an app owns. The source holds only the keys that
the tool enforces. It has no bootstrap header. A YAML or TOML source can have
the
comment `# Merged by bootstrap: <path>`, and the sync strips nothing.

- The format is `json`, `yaml` or `toml`: the `format` key, else the extension.
  The JSONC rule does not apply.
- The order is render, parse check, merge with the live target, dprint, compare.
- The sync applies `remove` and `remove_items` to the live target, then it
  merges
  the source over the result. Tables merge key by key. A list keeps its items
  and
  adds each new source item. Any other source value replaces the live value. A
  key that is not in the source stays. `sync-engine/merge` has the full rules.
- The target path is the path in the repository. A target that is a symlink or
  is
  below a symlinked folder stays as it is, and the sync returns code 2.
- A merge target has no header, so the stale rule leaves it. It keeps the mode
  of an existing target, and a new one takes the mode of its source.

No tool uses merge mode now.

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
  is part of `Kotlin`, as `pnpm` is part of `node`.
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
  of the project. The `kotlin` tool ships only `kotlin-lsp`.

## Tasks of a shared tool for one platform

A task that is for Flutter on one platform goes in the `flutter` tool, not in
the platform tool. For example, a task that runs `flutter build apk` is in
`repo/tools/flutter/`, not in `repo/tools/android/`.

- Each Flutter consumer gets the task, also without an Android target.
- If the project has no `android/` folder, the task prints nothing and exits 0.

## Secrets

Each task that runs the application uses fnox. See [fnox.md](fnox.md).
