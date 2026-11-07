# U2 — Move the repo-sync library to sync-engine and add the forwarding file

- **Read first:** `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`,
  top to bottom, before you move it.
- **Lazy-load:** `tools/repo-sync/.config/mise/tasks/setup/configs/sync`
  (U1 changes it), `.config/mise/tasks/test/repo-sync` (U3 changes it).

## Ruling

From the design rulings of 2026-11-04:

> - The shared library folder is `sync-engine/` (not `engine/`).

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

> - Placement (accepted): `sync-engine/` holds the shared library and both
>   front-ends (repo-sync, setup-sync). `repo/` holds `tools/`, `packs.toml`
>   (out of `tools/`), `examples/` (with `last_sync.toml`). `macos/` holds
>   `tools/`, `packs.toml`, `fixtures/` (today
>   `macos-setup/tests/setup-sync/`). `docs/` stays at the root:
>   `docs/sync/` (no move; plugins name `docs/sync/vwf.md`), `docs/setup/`
>   (Mac guide: sync.md, tool-reference.md), `docs/design/` (both sides).
>   Test tasks: `.config/mise/tasks/test/repo-sync` and `test/setup-sync`.

From this plan:

D5: "It becomes the forwarding file and stays until the parked cleanup plan,
because the old consumer stubs and the old setup/configs/sync tasks source
it."

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

D14: "No. The sync never copies a file of sync-engine/ into a consumer, so the
header rule of tools/ does not apply. Each file starts with the bash shebang,
so shell_files_in_scope lints it, and a short comment that names its use."

D15: "It prints 'repo-sync: no clone - run mise run setup:configs' on stderr
and returns 2. It never clones and never reads the cache."

D18: "Mode 755 for every file of sync-engine/ that has a shebang, clone
included, because the check-shebang-scripts-are-executable hook refuses a file
with a shebang and no execute bit."

D19: "U2. It owns tools/repo-sync/tools.toml and changes its first comment
line: the tool ships the setup:configs tasks, the gate and the forwarding
file, not the sync function, which is in sync-engine/repo-sync."

## Edits

1. **Move.** `mkdir -p sync-engine` and
   `mv tools/repo-sync/.config/mise/tasks/_scripts/repo-sync
   sync-engine/repo-sync`.
   The file mode is 755 (D18); keep it. Do not stage.
2. **`sync-engine/repo-sync`**
   - Lines 2-6: remove the two header lines and the two `#MISE` lines (D14).
     Keep the shebang.
   - The usage comment (line 12): `source
     "${SYNC_ENGINE_CLONE}/sync-engine/repo-sync"`.
     Change every comment that names the old path, `REPO_SYNC_CLONE`,
     `REPO_SYNC_URL`, "over SSH" or the self-clone (for example lines
     60-61 and the Environment block at lines 198-208) to the new facts:
     `SYNC_ENGINE_CLONE` is the clone copy that `setup:configs` gives;
     repo_sync reads no URL. Keep the `REPO_SYNC_VALUES` and
     `REPO_SYNC_NO_AUTOUPDATE` lines (D7).
   - Return codes in the header: code 2 also covers "no
     `SYNC_ENGINE_CLONE` without `--check`"; code 5 becomes "the clone of
     `SYNC_ENGINE_CLONE` is not usable, or `pre-commit autoupdate` could not
     resolve the hook revisions".
   - In `repo_sync` (A:1870-1888): remove `local url=...` and the `elif`
     branch that clones. Before the `mktemp -d`, when `SYNC_ENGINE_CLONE` is
     empty: `_rs_log "SYNC_ENGINE_CLONE is not set - run mise run setup:configs,
     which clones bootstrap"`
     and `return 2`. Rename every other `REPO_SYNC_CLONE` to
     `SYNC_ENGINE_CLONE`. The `--check` path (before this code) does not
     change.
   - `_rs_context` (A:1428-1487): add the string key `sync_engine` to the
     context, with the value `<the clone copy of this run>/sync-engine` (the
     folder `$dir` of `repo_sync`). Use the same mechanism as the other keys
     (an environment variable read in the gomplate call). Add the key to the
     list of context keys in the header comment.
   - Do not change any other line: the `tools/` paths, the messages and the
     codes stay.
3. **`tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`** (new, the
   forwarding file, mode 755 as the old file). Content in this order:
   - `#!/usr/bin/env bash`, the bootstrap header for this path (two lines),
     `#MISE description="Forward to sync-engine/repo-sync of the clone"`,
     `#MISE hide=true`.
   - A comment of at most 6 lines: an old `setup/configs/sync` task sources
     this file from the clone with `REPO_SYNC_CLONE`; the file forwards to
     `sync-engine/repo-sync` of the same clone; the cleanup plan removes it.
   - `: "${SYNC_ENGINE_CLONE:=${REPO_SYNC_CLONE:-}}"`; when it is empty, the
     message of D15 on stderr and `return 2`; else
     `export SYNC_ENGINE_CLONE` and
     `source "${SYNC_ENGINE_CLONE}/sync-engine/repo-sync"`, with the
     `# shellcheck source=/dev/null` line above it.
4. **`tools/repo-sync/tools.toml`** — change the first comment line (D19), for
   example `# repo-sync: the setup:configs tasks, the gate code:lint:repo-sync
   and the forwarding file of the sync.`
   Keep every other line.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n sync-engine/repo-sync` and `bash -n
  tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
  pass.
- `grep -n 'REPO_SYNC_CLONE\|REPO_SYNC_URL\|git@github' sync-engine/repo-sync`
  finds nothing.
- `grep -n 'sync_engine' sync-engine/repo-sync` finds the context key.
- `test -x sync-engine/repo-sync` passes.
- `head -n 1 tools/repo-sync/tools.toml` does not say "the sync function".
- `grep -c 'Synced from bootstrap' sync-engine/repo-sync` is 0; the same
  count for the forwarding file is 1.
- `test:repo-sync` is U3's: it runs only after U3 has changed the test task.

## Guardrails

- Do not touch `sync-engine/clone` or the tasks of `tools/repo-sync/`; U1
  owns them.
- Do not change a rule, a message or a code of the sync other than the ones
  above. The current test cases must pass after U3 changes only their setup.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays.
- Delete with `rm`, never `git rm`; move with plain `mv`.

## Commit

`refactor: move the repo-sync library to sync-engine/repo-sync` — written by
the orchestrator after the unit gate and the wave gate, not by the unit.
