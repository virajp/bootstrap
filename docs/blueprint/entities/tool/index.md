---
type: vwf-entity
title: Tool
description: A named set of setup files for one tool, shipped inside bootstrap and applied to a repository as a unit.
status: reviewed
implementation: partial
owner: [cli]
---

# Entity: Tool

## Purpose

A tool is a named set of setup files for one tool, shipped inside bootstrap itself. It is a read-only catalog, versioned with bootstrap, and is never stored in the target repository. `core` is not a tool: it is the fixed set of tools always applied with no flag.

Used by: [Setup repository](../../flows/cli/110-setup-repository/index.md), [Check for drift](../../flows/cli/120-check-drift/index.md), [Update a repository](../../flows/cli/130-update-repository/index.md), [Add a tool](../../flows/cli/140-add-tool/index.md), [Remove a tool](../../flows/cli/150-remove-tool/index.md), [Documentation](../../flows/site/110-documentation/index.md).

Scale: 15 tools at 1.0 (13 core, 2 non-core); grows only with new bootstrap releases.

## Out of Scope

- User-defined or repository-local tools.
- Per-file selection inside a tool; the tool is the unit of selection.

## Lifecycle / State Machine

N/A — immutable catalog shipped with bootstrap. Contents change only with a new bootstrap release.

## Catalog (1.0)

The enumerated set below is the contract. Core tools are always applied with no flag and cannot be deselected or removed. Non-core tools are chosen with `--tool <name>` (repeatable) at init and with `bootstrap add <tool>` / `bootstrap remove <tool>` later; `github` and `gitlab` may be selected together. Renaming or removing a tool is a major version; adding a tool is minor ([conventions](../../conventions.md#changelog)). Recorded exemption from the no-vendor-names rule, by user decision: every tool is named after the tool whose files it renders, and its target paths name that tool's config files.

| Name             | Kind     | Purpose                                              |
| ---------------- | -------- | ---------------------------------------------------- |
| `mise`           | core     | Task runner and tool versions                        |
| `git`            | core     | Ignore rules and git config                          |
| `pre-commit`     | core     | Commit hooks and conventional-commit rules           |
| `vscode`         | core     | Editor defaults                                      |
| `dprint`         | core     | Formatter                                            |
| `taplo`          | core     | TOML formatter                                       |
| `gitleaks`       | core     | Secret scanning                                      |
| `grype`          | core     | Vulnerability scanning                               |
| `virajp-linter`  | core     | Linting                                              |
| `claude`         | core     | AI coding-assistant settings and status line         |
| `graphify`       | core     | Code knowledge graph                                 |
| `mempalace`      | core     | Memory store config                                  |
| `fnox`           | core     | Secrets provider                                     |
| `github`         | non-core | Pull-request and issue templates                     |
| `gitlab`         | non-core | Merge-request and issue templates                    |

### Target paths (1.0)

Each tool's target paths are pinned: paths relative to the target repository root, where bootstrap writes them. File contents are not part of the contract, except invariant 3. Paths marked `tasks/` are under `.config/mise/tasks/`. The setup config `.config/bootstrap.yaml` belongs to no tool; init writes it. Adding a path is a minor version; removing or renaming a path is a major version ([changelog](../../conventions.md#changelog)).

The `mise` dependency dispatchers (`tasks/setup/deps/<verb>`) run the matching `setup/deps/<verb>/<tool>` task of every selected tool that has one; no 1.0 tool has one, so in 1.0 they do nothing.

| Tool            | Paths |
| --------------- | ----- |
| `mise`          | `.config/mise.toml`, `.config/miserc.toml`, `.config/mise/conf.d/_base/mise.toml`, `.config/mise/conf.d/_base/mise.dev.toml`, `.config/mise/conf.d/_base/mise.ci.toml`, `tasks/_scripts/helpers`, `tasks/_scripts/checks`, `tasks/_scripts/merge`, `tasks/_scripts/placeholder`, `tasks/code/all`, `tasks/code/count`, `tasks/code/worktrees`, `tasks/code/merge/develop`, `tasks/code/merge/main`, `tasks/setup/all`, `tasks/setup/mise`, `tasks/setup/worktree`, `tasks/setup/cleanup`, `tasks/setup/deps/all`, `tasks/setup/deps/install`, `tasks/setup/deps/outdated`, `tasks/setup/deps/audit`, `tasks/setup/deps/upgrade`, `tasks/setup/deps/cleanup` |
| `git`           | `.gitignore`, `tasks/code/git-config` |
| `pre-commit`    | `.config/pre-commit-config.yaml`, `.config/git-conventional-commits.yaml`, `tasks/code/precommit`, `tasks/setup/precommit` |
| `vscode`        | `.vscode/settings.json`, `.vscode/extensions.json` |
| `dprint`        | `dprint.json`, `.config/dprint.json`, `tasks/code/format` |
| `taplo`         | `.config/taplo.toml` |
| `gitleaks`      | `.config/gitleaks.toml`, `tasks/code/sec` |
| `grype`         | `.config/grype.yaml` |
| `virajp-linter` | `.config/linter.yaml`, `eslint.config.mjs`, `tasks/code/lint` |
| `claude`        | `.config/mise/conf.d/ai/mise.dev.toml`, `.config/claude-status.json`, `tasks/setup/ai` |
| `graphify`      | `tasks/code/graph`, `.graphifyignore` |
| `mempalace`     | `mempalace.yaml` |
| `fnox`          | `.config/fnox.toml`, `tasks/setup/secrets` |
| `github`        | `.github/pull_request_template.md`, `.github/ISSUE_TEMPLATE/bug.yml`, `.github/ISSUE_TEMPLATE/feature.yml`, `.github/ISSUE_TEMPLATE/config.yml` |
| `gitlab`        | `.gitlab/merge_request_templates/Default.md`, `.gitlab/issue_templates/Bug.md`, `.gitlab/issue_templates/Feature.md` |

## Invariants

1. No file path appears in more than one tool.
2. Core tools always render; none can be deselected or removed.
3. The `git` tool's `.gitignore` ignores `*.bak` (backups per [backups](../../conventions.md#backups)).

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)

## Relationships

| Related entity                           | Cardinality | Ownership | On delete | Required |
| ---------------------------------------- | ----------- | --------- | --------- | -------- |
| [Setup config](../setup-config/index.md) | N–M         | reference | N/A — tools are never deleted | No |

## Concurrency & Consistency

- Concurrent-write resolution: N/A — read-only catalog, never written at runtime.
- Uniqueness guarantees under races: N/A — read-only.
- Idempotency of each mutating action: N/A — no mutating actions.

## References

- [baseline](../../conventions.md#baseline)
- [backups](../../conventions.md#backups)
- [changelog](../../conventions.md#changelog)
