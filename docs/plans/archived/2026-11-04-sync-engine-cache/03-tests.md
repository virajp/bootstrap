# U3 — Tests for the new cache and the moved library

- **Read first:** `.config/mise/tasks/test/repo-sync`, top to bottom, before
  editing.
- **Lazy-load:** `sync-engine/clone`, `sync-engine/repo-sync`, the forwarding
  file `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`, and the three
  tasks of `tools/repo-sync/.config/mise/tasks/` (the output of U1 and U2).

## Ruling

From the design rulings of 2026-11-04:

> - One shared cache for repos and the Mac: the clone in a `repo/` subfolder,
>   the lock and `macos-generated.tsv` next to it. Origin compare is
>   normalized (repos use SSH, the Mac uses HTTPS). Branch `main` explicit. A
>   non-git, non-empty clone folder stops with exit 2 (no delete). A dry run
>   releases the lock early.

> - The new engine treats a clone that is not in `repo/` (the old layout, the
>   clone at the cache root) as an old clone and removes it before it clones
>   into `repo/`.

> - The shared cache folder is `~/.cache/sync-engine/` (`repo/`, `lock/`,
>   `macos-generated.tsv`), not `~/.cache/bootstrap/` and not `bootstrap2`.
>   **Why:** an old consumer stub runs once more after the update; on a new
>   layout in `~/.cache/bootstrap` it does `rm -rf` and deletes the manifest,
>   and its lock `~/.cache/bootstrap.lock` does not exclude the new lock. A
>   version number in the name has no meaning after the transition. Each run
>   removes the legacy `~/.cache/bootstrap/`, `~/.cache/bootstrap.lock` and
>   `~/.cache/macos-setup/` (not while an old stub holds the old lock). The
>   first run moves `~/.cache/macos-setup/generated.txt` to
>   `macos-generated.tsv` before that removal.

Note: D12 of this plan narrows the last two sentences. This plan removes only
`~/.cache/bootstrap/` and `~/.cache/bootstrap.lock`; plan 10 handles
`~/.cache/macos-setup/`.

> - Engine variables: `SYNC_ENGINE_CLONE` (temp copy, stub → child task),
>   `SYNC_ENGINE_URL`, `SYNC_ENGINE_CACHE`. Internal only, not user config:
>   test:repo-sync and test:setup-sync use them (today `REPO_SYNC_URL` 8x,
>   `MACOS_SETUP_REPO_URL` 5x, `MACOS_SETUP_REPO_PATH` 3x), and
>   `SYNC_ENGINE_URL`
>   tests an unpushed change. The user saw no value in them and let me keep
>   them on this reason. The forwarding file at the old repo-sync path accepts
>   `REPO_SYNC_CLONE` during the transition.

> - No local engine copy in a consumer. `setup:configs:sync` without
>   `SYNC_ENGINE_CLONE` stops with exit 2. The gate `code:lint:repo-sync`
>   sources the engine from `~/.cache/sync-engine/repo/`; no cache → silent
>   exit 0, never a clone.

From this plan:

D3: "This plan extends test:repo-sync with cases for its new rules. A current
case changes only where a ruling changes its expected result."

D6: "The stub template embeds sync-engine/clone at render time with the
gomplate function file.Read. The repo front-end adds the absolute path of the
sync-engine/ folder of its clone copy to the template context as
.sync_engine. One source, no copy in tools/."

D7: "REPO_SYNC_CLONE becomes SYNC_ENGINE_CLONE and REPO_SYNC_URL becomes
SYNC_ENGINE_URL. REPO_SYNC_VALUES and REPO_SYNC_NO_AUTOUPDATE stay, because
they are variables of the repo front-end. The internal gomplate variables of
the library (REPO_SYNC_TOOLS, REPO_SYNC_ENVS and the others) stay."

D8: "No. The self-clone path goes. repo_sync without SYNC_ENGINE_CLONE and
without --check returns 2 with a message that names mise run setup:configs."

D9: "Each function of sync-engine/clone starts with *se* and each global
variable with *SE*. The later shared libraries of sync-engine/ use the same
prefix. The library defines only functions; _se_clone_init sets the paths
from the environment when the caller calls it."

D10: "Through _se_say <info|error> <text>. When the caller defined print_info
and print_error (the helpers of the repo side), _se_say uses them. Else it
prints 'sync-engine: <text>' on stderr. The Mac entry task of a later plan has
no helpers."

D11: "Both SYNC_ENGINE_CLONE and REPO_SYNC_CLONE, set to the same copy, for
the root and for each member, until the parked cleanup plan. A member that has
not synced yet has the old setup/configs/sync, which reads only
REPO_SYNC_CLONE and then sources the forwarding file."

D12: "Each run, a dry run too, calls _se_legacy_cleanup while it holds the new
lock. For ${XDG_CACHE_HOME:-$HOME/.cache}/bootstrap it first takes the old
lock bootstrap.lock (mkdir, with the stale-lock rule of today for a dead
holder); when a live pid holds it, the run skips the removal and says nothing.
With the old lock taken, it removes bootstrap/ and then bootstrap.lock. Once
the run holds the old lock, it releases it on success, on failure and on EXIT:
_se_clone_cleanup, which the trap of the caller calls, releases it too. This
plan does not touch ~/.cache/macos-setup/: the old Mac task can still run
until plan 10, and the Mac entry task of plan 10 moves generated.txt to
macos-generated.tsv and removes that folder."

D15: "It prints 'repo-sync: no clone - run mise run setup:configs' on stderr
and returns 2. It never clones and never reads the cache."

## Edits

1. **`.config/mise/tasks/test/repo-sync` — the setup.**
   - The scratch clone (:46-50) copies `sync-engine/` next to `tools/`.
   - The header comment (:6-9) names `sync-engine/repo-sync` and the
     working tree's `tools/` and `sync-engine/`.
   - `case_body`: `unset SYNC_ENGINE_URL`; `CASE_SYNC_ENGINE_URL` replaces
     `CASE_REPO_SYNC_URL` and becomes `SYNC_ENGINE_URL`; export
     `SYNC_ENGINE_CLONE=$clone` in place of `REPO_SYNC_CLONE`; unset
     `SYNC_ENGINE_CLONE` for `--check`; source `$clone/sync-engine/repo-sync`.
     Rename the variable in the comment above `run_case` and in the case
     "--check, no clone" (:683).
   - Add a case variable, for example `CASE_LIB=forward`, with which
     `case_body` exports only `REPO_SYNC_CLONE=$clone` (no
     `SYNC_ENGINE_CLONE`) and sources the forwarding file
     `$clone/tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`.
2. **New `repo_sync` cases** (in the parallel block, before the collect
   loop):
   - "the forwarding file with REPO_SYNC_CLONE": `CASE_LIB=forward`, a valid
     values file, exit 0.
   - "the forwarding file without a clone": no clone variable, exit 2, the
     message of D15.
   - "repo_sync without SYNC_ENGINE_CLONE": exit 2, the message names
     `mise run setup:configs`.
   - "the template context has sync_engine": a change function adds a
     template to a tool of the scratch clone that prints
     `<% .sync_engine %>`; the dry run output has `<clone copy>/sync-engine`
     (match the `/sync-engine` end of the line). Use a tool of the pack
     `base`, and give the new template the bootstrap header.
3. **The `setup:configs` cases** (:812-972).
   - `ctx.json` gets `"sync_engine":"<root>/sync-engine"`.
   - `cache="$tmp/cache/sync-engine"`; the clone is `$cache/repo`; the lock
     is `$cache/lock`. The fake `mise` checks `$SYNC_ENGINE_CLONE/.git`,
     `$SYNC_ENGINE_CLONE` outside `$XDG_CACHE_HOME`, no
     `$XDG_CACHE_HOME/sync-engine/lock`, and that `$REPO_SYNC_CLONE` equals
     `$SYNC_ENGINE_CLONE` (D11). It records the commit of
     `$SYNC_ENGINE_CLONE`.
   - `run_configs` sets `SYNC_ENGINE_URL` in place of `REPO_SYNC_URL`, and
     unsets `SYNC_ENGINE_CACHE`.
   - `check_configs` and the live-holder case use `$cache/lock`, and grep
     the new wait message `Waiting for another sync (pid <pid>) ...`. The
     failed-fetch case checks `$cache/repo/.git`.
4. **New `setup:configs` cases**, each with its own `XDG_CACHE_HOME` folder
   under `$tmp` (or a clean-up of the folders between cases):
   - "removes the old bootstrap cache": `bootstrap/` with a file and no
     `bootstrap.lock`; exit 0; both are gone after the run.
   - "keeps the old cache of a live old stub": `bootstrap/` and
     `bootstrap.lock/pid` with the pid of a live `sleep`; exit 0 without a
     wait; `bootstrap/` stays. Kill the `sleep` after the case.
   - "leaves the old Mac cache": `macos-setup/generated.txt` exists; exit 0;
     `macos-setup/` and its file stay, and no `macos-generated.tsv` exists
     (D12).
   - "releases the old lock on a failure": `bootstrap/` and no old lock, and
     a URL that does not exist; exit 5; `bootstrap.lock` does not exist
     after the run (D12).
   - "a normalized origin is the same clone": a first run with
     `file://<remote>`, a second run with the plain path `<remote>`; the
     second log has `Updating the clone` and no `Cloning into`.
   - "a non-git folder stops with 2": `sync-engine/repo/` with one plain
     file; exit 2; the file is still there.
   - "an old clone at the cache root": a `git init` in `sync-engine/`
     itself (no `repo/`); exit 0; `sync-engine/.git` is gone and
     `sync-engine/repo/.git` exists.
5. **New gate cases** for
   `tools/repo-sync/.config/mise/tasks/code/lint/repo-sync`,
   run with `bash` in a scratch consumer that has
   `.config/mise/tasks/_scripts/helpers` and a valid values file, with
   `MISE_PROJECT_ROOT` set to it and `SYNC_ENGINE_CACHE` set to a scratch
   folder:
   - "the gate without a cache": exit 0 and an empty output.
   - "the gate with a cache": `$SYNC_ENGINE_CACHE/repo/sync-engine/repo-sync`
     copied from the working tree; exit 0 and the `valid (local rules)` line.
     When the gate exits 0 early because `mise x -- which gomplate` finds no
     gomplate in the scratch folder, run it from a folder where it finds
     gomplate, or report a `GAP:` with the assumption taken.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes: every current case and every
  new case.
- `grep -n 'REPO_SYNC_URL\|cache/bootstrap\|bootstrap.lock'
  .config/mise/tasks/test/repo-sync`
  finds only the lines of the legacy-cleanup cases.

## Guardrails

- Do not touch `sync-engine/**` or `tools/**`; U1 and U2 own them. Report a
  defect in them as a `GAP:` or an `UNRESOLVED:`.
- A current case changes only in its setup (paths, variable names, the wait
  message). Its expected code and message stay (D3).
- The test needs no network and writes nothing outside its temp folder. A
  live `sleep` that a case starts must end before the task ends.
- Write shell for bash 3.2 and bash 5.x.
- Delete with `rm`, never `git rm`.

## Commit

`test: cover the sync-engine cache, the legacy cleanup and the forwarding file`
— written by the orchestrator after the unit gate and the wave gate, not by
the unit.
