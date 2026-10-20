# Tools design

This document records the design decisions for the tools of `bootstrap` and
for a repository-level skill that adds a tool. The decisions come from a design
session on 2026-10-20. They are not implemented yet. See [Backlog](#backlog).

## The goal of the skill

The skill adds any new tool to `bootstrap`, with a correct dependency tree.
Flutter is only the example, because it is one of the most difficult tools.
The skill creates the tool files. It does not create the Flutter tool now. The
user stated:

> Please create skill only and not the tool itself

## How bootstrap is used

`bootstrap` uses AI and scripts together:

- **AI, for the non-deterministic part.** A user opens an AI session in the
  consumer repository with the prompt of `README.md`. The AI reads
  `docs/sync/index.md`, asks the questions, and writes `.config/bootstrap.toml`.
- **Script, for the deterministic part.** The `setup` task calls `repo-sync`.
  It reads `.config/bootstrap.toml` and syncs the files of the selected tools.

The repository solves two problems: add tools and dependencies to many
repositories without manual edits that drift, and keep the updates in sync.

The AI part is for a new repository, and after that only in rare situations.
The user stated:

> I want the user to use the AI part when their repo is new and post that in
> rare situations only

## Conditional dependencies become deterministic

The structure of the tools makes each dependency deterministic. The user
selects a tool, and the tool brings its dependencies. For example:

- A Flutter app: the user selects `flutter`. It brings `dart`, `android` and
  the other tools of the selected platforms.
- A native Android app: the user selects only `android`. The sync does not
  install `flutter` or `dart`.

## Multi-step questions

The AI asks the questions in layers. The user selects Flutter, then the
platforms, and so on. If the prompt of the user already gives the answers,
the AI does not ask those questions.

## The `options.toml` file

A tool folder can have `options.toml`. It lists the options that the user can
select. Each option names its dependencies. An option can mark a dependency as
optional.

- A folder can have `options.toml` only: a selection with no shared files.
- A folder can have `tool.toml` and `options.toml`: a real tool with options.
  For example, `tools/flutter/tool.toml` is the core (the Flutter SDK, Dart,
  and the common tasks), and `tools/flutter/options.toml` lists the platforms.
- One `options.toml` holds all the layers of the tree of that tool. Native
  Android and native iOS have their own `options.toml`, because they have more
  options than Flutter. The Flutter path does not ask the questions of the
  native Android file.

A first draft of the format, to show the idea. The format is not decided:

```toml
depends = [ "flutter-sdk" ] # always selected
# tools/flutter/options.toml
description = "Flutter apps for mobile, desktop and web."

[[question]]
id = "platforms"
multiple = true
prompt = "Which platforms does the app target?"

[[question.option]]
depends = [ "flutter-android" ]
id = "android"

[[question.option]]
depends = [ "flutter-ios" ]
id = "ios"
optional = [ "flutter-ios-pods" ] # the AI asks before it selects
```

## Optional dependencies

For each optional dependency, the AI asks a yes or no question.

## The answers go into the values file

The AI records all the answers in `.config/bootstrap.toml`: the answers of the
questions and the yes or no answers of the optional dependencies. For example,
`[options.flutter] platforms = ["android"]`.

`repo-sync` reads `options.toml` and resolves the answers to tools on each
sync. Thus, if `bootstrap` later adds a tool to an option, each consumer gets
it with no AI step.

## Folder layout

The layout is flat. Each real tool is a top-level folder in `tools/`:

```text
tools/
├── flutter/
│   ├── tool.toml        # the core of Flutter
│   └── options.toml     # the platform choices
├── flutter-android/     # the Android build and run tasks of Flutter
├── flutter-ios/         # the iOS build and run tasks of Flutter
├── flutter-web/         # the web build and run tasks of Flutter
└── android/             # native Android, also usable without Flutter
```

The names of the variant folders are examples, not decisions.

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

### The current state

- `tools/flutter/tool.toml` gives the mise tool `flutter` the static mise
  `depends` `android-sdk`. Thus each Flutter consumer gets `android` and
  `java`, also a web-only or a desktop-only app.
- Gradle does not come from mise. The Android tasks use the `gradlew` wrapper
  of the project. The `kotlin` tool ships only `kotlin-lsp`.

## Secrets

Each task that runs the application uses fnox. See [fnox.md](fnox.md).

## Open points

- The final format of `options.toml`.
- The format of the answers in `.config/bootstrap.toml`.
- The name and the location of the skill, for example
  `.claude/skills/<name>/SKILL.md`.
- The tasks that count as "tasks that run the application", and their names.

## Backlog

1. `repo-sync` reads `options.toml` and resolves the answers in
   `.config/bootstrap.toml` on each sync.
2. Rename `tool.toml` to `tools.toml` in all tools, the sync code and the
   docs. Most tools install more than one mise tool, and the name matches
   mise. Until this item lands, the skill writes `tool.toml`.
3. The `fnox` redesign. See [fnox.md](fnox.md#backlog).
