# U5 — Docs

- **Read first:** the branch delta, then each passage named below.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. The task rulings are quoted in
`01-git-tool.md` and `02-mise-tool.md`.

- **D1**, a reversal of the design text that kept the old names as aliases for
  one release: no aliases. One `bootstrap@4.0.0` tag covers R1a to R2, so no
  release ships between.

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings plus every
   `DOCS FALSIFIED:` line of the earlier units.
2. **`docs/design/mise.md`, `docs/design/packs.md`, `docs/design/repo.md`,
   `docs/design/tools.md`, `docs/design/bootstrap-config.md`,
   `docs/design/cli.md`** — every task table and passage that names
   `code:merge`, `code:merge:<b>`, `setup:worktree` or `code:worktrees`: the
   new names, `git:worktree:create` (with `--from` and `--no-setup`) and
   `delete` (D3, D13 to D16), and the destination rule of
   `git:merge:<branch>` (D2).
3. **`docs/sync/index.md`, `docs/sync/tools/{git,github,gitlab,mise}.md`,
   `docs/sync/vwf.md`** — the same names; `vwf.md` names the `git:*` tasks as
   the git actions a plugin calls.
4. **`docs/sync/migrate.md`** — a migrate item: the old task names are gone
   (D1); the new names; `bootstrap init` rewrites `setup:worktree` in
   `.config/wt.toml` and in members' `worktree_install` (D20); a hand script
   or hook that ran an old name must change by hand.
   4a. **`docs/design/cli.md`** — the D20 `init` change and the code 2 of
   `check` and `sync` for an old name in those places.
   4b. **`docs/sync/tools/git.md`** — a short D20 passage where it describes
   `.config/wt.toml`: `bootstrap init` renames the old command there (D34).
5. **`README.md`, `CLAUDE.md`** — each mention of the old names.
6. **`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`** — add to
   the amendments section (or add one, "Amendments 2026-11-14 (R2 planning)"):
   R2 is three plans (R2a git tasks, R2b LSP, R2c plugins); D1 (no aliases);
   D2 to D4. Do not rewrite the rest of the record.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rn 'code:merge\|setup:worktree\|code:worktrees' README.md CLAUDE.md
  docs/design docs/sync`
  finds only the migrate item and the D20 passages in `docs/design/cli.md` and
  `docs/sync/tools/git.md`, which name the old names on purpose (D30).

## Guardrails

- Touch nothing outside the owned paths. `docs/plans/` and `docs/scratchpad/`
  are not this unit's.
- Delete with `rm`, never `git rm`.

## Commit

`docs: the git:* tasks`
