# The `java` tool

The Java development kit (JDK), for Gradle and the Android SDK.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository builds with Gradle or Maven.
- Another selected tool adds it: `android` and `kotlin` have `java` in their
  `depends`.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
the root, or a folder one level below it, has a `*.gradle` or `*.gradle.kts`
file (see "Scripts").

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/conf.d/java/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version | Environments | Note                                                                                                                                                             |
| --------- | ------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `java`    | `lts`   | dev, ci      | mise resolves `lts` to the newest LTS release. The mise setting `java.shorthand_vendor = "temurin"` selects the vendor, so each machine installs the same build. |

## Tasks

None.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                                        |
| ---------------- | ------ | -------------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a Gradle file at the root or one level below it, else `none`. |

## Contributions

None.

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. A tool that adds `java` gives it its own environments (see
   [`docs/design/mise.md`](../../design/mise.md)). `kotlin` is in dev only, so
   `java` of `kotlin` is in dev only. To have it in CI, add `[tools.java]` to
   the values file. `flutter` adds `android`, and `android` adds `java`, in the
   environments of `flutter`.
2. If the repository pins a different JDK, tell the user. The values file
   can set it with `[overrides.mise.tools.java]`, for example
   `version = "21"`, and a different vendor with `[overrides.mise.settings]`,
   for example `java.shorthand_vendor = "openjdk"`.

## Migration

- Remove `java` from the repository-owned mise files.
