# The `java` tool

The Java development kit (JDK), for Gradle and the Android SDK.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository builds with Gradle or Maven.
- Another selected tool adds it: `android` and `kotlin` depend on `java`.

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/conf.d/java/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version | Environments | Note                                           |
| --------- | ------- | ------------ | ---------------------------------------------- |
| `java`    | `lts`   | dev, ci      | mise resolves `lts` to the newest LTS release. |

## Tasks

None.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. A tool that adds `java` gives it its own environments. `kotlin` is in dev
   only, so `java` of `kotlin` is in dev only. To have it in CI, add
   `[tools.java]` to the values file. Through the mise `depends` of `flutter`
   and `android-sdk`, `java` is in the environments of `flutter`.
2. If the repository pins a different JDK, tell the user.

## Migration

- Remove `java` from the repository-owned mise files.
