# Migrate a repository that has older copies

These instructions are for an AI agent that works in a consumer repository.
They are for a repository that has files from an older version of
`bootstrap`, or from `repo-config`, the old name of `bootstrap`. Read
[`index.md`](index.md) first. Its rules apply here too.

This file is temporary. Remove an item when all repositories have done it.

## Find the items that apply

| The repository has                                              | Do                                                                                                                                                                                                                                                                  |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.config/bootstrap.toml` with a `tools` list                    | The list stops the sync, and the old synced `setup/configs/sync` loads the removed `tools/base/`. Do step 2 of `index.md` first: copy the sync files by hand. Then write the tables of step 3, then M17 and M16. Then ask the user to run `mise run setup:configs`. |
| `.config/repo-config.toml` with a `tools` list                  | M15. Its M15.4 includes the row above.                                                                                                                                                                                                                              |
| No values file, but older shared files or a mise config         | The steps of `index.md`. Its step 5 sends you here for M1 to M14. Write the new name directly.                                                                                                                                                                      |
| `.config/mise/conf.d/_repo/`, or an `env` without `dev`         | M17, before the first sync.                                                                                                                                                                                                                                         |
| `.config/repo-sync/gitignore`                                   | M18, before the next sync.                                                                                                                                                                                                                                          |
| Files from the stackgen `tool-config` skill or packs            | M19, before the next sync.                                                                                                                                                                                                                                          |
| `[tools.vscode]` in the values file, or no `vscode` in the root | M20.                                                                                                                                                                                                                                                                |

**WARNING:** Do not delete a file before you move its repository-owned
content. The sync never restores repository content.

The sync deletes only three kinds of files: the merge tasks of the other
branch model, a `.config/mise/conf.d/*/mise*.toml` with the bootstrap header
that it did not write, and a file with the bootstrap header at a path where
the sync needs a folder. Thus you must delete each other obsolete file and
rename each old task by hand. Do the items that apply to the repository, in
this sequence. Show the user the list of changes before you make them.

## M1: Delete the old sync files

- Delete `.config/gist-sync/` (an old `.gitignore` with only a header).
- If the repository uses `gist_sync` (macos-setup), delete
  `.config/mise/tasks/_scripts/gist-sync`. Move the lines of
  `.config/gist-sync/gitignore` to `[overrides.gitignore]` as M18 tells you.
- Delete the old `.config/mise/tasks/setup/configs`. Step 2 of `index.md`
  replaces it.

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
Keep `setup:ai` and `code:graph`: the `claude` tool ships them to a submodule
too, and `setup:ai` skips there.

## M5: Move the content of `.config/mise.toml`

The sync replaces `.config/mise.toml`. It keeps only the shared `[settings]`.

1. Move `REPO_NAME`, `MEMBERS` and `MERGE_MODEL*` into the values file:
   - `REPO_NAME` becomes `repo_name`.
   - `MERGE_MODEL_DEVELOP` becomes `merge_model` of
     `[branch.develop]`, and `MERGE_MODEL_MAIN` becomes
     `merge_model` of `[branch.main]`. The old single
     `MERGE_MODEL` goes to both tables. A `direct` value is the default, so
     it can stay out. Remove the old keys `merge_model_develop` and
     `merge_model_main`: nothing reads them.
   - `MEMBERS` is not necessary for submodules. For folders, write
     `[member.<name>]` tables.
2. Move the repository's own `[env]` values to the seed
   `.config/mise/conf.d/zz-repo/mise.<env>.toml` of each environment that needs
   them. Examples are `GITHUB_OWNER`, and the emulator, Temporal and OTEL
   variables of 95octane backend (`mise.dev.toml`).
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
one, tell the user. Examples: node `26` and pnpm `12` in 95octane, where
`node` ships `lts` and `latest`.

## M7: Split an old `conf.d/_base`

claude-plugins has an old folder `.config/mise/conf.d/_base`, from before the
`tools/` layout. Before the first sync, move its content:

| Old content in `_base`                                           | New place                 |
| ---------------------------------------------------------------- | ------------------------- |
| node, pnpm                                                       | the `node` tool           |
| python, uv, `UV_*`                                               | the `python` tool         |
| pre-commit, `PRE_COMMIT_HOME`                                    | the `pre-commit` tool     |
| `MEMBERS`, `REPO_NAME`                                           | the values file           |
| `MEMPALACE_*` and other repository variables                     | `zz-repo/mise.<env>.toml` |
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
of the old file in `commit_scopes`.

## M12: The pre-commit hooks

**WARNING:** Do this before the first sync. The sync replaces
`.config/pre-commit-config.yaml`, and it does not keep the hooks of the
repository.

1. Compare the hooks of the existing file with the rendered file. Use
   `mise run setup:configs --dry-run`.
2. Move each hook that no tool ships to `.config/repo-sync/pre-commit.yaml`.
   Prefix each `entry` that calls a tool with `mise x --`.
3. Remove each hook that a tool now ships, for example `graphify-refresh`,
   `gitleaks` and `conventional-commits`.

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

The next sync removes the old `conf.d/*/mise.toml`, `_env/mise.toml` and
`_alias/mise.dev.toml` files, and the old `mise*.toml` files of
`conf.d/_task/`, that have the header of this repository
(`Synced from repo-config:` or `Synced from bootstrap:`). Delete each other
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
   that apply" now, but not its last sentence: copy the sync files by hand,
   write the tables, then do M17 and M16. If not, do M17 and M16 if they
   apply.
3. Run `mise run setup:configs --dry-run`.
4. Examine the diff. Each synced file changes its header. Other changes come
   from the commits in `bootstrap` since the last sync of this repository.
   Show the user the list of files with changes that are not only the header.
5. If the dry run replaces repository-owned content, stop. Tell the user.
   "Rules" in `index.md` lists the places for repository-owned content.
6. Ask the user to run `mise run setup:configs`. Wait until the user tells
   you the result.
7. Run `mise run init`. This makes all the task scripts executable.

If `repo_sync` fails, read the table of exit codes in step 6 of `index.md`.

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
3. Run `mise run setup:configs` again. It must not show the notice about the
   old values file. Each file must show `unchanged`, `kept seed`,
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

## M18: `.config/repo-sync/gitignore` is now `[overrides.gitignore]`

Do this item before the next sync. The old file stops the sync with code 2.

1. Read `.config/repo-sync/gitignore`. Show the user its lines.
2. For each line, find the type. Show the user the result:
   - A blank line or a comment line: drop it.
   - A line that cancels a shared pattern with `!`, for example
     `!.env.example`: put it in `add`.
   - A line that the repository needs to ignore: put it in `add`.
3. If the old file only cancels a shared line, for example `!*.log` for the
   shared `*.log`, ask the user whether to put `*.log` in `remove` instead.
   `remove` comments out the shared line, and the result is easier to read.
4. Write the lists in `.config/bootstrap.toml`:

   ```toml
   [overrides.gitignore]
   add = [ "/build/", "!.env.example" ]
   remove = [ "*.log" ]
   ```

5. Delete the old file:

   ```bash
   git rm .config/repo-sync/gitignore
   ```

6. Run `mise run setup:configs --dry-run`, and examine the diff of
   `.gitignore`. The lines of `add` are at the end, under
   `# ==== Overrides from .config/bootstrap.toml ====`.

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
   | `.graphifyignore`                             | Put each line that is not in the shared file or in `.gitignore` in `add` of `[overrides.graphifyignore]`. The sync replaces the file. The `claude` tool is necessary.        |
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

1. If the values file has `[tools.vscode]`, remove the table. The entry does
   nothing in a root repository, and a submodule refuses it with code 2.
2. If a root repository did not select `vscode` before, do "Adapt to the
   repository" of [`tools/vscode.md`](tools/vscode.md) before the next sync.
   The sync replaces `.vscode/settings.json`.
