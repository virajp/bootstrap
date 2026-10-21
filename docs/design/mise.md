# mise design

This document describes the way mise will be used across all projects. It is the
base tool which is required in all repos and is expected to be already installed
in the machine before setting up with `bootstrap`

## Rules

- `mise` is ONLY used in `dev` & `ci` environments
- `mise` configuration is always stored in `.config/` folder

## Usage

Mise will be used to manage tools, environments, and tasks consistently across
all projects

## Repository Types

### Polyrepo

Mise setup is straight forward in polyrepos, where each repository is
independent and contains its own tools and configuration

### Monorepo

A monorepo setup will be very similar to that of polyrepo, with the main
difference being that all tools and configurations are common for all projects
in the monorepo

### Repos with Submodules

In repositories with submodules, mise setup requires careful handling to ensure
that each submodule's tools and configurations are correctly integrated.

- Each repository (root as well as submodules) will have it's own `mise`
  configuration
- Root repo will have tasks that can call tasks in submodules as needed
- The root repo or submodule can be either a polyrepo or a monorepo

## Design Overview

Mise configuration are generated from templates and each tool will have its own
configuration section in the `tools.toml` file. The sync process reads these
sections and generates the necessary mise configuration files accordingly.

## The parts of the design

mise has three jobs in a consumer repository:

| Job          | Files                                                  |
| ------------ | ------------------------------------------------------ |
| Tools        | `[mise.tools.*]` tables of each `tools.toml`           |
| Environments | `.config/mise.toml` and `.config/mise/conf.d/`         |
| Tasks        | The task scripts in `.config/mise/tasks/` of each tool |

A tool never ships a mise config file. It writes its mise config in the
`[mise.*]` tables of its `tools.toml`, in mise syntax. The sync generates the
mise files from these tables.

## The files in a consumer repository

```text
.config/
├── bootstrap.toml                  # the values file; the sync reads it
├── mise.toml                       # synced; the only file for every environment
└── mise/
    ├── conf.d/
    │   ├── <tool>/mise.<env>.toml  # generated from tools/<tool>/tools.toml
    │   └── zz-repo/mise.<env>.toml # seeds; the repository owns them; loads last
    └── tasks/
        ├── _scripts/               # libraries, not tasks
        ├── code/                   # code:* tasks
        └── setup/                  # setup:* tasks
```

| File                                          | Kind      | Owner          | Loads when                  |
| --------------------------------------------- | --------- | -------------- | --------------------------- |
| `.config/mise.toml`                           | Template  | `mise`         | Always                      |
| `.config/mise/conf.d/<tool>/mise.<env>.toml`  | Generated | `<tool>`       | `MISE_ENV` includes `<env>` |
| `.config/mise/conf.d/zz-repo/mise.<env>.toml` | Seed      | The repository | `MISE_ENV` includes `<env>` |

Without `MISE_ENV`, no file in `conf.d/` loads. No tool ships or generates a
`conf.d/*/mise.toml`.

mise loads the folders of `conf.d/` in name order, and a later folder wins.
The name `zz-repo` sorts after each tool folder, so the config of the
repository wins over the config of each tool. A tool folder that is
`zz-repo`, or whose name sorts after `zz-repo` in byte order, stops the sync
with code 2.

## Environments

The consumer selects its environments with `environments` in the values file.

| Environment | Rule      | Use                                      |
| ----------- | --------- | ---------------------------------------- |
| `dev`       | Necessary | Local development. Each tool is in `dev` |
| `ci`        | Optional  | The CI build and the examination of code |

## `.config/mise.toml`

`min_version`: The minimum version of the mise cli required by the
consumer repository. It is not set now. When it is set, it must be continously
upgraded as new versions of the mise configuration are released.

The template is `tools/mise/.config/mise.toml.tmpl`. It holds the shared
`[settings]` and the `[env]` values from the values file.

| `[settings]` key                     | Value          |
| ------------------------------------ | -------------- |
| `activate_aggressive`                | `true`         |
| `all_compile`                        | `false`        |
| `env_shell_expand`                   | `true`         |
| `gpg_verify`                         | `true`         |
| `lockfile`                           | `false`        |
| `minimum_release_age`                | `"10h"`        |
| `raw`                                | `true`         |
| `status.missing_tools`               | `"always"`     |
| `task.disable_spec_from_run_scripts` | `true`         |
| `task.output`                        | `"interleave"` |
| `task.timings`                       | `true`         |

> As of now, we will NOT be supporting `monorepo_root` and `[monorepo]`
> The repository puts its own `[env]` values in the `zz-repo` seeds.

> `.config/mise.toml` does not set `os`. The `mise` tool sets `os = "macos"` in
> `dev`
> and `os = "linux"` in `ci`, with `[mise_env.<env>.settings]` tables.

## The mise config of a tool

### The `[mise.*]` sections

The sync sorts each `[mise.<section>]` table into one of three classes:

| Class    | Sections                                                                                                                                            | Goes to                                      |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| Entry    | `tools`, `tool_alias`, `plugins`                                                                                                                    | The environments of each entry               |
| Folder   | `settings`, `env`, `tool_config`, `task_config`, `task_templates`, `doctor`, `deps`, `oci`                                                          | Each environment where the folder has a tool |
| Dev only | `vars`, `secrets`, `redactions`, `shell_alias`, `hooks`, `watch_files`, `daemons`, `daemon_groups`, `daemons_settings`, `daemon_providers`, `tasks` | `dev` only                                   |

Any other section stops the sync for that tool, with code 2.

A `[mise_env.<env>.<section>]` table adds content to one environment only. Its
section must be in one of the three classes.

### The keys of a mise tool

A `[mise.tools.<name>]` table has the mise keys of the tool, for example
`version`, and two keys for the sync:

| Key       | Meaning                                                                            |
| --------- | ---------------------------------------------------------------------------------- |
| `envs`    | The environments that the mise tool allows. The default is `["dev"]`. Not written. |
| `depends` | The mise tools that this mise tool needs. mise reads it too, so it stays.          |

### The environments of a mise tool

The sync calculates the environments of each tool in this sequence. Each
mise tool gets the environments of its tool.

1. A tool of a `[tools.<folder>]` entry gets its `env`, else the
   `default_envs` of its `tools.toml`. A tool of a `[packs.<name>]` entry gets
   the `env` of that entry, else its own `default_envs`. Each tool of the
   `base` pack gets its own `default_envs`. A tool that is selected more than
   once gets each of these environments.
2. Each dependency gets the environments of the tool that needs it: a tool of
   the `depends` key, and the tool of a mise `depends`. The sync repeats this
   step until nothing changes.
3. The sync keeps only the environments that the values file selects.
4. Each tool must be in `dev`. An `env` without `dev` stops the sync with
   code 2. A `default_envs` without `dev` also stops the sync with code 2.
   The sync examines the `default_envs` of each tool of the clone, also of a
   tool that the values file does not select, so the tests of `bootstrap`
   find it first.
5. The `envs` of each mise tool must include each environment of its tool. If
   not, the sync stops with code 2. The message names the cause: the
   `default_envs`, the `env` of the values file, or the tool that needs the
   dependency. For example, if `node` has
   `envs = ["dev", "ci"]` and needs `pnpm` with `envs = ["dev"]`, a consumer
   with `ci` gets this error. The error shows a wrong `envs` in a `tools.toml`
   of `bootstrap`, so the tests of `bootstrap` must find it first.

### The dependencies between tools

A tool gets other tools in two ways:

| Link                       | Example                                | Effect                                      |
| -------------------------- | -------------------------------------- | ------------------------------------------- |
| `depends` of `tools.toml`  | `kotlin` has `depends = ["java"]`      | The sync adds the tool `java`.              |
| mise `depends` of an entry | `android-sdk` has `depends = ["java"]` | The sync adds the tool that defines `java`. |

The sync resolves the tools depth first. A dependency comes before the tool
that needs it. Do not repeat a mise `depends` in the `depends` key.

Two tools must not define the same mise tool. The sync reads all the
`tools.toml` files of `bootstrap`, also the files of the tools that are not
selected, and stops with code 2 for a duplicate.

### The generated files

The sync writes one file for each selected environment that has content:
`.config/mise/conf.d/<tool>/mise.<env>.toml`. The folder is the name of the
tool. Each file starts with the bootstrap header. There is no `_base` folder,
because `base` is a pack, not a tool. See [packs.md](packs.md#the-base-pack).

- If a tool ships a `conf.d/*/mise.<env>.toml` file and also generates it, the
  sync stops for that tool.
- The sync skips a shipped `conf.d/*/mise.<env>.toml` of an environment that
  the consumer does not select.
- After the sync, it removes each `conf.d/*/mise*.toml` with the bootstrap
  header that it did not write. A file without the header stays.

## The mise tools

This table is the target state, with the tools of the `base` pack. The
Flutter packs add `android` or `swift`, so the mise tool `flutter` has no mise
`depends`.

| Tool          | mise tool              | Version  | `envs`  | mise `depends` | Other `[mise.*]` sections                           |
| ------------- | ---------------------- | -------- | ------- | -------------- | --------------------------------------------------- |
| `android`     | `android-sdk`          | `23`     | dev, ci | `java`         |                                                     |
| `dprint`      | `dprint`               | `latest` | dev, ci |                |                                                     |
| `gitleaks`    | `gitleaks`             | `latest` | dev, ci |                |                                                     |
| `grype`       | `grype`                | `latest` | dev, ci |                |                                                     |
| `linter`      | `npm:@askviraj/linter` | `latest` | dev, ci | `node`, `pnpm` |                                                     |
| `linter`      | `shellcheck`           | `latest` | dev, ci |                |                                                     |
| `mise`        | `jq`                   | `latest` | dev, ci |                | `tasks` (`init`), `shell_alias`; `mise_env` (`os`)  |
| `mise`        | `yq`                   | `latest` | dev, ci |                |                                                     |
| `osv-scanner` | `osv-scanner`          | `latest` | dev, ci |                |                                                     |
| `pre-commit`  | `pre-commit`           | `latest` | dev     | `python`, `uv` | `env` (`PRE_COMMIT_HOME`), `shell_alias`            |
| `repo-sync`   | `gomplate`             | `latest` | dev     |                |                                                     |
| `bun`         | `bun`                  | `latest` | dev, ci |                |                                                     |
| `claude`      | `pipx:graphifyy`       | `latest` | dev     | `python`, `uv` |                                                     |
| `claude`      | `pipx:mempalace`       | `latest` | dev     | `python`, `uv` |                                                     |
| `cloudflare`  | `npm:cf`               | `latest` | dev, ci | `node`, `pnpm` |                                                     |
| `doppler`     | `doppler`              | `latest` | dev     |                |                                                     |
| `flutter`     | `flutter`              | `3`      | dev, ci |                |                                                     |
| `fnox`        | `fnox`                 | `latest` | dev     |                | `env` (`FNOX_CONFIG_DIR`)                           |
| `gcloud`      | `gcloud`               | `latest` | dev, ci | `python`, `uv` | `env` (`CLOUDSDK_PYTHON`)                           |
| `github`      | `gh`                   | `latest` | dev, ci |                |                                                     |
| `github`      | `actionlint`           | `latest` | dev, ci |                |                                                     |
| `gitlab`      | `glab`                 | `latest` | dev, ci |                |                                                     |
| `java`        | `java`                 | `lts`    | dev, ci |                |                                                     |
| `kotlin`      | `http:kotlin-lsp`      | `latest` | dev     |                | Platform URLs for macOS                             |
| `node`        | `node`                 | `lts`    | dev, ci |                | `settings`, `env` (`_.path`), `shell_alias` (`npx`) |
| `node`        | `pnpm`                 | `latest` | dev, ci | `node`         |                                                     |
| `pulumi`      | `pulumi`               | `latest` | dev     |                |                                                     |
| `python`      | `python`               | `3.14`   | dev, ci |                | `settings`, `env` (`UV_*`)                          |
| `python`      | `uv`                   | `latest` | dev, ci | `python`       |                                                     |
| `rust`        | `core:rust`            | `latest` | dev, ci |                | `profile = "default"`                               |
| `rust`        | `zola`                 | `latest` | dev, ci |                |                                                     |

These tools have no mise tools: `docker`, `swift` (the host Xcode), and
`vscode`. `kotlin` gets `java` through its `depends` key.

Because the `base` pack needs `node`, `pnpm`, `python` and `uv`, each consumer
gets the `node` and `python` tools. They get the environments of the tools
that need them: `dev` and `ci` for `node` and `pnpm` (the `linter` tool), and
`dev` for `python` and `uv` (the `pre-commit` tool).

## Tasks

### Names and files

mise uses the path below `.config/mise/tasks/` as the task name. For example,
`code/merge/develop` is `code:merge:develop`. A `_default` file is the task of
its folder. A `.tmpl` task is rendered by the sync, and the output keeps the
git mode of the template.

Only the `init` task is in mise config, in `[mise.tasks.init]` of
`tools/mise/tools.toml`. It makes the task scripts executable. The task files
load in each environment.

### Orchestrators and groups

An orchestrator runs each task of a group with `own_tasks <group>` from
`helpers`. A tool adds a step to an orchestrator when it ships a task in that
group. The orchestrator finds only the task files of the repository itself,
not the tasks of a parent repository.

| Orchestrator     | Groups that it runs                                       | Other steps                                                                                                              |
| ---------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `setup`          | `setup:secrets:*`, then `setup:deps:*` (with `--upgrade`) | `setup:mise`, `setup:precommit`; with `--upgrade`: `setup:external:update`, `setup:dprint`, `setup:ai`; then each member |
| `setup:worktree` | `setup:secrets:*`, then `setup:deps:*`                    | `setup:mise`, `setup:precommit`; then each member                                                                        |
| `code`           | None                                                      | `code:format`, `code:lint`, `code:sec`                                                                                   |
| `code:format`    | `code:format:*` (with `--fix` and the files)              | `dprint`; then each selected member                                                                                      |
| `code:lint`      | `code:lint:*` (with `--fix` and the files)                | Each selected member                                                                                                     |
| `code:sec`       | `code:sec:*` (with `--staged` and `--debug`)              | First, it stops if `code:sec:gitleaks`, `code:sec:grype` or `code:sec:osv-scanner` is missing or not executable          |

`has_own_task` finds an optional task, for example `setup:ai` or
`setup:external:update`.

### The tasks of each tool

| Tool             | Tasks                                                                                                                 |
| ---------------- | --------------------------------------------------------------------------------------------------------------------- |
| `mise`           | `setup`, `setup:mise`, `setup:worktree`, `code`, `code:worktrees`                                                     |
| `repo-sync`      | `setup:configs`, `setup:configs:sync`                                                                                 |
| `git`            | `code:git-config`, `code:merge` or `code:merge:develop` and `code:merge:main`                                         |
| `pre-commit`     | `setup:precommit`, `code:precommit`                                                                                   |
| `dprint`         | `code:format`, `setup:dprint`                                                                                         |
| `linter`         | `code:lint`, `setup:linter`                                                                                           |
| `_security_base` | `code:sec`                                                                                                            |
| `gitleaks`       | `code:sec:gitleaks`                                                                                                   |
| `grype`          | `code:sec:grype`                                                                                                      |
| `osv-scanner`    | `code:sec:osv-scanner`                                                                                                |
| `android`        | `setup:deps:android`, `code:lint:android`                                                                             |
| `claude`         | `setup:ai`, `code:graph`                                                                                              |
| `docker`         | `setup:external:check`, `setup:external:pull`, `setup:external:start`, `setup:external:stop`, `setup:external:update` |
| `doppler`        | `setup:secrets:doppler`                                                                                               |
| `flutter`        | `setup:deps:flutter`, `code:format:dart`, `code:lint:dart`                                                            |
| `fnox`           | `setup:secrets:fnox`                                                                                                  |
| `kotlin`         | `setup:deps:kotlin`                                                                                                   |
| `node`           | `setup:deps:node`                                                                                                     |
| `python`         | `setup:deps:python`                                                                                                   |
| `swift`          | `setup:deps:swift`                                                                                                    |
| `vscode`         | `setup:vscode`                                                                                                        |

### The libraries in `_scripts/`

| Library       | Contents                                                                            |
| ------------- | ----------------------------------------------------------------------------------- |
| `helpers`     | The print functions, `has_own_task`, `own_tasks`, `members`, `shell_files_in_scope` |
| `checks`      | git predicates. They return an exit status only.                                    |
| `merge`       | `merge_to_destination_branch`, for the `code:merge` tasks                           |
| `placeholder` | For a task that no stack fills yet. It always exits 0.                              |
| `repo-sync`   | The sync function `repo_sync`                                                       |

### The rules for task scripts

- Each tool call has the prefix `mise x --`. System commands and `mise` itself
  do not.
- `mise x -- which <tool>` finds a tool, not `command -v`.
- The shell is for bash 3.2 and bash 5.x.
- `#MISE` and `#USAGE` header comments give the metadata and the flags.
- A task that does not apply prints nothing and exits 0.
- The output rules are in "Task output: headers and subheaders" of
  `CLAUDE.md`.

## Shell aliases

| Alias               | Command                        | Source                                                |
| ------------------- | ------------------------------ | ----------------------------------------------------- |
| `setup`             | `mise run setup`               | `tools/mise/tools.toml`                               |
| `precommit`         | `mise run code:precommit`      | `tools/pre-commit/tools.toml`                         |
| `setup-all`         | `mise run setup --all`         | The sync, only with git submodules, in `conf.d/mise/` |
| `setup-<submodule>` | `mise run setup --<submodule>` | The sync, one for each git submodule                  |
| `npx`               | `pnpm dlx`                     | `tools/node/tools.toml`                               |

All aliases are in `dev` only, because `shell_alias` is a dev-only section.

## Observations

These facts came from the examination. They are not decisions.

- No tool ships a task that runs, tests or builds the application.
- The mise tool `flutter` has the static mise `depends` `android-sdk`. Thus
  each Flutter consumer gets `android` and `java`. See [tools.md](tools.md).
- Gradle does not come from mise. The Gradle tasks use the `gradlew` wrapper
  of the project.
- `fnox` and `doppler` install their CLI in `dev` only. See
  [fnox.md](fnox.md) for the new fnox design.
- `tools/mise/.config/mise/tasks/code/_default` has the shebang
  `#!/usr/bin/env zsh`. The other task scripts use bash.
- No file of `tools/` sets `MISE_ENV`. The session of this examination had
  `MISE_ENV=dev`, but its source is not known. It is not in
  `~/.config/mise/` or in `~/.config/fish/`.
