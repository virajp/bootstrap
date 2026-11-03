# U2 — Move every tools.toml to the [files] tables

- **Read first:** every `tools/*/tools.toml`, top to bottom, before editing.
- **Lazy-load:** the source files that a seed names, to confirm each target
  (for example `tools/fnox/.config/fnox.toml.tmpl`).

## Ruling

From the design rulings of 2026-11-04:

> - Seeds on both sides use the setup-sync form: `[files."<target>"]` (target
>   path, no `.tmpl`), `mode = "seed"`, `requires = [...]` (exact paths
>   relative to the target root, each must exist). `seed = [...]` and
>   `[seed_requires]` go (exit 2 naming the new key). No `[files]` table =
>   `mode = "generated"`. The repo mise-entry `requires` (a file name searched
>   4 folders deep) is a second meaning of the word: solve it in the "mise
>   config" row.

> - `requires` has one meaning: exact paths, in `[files]` tables. New key
>   `requires_found` = a list of file names searched down to 4 folders below
>   the root, valid in a mise entry and in a `[files]` table, repo side only
>   (Mac → exit 2). The `files` list of a mise entry goes: each file states
>   its condition in its own `[files]` table (python, node).

From this plan:

D5: "The seed key goes from every tools.toml, the empty ones too. A seed key
with any value stops the sync with code 2 and names [files.\"<target>\"] mode =
\"seed\". A tools.toml now has four keys that the sync reads without a
default: default_envs, depends, root_only and extensions."

D9: "requires: a list of exact paths relative to the repository root
(MISE_PROJECT_ROOT); each must exist (-e); an absolute path or a .. part stops
with code 2. requires_found: a list of file names; each must be found by the
rule of `_rs_has_file` (git tracks it or does not ignore it, at the root or down
to 4 folders below it, no submodule); a name with a / stops with code 2. Both
lists must be non-empty when present. A target with both lists needs every
entry of both."

## Edits

In each file, the `[files."<target>"]` tables go after the top-level keys and
before the first other table. The target never has `.tmpl`. Keep each
comment; move a comment that describes a seed above its new table.

1. **Every `tools/*/tools.toml`** — remove the line `seed = ...`.
2. **The seeds** — add one table each, with `mode = "seed"`:
   - `tools/bun/tools.toml`: `bunfig.toml`.
   - `tools/docker/tools.toml`: `.tools/docker-compose.yaml`.
   - `tools/fnox/tools.toml`: `.config/fnox.toml`.
   - `tools/git/tools.toml`: `.config/wt.toml`.
   - `tools/gitleaks/tools.toml`: `.config/gitleaks.toml`.
   - `tools/grype/tools.toml`: `.config/grype.yaml`.
   - `tools/kotlin/tools.toml`: `.config/detekt.yml`,
     `.config/ktlint.editorconfig`.
   - `tools/mise/tools.toml`: `.config/mise/conf.d/zz-repo/mise.ci.toml`,
     `.config/mise/conf.d/zz-repo/mise.dev.toml`.
   - `tools/node/tools.toml`: `pnpm-workspace.yaml` with
     `requires = ["package.json"]`; remove `[seed_requires]` and move its
     comment above the new table.
   - `tools/osv-scanner/tools.toml`: `.config/osv-scanner.toml`.
   - `tools/swift/tools.toml`: `.config/swift-format.json`,
     `.config/swiftlint.yml`.
3. **`tools/python/tools.toml`**
   - `[mise.tools.ruff]` and `[mise.tools."pipx:pyrefly"]`: remove `files`
     and `requires`; add `requires_found = ["pyproject.toml"]`.
   - Add `[files.".config/mise/tasks/code/format/ruff"]`,
     `[files.".config/mise/tasks/code/lint/ruff"]` and
     `[files.".config/mise/tasks/code/lint/pyrefly"]`, each with
     `requires_found = ["pyproject.toml"]` and no `mode`.
   - Keep the comment above the gate entries; change "only when the
     repository has a pyproject.toml" so it names `requires_found`.
4. The comment `# The mise config, generated into ...` of each file stays.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass
  (taplo and the check-toml hook read each file).
- `grep -ln '^seed\|seed_requires' tools/*/tools.toml` finds nothing.
- `grep -c '^\[files\.' tools/kotlin/tools.toml tools/mise/tools.toml
  tools/swift/tools.toml`
  prints 2 for each; `tools/python/tools.toml` has 3.
- `grep -n 'requires = "' tools/*/tools.toml` finds nothing.
- For each `[files."<t>"]` of tool `<x>`, `tools/<x>/<t>` or
  `tools/<x>/<t>.tmpl` is a file.

## Guardrails

- Touch only `tools/*/tools.toml`. Do not touch the `files` regex of the
  `[[precommit.hooks]]` of `github` and `gitlab`: it is a pre-commit key.
- Add no `[files]` table for a generated file without a condition: no table
  means `mode = "generated"`.
- No tool gets `mode = "merge"`.
- Delete with `rm`, never `git rm`.

## Commit

`ops: move the tools.toml seeds and conditions to [files] tables` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
