# U6 — Templates with no source: fnox, github, gitlab

- **Wave:** 4
- **Depends on:** U3
- **Owns:** `cli/templates/fnox/**`, `cli/templates/github/**`,
  `cli/templates/gitlab/**`, `cli/test/tool/new-templates.test.ts`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/tool/new-templates.test.ts` — rendering `fnox`
  alone, `github` alone and `gitlab` alone with the sample values yields exactly
  each tool's catalog paths; `.github/ISSUE_TEMPLATE/config.yml` turns off blank
  issues; every `.yml` parses as YAML; `.config/fnox.toml` parses as TOML; the
  GitLab merge-request template is at
  `.gitlab/merge_request_templates/Default.md`. Fails because no template
  exists.
- **Read first:** `cli/src/tool/catalog.ts`, `cli/src/tool/render.ts`,
  `.config/mise/tasks/setup/secrets` (the current secrets task,
  Doppler-flavoured).
- **Lazy-load:** fnox, GitHub issue-form and GitLab description-template docs
  via Context7 or the vendors' docs.

## Ruling

> | 4 | Template storage | "Embed at build": templates live as files under
> `cli/templates/<tool>/<target path>`; the build bundles them into the JS as an
> in-memory map; nothing reads template files at run time | ship templates as
> package files read at run time | U3, U5, U6 | | 7 | Drift: repo sources vs
> catalog | the templates conform to the Tool catalog and the setup values; the
> repo's own tracked files stay unchanged (they are sources, not targets) | edit
> this repo's files to match | U5, U6 |

From the blueprint (user decision): `github` holds
`.github/pull_request_template.md` and
`.github/ISSUE_TEMPLATE/{bug.yml,feature.yml,config.yml}` —
"`.github/workflows/ci.yml` belong to `stackgen`", so no workflow; `gitlab`
holds `.gitlab/merge_request_templates/Default.md` and
`.gitlab/issue_templates/{Bug.md,Feature.md}`.

## Edits

1. **`cli/templates/fnox/.config/fnox.toml`** — a minimal fnox config with no
   secret values and no provider credentials.
2. **`cli/templates/fnox/.config/mise/tasks/setup/secrets`** — sets up fnox for
   the repository; no Doppler.
3. **`cli/templates/github/**`** — the pull-request template (a short checklist)
   and issue forms for a bug and a feature; `config.yml` with
   `blank_issues_enabled: false`.
4. **`cli/templates/gitlab/**`** — the same content in GitLab's Markdown
   description-template form.
5. **`cli/test/tool/new-templates.test.ts`** — the test-first assertions.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:build` succeeds.

## Guardrails

- No secret value, token or account name in any template.
- Do not touch U5's template folders.

## Commit

`feat: add the fnox, github and gitlab templates`
