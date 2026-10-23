# The `kotlin` tool

The Kotlin language server, the ktlint and detekt gates, and the task that
resolves the Gradle dependencies against the lockfiles.

| Property     | Value                                               |
| ------------ | --------------------------------------------------- |
| Root only    | No                                                  |
| Depends      | `java`                                              |
| Default envs | dev                                                 |
| Seeds        | `.config/ktlint.editorconfig`, `.config/detekt.yml` |

## When the tool is relevant

- The repository has Kotlin code: `*.kt` or `*.kts` files, for example the
  Android part of a Flutter app.
- The user edits Kotlin with an editor or an agent that uses a language
  server.
- The repository root has a Gradle build with an executable `./gradlew`.
  `setup:deps:kotlin` needs a `distributionSha256Sum` in
  `gradle/wrapper/gradle-wrapper.properties`. It checks the lockfiles only of
  the projects with `dependencyLocking`. The `android/` build of a Flutter app
  is for `setup:deps:flutter`, not for this task.

## What the tool ships

| Path in the repository                       | Kind      | Use                                                |
| -------------------------------------------- | --------- | -------------------------------------------------- |
| `.config/mise/tasks/setup/deps/kotlin`       | Synced    | The task `setup:deps:kotlin`.                      |
| `.config/mise/tasks/code/format/ktlint`      | Synced    | The task `code:format:ktlint`.                     |
| `.config/mise/tasks/code/lint/detekt`        | Synced    | The task `code:lint:detekt`.                       |
| `.config/ktlint.editorconfig`                | Seed      | The ktlint rules: `ktlint_official`, lines of 120. |
| `.config/detekt.yml`                         | Seed      | The changes to the default detekt rules.           |
| `.config/mise/conf.d/kotlin/mise.<env>.toml` | Generated | The mise config of this tool.                      |

## mise tools

| mise tool              | Version  | Environments | Note                                                                                   |
| ---------------------- | -------- | ------------ | -------------------------------------------------------------------------------------- |
| `ktlint`               | `latest` | dev, ci      | The layout gate.                                                                       |
| `github:detekt/detekt` | `latest` | dev, ci      | The correctness gate. `detekt-cli` runs on the JDK of `java`.                          |
| `http:kotlin-lsp`      | `latest` | dev          | From the JetBrains downloads, for macOS arm64 and x64 only. In `[mise_env.dev.tools]`. |

`kotlin-lsp` is in `[mise_env.dev.tools]`, not in `[mise.tools]`. Thus it
stays in dev, and it does not stop `kotlin` in ci.

## Tasks

| Task                | Use                                                                                                                                                      |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:deps:kotlin` | Resolve the resolvable configurations of every Gradle project, and fail when the resolve does not match the lockfiles. `--upgrade` writes the lockfiles. |

- The task applies only when the repository root has an executable
  `./gradlew`. Else it prints nothing. It does not look in `android/` or in
  another folder.
- It fails when `gradle/wrapper/gradle-wrapper.properties` has no
  `distributionSha256Sum` of 64 hex digits.
- A normal run sets `LockMode.STRICT` on the settings buildscript classpath,
  and on each project and on its buildscript classpath, then resolves without
  `--write-locks`. Gradle enforces the lockfiles on the disk: a dynamic version
  stays at its locked version, and a
  resolve that does not match a lockfile fails. A locked configuration with no
  lock state also fails, for example a new module without a lockfile. The Gradle
  output names the project and the configuration. On any failure of
  `./gradlew`, the task tells the user to run `--upgrade` only if the output
  names the dependency lock state. The run changes no file.
- The task resolves each resolvable configuration, except the legacy
  `default` and `archives` configurations. A configuration from
  `configurations.create()`, for example of detekt or ktlint, is also
  consumable. The task resolves it, so its lock is examined and written.
- Gradle writes a lockfile only for a project with `dependencyLocking`. The
  task does not check a project without it. When no project has it, the
  normal run prints a warning and passes. The task finds the lockfiles below
  the Gradle root, but not in a folder with `.git` or `settings.gradle(.kts)`:
  a git submodule or a nested Gradle build.
- Each `./gradlew` call of the task passes `--no-configuration-cache`. The
  resolve and `dependencyUpdates` cannot use the configuration cache, so Gradle
  prints no cache warnings.
- `--upgrade` resolves with `--write-locks --refresh-dependencies` and keeps
  the new lockfiles. Then it runs `./gradlew dependencyUpdates` if the
  ben-manes plugin `com.github.ben-manes.versions` is applied. It ends with
  "Commit the new lockfiles." when the resolve added, changed or removed a
  lockfile. It compares a checksum of each lockfile before and after the
  resolve, so it does not use git. It ends with a warning when the build has
  no lockfile.

`code:format:ktlint` checks the Kotlin files with ktlint, and `--fix` formats
them. `code:lint:detekt` lints them with detekt. `code:format` and `code:lint`
run them. A gate skips without its tool and without a Kotlin file.

- Without files, the list comes from `git ls-files`, so a gate never reads an
  ignored folder. Both gates skip the files in a `generated/` folder.
- With a list of files, ktlint uses only the `.kt` and `.kts` files. detekt
  needs only one Kotlin file in the list: then it lints each Kotlin file,
  because a finding of detekt can be about more than one file.
- detekt reads `.config/detekt.yml` on the default rules
  (`--build-upon-default-config`). A path with whitespace, or with the
  separator of `--input`, is skipped with a warning.

## Pre-commit hooks

None.

## Files the repository supplies

None.

## Adapt to the repository

1. `kotlin-lsp` has binaries for macOS only. If the user develops on a
   different platform, tell the user. ktlint and detekt work on macOS and
   Linux.
2. To run the gates in CI, set `env = ["dev", "ci"]` in `[tools.kotlin]`.
   Then `java` is in ci too.
3. The tool adds `java` as a dependency, and `java` gets the environments of
   `kotlin`: dev only. Another selected tool can put `java` in CI through a
   mise `depends` (for example `flutter`). `setup` runs `setup:deps:kotlin` in
   every environment, and the task runs `./gradlew` when the root has one. If
   the root has an executable `./gradlew`, CI runs `mise run setup`, and `java`
   is not in CI, add `[tools.java]` to the values file. See `java.md`.
4. If a project of the root Gradle build does not lock its dependencies, tell
   the user that `setup:deps:kotlin` does not check that project. It is
   checked after the project has
   `dependencyLocking { lockAllConfigurations() }` and its lockfile is
   committed (`mise run setup:deps:kotlin --upgrade` writes it).
5. If `gradle/wrapper/gradle-wrapper.properties` has no
   `distributionSha256Sum`, tell the user to add the checksum that Gradle
   publishes for the distribution.
6. In a Flutter app, the Gradle build is in `android/`, and
   `setup:deps:flutter` handles it. `setup:deps:kotlin` does nothing there.
7. If the repository has its own ktlint `.editorconfig` or detekt config, tell
   the user. The gates read only `.config/ktlint.editorconfig` and
   `.config/detekt.yml`.

## Migration

- Remove `http:kotlin-lsp` from the repository-owned mise files.
- Delete the old stackgen copies: `conf.d/ktlint/` and `conf.d/detekt/`, and
  the task `code:lint:ktlint` (M19 of `migrate.md`).
- The old `.config/ktlint.editorconfig` and `.config/detekt.yml` of stackgen
  stay as they are: the sync does not write a seed that exists.
