# Adopt repo-config in a repository

These instructions are for an AI agent that works in a consumer repository.
Do the steps in sequence to make the repository get its shared files from
`repo-config` (`git@github.com:virajp/repo-config.git`). Read `README.md` in
`repo-config` for the sync contract.

## Rules

- Ask the user before each choice that this file does not decide. Do not
  assume a tool, a value, or a path.
- The shared files are grouped by tool. Each tool is a folder
  `tools/<tool>/` in `repo-config`. A file at `tools/<tool>/<path>` goes to
  `<path>` in the consumer repository.
- The values file `.config/repo-config.toml` lists the tools of the
  repository. `repo_sync` syncs all the files of these tools.
- Do not edit a synced file in the consumer repository. The next sync replaces
  it. Put repository-owned content in the places that the sync never writes:
  - mise config: `.config/mise/conf.d/repo/mise.toml` and
    `.config/mise/conf.d/repo/mise.dev.toml`.
  - shell functions: `.config/mise/tasks/_scripts/local`.
  - ignore lines: `.config/repo-sync/gitignore`.
  - tasks: any task name that no tool ships.
- A seed file is different. The sync writes it only if it is missing. After
  that, the repository owns it.
- To change a shared file, change it in `repo-config`, push it to `main`, then
  sync again.
- Prefix each tool call with `mise x --`, also in the task files and the hook
  entries of `.config/pre-commit-config.yaml`.

## Requirements

- The machine has SSH access to `git@github.com:virajp/repo-config.git`. The
  sync uses `BatchMode=yes`, so it does not ask for a password.
- mise is installed. The `setup:configs` task gets gomplate through mise.

## Repositories with git submodules

Each git submodule is a separate repository. It needs its own values file and
its own copy of the shared files. Do the steps in each repository of the
tree, not only in the root repository.

1. Get the list of submodules, nested submodules included.

   ```bash
   git submodule foreach --quiet --recursive 'echo $displaypath'
   ```

2. Show the list to the user. Ask which submodules to include.
3. Do the steps in each included submodule. Start with the deepest
   submodules. Do the root repository last.
4. In each submodule, run the commands from the root of that submodule.
   `MISE_PROJECT_ROOT` and all the paths in this file are relative to it.

The tools and the values of a submodule can be different from those of the
root repository. Ask the user for each submodule. You can show the answers of
a different repository as a default.

A tool with `root_only = true` is for the root repository only. The sync
refuses it in a submodule. When a submodule is in its root checkout, mise
also loads the tools of the root. Thus a nested submodule still finds, for
example, graphify.

If a repository already has a `tools` list in `.config/repo-config.toml`,
tell the user. Then ask whether to do the steps again in that repository, or
to go to the next repository.

## Step 1: Select the tools

Make a shallow clone of `repo-config` into a temporary directory. Keep it
until all repositories are complete.

```bash
tmp=$(mktemp -d)
git clone --quiet --depth 1 git@github.com:virajp/repo-config.git "$tmp"
```

Read `"$tmp"/tools/*/tool.toml`. Show the user each tool with its
description. Then ask which tools the repository needs. `base` is necessary
in each repository.

Each `tool.toml` has these keys:

| Key          | Meaning                                                              |
| ------------ | -------------------------------------------------------------------- |
| `depends`    | The sync adds these tools automatically.                             |
| `root_only`  | If `true`, the sync refuses the tool in a submodule.                 |
| `extensions` | The VS Code extensions of the tool, for `.vscode/extensions.json`.   |
| `seed`       | Files that the sync writes only if they are missing.                 |

Ask which of `python`, `node` and `flutter` is the core language of the
repository. A core tool goes in `core_tools`. Then its
`conf.d/<tool>/mise.dev.toml` is written as `mise.toml`, and mise loads it in
every environment, CI included. A TypeScript repository has `node` as its
core tool.

## Step 2: Copy the sync files by hand

**CAUTION:** Do this before the first sync. An old copy of `repo-sync` reads
paths, not tools, and it cannot sync the new layout.

Copy these three files from the clone into the repository:

```bash
base="$tmp/tools/base/.config/mise/tasks"
mkdir -p .config/mise/tasks/_scripts .config/mise/tasks/setup
cp "$base/_scripts/repo-sync" "$base/_scripts/helpers" .config/mise/tasks/_scripts/
cp "$base/setup/configs" .config/mise/tasks/setup/
```

## Step 3: Write the values file

Write `.config/repo-config.toml`. Ask the user for each value.

| Key                                | Used by                     | Values                                                                                     |
| ---------------------------------- | --------------------------- | ------------------------------------------------------------------------------------------ |
| `tools`                            | `repo_sync`                 | The tools from step 1. `base` is necessary.                                                |
| `core_tools`                       | `repo_sync`                 | The core language tools from step 1, for example `["node"]`. The default is none.          |
| `repo_name`                        | `_env`                      | The repository name, for example `my-repo`. It is a literal, not the folder name of a worktree. |
| `merge_model_develop`              | `_env`                      | `direct` (local merge) or `pr` (pull request)                                              |
| `merge_model_main`                 | `_env`                      | `direct` or `pr`                                                                           |
| `commit_scopes`                    | `git-conventional-commits`  | The commit scopes, for example `["backend", "web"]`. The default is no scopes.             |
| `repo_url`                         | `git-conventional-commits`  | The web URL of the repository, for the changelog links. GitHub and GitLab are supported.   |
| `worktree_files`                   | `setup:worktree`            | The gitignored files to copy from the main worktree into a new worktree, for example `[".config/mise.local.toml"]` |
| `[member.<name>]`                  | `setup`, `setup:worktree`   | One table for each member folder                                                           |
| `[member.<name>].label`            | `setup`                     | The display name. The default is `<name>`.                                                 |
| `[member.<name>].worktree_install` | `setup:worktree`            | The install command for a new worktree. If it is not set, the member's own `setup:worktree` runs. |

Example:

```toml
tools = ["base", "node", "fnox", "claude", "vscode"]
core_tools = ["node"]
merge_model_develop = "direct"
merge_model_main = "pr"
repo_name = "my-repo"
commit_scopes = ["backend", "web"]
repo_url = "https://github.com/me/my-repo"
worktree_files = [".config/mise.local.toml"]
```

The member list is the git submodules, then each `[member.<name>]` table. A
`[member.<name>]` table that is not a submodule must have a folder with that
name in the repository. If not, the render fails.

## Step 4: Migrate the existing files

Do the "Migration" section below before the first sync. If the repository
has no mise config and no task files, skip this step.

## Step 5: Run the sync

1. If mise says that the config files are not trusted, ask the user to run
   `mise trust`.
2. Run `mise run setup:configs --dry-run`.
3. Examine the diff for each file that already exists in the repository.
4. If the dry run replaces repository-owned content, stop. Move that content
   to a repository-owned place (see "Rules"), then do the dry run again.
5. Run `mise run setup:configs`.
6. Run `mise run init`. This makes all the task scripts executable.

If `repo_sync` fails, read the exit code:

| Code | Cause                                                                          | Action                                          |
| ---- | ------------------------------------------------------------------------------ | ----------------------------------------------- |
| 2    | No `tools` in the values file, a root-only tool in a submodule, or a template that cannot render | Correct the values file          |
| 3    | A tool that is not in `repo-config`                                            | Correct the `tools` list                        |
| 5    | The clone failed                                                               | Examine the network, the SSH key and the access |
| 1    | A local file system error                                                      | Read the `repo-sync:` lines on stderr           |

A failure does not change an existing file.

## Step 6: Supply the files that `repo-config` does not have

The shared files refer to files that each repository must supply. Make sure
that the repository has each file for the tools that it uses. If a file is
missing, ask the user for its content. Do not write a placeholder that hides
the problem.

| File                             | Used by                                              |
| -------------------------------- | ---------------------------------------------------- |
| `.config/pre-commit-config.yaml` | `code:precommit`, `setup:precommit`, the merge tasks |
| `.config/gitleaks.toml`          | `code:sec`                                           |
| `.config/grype.yaml`             | `code:sec`                                           |
| `dprint.jsonc`                   | `code:format`. It only excludes paths, for example the submodules. The plugins are in the global dprint config. |
| `~/.config/dprint/dprint.jsonc`  | `setup:dprint`                                       |

The merge tasks also need the `develop` and `main` branches.

## Step 7: Make sure that the result is correct

1. Run `mise config ls`. Make sure that mise loads the synced `conf.d` files.
2. Run `mise tasks`. Make sure that the synced tasks are in the list.
3. Run `mise run setup`. Make sure that it completes.
4. Run `mise run setup:configs` again. Each file must show `unchanged` or
   `kept seed`.

mise loads `mise.toml` from each folder in `conf.d`. It loads a
`mise.dev.toml` file only when the `dev` config environment is active
(`MISE_ENV`).

## Step 8: Commit

Commit these files:

- The synced files and the seed files.
- `.config/repo-config.toml` and `.config/repo-sync/gitignore`.
- The repository-owned files from step 4 and step 6.

Commit in each submodule before you commit in the repository that contains
it. Then the commit in the containing repository includes the new submodule
commit. Ask the user before you push. Push each submodule before the
repository that contains it. If not, the containing repository points to a
commit that the remote does not have.

## Migration

This section is temporary. It is for the repositories that existed before the
`tools/` layout. Remove it when all repositories are migrated.

**WARNING:** Do not delete a file before you move its repository-owned
content. The sync never restores repository content.

The sync never deletes files. Thus you must delete each obsolete file and
rename each old task by hand. Do the items that apply to the repository, in
this sequence. Show the user the list of changes before you make them.

### M1: Delete the old sync files

- Delete `.config/gist-sync/` (an old `.gitignore` with only a header).
- If the repository uses `gist_sync` (macos-setup), delete
  `.config/mise/tasks/_scripts/gist-sync`. Rename
  `.config/gist-sync/gitignore` to `.config/repo-sync/gitignore`.
- Delete the old `.config/mise/tasks/setup/configs`. Step 2 replaces it.

### M2: Delete the old dependency tasks

**CAUTION:** Do this before the first `mise run setup`. The new `setup` runs
each task in `setup/deps/`. An old task there runs too, and it can fail.

- Delete `setup/deps/{all,install,upgrade,audit,cleanup,outdated,check,restart}`
  and `setup/pnpm/*`. The tool tasks `setup:deps:<tool>` replace them, with
  `--upgrade` for the upgrade.
- If a repository needs a dependency step that no tool ships, write it as
  `setup/deps/<name>`, with a `--upgrade` flag.
- Delete each `#PLACEHOLDER` stub in `setup/deps/` and `setup/external/`.

### M3: Rename the old tasks

| Old task                      | New task                     | Action                                                                                |
| ----------------------------- | ---------------------------- | ------------------------------------------------------------------------------------- |
| `setup:all` (`setup/all`)     | `setup` (`setup/_default`)   | Delete the old file. `base` ships the new one. Members are called with `setup`.       |
| `code:all` (`code/all`)       | `code` (`code/_default`)     | Delete the old file. Keep `code:check` and `code:test` as repository-owned tasks.     |
| `setup:secrets` (fnox)        | `setup:secrets:fnox`         | Delete the old file. Select the `fnox` tool.                                          |
| `setup:secrets` (Doppler)     | `setup:secrets:doppler`      | Move the file to `setup/secrets/doppler`. It stays repository-owned.                  |
| `setup:secrets` (stub)        | none                         | Delete it.                                                                            |
| `setup:deps:check`            | `setup:external:check`       | Delete it. The `docker` tool ships the new one.                                       |
| `setup:git-config`            | `code:git-config`            | Delete it.                                                                            |
| top-level `count`, `leaks`, `lint` | `code:*`                | Delete them.                                                                          |

Update each shell alias that calls an old name, for example
`setup = "mise run setup:all"` becomes `setup = "mise run setup"`.

### M4: Delete root-only items from submodules

In a submodule, delete `setup:ai`, `setup/external/*` and `conf.d/claude/`.
Keep `code:graph`: `base` ships it to each repository.

### M5: Move the content of `.config/mise.toml`

The sync replaces `.config/mise.toml`. It keeps only the shared `[settings]`.

1. Move `REPO_NAME`, `MEMBERS` and `MERGE_MODEL*` into the values file:
   - `REPO_NAME` becomes `repo_name`.
   - The old single `MERGE_MODEL` becomes both `merge_model_develop` and
     `merge_model_main`, with the same value.
   - `MEMBERS` is not necessary for submodules. For folders, write
     `[member.<name>]` tables.
2. Move repository-owned `[settings]`, `[env]`, `[tools]`, `[tasks]` and
   `[shell_alias]` to `.config/mise/conf.d/repo/mise.toml`. Examples are
   `lockfile = true`, `task.timings = true`, `CLOUDSDK_PYTHON` and
   `GITHUB_OWNER`.
3. Do the same for `.config/mise.dev.toml`, into
   `.config/mise/conf.d/repo/mise.dev.toml`. Examples are the emulator, Temporal
   and OTEL variables of 95octane backend.
4. Keep `.config/mise.ci.toml`, `.config/mise.test.toml` and
   `.config/miserc.toml`. The sync does not write them.

### M6: Remove the tools that the shared files now install

From the repository-owned mise files, remove each tool that a selected tool
installs:

- `base`: dprint, `npm:@askviraj/linter`, gitleaks, grype, osv-scanner,
  actionlint, shellcheck, shfmt, jq, yq, pre-commit.
- `python`: python, uv. `node`: node, pnpm. `fnox`: fnox.
- `claude`: `pipx:graphifyy`, `pipx:mempalace`.
- `flutter`: flutter, android-sdk, java. `bun`: bun.
- `rust`: `core:rust`, zola. `cloudflare`: `npm:cf`. `pulumi`: pulumi, gcloud.
- `kotlin`: `http:kotlin-lsp`.

Remove taplo too. dprint formats TOML now. Also remove the `[toml]` and
`evenBetterToml.*` keys from `.vscode/settings.json`.

Keep a tool that no selected tool installs, for example `npm:wrangler`,
`npm:sort-package-json`, `doppler` and `rtk`. Move it to
`conf.d/repo/`.

If the old tool has a pinned version that is different from the shared one,
tell the user. Examples: node `26` and pnpm `12` in 95octane, where `node`
ships `lts` and `latest`.

### M7: Split an old `conf.d/_base`

bootstrap and claude-plugins have an old `.config/mise/conf.d/_base/`. Its
name is the same as the shared `_base`, but the content is different. Before
the first sync, move its content:

| Old content in `_base`               | New place                                  |
| ------------------------------------ | ------------------------------------------ |
| node, pnpm                           | the `node` tool                            |
| python, uv, `UV_*`                   | the `python` tool                          |
| pre-commit, `PRE_COMMIT_HOME`        | the `base` tool                            |
| `MEMBERS`, `REPO_NAME`               | the values file                            |
| `MEMPALACE_*`, `min_version`, `task.run_auto_install`, other repository content | `conf.d/repo/` |

Then delete the old `_base/mise.toml`, `_base/mise.dev.toml` and
`_base/mise.ci.toml`. Also delete an old `conf.d/ai/`: the `claude` tool
replaces it. Move its repository-owned content to `conf.d/repo/` first.
Delete the old `conf.d/_formatter`, `_linter` and `_security`.

### M8: Move the fnox config

If the repository has `fnox.toml` at its root, move it to
`.config/fnox.toml`. Then the seed does not replace it. Change each
`fnox exec` call to `mise x -- fnox --config .config/fnox.toml exec`.

### M9: Clean the old markers

- Delete `.config/vscode.d/`. Remove the `// >>> vscode.d` blocks from
  `.vscode/settings.json` and `.config/pre-commit-config.yaml`. The
  `vscode` tool writes `.vscode/extensions.json` from the tools.
- Move the repository-owned keys of `.vscode/settings.json` to the user
  settings or to the workspace, and tell the user. The sync replaces the file.
  Examples are `[dart]`, `files.associations` and `markdownlint.*`.
- If an extension is not in a tool, for example astro, tailwind or effect,
  tell the user. The sync replaces `.vscode/extensions.json`.
- Replace each `MARKED POSITION` block with its value, in a repository-owned
  file or in the values file.

### M10: Change the old helper files

- The shared helpers are `_scripts/helpers`, `_scripts/checks` and
  `_scripts/merge`. Change each `source` of `_helpers`, `_checks` or `_merge`
  to the new name. Then delete the old files.
- If a task calls a function that `helpers` does not have, for example
  `print_ok`, put that function in `_scripts/local` and source it from the
  task.

### M11: The commit config

The template has 10 types: `feat`, `fix`, `perf`, `refactor`, `revert`,
`test`, `ops`, `docs`, `merge`, `wip`. Other types are not used. The old types
`spec`, `security` and `blueprint` are dropped. Tell the user. Put the scopes
of the old file in `commit_scopes`.

### M12: The pre-commit hooks

In `.config/pre-commit-config.yaml`, prefix each `entry` that calls a tool
with `mise x --`. The `graphify-refresh` hook must call
`mise x -- mise run code:graph`. Keep the hook in each repository, submodules
included.
