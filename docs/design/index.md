# **Bootstrap** Design

## Overview

`bootstrap` allows users to setup base configurations and keep those
configuration in-sync and consistent across all their repositories. It also
keeps the configuration of a Mac in sync.

It does not limit itself to the base configuration and can also setup the
various tools that it supports

## Terminologies

- `bootstrap`: This repo
- `consumer repo`: The repo which uses `bootstrap` to create base configuration
  and keep them in sync
- `repo-sync`: The front-end of the sync for a repository. A consumer repo runs
  it regularly to keep the base configuration in sync
- `setup-sync`: The front-end of the sync for a Mac. It renders the Mac tools
  into `$HOME`
- `sync engine`: The shared libraries of both front-ends, in `sync-engine/`
- `tools`: Tools that are supported by `bootstrap` and can be setup in consumer
  repos or on a Mac

## Requirements

`mise` MUST be pre-installed on the machine

## How bootstrap is used

`bootstrap` can be used via AI and scripts together:

- **AI, for the non-deterministic part.** A user opens an AI session in the
  consumer repository with the prompt of `README.md`. The AI reads
  `docs/sync/index.md`, asks the questions, and writes `.config/bootstrap.toml`
  in consumer repo
- **Script, for the deterministic part.** Once the `.config/bootstrap.toml` and
  base configuration is done in the consumer repo, they will have the
  `setup:configs` task in mise. The `setup:configs` task, when run, calls
  `repo-sync` which reads
  `.config/bootstrap.toml` and syncs the files of the selected tools.

A Mac uses the same two parts. The AI reads `docs/setup/index.md`, asks the
questions, and writes `macos.toml`. After that, the `updateall:configs` task
calls `setup-sync`, which reads `macos.toml` and renders the selected tools into
`$HOME`.

The repository solves two problems: add tools and dependencies to many
repositories without manual edits that drift, and keep the updates in sync.

The AI part is for a new repository, or for the first adoption or the migration
of a Mac that has mise and the clone, and after that only in rare situations. A
new Mac has neither, so it uses the parked `setup` script of `macos-setup`.

## Bootstrap config

The consuming repo gets `.config/bootstrap.toml` file which contains all the
selection made during onboarding. `repo-sync` script relies on this config to
sync the updates from `bootstrap` repo to consumer repo. The Mac has its own
values file, `macos.toml`.

## References

- [sync-engine.md](./sync-engine.md): the rules that both front-ends share: the
  cache and the clone, the tool calls, the templates, the write of one target,
  merge mode, the header, the manifest and the stale rule, the values merge,
  `bootstrap_commit` and the exit codes 2, 3 and 5.
- [repo-sync.md](./repo-sync.md): the repo front-end: the entry tasks, members,
  the template context, the branch model, `direct_commit`, the pre-commit
  render,
  `--check` and its gate.
- [setup-sync.md](./setup-sync.md): the Mac front-end: `macos.toml` and its
  keys,
  the selection, the stages and exit 4, the login shell check and the symlink
  resolve.
- [packs.md](./packs.md): the packs model of both sides.
- [tools.md](./tools.md): the `tools.toml` schema, the `[files]` keys and the
  rule
  for the header of a source file.
- [mise.md](./mise.md): the mise generator, the environments,
  `[overrides.mise*]`,
  `zz-repo` and `zz-macos`, and the tasks.
- [bootstrap-config.md](./bootstrap-config.md): the keys of
  `.config/bootstrap.toml`.
- [fnox.md](./fnox.md): the design of the `fnox` tool and of the tasks that use
  it.
