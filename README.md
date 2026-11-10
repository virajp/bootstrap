# bootstrap

Shared configuration files, mise task scripts and templates for my
repositories, and the CLI `bootstrap` that syncs them. Each repository selects
its tools, and `bootstrap sync` copies the files of those tools.

## Install the CLI

mise installs the CLI from the GitHub releases of this repository:

```bash
mise use -g github:virajp/bootstrap
```

| Command                          | Job                                                                              |
| -------------------------------- | -------------------------------------------------------------------------------- |
| `bootstrap init [tool…]`         | Select `base` and the named tools, migrate an old form, ask the questions, sync. |
| `bootstrap sync`                 | Sync the repository of the current folder. `setup:configs` runs it.              |
| `bootstrap check`                | Examine the values file, and run the check scripts of the tools.                 |
| `bootstrap questions [tool…]`    | List the questions, their proposals and their answers. `--json` for an agent.    |
| `bootstrap detect`               | Run the detect scripts, and offer to add each tool that matches.                 |
| `bootstrap add <tool…>`          | Add the tools and their dependencies, then sync.                                 |
| `bootstrap remove <tool…>`       | Remove the tools and the dependencies that no tool needs, then sync.             |
| `bootstrap replace <old> <new>`  | Move from one tool to another with the same parent, then sync.                   |
| `bootstrap list`, `bootstrap ls` | List the selected tools, with the reason of each one.                            |
| `bootstrap --version`            | Print the version.                                                               |
| `bootstrap mac <command>`        | The same commands for the Mac, with `macos.toml`.                                |

Each writing command shows its change list first. `--dry-run` writes nothing,
and `--yes` writes without a prompt.
[`docs/design/cli.md`](docs/design/cli.md#the-command-line) has
the rules.

## Adopt in a repository

A repository uses this repository in two ways, in this sequence:

1. **Adoption and update, by an AI agent.** [`docs/sync/`](docs/sync/index.md)
   tells the agent how to apply the rules of this repository to a consumer
   repository, the first time and each time that the repository is not up to
   date. The agent selects the tools, gets the answers of their questions, and
   runs `bootstrap init`. `docs/sync/index.md` has the general steps,
   `migrate.md` has the items for a repository with older copies, one file for
   each tool in `docs/sync/tools/` tells when the tool is relevant and how to
   use it, and one file for each pack is in `docs/sync/packs/`.
   `docs/sync/vwf.md` is the entry point for the vwf and stackgen plugins: they
   give it the `init`, `add`, `remove`, `add plugins` and `remove plugins`
   instructions.
2. **Sync, each time after that, by `bootstrap sync`.** The adoption installs
   the `setup:configs` task, which runs `bootstrap sync` and keeps the shared
   files equal to this repository.

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

The shared files are in `repo/tools/`, one folder for each tool. The
`tools.toml` of a tool gives the path that each file of its `files/` and
`tasks/` folders gets in a repository. For example,
`repo/tools/git/tasks/code-merge-develop` here is
`.config/mise/tasks/code/merge/develop` in every repository with `main` and
`develop`, because the `base` pack adds the `git` tool. The `scripts/` folder
of a tool holds its scripts, which the CLI runs from its clone and never copies.
This repository also syncs its own `base` pack.

The Mac tools are in `macos/tools/`, one folder for each tool, and their packs
are in `macos/packs.toml`. `macos/examples/macos.toml` is the sample values
file of the Mac side. The task `updateall:configs` of the Mac tool `mise` is
the entry of the Mac sync.

The sample values files are in `repo/examples/`.

The code of the CLI is in `cli/`: the Go module (`cmd/` and
`internal/`) and `VERSION`, the minimum version of the CLI that a clone needs.
This folder is not a tool, so the sync does not copy it to a repository.

A pack is a named group of tools in `repo/packs.toml`. It has no files of its
own. The sync always adds the `base` pack:

<!-- dprint-ignore -->
```toml
[pack.base]
mise = { default = true, required = true }
git = { default = true, required = true }
pre-commit = { default = true, required = true }
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
nested pack. `required = true` marks a tool that `bootstrap remove` refuses.
[`docs/design/packs.md`](docs/design/packs.md) has the rules.

Each tool folder has a `tools.toml` manifest:

```toml
# The Android SDK command-line tools.
default_envs = [
  "dev",
  "ci",
] # the environments when the values file gives no env
depends = [ "java" ] # tools added with this one
root_only = false # true: refused in a submodule; skipped there from a pack

# The mise config, in mise syntax. The sync generates
# .config/mise/conf.d/android/mise.<env>.toml.
[mise.tools.android-sdk]
envs = [ "dev", "ci" ] # the allowed environments; the default is [ "dev" ]
version = "23"
```

The sync removes the prefix `mise.` from each table: `[mise.settings]` becomes
`[settings]` of the generated file. The section decides the environments that
get it. The sync computes the mise `depends` of each mise tool from `depends`
and `parent`. [`docs/design/mise.md`](docs/design/mise.md) has the full list and
an example.

A tool can also have the optional tables `[defaults]` (its values),
`[questions.<key>]` (the question for one value) and `[contributes.<tool>]`
(values for another selected tool). For example, a tool adds a hook to
`.config/pre-commit-config.yaml` with a contribution to `pre-commit`:

```toml
[[contributes.pre-commit.repos]]
id = "format"
repo = "local"

[[contributes.pre-commit.repos.hooks]]
entry = "mise x -- mise run code:format --fix"
id = "format"
language = "system"
name = "Format (mise run code:format)"
```

Each file of `files/` and `tasks/` has a `[files."<target>"]` or a
`[tasks."<path>"]` table:

```toml
[files."pnpm-workspace.yaml"]
mode = "seed" # copy, template, seed or merge
requires = [ "package.json" ] # exact paths; each must exist
source = "pnpm-workspace.yaml"
```

[`docs/design/tools.md`](docs/design/tools.md) has each key of a `tools.toml`,
of the `[files]` and `[tasks]` tables, and the scripts.

## How a repository syncs

1. `setup:configs` runs `bootstrap sync`. It fetches the branch `main` of
   `https://github.com/virajp/bootstrap.git` into a shared clone in the cache
   `${XDG_CACHE_HOME:-~/.cache}/bootstrap/`, copies that clone, and stops when
   the CLI is older than `cli/VERSION` of the clone.
2. It reads the `[tools.*]` and `[packs.*]` tables of the repository's
   `.config/bootstrap.toml`, and adds the `base` pack and the `depends` of each
   tool.
3. It merges the values of each tool, runs the derive and check scripts, and
   generates the mise config of each tool.
4. It writes each file of each tool. An owned file changes on each sync, a seed
   is written only when it is missing, and a shared file gets the bootstrap
   items and keeps the changes of the user.
5. After a sync without a failure, it removes each stale file, it writes
   `.config/repo-generated.tsv`, and it sets `bootstrap_commit` of the values
   file to the commit of its clone.

Between two syncs, the gate `code:lint:60-bootstrap` runs `bootstrap check` at
each commit that changes `.config/bootstrap.toml`.

The sync needs `mise` and a git repository. It lists only the files that
changed, and it ends with a table of counts. Use `--dry-run` to see the changes
without writing them. Use `--verbose` of `setup:configs` to list each file,
also an unchanged one. The rules are in `docs/design/`:

- [`cli.md`](docs/design/cli.md): the command line, the clone,
  the scripts, the write of one file, shared files, the header, the manifest and
  the stale rule, the values, `bootstrap_commit` and the exit codes.
- [`repo.md`](docs/design/repo.md): the entry tasks, members, the branch model,
  `direct_commit`, the pre-commit config, `bootstrap check` and its gate.
- [`mise.md`](docs/design/mise.md): the generated mise config, the environments
  and `[overrides.mise*]`.
- [`bootstrap-config.md`](docs/design/bootstrap-config.md): the keys of
  `.config/bootstrap.toml`, and its old forms.

A repository with an older form of the values file runs `bootstrap init` once:
it migrates each old form. `docs/sync/migrate.md` has the manual items.

## Templates

A file with `mode = "template"` is a template. `bootstrap sync` renders it with
Go `text/template` and its own copy of the [gomplate](https://docs.gomplate.ca)
functions. The delimiters are `<% %>`. mise uses `{{ }}` and bash uses `[[ ]]`,
so both reach the output unchanged. The values come from the tools and from the
repository's `.config/bootstrap.toml`.
[`docs/design/cli.md`](docs/design/cli.md#templates) has the
rules, and
[`docs/design/repo.md`](docs/design/repo.md#the-template-context) has each value
of the context.

Example values file:

```toml
environments = [ "dev", "ci" ]
product_name = "my-product"
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"

[tools.claude]
# The Claude Code plugins that setup:ai of the claude tool installs.
[tools.claude.plugins.user]
"taste-skill@taste-skill" = "Leonxlnx/taste-skill"

[tools.fnox]

[tools.git]
commit_scopes = [ "backend", "web" ]

# The merge config of this repository.
[tools.git.branch.main]
merge_from = [ "develop" ]
merge_model = "pr"

[tools.git.branch.develop]
# Allow a direct commit to develop. main can allow it only when develop does.
# direct_commit = true

[tools.github]

[tools.mise]
worktree_files = [ ".config/mise.local.toml" ]

[tools.node]
env = [ "dev", "ci" ] # optional; without it, the default_envs of the tool

[member.backend]
folder = "projects/backend" # optional; the default is the name
label = "Backend"
```

## Shared files

`.gitignore`, `.config/pre-commit-config.yaml`, `dprint.jsonc` and
`.vscode/extensions.json` are shared files. The sync adds the items of the
selected tools, and keeps each item that you add to the file: own lines, own
hook blocks, own excludes, own extensions. A removed bootstrap item comes back
at the next sync, unless the values file has an override for it:

```toml
[tools.git.files.".gitignore"]
"*.log" = { "$mode" = "remove" }

[tools.pre-commit.files.".config/pre-commit-config.yaml".repos]
"conventional-commits" = { "$mode" = "remove" }
```

[`docs/design/cli.md`](docs/design/cli.md#shared-files) has the
rules.

## Tests

`mise run test:repo` runs `bootstrap sync --dry-run`, `bootstrap check` and,
for a case that compares a written file, `bootstrap sync`, built from the
working tree's `cli/`, with its `repo/`, against scratch values files
and scratch consumers. The init cases run `bootstrap init`. It checks the exit
code and the message of each case. It needs no network when gomplate, jq and yq
are installed.

`mise run test:mac` runs `bootstrap mac sync`, built from the working tree's
`cli/`, against the fixtures in `macos/fixtures/`. It checks the exit
code, the output and the files of each case, and it needs no network when
gomplate, jq and yq are installed. The entry cases run the rendered task
`updateall:configs` against a local bare origin.

`mise run test:macos-tools` runs `bootstrap mac sync` with each tool of
`macos/tools/` and a copy of `macos/examples/macos.toml`, in a temporary `HOME`.
It checks that the example selects each tool and that each rendered file
parses, and it needs no network when gomplate, jq and yq are installed.

`mise run test:go` runs `go vet`, a `gofmt` check and `go test` in
`cli/`. A test fails when the engine code names a tool.

`mise run test` runs each test task in parallel. No hook runs a test. Run the
tests by hand before you commit: `test` for a change to `cli/`,
`test:repo` for a change to `repo/`, and `test:mac` and `test:macos-tools` for
a change to `macos/`.

## Change a shared file

1. Change the file in `repo/tools/<tool>/`, and push to `main`.
2. In each repository, run `mise run setup:configs`, then commit the result.
