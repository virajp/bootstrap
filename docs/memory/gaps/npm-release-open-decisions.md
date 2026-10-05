G:npm-release 2026-10-06 cli npm-package — parked open decisions from the generated npm-registry doctrine

# Open release decisions for `cli` (parked 2026-10-06)

Raised while generating the `npm-package` deploy doctrine
(`.claude/skills/npm-registry/`); to be settled in `/vwf:blueprint`
(conventions#pipeline) or `/vwf:plan`.

- **Tag shape** — `v*` (the user's earlier release intent) vs stackgen's
  release-trigger contract `cli-prod-v<semver>`. The doctrine only requires the
  tag's semver to equal `cli/package.json`'s version. Pick one and record it.
- **npm on the runner** — trusted publishing needs npm >= 11.5.1 on Node >=
  22.14. Unconfirmed which npm the pinned Node bundles; if older, pin a newer npm
  in the toolchain config (`npm:npm@12.2.0` vetted `warn`: no provenance on the
  npm package itself).
- **Staged publishing** — not adopted; accepted security risk (a compromised CI
  step could publish with valid provenance). Revisit when more than one person
  can push release tags.

Already decided and recorded in the doctrine: the publish task must be
rewritten (not just renamed from `p:i:publish`); `exports` closed to
`./package.json` only (user, 2026-10-06); first publish is a hand-published
placeholder before configuring the trusted publisher.
