---
type: vwf-flow
title: Set up a repository
description: One command sets up a git repository from the shared source, and
  the same command on a set-up repository keeps it on that source, with nothing
  lost and nothing half-written.
status: reviewed
implementation: none
---

# Flow: Set up a repository

## Purpose

`bootstrap init` writes the baseline files of the selected tools into a git
repository and records what it did in one config file. Run again on a set-up
repository it is a re-run: it renders with the running bootstrap and writes only
what changed.

Serves: [Fast new-repo setup](../../../product.md#goal-fast-setup),
[Zero setup drift](../../../product.md#goal-zero-drift)

## Trigger & Actors

| Actor                      | May trigger                                                       | Authorization                       | Audit-recorded                                                             |
| -------------------------- | ----------------------------------------------------------------- | ----------------------------------- | -------------------------------------------------------------------------- |
| Repo owner, AI agent or CI | the command `bootstrap init`, in any terminal, with flags or `-y` | write access to the repository root | no — no audit foundation; git history keeps every replaced or deleted file |

Interactive setup is `bootstrap tui`
([Select tools](../160-select-tools/index.md)).

## Steps

Usage errors come first and exit 2 before the preflight and before
`.config/bootstrap.yaml` is read. Nothing is written. They are:

- an unknown flag;
- a flag value that breaks its rule (`--merge-develop` or `--merge-main` not
  `direct`\|`pr`; an `--add-scope` or `--remove-scope` value not lowercase
  kebab-case; `--repo` not two or more `/`-separated segments), with a message
  naming the flag and the allowed form;
- `--reset-scope` together with `--add-scope` or `--remove-scope`;
- the same scope value given to both `--add-scope` and `--remove-scope`;
- an unknown `--tool` name ([Tool](../../../entities/tool/index.md)), or two
  `--tool` names of one `max: one` category
  ([Tool category](../../../entities/tool-category/index.md)).

1. Init runs the preflight per [errors](../../../conventions.md#errors); nothing
   is written on failure.
2. Init looks for `.config/bootstrap.yaml`. Absent → a first run. Present → a
   re-run: init validates it per
   [Setup config validity](../../../entities/setup-config/index.md#validity),
   repairs it per
   [repair on read](../../../entities/setup-config/index.md#repair), and takes
   every value and the tool selection from it. Init is exempt from the version
   guard: an older `version` or `format` is never refused, it is an upgrade
   re-run handled as any re-run (step 3 onward). A newer `version` or `format`
   fails that validity check.

   [Setup config](../../../entities/setup-config/index.md)
3. Actor supplies the values:

   | Value                | Flag                                                                         | Required          | Default                                                |
   | -------------------- | ---------------------------------------------------------------------------- | ----------------- | ------------------------------------------------------ |
   | repo path            | `--repo`                                                                     | yes               | the recorded value, else read from the remote `origin` |
   | commit scopes        | `--add-scope` (repeatable) / `--remove-scope` (repeatable) / `--reset-scope` | no — zero allowed | the recorded list, else none                           |
   | merge model, develop | `--merge-develop` `direct`\|`pr`                                             | yes               | `direct`                                               |
   | merge model, main    | `--merge-main` `direct`\|`pr`                                                | yes               | `pr`                                                   |

   The `--repo` default and the host (step 4) are read from the remote named
   `origin` only, on any host:

   | `origin` URL form                     | Repo default (`<path>`: owner segments then the name) | Host |
   | ------------------------------------- | ----------------------------------------------------- | ---- |
   | scp-like `<user>@<host>:<path>(.git)` | yes                                                   | yes  |
   | `ssh://…`                             | no                                                    | yes  |
   | `https://<host>/<path>(.git)`         | yes                                                   | yes  |
   | `http://…`                            | no                                                    | yes  |
   | no `origin`, or an unreadable URL     | no                                                    | no   |

   A `<path>` is `owner/name`, or nested `group/subgroup/name`. A `--repo` flag
   always overrides the origin default.

   The scope flags change the recorded list (`values.commit_scopes`):
   - `--add-scope <scope>` adds a scope; a scope already in the list is used
     once.
   - `--remove-scope <scope>` removes a scope; a scope not in the list is
     ignored and reported in `warnings` ("scope `<s>` is not recorded").
   - `--reset-scope` empties the list (zero scopes).
     ([config](../../../conventions.md#config))

   A repo path from `origin` is used without `-y`; with no repo path at all,
   init exits 2 naming `--repo`. For other required values (merge models) with
   no flag, recorded value or accepted default: init exits 2 naming the missing
   flag, and the error's next command offers `bootstrap tui` (interactive setup)
   or the missing flag; it offers `-y` only when `-y` would supply the value (a
   default exists and `-y` was not given). With `--json` the same text is the
   error's `next_command` ([errors](../../../conventions.md#errors)).
   [Setup config](../../../entities/setup-config/index.md) (recorded values)
4. Actor selects the tools. Tools and their categories come from the
   [Tool](../../../entities/tool/index.md) and
   [Tool category](../../../entities/tool-category/index.md) catalogs.
   - First-run defaults: every tool whose `default` is `on`, plus each `origin`
     tool whose `origin_hosts` contains the host of the `origin` remote (the
     host column of the step 3 table).
   - First run without `--tool`: with `-y` the selection is the default tools;
     without `-y` it is only the unremovable tools. A re-run without `--tool`
     keeps the recorded selection.
   - `--tool <name>` (repeatable) gives the full list of selected removable
     tools; unremovable tools are always added. A name given more than once is
     used once.
   - A selected tool whose `requires` tool is not in the selection after the
     request (the tools given, plus the unremovable ones) exits 2 naming the
     required tool.
5. Replacement, re-run only: a selection that puts a different tool in a
   `max: one` category than the recorded one is a replacement. It needs
   `--replace`, else exit 2; that consent also covers deleting the replaced
   tool's files, so no `-y` is needed. `-y` is never consent to replace.
   Replacing a recorded tool whose `replaceable` is false exits 2. `--replace`
   with no replacement to make is ignored.
   [Tool category](../../../entities/tool-category/index.md)
6. Init computes every render and decision in memory per
   [safety](../../../conventions.md#safety).
   - A re-run that drops a recorded removable tool, or replaces one, removes it
     as [Remove a tool](../150-remove-tool/index.md) does.

     Behaviour by mode:

     | Mode                | Drop of a recorded removable tool, not a replacement |
     | ------------------- | ---------------------------------------------------- |
     | plain run           | exits 2 "removing files needs -y"                    |
     | `-y` (long `--yes`) | proceeds (`-y` given)                                |
     | `--dry-run`         | reported, needs no `-y`                              |

   [Tool](../../../entities/tool/index.md)
7. Init writes only the changes: creates, changes and deletes the targets.
   [Tool](../../../entities/tool/index.md) (target files)
8. Init writes `.config/bootstrap.yaml` last, recording the format, the
   bootstrap version running, the values (`values.tools` = every selected tool)
   and `files` (every tool path init renders for the selected tools, plus any
   `orphaned` path; not `.config/bootstrap.yaml` itself). It rewrites the whole
   file per
   [write format](../../../entities/setup-config/index.md#write-format); on an
   upgrade re-run `version` and `format` move to the running bootstrap's.
   [Setup config](../../../entities/setup-config/index.md)
9. Init prints the result (`.config/bootstrap.yaml` is in `created` on a first
   run and in `changed` when rewritten) and the next command per
   [errors](../../../conventions.md#errors), with the closing line per the same
   rule. The `--json` success document has top-level keys `exit`, `created`,
   `changed`, `deleted`, `unchanged`, `kept`, `orphaned` (paths), `replaced`
   (objects `{from, to}`), `warnings` and, in that case, `next_command`. When
   there is nothing to create, change, delete or replace, init exits 0, lists
   every path by its status, the human output has no closing line and the
   `--json` document has no `next_command` key. Init never installs software or
   runs that task itself.

`--dry-run` runs the usage check and steps 1–6 and writes nothing; it exits 0 on
success and otherwise with the code of the step that fails; its output is as in
step 9; `--json` adds `"dry_run": true` (to the success or the error document).

## Guarantees

| Step             | Consistency                 | On failure                                                                                                                                                                                                                                                                                                                                                                   | Idempotency                                                                               | Load & latency          |
| ---------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ----------------------- |
| usage check, 1–6 | atomic — nothing is written | none — nothing written yet; the usage check exits 2, steps 1–6 exit 1, 2, 3 or 130 per step                                                                                                                                                                                                                                                                                  | n/a — a retry starts from the same repo state                                             | n/a — one local command |
| 7–8              | atomic — all-or-nothing     | restore per [safety](../../../conventions.md#safety); a write failure exits 3 naming the failing path, an interrupt exits 130; a failed restore exits 3 per [errors](../../../conventions.md#errors). An uncatchable kill is out of the guarantee; the config is written last, so the repo is then not marked set up (first run) or still records the earlier state (re-run) | a re-run with nothing changed writes nothing and exits 0; after a restore it starts clean | n/a — one local command |
| 9                | atomic — output only        | none — changes no repo state                                                                                                                                                                                                                                                                                                                                                 | n/a                                                                                       | n/a — one local command |

## Diagram

```mermaid
sequenceDiagram
    actor O as Actor
    participant I as Init
    participant R as Git repository
    participant K as Tool category
    participant G as Tool
    participant C as Setup config
    O->>I: bootstrap init, flags, -y
    I->>K: usage check (flags, --tool names, category limits)
    I->>R: preflight
    I->>C: read if present
    I->>G: render in memory
    I->>R: check targets
    alt exit 2
        I-->>O: usage or consent refused
    else exit 1
        I-->>O: dirty target
    else exit 3
        I-->>O: preflight, config refused, not a file
    else exit 130
        I-->>O: interrupted before any write
    else nothing to change
        I-->>O: result, exit 0, no next command
    else success
        I->>R: write changes
        I->>C: write config
        alt write failure or interrupt
            I->>R: restore from HEAD (delete created files)
            alt interrupt
                I-->>O: exit 130
            else write failure
                I-->>O: exit 3 (unrestored paths listed when restore failed)
            end
        else written
            I-->>O: result, next command
        end
    end
```

## Background Jobs

N/A — runs synchronously in one command invocation.

## Acceptance

- Given an empty git repository, when `bootstrap init` runs with all flags, then
  the files of the selected tools exist, `.config/bootstrap.yaml` records the
  version, values (`values.tools` = every selected tool) and files, and the exit
  code is 0.
- Given a first run with every required value flagged, no `-y` and no `--tool`,
  when init runs, then only the unremovable tools' files are written and the
  exit code is 0.
- Given `bootstrap init --yes` with no `--tool` on a first run, when init runs,
  then the default tools are selected and the exit code is 0.
- Given the remote `origin` is `git@<host>:o/r.git` or `https://<host>/o/r.git`,
  when init reads defaults, then the repo default is `o/r`; for
  `https://<host>/a/b/c.git` it is `a/b/c`; for `ssh://git@github.com/o/r.git`
  the repo default is not readable; and whenever the `origin` host is
  `github.com`, `github` is selected by default.
- Given a set-up repository and a re-run with nothing to create, change, delete
  or replace, when init runs (in any terminal, with or without `-y`), then
  nothing is written, an existing `create_only` file is listed `kept`, an
  unrendered recorded path is listed `orphaned`, every other file is listed
  `unchanged`, there is no closing line and no `next_command`, and the exit code
  is 0 (a `--dry-run` follows the same rule).
- Given a re-run with `-y` and `--merge-main direct` over a recorded `pr`, when
  init runs, then the flag value is rendered and recorded and the other recorded
  values are kept as supplied.
- Given a re-run whose config records a newer version or format, when init runs,
  then nothing is written and the exit code is 3 "upgrade bootstrap".
- Given a repository set up by an older bootstrap `version` or `format`, when a
  newer bootstrap re-runs init, then the older format is read and not refused,
  the changed files are rewritten, `version` and `format` move to the running
  bootstrap's, `files` is refreshed and orphaned paths are reported `orphaned`.
- Given a first run, when init finishes, then no path is reported `orphaned`.
- Given a config that fails the schema, or a `files` entry that is empty,
  absolute, uses `..` or has a wildcard, when init runs, then nothing is
  written, the exit code is 3 and the message names the failing field and says
  "fix the file, then run again".
- Given an unknown flag, an invalid flag value (for example
  `--merge-main squash`), an unknown `--tool` name or two `--tool` names of one
  `max: one` category, when init runs (even in a directory that is not a git
  repository or with a config that fails the schema), then nothing is written
  and the exit code is 2, not 3; for an invalid flag value the message names the
  flag and the allowed form.
- Given a `values.tools` name the running catalog does not have, or two tools of
  one `max: one` category, or a tool whose `requires` tool is not selected, when
  init runs, then nothing is written and the exit code is 3 naming the fix (for
  example "remove `fnox` from values.tools").
- Given a `values.tools` that misses an unremovable tool and no other tool of
  its `max: one` category is recorded, when init runs, then the tool is added
  again, the warning "added back unremovable tool `<name>`" is reported
  (`warnings`) and the exit code is 0.
- Given `.config/bootstrap.yaml` has uncommitted changes and its content differs
  from the render, when init re-runs, then nothing is written and the exit code
  is 1 listing it; once committed, it is listed `changed` when rewritten and
  `unchanged` when not.
- Given `--tool taplo` without `--tool dprint` (even with `dprint` recorded),
  when init runs, then nothing is written and the exit code is 2 naming
  `dprint`; given `--tool taplo --tool dprint`, the requires check passes on the
  selection after the request and both are selected.
- Given `--tool github --tool github`, when init runs, then github is applied
  once.
- Given a recorded list without `api`, when init runs with `--add-scope api`
  (also repeated as `--add-scope api --add-scope api`), then `api` is recorded
  once, the files are rendered again and every changed file is a target; given
  `api` already in the list, the list is unchanged.
- Given a recorded list with `api`, when init runs with `--remove-scope api`,
  then `api` is no longer recorded and the changed files are rewritten; given
  `--remove-scope web` with `web` not in the list, then the list is unchanged
  and `warnings` has "scope `web` is not recorded".
- Given a recorded list with scopes, when init runs with `--reset-scope`, then
  `values.commit_scopes` is empty and the changed files are rewritten; it also
  works together with every non-scope flag.
- Given `--reset-scope` with `--add-scope` or `--remove-scope`, when init runs,
  then nothing is written and the exit code is 2.
- Given `--add-scope api --remove-scope api`, when init runs, then nothing is
  written and the exit code is 2.
- Given `--add-scope api --remove-scope web` with `web` recorded, when init
  runs, then both are applied: `api` is added and `web` is removed.
- Given a first run with `--add-scope api`, when init runs, then the list is
  `api` alone.
- Given `--add-scope Api` (not lowercase kebab-case), when init runs, then
  nothing is written and the exit code is 2 naming the flag and the allowed
  form.
- Given the directory is not inside a git repository, when init runs, then
  nothing is written and the exit code is 3.
- Given `mise` is not on `PATH`, when init runs, then nothing is written, the
  exit code is 3 and the message says "mise not installed" with the install
  command.
- Given `mise` is found at a path other than `~/.local/bin/mise`, when init
  runs, then it continues and prints one warning naming that path (`--json`: in
  `warnings`).
- Given init runs from a subdirectory of a repository, when it succeeds, then
  `.config/bootstrap.yaml` and every tool file are at the repository root.
- Given a target path that is modified, staged, untracked or ignored in git,
  when init runs, then nothing is written, the exit code is 1 and the output
  lists the path with "commit or stash these files, then run again".
- Given a tracked target deleted in the working tree with the deletion not
  committed, when init runs, then nothing is written and the exit code is 1
  listing the path; once the deletion is committed, init creates the file again.
- Given a dirty file (including `.config/bootstrap.yaml`) whose content and mode
  already equal the render, when init runs, then it is listed `unchanged` and
  the run is not refused.
- Given a tool target path that exists as a directory or a symlink, or whose
  parent directory is a symlink or a regular file, when init runs, then nothing
  is written and the exit code is 3 naming that path.
- Given both a not-a-file target and an uncommitted target, when init runs, then
  nothing is written, the exit code is 3 and the output lists every path of both
  kinds.
- Given an existing `create_only` file, when init runs, then it is untouched and
  listed `kept`.
- Given a path in `files` that the running bootstrap no longer renders for a
  selected tool, when init runs, then it is left in place and listed `orphaned`;
  a recorded path that is no longer rendered and is absent on disk leaves
  `files` silently.
- Given a re-run that needs a replacement (a different tool in a `max: one`
  category than recorded) without `--replace`, when init runs in any terminal
  (also with `-y`, `--dry-run` or `--json`), then it does not prompt, nothing is
  written and the exit code is 2; with `--replace` the replacement is applied
  (deleting the replaced tool's files) and listed in `replaced`.
- Given a re-run that replaces a tool with `--replace` and without `-y`, in any
  terminal, when init runs, then the replacement is applied, the replaced tool's
  non-`create_only` files are deleted, it is listed in `replaced` and the exit
  code is 0.
- Given `--replace` and no replacement to make, when init runs, then the flag is
  ignored and the run proceeds as without it.
- Given `-y` or `--yes`, when init runs, then both behave identically.
- Given a replacement of a tool whose `replaceable` is false, when init runs,
  then nothing is written and the exit code is 2.
- Given a re-run with `-y` whose selection drops a recorded removable tool, when
  init runs, then that tool's non-`create_only` files are deleted and its
  `create_only` files are kept and reported.
- Given a re-run whose selection drops a recorded removable tool (whether or not
  it deletes files), when init runs (including `--json`) without `-y`, then
  nothing is written and the exit code is 2 "removing files needs -y"; with `-y`
  the files are deleted, and `--dry-run` (also `--dry-run --json`) reports the
  deletions without needing `-y`.
- Given a successful run, when init finishes, then a first run's `created` lists
  `.config/bootstrap.yaml`, every task file has mode `0755` and every path list
  is sorted by path.
- Given a write failure injected mid-run, when init runs, then every changed
  target is restored from `HEAD`, every created file is deleted and the exit
  code is 3 naming the failing path.
- Given an interrupt signal mid-write, when init is running, then the working
  tree is as before and the exit code is 130.
- Given `--dry-run`, when init runs, then no file changes, the would-be created,
  changed, deleted, kept, orphaned and replaced paths are reported, and the exit
  code is 0.
- Given a required value with no flag, no recorded value and no default accepted
  by `-y` (also with `--dry-run` or `--json`), when init runs, then it does not
  prompt, the exit code is 2 and the message names the missing flag; with
  `--json` stdout is one JSON document carrying the error.
- Given an interactive terminal and a missing required value without `-y`, when
  init runs, then it does not prompt, the exit code is 2 naming the flag, and
  the error's next command offers `bootstrap tui` or the missing flag, and `-y`
  only when `-y` would supply the value.
- Given a first run, a repo path readable from `origin`, `--merge-develop` and
  `--merge-main` given and no `-y`, when init runs with no `--repo`, then the
  path from `origin` is used and the exit code is 0.
- Given a re-run with a recorded `values.repo` that differs from the path read
  from `origin`, when init runs with no `--repo`, then the recorded value is
  kept; with `--repo` the flag value is rendered and recorded.
- Given a first run, no `--repo` and no repo path readable from `origin`, when
  init runs with or without `-y`, then there is no default, it does not prompt,
  the exit code is 2 and the message names `--repo`.
- Given a successful run or a `--dry-run` that adds a tool or creates, changes
  or deletes a `mise` config file, when init finishes, then the human output
  ends with exactly "commit the changes, then run
  `MISE_ENV=dev mise run setup:all`"; with `--json` the success document has
  `exit` 0, `created`, `changed`, `deleted`, `unchanged`, `kept`, `orphaned`,
  `replaced` (each `{from, to}`), `warnings`, and `next_command` equal to
  `MISE_ENV=dev mise run setup:all`.
- Given a re-run that only changes a value (for example `--merge-main direct`)
  and no `mise` config file, when init finishes, then the human output ends with
  exactly "commit the changes", there is no `next_command` and the exit code
  is 0.
- Given `--dry-run --json`, when init runs, then the document has the same keys
  and values a real run would return plus `"dry_run": true`, and no file
  changes; on a non-zero exit it is the same error document plus
  `"dry_run": true`.
- Given `--json`, when init runs, then stdout parses as exactly one JSON
  document and contains nothing else.
- Given a recorded `values.tools` that misses an unremovable tool while another
  tool of its `max: one` category is recorded, when init runs, then nothing is
  written and the exit code is 3 naming the fix
  ([Validity](../../../entities/setup-config/index.md#validity)).
- Abuse case: n/a — runs locally with the caller's own permissions on the
  caller's own repository; no remote surface; every input is validated per
  [baseline](../../../conventions.md#baseline) boundary-validation.

## References

- [baseline](../../../conventions.md#baseline),
  [safety](../../../conventions.md#safety),
  [errors](../../../conventions.md#errors),
  [config](../../../conventions.md#config),
  [tool-versions](../../../conventions.md#tool-versions)
- [Terminal UX](../../../design-system.md#terminal-ux)
- API surface: N/A — no service project; the flow is a local command
- Screens surface: N/A — cli has no screen platform
