# Tools design

`Tools` are fundamental building block of `bootstrap`. They define `mise`
configuration to install the required applications (also called as tools in
mise), tasks, environment, dependencies, etc. They also contain configuration of
the tool they are installing

## The `tools.toml` file

Each tool folder has a `tools.toml` with five necessary keys. The sync reads
them without defaults.

| Key            | Meaning                                                                                                                    |
| -------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `default_envs` | A list: the environments of the tool when the values file gives it no `env`, for example `["dev"]`. It must include `dev`. |
| `depends`      | The tools that the sync adds with this one. Each gets the environments of this tool.                                       |
| `root_only`    | `true`: the sync refuses the tool in a git submodule.                                                                      |
| `extensions`   | The VS Code extensions of the tool.                                                                                        |
| `seed`         | The files that the sync writes only when they are missing.                                                                 |

`default_envs` and the `envs` of a mise tool can name `dev`, `ci`, `prod` and
`stage`. A name of `default_envs` that the values file selects, and that a
mise tool of the tool does not allow in its `envs`, stops the sync with
code 2. See [mise.md](mise.md#the-environments-of-a-mise-tool).

A `default_envs` without `dev` stops the sync with code 2, also for a tool
that the values file does not select. A tool folder that is `zz-repo`, or
whose name sorts after `zz-repo` in byte order, also stops the sync with
code 2: the folder `conf.d/zz-repo/` of the repository must load last.

## Kinds of tool

A tool is one of two kinds:

- **A plain tool**, for example `claude`. It is a set of config for mise and
  for other CLIs.
- **A group tool**, for example `flutter`. It has its own config, as a plain
  tool has. It also has packs, which select other tools. See
  [packs.md](packs.md).

## Packs

A group tool has packs. See [packs.md](packs.md).

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

`tools/flutter/tools.toml` has the mise tool `flutter`, which is the Flutter
SDK. `flutter` is its name in the mise registry. `flutter-sdk` is not a name in
the registry.

### The current state

- `tools/flutter/tools.toml` gives the mise tool `flutter` the static mise
  `depends` `android-sdk`. Thus each Flutter consumer gets `android` and
  `java`, also a web-only or a desktop-only app.
- Gradle does not come from mise. The Android tasks use the `gradlew` wrapper
  of the project. The `kotlin` tool ships only `kotlin-lsp`.

## Tasks of a group tool for one platform

A task that is for Flutter on one platform goes in the `flutter` tool, not in
the platform tool. For example, a task that runs `flutter build apk` is in
`tools/flutter/`, not in `tools/android/`.

- Each Flutter consumer gets the task, also without an Android target.
- If the project has no `android/` folder, the task prints nothing and exits 0.

## Secrets

Each task that runs the application uses fnox. See [fnox.md](fnox.md).
