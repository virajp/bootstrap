# The `base` tool

The files that each repository gets: mise, the task library, the formatter,
the linters, the scanners and pre-commit.

| Property  | Value                                                                  |
| --------- | ---------------------------------------------------------------------- |
| Root only | No                                                                     |
| Depends   | None. Its mise tools add `node` and `python` (see "mise tools").       |
| Seeds     | `gitleaks.toml`, `grype.yaml`, `osv-scanner.toml`, `_repo/mise.*.toml` |

## When the tool is relevant

Always. Each repository selects `base`.

## What the tool ships

| Path in the repository                           | Kind           | Use                                                                                                |
| ------------------------------------------------ | -------------- | -------------------------------------------------------------------------------------------------- |
| `.config/mise.toml`                              | Template       | The shared `[settings]`, and the `[env]` values from the values file.                              |
| `.config/mise/conf.d/_base/mise.<env>.toml`      | Generated      | The mise config of this tool, one file for each environment.                                       |
| `.config/mise/conf.d/_repo/mise.<env>.toml`      | Seed           | The repository's own mise config and `[env]` values, one file for each environment.                |
| `.config/mise/tasks/_scripts/`                   | Synced         | The libraries `helpers`, `checks`, `merge`, `placeholder` and `repo-sync`.                         |
| `.config/mise/tasks/code/`, `setup/`             | Synced         | The tasks in "Tasks".                                                                              |
| `.config/pre-commit-config.yaml`                 | Template       | The hooks of the selected tools, then `.config/repo-sync/pre-commit.yaml`.                         |
| `.config/git-conventional-commits.yaml`          | Template       | The commit types, the scopes from `commit_scopes`, and the tag pattern from `release_tag_pattern`. |
| `.config/gitleaks.toml`                          | Seed, template | The gitleaks rules.                                                                                |
| `.config/grype.yaml`, `.config/osv-scanner.toml` | Seed           | The config of the vulnerability scanners.                                                          |
| `.gitignore`                                     | Synced         | The shared ignore lines, then the lines of `.config/repo-sync/gitignore`.                          |

## mise tools

| mise tool              | Version  | Environments | Note                  |
| ---------------------- | -------- | ------------ | --------------------- |
| `gitleaks`             | `latest` | dev, ci      | Secret scanner        |
| `grype`                | `latest` | dev, ci      | Vulnerability scanner |
| `osv-scanner`          | `latest` | dev, ci      | Vulnerability scanner |
| `dprint`               | `latest` | dev, ci      | Formatter             |
| `npm:@askviraj/linter` | `latest` | dev, ci      | Linter. Needs node.   |
| `shellcheck`           | `latest` | dev, ci      | Shell linter          |
| `gomplate`             | `latest` | dev          | Renders the templates |
| `jq`, `yq`             | `latest` | dev          | JSON and YAML queries |
| `pre-commit`           | `latest` | dev          | Hooks. Needs python.  |

The tool also sets:

- `PRE_COMMIT_HOME = "$HOME/.cache/pre-commit"`.
- The shell aliases `precommit` and `setup`.
- The hidden task `init`, which makes the task scripts executable.
- `os = "linux"` in `ci`, `prod` and `stage`.

## Tasks

| Task                          | Use                                                                            |
| ----------------------------- | ------------------------------------------------------------------------------ |
| `setup`                       | Install tools, dependencies and hooks. `--upgrade` also upgrades them.         |
| `setup:configs`               | Sync the shared files from `bootstrap`. `--dry-run` shows the diff.            |
| `setup:mise`                  | Reshim, doctor and install. `--upgrade` also upgrades.                         |
| `setup:precommit`             | Install the hooks. `--upgrade` also updates them.                              |
| `setup:dprint`                | Upgrade the dprint plugins.                                                    |
| `setup:linter`                | Write the linter config files, if they are missing.                            |
| `setup:worktree`              | Initialize a new git worktree. It is lighter than `setup`.                     |
| `code`                        | Format, lint and scan all code.                                                |
| `code:format`, `code:lint`    | Run each `code:format:<name>` or `code:lint:<name>`. `--fix` applies fixes.    |
| `code:sec`                    | Run grype, osv-scanner and gitleaks. `--staged` scans only the staged changes. |
| `code:precommit`              | Run the hooks on all files.                                                    |
| `code:git-config`             | Find forbidden keys in the local git config. `--fix` removes them.             |
| `code:graph`                  | Refresh the graphify graph. Only the post-commit hook runs it.                 |
| `code:worktrees`              | List the worktrees of the root repository and of each submodule.               |
| `code:merge <branch>`         | Land a branch on the only long-lived branch.                                   |
| `code:merge:develop <branch>` | Land a branch on `develop`.                                                    |
| `code:merge:main`             | Land `develop` on `main`.                                                      |

The sync writes the merge tasks of the branch model that it finds (step 4 of
`index.md`). `MERGE_MODEL_DEVELOP` and `MERGE_MODEL_MAIN` select `direct` or
`pr`.

## Pre-commit hooks

- Local hooks that call a task: `git-config`, `format`, `lint`, `sec`, and
  `graphify-refresh` (post-commit stage).
- `pre-commit/pre-commit-hooks`: large files, case conflicts, Windows names,
  shebangs, merge conflicts, symlinks, JSON, TOML and YAML syntax, private
  keys, end of file, trailing whitespace, line endings, and
  `no-commit-to-branch` for `main` and `develop`.
- `qoomon/git-conventional-commits` in the commit-msg stage.

## Files the repository supplies

| File                            | Used by                                                                                                                                               |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `dprint.jsonc`                  | `code:format`. It must have `"extends": "~/.config/dprint/dprint.jsonc"`, which has the plugins. It only adds `excludes`, for example the submodules. |
| `~/.config/dprint/dprint.jsonc` | `setup:dprint`                                                                                                                                        |

## Adapt to the repository

1. Write the values file (step 3 of `index.md`) and select the branch model
   (step 4).
2. Examine an existing `.config/gitleaks.toml`. A seed keeps an existing
   file. Without `[extend] useDefault = true`, gitleaks uses only the rules of
   the file. Tell the user.
3. Move each repository hook that no tool ships to
   `.config/repo-sync/pre-commit.yaml`.
4. Move each repository formatter or linter step to a task
   `code:format:<name>` or `code:lint:<name>`.
5. Move each repository ignore line to `.config/repo-sync/gitignore`.
6. Move each repository shell function to `.config/mise/tasks/_scripts/local`.

## Migration

- Remove these old mise tools from the repository-owned mise files: dprint,
  `npm:@askviraj/linter`, gitleaks, grype, osv-scanner, shellcheck, gomplate,
  jq, yq and pre-commit.
- Remove shfmt. dprint formats shell scripts now.
- Remove taplo. dprint formats TOML now. Also remove the `[toml]` and
  `evenBetterToml.*` keys from `.vscode/settings.json`.
- The general items M1 to M5, M10 to M12 and M14 of `migrate.md` are about
  the files of this tool.
