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
  `setup:deps:kotlin` needs a `distributionSha256Sum` in
  `gradle/wrapper/gradle-wrapper.properties`. It checks the lockfiles only of
  the projects with `dependencyLocking`. The `android/` build of a Flutter app
  is for `setup:deps:flutter`, not for this
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

| Task                | Use                                                                                                                               |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:kotlin` | Resolve every configuration of every Gradle project, and fail when a lockfile differs from git. `--upgrade` writes the lockfiles. |

- The task applies only when the repository root has an executable
  `./gradlew`. Else it prints nothing. It does not look in `android/` or in
  another folder.
- It fails when `gradle/wrapper/gradle-wrapper.properties` has no
  `distributionSha256Sum` of 64 hex digits.
- A normal run resolves with `--write-locks`, then compares each
  `*gradle.lockfile` with git. A lockfile that differs from git, or that git
  does not track, fails the run and is named. So a module with
  `dependencyLocking` but without a committed lockfile fails even without
  `LockMode.STRICT`. The task tells `--upgrade` for a lockfile that the
  resolve changed or created, and a commit for one that already differed from
  git before the run. The run restores the lockfiles as it found them.
- Gradle writes a lockfile only for a project with `dependencyLocking`. The
  task does not check a project without it. When no project has it, the
  normal run prints a warning and passes.
- `--upgrade` resolves with `--write-locks --refresh-dependencies` and keeps
  the new lockfiles. Then it runs `./gradlew dependencyUpdates` if the
  ben-manes plugin `com.github.ben-manes.versions` is applied. It ends with
  "Commit the new lockfiles." when a lockfile differs from git, and with a
  warning when the build has no lockfile.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. The tool has binaries for macOS only. If the user develops on a different
   platform, tell the user.
2. The tool adds `java` only as a dependency, so `java` is in dev only,
   unless another selected tool puts `java` in CI through a mise `depends`
   (for example `flutter`). `setup` runs `setup:deps:kotlin` in every
   environment, and the task runs `./gradlew` when the root has one. If the
   root has an executable `./gradlew`, CI runs `mise run setup`, and `java` is
   not in CI, add `java` to `tools`. See `java.md`.
3. If a project of the root Gradle build does not lock its dependencies, tell
   the user that `setup:deps:kotlin` does not check that project. It is
   checked after the project has
   `dependencyLocking { lockAllConfigurations() }` and its lockfile is
   committed (`mise run setup:deps:kotlin --upgrade` writes it).
4. If `gradle/wrapper/gradle-wrapper.properties` has no
   `distributionSha256Sum`, tell the user to add the checksum that Gradle
   publishes for the distribution.
5. In a Flutter app, the Gradle build is in `android/`, and
   `setup:deps:flutter` handles it. `setup:deps:kotlin` does nothing there.

## Migration

- Remove `http:kotlin-lsp` from the repository-owned mise files.
