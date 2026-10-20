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
tools = [ "github", "node", "fnox", "claude", "vscode" ]
worktree_files = [ ".config/mise.local.toml" ]

[member.web]
folder = "projects/web"
label = "Web"
worktree_install = "mise run setup"

[member.web.branch.main]
merge_from = [ "develop" ]
merge_model = "pr"

[member.web.branch.develop]
merge_model = "direct"
merge_never = [ "main" ]
```

## Summary of the keys

| Key                                           | Type            | Necessary | Read by                                |
| --------------------------------------------- | --------------- | --------- | -------------------------------------- |
| `tools`                                       | List of strings | No        | `repo-sync`                            |
| `environments`                                | List of strings | No        | `repo-sync`                            |
| `repo_name`                                   | String          | Yes       | `mise.toml`, `gitleaks.toml`           |
| `product_name`                                | String          | No        |                                        |
| `commit_scopes`                               | List of strings | No        | `git-conventional-commits.yaml`        |
| `repo_url`                                    | String          | No        | `github` tool                          |
| `worktree_files`                              | List of strings | No        | `setup:worktree`                       |
| `extensions`                                  | List of strings | No        | `vscode` tool                          |
| `[member.<name>]`                             | Table           | No        | `repo-sync`, `setup`, `setup:worktree` |
| `[member.<name>].label`                       | String          | No        | `setup`, `setup:worktree`              |
| `[member.<name>].folder`                      | String          | Yes       | `repo-sync`, `setup`, `setup:worktree` |
| `[member.<name>].worktree_install`            | String          | No        | `setup:worktree`                       |
| `[member.<name>.branch.<branch>].merge_model` | String          | No        | `code:merge:<branch>`                  |
| `[member.<name>.branch.<branch>].merge_from`  | List of strings | No        | `code:merge:<branch>`                  |
| `[member.<name>.branch.<branch>].merge_never` | List of strings | No        | `code:merge:<branch>`                  |

All the paths in this document are relative to the root of the consumer
repository.

## The keys

### `tools`

The tools of the repository. Each name is a folder `tools/<name>/` of
`bootstrap`. Do not list the tools of the `base` pack: `repo-sync` always adds
the `base` pack. See [packs.md](packs.md#the-base-pack)

- `repo-sync` adds the `depends` of each tool, and the tool of each mise
  `depends`, depth first
- All dependencies are configured for same environment as parent
- A tool that is installed for `dev` environment will have all it's dependencies
  configured for `dev`
- The `envs` of a dependency must include each environment of its parent. If
  not, the sync stops with code 2. See
  [mise.md](mise.md#the-environments-of-a-mise-tool)
- Without the key, or with an empty list, the repository gets only the `base`
  pack
- A name that is not in `bootstrap` stops the sync with code 3
- A tool with `root_only = true` in a git submodule stops the sync with
  code 2

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

- A `[member.<name>]` table that is not a submodule must have the folder of
  its `folder` key in the repository. If not, the render stops with code 2.
- The templates get the list as `.members`. `MEMBERS` in `.config/mise.toml`
  is this list, with one space between the names.
- `setup`, `code:format` and `code:lint` get a `--<name>` flag and an `--all`
  flag for the members.
- A repo without any member will have itself as a member

### `[member.<name>].label`

The display name of the member in the headers of `setup` and
`setup:worktree`. The default is `<name>`.

### `[member.<name>].folder`

This must contain the folder of the submodule or member project. For a polyrepo,
it can simply be "."

### `[member.<name>].worktree_install`

The install command of the member for a new worktree, for example
`pnpm install`. `setup:worktree` runs it as
`mise x --cd <folder> -- <command>`. Without the key, `setup:worktree` runs the
own `setup:worktree` task of the member, if the member has one.

### `[member.<name>.branch.<branch>].merge_model`

How `code:merge:<branch>` merges a branch on specified branch: `direct`
or `pr`

| Value    | Effect                                                         |
| -------- | -------------------------------------------------------------- |
| `direct` | A local `--no-ff` merge, then a push.                          |
| `pr`     | A push of the branch, then a pull request with `gh` or `glab`. |

The merge script uses `direct` when the value is empty

### `[member.<name>.branch.<branch>].merge_from`

List of branches the merge can be allowed from. Any branch can be merged from if
the list is empty, except from the ones mentioned in
`[member.<name>.branch.<branch>].merge_never`

- This can be set ONLY if `[member.<name>.branch.<branch>].merge_never` is empty

### `[member.<name>.branch.<branch>].merge_never`

List of branches the merge can NEVER be allowed from. If the list is empty and
`[member.<name>.branch.<branch>].merge_from` is set, only the branches of
`merge_from` will be allowed

- This can be set ONLY if `[member.<name>.branch.<branch>].merge_from` is empty
- Both lists are empty only in a repository with a single protected branch
  (`main` or `develop`). Then any branch can be merged into that branch

## Reserved names

The sync merges these keys into the template context before the values file:
`members`, `submodules`, `tools`, `extensions` and `precommit`. Thus a key of
the same name in the values file does not replace the value of the sync.
`tools` and `extensions` of the values file are the input of the sync, not the
values that the templates get.

## Old keys and old names

| Old item                   | Effect                                                                                           |
| -------------------------- | ------------------------------------------------------------------------------------------------ |
| `core_tools`               | Not used. The sync prints a message: delete it. The `envs` of each tool decide its environments. |
| `.config/repo-config.toml` | The old name of the values file. The sync reads it and prints a message: rename it.              |

## The location of the file

The default path is `.config/bootstrap.toml`. `REPO_SYNC_VALUES` sets a
different path. Without the file, the sync stops with code 2.

## Planned keys

`packs`: a list of the packs of `tools/packs.toml`, for example
`["flutter-android", "flutter-ios"]`. `repo-sync` adds the tools of each pack
to `tools`. See [packs.md](packs.md). `repo-sync` does not read it yet.
