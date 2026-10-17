# The `python` tool

Python and uv.

| Property  | Value |
| --------- | ----- |
| Root only | No    |
| Depends   | None  |
| Seeds     | None  |

## When the tool is relevant

- The repository has `pyproject.toml`.
- The core language of the repository is Python.

`base` adds `python` in dev for pre-commit. A repository that uses Python only
as a development tool does not need to list `python`.

## What the tool ships

| Path in the repository                       | Kind      | Use                           |
| -------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/tasks/setup/deps/python`       | Synced    | The task `setup:deps:python`. |
| `.config/mise/conf.d/python/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool | Version  | Environments         |
| --------- | -------- | -------------------- |
| `python`  | `3.14`   | dev, ci, prod, stage |
| `uv`      | `latest` | dev, ci, prod, stage |

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

The install uses `--locked`. It fails, and does not change `uv.lock`, when
the lockfile does not agree with `pyproject.toml`.

## Pre-commit hooks

None.

## Files the repository supplies

| File             | Used by                                             |
| ---------------- | --------------------------------------------------- |
| `pyproject.toml` | `setup:deps:python`. The task skips without it.     |
| `uv.lock`        | `setup:deps:python`. The install uses its versions. |

## Adapt to the repository

1. If the repository needs a different Python version, tell the user.
2. If the repository uses pip, Poetry or a `requirements.txt`, tell the user.
   The task uses uv.

## Migration

- Remove python and uv from the repository-owned mise files.
- Move the old `UV_*` variables (M7 of `migrate.md`).
