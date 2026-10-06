# U3 — Template renderer and build-time embedding

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/src/tool/render.ts`, `cli/src/tool/templates.ts`,
  `cli/test/tool/render.test.ts`, `cli/test/fixtures/templates/**`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/tool/render.test.ts` — rendering a fixture template
  that uses `<% if %>` and `<%= repo %>` with sample setup values returns the
  expected text; a template that reads an undefined variable fails with a typed
  render error; rendering the selected tool set returns one entry per catalog
  path of those tools. Fails because the module does not exist.
- **Read first:** `cli/src/tool/catalog.ts`,
  `docs/blueprint/entities/setup-config/schema.yaml` (the `values` shape),
  `cli/tsdown.config.ts`, `cli/vitest.config.ts`.
- **Lazy-load:** LiquidJS options and Effect error-channel docs via Context7.

## Ruling

> | 3 | Template engine | LiquidJS with tag delimiters `<%` `%>` and output
> delimiters `<%=` `%>`, `strictVariables: true` (earlier product decision) |
> default `{% %}` / `{{ }}` delimiters | U3 | | 4 | Template storage | "Embed at
> build": templates live as files under `cli/templates/<tool>/<target path>`;
> the build bundles them into the JS as an in-memory map; nothing reads template
> files at run time | ship templates as package files read at run time | U3, U5,
> U6 | | 11 | Template values | every template reads only the setup values:
> `repo`, `commit_scopes`, `merge_model.develop`, `merge_model.main`, `tools`;
> derived names (e.g. the repo name, the last path segment of `repo`) are
> computed in the renderer, not stored | per-tool values | U3, U5 |

## Edits

1. **`cli/src/tool/templates.ts`** — the single place that turns
   `cli/templates/**` into an in-memory map `target path → template text` at
   build time, the same way in the tsdown build and under Vitest (for example a
   bundler glob import of raw text, or a build plugin; choose one mechanism that
   works in both, and record it in `DECIDED:`). A target path is the file's path
   under `cli/templates/<tool>/`. Nothing reads the file system at run time.
2. **`cli/src/tool/render.ts`** — a LiquidJS engine with the delimiters above
   and `strictVariables: true`, rendering from the in-memory map only. Input:
   the selected tool set and the setup values. Output: the rendered content per
   target path for every path of those tools. Failures are typed values in the
   Effect error channel (template missing for a catalog path, render error with
   the path), never thrown. Derived values (repo name = last segment of `repo`,
   owner path = the segments before it) are computed here.
3. **`cli/test/fixtures/templates/**`** — the fixture templates the test uses.
4. **`cli/test/tool/render.test.ts`** — the test-first assertions.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:build` succeeds and `cli/dist/bin.mjs` contains no
  reference to `cli/templates` as a run-time path.

## Guardrails

- Do not create files under `cli/templates/` (U5, U6 own them).
- Do not edit `cli/src/tool/catalog.ts` (U2) or the build configs (U1); if the
  embed needs a build-config change, return `UNRESOLVED:` naming it.

## Commit

`feat: render tool templates with liquidjs`
