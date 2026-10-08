# U5 — Renderer: create-only files and selection-dependent renders

- **Wave:** 4
- **Depends on:** U3, U4
- **Owns:** `cli/src/tool/render.ts`, `cli/test/tool/render.test.ts`,
  `cli/templates/mise/.config/mise/conf.d/_base/mise.dev.toml`,
  `cli/templates/pre-commit/.config/git-conventional-commits.yaml`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/tool/render.test.ts`:
  - The render input is the values, the selection (tool names) and the `origin`
    host (or none). Each rendered file carries `path`, `content`, `executable`
    and `createOnly`; `mempalace.yaml` is `createOnly`.
  - `conf.d/_base/mise.dev.toml` holds node, pnpm, python, uv, jq and yq, plus
    the entries of the selected base tools only (pre-commit; dprint and
    sort-package-json; taplo; the house linter). With `dprint` not selected, no
    dprint entry is rendered.
  - `git-conventional-commits.yaml`: with `github` and `gitlab` selected and
    host `gitlab.com`, the links are GitLab's; with host `example.org`, the
    links are GitHub's (the first selected forge in catalog order); with no
    forge selected, the file has no forge links.
  - Every file under `.config/mise/tasks/` is executable; no other file is.
  - Fails because the render takes no selection or host and has no `createOnly`.
- **Read first:** `docs/blueprint/entities/tool/index.md` (Selection-dependent
  files, invariants 5–6), `docs/blueprint/conventions.md` `#tool-versions`,
  every owned file, `cli/src/tool/catalog.ts`.
- **Lazy-load:** `cli/src/setup-config/schema.ts` (`Values`); LiquidJS docs via
  Context7.

## Ruling

| #  | Decision         | Ruling                                                                                                          | Rejected                    | Unit   |
| -- | ---------------- | --------------------------------------------------------------------------------------------------------------- | --------------------------- | ------ |
| 8  | Renderer inputs  | The selection and the `origin` host are inputs to the render; plan 2 reads git and supplies them                | Read git in the renderer    | U5     |
| 10 | Template content | Templates conform to the catalog; this repo's own `.config/` files are sources, not targets, and stay untouched | Edit this repo's `.config/` | U3, U5 |

## Edits

1. **`cli/src/tool/render.ts`** — a render context with `values`, `selection`
   and `originHost`; `renderFiles` renders the files of the selected tools only
   and carries `createOnly` from the catalog. Keep the `<%=` swap (G6) and
   strict variables.
2. **`cli/templates/mise/.config/mise/conf.d/_base/mise.dev.toml`** — the fixed
   entries plus one conditional block per base tool with mise entries, all at
   `latest`.
3. **`cli/templates/pre-commit/.config/git-conventional-commits.yaml`** — the
   forge links from a context value that the renderer computes (the selected
   forge whose `origin_hosts` holds the host; else the first selected forge in
   catalog order; else none).
4. **`cli/test/tool/render.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:build` exits 0.

## Guardrails

- Do not touch any other file under `cli/templates/**` (U3) or
  `cli/src/setup-config/**` (U4).
- The renderer reads no git and no disk. Plan 2 supplies the host.

## Commit

`feat: render create-only files and selection-dependent tool files`
