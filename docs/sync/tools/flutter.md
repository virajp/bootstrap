# The `flutter` tool

Flutter and Dart. The tool adds the `android` tool for the Android build.

| Property     | Value     |
| ------------ | --------- |
| Root only    | No        |
| Depends      | `android` |
| Parent       | None      |
| Default envs | dev, ci   |
| Seeds        | None      |

## When the tool is relevant

- The repository has `pubspec.yaml`.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
`pubspec.yaml` depends on the flutter SDK (see "Scripts").

## What the tool ships

| Path in the repository                        | Kind      | Use                               |
| --------------------------------------------- | --------- | --------------------------------- |
| `.config/mise/tasks/setup/deps/20-flutter`    | Synced    | The task `setup:deps:20-flutter`. |
| `.config/mise/tasks/code/format/10-dart`      | Synced    | The task `code:format:10-dart`.   |
| `.config/mise/tasks/code/lint/20-dart`        | Synced    | The task `code:lint:20-dart`.     |
| `.config/mise/conf.d/flutter/mise.<env>.toml` | Generated | The mise config of this tool.     |

## mise tools

| mise tool | Version | Environments |
| --------- | ------- | ------------ |
| `flutter` | `3`     | dev, ci      |

The tool has `android` in its `depends`, and `android` has `java`. Thus the
sync adds both tools, in the environments of `flutter` (dev and ci). See
[`android.md`](android.md) and [`java.md`](java.md).

## Tasks

| Task                    | Use                                                                                                                                                                         |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:20-flutter` | Install the Dart, Flutter, CocoaPods and Gradle dependencies from the lockfiles, then report the outdated ones. `--upgrade` cleans the caches and lockfiles, then upgrades. |
| `code:format:10-dart`   | Check or format the existing `.dart` files of the `[files]` arguments, else `lib/` and `test/`.                                                                             |
| `code:lint:20-dart`     | Analyze the existing `.dart` files of the `[files]` arguments, else the whole project. An info also fails the task (`--fatal-infos`). `dart analyze` only reports.          |

The two Dart tasks apply only when the root has `pubspec.yaml`. Without
files, they also need a `.dart` file that git does not ignore.

`setup:deps:20-flutter` uses CocoaPods only with `ios/Podfile`, and Gradle only
with an executable `android/gradlew`. It accepts the Android licenses. It
installs `platform-tools` with `android-sdk`, which comes from the `android`
tool.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                                                                      |
| ---------------- | ------ | -------------------------------------------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` when `pubspec.yaml` has `sdk: flutter`. `none` for a `pubspec.yaml` of Dart only, or without it. |

## Contributions

| Tool     | Shared file               | Items                                      |
| -------- | ------------------------- | ------------------------------------------ |
| `vscode` | `.vscode/extensions.json` | `dart-code.dart-code`, `dart-code.flutter` |

The tool adds no hook. The `format` hook of `dprint` and the `lint` hook of
`linter` run the Dart tasks through `code:format` and `code:lint`.

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File           | Used by                                                                 |
| -------------- | ----------------------------------------------------------------------- |
| `pubspec.yaml` | `setup:deps:20-flutter`. The task skips without it.                     |
| `pubspec.lock` | `setup:deps:20-flutter`. The install uses the versions of the lockfile. |

## Adapt to the repository

1. If the repository pins a different Flutter version, tell the user.
2. Move the repository-owned `[dart]` keys of `.vscode/settings.json` (see
   "Adapt to the repository" in `vscode.md`).

## Migration

- Remove `flutter` from the repository-owned mise files. For `android-sdk`
  and `java`, see `android.md` and `java.md`.
- M16 of `migrate.md` has the move of `android-sdk` and `java` out of
  `flutter`.
- If the Gradle build uses the Groovy DSL (`*.gradle`), do the migration of
  [`kotlin-dsl.md`](../kotlin-dsl.md) when the user agrees.
