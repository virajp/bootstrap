# The `python` tool

Python and uv. In a repository with a `pyproject.toml`, also ruff and pyrefly,
with their `code:lint` and `code:format` tasks.

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
| `.config/mise/tasks/setup/deps/50-python`    | Synced    | The task `setup:deps:50-python`.                                                  |
| `.config/mise/tasks/code/format/40-ruff`     | Synced    | The task `code:format:40-ruff`. Only with a `pyproject.toml` (`requires_found`).  |
| `.config/mise/tasks/code/lint/50-pyrefly`    | Synced    | The task `code:lint:50-pyrefly`. Only with a `pyproject.toml` (`requires_found`). |
| `.config/mise/tasks/code/lint/70-ruff`       | Synced    | The task `code:lint:70-ruff`. Only with a `pyproject.toml` (`requires_found`).    |
| `.config/mise/tasks/code/lint/90-uv`         | Synced    | The task `code:lint:90-uv`.                                                       |
| `.config/mise/conf.d/python/mise.<env>.toml` | Generated | The mise config of this tool.                                                     |

## mise tools

| mise tool      | Version  | Environments | `requires_found` |
| -------------- | -------- | ------------ | ---------------- |
| `python`       | `3.14`   | dev, ci      |                  |
| `uv`           | `latest` | dev, ci      |                  |
| `ruff`         | `latest` | dev, ci      | `pyproject.toml` |
| `pipx:pyrefly` | `latest` | dev, ci      | `pyproject.toml` |

Many repositories get `python` only as a dependency of other tools. Thus the
sync installs `ruff` and `pyrefly`, and writes their three tasks, only when
the repository has a `pyproject.toml`. The rule of `requires_found` is in
[`docs/design/tools.md`](../../design/tools.md). If the repository adds its
first `pyproject.toml` later, run `mise run setup:configs` again.

The tool also sets:

- `python.uv_venv_auto = "create|source"`: mise makes and loads the virtual
  environment.
- `pipx.uvx = true` and `python.compile = false`.
- `UV_NO_CACHE`, `UV_NO_MANAGED_PYTHON` and `UV_PYTHON_DOWNLOADS = "never"`.
  Thus uv uses only the Python of mise.

The `gcloud` tool uses the same Python minor version. Change both together.

## Tasks

| Task                   | Use                                                                                   |
| ---------------------- | ------------------------------------------------------------------------------------- |
| `setup:deps:50-python` | Install from `uv.lock`. `--upgrade` also upgrades. It skips without `pyproject.toml`. |
| `code:format:40-ruff`  | Check the format of the `.py` and `.pyi` files. `--fix` formats them.                 |
| `code:lint:50-pyrefly` | Type check the `.py` and `.pyi` files with `pyrefly check`.                           |
| `code:lint:70-ruff`    | Lint the `.py` and `.pyi` files with `ruff check`. `--fix` applies the fixes.         |
| `code:lint:90-uv`      | Fail when a `uv.lock` does not agree with the `pyproject.toml` next to it.            |

`code:lint` and `code:format` run the four gates. Without ruff or pyrefly,
the gate skips. Without files, the ruff and pyrefly gates also skip when the
repository has no `.py` or `.pyi` file that git does not ignore.
`code:lint:90-uv` runs `uv lock --check` for each project with a `uv.lock` and a
`pyproject.toml`. With a list of files, it checks only the projects of a given
`uv.lock` or `pyproject.toml`. With a list of files, the other gates use only
the Python files. ruff uses `--force-exclude`, so a file that the ruff config
excludes stays excluded.

The install uses `--locked`. It fails, and does not change `uv.lock`, when
the lockfile does not agree with `pyproject.toml`.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                           |
| ---------------- | ------ | --------------------------------------------- |
| `scripts/detect` | detect | `match` with a `pyproject.toml`, else `none`. |

## Contributions

| Tool     | Shared file               | Items                                                                                   |
| -------- | ------------------------- | --------------------------------------------------------------------------------------- |
| `vscode` | `.vscode/extensions.json` | `meta.pyrefly`.                                                                         |
| `dprint` | `dprint.jsonc`            | The block `ruff` (`indentWidth = 2`, `lineLength = 120`) and the plugin `@dprint/ruff`. |

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
