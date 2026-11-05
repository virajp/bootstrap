# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Purpose

This repository is the source of shared files for other repositories. It has no
build and no application code. Its tests are `test:repo-sync`, `test:setup-sync`
and `test:macos-tools`, and the task `test` runs each test task (see "How the
sync operates"). Other repositories copy files from it with their own
`setup:configs` task. Read `README.md` for the sync contract.

## Layout rule

The shared files of a repository are in `repo/tools/<tool>/`, one folder for
each tool. The packs are in `repo/packs.toml`, outside each tool folder, so the
sync never copies it. The examples are in `repo/examples/`, with
`last_sync.toml`. The Mac tools are in `macos/tools/<tool>/`, one folder for
each tool, and their packs are in `macos/packs.toml`. `tools/` holds only the
forwarding file `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`, which
the sync does not ship: an old consumer stub sources it from the clone. These
files are for this repository only:

- At the root: `README.md`, `CLAUDE.md`, `LICENSE`, `dprint.jsonc` (with the
  excludes of this repository) and `eslint.config.mjs`.
- `sync-engine/`: the shared library folder of the sync, with `clone`,
  `common`, `write`, `merge`, `select`, `manifest`, `values`, `mise`,
  `repo-sync` and `setup-sync`, the Mac front-end. It is not a
  tool, so the sync never copies it to a consumer.
- `docs/design/`: the design notes of the sync (packs, tools, mise, fnox and
  the values file). `docs/plans/`: the vwf plan index.
- `.config/mise/tasks/test/`: the tasks `test:repo-sync` (`repo-sync`),
  `test:setup-sync` (`setup-sync`), `test:macos-tools` (`macos-tools`) and
  `test` (`_default`). They are not in `repo/tools/` or `macos/tools/`, so no
  consumer gets them.
- `macos/fixtures/`: the fixtures of `test:setup-sync`.
- `macos/examples/macos.toml`: the sample values file of the Mac side, and the
  model of `test:macos-tools`. A change to a Mac tool also changes this file
  when the tool needs a value or a selection, because `test:macos-tools` fails
  for an unselected tool.

`repo/examples/<example>/` holds sample values files (`bootstrap.toml`), one
folder for each repository shape: `polyrepo`, `polyrepo-main`,
`polyrepo-develop`, `monorepo`, `repo-submodules` and `repo-siblings`.
`repo-submodules` also has one file in each submodule folder, and
`repo-siblings` has one file in each sibling folder only. The sync never
copies or generates them. `docs/sync/index.md` and `docs/sync/vwf.md`
use them as the model of a values file, so a change to the contract of the
values file changes the examples in the same commit.

`repo/examples/last_sync.toml` holds `commit_id`, the `bootstrap` commit that
the examples agree with. This id tracks the examples only. It is not for
consumers, and no consumer reads it. Each example has an empty
`bootstrap_commit`, so a consumer that copies an example gets no wrong
value: `repo_sync` replaces the empty line with the commit of its sync. A
consumer value is newer than `commit_id`, and that is correct.

- To find whether the examples can be out of date, run
  `git log <commit_id>..HEAD -- tools repo docs/sync CLAUDE.md`. A contract
  change in that log needs a review of the examples. The command keeps `tools`
  until `commit_id` is set again, because the contract changes before the move
  to `repo/` are under `tools/`.
- All the examples change together. After a review, set `commit_id` to
  `git rev-parse HEAD`. The value is always older than `main`, because a
  commit cannot contain its own id.

`docs/sync/` tells an AI agent how to adopt `bootstrap` in a consumer
repository, how to bring it up to date, and how to add and remove tools after
the adoption. After the adoption, `repo-sync` keeps the consumer in sync.

- `docs/sync/index.md`: the adoption prompt. It does the first adoption, and
  it brings an adopted repository up to date with `bootstrap_commit`.
- `docs/sync/migrate.md`: the items for a repository with older copies.
- `docs/sync/kotlin-dsl.md`: the static prompt that migrates a Gradle build
  from the Groovy DSL to the Kotlin DSL. `vwf.md` and the "Migration"
  sections of `kotlin.md`, `android.md` and `flutter.md` send an agent there.
- `docs/sync/vwf.md`: the entry point for the vwf and stackgen plugins. It
  accepts the `adopt`, `add tools`, `remove tools`, `add plugins` and
  `remove plugins` instructions. A plugin names this path
  and no other bootstrap path.
- `docs/sync/tools/<tool>.md`: one file for each folder in `repo/tools/`, with
  the same sections in each file. The file of the `claude` tool is
  `claude-tool.md`, because a file system that ignores case reads `claude.md`
  as `CLAUDE.md`.
- `docs/sync/packs/<pack>.md`: one file for each pack of `repo/packs.toml`,
  with the tools that it adds and the questions of the adopting AI.

This repository is also a consumer of its own `base` pack and of the `claude`
tool. Its values file is `.config/bootstrap.toml`, and `setup:configs` writes
the synced copies at the root (`.config/`, `.vscode/`, `.gitignore`,
`.gitattributes`). Do not edit a synced copy. Change the
file in `repo/tools/<tool>/`, push it, then run `mise run setup:configs` here
too.

Each file in `repo/tools/<tool>/` is at the path that it gets in a consumer
repository, without the `.tmpl` suffix of a template. For example,
`repo/tools/git/.config/mise/tasks/code/merge/develop` becomes
`.config/mise/tasks/code/merge/develop` in a consumer.

- Each tool folder has a `tools.toml` with four keys: `default_envs`,
  `depends`, `root_only` and `extensions`. All four keys are
  necessary, because the sync reads them without defaults. `default_envs` is
  the list of the environments of the tool when the values file gives it no
  `env`. It must include `dev`. A `root_only` tool is not for a git
  submodule: in a submodule, the sync skips it when a pack adds it, and stops
  with code 2 when a `[tools.*]` entry names it. The tables `[defaults]`,
  `[questions.<key>]` and `[contributes.<tool>]` are optional (see "Templates"
  and `docs/design/tools.md`).
- A tool folder must not be `zz-repo`, and its name must not sort after
  `zz-repo` in byte order. If it does, the sync stops with code 2. The folder
  `conf.d/zz-repo/` of the repository must load last (see "mise config").
- A pack is a `[pack.<name>]` table of `repo/packs.toml`. Each key is a
  member, on one line: `<tool> = { default = true }` or `<tool> = {}`, and
  `"pack:<name>" = { default = true }` for a nested pack. `[packs.<name>]` of
  a values file adds each `default = true` member. A member with `{}` arrives
  only with `[tools.<name>]`, and a nested pack with `{}` only with its own
  `[packs.<name>]`. A tool in several packs is added once. The old keys
  `depends`, `required`, `tools` and `default` with a list, and the prefix
  `p:`, stop the sync with code 2. A member that is not a tool folder or a
  pack stops it with code 3. A pack name must not be a tool folder, and a pack
  must not include itself. Each sync examines every `[pack.*]` table. The sync
  always adds the `base` pack; its name is fixed in `repo-sync`.
- The member order of a pack is the sync order of `repo-sync`. `setup-sync`
  sorts the tools by name in each stage. The TOML plugin of dprint sorts
  keys, so the root `dprint.jsonc` excludes `repo/packs.toml` and
  `macos/packs.toml`. Keep the order by hand.
- A common tool of a pack holds the items that the other tools of the pack
  use, for example an orchestrator and its hook. Its name starts with `_`, for
  example `_security_base`. It is not a member: each member tool has it in its
  `depends`, so one member tool alone brings it.
- Two tools must not ship the same path. The sync does not examine this rule,
  so examine it yourself when you add a file.
- A file that the consumer owns after the first write is a seed. Give it a
  `[files."<target>"]` table with `mode = "seed"`. A config file that a project
  tunes, for example the rules of a
  linter (`.config/swiftlint.yml`, `.config/detekt.yml`), is a seed.
- A seed that applies only when the consumer has a specific file also gets
  `requires` in that table, for example `requires = ["package.json"]`.
- An optional `[files."<target>"]` table gives the mode and the condition of
  one file of the tool. Without a table, the file is `generated`. The target
  is the path of the file without `.tmpl`. The target, or the target with
  `.tmpl`, must be the same string as a path of `find -type f` in the tool
  folder. Thus a case variant and a path through a symlinked folder fail. A
  target that fails stops the sync with code 2. The keys of the table are:
  - `mode`: `generated` (the default), `seed` or `merge`.
  - `requires`: a list of exact paths, relative to the root of the repository.
    Each path must exist.
  - `requires_found`: a list of file names. The sync must find each name at the
    root or down to 4 folders below it, in a file that git tracks or does not
    ignore.
  - `format`: `text`, `json`, `yaml` or `toml`. It selects the parse check. The
    default comes from the extension of the target.
  - `remove`: only with `mode = "merge"`. A list of key paths to delete from
    the live target.
  - `remove_items`: only with `mode = "merge"`. A list of tables with `path`
    and `items`. It deletes each listed item from the list at `path`, in the
    live target and in the source.
- When a `requires` or `requires_found` entry is not there, the sync does not
  write the target, and it logs the skip. A generated target gets no row in
  `.config/repo-generated.tsv`, so the stale rule removes a listed copy with the
  bootstrap header (see "How the sync operates"). A copy without the header, a
  seed and a merge target stay.
- Merge mode is for a target that an app owns. The source holds only the keys
  that the tool enforces. The sync merges it into the live target, and the
  other keys stay. The format is `json`, `yaml` or `toml`. The order is render,
  parse check, merge, dprint, compare. A merge target has no header, so the
  stale rule leaves it (see "How the sync operates"). It follows no symlink and
  keeps the mode of an existing target. A new target takes the mode of its
  source. `sync-engine/merge` has the merge rules. The JSONC rule does not apply
  to a merge target. No tool uses merge mode now.
- These keys are gone, and each stops the sync with code 2 and a message that
  names the new key: the top-level `seed` key, the old table of seed conditions,
  and the `requires` and the `files` of a mise entry.
- A tool folder with both `<path>` and `<path>.tmpl` stops the sync with code 2.
- Each file that is neither a seed nor a merge source starts with this header,
  in the comment syntax
  of the file (after the shebang in a script, after `---` in YAML when the
  file has it, inside `{` in JSONC, as an HTML comment in Markdown). In a
  template, `<path>` keeps the `.tmpl` suffix:

  ```text
  # Synced from bootstrap: repo/tools/<tool>/<path>. Do not edit this file.
  # The next sync replaces it, and a manual change is lost.
  ```

  A file of a Mac tool has `macos/tools/<tool>/<path>` in the header. The
  forwarding file keeps its old header with `tools/repo-sync/...`, because the
  sync does not ship it.

  A YAML or TOML merge source can have the comment
  `# Merged by bootstrap: <path>`. A JSON merge source has no comment. The sync
  strips nothing from a merge source.
- A tool can add `[[precommit]]` tables to its `tools.toml`. Each table is one
  `repos:` entry of `.config/pre-commit-config.yaml`, with its
  `[[precommit.hooks]]`. This key is optional.
- A tool does not ship mise files. The only exception is the `mise` tool, with
  `.config/mise.toml.tmpl` and the `zz-repo` seeds. A tool writes its mise
  config in `[mise.*]` and `[mise_env.<env>.*]` tables of its `tools.toml`, in
  mise syntax. The sync generates `conf.d/<tool>/mise.<env>.toml` from them,
  one file for each environment of the consumer. A tool that ships a file that
  the sync also generates stops the sync with code 2. See "mise config".

Some files refer to files that are not in `repo/tools/`, for example
`dprint.jsonc`. The consumer repository supplies them.

## How the sync operates

- `sync-engine/repo-sync` defines `repo_sync`. `sync-engine/clone` holds the
  functions of the cache and the clone. The sync embeds that file in the
  consumer task `setup:configs` (`setup/configs/_default`) when it renders the
  template of the task. The task keeps one shallow clone of this repository in
  `repo/` of the cache `${XDG_CACHE_HOME:-~/.cache}/sync-engine/`
  (`SYNC_ENGINE_CACHE` changes the cache folder). All the consumers of the
  machine share it. The URL is `https://github.com/virajp/bootstrap.git`
  (`SYNC_ENGINE_URL` changes it). The task fetches the branch `main` from the
  URL, not from `origin`, into the clone, and copies the clone to a temporary
  folder. Then it runs `setup:configs:sync` in a new process with
  `SYNC_ENGINE_CLONE` set to the copy. It also sets `REPO_SYNC_CLONE` for a
  member that still has the old sync task. `setup:configs:sync` loads
  `sync-engine/repo-sync` from the copy, so a change to the sync code applies
  in the same run. Without `SYNC_ENGINE_CLONE`, it stops with code 2.
- `sync-engine/repo-sync` sources seven libraries from its own folder.
  `common` holds the log, the file mode, the clean-up and its traps, the header
  detection, the record of the commit and the tool wrappers. `write` holds the
  write of one target: the format, the parse check, the temp file, the seed
  gate, the unblock of a folder path, the checks of a `[files]` table, and the
  compare, the mode, the dry-run diff and the rename. `merge` holds merge mode.
  `select` holds the pack expansion and the depends walk. `manifest` holds the
  read and the write of the manifest and the stale rule. Its `macos` link mode
  is for the Mac front-end, `sync-engine/setup-sync`: it resolves
  each link, removes the real file and prunes up to the target root. `values`
  holds the merge of the values of the tools (see "Templates"). `mise` holds
  the mise config generator, the environments of each tool and the checks of
  `[overrides.mise*]` (see "mise config").
- The sync needs `mise`. Without it, `repo_sync` stops with code 2. The
  wrappers `_se_gomplate`, `_se_jq` and `_se_yq` of `common` start gomplate, jq
  and yq from `/`, so no mise config of the repository selects a version. An
  installed tool runs as `MISE_OFFLINE=1 mise x <tool> -- <tool>`, with no
  network. A tool that is not installed runs as
  `mise x <tool>@latest -- <tool>`. Thus the sync operates before a first sync,
  on a new machine and in a scratch folder.
- A file counts as a synced file (it "has the bootstrap header") when one of
  its first 3 lines is a comment line that starts with a header text. The
  comment line has an optional indent, then `#`, `//` or `<!--`. The header
  texts are `Synced from bootstrap: `, `Synced from repo-config: ` and
  `Generated by macos-setup`. A text directly after a `<!--` line that has
  nothing else also counts. A text in the middle of a comment does not count.
- The lock folder `lock/` of the cache makes a second `setup:configs` wait
  until the first has copied the clone. A waiter removes the lock of a dead
  holder, and a lock without a holder that is older than one minute. A failed
  fetch or clone stops the task with code 5. The origin compare is normalized,
  so the SSH and the HTTPS URL of one repository match. A clone with another
  `origin`, for example after a test with `SYNC_ENGINE_URL`, is cloned again.
  A `repo/` that is not a git clone and not empty stops the task with code 2.
- The old stubs used `bootstrap/` and `bootstrap.lock` in
  `${XDG_CACHE_HOME:-~/.cache}`. Each run removes these two while it holds the
  old lock. When a live old stub holds that lock, the run skips the removal,
  and a later run does it. A failed removal does not fail the sync.
- A git repository at the cache root is the old layout. When its raw `origin`
  normalizes to the URL, the run removes each entry of the cache root, except
  `repo/`, the lock and `macos-generated.tsv`. Another `origin` or no `origin`
  stops the task with code 2 and deletes nothing in the cache root.
- `repo_sync` reads the `[tools.<folder>]` and `[packs.<name>]` tables of the
  consumer file `.config/bootstrap.toml`. `env` selects the environments, and
  each other key of a `[tools.<name>]` table is a value of that tool (see
  "Templates"). A `[packs.<name>]` entry has only `env`: another key stops the
  sync with code 2. A list (`tools = [...]`, `packs = [...]`) is the old format
  and stops the
  sync with code 2. It copies the clone of `SYNC_ENGINE_CLONE` and never clones.
  Then it adds the `base` pack, expands each pack to its default members, adds
  the `depends` of each tool, and copies each file of each tool.
- Exit codes: 2 for a values file, a `tools.toml` or a `[pack.*]` table that
  breaks a rule (also a
  missing values file, a missing `mise`, a target that does not parse, or a bad
  `direct_commit`: see "Branch model in the merge tasks"), for a merge source
  or a live merge target that cannot merge, for a merge target that is a symlink
  or is below a symlinked folder, for a `.config/repo-generated.tsv` that is a
  symlink, is below a symlinked folder or is not a regular file, for a
  repository at the cache root with another `origin` or no `origin`, for a
  `repo/` of the cache that is not a clone and not empty, for a sync without
  `SYNC_ENGINE_CLONE`, for a bad `[defaults]`, `[questions]` or `[contributes]`
  table of a selected tool, and for two contributions or a user value that
  conflict. 3 for a tool or a pack that is not in bootstrap, also a member of a
  pack, and for a contribution of a selected tool to a tool that is not in
  bootstrap. 5 for a clone, a fetch or an `autoupdate` that fails, and
  for a `SYNC_ENGINE_CLONE` that is not usable. A failure of one file does not
  stop the other files. The sync returns the highest code of all its failures.
  `repo_sync --check` returns the highest code of its checks, at most 2.
  `--check` with `--dry-run` is a usage error, code 2. The merge of the values
  runs before the mise config and the file loop, so a failure there stops the
  sync before it writes a file of a tool. Each message of each tool logs, and
  the highest code returns, but inside one target tool the merge stops at the
  first contribution conflict or type change of that tool.
- `repo_sync --check` examines only the local rules of the values file: the
  rules that need the values file and the local git repository, for example
  `env`, `environments`, `[member.*]`, the structure of `[overrides]`,
  `direct_commit`, a missing `main` and `develop`, and the profiles of
  `.config/fnox.toml`. It runs after the
  checks of the values file and of `mise`, and before the clone. It makes
  no clone, reads no `SYNC_ENGINE_URL` or `SYNC_ENGINE_CLONE`, writes no file
  and
  does not set `bootstrap_commit`. It runs each local check and prints each
  message, so every error shows at once. The rules that need the clone stay
  with the sync: a tool or a pack that is not in bootstrap, a pack loop, a
  pack named like a tool or another bad `[pack.*]` table, an environment that a
  mise tool does not allow, an
  `[overrides.mise*]` path that no selected tool has, a `remove` line that
  the shared file does not have, a `root_only` tool in a submodule, an old key
  or a bad `[files]` table of a `tools.toml`, and a bad `[defaults]`,
  `[questions]` or `[contributes]` table or a conflict of the values of the
  tools.
- The gate `code:lint:repo-sync` of the `repo-sync` tool runs
  `repo_sync --check`. The `lint` hook runs `code:lint`, which runs the gate
  with the staged files, so the tool adds no hook. The gate is hidden, and it
  ignores `--fix`. It prints nothing and exits 0 without gomplate
  (`mise x -- which gomplate`), with a list of files that does not include
  `.config/bootstrap.toml`, or with no list and no values file. It also prints
  nothing and exits 0, without a validation, in two cases: the cache has no
  clone with `sync-engine/repo-sync`, or the raw `origin` of `<cache>/repo` is
  not one of five literal forms of the default URL. Thus it never sources code
  from another repository at commit time. Else it unsets `SYNC_ENGINE_CLONE`
  and `REPO_SYNC_CLONE`, and loads `sync-engine/repo-sync` of the cache. It
  takes no lock and never clones.
- Each JSON, YAML or TOML target must parse, after its render or copy and
  before dprint. jq reads JSON, and yq reads YAML and TOML. The `format` key of
  its `[files]` table selects the format, else the extension of the target
  without `.tmpl`. A `.json` target without a `format` key with a line that
  starts with `//` after spaces is JSONC, which jq cannot read, so the sync does
  not examine it. An empty JSON target does not parse. A target that does not
  parse keeps its existing file, the other files continue, and the sync returns
  code 2 and does not set `bootstrap_commit`. A dry run examines each target
  too. An existing seed gets no examination, because the sync does not write it.
- The sync formats each rendered template, each generated mise file and each
  merge target with the dprint config of the consumer before it compares the
  file.
- gomplate reads the TOML files: the values file, `repo/packs.toml` (an awk
  scan also gives the file order of its members), each
  `tools.toml` and the consumer file `.config/fnox.toml`.
- A repository does not use the fnox `default` profile, which is only for the
  user-level fnox settings. When `.config/fnox.toml` exists, the sync and
  `repo_sync --check` stop with code 2 on a top-level `[secrets]` table (the
  `default` profile, also when empty), a `[profiles.default]` table, or no
  `[profiles.<name>]` table. A repository with one profile uses its
  `repo_name`. Each `fnox` call passes `-P <name>`.
- The sync never deletes a file, except the merge tasks of the other branch
  model (see "Branch model in the merge tasks"), a synced file at a path where
  the sync needs a folder, and the stale rule: a file with the bootstrap header
  that `.config/repo-generated.tsv` lists and this run does not write. The merge
  tasks and the stale rule skip only a seed and a merge target of a tool of this
  run (a selected tool or a dependency, also through a mise `depends`). At a
  folder path, the sync refuses to remove such a path and fails that file. Thus
  the stale rule removes a listed path with the header at a seed or a merge path
  of a tool that is not in the run. A merge target has no header, so the rule
  leaves it.
  `docs/sync/migrate.md` tells consumers what to delete.
- The manifest `.config/repo-generated.tsv` is committed, and each git submodule
  has its own. Its first line is
  `# Written by bootstrap. Do not edit this file.`. Each other line is
  `<tool><TAB><path>`, sorted, one for each generated file of the run. A seed
  and a merge target get no row. After a clean run, the order is the stale rule,
  the write of the manifest, then `bootstrap_commit`. A run with a failure runs
  no stale rule and writes no manifest. A failed stale removal is a failure, and
  the other rows still run. The first sync has no manifest, so the stale rule
  removes nothing. A real run writes the manifest only when its content changes.
  A dry run prints
  `would remove <path>` for each stale file, and
  `would write .config/repo-generated.tsv` when the content differs.
- The stale rule examines each row of the old manifest whose path this run did
  not handle. Handled paths are the files of the selected tools that this run
  wrote or kept (generated, seed or merge), and the merge tasks of the other
  branch model. A seed or a merge target of a selected tool that the run left in
  place counts as kept, also one that `requires` or its environment skips. A
  generated file that `requires` or its environment skips is not handled. The
  rule removes such a file when it has the bootstrap header, then its empty
  parent folders. A file without the header always stays.
- The sync does not trust the manifest. A row with no TAB, an empty tool or
  path, an absolute path, or an empty, `.`, `..` or `.git` segment is skipped
  with a log line. A listed path that is a symlink or is below a symlinked
  folder is not removed. A listed file below a folder with its own `.git` (a git
  submodule or a nested clone) stays and logs
  `kept <path> (another repository)`. A folder member without its own `.git` is
  part of the root tree. A listed path that is the same file as a handled path
  stays, which covers a case variant. A manifest that is a symlink, is below a
  symlinked folder or is not a regular file is not read. The sync reads it after
  the file loop, so the run keeps its writes and its removals of the merge
  tasks. It logs `FAILED: .config/repo-generated.tsv`, returns code 2, and skips
  the stale rule, the write of the manifest and `bootstrap_commit`.
- After a sync without a failure, `repo_sync` sets the top-level
  `bootstrap_commit` of the values file to the full commit of its clone. It
  replaces only that line, or adds it after the header comments. This is the
  only write to the values file. An agent compares it with `main` to find
  what changed since the last sync. A dry run prints
  `would set bootstrap_commit <commit>`.
- The sync uses the pushed `main` branch. Local changes have no effect on
  consumers until you push them.
- To test a change before you push it, set `SYNC_ENGINE_URL` to a local clone
  with a commit of the change on `main` of that clone. Use
  `mise run setup:configs --dry-run` to show a diff. The next run without that
  variable clones the default URL again.
- The sync writes the shared `.gitignore` with `[overrides.gitignore]` of the
  values file applied: it comments out each line of `remove`, then adds the
  lines of `add` at the end. A line of `remove` that the shared `.gitignore`
  does not have, a line in both lists, and the old consumer file
  `.config/repo-sync/gitignore` stop the sync with code 2.
- The sync does not read `[claude.plugins.user]` and
  `[claude.plugins.project]` of the values file. `setup:ai` of the `claude`
  tool reads them when it runs, as the merge tasks read `[branch.*]`. Each key
  is `"<plugin>@<marketplace>" = "<source>"`. It adds a missing marketplace
  from the source, installs each plugin that is missing at that scope, and
  updates the marketplaces and the plugins.
- The `pre-commit` tool renders `.config/pre-commit-config.yaml` from the
  `[[precommit]]` tables of the selected tools. Then it adds the entries of the
  consumer file `.config/repo-sync/pre-commit.yaml`. Before the render, the
  sync removes each `"--branch", "<b>"` pair of a `no-commit-to-branch` hook
  whose core branch does not exist or has `direct_commit = true`. A hook with
  no pair left goes, and an entry with no hook left goes too (see "Branch
  model in the merge tasks").
- The sync runs `pre-commit autoupdate` on the rendered pre-commit config, so
  each `rev` becomes the newest tag. A later `autoupdate` in the consumer
  picks the same tags, so the two never undo each other. Keep a real tag in
  each `rev` of a `tools.toml`: an older `repo-sync` copies it as it is.
  `REPO_SYNC_NO_AUTOUPDATE=1` skips `autoupdate`, so no network is necessary.
  If `pre-commit` is missing or `autoupdate` fails, the sync keeps the
  existing file and returns code 5.
- `test:repo-sync` is the task of this repository in
  `.config/mise/tasks/test/repo-sync`. It runs `repo_sync --dry-run`,
  `repo_sync --check` and, for a case that compares a written file, `repo_sync`
  of the working tree's `repo/`, `sync-engine/` and `tools/` (the forwarding
  file) against scratch values files and scratch consumers. Some cases of the
  manifest also call `_se_manifest_stale` and `_se_manifest_read` directly, and
  the cases of the `macos` side call `sync-engine/mise` directly. It
  compares the exit code and the message of each case. It needs no network when
  gomplate, jq and yq are installed. Before its
  first `git` call, it unsets the `GIT_*` variables of a git hook, so its
  scratch repositories never write to this repository. It runs the cases in
  parallel, one job for each CPU, and each case in its own folder.
- `test:setup-sync` is the task of this repository in
  `.config/mise/tasks/test/setup-sync`. It runs `setup_sync` of the working
  tree's `sync-engine/` against the fixtures in `macos/fixtures/`, and it
  compares the exit code, the output and the files of each case. The entry cases
  run the rendered task `updateall:configs` of `macos/tools/macos-setup/`
  against a local bare origin, for the clone, the lock and the temporary copy.
  Each case has its own temporary `HOME`, so it never changes the real `$HOME`.
  It needs no network when gomplate, jq and yq are installed.
- `test:macos-tools` is the task of this repository in
  `.config/mise/tasks/test/macos-tools`. It runs `setup_sync` of the working
  tree's `sync-engine/` with each tool of `macos/tools/` and a copy of
  `macos/examples/macos.toml`, in a temporary `HOME`. It examines three items:
  no `[files]` table gives `mode = "seed"` to a JSON, YAML or TOML target, the
  example selects each tool and each rendered file parses, and an example
  without the answer `config_path` of `macos-setup` stops the sync with code 2.
  It needs no network when gomplate, jq and yq are installed.
- `test` (`test/_default`) runs each test task in parallel. No hook and no
  `code:merge` task runs a test, because the tests are slow and a consumer does
  not have them. Run the tests by hand before you commit:
  - A change to `sync-engine/`: `mise run test`.
  - A change to `repo/`: `mise run test:repo-sync`.
  - A change to `macos/`: `mise run test:setup-sync` and
    `mise run test:macos-tools`.

## Templates (`*.tmpl`)

- gomplate renders each `.tmpl` file. The output file has no `.tmpl` suffix.
- The output file takes the git mode of the `.tmpl` file. Make a `.tmpl` that
  renders a task or a script executable (`755`).
- The delimiters are `<% %>`. Do not use `{{ }}` or `[[ ]]` for template
  logic. mise uses `{{ }}` and bash uses `[[ ]]`, and both must reach the
  output unchanged.
- The values come from the consumer file `.config/bootstrap.toml`. The
  keys are in `docs/sync/index.md`, step 3.
- `.members` is the list of member names: the git submodules (the key of the
  `[member.<name>]` table whose `folder` is the submodule path, else the
  `repo_name` of the values file of the submodule, else the path with each
  `/` changed to `_`), then each other `[member.<name>]` key. A name has only
  letters, digits, `-` and `_`, and two members never have the same name. If
  not, the sync stops with code 2. `.member_folders` maps each
  name to its folder (`folder`, default the name), and `.member_types` to
  `submodule` or `folder` (`type`, default from the folder). A task that runs in
  a member uses its folder. The
  member flags (`--all`, `--<name>`) of `setup`, `setup:configs`, `code:format`
  and `code:lint` are only for the submodules: a folder member is part of the
  root tree.
- `.submodules` is the list of the git submodules only, nested ones included.
  It is empty in a repository without submodules.
- `.tools` is the list of the selected tools, with the packs expanded and the
  dependencies added. `.extensions` is the sorted list
  of their VS Code extensions and of the values key `extensions`. In the root
  repository, it also has those of each git submodule. `.precommit` is the list
  of their `[[precommit]]` tables, then the entries of the consumer file
  `.config/repo-sync/pre-commit.yaml`, with the `--branch` pairs of
  `no-commit-to-branch` filtered by `direct_commit` (see "How the sync
  operates").
- `.sync_engine` is the absolute path of the `sync-engine/` folder of the clone
  copy of the run. The template of `setup:configs` reads `clone` from it.
- `.config` is the values of the tool of the template, and `.all` is the values
  of each selected tool, by tool name. A tool without values has an empty table.
  A top-level key `all` or `config` of the values file does not reach a
  template. `sync-engine/values` merges the values of each selected tool in this
  order: `[defaults]` of the tool, then `[contributes.<tool>]` of each selected
  tool in tool-name order, then `[tools.<name>]` without `env`.
- A contribution merges over `[defaults]`: tables merge key by key, a list
  appends its new items, and any other value replaces the default. Two
  contributions that set one key to different values stop the sync with code 2,
  unless both values are tables or both are lists. A contribution to a tool of
  `repo/tools/` that is not selected is ignored. A contribution of a selected
  tool to a tool that is not in `repo/tools/` stops the sync with code 3.
- A value of `[tools.<name>]` depends on the tool. A base or dependency tool is
  a
  tool that pack `base` adds, also through a nested `default = true` pack, or a
  tool in the depends closure of a selected tool, a tool that a mise `depends`
  adds included. A `<tool> = {}` member that the pack does not add does not
  count. It is a base or dependency tool also when a `[tools.<name>]` table
  names
  it. For such a tool, a list appends its new items, and a value whose type
  differs from the merged value stops the sync with code 2. For any other tool,
  a value replaces the merged one. For both, tables merge key by key.
- A TOML, JSON or YAML template of a Mac tool (`macos/tools/`) encodes a string
  value with `data.ToJSON`, for example
  `name = <% .config.name | data.ToJSON %>`. Its value comes from
  `[tools.<name>]` of `macos.toml`, and a string with a quote and a newline can
  add tables or keys to the rendered target. The templates of `repo/tools/` do
  not follow this rule yet, for example
  `repo/tools/mise/.config/mise.toml.tmpl`.
- A missing answer is not an error of the engine. The sync examines only the
  shape of `[questions]`, and it asks nothing. A template that needs an answer
  uses `required`, which fails that file with its message. Exit code 4 is only
  for the Mac front-end.

## Task scripts

Task scripts are in `repo/tools/<tool>/.config/mise/tasks/`. mise uses the
directory path as the task name, for example `code/merge/develop` is
`code:merge:develop`. A `_default` file is the task for the directory name.

- `_scripts/` contains libraries, not tasks. Each task sources them through
  `${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/<name>`.
  - `helpers`: the print functions (`print_header`, `print_success`,
    `print_wait`, and others), `members`, and `shell_files_in_scope`. Use these
    functions for all task output.
  - `checks`: git predicates. They return an exit status only and print
    nothing. The one exception is `get_main_worktree_path`, which prints the
    path of the main worktree.
  - `merge`: `merge_to_destination_branch`, which the `code:merge` tasks
    use.
  - `placeholder`: for a task that no stack fills yet. A placeholder task
    contains the `#PLACEHOLDER` marker and always exits 0. A stack overlay
    replaces the file. Do not edit a placeholder in place. No task in
    `repo/tools/` uses it now.
  - `xcode`: `require_xcode`, `find_xcodeproj` and the simulator functions for
    the Apple tasks.
  - `repo-sync`: the forwarding file, at
    `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync` of this repository.
    The sync does not ship it. An old `setup/configs/sync` loads it from the
    clone with `REPO_SYNC_CLONE`, and it loads `sync-engine/repo-sync` of the
    same clone. The sync code is in `sync-engine/`, not in a tool. The cleanup
    plan removes the forwarding file.
  - `worktrees-page`: `write_worktrees_page`, which writes the HTML page of
    `code:worktrees`. It holds a small font with the glyphs of that task.

  The `mise` tool ships `helpers`, `checks`, `placeholder` and
  `worktrees-page`, the `git` tool ships `merge`, and the `swift` tool ships
  `xcode`.
- Prefix each tool call with `mise x --`, for example `mise x -- dprint check`.
  To find whether a tool is available, use `mise x -- which <tool>`, not
  `command -v`. System commands (`git`, `awk`, `sed`) and `mise` itself do not
  get the prefix. The tools that mise does not install also do not get it, for
  example `xcodebuild`, `swift`, `pod`, `./gradlew` and the VS Code `code`
  command.
- Exception: a tool that a task must have in every repository and environment
  gets a `#MISE tools.<tool>="latest"` directive in that task, and the task
  and the libraries that it sources call it without `mise x --` and find it
  with `command -v`. `mise x` finds only the tools of a config file, and a
  repository before its first sync, or a shell without `MISE_ENV=dev`, has no
  config that names them. The tools are `pre-commit` in
  `setup:configs:sync`, `setup:precommit` and `code:precommit`, and
  `pre-commit` and `yq` in the `code:merge` tasks, and `yq` in
  `code:worktrees`. The engine in `sync-engine/`
  calls gomplate, jq and yq only through its three wrappers.
- To find whether the repository has an optional task, use `has_own_task` and
  `own_tasks` from `helpers`. Do not use `mise tasks info` or a wildcard
  `mise run`. In a submodule, mise also finds the tasks of the root
  repository, and it runs them in the root directory.
- Write shell for bash 3.2 and bash 5.x. macOS has bash 3.2 as `/bin/bash`. Do
  not use namerefs, `${var,,}`, or associative arrays.
- Tasks use `#MISE` and `#USAGE` header comments for metadata and flags. A
  `#USAGE flag "--fix"` is available in the script as `usage_fix`.
- Under `set -e`, write a command substitution that can fail as
  `VAR=$(cmd) || VAR=''`. The existing scripts explain this in comments.
- Only the `init` task is in mise config (`[mise.tasks.init]` of
  `repo/tools/mise/tools.toml`, always dev). It makes the task scripts
  executable. The task files load in every environment.
- A gate of a language is a `code:lint:<gate>` or `code:format:<gate>` task
  in the tool of that language, for example `code:lint:ruff` in `python`.
  `code:lint:repo-sync` of `repo-sync` follows the same rules for the values
  file `.config/bootstrap.toml`:
  - Without its tool (`mise x -- which <tool>`) or without a file of its
    language, the gate prints nothing and exits 0. Some gates examine a
    project file instead, for example `pubspec.yaml` for dart and `gradlew`
    for android. The task files load in every environment, also where the
    tool is not installed.
  - With a list of files, it uses only the files of its language. Without
    files, it lists them with `git ls-files --cached --others
    --exclude-standard`, because most linters do not read `.gitignore`. A
    tool that scans its project itself (ruff, pyrefly, detekt) uses that list
    only to find whether a file exists.
  - Without `--fix`, a format task is the check of its tool. Do not add a
    lint task that runs the same check, for example `code:lint:ktlint` next
    to `code:format:ktlint`.
  - Its config file, when it has one, is a seed.

`bootstrap` does not ship these items on purpose. Do not add them again
without a new reason:

- `code:count`: it only counts the lines of code.
- `.config/claude-status.json`: claude-status finds the name of the
  repository from `origin`. The file is plain JSON, so it cannot have the
  bootstrap header.
- `.config/miserc.toml` with `env_conf_d`: the sync writes only folder
  fragments `conf.d/<tool>/mise.<env>.toml`, and that setting does not apply
  to them.
- A `code:check` task group: a check of a language is a `code:lint:<name>`
  task, for example `code:lint:uv` for `uv lock --check`.

## Task output: headers and subheaders

An orchestrator is a task that runs other tasks with `mise run`. A leaf is a
task that runs only commands.

1. A task first checks whether it applies. If it does not apply, it prints
   nothing and exits 0. For example, `setup:deps:python` without a
   `pyproject.toml`.
2. A leaf prints `print_header` first, and only once. Nothing comes before
   it.
3. Each later step of a leaf that needs separation is a `print_subheader`.
4. Status lines (`print_wait`, `print_success`, `print_info`, `print_warn`,
   `print_error`) go under a header or a subheader. They never replace one.
5. An orchestrator prints nothing for a `mise run`. The subtask prints its own
   header. Exception: before `mise run --cd <member>`, it prints a header that
   names the member, because a member is a separate repository.
6. A group of commands in an orchestrator is its own subtask, for example
   `setup:linter`. A single direct command in an orchestrator gets a
   `print_header`, because it is at the level of the sibling subtasks.
7. The top-level orchestrator ends with `line_sep1` and
   `print_success "<Done message>"`, not with a header.
8. Header and subheader text is an "-ing" verb phrase with " ..." at the end,
   for example "Installing pre-commit hooks ...".
9. A task that runs only inside a git hook (`code:git-config`) prints status
   lines only, no header.
10. Exception to rules 2 and 5: `code:sec` prints a `print_header` before it
    runs its `code:sec:<scanner>` tasks. Each scanner task prints only a
    `print_subheader` under that header. With `--staged`, a scanner task
    without a staged mode prints nothing.
11. Exception to rule 2: when its output is not a terminal, `code:worktrees`
    prints only JSON Lines, because an agent reads each line.

## Branch model in the merge tasks

The branches `main` and `develop` select the model. A branch counts if it
exists locally or on `origin`. `repo_sync` writes the merge tasks of the model
that it finds, and removes the others. It stops if neither branch exists.

- `main` and `develop`: work starts on feature branches.
  `code:merge:develop <branch>` lands a branch on `develop`.
  `code:merge:main` lands `develop` on `main`, and runs only from `develop`.
- Only one of them: `code:merge <branch>` (`code/merge/_default`) lands any
  feature branch on that branch.
- Merges use `--no-ff`. A hook pass over all files runs first as a safety net,
  with `SKIP=no-commit-to-branch`: that hook guards a commit, and the safety
  net makes none. Thus `code:merge:main` also runs from a protected
  `develop`.
- The merge tasks read `[branch.<destination>]` of the repository's own
  `.config/bootstrap.toml` when they run: the root and each submodule read
  their own file. `merge_model` selects
  `direct` (local merge and push, the default) or `pr` (push and open a pull
  request with `gh` or `glab`). Any other value reads as `direct`. Without an
  `origin` remote, `direct` merges locally and skips the pull and the push,
  and `pr` stops before the hook pass. The `github`
  and `gitlab` tools install these
  CLIs. `merge_from` lists the only source branches, and `merge_never` the
  forbidden ones. A table with both stops the merge. Without both, the
  defaults are: into `main` with `develop`, only from `develop`; into another
  branch, never from `main`.
- `direct_commit = true` in `[branch.main]` or `[branch.develop]` allows a
  direct commit to that branch. `repo_sync` reads it, and the merge tasks do
  not. The `no-commit-to-branch` hook protects only the core branches that
  exist and do not allow direct commits: the sync removes the `--branch` pair
  of each other core branch, and the hook when no pair is left. Thus a
  repository with one core branch gets only `--branch <its branch>`. With
  both branches, `main` can allow direct commits only when `develop` does.
- `repo_sync` validates `direct_commit` before it writes a file, also when it
  is `false`. A value that is not a boolean, the key in another
  `[branch.*]` table, the key of a core branch that does not exist, and
  `main` true without `develop` true when both exist stop the sync with
  code 2.

## mise config

- The consumer selects its environments with `environments` in the values
  file: `dev` (necessary) and `ci`. Any other name stops the sync with code 2.
  bootstrap is only for `dev` and `ci`: a `tools.toml` names only these two in
  `default_envs` and `envs`, and any other name stops the sync with code 2.
  The sync ignores each environment that `environments` does not select.
  Without `environments`, the sync uses `dev` only and logs a warning.
- `.config/mise.toml` (`repo/tools/mise/.config/mise.toml.tmpl`) is the only
  file for every environment. It holds the shared `[settings]` and the `[env]`
  values `REPO_NAME` and `MEMBERS` (the folders of the members, see
  "Templates"). It does not set
  `min_version`. `monorepo_root` and `[monorepo]` are not supported.
- Each file in `.config/mise/conf.d/` is for one environment:
  `mise.<env>.toml`. mise loads it only when `MISE_ENV` includes that
  environment. In a repository, no tool ships or generates a
  `conf.d/*/mise.toml`.
- The generator is `sync-engine/mise`, with the sides `repo` and `macos`, and
  `docs/design/mise.md` has the rules of each side.
- The sync finds the environments of each tool:
  - A tool of a `[tools.*]` entry gets its `env`, else its `default_envs`. A
    tool of a `[packs.*]` entry gets the `env` of that entry, else its own
    `default_envs`. Each tool of the `base` pack gets its `default_envs`. A
    tool that is selected more than once gets each of these environments.
  - A dependency (a tool of `depends`, or the tool that defines a mise
    `depends`) gets each environment of each tool that needs it.
  - `environments` limits every result.
  - An `env` name that is not in `environments`, an empty `env`, or an `env`
    item with a space stops the sync with code 2.
  - Each tool must be in `dev`. An `env` without `dev` stops the sync with
    code 2. A `default_envs` without `dev` also stops the sync with code 2.
    The sync examines the `default_envs` of each tool of the clone, also of a
    tool that the consumer does not select.
  - The `envs` of a mise tool (default `["dev"]`) lists the environments that
    it allows. A tool in an environment that one of its mise tools does not
    allow stops the sync with code 2.
- The sync generates `conf.d/<tool>/mise.<env>.toml` from the `[mise.*]`
  tables of each selected tool:
  - Each mise tool is in the environments of its tool.
  - `[tools]`, `[tool_alias]` and `[plugins]` follow each entry. `[settings]`,
    `[env]`, `[tool_config]`, `[task_config]`, `[task_templates]`, `[doctor]`,
    `[deps]` and `[oci]` go to each environment of the tool.
  - `[vars]`, `[secrets]`, `redactions`, `[shell_alias]`, `[hooks]`,
    `[[watch_files]]`, `[tasks]`, `[bootstrap.*]`, `[dotfiles]` and the daemon
    sections are only for dev, when the tool is in dev.
  - Any other section stops the mise config of that tool with code 2. The
    sync keeps the existing `conf.d` files of that tool, and it still syncs
    the other files of the tool.
  - A `[mise.*]` string can hold `<% %>`. It is rendered with `.config`, `.all`
    and `.env` (`HOME`, `USER`), and each string literal inside the expression
    is in backticks. A failed render stops the mise config of that tool with
    code 2, as above. The overrides of the values file are not rendered.
  - `[mise_env.<env>.<section>]` adds content to one environment only, also
    a dev-only section. `mise`
    uses it for `os = "macos"` in `dev` and `os = "linux"` in `ci`. A mise
    tool in `[mise_env.<env>.tools]` is only in that environment, and its tool
    gets no `envs` check for it. The sync reads `envs` and `requires_found` only
    in `[mise.tools]`, and `requires_found`, `requires` or `files` in a
    `[mise_env.<env>.tools]` entry stops the sync with code 2. For example,
    `kotlin-lsp` of `kotlin` is in
    `[mise_env.dev.tools]`, so `kotlin` can be in `ci`.
  - The mise `depends` of an entry adds the tool that defines that mise tool.
    Do not repeat it in the `depends` key. Two tools must not define the same
    mise tool.
  - The optional `requires_found` of an entry is a list of file names. The
    sync writes the entry only when the repository has a file with each name,
    at the root or down to 4 folders below it, that git tracks or does not
    ignore. The sync does not look in git submodules. The generated file has
    no `requires_found`. For example, `python` installs `ruff` only with a
    `pyproject.toml`.
  - Each file states its condition in its own `[files]` table (see "Layout
    rule"), for example the tasks of a mise tool. A mise entry has no list of
    files. For example, `python` writes `code:lint:ruff` only with a
    `pyproject.toml`, because the table of that task has
    `requires_found = ["pyproject.toml"]`.
- The sync skips a shipped `conf.d/*/mise.<env>.toml` of an environment that
  the consumer did not select.
- The stale rule removes a `conf.d/*/mise*.toml` with the bootstrap header that
  `.config/repo-generated.tsv` lists and this run does not write (see "How the
  sync operates"). Thus a removed tool, environment or folder disappears from
  consumers.
- The shell aliases `setup`, `precommit` and `worktrees` come from
  `repo/tools/mise/tools.toml`, and the sync writes them in the generated
  `mise/mise.dev.toml`. For a repository with git submodules, the sync adds
  `setup-all` and `setup-<submodule>` there, also for nested submodules.
- `[overrides]` of the values file replaces the defaults of the tools.
  `[overrides.mise.<path>]` applies to each environment, and
  `[overrides.mise_env.<env>.<path>]` to one environment, where it wins.
  `<path>` is the path below `[mise]` or `[mise_env.<env>]` of a
  `tools.toml`, without the tool name. The sync applies it to each selected
  tool with a value at that path. A table merges key by key; any other value
  replaces the default. An override replaces only a key that a selected tool
  has: it cannot add a key. A path that no selected tool has, a key of
  `[overrides]` other than `mise`, `mise_env` and `gitignore`, an
  environment outside `environments`, and the `envs`, the `depends`, the
  `requires`, the `requires_found` or the `files` of a mise tool stop the sync
  with code 2. Each overridden line gets an
  `# override: .config/bootstrap.toml` comment. `[overrides.gitignore]` is
  different: see "How the sync operates".
- `conf.d/zz-repo/` is the folder that each repository owns and customizes.
  `mise` ships the seeds `zz-repo/mise.<env>.toml`, for example for the
  repository's own `[env]` values. The sync writes a seed only when it is
  missing. mise loads the folders of `conf.d/` in name order, and a later
  folder wins. `zz-repo` sorts after each tool folder, so the config of the
  repository wins over the config of each tool.
- Each mise tool comes from a bootstrap tool. A tool that a repository needs
  gets added to a tool in `repo/tools/`, or becomes a new tool, for example
  `doppler`.

## Change a shared file

1. Change the file in `repo/tools/<tool>/`.
2. If the change adds, removes or changes a file, a mise tool, a task, a hook,
   a seed or a value of the tool, change `docs/sync/tools/<tool>.md` too
   (`claude-tool.md` for `claude`). A
   new tool needs a new file in `docs/sync/tools/`, and a line in the list of
   `docs/sync/index.md`. A removed file needs a row in "Files that bootstrap no
   longer ships" of `docs/sync/index.md`, and a `migrate.md` item. A change to
   `repo/packs.toml` changes
   `docs/sync/packs/<pack>.md` too, and a new pack needs a new file there.
3. Commit on a local work branch. Land it on `main` with
   `mise run code:merge <branch>`, which merges and pushes.
4. In this repository and in each consumer repository, run
   `mise run setup:configs`, then commit the result. This repository sets no
   `direct_commit`, so the `no-commit-to-branch` hook refuses a commit on
   `main`. Thus in this
   repository, sync on a local `sync-<name>` branch, commit
   `ops: sync bootstrap with <change>`, and land it with
   `mise run code:merge sync-<name>`. The sync also updates
   `bootstrap_commit`, which stays one merge behind `main`.
