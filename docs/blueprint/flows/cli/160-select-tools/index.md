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

| Condition                                                                                                                                                                                                                                        | Exit | Message / next command                                                                                                                                                                                                                                                      |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--help` or `--version`                                                                                                                                                                                                                          | 0    | per [Terminal UX](../../../design-system.md#terminal-ux)                                                                                                                                                                                                                    |
| A usage error: `--quiet`, `--json`, or a flag tui does not accept (`-y`, `--dry-run`, `--replace`, `--tool`, `--repo`, `--add-scope`, `--remove-scope`, `--reset-scope`, `--merge-develop`, `--merge-main`); all usage errors are found together | 2    | with `--json` (alone or with `--quiet` or another flag tui does not accept): the one JSON result with `exit` 2 and an `error` (what, why, next command `bootstrap add`/`remove`) naming every usage error found; without `--json`: the short usage naming every usage error |
| No git repository, or `mise` not on the `PATH` (preflight, as [Set up a repository](../110-setup-repository/index.md) step 1)                                                                                                                    | 3    | the preflight message of [errors](../../../conventions.md#errors)                                                                                                                                                                                                           |
| Not interactive (stdin or stdout not a terminal)                                                                                                                                                                                                 | 2    | "tui needs an interactive terminal — use `bootstrap init`/`add`/`remove`"                                                                                                                                                                                                   |
| Setup file fails [Setup config validity](../../../entities/setup-config/index.md#validity)                                                                                                                                                       | 3    | newer: "upgrade bootstrap"; schema failure: the failing field and "fix the file, then run again"; selection failure: names the fix (for example "remove `fnox` from values.tools")                                                                                          |
| Recorded `version` older than the running cli                                                                                                                                                                                                    | 1    | next command `bootstrap init`                                                                                                                                                                                                                                               |
| Setup file changed since the view opened (checked at apply): the content differs from the bytes read when the view opened, or the file appeared or vanished; a touch with identical content is not a change                                      | 1    | the view closes; "setup file changed — run `bootstrap tui` again" (result per step 10)                                                                                                                                                                                      |
| A target is a directory or symlink (checked before dirty targets, also when both are found)                                                                                                                                                      | 3    | names every such path, per [safety](../../../conventions.md#safety) precedence                                                                                                                                                                                              |
| A target is not clean per [safety](../../../conventions.md#safety)                                                                                                                                                                               | 1    | the view closes; lists the paths                                                                                                                                                                                                                                            |
| The actor quits without applying, or no target (selection and values unchanged and the render matches the working copy)                                                                                                                          | 0    | "no change"                                                                                                                                                                                                                                                                 |

1. Tui accepts `--no-color` (honored in the view) and `--verbose` (each per-file
   decision in the plan).
2. The warning for `mise` not at `~/.local/bin/mise` goes in the result
   `warnings`.
3. Tui reads `.config/bootstrap.yaml`. With the file absent, tui opens in
   first-run mode, not an exit: the view is pre-filled with the first-run
   defaults, per [Terminal UX](../../../design-system.md#terminal-ux). A
   recorded `format` older than the running cli's is read per
   [write format](../../../entities/setup-config/index.md#write-format). A
   `values.tools` that misses an unremovable tool is handled per
   [repair on read](../../../entities/setup-config/index.md#repair); the warning
   goes in `warnings`. [Setup config](../../../entities/setup-config/index.md)
4. Tui opens the view: every
   [Tool category](../../../entities/tool-category/index.md) as a group with its
   tools from the [Tool](../../../entities/tool/index.md) catalog, the tools in
   `values.tools` checked (on a first run, the default tools of step 3). A tool
   with `removable: false` is selected and locked. Groups, locked controls,
   keys, marks and the terminal-size rule follow
   [Terminal UX](../../../design-system.md#terminal-ux). Ctrl-C at any moment in
   the view writes nothing, exits 130 and prints no result, per
   [errors](../../../conventions.md#errors).
5. Actor changes the selection per this table. The consent prompt is the one of
   [Terminal UX](../../../design-system.md#terminal-ux); declining keeps the old
   choice. A "none" choice is listed in the apply plan and its single confirm
   covers it.

   | Case                                       | Replace another tool?                                                                                      | Choose "none"?           | Consent prompt?       |
   | ------------------------------------------ | ---------------------------------------------------------------------------------------------------------- | ------------------------ | --------------------- |
   | `max: one`, held tool `replaceable: true`  | yes                                                                                                        | per `removable`          | yes, on a replacement |
   | `max: one`, held tool `replaceable: false` | no, the other tools of the category are locked                                                             | per `removable`          | none                  |
   | `max: one`, held tool `removable: true`    | per `replaceable`                                                                                          | yes, a removal           | no, on "none"         |
   | `max: one`, held tool `removable: false`   | no, locked (an unremovable tool is never replaceable, [Tool](../../../entities/tool/index.md) invariant 7) | not offered              | none                  |
   | `max: many`                                | n/a, `replaceable` has no effect                                                                           | n/a, a tool is unchecked | none                  |

   The requires check uses the selection as it stands after the change:
   selecting a tool whose `requires` tool is not selected, or deselecting a tool
   that another selected tool still requires, is refused in place with a message
   naming the dependency.
   [Tool category](../../../entities/tool-category/index.md)
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
   unchanged, kept, orphaned, replacements, value changes) and asks once to
   confirm. Declining returns to the view (step 4).
   [Tool](../../../entities/tool/index.md)
8. Actor quits without applying; the exit is in [exits](#exits).
9. On confirm, tui writes per [safety](../../../conventions.md#safety). It
   rewrites `.config/bootstrap.yaml` last (values, `values.tools` and `files`)
   and never commits. On a first run it renders and writes exactly as a first
   run of [Set up a repository](../110-setup-repository/index.md) does (the same
   targets, safety rules and file modes), and `.config/bootstrap.yaml` is
   created last and listed `created`.
   [Setup config](../../../entities/setup-config/index.md)
10. On every exit after the view closes except an interrupt (quit, no change, a
    refused apply, a failure, success), tui prints its human result, warnings
    included, in the result layout of
    [Terminal UX](../../../design-system.md#terminal-ux). A successful apply
    prints the result with the keys of add and remove, plus `replaced` (a list
    of `{from, to}`) and `values_changed` (`{name, from, to}` sorted by name).
    The next command `MISE_ENV=dev mise run setup:all` is printed only when the
    apply added a tool (a replacement, a repair or a first run included);
    otherwise no next command is printed and the human result has none. On a
    first run `added` lists every selected tool and `values_changed` lists every
    value (repo, commit_scopes, merge_model.develop, merge_model.main) as
    `{name, from: null, to}` sorted by name. A successful apply exits 0. The
    result is human text only; there is no `--json` result on a successful run.

## Guarantees

| Step / group                                                 | Consistency                                                         | On failure                                                                                                                                                                                                                                                                                                 | Idempotency                                                             | Load & latency         |
| ------------------------------------------------------------ | ------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- | ---------------------- |
| 1–8 (the view, the re-read, the render and the target check) | atomic, nothing is written                                          | none, nothing written yet; exit 0, 1, 2, 3 or 130 per step                                                                                                                                                                                                                                                 | n/a, a re-run starts from the same repo state                           | n/a, one local command |
| 9 (apply)                                                    | atomic, all-or-nothing per [safety](../../../conventions.md#safety) | a write failure or an interrupt restores every target from `HEAD` and deletes every created file per [baseline](../../../conventions.md#baseline) atomic-multi-write and graceful-shutdown; exit 3 naming the failing path, or 130; a failed restore exits 3 listing the paths not restored (`unrestored`) | a re-run with the same selection and values finds no change and exits 0 | n/a, one local command |
| 10                                                           | atomic, output only                                                 | none, changes no repo state                                                                                                                                                                                                                                                                                | n/a                                                                     | n/a, one local command |

## Diagram

```mermaid
sequenceDiagram
    actor O as Actor
    participant U as Select
    participant C as Setup config
    participant G as Tool and category
    participant R as Repository
    O->>U: bootstrap tui
    alt --help or --version
        U-->>O: help or version, exit 0
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
                alt config changed
                    U-->>O: exit 1
                else target is a directory or symlink
                    U-->>O: exit 3
                else dirty target
                    U-->>O: exit 1
                else no target
                    U-->>O: no change, exit 0
                else
                    U-->>O: plan
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
```

## Background Jobs

N/A, runs synchronously in one command invocation.

## Acceptance

- Given stdin or stdout is not a terminal, when `bootstrap tui` runs, then
  nothing is written and the exit code is 2.
- Given `--json`, when `bootstrap tui` runs, then nothing is written, stdout
  holds one JSON result with `exit` 2 and an `error` with `what`, `why` and
  `next_command`, and the exit code is 2.
- Given `--dry-run`, `-y`, `--replace`, `--tool`, `--repo`, `--add-scope`,
  `--remove-scope`, `--reset-scope`, `--merge-develop` or `--merge-main`, when
  `bootstrap tui` runs, then nothing is written and the exit code is 2 with
  usage.
- Given `--quiet`, when `bootstrap tui` runs, then nothing is written and the
  exit code is 2.
- Given `--json` together with `--quiet` or another flag tui does not accept
  (for example `--tool`), when `bootstrap tui` runs, then stdout is exactly one
  JSON document with `exit` 2 and an `error` naming every usage error found,
  nothing is written, and the exit code is 2.
- Given `bootstrap tui --help` (or `--version`) in any terminal, when it runs,
  then it prints the help (the version), writes nothing and exits 0, even
  outside a git repository.
- Given `--no-color`, when the view opens, then it shows without color. Given
  `--verbose`, when the plan is shown, then it lists each per-file decision.
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
- Given a recorded `values.tools` that misses an unremovable tool, when tui
  runs, then the tool shows selected, and the result `warnings` reports "added
  back unremovable tool `<name>`". Given a setup file recorded with an older
  `format`, then it is read and not refused, and the next write records the
  running cli's format. Given two tools of one `max: one` category, an unknown
  tool name, or a tool whose `requires` is not selected, then nothing is written
  and the exit code is 3 naming the fix.
- Given no `.config/bootstrap.yaml` in an interactive terminal, when tui runs,
  then the view opens with the default tools and the origin-derived repo
  pre-filled; when the actor applies and confirms, then the files of the
  selected tools and `.config/bootstrap.yaml` are written, the exit code is 0,
  the next command is `MISE_ENV=dev mise run setup:all`, and `values_changed`
  lists repo, commit_scopes, merge_model.develop and merge_model.main as
  `{name, from: null, to}` sorted by name. Given a `.config/bootstrap.yaml` that
  appears between the view opening and apply, then nothing is written and the
  exit code is 1 "setup file changed — run `bootstrap tui` again".
- Given no `.config/bootstrap.yaml` and no readable origin, when the actor
  applies, then apply is refused in place until a valid repo value is entered.
- Given no `.config/bootstrap.yaml`, when the actor quits without applying, then
  nothing is written, no setup file is created and the exit code is 0 with "no
  change".
- Given a recorded version older than the running cli, when tui runs, then the
  exit code is 1 and the next command is `bootstrap init`.
- Given a setup file failing Setup config validity, or a newer recorded version,
  when tui runs, then nothing is written and the exit code is 3.
- Given a tool with `removable: false`, when the view opens, then it shows
  selected and locked, its category offers no "none", and the actor cannot
  deselect it.
- Given a `max: one` category holding a tool with `replaceable: true`, when the
  actor chooses another, then tui asks "replace <old> with <new>?"; declining
  keeps the old tool.
- Given a `max: one` category holding a removable tool, when the actor chooses
  "none", then no consent prompt is asked, the apply plan lists the removal, the
  single confirm covers it, and on success the tool is in `removed`, not in
  `replaced`.
- Given a held tool with `replaceable: false` and `removable: true` in a
  `max: one` category, when the actor tries to choose another tool of that
  category, then the other tools are locked and nothing changes; when the actor
  chooses "none", then it is allowed as a removal. Given a `max: many` category,
  then `replaceable` has no effect.
- Given a selection change, when the actor quits without applying, then nothing
  is written and the exit code is 0 with "no change".
- Given a selection change, when the actor applies and declines the plan, then
  the view returns with the selection intact and nothing is written.
- Given a tool whose `requires` tool is not selected, or a selected tool that
  another selected tool requires, when the actor selects or deselects it, then
  the change is refused in place with a message naming the dependency and the
  view stays open.
- Given `taplo` and `dprint` selected, when the actor deselects `taplo` and then
  `dprint`, then both changes are accepted, because the check uses the selection
  after each change.
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
- Given a recorded path the running cli does not render, when apply succeeds,
  then it is not deleted, it stays in `files` and the result lists it
  `orphaned`.
- Given only a repo or merge-model value changed, when apply succeeds, then the
  exit code is 0 and `values_changed` lists `{name, from, to}` sorted by name.
- Given a set-up repository and an apply that only removes tools or only changes
  values (no tool added and no repair), when the apply succeeds, then the result
  prints no next command and the exit code is 0.
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
