# The `base` pack

The tools that each repository gets: mise, the sync, git, pre-commit, the
formatter, the linters, VS Code and the scanners.

| Property | Value                                                                                     |
| -------- | ----------------------------------------------------------------------------------------- |
| Members  | `mise`, `repo-sync`, `git`, `pre-commit`, `dprint`, `linter`, `vscode`, `"pack:security"` |

## When the pack is relevant

Always. `repo-sync` always adds the `base` pack. Its name is fixed in
`repo-sync`. Do not write `[packs.base]` in the values file.

## What the pack adds

Each member is `default = true` in `tools/packs.toml`, for example
`mise = { default = true }`.

| Member            | Default | Tool file                                         |
| ----------------- | ------- | ------------------------------------------------- |
| `mise`            | Yes     | [`tools/mise.md`](../tools/mise.md)               |
| `repo-sync`       | Yes     | [`tools/repo-sync.md`](../tools/repo-sync.md)     |
| `git`             | Yes     | [`tools/git.md`](../tools/git.md)                 |
| `pre-commit`      | Yes     | [`tools/pre-commit.md`](../tools/pre-commit.md)   |
| `dprint`          | Yes     | [`tools/dprint.md`](../tools/dprint.md)           |
| `linter`          | Yes     | [`tools/linter.md`](../tools/linter.md)           |
| `vscode`          | Yes     | [`tools/vscode.md`](../tools/vscode.md)           |
| `"pack:security"` | Yes     | The `security` pack, [`security.md`](security.md) |

The mise `depends` of `pre-commit` and `linter` add the `python` and `node`
tools too.

`vscode` is a `root_only` tool. In a git submodule, the sync skips it.

## Environments

Each tool of the pack gets its own `default_envs`, limited by `environments`
of the values file. Each dependency gets the environments of the tools that
need it.

## Questions

None. The pack is always added. A consumer cannot remove a tool of the pack.

## Adapt to the repository

Do the items in "Adapt to the repository" and "Migration" of each tool file
of the pack.
