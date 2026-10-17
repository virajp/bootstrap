# Adopt bootstrap in a repository

These instructions are for an AI agent that works in a consumer repository.
Do the steps in sequence to make the repository get its shared files from
`bootstrap` (`git@github.com:virajp/bootstrap.git`). Read `README.md` in
`bootstrap` for the sync contract.

## Two ways to use bootstrap

A repository uses `bootstrap` in two ways, in this sequence:

1. **Adoption, one time, by an AI agent.** The agent reads this folder and
   applies the rules of `bootstrap` to the repository. The agent makes sure
   that each rule is relevant to the repository: it selects only the tools
   that the repository needs, it keeps the repository-owned content, and it
   asks the user about each difference.
2. **Sync, each time after that, by `repo-sync`.** The adoption installs the
   `setup:configs` task. The task runs `repo_sync`, which copies the newest
   shared files of the selected tools into the repository.

The adoption makes sure that the repository is ready for the sync. The sync
then keeps the repository equal to `bootstrap`, without an agent.

## Files in this folder

| File         | Use                                                                         |
| ------------ | --------------------------------------------------------------------------- |
| `index.md`   | This file: the general steps of an adoption.                                |
| `migrate.md` | The items for a repository that has older copies, or the old name.          |
| `<tool>.md`  | One file for each folder in `tools/`: when the tool is relevant, and usage. |

The tool files are `base.md`, `bun.md`, `claude-tool.md`, `cloudflare.md`,
`docker.md`, `doppler.md`, `flutter.md`, `fnox.md`, `gcloud.md`, `github.md`,
`gitlab.md`, `kotlin.md`, `node.md`, `pulumi.md`, `python.md`, `rust.md` and
`vscode.md`. The file of the `claude` tool is `claude-tool.md`, because a
file system that ignores case reads `claude.md` as `CLAUDE.md`, and AI agents
load that name as instructions.

## Rules

- Ask the user before each choice that this file does not decide. Do not
  assume a tool, a value, or a path.
- Apply only what is relevant to the repository. Before you select a tool,
  read its tool file. Examine the repository for the signs in "When the tool
  is relevant", then show the user the result.
- The shared files are grouped by tool. Each tool is a folder
  `tools/<tool>/` in `bootstrap`. A file at `tools/<tool>/<path>` goes to
  `<path>` in the consumer repository.
- The values file `.config/bootstrap.toml` lists the tools of the
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
  - formatters and linters: a task `code:format:<name>` or `code:lint:<name>`
    (`.config/mise/tasks/code/format/<name>`). `code:format` and `code:lint`
    run each of them with the same `--fix` and files. A pre-commit hook only
    calls a task, so do not call a formatter or a linter from a hook.
- A seed file is different. The sync writes it only if it is missing. After
  that, the repository owns it.
- To change a shared file, change it in `bootstrap`, push it to `main`, then
  sync again.
- Prefix each tool call with `mise x --`, also in the task files and the hook
  entries of `.config/repo-sync/pre-commit.yaml`.

## Requirements

- The machine has SSH access to `git@github.com:virajp/bootstrap.git`. The
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

If a repository already has a `tools` list in `.config/bootstrap.toml`,
tell the user. Then ask whether to do the steps again in that repository, or
to go to the next repository.

## Step 1: Select the tools

Make a shallow clone of `bootstrap` into a temporary directory. Keep it
until all repositories are complete.

```bash
tmp=$(mktemp -d)
git clone --quiet --depth 1 git@github.com:virajp/bootstrap.git "$tmp"
```

Read `"$tmp"/tools/*/tool.toml` and `"$tmp"/docs/sync/<tool>.md`. Examine
the repository for the signs in "When the tool is relevant" of each tool
file. Show the user each tool with its description and the result. Then ask
which tools the repository needs. `base` is necessary in each repository.

For each selected tool, do the items in "Adapt to the repository" of its tool
file.

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

Write `.config/bootstrap.toml`. Ask the user for each value.

| Key                                | Used by                    | Values                                                                                                                                                                                                                                                      |
| ---------------------------------- | -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `tools`                            | `repo_sync`                | The tools from step 1. `base` is necessary.                                                                                                                                                                                                                 |
| `environments`                     | `repo_sync`                | The environments from step 1, for example `["dev", "ci"]`. The default is `["dev"]`.                                                                                                                                                                        |
| `repo_name`                        | `.config/mise.toml`        | The repository name, for example `my-repo`. It is a literal, not the folder name of a worktree.                                                                                                                                                             |
| `product_name`                     | `fnox.toml`                | The product name, for example `95octane`. Set it in each repository of a product with members, the root included. Then the fnox keychain uses it as `service`, and `<repo_name>/` as `prefix`. Without it, `service` is `repo_name` and there is no prefix. |
| `merge_model_develop`              | `.config/mise.toml`        | `direct` (local merge) or `pr` (pull request)                                                                                                                                                                                                               |
| `merge_model_main`                 | `.config/mise.toml`        | `direct` or `pr`                                                                                                                                                                                                                                            |
| `commit_scopes`                    | `git-conventional-commits` | The commit scopes, for example `["backend", "web"]`. The default is the member list.                                                                                                                                                                        |
| `repo_url`                         | `git-conventional-commits` | The web URL of the repository, for the changelog links. GitHub and GitLab are supported.                                                                                                                                                                    |
| `release_tag_pattern`              | `git-conventional-commits` | The pattern of the release tags, for example `'^web-v\d+$'`. Use a TOML literal string (single quotes), so that each backslash stays. The default is `"*"`, all tags.                                                                                       |
| `worktree_files`                   | `setup:worktree`           | The gitignored files to copy from the main worktree into a new worktree, for example `[".config/mise.local.toml"]`                                                                                                                                          |
| `extensions`                       | `vscode` (in the root)     | VS Code extensions that no tool has, for example `["astro-build.astro-vscode"]`. A submodule lists its own too.                                                                                                                                             |
| `[member.<name>]`                  | `setup`, `setup:worktree`  | One table for each member folder                                                                                                                                                                                                                            |
| `[member.<name>].label`            | `setup`                    | The display name. The default is `<name>`.                                                                                                                                                                                                                  |
| `[member.<name>].worktree_install` | `setup:worktree`           | The install command for a new worktree. If it is not set, the member's own `setup:worktree` runs.                                                                                                                                                           |

Example:

```toml
commit_scopes = [ "backend", "web" ]
environments = [ "dev", "ci", "prod" ]
merge_model_develop = "direct"
merge_model_main = "pr"
product_name = "my-product"
release_tag_pattern = '^my-repo-v\d+$'
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

Do the items of [`migrate.md`](migrate.md) that apply, before the first sync.
Also do the "Migration" section of each selected tool file. If the repository
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
| 3    | A tool that is not in `bootstrap`                                                                | Correct the `tools` list                        |
| 5    | The clone failed, or `pre-commit autoupdate` failed                                              | Examine the network, the SSH key and the access |
| 1    | A local file system error                                                                        | Read the `repo-sync:` lines on stderr           |

A failure does not change an existing file.

## Step 7: Supply the files that `bootstrap` does not have

The shared files refer to files that each repository must supply. Make sure
that the repository has each file for the tools that it uses. If a file is
missing, ask the user for its content. Do not write a placeholder that hides
the problem.

Each tool file lists these files in "Files the repository supplies". Do this
for each selected tool.

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
- `.config/bootstrap.toml`, `.config/repo-sync/gitignore` and
  `.config/repo-sync/pre-commit.yaml`.
- The repository-owned files from step 5 and step 7.

Commit in each submodule before you commit in the repository that contains
it. Then the commit in the containing repository includes the new submodule
commit. Ask the user before you push. Push each submodule before the
repository that contains it. If not, the containing repository points to a
commit that the remote does not have.
