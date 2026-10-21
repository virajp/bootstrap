# U6 — Gates: sync this repository and pass the final gate

- **Wave:** 6
- **Depends on:** U5
- **Owns:** `.config/bootstrap.toml`; every synced or generated copy in
  `.config/` (including `.config/mise/conf.d/**`, `.config/mise/tasks/**`
  except `test/`, `.config/pre-commit-config.yaml`) and `.gitignore`
- **Model:** opus
- **Kind:** edit
- **Read first:** `.config/bootstrap.toml`, `docs/sync/index.md` step 6 and
  step 8.

## Ruling

Quoted from `index.md`:

- D2: "`[tools.<folder>]` and `[packs.<name>]` tables."
- D7: "The `base` pack name is fixed in `repo-sync` and is always added. Empty
  or absent `[tools]` is valid."
- Gates the orchestrator keeps: "a second `mise run setup:configs` shows only
  `unchanged`, `kept seed` or `skip env`; `.config/mise/conf.d/_base/` is
  gone."

## Edits

1. **`.config/bootstrap.toml`:** remove `tools = [ "base" ]`. This repository
   needs only the `base` pack, so it has no `[tools]` and no `[packs]` entry.
   Keep the other keys.
2. **Sync this repository from the run's commit.** Set `REPO_SYNC_URL` to the
   path of the run's worktree, then run `mise run setup:configs`. Make sure
   that the line `using bootstrap at <rev>` shows `git rev-parse --short HEAD`
   of the run branch. If it shows a different commit, report a `GAP:` and use
   a scratch clone of the run branch as `REPO_SYNC_URL`.
3. Run `mise run init`.
4. Run the same `mise run setup:configs` a second time.

## Verification

- The second sync shows only `unchanged`, `kept seed` or `skip env`.
- `test ! -e .config/mise/conf.d/_base`
- `.config/pre-commit-config.yaml` has the `test-repo-sync` hook.
- `mise x -- mise run test:repo-sync` passes.
- `mise x -- mise run code:precommit` passes (the full wave gate).

## Guardrails

- Edit only `.config/bootstrap.toml` by hand. Every other owned file changes
  only as the output of the sync.
- Do not edit `.config/mise/tasks/test/` or `.config/repo-sync/`: U3 owns
  them.
- Do not bump a version.

## Commit

`ops: sync bootstrap with the split base tools`
