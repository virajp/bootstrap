DECISIONS|tool-setup-config run 2026-10-07|source:
docs/plans/2026-10-06-2332-tool-setup-config

- cli scaffold: Effect v4 `effect/cli`, `NodeServices.layer`; `--version` prints
  the bare version through a `CliOutput.Formatter` override; a bare command
  exits 2 (`NoCommand`).
- TypeScript is pinned to `^6.0.3`: the TypeScript 7 Go binary fails grype.
- pnpm 12 writes a two-document lockfile, so `check-yaml` takes
  `--allow-multiple-documents` (user ruling). pnpm 12 ignores `.npmrc` for
  `ignore-scripts`; `ignoreScripts: true` lives in `pnpm-workspace.yaml` (user
  ruling, G17).
- Templates are embedded at build time through `import.meta.glob` with `?raw`
  and an inline rolldown raw loader in `tsdown.config.ts` (user ruling, G7).
  Keys are `<tool>/<path>`; `.DS_Store` is excluded.
- LiquidJS matches `<%` before `<%=`, so the renderer swaps `<%=` for an
  internal output delimiter. `strictVariables` and `strictFilters` are on.
- `renderFiles` returns `{ content, executable }`; a file is executable when its
  path is under `.config/mise/tasks/`.
- Setup config: `repo` segments are `[A-Za-z0-9._-]` (this blocks Tera injection
  in mise TOML). Paths are normalised per segment and refuse `:`, `.git` in any
  spelling, `git~N`, and empty, `.` or `..` segments.
- Templates: a core file must not depend on the non-core tool selection (the
  changelog forge URLs were dropped). Reversed by plan
  2026-10-08-1124-tool-catalog (assumed decision #21): the rescoped blueprint
  drops the core/non-core model, and `git-conventional-commits.yaml` renders the
  forge links again, chosen from the selected tools with `origin_hosts` and the
  `origin` host. gitleaks keeps its default rules (`useDefault = true`), prints
  with `--redact=100`, and scans history with a `.gitleaksignore` hint. fnox
  runs with `--config .config/fnox.toml`. The linter and sort-package-json are
  mise npm tools on `latest`; the float is a risk the user accepted (G18). This
  repo's own `bin/` and related ignore lines stay in the templates (user ruling,
  R4).
