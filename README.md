# bootstrap

Shared configuration files, mise task scripts and templates for my
repositories. Each repository selects its tools, and its `setup:configs` task
copies the files of those tools.

## Adopt in a repository

A repository uses this repository in two ways, in this sequence:

1. **Adoption, one time, by an AI agent.** [`docs/sync/`](docs/sync/index.md)
   tells the agent how to apply the rules of this repository to a consumer
   repository. The agent makes sure that each rule is relevant to that
   repository. `docs/sync/index.md` has the general steps, `migrate.md` has
   the items for a repository with older copies, and one file for each tool
   tells when the tool is relevant and how to use it.
2. **Sync, each time after that, by `repo-sync`.** The adoption installs the
   `setup:configs` task, which keeps the shared files equal to this
   repository.

Copy this prompt into an AI session in the consumer repository:

```text
Make this repository and each of its git submodules get their shared files
from bootstrap. Make a shallow clone of
git@github.com:virajp/bootstrap.git into a temporary directory. Read
docs/sync/index.md in the clone, and do its steps in sequence. Read the other
files of docs/sync/ when index.md tells you to. Apply only what is relevant to
this repository. Ask me before each choice that docs/sync/ does not decide.
```

## Layout

The shared files are in `tools/`, one folder for each tool. This repository
also syncs its own `base` tool. Each file in `tools/<tool>/` is at the path that
it gets in a repository. For example,
`tools/base/.config/mise/tasks/code/merge/develop` here is
`.config/mise/tasks/code/merge/develop` in every repository that selects
`base`.

Each tool folder has a `tool.toml` manifest:

```toml
# Flutter and Dart, with the Android toolchain.
depends = [] # tools added with this one
extensions = [ "dart-code.dart-code", "dart-code.flutter" ] # VS Code extensions
root_only = false # true: refused in a submodule
seed = [] # written only when missing

# The mise config, in mise syntax. The sync generates
# .config/mise/conf.d/flutter/mise.<env>.toml.
[mise.tools.android-sdk]
depends = [ "java" ] # also adds the tool that defines java
envs = [ "dev", "ci" ] # the default is [ "dev" ]
version = "23"

# Content for one environment only
[mise_env.ci.env]
ANDROID_HOME = "/opt/android"
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

1. `setup:configs` clones the default branch of this repository once
   (shallow, over SSH).
2. It runs `setup:configs:sync`, which loads `repo-sync` from that clone and
   calls `repo_sync`. Thus a change to the sync code applies in the same run.
3. `repo_sync` reads `tools` from the repository's `.config/bootstrap.toml`.
4. It adds the `depends` of each tool. It refuses a `root_only` tool in a git
   submodule.
5. It finds the branch model from the branches `main` and `develop`, local or
   on origin. It stops if neither branch exists.
6. It copies each file of each tool.

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
  (necessary), `ci`, and `prod` and `stage` (only with `ci`).
- The mise config of each tool is generated from its `[mise.*]` tables into
  `.config/mise/conf.d/<tool>/mise.<env>.toml` (`_base` for `base`), one file
  for each environment. A tool that the repository lists in `tools` goes to the
  environments of its `envs`. A tool that only a dependency adds goes to dev,
  unless a tool in another environment needs it. Aliases, hooks, vars and
  tasks go only to dev. `.config/mise.toml` is the only file for every
  environment. `CLAUDE.md` has the full rules.
- The sync never deletes a file, with two exceptions. The first is the merge
  tasks. With `main` and `develop`, the repository gets `code:merge:develop`
  and `code:merge:main`. With only one of them, it gets `code:merge`. The sync
  removes the merge tasks of the other model. The second is a
  `.config/mise/conf.d/*/mise*.toml` file with the bootstrap header that the
  sync did not write. A seed, such as `_repo/mise.<env>.toml`, has no header, so
  the sync keeps it.

A copy of `repo-sync` from before the `tools/` layout cannot sync. Copy
`tools/base/.config/mise/tasks/_scripts/repo-sync` and `helpers` into the
repository by hand one time. Step 2 of `docs/sync/index.md` tells you how.

## Templates

A file whose name ends in `.tmpl` is a template. `repo_sync` renders it with
[gomplate](https://docs.gomplate.ca) and writes it without the `.tmpl` suffix.

- The delimiters are `<% %>`. mise uses `{{ }}` and bash uses `[[ ]]`, so
  both reach the output unchanged.
- The values come from the repository's `.config/bootstrap.toml`.
- `.members` is the list of member names: the git submodules (else `MEMBERS`),
  then each `[member.<name>]` table whose folder exists in the repository.
- `.submodules` is the list of the git submodules only. It is empty in a
  repository without submodules.
- `.tools` is the list of the selected tools, dependencies included.
- `.extensions` is the sorted list of the VS Code extensions of those tools,
  plus the `extensions` list of the values file. In the root repository, it
  also has the extensions of each git submodule: the tools and the
  `extensions` list of its values file.
- `.precommit` is the list of the `[[precommit]]` entries of those tools, in
  tool order, then the entries of the repository's
  `.config/repo-sync/pre-commit.yaml`.

Example values file:

```toml
commit_scopes = [ "backend", "web" ]
environments = [ "dev", "ci", "prod" ]
merge_model_develop = "direct"
merge_model_main = "pr"
product_name = "my-product"
repo_name = "my-repo"
repo_url = "https://github.com/me/my-repo"
tools = [ "base", "github", "node", "fnox", "claude", "vscode" ]
worktree_files = [ ".config/mise.local.toml" ]

[member.backend]
label = "Backend"
worktree_install = "pnpm install --frozen-lockfile"
```

## `.gitignore`

`repo_sync` writes the shared `.gitignore` first. Then it adds the lines of the
repository's `.config/repo-sync/gitignore`, if that file exists. A later line
wins, so a repository can negate a shared pattern with `!`.

## `.config/pre-commit-config.yaml`

`base` renders this file from the `[[precommit]]` entries of the selected
tools. Then it adds the entries of the repository's
`.config/repo-sync/pre-commit.yaml`, if that file exists. That file is a YAML
list of `repos:` entries:

```yaml
- repo: local
  hooks:
    - id: my-hook
      name: My hook
      entry: mise x -- mise run my:task
      language: system
```

## Change a shared file

1. Change the file in `tools/<tool>/`, and push to `main`.
2. In each repository, run `mise run setup:configs`, then commit the result.
