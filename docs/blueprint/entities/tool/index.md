---
type: vwf-entity
title: Tool
description: A named set of setup files for one tool, shipped inside bootstrap
  and applied to a repository as a unit.
status: reviewed
implementation: partial
owner: [ cli ]
---

# Entity: Tool

## Purpose

A tool is a named set of setup files for one tool, shipped inside bootstrap
itself. It is a read-only catalog, versioned with bootstrap, and is never stored
in the target repository. Each tool belongs to a
[Tool category](../tool-category/index.md).

Used by: [Setup repository](../../flows/cli/110-setup-repository/index.md),
[Add a tool](../../flows/cli/140-add-tool/index.md),
[Remove a tool](../../flows/cli/150-remove-tool/index.md),
[Select tools](../../flows/cli/160-select-tools/index.md),
[Show the setup](../../flows/cli/170-show-setup/index.md),
[Home](../../flows/site/100-home/index.md),
[Documentation](../../flows/site/110-documentation/index.md).

Scale: 15 tools at 1.0; grows only with new bootstrap releases.

## Out of Scope

- User-defined or repository-local tools.
- Per-file selection inside a tool; the tool is the unit of selection.

## Lifecycle / State Machine

N/A — immutable catalog shipped with bootstrap. Contents change only with a new
bootstrap release.

## Catalog (1.0)

The enumerated set below is the contract. Tools are chosen with `--tool <name>`
(repeatable) at init and with `bootstrap add <tool>` / `bootstrap remove <tool>`
later; `github` and `gitlab` may be selected together. Renaming or removing a
tool or a path is a major version; adding a tool or path is minor
([changelog](../../conventions.md#changelog)).

| Name            | Category              | Purpose                                      | Base | Removable | Replaceable | Default               | Requires |
| --------------- | --------------------- | -------------------------------------------- | ---- | --------- | ----------- | --------------------- | -------- |
| `mise`          | tool-manager          | Installs the tools and runs the task library | yes  | no        | no          | on                    | none     |
| `git`           | version-control       | Ignore rules and git checks                  | yes  | no        | no          | on                    | none     |
| `pre-commit`    | git-hooks             | Runs the gates before each commit            | yes  | no        | no          | on                    | none     |
| `vscode`        | editor                | Editor settings and extensions               | yes  | yes       | yes         | on                    | none     |
| `dprint`        | formatter             | Formats code and documents                   | yes  | yes       | yes         | on                    | none     |
| `taplo`         | formatter             | Formats TOML files, run by dprint            | yes  | yes       | yes         | on                    | `dprint` |
| `virajp-linter` | linter                | Lints code with the house rules              | yes  | yes       | yes         | on                    | none     |
| `gitleaks`      | secret-scanner        | Finds secrets in code and commits            | no   | yes       | yes         | on                    | none     |
| `grype`         | vulnerability-scanner | Finds known vulnerabilities                  | no   | yes       | yes         | on                    | none     |
| `fnox`          | secrets-manager       | Gives secrets to the tasks                   | no   | yes       | yes         | off                   | none     |
| `claude`        | ai-agent              | Status line and settings for Claude Code     | no   | yes       | yes         | on                    | none     |
| `graphify`      | knowledge-graph       | Builds a knowledge graph of the code         | no   | yes       | yes         | on                    | none     |
| `mempalace`     | agent-memory          | Memory store for AI agents                   | no   | yes       | yes         | on                    | none     |
| `github`        | forge                 | Pull request and issue templates for GitHub  | no   | yes       | yes         | origin (`github.com`) | none     |
| `gitlab`        | forge                 | Merge request and issue templates for GitLab | no   | yes       | yes         | origin (`gitlab.com`) | none     |

`init --tool` or `add` of a tool whose required tool is not in the selection
after the request (for `init --tool`: the named tools plus the unremovable ones;
for `add`: the held tools minus the replaced ones, plus the named ones) exits 2
([errors](../../conventions.md#errors)) naming the missing tool; `remove` of a
tool that a tool still selected after the request requires exits 2 naming the
dependent tool, so removing both in one request passes; an `add` that replaces a
tool another selected tool requires exits 2 naming the dependent tool. In the
tui, selecting a tool whose required tool is not selected, or deselecting a tool
that a selected tool requires, is refused in place, per
[Select tools](../../flows/cli/160-select-tools/index.md) step 5. Init, add and
remove never prompt; the tui is the only interactive interface. Display order
(tui, docs) is the order of this table; changing it is a minor version.

### Target paths (1.0)

Paths are relative to the target repository root, where bootstrap writes them.
`tasks/` stands for `.config/mise/tasks/` and `conf.d/` for
`.config/mise/conf.d/`. File contents are not part of the contract, except
invariants 5 and 6, the dispatcher behaviour under Task dispatchers (what each
`_default` runs, "no tool", exit 0), the mise entries listed for
`conf.d/_base/mise.toml` and the selection-dependent files below. The setup
config `.config/bootstrap.yaml` belongs to no tool; init writes it.

| Tool            | Paths                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mise`          | `.config/mise.toml`, `.config/miserc.toml`, `conf.d/_base/mise.toml`, `conf.d/_base/mise.dev.toml`, `conf.d/_base/mise.ci.toml`, `tasks/_scripts/helpers`, `tasks/_scripts/checks`, `tasks/_scripts/merge`, `tasks/_scripts/placeholder`, `tasks/code/all`, `tasks/code/count`, `tasks/code/worktrees`, `tasks/code/merge/develop`, `tasks/code/merge/main`, `tasks/code/format/_default`, `tasks/code/lint/_default`, `tasks/code/sec/_default`, `tasks/code/graph/_default`, `tasks/setup/all`, `tasks/setup/mise`, `tasks/setup/worktree`, `tasks/setup/cleanup`, `tasks/setup/deps/all`, `tasks/setup/deps/install/_default`, `tasks/setup/deps/outdated/_default`, `tasks/setup/deps/audit/_default`, `tasks/setup/deps/upgrade/_default`, `tasks/setup/deps/cleanup/_default`, `tasks/setup/secrets/_default`, `tasks/setup/ai/_default` |
| `git`           | `.gitignore`, `tasks/code/git-config`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `pre-commit`    | `.config/pre-commit-config.yaml`, `.config/git-conventional-commits.yaml`, `tasks/code/precommit`, `tasks/setup/precommit`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `vscode`        | `.vscode/settings.json`, `.vscode/extensions.json`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `dprint`        | `dprint.json` (repository root; the full formatter config), `tasks/code/format/dprint`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `taplo`         | `.config/taplo.toml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `virajp-linter` | `.config/linter.yaml`, `eslint.config.mjs`, `tasks/code/lint/virajp-linter`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `gitleaks`      | `.config/gitleaks.toml`, `conf.d/gitleaks/mise.dev.toml`, `tasks/code/sec/gitleaks`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `grype`         | `.config/grype.yaml`, `conf.d/grype/mise.dev.toml`, `tasks/code/sec/grype`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `fnox`          | `.config/fnox.toml`, `conf.d/fnox/mise.dev.toml`, `tasks/setup/secrets/fnox`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `claude`        | `.config/claude-status.json`, `tasks/setup/ai/claude`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `graphify`      | `.graphifyignore`, `conf.d/graphify/mise.dev.toml`, `tasks/code/graph/graphify`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `mempalace`     | `mempalace.yaml` (`create_only`), `conf.d/mempalace/mise.dev.toml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `github`        | `.github/pull_request_template.md`, `.github/ISSUE_TEMPLATE/bug.yml`, `.github/ISSUE_TEMPLATE/feature.yml`, `.github/ISSUE_TEMPLATE/config.yml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `gitlab`        | `.gitlab/merge_request_templates/Default.md`, `.gitlab/issue_templates/Bug.md`, `.gitlab/issue_templates/Feature.md`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |

### Task dispatchers

A `_default` file under a task folder defines the folder's task
(`tasks/code/sec/_default` is `code:sec`) and runs every per-tool subtask
present in that folder (`code:sec:gitleaks`, `code:sec:grype`); with none
present it prints "no tool" and exits 0. `code:all` and `setup:all` call the
dispatchers and never test which tools are present. The `setup/deps/<verb>`
dispatchers are folders (`tasks/setup/deps/<verb>/_default`;
`tasks/setup/deps/all` stays a file) and run each selected tool's
`setup/deps/<verb>/<tool>` task; no 1.0 tool has one, so in 1.0 they do nothing.
`setup:deps:all` runs the `cleanup`, `install`, `upgrade`, `outdated` and
`audit` dispatchers in that order.

### Selection-dependent files

Files whose content depends on the selection are re-rendered by every command
that changes the selection. The complete list:

- `.config/mise/conf.d/_base/mise.dev.toml`: node, pnpm, python, uv, jq and yq,
  plus the entries of the selected base tools (pre-commit; dprint and
  sort-package-json; taplo; the house linter).
- `.config/git-conventional-commits.yaml`: the forge links, from the selected
  forge whose host matches `origin`; if none matches, the first selected forge
  in catalog order; with no forge selected, the file has no forge links.

`conf.d/_base/mise.toml` carries osv-scanner, pinned exactly. Versions follow
[tool-versions](../../conventions.md#tool-versions).

## Invariants

1. No file path appears in more than one tool.
2. A tool with `removable` false is always selected and cannot be removed or
   replaced.
3. Every tool of the catalog names an existing category.
4. A category with `max: one` never holds two selected tools.
5. Every task file is written executable
   ([safety](../../conventions.md#safety)).
6. Every tool a rendered task runs is installed by mise
   ([tool-versions](../../conventions.md#tool-versions)), except the named
   exemptions: `git`, `mise` itself, and the AI agent's own CLI (`claude`).
7. `removable: false` implies `replaceable: false` and `default: on`.
8. `origin_hosts` is present only when `default` is `origin`.
9. No two `origin` tools of one `max: one` category share a host.
10. `base` matters only for a tool with mise entries.
11. Every `requires` entry names a tool of the catalog.
12. A `max: one` category holds at most one `default: on` tool and at most one
    unremovable tool; every category holds at least one tool.
13. A recorded selection is checked per
    [Validity](../setup-config/index.md#validity) and
    [Repair on read](../setup-config/index.md#repair).

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)

## Relationships

| Related entity                             | Cardinality | Ownership | On delete        | Required |
| ------------------------------------------ | ----------- | --------- | ---------------- | -------- |
| [Tool category](../tool-category/index.md) | N–1         | reference | per invariant 13 | Yes      |
| [Setup config](../setup-config/index.md)   | N–M         | reference | per invariant 13 | No       |
| [Tool](./index.md) (`requires`, self)      | N–M         | reference | per invariant 13 | No       |

## Concurrency & Consistency

- Concurrent-write resolution: N/A — read-only catalog, never written at
  runtime.
- Uniqueness guarantees under races: N/A — read-only.
- Idempotency of each mutating action: N/A — no mutating actions.

## References

- [errors](../../conventions.md#errors)
- [baseline](../../conventions.md#baseline)
- [safety](../../conventions.md#safety)
- [changelog](../../conventions.md#changelog)
- [tool-versions](../../conventions.md#tool-versions)
