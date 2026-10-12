# repo-config

Shared configuration files, mise task scripts and templates for my
repositories. Each repository copies the files it needs with its own
`setup:configs` task.

## Layout

The shared files are in `files/`. This repository does not use them itself.
Each file in `files/` is at the path that it gets in a repository. For
example, `files/.config/mise/tasks/code/merge/develop` here is
`.config/mise/tasks/code/merge/develop` in every repository that uses it.

## Adopt in a repository

[`sync.md`](sync.md) tells an AI agent how to make a repository use these
files. Copy this prompt into an AI session in that repository:

```text
Make this repository get its shared files from repo-config.
Make a shallow clone of git@github.com:virajp/repo-config.git into a
temporary directory. Read sync.md in the clone, and do its steps in sequence.
Ask me before each choice that sync.md does not decide.
```

## How a repository syncs

1. `setup:configs` sources `.config/mise/tasks/_scripts/repo-sync`.
2. It calls `repo_sync -- <path> ...` with the paths that the repository
   wants. A path does not start with `files/`.
3. `repo_sync` clones the default branch of this repository once (shallow,
   over SSH), then copies each path from `files/`.

A file is only replaced when its content changes. Use `--dry-run` to see the
changes without writing them.

A copy of `repo-sync` from before the `files/` layout reads the paths from the
root of this repository, and it cannot find them. Copy
`files/.config/mise/tasks/_scripts/repo-sync` into the repository by hand one
time. Then `setup:configs` operates again.

## Templates

A file whose name ends in `.tmpl` is a template. `repo_sync` renders it with
[gomplate](https://docs.gomplate.ca) and writes it without the `.tmpl` suffix.

- The delimiters are `<% %>`. mise uses `{{ }}` and bash uses `[[ ]]`, so
  both reach the output unchanged.
- The values come from the repository's `.config/repo-config.toml`.
- `.members` is the list of member names: the git submodules (else `MEMBERS`),
  then each `[member.<name>]` table whose folder exists in the repository.

Example values file:

```toml
merge_model_develop = "direct"
merge_model_main = "pr"
repo_name = "my-repo"

[member.backend]
label = "Backend"
worktree_install = "pnpm install --frozen-lockfile"
```

## `.gitignore`

`repo_sync` writes the shared `.gitignore` first. Then it adds the lines of the
repository's `.config/repo-sync/gitignore`, if that file exists. A later line
wins, so a repository can negate a shared pattern with `!`.

## Change a shared file

1. Change the file in `files/`, and push to `main`.
2. In each repository, run `mise run setup:configs`, then commit the result.
