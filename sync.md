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
  - mise config and environment variables: the seeds
    `.config/mise/conf.d/_repo/mise.<env>.toml`, one for each environment.
  - shell functions: `.config/mise/tasks/_scripts/local`.
  - ignore lines: `.config/repo-sync/gitignore`.
  - pre-commit hooks: `.config/repo-sync/pre-commit.yaml`, a YAML list of
    `repos:` entries.
  - tasks: any task name that no tool ships.
- A seed file is different. The sync writes it only if it is missing. After
  that, the repository owns it.
- To change a shared file, change it in `repo-config`, push it to `main`, then
  sync again.
- Prefix each tool call with `mise x --`, also in the task files and the hook
  entries of `.config/repo-sync/pre-commit.yaml`.

## Requirements

- The machine has SSH access to `git@github.com:virajp/repo-config.git`. The
  sync uses `BatchMode=yes`, so it does not ask for a password.
- mise is installed. The `setup:configs:sync` task gets gomplate through mise.

## An empty repository

If the repository has no commit, do these steps first:

1. Write `.gitignore` with these two lines:

   ```text
   .worktrees/
   .claude/worktrees/
   ```

2. Commit it on `main`. This is the first commit.

Then do the other steps. A worktree in `.worktrees/` is now ignored, and the
branch `main` exists. The first sync replaces `.gitignore` with the shared one,
which has the same two lines.

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

| Key          | Meaning                                                            |
| ------------ | ------------------------------------------------------------------ |
| `depends`    | The sync adds these tools automatically.                           |
| `root_only`  | If `true`, the sync refuses the tool in a submodule.               |
| `extensions` | The VS Code extensions of the tool, for `.vscode/extensions.json`. |
| `seed`       | Files that the sync writes only if they are missing.               |

A `tool.toml` can also have a `[seed_requires]` table. It is optional. Each
key is a seed, and its value is a file of the repository. The sync writes the
seed only if the repository has that file. For example, the `node` tool writes
`pnpm-workspace.yaml` only if the repository has a `package.json`.

The `[mise.*]` tables of a `tool.toml` hold the mise config of the tool. The
sync generates `.config/mise/conf.d/<tool>/mise.<env>.toml` from them, one
file for each environment. A mise `depends` adds the tool that defines the
dependency. For example, `base` adds `python` and `node`, because pre-commit
and the linter need them.

A tool that you list in `tools` is installed in the environments of its
`envs`. A tool that only a dependency adds is installed in dev only. Thus list
the core language of the repository in `tools`. A TypeScript repository lists
`node`, which then goes to every selected environment.

Ask which mise environments the repository needs. Offer only these:

| Environment | Use                                                                                         |
| ----------- | ------------------------------------------------------------------------------------------- |
| `dev`       | Local development. It usually has everything. Necessary.                                    |
| `ci`        | The CI environment, to build and examine the code. Optional.                                |
| `prod`      | Production. Only the tools that are necessary to run the application. Optional, needs `ci`. |
| `stage`     | Staging. The same tools as `prod`, with different `[env]` values. Optional, needs `ci`.     |

The sync stops if `dev` is missing, or if `prod` or `stage` is selected
without `ci`.

Ask whether the repository is on GitHub or on GitLab. Select `github` or
`gitlab` for its forge. The `pr` merge model needs the CLI of the forge: `gh`
or `glab`.

## Step 2: Copy the sync files by hand

**CAUTION:** Do this before the first sync. An old copy of `repo-sync` reads
paths, not tools, and it cannot sync the new layout.

Copy these four files from the clone into the repository. An old
`.config/mise/tasks/setup/configs` is a file, and `setup/configs/` is now a
folder: delete the old file first.

```bash
base="$tmp/tools/base/.config/mise/tasks"
rm -f .config/mise/tasks/setup/configs
mkdir -p .config/mise/tasks/_scripts .config/mise/tasks/setup/configs
cp "$base/_scripts/repo-sync" "$base/_scripts/helpers" .config/mise/tasks/_scripts/
cp "$base/setup/configs/_default" "$base/setup/configs/sync" .config/mise/tasks/setup/configs/
```

## Step 3: Write the values file

Write `.config/repo-config.toml`. Ask the user for each value.

| Key                                | Used by                    | Values                                                                                                             |
| ---------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `tools`                            | `repo_sync`                | The tools from step 1. `base` is necessary.                                                                        |
| `environments`                     | `repo_sync`                | The environments from step 1, for example `["dev", "ci"]`. The default is `["dev"]`.                               |
| `repo_name`                        | `.config/mise.toml`        | The repository name, for example `my-repo`. It is a literal, not the folder name of a worktree.                    |
| `merge_model_develop`              | `.config/mise.toml`        | `direct` (local merge) or `pr` (pull request)                                                                      |
| `merge_model_main`                 | `.config/mise.toml`        | `direct` or `pr`                                                                                                   |
| `commit_scopes`                    | `git-conventional-commits` | The commit scopes, for example `["backend", "web"]`. The default is the member list.                               |
| `repo_url`                         | `git-conventional-commits` | The web URL of the repository, for the changelog links. GitHub and GitLab are supported.                           |
| `worktree_files`                   | `setup:worktree`           | The gitignored files to copy from the main worktree into a new worktree, for example `[".config/mise.local.toml"]` |
| `[member.<name>]`                  | `setup`, `setup:worktree`  | One table for each member folder                                                                                   |
| `[member.<name>].label`            | `setup`                    | The display name. The default is `<name>`.                                                                         |
| `[member.<name>].worktree_install` | `setup:worktree`           | The install command for a new worktree. If it is not set, the member's own `setup:worktree` runs.                  |

Example:

```toml
commit_scopes = [ "backend", "web" ]
environments = [ "dev", "ci", "prod" ]
merge_model_develop = "direct"
merge_model_main = "pr"
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"
tools = [ "base", "github", "node", "fnox", "claude", "vscode" ]
worktree_files = [ ".config/mise.local.toml" ]
```

The member list is the git submodules, then each `[member.<name>]` table. A
`[member.<name>]` table that is not a submodule must have a folder with that
name in the repository. If not, the render fails.

## Step 4: Select the branch model

The merge tasks follow the branches `main` and `develop`. A branch counts if it
exists locally or on `origin`.

| Branches             | Merge tasks                             | Rule                                                                      |
| -------------------- | --------------------------------------- | ------------------------------------------------------------------------- |
| `main` and `develop` | `code:merge:develop`, `code:merge:main` | Feature branches merge into `develop`. Only `develop` merges into `main`. |
| only one of them     | `code:merge`                            | Any feature branch merges into that branch.                               |
| neither              | none                                    | The sync stops with code 2.                                               |

1. Find the branches:

   ```bash
   for b in main develop; do
     git show-ref --verify --quiet "refs/heads/$b" ||
       git show-ref --verify --quiet "refs/remotes/origin/$b" && echo "$b"
   done
   ```

2. Show the result to the user.
3. If a branch is missing, ask the user whether to create it. If the user
   agrees, create it from the existing branch, and ask before you push it.

`repo_sync` does not ask. It writes the merge tasks of the model that it
finds, and removes the merge tasks of the other model. If the branches change
later, run `setup:configs` again.

## Step 5: Migrate the existing files

Do the "Migration" section below before the first sync. If the repository
has no mise config and no task files, skip this step.

## Step 6: Run the sync

1. If mise says that the config files are not trusted, ask the user to run
   `mise trust`.
2. Run `mise run setup:configs --dry-run`.
3. Examine the diff for each file that already exists in the repository.
4. If the dry run replaces repository-owned content, stop. Move that content
   to a repository-owned place (see "Rules"), then do the dry run again.
5. Run `mise run setup:configs`.
6. Run `mise run init`. This makes all the task scripts executable.

If `repo_sync` fails, read the exit code:

| Code | Cause                                                                                            | Action                                          |
| ---- | ------------------------------------------------------------------------------------------------ | ----------------------------------------------- |
| 2    | No `tools` in the values file, a root-only tool in a submodule, or a template that cannot render | Correct the values file                         |
| 2    | Neither `main` nor `develop` exists                                                              | Do step 4                                       |
| 3    | A tool that is not in `repo-config`                                                              | Correct the `tools` list                        |
| 5    | The clone failed, or `pre-commit autoupdate` failed                                              | Examine the network, the SSH key and the access |
| 1    | A local file system error                                                                        | Read the `repo-sync:` lines on stderr           |

A failure does not change an existing file.

## Step 7: Supply the files that `repo-config` does not have

The shared files refer to files that each repository must supply. Make sure
that the repository has each file for the tools that it uses. If a file is
missing, ask the user for its content. Do not write a placeholder that hides
the problem.

| File                            | Used by                                                                                                                                               |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `dprint.jsonc`                  | `code:format`. It must have `"extends": "~/.config/dprint/dprint.jsonc"`, which has the plugins. It only adds `excludes`, for example the submodules. |
| `~/.config/dprint/dprint.jsonc` | `setup:dprint`                                                                                                                                        |

`base` supplies `.config/pre-commit-config.yaml` (synced), and
`.config/gitleaks.toml`, `.config/grype.yaml` and `.config/osv-scanner.toml`
(seeds). A seed keeps an existing file. Examine an existing `gitleaks.toml`:
without `[extend] useDefault = true`, gitleaks uses only the rules of the file.

## Step 8: Make sure that the result is correct

1. For each selected environment, run `MISE_ENV=<env> mise config ls`. Make
   sure that mise loads the `conf.d/*/mise.<env>.toml` files of that
   environment.
2. Run `mise tasks`. Make sure that the synced tasks are in the list.
3. Run `mise run setup`. Make sure that it completes.
4. Run `mise run setup:configs` again. Each file must show `unchanged`,
   `kept seed` or `skip env`.

mise always loads `.config/mise.toml`. It loads a `conf.d/*/mise.<env>.toml`
file only when `MISE_ENV` includes that environment. Without `MISE_ENV`, no
file in `conf.d` loads.

## Step 9: Commit

Commit these files:

- The synced files and the seed files.
- `.config/repo-config.toml`, `.config/repo-sync/gitignore` and
  `.config/repo-sync/pre-commit.yaml`.
- The repository-owned files from step 5 and step 7.

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

The sync deletes only two kinds of files: the merge tasks of the other branch
model, and a `.config/mise/conf.d/*/mise*.toml` with the repo-config header
that it did not write. Thus you must delete each other obsolete file and
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

| Old task                           | New task                   | Action                                                                            |
| ---------------------------------- | -------------------------- | --------------------------------------------------------------------------------- |
| `setup:all` (`setup/all`)          | `setup` (`setup/_default`) | Delete the old file. `base` ships the new one. Members are called with `setup`.   |
| `code:all` (`code/all`)            | `code` (`code/_default`)   | Delete the old file. Keep `code:check` and `code:test` as repository-owned tasks. |
| `setup:secrets` (fnox)             | `setup:secrets:fnox`       | Delete the old file. Select the `fnox` tool.                                      |
| `setup:secrets` (Doppler)          | `setup:secrets:doppler`    | Delete the old file. Select the `doppler` tool. It needs a `doppler.yaml`.        |
| `setup:secrets` (stub)             | none                       | Delete it.                                                                        |
| `setup:deps:check`                 | `setup:external:check`     | Delete it. The `docker` tool ships the new one.                                   |
| `setup:git-config`                 | `code:git-config`          | Delete it.                                                                        |
| top-level `count`, `leaks`, `lint` | `code:*`                   | Delete them.                                                                      |

Update each shell alias that calls an old name. Remove a repository-owned
`setup` alias: `base` ships `setup = "mise run setup"`.

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
2. Move the repository's own `[env]` values to the seed
   `.config/mise/conf.d/_repo/mise.<env>.toml` of each environment that needs
   them. Examples are `GITHUB_OWNER`, and the emulator, Temporal and OTEL
   variables of 95octane backend (`mise.dev.toml`).
3. Delete what a tool now supplies: `CLOUDSDK_PYTHON` (the `gcloud` tool) and
   `task.timings = true` (the shared `[settings]`). Delete `lockfile = true`.
4. Keep `.config/mise.test.toml` and `.config/miserc.toml`. The sync does not
   write them.

### M6: Remove the tools that the shared files now install

From the repository-owned mise files, remove each tool that a selected tool
installs:

- `base`: dprint, `npm:@askviraj/linter`, gitleaks, grype, osv-scanner,
  shellcheck, gomplate, jq, yq, pre-commit. Remove shfmt
  too: dprint formats shell scripts now.
- `python`: python, uv. `node`: node, pnpm. `fnox`: fnox.
- `github`: gh, actionlint. `gitlab`: glab.
- `claude`: `pipx:graphifyy`, `pipx:mempalace`.
- `flutter`: flutter, android-sdk, java. `bun`: bun.
- `rust`: `core:rust`, zola. `cloudflare`: `npm:cf`. `pulumi`: pulumi.
  `gcloud`: gcloud. `doppler`: doppler.
- `kotlin`: `http:kotlin-lsp`.

Remove taplo too. dprint formats TOML now. Also remove the `[toml]` and
`evenBetterToml.*` keys from `.vscode/settings.json`. Remove `npm:wrangler`:
it is deprecated, and `npm:cf` of the `cloudflare` tool replaces it.

Each tool must come from a repo-config tool. If the repository needs a tool
that no repo-config tool installs, for example `npm:sort-package-json` or
`rtk`, tell the user. Then add it to a repo-config tool, or make a new tool
for it.

If the old tool has a pinned version that is different from the shared one,
tell the user. Examples: node `26` and pnpm `12` in 95octane, where `node`
ships `lts` and `latest`.

### M7: Split an old `conf.d/_base`

bootstrap and claude-plugins have an old `.config/mise/conf.d/_base/`. Its
name is the same as the shared `_base`, but the content is different. Before
the first sync, move its content:

| Old content in `_base`                                           | New place               |
| ---------------------------------------------------------------- | ----------------------- |
| node, pnpm                                                       | the `node` tool         |
| python, uv, `UV_*`                                               | the `python` tool       |
| pre-commit, `PRE_COMMIT_HOME`                                    | the `base` tool         |
| `MEMBERS`, `REPO_NAME`                                           | the values file         |
| `MEMPALACE_*` and other repository variables                     | `_repo/mise.<env>.toml` |
| `min_version`, `task.run_auto_install`, other repository content | step M5                 |

Then delete the old `_base/mise.toml`, `_base/mise.dev.toml` and
`_base/mise.ci.toml`. Also delete an old `conf.d/ai/`: the `claude` tool
replaces it. Move its repository-owned content first (step M5).
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

**WARNING:** Do this before the first sync. The sync replaces
`.config/pre-commit-config.yaml`, and it does not keep the hooks of the
repository.

1. Compare the hooks of the existing file with the rendered file. Use
   `mise run setup:configs --dry-run`.
2. Move each hook that no tool ships to `.config/repo-sync/pre-commit.yaml`.
   Prefix each `entry` that calls a tool with `mise x --`.
3. Remove each hook that a tool now ships, for example `graphify-refresh`,
   `gitleaks` and `conventional-commits`.

### M13: The forge files

`github` and `gitlab` replace these files, if they exist:

- `github`: `.github/ISSUE_TEMPLATE/bug_report.yml`, `feature_request.yml`,
  `config.yml`, and `.github/pull_request_template.md`.
- `gitlab`: `.gitlab/issue_templates/Bug.md`, `Feature.md`, and
  `.gitlab/merge_request_templates/Default.md`.

Before the first sync, compare them with the shared files. Tell the user
about each repository-specific change, for example an extra contact link.
Remove `actionlint`, `gh` and `glab` from the repository-owned mise files.

### M14: The mise environments

1. Add `environments` to `.config/repo-config.toml` (step 1). Delete an old
   `core_tools` key: the `envs` of each tool decide its environments now.
2. Move the repository variables of an old `_env/mise.toml` or
   `conf.d/repo/` to the `_repo/mise.<env>.toml` seeds (step M5). Move each
   tool of an old `conf.d/repo/` to a repo-config tool (step M6). Then delete
   `conf.d/repo/`.
3. Move the seeds of an old `conf.d/_env/` folder to `conf.d/_repo/`. The sync
   does not remove them, because a seed has no header.

The next sync removes the old `conf.d/*/mise.toml`, `_env/mise.toml`,
`_alias/mise.dev.toml`, `conf.d/_task/` and `conf.d/pre-commit/` files that
have the repo-config header.
