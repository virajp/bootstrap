# U5 — Docs: zz-repo, the dev rule and the first sync

- **Wave:** 3
- **Depends on:** U4
- **Owns:** `docs/sync/**`, `README.md`, `CLAUDE.md`, `docs/design/*.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** `docs/sync/index.md`, `docs/sync/migrate.md`, `CLAUDE.md`.

## Ruling

Quoted from `index.md`:

- D1: "The AI copies the sync files by hand, then asks the user to run
  `mise run setup:configs`."
- D3: "Rename `conf.d/_repo/` to `conf.d/zz-repo/`. A tool folder whose name
  sorts after `zz-repo` stops the sync with code 2."
- D4: "Each `env` of the values file and each `default_envs` without `dev`
  stops the sync with code 2."

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings plus each
   `DOCS FALSIFIED:` line of U1 to U3.
2. Change `_repo` to `zz-repo` in each doc that the facts of `index.md` list,
   and say why the name sorts last.
3. `docs/sync/index.md` (step 2 and step 6) and `docs/sync/migrate.md`: the AI
   copies the sync files by hand, then asks the user to run
   `mise run setup:configs`. `migrate.md` gets an item: rename an existing
   `conf.d/_repo/` to `conf.d/zz-repo/` before the first sync.
4. `CLAUDE.md`, `docs/design/mise.md`, `docs/design/bootstrap-config.md`: each
   `env` and each `default_envs` must include `dev`; a tool folder must not
   sort after `zz-repo`.

## Verification

- `rg -n 'conf\.d/_repo|_repo/|`_repo`' README.md CLAUDE.md docs/sync
  docs/design`
  prints nothing.
- `mise x -- mise run code:precommit` passes.

## Guardrails

- Touch only docs. Do not edit `docs/plans/**`.
- Write in ASD-STE100 Simplified Technical English, as the existing docs do.

## Commit

`docs: describe zz-repo, the dev rule and the first sync`
