---
type: vwf-flow
title: Remove a tool
description: One command removes removable tools from a set-up repository,
  deleting their files and nothing half-written; git history keeps every deleted
  file.
status: reviewed
implementation: none
---

# Flow: Remove a tool

## Purpose

`bootstrap remove <tool>...` removes one or more removable tools from a
repository that is already set up: it deletes their files, and prunes each
dependency that no remaining tool needs; a tool that would dangle without a
named tool is never removed silently: it must be named too
([Tool](../../../entities/tool/index.md#requires-and-dependencies)). It never
commits.

Serves: [Zero setup drift](../../../product.md#goal-zero-drift)

## Trigger & Actors

| Actor                            | May trigger                                                                                                                                                 | Authorization                         | Audit-recorded                                                                 |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- | ------------------------------------------------------------------------------ |
| Repo owner (any terminal)        | the command `bootstrap remove <tool> [<tool>...]` with flags; the interactive alternative is `bootstrap tui` ([Select tools](../160-select-tools/index.md)) | write access to the working directory | no — the product has no audit foundation; git history keeps every deleted file |
| AI agent or CI (non-interactive) | `bootstrap remove <tool>...`; `-y` is necessary for a run that changes anything, since no prompt is possible without terminals or with `--json`             | write access to the working directory | no — the product has no audit foundation; git history keeps every deleted file |

## Steps

Usage errors come first: no tool name given, an unknown tool name (`mise` is
one: a hidden tool is never nameable), a tool whose `removable` value in the
[Tool](../../../entities/tool/index.md) catalog is false (`git`, `pre-commit`),
or an unknown flag or a bad flag value, exits 2 with short usage before the
preflight and before `.config/bootstrap.yaml` is read, all reported together
([errors](../../../conventions.md#errors)). Nothing is written.

An old tool name (one the catalog lists in a tool's `renamed_from`) typed on the
command line is an unknown tool name: a usage error, exit 2, whose error names
the new name as the fix, "unknown tool `<old>` — use `<new>`". Only a recorded
old name is repaired (step 2).

1. Remove runs, after the usage check, the preflight of
   [Set up a repository](../110-setup-repository/index.md) step 1.
2. Remove reads `.config/bootstrap.yaml`
   ([Setup config](../../../entities/setup-config/index.md)). Absent → writes
   nothing, exits 1, next command `bootstrap init`. Invalid (a newer `version`
   or `format` included) → writes nothing, exits 3 per
   [Validity](../../../entities/setup-config/index.md#validity); an older
   `version` → exits 1 per
   [Version guard](../../../entities/setup-config/index.md#version-guard). A
   missing unremovable tool (`git` or `pre-commit`), or a missing dependency
   that a selected tool needs, is repaired with a warning per
   [Repair on read](../../../entities/setup-config/index.md#repair); a missing
   dependency that no tool of the selection after the request needs is not
   repaired and not reported. A recorded old name of a renamed tool (in
   `values.tools` or `values.dependencies`) is read as the new name and
   rewritten under it as a repair (a change that needs consent), with the
   warning "renamed tool `<old>` to `<new>`". A `values.dependencies` name the
   running catalog does not have is dropped as a repair (a change on the list
   that needs consent, so the list is not empty) at the step 8 rewrite, with the
   warning "dropped unknown dependency `<name>`".
3. Remove validates each name. A name given more than once is used once, with no
   error. A recorded tool whose `host_os` is not the running host may be removed
   like any other. A name is selected when it is in `values.tools` (direct) or
   in `values.dependencies` (a dependency). The selection after the request is
   the direct tools without the named ones, plus the dependencies they need
   (derived per the requires check of
   [Tool](../../../entities/tool/index.md#requires-and-dependencies)). Each name
   falls in one case:

   | Case                                                                                                                                                                                   | Result                                                                                                                                                          |
   | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
   | Selected directly, and no tool of the selection after the request would dangle without it                                                                                              | `removed`                                                                                                                                                       |
   | Selected (direct or dependency), and a tool of the selection after the request would dangle without it (it requires the named tool, or an engine whose last selected runtime is named) | exit 2 naming every direct tool that would dangle (`pre-commit` for `python`; `effect` and `pnpm` for `node`), unless all of them are named in the same request |
   | Selected only as a dependency, and no tool of the selection after the request would dangle without it                                                                                  | `not_added`, like an unselected name; pruned per the rules below and listed in `dependencies_pruned` when pruned                                                |
   | Selected neither way                                                                                                                                                                   | `not_added`, changes nothing                                                                                                                                    |

   Which tools would dangle follows the
   [Dangling rule](../../../entities/tool/index.md#requires-and-dependencies):
   it is transitive, so a package manager whose runtime leaves (`pnpm` or `yarn`
   when `node` leaves) dangles too, and a runtime leaves without blocking
   anything while another runtime of the same engine stays selected for each of
   its parents (their links move to it). When several named tools are blocked,
   every blocked name is reported with its parents in the one exit 2 error, in
   step order; they are reported together after the usage check
   ([errors](../../../conventions.md#errors)). The `next_command` is the same
   command plus every named parent, sorted by name (for example
   `bootstrap remove node -y` → `bootstrap remove node effect pnpm -y`); naming
   them removes them all. When an unremovable tool (`git` or `pre-commit`) is a
   blocking parent of any named tool, the error has no next command, even if
   other names could be removed.

   A dependency that the request leaves without a parent (a tool of the recorded
   selection needed it, no tool of the selection after the request does) is
   pruned: it is on the list of changes, like a removed tool. A stale dependency
   (no tool of the recorded selection needs it, so the request did not cause its
   loss of parents) is bookkeeping, like a wrong parent list in
   `values.dependencies`: when the run writes for another change, the stale
   dependency is pruned too (its files deleted and reported as usual, listed in
   `dependencies_pruned`) and the parent lists are corrected; on its own it is
   not a change. If no named tool is selected, the request leaves no dependency
   without a parent and step 2 made no repair, the list of changes is empty
   (bookkeeping alone does not fill it): remove first renders the recorded
   selection in memory to find the `orphaned` paths (as
   [Add a tool](../140-add-tool/index.md) step 3 does), writes nothing, shows no
   prompt, reports them and exits 0.
4. Remove renders in memory the full file set for the new selection with the
   recorded values (exact mise pins already in the repository are kept,
   [tool versions](../../../conventions.md#tool-versions)) and compares each
   file with the working copy: the new-selection render judges the files to
   keep, create or change. Remove also renders the recorded selection: that
   render tells a removed tool's file (`deleted` or `kept`) from a path the
   running cli no longer renders (`orphaned`, per
   [Files no longer rendered](../../../conventions.md#safety)). In this table a
   removed tool is a named tool or a dependency step 3 prunes: the files of a
   pruned dependency follow the removed-tool rows, and the files of a dependency
   that stays (its parents updated) follow the still-selected rows. Each
   recorded or rendered path falls in one case. A shared file is one of the
   files listed in
   [Selection-dependent files](../../../entities/tool/index.md#selection-dependent-files)
   of the [Tool](../../../entities/tool/index.md) entity. A shared file that the
   new selection still renders is a file of a still-selected tool, never of a
   removed tool.

   | Case                                                                                                                                                                                                                | Target?                           | Report key                            | Stays in `files`?             |
   | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------- | ------------------------------------- | ----------------------------- |
   | Recorded file (not `create_only`) of a removed tool, present on disk                                                                                                                                                | yes — deleted                     | `deleted`                             | no                            |
   | Recorded `create_only` file of a removed tool, present on disk                                                                                                                                                      | no — never deleted, stays on disk | `kept`                                | no                            |
   | Recorded path of a removed tool, absent on disk, absence committed                                                                                                                                                  | no                                | none — dropped silently               | no                            |
   | Recorded path of a removed tool, absent on disk, deletion not committed                                                                                                                                             | yes — dirty, step 5               | none                                  | n/a — nothing written, exit 1 |
   | Recorded file (not `create_only`) of a still-selected tool, present on disk, render differs from the working copy (content or mode, including a committed hand edit, overwritten as at a re-run of init)            | yes                               | `changed`                             | yes                           |
   | Recorded `create_only` file of a still-selected tool, present on disk                                                                                                                                               | no — never changed                | `kept`                                | yes                           |
   | Recorded shared file of a still-selected tool, absent on disk, absence committed                                                                                                                                    | yes — created again               | `created`                             | yes                           |
   | Recorded file (not shared) of a still-selected tool, absent on disk, absence committed or not                                                                                                                       | no — left alone                   | none                                  | yes                           |
   | Recorded shared file of a still-selected tool, absent on disk, deletion not committed                                                                                                                               | yes — dirty, step 5               | none                                  | n/a — nothing written, exit 1 |
   | File of a tool or dependency added back by the step 2 [Repair on read](../../../entities/setup-config/index.md#repair) (even when none of the named tools is selected), except a `create_only` file present on disk | yes                               | `created`, `changed` or `unchanged`   | yes                           |
   | `create_only` file of a tool or dependency added back by the step 2 [Repair on read](../../../entities/setup-config/index.md#repair), present on disk                                                               | no — never changed                | `kept`                                | yes                           |
   | Recorded file (not `create_only`) of a still-selected tool, render equals the working copy                                                                                                                          | no                                | `unchanged`                           | yes                           |
   | Path the running cli renders for a tool that stays selected, not recorded in `files` (absent on disk, or present in any git state)                                                                                  | no — left alone                   | none — not reported                   | no                            |
   | Path of a removed tool, not recorded in `files`, present on disk                                                                                                                                                    | no — left alone                   | none — not reported                   | no                            |
   | Recorded path the running cli does not render, whichever tool it belonged to, present on disk                                                                                                                       | no — never deleted, left in place | `orphaned`                            | yes                           |
   | Recorded path the running cli does not render, whichever tool it belonged to, absent on disk                                                                                                                        | no                                | none — dropped silently, not reported | no                            |
   | `.config/bootstrap.yaml`                                                                                                                                                                                            | yes, like any other file          | `changed`                             | no, never listed in `files`   |

   A recorded path that leaves `files` is dropped only when the run rewrites the
   setup file for another change; on its own it is not a change, so a list that
   is otherwise empty stays empty (step 3).
5. Remove checks the targets per [safety](../../../conventions.md#safety). Steps
   1–5 find every refusal, so a refusal never follows a prompt.
6. Consent, per [config](../../../conventions.md#config) step 4, for the list of
   changes of steps 3–5.
7. Ctrl-C during the run exits per [errors](../../../conventions.md#errors).
8. Remove writes every step 4 target: it deletes the `deleted` files, applies
   [Empty directories](../../../conventions.md#safety) after the deletions,
   writes the `created` and `changed` files, then rewrites
   `.config/bootstrap.yaml` last: `values.tools` without the named names and
   with any tool step 2 repaired, `values.dependencies` without the pruned
   dependencies, with the parents of the others recomputed and with any
   dependency step 2 repaired, without any unknown dependency name step 2 found,
   and `files` per
   [invariant 2](../../../entities/setup-config/index.md#invariants). It never
   commits. When the write created or changed `.vscode/extensions.json`, remove
   then runs the editor's command line once to sync the `REPO_NAME` profile to
   it: every listed extension installed, and every extension no remaining
   selected tool lists (the removed tool's, unless another tool lists it)
   uninstalled; absent, or a failed install or uninstall, is a warning and the
   exit code stays 0 ([errors](../../../conventions.md#errors)), the written
   files stay.
9. Remove reports, in human output and under `--json`: `removed`, `not_added`,
   `dependencies_added`, `dependencies_pruned`, `deleted`, `created`, `changed`,
   `unchanged`, `kept`, `orphaned` and `warnings`. `removed` lists only the
   named tools that were selected directly; `dependencies_pruned` lists the
   dependencies step 3 pruned, and `dependencies_added` the dependencies a step
   2 repair added back. `warnings` are listed in the order they were found; the
   path lists are sorted by path; `removed`, `not_added`, `dependencies_added`
   and `dependencies_pruned` are sorted by name. Each path appears under the key
   its case gives in step 4. A hidden tool (`mise`) is never listed in a name
   list of the report; the paths of its files are listed like any path. A
   directory step 8 removed is not reported, in human output, `--json` or the
   `--dry-run` list. The next command follows
   [errors](../../../conventions.md#errors). A run that changed a file ends its
   human output with the closing hint "review `git diff`; restore your own lines
   with `git restore -p <file>`". Exit 0 on success.

Modes: `--dry-run` shows the list of changes, writes nothing, never prompts (in
any terminal, with or without `-y`) and exits with the code the real run (with
`-y`) would return. `--dry-run --json` prints the JSON result the real run would
return plus `"dry_run": true`; on a non-zero exit it prints the same error
document plus `"dry_run": true`.

## Guarantees

| Step / group                       | Consistency                                      | On failure                                                                                                                                                                                                | Idempotency                                                                  | Load & latency          |
| ---------------------------------- | ------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ----------------------- |
| 1–7                                | atomic — nothing is written                      | none — nothing written; exit codes per steps 1–7                                                                                                                                                          | n/a — a re-run starts from the same repo state                               | n/a — one local command |
| 8                                  | atomic — all-or-nothing                          | per [safety](../../../conventions.md#safety) and [errors](../../../conventions.md#errors)                                                                                                                 | a re-run with the same names reports "not added", writes nothing and exits 0 | n/a — one local command |
| editor profile sync (after step 8) | best effort — outside the repository, not atomic | a failed editor command line, install or uninstall is a warning, the written files stay, exit 0; an interrupt stops the sync, the written files stay, exit 130 ([errors](../../../conventions.md#errors)) | n/a — runs only after a write that changed `.vscode/extensions.json`         | n/a — one local command |
| 9                                  | atomic — output only                             | none — changes no repo state                                                                                                                                                                              | n/a                                                                          | n/a — one local command |

## Diagram

```mermaid
sequenceDiagram
    actor O as Actor
    participant M as Remove
    participant C as Setup config
    participant G as Tool
    participant R as Repository
    participant E as Editor command line
    O->>M: bootstrap remove tools
    M->>M: usage check
    alt usage error
        M-->>O: exit 2
    end
    M->>R: preflight
    alt preflight failed
        M-->>O: exit 3
    end
    M->>C: read config
    alt not set up
        M-->>O: exit 1
    else config invalid or newer version or format
        M-->>O: exit 3
    else older version
        M-->>O: exit 1
    end
    M->>G: validate names
    alt name still needed by a selected tool
        M-->>O: exit 2
    end
    M->>M: render in memory
    alt no change
        M-->>O: exit 0
    end
    M->>R: check targets
    alt not a file
        M-->>O: exit 3
    else dirty target
        M-->>O: exit 1
    end
    alt dry run
        M-->>O: show the list, exit 0
    else no prompt possible and no -y
        M-->>O: changes need --yes, exit 2
    else consent
        M-->>O: ask consent per config
        O-->>M: answer
        alt declined
            M-->>O: nothing changed, exit 0
        end
    end
    alt interrupt before the write
        M-->>O: exit 130
    else proceed
        M->>R: delete and rewrite files, rewrite config last
        alt write failure or interrupt
            M->>R: restore from HEAD, delete created files
            alt restore failed or write failure
                M-->>O: exit 3
            else interrupt
                M-->>O: exit 130
            end
        else written
            opt extensions.json changed
                M->>E: sync editor profile
                alt command line absent, install or uninstall failed
                    E-->>M: warning, written files stay
                else Ctrl-C during the sync
                    M-->>O: written files stay, exit 130
                end
            end
            M-->>O: report, exit 0
        end
    end
```

## Background Jobs

N/A — runs synchronously in one command invocation.

## Acceptance

A criterion that says only "when remove runs" runs with `-y`; a criterion that
names its flags (no `-y`, `--dry-run`) runs with those only.

A case that records a missing `git` while another tool of its `max: one`
category is recorded cannot be reached with the 1.0 catalog (`version-control`
holds only `git`); it holds for a later catalog.

- Given a repository with github added, when `bootstrap remove github -y` runs,
  then github's files are deleted, `values.tools` no longer contains github,
  `files` no longer lists them, nothing is committed and the exit code is 0.
- Given `dprint` is selected and `taplo` is not, when
  `bootstrap remove dprint -y` runs, then
  `.config/mise/conf.d/dprint/mise.dev.toml` and the tool's other files,
  `dprint.jsonc` included, are deleted, the next command is
  `MISE_ENV=dev mise run setup:all` and the exit code is 0.
- Given a removed tool whose deletions empty `.config/mise/conf.d/<tool>/`, when
  remove runs with `-y`, then that directory no longer exists and is not
  reported; a directory that still holds another entry stays.
- Given `grype` is selected, when `bootstrap remove grype -y` runs, then
  `.config/grype.yaml` (`create_only`) stays on disk, is reported `kept` and is
  no longer in `files`, the tool's other files are deleted and the exit code is
  0.
- Given `effect`, `node` and `bun` are selected directly, when
  `bootstrap remove node -y` runs, then `node` is removed, `effect`'s
  `javascript` engine is met by `bun` and the exit code is 0.
- Given `effect` with `node` recorded as its dependency and `bun` selected
  directly, when `bootstrap remove node -y` runs, then `effect`'s link moves to
  `bun`, `node` is pruned (listed in `dependencies_pruned`, its files deleted),
  it is reported `not_added` and the exit code is 0.
- Given `effect` and `pnpm` selected and `node` their only `javascript` runtime,
  when `bootstrap remove node -y` runs, then nothing is written and the exit
  code is 2 with one error naming `effect` and `pnpm`, and the next command is
  `bootstrap remove node effect pnpm -y`; when that command runs, then all three
  are removed and the exit code is 0.
- Given `effect`, `pnpm` and `bun` selected and `node` selected directly, when
  `bootstrap remove node -y` runs, then nothing is written and the exit code is
  2 naming `pnpm` only (`bun` keeps `effect` met), with the next command
  `bootstrap remove node pnpm -y`.
- Given a recorded shared file missing from disk whose tool stays selected, when
  remove runs and its absence is committed, then the file is created again,
  reported `created` and the exit code is 0.
- Given a recorded shared file of a still-selected tool that is deleted in the
  working tree with the deletion not committed, when remove runs, then nothing
  is written, the path is listed and the exit code is 1.
- Given a recorded path of a removed tool that is deleted in the working tree
  with the deletion not committed, when remove runs, then nothing is written,
  the path is listed and the exit code is 1.
- Given a committed hand edit to a file of a still-selected tool, when remove
  runs with `-y`, then the file is overwritten with its render, reported
  `changed` and the exit code is 0.
- Given a tool whose `removable` is false (e.g. `git`), when remove runs, then
  nothing is written and the exit code is 2.
- Given an unknown tool name, when remove runs, then the exit code is 2.
- Given an old tool name that the catalog lists in a tool's `renamed_from`, when
  `bootstrap remove <old>` runs, then nothing is written, the error is "unknown
  tool `<old>` — use `<new>`" and the exit code is 2.
- Given `bootstrap remove mise -y`, when it runs, then `mise` is an unknown
  name: nothing is written, the exit code is 2 and `mise` is never listed as a
  tool.
- Given a tool not in `values.tools`, when remove runs with its name, then
  nothing is written, it is reported `not_added` and the exit code is 0.
- Given `bootstrap remove` with no names, when it runs, then nothing is written,
  a short usage is shown and the exit code is 2.
- Given `taplo` is selected, when `bootstrap remove dprint -y` runs, then
  nothing is written and the exit code is 2 naming `taplo`, with the next
  command `bootstrap remove dprint taplo -y`.
- Given `taplo` and `dprint` are selected, when
  `bootstrap remove taplo dprint -y` runs, then the requires check passes on the
  selection after the request, both are removed and the exit code is 0.
- Given `effect` and `virajp-linter` are selected with `node` a dependency
  needed only by them, when `bootstrap remove effect virajp-linter -y` runs,
  then `node` is pruned: it is on the list of changes before consent, its files
  are deleted (`create_only` files kept), it leaves `values.dependencies`,
  `dependencies_pruned` lists `node`, `removed` lists `effect` and
  `virajp-linter` alone and the exit code is 0.
- Given `effect` and `virajp-linter` are selected with `node` a dependency of
  both, when `bootstrap remove effect -y` runs, then `node` stays with its
  files, `values.dependencies` records `node` with parents `[virajp-linter]`,
  and the exit code is 0.
- Given a pruned dependency whose deletions empty its `conf.d/` directory, when
  remove runs with `-y`, then the directory no longer exists and is not reported
  ([safety](../../../conventions.md#safety)).
- Given `pre-commit` is selected (unremovable) with `python` and `uv` as its
  dependencies, when `bootstrap remove python uv -y` runs, then nothing is
  written and the exit code is 2 with one error that names both `python` and
  `uv` with their parent `pre-commit`, in step order, and no next command.
- Given `dprint` is only a dependency of `taplo`, when
  `bootstrap remove dprint -y` runs, then nothing is written and the exit code
  is 2 naming `taplo`.
- Given `pre-commit` is selected (unremovable) with `python` and `uv` as its
  dependencies, when `bootstrap remove python -y` runs, then nothing is written
  and the exit code is 2 naming `pre-commit` with no next command.
- Given a selected tool whose dependency is missing from `values.dependencies`,
  when remove runs, then the dependency is added back with the warning "added
  back dependency `<name>`", it is listed in `dependencies_added`, its files are
  `created` (a `create_only` file present on disk is `kept`), the config is
  `changed` and the exit code is 0.
- Given `bootstrap remove` with no names in a directory that is not a git
  repository or not set up, when it runs, then the exit code is 2, not 3 or 1.
- Given a recorded path the running cli no longer renders that is present on
  disk, when remove runs, then the path is not deleted, stays in `files` and is
  reported `orphaned`.
- Given a recorded path the running cli no longer renders that is absent on
  disk, when remove runs, then the path is dropped from `files`, is not reported
  and the exit code is 0.
- Given a tool name given twice, when `bootstrap remove github github -y` runs,
  then the tool is removed once and the exit code is 0.
- Given a recorded path of the removed tool already absent from the repository,
  with the absence committed, when remove runs, then the path is dropped from
  `files` without a report entry and the exit code is 0.
- Given a file of the tool that is modified or staged, when remove runs, then
  nothing is written, the path is listed and the exit code is 1.
- Given a target that is untracked or ignored in git, when remove runs, then
  nothing is written, the path is listed and the exit code is 1.
- Given a target that is a directory or a symlink, when remove runs, then
  nothing is written and the exit code is 3 naming that path.
- Given one target that is a directory and another that is modified, when remove
  runs, then nothing is written, the exit code is 3 and the message lists both
  paths.
- Given `.config/bootstrap.yaml` has an uncommitted change, when remove runs,
  then nothing is written and the exit code is 1 listing it.
- Given a recorded `values.tools` missing an unremovable tool and no other tool
  of its `max: one` category is recorded, when remove runs, even with none of
  the named tools selected, then the repair is written: the tool is in
  `values.tools`, its files are rendered and reported `created` or `changed`,
  `.config/bootstrap.yaml` is listed `changed`, the warning "added back
  unremovable tool `<name>`" is in `warnings` and the exit code is 0.
- Given a recorded `values.tools` with a tool name the running catalog does not
  have, when remove runs, then nothing is written and the exit code is 3 naming
  the fix.
- Given stdin and stdout are terminals and no `--json`, when remove runs without
  `-y` with a named tool selected (or a repair), then the list of changes and
  "apply these changes? y/N" are shown; on "n", "no" or Enter nothing is
  written, "nothing changed" is shown and the exit code is 0; on "y" or "yes"
  the changes are applied and the exit code is 0.
- Given no terminal (stdin or stdout is not a terminal) or `--json`, when remove
  runs without `-y` with a named tool selected (or a repair), then no prompt is
  shown, nothing is written, the exit code is 2, the message is "changes need
  --yes" and the next command is the same command plus `--yes`.
- Given a dirty target, when remove runs without `-y` in a terminal, then the
  exit code is 1 and no prompt is shown (every refusal comes before the
  consent).
- Given none of the named tools is selected, no repair happened and the request
  leaves no dependency without a parent, when remove runs without `-y`, then
  nothing is written, no prompt is shown, `not_added` is reported and the exit
  code is 0.
- Given a stale dependency (no tool of the recorded selection needs it) and none
  of the named tools selected, when remove runs, then it is bookkeeping alone:
  nothing is written, no prompt is shown, the dependency is not pruned and the
  exit code is 0.
- Given a stale dependency and a named tool that is selected, when
  `bootstrap remove <tool> -y` runs, then the stale dependency is pruned with
  the removal (files deleted, listed in `dependencies_pruned`), the parent lists
  are corrected and the exit code is 0.
- Given a `--dry-run` without `-y`, in any terminal, when remove runs, then the
  plan is shown, no prompt is shown, nothing is written and the exit code is 0.
- Given `--json`, when remove exits on any code, then exactly one JSON result is
  printed, with `removed`, `not_added`, `dependencies_added`,
  `dependencies_pruned`, `deleted`, `created`, `changed`, `unchanged`, `kept`,
  `orphaned` and `warnings` on success; a successful JSON result has a top-level
  `next_command` key only per [errors](../../../conventions.md#errors), and an
  error document still carries `error.next_command`.
- Given `--dry-run --json`, when remove runs, then nothing is written and the
  JSON result the real run would return is printed plus `"dry_run": true`; on a
  non-zero exit it is the same error document plus `"dry_run": true`.
- Given no `.config/bootstrap.yaml`, when remove runs, then the exit code is 1
  and the message names `bootstrap init`.
- Given a recorded version older than the running cli, when remove runs, then
  the exit code is 1 and the next command is `bootstrap init`.
- Given an invalid setup file or a newer recorded version, when remove runs,
  then the exit code is 3 and nothing is written.
- Given a directory outside a git repository, or no `mise` on the `PATH`, when
  remove runs, then nothing is written and the exit code is 3.
- Given `mise` found at a path other than `~/.local/bin/mise`, when remove runs,
  then one warning names the path, it is listed in `warnings` and the run
  continues.
- Given a write failure injected mid-run, when remove runs, then every changed
  or deleted target is restored from `HEAD`, every created file is deleted and
  the exit code is 3.
- Given a run, when the actor presses Ctrl-C during the write, then the
  repository is as it was before and the exit code is 130.
- Given a write failure or an interrupt during the writes whose restore also
  fails, when remove runs, then the exit code is 3 and the output lists every
  path not restored (`--json`: `unrestored`).
- Given a path the cli renders for a tool that stays selected but that is not in
  `files` and is absent on disk, when remove runs, then that path is not created
  and not reported.
- Given a recorded `values.tools` that names `mise`, when remove runs, then
  nothing is written and the exit code is 3 naming the fix "remove `mise` from
  values.tools" ([Validity](../../../entities/setup-config/index.md#validity)).
- Given a recorded `values.tools` that misses `git` while another tool of its
  `max: one` category is recorded, when remove runs, then the other tool is a
  valid replacement: there is no exit 3 and no repair.
- Given a recorded `values.dependencies` with a name the running catalog does
  not have and `github` selected, when `bootstrap remove github -y` runs, then
  the name is dropped from `values.dependencies` at the rewrite, the warning
  "dropped unknown dependency `<name>`" is in `warnings` and the exit code is 0.
- Given `claude` selected (the only tool that requires `jq`) and `jq` missing
  from `values.dependencies`, when `bootstrap remove claude -y` runs, then `jq`
  is not added back, is in neither `dependencies_added` nor
  `dependencies_pruned`, and no "added back dependency" warning is reported.
- Given a recorded `values.tools` or `values.dependencies` that names a tool by
  an old name the running catalog lists in `renamed_from`, when remove runs with
  `-y`, even with none of the named tools selected, then the new name replaces
  it on the rewrite, `warnings` reports "renamed tool `<old>` to `<new>`",
  `.config/bootstrap.yaml` is listed `changed` and the exit code is 0; without
  `-y` and with no terminal, nothing is written and the exit code is 2 "changes
  need --yes".
- Given a recorded tool whose `host_os` is not the running host, when
  `bootstrap remove <tool> -y` runs, then it is removed like any other tool and
  the exit code is 0.
- Given `taplo` selected, when `bootstrap remove taplo -y` runs, then
  `.vscode/extensions.json` changes and the editor's command line runs once to
  sync the `REPO_NAME` profile: `tamasfe.even-better-toml` is uninstalled from
  it (no remaining tool lists it); given the command line is not on the `PATH`,
  then one warning names the fix, the files stay written and the exit code is 0.
- Given a run that changed a file, when remove ends, then the human output ends
  with "review `git diff`; restore your own lines with `git restore -p <file>`".
- Abuse case: n/a — runs locally with the caller's own permissions on the
  caller's own repository; every input is validated per
  [baseline](../../../conventions.md#baseline) boundary-validation.

## References

- [baseline](../../../conventions.md#baseline),
  [safety](../../../conventions.md#safety),
  [errors](../../../conventions.md#errors),
  [config](../../../conventions.md#config),
  [Terminal UX](../../../design-system.md#terminal-ux)
- [Set up a repository](../110-setup-repository/index.md),
  [Add a tool](../140-add-tool/index.md) (inverse)
- API surface: N/A — no service project. Screens surface: N/A — cli has no
  screen platform.
