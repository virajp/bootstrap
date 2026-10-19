# The `android` tool

The Android SDK command-line tools.

| Property  | Value                                               |
| --------- | --------------------------------------------------- |
| Root only | No                                                  |
| Depends   | None. Its mise tool adds `java` (see "mise tools"). |
| Seeds     | None                                                |

## When the tool is relevant

- The repository has an Android build: a `settings.gradle.kts` with the
  Android Gradle plugin, or the `android/` folder of a Flutter app.
- The `flutter` tool adds it.

## What the tool ships

| Path in the repository                        | Kind      | Use                            |
| --------------------------------------------- | --------- | ------------------------------ |
| `.config/mise/tasks/setup/deps/android`       | Synced    | The task `setup:deps:android`. |
| `.config/mise/tasks/code/lint/android`        | Synced    | The task `code:lint:android`.  |
| `.config/mise/conf.d/android/mise.<env>.toml` | Generated | The mise config of this tool.  |

## mise tools

| mise tool     | Version | Environments | Note                                                    |
| ------------- | ------- | ------------ | ------------------------------------------------------- |
| `android-sdk` | `23`    | dev, ci      | Needs java. Sets `ANDROID_HOME` and `ANDROID_SDK_ROOT`. |

## Tasks

| Task                 | Use                                                                                                                                |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:android` | Install `platform-tools`, the platform of `compileSdk`, `emulator` and the emulator image. `--upgrade` installs the same packages. |
| `code:lint:android`  | Run Android lint on the Gradle modules of the given files. `--fix` runs `lintFix`.                                                 |

`setup:deps:android` applies only when the root has an executable `./gradlew`
and a `*.gradle.kts` file, `build-logic/` or `gradle/libs.versions.toml`
applies `com.android.application` or `com.android.library`. The `android/`
folder of a Flutter app is for `setup:deps:flutter`.

- The level is the env value `COMPILE_SDK`, else the first `compileSdk` in the
  `build.gradle.kts` files, then in `build-logic/`, then in
  `gradle/libs.versions.toml`. Only a literal number counts. Without a level,
  the task fails. A bare level of 37 or more becomes `<level>.0`.
- The emulator image is the env value `EMULATOR_IMAGE`, else
  `system-images/android-<targetSdk>/google_apis`, with `targetSdk` found as
  `compileSdk`. Without a `targetSdk` level, the task fails and tells the user
  to set `EMULATOR_IMAGE`. The task adds the ABI of the host: `arm64-v8a` on
  arm64, else `x86_64`. The ids use `/`, the form of `android sdk list`; an
  `EMULATOR_IMAGE` with `;` (the sdkmanager form) is accepted.
- A package is installed when `$ANDROID_HOME/<id>/package.xml` exists. If
  the image is not installed, the task checks it against
  `android sdk list --all`. If the image is not in the list and its `_ps16k`
  variant is, the task installs the variant and names it. From
  `android-37.1`, only the `_ps16k` images exist. If neither is there, the
  task fails and names both.
- It installs the packages of `platform-tools`, `platforms/android-<level>`,
  `emulator` and the image that are not installed, with `android sdk install`
  of the `android-sdk` mise tool. When all four are installed, the task does
  not list or install, so it needs no network. Only the image is checked
  against the list. The install accepts the SDK license and writes
  `licenses/`.

`code:lint:android` uses the Gradle root `./gradlew`, else `android/gradlew`,
and applies only when `$ANDROID_HOME/platforms` exists. Given files, it runs
only for a `src/` `.kt`, `.java` or `.xml` file, a `*.gradle.kts`, a `*.pro`,
`libs.versions.toml`, `lint.xml` or `lint-baseline.xml`. It lints only the
modules of those files: the nearest folder with `build.gradle.kts` or
`build.gradle`. A file at the Gradle root, in `gradle/`, `build-logic/` or
`buildSrc/` (not a Gradle project), or `settings.gradle.kts` lints every
module, as a run without files does. Every module is each Gradle project
below the Gradle root, so the pub plugins of a Flutter app, which are
projects of its `android/` build, are not linted. A temporary init script
runs `lint` (`lintFix` with `--fix`) in each of these modules that has the
task; a module without the Android Gradle plugin has none and is skipped.
`--debug` adds `--info`.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository sets `ANDROID_HOME` by hand, tell the user that the tool
   sets it.
2. If `setup:deps:android` finds no `compileSdk` level, for example when it
   comes from a variable such as `libs.versions.sdk.get().toInt()`, set
   `COMPILE_SDK` in the `[env]` of
   `.config/mise/conf.d/_repo/mise.<env>.toml`. If it finds no `targetSdk`
   level, or for another image, set `EMULATOR_IMAGE` there, for example
   `system-images/android-36/google_apis_playstore`.

## Migration

- Remove `android-sdk` from the repository-owned mise files.
