# The `kotlin` tool

The Kotlin language server, and the task that resolves the Gradle
dependencies against the lockfiles.

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

| Task                | Use                                                                                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:kotlin` | Resolve the resolvable-only configurations of every Gradle project, and fail when the resolve does not match the lockfiles. `--upgrade` writes the lockfiles. |

- The task applies only when the repository root has an executable
  `./gradlew`. Else it prints nothing. It does not look in `android/` or in
  another folder.
- It fails when `gradle/wrapper/gradle-wrapper.properties` has no
  `distributionSha256Sum` of 64 hex digits.
- A normal run sets `LockMode.STRICT` on each project and on its buildscript
  classpath, then resolves without
  `--write-locks`. Gradle enforces the lockfiles on the disk: a dynamic version
  stays at its locked version, and a resolve that does not match a lockfile
  fails. A locked configuration with no lock state also fails, for example a
  new module without a lockfile. The Gradle output names the project
  and the configuration, and the task tells `--upgrade`. The run changes no
  file.
- The task resolves only the configurations that are resolvable and not
  consumable (`canBeConsumed`). So it does not resolve the legacy `default`
  and `archives` configurations.
- Gradle writes a lockfile only for a project with `dependencyLocking`. The
  task does not check a project without it. When no project has it, the
  normal run prints a warning and passes.
- `--upgrade` resolves with `--write-locks --refresh-dependencies` and keeps
  the new lockfiles. Then it runs `./gradlew dependencyUpdates` if the
  ben-manes plugin `com.github.ben-manes.versions` is applied. It ends with
  "Commit the new lockfiles." when the resolve added, changed or removed a
  lockfile. It compares a checksum of each lockfile before and after the
  resolve, so it does not use git. It ends with a warning when the build has
  no lockfile.

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
