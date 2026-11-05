# `.config/bootstrap.toml` design

This document describes each key of the values file `.config/bootstrap.toml`. It
is
the one place of these keys. The adopting AI writes the file in a consumer
repository
(step 3 of `docs/sync/index.md`). `repo-sync` and the templates of `repo/tools/`
read
it. The keys of the Mac values file are in
[setup-sync.md](setup-sync.md#macostoml).

See also: [repo-sync.md](repo-sync.md), [packs.md](packs.md),
[tools.md](tools.md),
[mise.md](mise.md), [sync-engine.md](sync-engine.md) and [fnox.md](fnox.md).
The exit code of each rule that stops the sync or the render is in
[sync-engine.md](sync-engine.md#the-exit-codes).

## Example

```toml
commit_scopes = [ "backend", "web" ]
environments = [ "dev", "ci" ]
extensions = [ "astro-build.astro-vscode" ]
product_name = "my-product"
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"
worktree_files = [ ".config/mise.local.toml" ]

[tools.claude]

[tools.fnox]

[tools.github]

[tools.node]
env = [ "dev", "ci" ]

[tools.vscode]

[member.web]
folder = "projects/web"
label = "Web"

[branch.main]
merge_from = [ "develop" ]
merge_model = "pr"

[branch.develop]
merge_model = "direct"
merge_never = [ "main" ]
# direct_commit = true
```

## Summary of the keys

| Key                                 | Type            | Necessary | Read by                                |
| ----------------------------------- | --------------- | --------- | -------------------------------------- |
| `[tools.<folder>]`                  | Table           | No        | `repo-sync`                            |
| `[tools.<folder>].env`              | List of strings | No        | `repo-sync`                            |
| `[tools.<folder>].<key>`            | Any             | No        | The templates of the tools             |
| `[packs.<name>]`                    | Table           | No        | `repo-sync`                            |
| `[packs.<name>].env`                | List of strings | No        | `repo-sync`                            |
| `environments`                      | List of strings | No        | `repo-sync`                            |
| `repo_name`                         | String          | Yes       | `mise.toml`, `gitleaks.toml`           |
| `product_name`                      | String          | No        | `fnox.toml`                            |
| `commit_scopes`                     | List of strings | No        | `git-conventional-commits.yaml`        |
| `repo_url`                          | String          | No        | `github` tool                          |
| `worktree_files`                    | List of strings | No        | `setup:worktree`                       |
| `extensions`                        | List of strings | No        | `vscode` tool                          |
| `[member.<name>]`                   | Table           | No        | `repo-sync`, `setup`, `setup:worktree` |
| `[member.<name>].label`             | String          | No        | `setup`, `setup:worktree`              |
| `[member.<name>].folder`            | String          | Yes       | `repo-sync`, `setup`, `setup:worktree` |
| `[member.<name>].type`              | String          | No        | `repo-sync`                            |
| `[member.<name>].worktree_install`  | String          | No        | `setup:worktree`                       |
| `[branch.<branch>].merge_model`     | String          | No        | `code:merge:<branch>`                  |
| `[branch.<branch>].merge_from`      | List of strings | No        | `code:merge:<branch>`                  |
| `[branch.<branch>].merge_never`     | List of strings | No        | `code:merge:<branch>`                  |
| `[branch.<branch>].direct_commit`   | Boolean         | No        | `repo-sync`                            |
| `[overrides.mise.<path>]`           | Any             | No        | `repo-sync`                            |
| `[overrides.mise_env.<env>.<path>]` | Any             | No        | `repo-sync`                            |
| `[overrides.gitignore]`             | Table           | No        | `repo-sync`                            |
| `[claude.plugins.user]`             | Table           | No        | `setup:ai`                             |
| `[claude.plugins.project]`          | Table           | No        | `setup:ai`                             |
| `bootstrap_commit`                  | String          | No        | `repo-sync` writes it                  |

The gate `code:lint:repo-sync` runs `repo_sync --check` at each commit that
changes
this file. See [repo-sync.md](repo-sync.md#--check) for the rules that it
examines and
the rules that only `setup:configs` examines.

All the paths in this document are relative to the root of the consumer
repository.

## The keys

### `[tools.<folder>]`

One table for each tool of the repository. Each key is a folder
`repo/tools/<folder>/` of `bootstrap`. Do not select the tools of the `base`
pack: `repo-sync` always adds the `base` pack. A table for a base tool only
gives values. See
[packs.md](packs.md#the-base-pack).
A table for a dependency also selects the tool, and its values follow the rule
of a base or dependency tool in `[tools.<folder>].<key>`.

- `env` selects the environments, and each other key is a value of the tool. See
  `[tools.<folder>].<key>`
- `repo-sync` adds the `depends` of each tool, and the tool of each mise
  `depends`, depth first
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
  [repo-sync.md](repo-sync.md#root_only)

### `[tools.<folder>].env`

The environments of the tool, for example `["dev", "ci"]`. Without the key,
the tool gets the `default_envs` of its `tools.toml`. See
[mise.md](mise.md#the-environments-of-a-mise-tool) for the rules of this list.

### `[tools.<folder>].<key>`

A value of the tool, for its templates. The top-level keys of this file, for
example `commit_scopes` and `extensions`, stay at the top level.

```toml
[tools.node]
env = [ "dev", "ci" ]
registry = "https://registry.example.com"
```

The sync merges the value over the `[defaults]` of the tool and the
contributions
of the selected tools, and a template reads the result as `.config` or `.all`.
See
[sync-engine.md](sync-engine.md#the-values-merge).

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

| Output                  | Use                                         |
| ----------------------- | ------------------------------------------- |
| `.config/mise.toml`     | `REPO_NAME` in `[env]`                      |
| `.config/gitleaks.toml` | The `title` of the gitleaks config (a seed) |

The `setup:vscode` task uses `REPO_NAME` as the name of the VS Code profile.

### `product_name`

The name of the product, for example `my-product`. Set it in each repository of
a product with members, the root included.

### `commit_scopes`

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

### `worktree_files`

The gitignored files that `setup:worktree` copies from the main worktree into
a new worktree, for example `[".config/mise.local.toml"]`.

- The task copies a file only in a linked worktree.
- It does not replace a file that exists.
- It prints a warning for a file that is not in the main worktree.
- Without the key, the task has no copy step.

### `extensions`

VS Code extensions that no tool has, for example
`["astro-build.astro-vscode"]`.

- The sync adds them to the extensions of the selected tools, sorts the list,
  and removes the duplicates. The `vscode` tool writes the result to
  `.vscode/extensions.json`.
- In the root repository, the list also has the tools and the `extensions` of
  each git submodule, because VS Code opens the root only.
- A submodule lists its own extensions in its own values file.

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
  is in [repo-sync.md](repo-sync.md#the-template-context).
- A `[member.<name>]` table that is not a submodule must have its folder in
  the repository. If not, the render stops.
- No member is the repository itself: a `folder` of `.` stops the render. The
  merge config of the repository, and its `direct_commit`,
  are in `[branch.<branch>]`.
- The templates get these values. See
  [repo-sync.md](repo-sync.md#the-template-context). `MEMBERS` in
  `.config/mise.toml` is the list of the folders, with one space
  between them.
- `setup`, `setup:configs`, `code:format` and `code:lint` get a `--<name>`
  flag and an `--all` flag only for the members of type `submodule`. They run in
  the folder of the
  submodule. A folder member of a monorepo is part of the tree of the root, so
  it gets no flag. `setup:worktree` runs `worktree_install` for each submodule
  member.

### `[member.<name>].label`

The display name of the member in the headers of `setup` and
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

### `[branch.<branch>].merge_model`

How `code:merge:<branch>` merges a branch on specified branch: `direct`
or `pr`

Each repository reads only `[branch.<branch>]` of its own
`.config/bootstrap.toml`.
This applies to a repository without submodules, to the root of a repository
with
submodules, and to each submodule. Each merge task declares `yq` with
`#MISE tools.yq="latest"`. The merge rules are in
[repo-sync.md](repo-sync.md#the-branch-model-and-the-merge-tasks).

| Value    | Effect                                                         |
| -------- | -------------------------------------------------------------- |
| `direct` | A local `--no-ff` merge, then a push.                          |
| `pr`     | A push of the branch, then a pull request with `gh` or `glab`. |

The merge script uses `direct` when the value is empty

### `[branch.<branch>].merge_from`

List of branches the merge can be allowed from. Any branch can be merged from if
the list is empty, except from the ones mentioned in
`[branch.<branch>].merge_never`

- This can be set ONLY if `[branch.<branch>].merge_never` is empty

### `[branch.<branch>].merge_never`

List of branches the merge can NEVER be allowed from. If the list is empty and
`[branch.<branch>].merge_from` is set, only the branches of
`merge_from` will be allowed

- This can be set ONLY if `[branch.<branch>].merge_from` is empty

### `[branch.<branch>].direct_commit`

`true` allows a direct commit to the branch. `repo-sync` reads the key. The
merge
tasks do not read it. The key is only for `[branch.main]` and
`[branch.develop]`, the
core branches. Without the key, or with `false`, the `no-commit-to-branch` hook
of the
`git` tool refuses a direct commit to the branch. The effect on the hook and the
checks
of the key are in [repo-sync.md](repo-sync.md#direct_commit).

### `[overrides.mise.<path>]` and `[overrides.mise_env.<env>.<path>]`

Values that replace the defaults of the mise config of the tools. See
[mise.md](mise.md#the-overrides).

### `[overrides.gitignore]`

The keys `add` and `remove`, each a list of lines of the shared `.gitignore`.
See
[repo-sync.md](repo-sync.md#overridesgitignore).

### `[claude.plugins.user]` and `[claude.plugins.project]`

The Claude plugins of each scope. Each key is `"<plugin>@<marketplace>" =
"<source>"`.
See [repo-sync.md](repo-sync.md#claudeplugins).

### `bootstrap_commit`

The commit of the last sync without a failure. The sync writes it. An example
has an
empty value. See [sync-engine.md](sync-engine.md#bootstrap_commit).

## Reserved names

The sync merges these keys into the template context before the values file:
`members`, `member_folders`, `member_types`, `submodules`, `tools`,
`extensions`,
`sync_engine`, `precommit`, `config` and `all`.
Thus a key of
the same name in the values file does not replace the value of the sync.
`tools` and `extensions` of the values file are the input of the sync, not the
values that the templates get.

## Old keys and old names

| Old item                   | Effect                                                                                                                  |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `core_tools`               | Not used. The sync prints a message: delete it. The `env` of each entry, else `default_envs`, decides the environments. |
| `.config/repo-config.toml` | The old name of the values file. The sync reads it and prints a message: rename it.                                     |

## The location of the file

The default path is `.config/bootstrap.toml`. `REPO_SYNC_VALUES` sets a
different path. Without the file, the sync stops.
