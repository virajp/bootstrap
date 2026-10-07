---
type: vwf-entity
title: Setup Config
description: The file in a repository that records its bootstrap values and selected tools, and marks it as set up.
status: reviewed
implementation: partial
owner: [cli]
---

# Entity: Setup Config

## Purpose

The setup config records the values and the selected tools a repository was set
up with, so that later commands re-render the same setup. It is a
human-readable, hand-editable file kept in the target repository at
`.config/bootstrap.yaml`; that path is product contract. Its presence means
"this repository has been set up".

Used by: [Set up a repository](../../flows/cli/110-setup-repository/index.md),
[Add a tool](../../flows/cli/140-add-tool/index.md),
[Remove a tool](../../flows/cli/150-remove-tool/index.md),
[Select tools](../../flows/cli/160-select-tools/index.md),
[Documentation](../../flows/site/110-documentation/index.md)

Scale: N/A — a local file on the user's machine, one per set-up repository; no
scale to measure.

## Out of Scope

- Deleting the file: removal is manual and not a contract transition.
- Any record of earlier states: git history is the only one
  ([safety](../../conventions.md#safety)).

## Lifecycle / State Machine

| From    | To      | Trigger (actor/system)                                                                                                   | Guard                                                                                                                                                                                  | Side effect                                                                                                                                                                                                                                |
| ------- | ------- | ------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| absent  | present | Repo owner or agent running init, first run ([flow 110](../../flows/cli/110-setup-repository/index.md))                  | none                                                                                                                                                                                   | `format` and `version` set to the running cli's; `values` set to the init answers; `files` set to the set the running cli renders                                                                                                          |
| present | present | Repo owner or agent re-running init ([flow 110](../../flows/cli/110-setup-repository/index.md))                          | file is valid ([Validity](#validity))                                                                                                                                                  | `values` from flags, else recorded; `values.tools` as selected; rewrite per invariant 2                                                                                                                                                    |
| present | present | Repo owner or agent running [Add a tool](../../flows/cli/140-add-tool/index.md)                                          | file is valid; recorded `version` equals the running cli ([version guard](#version-guard))                                                                                             | Named tools added to `values.tools`; a replacement takes the replaced tool out; rewrite per invariant 2                                                                                                                                    |
| present | present | Repo owner or agent running [Remove a tool](../../flows/cli/150-remove-tool/index.md)                                    | file is valid; recorded `version` equals the running cli ([version guard](#version-guard))                                                                                             | Named tools taken out of `values.tools`; rewrite per invariant 2                                                                                                                                                                           |
| present | present | Repo owner running [Select tools](../../flows/cli/160-select-tools/index.md)                                    | file is valid; recorded `version` equals the running cli ([version guard](#version-guard))                                                                                             | `values.tools` and the other `values` changed as selected; rewrite per invariant 2                                                                                                                                                         |
| present | present | Repo owner hand-edits any field ([config](../../conventions.md#config))                     | none at the edit; once the edit is committed (invariant 1), the next command applies [Validity](#validity) and [Repair on read](#repair)                               | none by bootstrap; the edit takes effect on the next command                                                                                                                                                                               |

## Validity {#validity}

Every command that reads the file applies these checks. A failed check stops
the command with exit 3 ([errors](../../conventions.md#errors)), naming the
problem and the fix. Nothing is written.

- The file conforms to [schema.yaml](./schema.yaml).
- `format` is not newer than the running cli understands; the fix is "upgrade
  bootstrap".
- `version` is not newer than the running cli; the fix is "upgrade bootstrap".
  Versions are compared by semver precedence; build metadata is ignored (so
  `1.0.0+abc` equals `1.0.0`, and `1.0.0-rc.1` is older than `1.0.0`).
- Every name in `values.tools` is a [Tool](../tool/index.md) the running cli
  ships. A removed name, or any name the running catalog does not have, is
  reported with the fix to remove it from the file (for example "remove `fnox`
  from values.tools"). Nothing is guessed.
- No two tools in `values.tools` share a `max: one`
  [Tool category](../tool-category/index.md).
- Every tool in `values.tools` has its `requires` tools in `values.tools`.
- A `files` entry that breaks the schema pattern: the fix is to correct or
  remove the entry.

### Repair on read {#repair}

One hand-edit break is repaired rather than refused: a `values.tools` that
misses an unremovable tool while no other tool of its `max: one` category is
recorded. The tool is added again and the warning "added back unremovable tool
`<name>`" is reported (`warnings`). When
another tool of that category is recorded, it is a Validity failure (two tools
of one `max: one` category). The repair is a change under invariant 7: the file
is rewritten and listed `changed`.

### Write format {#write-format}

Bootstrap rewrites the whole file on a write. Comments are not kept. An older
`format` is read, never refused, and the write records the running cli's
`format`. Keys are in
schema order; `values.tools`, `values.commit_scopes` and `files` are sorted.

### Version guard {#version-guard}

Add, remove and select tools need the recorded `version` equal to the running
cli (compared as in [Validity](#validity)).

- Older recorded `version`: writes nothing, exit 1 with next command
  `bootstrap init` (a re-run), per [errors](../../conventions.md#errors).
- Absent file: exit 1 "not set up — run `bootstrap init`".

## Invariants

1. Written last of all writes in a run, so a failed run leaves no new setup
   config behind. The file is itself a target under
   [safety](../../conventions.md#safety), like any file: an uncommitted change
   to it (a hand edit) makes the run exit 1, so commit hand edits first. In
   results it is listed `created`, `changed` or `unchanged` like any other file.
2. Every write rewrites `format` and `version` to the running cli's and
   refreshes `files` to the set the running cli renders for the selected tools,
   subject to invariants 4–6.
3. `values.tools` lists every selected tool, base tools and unremovable tools
   included, with no duplicates.
4. Regardless of tool, a recorded path the running cli does not render stays in
   `files` and is reported `orphaned` by every command until the owner deletes
   it; remove never deletes an orphaned path.
5. Any recorded path absent on disk leaves `files` silently on every write.
6. `files` never lists this file; an entry for it added by a hand edit is
   dropped on the next write, without a report.
7. An `init` re-run, and an apply in select tools, rewrites the file, listed
   `changed`, whenever it differs from its render: a running `version` or
   `format` that differs from the recorded one, or a committed edit that
   changes only comments, key order or list order. Add and remove rewrite it
   only on a run that writes another change (a tool change or a repair); a run
   of theirs with nothing to add or remove and no repair writes nothing.

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)


## Relationships

| Related entity           | Cardinality | Ownership | On delete                                                               | Required |
| ------------------------ | ----------- | --------- | ----------------------------------------------------------------------- | -------- |
| [Tool](../tool/index.md) | N–M         | reference | N/A — tools ship with the cli, not across repositories; see Validity    | yes — at least the unremovable tools |

The relationship is carried by `values.tools`.

## Concurrency & Consistency

- Concurrent-write resolution: default — per
  [baseline](../../conventions.md#baseline) (single local writer)
- Idempotency of each mutating action: init re-run, add, remove and select
  tools with nothing to change do not rewrite the file (invariant 7).

## References

- [baseline](../../conventions.md#baseline)
- [config](../../conventions.md#config)
- [errors](../../conventions.md#errors)
- [safety](../../conventions.md#safety)

Retention: lives with the repository's git history. PII: none.
