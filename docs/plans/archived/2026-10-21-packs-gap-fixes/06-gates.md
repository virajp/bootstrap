# U6 — Gates: bring this repository in sync

- **Wave:** 4
- **Depends on:** U5
- **Owns:** `.config/bootstrap.toml`, `.config/mise/conf.d/**`, every synced
  copy in `.config/` and at the root
- **Model:** opus
- **Kind:** edit
- **Read first:** `.config/bootstrap.toml`,
  `.config/mise/conf.d/_repo/mise.dev.toml`.

## Ruling

Quoted from `index.md`:

- D5: "This repository selects `[tools.claude]`, moves its `conf.d/_repo/`
  files to `conf.d/zz-repo/`, and removes `"pipx:graphifyy"` from its seed,
  because `claude` supplies it."

## Edits

1. `.config/bootstrap.toml`: add `[tools.claude]` with no `env`.
2. Move `.config/mise/conf.d/_repo/` to `.config/mise/conf.d/zz-repo/` with
   `mv`. In `zz-repo/mise.dev.toml`, remove the `[tools]` line
   `"pipx:graphifyy" = "latest"`, and the `[tools]` table if it is then empty.
3. Set `REPO_SYNC_URL` to the path of the run's worktree and run
   `mise run setup:configs`. Make sure that `using bootstrap at <rev>` shows
   `git rev-parse --short HEAD` of the run branch; if not, report a `GAP:` and
   use a scratch clone of the run branch.
4. Run `mise run init`, then `mise run setup:configs` a second time.

## Verification

- `test -d .config/mise/conf.d/zz-repo && test ! -e .config/mise/conf.d/_repo`
- `.config/pre-commit-config.yaml` has the `graphify-refresh` hook.
- `.config/mise/tasks/code/graph` has the header `tools/claude/...`.
- The second sync shows only `unchanged`, `kept seed` or `skip env`.
- `mise x -- mise run test:repo-sync` and `mise x -- mise run code:precommit`
  pass.

## Guardrails

- Edit only `.config/bootstrap.toml` and `zz-repo/mise.dev.toml` by hand.
  Every other owned file changes only as the output of the sync.
- Do not edit `.config/mise/tasks/test/` or `.config/repo-sync/`.

## Commit

`ops: sync bootstrap after the packs gap fixes`
