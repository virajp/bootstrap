# `.config/bootstrap.toml` design

This document describes each key of the values file `.config/bootstrap.toml`,
and its old forms. It is the one place of these keys. `bootstrap init` writes
the file in a consumer repository, and the adopting AI gives the answers (see
`docs/sync/index.md`). The sync and the templates of `repo/tools/` read it. The
keys of the Mac values file are in [mac.md](mac.md#macostoml).

See also: [repo.md](repo.md), [packs.md](packs.md),
[tools.md](tools.md),
[mise.md](mise.md), [sync-engine.md](sync-engine.md) and [fnox.md](fnox.md).
The exit code of each rule that stops the sync or the render is in
[sync-engine.md](sync-engine.md#the-exit-codes).

## Example

```toml
environments = [ "dev", "ci" ]
product_name = "my-product"
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"

[tools.claude]

[tools.claude.plugins.project]
"vwf@my-marketplace" = "github:me/claude-plugins"

[tools.fnox]

[tools.git]
commit_scopes = [ "backend", "web" ]

[tools.git.branch.main]
merge_from = [ "develop" ]
merge_model = "pr"

[tools.git.branch.develop]
merge_model = "direct"
merge_never = [ "main" ]
# direct_commit = true

[tools.github]

[tools.mise]
worktree_files = [ ".config/mise.local.toml" ]

[tools.node]
env = [ "dev", "ci" ]

[member.web]
folder = "projects/web"
label = "Web"
```

## Summary of the keys

| Key                                         | Type            | Necessary | Read by                                                                          |
| ------------------------------------------- | --------------- | --------- | -------------------------------------------------------------------------------- |
| `[tools.<folder>]`                          | Table           | No        | The sync                                                                         |
| `[tools.<folder>].env`                      | List of strings | No        | The sync                                                                         |
| `[tools.<folder>].<key>`                    | Any             | No        | The templates and the scripts of the tools                                       |
| `[tools.<folder>.files."<target>"]`         | Table           | No        | The sync: the overrides of a shared file                                         |
| `[packs.<name>]`                            | Table           | No        | The sync                                                                         |
| `[packs.<name>].env`                        | List of strings | No        | The sync                                                                         |
| `environments`                              | List of strings | No        | The sync                                                                         |
| `repo_name`                                 | String          | Yes       | `mise.toml`, `gitleaks.toml`, `fnox.toml`                                        |
| `product_name`                              | String          | No        | `fnox.toml`                                                                      |
| `repo_url`                                  | String          | No        | `github` tool                                                                    |
| `[tools.git].commit_scopes`                 | List of strings | No        | `git-conventional-commits.yaml`                                                  |
| `[tools.git.branch.<branch>].merge_model`   | String          | No        | `code:merge:<branch>`                                                            |
| `[tools.git.branch.<branch>].merge_from`    | List of strings | No        | `code:merge:<branch>`                                                            |
| `[tools.git.branch.<branch>].merge_never`   | List of strings | No        | `code:merge:<branch>`                                                            |
| `[tools.git.branch.<branch>].direct_commit` | Boolean         | No        | The scripts of `git`                                                             |
| `[tools.mise].worktree_files`               | List of strings | No        | `setup:worktree`                                                                 |
| `[tools.claude.plugins.user]`               | Table           | No        | `setup:ai`                                                                       |
| `[tools.claude.plugins.project]`            | Table           | No        | `setup:ai`                                                                       |
| `[member.<name>]`                           | Table           | No        | The sync, `setup`, `setup:configs`, `setup:worktree`                             |
| `[member.<name>].label`                     | String          | No        | `setup`, `setup:configs`, `setup:worktree`                                       |
| `[member.<name>].folder`                    | String          | No        | The sync, `setup`, `setup:configs`, `setup:worktree`, `code:format`, `code:lint` |
| `[member.<name>].type`                      | String          | No        | The sync                                                                         |
| `[member.<name>].worktree_install`          | String          | No        | `setup:worktree`                                                                 |
| `[overrides.mise.<path>]`                   | Any             | No        | The sync                                                                         |
| `[overrides.mise_env.<env>.<path>]`         | Any             | No        | The sync                                                                         |
| `bootstrap_commit`                          | String          | No        | The sync writes it                                                               |

A key of a question is a `[tools.<folder>].<key>`: `bootstrap init` and
`bootstrap add` write the answer there. See
[sync-engine.md](sync-engine.md#the-questions).

The gate `code:lint:60-bootstrap` runs `bootstrap check` at each commit that
changes this file. See [repo.md](repo.md#bootstrap-check).

All the paths in this document are relative to the root of the consumer
repository.

## The keys

### `[tools.<folder>]`

One table for each tool of the repository. Each key is a folder
`repo/tools/<folder>/` of `bootstrap`. Do not select the tools of the `base`
pack: the sync always adds the `base` pack. `bootstrap add` and
`bootstrap remove` change these tables. A table for a base tool gives
values. An `env` in that table can only add environments, and the sync examines
it. See
[packs.md](packs.md#the-base-pack).
A table for a dependency also selects the tool, and its values follow the rule
of a base or dependency tool in `[tools.<folder>].<key>`.

- `env` selects the environments, and each other key is a value of the tool. See
  `[tools.<folder>].<key>`
- The sync adds the `depends` of each tool, depth first
- All dependencies are configured for same environment as parent
- A tool that is installed for `dev` environment will have all it's dependencies
  configured for `dev`
- The `envs` of a dependency must include each environment of its parent. If
  not, the sync stops. See
  [mise.md](mise.md#the-environments-of-a-mise-tool)
- Without `[tools]` and `[packs]`, the repository gets only the `base` pack
- A list, `tools = [...]`, is the old format. It stops the sync, and the
  message names the table format
- A folder that is not in `bootstrap` stops the sync
- A tool with `root_only = true` in a git submodule stops the sync. See
  [repo.md](repo.md#root_only)

### `[tools.<folder>].env`

The environments of the tool, for example `["dev", "ci"]`. Without the key,
the tool gets the `default_envs` of its `tools.toml`. See
[mise.md](mise.md#the-environments-of-a-mise-tool) for the rules of this list.

### `[tools.<folder>].<key>`

A value of the tool, for its templates and its scripts. A key of a tool is
always in its table, for example `commit_scopes` in `[tools.git]`.

```toml
[tools.node]
env = [ "dev", "ci" ]
registry = "https://registry.example.com"
```

The sync merges the value over the `[defaults]` of the tool and the
contributions
of the selected tools and the derived values, and a template reads the result
as `.config` or `.all`. See
[sync-engine.md](sync-engine.md#the-values-of-a-tool).

### `[packs.<name>]`

One table for each pack of the repository. Each key is a pack of
`repo/packs.toml`, without the prefix `pack:`. Do not write `base`. See
[packs.md](packs.md)

- The only key of an entry is `env`. Another key stops the sync
- A list, `packs = [...]`, is the old format. It stops the sync
- A pack that is not in `repo/packs.toml` stops the sync

### `[packs.<name>].env`

The environments of each tool of the pack, with the same rules as
`[tools.<folder>].env`. Without the key, each tool of the pack gets its own
`default_envs`.

### `environments`

The mise environments of the repository, from `dev` and `ci`

- `dev` is necessary. `ci` is optional
- Without the key, the sync uses `dev` only and prints a message.
- An unknown name, or a list without `dev`, stops the sync.
- The sync writes `conf.d/<folder>/mise.<env>.toml` only for these
  environments. See [mise.md](mise.md).

### `repo_name`

The name of the repository, for example `my-repo`. It is a literal, not the
folder name of a worktree.

| Output                  | Use                                                         |
| ----------------------- | ----------------------------------------------------------- |
| `.config/mise.toml`     | `REPO_NAME` in `[env]`                                      |
| `.config/gitleaks.toml` | The `title` of the gitleaks config (a seed)                 |
| `.config/fnox.toml`     | The profile and the keychain `service` or `prefix` (a seed) |

The `setup:vscode` task uses `REPO_NAME` as the name of the VS Code profile.

### `product_name`

The name of the product, for example `my-product`. Set it in each repository of
a product with members, the root included.

With this key, the fnox keychain uses it as `service`, and `<repo_name>/` as
`prefix`. Without this key, `service` is `repo_name` and there is no prefix.

### `[tools.git].commit_scopes`

The commit scopes of `.config/git-conventional-commits.yaml`, for example
`["backend", "web"]`.

- Without the key, or with an empty list, the scopes are the members of the
  repository.
- Without members, the list of scopes is empty.

### `repo_url`

The web URL of the repository, for example `https://github.com/me/my-repo`.

| Output                              | Use                                                     |
| ----------------------------------- | ------------------------------------------------------- |
| `.github/ISSUE_TEMPLATE/config.yml` | The links to the README and to the security advisories. |

Without the key, the templates write no links.

### `[tools.mise].worktree_files`

The gitignored files that `setup:worktree` copies from the main worktree into
a new worktree, for example `[".config/mise.local.toml"]`.

- The task copies a file only in a linked worktree.
- It does not replace a file that exists.
- It prints a warning for a file that is not in the main worktree.
- Without the key, the task has no copy step.

### `[member.<name>]`

One table for each member folder of the repository. The member list is the
git submodules (else the `MEMBERS` value), then each `[member.<name>]` key
that is not a submodule.

- A submodule without a table is a member. Its folder is its path. Its name
  is the `repo_name` of its own `.config/bootstrap.toml`. Without that value,
  its name is its path with each `/` changed to `_`, for example
  `backend_lib`. A table whose `folder` is a submodule path gives that
  submodule its name.
- A member name is a flag and a part of a shell variable. The rule for the name
  is in [repo.md](repo.md#the-template-context).
- A `[member.<name>]` table that is not a submodule must have its folder in
  the repository. If not, the render stops.
- No member is the repository itself: a `folder` of `.` stops the render. The
  merge config of the repository, and its `direct_commit`, are in
  `[tools.git.branch.<branch>]`.
- The templates get these values. See
  [repo.md](repo.md#the-template-context). `MEMBERS` in
  `.config/mise.toml` is the list of the folders, with one space
  between them.
- `setup`, `setup:configs`, `code:format` and `code:lint` get a `--<name>`
  flag and an `--all` flag only for the members of type `submodule`. They run in
  the folder of the
  submodule. A folder member of a monorepo is part of the tree of the root, so
  it gets no flag. `setup:worktree` runs `worktree_install` for each submodule
  member.

### `[member.<name>].label`

The display name of the member in the headers of `setup`, `setup:configs` and
`setup:worktree`. The default is `<name>`.

### `[member.<name>].folder`

The folder of the member, relative to the root of the repository, for example
`projects/web`. The default is `<name>`. A polyrepo has no members, so it has
no `[member.<name>]` table.

### `[member.<name>].type`

`submodule` or `folder`. Optional: without it, the type is `submodule` when
`.gitmodules` has the `folder` path, else `folder`. With it, the render stops
when the type is not correct, or when the value is a different
word.

### `[member.<name>].worktree_install`

The install command of the member for a new worktree, for example
`pnpm install`. `setup:worktree` runs it as
`mise x --cd <folder> -- <command>`. Without the key, `setup:worktree` runs the
own `setup:worktree` task of the member, if the member has one.

The key is only for a member of type `submodule`. A member of type `folder`
never has it.

### `[tools.git.branch.<branch>].merge_model`

How `code:merge:<branch>` merges a branch on specified branch: `direct`
or `pr`

Each repository reads only `[tools.git.branch.<branch>]` of its own
`.config/bootstrap.toml`.
This applies to a repository without submodules, to the root of a repository
with
submodules, and to each submodule. Each merge task declares `yq` with
`#MISE tools.yq="latest"`. The merge rules are in
[repo.md](repo.md#the-branch-model-and-the-merge-tasks).

| Value    | Effect                                                         |
| -------- | -------------------------------------------------------------- |
| `direct` | A local `--no-ff` merge, then a push.                          |
| `pr`     | A push of the branch, then a pull request with `gh` or `glab`. |

The merge script uses `direct` when the value is empty

### `[tools.git.branch.<branch>].merge_from`

The only branches that can merge into the branch. An empty list gives no limit.

- A table with both `merge_from` and `merge_never` stops the merge, also when a
  list is empty.
- The defaults without both keys are in
  [repo.md](repo.md#the-branch-model-and-the-merge-tasks).

### `[tools.git.branch.<branch>].merge_never`

The branches that must not merge into the branch. An empty list gives no limit.
The rule for both keys and the defaults are in
`[tools.git.branch.<branch>].merge_from`.

### `[tools.git.branch.<branch>].direct_commit`

`true` lets a direct commit to the branch. The scripts of the `git` tool read
the key. The merge tasks do not read it. The key is only for
`[tools.git.branch.main]` and `[tools.git.branch.develop]`, the core branches.
Without the key, or with `false`, the `no-commit-to-branch` hook of the `git`
tool refuses a direct commit to the branch. The effect on the hook and the
checks of the key are in [repo.md](repo.md#direct_commit).

### `[overrides.mise.<path>]` and `[overrides.mise_env.<env>.<path>]`

Values that replace the defaults of the mise config of the tools. See
[mise.md](mise.md#the-overrides).

### `[tools.<folder>.files."<target>"]`

The overrides of the bootstrap items of a shared file of the tool, for example
`[tools.git.files.".gitignore"]` with `"*.log" = { "$mode" = "remove" }`. See
[sync-engine.md](sync-engine.md#overrides-of-a-shared-file). The own items of
the user go directly in the file.

### `[tools.claude.plugins.user]` and `[tools.claude.plugins.project]`

The Claude plugins of each scope. Each key is `"<plugin>@<marketplace>" =
"<source>"`. See [repo.md](repo.md#the-claude-plugins).

### `bootstrap_commit`

The commit of the last sync without a failure. The sync writes it. An example
has an
empty value. See [sync-engine.md](sync-engine.md#bootstrap_commit).

## Reserved names

The sync merges these keys into the template context before the values file:
`members`, `member_folders`, `member_types`, `submodules`, `tools`,
`sync_engine`, `config` and `all`. Thus a key of the same name in the values
file does not replace the value of the sync. `tools` of the values file is the
input of the sync, not the value that the templates get.

## Old forms

`bootstrap init` migrates each old form, with the comments of the values file
kept. Each other command stops with code 2 and the message "run bootstrap
init". A dry run of `init` shows the migration and writes nothing.

| Old form                        | `bootstrap init` moves it to                                                                       |
| ------------------------------- | -------------------------------------------------------------------------------------------------- |
| `commit_scopes`                 | `[tools.git].commit_scopes`                                                                        |
| `[branch.<branch>]`             | `[tools.git.branch.<branch>]`                                                                      |
| `worktree_files`                | `[tools.mise].worktree_files`                                                                      |
| `[claude.plugins.*]`            | `[tools.claude.plugins.*]`                                                                         |
| `extensions`                    | `.vscode/extensions.json`, as own items of the shared file                                         |
| `[overrides.gitignore]`         | `add`: lines of `.gitignore`. `remove`: `[tools.git.files.".gitignore"]` with `"$mode" = "remove"` |
| The table of the old entry tool | Removed: the `mise` tool has the entry now                                                         |
| The old file of the own hooks   | Its hooks move to `.config/pre-commit-config.yaml` as own blocks, and the file goes                |

`docs/sync/migrate.md` names the old entry tool and its old files. These old
items are not migrated:

| Old item                                   | Effect                                                                                                                  |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| The old file of the own `.gitignore` lines | Stops the sync: move its lines into `.gitignore`, then delete it.                                                       |
| `[overrides.graphifyignore]`               | Stops the sync: remove the table.                                                                                       |
| `core_tools`                               | Not used. The sync prints a message: delete it. The `env` of each entry, else `default_envs`, decides the environments. |
| `.config/repo-config.toml`                 | The old name of the values file. The sync reads it and prints a message: rename it.                                     |

## The location of the file

The default path is `.config/bootstrap.toml`. `REPO_SYNC_VALUES` sets a
different path. Without the file, each command except `bootstrap init` stops.
The overrides file is `.config/bootstrap.overrides.toml` next to it, with only
`[tools.<folder>]` tables. See
[sync-engine.md](sync-engine.md#the-values-of-a-tool).
