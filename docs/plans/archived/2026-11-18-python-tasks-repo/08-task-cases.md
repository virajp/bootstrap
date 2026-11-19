# U8 — The Go cases of the Python tasks

- **Read first:** `e2e/internal/harness/harness.go`,
  `e2e/internal/repocase/repocase.go`, `e2e/repo/commands/tasks_test.go`,
  `e2e/repo/tasks/*_test.go` and `e2e/repo/pylib/*_test.go`.
- **Lazy-load:** the other packages of `e2e/repo/` and
  `e2e/macostools/macostools_test.go`, when `grep` finds an old task name or
  a `bash` call of a task in them.

## Ruling

- **D9:** "U8 updates each Go case of e2e/repo/ and e2e/macostools/ that runs
  or names a converted task, and changes the byte compare of e2e/repo/pylib/
  from the shell helpers to fixed expected bytes, because U9 deletes the shell
  helpers."
- **D2:** the task names, the flags, the output and the exit codes do not
  change, so an expected output of a case does not change. A case that fails
  only because the output changed is a QUESTION to the orchestrator, which
  routes it to the unit of the task.
- **D3:** the target of a task gets `.py`; a template source is
  `<name>.py.tmpl`.
- **D8:** no task sources a shell library after this plan.

## Edits

1. Each case that runs a rendered task with `bash <file>`: run the `.py` file
   directly (its shebang runs uv), or with
   `uv run --quiet --script <file>`, in the form that the harness gives.
2. The stub `merge` library and the stub engines that are Go string
   constants: the stub `merge` becomes a Python `merge.py` stub that prints
   its arguments in the same form; a stub engine stays a shell script,
   because it stands for the binary.
3. Each expected list of synced files and each task path in
   `e2e/repo/{config,detect,files,migrate,sync,tasks,commands}/` and
   `e2e/macostools/`: the new `.py` names, and no shell library.
4. `e2e/repo/pylib/`: replace each compare with the shell `helpers`,
   `checks`, `worktrees-page`, `merge` and `xcode` by fixed expected bytes or
   values. U9 deletes the shell libraries before this unit: take the expected
   bytes from a run of the shell library of the branch base
   (`git show <base>:<path>` into a scratch folder). The same holds for
   `e2e/repo/migrate/migrate_test.go` (about line 523), which sources the
   shell `_scripts/helpers` for its expected bytes: it uses fixed expected
   bytes.

## Verification

- `mise x -- go vet ./e2e/...` passes.
- `mise x -- go test ./e2e/...` passes, and `-count=1` passes again.
- `grep -rn '"bash"' e2e/repo e2e/macostools` finds no call that runs a
  task file (a stub engine or a git script is allowed).

## Guardrails

- Do not edit a file below `e2e/internal/`.
- Do not edit a task: a task failure is a QUESTION.
- Each case uses `t.Parallel()`, `harness.Dir(t)` and `harness.Env()`;
  never `os.Setenv` or `os.Chdir`.
- Keep each check name text: a renamed check hides a lost case.
- Write each comment in ASD-STE100.

## Commit

`test: run the Python tasks in the Go cases`
