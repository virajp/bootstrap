# U4 — Review: the Mac tasks and their cases

## Scope

D8: the plan lands runnable code (every Mac task). The review covers U1 to
U3, the whole branch delta since the branch base.

Focus: each task keeps its name, flags, output and exit codes; `func-claude`
and `func-myps` end with `os.execvp` and the same arguments (D3); each
`subprocess` call takes a list and no `shell=True`; the entry task
`updateall:configs` keeps its lock, clone and temporary copy, and never
writes outside the paths it writes now; the Mac helpers print the same bytes
(D2); no Go case lost a check name.
