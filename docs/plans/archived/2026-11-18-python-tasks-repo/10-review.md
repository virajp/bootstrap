# U10 — Review: the Python tasks of repo/tools and their cases

## Scope

D11: the plan lands runnable code (every task that `repo/tools/` ships, the
own tasks of this repository and the engine fix). The review covers U1 to
U9, U13 and U14, the whole branch delta since the branch base.

Focus: each task keeps its name, flags, output and exit codes (D2); each
`subprocess` call takes a list and no `shell=True`, and no value of a file,
a branch name, a member URL or the engine answer reaches a shell; the
templates render to valid Python with no gomplate delimiter left (D5); the
tool exceptions are only the three of D4; the `code:sec` scanners keep their
own lists; the merge and worktree tasks keep their git commands in the same
order; no Go case lost a check name; the own tasks of U13 keep their behavior;
the
engine fix of U14 changes only the compare of `UnnamedParts`.
