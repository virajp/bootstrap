---
type: vwf-entity
title: Tool category
description: A group of tools that do the same job, with a limit on how many a
  repository can select at once.
status: reviewed
implementation: partial
owner: [ cli ]
---

# Entity: Tool category

## Purpose

A tool category groups tools that do the same job. It is a read-only catalog
shipped inside bootstrap, versioned with it, and never stored in the target
repository. A tool can belong to several categories (ktlint and ruff are both
formatter and linter).

Scale: 17 categories at 1.0; grows only when a release adds one.

Used by: [Set up a repository](../../flows/cli/110-setup-repository/index.md),
[Add a tool](../../flows/cli/140-add-tool/index.md),
[Select tools](../../flows/cli/160-select-tools/index.md),
[Show the setup](../../flows/cli/170-show-setup/index.md),
[Home](../../flows/site/100-home/index.md),
[Documentation](../../flows/site/110-documentation/index.md).

## Out of Scope

- User-defined or repository-local categories.
- Choosing which tool within a category is preferred; the category only limits
  how many are selected.

## Lifecycle / State Machine

N/A — immutable catalog shipped with bootstrap. Contents change only with a new
bootstrap release.

## Catalog (1.0)

The enumerated set below is the contract; each tool's categories are recorded on
the [Tool](../tool/index.md) entity.

| Name                    | Max    | Purpose                                                                    |
| ----------------------- | ------ | -------------------------------------------------------------------------- |
| `tool-manager`          | `one`  | Installs the tools and runs the task library                               |
| `version-control`       | `one`  | Ignore rules and git settings                                              |
| `git-hooks`             | `one`  | Runs the gates before each commit                                          |
| `editor`                | `many` | Editor defaults for the repository                                         |
| `runtime`               | `many` | Language runtimes and command-line helpers that other tools run on         |
| `package-manager`       | `many` | Package managers that install a runtime's dependencies                     |
| `framework`             | `many` | Frameworks whose editor support and command-line tools the repository uses |
| `formatter`             | `many` | Formats the repository's files                                             |
| `linter`                | `many` | Finds errors in code                                                       |
| `secret-scanner`        | `one`  | Finds secrets in code and commits                                          |
| `vulnerability-scanner` | `many` | Finds known vulnerabilities in dependencies                                |
| `secrets-manager`       | `many` | Gives the repository's secrets to its tasks                                |
| `ai-agent`              | `many` | Settings for an AI coding agent                                            |
| `knowledge-graph`       | `one`  | Builds a knowledge graph of the code                                       |
| `agent-memory`          | `one`  | Memory store for AI agents                                                 |
| `local-services`        | `many` | Local services the project runs against, such as emulators                 |
| `forge`                 | `many` | Pull request and issue templates of the code host                          |

The table order is the display order: the tui and the docs list categories in
this order, and the `purpose` shows in both. A tool in several categories
appears in each of its category groups; selecting it in one group selects it in
all of them. A category whose every tool is hidden (`tool-manager`, whose only
tool is `mise`) is not shown: not in the tui, the docs or `show`. Changing the
order is a minor version. Adding a category is a minor version; renaming or
removing one, or changing `max`, is a major version
([changelog](../../conventions.md#changelog)).

## Invariants

1. In a `max: many` category, any number of its tools may be selected together.
2. Two tools of one `max: one` category named in a single request (init
   `--tool`, add) are an error and exit 2
   ([errors](../../conventions.md#errors)).
3. Selecting a tool in a `max: one` category that holds another tool is a
   replacement, and a replacement needs its own consent: in the tui, the
   in-place prompt "replace <old> with <new>?"; in init and add, the `--replace`
   flag, without which the run exits 2. The general consent to apply the changes
   follows per [config](../../conventions.md#config); `-y` and the consent
   prompt of init and add are never consent to replace. A replacement exists
   only in a `max: one` category (and among the runtimes of one engine, owned by
   the [Tool](../tool/index.md#requires-and-dependencies) entity); in a
   `max: many` category a user adds one tool and removes the other. A tool in
   several categories is replaced in a `max: one` category only with respect to
   that category: the replacement concerns that category alone, and the tool
   stays selected for its other categories.
4. Every shown tool can be replaced within its `max: one` category.
   `removable: false` (see [Tool](../tool/index.md)) does not block replacement:
   `git` and `pre-commit` are unremovable but replaceable by another tool of
   their `max: one` category. In the tui "none" stays allowed unless the
   category's tool is unremovable.
5. A `max: one` category holds at most one tool with `default: "on"`, and at
   most one tool with `removable: false`. Catalog releases must not ship
   defaults that break the limit.
6. Every category holds at least one tool.
7. A `max: one` category whose tool is unremovable is never empty: that tool
   stays selected unless it is replaced by another tool of the category. A
   missing unremovable tool is added back per
   [Repair on read](../setup-config/index.md#repair) unless its `max: one`
   category holds another tool.

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)

## Relationships

| Related entity           | Cardinality | Ownership | On delete                                                                                                                                                                                                          | Required |
| ------------------------ | ----------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- |
| [Tool](../tool/index.md) | N–M         | reference | a release that drops a category drops only the tools with no other category (a recorded selection that names one exits 3 naming the fix); a tool in another category stays selected and loses the dropped category | Yes      |

## Concurrency & Consistency

- Concurrent-write resolution: N/A — read-only catalog, never written at
  runtime.
- Uniqueness guarantees under races: N/A — read-only.
- Idempotency of each mutating action: N/A — no mutating actions.

## References

- [errors](../../conventions.md#errors)
- [config](../../conventions.md#config)
- [changelog](../../conventions.md#changelog)
- [baseline](../../conventions.md#baseline)
