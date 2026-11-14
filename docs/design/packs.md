# Packs design

This document describes the packs model of both sides of `bootstrap`. A pack is
outside `tools.toml`, so it has its own document.

See also: [tools.md](tools.md), [repo.md](repo.md), [mac.md](mac.md) and
[mise.md](mise.md). The exit code of each
rule that stops the sync is in
[cli.md](cli.md#the-exit-codes).

A pack is a named group of members, and a shortcut of `add pack`. It is only
a pointer: each key of its table is a member. The config, the tasks and the
files are in the tools that the pack adds. A tool adds only the tools that it
needs, with its `depends`. If several packs share config, tasks or files, they
are in one tool that each of these packs names as a member, or in a common
tool `_<name>`.

The values file keeps no record of a pack. `bootstrap add pack <pack>` (or
`bootstrap mac add pack <pack>`) writes a `[tools.<name>]` table for each
member, and the sync reads only these tables. Thus a later change of a pack
other than `base` reaches no values file.

## The packs files

Each side has one file. It is not in a tool folder, so the sync does not copy
it. The file must exist, and it must have at least `[pack.base]`. Else the sync
stops with code 3: `base is not a bootstrap pack (named in <side>/packs.toml)`.

| Side    | Packs file         | Tools            | Values file              |
| ------- | ------------------ | ---------------- | ------------------------ |
| `repo`  | `repo/packs.toml`  | `repo/tools/*/`  | `.config/bootstrap.toml` |
| `macos` | `macos/packs.toml` | `macos/tools/*/` | `macos.toml`             |

Each pack is a `[pack.<name>]` table:

```toml
# macos/packs.toml
[pack.iwork]
description = "The iWork apps: Keynote, Numbers and Pages."
keynote = {}
numbers = {}
pages = {}
```

- Each key of the table is a member, and each member is a default: `add pack`
  adds it. The member form is `<tool> = {}`. In `[pack.base]` only, a tool can
  also be `<tool> = { required = true }`.
- A member is a tool folder, or a nested pack with the prefix `pack:`, in
  quotes, for example `"pack:iwork" = {}`. `add pack` expands it in turn.
- A tool can be in several packs, or in no pack.
- `description` is the one reserved key of a `[pack.<name>]` table: a string
  value at the key `description` is the text of the pack, not a member, for
  example `description = "The AI tools."`. A `description` that is not a
  string stops the sync with code 2. A string at another key also stops the
  sync. A pack has no `summary`. The one key of a member is `required`, a
  boolean. Each member is on one line.
- `required = true` marks a tool that each consumer must have. It is valid only
  for a tool of `base`, not for a nested pack. Else the sync stops.
  `bootstrap remove tool <tool…>` refuses a required tool.
- The repo sync order is the required members of `base` in the order of the
  packs file, then the tools of the `[tools.*]` tables in name order, each
  dependency before its tool. The Mac side sorts the tools by name in each
  stage.
- The TOML plugin of dprint sorts keys, so the root `dprint.jsonc` excludes
  `repo/packs.toml` and `macos/packs.toml`. Keep the order by hand.
- The old keys stop the sync: `depends`, `required`, `tools` or `default` with
  a list as a pack key (not as a member key), and a member with the prefix
  `p:`. The key `default` in a member is the old form, and it stops the sync. A
  member that is not a table or is not on one line, and another key in a
  member, also stop the sync.
- A member that is not a tool folder or a pack stops the sync. The
  message names the pack.
- A pack name must not be the name of a tool folder. The sync stops for a
  duplicate.
- A pack must not include itself, directly or through other packs. The sync
  stops for a loop, and it shows the packs of the loop, for example
  `a -> b -> a`.
- Each sync examines every `[pack.*]` table. Thus a problem shows before a
  consumer adds that pack.

## The questions

The packs files hold only the relationships. The questions of a value are in
the `[questions]` tables of the tools: `bootstrap init` and `bootstrap add` ask
them. See [cli.md](cli.md#the-questions). The questions of a
selection, for example which pack a repository needs, are in
`docs/sync/packs/<pack>.md` and `docs/sync/tools/<tool>.md`. The adopting AI
asks them, and it does not ask a question that the prompt of the user already
answers.

## The values file holds tools

The values file has one `[tools.<name>]` table for each selected tool, also
for each required member of `base` and each dependency, so that a user can
write values in it. It has no `[packs.<name>]` table: that is an old form. See
[bootstrap-config.md](bootstrap-config.md#old-forms).

```toml
format = 2

[tools.dprint]

[tools.github]
env = [ "dev" ]
```

| Rule                               | `repo`                        | `macos`              |
| ---------------------------------- | ----------------------------- | -------------------- |
| `env` of a `[tools.<name>]` entry  | The environments of the tool. | It stops the sync.   |
| Each other key of `[tools.<name>]` | A value of the tool.          | A value of the tool. |

The values are in [bootstrap-config.md](bootstrap-config.md#toolsfolderkey) and
in [mac.md](mac.md#macostoml).

A tool is selected when it has a table, when a selected tool needs it
(`depends`), or when it is a required member of `base`. On each sync, the
front-end reads its packs file. It adds the required members of `base`, then
the tools of `[tools.*]`. Then it resolves the dependencies of each tool. The
sync writes an empty table for each required member of `base` that has none.
See [cli.md](cli.md#the-tables-of-the-required-members).

## Environments

Only the repo side has environments. A pack has no environments. Each tool gets
the `env` of its `[tools.<name>]` table, else its own `default_envs`. Each
dependency gets the environments of the tool that needs it. `[shell_alias]` is
only in `dev`. See
[mise.md](mise.md#the-environments-of-a-mise-tool).

## The `base` pack

`base` is a pack, not a tool. Each side must have it. Only `bootstrap` changes
the tools of the `base` pack, in the packs file.

- A new values file gets a table for each member of `base`: `init` adds
  `base` as `add pack base` does. A later `init` does not add a removed member
  again.
- The required members of `base` are selected also without a table, and
  `bootstrap remove tool <tool…>` cannot remove them. The other members of
  `base` are plain tables, and the user can remove them.
- On the repo side, an `env` in the table of a required tool can only add
  environments, and the sync examines it.
- The `mise` tool of each side has `required = true`, because it ships the
  entry of the sync. The sync does not name these tools.

<!-- dprint-ignore -->
```toml
# repo/packs.toml
[pack.base]
description = "The tools of each repository: ..."
mise = { required = true }
git = { required = true }
pre-commit = { required = true }
dprint = { required = true }
linter = { required = true }
gitleaks = { required = true }
grype = { required = true }
osv-scanner = { required = true }
vscode = {}
```

In a git submodule, the repo side skips a `root_only` tool of a pack. See
[repo.md](repo.md#root_only).

The Mac `base` pack is in `macos/packs.toml`. Its required members are `mise`,
`homebrew`, the shells, `git`, `ssh`, `cli`, `python`, `pre-commit` and the
other tools of each Mac. `visual-studio-code` and `uv` are members that a Mac
can remove.

### Common tools

A pack can have a common tool. It holds the items that all the other tools of
the pack use, for example an orchestrator task and its hook. The name of a
common tool starts with `_`, for example `_security_base`.

A common tool is not a member of the pack. It is in the `depends` of each
member tool. For example, `gitleaks`, `grype` and `osv-scanner` each have
`depends = [ "_security_base" ]`.

- Each `[tools.<member>]` entry brings the common tool, with the environments
  of the member tools that need it.
- The values file does not select it, and it has `selectable = false`, so
  `add pack` writes no table for it. A `[tools.<name>]` table for it gives
  values.
- The sync adds a dependency before the tool that needs it. Thus the tool order
  of the `base` pack is `... linter _security_base gitleaks grype
  osv-scanner`.
  Of these four tools, only `_security_base` contributes a hook.

### The tools of the repo `base` pack

| Tool             | Files and config                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mise`           | `.config/mise.toml`, the `zz-repo` seeds, the libraries `helpers`, `checks`, `placeholder` and `worktrees-page`, each orchestrator `_default`, `setup:configs`, `code:lint:60-bootstrap`, `code:merge` (with one of `main` and `develop`), `setup:mise`, `setup:worktree`, `code:worktrees`, the `init` task, the `setup`, `precommit` and `worktrees` aliases, the `os` settings of each environment. The mise tools `jq` and `yq` (dev, ci), and `github:virajp/bootstrap` (dev). The derive script: the member aliases |
| `git`            | `.gitignore` (shared), `.gitattributes`, `git-conventional-commits.yaml`, `.config/wt.toml`, `_scripts/merge`, `code:merge:develop` and `code:merge:main` (with both branches), `code:git-config`. The hooks `git-config`, `no-commit-to-branch` and `conventional-commits`. The mise tool `worktrunk`. The derive and check scripts of the branch model                                                                                                                                                                  |
| `pre-commit`     | `.config/pre-commit-config.yaml` (shared), `setup:precommit`, `code:precommit`, the mise tool `pre-commit`, `PRE_COMMIT_HOME`, the `precommit` alias. The file-check hooks from `pre-commit/pre-commit-hooks`                                                                                                                                                                                                                                                                                                             |
| `dprint`         | `dprint.jsonc` (shared), the mise tool `dprint`, `code:format:05-dprint`, `setup:dprint`, the `format` hook, the `[formatter]` of the sync                                                                                                                                                                                                                                                                                                                                                                                |
| `linter`         | The mise tools `npm:@virajp.dev/linter` and `shellcheck`, `code:lint:05-linter`, `setup:linter`, the `lint` hook                                                                                                                                                                                                                                                                                                                                                                                                          |
| `vscode`         | `.vscode/settings.json`, `.vscode/extensions.json` (shared), `setup:vscode`, the `code` alias. Root only                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `gitleaks`       | The mise tool `gitleaks`, `.config/gitleaks.toml` (a seed), `code:sec:10-gitleaks`                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `grype`          | The mise tool `grype`, `.config/grype.yaml` (a seed), `code:sec:20-grype`                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `_security_base` | The `sec` hook, which runs `code:sec`. The `mise` tool ships `code:sec`                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `osv-scanner`    | The mise tool `osv-scanner`, `.config/osv-scanner.toml` (a seed), `code:sec:30-osv-scanner`                                                                                                                                                                                                                                                                                                                                                                                                                               |

Each hook is a contribution of the tool that owns its task. Thus a removed tool
also removes its hook, and no hook calls a task that is missing.

### Changes to other tools

- `claude` has `root_only = false`. `setup:ai` skips in a git submodule.

## The Flutter packs

These packs are a plan, not in `repo/packs.toml`. The tools of each Flutter pack
are
not decided.

| Pack              | Use                   |
| ----------------- | --------------------- |
| `flutter-ios`     | Develop iOS apps.     |
| `flutter-android` | Develop Android apps. |
| `flutter-web`     | Develop web apps.     |
| `flutter-macos`   | Develop macOS apps.   |
| `flutter-win`     | Develop Windows apps. |
| `flutter-linux`   | Develop Linux apps.   |

## Generated folders

- There is no generated folder `.config/mise/conf.d/_base`. Each tool of the
  pack
  gets its own folder `.config/mise/conf.d/<tool>/`. The consumer does not edit
  these
  folders. See [mise.md](mise.md#zz-repo-and-zz-macos) for the folder of the
  owner.
