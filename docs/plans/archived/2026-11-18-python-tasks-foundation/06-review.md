# U6 — Review: the Python libraries, the gate fixes and their cases

## Scope

D20: the plan lands runnable code (Python libraries and shell gate fixes that
each consumer runs). The review covers U1 to U5, the whole branch delta since
the branch base.

Focus: each `subprocess` call takes a list and no `shell=True`, and no value
of a file, a branch name or the engine answer reaches a shell; `bootstrap_get`
keeps the exit codes 1 and 2 of D10; the print functions give the same bytes
as the shell (D9); `own_tasks` returns task names without `.py` (D18); the
modules use the standard library only (D7); the shell edits of U4 stay valid
for bash 3.2; no Go case uses `os.Setenv` or `os.Chdir`.
