# `.config/bootstrap/` design

This document describes the values folder `.config/bootstrap/` of a repository,
each key of its files, and their old forms. It is the one place of these keys.
`bootstrap migrate` writes the folder in a consumer repository, and the adopting
AI gives the answers (see `docs/sync/index.md`). The sync and the templates of
`repo/tools/` read it. `bootstrap config get|set|unset|show` reads and writes
each key (see [cli.md](cli.md#the-command-line)). The keys of the Mac values
file are in [mac.md](mac.md#macostoml).

See also: [repo.md](repo.md), [packs.md](packs.md),
[tools.md](tools.md),
[mise.md](mise.md), [cli.md](cli.md) and [fnox.md](fnox.md).
The exit code of each rule that stops the sync or the render is in
[cli.md](cli.md#the-exit-codes).

## The files of the folder

| File                 | Holds                                                                                                                       | Written by                                        |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| `repo.toml`          | The facts of the repository: names, branches, commits, members, environments                                                | `bootstrap migrate`, `bootstrap config`, the sync |
| `tools.toml`         | The tool selection, the values of each tool, and the overrides                                                              | `bootstrap migrate`, `add`, `remove`, `config`    |
| `releases.toml`      | The release level of each project. Optional: only `bootstrap config set` makes it                                           | `bootstrap config set releases.<project> <level>` |
| `repo-generated.tsv` | The manifest of the sync. See [repo.md](repo.md#the-manifest)                                                               | The sync                                          |
| `config-files.toml`  | The global file lists of the gates, from `code_files` of the `mise` tool. See [mise.md](mise.md#the-file-sets-of-the-gates) | The sync (a template file of the `mise` tool)     |

The three TOML files `repo.toml`, `tools.toml` and `releases.toml` carry
`format = 3`. The engine writes `format` first, after the header comments, in
a block of its own: a blank line follows it, because the dprint `sortKeys` of a
consumer sorts only the keys of one block. `bootstrap_commit` goes in the next
block. The manifest and `config-files.toml` have no format line. The
engine reads only these three, so it never reads `config-files.toml` as a
values file.

## Example

`repo.toml`:

```toml
format = 3

bootstrap_commit = ""
core_branches = [ "main", "develop" ]
default_branch = "develop"
environments = [ "dev", "ci" ]
product_name = "my-product"
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"

[branches.main]
landing = "pr"
merge_from = [ "develop" ]

[branches.develop]
landing = "direct-merge"
merge_never = [ "main" ]

[commits]
scopes = [ "backend", "web" ]

[worktree]
files = [ ".config/mise.local.toml" ]

[members.web]
folder = "projects/web"
label = "Web"
type = "folder"
```

`tools.toml`:

```toml
format = 3

[tools.claude]

[tools.claude.plugins.project]
"vwf@my-marketplace" = "github:me/claude-plugins"

[tools.dprint]

[tools.fnox]

[tools.git]

[tools.github]

[tools.gitleaks]

[tools.grype]

[tools.linter]

[tools.mise]

[tools.node]
env = [ "dev", "ci" ]

[tools.osv-scanner]

[tools.pre-commit]

[tools.vscode]
```

`releases.toml`:

```toml
backend = "MINOR"
format = 3
web = "PATCH"
```

## Summary of the keys

`repo.toml`:

| Key                                 | Type            | Necessary | Read by                                                                       |
| ----------------------------------- | --------------- | --------- | ----------------------------------------------------------------------------- |
| `format`                            | Integer         | Yes       | The engine: `3`, the form of the file                                         |
| `repo_name`                         | String          | Yes       | `mise.toml`, `gitleaks.toml`, `fnox.toml`                                     |
| `product_name`                      | String          | No        | `fnox.toml`                                                                   |
| `repo_url`                          | String          | No        | `github` tool, `[home].url` of a member                                       |
| `core_branches`                     | List of strings | Yes       | The scripts of `git`                                                          |
| `default_branch`                    | String          | No        | `setup:configs`, `git:pull`, `git:worktree:create`                            |
| `[branches.<branch>].landing`       | String          | No        | `git:merge:<branch>`, the scripts of `git`                                    |
| `[branches.<branch>].merge_from`    | List of strings | No        | `git:merge:<branch>`                                                          |
| `[branches.<branch>].merge_never`   | List of strings | No        | `git:merge:<branch>`                                                          |
| `[commits].style`                   | String          | No        | Nothing yet                                                                   |
| `[commits].scopes`                  | List of strings | No        | `git-conventional-commits.yaml`                                               |
| `environments`                      | List of strings | No        | The sync                                                                      |
| `bootstrap_commit`                  | String          | No        | The sync writes it                                                            |
| `[worktree].files`                  | List of strings | No        | `git:worktree:setup`                                                          |
| `topology`                          | String          | No        | Nothing in bootstrap                                                          |
| `topology_reason`                   | String          | No        | Nothing in bootstrap                                                          |
| `[members.<name>]`                  | Table           | No        | The sync, `setup`, `setup:configs`, `git:worktree:setup`, `git:worktree:list` |
| `[members.<name>].folder`           | String          | No        | The sync, `setup`, `setup:configs`, `git:worktree:setup`, `git:worktree:list` |
| `[members.<name>].type`             | String          | Yes       | The sync, `setup`, `setup:configs`, `git:worktree:list`                       |
| `[members.<name>].url`              | String          | No        | `git:worktree:list`, the sync of a missing sibling                            |
| `[members.<name>].label`            | String          | No        | `setup`, `setup:configs`, `git:worktree:setup`                                |
| `[members.<name>].worktree_install` | String          | No        | `git:worktree:setup`                                                          |
| `[home]`                            | Table           | No        | The sync writes it; `git:worktree:list`                                       |
| `[projects.<name>]`                 | Table           | No        | `detect`, `check`, `describe`; vwf through `bootstrap config`                 |
| `[projects.<name>].folder`          | String          | Yes       | `detect`, `check`                                                             |
| `release_env`                       | String          | No        | Nothing in bootstrap                                                          |
| `[deploy.<env>].<project>`          | String          | No        | Nothing in bootstrap                                                          |

`tools.toml`:

| Key                                 | Type            | Necessary | Read by                                    |
| ----------------------------------- | --------------- | --------- | ------------------------------------------ |
| `format`                            | Integer         | Yes       | The engine: `3`, the form of the file      |
| `[tools.<folder>]`                  | Table           | No        | The sync                                   |
| `[tools.<folder>].env`              | List of strings | No        | The sync                                   |
| `[tools.<folder>].projects`         | List of strings | No        | `check`, `describe`; `detect` writes it    |
| `[tools.<folder>].<key>`            | Any             | No        | The templates and the scripts of the tools |
| `[tools.<folder>.files."<target>"]` | Table           | No        | The sync: the overrides of a shared file   |
| `[tools.claude.plugins.user]`       | Table           | No        | `setup:ai`                                 |
| `[tools.claude.plugins.project]`    | Table           | No        | `setup:ai`                                 |
| `[tools.claude.plugins].installed`  | List of strings | No        | `setup:ai`, which writes it                |
| `[overrides.tools.<folder>]`        | Table           | No        | The sync: the local values of a tool       |
| `[overrides.mise.<path>]`           | Any             | No        | The sync                                   |
| `[overrides.mise_env.<env>.<path>]` | Any             | No        | The sync                                   |

`releases.toml`:

| Key         | Type    | Necessary | Read by                                     |
| ----------- | ------- | --------- | ------------------------------------------- |
| `format`    | Integer | Yes       | The engine: `3`, the form of the file       |
| `<project>` | String  | No        | `bootstrap config`; the release tasks later |

Keys are snake case. A key of a question is a `[tools.<folder>].<key>`:
`bootstrap migrate` and `bootstrap add tool <tool…>` write the answer there. See
[cli.md](cli.md#the-questions).

The gate `code:lint:60-bootstrap` runs `bootstrap check` at each commit that
changes a file of the folder. See [repo.md](repo.md#bootstrap-check).

All the paths in this document are relative to the root of the consumer
repository.

## The schema

The engine checks each key of the folder against a schema, in
`cli/internal/schema`. `check`, the sync, `config` and each other command that
reads the values use it. An unknown key, a value of the wrong type, a value
outside its enum and a missing required key each stop the command with code 2,
and the message names the key and the file.

- `repo.toml` and `releases.toml`: one table of keys in Go, with the key path,
  the type, the enum values, the default and whether the key is required.
- `tools.toml`: a key of `[tools.<folder>]` is valid when the tool is in the
  catalog of the clone and the key is `env`, `files`, `projects`, or one of its
  questions or `[defaults]`. A tool that the clone does not have stops the
  selection with
  code 3. The keys of `[overrides.tools.*]` keep the rules of
  [cli.md](cli.md#the-values-of-a-tool). A key that lives in `repo.toml` now
  (`[tools.git].commit_scopes`, `[tools.git.branch.*]`,
  `[tools.mise].worktree_files`) stops the command in `[tools.*]` and in
  `[overrides.tools.*]`, and the message names its new key.
- The rules between keys of `repo.toml`: each `[branches.<b>]` names a branch
  of `core_branches`; when `main` and `develop` are both core branches and
  `main` has `landing = "direct-commit"`, `develop` has it too; each
  `[members.<name>]` has a `type` and a name of letters, digits, `-` and `_`;
  a `sibling` member has a `url`; `[home]` has `folder` and `type`; each
  `[projects.<name>]` has a lowercase kebab-case name and a `folder` of the
  repository, and no two projects have one folder.
- The rule between the two files: each name of a `[tools.<folder>].projects`
  is a `[projects.<name>]` of `repo.toml`, and a `root_only` tool has no
  `projects`.

`bootstrap config set` and `unset` refuse, with code 2, a change whose result
breaks a rule, and write nothing.

## The keys of `repo.toml`

### `format`

The form of the file. The value is `3`. `bootstrap migrate` writes it in a new
file and in a migrated file. Another value stops the command with code 2.

### `repo_name`

The name of the repository, for example `my-repo`. It is a literal, not the
folder name of a worktree. A new `repo.toml` gets the name of the folder of the
repository.

| Output                  | Use                                                         |
| ----------------------- | ----------------------------------------------------------- |
| `.config/mise.toml`     | `REPO_NAME` in `[env]`                                      |
| `.config/gitleaks.toml` | The `title` of the gitleaks config (a seed)                 |
| `.config/fnox.toml`     | The profile and the keychain `service` or `prefix` (a seed) |

The `setup:vscode` task uses `REPO_NAME` as the name of the VS Code profile.

### `product_name`

The name of the product, for example `my-product`. Set it in each repository of
a product with members, the root included.

With this key, the fnox keychain uses it as `service`, and `<repo_name>/` as
`prefix`. Without this key, `service` is `repo_name` and there is no prefix.

### `repo_url`

The web URL of the repository, for example `https://github.com/me/my-repo`.

| Output                              | Use                                                     |
| ----------------------------------- | ------------------------------------------------------- |
| `.github/ISSUE_TEMPLATE/config.yml` | The links to the README and to the security advisories. |
| `[home].url` of each member         | The sync of a member writes it there.                   |

Without the key, the templates write no links.

### `core_branches`

The core branches of the repository, in order, for example
`["main", "develop"]`. Each one must exist in git, locally or on origin: the
check script of `git` stops the sync when one does not. A new `repo.toml` gets
the `main` and `develop` that exist, in that order. With neither, it gets the
current branch, born or unborn, of any name. With a detached `HEAD` and neither
branch, `migrate` stops with code 2 and the hint "check out a branch, then run
bootstrap migrate again".

- `branch_pair` of `git` is true when the list has both `main` and `develop`.
  `has_main` and `has_develop` of `git` are true when the list has `main`, or
  `develop`: each gives the task `git:merge:<branch>` of that branch. See
  [repo.md](repo.md#the-branch-model-and-the-merge-tasks).
- `protected_branches` of `git` is each core branch whose landing is not
  `direct-commit`. See [repo.md](repo.md#the-protected-branches).

### `default_branch`

The base of each change: the branch that a new branch starts from, and that
the change lands on. It can be different from the default branch of the
remote. It must be in `core_branches`. Without the key, it is `develop` when
`develop` is a core branch, else `main` when `main` is a core branch, else the
first core branch. `helpers.default_branch()` gives it to the tasks.

- `setup:configs` starts `ops/bootstrap-sync` from it, and lands the sync on
  it. See [repo.md](repo.md#the-entry-tasks).
- `git:worktree:create` starts a new branch from it. See
  [repo.md](repo.md#the-worktree-tasks).
- `git:pull` brings it up to date with the core branches.

The schema refuses `default_branch = "main"` when `develop` is a core branch
and `[branches.main]` has neither `merge_from` nor `merge_never`: the defaults
let only `develop` merge into `main`, so a branch from `main` cannot land.

### `[branches.<branch>].landing`

How a branch lands on `<branch>`. `<branch>` must be in `core_branches`.

| Value           | Effect                                                                                                                                                                                     |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `pr`            | A push of the source branch, then a pull request with `gh` or `glab`. Nothing merges locally.                                                                                              |
| `direct-merge`  | The default. A local merge, then a push: a fast-forward for `develop` into `main` when `develop` is a core branch, else `--no-ff`. The `no-commit-to-branch` hook refuses a direct commit. |
| `direct-commit` | The same merge as `direct-merge`, and a direct commit to the branch is allowed.                                                                                                            |

Each repository reads only the `[branches.<branch>]` of its own `repo.toml`,
with `bootstrap config get`. This applies to a repository without submodules,
to the root of a repository with submodules, and to each member. The merge
rules are in [repo.md](repo.md#the-branch-model-and-the-merge-tasks).

### `[branches.<branch>].merge_from`

The only branches that can merge into the branch. An empty list gives no limit.

- A table with both `merge_from` and `merge_never` stops the merge, also when a
  list is empty.
- The defaults without both keys are in
  [repo.md](repo.md#the-branch-model-and-the-merge-tasks).

### `[branches.<branch>].merge_never`

The branches that must not merge into the branch. An empty list gives no limit.
The rule for both keys and the defaults are in
`[branches.<branch>].merge_from`.

### `[commits].style`

The commit model. The one value is `conventional`, the default. Nothing reads
it yet.

### `[commits].scopes`

The commit scopes of `.config/git-conventional-commits.yaml`, for example
`["backend", "web"]`.

- Without the key, or with an empty list, the scopes are the names of the
  `[members.*]` tables, sorted.
- Without members, the list of scopes is empty.

### `environments`

The mise environments of the repository, from `dev` and `ci`

- `dev` is necessary. `ci` is optional
- Without the key, the sync uses `dev` only and prints a message.
- An unknown name, or a list without `dev`, stops the sync.
- The sync writes `conf.d/<folder>/mise.<env>.toml` only for these
  environments. See [mise.md](mise.md).

### `bootstrap_commit`

The commit of the last sync without a failure that changed a file. The sync
writes it. An example has an empty value. See [cli.md](cli.md#bootstrap_commit).

### `[worktree].files`

The gitignored files that `git:worktree:setup` copies from the main worktree
into a new worktree, for example `[".config/mise.local.toml"]`.

- The task copies a file only in a linked worktree.
- It does not replace a file that exists.
- It prints a warning for a file that is not in the main worktree.
- Without the key, the task has no copy step.

### `topology` and `topology_reason`

The shape of the product: `repo`, `monorepo` or `multi-repo`, and the reason
for it in words. Bootstrap does not read them. A plugin reads and writes them
with `bootstrap config`.

### `[members.<name>]`

One table for each member of the repository. The member list of the sync is
the git submodules, then each `[members.<name>]` table that is not a
submodule: the folder members and the sibling members.

- A submodule without a table is a member of the sync. Its folder is its path.
  Its name is the `repo_name` of its own `.config/bootstrap/repo.toml`. Without
  that value, its name is its path with each `/` changed to `_`, for example
  `backend_lib`. A table whose `folder` is a submodule path gives that
  submodule its name. A submodule without a table gets no flag of a task.
- A member name is a flag and a part of a shell variable: it has only letters,
  digits, `-` and `_`. `check` and `bootstrap config set` refuse another name
  with code 2. See [repo.md](repo.md#the-template-context).
- A `folder` member must have its folder in the repository. If not, the render
  stops. A `sibling` member whose folder is not there is left out of the sync,
  with an info line that names its `url`.
- No member is the repository itself: a `folder` of `.` stops the render. The
  merge config of the repository is in `[branches.<branch>]`.
- The templates get these values. See
  [repo.md](repo.md#the-template-context). `MEMBERS` in
  `.config/mise.toml` is the list of the folders, with one space
  between them. No task reads it.
- The tasks read the members with `helpers.members()`: each `submodule` and
  `sibling` member that has a table, through `bootstrap config get
  repo.members`.
- `setup` and `setup:configs` get a `--<name>` flag and an `--all` flag only
  for the `submodule` and `sibling` members that have a table. They run in the
  folder of the member, through `mise run --cd`. A folder member of a monorepo
  is part of the tree of the root, so it gets no flag. `setup:configs` skips a
  member that is not on the machine, or that has no `.config/bootstrap/`, with
  a warning. `code:format` and `code:lint` have no member flags: each
  repository and each submodule runs its own. `git:worktree:setup` runs
  `worktree_install` for each member.

### `[members.<name>].folder`

The folder of the member, relative to the root of the repository, for example
`projects/web`. The folder of a sibling starts with `../`. The default is
`<name>`. A polyrepo has no members, so it has no `[members.<name>]` table.

### `[members.<name>].type`

`submodule`, `folder` or `sibling`. Required. The render stops when the type is
not correct: a `submodule` that is not in `.gitmodules`, or a `folder` or
`sibling` that is a submodule.

| Type        | The member                                                                 |
| ----------- | -------------------------------------------------------------------------- |
| `submodule` | A git submodule of the repository, with its own `.config/bootstrap/`.      |
| `folder`    | A folder of a monorepo, part of the tree of the root.                      |
| `sibling`   | A separate repository next to this one, with its own `.config/bootstrap/`. |

### `[members.<name>].url`

The URL of the repository of the member. Required for a `sibling` member.
`git:worktree:list` shows it, and the sync names it when the folder of a sibling
is not there.

### `[members.<name>].label`

The display name of the member in the headers of `setup`, `setup:configs` and
`git:worktree:setup`. The default is `<name>`.

### `[members.<name>].worktree_install`

The install command of the member for a new worktree, for example
`pnpm install`. `git:worktree:setup` runs it as
`mise x --cd <folder> -- bash -c <command>`, so the whole command line runs in
the folder of the member, also the part after `&&`. Without the key,
`git:worktree:setup` runs the own `git:worktree:setup` task of the member, if
the member has one.

### `[home]`

The back-link of a `submodule` or `sibling` member to its home repository. The
sync of the member writes it, when `setup:configs` of the home runs it with
`BOOTSTRAP_HOME` and `BOOTSTRAP_MEMBER`. A folder member gets none.

| Key      | Value                                                          |
| -------- | -------------------------------------------------------------- |
| `folder` | The path from the member back to the home, for example `..`.   |
| `type`   | `submodule` or `sibling`: the type of the member in the home.  |
| `url`    | The `repo_url` of the home. Without it, `[home]` has no `url`. |

A member with `type = "submodule"` leaves out the tools with
`root_only = true`. A sibling member is its own root, so it keeps them. A
repository without `[home].type` uses the git test: a git submodule leaves them
out. See [repo.md](repo.md#root_only). `git:worktree:list` uses `[home].folder`
to find the home of the product.

### `[projects.<name>]`

A project of the repository: a folder that builds one thing, for example a
package of a workspace or a Go module. Each repository lists its own projects.
vwf keeps the other facts of a project in its own files, by the same name.

```toml
[projects.web]
folder = "apps/web"

[projects.api]
folder = "services/api"
```

- The name is lowercase kebab case, for example `my-app`.
- `folder` is the path of the folder from the root of the repository, without
  `..`. `.` is the root. Two projects cannot have one folder.
- `bootstrap detect` proposes the tables, and `detect --add` writes them. vwf
  writes a new one with `bootstrap config set repo.projects.<name>`. See
  [cli.md](cli.md#the-projects-of-detect).
- `bootstrap check` stops with code 2 when the folder is not there.
- `bootstrap config unset repo.projects.<name>` stops with code 2 while a tool
  lists the name.

No synced file uses the projects.

### `release_env` and `[deploy.<env>]`

`release_env` names the environment of a release. `[deploy.<env>]` holds the
URL of each project in an environment, `<project> = "<url>"`. Bootstrap does
not read them.

## The keys of `tools.toml`

### `format`

The form of the file. The value is `3`, as in `repo.toml`.

### `[tools.<folder>]`

One table for each selected tool of the repository, also for each required
member of the `base` pack and each dependency, so that a user can write values
in it. Each key is a folder `repo/tools/<folder>/` of `bootstrap`.
`bootstrap add tool <tool…>`, `bootstrap add pack <pack…>` and
`bootstrap remove tool <tool…>` change these tables. The sync writes an empty
table for each required member of `base` that has none, also when the user
deleted it by hand. See [cli.md](cli.md#the-tables-of-the-required-members). An
`env` in the table of a required tool can only add environments, and the sync
examines it. See [packs.md](packs.md#the-base-pack). A table for a dependency
also selects the tool, and its values follow the rule of a base or dependency
tool in `[tools.<folder>].<key>`.

- `env` selects the environments, and each other key is a value of the tool. See
  `[tools.<folder>].<key>`
- The sync adds the `depends` of each tool, depth first
- All dependencies are configured for same environment as parent
- A tool that is installed for `dev` environment will have all it's dependencies
  configured for `dev`
- The `envs` of a dependency must include each environment of its parent. If
  not, the sync stops. See
  [mise.md](mise.md#the-environments-of-a-mise-tool)
- Without `[tools]`, the repository gets only the required members of the
  `base` pack, and the sync writes their tables
- A list, `tools = [...]`, is the old format. It stops the sync, and the
  message names the table format
- A folder that is not in `bootstrap` stops the sync
- A tool with `root_only = true` in a submodule member stops the sync. See
  [repo.md](repo.md#root_only)

### `[tools.<folder>].env`

The environments of the tool, for example `["dev", "ci"]`. Without the key,
the tool gets the `default_envs` of its `tools.toml`. See
[mise.md](mise.md#the-environments-of-a-mise-tool) for the rules of this list.

### `[tools.<folder>].projects`

The projects that the tool serves: names of `[projects.<name>]` tables of
`repo.toml`, for example `projects = ["web"]`. Without the key, the list is
empty. `bootstrap detect` proposes the names, and `detect --add` writes them.
`bootstrap check` warns when the detect script of the tool no longer finds the
tool in the folder of a project. A `root_only` tool has no `projects`. See
[cli.md](cli.md#the-projects-of-detect).

`env` and `projects` are the engine keys of a tool: the values merge does not
merge them, and the `values` of a script of the tool do not have them. No synced
file uses `projects`.

### `[tools.<folder>].<key>`

A value of the tool, for its templates and its scripts: one of its questions
or `[defaults]`.

```toml
[tools.node]
env = [ "dev", "ci" ]
registry = "https://registry.example.com"
```

The sync merges the value over the `[defaults]` of the tool and the
contributions
of the selected tools and the derived values, and a template reads the result
as `.config` or `.all`. See
[cli.md](cli.md#the-values-of-a-tool).

The sync also gives the tools the keys of `repo.toml` that they read in their
tables before the values folder, under their old names: `[commits].scopes` as
`[tools.git].commit_scopes`, `[branches.<b>]` as `[tools.git.branch.<b>]`
(with `merge_model` and `direct_commit` from `landing`), and
`[worktree].files` as `[tools.mise].worktree_files`. A user never writes them
in `tools.toml`.

### `[tools.<folder>.files."<target>"]`

The overrides of the bootstrap items of a shared file of the tool, for example
`[tools.git.files.".gitignore"]` with `"*.log" = { "$mode" = "remove" }`. See
[cli.md](cli.md#overrides-of-a-shared-file). The own items of
the user go directly in the file.

### `[tools.claude.plugins.user]` and `[tools.claude.plugins.project]`

The Claude plugins of each scope. Each key is `"<plugin>@<marketplace>" =
"<source>"`. Both are `[defaults]` of the `claude` tool, empty by default.
Another selected tool can add entries in `[contributes.claude]`, for example
`vwf@virajp-plugins` of `virajp-plugins`: `bootstrap config get` returns them
with `contributed_by`. See [repo.md](repo.md#the-claude-plugins).

### `[tools.claude.plugins].installed`

The record of `setup:ai`: a list of `"<plugin>@<marketplace>"`, the project
plugins that it installed, or found installed while listed. A `[defaults]` of
the `claude` tool, empty by default. `setup:ai` writes it with
`bootstrap config set`, and uninstalls a recorded plugin that leaves
`[tools.claude.plugins.project]`. A user does not write it. See
[repo.md](repo.md#the-claude-plugins).

### `[overrides.tools.<folder>]`

The local values of a tool, which win over its `[tools.<folder>]` table, with
the keys of that table. See [cli.md](cli.md#the-values-of-a-tool). A key of
`repo.toml` has no override: change it in `repo.toml`, with
`bootstrap config set`.

### `[overrides.mise.<path>]` and `[overrides.mise_env.<env>.<path>]`

Values that replace the defaults of the mise config of the tools. See
[mise.md](mise.md#the-overrides).

## The keys of `releases.toml`

### `<project>`

The release level of a project: `NONE`, `PATCH`, `MINOR` or `MAJOR`. The
default is `NONE`. `bootstrap config set releases.<project> <level>` keeps the
higher level when the file already has one, and creates the file with
`format = 3` when it is not there. `bootstrap config unset releases.<project>`
clears it. The release tasks of the plugins still read `.config/releases.yaml`
until they move to `bootstrap config`; `migrate` does not move that file.

## Reserved names

The sync merges these keys into the template context before the values:
`members`, `member_folders`, `member_types`, `member_tables`, `submodules`,
`tools`, `cli_dir`, `config` and `all`. Thus a key of the same name in the
values does not replace the value of the sync. `tools` of the values is the
input of the sync, not the value that the templates get. `member_tables` is the
names of the members that have a `[members.<name>]` table. The context also
has `member`, the old name of `[members.*]`, with the same tables.

## Old forms

`bootstrap migrate` migrates each old form, and `bootstrap sync` runs `migrate`
first (see [cli.md](cli.md#migrate-before-sync)). Each other command stops with
code 2 and the message "run bootstrap migrate", except `version`, `settings` and
each `mac` command. A dry run of `migrate` shows the migration and writes
nothing. A
task of the repository at the place of a test port is an old form too, though
it is not a value. It stops only `sync`, the sync of `add`, `remove` and
`replace`, and `check`: see [The own test tasks](#the-own-test-tasks). The
old worktree command is an old form in the same way: see
[The old worktree command](#the-old-worktree-command).

### The values file `.config/bootstrap.toml`

`.config/bootstrap.toml`, with its overrides file
`.config/bootstrap.overrides.toml` and its manifest
`.config/repo-generated.tsv`, is the old form of the values folder. `migrate`
first migrates each older form of the file (below), then splits it:

- `repo.toml` gets each top-level key of the old file, with the moves of the
  table below, `core_branches` and `format = 3`.
- `tools.toml` gets its `[tools.*]` and `[overrides.*]` tables, without the
  keys that move to `repo.toml`, and `format = 3`.
- The `[tools.*]` tables of the overrides file go to `[overrides.tools.*]` of
  `tools.toml`. A key of the overrides file that moves to `repo.toml` goes
  there directly, and wins over the value of the old file.
- `.config/repo-generated.tsv` moves to `.config/bootstrap/repo-generated.tsv`.
- The old values file and the overrides file go.
- `.config/releases.yaml` stays where it is.

`core_branches` is the `main` and `develop` that exist in git, locally or on
origin, in that order. With neither, `migrate` stops with code 2. A
`[tools.git.branch.<b>]` table of a branch that is not `main` or `develop`, or
that does not exist, stops `migrate` with code 2, and the message names the
table
with the fix "remove the table or create the branch".

The migration is atomic: it writes the folder to a temporary path, renames it
into place, then removes the old files. When the sync after it fails, the old
files come back and the folder goes. The comments of a table that moves to
`repo.toml` are lost; the comments of each other key stay.

With both the old file and the folder, each command, `migrate` included, stops
with code 2 and the fix "remove one: compare .config/bootstrap.toml with
.config/bootstrap/". A leftover `.config/bootstrap.overrides.toml` next to the
folder is not read and stops nothing. A folder without `repo.toml` or
`tools.toml` stops each command but `migrate` with code 2; `migrate` creates the
missing file.

| Old key                                                       | `bootstrap migrate` moves it to                                                                                |
| ------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `[tools.git].commit_scopes`                                   | `[commits].scopes` of `repo.toml`                                                                              |
| `[tools.git.branch.<b>].merge_model = "pr"`                   | `[branches.<b>].landing = "pr"`                                                                                |
| `merge_model = "direct"` or none, with `direct_commit = true` | `[branches.<b>].landing = "direct-commit"`                                                                     |
| Each other `[tools.git.branch.<b>]`                           | `[branches.<b>].landing = "direct-merge"`                                                                      |
| `[tools.git.branch.<b>].merge_from`, `merge_never`            | `[branches.<b>]`, unchanged                                                                                    |
| `[tools.mise].worktree_files`                                 | `[worktree].files` of `repo.toml`                                                                              |
| `[member.<name>]`                                             | `[members.<name>]`; a table without `type` gets `submodule` when its folder is in `.gitmodules`, else `folder` |
| Each other top-level key                                      | `repo.toml`, unchanged                                                                                         |
| `[tools.*]` and `[overrides.*]`                               | `tools.toml`                                                                                                   |
| `.config/bootstrap.overrides.toml`                            | `[overrides.tools.*]` of `tools.toml`                                                                          |
| `.config/repo-generated.tsv`                                  | `.config/bootstrap/repo-generated.tsv`                                                                         |

`[commits].style = "conventional"` is written too. A sibling member and its
`url` are not migrated: add them with `bootstrap config set`.

### Older forms of the values file

A file without `format = 2`, or with a `[packs.<name>]` table, is an older
form. `migrate` expands each `[packs.<name>]` table and the old implicit pack
`base` with a frozen copy of the packs files of before `format = 2` (only the
`default = true` members, and the nested packs). It writes a
`[tools.<name>]` table for each such tool and each new dependency that has no
table, removes the `[packs.*]` tables, and then splits the file as above. A new
table gets the `env` of its old pack. A tool that already has a table keeps its
own `env`. When two old packs give one tool different `env` lists, the table
gets both. When one source has no `env`, the table has no `env`. A tool of the
frozen copy that the clone no longer has gets no table. A `[tools.<name>]`
table of a tool that the clone no longer has (for example `devtools` on the
Mac) is removed, and the change list names it. A file with `format = 2` and a
`[packs.*]` table gets only the tables of its packs, not the members of
`base` again. A `[packs.<name>]` table of a pack that the frozen copy does not
have stops `migrate` with code 3 and the message
`<pack> is not a bootstrap pack (named in <file>)`, as the old sync did. The
user fixes the values file by hand, then runs `migrate` again.

| Old form                        | `bootstrap migrate` moves it to                                                                                                                                                                                      |
| ------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| No `format = 2`                 | A table for each tool of the old pack `base`                                                                                                                                                                         |
| `[packs.<name>]`                | A `[tools.<name>]` table for each tool of the old pack, with the `env` of the pack                                                                                                                                   |
| `commit_scopes`                 | `[commits].scopes`                                                                                                                                                                                                   |
| `[branch.<branch>]`             | `[branches.<branch>]`, with `landing`                                                                                                                                                                                |
| `worktree_files`                | `[worktree].files`                                                                                                                                                                                                   |
| `[claude.plugins.*]`            | `[tools.claude.plugins.*]`                                                                                                                                                                                           |
| `extensions`                    | `.vscode/extensions.json`, as own items of the shared file                                                                                                                                                           |
| `[overrides.gitignore]`         | `add`: lines of `.gitignore`. `remove`: `[tools.git.files.".gitignore"]` with `"$mode" = "remove"`                                                                                                                   |
| The table of the old entry tool | Removed: the `mise` tool has the entry now                                                                                                                                                                           |
| The old file of the own hooks   | Its hooks move to `.config/pre-commit-config.yaml` as own blocks, and the file goes. The file holds its blocks in `repos`, or is a list of the blocks. A hook whose `id` is already in the config is not added again |

`docs/sync/migrate.md` names the old entry tool and its old files. These old
items are not migrated:

| Old item                                   | Effect                                                                                                                     |
| ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- |
| The old file of the own `.gitignore` lines | Stops the sync: move its lines into `.gitignore`, then delete it.                                                          |
| `[overrides.graphifyignore]`               | Stops the sync: remove the table.                                                                                          |
| `core_tools`                               | Not used. The sync prints a message: delete it. The `env` of each entry, else `default_envs`, decides the environments.    |
| `.config/repo-config.toml`                 | The oldest name of the values file. Nothing reads it: rename it to `.config/bootstrap.toml`, then run `bootstrap migrate`. |

### The own test tasks

The `mise` tool ships the dispatchers `test`, `test:e2e` and `test:golden`. A
task of the repository at one of their paths is an old form. It is a regular
file that the manifest does not list, without the bootstrap header.

| Old form, below `.config/mise/tasks/` | `bootstrap migrate` moves it to |
| ------------------------------------- | ------------------------------- |
| `test`, `test/_default`               | `test/50-local`                 |
| `test/e2e`, `test/e2e/_default`       | `test/e2e/50-local`             |
| `test/golden`, `test/golden/_default` | `test/golden/50-local`          |

`check` reports it in the stage `old-forms`, and `sync` stops on it, both with
code 2. When the target exists, `migrate` stops with code 2 and moves nothing.
See [cli.md](cli.md#migrate).

A known old copy of a part of a selected tool is not moved: `migrate` removes
it first. It is a file at the old path or at `<group>/50-local` whose sha256
is in the `replaces` of the part, for example the old `test/golden` task of
the compose pack of stackgen. See
[tools.md](tools.md#the-tasks-table). At the old path, `check` reports it in
the stage `old-forms` as above. At `<group>/50-local`, `check` reports it in
the stage `scripts`, and `sync` stops on it, both with code 2 and "run
bootstrap migrate".

### The old worktree command

`mise run setup:worktree` runs the old name of `git:worktree:setup`, and no
task has that name now. It is an old form in the seed `.config/wt.toml` and in
a `[members.<name>].worktree_install` of `repo.toml`. `bootstrap migrate` writes
`mise run git:worktree:setup` in its place, and the rest of the file or value
stays. `check` reports it in the stage `old-forms`, with code 2 and "run
bootstrap
migrate". `bootstrap sync` runs `migrate` first. See [cli.md](cli.md#migrate).

## The location of the folder

The default path is `.config/bootstrap/`. `BOOTSTRAP_VALUES` names a different
folder; its files keep their names inside it. Without the folder, each command
except `bootstrap migrate`, `version`, `settings` and the `mac` commands stops
with code 2 and the fix "run bootstrap migrate". During `migrate` and each sync,
the
engine points `BOOTSTRAP_VALUES` at the temporary folder of the planned values,
so a script that runs `bootstrap config get` reads the new keys, also during a
migration. See [cli.md](cli.md#the-engine-variables).
