D:architecture 2026-10-05 bootstrap monorepo cli(system/cli)+site(frontend/site)

# Architecture decisions (2026-10-05)

- `cli` is `system` / `[cli]` (it scaffolds other repos), not `frontend` as
  first confirmed at setup — the platform guidance routes a scaffolding tool to
  `system`; nothing downstream keys on the role.
- `site` is `frontend` / `[site]`; design system is therefore mandatory.
- Stack pins: repo `pnpm-workspace`; cli `typescript-effect-cli`, deploy
  `npm-package`; site `astro-ssg`, deploy `cloudflare-workers-static`, design
  `claude-code`, stylesheet `plain-css`; cicd `github-actions` for both;
  backing `[]` for both.
- **Supersedes** the earlier citty + @clack/prompts + tsdown choice for the CLI:
  the user picked the Effect CLI template. Library choices are `plan`'s to
  settle against that template's conventions.
- Foundations: core five adapted, not deferred (users single-class-no-accounts,
  observability local-logs-only, reliability static-site-best-effort, DR
  git-is-source-of-truth, incident issue-tracker-and-security-policy); change
  logs accepted as keepachangelog; the other electives not applicable.
