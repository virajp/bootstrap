# The repo side

This document describes the repo side of the sync: `bootstrap sync` and the
front-end `internal/repo` of the engine. The front-end writes the files of the
selected tools of `repo/tools/` into a consumer repository. This document holds
the rules that only a repository has: the entry tasks, members and submodules,
the template context, the branch model, the pre-commit config, `check` and its
gate.

See also: [sync-engine.md](sync-engine.md) for the command line, the shared
rules and the exit codes, [bootstrap-config.md](bootstrap-config.md) for the
keys of the values file, [packs.md](packs.md), [tools.md](tools.md),
[mise.md](mise.md) and [fnox.md](fnox.md).

## The entry tasks

The `mise` tool of `repo/tools/` ships the entry of the sync, so each
repository has it. The tool also has the mise tool `github:virajp/bootstrap`
in `[mise_env.dev.tools]`.

| Task                     | Job                                                                                  |
| ------------------------ | ------------------------------------------------------------------------------------ |
| `setup:configs`          | Runs `bootstrap sync` in the repository, and in each selected submodule.             |
| `code:lint:60-bootstrap` | Hidden. The gate of the values file. See [The gate](#the-gate-codelint60-bootstrap). |

`setup:configs` (`setup/configs/_default`) has
`#MISE tools."github:virajp/bootstrap"="latest"`, so it runs also before the
first sync. With `BOOTSTRAP_BIN` set, it runs that file in place of the
`bootstrap` of mise. Its order is:

1. It prints the header `Syncing the shared files from bootstrap ...`. Then it
   runs `bootstrap sync` in the root of the repository, with
   `MISE_PROJECT_ROOT` set to the root.
2. For each selected submodule, it prints the header `Syncing <label> ...`.
   Then it does step 1 in the folder of the submodule, with
   `MISE_PROJECT_ROOT` set to that folder.

The first failure stops the task, and the exit code of the engine is the exit
code of the task. `bootstrap sync` takes the clone, examines its version and
syncs: see [sync-engine.md](sync-engine.md#the-cache-and-the-clone).

| Flag                | Effect                                                                                          |
| ------------------- | ----------------------------------------------------------------------------------------------- |
| `--dry-run`         | Shows the changes. Writes nothing.                                                              |
| `--verbose`         | Sets `BOOTSTRAP_VERBOSE=1`. The task does not pass the flag to the engine.                      |
| `--all`, `--<name>` | Also syncs each submodule member, or the named one. Only a repository with submodules has them. |

A submodule is a separate repository with its own values file. A submodule
without a values file stops the engine with code 2.

By default, the sync starts with one info line:
`bootstrap <short commit> · <n> tools · environments: <environments>`.

`bootstrap sync` makes its own private copy of the clone, because it writes
the generated mise config into the copy.

## The values file and the selection

`bootstrap sync` reads the consumer file `.config/bootstrap.toml`, and its
overrides file `.config/bootstrap.overrides.toml`. The keys are in
[bootstrap-config.md](bootstrap-config.md).

- It adds the pack `base`. Its name is fixed in the engine. See
  [packs.md](packs.md).
- It expands each pack to its default members, adds the `depends` of each tool
  and the tool of each mise `depends`, depth first, and copies each file of
  each tool. A dependency comes before the tool that needs it.
- The values file is necessary. A missing file stops the sync. The old name
  `.config/repo-config.toml` is read, with a message: rename it.
- It examines the `tools.toml` of each tool of the clone, selected or not. See
  [tools.md](tools.md).
- An old form of the values file stops the sync with "run bootstrap init". See
  [bootstrap-config.md](bootstrap-config.md#old-forms).

## Members and submodules

A member is a git submodule or a folder below the root. The keys are in
[`[member.<name>]`](bootstrap-config.md#membername).

- The member flags (`--all`, `--<name>`) of `setup`, `setup:configs`,
  `code:format` and `code:lint` are only for the submodules. A folder member is
  part of the root tree.
- A task that runs in a member uses its folder.
- A submodule is a separate repository, with its own values file and its own
  manifest. The sync never writes a file of a submodule.
- The members are in the input of each script: `repo.members`. See
  [sync-engine.md](sync-engine.md#the-scripts-of-a-tool). The derive script of
  `mise` makes the aliases `setup-all` and `setup-<member>` from them.

### `root_only`

A `root_only` tool is not for a git submodule. In a submodule, the sync skips it
when a pack adds it, and stops when a `[tools.*]` entry names it.

## The template context

The context of a template is the values file, with these values merged over it.
A key of the same name in the values file does not replace them. The reserved
names are in [bootstrap-config.md](bootstrap-config.md#reserved-names).

| Value             | Content                                                                                                                                                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `.members`        | The list of member names: the git submodules (the key of the `[member.<name>]` table whose `folder` is the submodule path, else the `repo_name` of the values file of the submodule, else the path with each `/` changed to `_`), then each other `[member.<name>]` key. |
| `.member_folders` | A map from each name to its folder (`folder`, default the name).                                                                                                                                                                                                         |
| `.member_types`   | A map from each name to `submodule` or `folder` (`type`, default from the folder).                                                                                                                                                                                       |
| `.submodules`     | The list of the git submodules only, nested ones included. It is empty in a repository without submodules.                                                                                                                                                               |
| `.tools`          | The list of the selected tools, with the packs expanded and the dependencies added.                                                                                                                                                                                      |
| `.sync_engine`    | The absolute path of the `sync-engine/` folder of the clone copy of the run.                                                                                                                                                                                             |
| `.config`, `.all` | The values of the tools. See [sync-engine.md](sync-engine.md#templates).                                                                                                                                                                                                 |

A name has only letters, digits, `-` and `_`, and two members never have the
same name. If not, the sync stops.

## The branch model and the merge tasks

The `git` tool holds the branch model. The engine has no branch rule.

- The derive script of `git` gives `branch_pair` (true when the repository has
  both `main` and `develop`, locally or on `origin`) and `protected_branches`
  (`main` and `develop`, without each one with `direct_commit = true`). It reads
  `direct_commit` from the values file and its overrides file.
- The tasks `code:merge:develop` and `code:merge:main` of `git` have
  `when = "git.branch_pair"`. The task `code:merge` (`code/merge/_default`) of
  `mise` has `when = "!git.branch_pair"`. Thus the sync writes the merge tasks
  of the model that it finds, and the stale rule removes the others.
- The check script of `git` stops the sync when neither branch exists.

| Branches             | Tasks                                                                                                                                                                  |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `main` and `develop` | Work starts on feature branches. `code:merge:develop <branch>` lands a branch on `develop`. `code:merge:main` lands `develop` on `main`, and runs only from `develop`. |
| Only one of them     | `code:merge <branch>` lands any feature branch on that branch.                                                                                                         |

- Merges use `--no-ff`. A hook pass over all files runs first as a safety net,
  with `SKIP=no-commit-to-branch`: that hook guards a commit, and the safety net
  makes none. Thus `code:merge:main` also runs from a protected `develop`.
- The merge tasks read `[tools.git.branch.<destination>]` of the repository's
  own `.config/bootstrap.toml` when they run, with `yq`: the root and each
  submodule read their own file.
- A merge task stops before it touches a branch when the git top level of
  `MISE_ORIGINAL_CWD` is not the git top level of `MISE_PROJECT_ROOT`. In a
  submodule without that task, mise runs the root's task in the root folder,
  and the task would land a branch of the root.
- `merge_model` selects `direct` (local merge and push, the default) or `pr`
  (push and open a pull request with `gh` or `glab`). Any other value reads as
  `direct`. Without an `origin` remote, `direct` merges locally and skips the
  pull and the push, and `pr` stops before the hook pass. The `github` and
  `gitlab` tools install these CLIs.
- `merge_from` lists the only source branches, and `merge_never` the forbidden
  ones. A table with both stops the merge. Without both, the defaults are: into
  `main` with `develop`, only from `develop`; into another branch, never from
  `main`.

### `direct_commit`

`direct_commit = true` in `[tools.git.branch.main]` or
`[tools.git.branch.develop]` lets a direct commit to that branch. The sync reads
it through the scripts of `git`, and the merge tasks do not read it.

The `no-commit-to-branch` hook of the `git` tool protects only the core branches
in `protected_branches`. The pre-commit spec gives the hook one
`"--branch", "<branch>"` pair for each of them, and removes the hook when no
branch is left, because without `--branch` the hook protects `main` and
`master`.

| Core branches of the repository | `direct_commit = true` in         | Hook args                         |
| ------------------------------- | --------------------------------- | --------------------------------- |
| `main` only                     | `[tools.git.branch.main]`         | No hook                           |
| `develop` only                  | `[tools.git.branch.develop]`      | No hook                           |
| `main` and `develop`            | `[tools.git.branch.develop]` only | `--branch main`                   |
| `main` and `develop`            | Both tables                       | No hook                           |
| `main` and `develop`            | `[tools.git.branch.main]` only    | The sync stops                    |
| `main` and `develop`            | No table                          | `--branch main --branch develop`  |
| `main` only, or `develop` only  | No table                          | `--branch <branch>` of the branch |

The check script of `git` examines the key before the sync writes a file, also
when it is `false`. These values stop the sync:

- A value that is not a boolean: `direct_commit must be true or false`.
- The key in a `[tools.git.branch.*]` table other than `main` and `develop`.
- The key for a core branch that does not exist, locally or on `origin`. The
  message names the table and the branch: create the branch, or remove the key.
- With `main` and `develop`, `direct_commit = true` for `main` but not for
  `develop`. A repository with both branches never lets a direct commit to
  `main` and refuses it on `develop`.

## The pre-commit config

`.config/pre-commit-config.yaml` is a shared file of the `pre-commit` tool. Its
spec is `repo/tools/pre-commit/files/.config/pre-commit-config.yaml.merge.toml`.
See [sync-engine.md](sync-engine.md#shared-files).

- The hook repositories come from `[[defaults.repos]]` of `pre-commit` and from
  `[[contributes.pre-commit.repos]]` of each selected tool. Each repository has
  an `id`, and the spec writes it as one block, found by `repo` and `hooks.id`.
- The `rev` of a found block stays (`ignore = ["rev"]`). `setup:precommit` runs
  `pre-commit autoupdate`, so each `rev` becomes the newest tag. The sync never
  calls `pre-commit`.
- The top-level `exclude` comes from `.config.exclude`, the list of the values
  of the tool: `[defaults]`, then `[contributes.pre-commit]` of each selected
  tool, then `[tools.pre-commit]` of the values file. pre-commit reads one
  regular expression, so the spec joins the items with `|`.
- The own hooks of a repository are blocks that the user writes directly in
  the file. The sync never touches them.
- An override removes or skips a bootstrap block by its `id`, for example
  `[tools.pre-commit.files.".config/pre-commit-config.yaml".repos]` with
  `"conventional-commits" = { "$mode" = "remove" }`.

## fnox profiles

A repository does not use the fnox `default` profile, which is only for the
user-level fnox settings. The check script of the `fnox` tool examines
`.config/fnox.toml` when `fnox` is selected. It stops the sync on a top-level
`[secrets]` table (the `default` profile, also when empty), a
`[profiles.default]` table, or no `[profiles.<name>]` table. A repository with
one profile uses its `repo_name`. Each `fnox` call passes `-P <name>`. The rest
of the design is in [fnox.md](fnox.md).

## `requires_found`

The `[files]` and `[tasks]` key `requires_found` and the same key of a mise tool
hold a list of file names. The sync must find each name at the root or down to
4 folders below it, in a file that git tracks or does not ignore. The sync does
not look in git submodules, because a submodule has its own sync. A name has no
`/`. A target with `requires` and `requires_found` needs each entry of both.

## `bootstrap check`

`bootstrap check` examines the values file. It writes no file and does not set
`bootstrap_commit`. It runs each check and prints each message, so each error
shows at once.

1. The local rules: the old forms, `environments`, `env`, the structure of
   `[overrides]` and of `[member.*]`.
2. With a clone, the rules of the clone: the selection, the values of the tools
   (with the derive scripts), the questions without an answer, and the check
   script of each selected tool.

The command takes a clone as each command does, so it examines both parts. The
front-end skips part 2, with a line, when it has no clone.

### The gate `code:lint:60-bootstrap`

The gate runs `bootstrap check`. The `lint` hook runs `code:lint`, which runs
the gate with the staged files, so the tool adds no hook. The gate is hidden,
and it ignores `--fix`.

It prints nothing and exits 0 in these cases:

- Without `bootstrap` (`mise x -- which bootstrap`), and without
  `BOOTSTRAP_BIN`.
- With a list of files that does not include `.config/bootstrap.toml`, or with
  no list and no values file.
- The cache has no clone with `sync-engine/VERSION`.
- The raw `origin` of `<cache>/repo` is not one of five literal forms of the
  default URL. Thus the gate never uses code of another repository at commit
  time.

Else it runs `bootstrap check` with `BOOTSTRAP_CLONE` set to the cache clone.
Thus the gate takes no lock and never fetches. A binary that is older than
`VERSION` of the cache clone stops the gate with code 2.

## The manifest

The manifest is `.config/repo-generated.tsv`. It is committed, and each git
submodule has its own. Its shape and the stale rule are in
[sync-engine.md](sync-engine.md#the-manifest-and-the-stale-rule).

After a clean run, the order is the stale rule, the write of the manifest, then
`bootstrap_commit`. A run with a failure runs no stale rule and writes no
manifest. A manifest that the sync cannot read fails the run: it logs
`FAILED: .config/repo-generated.tsv`, and it skips the stale rule, the write of
the manifest and `bootstrap_commit`.

Handled paths are the files of the selected tools and their dependencies, also
through a mise `depends`, that this run wrote or kept (owned, seed or shared).
A seed or a shared file of such a tool that the run left in place counts as
kept, also one that `requires` or its environment skips. An owned file that
`requires`, its environment or a false `when` skips is not handled. A folder
member without its own `.git` is part of the root tree.

### What the sync deletes

The sync never deletes a file, except these:

- A synced file at a path where the sync needs a folder.
- The stale rule: a file with the bootstrap header that the manifest lists and
  this run does not handle, for example a merge task of the other branch model.
- An item that bootstrap added to a shared file, when no spec has it now.

The stale rule skips a seed and a shared file of a tool of this run. At a folder
path, the sync refuses to remove such a path and fails that file.
`docs/sync/migrate.md` tells consumers what to delete.

## The Claude plugins

The sync does not read `[tools.claude.plugins.user]` and
`[tools.claude.plugins.project]` of the values file. `setup:ai` of the `claude`
tool reads them when it runs, as the merge tasks read `[tools.git.branch.*]`.
Each key is `"<plugin>@<marketplace>" = "<source>"`. It adds a missing
marketplace from the source, installs each plugin that is missing at that
scope, and updates the marketplaces and the plugins.
