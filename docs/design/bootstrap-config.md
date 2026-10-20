# `.config/bootstrap.toml` design

This document describes each key of the values file `.config/bootstrap.toml`.
The adopting AI writes the file in a consumer repository (step 3 of
`docs/sync/index.md`). `repo-sync` and the templates of `tools/` read it. The
content is reverse engineered on 2026-10-20 from `tools/` and from
`tools/base/.config/mise/tasks/_scripts/repo-sync`.

## Example

```toml
commit_scopes = [ "backend", "web" ]
environments = [ "dev", "ci", "prod" ]
extensions = [ "astro-build.astro-vscode" ]
merge_model_develop = "direct"
merge_model_main = "pr"
product_name = "my-product"
release_tag_pattern = '^my-repo-v\d+$'
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"
tools = [ "base", "github", "node", "fnox", "claude", "vscode" ]
worktree_files = [ ".config/mise.local.toml" ]

[member.web]
label = "Web"
worktree_install = "pnpm install"
```

## Summary of the keys

| Key                                | Type            | Necessary | Read by                                        |
| ---------------------------------- | --------------- | --------- | ---------------------------------------------- |
| `tools`                            | List of strings | Yes       | `repo-sync`                                    |
| `environments`                     | List of strings | No        | `repo-sync`                                    |
| `repo_name`                        | String          | Yes       | `mise.toml`, `gitleaks.toml`, `fnox.toml`      |
| `product_name`                     | String          | No        | `fnox.toml`                                    |
| `merge_model_develop`              | String          | Yes       | `mise.toml`                                    |
| `merge_model_main`                 | String          | Yes       | `mise.toml`                                    |
| `commit_scopes`                    | List of strings | No        | `git-conventional-commits.yaml`                |
| `repo_url`                         | String          | No        | `git-conventional-commits.yaml`, `github` tool |
| `release_tag_pattern`              | String          | No        | `git-conventional-commits.yaml`                |
| `worktree_files`                   | List of strings | No        | `setup:worktree`                               |
| `extensions`                       | List of strings | No        | `vscode` tool                                  |
| `[member.<name>]`                  | Table           | No        | `repo-sync`, `setup`, `setup:worktree`         |
| `[member.<name>].label`            | String          | No        | `setup`, `setup:worktree`                      |
| `[member.<name>].worktree_install` | String          | No        | `setup:worktree`                               |

All the paths in this document are relative to the root of the consumer
repository.

## The keys

### `tools`

The tools of the repository. Each name is a folder `tools/<name>/` of
`bootstrap`. `base` is necessary in each repository.

- `repo-sync` adds the `depends` of each tool, and the tool of each mise
  `depends`, depth first.
- A tool that `tools` lists gets its mise tools in the environments of their
  `envs`. A tool that only a dependency adds gets its mise tools in `dev`
  only. Thus list the core language of the repository, for example `node`.
- Without the key, or with an empty list, the sync stops with code 2.
- A name that is not in `bootstrap` stops the sync with code 3.
- A tool with `root_only = true` in a git submodule stops the sync with
  code 2.

### `environments`

The mise environments of the repository, from `dev`, `ci`, `prod` and `stage`.

- `dev` is necessary. `prod` and `stage` need `ci`.
- Without the key, the sync uses `dev` only and prints a message.
- An unknown name, a list without `dev`, or `prod` or `stage` without `ci`
  stops the sync with code 2.
- The sync writes `conf.d/<folder>/mise.<env>.toml` only for these
  environments. See [mise.md](mise.md).

### `repo_name`

The name of the repository, for example `my-repo`. It is a literal, not the
folder name of a worktree.

| Output                  | Use                                                                  |
| ----------------------- | -------------------------------------------------------------------- |
| `.config/mise.toml`     | `REPO_NAME` in `[env]`                                               |
| `.config/gitleaks.toml` | The `title` of the gitleaks config (a seed)                          |
| `.config/fnox.toml`     | The keychain `service`, or the `prefix` with `product_name` (a seed) |

The `setup:vscode` task uses `REPO_NAME` as the name of the VS Code profile.

### `product_name`

The name of the product, for example `95octane`. Set it in each repository of
a product with members, the root included.

- With the key, the fnox keychain provider uses `service = "<product_name>"`
  and `prefix = "<repo_name>/"`.
- Without the key, it uses `service = "<repo_name>"` and no prefix.

### `merge_model_develop`

How `code:merge:develop` lands a feature branch on `develop`: `direct` or
`pr`. The template writes it to `MERGE_MODEL_DEVELOP` in `.config/mise.toml`.

| Value    | Effect                                                         |
| -------- | -------------------------------------------------------------- |
| `direct` | A local `--no-ff` merge, then a push.                          |
| `pr`     | A push of the branch, then a pull request with `gh` or `glab`. |

The merge script uses `direct` when the environment value is empty.

### `merge_model_main`

How `code:merge:main` lands `develop` on `main`: `direct` or `pr`. The
template writes it to `MERGE_MODEL_MAIN` in `.config/mise.toml`. The values
are the same as for `merge_model_develop`.

A repository with only one of `main` and `develop` uses `code:merge`. See
"Branch model in the merge tasks" in `CLAUDE.md`.

### `commit_scopes`

The commit scopes of `.config/git-conventional-commits.yaml`, for example
`["backend", "web"]`.

- Without the key, or with an empty list, the scopes are the members of the
  repository.
- Without members, the list of scopes is empty.

### `repo_url`

The web URL of the repository, for example `https://github.com/me/my-repo`.

| Output                                  | Use                                                     |
| --------------------------------------- | ------------------------------------------------------- |
| `.config/git-conventional-commits.yaml` | The commit, compare and issue links of the changelog.   |
| `.github/ISSUE_TEMPLATE/config.yml`     | The links to the README and to the security advisories. |

If the URL contains `gitlab`, the changelog links use the `/-/` path of
GitLab. Without the key, the templates write no links.

### `release_tag_pattern`

The pattern of the release tags, for example `'^web-v\d+$'`. The template
writes it to `releaseTagGlobPattern` of `.config/git-conventional-commits.yaml`.

- Use a TOML literal string (single quotes), so that each backslash stays.
- Without the key, the value is `"*"`: all tags.

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

- A `[member.<name>]` table that is not a submodule must have a folder with
  that name in the repository. If not, the render stops with code 2.
- The templates get the list as `.members`. `MEMBERS` in `.config/mise.toml`
  is this list, with one space between the names.
- `setup`, `code:format` and `code:lint` get a `--<name>` flag and an `--all`
  flag for the members.

### `[member.<name>].label`

The display name of the member in the headers of `setup` and
`setup:worktree`. The default is `<name>`.

### `[member.<name>].worktree_install`

The install command of the member for a new worktree, for example
`pnpm install`. `setup:worktree` runs it as
`mise x --cd <name> -- <command>`. Without the key, `setup:worktree` runs the
own `setup:worktree` task of the member, if the member has one.

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

The design in [tools.md](tools.md) adds the answers of the questions of each
`options.toml`, for example `[options.flutter] platforms = ["android"]`. The
format is not decided, and `repo-sync` does not read it yet.
