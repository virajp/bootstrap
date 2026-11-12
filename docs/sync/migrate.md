# Migrate a repository that has older copies

These instructions are for an AI agent that works in a consumer repository.
They are for a repository that has files from an older version of
`bootstrap`, or from `repo-config`, the old name of `bootstrap`. Read
[`index.md`](index.md) first. Its rules apply here too.

This file is temporary. Remove an item when all repositories have done it.

## Find the items that apply

| The repository has                                                                            | Do                                                                                                                                                      |
| --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.config/bootstrap.toml` with a `tools` list                                                  | The list stops `bootstrap`, and `bootstrap init` does not migrate it. Write one `[tools.<name>]` table for each tool by hand. Then do M17, M16 and M34. |
| `.config/repo-config.toml` with a `tools` list                                                | M15. Its M15.4 includes the row above.                                                                                                                  |
| No values file, but older shared files or a mise config                                       | The steps of `index.md`. Its step 5 sends you here for M1 to M14. Write the new name directly.                                                          |
| `.config/mise/conf.d/_repo/`, or an `env` without `dev`                                       | M17, before the first sync.                                                                                                                             |
| `.config/repo-sync/gitignore`                                                                 | M18, before the next sync.                                                                                                                              |
| Files from the stackgen `tool-config` skill or packs                                          | M19, before the next sync.                                                                                                                              |
| `[tools.vscode]` without values, or no `vscode` in the root                                   | M20.                                                                                                                                                    |
| `@askviraj/linter` in a file of the repository                                                | M21, before the next sync.                                                                                                                              |
| A core branch that must allow direct commits                                                  | M22. Optional.                                                                                                                                          |
| A sync that adds `code:lint:repo-sync`                                                        | M23, after the sync.                                                                                                                                    |
| `[overrides.graphifyignore]`, `.graphifyignore` or `code:graph`                               | M24, before the next sync.                                                                                                                              |
| `.config/fnox.toml` with a top-level `[secrets]` or no profile                                | M25, before the next sync.                                                                                                                              |
| A git submodule without a `[member.<name>]` table                                             | M26, before the next sync.                                                                                                                              |
| A sync that adds `sync-engine/`                                                               | M27, after the sync.                                                                                                                                    |
| A values file that selects one scanner, or a repository that read `required`                  | M28. Nothing to do.                                                                                                                                     |
| A sync that writes the first `.config/repo-generated.tsv`                                     | M29, after the sync.                                                                                                                                    |
| `.config/mise/tasks/setup/configs/sync`                                                       | M31, after the sync.                                                                                                                                    |
| `SYNC_ENGINE_` in a file or a script, or a folder `~/.cache/sync-engine/`                     | M32, after the sync.                                                                                                                                    |
| A file or a script that names a task of the table in M33                                      | M33, after the sync.                                                                                                                                    |
| A `setup:configs` that does not run `bootstrap sync`, or an old form that `bootstrap` reports | M34, before the next sync.                                                                                                                              |
| `version_prefix = "cli-v"` or `version_prefix="cli-v"` in a file of the repository            | M35, before the next sync.                                                                                                                              |
| A Mac with `[tools.fonts]` or `[packs.security]` in `macos.toml`                              | M36, before the next Mac sync.                                                                                                                          |

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
- Delete the old `.config/mise/tasks/setup/configs`. `bootstrap init` (M34)
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
   tools. Use `bootstrap init --dry-run`.
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

An old `[overrides.gitignore]` table needs no manual step: `bootstrap init`
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
`develop`) that exist and do not allow direct commits. The key
`direct_commit` of `[tools.git.branch.main]` and `[tools.git.branch.develop]`
is new. No old form
of it exists.

1. If the repository does not want direct commits on a core branch, do
   nothing. At the next sync, a repository with one core branch gets only
   `--branch <its branch>` in the hook.
2. To allow direct commits on a core branch, set `direct_commit = true` in its
   `[tools.git.branch.<branch>]` table. With `main` and `develop`, `main` can
   allow them only when `develop` allows them too. See "Step 2: Select the
   branch model" of `index.md`.
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
3. Do the same in each git submodule. A submodule without
   `.config/bootstrap.toml` now stops `setup:configs --all` with code 2: adopt
   `bootstrap` in it first.

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

1. Run `mise run setup:configs`. Commit the result.
2. In the files of the repository, replace each old name with its new name,
   for example in a CI workflow, a script or a document. Show the user each
   change before you write it.
3. A task of the repository in one of these groups, without a number prefix,
   runs after the tasks of the tools. If it must run before one of them, give
   it a lower number prefix.

## M34: The CLI `bootstrap` replaces the `repo-sync` tool

The `repo-sync` tool is gone. The `mise` tool now ships the entry:
`setup:configs` runs `bootstrap sync`, and the gate `code:lint:60-bootstrap`
replaces `code:lint:60-repo-sync`. mise installs the CLI as
`github:virajp/bootstrap`. Only `bootstrap init` migrates an old form.
`bootstrap sync` and `bootstrap check` stop with code 2 and the message
`run bootstrap init`.

The code of the CLI moved from `sync-engine/` to `cli/` of `bootstrap`, and the
release tags changed (see M35). The old entry task `setup:configs` reads
`sync-engine/VERSION` of the clone, and installs the release `vX.Y.Z`. Thus it
stops with an error after this release. Do the steps of this item: the old
entry task cannot do them.

`bootstrap init` moves these old forms:

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
3. Commit or discard the changes of the work tree. `bootstrap init` writes no
   file that has changes that are not committed.
4. Run `bootstrap init --dry-run`. Show the user the change list.
5. Ask the user whether to apply the change list. If the user agrees, run
   `bootstrap init --yes`. It migrates each old form, then syncs. Give the
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

The Mac tools `cli`, `network`, `git`, `fnox`, `fonts` and `python` installed
more than one package. Now each package is a tool of its own, in the pack
`base` with `default = true`:

| Old tool  | New tools                                                                                                                             |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `cli`     | `bat`, `coreutils`, `curl`, `eza`, `fzf`, `jq`, `ripgrep`, `zoxide`                                                                   |
| `network` | `doggo`, `gping`, `httping`, `ipcalc`, `prettyping`                                                                                   |
| `git`     | `lazygit`, `worktrunk`                                                                                                                |
| `fnox`    | `age`                                                                                                                                 |
| `fonts`   | `font-fira-code-nerd-font`, `font-hack-nerd-font`, `font-iosevka-nerd-font`, `font-iosevka-term-nerd-font`, `font-meslo-lg-nerd-font` |
| `python`  | `uv`                                                                                                                                  |

`ca-certificates` moved from `cli` to `mise`. The tool `fonts` is gone. A Mac
that selects only `base` gets the same packages after the sync.

The member `1password` of the pack `security` is now `{}`. Thus
`[packs.security]` no longer adds the 1Password app.

1. If `macos.toml` has `[tools.fonts]`, remove that table. Else the Mac sync
   stops with code 3.
2. If `macos.toml` has `[packs.security]` and the owner wants the 1Password
   app, add `[tools.1password]`, with its answer `ssh_vault`. Ask the owner
   first.
3. Run `mise run updateall:configs`. The stale rule removes
   `~/.config/mise/conf.d/fonts/mise.toml`.
