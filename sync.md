# Adopt repo-config in a repository

These instructions are for an AI agent that works in a consumer repository.
Do the steps in sequence to make the repository get its shared files from
`repo-config` (`git@github.com:virajp/repo-config.git`). Read `README.md` in
`repo-config` for the sync contract.

## Rules

- Ask the user before each choice that this file does not decide. Do not
  assume a stack, a value, or a path.
- Each shared file is at `files/<path>` in `repo-config`. In the consumer
  repository, it is at `<path>`. Give `repo_sync` the `<path>` only, without
  `files/`.
- Do not edit a synced file in the consumer repository. The next sync replaces
  it. Put a function that only one repository uses in a local sibling file,
  for example `.config/mise/tasks/_scripts/<name>.env`.
- To change a shared file, change it in `repo-config`, push it to `main`, then
  sync again.

## Requirements

- The machine has SSH access to `git@github.com:virajp/repo-config.git`. The
  sync uses `BatchMode=yes`, so it does not ask for a password.
- mise is installed. The `setup:configs` task gets gomplate through mise.

## Step 1: Select the files

Show the user this catalog. Then ask which groups the repository needs.

| Group                  | Paths                                                                                                                                                                 |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Sync and task library  | `.config/mise/tasks/_scripts/repo-sync`, `.config/mise/tasks/_scripts/helpers`, `.config/mise/tasks/_scripts/checks`, `.config/mise/tasks/_scripts/placeholder`        |
| Base config            | `.gitignore`, `.config/mise.toml`, `.config/mise/conf.d/_env/mise.toml.tmpl`, `.config/mise/conf.d/_task/mise.dev.toml`                                               |
| Code quality           | `.config/mise/conf.d/_formatter/mise.dev.toml`, `.config/mise/conf.d/_linter/mise.dev.toml`, `.config/mise/conf.d/_security/mise.toml`                                |
| Code tasks             | `.config/mise/tasks/code/_default`, `.config/mise/tasks/code/format`, `.config/mise/tasks/code/lint`, `.config/mise/tasks/code/sec`, `.config/mise/tasks/code/git-config`, `.config/mise/tasks/code/worktrees` |
| Merge tasks            | `.config/mise/tasks/_scripts/merge`, `.config/mise/tasks/code/merge/develop`, `.config/mise/tasks/code/merge/main`                                                    |
| Setup tasks            | `.config/mise/tasks/setup/_default.tmpl`, `.config/mise/tasks/setup/mise`, `.config/mise/tasks/setup/dprint`, `.config/mise/tasks/setup/vscode`, `.config/mise/tasks/setup/worktree.tmpl` |
| Node                   | `.config/mise/conf.d/node/mise.toml`                                                                                                                                  |
| Python                 | `.config/mise/conf.d/python/mise.toml`                                                                                                                                |
| pre-commit             | `.config/mise/conf.d/pre-commit/mise.dev.toml`, `.config/mise/tasks/code/precommit`, `.config/mise/tasks/setup/precommit`                                             |
| AI tools               | `.config/mise/conf.d/claude/mise.dev.toml`                                                                                                                            |
| graphify               | `.config/mise/tasks/code/graph`                                                                                                                                       |
| VS Code                | `.vscode/settings.json`                                                                                                                                               |

Before you ask, read the current file list of `repo-config` (step 2 makes a
clone). If the catalog and the clone are different, the clone is correct. Tell
the user about the difference.

The "Sync and task library" group is necessary. Every task sources `helpers`,
and `setup:configs` sources `repo-sync`.

Some groups need other groups:

- "Merge tasks" needs "Sync and task library" and "Base config". The merge
  reads `MERGE_MODEL_DEVELOP` and `MERGE_MODEL_MAIN` from `_env`.
- "pre-commit", "AI tools" and "graphify" need "Python", because their tools
  use `python` and `uv`.
- "Setup tasks" runs `pnpm dlx`, so it needs "Node".

## Step 2: Copy the sync script by hand

The sync script cannot sync itself before it exists. A copy from before the
`files/` layout also cannot sync, because it reads from the root of
`repo-config`.

1. Make a shallow clone of `repo-config` into a temporary directory.

   ```bash
   tmp=$(mktemp -d)
   git clone --quiet --depth 1 git@github.com:virajp/repo-config.git "$tmp"
   ```

2. Copy the two library files into the repository.

   ```bash
   mkdir -p .config/mise/tasks/_scripts
   cp "$tmp/files/.config/mise/tasks/_scripts/repo-sync" .config/mise/tasks/_scripts/
   cp "$tmp/files/.config/mise/tasks/_scripts/helpers" .config/mise/tasks/_scripts/
   ```

3. Keep the clone until step 1 is complete. Then remove it with
   `rm -rf "$tmp"`.

The template render needs `helpers`, because it calls `members` from it. Thus
copy `helpers` now, also if the first sync is a dry run.

## Step 3: Write the values file

**CAUTION:** Do not omit a key that a selected template uses. The render
fails, and the sync keeps the old file.

Write `.config/repo-config.toml` if the selected paths include a `.tmpl` file.
Ask the user for each value.

| Key                           | Used by                     | Values                                       |
| ----------------------------- | --------------------------- | -------------------------------------------- |
| `repo_name`                   | `_env`                      | The repository name, for example `my-repo`   |
| `merge_model_develop`         | `_env`                      | `direct` (local merge) or `pr` (pull request) |
| `merge_model_main`            | `_env`                      | `direct` or `pr`                             |
| `[member.<name>]`             | `setup/_default`, `setup/worktree` | One table for each member folder     |
| `[member.<name>].label`       | `setup/_default`            | The display name. The default is `<name>`.   |
| `[member.<name>].worktree_install` | `setup/worktree`       | The install command for a new worktree       |

Example:

```toml
merge_model_develop = "direct"
merge_model_main = "pr"
repo_name = "my-repo"

[member.backend]
label = "Backend"
worktree_install = "pnpm install --frozen-lockfile"
```

The member list is the git submodules, then each `[member.<name>]` table. A
`[member.<name>]` table that is not a submodule must have a folder with that
name in the repository. If not, the render fails.

## Step 4: Add repository-only ignore lines

If the repository needs ignore lines that the shared `.gitignore` does not
have, write them to `.config/repo-sync/gitignore`. The sync adds them after
the shared lines. A later line wins, so a `!<pattern>` line negates a shared
pattern.

Do not edit `.gitignore` directly. The next sync replaces it.

## Step 5: Write the `setup:configs` task

Write `.config/mise/tasks/setup/configs`. List the paths from step 1. Put
`_scripts/helpers` before all `.tmpl` paths.

```bash
#!/usr/bin/env bash

#MISE description="Sync the shared files from repo-config"
#MISE tools.gomplate="latest"

#USAGE flag "--dry-run" help="show the changes without writing them"

set -euo pipefail
# shellcheck source=/dev/null
source "${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/repo-sync"

FLAGS=()
[[ "${usage_dry_run:-false}" == "true" ]] && FLAGS+=(--dry-run)

repo_sync ${FLAGS[@]+"${FLAGS[@]}"} -- \
  .config/mise/tasks/_scripts/repo-sync \
  .config/mise/tasks/_scripts/helpers \
  .gitignore \
  .config/mise.toml \
  .config/mise/conf.d/_env/mise.toml.tmpl
```

Write the scripts for bash 3.2. macOS has bash 3.2 as `/bin/bash`. The
`${FLAGS[@]+...}` form prevents an "unbound variable" error for an empty array
in bash 3.2.

Make the task executable with `chmod 755 .config/mise/tasks/setup/configs`.

## Step 6: Run the sync

1. If mise says that the config files are not trusted, ask the user to run
   `mise trust`.
2. Run `mise run setup:configs --dry-run`.
3. Examine the diff for each file that already exists in the repository.
4. If the dry run replaces repository-only content, stop and ask the user.
5. Run `mise run setup:configs`.
6. Run `mise run init`. This makes all the task scripts executable.

If `repo_sync` fails, read the exit code:

| Code | Cause                                                        | Action                                         |
| ---- | ------------------------------------------------------------ | ---------------------------------------------- |
| 2    | A usage error, or a template that cannot render              | Examine the values file and the gomplate output |
| 3    | A path that is not in `files/` of `repo-config`              | Correct the path in the task                   |
| 5    | The clone failed                                             | Examine the network, the SSH key and the access |
| 1    | A local file system error                                    | Read the `repo-sync:` lines on stderr          |

A failure does not change an existing file.

## Step 7: Supply the files that `repo-config` does not have

The shared files refer to files that each repository must supply. Make sure
that the repository has each file for the groups that it uses. If a file is
missing, ask the user for its content. Do not write a placeholder that hides
the problem.

| File or task                        | Used by                                             |
| ----------------------------------- | --------------------------------------------------- |
| `.config/pre-commit-config.yaml`    | `code:precommit`, `setup:precommit`, the merge tasks |
| `.config/gitleaks.toml`             | `code:sec`                                          |
| `.config/grype.yaml`                | `code:sec`                                          |
| `~/.config/dprint/dprint.jsonc`     | `setup:dprint`                                      |
| `.vscode/extensions.json`           | `setup:vscode`                                      |
| `setup:external:update` task        | `setup` (`setup/_default`)                          |
| `setup:ai` task                     | `setup` (`setup/_default`)                          |

The merge tasks also need the `develop` and `main` branches.

## Step 8: Make sure that the result is correct

1. Run `mise config ls`. Make sure that mise loads the synced `conf.d` files.
2. Run `mise tasks`. Make sure that the synced tasks are in the list.
3. Run `mise run setup:configs` again. Each file must show `unchanged`.

mise loads `mise.toml` from each folder in `conf.d`. It loads a
`mise.dev.toml` file only when the `dev` config environment is active
(`MISE_ENV`).

## Step 9: Commit

Commit these files:

- The synced files.
- `.config/repo-config.toml`, `.config/repo-sync/gitignore`, and
  `.config/mise/tasks/setup/configs`.
- The files from step 7.
