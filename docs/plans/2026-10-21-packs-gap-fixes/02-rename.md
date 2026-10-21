# U2 — Rename the repository seed folder to zz-repo

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/mise/.config/mise/conf.d/_repo/` (moves to
  `tools/mise/.config/mise/conf.d/zz-repo/`), `tools/mise/tools.toml`,
  `tools/mise/.config/mise.toml.tmpl`
- **Model:** opus
- **Kind:** edit
- **Read first:** the owned files.

## Ruling

Quoted from `index.md`:

- D3: "Rename `conf.d/_repo/` to `conf.d/zz-repo/`."

## Edits

1. Move `tools/mise/.config/mise/conf.d/_repo/` to
   `tools/mise/.config/mise/conf.d/zz-repo/` with `mv`. Keep the content of
   the two seeds.
2. `tools/mise/tools.toml`: change the two `seed` paths to `zz-repo`.
3. `tools/mise/.config/mise.toml.tmpl`: change the comment at line 21 to
   `.config/mise/conf.d/zz-repo/mise.<env>.toml`.

## Verification

- `rg -n 'conf\.d/_repo|_repo/' tools/` prints nothing.
- `mise x -- mise run code:precommit` passes after wave 1.

## Guardrails

- Touch only the owned files. Move with `mv`, never `git mv`.

## Commit

`refactor: rename the repository seed folder to zz-repo`
