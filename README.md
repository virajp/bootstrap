# bootstrap

Shared configuration files, mise task scripts and templates for my
repositories. Each repository selects its tools, and its `setup:configs` task
copies the files of those tools.

## Adopt in a repository

A repository uses this repository in two ways, in this sequence:

1. **Adoption and update, by an AI agent.** [`docs/sync/`](docs/sync/index.md)
   tells the agent how to apply the rules of this repository to a consumer
   repository, the first time and each time that the repository is not up to
   date. The agent makes sure that each rule is relevant to that
   repository. `docs/sync/index.md` has the general steps, `migrate.md` has
   the items for a repository with older copies, one file for each tool in
   `docs/sync/tools/` tells when the tool is relevant and how to use it, and
   one file for each pack is in `docs/sync/packs/`. `docs/sync/vwf.md` is
   the entry point for the vwf and stackgen plugins: they give it the
   `adopt`, `add tools`, `remove tools`, `add plugins` and `remove plugins`
   instructions.
2. **Sync, each time after that, by `repo-sync`.** The adoption installs the
   `setup:configs` task, which keeps the shared files equal to this
   repository.

Copy this prompt into an AI session in the consumer repository. It does the
first adoption, and it also brings an adopted repository up to date:

```text
Make this repository, each of its git submodules and each of its sibling
repositories get their shared files from bootstrap. Clone
git@github.com:virajp/bootstrap.git with --filter=blob:none into a temporary
directory. Read docs/sync/index.md in the clone, and do its steps in sequence. Read the other
files of docs/sync/ when index.md tells you to. Apply only what is relevant to
this repository. Ask me before each choice that docs/sync/ does not decide.
```

## Layout

The shared files are in `tools/`, one folder for each tool. Each file in
`tools/<tool>/` is at the path that it gets in a repository. For example,
`tools/git/.config/mise/tasks/code/merge/develop` here is
`.config/mise/tasks/code/merge/develop` in every repository, because the
`base` pack adds the `git` tool. This repository also syncs its own `base`
pack.

A pack is a named list of tools in `tools/packs.toml`. It has no files of its
own. The sync always adds the `base` pack:

```toml
[pack.base]
depends = [
  "mise",
  "repo-sync",
  "git",
  "pre-commit",
  "dprint",
  "linter",
  "p:security",
]

[pack.security]
depends = [ "gitleaks", "grype", "osv-scanner" ]
required = [ "_security_base" ] # always added with the pack
```

A name with the prefix `p:` is a pack.

Each tool folder has a `tools.toml` manifest:

```toml
# The Android SDK command-line tools.
default_envs = [
  "dev",
  "ci",
] # the environments when the values file gives no env
depends = [] # tools added with this one
extensions = [] # VS Code extensions
root_only = false # true: refused in a submodule; skipped there from a pack
seed = [] # written only when missing

# The mise config, in mise syntax. The sync generates
# .config/mise/conf.d/android/mise.<env>.toml.
[mise.tools.android-sdk]
depends = [ "java" ] # also adds the tool that defines java
envs = [ "dev", "ci" ] # the allowed environments; the default is [ "dev" ]
version = "23"
```

The sync removes the prefix `mise.` from each table: `[mise.settings]` becomes
`[settings]` of the generated file. The section decides the environments that
get it. `tools`, `tool_alias` and `plugins` follow the `envs` of each entry.
`settings`, `env` and the other folder sections go to each environment of the
tool. `shell_alias`, `tasks` and the other dev-only sections go only to `dev`.
`docs/design/mise.md` ("The `[mise.*]` sections") has the full list and an
example.

A `[mise_env.<env>.<section>]` table adds content to one environment only.
For example, the `mise` tool sets the mise `os` setting in `ci`:

```toml
[mise_env.ci.settings]
os = "linux"
```

A tool can also add hooks to `.config/pre-commit-config.yaml` with
`[[precommit]]` tables. Each table is one `repos:` entry, with its
`[[precommit.hooks]]`:

```toml
[[precommit]]
repo = "local"

[[precommit.hooks]]
entry = "mise x -- mise run code:format --fix"
id = "format"
language = "system"
name = "Format (mise run code:format)"
```

A tool can also make a seed depend on a file of the repository with a
`[seed_requires]` table. The sync then writes the seed only if the repository
has that file:

```toml
[seed_requires]
"pnpm-workspace.yaml" = "package.json"
```

## How a repository syncs

1. `setup:configs` fetches the default branch of this repository (shallow,
   over SSH) into a clone in `${XDG_CACHE_HOME:-~/.cache}/bootstrap`. Each
   consumer of the machine uses that clone. A lock makes a second sync wait
   until the first has copied the clone.
2. It runs `setup:configs:sync`, which loads `repo-sync` from that copy and
   calls `repo_sync`. Thus a change to the sync code applies in the same run.
3. `repo_sync` reads the `[tools.*]` and `[packs.*]` tables of the
   repository's `.config/bootstrap.toml`, and it adds the `base` pack.
4. It expands each pack, and it adds the `depends` of each tool. It refuses a
   `root_only` tool in a git submodule. It skips a `root_only` tool that only
   a pack adds.
5. It finds the branch model from the branches `main` and `develop`, local or
   on origin. It stops if neither branch exists. It validates `direct_commit`
   of `[branch.main]` and `[branch.develop]`, and stops with code 2 for a bad
   value.
6. It copies each file of each tool.
7. After a sync without a failure, it sets `bootstrap_commit` of the values
   file to the commit of its clone. This is its only write to the values
   file. An agent compares the value with `main` to find what changed since
   the last sync.

These rules apply to each file:

- A file is only replaced when its content changes. Use `--dry-run` to see the
  changes without writing them.
- A file takes the mode of its source, also when the file exists already. A
  template keeps the mode of its `.tmpl`, so an executable `.tmpl` gives an
  executable file.
- A seed file is written only when it is missing. After that, the repository
  owns it. A seed in `seed_requires` is written only if the repository has the
  named file.
- In `.config/pre-commit-config.yaml`, each hook `rev` becomes the newest tag:
  the sync runs `pre-commit autoupdate` on it. This needs the network.
- The repository selects its environments in `environments`: `dev`
  (necessary) and `ci`.
- The mise config of each tool is generated from its `[mise.*]` tables into
  `.config/mise/conf.d/<tool>/mise.<env>.toml`, one file for each environment.
  A tool of a `[tools.*]` or a `[packs.*]` entry goes to the environments of
  its `env`, else of its `default_envs`. A dependency goes to each environment
  of the tools that need it. The `envs` of a mise tool lists the environments
  that it allows; a tool in a different environment stops the sync. The
  `requires` of a mise tool names a file: the sync writes the mise tool only
  when the repository has that file. Its `files` lists the files of the tool,
  for example its tasks, that follow the same condition. Each tool
  must be in dev: an `env` or a `default_envs` without dev stops the sync.
  Aliases, hooks, vars and tasks go only to dev. `.config/mise.toml` is the
  only file for every environment. `CLAUDE.md` has the full rules.
- `[overrides]` of the values file replaces a default of a tool, for example
  `[overrides.mise.tools.java]` with `version = "21"`. Each overridden line
  gets an `# override: .config/bootstrap.toml` comment.
  `docs/sync/index.md` has the rules.
- The sync never deletes a file, with four exceptions. The first is the merge
  tasks. With `main` and `develop`, the repository gets `code:merge:develop`
  and `code:merge:main`. With only one of them, it gets `code:merge`. The sync
  removes the merge tasks of the other model. The second is a
  `.config/mise/conf.d/*/mise*.toml` file with the bootstrap header that the
  sync did not write. A seed, such as `zz-repo/mise.<env>.toml`, has no header,
  so the sync keeps it. The third is a file with the bootstrap header at a path
  where the sync needs a folder. The fourth is a file with the bootstrap header
  of a `files` list of a mise tool, when the `requires` file of that mise tool
  is missing.

A repository synced before the split of the base tool (`c95e4ff`) cannot sync.
Its
`setup/configs/sync` loads `tools/base/.../repo-sync` from the clone, and the
`tools/base/` folder no longer exists. Copy the sync files into the repository
by hand one time, as step 2 of `docs/sync/index.md` tells you.

## Templates

A file whose name ends in `.tmpl` is a template. `repo_sync` renders it with
[gomplate](https://docs.gomplate.ca) and writes it without the `.tmpl` suffix.

- The delimiters are `<% %>`. mise uses `{{ }}` and bash uses `[[ ]]`, so
  both reach the output unchanged.
- The values come from the repository's `.config/bootstrap.toml`.
- `.members` is the list of member names: the git submodules, then each other
  `[member.<name>]` table, whose `folder` (default the name) must exist.
  `.member_folders` and `.member_types` give the folder and the type
  (`submodule` or `folder`) of each name.
- `.submodules` is the list of the git submodules only. It is empty in a
  repository without submodules.
- `.tools` is the list of the selected tools, packs expanded and dependencies
  included.
- `.extensions` is the sorted list of the VS Code extensions of those tools,
  plus the `extensions` list of the values file. In the root repository, it
  also has the extensions of each git submodule: the tools and the
  `extensions` list of its values file.
- `.precommit` is the list of the `[[precommit]]` entries of those tools, in
  tool order, then the entries of the repository's
  `.config/repo-sync/pre-commit.yaml`. A `no-commit-to-branch` hook in it
  loses each `--branch` pair of a core branch that does not exist or that
  has `direct_commit = true`. A hook with no pair left goes, and an entry with
  no hook left goes too.

Example values file:

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
env = [ "dev", "ci" ] # optional; without it, the default_envs of the tool

[member.backend]
folder = "projects/backend" # optional; the default is the name
label = "Backend"

# The merge config of this repository.
[branch.main]
merge_from = [ "develop" ]
merge_model = "pr"

[branch.develop]
# Allow a direct commit to develop. main can allow it only when develop does.
# direct_commit = true

# The Claude Code plugins that setup:ai of the claude tool installs.
[claude.plugins.user]
"taste-skill@taste-skill" = "Leonxlnx/taste-skill"
```

## `.gitignore`

`repo_sync` writes the shared `.gitignore`, with `[overrides.gitignore]` of the
values file applied. It comments out each line of `remove`, then it adds the
lines of `add` at the end. A later line wins, so a line of `add` can also
negate a shared pattern with `!`. `docs/sync/index.md` has the rules.

The `claude` tool ships `.graphifyignore`. The sync applies
`[overrides.graphifyignore]` to it in the same way.

## `.config/pre-commit-config.yaml`

The `pre-commit` tool renders this file from the `[[precommit]]` entries of the
selected tools. Then it adds the entries of the repository's
`.config/repo-sync/pre-commit.yaml`, if that file exists. The args of the
`no-commit-to-branch` hook follow the core branches (`main`, `develop`) that
exist and `direct_commit` of their `[branch.*]` tables: the hook protects
only the core branches that exist and do not allow direct commits. Without a
protected branch, the sync removes the hook.
`.config/repo-sync/pre-commit.yaml` is a YAML list of `repos:` entries:

```yaml
- repo: local
  hooks:
    - id: my-hook
      name: My hook
      entry: mise x -- mise run my:task
      language: system
```

## Tests

`mise run test:repo-sync` runs `repo_sync --dry-run` of the working tree's
`tools/` against scratch values files, and checks the exit code and the
message of each case. It needs no network. No hook runs it: run it by hand
before you commit a change to `tools/`.

## Change a shared file

1. Change the file in `tools/<tool>/`, and push to `main`.
2. In each repository, run `mise run setup:configs`, then commit the result.
