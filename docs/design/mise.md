# mise design

This document describes the way mise will be used across. It is the base tool
which is required in any repo and is expected to be already installed in the
machine before setting up with `bootstrap`

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
    │   ├── _base/mise.<env>.toml   # generated from tools/base/tools.toml
    │   ├── <tool>/mise.<env>.toml  # generated from tools/<tool>/tools.toml
    │   └── _repo/mise.<env>.toml   # seeds; the repository owns them
    └── tasks/
        ├── _scripts/               # libraries, not tasks
        ├── code/                   # code:* tasks
        └── setup/                  # setup:* tasks
```

| File                                         | Kind      | Owner          | Loads when                  |
| -------------------------------------------- | --------- | -------------- | --------------------------- |
| `.config/mise.toml`                          | Template  | `base`         | Always                      |
| `.config/mise/conf.d/_base/mise.<env>.toml`  | Generated | `base`         | `MISE_ENV` includes `<env>` |
| `.config/mise/conf.d/<tool>/mise.<env>.toml` | Generated | `<tool>`       | `MISE_ENV` includes `<env>` |
| `.config/mise/conf.d/_repo/mise.<env>.toml`  | Seed      | The repository | `MISE_ENV` includes `<env>` |

Without `MISE_ENV`, no file in `conf.d/` loads. No tool ships or generates a
`conf.d/*/mise.toml`.

## `.config/mise.toml`

The template is `tools/base/.config/mise.toml.tmpl`. It holds the shared
`[settings]` and the `[env]` values from the values file.

| `[settings]` key                     | Value          |
| ------------------------------------ | -------------- |
| `activate_aggressive`                | `true`         |
| `all_compile`                        | `false`        |
| `env_shell_expand`                   | `true`         |
| `gpg_verify`                         | `true`         |
| `lockfile`                           | `false`        |
| `minimum_release_age`                | `"10h"`        |
| `os`                                 | `"macos"`      |
| `raw`                                | `true`         |
| `status.missing_tools`               | `"always"`     |
| `task.disable_spec_from_run_scripts` | `true`         |
| `task.output`                        | `"interleave"` |
| `task.timings`                       | `true`         |

| `[env]` key           | Source in `.config/bootstrap.toml`                     |
| --------------------- | ------------------------------------------------------ |
| `MEMBERS`             | `.members`: the git submodules, then each `[member.*]` |
| `MERGE_MODEL_DEVELOP` | `merge_model_develop`                                  |
| `MERGE_MODEL_MAIN`    | `merge_model_main`                                     |
| `REPO_NAME`           | `repo_name`                                            |

`min_version`, `monorepo_root` and `[monorepo]` go only in the own copy of a
repository. The repository puts its own `[env]` values in the `_repo` seeds.

## Environments

The consumer selects its environments with `environments` in the values file.

| Environment | Rule                     | Use                                       |
| ----------- | ------------------------ | ----------------------------------------- |
| `dev`       | Necessary                | Local development                         |
| `ci`        | Optional                 | The CI build and the examination of code  |
| `prod`      | Optional, only with `ci` | Only the tools that run the application   |
| `stage`     | Optional, only with `ci` | The tools of `prod`, other `[env]` values |

Without the key, the sync uses `dev` only. The sync stops with code 2 for an
unknown name, for a list without `dev`, or for `prod` or `stage` without `ci`.

`.config/mise.toml` sets `os = "macos"`. `base` sets `os = "linux"` in `ci`,
`prod` and `stage` with `[mise_env.<env>.settings]` tables.

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

| Key       | Meaning                                                                   |
| --------- | ------------------------------------------------------------------------- |
| `envs`    | The environments of the mise tool. The default is `["dev"]`. Not written. |
| `depends` | The mise tools that this mise tool needs. mise reads it too, so it stays. |

### The environments of a mise tool

The sync calculates the environments of each mise tool in this sequence:

1. A mise tool of a tool that the values file lists gets the environments of
   its `envs`.
2. A mise tool of a tool that only a dependency adds gets `dev` only.
3. Each mise `depends` gets the environments of the mise tool that needs it.
   The sync repeats this step until nothing changes.
4. The sync keeps only the environments that the values file selects.

Thus a consumer lists its core language in `tools`. For example, a TypeScript
repository lists `node`, so that `node` goes to each selected environment.

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
`.config/mise/conf.d/<folder>/mise.<env>.toml`. The folder is `_base` for
`base`, else the tool name. Each file starts with the bootstrap header.

- If a tool ships a `conf.d/*/mise.<env>.toml` file and also generates it, the
  sync stops for that tool.
- The sync skips a shipped `conf.d/*/mise.<env>.toml` of an environment that
  the consumer does not select.
- After the sync, it removes each `conf.d/*/mise*.toml` with the bootstrap
  header that it did not write. A file without the header stays.

## The current mise tools

| Tool         | mise tool              | Version  | `envs`               | mise `depends` | Other `[mise.*]` sections                           |
| ------------ | ---------------------- | -------- | -------------------- | -------------- | --------------------------------------------------- |
| `android`    | `android-sdk`          | `23`     | dev, ci              | `java`         |                                                     |
| `base`       | `gitleaks`             | `latest` | dev, ci              |                | `env`, `shell_alias`, `tasks`; `mise_env` (`os`)    |
| `base`       | `grype`                | `latest` | dev, ci              |                |                                                     |
| `base`       | `osv-scanner`          | `latest` | dev, ci              |                |                                                     |
| `base`       | `dprint`               | `latest` | dev, ci              |                |                                                     |
| `base`       | `npm:@askviraj/linter` | `latest` | dev, ci              | `node`, `pnpm` |                                                     |
| `base`       | `shellcheck`           | `latest` | dev, ci              |                |                                                     |
| `base`       | `gomplate`             | `latest` | dev                  |                |                                                     |
| `base`       | `jq`                   | `latest` | dev                  |                |                                                     |
| `base`       | `yq`                   | `latest` | dev                  |                |                                                     |
| `base`       | `pre-commit`           | `latest` | dev                  | `python`, `uv` |                                                     |
| `bun`        | `bun`                  | `latest` | dev, ci, prod, stage |                |                                                     |
| `claude`     | `pipx:graphifyy`       | `latest` | dev                  | `python`, `uv` |                                                     |
| `claude`     | `pipx:mempalace`       | `latest` | dev                  | `python`, `uv` |                                                     |
| `cloudflare` | `npm:cf`               | `latest` | dev, ci              | `node`, `pnpm` |                                                     |
| `doppler`    | `doppler`              | `latest` | dev                  |                |                                                     |
| `flutter`    | `flutter`              | `3`      | dev, ci              | `android-sdk`  |                                                     |
| `fnox`       | `fnox`                 | `latest` | dev                  |                | `env` (`FNOX_CONFIG_DIR`)                           |
| `gcloud`     | `gcloud`               | `latest` | dev, ci              | `python`, `uv` | `env` (`CLOUDSDK_PYTHON`)                           |
| `github`     | `gh`                   | `latest` | dev, ci              |                |                                                     |
| `github`     | `actionlint`           | `latest` | dev, ci              |                |                                                     |
| `gitlab`     | `glab`                 | `latest` | dev, ci              |                |                                                     |
| `java`       | `java`                 | `lts`    | dev, ci              |                |                                                     |
| `kotlin`     | `http:kotlin-lsp`      | `latest` | dev                  |                | Platform URLs for macOS                             |
| `node`       | `node`                 | `lts`    | dev, ci, prod, stage |                | `settings`, `env` (`_.path`), `shell_alias` (`npx`) |
| `node`       | `pnpm`                 | `latest` | dev, ci, prod, stage | `node`         |                                                     |
| `pulumi`     | `pulumi`               | `latest` | dev                  |                |                                                     |
| `python`     | `python`               | `3.14`   | dev, ci, prod, stage |                | `settings`, `env` (`UV_*`)                          |
| `python`     | `uv`                   | `latest` | dev, ci, prod, stage | `python`       |                                                     |
| `rust`       | `core:rust`            | `latest` | dev, ci              |                | `profile = "default"`                               |
| `rust`       | `zola`                 | `latest` | dev, ci              |                |                                                     |

These tools have no mise tools: `docker`, `swift` (the host Xcode), and
`vscode`. `kotlin` gets `java` through its `depends` key.

Because `base` needs `node`, `pnpm`, `python` and `uv`, each consumer gets the
`node` and `python` tools. If the consumer does not list them in `tools`, they
are in `dev` and in the environments of the `base` mise tools that need them
(`ci` for `npm:@askviraj/linter`).

## Tasks

### Names and files

mise uses the path below `.config/mise/tasks/` as the task name. For example,
`code/merge/develop` is `code:merge:develop`. A `_default` file is the task of
its folder. A `.tmpl` task is rendered by the sync, and the output keeps the
git mode of the template.

Only the `init` task is in mise config, in `[mise.tasks.init]` of
`tools/base/tools.toml`. It makes the task scripts executable. The task files
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

`has_own_task` finds an optional task, for example `setup:ai` or
`setup:external:update`.

### The current tasks of each tool

| Tool      | Tasks                                                                                                                                                                                                                                                                                                              |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `base`    | `setup`, `setup:mise`, `setup:configs`, `setup:configs:sync`, `setup:dprint`, `setup:linter`, `setup:precommit`, `setup:worktree`, `code`, `code:format`, `code:lint`, `code:sec`, `code:precommit`, `code:git-config`, `code:graph`, `code:worktrees`, `code:merge` or `code:merge:develop` and `code:merge:main` |
| `android` | `setup:deps:android`, `code:lint:android`                                                                                                                                                                                                                                                                          |
| `claude`  | `setup:ai`                                                                                                                                                                                                                                                                                                         |
| `docker`  | `setup:external:check`, `setup:external:pull`, `setup:external:start`, `setup:external:stop`, `setup:external:update`                                                                                                                                                                                              |
| `doppler` | `setup:secrets:doppler`                                                                                                                                                                                                                                                                                            |
| `flutter` | `setup:deps:flutter`, `code:format:dart`, `code:lint:dart`                                                                                                                                                                                                                                                         |
| `fnox`    | `setup:secrets:fnox`                                                                                                                                                                                                                                                                                               |
| `kotlin`  | `setup:deps:kotlin`                                                                                                                                                                                                                                                                                                |
| `node`    | `setup:deps:node`                                                                                                                                                                                                                                                                                                  |
| `python`  | `setup:deps:python`                                                                                                                                                                                                                                                                                                |
| `swift`   | `setup:deps:swift`                                                                                                                                                                                                                                                                                                 |
| `vscode`  | `setup:vscode`                                                                                                                                                                                                                                                                                                     |

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

| Alias               | Command                        | Source                               |
| ------------------- | ------------------------------ | ------------------------------------ |
| `setup`             | `mise run setup`               | `tools/base/tools.toml`              |
| `precommit`         | `mise run code:precommit`      | `tools/base/tools.toml`              |
| `setup-all`         | `mise run setup --all`         | The sync, only with git submodules   |
| `setup-<submodule>` | `mise run setup --<submodule>` | The sync, one for each git submodule |
| `npx`               | `pnpm dlx`                     | `tools/node/tools.toml`              |

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
- `tools/base/.config/mise/tasks/code/_default` has the shebang
  `#!/usr/bin/env zsh`. The other task scripts use bash.
- No file of `tools/` sets `MISE_ENV`. The session of this examination had
  `MISE_ENV=dev`, but its source is not known. It is not in
  `~/.config/mise/` or in `~/.config/fish/`.
