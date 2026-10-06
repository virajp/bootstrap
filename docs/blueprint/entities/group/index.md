---
type: vwf-entity
title: Group
description: A named set of setup files shipped inside the tool, applied to a repository as a unit.
status: reviewed
implementation: none
owner: [cli]
---

# Entity: Group

## Purpose

A group is a named set of setup files shipped inside the bootstrap tool itself. It is a read-only catalog, versioned with the tool, and is never stored in the target repository. Users select optional groups at init (`--group <name>`) and add them later by name; core groups are always applied.

Used by: [Setup repository](../../flows/cli/110-setup-repository/index.md), [Add an optional group](../../flows/cli/140-add-group/index.md).

Scale: 12 groups at 1.0 (8 core, 4 optional); grows only with new tool releases.

## Out of Scope

- Language-toolchain groups (product non-goal for 1.0).
- User-defined or repository-local groups.
- Per-file selection inside a group; the group is the unit of selection.

## Lifecycle / State Machine

N/A — immutable catalog shipped with the tool. Contents change only with a new bootstrap release.

## Catalog (1.0)

The enumerated set below is the contract. Names, the core/optional split and the purposes are fixed; renaming or removing a group is a major version ([conventions](../../conventions.md#changelog)). Two optional groups, `github` and `renovate`, are named after the tool whose files they render — a recorded exemption from the no-vendor-names rule, because the name is what users type and the files are specific to that tool.

| Name           | Kind     | Purpose                                              |
| -------------- | -------- | ---------------------------------------------------- |
| `tasks`        | core     | Task runner config and base task library             |
| `format`       | core     | Formatters                                           |
| `commit-gates` | core     | Pre-commit hooks and conventional-commit rules       |
| `secret-scan`  | core     | Secret scanning                                      |
| `vuln-scan`    | core     | Vulnerability scanning                               |
| `lint`         | core     | Linting                                              |
| `git`          | core     | Ignore and attributes files                          |
| `editor`       | core     | Editor defaults                                      |
| `ai`           | optional | AI coding-assistant tooling and status line          |
| `github`       | optional | CI workflows and repository templates                |
| `renovate`     | optional | Dependency update bot config                         |
| `secrets`      | optional | Secrets provider setup                               |

## Invariants

1. No file path appears in more than one group.
2. Core groups always render; none can be deselected.
3. The core `git` group's ignore file ignores `*.bak` (backups per [backups](../../conventions.md#backups)).

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)

## Relationships

| Related entity                           | Cardinality | Ownership | On delete | Required |
| ---------------------------------------- | ----------- | --------- | --------- | -------- |
| [Setup config](../setup-config/index.md) | N–M         | reference | N/A — groups are never deleted | No |

Setup config's `groups` list names optional groups; core groups are implicit and never listed.

## Concurrency & Consistency

- Concurrent-write resolution: N/A — read-only catalog, never written at runtime.
- Uniqueness guarantees under races: N/A — read-only.
- Idempotency of each mutating action: N/A — no mutating actions.

## References

- [baseline](../../conventions.md#baseline)
- [backups](../../conventions.md#backups)
