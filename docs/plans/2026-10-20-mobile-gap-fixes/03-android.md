# U3 — The android tasks: project paths, Groovy DSL, install checks

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/android/.config/mise/tasks/setup/deps/android`,
  `tools/android/.config/mise/tasks/code/lint/android`,
  `docs/sync/android.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** the three owned files, top to bottom; `CLAUDE.md` sections
  "Task scripts" and "Task output: headers and subheaders".
- **Lazy-load:** `tools/base/.config/mise/tasks/_scripts/helpers` (status
  printers); the Gradle user guide through Context7
  (`/websites/gradle_current_userguide`) for `Project.projectDir` and
  `gradle.projectsEvaluated`.

## Ruling

From index.md, assumed decisions:

> R8 — `code:lint:android` gives the init script the module folders, not
> Gradle paths. The init script finds the project whose `projectDir` is that
> folder. A folder with no project lints every module, which also covers
> `buildSrc/` (G5).

> R9 — After the install, the task examines `$ANDROID_HOME/<id>/package.xml`
> for each package it asked for. If one is missing, it fails and names the id.
> No network.

> R10 — Supported: `build.gradle` and `*.gradle` in the AGP detect, the level
> detect and the lint file filter, and `settings.gradle` in the lint-all list.
> The `compileSdk` and `targetSdk` detects also accept the Groovy form without
> `=` (`compileSdk 35`, `compileSdkVersion 35`).

> R11 — No AGP check in `code:lint:android`. The departures from E6 are
> ratified: `:bootstrapLint` from a temporary init script depends on `lint` or
> `lintFix` only where the task exists, and "every module" means the projects
> below the Gradle root.

> R12 — The AGP and level detects read the file list from
> `git ls-files --cached --others --exclude-standard`, so ignored folders
> (`.worktrees/`, `.claude/worktrees/`, `node_modules`) are not read. The
> installed check also accepts an installed `_ps16k` variant of the image,
> uses it, and prints which id it used, with no network list.

> R13 — `docs/sync/android.md` keeps the version-catalog example (no Flutter
> example, by E2). The package ids use the `/` form of the `android` CLI, and a
> `;` in `EMULATOR_IMAGE` is converted. No change.

## Edits

1. **`setup/deps/android`, detects (:24-28, :47-48)** (R10, R12):
   - Get the file list once from
     `git ls-files --cached --others --exclude-standard` in
     `${MISE_PROJECT_ROOT}`; filter it with `grep` or a `case` for
     `*.gradle.kts`, `*.gradle`, `build-logic/*` and
     `gradle/libs.versions.toml`. Remove the `grep -r` and `find .` walks.
   - The AGP detect and the level detect read `build.gradle.kts`,
     `build.gradle` and `*.gradle` files. Keep "the first hit", in sorted
     order, as now.
   - Extend the `compileSdk` and `targetSdk` patterns to the Groovy forms
     `compileSdk 35`, `compileSdkVersion 35`, `targetSdk 35`,
     `targetSdkVersion 35`, and keep the Kotlin `=` forms.
2. **`setup/deps/android`, installed check (:122-160)** (R12): before the
   network list, when the plain image id is not installed and its `_ps16k`
   variant is (`$ANDROID_HOME/<variant path>/package.xml` exists), use the
   variant, `print_info` which id it used, and do not list. The variant id
   inserts `_ps16k` after the tag, as the existing swap at :153 does.
3. **`setup/deps/android`, after the install (:172)** (R9): for each package
   that the task gave to `android sdk install`, examine
   `$ANDROID_HOME/<id>/package.xml`. If one is missing, `print_error` names
   the id and the task exits 1. Print "installed" only after the check.
4. **`code/lint/android`, `module_of` and the init script (:38-55, :90-104)**
   (R8):
   - `module_of` keeps the walk to the nearest folder with
     `build.gradle.kts` or `build.gradle`, and prints the folder path relative
     to the Gradle root (`.` for the root), not a Gradle path.
   - The init script receives the folders. For each folder, it finds the
     project in `root.allprojects` whose canonical `projectDir` equals the
     canonical `rootDir/<folder>`. A folder with no such project (for example
     `buildSrc`, or an included build) makes the lint run on every module.
   - The lint-all list (:41) adds `settings.gradle` and keeps
     `settings.gradle.kts`, `buildSrc/*`, `build-logic/*` and `gradle/*`.
   - Keep R11: `:bootstrapLint` depends on `lint` or `lintFix` only in each
     found project that has the task; "every module" stays the projects whose
     `projectDir` is under the root dir.
5. **`code/lint/android`, file filter (:67)** (R10): add `*.gradle`.
6. **`docs/sync/android.md`**:
   - :38-39: the detects read `*.gradle.kts` and `*.gradle` (Groovy DSL is
     supported) and only the files that git tracks or does not ignore.
   - :61-63: keep the offline promise; add that an installed `_ps16k` image
     counts as installed.
   - Add one line: an unknown platform or image id that installs nothing
     fails the task and names the id.
   - :71-77: the lint maps a folder to the project with that `projectDir`, so
     a remapped `projectDir` in `settings.gradle(.kts)` works; a folder with
     no project lints every module; add `settings.gradle` to the lint-all
     list.
   - Do not change the version-catalog example or the `/` id text (R13).

## Verification

- Wave-gate lines 1 and 2 pass (ShellCheck runs in line 1).
- `bash -n` on both tasks passes.
- `grep -n 'grep -r\|find \.'
  tools/android/.config/mise/tasks/setup/deps/android`
  finds no file walk.
- With `MISE_PROJECT_ROOT` set to an empty temporary git repository, each task
  prints nothing and exits 0.

## Guardrails

- Do not touch `tools/android/tools.toml`, any other tool, or the docs that U6
  owns. Report each falsified passage as `DOCS FALSIFIED:`.
- Do not install SDK packages on this host to test; the review row and the
  wave review judge the change.
- Keep the scripts safe for bash 3.2 and 5.x; write `VAR=$(cmd) || VAR=''`
  under `set -e`.
- Delete with `rm`, never `git rm`.

## Commit

`feat(android): support Groovy DSL and remapped projects, check installs`
