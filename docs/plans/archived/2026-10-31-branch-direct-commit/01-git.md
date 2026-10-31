# U1 — git: neutral hook text and the safety-net skip

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:**
  `tools/git/.config/mise/tasks/code/merge/{_default,develop,main}`
  (the callers of the library),
  `tools/mise/.config/mise/tasks/_scripts/helpers`.

## Ruling

D3: "The safety net always runs with SKIP=no-commit-to-branch. The hook guards
a commit, and the safety net does not make a commit; a merge never runs the
hook, because the hook types do not include pre-merge-commit."

D6: "Neutral fixed text, for example \"Refuse a direct commit to a protected
branch\". The --branch args show which branches the hook protects."

## Edits

1. **`tools/git/tools.toml`** — in the `[[precommit.hooks]]` table of
   `no-commit-to-branch` (`:25-29`), change the `name` (`:29`) from
   `Refuse a direct commit to main or develop` to
   `Refuse a direct commit to a protected branch`. The `description` (`:27`)
   names no branch ("These branches take merges only. [...]"), so it stays.
   Do not change the `id`, the `args` or any other table.
2. **`tools/git/.config/mise/tasks/_scripts/merge`** — in
   `precommit_safety_net` (`:63-108`), run the `pre-commit run --all-files`
   call (`:78`) with `no-commit-to-branch` added to `SKIP`, and keep a value
   that the caller already set: `SKIP="${SKIP:+${SKIP},}no-commit-to-branch"`
   as a prefix of that one command, not an `export`. Add a short comment in
   the style of the file: the hook guards a commit, the safety net makes none,
   and `code:merge:main` runs from `develop`, which the hook can protect.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'direct commit to main or develop' tools/git/tools.toml` finds
  nothing.
- `grep -n 'no-commit-to-branch' tools/git/.config/mise/tasks/_scripts/merge`
  finds the `SKIP` prefix of the safety-net call.
- `bash -n tools/git/.config/mise/tasks/_scripts/merge` passes.

## Guardrails

- Do not touch the synced copy `.config/mise/tasks/_scripts/merge` of this
  repository; the next sync replaces it.
- Do not touch `tools/repo-sync/` or the template; U2 owns the args.
- Write shell for bash 3.2 and bash 5.x.
- Delete with `rm`, never `git rm`.

## Commit

`fix: skip no-commit-to-branch in the merge safety net` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
