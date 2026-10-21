# U1 — repo-sync: each tool in dev, and zz-repo sorts last

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
- **Model:** opus
- **Kind:** edit
- **Read first:** the owned file, top to bottom.

## Ruling

Quoted from `index.md`:

- D3: "Rename `conf.d/_repo/` to `conf.d/zz-repo/`. A tool folder whose name
  sorts after `zz-repo` stops the sync with code 2."
- D4: "'yes, every tool must be in dev'. Each `env` of the values file and
  each `default_envs` without `dev` stops the sync with code 2. The pre-commit
  code does not change."

## Edits

1. Where the sync reads the `env` of a `[tools.*]` or `[packs.*]` entry, stop
   with code 2 when the list has no `dev`. The message names the entry and
   says that each tool must be in `dev`, for example `env = ["dev", "ci"]`.
2. Where the sync reads `default_envs` of a `tools.toml`, stop with code 2
   when the list has no `dev`. The message names `tools/<tool>/tools.toml`.
3. Where the sync reads the tool folders of the clone, stop with code 2 for a
   folder whose name sorts after `zz-repo` in byte order (`LC_ALL=C`). The
   message names the folder and says that `conf.d/zz-repo/` must load last.
4. Update the header comments for both rules.

## Verification

- `bash -n` passes; `mise x -- shellcheck -x -s bash` shows no new warning on
  the changed lines.
- `mise x -- mise run code:precommit` and `mise x -- mise run test:repo-sync`
  pass after wave 1, with U3.

## Guardrails

- Touch only the owned file. Do not change the pre-commit code (D4).
- Write shell for bash 3.2 and 5.x.
- A failure must still leave every existing file unchanged.

## Commit

`feat: require dev in each tool and keep zz-repo last in repo-sync`
