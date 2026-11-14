# Packs design

This document describes the packs model of both sides of `bootstrap`. A pack is
outside `tools.toml`, so it has its own document.

See also: [tools.md](tools.md), [repo.md](repo.md), [mac.md](mac.md) and
[mise.md](mise.md). The exit code of each
rule that stops the sync is in
[cli.md](cli.md#the-exit-codes).

A pack is a named group of members. It is only a pointer: each key of its
table is a member, and it has no other keys. The config, the tasks and the
files are in the tools that the pack selects. The pack is the group: a tool
adds only the tools that it needs, with its `depends`. If several packs share
config, tasks or files, they are in one tool that each of these packs names as
a member, or in a common tool `_<name>`.

## The packs files

Each side has one file. It is not in a tool folder, so the sync does not copy
it.

| Side    | Packs file         | Tools            | Values file              |
| ------- | ------------------ | ---------------- | ------------------------ |
| `repo`  | `repo/packs.toml`  | `repo/tools/*/`  | `.config/bootstrap.toml` |
| `macos` | `macos/packs.toml` | `macos/tools/*/` | `macos.toml`             |

Each pack is a `[pack.<name>]` table:

```toml
# repo/packs.toml
[pack.security]
gitleaks = { default = true }
grype = { default = true }
osv-scanner = { default = true }
```

```toml
# macos/packs.toml
[pack.ai]
claude-cli = {}
claude-desktop = { default = true }
```

- Each key of the table is a member. The member forms are
  `<tool> = { default = true }` and `<tool> = {}`.
- `[packs.<name>]` of a values file adds each `default = true` member. A member
  with `{}` adds nothing: it arrives only with `[tools.<name>]`.
- A member is a tool folder, or a nested pack with the prefix `pack:`, in
  quotes,
  for example `"pack:security" = { default = true }`. The sync expands it in
  turn.
  `"pack:<name>" = {}` does not expand: that pack arrives only through its own
  `[packs.<name>]`.
- A tool can be in several packs. The sync adds it once.
- `description` is the one reserved key of a `[pack.<name>]` table: a string
  value at the key `description` is the text of the pack, not a member, for
  example `description = "The AI tools."`. A `description` that is not a
  string stops the sync with code 2. A string at another key also stops the
  sync. A pack has no `summary`. The keys of a member are `default` and
  `required`, each a boolean. Each member is on one line.
- `required = true` marks a tool that each consumer must have. It is valid only
  for a member of `base` with `default = true`, for example
  `mise = { default = true, required = true }`. Else the sync stops.
  `bootstrap remove` refuses a required tool.
- The member order of a pack is the sync order of the repo side. The Mac side
  sorts the tools by name in each stage.
- The TOML plugin of dprint sorts keys, so the root `dprint.jsonc` excludes
  `repo/packs.toml` and `macos/packs.toml`. Keep the order by hand.
- The old keys stop the sync: `depends`, `required`, `tools` or
  `default` with a list as a pack key (not as a member key), and a member with
  the prefix `p:`. A
  member that is
  not a table or is not on one line, another key in a member, and a `default`
  that is not a boolean also stop the sync.
- A member that is not a tool folder or a pack stops the sync. The
  message names the pack.
- A pack name must not be the name of a tool folder. The sync stops for a
  duplicate.
- A pack must not include itself, directly or through other packs, with any
  `default`. The sync stops for a loop, and it shows the packs of
  the loop, for example `a -> b -> a`.
- Each sync examines every `[pack.*]` table, selected or not. Thus a problem
  shows before a consumer selects that pack.

## The questions

The packs files hold only the relationships. The questions of a value are in
the `[questions]` tables of the tools: `bootstrap init` and `bootstrap add` ask
them. See [cli.md](cli.md#the-questions). The questions of a
selection, for example which pack a repository needs, are in
`docs/sync/packs/<pack>.md` and `docs/sync/tools/<tool>.md`. The adopting AI
asks them, and it does not ask a question that the prompt of the user already
answers.

## The answers go into the values file

The values file has one `[packs.<name>]` table for each selected pack, next to
the
`[tools.<name>]` tables:

```toml
[packs.security]
env = [ "dev" ]

[tools.github]
```

| Rule                                 | `repo`                                  | `macos`                     |
| ------------------------------------ | --------------------------------------- | --------------------------- |
| The keys of a `[packs.<name>]` entry | Only `env`. Another key stops the sync. | None. A key stops the sync. |
| `env` of a `[tools.<name>]` entry    | The environments of the tool.           | It stops the sync.          |
| Each other key of `[tools.<name>]`   | A value of the tool.                    | A value of the tool.        |

The values are in [bootstrap-config.md](bootstrap-config.md#toolsfolderkey) and
in [mac.md](mac.md#macostoml). A list (`packs = [...]` or
`tools = [...]`) is the old format, and it stops the sync.

On each sync, the front-end reads its packs file. It adds the default members of
the `base` pack, then the default members of each selected pack, then the tools
of
`[tools.*]`. Then it resolves the dependencies of each tool. Thus, if
`bootstrap`
later adds a `default = true` member to a pack, each consumer of the pack gets
it
with no AI step. A member with `{}` arrives only with a `[tools.<name>]` entry.

A pack name that is not in the packs file stops the sync, as an
unknown
tool does.

The `[packs.*]` keys have the names without the prefix `pack:`, because each
name
in it is a pack.

## Environments

Only the repo side has environments. A tool that a pack adds gets the `env` of
the
`[packs.<name>]` entry, also through a nested pack, else its own `default_envs`.
Each tool of the `base` pack gets its own `default_envs`. A tool that is
selected
more than once gets each of these environments. Each dependency gets the
environments of the tool that needs it. `[shell_alias]` is only in `dev`. See
[mise.md](mise.md#the-environments-of-a-mise-tool).

## The `base` pack

`base` is a pack, not a tool. It is mandatory on each side: each consumer gets
it, and `bootstrap remove` cannot remove a tool of it, because only a pack
selects it. Only `bootstrap` changes the tools of
the
`base` pack, in the packs file.

- The front-end always adds the `base` pack. Its name is fixed in the front-end,
  so
  the values file does not select it. A `[tools.<name>]` table for one of its
  tools
  gives values. On the repo side, an `env` in that table can only add
  environments, and the sync examines it.
- Its members are in the packs file, as for each other pack. Thus a change to
  the
  pack is a change to the packs file only.
- The `mise` tool of each side has `required = true`, because it ships the
  entry of the sync. On the repo side, `git` and `pre-commit` are required too.
  The sync does not name these tools.

<!-- dprint-ignore -->
```toml
# repo/packs.toml
[pack.base]
mise = { default = true, required = true }
git = { default = true, required = true }
pre-commit = { default = true, required = true }
dprint = { default = true }
linter = { default = true }
vscode = { default = true }
"pack:security" = { default = true }

[pack.security]
gitleaks = { default = true }
grype = { default = true }
osv-scanner = { default = true }
```

`"pack:security"` adds the default members of the pack `security` to the `base`
pack. In a git submodule, the repo side skips a `root_only` tool that a pack
adds.
See [repo.md](repo.md#root_only).

The Mac `base` pack is in `macos/packs.toml`. It has `mise` (required),
`homebrew`, the shells, `git`, `ssh`, `cli` and the other tools of each Mac.

### Common tools

A pack can have a common tool. It holds the items that all the other tools of
the pack use, for example an orchestrator task and its hook. The name of a
common tool starts with `_`, for example `_security_base`.

A common tool is not a member of the pack. It is in the `depends` of each
member tool. For example, `gitleaks`, `grype` and `osv-scanner` each have
`depends = [ "_security_base" ]`.

- A pack selection and a single `[tools.<member>]` entry both bring the common
  tool, with the environments of the member tools that need it.
- The values file does not select it. A `[tools.<name>]` table for it gives
  values.
- The sync adds a dependency before the tool that needs it. Thus the tool order
  of the `base` pack ends `... vscode _security_base gitleaks grype
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

## Open points

- Later, not now: a key of the values file that removes one default member of a
  pack, for example `grype` of the `security` pack. Until then, a user cannot
  remove
  a default member of a selected pack, also not by hand.

## Generated folders

- There is no generated folder `.config/mise/conf.d/_base`. Each tool of the
  pack
  gets its own folder `.config/mise/conf.d/<tool>/`. The consumer does not edit
  these
  folders. See [mise.md](mise.md#zz-repo-and-zz-macos) for the folder of the
  owner.
