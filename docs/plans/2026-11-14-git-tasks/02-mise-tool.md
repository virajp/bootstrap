# U2 — The mise tool: worktree setup and list under git:worktree

- **Read first:** `repo/tools/mise/tools.toml` (the alias table, about line 29,
  and the `[tasks.*]` entries, about lines 98-122),
  `repo/tools/mise/tasks/code-merge`,
  `repo/tools/mise/tasks/setup-worktree.tmpl`,
  the header and the messages of `repo/tools/mise/tasks/code-worktrees`.
- **Lazy-load:** `repo/tools/mise/tasks/helpers` (about line 277),
  `repo/tools/mise/tasks/worktrees-page` (about lines 5 and 233).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1:** no aliases for the old task names.
- **D2:** mise's `code/merge/_default` (the one-branch `code:merge <src>`) goes
  away; the git tool's `git:merge:<branch>` replaces it (U1).
- **D4:** `code:worktrees` becomes `git:worktree:list`
  (`git/worktree/list`); the shell alias `worktrees` runs
  `mise run git:worktree:list`. Its content does not change.
- **D5:** `setup:worktree` becomes `git:worktree:setup`
  (`git/worktree/setup`) and stays in the mise tool; its content does not
  change.

## Edits

1. **`repo/tools/mise/tools.toml`** — `[tasks."git/worktree/setup"]` (source
   `setup-worktree.tmpl`, template mode as today) and
   `[tasks."git/worktree/list"]` (source `code-worktrees`); remove
   `[tasks."code/merge/_default"]`, `[tasks."setup/worktree"]` and
   `[tasks."code/worktrees"]`; the alias `worktrees = "mise run
   git:worktree:list"`.
2. **`repo/tools/mise/tasks/code-merge`** — remove with `rm`.
3. **`repo/tools/mise/tasks/setup-worktree.tmpl`, `code-worktrees`,
   `worktrees-page`, `helpers`** — only the `#MISE description`, comments and
   messages that name `setup:worktree`, `code:worktrees` or `code:merge`. The
   source file names may stay.
4. **`repo/tools/mise/tasks/setup.tmpl`** and any other file of the mise tool
   that runs `setup:worktree` or `code:worktrees` by name — the new names.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rn 'code:merge\|code:worktrees\|setup:worktree\|code/merge'
  repo/tools/mise`
  finds nothing.

## Guardrails

- Touch nothing outside `repo/tools/mise/`. `repo/tools/git/` is U1's.
- Do not change what the setup or the lister does.
- Delete with `rm`, never `git rm`.

## Commit

`feat: git:worktree:setup and git:worktree:list replace the old names`
