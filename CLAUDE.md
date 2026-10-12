# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

This repository is the source of shared files for other repositories. It has
no build, no tests, and no application code. Other repositories copy files from
it with their own `setup:configs` task. Read `README.md` for the sync contract.

## Layout rule

All shared files are in `files/`. This repository does not use them itself.
The files at the root (`README.md`, `CLAUDE.md`) are for this repository only.

Each file in `files/` is at the path that it gets in a consumer repository.
For example, `files/.config/mise/tasks/code/merge/develop` becomes
`.config/mise/tasks/code/merge/develop` in a consumer. Do not move a file in
`files/`. A move breaks every consumer that names the old path in its
`repo_sync -- <path> ...` call.

Many files refer to files that are not in this repository, for example
`.config/pre-commit-config.yaml`, `.config/gitleaks.toml`, `.config/grype.yaml`,
`.vscode/extensions.json`, and the `setup:ai` and `setup:external:update`
tasks. The consumer repository or a stack overlay supplies them. Do not run
the shared tasks in this repository.

## How the sync operates

- `files/.config/mise/tasks/_scripts/repo-sync` defines `repo_sync`. A
  consumer task sources its own copy of that file. It does a shallow clone of
  the default branch of this repository over SSH, then copies each path.
- A consumer gives paths without the `files/` prefix. The sync reads each
  path from `files/` in the clone.
- The sync uses the pushed `main` branch. Local changes have no effect on
  consumers until you push them.
- To test a change before you push it, set `REPO_SYNC_URL` to a local clone
  with a commit of the change. Use `repo_sync --dry-run` to show a diff.
- The sync writes the shared `.gitignore` first. Then it adds the lines of the
  consumer file `.config/repo-sync/gitignore`.

## Templates (`*.tmpl`)

- gomplate renders each `.tmpl` file. The output file has no `.tmpl` suffix.
- The delimiters are `<% %>`. Do not use `{{ }}` or `[[ ]]` for template
  logic. mise uses `{{ }}` and bash uses `[[ ]]`, and both must reach the
  output unchanged.
- The values come from the consumer file `.config/repo-config.toml`
  (`repo_name`, `merge_model_develop`, `merge_model_main`, `[member.<name>]`).
- `.members` is the list of member names: the git submodules (else the
  `MEMBERS` env value), then each `[member.<name>]` key with a folder in the
  consumer repository.

## Task scripts

Task scripts are in `files/.config/mise/tasks/`. mise uses the directory path
as the task name, for example `code/merge/develop` is `code:merge:develop`. A
`_default` file is the task for the directory name.

- `_scripts/` contains libraries, not tasks. Each task sources them through
  `${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/<name>`.
  - `helpers`: the print functions (`print_header`, `print_success`,
    `print_wait`, and others), `members`, and `shell_files_in_scope`. Use these
    functions for all task output.
  - `checks`: git predicates. They return an exit status only and print
    nothing.
  - `merge`: `merge_to_destination_branch`, which the two `code:merge:*` tasks
    use.
  - `placeholder`: for a task that no stack fills yet. A placeholder task
    contains the `#PLACEHOLDER` marker and always exits 0. A stack overlay
    replaces the file. Do not edit a placeholder in place.
  - `repo-sync`: the sync function.
- Write shell for bash 3.2 and bash 5.x. macOS has bash 3.2 as `/bin/bash`. Do
  not use namerefs, `${var,,}`, or associative arrays.
- Tasks use `#MISE` and `#USAGE` header comments for metadata and flags. A
  `#USAGE flag "--fix"` is available in the script as `usage_fix`.
- Under `set -e`, write a command substitution that can fail as
  `VAR=$(cmd) || VAR=''`. The existing scripts explain this in comments.
- Only the `init` task is in mise config (`_task/mise.dev.toml`). It makes the
  task scripts executable.

## Branch model in the merge tasks

The `merge` library enforces this model:

- Work starts on feature branches. `code:merge:develop <branch>` lands a
  branch on `develop`. `code:merge:main` lands `develop` on `main`.
- Nothing merges from `main`. Only `develop` merges into `main`.
- Merges use `--no-ff`. A hook pass over all files runs first as a safety net.
- `MERGE_MODEL_DEVELOP` and `MERGE_MODEL_MAIN` select `direct` (local merge and
  push) or `pr` (push and open a pull request with `gh` or `glab`).

## mise config

- `files/.config/mise.toml` holds the shared `[settings]`.
- `files/.config/mise/conf.d/<topic>/` holds one topic each. `mise.toml` files
  are for all environments. `mise.dev.toml` files are for development only.
- Directories with a `_` prefix (`_env`, `_formatter`, `_linter`, `_security`,
  `_task`) are for every repository. The other directories (`node`, `python`,
  `claude`, `pre-commit`) are for each stack or tool.

## Change a shared file

1. Change the file in `files/`.
2. Commit and push to `main`.
3. In each consumer repository, run `mise run setup:configs`, then commit the
   result.
