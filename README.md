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

The owner of a Mac uses [`docs/setup/index.md`](docs/setup/index.md), not this
prompt.

## Layout

The shared files are in `repo/tools/`, one folder for each tool. Each file in
`repo/tools/<tool>/` is at the path that it gets in a repository. For example,
`repo/tools/git/.config/mise/tasks/code/merge/develop` here is
`.config/mise/tasks/code/merge/develop` in every repository, because the
`base` pack adds the `git` tool. This repository also syncs its own `base`
pack.

The Mac tools are in `macos/tools/`, one folder for each tool, and their packs
are in `macos/packs.toml`. `macos/examples/macos.toml` is the sample values
file of the Mac side. The task `updateall:configs` of the tool `macos-setup`
is the entry of the Mac sync.

The sample values files are in `repo/examples/`. `tools/` holds only a
forwarding file, for an old sync task of a repository. The sync does not copy
it.

The sync code is in `sync-engine/`: `clone` for the shared cache and the clone,
`common` for the log, the header detection and the wrappers of gomplate, jq and
yq, `write` for the write of one file, `merge` for merge mode, `select` for the
packs and the `depends`, `manifest` for the manifest and the stale rule,
`values`
for the merge of the values of the tools, `mise` for the mise config generator,
`repo-sync` for the function `repo_sync`, and `setup-sync` for the function
`setup_sync`, the Mac front-end. This
folder is not a tool, so the
sync does not copy it to a repository.

A pack is a named group of tools in `repo/packs.toml`. It has no files of its
own. The sync always adds the `base` pack:

<!-- dprint-ignore -->
```toml
[pack.base]
mise = { default = true }
repo-sync = { default = true }
git = { default = true }
pre-commit = { default = true }
dprint = { default = true }
linter = { default = true }
vscode = { default = true }
"pack:security" = { default = true }

[pack.security]
gitleaks = { default = true }
grype = { default = true }
osv-scanner = { default = true }
```

`[packs.<name>]` of a values file adds each `default = true` member. A member
with `{}` arrives only with `[tools.<name>]`. A key with the prefix `pack:` is a
nested pack. [`docs/design/packs.md`](docs/design/packs.md) has the rules.

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

# The mise config, in mise syntax. The sync generates
# .config/mise/conf.d/android/mise.<env>.toml.
[mise.tools.android-sdk]
depends = [ "java" ] # also adds the tool that defines java
envs = [ "dev", "ci" ] # the allowed environments; the default is [ "dev" ]
version = "23"
```

The sync removes the prefix `mise.` from each table: `[mise.settings]` becomes
`[settings]` of the generated file. The section decides the environments that
get it. [`docs/design/mise.md`](docs/design/mise.md) ("The `[mise.*]` sections")
has the full list and an example.

A tool can also have the optional tables `[defaults]` (its values),
`[questions.<key>]` (the question for one value) and `[contributes.<tool>]`
(values for another selected tool). `docs/design/tools.md` has their keys and
their checks.

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

A tool can also give one of its files a `[files."<target>"]` table. A file
without a table is `generated`: the sync writes it on each run.

```toml
[files."pnpm-workspace.yaml"]
mode = "seed" # generated, seed or merge
requires = [ "package.json" ] # exact paths; each must exist
```

[`docs/design/tools.md`](docs/design/tools.md) has each key of a `tools.toml`
and
of the `[files]` table.

## How a repository syncs

1. `setup:configs` fetches the branch `main` of
   `https://github.com/virajp/bootstrap.git` into a shared clone in the cache
   `${XDG_CACHE_HOME:-~/.cache}/sync-engine/`, and it copies that clone.
2. It updates the local `setup/configs/sync` and `_scripts/helpers` from that
   copy. Then it runs `setup:configs:sync`, which loads `sync-engine/repo-sync`
   from that copy and calls `repo_sync`. Thus a change to the sync code applies
   in the same run.
3. `repo_sync` reads the `[tools.*]` and `[packs.*]` tables of the repository's
   `.config/bootstrap.toml`, adds the `base` pack and the `depends` of each
   tool,
   and finds the branch model from the branches `main` and `develop`.
4. It generates the mise config of each tool, and it writes each file of each
   tool. A generated file changes on each sync, and a seed is written only when
   it is missing. A merge target gets only the keys that the tool enforces.
5. After a sync without a failure, it removes each stale file, it writes
   `.config/repo-generated.tsv`, and it sets `bootstrap_commit` of the values
   file to the commit of its clone.

Between two syncs, the gate `code:lint:repo-sync` runs `repo_sync --check` at
each commit that changes `.config/bootstrap.toml`.

The sync needs `mise`. It lists only the files that changed, and it ends with a
table of counts. Use `--dry-run` to see the changes without writing them. Use
`--verbose` to list each file, also an unchanged one.
The rules are in `docs/design/`:

- [`sync-engine.md`](docs/design/sync-engine.md): the clone, the write of one
  file, merge mode, the header, the manifest and the stale rule, the values
  merge, `bootstrap_commit` and the exit codes.
- [`repo-sync.md`](docs/design/repo-sync.md): the entry tasks, members, the
  branch model, `direct_commit`, the pre-commit config, `--check` and its gate.
- [`mise.md`](docs/design/mise.md): the generated mise config, the environments
  and `[overrides.mise*]`.
- [`bootstrap-config.md`](docs/design/bootstrap-config.md): the keys of
  `.config/bootstrap.toml`.

A repository synced before the split of the base tool (`c95e4ff`) cannot sync.
Its
`setup/configs/sync` loads `tools/base/.../repo-sync` from the clone, and the
`tools/base/` folder no longer exists. Copy the sync files into the repository
by hand one time, as step 2 of `docs/sync/index.md` tells you.

## Templates

A file whose name ends in `.tmpl` is a template. `repo_sync` renders it with
[gomplate](https://docs.gomplate.ca) and writes it without the `.tmpl` suffix.
The delimiters are `<% %>`. mise uses `{{ }}` and bash uses `[[ ]]`, so both
reach the output unchanged. The values come from the repository's
`.config/bootstrap.toml`.
[`docs/design/sync-engine.md`](docs/design/sync-engine.md#templates)
has the rules, and
[`docs/design/repo-sync.md`](docs/design/repo-sync.md#the-template-context)
has each value of the context.

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
values file applied.
[`docs/design/repo-sync.md`](docs/design/repo-sync.md#overridesgitignore) has
the
rules.

## `.config/pre-commit-config.yaml`

The `pre-commit` tool renders this file from the `[[precommit]]` entries of the
selected tools. Then it adds the entries of the repository's
`.config/repo-sync/pre-commit.yaml`, if that file exists.
[`docs/design/repo-sync.md`](docs/design/repo-sync.md#the-pre-commit-config) has
the rules. `.config/repo-sync/pre-commit.yaml` is a YAML list of `repos:`
entries:

```yaml
- repo: local
  hooks:
    - id: my-hook
      name: My hook
      entry: mise x -- mise run my:task
      language: system
```

## Tests

`mise run test:repo-sync` runs `repo_sync --dry-run`, `repo_sync --check` and,
for a case that compares a written file, `repo_sync` of the working tree's
`repo/`, `sync-engine/` and `tools/` (the forwarding file) against scratch
values files and scratch consumers.
Some cases of the manifest also call `_se_manifest_stale` and
`_se_manifest_read` directly, and the cases of the `macos` side call
`sync-engine/mise` directly. It checks the exit code and the message of each
case. It needs no network when gomplate, jq and yq are installed.

`mise run test:setup-sync` runs `setup_sync` of the working tree's
`sync-engine/` against the fixtures in `macos/fixtures/`. It checks the exit
code, the output and the files of each case, and it needs no network when
gomplate, jq and yq are installed. The entry cases run the rendered task
`updateall:configs` against a local bare origin.

`mise run test:macos-tools` runs `setup_sync` with each tool of `macos/tools/`
and a copy of `macos/examples/macos.toml`, in a temporary `HOME`. It checks that
the example selects each tool and that each rendered file parses, and it needs
no network when gomplate, jq and yq are installed.

`mise run test` runs each test task in parallel. No hook runs a test. Run the
tests by hand before you commit: `test` for a change to `sync-engine/`,
`test:repo-sync` for a change to `repo/`, and `test:setup-sync` and
`test:macos-tools` for a change to `macos/`.

## Change a shared file

1. Change the file in `repo/tools/<tool>/`, and push to `main`.
2. In each repository, run `mise run setup:configs`, then commit the result.
