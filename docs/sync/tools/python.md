# The `python` tool

Python and uv. dprint formats the Python files. In a repository with a
`pyproject.toml`, also ruff (lint) and pyrefly (types), with their `code:lint`
tasks.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `pyproject.toml`.
- The core language of the repository is Python.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
the root has a `pyproject.toml` (see "Scripts").

The `pre-commit` and `gcloud` tools have `python` in their `depends`. Thus each
repository gets `python` through the `base` pack. A repository that uses
Python only as a development tool does not need to list `python`. The
environments of a dependency are in
[`docs/design/mise.md`](../../design/mise.md).

## What the tool ships

| Path in the repository                       | Kind      | Use                                                                               |
| -------------------------------------------- | --------- | --------------------------------------------------------------------------------- |
| `.config/mise/tasks/setup/deps/50-python.py` | Synced    | The task `setup:deps:50-python`.                                                  |
| `.config/mise/tasks/code/lint/50-pyrefly.py` | Synced    | The task `code:lint:50-pyrefly`. Only with a `pyproject.toml` (`requires_found`). |
| `.config/mise/tasks/code/lint/70-ruff.py`    | Synced    | The task `code:lint:70-ruff`. Only with a `pyproject.toml` (`requires_found`).    |
| `.config/mise/tasks/code/lint/90-uv.py`      | Synced    | The task `code:lint:90-uv`.                                                       |
| `.config/mise/conf.d/python/mise.<env>.toml` | Generated | The mise config of this tool.                                                     |

## mise tools

| mise tool      | Version  | Environments | `requires_found` |
| -------------- | -------- | ------------ | ---------------- |
| `python`       | `3.14`   | dev, ci      |                  |
| `uv`           | `latest` | dev, ci      |                  |
| `ruff`         | `latest` | dev, ci      | `pyproject.toml` |
| `pipx:pyrefly` | `latest` | dev, ci      | `pyproject.toml` |

Many repositories get `python` only as a dependency of other tools. Thus the
sync installs `ruff` and `pyrefly`, and writes their two tasks, only when
the repository has a `pyproject.toml`. The rule of `requires_found` is in
[`docs/design/tools.md`](../../design/tools.md). If the repository adds its
first `pyproject.toml` later, run `mise run setup:configs` again.

The tool also sets:

- `python.uv_venv_auto = "create|source"`: mise makes and loads the virtual
  environment.
- `pipx.uvx = true` and `python.compile = false`.
- `UV_NO_MANAGED_PYTHON` and `UV_PYTHON_DOWNLOADS = "never"`. Thus uv uses
  only the Python of mise. The uv cache is on, so uv keeps one environment for
  each script path of a Python task.

These places use the same Python version (3.14). Change them together:

- The `python` tool: `[mise.tools.python]`.
- The `gcloud` tool: `CLOUDSDK_PYTHON`.
- Each Python library of `_scripts/`: `helpers.py`, `gate_files.py`,
  `checks.py`, `placeholder.py` and `worktrees_page.py` (the `mise` tool),
  `merge.py` (the `git` tool) and `xcode.py` (the `swift` tool), and each task
  file of `repo/tools/`. Each has `requires-python = ">=3.14"` and
  `python="3.14"` in its `# MISE tools={...}` line.

`helpers.gate_files` runs `gate_files.py` with `sys.executable`, the
Python of the task. Thus it has no version of its own.

The tool declares the language server `pyrefly` (`pyrefly lsp`, for `.py` and
`.pyi`), from the `pipx:pyrefly` pin, so it is listed only with a
`pyproject.toml`. The language server is in
[`docs/design/tools.md`](../../design/tools.md) ("The `[lsp]` table").
`bootstrap lsp` lists it, and `code:lsp:check` of the `mise` tool tests that it
resolves.

## Tasks

| Task                   | Use                                                                                   |
| ---------------------- | ------------------------------------------------------------------------------------- |
| `setup:deps:50-python` | Install from `uv.lock`. `--upgrade` also upgrades. It skips without `pyproject.toml`. |
| `code:lint:50-pyrefly` | Type check the `.py` and `.pyi` files with `pyrefly check`.                           |
| `code:lint:70-ruff`    | Lint the `.py` and `.pyi` files with `ruff check`. `--fix` applies the fixes.         |
| `code:lint:90-uv`      | Fail when a `uv.lock` does not agree with the `pyproject.toml` next to it.            |

`code:lint` runs the three gates. Without ruff or pyrefly, the gate skips.
The tool has no format gate: dprint (`code:format:05-dprint`, with the block
`ruff` of `dprint.jsonc`) is the one formatter of each `.py` and `.pyi` file,
also the Python task files. Each gate gets its files from `gate_files`: the
given files,
else with `--all` every file, else the uncommitted files. The ruff and pyrefly
gates use only the `.py` and `.pyi` files, and skip without one. Their
`# FILES` header excludes `.config/mise/tasks/**`: the synced Python task files
import the libraries of `_scripts/` at run time, and are not for these gates.
Each tool
gets only that list: a path exclude of the ruff or pyrefly config does not
act in a gate run, so put it in `code_files.excludes` of `[tools.mise]`.
`code:lint:90-uv` runs `uv lock --check` for each project of a `uv.lock` or a
`pyproject.toml` of its list, when the project has both files. A change that
only deletes files gives an empty list: run `code:lint --all` then. See
[`mise.md`](mise.md#the-file-sets-of-the-gates).

The install uses `--locked`. It fails, and does not change `uv.lock`, when
the lockfile does not agree with `pyproject.toml`.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                           |
| ---------------- | ------ | --------------------------------------------- |
| `scripts/detect` | detect | `match` with a `pyproject.toml`, else `none`. |

## Contributions

| Tool     | Shared file               | Items                                                                                                                                                                                                   |
| -------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `vscode` | `.vscode/extensions.json` | `meta.pyrefly`.                                                                                                                                                                                         |
| `dprint` | `dprint.jsonc`            | The block `ruff` (`indentWidth = 2`, `lineLength = 120`) and the plugin `@dprint/ruff`.                                                                                                                 |
| `claude` | `CLAUDE.md`               | The gate lines of `code:lint:50-pyrefly`, `code:lint:70-ruff` and `code:lint:90-uv` (`[contributes.claude.claude_md.gates]`), and the dependency line `python` (`[contributes.claude.claude_md.deps]`). |

The `claude` tool writes each line into the bootstrap section of
`CLAUDE.md`, only when the repository selects `claude`. See
[`claude-tool.md`](claude-tool.md).

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File             | Used by                                                                                                            |
| ---------------- | ------------------------------------------------------------------------------------------------------------------ |
| `pyproject.toml` | `setup:deps:50-python`, and the sync for ruff and pyrefly. `[tool.ruff]` and `[tool.pyrefly]` configure the gates. |
| `uv.lock`        | `setup:deps:50-python`. The install uses its versions.                                                             |

## Adapt to the repository

1. If the repository needs a different Python version, tell the user.
2. If the repository uses pip, Poetry or a `requirements.txt`, tell the user.
   The task uses uv.
3. If the repository has its own ruff or pyrefly config, keep it. If it pins
   ruff or pyrefly in its dev dependencies, tell the user. The gates use the
   mise versions.

## Migration

- Remove python and uv from the repository-owned mise files.
- Move the old `UV_*` variables (M7 of `migrate.md`).
- The old `code:check:uv` task of stackgen is now `code:lint:90-uv` (M19 of
  `migrate.md`).
