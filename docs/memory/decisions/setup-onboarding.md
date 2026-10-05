D:setup-onboarding 2026-10-05 bootstrap monorepo cli+site

# Setup onboarding decisions (2026-10-05)

- Topology: `monorepo` — pnpm workspace with `cli/` and `site/`.
- Roles and platforms, confirmed by the user, for `/vwf:architecture` to write
  into the registry: `cli` is `frontend` / `[cli]`; `site` is `frontend` /
  `[site]` (a screen platform, so the design system is mandatory; the CLI's
  terminal output belongs in its Terminal UX section).
- Stack axes deferred (`unresolved`) — architecture decides from stackgen's
  menu.
- `/vwf:init` offered and declined: this repo's tooling is the hand-copied seed
  that the bootstrap CLI will replace stackgen's tool-config with. Unlock:
  `/vwf:setup reshape`.
- Product decisions already agreed (package `@virajp.dev/bootstrap`, LiquidJS
  with `<% %>` delimiters, citty, @clack/prompts, tsdown, Starlight on
  Cloudflare, drift report as JSON for an AI tool, dogfooding via
  `bootstrap check` in CI) are in the session memory `bootstrap-cli-decisions`;
  `/vwf:product` and `/vwf:architecture` should carry them.
