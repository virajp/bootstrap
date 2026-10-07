---
type: vwf-entity
title: Tool category
description: A group of tools that do the same job, with a limit on how many a repository can select at once.
status: reviewed
implementation: none
owner: [cli]
---

# Entity: Tool category

## Purpose

A tool category groups tools that do the same job. It is a read-only catalog shipped inside bootstrap, versioned with it, and never stored in the target repository.

Scale: 13 categories at 1.0; grows only when a release adds one.

Used by: [Set up a repository](../../flows/cli/110-setup-repository/index.md), [Add a tool](../../flows/cli/140-add-tool/index.md), [Select tools](../../flows/cli/160-select-tools/index.md), [Documentation](../../flows/site/110-documentation/index.md).

## Out of Scope

- User-defined or repository-local categories.
- Choosing which tool within a category is preferred; the category only limits how many are selected.

## Lifecycle / State Machine

N/A — immutable catalog shipped with bootstrap. Contents change only with a new bootstrap release.

## Catalog (1.0)

The enumerated set below is the contract; each tool's category is recorded on the [Tool](../tool/index.md) entity.

| Name | Max | Purpose |
| ---- | --- | ------- |
| `tool-manager` | `one` | Installs the tools and runs the task library |
| `version-control` | `one` | Ignore rules and git settings |
| `git-hooks` | `one` | Runs the gates before each commit |
| `editor` | `many` | Editor defaults for the repository |
| `formatter` | `many` | Formats the repository's files |
| `linter` | `many` | Finds errors in code |
| `secret-scanner` | `one` | Finds secrets in code and commits |
| `vulnerability-scanner` | `many` | Finds known vulnerabilities in dependencies |
| `secrets-manager` | `one` | Gives the repository's secrets to its tasks |
| `ai-agent` | `many` | Settings for an AI coding agent |
| `knowledge-graph` | `one` | Builds a knowledge graph of the code |
| `agent-memory` | `one` | Memory store for AI agents |
| `forge` | `many` | Pull request and issue templates of the code host |

The table order is the display order: the init picker, the tui and the docs list categories in this order, and the `purpose` shows in all three. Changing the order is a minor version. Adding a category is a minor version; renaming or removing one, or changing `max`, is a major version ([changelog](../../conventions.md#changelog)).

## Invariants

1. In a `max: many` category, any number of its tools may be selected together.
2. In a `max: one` category, a new tool chosen against the tool the repository already holds is a replacement of the held one. Two tools of one `max: one` category named in a single request (init `--tool`, add) are an error and exit 2 ([errors](../../conventions.md#errors)).
3. A replacement needs the user's consent: an interactive prompt at the moment of selection, or the `--replace` flag in any mode (`--replace` is consent without a prompt). `-y` is never consent, and a `-y` run that needs a replacement without `--replace` exits 2 ([config](../../conventions.md#config)).
4. A tool whose `replaceable` is false (see [Tool](../tool/index.md)) is never replaced. In the command-line commands (init, add) an attempt exits 2 ([errors](../../conventions.md#errors)); in the tui the other tools of its category are shown locked, and "none" stays allowed when the tool is removable.
5. A `max: one` category holds at most one tool with `default: "on"`, and at most one tool with `removable: false`. Catalog releases must not ship defaults that break the limit.
6. Every category holds at least one tool.
7. A `max: one` category whose tool is unremovable always has that tool selected.

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)

## Relationships

| Related entity           | Cardinality | Ownership | On delete                         | Required |
| ------------------------ | ----------- | --------- | --------------------------------- | -------- |
| [Tool](../tool/index.md) | 1–N         | reference | a release that drops a category drops its tools: a recorded selection that names one exits 3 naming the fix | Yes      |

## Concurrency & Consistency

- Concurrent-write resolution: N/A — read-only catalog, never written at runtime.
- Uniqueness guarantees under races: N/A — read-only.
- Idempotency of each mutating action: N/A — no mutating actions.

## References

- [errors](../../conventions.md#errors)
- [config](../../conventions.md#config)
- [changelog](../../conventions.md#changelog)
- [baseline](../../conventions.md#baseline)
