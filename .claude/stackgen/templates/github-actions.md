---
slug: github-actions
axis: cicd
kind: ci-system
components:
  - ci-system/github-actions@0.2.1
platforms: []
languages: []
language_facts: {}
optional_languages: []
frameworks: []
dependencies: []
capabilities: []
artifact: n/a
harness: n/a
materialized: "2026-10-05"
---
# CI — GitHub Actions

vwf's delivery-pipeline contract, implemented on GitHub Actions: workflow
layout, toolchain installation through the repo's own manager, the gate
sequence, and the tag-triggered release shape.

The neutral contract it cites is stackgen's release-trigger contract — the
recommended release-trigger mechanism vwf's rules deliberately leave open:
the tag grammar, the branch mapping, the release task names and how far a
deploy path may be split. It is the bundle's second half, the way a
capability contract is elsewhere.

**Exactly one CI system per repo.** Generating for a second produces a pipeline
nobody runs and nobody updates, which is worse than none — a green check that
means nothing.

This bundle exists because the pack behind it had **no way to be offered**. CI
is chosen by the `projects.<name>.cicd` config key, and until the `cicd` axis
the menu was the only door a template could come through — so a pack landed
that nothing could ever materialize, silently. The slug is that config token.

# GitHub Actions — conventions

The repo's delivery pipeline, implementing vwf's delivery-pipeline contract on
GitHub Actions.

**The pipeline installs the toolchain manager and nothing else.** No
per-language setup action, no system package install, no global install. Every
tool a job needs is declared in the repo's mise config; only *how mise itself is
installed* is the pipeline's business.

**Steps run through the task library** — `mise run <task>` — never a binary the
toolchain step did not put on `PATH`. The same task names a developer runs
locally are the ones CI runs, which is what makes local and CI the same gate.

**`MISE_ENV: ci` is set** when the repo defines a `mise.ci.toml` variant, and
omitted when it has only a flat `mise.toml`.

**Every third-party action is pinned to an explicit version.** An unpinned
action is remote code executing with the pipeline's credentials.

**Releases are tag-triggered** on `<project>-<env>-v<semver>`, validated against
the branch before publishing, and tested before released. The grammar, the
branch mapping and the release task names are stackgen's release-trigger
contract's, which serves vwf's delivery-pipeline contract; both are cited rather
than restated.

**Prefer federated identity to stored tokens.** A stored long-lived credential
is the finding most worth preventing here.

Full judgment: the `github-actions` skill's references.
