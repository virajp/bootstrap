# repo-config

Shared configuration files, mise task scripts and templates for my
repositories. Each repository copies the files it needs with its own
`setup:configs` task.

## Layout

Each file is at the path that it gets in a repository. For example,
`.config/mise/tasks/code/merge/develop` here is
`.config/mise/tasks/code/merge/develop` in every repository that uses it.

## How a repository syncs

1. `setup:configs` sources `.config/mise/tasks/_scripts/repo-sync`.
2. It calls `repo_sync -- <path> ...` with the paths that the repository
   wants.
3. `repo_sync` clones the default branch of this repository once (shallow,
   over SSH), then copies each path.

A file is only replaced when its content changes. Use `--dry-run` to see the
changes without writing them.

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

1. Change the file here, and push to `main`.
2. In each repository, run `mise run setup:configs`, then commit the result.
