# The `python` tool

Python and uv. In a Python project, also the ruff and pyrefly gates.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- The repository has `pyproject.toml`.
- The core language of the repository is Python.

The `pre-commit` and `claude` tools add `python` in dev, and the `gcloud` tool
adds it in dev and ci, through their mise `depends`. A repository that uses
Python only as a development tool does not
need to list `python`.

## What the tool ships

| Path in the repository                       | Kind      | Use                                                         |
| -------------------------------------------- | --------- | ----------------------------------------------------------- |
| `.config/mise/tasks/setup/deps/python`       | Synced    | The task `setup:deps:python`.                               |
| `.config/mise/tasks/code/lint/ruff`          | Synced    | The task `code:lint:ruff`. Only with a `pyproject.toml`.    |
| `.config/mise/tasks/code/lint/pyrefly`       | Synced    | The task `code:lint:pyrefly`. Only with a `pyproject.toml`. |
| `.config/mise/tasks/code/lint/uv`            | Synced    | The task `code:lint:uv`.                                    |
| `.config/mise/tasks/code/format/ruff`        | Synced    | The task `code:format:ruff`. Only with a `pyproject.toml`.  |
| `.config/mise/conf.d/python/mise.<env>.toml` | Generated | The mise config of this tool.                               |

The VS Code extension is `meta.pyrefly`.

## mise tools

| mise tool      | Version  | Environments | Requires         |
| -------------- | -------- | ------------ | ---------------- |
| `python`       | `3.14`   | dev, ci      |                  |
| `uv`           | `latest` | dev, ci      |                  |
| `ruff`         | `latest` | dev, ci      | `pyproject.toml` |
| `pipx:pyrefly` | `latest` | dev, ci      | `pyproject.toml` |

Many repositories get `python` only as a dependency of other tools. Thus the
sync installs `ruff` and `pyrefly` only when the repository has a
`pyproject.toml`, at the root or down to 4 folders below it, that git does not
ignore. The `files` of the two entries apply the same condition to their
tasks: `code:lint:ruff`, `code:format:ruff` and `code:lint:pyrefly`. Without
a `pyproject.toml`, the sync does not write them, and it removes the copies
of an earlier sync. If the repository adds its first `pyproject.toml` later,
run `mise run setup:configs` again.

The tool also sets:

- `python.uv_venv_auto = "create|source"`: mise makes and loads the virtual
  environment.
- `pipx.uvx = true` and `python.compile = false`.
- `UV_NO_CACHE`, `UV_NO_MANAGED_PYTHON` and `UV_PYTHON_DOWNLOADS = "never"`.
  Thus uv uses only the Python of mise.

The `gcloud` tool uses the same Python minor version. Change both together.

## Tasks

| Task                | Use                                                                                   |
| ------------------- | ------------------------------------------------------------------------------------- |
| `setup:deps:python` | Install from `uv.lock`. `--upgrade` also upgrades. It skips without `pyproject.toml`. |
| `code:lint:ruff`    | Lint the `.py` and `.pyi` files with `ruff check`. `--fix` applies the fixes.         |
| `code:lint:pyrefly` | Type check the `.py` and `.pyi` files with `pyrefly check`.                           |
| `code:format:ruff`  | Check the format of the `.py` and `.pyi` files. `--fix` formats them.                 |
| `code:lint:uv`      | Fail when a `uv.lock` does not agree with the `pyproject.toml` next to it.            |

`code:lint` and `code:format` run the four gates. Without ruff or pyrefly,
the gate skips. Without files, the ruff and pyrefly gates also skip when the
repository has no `.py` or `.pyi` file that git does not ignore. `code:lint:uv`
runs `uv lock --check` for each project with a
`uv.lock` and a `pyproject.toml`. With a list of files, it checks only the
projects of a given `uv.lock` or `pyproject.toml`. It replaces the
`code:check:uv` task of stackgen. With a list of files, a gate uses only the
Python files. ruff
uses `--force-exclude`, so a file that the ruff config excludes stays
excluded.

The install uses `--locked`. It fails, and does not change `uv.lock`, when
the lockfile does not agree with `pyproject.toml`.

## Pre-commit hooks

None.

## Files the repository supplies

| File             | Used by                                                                                                         |
| ---------------- | --------------------------------------------------------------------------------------------------------------- |
| `pyproject.toml` | `setup:deps:python`, and the sync for ruff and pyrefly. `[tool.ruff]` and `[tool.pyrefly]` configure the gates. |
| `uv.lock`        | `setup:deps:python`. The install uses its versions.                                                             |

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
