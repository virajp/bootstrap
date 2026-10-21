# Packs design

This document describes the packs of `bootstrap`. A pack is outside
`tools.toml`, so it has its own document. See [tools.md](tools.md) for the
tools.

A pack is a named list of tools. It is only a pointer: it has `depends` and
an optional `required`, and nothing else. The config, the tasks and the files
are in the tools that the pack selects. If the group tool has config, tasks or
files that all its packs use, they are in the group tool.

## `tools/packs.toml`

All the packs are in one file, `tools/packs.toml`. It is not in a tool folder,
so the sync does not copy it to a consumer. Each pack is a `[pack.<name>]`
table:

```toml
# tools/packs.toml
[pack.flutter-android]
depends = [ "flutter", "android" ]

[pack.flutter-ios]
depends = [ "flutter", "swift" ]

[pack.flutter-web]
depends = [ "flutter" ]
```

- A pack names all its tools in `depends`, the group tool included.
- A name in `depends` is a tool folder in `tools/`, or a pack with the prefix
  `p:`, for example `p:security`.
- A pack name must not be the name of a tool folder. The sync reads all the
  packs and all the tool folders, and it stops with code 2 for a duplicate.
- A pack must not include itself, directly or through other packs. The sync
  stops with code 2 for a loop, and it shows the packs of the loop, for example
  `a -> b -> a`.
- The `depends` lists of this document are examples. The packs of each group
  tool are decided when the tool is designed.

## The questions are in `docs/sync/`

`tools.toml` and `tools/packs.toml` hold only the relationships. The
questions are in `docs/sync/packs/<pack>.md` of each pack, and in
`docs/sync/tools/<tool>.md` of the group tool. The adopting AI asks them, and
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

The only key of an entry is `env`. A list (`packs = [...]` or
`tools = [...]`) is the old format, and it stops the sync with code 2.

On each sync, `repo-sync` reads `tools/packs.toml`. It adds the tools of the
`base` pack, then the tools of each selected pack, then the tools of
`[tools.*]`. Then it resolves the dependencies of each tool. Thus, if
`bootstrap` later adds a tool to a pack, each consumer gets it with no AI
step.

A pack name that is not in `tools/packs.toml` stops the sync with code 3, as
an unknown tool does.

The `[packs.*]` keys have the names without the prefix `p:`, because each
name in it is a pack.

## Environments

A tool that a pack adds gets the `env` of the `[packs.<name>]` entry, else
its own `default_envs`. Each tool of the `base` pack gets its own
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
`base` pack, in `tools/packs.toml`.

- `repo-sync` always adds the `base` pack. Its name is fixed in `repo-sync`,
  so the values file does not list it.
- Its `depends` are in `tools/packs.toml`, as for each other pack. Thus a
  change to the pack is a change to `tools/packs.toml` only.

The old `base` tool is now independent tools. Thus a change adds or removes
one tool of the pack, and does not change one large tool.

```toml
# tools/packs.toml
[pack.base]
depends = [
  "mise",
  "repo-sync",
  "git",
  "pre-commit",
  "dprint",
  "linter",
  "p:security",
]

[pack.security]
depends = [ "gitleaks", "grype", "osv-scanner" ]
required = [ "_security_base" ]
```

`p:security` adds the pack `security` to the `base` pack.

### Common tools and required tools

A pack can have a common tool. It holds the items that all the other tools of
the pack use, for example an orchestrator task and its hook. The name of a
common tool starts with `_`, for example `_security_base`.

`required` lists the tools of a pack that a consumer cannot remove.

- A name in `required` is not in `depends`.
- When a pack is selected, `repo-sync` always installs the tools of its
  `required`. The values file does not list them.
- `bootstrap` gives a consumer no way to remove a tool of a pack. But a user
  can edit `.config/bootstrap.toml` by hand. `repo-sync` follows the edit if
  it does not break a tool. It never removes a tool of `required`.

### The tools of the `base` pack

| Tool             | Files and config                                                                                                                                                                                                                                                                                          |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mise`           | `.config/mise.toml.tmpl`, the `_repo` seeds, `_scripts/helpers`, `_scripts/checks`, `_scripts/placeholder`, `setup`, `setup:mise`, `setup:worktree`, `code`, `code:worktrees`, the `init` task, the `setup` alias, the `os` settings of each environment. The mise tools `jq` and `yq`, in `dev` and `ci` |
| `repo-sync`      | `_scripts/repo-sync`, `setup:configs`, `setup:configs:sync`, the mise tool `gomplate`                                                                                                                                                                                                                     |
| `git`            | `.gitignore`, `git-conventional-commits.yaml.tmpl`, `_scripts/merge`, `code:merge` and `code:merge:*`, `code:git-config`. The hooks `git-config`, `no-commit-to-branch` and `conventional-commits`                                                                                                        |
| `pre-commit`     | `pre-commit-config.yaml.tmpl`, `setup:precommit`, `code:precommit`, the mise tool `pre-commit`, `PRE_COMMIT_HOME`, the `precommit` alias. The file-check hooks from `pre-commit/pre-commit-hooks`                                                                                                         |
| `dprint`         | The mise tool `dprint`, `code:format`, `setup:dprint`, the `format` hook                                                                                                                                                                                                                                  |
| `linter`         | The mise tools `npm:@askviraj/linter` and `shellcheck`, `code:lint`, `setup:linter`, the `lint` hook                                                                                                                                                                                                      |
| `gitleaks`       | The mise tool `gitleaks`, `.config/gitleaks.toml.tmpl` (a seed), `code:sec:gitleaks`                                                                                                                                                                                                                      |
| `grype`          | The mise tool `grype`, `.config/grype.yaml` (a seed), `code:sec:grype`                                                                                                                                                                                                                                    |
| `_security_base` | `code:sec`, which runs each `code:sec:*` task. The `sec` hook                                                                                                                                                                                                                                             |
| `osv-scanner`    | The mise tool `osv-scanner`, `.config/osv-scanner.toml` (a seed), `code:sec:osv-scanner`                                                                                                                                                                                                                  |

Each hook moves to the tool that owns its task. Thus a removed tool also
removes its hook, and no hook calls a task that is missing.

### Changes to other tools

- `claude` has `code:graph` and the `graphify-refresh` hook, because they are
  for graphify.
- `claude` has `root_only = false`, so that a submodule also gets
  `code:graph`. `setup:ai` skips in a git submodule.

## Open points

- Later, not now: a key of `.config/bootstrap.toml` that removes a tool of a
  pack, for example `grype` of the `security` pack. Until then, a user cannot
  remove a tool of a pack, also not by hand.

## Generated folders

- There is no generated folder `.config/mise/conf.d/_base`. Each tool of
  the pack gets its own folder `.config/mise/conf.d/<tool>/`. The consumer does
  not edit these folders.
- `.config/mise/conf.d/_repo/` stays. It holds the customization of the
  repository.
