# U4 — Review: the git tasks and the tests

## Scope

U1, U2, U3 and U7, the whole branch delta since the branch base. The tasks merge
branches, push, and remove worktrees and branches in a user's repo (D7).
Focus: `git:merge:<branch>` always merges into `<branch>` and never the
reverse (D2); the `when` keys give exactly the tasks of the repo's core
branches; `git:worktree:delete` runs `git branch -D` only after it proves the
branch is merged into a core branch, never deletes a core branch, and never
removes a dirty worktree (D3, D15, D22); `init` changes only the old command
text in `wt.toml` and `worktree_install` (D20); no file still runs an old task
name (D1).
