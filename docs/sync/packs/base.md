# The `base` pack

The tools that each repository gets: mise with the entry of the sync, git,
pre-commit, the formatter, the linters, the scanners and VS Code.

| Property | Value                                                                                         |
| -------- | --------------------------------------------------------------------------------------------- |
| Members  | `mise`, `git`, `pre-commit`, `dprint`, `linter`, `gitleaks`, `grype`, `osv-scanner`, `vscode` |

## When the pack is relevant

Always. `bootstrap init` writes a `[tools.<name>]` table for each member in a
new values file, as `bootstrap add pack base` does. The values file keeps no
record of the pack: do not write `[packs.base]`.

## What the pack adds

| Member        | Required | Tool file                                         |
| ------------- | -------- | ------------------------------------------------- |
| `mise`        | Yes      | [`tools/mise.md`](../tools/mise.md)               |
| `git`         | Yes      | [`tools/git.md`](../tools/git.md)                 |
| `pre-commit`  | Yes      | [`tools/pre-commit.md`](../tools/pre-commit.md)   |
| `dprint`      | Yes      | [`tools/dprint.md`](../tools/dprint.md)           |
| `linter`      | Yes      | [`tools/linter.md`](../tools/linter.md)           |
| `gitleaks`    | Yes      | [`tools/gitleaks.md`](../tools/gitleaks.md)       |
| `grype`       | Yes      | [`tools/grype.md`](../tools/grype.md)             |
| `osv-scanner` | Yes      | [`tools/osv-scanner.md`](../tools/osv-scanner.md) |
| `vscode`      | No       | [`tools/vscode.md`](../tools/vscode.md)           |

`required = true` marks a tool that each repository must have. The sync
selects a required member also without a table, and it writes an empty table
for it. The rules are in [`docs/design/packs.md`](../../design/packs.md).

The `depends` of the members add more tools: `pre-commit` adds `python`,
`linter` adds `node`, and each scanner adds
[`_security_base`](../tools/_security_base.md), the common tool of the
scanners. It adds the pre-commit hook of `code:sec`.

`vscode` is a `root_only` tool. In a git submodule, `bootstrap init` writes no
table for it.

## Environments

Each tool of the pack gets the `env` of its table, else its own
`default_envs`, limited by `environments` of the values file. Each dependency
gets the environments of the tools that need it.

## Questions

None. `bootstrap remove tool <tool…>` refuses a `required` tool. `vscode` is the
one member that a repository can remove: `bootstrap remove tool vscode`. A
`[tools.<name>]` table of a required member gives values only.

## Adapt to the repository

Do the items in "Adapt to the repository" and "Migration" of each tool file
of the pack.
