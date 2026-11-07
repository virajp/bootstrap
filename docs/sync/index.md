# Adopt bootstrap in a repository, or bring it up to date

These instructions are for an AI agent that works in a consumer repository.
Do the steps in sequence to make the repository get its shared files from
`bootstrap` (`git@github.com:virajp/bootstrap.git`). The contract of the sync is
in `docs/design/` of `bootstrap` (see "Where the rules are").

This file is a prompt. A user or an agent of a consumer repository can use it
directly, for example: "Read `docs/sync/index.md` of bootstrap and do its
steps in this repository." The `bootstrap: adopt` instruction of
[`vwf.md`](vwf.md) also sends an agent here.

Use this file for the first adoption, and also later, to bring a repository
up to date with `bootstrap`. For a repository that already has
`.config/bootstrap.toml`, start with
"A repository that already has `.config/bootstrap.toml`".

A Mac uses [`docs/setup/index.md`](../setup/index.md), not this prompt.

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

| File                | Use                                                                                                                                          |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `index.md`          | This file: the general steps of an adoption.                                                                                                 |
| `migrate.md`        | The items for a repository that has older copies, or the old name.                                                                           |
| `kotlin-dsl.md`     | A static prompt that migrates a Gradle build from the Groovy DSL to the Kotlin DSL.                                                          |
| `./tools/<tool>.md` | One file for each folder in `repo/tools/`: when the tool is relevant, and usage.                                                             |
| `packs/<pack>.md`   | One file for each pack of `repo/packs.toml`: the tools that it adds, and its questions.                                                      |
| `vwf.md`            | The entry point for the vwf and stackgen plugins: the `adopt`, `add tools`, `remove tools`, `add plugins` and `remove plugins` instructions. |

The tool files in `tools/` are `_security_base.md`, `android.md`, `apko.md`,
`aube.md`, `bun.md`, `claude-tool.md`, `cloudflare.md`, `cosign.md`, `crane.md`,
`docker.md`, `doppler.md`, `dprint.md`, `firebase.md`, `flutter.md`, `fnox.md`,
`gcloud.md`, `git.md`, `github.md`, `gitlab.md`, `gitleaks.md`, `grype.md`,
`java.md`, `kotlin.md`, `linter.md`, `melange.md`, `mise.md`, `node.md`,
`osv-scanner.md`, `pnpm.md`, `pre-commit.md`, `pulumi.md`, `python.md`,
`repo-sync.md`, `rust.md`, `swift.md` and `vscode.md`. The file of the
`claude` tool is `claude-tool.md`, because a file system that ignores case
reads `claude.md` as `CLAUDE.md`, and AI agents load that name as
instructions.

The pack files in `packs/` are `base.md`, `container-build.md` and
`security.md`.

## Where the rules are

This file holds the instructions. Each rule of the sync has one place in
`docs/design/` of the clone of step 1. Read the doc when a step needs the rule:

- Read [`docs/design/bootstrap-config.md`](../design/bootstrap-config.md) before
  you write a key of `.config/bootstrap.toml`.
- Read [`docs/design/packs.md`](../design/packs.md) before you select a pack.
- Read [`docs/design/tools.md`](../design/tools.md) before you read a
  `tools.toml`: its keys, the `[files]` tables, the seeds and the header.
- Read [`docs/design/mise.md`](../design/mise.md) before you select an
  environment or write `[overrides.mise*]`.
- Read [`docs/design/repo-sync.md`](../design/repo-sync.md) before you select
  the
  branch model or members, or write `[overrides.gitignore]`.
- Read [`docs/design/sync-engine.md`](../design/sync-engine.md) when you examine
  the clone, a parse failure, a stale file or an exit code.
- Read [`docs/design/fnox.md`](../design/fnox.md) before you select `fnox`.

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
  `repo/tools/<tool>/` in `bootstrap`. A file at `repo/tools/<tool>/<path>` goes
  to `<path>` in the consumer repository.
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

- The machine can reach `https://github.com/virajp/bootstrap.git`. The sync
  uses this URL and needs no SSH key. It does not ask for a password.
- The temporary clone of this prompt (step 1) uses
  `git@github.com:virajp/bootstrap.git`, which needs SSH access.
- mise is installed. The sync starts gomplate, jq and yq through mise, and mise
  installs each one that is missing. This install needs the network.

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

`repo/examples/` of `bootstrap` has one folder for each shape. Use the folder of
the shape of the product as the model of each values file:

| Shape                                                  | Example                           |
| ------------------------------------------------------ | --------------------------------- |
| One repository, with `main` and `develop`              | `repo/examples/polyrepo/`         |
| One repository, with `main` only                       | `repo/examples/polyrepo-main/`    |
| One repository, with `develop` only                    | `repo/examples/polyrepo-develop/` |
| One repository, with member folders                    | `repo/examples/monorepo/`         |
| A base repository with git submodules                  | `repo/examples/repo-submodules/`  |
| A base repository with sibling repositories next to it | `repo/examples/repo-siblings/`    |

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
checkout, mise also loads the tools of the root. Thus a nested submodule also
finds the tools of the root.

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

   | `bootstrap_commit` | The checks of the kind `bootstrap`                                                                                                   | The checks of the kind `repository` |
   | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------- |
   | The same as `main` | Skip them. The shared files are current.                                                                                             | Do them.                            |
   | Older than `main`  | Do them only for the tools and the migrate items that changed: `git -C "$tmp" log --oneline <commit>..HEAD -- tools repo docs/sync`. | Do them.                            |
   | Missing or empty   | Do them all.                                                                                                                         | Do them.                            |

   The checks of the kind `repository` are always necessary: the repository
   can change also when `bootstrap` does not, for example with a new
   `pyproject.toml` or a new submodule.
4. Do the selected checks of this table. Each check that fails sends you to
   a step:

   | Check                                                                                                                      | Kind       | If it fails, do         |
   | -------------------------------------------------------------------------------------------------------------------------- | ---------- | ----------------------- |
   | The values file has no old format and no old key, for example `tools = [...]` or `.config/repo-sync/gitignore`.            | bootstrap  | Step 3 and `migrate.md` |
   | No item of `migrate.md` applies, and no "Migration" item of a selected tool applies.                                       | bootstrap  | Step 5                  |
   | No path of "Files that bootstrap no longer ships" exists.                                                                  | bootstrap  | Step 5                  |
   | `mise run setup:configs --dry-run` shows no change.                                                                        | bootstrap  | Step 6                  |
   | If the repository has the task `code:lint:repo-sync`, `mise run code:lint:repo-sync` shows no error in the values file.    | repository | Step 3                  |
   | The selected tools agree with the signs of their tool files: no tool is missing, and no tool is selected without a reason. | repository | Step 1                  |
   | Each value that you can find agrees with the file, for example `repo_url` and `origin`, or the members and `.gitmodules`.  | repository | Step 3                  |
   | Each file of "Files the repository supplies" of a selected tool exists.                                                    | repository | Step 7                  |

5. If all the checks pass, tell the user that the repository is up to date,
   and stop.

### Files that bootstrap no longer ships

An earlier sync wrote these paths, and the sync of today does not write them.
The sync does not delete them, so step 5 deletes them with the item of
`migrate.md`. Delete a file only when it starts with the bootstrap header. A
file without the header belongs to the repository: keep it.

| Path                                                                                                                                                                                 | Do                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------- |
| `.graphifyignore`                                                                                                                                                                    | Delete it (M24).                      |
| `.config/mise/tasks/code/graph`                                                                                                                                                      | Delete it (M24).                      |
| The `post-commit` and `post-merge` git hooks                                                                                                                                         | Uninstall them with pre-commit (M24). |
| `.config/mise/conf.d/_repo/`                                                                                                                                                         | Rename it to `conf.d/zz-repo/` (M17). |
| A `.config/mise/conf.d/*/mise*.toml` that `.config/repo-generated.tsv` does not list and the sync did not write, for example in `conf.d/_base/`, `conf.d/_env/` and `conf.d/claude/` | Delete it (M29).                      |

The sync removes these old paths itself, so they need no step:

- A file with the bootstrap header that `.config/repo-generated.tsv` lists and
  the sync does not write. The first sync has no manifest, so it removes
  nothing.
- A synced file at a path where the sync now needs a folder, for example the
  old `code/format`, `code/lint`, `code/sec` and `setup/configs` files.
- The merge tasks of the other branch model.

When a change removes a shipped file from `repo/tools/`, add its path to this
table and to an item of `migrate.md`.

## Step 1: Select the tools and the packs

Clone `bootstrap` into a temporary directory. Keep it until all repositories
are complete. `--filter=blob:none` gets all the commits, for the `git log` of
an update, but only the files of the newest commit.

```bash
tmp=$(mktemp -d)
git clone --quiet --filter=blob:none git@github.com:virajp/bootstrap.git "$tmp"
```

Read `"$tmp"/repo/tools/*/tools.toml`, `"$tmp"/repo/packs.toml`,
`"$tmp"/docs/sync/tools/<tool>.md` and `"$tmp"/docs/sync/packs/<pack>.md`.
Examine the repository for the signs in "When the tool is relevant" of each
tool file. Show the user each tool and each pack with its description and the
result. Then ask which tools and packs the repository needs. Ask the questions
of each pack file.

The `base` pack is always added: `repo_sync` adds it, and the values file does
not list it. Its tools are `mise`, `repo-sync`, `git`, `pre-commit`, `dprint`,
`linter`, `vscode`, and the tools of the `security` pack. Do not select a tool
of the
`base` pack: a `[tools.<name>]` table for it gives values. An `env` in that
table can only add environments, and the sync examines it. See
[`packs/base.md`](packs/base.md).

A repository with a `package.json` must select one package manager: `pnpm` or
`aube`. If it selects none, or both, `setup:deps:node` stops the setup. See
[`tools/node.md`](tools/node.md).

For each selected tool, and each tool of the `base` pack, do the items in
"Adapt to the repository" of its tool file.

The form of a pack is in [`docs/design/packs.md`](../design/packs.md). The keys
of a `tools.toml` and its `[files]` tables are in
[`docs/design/tools.md`](../design/tools.md). The `[mise.*]` tables and the
environments of
each tool are in [`docs/design/mise.md`](../design/mise.md).

A mise tool with `requires_found`, and a file with `requires` or
`requires_found`
in its `[files]` table, need a file in the repository. For example, the `python`
tool installs `ruff` and `pyrefly`, and writes `code:lint:ruff`,
`code:format:ruff` and `code:lint:pyrefly`, only in a repository with a
`pyproject.toml`. If the repository adds its first `pyproject.toml` later,
run `mise run setup:configs` again.

In a tool file, the "Environments" column of "mise tools" shows the environments
that each mise tool allows. If a tool gets an environment that one of its mise
tools does not allow, the sync stops with code 2.

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

Copy these three files from the clone into the repository: the `setup:configs`
tasks of the `repo-sync` tool, and `helpers` of the `mise` tool. An old
`.config/mise/tasks/setup/configs` is a file, and
`setup/configs/` is now a folder: delete the old file first.

```bash
sync="$tmp/repo/tools/repo-sync/.config/mise/tasks"
lib="$tmp/repo/tools/mise/.config/mise/tasks"
rm -f .config/mise/tasks/setup/configs
mkdir -p .config/mise/tasks/_scripts .config/mise/tasks/setup/configs
cp "$lib/_scripts/helpers" .config/mise/tasks/_scripts/
cp "$sync/setup/configs/sync" .config/mise/tasks/setup/configs/
printf '{"members":[],"sync_engine":"%s/sync-engine"}\n' "$tmp" > "$tmp/context.json"
mise x gomplate@latest -- gomplate --left-delim '<%' --right-delim '%>' \
  -c ".=$tmp/context.json" -f "$sync/setup/configs/_default.tmpl" \
  -o .config/mise/tasks/setup/configs/_default
chmod 755 .config/mise/tasks/setup/configs/_default
```

`setup:configs/_default` is a template. The command above renders it without
the member flags. The first sync writes it again with the member flags. The
template embeds `sync-engine/clone`, so its context needs `sync_engine`, the
absolute path of the `sync-engine` folder of the clone.

Do not run the sync, `mise run setup:configs` without `--dry-run`, before
step 6. You can run a dry run. When the repository already has the task
`code:lint:repo-sync`, for example at an update, you can also run
`mise run code:lint:repo-sync`. It validates the values file with
`repo_sync --check`, without a clone and without a sync, and it shows each
error at once. The task loads the sync code from the clone of the shared cache.
Without that clone, or when the clone has another `origin`, the task prints
nothing, exits 0 and does not validate the values file. Thus no output does not
show that the values file is correct. A dry run makes the clone.

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
   | `[branch.<branch>]`      | Step 4. Without a sign, `merge_model = "direct"` and no `direct_commit`.                                            |
   | `[member.<name>]`        | `.gitmodules`, and the member folders of a monorepo.                                                                |
   | `commit_scopes`          | The scopes of the existing commits, `git log --format=%s`. Else the members.                                        |
   | `worktree_files`         | Gitignored files that a new worktree needs, for example `.config/mise.local.toml`.                                  |
   | `extensions`             | The `recommendations` of `.vscode/extensions.json` that no selected tool has.                                       |
   | `[overrides.gitignore]`  | The lines of the existing `.gitignore` that the shared one does not have.                                           |
   | `[claude.plugins.*]`     | The plugins of `.claude/settings.json` (project scope).                                                             |

2. Use the example of `repo/examples/` with the shape of the repository as the
   model (see "A product with more than one repository").
3. Show the user the full draft of `.config/bootstrap.toml`. Ask whether the
   user wants to change a value.
4. Write the file with the values of the user.

The keys are in
[`docs/design/bootstrap-config.md`](../design/bootstrap-config.md).
Read it, and write the file with these instructions:

- Do not write `bootstrap_commit`, and do not copy it from an example: each
  example has an empty value. `repo_sync` writes the `bootstrap` commit of each
  sync without a failure. See "A repository that already has
  `.config/bootstrap.toml`".
- Write one `[tools.<folder>]` table for each tool from step 1, and one
  `[packs.<name>]` table for each pack from step 1, without the prefix `pack:`.
  Do not write `base`.
- A tool value goes into the `[tools.<name>]` table of that tool, next to `env`.
  The top-level keys of the values file stay at the top level.
- Set `product_name` in each repository of a product with members, the root
  included. See `tools/fnox.md`.
- Write one `[member.<name>]` table for each member folder.
- Do not set `merge_from` with `merge_never`. For `direct_commit`, see step 4.
- For the format of `[claude.plugins.*]`, see `tools/claude-tool.md`.

### Overrides

`[overrides]` replaces a default of a tool. The path after `overrides.` is the
path in `repo/tools/<tool>/tools.toml`, without the tool name. The sync applies
the override to each selected tool that has a value at that path.

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

An override cannot add a key. Put a new key in
`conf.d/zz-repo/mise.<env>.toml`. The other rules of `[overrides.mise*]` are in
[`docs/design/mise.md`](../design/mise.md#the-overrides).

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
- The old table `[overrides.graphifyignore]` also stops the sync with code 2.
  See M24 of `migrate.md`.

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
4. Ask the user whether a core branch allows a direct commit. If it does, set
   `direct_commit = true` in its `[branch.<branch>]` table. With both
   branches, `main` can allow it only when `develop` allows it too. Without a
   sign, do not write the key.

`repo_sync` does not ask. It writes the merge tasks of the model that it
finds, and removes the merge tasks of the other model. If the branches change
later, run `setup:configs` again.

The `no-commit-to-branch` hook of the `git` tool protects only the core
branches that exist and do not allow direct commits. The effect of
`direct_commit` on the hook and the checks of the key are in
[`docs/design/repo-sync.md`](../design/repo-sync.md#direct_commit).

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
2. Run `mise run setup:configs --dry-run`. To validate only the values file,
   without a sync, run `mise run code:lint:repo-sync` when the repository has
   that task. It does not examine the rules that need the `bootstrap` clone,
   for example a tool that is not in `bootstrap`. Without a clone in the shared
   cache, or with another `origin` there, it prints nothing and does not
   validate. Thus do the dry run first.
3. Examine the diff for each file that already exists in the repository.
4. If the dry run replaces repository-owned content, stop. Move that content
   to a repository-owned place (see "Rules"), then do the dry run again.
5. If the dry run fails with `pre-commit is not available`, and its diff adds
   a `#MISE tools.pre-commit` line to `.config/mise/tasks/setup/configs/sync`,
   the cause is the old task file
   of the repository, not the values file. `setup:configs` runs the
   repository's own copy of that task, and only a real sync writes the new
   copy. Thus the dry run cannot pass before a real sync. Continue with item
   6, and at item 7 use the two runs of the note under the list.
6. Ask the user whether to run `mise run setup:configs`. In the same
   question, ask whether to commit the changes at the end (step 9), or to
   leave them for review.
7. If the user agrees, run `mise run setup:configs`. A sync without a failure
   removes the stale files, writes `.config/repo-generated.tsv` and sets
   `bootstrap_commit` in the values file.
8. Run `mise run init`. This makes all the task scripts executable.

**An old task file (item 5).** Run `mise run setup:configs` two times:

1. The first run writes the new task file and the other files. It fails on
   the file that needs the tool, keeps the old copy of that file, and does not
   set `bootstrap_commit`.
2. The second run uses the new task file, so mise installs the tool before
   the task runs. This run must pass. If it fails again, read the exit code
   below.

If `repo_sync` fails, read the exit code:

| Code | Cause                                                                                                                                                                                                                                                                                                    | Action                                                                                                                                                                                           |
| ---- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 2    | The old list format, a bad `[tools.*]` or `[packs.*]` entry, a bad `env` (also an `env` without `dev`) or `environments`, a bad `[overrides]`, a root-only tool in a submodule, or a template that cannot render                                                                                         | Correct the values file                                                                                                                                                                          |
| 2    | A tool in an environment that one of its mise tools does not allow                                                                                                                                                                                                                                       | Correct the `env`, or tell the `bootstrap` owner                                                                                                                                                 |
| 2    | A pack loop, a pack name that is also a tool folder, or an old key or a bad member of `repo/packs.toml`                                                                                                                                                                                                  | Tell the `bootstrap` owner                                                                                                                                                                       |
| 2    | A `default_envs` without `dev`, or a tool folder that is `zz-repo` or that sorts after `zz-repo`                                                                                                                                                                                                         | Tell the `bootstrap` owner                                                                                                                                                                       |
| 2    | Two tools define the same mise tool, a mise `depends` that no tool defines, a tool that ships and generates the same `conf.d` file, an unsupported `[mise.*]` or `[mise_env.*]` section, an old key or a bad `[files]` table of a `tools.toml`, or a `default_envs` or `envs` name that is not dev or ci | Tell the `bootstrap` owner                                                                                                                                                                       |
| 2    | A bad `[defaults]`, `[questions]` or `[contributes]` table of a selected tool, or two contributions that set one key to different values. The sync writes no file of a tool. Inside one target tool, only the first conflict shows                                                                       | Tell the `bootstrap` owner                                                                                                                                                                       |
| 2    | A value of `[tools.<name>]` whose type differs from the merged value of a base or dependency tool. The sync writes no file of a tool                                                                                                                                                                     | Correct the value in the values file                                                                                                                                                             |
| 2    | A merge source or a live merge target cannot merge, or a merge target is a symlink or is below a symlinked folder. The existing file stays, and the other files continue                                                                                                                                 | Read the `repo-sync:` line. Correct the live file or the symlink, or tell the `bootstrap` owner                                                                                                  |
| 2    | A JSON, YAML or TOML target does not parse. The existing file stays, and the other files continue                                                                                                                                                                                                        | Read the `repo-sync:` line. A mise failure indented under that line means that a tool could not start: examine mise and the network. Else correct the values file, or tell the `bootstrap` owner |
| 2    | `.config/repo-generated.tsv` is a symlink, is below a symlinked folder or is not a regular file. The other files stay written. The sync runs no stale rule, writes no manifest and does not set `bootstrap_commit`                                                                                       | Replace it with a regular file, or delete it                                                                                                                                                     |
| 2    | The values file is missing, or mise is not available                                                                                                                                                                                                                                                     | Correct the values file, or install mise                                                                                                                                                         |
| 2    | Neither `main` nor `develop` exists                                                                                                                                                                                                                                                                      | Do step 4                                                                                                                                                                                        |
| 2    | A git repository at the root of the cache with another `origin` or no `origin`, or a `repo/` of the cache that is not a clone and not empty                                                                                                                                                              | Move it, or set `SYNC_ENGINE_CACHE` to another folder                                                                                                                                            |
| 2    | `setup:configs:sync` ran without `SYNC_ENGINE_CLONE`                                                                                                                                                                                                                                                     | Run `mise run setup:configs`                                                                                                                                                                     |
| 3    | A tool or a pack that is not in `bootstrap`                                                                                                                                                                                                                                                              | Correct the `[tools.*]` and `[packs.*]` entries                                                                                                                                                  |
| 3    | A member of a pack that is not a tool folder or a pack                                                                                                                                                                                                                                                   | Tell the `bootstrap` owner                                                                                                                                                                       |
| 3    | A contribution of a selected tool to a tool that is not in `bootstrap`                                                                                                                                                                                                                                   | Tell the `bootstrap` owner                                                                                                                                                                       |
| 5    | The clone or the fetch failed, `SYNC_ENGINE_CLONE` is not usable, pre-commit is not available (see item 5 of step 6), or `pre-commit autoupdate` failed                                                                                                                                                  | Examine the network and the access                                                                                                                                                               |
| 1    | A local file system error                                                                                                                                                                                                                                                                                | Read the `repo-sync:` lines on stderr                                                                                                                                                            |

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
4. Run `mise run setup:configs --verbose`. Each file must show `unchanged`,
   `kept seed`, `skip seed` or `skip env`.

mise always loads `.config/mise.toml`. It loads a `conf.d/*/mise.<env>.toml`
file only when `MISE_ENV` includes that environment. Without `MISE_ENV`, no
file in `conf.d` loads.

## Step 9: Commit

If the user chose to leave the changes for review in step 6, show
`git status --short` in each repository, and stop. Else commit these files:

- The synced files and the seed files.
- `.config/bootstrap.toml` and `.config/repo-sync/pre-commit.yaml`.
- `.config/repo-generated.tsv`.
- The repository-owned files from step 5 and step 7.

Commit in each submodule before you commit in the repository that contains
it. Then the commit in the containing repository includes the new submodule
commit. Ask the user before you push. Push each submodule before the
repository that contains it. If not, the containing repository points to a
commit that the remote does not have.
