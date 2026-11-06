# mise design

This document describes the way mise will be used across all projects. It is the
base tool which is required in all repos and is expected to be already installed
in the machine before setting up with `bootstrap`. It holds the mise generator
of
both sides, the environments, `[overrides.mise*]`, `zz-repo` and `zz-macos`, and
the tasks.

See also: [tools.md](tools.md), [packs.md](packs.md),
[sync-engine.md](sync-engine.md) for the stale rule,
[repo-sync.md](repo-sync.md), [setup-sync.md](setup-sync.md) and
[bootstrap-config.md](bootstrap-config.md). The exit code of each rule that
stops the sync is in
[sync-engine.md](sync-engine.md#the-exit-codes).

## Rules

- In a repository, `mise` is ONLY used in `dev` & `ci` environments. The Mac
  side
  has no environments
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

A tool never ships a mise config file, except the two `mise` tools (see
[The `[mise.*]` sections](#the-mise-sections)). It writes its mise config in the
`[mise.*]` tables of its `tools.toml`, in mise syntax. The sync generates the
mise files from these tables.

## The files in a consumer repository

```text
.config/
├── bootstrap.toml                  # the values file; the sync reads it
├── mise.toml                       # synced; the only file for every environment
└── mise/
    ├── conf.d/
    │   ├── <tool>/mise.<env>.toml  # generated from repo/tools/<tool>/tools.toml
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

Each file in `.config/mise/conf.d/` is for one environment: `mise.<env>.toml`.
mise
loads it only when `MISE_ENV` includes that environment. Without `MISE_ENV`, no
file
in `conf.d/` loads. In a repository, no tool ships or generates a
`conf.d/*/mise.toml`. Only the `macos` side of the generator writes that name.
See
[The two sides](#the-two-sides-of-the-generator).

The folder that the owner customizes is in
[`zz-repo` and `zz-macos`](#zz-repo-and-zz-macos).

## Environments

The consumer selects its environments with `environments` in the values file.

| Environment | Rule      | Use                                      |
| ----------- | --------- | ---------------------------------------- |
| `dev`       | Necessary | Local development. Each tool is in `dev` |
| `ci`        | Optional  | The CI build and the examination of code |

Any other name stops the sync. `bootstrap` is only for `dev` and
`ci`: a
`tools.toml` names only these two in `default_envs` and `envs`, and any other
name
stops the sync. The sync ignores each environment that
`environments`
does not select. Without `environments`, the sync uses `dev` only and logs a
warning.

## `.config/mise.toml`

`min_version`: The minimum version of the mise cli required by the
consumer repository. It is not set now. When it is set, it must be continously
upgraded as new versions of the mise configuration are released.

The template is `repo/tools/mise/.config/mise.toml.tmpl`. It is the only file
for every environment.
It holds the shared `[settings]` and the `[env]` values `REPO_NAME` and
`MEMBERS` (the
folders of the members, see [repo-sync.md](repo-sync.md#the-template-context)).

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

A tool does not ship mise files. The exceptions are the `mise` tool of
`repo/tools/`, with `.config/mise.toml.tmpl` and the `zz-repo` seeds, and the
`mise` tool of `macos/tools/`, with the user-level
`~/.config/mise/config.toml` and `~/.config/mise/miserc.toml`. `config.toml` has
the top-level key `min_version`, which no `[mise.*]` section holds.

A tool writes its mise config in `[mise.*]`
tables of its `tools.toml`, in mise syntax. The sync removes the prefix
`mise.` and writes each table into `.config/mise/conf.d/<tool>/mise.<env>.toml`
as a normal mise section: `[mise.settings]` becomes `[settings]`, `[mise.env]`
becomes `[env]`, and `[mise.shell_alias]` becomes `[shell_alias]`. The
generator is the library `sync-engine/mise`: `_SE_MISE_TEMPLATE` builds the
output, and `_se_mise_render` writes it. `_se_mise_envs` finds the environments
of each tool, and `_se_mise_overrides` examines `[overrides.mise*]`. The
template reads the `SYNC_ENGINE_MISE_*` environment variables.

These parts stay in `sync-engine/repo-sync`:

- `[overrides.gitignore]` and its checks.
- The `setup-all` and `setup-<submodule>` aliases.
- The records of the generated files, for the file loop and the manifest.
- `_rs_envs`, which reads `environments`, and `_rs_env_rules`, which examines
  the `env` of each entry of the values file.
- `_rs_mise_index`, which reads each `tools.toml` of the clone, and
  `_rs_mise_owner`, which finds the tool that defines a mise tool.
- The call of the library with the side `repo`.

The sync sorts each `[mise.<section>]` table into one of three classes:

| Class    | Sections                                                                                                                                                                     | Goes to                                      |
| -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| Entry    | `tools`, `tool_alias`, `plugins`                                                                                                                                             | The environments of each entry               |
| Folder   | `settings`, `env`, `tool_config`, `task_config`, `task_templates`, `doctor`, `deps`, `oci`                                                                                   | Each environment where the folder has a tool |
| Dev only | `vars`, `secrets`, `redactions`, `shell_alias`, `hooks`, `watch_files`, `daemons`, `daemon_groups`, `daemons_settings`, `daemon_providers`, `tasks`, `bootstrap`, `dotfiles` | `dev` only                                   |

`[mise.bootstrap.*]` and `[mise.dotfiles]` are dev only, like `[tasks]`, because
CI is not a machine to provision.

A `[mise_env.<env>.<section>]` table adds content to one environment only. Its
section must be in one of the three classes. The sync writes it only when the
tool is in that environment. If `[mise.<section>]` has the same section, the
sync merges the two, and `[mise_env.<env>.<section>]` wins on a key that is in
both. It can add a dev-only section. `[mise_env.<env>.tools]` holds a mise tool
that
is only in that environment, and its tool gets no `envs` check for it. For
example,
`kotlin-lsp` of `kotlin` is in `[mise_env.dev.tools]`, so `kotlin` can be in
`ci`.

Any other section stops the mise config of that tool. The sync keeps
the
existing `conf.d` files of that tool, and it still syncs the other files of the
tool.

#### An example

The `node` tool is in `dev` and `ci`. Its `tools.toml` has:

```toml
[mise.settings] # folder class: dev and ci
node.compile = false
npm.package_manager = "aube"

[mise.env] # folder class: dev and ci
_.path = { path = "node_modules/.bin", tools = true }
```

The `pnpm` tool is also in `dev` and `ci`. Its `tools.toml` has:

```toml
[mise.shell_alias] # dev only: an alias is for a person at a terminal
npx = "pnpm dlx"
```

So `conf.d/node/mise.dev.toml` and `conf.d/node/mise.ci.toml` each get
`[settings]` and `[env]`. `conf.d/pnpm/mise.dev.toml` gets `[shell_alias]`, and
`conf.d/pnpm/mise.ci.toml` does not get it.

The `mise` tool selects the binaries of the host with one table for each
environment:

```toml
[mise_env.dev.settings]
os = "macos" # a developer works on macOS

[mise_env.ci.settings]
os = "linux" # CI runs on Linux
```

So `conf.d/mise/mise.dev.toml` gets `os = "macos"`, and
`conf.d/mise/mise.ci.toml` gets `os = "linux"`.

mise loads a `conf.d/*/mise.<env>.toml` only when `MISE_ENV` includes that
environment. It loads the folders in name order, and `conf.d/zz-repo/` loads
last. Thus a value in the repository's own `zz-repo/mise.<env>.toml` wins over
the value of each tool.

### `<% %>` in a `[mise.*]` string

A `[mise.*]` string can hold `<% %>`. When the `tools.toml` of a tool contains
`<%`, the engine renders that file as a gomplate template into a temp copy. The
generator reads the copy, so the generated file holds the rendered value. The
overrides of the values file are not rendered.

The render context is the same on both sides:

| Key       | Value                            |
| --------- | -------------------------------- |
| `.config` | The values of the tool           |
| `.all`    | The values of each selected tool |
| `.env`    | `HOME` and `USER`                |

An expression sits inside a TOML basic string (`"..."`), and each string literal
inside the expression is a Go raw string in backticks:

```toml
[mise.env]
GIT_NAME = "<% required `git needs a name` .config.name %>"
```

A key that is missing reaches `required` as a nil, and `required` fails the
render. A plain `.config.x` writes the text `<no value>`, and the engine rejects
a copy that holds this text. A function that does not accept a nil, for example
`len`, also fails the render. Each other use of the key renders without an
error:
``.config.x | default `d` `` gives `d`, `if .config.x` takes the false branch,
and `print .config.x` writes `<nil>` in the value.

Each of these failures stops the mise config of that tool. The
existing `conf.d` files of the tool stay, and the other tools continue.
`<shown>` is `<tools root>/<tool>/tools.toml`, for example
`repo/tools/node/tools.toml`.

| Failure                                  | Message                                                                 |
| ---------------------------------------- | ----------------------------------------------------------------------- |
| A `required` or `test.Fail` check        | `<shown>: cannot render: <message>`                                     |
| An expression uses a key that is missing | `<shown>: cannot render: an expression has no value - a key is missing` |
| The rendered copy is not valid TOML      | `<shown>: not valid TOML after render: <message>`                       |
| The copy has other key paths             | `<shown>: a rendered value changes the keys of the file`                |

A raw file that is not valid TOML does not follow this rule. On the `repo` side,
`_rs_mise_index` reads each `tools.toml` of the clone before any render, so the
whole sync stops with `cannot read repo/tools/<tool>/tools.toml`. The
message `<shown>: not valid TOML before render: <message>` comes only from a
direct call of `_se_mise_render`.

The rendered copy must have the same key paths as the raw file, so a rendered
value cannot add a key or a table.

The values file is trusted. Its writer can already add mise config directly, so
the key-path check is the whole check. A rendered value that replaces a fixed
value and keeps the key paths equal is not a defect. A key that yq and gomplate
read differently is not a defect either.

### The two sides of the generator

Each function of `sync-engine/mise` takes a side as its first parameter: `repo`
or `macos`. Any other value is a usage error. The `repo` side writes one
file for each environment. The `macos` side writes one file for each tool and
refuses environments.

| Rule                            | `repo`                                                    | `macos`                                                   |
| ------------------------------- | --------------------------------------------------------- | --------------------------------------------------------- |
| Output                          | `conf.d/<tool>/mise.<env>.toml`, one for each environment | `conf.d/<tool>/mise.toml`, one for each tool with content |
| Dev-only sections               | Only in `dev`                                             | Written, because the Mac uses no environments             |
| `[mise_env.*]`                  | Valid                                                     | A manifest error                                          |
| `envs` of a mise tool           | Valid                                                     | A tool error                                              |
| `requires_found` of a mise tool | Valid                                                     | A tool error                                              |
| `[overrides.mise.*]`            | Valid                                                     | Valid. An override cannot add a key                       |
| `[overrides.mise_env.*]`        | Valid                                                     | A values-file error                                       |
| Override comment                | `# override: .config/bootstrap.toml`                      | `# override: macos.toml` (the file name only)             |

After a tool error, the files of that tool stay, the other tools continue, and
the run fails. A values-file error stops the whole run before any write. A
manifest error also stops the whole run before any write, for each tool of the
clone, selected or not. See [tools.md](tools.md#the-toolstoml-file).

### The own config of a values file

`_se_mise_own_config`, only on the `macos` side, generates the `[mise.*]` tables
of the values file into `conf.d/zz-macos/mise.toml`. It uses the same section
list. The header is:

```text
# Synced from bootstrap: macos.toml [mise]. Do not edit this file.
# The next sync replaces it, and a manual change is lost.
```

A `<% %>` string renders with `.all` and `.env`, and has no `.config`.
`[overrides.mise.*]` does not apply to this file. A values file without a
`[mise]` table gets no file.

### The keys of a mise tool

A `[mise.tools.<name>]` table has the mise keys of the tool, for example
`version`, and three keys for the sync:

| Key              | Meaning                                                                                                                              |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `envs`           | The environments that the mise tool allows. The default is `["dev"]`. Not written.                                                   |
| `depends`        | The mise tools that this mise tool needs. mise reads it too, so it stays.                                                            |
| `requires_found` | A list of file names. The sync writes the entry only when it finds each name at the root or down to 4 folders below it. Not written. |

A file that follows the same condition, for example the task of that mise tool,
states it in its own `[files]` table. See [tools.md](tools.md#the-files-table).
`requires_found` is only for `[mise.tools]`. The sync reads `envs` and
`requires_found` only there, and `requires_found`, `requires` or `files` in a
`[mise_env.<env>.tools]` entry stops the sync. A mise entry has no
list of
files. For example, `python` installs `ruff` only with a `pyproject.toml`, and
it writes
`code:lint:ruff` only with a `pyproject.toml`, because the table of that task
has
`requires_found = ["pyproject.toml"]`. The search is in
[repo-sync.md](repo-sync.md#requires_found).

### The environments of a mise tool

The sync calculates the environments of each tool in this sequence. Each
mise tool gets the environments of its tool.

1. A tool of a `[tools.<folder>]` entry gets its `env`, else the
   `default_envs` of its `tools.toml`. A tool of a `[packs.<name>]` entry gets
   the `env` of that entry, else its own `default_envs`. Each tool of the
   `base` pack gets its own `default_envs`. A tool that is selected more than
   once gets each of these environments. An `env` name that is not in
   `environments`, an empty `env`, or an `env` item with a space stops the sync.
2. Each dependency gets the environments of the tool that needs it: a tool of
   the `depends` key, and the tool of a mise `depends`. The sync repeats this
   step until nothing changes.
3. The sync keeps only the environments that the values file selects.
4. Each tool must be in `dev`. An `env` without `dev` stops the sync. A
   `default_envs` without `dev` also
   stops the sync.
   The sync examines the `default_envs` of each tool of the clone, also of a
   tool that the values file does not select, so the tests of `bootstrap`
   find it first.
5. The `envs` of each mise tool must include each environment of its tool. If
   not, the sync stops. The message names the cause: the
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
selected, and stops for a duplicate.

### The generated files

The sync writes one file for each selected environment that has content:
`.config/mise/conf.d/<tool>/mise.<env>.toml`. The folder is the name of the
tool. Each file starts with the bootstrap header. There is no `_base` folder,
because `base` is a pack, not a tool. See [packs.md](packs.md#the-base-pack).

- If a tool ships a `conf.d/*/mise.<env>.toml` file and also generates it, the
  sync stops for that tool.
- The sync skips a shipped `conf.d/*/mise.<env>.toml` of an environment that
  the consumer does not select.
- The stale rule removes a `conf.d/*/mise*.toml` with the bootstrap header that
  the manifest lists and this run does not write. Thus a removed tool,
  environment
  or folder disappears from consumers. See
  [sync-engine.md](sync-engine.md#the-manifest-and-the-stale-rule).

The header of a generated file names `<tools root>/<tool>/tools.toml`, for
example
`repo/tools/node/tools.toml`. Each mise tool comes from a bootstrap tool. A tool
that a
repository needs gets added to a tool in `repo/tools/`, or becomes a new tool,
for
example `doppler`.

### The overrides

`[overrides]` of the values file replaces the defaults of the tools.

- `[overrides.mise.<path>]` applies to each environment, and
  `[overrides.mise_env.<env>.<path>]` to one environment, where it wins.
  `<path>` is the
  path below `[mise]` or `[mise_env.<env>]` of a `tools.toml`, without the tool
  name.
- The sync applies it to each selected tool with a value at that path. A table
  merges
  key by key; any other value replaces the default. An override replaces only a
  key that a
  selected tool has: it cannot add a key.
- These stop the sync: a path that no selected tool has, a key of
  `[overrides]`
  other than `mise`, `mise_env` and `gitignore`, an environment outside
  `environments`, and
  the `envs`, the `depends`, the `requires`, the `requires_found` or the `files`
  of a mise
  tool.
- Each overridden line gets an `# override: <values file>` comment.
- The sync examines the overrides before any write. `repo_sync --check` examines
  their
  structure, without the check of each path.
- The overrides are not rendered.

The Mac side has only `[overrides.mise.*]`. See
[The two sides](#the-two-sides-of-the-generator). `[overrides.gitignore]` is
different: see
[repo-sync.md](repo-sync.md#overridesgitignore).

### `zz-repo` and `zz-macos`

mise loads the folders of `conf.d/` in name order, and a later folder wins.

| Folder     | Side    | Owner          | Files                                                                                                                   |
| ---------- | ------- | -------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `zz-repo`  | `repo`  | The repository | The `mise` tool ships the seeds `zz-repo/mise.<env>.toml`, for example for the own `[env]` values of the repository.    |
| `zz-macos` | `macos` | The user       | The sync generates `zz-macos/mise.toml` from `[mise.*]` of `macos.toml`. See [below](#the-own-config-of-a-values-file). |

Each name sorts after each tool folder, so the config of the owner wins over the
config of
each tool. A tool folder that has that name, or whose name sorts after it in
byte order,
stops the sync. The sync writes a seed only when it is missing.

## The mise tools

This table has the mise tools of `repo/tools/` now.

| Tool          | mise tool                | Version  | `envs`  | mise `depends` | Other `[mise.*]` sections                          |
| ------------- | ------------------------ | -------- | ------- | -------------- | -------------------------------------------------- |
| `android`     | `android-sdk`            | `23`     | dev, ci | `java`         |                                                    |
| `dprint`      | `dprint`                 | `latest` | dev, ci |                |                                                    |
| `git`         | `worktrunk`              | `latest` | dev     |                |                                                    |
| `gitleaks`    | `gitleaks`               | `latest` | dev, ci |                |                                                    |
| `grype`       | `grype`                  | `latest` | dev, ci |                |                                                    |
| `linter`      | `npm:@virajp.dev/linter` | `latest` | dev, ci | `node`         |                                                    |
| `linter`      | `shellcheck`             | `latest` | dev, ci |                |                                                    |
| `mise`        | `jq`                     | `latest` | dev, ci |                | `tasks` (`init`), `shell_alias`; `mise_env` (`os`) |
| `mise`        | `yq`                     | `latest` | dev, ci |                |                                                    |
| `osv-scanner` | `osv-scanner`            | `latest` | dev, ci |                |                                                    |
| `pre-commit`  | `pre-commit`             | `latest` | dev     | `python`, `uv` | `env` (`PRE_COMMIT_HOME`), `shell_alias`           |
| `repo-sync`   | `gomplate`               | `latest` | dev     |                |                                                    |
| `apko`        | `apko`                   | `latest` | dev, ci |                |                                                    |
| `aube`        | `aube`                   | `latest` | dev, ci | `node`         | `shell_alias` (`npx`)                              |
| `bun`         | `bun`                    | `latest` | dev, ci |                |                                                    |
| `cloudflare`  | `npm:cf`                 | `latest` | dev, ci | `node`         |                                                    |
| `cosign`      | `cosign`                 | `latest` | dev, ci |                |                                                    |
| `crane`       | `crane`                  | `latest` | dev, ci |                |                                                    |
| `doppler`     | `doppler`                | `latest` | dev     |                |                                                    |
| `firebase`    | `npm:firebase-tools`     | `latest` | dev, ci | `node`         |                                                    |
| `flutter`     | `flutter`                | `3`      | dev, ci | `android-sdk`  |                                                    |
| `fnox`        | `fnox`                   | `latest` | dev     |                | `env` (`FNOX_CONFIG_DIR`)                          |
| `gcloud`      | `gcloud`                 | `latest` | dev, ci | `python`, `uv` | `env` (`CLOUDSDK_PYTHON`)                          |
| `github`      | `gh`                     | `latest` | dev, ci |                |                                                    |
| `github`      | `actionlint`             | `latest` | dev, ci |                |                                                    |
| `gitlab`      | `glab`                   | `latest` | dev, ci |                |                                                    |
| `java`        | `java`                   | `lts`    | dev, ci |                | `settings`                                         |
| `kotlin`      | `ktlint`                 | `latest` | dev, ci |                |                                                    |
| `kotlin`      | `github:detekt/detekt`   | `latest` | dev, ci |                |                                                    |
| `kotlin`      | `http:kotlin-lsp`        | `latest` | dev     |                | In `[mise_env.dev.tools]`. Platform URLs for macOS |
| `melange`     | `melange`                | `latest` | dev, ci |                |                                                    |
| `node`        | `node`                   | `lts`    | dev, ci |                | `settings`, `env` (`_.path`)                       |
| `pnpm`        | `pnpm`                   | `latest` | dev, ci | `node`         | `shell_alias` (`npx`)                              |
| `pulumi`      | `pulumi`                 | `latest` | dev     |                |                                                    |
| `python`      | `python`                 | `3.14`   | dev, ci |                | `settings`, `env` (`UV_*`)                         |
| `python`      | `uv`                     | `latest` | dev, ci | `python`       |                                                    |
| `python`      | `ruff`                   | `latest` | dev, ci |                | `requires_found` (`pyproject.toml`)                |
| `python`      | `pipx:pyrefly`           | `latest` | dev, ci | `python`, `uv` | `requires_found` (`pyproject.toml`)                |
| `rust`        | `core:rust`              | `latest` | dev, ci |                | `profile = "default"`                              |
| `rust`        | `zola`                   | `latest` | dev, ci |                |                                                    |
| `swift`       | `swiftlint`              | `latest` | dev, ci |                | `mise_env` (`os` in `ci`)                          |

These tools have no mise tools: `_security_base`, `claude`, `docker` and
`vscode`.
`kotlin` gets `java` through its `depends` key. `vscode` has a `code` alias.

Because the `base` pack needs `node`, `python` and `uv`, each consumer gets the
`node` and `python` tools. They get the environments of the tools that need
them: `dev` and `ci` for `node` (the `linter` tool), and `dev` for `python` and
`uv` (the `pre-commit` tool). The `node` tool brings no package manager: a
repository with a `package.json` selects the `pnpm` or the `aube` tool.

## Tasks

### Names and files

mise uses the path below `.config/mise/tasks/` as the task name. For example,
`code/merge/develop` is `code:merge:develop`. A `_default` file is the task of
its folder. A `.tmpl` task is rendered by the sync, and the output keeps the
git mode of the template.

Only the `init` task is in mise config, in `[mise.tasks.init]` of
`repo/tools/mise/tools.toml`. It makes the task scripts executable. The task
files load in each environment.

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
| `repo-sync`      | `setup:configs`, `setup:configs:sync`, `code:lint:repo-sync`                                                          |
| `git`            | `code:git-config`, `code:merge` or `code:merge:develop` and `code:merge:main`                                         |
| `pre-commit`     | `setup:precommit`, `code:precommit`                                                                                   |
| `dprint`         | `code:format`, `setup:dprint`                                                                                         |
| `linter`         | `code:lint`, `setup:linter`                                                                                           |
| `_security_base` | `code:sec`                                                                                                            |
| `gitleaks`       | `code:sec:gitleaks`                                                                                                   |
| `grype`          | `code:sec:grype`                                                                                                      |
| `osv-scanner`    | `code:sec:osv-scanner`                                                                                                |
| `android`        | `setup:deps:android`, `code:lint:android`                                                                             |
| `claude`         | `setup:ai`                                                                                                            |
| `docker`         | `setup:external:check`, `setup:external:pull`, `setup:external:start`, `setup:external:stop`, `setup:external:update` |
| `doppler`        | `setup:secrets:doppler`                                                                                               |
| `flutter`        | `setup:deps:flutter`, `code:format:dart`, `code:lint:dart`                                                            |
| `fnox`           | `setup:secrets:fnox`                                                                                                  |
| `kotlin`         | `setup:deps:kotlin`, `code:format:ktlint`, `code:lint:detekt`                                                         |
| `node`           | `setup:deps:node`                                                                                                     |
| `python`         | `setup:deps:python`, `code:format:ruff`, `code:lint:ruff`, `code:lint:pyrefly`, `code:lint:uv`                        |
| `swift`          | `setup:deps:swift`, `code:format:swift-format`, `code:lint:swiftlint`                                                 |
| `vscode`         | `setup:vscode`                                                                                                        |

### The libraries in `_scripts/`

| Library          | Contents                                                                            |
| ---------------- | ----------------------------------------------------------------------------------- |
| `helpers`        | The print functions, `has_own_task`, `own_tasks`, `members`, `shell_files_in_scope` |
| `checks`         | git predicates. They return an exit status only.                                    |
| `merge`          | `merge_to_destination_branch`, for the `code:merge` tasks                           |
| `placeholder`    | For a task that no stack fills yet. It always exits 0.                              |
| `xcode`          | `require_xcode`, `find_xcodeproj` and the simulator functions, of the `swift` tool  |
| `worktrees-page` | `write_worktrees_page`, the HTML page of `code:worktrees`, of the `mise` tool       |

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
| `setup`             | `mise run setup`               | `repo/tools/mise/tools.toml`                          |
| `precommit`         | `mise run code:precommit`      | `repo/tools/mise/tools.toml` and `pre-commit`         |
| `worktrees`         | `mise run code:worktrees`      | `repo/tools/mise/tools.toml`                          |
| `setup-all`         | `mise run setup --all`         | The sync, only with git submodules, in `conf.d/mise/` |
| `setup-<submodule>` | `mise run setup --<submodule>` | The sync, one for each git submodule, also nested     |
| `npx`               | `pnpm dlx`                     | `repo/tools/pnpm/tools.toml`                          |
| `npx`               | `aubx`                         | `repo/tools/aube/tools.toml`                          |
| `code`              | `command code --profile ...`   | `repo/tools/vscode/tools.toml`                        |

All aliases are in `dev` only, because `shell_alias` is a dev-only section. The
sync
writes the aliases of `mise` and the submodule aliases in the generated
`mise/mise.dev.toml`.

## Observations

These facts came from the examination. They are not decisions.

- No tool ships a task that runs, tests or builds the application.
- The mise tool `flutter` has the static mise `depends` `android-sdk`. Thus
  each Flutter consumer gets `android` and `java`. See [tools.md](tools.md).
- Gradle does not come from mise. The Gradle tasks use the `gradlew` wrapper
  of the project.
- `fnox` and `doppler` install their CLI in `dev` only. See
  [fnox.md](fnox.md) for the new fnox design.
- `repo/tools/mise/.config/mise/tasks/code/_default` has the shebang
  `#!/usr/bin/env zsh`. The other task scripts use bash.
- No file of `repo/tools/` sets `MISE_ENV`. The session of this examination had
  `MISE_ENV=dev`, but its source is not known. It is not in
  `~/.config/mise/` or in `~/.config/fish/`.
