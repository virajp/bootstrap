# U5 — Core tool templates

- **Wave:** 4
- **Depends on:** U3
- **Owns:** `cli/templates/mise/**`, `cli/templates/git/**`,
  `cli/templates/pre-commit/**`, `cli/templates/vscode/**`,
  `cli/templates/dprint/**`, `cli/templates/taplo/**`,
  `cli/templates/gitleaks/**`, `cli/templates/grype/**`,
  `cli/templates/virajp-linter/**`, `cli/templates/claude/**`,
  `cli/templates/graphify/**`, `cli/templates/mempalace/**`,
  `cli/test/tool/core-templates.test.ts`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/tool/core-templates.test.ts` — rendering core with
  sample values (`repo: acme/widgets`, scopes `[api]`,
  `merge_model: {develop: direct, main: pr}`, no non-core tools) yields exactly
  the catalog paths of the 12 tools this unit owns; no rendered file contains
  `virajp/bootstrap`, `bootstrap` as a repo name, `doppler`, `vwf`, `stackgen`,
  `site/`, `claude-plugins`, `flutter`, `dart` or `pubspec`; the rendered
  `.gitignore` contains a `*.bak` line; the rendered
  `.config/git-conventional-commits.yaml` lists the scope `api`; the rendered
  mise base config sets `MERGE_MODEL_DEVELOP = "direct"` and
  `MERGE_MODEL_MAIN = "pr"` and no `MERGE_MODEL`; every rendered task file
  starts with a shebang. Fails because no template exists.
- **Read first:** `cli/src/tool/catalog.ts`, `cli/src/tool/render.ts`,
  `cli/src/tool/templates.ts`, then each source file named in the index's facts
  before writing its template.
- **Lazy-load:** `docs/blueprint/entities/tool/index.md`,
  `docs/blueprint/conventions.md` `#backups`.

## Ruling

| #  | Decision                       | Ruling                                                                                                                                                                                                                                | Rejected                                         | Unit       |
| -- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ | ---------- |
| 4  | Template storage               | "Embed at build": templates live as files under `cli/templates/<tool>/<target path>`; the build bundles them into the JS as an in-memory map; nothing reads template files at run time                                                | ship templates as package files read at run time | U3, U5, U6 |
| 7  | Drift: repo sources vs catalog | the templates conform to the Tool catalog and the setup values; the repo's own tracked files stay unchanged (they are sources, not targets)                                                                                           | edit this repo's files to match                  | U5, U6     |
| 8  | Merge model in templates       | templates set `MERGE_MODEL_DEVELOP` from `values.merge_model.develop` and `MERGE_MODEL_MAIN` from `values.merge_model.main`; `MEMBERS` renders empty (the setup config has no members value)                                          | one `MERGE_MODEL`                                | U5         |
| 9  | `setup/deps` dispatchers       | authored new: each of `setup/deps/{all,install,outdated,audit,upgrade,cleanup}` runs the `setup/deps/<verb>/<tool>` task of every tool that has one and does nothing when there is none; no 1.0 tool has one                          | copy the pnpm tasks                              | U5         |
| 10 | `setup/all`                    | orchestrates by task name only; it runs a tool's setup task (`setup:ai`, `setup:secrets`, `setup:precommit`) only when that task exists                                                                                               | call tool tasks unconditionally                  | U5         |
| 11 | Template values                | every template reads only the setup values: `repo`, `commit_scopes`, `merge_model.develop`, `merge_model.main`, `tools`; derived names (e.g. the repo name, the last path segment of `repo`) are computed in the renderer, not stored | per-tool values                                  | U3, U5     |

## Edits

1. **`cli/templates/<tool>/<target path>`** — one template per catalog path of
   the 12 core tools above (every core tool except `fnox`). Start from the
   source file the facts name (byte-copy, then edit — never retype), then:
   replace this repo's name and `virajp/bootstrap` with template values; split
   `MERGE_MODEL`; render `commitScopes` from `commit_scopes`; set `MEMBERS`
   empty; remove vwf, stackgen, Doppler, `site/`, claude-plugins, dart, flutter,
   astro, tailwind and pubspec lines; fix comments that point at files that do
   not exist (`tools.dev.toml`, the `miserc.toml` layout comment); add `*.bak`
   to `.gitignore`; keep the repo-tool behaviour otherwise unchanged.
2. **`setup/deps/*` and `setup/all`** — write per rulings 9 and 10. The
   placeholder library (`_scripts/placeholder`) stays the way an unfilled slot
   announces itself.
3. **`cli/test/tool/core-templates.test.ts`** — the test-first assertions.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:build` succeeds.
- Render core for the sample values into a temp directory and run `bash -n` on
  every rendered task file (no syntax error).

## Guardrails

- Do not edit this repo's own `.config/**`, `.gitignore`, `.vscode/**` or other
  source files (ruling 7).
- Do not create `cli/templates/{fnox,github,gitlab}/` (U6).
- Every template file mode for a task is executable.

## Commit

`feat: add the core tool templates`
