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

| Path in the repository                        | Kind      | Use                           |
| --------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/android/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool     | Version | Environments | Note                                                    |
| ------------- | ------- | ------------ | ------------------------------------------------------- |
| `android-sdk` | `23`    | dev, ci      | Needs java. Sets `ANDROID_HOME` and `ANDROID_SDK_ROOT`. |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. If the repository sets `ANDROID_HOME` by hand, tell the user that the tool
   sets it.

## Migration

- Remove `android-sdk` from the repository-owned mise files.
