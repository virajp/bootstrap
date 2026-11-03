# U3 — Tests: the repo/ layout

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `sync-engine/repo-sync` (U2's output), `repo/tools/` (U1's
  output), the forwarding file.

## Ruling

From the design rulings of 2026-11-04:

> - Placement (accepted): `sync-engine/` holds the shared library and both
>   front-ends (repo-sync, setup-sync). `repo/` holds `tools/`, `packs.toml`
>   (out of `tools/`), `examples/` (with `last_sync.toml`). `macos/` holds
>   `tools/`, `packs.toml`, `fixtures/` (today
>   `macos-setup/tests/setup-sync/`). `docs/` stays at the root:
>   `docs/sync/` (no move; plugins name `docs/sync/vwf.md`), `docs/setup/`
>   (Mac guide: sync.md, tool-reference.md), `docs/design/` (both sides).
>   Test tasks: `.config/mise/tasks/test/repo-sync` and `test/setup-sync`.

D3: "This plan extends test:repo-sync with cases for its new rules. An
existing case changes only where a ruling changes its expected result."

D7: "Each message that names a tool file names the path in the clone,
repo/tools/<tool>/<path>, and repo/packs.toml for the packs. The tests that
compare these messages change their expected text."

## Edits

1. **`.config/mise/tasks/test/repo-sync`**:
   - The bootstrap copy (near line 46): copy `$root/repo`, `$root/sync-engine`
     and `$root/tools` (the forwarding file) into the scratch clone. Update
     the header comment ("the working tree's tools/").
   - Each path of the scratch clone `$clone/tools/<tool>` becomes
     `$clone/repo/tools/<tool>`, and `$clone/tools/packs.toml` becomes
     `$clone/repo/packs.toml`. The source of the engine stays as P01 set it.
   - Each expected message that names `tools/<tool>/...` names
     `repo/tools/<tool>/...` (D7).
   - New cases: a real run (`case_args=' '`) writes a file whose second line
     has `Synced from bootstrap: repo/tools/`; a generated conf.d file has
     `Synced from bootstrap: repo/tools/<tool>/tools.toml`; a consumer file
     with the old header `Synced from bootstrap: tools/...` counts as synced
     (a manifest from P05 lists it, the tool no longer ships it, and the dry
     run has `would remove`); a case that sources the forwarding file
     `$clone/tools/repo-sync/.config/mise/tasks/_scripts/repo-sync` with
     `REPO_SYNC_CLONE` set and `SYNC_ENGINE_CLONE` unset gives exit 0.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes, every old and new case.
- `grep -n '\$clone/tools/' .config/mise/tasks/test/repo-sync` finds only the
  forwarding-file case.

## Guardrails

- Do not touch `sync-engine/**`, `repo/**` or `tools/**`. Report an engine
  bug as `UNRESOLVED:` with the case.
- Write shell for bash 3.2 and bash 5.x. Under `set -e`, write a command
  substitution that can fail as `VAR=$(cmd) || VAR=''`.
- Delete with `rm`, never `git rm`.

## Commit

`test: run test:repo-sync on the repo/ layout` — written by the orchestrator
after the unit gate and the wave gate, not by the unit.
