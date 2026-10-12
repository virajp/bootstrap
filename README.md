# repo-config

Shared configuration files, mise task scripts and templates for my
repositories. Each repository selects its tools, and its `setup:configs` task
copies the files of those tools.

## Layout

The shared files are in `tools/`, one folder for each tool. This repository
does not use them itself. Each file in `tools/<tool>/` is at the path that it
gets in a repository. For example,
`tools/base/.config/mise/tasks/code/merge/develop` here is
`.config/mise/tasks/code/merge/develop` in every repository that selects
`base`.

Each tool folder has a `tool.toml` manifest:

```toml
# Flutter and Dart, with the Android toolchain.
depends    = []                                          # tools added with this one
root_only  = false                                       # true: refused in a submodule
extensions = ["dart-code.dart-code", "dart-code.flutter"] # VS Code extensions
seed       = []                                          # written only when missing
```

## Adopt in a repository

[`sync.md`](sync.md) tells an AI agent how to make a repository use these
files, and how to migrate a repository that has older copies. Copy this
prompt into an AI session in that repository:

```text
Make this repository and each of its git submodules get their shared files
from repo-config. Make a shallow clone of
git@github.com:virajp/repo-config.git into a temporary directory. Read
sync.md in the clone, and do its steps in sequence. Ask me before each choice
that sync.md does not decide.
```

## How a repository syncs

1. `setup:configs` sources `.config/mise/tasks/_scripts/repo-sync` and calls
   `repo_sync`.
2. `repo_sync` reads `tools` from the repository's `.config/repo-config.toml`.
3. It clones the default branch of this repository once (shallow, over SSH).
4. It adds the `depends` of each tool. It refuses a `root_only` tool in a git
   submodule.
5. It copies each file of each tool.

These rules apply to each file:

- A file is only replaced when its content changes. Use `--dry-run` to see the
  changes without writing them.
- A seed file is written only when it is missing. After that, the repository
  owns it.
- For a tool in `core_tools`, `conf.d/<tool>/mise.dev.toml` is written as
  `mise.toml`. Thus mise loads it in every environment.
- The sync never deletes a file.

A copy of `repo-sync` from before the `tools/` layout cannot sync. Copy
`tools/base/.config/mise/tasks/_scripts/repo-sync` and `helpers` into the
repository by hand one time. `sync.md` tells you how.

## Templates

A file whose name ends in `.tmpl` is a template. `repo_sync` renders it with
[gomplate](https://docs.gomplate.ca) and writes it without the `.tmpl` suffix.

- The delimiters are `<% %>`. mise uses `{{ }}` and bash uses `[[ ]]`, so
  both reach the output unchanged.
- The values come from the repository's `.config/repo-config.toml`.
- `.members` is the list of member names: the git submodules (else `MEMBERS`),
  then each `[member.<name>]` table whose folder exists in the repository.
- `.tools` is the list of the selected tools, dependencies included.
- `.extensions` is the sorted list of the VS Code extensions of those tools.

Example values file:

```toml
tools = ["base", "node", "fnox", "claude", "vscode"]
core_tools = ["node"]
merge_model_develop = "direct"
merge_model_main = "pr"
repo_name = "my-repo"
commit_scopes = ["backend", "web"]
repo_url = "https://github.com/me/my-repo"
worktree_files = [".config/mise.local.toml"]

[member.backend]
label = "Backend"
worktree_install = "pnpm install --frozen-lockfile"
```

## `.gitignore`

`repo_sync` writes the shared `.gitignore` first. Then it adds the lines of the
repository's `.config/repo-sync/gitignore`, if that file exists. A later line
wins, so a repository can negate a shared pattern with `!`.

## Change a shared file

1. Change the file in `tools/<tool>/`, and push to `main`.
2. In each repository, run `mise run setup:configs`, then commit the result.
