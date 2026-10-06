---
type: vwf-conventions
title: Conventions
description: Cross-cutting decisions every bootstrap flow and entity doc links to
  rather than repeats.
status: draft # draft | reviewed | stable
timestamp: 2026-10-06
---

# Conventions

Cross-cutting decisions referenced by flow and entity docs. Defined once; docs
link the anchor rather than repeating it. Mirrors the `cross_cutting` block in
[registry.yaml](registry.yaml).

## Auth {#auth}

None. The product has no accounts, no sign-in and no server (`auth: none`,
`users: single-class-no-accounts`). Every action runs with the invoking
person's or agent's own file-system and git permissions.

## Errors {#errors}

Every `cli` command reports outcome by **exit code** and, under `--json`, one
**structured report** (`errors: exit-codes-and-structured-report`).

| Exit | Meaning                                                                           |
| ---- | --------------------------------------------------------------------------------- |
| `0`  | success; for a read-only check, no drift                                          |
| `1`  | drift found, or the command declined to act (e.g. the repository is already set up) |
| `2`  | usage error — bad or missing flag/argument; prints short usage                    |
| `3`  | failure — not a git repository, unreadable setup file, render error, write failed |

- Results go to stdout; progress, warnings and errors go to stderr.
- `--json` prints exactly one JSON document to stdout and nothing else, carrying
  the same outcome as the exit code; it is the machine contract an agent reads.
- Every error message states what happened, why, and the exact next command to
  run. No internal trace unless `--verbose`.
- A command that cannot complete its writes leaves the repository as it found
  it — see [#backups](#backups) and `baseline/atomic-multi-write`.

## Config {#config}

Configuration reaches `cli` two ways only, and nothing is read from the
environment or a secrets store (`config: flags-and-repo-config-file`):

- **Flags** on the command line, which always win.
- **The repository setup file** `.config/bootstrap.yaml`
  ([setup-config](entities/setup-config/index.md)), committed with the
  repository. Its presence is the marker that a repository has been set up.

Detected values (e.g. the repository's `owner/name` from its git remote) are
offered as defaults, never applied unconfirmed in an interactive run. A
non-interactive run with a required value neither flagged, recorded nor
detectable is a usage error (exit `2`). The product has no secrets and no
external integration, so there is no environment catalog.

## Backups {#backups}

Bootstrap never destroys a file it did not write. Before writing a path that
already exists and that bootstrap does not already manage, it moves the existing
file aside:

- an existing file already byte-identical to what would be written is left
  untouched — no backup, no rewrite — and reported as `unchanged`;
- first backup: `<name>.bak` (e.g. `.gitignore.bak`);
- if that exists: `<name>.1.bak`, then `<name>.2.bak`, … — the lowest free
  number; an existing backup is never overwritten;
- the `.bak` suffix is always last, so one ignore pattern (`*.bak`) covers every
  backup — the core `git` group's ignore file carries it;
- every backup made is listed in the command's output (original → backup path),
  human and `--json`, so a person or agent can merge hand edits back;
- on a failed or interrupted run, backups are restored to their original paths
  as part of the rollback.

## Observability {#observability}

Local only (`observability: local-logs-only`). Diagnostics go to stderr, with
`--verbose` adding each per-file decision. No telemetry, no analytics, no
network calls home — a product non-goal. The `site` collects no visitor
analytics.

## Theming {#theming}

The `site` is **dark-only**: one theme, no light variant and no theme switch;
it does not follow the visitor's system preference. Colour roles come from
[design-system.md](design-system.md). The `cli` uses the terminal's own colours
mapped to semantic roles (design-system Terminal UX) — never fixed values.

## Release & changelog {#changelog}

Each project keeps a human-written changelog in the keep-a-changelog shape
(`changelog: keepachangelog`): an `Unreleased` section, then one section per
released version, grouped Added / Changed / Deprecated / Removed / Fixed /
Security. A release moves `Unreleased` under its version. Versions follow
semantic versioning; for `cli`, a change to the setup-file format, a removed or renamed
group, or a changed exit-code meaning is a major version.

## Reliability targets {#reliability}

`reliability: static-site-best-effort`. The `site` is static files served from
the host's edge with no stated availability target. The `cli` runs locally; its
reliability contract is correctness — the acceptance criteria of each flow —
not uptime.

## Disaster recovery {#disaster-recovery}

`dr: git-is-source-of-truth`. Everything the product is — source, templates,
docs, the site's content — lives in the git repository and its remote; a lost
host or package release is recovered by rebuilding from a tag. In a target
repository, bootstrap's own writes are recoverable the same way, and through
[#backups](#backups) for files it replaced.

## Incident response {#incident}

`incident: issue-tracker-and-security-policy`. Bugs and regressions are reported
on the repository's issue tracker. Vulnerabilities are reported privately per
the repository's security policy, never in a public issue; a fix ships as a
patch release with a `Security` changelog entry.

## Web metadata {#web-metadata}

Product-wide `site` metadata defaults (site name, default description, locale,
social handle, organisation facts) are elicited with the first `site` flow and
recorded here.

## Engineering baseline {#baseline}

Defaults every doc and cycle follows; only exceptions are documented, on the
deviating doc and as an `enforcement.rules` waiver.

- **`baseline/atomic-multi-write`** — a command that writes more than one file
  is all-or-nothing: on any failure or interruption it removes what it wrote
  and restores what it moved aside, so a partial set of writes is never left
  observable.
- **`baseline/boundary-validation`** — every input crossing a boundary is
  validated against its contract and rejected, never coerced: flags and
  arguments, the setup file
  ([setup-config schema](entities/setup-config/schema.yaml)), and every `--json`
  report before it is printed.
- **`baseline/business-technical-separation`** — the setup decisions (what to
  render, what counts as drift, what to back up) never live inside technical
  helpers (file system, git, terminal, templating), which sit in shared layers
  the decisions consume.
- **`baseline/graceful-shutdown`** — an interrupt during a writing command
  triggers the same rollback as a failure; acknowledged work is never half
  applied.
- **`baseline/structured-logs-no-pii`** — `--verbose` diagnostics are
  key-value, and never contain secrets or personal data; paths and group names
  are the identifiers. No telemetry pipeline exists (see
  [#observability](#observability)).
- **`baseline/expand-contract`** — the setup file is stored data with a `format`
  number: a newer `cli` reads every older format it has shipped; a format change
  that removes or renames a field lands only in a major release, and a file
  newer than the running `cli` is refused, never guessed at.

**Inapplicable** (no surface — not waivers): with **no datastore and no
concurrent writer** — `write-versioning`, `soft-delete`, `server-time`,
`integer-money`; with **no API** — `idempotency-keys`, `error-envelope`
(replaced by [#errors](#errors)), `cursor-pagination`, `retry-discipline`; with
**no events, services or workers** — `tolerant-reader`, `stateless-processes`.

## Delivery pipeline {#pipeline}

Environments use the canonical names only — synonyms are drift.

| Environment   | Serves                    | Built from | Deployed by           |
| ------------- | ------------------------- | ---------- | --------------------- |
| `development` | the developer's machine   | any branch | never — run locally   |
| `staging`     | testers only              | `develop`  | a deliberate release act |
| `production`  | users                     | `main`     | a deliberate release act |

- **`pipeline/mise-built`** — every CI tool comes from the repository's
  toolchain config; never a language-setup step or a global install.
- **`pipeline/tag-triggered-deploys`** — a deploy is an explicit, recorded act
  naming one project and one environment, never a side effect of a push;
  recommended shape `<project>-<env>-v<semver>`.
- **`pipeline/branch-validated`** — the deploy proves the commit it ships is
  reachable from that environment's branch, and fails otherwise.
- **`pipeline/staging-is-not-a-release`** — a staging deploy is never announced,
  changelogged as released, or frozen.
- **`pipeline/tested-before-release`** — no deploy step runs until the project's
  and its dependents' tests pass in the same run.
- **`pipeline/load-proven`** — a flow whose declared peak rate reaches `~10/s`
  is load-proven on staging before its first production release.
- **`pipeline/rollback-path`** — every production release states its rollback;
  the previous release stays installable/deployable.
- **`pipeline/dependency-audit`** — every release-capable run audits the
  lockfile; a known-critical advisory fails it, waivable per advisory with a
  reason and a date.
