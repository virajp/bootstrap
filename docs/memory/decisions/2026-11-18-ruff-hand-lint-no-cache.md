DECISION 2026-11-18 ★3 | repo:bootstrap, repo/tools/*/tasks/, the hand lint of
D21 in docs/plans/2026-11-18-python-tasks-foundation | a hand-run ruff check of
a task file runs with --no-cache, so no .ruff_cache folder appears below a
tools folder | alt rejected: delete the cache after each run only; accept a
smoke check against a clean clone

# The hand lint of a Python task file runs ruff with --no-cache

This repository has no `pyproject.toml`, so no gate lints its `.py` task files
(D13, D21 of `docs/plans/2026-11-18-python-tasks-foundation`). Each unit lints
its own files by hand with `mise x ruff@latest -- ruff check`.

In the run of that plan, a hand-run `ruff check` wrote the gitignored folder
`repo/tools/mise/tasks/.ruff_cache/`. `git status` does not show it, but the
engine inventory reads each file below a `tasks/` folder. So
`bootstrap sync --clone <worktree>` stopped with code 2: each cache file "has
no entry" in `tools.toml`. The smoke check K1 failed until the run deleted the
folder.

The user ruled (Q4, D34): delete the cache, and a hand lint of a task file in
plan 2 (`docs/plans/2026-11-18-python-tasks-repo`) and later runs ruff with
`--no-cache`, so the cache does not come back:

    mise x ruff@latest -- ruff check --no-cache <files>
