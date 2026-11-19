# U3 — The Go cases of the Mac tasks

- **Read first:** `e2e/internal/harness/harness.go`,
  `e2e/mac/commands_test.go`, `e2e/mac/helpers_test.go`.
- **Lazy-load:** `e2e/mac/sync_test.go`, `e2e/mac/mac_test.go`,
  `e2e/macostools/macostools_test.go`, when `grep` finds an old task name.

## Ruling

- **D10:** "uv becomes a required member of the Mac base pack: uv = { required =
  true } in macos/packs.toml [...]. [...] the Mac mise tool gets depends =
  [\"cli\", \"python\", \"uv\"]: then both are in the init stage."
- **D11:** "A direct run (outside mise run) needs uv on PATH for the
  shebang. [...] The Go cases run it directly with a PATH that gives uv (mise
  x uv@latest python@3.14 -- <file>)."
- **D6:** "U3 updates the Go cases of e2e/mac/ and e2e/macostools/."
- Repo plan D2: the task names, flags, output and exit codes do not change,
  so an expected output does not change. A case that fails only because the
  output changed is a QUESTION, which the orchestrator routes to U1 or U2.

## Edits

1. `entryRun` and each other case that runs a Mac task with `bash <file>`:
   run the `.py` file directly with uv on PATH (D11), for example
   `mise x uv@latest python@3.14 -- <file>`. `e2e/mac/helpers_test.go` copies
   `updateall-configs.py` in place of `updateall-configs`.
2. A case that expects uv to be optional in the Mac base pack expects it
   required now, and a case that lists the init-stage tools expects `python`
   and `uv` among them (D10).
3. Each expected list of synced files and each task path in `e2e/mac/` and
   `e2e/macostools/`: the new `.py` names, and no shell Mac helpers.

## Verification

- `mise x -- go vet ./e2e/mac/ ./e2e/macostools/` passes.
- `mise x -- go test ./e2e/mac/ ./e2e/macostools/` passes, and `-count=1`
  passes again.

## Guardrails

- Do not edit a file below `e2e/internal/`. Do not edit a task.
- Keep each check name text, except the words "under bash" in the case
  names of `e2e/mac/mac_test.go:205-206` and the comments of
  `e2e/mac/commands_test.go:368` and `:617`: they become "directly" (D11).
- Each case uses `t.Parallel()` and the environment of the harness; never
  `os.Setenv` or `os.Chdir`.
- Write each comment in ASD-STE100.

## Commit

`test: run the Python Mac tasks in the Go cases`
