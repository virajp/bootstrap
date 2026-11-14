# The Mac side

This document describes the Mac side of the sync: `bootstrap mac <command>` and
the front-end `internal/macos` of the engine. The front-end renders the selected
tools of `macos/tools/` into `$HOME`, then runs `mise bootstrap`. This document
holds the rules that only the Mac has: the repository of `macos.toml`, the
entry tasks, `macos.toml`, the stages and exit 4, the place of mise and the
symlink resolve.

See also: [cli.md](cli.md) for the command line, the shared
rules and the other exit codes, [packs.md](packs.md), [tools.md](tools.md) and
[mise.md](mise.md).

## The repository of `macos.toml`

`macos.toml` is in a git repository of the user: the private `macos-setup`
clone. Each `mac` command needs that repository.

- `bootstrap mac init` runs in the current folder, which must be a git
  repository. Else it tells the user to create one with `git init`.
- Each other `mac` command runs from any folder, and uses `--repo-path <path>`
  or `MACOS_SETUP_REPO_PATH`. The flag sets the variable for the run. A path
  that is not set, or that does not name a git repository, stops the command
  with code 2. The one exception is `mac tui`: without a path, it uses the
  current folder.

The Mac tool `mise` writes `MACOS_SETUP_REPO_PATH` in its `[mise.env]`, from
the answer `config_path` of `[tools.mise]`. `~` at the start becomes `$HOME`.
Thus the variable is set in each shell after the first sync. Before the first
sync, `mac init` sets it for its own run.

The config file of the engine has no key for the repository of `macos.toml`.
The mise env exports `MACOS_SETUP_REPO_PATH` in each shell, so the variable
always comes before a key of the file. `bootstrap config` shows the line
`mac.repo_path` with the value of `--repo-path` or of the variable. See
[cli.md](cli.md#the-config-file).

## The entry tasks

The Mac tool `mise` ships the tasks into `~/.config/mise/tasks/`, and the mise
tool `github:virajp/bootstrap`.

| Task                                                               | Job                                                                           |
| ------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| `updateall:configs`                                                | Runs `bootstrap mac sync` with its flags.                                     |
| `updateall`                                                        | Upgrades the machine. It runs `updateall:configs --no-bootstrap` as one step. |
| `upgrade:macos:download`, `upgrade:macos:upgrade`, `upgrade:power` | The upgrade tasks of macOS and of the power profile.                          |

The alias `updateall` runs `mise run updateall` from any folder.

```shell
mise run updateall:configs                # render, then mise bootstrap
mise run updateall:configs --dry-run      # show the changes, write nothing
mise run updateall:configs --no-bootstrap # skip each mise bootstrap step
mise run updateall:configs --verbose      # list each file
```

| Flag             | Effect                                                                                                                 |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `--dry-run`      | Shows the changes. Writes nothing.                                                                                     |
| `--no-bootstrap` | Skips each `mise bootstrap` step.                                                                                      |
| `--verbose`      | Lists each file, also an unchanged one. Sets `BOOTSTRAP_VERBOSE=1`, and the task does not pass the flag to the engine. |
| `-h`, `--help`   | Shows the usage.                                                                                                       |

`updateall:configs` does these steps:

1. It parses the flags from `"$@"`, so the task works the same under `mise run`
   and under `bash`. An unknown flag stops the task with code 2.
2. It stops a nested run with code 2.
3. It does the cleanup of the old Mac cache. See below.
4. It runs `bootstrap mac sync` with the flags, or `$BOOTSTRAP_BIN mac sync`.
   The exit code of the engine, exit 4 too, is the exit code of the task.

The old Mac cache `${XDG_CACHE_HOME:-~/.cache}/macos-setup/` is gone. The task
moves its old list `generated.txt` to `macos-generated.tsv` when the new cache
has no manifest, then removes the old folder. It keeps the old folder when the
old Mac sync still runs, or when the list cannot move. A failure here does not
fail the sync.

`updateall` runs `updateall:configs --no-bootstrap`. On exit 4, it prints a
message and stops at once with exit 4. It runs its other upgrade steps only when
`updateall:configs` returns 0.

A new Mac has no mise and no clone. The `setup` script of the user's
`macos-setup` repository installs mise and `bootstrap`
(`mise use -g github:virajp/bootstrap`), then runs
`bootstrap mac sync`. After
that, the Mac tool `mise` keeps `bootstrap` as a mise tool, and the version
check applies.

## mise

mise must be the standalone install of `curl https://mise.run | sh`, at
`~/.local/bin/mise`, so that `mise self-update` manages it. The engine resolves
the `mise` on `PATH` through each symlink. When there is no `mise`, or its real
path does not end with `/.local/bin/mise` (for example a Homebrew copy), the run
stops with code 2 and names the path that it found. This check is the first
step of the front-end.

`BOOTSTRAP_CLONE` must be a folder with `cli/` and `macos/tools/`. The
command line sets it. See
[cli.md](cli.md#the-cache-and-the-clone).

## `macos.toml`

`macos.toml` is the values file of the Mac side. It is at the root of the
repository of `macos.toml`. `macos/examples/macos.toml` is the sample. Its
overrides file is `macos.overrides.toml` next to it. See
[cli.md](cli.md#the-values-of-a-tool).

| Key                       | Meaning                                                                                                                          |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `format`                  | The integer `2`, the form of the file. A file without it is an old form: run `bootstrap mac init`.                               |
| `bootstrap_commit`        | A string. The commit of the last clean run. The sync writes it. See [cli.md](cli.md#bootstrap_commit).                           |
| `[tools.<name>]`          | A table. It selects the tool, and each key is a value of the tool. See [cli.md](cli.md#the-values-of-a-tool).                    |
| `[overrides.mise.<path>]` | It changes the generated mise config. See [mise.md](mise.md#the-overrides).                                                      |
| `[mise.*]`                | The own mise config of the user. It goes to `conf.d/zz-macos/mise.toml`. See [mise.md](mise.md#the-own-config-of-a-values-file). |

These stop the sync:

- Any other top-level key.
- `tools = [...]`, the old format.
- `env` in a `[tools.<name>]` table: the Mac side has no environments.
- A `format` other than `2`.
- `[overrides.mise_env]`, and any other key of `[overrides]`.
- A `[tools.<name>]` that is not a table, a `[mise]` that is not a table, and a
  `bootstrap_commit` that is not a string.
- An old form: run `bootstrap mac init`. The old forms are a file without
  `format = 2`, a `[packs.<name>]` table or a `packs` key, and the table of
  the old Mac entry tool, whose values go to `[tools.mise]`.
  `docs/sync/migrate.md` names the old tool. `mac init` migrates them as on
  the repo side: see
  [bootstrap-config.md](bootstrap-config.md#old-forms).

Without `macos.toml`, the selection is only the required members of the pack
`base` and their `depends`. A clean real run creates a missing file with a
header comment, `format = 2`, a table for each required member of `base`, and
the `bootstrap_commit` line.

## The selection

The selected tools are the required members of the pack `base`, each
`[tools.<name>]`, then the `depends` of each, transitively. The values file
keeps no record of a pack: `mac add pack` writes the tables of its members.

- The pack `base` is fixed by name. Its member `mise` has
  `required = true`, because each Mac needs the entry tasks. The sync writes
  an empty table for each required member that has none. See
  [cli.md](cli.md#the-tables-of-the-required-members).
- A tool that `macos/` does not have stops the sync. A name with
  whitespace or a `/`, or a name that starts with a dot, is such a name.
- The sync examines each `tools.toml` of `macos/tools/`, selected or not. It
  examines each `[pack.*]` table too.
- A Mac `tools.toml` accepts the catalog keys `description`, `summary`,
  `category` and `selectable`, with the rules of
  [tools.md](tools.md#the-catalog-keys). `summary` and `category` are
  necessary: a tool without one of them stops the sync with code 2. Its
  `category` names an id of `macos/categories.toml`. `hosts` is for repo tools
  only, and it stops the
  sync with `hosts is for repo tools only`.
- `mac add tool` and `mac replace tool` refuse a named tool with
  `selectable = false`, and `mac check` shows a warning for it. See
  [cli.md](cli.md#named-tools-that-a-command-refuses).
- The tools are sorted by name in each stage. See [packs.md](packs.md).

## The template context

| Value      | Content                                                   |
| ---------- | --------------------------------------------------------- |
| `.config`  | The values of the tool of the template.                   |
| `.all`     | The values of each selected tool, by name.                |
| `.tools`   | The names of the selected tools.                          |
| `.env`     | `HOME`, `USER` and `MACOS_SETUP_REPO_PATH`.               |
| `.cli_dir` | The absolute path of the `cli/` folder of the clone copy. |

A TOML, JSON or YAML template of a Mac tool encodes a string value with
`data.ToJSON`, and a merge spec with `data.ToTOMLInline`, for example
`name = <% .config.name | data.ToJSON %>`. Its value comes from
`[tools.<name>]` of `macos.toml`, and a string with a quote and a newline can
add tables or keys to the rendered target.

Two Mac sources do not follow this rule, each from its old copy:

| Source                                                              | Value                                            |
| ------------------------------------------------------------------- | ------------------------------------------------ |
| `macos/tools/1password/files/.config/1Password/ssh/agent.toml.tmpl` | The quoted value of `ssh_vault`.                 |
| `macos/tools/claude-cli/files/.claude/settings.json.merge.toml`     | The quoted paths in `MISE_TRUSTED_CONFIG_PATHS`. |

## The run order

1. The front-end examines its flags, mise, the clone and the repository of
   `macos.toml`.
2. It reads `macos.toml` and each `tools.toml`, and it resolves the selection.
3. It merges the values of the tools, with the derive scripts. A question of a
   selected tool without an answer stops the run with code 2. An optional
   question can have no answer. A question of a group needs no answer, but its
   group needs one answer that is not empty, else the run stops with code 2.
   The message names the question or the group, and the hint is
   `run bootstrap mac init`.
4. It runs the check scripts.
5. It examines `[overrides.mise*]`, and generates the mise config of each tool
   and of `macos.toml` into a work folder.
6. It makes the file plan: each file of a selected tool, relative to `$HOME`.
   The plan has each stage, with the `.config/mise/` files of the stage first
   and the generated mise files first of all.
7. It runs the init stage, then the main stage.
8. It writes the manifest, runs `mise bootstrap --yes` from `$HOME`, runs the
   `after` scripts of the main stage, and sets `bootstrap_commit`.

The write of one file is in
[cli.md](cli.md#the-write-of-one-target). A file takes this
order: the real path of its target, the seed gate, `requires`, the unblock of a
folder path, the `dir_perm` of its folder, then the render or the copy, the
parse check, the merge and the compare.

`requires` holds paths relative to `$HOME`. An owned file with an entry that is
not there is not handled, so the stale rule removes a listed copy with the
header. A seed and a shared file stay as they are.

The `ssh` tool gives each file below `.ssh/` `dir_perm = "700"`, so `~/.ssh` has
mode 700. The check uses the path in the home folder, not the real path of a
linked file.

## Stages and exit 4

The init stage holds each selected tool with `stage = "init"` and,
transitively, each tool in their `depends`. Each other selected tool, and the
file `zz-macos/mise.toml`, is in the main stage. `bash`, `zsh` and `fish` have
`stage = "init"` and `depends = ["homebrew", "mise"]`. `homebrew` depends on
`fnox`, and `mise` depends on `cli`.

When the init stage is empty, the sync skips it. Else a real run does these
steps:

1. It renders the init-stage tools.
2. It applies the stale rule to the manifest rows of the init-stage tools only.
3. A failure of a file or of the stale rule ends the run here, with no later
   stage.
4. It runs `mise bootstrap --yes`, unless `--no-bootstrap` is set. A failure of
   this step ends the run with code 1.
5. It runs the `after` script of each init-stage tool. A script that asks for a
   restart gives a reason.
6. If the stage changed a file, a seed, a shared file, a stale file or a folder
   mode, or a script asked for a restart, the run writes the manifest with the
   rows of this stage and the old rows of the main stage. It prints the reasons
   and `open a new terminal, then run: mise run updateall:configs`, and it
   stops with **exit 4**.
7. Else it continues with the main stage.

After the main stage, an `after` script can also ask for a restart: the run
prints the reasons and stops with exit 4, before it sets `bootstrap_commit`.

**Exit 4** is not a failure of a file: the shells, Homebrew or mise changed, and
the rest needs a new terminal. Open a new terminal, then run the task again. The
next run finds no change in the init stage and continues. `updateall` also stops
with exit 4, and it names `mise run updateall`.

A dry run renders both stages at once, runs no `mise bootstrap` and no `after`
script, and never stops with exit 4.

The other exit codes are in [cli.md](cli.md#the-exit-codes).

## The output

The list, the last line and verbose are in
[cli.md](cli.md#the-output-of-a-run). These rules are only for
the Mac:

- The first line is `↻ bootstrap <short commit> · <n> tools`. With an empty
  commit, it reads `↻ bootstrap an unknown commit · <n> tools`.
- `mise bootstrap` prints `▸ Running mise bootstrap --yes` before it and
  `✔ mise bootstrap done` after a success. The output of mise stays as it is.
- The run has one last line across the init and main stages, and it prints after
  `mise bootstrap`. A return of exit 4 after the init stage has no last line.
- `FAILED: <list>` stays by default, before the last line.
- `log` of the entry task is an info line, and `log_error` is an error line.
  With the style rule on, they show the glyphs `·` and `✗`. Otherwise they have
  the prefixes `[INFO] ` and `[ERROR] `.

## The failures

Inside a stage, a failed file stops no other file. After a failure in a stage:

- The stale rule of that stage still runs. A target that failed but is still
  planned stays handled, so its existing copy stays. A tool whose mise config
  cannot generate keeps its existing `conf.d/<tool>/mise.toml` in the same way.
- The manifest keeps its old content.
- No `mise bootstrap` runs, and `bootstrap_commit` stays.

## Symlinks

The Mac side resolves each target through each symlink, of the file or of a
parent folder, to its real path. The compare, the temp file, the `mv` and a
stale removal all act on that real path. Thus the link stays and its source
changes. Use this to keep your files in your own repository: link them into
`$HOME`, and the sync writes through the links.

A link whose source is missing is written as a new file at the source path, with
its parent folders. A seed or a shared file behind such a link counts as
missing. A symlink loop fails that target.

The repo side never follows a link, because a link can point outside the
repository. This difference is on purpose.

## The manifest

The manifest is `macos-generated.tsv` of the cache. It is outside the clone, so
`git clean` does not remove it. Its shape and the stale rule are in
[cli.md](cli.md#the-manifest-and-the-stale-rule).

Each owned target of a selected tool, each generated mise file, and each item
that bootstrap added to a shared file has one row. The file
`conf.d/zz-macos/mise.toml` has the tool `zz-macos`. A clean run writes the
manifest after the stale rule of the main stage and before `mise bootstrap`. A
dry run prints the stale files and writes no manifest. A tool that is not
selected is not in the init stage, so the main stage examines its rows.

## Where the Mac side writes

- `$HOME`, and the sources that its symlinks point to.
- The cache: the clone, the lock and `macos-generated.tsv`.
- `macos.toml`: `bootstrap_commit`, the tables of the required members of
  `base`, and the writes of `mac init`, `mac add`, `mac remove`,
  `mac replace` and `mac detect --add`.
- Temp folders, which it removes on exit.

`mise bootstrap` writes where mise writes.
