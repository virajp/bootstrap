# The `flutter` tool

Flutter and Dart. The tool adds the `android` tool for the Android build.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `pubspec.yaml`.

## What the tool ships

| Path in the repository                        | Kind      | Use                            |
| --------------------------------------------- | --------- | ------------------------------ |
| `.config/mise/tasks/setup/deps/flutter`       | Synced    | The task `setup:deps:flutter`. |
| `.config/mise/tasks/code/format/dart`         | Synced    | The task `code:format:dart`.   |
| `.config/mise/tasks/code/lint/dart`           | Synced    | The task `code:lint:dart`.     |
| `.config/mise/conf.d/flutter/mise.<env>.toml` | Generated | The mise config of this tool.  |

VS Code extensions: `dart-code.dart-code`, `dart-code.flutter`.

## mise tools

| mise tool | Version | Environments | Note               |
| --------- | ------- | ------------ | ------------------ |
| `flutter` | `3`     | dev, ci      | Needs android-sdk. |

`flutter` has the mise `depends` `android-sdk`. Thus the sync adds the
`android` tool, and `android-sdk` and `java` are in the environments of
`flutter` (dev and ci). See `android.md` and `java.md`.

## Tasks

| Task                 | Use                                                                                                                                                                         |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:flutter` | Install the Dart, Flutter, CocoaPods and Gradle dependencies from the lockfiles, then report the outdated ones. `--upgrade` cleans the caches and lockfiles, then upgrades. |
| `code:format:dart`   | Check or format the existing `.dart` files of the `[files]` arguments, else `lib/` and `test/`.                                                                             |
| `code:lint:dart`     | Analyze the existing `.dart` files of the `[files]` arguments, else the whole project. An info also fails the task (`--fatal-infos`). `dart analyze` only reports.          |

The two Dart tasks apply only when the root has `pubspec.yaml`.

`setup:deps:flutter` uses CocoaPods only with `ios/Podfile`, and Gradle only
with an executable `android/gradlew`. It accepts the Android licenses. It
installs `platform-tools` with `android-sdk`, which comes from the `android`
tool.

## Pre-commit hooks

None. The `format` hook of `dprint` and the `lint` hook of `linter` run the Dart
tasks.

## Files the repository supplies

| File           | Used by                                                              |
| -------------- | -------------------------------------------------------------------- |
| `pubspec.yaml` | `setup:deps:flutter`. The task skips without it.                     |
| `pubspec.lock` | `setup:deps:flutter`. The install uses the versions of the lockfile. |

## Adapt to the repository

1. If the repository pins a different Flutter version, tell the user.
2. Move the repository-owned `[dart]` keys of `.vscode/settings.json` (see
   "Adapt to the repository" in `vscode.md`).

## Migration

- Remove `flutter` from the repository-owned mise files. For `android-sdk`
  and `java`, see `android.md` and `java.md`.
