# The `flutter` tool

Flutter and Dart, with the Android toolchain.

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

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

| mise tool     | Version | Environments | Note       |
| ------------- | ------- | ------------ | ---------- |
| `flutter`     | `3`     | dev, ci      |            |
| `android-sdk` | `23`    | dev, ci      | Needs java |
| `java`        | `21`    | dev, ci      |            |

## Tasks

| Task                 | Use                                                                                                                                                                         |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:flutter` | Install the Dart, Flutter, CocoaPods and Gradle dependencies from the lockfiles, then report the outdated ones. `--upgrade` cleans the caches and lockfiles, then upgrades. |
| `code:format:dart`   | Check or format the `.dart` files. The default is `lib/` and `test/`.                                                                                                       |
| `code:lint:dart`     | Analyze the `.dart` files. `dart analyze` only reports.                                                                                                                     |

`setup:deps:flutter` uses CocoaPods only with `ios/Podfile`, and Gradle only
with an executable `android/gradlew`. It accepts the Android licenses.

## Pre-commit hooks

None. The `format` and `lint` hooks of `base` run the Dart tasks.

## Files the repository supplies

| File           | Used by                                                              |
| -------------- | -------------------------------------------------------------------- |
| `pubspec.yaml` | `setup:deps:flutter`. The task skips without it.                     |
| `pubspec.lock` | `setup:deps:flutter`. The install uses the versions of the lockfile. |

## Adapt to the repository

1. If the repository pins a different Flutter, Android SDK or Java version,
   tell the user.
2. Move the repository-owned `[dart]` keys of `.vscode/settings.json` (M9 of
   `migrate.md`).

## Migration

- Remove flutter, android-sdk and java from the repository-owned mise files.
