# The `base` pack

The tools that each repository gets: mise, the sync, git, pre-commit, the
formatter, the linters and the scanners.

| Property | Value                                                                      |
| -------- | -------------------------------------------------------------------------- |
| Depends  | `mise`, `repo-sync`, `git`, `pre-commit`, `dprint`, `linter`, `p:security` |
| Required | None                                                                       |

## When the pack is relevant

Always. `repo-sync` always adds the `base` pack. Its name is fixed in
`repo-sync`. Do not write `[packs.base]` in the values file.

## What the pack adds

| Tool         | Tool file                                         |
| ------------ | ------------------------------------------------- |
| `mise`       | [`tools/mise.md`](../tools/mise.md)               |
| `repo-sync`  | [`tools/repo-sync.md`](../tools/repo-sync.md)     |
| `git`        | [`tools/git.md`](../tools/git.md)                 |
| `pre-commit` | [`tools/pre-commit.md`](../tools/pre-commit.md)   |
| `dprint`     | [`tools/dprint.md`](../tools/dprint.md)           |
| `linter`     | [`tools/linter.md`](../tools/linter.md)           |
| `p:security` | The `security` pack, [`security.md`](security.md) |

The mise `depends` of `pre-commit` and `linter` add the `python` and `node`
tools too.

## Environments

Each tool of the pack gets its own `default_envs`, limited by `environments`
of the values file. Each dependency gets the environments of the tools that
need it.

## Questions

None. The pack is always added. A consumer cannot remove a tool of the pack.

## Adapt to the repository

Do the items in "Adapt to the repository" and "Migration" of each tool file
of the pack.
