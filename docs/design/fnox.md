# fnox design

This document records the design decisions for the `fnox` tool and for the
tasks that use it. The decisions come from a design session on 2026-10-20.
They are not implemented yet. See [Backlog](#backlog).

## The rule for run tasks

Each task that runs the application uses fnox to supply the secrets, with a
profile. The user stated the rule as:

> Every task that run the application must use `fnox` to provide secrets along
> with the specific profile.

The task runs the command through fnox:

```bash
mise x -- fnox exec --profile "${MISE_ENV}" -- <command>
```

- The profile is the mise environment: `dev` or `ci`.
- The task runs `fnox exec` from the folder of its project. Thus fnox loads the
  shared files of `.config/` and the files of that project folder.

## Environments

| Environment | Secrets source                                                |
| ----------- | ------------------------------------------------------------- |
| `dev`       | fnox. The `age` key comes from the keychain.                  |
| `ci`        | fnox. The `age` key comes from the secrets manager of the CI. |

The current `fnox` tool installs fnox in dev only. The new design installs it
in dev and ci.

## Provider and files

- fnox uses the `age` provider. The encrypted values are in the fnox config
  files, and git tracks them.
- Profiles are mandatory in a repository that uses fnox.
- A profile is a file, `fnox.<profile>.toml`, not a `[profiles.<name>]` table.
  There is one file for each environment: `fnox.dev.toml` and `fnox.ci.toml`.
- In dev, the `keychain` provider supplies the `age` key (the identity).
- In ci, the secrets manager of the CI supplies the `age` key.

## Folders limit the exposure of secrets

Each project folder has its own profile files. A task of a project gets only
the shared secrets and the secrets of that project.

```text
repo/
├── .config/
│   ├── fnox.toml               # root = true;
│   ├── fnox.dev.toml           # age-dev provider; dev secrets shared by every project
│   └── fnox.ci.toml            # age-ci provider;  CI secrets shared by every project
├── packages/
│   └── common/                 # no fnox files
└── projects/
    ├── service/
    │   ├── fnox.dev.toml       # service-only dev secrets
    │   └── fnox.ci.toml        # service-only CI secrets
    ├── worker/
    │   ├── fnox.dev.toml
    │   └── fnox.ci.toml
    └── console/
        ├── fnox.dev.toml
        └── fnox.ci.toml
```

## Projects and members

Each task belongs to one project. A project is the repository itself, or one
project or package of a monorepo. The `MEMBERS` list names the members of a
monorepo.

A git submodule is also in `MEMBERS`, but the parent repository does not
create tasks for it. The user stated:

> a parent repo will always call the task in the submodule if needed and never
> run the task directly in the parent repo

## Superseded decisions

These decisions came before the final design. They are not used.

- The profile is the project name: a name of `MEMBERS`, or `repo_name` in a
  repository without members. The final design uses the environment as the
  profile, and the folder as the project.
- In `ci`, the task runs the command directly, because the CI supplies its own
  secrets as environment variables. The final design uses fnox in `ci` too.

## Fact from the fnox documentation

The fnox documentation (Context7 `/jdx/fnox`) gives this merge order, from
the lowest to the highest priority:

1. The global config, `~/.config/fnox/config.toml`.
2. The `fnox.toml` of the outer folder.
3. The `fnox.<profile>.toml` of the outer folder, for each active profile.
4. The `fnox.local.toml` of the outer folder.
5. The same three files of each inner folder.

`root = true` stops the search of the parent folders. The global config still
loads. The current `fnox` tool sets `FNOX_CONFIG_DIR` to `.config/fnox/` for
this reason.

## Open points

- The docs say that fnox loads "discovered project directories". The project
  folders in the design have no `fnox.toml`. It is possible that fnox does not
  load `projects/service/fnox.dev.toml` without a `fnox.toml` in the same
  folder. Test it with `mise x -- fnox config-files` from a project folder.
- The tasks that count as "tasks that run the application" are not decided:
  the start of the app only, or also the tests and the release build. Their
  names are not decided.

## Backlog

- Change the `fnox` tool to this design: dev and ci, the `age` provider, one
  file for each environment, the `age` key from the keychain in dev and from
  the CI secrets in ci, and one folder for each project. Include the test of
  the first open point.
