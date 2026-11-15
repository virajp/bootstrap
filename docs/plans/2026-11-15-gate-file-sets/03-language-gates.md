# U3 — The language gates

- **Read first:** each gate file and the `tools.toml` of each owned tool:
  `repo/tools/flutter/tasks/code-format-dart`, `code-lint-dart`;
  `repo/tools/kotlin/tasks/code-format-ktlint`, `code-lint-detekt`;
  `repo/tools/python/tasks/code-format-ruff`, `code-lint-pyrefly`,
  `code-lint-ruff`, `code-lint-uv`;
  `repo/tools/swift/tasks/code-format-swift-format`,
  `code-lint-swiftlint`; `repo/tools/go/tasks/code-lint-go`, `code-lint-gomod`;
  `repo/tools/android/tasks/code-lint-android`. Then the U1 result:
  `repo/tools/mise/tasks/helpers` (`gate_files`) and
  `repo/tools/mise/tasks/gate_files.py` (read only).
- **Lazy-load:** `repo/tools/swift/files/.config/swiftlint.yml` (read only; a
  seed, D8).

## Ruling

- **D13:** each gate calls `gate_files "$0"` with its flags and files, and
  gives the list file to its tool. An empty list: exit 0 with no output. A gate
  keeps its check that its tool exists (no tool: exit 0 with no output) before
  the call.
- **D14, D26:** detekt keeps its own argument file: the gate reads the list
  file and builds `--input` with the paths, as today (with the skip of a path
  with a comma or whitespace). Each other tool that takes files gets the list
  through `xargs`.
- **D23:** each `xargs` call is `tr '\n' '\0' < "<list file>" | xargs -0
  <tool>`.
- **D24:** right after the `gate_files` call: `trap 'rm -f "<list file>"'
  EXIT`.
- **D50:** detekt gets only the `gate_files` list, not every Kotlin file of
  the repository; D26 keeps only its argument file.
- **D51:** where a gate already sets a `trap … EXIT` (`code-lint-android`,
  `code-lint-detekt`), one trap removes both its temporary folder and the
  list file.
- **D54:** the AndroidManifest check of `code-lint-android` runs in each mode.
- **D57:** change the comment at `repo/tools/python/tools.toml:28` so that it
  also names `python@3.14` in `gate_files` of `helpers`.
- **D31:** a change that only deletes files gives an empty list, so a module
  gate does not run; `--all` covers that case.
- **D7:** the gates drop every own filter, the `git ls-files` calls and
  `--force-exclude`. User: "I am more keen on standardizing the approach so it
  becomes easier to debug".
- **D12 (the headers):**
  - `code-format-dart`, `code-lint-dart`: `#FILES include="*.dart"
    exclude="*.g.dart *.freezed.dart *.mocks.dart"`.
  - `code-format-ktlint`, `code-lint-detekt`: `#FILES include="*.kt *.kts"
    exclude="**/generated/**"`.
  - `code-format-ruff`, `code-lint-pyrefly`, `code-lint-ruff`:
    `#FILES include="*.py *.pyi"`.
  - `code-format-swift-format`, `code-lint-swiftlint`: `#FILES
    include="*.swift" exclude="*.generated.swift .build/** .swiftpm/**
    **/Derived/** **/DerivedData/**"`.
  - `code-lint-go`, `code-lint-gomod`: `#FILES include="*.go go.mod go.sum"`.
  - `code-lint-uv`: `#FILES include="pyproject.toml uv.lock"`.
  - `code-lint-android`: `#FILES` with the Android sources and build files of
    its filter of today (`src/` Kotlin, Java and XML files, `*.gradle.kts`,
    `*.gradle`, `*.pro`, `libs.versions.toml`, `lint.xml`,
    `lint-baseline.xml`), written as patterns.
- **D12 (the module gates):** android, go, gomod and uv change the list into
  modules or projects as today. In the default mode they examine the modules
  of the changed files; with `--all`, each module of the list.
- **D17:** each gate declares `#USAGE flag "--all"`, `--fix`, `--debug` and the
  variadic `[files]`; the `[files]` help says "the default is the uncommitted
  files".
- **Dart:** `dart format` and `dart analyze` get the files of the list, not
  `lib/` and `test/` (D7).
- **D3, D4:** the gate names its lists in its `#FILES` header and gets its
  list from `gate_files`; the orchestrator computes no list.
- **D19:** U2 edits the dprint, linter and ShellCheck gates in the same wave;
  the two units share no file.

## Edits

1. **Each gate file named under Read first** — add `--all` and the `#FILES`
   header; replace the filter code (the `FILES`/`TARGETS` loops, the
   `git ls-files` calls, the `*/generated/*` and `*.generated.swift` checks)
   with one `gate_files` call; give the list to the tool per D14; drop
   `--force-exclude`. Keep each header text, `print_header` text, the `--fix`
   and `--debug` handling, and the "does not apply" checks of today, save the
   ones that only looked for a file of the language (the empty list does
   that now).
2. **The `tools.toml` of each owned tool** — no change, unless a comment there
   describes the old filter; then change the comment. Change the Python
   version comment of `repo/tools/python/tools.toml` (D57).

## Verification

- `mise x -- mise run code:format` passes.
- `find repo/tools/{flutter,kotlin,python,swift,go,android}/tasks -name
  'code-format-*' -o -name 'code-lint-*' | xargs grep -n --
  '--force-exclude\|ls-files\|generated'`
  (D49) finds only the `#FILES` headers and the AndroidManifest check of
  `code-lint-android` (D41: that "does not apply" check stays).
- D44: each `#FILES` line is right after the `#USAGE` lines of its gate.
- `bash -n` passes on each changed task file.

## Guardrails

- Touch nothing outside the six owned tool folders. Do not edit
  `repo/tools/mise/` (U1's); report a needed change there as a QUESTION.
- Do not edit the SwiftLint seed (D8).
- Write shell for bash 3.2 and bash 5.x. Quote the list file path.
- Delete with `rm`, never `git rm`.

## Commit

`feat: the language gates take their files from gate_files`
