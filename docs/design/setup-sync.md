# `setup-sync` design

This document describes the Mac front-end of the sync, the command
`sync-engine macos` of the Go engine (`internal/macos`). The
front-end
renders the selected tools of `macos/tools/` into `$HOME`, then runs `mise
bootstrap`.
It holds the rules that only the Mac has: `macos.toml`, the stages and exit 4,
the login
shell check, the place of mise and the symlink resolve.

See also: [sync-engine.md](sync-engine.md) for the shared rules and the other
exit
codes, [packs.md](packs.md), [tools.md](tools.md) and [mise.md](mise.md).

## The entry task

```shell
mise run updateall:configs                # render, then mise bootstrap
mise run updateall:configs --dry-run      # show the changes, write nothing
mise run updateall:configs --no-bootstrap # skip each mise bootstrap step
mise run updateall:configs --verbose      # list each file
```

| Flag             | Effect                                                                                                                   |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `--dry-run`      | Shows the changes. Writes nothing.                                                                                       |
| `--no-bootstrap` | Skips each `mise bootstrap` step.                                                                                        |
| `--verbose`      | Lists each file, also an unchanged one. Sets `SYNC_ENGINE_VERBOSE=1`, and the task does not pass the flag to the engine. |
| `-h`, `--help`   | Shows the usage.                                                                                                         |

An unknown flag, and a run inside a run, stop the task.

The `macos-setup` tool ships the tasks into `~/.config/mise/tasks/`:

| Task                | Job                                                                                   |
| ------------------- | ------------------------------------------------------------------------------------- |
| `updateall:configs` | Refreshes the cache clone and copies it. Then it runs `sync-engine macos` of the pin. |
| `updateall`         | Upgrades the machine. It runs `updateall:configs -- --no-bootstrap` as one step.      |

The old task `updateall/configs/sync` loaded `sync-engine/setup-sync` from the
copy and ran `setup_sync`. The tool no longer ships it, so the stale rule
removes it.

`updateall:configs` embeds `sync-engine/clone`. See
[sync-engine.md](sync-engine.md#how-an-entry-task-embeds-clone). Its order is:

1. It parses the flags from `"$@"`, so the task works the same under `mise run`
   and under
   `bash`.
2. It stops for a nested run, and for a `MACOS_SETUP_CONFIG_PATH` that is not
   set or is not
   a folder. This comes before any network use.
3. It takes the lock of the cache, and it does the legacy cleanup.
4. It refreshes the clone, and copies it to a temporary folder.
5. A dry run releases the lock here. A real run holds the lock to the end of
   the engine.
6. It selects the engine. With `SYNC_ENGINE_BIN` set, it uses that file: a path
   that is not an executable file stops the task with exit 2. Else it reads
   `sync-engine/VERSION` of the copy with `read -r`. A value that is not
   `MAJOR.MINOR.PATCH` stops the task with exit 5. Then it runs
   `mise install github:virajp/bootstrap@<version>`, and a failure stops the
   task with exit 5. See
   [sync-engine.md](sync-engine.md#the-release-of-the-engine).
7. It runs `sync-engine macos` with the flags, from the current folder, with
   `SYNC_ENGINE_CLONE` set to the copy: `"$SYNC_ENGINE_BIN" macos`, or
   `mise x github:virajp/bootstrap@<version> -- sync-engine macos`. The exit
   code of the engine, exit 4 too, is the exit code of the task.

The old Mac cache `${XDG_CACHE_HOME:-~/.cache}/macos-setup/` is gone. The entry
task moves
its old list `generated.txt` to `macos-generated.tsv` when the new cache has no
manifest,
then removes the old folder. It keeps the old folder when the old Mac sync still
runs, or
when the list cannot move. A failure here does not fail the sync.

`updateall` runs `updateall:configs --no-bootstrap`. On exit 4, it prints a
message and
stops at once with exit 4. It runs its other upgrade steps only when
`updateall:configs` returns 0.

## The variable `MACOS_SETUP_CONFIG_PATH`

`MACOS_SETUP_CONFIG_PATH` is required. It names the folder of `macos.toml`, the
root of the
private `macos-setup` clone of the user. It has no default. The `macos-setup`
tool writes it
in `[mise.env]` from the answer `config_path` of `[tools.macos-setup]`, and `~`
at the start
becomes `$HOME`. Before a first sync, the user sets it by hand.

`SYNC_ENGINE_CLONE` must be a folder with `sync-engine/` and `macos/tools/`. See
[sync-engine.md](sync-engine.md#the-engine-variables).

## mise

mise must be the standalone install of `curl https://mise.run | sh`, at
`~/.local/bin/mise`, so that `mise self-update` manages it. The engine
resolves the `mise`
on `PATH` through each symlink. When there is no `mise`, or its real path does
not end with
`/.local/bin/mise` (for example a Homebrew copy), the run stops and names the
path that it
found. This check is the first step of the engine, after the entry task has
refreshed the
clone.

## `macos.toml`

`macos.toml` is the values file of the Mac side. It is in the folder of
`MACOS_SETUP_CONFIG_PATH`. `macos/examples/macos.toml` is the sample.

| Key                       | Meaning                                                                                                                          |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `bootstrap_commit`        | A string. The commit of the last clean run. The sync writes it. See [sync-engine.md](sync-engine.md#bootstrap_commit).           |
| `[tools.<name>]`          | A table. It selects the tool, and each key is a value of the tool. See [sync-engine.md](sync-engine.md#the-values-merge).        |
| `[packs.<name>]`          | An empty table. It adds each `default = true` member of the pack. See [packs.md](packs.md).                                      |
| `[overrides.mise.<path>]` | It changes the generated mise config. See [mise.md](mise.md#the-overrides).                                                      |
| `[mise.*]`                | The own mise config of the user. It goes to `conf.d/zz-macos/mise.toml`. See [mise.md](mise.md#the-own-config-of-a-values-file). |

These stop the sync:

- Any other top-level key.
- `tools = [...]`, the old format, and a `packs` key that is not a table.
- `env` in a `[tools.<name>]` table: the Mac side has no environments.
- A key in a `[packs.<name>]` table.
- `[overrides.mise_env]`, and any other key of `[overrides]`.
- A `[tools.<name>]` that is not a table, a `[mise]` that is not a table, and a
  `bootstrap_commit` that is not a string.

Without `macos.toml`, the selection is empty: only the pack `base` and the
`depends`. A clean
real run creates a missing file with a header comment and the `bootstrap_commit`
line. This
line is the only change that the sync makes in the config folder.

## The selection

The selected tools are the pack `base`, the `default = true` members of each
`[packs.<name>]`,
each `[tools.<name>]`, then the `depends` of each, transitively.

- The pack `base` is fixed by name, and it must have the member
  `macos-setup = { default = true }`, because `updateall` runs
  `updateall:configs` with no
  guard.
- A tool or a pack that `macos/` does not have stops the sync. A name with
  whitespace or a
  `/`, or a name that starts with a dot, is such a name.
- The sync examines each `tools.toml` of `macos/tools/`, selected or not. It
  examines each
  `[pack.*]` table too.
- The tools are sorted by name in each stage. See [packs.md](packs.md).

## The template context

| Value          | Content                                                           |
| -------------- | ----------------------------------------------------------------- |
| `.config`      | The values of the tool of the template.                           |
| `.all`         | The values of each selected tool, by name.                        |
| `.tools`       | The names of the selected tools.                                  |
| `.env`         | `HOME`, `USER` and `MACOS_SETUP_CONFIG_PATH`.                     |
| `.sync_engine` | The absolute path of the `sync-engine/` folder of the clone copy. |

A TOML, JSON or YAML template of a Mac tool encodes a string value with
`data.ToJSON`, for
example `name = <% .config.name | data.ToJSON %>`. Its value comes from
`[tools.<name>]` of
`macos.toml`, and a string with a quote and a newline can add tables or keys to
the rendered
target.

Three Mac templates do not follow this rule, each from its old copy:

| Template                                                      | Value                                            |
| ------------------------------------------------------------- | ------------------------------------------------ |
| `macos/tools/gh/.config/gh/config.yml.tmpl`                   | The plain value of `editor`.                     |
| `macos/tools/1password/.config/1Password/ssh/agent.toml.tmpl` | The quoted value of `ssh_vault`.                 |
| `macos/tools/claude-cli/.claude/settings.json.tmpl`           | The quoted paths in `MISE_TRUSTED_CONFIG_PATHS`. |

## The mode of `~/.ssh`

When the engine handles a target below `.ssh/`, it makes sure once in each run
that the folder `$_SS_HOME/.ssh` has mode 700, on a new folder and on an
existing folder. The check uses the path in the home folder, not the real path
of a linked file. If `~/.ssh` is itself a link, the check reads and sets the
mode of the target of the link.

- A real run sets the mode, reports it and sets `_SS_CHANGED`.
- A dry run changes nothing. It reports the change only for an existing folder
  with another mode, and it reports nothing for a missing folder.
- When the mode is already 700, the sync prints nothing about the folder.
- A failed `chmod` fails that target and each later target below `.ssh/`.

## The run order

1. The engine parses the flags and examines mise, the clone and the config
   folder.
2. It reads `macos.toml` and each `tools.toml`, and it resolves the selection.
3. It merges the values of the tools.
4. It examines `[overrides.mise*]`, and generates the mise config of each tool
   and of
   `macos.toml` into a work folder.
5. It makes the file plan: each file of a selected tool, relative to `$HOME`,
   without
   `.tmpl`. The plan has each stage, with the `.config/mise/` files of the stage
   first and the
   generated mise files first of all.
6. It runs the init stage, then the main stage.
7. It writes the manifest, runs `mise bootstrap --yes` from `$HOME` and sets
   `bootstrap_commit`.

The write of one file is in
[sync-engine.md](sync-engine.md#the-write-of-one-target). A file
takes this order: the real path of its target, the seed gate, `requires`, the
unblock of a
folder path, the mode of `~/.ssh` for a target below `.ssh/`, then the render or
the copy, the parse check, the merge and the
compare.

`requires` holds paths relative to `$HOME`. A generated file with an entry that
is not there
is not handled, so the stale rule removes a listed copy with the header. A seed
and a merge
target stay as they are.

A folder where a file goes fails that target.

## Stages and exit 4

The init stage holds each selected tool with `stage = "init"` and, transitively,
each tool in
their `depends`. Each other selected tool, and the file `zz-macos/mise.toml`, is
in the main
stage. `bash`, `zsh` and `fish` have `stage = "init"` and
`depends = ["homebrew", "mise", "macos-setup"]`. `homebrew` depends on `fnox`,
and `mise`
depends on `cli`.

When the init stage is empty, the sync skips it. Else a real run does these
steps:

1. It reads the login shell with `dscl . -read ~ UserShell`.
2. It renders the init-stage tools.
3. It applies the stale rule to the manifest rows of the init-stage tools only.
4. A failure of a file or of the stale rule ends the run here, with no later
   stage.
5. It runs `mise bootstrap --yes`, unless `--no-bootstrap` is set. A failure of
   this step
   ends the run with code 1.
6. It reads the login shell again.
7. It examines whether the stage changed: it created, updated or `chmod`-ed a
   file, wrote a
   seed or a merge target, removed a stale file, changed the mode of `~/.ssh`,
   or changed the login shell.
8. If the stage changed, it writes the manifest with the rows of this stage and
   the old rows
   of the main stage. It prints `open a new terminal, then run: mise run
   updateall:configs`,
   and it stops with **exit 4**.
9. If the stage changed nothing, it continues with the main stage.

**Exit 4** is not a failure of a file: the shells, Homebrew or mise changed, and
the rest
needs a new terminal. Open a new terminal, then run the task again. The next run
finds no
change in the init stage and continues. `updateall` also stops with exit 4, and
it names
`mise run updateall`.

A dry run renders both stages at once, runs no `mise bootstrap` and never stops
with exit 4.

The other exit codes are in [sync-engine.md](sync-engine.md#the-exit-codes).

## The output

The list, the table, the last line and verbose are in
[sync-engine.md](sync-engine.md#the-output-of-a-run). These rules are only for
the Mac:

- The first line is the info line `bootstrap <short commit> · <n> tools`. With
  an
  empty commit, it reads `bootstrap an unknown commit · <n> tools`.
- The run has one table across the init and main stages, and it prints after
  `mise bootstrap`. A return of exit 4 after the init stage has no table.
- `FAILED: <list>` stays by default, before the table.
- The entry task has no line for the local sync task, because it runs the
  engine of the pin with the clone copy.
- `log` of the entry task is an info line, and `log_error` is an error line.
  With
  the style rule on, they show the info glyph or the error glyph and no prefix.
  Otherwise they keep the prefix `setup-sync: `.
- `setup-sync: running <engine command> macos` prints only with verbose, from
  the entry task. The command is `$SYNC_ENGINE_BIN`, or
  `mise x github:virajp/bootstrap@<version> -- sync-engine`.

## The failures

Inside a stage, a failed file stops no other file. After a failure in a stage:

- The stale rule of that stage still runs. A target that failed but is still
  planned stays
  handled, so its existing copy stays. A tool whose mise config cannot generate
  keeps its
  existing `conf.d/<tool>/mise.toml` in the same way.
- The manifest keeps its old content.
- No `mise bootstrap` runs, and `bootstrap_commit` stays.

## Symlinks

The Mac side resolves each target through each symlink, of the file or of a
parent folder, to
its real path. The compare, the temp file, the `mv` and a stale removal all act
on that real
path. Thus the link stays and its source changes. Use this to keep your files in
your own
repository: link them into `$HOME`, and the sync writes through the links.

A link whose source is missing is written as a new file at the source path, with
its parent
folders. A seed or a merge target behind such a link counts as missing. A
symlink loop fails
that target.

The repo side never follows a link, because a link can point outside the
repository. This
difference is on purpose.

## The manifest

The manifest is `macos-generated.tsv` of the cache. It is outside the clone, so
`git clean`
does not remove it. Its shape and the stale rule are in
[sync-engine.md](sync-engine.md#the-manifest-and-the-stale-rule).

Each generated target of a selected tool, and each generated mise file, has one
row. The file
`conf.d/zz-macos/mise.toml` has the tool `zz-macos`. A clean run writes the
manifest after the
stale rule of the main stage and before `mise bootstrap`. A dry run prints the
stale files and
writes no manifest. A tool that is not selected is not in the init stage, so the
main stage
examines its rows.

## Where `setup-sync` writes

- `$HOME`, and the sources that its symlinks point to.
- The cache: the clone, the lock and `macos-generated.tsv`.
- The `bootstrap_commit` line of `macos.toml`.
- Temp folders, which it removes on exit.

`mise bootstrap` writes where mise writes.
