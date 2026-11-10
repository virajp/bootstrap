# **Bootstrap** Design

## Overview

`bootstrap` lets users install base configurations and keep those
configurations in sync and consistent across all their repositories. It also
keeps the configuration of a Mac in sync.

It does not limit itself to the base configuration. It can also install the
various tools that it supports.

## Terminologies

- `bootstrap`: This repository, and its CLI `bootstrap`.
- `consumer repo`: The repository which uses `bootstrap` to create base
  configuration and keep it in sync.
- `the repo side`: `bootstrap sync` and the other commands without `mac`. They
  write the files of the selected tools of `repo/tools/` into a repository.
- `the Mac side`: `bootstrap mac <command>`. It renders the Mac tools of
  `macos/tools/` into `$HOME`.
- `sync engine`: The Go code of the CLI, in `cli/`.
- `tools`: Tools that `bootstrap` supports, and that a consumer repo or a Mac
  can select.

## Requirements

`mise` MUST be installed on the machine. mise installs the CLI:
`mise use -g github:virajp/bootstrap`.

## How bootstrap is used

`bootstrap` can be used via AI and the CLI together:

- **AI, for the non-deterministic part.** A user opens an AI session in the
  consumer repository with the prompt of `README.md`. The AI reads
  `docs/sync/index.md`, selects the tools, and gets the answers of their
  questions. Then it runs `bootstrap init <tool…> --answer …`, which writes
  `.config/bootstrap.toml` and syncs.
- **CLI, for the deterministic part.** After the adoption, the repository has
  the `setup:configs` task in mise. The task runs `bootstrap sync`, which reads
  `.config/bootstrap.toml` and syncs the files of the selected tools.
  `bootstrap add`, `remove` and `replace` change the selection later.

A Mac uses the same two parts. The AI reads `docs/setup/index.md`, and runs
`bootstrap mac init`, which writes `macos.toml`. After that, the
`updateall:configs` task runs `bootstrap mac sync`, which reads `macos.toml`
and renders the selected tools into `$HOME`.

The repository solves two problems: add tools and dependencies to many
repositories without manual edits that drift, and keep the updates in sync.

The AI part is for a new repository, or for the first adoption or the migration
of a Mac, and after that only in rare situations. `bootstrap init` also brings
an adopted repository up to date.

## Bootstrap config

The consumer repository gets the file `.config/bootstrap.toml`, which contains
all the selections of the adoption. The sync reads it to bring the updates from
the `bootstrap` repository to the consumer repository. The Mac has its own
values file, `macos.toml`.

## References

- [cli.md](./cli.md): the rules that both sides share: the
  command line, the cache and the clone, the version check, the scripts of a
  tool, the templates, the write of one target, shared files, the header, the
  manifest and the stale rule, the values of a tool, `bootstrap_commit`, the
  exit codes and the release.
- [repo.md](./repo.md): the repo side: the entry tasks, members, the template
  context, the branch model, `direct_commit`, the pre-commit config,
  `bootstrap check` and its gate.
- [mac.md](./mac.md): the Mac side: the repository of `macos.toml`, the entry
  tasks, `macos.toml` and its keys, the selection, the stages and exit 4, and
  the symlink resolve.
- [packs.md](./packs.md): the packs model of both sides.
- [tools.md](./tools.md): the `tools.toml` schema, the `[files]`, `[tasks]` and
  `[formatter]` keys, the scripts, and the rule for the header of a source
  file.
- [mise.md](./mise.md): the mise generator, the environments,
  `[overrides.mise*]`, `zz-repo` and `zz-macos`, and the tasks.
- [bootstrap-config.md](./bootstrap-config.md): the keys of
  `.config/bootstrap.toml`, and its old forms.
- [fnox.md](./fnox.md): the design of the `fnox` tool and of the tasks that use
  it.
