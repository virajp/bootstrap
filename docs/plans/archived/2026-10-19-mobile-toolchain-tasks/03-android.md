# U3 — The Android SDK and lint tasks of the android tool

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/android/.config/mise/tasks/setup/deps/android` (new, mode
  `755`), `tools/android/.config/mise/tasks/code/lint/android` (new, mode
  `755`), `docs/sync/android.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** `tools/flutter/.config/mise/tasks/setup/deps/flutter` (the
  model deps task; its Android step at :64-70),
  `tools/flutter/.config/mise/tasks/code/lint/dart` (the model lint task:
  `--fix`, `--debug`, `[files]`), `docs/sync/android.md`, `CLAUDE.md`
  sections "Task scripts" and "Task output: headers and subheaders".
- **Lazy-load:** `tools/base/.config/mise/tasks/_scripts/helpers`,
  `tools/base/.config/mise/tasks/code/lint/_default.tmpl:57-59` (how the lint
  dispatcher passes flags and files), `mise x -- android sdk --help` (the CLI
  of the `android-sdk` tool, to find the list and install commands).

## Ruling

From index.md, assumed decisions:

> E1 — One task `setup:deps:<tool>` per tool, with a `--upgrade` flag, as
> `setup:deps:flutter`. A normal run installs only what the lockfiles name. No
> audit step: `code:sec` runs osv-scanner and grype over the repository.

> E2 — `setup:deps:*` acts on the repository root only (`./gradlew`,
> `Package.swift`, one `*.xcodeproj`). In a Flutter repository,
> `setup:deps:flutter` keeps the `android/` and `ios/` deps, so nothing runs
> twice. `code:lint:android` uses the root `./gradlew`, else
> `android/gradlew`, so a Flutter app's Android code is linted too.

> E5 — Applies when the root has an executable `./gradlew` and a build file
> applies the Android Gradle plugin (`com.android.application` or
> `com.android.library`, in `*.gradle.kts`, `build-logic/` or
> `gradle/libs.versions.toml`); else it prints nothing and exits 0. The level:
> `COMPILE_SDK` from the env when set, else the first hit of `compileSdk` in
> `build.gradle.kts` files, then `build-logic/`, then
> `gradle/libs.versions.toml` (excluding `build/`, `.gradle/`); no hit fails
> and tells the user to set `COMPILE_SDK` in `conf.d/_repo`. A bare level of
> 37 or more becomes `<level>.0`; the level must match
> `^[0-9]+(\.[0-9]+)?$`. The image: `EMULATOR_IMAGE` from the env when set,
> else `system-images;android-<target>;google_apis` with `<target>` detected
> the same way from `targetSdk`. ABI `arm64-v8a` on arm64/aarch64, else
> `x86_64`. G7: if the image id is not in the package list and its `_ps16k`
> variant is, install the variant and print the id used; if neither is, stop
> and name both. Packages: `platform-tools`, `platforms;android-<lvl>`,
> `emulator`, the image with the ABI — installed through the `android-sdk`
> tool's CLI (`mise x -- android sdk install …`, as the flutter task does),
> which writes `licenses/`.

> E6 — Applies when the root or `android/` has an executable `gradlew` and
> `$ANDROID_HOME/platforms` exists; else it prints nothing and exits 0. With a
> file list, it runs only when a file matches `*src/*.kt`, `*src/*.java`,
> `*src/*.xml`, `*.gradle.kts`, `*.pro`, `*libs.versions.toml`, `*lint.xml` or
> `*lint-baseline.xml`; it maps each matching file to its Gradle module (the
> nearest folder with `build.gradle.kts` or `build.gradle`, below the Gradle
> root) and runs `:<module>:lint` for each module (`:<module>:lintFix` with
> `--fix`). A file that maps to the Gradle root or to
> `gradle/libs.versions.toml`, `settings.gradle.kts` or `build-logic/` lints
> every module. With no file list, `lint` (or `lintFix` with `--fix`) for
> every module. `--debug` adds `--info`.

> E7 — `android-sdk` stays at `"23"`: bootstrap owns the pin, and mise takes
> the newest 23.x.

> E10 — Each task unit owns its `docs/sync/<tool>.md` (the "Tasks" table and
> "What the tool ships"), as D12 of plan 1. Tool calls go through `mise x --`;
> host commands (`git`, `awk`, `sed`, `find`, `uname`, `swift`, `xcodebuild`)
> do not.

G7 text (claude-plugins, 2026-10-18): "The install task checks
`EMULATOR_IMAGE` with `sdkmanager --list`. If the id is missing and its
`_ps16k` variant exists, it installs that and prints which id it used. If
neither exists, it stops and names both ids."

## Edits

1. **`tools/android/.config/mise/tasks/setup/deps/android`** (new, `755`):
   - shebang, the bootstrap header for this path, `#MISE` and `#USAGE`
     headers in the flutter task's style, with `--upgrade` (accepted; an
     upgrade installs the same packages — the SDK packages have no lockfile);
   - `set -euo pipefail`, source the helpers, `cd "${MISE_PROJECT_ROOT}"`;
   - the applicability check of E5 → exit 0, no output;
   - `print_header` once, `print_subheader` per step (detect, list, install);
   - the detects of E5, in the order given; each with
     `grep -rhoE … --exclude-dir=build --exclude-dir=.gradle` style commands,
     bash 3.2 safe; print the detected level, target and source file with
     `print_info`;
   - the package list: find the list command of the `android` CLI of the
     `android-sdk` tool (`mise x -- android sdk --help`); if it has none, use
     `mise x -- sdkmanager --list` (cmdline-tools is on `PATH` through the
     tool). Record which one you used as a `DECIDED:` line;
   - the G7 fallback, then one install call with the four packages, through
     the same CLI the flutter task uses (`mise x -- android sdk install …`);
   - exit 1 with `print_error` on a failed detect, a bad level, or a missing
     image.
2. **`tools/android/.config/mise/tasks/code/lint/android`** (new, `755`):
   - shebang, header, `#MISE` and `#USAGE` in `code/lint/dart`'s style:
     `--fix`, `--debug`, `[files]` (variadic);
   - the Gradle root: `.` when `./gradlew` is executable, else `android`
     when `android/gradlew` is; none, or no `$ANDROID_HOME/platforms` → exit
     0, no output;
   - the E6 filter and the module map: for each matching file under the
     Gradle root, walk up from its folder to the Gradle root and take the
     first folder with `build.gradle.kts` or `build.gradle`; turn its path
     relative to the Gradle root into a Gradle path (`app/feature` →
     `:app:feature`); a file at the root, in `gradle/`, `settings.gradle.kts`
     or `build-logic/` → every module. Remove duplicates (no associative
     arrays: a sorted list with `sort -u`);
   - run `(cd <root> && ./gradlew [--info] <tasks…>)` once with all module
     tasks;
   - `print_header` once; nothing printed when nothing matches.
3. **`docs/sync/android.md`**: "What the tool ships" gets both task rows;
   "Tasks" describes `setup:deps:android` (detects, env overrides, the image
   and its `_ps16k` fallback, the packages) and `code:lint:android` (Gradle
   root, trigger files, changed modules only, `--fix`); "Adapt to the
   repository" says how to set `COMPILE_SDK` or `EMULATOR_IMAGE` in
   `.config/mise/conf.d/_repo/mise.<env>.toml` when the detect cannot find
   them (for example a Flutter app's `flutter.compileSdkVersion`).

## Verification

- Wave-gate lines 1 and 2 pass.
- `bash -n` on both tasks passes; the unit sets `chmod 755` on both.
- With `MISE_PROJECT_ROOT` set to an empty temporary git repository, each
  task prints nothing and exits 0.
- `grep -n 'lint-baseline.xml'
  tools/android/.config/mise/tasks/code/lint/android`
  prints a line.

## Guardrails

- Do not touch `tools/android/tools.toml`, any other tool, or the docs that U5
  owns. Report each falsified passage as `DOCS FALSIFIED:`.
- Do not change `tools/flutter/**`; the Flutter `android/` deps stay there.
- Do not add a `test:*` task.
- No associative arrays, namerefs or `${var,,}` (bash 3.2).
- Delete with `rm`, never `git rm`.

## Commit

`feat(android): add setup:deps:android and code:lint:android`
