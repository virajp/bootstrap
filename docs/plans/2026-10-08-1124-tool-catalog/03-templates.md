# U3 — Templates for the rescoped target paths and dispatchers

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/templates/**`, `cli/test/tool/core-templates.test.ts`,
  `cli/test/tool/new-templates.test.ts`, `cli/test/fixtures/**`
- **Model:** opus
- **Kind:** code
- **Test first:**
  - Every `{ tool, path }` of the catalog has exactly one template at
    `cli/templates/<tool>/<path>`, and no template exists that the catalog does
    not name.
  - Each `_default` dispatcher (`code/{format,lint,sec,graph}`,
    `setup/{secrets,ai}`, `setup/deps/<verb>` for the five verbs), run by `mise`
    in a temporary directory with no per-tool subtask, prints `no tool` and
    exits 0. With one per-tool subtask present, it runs that subtask. Skip the
    run test when `mise` is not on `PATH`, as the G12 test does with `taplo`.
  - `setup/deps/all` runs the `cleanup`, `install`, `upgrade`, `outdated` and
    `audit` dispatchers in that order.
  - Every `conf.d/<tool>/mise.dev.toml` asks for `latest`;
    `conf.d/_base/mise.toml` pins osv-scanner to an exact version.
  - Fails because the templates have the old layout.
- **Read first:** `docs/blueprint/entities/tool/index.md` (Target paths, Task
  dispatchers, Selection-dependent files), `docs/blueprint/conventions.md`
  `#tool-versions`, `cli/src/tool/catalog.ts`, both owned test files.
- **Lazy-load:** this repo's own `.config/mise/tasks/**`,
  `.config/mise/conf.d/**`, `dprint.json` — read as sources only.

## Ruling

| #  | Decision                                                                                                                       | Ruling                                                                                                          | Rejected                    | Unit   |
| -- | ------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------- | --------------------------- | ------ |
| 4  | Drift: plain task files `code/format`, `code/lint`, `code/sec`, `code/graph`, `setup/secrets`, `setup/ai`, `setup/deps/<verb>` | `_default` dispatchers owned by `mise`, plus per-tool subtasks, as Tool Target paths and Task dispatchers say   | Keep single-tool task files | U3     |
| 10 | Template content                                                                                                               | Templates conform to the catalog; this repo's own `.config/` files are sources, not targets, and stay untouched | Edit this repo's `.config/` | U3, U5 |

## Edits

1. **`cli/templates/mise/.config/mise/tasks/**`** — add the `_default`
   dispatchers. Each runs every per-tool subtask file present in its folder
   (`code:sec` runs `code:sec:gitleaks` and `code:sec:grype` when present), else
   prints `no tool` and exits 0. Move each `setup/deps/<verb>` file to
   `setup/deps/<verb>/_default`. Keep `setup/deps/all` a file. `code/all` and
   `setup/all` call the dispatchers and never test which tools are present.
2. **Per-tool subtasks** — move the content of the old single-tool tasks:
   `dprint/.../tasks/code/format` → `.../code/format/dprint`;
   `virajp-linter/.../code/lint` → `.../code/lint/virajp-linter`;
   `gitleaks/.../code/sec` → `.../code/sec/gitleaks`, and split the grype part
   into `grype/.../code/sec/grype`; `fnox/.../setup/secrets` →
   `.../setup/secrets/fnox`; `claude/.../setup/ai` → `.../setup/ai/claude`;
   `graphify/.../code/graph` → `.../code/graph/graphify`.
3. **`conf.d/<tool>/mise.dev.toml`** — add one for `gitleaks`, `grype`, `fnox`,
   `graphify` and `mempalace`, each with the tool's mise entry at `latest`.
   Remove the grype, gitleaks and fnox entries from
   `mise/.config/mise/conf.d/_base/mise.dev.toml` (U5 makes that file
   selection-dependent).
4. **Remove** `cli/templates/dprint/.config/dprint.json` and
   `cli/templates/claude/.config/mise/conf.d/ai/mise.dev.toml` with `rm`.
5. **`cli/test/tool/core-templates.test.ts`**, **`new-templates.test.ts`**,
   **`cli/test/fixtures/**`** — the test-first cases; adapt the existing
   assertions to the new paths.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:build` exits 0 (the `?raw` embed picks up the new
  files).

## Guardrails

- Do not touch `cli/src/**` (U2, U4, U5).
- Do not edit this repo's own `.config/` files or `dprint.json`. They are
  sources, not targets.
- Task files have no extension and start with a shebang. The renderer sets the
  mode; do not rely on the mode of the template file.
- Template syntax uses the `<% %>` / `<%= %>` delimiters (G6).
- Delete with `rm`, never `git rm`.

## Commit

`feat: lay out task dispatchers and per-tool subtasks in the templates`
