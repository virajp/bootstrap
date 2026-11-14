# Adopt bootstrap in a repository, or bring it up to date

These instructions are for an AI agent that works in a consumer repository.
Do the steps in sequence to make the repository get its shared files from
`bootstrap` (`https://github.com/virajp/bootstrap.git`). The agent uses the
CLI `bootstrap`. The contract of the sync is in `docs/design/` of `bootstrap`
(see "Where the rules are").

This file is a prompt. A user or an agent of a consumer repository can use it
directly, for example: "Read `docs/sync/index.md` of bootstrap and do its
steps in this repository." The `bootstrap: init` instruction of
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
   for the adoption, and again when the repository is not up to date. The
   agent makes sure that each rule is relevant to the repository: it selects
   only the tools that the repository needs, it keeps the repository-owned
   content, and it asks the user about each difference. The agent runs
   `bootstrap init` to write the values file and to do the first sync, and
   `bootstrap add tool <tool…>` to add the other tools.
2. **Sync, each time after that, by `setup:configs`.** The `mise` tool ships
   the task `setup:configs`. The task runs `bootstrap sync`, which copies the
   newest shared files of the selected tools into the repository.

The adoption makes sure that the repository is ready for the sync. The sync
then keeps the repository equal to `bootstrap`, without an agent.

After the adoption, `bootstrap add tool <tool…>`, `bootstrap add pack <pack…>`,
`bootstrap remove tool <tool…>` and `bootstrap replace tool <old> <new>` change
the tools. See [`vwf.md`](vwf.md).

## Files in this folder

| File                | Use                                                                                                                             |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `index.md`          | This file: the general steps of an adoption.                                                                                    |
| `migrate.md`        | The items for a repository that has older copies, or the old name.                                                              |
| `kotlin-dsl.md`     | A static prompt that migrates a Gradle build from the Groovy DSL to the Kotlin DSL.                                             |
| `./tools/<tool>.md` | One file for each folder in `repo/tools/`: when the tool is relevant, and usage.                                                |
| `packs/<pack>.md`   | One file for each pack of `repo/packs.toml`: the tools that it adds, and its questions.                                         |
| `vwf.md`            | The entry point for the vwf and stackgen plugins: the `init`, `add`, `remove`, `add plugins` and `remove plugins` instructions. |

The tool files in `tools/` are `_security_base.md`, `android.md`, `apko.md`,
`aube.md`, `bun.md`, `claude-tool.md`, `cloudflare.md`, `cosign.md`, `crane.md`,
`docker.md`, `doppler.md`, `dprint.md`, `firebase.md`, `flutter.md`, `fnox.md`,
`gcloud.md`, `git.md`, `github.md`, `gitlab.md`, `gitleaks.md`, `go.md`,
`grype.md`, `java.md`, `kotlin.md`, `linter.md`, `melange.md`, `mise.md`,
`node.md`, `osv-scanner.md`, `pnpm.md`, `pre-commit.md`, `pulumi.md`,
`python.md`, `rust.md`, `swift.md` and `vscode.md`. The file of the `claude`
tool is `claude-tool.md`, because a file system that ignores case reads
`claude.md` as `CLAUDE.md`, and AI agents load that name as instructions.

The pack file in `packs/` is `base.md`.

## Where the rules are

This file holds the instructions. Each rule of the sync has one place in
`docs/design/` of the clone of step 1. Read the doc when a step needs the rule:

- Read [`docs/design/bootstrap-config.md`](../design/bootstrap-config.md) before
  you write a key of `.config/bootstrap.toml`.
- Read [`docs/design/packs.md`](../design/packs.md) before you add a pack.
- Read [`docs/design/tools.md`](../design/tools.md) before you read a
  `tools.toml`: its keys, the `[files]` and `[tasks]` tables, the shared
  files, the scripts, the seeds and the header.
- Read [`docs/design/mise.md`](../design/mise.md) before you select an
  environment or write `[overrides.mise*]`.
- Read [`docs/design/repo.md`](../design/repo.md) before you select the
  branch model or the members.
- Read [`docs/design/cli.md`](../design/cli.md) for the
  commands of `bootstrap`, the clone, a parse failure, a stale file or an exit
  code.
- Read [`docs/design/fnox.md`](../design/fnox.md) before you select `fnox`.

## Rules

- Ask each question of this file. Do not skip a question: only the user can
  skip one. Before you ask, find each value that you can, and show it as the
  default. A value of a `bootstrap: init` instruction is a default too.
- Ask the user before each choice that this file does not decide. Do not
  use a tool, a value, or a path that the user did not see.
- Apply only what is relevant to the repository. Before you select a tool,
  read its tool file. Examine the repository for the signs in "When the tool
  is relevant", then show the user the result.
- The shared files are grouped by tool. Each tool is a folder
  `repo/tools/<tool>/` in `bootstrap`. Its `tools.toml` gives the path in the
  consumer repository of each file of its `files/` and `tasks/` folders.
- The values file `.config/bootstrap.toml` has one `[tools.<name>]` table for
  each tool of the repository, and no pack. `bootstrap sync` always adds the
  required tools of the `base` pack, and it syncs all the files of these
  tools.
- A command of `bootstrap` that writes shows the change list first. An AI
  agent has no terminal, so it gives `--yes` after the user agrees to the
  list. Run the command with `--dry-run` first to get the list.
- In a terminal, `bootstrap init` opens the TUI, which is for a person. The
  CLI is for an agent: `--no-tui`, `--yes`, `--dry-run` and `--answer` each
  keep `bootstrap init` on the CLI path. Do not run `bootstrap tui`.
- `bootstrap` writes no file that has changes that are not committed. It
  stops with code 2 and names the file. Thus commit or discard the changes of
  each target before a command that writes. Ask the user before each commit.
- Do not edit a synced file in the consumer repository. The next sync replaces
  it. Put repository-owned content in the places that the sync never writes:
  - a different default of a mise tool, for example a version, a mise setting
    or an env value: `[overrides.mise*]` of the values file (see step 6).
  - mise config and environment variables that no tool has: the seeds
    `.config/mise/conf.d/zz-repo/mise.<env>.toml`, one for each environment.
    mise loads the folders of `conf.d/` in name order, and a later folder
    wins. The name `zz-repo` sorts last, so the repository config wins over
    the config of each tool.
  - shell functions: `.config/mise/tasks/_scripts/local`.
  - tasks: any task name that no tool ships.
  - formatters and linters: a task `code:format:<name>` or `code:lint:<name>`
    (`.config/mise/tasks/code/format/<name>`). `code:format` and `code:lint`
    run each of them with the same `--fix` and files. A pre-commit hook only
    calls a task, so do not call a formatter or a linter from a hook.
- A shared file is different. The sync merges the items of the selected tools
  into it, and it keeps the other items. Thus put the repository-owned items
  directly in the file: ignore lines in `.gitignore`, hooks in
  `.config/pre-commit-config.yaml`, extensions in `.vscode/extensions.json`,
  and keys, excludes and plugins in `dprint.jsonc`. To change or remove an
  item of `bootstrap`, write an override in
  `[tools.<tool>.files."<path>"]` of the values file (see step 6).
- A seed file is different too. The sync writes it only if it is missing.
  After that, the repository owns it.
- To change a shared file of `bootstrap`, change it in `bootstrap`, push it to
  `main`, then sync again.
- Prefix each tool call with `mise x --`, also in the task files and the hook
  entries of `.config/pre-commit-config.yaml`.

## Requirements

- The repository is a git repository. `bootstrap` stops outside of one.
- mise is installed.
- The CLI `bootstrap` is installed. Install it with
  `mise use -g github:virajp/bootstrap`. This install
  needs the network.
  After the adoption, the `mise` tool also installs it in `dev`.
- The machine can reach `https://github.com/virajp/bootstrap.git`.
  `bootstrap` clones this URL into `~/.cache/bootstrap/`, and needs no SSH
  key. `BOOTSTRAP_CACHE`, or the key `cache_dir` of the user config file
  `~/.config/bootstrap/config.toml`, changes the folder. `bootstrap config`
  shows each setting of the CLI and its source. See
  [The config file](../design/cli.md#the-config-file).
- When `bootstrap` is older than the version that the clone needs, it stops.
  Upgrade it with `mise use -g github:virajp/bootstrap`.

## An empty repository

If the repository has no commit, do these steps first:

1. Write `.gitignore` with these two lines:

   ```text
   .worktrees/
   .claude/worktrees/
   ```

2. Ask the user whether to commit it on `main`. This is the first commit.
   Without a commit, no branch exists, and the sync stops (step 2).

Then do the other steps. A worktree in `.worktrees/` is now ignored, and the
branch `main` exists. The first sync adds the shared lines to `.gitignore`,
and it keeps these two lines.

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

A tool with `root_only = true` is for the base repository only. The sync refuses
it in a submodule. `bootstrap init` and `bootstrap add pack <pack…>` skip a
`root_only` member of a pack in a submodule, for example `vscode` of `base`.
When a submodule is in its root checkout, mise also loads the tools of the root.
Thus a nested submodule also finds the tools of the root.

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
change a file. `bootstrap init` on an existing values file is the update: it
examines the version, migrates each old form of the values file, asks each
new question, then syncs. A values file without `format = 2`, or with a
`[packs.*]` table, is an old form: `bootstrap init` moves it to a
`[tools.<name>]` table for each tool (M38 of `migrate.md`).

1. Make the clone of step 1, and get the newest commit of `bootstrap`:
   `git -C "$tmp" rev-parse HEAD`.
2. Read `bootstrap_commit` of the values file. `bootstrap sync` writes it
   after each sync without a failure.
3. Use `bootstrap_commit` to select the checks of step 4:

   | `bootstrap_commit` | The checks of the kind `bootstrap`                                                                                             | The checks of the kind `repository` |
   | ------------------ | ------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------- |
   | The same as `main` | Skip them. The shared files are current.                                                                                       | Do them.                            |
   | Older than `main`  | Do them only for the tools and the migrate items that changed: `git -C "$tmp" log --oneline <commit>..HEAD -- repo docs/sync`. | Do them.                            |
   | Missing or empty   | Do them all.                                                                                                                   | Do them.                            |

   The checks of the kind `repository` are always necessary: the repository
   can change also when `bootstrap` does not, for example with a new
   `pyproject.toml` or a new submodule.
4. Do the selected checks of this table. Each check that fails sends you to
   a step:

   | Check                                                                                                                      | Kind       | If it fails, do     |
   | -------------------------------------------------------------------------------------------------------------------------- | ---------- | ------------------- |
   | No item of `migrate.md` applies, and no "Migration" item of a selected tool applies.                                       | bootstrap  | Step 4              |
   | No path of "Files that bootstrap no longer ships" exists.                                                                  | bootstrap  | Step 4              |
   | `bootstrap init --dry-run` shows no change. It shows the migration of an old form, a new question and each file to sync.   | bootstrap  | Step 3, then step 5 |
   | `bootstrap check` shows no error in the values file.                                                                       | repository | Step 6              |
   | `bootstrap detect` exits 0: each tool that a detect script finds is selected.                                              | repository | Step 1              |
   | The selected tools agree with the signs of their tool files: no tool is missing, and no tool is selected without a reason. | repository | Step 1              |
   | Each value that you can find agrees with the file, for example `repo_url` and `origin`, or the members and `.gitmodules`.  | repository | Step 6              |
   | Each file of "Files the repository supplies" of a selected tool exists.                                                    | repository | Step 7              |

5. If all the checks pass, tell the user that the repository is up to date,
   and stop.

For an update, step 5 runs `bootstrap init`. Add a new tool with
`bootstrap add tool <tool…>`.

### Files that bootstrap no longer ships

An earlier sync wrote these paths, and the sync of today does not write them.
Step 4 deletes them with the item of `migrate.md`. Delete a file only when it
starts with the bootstrap header. A file without the header belongs to the
repository: keep it.

| Path                                                                                                                                                                                 | Do                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------- |
| `.graphifyignore`                                                                                                                                                                    | Delete it (M24).                      |
| `.config/mise/tasks/code/graph`                                                                                                                                                      | Delete it (M24).                      |
| The `post-commit` and `post-merge` git hooks                                                                                                                                         | Uninstall them with pre-commit (M24). |
| `.config/mise/conf.d/_repo/`                                                                                                                                                         | Rename it to `conf.d/zz-repo/` (M17). |
| A `.config/mise/conf.d/*/mise*.toml` that `.config/repo-generated.tsv` does not list and the sync did not write, for example in `conf.d/_base/`, `conf.d/_env/` and `conf.d/claude/` | Delete it (M29).                      |
| `.config/mise/tasks/setup/configs/sync`                                                                                                                                              | Delete it (M31, M34).                 |
| `.config/repo-sync/gitignore`                                                                                                                                                        | Move its lines to `.gitignore` (M18). |
| `.config/repo-sync/pre-commit.yaml`                                                                                                                                                  | `bootstrap init` moves it (M34).      |
| The files of the old `repo-sync` tool: `.config/mise/tasks/_scripts/repo-sync`, `.config/mise/tasks/code/lint/60-repo-sync` and `.config/mise/conf.d/repo-sync/`                     | Delete them (M34).                    |
| `.config/mise/tasks/code/format/20-gofmt`                                                                                                                                            | Delete it (M39).                      |

The sync removes these old paths itself, so they need no step:

- A file with the bootstrap header that `.config/repo-generated.tsv` lists and
  the sync does not write. The first sync has no manifest, so it removes
  nothing.
- A synced file at a path where the sync now needs a folder, for example the
  old `code/format`, `code/lint`, `code/sec` and `setup/configs` files.
- An item of a shared file that `bootstrap` added and that no selected tool
  has now.

When a change removes a shipped file from `repo/tools/`, add its path to this
table and to an item of `migrate.md`.

## Step 1: Select the tools

Clone `bootstrap` into a temporary directory to read its docs. Keep it until
all repositories are complete. `--filter=blob:none` gets all the commits, for
the `git log` of an update, but only the files of the newest commit.

```bash
tmp=$(mktemp -d)
git clone --quiet --filter=blob:none https://github.com/virajp/bootstrap.git "$tmp"
```

Read `"$tmp"/repo/tools/*/tools.toml`, `"$tmp"/repo/packs.toml`,
`"$tmp"/docs/sync/tools/<tool>.md` and `"$tmp"/docs/sync/packs/<pack>.md`.
Examine the repository for the signs in "When the tool is relevant" of each
tool file. Show the user each tool with its description and the result. Then
ask which tools the repository needs. In a repository with a values file,
`bootstrap detect` also lists each tool that a detect script finds and that is
not selected.

The `base` pack is the one pack of `repo/packs.toml`. `bootstrap init` writes a
table for each of its members in a new values file. Its tools are `mise`,
`git`, `pre-commit`, `dprint`, `linter`, `gitleaks`, `grype`, `osv-scanner`
and `vscode`. Each of them except `vscode` is required: the repository cannot
remove it, and its table gives values. An `env` in that table can only add
environments, and the sync examines it. Ask whether the repository needs
`vscode`. See [`packs/base.md`](packs/base.md).

A repository with a `package.json` must select one package manager: `pnpm` or
`aube`. If it selects none, or both, `setup:deps:40-node` stops the setup. See
[`tools/node.md`](tools/node.md).

For each selected tool, and each tool of the `base` pack, do the items in
"Adapt to the repository" of its tool file.

The form of a pack is in [`docs/design/packs.md`](../design/packs.md). The keys
of a `tools.toml` are in [`docs/design/tools.md`](../design/tools.md). The
`[mise.*]` tables and the environments of each tool are in
[`docs/design/mise.md`](../design/mise.md).

A file with `requires` or `requires_found` in its `[files]` or `[tasks]`
table needs a file in the repository. For example, the `python` tool writes
`code:lint:70-ruff`, `code:format:40-ruff` and `code:lint:50-pyrefly` only in
a repository with a `pyproject.toml`. If the repository adds its first
`pyproject.toml` later, run `mise run setup:configs` again.

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

## Step 2: Select the branch model

The merge tasks follow the branches `main` and `develop`. A branch counts if it
exists locally or on `origin`.

| Branches             | Merge tasks                             | Rule                                                                      |
| -------------------- | --------------------------------------- | ------------------------------------------------------------------------- |
| `main` and `develop` | `code:merge:develop`, `code:merge:main` | Feature branches merge into `develop`. Only `develop` merges into `main`. |
| only one of them     | `code:merge`                            | Any feature branch merges into that branch.                               |
| neither              | none                                    | The check script of the `git` tool stops the sync with code 2.            |

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
4. Ask the user whether a core branch allows a direct commit. If it does, step
   6 sets `direct_commit = true` in its `[tools.git.branch.<branch>]` table.
   With both branches, `main` can allow it only when `develop` allows it too.
   Without a sign, do not write the key.
5. Ask the user for the merge model of each core branch: `direct` (the
   default) or `pr`. Step 6 writes it as `merge_model`.

`bootstrap sync` does not ask. It writes the merge tasks of the model that it
finds, and removes the merge tasks of the other model. If the branches change
later, run `setup:configs` again.

The `no-commit-to-branch` hook of the `git` tool protects only the core
branches that exist and do not allow direct commits. The rules of
`direct_commit` and of the merge keys are in
[`docs/design/repo.md`](../design/repo.md) and
[`tools/git.md`](tools/git.md).

## Step 3: Answer the questions

A tool can have questions. Their answers are keys of its `[tools.<tool>]`
table. Get the questions of the tools that the values file selects (without a
values file, the tools of the `base` pack), then of the tools of step 1:

```text
bootstrap questions --json
bootstrap questions <tool…> --json
```

Each item has `tool`, `key`, `question`, `proposal`, `source`, `answer`,
`group`, `note` and `optional`. `proposal` is the value that `bootstrap` finds
in the repository, and `source` tells where it comes from. `answer` is the
value of the values file, or `null`. An `answer` of `""` means no answer, and
the question is answered. `note` is a line to show under the question.

1. For each item with a `null` answer, ask the user the question. Show the
   proposal as the default.
2. Keep each answer for step 5, as `--answer <tool>.<key>=<value>`.
3. The user can leave an item with `optional` = `true` empty. Keep it as
   `--answer <tool>.<key>=`, so a later `bootstrap init` does not ask it.
4. The items with the same `tool` and a `group` that is not empty need at
   least one answer that is not empty. Keep each empty answer of the group as
   `--answer <tool>.<key>=`. Another item needs an answer that is not empty.

Without questions, the command prints `[]`.

## Step 4: Migrate the existing files

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
5. `bootstrap` writes no target that has changes that are not committed.
   Thus ask the user to commit the changes of this step before step 5.

## Step 5: Run `bootstrap init` and `bootstrap add tool <tool…>`

1. If mise says that the config files are not trusted, ask the user to run
   `mise trust`.
2. Run `bootstrap init` with the answers of step 3 for the tools of the `base`
   pack:

   ```text
   bootstrap init --answer <tool>.<key>=<value> --dry-run
   ```

   `bootstrap init` takes no tool names. For an update, this is the only command
   of this step.
3. Examine the change list, and the diff of each file that already exists in
   the repository.
4. If the dry run replaces repository-owned content, stop. Move that content
   to a repository-owned place (see "Rules"), then do the dry run again.
5. Ask the user whether to apply the change list. In the same question, ask
   whether to commit the result of each step (step 9), or to leave the
   changes for review.
6. If the user agrees, run the same command with `--yes` in place of
   `--dry-run`.
7. Do items 3 to 6 again with `bootstrap add tool <tool…>`, the other selected
   tools of step 1, and their answers of step 3:

   ```text
   bootstrap add tool <tool…> --answer <tool>.<key>=<value> --dry-run
   ```

   If the repository does not need `vscode`, also run
   `bootstrap remove tool vscode`.

Without a values file, `bootstrap init` writes it with `format = 2`,
`environments` and `repo_name`, the name of the folder. It adds a
`[tools.<name>]` table for each member of the `base` pack and each answer, then
syncs. `bootstrap add tool <tool…>` adds a table for each named tool and each
new dependency, and each answer, then syncs. A sync without a failure removes
the stale files, writes `.config/repo-generated.tsv` and sets `bootstrap_commit`
in the values file.

## Step 6: Write the other values and sync again

1. Find each value of the table below that you can. These are the usual
   places:

   | Value                              | Where to find it                                                                                                    |
   | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
   | `repo_name`                        | The last part of the path of `origin`, without `.git`. Else the folder name, but not the name of a worktree folder. |
   | `repo_url`                         | `origin`, as an `https://` URL.                                                                                     |
   | `product_name`                     | `product.name` of `.config/vwf.yaml`, or the value of the other repositories of the product.                        |
   | `environments`                     | Step 1.                                                                                                             |
   | `env` of `[tools.*]`               | Step 1.                                                                                                             |
   | `[tools.git.branch.<branch>]`      | Step 2. Without a sign, `merge_model = "direct"` and no `direct_commit`.                                            |
   | `[member.<name>]`                  | `.gitmodules`, and the member folders of a monorepo.                                                                |
   | `commit_scopes` of `[tools.git]`   | The scopes of the existing commits, `git log --format=%s`. Else the members.                                        |
   | `worktree_files` of `[tools.mise]` | Gitignored files that a new worktree needs, for example `.config/mise.local.toml`.                                  |
   | `[tools.claude.plugins.*]`         | The plugins of `.claude/settings.json` (project scope).                                                             |

2. Use the example of `repo/examples/` with the shape of the repository as the
   model (see "A product with more than one repository").
3. Show the user the full draft of `.config/bootstrap.toml`. Ask whether the
   user wants to change a value.
4. If the result of step 5 is not committed, ask the user to commit it now.
   The sync writes no target that has changes that are not committed. The
   values file is not a target.
5. Write the file with the values of the user.
6. Run `bootstrap check`. It must show no error.
7. Run `bootstrap sync --dry-run`, and examine the diff. Then run
   `mise run setup:configs`.

The keys are in
[`docs/design/bootstrap-config.md`](../design/bootstrap-config.md).
Read it, and write the file with these instructions:

- Do not write `bootstrap_commit`, and do not copy it from an example: each
  example has an empty value. `bootstrap sync` writes the `bootstrap` commit
  of each sync without a failure.
- Do not write a `[packs.<name>]` table: it is an old form. Keep the
  `[tools.<name>]` tables and `format = 2` of step 5.
- A tool value goes into the `[tools.<name>]` table of that tool, next to `env`.
  The top-level keys of the values file stay at the top level.
- Set `product_name` in each repository of a product with members, the root
  included. See `tools/fnox.md`.
- Write one `[member.<name>]` table for each member folder.
- Do not set `merge_from` with `merge_never`. For `direct_commit`, see step 2.
- For the format of `[tools.claude.plugins.*]`, see `tools/claude-tool.md`.

### Overrides of a mise default

`[overrides]` replaces a default of a mise tool, a mise setting or an env
value. The path after `overrides.` is the path in
`repo/tools/<tool>/tools.toml`, without the tool name. The sync applies the
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

An override cannot add a key. Put a new key in
`conf.d/zz-repo/mise.<env>.toml`. The only keys of `[overrides]` are `mise`
and `mise_env`. The other rules are in
[`docs/design/mise.md`](../design/mise.md).

### Overrides of an item of a shared file

Put a repository-owned item directly in the shared file. To change or remove
an item of `bootstrap`, write an override in `[tools.<tool>.files."<path>"]`,
with the `<tool>` that owns the file:

```toml
# Remove a line of the shared .gitignore.
[tools.git.files.".gitignore"]
"*.log" = { "$mode" = "remove" }

# Remove the conventional-commits hook block.
[tools.pre-commit.files.".config/pre-commit-config.yaml".repos]
"conventional-commits" = { "$mode" = "remove" }
```

`"$mode" = "remove"` removes the item. `"$skip" = true` tells the sync not to
add a missing item. The rules of the merge are in
[`docs/design/tools.md`](../design/tools.md).

## The exit codes

If a command of `bootstrap` fails, read its messages and its exit code. A
failure does not change an existing file. The full list is in
[`docs/design/cli.md`](../design/cli.md).

| Code | Cause                                                                                                                              | Action                                                        |
| ---- | ---------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| 2    | An old form of the values file, also a file without `format = 2`. The message ends with `run bootstrap init`                       | Run `bootstrap init` (step 5)                                 |
| 2    | A target has changes that are not committed                                                                                        | Commit or discard the changes of the target                   |
| 2    | A question without an answer, or a group without an answer that is not empty, with `--yes` or without a terminal                   | Give the answer with `--answer <tool>.<key>=<value>` (step 3) |
| 2    | `changes need --yes`                                                                                                               | Ask the user, then run the command with `--yes`               |
| 2    | A bad key or value of the values file, a root-only tool in a submodule, or a check script that failed. The message names the cause | Correct the values file, or the repository                    |
| 2    | A bad `tools.toml`, packs file or merge spec of `bootstrap`                                                                        | Tell the `bootstrap` owner                                    |
| 2    | `bootstrap` is older than the version that the clone needs                                                                         | Run `mise use -g github:virajp/bootstrap`                     |
| 2    | Not a git repository, or mise is not available                                                                                     | Run `git init`, or install mise                               |
| 3    | A tool or a pack that is not in `bootstrap`                                                                                        | Correct the `[tools.*]` entries, or the names of the command  |
| 5    | The clone or the fetch failed, or `cli/VERSION` of the clone is not `MAJOR.MINOR.PATCH`                                            | Examine the network and the access                            |
| 1    | A local file system error. `bootstrap detect` also exits 1 for a detected tool that is not selected                                | Read the `bootstrap:` lines on stderr                         |

## Step 7: Supply the files that `bootstrap` does not have

The shared files refer to files that each repository must supply. Make sure
that the repository has each file for the tools that it uses. If a file is
missing, ask the user for its content. Do not write a placeholder that hides
the problem.

Each tool file lists these files in "Files the repository supplies". Do this
for each selected tool.

## Step 8: Make sure that the result is correct

1. Run `mise run init`. This makes all the task scripts executable.
2. For each selected environment, run `MISE_ENV=<env> mise config ls`. Make
   sure that mise loads the `conf.d/*/mise.<env>.toml` files of that
   environment.
3. Run `mise tasks`. Make sure that the synced tasks are in the list.
4. Run `mise run setup`. Make sure that it completes.
5. Run `mise run setup:configs --verbose`. Each file must show `unchanged`,
   `kept seed`, `skip seed`, `skip merge` or `skip env`.

mise always loads `.config/mise.toml`. It loads a `conf.d/*/mise.<env>.toml`
file only when `MISE_ENV` includes that environment. Without `MISE_ENV`, no
file in `conf.d` loads.

## Step 9: Commit

If the user chose to leave the changes for review in step 5, show
`git status --short` in each repository, and stop. Else commit these files:

- The synced files, the shared files and the seed files.
- `.config/bootstrap.toml`.
- `.config/repo-generated.tsv`.
- The repository-owned files from step 4 and step 7.

Commit in each submodule before you commit in the repository that contains
it. Then the commit in the containing repository includes the new submodule
commit. Ask the user before you push. Push each submodule before the
repository that contains it. If not, the containing repository points to a
commit that the remote does not have.
