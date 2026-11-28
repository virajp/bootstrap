# The Mac side

This document describes the Mac side of the sync: `bootstrap mac <command>` and
the front-end `internal/macos` of the engine. The front-end renders the selected
tools of `macos/tools/` into `$HOME`. It installs nothing: `updateall` runs
`mise bootstrap`. Only the first setup of a Mac, which `mac migrate` runs, runs
`mise bootstrap` once. This document holds the rules that only the Mac has: the
place of `macos.toml`, the entry tasks, `macos.toml`, the first setup and
exit 4, the place of mise and the symlink resolve.

See also: [cli.md](cli.md) for the command line, the shared
rules and the other exit codes, [packs.md](packs.md), [tools.md](tools.md) and
[mise.md](mise.md).

## The place of `macos.toml`

`macos.toml` has one fixed place on a Mac:
`$XDG_CONFIG_HOME/bootstrap/macos.toml`, or
`~/.config/bootstrap/macos.toml` when `XDG_CONFIG_HOME` is not set. The engine
config file `config.toml` is in the same folder. Each `mac` command uses this
path from any folder, `mac migrate` too. Thus a `mac migrate` in a different
folder, for example in a clone of a project, never writes a second
`macos.toml`.

The folder needs no git repository. A clean real run creates the folder when
it is not there.

A Mac where a sync finished (it has a manifest) and with no `macos.toml` at
this place stops each `mac sync` and `mac migrate` with code 2:
`a sync finished on this Mac, but <path> does not exist - move your macos.toml
there (docs/sync/migrate.md, M56)`. Without this stop, the run selects only
the members of `base`, and the stale rule removes the files of each other
tool. The backup of `macos.toml` in git comes from the key `backup`:
the sync links `~/.config/bootstrap/` into the backup clone. See
[The backup clone](#the-backup-clone).

The engine sets `BOOTSTRAP_VALUES` to a temporary folder for the plan of a
change, as on the repo side. Thus the dry run of a plan reads the planned
`macos.toml`. A user does not set this variable.

Old forms: the flag `--repo-path`, the variable `MACOS_SETUP_REPO_PATH`, the
setting `mac.repo_path` and the answer `config_path` of `[tools.mise]` are
gone. `mac migrate` moves the answer `config_path` to the key `backup`. See
`docs/sync/migrate.md` (M56).

## The entry tasks

The Mac tool `mise` ships the tasks into `~/.config/mise/tasks/`, and the mise
tool `github:virajp/bootstrap`. Each task is a Python uv script with a `.py`
name, and mise removes `.py` from the task name. The tasks need uv and Python
3.14, so the Mac `mise` tool depends on `python` and `uv` (see
[The first setup and exit 4](#the-first-setup-and-exit-4)). `updateall:configs`
is standalone:
it imports no library.

| Task                                                               | Job                                                                                   |
| ------------------------------------------------------------------ | ------------------------------------------------------------------------------------- |
| `updateall:configs`                                                | Runs `bootstrap mac sync` with its flags.                                             |
| `updateall`                                                        | Upgrades the machine. It runs `updateall:configs` as one step, then `mise bootstrap`. |
| `upgrade:macos:download`, `upgrade:macos:upgrade`, `upgrade:power` | The upgrade tasks of macOS and of the power profile.                                  |

The alias `updateall` runs `mise run updateall` from any folder.

```shell
mise run updateall:configs                # render the configs
mise run updateall:configs --dry-run      # show the changes, write nothing
mise run updateall:configs --no-bootstrap # skip the mise bootstrap of a first setup
mise run updateall:configs --verbose      # list each file
```

| Flag             | Effect                                                                                                                 |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `--dry-run`      | Shows the changes. Writes nothing.                                                                                     |
| `--no-bootstrap` | Skips the `mise bootstrap` of a first setup, when the sync runs a pending `mac migrate`.                               |
| `--verbose`      | Lists each file, also an unchanged one. Sets `BOOTSTRAP_VERBOSE=1`, and the task does not pass the flag to the engine. |
| `-h`, `--help`   | Shows the usage.                                                                                                       |

`updateall:configs` does these steps:

1. It parses the flags from its arguments, so the task works the same under
   `mise run` and directly, with uv on PATH. An unknown flag stops the task
   with code 2.
2. It stops a nested run with code 2.
3. It does the cleanup of the old Mac cache. See below.
4. It runs `bootstrap mac sync` with the flags, or `$BOOTSTRAP_BIN mac sync`.
   The exit code of the engine, exit 4 too, is the exit code of the task.

The old Mac cache `${XDG_CACHE_HOME:-~/.cache}/macos-setup/` is gone. The task
moves its old list `generated.txt` to `macos-generated.tsv` when the new cache
has no manifest, then removes the old folder. It keeps the old folder when the
old Mac sync still runs, or when the list cannot move. A failure here does not
fail the sync.

`updateall` runs `updateall:configs`. On exit 4, it prints a
message and stops at once with exit 4. It runs its other upgrade steps only when
`updateall:configs` returns 0.

A new Mac has no mise and no clone. The script `macos/setup` of bootstrap
prepares it:

```shell
BACKUP_REPO= < owner > / < repo > curl -fsSL https://raw.githubusercontent.com/virajp/bootstrap/main/macos/setup | sh
```

1. It installs Homebrew (with the Xcode Command Line Tools and git) and the
   standalone mise, each only when it is missing.
2. It reads the backup repository from `BACKUP_REPO`, else it asks for it.
   The clone goes to `${PROJECTS_DIR:-~/Projects}/github.com/<owner>/<repo>`.
3. When gh has no login, it runs `gh auth login` through `mise x`. It clones
   the private repository with the token of gh for that one clone, and sets
   the remote to the ssh URL. The token is not written to the clone.
4. When the clone has `dotfiles/.config/bootstrap/macos.toml`, it links
   `~/.config/bootstrap` to that folder and runs `bootstrap mac migrate`. The
   file already has `backup`. Else it runs
   `bootstrap mac migrate --backup <clone>`, which writes a new `macos.toml`
   and asks the questions in the terminal. It stops with code 2 when
   `~/.config/bootstrap` exists and is not that link.
5. It runs the engine with `mise x github:virajp/bootstrap@latest --`, so no
   mise config is written before the backup step makes the link
   `~/.config/mise`.
6. On exit 4 of the first setup, it tells the owner to open a new terminal and
   run `mise run updateall`, then parks the old terminal.

After that, the Mac tool `mise` keeps `bootstrap` as a mise tool, and the
version check applies. A Mac with configs of its own, or with the old
`dotfiles/` links, is not a new Mac: [the adoption
prompt](../setup/index.md) handles it.

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

`macos.toml` is the values file of the Mac side. It is at the fixed place of
[The place of `macos.toml`](#the-place-of-macostoml).
`macos/examples/macos.toml`
is the sample. Its
overrides file is `macos.overrides.toml` next to it. See
[cli.md](cli.md#the-values-of-a-tool).

| Key                       | Meaning                                                                                                                          |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `backup`                  | A string, not empty. The backup clone. `~` at the start is `$HOME`. See [The backup clone](#the-backup-clone).                   |
| `format`                  | The integer `2`, the form of the file. A file without it is an old form: run `bootstrap mac migrate`.                            |
| `bootstrap_commit`        | A string. The commit of the last clean run that changed a file. The sync writes it. See [cli.md](cli.md#bootstrap_commit).       |
| `[tools.<name>]`          | A table. It selects the tool, and each key is a value of the tool. See [cli.md](cli.md#the-values-of-a-tool).                    |
| `[overrides.mise.<path>]` | It changes the generated mise config. See [mise.md](mise.md#the-overrides).                                                      |
| `[mise.*]`                | The own mise config of the user. It goes to `conf.d/zz-macos/mise.toml`. See [mise.md](mise.md#the-own-config-of-a-values-file). |

These stop the sync:

- Any other top-level key.
- `tools = [...]`, the old format.
- `env` in a `[tools.<name>]` table: the Mac side has no environments.
- A `format` other than `2`.
- `[overrides.mise_env]`, and any other key of `[overrides]`.
- A `[tools.<name>]` that is not a table, a `[mise]` that is not a table, a
  `bootstrap_commit` that is not a string, and a `backup` that is not a string
  or is empty.
- An old form: run `bootstrap mac migrate`. The old forms are a file without
  `format = 2`, a `[packs.<name>]` table or a `packs` key, and the table of
  the old Mac entry tool, whose values go to `[tools.mise]`, and the answer
  `config_path` of `[tools.mise]`, which goes to `backup`.
  `docs/sync/migrate.md` names the old tool. `mac migrate` migrates them as on
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
| `.env`     | `HOME` and `USER`.                                        |
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
| `macos/tools/claude/files/.claude/settings.json.merge.toml`         | The quoted paths in `MISE_TRUSTED_CONFIG_PATHS`. |

## The run order

1. The front-end examines its flags, mise and the clone, and finds the place of
   `macos.toml`.
2. It reads `macos.toml` and each `tools.toml`, and it resolves the selection.
3. It merges the values of the tools, with the derive scripts. A question of a
   selected tool without an answer stops the run with code 2. An optional
   question can have no answer. A question of a group needs no answer, but its
   group needs one answer that is not empty, else the run stops with code 2.
   The message names the question or the group, and the hint is
   `run bootstrap mac migrate`.
4. It runs the check scripts.
5. It examines `[overrides.mise*]`, and generates the mise config of each tool
   and of `macos.toml` into a work folder.
6. It makes the file plan: each file of a selected tool, relative to `$HOME`
   (or to the backup clone with `root = "backup"`), with the `.config/mise/`
   files first and the generated mise files first of
   all.
7. It makes the links of the backup clone (see
   [The backup clone](#the-backup-clone)), then writes each file of the plan in
   one pass, and applies the stale rule.
8. It writes the manifest, runs the `after` scripts, and sets
   `bootstrap_commit`. It runs no `mise bootstrap`.

The write of one file is in
[cli.md](cli.md#the-write-of-one-target). A file takes this
order: the real path of its target, the seed gate, `requires`, the unblock of a
folder path, the `dir_perm` of its folder, then the render or the copy, the
parse check, the merge and the compare.

`requires` holds paths relative to `$HOME`, or to the backup clone for a
`root = "backup"` entry. An owned file with an entry that is
not there is not handled, so the stale rule removes a listed copy with the
header. A seed and a shared file stay as they are.

The `ssh` tool gives each file below `.ssh/` `dir_perm = "700"`, so `~/.ssh` has
mode 700. The check uses the path in the home folder, not the real path of a
linked file.

## The first setup and exit 4

The first setup installs what `mise bootstrap` and the entry tasks need on a
new Mac: Homebrew, the shells, mise, `uv` and Python. A Mac needs it when no
sync finished on it: the cache has no manifest `macos-generated.tsv`. The
manifest is per Mac, so a second Mac with a clone of the backup also needs
it.

`bootstrap mac migrate` runs no sync. It writes `macos.toml`, and on a Mac with
no manifest it runs the first setup in place of the sync:

1. It renders the tools of the init stage. The init stage holds each selected
   tool with `stage = "init"` and, transitively, each tool in their
   `depends`. `bash`, `zsh` and `fish` have `stage = "init"` and
   `depends = ["homebrew", "mise"]`. `homebrew` depends on `fnox`, and `mise`
   depends on `cli`, `python` and `uv`.
2. A failure of a file ends the run with no manifest. `Apply` puts back
   `macos.toml`.
3. It writes the manifest with the rows of the init stage.
4. It runs `mise bootstrap --yes` from `$HOME`, unless `--no-bootstrap` is set.
   A failure of this step ends the run with code 1.
5. It runs the `after` script of each init-stage tool, prints the reasons of a
   script that asks for a restart, and prints
   `the first setup is done: open a new terminal, then run: bootstrap mac sync`.
   It stops with **exit 4**, because the login shell and `PATH` changed.

With `--no-bootstrap`, the first setup ends after step 3 with exit 0. The first
setup sets no `bootstrap_commit`. On a Mac with a manifest, `mac migrate`
writes only `macos.toml`. After each `mac migrate` that ends with exit 0, it
prints `Next: run bootstrap mac sync`.

`bootstrap mac sync` writes each selected tool in one pass, and never runs
`mise bootstrap`. On a Mac with no manifest, it first prints the warning
`no sync finished on this Mac yet: run bootstrap mac migrate first`, then it
writes the files. When `mac sync` finds a pending migrate, it runs that migrate
first: on a Mac with no manifest, that is the first setup, and the sync stops
there with its exit 4.

**Exit 4** is not a failure of a file: the rest needs a new terminal. A sync
also stops with exit 4 when an `after` script asks for a restart, before it
sets `bootstrap_commit`. Open a new terminal, then run the command again.
`updateall` also stops with exit 4, and it names `mise run updateall`.

A dry run runs no `mise bootstrap` and no `after` script, and never stops with
exit 4.

The other exit codes are in [cli.md](cli.md#the-exit-codes).

## The output

The list, the last line and verbose are in
[cli.md](cli.md#the-output-of-a-run). These rules are only for
the Mac:

- The first line is `↻ bootstrap <short commit> · <n> tools`. With an empty
  commit, it reads `↻ bootstrap an unknown commit · <n> tools`.
- The `mise bootstrap` of a first setup prints `▸ Running mise bootstrap --yes`
  before it and `✔ mise bootstrap done` after a success. The output of mise
  stays as it is.
- The run has one last line. A return of exit 4 has no last line.
- `FAILED: <list>` stays by default, before the last line.
- `log` of the entry task is an info line, and `log_error` is an error line.
  With the style rule on, they show the glyphs `·` and `✗`. Otherwise they have
  the prefixes `[INFO] ` and `[ERROR] `.

## The failures

A failed file stops no other file. After a failure:

- The stale rule still runs. A target that failed but is still
  planned stays handled, so its existing copy stays. A tool whose mise config
  cannot generate keeps its existing `conf.d/<tool>/mise.toml` in the same way.
- The manifest keeps its old content.
- `bootstrap_commit` stays. A first setup with a failure runs no
  `mise bootstrap`.

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

## The backup clone

The key `backup` of `macos.toml` names a git clone of the user, for example the
private `macos-setup` clone. The clone holds the backup of the synced files
below `<backup>/dotfiles/`, with the same paths as in `$HOME`. The sync links
the targets into the clone, and then writes through the links (see
[Symlinks](#symlinks)). bootstrap never commits: the user examines the clone
with git and commits the changes. Without `backup`, the sync writes real files
and makes no link.

A `[files]` entry with `root = "backup"` (see
[tools.md](tools.md#the-files-table)) writes its target as a real file at
`<backup>/<target>`, at the root of the clone, not below `dotfiles/`. The link
rule ignores it: it gets no link from `$HOME`. A dry run and `mac check` show
its path in the clone: `~/...` when the clone is below `$HOME`, else the
absolute path. Without `backup`, the sync skips the entry as a skipped target,
with the line `skip file  @backup/<target> (no backup in macos.toml)`.
bootstrap never commits these files: the user commits them in the clone, as
the changes below `dotfiles/`.

The link rule picks the level of each link:

1. Link a folder when no selected tool declares a runtime file in it.
2. Else do not link the folder. Link each child that holds a synced file: a
   child folder with the same rule, and a child file as a file.
3. `~` and `~/.config` are never linked, so their children are linked one by
   one. `~/.config/bootstrap/` is a folder link, so `macos.toml` is in the
   backup.

A runtime file is a file that the program of a tool writes, which must not go
into git, for example `.config/fish/fish_variables`. A runtime path can also be
a folder where other programs write their own items, for example
`.claude/skills` of the `claude` tool: the folder is never linked, so the sync
links each synced item in it one by one, and each other item stays a real item
in `$HOME`. The key `runtime` of the
`tools.toml` of a Mac tool names them: a list of paths relative to `$HOME`,
with no `~`, no `..` and no `/` at the start or the end.

The engine makes the links itself, also the link `~/.config/mise`. It writes
no `[mise.dotfiles]` entry for them: mise reads its entries from
`~/.config/mise`, and it refuses an entry whose source is missing. The own
`[mise.dotfiles]` entries of the user in `macos.toml` stay with mise.

Each run with `backup` (`mac sync`, and the first setup of `mac migrate`)
does these steps before it writes a file:

1. It stops with code 2 when the clone is not a folder.
2. It makes the list of links from the targets of the plan, the generated
   mise files and `macos.toml` (when its folder is below `$HOME`).
3. It examines each link. The run stops with code 2 before any change when one
   link has a conflict, and each conflict has one error line with the fix
   `fix it by hand, then run the sync again`.
4. It does the steps of the safe cases, then makes each missing link. A folder
   link with no folder in the clone gets an empty folder there.

| State of the target in `$HOME`                                                      | Action                                                    |
| ----------------------------------------------------------------------------------- | --------------------------------------------------------- |
| The link to its place in the clone                                                  | none                                                      |
| Nothing                                                                             | the link                                                  |
| A real file or folder, and the clone has nothing there                              | move it into the clone, then the link                     |
| A real file with the same bytes as the clone                                        | remove the real file, then the link                       |
| A link that resolves to its place in the clone (for example a link of an old level) | remove that link, then the link                           |
| A real folder, and the clone has a folder there                                     | the same rules for each item of the folder, then the link |
| A real file that differs from the clone                                             | conflict                                                  |
| A link to a different place                                                         | conflict                                                  |
| A file and a folder at the same place                                               | conflict                                                  |
| A folder above the link is a link (the rule now splits that folder)                 | conflict: replace the link with a folder                  |

A dry run examines the links, stops on a conflict, and shows
`would move <path> into the backup <clone path>` and
`would link <path> to <clone path>`. It changes nothing, so the render of the
dry run compares the targets at their place of today. A real run shows
`moved <path> into the backup` and `linked <path> to <clone path>`.

A link stays when its tool is no longer selected: the files in the clone are
the user's backup. bootstrap never commits and never runs git in the clone.

Keep the main checkout of the backup clone on its default branch. Do not
create a branch in it and do not switch it. `$HOME` links into this checkout,
so a `git switch` changes the live files of the Mac at once, for example
`~/.config/mise`, `~/.config/git` and `~/.ssh/config.identity`. A branch with
an old or a broken file can then stop git and ssh. Make each change in a git
worktree with its own branch (`git worktree add`), commit it there, land it
with `git:merge:<branch>` in the worktree (the task merges in the main
checkout), then remove the worktree.

## The manifest

The manifest is `macos-generated.tsv` of the cache. It is outside the clone, so
`git clean` does not remove it. Its shape and the stale rule are in
[cli.md](cli.md#the-manifest-and-the-stale-rule).

Each owned target of a selected tool, each generated mise file, and each item
that bootstrap added to a shared file has one row. The row of a target in the
backup clone (`root = "backup"`) has the path `@backup/<target>`. The file
`conf.d/zz-macos/mise.toml` has the tool `zz-macos`. A clean run writes the
manifest after the stale rule. The first setup writes the rows of the init
stage before its `mise bootstrap`. A
dry run prints the stale files and writes no manifest. The stale rule examines
the rows of each tool, also of a tool that is not selected.

## Where the Mac side writes

- `$HOME`, and the sources that its symlinks point to.
- The backup clone: the moves and the empty folders of the links, below
  `<backup>/dotfiles/`, and the targets of the `root = "backup"` entries, at
  the root of the clone.
- The cache: the clone, the lock and `macos-generated.tsv`.
- `macos.toml`: `bootstrap_commit`, the tables of the required members of
  `base`, and the writes of `mac migrate`, `mac add`, `mac remove`,
  `mac replace` and `mac detect --add`.
- Temp folders, which it removes on exit.

`mise bootstrap` writes where mise writes.
