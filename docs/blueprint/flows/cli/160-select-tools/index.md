---
type: vwf-flow
title: Select tools
description: One interactive view shows every tool by category, so the owner
  sets up a repository, adds, removes and replaces tools and edits the recorded
  values in one place, with nothing half-written.
status: reviewed
implementation: none
---

# Flow: Select tools

## Purpose

`bootstrap tui` opens one view of every tool category with its tools and the
recorded values of a repository. Serves:
[Fast new-repo setup](../../../product.md#goal-fast-setup)

## Trigger & Actors

| Actor                                  | May trigger                 | Authorization                         | Audit-recorded                          |
| -------------------------------------- | --------------------------- | ------------------------------------- | --------------------------------------- |
| Repo owner (interactive terminal only) | the command `bootstrap tui` | write access to the working directory | no, the product has no audit foundation |

## Steps

### Exits

Every refusal below writes nothing. Checked in this order; the first that fails
decides the exit code.

| Condition                                                                                                                                                                                                   | Exit | Message / next command                                                                                                                                                                |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--help`, also with other flags or arguments (help wins)                                                                                                                                                    | 0    | the tui help, which states that tui takes no flags or arguments besides `--help`                                                                                                      |
| A usage error: any argument, or any flag other than `--help` (for example `-h`, `--json`, `--quiet`, `--version`, `--no-color`, `-y`, `--tool`); all usage errors are found together                        | 2    | the short usage naming every usage error found; no JSON result is printed                                                                                                             |
| No git repository, or `mise` not on the `PATH` (preflight, as [Set up a repository](../110-setup-repository/index.md) step 1)                                                                               | 3    | the preflight message of [errors](../../../conventions.md#errors)                                                                                                                     |
| Not interactive (stdin or stdout not a terminal)                                                                                                                                                            | 2    | "tui needs an interactive terminal — use `bootstrap init`/`add`/`remove`"                                                                                                             |
| Setup file fails [Setup config validity](../../../entities/setup-config/index.md#validity)                                                                                                                  | 3    | newer: "upgrade bootstrap"; schema failure: the failing field and "fix the file, then run again"; selection failure: names the fix (for example "remove `oldtool` from values.tools") |
| Recorded `version` older than the running cli                                                                                                                                                               | 1    | next command `bootstrap init`                                                                                                                                                         |
| Setup file changed since the view opened (checked at apply): the content differs from the bytes read when the view opened, or the file appeared or vanished; a touch with identical content is not a change | 1    | the view closes; "setup file changed — run `bootstrap tui` again" (result per step 9)                                                                                                 |
| A target is a directory or symlink (checked before dirty targets, also when both are found)                                                                                                                 | 3    | names every such path, per [safety](../../../conventions.md#safety) precedence                                                                                                        |
| A target is not clean per [safety](../../../conventions.md#safety)                                                                                                                                          | 1    | the view closes; lists the paths                                                                                                                                                      |
| The actor quits without applying, or no target (selection and values unchanged, the render matches the working copy, and no repair; bookkeeping alone is not a change)                                      | 0    | "no change", then the `orphaned` paths of the recorded render ([safety](../../../conventions.md#safety))                                                                              |

1. Tui takes no argument and no flag except `--help` ([exits](#exits)). Color
   follows [Terminal UX](../../../design-system.md#terminal-ux).
2. The warning for `mise` not at `~/.local/bin/mise` is printed once on stderr
   before the view opens and is repeated in the result `warnings`.
3. Tui reads `.config/bootstrap.yaml`. With the file absent, tui opens in
   first-run mode, not an exit: the view is pre-filled with the first-run
   defaults, per [Terminal UX](../../../design-system.md#terminal-ux). A
   recorded `format` older than the running cli's is read per
   [write format](../../../entities/setup-config/index.md#write-format). A
   `values.tools` that misses an unremovable tool, or a dependency that a
   selected tool needs and `values.dependencies` lacks, is handled per
   [repair on read](../../../entities/setup-config/index.md#repair); the warning
   goes in `warnings`. The warning for a missing dependency is reported only
   when the selection after the request needs it (the condition of step 7).
   [Setup config](../../../entities/setup-config/index.md)
4. Tui opens the view: every
   [Tool category](../../../entities/tool-category/index.md) as a group with its
   tools from the [Tool](../../../entities/tool/index.md) catalog, the tools in
   `values.tools` checked (on a first run, the default tools of step 3). Hidden
   tools and categories whose every tool is hidden are not shown, and a hidden
   tool is never listed in the view, the plan or the result tool lists;
   unremovable tools are selected with no `none` choice and are replaced in
   place (step 5), per [Tool](../../../entities/tool/index.md). Each tool row
   shows the tool's `purpose` after its name, in the muted role, for example
   `[x] dprint  Formats code and documents`; a category row keeps its purpose. A
   dependency (a tool a selected tool needs that the actor did not select) is
   shown unchecked, with no row-level mark of its own; it appears, marked
   `dependency`, in the tree of each selected tool that needs it. In a
   `max: one` category, a dependency held there (not reachable with the 1.0
   catalog) leaves the radio group of that category with no marked row, and
   `none` is not marked either. Groups, keys, marks and the terminal-size rule
   follow [Terminal UX](../../../design-system.md#terminal-ux). Ctrl-C at any
   moment in the view writes nothing, exits 130 and prints no result, per
   [errors](../../../conventions.md#errors).

   Dependency tree. Under each selected (checked) tool the view shows the full
   tree of its required tools, nested by `requires`, one child per slot (the
   selected alternative of the slot, else its first alternative; "selected"
   includes a dependency, checked or not, so a slot met by a non-first
   alternative shows that alternative). The mark `dependency` never carries
   meaning by color alone. A direct tool appears unmarked and does not expand.
   Repeats of dependencies are shown (the tree is not deduplicated). Unselected
   rows show no tree. The exact form (pinned):

   ```text
   formatter  Formats the repository's files
     [x] dprint  Formats code and documents
         ├─ node  dependency
         └─ pnpm  dependency
            └─ node  dependency
     [x] taplo  Formats TOML files, run by dprint
         └─ dprint
   ```

   Tree lines are not selectable: focus moves over tool rows only, under the
   existing focus rules of [Terminal UX](../../../design-system.md#terminal-ux).
5. Actor changes the selection per this table. The consent prompt is the one of
   [Terminal UX](../../../design-system.md#terminal-ux); declining keeps the old
   choice. A "none" choice is listed in the apply plan and its single confirm
   covers it.

   | Case                                     | Replace another tool?                                             | Choose "none"?           | Consent prompt?                       |
   | ---------------------------------------- | ----------------------------------------------------------------- | ------------------------ | ------------------------------------- |
   | `max: one`, held tool `removable: true`  | yes                                                               | yes, a removal           | yes, on a replacement; none on "none" |
   | `max: one`, held tool `removable: false` | yes, in place: `git` and `pre-commit` are replaced, never removed | not offered              | yes, on a replacement                 |
   | `max: many`                              | n/a, a tool is unchecked                                          | n/a, a tool is unchecked | none                                  |

   [Tool category](../../../entities/tool-category/index.md)

   Dependencies. The requires check uses the selection as it stands after the
   change ([Tool](../../../entities/tool/index.md#requires-and-dependencies)):
   - Selecting a tool whose required tools are not selected is not refused: each
     required tool that is not selected becomes a dependency of it (the first
     alternative of an unmet slot), shown marked `dependency` in the tool's
     tree.
   - Selecting a tool that is already a dependency makes it direct: it is
     checked, no parent changes, it shows its own tree, and it appears unmarked
     in its parents' trees.
   - Deselecting a tool that a selected tool, direct or dependency, still needs
     is refused in place, unless another alternative of that slot is selected,
     with the message "cannot deselect `<tool>`: needed by `<parents>`"
     (`<parents>` in Catalog order, joined by `,`); the view stays open and the
     selection is unchanged.
   - Deselecting the last parent of a dependency leaves the dependency without a
     parent; it is pruned at apply and listed in the plan (step 7), not in the
     view.
   - Replacing within the alternatives of a slot needs the in-place consent
     prompt like any replacement.
   - A change to a dependency held in a `max: one` category (its replacement, or
     `none` in its radio group) is out of scope: the 1.0 catalog cannot reach
     it.
   - A replacement in a `max: one` category whose held tool a selected tool
     still needs, while the new tool is not an alternative of that slot and no
     other alternative of it is selected, is refused in place like a deselect,
     with the message "cannot replace `<tool>`: needed by `<parents>`"; no
     consent prompt is shown and the selection is unchanged.
6. Actor shows and edits the recorded values (repo, commit scopes, merge model
   for develop and main) in the values section, with the validation rules of
   [Set up a repository](../110-setup-repository/index.md) step 3; an invalid
   value is refused in place with the rule. When `origin` is not readable on a
   first run the repo value is empty, and apply is refused in place until a
   valid repo value is entered.
   [Setup config](../../../entities/setup-config/index.md)
7. Actor applies. Tui first re-reads `.config/bootstrap.yaml`, then computes the
   full render for the new selection and values in memory and finds the targets
   per [safety](../../../conventions.md#safety); the setup file is a target
   ([Setup config](../../../entities/setup-config/index.md)). If none of the
   [exits](#exits) applies, it shows the plan (create, change, delete,
   unchanged, kept, orphaned, replacements, value changes, and the dependencies
   added or pruned) and asks once to confirm. Bookkeeping alone is never a
   change: a recorded path the running cli no longer renders and that is absent
   on disk, a stale recorded dependency (a name the running catalog has) that no
   tool of the selection needs (its parents not lost by this request), and a
   wrong parent list in `values.dependencies` make no target and no plan on
   their own, so with nothing else different tui writes nothing, asks no consent
   and exits 0 "no change". When apply writes for another change (a tool, a
   value, a repair or a file change), it also drops the stale path, prunes the
   stale dependency (its files deleted and reported as usual, listed in
   `dependencies_pruned`) and corrects the parent lists. A dependency left
   without a parent by the request itself is a real change, pruned and planned
   as before. Repairs apply only to the dependencies the selection after the
   request needs: a missing dependency the request leaves unneeded is never
   added back and never reported. A `values.dependencies` name the running
   catalog does not have is not bookkeeping: dropping it is a repair, a change
   in the plan that needs consent even alone, with the warning "dropped unknown
   dependency `<name>`". Declining returns to the view (step 4).
   [Tool](../../../entities/tool/index.md)
8. On confirm, tui writes per [safety](../../../conventions.md#safety). It
   rewrites `.config/bootstrap.yaml` last (values, `values.tools`,
   `values.dependencies` with each dependency's parents, and `files`) and never
   commits. The dependencies are derived from the selection, each with its
   sorted parents, and a dependency no parent needs is pruned (its files deleted
   as for a removed tool). On a first run it renders and writes exactly as a
   first run of [Set up a repository](../110-setup-repository/index.md) does
   (the same targets, safety rules and file modes), and `.config/bootstrap.yaml`
   is created last and listed `created`. A `values.dependencies` name the
   running catalog does not have is dropped on this rewrite as a repair (a
   change in the plan, so apply is not a no-change run) and reported in
   `warnings` as "dropped unknown dependency `<name>`". The drop alone creates
   or changes no `mise` config file, so it does not trigger the next command
   `MISE_ENV=dev mise run setup:all` (per
   [errors](../../../conventions.md#errors)).
   [Setup config](../../../entities/setup-config/index.md)
9. On every exit after the view closes except an interrupt (quit, no change, a
   refused apply, a failure, success), tui prints its human result, warnings
   included, in the result layout of
   [Terminal UX](../../../design-system.md#terminal-ux). A successful apply
   prints the result groups in this order: `added`, `removed`,
   `dependencies_added`, `dependencies_pruned`, `replaced` (a list of
   `{from, to}`), `values_changed` (`{name, from, to}` sorted by name),
   `created`, `changed`, `deleted`, `unchanged`, `kept`, `orphaned` and
   `warnings`, with the meanings of [Add a tool](../140-add-tool/index.md) and
   [Remove a tool](../150-remove-tool/index.md); `added` and `removed` list only
   the tools the actor selected or deselected directly, while
   `dependencies_added` and `dependencies_pruned` list the dependencies that
   entered or left `values.dependencies`. The next command
   `MISE_ENV=dev mise run setup:all` is printed per
   [errors](../../../conventions.md#errors) (a tool added, direct or dependency,
   a replacement, a repair that adds a tool or a first run included, or a
   created, changed or deleted `mise` config file), so an apply that only adds
   dependencies prints it; the drop of an unknown dependency alone does not;
   otherwise no next command is printed and the human result has none. On a
   first run `added` lists the direct tools only, `dependencies_added` lists the
   dependencies and `values_changed` lists every value (repo, commit_scopes,
   merge_model.develop, merge_model.main) as `{name, from: null, to}` sorted by
   name. A successful apply exits 0. The result is human text only; tui never
   prints a JSON result.

## Guarantees

| Step / group                                                 | Consistency                                                         | On failure                                                                                                                                                                                                                                                                                                 | Idempotency                                                             | Load & latency         |
| ------------------------------------------------------------ | ------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- | ---------------------- |
| 1–7 (the view, the re-read, the render and the target check) | atomic, nothing is written                                          | none, nothing written yet; exit 0, 1, 2, 3 or 130 per step                                                                                                                                                                                                                                                 | n/a, a re-run starts from the same repo state                           | n/a, one local command |
| 8 (apply)                                                    | atomic, all-or-nothing per [safety](../../../conventions.md#safety) | a write failure or an interrupt restores every target from `HEAD` and deletes every created file per [baseline](../../../conventions.md#baseline) atomic-multi-write and graceful-shutdown; exit 3 naming the failing path, or 130; a failed restore exits 3 listing the paths not restored (`unrestored`) | a re-run with the same selection and values finds no change and exits 0 | n/a, one local command |
| 9                                                            | atomic, output only                                                 | none, changes no repo state                                                                                                                                                                                                                                                                                | n/a                                                                     | n/a, one local command |

## Diagram

```mermaid
sequenceDiagram
    actor O as Actor
    participant U as Select
    participant C as Setup config
    participant G as Tool and category
    participant R as Repository
    O->>U: bootstrap tui
    alt --help
        U-->>O: help, exit 0
    else usage error
        U-->>O: exit 2
    else preflight failed
        U-->>O: exit 3
    else not interactive
        U-->>O: exit 2
    else
        U->>C: read config (absent: first-run defaults)
        alt config invalid or newer version
            U-->>O: exit 3
        else older version
            U-->>O: exit 1
        else
            U->>G: read categories and tools
            loop until apply or quit
                O->>U: select, edit values
                opt replace in max-one
                    U-->>O: replace?
                    O-->>U: consent or decline
                end
            end
            alt Ctrl-C in the view
                U-->>O: interrupted, nothing written, exit 130
            else quit
                U-->>O: no change, exit 0
            else apply
                U->>C: re-read config
                U->>G: render in memory, find targets
                alt config invalid or newer version
                    U-->>O: exit 3
                else older version
                    U-->>O: exit 1
                else config changed
                    U-->>O: exit 1
                else target is a directory or symlink
                    U-->>O: exit 3
                else dirty target
                    U-->>O: exit 1
                else no target
                    U-->>O: no change, exit 0
                else
                    U-->>O: plan
                    alt decline
                        O-->>U: decline
                        U-->>O: back to the view (step 4)
                    else confirm
                        O-->>U: confirm
                        U->>R: write files
                        U->>C: rewrite config last
                        alt write failure or Ctrl-C
                            U->>R: restore from HEAD
                            alt restore failed
                                U-->>O: unrestored, exit 3
                            else write failure
                                U-->>O: failing path, exit 3
                            else Ctrl-C
                                U-->>O: exit 130
                            end
                        else
                            U-->>O: result, exit 0
                        end
                    end
                end
            end
        end
    end
```

## Background Jobs

N/A, runs synchronously in one command invocation.

## Acceptance

- Given a git repository with `mise` on the `PATH` and stdin or stdout not a
  terminal, when `bootstrap tui` runs, then nothing is written and the exit code
  is 2.
- Given any argument (for example `bootstrap tui foo`) or any flag other than
  `--help` (for example `-h`, `--json`, `--quiet`, `--version`, `--no-color`,
  `--verbose`, `-y`, `--dry-run`, `--tool` or `--repo`), when `bootstrap tui`
  runs, then nothing is written, the short usage names every usage error found,
  no JSON result is printed and the exit code is 2.
- Given `bootstrap tui --help --json` or `bootstrap tui --help foo`, when it
  runs, then it prints the tui help and exits 0.
- Given `bootstrap tui --help` in any terminal, when it runs, then it prints the
  tui help stating that tui takes no flags or arguments besides `--help`, writes
  nothing and exits 0, even outside a git repository.
- Given a flag tui does not accept in a directory that is not a git repository,
  when `bootstrap tui` runs, then the exit code is 2, not 3.
- Given a terminal under 80x24, when the view would open, then "make the
  terminal larger" is shown and tui waits for a resize; Ctrl-C exits 130.
- Given a successful apply, when tui finishes, then the result is human text
  only and no JSON result is printed.
- Given a repository that is not set up and a non-interactive run, when tui
  runs, then the exit code is 2.
- Given `mise` at a path other than `~/.local/bin/mise`, when tui runs and
  applies, then one warning is printed on stderr before the view opens and the
  result `warnings` repeats it.
- Given a recorded `values.tools` that misses an unremovable tool and no other
  tool of its `max: one` category is recorded, when tui runs, then the tool
  shows selected, and the result `warnings` reports "added back unremovable tool
  `<name>`". Given a setup file recorded with an older `format`, then it is read
  and not refused, and the next write records the running cli's format. Given
  two tools of one `max: one` category or an unknown tool name, then nothing is
  written and the exit code is 3 naming the fix. Given a selected tool whose
  required tool is in neither `values.tools` nor `values.dependencies`, then it
  is not a failure: the dependency is added back, "added back dependency
  `<name>`" is reported in `warnings`, and the view shows it marked `dependency`
  in the tool's tree.
- Given no `.config/bootstrap.yaml` in an interactive terminal, when tui runs,
  then the view opens with the default tools and the origin-derived repo
  pre-filled, the runtimes unchecked and shown only, marked `dependency`, in the
  trees under the default tools that need them (for example under
  `(•) pre-commit` the lines `├─ python  dependency`, `└─ uv  dependency` and,
  nested under `uv`, `└─ python  dependency`), with no `needs` or
  `also needed by` note; when the actor applies and confirms, then the files of
  the selected tools and `.config/bootstrap.yaml` are written, the exit code is
  0, the next command is `MISE_ENV=dev mise run setup:all`, `added` lists the
  direct tools only, `dependencies_added` lists the dependencies, and
  `values_changed` lists repo, commit_scopes, merge_model.develop and
  merge_model.main as `{name, from: null, to}` sorted by name.
- Given no `.config/bootstrap.yaml` and no readable origin, when the actor
  applies, then apply is refused in place until a valid repo value is entered.
- Given no `.config/bootstrap.yaml`, when the actor quits without applying, then
  nothing is written, no setup file is created and the exit code is 0 with "no
  change".
- Given a recorded version older than the running cli, when tui runs, then the
  exit code is 1 and the next command is `bootstrap init`.
- Given a setup file failing Setup config validity, or a newer recorded version,
  when tui runs, then nothing is written and the exit code is 3.
- Given the catalog holds the hidden tool `mise` and the `tool-manager`
  category, when the view opens, then neither is shown, and no tool list of the
  plan or the result names `mise`; the paths of its files are listed like any
  path.
- Given tool rows in the view, when it opens, then each tool row shows the
  tool's `purpose` after its name in the muted role (for example
  `[x] dprint  Formats code and documents`), and a category row keeps its
  purpose.
- Given `git` or `pre-commit` (`removable: false`), when the view opens, then it
  shows selected, its category offers no "none", and choosing another tool of
  the category asks "replace <old> with <new>?"; declining keeps the old tool.
- Given a `max: one` category holding a tool, when the actor chooses another,
  then tui asks "replace <old> with <new>?"; declining keeps the old tool.
- Given a `max: one` category holding a removable tool, when the actor chooses
  "none", then no consent prompt is asked, the apply plan lists the removal, the
  single confirm covers it, and on success the tool is in `removed`, not in
  `replaced`.
- Given a `values.dependencies` name the running catalog does not have, when the
  actor applies, then the drop is a repair listed in the plan that needs the
  confirm even alone; on confirm the name is dropped from `values.dependencies`
  on the rewrite, `warnings` reports "dropped unknown dependency `<name>`", and
  with no other change the result prints no next command.
- Given a selection change, when the actor quits without applying, then nothing
  is written and the exit code is 0 with "no change".
- Given a selection change, when the actor applies and declines the plan, then
  the view returns with the selection intact and nothing is written.
- Given `dprint` unchecked and `node` and `pnpm` unselected, when the actor
  selects `dprint`, then it is not refused, the row shows the tree
  `├─ node  dependency` and `└─ pnpm  dependency` (with `└─ node  dependency`
  nested under `pnpm`), `node` and `pnpm` stay unchecked, and after apply
  `values.tools` lists `dprint`, `values.dependencies` records `node` and `pnpm`
  with their parents, `added` lists `dprint` and `dependencies_added` lists
  `node` and `pnpm`.
- Given `dprint` selected with `node` a dependency, when the actor selects
  `node`, then it is checked as a direct tool, it shows its own tree if it
  requires anything, it appears unmarked in the trees of `dprint` and `pnpm`,
  and after apply `node` is in `values.tools` and not in `values.dependencies`,
  and `added` lists `node`.
- Given the pinned example view, when the actor moves focus, then it moves over
  the tool rows only and never lands on a tree line.
- Given a selected tool that another selected tool still needs and no other
  alternative of its slot selected, when the actor deselects it, then the change
  is refused in place with "cannot deselect `<tool>`: needed by `<parents>`" and
  the view stays open.
- Given `taplo` and `dprint` selected, when the actor deselects `dprint` first,
  then it is refused naming `taplo`; when the actor deselects `taplo` and then
  `dprint`, then both changes are accepted, because the check uses the selection
  after each change.
- Given a dependency whose last parent is deselected, when the actor applies,
  then the plan lists the dependency as pruned, its files are deleted,
  `values.dependencies` no longer records it, and the result lists it in
  `dependencies_pruned`, not in `removed`.
- Given a target that is modified, staged, untracked or ignored (the setup file
  included), when the actor applies, then the view closes, nothing is written
  and the exit code is 1 listing the paths.
- Given a target that is a directory or symlink, when the actor applies, then
  nothing is written and the exit code is 3 naming the path, also when dirty
  targets are found together.
- Given `.config/bootstrap.yaml` changed after the view opened (its content
  differs from the bytes read when the view opened, or it appeared or vanished),
  when the actor applies, then nothing is written and the exit code is 1 "setup
  file changed — run `bootstrap tui` again". Given a touch with identical
  content, then it is not a change and apply proceeds.
- Given the selection and values are unchanged, when the actor applies, then
  nothing is written and the exit code is 0 "no change". Given an unchanged
  selection but a render that differs from the working copy, then the differing
  paths are targets and the plan is shown.
- Given a removable tool deselected, when apply succeeds, then its files that
  are not `create_only` are deleted, shared files are reported `changed` and
  `values.tools` drops it.
- Given a replacement consented in a `max: one` category, when apply succeeds,
  then the files are swapped and `replaced` lists `{from, to}`.
- Given a recorded path the running cli does not render and that is present on
  disk, when apply succeeds, then it is not deleted, it stays in `files` and the
  result lists it `orphaned`.
- Given only bookkeeping (a recorded path not rendered and absent on disk, a
  stale dependency the catalog has that no selected tool needs, or a wrong
  parent list), when the actor applies, then nothing is written, no prompt is
  asked, "no change" is printed and the exit code is 0. Given the same
  bookkeeping and another change, when apply succeeds, then the stale path is
  dropped, the stale dependency is pruned (listed in `dependencies_pruned`) and
  the parent lists are corrected.
- Given a missing dependency that the selection after the request leaves
  unneeded, when apply succeeds, then it is not added back and not reported.
- Given a `max: one` category holding a dependency (not reachable in 1.0), when
  the view opens, then its radio group shows no marked row and `none` is not
  marked; when the actor chooses another tool and consents, then that tool is
  recorded in `values.dependencies`, not in `values.tools`, and `added` does not
  list it.
- Given only a repo or merge-model value changed, when apply succeeds, then the
  exit code is 0 and `values_changed` lists `{name, from, to}` sorted by name.
- Given a set-up repository and an apply with no tool added, no repair and no
  `mise` config file created, changed or deleted, when the apply succeeds, then
  the result prints no next command and the exit code is 0.
- Given a set-up repository, when the actor applies a change whose only addition
  is one or more dependencies (no direct tool added, no replacement, no repair),
  then `dependencies_added` lists them, the result prints the next command
  `MISE_ENV=dev mise run setup:all` per [errors](../../../conventions.md#errors)
  and the exit code is 0.
- Given a failed restore after a write failure, when tui runs, then the exit
  code is 3 listing the paths not restored.
- Given an edited repo or merge-model value, when the actor applies and
  confirms, then every rendered file that uses the value shows the new value and
  the setup file records it.
- Given an invalid value (for example a merge model `squash`), when the actor
  enters it, then it is refused in place with the rule and the view stays open.
- Given a deselected tool with a `create_only` file, when the actor applies,
  then the file stays, it is reported `kept` and it is not deleted.
- Given a tool added in the view, when apply succeeds, then its files exist,
  `values.tools` lists it, the result prints `added` and the next command
  `MISE_ENV=dev mise run setup:all`, and a re-run of `bootstrap init -y` reports
  every file `unchanged`.
- Given a write failure injected mid-apply, when the actor confirms, then the
  repository is byte-identical to before and the exit code is 3 naming the
  failing path.
- Given an interrupt (Ctrl-C) in the view or during the write, when tui is
  running, then the repository is byte-identical to before and the exit code
  is 130.
- Given `mise` not on the `PATH`, or no git repository, when tui runs, then
  nothing is written and the exit code is 3.
- Given a `max: one` category whose held tool (direct or a dependency) a
  selected tool still needs, when the actor chooses another tool of the category
  that is not an alternative of that slot, then no consent prompt is shown,
  "cannot replace `<tool>`: needed by `<parents>`" is shown, and the selection
  is unchanged (no 1.0 catalog tool reaches this case).
- Abuse case: n/a, runs locally with the caller's own permissions on the
  caller's own repository; every value is validated per
  [baseline](../../../conventions.md#baseline) boundary-validation.

## References

- [baseline](../../../conventions.md#baseline),
  [safety](../../../conventions.md#safety),
  [errors](../../../conventions.md#errors),
  [config](../../../conventions.md#config)
- [design-system](../../../design-system.md#terminal-ux), Terminal UX
- API surface: N/A, no service project; the flow is a local command
- Screens surface: N/A, cli has no screen platform
