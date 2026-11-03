# Migrate a Gradle build from the Groovy DSL to the Kotlin DSL

These instructions are for an AI agent that works in a consumer repository.
They migrate each Gradle build script of the repository from the Groovy DSL
(`*.gradle`) to the Kotlin DSL (`*.gradle.kts`).

`bootstrap` uses the Kotlin DSL for Gradle. The `kotlin`, `android` and
`flutter` tools support a Groovy build, but this migration is the recommended
next step for it. [`vwf.md`](vwf.md) and the "Migration" sections of
`tools/kotlin.md`, `tools/android.md` and `tools/flutter.md` send an agent
here.

The source of the rules below is the Gradle guide "Migrating build logic from
Groovy to Kotlin":
<https://docs.gradle.org/current/userguide/migrating_from_groovy_to_kotlin_dsl.html>.
Read it when a script has a construct that this file does not show.

## When the migration applies

The migration applies when both conditions are true:

1. The repository has a Gradle build: an executable `gradlew` at the root, or
   in `android/` for a Flutter app. The folder of `gradlew` is the root of the
   build.
2. In the root of the build, `git ls-files` lists a file whose name ends in
   `.gradle`.

```bash
for root in . android; do
  [[ -x "${root}/gradlew" ]] || continue
  git -C "${root}" ls-files -- '*.gradle'
done
```

Each file of the output is a Groovy script that this migration changes. A
build that is partly migrated, for example `settings.gradle.kts` with one
module that still has `build.gradle`, also applies. Then migrate only the
files that are left.

If the output is empty, the migration does not apply. Change nothing.

The init scripts in `~/.gradle/` and the scripts outside the repository are
not part of the repository. Do not change them.

## Before you start

1. Tell the user which files the migration changes, and get a yes. This
   migration changes the build of the repository.
2. Make sure that the git working tree is clean. If it is not, stop and tell
   the user.
3. In the root of each build, run the build one time before you change a
   file, and keep the result:
   - `./gradlew help`, for the configuration of all the projects;
   - the check of the repository, for example `./gradlew build`, or the
     command of its CI. For a Flutter app, use `flutter build apk --debug`
     at the repository root.

   If the build fails before the migration, stop and tell the user. A
   migration cannot prove a build that already fails.
4. For a Flutter app, make sure that `android/app/build.gradle` applies the
   Flutter plugin in the `plugins {}` block
   (`id "dev.flutter.flutter-gradle-plugin"`). If it applies
   `flutter.gradle` with `apply from:`, do the Flutter migration to the
   `plugins {}` block first:
   <https://docs.flutter.dev/release/breaking-changes/flutter-gradle-plugin-apply>.

## Migrate the scripts

Migrate one build at a time. In a build, migrate `settings.gradle` first, then
the root `build.gradle`, then each module, then each script that another
script applies. Run `./gradlew help` after each file. A build can have Groovy
and Kotlin scripts at the same time, so each step leaves a build that
operates.

For each Groovy script:

1. Prepare the script in Groovy, as the Gradle guide tells:
   - Use double quotes for each string.
   - Use parentheses for each function call, for example
     `implementation("com.acme:example:1.0")`.
   - Use `=` for each property assignment, for example `group = "com.acme"`.
2. Rename the file with `git mv`, so that git keeps its history:
   `git mv build.gradle build.gradle.kts`. A script that is applied to
   `Settings` gets the name `*.settings.gradle.kts`.
3. Change the syntax to Kotlin. The table shows the frequent changes.
4. Change each reference to the old name, for example
   `apply from: "gradle/publish.gradle"` becomes
   `apply(from = "gradle/publish.gradle.kts")`.
5. Run `./gradlew help`. If it fails, correct the script before you go to the
   next file.

| Groovy DSL                                   | Kotlin DSL                                                  |
| -------------------------------------------- | ----------------------------------------------------------- |
| `id 'com.android.application'`               | `id("com.android.application")`                             |
| `id 'java'`, `id 'maven-publish'`            | `java`, `` `maven-publish` `` (core plugins)                |
| `apply plugin: 'java'`                       | `apply(plugin = "java")`, or better, the `plugins {}` block |
| `include ':app'`                             | `include(":app")`                                           |
| `ext.foo = 'x'`, `ext { foo = 'x' }`         | `extra["foo"] = "x"`, read with `val foo: String by extra`  |
| `def x = 1`                                  | `val x = 1`                                                 |
| `task hello { ... }`                         | `tasks.register("hello") { ... }`                           |
| `tasks.withType(Test) { ... }`               | `tasks.withType<Test>().configureEach { ... }`              |
| `compileSdk 35`, `minSdkVersion 24`          | `compileSdk = 35`, `minSdk = 24`                            |
| `namespace 'com.acme.app'`                   | `namespace = "com.acme.app"`                                |
| `minifyEnabled true`, `shrinkResources true` | `isMinifyEnabled = true`, `isShrinkResources = true`        |
| `buildTypes { release { ... } }`             | `buildTypes { getByName("release") { ... } }`               |
| `signingConfigs { release { ... } }`         | `signingConfigs { create("release") { ... } }`              |
| `signingConfig signingConfigs.release`       | `signingConfig = signingConfigs.getByName("release")`       |
| `flutter { source '../..' }`                 | `flutter { source = "../.." }`                              |

A Groovy script often reads `local.properties` or `key.properties` with a
`Properties` object. In Kotlin, use `java.util.Properties()` and
`file("key.properties").inputStream().use { props.load(it) }`, then read a key
with `props.getProperty("storePassword")`.

Keep the `.gradle` suffix in each line that is not a script name, for example
the `.gradle/` cache folder in `.gitignore`, and the
`gradle/libs.versions.toml` catalog.

## A script that cannot be migrated

Some old plugins operate only with dynamic Groovy. The Gradle guide lets a
Kotlin script apply such a Groovy script, for example
`apply(from = "legacy-plugin.gradle")`. If you cannot migrate a script:

1. Keep that script in Groovy, and migrate all the other scripts.
2. Tell the user the name of the script and the reason.

The condition of "When the migration applies" finds that script again in the
next update. Tell the user so.

## After the migration

1. Make sure that no Groovy script is left: run the command of "When the
   migration applies" again. Its output must be empty, or only the scripts of
   "A script that cannot be migrated".
2. Run the same build commands as in step 3 of "Before you start". Each one
   must pass, like before the migration.
3. Run `mise run code:lint` and `mise run code:format`, if the repository has
   these tasks.
4. Show the changed files to the user with `git status --short`. Do not
   commit, unless the user tells you to.
