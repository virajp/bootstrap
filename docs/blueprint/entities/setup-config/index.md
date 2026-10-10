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

| From    | To      | Trigger (actor/system)                                                                                                   | Guard                                                                                                                                                                                                                                                                                | Side effect                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------- | ------- | ------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| absent  | present | Repo owner, AI agent or CI running init, first run ([flow 110](../../flows/cli/110-setup-repository/index.md))           | none                                                                                                                                                                                                                                                                                 | `format` and `version` set to the running cli's; `values` set to the values from flags, defaults and the remote `origin`, with `values.dependencies` derived; `files` set to the set the running cli renders                                                                                                                                                                                                                                        |
| absent  | present | Repo owner running [Select tools](../../flows/cli/160-select-tools/index.md) on a repository not set up (first-run mode) | none                                                                                                                                                                                                                                                                                 | `format` and `version` set to the running cli's; `values` set as applied, with `values.dependencies` derived; `files` set to the set the running cli renders                                                                                                                                                                                                                                                                                        |
| present | present | Repo owner, AI agent or CI re-running init ([flow 110](../../flows/cli/110-setup-repository/index.md))                   | file is valid ([Validity](#validity))                                                                                                                                                                                                                                                | `values` from flags, else recorded; `values.tools` as selected; `values.dependencies` derived; stale dependencies and wrong parent lists fixed only when the run writes for another change (invariants 7, 8); rewrite per invariant 2                                                                                                                                                                                                               |
| present | present | Repo owner, AI agent or CI running [Add a tool](../../flows/cli/140-add-tool/index.md)                                   | file is valid; recorded `version` equals the running cli ([version guard](#version-guard)); an add that would put two tools of one `max: one` category into `values.tools` without `--replace` exits 2 and writes nothing ([flow 140](../../flows/cli/140-add-tool/index.md) step 4) | Named tools added to `values.tools`, a named tool already in `values.dependencies` moving from it to `values.tools` (it becomes direct); the tools they require added to `values.dependencies`; a replacement (`--replace`) takes the replaced tool out and moves its parents to the replacement, pruning dependencies no parent needs; a tool [Repair on read](#repair) adds back is kept unless the same run replaces it; rewrite per invariant 2 |
| present | present | Repo owner, AI agent or CI running [Remove a tool](../../flows/cli/150-remove-tool/index.md)                             | file is valid; recorded `version` equals the running cli ([version guard](#version-guard)); a remove that names `git` or `pre-commit` (an unremovable tool) exits 2 and writes nothing ([flow 150](../../flows/cli/150-remove-tool/index.md) usage check)                            | Named tools taken out of `values.tools`; dependencies no remaining tool needs pruned from `values.dependencies`, parents recomputed (also fixing any bookkeeping, invariant 7); a tool [Repair on read](#repair) adds back is kept; rewrite per invariant 2                                                                                                                                                                                         |
| present | present | Repo owner running [Select tools](../../flows/cli/160-select-tools/index.md)                                             | file is valid; recorded `version` equals the running cli ([version guard](#version-guard))                                                                                                                                                                                           | `values.tools` and the other `values` changed as selected; `values.dependencies` derived (invariant 8); rewrite per invariant 2                                                                                                                                                                                                                                                                                                                     |
| present | present | Repo owner hand-edits any field ([config](../../conventions.md#config))                                                  | none at the edit; once the edit is committed (invariant 1), the next command applies [Validity](#validity) and [Repair on read](#repair) (`show` applies Validity only)                                                                                                              | none by bootstrap; the edit takes effect on the next command                                                                                                                                                                                                                                                                                                                                                                                        |

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
- `values.tools` never holds `mise`, a hidden tool that is always applied and
  never recorded; a recorded `mise` fails like a name the running catalog does
  not have, with the fix "remove `mise` from values.tools".
- No two tools in `values.tools` share a `max: one`
  [Tool category](../tool-category/index.md); the fix is "remove `<a>` or `<b>`
  from values.tools", naming the two tools.

### Repair on read {#repair}

Three hand-edit breaks are repaired rather than refused. All are changes under
invariant 7: the file is rewritten and listed `changed`, and each needs consent
([config](../../conventions.md#config)), even when it is the only change. Every
writing command (init re-run, add, remove, select tools apply) repairs only what
the request needs; a missing piece left unneeded is never added back and never
reported.

- A non-empty `values.tools` that misses `git` or `pre-commit` (the unremovable
  tools), unless another tool of its `max: one` category is recorded. Only these
  two are added back; the warning "added back unremovable tool `<name>`" is
  reported (`warnings`).
- A dependency that a selected tool needs and that `values.dependencies` lacks.
  The missing one is added back (the default, first alternative of an unmet
  `requires` slot) and the warning "added back dependency `<name>`" is reported.
- A name in `values.dependencies` that the running catalog does not have is
  dropped. The warning "dropped unknown dependency `<name>`" is reported
  (`warnings`); `show` does not warn.

`show` never repairs: it reports a missing unremovable tool or a missing
dependency as an error with exit 1
([Show the setup](../../flows/cli/170-show-setup/index.md)).

Bookkeeping differences are neither a read failure nor a change on their own;
see invariant 7.

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
   duplicates. Every write records a name in `values.tools` or in
   `values.dependencies`, never in both; a hand-edited file can hold both
   (invariant 7).
4. Regardless of tool, a recorded path the running cli does not render stays in
   `files` and is reported `orphaned` by every command until the owner deletes
   it; remove never deletes an orphaned path.
5. A recorded path the running cli no longer renders and that is absent on disk
   is bookkeeping (invariant 7).
6. `files` never lists this file; an entry for it added by a hand edit is
   dropped on the next write, without a report.
7. An `init` re-run, and an apply in select tools, rewrites the file, listed
   `changed`, for a change: a running `version` or `format` that differs from
   the recorded one, a committed edit that changes only comments, key order or
   list order, a repair or a tool change. Add and remove rewrite it only on a
   run that writes another change (a tool change or a repair); a run of theirs
   with nothing to add or remove and no repair writes nothing. Bookkeeping alone
   is not a reason to rewrite, in every writing command: a recorded path the
   running cli no longer renders and that is absent on disk (invariant 5), a
   stale recorded dependency (a name the running catalog has that no tool of the
   selection needs and whose loss of parents the request did not cause), a wrong
   parent list in `values.dependencies`, or a name recorded in both
   `values.tools` and `values.dependencies`. A dependency name the running
   catalog does not have is not bookkeeping: it stays a repair
   ([Repair on read](#repair), third bullet), a change that needs consent even
   alone. When the file is rewritten for another change, the run also fixes the
   bookkeeping: it drops the stale path, prunes the stale dependency (its files
   deleted and reported as usual, listed in `dependencies_pruned`), corrects the
   parent lists and keeps a doubly recorded name in `values.tools` only. When
   bookkeeping is the only difference, nothing is written, no consent is asked
   and the exit is 0 (an init re-run reports the "nothing to change" result). A
   dependency left without a parent by the request (a drop, a remove, a
   replacement) is a real change and is pruned. `orphaned` paths (present on
   disk) are unchanged: reported, never a change.
8. `values.dependencies` is derived, never authoritative: every command that
   reads the file computes it from `values.tools` (plus any unremovable tool
   that [Repair on read](#repair) would add back) and the Tool catalog. It holds
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
  rewrite the file when it does not differ from its render, bookkeeping aside;
  add and remove do not rewrite it on a run with no tool change and no repair
  (invariant 7).

## References

- [baseline](../../conventions.md#baseline)
- [config](../../conventions.md#config)
- [errors](../../conventions.md#errors)
- [safety](../../conventions.md#safety)

Retention: lives with the repository's git history. PII: none.
