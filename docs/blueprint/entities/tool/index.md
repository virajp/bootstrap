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

The tools form a read-only catalog that is versioned with bootstrap and never
stored in the target repository.

Used by: [Setup repository](../../flows/cli/110-setup-repository/index.md),
[Add a tool](../../flows/cli/140-add-tool/index.md),
[Remove a tool](../../flows/cli/150-remove-tool/index.md),
[Select tools](../../flows/cli/160-select-tools/index.md),
[Show the setup](../../flows/cli/170-show-setup/index.md),
[Home](../../flows/site/100-home/index.md),
[Documentation](../../flows/site/110-documentation/index.md).

Scale: 22 tools at 1.0; grows only with new bootstrap releases.

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

| Name            | Category              | Purpose                                      | Removable | Hidden | Default               | Requires       |
| --------------- | --------------------- | -------------------------------------------- | --------- | ------ | --------------------- | -------------- |
| `mise`          | tool-manager          | Installs the tools and runs the task library | no        | yes    | on                    | none           |
| `git`           | version-control       | Ignore rules and git checks                  | no        | no     | on                    | none           |
| `pre-commit`    | git-hooks             | Runs the gates before each commit            | no        | no     | on                    | `python`, `uv` |
| `vscode`        | editor                | Editor settings and extensions               | yes       | no     | on                    | none           |
| `dprint`        | formatter             | Formats code and documents                   | yes       | no     | on                    | `node`, `pnpm` |
| `taplo`         | formatter             | Formats TOML files, run by dprint            | yes       | no     | on                    | `dprint`       |
| `virajp-linter` | linter                | Lints code with the house rules              | yes       | no     | on                    | `node`, `pnpm` |
| `node`          | runtime               | JavaScript runtime                           | yes       | no     | off                   | none           |
| `pnpm`          | runtime               | JavaScript package manager                   | yes       | no     | off                   | `node`         |
| `python`        | runtime               | Python runtime                               | yes       | no     | off                   | none           |
| `uv`            | runtime               | Python package and tool manager              | yes       | no     | off                   | `python`       |
| `jq`            | runtime               | Processes JSON on the command line           | yes       | no     | off                   | none           |
| `yq`            | runtime               | Processes YAML on the command line           | yes       | no     | off                   | none           |
| `gitleaks`      | secret-scanner        | Finds secrets in code and commits            | yes       | no     | on                    | none           |
| `grype`         | vulnerability-scanner | Finds known vulnerabilities                  | yes       | no     | on                    | none           |
| `osv-scanner`   | vulnerability-scanner | Scans lockfiles for known vulnerabilities    | yes       | no     | on                    | none           |
| `fnox`          | secrets-manager       | Gives secrets to the tasks                   | yes       | no     | off                   | none           |
| `claude`        | ai-agent              | Status line and settings for Claude Code     | yes       | no     | on                    | `jq`, `yq`     |
| `graphify`      | knowledge-graph       | Builds a knowledge graph of the code         | yes       | no     | on                    | `python`, `uv` |
| `mempalace`     | agent-memory          | Memory store for AI agents                   | yes       | no     | on                    | `python`, `uv` |
| `github`        | forge                 | Pull request and issue templates for GitHub  | yes       | no     | origin (`github.com`) | none           |
| `gitlab`        | forge                 | Merge request and issue templates for GitLab | yes       | no     | origin (`gitlab.com`) | none           |

The file paths of a hidden tool (`mise` only, invariant 7) are still listed like
any path. A category whose every tool is hidden (`tool-manager`) is not shown.
Each shown tool's `purpose` appears after its name in the tui and the docs;
hidden tools are left out of the site Tools and Tool categories pages.

### Requires and dependencies {#requires-and-dependencies}

`requires` is a list of slots. Each slot lists one or more alternative tools;
the first alternative is the default. In the Requires column, slots are
separated by commas and the alternatives of a slot by `|` (for example `python`,
`uv` is two slots of one alternative each). No 1.0 slot has more than one.

A tool that a selected tool requires and that the user did not select directly
is a dependency. A dependency is installed and rendered like any selected tool
and is recorded in the [Setup config](../setup-config/index.md).

The `requires` check, per request ([errors](../../conventions.md#errors)):

| Request                                                                     | Checked against                                                 | Fails when                                                                                     | Outcome                                                                                                                                      |
| --------------------------------------------------------------------------- | --------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `init` first run (default or `--tool` selection), or a re-run with `--tool` | the selected tools (default or named) plus the unremovable ones | never for a missing required tool                                                              | each slot with no selected tool gets its first alternative as a dependency                                                                   |
| `add`                                                                       | the held tools minus the replaced ones, plus the named          | never for a missing required tool                                                              | the same; `add` of a held dependency makes it direct (no parents change); each dependency that no parent needs after the request is pruned   |
| `add <alternative> --replace` on a slot                                     | the same                                                        | never                                                                                          | the parents of the replaced alternative move to the new one; the old one is pruned unless it is selected directly or another parent needs it |
| `add` or an init re-run that replaces a tool outside a slot                 | the selection after the request                                 | a selected tool requires the replaced tool and no other alternative of its slot is selected    | exit 2 naming the parent tool                                                                                                                |
| `remove`                                                                    | the selection after the request                                 | a still-selected tool requires a removed tool and no other alternative of its slot is selected | exit 2 naming the parent tool; removing the parent in the same request passes                                                                |
| `remove` and an init re-run (drops)                                         | the selection after the request                                 | never                                                                                          | each dependency that no parent needs is pruned                                                                                               |
| tui select or deselect                                                      | the selection after the change                                  | per [Select tools](../../flows/cli/160-select-tools/index.md)                                  | per flow 160                                                                                                                                 |

Display order (tui, docs, shown tools only) is the order of the Catalog table;
changing it is a minor version.

### Target paths (1.0)

Paths are relative to the target repository root, where bootstrap writes them.
`tasks/` stands for `.config/mise/tasks/` and `conf.d/` for
`.config/mise/conf.d/`. File contents are not part of the contract, except
invariants 5 and 6, the dispatcher behaviour under Task dispatchers (what each
`_default` runs, "no tool", exit 0), the mise entries listed for the `conf.d/`
files below and the selection-dependent files below. The setup config
`.config/bootstrap.yaml` belongs to no tool; init, add, remove and the tui write
it.

| Tool            | Paths                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mise`          | `.config/mise.toml`, `.config/miserc.toml`, `conf.d/env.toml`, `conf.d/task-init.dev.toml`, `conf.d/alias.dev.toml`, `tasks/_scripts/helpers`, `tasks/_scripts/checks`, `tasks/_scripts/merge`, `tasks/_scripts/placeholder`, `tasks/code/all`, `tasks/code/count`, `tasks/code/worktrees`, `tasks/code/merge/develop`, `tasks/code/merge/main`, `tasks/code/format/_default`, `tasks/code/lint/_default`, `tasks/code/sec/_default`, `tasks/code/graph/_default`, `tasks/code/git-hooks/_default`, `tasks/setup/all`, `tasks/setup/mise`, `tasks/setup/worktree`, `tasks/setup/cleanup`, `tasks/setup/deps/all`, `tasks/setup/deps/install/_default`, `tasks/setup/deps/outdated/_default`, `tasks/setup/deps/audit/_default`, `tasks/setup/deps/upgrade/_default`, `tasks/setup/deps/cleanup/_default`, `tasks/setup/secrets/_default`, `tasks/setup/git-hooks/_default`, `tasks/setup/ai/_default` |
| `git`           | `.gitignore`, `tasks/code/git-config`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `pre-commit`    | `.config/pre-commit-config.yaml`, `.config/git-conventional-commits.yaml`, `conf.d/pre-commit/mise.dev.toml` (pre-commit and the shell alias `precommit`), `tasks/code/git-hooks/pre-commit`, `tasks/setup/git-hooks/pre-commit`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `vscode`        | `.vscode/settings.json`, `.vscode/extensions.json`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `dprint`        | `dprint.json` (repository root; the full formatter config), `conf.d/dprint/mise.dev.toml` (dprint and sort-package-json, which is part of dprint and not a catalog tool), `tasks/code/format/dprint`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `taplo`         | `.config/taplo.toml`, `conf.d/taplo/mise.dev.toml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `virajp-linter` | `.config/linter.yaml`, `eslint.config.mjs`, `conf.d/virajp-linter/mise.dev.toml`, `tasks/code/lint/virajp-linter`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `node`          | `conf.d/node/mise.toml` (setting node.compile; the `node_modules/.bin` path env), `conf.d/node/mise.dev.toml` (node), `conf.d/node/mise.ci.toml` (setting node.gpg_verify)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `pnpm`          | `conf.d/pnpm/mise.toml` (setting npm.package_manager), `conf.d/pnpm/mise.dev.toml` (pnpm)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `python`        | `conf.d/python/mise.dev.toml` (python; settings python.compile and python.uv_venv_auto)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `uv`            | `conf.d/uv/mise.dev.toml` (uv; setting pipx.uvx; env UV_NO_CACHE, UV_NO_MANAGED_PYTHON, UV_PYTHON_DOWNLOADS)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `jq`            | `conf.d/jq/mise.dev.toml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `yq`            | `conf.d/yq/mise.dev.toml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `gitleaks`      | `.config/gitleaks.toml`, `conf.d/gitleaks/mise.dev.toml`, `tasks/code/sec/gitleaks`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `grype`         | `.config/grype.yaml`, `conf.d/grype/mise.dev.toml`, `tasks/code/sec/grype`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `osv-scanner`   | `conf.d/osv-scanner/mise.toml` (osv-scanner, pinned exactly), `tasks/code/sec/osv-scanner`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `fnox`          | `.config/fnox.toml`, `conf.d/fnox/mise.dev.toml`, `tasks/setup/secrets/fnox`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `claude`        | `.config/claude-status.json`, `tasks/setup/ai/claude`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `graphify`      | `.graphifyignore`, `conf.d/graphify/mise.dev.toml`, `tasks/code/graph/graphify`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `mempalace`     | `mempalace.yaml` (`create_only`), `conf.d/mempalace/mise.dev.toml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `github`        | `.github/pull_request_template.md`, `.github/ISSUE_TEMPLATE/bug.yml`, `.github/ISSUE_TEMPLATE/feature.yml`, `.github/ISSUE_TEMPLATE/config.yml`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `gitlab`        | `.gitlab/merge_request_templates/Default.md`, `.gitlab/issue_templates/Bug.md`, `.gitlab/issue_templates/Feature.md`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |

Mise entries go in the env file they belong to (`mise.toml`, `mise.dev.toml`,
`mise.ci.toml`); a setting lives with the tool it applies to. Besides the tool
folders, `conf.d/` holds files owned by the `mise` tool:

- `conf.d/env.toml`: the product env values `MEMBERS`, `MERGE_MODEL_DEVELOP`,
  `MERGE_MODEL_MAIN` and `REPO_NAME`; `conf.d/env.<env>.toml` holds values
  specific to one `MISE_ENV` (none in 1.0).
- `conf.d/task-init.dev.toml`: the `init` task.
- `conf.d/alias.dev.toml`: the shell aliases `setup` and `worktrees`.

Mise-wide settings stay in `.config/mise.toml`.

### Task dispatchers

A `_default` file under a task folder defines the folder's task
(`tasks/code/sec/_default` is `code:sec`) and runs every per-tool subtask
present in that folder (`code:sec:gitleaks`, `code:sec:grype`,
`code:sec:osv-scanner`); with none present it prints "no tool" and exits 0. The
hook tasks are dispatchers too, owned by `mise`:
`tasks/setup/git-hooks/_default` (`setup:git-hooks`) and
`tasks/code/git-hooks/_default` (`code:git-hooks`). Each git-hooks tool supplies
its subtasks; `pre-commit` supplies `setup:git-hooks:pre-commit` and
`code:git-hooks:pre-commit`. The `precommit` shell alias stays with `pre-commit`
(a replacing tool brings its own alias). `code:all` and `setup:all` call the
dispatchers and never test which tools are present. `code:all` runs
`code:format`, `code:lint` and `code:sec` in that order; `setup:all` runs
`setup:mise`, `setup:secrets`, `setup:deps:all`, `setup:git-hooks` and
`setup:ai` in that order. The `setup/deps/<verb>` dispatchers are folders
(`tasks/setup/deps/<verb>/_default`; `tasks/setup/deps/all` stays a file) and
run each selected tool's `setup/deps/<verb>/<tool>` task; no 1.0 tool has one,
so in 1.0 they do nothing. `setup:deps:all` runs the `cleanup`, `install`,
`upgrade`, `outdated` and `audit` dispatchers in that order.

### Selection-dependent files {#selection-dependent-files}

Files whose content depends on the selection are re-rendered by every command
that changes the selection. The complete list:

- `.config/git-conventional-commits.yaml`: the forge links, from the selected
  forge whose host matches `origin`; if none matches, the first selected forge
  in catalog order; with no forge selected, the file has no forge links. The
  three links use `<host>`, the host of `origin` (with no readable `origin`
  host, the forge's first `origin_hosts` entry), and `<repo>`, the recorded
  repository path:

  | Link    | `github`                                                 | `gitlab`                                        |
  | ------- | -------------------------------------------------------- | ----------------------------------------------- |
  | commit  | `https://<host>/<repo>/commit/<sha>`                     | `https://<host>/<repo>/-/commit/<sha>`          |
  | compare | `https://<host>/<repo>/compare/<from>...<to>?diff=split` | `https://<host>/<repo>/-/compare/<from>...<to>` |
  | issue   | `https://<host>/<repo>/issues/<n>`                       | `https://<host>/<repo>/-/issues/<n>`            |
- `dprint.json`: its TOML formatting (the exec plugin and its command that run
  taplo for TOML files) is present only when `taplo` is selected; without taplo,
  dprint does not format TOML files.

Versions follow [tool-versions](../../conventions.md#tool-versions).

## Invariants

1. No file path appears in more than one tool.
2. A tool with `removable` false is always selected and cannot be removed; if
   shown, it can be replaced in place within its `max: one` category.
3. Every tool of the catalog names an existing category.
4. A category with `max: one` never holds two selected tools.
5. Every task file is written executable
   ([safety](../../conventions.md#safety)).
6. Every tool a rendered task runs is installed by mise
   ([tool-versions](../../conventions.md#tool-versions)), except the named
   exemptions: `git`, `mise` itself and the AI agent's own CLI (`claude`).
7. `mise` is the only hidden tool: it is always applied (init, add, remove and
   the tui render its files as for any selected tool), never recorded in
   `values.tools` or `values.dependencies`, and never shown or named in any
   tool-name list (tui, human and `--json` output, `show`), so no request can
   name it (naming it in `init --tool`, `add` or `remove` is an unknown tool
   name), replace it or remove it; `removable: false` implies `default: on`.
8. `origin_hosts` is present only when `default` is `origin`.
9. No two `origin` tools of one `max: one` category share a host.
10. Every tool but `mise` renders its mise entries only into
    `.config/mise/conf.d/<tool>/` (the `mise` tool's own files are listed in its
    Target paths row), and each mise entry declares mise `depends` on the
    selected alternative of each `requires` slot it runs on, so mise installs
    them first.
11. Every alternative of every `requires` slot names a tool of the catalog; no
    tool requires itself, and the `requires` graph has no cycle.
12. A `max: one` category holds at most one `default: on` tool and at most one
    unremovable tool; every category holds at least one tool.
13. A recorded selection is checked per
    [Validity](../setup-config/index.md#validity) and
    [Repair on read](../setup-config/index.md#repair).
14. A replacement applies only in a `max: one` category or within the
    alternatives of a `requires` slot; elsewhere a tool is added and the other
    removed.

## Data Model

Authoritative schema: [schema.yaml](./schema.yaml)

## Relationships

| Related entity                             | Cardinality | Ownership | On delete                                     | Required |
| ------------------------------------------ | ----------- | --------- | --------------------------------------------- | -------- |
| [Tool category](../tool-category/index.md) | N–1         | reference | cannot happen within a release (invariant 3)  | Yes      |
| [Setup config](../setup-config/index.md)   | N–M         | reference | per invariant 13                              | No       |
| [Tool](./index.md) (`requires`, self)      | N–M         | reference | cannot happen within a release (invariant 11) | No       |

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
