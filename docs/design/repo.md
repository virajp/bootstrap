# The repo side

This document describes the repo side of the sync: `bootstrap sync` and the
front-end `internal/repo` of the engine. The front-end writes the files of the
selected tools of `repo/tools/` into a consumer repository. This document holds
the rules that only a repository has: the entry tasks, members and submodules,
the template context, the branch model, the pre-commit config, `check` and its
gate.

See also: [cli.md](cli.md) for the command line, the shared
rules and the exit codes, [bootstrap-config.md](bootstrap-config.md) for the
keys of the values folder `.config/bootstrap/`, [packs.md](packs.md),
[tools.md](tools.md),
[mise.md](mise.md) and [fnox.md](fnox.md).

## The entry tasks

The `mise` tool of `repo/tools/` ships the entry of the sync, so each
repository has it. The tool also has the mise tool `github:virajp/bootstrap`
in `[mise_env.dev.tools]`.

| Task                     | Job                                                                             |
| ------------------------ | ------------------------------------------------------------------------------- |
| `setup:configs`          | Syncs and lands the sync in the repository, and in each selected member.        |
| `code:lint:60-bootstrap` | Hidden. The gate of the values. See [The gate](#the-gate-codelint60-bootstrap). |

`setup:configs` (`setup/configs/_default.py`) has
`github:virajp/bootstrap` in its one `# MISE tools={...}` line, so it runs
also before the first sync. With `BOOTSTRAP_BIN` set, it runs that file in place
of the `bootstrap` of mise. Its order is:

1. Without `--dry-run`, it stops with exit 1 when the root, or a selected
   member with `.config/bootstrap/`, has uncommitted or untracked files. It
   examines each repository before it changes one.
2. It prints the header `Syncing the shared files from bootstrap ...`. Then it
   fetches `origin` and finds the default branch (`helpers.default_branch()`,
   see [`default_branch`](bootstrap-config.md#default_branch)). It stops when
   the default branch has diverged from `origin`.
3. When the repository has the task `git:merge:<default branch>`, it stops if
   the branch `ops/bootstrap-sync` exists, locally or on `origin`. Else it
   creates `ops/bootstrap-sync` from the default branch, or from
   `origin/<default branch>` when origin is ahead, and runs `bootstrap sync`
   there, with `MISE_PROJECT_ROOT` set to the root. When the sync changed
   files, it commits them as `ops: sync bootstrap <commit>` and runs
   `mise run --cd <folder> git:merge:<default branch> ops/bootstrap-sync`:
   a merge and a push, or a pull request under `landing = "pr"`. Then it goes
   back to the branch where it started and deletes the local
   `ops/bootstrap-sync`: under `pr`, origin keeps the branch for the request,
   and else the default branch must have it. A sync with no change makes no
   commit and deletes the branch.
4. Without that task, for example at the first sync, it checks out the default
   branch, runs `bootstrap sync` and commits the changes there, then goes back.
5. With `--dry-run`, it only runs `bootstrap sync --dry-run`: no branch and no
   commit.
6. With a member flag, it reads the members with `helpers.members()`
   (`bootstrap config get repo.members`). For each selected member, it prints
   the header `Syncing <label> ...`. Then it does steps 2 to 5 in the folder of
   the member, with `MISE_PROJECT_ROOT` set to that folder, `BOOTSTRAP_HOME` set
   to
   the root and `BOOTSTRAP_MEMBER` set to the name of the member, so the sync
   of the member writes its `[home]`.
7. After the members, it looks for submodule members whose landing moved
   their checkout, so the root has a new pointer for them. It commits only
   those pointers as `ops: update the member pointers`, with the flow of steps
   2 to 4 in the root, and lands them. A sibling member is outside the root,
   so it has no pointer.

A commit of the task runs the hooks of the repository. When a hook changes a
file, for example a formatter, git makes no commit: the task adds the files
and commits one more time, and stops only when that commit also fails.

The first failure stops the task, and the exit code of the engine is the exit
code of the task. `bootstrap sync` takes the clone, examines its version and
syncs: see [cli.md](cli.md#the-cache-and-the-clone).

| Flag                | Effect                                                                                                                                             |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--dry-run`         | Shows the changes. Writes nothing, and makes no branch and no commit.                                                                              |
| `--verbose`         | Sets `BOOTSTRAP_VERBOSE=1`. The task does not pass the flag to the engine.                                                                         |
| `--all`, `--<name>` | Also syncs each submodule and sibling member, or the named one. Only a repository with such a member that has a `[members.<name>]` table has them. |

A submodule or sibling member is a separate repository with its own
`.config/bootstrap/`. The task skips, with a warning, a member that is not on
the machine (`member <name>: <folder> is not on this machine`) and a member
without `.config/bootstrap/`, or with only the old form
(`member <name>: run bootstrap migrate there`), and goes on.

By default, the sync starts with one line:
`↻ bootstrap <short commit> · <n> tools · env <environments>`.

`bootstrap sync` makes its own private copy of the clone, because it writes
the generated mise config into the copy. On macOS, the copy is `cp -cR`, which
makes copy-on-write clones on APFS. When `cp -cR` fails, the sync empties the
copy and uses `cp -R`, as on Linux.

## The values folder and the selection

`bootstrap sync` reads the values folder `.config/bootstrap/` of the consumer:
`repo.toml` and `tools.toml`, with the overrides in `[overrides.tools.*]` of
`tools.toml`. The keys and their schema are in
[bootstrap-config.md](bootstrap-config.md).

- It adds the required members of the pack `base`. The name `base` is fixed in
  the engine. See [packs.md](packs.md). It writes an empty table for each
  required member that has none. See
  [cli.md](cli.md#the-tables-of-the-required-members).
- It adds the tool of each `[tools.<name>]` table, adds the `depends` of each
  tool and the tool of each mise `depends`, depth first, and copies each file of
  each tool. A dependency comes before the tool that needs it.
- The command `bootstrap sync` runs `migrate` first when it has a change to
  make (see [cli.md](cli.md#migrate-before-sync)), so the stops of an old form
  below show only when the front-end runs alone, for example under `check`.
- The values folder is necessary. A missing folder, `repo.toml` or
  `tools.toml` stops the sync with code 2 and "run bootstrap migrate". An
  unknown
  key stops it with code 2.
- It examines the `tools.toml` of each tool of the clone, selected or not. See
  [tools.md](tools.md).
- An old form of the values stops the sync with "run bootstrap migrate". See
  [bootstrap-config.md](bootstrap-config.md#old-forms).
- A task of the repository at the place of a test port (for example its own
  `.config/mise/tasks/test/_default`) stops the sync with code 2 and "run
  bootstrap migrate", before it writes a file. `migrate` moves it to a
  `50-local`
  part. See [cli.md](cli.md#migrate).
- A known old copy of a part of a selected tool at `<group>/50-local` (for
  example the old `test/golden` task of a stackgen pack) stops the sync with
  code 2 and "run bootstrap migrate". `migrate` removes such a copy, there and
  at
  the old path of the group, and does not move it. See
  [cli.md](cli.md#migrate).

## Members and submodules

A member is a git submodule, a folder below the root, or a sibling repository
next to it. The keys are in
[`[members.<name>]`](bootstrap-config.md#membersname). No bootstrap file reads
a vwf file for the members.

- The member flags (`--all`, `--<name>`) of `setup` and `setup:configs` are
  only for the submodule and sibling members that have a `[members.<name>]`
  table. A folder member is part of the root tree. `code:format` and
  `code:lint` have no member flags: their `--all` means every file of the
  current repository, not the members. Each repository and each submodule runs
  its own, and a task runs in a member only through `mise run --cd`.
- A task that runs in a member uses its folder.
- A submodule or sibling member is a separate repository, with its own values
  folder and its own manifest. The sync never writes a file of a member, except
  the `[home]` of its own `repo.toml` when `setup:configs` of the home runs its
  sync. See [`[home]`](bootstrap-config.md#home).
- A sibling member whose folder is not there is left out, with an info line.
- The members are in the input of each script: `repo.members`. See
  [cli.md](cli.md#the-scripts-of-a-tool). The derive script of
  `mise` makes the aliases `setup-all` and `setup-<member>` for the submodule
  and sibling members that have a `[members.<name>]` table: it reads the
  tables with `bootstrap config get repo.members`.

### `root_only`

A `root_only` tool is not for a submodule member: a repository whose `[home]`
has `type = "submodule"`, or, without `[home].type`, a git submodule. In such a
member, `migrate`, `add pack` and the migration of an old form skip it, and the
sync stops when a `[tools.*]` entry names it. A sibling member is its own root,
so it keeps them. A `root_only` tool has no `projects` list, and `detect` gives
it no link to a project.

## The template context

The context of a template is the values: the keys of `repo.toml`, then the
`[tools.*]` and `[overrides.*]` tables of `tools.toml` without
`[overrides.tools.*]`, and `member`, the old name of `[members.*]`. These
values are merged over it. A key of the same name in the values does not
replace them. The reserved
names are in [bootstrap-config.md](bootstrap-config.md#reserved-names).

| Value             | Content                                                                                                                                                                                                                                                                                                                     |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.members`        | The list of member names: the git submodules (the key of the `[members.<name>]` table whose `folder` is the submodule path, else the `repo_name` of the `repo.toml` of the submodule, else the path with each `/` changed to `_`), then each other `[members.<name>]` key. A sibling whose folder is not there is left out. |
| `.member_folders` | A map from each name to its folder (`folder`, default the name).                                                                                                                                                                                                                                                            |
| `.member_types`   | A map from each name to `submodule`, `folder` or `sibling`.                                                                                                                                                                                                                                                                 |
| `.member_tables`  | The list of the members that have a `[members.<name>]` table. A git submodule without one is a member, but gets no flag of a task.                                                                                                                                                                                          |
| `.submodules`     | The list of the git submodules only, nested ones included. It is empty in a repository without submodules.                                                                                                                                                                                                                  |
| `.tools`          | The list of the selected tools, with the required members of `base` and the dependencies added.                                                                                                                                                                                                                             |
| `.cli_dir`        | The absolute path of the `cli/` folder of the clone copy of the run.                                                                                                                                                                                                                                                        |
| `.config`, `.all` | The values of the tools. See [cli.md](cli.md#templates).                                                                                                                                                                                                                                                                    |

A name has only letters, digits, `-` and `_`, and two members never have the
same name. If not, the sync stops.

## The branch model and the merge tasks

The `git` tool holds the branch model. The engine has no branch rule.

- The derive script of `git` gives `branch_pair` (true when `core_branches` of
  `repo.toml` has both `main` and `develop`), `has_main` and `has_develop`
  (true when `core_branches` has `main`, or `develop`) and `protected_branches`
  (each core branch whose `landing` is not `direct-commit`). It reads
  `core_branches` and `[branches.*]` with `bootstrap config get`
  (`BOOTSTRAP_BIN`, else the `bootstrap` of mise). `repo.toml` has no
  overrides.
- The task `git:merge:<branch>` lands a branch on `<branch>`: the name is
  always the destination. `git` ships `git:merge:main` with
  `when = "git.has_main"` and `git:merge:develop` with
  `when = "git.has_develop"`. Thus the sync writes the merge task of each core
  branch, and the stale rule removes the others. No other task merges, and the
  old names have no alias.
- The check script of `git` stops the sync when a branch of `core_branches`
  does not exist, locally or on `origin`, and is not the unborn branch that
  HEAD names. A detached HEAD names no branch. Without an engine that answers
  `bootstrap config get`, for example one older than 4.0.0, it stops the sync
  with code 2.

| Branches             | Tasks                                                                                                                                                                                                           |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `main` and `develop` | Work starts on feature branches. `git:merge:develop [<branch>]` lands a branch on `develop`; the default is the current branch. `git:merge:main [<branch>]` lands a branch on `main`; the default is `develop`. |
| Only one of them     | `git:merge:<that branch> [<branch>]` lands any feature branch on that branch; the default is the current branch.                                                                                                |

- When `core_branches` has `main` and `develop`, a direct landing of
  `develop` on `main` is a fast-forward
  (`git merge --ff-only`), so `main` is always an ancestor of `develop`. Before
  the hook pass and before a branch changes, the task fetches `main` from
  `origin` and makes sure that `main` and `origin/main` are ancestors of
  `develop`. If not, it stops with exit 1, lists the commits that `develop`
  does not have, and gives the fix: merge `main` into `develop`, push
  `develop`, then land again. It never falls back to `--no-ff`. Each other
  direct landing uses `--no-ff`, so the merge commit records which commits
  of a branch landed. `pr` does not change: the forge merges.
- A hook pass over all files runs first as a safety net,
  with `SKIP=no-commit-to-branch`: that hook guards a commit, and the safety net
  makes none. Thus `git:merge:main` also runs from a protected `develop`.
- The merge tasks read `[branches.<destination>]` of the repository's own
  `.config/bootstrap/repo.toml` when they run, with `helpers.bootstrap_get`
  (`bootstrap config get`): the root and each member read their own
  values. A read failure stops the merge with code 1. Without an engine that
  answers, for example one older than 4.0.0, the merge stops with code 2.
- A merge task stops before it touches a branch when `MISE_ORIGINAL_CWD` is
  inside the tree of the git top level of `MISE_PROJECT_ROOT`, but has another
  git top level. In a submodule without that task, mise runs the root's task in
  the root folder, and the task would land a branch of the root.
  `MISE_ORIGINAL_CWD` is the folder of the shell, also with
  `mise run --cd <folder>` and `mise -C <folder>`, and no variable gives
  `<folder>`. Thus a run from a folder outside the tree passes, for example
  the merge of a member that `setup:configs --all` runs from the root, and the
  stop cannot find a `--cd` into a submodule without the task.
- `landing` selects `direct-merge` (local merge and push, the default),
  `direct-commit` (the same merge) or `pr` (push and open a pull request with
  `gh` or `glab`). The schema allows no other value. The header names the
  landing: `Landing <source> on <destination> (<landing>) ...`. Without an
  `origin` remote, a direct landing merges locally and skips the pull and the
  push, and `pr` stops before the hook pass, with the fix
  `bootstrap config set 'repo.branches."<branch>".landing' direct-merge`. The
  `github` and `gitlab` tools install these CLIs.
- `merge_from` lists the only source branches, and `merge_never` the forbidden
  ones. A table with both stops the merge. Without both, the defaults are: into
  `main` with `develop`, only from `develop`; into another branch, never from
  `main`.

### The protected branches

`landing = "direct-commit"` in `[branches.main]` or `[branches.develop]` lets a
direct commit to that branch. The sync reads it through the scripts of `git`.
The merge tasks merge the same way for `direct-merge` and `direct-commit`.

The `no-commit-to-branch` hook of the `git` tool protects only the core branches
in `protected_branches`. The pre-commit spec gives the hook one
`"--branch", "<branch>"` pair for each of them, and removes the hook when no
branch is left, because without `--branch` the hook protects `main` and
`master`.

| `core_branches`      | `landing = "direct-commit"` in | Hook args                         |
| -------------------- | ------------------------------ | --------------------------------- |
| `main` only          | `[branches.main]`              | No hook                           |
| `develop` only       | `[branches.develop]`           | No hook                           |
| `main` and `develop` | `[branches.develop]` only      | `--branch main`                   |
| `main` and `develop` | Both tables                    | No hook                           |
| `main` and `develop` | `[branches.main]` only         | The command stops                 |
| `main` and `develop` | No table                       | `--branch main --branch develop`  |
| One branch           | No table                       | `--branch <branch>` of the branch |

The schema of `repo.toml` examines the tables before the sync writes a file,
and each command that reads the values stops with code 2 on these:

- A `[branches.<b>]` table for a branch that is not in `core_branches`.
- A `landing` that is not `pr`, `direct-merge` or `direct-commit`.
- With `main` and `develop`, `landing = "direct-commit"` for `main` but not for
  `develop`. A repository with both branches never lets a direct commit to
  `main` and refuses it on `develop`.

The check script of `git` stops the sync when a branch of `core_branches` does
not exist, locally or on `origin`, and is not the unborn branch that HEAD
names: create the branch, or remove it from
`core_branches`.

### The task `git:pull`

The `git` tool ships `git:pull`. It fetches `origin` (with `--prune` and
`--tags`) in the repository and in each submodule and sibling member of
`helpers.members()` that is on the machine. Then, for each branch of the set,
it creates a local branch that tracks `origin/<branch>` when none exists, and
fast-forwards the local branch when origin is ahead. A branch that a worktree
has checked out moves with `git merge --ff-only` in that worktree, so its files
move too; each other branch moves with `git update-ref`. A local branch that is
ahead of origin, or that has diverged, stays as it is: a diverged branch gets a
warning. The task never merges, rebases or deletes a branch.

The set is `core_branches` and the default branch of the repository, or
`main` and `develop` in a member without `.config/bootstrap/`. With `--all`,
it is each branch of origin. A failure of a fetch or a fast-forward gives exit
1 after the task has done each other repository.

### The worktree tasks

The `git` tool ships `git:worktree:create` and `git:worktree:delete`, and the
`mise` tool ships `git:worktree:setup` and `git:worktree:list`. They use raw
git, not worktrunk. `create` and `delete` read `core_branches` with
`helpers.bootstrap_get`, as the merge tasks do.

- `git:worktree:create <branch> [--from <base> | --current] [--no-setup]`
  adds the worktree
  in `.worktrees/<branch>` of the main checkout. It uses `worktrees/` only when
  `.worktrees/` is absent and `worktrees/` exists. A local branch is checked
  out. A branch only on `origin` becomes a local branch that tracks it, and the
  task prints a warning first: it trusts and runs the mise config of that
  branch as is. Any other branch is new, cut from `--from`, from the current
  branch with `--current`, else from the default branch
  (`helpers.default_branch()`, see
  [`default_branch`](bootstrap-config.md#default_branch)): the local branch
  when it exists, else `origin/<branch>`, with `--no-track`. Then it runs `mise
  trust` and `git:worktree:setup` in the new
  worktree. `--no-setup` skips only the setup.
- `create` stops with code 2 for a core branch, for `--from` and `--current`
  together, and for `--from` or `--current` with a branch that exists. It stops
  with code 1 when the path exists, or when git refuses
  the add, for example for a branch that another worktree has checked out.
  When the trust or the setup fails, the worktree stays: the task prints the
  command that runs the setup again and the `delete` command, and exits 1.
- `git:worktree:delete <branch>` finds the worktree of the branch in
  `git worktree list`, at any path, and runs `git worktree remove` on it. git
  refuses a worktree with changes (code 1). Then the task deletes the branch
  with `git branch -D` only when a core branch has each of its commits: it
  tests each core branch, the local ref, else `origin/<core>`, else none. With
  no such core branch, the branch stays, and the task prints the `-D` command
  and exits 1.
- `delete` stops with code 2 for a core branch, and when it runs inside the
  worktree that it would remove. It stops with code 1 when no worktree has the
  branch.

## The pre-commit config

`.config/pre-commit-config.yaml` is a shared file of the `pre-commit` tool. Its
spec is `repo/tools/pre-commit/files/.config/pre-commit-config.yaml.merge.toml`.
See [cli.md](cli.md#shared-files).

- The hook repositories come from `[[defaults.repos]]` of `pre-commit` and from
  `[[contributes.pre-commit.repos]]` of each selected tool. Each repository has
  an `id`, and the spec writes it as one block, found by `repo` and `hooks.id`.
- The `rev` of a found block stays (`ignore = ["rev"]`). `setup:precommit` runs
  `pre-commit autoupdate`, so each `rev` becomes the newest tag. The sync never
  calls `pre-commit`.
- The top-level `exclude` comes from `.config.exclude`, the list of the values
  of the tool: `[defaults]`, then `[contributes.pre-commit]` of each selected
  tool, then `[tools.pre-commit]` of `tools.toml`. pre-commit reads one
  regular expression, so the spec joins the items with `|`.
- The own hooks of a repository are blocks that the user writes directly in
  the file. The sync never touches them.
- An override removes or skips a bootstrap block by its `id`, for example
  `[tools.pre-commit.files.".config/pre-commit-config.yaml".repos]` with
  `"conventional-commits" = { "$mode" = "remove" }`.

## Contributed items

A tool can add items to a merge-mode file of another tool, the owner, with
`[[contributes_items.<owner>."<target>".item]]` in its `tools.toml`. The owner
stays the one owner of the file. For example, `virajp-plugins` adds the vwf
lines to `.gitignore` of `git`. The rules of the key are in
[tools.md](tools.md#the-defaults-questions-and-contributes-tables).

- The sync adds the contributed items of the selected tools after the items of
  the spec of the owner, in the order of the names of the contributing tools,
  each tool's items in the order of its file. An item whose identity is
  already there is not added again.
- The override table of the owner's file applies to the contributed items too,
  for example `[tools.git.files.".gitignore"]` with
  `"docs/scratchpad/" = { "$mode" = "remove" }`. An override of an item that a
  tool of the catalog contributes is known, also when that tool is not
  selected: it is not an unused override.
- A contribution to an owner that is not selected adds nothing.

## fnox profiles

A repository does not use the fnox `default` profile, which is only for the
user-level fnox settings. The check script of the `fnox` tool examines
`.config/fnox.toml` when `fnox` is selected. It stops the sync on a top-level
`[secrets]` table (the `default` profile, also when empty), a
`[profiles.default]` table, or no `[profiles.<name>]` table. A repository with
one profile uses its `repo_name`. Each `fnox` call passes `-P <name>`. The rest
of the design is in [fnox.md](fnox.md).

## `requires_found`

The `[files]` and `[tasks]` key `requires_found` and the same key of a
`[mise.tools]` or a `[mise_env.<env>.tools]` entry hold a list of file names.
The list is met when the sync finds any one of the names at the root or down to
4 folders below it, in a file that git tracks or does not ignore. The sync does
not look in git submodules, because a submodule has its own sync. A name has no
`/`. A target with `requires` and `requires_found` needs each `requires` path
and one name of `requires_found`. The line of a skipped target names the first
missing `requires` path, else each name of `requires_found`: `(no a)`,
`(no a or b)`, `(no a, b or c)`.

## `bootstrap check`

`bootstrap check` examines the values folder. It writes no file and does not set
`bootstrap_commit`. It runs each check and prints each message, so each error
shows at once.

1. The local rules: the old forms (also a task of the repository at the place
   of a test port), the schema of `repo.toml`, `tools.toml` and
   `releases.toml` (an unknown key is code 2), `environments`, `env`, the
   structure of `[overrides]`, the members and the folder of each
   `[projects.<name>]`. A task without an `NN-` name in the folder of a
   dispatcher is a warning.
2. With a clone, the rules of the clone: the selection, the values of the tools
   (with the derive scripts), the questions without an answer, the check
   script of each selected tool, and the `projects` list of each tool: its
   detect script runs in the folder of each of its projects, and a link that no
   longer matches is a warning. See
   [cli.md](cli.md#the-json-contract) ("The stages of `check`").

The command takes a clone as each command does, so it examines both parts. The
front-end skips part 2, with a line, when it has no clone.

### The gate `code:lint:60-bootstrap`

The gate runs `bootstrap check`. The `lint` hook runs `code:lint`, which runs
the gate with the staged files, so the tool adds no hook. The gate is hidden,
and it ignores `--fix`.

It prints nothing and exits 0 in these cases:

- Without `bootstrap` (`mise x -- which bootstrap`), and without
  `BOOTSTRAP_BIN`.
- The list of `gate_files` is empty. Its `# FILES` header has
  `include=".config/bootstrap/*.toml"` and
  `exclude=".config/bootstrap/config-files.toml"`, so the gate runs when the
  given files, the uncommitted files or, with `--all`, the files of the
  repository have a values file. The rendered `config-files.toml` does not
  count.
- The cache has no clone with `cli/VERSION`.
- The raw `origin` of `<cache>/repo` is not one of five literal forms of the
  default URL. Thus the gate never uses code of another repository at commit
  time.

Else it runs `bootstrap check` with `BOOTSTRAP_CLONE` set to the cache clone.
Thus the gate takes no lock and never fetches. A binary that is older than
`VERSION` of the cache clone stops the gate with code 2.

## The manifest

The manifest is `.config/bootstrap/repo-generated.tsv`, in the values folder
(`$BOOTSTRAP_VALUES/repo-generated.tsv`). It is committed, and each git
submodule has its own. The plan of a migration copies the old
`.config/repo-generated.tsv` into its temporary values folder, so a dry run of
`migrate` lists the same removals as `migrate`. Its shape and the stale rule are
in
[cli.md](cli.md#the-manifest-and-the-stale-rule).

After a clean run, the order is the stale rule, the write of the manifest, then
`bootstrap_commit`. A run with a failure runs no stale rule and writes no
manifest. A manifest that the sync cannot read fails the run: it logs
`FAILED: .config/bootstrap/repo-generated.tsv`, and it skips the stale rule, the
write of
the manifest and `bootstrap_commit`.

Handled paths are the files of the selected tools and their dependencies, also
through a mise `depends`, that this run wrote or kept (owned, seed or shared).
A seed or a shared file of such a tool that the run left in place counts as
kept, also one that `requires` or its environment skips. An owned file that
`requires`, its environment or a false `when` skips is not handled. A folder
member without its own `.git` is part of the root tree.

### What the sync deletes

The sync never deletes a file, except these:

- A synced file at a path where the sync needs a folder.
- The stale rule: a file with the bootstrap header that the manifest lists and
  this run does not handle, for example a merge task of the other branch model.
- An item that bootstrap added to a shared file, when no spec and no
  contributed item of a selected tool has it now.

The stale rule skips a seed and a shared file of a tool of this run. At a folder
path, the sync refuses to remove such a path and fails that file.
`docs/sync/migrate.md` tells consumers what to delete.

## The Claude plugins

The sync does not read `[tools.claude.plugins.user]`,
`[tools.claude.plugins.project]` and `tools.claude.plugins.installed` of
`tools.toml`. `setup:ai` of the `claude` tool reads them with
`bootstrap config get` when it runs, as the merge tasks read `[branches.*]`, so
it gets the entries that other selected tools contribute in
`[contributes.claude]` too, for example `vwf@virajp-plugins` of
`virajp-plugins`. A failed read, for example with no catalog (code 2), stops
it. The three are `[defaults]` of the tool, empty by default. Each key of a
plugin table is `"<plugin>@<marketplace>" = "<source>"`. `setup:ai` names no
plugin.

`setup:ai` reads the installed plugins once, then:

1. It drops from the record `plugins.installed` each plugin that is no longer
   installed at project scope.
2. It adds a missing marketplace from the source, installs each listed plugin
   that is missing at its scope, and records each listed project plugin that is
   installed, only after its install succeeds.
3. It uninstalls each recorded plugin that left `[tools.claude.plugins.project]`
   with `claude plugin uninstall --scope project --prune -y`. A failed
   uninstall keeps it in the record. It never uninstalls a plugin at user
   scope, or a plugin that is not in the record.
4. With `--upgrade` only, it updates the marketplaces and each listed plugin
   that is installed.
5. It writes the record with `bootstrap config set` when it changed. A failed
   write is a failed step (exit 1).

A plain run with nothing to change changes nothing.
