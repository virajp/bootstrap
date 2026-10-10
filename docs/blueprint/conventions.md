---
type: vwf-conventions
title: Conventions
description: Cross-cutting decisions every bootstrap flow and entity doc links
  to
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
`users: single-class-no-accounts`). Every action runs with the invoking person's
or agent's own file-system and git permissions.

## Errors {#errors}

Every `cli` command reports outcome by **exit code** and, under `--json`, one
**JSON result** (`tui` takes no `--json`; it is a usage error there)
(`errors: exit-codes-and-structured-report`).

| Exit  | Meaning                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `0`   | success                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `1`   | declined to act — a target has uncommitted changes ([#safety](#safety)), the setup file changed while the `tui` view was open (next command `bootstrap tui`), `add`, `remove` or `show` on a repository that is not set up, `show` on a selection that misses an unremovable tool or a dependency that [Repair on read](entities/setup-config/index.md#repair) would add back, or `add`, `remove` or `tui` on a repository set up by an older bootstrap (next command `bootstrap init`)                                  |
| `2`   | usage error — bare `bootstrap` with no command (prints the top-level help), bad or missing flag/argument (prints short usage), an unremovable tool named for removal, a tool named for removal, or replaced outside a `requires` slot, that a selected tool still needs (names that tool), a change in init, add or remove without `--yes` where no prompt is possible ("changes need --yes"), a replacement with no consent or of a tool that is not replaceable, `tui` without terminals (checked after the preflight) |
| `3`   | failure — not a git repository, `mise` not installed, unreadable setup file, render error, a target path that is not a regular file, write failed                                                                                                                                                                                                                                                                                                                                                                        |
| `130` | interrupted (Ctrl-C) — nothing written, or the repository was restored                                                                                                                                                                                                                                                                                                                                                                                                                                                   |

- An interrupt (Ctrl-C) exits `130` in every command and at any moment — in the
  `tui` view, while reading, or while writing. During a write it first triggers
  the full rollback ("interrupted — repository restored"); before any write it
  says "interrupted — nothing written". A `--json` run prints the same as its
  `error`. A restore that itself fails exits `3` listing every path not
  restored.
- After `--help` (and `bootstrap --version`) and the flag and argument check,
  every command runs only inside a git repository; anywhere else it writes
  nothing and exits `3` with "not a git repository — run `git init` first".
- Every command except `show` needs `mise` on the `PATH` (`show` runs no tool);
  without it, the command writes nothing and exits `3` with "mise not installed"
  and the install command. When the `mise` found is not `~/.local/bin/mise`, the
  command prints one warning naming the path found and continues. From any
  subdirectory, every command works at the repository root: the setup file and
  every tool path are relative to the root.
- Results go to stdout; progress, warnings and errors go to stderr.
- On every command but `tui`, `--json` prints exactly one JSON document, the
  JSON result, to stdout and nothing else, carrying the same outcome as the exit
  code; it is the machine contract an agent reads. This holds on every exit, `2`
  and `3` included: the document then carries an `error` (what happened, why,
  and the exact next command) in place of results. Every document has a
  top-level `exit` (the exit code). The `error` object has exactly the keys
  `what`, `why` and `next_command`; a failed restore adds `unrestored`, a list
  of every path not restored. Warnings go in a top-level `warnings` list of
  strings, empty when there is none. One exception: `bootstrap show` on a
  selection that misses an unremovable tool or a dependency that
  [Repair on read](entities/setup-config/index.md#repair) would add back, exits
  `1` and its document carries the selected fields and the `error` together
  ([Show the setup](flows/cli/170-show-setup/index.md)).
- `bootstrap tui` is the only full-screen interactive command, and it needs
  stdin and stdout to be terminals. `show` never prompts. `init`, `add` and
  `remove` take their values and decisions from flags, the recorded setup file,
  the defaults and, for the repo path, the remote `origin`; their one prompt is
  the consent to apply the changes ([#config](#config)).
- After a successful write (or the `--dry-run` of one), the next command
  `MISE_ENV=dev mise run setup:all` (`--json`: top-level `next_command`) is
  printed if and only if the run added a tool, direct or dependency (a first
  run, a new tool, a replacement or a repair) or created, changed or deleted a
  `mise` config file (`.config/mise.toml`, `.config/miserc.toml` or a file under
  `.config/mise/conf.d/`); otherwise there is no next command and no
  `next_command` key. In human output, a run of `init`, `add`, `remove` or the
  `tui` that created, changed, deleted or replaced a file ends with exactly
  "commit the changes, then run `MISE_ENV=dev mise run setup:all`" when the next
  command is printed, else with exactly "commit the changes"; a run with nothing
  to create, change, delete or replace has no closing line, and a `--dry-run`
  shows the closing line of the real run.
- Every error message states what happened, why, and the exact next command to
  run. No internal trace unless `--verbose`.
- A run reports its exit `2` refusals together, in one error: `what` and `why`
  name each refusal in step order, and `next_command` is the one command that
  fixes all of them. The usage check, which needs only the flags and the catalog
  and runs before the preflight, reports all its refusals and stops; the
  refusals that need the setup file are found after it and reported together in
  the same way.
- A command that cannot complete its writes leaves the repository as it found it
  — see [#safety](#safety) and `baseline/atomic-multi-write`.

## Config {#config}

Configuration reaches `cli` two ways only, and nothing is read from the
environment or a secrets store (`config: flags-and-repo-config-file`):

- **Flags** on the command line, which always win.
- **The repository setup file** `.config/bootstrap.yaml`
  ([setup-config](entities/setup-config/index.md)), committed with the
  repository. Its presence is the marker that a repository has been set up.

Detected values (e.g. the repository path `owner/name` from its git remote) and
fixed values are defaults, and a default applies without consent. A flag wins
over a recorded value, and a recorded value wins over a detected one. A required
value neither flagged, recorded nor defaulted is a usage error (exit `2`) whose
next command offers `bootstrap tui` or the missing flag. The `tui` shows the
defaults pre-filled, and the owner confirms them at apply.

`init`, `add` and `remove` run in this order:

1. Read the flags and the recorded setup file, and prepare the full list of
   changes: each file to create, change, delete or replace, and a repair of the
   setup file ([Repair on read](entities/setup-config/index.md#repair)). Every
   refusal is found here, before consent: a usage error, a replacement without
   `--replace` or of a tool that is not replaceable, a dirty target
   ([#safety](#safety)).
2. An empty list: write nothing, exit `0`.
3. `--dry-run`: show the list, write nothing.
4. Consent. `--yes` (`-y`) gives it. Without `--yes`, when stdin and stdout are
   terminals and `--json` is not given, show the list and ask "apply these
   changes? y/N": "y" or "yes" applies; Enter, "n" or "no" writes nothing and
   exits `0` with "nothing changed". Without `--yes` and with no prompt
   possible, write nothing and exit `2` "changes need --yes", the next command
   being the same command plus `--yes`.
5. Apply the list.

A replacement needs `--replace` at step 1 and consent at step 4. In the `tui`
the in-place prompt and the apply confirm give consent. `--yes` and the consent
prompt are never consent to replace a tool
([Tool category](entities/tool-category/index.md)). The product has no secrets
and no external integration, so there is no environment catalog.

## Safety {#safety}

The repository's git history is the only record of earlier states. Bootstrap
makes no backup copies and never commits; the owner examines the changes and
commits them.

- **Targets.** The targets of a run are the paths it would create, change
  (content or file mode) or delete. A path whose rendered content and mode equal
  the working copy is not a target; it is reported as `unchanged`. The setup
  file `.config/bootstrap.yaml` is a target like any other file, so a hand edit
  to it must be committed before the next command.
- **Clean targets only.** If any target is modified, staged, untracked, ignored,
  or deleted in the working tree with the deletion not committed, the command
  writes nothing and exits `1`, listing every such path with the fix "commit or
  stash these files, then run again".
- **Write last.** A command computes every render, decision and check in memory
  first; it writes only after every check passes and every confirm is given.
- **Restore on failure.** If a write fails or the run is interrupted during the
  writes, the command restores every changed or deleted target from `HEAD` and
  deletes every file it created, so the working tree is as it was. A restore
  that fails exits `3` listing every path not restored.
- **Not a file.** A target that is a directory or a symlink, or whose parent
  directory is a symlink or a regular file, is never written, deleted or
  followed: the command exits `3` naming every such path, with the fix "move it
  away and run again".
- **Create-only files.** A file the [Tool](entities/tool/index.md) catalog marks
  `create_only` is written only when it is absent; it is never changed and never
  deleted, and an existing one is reported as `kept`.
- **Empty directories.** After its deletions, a run removes each directory that
  a deletion left empty, then each parent that is left empty in turn, never the
  repository root. A directory that holds any other entry stays.
- **File mode.** Every task file is written executable (`0755`); every other
  file is written `0644`.
- **Files no longer rendered.** A path in the setup config's `files` that the
  running bootstrap does not render is never deleted by any command; it stays in
  `files` and every command reports it as `orphaned` until the owner deletes it,
  also on a run that writes nothing (each command computes the render of the
  recorded selection for this). A recorded path that is no longer rendered and
  is absent on disk leaves `files` on the next write.
- **Precedence.** When one run finds both kinds of refusal, a not-a-file target
  (exit `3`) wins over an uncommitted target (exit `1`); the message lists every
  path of both kinds. A refusal decided from the flags, the catalog and the
  setup file (exit `2`, e.g. a tool still needed by another or a replacement
  without `--replace`) comes before the target checks, so it wins over both.

Every path list in a command's output (human and `--json`) is sorted by path in
byte order.

## Tool versions {#tool-versions}

Every tool a rendered task or hook runs is installed by `mise` from the rendered
config; no task fetches and runs a package by itself. Every `mise.dev.toml` asks
for the `latest` version. Every other `mise` config file pins an exact version,
and a new bootstrap release moves those pins.

## Observability {#observability}

Local only (`observability: local-logs-only`). Diagnostics go to stderr, with
`--verbose` adding each per-file decision. No telemetry, no analytics, no
network calls home — a product non-goal. The `site` collects no visitor
analytics.

## Theming {#theming}

The `site` is **dark-only**: one theme, no light variant and no theme switch; it
does not follow the visitor's system preference. Colour roles come from
[design-system.md](design-system.md). The `cli` uses the terminal's own colours
mapped to semantic roles (design-system Terminal UX) — never fixed values.

## Release & changelog {#changelog}

Each project keeps a human-written changelog in the keep-a-changelog shape
(`changelog: keepachangelog`): an `Unreleased` section, then one section per
released version, grouped Added / Changed / Deprecated / Removed / Fixed /
Security. A release moves `Unreleased` under its version. Versions follow
semantic versioning; for `cli`, a change to the setup-file format, a removed or
renamed tool, a removed or renamed path in a tool, or a changed exit-code
meaning is a major version. A tool or a path added to a tool is a minor version.

## Reliability targets {#reliability}

`reliability: static-site-best-effort`. The `site` is static files served from
the host's edge with no stated availability target. Every page's content and
links work with client scripting off; the only scripted controls (the copy
button, the search input, the not-found page's search link and, on documentation
pages, the mobile-menu button) are hidden when scripting is off, never shown
inert — on narrow views the navigation then shows inline. The `cli` runs
locally; its reliability contract is correctness — the acceptance criteria of
each flow — not uptime.

## Disaster recovery {#disaster-recovery}

`dr: git-is-source-of-truth`. Everything the product is — source, templates,
docs, the site's content — lives in the git repository and its remote; a lost
host or package release is recovered by rebuilding from a tag. In a target
repository, every file bootstrap changed is recoverable from that repository's
own git history, per [#safety](#safety).

## Incident response {#incident}

`incident: issue-tracker-and-security-policy`. Bugs and regressions are reported
on the repository's issue tracker. Vulnerabilities are reported privately per
the repository's security policy, never in a public issue; a fix ships as a
patch release with a `Security` changelog entry.

## Web metadata {#web-metadata}

Defaults every `site` page inherits; each screen's Metadata block overrides only
what it states. Visual assets (favicon mark, social preview, theme colour) are
the design system's Brand assets.

| Field               | Value                                                                                                    |
| ------------------- | -------------------------------------------------------------------------------------------------------- |
| Site name           | bootstrap                                                                                                |
| Home address        | `https://bootstrap.virajp.dev/` — the canonical origin of every page                                     |
| Default description | One shared setup for every repository: formatters, commit gates, scanners and tasks, chosen by category. |
| Content locale      | `en`                                                                                                     |
| Social handle       | none                                                                                                     |
| Organisation        | Viraj Patel — `https://virajp.dev`; logo: the brand mark                                                 |
| Source repository   | `https://github.com/virajp/bootstrap` — the target of every "Source ↗" link                              |

## Engineering baseline {#baseline}

Defaults every doc and cycle follows; only exceptions are documented, on the
deviating doc and as an `enforcement.rules` waiver.

- **`baseline/atomic-multi-write`** — a command that writes more than one file
  is all-or-nothing: on any failure or interruption it removes what it wrote and
  restores what it moved aside, so a partial set of writes is never left
  observable.
- **`baseline/boundary-validation`** — every input crossing a boundary is
  validated against its contract and rejected, never coerced: flags and
  arguments, the setup file
  ([setup-config schema](entities/setup-config/schema.yaml)), and every `--json`
  report before it is printed.
- **`baseline/business-technical-separation`** — the setup decisions (what to
  render, which paths are targets, what needs consent) never live inside
  technical helpers (file system, git, terminal, templating), which sit in
  shared layers the decisions consume.
- **`baseline/graceful-shutdown`** — an interrupt during a writing command
  triggers the same rollback as a failure; acknowledged work is never half
  applied.
- **`baseline/structured-logs-no-pii`** — `--verbose` diagnostics are key-value,
  and never contain secrets or personal data; paths and tool names are the
  identifiers. No telemetry pipeline exists (see
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

| Environment   | Serves                  | Built from | Deployed by              |
| ------------- | ----------------------- | ---------- | ------------------------ |
| `development` | the developer's machine | any branch | never — run locally      |
| `staging`     | testers only            | `develop`  | a deliberate release act |
| `production`  | users                   | `main`     | a deliberate release act |

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
