---
slug: npm-package
axis: deploy
kind: deploy-target
components:
  - deploy-target/npm-registry@generated
platforms: []
languages: []
language_facts: {}
optional_languages: []
frameworks: []
dependencies: []
capabilities: []
artifact: npm-package
harness: n/a
materialized: "2026-10-06"
---
# Deploy — Package registry · npm

The target for a project users **install** rather than one you run: a CLI, or a
library published outside the workspace. The registry is the host; there is no
environment to deploy into and nothing to keep warm.

## Artifact

One published package per project, built from the workspace and pruned to what
consumers need — the `files` allowlist, the `bin` entry for a CLI, and an
`exports` map for a library. The tarball is the only artifact; it carries no
environment-specific configuration, because a published version is immutable and
the same version is what every consumer resolves.

## Pipeline

Build → `publish`, wrapped in mise `release:*` tasks so the same command runs
locally and in CI.

Publishes obey vwf's delivery-pipeline contract — named here, never copied,
because no path from this plugin spells vwf's root. The caller fetching these
conventions is vwf, which already has it: tag-triggered only, shaped
`<project>-<env>-v<semver>`, branch-validated, and gated on the tagged
project's tests plus its dependents'. Two rules are specific to a registry:

- **Publish from CI with a trusted publisher (OIDC)** rather than a stored
  token, so provenance is attached automatically and no long-lived credential
  exists to leak.
- **The publish step is idempotent** — a version already on the registry is a
  skip, not a failure. Tag re-points and re-runs are then safe, which matters
  because a registry publish cannot be rolled back, only superseded.

## Versioning

The published version **is** the release record — semver, and the same number
the tag carries. A version is never republished with different contents; a
mistake ships as the next patch. There is no promotion between environments: a
`staging` publish is a prerelease tag (`1.4.0-rc.1`) on the same line, not a
separate artifact.

## Configuration and secrets

None ship in the package. A CLI reads configuration from its own flags,
environment, and per-user config file at run time on the user's machine — so the
`environment.md` catalog for this project names what it *reads*, and the package
itself embeds nothing. A published secret is unrecallable; the registry is the
one target where a leak cannot be revoked by redeploying.

## Private plane

A package that must not be public is published **scoped and private** to the
organization, or to a private registry. There is no network layer to hide behind
here: publication *is* the exposure, so the access setting is the whole control.

## Health

Not applicable — nothing is running to probe, so this project carries no entry
under `environments:` and no `harness.health` path. `/vwf:verify` has no URL to
probe for it and says so rather than reporting a failure.

The equivalent post-release check belongs in the pipeline, not in `verify`: an
**install smoke test** that installs the published version from the registry
into a clean environment and runs its entrypoint. Make it the last step of the
publish job — it is the only check that exercises what consumers actually
resolve, rather than what the workspace built.

# npm registry · public package — conventions

The deploy target for `cli`, a command-line tool people **run** rather than a
service anyone hosts. The registry is the host; there is no environment to
deploy into and nothing to keep warm. Consumers run it with
`npx @virajp.dev/bootstrap`.

**One tarball, built from the workspace, is the only artifact.** It is
packed by pnpm, because only pnpm rewrites `workspace:` and `catalog:`
specifiers and applies `publishConfig` overrides as it packs, and it is
uploaded by npm, the client whose trusted-publishing support the registry
documents. The tarball that the dry run lists is the tarball that ships.

**The manifest's `files` allowlist decides what ships**, not an ignore
list. A local `.env` is not on npm's always-excluded list, so a deny-list
leaks the variant nobody wrote down. Run a pack listing before every
release and read it.

**Exactly one `bin` entry**, pointing at a built file that starts with
`#!/usr/bin/env node`. `npx` runs the single bin entry. Without the shebang
the script starts without Node.

**Publishing happens from CI only, through a trusted publisher (OIDC).**
No npm token exists anywhere. Provenance comes from trusted publishing
itself, and only for a public repository built on GitHub-hosted runners
(docs.npmjs.com/trusted-publishers). The
registry needs the package to exist before a trusted publisher can be
configured, so the first version is a one-time hand publish with 2FA.
After that, token publishing is switched off for the package.

**There is one break-glass path: the maintainer's interactive publish.**
When CI cannot publish, the package's maintainer, the only 2FA holder,
runs the repo's local publish task by hand. They answer a 2FA prompt for
that publish, run the same dry run, pack listing and smoke test, then log
out and record why CI was bypassed. The CI path is fixed before the next
release. Token publishing stays off throughout. A version published this
way has no provenance, which is why this is the exception and not a second
release path.

**A version is published once and never again.** Releasing an
already-published version does not publish again. It runs the install
smoke test against what is already there and passes or fails on that, and a mistake ships
as the next patch. The registry refuses republishing anyway. An unpublished
version number is gone for good, so treat an unpublish as unavailable and
use deprecate.

**The release is a repo task with a `--dry-run` path.** The same command
runs locally and in CI, and the dry run reports everything a publish would
do except the upload. The workflow that triggers it, and the tag grammar it
validates, belong to the CI system and stackgen's release-trigger contract.
They are cited, not restated.

**Nothing environment-specific or secret is inside the package.** The CLI
reads its configuration from flags, the environment and its repo config
file on the user's machine, at run time. A published secret cannot be
recalled.

**Health is `n/a`.** Nothing runs. The equivalent check is the install
smoke test: after publishing, install the published version from the
registry into an empty directory and run it.

Full judgment: the `npm-registry` skill's references, and the
`npm-package-manifest` skill while editing `cli/package.json`.
