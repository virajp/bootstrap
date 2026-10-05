---
name: npm-registry
version: 0.1.0
category: development
description: >-
  The public npm registry as this CLI's deploy target. Covers when a registry
  is the right target, the trusted-publisher (OIDC) release with provenance
  and no stored token, the idempotent publish and its dry run, the install
  smoke test, and why no configuration or secret ships inside the package.
license: MIT
allowed-tools: Read Grep Glob Edit Write Bash
---

# npm registry · public package

The deploy target for `cli`, published as `@virajp.dev/bootstrap` and run by
consumers with `npx`. This skill holds the judgment. Command flags belong to
Context7, looked up when you use them.

Read the reference that matches what you are doing. Read one, not all of
them.

| Doing | Read |
| --- | --- |
| Choosing, or questioning, this target | [Pick & trade](references/pick-and-trade.md) |
| Releasing, the dry run, first publish, trusted publisher, staged publishing, break-glass, smoke test | [Promotion & release](references/promotion.md) |
| Deciding where the CLI's configuration or credentials come from | [Config & secrets](references/config-and-secrets.md) |

The sibling `npm-package-manifest` skill covers the package itself: what
goes in the tarball and what must stay out. It applies automatically while
you edit `cli/package.json`.

**One rule applies before any reference:** publish from CI through the
trusted publisher, never with a token, and never publish one version twice.
Neither a leaked token nor a published version can be taken back.

**Health is `n/a` for this target.** Nothing runs, so there is no
readiness endpoint and no `environments:` entry. The post-release check is
the install smoke test in [Promotion & release](references/promotion.md).

**Out of scope for this skill:**

- The workflow file and the tag grammar belong to the CI system, under
  stackgen's release-trigger contract.
- Build commands belong to the language bundle.
- The local stack belongs to the harness contract.
