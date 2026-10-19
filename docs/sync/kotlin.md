# The `kotlin` tool

The Kotlin language server, and the task that resolves the Gradle
dependencies against the committed lockfiles.

| Property  | Value  |
| --------- | ------ |
| Root only | No     |
| Depends   | `java` |
| Seeds     | None   |

## When the tool is relevant

- The repository has Kotlin code: `*.kt` or `*.kts` files, for example the
  Android part of a Flutter app.
- The user edits Kotlin with an editor or an agent that uses a language
  server.
- The repository root has a Gradle build with an executable `./gradlew`.
  `setup:deps:kotlin` needs committed lockfiles (`dependencyLocking`) and a
  `distributionSha256Sum` in `gradle/wrapper/gradle-wrapper.properties`. The
  `android/` build of a Flutter app is for `setup:deps:flutter`, not for this
  task.

## What the tool ships

| Path in the repository                     | Kind      | Use                           |
| ------------------------------------------ | --------- | ----------------------------- |
| `.config/mise/tasks/setup/deps/kotlin`     | Synced    | The task `setup:deps:kotlin`. |
| `.config/mise/conf.d/kotlin/mise.dev.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool         | Version  | Environments | Note                                                        |
| ----------------- | -------- | ------------ | ----------------------------------------------------------- |
| `http:kotlin-lsp` | `latest` | dev          | From the JetBrains downloads, for macOS arm64 and x64 only. |

## Tasks

| Task                | Use                                                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `setup:deps:kotlin` | Resolve every configuration of every Gradle project, and fail when the committed lockfiles do not cover it. `--upgrade` writes them. |

- The task applies only when the repository root has an executable
  `./gradlew`. Else it prints nothing. It does not look in `android/` or in
  another folder.
- It fails when `gradle/wrapper/gradle-wrapper.properties` has no
  `distributionSha256Sum` of 64 hex digits.
- A normal run resolves with `--write-locks`, then compares each
  `*gradle.lockfile` with git. A changed or new lockfile fails the run and is
  named, so a module without a committed lockfile fails even without
  `LockMode.STRICT`. The run restores the lockfiles as it found them.
- `--upgrade` resolves with `--write-locks --refresh-dependencies` and keeps
  the new lockfiles. Then it runs `./gradlew dependencyUpdates` if the
  ben-manes plugin `com.github.ben-manes.versions` is applied. Commit the
  lockfiles after it.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. The tool has binaries for macOS only. If the user develops on a different
   platform, tell the user.
2. The tool adds `java` only as a dependency, so `java` is in dev only,
   unless another selected tool puts `java` in CI through a mise `depends`
   (for example `flutter`). If CI builds with Gradle and `java` is not in CI,
   add `java` to `tools`. See `java.md`.
3. If the root Gradle build does not lock its dependencies, tell the user that
   `setup:deps:kotlin` fails until each project has
   `dependencyLocking { lockAllConfigurations() }` and the lockfiles are
   committed (`mise run setup:deps:kotlin --upgrade` writes them).
4. If `gradle/wrapper/gradle-wrapper.properties` has no
   `distributionSha256Sum`, tell the user to add the checksum that Gradle
   publishes for the distribution.
5. In a Flutter app, the Gradle build is in `android/`, and
   `setup:deps:flutter` handles it. `setup:deps:kotlin` does nothing there.

## Migration

- Remove `http:kotlin-lsp` from the repository-owned mise files.
