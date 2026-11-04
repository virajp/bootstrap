# Packs design

This document describes the packs of `bootstrap`. A pack is outside
`tools.toml`, so it has its own document. See [tools.md](tools.md) for the
tools.

A pack is a named group of members. It is only a pointer: each key of its
table is a member, and it has no other keys. The config, the tasks and the
files are in the tools that the pack selects. The pack is the group: a tool
adds only the tools that it needs, with its `depends`. If several packs share
config, tasks or files, they are in one tool that each of these packs names as
a member, for example `flutter`, or in a common tool `_<name>`.

## `repo/packs.toml`

All the packs are in one file, `repo/packs.toml`. It is not in a tool folder,
so the sync does not copy it to a consumer. Each pack is a `[pack.<name>]`
table:

```toml
# repo/packs.toml
[pack.flutter-android]
android = { default = true }
flutter = { default = true }

[pack.flutter-ios]
flutter = { default = true }
swift = { default = true }

[pack.flutter-web]
flutter = { default = true }
```

- Each key of the table is a member. The member forms are
  `<tool> = { default = true }` and `<tool> = {}`.
- `[packs.<name>]` of a values file adds each `default = true` member. A member
  with `{}` adds nothing: it arrives only with `[tools.<name>]`.
- A member is a tool folder in `repo/tools/`, or a nested pack with the prefix
  `pack:`, in quotes, for example `"pack:security" = { default = true }`. The
  sync expands it in turn. `"pack:<name>" = {}` does not expand: that pack
  arrives only through its own `[packs.<name>]`.
- A tool can be in several packs. The sync adds it once.
- A pack has no pack-level keys, and the only key of a member is `default`, a
  boolean. Each member is on one line.
- The sync takes the members in file order. The TOML plugin of dprint sorts
  keys, so the root `dprint.jsonc` excludes `repo/packs.toml`. Keep the order
  by hand.
- The old keys stop the sync with code 2: `depends`, `required`, `tools` or
  `default` with a list, and a member with the prefix `p:`. A member that is
  not a table or is not on one line, another key in a member, and a `default`
  that is not a boolean also stop the sync with code 2.
- A member that is not a tool folder or a pack stops the sync with code 3. The
  message names the pack.
- A pack name must not be the name of a tool folder. The sync reads all the
  packs and all the tool folders, and it stops with code 2 for a duplicate.
- A pack must not include itself, directly or through other packs, with any
  `default`. The sync stops with code 2 for a loop, and it shows the packs of
  the loop, for example `a -> b -> a`.
- Each sync examines every `[pack.*]` table, selected or not. Thus a problem
  shows before a consumer selects that pack.
- The Flutter tables of this document are examples. The Flutter packs are
  decided when they are designed.

## The questions are in `docs/sync/`

`tools.toml` and `repo/packs.toml` hold only the relationships. The
questions are in `docs/sync/packs/<pack>.md` of each pack, and in
`docs/sync/tools/<tool>.md` of the shared tool. The adopting AI asks them, and
it does not ask a question that the prompt of the user already answers.

## The answers go into the values file

The adopting AI writes one `[packs.<name>]` table for each selected pack in
`.config/bootstrap.toml`, next to the `[tools.<folder>]` tables:

```toml
[packs.flutter-android]

[packs.flutter-ios]
env = [ "dev" ]

[tools.github]
```

The only key of a `[packs.<name>]` entry is `env`. In a `[tools.<name>]` entry,
each other key is a value of the tool. See
[bootstrap-config.md](bootstrap-config.md#toolsfolderkey). A list
(`packs = [...]` or `tools = [...]`) is the old format, and it stops the sync
with code 2.

On each sync, `repo-sync` reads `repo/packs.toml`. It adds the default members
of the `base` pack, then the default members of each selected pack, then the
tools of `[tools.*]`. Then it resolves the dependencies of each tool. Thus, if
`bootstrap` later adds a `default = true` member to a pack, each consumer of
the pack gets it with no AI step. A member with `{}` arrives only in a consumer
with a `[tools.<name>]` entry.

A pack name that is not in `repo/packs.toml` stops the sync with code 3, as
an unknown tool does.

The `[packs.*]` keys have the names without the prefix `pack:`, because each
name in it is a pack.

## Environments

A tool that a pack adds gets the `env` of the `[packs.<name>]` entry, also
through a nested pack, else its own `default_envs`. Each tool of the `base` pack
gets its own
`default_envs`. A tool that is selected more than once gets each of these
environments. Each dependency gets the environments of the tool that needs it.
`[shell_alias]` is only in `dev`. See
[mise.md](mise.md#the-environments-of-a-mise-tool).

## The Flutter packs

| Pack              | Use                   |
| ----------------- | --------------------- |
| `flutter-ios`     | Develop iOS apps.     |
| `flutter-android` | Develop Android apps. |
| `flutter-web`     | Develop web apps.     |
| `flutter-macos`   | Develop macOS apps.   |
| `flutter-win`     | Develop Windows apps. |
| `flutter-linux`   | Develop Linux apps.   |

The tools of each Flutter pack are not decided.

## The `base` pack

`base` is a pack, not a tool. It is mandatory: each consumer gets it, and a
consumer cannot remove a tool of it. Only `bootstrap` changes the tools of the
`base` pack, in `repo/packs.toml`.

- `repo-sync` always adds the `base` pack. Its name is fixed in `repo-sync`,
  so the values file does not select it. A `[tools.<name>]` table for one of its
  tools only gives values.
- Its members are in `repo/packs.toml`, as for each other pack. Thus a change
  to the pack is a change to `repo/packs.toml` only.

The old `base` tool is now independent tools. Thus a change adds or removes
one tool of the pack, and does not change one large tool.

<!-- dprint-ignore -->
```toml
# repo/packs.toml
[pack.base]
mise = { default = true }
repo-sync = { default = true }
git = { default = true }
pre-commit = { default = true }
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
pack.

### Common tools

A pack can have a common tool. It holds the items that all the other tools of
the pack use, for example an orchestrator task and its hook. The name of a
common tool starts with `_`, for example `_security_base`.

A common tool is not a member of the pack. It is in the `depends` of each
member tool. For example, `gitleaks`, `grype` and `osv-scanner` each have
`depends = [ "_security_base" ]`.

- A pack selection and a single `[tools.<member>]` entry both bring the common
  tool, with the environments of the member tools that need it.
- The values file does not select it. A `[tools.<name>]` table for it only gives
  values.
- The sync adds a dependency before the tool that needs it. Thus the tool order
  of the `base` pack ends `... vscode _security_base gitleaks grype
  osv-scanner`.
  The hook order does not change, because of these four tools only
  `_security_base` has hooks.

### The tools of the `base` pack

| Tool             | Files and config                                                                                                                                                                                                                                                                                            |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mise`           | `.config/mise.toml.tmpl`, the `zz-repo` seeds, `_scripts/helpers`, `_scripts/checks`, `_scripts/placeholder`, `setup`, `setup:mise`, `setup:worktree`, `code`, `code:worktrees`, the `init` task, the `setup` alias, the `os` settings of each environment. The mise tools `jq` and `yq`, in `dev` and `ci` |
| `repo-sync`      | `setup:configs`, `setup:configs:sync`, `code:lint:repo-sync`, the mise tool `gomplate`                                                                                                                                                                                                                      |
| `git`            | `.gitignore`, `git-conventional-commits.yaml.tmpl`, `_scripts/merge`, `code:merge` and `code:merge:*`, `code:git-config`. The hooks `git-config`, `no-commit-to-branch` (its args follow `direct_commit`) and `conventional-commits`                                                                        |
| `pre-commit`     | `pre-commit-config.yaml.tmpl`, `setup:precommit`, `code:precommit`, the mise tool `pre-commit`, `PRE_COMMIT_HOME`, the `precommit` alias. The file-check hooks from `pre-commit/pre-commit-hooks`                                                                                                           |
| `dprint`         | The mise tool `dprint`, `code:format`, `setup:dprint`, the `format` hook                                                                                                                                                                                                                                    |
| `linter`         | The mise tools `npm:@virajp.dev/linter` and `shellcheck`, `code:lint`, `setup:linter`, the `lint` hook                                                                                                                                                                                                      |
| `gitleaks`       | The mise tool `gitleaks`, `.config/gitleaks.toml.tmpl` (a seed), `code:sec:gitleaks`                                                                                                                                                                                                                        |
| `grype`          | The mise tool `grype`, `.config/grype.yaml` (a seed), `code:sec:grype`                                                                                                                                                                                                                                      |
| `_security_base` | `code:sec`, which runs each `code:sec:*` task. The `sec` hook                                                                                                                                                                                                                                               |
| `osv-scanner`    | The mise tool `osv-scanner`, `.config/osv-scanner.toml` (a seed), `code:sec:osv-scanner`                                                                                                                                                                                                                    |

Each hook moves to the tool that owns its task. Thus a removed tool also
removes its hook, and no hook calls a task that is missing.

### Changes to other tools

- `claude` has `root_only = false`. `setup:ai` skips in a git submodule.

## Open points

- Later, not now: a key of `.config/bootstrap.toml` that removes one default
  member of a pack, for example `grype` of the `security` pack. Until then, a
  user cannot remove a default member of a selected pack, also not by hand.

## Generated folders

- There is no generated folder `.config/mise/conf.d/_base`. Each tool of
  the pack gets its own folder `.config/mise/conf.d/<tool>/`. The consumer does
  not edit these folders.
- `.config/mise/conf.d/zz-repo/` holds the customization of the repository.
  It loads after each tool folder, so its config wins.
