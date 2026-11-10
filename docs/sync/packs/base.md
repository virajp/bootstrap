# The `base` pack

The tools that each repository gets: mise with the entry of the sync, git,
pre-commit, the formatter, the linters, VS Code and the scanners.

| Property | Value                                                                        |
| -------- | ---------------------------------------------------------------------------- |
| Members  | `mise`, `git`, `pre-commit`, `dprint`, `linter`, `vscode`, `"pack:security"` |

## When the pack is relevant

Always. `bootstrap sync` always adds the `base` pack. Do not write
`[packs.base]` in the values file.

## What the pack adds

| Member            | Default | Required | Tool file                                         |
| ----------------- | ------- | -------- | ------------------------------------------------- |
| `mise`            | Yes     | Yes      | [`tools/mise.md`](../tools/mise.md)               |
| `git`             | Yes     | Yes      | [`tools/git.md`](../tools/git.md)                 |
| `pre-commit`      | Yes     | Yes      | [`tools/pre-commit.md`](../tools/pre-commit.md)   |
| `dprint`          | Yes     | No       | [`tools/dprint.md`](../tools/dprint.md)           |
| `linter`          | Yes     | No       | [`tools/linter.md`](../tools/linter.md)           |
| `vscode`          | Yes     | No       | [`tools/vscode.md`](../tools/vscode.md)           |
| `"pack:security"` | Yes     | No       | The `security` pack, [`security.md`](security.md) |

`required = true` marks a tool that each repository must have. The rules of
`default` and `required` are in
[`docs/design/packs.md`](../../design/packs.md).

The `depends` of the members add more tools: `pre-commit` adds `python`, and
`linter` adds `node`.

`vscode` is a `root_only` tool. In a git submodule, the sync skips it.

## Environments

Each tool of the pack gets its own `default_envs`, limited by `environments`
of the values file. Each dependency gets the environments of the tools that
need it.

## Questions

None. The pack is always added. `bootstrap remove` refuses a `required` tool,
and a tool that only a pack selects. Thus a consumer cannot remove a tool of
the pack. A `[tools.<name>]` table of a member gives values only.

## Adapt to the repository

Do the items in "Adapt to the repository" and "Migration" of each tool file
of the pack.
