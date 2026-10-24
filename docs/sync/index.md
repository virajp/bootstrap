# Adopt bootstrap in a repository, or bring it up to date

These instructions are for an AI agent that works in a consumer repository.
Do the steps in sequence to make the repository get its shared files from
`bootstrap` (`git@github.com:virajp/bootstrap.git`). Read `README.md` in
`bootstrap` for the sync contract.

This file is a prompt. A user or an agent of a consumer repository can use it
directly, for example: "Read `docs/sync/index.md` of bootstrap and do its
steps in this repository." The `bootstrap: adopt` instruction of
[`vwf.md`](vwf.md) also sends an agent here.

Use this file for the first adoption, and also later, to bring a repository
up to date with `bootstrap`. For a repository that already has
`.config/bootstrap.toml`, start with
"A repository that already has `.config/bootstrap.toml`".

## Two ways to use bootstrap

A repository uses `bootstrap` in two ways, in this sequence:

1. **Adoption and update, by an AI agent.** The agent reads this folder and
   applies the rules of `bootstrap` to the repository. It does this one time
   for the adoption, and again when the repository is not up to date. The agent
   makes sure
   that each rule is relevant to the repository: it selects only the tools
   that the repository needs, it keeps the repository-owned content, and it
   asks the user about each difference.
2. **Sync, each time after that, by `repo-sync`.** The adoption installs the
   `setup:configs` task. The task runs `repo_sync`, which copies the newest
   shared files of the selected tools into the repository.

The adoption makes sure that the repository is ready for the sync. The sync
then keeps the repository equal to `bootstrap`, without an agent.

## Files in this folder

| File              | Use                                                                                                                                          |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `index.md`        | This file: the general steps of an adoption.                                                                                                 |
| `migrate.md`      | The items for a repository that has older copies, or the old name.                                                                           |
| `tools/<tool>.md` | One file for each folder in `tools/`: when the tool is relevant, and usage.                                                                  |
| `packs/<pack>.md` | One file for each pack of `tools/packs.toml`: the tools that it adds, and its questions.                                                     |
| `vwf.md`          | The entry point for the vwf and stackgen plugins: the `adopt`, `add tools`, `remove tools`, `add plugins` and `remove plugins` instructions. |

The tool files in `tools/` are `_security_base.md`, `android.md`, `bun.md`,
`claude-tool.md`, `cloudflare.md`, `docker.md`, `doppler.md`, `dprint.md`,
`flutter.md`, `fnox.md`, `gcloud.md`, `git.md`, `github.md`, `gitlab.md`,
`gitleaks.md`, `grype.md`, `java.md`, `kotlin.md`, `linter.md`, `mise.md`,
`node.md`, `osv-scanner.md`, `pre-commit.md`, `pulumi.md`, `python.md`,
`repo-sync.md`, `rust.md`, `swift.md` and `vscode.md`. The file of the
`claude` tool is `claude-tool.md`, because a file system that ignores case
reads `claude.md` as `CLAUDE.md`, and AI agents load that name as
instructions.

The pack files in `packs/` are `base.md` and `security.md`.

## Rules

- Ask each question of this file. Do not skip a question: only the user can
  skip one. Before you ask, find each value that you can, and show it as the
  default. A value of a `bootstrap: adopt` instruction is a default too.
- Ask the user before each choice that this file does not decide. Do not
  use a tool, a value, or a path that the user did not see.
- Apply only what is relevant to the repository. Before you select a tool,
  read its tool file. Examine the repository for the signs in "When the tool
  is relevant", then show the user the result.
- The shared files are grouped by tool. Each tool is a folder
  `tools/<tool>/` in `bootstrap`. A file at `tools/<tool>/<path>` goes to
  `<path>` in the consumer repository.
  A tool folder must not be `zz-repo`, and its name must not sort after
  `zz-repo` in byte order. If it does, the sync stops with code 2.
- The values file `.config/bootstrap.toml` lists the tools and the packs of
  the repository. `repo_sync` always adds the `base` pack, and it syncs all
  the files of these tools.
- Do not edit a synced file in the consumer repository. The next sync replaces
  it. Put repository-owned content in the places that the sync never writes:
  - a different default of a tool, for example a version, a mise setting or
    an env value: `[overrides]` of the values file (see step 3).
  - mise config and environment variables that no tool has: the seeds
    `.config/mise/conf.d/zz-repo/mise.<env>.toml`, one for each environment.
    mise loads the folders of `conf.d/` in name order, and a later folder
    wins. The name `zz-repo` sorts last, so the repository config wins over
    the config of each tool.
  - shell functions: `.config/mise/tasks/_scripts/local`.
  - ignore lines: `[overrides.gitignore]` of the values file (see step 3).
    For graphify, `[overrides.graphifyignore]`.
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

2. Ask the user whether to commit it on `main`. This is the first commit.
   Without a commit, no branch exists, and the sync stops (step 4).

Then do the other steps. A worktree in `.worktrees/` is now ignored, and the
branch `main` exists. The first sync replaces `.gitignore` with the shared one,
which has the same two lines.

## A product with more than one repository

A product can have more than one repository. Each repository has its own
values file and its own copy of the shared files. Do the steps in each
repository of the product, not only in the base repository. Run the commands
from the root of each repository: `MISE_PROJECT_ROOT` and all the paths in
this file are relative to it.

The tools and the values of each repository can be different. For each
repository, show the answers of the repositories before it as the defaults.
Each repository of a product has the same `product_name`.

`examples/` of `bootstrap` has one folder for each shape. Use the folder of
the shape of the product as the model of each values file:

| Shape                                                  | Example                      |
| ------------------------------------------------------ | ---------------------------- |
| One repository, with `main` and `develop`              | `examples/polyrepo/`         |
| One repository, with `main` only                       | `examples/polyrepo-main/`    |
| One repository, with `develop` only                    | `examples/polyrepo-develop/` |
| One repository, with member folders                    | `examples/monorepo/`         |
| A base repository with git submodules                  | `examples/repo-submodules/`  |
| A base repository with sibling repositories next to it | `examples/repo-siblings/`    |

### Git submodules

The base repository has the members as git submodules, for example
`backend/` and `frontend/`.

1. Get the list of submodules, nested submodules included.

   ```bash
   git submodule foreach --quiet --recursive 'echo $displaypath'
   ```

2. Show the list to the user. Ask which submodules to include. If a
   submodule is not initialized, ask whether to run
   `git submodule update --init <path>`, or to skip it.
3. Do the steps in each included submodule. Start with the deepest
   submodules. Do the base repository last.
4. In the base repository, write a `[member.<name>]` table for each
   submodule, with its `label`.

A tool with `root_only = true` is for the base repository only. The sync
refuses it in a submodule. A `root_only` tool of a pack, for example `vscode`
of `base`, is skipped in a submodule. When a submodule is in its root
checkout, mise also loads the tools of the root. Thus a nested submodule still
finds, for example, graphify.

### Siblings

The members are separate checkouts next to the base repository, for example
`../acme-api` and `../acme-app`. They are not submodules. vwf names this
shape `linkage: siblings`, and lists the paths in `members:` of
`.config/vwf.yaml`.

1. Get the paths of the members from `members:` of `.config/vwf.yaml`, or ask
   the user for them. Show the list to the user. Ask which members to include.
2. If a member checkout is missing, ask whether to clone it to its path, or
   to skip it.
3. Do the steps in each included member, then in the base repository.
4. Do not write a `[member.<name>]` table for a sibling. A member table names
   a folder below the root, and a sibling is not below the root.

## A repository that already has `.config/bootstrap.toml`

Do not stop. Find whether the repository is up to date, then do only the
steps that are not. Show the result of each check to the user before you
change a file.

1. Make the clone of step 1, and get the newest commit of `bootstrap`:
   `git -C "$tmp" rev-parse HEAD`.
2. Read `bootstrap_commit` of the values file. `repo_sync` writes it after
   each sync without a failure.
3. Use `bootstrap_commit` to select the checks of step 4:

   | `bootstrap_commit` | The checks of the kind `bootstrap`                                                                                              | The checks of the kind `repository` |
   | ------------------ | ------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
   | The same as `main` | Skip them. The shared files are current.                                                                                        | Do them.                            |
   | Older than `main`  | Do them only for the tools and the migrate items that changed: `git -C "$tmp" log --oneline <commit>..HEAD -- tools docs/sync`. | Do them.                            |
   | Missing or empty   | Do them all.                                                                                                                    | Do them.                            |

   The checks of the kind `repository` are always necessary: the repository
   can change also when `bootstrap` does not, for example with a new
   `pyproject.toml` or a new submodule.
4. Do the selected checks of this table. Each check that fails sends you to
   a step:

   | Check                                                                                                                      | Kind       | If it fails, do         |
   | -------------------------------------------------------------------------------------------------------------------------- | ---------- | ----------------------- |
   | The values file has no old format and no old key, for example `tools = [...]` or `.config/repo-sync/gitignore`.            | bootstrap  | Step 3 and `migrate.md` |
   | No item of `migrate.md` applies, and no "Migration" item of a selected tool applies.                                       | bootstrap  | Step 5                  |
   | `mise run setup:configs --dry-run` shows no change.                                                                        | bootstrap  | Step 6                  |
   | The selected tools agree with the signs of their tool files: no tool is missing, and no tool is selected without a reason. | repository | Step 1                  |
   | Each value that you can find agrees with the file, for example `repo_url` and `origin`, or the members and `.gitmodules`.  | repository | Step 3                  |
   | Each file of "Files the repository supplies" of a selected tool exists.                                                    | repository | Step 7                  |

5. If all the checks pass, tell the user that the repository is up to date,
   and stop.

## Step 1: Select the tools and the packs

Clone `bootstrap` into a temporary directory. Keep it until all repositories
are complete. `--filter=blob:none` gets all the commits, for the `git log` of
an update, but only the files of the newest commit.

```bash
tmp=$(mktemp -d)
git clone --quiet --filter=blob:none git@github.com:virajp/bootstrap.git "$tmp"
```

Read `"$tmp"/tools/*/tools.toml`, `"$tmp"/tools/packs.toml`,
`"$tmp"/docs/sync/tools/<tool>.md` and `"$tmp"/docs/sync/packs/<pack>.md`.
Examine the repository for the signs in "When the tool is relevant" of each
tool file. Show the user each tool and each pack with its description and the
result. Then ask which tools and packs the repository needs. Ask the questions
of each pack file.

The `base` pack is always added: `repo_sync` adds it, and the values file does
not list it. Its tools are `mise`, `repo-sync`, `git`, `pre-commit`, `dprint`,
`linter`, and the tools of the `security` pack. Do not select a tool of the
`base` pack. See [`packs/base.md`](packs/base.md).

For each selected tool, and each tool of the `base` pack, do the items in
"Adapt to the repository" of its tool file.

A pack is a `[pack.<name>]` table of `tools/packs.toml`. Its `depends` lists
tool folders, or a pack with the prefix `p:`, for example `p:security`. Its
optional `required` lists the tools that the pack always adds. A pack name is
never a tool folder.

Each `tools.toml` has these keys:

| Key            | Meaning                                                                                      |
| -------------- | -------------------------------------------------------------------------------------------- |
| `default_envs` | The environments of the tool when the values file gives it no `env`. It must include `dev`.  |
| `depends`      | The sync adds these tools automatically.                                                     |
| `root_only`    | If `true`, the sync refuses the tool in a submodule, and skips it there when a pack adds it. |
| `extensions`   | The VS Code extensions of the tool, for `.vscode/extensions.json`.                           |
| `seed`         | Files that the sync writes only if they are missing.                                         |

A `tools.toml` can also have a `[seed_requires]` table. It is optional. Each
key is a seed, and its value is a file of the repository. The sync writes the
seed only if the repository has that file. For example, the `node` tool writes
`pnpm-workspace.yaml` only if the repository has a `package.json`.

The `[mise.*]` tables of a `tools.toml` hold the mise config of the tool. The
sync generates `.config/mise/conf.d/<tool>/mise.<env>.toml` from them, one
file for each environment. A mise `depends` adds the tool that defines the
dependency. For example, the `pre-commit` tool adds `python`, and the
`linter` tool adds `node`.

A mise tool can have a `requires` key with a file name. The sync writes that
mise tool only when the repository has a file with that name, at the root or
down to 4 folders below it, that git does not ignore. For example, the
`python` tool installs `ruff` and `pyrefly` only in a repository with a
`pyproject.toml`. If the repository adds its first `pyproject.toml` later,
run `mise run setup:configs` again.

The `files` key of a mise tool with `requires` lists files of its tool, for
example the tasks of that mise tool. The sync writes them under the same
condition. Without the file, it removes a copy with the bootstrap header. For
example, the `python` tool writes `code:lint:ruff`, `code:format:ruff` and
`code:lint:pyrefly` only in a repository with a `pyproject.toml`.

The sync finds the environments of each tool in this sequence:

1. A tool of a `[tools.<folder>]` entry gets its `env`, else its
   `default_envs`. A tool of a `[packs.<name>]` entry gets the `env` of the
   pack entry, else its own `default_envs`. Each tool of the `base` pack gets
   its own `default_envs`. A tool that is selected more than once gets each of
   these environments.
2. A dependency gets each environment of each tool that needs it, through
   `depends` or a mise `depends`. For example, `flutter` adds the `android` and
   `java` tools in dev and ci.
3. `environments` of the values file limits every result.
4. Each tool must be in `dev`. An `env` without `dev` stops the sync with
   code 2. A `default_envs` without `dev` in a `tools.toml` of `bootstrap`
   also stops the sync with code 2, also for a tool that the repository does
   not select.

The `envs` of a mise tool (default `["dev"]`) lists the environments that the
mise tool allows. In a tool file, the "Environments" column of "mise tools"
shows them. If a tool gets an environment that one of its mise tools does not
allow, the sync stops with code 2.

A tool that only a dependency adds goes only to the environments of the tools
that need it. Thus list the core language of the repository in `[tools.*]`. A
TypeScript repository lists `[tools.node]`, which then goes to each
environment of its `default_envs`.

Ask which mise environments the repository needs. A CI config, for example
`.github/workflows/` or `.gitlab-ci.yml`, is a sign of `ci`. Offer only
these:

| Environment | Use                                                          |
| ----------- | ------------------------------------------------------------ |
| `dev`       | Local development. It usually has everything. Necessary.     |
| `ci`        | The CI environment, to build and examine the code. Optional. |

The sync stops with code 2 if `dev` is missing, or for a different name.

Find the forge from the host of `origin`, `github.com` or `gitlab.com`. Show
it, and ask whether the repository is on GitHub or on GitLab. Select `github`
or `gitlab` for its forge. The `pr` merge model needs the CLI of the forge: `gh`
or `glab`.

## Step 2: Copy the sync files by hand

**CAUTION:** Do this before the first sync. An old synced
`setup/configs/sync` loads `tools/base/.../repo-sync` from the clone, and the
`tools/base/` folder no longer exists. An old copy of `repo-sync` reads paths,
not tools. Thus neither can sync the new layout.

For an update, skip this step when `.config/mise/tasks/setup/configs/_default`
and `.config/mise/tasks/setup/configs/sync` exist and have the bootstrap
header. Then `setup:configs` loads the sync code from its own clone.

Copy these four files from the clone into the repository: `repo-sync` and the
`setup:configs` tasks of the `repo-sync` tool, and `helpers` of the `mise`
tool. An old `.config/mise/tasks/setup/configs` is a file, and
`setup/configs/` is now a folder: delete the old file first.

```bash
sync="$tmp/tools/repo-sync/.config/mise/tasks"
lib="$tmp/tools/mise/.config/mise/tasks"
rm -f .config/mise/tasks/setup/configs
mkdir -p .config/mise/tasks/_scripts .config/mise/tasks/setup/configs
cp "$sync/_scripts/repo-sync" "$lib/_scripts/helpers" .config/mise/tasks/_scripts/
cp "$sync/setup/configs/_default" "$sync/setup/configs/sync" .config/mise/tasks/setup/configs/
```

Do not run the sync, `mise run setup:configs` without `--dry-run`, before
step 6. You can run a dry run.

## Step 3: Write the values file

1. Find each value of the table below that you can. These are the usual
   places:

   | Value                    | Where to find it                                                                                                    |
   | ------------------------ | ------------------------------------------------------------------------------------------------------------------- |
   | `repo_name`              | The last part of the path of `origin`, without `.git`. Else the folder name, but not the name of a worktree folder. |
   | `repo_url`               | `origin`, as an `https://` URL.                                                                                     |
   | `product_name`           | `product.name` of `.config/vwf.yaml`, or the value of the other repositories of the product.                        |
   | `environments`           | Step 1.                                                                                                             |
   | `[tools.*]`, `[packs.*]` | Step 1.                                                                                                             |
   | `[branch.<branch>]`      | Step 4. Without a sign, `merge_model = "direct"`.                                                                   |
   | `[member.<name>]`        | `.gitmodules`, and the member folders of a monorepo.                                                                |
   | `commit_scopes`          | The scopes of the existing commits, `git log --format=%s`. Else the members.                                        |
   | `worktree_files`         | Gitignored files that a new worktree needs, for example `.config/mise.local.toml`.                                  |
   | `extensions`             | The `recommendations` of `.vscode/extensions.json` that no selected tool has.                                       |
   | `[overrides.gitignore]`  | The lines of the existing `.gitignore` that the shared one does not have.                                           |
   | `[claude.plugins.*]`     | The plugins of `.claude/settings.json` (project scope).                                                             |

2. Use the example of `examples/` with the shape of the repository as the
   model (see "A product with more than one repository").
3. Show the user the full draft of `.config/bootstrap.toml`. Ask whether the
   user wants to change a value.
4. Write the file with the values of the user.

The keys:

| Key                                    | Used by                                                            | Values                                                                                                                                                                                                                                                                  |
| -------------------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bootstrap_commit`                     | `repo_sync`, this file                                             | Do not write it, and do not copy it from an example: each example has an empty value. `repo_sync` writes the `bootstrap` commit of each sync without a failure. See "A repository that already has `.config/bootstrap.toml`".                                           |
| `[tools.<folder>]`                     | `repo_sync`                                                        | One table for each tool from step 1. Do not write a tool of the `base` pack.                                                                                                                                                                                            |
| `[tools.<folder>].env`                 | `repo_sync`                                                        | Optional. The environments of the tool, for example `["dev", "ci"]`. Without it, the tool gets its `default_envs`.                                                                                                                                                      |
| `[packs.<name>]`                       | `repo_sync`                                                        | One table for each pack from step 1, without the prefix `p:`. Do not write `base`.                                                                                                                                                                                      |
| `[packs.<name>].env`                   | `repo_sync`                                                        | Optional. The environments of each tool of the pack. Without it, each tool gets its own `default_envs`.                                                                                                                                                                 |
| `[overrides.mise.<section>]`           | `repo_sync`                                                        | Optional. New values for the `[mise.<section>]` defaults of the selected tools, in all environments, for example `[overrides.mise.tools.java]` with `version = "21"`. See "Overrides".                                                                                  |
| `[overrides.mise_env.<env>.<section>]` | `repo_sync`                                                        | Optional. New values for one environment only. They win over `[overrides.mise]`. See "Overrides".                                                                                                                                                                       |
| `[overrides.gitignore].remove`         | `repo_sync`                                                        | Optional. The lines of the shared `.gitignore` to comment out, for example `["*.log"]`. See "Overrides".                                                                                                                                                                |
| `[overrides.gitignore].add`            | `repo_sync`                                                        | Optional. The repository's own lines, added at the end of `.gitignore`, for example `["/build/"]`. See "Overrides".                                                                                                                                                     |
| `[overrides.graphifyignore].remove`    | `repo_sync`                                                        | Optional. The lines of the shared `.graphifyignore` to comment out, for example `["docs/prompts/"]`. Only with the `claude` tool. See "Overrides".                                                                                                                      |
| `[overrides.graphifyignore].add`       | `repo_sync`                                                        | Optional. The repository's own lines, added at the end of `.graphifyignore`, for example `["fixtures/"]`. Only with the `claude` tool. See "Overrides".                                                                                                                 |
| `environments`                         | `repo_sync`                                                        | The environments from step 1, for example `["dev", "ci"]`. The default is `["dev"]`.                                                                                                                                                                                    |
| `repo_name`                            | `.config/mise.toml`, `gitleaks.toml`, `fnox.toml`                  | The repository name, for example `my-repo`. It is a literal, not the folder name of a worktree.                                                                                                                                                                         |
| `product_name`                         | `fnox.toml`                                                        | The product name, for example `acme`. Set it in each repository of a product with members, the root included. Then the fnox keychain uses it as `service`, and `<repo_name>/` as `prefix`. Without it, `service` is `repo_name` and there is no prefix.                 |
| `[claude.plugins.user]`                | `setup:ai`                                                         | Optional. The Claude Code plugins to install at user scope: `"<name>@<marketplace>" = "<source of the marketplace>"`, for example `"taste-skill@taste-skill" = "Leonxlnx/taste-skill"`. See `tools/claude-tool.md`.                                                     |
| `[claude.plugins.project]`             | `setup:ai`                                                         | Optional. The Claude Code plugins to install at project scope, in the same format.                                                                                                                                                                                      |
| `[branch.<branch>].merge_model`        | `code:merge:*`                                                     | The merge config of this repository, for the destination `<branch>`: `direct` (local merge, the default) or `pr` (pull request). Each repository has its own, also a submodule.                                                                                         |
| `[branch.<branch>].merge_from`         | `code:merge:*`                                                     | Optional. The only branches that can merge into `<branch>`. Do not set it with `merge_never`.                                                                                                                                                                           |
| `[branch.<branch>].merge_never`        | `code:merge:*`                                                     | Optional. The branches that never merge into `<branch>`. Without both lists: into `main` with `develop`, only from `develop`; into any other branch, never from `main`.                                                                                                 |
| `commit_scopes`                        | `git-conventional-commits`                                         | The commit scopes, for example `["backend", "web"]`. The default is the member list.                                                                                                                                                                                    |
| `repo_url`                             | `github` tool                                                      | The web URL of the repository, for the links of the GitHub issue form: the README and a new security advisory.                                                                                                                                                          |
| `worktree_files`                       | `setup:worktree`                                                   | The gitignored files to copy from the main worktree into a new worktree, for example `[".config/mise.local.toml"]`                                                                                                                                                      |
| `extensions`                           | `vscode` (in the root)                                             | VS Code extensions that no tool has, for example `["astro-build.astro-vscode"]`. A submodule lists its own too.                                                                                                                                                         |
| `[member.<name>]`                      | `setup`, `setup:worktree`                                          | One table for each member folder                                                                                                                                                                                                                                        |
| `[member.<name>].folder`               | `repo_sync`, `setup`, `setup:worktree`, `code:format`, `code:lint` | Optional. The folder of the member below the root, for example `projects/web`. The default is `<name>`. A table whose folder is a submodule path names that submodule. A submodule without a table has its path as its name. The folder `.` stops the sync with code 2. |
| `[member.<name>].type`                 | `repo_sync`                                                        | Optional. `submodule` or `folder`. Without it, the sync finds the type. A type that is not correct stops the sync with code 2.                                                                                                                                          |
| `[member.<name>].label`                | `setup`, `setup:worktree`                                          | The display name. The default is `<name>`.                                                                                                                                                                                                                              |
| `[member.<name>].worktree_install`     | `setup:worktree`                                                   | Only for a submodule member. The install command for a new worktree. If it is not set, the member's own `setup:worktree` runs. A folder member never has it.                                                                                                            |

Example:

```toml
commit_scopes = [ "backend", "web" ]
environments = [ "dev", "ci" ]
product_name = "my-product"
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"
worktree_files = [ ".config/mise.local.toml" ]

[tools.claude]

[tools.fnox]

[tools.github]

[tools.node]
env = [ "dev", "ci" ]

[branch.main]
merge_from = [ "develop" ]
merge_model = "pr"
```

The only key of a `[tools.*]` or a `[packs.*]` entry is `env`. Each name in
`env` must be in `environments`, and `env` must include `dev`. A list
(`tools = [...]` or `packs = [...]`) is the old format, and the sync stops
with code 2.

The member list is the git submodules, then each `[member.<name>]` table. A
`[member.<name>]` table that is not a submodule must have a folder with that
name in the repository. If not, the render fails.

### Overrides

`[overrides]` replaces a default of a tool. The path after `overrides.` is the
path in `tools/<tool>/tools.toml`, without the tool name. The sync applies the
override to each selected tool that has a value at that path.

```toml
# Java 21, not lts, in all environments.
[overrides.mise.tools.java]
version = "21"

# Node 22 in CI only. This wins over [overrides.mise.tools.node].
[overrides.mise_env.ci.tools.node]
version = "22"

# A mise setting and an env value of the python tool.
[overrides.mise.settings]
python.compile = true

[overrides.mise.env]
UV_NO_CACHE = false
```

- A table merges key by key. Any other value, also a list, replaces the
  default.
- `[overrides.mise_env.<env>]` matches a value in `[mise.*]` or in
  `[mise_env.<env>.*]` of a tool. The environment must be in `environments`.
- Each line that an override changes gets the comment
  `# override: .config/bootstrap.toml` in
  `.config/mise/conf.d/<tool>/mise.<env>.toml`.
- These stop the sync with code 2: a key of `[overrides]` other than `mise`,
  `mise_env`, `gitignore` and `graphifyignore`, a path that no selected tool
  has, and the
  `envs`, the `depends`, the `requires` or the `files` of a mise tool. An
  override cannot
  add a key. Put a
  new key in `conf.d/zz-repo/mise.<env>.toml`.

`[overrides.gitignore]` changes the shared `.gitignore`. It has two lists of
lines. The sync applies `remove` first, then `add`.

```toml
[overrides.gitignore]
add = [ "/build/", "/coverage/", "!.vscode/settings.json" ]
remove = [ ".env", "*.log" ]
```

- `remove` comments out each matching line of the shared `.gitignore` and
  adds `# override: removed by .config/bootstrap.toml`. A line matches when it
  is the same after the sync removes the spaces at its start and its end.
- `add` puts its lines at the end, under the heading
  `# ==== Overrides from .config/bootstrap.toml ====`. A later line wins, so
  a line with `!` cancels a shared pattern.
- These stop the sync with code 2: a key other than `add` and `remove`, a
  value that is not a list of one-line strings, an empty line, a line in both
  lists, and a line of `remove` that the shared `.gitignore` does not have.
- The old file `.config/repo-sync/gitignore` also stops the sync with code 2.
  See M18 of `migrate.md`.

`[overrides.graphifyignore]` changes the `.graphifyignore` of the `claude`
tool with the same rules. graphify also reads `.gitignore`, so add only the
committed files that are not code, for example test fixtures.
`[overrides.graphifyignore]` without the `claude` tool stops the sync with
code 2.

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

The items of [`migrate.md`](migrate.md) that apply come before the first
sync. So do the "Migration" sections of each selected tool file, and of each
tool of the `base` pack. If the repository has no mise config and no task
files, skip this step.

1. Find the items that apply. For an update, find only the items of the
   checks of "A repository that already has `.config/bootstrap.toml`".
2. Show the user the list, with the files that each item changes.
3. Ask one question: do all the items? The user can name the items to leave
   out.
4. Do the items, without a question for each step. The warning of
   `migrate.md` still applies: move the repository-owned content before you
   delete a file.

## Step 6: Run the sync

1. If mise says that the config files are not trusted, ask the user to run
   `mise trust`.
2. Run `mise run setup:configs --dry-run`.
3. Examine the diff for each file that already exists in the repository.
4. If the dry run replaces repository-owned content, stop. Move that content
   to a repository-owned place (see "Rules"), then do the dry run again.
5. Ask the user whether to run `mise run setup:configs`. In the same
   question, ask whether to commit the changes at the end (step 9), or to
   leave them for review.
6. If the user agrees, run `mise run setup:configs`. A sync without a failure
   sets `bootstrap_commit` in the values file.
7. Run `mise run init`. This makes all the task scripts executable.

If `repo_sync` fails, read the exit code:

| Code | Cause                                                                                                                                                                                                                                             | Action                                           |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| 2    | The old list format, a bad `[tools.*]` or `[packs.*]` entry, a bad `env` (also an `env` without `dev`) or `environments`, a bad `[overrides]`, a root-only tool in a submodule, or a template that cannot render                                  | Correct the values file                          |
| 2    | A tool in an environment that one of its mise tools does not allow                                                                                                                                                                                | Correct the `env`, or tell the `bootstrap` owner |
| 2    | A pack loop, or a pack name that is also a tool folder                                                                                                                                                                                            | Tell the `bootstrap` owner                       |
| 2    | A `default_envs` without `dev`, or a tool folder that is `zz-repo` or that sorts after `zz-repo`                                                                                                                                                  | Tell the `bootstrap` owner                       |
| 2    | Two tools define the same mise tool, a mise `depends` that no tool defines, a tool that ships and generates the same `conf.d` file, an unsupported `[mise.*]` or `[mise_env.*]` section, or a `default_envs` or `envs` name that is not dev or ci | Tell the `bootstrap` owner                       |
| 2    | The values file is missing, or gomplate is not available                                                                                                                                                                                          | Correct the values file, or run `mise install`   |
| 2    | Neither `main` nor `develop` exists                                                                                                                                                                                                               | Do step 4                                        |
| 3    | A tool or a pack that is not in `bootstrap`                                                                                                                                                                                                       | Correct the `[tools.*]` and `[packs.*]` entries  |
| 5    | The clone failed, `REPO_SYNC_CLONE` is not usable, pre-commit is not available, or `pre-commit autoupdate` failed                                                                                                                                 | Examine the network, the SSH key and the access  |
| 1    | A local file system error                                                                                                                                                                                                                         | Read the `repo-sync:` lines on stderr            |

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
   `kept seed`, `skip seed` or `skip env`.

mise always loads `.config/mise.toml`. It loads a `conf.d/*/mise.<env>.toml`
file only when `MISE_ENV` includes that environment. Without `MISE_ENV`, no
file in `conf.d` loads.

## Step 9: Commit

If the user chose to leave the changes for review in step 6, show
`git status --short` in each repository, and stop. Else commit these files:

- The synced files and the seed files.
- `.config/bootstrap.toml` and `.config/repo-sync/pre-commit.yaml`.
- The repository-owned files from step 5 and step 7.

Commit in each submodule before you commit in the repository that contains
it. Then the commit in the containing repository includes the new submodule
commit. Ask the user before you push. Push each submodule before the
repository that contains it. If not, the containing repository points to a
commit that the remote does not have.
