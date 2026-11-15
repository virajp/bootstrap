# U4 — Tests

- **Read first:** `.config/mise/tasks/test/repo` (its case helpers, the
  `run_gate` cases of `code-lint-bootstrap` at about lines 4214-4260, and the
  render cases of `setup-configs.tmpl` at about lines 3904-3958; the line
  numbers are from before the plans that this plan requires). Then the U1 to
  U3 results (read only): `repo/tools/mise/tasks/gate_files.py`,
  `repo/tools/mise/tasks/helpers`, `code-format.tmpl`, `code-lint.tmpl`,
  `repo/tools/mise/files/.config/bootstrap/config-files.toml.tmpl`.
- **Lazy-load:** CLAUDE.md ("Tests").

## Ruling

- **D1, D16, D12, D11, D13, D17, D9** as U1 quotes them. The cases assert each
  rule.
- Each case runs in its own scratch folder, with the `GIT_*` variables unset,
  as the other cases of the task do.

## Edits

1. **`.config/mise/tasks/test/repo`** — add cases that run `gate_files.py`
   (and, where a case needs it, the rendered `helpers` function) in a scratch
   git repo:
   - **default mode:** a staged change, an unstaged change, a deleted file and
     a new file; the list holds the staged, the changed and the new paths only
     (D1);
   - **`--all`:** every tracked and new path, no deleted path;
   - **given files:** an ignored path, a path in a submodule and a missing path
     are dropped (D16);
   - **submodule and nested repository:** no path below either in any mode
     (D16);
   - **global filter:** a `config-files.toml` with an exclude and an include;
     a missing `config-files.toml` means two empty lists (D9);
   - **gate filter:** a `#FILES` header with `include` and `exclude`; a gate
     with no header gets every file (D3); the `shebang` key takes a file with
     `#!/usr/bin/env bash` and leaves a zsh file (D11);
   - **pattern syntax:** a pattern without `/` matches in each folder; `**`
     matches any number of folders (D12);
   - **`--debug`:** stderr has `drop <path>: <layer> <kind> "<pattern>"` for a
     dropped path (D13);
   - **empty list:** an empty list file, exit 0;
   - **malformed header (D44):** two `#FILES` lines, an unknown key and a
     value without quotes each exit 2 with a message;
   - **symlink and missing path (D48):** a tracked symlink prints
     `repo symlink`; a given path that git does not list prints
     `repo unlisted`;
   - **a pattern without `/` (D46):** `node_modules` drops each path below a
     `node_modules` folder at any depth;
   - **shebang (D47):** a `*.sh` file with `#!/usr/bin/env python3` stays; one
     with `#!/bin/zsh` is dropped (`gate shebang`);
   - **orchestrator:** the rendered `code:lint` with `--all` and a file stops
     with exit 2 (D17);
   - **render:** `bootstrap sync` writes `.config/bootstrap/config-files.toml`
     with the mise defaults and a value that a consumer sets in
     `[tools.mise] code_files.excludes`, and `dprint.jsonc` gets the same
     excludes (D8, D9).
     **D58:** the `run_gate` cases of `code-lint-bootstrap` copy
     `repo/tools/mise/tasks/helpers` and `gate_files.py` (not the synced copy at
     the root) and run the gate with `--all`.
     Change each old case that this plan falsifies (for example a case that
     expects `excludes` from the dprint `[defaults]`).

## Verification

- `mise x -- mise run test:repo` passes.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo`. A defect in a U1 to U3
  file is a QUESTION, never a fix here.
- Write shell for bash 3.2 and bash 5.x.
- Delete with `rm`, never `git rm`.

## Commit

`test: the gate file sets`
