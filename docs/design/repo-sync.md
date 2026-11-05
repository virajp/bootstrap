# `repo-sync` design

This document describes the repo front-end of the sync, `sync-engine/repo-sync`.
It
writes the files of the selected tools of `repo/tools/` into a consumer
repository.
It holds the rules that only a repository has: members and submodules, the
branch
model, the pre-commit render, `--check` and its gate.

See also: [sync-engine.md](sync-engine.md) for the shared rules and the exit
codes,
[bootstrap-config.md](bootstrap-config.md) for the keys of the values file,
[packs.md](packs.md), [tools.md](tools.md), [mise.md](mise.md) and
[fnox.md](fnox.md).

## The entry tasks

The `repo-sync` tool of `repo/tools/` ships the three tasks.

| Task                  | Job                                                                                          |
| --------------------- | -------------------------------------------------------------------------------------------- |
| `setup:configs`       | Refreshes the cache clone and copies it. Then it runs `setup:configs:sync` in a new process. |
| `setup:configs:sync`  | Hidden. It loads `sync-engine/repo-sync` from the copy and runs `repo_sync`.                 |
| `code:lint:repo-sync` | Hidden. The gate of the values file. See [The gate](#the-gate-codelintrepo-sync).            |

`setup:configs` (`setup/configs/_default`) embeds `sync-engine/clone`. See
[sync-engine.md](sync-engine.md#how-an-entry-task-embeds-clone). It sets
`SYNC_ENGINE_CLONE` to the copy. It also sets `REPO_SYNC_CLONE` for a member
that
still has the old sync task. `setup:configs:sync` stops without
`SYNC_ENGINE_CLONE`.

`--all` and `--<name>` also sync each selected submodule from the same copy, so
all
get the same commit. The first failure stops the task. A member without a
`setup:configs:sync` task stops the task: without that task, mise finds the task
of
the root and runs it in the root.

`repo_sync` makes its own private copy of the clone, because it writes the
generated
mise config into the copy. It never clones and reads no URL.

| Variable                  | Use                                                            |
| ------------------------- | -------------------------------------------------------------- |
| `REPO_SYNC_VALUES`        | The values file. The default is `.config/bootstrap.toml`.      |
| `REPO_SYNC_NO_AUTOUPDATE` | `1` skips `pre-commit autoupdate`, so no network is necessary. |

## The values file and the selection

`repo_sync` reads the consumer file `.config/bootstrap.toml`. The keys are in
[bootstrap-config.md](bootstrap-config.md).

- `repo_sync` adds the pack `base`. Its name is fixed in `repo-sync`. See
  [packs.md](packs.md).
- It expands each pack to its default members, adds the `depends` of each tool
  and the
  tool of each mise `depends`, depth first, and copies each file of each tool. A
  dependency comes before the tool that needs it.
- The values file is required. A missing file stops the sync. The old name
  `.config/repo-config.toml` is read, with a message: rename it.
- `repo_sync` copies the clone of `SYNC_ENGINE_CLONE`. A clone that it cannot
  use stops
  the run.
- `repo_sync` examines the manifest of each tool of the clone, selected or not.
  See
  [tools.md](tools.md).

The exit codes are in [sync-engine.md](sync-engine.md#the-exit-codes).

## Members and submodules

A member is a git submodule or a folder below the root. The keys are in
[`[member.<name>]`](bootstrap-config.md#membername).

- The member flags (`--all`, `--<name>`) of `setup`, `setup:configs`,
  `code:format` and
  `code:lint` are only for the submodules. A folder member is part of the root
  tree.
- A task that runs in a member uses its folder.
- A submodule is a separate repository, with its own values file and its own
  manifest.

### `root_only`

A `root_only` tool is not for a git submodule. In a submodule, the sync skips it
when a
pack adds it, and stops when a `[tools.*]` entry names it.

## The template context

The context of a template is the values file, with these values merged over it.
A key
of the same name in the values file does not replace them. The reserved names
are in
[bootstrap-config.md](bootstrap-config.md#reserved-names).

| Value             | Content                                                                                                                                                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `.members`        | The list of member names: the git submodules (the key of the `[member.<name>]` table whose `folder` is the submodule path, else the `repo_name` of the values file of the submodule, else the path with each `/` changed to `_`), then each other `[member.<name>]` key. |
| `.member_folders` | A map from each name to its folder (`folder`, default the name).                                                                                                                                                                                                         |
| `.member_types`   | A map from each name to `submodule` or `folder` (`type`, default from the folder).                                                                                                                                                                                       |
| `.submodules`     | The list of the git submodules only, nested ones included. It is empty in a repository without submodules.                                                                                                                                                               |
| `.tools`          | The list of the selected tools, with the packs expanded and the dependencies added.                                                                                                                                                                                      |
| `.extensions`     | The sorted list of their VS Code extensions and of the values key `extensions`, without duplicates. In the root repository, it also has those of each git submodule.                                                                                                     |
| `.precommit`      | The list of their `[[precommit]]` tables, in tool order, then the entries of the consumer file `.config/repo-sync/pre-commit.yaml`, with the `--branch` pairs filtered. See [The pre-commit config](#the-pre-commit-config).                                             |
| `.sync_engine`    | The absolute path of the `sync-engine/` folder of the clone copy of the run.                                                                                                                                                                                             |
| `.config`, `.all` | The values of the tools. See [sync-engine.md](sync-engine.md#templates).                                                                                                                                                                                                 |

A name has only letters, digits, `-` and `_`, and two members never have the
same name.
If not, the sync stops.

A template of `repo/tools/` does not encode a string value with `data.ToJSON`
yet, for
example `repo/tools/mise/.config/mise.toml.tmpl`. The rule for a Mac template is
in
[setup-sync.md](setup-sync.md#the-template-context).

## The branch model and the merge tasks

The branches `main` and `develop` select the model. A branch counts if it exists
locally
or on `origin`. `repo_sync` writes the merge tasks of the model that it finds,
and
removes the others. It stops if neither branch exists.

| Branches             | Tasks                                                                                                                                                                  |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `main` and `develop` | Work starts on feature branches. `code:merge:develop <branch>` lands a branch on `develop`. `code:merge:main` lands `develop` on `main`, and runs only from `develop`. |
| Only one of them     | `code:merge <branch>` (`code/merge/_default`) lands any feature branch on that branch.                                                                                 |

- Merges use `--no-ff`. A hook pass over all files runs first as a safety net,
  with
  `SKIP=no-commit-to-branch`: that hook guards a commit, and the safety net
  makes none.
  Thus `code:merge:main` also runs from a protected `develop`.
- The merge tasks read `[branch.<destination>]` of the repository's own
  `.config/bootstrap.toml` when they run, with `yq`: the root and each submodule
  read
  their own file.
- `merge_model` selects `direct` (local merge and push, the default) or `pr`
  (push and
  open a pull request with `gh` or `glab`). Any other value reads as `direct`.
  Without an
  `origin` remote, `direct` merges locally and skips the pull and the push, and
  `pr` stops
  before the hook pass. The `github` and `gitlab` tools install these CLIs.
- `merge_from` lists the only source branches, and `merge_never` the forbidden
  ones. A
  table with both stops the merge. Without both, the defaults are: into `main`
  with
  `develop`, only from `develop`; into another branch, never from `main`.

### `direct_commit`

`direct_commit = true` in `[branch.main]` or `[branch.develop]` allows a direct
commit to
that branch. `repo_sync` reads it, and the merge tasks do not.

The `no-commit-to-branch` hook of the `git` tool protects only the core branches
that
exist and do not allow direct commits. The sync removes the `"--branch",
"<branch>"` pair
of each other core branch, and the hook when no pair is left.

| Core branches of the repository | `direct_commit = true` in | Hook args                         |
| ------------------------------- | ------------------------- | --------------------------------- |
| `main` only                     | `[branch.main]`           | No hook                           |
| `develop` only                  | `[branch.develop]`        | No hook                           |
| `main` and `develop`            | `[branch.develop]` only   | `--branch main`                   |
| `main` and `develop`            | Both tables               | No hook                           |
| `main` and `develop`            | `[branch.main]` only      | The sync stops                    |
| `main` and `develop`            | No table                  | `--branch main --branch develop`  |
| `main` only, or `develop` only  | No table                  | `--branch <branch>` of the branch |

`repo_sync` validates the key before it writes a file, also when it is `false`.
These
values stop the sync:

- A value that is not a boolean: `direct_commit must be true or false`.
- The key in a `[branch.*]` table other than `[branch.main]` and
  `[branch.develop]`:
  `direct_commit is a key of [branch.main] and [branch.develop] only`.
- The key for a core branch that does not exist, locally or on `origin`. The
  message names
  the table and the branch: create the branch, or remove the key.
- With `main` and `develop`, `direct_commit = true` for `main` but not for
  `develop`. A
  repository with both branches never allows a direct commit to `main` and
  refuses it on
  `develop`.

## The pre-commit config

The `pre-commit` tool renders `.config/pre-commit-config.yaml` from
`.precommit`: the
`[[precommit]]` tables of the selected tools, then the entries of the consumer
file
`.config/repo-sync/pre-commit.yaml`.

Before the render, the sync removes each `"--branch", "<b>"` pair of a
`no-commit-to-branch` hook whose core branch does not exist or has
`direct_commit = true`.
This also applies to a hook in the consumer file. Only that pair form changes: a
hook
without args, or with `--branch=<b>` or `-b <b>`, does not change. A hook that
had
`--branch` pairs and has none left goes, also when it has other args, because
without
`--branch` the hook protects `main` and `master`. An entry with no hook left
goes too.

The sync runs `pre-commit autoupdate` on the rendered config, before the parse
check, so
each `rev` becomes the newest tag. A later `autoupdate` in the consumer picks
the same
tags, so the two never undo each other. Keep a real tag in each `rev` of a
`tools.toml`: an
older `repo-sync` copies it as it is. `REPO_SYNC_NO_AUTOUPDATE=1` skips
`autoupdate`. If
`pre-commit` is missing or `autoupdate` fails, the sync keeps the existing file.
The task
`setup:configs:sync` provides `pre-commit` with `#MISE
tools.pre-commit="latest"`.

## `[overrides.gitignore]`

The sync writes the shared `.gitignore` with `[overrides.gitignore]` of the
values file
applied. It comments out each line of `remove`, then adds the lines of `add` at
the end. A
later line wins, so `add` can also negate a shared pattern with `!`. This step
comes after
the format step of a target.

These stop the sync: a line of `remove` that the shared `.gitignore` does not
have, a line
in both lists, an empty line, a key other than `add` and `remove`, the old
`[overrides.graphifyignore]` and the old consumer file
`.config/repo-sync/gitignore`.

`[overrides.mise*]` is in [mise.md](mise.md#the-overrides).

## fnox profiles

A repository does not use the fnox `default` profile, which is only for the
user-level
fnox settings. When `.config/fnox.toml` exists, the sync and `repo_sync --check`
stop on a
top-level `[secrets]` table (the `default` profile, also when empty), a
`[profiles.default]` table, or no `[profiles.<name>]` table. A repository with
one profile
uses its `repo_name`. Each `fnox` call passes `-P <name>`. The rest of the
design is in
[fnox.md](fnox.md).

## dprint

The sync formats each rendered template, each generated mise file and each merge
target
with the dprint config of the consumer before it compares the file. Without this
step, the
format hook changes the render, and the next sync writes it back. A plain file
needs no
step, because dprint formats its source in `bootstrap`.

The sync copies the files to a stage folder, at their paths in the repository,
with a copy
of the dprint config, and it formats the stage in one dprint call. Without
dprint, or if
dprint fails for a file, the raw render of that file stays.

## `requires_found`

The `[files]` key `requires_found` and the same key of a mise tool hold a list
of file
names. The sync must find each name at the root or down to 4 folders below it,
in a file
that git tracks or does not ignore. The sync does not look in git submodules,
because a
submodule has its own sync. A name has no `/`. A target with `requires` and
`requires_found` needs each entry of both.

## The flags

### `--dry-run`

`repo_sync --dry-run` shows each change and writes no file. It examines each
target too.

### `--check`

`repo_sync --check` examines only the local rules of the values file: the rules
that need
the values file and the local git repository, for example `env`, `environments`,
`[member.*]`, the structure of `[overrides]`, `direct_commit`, a missing `main`
and
`develop`, and the profiles of `.config/fnox.toml`. It runs after the checks of
the values
file and of `mise`, and before the clone. It makes no clone, reads no
`SYNC_ENGINE_URL` or
`SYNC_ENGINE_CLONE`, writes no file and does not set `bootstrap_commit`. It runs
each local
check and prints each message, so each error shows at once.

The rules that need the clone stay with the sync:

- A tool or a pack that is not in `bootstrap`.
- A pack loop, a pack named like a tool, or another bad `[pack.*]` table (an old
  key or a
  bad member).
- An environment that a mise tool does not allow.
- An `[overrides.mise*]` path that no selected tool has.
- A `remove` line of `[overrides.gitignore]` that the shared file does not have.
- A `root_only` tool in a submodule.
- An old key or a bad `[files]` table of a `tools.toml`.
- A bad `[defaults]`, `[questions]` or `[contributes]` table, or a conflict of
  the values of
  the tools.

`--check` with `--dry-run` is a usage error.

### The gate `code:lint:repo-sync`

The gate runs `repo_sync --check`. The `lint` hook runs `code:lint`, which runs
the gate
with the staged files, so the tool adds no hook. The gate is hidden, and it
ignores `--fix`.

It prints nothing and exits 0 in these cases:

- Without gomplate (`mise x -- which gomplate`).
- With a list of files that does not include `.config/bootstrap.toml`, or with
  no list and
  no values file.
- The cache has no clone with `sync-engine/repo-sync`.
- The raw `origin` of `<cache>/repo` is not one of five literal forms of the
  default URL.
  Thus the gate never sources code from another repository at commit time.

Else it unsets `SYNC_ENGINE_CLONE` and `REPO_SYNC_CLONE`, and loads
`sync-engine/repo-sync`
of the cache. It takes no lock and never clones.

## The manifest

The manifest is `.config/repo-generated.tsv`. It is committed, and each git
submodule has
its own. Its shape and the stale rule are in
[sync-engine.md](sync-engine.md#the-manifest-and-the-stale-rule).

After a clean run, the order is the stale rule, the write of the manifest, then
`bootstrap_commit`. A run with a failure runs no stale rule and writes no
manifest. The sync
reads the manifest after the file loop. A manifest that it cannot read fails the
run after
its writes and its removals of the merge tasks: it logs `FAILED:
.config/repo-generated.tsv`,
and it skips the stale rule, the write of the manifest and `bootstrap_commit`.

Handled paths are the files of the selected tools that this run wrote or kept
(generated,
seed or merge), and the merge tasks of the other branch model. A seed or a merge
target of a
selected tool that the run left in place counts as kept, also one that
`requires` or its
environment skips. A generated file that `requires` or its environment skips is
not handled.
A folder member without its own `.git` is part of the root tree.

### What the sync deletes

The sync never deletes a file, except these:

- The merge tasks of the other branch model.
- A synced file at a path where the sync needs a folder.
- The stale rule: a file with the bootstrap header that the manifest lists and
  this run does
  not handle.

The merge tasks and the stale rule skip only a seed and a merge target of a tool
of this run
(a selected tool or a dependency, also through a mise `depends`). At a folder
path, the sync
refuses to remove such a path and fails that file. Thus the stale rule removes a
listed path
with the header at a seed or a merge path of a tool that is not in the run. A
merge target has
no header, so the rule leaves it. `docs/sync/migrate.md` tells consumers what to
delete.

## `[claude.plugins.*]`

The sync does not read `[claude.plugins.user]` and `[claude.plugins.project]` of
the values
file. `setup:ai` of the `claude` tool reads them when it runs, as the merge
tasks read
`[branch.*]`. Each key is `"<plugin>@<marketplace>" = "<source>"`. It adds a
missing
marketplace from the source, installs each plugin that is missing at that scope,
and updates
the marketplaces and the plugins.
