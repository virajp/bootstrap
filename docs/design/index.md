# **Bootstrap** Design

## Overview

`bootstrap` allows users to setup base configurations and keep those
configuration in-sync and consistent across all their repositories

It does not limit itself to the base configuration and can also setup the
various tools that it supports

## Terminologies

- `bootstrap`: This repo
- `consumer repo`: The repo which uses `bootstrap` to create base configuration
  and keep them in sync
- `repo-sync`: Script which can be run regularly to keep the base configuration
  in sync
- `tools`: Tools that are supported by `bootstrap` and can be setup in consumer
  repos

## Requirements

`mise` MUST be pre-installed on the machine

## How bootstrap is used

`bootstrap` can be used via AI and scripts together:

- **AI, for the non-deterministic part.** A user opens an AI session in the
  consumer repository with the prompt of `README.md`. The AI reads
  `docs/sync/index.md`, asks the questions, and writes `.config/bootstrap.toml`
  in consumer repo
- **Script, for the deterministic part.** Once the `.config/bootstrap.toml` and
  base configuration is done in the consumer repo, they will have `setup` task
  in mise. The `setup` task, when run, calls `repo-sync` which reads
  `.config/bootstrap.toml` and syncs the files of the selected tools.

The repository solves two problems: add tools and dependencies to many
repositories without manual edits that drift, and keep the updates in sync.

The AI part is for a new repository, and after that only in rare situations.

## Bootstrap config

The consuming repo gets `.config/bootstrap.toml` file which contains all the
selection made during onboarding. `repo-sync` script relies on this config to
sync the updates from `bootstrap` repo to consumer repo

## References

- [tools](./tools.md)
- [.config/bootstrap.toml](./bootstrap-config.md)

## Tools

- [mise](./mise.md)
- [fnox](./fnox.md)
