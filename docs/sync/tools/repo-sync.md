# The `repo-sync` tool

The `setup:configs` tasks that copy the tools from `bootstrap`, with the sync
function `repo_sync` of `sync-engine/repo-sync` in the clone.

| Property     | Value |
| ------------ | ----- |
| Root only    | No    |
| Depends      | None  |
| Default envs | dev   |
| Seeds        | None  |

## When the tool is relevant

Always. The `base` pack always adds this tool. Do not select it in the values
file. A `[tools.repo-sync]` table only gives values.

## What the tool ships

| Path in the repository                        | Kind      | Use                                                                                                                                                                                                    |
| --------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `.config/mise/tasks/setup/configs/_default`   | Template  | The task `setup:configs`. It fetches `bootstrap` into the cache. It updates the local sync task from the clone, then runs it. The sync embeds `sync-engine/clone` in it.                               |
| `.config/mise/tasks/setup/configs/sync`       | Synced    | The task `setup:configs:sync`. It loads `sync-engine/repo-sync` from the clone, and stops with code 2 without `SYNC_ENGINE_CLONE`. `setup:configs` updates this file from the clone before it runs it. |
| `.config/mise/tasks/code/lint/repo-sync`      | Synced    | The gate `code:lint:repo-sync`. It runs `repo_sync --check`.                                                                                                                                           |
| `.config/mise/conf.d/repo-sync/mise.dev.toml` | Generated | The mise config of this tool.                                                                                                                                                                          |

The sync no longer ships the forwarding file. It stays in `bootstrap` at
`tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`, because an old
`setup/configs/sync` loads it from the clone. Nothing in a repository sources
the synced copy `.config/mise/tasks/_scripts/repo-sync`. With a manifest
`.config/repo-generated.tsv` that lists the copy, the stale rule removes it.
Without such a manifest, the copy stays until the cleanup of the forwarding
file, whose `migrate.md` item covers it.

## mise tools

| mise tool  | Version  | Environments | Note                               |
| ---------- | -------- | ------------ | ---------------------------------- |
| `gomplate` | `latest` | dev          | For the gate `code:lint:repo-sync` |

The sync does not use this entry. It starts gomplate, jq and yq itself through
mise, from `/`.

## Tasks

| Task                  | Use                                                                                                                                                                                                         |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup:configs`       | Update the local sync task and sync the shared files from `bootstrap`. `--dry-run` shows the diff. `--verbose` lists each file, also an unchanged one. `--all` and `--<name>` also sync the git submodules. |
| `setup:configs:sync`  | Run `repo_sync` from the clone of `setup:configs`. Do not run it alone.                                                                                                                                     |
| `code:lint:repo-sync` | Validate `.config/bootstrap.toml` with `repo_sync --check`. Hidden.                                                                                                                                         |

`setup:configs` keeps a shallow clone of `bootstrap` in `repo/` of the cache
`${XDG_CACHE_HOME:-~/.cache}/sync-engine/`, for each consumer of the machine:

- Each run fetches the branch `main` of
  `https://github.com/virajp/bootstrap.git` into the clone, and gives a copy
  of it to `setup:configs:sync`. The fetch uses the URL, not `origin`. It needs
  no SSH key.
- The lock folder `lock/` of the cache makes a second run wait until the first
  has made its copy. A waiter removes the lock of a dead holder.
- The origin compare is normalized, so the SSH and the HTTPS URL of one
  repository match. A clone with another `origin`, for example after a test
  with `SYNC_ENGINE_URL`, is cloned again.
- A `repo/` that is not a git clone and not empty stops `setup:configs` with
  code 2. The task does not delete it.
- A git repository at the cache root is the old layout. When its raw `origin`
  normalizes to the URL, the run removes it and keeps `repo/`, the lock and
  `macos-generated.tsv`. Another `origin` or no `origin` stops `setup:configs`
  with code 2, and the task deletes nothing in the cache root.
- Each run removes the old cache `bootstrap/` and `bootstrap.lock` of
  `${XDG_CACHE_HOME:-~/.cache}` while it holds the old lock. When a live old
  stub holds that lock, a later run does the removal. A failed removal does
  not fail the sync. The line `Removed the old cache <path>` prints only when
  the run removed the old folder.
- If the fetch or the clone fails, `setup:configs` stops with code 5. It
  never uses an old clone.

Before each run of `setup:configs:sync`, for the root and for each submodule,
`setup:configs` updates the local sync task. A running task keeps its old code
when the sync writes its file, so the task compares the local
`setup/configs/sync` and `_scripts/helpers` with the same files in the copy of
the clone:

- It copies a file that differs, in its content or in its mode. It prints
  `Updated the local sync task to <short commit>`, below the line
  `Updating the clone`. If no file differs, it prints
  `The local sync task is up to date`.
- A dry run copies nothing. If a file differs, it prints
  `Would update the local sync task to <short commit>`. The sync then shows the
  file as `would change`.
- It skips a file that does not exist locally, a file that the clone does not
  have and a local symlink. If neither file exists locally, it prints no line.
- A failed copy prints a warning, and the run continues with the local file.
- A copied file counts as unchanged in the sync, so the last line can read
  `Synced: no file changed`.
- `_default` stays old for that run, because it needs the full render. The sync
  writes it for the next run.

The sync lists only what changed, then prints a table of counts and a last line:

```text
  ✎ changed   .gitignore
  ✚ new       .config/mise/tasks/code/lint/repo-sync
  ❋ seeded    .config/swiftlint.yml
  ✖ removed   .config/mise/conf.d/old-tool/mise.dev.toml

  ┌───────────────┬───────┐
  │ ✎ Changed     │     1 │
  │ ✚ New         │     1 │
  │ ❋ Seeded      │     1 │
  │ ✖ Removed     │     1 │
  │ = Unchanged   │   143 │
  │ ◦ Kept seeds  │     6 │
  │ ⊘ Skipped     │    11 │
  │ ✗ Failed      │     0 │
  ├───────────────┼───────┤
  │ Total         │   164 │
  └───────────────┴───────┘
✅ Synced: 4 files changed
```

- A failed file also has a list line, `✗ failed <path>: <reason>`. The line
  `FAILED: <list>` prints before the table, and the last line gives the exit
  code.
- A dry run has the words `would change`, `would add`, `would seed` and
  `would remove` in the list, with the diff of a change. The table counts what
  the run would do.
- A run that stops before it writes files has no table.
- `NO_COLOR`, an output that is not a terminal, or `TERM=dumb` removes the
  glyphs and the colors. `FORCE_COLOR` turns them on.
- `--verbose`, or `SYNC_ENGINE_VERBOSE=1`, prints the log line of each file in
  place of the list: `unchanged`, `updated`, `kept seed`, `skip seed`,
  `skip env` and the others. The table and the last line stay.

A repository with an old `_default` must run `mise run setup:configs` once
without `--verbose`, because mise refuses a flag that the old task does not
define. That run writes the new `_default`.

The member flags of `setup:configs` are the same as the flags of `setup`.
They are only for the git submodules, nested submodules included. A folder
member is part of the root tree, so it gets no flag.

- `--all` syncs the root, then each submodule. `--<name>` syncs the root,
  then the submodule `<name>`. Without a flag, only the root syncs.
- The task fetches one time. The root and each submodule sync from the same
  copy, so all get the same `bootstrap_commit`. `--dry-run` and `--verbose`
  apply to each. Each one prints its own list and table.
- Before each submodule, the task prints a header with the `label` of the
  member, for example "Syncing Backend ...".
- The first failure stops the task with its code. The submodules after it
  get no sync.
- A selected submodule without its own `setup/configs/sync` stops the task
  with code 2. Without that file, mise runs the task of the root in the root.

`code:lint:repo-sync` is a gate of `code:lint`. It accepts `--debug` and
`--fix`, and ignores them, because the check has no fixes. It prints nothing
and exits 0 in these conditions:

- gomplate is not available (`mise x -- which gomplate`). The gate does not
  install it, so a commit never needs the network.
- A list of files does not include `.config/bootstrap.toml`.
- No list of files is given, and `.config/bootstrap.toml` does not exist.
- The cache has no clone with `sync-engine/repo-sync`. The gate never clones.
- The raw `origin` of `<cache>/repo` is not one of five literal forms of the
  default URL: the HTTPS URL and the `git@github.com:` URL, each with and
  without `.git`, and the `ssh://git@github.com/` URL with `.git`. The gate
  sources no code from another repository at commit time.

In the last two conditions, the gate does not validate the values file. Thus
no output does not show that the values file is correct. A run of
`setup:configs` makes the clone.

Else it unsets `SYNC_ENGINE_CLONE` and `REPO_SYNC_CLONE`, loads
`sync-engine/repo-sync` of the cache without its lock, and runs
`repo_sync --check`.

`repo_sync --check` examines only the rules that need the values file and the
local git repository. It runs after the checks of the values file and of
`mise`, and it stops before the clone. It makes no clone, writes no file and
does not set `bootstrap_commit`. It runs each local check, prints each
message, and returns the highest code. `--check` with `--dry-run` is a usage
error, code 2. These rules need the clone, so only `setup:configs` examines
them: a tool or a pack that is not in `bootstrap`, a pack loop, a pack named
like a tool or another bad `[pack.*]` table (an old key or a bad member), an
environment that a mise tool does not allow, an
`[overrides.mise*]` path that no selected tool has, a `remove` line that the
shared file does not have, a `root_only` tool in a submodule, an old key or a
bad `[files]` table of a `tools.toml`, and a bad `[defaults]`, `[questions]` or
`[contributes]` table or a conflict of the values of the tools.

Besides the copy of the files, `repo_sync` does these steps:

- It stops with code 2 without `mise`. An installed gomplate, jq or yq runs as
  `MISE_OFFLINE=1 mise x <tool> -- <tool>`, with no network. A tool that is not
  installed runs as `mise x <tool>@latest -- <tool>`. Each call runs from `/`,
  so no mise config of the repository selects a version.
- It examines each JSON, YAML and TOML target after its render or copy, and
  before dprint: jq for JSON, yq for YAML and TOML. The `format` key of its
  `[files]` table selects the format, else the extension. A `.json` target
  without a `format` key, with a line that starts with `//` after spaces, is
  JSONC, and it does not examine that target. This JSONC rule does not apply to
  a merge target. A target that does not parse keeps its existing file, the
  other files continue, and the sync returns code 2 and does not set
  `bootstrap_commit`.
- It examines the `[files]` tables of each `tools.toml` of the clone. An old key
  or a bad table stops the sync with code 2. A target passes only when it, or
  the
  target with `.tmpl`, is the same string as a path of `find -type f` in the
  tool
  folder.
- It merges each `mode = "merge"` target with the live file, before dprint. A
  merge source or a live target that cannot merge, and a merge target that is a
  symlink or is below a symlinked folder, keep the existing file. The other
  files
  continue, and the sync returns code 2.
- It writes the merge tasks of the branch model that it finds, and removes the
  others. If neither `main` nor `develop` exists, it stops with code 2.
- After a sync without a failure, it runs the stale rule, then it writes the
  manifest `.config/repo-generated.tsv`, then it sets `bootstrap_commit`. A sync
  with a failure runs no stale rule and writes no manifest. A failed stale
  removal is a failure, and the other rows still run.
  - The stale rule removes each file with the bootstrap header that the old
    manifest lists and this run does not write, then its empty parent folders. A
    file without the header stays. It skips a seed and a merge target of a
    selected tool. The first sync has no manifest, so it removes nothing.
  - The manifest has one `<tool><TAB><path>` row for each generated file. The
    sync skips a row with no TAB, an empty tool or path, an absolute path, or an
    empty, `.`, `..` or `.git` segment. It does not remove a listed path that is
    a symlink or is below a symlinked folder, or a listed file of another
    repository.
  - A manifest that is a symlink, is below a symlinked folder or is not a
    regular file is not read. The sync returns code 2 after its other work, and
    it skips the stale rule, the write of the manifest and `bootstrap_commit`.
  - A dry run lists each stale file as `would remove`, and
    `.config/repo-generated.tsv` as `would change` or `would add` when the
    content differs. With `--verbose`, the lines are `would remove <path>` and
    `would write .config/repo-generated.tsv`.
- It merges the values of each selected tool (`sync-engine/values`): the
  `[defaults]` of the tool, the contributions of the selected tools, then
  `[tools.<name>]` without `env`. A template reads them as `.config` and `.all`.
  A failure stops the sync with code 2 or 3 before it writes a file of a tool.
  See `docs/design/tools.md`.
- It applies `[overrides]` of the values file to the generated mise config
  and to `.gitignore`. Each overridden line gets an `# override:` comment.
  A bad `[overrides]` stops the sync with code 2. See "Overrides" in step 3
  of `index.md`.
- It validates `direct_commit` of the `[branch.*]` tables before it writes a
  file. These values stop the sync with code 2: a value that is not a
  boolean, the key in a table other than `[branch.main]` and
  `[branch.develop]`, the key for a core branch that does not exist (also
  `false`), and `direct_commit = true` for `main` but not for `develop` when
  both branches exist.
- It changes the `no-commit-to-branch` hook in `.precommit`, the hooks of the
  tools and of `.config/repo-sync/pre-commit.yaml`. It removes each
  `"--branch", "<branch>"` pair of a core branch that does not exist or that
  has `direct_commit = true`. A hook with no pair left goes, and an entry with
  no hook left goes too. Thus a repository with one core branch gets only
  `--branch <its branch>`.
- After a sync without a failure, it sets the top-level `bootstrap_commit` of
  the values file to the full commit of its clone. It replaces only that
  line, or adds it after the header comments. It prints
  `Set bootstrap_commit to <short commit>`, and a dry run prints
  `Would set bootstrap_commit to <short commit>`. With `--verbose`, the lines
  have the full commit and the path of the file. See "A repository that already
  has `.config/bootstrap.toml`" of `index.md`.

These environment variables change the sync. `--check` does not read
`SYNC_ENGINE_URL` and `SYNC_ENGINE_CLONE`, because it makes no clone:

| Variable                  | Use                                                                                                                                    |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `SYNC_ENGINE_URL`         | The repository to clone. The default is `https://github.com/virajp/bootstrap.git`.                                                     |
| `SYNC_ENGINE_CACHE`       | The cache folder. The default is `sync-engine` in the base of the cache.                                                               |
| `XDG_CACHE_HOME`          | The base of the cache, and of the old cache that each run removes. The default is `~/.cache`.                                          |
| `SYNC_ENGINE_CLONE`       | The copy of the clone. `setup:configs` sets it. An old `setup/configs/sync` reads `REPO_SYNC_CLONE`, so `setup:configs` sets that too. |
| `SYNC_ENGINE_VERBOSE`     | `1` prints the log line of each file in place of the list. `--verbose` sets it. Another value does nothing.                            |
| `REPO_SYNC_VALUES`        | The values file. The default is `.config/bootstrap.toml`.                                                                              |
| `REPO_SYNC_NO_AUTOUPDATE` | `1` skips `pre-commit autoupdate`, so no network is necessary.                                                                         |

## Pre-commit hooks

None. The `lint` hook of the `linter` tool runs `code:lint`, and `code:lint`
runs `code:lint:repo-sync` with the staged files.

## Files the repository supplies

| File                                | Used by                                                                                          |
| ----------------------------------- | ------------------------------------------------------------------------------------------------ |
| `.config/bootstrap.toml`            | `repo_sync`. The values file of step 3 of `index.md`. `repo_sync` writes its `bootstrap_commit`. |
| `.config/repo-sync/pre-commit.yaml` | `.config/pre-commit-config.yaml`. Optional: its own hooks.                                       |
| `.config/repo-generated.tsv`        | `repo_sync` writes and reads it. Commit it. Do not edit it.                                      |

## Adapt to the repository

1. Copy the sync files by hand before the first sync (step 2 of `index.md`).

## Migration

- Remove `gomplate` from the repository-owned mise files.
- The general items M1, M15, M23, M27 and M29 of `migrate.md` are about the
  files of this tool.
