# `.config/bootstrap.toml` design

This document describes each key of the values file `.config/bootstrap.toml`.
The adopting AI writes the file in a consumer repository (step 3 of
`docs/sync/index.md`). `repo-sync` and the templates of `tools/` read it

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

| Key                                | Type            | Necessary | Read by                                |
| ---------------------------------- | --------------- | --------- | -------------------------------------- |
| `[tools.<folder>]`                 | Table           | No        | `repo-sync`                            |
| `[tools.<folder>].env`             | List of strings | No        | `repo-sync`                            |
| `[packs.<name>]`                   | Table           | No        | `repo-sync`                            |
| `[packs.<name>].env`               | List of strings | No        | `repo-sync`                            |
| `environments`                     | List of strings | No        | `repo-sync`                            |
| `repo_name`                        | String          | Yes       | `mise.toml`, `gitleaks.toml`           |
| `product_name`                     | String          | No        | `fnox.toml`                            |
| `commit_scopes`                    | List of strings | No        | `git-conventional-commits.yaml`        |
| `repo_url`                         | String          | No        | `github` tool                          |
| `worktree_files`                   | List of strings | No        | `setup:worktree`                       |
| `extensions`                       | List of strings | No        | `vscode` tool                          |
| `[member.<name>]`                  | Table           | No        | `repo-sync`, `setup`, `setup:worktree` |
| `[member.<name>].label`            | String          | No        | `setup`, `setup:worktree`              |
| `[member.<name>].folder`           | String          | Yes       | `repo-sync`, `setup`, `setup:worktree` |
| `[member.<name>].worktree_install` | String          | No        | `setup:worktree`                       |
| `[branch.<branch>].merge_model`    | String          | No        | `code:merge:<branch>`                  |
| `[branch.<branch>].merge_from`     | List of strings | No        | `code:merge:<branch>`                  |
| `[branch.<branch>].merge_never`    | List of strings | No        | `code:merge:<branch>`                  |
| `[branch.<branch>].direct_commit`  | Boolean         | No        | `repo-sync`                            |

All the paths in this document are relative to the root of the consumer
repository.

## The keys

### `[tools.<folder>]`

One table for each tool of the repository. Each key is a folder
`tools/<folder>/` of `bootstrap`. Do not write the tools of the `base` pack:
`repo-sync` always adds the `base` pack. See
[packs.md](packs.md#the-base-pack)

- The only key of an entry is `env`. Another key stops the sync with code 2
- `repo-sync` adds the `depends` of each tool, and the tool of each mise
  `depends`, depth first
- All dependencies are configured for same environment as parent
- A tool that is installed for `dev` environment will have all it's dependencies
  configured for `dev`
- The `envs` of a dependency must include each environment of its parent. If
  not, the sync stops with code 2. See
  [mise.md](mise.md#the-environments-of-a-mise-tool)
- Without `[tools]` and `[packs]`, the repository gets only the `base` pack
- A list, `tools = [...]`, is the old format. It stops the sync with code 2,
  and the message names the table format
- A folder that is not in `bootstrap` stops the sync with code 3
- A tool folder of `bootstrap` that is `zz-repo`, or whose name sorts after
  `zz-repo` in byte order, stops the sync with code 2. The folder
  `conf.d/zz-repo/` of the repository must load last. See
  [mise.md](mise.md)
- A tool with `root_only = true` in a git submodule stops the sync with
  code 2

### `[tools.<folder>].env`

The environments of the tool, for example `["dev", "ci"]`. Without the key,
the tool gets the `default_envs` of its `tools.toml`. Each `default_envs`
must include `dev`. A `default_envs` without `dev` stops the sync with code 2,
also for a tool that the values file does not select.

- Each name must be in `environments`. If not, the sync stops with code 2
- The list must include `dev`. If not, the sync stops with code 2
- An empty list, or an item with a space, stops the sync with code 2
- Each mise tool of the tool must allow each environment in its `envs`. If
  not, the sync stops with code 2
- A tool that is selected more than once, for example also by a pack, gets
  each of these environments

### `[packs.<name>]`

One table for each pack of the repository. Each key is a pack of
`tools/packs.toml`, without the prefix `p:`. Do not write `base`. See
[packs.md](packs.md)

- The only key of an entry is `env`. Another key stops the sync with code 2
- A list, `packs = [...]`, is the old format. It stops the sync with code 2
- A pack that is not in `tools/packs.toml` stops the sync with code 3

### `[packs.<name>].env`

The environments of each tool of the pack, with the same rules as
`[tools.<folder>].env`. Without the key, each tool of the pack gets its own
`default_envs`.

### `environments`

The mise environments of the repository, from `dev` and `ci`

- `dev` is necessary. `ci` is optional
- Without the key, the sync uses `dev` only and prints a message.
- An unknown name, or a list without `dev`, stops the sync with code 2.
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

- A submodule without a table is a member. Its name and its folder are its
  path. A table whose `folder` is a submodule path gives that submodule its
  name.
- A `[member.<name>]` table that is not a submodule must have its folder in
  the repository. If not, the render stops with code 2.
- No member is the repository itself: a `folder` of `.` stops the render
  with code 2. The merge config of the repository, and its `direct_commit`,
  are in `[branch.<branch>]`.
- The templates get the names as `.members`, the folder of each name as
  `.member_folders`, and the type of each name as `.member_types`.
  `MEMBERS` in `.config/mise.toml` is the list of the folders, with one space
  between them.
- `setup`, `code:format` and `code:lint` get a `--<name>` flag and an `--all`
  flag only for the members of type `submodule`. They run in the folder of the
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
with code 2 when the type is not correct, or when the value is a different
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
`.config/bootstrap.toml`. This applies to a repository without submodules, to
the root of a repository with submodules, and to each submodule. The merge tasks
read the file when they run, with
`mise x -- yq`.

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
- Without both lists, the defaults are the branch model: into `main` when
  `develop` exists, only from `develop` (`merge_from = ["develop"]`); into any
  other branch, never from `main` (`merge_never = ["main"]`)
- A table that sets both lists stops the merge

### `[branch.<branch>].direct_commit`

`true` allows a direct commit to the branch. `repo-sync` reads the key. The
merge tasks do not read it. The key is only for `[branch.main]` and
`[branch.develop]`, the core branches. Without the key, or with `false`, the
`no-commit-to-branch` hook of the `git` tool refuses a direct commit to the
branch.

The hook protects only the core branches that exist and that do not allow
direct commits. The sync removes the `"--branch", "<branch>"` pair of each
core branch that does not exist and of each core branch with
`direct_commit = true`. When no pair is left, the sync removes the hook.

| Core branches of the repository | `direct_commit = true` in | Hook args                         |
| ------------------------------- | ------------------------- | --------------------------------- |
| `main` only                     | `[branch.main]`           | No hook                           |
| `develop` only                  | `[branch.develop]`        | No hook                           |
| `main` and `develop`            | `[branch.develop]` only   | `--branch main`                   |
| `main` and `develop`            | Both tables               | No hook                           |
| `main` and `develop`            | `[branch.main]` only      | The sync stops with code 2        |
| `main` and `develop`            | No table                  | `--branch main --branch develop`  |
| `main` only, or `develop` only  | No table                  | `--branch <branch>` of the branch |

These values stop the sync with code 2, before it writes a file. The sync
validates the key when it is present, also when it is `false`:

- A value that is not a boolean: `direct_commit must be true or false`
- The key in a `[branch.*]` table other than `[branch.main]` and
  `[branch.develop]`:
  `direct_commit is a key of [branch.main] and [branch.develop] only`
- The key for a core branch that does not exist, locally or on `origin`. The
  message names the table and the branch: create the branch, or remove the
  key
- With `main` and `develop`, `direct_commit = true` for `main` but not for
  `develop`. A repository with both branches never allows a direct commit to
  `main` and refuses it on `develop`

The same rule applies to a `no-commit-to-branch` hook in
`.config/repo-sync/pre-commit.yaml`. The sync changes only its
`"--branch", "<branch>"` pairs. A hook without args, or with `--branch=<b>` or
`-b <b>`, does not change. A hook that had `--branch` pairs and has none left
goes, also when it has other args: without `--branch`, the hook protects
`main` and `master`. An entry with no hook left goes too.

## Reserved names

The sync merges these keys into the template context before the values file:
`members`, `submodules`, `tools`, `extensions` and `precommit`. Thus a key of
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
different path. Without the file, the sync stops with code 2.
