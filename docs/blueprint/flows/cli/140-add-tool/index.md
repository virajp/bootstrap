---
type: vwf-flow
title: Add a tool
description: One command adds tools to a set-up repository, replacing a held
  tool only with consent, with nothing half-written.
status: reviewed
implementation: none
---

# Flow: Add a tool

## Purpose

`bootstrap add <tool>...` adds one or more tools of the
[Tool](../../../entities/tool/index.md) catalog to a repository that is already
set up, adds the tools they require as dependencies, replaces a tool of a
`max: one` [Tool category](../../../entities/tool-category/index.md) (or a
runtime of an engine) only with consent, and records the new selection in the
setup config.

Serves: [Fast new-repo setup](../../../product.md#goal-fast-setup)

## Trigger & Actors

| Actor                            | May trigger                                               | Authorization                         | Audit-recorded                                                  |
| -------------------------------- | --------------------------------------------------------- | ------------------------------------- | --------------------------------------------------------------- |
| Repo owner (any terminal)        | the command `bootstrap add <tool> [<tool>...]` with flags | write access to the working directory | no — no audit foundation; git history keeps every replaced file |
| AI agent or CI (non-interactive) | `bootstrap add <tool>...`                                 | write access to the working directory | no — no audit foundation; git history keeps every replaced file |

## Steps

Add accepts `-y`/`--yes`, `--replace`, `--dry-run`, `--json`, `--quiet`,
`--verbose`, `--no-color` and `--help`. It takes none of init's value flags
(step 5 renders with the recorded `values`); any other flag is a usage error.

Usage errors come first: no tool name given, an unknown flag or a bad flag
value, an unknown tool name (`mise` is one: it is never shown as a tool), or two
distinct names of one `max: one` category or of one `max: one` engine (repeated
names are merged first) exits 2 with short usage before the preflight and before
`.config/bootstrap.yaml` is read. Nothing is written. The usage errors are
reported together in one error ([errors](../../../conventions.md#errors)).

An old tool name (one the catalog lists in a tool's `renamed_from`) typed on the
command line is an unknown tool name: a usage error, exit 2, whose error names
the new name as the fix, "unknown tool `<old>` — use `<new>`". Only a recorded
old name is repaired (step 2).

1. Add runs, after the usage check, the preflight of
   [Set up a repository](../110-setup-repository/index.md) step 1.
2. Add reads `.config/bootstrap.yaml` under
   [Setup config](../../../entities/setup-config/index.md), applying its
   [validity and repair](../../../entities/setup-config/index.md#validity) and
   [version guard](../../../entities/setup-config/index.md#version-guard). A
   repair (an unremovable tool or a dependency added back, or a renamed tool
   read under its new name: the recorded old name becomes the new one in the
   rewrite, warning "renamed tool `<old>` to `<new>`", per
   [Repair on read](../../../entities/setup-config/index.md#repair)) lets the
   run continue, its warning goes in `warnings`, and it is written back in
   step 7. Every repair is a change: it is on the list of changes, needs consent
   (step 6) and defeats the "already added" shortcut of step 3. A missing
   dependency is repaired only when the selection after the request needs it
   (the replacements of step 4 count): one the request leaves unneeded is never
   added back and never reported (no warning, not in `dependencies_added` or
   `dependencies_pruned`). A `values.dependencies` name the running catalog does
   not have is dropped as such a repair on the step 7 rewrite and reported in
   `warnings` as "dropped unknown dependency `<name>`". Bookkeeping is never a
   change on its own: a recorded path in `files` that the running cli no longer
   renders and that is absent on disk, a stale dependency (one no tool of the
   selection needs, whose loss of parents the request did not cause) and a wrong
   parent list in `values.dependencies`. When the run rewrites the setup file
   for another change, it also drops the stale path, prunes the stale dependency
   (its files deleted and reported as usual, listed in `dependencies_pruned`)
   and corrects the parent lists; when bookkeeping is the only difference,
   nothing is written, no prompt is shown and the exit code is 0.
3. Add reads the [Tool](../../../entities/tool/index.md) `requires` and the
   [Setup config](../../../entities/setup-config/index.md) `values.tools` and
   `values.dependencies`, and derives the dependencies of the selection after
   the request (the held tools minus the replaced ones found in step 4, plus the
   named ones): a missing requirement is not a refusal; each unmet tool
   requirement adds that tool, and each unmet engine requirement the engine's
   default runtime, as a dependency recorded with its parents
   ([Requires and dependencies](../../../entities/tool/index.md#requires-and-dependencies)).
   A name already in `values.tools` is reported "already added" and changes
   nothing; a name recorded in `values.dependencies` is not "already added": it
   becomes direct (moves from `values.dependencies` to `values.tools`, no parent
   changes). If every name is already in `values.tools` and step 2 made no
   repair, add renders in memory only to find the `orphaned` paths and skips to
   step 8; that run has nothing to change (bookkeeping alone is not a change),
   shows no prompt and exits 0.
4. Add finds the replacements: a tool that shares a
   [Tool category](../../../entities/tool-category/index.md) `max: one` with a
   held tool replaces it (a tool has a list of categories; only that category
   concerns the replacement, the other categories of either tool decide
   nothing). A runtime named while another runtime of a `max: one` engine is
   held (for example `add temurin` while `openjdk` is held) is always a
   replacement and needs `--replace` like any replacement; with it, it follows
   the `--replace` case below. A runtime named while another runtime of a
   `max: many` engine is held (for example `add bun` while `node` meets
   `effect`'s `javascript` engine) is a replacement only with `--replace`:
   - Without `--replace` both are kept: the held runtime keeps its links and its
     parents do not move, the named one is added as a direct tool, nothing is
     pruned and there is no `replaced` entry. This is not a replacement, so it
     is never refused for a missing `--replace`.
   - With `--replace` the engine links of the held runtime move to the named one
     and the held runtime leaves the selection: pruned when it was only a
     dependency (listed in `dependencies_pruned`), taken out of `values.tools`
     when it was selected directly. `replaced` lists `{from, to}` and its files
     are removed as for a removed tool (`create_only` files kept). The
     exception: when another selected tool requires the held runtime itself (for
     example `pnpm` or `yarn` requiring `node`), it stays as a dependency of
     that tool, keeps its files, only the engine links move, and `replaced` has
     no entry for it.
   - A `max: many` engine that holds two or more other runtimes is out of scope:
     each `max: many` engine has two runtimes in 1.0.

   The replaced ones are the held tools that leave the selection, and only they
   get a `replaced` `{from, to}` entry. A dependency that no parent needs after
   the request is pruned. A repair from step 2 holds an unremovable tool (`git`
   or `pre-commit`) again, so a named tool of its `max: one` category is a
   replacement of it and needs `--replace` like any replacement
   ([Tool category](../../../entities/tool-category/index.md)). A replacement
   without `--replace` exits 2 with nothing written, before consent: no prompt
   is shown in any terminal, with or without `-y`. The next command is the same
   command plus `--replace`. A replacement that would leave a selected tool
   dangling (the
   [Dangling rule](../../../entities/tool/index.md#requires-and-dependencies))
   exits 2 naming that tool (checked on the selection after the request, before
   any write). `--replace` with nothing to replace is ignored. A named tool
   whose `host_os` is not the running host is refused too (exit 2, a new
   selection is refused; a tool already recorded is "already added" on any
   host); its dependencies follow the `requires` model unchanged. All the
   refusals of step 4 are reported together in one error
   ([errors](../../../conventions.md#errors)).
5. Add renders the full file set of the new selection (the held
   [Tool](../../../entities/tool/index.md)s, minus the replaced ones, plus the
   added ones and the dependencies added, minus the dependencies pruned) with
   the recorded `values`, and so prepares the full list of changes: each file to
   create, change, delete or replace (a dependency's files are created like any
   added tool's, a pruned dependency's are removed). Shared files are targets
   too; the files of a replaced tool or a pruned dependency are removed as in
   [Remove a tool](../150-remove-tool/index.md). An exact mise pin already in
   the repository is kept; a new tool line gets bootstrap's exact pin
   ([tool versions](../../../conventions.md#tool-versions)). Planning, the setup
   file `.config/bootstrap.yaml` as a target and `orphaned` paths follow
   [safety](../../../conventions.md#safety), with the refusals of step 4 first
   ([precedence](../../../conventions.md#safety)).
6. The actor gives consent to the list of changes to the repository and the
   [Setup config](../../../entities/setup-config/index.md), as in
   [config](../../../conventions.md#config) steps 2 to 4 (and
   [errors](../../../conventions.md#errors) for "changes need --yes").
7. Add writes the planned changes, then rewrites `.config/bootstrap.yaml` last:
   `values.tools` gains the added tools (including a dependency made direct) and
   loses the replaced ones, `values.dependencies` is re-derived (dependencies
   added with their parents, parents moved, unneeded ones pruned), and `files`
   is refreshed per
   [Setup config invariants](../../../entities/setup-config/index.md#invariants).
   Add never commits. When the write created or changed
   `.vscode/extensions.json` (a tool with VS Code extensions was added or
   replaced), add then runs the editor's command line once to sync the
   `REPO_NAME` profile to it (every listed extension installed, every extension
   no selected tool lists uninstalled); absent, or a failed install or
   uninstall, is a warning and the exit code stays 0
   ([errors](../../../conventions.md#errors)), the written files stay.
8. Add reports `added`, `already_added`, `replaced` (each `{from, to}`),
   `dependencies_added`, `dependencies_pruned`, `created`, `changed`, `deleted`,
   `unchanged`, `kept`, `orphaned`, `warnings` and `next_command`, in human
   output and under `--json`. `added`, `already_added` and `replaced` hold only
   the tools the actor named; `dependencies_added` lists the tools add added as
   dependencies in this run and `dependencies_pruned` every dependency it pruned
   (step 4); both are tool names sorted by name, like the other name lists. The
   next command is printed per [errors](../../../conventions.md#errors);
   otherwise `next_command` is absent. Every other report key is always present.
   A run that changed a file ends its human output with the closing hint "review
   `git diff`; restore your own lines with `git restore -p <file>`".

Modes: `--json` follows [errors](../../../conventions.md#errors), with the keys
of step 8 on success. `--dry-run` exits with the code the real run (with
`--replace` only if flagged, and as if `-y` were given) would return.
`--dry-run --json` adds `"dry_run": true` to the document the real run would
return, an error document included. Both otherwise behave as in
[Set up a repository](../110-setup-repository/index.md). Output rules follow
[Terminal UX](../../../design-system.md#terminal-ux).

## Guarantees

| Step / group                       | Consistency                                      | On failure                                                                                                                                                                                                                                                                                         | Idempotency                                                                      | Load & latency          |
| ---------------------------------- | ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | ----------------------- |
| 1–6                                | atomic — nothing is written                      | none — nothing written yet; exits 0, 1, 2, 3 or 130 per step                                                                                                                                                                                                                                       | n/a — a re-run starts from the same repo state                                   | n/a — one local command |
| 7                                  | atomic — all-or-nothing                          | a write failure or an interrupt restores every changed or deleted target from `HEAD` and deletes every created file per [safety](../../../conventions.md#safety); a write failure exits 3 naming the failing path, an interrupt exits 130; a failed restore exits 3 listing the paths not restored | a re-run with the same names reports "already added", writes nothing and exits 0 | n/a — one local command |
| editor profile sync (after step 7) | best effort — outside the repository, not atomic | a failed editor command line, install or uninstall is a warning, the written files stay, exit 0; an interrupt stops the sync, the written files stay, exit 130 ([errors](../../../conventions.md#errors))                                                                                          | n/a — runs only after a write that changed `.vscode/extensions.json`             | n/a — one local command |
| 8                                  | atomic — output only                             | none — changes no repo state                                                                                                                                                                                                                                                                       | n/a                                                                              | n/a — one local command |

## Diagram

```mermaid
sequenceDiagram
    actor O as Actor
    participant A as Add
    participant C as Setup config
    participant T as Tool
    participant K as Tool category
    participant R as Repository
    O->>A: bootstrap add tools
    A->>A: usage check
    opt usage error
        A-->>O: exit 2
    end
    A->>R: preflight
    opt preflight failed
        A-->>O: exit 3
    end
    A->>C: read config
    alt not set up, older version
        A-->>O: exit 1
    else config invalid, newer version
        A-->>O: exit 3
    else valid, same version
        A->>A: continue
    end
    A->>T: requires, derive dependencies
    A->>K: find replacements
    A->>R: plan targets (list of changes)
    alt replacement or other-host tool refused
        A-->>O: exit 2
    else all already added no repair, or empty list
        A-->>O: exit 0
    else not a file
        A-->>O: exit 3
    else dirty target
        A-->>O: exit 1
    else dry-run
        A-->>O: show list, exit 0
    else no -y, no prompt possible
        A-->>O: exit 2, changes need --yes
    else no -y, terminal, answer is not y
        A->>O: show list, apply these changes? y/N
        A-->>O: nothing changed, exit 0
    else proceed (-y, or terminal and answer y)
        A->>R: write files
        A->>C: rewrite config
        alt write failure or interrupt
            A->>R: restore targets from HEAD, delete created files
            alt restore failed or write failure
                A-->>O: exit 3
            else interrupt
                A-->>O: exit 130
            end
        else written
            opt extensions.json changed
                A->>R: update editor profile, warning if absent
            end
            A-->>O: report, exit 0
        end
    end
```

## Background Jobs

N/A — runs synchronously in one command invocation.

## Acceptance

- Given a repository at the running version, when `bootstrap add <tool> -y` runs
  for a tool in no occupied `max: one` category, then its files exist,
  `values.tools` contains it, `files` includes them and the exit code is 0; then
  a re-run of `bootstrap init -y` reports every file `unchanged`.
- Given a repository at the running version and stdin and stdout that are
  terminals, when `bootstrap add <tool>` for a new tool runs without `-y` and
  without `--json`, then the list of changes is shown and "apply these changes?
  y/N" is asked; "y" or "yes" applies it and the exit code is 0; Enter, "n" or
  "no" writes nothing, prints "nothing changed" and exits 0.
- Given a repository at the running version and no terminal on stdin or stdout,
  or `--json`, when `bootstrap add <tool>` for a new tool runs without `-y`,
  then nothing is written, no prompt is shown, the exit code is 2, the message
  is "changes need --yes" and the next command is the same command plus `--yes`.
- Given a held tool and a terminal, when add runs with `--replace` but without
  `-y`, then the list is shown and the prompt is asked; "y" replaces the tool
  and any other answer writes nothing and exits 0.
- Given a held tool and no terminal, when add runs with `--replace` but without
  `-y`, then nothing is written and the exit code is 2 with the message "changes
  need --yes".
- Given two tool names, when add runs with `-y`, then both tools are added.
- Given a tool name given twice, when add runs with `-y`, then the tool is added
  once and the exit code is 0.
- Given a tool already in `values.tools` and no repair to make, when add runs
  with its name and without `-y`, then nothing is written, no prompt is shown,
  the full step 8 report is printed with the tool in `already_added`, empty path
  lists except `orphaned`, `warnings` and no next command (no `next_command` key
  under `--json`), and the exit code is 0.
- Given `bootstrap add` with no tool name, when add runs, then nothing is
  written, short usage is printed and the exit code is 2.
- Given an unknown tool name, or `mise`, when add runs, then the exit code is 2
  and nothing is written.
- Given an old tool name that the catalog lists in a tool's `renamed_from`, when
  `bootstrap add <old>` runs, then nothing is written, the error is "unknown
  tool `<old>` — use `<new>`" and the exit code is 2.
- Given `bootstrap add effect -y` with no `javascript` runtime selected, when
  add runs, then `effect` is added to `values.tools`, `node` (the engine's
  default runtime) is added to `values.dependencies` with parents `[effect]`,
  its files are created and listed in `created`, `added` lists `effect` alone,
  `dependencies_added` lists `node`, and the exit code is 0 (a missing
  requirement is never a refusal).
- Given `bun` selected directly, when `bootstrap add effect -y` runs, then `bun`
  meets the `javascript` engine, no runtime is added, `dependencies_added` is
  empty and the exit code is 0.
- Given `bootstrap add pnpm -y` with `node` absent, when add runs, then `node`
  is added to `values.dependencies` with parents `[pnpm]` and the exit code is
  0.
- Given `bootstrap add taplo dprint` with neither held, when add runs with `-y`,
  then both are direct, `dprint` is not recorded as a dependency of `taplo`,
  `dependencies_added` is empty and `added` lists `dprint` and `taplo`.
- Given `bootstrap add taplo -y` with `dprint` absent, when add runs, then
  `dprint` is added to `values.dependencies` with parents `[taplo]` and listed
  in `dependencies_added`.
- Given a tool recorded in `values.dependencies`, when add runs with its name
  and `-y`, then it moves from `values.dependencies` to `values.tools` (it
  becomes direct), its parents change nothing, it is not reported "already
  added" and the exit code is 0.
- Given `effect` selected and `node` recorded as its dependency, when
  `bootstrap add bun -y` runs without `--replace`, then `bun` is added to
  `values.tools`, `node` stays in `values.dependencies` with parents `[effect]`
  and keeps its files, nothing is pruned, `replaced` is empty and the exit code
  is 0.
- Given `effect` selected and `node` meeting its `javascript` engine (as a
  dependency, or selected directly), when `bootstrap add bun --replace -y` runs,
  then `effect`'s link moves from `node` to `bun`, `node` leaves the selection
  (pruned and listed in `dependencies_pruned` when it was a dependency, taken
  out of `values.tools` when it was direct), its files are removed except
  `create_only` ones, `replaced` lists `{from: node, to: bun}` and the exit code
  is 0. Given `pnpm` also selected (it requires `node` itself), then `node`
  stays in `values.dependencies` with parents `[pnpm]` and its files, only
  `effect`'s link moves to `bun`, `replaced` has no entry for `node` and the
  exit code is 0.
- Given `kotlin` selected and `openjdk` held for its `java` engine, when
  `bootstrap add temurin -y` runs without `--replace`, then nothing is written,
  the exit code is 2 and the next command is the same command plus `--replace`;
  with `--replace`, `kotlin`'s link moves to `temurin`, `openjdk` leaves the
  selection, `replaced` lists `{from: openjdk, to: temurin}` and the exit code
  is 0.
- Given `bootstrap add openjdk temurin -y`, when add runs, then the exit code is
  2 before the preflight (two runtimes of the `max: one` `java` engine) and
  nothing is written.
- Given a replacement in a `max: one` category whose replaced tool another
  selected tool requires, with no other way to meet that requirement, when add
  runs, then the exit code is 2, the message names the dependent tool and
  nothing is written.
- Given an add whose named tools would replace held tools (in two `max: one`
  categories, or in one), when add runs without `--replace` in any terminal,
  interactive or not, with or without `-y`, then the exit code is 2, nothing is
  written, the refusal comes before consent so no prompt is shown, the one error
  names every refusal, and the next command is the same command plus
  `--replace`.
- Given an unknown tool name and two names of one `max: one` category, when add
  runs, then the exit code is 2 before the preflight, nothing is written and the
  one error names both usage errors; with only the two names of one category,
  the exit code is 2 and nothing is written.
- Given `bootstrap add` with no name in a directory that is not a git repository
  or not set up, when add runs, then the exit code is 2, not 3 or 1.
- Given a held tool, when add runs with `--replace` and `-y` in any terminal,
  then the new tool's files are written, the old tool's files are removed except
  `create_only` ones, `values.tools` swaps the two and `replaced` lists
  `{from, to}`, and the exit code is 0.
- Given an interactive terminal and a dirty target, when add runs without `-y`,
  then the exit code is 1, the path is listed, nothing is written and no prompt
  is shown.
- Given `--replace` and no replacement to make, when add runs with `-y`, then
  the flag is ignored and the run proceeds as without it.
- Given a recorded `values.tools` that misses `git` (or `pre-commit`) and no
  other tool of its `max: one` category is recorded, when add names a tool of
  that category with `--replace` and `-y`, then the repair adds it back and the
  named tool replaces it (`replaced` lists `{from, to}`); without `--replace`
  the exit code is 2 and nothing is written.
- Given a shared file the render would change (for example `dprint.jsonc` when
  `taplo` is added) that is modified, staged, untracked or ignored, when add
  runs, then nothing is written, the exit code is 1 and the path is listed.
- Given a tool with mise entries is added, when add runs with `-y`, then the
  full-set render creates that tool's own `.config/mise/conf.d/<tool>/` file(s),
  and they are listed in `created`.
- Given a setup file that is modified, staged, untracked or ignored, when add
  runs, then nothing is written, the exit code is 1 and `.config/bootstrap.yaml`
  is listed; on a clean run with `-y` it is listed `changed`.
- Given a recorded `values.tools` that misses an unremovable tool and no other
  tool of its `max: one` category is recorded, when add runs with `-y`, even
  with every named tool already added, then the tool is added again, the repair
  is written back, `.config/bootstrap.yaml` is listed `changed`, `warnings`
  reports "added back unremovable tool `<name>`" and the exit code is 0.
- Given the same repair and every named tool already added, when add runs
  without `-y` and with no terminal, then nothing is written, the exit code is
  2, the message is "changes need --yes" and the next command is the same
  command plus `--yes`; in a terminal the list is shown and the prompt is asked
  (the repair is a change and needs consent).
- Given a recorded selection that misses a dependency a selected tool needs,
  when add runs with `-y`, even with every named tool already added, then the
  dependency is added back with its files, the repair is written back,
  `.config/bootstrap.yaml` is listed `changed`, `warnings` reports "added back
  dependency `<name>`", `dependencies_added` lists it and the exit code is 0;
  without `-y` and with no terminal, nothing is written and the exit code is 2
  "changes need --yes" (the repair is a change and needs consent).
- Given a recorded selection that misses a dependency which only the selection
  before the request needed (the request replaces or leaves unneeded the tool
  that required it), when add runs with `-y`, then the dependency is not added
  back, there is no warning and it is in neither `dependencies_added` nor
  `dependencies_pruned`.
- Given a recorded stale dependency, a wrong parent list in
  `values.dependencies` or a recorded path in `files` that the running cli no
  longer renders and that is absent on disk, when add runs with `-y` for another
  change, then the setup file rewrite prunes the dependency (its files deleted
  and listed in `dependencies_pruned`), corrects the parent list and drops the
  path; when that is the only difference, nothing is written, no prompt is shown
  and the exit code is 0.
- Given a recorded `values.tools` with two tools of one `max: one` category,
  when add runs, then nothing is written and the exit code is 3 naming the fix.
- Given a recorded `values.tools` with a tool name the running catalog does not
  have, when add runs, then nothing is written and the exit code is 3 naming the
  fix.
- Given a path in `files` that the running cli does not render for a selected
  tool and that is present on disk, when add runs, then it is left in place,
  stays in `files` and is listed `orphaned`, and it is never a change.
- Given a recorded version older than the running cli, when add runs, then the
  exit code is 1, the next command is `bootstrap init` and nothing is written.
- Given a recorded version newer than the running cli, or a setup file failing
  Setup config validity, when add runs, then the exit code is 3 and nothing is
  written.
- Given no `.config/bootstrap.yaml`, when add runs, then the exit code is 1 and
  the message names `bootstrap init`.
- Given a directory outside a git repository, or no `mise` on the `PATH`, when
  add runs, then nothing is written and the exit code is 3.
- Given `mise` found elsewhere than `~/.local/bin/mise`, when add runs, then a
  warning names the path and the run continues.
- Given a target that is a directory or a symlink, when add runs, then nothing
  is written and the exit code is 3 naming that path.
- Given a run that created or changed `.vscode/extensions.json` and an editor
  command line whose extension install or uninstall fails, when add runs with
  `-y`, then there is one warning, the written files stay and the exit code
  is 0.
- Given an init value flag (for example `--member <path>`) or an unknown flag,
  when add runs, then the exit code is 2 before the preflight, short usage is
  printed and nothing is written.
- Given an existing `create_only` file of a new tool, when add runs with `-y`,
  then it is not changed and is listed in `kept`.
- Given a path whose render equals the working copy, when add runs with `-y`,
  then it is listed `unchanged` and not rewritten.
- Given a write failure injected mid-run, when add runs with `-y`, then the
  tracked tree matches `HEAD` for every target, created files are gone and the
  exit code is 3.
- Given a write failure whose restore also fails, when add runs with `-y`, then
  the exit code is 3, the output lists every path not restored and, under
  `--json`, `error.unrestored` lists the same paths.
- Given an interrupt (Ctrl-C) during the write, when add runs with `-y`, then
  the targets are restored, created files are gone and the exit code is 130.
- Given `--dry-run`, when add runs, then the list is shown, nothing is written,
  no prompt is shown, no `-y` is needed and the exit code is the one the real
  run would return, including 2 for a replacement without `--replace`.
- Given an interrupt (Ctrl-C) at the prompt, when add runs in a terminal, then
  nothing is written, the message is "interrupted — nothing written" and the
  exit code is 130.
- Given `--json` and `-y`, when add runs on any exit, then stdout is exactly one
  document with `exit`; on success it has `added`, `already_added`, `replaced`,
  `dependencies_added`, `dependencies_pruned`, `created`, `changed`, `deleted`,
  `unchanged`, `kept`, `orphaned`, `warnings` and `next_command` equal to
  `MISE_ENV=dev mise run setup:all` when the run added a tool, direct or
  dependency, or changed a `mise` config file, else absent
  ([errors](../../../conventions.md#errors)), with sorted paths.
- Given `--dry-run --json`, when add runs, then the document is the one the real
  run would return plus `"dry_run": true`, and on a non-zero exit it is the same
  error document plus `"dry_run": true`; no file changes.
- Given a recorded `values.tools` that names `mise`, when add runs, then nothing
  is written and the exit code is 3 naming the fix "remove `mise` from
  values.tools" ([Validity](../../../entities/setup-config/index.md#validity)).
- Given a recorded `values.dependencies` that names a tool the running catalog
  does not have, when add runs with `-y`, then the name is dropped on the
  `.config/bootstrap.yaml` rewrite, `warnings` reports "dropped unknown
  dependency `<name>`" and the exit code is 0.
- Given a recorded `values.tools` or `values.dependencies` that names a tool by
  an old name the running catalog lists in `renamed_from`, when add runs with
  `-y`, then the new name replaces it on the rewrite, `warnings` reports
  "renamed tool `<old>` to `<new>`", the setup file is listed `changed` and the
  exit code is 0; without `-y` and with no terminal, nothing is written and the
  exit code is 2 "changes need --yes" (the repair is a change and needs
  consent).
- Given a tool whose `host_os` is not the running host, when add names it, then
  nothing is written and the exit code is 2; given the tool is already recorded,
  then it is reported "already added" and the exit code is 0.
- Given an exact mise pin already in the repository for a held tool, when add
  runs with `-y`, then the pin is kept; a new tool's mise line gets bootstrap's
  exact pin ([tool versions](../../../conventions.md#tool-versions)).
- Given an added or replaced tool with VS Code extensions, when add runs with
  `-y`, then `.vscode/extensions.json` is created or changed and the editor's
  command line runs once to sync the `REPO_NAME` profile (a replaced tool's
  extensions that no selected tool lists are uninstalled); given it is not on
  the `PATH`, then one warning names the fix, the files stay written and the
  exit code is 0.
- Given a run that changed a file, when add ends, then the human output ends
  with "review `git diff`; restore your own lines with `git restore -p <file>`".
- Given a named tool and a held tool that share a `max: one` category among
  several, when add runs with `--replace` and `-y`, then only that category
  decides the replacement and the tools' other categories change nothing.
- Abuse case: n/a — runs locally with the caller's own permissions on the
  caller's own repository; every input is validated per
  [baseline](../../../conventions.md#baseline) boundary-validation.

## References

- [baseline](../../../conventions.md#baseline),
  [safety](../../../conventions.md#safety),
  [errors](../../../conventions.md#errors),
  [config](../../../conventions.md#config)
- [design-system](../../../design-system.md#terminal-ux) — Terminal UX
- [Remove a tool](../150-remove-tool/index.md) — the inverse flow
- [Select tools](../160-select-tools/index.md) — the interactive alternative
- API surface: N/A — no service project; the flow is a local command
- Screens surface: N/A — cli has no screen platform
