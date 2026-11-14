# U2 — The virajp-plugins tool and the move of the vwf lines

- **Read first:** `repo/tools/git/files/gitignore.merge.toml` (the vwf lines:
  the section "==== vwf: working notes and scratch ====", about lines 611-619,
  and the comment and negation `!docs/plans/**/*.log`, about lines 182-186),
  `repo/tools/claude/tools.toml`, a small selectable tool as a model (for
  example `repo/tools/github/tools.toml`), and U1's contributed-items format
  (`01-contributed-items.md`).
- **Lazy-load:** `repo/packs.toml`, `docs/design/tools.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1:** the vwf `.gitignore` lines move out of the git tool's
  `gitignore.merge.toml` into `virajp-plugins`, as contributed items
  `[[contributes_items.git.".gitignore".item]]` (D11). The lines:
  `docs/memory/handoff/`, `docs/memory/doctor/`, `docs/memory/runs/`,
  `docs/scratchpad/`, and the negation `!docs/plans/**/*.log`. The generic
  `*.log` line stays in the git tool.
- **D16:** remove the vwf sentence of the comment above `*.log` from the git
  file; in `virajp-plugins/tools.toml`, a TOML comment above the negation item
  says that vwf's plan logs stay tracked.
- **D2:** `virajp-plugins` contributes the plugin entry to the claude tool's
  values: `[contributes.claude.plugins.user]`
  `"vwf@virajp-plugins" = "virajp/claude-plugins"`.
- **D6:** `virajp-plugins` is a selectable tool, not in the base pack. It
  depends on `claude` and `git`. Its description says it carries vwf's
  ignore lines and plugin entry.
- **D21:** after this unit, the git tool's `gitignore.merge.toml` names no vwf
  path. Other vwf mentions outside R2c's files are parked.

## Edits

1. **`repo/tools/virajp-plugins/tools.toml`** (new) — the tool: category,
   description, `depends = ["claude", "git"]`, `default_envs`, `root_only`,
   selectable, and the D1 and D2 contributions.
2. **`repo/tools/git/files/gitignore.merge.toml`** — remove the vwf section
   and the `!docs/plans/**/*.log` negation with its comment (D1).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'docs/memory\|docs/scratchpad\|docs/plans'
  repo/tools/git/files/gitignore.merge.toml`
  finds nothing.

## Guardrails

- Touch nothing outside `repo/tools/virajp-plugins/` and
  `repo/tools/git/files/gitignore.merge.toml`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: the virajp-plugins tool carries vwf's ignore lines and plugin entry`
