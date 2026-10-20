# U2 — setup:deps:kotlin enforces the locks with STRICT

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/kotlin/.config/mise/tasks/setup/deps/kotlin`,
  `docs/sync/kotlin.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** `tools/kotlin/.config/mise/tasks/setup/deps/kotlin` and
  `docs/sync/kotlin.md`, top to bottom; `CLAUDE.md` sections "Task scripts"
  and "Task output: headers and subheaders".
- **Lazy-load:** `tools/base/.config/mise/tasks/_scripts/helpers` (the status
  printers, :163-237); the Gradle user guide on dependency locking, through
  Context7 (`/websites/gradle_current_userguide`), before you write the
  init-script change.

## Ruling

From index.md, assumed decisions:

> R4 — The init script sets `LockMode.STRICT` on each project. The normal run
> resolves without `--write-locks`, so Gradle pins the locked versions and
> fails when a locked configuration has no lock state (a new module). The git
> compare, the RESOLVED and UNCOMMITTED split and the restore leave the normal
> run. `--upgrade` does not change (`--write-locks --refresh-dependencies`).

> R5 — The init script resolves only the configurations that are
> `canBeResolved` and not `canBeConsumed`. A failure in what remains still
> fails the task.

> R6 — A project with no `dependencyLocking` is not checked. Ratified;
> `docs/sync/kotlin.md:83-86` already says so. The warning for a build with no
> lockfile stays.

> R7 — The init script copies the `matching {}` view to a list before it
> resolves. The hints use status printers (`print_info` or `print_warn`), not
> `print_yellow`.

## Edits

1. **`tools/kotlin/.config/mise/tasks/setup/deps/kotlin`, init script
   (:77-85)**:
   - Resolve `configurations.matching { it.canBeResolved && !it.canBeConsumed }`
     copied to a list (`.toList()`) before the loop (R5, R7).
   - For the normal run only, set the lock mode of each project to
     `LockMode.STRICT` before any configuration resolves (for example in
     `allprojects { … }` on `dependencyLocking.lockMode`). Use the API form that
     the Gradle docs give for the wrapper version of the build; a script that
     must work on Gradle 8 and 9 uses the form that both accept. `--upgrade`
     does not set STRICT (a new module must get its first lockfile there).
     Two init-script variants, or one script with a system property that the
     task passes, are both acceptable; record the choice as `DECIDED:`.
2. **Same file, normal run (:125-187)**:
   - Run `./gradlew --init-script <tmp> bootstrapResolveAll` without
     `--write-locks`.
   - Remove the lockfile snapshot, the git compare, the RESOLVED and
     UNCOMMITTED split, the restore and its trap (:125-159), and the hints that
     only they used. Remove `print_files` if nothing else uses it.
   - On a non-zero Gradle exit: `print_error` that the resolve does not match
     the lockfiles, then the fix as status lines: run
     `mise run setup:deps:kotlin --upgrade` and commit the lockfiles. Exit 1.
     Gradle's own output names the project and the configuration; do not hide
     it.
   - Keep the warning and exit 0 for a build with no `*gradle.lockfile` (R6,
     :184-187), and keep the checksum check (:31-37).
3. **Same file, all hints**: replace `print_yellow` at :34, :177 and :181 (or
   their new places) with `print_info` or `print_warn`, under a header or
   subheader (CLAUDE.md "Task output" rule 4).
4. **Same file, `--upgrade` (:89-118)**: no behaviour change, except the shared
   init-script changes of edit 1 (filter and list copy, no STRICT).
5. **`docs/sync/kotlin.md`**:
   - Normal run (:49-55): Gradle enforces the committed lockfiles in STRICT
     mode, set by the task; a dynamic version stays at its locked version; a
     locked configuration with no lock state (a new module) fails; the task
     changes no file.
   - Lockfiles only with `dependencyLocking` (:56-58) and step 3 (:83-86): no
     change of meaning (R6).
   - Add one line: the task resolves only the resolvable-only configurations
     (not `canBeConsumed`), so legacy `default` and `archives` are not
     resolved.

## Verification

- Wave-gate lines 1 and 2 pass (ShellCheck runs in line 1).
- `bash -n` on the task passes.
- `grep -n 'print_yellow\|--write-locks'
  tools/kotlin/.config/mise/tasks/setup/deps/kotlin`
  shows `--write-locks` only in the `--upgrade` path and no `print_yellow`.
- With `MISE_PROJECT_ROOT` set to an empty temporary git repository, the task
  prints nothing and exits 0.

## Guardrails

- Do not touch `tools/kotlin/tools.toml`, any other tool, or the docs that U6
  owns. Report each falsified passage as `DOCS FALSIFIED:`.
- Do not write or commit any lockfile in this repository.
- Keep the script safe for bash 3.2 and 5.x; write `VAR=$(cmd) || VAR=''`
  under `set -e`.
- Delete with `rm`, never `git rm`.

## Commit

`fix(kotlin): enforce the Gradle locks with STRICT in setup:deps:kotlin`
