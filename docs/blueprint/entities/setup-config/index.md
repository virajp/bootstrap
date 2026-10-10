---
type: vwf-entity
title: Setup Config
description: The file in a repository that records its bootstrap values and
  selected tools, and marks it as set up.
status: reviewed
implementation: partial
owner: [ cli ]
---

# Entity: Setup Config

## Purpose

The setup config records the values and the selected tools a repository was set
up with. It is a human-readable, hand-editable file kept in the target
repository at `.config/bootstrap.yaml`; that path is product contract. Its
presence means "this repository has been set up".

Used by: [Set up a repository](../../flows/cli/110-setup-repository/index.md),
[Add a tool](../../flows/cli/140-add-tool/index.md),
[Remove a tool](../../flows/cli/150-remove-tool/index.md),
[Select tools](../../flows/cli/160-select-tools/index.md),
[Show the setup](../../flows/cli/170-show-setup/index.md),
[Home](../../flows/site/100-home/index.md),
[Documentation](../../flows/site/110-documentation/index.md)

Scale: N/A — a local file on the user's machine, one per set-up repository; no
scale to measure.

## Out of Scope

- Deleting the file: removal is manual and not a contract transition.
- Any record of earlier states: git history is the only one
  ([safety](../../conventions.md#safety)).

## Lifecycle / State Machine

| From    | To      | Trigger (actor/system)                                                                                                   | Guard                                                                                                                                                                   | Side effect                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------- | ------- | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| absent  | present | Repo owner, AI agent or CI running init, first run ([flow 110](../../flows/cli/110-setup-repository/index.md))           | none                                                                                                                                                                    | `format` and `version` set to the running cli's; `values` set to the values from flags, defaults and the remote `origin`, with `values.dependencies` derived; `files` set to the set the running cli renders                                                                                                                                                                                                                                        |
| absent  | present | Repo owner running [Select tools](../../flows/cli/160-select-tools/index.md) on a repository not set up (first-run mode) | none                                                                                                                                                                    | `format` and `version` set to the running cli's; `values` set as applied, with `values.dependencies` derived; `files` set to the set the running cli renders                                                                                                                                                                                                                                                                                        |
| present | present | Repo owner, AI agent or CI re-running init ([flow 110](../../flows/cli/110-setup-repository/index.md))                   | file is valid ([Validity](#validity))                                                                                                                                   | `values` from flags, else recorded; `values.tools` as selected; `values.dependencies` derived and pruned of those no parent needs (invariant 8); rewrite per invariant 2                                                                                                                                                                                                                                                                            |
| present | present | Repo owner, AI agent or CI running [Add a tool](../../flows/cli/140-add-tool/index.md)                                   | file is valid; recorded `version` equals the running cli ([version guard](#version-guard))                                                                              | Named tools added to `values.tools`, a named tool already in `values.dependencies` moving from it to `values.tools` (it becomes direct); the tools they require added to `values.dependencies`; a replacement (`--replace`) takes the replaced tool out and moves its parents to the replacement, pruning dependencies no parent needs; a tool [Repair on read](#repair) adds back is kept unless the same run replaces it; rewrite per invariant 2 |
| present | present | Repo owner, AI agent or CI running [Remove a tool](../../flows/cli/150-remove-tool/index.md)                             | file is valid; recorded `version` equals the running cli ([version guard](#version-guard))                                                                              | Named tools taken out of `values.tools`; dependencies no remaining tool needs pruned from `values.dependencies`, parents recomputed; a tool [Repair on read](#repair) adds back is kept; rewrite per invariant 2                                                                                                                                                                                                                                    |
| present | present | Repo owner running [Select tools](../../flows/cli/160-select-tools/index.md)                                             | file is valid; recorded `version` equals the running cli ([version guard](#version-guard))                                                                              | `values.tools` and the other `values` changed as selected; `values.dependencies` derived (invariant 8); rewrite per invariant 2                                                                                                                                                                                                                                                                                                                     |
| present | present | Repo owner hand-edits any field ([config](../../conventions.md#config))                                                  | none at the edit; once the edit is committed (invariant 1), the next command applies [Validity](#validity) and [Repair on read](#repair) (`show` applies Validity only) | none by bootstrap; the edit takes effect on the next command                                                                                                                                                                                                                                                                                                                                                                                        |

## Validity {#validity}

Every command that reads the file applies these checks. A failed check stops the
command with exit 3 ([errors](../../conventions.md#errors)), naming the problem
and the fix. Nothing is written.

- The file conforms to [schema.yaml](./schema.yaml); for a `files` entry that
  breaks the pattern, the fix is to correct or remove the entry.
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
  [Tool category](../tool-category/index.md); the fix is "remove `<a>` or `<b>`
  from values.tools", naming the two tools.
- No `max: one` category whose unremovable tool has `replaceable` false is
  missing that tool from `values.tools` while it holds another tool in its
  place; the fix is "replace `<other>` with `<tool>` in values.tools", where
  `<tool>` is the unremovable tool and `<other>` the tool in its place. A
  required tool a selected tool lacks is not a Validity failure: it is a
  dependency, repaired on read (below).

### Repair on read {#repair}

Two hand-edit breaks are repaired rather than refused. Both are changes under
invariant 7: the file is rewritten and listed `changed`, and each needs consent
([config](../../conventions.md#config)).

- A non-empty `values.tools` that misses an unremovable tool, unless its
  category is `max: one` and another tool of that category is recorded. The tool
  is added again and the warning "added back unremovable tool `<name>`" is
  reported (`warnings`).
- A dependency that a selected tool needs and that `values.dependencies` lacks.
  Dependencies are derived from `values.tools` and the Tool catalog on every
  read; the missing one is added back (the default, first alternative of an
  unmet `requires` slot) and the warning "added back dependency `<name>`" is
  reported.

`show` never repairs: it reports a missing unremovable tool or a missing
dependency as an error with exit 1
([Show the setup](../../flows/cli/170-show-setup/index.md)).

A wrong parent list, a recorded dependency no parent needs, a name in
`values.dependencies` the running catalog does not have, or a name recorded in
both `values.tools` and `values.dependencies` is not a read failure: it is
corrected or pruned on the next write. An unknown dependency name is reported
with the warning "dropped unknown dependency `<name>`".

### Write format {#write-format}

Bootstrap rewrites the whole file on a write. Comments are not kept. An older
`format` is read, never refused, and the write records the running cli's
`format`. Keys are in schema order; `values.tools`, `values.commit_scopes` and
`files` are sorted; so are the keys of `values.dependencies` and each parent
list.

### Version guard {#version-guard}

Add, remove and select tools need the recorded `version` equal to the running
cli (compared as in [Validity](#validity)).

- Older recorded `version`: writes nothing, exit 1 with next command
  `bootstrap init` (a re-run), per [errors](../../conventions.md#errors).
  [Validity](#validity) is checked first: a file that fails Validity exits 3,
  also when its `version` is older.
- Absent file: add, remove and `show` exit 1 "not set up — run `bootstrap init`"
  (`show` takes no version guard otherwise). Select tools (the tui) instead
  opens in first-run mode
  ([flow 160](../../flows/cli/160-select-tools/index.md)).

## Invariants

1. Written last of all writes in a run, so a failed run leaves no new setup
   config behind. The file is itself a target under
   [safety](../../conventions.md#safety), like any file: when its render differs
   from an uncommitted working copy (a hand edit), the run exits 1, so commit
   hand edits first. In results it is listed `created`, `changed` or `unchanged`
   like any other file.
2. Every write rewrites `format` and `version` to the running cli's. Init, add
   and an apply in select tools refresh `files` to the set the running cli
   renders for the selected tools; remove only takes out the paths of the
   removed tools and of the dependencies it prunes and adds none, except the
   paths of a tool or dependency that [Repair on read](#repair) adds back. All
   subject to invariants 4–6.
3. `values.tools` lists every direct tool, unremovable tools included, with no
   duplicates; a name in `values.tools` is never in `values.dependencies`.
4. Regardless of tool, a recorded path the running cli does not render stays in
   `files` and is reported `orphaned` by every command until the owner deletes
   it; remove never deletes an orphaned path.
5. A recorded path that the running cli no longer renders and that is absent on
   disk leaves `files` silently on every write.
6. `files` never lists this file; an entry for it added by a hand edit is
   dropped on the next write, without a report.
7. An `init` re-run, and an apply in select tools, rewrites the file, listed
   `changed`, whenever it differs from its render: a running `version` or
   `format` that differs from the recorded one, or a committed edit that changes
   only comments, key order or list order. Add and remove rewrite it only on a
   run that writes another change (a tool change or a repair); a run of theirs
   with nothing to add or remove and no repair writes nothing.
8. `values.dependencies` is derived, never authoritative: every command that
   reads the file computes it from `values.tools` and the Tool catalog. It holds
   exactly the tools that a selected tool, direct or dependency, requires and
   that are not direct, each with its sorted parents. A file without the key
   reads as an empty map, so Repair on read adds the missing dependencies; every
   write records the derived result.

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)

## Relationships

| Related entity           | Cardinality | Ownership | On delete                                                            | Required                             |
| ------------------------ | ----------- | --------- | -------------------------------------------------------------------- | ------------------------------------ |
| [Tool](../tool/index.md) | N–M         | reference | N/A — tools ship with the cli, not across repositories; see Validity | yes — at least the unremovable tools |

The relationship is carried by `values.tools` and `values.dependencies`.

## Concurrency & Consistency

- Concurrent-write resolution: default — per
  [baseline](../../conventions.md#baseline) (single local writer)
- Idempotency of each mutating action: an init re-run and select tools do not
  rewrite the file when it does not differ from its render; add and remove do
  not rewrite it on a run with no tool change and no repair (invariant 7).

## References

- [baseline](../../conventions.md#baseline)
- [config](../../conventions.md#config)
- [errors](../../conventions.md#errors)
- [safety](../../conventions.md#safety)

Retention: lives with the repository's git history. PII: none.
