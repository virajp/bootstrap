# Migrate a repository that has older copies

These instructions are for an AI agent that works in a consumer repository.
They are for a repository that has files from an older version of
`bootstrap`, or from `repo-config`, the old name of `bootstrap`. Read
[`index.md`](index.md) first. Its rules apply here too.

This file is temporary. Remove an item when all repositories have done it.

The items before M41 name the old values file `.config/bootstrap.toml`, its
overrides file and `.config/repo-generated.tsv`, with their old keys. Do such
an item on the old file, before M41: `bootstrap migrate` (M41) then moves each
key
into `.config/bootstrap/`.

## Find the items that apply

| The repository has                                                                                                                  | Do                                                                                                                                                         |
| ----------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.config/bootstrap.toml` with a `tools` list                                                                                        | The list stops `bootstrap`, and `bootstrap migrate` does not migrate it. Write one `[tools.<name>]` table for each tool by hand. Then do M17, M16 and M34. |
| `.config/repo-config.toml` with a `tools` list                                                                                      | M15. Its M15.4 includes the row above.                                                                                                                     |
| No values file, but older shared files or a mise config                                                                             | The steps of `index.md`. Its step 5 sends you here for M1 to M14. Write the new name directly.                                                             |
| `.config/mise/conf.d/_repo/`, or an `env` without `dev`                                                                             | M17, before the first sync.                                                                                                                                |
| `.config/repo-sync/gitignore`                                                                                                       | M18, before the next sync.                                                                                                                                 |
| Files from the stackgen `tool-config` skill or packs                                                                                | M19, before the next sync.                                                                                                                                 |
| `[tools.vscode]` without values, or no `vscode` in the root                                                                         | M20.                                                                                                                                                       |
| `@askviraj/linter` in a file of the repository                                                                                      | M21, before the next sync.                                                                                                                                 |
| A core branch that must allow direct commits                                                                                        | M22. Optional.                                                                                                                                             |
| A sync that adds `code:lint:repo-sync`                                                                                              | M23, after the sync.                                                                                                                                       |
| `[overrides.graphifyignore]`, `.graphifyignore` or `code:graph`                                                                     | M24, before the next sync.                                                                                                                                 |
| `.config/fnox.toml` with a top-level `[secrets]` or no profile                                                                      | M25, before the next sync.                                                                                                                                 |
| A git submodule without a `[member.<name>]` table                                                                                   | M26, before the next sync.                                                                                                                                 |
| A sync that adds `sync-engine/`                                                                                                     | M27, after the sync.                                                                                                                                       |
| A values file that selects one scanner, or a repository that read `required`                                                        | M28. Nothing to do.                                                                                                                                        |
| A sync that writes the first `.config/repo-generated.tsv`                                                                           | M29, after the sync.                                                                                                                                       |
| `.config/mise/tasks/setup/configs/sync`                                                                                             | M31, after the sync.                                                                                                                                       |
| `SYNC_ENGINE_` in a file or a script, or a folder `~/.cache/sync-engine/`                                                           | M32, after the sync.                                                                                                                                       |
| A file or a script that names a task of the table in M33                                                                            | M33, after the sync.                                                                                                                                       |
| A `setup:configs` that does not run `bootstrap sync`, or an old form that `bootstrap` reports                                       | M34, before the next sync.                                                                                                                                 |
| `version_prefix = "cli-v"` or `version_prefix="cli-v"` in a file of the repository                                                  | M35, before the next sync.                                                                                                                                 |
| A Mac with `[tools.gem]` or `[packs.security]` in `macos.toml`                                                                      | M36, before the next Mac sync.                                                                                                                             |
| `[tools._security_base]` in the values file                                                                                         | M37.                                                                                                                                                       |
| A values file without `format = 2`, or with `[packs.*]`, also `macos.toml`                                                          | M38, after the release `bootstrap@3.0.0`.                                                                                                                  |
| `[tools.go]`, or `.config/mise/tasks/code/format/20-gofmt`                                                                          | M39, after the sync.                                                                                                                                       |
| An `exec` block in `dprint.jsonc`, for example from `[tools.rust]` or `[tools.go]`                                                  | M40, after the sync.                                                                                                                                       |
| `.config/bootstrap.toml`, `.config/bootstrap.overrides.toml` or `.config/repo-generated.tsv`                                        | M41, after the release `bootstrap@4.0.0`, and after each other item.                                                                                       |
| A task `test`, `test:e2e` or `test:golden` of the repository, or a task without `NN-` in a task group                               | M42, with M41.                                                                                                                                             |
| `code:merge`, `setup:worktree` or `code:worktrees` in a file, a script or a hook of the repository                                  | M43, with M41.                                                                                                                                             |
| A vwf line in `.gitignore`, `vwf@virajp-plugins`, or a use of the `pnpx` install of `setup:ai`                                      | M44, before the next sync.                                                                                                                                 |
| Excludes in `.config/linter.yaml`, `dprint.jsonc`, `[tools.dprint] excludes` or a tool config, or an own gate                       | M45, after the sync.                                                                                                                                       |
| `UV_NO_CACHE` in `[overrides.mise.env]` of the values                                                                               | M46, before the next sync.                                                                                                                                 |
| `[tool.ruff]` or `[tool.ruff.format]` format settings in a `pyproject.toml`, or `.config/mise/tasks/code/format/40-ruff`            | M47, after the sync.                                                                                                                                       |
| An own task that sources a shell library of `.config/mise/tasks/_scripts/`, or an own part that calls a shell function of `helpers` | M48, after the sync.                                                                                                                                       |
| `main` and `develop` in `core_branches`, and `main` with a commit that `develop` does not have                                      | M49, before the next `git:merge:main`.                                                                                                                     |
| A script, a hook or a habit that runs `setup:configs` on a dirty tree, or commits its result by hand                                | M50, after the release with `default_branch`.                                                                                                              |
| An own task that is not a Python uv script, or own shell functions in `.config/mise/tasks/_scripts/local`                           | M51, after M48.                                                                                                                                            |
| `release_env` or a `[deploy.<env>]` table in `.config/bootstrap/repo.toml`                                                          | M52, before the next sync.                                                                                                                                 |
| An own task that bumps a version, commits a release or makes a release tag, for example `p:release`                                 | M53, after the release with the release tasks.                                                                                                             |
| `type = "local"` in a `[release.<project>.<kind>]` table of `.config/bootstrap/repo.toml`                                           | M54, before the next sync.                                                                                                                                 |
| `settings` or `lists` in `[tools.claude-cli]` of `macos.toml`, or a `~/.claude/CLAUDE.md` that bootstrap wrote as a seed            | M55, before the next Mac sync.                                                                                                                             |
| `macos.toml` at the root of a clone, `config_path` in `[tools.mise]`, or `MACOS_SETUP_REPO_PATH` in a script                        | M56, before the next Mac sync.                                                                                                                             |

**WARNING:** Do not delete a file before you move its repository-owned
content. The sync never restores repository content.

The sync deletes only two kinds of files: a file with the bootstrap header
that `.config/repo-generated.tsv` lists and the sync does not write (M29), and
a file with the bootstrap header at a path where the sync needs a folder. Thus
you must delete each other obsolete file and rename each old task by hand. Do
the items that apply to the repository, in
this sequence. Show the user the list of changes before you make them.

## M1: Delete the old sync files

- Delete `.config/gist-sync/` (an old `.gitignore` with only a header).
- If the repository uses `gist_sync`, delete
  `.config/mise/tasks/_scripts/gist-sync`. Move the lines of
  `.config/gist-sync/gitignore` into `.gitignore` as M18 tells you.
- Delete the old `.config/mise/tasks/setup/configs`. `bootstrap migrate` (M34)
  writes the new `setup/configs/_default`.

## M2: Delete the old dependency tasks

**CAUTION:** Do this before the first `mise run setup`. The new `setup` runs
each task in `setup/deps/`. An old task there runs too, and it can fail.

- Delete `setup/deps/{all,install,upgrade,audit,cleanup,outdated,check,restart}`
  and `setup/pnpm/*`. The tool tasks `setup:deps:<tool>` replace them, with
  `--upgrade` for the upgrade.
- If a repository needs a dependency step that no tool ships, write it as
  `setup/deps/<name>`, with a `--upgrade` flag.
- Delete each `#PLACEHOLDER` stub in `setup/deps/` and `setup/external/`.

## M3: Rename the old tasks

| Old task                           | New task                   | Action                                                                            |
| ---------------------------------- | -------------------------- | --------------------------------------------------------------------------------- |
| `setup:all` (`setup/all`)          | `setup` (`setup/_default`) | Delete the old file. `mise` ships the new one. Members are called with `setup`.   |
| `code:all` (`code/all`)            | `code` (`code/_default`)   | Delete the old file. Keep `code:check` and `code:test` as repository-owned tasks. |
| `code:format` (`code/format`)      | `code/format/_default`     | Delete the old file. Move each extra formatter step to `code/format/<name>`.      |
| `code:lint` (`code/lint`)          | `code/lint/_default`       | Delete the old file. Move each extra linter step to `code/lint/<name>`.           |
| `setup:secrets` (fnox)             | `setup:secrets:fnox`       | Delete the old file. Select the `fnox` tool.                                      |
| `setup:secrets` (Doppler)          | `setup:secrets:doppler`    | Delete the old file. Select the `doppler` tool. It needs a `doppler.yaml`.        |
| `setup:secrets` (stub)             | none                       | Delete it.                                                                        |
| `setup:deps:check`                 | `setup:external:check`     | Delete it. The `docker` tool ships the new one.                                   |
| `setup:git-config`                 | `code:git-config`          | Delete it.                                                                        |
| top-level `count`, `leaks`, `lint` | `code:*`                   | Delete them.                                                                      |

Update each shell alias that calls an old name. Remove a repository-owned
`setup` alias: `mise` ships `setup = "mise run setup"`.

## M4: Delete root-only items from submodules

In a submodule, delete `setup/external/*`, `setup/vscode` and `.vscode/`.
Keep `setup:ai`: the `claude` tool ships it to a submodule too, and it skips
there.

## M5: Move the content of `.config/mise.toml`

The sync replaces `.config/mise.toml`. It keeps only the shared `[settings]`.

1. Move `REPO_NAME`, `MEMBERS` and `MERGE_MODEL*` into the values file:
   - `REPO_NAME` becomes `repo_name`.
   - `MERGE_MODEL_DEVELOP` becomes `merge_model` of
     `[tools.git.branch.develop]`, and `MERGE_MODEL_MAIN` becomes
     `merge_model` of `[tools.git.branch.main]`. The old single
     `MERGE_MODEL` goes to both tables. A `direct` value is the default, so
     it can stay out. Remove the old keys `merge_model_develop` and
     `merge_model_main`: nothing reads them.
   - `MEMBERS` is not necessary for submodules. For folders, write
     `[member.<name>]` tables.
2. Move the repository's own `[env]` values to the seed
   `.config/mise/conf.d/zz-repo/mise.<env>.toml` of each environment that needs
   them. Examples are `GITHUB_OWNER`, and the variables of the local services
   of a backend (`mise.dev.toml`).
3. Delete what a tool now supplies: `CLOUDSDK_PYTHON` (the `gcloud` tool) and
   `task.timings = true` (the shared `[settings]`). Delete `lockfile = true`.
4. Keep `.config/mise.test.toml` and `.config/miserc.toml`. The sync does not
   write them.

## M6: Remove the tools that the shared files now install

Each tool file in [`tools/`](tools/) lists the mise tools of that tool
(section "mise tools"). From the repository-owned mise files, remove each mise
tool that a selected tool or a tool of the `base` pack installs. The
"Migration" section of a tool file lists the old tools that the tool replaces,
for example taplo (`dprint`) and `npm:wrangler` (`cloudflare`).

Each mise tool must come from a bootstrap tool. If the repository needs a
mise tool that no bootstrap tool installs, for example
`npm:sort-package-json` or `rtk`, tell the user. Then add it to a bootstrap
tool, or make a new tool for it.

If the old mise tool has a pinned version that is different from the shared
one, tell the user. For example, the repository pins node `26`, and `node`
ships `lts` and `latest`.

## M7: Split an old `conf.d/_base`

A repository can have an old folder `.config/mise/conf.d/_base`, from before
the `tools/` layout. Before the first sync, move its content:

| Old content in `_base`                                           | New place                 |
| ---------------------------------------------------------------- | ------------------------- |
| node, pnpm                                                       | the `node` tool           |
| python, uv, `UV_*`                                               | the `python` tool         |
| pre-commit, `PRE_COMMIT_HOME`                                    | the `pre-commit` tool     |
| `MEMBERS`, `REPO_NAME`                                           | the values file           |
| Other repository variables                                       | `zz-repo/mise.<env>.toml` |
| `min_version`, `task.run_auto_install`, other repository content | step M5                   |

Then delete the old files `mise.toml`, `mise.dev.toml` and `mise.ci.toml` of
that folder. Also delete an old `conf.d/ai/`: the `claude` tool replaces it.
Move its repository-owned content first (step M5). Delete the old
`conf.d/_formatter`, `_linter` and `_security`.

## M8: Move the fnox config

See "Migration" in [`tools/fnox.md`](tools/fnox.md).

## M9: Clean the old markers

- Replace each `MARKED POSITION` block with its value, in a repository-owned
  file or in the values file.
- For the `// >>> vscode.d` blocks and `.config/vscode.d/`, see "Migration"
  in [`tools/vscode.md`](tools/vscode.md).

## M10: Change the old helper files

- The shared helpers are `_scripts/helpers`, `_scripts/checks` and
  `_scripts/merge`. Change each `source` of `_helpers`, `_checks` or `_merge`
  to the new name. Then delete the old files.
- If a task calls a function that `helpers` does not have, for example
  `print_ok`, put that function in `_scripts/local` and source it from the
  task.
- M48 removes the shell libraries later: the shared libraries are now the
  Python modules `helpers.py`, `checks.py` and `merge.py`. Do M48 for a task
  that sources one of them, and M51 for each other own task. M51 moves
  `_scripts/local` to `_scripts/local.py`.

## M11: The commit config

The template has 10 types: `feat`, `fix`, `perf`, `refactor`, `revert`,
`test`, `ops`, `docs`, `merge`, `wip`. Other types are not used. The old types
`spec`, `security` and `blueprint` are dropped. Tell the user. Put the scopes
of the old file in `commit_scopes` of `[tools.git]`.

## M12: The pre-commit hooks

Do this before the first sync. `.config/pre-commit-config.yaml` is a shared
file: the sync merges the hooks of the tools into it, and it keeps the other
hooks of the file.

1. Compare the hooks of the existing file with the hooks of the selected
   tools. Use `bootstrap migrate --dry-run`.
2. Keep each hook that no tool ships in the file. Prefix each `entry` that
   calls a tool with `mise x --`.
3. Remove each hook that a tool now ships, for example `gitleaks` and
   `conventional-commits`.

## M13: The forge files

See "Migration" in [`tools/github.md`](tools/github.md) or
[`tools/gitlab.md`](tools/gitlab.md).

## M14: The mise environments

1. Add `environments` to `.config/bootstrap.toml` (step 1 of `index.md`).
   Delete an old `core_tools` key: the `env` of each `[tools.*]` entry, else
   the `default_envs` of the tool, decides its environments now.
2. Move the repository variables of an old `_env/mise.toml` or
   `conf.d/repo/` to the `zz-repo/mise.<env>.toml` seeds (step M5). Move each
   tool of an old `conf.d/repo/` to a bootstrap tool (step M6). Then delete
   `conf.d/repo/`.
3. Move the seeds of an old `conf.d/_env/` folder to `conf.d/zz-repo/`. The sync
   does not remove them, because a seed has no header.

Delete the old `conf.d/*/mise.toml`, `_env/mise.toml` and
`_alias/mise.dev.toml` files, and the old `mise*.toml` files of
`conf.d/_task/`, that have the header of this repository
(`Synced from repo-config:` or `Synced from bootstrap:`). The sync does not
remove them, because no manifest lists them (M29). Delete each other
file of `conf.d/_task/` by hand. `conf.d/pre-commit/` is now the generated
folder of the `pre-commit` tool, and the sync writes its `mise.dev.toml`.

## M15: The rename to bootstrap

The repository `virajp/repo-config` is now `virajp/bootstrap`
(`git@github.com:virajp/bootstrap.git`).

| Item              | Old                                     | New                                   |
| ----------------- | --------------------------------------- | ------------------------------------- |
| Source repository | `git@github.com:virajp/repo-config.git` | `git@github.com:virajp/bootstrap.git` |
| Values file       | `.config/repo-config.toml`              | `.config/bootstrap.toml`              |
| Synced header     | `# Synced from repo-config: ...`        | `# Synced from bootstrap: ...`        |

The old names still work for a short time. Do not rely on these fallbacks. A
later version of `bootstrap` can remove them.

- GitHub redirects the old URL. Thus an old copy of `setup:configs` still
  clones the new repository. The sync code then comes from that clone.
- The sync reads `.config/repo-config.toml` when `.config/bootstrap.toml` is
  missing. It shows a notice: `rename it to .config/bootstrap.toml`.
- The sync accepts both headers. Thus it still finds the files that it wrote
  with the old header.

Do not change the tools or other values in the values file. This item changes
only the name of the file.

### M15.1: Find the repositories to migrate

1. Get the list of submodules, nested submodules included:

   ```bash
   git submodule foreach --quiet --recursive 'echo $displaypath'
   ```

2. For the root repository and each submodule, find whether
   `.config/repo-config.toml` exists.
3. Show the user the list of repositories that have the file. A repository
   without `.config/repo-config.toml` and without `.config/bootstrap.toml`
   does not use `bootstrap`. Do not migrate it.
4. Ask the user which repositories to migrate.

Do M15.2 to M15.7 in each selected repository. Start with the deepest
submodules. Do the root repository last. Run the commands from the root of
the repository that you migrate.

### M15.2: Make sure that the repository is ready

1. Run `git status`. If the tree has changes that are not committed, stop.
   Ask the user what to do with them.
2. In the root repository, run `git submodule status`. A `+` before a
   submodule shows that its checkout is not at the commit that the root
   records. Tell the user, and ask whether to continue.
3. If the repository already has `.config/bootstrap.toml`, tell the user.
   Then go to the next repository.

### M15.3: Rename the values file

```bash
git mv .config/repo-config.toml .config/bootstrap.toml
```

Do not change the content of the file.

### M15.4: Run the sync

This is the first sync with `bootstrap`. Do not run it yourself. You can run
a dry run.

1. If mise says that the config files are not trusted, ask the user to run
   `mise trust`.
2. If the values file has a `tools` list, do the first row of "Find the items
   that apply" now: write the tables, then do M17, M16 and M34. If not, do
   M17, M16 and M34 if they apply.
3. Run `mise run setup:configs --dry-run`.
4. Examine the diff. Each synced file changes its header. Other changes come
   from the commits in `bootstrap` since the last sync of this repository.
   Show the user the list of files with changes that are not only the header.
5. If the dry run replaces repository-owned content, stop. Tell the user.
   "Rules" in `index.md` lists the places for repository-owned content.
6. Ask the user to run `mise run setup:configs`. Wait until the user tells
   you the result.
7. Run `mise run init`. This makes all the task scripts executable.

If `bootstrap` fails, read "The exit codes" of `index.md`.

### M15.5: Find the files with the old header

The sync replaces only the files of the tools that the values file lists. A
file with the old header that remains is from a tool or a path that
`bootstrap` no longer ships.

1. Find these files:

   ```bash
   git grep -l 'Synced from repo-config:'
   ```

2. If the list is empty, go to M15.6.
3. Show the list to the user. For each file, ask whether to delete it.

### M15.6: Find the other mentions of the old name

1. Find the other mentions:

   ```bash
   git grep -n 'repo-config'
   ```

2. Ignore the mentions in the synced files. The sync owns them.
3. Show the user the other mentions, for example in `README.md`,
   `CLAUDE.md`, the docs and `.config/repo-sync/`.
4. Ask the user which mentions to change. Do not change a mention in a
   changelog, in dated history or in a quote.

### M15.7: Make sure that the result is correct

1. Make sure that `.config/repo-config.toml` does not exist.
2. Run `git grep -l 'Synced from repo-config:'`. The list must be empty,
   except for the files that the user decided to keep in M15.5.
3. Run `mise run setup:configs --verbose`. It must not show the notice about
   the old values file. Each file must show `unchanged`, `kept seed`,
   `skip seed` or `skip env`.
4. Run `mise tasks`. Make sure that the synced tasks are in the list.
5. Run `mise run setup`. Make sure that it completes.

### M15.8: Commit

Commit these changes in each migrated repository:

- The rename of the values file.
- The synced files.
- The deletes from M15.5 and the changes from M15.6.

Use a commit message such as
`ops(repo-sync): migrate from repo-config to bootstrap`. Read
`.config/git-conventional-commits.yaml` for the types and scopes of the
repository. Use the git workflow of the repository, and ask the user before
each commit.

Commit in each submodule before you commit in the repository that contains
it. Ask the user before you push. Push each submodule before the repository
that contains it. If not, the containing repository points to a commit that
the remote does not have.

## M16: `android-sdk` and `java` moved out of `flutter`

Do this item only if the values file selects `flutter` or `kotlin`.

The `flutter` tool no longer defines `android-sdk` and `java`. The `android`
tool defines `android-sdk`, and the `java` tool defines `java`. The mise
`depends` of `flutter` adds both tools, in dev and ci. Do not add
`[tools.android]` or `[tools.java]` for `flutter`. See
[`tools/android.md`](tools/android.md) and [`tools/java.md`](tools/java.md).

1. The `java` version changes from `21` to `lts`. If the repository needs
   Java 21, for example for its Gradle or Android Gradle Plugin version, tell
   the user. Then ask how to keep Java 21.
2. The `kotlin` tool now depends on `java`. Do the items in "Adapt to the
   repository" of [`tools/kotlin.md`](tools/kotlin.md).
3. Run `mise run setup:configs --dry-run`. Examine the output:
   - With `flutter`: it must list the `.config/mise/conf.d/android/` and
     `.config/mise/conf.d/java/` files for each environment of `flutter`.
   - With `kotlin` and no `flutter`: it must list
     `.config/mise/conf.d/java/mise.dev.toml`. `java` is in dev only. If CI
     builds with Gradle, add `[tools.java]`, as `kotlin.md` says. Then the
     output must also list `.config/mise/conf.d/java/mise.ci.toml`.
4. Ask the user to run `mise run setup:configs`. Wait until the user tells
   you the result. If you do this item before the first sync, skip this step:
   the first sync is a later step of the row, of M15.4 or of `index.md`.

## M17: `conf.d/_repo/` is now `conf.d/zz-repo/`, and each tool is in `dev`

Do this item before the first sync.

1. Rename an existing `.config/mise/conf.d/_repo/` to
   `.config/mise/conf.d/zz-repo/`:

   ```bash
   git mv .config/mise/conf.d/_repo .config/mise/conf.d/zz-repo
   ```

   mise loads the folders of `conf.d/` in name order, and a later folder
   wins. `_repo` sorts before the tool folders, so a tool overrides the
   repository config. `zz-repo` sorts last, so the repository config wins. If
   you do not rename the folder, the sync writes new seeds in `zz-repo/`, and
   the old `_repo/` stays.
2. Make sure that each `env` of a `[tools.*]` or a `[packs.*]` entry includes
   `dev`, for example `env = ["dev", "ci"]`. An `env` without `dev`, for
   example `env = ["ci"]`, stops the sync with code 2. If the user needs a
   tool in `ci` only, tell the user that `bootstrap` does not support it.

## M18: `.config/repo-sync/gitignore` is now part of `.gitignore`

Do this item before the next sync. The old file stops `bootstrap` with code 2.
`.gitignore` is a shared file: the sync adds the lines of `bootstrap`, and it
keeps the other lines of the file.

1. Read `.config/repo-sync/gitignore`. Show the user its lines.
2. For each line, find the type. Show the user the result:
   - A blank line or a comment line: drop it.
   - A line that the repository needs to ignore: add it to `.gitignore`.
   - A line that cancels a shared pattern with `!`, for example
     `!.env.example`: add it to `.gitignore`. A later line wins.
3. If the old file only cancels a shared line, for example `!*.log` for the
   shared `*.log`, ask the user whether to remove the shared line instead:

   ```toml
   [tools.git.files.".gitignore"]
   "*.log" = { "$mode" = "remove" }
   ```

4. Delete the old file:

   ```bash
   git rm .config/repo-sync/gitignore
   ```

5. Run `bootstrap sync --dry-run`, and examine the diff of `.gitignore`.

An old `[overrides.gitignore]` table needs no manual step: `bootstrap migrate`
moves it (M34).

## M19: The stackgen files that bootstrap now ships

The `tool-config` skill and the packs of stackgen wrote some files that
bootstrap now ships. Do this item before the next sync. The old files do not
have the bootstrap header, so the sync does not remove them.

1. Find the old files and show the user the list:

   | Old file                                      | Do                                                                                                                                                                           |
   | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
   | `.config/mise/conf.d/ktlint/mise.toml`        | Delete it. The `kotlin` tool installs ktlint.                                                                                                                                |
   | `.config/mise/conf.d/detekt/mise.toml`        | Delete it. The `kotlin` tool installs detekt.                                                                                                                                |
   | `.config/mise/conf.d/swiftlint/mise.toml`     | Delete it. The `swift` tool installs SwiftLint.                                                                                                                              |
   | `.config/mise/conf.d/swiftui/mise.toml`       | Move each value of its `[env]` that the user wants to keep to the `[env]` of `conf.d/zz-repo/mise.dev.toml`. Then delete it. `_scripts/xcode` finds a value that is not set. |
   | `.config/mise/tasks/code/lint/ktlint`         | Delete it. `code:format:ktlint` reports the same problems.                                                                                                                   |
   | `.config/mise/tasks/code/lint/swift-format`   | Delete it. `code:format:swift-format` reports the same problems.                                                                                                             |
   | `.config/mise/tasks/code/check/uv`            | Delete it. `code:lint:uv` of the `python` tool does the same check. If `code/check/all` runs only this task, delete `code/check/all` too.                                    |
   | `.config/mise/tasks/code/count`               | Delete it, unless the user wants to keep it. bootstrap does not ship it.                                                                                                     |
   | A ruff or pyrefly pin in `pyproject.toml`     | Tell the user. The gates of the `python` tool use the mise version.                                                                                                          |
   | `.graphifyignore`                             | Keep it. bootstrap does not ship it, and the sync does not change it.                                                                                                        |
   | `.gitattributes` with lines of the repository | Tell the user. The sync replaces the file, and it has no overrides.                                                                                                          |

2. Keep these files. The sync writes the same path, but only when it is
   missing, or it does not write the path:
   - `.config/swiftlint.yml`, `.config/swift-format.json`,
     `.config/ktlint.editorconfig` and `.config/detekt.yml`: they are seeds.
   - `.config/miserc.toml` and `.config/claude-status.json`: bootstrap does
     not ship them.

3. Run `mise run setup:configs --dry-run`, and examine the diff. The sync
   replaces the old copies of `code/lint/ruff`, `code/format/ruff`,
   `code/lint/swiftlint`, `code/format/swift-format`, `code/format/ktlint`,
   `code/lint/detekt` and `_scripts/xcode`.

## M20: The `base` pack now adds `vscode`

The sync adds `vscode` to each root repository, and it skips the tool in a
git submodule.

1. If the values file has a `[tools.vscode]` table without values, remove the
   table. The entry does nothing in a root repository, and a submodule refuses
   it with code 2.
2. If a root repository did not select `vscode` before, do "Adapt to the
   repository" of [`tools/vscode.md`](tools/vscode.md) before the next sync.
   The sync replaces `.vscode/settings.json`.

## M21: The linter package is now `@virajp.dev/linter`

The `linter` tool installs `npm:@virajp.dev/linter`, not
`npm:@askviraj/linter`. `setup:linter` writes the linter config files only
when they are missing. Thus it does not change the old files.

1. Find the old name:

   ```bash
   git grep -n '@askviraj/linter'
   ```

2. Ignore the mentions in the synced files. The sync owns them.
3. In `eslint.config.mjs` and `.config/linter.yaml`, replace
   `@askviraj/linter` with `@virajp.dev/linter`.
4. Remove `npm:@askviraj/linter` from the repository-owned mise files.
5. Run `mise run setup:configs`, then `mise run code:lint`. The lint must
   complete without an error about a missing module.

## M22: Allow direct commits on a core branch (optional)

The `no-commit-to-branch` hook protects only the core branches (`main`,
`develop`) that exist and do not allow direct commits. Before M41 the key was
`direct_commit = true` of `[tools.git.branch.main]` and
`[tools.git.branch.develop]`. With `.config/bootstrap/`, it is
`landing = "direct-commit"` of `[branches.<branch>]` in `repo.toml`.

1. If the repository does not want direct commits on a core branch, do
   nothing. At the next sync, a repository with one core branch gets only
   `--branch <its branch>` in the hook.
2. To allow direct commits on a core branch, run
   `bootstrap config set repo.branches.<branch>.landing direct-commit`. With
   `main` and `develop`, `main` can allow them only when `develop` allows them
   too. See "Step 2: Select the branch model" of `index.md`.
3. Run `mise run setup:configs`.

## M23: `code:lint:repo-sync` validates the values file at each commit

The `repo-sync` tool now ships the gate `code:lint:repo-sync`. The `lint` hook
runs it through `code:lint` at each commit that changes
`.config/bootstrap.toml`. It runs `repo_sync --check`, which examines the
rules of the values file that need no `bootstrap` clone and shows each error
at once. A values file that the sync accepts also passes the check, because
the check uses the same rules.

1. After the sync, run `mise run code:lint` once.
2. If the gate reports an error, correct the values file. See "Step 6: Write
   the other values and sync again" of `index.md`. The gate is now
   `code:lint:60-bootstrap` (M34).
3. A repository with a valid values file does nothing.

## M24: The `claude` tool no longer ships graphify and mempalace

The `claude` tool now ships only `setup:ai`. It no longer ships
`.graphifyignore`, `code:graph`, the `graphify-refresh` hook and the mise tools
`pipx:graphifyy` and `pipx:mempalace`. `[overrides.graphifyignore]` of the
values file stops the sync with code 2. The shared `.gitignore` still ignores
`graphify-out/` and `/entities.json`.

1. Remove the `[overrides.graphifyignore]` table from `.config/bootstrap.toml`.
2. Run `mise run setup:configs`. The sync removes
   `.config/mise/conf.d/claude/mise.<env>.toml` when
   `.config/repo-generated.tsv` lists it. Else delete it by hand when it starts
   with the bootstrap header (M29). The pre-commit config loses the
   `graphify-refresh` hook.
3. Delete `.graphifyignore` and `.config/mise/tasks/code/graph` when they
   start with the bootstrap header. A file without the header belongs to the
   repository: keep it.
4. Remove the `post-commit` and `post-merge` hooks of pre-commit:

   ```bash
   mise x -- pre-commit uninstall --hook-type post-commit --hook-type post-merge
   ```

## M25: A repository does not use the fnox `default` profile

A repository keeps its fnox secrets only in named profiles,
`[profiles.<name>.secrets]` of `.config/fnox.toml`. The top-level `[secrets]`
table is the fnox `default` profile, and a repository does not use it.
`default` is only for the user-level fnox settings. A repository with one
profile uses its `repo_name` as the profile name. The sync and
`repo_sync --check` stop with code 2 when `.config/fnox.toml` has a top-level
`[secrets]` table (also an empty one), a `[profiles.default]` table, or no
named profile. `setup:secrets:fnox` checks only the named profiles.

The old seed ends with an empty `[secrets]` table, so each repository with
`.config/fnox.toml` must do this item.

1. In `.config/fnox.toml`, find each secret of the top-level `[secrets]`
   table.
2. Write a `[profiles.<repo_name>.secrets]` table, with the `repo_name` of
   `.config/bootstrap.toml`. Move each secret into it. A repository with more
   sets of secrets writes one `[profiles.<name>.secrets]` table for each set.
3. Delete the top-level `[secrets]` table. Rename a `[profiles.default]`
   table.
4. Replace the comment above the table with the comment of the new seed
   `repo/tools/fnox/files/.config/fnox.toml.tmpl`.
5. Add `-P <profile>` to each `fnox` call of the repository, for example
   `mise x -- fnox -P <repo_name> exec -- <command>`. Without `-P`, fnox uses
   the `default` profile, and the command gets no secret.
6. The keychain values do not move: a profile secret uses the same provider
   and the same key. Run `mise run setup:secrets:fnox` to make sure that each
   profile has its secrets.

## M26: A submodule without a table gets its `repo_name` as its name

A member name is now a flag (`--<name>`) and a part of a shell variable.
Thus a git submodule without a `[member.<name>]` table no longer gets its path
as its name. It gets the `repo_name` of its own `.config/bootstrap.toml`. A
submodule without that value gets its path with each `/` changed to `_`. A
name has only letters, digits, `-` and `_`, and two members never have the
same name. If not, the sync and `repo_sync --check` stop with code 2.

The name is in the flags of `setup`, `setup:configs`, `code:format` and
`code:lint`, and in the default commit scopes. Thus a new name changes these
flags and scopes.

1. Get the submodules, nested submodules included:

   ```bash
   git submodule foreach --quiet --recursive 'echo $displaypath'
   ```

2. Find each submodule that has no `[member.<name>]` table, with its path as
   `folder` or as the table name.
3. If the repository has no such submodule, do nothing more.
4. For each such submodule, show the user the old name (the path) and the new
   name (its `repo_name`, else the path with `_`). Ask whether to keep the
   new name, or to write a `[member.<name>]` table with `folder = "<path>"`
   and another name.
5. Run `bootstrap check`. It must pass.
6. Run `mise run setup:configs --dry-run`, and examine the new flags.

## M27: The cache moved to `~/.cache/sync-engine/`

The sync code moved to `sync-engine/` of `bootstrap`, and the cache of the
sync moved from `~/.cache/bootstrap/` to `~/.cache/sync-engine/`. The sync now
clones `https://github.com/virajp/bootstrap.git`, so it needs no SSH key. The
synced `.config/mise/tasks/_scripts/repo-sync` is now only a forwarding file.

The first sync after the update runs the old `setup:configs` one time. It
writes the new task. The second run uses the new task. M32 moves the cache
back to `~/.cache/bootstrap/`, so do M32 after this item.

1. Run `mise run setup:configs` two times. Commit the result of the first run.
   Without a new commit in `bootstrap`, the second run changes no file of the
   repository.
2. Before the second run, `code:lint:repo-sync` prints nothing and does not
   validate, because the new cache has no clone yet.
3. If the second run stops with code 2 and names the cache folder, that folder
   holds another repository or a `repo/` that is not a clone. Move it, or set
   `BOOTSTRAP_CACHE` to another folder. The sync did not delete the other
   repository or the `repo/`.

## M28: The merged packs model, and `required` is gone

Each key of a `[pack.<name>]` table of `repo/packs.toml` is now a member, and
the key `required` is gone. `gitleaks`, `grype` and `osv-scanner` now have
`_security_base` in their `depends`, so a single scanner now brings
`_security_base`.

Values files do not change. `[packs.<name>]` and `[tools.<name>]` have the same
form. The consumer does nothing.

## M30: The `node` tool has no package manager

The `node` tool no longer installs pnpm and no longer has the
`pnpm-workspace.yaml` seed or the alias `npx`. The new `pnpm` and `aube` tools
have them. A repository with a `package.json` selects one of them. Without a
selection, `setup:deps:node` stops the setup. A repository without a
`package.json` does nothing. A repository with only the `bun` tool does nothing.

1. If the repository stays on pnpm, add `[tools.pnpm]` to
   `.config/bootstrap.toml`. If it has `[tools.node]` with an `env` key, give
   `[tools.pnpm]` the same `env`. The existing `pnpm-workspace.yaml` stays.
2. If the repository moves to aube, add `[tools.aube]`. Then do "Adapt to the
   repository" of `tools/aube.md`. Rename `pnpm-workspace.yaml` to
   `aube-workspace.yaml` before the sync.
3. Run `mise run setup:configs`, then `mise run setup`.

No user-level config is necessary. The Mac tools no longer install pnpm, write
`Library/Preferences/pnpm/config.yaml` or set `PNPM_HOME`.

## M29: The sync writes `.config/repo-generated.tsv`

The sync now writes a manifest, `.config/repo-generated.tsv`, with one
`<tool><TAB><path>` row for each generated file. From then on, after a sync
without a failure, the sync removes each listed file with the bootstrap header
that it does not write again, then the empty parent folders. A file without the
header stays.

The first sync has no manifest, so the stale rule removes nothing. A file that
no manifest lists stays also after that sync. Thus delete the old leftovers by
hand, as before.

1. Run `mise run setup:configs`. Commit `.config/repo-generated.tsv` with the
   other files. Each git submodule has its own manifest: commit it in the
   submodule.
2. Delete each of these files that starts with the bootstrap header and that the
   manifest does not list. A file without the header belongs to the repository:
   keep it.
   - A `.config/mise/conf.d/*/mise*.toml`.
   - A generated `[files]` target whose `requires` or `requires_found` condition
     fails.
3. If the sync returns code 2 and names `.config/repo-generated.tsv`, that path
   is a symlink, is below a symlinked folder or is not a regular file. The sync
   wrote the other files, and it skipped the stale rule, the manifest and
   `bootstrap_commit`. Replace the path with a regular file, or delete it. Then
   run the sync again.

## M31: `setup:configs` runs the sync engine

`setup:configs` now runs the Go sync engine, `bootstrap sync`, of the pin
`sync-engine/VERSION` of the clone. The task `setup:configs:sync`
(`.config/mise/tasks/setup/configs/sync`) is gone, and `setup:configs` no
longer copies a local sync task from the clone.

The first sync after the update runs the old `setup:configs` one time. Its old
`setup/configs/sync` loads a shim from the clone, which runs the engine. That
run writes the new `_default`.

1. Run `mise run setup:configs`. Commit the result.
2. If `.config/repo-generated.tsv` listed
   `.config/mise/tasks/setup/configs/sync`
   before the sync, the stale rule removed it. Else delete
   `.config/mise/tasks/setup/configs/sync` when it starts with the bootstrap
   header. A file without the header belongs to the repository: keep it.
3. Do the same in each git submodule. `setup:configs --all` skips a member
   without `.config/bootstrap/` with a warning: adopt `bootstrap` in it
   first.

## M32: The engine is the command `bootstrap`

The Go engine is now the command `bootstrap`. `bootstrap sync` replaces
`sync-engine repo`, `bootstrap check` replaces `sync-engine repo --check`, and
`bootstrap mac sync` replaces `sync-engine macos`. Each variable
`SYNC_ENGINE_<name>` is now `BOOTSTRAP_<name>`, for example `BOOTSTRAP_BIN`,
`BOOTSTRAP_VERBOSE` and `BOOTSTRAP_CACHE`. The cache moved from
`~/.cache/sync-engine/` to `~/.cache/bootstrap/`. The sync no longer removes
the cache of the old stubs (`~/.cache/bootstrap.lock`).

1. Run `mise run setup:configs`. Commit the result.
2. In the files of the repository, replace each `SYNC_ENGINE_<name>` with
   `BOOTSTRAP_<name>`. Show the user each change before you write it.
3. Tell the user to replace `SYNC_ENGINE_<name>` in their shell config, if it
   sets one.
4. Tell the user to delete `~/.cache/sync-engine/` and
   `~/.cache/bootstrap.lock`, if they exist. The sync does not delete them.

## M33: The tasks of a task group have a number prefix

Each task that an orchestrator runs for each tool (`code:format`,
`code:lint`, `code:sec`, `setup:deps` and `setup:secrets`) has a number
prefix now. The prefix gives the run order. The sync writes the new files and
removes the old copies, because `.config/repo-generated.tsv` lists them.

| Old name                   | New name                      |
| -------------------------- | ----------------------------- |
| `code:format:dart`         | `code:format:10-dart`         |
| `code:format:gofmt`        | `code:format:20-gofmt`        |
| `code:format:ktlint`       | `code:format:30-ktlint`       |
| `code:format:ruff`         | `code:format:40-ruff`         |
| `code:format:swift-format` | `code:format:50-swift-format` |
| `code:lint:android`        | `code:lint:10-android`        |
| `code:lint:dart`           | `code:lint:20-dart`           |
| `code:lint:detekt`         | `code:lint:30-detekt`         |
| `code:lint:go`             | `code:lint:40-go`             |
| `code:lint:pyrefly`        | `code:lint:50-pyrefly`        |
| `code:lint:repo-sync`      | `code:lint:60-repo-sync`      |
| `code:lint:ruff`           | `code:lint:70-ruff`           |
| `code:lint:swiftlint`      | `code:lint:80-swiftlint`      |
| `code:lint:uv`             | `code:lint:90-uv`             |
| `code:sec:gitleaks`        | `code:sec:10-gitleaks`        |
| `code:sec:grype`           | `code:sec:20-grype`           |
| `code:sec:osv-scanner`     | `code:sec:30-osv-scanner`     |
| `setup:deps:android`       | `setup:deps:10-android`       |
| `setup:deps:flutter`       | `setup:deps:20-flutter`       |
| `setup:deps:kotlin`        | `setup:deps:30-kotlin`        |
| `setup:deps:node`          | `setup:deps:40-node`          |
| `setup:deps:python`        | `setup:deps:50-python`        |
| `setup:deps:swift`         | `setup:deps:60-swift`         |
| `setup:secrets:doppler`    | `setup:secrets:10-doppler`    |
| `setup:secrets:fnox`       | `setup:secrets:20-fnox`       |

M47 removes `code:format:40-ruff` later: replace an old `code:format:ruff`
with `code:format:05-dprint`.

1. Run `mise run setup:configs`. Commit the result.
2. In the files of the repository, replace each old name with its new name,
   for example in a CI workflow, a script or a document. Show the user each
   change before you write it.
3. A task of the repository in one of these groups needs a number prefix too:
   see M42.

## M34: The CLI `bootstrap` replaces the `repo-sync` tool

The `repo-sync` tool is gone. The `mise` tool now ships the entry:
`setup:configs` runs `bootstrap sync`, and the gate `code:lint:60-bootstrap`
replaces `code:lint:60-repo-sync`. mise installs the CLI as
`github:virajp/bootstrap`. Only `bootstrap migrate` migrates an old form.
`bootstrap sync` and `bootstrap check` stop with code 2 and the message
`run bootstrap migrate`.

The code of the CLI moved from `sync-engine/` to `cli/` of `bootstrap`, and the
release tags changed (see M35). The old entry task `setup:configs` reads
`sync-engine/VERSION` of the clone, and installs the release `vX.Y.Z`. Thus it
stops with an error after this release. Do the steps of this item: the old
entry task cannot do them.

`bootstrap migrate` moves these old forms:

| Old form                            | New place                                                               |
| ----------------------------------- | ----------------------------------------------------------------------- |
| `commit_scopes`                     | `commit_scopes` of `[tools.git]`                                        |
| `[branch.<name>]`                   | `[tools.git.branch.<name>]`                                             |
| `worktree_files`                    | `worktree_files` of `[tools.mise]`                                      |
| `[claude.plugins.<scope>]`          | `[tools.claude.plugins.<scope>]`                                        |
| `extensions`                        | The `recommendations` of `.vscode/extensions.json`                      |
| `add` of `[overrides.gitignore]`    | The lines of `.gitignore`                                               |
| `remove` of `[overrides.gitignore]` | `"<line>" = { "$mode" = "remove" }` of `[tools.git.files.".gitignore"]` |
| `.config/repo-sync/pre-commit.yaml` | The `repos` of `.config/pre-commit-config.yaml`. The old file goes.     |
| `[tools.repo-sync]`                 | Nothing. The `mise` tool has the entry.                                 |

1. Install the CLI:

   ```bash
   mise use -g github:virajp/bootstrap
   ```

2. If `.config/mise/tasks/setup/configs/sync` exists, delete it. This old task
   sources `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync` from the
   clone, and that file is gone. A file without the bootstrap header belongs
   to the repository: ask the user first.
3. Commit or discard the changes of the work tree. `bootstrap migrate` writes no
   file that has changes that are not committed.
4. Run `bootstrap migrate --dry-run`. Show the user the change list.
5. Ask the user whether to apply the change list. If the user agrees, run
   `bootstrap migrate --yes`. It migrates each old form, then syncs. Give the
   answer of each new question with `--answer <tool>.<key>=<value>`.
6. If `.config/mise/tasks/_scripts/repo-sync`,
   `.config/mise/tasks/code/lint/60-repo-sync` or
   `.config/mise/conf.d/repo-sync/` stays, delete it when it starts with the
   bootstrap header.
7. Replace each `code:lint:60-repo-sync` in the files of the repository with
   `code:lint:60-bootstrap`. Show the user each change before you write it.
8. Commit the result. Do the same in each git submodule.

## M35: The release tags are `bootstrap@X.Y.Z`

The release tags of the CLI changed from `cli-vX.Y.Z` to `bootstrap@X.Y.Z`.
mise removes a tag prefix that is the name of the repository, so a mise command
needs no `version_prefix` now:

```bash
mise x github:virajp/bootstrap@latest -- bootstrap <command>
```

From 2.0.2, `cli/VERSION` of the clone is 2.0.2 or newer. The old entry task
`setup:configs` has `version_prefix="cli-v"`, so it installs 2.0.1, and that
binary stops with code 2. Do the steps of this item: the old entry task cannot
do them.

A Mac that has an older Mac sync has `version_prefix = "cli-v"` in
`~/.config/mise/conf.d/mise/mise.toml` too. mise uses that line in each
folder, so the command of step 3 does not find the release. The setting
`minimum_release_age` of mise can also hide a new release, and then `@latest`
gets an old release.

1. Commit or discard the changes of the work tree.
2. If `~/.config/mise/conf.d/mise/mise.toml` has `version_prefix = "cli-v"`,
   show the user the line, and remove it if the user agrees. The next Mac sync
   writes the file again without the line (`docs/setup/index.md`).
3. Run the new CLI once. If mise hides the release, use the version of
   `cli/VERSION` of the clone in place of `latest`:

   ```bash
   mise x github:virajp/bootstrap@latest -- bootstrap sync --dry-run
   ```

4. Show the user the change list. If the user agrees, run the same command
   without `--dry-run`. The sync writes the new entry task and
   the new mise config, with no `version_prefix`.
5. If `version_prefix = "cli-v"` stays in a file that does not have the
   bootstrap header, remove that line. Show the user each change before you
   write it.
6. Commit the result. Do the same in each git submodule.

## M36: Each Mac package is a tool of its own

The Mac tools `cli`, `network`, `git`, `fnox` and `python` installed
more than one package. Now each package is a tool of its own, in the pack
`base` with `default = true`:

| Old tool  | New tools                                                           |
| --------- | ------------------------------------------------------------------- |
| `cli`     | `bat`, `coreutils`, `curl`, `eza`, `fzf`, `jq`, `ripgrep`, `zoxide` |
| `network` | `doggo`, `gping`, `httping`, `ipcalc`, `prettyping`                 |
| `git`     | `lazygit`, `worktrunk`                                              |
| `fnox`    | `age`                                                               |
| `python`  | `uv`                                                                |

`ca-certificates` moved from `cli` to `mise`. The tool `fonts` keeps its five
fonts. `ruby` moved from `mise` to the new tool `ruby` in `base`, and the tool
`gem` is now part of `ruby`. A Mac that selects only `base` gets the same
packages after the sync.

The member `1password` of the pack `security` is now `{}`. Thus
`[packs.security]` no longer adds the 1Password app.

1. If `macos.toml` has `[tools.gem]`, remove that table. Else the Mac sync
   stops with code 3. `ruby` in `base` writes `~/.gemrc`.
2. If `macos.toml` has `[packs.security]` and the owner wants the 1Password
   app, add `[tools.1password]`, with its answer `ssh_vault`. Ask the owner
   first.
3. Run `mise run updateall:configs`.

## M37: `_security_base` is not selectable

The tool `_security_base` now has `selectable = false`. `gitleaks`, `grype` or
`osv-scanner` selects it, because each of them has it in its `depends`.
`bootstrap migrate` and `bootstrap add` refuse it by name with code 2. `sync`
accepts a values file that names it, and `check` shows this warning:
`_security_base is not selectable - remove [tools._security_base] from <file>; a
tool that needs it selects it`.

1. If the values file has a `[tools._security_base]` table, show the user the
   table, and remove it if the user agrees. If the table has an `env` that the
   scanners do not have, tell the user first.
2. Run `mise run setup:configs`. The sync gives the same result.
3. Do the same in each git submodule.

A Mac needs no change. The Mac git tool now needs one email only, so an empty
`email_github` or `email_gitlab` in `macos.toml` is valid.

## M38: The values file holds tools, and the grammar has a kind word

`bootstrap@3.0.0` changes the form of the values file and the grammar of the
CLI. An engine older than `3.0.0` stops with code 2 on the new packs files.

- The values file has the top-level key `format = 2`, and one
  `[tools.<name>]` table for each selected tool, also for each member of the
  `base` pack. It has no `[packs.<name>]` table: a pack is only a shortcut of
  `bootstrap add pack`. A file without `format = 2`, or with a `[packs.*]`
  table, is an old form: `sync` and `check` stop with code 2 and "run
  bootstrap migrate".
- The commands have a kind word: `bootstrap add tool <tool…>`,
  `bootstrap add pack <pack…> [--exclude <member>]`,
  `bootstrap remove tool <tool…>` and `bootstrap replace tool <old> <new>`.
  The same on the Mac: `bootstrap mac add tool <tool…>`. A command without the
  kind word stops with code 2.
- `bootstrap migrate` takes no tool names.
- `bootstrap detect` without a flag only lists. `bootstrap detect --add` adds
  each detected tool. `--no` is gone.
- The repo packs `security` and `container-build` are gone. `gitleaks`,
  `grype` and `osv-scanner` are required members of `base`, and so are
  `dprint` and `linter`. `vscode` is a member of `base` that a repository can
  remove.
- The Mac packs are now `base`, `claude`, `essentials`, `1password-apps`,
  `dev`, `terminal`, `mobile-dev`, `iwork`, `ms-office`, `disk` and `media`.
  See `macos/packs.toml`.
- The Mac tool `devtools` is gone. Its `PRE_COMMIT_HOME` is in the new Mac tool
  `pre-commit`, a required member of the Mac `base`. Its
  `DOTNET_CLI_TELEMETRY_OPTOUT` is gone: a later .NET tool brings it. Its
  `RUNTIME_ENV` is gone: each repository sets it, and no bootstrap tool sets
  it.

Do these steps in each repository, after the release `bootstrap@3.0.0`:

1. Upgrade the CLI: `mise use -g github:virajp/bootstrap`.
2. Run `bootstrap migrate --dry-run`. It writes a `[tools.<name>]` table for
   each
   required tool of the old pack `base` and for each tool of each `[packs.*]`
   table, with the `env` of its pack, removes the `[packs.*]` tables, and
   writes `format = 2`. It writes no table for `vscode`, because `vscode` is
   not a required member of `base`. Examine the change list.
3. If the user agrees, run `bootstrap migrate --yes`. The sync gives the same
   tools as before, but without `vscode`.
4. Add `vscode` again with `bootstrap add tool vscode`, in the correct
   repository:
   - In a monorepo or a polyrepo, add it in that repository.
   - In a repository with members (git submodules or siblings), add it in the
     root repository only. Do not add it in a member.
5. Change each script and each doc of the repository that runs
   `bootstrap add`, `bootstrap remove`, `bootstrap replace` without a kind
   word, `bootstrap migrate` with tool names, or `bootstrap detect --no`.
6. Do steps 2, 3 and 5 in each git submodule.

On the Mac, run `bootstrap mac migrate --dry-run`, then `bootstrap mac migrate
--yes`,
in the repository of `macos.toml`. It also removes `[tools.devtools]` and lists
it. Then run `mise run updateall:configs`.

## M39: The `go` tool formats with dprint, and lints with golangci-lint

The `go` tool changed in four ways:

- dprint formats the `.go` files. The `go` tool adds an `exec` command for
  `gofmt` to the shared `dprint.jsonc`. The task `code:format:20-gofmt`
  (`.config/mise/tasks/code/format/20-gofmt`) is gone.
- `code:lint:40-go` runs `golangci-lint` with its standard linters, not
  `go vet`. These linters find more problems than `go vet`.
- The new check `code:lint:45-gomod` fails when a `go.mod` or a `go.sum` is
  not tidy.
- The new task `setup:deps:25-go` downloads and verifies the modules, reports
  the outdated ones, and audits them with `govulncheck`.

The sync installs `golangci-lint` and `govulncheck`, and writes
`code:lint:40-go` and `setup:deps:25-go`, only when it finds a `go.mod`.

Do these steps in each repository that selects `go`:

1. Run `mise run setup:configs`, then `mise install`.
2. If `.config/repo-generated.tsv` listed
   `.config/mise/tasks/code/format/20-gofmt` before the sync, the sync removed
   it. Else delete it when it starts with the bootstrap header. A file without
   the header belongs to the repository: keep it.
3. Run `mise run code:format`. If dprint changes a `.go` file, commit the
   result.
4. Run `mise run code:lint`. If `code:lint:40-go` finds problems, show them to
   the user. Fix them, or let the user add a `.golangci.yml` to the module.
5. If `code:lint:45-gomod` fails, run `mise run code:lint:45-gomod --fix`.
6. Commit the result.

## M40: Each exec command of `dprint.jsonc` is its own item

Before this change, the sync wrote the list `exec.commands` of `dprint.jsonc`
only one time. A later tool did not add its command, and a removed tool left
its command. The rustfmt command of the `rust` tool also had `cacheKeyFiles`.
The exec plugin stops when a listed file does not exist, so dprint formatted
no `.rs` and no `.go` files in a repository without `rustfmt.toml`.

Now the sync finds each command by its `exts`, and the rustfmt command has no
`cacheKeyFiles`. The first sync after this change removes the old list, and
the second sync writes each command again. Commit the first sync before the
second one.

Do these steps in each repository with an `exec` block in `dprint.jsonc`:

1. Copy the list `exec.commands` of `dprint.jsonc`.
2. Run `mise run setup:configs`, then commit the result. The next sync does
   not write a file with changes that are not committed.
3. Run `mise run setup:configs` again.
4. Make sure that `exec.commands` has one command for each selected tool with
   a formatter through exec, for example `gofmt` for `go` and `rustfmt` for
   `rust`.
5. If the copy of step 1 had a command that the repository added, add it
   again. The sync keeps it from now on.
6. Run `mise run code:format`, then commit the result.

## M41: The values move into the folder `.config/bootstrap/`

`bootstrap@4.0.0` moves the values of a repository from one file into the
folder `.config/bootstrap/`. An engine older than `4.0.0` cannot read it, and
the tool scripts of `bootstrap` stop with code 2 on an older engine.

- `repo.toml` holds the facts of the repository: `repo_name`, `product_name`,
  `repo_url`, `environments`, `bootstrap_commit`, the new key `core_branches`,
  `[branches.<branch>]`, `[commits]`, `[worktree]` and `[members.<name>]`.
- `tools.toml` holds the `[tools.<name>]` tables and `[overrides.*]`.
- `releases.toml` is new and optional. `.config/releases.yaml` stays where it
  is: a later release moves it.
- `repo-generated.tsv` is the manifest, moved from `.config/`.
- Each TOML file has `format = 3`. An unknown key now stops each command with
  code 2.
- These keys move or change:

  | Old key                                                       | New key                                    |
  | ------------------------------------------------------------- | ------------------------------------------ |
  | `[tools.git].commit_scopes`                                   | `[commits].scopes` of `repo.toml`          |
  | `[tools.git.branch.<b>]`                                      | `[branches.<b>]` of `repo.toml`            |
  | `merge_model = "pr"`                                          | `landing = "pr"`                           |
  | `merge_model = "direct"` or none, with `direct_commit = true` | `landing = "direct-commit"`                |
  | Each other branch table                                       | `landing = "direct-merge"`, the default    |
  | `[tools.mise].worktree_files`                                 | `[worktree].files` of `repo.toml`          |
  | `[member.<name>]`                                             | `[members.<name>]`, with a required `type` |
  | `.config/bootstrap.overrides.toml`                            | `[overrides.tools.<name>]` of `tools.toml` |

- A key of the overrides file that moves to `repo.toml` goes into `repo.toml`
  directly, and wins over the value of the old file. `repo.toml` has no
  overrides.
- A member can now be a sibling repository: `type = "sibling"`, with `folder`
  (for example `../acme-api`) and `url`. Each submodule and sibling member gets
  `[home]` in its own `repo.toml`, the back-link that the sync of the member
  writes. No bootstrap task reads `.config/vwf.yaml` or
  `.config/vwf-membership.yaml` any more: `git:worktree:list` and the member
  flags read `[members.*]` and `[home]`.
- `code:format` and `code:lint` have no member flags (`--all`, `--<member>`)
  any more: each repository and each submodule runs its own.
- `bootstrap config` is now the reader and writer of the values:
  `bootstrap config get|set|unset|show`. The old `bootstrap config`, which
  printed the settings of the engine, is now `bootstrap settings`.
- Each other command stops with code 2 and "run bootstrap migrate" on the old
  file. With both the old file and the folder, each command stops with code 2.

Do these steps in each repository, after the release `bootstrap@4.0.0`, and
after each other item of this file that applies:

1. Upgrade the CLI: `mise use -g github:virajp/bootstrap`.
2. Make sure that `main` or `develop` exists, locally or on `origin`. Without
   both, `bootstrap migrate` stops with code 2. A `[tools.git.branch.<b>]` table
   of another branch, or of a branch that does not exist, also stops it: remove
   the table, or create the branch.
3. Run `bootstrap migrate --dry-run`. It lists the split of
   `.config/bootstrap.toml` into `.config/bootstrap/repo.toml` and
   `.config/bootstrap/tools.toml`, the move of the overrides file and of
   `.config/repo-generated.tsv`, the removal of the old files, and the files
   that the sync changes. Examine the change list.
4. If the user agrees, run `bootstrap migrate --yes`. The comments of a table
   that
   moves to `repo.toml`, for example `[tools.git.branch.main]`, are lost: tell
   the user, and write them again in `repo.toml` if the user wants them. A
   member table without `type` gets `submodule` when its folder is in
   `.gitmodules`, else `folder`.
5. For a product with sibling repositories, write a `[members.<name>]` table
   for each sibling in the base repository:
   `bootstrap config set repo.members.<name>
   '{"type":"sibling","folder":"../<name>","url":"<url>"}'`.
6. Change each script and each doc of the repository that reads
   `.config/bootstrap.toml`, `.config/bootstrap.overrides.toml` or
   `.config/repo-generated.tsv`, or a key that moved. A script reads a value
   with `bootstrap config get <key> --format json` and its `.value`. Change
   each call of `bootstrap config` that printed the settings to
   `bootstrap settings`.
7. Do steps 2 to 6 in each member, then run `mise run setup:configs --all` in
   the base repository. The sync of each submodule and sibling member writes its
   `[home]`. `setup:configs --all` skips a member that has not done this item,
   with the warning `member <name>: run bootstrap migrate there`.
8. Commit `.config/bootstrap/` and the removal of the old files.

## M42: The test ports, and a part needs a number prefix

`bootstrap@4.0.0` ships the tasks `test`, `test:e2e` and `test:golden` in the
`mise` tool: `.config/mise/tasks/test/_default`, `test/e2e/_default` and
`test/golden/_default`. Each runs the `NN-<name>` tasks of its group in number
order, for example `test:10-unit`, and gets the arguments unchanged. `test`
does not run `test:e2e` or `test:golden`. With no part, a task prints one line
and exits 0.

- A task of the repository at one of these paths stops `bootstrap sync` and
  `bootstrap check` with code 2, and the fix `run bootstrap migrate`: `test`,
  `test/_default`, `test/e2e`, `test/e2e/_default`, `test/golden` and
  `test/golden/_default`, below `.config/mise/tasks/`. A file with the
  bootstrap header is not a task of the repository.
- `bootstrap migrate` moves each of them to the part `50-local` of its group:
  `test/50-local`, `test/e2e/50-local` or `test/golden/50-local`.
  `mise run test` runs it as a part, with its depends. When the part
  `50-local` exists, `migrate` stops with code 2 and moves nothing.
- `bootstrap migrate` removes the old stackgen copies of the test tasks of the
  `android` and `swift` tools, and does not move them: the change
  `remove old copies of the <tool> test parts`. See
  [`tools/android.md`](tools/android.md#migration) and
  [`tools/swift.md`](tools/swift.md#migration).
- `code:format`, `code:lint`, `code:sec`, `setup:deps`, `setup:secrets` and the
  test tasks now run only an executable file named `NN-<name>`. A task of the
  repository in one of these folders without the prefix does not run any more.
  `bootstrap check` warns about it. A task that a `#MISE depends` of another
  file of the same folder names gets no warning.

Do these steps in each repository:

1. A `test` task in `.config/mise.toml` (`[tasks.test]`) or in another task
   folder: move it by hand into `.config/mise/tasks/test/50-local`, before
   `bootstrap migrate`.
2. Run `bootstrap migrate --dry-run`. The change list shows the change
   `move own test tasks into 50-local parts`, with each file and its target.
   If a part `50-local` exists, join the two tasks in it, then run
   `bootstrap migrate --dry-run` again.
3. Run `bootstrap migrate --yes`, the same run as M41.
4. Run `bootstrap check`. For each warning `has no NN- name`, rename the file
   to `NN-<name>`, for example `code/lint/mylint` to `code/lint/50-mylint`.
5. Commit the moved and renamed files.

## M43: The git tasks have `git:*` names

`bootstrap@4.0.0` renames the git tasks, with no alias for an old name:

| Old task                      | New task                                                                  |
| ----------------------------- | ------------------------------------------------------------------------- |
| `code:merge:develop <branch>` | `git:merge:develop [<branch>]`                                            |
| `code:merge:main`             | `git:merge:main [<branch>]`                                               |
| `code:merge <branch>`         | `git:merge:<core branch> <branch>`, for example `git:merge:main <branch>` |
| `setup:worktree`              | `git:worktree:setup`                                                      |
| `code:worktrees`              | `git:worktree:list`                                                       |

- The name of a merge task is its destination. The sync writes
  `git:merge:<branch>` for each branch of `core_branches`. The source is
  optional: `git:merge:main` takes `develop` when it is a core branch, else
  the current branch, and `git:merge:develop` takes the current branch.
  `git:merge:main` no longer needs `develop` as the current branch:
  `merge_from` of `[branches.main]` decides the source.
- The new tasks `git:worktree:create <branch> [--from <base>] [--no-setup]`
  and `git:worktree:delete <branch>` add and remove a worktree. See
  [`tools/git.md`](tools/git.md).
- The sync removes the old task files itself, through the manifest. The alias
  `worktrees` now runs `mise run git:worktree:list`.
- `.config/wt.toml` is a seed, so the sync does not write it again. Its
  `pre-start` step, and a `[members.<name>].worktree_install` of `repo.toml`,
  can run `mise run setup:worktree`. That stops `bootstrap sync` and
  `bootstrap check` with code 2 and the fix `run bootstrap migrate`.
  `bootstrap migrate` renames the command to `mise run git:worktree:setup`, and
  the rest of the file or value stays.

Do these steps in each repository:

1. Run `bootstrap migrate --dry-run`. The change list shows the change
   `rename setup:worktree` when `.config/wt.toml` or a member runs the old
   command, with each file and key.
2. Run `bootstrap migrate --yes`, the same run as M41.
3. Find each other use of an old name, in the scripts, the hooks, the CI
   files, the docs, `CLAUDE.md` and the shell aliases of the repository:

   ```bash
   git grep -n -e 'code:merge' -e 'setup:worktree' -e 'code:worktrees'
   ```

   Change each one to its new name by hand. A file with the bootstrap header
   is not one: the sync writes it.
4. Commit the changed files.

## M44: vwf comes from the `virajp-plugins` tool

The `git` tool no longer ships the vwf lines of `.gitignore`
(`docs/memory/handoff/`, `docs/memory/doctor/`, `docs/memory/runs/`,
`docs/scratchpad/` and `!docs/plans/**/*.log`). `setup:ai` of the `claude`
tool no longer runs `pnpx @virajp.dev/claude-plugins@latest --all` for
`vwf@virajp-plugins`, and it names no plugin. The new tool `virajp-plugins`
has both: it adds the vwf lines to `.gitignore` of the `git` tool, and the
entry `"vwf@virajp-plugins" = "virajp/claude-plugins"` to
`[tools.claude.plugins.user]`. It depends on `claude` and `git`, and only the
root repository can select it.

The sync removes an item of a shared file that no selected tool has now. So
without `virajp-plugins`, the next sync removes the vwf lines from
`.gitignore`.

`setup:ai` also changes:

- It updates the marketplaces and the plugins only with `--upgrade`:
  `mise run setup:ai --upgrade`. `mise run setup --upgrade` runs it with the
  flag.
- It no longer runs `claude plugin autoremove`.
- It keeps a record of the project plugins that it installed,
  `tools.claude.plugins.installed`, and it uninstalls a recorded plugin that
  leaves `[tools.claude.plugins.project]`. It never uninstalls a plugin at user
  scope, or a plugin that is not in the record.

Do these steps in a repository that uses vwf, before the next sync:

1. Run `bootstrap add tool virajp-plugins` in the root repository. A
   submodule member cannot select it.
2. If `[tools.claude.plugins.user]` or `[tools.claude.plugins.project]` of
   `.config/bootstrap/tools.toml` lists `vwf@virajp-plugins`, you can remove
   the entry: `virajp-plugins` contributes it.
3. Run `mise run setup:configs`, then `mise run setup:ai`.
4. Commit the changed files.

## M45: The gates examine the uncommitted files, and `code_files` holds the excludes

`code:format` and `code:lint` now examine only the uncommitted files of the
repository (staged, changed and new) when they get no files. `--all` examines
every file of the repository, and `mise run code --all` passes it to both.
`code:sec` does not change: it keeps its own file lists. Each gate gets its
files from `gate_files` of `helpers` and its `#FILES` header, and in a gate run
`gate_files` is the only place that removes a file. See
[`tools/mise.md`](tools/mise.md#the-file-sets-of-the-gates).

- The new key `code_files` of `[tools.mise]` holds two lists of patterns,
  `includes` and `excludes`. The sync renders them to
  `.config/bootstrap/config-files.toml`, and the excludes also into
  `dprint.jsonc`. The base excludes of the `dprint` tool moved there, and
  `[tools.dprint] excludes` now adds only to `dprint.jsonc`.
- A gate no longer reads the path excludes of its tool config: the ignores of
  `.config/linter.yaml`, the excludes of the ruff and pyrefly config, and the
  `excluded` paths of `.config/swiftlint.yml` do not act in a gate run. The
  own excludes of `dprint.jsonc` still act in dprint, but `--debug` does not
  name the files that they drop.
- ShellCheck is now the gate `code:lint:06-shellcheck` of the `linter` tool.
  It does not check a zsh file. `shell_files_in_scope` of `helpers` is gone.
- `code:format` and `code:lint` pass `--all` only to a gate that declares
  `#USAGE flag "--all"`. An own gate without it examines what it examined
  before.
- The sync removes the `associations` of the `shuck` block of `dprint.jsonc`.
  shuck formats only the shell files with an extension (`.sh`, `.bash`, `.zsh`
  and `.ksh`), so it no longer formats the task files or `gate_files.py`.

Do these steps in each repository, after the sync:

1. Move each path exclude of these places into `code_files.excludes` of
   `[tools.mise]` in `.config/bootstrap/tools.toml`, and remove it there:
   - the `ignores` of `.config/linter.yaml`;
   - `[tools.dprint] excludes` of `.config/bootstrap/tools.toml`;
   - the own `excludes` of `dprint.jsonc`, for example a vendored folder (a
     submodule needs none: a gate never examines a file of a submodule);
   - the path excludes of the ruff, pyrefly and SwiftLint configs.

   ```toml
   [tools.mise.code_files]
   excludes = [ "vendor/", "**/*.snap" ]
   ```

   A pattern has the gitignore glob form, without negation.
2. For each own `code:format:NN-<name>` and `code:lint:NN-<name>` gate of the
   repository, add `#USAGE flag "--all"`, a `#FILES` header with its file
   types, and the call of `gate_files`. Then give its tool only that list:

   ```bash
   #USAGE flag "--all" help="lint every file of the repository"
   #FILES include="*.proto"

   args=()
   [[ "${usage_all:-false}" == "true" ]] && args+=(--all)
   [[ "${usage_debug:-false}" == "true" ]] && args+=(--debug)
   [[ "${usage_fix:-false}" == "true" ]] && args+=(--fix)
   FILES=()
   [[ -n "${usage_files:-}" ]] && eval "FILES=(${usage_files})"
   list=$(gate_files "$0" ${args[@]+"${args[@]}"} ${FILES[@]+"${FILES[@]}"}) || exit $?
   trap 'rm -f "${list}"' EXIT
   [[ -s "${list}" ]] || exit 0
   ```

3. A script, a CI job or a doc that runs `code:format` or `code:lint` to
   examine the whole repository: add `--all`.
4. Run `mise run setup:configs`, then `mise run code --all --debug`. Each
   `drop` line names a removed file and its pattern.
5. Commit the changed files.

M48 removes the shell `helpers` later. After M48, write an own gate in
Python, as M48 shows.

## M46: The `python` tool no longer sets `UV_NO_CACHE`

The `python` tool, and the `uv` tool of the Mac, no longer set
`UV_NO_CACHE = true` in `[mise.env]`. The uv cache is on, so uv keeps one
environment for each script path of a Python task, and a task does not build
its environment again at each run.

An override cannot add a key. Thus an `[overrides.mise.env] UV_NO_CACHE` of
the values now stops the sync. Do these steps before the next sync:

1. Find `UV_NO_CACHE` in `[overrides.mise.env]` of
   `.config/bootstrap/tools.toml`.
2. Remove it. If the repository must keep the cache off, put the key in
   `[env]` of `.config/mise/conf.d/zz-repo/mise.<env>.toml` instead:

   ```toml
   [env]
   UV_NO_CACHE = true
   ```

3. Run `mise run setup:configs`, then commit the changed files.

For the Mac, see [`docs/setup/index.md`](../setup/index.md).

## M47: dprint is the one formatter of the Python files

The `python` tool no longer ships the gate `code:format:40-ruff`. The block
`ruff` of `dprint.jsonc` (the dprint ruff plugin) formats each `.py` and
`.pyi` file. The sync removes the synced file
`.config/mise/tasks/code/format/40-ruff` itself. `ruff format` no longer runs,
so the format settings of `[tool.ruff]` and `[tool.ruff.format]` in
`pyproject.toml` no longer act. The lint gate `code:lint:70-ruff` and the type
gate `code:lint:50-pyrefly` do not change, and still run only with a
`pyproject.toml`. Both now skip `.config/mise/tasks/`.

Do these steps in each repository with a `pyproject.toml`, after the sync:

1. Find the format settings in `pyproject.toml`: the keys of
   `[tool.ruff.format]`, and `line-length`, `indent-width` and
   `target-version` of `[tool.ruff]`.
2. Write each one in the block `ruff` of `dprint.jsonc`, with the name of the
   dprint ruff plugin:

   | `pyproject.toml`                                | Block `ruff` of `dprint.jsonc` |
   | ----------------------------------------------- | ------------------------------ |
   | `line-length`                                   | `lineLength`                   |
   | `indent-width`                                  | `indentWidth`                  |
   | `target-version`                                | `targetVersion`                |
   | `[tool.ruff.format] quote-style`                | `quoteStyle`                   |
   | `[tool.ruff.format] indent-style`               | `indentStyle`                  |
   | `[tool.ruff.format] line-ending`                | `lineEnding`                   |
   | `[tool.ruff.format] skip-magic-trailing-comma`  | `skipMagicTrailingComma`       |
   | `[tool.ruff.format] docstring-code-format`      | `docstringCodeFormat`          |
   | `[tool.ruff.format] docstring-code-line-length` | `docstringCodeLineWidth`       |
   | `[tool.ruff.format] preview`                    | `preview`                      |

   `lineLength` and `indentWidth` are items of `bootstrap` (`120` and `2`), so
   change them with an override in `.config/bootstrap/tools.toml`:

   ```toml
   [tools.dprint.files."dprint.jsonc".ruff]
   lineLength = 100
   ```

   Put each other key directly in the block `ruff` of `dprint.jsonc`. The sync
   keeps it.
3. Remove `[tool.ruff.format]` from `pyproject.toml`. Keep `line-length` and
   `target-version` of `[tool.ruff]` when the lint uses them, for example the
   rule `E501`.
4. In the files of the repository, replace each use of
   `code:format:40-ruff` with `code:format:05-dprint`, for example in a CI
   workflow, a script, a hook or a document. Show the user each change before
   you write it.
5. Run `mise run setup:configs`, then `mise run code:format --all --fix`.
   Commit the changed files.

## M48: The tasks are Python scripts, and the shell libraries are gone

Each task file that `bootstrap` ships is now a Python uv script with a `.py`
name, for example `.config/mise/tasks/setup/mise.py`. The task names, the
flags, the output and the exit codes do not change: mise removes `.py` from
the task name. The sync removes each old copy without the extension that the
manifest lists. An old dispatcher `<dir>/_default` with the bootstrap header
that the manifest does not list hides the new `<dir>/_default.py`, because
mise runs the file without the extension. `bootstrap migrate` removes it, in
the change `remove old copies of the dispatchers`.

The shell libraries of `.config/mise/tasks/_scripts/` are gone, and the sync
removes them. A Python module replaces each:

| Shell library    | Python module       | Tool    |
| ---------------- | ------------------- | ------- |
| `helpers`        | `helpers.py`        | `mise`  |
| `checks`         | `checks.py`         | `mise`  |
| `placeholder`    | `placeholder.py`    | `mise`  |
| `worktrees-page` | `worktrees_page.py` | `mise`  |
| `merge`          | `merge.py`          | `git`   |
| `xcode`          | `xcode.py`          | `swift` |

A task of the repository that sources one of these libraries, or an own part
of a task group that calls a shell function of `helpers` (for example
`print_header`, `bootstrap_get` or `gate_files`), stops at its next run. Move
each such task to a Python uv script. Without such a task, there is nothing
to do.

Do these steps in each repository, after the sync:

1. Run `bootstrap migrate`. It removes an old dispatcher `<dir>/_default`
   with the bootstrap header.
2. Find the tasks of the repository that use a shell library. A file with the
   bootstrap header is not a task of the repository:

   ```bash
   grep -rlE '_scripts/(helpers|checks|placeholder|worktrees-page|merge|xcode)' .config/mise/tasks \
     | xargs grep -L 'Synced from bootstrap'
   ```

   Also examine `.config/mise/tasks/_scripts/local`, and each own part of a
   task group, for example `code/lint/NN-<name>` or `test/50-local`.
3. Write each such task as `<name>.py` with the rules of "Task scripts" in
   the `CLAUDE.md` of `bootstrap`: mode 755, the standard library only, and
   this form. A task of the repository has no `Synced from bootstrap` header.

   ```python
   #!/usr/bin/env -S uv run --script
   # /// script
   # requires-python = ">=3.14"
   # dependencies = []
   # ///
   # MISE description="Lint the proto files"
   # MISE tools={uv="latest", python="3.14"}

   # USAGE flag "--all" help="lint every file of the repository"
   # USAGE flag "--debug" help="emit debug logs"
   # USAGE flag "--fix" help="apply fixes"
   # USAGE arg "[files]" var=#true help="the files to lint"
   # FILES include="*.proto"

   import os
   import sys

   sys.dont_write_bytecode = True
   sys.path.insert(0, f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")
   import helpers  # noqa: E402
   ```

   Change each shell call to its Python form:

   | Shell                                      | Python                                                                                                             |
   | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
   | `source .../_scripts/helpers`              | The preamble above, then `import checks`, `import merge` as necessary                                              |
   | `print_header "Linting ..."`               | `helpers.print_header("Linting ...")`                                                                              |
   | `print_error --bold "..." >&2`             | `helpers.print_error("...", bold=True, file=sys.stderr)`                                                           |
   | `VAR=$(bootstrap_get <key>) \|\| exit $?`  | `var = helpers.bootstrap_get("<key>")["value"]`                                                                    |
   | `members`                                  | `helpers.members()`, a list of `(name, folder, label, url)` tuples                                                 |
   | `own_tasks <group>`, `has_own_task <task>` | `helpers.own_tasks("<group>")`, `helpers.has_own_task("<task>")`                                                   |
   | `mise x -- <tool> <args>`                  | `helpers.tool("<tool>", "<args>")`                                                                                 |
   | `mise x -- which <tool>`                   | `helpers.has_tool("<tool>")`                                                                                       |
   | `${usage_fix}`                             | `os.environ.get("usage_fix")`                                                                                      |
   | `list=$(gate_files "$0" ...) \|\| exit $?` | `listing = helpers.gate_files(__file__, *args, *helpers.usage_files())`, then delete the list in a `finally` block |
   | `yq` or `jq` on the JSON of a command      | The `json` module                                                                                                  |

   A gate gets its files from `helpers.gate_files`, as in
   [`tools/mise.md`](tools/mise.md#the-file-sets-of-the-gates).
4. Delete the old shell file. Keep the task name: `code/lint/80-proto`
   becomes `code/lint/80-proto.py`, and it is still `code:lint:80-proto`.
5. Run `mise run init`, then run each moved task, for example
   `mise run code:lint --all`. Make sure that the output and the exit code are
   the same as before.
6. Commit the changed files.

## M49: `develop` lands on `main` with a fast-forward

In a repository with `main` and `develop` in `core_branches`, a direct
landing of `develop` on `main` (`landing = "direct-merge"` or
`"direct-commit"` of `[branches.main]`) is now a fast-forward. `main` only
moves forward to a commit of `develop`, so `main` is always an ancestor of
`develop`. Before, each landing added a merge commit to `main` that `develop`
did not get, and `git:worktree:list` showed `main` ahead of `develop`.

`git:merge:main` stops before it changes a branch when `main` or `origin/main`
has a commit that `develop` does not have. Do these steps one time, after the
sync:

1. Find the commits of `main` that `develop` does not have:

   ```bash
   git fetch origin main develop
   git log --oneline origin/develop..origin/main
   ```

   If the list is empty, stop here.
2. Merge `main` into `develop`, and push `develop`:

   ```bash
   git switch develop
   git merge origin/main
   git push origin develop
   ```

3. Run `mise run git:merge:main`. It now fast-forwards `main` to `develop`.

A `landing = "pr"` of `[branches.main]` does not change: the forge merges. A
pull request on GitHub cannot fast-forward, so a repository that wants this
model sets `landing = "direct-merge"` for `main`.

## M50: `setup:configs` commits and lands the sync

`setup:configs` now lands each sync as a change of its own. It stops on a
tree with uncommitted or untracked files. It syncs on the branch
`ops/bootstrap-sync`, commits the changes as `ops: sync bootstrap <commit>`,
lands the branch on the default branch with `git:merge:<branch>`, and deletes
the local branch. Under `landing = "pr"`, it opens a pull request, and the
branch stays on `origin` until the request merges. The new key
`default_branch` of `repo.toml` names the default branch. Without it, the
default branch is `develop` when it is a core branch, else `main`.

The new tasks read `default_branch` with `bootstrap config get`, so they need
the release of `bootstrap` that has the key. Do these steps after that
release:

1. Make sure that `mise x -- bootstrap --version` gives that release or newer.
   If not, run `mise up github:virajp/bootstrap`.
2. Commit or stash each change. Then run `mise run setup:configs`. Do not
   commit its result by hand: the task does it.
3. Change each script or hook that runs `setup:configs` and then commits, so
   that it no longer commits.
4. If the base of the changes is not `develop` or `main`, set it:

   ```bash
   bootstrap config set repo.default_branch <branch>
   ```

   The branch must be in `core_branches`.

`git:pull` is a new task: it brings the core branches up to date from
`origin`, also in each member. `git:worktree:create` now cuts a new branch from
the default branch. Use `--current` for the old behavior.

## M51: Each own task is a Python script

All the tasks that `bootstrap` ships are Python uv scripts (M48). The own
tasks of a repository now follow the same rule, also a task that uses no
shell library. Then each task has one form, one library (`helpers.py`) and one
set of rules. M48 moved only the own tasks that source a shell library.

1. Find the own tasks that are not Python. A file with the bootstrap header
   is not an own task:

   ```bash
   find .config/mise/tasks -type f ! -name '*.py' ! -name '*.py.tmpl' \
     | xargs grep -L 'Synced from bootstrap'
   ```

2. Write each such task as `<name>.py` in the form of step 3 of M48: the uv
   shebang, the `# /// script` block, the `# MISE` and `# USAGE` lines, and
   the preamble that imports `helpers`. Keep the `# MISE` and `# USAGE` lines
   of the old file, for example `depends`. Call each tool with
   `helpers.tool(...)` and each other task with `mise run`, and stop with the
   exit code of a failed command. Keep the task name: `p/build` becomes
   `p/build.py`, and it is still `p:build`.
3. Use the output rules of the shared tasks: one `helpers.print_header` first,
   a `helpers.print_subheader` for each later step, and
   `helpers.print_success` for the last line.
4. Move each own shell function of `.config/mise/tasks/_scripts/local` to the
   Python module `.config/mise/tasks/_scripts/local.py`, and import it with
   `import local` after the preamble. Then delete the shell file. An own shell
   library with another name, for example an old `_scripts/_helpers`, gets
   the same move: use the functions of `helpers.py` where they exist.
5. Delete each old file. mise runs the file without the extension first, so
   an old file hides its new `.py` file.
6. Run `mise run init`. Then run each moved task, and make sure that its
   steps and its exit code are the same as before. For a task that changes
   the repository (a release, a version bump, an upgrade), examine the code
   instead, or run it and discard its changes.
7. Change each doc that names an old file path of a task.
8. Commit the changed files.

## M52: `release_env` and `[deploy.<env>]` are gone

`repo.toml` has no `release_env` key and no `[deploy.<env>]` table now. No
task read them. The `[release.<project>.<kind>]` tables describe the releases
of a repository (M53). A sync or `bootstrap check` stops with code 2 while
the file has one of them.

1. Delete the `release_env` line and each `[deploy.<env>]` table of
   `.config/bootstrap/repo.toml` in an editor. `bootstrap config unset` does
   not know these keys now.
2. Run `bootstrap check`, and make sure that it exits 0.
3. Commit `.config/bootstrap/repo.toml`.

## M53: Own release tasks become `[release.<project>.<kind>]` tables

The `git` tool ships `git:tag` and one task `release:<project>:<kind>` for
each `[release.<project>.<kind>]` table of `repo.toml`. An own task that bumps
a version, commits the bump and makes a tag, for example `p:release` and
`p:version`, does the same work. Move it to a table. The keys are in
`docs/design/bootstrap-config.md` of `bootstrap`
("`[release.<project>.<kind>]`").

1. Find the own release tasks, and the release workflows of the repository:

   ```bash
   grep -rlE 'git tag|version --no-git-tag-version|cargo set-version' .config/mise/tasks
   grep -lE '^\s+tags:' .github/workflows/*.y*ml
   ```

2. For each project that has a release, make sure that it has a
   `[projects.<name>]` table (`bootstrap detect --add`).
3. Write one table for each release of the project with
   `bootstrap config set`:
   - A release that CI starts from a tag: `type = "tag"`, the `tag` with
     `{version}` (the tag pattern of the workflow), `version` (the version
     file, or `"tag"`), the `branch` of the tag, `preflight` (the checks that
     the old task ran first) and `workflow` (the workflow file).
   - A second environment of the same project, for example production after
     staging: `type = "tag"` with `from` the first kind.
   - A release that an own task does, for example a build that installs on
     this machine: `type = "task"`, `task` (the own task) and, when the task
     needs fixed arguments, `args`.
4. Move the pending level of each project to
   `bootstrap config set releases.<project> <level>`. A file such as
   `.config/releases.yaml` goes.
5. Run `mise run setup:configs`, then
   `mise run release:<project>:<kind> --dry-run` for each table. Make sure
   that the tag and the version are the ones that the old task gives.
6. Delete the old release tasks, and change each doc and workflow that names
   them.
7. Commit the changed files.

## M54: The release type `local` is now `task`

A `[release.<project>.<kind>]` table of `type = "task"` runs an own task to do
the release: a build on this machine, or a production release that the own
task bumps and tags. The new key `args` gives the task fixed arguments. A sync
or `bootstrap check` stops with code 2 while a table has `type = "local"`.

1. Find the tables:

   ```bash
   grep -n 'type = "local"' .config/bootstrap/repo.toml
   ```

2. In each table that the command finds, change `type = "local"` to
   `type = "task"` in an editor.
3. Run `bootstrap check`, and make sure that it exits 0.
4. Run `mise run setup:configs`. It writes the release task again.

## M55: The Mac tool `claude` owns the shared Claude files

The new Mac tool `claude` has `~/.claude/settings.json`, `~/.claude/CLAUDE.md`
and the task `upgrade:ai`. `claude-cli` and `claude-desktop` depend on it, so
the sync adds it with either one. The values `settings` and `lists` are now in
`[tools.claude]`, and a Mac sync stops with code 2 while `[tools.claude-cli]`
has them. `CLAUDE.md` is now a shared file: bootstrap owns the section between
the `bootstrap:start` and `bootstrap:end` lines at the top, and the own rules
of the owner go after it. The pack `claude` is now `claude-apps`; `macos.toml`
keeps no record of a pack, so this name matters only to `mac add pack`.

1. In `macos.toml`, rename the tables `[tools.claude-cli.settings]` and
   `[tools.claude-cli.lists]` to `[tools.claude.settings]` and
   `[tools.claude.lists]`. Keep the empty `[tools.claude-cli]` table.
2. Run `bootstrap mac sync`, and make sure that it exits 0. The sync adds the
   empty `[tools.claude]` table when step 1 did not write it.
3. Open `~/.claude/CLAUDE.md`. The first sync puts the bootstrap section at the
   top and keeps the old file below the `bootstrap:end` line. Remove from the
   old part each line that the section already has, and keep the own rules of
   the owner.
4. Run `bootstrap mac sync` again, and make sure that it changes nothing.

## M56: `macos.toml` has a fixed place, and the backup clone

`macos.toml` is now at `~/.config/bootstrap/macos.toml` (below
`$XDG_CONFIG_HOME` when it is set). Each `mac` command reads it there, from any
folder. The flag `--repo-path`, the variable `MACOS_SETUP_REPO_PATH`, the
setting `mac.repo_path` and the question `config_path` of the Mac tool `mise`
are gone. The new top-level key `backup` names a git clone: the sync links the
synced files into `<backup>/dotfiles/` and writes through the links, and the
owner commits them. The Mac tools `fish`, `claude` and `ssh` declare their
runtime files with the new key `runtime`, so the sync links their synced
files, not their folders. See `docs/design/mac.md` ("The place of
`macos.toml`" and "The backup clone").

1. Move `macos.toml` from the root of the clone to
   `~/.config/bootstrap/macos.toml`. Make the folder when it does not exist.
   Do not copy it into the clone: the sync of step 4 moves the folder
   `~/.config/bootstrap/` into the clone and links it.
2. A clone of the old symlink model also has the old mise templates of the
   files that the Mac sync renders, with `{{ vars.<name> }}` in them, for
   example `dotfiles/.config/git/identity` and `dotfiles/.ssh/config.identity`.
   They differ from the files in `$HOME`, and the plan of step 3 runs the sync
   in dry-run mode, so it stops on them. They are obsolete: remove them from
   the clone with `git rm`. Also remove `dotfiles/.config/bootstrap/` when an
   earlier try left it in the clone.
3. Run `bootstrap mac migrate --dry-run`. It moves the answer `config_path` of
   `[tools.mise]` to the key `backup`. Make sure that `backup` names the clone,
   then run `bootstrap mac migrate --yes`.
4. Run `bootstrap mac sync --dry-run`. Read each `would move` and `would link`
   line. A conflict stops the run with code 2 and names the file: fix it by
   hand. A file that differs from its copy in the clone needs a decision of the
   owner. A link to a different place, or a link of a folder that the rule now
   splits (for example a link of `~/.config/fish`), must become the correct
   link or a real folder.
5. Run `bootstrap mac sync`, and make sure that it exits 0. Run it again, and
   make sure that it changes nothing.
6. Remove each use of `MACOS_SETUP_REPO_PATH` and `--repo-path` from your own
   scripts and from the old mise config of the clone.
7. In the clone, examine `git status`, and commit the moved files, the removal
   of the old `macos.toml` at the root and
   `dotfiles/.config/bootstrap/macos.toml`.
